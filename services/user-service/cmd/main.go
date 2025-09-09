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

	"github.com/gin-gonic/gin"
	"user-service/config"
	"user-service/internal/handlers"
	"user-service/internal/repository"
	"user-service/internal/services"
	"user-service/pkg/database"
	"user-service/pkg/logger"
	"user-service/pkg/middleware"
	"user-service/pkg/redis"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 初始化日志
	logger.Init(cfg.Log)

	// 初始化数据库
	db, err := database.Init(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 初始化Redis
	redisClient, err := redis.Init(cfg.Redis)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// 初始化仓库层
	userRepo := repository.NewUserRepository(db)
	
	// 初始化服务层
	userService := services.NewUserService(userRepo, redisClient, cfg.JWT)
	
	// 初始化处理器
	userHandler := handlers.NewUserHandler(userService)

	// 设置Gin模式
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()
	
	// 添加中间件
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "user-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// API路由
	api := router.Group("/api/v1")
	
	// 认证路由 (公开)
	auth := api.Group("/auth")
	{
		auth.POST("/login", userHandler.Login)
		auth.POST("/register", userHandler.Register)
		auth.POST("/refresh", userHandler.RefreshToken)
	}

	// 用户路由 (需要认证)
	users := api.Group("/users")
	users.Use(middleware.JWTAuth(cfg.JWT.Secret))
	{
		users.GET("/profile", userHandler.GetProfile)
		users.PUT("/profile", userHandler.UpdateProfile)
		users.POST("/change-password", userHandler.ChangePassword)
		users.GET("/quota", userHandler.GetQuota)
		users.GET("/activity", userHandler.GetActivity)
	}

	// 管理员路由 (需要管理员权限)
	admin := api.Group("/admin/users")
	admin.Use(middleware.JWTAuth(cfg.JWT.Secret))
	admin.Use(middleware.RequireRole([]int{1, 2})) // 管理员和超级管理员
	{
		admin.GET("", userHandler.ListUsers)
		admin.GET("/:id", userHandler.GetUser)
		admin.PUT("/:id/status", userHandler.UpdateUserStatus)
		admin.PUT("/:id/quota", userHandler.UpdateUserQuota)
		admin.DELETE("/:id", userHandler.DeleteUser)
	}

	// 启动服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("User Service started on port %d", cfg.Port)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}