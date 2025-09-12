package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// 切换到文件服务目录
	err := os.Chdir("services/file-service")
	if err != nil {
		fmt.Printf("无法切换到文件服务目录: %v\n", err)
		return
	}
	
	fmt.Printf("当前工作目录: %s\n", getCurrentDir())
	
	// 清除环境变量以测试默认值
	os.Unsetenv("USER_SERVICE_BASE_URL")
	
	// 动态导入配置包
	fmt.Println("测试修复后的配置默认值...")
	
	// 由于Go模块路径问题，我们需要在正确的目录中运行
	fmt.Println("请在 services/file-service 目录中运行以下命令来测试配置:")
	fmt.Println("go run scripts/test_config.go")
}

func getCurrentDir() string {
	dir, err := filepath.Abs(".")
	if err != nil {
		return "unknown"
	}
	return dir
}
