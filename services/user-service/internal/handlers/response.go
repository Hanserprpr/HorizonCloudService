package handlers

import (
	"net/http"
	"time"
	
	"github.com/gin-gonic/gin"
)

// Response 通用响应结构
type Response struct {
	Code      int         `json:"code"`      // 业务状态码
	Message   string      `json:"message"`   // 消息
	Data      interface{} `json:"data"`      // 数据
	Success   bool        `json:"success"`   // 是否成功
	Timestamp int64       `json:"timestamp"` // 时间戳
}

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, httpCode int, message string, data interface{}) {
	response := Response{
		Code:      httpCode,
		Message:   message,
		Data:      data,
		Success:   true,
		Timestamp: time.Now().Unix(),
	}
	c.JSON(httpCode, response)
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, httpCode int, message string, detail string) {
	response := Response{
		Code:      httpCode,
		Message:   message,
		Data:      detail,
		Success:   false,
		Timestamp: time.Now().Unix(),
	}
	c.JSON(httpCode, response)
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Items    interface{} `json:"items"`     // 数据项
	Total    int64       `json:"total"`     // 总数
	Page     int         `json:"page"`      // 当前页
	PageSize int         `json:"page_size"` // 页大小
	Pages    int64       `json:"pages"`     // 总页数
}

// ValidationError 验证错误详情
type ValidationError struct {
	Field   string `json:"field"`   // 字段名
	Message string `json:"message"` // 错误消息
}

// ValidationErrorResponse 验证错误响应
func ValidationErrorResponse(c *gin.Context, errors []ValidationError) {
	response := Response{
		Code:      http.StatusBadRequest,
		Message:   "请求参数验证失败",
		Data:      errors,
		Success:   false,
		Timestamp: time.Now().Unix(),
	}
	c.JSON(http.StatusBadRequest, response)
}

// HealthCheckResponse 健康检查响应
type HealthCheckResponse struct {
	Status    string                 `json:"status"`    // 服务状态
	Version   string                 `json:"version"`   // 版本号
	Timestamp int64                  `json:"timestamp"` // 时间戳
	Services  map[string]interface{} `json:"services"`  // 依赖服务状态
}

// UnauthorizedResponse 未授权响应
func UnauthorizedResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, "未授权", message)
}

// ForbiddenResponse 禁止访问响应
func ForbiddenResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, "权限不足", message)
}

// NotFoundResponse 未找到响应
func NotFoundResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, "资源不存在", message)
}

// InternalServerErrorResponse 服务器内部错误响应
func InternalServerErrorResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, "服务器内部错误", message)
}