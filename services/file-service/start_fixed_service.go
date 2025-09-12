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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"file-service/internal/config"
	"file-service/internal/middleware"
	"file-service/internal/models"
	"file-service/internal/repository"
	"file-service/internal/routes"
	"file-service/internal/services"
	"file-service/internal/storage"
)

func main() {
	fmt.Println("🚀 启动修复后的文件服务...")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to load config, using defaults: %v", err)
		cfg = getDefaultConfig()
	}

	// 打印配置信息用于调试
	log.Printf("Loaded configuration:")
	log.Printf("  App.Name: %s", cfg.App.Name)
	log.Printf("  Server.Port: %d", cfg.Server.Port)
	log.Printf("  UserService.BaseURL: %s", cfg.UserService.BaseURL)
	log.Printf("  UserService.TimeoutSeconds: %d", cfg.UserService.TimeoutSeconds)

	// 设置运行模式
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化数据库
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 运行数据库迁移
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// 初始化存储客户端
	storageClient, err := initStorageClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage client: %v", err)
	}

	// 初始化仓库层
	repo := repository.NewRepository(db)

	// 初始化用户服务客户端
	userServiceClient := initUserServiceClient(cfg)

	// 初始化服务层
	serviceContainer := initServices(repo, storageClient)

	// 初始化中间件配置
	authConfig := &middleware.AuthConfig{
		JWTSecret:     []byte(cfg.JWT.Secret),
		JWTExpiration: time.Duration(cfg.JWT.ExpirationHours) * time.Hour,
		SkipPaths: []string{
			"/health",
			"/api/v1/health",
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/metrics",
			"/favicon.ico",
		},
	}

	authMiddleware := middleware.NewAuthMiddleware(authConfig)

	// 初始化配额中间件
	quotaMiddleware := middleware.NewQuotaMiddleware(userServiceClient, &middleware.QuotaConfig{
		DefaultStorageQuota:  cfg.Quota.DefaultStorageQuota,
		DefaultFileCount:     cfg.Quota.DefaultFileCount,
		CheckIntervalMinutes: cfg.Quota.CheckIntervalMinutes,
		GraceBuffer:          cfg.Quota.GraceBuffer,
		EnableWarnings:       cfg.Quota.EnableWarnings,
		WarningThreshold:     cfg.Quota.WarningThreshold,
	})

	// 创建Gin引擎
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// 设置路由
	routes.SetupRoutes(engine, serviceContainer, authMiddleware, quotaMiddleware)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:        engine,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:    time.Duration(cfg.Server.IdleTimeoutSeconds) * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// 启动服务器
	go func() {
		log.Printf("File service starting on port %d", cfg.Server.Port)
		log.Printf("Environment: %s", cfg.App.Environment)
		log.Printf("Storage backend: %s", cfg.Storage.Backend)
		log.Printf("User service URL: %s", cfg.UserService.BaseURL)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号进行优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 创建上下文进行优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// 辅助函数
func getDefaultConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:        "file-service",
			Version:     "1.0.0",
			Environment: "development",
			Debug:       true,
		},
		Server: config.ServerConfig{
			Port:               8002,
			ReadTimeoutSeconds: 30,
			WriteTimeoutSeconds: 30,
			IdleTimeoutSeconds: 120,
			MaxHeaderBytes:     1048576,
		},
		JWT: config.JWTConfig{
			Secret:          "your-development-secret-key",
			ExpirationHours: 24,
		},
		Storage: config.StorageConfig{
			Backend: "local",
			Local: config.LocalConfig{
				RootPath: "./uploads",
			},
		},
		UserService: config.UserServiceConfig{
			BaseURL:                "http://localhost:8001",
			APIKey:                 "",
			TimeoutSeconds:         30,
			RetryCount:             3,
			RetryIntervalSeconds:   1,
			EnableCircuitBreaker:   false,
		},
	}
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("file-service.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.File{},
		&models.Folder{},
		&models.UploadSession{},
		&models.FileVersion{},
		&models.Share{},
	)
}

func initStorageClient(cfg *config.Config) (storage.Storage, error) {
	return storage.NewLocalStorage(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: cfg.Storage.Local.RootPath,
	})
}

func initUserServiceClient(cfg *config.Config) services.UserServiceClient {
	userServiceConfig := &services.UserServiceConfig{
		BaseURL:              cfg.UserService.BaseURL,
		APIKey:               cfg.UserService.APIKey,
		Timeout:              time.Duration(cfg.UserService.TimeoutSeconds) * time.Second,
		RetryCount:           cfg.UserService.RetryCount,
		RetryInterval:        time.Duration(cfg.UserService.RetryIntervalSeconds) * time.Second,
		EnableCircuitBreaker: cfg.UserService.EnableCircuitBreaker,
	}

	return services.NewUserServiceClient(userServiceConfig)
}

func initServices(repo *repository.Repository, storageClient storage.Storage) *services.Services {
	serviceConfig := &services.ServicesConfig{
		DefaultChunkSize:    5 * 1024 * 1024,
		ThumbnailSizes:      []string{"small", "medium", "large"},
		ThumbnailQuality:    85,
		ThumbnailTimeout:    30 * time.Second,
		MaxBatchSize:        100,
		BatchConcurrency:    5,
		SearchLimit:         100,
	}

	return services.NewServices(repo, storageClient, serviceConfig)
}
