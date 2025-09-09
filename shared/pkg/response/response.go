package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 响应码定义
const (
	CodeSuccess = 200
	CodeError   = 500
	CodeInvalidRequest = 400
	CodeUnauthorized   = 401
	CodeForbidden      = 403
	CodeNotFound       = 404
	CodeConflict       = 409
	CodeTooManyRequests = 429
)

// Success 成功响应
func Success(c *gin.Context, data ...interface{}) {
	response := Response{
		Code:    CodeSuccess,
		Message: "success",
	}
	
	if len(data) > 0 && data[0] != nil {
		response.Data = data[0]
	}
	
	c.JSON(http.StatusOK, response)
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	response := Response{
		Code:    code,
		Message: message,
	}
	
	statusCode := http.StatusInternalServerError
	switch code {
	case CodeInvalidRequest:
		statusCode = http.StatusBadRequest
	case CodeUnauthorized:
		statusCode = http.StatusUnauthorized
	case CodeForbidden:
		statusCode = http.StatusForbidden
	case CodeNotFound:
		statusCode = http.StatusNotFound
	case CodeConflict:
		statusCode = http.StatusConflict
	case CodeTooManyRequests:
		statusCode = http.StatusTooManyRequests
	}
	
	c.JSON(statusCode, response)
}

// BadRequest 400响应
func BadRequest(c *gin.Context, message string) {
	Error(c, CodeInvalidRequest, message)
}

// Unauthorized 401响应
func Unauthorized(c *gin.Context, message string) {
	Error(c, CodeUnauthorized, message)
}

// Forbidden 403响应
func Forbidden(c *gin.Context, message string) {
	Error(c, CodeForbidden, message)
}

// NotFound 404响应
func NotFound(c *gin.Context, message string) {
	Error(c, CodeNotFound, message)
}

// InternalServerError 500响应
func InternalServerError(c *gin.Context, message string) {
	Error(c, CodeError, message)
}