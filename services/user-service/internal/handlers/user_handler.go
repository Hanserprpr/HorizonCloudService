package handlers

import (
	"net/http"
	"strconv"
	"user-service/internal/services"
	
	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService services.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile 获取用户档案
// @Summary 获取用户档案
// @Description 获取当前用户的详细信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response{data=services.UserProfile} "获取成功"
// @Failure 401 {object} Response "未授权"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	// 从中间件获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ErrorResponse(c, http.StatusUnauthorized, "未授权", "无效的用户认证")
		return
	}
	
	// 调用服务层
	profile, err := h.userService.GetUserProfile(c.Request.Context(), userID.(uint))
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "获取用户信息失败", err.Error())
		return
	}
	
	SuccessResponse(c, http.StatusOK, "获取成功", profile)
}

// UpdateProfile 更新用户档案
// @Summary 更新用户档案
// @Description 更新当前用户的个人信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.UpdateProfileRequest true "更新请求"
// @Success 200 {object} Response "更新成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/user/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// 从中间件获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ErrorResponse(c, http.StatusUnauthorized, "未授权", "无效的用户认证")
		return
	}
	
	var req services.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数格式错误", err.Error())
		return
	}
	
	// 调用服务层
	err := h.userService.UpdateUserProfile(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "更新用户信息失败", err.Error())
		return
	}
	
	SuccessResponse(c, http.StatusOK, "更新成功", nil)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户的登录密码
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} Response "修改成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/user/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	// 从中间件获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ErrorResponse(c, http.StatusUnauthorized, "未授权", "无效的用户认证")
		return
	}
	
	var req services.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数格式错误", err.Error())
		return
	}
	
	// 调用服务层
	err := h.userService.ChangePassword(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "修改密码失败", err.Error())
		return
	}
	
	SuccessResponse(c, http.StatusOK, "修改密码成功", nil)
}

// GetQuota 获取用户存储配额
// @Summary 获取存储配额
// @Description 获取当前用户的存储配额信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response{data=models.UserQuota} "获取成功"
// @Failure 401 {object} Response "未授权"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/user/quota [get]
func (h *UserHandler) GetQuota(c *gin.Context) {
	// 从中间件获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ErrorResponse(c, http.StatusUnauthorized, "未授权", "无效的用户认证")
		return
	}
	
	// 调用服务层
	quota, err := h.userService.GetUserQuota(c.Request.Context(), userID.(uint))
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "获取配额信息失败", err.Error())
		return
	}
	
	SuccessResponse(c, http.StatusOK, "获取成功", quota)
}

// GetActivityLogs 获取用户活动日志
// @Summary 获取活动日志
// @Description 获取当前用户的活动历史记录
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} Response{data=services.GetActivityLogsResponse} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/user/activity-logs [get]
func (h *UserHandler) GetActivityLogs(c *gin.Context) {
	// 从中间件获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ErrorResponse(c, http.StatusUnauthorized, "未授权", "无效的用户认证")
		return
	}
	
	// 解析查询参数
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	
	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}
	
	req := &services.GetActivityLogsRequest{
		UserID:   userID.(uint),
		Page:     page,
		PageSize: pageSize,
	}
	
	// 调用服务层
	result, err := h.userService.GetUserActivityLogs(c.Request.Context(), req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "获取活动日志失败", err.Error())
		return
	}
	
	SuccessResponse(c, http.StatusOK, "获取成功", result)
}