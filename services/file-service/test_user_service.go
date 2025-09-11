package main

import (
	"fmt"
	"os"

	"file-service/internal/config"
	"file-service/internal/services"
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
	fmt.Printf("用户服务BaseURL: '%s'\n", cfg.UserService.BaseURL)
	
	// 初始化用户服务客户端
	var userServiceClient services.UserServiceClient
	if cfg.UserService.BaseURL != "" {
		fmt.Println("使用真实用户服务客户端")
		userServiceConfig := &services.UserServiceConfig{
			BaseURL:              cfg.UserService.BaseURL,
			APIKey:               cfg.UserService.APIKey,
			Timeout:              30,
			RetryCount:           cfg.UserService.RetryCount,
			RetryInterval:        5,
			EnableCircuitBreaker: cfg.UserService.EnableCircuitBreaker,
		}
		userServiceClient = services.NewUserServiceClient(userServiceConfig)
	} else {
		fmt.Println("使用Mock用户服务客户端")
		userServiceClient = services.NewMockUserServiceClient()
	}
	
	// 测试获取用户信息
	fmt.Println("测试获取用户信息...")
	userInfo, err := userServiceClient.GetUser(nil, 1)
	if err != nil {
		fmt.Printf("获取用户信息失败: %v\n", err)
		return
	}
	
	fmt.Printf("获取用户信息成功: ID=%d, Username=%s, Email=%s\n", userInfo.ID, userInfo.Username, userInfo.Email)
}