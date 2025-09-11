package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"file-service/internal/config"
	"file-service/internal/middleware"
)

func main() {
	// 设置环境变量
	os.Setenv("JWT_SECRET", "your-development-secret-key")
	os.Setenv("USER_SERVICE_BASE_URL", "http://localhost:8001")
	
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	
	// 创建Gin引擎
	gin.SetMode(gin.DebugMode)
	r := gin.New()
	
	// 添加日志中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	
	// 创建认证中间件
	authConfig := &middleware.AuthConfig{
		JWTSecret:     []byte(cfg.JWT.Secret),
		JWTExpiration: 24,
		SkipPaths: []string{
			"/health",
			"/api/v1/health",
		},
	}
	
	authMiddleware := middleware.NewAuthMiddleware(authConfig)
	
	// 设置路由
	api := r.Group("/api/v1")
	{
		// 健康检查路由
		health := api.Group("/health")
		{
			health.GET("", func(c *gin.Context) {
				c.JSON(200, gin.H{"status": "healthy"})
			})
		}
		
		// 受保护的路由
		protected := api.Group("")
		protected.Use(authMiddleware.AuthRequired())
		{
			protected.GET("/files", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "files list"})
			})
		}
	}
	
	// 启动服务器
	fmt.Printf("启动简化测试服务器，端口: 8002\n")
	fmt.Printf("JWT密钥: '%s'\n", cfg.JWT.Secret)
	
	if err := r.Run(":8002"); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}