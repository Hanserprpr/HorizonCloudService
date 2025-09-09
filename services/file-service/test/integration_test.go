package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

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

type IntegrationTestSuite struct {
	suite.Suite
	db             *gorm.DB
	router         *gin.Engine
	storage        storage.Storage
	services       *services.Services
	authMiddleware *middleware.AuthMiddleware
	testUser       *TestUser
	adminUser      *TestUser
}

type TestUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Token    string `json:"token"`
}

func (s *IntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	// 设置内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	s.NoError(err)
	s.db = db

	// 运行数据库迁移
	err = db.AutoMigrate(
		&models.File{},
		&models.Folder{},
		&models.Thumbnail{},
		&models.UploadSession{},
	)
	s.NoError(err)

	// 创建存储实例（内存存储用于测试）
	s.storage = storage.NewMemoryStorage()

	// 创建仓库
	repos := &repository.Repositories{
		File:   repository.NewFileRepository(db),
		Folder: repository.NewFolderRepository(db),
		Upload: repository.NewUploadRepository(db),
	}

	// 创建服务配置
	config := &services.ServiceConfig{
		Storage:          s.storage,
		UserService:      services.NewMockUserServiceClient(),
		DefaultChunkSize: 1024 * 1024, // 1MB
	}

	// 创建服务
	s.services = services.NewServices(repos, config)

	// 创建认证中间件
	authConfig := &middleware.AuthConfig{
		JWTSecret:     []byte("test-integration-secret"),
		JWTExpiration: time.Hour,
	}
	s.authMiddleware = middleware.NewAuthMiddleware(authConfig)

	// 创建配额中间件
	quotaConfig := &middleware.QuotaConfig{
		DefaultStorageQuota: 100 * 1024 * 1024, // 100MB
		DefaultFileCount:    1000,
		CheckInterval:       time.Minute,
		GraceBuffer:         0.1,
	}
	quotaMiddleware := middleware.NewQuotaMiddleware(quotaConfig, config.UserService, repos.File)

	// 创建处理器
	fileHandler := handlers.NewFileHandler(s.services.File, repos.File)
	folderHandler := handlers.NewFolderHandler(s.services.Folder, repos.Folder)
	uploadHandler := handlers.NewUploadHandler(s.services.Upload)
	thumbnailHandler := handlers.NewThumbnailHandler(s.services.Thumbnail)
	healthHandler := handlers.NewHealthHandler(s.services, repos, s.storage)

	// 设置路由
	s.router = gin.New()
	s.router.Use(gin.Recovery())

	// 健康检查路由（无需认证）
	s.router.GET("/health", healthHandler.Health)
	s.router.GET("/health/ready", healthHandler.ReadinessCheck)

	// API路由（需要认证）
	api := s.router.Group("/api/v1")
	api.Use(s.authMiddleware.AuthRequired())
	api.Use(quotaMiddleware.CheckQuota())

	// 文件管理API
	files := api.Group("/files")
	{
		files.GET("", fileHandler.ListFiles)
		files.POST("", fileHandler.CreateFile)
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
		upload.GET("/status/:session_id", uploadHandler.GetUploadStatus)
		upload.GET("/sessions", uploadHandler.ListUploadSessions)
	}

	// 缩略图API
	thumbnails := api.Group("/thumbnails")
	{
		thumbnails.GET("/:file_id", thumbnailHandler.GetThumbnail)
		thumbnails.POST("/:file_id/generate", thumbnailHandler.GenerateThumbnail)
		thumbnails.GET("/:file_id/list", thumbnailHandler.ListThumbnails)
		thumbnails.DELETE("/:file_id", thumbnailHandler.DeleteThumbnail)
	}

	// 创建测试用户
	s.setupTestUsers()
}

func (s *IntegrationTestSuite) setupTestUsers() {
	// 创建普通用户
	testUserToken, err := s.authMiddleware.GenerateJWT(1, "testuser", "test@example.com", "user")
	s.NoError(err)
	s.testUser = &TestUser{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		Token:    testUserToken,
	}

	// 创建管理员用户
	adminUserToken, err := s.authMiddleware.GenerateJWT(2, "admin", "admin@example.com", "admin")
	s.NoError(err)
	s.adminUser = &TestUser{
		ID:       2,
		Username: "admin",
		Email:    "admin@example.com",
		Role:     "admin",
		Token:    adminUserToken,
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	// 清理数据库
	s.db.Exec("DELETE FROM files")
	s.db.Exec("DELETE FROM folders")
	s.db.Exec("DELETE FROM thumbnails")
	s.db.Exec("DELETE FROM upload_sessions")
}

// 辅助方法
func (s *IntegrationTestSuite) makeRequest(method, url string, body io.Reader, user *TestUser, contentType ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, body)
	if user != nil {
		req.Header.Set("Authorization", "Bearer "+user.Token)
	}
	if len(contentType) > 0 {
		req.Header.Set("Content-Type", contentType[0])
	}

	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)
	return resp
}

func (s *IntegrationTestSuite) makeJSONRequest(method, url string, data interface{}, user *TestUser) *httptest.ResponseRecorder {
	jsonData, _ := json.Marshal(data)
	return s.makeRequest(method, url, bytes.NewBuffer(jsonData), user, "application/json")
}

func (s *IntegrationTestSuite) makeMultipartRequest(method, url string, fields map[string]string, files map[string][]byte, user *TestUser) *httptest.ResponseRecorder {
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
	return s.makeRequest(method, url, body, user, writer.FormDataContentType())
}

// 测试用例

func (s *IntegrationTestSuite) TestHealthCheck() {
	// 测试健康检查
	resp := s.makeRequest("GET", "/health", nil, nil)
	s.Equal(http.StatusOK, resp.Code)

	var result map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &result)
	s.Equal("healthy", result["status"])
}

func (s *IntegrationTestSuite) TestCompleteFileUploadWorkflow() {
	// 完整的文件上传和管理工作流测试

	// 1. 创建文件夹
	folderData := map[string]interface{}{
		"name":        "integration-test-folder",
		"description": "Integration test folder",
	}
	resp := s.makeJSONRequest("POST", "/api/v1/folders", folderData, s.testUser)
	s.Equal(http.StatusCreated, resp.Code)

	var folderResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &folderResult)
	folderID := int(folderResult["data"].(map[string]interface{})["id"].(float64))

	// 2. 上传文件到文件夹
	fileContent := "This is an integration test file content"
	files := map[string][]byte{
		"integration-test.txt": []byte(fileContent),
	}
	fields := map[string]string{
		"folder_id": fmt.Sprintf("%d", folderID),
	}

	resp = s.makeMultipartRequest("POST", "/api/v1/upload/simple", fields, files, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	var uploadResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &uploadResult)
	fileID := int(uploadResult["data"].(map[string]interface{})["id"].(float64))

	// 3. 获取文件信息
	resp = s.makeRequest("GET", fmt.Sprintf("/api/v1/files/%d", fileID), nil, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	var fileResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &fileResult)
	fileData := fileResult["data"].(map[string]interface{})
	s.Equal("integration-test.txt", fileData["name"])
	s.Equal(float64(len(fileContent)), fileData["size"])

	// 4. 更新文件信息
	updateData := map[string]interface{}{
		"name":        "updated-integration-test.txt",
		"description": "Updated description",
		"tags":        []string{"integration", "test", "updated"},
	}

	resp = s.makeJSONRequest("PUT", fmt.Sprintf("/api/v1/files/%d", fileID), updateData, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	// 5. 验证文件更新
	resp = s.makeRequest("GET", fmt.Sprintf("/api/v1/files/%d", fileID), nil, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	json.Unmarshal(resp.Body.Bytes(), &fileResult)
	fileData = fileResult["data"].(map[string]interface{})
	s.Equal("updated-integration-test.txt", fileData["name"])
	s.Equal("Updated description", fileData["description"])

	// 6. 复制文件
	copyData := map[string]interface{}{
		"folder_id": nil, // 复制到根目录
	}

	resp = s.makeJSONRequest("POST", fmt.Sprintf("/api/v1/files/%d/copy", fileID), copyData, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	var copyResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &copyResult)
	copiedFileID := int(copyResult["data"].(map[string]interface{})["id"].(float64))

	// 7. 获取文件列表
	resp = s.makeRequest("GET", "/api/v1/files", nil, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	var listResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &listResult)
	files_list := listResult["data"].(map[string]interface{})["data"].([]interface{})
	s.GreaterOrEqual(len(files_list), 2) // 原文件和复制的文件

	// 8. 搜索文件
	resp = s.makeRequest("GET", "/api/v1/files/search?q=integration", nil, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	var searchResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &searchResult)
	searchFiles := searchResult["data"].(map[string]interface{})["data"].([]interface{})
	s.GreaterOrEqual(len(searchFiles), 1)

	// 9. 获取下载URL
	resp = s.makeRequest("GET", fmt.Sprintf("/api/v1/files/%d/download", fileID), nil, s.testUser)
	s.Equal(http.StatusFound, resp.Code)
	s.NotEmpty(resp.Header().Get("Location"))

	// 10. 删除复制的文件
	resp = s.makeRequest("DELETE", fmt.Sprintf("/api/v1/files/%d", copiedFileID), nil, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	// 11. 获取文件夹内容
	resp = s.makeRequest("GET", fmt.Sprintf("/api/v1/folders/%d/contents", folderID), nil, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	var contentsResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &contentsResult)
	contentsData := contentsResult["data"].(map[string]interface{})
	files_in_folder := contentsData["files"].([]interface{})
	s.Len(files_in_folder, 1) // 只剩一个原文件

	// 12. 删除文件夹（连同文件一起删除）
	resp = s.makeRequest("DELETE", fmt.Sprintf("/api/v1/folders/%d", folderID), nil, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	// 13. 验证文件也被删除
	resp = s.makeRequest("GET", fmt.Sprintf("/api/v1/files/%d", fileID), nil, s.testUser)
	s.Equal(http.StatusNotFound, resp.Code)
}

func (s *IntegrationTestSuite) TestChunkedUploadWorkflow() {
	// 分片上传完整工作流测试

	// 1. 初始化分片上传
	fileSize := int64(1024 * 5) // 5KB
	chunkSize := int64(1024)    // 1KB chunks
	initData := map[string]interface{}{
		"file_name":    "chunked-upload-test.txt",
		"file_size":    fileSize,
		"content_type": "text/plain",
		"chunk_size":   chunkSize,
	}

	resp := s.makeJSONRequest("POST", "/api/v1/upload/initiate", initData, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	var initResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &initResult)
	sessionID := initResult["data"].(map[string]interface{})["session_id"].(string)
	totalChunks := int(initResult["data"].(map[string]interface{})["total_chunks"].(float64))

	// 2. 获取上传状态
	resp = s.makeRequest("GET", fmt.Sprintf("/api/v1/upload/status/%s", sessionID), nil, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	var statusResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &statusResult)
	s.Equal("initiated", statusResult["data"].(map[string]interface{})["status"])

	// 3. 上传所有分片
	for i := 0; i < totalChunks; i++ {
		chunkContent := strings.Repeat(fmt.Sprintf("%d", i%10), int(chunkSize))
		chunkData := map[string]interface{}{
			"session_id":  sessionID,
			"chunk_index": i,
			"chunk_size":  chunkSize,
			"chunk_data":  chunkContent,
		}

		resp = s.makeJSONRequest("POST", "/api/v1/upload/chunk", chunkData, s.testUser)
		s.Equal(http.StatusOK, resp.Code)

		var chunkResult map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &chunkResult)
		s.True(chunkResult["data"].(map[string]interface{})["success"].(bool))
	}

	// 4. 获取上传进度
	resp = s.makeRequest("GET", fmt.Sprintf("/api/v1/upload/progress/%s", sessionID), nil, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	var progressResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &progressResult)
	progressData := progressResult["data"].(map[string]interface{})
	s.Equal(float64(100), progressData["percentage"]) // 100% 完成

	// 5. 完成上传
	completeData := map[string]interface{}{
		"session_id": sessionID,
	}

	resp = s.makeJSONRequest("POST", "/api/v1/upload/complete", completeData, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	var completeResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &completeResult)
	fileData := completeResult["data"].(map[string]interface{})["file"].(map[string]interface{})
	s.Equal("chunked-upload-test.txt", fileData["name"])
	s.Equal(float64(fileSize), fileData["size"])

	// 6. 验证会话已完成
	resp = s.makeRequest("GET", fmt.Sprintf("/api/v1/upload/status/%s", sessionID), nil, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	json.Unmarshal(resp.Body.Bytes(), &statusResult)
	s.Equal("completed", statusResult["data"].(map[string]interface{})["status"])
}

func (s *IntegrationTestSuite) TestBatchOperations() {
	// 批量操作测试

	// 1. 创建目标文件夹
	folderData := map[string]interface{}{
		"name": "batch-target-folder",
	}
	resp := s.makeJSONRequest("POST", "/api/v1/folders", folderData, s.testUser)
	s.Equal(http.StatusCreated, resp.Code)

	var folderResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &folderResult)
	targetFolderID := int(folderResult["data"].(map[string]interface{})["id"].(float64))

	// 2. 创建多个测试文件
	fileIDs := make([]int, 0, 3)
	for i := 1; i <= 3; i++ {
		fileContent := fmt.Sprintf("Batch test file %d content", i)
		files := map[string][]byte{
			fmt.Sprintf("batch-file-%d.txt", i): []byte(fileContent),
		}

		resp = s.makeMultipartRequest("POST", "/api/v1/upload/simple", map[string]string{}, files, s.testUser)
		s.Equal(http.StatusOK, resp.Code)

		var uploadResult map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &uploadResult)
		fileID := int(uploadResult["data"].(map[string]interface{})["id"].(float64))
		fileIDs = append(fileIDs, fileID)
	}

	// 3. 批量移动文件
	batchMoveData := map[string]interface{}{
		"file_ids": fileIDs,
		"action":   "move",
		"target":   targetFolderID,
	}

	resp = s.makeJSONRequest("POST", "/api/v1/files/batch", batchMoveData, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	var batchResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &batchResult)
	batchData := batchResult["data"].(map[string]interface{})
	s.Equal(float64(3), batchData["total"])
	s.Equal(float64(3), batchData["successful"])
	s.Equal(float64(0), batchData["failed"])

	// 4. 验证文件已移动到目标文件夹
	resp = s.makeRequest("GET", fmt.Sprintf("/api/v1/folders/%d/contents", targetFolderID), nil, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	var contentsResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &contentsResult)
	contentsData := contentsResult["data"].(map[string]interface{})
	files := contentsData["files"].([]interface{})
	s.Len(files, 3) // 三个文件都在目标文件夹中

	// 5. 批量删除文件
	batchDeleteData := map[string]interface{}{
		"file_ids": fileIDs,
		"action":   "delete",
	}

	resp = s.makeJSONRequest("POST", "/api/v1/files/batch", batchDeleteData, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	json.Unmarshal(resp.Body.Bytes(), &batchResult)
	batchData = batchResult["data"].(map[string]interface{})
	s.Equal(float64(3), batchData["successful"])

	// 6. 验证文件已被删除
	for _, fileID := range fileIDs {
		resp = s.makeRequest("GET", fmt.Sprintf("/api/v1/files/%d", fileID), nil, s.testUser)
		s.Equal(http.StatusNotFound, resp.Code)
	}
}

func (s *IntegrationTestSuite) TestUserIsolation() {
	// 用户隔离测试

	// 1. 用户1创建文件夹和文件
	folderData := map[string]interface{}{
		"name": "user1-private-folder",
	}
	resp := s.makeJSONRequest("POST", "/api/v1/folders", folderData, s.testUser)
	s.Equal(http.StatusCreated, resp.Code)

	var folderResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &folderResult)
	user1FolderID := int(folderResult["data"].(map[string]interface{})["id"].(float64))

	fileContent := "User1 private file content"
	files := map[string][]byte{
		"user1-private.txt": []byte(fileContent),
	}
	fields := map[string]string{
		"folder_id": fmt.Sprintf("%d", user1FolderID),
	}

	resp = s.makeMultipartRequest("POST", "/api/v1/upload/simple", fields, files, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	var uploadResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &uploadResult)
	user1FileID := int(uploadResult["data"].(map[string]interface{})["id"].(float64))

	// 2. 创建第三个用户token（模拟另一个用户）
	user2Token, err := s.authMiddleware.GenerateJWT(3, "user2", "user2@example.com", "user")
	s.NoError(err)
	user2 := &TestUser{
		ID:       3,
		Username: "user2",
		Email:    "user2@example.com",
		Role:     "user",
		Token:    user2Token,
	}

	// 3. 用户2尝试访问用户1的文件夹（应该失败）
	resp = s.makeRequest("GET", fmt.Sprintf("/api/v1/folders/%d", user1FolderID), nil, user2)
	s.Equal(http.StatusNotFound, resp.Code) // 返回404避免信息泄露

	// 4. 用户2尝试访问用户1的文件（应该失败）
	resp = s.makeRequest("GET", fmt.Sprintf("/api/v1/files/%d", user1FileID), nil, user2)
	s.Equal(http.StatusNotFound, resp.Code)

	// 5. 用户2创建自己的文件
	user2FileContent := "User2 private file content"
	files = map[string][]byte{
		"user2-private.txt": []byte(user2FileContent),
	}

	resp = s.makeMultipartRequest("POST", "/api/v1/upload/simple", map[string]string{}, files, user2)
	s.Equal(http.StatusOK, resp.Code)

	// 6. 验证用户只能看到自己的文件
	resp = s.makeRequest("GET", "/api/v1/files", nil, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	var user1FilesResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &user1FilesResult)
	user1Files := user1FilesResult["data"].(map[string]interface{})["data"].([]interface{})

	resp = s.makeRequest("GET", "/api/v1/files", nil, user2)
	s.Equal(http.StatusOK, resp.Code)

	var user2FilesResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &user2FilesResult)
	user2Files := user2FilesResult["data"].(map[string]interface{})["data"].([]interface{})

	// 验证文件列表不重叠
	s.GreaterOrEqual(len(user1Files), 1)
	s.GreaterOrEqual(len(user2Files), 1)
	
	// 验证用户1不能在文件列表中看到用户2的文件
	for _, file := range user1Files {
		fileData := file.(map[string]interface{})
		s.Equal(float64(s.testUser.ID), fileData["user_id"])
	}

	for _, file := range user2Files {
		fileData := file.(map[string]interface{})
		s.Equal(float64(user2.ID), fileData["user_id"])
	}
}

func (s *IntegrationTestSuite) TestAuthenticationAndAuthorization() {
	// 认证和授权测试

	// 1. 无token访问（应该失败）
	resp := s.makeRequest("GET", "/api/v1/files", nil, nil)
	s.Equal(http.StatusUnauthorized, resp.Code)

	// 2. 无效token访问（应该失败）
	invalidUser := &TestUser{
		Token: "invalid-token",
	}
	resp = s.makeRequest("GET", "/api/v1/files", nil, invalidUser)
	s.Equal(http.StatusUnauthorized, resp.Code)

	// 3. 有效token访问（应该成功）
	resp = s.makeRequest("GET", "/api/v1/files", nil, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	// 4. 管理员访问普通用户资源（通过管理员API）
	// 注意：这里需要在路由中定义管理员API端点
	// 暂时跳过，因为当前实现中没有管理员特殊API
}

func (s *IntegrationTestSuite) TestErrorHandling() {
	// 错误处理测试

	// 1. 访问不存在的文件
	resp := s.makeRequest("GET", "/api/v1/files/99999", nil, s.testUser)
	s.Equal(http.StatusNotFound, resp.Code)

	var errorResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &errorResult)
	s.Contains(errorResult, "error")

	// 2. 上传到不存在的文件夹
	files := map[string][]byte{
		"test.txt": []byte("test content"),
	}
	fields := map[string]string{
		"folder_id": "99999", // 不存在的文件夹
	}

	resp = s.makeMultipartRequest("POST", "/api/v1/upload/simple", fields, files, s.testUser)
	s.Equal(http.StatusBadRequest, resp.Code)

	// 3. 上传空文件
	files = map[string][]byte{
		"empty.txt": []byte(""),
	}

	resp = s.makeMultipartRequest("POST", "/api/v1/upload/simple", map[string]string{}, files, s.testUser)
	s.Equal(http.StatusBadRequest, resp.Code)

	// 4. 创建重名文件夹
	folderData := map[string]interface{}{
		"name": "duplicate-folder",
	}

	resp = s.makeJSONRequest("POST", "/api/v1/folders", folderData, s.testUser)
	s.Equal(http.StatusCreated, resp.Code)

	// 尝试创建同名文件夹
	resp = s.makeJSONRequest("POST", "/api/v1/folders", folderData, s.testUser)
	s.Equal(http.StatusConflict, resp.Code)

	// 5. 无效的JSON请求
	invalidJSON := strings.NewReader(`{"invalid": json}`)
	resp = s.makeRequest("POST", "/api/v1/folders", invalidJSON, s.testUser, "application/json")
	s.Equal(http.StatusBadRequest, resp.Code)
}

func (s *IntegrationTestSuite) TestConcurrentOperations() {
	// 并发操作测试

	const concurrentRequests = 10
	results := make(chan *httptest.ResponseRecorder, concurrentRequests)

	// 并发创建文件夹
	for i := 0; i < concurrentRequests; i++ {
		go func(index int) {
			folderData := map[string]interface{}{
				"name": fmt.Sprintf("concurrent-folder-%d", index),
			}
			resp := s.makeJSONRequest("POST", "/api/v1/folders", folderData, s.testUser)
			results <- resp
		}(i)
	}

	// 收集结果
	successCount := 0
	for i := 0; i < concurrentRequests; i++ {
		resp := <-results
		if resp.Code == http.StatusCreated {
			successCount++
		}
	}

	// 验证所有请求都成功
	s.Equal(concurrentRequests, successCount)

	// 验证数据库中确实有这些文件夹
	var count int64
	s.db.Model(&models.Folder{}).Where("name LIKE 'concurrent-folder-%'").Count(&count)
	s.Equal(int64(concurrentRequests), count)
}

func (s *IntegrationTestSuite) TestServiceIntegration() {
	// 服务间集成测试

	// 1. 上传图片文件
	imageContent := []byte("fake-image-content-for-testing")
	files := map[string][]byte{
		"test-image.jpg": imageContent,
	}

	resp := s.makeMultipartRequest("POST", "/api/v1/upload/simple", map[string]string{}, files, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	var uploadResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &uploadResult)
	fileID := int(uploadResult["data"].(map[string]interface{})["id"].(float64))

	// 2. 生成缩略图（如果支持）
	generateData := map[string]interface{}{
		"sizes": []string{"small", "medium"},
	}

	resp = s.makeJSONRequest("POST", fmt.Sprintf("/api/v1/thumbnails/%d/generate", fileID), generateData, s.testUser)
	// 可能返回202 Accepted（异步处理）或者直接成功
	s.True(resp.Code == http.StatusOK || resp.Code == http.StatusAccepted)

	// 3. 获取文件统计信息
	resp = s.makeRequest("GET", "/api/v1/files/stats", nil, s.testUser)
	s.Equal(http.StatusOK, resp.Code)

	var statsResult map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &statsResult)
	stats := statsResult["data"].(map[string]interface{})
	s.GreaterOrEqual(stats["file_count"], float64(1))
	s.GreaterOrEqual(stats["total_size"], float64(len(imageContent)))
}

// 性能测试
func (s *IntegrationTestSuite) TestPerformance() {
	// 创建大量文件测试性能
	const fileCount = 100
	start := time.Now()

	for i := 0; i < fileCount; i++ {
		fileContent := fmt.Sprintf("Performance test file %d content", i)
		files := map[string][]byte{
			fmt.Sprintf("perf-file-%d.txt", i): []byte(fileContent),
		}

		resp := s.makeMultipartRequest("POST", "/api/v1/upload/simple", map[string]string{}, files, s.testUser)
		s.Equal(http.StatusOK, resp.Code)
	}

	duration := time.Since(start)
	s.Less(duration, time.Second*30) // 100个文件上传应该在30秒内完成

	// 测试文件列表性能
	start = time.Now()
	resp := s.makeRequest("GET", "/api/v1/files?page_size=50", nil, s.testUser)
	duration = time.Since(start)
	s.Equal(http.StatusOK, resp.Code)
	s.Less(duration, time.Millisecond*500) // 文件列表应该在500ms内返回
}

func TestIntegrationSuite(t *testing.T) {
	// 检查是否设置了集成测试环境变量
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("跳过集成测试，设置 INTEGRATION_TESTS=1 环境变量来运行")
	}

	suite.Run(t, new(IntegrationTestSuite))
}