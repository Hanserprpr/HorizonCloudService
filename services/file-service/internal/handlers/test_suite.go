package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strings"
	"time"
	"file-service/internal/middleware"
	"file-service/internal/models"
	"file-service/internal/repository"
	"file-service/internal/services"
	"file-service/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestSuite 测试套件基类
type TestSuite struct {
	suite.Suite
	DB         *gorm.DB
	Router     *gin.Engine
	Storage    storage.Storage
	Services   *services.Services
	UserClient services.UserServiceClient
	Repos      *repository.Repository
	TestUser   *TestUser
	AdminUser  *TestUser
}

// TestUser 测试用户
type TestUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Token    string `json:"token"`
}

// SetupSuite 设置测试套件
func (s *TestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	// 设置内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	s.NoError(err)
	s.DB = db

	// 运行数据库迁移
	err = db.AutoMigrate(
		&models.File{},
		&models.Folder{},
		&models.Thumbnail{},
		&models.UploadSession{},
	)
	s.NoError(err)

	// 创建本地存储用于测试
	tmpDir, err := os.MkdirTemp("", "file-service-test-*")
	s.NoError(err)

	localStorage, err := storage.NewLocalStorage(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: tmpDir,
	})
	s.NoError(err)
	s.Storage = localStorage

	// 创建仓库
	s.Repos = repository.NewRepository(db)

	// 创建用户服务客户端Mock
	s.UserClient = services.NewMockUserServiceClient()

	// 创建服务配置
	config := &services.ServicesConfig{
		DefaultChunkSize:    1024 * 1024, // 1MB
		ThumbnailSizes:      []string{"small", "medium", "large"},
		ThumbnailQuality:    85,
		ThumbnailTimeout:    30 * time.Second,
		MaxBatchSize:        100,
		BatchConcurrency:    5,
		SearchLimit:         100,
	}

	// 创建服务
	s.Services = services.NewServices(s.Repos, s.Storage, config)

	// 设置测试路由
	s.Router = s.setupRouter()

	// 创建测试用户
	s.setupTestUsers()
}

// TearDownSuite 清理测试套件
func (s *TestSuite) TearDownSuite() {
	if s.Storage != nil {
		// 清理测试存储目录
		// TODO: Implement storage cleanup through public interface if needed
		// For now, skip cleanup to avoid accessing private localStorage type
	}
}

// SetupTest 每个测试前的设置
func (s *TestSuite) SetupTest() {
	// 清理数据库
	s.DB.Exec("DELETE FROM files")
	s.DB.Exec("DELETE FROM folders")
	s.DB.Exec("DELETE FROM shares")
	s.DB.Exec("DELETE FROM thumbnails")
	s.DB.Exec("DELETE FROM upload_sessions")
}

// setupRouter 设置路由
func (s *TestSuite) setupRouter() *gin.Engine {
	// JWT 配置
	authConfig := &middleware.AuthConfig{
		JWTSecret:     []byte("test-jwt-secret-key"),
		JWTExpiration: time.Hour,
	}

	// 配额配置
	quotaConfig := &middleware.QuotaConfig{
		DefaultStorageQuota:   100 * 1024 * 1024, // 100MB
		DefaultFileCount:      1000,
		CheckInterval:         time.Minute,
		GraceBuffer:           0.1,
		EnableWarnings:        true,
		WarningThreshold:      0.8,
	}

	// 创建路由引擎 - 避免循环导入，直接内联路由设置
	router := gin.New()
	router.Use(gin.Recovery())

	// 创建中间件
	authMiddleware := middleware.NewAuthMiddleware(authConfig)
	quotaMiddleware := middleware.NewQuotaMiddleware(quotaConfig, s.UserClient, s.Repos.File)

	// 创建处理器
	fileHandler := NewFileHandler(s.Services)
	folderHandler := NewFolderHandler(s.Services)
	uploadHandler := NewUploadHandler(s.Services)
	thumbnailHandler := NewThumbnailHandler(s.Services)
	healthHandler := NewHealthHandler(s.Services)

	// 健康检查路由（无需认证）
	router.GET("/health", healthHandler.Health)
	router.GET("/health/ready", healthHandler.Ready)

	// API路由（需要认证）
	api := router.Group("/api/v1")
	api.Use(authMiddleware.AuthRequired())
	api.Use(quotaMiddleware.CheckQuota())

	// 文件管理API
	files := api.Group("/files")
	{
		files.GET("", fileHandler.ListFiles)
		files.POST("", fileHandler.UploadFile)
		files.GET("/:id", fileHandler.GetFile)
		files.PUT("/:id", fileHandler.UpdateFile)
		files.DELETE("/:id", fileHandler.DeleteFile)
		files.GET("/:id/download", fileHandler.DownloadFile)
		files.POST("/:id/copy", fileHandler.CopyFile)
		files.PUT("/:id/move", fileHandler.MoveFile)
		files.GET("/search", fileHandler.SearchFiles)
		files.POST("/batch", fileHandler.BatchOperation)
		files.GET("/stats", fileHandler.GetUserStats)
	}

	// 文件夹管理API
	folders := api.Group("/folders")
	{
		folders.GET("", folderHandler.ListFolders)
		folders.POST("", folderHandler.CreateFolder)
		folders.GET("/:id", folderHandler.GetFolder)
		folders.PUT("/:id", folderHandler.UpdateFolder)
		folders.DELETE("/:id", folderHandler.DeleteFolder)
		folders.GET("/:id/contents", folderHandler.GetFolderContents)
		folders.GET("/tree", folderHandler.GetFolderTree)
	}

	// 上传API
	upload := api.Group("/upload")
	{
		upload.POST("/simple", uploadHandler.SimpleUpload)
		upload.POST("/initiate", uploadHandler.InitiateUpload)
		upload.POST("/chunk", uploadHandler.UploadChunk)
		upload.POST("/complete", uploadHandler.CompleteUpload)
		upload.POST("/abort", uploadHandler.AbortUpload)
		upload.GET("/status/:session_id", uploadHandler.GetUploadProgress)
		upload.GET("/sessions", uploadHandler.ListUploadSessions)
	}

	// 缩略图API
	thumbnails := api.Group("/thumbnails")
	{
		thumbnails.GET("/:file_id", thumbnailHandler.GetThumbnail)
		thumbnails.POST("/:file_id/generate", thumbnailHandler.GenerateThumbnail)
		thumbnails.GET("/:file_id/list", thumbnailHandler.GetFileThumbnails)
		thumbnails.DELETE("/:file_id", thumbnailHandler.DeleteThumbnail)
	}

	return router
}

// setupTestUsers 设置测试用户
func (s *TestSuite) setupTestUsers() {
	// 创建JWT中间件用于生成token
	authConfig := &middleware.AuthConfig{
		JWTSecret:     []byte("test-jwt-secret-key"),
		JWTExpiration: time.Hour,
	}
	authMiddleware := middleware.NewAuthMiddleware(authConfig)

	// 创建普通用户
	testUserToken, err := authMiddleware.GenerateJWT(1, "testuser", "test@example.com", "user")
	s.NoError(err)
	s.TestUser = &TestUser{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		Token:    testUserToken,
	}

	// 创建管理员用户
	adminUserToken, err := authMiddleware.GenerateJWT(2, "admin", "admin@example.com", "admin")
	s.NoError(err)
	s.AdminUser = &TestUser{
		ID:       2,
		Username: "admin",
		Email:    "admin@example.com",
		Role:     "admin",
		Token:    adminUserToken,
	}
}

// MakeRequest 发送HTTP请求
func (s *TestSuite) MakeRequest(method, url string, body io.Reader, user *TestUser, contentType ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, body)
	if user != nil {
		req.Header.Set("Authorization", "Bearer "+user.Token)
	}
	if len(contentType) > 0 {
		req.Header.Set("Content-Type", contentType[0])
	}

	resp := httptest.NewRecorder()
	s.Router.ServeHTTP(resp, req)
	return resp
}

// MakeJSONRequest 发送JSON请求
func (s *TestSuite) MakeJSONRequest(method, url string, data interface{}, user *TestUser) *httptest.ResponseRecorder {
	jsonData, _ := json.Marshal(data)
	return s.MakeRequest(method, url, bytes.NewBuffer(jsonData), user, "application/json")
}

// MakeMultipartRequest 发送multipart请求
func (s *TestSuite) MakeMultipartRequest(method, url string, fields map[string]string, files map[string][]byte, user *TestUser) *httptest.ResponseRecorder {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加字段
	for key, value := range fields {
		writer.WriteField(key, value)
	}

	// 添加文件
	for filename, content := range files {
		part, err := writer.CreateFormFile("file", filename)
		s.NoError(err)
		part.Write(content)
	}

	writer.Close()
	return s.MakeRequest(method, url, body, user, writer.FormDataContentType())
}

// AssertSuccessResponse 断言成功响应
func (s *TestSuite) AssertSuccessResponse(resp *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	s.Equal(expectedStatus, resp.Code)
	
	var result map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &result)
	s.NoError(err)
	s.Equal(true, result["success"])
	
	return result
}

// AssertErrorResponse 断言错误响应
func (s *TestSuite) AssertErrorResponse(resp *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	s.Equal(expectedStatus, resp.Code)
	
	var result map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &result)
	s.NoError(err)
	s.Equal(false, result["success"])
	s.Contains(result, "error")
	
	return result
}

// AssertPaginatedResponse 断言分页响应
func (s *TestSuite) AssertPaginatedResponse(resp *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	result := s.AssertSuccessResponse(resp, expectedStatus)
	
	data, ok := result["data"].(map[string]interface{})
	s.True(ok, "Response data should be an object")
	
	s.Contains(data, "data", "Response should contain data array")
	s.Contains(data, "total", "Response should contain total count")
	s.Contains(data, "page", "Response should contain page number")
	s.Contains(data, "page_size", "Response should contain page size")
	
	return result
}

// CreateTestFile 创建测试文件
func (s *TestSuite) CreateTestFile(userID uint, name, content string) *models.File {
	file := &models.File{
		Name:        name,
		Size:        int64(len(content)),
		ContentType: "text/plain",
		UserID:      userID,
		Hash:        fmt.Sprintf("hash-%s", name),
		Path:        fmt.Sprintf("/test/%s", name),
		Status:      models.FileStatusActive,
		Category:    "document",
	}
	
	err := s.DB.Create(file).Error
	s.NoError(err)
	
	// 上传到存储
	ctx := context.Background()
	_, err = s.Storage.Upload(ctx, file.Path, strings.NewReader(content), file.Size, &storage.UploadOptions{
		ContentType: file.ContentType,
	})
	s.NoError(err)
	
	return file
}

// CreateTestFolder 创建测试文件夹
func (s *TestSuite) CreateTestFolder(userID uint, name string, parentID *uint) *models.Folder {
	folder := &models.Folder{
		Name:        name,
		UserID:      userID,
		ParentID:    parentID,
		Description: fmt.Sprintf("Test folder: %s", name),
		Path:        fmt.Sprintf("/%s", name),
	}
	
	if parentID != nil {
		var parent models.Folder
		err := s.DB.First(&parent, *parentID).Error
		s.NoError(err)
		folder.Path = fmt.Sprintf("%s/%s", parent.Path, name)
	}
	
	err := s.DB.Create(folder).Error
	s.NoError(err)
	
	return folder
}