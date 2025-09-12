package main

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// JWT Claims structure from user service 
type UserJWTClaims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	StudentID string `json:"student_id"`
	RoleID    int    `json:"role_id"`
	Type      string `json:"type"`
	jwt.RegisteredClaims
}

// JWT Claims structure expected by file service
type FileJWTClaims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	StudentID string `json:"student_id"`
	RoleID    int    `json:"role_id"`
	jwt.RegisteredClaims
}

func main() {
	fmt.Println("🔍 JWT Secret and Claims Debug Test")
	fmt.Println("=====================================")

	// Test different JWT secrets
	secrets := []string{
		"your-development-secret-key",  // Current standardized secret
		"your-jwt-secret-key",          // Old user service default
		"your-secret-key",              // Generic default
	}

	// Sample JWT token from network test (from actual user service)
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyLCJ1c2VybmFtZSI6InRlc3QwMDEiLCJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJyb2xlIjoiYWRtaW4iLCJzdHVkZW50X2lkIjoidGVzdDAwMSIsInJvbGVfaWQiOjEsInR5cGUiOiJhY2Nlc3MiLCJpc3MiOiJob3Jpem9uLWNsb3VkIiwic3ViIjoiXHUwMDAyIiwiZXhwIjoxNzU3NTIyNzA4LCJuYmYiOjE3NTc1MTU1MDgsImlhdCI6MTc1NzUxNTUwOH0.OK2tcXBekioKaFC10btF6yMCZmQUuhJMuhC-DovbKd4"

	fmt.Printf("📝 Testing JWT Token: %.60s...\n", tokenString)
	fmt.Println()

	// Test each secret
	for i, secret := range secrets {
		fmt.Printf("🔐 Test %d: Testing with secret '%s'\n", i+1, secret)
		fmt.Printf("   Secret length: %d bytes\n", len(secret))
		
		// Parse with user service claims
		userClaims, userErr := parseWithSecret(tokenString, secret, &UserJWTClaims{})
		if userErr == nil {
			fmt.Printf("   ✅ SUCCESS with User Claims: UserID=%d, Username=%s, Role=%s, Type=%s\n", 
				userClaims.(*UserJWTClaims).UserID,
				userClaims.(*UserJWTClaims).Username,
				userClaims.(*UserJWTClaims).Role,
				userClaims.(*UserJWTClaims).Type)
		} else {
			fmt.Printf("   ❌ FAILED with User Claims: %v\n", userErr)
		}

		// Parse with file service claims
		fileClaims, fileErr := parseWithSecret(tokenString, secret, &FileJWTClaims{})
		if fileErr == nil {
			fmt.Printf("   ✅ SUCCESS with File Claims: UserID=%d, Username=%s, Role=%s\n", 
				fileClaims.(*FileJWTClaims).UserID,
				fileClaims.(*FileJWTClaims).Username,
				fileClaims.(*FileJWTClaims).Role)
		} else {
			fmt.Printf("   ❌ FAILED with File Claims: %v\n", fileErr)
		}
		
		fmt.Println()
	}

	// Environment variable check
	fmt.Println("📋 Environment Variable Check:")
	fmt.Printf("   JWT_SECRET = '%s'\n", os.Getenv("JWT_SECRET"))
	fmt.Printf("   JWT_EXPIRATION_HOURS = '%s'\n", os.Getenv("JWT_EXPIRATION_HOURS"))
	fmt.Printf("   SERVER_PORT = '%s'\n", os.Getenv("SERVER_PORT"))
	fmt.Println()

	// Generate test token with standardized secret
	fmt.Println("🔧 Generating test token with standardized secret:")
	standardSecret := "your-development-secret-key"
	testToken, err := generateTestToken(standardSecret)
	if err != nil {
		log.Printf("Failed to generate test token: %v", err)
		return
	}
	
	fmt.Printf("   Generated token: %.60s...\n", testToken)
	
	// Verify the generated token
	claims, err := parseWithSecret(testToken, standardSecret, &FileJWTClaims{})
	if err == nil {
		fmt.Printf("   ✅ Self-verification successful: UserID=%d, Username=%s\n", 
			claims.(*FileJWTClaims).UserID,
			claims.(*FileJWTClaims).Username)
	} else {
		fmt.Printf("   ❌ Self-verification failed: %v\n", err)
	}
}

func parseWithSecret(tokenString, secret string, claims jwt.Claims) (jwt.Claims, error) {
	secretBytes := []byte(secret)
	
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretBytes, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return token.Claims, nil
}

func generateTestToken(secret string) (string, error) {
	claims := &FileJWTClaims{
		UserID:    2,
		Username:  "test001",
		Email:     "test@example.com", 
		Role:      "admin",
		StudentID: "test001",
		RoleID:    1,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  "test-generator",
			Subject: "2",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}