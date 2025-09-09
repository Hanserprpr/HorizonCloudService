package repository

import (
	"context"
	"errors"
	"sync"
	"user-service/internal/models"
	"gorm.io/gorm"
)

// MockUserRepository Mock用户仓库实现，用于单元测试
type MockUserRepository struct {
	mu           sync.RWMutex
	users        map[uint]*models.User
	studentIDMap map[string]uint
	emailMap     map[string]uint
	activityLogs map[uint][]*models.ActivityLog
	nextID       uint
}

// NewMockUserRepository 创建Mock用户仓库
func NewMockUserRepository() UserRepository {
	return &MockUserRepository{
		users:        make(map[uint]*models.User),
		studentIDMap: make(map[string]uint),
		emailMap:     make(map[string]uint),
		activityLogs: make(map[uint][]*models.ActivityLog),
		nextID:       1,
	}
}

// Create 创建用户
func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 检查学号是否已存在
	if _, exists := m.studentIDMap[user.StudentID]; exists {
		return errors.New("student_id already exists")
	}
	
	// 检查邮箱是否已存在
	if user.Email != "" {
		if _, exists := m.emailMap[user.Email]; exists {
			return errors.New("email already exists")
		}
	}
	
	user.ID = m.nextID
	m.nextID++
	
	m.users[user.ID] = user
	m.studentIDMap[user.StudentID] = user.ID
	if user.Email != "" {
		m.emailMap[user.Email] = user.ID
	}
	
	return nil
}

// GetByID 根据ID获取用户
func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	user, exists := m.users[id]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

// GetByStudentID 根据学号获取用户
func (m *MockUserRepository) GetByStudentID(ctx context.Context, studentID string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	userID, exists := m.studentIDMap[studentID]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return m.users[userID], nil
}

// GetByEmail 根据邮箱获取用户
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	userID, exists := m.emailMap[email]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return m.users[userID], nil
}

// Update 更新用户信息
func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.users[user.ID]; !exists {
		return gorm.ErrRecordNotFound
	}
	
	m.users[user.ID] = user
	return nil
}

// Delete 删除用户
func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	user, exists := m.users[id]
	if !exists {
		return gorm.ErrRecordNotFound
	}
	
	delete(m.users, id)
	delete(m.studentIDMap, user.StudentID)
	if user.Email != "" {
		delete(m.emailMap, user.Email)
	}
	
	return nil
}

// List 分页获取用户列表
func (m *MockUserRepository) List(ctx context.Context, offset, limit int, status int) ([]*models.User, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var users []*models.User
	for _, user := range m.users {
		if status == 0 || user.Status == status {
			users = append(users, user)
		}
	}
	
	total := int64(len(users))
	
	// 简单分页
	start := offset
	end := offset + limit
	if start >= len(users) {
		return []*models.User{}, total, nil
	}
	if end > len(users) {
		end = len(users)
	}
	
	return users[start:end], total, nil
}

// Search 搜索用户
func (m *MockUserRepository) Search(ctx context.Context, keyword string, offset, limit int) ([]*models.User, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var users []*models.User
	for _, user := range m.users {
		// 简单的包含搜索
		if contains(user.StudentID, keyword) || contains(user.NickName, keyword) || contains(user.Email, keyword) {
			users = append(users, user)
		}
	}
	
	total := int64(len(users))
	
	// 简单分页
	start := offset
	end := offset + limit
	if start >= len(users) {
		return []*models.User{}, total, nil
	}
	if end > len(users) {
		end = len(users)
	}
	
	return users[start:end], total, nil
}

// GetUsersByRole 根据角色获取用户列表
func (m *MockUserRepository) GetUsersByRole(ctx context.Context, roleID int, offset, limit int) ([]*models.User, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var users []*models.User
	for _, user := range m.users {
		if user.RoleID == roleID {
			users = append(users, user)
		}
	}
	
	total := int64(len(users))
	
	// 简单分页
	start := offset
	end := offset + limit
	if start >= len(users) {
		return []*models.User{}, total, nil
	}
	if end > len(users) {
		end = len(users)
	}
	
	return users[start:end], total, nil
}

// UpdateStorageUsed 更新用户存储使用量
func (m *MockUserRepository) UpdateStorageUsed(ctx context.Context, userID uint, used int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	user, exists := m.users[userID]
	if !exists {
		return gorm.ErrRecordNotFound
	}
	
	user.StorageUsed = used
	return nil
}

// GetUserQuota 获取用户配额信息
func (m *MockUserRepository) GetUserQuota(ctx context.Context, userID uint) (*models.UserQuota, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	user, exists := m.users[userID]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	
	usagePercent := float64(user.StorageUsed) / float64(user.StorageQuota) * 100
	
	return &models.UserQuota{
		UserID:       user.ID,
		StorageQuota: user.StorageQuota,
		StorageUsed:  user.StorageUsed,
		FileCount:    0, // Mock暂时设为0
		UsagePercent: usagePercent,
	}, nil
}

// UpdateStorageQuota 更新用户存储配额
func (m *MockUserRepository) UpdateStorageQuota(ctx context.Context, userID uint, quota int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	user, exists := m.users[userID]
	if !exists {
		return gorm.ErrRecordNotFound
	}
	
	user.StorageQuota = quota
	return nil
}

// UpdateStatus 更新用户状态
func (m *MockUserRepository) UpdateStatus(ctx context.Context, userID uint, status int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	user, exists := m.users[userID]
	if !exists {
		return gorm.ErrRecordNotFound
	}
	
	user.Status = status
	return nil
}

// UpdateLoginInfo 更新用户登录信息
func (m *MockUserRepository) UpdateLoginInfo(ctx context.Context, userID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	user, exists := m.users[userID]
	if !exists {
		return gorm.ErrRecordNotFound
	}
	
	user.LoginCount++
	return nil
}

// CreateActivityLog 创建活动日志
func (m *MockUserRepository) CreateActivityLog(ctx context.Context, log *models.ActivityLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.activityLogs[log.UserID] == nil {
		m.activityLogs[log.UserID] = []*models.ActivityLog{}
	}
	
	m.activityLogs[log.UserID] = append(m.activityLogs[log.UserID], log)
	return nil
}

// GetUserActivityLogs 获取用户活动日志
func (m *MockUserRepository) GetUserActivityLogs(ctx context.Context, userID uint, offset, limit int) ([]*models.ActivityLog, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	logs := m.activityLogs[userID]
	if logs == nil {
		return []*models.ActivityLog{}, 0, nil
	}
	
	total := int64(len(logs))
	
	// 简单分页
	start := offset
	end := offset + limit
	if start >= len(logs) {
		return []*models.ActivityLog{}, total, nil
	}
	if end > len(logs) {
		end = len(logs)
	}
	
	return logs[start:end], total, nil
}

// 辅助函数：检查字符串包含
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0)
}