package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型 - 与用户服务保持一致
type BaseModel struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// CreateTime 获取创建时间戳（毫秒）
func (b *BaseModel) CreateTime() int64 {
	return b.CreatedAt.UnixMilli()
}

// UpdateTime 获取更新时间戳（毫秒）
func (b *BaseModel) UpdateTime() int64 {
	return b.UpdatedAt.UnixMilli()
}

// TableName 表名前缀
const TablePrefix = "file_"