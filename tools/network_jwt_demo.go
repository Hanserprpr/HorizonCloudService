package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LoginRequest 登录请求 - 匹配用户服务的期望格式
type LoginRequest struct {
	StudentID string `json:"student_id"`
	Password  string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Success   bool        `json:"success"`
	Timestamp int64       `json:"timestamp"`
}

// AuthData JWT认证数据
type AuthData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
	User         struct {
		ID        uint   `json:"id"`
		StudentID string `json:"student_id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
	} `json:"user"`
}

// FilesResponse 文件服务响应
type FilesResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Success bool        `json:"success"`
}

func main() {
	fmt.Println("🌐 Network Layer JWT Authentication Test")
	fmt.Println("========================================")

	// Phase 1: 测试用户服务注册（如果用户不存在）
	fmt.Println("📋 Phase 1: Ensure test user exists...")
	
	registerReq := map[string]interface{}{
		"student_id": "test001",
		"nick_name":  "testuser",
		"password":   "testpassword123",
		"email":      "test@example.com",
	}
	
	regBody, _ := json.Marshal(registerReq)
	regResp, err := http.Post("http://localhost:8001/api/v1/auth/register", 
		"application/json", bytes.NewBuffer(regBody))
	
	if err != nil {
		fmt.Printf("⚠️  Register request failed: %v (user may already exist)\n", err)
	} else {
		defer regResp.Body.Close()
		regRespBody, _ := io.ReadAll(regResp.Body)
		fmt.Printf("📝 Register response: %s\n", string(regRespBody)[:min(200, len(regRespBody))])
	}
	fmt.Println()

	// Phase 2: 从用户服务获取JWT token
	fmt.Println("📋 Phase 2: Get JWT token from user service...")
	
	loginReq := LoginRequest{
		StudentID: "test001",
		Password:  "testpassword123",
	}

	loginBody, err := json.Marshal(loginReq)
	if err != nil {
		panic(fmt.Errorf("failed to marshal login request: %v", err))
	}

	fmt.Printf("🔐 Login request: %s\n", string(loginBody))

	// 发送登录请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post("http://localhost:8001/api/v1/auth/login", 
		"application/json", bytes.NewBuffer(loginBody))
	
	if err != nil {
		panic(fmt.Errorf("login request failed: %v", err))
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Errorf("failed to read response: %v", err))
	}

	fmt.Printf("📊 Login response status: %d\n", resp.StatusCode)
	fmt.Printf("📝 Login response body: %s\n", string(respBody))

	// 解析登录响应
	var loginResp LoginResponse
	if err := json.Unmarshal(respBody, &loginResp); err != nil {
		panic(fmt.Errorf("failed to unmarshal login response: %v", err))
	}

	if resp.StatusCode != 200 || !loginResp.Success {
		panic(fmt.Errorf("login failed: %s", loginResp.Message))
	}

	// 从Data中提取JWT token
	dataBytes, _ := json.Marshal(loginResp.Data)
	var authData AuthData
	if err := json.Unmarshal(dataBytes, &authData); err != nil {
		panic(fmt.Errorf("failed to extract auth data: %v", err))
	}

	accessToken := authData.AccessToken
	if accessToken == "" {
		panic(fmt.Errorf("no access token in response"))
	}

	fmt.Printf("✅ JWT token obtained: %s...\n", accessToken[:30])
	fmt.Printf("👤 User info: ID=%d, StudentID=%s, Username=%s\n", 
		authData.User.ID, authData.User.StudentID, authData.User.Username)
	fmt.Println()

	// Phase 3: 使用JWT token访问文件服务
	fmt.Println("📋 Phase 3: Test file service with JWT token...")

	requestID := fmt.Sprintf("test-network-%d", time.Now().Unix())
	
	// 创建文件服务请求
	req, err := http.NewRequest("GET", "http://localhost:8002/api/v1/files", nil)
	if err != nil {
		panic(fmt.Errorf("failed to create file service request: %v", err))
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-Request-ID", requestID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:3000")

	fmt.Printf("🔍 Request headers:\n")
	for name, values := range req.Header {
		for _, value := range values {
			fmt.Printf("   %s: %s\n", name, value)
		}
	}
	fmt.Println()

	// 发送文件服务请求
	fileResp, err := client.Do(req)
	if err != nil {
		panic(fmt.Errorf("file service request failed: %v", err))
	}
	defer fileResp.Body.Close()

	fileRespBody, err := io.ReadAll(fileResp.Body)
	if err != nil {
		panic(fmt.Errorf("failed to read file service response: %v", err))
	}

	fmt.Printf("📊 File service response status: %d\n", fileResp.StatusCode)
	fmt.Printf("🔍 Response headers:\n")
	for name, values := range fileResp.Header {
		if name == "X-Request-Id" || name == "Access-Control-Expose-Headers" {
			for _, value := range values {
				fmt.Printf("   %s: %s\n", name, value)
			}
		}
	}
	fmt.Printf("📝 File service response: %s\n", string(fileRespBody)[:min(500, len(fileRespBody))])

	// 解析文件服务响应
	var filesResp FilesResponse
	if err := json.Unmarshal(fileRespBody, &filesResp); err != nil {
		fmt.Printf("⚠️  Failed to unmarshal file response: %v\n", err)
	}

	fmt.Println()

	// Phase 4: 结果分析
	fmt.Println("📋 Phase 4: Analysis and Summary...")

	if fileResp.StatusCode == 200 && filesResp.Success {
		fmt.Println("✅ JWT authentication across services working correctly!")
	} else if fileResp.StatusCode == 401 {
		fmt.Println("❌ JWT authentication failed - 401 Unauthorized")
		fmt.Println("🔍 This indicates the file service couldn't validate the JWT token")
		
		// 分析可能的原因
		fmt.Println("🔧 Possible causes:")
		fmt.Println("   1. JWT secret mismatch between services")
		fmt.Println("   2. JWT claims structure mismatch")
		fmt.Println("   3. Token expiration issues")
		fmt.Println("   4. Middleware configuration problems")
	} else {
		fmt.Printf("⚠️  Unexpected response status: %d\n", fileResp.StatusCode)
		fmt.Printf("📝 Response: %s\n", string(fileRespBody))
	}

	fmt.Println()
	fmt.Println("🎯 Test Summary:")
	fmt.Printf("   - User Service Login: %s\n", boolToStatus(resp.StatusCode == 200))
	fmt.Printf("   - JWT Token Generation: %s\n", boolToStatus(accessToken != ""))
	fmt.Printf("   - File Service Authentication: %s\n", boolToStatus(fileResp.StatusCode == 200))
	fmt.Printf("   - Request Tracing: %s\n", boolToStatus(fileResp.Header.Get("X-Request-Id") != ""))
	fmt.Printf("   - CORS Headers: %s\n", boolToStatus(fileResp.Header.Get("Access-Control-Allow-Origin") != ""))

}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func boolToStatus(b bool) string {
	if b {
		return "✅ Working"
	}
	return "❌ Failed"
}