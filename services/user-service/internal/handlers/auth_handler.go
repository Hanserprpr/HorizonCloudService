package handlers

import (
	"net/http"
	"user-service/internal/services"
	
	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userService services.UserService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(userService services.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body services.RegisterRequest true "注册请求"
// @Success 201 {object} Response{data=services.AuthResponse} "注册成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 409 {object} Response "用户已存在"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数格式错误", err.Error())
		return
	}
	
	// 获取客户端信息
	req.IPAddress = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")
	
	// 调用服务层
	result, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		// 根据错误类型返回不同状态码
		if err.Error() == "学号已存在" || err.Error() == "邮箱已存在" {
			ErrorResponse(c, http.StatusConflict, "用户已存在", err.Error())
		} else {
			ErrorResponse(c, http.StatusBadRequest, "注册失败", err.Error())
		}
		return
	}
	
	SuccessResponse(c, http.StatusCreated, "注册成功", result)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户账户登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body services.LoginRequest true "登录请求"
// @Success 200 {object} Response{data=services.AuthResponse} "登录成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "认证失败"
// @Failure 403 {object} Response "用户被禁用"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数格式错误", err.Error())
		return
	}
	
	// 获取客户端信息
	req.IPAddress = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")
	
	// 调用服务层
	result, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		// 根据错误类型返回不同状态码
		switch err.Error() {
		case "用户名或密码错误":
			ErrorResponse(c, http.StatusUnauthorized, "认证失败", err.Error())
		case "用户已被禁用":
			ErrorResponse(c, http.StatusForbidden, "用户被禁用", err.Error())
		default:
			ErrorResponse(c, http.StatusBadRequest, "登录失败", err.Error())
		}
		return
	}
	
	SuccessResponse(c, http.StatusOK, "登录成功", result)
}

// RefreshToken 刷新访问令牌
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新令牌请求"
// @Success 200 {object} Response{data=services.AuthResponse} "刷新成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "令牌无效"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数格式错误", err.Error())
		return
	}
	
	// 调用服务层
	result, err := h.userService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "令牌刷新失败", err.Error())
		return
	}
	
	SuccessResponse(c, http.StatusOK, "令牌刷新成功", result)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户账户登出
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response "登出成功"
// @Failure 401 {object} Response "未授权"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从中间件获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ErrorResponse(c, http.StatusUnauthorized, "未授权", "无效的用户认证")
		return
	}
	
	// 调用服务层
	err := h.userService.Logout(c.Request.Context(), userID.(uint))
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "登出失败", err.Error())
		return
	}
	
	SuccessResponse(c, http.StatusOK, "登出成功", nil)
}

// ValidateToken 验证访问令牌
// @Summary 验证访问令牌
// @Description 验证访问令牌的有效性
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response "令牌有效"
// @Failure 401 {object} Response "令牌无效"
// @Router /api/v1/auth/validate [get]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	// 从中间件获取用户ID (如果中间件通过了，说明令牌有效)
	userID, exists := c.Get("user_id")
	if !exists {
		ErrorResponse(c, http.StatusUnauthorized, "令牌无效", "无效的访问令牌")
		return
	}
	
	SuccessResponse(c, http.StatusOK, "令牌有效", gin.H{
		"user_id": userID,
		"valid":   true,
	})
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改用户密码
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} Response "密码修改成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/users/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// 从中间件获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ErrorResponse(c, http.StatusUnauthorized, "未授权", "无效的用户认证")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数格式错误", err.Error())
		return
	}

	// 验证新密码和确认密码是否一致
	if req.NewPassword != req.ConfirmPassword {
		ErrorResponse(c, http.StatusBadRequest, "参数错误", "新密码和确认密码不一致")
		return
	}

	// 调用服务层
	changeReq := &services.ChangePasswordRequest{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}

	err := h.userService.ChangePassword(c.Request.Context(), userID.(uint), changeReq)
	if err != nil {
		if err.Error() == "原密码错误" {
			ErrorResponse(c, http.StatusBadRequest, "密码修改失败", err.Error())
		} else {
			ErrorResponse(c, http.StatusInternalServerError, "密码修改失败", err.Error())
		}
		return
	}

	SuccessResponse(c, http.StatusOK, "密码修改成功", nil)
}

// ChangePasswordRequest 修改密码请求结构
type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// RefreshTokenRequest 刷新令牌请求结构
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}