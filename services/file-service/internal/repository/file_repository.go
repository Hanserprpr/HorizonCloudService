package repository

import (
	"context"
	"file-service/internal/models"

	"gorm.io/gorm"
)

// UserStorageInfo 用户存储信息
type UserStorageInfo struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	StorageUsed int64  `json:"storage_used"`
}

// FileRepository 文件仓库接口
type FileRepository interface {
	// 基本CRUD操作
	Create(ctx context.Context, file *models.File) error
	GetByID(ctx context.Context, id uint) (*models.File, error)
	GetByHash(ctx context.Context, hash string) (*models.File, error)
	GetByPath(ctx context.Context, path string) (*models.File, error)
	Update(ctx context.Context, file *models.File) error
	Delete(ctx context.Context, id uint) error
	
	// 查询操作
	List(ctx context.Context, userID uint, folderID *uint, offset, limit int, filters *FileFilters) ([]*models.File, int64, error)
	Search(ctx context.Context, userID uint, keyword string, offset, limit int, filters *FileFilters) ([]*models.File, int64, error)
	GetByCategory(ctx context.Context, userID uint, category string, offset, limit int) ([]*models.File, int64, error)
	GetRecentFiles(ctx context.Context, userID uint, days int, limit int) ([]*models.File, error)
	
	// 文件夹相关
	GetFilesByFolderID(ctx context.Context, folderID uint, offset, limit int) ([]*models.File, int64, error)
	MoveToFolder(ctx context.Context, fileID, folderID uint) error
	BatchMoveToFolder(ctx context.Context, fileIDs []uint, folderID uint) error
	
	// 版本控制
	GetVersions(ctx context.Context, parentID uint) ([]*models.File, error)
	GetLatestVersion(ctx context.Context, parentID uint) (*models.File, error)
	CreateVersion(ctx context.Context, file *models.File, parentID uint) error
	
	// 去重相关
	FindDuplicates(ctx context.Context, hash string, userID uint) ([]*models.File, error)
	GetFilesByHashes(ctx context.Context, hashes []string, userID uint) ([]*models.File, error)
	
	// 统计相关
	GetUserFileCount(ctx context.Context, userID uint) (int64, error)
	GetUserStorageUsed(ctx context.Context, userID uint) (int64, error)
	GetCategoryStats(ctx context.Context, userID uint) (map[string]int64, error)
	GetStorageTierStats(ctx context.Context, userID uint) (map[string]int64, error)
	
	// 管理员级别统计
	GetTotalFileCount(ctx context.Context) (int64, error)
	GetTotalStorageUsed(ctx context.Context) (int64, error)
	GetGlobalCategoryStats(ctx context.Context) (map[string]int64, error)
	GetUserStorageList(ctx context.Context) ([]UserStorageInfo, error)
	
	// 批量操作
	BatchCreate(ctx context.Context, files []*models.File) error
	BatchUpdate(ctx context.Context, files []*models.File) error
	BatchDelete(ctx context.Context, ids []uint) error
	BatchUpdateStatus(ctx context.Context, ids []uint, status int) error
	
	// 清理操作
	CleanupDeleted(ctx context.Context, beforeDays int) (int64, error)
	CleanupOrphaned(ctx context.Context) (int64, error)
}

// FileFilters 文件过滤器
type FileFilters struct {
	Category     string   `json:"category,omitempty"`
	ContentType  string   `json:"content_type,omitempty"`
	MinSize      *int64   `json:"min_size,omitempty"`
	MaxSize      *int64   `json:"max_size,omitempty"`
	Extensions   []string `json:"extensions,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Status       *int     `json:"status,omitempty"`
	StorageTier  string   `json:"storage_tier,omitempty"`
	DateFrom     *string  `json:"date_from,omitempty"`
	DateTo       *string  `json:"date_to,omitempty"`
	SortBy       string   `json:"sort_by,omitempty"`      // name/size/created_at/updated_at
	SortOrder    string   `json:"sort_order,omitempty"`   // asc/desc
}

// fileRepository 文件仓库实现
type fileRepository struct {
	db *gorm.DB
}

// NewFileRepository 创建文件仓库实例
func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

// Create 创建文件记录
func (r *fileRepository) Create(ctx context.Context, file *models.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

// GetByID 根据ID获取文件
func (r *fileRepository) GetByID(ctx context.Context, id uint) (*models.File, error) {
	var file models.File
	err := r.db.WithContext(ctx).
		Preload("Folder").
		Preload("Thumbnails").
		First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetByHash 根据哈希值获取文件
func (r *fileRepository) GetByHash(ctx context.Context, hash string) (*models.File, error) {
	var file models.File
	err := r.db.WithContext(ctx).
		Where("hash = ?", hash).
		First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetByPath 根据路径获取文件
func (r *fileRepository) GetByPath(ctx context.Context, path string) (*models.File, error) {
	var file models.File
	err := r.db.WithContext(ctx).
		Where("path = ?", path).
		First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// Update 更新文件记录
func (r *fileRepository) Update(ctx context.Context, file *models.File) error {
	return r.db.WithContext(ctx).Save(file).Error
}

// Delete 删除文件记录(软删除)
func (r *fileRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.File{}, id).Error
}

// List 获取文件列表
func (r *fileRepository) List(ctx context.Context, userID uint, folderID *uint, offset, limit int, filters *FileFilters) ([]*models.File, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.File{}).Where("user_id = ?", userID)
	
	// 文件夹过滤
	if folderID != nil {
		query = query.Where("folder_id = ?", *folderID)
	} else {
		query = query.Where("folder_id IS NULL")
	}
	
	// 应用过滤器
	query = r.applyFilters(query, filters)
	
	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 应用排序
	query = r.applySorting(query, filters)
	
	// 分页查询
	var files []*models.File
	err := query.Preload("Folder").
		Preload("Thumbnails").
		Offset(offset).
		Limit(limit).
		Find(&files).Error
	
	return files, total, err
}

// Search 搜索文件
func (r *fileRepository) Search(ctx context.Context, userID uint, keyword string, offset, limit int, filters *FileFilters) ([]*models.File, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.File{}).Where("user_id = ?", userID)
	
	// 全文搜索 - 兼容SQLite和PostgreSQL
	if keyword != "" {
		// 使用LIKE搜索结合LOWER函数实现不区分大小写搜索（SQLite兼容）
		query = query.Where(
			"LOWER(name) LIKE LOWER(?) OR LOWER(original_name) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?)",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	
	// 应用过滤器
	query = r.applyFilters(query, filters)
	
	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 应用排序
	query = r.applySorting(query, filters)
	
	// 分页查询
	var files []*models.File
	err := query.Preload("Folder").
		Preload("Thumbnails").
		Offset(offset).
		Limit(limit).
		Find(&files).Error
	
	return files, total, err
}

// GetByCategory 根据分类获取文件
func (r *fileRepository) GetByCategory(ctx context.Context, userID uint, category string, offset, limit int) ([]*models.File, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.File{}).
		Where("user_id = ? AND category = ?", userID, category)
	
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	var files []*models.File
	err := query.Preload("Folder").
		Preload("Thumbnails").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&files).Error
	
	return files, total, err
}

// GetRecentFiles 获取最近文件
func (r *fileRepository) GetRecentFiles(ctx context.Context, userID uint, days int, limit int) ([]*models.File, error) {
	var files []*models.File
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at >= NOW() - INTERVAL ? DAY", userID, days).
		Preload("Folder").
		Preload("Thumbnails").
		Order("created_at DESC").
		Limit(limit).
		Find(&files).Error
	
	return files, err
}

// GetFilesByFolderID 获取文件夹下的文件
func (r *fileRepository) GetFilesByFolderID(ctx context.Context, folderID uint, offset, limit int) ([]*models.File, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.File{}).Where("folder_id = ?", folderID)
	
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	var files []*models.File
	err := query.Preload("Thumbnails").
		Order("name ASC").
		Offset(offset).
		Limit(limit).
		Find(&files).Error
	
	return files, total, err
}

// MoveToFolder 移动文件到文件夹
func (r *fileRepository) MoveToFolder(ctx context.Context, fileID, folderID uint) error {
	return r.db.WithContext(ctx).
		Model(&models.File{}).
		Where("id = ?", fileID).
		Update("folder_id", folderID).Error
}

// BatchMoveToFolder 批量移动文件到文件夹
func (r *fileRepository) BatchMoveToFolder(ctx context.Context, fileIDs []uint, folderID uint) error {
	return r.db.WithContext(ctx).
		Model(&models.File{}).
		Where("id IN ?", fileIDs).
		Update("folder_id", folderID).Error
}

// 应用过滤器
func (r *fileRepository) applyFilters(query *gorm.DB, filters *FileFilters) *gorm.DB {
	if filters == nil {
		return query
	}
	
	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}
	
	if filters.ContentType != "" {
		query = query.Where("content_type = ?", filters.ContentType)
	}
	
	if filters.MinSize != nil {
		query = query.Where("size >= ?", *filters.MinSize)
	}
	
	if filters.MaxSize != nil {
		query = query.Where("size <= ?", *filters.MaxSize)
	}
	
	if len(filters.Extensions) > 0 {
		query = query.Where("extension IN ?", filters.Extensions)
	}
	
	if len(filters.Tags) > 0 {
		for _, tag := range filters.Tags {
			query = query.Where("? = ANY(tags)", tag)
		}
	}
	
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	
	if filters.StorageTier != "" {
		query = query.Where("storage_tier = ?", filters.StorageTier)
	}
	
	if filters.DateFrom != nil {
		query = query.Where("created_at >= ?", *filters.DateFrom)
	}
	
	if filters.DateTo != nil {
		query = query.Where("created_at <= ?", *filters.DateTo)
	}
	
	return query
}

// 应用排序
func (r *fileRepository) applySorting(query *gorm.DB, filters *FileFilters) *gorm.DB {
	if filters == nil {
		return query.Order("created_at DESC")
	}
	
	sortBy := "created_at"
	if filters.SortBy != "" {
		switch filters.SortBy {
		case "name", "size", "created_at", "updated_at":
			sortBy = filters.SortBy
		}
	}
	
	sortOrder := "DESC"
	if filters.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	
	return query.Order(sortBy + " " + sortOrder)
}

// GetVersions 获取文件版本列表
func (r *fileRepository) GetVersions(ctx context.Context, parentID uint) ([]*models.File, error) {
	var versions []*models.File
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("version DESC").
		Find(&versions).Error
	return versions, err
}

// GetLatestVersion 获取最新版本
func (r *fileRepository) GetLatestVersion(ctx context.Context, parentID uint) (*models.File, error) {
	var file models.File
	err := r.db.WithContext(ctx).
		Where("parent_id = ? AND is_latest = true", parentID).
		First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// CreateVersion 创建文件版本
func (r *fileRepository) CreateVersion(ctx context.Context, file *models.File, parentID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 设置旧版本为非最新
		if err := tx.Model(&models.File{}).
			Where("parent_id = ? AND is_latest = true", parentID).
			Update("is_latest", false).Error; err != nil {
			return err
		}
		
		// 创建新版本
		file.ParentID = &parentID
		file.IsLatest = true
		return tx.Create(file).Error
	})
}

// FindDuplicates 查找重复文件
func (r *fileRepository) FindDuplicates(ctx context.Context, hash string, userID uint) ([]*models.File, error) {
	var files []*models.File
	err := r.db.WithContext(ctx).
		Where("hash = ? AND user_id = ?", hash, userID).
		Find(&files).Error
	return files, err
}

// GetFilesByHashes 根据哈希值批量获取文件
func (r *fileRepository) GetFilesByHashes(ctx context.Context, hashes []string, userID uint) ([]*models.File, error) {
	var files []*models.File
	err := r.db.WithContext(ctx).
		Where("hash IN ? AND user_id = ?", hashes, userID).
		Find(&files).Error
	return files, err
}

// GetUserFileCount 获取用户文件数量
func (r *fileRepository) GetUserFileCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.File{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// GetUserStorageUsed 获取用户存储使用量
func (r *fileRepository) GetUserStorageUsed(ctx context.Context, userID uint) (int64, error) {
	var result struct {
		TotalSize int64
	}
	err := r.db.WithContext(ctx).
		Model(&models.File{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(size), 0) as total_size").
		Scan(&result).Error
	return result.TotalSize, err
}

// GetCategoryStats 获取分类统计
func (r *fileRepository) GetCategoryStats(ctx context.Context, userID uint) (map[string]int64, error) {
	var results []struct {
		Category string
		Count    int64
	}
	
	err := r.db.WithContext(ctx).
		Model(&models.File{}).
		Where("user_id = ?", userID).
		Select("category, COUNT(*) as count").
		Group("category").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	
	stats := make(map[string]int64)
	for _, result := range results {
		stats[result.Category] = result.Count
	}
	return stats, nil
}

// GetStorageTierStats 获取存储层级统计
func (r *fileRepository) GetStorageTierStats(ctx context.Context, userID uint) (map[string]int64, error) {
	var results []struct {
		StorageTier string
		TotalSize   int64
	}
	
	err := r.db.WithContext(ctx).
		Model(&models.File{}).
		Where("user_id = ?", userID).
		Select("storage_tier, COALESCE(SUM(size), 0) as total_size").
		Group("storage_tier").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	
	stats := make(map[string]int64)
	for _, result := range results {
		stats[result.StorageTier] = result.TotalSize
	}
	return stats, nil
}

// BatchCreate 批量创建文件
func (r *fileRepository) BatchCreate(ctx context.Context, files []*models.File) error {
	return r.db.WithContext(ctx).CreateInBatches(files, 100).Error
}

// BatchUpdate 批量更新文件
func (r *fileRepository) BatchUpdate(ctx context.Context, files []*models.File) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, file := range files {
			if err := tx.Save(file).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchDelete 批量删除文件
func (r *fileRepository) BatchDelete(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Delete(&models.File{}, ids).Error
}

// BatchUpdateStatus 批量更新状态
func (r *fileRepository) BatchUpdateStatus(ctx context.Context, ids []uint, status int) error {
	return r.db.WithContext(ctx).
		Model(&models.File{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}

// CleanupDeleted 清理已删除的文件
func (r *fileRepository) CleanupDeleted(ctx context.Context, beforeDays int) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("deleted_at IS NOT NULL AND deleted_at < NOW() - INTERVAL ? DAY", beforeDays).
		Unscoped().
		Delete(&models.File{})
	return result.RowsAffected, result.Error
}

// CleanupOrphaned 清理孤立文件记录
func (r *fileRepository) CleanupOrphaned(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("folder_id IS NOT NULL AND folder_id NOT IN (SELECT id FROM file_folders)").
		Update("folder_id", nil)
	return result.RowsAffected, result.Error
}

// ====== 管理员级别统计方法 ======

// GetTotalFileCount 获取全系统文件总数
func (r *fileRepository) GetTotalFileCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.File{}).
		Where("deleted_at IS NULL").
		Count(&count).Error
	return count, err
}

// GetTotalStorageUsed 获取全系统存储使用量
func (r *fileRepository) GetTotalStorageUsed(ctx context.Context) (int64, error) {
	var result struct {
		TotalSize int64
	}
	err := r.db.WithContext(ctx).
		Model(&models.File{}).
		Where("deleted_at IS NULL").
		Select("COALESCE(SUM(size), 0) as total_size").
		Scan(&result).Error
	return result.TotalSize, err
}

// GetGlobalCategoryStats 获取全系统按文件类型的统计
func (r *fileRepository) GetGlobalCategoryStats(ctx context.Context) (map[string]int64, error) {
	var results []struct {
		Category string
		Count    int64
	}
	
	err := r.db.WithContext(ctx).
		Model(&models.File{}).
		Where("deleted_at IS NULL").
		Select("category, COUNT(*) as count").
		Group("category").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	
	stats := make(map[string]int64)
	for _, result := range results {
		category := result.Category
		if category == "" {
			category = "uncategorized"
		}
		stats[category] = result.Count
	}
	
	return stats, nil
}

// GetUserStorageList 获取用户存储使用列表
func (r *fileRepository) GetUserStorageList(ctx context.Context) ([]UserStorageInfo, error) {
	var results []struct {
		UserID      uint
		StorageUsed int64
	}
	
	err := r.db.WithContext(ctx).
		Model(&models.File{}).
		Where("deleted_at IS NULL").
		Select("user_id, COALESCE(SUM(size), 0) as storage_used").
		Group("user_id").
		Order("storage_used DESC").
		Limit(100). // 限制返回前100个用户
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	
	// 转换为UserStorageInfo结构
	storageList := make([]UserStorageInfo, len(results))
	for i, result := range results {
		storageList[i] = UserStorageInfo{
			UserID:      result.UserID,
			Username:    "", // 这里需要从用户服务获取用户名，暂时留空
			StorageUsed: result.StorageUsed,
		}
	}
	
	return storageList, nil
}