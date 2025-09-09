package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"user-service/internal/handlers"
	"user-service/internal/repository"
	"user-service/internal/services"
)

func main() {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 初始化Mock仓库层（用于测试）
	userRepo := repository.NewMockUserRepository()
	
	// 初始化服务层
	userService := services.NewUserService(userRepo, "your-jwt-secret-key")
	
	// 初始化处理器
	authHandler := handlers.NewAuthHandler(userService)
	userHandler := handlers.NewUserHandler(userService)
	adminHandler := handlers.NewAdminHandler(userService)
	quotaHandler := handlers.NewQuotaHandler(userService)
	activityHandler := handlers.NewActivityHandler(userService)

	// 创建路由
	router := gin.New()
	
	// 添加中间件
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// CORS中间件
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "user-service",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		})
	})

	// API路由
	api := router.Group("/api/v1")
	
	// 认证路由 (公开)
	auth := api.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", authHandler.Logout)
	}

	// 用户路由 (需要认证，暂时不使用JWT中间件，便于测试)
	users := api.Group("/users")
	{
		users.GET("/profile", userHandler.GetProfile)
		users.PUT("/profile", userHandler.UpdateProfile)
		users.POST("/change-password", userHandler.ChangePassword)
		users.GET("/quota", userHandler.GetQuota)
		users.GET("/activity", userHandler.GetActivityLogs)
	}

	// 管理员路由 (暂时不使用权限中间件，便于测试)
	admin := api.Group("/admin/users")
	{
		admin.GET("", adminHandler.ListUsers)
		admin.POST("/search", adminHandler.SearchUsers)
		admin.GET("/by-role", adminHandler.GetUsersByRole)
		admin.PUT("/:id/status", adminHandler.UpdateUserStatus)
		admin.PUT("/:id/quota", adminHandler.UpdateUserQuota)
		admin.DELETE("/:id", adminHandler.DeleteUser)
	}
	
	// 配额管理路由
	quota := api.Group("/admin/quota")
	{
		quota.GET("/:user_id", quotaHandler.GetQuotaInfo)
		quota.PUT("/:user_id/used", quotaHandler.UpdateStorageUsed)
		quota.POST("/batch", quotaHandler.BatchUpdateQuota)
		quota.GET("/statistics", quotaHandler.GetQuotaStatistics)
	}
	
	// 活动日志路由
	activity := api.Group("/admin/activity")
	{
		activity.POST("/log", activityHandler.LogActivity)
		activity.GET("/:user_id/logs", activityHandler.GetUserActivityLogs)
		activity.GET("/system", activityHandler.GetSystemActivityLogs)
		activity.GET("/statistics", activityHandler.GetActivityStatistics)
		activity.POST("/batch-delete", activityHandler.BatchDeleteLogs)
	}

	// 启动服务器
	port := 8001
	if p := os.Getenv("PORT"); p != "" {
		port = 8001 // 使用默认端口，或者可以解析环境变量
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("🚀 User Service started successfully on port %d", port)
	log.Printf("📖 API Documentation: http://localhost:%d/api/v1", port)
	log.Printf("💓 Health Check: http://localhost:%d/health", port)
	log.Printf("🔐 Authentication: POST http://localhost:%d/api/v1/auth/register", port)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("📴 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("❌ Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited gracefully")
}