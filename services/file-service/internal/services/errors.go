package services

import (
	"errors"
	"fmt"
	"time"
)

var (
	// 基础错误
	ErrServiceUnavailable = errors.New("service temporarily unavailable")
	ErrInvalidRequest     = errors.New("invalid request parameters")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrResourceNotFound   = errors.New("resource not found")
	ErrResourceConflict   = errors.New("resource conflict")
	
	// 文件操作错误
	ErrFileNotFound        = errors.New("file not found")
	ErrFileAlreadyExists   = errors.New("file already exists")
	ErrInvalidFileName     = errors.New("invalid file name")
	ErrFileTooLarge        = errors.New("file too large")
	ErrUnsupportedFileType = errors.New("unsupported file type")
	ErrFileCorrupted       = errors.New("file corrupted")
	
	// 文件夹操作错误
	ErrFolderNotFound      = errors.New("folder not found")
	ErrFolderNotEmpty      = errors.New("folder not empty")
	ErrInvalidFolderName   = errors.New("invalid folder name")
	ErrCircularReference   = errors.New("circular reference detected")
	ErrSystemFolder        = errors.New("cannot modify system folder")
	
	// 上传错误
	ErrUploadSessionNotFound = errors.New("upload session not found")
	ErrUploadSessionExpired  = errors.New("upload session expired")
	ErrInvalidChunkIndex     = errors.New("invalid chunk index")
	ErrChunkAlreadyUploaded  = errors.New("chunk already uploaded")
	ErrIncompleteUpload      = errors.New("upload incomplete")
	ErrUploadAborted         = errors.New("upload aborted")
	
	// 分享错误
	ErrShareNotFound        = errors.New("share not found")
	ErrShareExpired         = errors.New("share expired")
	ErrSharePasswordInvalid = errors.New("invalid share password")
	ErrShareAccessDenied    = errors.New("share access denied")
	ErrShareLimitReached    = errors.New("share download limit reached")
	
	// 配额和限制错误
	ErrQuotaExceeded       = errors.New("storage quota exceeded")
	ErrTooManyFiles        = errors.New("too many files")
	ErrTooManyFolders      = errors.New("too many folders")
	ErrRateLimitExceeded   = errors.New("rate limit exceeded")
	
	// 存储错误
	ErrStorageUnavailable = errors.New("storage service unavailable")
	ErrStorageQuotaFull   = errors.New("storage quota full")
	ErrStorageCorruption  = errors.New("storage corruption detected")
)

// ServiceError 服务错误结构体
type ServiceError struct {
	Type      string `json:"type"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Cause     error  `json:"-"`
}

// Error 实现error接口
func (e *ServiceError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s", e.Type, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap 支持error.Unwrap
func (e *ServiceError) Unwrap() error {
	return e.Cause
}

// NewServiceError 创建服务错误
func NewServiceError(errorType, message string, cause error) *ServiceError {
	return &ServiceError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}

// NewServiceErrorWithCode 创建带错误码的服务错误
func NewServiceErrorWithCode(errorType, code, message string, cause error) *ServiceError {
	return &ServiceError{
		Type:    errorType,
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewServiceErrorWithDetails 创建带详细信息的服务错误
func NewServiceErrorWithDetails(errorType, message, details string, cause error) *ServiceError {
	return &ServiceError{
		Type:    errorType,
		Message: message,
		Details: details,
		Cause:   cause,
	}
}

// 错误类型常量
const (
	// 请求错误
	ErrorTypeInvalidRequest = "InvalidRequest"
	ErrorTypeMissingParam   = "MissingParameter"
	ErrorTypeInvalidParam   = "InvalidParameter"
	
	// 权限错误
	ErrorTypePermissionDenied = "PermissionDenied"
	ErrorTypeAccessDenied     = "AccessDenied"
	ErrorTypeUnauthorized     = "Unauthorized"
	
	// 资源错误
	ErrorTypeResourceNotFound = "ResourceNotFound"
	ErrorTypeResourceConflict = "ResourceConflict"
	ErrorTypeResourceLocked   = "ResourceLocked"
	
	// 文件错误
	ErrorTypeFileNotFound        = "FileNotFound"
	ErrorTypeFileAlreadyExists   = "FileAlreadyExists"
	ErrorTypeInvalidFileName     = "InvalidFileName"
	ErrorTypeFileTooLarge        = "FileTooLarge"
	ErrorTypeUnsupportedFileType = "UnsupportedFileType"
	ErrorTypeFileCorrupted       = "FileCorrupted"
	ErrorTypeHashMismatch        = "HashMismatch"
	
	// 文件夹错误
	ErrorTypeFolderNotFound    = "FolderNotFound"
	ErrorTypeFolderNotEmpty    = "FolderNotEmpty"
	ErrorTypeInvalidFolderName = "InvalidFolderName"
	ErrorTypeCircularReference = "CircularReference"
	ErrorTypeSystemFolder      = "SystemFolder"
	
	// 上传错误
	ErrorTypeUploadSessionNotFound = "UploadSessionNotFound"
	ErrorTypeUploadSessionExpired  = "UploadSessionExpired"
	ErrorTypeInvalidChunkIndex     = "InvalidChunkIndex"
	ErrorTypeChunkAlreadyUploaded  = "ChunkAlreadyUploaded"
	ErrorTypeIncompleteUpload      = "IncompleteUpload"
	ErrorTypeUploadAborted         = "UploadAborted"
	ErrorTypeMultipartInitFailed   = "MultipartInitFailed"
	ErrorTypeChunkUploadFailed     = "ChunkUploadFailed"
	
	// 分享错误
	ErrorTypeShareNotFound        = "ShareNotFound"
	ErrorTypeShareExpired         = "ShareExpired"
	ErrorTypeSharePasswordInvalid = "SharePasswordInvalid"
	ErrorTypeShareAccessDenied    = "ShareAccessDenied"
	ErrorTypeShareLimitReached    = "ShareLimitReached"
	
	// 配额错误
	ErrorTypeQuotaExceeded     = "QuotaExceeded"
	ErrorTypeTooManyFiles      = "TooManyFiles"
	ErrorTypeTooManyFolders    = "TooManyFolders"
	ErrorTypeRateLimitExceeded = "RateLimitExceeded"
	
	// 存储错误
	ErrorTypeStorageUnavailable = "StorageUnavailable"
	ErrorTypeStorageQuotaFull   = "StorageQuotaFull"
	ErrorTypeStorageCorruption  = "StorageCorruption"
	ErrorTypeStorageTimeout     = "StorageTimeout"
	
	// 系统错误
	ErrorTypeInternalError      = "InternalError"
	ErrorTypeServiceUnavailable = "ServiceUnavailable"
	ErrorTypeTimeout            = "Timeout"
	ErrorTypeDatabaseError      = "DatabaseError"
)

// 错误检查函数

// IsServiceError 检查是否为服务错误
func IsServiceError(err error) bool {
	var serviceErr *ServiceError
	return errors.As(err, &serviceErr)
}

// GetServiceError 获取服务错误
func GetServiceError(err error) *ServiceError {
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		return serviceErr
	}
	return nil
}

// IsErrorType 检查是否为特定类型的错误
func IsErrorType(err error, errorType string) bool {
	if serviceErr := GetServiceError(err); serviceErr != nil {
		return serviceErr.Type == errorType
	}
	return false
}

// IsPermissionError 检查是否为权限错误
func IsPermissionError(err error) bool {
	return IsErrorType(err, ErrorTypePermissionDenied) ||
		IsErrorType(err, ErrorTypeAccessDenied) ||
		IsErrorType(err, ErrorTypeUnauthorized)
}

// IsNotFoundError 检查是否为资源未找到错误
func IsNotFoundError(err error) bool {
	return IsErrorType(err, ErrorTypeResourceNotFound) ||
		IsErrorType(err, ErrorTypeFileNotFound) ||
		IsErrorType(err, ErrorTypeFolderNotFound) ||
		IsErrorType(err, ErrorTypeUploadSessionNotFound) ||
		IsErrorType(err, ErrorTypeShareNotFound)
}

// IsValidationError 检查是否为验证错误
func IsValidationError(err error) bool {
	return IsErrorType(err, ErrorTypeInvalidRequest) ||
		IsErrorType(err, ErrorTypeMissingParam) ||
		IsErrorType(err, ErrorTypeInvalidParam) ||
		IsErrorType(err, ErrorTypeInvalidFileName) ||
		IsErrorType(err, ErrorTypeInvalidFolderName)
}

// IsQuotaError 检查是否为配额错误
func IsQuotaError(err error) bool {
	return IsErrorType(err, ErrorTypeQuotaExceeded) ||
		IsErrorType(err, ErrorTypeTooManyFiles) ||
		IsErrorType(err, ErrorTypeTooManyFolders) ||
		IsErrorType(err, ErrorTypeStorageQuotaFull)
}

// IsStorageError 检查是否为存储错误
func IsStorageError(err error) bool {
	return IsErrorType(err, ErrorTypeStorageUnavailable) ||
		IsErrorType(err, ErrorTypeStorageQuotaFull) ||
		IsErrorType(err, ErrorTypeStorageCorruption) ||
		IsErrorType(err, ErrorTypeStorageTimeout)
}

// 预定义错误创建函数

// NewFileNotFoundError 创建文件未找到错误
func NewFileNotFoundError(fileID uint) *ServiceError {
	return NewServiceErrorWithDetails(
		ErrorTypeFileNotFound,
		"File not found",
		fmt.Sprintf("File with ID %d not found", fileID),
		ErrFileNotFound,
	)
}

// NewFolderNotFoundError 创建文件夹未找到错误
func NewFolderNotFoundError(folderID uint) *ServiceError {
	return NewServiceErrorWithDetails(
		ErrorTypeFolderNotFound,
		"Folder not found",
		fmt.Sprintf("Folder with ID %d not found", folderID),
		ErrFolderNotFound,
	)
}

// NewPermissionDeniedError 创建权限拒绝错误
func NewPermissionDeniedError(resource string) *ServiceError {
	return NewServiceErrorWithDetails(
		ErrorTypePermissionDenied,
		"Permission denied",
		fmt.Sprintf("Access denied to %s", resource),
		ErrPermissionDenied,
	)
}

// NewInvalidFileNameError 创建无效文件名错误
func NewInvalidFileNameError(fileName string) *ServiceError {
	return NewServiceErrorWithDetails(
		ErrorTypeInvalidFileName,
		"Invalid file name",
		fmt.Sprintf("File name '%s' is invalid", fileName),
		ErrInvalidFileName,
	)
}

// NewQuotaExceededError 创建配额超限错误
func NewQuotaExceededError(used, quota int64) *ServiceError {
	return NewServiceErrorWithDetails(
		ErrorTypeQuotaExceeded,
		"Storage quota exceeded",
		fmt.Sprintf("Used %d bytes out of %d bytes quota", used, quota),
		ErrQuotaExceeded,
	)
}

// NewUploadSessionNotFoundError 创建上传会话未找到错误
func NewUploadSessionNotFoundError(sessionID string) *ServiceError {
	return NewServiceErrorWithDetails(
		ErrorTypeUploadSessionNotFound,
		"Upload session not found",
		fmt.Sprintf("Upload session %s not found", sessionID),
		ErrUploadSessionNotFound,
	)
}

// NewShareNotFoundError 创建分享未找到错误
func NewShareNotFoundError(token string) *ServiceError {
	return NewServiceErrorWithDetails(
		ErrorTypeShareNotFound,
		"Share not found",
		fmt.Sprintf("Share with token %s not found", token),
		ErrShareNotFound,
	)
}

// 错误转换函数

// ConvertStorageError 转换存储错误为服务错误
func ConvertStorageError(err error) *ServiceError {
	if err == nil {
		return nil
	}
	
	// TODO: 根据存储错误类型转换为对应的服务错误
	return NewServiceError(ErrorTypeStorageUnavailable, "Storage operation failed", err)
}

// ConvertDatabaseError 转换数据库错误为服务错误
func ConvertDatabaseError(err error) *ServiceError {
	if err == nil {
		return nil
	}
	
	// TODO: 根据数据库错误类型转换为对应的服务错误
	return NewServiceError(ErrorTypeDatabaseError, "Database operation failed", err)
}

// 错误统计和监控

// ErrorMetrics 错误统计指标
type ErrorMetrics struct {
	TotalErrors       int64            `json:"total_errors"`
	ErrorsByType      map[string]int64 `json:"errors_by_type"`
	ErrorsByCode      map[string]int64 `json:"errors_by_code"`
	RecentErrors      []*ServiceError  `json:"recent_errors"`
	ErrorRate         float64          `json:"error_rate"`
	LastErrorTime     int64            `json:"last_error_time"`
}

// ErrorCollector 错误收集器
type ErrorCollector struct {
	metrics    *ErrorMetrics
	maxRecent  int
	totalReqs  int64
}

// NewErrorCollector 创建错误收集器
func NewErrorCollector(maxRecent int) *ErrorCollector {
	return &ErrorCollector{
		metrics: &ErrorMetrics{
			ErrorsByType: make(map[string]int64),
			ErrorsByCode: make(map[string]int64),
			RecentErrors: make([]*ServiceError, 0, maxRecent),
		},
		maxRecent: maxRecent,
	}
}

// RecordError 记录错误
func (ec *ErrorCollector) RecordError(err *ServiceError) {
	if err == nil {
		return
	}
	
	ec.metrics.TotalErrors++
	ec.metrics.ErrorsByType[err.Type]++
	if err.Code != "" {
		ec.metrics.ErrorsByCode[err.Code]++
	}
	
	// 记录最近的错误
	if len(ec.metrics.RecentErrors) >= ec.maxRecent {
		ec.metrics.RecentErrors = ec.metrics.RecentErrors[1:]
	}
	ec.metrics.RecentErrors = append(ec.metrics.RecentErrors, err)
	
	ec.metrics.LastErrorTime = time.Now().Unix()
}

// RecordRequest 记录请求
func (ec *ErrorCollector) RecordRequest() {
	ec.totalReqs++
	if ec.totalReqs > 0 {
		ec.metrics.ErrorRate = float64(ec.metrics.TotalErrors) / float64(ec.totalReqs) * 100
	}
}

// GetMetrics 获取错误统计
func (ec *ErrorCollector) GetMetrics() *ErrorMetrics {
	return ec.metrics
}