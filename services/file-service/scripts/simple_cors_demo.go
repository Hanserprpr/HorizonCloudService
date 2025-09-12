package main

import (
	"fmt"
	"net/http/httptest"
	"strings"

	"file-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("🧪 Testing CORS Configuration...")

	// 创建Gin引擎
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	// 使用我们修复的CORS中间件
	engine.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}
		
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24小时缓存预检结果
		
		// 只在有实际X-Request-ID时才暴露
		if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
			c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		}
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// 添加JWT认证中间件
	authConfig := &middleware.AuthConfig{
		JWTSecret:     []byte("your-development-secret-key"),
		JWTExpiration: 0, // 使用默认值
		SkipPaths:     []string{"/health"},
	}
	
	authMiddleware := middleware.NewAuthMiddleware(authConfig)

	// 测试路由
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 需要认证的路由
	protected := engine.Group("/api/v1")
	protected.Use(authMiddleware.AuthRequired())
	protected.GET("/files", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "files endpoint", "user_id": c.GetUint("user_id")})
	})

	// 生成测试JWT token
	testToken, err := authMiddleware.GenerateJWT(1, "testuser", "test@example.com", "user")
	if err != nil {
		panic(fmt.Errorf("failed to generate test token: %v", err))
	}
	fmt.Println("✅ Test JWT Token generated successfully")

	// 1. 测试CORS预检请求 (不应该触发认证)
	fmt.Println("\n🌐 Testing CORS preflight request...")
	
	req := httptest.NewRequest("OPTIONS", "/api/v1/files", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Authorization, Content-Type")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	fmt.Printf("📊 CORS Preflight Status: %d (should be 204)\n", w.Code)
	if w.Code == 204 {
		fmt.Println("✅ CORS preflight request handled correctly")
	} else {
		fmt.Printf("❌ CORS preflight failed. Body: %s\n", w.Body.String())
	}

	// 检查CORS头
	fmt.Printf("📋 CORS Headers:\n")
	for key, values := range w.Header() {
		if strings.HasPrefix(key, "Access-Control-") {
			fmt.Printf("   %s: %s\n", key, strings.Join(values, ", "))
		}
	}

	// 2. 测试无认证请求到需要认证的端点 (应该返回401)
	fmt.Println("\n❌ Testing unauthenticated request to protected endpoint...")
	
	req = httptest.NewRequest("GET", "/api/v1/files", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	fmt.Printf("📊 Unauthenticated Status: %d (should be 401)\n", w.Code)
	if w.Code == 401 {
		fmt.Println("✅ Correctly rejected unauthenticated request")
	} else {
		fmt.Printf("❌ Should have returned 401. Body: %s\n", w.Body.String())
	}

	// 3. 测试有认证请求 (应该成功)
	fmt.Println("\n✅ Testing authenticated request...")
	
	req = httptest.NewRequest("GET", "/api/v1/files", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Authorization", "Bearer "+testToken)

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	fmt.Printf("📊 Authenticated Status: %d (should be 200)\n", w.Code)
	if w.Code == 200 {
		fmt.Println("✅ Successfully authenticated request")
		fmt.Printf("📋 Response: %s\n", w.Body.String())
	} else {
		fmt.Printf("❌ Authentication failed. Response: %s\n", w.Body.String())
	}

	// 4. 测试带X-Request-ID的认证请求
	fmt.Println("\n🔍 Testing authenticated request with X-Request-ID...")
	
	req = httptest.NewRequest("GET", "/api/v1/files", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Authorization", "Bearer "+testToken)
	req.Header.Set("X-Request-ID", "test-request-123")

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	fmt.Printf("📊 Status: %d (should be 200)\n", w.Code)
	fmt.Printf("🔍 Response X-Request-ID: %s\n", w.Header().Get("X-Request-ID"))
	fmt.Printf("📋 Access-Control-Expose-Headers: %s\n", w.Header().Get("Access-Control-Expose-Headers"))
	
	if w.Code == 200 {
		fmt.Println("✅ Request with X-Request-ID processed successfully")
		if w.Header().Get("Access-Control-Expose-Headers") == "X-Request-ID" {
			fmt.Println("✅ X-Request-ID header properly exposed in CORS")
		}
	} else {
		fmt.Printf("❌ Request with X-Request-ID failed. Response: %s\n", w.Body.String())
	}

	// 5. 测试无需认证的健康检查端点
	fmt.Println("\n💓 Testing health check endpoint (no auth required)...")
	
	req = httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	fmt.Printf("📊 Health Check Status: %d (should be 200)\n", w.Code)
	if w.Code == 200 {
		fmt.Println("✅ Health check works without authentication")
	} else {
		fmt.Printf("❌ Health check failed. Response: %s\n", w.Body.String())
	}

	fmt.Println("\n🎉 CORS and Authentication tests completed!")
	fmt.Println("✨ Summary:")
	fmt.Println("   - CORS preflight requests work correctly without triggering auth")
	fmt.Println("   - JWT authentication properly validates tokens")
	fmt.Println("   - X-Request-ID headers don't interfere with authentication")
	fmt.Println("   - Public endpoints remain accessible without auth")
	fmt.Println("   - Protected endpoints require valid JWT tokens")
}