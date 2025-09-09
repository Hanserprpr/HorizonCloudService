package services

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// 认证相关请求/响应结构

// RegisterRequest 注册请求
type RegisterRequest struct {
	StudentID string `json:"student_id" validate:"required,min=3,max=20"`
	Password  string `json:"password" validate:"required,min=6,max=50"`
	NickName  string `json:"nick_name" validate:"required,min=1,max=50"`
	Email     string `json:"email" validate:"omitempty,email,max=100"`
	Phone     string `json:"phone" validate:"omitempty,max=20"`
	IPAddress string `json:"-"`
	UserAgent string `json:"-"`
}

// Validate 验证注册请求
func (r *RegisterRequest) Validate() error {
	if strings.TrimSpace(r.StudentID) == "" {
		return errors.New("学号不能为空")
	}
	if len(r.StudentID) < 3 || len(r.StudentID) > 20 {
		return errors.New("学号长度必须在3-20字符之间")
	}
	if strings.TrimSpace(r.Password) == "" {
		return errors.New("密码不能为空")
	}
	if len(r.Password) < 6 || len(r.Password) > 50 {
		return errors.New("密码长度必须在6-50字符之间")
	}
	if strings.TrimSpace(r.NickName) == "" {
		return errors.New("昵称不能为空")
	}
	if len(r.NickName) > 50 {
		return errors.New("昵称长度不能超过50字符")
	}
	if r.Email != "" && !isValidEmail(r.Email) {
		return errors.New("邮箱格式不正确")
	}
	return nil
}

// LoginRequest 登录请求
type LoginRequest struct {
	StudentID string `json:"student_id" validate:"required"`
	Password  string `json:"password" validate:"required"`
	IPAddress string `json:"-"`
	UserAgent string `json:"-"`
}

// Validate 验证登录请求
func (r *LoginRequest) Validate() error {
	if strings.TrimSpace(r.StudentID) == "" {
		return errors.New("学号不能为空")
	}
	if strings.TrimSpace(r.Password) == "" {
		return errors.New("密码不能为空")
	}
	return nil
}

// AuthResponse 认证响应
type AuthResponse struct {
	User         *UserProfile `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

// UserProfile 用户档案
type UserProfile struct {
	ID           uint       `json:"id"`
	StudentID    string     `json:"student_id"`
	NickName     string     `json:"nick_name"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	Avatar       string     `json:"avatar"`
	RoleID       int        `json:"role_id"`
	Status       int        `json:"status"`
	StorageQuota int64      `json:"storage_quota"`
	StorageUsed  int64      `json:"storage_used"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	LoginCount   int        `json:"login_count"`
	CreatedAt    int64      `json:"created_at"`
	UpdatedAt    int64      `json:"updated_at"`
}

// UpdateProfileRequest 更新档案请求
type UpdateProfileRequest struct {
	NickName string `json:"nick_name" validate:"required,min=1,max=50"`
	Email    string `json:"email" validate:"omitempty,email,max=100"`
	Phone    string `json:"phone" validate:"omitempty,max=20"`
	Avatar   string `json:"avatar" validate:"omitempty,url,max=500"`
}

// Validate 验证更新档案请求
func (r *UpdateProfileRequest) Validate() error {
	if strings.TrimSpace(r.NickName) == "" {
		return errors.New("昵称不能为空")
	}
	if len(r.NickName) > 50 {
		return errors.New("昵称长度不能超过50字符")
	}
	if r.Email != "" && !isValidEmail(r.Email) {
		return errors.New("邮箱格式不正确")
	}
	if len(r.Phone) > 20 {
		return errors.New("电话号码长度不能超过20字符")
	}
	if len(r.Avatar) > 500 {
		return errors.New("头像URL长度不能超过500字符")
	}
	return nil
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6,max=50"`
}

// Validate 验证修改密码请求
func (r *ChangePasswordRequest) Validate() error {
	if strings.TrimSpace(r.OldPassword) == "" {
		return errors.New("原密码不能为空")
	}
	if strings.TrimSpace(r.NewPassword) == "" {
		return errors.New("新密码不能为空")
	}
	if len(r.NewPassword) < 6 || len(r.NewPassword) > 50 {
		return errors.New("新密码长度必须在6-50字符之间")
	}
	if r.OldPassword == r.NewPassword {
		return errors.New("新密码不能与原密码相同")
	}
	return nil
}

// 管理员相关请求/响应结构

// ListUsersRequest 用户列表请求
type ListUsersRequest struct {
	Page     int `json:"page" validate:"required,min=1"`
	PageSize int `json:"page_size" validate:"required,min=1,max=100"`
	Status   int `json:"status" validate:"omitempty,oneof=0 1 2"`
}

// Validate 验证用户列表请求
func (r *ListUsersRequest) Validate() error {
	if r.Page < 1 {
		return errors.New("页码必须大于0")
	}
	if r.PageSize < 1 || r.PageSize > 100 {
		return errors.New("每页大小必须在1-100之间")
	}
	if r.Status != 0 && r.Status != 1 && r.Status != 2 {
		return errors.New("状态值必须为0、1或2")
	}
	return nil
}

// ListUsersResponse 用户列表响应
type ListUsersResponse struct {
	Users    []*UserProfile `json:"users"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Pages    int64          `json:"pages"`
}

// SearchUsersRequest 搜索用户请求
type SearchUsersRequest struct {
	Keyword  string `json:"keyword" validate:"required,min=1,max=100"`
	Page     int    `json:"page" validate:"required,min=1"`
	PageSize int    `json:"page_size" validate:"required,min=1,max=100"`
}

// Validate 验证搜索用户请求
func (r *SearchUsersRequest) Validate() error {
	if strings.TrimSpace(r.Keyword) == "" {
		return errors.New("搜索关键词不能为空")
	}
	if len(r.Keyword) > 100 {
		return errors.New("搜索关键词长度不能超过100字符")
	}
	if r.Page < 1 {
		return errors.New("页码必须大于0")
	}
	if r.PageSize < 1 || r.PageSize > 100 {
		return errors.New("每页大小必须在1-100之间")
	}
	return nil
}

// SearchUsersResponse 搜索用户响应
type SearchUsersResponse struct {
	Users    []*UserProfile `json:"users"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Pages    int64          `json:"pages"`
	Keyword  string         `json:"keyword"`
}

// GetUsersByRoleRequest 根据角色获取用户请求
type GetUsersByRoleRequest struct {
	RoleID   int `json:"role_id" validate:"required,min=1"`
	Page     int `json:"page" validate:"required,min=1"`
	PageSize int `json:"page_size" validate:"required,min=1,max=100"`
}

// Validate 验证根据角色获取用户请求
func (r *GetUsersByRoleRequest) Validate() error {
	if r.RoleID < 1 {
		return errors.New("角色ID必须大于0")
	}
	if r.Page < 1 {
		return errors.New("页码必须大于0")
	}
	if r.PageSize < 1 || r.PageSize > 100 {
		return errors.New("每页大小必须在1-100之间")
	}
	return nil
}

// GetUsersByRoleResponse 根据角色获取用户响应
type GetUsersByRoleResponse struct {
	Users    []*UserProfile `json:"users"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Pages    int64          `json:"pages"`
	RoleID   int            `json:"role_id"`
}

// 活动日志相关结构

// LogActivityRequest 记录活动请求
type LogActivityRequest struct {
	UserID    uint   `json:"user_id" validate:"required"`
	Action    string `json:"action" validate:"required,max=50"`
	Resource  string `json:"resource" validate:"omitempty,max=100"`
	Detail    string `json:"detail" validate:"omitempty"`
	IPAddress string `json:"ip_address" validate:"omitempty,ip"`
	UserAgent string `json:"user_agent" validate:"omitempty"`
}

// Validate 验证记录活动请求
func (r *LogActivityRequest) Validate() error {
	if r.UserID == 0 {
		return errors.New("用户ID不能为空")
	}
	if strings.TrimSpace(r.Action) == "" {
		return errors.New("操作类型不能为空")
	}
	if len(r.Action) > 50 {
		return errors.New("操作类型长度不能超过50字符")
	}
	if len(r.Resource) > 100 {
		return errors.New("资源类型长度不能超过100字符")
	}
	return nil
}

// GetActivityLogsRequest 获取活动日志请求
type GetActivityLogsRequest struct {
	UserID   uint `json:"user_id" validate:"required"`
	Page     int  `json:"page" validate:"required,min=1"`
	PageSize int  `json:"page_size" validate:"required,min=1,max=100"`
}

// Validate 验证获取活动日志请求
func (r *GetActivityLogsRequest) Validate() error {
	if r.UserID == 0 {
		return errors.New("用户ID不能为空")
	}
	if r.Page < 1 {
		return errors.New("页码必须大于0")
	}
	if r.PageSize < 1 || r.PageSize > 100 {
		return errors.New("每页大小必须在1-100之间")
	}
	return nil
}

// ActivityLogItem 活动日志项
type ActivityLogItem struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
	Detail    string `json:"detail"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	CreatedAt int64  `json:"created_at"`
}

// GetActivityLogsResponse 获取活动日志响应
type GetActivityLogsResponse struct {
	Logs     []*ActivityLogItem `json:"logs"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Pages    int64              `json:"pages"`
}

// 辅助函数

// isValidEmail 验证邮箱格式
func isValidEmail(email string) bool {
	// 简单的邮箱正则验证
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}