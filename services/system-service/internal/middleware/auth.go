package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// SimpleAuthMiddleware 简化认证中间件
type SimpleAuthMiddleware struct {
	jwtSecret []byte
}

// NewSimpleAuthMiddleware 创建简化认证中间件
func NewSimpleAuthMiddleware() *SimpleAuthMiddleware {
	// 从环境变量获取JWT密钥，如果没有则使用默认值（仅用于开发）
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key" // 开发环境默认密钥
	}

	return &SimpleAuthMiddleware{
		jwtSecret: []byte(secret),
	}
}

// RequireAuth 需要认证的中间件
func (m *SimpleAuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开发模式：如果设置了跳过认证，则直接通过
		if os.Getenv("SKIP_AUTH") == "true" {
			// 在开发模式下，设置一个模拟的用户ID
			c.Set("user_id", uint(1))
			c.Set("is_admin", true)
			c.Next()
			return
		}

		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authorization header required",
				"message": "Please provide a valid authorization token",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authorization format",
				"message": "Authorization header must start with 'Bearer '",
			})
			c.Abort()
			return
		}

		// 提取token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Token required",
				"message": "Please provide a valid authorization token",
			})
			c.Abort()
			return
		}

		// 解析和验证JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 确保使用HMAC签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"message": "The provided token is invalid or expired",
			})
			c.Abort()
			return
		}

		// 检查token是否有效
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"message": "The provided token is invalid",
			})
			c.Abort()
			return
		}

		// 提取claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// 设置用户信息到上下文
			if userID, exists := claims["user_id"]; exists {
				c.Set("user_id", userID)
			}
			if isAdmin, exists := claims["is_admin"]; exists {
				c.Set("is_admin", isAdmin)
			}
			if roleID, exists := claims["role_id"]; exists {
				c.Set("role_id", roleID)
			}
		}

		c.Next()
	}
}

// RequireAdmin 需要管理员权限的中间件
func (m *SimpleAuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 首先检查是否已认证
		_, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication required",
				"message": "Please login to access this resource",
			})
			c.Abort()
			return
		}

		// 检查是否是管理员
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Admin access required",
				"message": "This resource requires administrator privileges",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) uint {
	if userID, exists := c.Get("user_id"); exists {
		switch id := userID.(type) {
		case uint:
			return id
		case float64:
			return uint(id)
		case int:
			return uint(id)
		}
	}
	return 0
}

// IsAdmin 检查当前用户是否是管理员
func IsAdmin(c *gin.Context) bool {
	if isAdmin, exists := c.Get("is_admin"); exists {
		if admin, ok := isAdmin.(bool); ok {
			return admin
		}
	}
	return false
}