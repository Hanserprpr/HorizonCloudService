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
	"gateway/internal/auth"
	"gateway/internal/monitoring"
	"gateway/internal/ratelimit"
	"gateway/internal/routing"
	"gateway/config"
)

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 初始化 Redis 连接 (用于限流)
	ratelimit.InitRedis(cfg.Redis)

	// 初始化监控
	monitoring.Init()

	// 设置 Gin 模式
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()

	// 添加中间件
	router.Use(gin.Recovery())
	router.Use(monitoring.PrometheusMiddleware())
	router.Use(ratelimit.Middleware())

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "gateway",
			"timestamp": time.Now().Unix(),
		})
	})

	// 配置路由转发
	routing.SetupRoutes(router, cfg)

	// 启动服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("API Gateway started on port %d", cfg.Port)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 给服务器 5 秒时间来关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}