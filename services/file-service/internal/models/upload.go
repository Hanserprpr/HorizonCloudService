package models

import (
	"time"
)

// UploadSession 文件上传会话模型
type UploadSession struct {
	BaseModel
	
	// 基本信息
	SessionID    string `gorm:"not null;uniqueIndex;size:64" json:"session_id"`        // 会话ID
	FileName     string `gorm:"not null;size:255" json:"file_name"`                    // 原始文件名
	FileSize     int64  `gorm:"not null" json:"file_size"`                             // 文件总大小
	ContentType  string `gorm:"size:100" json:"content_type"`                          // MIME类型
	
	// 用户和目标信息
	UserID       uint   `gorm:"not null;index" json:"user_id"`                         // 用户ID
	FolderID     *uint  `gorm:"index" json:"folder_id,omitempty"`                      // 目标文件夹ID
	
	// 分片上传信息
	ChunkSize    int64  `gorm:"not null" json:"chunk_size"`                            // 分片大小
	TotalChunks  int    `gorm:"not null" json:"total_chunks"`                          // 总分片数
	UploadedChunks int  `gorm:"default:0" json:"uploaded_chunks"`                      // 已上传分片数
	UploadID     string `gorm:"size:100" json:"upload_id,omitempty"`                   // 分片上传ID
	
	// 状态和进度
	Status       int    `gorm:"not null;default:1;index" json:"status"`                // 状态
	Progress     float64 `gorm:"default:0" json:"progress"`                            // 上传进度(0-100)
	
	// 文件标识
	Hash         string `gorm:"size:64" json:"hash,omitempty"`                         // 文件哈希值
	TempPath     string `gorm:"size:1000" json:"temp_path,omitempty"`                  // 临时文件路径
	StoragePath  string `gorm:"size:1000" json:"storage_path,omitempty"`               // 最终存储路径
	
	// 时间控制
	ExpiresAt    time.Time `gorm:"index" json:"expires_at"`                            // 过期时间
	CompletedAt  *time.Time `json:"completed_at,omitempty"`                           // 完成时间
	
	// 错误信息
	ErrorMessage string `gorm:"size:1000" json:"error_message,omitempty"`              // 错误消息
	RetryCount   int    `gorm:"default:0" json:"retry_count"`                          // 重试次数
	
	// 关联关系
	Chunks       []UploadChunk `gorm:"foreignKey:SessionID;references:SessionID;constraint:OnDelete:CASCADE" json:"chunks,omitempty"`
	ResultFile   *File         `gorm:"foreignKey:ID;references:ID" json:"result_file,omitempty"`
}

// UploadChunk 上传分片模型
type UploadChunk struct {
	BaseModel
	
	// 关联信息
	SessionID    string `gorm:"not null;index;size:64" json:"session_id"`              // 会话ID
	
	// 分片信息
	ChunkIndex   int    `gorm:"not null" json:"chunk_index"`                           // 分片索引(从0开始)
	ChunkSize    int64  `gorm:"not null" json:"chunk_size"`                            // 分片大小
	StartOffset  int64  `gorm:"not null" json:"start_offset"`                          // 起始偏移量
	EndOffset    int64  `gorm:"not null" json:"end_offset"`                            // 结束偏移量
	
	// 状态和校验
	Status       int    `gorm:"not null;default:1;index" json:"status"`                // 状态
	Hash         string `gorm:"size:64" json:"hash,omitempty"`                         // 分片哈希值
	ETag         string `gorm:"size:64" json:"etag,omitempty"`                         // 存储系统返回的ETag
	
	// 存储信息
	TempPath     string    `gorm:"size:1000" json:"temp_path,omitempty"`               // 临时存储路径
	UploadedAt   *time.Time `json:"uploaded_at,omitempty"`                            // 上传时间
	
	// 错误处理
	ErrorMessage string `gorm:"size:500" json:"error_message,omitempty"`               // 错误消息
	RetryCount   int    `gorm:"default:0" json:"retry_count"`                          // 重试次数
}

// TableName 指定表名
func (UploadSession) TableName() string {
	return TablePrefix + "upload_sessions"
}

// TableName 指定表名
func (UploadChunk) TableName() string {
	return TablePrefix + "upload_chunks"
}

// UploadStatus 上传状态常量
const (
	UploadStatusPending     = 0 // 待处理 (alias for backward compatibility)
	UploadStatusInitialized = 1 // 已初始化
	UploadStatusUploading   = 2 // 上传中
	UploadStatusPaused      = 3 // 已暂停
	UploadStatusCompleted   = 4 // 已完成
	UploadStatusFailed      = 5 // 上传失败
	UploadStatusCanceled    = 6 // 已取消
	UploadStatusExpired     = 7 // 已过期
)

// ChunkStatus 分片状态常量
const (
	ChunkStatusPending    = 1 // 等待上传
	ChunkStatusUploading  = 2 // 上传中
	ChunkStatusCompleted  = 3 // 上传完成
	ChunkStatusFailed     = 4 // 上传失败
)

// IsCompleted 检查上传是否完成
func (us *UploadSession) IsCompleted() bool {
	return us.Status == UploadStatusCompleted
}

// IsFailed 检查上传是否失败
func (us *UploadSession) IsFailed() bool {
	return us.Status == UploadStatusFailed
}

// IsExpired 检查是否过期
func (us *UploadSession) IsExpired() bool {
	return us.Status == UploadStatusExpired || time.Now().After(us.ExpiresAt)
}

// CanUpload 检查是否可以继续上传
func (us *UploadSession) CanUpload() bool {
	return us.Status == UploadStatusInitialized || 
		   us.Status == UploadStatusUploading || 
		   us.Status == UploadStatusPaused
}

// UpdateProgress 更新上传进度
func (us *UploadSession) UpdateProgress() {
	if us.TotalChunks == 0 {
		us.Progress = 0
		return
	}
	us.Progress = float64(us.UploadedChunks) / float64(us.TotalChunks) * 100
}

// MarkCompleted 标记为完成
func (us *UploadSession) MarkCompleted() {
	us.Status = UploadStatusCompleted
	us.Progress = 100
	now := time.Now()
	us.CompletedAt = &now
}

// MarkFailed 标记为失败
func (us *UploadSession) MarkFailed(errorMsg string) {
	us.Status = UploadStatusFailed
	us.ErrorMessage = errorMsg
}

// GetStatusText 获取状态文本
func (us *UploadSession) GetStatusText() string {
	switch us.Status {
	case UploadStatusInitialized:
		return "已初始化"
	case UploadStatusUploading:
		return "上传中"
	case UploadStatusPaused:
		return "已暂停"
	case UploadStatusCompleted:
		return "已完成"
	case UploadStatusFailed:
		return "上传失败"
	case UploadStatusCanceled:
		return "已取消"
	case UploadStatusExpired:
		return "已过期"
	default:
		return "未知状态"
	}
}

// IsCompleted 检查分片是否完成
func (uc *UploadChunk) IsCompleted() bool {
	return uc.Status == ChunkStatusCompleted
}

// IsFailed 检查分片是否失败
func (uc *UploadChunk) IsFailed() bool {
	return uc.Status == ChunkStatusFailed
}

// MarkCompleted 标记分片完成
func (uc *UploadChunk) MarkCompleted(etag string) {
	uc.Status = ChunkStatusCompleted
	uc.ETag = etag
	now := time.Now()
	uc.UploadedAt = &now
}

// MarkFailed 标记分片失败
func (uc *UploadChunk) MarkFailed(errorMsg string) {
	uc.Status = ChunkStatusFailed
	uc.ErrorMessage = errorMsg
	uc.RetryCount++
}