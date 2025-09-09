package services

import (
	"context"
	"testing"
	"time"
	"user-service/internal/repository"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// UserServiceTestSuite 用户服务测试套件
type UserServiceTestSuite struct {
	suite.Suite
	userService UserService
	mockRepo    repository.UserRepository
	ctx         context.Context
}

// SetupSuite 设置测试套件
func (s *UserServiceTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.mockRepo = repository.NewMockUserRepository()
	s.userService = NewUserService(s.mockRepo, "test-jwt-secret")
}

// TestRegister 测试用户注册
func (s *UserServiceTestSuite) TestRegister() {
	tests := []struct {
		name        string
		req         *RegisterRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "成功注册",
			req: &RegisterRequest{
				StudentID: "test001",
				Password:  "password123",
				NickName:  "测试用户",
				Email:     "test@example.com",
				Phone:     "13800138000",
			},
			expectError: false,
		},
		{
			name: "学号为空",
			req: &RegisterRequest{
				StudentID: "",
				Password:  "password123",
				NickName:  "测试用户",
			},
			expectError: true,
			errorMsg:    "学号不能为空",
		},
		{
			name: "密码过短",
			req: &RegisterRequest{
				StudentID: "test002",
				Password:  "123",
				NickName:  "测试用户",
			},
			expectError: true,
			errorMsg:    "密码长度必须在6-50字符之间",
		},
		{
			name: "昵称为空",
			req: &RegisterRequest{
				StudentID: "test003",
				Password:  "password123",
				NickName:  "",
			},
			expectError: true,
			errorMsg:    "昵称不能为空",
		},
		{
			name: "邮箱格式错误",
			req: &RegisterRequest{
				StudentID: "test004",
				Password:  "password123",
				NickName:  "测试用户",
				Email:     "invalid-email",
			},
			expectError: true,
			errorMsg:    "邮箱格式不正确",
		},
	}
	
	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := s.userService.Register(s.ctx, tt.req)
			
			if tt.expectError {
				assert.Error(s.T(), err)
				assert.Contains(s.T(), err.Error(), tt.errorMsg)
				assert.Nil(s.T(), result)
			} else {
				assert.NoError(s.T(), err)
				assert.NotNil(s.T(), result)
				assert.NotEmpty(s.T(), result.AccessToken)
				assert.NotEmpty(s.T(), result.RefreshToken)
				assert.Equal(s.T(), tt.req.StudentID, result.User.StudentID)
				assert.Equal(s.T(), tt.req.NickName, result.User.NickName)
			}
		})
	}
}

// TestRegisterDuplicateStudentID 测试重复学号注册
func (s *UserServiceTestSuite) TestRegisterDuplicateStudentID() {
	// 先注册一个用户
	req1 := &RegisterRequest{
		StudentID: "duplicate001",
		Password:  "password123",
		NickName:  "用户1",
	}
	
	_, err := s.userService.Register(s.ctx, req1)
	assert.NoError(s.T(), err)
	
	// 尝试使用相同学号注册
	req2 := &RegisterRequest{
		StudentID: "duplicate001",
		Password:  "password456",
		NickName:  "用户2",
	}
	
	_, err = s.userService.Register(s.ctx, req2)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "学号已存在")
}

// TestLogin 测试用户登录
func (s *UserServiceTestSuite) TestLogin() {
	// 先注册一个用户
	registerReq := &RegisterRequest{
		StudentID: "login001",
		Password:  "password123",
		NickName:  "登录测试用户",
	}
	
	registerResult, err := s.userService.Register(s.ctx, registerReq)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), registerResult)
	
	tests := []struct {
		name        string
		req         *LoginRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "成功登录",
			req: &LoginRequest{
				StudentID: "login001",
				Password:  "password123",
			},
			expectError: false,
		},
		{
			name: "学号不存在",
			req: &LoginRequest{
				StudentID: "notexist",
				Password:  "password123",
			},
			expectError: true,
			errorMsg:    "用户名或密码错误",
		},
		{
			name: "密码错误",
			req: &LoginRequest{
				StudentID: "login001",
				Password:  "wrongpassword",
			},
			expectError: true,
			errorMsg:    "用户名或密码错误",
		},
		{
			name: "学号为空",
			req: &LoginRequest{
				StudentID: "",
				Password:  "password123",
			},
			expectError: true,
			errorMsg:    "学号不能为空",
		},
	}
	
	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := s.userService.Login(s.ctx, tt.req)
			
			if tt.expectError {
				assert.Error(s.T(), err)
				assert.Contains(s.T(), err.Error(), tt.errorMsg)
				assert.Nil(s.T(), result)
			} else {
				assert.NoError(s.T(), err)
				assert.NotNil(s.T(), result)
				assert.NotEmpty(s.T(), result.AccessToken)
				assert.NotEmpty(s.T(), result.RefreshToken)
				assert.Equal(s.T(), tt.req.StudentID, result.User.StudentID)
			}
		})
	}
}

// TestGetUserProfile 测试获取用户档案
func (s *UserServiceTestSuite) TestGetUserProfile() {
	// 先注册一个用户
	registerReq := &RegisterRequest{
		StudentID: "profile001",
		Password:  "password123",
		NickName:  "档案测试用户",
		Email:     "profile@example.com",
	}
	
	registerResult, err := s.userService.Register(s.ctx, registerReq)
	assert.NoError(s.T(), err)
	
	// 获取用户档案
	profile, err := s.userService.GetUserProfile(s.ctx, registerResult.User.ID)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), profile)
	assert.Equal(s.T(), registerReq.StudentID, profile.StudentID)
	assert.Equal(s.T(), registerReq.NickName, profile.NickName)
	assert.Equal(s.T(), registerReq.Email, profile.Email)
	
	// 获取不存在的用户档案
	_, err = s.userService.GetUserProfile(s.ctx, 99999)
	assert.Error(s.T(), err)
}

// TestUpdateUserProfile 测试更新用户档案
func (s *UserServiceTestSuite) TestUpdateUserProfile() {
	// 先注册一个用户
	registerReq := &RegisterRequest{
		StudentID: "update001",
		Password:  "password123",
		NickName:  "更新测试用户",
	}
	
	registerResult, err := s.userService.Register(s.ctx, registerReq)
	assert.NoError(s.T(), err)
	
	tests := []struct {
		name        string
		req         *UpdateProfileRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "成功更新",
			req: &UpdateProfileRequest{
				NickName: "新昵称",
				Email:    "new@example.com",
				Phone:    "13900139000",
				Avatar:   "http://example.com/avatar.jpg",
			},
			expectError: false,
		},
		{
			name: "昵称为空",
			req: &UpdateProfileRequest{
				NickName: "",
				Email:    "test@example.com",
			},
			expectError: true,
			errorMsg:    "昵称不能为空",
		},
		{
			name: "邮箱格式错误",
			req: &UpdateProfileRequest{
				NickName: "测试昵称",
				Email:    "invalid-email",
			},
			expectError: true,
			errorMsg:    "邮箱格式不正确",
		},
	}
	
	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.userService.UpdateUserProfile(s.ctx, registerResult.User.ID, tt.req)
			
			if tt.expectError {
				assert.Error(s.T(), err)
				assert.Contains(s.T(), err.Error(), tt.errorMsg)
			} else {
				assert.NoError(s.T(), err)
				
				// 验证更新是否成功
				profile, err := s.userService.GetUserProfile(s.ctx, registerResult.User.ID)
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), tt.req.NickName, profile.NickName)
				if tt.req.Email != "" {
					assert.Equal(s.T(), tt.req.Email, profile.Email)
				}
			}
		})
	}
}

// TestChangePassword 测试修改密码
func (s *UserServiceTestSuite) TestChangePassword() {
	// 先注册一个用户
	registerReq := &RegisterRequest{
		StudentID: "changepwd001",
		Password:  "oldpassword123",
		NickName:  "修改密码测试用户",
	}
	
	registerResult, err := s.userService.Register(s.ctx, registerReq)
	assert.NoError(s.T(), err)
	
	tests := []struct {
		name        string
		req         *ChangePasswordRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "成功修改密码",
			req: &ChangePasswordRequest{
				OldPassword: "oldpassword123",
				NewPassword: "newpassword123",
			},
			expectError: false,
		},
		{
			name: "原密码错误",
			req: &ChangePasswordRequest{
				OldPassword: "wrongpassword",
				NewPassword: "newpassword456",
			},
			expectError: true,
			errorMsg:    "原密码错误",
		},
		{
			name: "新密码过短",
			req: &ChangePasswordRequest{
				OldPassword: "oldpassword123",
				NewPassword: "123",
			},
			expectError: true,
			errorMsg:    "新密码长度必须在6-50字符之间",
		},
		{
			name: "新密码与原密码相同",
			req: &ChangePasswordRequest{
				OldPassword: "oldpassword123",
				NewPassword: "oldpassword123",
			},
			expectError: true,
			errorMsg:    "新密码不能与原密码相同",
		},
	}
	
	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.userService.ChangePassword(s.ctx, registerResult.User.ID, tt.req)
			
			if tt.expectError {
				assert.Error(s.T(), err)
				assert.Contains(s.T(), err.Error(), tt.errorMsg)
			} else {
				assert.NoError(s.T(), err)
				
				// 验证新密码是否生效
				loginReq := &LoginRequest{
					StudentID: registerReq.StudentID,
					Password:  tt.req.NewPassword,
				}
				loginResult, err := s.userService.Login(s.ctx, loginReq)
				assert.NoError(s.T(), err)
				assert.NotNil(s.T(), loginResult)
			}
		})
	}
}

// TestUserQuota 测试用户配额管理
func (s *UserServiceTestSuite) TestUserQuota() {
	// 先注册一个用户
	registerReq := &RegisterRequest{
		StudentID: "quota001",
		Password:  "password123",
		NickName:  "配额测试用户",
	}
	
	registerResult, err := s.userService.Register(s.ctx, registerReq)
	assert.NoError(s.T(), err)
	userID := registerResult.User.ID
	
	// 获取初始配额
	quota, err := s.userService.GetUserQuota(s.ctx, userID)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), quota)
	assert.Equal(s.T(), int64(5368709120), quota.StorageQuota) // 默认5GB
	assert.Equal(s.T(), int64(0), quota.StorageUsed)
	
	// 更新存储使用量
	err = s.userService.UpdateStorageUsed(s.ctx, userID, 1073741824) // 1GB
	assert.NoError(s.T(), err)
	
	quota, err = s.userService.GetUserQuota(s.ctx, userID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1073741824), quota.StorageUsed)
	
	// 测试超出配额限制
	err = s.userService.UpdateStorageUsed(s.ctx, userID, 6442450944) // 6GB (超出5GB限制)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "存储使用量超出配额限制")
	
	// 更新用户配额
	err = s.userService.UpdateUserQuota(s.ctx, userID, 10737418240) // 10GB
	assert.NoError(s.T(), err)
	
	quota, err = s.userService.GetUserQuota(s.ctx, userID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(10737418240), quota.StorageQuota)
}

// TestRefreshToken 测试刷新令牌
func (s *UserServiceTestSuite) TestRefreshToken() {
	// 先注册并登录用户
	registerReq := &RegisterRequest{
		StudentID: "refresh001",
		Password:  "password123",
		NickName:  "刷新令牌测试用户",
	}
	
	registerResult, err := s.userService.Register(s.ctx, registerReq)
	assert.NoError(s.T(), err)
	
	// 使用刷新令牌获取新的访问令牌
	refreshResult, err := s.userService.RefreshToken(s.ctx, registerResult.RefreshToken)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), refreshResult)
	assert.NotEmpty(s.T(), refreshResult.AccessToken)
	assert.NotEmpty(s.T(), refreshResult.RefreshToken)
	
	// 测试无效的刷新令牌
	_, err = s.userService.RefreshToken(s.ctx, "invalid-refresh-token")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "无效的刷新令牌")
}

// TestActivityLog 测试活动日志
func (s *UserServiceTestSuite) TestActivityLog() {
	// 先注册一个用户
	registerReq := &RegisterRequest{
		StudentID: "activity001",
		Password:  "password123",
		NickName:  "活动日志测试用户",
	}
	
	registerResult, err := s.userService.Register(s.ctx, registerReq)
	assert.NoError(s.T(), err)
	userID := registerResult.User.ID
	
	// 记录活动日志
	logReq := &LogActivityRequest{
		UserID:    userID,
		Action:    "test_action",
		Resource:  "test_resource",
		Detail:    "测试活动详情",
		IPAddress: "192.168.1.100",
		UserAgent: "Test-Agent/1.0",
	}
	
	err = s.userService.LogActivity(s.ctx, logReq)
	assert.NoError(s.T(), err)
	
	// 等待异步日志写入完成
	time.Sleep(100 * time.Millisecond)
	
	// 获取活动日志
	getLogsReq := &GetActivityLogsRequest{
		UserID:   userID,
		Page:     1,
		PageSize: 10,
	}
	
	logs, err := s.userService.GetUserActivityLogs(s.ctx, getLogsReq)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), logs)
	assert.True(s.T(), logs.Total >= 1) // 至少有注册时的日志
}

// 运行测试套件
func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}