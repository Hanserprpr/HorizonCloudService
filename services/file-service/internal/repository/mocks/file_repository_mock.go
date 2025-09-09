package mocks

import (
	"context"
	"file-service/internal/models"
	"file-service/internal/repository"

	"github.com/stretchr/testify/mock"
)

// MockFileRepository 文件仓库模拟实现
type MockFileRepository struct {
	mock.Mock
}

// NewMockFileRepository 创建文件仓库模拟实例
func NewMockFileRepository() *MockFileRepository {
	return &MockFileRepository{}
}

// Create 创建文件记录
func (m *MockFileRepository) Create(ctx context.Context, file *models.File) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

// GetByID 根据ID获取文件
func (m *MockFileRepository) GetByID(ctx context.Context, id uint) (*models.File, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.File), args.Error(1)
}

// GetByHash 根据哈希值获取文件
func (m *MockFileRepository) GetByHash(ctx context.Context, hash string) (*models.File, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.File), args.Error(1)
}

// GetByPath 根据路径获取文件
func (m *MockFileRepository) GetByPath(ctx context.Context, path string) (*models.File, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.File), args.Error(1)
}

// Update 更新文件记录
func (m *MockFileRepository) Update(ctx context.Context, file *models.File) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

// Delete 删除文件记录
func (m *MockFileRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// List 获取文件列表
func (m *MockFileRepository) List(ctx context.Context, userID uint, folderID *uint, offset, limit int, filters *repository.FileFilters) ([]*models.File, int64, error) {
	args := m.Called(ctx, userID, folderID, offset, limit, filters)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.File), args.Get(1).(int64), args.Error(2)
}

// Search 搜索文件
func (m *MockFileRepository) Search(ctx context.Context, userID uint, keyword string, offset, limit int, filters *repository.FileFilters) ([]*models.File, int64, error) {
	args := m.Called(ctx, userID, keyword, offset, limit, filters)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.File), args.Get(1).(int64), args.Error(2)
}

// GetByCategory 根据分类获取文件
func (m *MockFileRepository) GetByCategory(ctx context.Context, userID uint, category string, offset, limit int) ([]*models.File, int64, error) {
	args := m.Called(ctx, userID, category, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.File), args.Get(1).(int64), args.Error(2)
}

// GetRecentFiles 获取最近文件
func (m *MockFileRepository) GetRecentFiles(ctx context.Context, userID uint, days int, limit int) ([]*models.File, error) {
	args := m.Called(ctx, userID, days, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.File), args.Error(1)
}

// GetFilesByFolderID 获取文件夹下的文件
func (m *MockFileRepository) GetFilesByFolderID(ctx context.Context, folderID uint, offset, limit int) ([]*models.File, int64, error) {
	args := m.Called(ctx, folderID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.File), args.Get(1).(int64), args.Error(2)
}

// MoveToFolder 移动文件到文件夹
func (m *MockFileRepository) MoveToFolder(ctx context.Context, fileID, folderID uint) error {
	args := m.Called(ctx, fileID, folderID)
	return args.Error(0)
}

// BatchMoveToFolder 批量移动文件到文件夹
func (m *MockFileRepository) BatchMoveToFolder(ctx context.Context, fileIDs []uint, folderID uint) error {
	args := m.Called(ctx, fileIDs, folderID)
	return args.Error(0)
}

// GetVersions 获取文件版本列表
func (m *MockFileRepository) GetVersions(ctx context.Context, parentID uint) ([]*models.File, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.File), args.Error(1)
}

// GetLatestVersion 获取最新版本
func (m *MockFileRepository) GetLatestVersion(ctx context.Context, parentID uint) (*models.File, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.File), args.Error(1)
}

// CreateVersion 创建文件版本
func (m *MockFileRepository) CreateVersion(ctx context.Context, file *models.File, parentID uint) error {
	args := m.Called(ctx, file, parentID)
	return args.Error(0)
}

// FindDuplicates 查找重复文件
func (m *MockFileRepository) FindDuplicates(ctx context.Context, hash string, userID uint) ([]*models.File, error) {
	args := m.Called(ctx, hash, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.File), args.Error(1)
}

// GetFilesByHashes 根据哈希值批量获取文件
func (m *MockFileRepository) GetFilesByHashes(ctx context.Context, hashes []string, userID uint) ([]*models.File, error) {
	args := m.Called(ctx, hashes, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.File), args.Error(1)
}

// GetUserFileCount 获取用户文件数量
func (m *MockFileRepository) GetUserFileCount(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// GetUserStorageUsed 获取用户存储使用量
func (m *MockFileRepository) GetUserStorageUsed(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// GetCategoryStats 获取分类统计
func (m *MockFileRepository) GetCategoryStats(ctx context.Context, userID uint) (map[string]int64, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

// GetStorageTierStats 获取存储层级统计
func (m *MockFileRepository) GetStorageTierStats(ctx context.Context, userID uint) (map[string]int64, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

// BatchCreate 批量创建文件
func (m *MockFileRepository) BatchCreate(ctx context.Context, files []*models.File) error {
	args := m.Called(ctx, files)
	return args.Error(0)
}

// BatchUpdate 批量更新文件
func (m *MockFileRepository) BatchUpdate(ctx context.Context, files []*models.File) error {
	args := m.Called(ctx, files)
	return args.Error(0)
}

// BatchDelete 批量删除文件
func (m *MockFileRepository) BatchDelete(ctx context.Context, ids []uint) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

// BatchUpdateStatus 批量更新状态
func (m *MockFileRepository) BatchUpdateStatus(ctx context.Context, ids []uint, status int) error {
	args := m.Called(ctx, ids, status)
	return args.Error(0)
}

// CleanupDeleted 清理已删除的文件
func (m *MockFileRepository) CleanupDeleted(ctx context.Context, beforeDays int) (int64, error) {
	args := m.Called(ctx, beforeDays)
	return args.Get(0).(int64), args.Error(1)
}

// CleanupOrphaned 清理孤立文件记录
func (m *MockFileRepository) CleanupOrphaned(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}