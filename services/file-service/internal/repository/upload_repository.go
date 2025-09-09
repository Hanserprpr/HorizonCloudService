package repository

import (
	"context"
	"file-service/internal/models"
	"time"

	"gorm.io/gorm"
)

// UploadRepository 上传仓库接口
type UploadRepository interface {
	// 上传会话管理
	CreateSession(ctx context.Context, session *models.UploadSession) error
	GetSession(ctx context.Context, sessionID string) (*models.UploadSession, error)
	GetUserSessions(ctx context.Context, userID uint, status *int, offset, limit int) ([]*models.UploadSession, int64, error)
	UpdateSession(ctx context.Context, session *models.UploadSession) error
	DeleteSession(ctx context.Context, sessionID string) error
	
	// 分片管理
	CreateChunk(ctx context.Context, chunk *models.UploadChunk) error
	GetChunk(ctx context.Context, sessionID string, chunkIndex int) (*models.UploadChunk, error)
	GetSessionChunks(ctx context.Context, sessionID string) ([]*models.UploadChunk, error)
	UpdateChunk(ctx context.Context, chunk *models.UploadChunk) error
	DeleteChunk(ctx context.Context, id uint) error
	
	// 会话状态管理
	UpdateSessionProgress(ctx context.Context, sessionID string) error
	CompleteSession(ctx context.Context, sessionID string, resultFileID uint) error
	FailSession(ctx context.Context, sessionID string, errorMsg string) error
	
	// 清理操作
	CleanExpiredSessions(ctx context.Context) (int64, error)
	CleanOrphanedChunks(ctx context.Context) (int64, error)
}

// uploadRepository 上传仓库实现
type uploadRepository struct {
	db *gorm.DB
}

// NewUploadRepository 创建上传仓库实例
func NewUploadRepository(db *gorm.DB) UploadRepository {
	return &uploadRepository{db: db}
}

// CreateSession 创建上传会话
func (r *uploadRepository) CreateSession(ctx context.Context, session *models.UploadSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetSession 获取上传会话
func (r *uploadRepository) GetSession(ctx context.Context, sessionID string) (*models.UploadSession, error) {
	var session models.UploadSession
	err := r.db.WithContext(ctx).
		Preload("Chunks").
		Where("session_id = ?", sessionID).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetUserSessions 获取用户上传会话列表
func (r *uploadRepository) GetUserSessions(ctx context.Context, userID uint, status *int, offset, limit int) ([]*models.UploadSession, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.UploadSession{}).Where("user_id = ?", userID)
	
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	var sessions []*models.UploadSession
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&sessions).Error
	
	return sessions, total, err
}

// UpdateSession 更新上传会话
func (r *uploadRepository) UpdateSession(ctx context.Context, session *models.UploadSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// DeleteSession 删除上传会话
func (r *uploadRepository) DeleteSession(ctx context.Context, sessionID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除所有分片
		if err := tx.Where("session_id = ?", sessionID).Delete(&models.UploadChunk{}).Error; err != nil {
			return err
		}
		
		// 删除会话
		return tx.Where("session_id = ?", sessionID).Delete(&models.UploadSession{}).Error
	})
}

// CreateChunk 创建上传分片
func (r *uploadRepository) CreateChunk(ctx context.Context, chunk *models.UploadChunk) error {
	return r.db.WithContext(ctx).Create(chunk).Error
}

// GetChunk 获取上传分片
func (r *uploadRepository) GetChunk(ctx context.Context, sessionID string, chunkIndex int) (*models.UploadChunk, error) {
	var chunk models.UploadChunk
	err := r.db.WithContext(ctx).
		Where("session_id = ? AND chunk_index = ?", sessionID, chunkIndex).
		First(&chunk).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

// GetSessionChunks 获取会话的所有分片
func (r *uploadRepository) GetSessionChunks(ctx context.Context, sessionID string) ([]*models.UploadChunk, error) {
	var chunks []*models.UploadChunk
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("chunk_index ASC").
		Find(&chunks).Error
	return chunks, err
}

// UpdateChunk 更新上传分片
func (r *uploadRepository) UpdateChunk(ctx context.Context, chunk *models.UploadChunk) error {
	return r.db.WithContext(ctx).Save(chunk).Error
}

// DeleteChunk 删除上传分片
func (r *uploadRepository) DeleteChunk(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.UploadChunk{}, id).Error
}

// UpdateSessionProgress 更新会话进度
func (r *uploadRepository) UpdateSessionProgress(ctx context.Context, sessionID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 计算已完成的分片数
		var completedChunks int64
		if err := tx.Model(&models.UploadChunk{}).
			Where("session_id = ? AND status = ?", sessionID, models.ChunkStatusCompleted).
			Count(&completedChunks).Error; err != nil {
			return err
		}
		
		// 更新会话进度
		return tx.Model(&models.UploadSession{}).
			Where("session_id = ?", sessionID).
			Updates(map[string]interface{}{
				"uploaded_chunks": completedChunks,
				"progress":        gorm.Expr("CASE WHEN total_chunks = 0 THEN 0 ELSE (? * 100.0 / total_chunks) END", completedChunks),
			}).Error
	})
}

// CompleteSession 完成上传会话
func (r *uploadRepository) CompleteSession(ctx context.Context, sessionID string, resultFileID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.UploadSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"status":       models.UploadStatusCompleted,
			"progress":     100,
			"completed_at": &now,
		}).Error
}

// FailSession 标记会话失败
func (r *uploadRepository) FailSession(ctx context.Context, sessionID string, errorMsg string) error {
	return r.db.WithContext(ctx).
		Model(&models.UploadSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"status":        models.UploadStatusFailed,
			"error_message": errorMsg,
		}).Error
}

// CleanExpiredSessions 清理过期会话
func (r *uploadRepository) CleanExpiredSessions(ctx context.Context) (int64, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Where("expires_at < ? AND status NOT IN (?)", 
			now, []int{models.UploadStatusCompleted, models.UploadStatusFailed}).
		Delete(&models.UploadSession{})
	
	return result.RowsAffected, result.Error
}

// CleanOrphanedChunks 清理孤立分片
func (r *uploadRepository) CleanOrphanedChunks(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("session_id NOT IN (SELECT session_id FROM file_upload_sessions)").
		Delete(&models.UploadChunk{})
	
	return result.RowsAffected, result.Error
}