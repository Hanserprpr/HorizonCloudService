package repository

import (
	"context"
	"file-service/internal/models"

	"gorm.io/gorm"
)

// ThumbnailRepository 缩略图仓库接口
type ThumbnailRepository interface {
	// 基本CRUD操作
	Create(ctx context.Context, thumbnail *models.Thumbnail) error
	GetByID(ctx context.Context, id uint) (*models.Thumbnail, error)
	GetByFileID(ctx context.Context, fileID uint) ([]*models.Thumbnail, error)
	GetByFileAndSize(ctx context.Context, fileID uint, size string) (*models.Thumbnail, error)
	Update(ctx context.Context, thumbnail *models.Thumbnail) error
	Delete(ctx context.Context, id uint) error
	
	// 批量操作
	BatchCreate(ctx context.Context, thumbnails []*models.Thumbnail) error
	BatchDelete(ctx context.Context, fileIDs []uint) error
	GetPendingGeneration(ctx context.Context, limit int) ([]*models.Thumbnail, error)
	
	// 状态管理
	UpdateGenerationStatus(ctx context.Context, id uint, status int, path string) error
	MarkGenerationFailed(ctx context.Context, id uint, errorMsg string) error
	
	// 清理操作
	CleanupFailed(ctx context.Context, beforeDays int) (int64, error)
	CleanupOrphaned(ctx context.Context) (int64, error)
	
	// 统计查询
	GetUserStats(ctx context.Context, userID uint) (*UserThumbnailStats, error)
	GetGenerationStats(ctx context.Context) (*ThumbnailStats, error)
	GetFileGenerationProgress(ctx context.Context, fileID uint) (*ThumbnailProgress, error)
}

// ThumbnailStats 缩略图统计信息
type ThumbnailStats struct {
	Total     int64 `json:"total"`
	Pending   int64 `json:"pending"`
	Generated int64 `json:"generated"`
	Failed    int64 `json:"failed"`
}

// ThumbnailProgress 文件缩略图生成进度
type ThumbnailProgress struct {
	FileID    uint  `json:"file_id"`
	Total     int   `json:"total"`
	Generated int   `json:"generated"`
	Failed    int   `json:"failed"`
	Progress  float64 `json:"progress"`
}

// UserThumbnailStats 用户缩略图统计
type UserThumbnailStats struct {
	TotalThumbnails     int64            `json:"total_thumbnails"`
	FilesWithThumbnails int64            `json:"files_with_thumbnails"`
	TotalSize          int64            `json:"total_size"`
	BySize             map[string]int64 `json:"by_size"`
}

// thumbnailRepository 缩略图仓库实现
type thumbnailRepository struct {
	db *gorm.DB
}

// NewThumbnailRepository 创建缩略图仓库实例
func NewThumbnailRepository(db *gorm.DB) ThumbnailRepository {
	return &thumbnailRepository{db: db}
}

// Create 创建缩略图记录
func (r *thumbnailRepository) Create(ctx context.Context, thumbnail *models.Thumbnail) error {
	return r.db.WithContext(ctx).Create(thumbnail).Error
}

// GetByID 根据ID获取缩略图
func (r *thumbnailRepository) GetByID(ctx context.Context, id uint) (*models.Thumbnail, error) {
	var thumbnail models.Thumbnail
	err := r.db.WithContext(ctx).
		Preload("File").
		First(&thumbnail, id).Error
	if err != nil {
		return nil, err
	}
	return &thumbnail, nil
}

// GetByFileID 获取文件的所有缩略图
func (r *thumbnailRepository) GetByFileID(ctx context.Context, fileID uint) ([]*models.Thumbnail, error) {
	var thumbnails []*models.Thumbnail
	err := r.db.WithContext(ctx).
		Where("file_id = ?", fileID).
		Order("size ASC").
		Find(&thumbnails).Error
	return thumbnails, err
}

// GetByFileAndSize 根据文件ID和尺寸获取缩略图
func (r *thumbnailRepository) GetByFileAndSize(ctx context.Context, fileID uint, size string) (*models.Thumbnail, error) {
	var thumbnail models.Thumbnail
	err := r.db.WithContext(ctx).
		Where("file_id = ? AND size = ?", fileID, size).
		First(&thumbnail).Error
	if err != nil {
		return nil, err
	}
	return &thumbnail, nil
}

// Update 更新缩略图记录
func (r *thumbnailRepository) Update(ctx context.Context, thumbnail *models.Thumbnail) error {
	return r.db.WithContext(ctx).Save(thumbnail).Error
}

// Delete 删除缩略图记录
func (r *thumbnailRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Thumbnail{}, id).Error
}

// BatchCreate 批量创建缩略图记录
func (r *thumbnailRepository) BatchCreate(ctx context.Context, thumbnails []*models.Thumbnail) error {
	return r.db.WithContext(ctx).CreateInBatches(thumbnails, 100).Error
}

// BatchDelete 批量删除文件的缩略图
func (r *thumbnailRepository) BatchDelete(ctx context.Context, fileIDs []uint) error {
	return r.db.WithContext(ctx).
		Where("file_id IN ?", fileIDs).
		Delete(&models.Thumbnail{}).Error
}

// GetPendingGeneration 获取待生成的缩略图
func (r *thumbnailRepository) GetPendingGeneration(ctx context.Context, limit int) ([]*models.Thumbnail, error) {
	var thumbnails []*models.Thumbnail
	err := r.db.WithContext(ctx).
		Where("status = ?", models.ThumbnailStatusGenerating).
		Preload("File").
		Order("created_at ASC").
		Limit(limit).
		Find(&thumbnails).Error
	return thumbnails, err
}

// UpdateGenerationStatus 更新生成状态
func (r *thumbnailRepository) UpdateGenerationStatus(ctx context.Context, id uint, status int, path string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if path != "" {
		updates["path"] = path
	}
	if status == models.ThumbnailStatusReady {
		updates["updated_at"] = gorm.Expr("NOW()")
	}
	
	return r.db.WithContext(ctx).
		Model(&models.Thumbnail{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// MarkGenerationFailed 标记生成失败
func (r *thumbnailRepository) MarkGenerationFailed(ctx context.Context, id uint, errorMsg string) error {
	return r.db.WithContext(ctx).
		Model(&models.Thumbnail{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        models.ThumbnailStatusFailed,
			"error_message": errorMsg,
		}).Error
}

// CleanupFailed 清理失败的缩略图记录
func (r *thumbnailRepository) CleanupFailed(ctx context.Context, beforeDays int) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("status = ? AND created_at < NOW() - INTERVAL ? DAY", 
			models.ThumbnailStatusFailed, beforeDays).
		Delete(&models.Thumbnail{})
	
	return result.RowsAffected, result.Error
}

// CleanupOrphaned 清理孤立的缩略图记录
func (r *thumbnailRepository) CleanupOrphaned(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("file_id NOT IN (SELECT id FROM file_files)").
		Delete(&models.Thumbnail{})
	
	return result.RowsAffected, result.Error
}

// GetGenerationStats 获取生成统计信息
func (r *thumbnailRepository) GetGenerationStats(ctx context.Context) (*ThumbnailStats, error) {
	var stats ThumbnailStats
	
	// 总数
	if err := r.db.WithContext(ctx).Model(&models.Thumbnail{}).Count(&stats.Total).Error; err != nil {
		return nil, err
	}
	
	// 各状态统计
	type statusCount struct {
		Status int   `json:"status"`
		Count  int64 `json:"count"`
	}
	
	var statusCounts []statusCount
	err := r.db.WithContext(ctx).
		Model(&models.Thumbnail{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts).Error
	if err != nil {
		return nil, err
	}
	
	for _, sc := range statusCounts {
		switch sc.Status {
		case models.ThumbnailStatusGenerating:
			stats.Pending = sc.Count
		case models.ThumbnailStatusReady:
			stats.Generated = sc.Count
		case models.ThumbnailStatusFailed:
			stats.Failed = sc.Count
		}
	}
	
	return &stats, nil
}

// GetFileGenerationProgress 获取文件缩略图生成进度
func (r *thumbnailRepository) GetFileGenerationProgress(ctx context.Context, fileID uint) (*ThumbnailProgress, error) {
	var progress ThumbnailProgress
	progress.FileID = fileID
	
	type statusCount struct {
		Status int   `json:"status"`
		Count  int   `json:"count"`
	}
	
	var statusCounts []statusCount
	err := r.db.WithContext(ctx).
		Model(&models.Thumbnail{}).
		Where("file_id = ?", fileID).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts).Error
	if err != nil {
		return nil, err
	}
	
	for _, sc := range statusCounts {
		progress.Total += sc.Count
		switch sc.Status {
		case models.ThumbnailStatusReady:
			progress.Generated = sc.Count
		case models.ThumbnailStatusFailed:
			progress.Failed = sc.Count
		}
	}
	
	if progress.Total > 0 {
		progress.Progress = float64(progress.Generated) / float64(progress.Total) * 100
	}
	
	return &progress, nil
}

// GetUserStats 获取用户缩略图统计
func (r *thumbnailRepository) GetUserStats(ctx context.Context, userID uint) (*UserThumbnailStats, error) {
	stats := &UserThumbnailStats{
		BySize: make(map[string]int64),
	}
	
	// 通过JOIN获取用户的缩略图统计
	type thumbnailCount struct {
		Size  string `json:"size"`
		Count int64  `json:"count"`
		TotalSize int64 `json:"total_size"`
	}
	
	var counts []thumbnailCount
	err := r.db.WithContext(ctx).
		Table("file_thumbnails t").
		Select("t.size, COUNT(*) as count, SUM(t.file_size) as total_size").
		Joins("JOIN file_files f ON t.file_id = f.id").
		Where("f.user_id = ? AND t.status = ?", userID, models.ThumbnailStatusReady).
		Group("t.size").
		Find(&counts).Error
	if err != nil {
		return nil, err
	}
	
	for _, count := range counts {
		stats.TotalThumbnails += count.Count
		stats.TotalSize += count.TotalSize
		stats.BySize[count.Size] = count.Count
	}
	
	// 获取有缩略图的文件数量
	err = r.db.WithContext(ctx).
		Table("file_files f").
		Joins("JOIN file_thumbnails t ON f.id = t.file_id").
		Where("f.user_id = ? AND t.status = ?", userID, models.ThumbnailStatusReady).
		Group("f.id").
		Count(&stats.FilesWithThumbnails).Error
	if err != nil {
		return nil, err
	}
	
	return stats, nil
}