package main

import (
	"fmt"
	"os"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"path/filepath"
	"bufio"
)

func main() {
	fmt.Println("🔐 JWT Key Verification Across Services")
	fmt.Println("=======================================")
	
	// 1. 检查环境变量
	envSecret := os.Getenv("JWT_SECRET")
	fmt.Printf("🌍 Environment JWT_SECRET: %s\n", envSecret)
	if envSecret != "" {
		hasher := sha256.New()
		hasher.Write([]byte(envSecret))
		envHash := hex.EncodeToString(hasher.Sum(nil))
		fmt.Printf("🔐 Environment Secret Hash: %s...\n", envHash[:16])
	}
	
	// 2. 扫描用户服务代码中的JWT密钥使用
	fmt.Println("\n📋 Scanning User Service JWT Key Usage...")
	scanJWTUsage("/mnt/d/暑期项目/HorizonCloudService-main/services/user-service")
	
	// 3. 扫描文件服务代码中的JWT密钥使用
	fmt.Println("\n📋 Scanning File Service JWT Key Usage...")
	scanJWTUsage("/mnt/d/暑期项目/HorizonCloudService-main/services/file-service")
	
	// 4. 比较.env文件
	fmt.Println("\n📋 Comparing Environment Files...")
	
	userEnvFile := "/mnt/d/暑期项目/HorizonCloudService-main/services/user-service/.env.development"
	fileEnvFile := "/mnt/d/暑期项目/HorizonCloudService-main/services/file-service/.env.development"
	
	fmt.Printf("🔍 User Service .env: %s\n", getJWTFromEnvFile(userEnvFile))
	fmt.Printf("🔍 File Service .env: %s\n", getJWTFromEnvFile(fileEnvFile))
	
	// 5. 检查hardcoded密钥
	fmt.Println("\n🚨 Looking for hardcoded JWT secrets...")
	findHardcodedSecrets("/mnt/d/暑期项目/HorizonCloudService-main/services/user-service")
	findHardcodedSecrets("/mnt/d/暑期项目/HorizonCloudService-main/services/file-service")
}

func scanJWTUsage(servicePath string) {
	// 扫描Go文件中的JWT相关代码
	err := filepath.Walk(servicePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		
		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			
			// 寻找JWT相关的代码行
			if strings.Contains(strings.ToLower(line), "jwt") && 
			   (strings.Contains(line, "secret") || strings.Contains(line, "key") || 
			    strings.Contains(line, "JWT_SECRET") || strings.Contains(line, "jwtSecret")) {
				relativePath, _ := filepath.Rel(servicePath, path)
				fmt.Printf("   📄 %s:%d: %s\n", relativePath, lineNum, strings.TrimSpace(line))
			}
		}
		
		return nil
	})
	
	if err != nil {
		fmt.Printf("❌ Error scanning %s: %v\n", servicePath, err)
	}
}

func getJWTFromEnvFile(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Sprintf("File not found: %s", filePath)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "JWT_SECRET=") {
			secret := strings.TrimPrefix(line, "JWT_SECRET=")
			hasher := sha256.New()
			hasher.Write([]byte(secret))
			hash := hex.EncodeToString(hasher.Sum(nil))
			return fmt.Sprintf("%s (hash: %s...)", secret, hash[:16])
		}
	}
	
	return "JWT_SECRET not found in file"
}

func findHardcodedSecrets(servicePath string) {
	hardcodedPatterns := []string{
		"your-jwt-secret-key",
		"your-development-secret-key",
		"jwt-secret",
		"secret-key",
	}
	
	err := filepath.Walk(servicePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		
		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			
			for _, pattern := range hardcodedPatterns {
				if strings.Contains(line, pattern) {
					relativePath, _ := filepath.Rel(servicePath, path)
					fmt.Printf("   🚨 %s:%d: %s\n", relativePath, lineNum, strings.TrimSpace(line))
				}
			}
		}
		
		return nil
	})
	
	if err != nil {
		fmt.Printf("❌ Error scanning %s: %v\n", servicePath, err)
	}
}