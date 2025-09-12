package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"file-service/internal/config"
	"file-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// 设置环境变量
	os.Setenv("JWT_SECRET", "your-development-secret-key")
	os.Setenv("USER_SERVICE_BASE_URL", "http://localhost:8001")
	
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}
	
	fmt.Printf("环境: %s\n", cfg.App.Environment)
	fmt.Printf("JWT密钥: '%s' (长度: %d)\n", cfg.JWT.Secret, len(cfg.JWT.Secret))
	
	// 创建认证中间件
	authConfig := &middleware.AuthConfig{
		JWTSecret:     []byte(cfg.JWT.Secret),
		JWTExpiration: 24,
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
	
	// 创建Gin引擎
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	// 添加中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	
	// 创建测试路由 - 使用认证中间件
	api := r.Group("/api/v1")
	api.Use(authMiddleware.AuthRequired())
	{
		api.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "认证成功!"})
		})
	}
	
	// 创建测试请求 - 使用用户服务生成的令牌
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbkBleGFtcGxlLmNvbSIsInJvbGUiOiJhZG1pbiIsInN0dWRlbnRfaWQiOiJhZG1pbiIsInJvbGVfaWQiOjEsInR5cGUiOiJhY2Nlc3MiLCJpc3MiOiJob3Jpem9uLWNsb3VkIiwic3ViIjoiXHUwMDAxIiwiZXhwIjoxNzU3NTE1MzExLCJuYmYiOjE3NTc1MDgxMTEsImlhdCI6MTc1NzUwODExMX0.al1ClSy7ExW014jXWouYnc_uKyLN16lriNEwPawCTMg"
	
	req, _ := http.NewRequest("GET", "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	r.ServeHTTP(w, req)
	
	// 输出结果
	fmt.Printf("状态码: %d\n", w.Code)
	fmt.Printf("响应内容: %s\n", w.Body.String())
}