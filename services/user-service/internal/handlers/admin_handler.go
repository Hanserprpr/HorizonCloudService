package handlers

import (
	"net/http"
	"strconv"
	"user-service/internal/services"
	
	"github.com/gin-gonic/gin"
)

// AdminHandler 管理员处理器
type AdminHandler struct {
	userService services.UserService
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler(userService services.UserService) *AdminHandler {
	return &AdminHandler{
		userService: userService,
	}
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 管理员获取系统用户列表
// @Tags 管理员
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param status query int false "用户状态" Enums(0, 1, 2)
// @Success 200 {object} Response{data=services.ListUsersResponse} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 403 {object} Response "权限不足"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
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
	
	status := 0
	if s := c.Query("status"); s != "" {
		if parsed, err := strconv.Atoi(s); err == nil && (parsed >= 0 && parsed <= 2) {
			status = parsed
		}
	}
	
	req := &services.ListUsersRequest{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
	}
	
	// 调用服务层
	result, err := h.userService.ListUsers(c.Request.Context(), req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "获取用户列表失败", err.Error())
		return
	}
	
	SuccessResponse(c, http.StatusOK, "获取成功", result)
}

// SearchUsers 搜索用户
// @Summary 搜索用户
// @Description 管理员根据关键词搜索用户
// @Tags 管理员
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param keyword query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} Response{data=services.SearchUsersResponse} "搜索成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 403 {object} Response "权限不足"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/admin/users/search [get]
func (h *AdminHandler) SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误", "搜索关键词不能为空")
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
	
	req := &services.SearchUsersRequest{
		Keyword:  keyword,
		Page:     page,
		PageSize: pageSize,
	}
	
	// 调用服务层
	result, err := h.userService.SearchUsers(c.Request.Context(), req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "搜索用户失败", err.Error())
		return
	}
	
	SuccessResponse(c, http.StatusOK, "搜索成功", result)
}

// GetUsersByRole 根据角色获取用户
// @Summary 根据角色获取用户
// @Description 管理员根据角色获取用户列表
// @Tags 管理员
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param role_id path int true "角色ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} Response{data=services.GetUsersByRoleResponse} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 403 {object} Response "权限不足"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/admin/users/role/{role_id} [get]
func (h *AdminHandler) GetUsersByRole(c *gin.Context) {
	// 解析路径参数
	roleIDStr := c.Param("role_id")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil || roleID <= 0 {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误", "无效的角色ID")
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
	
	req := &services.GetUsersByRoleRequest{
		RoleID:   roleID,
		Page:     page,
		PageSize: pageSize,
	}
	
	// 调用服务层
	result, err := h.userService.GetUsersByRole(c.Request.Context(), req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "获取角色用户失败", err.Error())
		return
	}
	
	SuccessResponse(c, http.StatusOK, "获取成功", result)
}

// UpdateUserStatus 更新用户状态
// @Summary 更新用户状态
// @Description 管理员更新用户状态（启用/禁用）
// @Tags 管理员
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "用户ID"
// @Param request body UpdateUserStatusRequest true "状态更新请求"
// @Success 200 {object} Response "更新成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 403 {object} Response "权限不足"
// @Failure 404 {object} Response "用户不存在"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/admin/users/{user_id}/status [put]
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	// 解析路径参数
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误", "无效的用户ID")
		return
	}
	
	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数格式错误", err.Error())
		return
	}
	
	// 调用服务层
	err = h.userService.UpdateUserStatus(c.Request.Context(), uint(userID), req.Status)
	if err != nil {
		if err.Error() == "用户不存在" {
			ErrorResponse(c, http.StatusNotFound, "用户不存在", err.Error())
		} else {
			ErrorResponse(c, http.StatusBadRequest, "更新用户状态失败", err.Error())
		}
		return
	}
	
	SuccessResponse(c, http.StatusOK, "更新成功", nil)
}

// UpdateUserQuota 更新用户配额
// @Summary 更新用户配额
// @Description 管理员更新用户存储配额
// @Tags 管理员
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "用户ID"
// @Param request body UpdateUserQuotaRequest true "配额更新请求"
// @Success 200 {object} Response "更新成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 403 {object} Response "权限不足"
// @Failure 404 {object} Response "用户不存在"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/admin/users/{user_id}/quota [put]
func (h *AdminHandler) UpdateUserQuota(c *gin.Context) {
	// 解析路径参数
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误", "无效的用户ID")
		return
	}
	
	var req UpdateUserQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数格式错误", err.Error())
		return
	}
	
	// 调用服务层
	err = h.userService.UpdateUserQuota(c.Request.Context(), uint(userID), req.Quota)
	if err != nil {
		if err.Error() == "用户不存在" {
			ErrorResponse(c, http.StatusNotFound, "用户不存在", err.Error())
		} else {
			ErrorResponse(c, http.StatusBadRequest, "更新用户配额失败", err.Error())
		}
		return
	}
	
	SuccessResponse(c, http.StatusOK, "更新成功", nil)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 管理员删除用户（软删除）
// @Tags 管理员
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "用户ID"
// @Success 200 {object} Response "删除成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 403 {object} Response "权限不足"
// @Failure 404 {object} Response "用户不存在"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/admin/users/{user_id} [delete]
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	// 解析路径参数
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误", "无效的用户ID")
		return
	}
	
	// 调用服务层
	err = h.userService.DeleteUser(c.Request.Context(), uint(userID))
	if err != nil {
		if err.Error() == "用户不存在" {
			ErrorResponse(c, http.StatusNotFound, "用户不存在", err.Error())
		} else {
			ErrorResponse(c, http.StatusInternalServerError, "删除用户失败", err.Error())
		}
		return
	}
	
	SuccessResponse(c, http.StatusOK, "删除成功", nil)
}

// 请求结构体

// UpdateUserStatusRequest 更新用户状态请求
type UpdateUserStatusRequest struct {
	Status int `json:"status" binding:"required,oneof=1 2"` // 1:正常 2:禁用
}

// UpdateUserQuotaRequest 更新用户配额请求
type UpdateUserQuotaRequest struct {
	Quota int64 `json:"quota" binding:"required,min=1"` // 配额，字节为单位
}