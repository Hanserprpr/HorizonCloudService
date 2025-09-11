package main

import (
	"fmt"
	"log"

	"github.com/golang-jwt/jwt/v5"
)

// UserClaims JWT用户声明
type UserClaims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	StudentID string `json:"student_id"`
	RoleID    int    `json:"role_id"`
	Type      string `json:"type"`
	jwt.RegisteredClaims
}

func main() {
	// 用户服务生成的实际令牌
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbkBleGFtcGxlLmNvbSIsInJvbGUiOiJhZG1pbiIsInN0dWRlbnRfaWQiOiJhZG1pbiIsInJvbGVfaWQiOjEsInR5cGUiOiJhY2Nlc3MiLCJpc3MiOiJob3Jpem9uLWNsb3VkIiwic3ViIjoiXHUwMDAxIiwiZXhwIjoxNzU3NTE1MzExLCJuYmYiOjE3NTc1MDgxMTEsImlhdCI6MTc1NzUwODExMX0.al1ClSy7ExW014jXWouYnc_uKyLN16lriNEwPawCTMg"
	
	// 使用的密钥
	jwtSecret := "your-development-secret-key"
	
	log.Printf("开始验证JWT令牌")
	log.Printf("令牌字符串长度: %d", len(tokenString))
	log.Printf("JWT密钥长度: %d", len(jwtSecret))
	log.Printf("JWT密钥内容: %s", jwtSecret)
	
	// 尝试手动解码令牌头部查看算法
	tokenParts := []string{}
	if len(tokenString) > 0 {
		tokenParts = []string{tokenString}
	}
	if len(tokenParts) >= 1 {
		parts := []string{}
		for i, part := range tokenParts {
			if i < len(parts) {
				parts[i] = part
			} else {
				parts = append(parts, part)
			}
		}
		log.Printf("令牌部分数量: %d", len(parts))
	}
	
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		log.Printf("令牌签名方法: %v", token.Method.Alg())
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("签名方法不匹配，期望: HMAC, 实际: %T", token.Method)
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		log.Printf("返回密钥进行验证，长度: %d", len([]byte(jwtSecret)))
		return []byte(jwtSecret), nil
	})

	if err != nil {
		log.Printf("JWT解析错误: %v", err)
		return
	}

	log.Printf("令牌解析完成，有效性: %v", token.Valid)
	if claims, ok := token.Claims.(*UserClaims); ok {
		log.Printf("声明类型正确，用户ID: %d, 用户名: %s", claims.UserID, claims.Username)
		if token.Valid {
			log.Printf("令牌有效，验证成功")
		} else {
			log.Printf("令牌无效")
		}
	} else {
		log.Printf("声明类型不匹配")
	}
}