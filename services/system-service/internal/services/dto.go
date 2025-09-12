package services

// SystemStatsResponse 系统统计响应
type SystemStatsResponse struct {
	TotalUsers       int64             `json:"total_users"`
	TotalFiles       int64             `json:"total_files"`
	TotalStorage     int64             `json:"total_storage"`
	UsedStorage      int64             `json:"used_storage"`
	AvailableStorage int64             `json:"available_storage"`
	UsagePercent     float64           `json:"usage_percent"`
	ActiveSessions   int64             `json:"active_sessions"`
	ServicesStatus   map[string]string `json:"services_status"`
	LastUpdateTime   int64             `json:"last_update_time"`
	Uptime          int64             `json:"uptime"`
}

// SystemHealthResponse 系统健康响应
type SystemHealthResponse struct {
	Status    string                          `json:"status"`
	Timestamp int64                           `json:"timestamp"`
	Services  map[string]ServiceHealthStatus  `json:"services"`
	Database  string                          `json:"database"`
	Cache     string                          `json:"cache"`
	Storage   string                          `json:"storage"`
	Version   string                          `json:"version"`
}

// ServiceHealthStatus 服务健康状态
type ServiceHealthStatus struct {
	Status       string `json:"status"`
	ResponseTime int64  `json:"response_time_ms"`
	LastChecked  int64  `json:"last_checked"`
	Version      string `json:"version,omitempty"`
	Error        string `json:"error,omitempty"`
}

// CacheOperationResponse 缓存操作响应
type CacheOperationResponse struct {
	Success      bool     `json:"success"`
	Message      string   `json:"message"`
	ClearedItems []string `json:"cleared_items,omitempty"`
	Timestamp    int64    `json:"timestamp"`
}

// SettingsResponse 设置响应
type SettingsResponse struct {
	Settings  map[string]map[string]interface{} `json:"settings"`
	Timestamp int64                             `json:"timestamp"`
}

// UpdateSettingsRequest 更新设置请求
type UpdateSettingsRequest struct {
	Settings map[string]map[string]interface{} `json:"settings" binding:"required"`
}

// StorageSettingsResponse 存储设置响应
type StorageSettingsResponse struct {
	StorageType     string   `json:"storage_type"`
	LocalPath       string   `json:"local_path,omitempty"`
	S3Bucket        string   `json:"s3_bucket,omitempty"`
	S3Region        string   `json:"s3_region,omitempty"`
	S3AccessKey     string   `json:"s3_access_key,omitempty"`
	MinIOEndpoint   string   `json:"minio_endpoint,omitempty"`
	MinIOAccessKey  string   `json:"minio_access_key,omitempty"`
	MinIOSecure     bool     `json:"minio_secure"`
	DefaultQuota    int64    `json:"default_quota"`
	MaxFileSize     int64    `json:"max_file_size"`
	AllowedTypes    []string `json:"allowed_types"`
	Timestamp       int64    `json:"timestamp"`
}

// UpdateStorageSettingsRequest 更新存储设置请求
type UpdateStorageSettingsRequest struct {
	StorageType     string   `json:"storage_type" binding:"required,oneof=local s3 minio"`
	LocalPath       string   `json:"local_path,omitempty"`
	S3Bucket        string   `json:"s3_bucket,omitempty"`
	S3Region        string   `json:"s3_region,omitempty"`
	S3AccessKey     string   `json:"s3_access_key,omitempty"`
	S3SecretKey     string   `json:"s3_secret_key,omitempty"`
	MinIOEndpoint   string   `json:"minio_endpoint,omitempty"`
	MinIOAccessKey  string   `json:"minio_access_key,omitempty"`
	MinIOSecretKey  string   `json:"minio_secret_key,omitempty"`
	MinIOSecure     bool     `json:"minio_secure"`
	DefaultQuota    int64    `json:"default_quota" binding:"required,min=1"`
	MaxFileSize     int64    `json:"max_file_size" binding:"required,min=1"`
	AllowedTypes    []string `json:"allowed_types" binding:"required"`
}

// TestStorageRequest 测试存储请求
type TestStorageRequest struct {
	StorageType     string `json:"storage_type" binding:"required,oneof=local s3 minio"`
	LocalPath       string `json:"local_path,omitempty"`
	S3Bucket        string `json:"s3_bucket,omitempty"`
	S3Region        string `json:"s3_region,omitempty"`
	S3AccessKey     string `json:"s3_access_key,omitempty"`
	S3SecretKey     string `json:"s3_secret_key,omitempty"`
	MinIOEndpoint   string `json:"minio_endpoint,omitempty"`
	MinIOAccessKey  string `json:"minio_access_key,omitempty"`
	MinIOSecretKey  string `json:"minio_secret_key,omitempty"`
	MinIOSecure     bool   `json:"minio_secure"`
}

// TestStorageResponse 测试存储响应
type TestStorageResponse struct {
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

// BasicHealthResponse 基础健康检查响应
type BasicHealthResponse struct {
	Service   string `json:"service"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	Version   string `json:"version"`
	Port      string `json:"port"`
}

// ReadinessResponse 就绪检查响应
type ReadinessResponse struct {
	Service    string            `json:"service"`
	Status     string            `json:"status"`
	Database   string            `json:"database"`
	Components map[string]string `json:"components"`
	Timestamp  int64             `json:"timestamp"`
}