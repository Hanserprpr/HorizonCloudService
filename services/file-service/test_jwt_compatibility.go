package main

import (
	"fmt"
	"log"
	"time"

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
	// 用户服务使用的JWT密钥
	userSecret := "your-development-secret-key"
	
	// 文件服务使用的JWT密钥
	fileSecret := "your-development-secret-key"
	
	// 检查密钥是否相同
	if userSecret != fileSecret {
		log.Fatal("用户服务和文件服务使用不同的JWT密钥!")
	}
	
	fmt.Printf("用户服务和文件服务使用相同的JWT密钥: %s\n", userSecret)
	
	// 创建一个测试令牌并用两个服务都验证
	claims := &UserClaims{
		UserID:    1,
		Username:  "admin",
		Email:     "admin@example.com",
		Role:      "admin",
		StudentID: "admin",
		RoleID:    1,
		Type:      "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "horizon-cloud",
			Subject:   "1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	
	// 用用户服务的签名方法创建令牌
	userToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	userTokenString, err := userToken.SignedString([]byte(userSecret))
	if err != nil {
		log.Fatalf("用户服务生成令牌失败: %v", err)
	}
	
	fmt.Printf("用户服务生成的令牌: %s\n", userTokenString)
	
	// 用文件服务的验证方法验证
	fileToken, err := jwt.ParseWithClaims(userTokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(fileSecret), nil
	})
	
	if err != nil {
		log.Fatalf("文件服务验证令牌失败: %v", err)
	}
	
	if fileToken.Valid {
		fmt.Println("文件服务验证成功!")
		if claims, ok := fileToken.Claims.(*UserClaims); ok {
			fmt.Printf("验证结果 - 用户ID: %d, 用户名: %s\n", claims.UserID, claims.Username)
		}
	} else {
		fmt.Println("文件服务验证失败 - 令牌无效")
	}
}