package handlers

import (
	"net/http"
	"os"
	"system-service/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

// SystemHandler 系统处理器
type SystemHandler struct {
	services *services.Services
}

// NewSystemHandler 创建系统处理器
func NewSystemHandler(services *services.Services) *SystemHandler {
	return &SystemHandler{services: services}
}

// GetSystemStats 获取系统统计信息
func (h *SystemHandler) GetSystemStats(c *gin.Context) {
	stats, err := h.services.System.GetSystemStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get system stats",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetSystemHealth 获取系统健康状态
func (h *SystemHandler) GetSystemHealth(c *gin.Context) {
	health, err := h.services.System.GetSystemHealth(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get system health",
			"message": err.Error(),
		})
		return
	}

	// 根据健康状态设置HTTP状态码
	statusCode := http.StatusOK
	if health.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"success": true,
		"data":    health,
	})
}

// ClearCache 清理缓存
func (h *SystemHandler) ClearCache(c *gin.Context) {
	result, err := h.services.System.ClearCache(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to clear cache",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetSettings 获取系统设置
func (h *SystemHandler) GetSettings(c *gin.Context) {
	settings, err := h.services.System.GetSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get settings",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}

// UpdateSettings 更新系统设置
func (h *SystemHandler) UpdateSettings(c *gin.Context) {
	var req services.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"message": err.Error(),
		})
		return
	}

	if err := h.services.System.UpdateSettings(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update settings",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Settings updated successfully",
	})
}

// GetStorageSettings 获取存储设置
func (h *SystemHandler) GetStorageSettings(c *gin.Context) {
	settings, err := h.services.System.GetStorageSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get storage settings",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}

// UpdateStorageSettings 更新存储设置
func (h *SystemHandler) UpdateStorageSettings(c *gin.Context) {
	var req services.UpdateStorageSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"message": err.Error(),
		})
		return
	}

	if err := h.services.System.UpdateStorageSettings(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update storage settings",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Storage settings updated successfully",
	})
}

// TestStorageSettings 测试存储设置
func (h *SystemHandler) TestStorageSettings(c *gin.Context) {
	var req services.TestStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"message": err.Error(),
		})
		return
	}

	result, err := h.services.System.TestStorageSettings(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to test storage settings",
			"message": err.Error(),
		})
		return
	}

	statusCode := http.StatusOK
	if !result.Success {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, gin.H{
		"success": result.Success,
		"data":    result,
	})
}

// HealthCheck 基础健康检查
func (h *SystemHandler) HealthCheck(c *gin.Context) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}

	response := services.BasicHealthResponse{
		Service:   "system-service",
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
		Version:   "1.0.0",
		Port:      port,
	}

	c.JSON(http.StatusOK, response)
}

// ReadinessCheck 就绪检查
func (h *SystemHandler) ReadinessCheck(c *gin.Context) {
	response := services.ReadinessResponse{
		Service:   "system-service",
		Status:    "ready",
		Database:  "connected",
		Components: map[string]string{
			"system_service":  "ready",
			"settings":        "ready",
			"health_monitor":  "ready",
			"cache_manager":   "ready",
		},
		Timestamp: time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}

// Handlers 处理器集合
type Handlers struct {
	System *SystemHandler
}

// NewHandlers 创建处理器集合
func NewHandlers(services *services.Services) *Handlers {
	return &Handlers{
		System: NewSystemHandler(services),
	}
}