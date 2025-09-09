package storage

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// PathUtils 路径工具函数
type PathUtils struct{}

// GenerateStoragePath 生成存储路径
// 格式: {userID}/{year}/{month}/{day}/{hash}.{ext}
func (PathUtils) GenerateStoragePath(userID uint, hash, extension string) string {
	now := time.Now()
	return fmt.Sprintf("%d/%04d/%02d/%02d/%s.%s",
		userID, now.Year(), now.Month(), now.Day(), hash, extension)
}

// GenerateThumbnailPath 生成缩略图路径
// 格式: {userID}/thumbnails/{hash}_{size}.jpg
func (PathUtils) GenerateThumbnailPath(userID uint, hash, size string) string {
	return fmt.Sprintf("%d/thumbnails/%s_%s.jpg", userID, hash, size)
}

// GenerateChunkPath 生成分片路径
// 格式: {userID}/chunks/{sessionID}/{chunkIndex}
func (PathUtils) GenerateChunkPath(userID uint, sessionID string, chunkIndex int) string {
	return fmt.Sprintf("%d/chunks/%s/%d", userID, sessionID, chunkIndex)
}

// CleanPath 清理路径，防止路径遍历攻击
func (PathUtils) CleanPath(path string) string {
	// 移除前导斜杠
	path = strings.TrimPrefix(path, "/")
	// 清理路径
	return filepath.Clean(path)
}

// ExtractExtension 提取文件扩展名
func (PathUtils) ExtractExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ""
	}
	return strings.TrimPrefix(ext, ".")
}

// GetMimeType 根据扩展名获取MIME类型
func (PathUtils) GetMimeType(extension string) string {
	mimeTypes := map[string]string{
		// 图片
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"webp": "image/webp",
		"bmp":  "image/bmp",
		"svg":  "image/svg+xml",
		"ico":  "image/x-icon",
		
		// 视频
		"mp4":  "video/mp4",
		"avi":  "video/x-msvideo",
		"mov":  "video/quicktime",
		"wmv":  "video/x-ms-wmv",
		"flv":  "video/x-flv",
		"webm": "video/webm",
		"mkv":  "video/x-matroska",
		
		// 音频
		"mp3":  "audio/mpeg",
		"wav":  "audio/wav",
		"flac": "audio/flac",
		"aac":  "audio/aac",
		"ogg":  "audio/ogg",
		"m4a":  "audio/mp4",
		
		// 文档
		"pdf":  "application/pdf",
		"doc":  "application/msword",
		"docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"xls":  "application/vnd.ms-excel",
		"xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"ppt":  "application/vnd.ms-powerpoint",
		"pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"txt":  "text/plain",
		"rtf":  "application/rtf",
		
		// 压缩包
		"zip": "application/zip",
		"rar": "application/vnd.rar",
		"7z":  "application/x-7z-compressed",
		"tar": "application/x-tar",
		"gz":  "application/gzip",
		
		// 代码文件
		"html": "text/html",
		"css":  "text/css",
		"js":   "application/javascript",
		"json": "application/json",
		"xml":  "application/xml",
		"yaml": "application/x-yaml",
		"yml":  "application/x-yaml",
	}
	
	if mimeType, exists := mimeTypes[strings.ToLower(extension)]; exists {
		return mimeType
	}
	
	return "application/octet-stream"
}

// IsImageFile 判断是否为图片文件
func (PathUtils) IsImageFile(extension string) bool {
	imageExts := []string{"jpg", "jpeg", "png", "gif", "webp", "bmp", "svg"}
	ext := strings.ToLower(extension)
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}

// IsVideoFile 判断是否为视频文件
func (PathUtils) IsVideoFile(extension string) bool {
	videoExts := []string{"mp4", "avi", "mov", "wmv", "flv", "webm", "mkv", "m4v"}
	ext := strings.ToLower(extension)
	for _, vidExt := range videoExts {
		if ext == vidExt {
			return true
		}
	}
	return false
}

// IsAudioFile 判断是否为音频文件
func (PathUtils) IsAudioFile(extension string) bool {
	audioExts := []string{"mp3", "wav", "flac", "aac", "ogg", "m4a", "wma"}
	ext := strings.ToLower(extension)
	for _, audExt := range audioExts {
		if ext == audExt {
			return true
		}
	}
	return false
}

// GetFileCategory 获取文件分类
func (p PathUtils) GetFileCategory(extension string) string {
	if p.IsImageFile(extension) {
		return "image"
	}
	if p.IsVideoFile(extension) {
		return "video"
	}
	if p.IsAudioFile(extension) {
		return "audio"
	}
	
	// 文档类型
	docExts := []string{"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "txt", "rtf"}
	ext := strings.ToLower(extension)
	for _, docExt := range docExts {
		if ext == docExt {
			return "document"
		}
	}
	
	// 压缩包
	archiveExts := []string{"zip", "rar", "7z", "tar", "gz"}
	for _, archiveExt := range archiveExts {
		if ext == archiveExt {
			return "archive"
		}
	}
	
	return "other"
}

// FormatFileSize 格式化文件大小
func (PathUtils) FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// ValidateFileName 验证文件名
func (PathUtils) ValidateFileName(filename string) error {
	if filename == "" {
		return NewStorageError("InvalidFileName", "filename cannot be empty", ErrInvalidFileKey)
	}
	
	if len(filename) > 255 {
		return NewStorageError("InvalidFileName", "filename too long", ErrInvalidFileKey)
	}
	
	// 检查非法字符
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\x00"}
	for _, char := range invalidChars {
		if strings.Contains(filename, char) {
			return NewStorageError("InvalidFileName", 
				fmt.Sprintf("filename contains invalid character: %s", char), ErrInvalidFileKey)
		}
	}
	
	// 检查保留名称（Windows）
	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}
	
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
	for _, reserved := range reservedNames {
		if strings.EqualFold(nameWithoutExt, reserved) {
			return NewStorageError("InvalidFileName", 
				fmt.Sprintf("filename is reserved: %s", reserved), ErrInvalidFileKey)
		}
	}
	
	return nil
}

// StorageConfig 存储配置管理器
type StorageConfig struct {
	configs map[string]*Config
	primary string
}

// NewStorageConfig 创建存储配置管理器
func NewStorageConfig() *StorageConfig {
	return &StorageConfig{
		configs: make(map[string]*Config),
	}
}

// AddConfig 添加存储配置
func (sc *StorageConfig) AddConfig(name string, config *Config) {
	sc.configs[name] = config
	if sc.primary == "" {
		sc.primary = name
	}
}

// GetConfig 获取存储配置
func (sc *StorageConfig) GetConfig(name string) (*Config, bool) {
	config, exists := sc.configs[name]
	return config, exists
}

// GetPrimaryConfig 获取主存储配置
func (sc *StorageConfig) GetPrimaryConfig() (*Config, bool) {
	if sc.primary == "" {
		return nil, false
	}
	return sc.GetConfig(sc.primary)
}

// SetPrimary 设置主存储
func (sc *StorageConfig) SetPrimary(name string) error {
	if _, exists := sc.configs[name]; !exists {
		return NewStorageError("ConfigNotFound", "storage config not found: "+name, ErrStorageNotFound)
	}
	sc.primary = name
	return nil
}

// ListConfigs 列出所有配置
func (sc *StorageConfig) ListConfigs() []string {
	var names []string
	for name := range sc.configs {
		names = append(names, name)
	}
	return names
}

// ValidateConfig 验证存储配置
func ValidateConfig(config *Config) error {
	if config == nil {
		return NewStorageError("InvalidConfig", "config cannot be nil", ErrInvalidConfig)
	}
	
	if config.Type == "" {
		return NewStorageError("InvalidConfig", "storage type is required", ErrInvalidConfig)
	}
	
	switch config.Type {
	case StorageTypeMinIO, StorageTypeS3:
		if config.Endpoint == "" {
			return NewStorageError("InvalidConfig", "endpoint is required for MinIO/S3", ErrInvalidConfig)
		}
		if config.Bucket == "" {
			return NewStorageError("InvalidConfig", "bucket is required for MinIO/S3", ErrInvalidConfig)
		}
		if config.AccessKeyID == "" || config.SecretAccessKey == "" {
			return NewStorageError("InvalidConfig", "access credentials are required for MinIO/S3", ErrMissingCredentials)
		}
	case StorageTypeLocal:
		if config.LocalPath == "" {
			return NewStorageError("InvalidConfig", "local path is required for local storage", ErrInvalidConfig)
		}
	default:
		return NewStorageError("InvalidConfig", "unsupported storage type: "+string(config.Type), ErrUnsupportedStorageType)
	}
	
	return nil
}