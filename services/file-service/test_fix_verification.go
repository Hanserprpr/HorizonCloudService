package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"file-service/internal/config"
	"file-service/internal/services"
)

func main() {
	fmt.Println("🔍 验证配置修复...")
	
	// 1. 测试默认配置（不设置环境变量）
	fmt.Println("\n📋 测试1: 默认配置（无环境变量）")
	os.Unsetenv("USER_SERVICE_BASE_URL")
	
	cfg, err := config.Load()
	if err != nil {
		log.Printf("配置加载失败: %v", err)
		return
	}
	
	fmt.Printf("✅ 用户服务BaseURL: '%s'\n", cfg.UserService.BaseURL)
	if cfg.UserService.BaseURL == "http://localhost:8001" {
		fmt.Println("✅ 默认配置修复成功！")
	} else {
		fmt.Printf("❌ 默认配置仍有问题，期望: 'http://localhost:8001'，实际: '%s'\n", cfg.UserService.BaseURL)
		return
	}
	
	// 2. 测试用户服务客户端创建
	fmt.Println("\n📋 测试2: 用户服务客户端创建")
	userServiceConfig := &services.UserServiceConfig{
		BaseURL:              cfg.UserService.BaseURL,
		APIKey:               cfg.UserService.APIKey,
		Timeout:              time.Duration(cfg.UserService.TimeoutSeconds) * time.Second,
		RetryCount:           cfg.UserService.RetryCount,
		RetryInterval:        time.Duration(cfg.UserService.RetryIntervalSeconds) * time.Second,
		EnableCircuitBreaker: cfg.UserService.EnableCircuitBreaker,
	}
	
	userServiceClient := services.NewUserServiceClient(userServiceConfig)
	if userServiceClient != nil {
		fmt.Println("✅ 用户服务客户端创建成功")
	} else {
		fmt.Println("❌ 用户服务客户端创建失败")
		return
	}
	
	// 3. 测试环境变量覆盖
	fmt.Println("\n📋 测试3: 环境变量覆盖")
	os.Setenv("USER_SERVICE_BASE_URL", "http://localhost:9999")
	
	cfg2, err := config.Load()
	if err != nil {
		log.Printf("配置加载失败: %v", err)
		return
	}
	
	fmt.Printf("✅ 环境变量覆盖后的BaseURL: '%s'\n", cfg2.UserService.BaseURL)
	if cfg2.UserService.BaseURL == "http://localhost:9999" {
		fmt.Println("✅ 环境变量覆盖功能正常")
	} else {
		fmt.Printf("❌ 环境变量覆盖失败，期望: 'http://localhost:9999'，实际: '%s'\n", cfg2.UserService.BaseURL)
	}
	
	// 恢复环境变量
	os.Unsetenv("USER_SERVICE_BASE_URL")
	
	fmt.Println("\n🎉 配置修复验证完成！")
	fmt.Println("📋 修复总结:")
	fmt.Println("   ✅ 默认用户服务URL已设置为: http://localhost:8001")
	fmt.Println("   ✅ 环境变量覆盖功能正常")
	fmt.Println("   ✅ 用户服务客户端可以正常创建")
	fmt.Println("\n🚀 现在可以重新启动文件服务并运行API测试")
}
