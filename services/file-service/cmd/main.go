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

	"file-service/internal/config"
	"file-service/internal/middleware"
	"file-service/internal/models"
	"file-service/internal/repository"
	"file-service/internal/routes"
	"file-service/internal/services"
	"file-service/internal/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to load config, using defaults: %v", err)
		cfg = getDefaultConfig()
	}

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

	// 初始化存储
	storageClient, err := initStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
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

	quotaConfig := &middleware.QuotaConfig{
		DefaultStorageQuota:  cfg.Quota.DefaultStorageQuota,
		DefaultFileCount:     cfg.Quota.DefaultFileCount,
		CheckInterval:        time.Duration(cfg.Quota.CheckIntervalMinutes) * time.Minute,
		GraceBuffer:          cfg.Quota.GraceBuffer,
		EnableWarnings:       cfg.Quota.EnableWarnings,
		WarningThreshold:     cfg.Quota.WarningThreshold,
	}

	// 初始化路由
	routerConfig := &routes.RouterConfig{
		Services:          serviceContainer,
		Repository:        repo,
		UserServiceClient: userServiceClient,
		AuthConfig:        authConfig,
		QuotaConfig:       quotaConfig,
	}

	router := routes.NewRouter(routerConfig)
	engine := router.GetEngine()

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

// getDefaultConfig 获取默认配置
func getDefaultConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:        "file-service",
			Version:     "1.0.0",
			Environment: "development",
			Debug:       true,
		},
		Server: config.ServerConfig{
			Port:                8083,
			ReadTimeoutSeconds:  30,
			WriteTimeoutSeconds: 30,
			IdleTimeoutSeconds:  120,
			MaxHeaderBytes:      1 << 20, // 1MB
		},
		Database: config.DatabaseConfig{
			Host:                     "localhost",
			Port:                     5432,
			User:                     "postgres",
			Password:                 "password",
			Name:                     "file_service",
			SSLMode:                  "disable",
			Timezone:                 "UTC",
			LogLevel:                 "info",
			MaxIdleConns:             10,
			MaxOpenConns:             100,
			ConnMaxLifetimeMinutes:   60,
		},
		JWT: config.JWTConfig{
			Secret:          "your-secret-key",
			ExpirationHours: 24,
		},
		Storage: config.StorageConfig{
			Backend: "local",
			Local: config.LocalConfig{
				RootPath: "./uploads",
			},
		},
		Quota: config.QuotaConfig{
			DefaultStorageQuota:    5 * 1024 * 1024 * 1024, // 5GB
			DefaultFileCount:       10000,
			CheckIntervalMinutes:   5,
			GraceBuffer:            0.1,
			EnableWarnings:         true,
			WarningThreshold:       0.8,
		},
		UserService: config.UserServiceConfig{
			BaseURL:                "",
			APIKey:                 "",
			TimeoutSeconds:         30,
			RetryCount:             3,
			RetryIntervalSeconds:   1,
			EnableCircuitBreaker:   false,
		},
	}
}

// initDatabase 初始化数据库连接
func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// 配置GORM日志级别
	var logLevel logger.LogLevel
	switch cfg.Database.LogLevel {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 尝试连接PostgreSQL，如果失败则使用SQLite
	if cfg.Database.Host != "" && cfg.Database.Host != "localhost" {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
			cfg.Database.Host,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Name,
			cfg.Database.Port,
			cfg.Database.SSLMode,
			cfg.Database.Timezone,
		)

		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err != nil {
			log.Printf("Failed to connect to PostgreSQL: %v, falling back to SQLite", err)
		} else {
			log.Println("Connected to PostgreSQL database")
			// 配置连接池（仅对PostgreSQL有效）
			if sqlDB, err := db.DB(); err == nil {
				sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
				sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
				sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetimeMinutes) * time.Minute)
			}
		}
	}

	// 如果PostgreSQL连接失败或未配置，使用SQLite
	if db == nil {
		log.Println("Using SQLite database for development")
		db, err = gorm.Open(sqlite.Open("file_service.db"), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
		}
	}

	log.Println("Database connected successfully")
	return db, nil
}

// runMigrations 运行数据库迁移
func runMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")
	
	err := db.AutoMigrate(
		&models.File{},
		&models.Folder{},
		&models.Thumbnail{},
		&models.UploadSession{},
		&models.UploadChunk{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// initStorage 初始化存储客户端
func initStorage(cfg *config.Config) (storage.Storage, error) {
	switch cfg.Storage.Backend {
	case "minio":
		return storage.NewMinIOStorage(&storage.Config{
			Type:            storage.StorageTypeMinIO,
			Endpoint:        cfg.Storage.MinIO.Endpoint,
			AccessKeyID:     cfg.Storage.MinIO.AccessKey,
			SecretAccessKey: cfg.Storage.MinIO.SecretKey,
			Bucket:          cfg.Storage.MinIO.BucketName,
			Region:          cfg.Storage.MinIO.Region,
			UseSSL:          cfg.Storage.MinIO.UseSSL,
		})
	case "s3":
		return storage.NewS3Storage(&storage.Config{
			Type:            storage.StorageTypeS3,
			Region:          cfg.Storage.S3.Region,
			AccessKeyID:     cfg.Storage.S3.AccessKey,
			SecretAccessKey: cfg.Storage.S3.SecretKey,
			Bucket:          cfg.Storage.S3.BucketName,
			Endpoint:        cfg.Storage.S3.Endpoint,
		})
	case "local":
		return storage.NewLocalStorage(&storage.Config{
			Type:      storage.StorageTypeLocal,
			LocalPath: cfg.Storage.Local.RootPath,
		})
	default:
		log.Printf("Unknown storage backend '%s', using local storage", cfg.Storage.Backend)
		return storage.NewLocalStorage(&storage.Config{
			Type:      storage.StorageTypeLocal,
			LocalPath: "./uploads",
		})
	}
}

// initUserServiceClient 初始化用户服务客户端
func initUserServiceClient(cfg *config.Config) services.UserServiceClient {
	if cfg.App.Environment == "development" || cfg.UserService.BaseURL == "" {
		// 开发环境使用Mock客户端
		log.Println("Using mock user service client")
		return services.NewMockUserServiceClient()
	}

	// 生产环境使用真实客户端
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

// initServices 初始化服务层
func initServices(repo *repository.Repository, storageClient storage.Storage) *services.Services {
	// 创建服务配置
	serviceConfig := &services.ServicesConfig{
		DefaultChunkSize:    5 * 1024 * 1024, // 5MB默认分片大小
		ThumbnailSizes:      []string{"small", "medium", "large"},
		ThumbnailQuality:    85,
		ThumbnailTimeout:    30 * time.Second,
		MaxBatchSize:        100,
		BatchConcurrency:    5,
		SearchLimit:         100,
	}

	return services.NewServices(repo, storageClient, serviceConfig)
}