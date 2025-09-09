package services

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"file-service/internal/models"
	"file-service/internal/repository"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockUploadRepository Mock上传仓库
type MockUploadRepository struct {
	mock.Mock
}

func (m *MockUploadRepository) CreateSession(ctx context.Context, session *models.UploadSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockUploadRepository) GetSession(ctx context.Context, sessionID string, userID uint) (*models.UploadSession, error) {
	args := m.Called(ctx, sessionID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UploadSession), args.Error(1)
}

func (m *MockUploadRepository) UpdateSession(ctx context.Context, session *models.UploadSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockUploadRepository) DeleteSession(ctx context.Context, sessionID string, userID uint) error {
	args := m.Called(ctx, sessionID, userID)
	return args.Error(0)
}

func (m *MockUploadRepository) ListSessions(ctx context.Context, userID uint, status *int, offset, limit int) ([]*models.UploadSession, int64, error) {
	args := m.Called(ctx, userID, status, offset, limit)
	return args.Get(0).([]*models.UploadSession), args.Get(1).(int64), args.Error(2)
}

func (m *MockUploadRepository) GetUserUploadStats(ctx context.Context, userID uint) (*repository.UserUploadStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.UserUploadStats), args.Error(1)
}

type UploadServiceTestSuite struct {
	suite.Suite
	uploadService *uploadService
	uploadRepo    *MockUploadRepository
	fileRepo      *MockFileRepository
	storage       *MockStorage
	userClient    *MockUserServiceClient
}

func (s *UploadServiceTestSuite) SetupSuite() {
	s.uploadRepo = new(MockUploadRepository)
	s.fileRepo = new(MockFileRepository)
	s.storage = new(MockStorage)
	s.userClient = new(MockUserServiceClient)

	config := &ServiceConfig{
		Storage:          s.storage,
		UserService:      s.userClient,
		DefaultChunkSize: 1024 * 1024, // 1MB
	}

	repos := &repository.Repositories{
		Upload: s.uploadRepo,
		File:   s.fileRepo,
	}

	services := NewServices(repos, config)
	s.uploadService = services.Upload.(*uploadService)
}

func (s *UploadServiceTestSuite) SetupTest() {
	// 重置Mock
	s.uploadRepo = new(MockUploadRepository)
	s.fileRepo = new(MockFileRepository)
	s.storage = new(MockStorage)
	s.userClient = new(MockUserServiceClient)

	s.uploadService.repo = s.uploadRepo
	s.uploadService.fileRepo = s.fileRepo
	s.uploadService.storage = s.storage
	s.uploadService.userService = s.userClient
}

func (s *UploadServiceTestSuite) TestInitiateUpload() {
	ctx := context.Background()
	userID := uint(1)

	req := &InitiateUploadRequest{
		FileName:    "test.txt",
		Size:        1024 * 1024, // 1MB
		ContentType: "text/plain",
		UserID:      userID,
		ChunkSize:   256 * 1024, // 256KB
	}

	// Mock创建上传会话
	s.uploadRepo.On("CreateSession", ctx, mock.AnythingOfType("*models.UploadSession")).
		Return(nil)

	// 执行测试
	result, err := s.uploadService.InitiateUpload(ctx, req)

	// 验证结果
	s.NoError(err)
	s.NotNil(result)
	s.NotEmpty(result.SessionID)
	s.Equal(req.FileName, result.FileName)
	s.Equal(req.Size, result.FileSize)
	s.Equal(models.UploadStatusInitiated, result.Status)
	s.Greater(result.TotalChunks, 0)

	s.uploadRepo.AssertExpectations(s.T())
}

func (s *UploadServiceTestSuite) TestUploadChunk() {
	ctx := context.Background()
	userID := uint(1)
	sessionID := "test-session-123"
	chunkIndex := 0
	chunkData := strings.NewReader("test chunk data")

	session := &models.UploadSession{
		SessionID:   sessionID,
		FileName:    "test.txt",
		FileSize:    1024,
		ChunkSize:   256,
		TotalChunks: 4,
		UserID:      userID,
		Status:      models.UploadStatusInitiated,
		UploadedChunks: []int{},
	}

	req := &UploadChunkRequest{
		SessionID:  sessionID,
		ChunkIndex: chunkIndex,
		ChunkSize:  256,
		ChunkData:  chunkData,
		UserID:     userID,
	}

	// Mock查询会话
	s.uploadRepo.On("GetSession", ctx, sessionID, userID).Return(session, nil)

	// Mock存储上传
	s.storage.On("Upload", ctx, mock.AnythingOfType("string"), chunkData, int64(256)).Return(nil)

	// Mock更新会话
	s.uploadRepo.On("UpdateSession", ctx, mock.AnythingOfType("*models.UploadSession")).Return(nil)

	// 执行测试
	result, err := s.uploadService.UploadChunk(ctx, req)

	// 验证结果
	s.NoError(err)
	s.NotNil(result)
	s.Equal(sessionID, result.SessionID)
	s.Equal(chunkIndex, result.ChunkIndex)
	s.True(result.Success)

	s.uploadRepo.AssertExpectations(s.T())
	s.storage.AssertExpectations(s.T())
}

func (s *UploadServiceTestSuite) TestCompleteUpload() {
	ctx := context.Background()
	userID := uint(1)
	sessionID := "test-session-123"

	session := &models.UploadSession{
		SessionID:      sessionID,
		FileName:       "test.txt",
		FileSize:       1024,
		ChunkSize:      256,
		TotalChunks:    4,
		UserID:         userID,
		Status:         models.UploadStatusUploading,
		UploadedChunks: []int{0, 1, 2, 3}, // 所有分片都已上传
		StoragePath:    "/uploads/test.txt",
	}

	// Mock查询会话
	s.uploadRepo.On("GetSession", ctx, sessionID, userID).Return(session, nil)

	// Mock合并分片
	s.storage.On("Upload", ctx, session.StoragePath, mock.AnythingOfType("*io.PipeReader"), session.FileSize).Return(nil)

	// Mock创建文件记录
	s.fileRepo.On("Create", ctx, mock.AnythingOfType("*models.File")).Return(nil)

	// Mock更新会话状态
	s.uploadRepo.On("UpdateSession", ctx, mock.AnythingOfType("*models.UploadSession")).Return(nil)

	// 执行测试
	result, err := s.uploadService.CompleteUpload(ctx, sessionID, userID)

	// 验证结果
	s.NoError(err)
	s.NotNil(result)
	s.NotNil(result.File)
	s.Equal(session.FileName, result.File.Name)
	s.Equal(session.FileSize, result.File.Size)

	s.uploadRepo.AssertExpectations(s.T())
	s.storage.AssertExpectations(s.T())
	s.fileRepo.AssertExpectations(s.T())
}

func (s *UploadServiceTestSuite) TestCompleteUploadIncomplete() {
	ctx := context.Background()
	userID := uint(1)
	sessionID := "test-session-123"

	session := &models.UploadSession{
		SessionID:      sessionID,
		FileName:       "test.txt",
		FileSize:       1024,
		ChunkSize:      256,
		TotalChunks:    4,
		UserID:         userID,
		Status:         models.UploadStatusUploading,
		UploadedChunks: []int{0, 1, 2}, // 缺少分片3
	}

	// Mock查询会话
	s.uploadRepo.On("GetSession", ctx, sessionID, userID).Return(session, nil)

	// 执行测试
	result, err := s.uploadService.CompleteUpload(ctx, sessionID, userID)

	// 验证结果
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "incomplete")

	s.uploadRepo.AssertExpectations(s.T())
}

func (s *UploadServiceTestSuite) TestAbortUpload() {
	ctx := context.Background()
	userID := uint(1)
	sessionID := "test-session-123"

	session := &models.UploadSession{
		SessionID:      sessionID,
		FileName:       "test.txt",
		UserID:         userID,
		Status:         models.UploadStatusUploading,
		UploadedChunks: []int{0, 1},
	}

	// Mock查询会话
	s.uploadRepo.On("GetSession", ctx, sessionID, userID).Return(session, nil)

	// Mock清理分片
	s.storage.On("Delete", ctx, mock.AnythingOfType("string")).Return(nil).Maybe()

	// Mock删除会话
	s.uploadRepo.On("DeleteSession", ctx, sessionID, userID).Return(nil)

	// 执行测试
	err := s.uploadService.AbortUpload(ctx, sessionID, userID)

	// 验证结果
	s.NoError(err)

	s.uploadRepo.AssertExpectations(s.T())
}

func (s *UploadServiceTestSuite) TestGetUploadSession() {
	ctx := context.Background()
	userID := uint(1)
	sessionID := "test-session-123"

	expectedSession := &models.UploadSession{
		SessionID:   sessionID,
		FileName:    "test.txt",
		FileSize:    1024,
		UserID:      userID,
		Status:      models.UploadStatusUploading,
	}

	// Mock查询会话
	s.uploadRepo.On("GetSession", ctx, sessionID, userID).Return(expectedSession, nil)

	// 执行测试
	session, err := s.uploadService.GetUploadSession(ctx, sessionID, userID)

	// 验证结果
	s.NoError(err)
	s.Equal(expectedSession, session)

	s.uploadRepo.AssertExpectations(s.T())
}

func (s *UploadServiceTestSuite) TestGetUploadProgress() {
	ctx := context.Background()
	userID := uint(1)
	sessionID := "test-session-123"

	session := &models.UploadSession{
		SessionID:      sessionID,
		FileName:       "test.txt",
		FileSize:       1024,
		TotalChunks:    4,
		UserID:         userID,
		Status:         models.UploadStatusUploading,
		UploadedChunks: []int{0, 1, 2}, // 3/4 chunks uploaded
	}

	// Mock查询会话
	s.uploadRepo.On("GetSession", ctx, sessionID, userID).Return(session, nil)

	// 执行测试
	progress, err := s.uploadService.GetUploadProgress(ctx, sessionID, userID)

	// 验证结果
	s.NoError(err)
	s.NotNil(progress)
	s.Equal(sessionID, progress.SessionID)
	s.Equal(4, progress.TotalChunks)
	s.Equal(3, progress.UploadedChunks)
	s.Equal(float64(75), progress.Percentage) // 3/4 * 100%

	s.uploadRepo.AssertExpectations(s.T())
}

func (s *UploadServiceTestSuite) TestListUploadSessions() {
	ctx := context.Background()
	userID := uint(1)
	offset := 0
	limit := 20

	expectedSessions := []*models.UploadSession{
		{
			SessionID: "session-1",
			FileName:  "file1.txt",
			UserID:    userID,
		},
		{
			SessionID: "session-2",
			FileName:  "file2.txt",
			UserID:    userID,
		},
	}
	expectedTotal := int64(2)

	// Mock查询会话列表
	s.uploadRepo.On("ListSessions", ctx, userID, (*int)(nil), offset, limit).
		Return(expectedSessions, expectedTotal, nil)

	// 执行测试
	sessions, total, err := s.uploadService.ListUploadSessions(ctx, userID, nil, offset, limit)

	// 验证结果
	s.NoError(err)
	s.Equal(expectedSessions, sessions)
	s.Equal(expectedTotal, total)

	s.uploadRepo.AssertExpectations(s.T())
}

func (s *UploadServiceTestSuite) TestBatchInitiateUpload() {
	ctx := context.Background()
	userID := uint(1)

	files := []*InitiateUploadRequest{
		{
			FileName:    "file1.txt",
			Size:        1024,
			ContentType: "text/plain",
			UserID:      userID,
		},
		{
			FileName:    "file2.txt",
			Size:        2048,
			ContentType: "text/plain",
			UserID:      userID,
		},
	}

	req := &BatchInitiateUploadRequest{
		Files:  files,
		UserID: userID,
	}

	// Mock创建多个上传会话
	s.uploadRepo.On("CreateSession", ctx, mock.AnythingOfType("*models.UploadSession")).
		Return(nil).Twice()

	// 执行测试
	result, err := s.uploadService.BatchInitiateUpload(ctx, req)

	// 验证结果
	s.NoError(err)
	s.NotNil(result)
	s.Len(result.Sessions, 2)
	s.Equal(2, result.Total)
	s.Equal(2, result.Successful)
	s.Equal(0, result.Failed)

	s.uploadRepo.AssertExpectations(s.T())
}

func (s *UploadServiceTestSuite) TestResumeUpload() {
	ctx := context.Background()
	userID := uint(1)
	sessionID := "test-session-123"

	session := &models.UploadSession{
		SessionID:   sessionID,
		FileName:    "test.txt",
		UserID:      userID,
		Status:      models.UploadStatusPaused,
	}

	// Mock查询会话
	s.uploadRepo.On("GetSession", ctx, sessionID, userID).Return(session, nil)

	// Mock更新会话状态
	s.uploadRepo.On("UpdateSession", ctx, mock.AnythingOfType("*models.UploadSession")).Return(nil)

	// 执行测试
	result, err := s.uploadService.ResumeUpload(ctx, sessionID, userID)

	// 验证结果
	s.NoError(err)
	s.NotNil(result)
	s.Equal(sessionID, result.SessionID)

	s.uploadRepo.AssertExpectations(s.T())
}

func (s *UploadServiceTestSuite) TestPauseUpload() {
	ctx := context.Background()
	userID := uint(1)
	sessionID := "test-session-123"

	session := &models.UploadSession{
		SessionID:   sessionID,
		FileName:    "test.txt",
		UserID:      userID,
		Status:      models.UploadStatusUploading,
	}

	// Mock查询会话
	s.uploadRepo.On("GetSession", ctx, sessionID, userID).Return(session, nil)

	// Mock更新会话状态
	s.uploadRepo.On("UpdateSession", ctx, mock.AnythingOfType("*models.UploadSession")).Return(nil)

	// 执行测试
	err := s.uploadService.PauseUpload(ctx, sessionID, userID)

	// 验证结果
	s.NoError(err)

	s.uploadRepo.AssertExpectations(s.T())
}

func (s *UploadServiceTestSuite) TestGetUploadStatistics() {
	ctx := context.Background()
	userID := uint(1)

	expectedStats := &repository.UserUploadStats{
		TotalSessions:     10,
		CompletedSessions: 8,
		FailedSessions:    1,
		PausedSessions:    1,
		TotalBytes:        1024 * 1024 * 100, // 100MB
		CompletedBytes:    1024 * 1024 * 80,  // 80MB
	}

	// Mock统计查询
	s.uploadRepo.On("GetUserUploadStats", ctx, userID).Return(expectedStats, nil)

	// 执行测试
	stats, err := s.uploadService.GetUploadStatistics(ctx, userID)

	// 验证结果
	s.NoError(err)
	s.Equal(expectedStats, stats)

	s.uploadRepo.AssertExpectations(s.T())
}

func (s *UploadServiceTestSuite) TestGenerateSessionID() {
	// 测试会话ID生成
	sessionID := s.uploadService.generateSessionID()
	s.NotEmpty(sessionID)
	s.Len(sessionID, 32) // 应该是32位的UUID

	// 测试唯一性
	sessionID2 := s.uploadService.generateSessionID()
	s.NotEqual(sessionID, sessionID2)
}

func (s *UploadServiceTestSuite) TestCalculateChunkPath() {
	sessionID := "test-session-123"
	chunkIndex := 5

	path := s.uploadService.calculateChunkPath(sessionID, chunkIndex)
	s.Contains(path, sessionID)
	s.Contains(path, "5")
	s.HasPrefix(path, "/chunks/")
}

func (s *UploadServiceTestSuite) TestIsUploadComplete() {
	// 测试完整上传
	uploadedChunks := []int{0, 1, 2, 3}
	totalChunks := 4
	s.True(s.uploadService.isUploadComplete(uploadedChunks, totalChunks))

	// 测试不完整上传
	uploadedChunks = []int{0, 1, 3}
	s.False(s.uploadService.isUploadComplete(uploadedChunks, totalChunks))

	// 测试空上传
	uploadedChunks = []int{}
	s.False(s.uploadService.isUploadComplete(uploadedChunks, totalChunks))
}

func TestUploadService(t *testing.T) {
	suite.Run(t, new(UploadServiceTestSuite))
}