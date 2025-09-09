package routing

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"gateway/config"
	"gateway/internal/auth"
)

// SetupRoutes 配置路由转发规则
func SetupRoutes(router *gin.Engine, cfg *config.Config) {
	// API 组 - 所有微服务路由
	api := router.Group("/api/v1")

	// 用户认证路由 (无需认证)
	authGroup := api.Group("/auth")
	{
		userServiceProxy := createReverseProxy(cfg.Services["user-service"])
		authGroup.POST("/login", gin.WrapH(userServiceProxy))
		authGroup.POST("/register", gin.WrapH(userServiceProxy))
	}

	// 受保护的路由 (需要认证)
	protected := api.Group("")
	protected.Use(auth.JWTMiddleware(cfg.JWT.Secret))

	// 用户管理路由
	userGroup := protected.Group("/users")
	{
		userServiceProxy := createReverseProxy(cfg.Services["user-service"])
		userGroup.Any("/*path", gin.WrapH(userServiceProxy))
	}

	// 权限管理路由
	permissionGroup := protected.Group("/permissions")
	{
		permissionServiceProxy := createReverseProxy(cfg.Services["permission-service"])
		permissionGroup.Any("/*path", gin.WrapH(permissionServiceProxy))
	}

	// 文件管理路由
	fileGroup := protected.Group("/files")
	{
		fileServiceProxy := createReverseProxy(cfg.Services["file-service"])
		fileGroup.Any("/*path", gin.WrapH(fileServiceProxy))
	}

	// AI处理路由
	aiGroup := protected.Group("/ai")
	{
		aiServiceProxy := createReverseProxy(cfg.Services["ai-service"])
		aiGroup.Any("/*path", gin.WrapH(aiServiceProxy))
	}

	// 模型管理路由
	modelGroup := protected.Group("/models")
	{
		modelServiceProxy := createReverseProxy(cfg.Services["model-service"])
		modelGroup.Any("/*path", gin.WrapH(modelServiceProxy))
	}

	// 搜索服务路由
	searchGroup := protected.Group("/search")
	{
		searchServiceProxy := createReverseProxy(cfg.Services["search-service"])
		searchGroup.Any("/*path", gin.WrapH(searchServiceProxy))
	}

	// 分享服务路由
	shareGroup := protected.Group("/shares")
	{
		shareServiceProxy := createReverseProxy(cfg.Services["share-service"])
		shareGroup.Any("/*path", gin.WrapH(shareServiceProxy))
	}

	// 通知服务路由
	notificationGroup := protected.Group("/notifications")
	{
		notificationServiceProxy := createReverseProxy(cfg.Services["notification-service"])
		notificationGroup.Any("/*path", gin.WrapH(notificationServiceProxy))
	}

	// 公共分享访问 (无需认证)
	api.GET("/public/shares/:token", func(c *gin.Context) {
		shareServiceProxy := createReverseProxy(cfg.Services["share-service"])
		shareServiceProxy.ServeHTTP(c.Writer, c.Request)
	})
}

// createReverseProxy 创建反向代理
func createReverseProxy(serviceConfig config.ServiceConfig) *httputil.ReverseProxy {
	target := fmt.Sprintf("http://%s:%d", serviceConfig.Host, serviceConfig.Port)
	targetURL, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// 自定义请求修改
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		// 移除网关前缀，转发到目标服务
		if serviceConfig.Prefix != "" {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, serviceConfig.Prefix)
		}
		
		// 添加原始主机头
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Host = targetURL.Host
	}

	return proxy
}