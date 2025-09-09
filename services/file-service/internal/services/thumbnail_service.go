package services

import (
	"bytes"
	"context"
	"file-service/internal/models"
	"file-service/internal/repository"
	"file-service/internal/storage"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
	"time"

	"golang.org/x/image/draw"
)


// thumbnailService 缩略图服务实现
type thumbnailService struct {
	repo     *repository.Repository
	storage  storage.Storage
	utils    PathUtils
	
	// 缩略图配置
	sizes    map[string]ThumbnailSize
	quality  int
	timeout  time.Duration
}

// ThumbnailSize 缩略图尺寸配置
type ThumbnailSize struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// NewThumbnailService 创建缩略图服务实例
func NewThumbnailService(repo *repository.Repository, storage storage.Storage) ThumbnailService {
	return &thumbnailService{
		repo:    repo,
		storage: storage,
		utils:   PathUtils{},
		sizes: map[string]ThumbnailSize{
			"small":  {Name: "small", Width: 150, Height: 150},
			"medium": {Name: "medium", Width: 300, Height: 300},
			"large":  {Name: "large", Width: 600, Height: 600},
		},
		quality: 85,
		timeout: 30 * time.Second,
	}
}

// GenerateThumbnail 生成单个缩略图
func (s *thumbnailService) GenerateThumbnail(ctx context.Context, req *GenerateThumbnailRequest) (*models.Thumbnail, error) {
	// 检查文件权限
	file, err := s.repo.File.GetByID(ctx, req.FileID)
	if err != nil {
		return nil, NewServiceError("FileNotFound", "file not found", err)
	}

	if file.UserID != req.UserID {
		return nil, NewServiceError("PermissionDenied", "access denied to file", nil)
	}

	// 检查是否支持缩略图生成
	if !s.isImageFile(file.ContentType) {
		return nil, NewServiceError("UnsupportedFileType", "thumbnail generation not supported for this file type", nil)
	}

	// 检查缩略图大小配置
	sizeConfig, exists := s.sizes[req.Size]
	if !exists {
		return nil, NewServiceError("InvalidThumbnailSize", "unsupported thumbnail size", nil)
	}

	// 检查是否已存在
	existing, err := s.repo.Thumbnail.GetByFileAndSize(ctx, req.FileID, req.Size)
	if err == nil && existing != nil {
		return existing, nil // 返回已存在的缩略图
	}

	// 下载原文件
	reader, err := s.storage.Download(ctx, file.Path)
	if err != nil {
		return nil, NewServiceError("FileDownloadFailed", "failed to download source file", err)
	}
	defer reader.Close()

	// 解码图片
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, NewServiceError("ImageDecodeFailed", "failed to decode image", err)
	}

	// 生成缩略图
	thumbnail := s.resizeImage(img, sizeConfig.Width, sizeConfig.Height)

	// 编码缩略图
	thumbnailData, err := s.encodeImage(thumbnail, format)
	if err != nil {
		return nil, NewServiceError("ThumbnailEncodeFailed", "failed to encode thumbnail", err)
	}

	// 生成存储路径
	thumbnailPath := s.utils.GenerateThumbnailPath(file.UserID, file.Hash, req.Size)

	// 获取缩略图数据
	thumbnailBytes := make([]byte, thumbnailData.Len())
	thumbnailData.Read(thumbnailBytes)
	thumbnailData.Seek(0, 0) // 重置reader位置
	
	// 上传缩略图
	uploadResult, err := s.storage.Upload(ctx, thumbnailPath, thumbnailData, int64(len(thumbnailBytes)), &storage.UploadOptions{
		ContentType: s.getThumbnailContentType(format),
	})
	if err != nil {
		return nil, NewServiceError("ThumbnailUploadFailed", "failed to upload thumbnail", err)
	}

	// 创建缩略图记录
	thumbnailRecord := &models.Thumbnail{
		FileID:      req.FileID,
		Size:        req.Size,
		Width:       sizeConfig.Width,
		Height:      sizeConfig.Height,
		Path:        thumbnailPath,
		ContentType: s.getThumbnailContentType(format),
		FileSize:    int64(len(thumbnailBytes)),
		Quality:     s.quality,
		Status:      models.ThumbnailStatusReady,
		ETag:        uploadResult.ETag,
	}

	if err := s.repo.Thumbnail.Create(ctx, thumbnailRecord); err != nil {
		return nil, NewServiceError("ThumbnailCreateFailed", "failed to create thumbnail record", err)
	}

	return thumbnailRecord, nil
}

// GenerateThumbnails 为文件生成所有尺寸的缩略图
func (s *thumbnailService) GenerateThumbnails(ctx context.Context, fileID uint, userID uint) ([]*models.Thumbnail, error) {
	var thumbnails []*models.Thumbnail

	for sizeName := range s.sizes {
		req := &GenerateThumbnailRequest{
			FileID: fileID,
			Size:   sizeName,
			UserID: userID,
		}

		thumbnail, err := s.GenerateThumbnail(ctx, req)
		if err != nil {
			// 记录错误但继续生成其他尺寸
			continue
		}

		thumbnails = append(thumbnails, thumbnail)
	}

	if len(thumbnails) == 0 {
		return nil, NewServiceError("ThumbnailGenerationFailed", "failed to generate any thumbnails", nil)
	}

	return thumbnails, nil
}

// GetThumbnail 获取缩略图
func (s *thumbnailService) GetThumbnail(ctx context.Context, thumbnailID uint, userID uint) (*models.Thumbnail, error) {
	thumbnail, err := s.repo.Thumbnail.GetByID(ctx, thumbnailID)
	if err != nil {
		return nil, NewServiceError("ThumbnailNotFound", "thumbnail not found", err)
	}

	// 检查文件权限
	file, err := s.repo.File.GetByID(ctx, thumbnail.FileID)
	if err != nil {
		return nil, NewServiceError("FileNotFound", "associated file not found", err)
	}

	if file.UserID != userID {
		return nil, NewServiceError("PermissionDenied", "access denied to thumbnail", nil)
	}

	return thumbnail, nil
}

// GetFileThumbnails 获取文件的所有缩略图
func (s *thumbnailService) GetFileThumbnails(ctx context.Context, fileID uint, userID uint) ([]*models.Thumbnail, error) {
	// 检查文件权限
	file, err := s.repo.File.GetByID(ctx, fileID)
	if err != nil {
		return nil, NewServiceError("FileNotFound", "file not found", err)
	}

	if file.UserID != userID {
		return nil, NewServiceError("PermissionDenied", "access denied to file", nil)
	}

	return s.repo.Thumbnail.GetByFileID(ctx, fileID)
}

// DeleteThumbnail 删除缩略图
func (s *thumbnailService) DeleteThumbnail(ctx context.Context, thumbnailID uint, userID uint) error {
	thumbnail, err := s.GetThumbnail(ctx, thumbnailID, userID)
	if err != nil {
		return err
	}

	// 删除数据库记录
	if err := s.repo.Thumbnail.Delete(ctx, thumbnailID); err != nil {
		return NewServiceError("ThumbnailDeleteFailed", "failed to delete thumbnail record", err)
	}

	// 异步删除存储文件
	go func() {
		if err := s.storage.Delete(context.Background(), thumbnail.Path); err != nil {
			// TODO: 记录日志
		}
	}()

	return nil
}

// GetThumbnailURL 获取缩略图访问URL
func (s *thumbnailService) GetThumbnailURL(ctx context.Context, fileID uint, size string, userID uint) (string, error) {
	// 检查文件权限
	file, err := s.repo.File.GetByID(ctx, fileID)
	if err != nil {
		return "", NewServiceError("FileNotFound", "file not found", err)
	}

	if file.UserID != userID {
		return "", NewServiceError("PermissionDenied", "access denied to file", nil)
	}

	// 获取缩略图
	thumbnail, err := s.repo.Thumbnail.GetByFileAndSize(ctx, fileID, size)
	if err != nil {
		// 缩略图不存在，尝试生成
		req := &GenerateThumbnailRequest{
			FileID: fileID,
			Size:   size,
			UserID: userID,
		}
		
		thumbnail, err = s.GenerateThumbnail(ctx, req)
		if err != nil {
			return "", err
		}
	}

	// 生成预签名URL
	url, err := s.storage.GetPresignedURL(ctx, thumbnail.Path, 1*time.Hour)
	if err != nil {
		return "", NewServiceError("URLGenerationFailed", "failed to generate thumbnail URL", err)
	}

	return url, nil
}

// DownloadThumbnail 下载缩略图
func (s *thumbnailService) DownloadThumbnail(ctx context.Context, fileID uint, size string, userID uint) (io.ReadCloser, string, error) {
	// 检查文件权限
	file, err := s.repo.File.GetByID(ctx, fileID)
	if err != nil {
		return nil, "", NewServiceError("FileNotFound", "file not found", err)
	}

	if file.UserID != userID {
		return nil, "", NewServiceError("PermissionDenied", "access denied to file", nil)
	}

	// 获取缩略图
	thumbnail, err := s.repo.Thumbnail.GetByFileAndSize(ctx, fileID, size)
	if err != nil {
		// 缩略图不存在，尝试生成
		req := &GenerateThumbnailRequest{
			FileID: fileID,
			Size:   size,
			UserID: userID,
		}
		
		thumbnail, err = s.GenerateThumbnail(ctx, req)
		if err != nil {
			return nil, "", err
		}
	}

	// 下载缩略图
	reader, err := s.storage.Download(ctx, thumbnail.Path)
	if err != nil {
		return nil, "", NewServiceError("ThumbnailDownloadFailed", "failed to download thumbnail", err)
	}

	return reader, thumbnail.ContentType, nil
}

// BatchGenerateThumbnails 批量生成缩略图
func (s *thumbnailService) BatchGenerateThumbnails(ctx context.Context, fileIDs []uint, userID uint) (*BatchOperationResponse, error) {
	if len(fileIDs) == 0 {
		return &BatchOperationResponse{
			Total:   0,
			Message: "No files to process",
		}, nil
	}

	var success []uint
	var failed []BatchError

	for _, fileID := range fileIDs {
		thumbnails, err := s.GenerateThumbnails(ctx, fileID, userID)
		if err != nil {
			failed = append(failed, BatchError{
				ID:      fileID,
				Message: err.Error(),
			})
		} else if len(thumbnails) > 0 {
			success = append(success, fileID)
		}
	}

	message := "Batch thumbnail generation completed"
	if len(failed) > 0 {
		message += " with some errors"
	}

	return &BatchOperationResponse{
		Success: success,
		Failed:  failed,
		Total:   len(fileIDs),
		Message: message,
	}, nil
}

// CleanupOrphanedThumbnails 清理孤立缩略图
func (s *thumbnailService) CleanupOrphanedThumbnails(ctx context.Context) (int64, error) {
	return s.repo.Thumbnail.CleanupOrphaned(ctx)
}

// RefreshThumbnails 刷新缩略图
func (s *thumbnailService) RefreshThumbnails(ctx context.Context, fileID uint, userID uint) ([]*models.Thumbnail, error) {
	// 删除现有缩略图
	existingThumbnails, err := s.GetFileThumbnails(ctx, fileID, userID)
	if err != nil {
		return nil, err
	}

	for _, thumbnail := range existingThumbnails {
		s.DeleteThumbnail(ctx, thumbnail.ID, userID)
	}

	// 重新生成缩略图
	return s.GenerateThumbnails(ctx, fileID, userID)
}

// GetThumbnailStats 获取缩略图统计
func (s *thumbnailService) GetThumbnailStats(ctx context.Context, userID uint) (*ThumbnailStats, error) {
	// 获取用户文件数量
	totalFiles, err := s.repo.File.GetUserFileCount(ctx, userID)
	if err != nil {
		return nil, NewServiceError("StatsRetrieveFailed", "failed to get file count", err)
	}

	// 获取缩略图统计
	thumbnailStats, err := s.repo.Thumbnail.GetUserStats(ctx, userID)
	if err != nil {
		return nil, NewServiceError("StatsRetrieveFailed", "failed to get thumbnail stats", err)
	}

	var coveragePercent float64
	if totalFiles > 0 {
		filesWithThumbnails := thumbnailStats.FilesWithThumbnails
		coveragePercent = float64(filesWithThumbnails) / float64(totalFiles) * 100
	}

	return &ThumbnailStats{
		TotalThumbnails:     thumbnailStats.TotalThumbnails,
		FilesWithThumbnails: thumbnailStats.FilesWithThumbnails,
		TotalSize:          thumbnailStats.TotalSize,
		CoveragePercent:    coveragePercent,
		BySize:             thumbnailStats.BySize,
	}, nil
}

// 辅助方法

// isImageFile 检查是否为图片文件
func (s *thumbnailService) isImageFile(contentType string) bool {
	imageTypes := []string{
		"image/jpeg", "image/jpg", "image/png", "image/gif",
		"image/webp", "image/bmp", "image/tiff",
	}

	for _, imageType := range imageTypes {
		if strings.EqualFold(contentType, imageType) {
			return true
		}
	}

	return false
}

// resizeImage 调整图片大小
func (s *thumbnailService) resizeImage(src image.Image, width, height int) image.Image {
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	// 计算缩放比例，保持宽高比
	scaleX := float64(width) / float64(srcWidth)
	scaleY := float64(height) / float64(srcHeight)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// 计算新尺寸
	newWidth := int(float64(srcWidth) * scale)
	newHeight := int(float64(srcHeight) * scale)

	// 创建目标图片
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// 使用高质量缩放算法
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, srcBounds, draw.Over, nil)

	return dst
}

// encodeImage 编码图片
func (s *thumbnailService) encodeImage(img image.Image, format string) (*bytes.Reader, error) {
	var buf bytes.Buffer

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: s.quality})
		if err != nil {
			return nil, err
		}
	case "png":
		err := png.Encode(&buf, img)
		if err != nil {
			return nil, err
		}
	case "gif":
		err := gif.Encode(&buf, img, nil)
		if err != nil {
			return nil, err
		}
	default:
		// 默认使用JPEG
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: s.quality})
		if err != nil {
			return nil, err
		}
	}

	return bytes.NewReader(buf.Bytes()), nil
}

// getThumbnailContentType 获取缩略图内容类型
func (s *thumbnailService) getThumbnailContentType(format string) string {
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	default:
		return "image/jpeg"
	}
}

