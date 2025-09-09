package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// SimpleAuthMiddleware 简化的认证中间件
type SimpleAuthMiddleware struct {
	*AuthMiddleware
}

// NewJWTAuthMiddleware 创建简化的JWT认证中间件
func NewJWTAuthMiddleware() *SimpleAuthMiddleware {
	config := &AuthConfig{
		JWTSecret:     []byte("your-development-secret-key"),
		JWTExpiration: 24 * time.Hour,
		SkipPaths: []string{
			"/health",
			"/health/ready",
			"/uploads",
			"/thumbnails",
			"/favicon.ico",
		},
	}

	authMiddleware := NewAuthMiddleware(config)
	
	return &SimpleAuthMiddleware{
		AuthMiddleware: authMiddleware,
	}
}

// Authenticate 认证中间件（开发环境下宽松验证）
func (m *SimpleAuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在开发环境下，如果没有token则设置默认用户信息
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 设置默认测试用户
			c.Set("user_id", uint(1))
			c.Set("username", "admin")
			c.Set("email", "admin@example.com")
			c.Set("role", "admin")
			c.Next()
			return
		}

		// 如果有token，则进行正常验证
		m.AuthMiddleware.AuthRequired()(c)
	}
}

// RequireAuth 要求认证
func (m *SimpleAuthMiddleware) RequireAuth() gin.HandlerFunc {
	return m.Authenticate()
}

// RequireAdmin 要求管理员权限
func (m *SimpleAuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.Authenticate()(c)
		if c.IsAborted() {
			return
		}

		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(403, gin.H{
				"code":    403,
				"message": "Admin privileges required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}