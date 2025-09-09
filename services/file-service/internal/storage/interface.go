package storage

import (
	"context"
	"io"
	"time"
)

// Storage 存储接口，支持多种存储后端
type Storage interface {
	// 文件基本操作
	Upload(ctx context.Context, key string, reader io.Reader, size int64, options *UploadOptions) (*UploadResult, error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	GetSize(ctx context.Context, key string) (int64, error)
	
	// 文件信息
	GetInfo(ctx context.Context, key string) (*FileInfo, error)
	GetURL(ctx context.Context, key string, expires time.Duration) (string, error)
	GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error)
	
	// 分片上传
	InitiateMultipartUpload(ctx context.Context, key string, options *MultipartUploadOptions) (*MultipartUploadResult, error)
	UploadPart(ctx context.Context, uploadID, key string, partNumber int, reader io.Reader, size int64) (*UploadPartResult, error)
	CompleteMultipartUpload(ctx context.Context, uploadID, key string, parts []UploadPartInfo) (*CompleteMultipartUploadResult, error)
	AbortMultipartUpload(ctx context.Context, uploadID, key string) error
	ListParts(ctx context.Context, uploadID, key string) ([]PartInfo, error)
	
	// 目录操作
	ListObjects(ctx context.Context, prefix string, options *ListOptions) (*ListResult, error)
	CopyObject(ctx context.Context, srcKey, destKey string, options *CopyOptions) error
	MoveObject(ctx context.Context, srcKey, destKey string) error
	
	// 批量操作
	BatchDelete(ctx context.Context, keys []string) (*BatchDeleteResult, error)
	
	// 健康检查
	HealthCheck(ctx context.Context) error
}

// UploadOptions 上传选项
type UploadOptions struct {
	ContentType     string            `json:"content_type,omitempty"`
	ContentEncoding string            `json:"content_encoding,omitempty"`
	CacheControl    string            `json:"cache_control,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	StorageClass    string            `json:"storage_class,omitempty"` // STANDARD/STANDARD_IA/GLACIER
	ServerSideEncryption bool         `json:"server_side_encryption,omitempty"`
}

// UploadResult 上传结果
type UploadResult struct {
	Key      string `json:"key"`
	ETag     string `json:"etag"`
	Location string `json:"location,omitempty"`
	Size     int64  `json:"size"`
}

// FileInfo 文件信息
type FileInfo struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	ETag         string            `json:"etag"`
	LastModified time.Time         `json:"last_modified"`
	ContentType  string            `json:"content_type"`
	StorageClass string            `json:"storage_class,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// MultipartUploadOptions 分片上传选项
type MultipartUploadOptions struct {
	ContentType  string            `json:"content_type,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	StorageClass string            `json:"storage_class,omitempty"`
}

// MultipartUploadResult 分片上传初始化结果
type MultipartUploadResult struct {
	UploadID string `json:"upload_id"`
	Key      string `json:"key"`
}

// UploadPartResult 分片上传结果
type UploadPartResult struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size"`
}

// UploadPartInfo 分片信息
type UploadPartInfo struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size,omitempty"`
}

// CompleteMultipartUploadResult 完成分片上传结果
type CompleteMultipartUploadResult struct {
	Key      string `json:"key"`
	ETag     string `json:"etag"`
	Location string `json:"location,omitempty"`
	Size     int64  `json:"size"`
}

// PartInfo 分片详情
type PartInfo struct {
	PartNumber   int       `json:"part_number"`
	ETag         string    `json:"etag"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
}

// ListOptions 列举选项
type ListOptions struct {
	Delimiter   string `json:"delimiter,omitempty"`
	MaxKeys     int    `json:"max_keys,omitempty"`
	ContinuationToken string `json:"continuation_token,omitempty"`
	StartAfter  string `json:"start_after,omitempty"`
}

// ListResult 列举结果
type ListResult struct {
	Objects               []ObjectInfo `json:"objects"`
	CommonPrefixes        []string     `json:"common_prefixes,omitempty"`
	IsTruncated           bool         `json:"is_truncated"`
	NextContinuationToken string       `json:"next_continuation_token,omitempty"`
}

// ObjectInfo 对象信息
type ObjectInfo struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	ETag         string    `json:"etag"`
	LastModified time.Time `json:"last_modified"`
	StorageClass string    `json:"storage_class,omitempty"`
}

// CopyOptions 复制选项
type CopyOptions struct {
	MetadataDirective string            `json:"metadata_directive,omitempty"` // COPY/REPLACE
	Metadata          map[string]string `json:"metadata,omitempty"`
	StorageClass      string            `json:"storage_class,omitempty"`
}

// BatchDeleteResult 批量删除结果
type BatchDeleteResult struct {
	Deleted []string       `json:"deleted"`
	Errors  []DeleteError  `json:"errors,omitempty"`
}

// DeleteError 删除错误
type DeleteError struct {
	Key     string `json:"key"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// StorageType 存储类型
type StorageType string

const (
	StorageTypeMinIO StorageType = "minio"
	StorageTypeS3    StorageType = "s3"
	StorageTypeLocal StorageType = "local"
)

// Config 存储配置
type Config struct {
	Type            StorageType `json:"type"`
	Endpoint        string      `json:"endpoint,omitempty"`
	Region          string      `json:"region,omitempty"`
	Bucket          string      `json:"bucket"`
	AccessKeyID     string      `json:"access_key_id"`
	SecretAccessKey string      `json:"secret_access_key"`
	UseSSL          bool        `json:"use_ssl"`
	LocalPath       string      `json:"local_path,omitempty"`
}

// StorageManager 存储管理器
type StorageManager struct {
	primary Storage
	backup  Storage
	config  *Config
}

// NewStorageManager 创建存储管理器
func NewStorageManager(config *Config) (*StorageManager, error) {
	primary, err := NewStorage(config)
	if err != nil {
		return nil, err
	}
	
	return &StorageManager{
		primary: primary,
		config:  config,
	}, nil
}

// NewStorage 根据配置创建存储实例
func NewStorage(config *Config) (Storage, error) {
	switch config.Type {
	case StorageTypeMinIO:
		return NewMinIOStorage(config)
	case StorageTypeS3:
		return NewS3Storage(config)
	case StorageTypeLocal:
		return NewLocalStorage(config)
	default:
		return nil, ErrUnsupportedStorageType
	}
}

// GetPrimary 获取主存储
func (sm *StorageManager) GetPrimary() Storage {
	return sm.primary
}

// GetBackup 获取备份存储
func (sm *StorageManager) GetBackup() Storage {
	return sm.backup
}

// SetBackup 设置备份存储
func (sm *StorageManager) SetBackup(backup Storage) {
	sm.backup = backup
}