package main

import (
	"fmt"
	"os"
	"time"

	"file-service/internal/config"
	"file-service/internal/middleware"
	"github.com/golang-jwt/jwt/v5"
)

// UserClaims JWT用户声明
type UserClaims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	StudentID string `json:"student_id"`
	RoleID    int    `json:"role_id"`
	Type      string `json:"type"`
	jwt.RegisteredClaims
}

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
	
	fmt.Printf("=== 配置加载结果 ===\n")
	fmt.Printf("环境: %s\n", cfg.App.Environment)
	fmt.Printf("JWT密钥: '%s' (长度: %d)\n", cfg.JWT.Secret, len(cfg.JWT.Secret))
	fmt.Printf("用户服务BaseURL: '%s'\n", cfg.UserService.BaseURL)
	
	// 创建认证中间件配置
	authConfig := &middleware.AuthConfig{
		JWTSecret:     []byte(cfg.JWT.Secret),
		JWTExpiration: time.Duration(cfg.JWT.ExpirationHours) * time.Hour,
		SkipPaths: []string{
			"/health",
			"/api/v1/health",
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/metrics",
			"/favicon.ico",
		},
	}
	
	// 创建认证中间件
	authMiddleware := middleware.NewAuthMiddleware(authConfig)
	
	// 用户服务实际生成的令牌（来自之前的请求）
	actualToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbkBleGFtcGxlLmNvbSIsInJvbGUiOiJhZG1pbiIsInN0dWRlbnRfaWQiOiJhZG1pbiIsInJvbGVfaWQiOjEsInR5cGUiOiJhY2Nlc3MiLCJpc3MiOiJob3Jpem9uLWNsb3VkIiwic3ViIjoiXHUwMDAxIiwiZXhwIjoxNzU3NTE3Nzc3LCJuYmYiOjE3NTc1MTA1NzcsImlhdCI6MTc1NzUxMDU3N30.XuFsHjXEJgHtre-zu722vrbtR5h-RESabEo2oOdAJpA"
	
	fmt.Printf("\n=== 实际令牌信息 ===\n")
	fmt.Printf("令牌字符串: %.50s...\n", actualToken)
	fmt.Printf("令牌长度: %d\n", len(actualToken))
	
	// 使用认证中间件验证令牌
	fmt.Printf("\n=== 开始验证过程 ===\n")
	claims, err := authMiddleware.ValidateTokenForTest(actualToken)
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