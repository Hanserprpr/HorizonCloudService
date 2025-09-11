package middleware

import (
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	jwtSecret string
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
	}
}

// JWTClaims JWT声明结构 - 与服务层保持一致
type JWTClaims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`  // 兼容文件服务
	Email     string `json:"email"`     // 兼容文件服务
	Role      string `json:"role"`      // 兼容文件服务（字符串类型）
	StudentID string `json:"student_id"`
	RoleID    int    `json:"role_id"`
	Type      string `json:"type"` // "access" 或 "refresh"
	jwt.RegisteredClaims
}

// Authenticate 认证中间件
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      http.StatusUnauthorized,
				"message":   "未授权",
				"data":      "缺少认证令牌",
				"success":   false,
				"timestamp": gin.H{},
			})
			c.Abort()
			return
		}
		
		// 检查Bearer格式
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      http.StatusUnauthorized,
				"message":   "未授权",
				"data":      "无效的令牌格式",
				"success":   false,
				"timestamp": gin.H{},
			})
			c.Abort()
			return
		}
		
		tokenString := tokenParts[1]
		
		// 解析并验证JWT令牌
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.jwtSecret), nil
		})
		
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      http.StatusUnauthorized,
				"message":   "未授权",
				"data":      "无效的令牌",
				"success":   false,
				"timestamp": gin.H{},
			})
			c.Abort()
			return
		}
		
		// 获取声明
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      http.StatusUnauthorized,
				"message":   "未授权",
				"data":      "无效的令牌声明",
				"success":   false,
				"timestamp": gin.H{},
			})
			c.Abort()
			return
		}
		
		// 检查令牌类型
		if claims.Type != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      http.StatusUnauthorized,
				"message":   "未授权",
				"data":      "令牌类型错误",
				"success":   false,
				"timestamp": gin.H{},
			})
			c.Abort()
			return
		}
		
		// 将用户信息保存到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("student_id", claims.StudentID)
		c.Set("role_id", claims.RoleID)
		
		c.Next()
	}
}

// RoleMiddleware 角色中间件
type RoleMiddleware struct {
	allowedRoles []int
}

// NewRoleMiddleware 创建角色中间件
func NewRoleMiddleware(allowedRoles []int) *RoleMiddleware {
	return &RoleMiddleware{
		allowedRoles: allowedRoles,
	}
}

// CheckRole 检查角色权限
func (m *RoleMiddleware) CheckRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取角色ID
		roleID, exists := c.Get("role_id")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code":      http.StatusForbidden,
				"message":   "权限不足",
				"data":      "无法获取用户角色",
				"success":   false,
				"timestamp": gin.H{},
			})
			c.Abort()
			return
		}
		
		// 检查角色是否在允许列表中
		userRoleID := roleID.(int)
		allowed := false
		for _, allowedRole := range m.allowedRoles {
			if userRoleID == allowedRole {
				allowed = true
				break
			}
		}
		
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"code":      http.StatusForbidden,
				"message":   "权限不足",
				"data":      "您没有执行此操作的权限",
				"success":   false,
				"timestamp": gin.H{},
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}