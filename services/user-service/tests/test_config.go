package tests

import (
	"user-service/internal/repository"
	"user-service/internal/services"
)

// TestConfig 测试配置
type TestConfig struct {
	JWTSecret string
}

// DefaultTestConfig 默认测试配置
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		JWTSecret: "test-jwt-secret-for-unit-testing-only",
	}
}

// NewTestServices 创建测试服务
func NewTestServices(config *TestConfig) services.UserService {
	mockRepo := repository.NewMockUserRepository()
	return services.NewUserService(mockRepo, config.JWTSecret)
}

// TestUserData 测试用户数据
type TestUserData struct {
	StudentID string
	Password  string
	NickName  string
	Email     string
	Phone     string
}

// DefaultTestUsers 默认测试用户
func DefaultTestUsers() []TestUserData {
	return []TestUserData{
		{
			StudentID: "admin001",
			Password:  "admin123456",
			NickName:  "测试管理员",
			Email:     "admin@test.com",
			Phone:     "13800138000",
		},
		{
			StudentID: "user001",
			Password:  "user123456",
			NickName:  "测试用户1",
			Email:     "user1@test.com",
			Phone:     "13800138001",
		},
		{
			StudentID: "user002",
			Password:  "user123456",
			NickName:  "测试用户2",
			Email:     "user2@test.com",
			Phone:     "13800138002",
		},
	}
}