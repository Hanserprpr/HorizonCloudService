package main

import (
	"fmt"
	"os"

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
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}
	
	fmt.Printf("环境: %s\n", cfg.App.Environment)
	fmt.Printf("JWT密钥: '%s' (长度: %d)\n", cfg.JWT.Secret, len(cfg.JWT.Secret))
	fmt.Printf("用户服务BaseURL: '%s'\n", cfg.UserService.BaseURL)
	
	// 创建认证中间件配置
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
	
	fmt.Printf("认证中间件配置JWT密钥: '%s' (长度: %d)\n", string(authConfig.JWTSecret), len(authConfig.JWTSecret))
	
	// 创建认证中间件
	authMiddleware := middleware.NewAuthMiddleware(authConfig)
	
	// 测试JWT验证 - 使用实际的用户服务令牌
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbkBleGFtcGxlLmNvbSIsInJvbGUiOiJhZG1pbiIsInN0dWRlbnRfaWQiOiJhZG1pbiIsInJvbGVfaWQiOjEsInR5cGUiOiJhY2Nlc3MiLCJpc3MiOiJob3Jpem9uLWNsb3VkIiwic3ViIjoiXHUwMDAxIiwiZXhwIjoxNzU3NTE1MzExLCJuYmYiOjE3NTc1MDgxMTEsImlhdCI6MTc1NzUwODExMX0.al1ClSy7ExW014jXWouYnc_uKyLN16lriNEwPawCTMg"
	
	fmt.Printf("\n开始验证JWT令牌:\n")
	fmt.Printf("令牌字符串: %.50s...\n", tokenString)
	
	// 调用验证方法
	claims, err := authMiddleware.ValidateTokenForTest(tokenString)
	if err != nil {
		fmt.Printf("验证失败: %v\n", err)
		return
	}
	
	fmt.Printf("验证成功!\n")
	fmt.Printf("用户ID: %d, 用户名: %s, 邮箱: %s\n", claims.UserID, claims.Username, claims.Email)
}