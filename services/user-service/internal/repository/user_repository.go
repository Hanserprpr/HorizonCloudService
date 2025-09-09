package repository

import (
	"context"
	"gorm.io/gorm"
	"user-service/internal/models"
)

// UserRepository 用户数据访问层接口
type UserRepository interface {
	// 用户基础操作
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uint) (*models.User, error)
	GetByStudentID(ctx context.Context, studentID string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error
	
	// 用户查询操作
	List(ctx context.Context, offset, limit int, status int) ([]*models.User, int64, error)
	Search(ctx context.Context, keyword string, offset, limit int) ([]*models.User, int64, error)
	GetUsersByRole(ctx context.Context, roleID int, offset, limit int) ([]*models.User, int64, error)
	
	// 存储配额管理
	UpdateStorageUsed(ctx context.Context, userID uint, used int64) error
	GetUserQuota(ctx context.Context, userID uint) (*models.UserQuota, error)
	UpdateStorageQuota(ctx context.Context, userID uint, quota int64) error
	
	// 用户状态管理
	UpdateStatus(ctx context.Context, userID uint, status int) error
	UpdateLoginInfo(ctx context.Context, userID uint) error
	
	// 活动日志
	CreateActivityLog(ctx context.Context, log *models.ActivityLog) error
	GetUserActivityLogs(ctx context.Context, userID uint, offset, limit int) ([]*models.ActivityLog, int64, error)
}

// userRepository GORM实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByStudentID 根据学号获取用户
func (r *userRepository) GetByStudentID(ctx context.Context, studentID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("student_id = ?", studentID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete 软删除用户
func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

// List 分页获取用户列表
func (r *userRepository) List(ctx context.Context, offset, limit int, status int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.User{})
	if status > 0 {
		query = query.Where("status = ?", status)
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 获取分页数据
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error
	return users, total, err
}

// Search 搜索用户
func (r *userRepository) Search(ctx context.Context, keyword string, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.User{}).Where(
		"student_id LIKE ? OR nick_name LIKE ? OR email LIKE ?",
		"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%",
	)
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 获取分页数据
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error
	return users, total, err
}

// GetUsersByRole 根据角色获取用户列表
func (r *userRepository) GetUsersByRole(ctx context.Context, roleID int, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.User{}).Where("role_id = ?", roleID)
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 获取分页数据
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error
	return users, total, err
}

// UpdateStorageUsed 更新用户存储使用量
func (r *userRepository) UpdateStorageUsed(ctx context.Context, userID uint, used int64) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("storage_used", used).Error
}

// GetUserQuota 获取用户配额信息
func (r *userRepository) GetUserQuota(ctx context.Context, userID uint) (*models.UserQuota, error) {
	var user models.User
	err := r.db.WithContext(ctx).Select("id, storage_quota, storage_used").First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	
	// 计算文件数量（需要连接file-service获取，这里暂时设为0）
	fileCount := 0
	usagePercent := float64(user.StorageUsed) / float64(user.StorageQuota) * 100
	
	return &models.UserQuota{
		UserID:       user.ID,
		StorageQuota: user.StorageQuota,
		StorageUsed:  user.StorageUsed,
		FileCount:    fileCount,
		UsagePercent: usagePercent,
	}, nil
}

// UpdateStorageQuota 更新用户存储配额
func (r *userRepository) UpdateStorageQuota(ctx context.Context, userID uint, quota int64) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("storage_quota", quota).Error
}

// UpdateStatus 更新用户状态
func (r *userRepository) UpdateStatus(ctx context.Context, userID uint, status int) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("status", status).Error
}

// UpdateLoginInfo 更新用户登录信息
func (r *userRepository) UpdateLoginInfo(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at": gorm.Expr("NOW()"),
			"login_count":   gorm.Expr("login_count + 1"),
		}).Error
}

// CreateActivityLog 创建活动日志
func (r *userRepository) CreateActivityLog(ctx context.Context, log *models.ActivityLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetUserActivityLogs 获取用户活动日志
func (r *userRepository) GetUserActivityLogs(ctx context.Context, userID uint, offset, limit int) ([]*models.ActivityLog, int64, error) {
	var logs []*models.ActivityLog
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.ActivityLog{}).Where("user_id = ?", userID)
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 获取分页数据
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}