package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"file-service/internal/handlers"
	"file-service/internal/middleware"
	"file-service/internal/models"
	"file-service/internal/repository"
	"file-service/internal/services"
	"file-service/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	// 加载环境变量
	if err := godotenv.Load(".env.development"); err != nil {
		log.Printf("Warning: Could not load .env.development file: %v", err)
	}
}

func main() {
	fmt.Println("🚀 Starting File Service HTTP Server (Simple Mode)...")

	// 1. 设置Gin为Release模式
	gin.SetMode(gin.ReleaseMode)

	// 2. 初始化数据库
	fmt.Println("1. Initializing SQLite database...")
	db, err := gorm.Open(sqlite.Open("file-service.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. 运行数据库迁移
	fmt.Println("2. Running database migrations...")
	err = db.AutoMigrate(
		&models.File{},
		&models.Folder{},
		&models.Thumbnail{},
		&models.UploadSession{},
		&models.UploadChunk{},
	)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// 4. 初始化上传目录
	fmt.Println("3. Initializing upload directories...")
	uploadDir := "./uploads"
	thumbnailDir := filepath.Join(uploadDir, "thumbnails")
	
	// 创建上传目录
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}
	if err := os.MkdirAll(thumbnailDir, 0755); err != nil {
		log.Fatalf("Failed to create thumbnail directory: %v", err)
	}

	// 5. 初始化存储
	fmt.Println("4. Initializing local storage...")
	localStorage, err := storage.NewLocalStorage(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: uploadDir,
	})
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// 6. 初始化仓库层
	fmt.Println("5. Initializing repository...")
	repo := repository.NewRepository(db)

	// 7. 初始化服务层
	fmt.Println("6. Initializing services...")
	config := &services.ServicesConfig{
		DefaultChunkSize:    5 * 1024 * 1024, // 5MB
		ThumbnailSizes:      []string{"small", "medium", "large"},
		ThumbnailQuality:    85,
		ThumbnailTimeout:    30 * time.Second,
		MaxBatchSize:        100,
		BatchConcurrency:    5,
		SearchLimit:         100,
	}

	serviceContainer := services.NewServices(repo, localStorage, config)

	// 8. 初始化处理器层
	fmt.Println("7. Initializing handlers...")
	handlerContainer := handlers.NewHandlers(serviceContainer)

	// 9. 初始化Gin路由器
	fmt.Println("8. Setting up HTTP router...")
	router := gin.New()

	// 10. 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS中间件
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// JWT中间件(简化版，只验证token格式)
	jwtMiddleware := middleware.NewJWTAuthMiddleware()

	// 11. 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   "file-service",
			"version":   "1.0.0",
			"port":      "8002",
			"timestamp": time.Now().Unix(),
		})
	})

	router.GET("/health/ready", func(c *gin.Context) {
		// 检查数据库连接
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(503, gin.H{
				"status": "not ready",
				"error":  "database connection failed",
			})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			c.JSON(503, gin.H{
				"status": "not ready",
				"error":  "database ping failed",
			})
			return
		}

		c.JSON(200, gin.H{
			"status":      "ready",
			"service":     "file-service",
			"database":    "connected",
			"storage":     "local",
			"upload_dir":  uploadDir,
			"components": gin.H{
				"file_service":      "ready",
				"folder_service":    "ready",
				"upload_service":    "ready",
				"thumbnail_service": "ready",
			},
		})
	})

	// 12. 静态文件服务
	router.Static("/uploads", "./uploads")
	router.Static("/thumbnails", "./uploads/thumbnails")

	// 13. API路由组
	api := router.Group("/api/v1")
	{
		// 文件相关路由
		files := api.Group("/files")
		files.Use(jwtMiddleware.Authenticate()) // 需要认证
		{
			files.GET("", handlerContainer.File.ListFiles)
			files.POST("", handlerContainer.File.UploadFile)
			files.GET("/:id", handlerContainer.File.GetFile)
			files.PUT("/:id", handlerContainer.File.UpdateFile)
			files.DELETE("/:id", handlerContainer.File.DeleteFile)
			files.POST("/:id/copy", handlerContainer.File.CopyFile)
			files.POST("/:id/move", handlerContainer.File.MoveFile)
			files.GET("/:id/download", handlerContainer.File.DownloadFile)
			files.POST("/batch", handlerContainer.File.BatchOperation)
			files.GET("/search", handlerContainer.File.SearchFiles)
			files.GET("/stats/user", handlerContainer.File.GetUserStats)
			files.GET("/stats/storage", handlerContainer.File.GetStorageStats)
		}

		// 文件夹相关路由
		folders := api.Group("/folders")
		folders.Use(jwtMiddleware.Authenticate()) // 需要认证
		{
			folders.POST("", handlerContainer.Folder.CreateFolder)
			folders.GET("/:id", handlerContainer.Folder.GetFolder)
			folders.PUT("/:id", handlerContainer.Folder.UpdateFolder)
			folders.DELETE("/:id", handlerContainer.Folder.DeleteFolder)
			folders.GET("/:id/contents", handlerContainer.Folder.GetFolderContents)
			folders.GET("/tree", handlerContainer.Folder.GetFolderTree)
			folders.GET("/:id/stats", handlerContainer.Folder.GetFolderStats)
		}

		// 上传相关路由
		uploads := api.Group("/upload")
		uploads.Use(jwtMiddleware.Authenticate()) // 需要认证
		{
			uploads.POST("/simple", handlerContainer.Upload.SimpleUpload)
			uploads.POST("/initiate", handlerContainer.Upload.InitiateUpload)
			uploads.POST("/chunk", handlerContainer.Upload.UploadChunk)
			uploads.POST("/complete/:session_id", handlerContainer.Upload.CompleteUpload)
			uploads.POST("/abort/:session_id", handlerContainer.Upload.AbortUpload)
			uploads.GET("/progress/:session_id", handlerContainer.Upload.GetUploadProgress)
			uploads.GET("/sessions", handlerContainer.Upload.ListUploadSessions)
		}

		// 缩略图相关路由
		thumbnails := api.Group("/thumbnails")
		thumbnails.Use(jwtMiddleware.Authenticate()) // 需要认证
		{
			thumbnails.POST("/:file_id/generate", handlerContainer.Thumbnail.GenerateThumbnails)
			thumbnails.GET("/:file_id", handlerContainer.Thumbnail.GetFileThumbnails)
			thumbnails.GET("/:file_id/:size", handlerContainer.Thumbnail.GetThumbnail)
			thumbnails.DELETE("/:file_id", handlerContainer.Thumbnail.DeleteThumbnail)
		}
	}

	// 14. 启动服务器
	port := ":8002"
	fmt.Printf("\n🎉 File Service HTTP Server started successfully!\n")
	fmt.Printf("📍 Server running on port %s\n", port)
	fmt.Printf("🔗 Health check: http://localhost%s/health\n", port)
	fmt.Printf("🔗 Ready check: http://localhost%s/health/ready\n", port)
	fmt.Printf("📁 Upload directory: %s\n", uploadDir)
	fmt.Printf("🗄️  Database: file-service.db\n")
	fmt.Printf("🌐 API Base: http://localhost%s/api/v1\n", port)
	fmt.Printf("\n📋 Available API Endpoints:\n")
	fmt.Printf("   • GET    /api/v1/files              - List files\n")
	fmt.Printf("   • POST   /api/v1/files              - Create file\n")
	fmt.Printf("   • GET    /api/v1/files/:id          - Get file\n")
	fmt.Printf("   • DELETE /api/v1/files/:id          - Delete file\n")
	fmt.Printf("   • POST   /api/v1/folders            - Create folder\n")
	fmt.Printf("   • GET    /api/v1/folders/:id/contents - Get folder contents\n")
	fmt.Printf("   • POST   /api/v1/upload/simple      - Simple upload\n")
	fmt.Printf("   • POST   /api/v1/upload/initiate    - Initiate chunked upload\n")
	fmt.Printf("   • POST   /api/v1/upload/chunk       - Upload chunk\n")
	fmt.Printf("   • POST   /api/v1/upload/complete    - Complete upload\n")
	fmt.Printf("\n⚠️  注意：所有API需要JWT认证（除健康检查外）\n")
	fmt.Printf("💡 前端已配置正确的端口：http://localhost:8002\n\n")

	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}