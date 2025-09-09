package repository

import (
	"context"
	"file-service/internal/models"
	"time"

	"gorm.io/gorm"
)

// ShareRepository 分享仓库接口
type ShareRepository interface {
	// 基本CRUD操作
	Create(ctx context.Context, share *models.Share) error
	GetByID(ctx context.Context, id uint) (*models.Share, error)
	GetByToken(ctx context.Context, token string) (*models.Share, error)
	Update(ctx context.Context, share *models.Share) error
	Delete(ctx context.Context, id uint) error
	
	// 查询操作
	GetByUserID(ctx context.Context, userID uint, offset, limit int, filters *ShareFilters) ([]*models.Share, int64, error)
	GetByFileID(ctx context.Context, fileID uint) ([]*models.Share, error)
	GetByFolderID(ctx context.Context, folderID uint) ([]*models.Share, error)
	GetPublicShares(ctx context.Context, offset, limit int) ([]*models.Share, int64, error)
	
	// 状态管理
	UpdateShareCounts(ctx context.Context, shareID uint, action string) error
	UpdateAccessTime(ctx context.Context, shareID uint) error
	ExpireShare(ctx context.Context, shareID uint) error
	
	// 访问日志
	CreateAccessLog(ctx context.Context, log *models.ShareAccessLog) error
	GetAccessLogs(ctx context.Context, shareID uint, offset, limit int) ([]*models.ShareAccessLog, int64, error)
	GetAccessStats(ctx context.Context, shareID uint, days int) (*ShareAccessStats, error)
	
	// 批量操作
	BatchExpireShares(ctx context.Context, shareIDs []uint) error
	BatchDelete(ctx context.Context, shareIDs []uint) error
	BatchUpdateStatus(ctx context.Context, shareIDs []uint, status int) error
	
	// 清理操作
	CleanupExpired(ctx context.Context) (int64, error)
	CleanupAccessLogs(ctx context.Context, beforeDays int) (int64, error)
	
	// 统计查询
	GetUserShareStats(ctx context.Context, userID uint) (*UserShareStats, error)
	GetSystemShareStats(ctx context.Context) (*SystemShareStats, error)
}

// ShareFilters 分享过滤器
type ShareFilters struct {
	Status       *int       `json:"status,omitempty"`
	ResourceType string     `json:"resource_type,omitempty"` // file/folder
	HasPassword  *bool      `json:"has_password,omitempty"`
	IsPublic     *bool      `json:"is_public,omitempty"`
	IsExpired    *bool      `json:"is_expired,omitempty"`
	DateFrom     *time.Time `json:"date_from,omitempty"`
	DateTo       *time.Time `json:"date_to,omitempty"`
	SortBy       string     `json:"sort_by,omitempty"`    // created_at/expires_at/view_count/download_count
	SortOrder    string     `json:"sort_order,omitempty"` // asc/desc
}

// ShareAccessStats 分享访问统计
type ShareAccessStats struct {
	TotalViews    int64            `json:"total_views"`
	TotalDownloads int64           `json:"total_downloads"`
	UniqueIPs     int64            `json:"unique_ips"`
	DailyStats    []DailyAccessStat `json:"daily_stats"`
	TopCountries  []CountryStat    `json:"top_countries"`
}

// DailyAccessStat 每日访问统计
type DailyAccessStat struct {
	Date      string `json:"date"`
	Views     int64  `json:"views"`
	Downloads int64  `json:"downloads"`
}

// CountryStat 国家统计
type CountryStat struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}

// UserShareStats 用户分享统计
type UserShareStats struct {
	TotalShares     int64 `json:"total_shares"`
	ActiveShares    int64 `json:"active_shares"`
	ExpiredShares   int64 `json:"expired_shares"`
	TotalViews      int64 `json:"total_views"`
	TotalDownloads  int64 `json:"total_downloads"`
}

// SystemShareStats 系统分享统计
type SystemShareStats struct {
	TotalShares    int64 `json:"total_shares"`
	ActiveShares   int64 `json:"active_shares"`
	PublicShares   int64 `json:"public_shares"`
	TotalViews     int64 `json:"total_views"`
	TotalDownloads int64 `json:"total_downloads"`
	TotalUsers     int64 `json:"total_users"`
}

// shareRepository 分享仓库实现
type shareRepository struct {
	db *gorm.DB
}

// NewShareRepository 创建分享仓库实例
func NewShareRepository(db *gorm.DB) ShareRepository {
	return &shareRepository{db: db}
}

// Create 创建分享记录
func (r *shareRepository) Create(ctx context.Context, share *models.Share) error {
	return r.db.WithContext(ctx).Create(share).Error
}

// GetByID 根据ID获取分享
func (r *shareRepository) GetByID(ctx context.Context, id uint) (*models.Share, error) {
	var share models.Share
	err := r.db.WithContext(ctx).
		Preload("File").
		Preload("Folder").
		First(&share, id).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// GetByToken 根据令牌获取分享
func (r *shareRepository) GetByToken(ctx context.Context, token string) (*models.Share, error) {
	var share models.Share
	err := r.db.WithContext(ctx).
		Preload("File").
		Preload("Folder").
		Where("token = ?", token).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// Update 更新分享记录
func (r *shareRepository) Update(ctx context.Context, share *models.Share) error {
	return r.db.WithContext(ctx).Save(share).Error
}

// Delete 删除分享记录
func (r *shareRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除访问日志
		if err := tx.Where("share_id = ?", id).Delete(&models.ShareAccessLog{}).Error; err != nil {
			return err
		}
		
		// 删除分享记录
		return tx.Delete(&models.Share{}, id).Error
	})
}

// GetByUserID 获取用户的分享列表
func (r *shareRepository) GetByUserID(ctx context.Context, userID uint, offset, limit int, filters *ShareFilters) ([]*models.Share, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Share{}).Where("user_id = ?", userID)
	
	// 应用过滤器
	query = r.applyShareFilters(query, filters)
	
	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 应用排序
	query = r.applyShareSorting(query, filters)
	
	// 分页查询
	var shares []*models.Share
	err := query.Preload("File").
		Preload("Folder").
		Offset(offset).
		Limit(limit).
		Find(&shares).Error
	
	return shares, total, err
}

// GetByFileID 获取文件的分享列表
func (r *shareRepository) GetByFileID(ctx context.Context, fileID uint) ([]*models.Share, error) {
	var shares []*models.Share
	err := r.db.WithContext(ctx).
		Where("file_id = ?", fileID).
		Order("created_at DESC").
		Find(&shares).Error
	return shares, err
}

// GetByFolderID 获取文件夹的分享列表
func (r *shareRepository) GetByFolderID(ctx context.Context, folderID uint) ([]*models.Share, error) {
	var shares []*models.Share
	err := r.db.WithContext(ctx).
		Where("folder_id = ?", folderID).
		Order("created_at DESC").
		Find(&shares).Error
	return shares, err
}

// GetPublicShares 获取公开分享列表
func (r *shareRepository) GetPublicShares(ctx context.Context, offset, limit int) ([]*models.Share, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Share{}).
		Where("is_public = true AND status = ?", models.ShareStatusActive)
	
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	var shares []*models.Share
	err := query.Preload("File").
		Preload("Folder").
		Order("view_count DESC").
		Offset(offset).
		Limit(limit).
		Find(&shares).Error
	
	return shares, total, err
}

// UpdateShareCounts 更新分享计数
func (r *shareRepository) UpdateShareCounts(ctx context.Context, shareID uint, action string) error {
	updates := make(map[string]interface{})
	now := time.Now()
	updates["last_accessed"] = &now
	
	switch action {
	case models.ShareActionView:
		updates["view_count"] = gorm.Expr("view_count + 1")
	case models.ShareActionDownload:
		updates["download_count"] = gorm.Expr("download_count + 1")
	}
	
	return r.db.WithContext(ctx).
		Model(&models.Share{}).
		Where("id = ?", shareID).
		Updates(updates).Error
}

// UpdateAccessTime 更新最后访问时间
func (r *shareRepository) UpdateAccessTime(ctx context.Context, shareID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Share{}).
		Where("id = ?", shareID).
		Update("last_accessed", &now).Error
}

// ExpireShare 过期分享
func (r *shareRepository) ExpireShare(ctx context.Context, shareID uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Share{}).
		Where("id = ?", shareID).
		Update("status", models.ShareStatusExpired).Error
}

// CreateAccessLog 创建访问日志
func (r *shareRepository) CreateAccessLog(ctx context.Context, log *models.ShareAccessLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetAccessLogs 获取访问日志
func (r *shareRepository) GetAccessLogs(ctx context.Context, shareID uint, offset, limit int) ([]*models.ShareAccessLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.ShareAccessLog{}).Where("share_id = ?", shareID)
	
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	var logs []*models.ShareAccessLog
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error
	
	return logs, total, err
}

// GetAccessStats 获取访问统计
func (r *shareRepository) GetAccessStats(ctx context.Context, shareID uint, days int) (*ShareAccessStats, error) {
	var stats ShareAccessStats
	
	// 总浏览和下载次数
	if err := r.db.WithContext(ctx).
		Model(&models.ShareAccessLog{}).
		Where("share_id = ? AND created_at >= NOW() - INTERVAL ? DAY", shareID, days).
		Select(`
			COUNT(CASE WHEN action = 'view' THEN 1 END) as total_views,
			COUNT(CASE WHEN action = 'download' THEN 1 END) as total_downloads,
			COUNT(DISTINCT ip_address) as unique_ips
		`).
		Scan(&stats).Error; err != nil {
		return nil, err
	}
	
	// 每日统计
	var dailyStats []DailyAccessStat
	if err := r.db.WithContext(ctx).
		Model(&models.ShareAccessLog{}).
		Where("share_id = ? AND created_at >= NOW() - INTERVAL ? DAY", shareID, days).
		Select(`
			DATE(created_at) as date,
			COUNT(CASE WHEN action = 'view' THEN 1 END) as views,
			COUNT(CASE WHEN action = 'download' THEN 1 END) as downloads
		`).
		Group("DATE(created_at)").
		Order("date DESC").
		Find(&dailyStats).Error; err != nil {
		return nil, err
	}
	stats.DailyStats = dailyStats
	
	// 国家统计
	var countryStats []CountryStat
	if err := r.db.WithContext(ctx).
		Model(&models.ShareAccessLog{}).
		Where("share_id = ? AND created_at >= NOW() - INTERVAL ? DAY AND country != ''", shareID, days).
		Select("country, COUNT(*) as count").
		Group("country").
		Order("count DESC").
		Limit(10).
		Find(&countryStats).Error; err != nil {
		return nil, err
	}
	stats.TopCountries = countryStats
	
	return &stats, nil
}

// 应用分享过滤器
func (r *shareRepository) applyShareFilters(query *gorm.DB, filters *ShareFilters) *gorm.DB {
	if filters == nil {
		return query
	}
	
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	
	if filters.ResourceType != "" {
		switch filters.ResourceType {
		case "file":
			query = query.Where("file_id IS NOT NULL")
		case "folder":
			query = query.Where("folder_id IS NOT NULL")
		}
	}
	
	if filters.HasPassword != nil {
		if *filters.HasPassword {
			query = query.Where("password != ''")
		} else {
			query = query.Where("password = ''")
		}
	}
	
	if filters.IsPublic != nil {
		query = query.Where("is_public = ?", *filters.IsPublic)
	}
	
	if filters.IsExpired != nil {
		if *filters.IsExpired {
			query = query.Where("expires_at IS NOT NULL AND expires_at < NOW()")
		} else {
			query = query.Where("expires_at IS NULL OR expires_at >= NOW()")
		}
	}
	
	if filters.DateFrom != nil {
		query = query.Where("created_at >= ?", *filters.DateFrom)
	}
	
	if filters.DateTo != nil {
		query = query.Where("created_at <= ?", *filters.DateTo)
	}
	
	return query
}

// 应用分享排序
func (r *shareRepository) applyShareSorting(query *gorm.DB, filters *ShareFilters) *gorm.DB {
	if filters == nil {
		return query.Order("created_at DESC")
	}
	
	sortBy := "created_at"
	if filters.SortBy != "" {
		switch filters.SortBy {
		case "created_at", "expires_at", "view_count", "download_count":
			sortBy = filters.SortBy
		}
	}
	
	sortOrder := "DESC"
	if filters.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	
	return query.Order(sortBy + " " + sortOrder)
}

// BatchExpireShares 批量过期分享
func (r *shareRepository) BatchExpireShares(ctx context.Context, shareIDs []uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Share{}).
		Where("id IN ?", shareIDs).
		Update("status", models.ShareStatusExpired).Error
}

// BatchDelete 批量删除分享
func (r *shareRepository) BatchDelete(ctx context.Context, shareIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除访问日志
		if err := tx.Where("share_id IN ?", shareIDs).Delete(&models.ShareAccessLog{}).Error; err != nil {
			return err
		}
		
		// 删除分享记录
		return tx.Where("id IN ?", shareIDs).Delete(&models.Share{}).Error
	})
}

// BatchUpdateStatus 批量更新状态
func (r *shareRepository) BatchUpdateStatus(ctx context.Context, shareIDs []uint, status int) error {
	return r.db.WithContext(ctx).
		Model(&models.Share{}).
		Where("id IN ?", shareIDs).
		Update("status", status).Error
}

// CleanupExpired 清理过期分享
func (r *shareRepository) CleanupExpired(ctx context.Context) (int64, error) {
	// 标记过期的分享
	result := r.db.WithContext(ctx).
		Model(&models.Share{}).
		Where("expires_at IS NOT NULL AND expires_at < NOW() AND status != ?", models.ShareStatusExpired).
		Update("status", models.ShareStatusExpired)
	
	return result.RowsAffected, result.Error
}

// CleanupAccessLogs 清理访问日志
func (r *shareRepository) CleanupAccessLogs(ctx context.Context, beforeDays int) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < NOW() - INTERVAL ? DAY", beforeDays).
		Delete(&models.ShareAccessLog{})
	
	return result.RowsAffected, result.Error
}

// GetUserShareStats 获取用户分享统计
func (r *shareRepository) GetUserShareStats(ctx context.Context, userID uint) (*UserShareStats, error) {
	var stats UserShareStats
	
	// 基本统计
	if err := r.db.WithContext(ctx).
		Model(&models.Share{}).
		Where("user_id = ?", userID).
		Select(`
			COUNT(*) as total_shares,
			COUNT(CASE WHEN status = ? THEN 1 END) as active_shares,
			COUNT(CASE WHEN status = ? OR (expires_at IS NOT NULL AND expires_at < NOW()) THEN 1 END) as expired_shares,
			COALESCE(SUM(view_count), 0) as total_views,
			COALESCE(SUM(download_count), 0) as total_downloads
		`, models.ShareStatusActive, models.ShareStatusExpired).
		Scan(&stats).Error; err != nil {
		return nil, err
	}
	
	return &stats, nil
}

// GetSystemShareStats 获取系统分享统计
func (r *shareRepository) GetSystemShareStats(ctx context.Context) (*SystemShareStats, error) {
	var stats SystemShareStats
	
	// 基本统计
	if err := r.db.WithContext(ctx).
		Model(&models.Share{}).
		Select(`
			COUNT(*) as total_shares,
			COUNT(CASE WHEN status = ? THEN 1 END) as active_shares,
			COUNT(CASE WHEN is_public = true THEN 1 END) as public_shares,
			COALESCE(SUM(view_count), 0) as total_views,
			COALESCE(SUM(download_count), 0) as total_downloads
		`, models.ShareStatusActive).
		Scan(&stats).Error; err != nil {
		return nil, err
	}
	
	// 用户数统计
	if err := r.db.WithContext(ctx).
		Model(&models.Share{}).
		Select("COUNT(DISTINCT user_id)").
		Scan(&stats.TotalUsers).Error; err != nil {
		return nil, err
	}
	
	return &stats, nil
}