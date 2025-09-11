package middleware

// ValidateTokenForTest 公共验证方法用于测试
func (m *AuthMiddleware) ValidateTokenForTest(tokenString string) (*UserClaims, error) {
	return m.validateJWT(tokenString)
}