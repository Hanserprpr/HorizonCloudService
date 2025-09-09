package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthMiddlewareTestSuite struct {
	suite.Suite
	middleware *AuthMiddleware
	engine     *gin.Engine
}

func (s *AuthMiddlewareTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	
	config := &AuthConfig{
		JWTSecret:     []byte("test-secret-key"),
		JWTExpiration: time.Hour,
		SkipPaths:     []string{"/public", "/health"},
	}
	
	s.middleware = NewAuthMiddleware(config)
	s.engine = gin.New()
}

func (s *AuthMiddlewareTestSuite) TestGenerateJWT() {
	token, err := s.middleware.GenerateJWT(1, "testuser", "test@example.com", "user")
	s.NoError(err)
	s.NotEmpty(token)

	// 验证生成的token
	claims, err := s.middleware.validateJWT(token)
	s.NoError(err)
	s.Equal(uint(1), claims.UserID)
	s.Equal("testuser", claims.Username)
	s.Equal("test@example.com", claims.Email)
	s.Equal("user", claims.Role)
}

func (s *AuthMiddlewareTestSuite) TestValidateJWT() {
	// 测试有效token
	validToken, _ := s.middleware.GenerateJWT(1, "testuser", "test@example.com", "user")
	claims, err := s.middleware.validateJWT(validToken)
	s.NoError(err)
	s.Equal(uint(1), claims.UserID)

	// 测试无效token
	_, err = s.middleware.validateJWT("invalid-token")
	s.Error(err)

	// 测试空token
	_, err = s.middleware.validateJWT("")
	s.Error(err)
}

func (s *AuthMiddlewareTestSuite) TestAuthRequiredMiddleware() {
	// 设置测试路由
	s.engine.Use(s.middleware.AuthRequired())
	s.engine.GET("/protected", func(c *gin.Context) {
		userID, _ := GetUserID(c)
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	// 测试无token访问
	req := httptest.NewRequest("GET", "/protected", nil)
	resp := httptest.NewRecorder()
	s.engine.ServeHTTP(resp, req)
	s.Equal(http.StatusUnauthorized, resp.Code)

	// 测试无效token
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp = httptest.NewRecorder()
	s.engine.ServeHTTP(resp, req)
	s.Equal(http.StatusUnauthorized, resp.Code)

	// 测试有效token
	validToken, _ := s.middleware.GenerateJWT(1, "testuser", "test@example.com", "user")
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	resp = httptest.NewRecorder()
	s.engine.ServeHTTP(resp, req)
	s.Equal(http.StatusOK, resp.Code)
}

func (s *AuthMiddlewareTestSuite) TestSkipPaths() {
	// 设置测试路由
	s.engine.Use(s.middleware.AuthRequired())
	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	s.engine.GET("/public/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"info": "public"})
	})

	// 测试跳过认证的路径
	req := httptest.NewRequest("GET", "/health", nil)
	resp := httptest.NewRecorder()
	s.engine.ServeHTTP(resp, req)
	s.Equal(http.StatusOK, resp.Code)

	req = httptest.NewRequest("GET", "/public/info", nil)
	resp = httptest.NewRecorder()
	s.engine.ServeHTTP(resp, req)
	s.Equal(http.StatusOK, resp.Code)
}

func (s *AuthMiddlewareTestSuite) TestRoleRequiredMiddleware() {
	// 创建新的引擎用于角色测试
	engine := gin.New()
	engine.Use(s.middleware.AuthRequired())
	engine.Use(s.middleware.RoleRequired("admin"))
	engine.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin only"})
	})

	// 测试普通用户访问管理员接口
	userToken, _ := s.middleware.GenerateJWT(1, "user", "user@example.com", "user")
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	s.Equal(http.StatusForbidden, resp.Code)

	// 测试管理员访问
	adminToken, _ := s.middleware.GenerateJWT(2, "admin", "admin@example.com", "admin")
	req = httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	s.Equal(http.StatusOK, resp.Code)
}

func (s *AuthMiddlewareTestSuite) TestAdminRequiredMiddleware() {
	// 创建新的引擎用于管理员测试
	engine := gin.New()
	engine.Use(s.middleware.AuthRequired())
	engine.Use(s.middleware.AdminRequired())
	engine.GET("/admin-only", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin only"})
	})

	// 测试普通用户
	userToken, _ := s.middleware.GenerateJWT(1, "user", "user@example.com", "user")
	req := httptest.NewRequest("GET", "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	s.Equal(http.StatusForbidden, resp.Code)

	// 测试管理员
	adminToken, _ := s.middleware.GenerateJWT(2, "admin", "admin@example.com", "admin")
	req = httptest.NewRequest("GET", "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	s.Equal(http.StatusOK, resp.Code)
}

func (s *AuthMiddlewareTestSuite) TestOptionalAuthMiddleware() {
	// 创建新的引擎用于可选认证测试
	engine := gin.New()
	engine.Use(s.middleware.OptionalAuth())
	engine.GET("/optional", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if exists {
			c.JSON(http.StatusOK, gin.H{"authenticated": true, "user_id": userID})
		} else {
			c.JSON(http.StatusOK, gin.H{"authenticated": false})
		}
	})

	// 测试无token访问
	req := httptest.NewRequest("GET", "/optional", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	s.Equal(http.StatusOK, resp.Code)

	// 测试有token访问
	validToken, _ := s.middleware.GenerateJWT(1, "testuser", "test@example.com", "user")
	req = httptest.NewRequest("GET", "/optional", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	resp = httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	s.Equal(http.StatusOK, resp.Code)
}

func (s *AuthMiddlewareTestSuite) TestRefreshToken() {
	// 创建一个即将过期的token (设置很短的过期时间)
	shortConfig := &AuthConfig{
		JWTSecret:     []byte("test-secret-key"),
		JWTExpiration: time.Microsecond, // 很短的过期时间
	}
	shortMiddleware := NewAuthMiddleware(shortConfig)
	
	token, err := shortMiddleware.GenerateJWT(1, "testuser", "test@example.com", "user")
	s.NoError(err)

	// 等待token过期
	time.Sleep(time.Millisecond)

	// 尝试刷新已过期的token（应该失败）
	_, err = s.middleware.RefreshToken(token)
	s.Error(err)

	// 测试有效token的刷新（在刷新窗口外）
	validToken, _ := s.middleware.GenerateJWT(1, "testuser", "test@example.com", "user")
	_, err = s.middleware.RefreshToken(validToken)
	s.Error(err) // 应该失败，因为token还没有到刷新窗口
}

func (s *AuthMiddlewareTestSuite) TestHelperFunctions() {
	// 创建测试上下文
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", uint(1))
	c.Set("username", "testuser")
	c.Set("role", "admin")

	// 测试GetUserID
	userID, err := GetUserID(c)
	s.NoError(err)
	s.Equal(uint(1), userID)

	// 测试GetUsername
	username := GetUsername(c)
	s.Equal("testuser", username)

	// 测试IsAuthenticated
	s.True(IsAuthenticated(c))

	// 测试IsAdmin
	s.True(IsAdmin(c))

	// 测试HasRole
	s.True(HasRole(c, "admin"))
	s.True(HasRole(c, "user")) // admin should have all roles
}

func TestAuthMiddleware(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareTestSuite))
}

// 基准测试
func BenchmarkJWTGeneration(b *testing.B) {
	config := &AuthConfig{
		JWTSecret:     []byte("test-secret-key"),
		JWTExpiration: time.Hour,
	}
	middleware := NewAuthMiddleware(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = middleware.GenerateJWT(1, "testuser", "test@example.com", "user")
	}
}

func BenchmarkJWTValidation(b *testing.B) {
	config := &AuthConfig{
		JWTSecret:     []byte("test-secret-key"),
		JWTExpiration: time.Hour,
	}
	middleware := NewAuthMiddleware(config)
	token, _ := middleware.GenerateJWT(1, "testuser", "test@example.com", "user")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = middleware.validateJWT(token)
	}
}

// 单元测试
func TestNewAuthMiddleware(t *testing.T) {
	config := &AuthConfig{
		JWTSecret: []byte("test-secret"),
	}

	middleware := NewAuthMiddleware(config)
	assert.NotNil(t, middleware)
	assert.Equal(t, 24*time.Hour, middleware.config.JWTExpiration) // 应该设置默认值
}

func TestCreateUserContext(t *testing.T) {
	ctx := CreateUserContext(nil, 123)
	userID, ok := GetUserFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, uint(123), userID)
}