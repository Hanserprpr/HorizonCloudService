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
	
	fmt.Printf("配置加载成功\n")
	fmt.Printf("JWT密钥: '%s' (长度: %d)\n", cfg.JWT.Secret, len(cfg.JWT.Secret))
	
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
	
	// 用户服务实际生成的令牌
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbkBleGFtcGxlLmNvbSIsInJvbGUiOiJhZG1pbiIsInN0dWRlbnRfaWQiOiJhZG1pbiIsInJvbGVfaWQiOjEsInR5cGUiOiJhY2Nlc3MiLCJpc3MiOiJob3Jpem9uLWNsb3VkIiwic3ViIjoiXHUwMDAxIiwiZXhwIjoxNzU3NTE3Nzc3LCJuYmYiOjE3NTc1MTA1NzcsImlhdCI6MTc1NzUxMDU3N30.XuFsHjXEJgHtre-zu722vrbtR5h-RESabEo2oOdAJpA"
	
	fmt.Printf("\n开始验证用户服务实际生成的令牌:\n")
	fmt.Printf("令牌: %.50s...\n", tokenString)
	
	// 调用验证方法
	claims, err := authMiddleware.ValidateTokenForTest(tokenString)
	if err != nil {
		fmt.Printf("验证失败: %v\n", err)
		return
	}
	
	fmt.Printf("验证成功!\n")
	fmt.Printf("用户ID: %d\n", claims.UserID)
	fmt.Printf("用户名: %s\n", claims.Username)
	fmt.Printf("邮箱: %s\n", claims.Email)
	fmt.Printf("角色: %s\n", claims.Role)
}