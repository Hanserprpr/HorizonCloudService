package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"file-service/internal/repository"
	"file-service/internal/services"

	"github.com/gin-gonic/gin"
)

// QuotaConfig 配额配置
type QuotaConfig struct {
	DefaultStorageQuota int64         // 默认存储配额（字节）
	DefaultFileCount    int64         // 默认文件数量限制
	CheckInterval       time.Duration // 配额检查间隔
	GraceBuffer         float64       // 宽限缓冲区（0.1表示10%）
	EnableWarnings      bool          // 是否启用警告
	WarningThreshold    float64       // 警告阈值（0.8表示80%）
}

// QuotaMiddleware 配额中间件
type QuotaMiddleware struct {
	config         *QuotaConfig
	userService    services.UserServiceClient
	fileRepository repository.FileRepository
}

// UserQuota 用户配额信息
type UserQuota struct {
	UserID           uint  `json:"user_id"`
	StorageQuota     int64 `json:"storage_quota"`     // 存储配额（字节）
	StorageUsed      int64 `json:"storage_used"`      // 已使用存储（字节）
	FileCountQuota   int64 `json:"file_count_quota"`  // 文件数量配额
	FileCount        int64 `json:"file_count"`        // 当前文件数量
	StorageUsageRate float64 `json:"storage_usage_rate"` // 存储使用率
	FileCountUsageRate float64 `json:"file_count_usage_rate"` // 文件数量使用率
	IsStorageExceeded bool `json:"is_storage_exceeded"` // 存储是否超额
	IsFileCountExceeded bool `json:"is_file_count_exceeded"` // 文件数量是否超额
	LastChecked      time.Time `json:"last_checked"`
}

// NewQuotaMiddleware 创建配额中间件
func NewQuotaMiddleware(config *QuotaConfig, userService services.UserServiceClient, fileRepo repository.FileRepository) *QuotaMiddleware {
	if config.DefaultStorageQuota == 0 {
		config.DefaultStorageQuota = 5 * 1024 * 1024 * 1024 // 5GB默认配额
	}
	if config.DefaultFileCount == 0 {
		config.DefaultFileCount = 10000 // 默认10000个文件
	}
	if config.CheckInterval == 0 {
		config.CheckInterval = 5 * time.Minute // 5分钟检查间隔
	}
	if config.GraceBuffer == 0 {
		config.GraceBuffer = 0.1 // 10%宽限缓冲
	}
	if config.WarningThreshold == 0 {
		config.WarningThreshold = 0.8 // 80%警告阈值
	}

	return &QuotaMiddleware{
		config:         config,
		userService:    userService,
		fileRepository: fileRepo,
	}
}

// CheckQuota 检查配额中间件
func (m *QuotaMiddleware) CheckQuota() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只有认证用户才检查配额
		if !IsAuthenticated(c) {
			c.Next()
			return
		}

		userID, err := GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "User not authenticated for quota check",
			})
			c.Abort()
			return
		}

		// 获取用户配额信息
		quota, err := m.GetUserQuota(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": fmt.Sprintf("Failed to check quota: %v", err),
			})
			c.Abort()
			return
		}

		// 将配额信息设置到上下文
		c.Set("user_quota", quota)

		// 检查存储配额
		if quota.IsStorageExceeded {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "Storage quota exceeded",
				"error": gin.H{
					"type":         "QUOTA_EXCEEDED",
					"quota_type":   "storage",
					"used":         quota.StorageUsed,
					"limit":        quota.StorageQuota,
					"usage_rate":   quota.StorageUsageRate,
					"description": "Your storage quota has been exceeded. Please delete some files or upgrade your plan.",
				},
			})
			c.Abort()
			return
		}

		// 检查文件数量配额
		if quota.IsFileCountExceeded {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "File count quota exceeded",
				"error": gin.H{
					"type":         "QUOTA_EXCEEDED",
					"quota_type":   "file_count",
					"used":         quota.FileCount,
					"limit":        quota.FileCountQuota,
					"usage_rate":   quota.FileCountUsageRate,
					"description": "Your file count quota has been exceeded. Please delete some files or upgrade your plan.",
				},
			})
			c.Abort()
			return
		}

		// 发送警告（如果启用）
		if m.config.EnableWarnings {
			m.checkAndSendWarnings(c, quota)
		}

		c.Next()
	}
}

// CheckUploadQuota 检查上传配额
func (m *QuotaMiddleware) CheckUploadQuota(fileSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAuthenticated(c) {
			c.Next()
			return
		}

		userID, err := GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "User not authenticated",
			})
			c.Abort()
			return
		}

		// 获取用户配额
		quota, err := m.GetUserQuota(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": fmt.Sprintf("Failed to check quota: %v", err),
			})
			c.Abort()
			return
		}

		// 检查上传后是否会超过存储配额
		projectedUsage := quota.StorageUsed + fileSize
		if projectedUsage > quota.StorageQuota {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "Upload would exceed storage quota",
				"error": gin.H{
					"type":            "QUOTA_EXCEEDED",
					"quota_type":      "storage",
					"current_used":    quota.StorageUsed,
					"upload_size":     fileSize,
					"projected_usage": projectedUsage,
					"limit":           quota.StorageQuota,
					"available":       quota.StorageQuota - quota.StorageUsed,
					"description":     "This upload would exceed your storage quota. Please delete some files or upgrade your plan.",
				},
			})
			c.Abort()
			return
		}

		// 检查文件数量配额
		if quota.FileCount >= quota.FileCountQuota {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "File count quota exceeded",
				"error": gin.H{
					"type":         "QUOTA_EXCEEDED",
					"quota_type":   "file_count",
					"current_count": quota.FileCount,
					"limit":        quota.FileCountQuota,
					"description":  "You have reached the maximum number of files allowed. Please delete some files or upgrade your plan.",
				},
			})
			c.Abort()
			return
		}

		c.Set("user_quota", quota)
		c.Next()
	}
}

// CheckBatchUploadQuota 检查批量上传配额
func (m *QuotaMiddleware) CheckBatchUploadQuota() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAuthenticated(c) {
			c.Next()
			return
		}

		userID, err := GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "User not authenticated",
			})
			c.Abort()
			return
		}

		// 从请求中获取文件信息
		var batchRequest struct {
			Files []struct {
				Size int64 `json:"size"`
			} `json:"files"`
		}

		if err := c.ShouldBindJSON(&batchRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "Invalid request format",
			})
			c.Abort()
			return
		}

		// 计算总文件大小和数量
		var totalSize int64
		fileCount := len(batchRequest.Files)

		for _, file := range batchRequest.Files {
			totalSize += file.Size
		}

		// 获取用户配额
		quota, err := m.GetUserQuota(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": fmt.Sprintf("Failed to check quota: %v", err),
			})
			c.Abort()
			return
		}

		// 检查存储配额
		projectedStorageUsage := quota.StorageUsed + totalSize
		if projectedStorageUsage > quota.StorageQuota {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "Batch upload would exceed storage quota",
				"error": gin.H{
					"type":            "QUOTA_EXCEEDED",
					"quota_type":      "storage",
					"current_used":    quota.StorageUsed,
					"upload_size":     totalSize,
					"projected_usage": projectedStorageUsage,
					"limit":           quota.StorageQuota,
					"file_count":      fileCount,
				},
			})
			c.Abort()
			return
		}

		// 检查文件数量配额
		projectedFileCount := quota.FileCount + int64(fileCount)
		if projectedFileCount > quota.FileCountQuota {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "Batch upload would exceed file count quota",
				"error": gin.H{
					"type":                "QUOTA_EXCEEDED",
					"quota_type":          "file_count",
					"current_count":       quota.FileCount,
					"upload_count":        fileCount,
					"projected_count":     projectedFileCount,
					"limit":               quota.FileCountQuota,
				},
			})
			c.Abort()
			return
		}

		c.Set("user_quota", quota)
		c.Set("batch_upload_size", totalSize)
		c.Set("batch_upload_count", fileCount)
		c.Next()
	}
}

// GetUserQuota 获取用户配额信息
func (m *QuotaMiddleware) GetUserQuota(ctx context.Context, userID uint) (*UserQuota, error) {
	// 获取用户信息（从用户服务）
	userInfo, err := m.userService.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// 获取用户文件统计
	storageUsed, err := m.fileRepository.GetUserStorageUsed(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user storage used: %w", err)
	}

	fileCount, err := m.fileRepository.GetUserFileCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user file count: %w", err)
	}

	// 构建配额信息
	quota := &UserQuota{
		UserID:         userID,
		StorageQuota:   userInfo.StorageQuota,
		FileCountQuota: userInfo.FileCountQuota,
		StorageUsed:    storageUsed,
		FileCount:      fileCount,
		LastChecked:    time.Now(),
	}

	// 如果用户没有设置配额，使用默认值
	if quota.StorageQuota == 0 {
		quota.StorageQuota = m.config.DefaultStorageQuota
	}
	if quota.FileCountQuota == 0 {
		quota.FileCountQuota = m.config.DefaultFileCount
	}

	// 计算使用率
	quota.StorageUsageRate = float64(quota.StorageUsed) / float64(quota.StorageQuota)
	quota.FileCountUsageRate = float64(quota.FileCount) / float64(quota.FileCountQuota)

	// 检查是否超额（考虑宽限缓冲）
	storageThreshold := float64(quota.StorageQuota) * (1.0 + m.config.GraceBuffer)
	fileCountThreshold := float64(quota.FileCountQuota) * (1.0 + m.config.GraceBuffer)

	quota.IsStorageExceeded = float64(quota.StorageUsed) > storageThreshold
	quota.IsFileCountExceeded = float64(quota.FileCount) > fileCountThreshold

	return quota, nil
}

// UpdateUserQuota 更新用户配额（管理员操作）
func (m *QuotaMiddleware) UpdateUserQuota(ctx context.Context, userID uint, storageQuota, fileCountQuota int64) error {
	return m.userService.UpdateUserQuota(ctx, userID, storageQuota, fileCountQuota)
}

// GetQuotaFromContext 从上下文获取配额信息
func GetQuotaFromContext(c *gin.Context) (*UserQuota, bool) {
	quota, exists := c.Get("user_quota")
	if !exists {
		return nil, false
	}

	userQuota, ok := quota.(*UserQuota)
	return userQuota, ok
}

// checkAndSendWarnings 检查并发送警告
func (m *QuotaMiddleware) checkAndSendWarnings(c *gin.Context, quota *UserQuota) {
	warnings := make([]string, 0)

	// 检查存储警告
	if quota.StorageUsageRate >= m.config.WarningThreshold && !quota.IsStorageExceeded {
		warnings = append(warnings, fmt.Sprintf("Storage usage is %.1f%% of quota", quota.StorageUsageRate*100))
	}

	// 检查文件数量警告
	if quota.FileCountUsageRate >= m.config.WarningThreshold && !quota.IsFileCountExceeded {
		warnings = append(warnings, fmt.Sprintf("File count is %.1f%% of quota", quota.FileCountUsageRate*100))
	}

	// 设置警告到响应头
	if len(warnings) > 0 {
		c.Header("X-Quota-Warnings", fmt.Sprintf("%v", warnings))
	}
}

// FormatFileSize 格式化文件大小显示
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetQuotaStatus 获取配额状态
func (m *QuotaMiddleware) GetQuotaStatus(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "User not authenticated",
		})
		return
	}

	quota, err := m.GetUserQuota(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": fmt.Sprintf("Failed to get quota status: %v", err),
		})
		return
	}

	// 添加格式化的显示信息
	response := gin.H{
		"user_id":     quota.UserID,
		"storage": gin.H{
			"used":        quota.StorageUsed,
			"quota":       quota.StorageQuota,
			"available":   quota.StorageQuota - quota.StorageUsed,
			"usage_rate":  quota.StorageUsageRate,
			"used_formatted": FormatFileSize(quota.StorageUsed),
			"quota_formatted": FormatFileSize(quota.StorageQuota),
			"available_formatted": FormatFileSize(quota.StorageQuota - quota.StorageUsed),
			"is_exceeded": quota.IsStorageExceeded,
		},
		"file_count": gin.H{
			"used":        quota.FileCount,
			"quota":       quota.FileCountQuota,
			"available":   quota.FileCountQuota - quota.FileCount,
			"usage_rate":  quota.FileCountUsageRate,
			"is_exceeded": quota.IsFileCountExceeded,
		},
		"last_checked": quota.LastChecked,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Quota status retrieved successfully",
		"data":    response,
	})
}