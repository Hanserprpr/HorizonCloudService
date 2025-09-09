package repository

import (
	"gorm.io/gorm"
)

// Repository 仓库注册器，管理所有仓库实例
type Repository struct {
	File      FileRepository
	Folder    FolderRepository
	Upload    UploadRepository
	Thumbnail ThumbnailRepository
	Share     ShareRepository
	db        *gorm.DB
}

// NewRepository 创建仓库注册器实例
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		File:      NewFileRepository(db),
		Folder:    NewFolderRepository(db),
		Upload:    NewUploadRepository(db),
		Thumbnail: NewThumbnailRepository(db),
		Share:     NewShareRepository(db),
		db:        db,
	}
}

// GetDB 获取数据库连接
func (r *Repository) GetDB() *gorm.DB {
	return r.db
}

// BeginTx 开始事务
func (r *Repository) BeginTx() *gorm.DB {
	return r.db.Begin()
}

// WithTx 使用事务创建新的仓库实例
func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{
		File:      NewFileRepository(tx),
		Folder:    NewFolderRepository(tx),
		Upload:    NewUploadRepository(tx),
		Thumbnail: NewThumbnailRepository(tx),
		Share:     NewShareRepository(tx),
		db:        tx,
	}
}