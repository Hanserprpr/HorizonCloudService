package storage

import (
	"errors"
	"fmt"
)

var (
	// 基础错误
	ErrUnsupportedStorageType = errors.New("unsupported storage type")
	ErrStorageNotFound        = errors.New("storage not found")
	ErrStorageConnectionFailed = errors.New("storage connection failed")
	
	// 文件操作错误
	ErrFileNotFound     = errors.New("file not found")
	ErrFileAlreadyExists = errors.New("file already exists")
	ErrInvalidFileKey   = errors.New("invalid file key")
	ErrFileTooLarge     = errors.New("file too large")
	ErrFileCorrupted    = errors.New("file corrupted")
	
	// 分片上传错误
	ErrInvalidUploadID     = errors.New("invalid upload id")
	ErrInvalidPartNumber   = errors.New("invalid part number")
	ErrPartNotFound        = errors.New("part not found")
	ErrPartTooSmall        = errors.New("part too small")
	ErrIncompleteMultipart = errors.New("incomplete multipart upload")
	
	// 权限错误
	ErrAccessDenied      = errors.New("access denied")
	ErrInsufficientQuota = errors.New("insufficient quota")
	
	// 配置错误
	ErrInvalidConfig     = errors.New("invalid storage config")
	ErrMissingCredentials = errors.New("missing storage credentials")
	ErrInvalidBucket     = errors.New("invalid bucket name")
)

// StorageError 存储错误结构体
type StorageError struct {
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Key     string `json:"key,omitempty"`
	Cause   error  `json:"-"`
}

// Error 实现error接口
func (e *StorageError) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("%s: %s (key: %s)", e.Type, e.Message, e.Key)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap 支持error.Unwrap
func (e *StorageError) Unwrap() error {
	return e.Cause
}

// NewStorageError 创建存储错误
func NewStorageError(errorType, message string, cause error) *StorageError {
	return &StorageError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}

// NewFileNotFoundError 创建文件未找到错误
func NewFileNotFoundError(key string) *StorageError {
	return &StorageError{
		Type:    "FileNotFound",
		Message: "file not found",
		Key:     key,
		Cause:   ErrFileNotFound,
	}
}

// NewAccessDeniedError 创建访问拒绝错误
func NewAccessDeniedError(key, message string) *StorageError {
	return &StorageError{
		Type:    "AccessDenied",
		Message: message,
		Key:     key,
		Cause:   ErrAccessDenied,
	}
}

// NewInvalidUploadIDError 创建无效上传ID错误
func NewInvalidUploadIDError(uploadID string) *StorageError {
	return &StorageError{
		Type:    "InvalidUploadID",
		Message: "invalid upload id: " + uploadID,
		Cause:   ErrInvalidUploadID,
	}
}

// NewPartTooSmallError 创建分片过小错误
func NewPartTooSmallError(partNumber int, size int64) *StorageError {
	return &StorageError{
		Type:    "PartTooSmall",
		Message: fmt.Sprintf("part %d is too small: %d bytes", partNumber, size),
		Cause:   ErrPartTooSmall,
	}
}

// IsFileNotFoundError 检查是否为文件未找到错误
func IsFileNotFoundError(err error) bool {
	var storageErr *StorageError
	if errors.As(err, &storageErr) {
		return storageErr.Type == "FileNotFound" || errors.Is(storageErr.Cause, ErrFileNotFound)
	}
	return errors.Is(err, ErrFileNotFound)
}

// IsAccessDeniedError 检查是否为访问拒绝错误
func IsAccessDeniedError(err error) bool {
	var storageErr *StorageError
	if errors.As(err, &storageErr) {
		return storageErr.Type == "AccessDenied" || errors.Is(storageErr.Cause, ErrAccessDenied)
	}
	return errors.Is(err, ErrAccessDenied)
}

// IsInvalidUploadIDError 检查是否为无效上传ID错误
func IsInvalidUploadIDError(err error) bool {
	var storageErr *StorageError
	if errors.As(err, &storageErr) {
		return storageErr.Type == "InvalidUploadID" || errors.Is(storageErr.Cause, ErrInvalidUploadID)
	}
	return errors.Is(err, ErrInvalidUploadID)
}