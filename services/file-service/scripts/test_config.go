package main

import (
	"fmt"
	"os"

	"file-service/internal/config"
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
	
	// 检查是否有隐藏字符
	for i, char := range cfg.JWT.Secret {
		if char < 32 || char > 126 {
			fmt.Printf("在位置 %d 发现隐藏字符: %d\n", i, char)
		}
	}
}