package services

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"file-service/internal/models"
	"file-service/internal/repository"
	"file-service/internal/storage"
	"io"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
)

// uploadService 上传服务实现
type uploadService struct {
	repo    *repository.Repository
	storage storage.Storage
	utils   PathUtils
}

// NewUploadService 创建上传服务实例
func NewUploadService(repo *repository.Repository, storage storage.Storage) UploadService {
	return &uploadService{
		repo:    repo,
		storage: storage,
		utils:   PathUtils{},
	}
}

// InitiateUpload 初始化上传
func (s *uploadService) InitiateUpload(ctx context.Context, req *InitiateUploadRequest) (*InitiateUploadResponse, error) {
	// 验证文件名
	if err := s.utils.ValidateFileName(req.FileName); err != nil {
		return nil, err
	}
	
	// 验证文件大小
	if req.Size <= 0 {
		return nil, NewServiceError("InvalidFileSize", "file size must be greater than 0", nil)
	}
	
	// 检查文件夹权限 (跳过根目录 folder_id = 0)
	if req.FolderID != nil && *req.FolderID != 0 {
		folder, err := s.repo.Folder.GetByID(ctx, *req.FolderID)
		if err != nil {
			return nil, NewServiceError("FolderNotFound", "target folder not found", err)
		}
		if folder.UserID != req.UserID {
			return nil, NewServiceError("PermissionDenied", "access denied to target folder", nil)
		}
	}
	
	// 设置默认分片大小 (5MB)
	chunkSize := req.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 5 * 1024 * 1024 // 5MB
	}
	
	// 计算分片数量
	totalChunks := int(math.Ceil(float64(req.Size) / float64(chunkSize)))
	if totalChunks > 10000 { // 限制最大分片数
		return nil, NewServiceError("TooManyChunks", "file too large or chunk size too small", nil)
	}
	
	// 生成会话ID
	sessionID := s.generateSessionID()
	
	// 提取文件扩展名
	extension := s.utils.ExtractExtension(req.FileName)
	
	// 创建上传会话
	session := &models.UploadSession{
		SessionID:   sessionID,
		FileName:    req.FileName,
		FileSize:    req.Size,
		ContentType: s.getContentType(req.ContentType, extension),
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
		UserID:      req.UserID,
		FolderID:    req.FolderID,
		Status:      models.UploadStatusUploading,
		Progress:    0,
		ExpiresAt:   time.Now().Add(24 * time.Hour), // 24小时过期
	}
	
	// 如果是单文件上传，直接处理
	if totalChunks == 1 {
		session.Status = models.UploadStatusInitialized
	} else {
		// 初始化多分片上传
		storagePath := s.utils.GenerateStoragePath(req.UserID, "", extension)
		uploadResult, err := s.storage.InitiateMultipartUpload(ctx, storagePath, &storage.MultipartUploadOptions{
			ContentType: session.ContentType,
			Metadata:    req.Metadata,
		})
		if err != nil {
			return nil, NewServiceError("MultipartInitFailed", "failed to initiate multipart upload", err)
		}
		session.UploadID = uploadResult.UploadID
	}
	
	if err := s.repo.Upload.CreateSession(ctx, session); err != nil {
		return nil, NewServiceError("SessionCreateFailed", "failed to create upload session", err)
	}
	
	// 生成预签名上传URL（如果存储支持）
	var uploadURLs []string
	// TODO: 实现预签名URL生成
	
	return &InitiateUploadResponse{
		SessionID:   sessionID,
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
		UploadURLs:  uploadURLs,
		ExpiresAt:   session.ExpiresAt.Unix(),
	}, nil
}

// UploadChunk 上传分片
func (s *uploadService) UploadChunk(ctx context.Context, req *UploadChunkRequest) (*UploadChunkResponse, error) {
	// 获取上传会话
	session, err := s.repo.Upload.GetSession(ctx, req.SessionID)
	if err != nil {
		return nil, NewServiceError("SessionNotFound", "upload session not found", err)
	}
	
	// 检查权限
	if session.UserID != req.UserID {
		return nil, NewServiceError("PermissionDenied", "access denied to upload session", nil)
	}
	
	// 检查会话状态
	if session.Status != models.UploadStatusUploading {
		return nil, NewServiceError("InvalidSessionStatus", "upload session is not in uploading status", nil)
	}
	
	// 检查会话是否过期
	if time.Now().After(session.ExpiresAt) {
		return nil, NewServiceError("SessionExpired", "upload session has expired", nil)
	}
	
	// 验证分片索引
	if req.ChunkIndex < 0 || req.ChunkIndex >= session.TotalChunks {
		return nil, NewServiceError("InvalidChunkIndex", "chunk index out of range", nil)
	}
	
	// 检查分片是否已存在
	existingChunk, err := s.repo.Upload.GetChunk(ctx, req.SessionID, req.ChunkIndex)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, NewServiceError("ChunkCheckFailed", "failed to check existing chunk", err)
	}
	
	if existingChunk != nil && existingChunk.Status == models.ChunkStatusCompleted {
		return &UploadChunkResponse{
			ChunkIndex: req.ChunkIndex,
			ETag:       existingChunk.ETag,
			Message:    "Chunk already uploaded",
		}, nil
	}
	
	// 计算分片哈希
	chunkHash, reader, err := s.calculateChunkHash(req.ChunkData)
	if err != nil {
		return nil, NewServiceError("ChunkHashFailed", "failed to calculate chunk hash", err)
	}
	
	// 验证分片哈希（如果提供）
	if req.ChunkHash != "" && req.ChunkHash != chunkHash {
		return nil, NewServiceError("ChunkHashMismatch", "chunk hash mismatch", nil)
	}
	
	var etag string
	var storagePath string
	
	if session.TotalChunks == 1 {
		// 单文件上传
		extension := s.utils.ExtractExtension(session.FileName)
		storagePath = s.utils.GenerateStoragePath(session.UserID, chunkHash, extension)
		
		uploadResult, err := s.storage.Upload(ctx, storagePath, reader, req.ChunkSize, &storage.UploadOptions{
			ContentType: session.ContentType,
		})
		if err != nil {
			return nil, NewServiceError("ChunkUploadFailed", "failed to upload chunk", err)
		}
		etag = uploadResult.ETag
		session.Hash = chunkHash // 单文件的哈希就是文件哈希
	} else {
		// 分片上传
		storagePath = s.utils.GenerateChunkPath(session.UserID, req.SessionID, req.ChunkIndex)
		
		uploadResult, err := s.storage.UploadPart(ctx, session.UploadID, storagePath, req.ChunkIndex+1, reader, req.ChunkSize)
		if err != nil {
			return nil, NewServiceError("ChunkUploadFailed", "failed to upload chunk", err)
		}
		etag = uploadResult.ETag
	}
	
	// 创建或更新分片记录
	chunk := &models.UploadChunk{
		SessionID:  req.SessionID,
		ChunkIndex: req.ChunkIndex,
		ChunkSize:  req.ChunkSize,
		Hash:       chunkHash,
		ETag:       etag,
		Status:     models.ChunkStatusCompleted,
		TempPath:   storagePath,
	}
	
	err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		
		// 保存分片记录
		if existingChunk != nil {
			chunk.ID = existingChunk.ID
			if err := txRepo.Upload.UpdateChunk(ctx, chunk); err != nil {
				return err
			}
		} else {
			if err := txRepo.Upload.CreateChunk(ctx, chunk); err != nil {
				return err
			}
		}
		
		// 更新会话进度
		return txRepo.Upload.UpdateSessionProgress(ctx, req.SessionID)
	})
	
	if err != nil {
		return nil, NewServiceError("ChunkSaveFailed", "failed to save chunk record", err)
	}
	
	return &UploadChunkResponse{
		ChunkIndex: req.ChunkIndex,
		ETag:       etag,
		Message:    "Chunk uploaded successfully",
	}, nil
}

// CompleteUpload 完成上传
func (s *uploadService) CompleteUpload(ctx context.Context, sessionID string, userID uint) (*CompleteUploadResponse, error) {
	// 这个方法在FileService中已经实现，这里委托调用
	fileService := NewFileService(s.repo, s.storage)
	return fileService.CompleteUpload(ctx, &CompleteUploadRequest{
		SessionID: sessionID,
		UserID:    userID,
	})
}

// AbortUpload 中止上传
func (s *uploadService) AbortUpload(ctx context.Context, sessionID string, userID uint) error {
	// 获取上传会话
	session, err := s.repo.Upload.GetSession(ctx, sessionID)
	if err != nil {
		return NewServiceError("SessionNotFound", "upload session not found", err)
	}
	
	if session.UserID != userID {
		return NewServiceError("PermissionDenied", "access denied to upload session", nil)
	}
	
	return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		
		// 如果是分片上传，中止存储层的分片上传
		if session.TotalChunks > 1 && session.UploadID != "" {
			// 生成存储路径（用于中止操作）
			extension := s.utils.ExtractExtension(session.FileName)
			storagePath := s.utils.GenerateStoragePath(session.UserID, "", extension)
			if err := s.storage.AbortMultipartUpload(ctx, session.UploadID, storagePath); err != nil {
				// 记录错误但不阻止数据库清理
				// TODO: 添加日志
			}
		}
		
		// 更新会话状态
		if err := txRepo.Upload.FailSession(ctx, sessionID, "Upload aborted by user"); err != nil {
			return err
		}
		
		// 删除会话和分片记录
		return txRepo.Upload.DeleteSession(ctx, sessionID)
	})
}

// GetUploadSession 获取上传会话
func (s *uploadService) GetUploadSession(ctx context.Context, sessionID string, userID uint) (*models.UploadSession, error) {
	session, err := s.repo.Upload.GetSession(ctx, sessionID)
	if err != nil {
		return nil, NewServiceError("SessionNotFound", "upload session not found", err)
	}
	
	if session.UserID != userID {
		return nil, NewServiceError("PermissionDenied", "access denied to upload session", nil)
	}
	
	return session, nil
}

// ListUploadSessions 列举上传会话
func (s *uploadService) ListUploadSessions(ctx context.Context, userID uint, status *int, offset, limit int) ([]*models.UploadSession, int64, error) {
	return s.repo.Upload.GetUserSessions(ctx, userID, status, offset, limit)
}

// GetUploadProgress 获取上传进度
func (s *uploadService) GetUploadProgress(ctx context.Context, sessionID string, userID uint) (*UploadProgressResponse, error) {
	session, err := s.GetUploadSession(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	
	// 计算预估剩余时间
	var estimatedTime int64
	if session.Progress > 0 && session.Progress < 100 {
		elapsed := time.Since(session.CreatedAt)
		totalTime := float64(elapsed) / (session.Progress / 100.0)
		remaining := totalTime - float64(elapsed)
		if remaining > 0 {
			estimatedTime = int64(remaining) / int64(time.Second)
		}
	}
	
	return &UploadProgressResponse{
		SessionID:       sessionID,
		FileName:        session.FileName,
		TotalSize:       session.FileSize,
		UploadedSize:    int64(float64(session.FileSize) * session.Progress / 100.0),
		TotalChunks:     session.TotalChunks,
		UploadedChunks:  session.UploadedChunks,
		Progress:        session.Progress,
		Status:          session.Status,
		EstimatedTime:   estimatedTime,
	}, nil
}

// BatchInitiateUpload 批量初始化上传
func (s *uploadService) BatchInitiateUpload(ctx context.Context, req *BatchInitiateUploadRequest) (*BatchInitiateUploadResponse, error) {
	var sessions []InitiateUploadResponse
	var failed []BatchError
	
	for i, fileReq := range req.Files {
		fileReq.UserID = req.UserID
		
		resp, err := s.InitiateUpload(ctx, fileReq)
		if err != nil {
			failed = append(failed, BatchError{
				ID:      uint(i), // 使用索引作为ID
				Message: err.Error(),
			})
		} else {
			sessions = append(sessions, *resp)
		}
	}
	
	return &BatchInitiateUploadResponse{
		Sessions: sessions,
		Failed:   failed,
		Total:    len(req.Files),
	}, nil
}

// ResumeUpload 断点续传
func (s *uploadService) ResumeUpload(ctx context.Context, sessionID string, userID uint) (*ResumeUploadResponse, error) {
	session, err := s.GetUploadSession(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	
	if session.Status == models.UploadStatusCompleted {
		return nil, NewServiceError("UploadAlreadyCompleted", "upload already completed", nil)
	}
	
	if session.Status == models.UploadStatusFailed {
		return nil, NewServiceError("UploadFailed", "upload has failed, cannot resume", nil)
	}
	
	// 获取已上传的分片
	chunks, err := s.repo.Upload.GetSessionChunks(ctx, sessionID)
	if err != nil {
		return nil, NewServiceError("ChunkRetrieveFailed", "failed to retrieve uploaded chunks", err)
	}
	
	// 计算下一个需要上传的分片
	uploadedChunks := 0
	nextChunkIndex := 0
	chunkStatus := make([]bool, session.TotalChunks)
	
	for _, chunk := range chunks {
		if chunk.Status == models.ChunkStatusCompleted {
			uploadedChunks++
			chunkStatus[chunk.ChunkIndex] = true
		}
	}
	
	// 找到第一个未上传的分片
	for i, uploaded := range chunkStatus {
		if !uploaded {
			nextChunkIndex = i
			break
		}
	}
	
	progress := float64(uploadedChunks) / float64(session.TotalChunks) * 100
	
	return &ResumeUploadResponse{
		SessionID:      sessionID,
		NextChunkIndex: nextChunkIndex,
		UploadedChunks: uploadedChunks,
		TotalChunks:    session.TotalChunks,
		Progress:       progress,
	}, nil
}

// CleanupExpiredSessions 清理过期会话
func (s *uploadService) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	return s.repo.Upload.CleanExpiredSessions(ctx)
}

// CleanupOrphanedChunks 清理孤立分片
func (s *uploadService) CleanupOrphanedChunks(ctx context.Context) (int64, error) {
	return s.repo.Upload.CleanOrphanedChunks(ctx)
}

// 辅助方法

// generateSessionID 生成会话ID
func (s *uploadService) generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// calculateChunkHash 计算分片哈希
func (s *uploadService) calculateChunkHash(reader io.Reader) (string, io.Reader, error) {
	hash := md5.New()
	teeReader := io.TeeReader(reader, hash)
	
	// 读取所有数据到内存（分片通常不大）
	data, err := io.ReadAll(teeReader)
	if err != nil {
		return "", nil, err
	}
	
	hashStr := hex.EncodeToString(hash.Sum(nil))
	newReader := strings.NewReader(string(data))
	
	return hashStr, newReader, nil
}

// getContentType 获取内容类型
func (s *uploadService) getContentType(provided, extension string) string {
	if provided != "" {
		return provided
	}
	return s.utils.GetMimeType(extension)
}

// 高级上传功能

// PauseUpload 暂停上传
func (s *uploadService) PauseUpload(ctx context.Context, sessionID string, userID uint) error {
	session, err := s.GetUploadSession(ctx, sessionID, userID)
	if err != nil {
		return err
	}
	
	if session.Status != models.UploadStatusUploading {
		return NewServiceError("InvalidSessionStatus", "cannot pause upload in current status", nil)
	}
	
	session.Status = models.UploadStatusPaused
	return s.repo.Upload.UpdateSession(ctx, session)
}

// ResumeUploadFromPause 从暂停状态恢复上传
func (s *uploadService) ResumeUploadFromPause(ctx context.Context, sessionID string, userID uint) error {
	session, err := s.GetUploadSession(ctx, sessionID, userID)
	if err != nil {
		return err
	}
	
	if session.Status != models.UploadStatusPaused {
		return NewServiceError("InvalidSessionStatus", "upload is not paused", nil)
	}
	
	session.Status = models.UploadStatusUploading
	return s.repo.Upload.UpdateSession(ctx, session)
}

// GetUploadStatistics 获取上传统计
func (s *uploadService) GetUploadStatistics(ctx context.Context, userID uint) (*UploadStatistics, error) {
	sessions, _, err := s.repo.Upload.GetUserSessions(ctx, userID, nil, 0, 1000)
	if err != nil {
		return nil, NewServiceError("StatsRetrieveFailed", "failed to retrieve upload sessions", err)
	}
	
	stats := &UploadStatistics{
		TotalSessions: len(sessions),
	}
	
	for _, session := range sessions {
		switch session.Status {
		case models.UploadStatusCompleted:
			stats.CompletedSessions++
		case models.UploadStatusUploading:
			stats.ActiveSessions++
		case models.UploadStatusFailed:
			stats.FailedSessions++
		case models.UploadStatusPaused:
			stats.PausedSessions++
		}
		
		stats.TotalBytes += session.FileSize
		if session.Status == models.UploadStatusCompleted {
			stats.UploadedBytes += session.FileSize
		}
	}
	
	if stats.TotalBytes > 0 {
		stats.SuccessRate = float64(stats.UploadedBytes) / float64(stats.TotalBytes) * 100
	}
	
	return stats, nil
}

