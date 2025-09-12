package repository

import (
	"time"
	"gorm.io/gorm"
)

// SystemSetting 系统设置模型
type SystemSetting struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Key       string         `gorm:"unique;not null;size:100" json:"key"`
	Value     string         `gorm:"type:text" json:"value"`
	Category  string         `gorm:"size:50" json:"category"`
	IsPublic  bool           `gorm:"default:false" json:"is_public"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (SystemSetting) TableName() string {
	return "system_settings"
}

// SettingsGroup 设置分组结构
type SettingsGroup struct {
	Category string                    `json:"category"`
	Settings map[string]SystemSetting `json:"settings"`
}

// SystemHealthCheck 系统健康检查结果
type SystemHealthCheck struct {
	Service     string            `json:"service"`
	Status      string            `json:"status"`
	Timestamp   int64             `json:"timestamp"`
	Services    map[string]string `json:"services"`
	Database    string            `json:"database"`
	Cache       string            `json:"cache"`
	Storage     string            `json:"storage"`
}

// SystemStats 系统统计信息
type SystemStats struct {
	TotalUsers       int64             `json:"total_users"`
	TotalFiles       int64             `json:"total_files"`
	TotalStorage     int64             `json:"total_storage"`
	UsedStorage      int64             `json:"used_storage"`
	ActiveSessions   int64             `json:"active_sessions"`
	ServicesStatus   map[string]string `json:"services_status"`
	LastUpdateTime   int64             `json:"last_update_time"`
}

// StorageSettings 存储设置结构
type StorageSettings struct {
	StorageType     string `json:"storage_type"`     // local, s3, minio
	LocalPath       string `json:"local_path"`       // 本地存储路径
	S3Bucket        string `json:"s3_bucket"`        // S3存储桶
	S3Region        string `json:"s3_region"`        // S3地区
	S3AccessKey     string `json:"s3_access_key"`    // S3访问密钥
	S3SecretKey     string `json:"s3_secret_key"`    // S3秘密密钥
	MinIOEndpoint   string `json:"minio_endpoint"`   // MinIO端点
	MinIOAccessKey  string `json:"minio_access_key"` // MinIO访问密钥
	MinIOSecretKey  string `json:"minio_secret_key"` // MinIO秘密密钥
	MinIOSecure     bool   `json:"minio_secure"`     // MinIO SSL
	DefaultQuota    int64  `json:"default_quota"`    // 默认用户配额
	MaxFileSize     int64  `json:"max_file_size"`    // 最大文件大小
	AllowedTypes    string `json:"allowed_types"`    // 允许的文件类型
}