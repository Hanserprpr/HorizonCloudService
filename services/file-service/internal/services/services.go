package services

import (
	"context"
	"file-service/internal/repository"
	"file-service/internal/storage"
	"fmt"
	"time"
)

// Services 服务注册器，管理所有服务实例
type Services struct {
	File      FileService
	Folder    FolderService
	Upload    UploadService
	Share     ShareService
	Thumbnail ThumbnailService
	Version   VersionService
	
	// 依赖
	repo    *repository.Repository
	storage storage.Storage
	
	// 配置
	config *ServicesConfig
	
	// 错误收集器
	errorCollector *ErrorCollector
}

// ServicesConfig 服务配置
type ServicesConfig struct {
	// 上传配置
	DefaultChunkSize    int64         `json:"default_chunk_size"`
	MaxFileSize         int64         `json:"max_file_size"`
	MaxChunks           int           `json:"max_chunks"`
	UploadExpiration    time.Duration `json:"upload_expiration"`
	
	// 分享配置
	DefaultShareExpires time.Duration `json:"default_share_expires"`
	MaxShareExpires     time.Duration `json:"max_share_expires"`
	SharePasswordMinLen int           `json:"share_password_min_len"`
	
	// 缩略图配置
	ThumbnailSizes      []string      `json:"thumbnail_sizes"`
	ThumbnailQuality    int           `json:"thumbnail_quality"`
	ThumbnailTimeout    time.Duration `json:"thumbnail_timeout"`
	
	// 存储配置
	StorageTiers        []string      `json:"storage_tiers"`
	ArchiveDays         int           `json:"archive_days"`
	CleanupInterval     time.Duration `json:"cleanup_interval"`
	
	// 批量操作配置
	MaxBatchSize        int           `json:"max_batch_size"`
	BatchConcurrency    int           `json:"batch_concurrency"`
	BatchTimeout        time.Duration `json:"batch_timeout"`
	
	// 搜索配置
	SearchLimit         int           `json:"search_limit"`
	SearchTimeout       time.Duration `json:"search_timeout"`
	
	// 统计配置
	StatsCache          time.Duration `json:"stats_cache"`
	StatsBatchSize      int           `json:"stats_batch_size"`
}

// DefaultServicesConfig 默认服务配置
func DefaultServicesConfig() *ServicesConfig {
	return &ServicesConfig{
		// 上传配置
		DefaultChunkSize:    5 * 1024 * 1024, // 5MB
		MaxFileSize:         10 * 1024 * 1024 * 1024, // 10GB
		MaxChunks:           10000,
		UploadExpiration:    24 * time.Hour,
		
		// 分享配置
		DefaultShareExpires: 7 * 24 * time.Hour, // 7天
		MaxShareExpires:     30 * 24 * time.Hour, // 30天
		SharePasswordMinLen: 6,
		
		// 缩略图配置
		ThumbnailSizes:   []string{"small", "medium", "large"},
		ThumbnailQuality: 85,
		ThumbnailTimeout: 30 * time.Second,
		
		// 存储配置
		StorageTiers:    []string{"hot", "warm", "cold", "archive"},
		ArchiveDays:     90,
		CleanupInterval: 1 * time.Hour,
		
		// 批量操作配置
		MaxBatchSize:     1000,
		BatchConcurrency: 10,
		BatchTimeout:     5 * time.Minute,
		
		// 搜索配置
		SearchLimit:   1000,
		SearchTimeout: 10 * time.Second,
		
		// 统计配置
		StatsCache:     5 * time.Minute,
		StatsBatchSize: 100,
	}
}

// NewServices 创建服务注册器实例
func NewServices(repo *repository.Repository, storage storage.Storage, config *ServicesConfig) *Services {
	if config == nil {
		config = DefaultServicesConfig()
	}
	
	errorCollector := NewErrorCollector(100) // 保留最近100个错误
	
	services := &Services{
		repo:           repo,
		storage:        storage,
		config:         config,
		errorCollector: errorCollector,
	}
	
	// 初始化各个服务
	services.File = NewFileService(repo, storage)
	services.Folder = NewFolderService(repo)
	services.Upload = NewUploadService(repo, storage)
	services.Version = NewVersionService(repo, storage)
	services.Thumbnail = NewThumbnailService(repo, storage)
	// services.Share = NewShareService(repo, storage) // TODO: 实现分享服务
	
	return services
}

// GetConfig 获取服务配置
func (s *Services) GetConfig() *ServicesConfig {
	return s.config
}

// GetErrorCollector 获取错误收集器
func (s *Services) GetErrorCollector() *ErrorCollector {
	return s.errorCollector
}

// HealthCheck 健康检查
func (s *Services) HealthCheck() map[string]interface{} {
	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	}
	
	// 检查存储健康状态
	if err := s.storage.HealthCheck(context.Background()); err != nil {
		health["storage"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
		health["status"] = "degraded"
	} else {
		health["storage"] = map[string]interface{}{
			"status": "ok",
		}
	}
	
	// 检查数据库连接
	if db := s.repo.GetDB(); db != nil {
		if sqlDB, err := db.DB(); err == nil {
			if err := sqlDB.Ping(); err != nil {
				health["database"] = map[string]interface{}{
					"status": "error",
					"error":  err.Error(),
				}
				health["status"] = "degraded"
			} else {
				health["database"] = map[string]interface{}{
					"status": "ok",
				}
			}
		}
	}
	
	// 添加错误统计
	metrics := s.errorCollector.GetMetrics()
	health["errors"] = map[string]interface{}{
		"total":      metrics.TotalErrors,
		"error_rate": metrics.ErrorRate,
		"by_type":    metrics.ErrorsByType,
	}
	
	return health
}

// GetSystemStats 获取系统统计信息
func (s *Services) GetSystemStats() map[string]interface{} {
	stats := map[string]interface{}{
		"timestamp": time.Now().Unix(),
	}
	
	// TODO: 添加系统统计信息
	// - 总文件数
	// - 总存储使用量
	// - 活跃用户数
	// - 上传统计
	// - 分享统计
	// - 缩略图统计
	
	return stats
}

// Cleanup 清理任务
func (s *Services) Cleanup() error {
	// 清理过期上传会话
	if expired, err := s.Upload.CleanupExpiredSessions(context.Background()); err != nil {
		// 记录错误但继续其他清理任务
		s.errorCollector.RecordError(NewServiceError("CleanupError", "Failed to cleanup expired sessions", err))
	} else if expired > 0 {
		// 记录清理数量
	}
	
	// 清理孤立分片
	if orphaned, err := s.Upload.CleanupOrphanedChunks(context.Background()); err != nil {
		s.errorCollector.RecordError(NewServiceError("CleanupError", "Failed to cleanup orphaned chunks", err))
	} else if orphaned > 0 {
		// 记录清理数量
	}
	
	// TODO: 添加其他清理任务
	// - 清理过期分享
	// - 清理失败的缩略图
	// - 清理临时文件
	// - 清理过期的统计缓存
	
	return nil
}

// StartBackgroundTasks 启动后台任务
func (s *Services) StartBackgroundTasks() {
	// 定期清理任务
	go func() {
		ticker := time.NewTicker(s.config.CleanupInterval)
		defer ticker.Stop()
		
		for range ticker.C {
			if err := s.Cleanup(); err != nil {
				s.errorCollector.RecordError(NewServiceError("BackgroundTaskError", "Cleanup task failed", err))
			}
		}
	}()
	
	// TODO: 添加其他后台任务
	// - 缩略图生成队列处理
	// - 存储层级自动迁移
	// - 统计数据更新
	// - 分享状态检查
}

// 服务工厂方法

// CreateFileService 创建文件服务
func CreateFileService(repo *repository.Repository, storage storage.Storage, config *ServicesConfig) FileService {
	return &fileServiceWithConfig{
		FileService: NewFileService(repo, storage),
		config:      config,
	}
}

// fileServiceWithConfig 带配置的文件服务包装器
type fileServiceWithConfig struct {
	FileService
	config *ServicesConfig
}

// 可以在包装器中添加配置相关的验证和处理
func (s *fileServiceWithConfig) UploadFile(ctx context.Context, req *UploadFileRequest) (*UploadFileResponse, error) {
	// 检查文件大小限制
	if req.Size > s.config.MaxFileSize {
		return nil, NewServiceErrorWithDetails(
			ErrorTypeFileTooLarge,
			"File too large",
			fmt.Sprintf("File size %d exceeds maximum allowed size %d", req.Size, s.config.MaxFileSize),
			ErrFileTooLarge,
		)
	}
	
	// 调用原始方法
	return s.FileService.UploadFile(ctx, req)
}

// ServiceMetrics 服务指标
type ServiceMetrics struct {
	RequestCount    int64             `json:"request_count"`
	RequestRate     float64           `json:"request_rate"`
	AvgResponseTime time.Duration     `json:"avg_response_time"`
	ErrorRate       float64           `json:"error_rate"`
	ActiveSessions  int               `json:"active_sessions"`
	CacheHitRate    float64           `json:"cache_hit_rate"`
	StorageUsage    int64             `json:"storage_usage"`
	Uptime          time.Duration     `json:"uptime"`
	LastError       *ServiceError     `json:"last_error,omitempty"`
	ByService       map[string]int64  `json:"by_service"`
}

// GetServiceMetrics 获取服务指标
func (s *Services) GetServiceMetrics() *ServiceMetrics {
	// TODO: 实现服务指标收集
	return &ServiceMetrics{
		Uptime: time.Since(time.Now()), // 临时实现
	}
}

// 服务中间件

// ServiceMiddleware 服务中间件接口
type ServiceMiddleware interface {
	Before(ctx context.Context, method string, args ...interface{}) error
	After(ctx context.Context, method string, result interface{}, err error) error
}

// LoggingMiddleware 日志中间件
type LoggingMiddleware struct {
	// TODO: 添加日志实现
}

// Before 方法调用前
func (m *LoggingMiddleware) Before(ctx context.Context, method string, args ...interface{}) error {
	// TODO: 记录请求日志
	return nil
}

// After 方法调用后
func (m *LoggingMiddleware) After(ctx context.Context, method string, result interface{}, err error) error {
	// TODO: 记录响应日志
	return nil
}

// MetricsMiddleware 指标中间件
type MetricsMiddleware struct {
	collector *ErrorCollector
}

// NewMetricsMiddleware 创建指标中间件
func NewMetricsMiddleware(collector *ErrorCollector) *MetricsMiddleware {
	return &MetricsMiddleware{collector: collector}
}

// Before 方法调用前
func (m *MetricsMiddleware) Before(ctx context.Context, method string, args ...interface{}) error {
	m.collector.RecordRequest()
	return nil
}

// After 方法调用后
func (m *MetricsMiddleware) After(ctx context.Context, method string, result interface{}, err error) error {
	if serviceErr := GetServiceError(err); serviceErr != nil {
		m.collector.RecordError(serviceErr)
	}
	return nil
}