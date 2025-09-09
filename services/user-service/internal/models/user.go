package models

import (
	"gorm.io/gorm"
	"time"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// CreateTime 获取创建时间戳
func (m *BaseModel) CreateTime() int64 {
	return m.CreatedAt.UnixMilli()
}

// UpdateTime 获取更新时间戳
func (m *BaseModel) UpdateTime() int64 {
	return m.UpdatedAt.UnixMilli()
}

// User 用户模型
type User struct {
	BaseModel
	StudentID     string `gorm:"type:varchar(20);uniqueIndex;not null" json:"student_id"`
	Password      string `gorm:"type:varchar(255);not null" json:"-"`
	RoleID        int    `gorm:"default:1;not null" json:"role_id"`
	NickName      string `gorm:"type:varchar(50);not null" json:"nick_name"`
	Email         string `gorm:"type:varchar(100);uniqueIndex" json:"email"`
	Phone         string `gorm:"type:varchar(20)" json:"phone"`
	Avatar        string `gorm:"type:varchar(500)" json:"avatar"`
	Status        int    `gorm:"default:1;not null" json:"status"` // 1:正常 2:禁用
	StorageQuota  int64  `gorm:"default:5368709120" json:"storage_quota"` // 默认5GB
	StorageUsed   int64  `gorm:"default:0" json:"storage_used"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	LoginCount    int    `gorm:"default:0" json:"login_count"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserQuota 用户存储配额信息
type UserQuota struct {
	UserID       uint  `json:"user_id"`
	StorageQuota int64 `json:"storage_quota"`
	StorageUsed  int64 `json:"storage_used"`
	FileCount    int   `json:"file_count"`
	UsagePercent float64 `json:"usage_percent"`
}

// ActivityLog 用户活动日志
type ActivityLog struct {
	BaseModel
	UserID    uint   `gorm:"not null;index" json:"user_id"`
	Action    string `gorm:"type:varchar(50);not null" json:"action"`
	Resource  string `gorm:"type:varchar(100)" json:"resource"`
	Detail    string `gorm:"type:text" json:"detail"`
	IPAddress string `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent string `gorm:"type:text" json:"user_agent"`
}

// TableName 指定表名
func (ActivityLog) TableName() string {
	return "user_activity_logs"
}