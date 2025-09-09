package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"file-service/internal/models"
	"file-service/internal/repository"
	"file-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MockFileRepository Mock文件仓库
type MockFileRepository struct {
	mock.Mock
}

func (m *MockFileRepository) Create(ctx context.Context, file *models.File) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockFileRepository) GetByID(ctx context.Context, id uint, userID uint) (*models.File, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.File), args.Error(1)
}

func (m *MockFileRepository) GetUserStats(ctx context.Context, userID uint) (*repository.UserFileStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.UserFileStats), args.Error(1)
}

func (m *MockFileRepository) List(ctx context.Context, userID uint, folderID *uint, offset, limit int, filters *repository.FileFilters) ([]*models.File, int64, error) {
	args := m.Called(ctx, userID, folderID, offset, limit, filters)
	return args.Get(0).([]*models.File), args.Get(1).(int64), args.Error(2)
}

func (m *MockFileRepository) Update(ctx context.Context, file *models.File) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockFileRepository) Delete(ctx context.Context, id uint, userID uint) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockFileRepository) GetByHash(ctx context.Context, hash string, userID uint) (*models.File, error) {
	args := m.Called(ctx, hash, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.File), args.Error(1)
}

func (m *MockFileRepository) Search(ctx context.Context, userID uint, query string, offset, limit int, filters *repository.FileFilters) ([]*models.File, int64, error) {
	args := m.Called(ctx, userID, query, offset, limit, filters)
	return args.Get(0).([]*models.File), args.Get(1).(int64), args.Error(2)
}

func (m *MockFileRepository) FindDuplicates(ctx context.Context, userID uint, folderID *uint) (map[string][]*models.File, error) {
	args := m.Called(ctx, userID, folderID)
	return args.Get(0).(map[string][]*models.File), args.Error(1)
}

func (m *MockFileRepository) GetDB() *gorm.DB {
	// 为测试返回一个内存数据库
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	return db
}

type QuotaMiddlewareTestSuite struct {
	suite.Suite
	quotaMiddleware   *QuotaMiddleware
	fileRepo          *MockFileRepository
	userServiceClient *services.MockUserServiceClient
	engine            *gin.Engine
	authMiddleware    *AuthMiddleware
}

func (s *QuotaMiddlewareTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	// 创建Mock仓库和服务
	s.fileRepo = new(MockFileRepository)
	s.userServiceClient = services.NewMockUserServiceClient().(*services.MockUserServiceClient)

	// 创建配额中间件
	quotaConfig := &QuotaConfig{
		DefaultStorageQuota:  5 * 1024 * 1024, // 5MB for testing
		DefaultFileCount:     10,
		CheckInterval:        time.Minute,
		GraceBuffer:          0.1, // 10%
		EnableWarnings:       true,
		WarningThreshold:     0.8, // 80%
	}

	s.quotaMiddleware = NewQuotaMiddleware(quotaConfig, s.userServiceClient, s.fileRepo)

	// 创建认证中间件
	authConfig := &AuthConfig{
		JWTSecret:     []byte("test-secret"),
		JWTExpiration: time.Hour,
	}
	s.authMiddleware = NewAuthMiddleware(authConfig)

	// 创建测试引擎
	s.engine = gin.New()
}

func (s *QuotaMiddlewareTestSuite) SetupTest() {
	// 重置Mock
	s.fileRepo = new(MockFileRepository)
	s.quotaMiddleware.fileRepository = s.fileRepo
}

func (s *QuotaMiddlewareTestSuite) TestGetUserQuota() {
	ctx := context.Background()
	userID := uint(2)

	// Mock用户文件统计
	expectedStats := &repository.UserFileStats{
		FileCount: 5,
		TotalSize: 2 * 1024 * 1024, // 2MB
	}

	s.fileRepo.On("GetUserStats", ctx, userID).Return(expectedStats, nil)

	// 获取用户配额
	quota, err := s.quotaMiddleware.GetUserQuota(ctx, userID)
	s.NoError(err)
	s.NotNil(quota)

	// 验证配额信息
	s.Equal(userID, quota.UserID)
	s.Equal(int64(5*1024*1024*1024), quota.StorageQuota) // 用户服务返回的配额
	s.Equal(int64(5), quota.FileCount)
	s.Equal(int64(2*1024*1024), quota.StorageUsed)
	s.Equal(float64(2*1024*1024)/float64(5*1024*1024*1024), quota.StorageUsageRate)
	s.False(quota.IsStorageExceeded)
	s.False(quota.IsFileCountExceeded)

	s.fileRepo.AssertExpectations(s.T())
}

func (s *QuotaMiddlewareTestSuite) TestCheckQuotaMiddleware() {
	// 创建测试路由
	s.engine.Use(func(c *gin.Context) {
		// 模拟认证用户
		c.Set("user_id", uint(2))
		c.Next()
	})
	s.engine.Use(s.quotaMiddleware.CheckQuota())
	s.engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Mock正常配额情况
	expectedStats := &repository.UserFileStats{
		FileCount: 3,
		TotalSize: 1 * 1024 * 1024, // 1MB，在限额内
	}
	s.fileRepo.On("GetUserStats", mock.Anything, uint(2)).Return(expectedStats, nil)

	// 测试正常情况
	req := httptest.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()
	s.engine.ServeHTTP(resp, req)
	s.Equal(http.StatusOK, resp.Code)

	s.fileRepo.AssertExpectations(s.T())
}

func (s *QuotaMiddlewareTestSuite) TestCheckQuotaExceeded() {
	// 创建测试路由
	s.engine.Use(func(c *gin.Context) {
		c.Set("user_id", uint(2))
		c.Next()
	})
	s.engine.Use(s.quotaMiddleware.CheckQuota())
	s.engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Mock超出配额的情况
	expectedStats := &repository.UserFileStats{
		FileCount: 15, // 超出文件数量限制
		TotalSize: 6 * 1024 * 1024, // 6MB，超出存储限制
	}
	s.fileRepo.On("GetUserStats", mock.Anything, uint(2)).Return(expectedStats, nil)

	// 测试超出配额情况
	req := httptest.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()
	s.engine.ServeHTTP(resp, req)
	s.Equal(http.StatusForbidden, resp.Code)

	s.fileRepo.AssertExpectations(s.T())
}

func (s *QuotaMiddlewareTestSuite) TestCheckUploadQuota() {
	fileSize := int64(1024 * 1024) // 1MB

	// 创建测试路由
	s.engine.Use(func(c *gin.Context) {
		c.Set("user_id", uint(2))
		c.Next()
	})
	s.engine.Use(s.quotaMiddleware.CheckUploadQuota(fileSize))
	s.engine.POST("/upload", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "upload success"})
	})

	// Mock当前使用情况
	expectedStats := &repository.UserFileStats{
		FileCount: 3,
		TotalSize: 3 * 1024 * 1024, // 3MB，上传1MB后不会超出5MB限制
	}
	s.fileRepo.On("GetUserStats", mock.Anything, uint(2)).Return(expectedStats, nil)

	// 测试正常上传
	req := httptest.NewRequest("POST", "/upload", nil)
	resp := httptest.NewRecorder()
	s.engine.ServeHTTP(resp, req)
	s.Equal(http.StatusOK, resp.Code)

	s.fileRepo.AssertExpectations(s.T())
}

func (s *QuotaMiddlewareTestSuite) TestCheckUploadQuotaExceeded() {
	fileSize := int64(3 * 1024 * 1024) // 3MB

	// 创建测试路由
	s.engine.Use(func(c *gin.Context) {
		c.Set("user_id", uint(2))
		c.Next()
	})
	s.engine.Use(s.quotaMiddleware.CheckUploadQuota(fileSize))
	s.engine.POST("/upload", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "upload success"})
	})

	// Mock当前使用情况：上传后会超出限制
	expectedStats := &repository.UserFileStats{
		FileCount: 3,
		TotalSize: 4 * 1024 * 1024, // 4MB，上传3MB后会超出5MB限制
	}
	s.fileRepo.On("GetUserStats", mock.Anything, uint(2)).Return(expectedStats, nil)

	// 测试上传超出配额
	req := httptest.NewRequest("POST", "/upload", nil)
	resp := httptest.NewRecorder()
	s.engine.ServeHTTP(resp, req)
	s.Equal(http.StatusForbidden, resp.Code)

	// 验证响应内容
	s.Contains(resp.Body.String(), "quota")
	s.Contains(resp.Body.String(), "exceeded")

	s.fileRepo.AssertExpectations(s.T())
}

func (s *QuotaMiddlewareTestSuite) TestCheckBatchUploadQuota() {
	// 创建测试路由
	s.engine.Use(func(c *gin.Context) {
		c.Set("user_id", uint(2))
		c.Next()
	})
	s.engine.Use(s.quotaMiddleware.CheckBatchUploadQuota())
	s.engine.POST("/batch-upload", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "batch upload success"})
	})

	// Mock当前使用情况
	expectedStats := &repository.UserFileStats{
		FileCount: 3,
		TotalSize: 1 * 1024 * 1024, // 1MB
	}
	s.fileRepo.On("GetUserStats", mock.Anything, uint(2)).Return(expectedStats, nil)

	// 测试批量上传
	requestBody := `{
		"files": [
			{"size": 1048576},
			{"size": 2097152}
		]
	}`

	req := httptest.NewRequest("POST", "/batch-upload", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	s.engine.ServeHTTP(resp, req)
	s.Equal(http.StatusOK, resp.Code)

	s.fileRepo.AssertExpectations(s.T())
}

func (s *QuotaMiddlewareTestSuite) TestGetQuotaStatus() {
	// 创建测试路由
	s.engine.Use(func(c *gin.Context) {
		c.Set("user_id", uint(2))
		c.Next()
	})
	s.engine.GET("/quota/status", s.quotaMiddleware.GetQuotaStatus)

	// Mock用户统计
	expectedStats := &repository.UserFileStats{
		FileCount: 5,
		TotalSize: 2 * 1024 * 1024, // 2MB
	}
	s.fileRepo.On("GetUserStats", mock.Anything, uint(2)).Return(expectedStats, nil)

	// 测试获取配额状态
	req := httptest.NewRequest("GET", "/quota/status", nil)
	resp := httptest.NewRecorder()
	s.engine.ServeHTTP(resp, req)
	s.Equal(http.StatusOK, resp.Code)

	// 验证响应包含配额信息
	s.Contains(resp.Body.String(), "storage")
	s.Contains(resp.Body.String(), "file_count")
	s.Contains(resp.Body.String(), "usage_rate")

	s.fileRepo.AssertExpectations(s.T())
}

func (s *QuotaMiddlewareTestSuite) TestFormatFileSize() {
	testCases := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, tc := range testCases {
		result := FormatFileSize(tc.bytes)
		s.Equal(tc.expected, result, "Failed for input %d", tc.bytes)
	}
}

func (s *QuotaMiddlewareTestSuite) TestGetQuotaFromContext() {
	// 创建测试上下文
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	
	// 测试没有配额信息的情况
	quota, exists := GetQuotaFromContext(c)
	s.False(exists)
	s.Nil(quota)

	// 设置配额信息
	expectedQuota := &UserQuota{
		UserID:       1,
		StorageQuota: 5 * 1024 * 1024,
		StorageUsed:  2 * 1024 * 1024,
	}
	c.Set("user_quota", expectedQuota)

	// 测试有配额信息的情况
	quota, exists = GetQuotaFromContext(c)
	s.True(exists)
	s.Equal(expectedQuota, quota)
}

func TestQuotaMiddleware(t *testing.T) {
	suite.Run(t, new(QuotaMiddlewareTestSuite))
}