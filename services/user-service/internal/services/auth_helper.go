package services

import (
	"context"
	"time"
	"user-service/internal/models"
	
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWTClaims JWT声明结构 - 兼容文件服务格式
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

// TokenPair 令牌对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// generateTokens 生成访问令牌和刷新令牌
func (s *userService) generateTokens(userID uint, username, email, studentID string, roleID int) (*TokenPair, error) {
	now := time.Now()
	
	// 将角色ID转换为角色字符串
	roleStr := "user" // 默认角色
	if roleID == 1 {
		roleStr = "admin"
	}
	
	// 访问令牌声明 (有效期2小时)
	accessClaims := &JWTClaims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		Role:      roleStr,
		StudentID: studentID,
		RoleID:    roleID,
		Type:      "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   string(rune(userID)),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "horizon-cloud",
		},
	}
	
	// 刷新令牌声明 (有效期7天)
	refreshClaims := &JWTClaims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		Role:      roleStr,
		StudentID: studentID,
		RoleID:    roleID,
		Type:      "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   string(rune(userID)),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "horizon-cloud",
		},
	}
	
	// 生成访问令牌
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}
	
	// 生成刷新令牌
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}
	
	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    2 * 60 * 60, // 2小时，以秒为单位
	}, nil
}

// hashPassword 加密密码
func (s *userService) hashPassword(password string) string {
	hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedBytes)
}

// verifyPassword 验证密码
func (s *userService) verifyPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// logUserActivity 记录用户活动的辅助方法
func (s *userService) logUserActivity(ctx context.Context, userID uint, action, resource, detail, ipAddress, userAgent string) {
	log := &models.ActivityLog{
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Detail:    detail,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
	
	// 异步记录日志，不影响主流程
	go func() {
		if err := s.userRepo.CreateActivityLog(context.Background(), log); err != nil {
			// 记录错误，但不中断程序执行
			// 在实际应用中，这里应该使用结构化日志记录错误
			_ = err
		}
	}()
}