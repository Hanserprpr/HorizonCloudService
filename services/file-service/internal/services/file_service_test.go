package services

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"file-service/internal/models"
	"file-service/internal/repository"
	"file-service/internal/storage"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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

func (m *MockFileRepository) GetUserStats(ctx context.Context, userID uint) (*repository.UserFileStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.UserFileStats), args.Error(1)
}

func (m *MockFileRepository) GetDB() *gorm.DB {
	return nil
}

// MockStorage Mock存储
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Upload(ctx context.Context, path string, reader io.Reader, size int64) error {
	args := m.Called(ctx, path, reader, size)
	return args.Error(0)
}

func (m *MockStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockStorage) Delete(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *MockStorage) Exists(ctx context.Context, path string) (bool, error) {
	args := m.Called(ctx, path)
	return args.Bool(0), args.Error(1)
}

func (m *MockStorage) GetURL(ctx context.Context, path string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, path, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) GetPresignedURL(ctx context.Context, path string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, path, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) Copy(ctx context.Context, srcPath, dstPath string) error {
	args := m.Called(ctx, srcPath, dstPath)
	return args.Error(0)
}

func (m *MockStorage) Move(ctx context.Context, srcPath, dstPath string) error {
	args := m.Called(ctx, srcPath, dstPath)
	return args.Error(0)
}

func (m *MockStorage) GetInfo(ctx context.Context, path string) (*storage.FileInfo, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.FileInfo), args.Error(1)
}

// MockUserServiceClient Mock用户服务客户端
type MockUserServiceClient struct {
	mock.Mock
}

func (m *MockUserServiceClient) GetUser(ctx context.Context, userID uint) (*UserInfo, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserInfo), args.Error(1)
}

func (m *MockUserServiceClient) GetUserByEmail(ctx context.Context, email string) (*UserInfo, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserInfo), args.Error(1)
}

func (m *MockUserServiceClient) UpdateUserQuota(ctx context.Context, userID uint, storageQuota, fileCountQuota int64) error {
	args := m.Called(ctx, userID, storageQuota, fileCountQuota)
	return args.Error(0)
}

func (m *MockUserServiceClient) ValidateUserPermission(ctx context.Context, userID uint, resource, action string) (bool, error) {
	args := m.Called(ctx, userID, resource, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserServiceClient) GetUserRole(ctx context.Context, userID uint) (*UserRole, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserRole), args.Error(1)
}

func (m *MockUserServiceClient) UpdateUserStorageUsage(ctx context.Context, userID uint, usedStorage int64) error {
	args := m.Called(ctx, userID, usedStorage)
	return args.Error(0)
}

func (m *MockUserServiceClient) NotifyQuotaExceeded(ctx context.Context, userID uint, quotaType string) error {
	args := m.Called(ctx, userID, quotaType)
	return args.Error(0)
}

type FileServiceTestSuite struct {
	suite.Suite
	fileService   *fileService
	fileRepo      *MockFileRepository
	storage       *MockStorage
	userClient    *MockUserServiceClient
}

func (s *FileServiceTestSuite) SetupSuite() {
	s.fileRepo = new(MockFileRepository)
	s.storage = new(MockStorage)
	s.userClient = new(MockUserServiceClient)

	config := &ServiceConfig{
		Storage:          s.storage,
		UserService:      s.userClient,
		DefaultChunkSize: 1024 * 1024, // 1MB
	}

	repos := &repository.Repositories{
		File: s.fileRepo,
	}

	services := NewServices(repos, config)
	s.fileService = services.File.(*fileService)
}

func (s *FileServiceTestSuite) SetupTest() {
	// 重置Mock
	s.fileRepo = new(MockFileRepository)
	s.storage = new(MockStorage)
	s.userClient = new(MockUserServiceClient)

	s.fileService.repo = s.fileRepo
	s.fileService.storage = s.storage
	s.fileService.userService = s.userClient
}

func (s *FileServiceTestSuite) TestUploadFile() {
	ctx := context.Background()
	userID := uint(1)
	fileName := "test.txt"
	content := "test file content"
	reader := strings.NewReader(content)

	req := &UploadFileRequest{
		FileName:    fileName,
		Size:        int64(len(content)),
		ContentType: "text/plain",
		UserID:      userID,
		Reader:      reader,
	}

	// Mock存储上传
	s.storage.On("Upload", mock.Anything, mock.AnythingOfType("string"), mock.Anything, req.Size).Return(nil)

	// Mock文件创建
	s.fileRepo.On("Create", ctx, mock.AnythingOfType("*models.File")).Return(nil)

	// 执行测试
	result, err := s.fileService.UploadFile(ctx, req)

	// 验证结果
	s.NoError(err)
	s.NotNil(result)
	s.Equal(fileName, result.Name)
	s.Equal(req.Size, result.Size)
	s.Equal(userID, result.UserID)

	// 验证Mock调用
	s.storage.AssertExpectations(s.T())
	s.fileRepo.AssertExpectations(s.T())
}

func (s *FileServiceTestSuite) TestGetFile() {
	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)

	expectedFile := &models.File{
		BaseModel: models.BaseModel{ID: fileID},
		Name:      "test.txt",
		Size:      1024,
		UserID:    userID,
		Status:    models.FileStatusActive,
	}

	// Mock文件查询
	s.fileRepo.On("GetByID", ctx, fileID, userID).Return(expectedFile, nil)

	// 执行测试
	file, err := s.fileService.GetFile(ctx, fileID, userID)

	// 验证结果
	s.NoError(err)
	s.Equal(expectedFile, file)

	s.fileRepo.AssertExpectations(s.T())
}

func (s *FileServiceTestSuite) TestGetFileNotFound() {
	ctx := context.Background()
	userID := uint(1)
	fileID := uint(999)

	// Mock文件不存在
	s.fileRepo.On("GetByID", ctx, fileID, userID).Return(nil, NewServiceError(ErrorTypeResourceNotFound, "file not found", nil))

	// 执行测试
	file, err := s.fileService.GetFile(ctx, fileID, userID)

	// 验证结果
	s.Error(err)
	s.Nil(file)
	s.Contains(err.Error(), "file not found")

	s.fileRepo.AssertExpectations(s.T())
}

func (s *FileServiceTestSuite) TestListFiles() {
	ctx := context.Background()
	userID := uint(1)
	offset := 0
	limit := 20

	expectedFiles := []*models.File{
		{
			BaseModel: models.BaseModel{ID: 1},
			Name:      "file1.txt",
			UserID:    userID,
		},
		{
			BaseModel: models.BaseModel{ID: 2},
			Name:      "file2.txt",
			UserID:    userID,
		},
	}
	expectedTotal := int64(2)

	// Mock文件列表查询
	s.fileRepo.On("List", ctx, userID, (*uint)(nil), offset, limit, mock.AnythingOfType("*repository.FileFilters")).
		Return(expectedFiles, expectedTotal, nil)

	// 执行测试
	files, total, err := s.fileService.ListFiles(ctx, userID, nil, offset, limit, nil)

	// 验证结果
	s.NoError(err)
	s.Equal(expectedFiles, files)
	s.Equal(expectedTotal, total)

	s.fileRepo.AssertExpectations(s.T())
}

func (s *FileServiceTestSuite) TestUpdateFile() {
	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)

	existingFile := &models.File{
		BaseModel: models.BaseModel{ID: fileID},
		Name:      "old-name.txt",
		UserID:    userID,
		Status:    models.FileStatusActive,
	}

	req := &UpdateFileRequest{
		FileID:      fileID,
		UserID:      userID,
		Name:        "new-name.txt",
		Description: "Updated description",
	}

	// Mock查询现有文件
	s.fileRepo.On("GetByID", ctx, fileID, userID).Return(existingFile, nil)

	// Mock更新文件
	s.fileRepo.On("Update", ctx, mock.AnythingOfType("*models.File")).Return(nil)

	// 执行测试
	err := s.fileService.UpdateFile(ctx, req)

	// 验证结果
	s.NoError(err)

	s.fileRepo.AssertExpectations(s.T())
}

func (s *FileServiceTestSuite) TestDeleteFile() {
	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)

	existingFile := &models.File{
		BaseModel:   models.BaseModel{ID: fileID},
		Name:        "test.txt",
		UserID:      userID,
		StoragePath: "/test/path",
		Status:      models.FileStatusActive,
	}

	// Mock查询现有文件
	s.fileRepo.On("GetByID", ctx, fileID, userID).Return(existingFile, nil)

	// Mock软删除文件
	s.fileRepo.On("Update", ctx, mock.AnythingOfType("*models.File")).Return(nil)

	// Mock存储删除
	s.storage.On("Delete", ctx, existingFile.StoragePath).Return(nil)

	// 执行测试
	err := s.fileService.DeleteFile(ctx, fileID, userID)

	// 验证结果
	s.NoError(err)

	s.fileRepo.AssertExpectations(s.T())
	s.storage.AssertExpectations(s.T())
}

func (s *FileServiceTestSuite) TestGetDownloadURL() {
	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)

	existingFile := &models.File{
		BaseModel:   models.BaseModel{ID: fileID},
		Name:        "test.txt",
		UserID:      userID,
		StoragePath: "/test/path",
		Status:      models.FileStatusActive,
	}

	expectedURL := "https://example.com/download/test.txt"

	// Mock查询文件
	s.fileRepo.On("GetByID", ctx, fileID, userID).Return(existingFile, nil)

	// Mock获取下载URL
	s.storage.On("GetURL", ctx, existingFile.StoragePath, mock.AnythingOfType("time.Duration")).
		Return(expectedURL, nil)

	// 执行测试
	url, err := s.fileService.GetDownloadURL(ctx, fileID, userID)

	// 验证结果
	s.NoError(err)
	s.Equal(expectedURL, url)

	s.fileRepo.AssertExpectations(s.T())
	s.storage.AssertExpectations(s.T())
}

func (s *FileServiceTestSuite) TestSearchFiles() {
	ctx := context.Background()
	userID := uint(1)
	query := "test"
	offset := 0
	limit := 20

	expectedFiles := []*models.File{
		{
			BaseModel: models.BaseModel{ID: 1},
			Name:      "test-file.txt",
			UserID:    userID,
		},
	}
	expectedTotal := int64(1)

	// Mock搜索
	s.fileRepo.On("Search", ctx, userID, query, offset, limit, mock.AnythingOfType("*repository.FileFilters")).
		Return(expectedFiles, expectedTotal, nil)

	// 执行测试
	files, total, err := s.fileService.SearchFiles(ctx, userID, query, offset, limit, nil)

	// 验证结果
	s.NoError(err)
	s.Equal(expectedFiles, files)
	s.Equal(expectedTotal, total)

	s.fileRepo.AssertExpectations(s.T())
}

func (s *FileServiceTestSuite) TestGetUserStats() {
	ctx := context.Background()
	userID := uint(1)

	expectedStats := &repository.UserFileStats{
		FileCount:      10,
		TotalSize:      1024 * 1024, // 1MB
		DocumentCount:  5,
		ImageCount:     3,
		VideoCount:     1,
		OtherCount:     1,
	}

	// Mock统计查询
	s.fileRepo.On("GetUserStats", ctx, userID).Return(expectedStats, nil)

	// 执行测试
	stats, err := s.fileService.GetUserStats(ctx, userID)

	// 验证结果
	s.NoError(err)
	s.Equal(expectedStats, stats)

	s.fileRepo.AssertExpectations(s.T())
}

func (s *FileServiceTestSuite) TestMoveFile() {
	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)
	newFolderID := uint(200)

	existingFile := &models.File{
		BaseModel: models.BaseModel{ID: fileID},
		Name:      "test.txt",
		UserID:    userID,
		FolderID:  nil,
		Status:    models.FileStatusActive,
	}

	// Mock查询现有文件
	s.fileRepo.On("GetByID", ctx, fileID, userID).Return(existingFile, nil)

	// Mock更新文件
	s.fileRepo.On("Update", ctx, mock.AnythingOfType("*models.File")).Return(nil)

	// 执行测试
	err := s.fileService.MoveFile(ctx, fileID, newFolderID, userID)

	// 验证结果
	s.NoError(err)

	s.fileRepo.AssertExpectations(s.T())
}

func (s *FileServiceTestSuite) TestCopyFile() {
	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)
	newFolderID := uint(200)

	existingFile := &models.File{
		BaseModel:   models.BaseModel{ID: fileID},
		Name:        "test.txt",
		UserID:      userID,
		FolderID:    nil,
		StoragePath: "/original/path",
		Status:      models.FileStatusActive,
	}

	// Mock查询原文件
	s.fileRepo.On("GetByID", ctx, fileID, userID).Return(existingFile, nil)

	// Mock存储复制
	s.storage.On("Copy", ctx, existingFile.StoragePath, mock.AnythingOfType("string")).Return(nil)

	// Mock创建新文件记录
	s.fileRepo.On("Create", ctx, mock.AnythingOfType("*models.File")).Return(nil)

	// 执行测试
	copiedFile, err := s.fileService.CopyFile(ctx, fileID, newFolderID, userID)

	// 验证结果
	s.NoError(err)
	s.NotNil(copiedFile)
	s.Contains(copiedFile.Name, "copy")

	s.fileRepo.AssertExpectations(s.T())
	s.storage.AssertExpectations(s.T())
}

func TestFileService(t *testing.T) {
	suite.Run(t, new(FileServiceTestSuite))
}