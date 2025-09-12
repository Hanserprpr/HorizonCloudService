package services

import (
	"context"
	"errors"
	"fmt"
	"user-service/internal/models"
	"user-service/internal/repository"
	
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// UserService 用户服务接口
type UserService interface {
	// 认证相关
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, userID uint) error
	
	// 用户管理
	GetUserProfile(ctx context.Context, userID uint) (*UserProfile, error)
	UpdateUserProfile(ctx context.Context, userID uint, req *UpdateProfileRequest) error
	ChangePassword(ctx context.Context, userID uint, req *ChangePasswordRequest) error
	
	// 用户查询（管理员功能）
	ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error)
	SearchUsers(ctx context.Context, req *SearchUsersRequest) (*SearchUsersResponse, error)
	GetUsersByRole(ctx context.Context, req *GetUsersByRoleRequest) (*GetUsersByRoleResponse, error)
	
	// 用户状态管理（管理员功能）
	UpdateUserStatus(ctx context.Context, userID uint, status int) error
	DeleteUser(ctx context.Context, userID uint) error
	
	// 存储配额管理
	GetUserQuota(ctx context.Context, userID uint) (*models.UserQuota, error)
	UpdateUserQuota(ctx context.Context, userID uint, quota int64) error
	UpdateStorageUsed(ctx context.Context, userID uint, used int64) error
	
	// 活动日志
	LogActivity(ctx context.Context, req *LogActivityRequest) error
	GetUserActivityLogs(ctx context.Context, req *GetActivityLogsRequest) (*GetActivityLogsResponse, error)
}

// userService 用户服务实现
type userService struct {
	userRepo repository.UserRepository
	jwtSecret string
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, jwtSecret string) UserService {
	return &userService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Register 用户注册
func (s *userService) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	// 检查学号是否已存在
	if _, err := s.userRepo.GetByStudentID(ctx, req.StudentID); err == nil {
		return nil, errors.New("学号已存在")
	}
	
	// 检查邮箱是否已存在
	if req.Email != "" {
		if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
			return nil, errors.New("邮箱已存在")
		}
	}
	
	// 加密密码
	hashedPassword := s.hashPassword(req.Password)
	
	// 创建用户
	user := &models.User{
		StudentID:    req.StudentID,
		Password:     hashedPassword,
		RoleID:       1, // 默认普通用户
		NickName:     req.NickName,
		Email:        req.Email,
		Phone:        req.Phone,
		Status:       1, // 正常状态
		StorageQuota: 5368709120, // 默认5GB
		StorageUsed:  0,
	}
	
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}
	
	// 记录注册活动
	s.logUserActivity(ctx, user.ID, "register", "user", "用户注册", req.IPAddress, req.UserAgent)
	
	// 生成JWT令牌
	tokens, err := s.generateTokens(user.ID, user.StudentID, user.Email, user.StudentID, user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}
	
	return &AuthResponse{
		User: &UserProfile{
			ID:           user.ID,
			StudentID:    user.StudentID,
			NickName:     user.NickName,
			Email:        user.Email,
			Phone:        user.Phone,
			Avatar:       user.Avatar,
			RoleID:       user.RoleID,
			Status:       user.Status,
			StorageQuota: user.StorageQuota,
			StorageUsed:  user.StorageUsed,
			CreatedAt:    user.CreateTime(),
		},
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	// 获取用户
	user, err := s.userRepo.GetByStudentID(ctx, req.StudentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	
	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("用户已被禁用")
	}
	
	// 验证密码
	if !s.verifyPassword(req.Password, user.Password) {
		return nil, errors.New("用户名或密码错误")
	}
	
	// 更新登录信息
	if err := s.userRepo.UpdateLoginInfo(ctx, user.ID); err != nil {
		// 记录错误但不影响登录流程
		fmt.Printf("更新登录信息失败: %v\n", err)
	}
	
	// 记录登录活动
	s.logUserActivity(ctx, user.ID, "login", "user", "用户登录", req.IPAddress, req.UserAgent)
	
	// 生成JWT令牌
	tokens, err := s.generateTokens(user.ID, user.StudentID, user.Email, user.StudentID, user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}
	
	return &AuthResponse{
		User: &UserProfile{
			ID:           user.ID,
			StudentID:    user.StudentID,
			NickName:     user.NickName,
			Email:        user.Email,
			Phone:        user.Phone,
			Avatar:       user.Avatar,
			RoleID:       user.RoleID,
			Status:       user.Status,
			StorageQuota: user.StorageQuota,
			StorageUsed:  user.StorageUsed,
			CreatedAt:    user.CreateTime(),
		},
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

// RefreshToken 刷新访问令牌
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// 解析并验证刷新令牌
	token, err := jwt.ParseWithClaims(refreshToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	
	if err != nil || !token.Valid {
		return nil, errors.New("无效的刷新令牌")
	}
	
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("无效的令牌声明")
	}
	
	// 检查令牌类型
	if claims.Type != "refresh" {
		return nil, errors.New("令牌类型错误")
	}
	
	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	
	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("用户已被禁用")
	}
	
	// 生成新的访问令牌
	tokens, err := s.generateTokens(user.ID, user.StudentID, user.Email, user.StudentID, user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}
	
	return &AuthResponse{
		User: &UserProfile{
			ID:           user.ID,
			StudentID:    user.StudentID,
			NickName:     user.NickName,
			Email:        user.Email,
			Phone:        user.Phone,
			Avatar:       user.Avatar,
			RoleID:       user.RoleID,
			Status:       user.Status,
			StorageQuota: user.StorageQuota,
			StorageUsed:  user.StorageUsed,
			CreatedAt:    user.CreateTime(),
		},
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

// Logout 用户登出
func (s *userService) Logout(ctx context.Context, userID uint) error {
	// 记录登出活动
	s.logUserActivity(ctx, userID, "logout", "user", "用户登出", "", "")
	
	// 实际应用中可能需要将token加入黑名单
	// 这里暂时只记录活动日志
	return nil
}

// GetUserProfile 获取用户档案
func (s *userService) GetUserProfile(ctx context.Context, userID uint) (*UserProfile, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	
	return &UserProfile{
		ID:           user.ID,
		StudentID:    user.StudentID,
		NickName:     user.NickName,
		Email:        user.Email,
		Phone:        user.Phone,
		Avatar:       user.Avatar,
		RoleID:       user.RoleID,
		Status:       user.Status,
		StorageQuota: user.StorageQuota,
		StorageUsed:  user.StorageUsed,
		LastLoginAt:  user.LastLoginAt,
		LoginCount:   user.LoginCount,
		CreatedAt:    user.CreateTime(),
		UpdatedAt:    user.UpdateTime(),
	}, nil
}

// UpdateUserProfile 更新用户档案
func (s *userService) UpdateUserProfile(ctx context.Context, userID uint, req *UpdateProfileRequest) error {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		return err
	}
	
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}
	
	// 更新字段
	user.NickName = req.NickName
	user.Email = req.Email
	user.Phone = req.Phone
	user.Avatar = req.Avatar
	
	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("更新用户信息失败: %w", err)
	}
	
	// 记录活动
	s.logUserActivity(ctx, userID, "update_profile", "user", "更新个人信息", "", "")
	
	return nil
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(ctx context.Context, userID uint, req *ChangePasswordRequest) error {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		return err
	}
	
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}
	
	// 验证原密码
	if !s.verifyPassword(req.OldPassword, user.Password) {
		return errors.New("原密码错误")
	}
	
	// 加密新密码
	user.Password = s.hashPassword(req.NewPassword)
	
	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}
	
	// 记录活动
	s.logUserActivity(ctx, userID, "change_password", "user", "修改密码", "", "")
	
	return nil
}

// 其他方法的实现将在下一个文件中继续...