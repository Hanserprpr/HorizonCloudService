package services

import (
	"context"
	"file-service/internal/models"
	"file-service/internal/repository"
	"io"
)

// FileService 文件服务接口
type FileService interface {
	// 文件上传
	UploadFile(ctx context.Context, req *UploadFileRequest) (*UploadFileResponse, error)
	CompleteUpload(ctx context.Context, req *CompleteUploadRequest) (*CompleteUploadResponse, error)
	
	// 文件查询
	GetFile(ctx context.Context, fileID uint, userID uint) (*models.File, error)
	GetFileByPath(ctx context.Context, path string, userID uint) (*models.File, error)
	ListFiles(ctx context.Context, req *ListFilesRequest) (*ListFilesResponse, error)
	SearchFiles(ctx context.Context, req *SearchFilesRequest) (*SearchFilesResponse, error)
	GetRecentFiles(ctx context.Context, userID uint, days, limit int) ([]*models.File, error)
	
	// 文件操作
	UpdateFile(ctx context.Context, req *UpdateFileRequest) error
	MoveFile(ctx context.Context, fileID, folderID uint, userID uint) error
	CopyFile(ctx context.Context, fileID uint, destFolderID uint, userID uint) (*models.File, error)
	RenameFile(ctx context.Context, fileID uint, newName string, userID uint) error
	DeleteFile(ctx context.Context, fileID uint, userID uint) error
	RestoreFile(ctx context.Context, fileID uint, userID uint) error
	
	// 文件下载
	DownloadFile(ctx context.Context, fileID uint, userID uint) (*DownloadFileResponse, error)
	GetFileURL(ctx context.Context, fileID uint, userID uint, expires int) (string, error)
	GetDownloadURL(ctx context.Context, fileID uint, userID uint) (string, error)
	
	// 文件版本控制
	GetFileVersions(ctx context.Context, fileID uint, userID uint) ([]*models.File, error)
	CreateFileVersion(ctx context.Context, fileID uint, req *UploadFileRequest) (*models.File, error)
	RestoreFileVersion(ctx context.Context, versionID uint, userID uint) error
	
	// 批量操作
	BatchMoveFiles(ctx context.Context, fileIDs []uint, folderID uint, userID uint) (*BatchOperationResponse, error)
	BatchDeleteFiles(ctx context.Context, fileIDs []uint, userID uint) (*BatchOperationResponse, error)
	BatchCopyFiles(ctx context.Context, fileIDs []uint, folderID uint, userID uint) (*BatchOperationResponse, error)
	
	// 统计信息
	GetUserStats(ctx context.Context, userID uint) (*UserFileStats, error)
	GetCategoryStats(ctx context.Context, userID uint) (map[string]int64, error)
	GetStorageStats(ctx context.Context, userID uint) (*StorageStats, error)
	GetAdminFileStats(ctx context.Context) (*AdminFileStats, error)
}

// FolderService 文件夹服务接口
type FolderService interface {
	// 文件夹CRUD
	CreateFolder(ctx context.Context, req *CreateFolderRequest) (*models.Folder, error)
	GetFolder(ctx context.Context, folderID uint, userID uint) (*models.Folder, error)
	GetFolderByPath(ctx context.Context, path string, userID uint) (*models.Folder, error)
	UpdateFolder(ctx context.Context, req *UpdateFolderRequest) error
	DeleteFolder(ctx context.Context, folderID uint, userID uint) error
	
	// 文件夹导航
	ListFolders(ctx context.Context, userID uint, parentID *uint) ([]*models.Folder, error)
	GetFolderTree(ctx context.Context, userID uint, rootID *uint) ([]*models.Folder, error)
	GetFolderPath(ctx context.Context, folderID uint, userID uint) ([]models.Folder, error)
	
	// 文件夹操作
	MoveFolder(ctx context.Context, folderID, newParentID uint, userID uint) error
	CopyFolder(ctx context.Context, folderID, destParentID uint, userID uint) (*models.Folder, error)
	RenameFolder(ctx context.Context, folderID uint, newName string, userID uint) error
	
	// 文件夹内容
	GetFolderContents(ctx context.Context, req *GetFolderContentsRequest) (*GetFolderContentsResponse, error)
	GetFolderStats(ctx context.Context, folderID uint, userID uint) (*repository.FolderStats, error)
	
	// 系统文件夹
	CreateSystemFolders(ctx context.Context, userID uint) error
	GetSystemFolder(ctx context.Context, userID uint, folderType string) (*models.Folder, error)
	
	// 统计和维护
	SyncFolderStats(ctx context.Context, folderID uint, userID uint) error
	RecalculateAllStats(ctx context.Context, userID uint) error
	SearchFolders(ctx context.Context, userID uint, keyword string, offset, limit int) ([]*models.Folder, int64, error)
}

// UploadService 上传服务接口
type UploadService interface {
	// 单文件上传
	InitiateUpload(ctx context.Context, req *InitiateUploadRequest) (*InitiateUploadResponse, error)
	UploadChunk(ctx context.Context, req *UploadChunkRequest) (*UploadChunkResponse, error)
	CompleteUpload(ctx context.Context, sessionID string, userID uint) (*CompleteUploadResponse, error)
	AbortUpload(ctx context.Context, sessionID string, userID uint) error
	
	// 上传会话管理
	GetUploadSession(ctx context.Context, sessionID string, userID uint) (*models.UploadSession, error)
	ListUploadSessions(ctx context.Context, userID uint, status *int, offset, limit int) ([]*models.UploadSession, int64, error)
	GetUploadProgress(ctx context.Context, sessionID string, userID uint) (*UploadProgressResponse, error)
	
	// 批量上传
	BatchInitiateUpload(ctx context.Context, req *BatchInitiateUploadRequest) (*BatchInitiateUploadResponse, error)
	
	// 断点续传
	ResumeUpload(ctx context.Context, sessionID string, userID uint) (*ResumeUploadResponse, error)
	
	// 清理操作
	CleanupExpiredSessions(ctx context.Context) (int64, error)
	CleanupOrphanedChunks(ctx context.Context) (int64, error)
	
	// 上传控制
	PauseUpload(ctx context.Context, sessionID string, userID uint) error
	ResumeUploadFromPause(ctx context.Context, sessionID string, userID uint) error
	GetUploadStatistics(ctx context.Context, userID uint) (*UploadStatistics, error)
}

// ShareService 分享服务接口
type ShareService interface {
	// 分享创建
	CreateShare(ctx context.Context, req *CreateShareRequest) (*models.Share, error)
	UpdateShare(ctx context.Context, shareID uint, req *UpdateShareRequest, userID uint) error
	DeleteShare(ctx context.Context, shareID uint, userID uint) error
	
	// 分享查询
	GetShare(ctx context.Context, shareID uint, userID uint) (*models.Share, error)
	GetShareByToken(ctx context.Context, token string) (*models.Share, error)
	ListShares(ctx context.Context, req *ListSharesRequest) (*ListSharesResponse, error)
	
	// 分享访问
	AccessShare(ctx context.Context, token string, req *AccessShareRequest) (*AccessShareResponse, error)
	DownloadSharedFile(ctx context.Context, token string, req *AccessShareRequest) (*DownloadFileResponse, error)
	PreviewSharedFile(ctx context.Context, token string, req *AccessShareRequest) (*PreviewFileResponse, error)
	
	// 分享统计
	GetShareStats(ctx context.Context, shareID uint, userID uint) (*repository.ShareAccessStats, error)
	GetUserShareStats(ctx context.Context, userID uint) (*repository.UserShareStats, error)
	
	// 分享管理
	ExpireShare(ctx context.Context, shareID uint, userID uint) error
	BatchExpireShares(ctx context.Context, shareIDs []uint, userID uint) error
	BatchDeleteShares(ctx context.Context, shareIDs []uint, userID uint) error
}

// ThumbnailService 缩略图服务接口
type ThumbnailService interface {
	// 缩略图生成
	GenerateThumbnail(ctx context.Context, req *GenerateThumbnailRequest) (*models.Thumbnail, error)
	GenerateThumbnails(ctx context.Context, fileID uint, userID uint) ([]*models.Thumbnail, error)
	
	// 缩略图管理
	GetThumbnail(ctx context.Context, thumbnailID uint, userID uint) (*models.Thumbnail, error)
	GetFileThumbnails(ctx context.Context, fileID uint, userID uint) ([]*models.Thumbnail, error)
	DeleteThumbnail(ctx context.Context, thumbnailID uint, userID uint) error
	
	// 缩略图下载
	GetThumbnailURL(ctx context.Context, fileID uint, size string, userID uint) (string, error)
	DownloadThumbnail(ctx context.Context, fileID uint, size string, userID uint) (io.ReadCloser, string, error)
	
	// 批量操作
	BatchGenerateThumbnails(ctx context.Context, fileIDs []uint, userID uint) (*BatchOperationResponse, error)
	CleanupOrphanedThumbnails(ctx context.Context) (int64, error)
	
	// 缩略图刷新
	RefreshThumbnails(ctx context.Context, fileID uint, userID uint) ([]*models.Thumbnail, error)
	
	// 统计功能
	GetThumbnailStats(ctx context.Context, userID uint) (*ThumbnailStats, error)
}

// 请求和响应结构体

// UpdateFileRequest 更新文件请求
type UpdateFileRequest struct {
	FileID      uint              `json:"file_id"`
	UserID      uint              `json:"user_id"`
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	IsPublic    *bool             `json:"is_public,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// UploadFileRequest 上传文件请求
type UploadFileRequest struct {
	FileName    string            `json:"file_name"`
	ContentType string            `json:"content_type,omitempty"`
	Size        int64             `json:"size"`
	FolderID    *uint             `json:"folder_id,omitempty"`
	UserID      uint              `json:"user_id"`
	Tags        []string          `json:"tags,omitempty"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Reader      io.Reader         `json:"-"`
}

// UploadFileResponse 上传文件响应
type UploadFileResponse struct {
	File       *models.File `json:"file"`
	Message    string       `json:"message"`
	Duplicated bool         `json:"duplicated,omitempty"`
}

// CompleteUploadRequest 完成上传请求
type CompleteUploadRequest struct {
	SessionID string `json:"session_id"`
	UserID    uint   `json:"user_id"`
}

// CompleteUploadResponse 完成上传响应
type CompleteUploadResponse struct {
	File    *models.File `json:"file"`
	Message string       `json:"message"`
}

// ListFilesRequest 列举文件请求
type ListFilesRequest struct {
	UserID   uint                        `json:"user_id"`
	FolderID *uint                       `json:"folder_id,omitempty"`
	Offset   int                         `json:"offset"`
	Limit    int                         `json:"limit"`
	Filters  *repository.FileFilters     `json:"filters,omitempty"`
}

// ListFilesResponse 列举文件响应
type ListFilesResponse struct {
	Files []*models.File `json:"files"`
	Total int64          `json:"total"`
}

// SearchFilesRequest 搜索文件请求
type SearchFilesRequest struct {
	UserID  uint                        `json:"user_id"`
	Keyword string                      `json:"keyword"`
	Offset  int                         `json:"offset"`
	Limit   int                         `json:"limit"`
	Filters *repository.FileFilters     `json:"filters,omitempty"`
}

// SearchFilesResponse 搜索文件响应
type SearchFilesResponse struct {
	Files []*models.File `json:"files"`
	Total int64          `json:"total"`
}

// DownloadFileResponse 下载文件响应
type DownloadFileResponse struct {
	File       *models.File   `json:"file"`
	Reader     io.ReadCloser  `json:"-"`
	URL        string         `json:"url,omitempty"`
	ExpireTime int64          `json:"expire_time,omitempty"`
}

// CreateFolderRequest 创建文件夹请求
type CreateFolderRequest struct {
	Name        string `json:"name"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	UserID      uint   `json:"user_id"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// UpdateFolderRequest 更新文件夹请求
type UpdateFolderRequest struct {
	FolderID    uint   `json:"folder_id"`
	UserID      uint   `json:"user_id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// GetFolderContentsRequest 获取文件夹内容请求
type GetFolderContentsRequest struct {
	FolderID *uint `json:"folder_id,omitempty"`
	UserID   uint  `json:"user_id"`
	Offset   int   `json:"offset"`
	Limit    int   `json:"limit"`
}

// GetFolderContentsResponse 获取文件夹内容响应
type GetFolderContentsResponse struct {
	Folders []*models.Folder `json:"folders"`
	Files   []*models.File   `json:"files"`
	Total   int64            `json:"total"`
}

// BatchOperationResponse 批量操作响应
type BatchOperationResponse struct {
	Success   []uint           `json:"success"`
	Failed    []BatchError     `json:"failed,omitempty"`
	Total     int              `json:"total"`
	Message   string           `json:"message"`
}

// BatchError 批量操作错误
type BatchError struct {
	ID      uint   `json:"id"`
	Message string `json:"message"`
}

// UserFileStats 用户文件统计
type UserFileStats struct {
	TotalFiles     int64 `json:"total_files"`
	TotalSize      int64 `json:"total_size"`
	TotalFolders   int64 `json:"total_folders"`
	StorageUsed    int64 `json:"storage_used"`
	StorageQuota   int64 `json:"storage_quota"`
	UsagePercent   float64 `json:"usage_percent"`
}

// StorageStats 存储统计
type StorageStats struct {
	TotalSize       int64            `json:"total_size"`
	UsedSize        int64            `json:"used_size"`
	AvailableSize   int64            `json:"available_size"`
	FileCount       int64            `json:"file_count"`
	CategoryStats   map[string]int64 `json:"category_stats"`
	TierStats       map[string]int64 `json:"tier_stats"`
}

// AdminFileStats 管理员文件统计（前端期望格式）
type AdminFileStats struct {
	TotalFiles      int64                    `json:"total_files"`
	TotalSize       int64                    `json:"total_size"`
	FilesByType     map[string]int64         `json:"files_by_type"`
	StorageByUser   []UserStorageInfo       `json:"storage_by_user"`
}

// UserStorageInfo 用户存储信息（与repository.UserStorageInfo一致）
type UserStorageInfo struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	StorageUsed int64  `json:"storage_used"`
}

// 上传服务相关结构体

// InitiateUploadRequest 初始化上传请求
type InitiateUploadRequest struct {
	FileName    string            `json:"file_name"`
	ContentType string            `json:"content_type,omitempty"`
	Size        int64             `json:"size"`
	ChunkSize   int64             `json:"chunk_size,omitempty"`
	FolderID    *uint             `json:"folder_id,omitempty"`
	UserID      uint              `json:"user_id"`
	Tags        []string          `json:"tags,omitempty"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// InitiateUploadResponse 初始化上传响应
type InitiateUploadResponse struct {
	SessionID    string `json:"session_id"`
	ChunkSize    int64  `json:"chunk_size"`
	TotalChunks  int    `json:"total_chunks"`
	UploadURLs   []string `json:"upload_urls,omitempty"`
	ExpiresAt    int64  `json:"expires_at"`
}

// UploadChunkRequest 上传分片请求
type UploadChunkRequest struct {
	SessionID   string    `json:"session_id"`
	ChunkIndex  int       `json:"chunk_index"`
	ChunkData   io.Reader `json:"-"`
	ChunkSize   int64     `json:"chunk_size"`
	ChunkHash   string    `json:"chunk_hash,omitempty"`
	UserID      uint      `json:"user_id"`
}

// UploadChunkResponse 上传分片响应
type UploadChunkResponse struct {
	ChunkIndex int    `json:"chunk_index"`
	ETag       string `json:"etag"`
	Message    string `json:"message"`
}

// UploadProgressResponse 上传进度响应
type UploadProgressResponse struct {
	SessionID       string  `json:"session_id"`
	FileName        string  `json:"file_name"`
	TotalSize       int64   `json:"total_size"`
	UploadedSize    int64   `json:"uploaded_size"`
	TotalChunks     int     `json:"total_chunks"`
	UploadedChunks  int     `json:"uploaded_chunks"`
	Progress        float64 `json:"progress"`
	Status          int     `json:"status"`
	EstimatedTime   int64   `json:"estimated_time,omitempty"`
}

// BatchInitiateUploadRequest 批量初始化上传请求
type BatchInitiateUploadRequest struct {
	Files  []*InitiateUploadRequest `json:"files"`
	UserID uint                     `json:"user_id"`
}

// BatchInitiateUploadResponse 批量初始化上传响应
type BatchInitiateUploadResponse struct {
	Sessions []InitiateUploadResponse `json:"sessions"`
	Failed   []BatchError             `json:"failed,omitempty"`
	Total    int                      `json:"total"`
}

// ResumeUploadResponse 断点续传响应
type ResumeUploadResponse struct {
	SessionID      string `json:"session_id"`
	NextChunkIndex int    `json:"next_chunk_index"`
	UploadedChunks int    `json:"uploaded_chunks"`
	TotalChunks    int    `json:"total_chunks"`
	Progress       float64 `json:"progress"`
}

// 分享服务相关结构体

// CreateShareRequest 创建分享请求
type CreateShareRequest struct {
	FileID        *uint  `json:"file_id,omitempty"`
	FolderID      *uint  `json:"folder_id,omitempty"`
	UserID        uint   `json:"user_id"`
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	Password      string `json:"password,omitempty"`
	ExpiresAt     *int64 `json:"expires_at,omitempty"`
	MaxDownloads  *int   `json:"max_downloads,omitempty"`
	AllowPreview  bool   `json:"allow_preview"`
	AllowDownload bool   `json:"allow_download"`
	IsPublic      bool   `json:"is_public"`
}

// UpdateShareRequest 更新分享请求
type UpdateShareRequest struct {
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	Password      string `json:"password,omitempty"`
	ExpiresAt     *int64 `json:"expires_at,omitempty"`
	MaxDownloads  *int   `json:"max_downloads,omitempty"`
	AllowPreview  *bool  `json:"allow_preview,omitempty"`
	AllowDownload *bool  `json:"allow_download,omitempty"`
	IsPublic      *bool  `json:"is_public,omitempty"`
	Status        *int   `json:"status,omitempty"`
}

// ListSharesRequest 列举分享请求
type ListSharesRequest struct {
	UserID  uint                         `json:"user_id"`
	Offset  int                          `json:"offset"`
	Limit   int                          `json:"limit"`
	Filters *repository.ShareFilters     `json:"filters,omitempty"`
}

// ListSharesResponse 列举分享响应
type ListSharesResponse struct {
	Shares []*models.Share `json:"shares"`
	Total  int64           `json:"total"`
}

// AccessShareRequest 访问分享请求
type AccessShareRequest struct {
	Password  string `json:"password,omitempty"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent,omitempty"`
	Referer   string `json:"referer,omitempty"`
}

// AccessShareResponse 访问分享响应
type AccessShareResponse struct {
	Share       *models.Share `json:"share"`
	File        *models.File  `json:"file,omitempty"`
	Folder      *models.Folder `json:"folder,omitempty"`
	CanDownload bool          `json:"can_download"`
	CanPreview  bool          `json:"can_preview"`
	Message     string        `json:"message,omitempty"`
}

// PreviewFileResponse 预览文件响应
type PreviewFileResponse struct {
	File        *models.File   `json:"file"`
	PreviewURL  string         `json:"preview_url,omitempty"`
	Reader      io.ReadCloser  `json:"-"`
	ContentType string         `json:"content_type"`
	Size        int64          `json:"size"`
}

// 缩略图服务相关结构体

// GenerateThumbnailRequest 生成缩略图请求
type GenerateThumbnailRequest struct {
	FileID uint   `json:"file_id"`
	Size   string `json:"size"`
	UserID uint   `json:"user_id"`
}

// ThumbnailStats 缩略图统计
type ThumbnailStats struct {
	TotalThumbnails     int64            `json:"total_thumbnails"`
	FilesWithThumbnails int64            `json:"files_with_thumbnails"`
	TotalSize          int64            `json:"total_size"`
	CoveragePercent    float64          `json:"coverage_percent"`
	BySize             map[string]int64 `json:"by_size"`
}

// UploadStatistics 上传统计信息
type UploadStatistics struct {
	TotalSessions     int     `json:"total_sessions"`
	CompletedSessions int     `json:"completed_sessions"`
	ActiveSessions    int     `json:"active_sessions"`
	FailedSessions    int     `json:"failed_sessions"`
	PausedSessions    int     `json:"paused_sessions"`
	TotalBytes        int64   `json:"total_bytes"`
	UploadedBytes     int64   `json:"uploaded_bytes"`
	SuccessRate       float64 `json:"success_rate"`
}