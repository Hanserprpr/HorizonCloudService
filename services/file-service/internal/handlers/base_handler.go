package handlers

import (
	"file-service/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BaseHandler 基础处理器，提供通用功能
type BaseHandler struct {
	services *services.Services
}

// NewBaseHandler 创建基础处理器
func NewBaseHandler(services *services.Services) *BaseHandler {
	return &BaseHandler{
		services: services,
	}
}

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Type    string `json:"type,omitempty"`
	Details string `json:"details,omitempty"`
}

// PaginationRequest 分页请求
type PaginationRequest struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=20" binding:"min=1,max=100"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Total    int64       `json:"total"`
	Pages    int         `json:"pages"`
	Data     interface{} `json:"data"`
}

// Success 成功响应
func (h *BaseHandler) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func (h *BaseHandler) Error(c *gin.Context, code int, message string, err error) {
	response := Response{
		Code:    code,
		Message: message,
	}

	if err != nil {
		if serviceErr := services.GetServiceError(err); serviceErr != nil {
			response.Error = &ErrorInfo{
				Type:    serviceErr.Type,
				Details: serviceErr.Details,
			}
		} else {
			response.Error = &ErrorInfo{
				Details: err.Error(),
			}
		}
	}

	c.JSON(code, response)
}

// BadRequest 400错误
func (h *BaseHandler) BadRequest(c *gin.Context, message string, err error) {
	h.Error(c, http.StatusBadRequest, message, err)
}

// Unauthorized 401错误
func (h *BaseHandler) Unauthorized(c *gin.Context, message string, err error) {
	h.Error(c, http.StatusUnauthorized, message, err)
}

// Forbidden 403错误
func (h *BaseHandler) Forbidden(c *gin.Context, message string, err error) {
	h.Error(c, http.StatusForbidden, message, err)
}

// NotFound 404错误
func (h *BaseHandler) NotFound(c *gin.Context, message string, err error) {
	h.Error(c, http.StatusNotFound, message, err)
}

// InternalError 500错误
func (h *BaseHandler) InternalError(c *gin.Context, message string, err error) {
	h.Error(c, http.StatusInternalServerError, message, err)
}

// Paginate 分页响应
func (h *BaseHandler) Paginate(c *gin.Context, data interface{}, total int64, req *PaginationRequest) {
	pages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data: PaginationResponse{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
			Pages:    pages,
			Data:     data,
		},
	})
}

// GetUserID 从上下文中获取用户ID
func (h *BaseHandler) GetUserID(c *gin.Context) uint {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
		if id, ok := userID.(float64); ok {
			return uint(id)
		}
		if id, ok := userID.(string); ok {
			if parsedID, err := strconv.ParseUint(id, 10, 32); err == nil {
				return uint(parsedID)
			}
		}
	}
	return 0
}

// GetUserIDParam 从URL参数获取用户ID
func (h *BaseHandler) GetUserIDParam(c *gin.Context) (uint, error) {
	param := c.Param("user_id")
	if param == "" {
		return 0, services.NewServiceError("MissingParameter", "user_id parameter is required", nil)
	}
	
	userID, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		return 0, services.NewServiceError("InvalidParameter", "invalid user_id parameter", err)
	}
	
	return uint(userID), nil
}

// GetIDParam 从URL参数获取ID
func (h *BaseHandler) GetIDParam(c *gin.Context, name string) (uint, error) {
	param := c.Param(name)
	if param == "" {
		return 0, services.NewServiceError("MissingParameter", name+" parameter is required", nil)
	}
	
	id, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		return 0, services.NewServiceError("InvalidParameter", "invalid "+name+" parameter", err)
	}
	
	return uint(id), nil
}

// GetQueryParam 获取查询参数
func (h *BaseHandler) GetQueryParam(c *gin.Context, name string) string {
	return c.Query(name)
}

// GetQueryParamInt 获取整数查询参数
func (h *BaseHandler) GetQueryParamInt(c *gin.Context, name string, defaultValue int) int {
	if value := c.Query(name); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetQueryParamBool 获取布尔查询参数
func (h *BaseHandler) GetQueryParamBool(c *gin.Context, name string, defaultValue bool) bool {
	if value := c.Query(name); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// ValidateRequest 验证请求参数
func (h *BaseHandler) ValidateRequest(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return services.NewServiceError("ValidationError", "request validation failed", err)
	}
	return nil
}

// BindQuery 绑定查询参数
func (h *BaseHandler) BindQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return services.NewServiceError("ValidationError", "query parameter validation failed", err)
	}
	return nil
}

// HandleServiceError 处理服务层错误
func (h *BaseHandler) HandleServiceError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	serviceErr := services.GetServiceError(err)
	if serviceErr == nil {
		h.InternalError(c, "Internal server error", err)
		return
	}

	switch serviceErr.Type {
	case services.ErrorTypeInvalidRequest, services.ErrorTypeMissingParam, services.ErrorTypeInvalidParam:
		h.BadRequest(c, serviceErr.Message, err)
	case services.ErrorTypePermissionDenied, services.ErrorTypeAccessDenied:
		h.Forbidden(c, serviceErr.Message, err)
	case services.ErrorTypeUnauthorized:
		h.Unauthorized(c, serviceErr.Message, err)
	case services.ErrorTypeResourceNotFound, services.ErrorTypeFileNotFound, services.ErrorTypeFolderNotFound:
		h.NotFound(c, serviceErr.Message, err)
	default:
		h.InternalError(c, serviceErr.Message, err)
	}
}

// RequireAuth 检查认证状态
func (h *BaseHandler) RequireAuth(c *gin.Context) uint {
	userID := h.GetUserID(c)
	if userID == 0 {
		h.Unauthorized(c, "Authentication required", nil)
		c.Abort()
		return 0
	}
	return userID
}

// RequireAuthWithRole 检查认证状态和角色权限
func (h *BaseHandler) RequireAuthWithRole(c *gin.Context, allowedRoles ...string) uint {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return 0
	}

	userRole, exists := c.Get("role")
	if !exists {
		h.Forbidden(c, "User role not found", nil)
		c.Abort()
		return 0
	}

	role, ok := userRole.(string)
	if !ok {
		h.Forbidden(c, "Invalid user role", nil)
		c.Abort()
		return 0
	}

	// 检查角色权限
	for _, allowedRole := range allowedRoles {
		if role == allowedRole || role == "admin" { // admin拥有所有权限
			return userID
		}
	}

	h.Forbidden(c, "Insufficient permissions", nil)
	c.Abort()
	return 0
}

// RequireAdmin 要求管理员权限
func (h *BaseHandler) RequireAdmin(c *gin.Context) uint {
	return h.RequireAuthWithRole(c, "admin")
}

// CheckOwnershipOrAdmin 检查资源所有权或管理员权限
func (h *BaseHandler) CheckOwnershipOrAdmin(c *gin.Context, resourceUserID uint) bool {
	currentUserID := h.GetUserID(c)
	if currentUserID == 0 {
		h.Unauthorized(c, "Authentication required", nil)
		c.Abort()
		return false
	}

	// 检查是否是资源拥有者
	if currentUserID == resourceUserID {
		return true
	}

	// 检查是否是管理员
	userRole, exists := c.Get("role")
	if exists {
		if role, ok := userRole.(string); ok && role == "admin" {
			return true
		}
	}

	h.Forbidden(c, "Access denied: insufficient permissions", nil)
	c.Abort()
	return false
}

// GetUsername 从上下文获取用户名
func (h *BaseHandler) GetUsername(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		if name, ok := username.(string); ok {
			return name
		}
	}
	return ""
}

// GetUserRole 从上下文获取用户角色
func (h *BaseHandler) GetUserRole(c *gin.Context) string {
	if role, exists := c.Get("role"); exists {
		if roleStr, ok := role.(string); ok {
			return roleStr
		}
	}
	return ""
}

// IsAdmin 检查是否是管理员
func (h *BaseHandler) IsAdmin(c *gin.Context) bool {
	return h.GetUserRole(c) == "admin"
}

// HasRole 检查是否拥有指定角色
func (h *BaseHandler) HasRole(c *gin.Context, requiredRole string) bool {
	role := h.GetUserRole(c)
	return role == requiredRole || role == "admin"
}

// LogRequest 记录请求信息
func (h *BaseHandler) LogRequest(c *gin.Context) {
	// TODO: 实现请求日志记录
}