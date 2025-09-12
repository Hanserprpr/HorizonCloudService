package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("🚀 启动最小化文件服务...")

	// 设置环境变量（如果没有设置的话）
	if os.Getenv("USER_SERVICE_BASE_URL") == "" {
		err := os.Setenv("USER_SERVICE_BASE_URL", "http://localhost:8001")
		if err != nil {
			return
		}
	}
	if os.Getenv("JWT_SECRET") == "" {
		err := os.Setenv("JWT_SECRET", "your-development-secret-key")
		if err != nil {
			return
		}
	}
	if os.Getenv("SERVER_PORT") == "" {
		err := os.Setenv("SERVER_PORT", "8002")
		if err != nil {
			return
		}
	}

	fmt.Printf("环境变量设置:\n")
	fmt.Printf("  USER_SERVICE_BASE_URL: %s\n", os.Getenv("USER_SERVICE_BASE_URL"))
	fmt.Printf("  JWT_SECRET: %s\n", os.Getenv("JWT_SECRET"))
	fmt.Printf("  SERVER_PORT: %s\n", os.Getenv("SERVER_PORT"))

	// 创建简单的Gin服务器
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 健康检查端点
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":           "healthy",
			"service":          "file-service",
			"user_service_url": os.Getenv("USER_SERVICE_BASE_URL"),
		})
	})

	// 基本的API端点
	api := r.Group("/api/v1")
	{
		api.GET("/files", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "files endpoint", "data": []interface{}{}})
		})

		api.GET("/folders", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "folders endpoint", "data": []interface{}{}})
		})

		api.GET("/quota/status", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "quota status",
				"data": gin.H{
					"used_storage":  0,
					"total_storage": 5368709120,
					"used_files":    0,
					"total_files":   10000,
				},
			})
		})

		api.POST("/files/upload", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "upload endpoint ready"})
		})

		api.POST("/files/batch", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "batch operation ready"})
		})

		api.POST("/folders", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "folder creation ready"})
		})
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8002"
	}

	fmt.Printf("🌐 文件服务启动在端口 %s\n", port)
	fmt.Printf("🔗 健康检查: http://localhost:%s/api/v1/health\n", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
