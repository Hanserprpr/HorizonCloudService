package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"file-service/internal/config"
	"file-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)
	
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}
	
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
	
	// 创建Gin引擎
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	
	// 创建受保护的路由
	api := r.Group("/api/v1")
	api.Use(authMiddleware.AuthRequired())
	{
		api.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
	}
	
	// 用户服务实际生成的令牌
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbkBleGFtcGxlLmNvbSIsInJvbGUiOiJhZG1pbiIsInN0dWRlbnRfaWQiOiJhZG1pbiIsInJvbGVfaWQiOjEsInR5cGUiOiJhY2Nlc3MiLCJpc3MiOiJob3Jpem9uLWNsb3VkIiwic3ViIjoiXHUwMDAxIiwiZXhwIjoxNzU3NTE3Nzc3LCJuYmYiOjE3NTc1MTA1NzcsImlhdCI6MTc1NzUxMDU3N30.XuFsHjXEJgHtre-zu722vrbtR5h-RESabEo2oOdAJpA"
	
	fmt.Printf("创建HTTP请求...\n")
	fmt.Printf("令牌长度: %d\n", len(tokenString))
	fmt.Printf("令牌内容: %.50s...\n", tokenString)
	
	// 创建HTTP请求
	req, _ := http.NewRequest("GET", "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set("Content-Type", "application/json")
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	fmt.Printf("执行HTTP请求...\n")
	r.ServeHTTP(w, req)
	
	// 输出结果
	fmt.Printf("状态码: %d\n", w.Code)
	fmt.Printf("响应内容: %s\n", w.Body.String())
}