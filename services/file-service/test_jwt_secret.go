package main

import (
	"fmt"
	"os"

	"file-service/internal/config"
)

func main() {
	// 显示所有环境变量中与JWT相关的变量
	fmt.Println("JWT相关环境变量:")
	for _, env := range os.Environ() {
		if len(env) > 4 && env[:4] == "JWT_" {
			fmt.Printf("  %s\n", env)
		}
	}
	
	// 显示JWT_SECRET环境变量的值
	jwtSecret := os.Getenv("JWT_SECRET")
	fmt.Printf("\nJWT_SECRET环境变量值: '%s' (长度: %d)\n", jwtSecret, len(jwtSecret))
	
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}
	
	fmt.Printf("\n配置文件加载结果:\n")
	fmt.Printf("  JWT密钥: '%s' (长度: %d)\n", cfg.JWT.Secret, len(cfg.JWT.Secret))
	fmt.Printf("  环境: %s\n", cfg.App.Environment)
	fmt.Printf("  用户服务BaseURL: '%s'\n", cfg.UserService.BaseURL)
	
	// 检查是否有任何隐藏字符
	fmt.Printf("\nJWT密钥字节分析:\n")
	for i, b := range []byte(cfg.JWT.Secret) {
		if b < 32 || b > 126 {
			fmt.Printf("  字节 %d: %d (非可打印字符)\n", i, b)
		}
	}
	
	// 检查环境变量中的JWT_SECRET是否与配置一致
	if jwtSecret != cfg.JWT.Secret {
		fmt.Printf("\n警告: 环境变量中的JWT_SECRET与配置加载的JWT密钥不一致!\n")
		fmt.Printf("  环境变量: '%s'\n", jwtSecret)
		fmt.Printf("  配置加载: '%s'\n", cfg.JWT.Secret)
	} else {
		fmt.Printf("\n环境变量与配置加载的JWT密钥一致。\n")
	}
}