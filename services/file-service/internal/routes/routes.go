package routes

import (
	"file-service/internal/handlers"
	"file-service/internal/middleware"
	"file-service/internal/repository"
	"file-service/internal/services"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Router 路由管理器
type Router struct {
	engine       *gin.Engine
	services     *services.Services
	authMW       *middleware.AuthMiddleware
	quotaMW      *middleware.QuotaMiddleware
	handlers     *Handlers
}

// Handlers 所有处理器的集合
type Handlers struct {
	File      *handlers.FileHandler
	Upload    *handlers.UploadHandler
	Folder    *handlers.FolderHandler
	Thumbnail *handlers.ThumbnailHandler
	Health    *handlers.HealthHandler
}

// RouterConfig 路由配置
type RouterConfig struct {
	Services          *services.Services
	Repository        *repository.Repository
	UserServiceClient services.UserServiceClient
	AuthConfig        *middleware.AuthConfig
	QuotaConfig       *middleware.QuotaConfig
}

// NewRouter 创建路由管理器
func NewRouter(config *RouterConfig) *Router {
	// 创建Gin引擎
	engine := gin.New()
	
	// 创建中间件
	authMW := middleware.NewAuthMiddleware(config.AuthConfig)
	quotaMW := middleware.NewQuotaMiddleware(config.QuotaConfig, config.UserServiceClient, config.Repository.File)

	// 创建处理器
	handlerInstances := &Handlers{
		File:      handlers.NewFileHandler(config.Services),
		Upload:    handlers.NewUploadHandler(config.Services),
		Folder:    handlers.NewFolderHandler(config.Services),
		Thumbnail: handlers.NewThumbnailHandler(config.Services),
		Health:    handlers.NewHealthHandler(config.Services),
	}

	router := &Router{
		engine:   engine,
		services: config.Services,
		authMW:   authMW,
		quotaMW:  quotaMW,
		handlers: handlerInstances,
	}

	router.setupMiddleware()
	router.setupRoutes()

	return router
}

// GetEngine 获取Gin引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// setupMiddleware 设置中间件
func (r *Router) setupMiddleware() {
	// 基础中间件
	r.engine.Use(gin.Logger())
	r.engine.Use(gin.Recovery())

	// CORS中间件
	r.engine.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// 超时中间件
	r.engine.Use(func(c *gin.Context) {
		// 为每个请求设置超时
		timeout := 30 * time.Second
		
		// 对于上传请求，设置更长的超时时间
		if c.Request.URL.Path == "/api/v1/upload/chunk" || 
		   c.Request.URL.Path == "/api/v1/upload/simple" {
			timeout = 5 * time.Minute
		}
		
		// 应用超时（示例实现）
		_ = timeout
		
		c.Request = c.Request.WithContext(c.Request.Context())
		c.Next()
	})

	// 请求ID中间件
	r.engine.Use(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	})
}

// setupRoutes 设置路由
func (r *Router) setupRoutes() {
	api := r.engine.Group("/api/v1")

	// 健康检查路由（无需认证）
	r.setupHealthRoutes(api)

	// 认证相关路由
	r.setupAuthRoutes(api)

	// 需要认证的路由
	authenticated := api.Group("")
	authenticated.Use(r.authMW.AuthRequired())
	{
		// 文件相关路由
		r.setupFileRoutes(authenticated)
		
		// 上传相关路由
		r.setupUploadRoutes(authenticated)
		
		// 文件夹相关路由
		r.setupFolderRoutes(authenticated)
		
		// 缩略图相关路由
		r.setupThumbnailRoutes(authenticated)

		// 配额相关路由
		r.setupQuotaRoutes(authenticated)
	}

	// 管理员路由
	admin := api.Group("/admin")
	admin.Use(r.authMW.AuthRequired())
	admin.Use(r.authMW.AdminRequired())
	{
		r.setupAdminRoutes(admin)
	}

	// 公共文件访问路由（可选认证）
	public := api.Group("/public")
	public.Use(r.authMW.OptionalAuth())
	{
		r.setupPublicRoutes(public)
	}
}

// setupHealthRoutes 设置健康检查路由
func (r *Router) setupHealthRoutes(api *gin.RouterGroup) {
	health := api.Group("/health")
	{
		health.GET("", r.handlers.Health.Health)
		health.GET("/ready", r.handlers.Health.Ready)
		health.GET("/live", r.handlers.Health.Live)
		health.GET("/metrics", r.handlers.Health.Metrics)
		health.GET("/stats", r.handlers.Health.Stats)
		health.GET("/version", r.handlers.Health.Version)
	}
}

// setupAuthRoutes 设置认证路由
func (r *Router) setupAuthRoutes(api *gin.RouterGroup) {
	auth := api.Group("/auth")
	{
		// 这些路由通常由用户服务提供，这里只是占位符
		auth.POST("/login", r.handleAuthPlaceholder)
		auth.POST("/logout", r.handleAuthPlaceholder)
		auth.POST("/refresh", r.handleAuthPlaceholder)
		auth.GET("/me", r.authMW.AuthRequired(), r.handleUserInfo)
	}
}

// setupFileRoutes 设置文件路由
func (r *Router) setupFileRoutes(api *gin.RouterGroup) {
	files := api.Group("/files")
	{
		// 文件基本操作
		files.GET("", r.handlers.File.ListFiles)
		files.POST("", r.quotaMW.CheckUploadQuota(0), r.handlers.File.UploadFile) // 动态检查大小
		files.GET("/:id", r.handlers.File.GetFile)
		files.PUT("/:id", r.handlers.File.UpdateFile)
		files.DELETE("/:id", r.handlers.File.DeleteFile)

		// 文件操作
		files.GET("/:id/download", r.handlers.File.DownloadFile)
		files.PUT("/:id/move", r.handlers.File.MoveFile)
		files.POST("/:id/copy", r.handlers.File.CopyFile)

		// 文件版本控制
		files.GET("/:id/versions", r.handlers.File.GetFileVersions)
		files.POST("/versions/:version_id/restore", r.handlers.File.RestoreFileVersion)

		// 批量操作
		files.POST("/batch", r.handlers.File.BatchOperation)

		// 搜索和统计
		files.GET("/search", r.handlers.File.SearchFiles)
		files.GET("/duplicates", r.handlers.File.GetDuplicateFiles)
		files.POST("/duplicates/cleanup", r.handlers.File.CleanupDuplicates)
		files.GET("/stats", r.handlers.File.GetUserStats)
		files.GET("/storage-stats", r.handlers.File.GetStorageStats)
	}
}

// setupUploadRoutes 设置上传路由
func (r *Router) setupUploadRoutes(api *gin.RouterGroup) {
	upload := api.Group("/upload")
	{
		// 简单上传
		upload.POST("/simple", r.quotaMW.CheckUploadQuota(0), r.handlers.Upload.SimpleUpload)

		// 分片上传
		upload.POST("/initiate", r.quotaMW.CheckUploadQuota(0), r.handlers.Upload.InitiateUpload)
		upload.POST("/chunk", r.handlers.Upload.UploadChunk)
		upload.POST("/:session_id/complete", r.handlers.Upload.CompleteUpload)
		upload.DELETE("/:session_id/abort", r.handlers.Upload.AbortUpload)

		// 批量上传
		upload.POST("/batch/initiate", r.quotaMW.CheckBatchUploadQuota(), r.handlers.Upload.BatchInitiateUpload)

		// 上传会话管理
		upload.GET("/sessions", r.handlers.Upload.ListUploadSessions)
		upload.GET("/:session_id", r.handlers.Upload.GetUploadSession)
		upload.GET("/:session_id/progress", r.handlers.Upload.GetUploadProgress)

		// 断点续传
		upload.POST("/:session_id/resume", r.handlers.Upload.ResumeUpload)
		upload.POST("/:session_id/pause", r.handlers.Upload.PauseUpload)
		upload.POST("/:session_id/resume-from-pause", r.handlers.Upload.ResumeUploadFromPause)

		// 上传统计
		upload.GET("/statistics", r.handlers.Upload.GetUploadStatistics)
	}
}

// setupFolderRoutes 设置文件夹路由
func (r *Router) setupFolderRoutes(api *gin.RouterGroup) {
	folders := api.Group("/folders")
	{
		// 文件夹基本操作
		folders.GET("", r.handlers.Folder.ListFolders)
		folders.POST("", r.handlers.Folder.CreateFolder)
		folders.GET("/:id", r.handlers.Folder.GetFolder)
		folders.PUT("/:id", r.handlers.Folder.UpdateFolder)
		folders.DELETE("/:id", r.handlers.Folder.DeleteFolder)

		// 文件夹操作
		folders.PUT("/:id/move", r.handlers.Folder.MoveFolder)
		folders.POST("/:id/copy", r.handlers.Folder.CopyFolder)
		folders.PUT("/:id/rename", r.handlers.Folder.RenameFolder)

		// 文件夹内容和导航
		folders.GET("/:id/contents", r.handlers.Folder.GetFolderContents)
		folders.GET("/:id/path", r.handlers.Folder.GetFolderPath)
		folders.GET("/tree", r.handlers.Folder.GetFolderTree)
		folders.GET("/by-path", r.handlers.Folder.GetFolderByPath)

		// 文件夹统计
		folders.GET("/:id/stats", r.handlers.Folder.GetFolderStats)
		folders.POST("/:id/sync-stats", r.handlers.Folder.SyncFolderStats)
		folders.POST("/recalculate-stats", r.handlers.Folder.RecalculateAllStats)

		// 系统文件夹
		folders.POST("/system/create", r.handlers.Folder.CreateSystemFolders)
		folders.GET("/system/:type", r.handlers.Folder.GetSystemFolder)

		// 搜索
		folders.GET("/search", r.handlers.Folder.SearchFolders)
	}
}

// setupThumbnailRoutes 设置缩略图路由
func (r *Router) setupThumbnailRoutes(api *gin.RouterGroup) {
	thumbnails := api.Group("/thumbnails")
	{
		// 缩略图生成
		thumbnails.POST("/generate", r.handlers.Thumbnail.GenerateThumbnail)
		thumbnails.POST("/files/:file_id/generate", r.handlers.Thumbnail.GenerateThumbnails)
		thumbnails.POST("/batch/generate", r.handlers.Thumbnail.BatchGenerateThumbnails)
		thumbnails.POST("/files/:file_id/refresh", r.handlers.Thumbnail.RefreshThumbnails)

		// 缩略图获取
		thumbnails.GET("/:id", r.handlers.Thumbnail.GetThumbnail)
		thumbnails.GET("/files/:file_id", r.handlers.Thumbnail.GetFileThumbnails)
		thumbnails.GET("/files/:file_id/info", r.handlers.Thumbnail.GetThumbnailInfo)
		thumbnails.GET("/files/:file_id/url/:size", r.handlers.Thumbnail.GetThumbnailURL)

		// 缩略图访问
		thumbnails.GET("/files/:file_id/serve/:size", r.handlers.Thumbnail.ServeThumbnail)
		thumbnails.GET("/files/:file_id/download/:size", r.handlers.Thumbnail.DownloadThumbnail)
		thumbnails.GET("/files/:file_id/preview", r.handlers.Thumbnail.PreviewThumbnail)

		// 缩略图管理
		thumbnails.DELETE("/:id", r.handlers.Thumbnail.DeleteThumbnail)
		thumbnails.GET("/stats", r.handlers.Thumbnail.GetThumbnailStats)
	}
}

// setupQuotaRoutes 设置配额路由
func (r *Router) setupQuotaRoutes(api *gin.RouterGroup) {
	quota := api.Group("/quota")
	{
		quota.GET("/status", r.quotaMW.GetQuotaStatus)
	}
}

// setupAdminRoutes 设置管理员路由
func (r *Router) setupAdminRoutes(api *gin.RouterGroup) {
	// 用户文件管理
	users := api.Group("/users")
	{
		users.GET("/:user_id/files", r.handlers.File.ListFiles)
		users.GET("/:user_id/folders", r.handlers.Folder.ListFolders)
		users.GET("/:user_id/stats", r.handlers.File.GetUserStats)
		users.GET("/:user_id/storage-stats", r.handlers.File.GetStorageStats)
		users.DELETE("/:user_id/files/:id", r.handlers.File.DeleteFile)
	}

	// 系统统计
	api.GET("/system/stats", r.handlers.Health.Stats)
	api.GET("/system/metrics", r.handlers.Health.Metrics)
}

// setupPublicRoutes 设置公共路由
func (r *Router) setupPublicRoutes(api *gin.RouterGroup) {
	// 公共文件访问（通过分享链接等）
	api.GET("/files/:id/download", r.handlers.File.DownloadFile)
	api.GET("/thumbnails/:id/preview", r.handlers.Thumbnail.PreviewThumbnail)
}

// handleAuthPlaceholder 认证占位符处理器
func (r *Router) handleAuthPlaceholder(c *gin.Context) {
	c.JSON(501, gin.H{
		"code":    501,
		"message": "Authentication endpoint should be handled by user service",
		"error": gin.H{
			"type": "NOT_IMPLEMENTED",
			"description": "This endpoint is implemented by the user service",
		},
	})
}

// handleUserInfo 用户信息处理器
func (r *Router) handleUserInfo(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(401, gin.H{
			"code":    401,
			"message": "Authentication required",
		})
		return
	}

	username := middleware.GetUsername(c)
	role, _ := c.Get("role")

	c.JSON(200, gin.H{
		"code":    200,
		"message": "User info retrieved successfully",
		"data": gin.H{
			"user_id":  userID,
			"username": username,
			"role":     role,
		},
	})
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	// 简单的请求ID生成，实际生产中应该使用更好的实现
	return fmt.Sprintf("%d", time.Now().UnixNano())
}