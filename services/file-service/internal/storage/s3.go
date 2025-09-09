package storage

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// s3Storage AWS S3存储实现
type s3Storage struct {
	client   *s3.Client
	bucket   string
	config   *Config
	uploader *manager.Uploader
}

// NewS3Storage 创建S3存储实例
func NewS3Storage(cfg *Config) (Storage, error) {
	if cfg.Bucket == "" {
		return nil, NewStorageError("InvalidConfig", "bucket name is required", ErrInvalidBucket)
	}
	
	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, NewStorageError("InvalidConfig", "access credentials are required", ErrMissingCredentials)
	}
	
	// 创建AWS配置
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, NewStorageError("ConfigLoadFailed", "failed to load AWS config", err)
	}
	
	// 如果有自定义endpoint，创建自定义解析器
	if cfg.Endpoint != "" {
		awsCfg.BaseEndpoint = aws.String(cfg.Endpoint)
	}
	
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.UsePathStyle = true // 强制使用路径风格
		}
	})
	
	uploader := manager.NewUploader(client)
	
	return &s3Storage{
		client:   client,
		bucket:   cfg.Bucket,
		config:   cfg,
		uploader: uploader,
	}, nil
}

// Upload 上传文件
func (s *s3Storage) Upload(ctx context.Context, key string, reader io.Reader, size int64, options *UploadOptions) (*UploadResult, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   reader,
	}
	
	if options != nil {
		if options.ContentType != "" {
			input.ContentType = aws.String(options.ContentType)
		}
		if options.ContentEncoding != "" {
			input.ContentEncoding = aws.String(options.ContentEncoding)
		}
		if options.CacheControl != "" {
			input.CacheControl = aws.String(options.CacheControl)
		}
		if len(options.Metadata) > 0 {
			input.Metadata = options.Metadata
		}
		if options.StorageClass != "" {
			input.StorageClass = types.StorageClass(options.StorageClass)
		}
		if options.ServerSideEncryption {
			input.ServerSideEncryption = types.ServerSideEncryptionAes256
		}
	}
	
	result, err := s.uploader.Upload(ctx, input)
	if err != nil {
		return nil, s.convertError(err, key)
	}
	
	return &UploadResult{
		Key:      key,
		ETag:     strings.Trim(*result.ETag, "\""),
		Location: result.Location,
		Size:     size, // S3 Upload不返回size，使用传入的值
	}, nil
}

// Download 下载文件
func (s *s3Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	
	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, s.convertError(err, key)
	}
	
	return result.Body, nil
}

// Delete 删除文件
func (s *s3Storage) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	
	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return s.convertError(err, key)
	}
	
	return nil
}

// Exists 检查文件是否存在
func (s *s3Storage) Exists(ctx context.Context, key string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	
	_, err := s.client.HeadObject(ctx, input)
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			if ae.ErrorCode() == "NotFound" || ae.ErrorCode() == "NoSuchKey" {
				return false, nil
			}
		}
		return false, s.convertError(err, key)
	}
	
	return true, nil
}

// GetSize 获取文件大小
func (s *s3Storage) GetSize(ctx context.Context, key string) (int64, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	
	result, err := s.client.HeadObject(ctx, input)
	if err != nil {
		return 0, s.convertError(err, key)
	}
	
	return *result.ContentLength, nil
}

// GetInfo 获取文件信息
func (s *s3Storage) GetInfo(ctx context.Context, key string) (*FileInfo, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	
	result, err := s.client.HeadObject(ctx, input)
	if err != nil {
		return nil, s.convertError(err, key)
	}
	
	info := &FileInfo{
		Key:          key,
		Size:         *result.ContentLength,
		ETag:         strings.Trim(*result.ETag, "\""),
		LastModified: *result.LastModified,
		Metadata:     result.Metadata,
	}
	
	if result.ContentType != nil {
		info.ContentType = *result.ContentType
	}
	if result.StorageClass != "" {
		info.StorageClass = string(result.StorageClass)
	}
	
	return info, nil
}

// GetURL 获取预签名URL
func (s *s3Storage) GetURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	presigner := s3.NewPresignClient(s.client)
	
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	
	result, err := presigner.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expires
	})
	if err != nil {
		return "", s.convertError(err, key)
	}
	
	return result.URL, nil
}

// GetPresignedURL 获取预签名URL（别名）
func (s *s3Storage) GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	return s.GetURL(ctx, key, expires)
}

// InitiateMultipartUpload 初始化分片上传
func (s *s3Storage) InitiateMultipartUpload(ctx context.Context, key string, options *MultipartUploadOptions) (*MultipartUploadResult, error) {
	input := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	
	if options != nil {
		if options.ContentType != "" {
			input.ContentType = aws.String(options.ContentType)
		}
		if len(options.Metadata) > 0 {
			input.Metadata = options.Metadata
		}
		if options.StorageClass != "" {
			input.StorageClass = types.StorageClass(options.StorageClass)
		}
	}
	
	result, err := s.client.CreateMultipartUpload(ctx, input)
	if err != nil {
		return nil, s.convertError(err, key)
	}
	
	return &MultipartUploadResult{
		UploadID: *result.UploadId,
		Key:      key,
	}, nil
}

// UploadPart 上传分片
func (s *s3Storage) UploadPart(ctx context.Context, uploadID, key string, partNumber int, reader io.Reader, size int64) (*UploadPartResult, error) {
	input := &s3.UploadPartInput{
		Bucket:     aws.String(s.bucket),
		Key:        aws.String(key),
		PartNumber: aws.Int32(int32(partNumber)),
		UploadId:   aws.String(uploadID),
		Body:       reader,
	}
	
	result, err := s.client.UploadPart(ctx, input)
	if err != nil {
		return nil, s.convertError(err, key)
	}
	
	return &UploadPartResult{
		PartNumber: partNumber,
		ETag:       strings.Trim(*result.ETag, "\""),
		Size:       size,
	}, nil
}

// CompleteMultipartUpload 完成分片上传
func (s *s3Storage) CompleteMultipartUpload(ctx context.Context, uploadID, key string, parts []UploadPartInfo) (*CompleteMultipartUploadResult, error) {
	completedParts := make([]types.CompletedPart, len(parts))
	for i, part := range parts {
		completedParts[i] = types.CompletedPart{
			PartNumber: aws.Int32(int32(part.PartNumber)),
			ETag:       aws.String(part.ETag),
		}
	}
	
	input := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}
	
	result, err := s.client.CompleteMultipartUpload(ctx, input)
	if err != nil {
		return nil, s.convertError(err, key)
	}
	
	// 计算总大小
	var totalSize int64
	for _, part := range parts {
		totalSize += part.Size
	}
	
	return &CompleteMultipartUploadResult{
		Key:      key,
		ETag:     strings.Trim(*result.ETag, "\""),
		Location: *result.Location,
		Size:     totalSize,
	}, nil
}

// AbortMultipartUpload 中止分片上传
func (s *s3Storage) AbortMultipartUpload(ctx context.Context, uploadID, key string) error {
	input := &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	}
	
	_, err := s.client.AbortMultipartUpload(ctx, input)
	if err != nil {
		return s.convertError(err, key)
	}
	
	return nil
}

// ListParts 列举已上传的分片
func (s *s3Storage) ListParts(ctx context.Context, uploadID, key string) ([]PartInfo, error) {
	input := &s3.ListPartsInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	}
	
	result, err := s.client.ListParts(ctx, input)
	if err != nil {
		return nil, s.convertError(err, key)
	}
	
	parts := make([]PartInfo, len(result.Parts))
	for i, part := range result.Parts {
		parts[i] = PartInfo{
			PartNumber:   int(*part.PartNumber),
			ETag:         strings.Trim(*part.ETag, "\""),
			Size:         *part.Size,
			LastModified: *part.LastModified,
		}
	}
	
	return parts, nil
}

// ListObjects 列举对象
func (s *s3Storage) ListObjects(ctx context.Context, prefix string, options *ListOptions) (*ListResult, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	}
	
	if options != nil {
		if options.MaxKeys > 0 {
			input.MaxKeys = aws.Int32(int32(options.MaxKeys))
		}
		if options.Delimiter != "" {
			input.Delimiter = aws.String(options.Delimiter)
		}
		if options.ContinuationToken != "" {
			input.ContinuationToken = aws.String(options.ContinuationToken)
		}
		if options.StartAfter != "" {
			input.StartAfter = aws.String(options.StartAfter)
		}
	}
	
	result, err := s.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, s.convertError(err, prefix)
	}
	
	objects := make([]ObjectInfo, len(result.Contents))
	for i, obj := range result.Contents {
		objects[i] = ObjectInfo{
			Key:          *obj.Key,
			Size:         *obj.Size,
			ETag:         strings.Trim(*obj.ETag, "\""),
			LastModified: *obj.LastModified,
		}
		if obj.StorageClass != "" {
			objects[i].StorageClass = string(obj.StorageClass)
		}
	}
	
	var commonPrefixes []string
	for _, cp := range result.CommonPrefixes {
		commonPrefixes = append(commonPrefixes, *cp.Prefix)
	}
	
	listResult := &ListResult{
		Objects:        objects,
		CommonPrefixes: commonPrefixes,
		IsTruncated:    result.IsTruncated != nil && *result.IsTruncated,
	}
	
	if result.NextContinuationToken != nil {
		listResult.NextContinuationToken = *result.NextContinuationToken
	}
	
	return listResult, nil
}

// CopyObject 复制对象
func (s *s3Storage) CopyObject(ctx context.Context, srcKey, destKey string, options *CopyOptions) error {
	copySource := s.bucket + "/" + srcKey
	
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		Key:        aws.String(destKey),
		CopySource: aws.String(copySource),
	}
	
	if options != nil {
		if options.MetadataDirective != "" {
			input.MetadataDirective = types.MetadataDirective(options.MetadataDirective)
		}
		if len(options.Metadata) > 0 {
			input.Metadata = options.Metadata
		}
		if options.StorageClass != "" {
			input.StorageClass = types.StorageClass(options.StorageClass)
		}
	}
	
	_, err := s.client.CopyObject(ctx, input)
	if err != nil {
		return s.convertError(err, srcKey)
	}
	
	return nil
}

// MoveObject 移动对象
func (s *s3Storage) MoveObject(ctx context.Context, srcKey, destKey string) error {
	// 先复制后删除
	err := s.CopyObject(ctx, srcKey, destKey, nil)
	if err != nil {
		return err
	}
	
	return s.Delete(ctx, srcKey)
}

// BatchDelete 批量删除对象
func (s *s3Storage) BatchDelete(ctx context.Context, keys []string) (*BatchDeleteResult, error) {
	objects := make([]types.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = types.ObjectIdentifier{
			Key: aws.String(key),
		}
	}
	
	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(s.bucket),
		Delete: &types.Delete{
			Objects: objects,
			Quiet:   aws.Bool(false),
		},
	}
	
	result, err := s.client.DeleteObjects(ctx, input)
	if err != nil {
		return nil, s.convertError(err, "")
	}
	
	var deleted []string
	for _, obj := range result.Deleted {
		deleted = append(deleted, *obj.Key)
	}
	
	var errors []DeleteError
	for _, err := range result.Errors {
		errors = append(errors, DeleteError{
			Key:     *err.Key,
			Code:    *err.Code,
			Message: *err.Message,
		})
	}
	
	return &BatchDeleteResult{
		Deleted: deleted,
		Errors:  errors,
	}, nil
}

// HealthCheck 健康检查
func (s *s3Storage) HealthCheck(ctx context.Context) error {
	input := &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	}
	
	_, err := s.client.HeadBucket(ctx, input)
	if err != nil {
		return NewStorageError("HealthCheckFailed", "failed to check bucket", err)
	}
	
	return nil
}

// convertError 转换AWS S3错误为存储错误
func (s *s3Storage) convertError(err error, key string) error {
	if err == nil {
		return nil
	}
	
	var ae smithy.APIError
	if errors.As(err, &ae) {
		switch ae.ErrorCode() {
		case "NoSuchKey", "NotFound":
			return NewFileNotFoundError(key)
		case "AccessDenied", "Forbidden":
			return NewAccessDeniedError(key, ae.ErrorMessage())
		case "NoSuchUpload", "InvalidUploadId":
			return NewInvalidUploadIDError(key)
		case "EntityTooSmall":
			return NewPartTooSmallError(0, 0)
		default:
			return NewStorageError("S3Error", ae.ErrorMessage(), err)
		}
	}
	
	return NewStorageError("S3Error", err.Error(), err)
}