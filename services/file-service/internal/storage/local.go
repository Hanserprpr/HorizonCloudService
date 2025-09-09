package storage

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// localStorage 本地文件系统存储实现
type localStorage struct {
	basePath string
	config   *Config
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(config *Config) (Storage, error) {
	if config.LocalPath == "" {
		return nil, NewStorageError("InvalidConfig", "local path is required", ErrInvalidConfig)
	}
	
	// 确保基础路径存在
	err := os.MkdirAll(config.LocalPath, 0755)
	if err != nil {
		return nil, NewStorageError("PathCreateFailed", "failed to create base path", err)
	}
	
	return &localStorage{
		basePath: config.LocalPath,
		config:   config,
	}, nil
}

// getFullPath 获取完整文件路径
func (s *localStorage) getFullPath(key string) string {
	// 清理路径，防止路径遍历攻击
	cleanKey := filepath.Clean(key)
	if strings.HasPrefix(cleanKey, "/") {
		cleanKey = cleanKey[1:]
	}
	return filepath.Join(s.basePath, cleanKey)
}

// Upload 上传文件
func (s *localStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64, options *UploadOptions) (*UploadResult, error) {
	fullPath := s.getFullPath(key)
	
	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, NewStorageError("DirectoryCreateFailed", "failed to create directory", err)
	}
	
	// 创建文件
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, NewStorageError("FileCreateFailed", "failed to create file", err)
	}
	defer file.Close()
	
	// 计算MD5哈希
	hash := md5.New()
	writer := io.MultiWriter(file, hash)
	
	// 复制数据
	written, err := io.Copy(writer, reader)
	if err != nil {
		os.Remove(fullPath) // 清理失败的文件
		return nil, NewStorageError("WriteError", "failed to write file", err)
	}
	
	etag := hex.EncodeToString(hash.Sum(nil))
	
	return &UploadResult{
		Key:  key,
		ETag: etag,
		Size: written,
	}, nil
}

// Download 下载文件
func (s *localStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := s.getFullPath(key)
	
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewFileNotFoundError(key)
		}
		return nil, NewStorageError("FileOpenFailed", "failed to open file", err)
	}
	
	return file, nil
}

// Delete 删除文件
func (s *localStorage) Delete(ctx context.Context, key string) error {
	fullPath := s.getFullPath(key)
	
	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return NewStorageError("DeleteFailed", "failed to delete file", err)
	}
	
	return nil
}

// Exists 检查文件是否存在
func (s *localStorage) Exists(ctx context.Context, key string) (bool, error) {
	fullPath := s.getFullPath(key)
	
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, NewStorageError("StatFailed", "failed to stat file", err)
	}
	
	return true, nil
}

// GetSize 获取文件大小
func (s *localStorage) GetSize(ctx context.Context, key string) (int64, error) {
	fullPath := s.getFullPath(key)
	
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, NewFileNotFoundError(key)
		}
		return 0, NewStorageError("StatFailed", "failed to stat file", err)
	}
	
	return info.Size(), nil
}

// GetInfo 获取文件信息
func (s *localStorage) GetInfo(ctx context.Context, key string) (*FileInfo, error) {
	fullPath := s.getFullPath(key)
	
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewFileNotFoundError(key)
		}
		return nil, NewStorageError("StatFailed", "failed to stat file", err)
	}
	
	// 计算ETag（文件的MD5）
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, NewStorageError("FileOpenFailed", "failed to open file for hash", err)
	}
	defer file.Close()
	
	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return nil, NewStorageError("HashError", "failed to calculate file hash", err)
	}
	
	etag := hex.EncodeToString(hash.Sum(nil))
	
	return &FileInfo{
		Key:          key,
		Size:         info.Size(),
		ETag:         etag,
		LastModified: info.ModTime(),
		ContentType:  "application/octet-stream", // 本地存储默认类型
	}, nil
}

// GetURL 获取文件URL（本地存储不支持预签名URL）
func (s *localStorage) GetURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	return "", NewStorageError("NotSupported", "local storage does not support presigned URLs", nil)
}

// GetPresignedURL 获取预签名URL（本地存储不支持）
func (s *localStorage) GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	return "", NewStorageError("NotSupported", "local storage does not support presigned URLs", nil)
}

// InitiateMultipartUpload 初始化分片上传（本地存储模拟实现）
func (s *localStorage) InitiateMultipartUpload(ctx context.Context, key string, options *MultipartUploadOptions) (*MultipartUploadResult, error) {
	// 生成上传ID
	uploadID := fmt.Sprintf("%d_%s", time.Now().UnixNano(), key)
	
	// 创建临时目录存储分片
	tempDir := filepath.Join(s.basePath, ".multipart", uploadID)
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		return nil, NewStorageError("TempDirCreateFailed", "failed to create temp directory", err)
	}
	
	return &MultipartUploadResult{
		UploadID: uploadID,
		Key:      key,
	}, nil
}

// UploadPart 上传分片
func (s *localStorage) UploadPart(ctx context.Context, uploadID, key string, partNumber int, reader io.Reader, size int64) (*UploadPartResult, error) {
	// 获取分片文件路径
	tempDir := filepath.Join(s.basePath, ".multipart", uploadID)
	partPath := filepath.Join(tempDir, fmt.Sprintf("part_%d", partNumber))
	
	// 创建分片文件
	file, err := os.Create(partPath)
	if err != nil {
		return nil, NewStorageError("PartCreateFailed", "failed to create part file", err)
	}
	defer file.Close()
	
	// 计算分片MD5
	hash := md5.New()
	writer := io.MultiWriter(file, hash)
	
	written, err := io.Copy(writer, reader)
	if err != nil {
		os.Remove(partPath)
		return nil, NewStorageError("PartWriteFailed", "failed to write part", err)
	}
	
	etag := hex.EncodeToString(hash.Sum(nil))
	
	return &UploadPartResult{
		PartNumber: partNumber,
		ETag:       etag,
		Size:       written,
	}, nil
}

// CompleteMultipartUpload 完成分片上传
func (s *localStorage) CompleteMultipartUpload(ctx context.Context, uploadID, key string, parts []UploadPartInfo) (*CompleteMultipartUploadResult, error) {
	tempDir := filepath.Join(s.basePath, ".multipart", uploadID)
	fullPath := s.getFullPath(key)
	
	// 确保目标目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, NewStorageError("DirectoryCreateFailed", "failed to create target directory", err)
	}
	
	// 创建最终文件
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, NewStorageError("FileCreateFailed", "failed to create final file", err)
	}
	defer file.Close()
	
	// 按顺序合并分片
	hash := md5.New()
	writer := io.MultiWriter(file, hash)
	var totalSize int64
	
	for _, part := range parts {
		partPath := filepath.Join(tempDir, fmt.Sprintf("part_%d", part.PartNumber))
		
		partFile, err := os.Open(partPath)
		if err != nil {
			os.Remove(fullPath) // 清理失败的文件
			return nil, NewStorageError("PartReadFailed", fmt.Sprintf("failed to read part %d", part.PartNumber), err)
		}
		
		written, err := io.Copy(writer, partFile)
		partFile.Close()
		
		if err != nil {
			os.Remove(fullPath)
			return nil, NewStorageError("PartMergeFailed", fmt.Sprintf("failed to merge part %d", part.PartNumber), err)
		}
		
		totalSize += written
	}
	
	// 清理临时文件
	os.RemoveAll(tempDir)
	
	etag := hex.EncodeToString(hash.Sum(nil))
	
	return &CompleteMultipartUploadResult{
		Key:  key,
		ETag: etag,
		Size: totalSize,
	}, nil
}

// AbortMultipartUpload 中止分片上传
func (s *localStorage) AbortMultipartUpload(ctx context.Context, uploadID, key string) error {
	tempDir := filepath.Join(s.basePath, ".multipart", uploadID)
	return os.RemoveAll(tempDir)
}

// ListParts 列举已上传的分片
func (s *localStorage) ListParts(ctx context.Context, uploadID, key string) ([]PartInfo, error) {
	tempDir := filepath.Join(s.basePath, ".multipart", uploadID)
	
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewInvalidUploadIDError(uploadID)
		}
		return nil, NewStorageError("ListPartsFailed", "failed to list parts", err)
	}
	
	var parts []PartInfo
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "part_") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			
			var partNumber int
			fmt.Sscanf(entry.Name(), "part_%d", &partNumber)
			
			parts = append(parts, PartInfo{
				PartNumber:   partNumber,
				Size:         info.Size(),
				LastModified: info.ModTime(),
			})
		}
	}
	
	return parts, nil
}

// ListObjects 列举对象
func (s *localStorage) ListObjects(ctx context.Context, prefix string, options *ListOptions) (*ListResult, error) {
	var objects []ObjectInfo
	maxKeys := 1000
	
	if options != nil && options.MaxKeys > 0 {
		maxKeys = options.MaxKeys
	}
	
	prefixPath := s.getFullPath(prefix)
	
	err := filepath.WalkDir(s.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if d.IsDir() {
			return nil
		}
		
		// 跳过临时文件
		if strings.Contains(path, ".multipart") {
			return nil
		}
		
		// 检查前缀匹配
		if prefix != "" && !strings.HasPrefix(path, prefixPath) {
			return nil
		}
		
		// 转换为相对路径
		relPath, err := filepath.Rel(s.basePath, path)
		if err != nil {
			return err
		}
		
		// 统一路径分隔符
		key := filepath.ToSlash(relPath)
		
		info, err := d.Info()
		if err != nil {
			return err
		}
		
		objects = append(objects, ObjectInfo{
			Key:          key,
			Size:         info.Size(),
			LastModified: info.ModTime(),
		})
		
		// 限制返回数量
		if len(objects) >= maxKeys {
			return filepath.SkipAll
		}
		
		return nil
	})
	
	if err != nil {
		return nil, NewStorageError("ListFailed", "failed to list objects", err)
	}
	
	return &ListResult{
		Objects:     objects,
		IsTruncated: len(objects) >= maxKeys,
	}, nil
}

// CopyObject 复制对象
func (s *localStorage) CopyObject(ctx context.Context, srcKey, destKey string, options *CopyOptions) error {
	srcPath := s.getFullPath(srcKey)
	destPath := s.getFullPath(destKey)
	
	// 确保目标目录存在
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return NewStorageError("DirectoryCreateFailed", "failed to create destination directory", err)
	}
	
	// 复制文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewFileNotFoundError(srcKey)
		}
		return NewStorageError("SourceReadFailed", "failed to open source file", err)
	}
	defer srcFile.Close()
	
	destFile, err := os.Create(destPath)
	if err != nil {
		return NewStorageError("DestCreateFailed", "failed to create destination file", err)
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		os.Remove(destPath) // 清理失败的文件
		return NewStorageError("CopyFailed", "failed to copy file", err)
	}
	
	return nil
}

// MoveObject 移动对象
func (s *localStorage) MoveObject(ctx context.Context, srcKey, destKey string) error {
	srcPath := s.getFullPath(srcKey)
	destPath := s.getFullPath(destKey)
	
	// 确保目标目录存在
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return NewStorageError("DirectoryCreateFailed", "failed to create destination directory", err)
	}
	
	err := os.Rename(srcPath, destPath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewFileNotFoundError(srcKey)
		}
		return NewStorageError("MoveFailed", "failed to move file", err)
	}
	
	return nil
}

// BatchDelete 批量删除对象
func (s *localStorage) BatchDelete(ctx context.Context, keys []string) (*BatchDeleteResult, error) {
	var deleted []string
	var errors []DeleteError
	
	for _, key := range keys {
		err := s.Delete(ctx, key)
		if err != nil {
			errors = append(errors, DeleteError{
				Key:     key,
				Code:    "DeleteFailed",
				Message: err.Error(),
			})
		} else {
			deleted = append(deleted, key)
		}
	}
	
	return &BatchDeleteResult{
		Deleted: deleted,
		Errors:  errors,
	}, nil
}

// HealthCheck 健康检查
func (s *localStorage) HealthCheck(ctx context.Context) error {
	// 检查基础路径是否可访问
	info, err := os.Stat(s.basePath)
	if err != nil {
		return NewStorageError("HealthCheckFailed", "base path not accessible", err)
	}
	
	if !info.IsDir() {
		return NewStorageError("HealthCheckFailed", "base path is not a directory", nil)
	}
	
	// 检查读写权限
	testFile := filepath.Join(s.basePath, ".health_check_test")
	file, err := os.Create(testFile)
	if err != nil {
		return NewStorageError("HealthCheckFailed", "cannot write to base path", err)
	}
	file.Close()
	os.Remove(testFile)
	
	return nil
}

// GetRootPath 获取根路径 (用于测试清理)
func (s *localStorage) GetRootPath() string {
	return s.basePath
}