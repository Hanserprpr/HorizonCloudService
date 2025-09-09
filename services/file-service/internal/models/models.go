package models

import (
	"gorm.io/gorm"
)

// AllModels 返回所有数据模型，用于数据库迁移
func AllModels() []interface{} {
	return []interface{}{
		&File{},
		&Folder{},
		&UploadSession{},
		&UploadChunk{},
		&Thumbnail{},
		&Share{},
		&ShareAccessLog{},
	}
}

// AutoMigrate 执行数据库自动迁移
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(AllModels()...)
}

// CreateIndexes 创建额外的数据库索引
func CreateIndexes(db *gorm.DB) error {
	// 复合索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_files_user_folder ON file_files(user_id, folder_id)").Error; err != nil {
		return err
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_files_user_category ON file_files(user_id, category)").Error; err != nil {
		return err
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_files_hash_size ON file_files(hash, size)").Error; err != nil {
		return err
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_folders_user_parent ON file_folders(user_id, parent_id)").Error; err != nil {
		return err
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_upload_sessions_user_status ON file_upload_sessions(user_id, status)").Error; err != nil {
		return err
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_shares_user_status ON file_shares(user_id, status)").Error; err != nil {
		return err
	}
	
	// 全文搜索索引(PostgreSQL)
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_files_name_gin ON file_files USING GIN(to_tsvector('simple', name))").Error; err != nil {
		return err
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_folders_name_gin ON file_folders USING GIN(to_tsvector('simple', name))").Error; err != nil {
		return err
	}
	
	return nil
}