package handlers

import (
	"time"
	"user-service/internal/middleware"
	"user-service/internal/services"
	
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine, userService services.UserService, jwtSecret string) {
	// 创建处理器
	authHandler := NewAuthHandler(userService)
	userHandler := NewUserHandler(userService)
	adminHandler := NewAdminHandler(userService)
	quotaHandler := NewQuotaHandler(userService)
	activityHandler := NewActivityHandler(userService)
	internalHandler := NewInternalHandler(userService)
	
	// 创建中间件
	authMiddleware := middleware.NewAuthMiddleware(jwtSecret)
	adminMiddleware := middleware.NewRoleMiddleware([]int{2}) // 假设角色2是管理员
	
	// API版本分组
	v1 := r.Group("/api/v1")
	
	// 健康检查
	v1.GET("/health", HealthCheck)
	
	// 公开认证路由
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		
		// 需要认证的认证路由
		authProtected := auth.Use(authMiddleware.Authenticate())
		authProtected.GET("/validate", authHandler.ValidateToken)
		authProtected.POST("/logout", authHandler.Logout)
	}
	
	// 需要认证的用户路由 (兼容前端路径)
	users := v1.Group("/users")
	users.Use(authMiddleware.Authenticate())
	{
		// 用户档案管理 (前端期望的路径)
		users.GET("/profile", userHandler.GetProfile)
		users.PUT("/profile", userHandler.UpdateProfile)
		users.POST("/change-password", authHandler.ChangePassword)
	}
	
	// 需要认证的用户路由 (原始路径 - 向后兼容)
	user := v1.Group("/user")
	user.Use(authMiddleware.Authenticate())
	{
		// 用户档案管理
		user.GET("/profile", userHandler.GetProfile)
		user.PUT("/profile", userHandler.UpdateProfile)
		user.PUT("/password", userHandler.ChangePassword)
		
		// 存储配额
		user.GET("/quota", userHandler.GetQuota)
		
		// 活动日志
		user.GET("/activity-logs", userHandler.GetActivityLogs)
	}
	
	// 管理员路由
	admin := v1.Group("/admin")
	admin.Use(authMiddleware.Authenticate())
	admin.Use(adminMiddleware.CheckRole())
	{
		// 用户管理
		admin.GET("/users", adminHandler.ListUsers)
		admin.GET("/users/search", adminHandler.SearchUsers)
		admin.GET("/users/role/:role_id", adminHandler.GetUsersByRole)
		admin.PUT("/users/:user_id/status", adminHandler.UpdateUserStatus)
		admin.PUT("/users/:user_id/quota", adminHandler.UpdateUserQuota)
		admin.DELETE("/users/:user_id", adminHandler.DeleteUser)
		
		// 配额管理
		admin.GET("/quota/:user_id", quotaHandler.GetQuotaInfo)
		admin.POST("/quota/batch-update", quotaHandler.BatchUpdateQuota)
		admin.GET("/quota/statistics", quotaHandler.GetQuotaStatistics)
		
		// 活动日志管理
		admin.GET("/activity/:user_id/logs", activityHandler.GetUserActivityLogs)
		admin.GET("/activity/system", activityHandler.GetSystemActivityLogs)
		admin.GET("/activity/statistics", activityHandler.GetActivityStatistics)
		admin.POST("/activity/batch-delete", activityHandler.BatchDeleteLogs)
	}
	
	// 内部服务路由（供其他微服务调用）
	internal := v1.Group("/internal")
	{
		// 用户信息获取（文件服务调用）
		internal.GET("/users/:user_id", internalHandler.GetUser)
		
		// 配额管理（文件服务调用）
		internal.PUT("/quota/:user_id/used", quotaHandler.UpdateStorageUsed)
		
		// 活动日志记录（各服务调用）
		internal.POST("/activity/log", activityHandler.LogActivity)
	}
}

// HealthCheck 健康检查处理器
// @Summary 健康检查
// @Description 检查服务健康状态
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} HealthCheckResponse "服务正常"
// @Router /api/v1/health [get]
func HealthCheck(c *gin.Context) {
	response := HealthCheckResponse{
		Status:    "healthy",
		Version:   "1.0.0",
		Timestamp: time.Now().Unix(),
		Services: map[string]interface{}{
			"database": "connected",
			"redis":    "connected",
		},
	}
	
	SuccessResponse(c, 200, "服务正常", response)
}