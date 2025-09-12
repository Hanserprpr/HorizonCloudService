package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"file-service/internal/config"
	"file-service/internal/handlers"
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
	DB            *gorm.DB
	Router        *gin.Engine
	Storage       storage.Storage
	Services      *services.Services
	UserClient    services.UserServiceClient
	Repos         *repository.Repository
	TestUser      *TestUser
	AdminUser     *TestUser
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
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建内存SQLite数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	s.Require().NoError(err)
	s.DB = db

	// 自动迁移数据库
	err = s.DB.AutoMigrate(
		&models.File{},
		&models.Folder{},
		&models.Share{},
		&models.Thumbnail{},
		&models.UploadSession{},
	)
	s.Require().NoError(err)

	// 创建本地存储
	tmpDir, err := os.MkdirTemp("", "file-service-test-*")
	s.Require().NoError(err)

	localStorage, err := storage.NewLocalStorage(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: tmpDir,
	})
	s.Require().NoError(err)
	s.Storage = localStorage

	// 创建Mock用户服务客户端
	s.UserClient = services.NewMockUserServiceClient()

	// 初始化仓库
	repo := repository.NewRepository(s.DB)

	// 保存repos引用
	s.Repos = repo

	// 初始化服务
	serviceConfig := &services.ServicesConfig{
		DefaultChunkSize:    1024 * 1024, // 1MB for testing
		ThumbnailSizes:      []string{"small", "medium", "large"},
		ThumbnailQuality:    85,
		ThumbnailTimeout:    30 * time.Second,
		MaxBatchSize:        100,
		BatchConcurrency:    5,
		SearchLimit:         100,
	}
	s.Services = services.NewServices(repo, s.Storage, serviceConfig)

	// 创建测试用户
	s.TestUser = &TestUser{
		ID:       2,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}

	s.AdminUser = &TestUser{
		ID:       1,
		Username: "admin",
		Email:    "admin@example.com",
		Role:     "admin",
	}

	// 生成JWT Token
	authMiddleware := middleware.NewAuthMiddleware(&middleware.AuthConfig{
		JWTSecret:     []byte("test-secret"),
		JWTExpiration: 24 * time.Hour,
	})

	s.TestUser.Token, err = authMiddleware.GenerateJWT(
		s.TestUser.ID, s.TestUser.Username, s.TestUser.Email, s.TestUser.Role,
	)
	s.Require().NoError(err)

	s.AdminUser.Token, err = authMiddleware.GenerateJWT(
		s.AdminUser.ID, s.AdminUser.Username, s.AdminUser.Email, s.AdminUser.Role,
	)
	s.Require().NoError(err)

	// 创建路由
	s.Router = s.createTestRouter()
}

// TearDownSuite 清理测试套件
func (s *TestSuite) TearDownSuite() {
	if s.Storage != nil {
		// Note: Storage cleanup is handled automatically for test storage
		// For production deployments, implement proper cleanup
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

// createTestRouter 创建测试路由
func (s *TestSuite) createTestRouter() *gin.Engine {
	authConfig := &middleware.AuthConfig{
		JWTSecret:     []byte("test-secret"),
		JWTExpiration: 24 * time.Hour,
		SkipPaths:     []string{"/api/v1/health"},
	}

	quotaConfig := &middleware.QuotaConfig{
		DefaultStorageQuota:  5 * 1024 * 1024 * 1024, // 5GB
		DefaultFileCount:     10000,
		CheckInterval:        5 * time.Minute,
		GraceBuffer:          0.1,
		EnableWarnings:       true,
		WarningThreshold:     0.8,
	}

	// 创建路由引擎 - 避免循环导入，直接内联路由设置
	router := gin.New()
	router.Use(gin.Recovery())

	// 创建中间件
	authMiddleware := middleware.NewAuthMiddleware(authConfig)
	quotaMiddleware := middleware.NewQuotaMiddleware(quotaConfig, s.UserClient, s.Repos.File)

	// 创建处理器
	fileHandler := handlers.NewFileHandler(s.Services)
	folderHandler := handlers.NewFolderHandler(s.Services)
	uploadHandler := handlers.NewUploadHandler(s.Services)
	thumbnailHandler := handlers.NewThumbnailHandler(s.Services)
	healthHandler := handlers.NewHealthHandler(s.Services)

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

// CreateTestFile 创建测试文件
func (s *TestSuite) CreateTestFile(userID uint, name, content string) *models.File {
	file := &models.File{
		Name:        name,
		Size:        int64(len(content)),
		ContentType: "text/plain",
		UserID:      userID,
		Status:      models.FileStatusActive,
		Hash:        fmt.Sprintf("hash-%s", name),
		Path: fmt.Sprintf("/%d/%s", userID, name),
		Category:    "document",
		StorageTier: "hot",
	}

	err := s.DB.Create(file).Error
	s.Require().NoError(err)

	return file
}

// CreateTestFolder 创建测试文件夹
func (s *TestSuite) CreateTestFolder(userID uint, name string, parentID *uint) *models.Folder {
	folder := &models.Folder{
		Name:     name,
		UserID:   userID,
		ParentID: parentID,
		Path:     "/" + name,
		IsSystem: false,
	}

	if parentID != nil {
		var parent models.Folder
		s.DB.First(&parent, *parentID)
		folder.Path = parent.Path + "/" + name
	}

	err := s.DB.Create(folder).Error
	s.Require().NoError(err)

	return folder
}

// MakeRequest 发送HTTP请求
func (s *TestSuite) MakeRequest(method, url string, body interface{}, user *TestUser) *httptest.ResponseRecorder {
	var reqBody io.Reader
	contentType := "application/json"

	if body != nil {
		switch v := body.(type) {
		case string:
			reqBody = strings.NewReader(v)
		case []byte:
			reqBody = bytes.NewReader(v)
		default:
			jsonData, err := json.Marshal(body)
			s.Require().NoError(err)
			reqBody = bytes.NewReader(jsonData)
		}
	}

	req := httptest.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", contentType)

	if user != nil {
		req.Header.Set("Authorization", "Bearer "+user.Token)
	}

	recorder := httptest.NewRecorder()
	s.Router.ServeHTTP(recorder, req)

	return recorder
}

// MakeMultipartRequest 发送multipart请求
func (s *TestSuite) MakeMultipartRequest(method, url string, fields map[string]string, files map[string][]byte, user *TestUser) *httptest.ResponseRecorder {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// 添加字段
	for key, value := range fields {
		err := writer.WriteField(key, value)
		s.Require().NoError(err)
	}

	// 添加文件
	for filename, content := range files {
		part, err := writer.CreateFormFile("file", filename)
		s.Require().NoError(err)
		_, err = part.Write(content)
		s.Require().NoError(err)
	}

	err := writer.Close()
	s.Require().NoError(err)

	req := httptest.NewRequest(method, url, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if user != nil {
		req.Header.Set("Authorization", "Bearer "+user.Token)
	}

	recorder := httptest.NewRecorder()
	s.Router.ServeHTTP(recorder, req)

	return recorder
}

// AssertSuccessResponse 断言成功响应
func (s *TestSuite) AssertSuccessResponse(response *httptest.ResponseRecorder, expectedCode ...int) map[string]interface{} {
	code := http.StatusOK
	if len(expectedCode) > 0 {
		code = expectedCode[0]
	}

	s.Assert().Equal(code, response.Code)

	var result map[string]interface{}
	err := json.Unmarshal(response.Body.Bytes(), &result)
	s.Require().NoError(err)

	s.Assert().Equal(float64(200), result["code"])
	s.Assert().Contains(result, "data")

	return result
}

// AssertErrorResponse 断言错误响应
func (s *TestSuite) AssertErrorResponse(response *httptest.ResponseRecorder, expectedCode int, expectedMessage ...string) map[string]interface{} {
	s.Assert().Equal(expectedCode, response.Code)

	var result map[string]interface{}
	err := json.Unmarshal(response.Body.Bytes(), &result)
	s.Require().NoError(err)

	s.Assert().Equal(float64(expectedCode), result["code"])

	if len(expectedMessage) > 0 {
		s.Assert().Contains(result["message"], expectedMessage[0])
	}

	return result
}

// AssertPaginatedResponse 断言分页响应
func (s *TestSuite) AssertPaginatedResponse(response *httptest.ResponseRecorder) map[string]interface{} {
	result := s.AssertSuccessResponse(response)

	data := result["data"].(map[string]interface{})
	s.Assert().Contains(data, "page")
	s.Assert().Contains(data, "page_size")
	s.Assert().Contains(data, "total")
	s.Assert().Contains(data, "pages")
	s.Assert().Contains(data, "data")

	return result
}

// CreateTestConfig 创建测试配置
func CreateTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:        "file-service-test",
			Version:     "1.0.0-test",
			Environment: "test",
			Debug:       true,
		},
		Server: config.ServerConfig{
			Port:               8002,
			ReadTimeoutSeconds: 30,
			WriteTimeoutSeconds: 30,
			IdleTimeoutSeconds: 120,
		},
		Database: config.DatabaseConfig{
			Host:     ":memory:",
			LogLevel: "silent",
		},
		JWT: config.JWTConfig{
			Secret:          "test-secret",
			ExpirationHours: 24,
		},
		Storage: config.StorageConfig{
			Backend: "local",
			Local: config.LocalConfig{
				RootPath: "./test-storage",
			},
		},
	}
}

// TestFileContent 测试文件内容
var TestFileContent = []byte("This is test file content for testing purposes.")

// TestImageContent 测试图片内容 (简单的PNG数据)
var TestImageContent = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
	0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49,
	0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
}

// WaitForAsync 等待异步操作完成
func (s *TestSuite) WaitForAsync(timeout time.Duration) {
	time.Sleep(100 * time.Millisecond) // 简单的等待，实际项目中可以用channel通知
}

// CleanupStorage 清理存储
func (s *TestSuite) CleanupStorage() {
	// Note: For test storage, cleanup is handled automatically
	// In production, implement proper cleanup through public interface
}

// GetResponseData 获取响应数据
func (s *TestSuite) GetResponseData(response *httptest.ResponseRecorder) map[string]interface{} {
	var result map[string]interface{}
	err := json.Unmarshal(response.Body.Bytes(), &result)
	s.Require().NoError(err)
	return result["data"].(map[string]interface{})
}