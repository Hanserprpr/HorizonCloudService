package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

// SystemRepository 系统设置仓库接口
type SystemRepository interface {
	// 设置管理
	GetSetting(ctx context.Context, key string) (*SystemSetting, error)
	GetSettingsByCategory(ctx context.Context, category string) ([]*SystemSetting, error)
	GetAllSettings(ctx context.Context, includePrivate bool) ([]*SystemSetting, error)
	SetSetting(ctx context.Context, key, value, category string, isPublic bool) error
	DeleteSetting(ctx context.Context, key string) error
	BatchSetSettings(ctx context.Context, settings map[string]interface{}, category string) error

	// 存储设置
	GetStorageSettings(ctx context.Context) (*StorageSettings, error)
	SaveStorageSettings(ctx context.Context, settings *StorageSettings) error

	// 系统统计
	GetSystemStats(ctx context.Context) (*SystemStats, error)
}

// systemRepository 系统仓库实现
type systemRepository struct {
	db *gorm.DB
}

// NewSystemRepository 创建系统仓库
func NewSystemRepository(db *gorm.DB) SystemRepository {
	return &systemRepository{db: db}
}

// GetSetting 获取单个设置
func (r *systemRepository) GetSetting(ctx context.Context, key string) (*SystemSetting, error) {
	var setting SystemSetting
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

// GetSettingsByCategory 获取分类设置
func (r *systemRepository) GetSettingsByCategory(ctx context.Context, category string) ([]*SystemSetting, error) {
	var settings []*SystemSetting
	err := r.db.WithContext(ctx).Where("category = ?", category).Find(&settings).Error
	return settings, err
}

// GetAllSettings 获取所有设置
func (r *systemRepository) GetAllSettings(ctx context.Context, includePrivate bool) ([]*SystemSetting, error) {
	var settings []*SystemSetting
	query := r.db.WithContext(ctx)
	
	if !includePrivate {
		query = query.Where("is_public = ?", true)
	}
	
	err := query.Find(&settings).Error
	return settings, err
}

// SetSetting 设置单个配置
func (r *systemRepository) SetSetting(ctx context.Context, key, value, category string, isPublic bool) error {
	setting := SystemSetting{
		Key:      key,
		Value:    value,
		Category: category,
		IsPublic: isPublic,
	}

	// 使用Upsert操作
	return r.db.WithContext(ctx).Save(&setting).Error
}

// DeleteSetting 删除设置
func (r *systemRepository) DeleteSetting(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).Where("key = ?", key).Delete(&SystemSetting{}).Error
}

// BatchSetSettings 批量设置
func (r *systemRepository) BatchSetSettings(ctx context.Context, settings map[string]interface{}, category string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, value := range settings {
			// 将值转换为JSON字符串存储
			valueJSON, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal value for key %s: %v", key, err)
			}

			setting := SystemSetting{
				Key:      key,
				Value:    string(valueJSON),
				Category: category,
				IsPublic: true, // 默认公开
			}

			if err := tx.Save(&setting).Error; err != nil {
				return fmt.Errorf("failed to save setting %s: %v", key, err)
			}
		}
		return nil
	})
}

// GetStorageSettings 获取存储设置
func (r *systemRepository) GetStorageSettings(ctx context.Context) (*StorageSettings, error) {
	settings, err := r.GetSettingsByCategory(ctx, "storage")
	if err != nil {
		return nil, err
	}

	storageSettings := &StorageSettings{
		StorageType:  "local", // 默认值
		LocalPath:    "./uploads",
		DefaultQuota: 10 * 1024 * 1024 * 1024, // 10GB
		MaxFileSize:  100 * 1024 * 1024,       // 100MB
		AllowedTypes: "image/*,text/*,application/pdf",
	}

	// 将设置映射到结构体
	for _, setting := range settings {
		switch setting.Key {
		case "storage_type":
			storageSettings.StorageType = setting.Value
		case "local_path":
			storageSettings.LocalPath = setting.Value
		case "s3_bucket":
			storageSettings.S3Bucket = setting.Value
		case "s3_region":
			storageSettings.S3Region = setting.Value
		case "minio_endpoint":
			storageSettings.MinIOEndpoint = setting.Value
		case "default_quota":
			json.Unmarshal([]byte(setting.Value), &storageSettings.DefaultQuota)
		case "max_file_size":
			json.Unmarshal([]byte(setting.Value), &storageSettings.MaxFileSize)
		case "allowed_types":
			storageSettings.AllowedTypes = setting.Value
		}
	}

	return storageSettings, nil
}

// SaveStorageSettings 保存存储设置
func (r *systemRepository) SaveStorageSettings(ctx context.Context, settings *StorageSettings) error {
	settingsMap := map[string]interface{}{
		"storage_type":     settings.StorageType,
		"local_path":       settings.LocalPath,
		"s3_bucket":        settings.S3Bucket,
		"s3_region":        settings.S3Region,
		"minio_endpoint":   settings.MinIOEndpoint,
		"default_quota":    settings.DefaultQuota,
		"max_file_size":    settings.MaxFileSize,
		"allowed_types":    settings.AllowedTypes,
		"minio_secure":     settings.MinIOSecure,
	}

	// 敏感信息单独处理（标记为私有）
	if settings.S3AccessKey != "" {
		if err := r.SetSetting(ctx, "s3_access_key", settings.S3AccessKey, "storage", false); err != nil {
			return err
		}
	}
	if settings.S3SecretKey != "" {
		if err := r.SetSetting(ctx, "s3_secret_key", settings.S3SecretKey, "storage", false); err != nil {
			return err
		}
	}
	if settings.MinIOAccessKey != "" {
		if err := r.SetSetting(ctx, "minio_access_key", settings.MinIOAccessKey, "storage", false); err != nil {
			return err
		}
	}
	if settings.MinIOSecretKey != "" {
		if err := r.SetSetting(ctx, "minio_secret_key", settings.MinIOSecretKey, "storage", false); err != nil {
			return err
		}
	}

	return r.BatchSetSettings(ctx, settingsMap, "storage")
}

// GetSystemStats 获取系统统计信息
func (r *systemRepository) GetSystemStats(ctx context.Context) (*SystemStats, error) {
	// 模拟系统统计数据（实际实现中应该从各个服务获取真实数据）
	stats := &SystemStats{
		TotalUsers:     150,
		TotalFiles:     1250,
		TotalStorage:   500 * 1024 * 1024 * 1024, // 500GB
		UsedStorage:    120 * 1024 * 1024 * 1024, // 120GB
		ActiveSessions: 25,
		ServicesStatus: map[string]string{
			"user-service":   "healthy",
			"file-service":   "healthy", 
			"system-service": "healthy",
		},
		LastUpdateTime: r.db.NowFunc().Unix(),
	}

	return stats, nil
}

// Repositories 仓库集合
type Repositories struct {
	System SystemRepository
}

// NewRepositories 创建仓库集合
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		System: NewSystemRepository(db),
	}
}