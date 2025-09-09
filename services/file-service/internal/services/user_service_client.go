package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// UserServiceClient 用户服务客户端接口
type UserServiceClient interface {
	GetUser(ctx context.Context, userID uint) (*UserInfo, error)
	GetUserByEmail(ctx context.Context, email string) (*UserInfo, error)
	UpdateUserQuota(ctx context.Context, userID uint, storageQuota, fileCountQuota int64) error
	ValidateUserPermission(ctx context.Context, userID uint, resource, action string) (bool, error)
	GetUserRole(ctx context.Context, userID uint) (*UserRole, error)
	UpdateUserStorageUsage(ctx context.Context, userID uint, usedStorage int64) error
	NotifyQuotaExceeded(ctx context.Context, userID uint, quotaType string) error
}

// UserInfo 用户信息
type UserInfo struct {
	ID               uint      `json:"id"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	Role             string    `json:"role"`
	Status           int       `json:"status"`
	StorageQuota     int64     `json:"storage_quota"`
	FileCountQuota   int64     `json:"file_count_quota"`
	StorageUsed      int64     `json:"storage_used"`
	FileCount        int64     `json:"file_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	LastLoginAt      *time.Time `json:"last_login_at"`
	Settings         map[string]interface{} `json:"settings"`
}

// UserRole 用户角色
type UserRole struct {
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

// userServiceClient 用户服务客户端实现
type userServiceClient struct {
	client  *resty.Client
	baseURL string
	apiKey  string
}

// UserServiceConfig 用户服务配置
type UserServiceConfig struct {
	BaseURL        string        `json:"base_url"`
	APIKey         string        `json:"api_key"`
	Timeout        time.Duration `json:"timeout"`
	RetryCount     int           `json:"retry_count"`
	RetryInterval  time.Duration `json:"retry_interval"`
	EnableCircuitBreaker bool    `json:"enable_circuit_breaker"`
}

// NewUserServiceClient 创建用户服务客户端
func NewUserServiceClient(config *UserServiceConfig) UserServiceClient {
	client := resty.New()
	
	// 设置基本配置
	client.SetBaseURL(config.BaseURL)
	client.SetTimeout(config.Timeout)
	client.SetRetryCount(config.RetryCount)
	client.SetRetryWaitTime(config.RetryInterval)
	
	// 设置认证头
	if config.APIKey != "" {
		client.SetAuthToken(config.APIKey)
	}
	
	// 设置通用请求头
	client.SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
		"User-Agent":   "file-service/1.0.0",
	})
	
	// 设置重试条件
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		return r.StatusCode() >= 500 || err != nil
	})

	return &userServiceClient{
		client:  client,
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
	}
}

// GetUser 获取用户信息
func (c *userServiceClient) GetUser(ctx context.Context, userID uint) (*UserInfo, error) {
	var response struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Data    UserInfo `json:"data"`
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetPathParam("user_id", fmt.Sprintf("%d", userID)).
		SetResult(&response).
		Get("/api/v1/users/{user_id}")

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("user service returned status %d: %s", resp.StatusCode(), resp.String())
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("user service error: %s", response.Message)
	}

	return &response.Data, nil
}

// GetUserByEmail 根据邮箱获取用户信息
func (c *userServiceClient) GetUserByEmail(ctx context.Context, email string) (*UserInfo, error) {
	var response struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Data    UserInfo `json:"data"`
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("email", email).
		SetResult(&response).
		Get("/api/v1/users/search")

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("user service returned status %d: %s", resp.StatusCode(), resp.String())
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("user service error: %s", response.Message)
	}

	return &response.Data, nil
}

// UpdateUserQuota 更新用户配额
func (c *userServiceClient) UpdateUserQuota(ctx context.Context, userID uint, storageQuota, fileCountQuota int64) error {
	requestBody := map[string]interface{}{
		"storage_quota":     storageQuota,
		"file_count_quota":  fileCountQuota,
	}

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetPathParam("user_id", fmt.Sprintf("%d", userID)).
		SetBody(requestBody).
		SetResult(&response).
		Put("/api/v1/users/{user_id}/quota")

	if err != nil {
		return fmt.Errorf("failed to update user quota: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("user service returned status %d: %s", resp.StatusCode(), resp.String())
	}

	if response.Code != 200 {
		return fmt.Errorf("user service error: %s", response.Message)
	}

	return nil
}

// ValidateUserPermission 验证用户权限
func (c *userServiceClient) ValidateUserPermission(ctx context.Context, userID uint, resource, action string) (bool, error) {
	requestBody := map[string]interface{}{
		"resource": resource,
		"action":   action,
	}

	var response struct {
		Code    int `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Allowed bool `json:"allowed"`
		} `json:"data"`
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetPathParam("user_id", fmt.Sprintf("%d", userID)).
		SetBody(requestBody).
		SetResult(&response).
		Post("/api/v1/users/{user_id}/permissions/check")

	if err != nil {
		return false, fmt.Errorf("failed to validate user permission: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return false, fmt.Errorf("permission service returned status %d: %s", resp.StatusCode(), resp.String())
	}

	if response.Code != 200 {
		return false, fmt.Errorf("permission service error: %s", response.Message)
	}

	return response.Data.Allowed, nil
}

// GetUserRole 获取用户角色
func (c *userServiceClient) GetUserRole(ctx context.Context, userID uint) (*UserRole, error) {
	var response struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Data    UserRole `json:"data"`
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetPathParam("user_id", fmt.Sprintf("%d", userID)).
		SetResult(&response).
		Get("/api/v1/users/{user_id}/role")

	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("user service returned status %d: %s", resp.StatusCode(), resp.String())
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("user service error: %s", response.Message)
	}

	return &response.Data, nil
}

// UpdateUserStorageUsage 更新用户存储使用量
func (c *userServiceClient) UpdateUserStorageUsage(ctx context.Context, userID uint, usedStorage int64) error {
	requestBody := map[string]interface{}{
		"storage_used": usedStorage,
	}

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetPathParam("user_id", fmt.Sprintf("%d", userID)).
		SetBody(requestBody).
		SetResult(&response).
		Put("/api/v1/users/{user_id}/storage-usage")

	if err != nil {
		return fmt.Errorf("failed to update storage usage: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("user service returned status %d: %s", resp.StatusCode(), resp.String())
	}

	if response.Code != 200 {
		return fmt.Errorf("user service error: %s", response.Message)
	}

	return nil
}

// NotifyQuotaExceeded 通知配额超限
func (c *userServiceClient) NotifyQuotaExceeded(ctx context.Context, userID uint, quotaType string) error {
	requestBody := map[string]interface{}{
		"quota_type": quotaType,
		"timestamp":  time.Now(),
	}

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetPathParam("user_id", fmt.Sprintf("%d", userID)).
		SetBody(requestBody).
		SetResult(&response).
		Post("/api/v1/users/{user_id}/quota-exceeded")

	if err != nil {
		return fmt.Errorf("failed to notify quota exceeded: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("user service returned status %d: %s", resp.StatusCode(), resp.String())
	}

	if response.Code != 200 {
		return fmt.Errorf("user service error: %s", response.Message)
	}

	return nil
}

// MockUserServiceClient Mock用户服务客户端（用于测试和开发）
type MockUserServiceClient struct {
	users map[uint]*UserInfo
}

// NewMockUserServiceClient 创建Mock用户服务客户端
func NewMockUserServiceClient() UserServiceClient {
	return &MockUserServiceClient{
		users: map[uint]*UserInfo{
			1: {
				ID:             1,
				Username:       "admin",
				Email:          "admin@example.com",
				Role:           "admin",
				Status:         1,
				StorageQuota:   10 * 1024 * 1024 * 1024, // 10GB
				FileCountQuota: 50000,
				StorageUsed:    0,
				FileCount:      0,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				Settings:       make(map[string]interface{}),
			},
			2: {
				ID:             2,
				Username:       "testuser",
				Email:          "test@example.com",
				Role:           "user",
				Status:         1,
				StorageQuota:   5 * 1024 * 1024 * 1024, // 5GB
				FileCountQuota: 10000,
				StorageUsed:    0,
				FileCount:      0,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				Settings:       make(map[string]interface{}),
			},
		},
	}
}

func (m *MockUserServiceClient) GetUser(ctx context.Context, userID uint) (*UserInfo, error) {
	if user, exists := m.users[userID]; exists {
		return user, nil
	}
	return nil, fmt.Errorf("user not found: %d", userID)
}

func (m *MockUserServiceClient) GetUserByEmail(ctx context.Context, email string) (*UserInfo, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found with email: %s", email)
}

func (m *MockUserServiceClient) UpdateUserQuota(ctx context.Context, userID uint, storageQuota, fileCountQuota int64) error {
	if user, exists := m.users[userID]; exists {
		user.StorageQuota = storageQuota
		user.FileCountQuota = fileCountQuota
		user.UpdatedAt = time.Now()
		return nil
	}
	return fmt.Errorf("user not found: %d", userID)
}

func (m *MockUserServiceClient) ValidateUserPermission(ctx context.Context, userID uint, resource, action string) (bool, error) {
	if user, exists := m.users[userID]; exists {
		// 简单权限检查：admin拥有所有权限
		return user.Role == "admin", nil
	}
	return false, fmt.Errorf("user not found: %d", userID)
}

func (m *MockUserServiceClient) GetUserRole(ctx context.Context, userID uint) (*UserRole, error) {
	if user, exists := m.users[userID]; exists {
		permissions := []string{"read"}
		if user.Role == "admin" {
			permissions = []string{"read", "write", "delete", "admin"}
		}
		return &UserRole{
			Role:        user.Role,
			Permissions: permissions,
		}, nil
	}
	return nil, fmt.Errorf("user not found: %d", userID)
}

func (m *MockUserServiceClient) UpdateUserStorageUsage(ctx context.Context, userID uint, usedStorage int64) error {
	if user, exists := m.users[userID]; exists {
		user.StorageUsed = usedStorage
		user.UpdatedAt = time.Now()
		return nil
	}
	return fmt.Errorf("user not found: %d", userID)
}

func (m *MockUserServiceClient) NotifyQuotaExceeded(ctx context.Context, userID uint, quotaType string) error {
	// Mock实现：只记录日志，不做实际通知
	fmt.Printf("Mock: Quota exceeded notification for user %d, type: %s\n", userID, quotaType)
	return nil
}