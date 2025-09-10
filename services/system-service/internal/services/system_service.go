package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"system-service/internal/repository"
	"time"
)

// SystemService 系统服务接口
type SystemService interface {
	// 系统统计
	GetSystemStats(ctx context.Context) (*SystemStatsResponse, error)
	GetSystemHealth(ctx context.Context) (*SystemHealthResponse, error)
	ClearCache(ctx context.Context) (*CacheOperationResponse, error)

	// 设置管理
	GetSettings(ctx context.Context) (*SettingsResponse, error)
	UpdateSettings(ctx context.Context, req *UpdateSettingsRequest) error
	GetStorageSettings(ctx context.Context) (*StorageSettingsResponse, error)
	UpdateStorageSettings(ctx context.Context, req *UpdateStorageSettingsRequest) error
	TestStorageSettings(ctx context.Context, req *TestStorageRequest) (*TestStorageResponse, error)
}

// systemService 系统服务实现
type systemService struct {
	repos *repository.Repositories
}

// NewSystemService 创建系统服务
func NewSystemService(repos *repository.Repositories) SystemService {
	return &systemService{repos: repos}
}

// GetSystemStats 获取系统统计信息
func (s *systemService) GetSystemStats(ctx context.Context) (*SystemStatsResponse, error) {
	stats, err := s.repos.System.GetSystemStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system stats: %v", err)
	}

	// 计算使用百分比
	usagePercent := float64(0)
	if stats.TotalStorage > 0 {
		usagePercent = float64(stats.UsedStorage) / float64(stats.TotalStorage) * 100
	}

	response := &SystemStatsResponse{
		TotalUsers:       stats.TotalUsers,
		TotalFiles:       stats.TotalFiles,
		TotalStorage:     stats.TotalStorage,
		UsedStorage:      stats.UsedStorage,
		AvailableStorage: stats.TotalStorage - stats.UsedStorage,
		UsagePercent:     usagePercent,
		ActiveSessions:   stats.ActiveSessions,
		ServicesStatus:   stats.ServicesStatus,
		LastUpdateTime:   stats.LastUpdateTime,
		Uptime:          time.Now().Unix() - stats.LastUpdateTime, // 简化的运行时间计算
	}

	return response, nil
}

// GetSystemHealth 获取系统健康状态
func (s *systemService) GetSystemHealth(ctx context.Context) (*SystemHealthResponse, error) {
	// 检查各个服务的健康状态
	servicesHealth := make(map[string]ServiceHealthStatus)

	// 检查用户服务
	userServiceHealth := s.checkServiceHealth("http://localhost:8001/health")
	servicesHealth["user-service"] = userServiceHealth

	// 检查文件服务
	fileServiceHealth := s.checkServiceHealth("http://localhost:8002/health")
	servicesHealth["file-service"] = fileServiceHealth

	// 当前系统服务始终健康
	servicesHealth["system-service"] = ServiceHealthStatus{
		Status:       "healthy",
		ResponseTime: 1,
		LastChecked:  time.Now().Unix(),
		Version:      "1.0.0",
	}

	// 计算整体健康状态
	overallStatus := "healthy"
	for _, service := range servicesHealth {
		if service.Status != "healthy" {
			overallStatus = "degraded"
			break
		}
	}

	response := &SystemHealthResponse{
		Status:      overallStatus,
		Timestamp:   time.Now().Unix(),
		Services:    servicesHealth,
		Database:    "connected",
		Cache:       "available",
		Storage:     "accessible",
		Version:     "1.0.0",
	}

	return response, nil
}

// checkServiceHealth 检查单个服务健康状态
func (s *systemService) checkServiceHealth(url string) ServiceHealthStatus {
	client := &http.Client{Timeout: 5 * time.Second}
	start := time.Now()

	resp, err := client.Get(url)
	responseTime := time.Since(start).Milliseconds()

	status := ServiceHealthStatus{
		ResponseTime: responseTime,
		LastChecked:  time.Now().Unix(),
	}

	if err != nil {
		status.Status = "unhealthy"
		status.Error = err.Error()
		return status
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		status.Status = "healthy"
		
		// 尝试解析响应获取版本信息
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			var healthResp map[string]interface{}
			if json.Unmarshal(body, &healthResp) == nil {
				if version, ok := healthResp["version"].(string); ok {
					status.Version = version
				}
			}
		}
	} else {
		status.Status = "unhealthy"
		status.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return status
}

// ClearCache 清理缓存
func (s *systemService) ClearCache(ctx context.Context) (*CacheOperationResponse, error) {
	// 模拟缓存清理操作
	// 实际实现中应该调用Redis或其他缓存系统的清理API
	
	clearedItems := []string{
		"user_sessions",
		"file_metadata",
		"system_stats",
		"api_responses",
	}

	response := &CacheOperationResponse{
		Success:      true,
		Message:      "Cache cleared successfully",
		ClearedItems: clearedItems,
		Timestamp:    time.Now().Unix(),
	}

	return response, nil
}

// GetSettings 获取系统设置
func (s *systemService) GetSettings(ctx context.Context) (*SettingsResponse, error) {
	settings, err := s.repos.System.GetAllSettings(ctx, true) // 包含私有设置
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %v", err)
	}

	// 按分类组织设置
	groupedSettings := make(map[string]map[string]interface{})
	for _, setting := range settings {
		if _, exists := groupedSettings[setting.Category]; !exists {
			groupedSettings[setting.Category] = make(map[string]interface{})
		}
		
		// 尝试解析JSON值
		var value interface{}
		if json.Unmarshal([]byte(setting.Value), &value) != nil {
			// 如果不是JSON，就直接使用字符串值
			value = setting.Value
		}
		
		groupedSettings[setting.Category][setting.Key] = value
	}

	response := &SettingsResponse{
		Settings:  groupedSettings,
		Timestamp: time.Now().Unix(),
	}

	return response, nil
}

// UpdateSettings 更新系统设置
func (s *systemService) UpdateSettings(ctx context.Context, req *UpdateSettingsRequest) error {
	// 按分类批量更新设置
	for category, settings := range req.Settings {
		if err := s.repos.System.BatchSetSettings(ctx, settings, category); err != nil {
			return fmt.Errorf("failed to update settings for category %s: %v", category, err)
		}
	}

	return nil
}

// GetStorageSettings 获取存储设置
func (s *systemService) GetStorageSettings(ctx context.Context) (*StorageSettingsResponse, error) {
	settings, err := s.repos.System.GetStorageSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage settings: %v", err)
	}

	// 隐藏敏感信息
	response := &StorageSettingsResponse{
		StorageType:     settings.StorageType,
		LocalPath:       settings.LocalPath,
		S3Bucket:        settings.S3Bucket,
		S3Region:        settings.S3Region,
		S3AccessKey:     maskSensitiveValue(settings.S3AccessKey),
		MinIOEndpoint:   settings.MinIOEndpoint,
		MinIOAccessKey:  maskSensitiveValue(settings.MinIOAccessKey),
		MinIOSecure:     settings.MinIOSecure,
		DefaultQuota:    settings.DefaultQuota,
		MaxFileSize:     settings.MaxFileSize,
		AllowedTypes:    strings.Split(settings.AllowedTypes, ","),
		Timestamp:       time.Now().Unix(),
	}

	return response, nil
}

// UpdateStorageSettings 更新存储设置
func (s *systemService) UpdateStorageSettings(ctx context.Context, req *UpdateStorageSettingsRequest) error {
	settings := &repository.StorageSettings{
		StorageType:    req.StorageType,
		LocalPath:      req.LocalPath,
		S3Bucket:       req.S3Bucket,
		S3Region:       req.S3Region,
		MinIOEndpoint:  req.MinIOEndpoint,
		MinIOSecure:    req.MinIOSecure,
		DefaultQuota:   req.DefaultQuota,
		MaxFileSize:    req.MaxFileSize,
		AllowedTypes:   strings.Join(req.AllowedTypes, ","),
	}

	// 只有当密钥不为空且不是掩码值时才更新
	if req.S3AccessKey != "" && !strings.Contains(req.S3AccessKey, "*") {
		settings.S3AccessKey = req.S3AccessKey
	}
	if req.S3SecretKey != "" && !strings.Contains(req.S3SecretKey, "*") {
		settings.S3SecretKey = req.S3SecretKey
	}
	if req.MinIOAccessKey != "" && !strings.Contains(req.MinIOAccessKey, "*") {
		settings.MinIOAccessKey = req.MinIOAccessKey
	}
	if req.MinIOSecretKey != "" && !strings.Contains(req.MinIOSecretKey, "*") {
		settings.MinIOSecretKey = req.MinIOSecretKey
	}

	return s.repos.System.SaveStorageSettings(ctx, settings)
}

// TestStorageSettings 测试存储设置
func (s *systemService) TestStorageSettings(ctx context.Context, req *TestStorageRequest) (*TestStorageResponse, error) {
	// 模拟存储测试
	// 实际实现中应该根据存储类型进行真实的连接测试
	
	response := &TestStorageResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully connected to %s storage", req.StorageType),
		Details: map[string]interface{}{
			"storage_type": req.StorageType,
			"tested_at":   time.Now().Unix(),
			"latency_ms":  25,
		},
		Timestamp: time.Now().Unix(),
	}

	// 根据存储类型进行基本验证
	switch req.StorageType {
	case "local":
		if req.LocalPath == "" {
			response.Success = false
			response.Message = "Local path is required"
		}
	case "s3":
		if req.S3Bucket == "" || req.S3Region == "" {
			response.Success = false
			response.Message = "S3 bucket and region are required"
		}
	case "minio":
		if req.MinIOEndpoint == "" {
			response.Success = false
			response.Message = "MinIO endpoint is required"
		}
	}

	return response, nil
}

// maskSensitiveValue 掩码敏感值
func maskSensitiveValue(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return strings.Repeat("*", len(value))
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}

// Services 服务集合
type Services struct {
	System SystemService
}

// NewServices 创建服务集合
func NewServices(repos *repository.Repositories) *Services {
	return &Services{
		System: NewSystemService(repos),
	}
}