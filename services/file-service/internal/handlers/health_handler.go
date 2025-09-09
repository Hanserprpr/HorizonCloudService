package handlers

import (
	"file-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	*BaseHandler
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(services *services.Services) *HealthHandler {
	return &HealthHandler{
		BaseHandler: NewBaseHandler(services),
	}
}

// HealthCheckResponse 健康检查响应
type HealthCheckResponse struct {
	Status    string                 `json:"status"`
	Timestamp int64                  `json:"timestamp"`
	Services  map[string]interface{} `json:"services"`
	Version   string                 `json:"version,omitempty"`
	Uptime    string                 `json:"uptime,omitempty"`
}

// Health 健康检查
func (h *HealthHandler) Health(c *gin.Context) {
	health := h.services.HealthCheck()
	
	status := http.StatusOK
	if health["status"] != "ok" {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, Response{
		Code:    200,
		Message: "Health check completed",
		Data:    health,
	})
}

// Ready 就绪检查
func (h *HealthHandler) Ready(c *gin.Context) {
	health := h.services.HealthCheck()
	
	// 检查核心服务是否就绪
	ready := true
	if health["status"] != "ok" {
		ready = false
	}

	status := http.StatusOK
	message := "Service is ready"
	
	if !ready {
		status = http.StatusServiceUnavailable
		message = "Service is not ready"
	}

	c.JSON(status, Response{
		Code:    200,
		Message: message,
		Data: gin.H{
			"ready": ready,
			"health": health,
		},
	})
}

// Live 存活检查
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "Service is alive",
		Data: gin.H{
			"alive": true,
		},
	})
}

// Metrics 服务指标
func (h *HealthHandler) Metrics(c *gin.Context) {
	metrics := h.services.GetServiceMetrics()
	
	h.Success(c, metrics)
}

// Stats 系统统计
func (h *HealthHandler) Stats(c *gin.Context) {
	stats := h.services.GetSystemStats()
	
	h.Success(c, stats)
}

// Version 版本信息
func (h *HealthHandler) Version(c *gin.Context) {
	h.Success(c, gin.H{
		"service": "file-service",
		"version": "1.0.0",
		"build":   "development",
		"go":      "1.23.0",
	})
}