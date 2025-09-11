package main

import (
	"fmt"
	"os"
	"crypto/sha256"
	"encoding/hex"
)

func main() {
	fmt.Println("🔍 JWT Environment Debug Tool")
	fmt.Println("====================================================")
	
	// 1. 检查环境变量
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		fmt.Println("⚠️  JWT_SECRET environment variable is not set")
		jwtSecret = "your-development-secret-key" // fallback
		fmt.Printf("📝 Using fallback value: %s\n", jwtSecret)
	} else {
		fmt.Printf("✅ JWT_SECRET found: %s\n", jwtSecret)
	}
	
	// 2. 计算JWT secret的哈希值用于验证一致性
	hasher := sha256.New()
	hasher.Write([]byte(jwtSecret))
	secretHash := hex.EncodeToString(hasher.Sum(nil))
	fmt.Printf("🔐 JWT_SECRET SHA256: %s\n", secretHash[:16]+"...")
	
	// 3. 检查其他相关环境变量
	fmt.Println("\n📋 Other Environment Variables:")
	envVars := []string{
		"JWT_EXPIRATION_HOURS",
		"SERVER_PORT", 
		"DATABASE_URL",
		"STORAGE_PATH",
		"DEBUG",
		"GIN_MODE",
	}
	
	for _, envVar := range envVars {
		value := os.Getenv(envVar)
		if value != "" {
			fmt.Printf("   %s=%s\n", envVar, value)
		} else {
			fmt.Printf("   %s=<not set>\n", envVar)
		}
	}
	
	// 4. 输出所有以JWT开头的环境变量
	fmt.Println("\n🔑 JWT-related Environment Variables:")
	for _, env := range os.Environ() {
		if len(env) >= 3 && env[:3] == "JWT" {
			fmt.Printf("   %s\n", env)
		}
	}
	
	// 5. 当前工作目录
	pwd, _ := os.Getwd()
	fmt.Printf("\n📁 Current Directory: %s\n", pwd)
	
	// 6. 检查.env文件存在性
	envFiles := []string{
		".env",
		".env.local", 
		".env.development",
		".env.production",
	}
	
	fmt.Println("\n📄 Environment Files Check:")
	for _, file := range envFiles {
		if _, err := os.Stat(file); err == nil {
			fmt.Printf("   ✅ %s exists\n", file)
		} else {
			fmt.Printf("   ❌ %s not found\n", file)
		}
	}
}