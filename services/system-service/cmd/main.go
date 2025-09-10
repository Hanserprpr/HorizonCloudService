package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"system-service/internal/handlers"
	"system-service/internal/middleware"
	"system-service/internal/repository"
	"system-service/internal/services"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// 初始化数据库
	db, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化仓库层
	repos := repository.NewRepositories(db)

	// 初始化服务层
	systemServices := services.NewServices(repos)

	// 初始化处理器
	systemHandlers := handlers.NewHandlers(systemServices)

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	if os.Getenv("GIN_MODE") == "debug" {
		gin.SetMode(gin.DebugMode)
	}

	// 创建路由器
	r := gin.New()

	// 添加中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS配置
	config := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(config))

	// 添加认证中间件（简化版）
	authMiddleware := middleware.NewSimpleAuthMiddleware()

	// 健康检查路由（无需认证）
	r.GET("/health", systemHandlers.System.HealthCheck)
	r.GET("/health/ready", systemHandlers.System.ReadinessCheck)

	// API路由
	api := r.Group("/api/v1")
	{
		// 系统管理API
		system := api.Group("/system")
		system.Use(authMiddleware.RequireAuth())
		{
			system.GET("/stats", systemHandlers.System.GetSystemStats)
			system.GET("/health", systemHandlers.System.GetSystemHealth)
			system.POST("/cache/clear", systemHandlers.System.ClearCache)
		}

		// 管理员设置API
		admin := api.Group("/admin")
		admin.Use(authMiddleware.RequireAuth())
		{
			admin.GET("/settings", systemHandlers.System.GetSettings)
			admin.PUT("/settings", systemHandlers.System.UpdateSettings)
			admin.GET("/settings/storage", systemHandlers.System.GetStorageSettings)
			admin.PUT("/settings/storage", systemHandlers.System.UpdateStorageSettings)
			admin.POST("/settings/test-storage", systemHandlers.System.TestStorageSettings)
		}
	}

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		log.Printf("🎉 System Service started successfully!")
		log.Printf("📍 Server running on port :%s", port)
		log.Printf("🔗 Health check: http://localhost:%s/health", port)
		log.Printf("🔗 Ready check: http://localhost:%s/health/ready", port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 5秒超时关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

// initDatabase 初始化数据库
func initDatabase() (*gorm.DB, error) {
	// 使用SQLite进行开发
	db, err := gorm.Open(sqlite.Open("system-service.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(
		&repository.SystemSetting{},
	)
	if err != nil {
		return nil, err
	}

	log.Println("✅ Database initialized successfully")
	return db, nil
}