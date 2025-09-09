package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthConfig JWT认证配置
type AuthConfig struct {
	JWTSecret     []byte
	JWTExpiration time.Duration
	SkipPaths     []string // 跳过认证的路径
}

// UserClaims JWT用户声明
type UserClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware JWT认证中间件
type AuthMiddleware struct {
	config *AuthConfig
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(config *AuthConfig) *AuthMiddleware {
	if config.JWTExpiration == 0 {
		config.JWTExpiration = 24 * time.Hour // 默认24小时
	}
	
	return &AuthMiddleware{
		config: config,
	}
}

// AuthRequired JWT认证中间件
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过认证
		if m.shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 获取Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.respondUnauthorized(c, "Missing authorization header")
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.respondUnauthorized(c, "Invalid authorization header format")
			return
		}

		tokenString := parts[1]
		
		// 解析和验证JWT
		claims, err := m.validateJWT(tokenString)
		if err != nil {
			m.respondUnauthorized(c, fmt.Sprintf("Invalid token: %v", err))
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("user_claims", claims)

		c.Next()
	}
}

// OptionalAuth 可选认证中间件（不强制要求认证）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := m.validateJWT(tokenString)
		if err == nil {
			// 设置用户信息
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)
			c.Set("user_claims", claims)
		}

		c.Next()
	}
}

// RoleRequired 角色权限中间件
func (m *AuthMiddleware) RoleRequired(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			m.respondForbidden(c, "User role not found")
			return
		}

		role, ok := userRole.(string)
		if !ok {
			m.respondForbidden(c, "Invalid user role")
			return
		}

		// 检查角色权限
		for _, allowedRole := range allowedRoles {
			if role == allowedRole || role == "admin" { // admin拥有所有权限
				c.Next()
				return
			}
		}

		m.respondForbidden(c, "Insufficient permissions")
	}
}

// AdminRequired 管理员权限中间件
func (m *AuthMiddleware) AdminRequired() gin.HandlerFunc {
	return m.RoleRequired("admin")
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) (uint, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New("user not authenticated")
	}

	id, ok := userID.(uint)
	if !ok {
		return 0, errors.New("invalid user ID type")
	}

	return id, nil
}

// GetUserIDFromParam 从URL参数获取用户ID（管理员操作）
func GetUserIDFromParam(c *gin.Context) (uint, error) {
	userIDStr := c.Param("user_id")
	if userIDStr == "" {
		return GetUserID(c) // 回退到当前用户
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID: %v", err)
	}

	return uint(userID), nil
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}

	name, ok := username.(string)
	if !ok {
		return ""
	}

	return name
}

// GetUserClaims 从上下文获取用户声明
func GetUserClaims(c *gin.Context) (*UserClaims, error) {
	claims, exists := c.Get("user_claims")
	if !exists {
		return nil, errors.New("user claims not found")
	}

	userClaims, ok := claims.(*UserClaims)
	if !ok {
		return nil, errors.New("invalid user claims type")
	}

	return userClaims, nil
}

// IsAuthenticated 检查用户是否已认证
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}

// IsAdmin 检查用户是否是管理员
func IsAdmin(c *gin.Context) bool {
	role, exists := c.Get("role")
	if !exists {
		return false
	}

	roleStr, ok := role.(string)
	return ok && roleStr == "admin"
}

// HasRole 检查用户是否拥有指定角色
func HasRole(c *gin.Context, requiredRole string) bool {
	role, exists := c.Get("role")
	if !exists {
		return false
	}

	roleStr, ok := role.(string)
	return ok && (roleStr == requiredRole || roleStr == "admin")
}

// GenerateJWT 生成JWT token
func (m *AuthMiddleware) GenerateJWT(userID uint, username, email, role string) (string, error) {
	now := time.Now()
	claims := &UserClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.JWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "file-service",
			Subject:   strconv.FormatUint(uint64(userID), 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.config.JWTSecret)
}

// RefreshToken 刷新JWT token
func (m *AuthMiddleware) RefreshToken(tokenString string) (string, error) {
	claims, err := m.validateJWT(tokenString)
	if err != nil {
		return "", err
	}

	// 检查是否在刷新窗口内（剩余时间少于1小时时允许刷新）
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", errors.New("token is not eligible for refresh yet")
	}

	// 生成新token
	return m.GenerateJWT(claims.UserID, claims.Username, claims.Email, claims.Role)
}

// validateJWT 验证JWT token
func (m *AuthMiddleware) validateJWT(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.config.JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// shouldSkipAuth 检查路径是否应该跳过认证
func (m *AuthMiddleware) shouldSkipAuth(path string) bool {
	for _, skipPath := range m.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// respondUnauthorized 返回401未授权响应
func (m *AuthMiddleware) respondUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":    http.StatusUnauthorized,
		"message": message,
		"error": gin.H{
			"type":        "UNAUTHORIZED",
			"description": message,
		},
	})
	c.Abort()
}

// respondForbidden 返回403禁止访问响应
func (m *AuthMiddleware) respondForbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{
		"code":    http.StatusForbidden,
		"message": message,
		"error": gin.H{
			"type":        "FORBIDDEN",
			"description": message,
		},
	})
	c.Abort()
}

// CreateUserContext 创建用户上下文
func CreateUserContext(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, "user_id", userID)
}

// GetUserFromContext 从上下文获取用户ID
func GetUserFromContext(ctx context.Context) (uint, bool) {
	userID, ok := ctx.Value("user_id").(uint)
	return userID, ok
}