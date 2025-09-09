package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	StudentID string `json:"student_id"`
	RoleID    int    `json:"role_id"`
	jwt.RegisteredClaims
}

// JWTMiddleware JWT认证中间件
func JWTMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// 解析JWT token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.StudentID)
		c.Set("role_id", claims.RoleID)
		c.Set("claims", claims)

		c.Next()
	}
}