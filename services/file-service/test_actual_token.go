package main

import (
	"fmt"

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
	// 用户服务实际生成的令牌
	actualToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbkBleGFtcGxlLmNvbSIsInJvbGUiOiJhZG1pbiIsInN0dWRlbnRfaWQiOiJhZG1pbiIsInJvbGVfaWQiOjEsInR5cGUiOiJhY2Nlc3MiLCJpc3MiOiJob3Jpem9uLWNsb3VkIiwic3ViIjoiXHUwMDAxIiwiZXhwIjoxNzU3NTE3Nzc3LCJuYmYiOjE3NTc1MTA1NzcsImlhdCI6MTc1NzUxMDU3N30.XuFsHjXEJgHtre-zu722vrbtR5h-RESabEo2oOdAJpA"
	
	// 使用的JWT密钥
	jwtSecret := "your-development-secret-key"
	
	fmt.Printf("开始验证用户服务实际生成的令牌\n")
	fmt.Printf("令牌字符串: %.50s...\n", actualToken)
	fmt.Printf("JWT密钥: %s\n", jwtSecret)
	
	// 用文件服务的验证方法验证实际令牌
	fileToken, err := jwt.ParseWithClaims(actualToken, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		fmt.Printf("令牌签名方法: %v\n", token.Method.Alg())
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		fmt.Printf("返回密钥进行验证，长度: %d\n", len([]byte(jwtSecret)))
		return []byte(jwtSecret), nil
	})
	
	if err != nil {
		fmt.Printf("JWT解析错误: %v\n", err)
		return
	}
	
	fmt.Printf("令牌解析完成，有效性: %v\n", fileToken.Valid)
	if claims, ok := fileToken.Claims.(*UserClaims); ok {
		fmt.Printf("声明类型正确，用户ID: %d, 用户名: %s\n", claims.UserID, claims.Username)
		if fileToken.Valid {
			fmt.Printf("令牌有效，验证成功!\n")
		} else {
			fmt.Printf("令牌无效\n")
		}
	} else {
		fmt.Printf("声明类型不匹配\n")
	}
}