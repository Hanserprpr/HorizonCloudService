package services

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"file-service/internal/models"
	"file-service/internal/repository"
	"file-service/internal/storage"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

// fileService 文件服务实现
type fileService struct {
	repo    *repository.Repository
	storage storage.Storage
	utils   PathUtils
}

// NewFileService 创建文件服务实例
func NewFileService(repo *repository.Repository, storage storage.Storage) FileService {
	return &fileService{
		repo:    repo,
		storage: storage,
		utils:   PathUtils{},
	}
}

// UploadFile 上传文件
func (s *fileService) UploadFile(ctx context.Context, req *UploadFileRequest) (*UploadFileResponse, error) {
	// 验证文件名
	if err := s.utils.ValidateFileName(req.FileName); err != nil {
		return nil, err
	}
	
	// 计算文件哈希
	hash, reader, err := s.calculateFileHash(req.Reader)
	if err != nil {
		return nil, NewServiceError("HashCalculationFailed", "failed to calculate file hash", err)
	}
	
	// 检查文件是否已存在（去重）
	existingFile, err := s.repo.File.GetByHash(ctx, hash)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, NewServiceError("DuplicateCheckFailed", "failed to check file duplicates", err)
	}
	
	// 如果文件已存在，创建引用
	if existingFile != nil && existingFile.UserID == req.UserID {
		duplicatedFile := &models.File{
			Name:         req.FileName,
			OriginalName: req.FileName,
			Hash:         existingFile.Hash,
			Size:         existingFile.Size,
			ContentType:  existingFile.ContentType,
			Extension:    existingFile.Extension,
			Path:         existingFile.Path, // 共享存储路径
			FolderID:     req.FolderID,
			UserID:       req.UserID,
			Category:     existingFile.Category,
			StorageTier:  existingFile.StorageTier,
			Tags:         req.Tags,
			Description:  req.Description,
			Status:       models.FileStatusActive,
		}
		
		if err := s.repo.File.Create(ctx, duplicatedFile); err != nil {
			return nil, NewServiceError("FileCreateFailed", "failed to create file record", err)
		}
		
		return &UploadFileResponse{
			File:       duplicatedFile,
			Message:    "File uploaded successfully (deduplicated)",
			Duplicated: true,
		}, nil
	}
	
	// 生成存储路径
	extension := s.utils.ExtractExtension(req.FileName)
	storagePath := s.utils.GenerateStoragePath(req.UserID, hash, extension)
	
	// 上传到存储
	uploadOpts := &storage.UploadOptions{
		ContentType: req.ContentType,
		Metadata:    req.Metadata,
		StorageClass: "STANDARD",
	}
	
	uploadResult, err := s.storage.Upload(ctx, storagePath, reader, req.Size, uploadOpts)
	if err != nil {
		return nil, NewServiceError("StorageUploadFailed", "failed to upload file to storage", err)
	}
	
	// 创建文件记录
	file := &models.File{
		Name:         req.FileName,
		OriginalName: req.FileName,
		Hash:         hash,
		Size:         uploadResult.Size,
		ContentType:  s.getContentType(req.ContentType, extension),
		Extension:    extension,
		Path:         storagePath,
		FolderID:     req.FolderID,
		UserID:       req.UserID,
		Category:     s.utils.GetFileCategory(extension),
		StorageTier:  "hot",
		Tags:         req.Tags,
		Description:  req.Description,
		Status:       models.FileStatusActive,
		Metadata:     s.extractFileMetadata(req),
	}
	
	if err := s.repo.File.Create(ctx, file); err != nil {
		// 如果数据库操作失败，清理存储文件
		s.storage.Delete(ctx, storagePath)
		return nil, NewServiceError("FileCreateFailed", "failed to create file record", err)
	}
	
	return &UploadFileResponse{
		File:    file,
		Message: "File uploaded successfully",
	}, nil
}

// CompleteUpload 完成上传
func (s *fileService) CompleteUpload(ctx context.Context, req *CompleteUploadRequest) (*CompleteUploadResponse, error) {
	// 获取上传会话
	session, err := s.repo.Upload.GetSession(ctx, req.SessionID)
	if err != nil {
		return nil, NewServiceError("SessionNotFound", "upload session not found", err)
	}
	
	if session.UserID != req.UserID {
		return nil, NewServiceError("PermissionDenied", "access denied to upload session", nil)
	}
	
	if session.Status != models.UploadStatusUploading {
		return nil, NewServiceError("InvalidSessionStatus", "upload session is not in uploading status", nil)
	}
	
	// 检查所有分片是否完成
	chunks, err := s.repo.Upload.GetSessionChunks(ctx, req.SessionID)
	if err != nil {
		return nil, NewServiceError("ChunkRetrieveFailed", "failed to retrieve upload chunks", err)
	}
	
	completedChunks := 0
	for _, chunk := range chunks {
		if chunk.Status == models.ChunkStatusCompleted {
			completedChunks++
		}
	}
	
	if completedChunks != session.TotalChunks {
		return nil, NewServiceError("IncompleteUpload", "not all chunks have been uploaded", nil)
	}
	
	// 合并分片（如果是分片上传）
	extension := s.utils.ExtractExtension(session.FileName)
	storagePath := s.utils.GenerateStoragePath(session.UserID, session.Hash, extension)
	
	if session.TotalChunks > 1 {
		// 完成分片上传
		parts := make([]storage.UploadPartInfo, len(chunks))
		for i, chunk := range chunks {
			parts[i] = storage.UploadPartInfo{
				PartNumber: chunk.ChunkIndex + 1, // 存储层通常从1开始
				ETag:       chunk.ETag,
				Size:       chunk.ChunkSize,
			}
		}
		
		_, err = s.storage.CompleteMultipartUpload(ctx, session.UploadID, storagePath, parts)
		if err != nil {
			return nil, NewServiceError("MultipartCompleteFailed", "failed to complete multipart upload", err)
		}
	}
	
	// 创建文件记录
	file := &models.File{
		Name:         session.FileName,
		OriginalName: session.FileName,
		Hash:         session.Hash,
		Size:         session.FileSize,
		ContentType:  session.ContentType,
		Extension:    extension,
		Path:         storagePath,
		FolderID:     session.FolderID,
		UserID:       session.UserID,
		Category:     s.utils.GetFileCategory(extension),
		StorageTier:  "hot",
		Status:       models.FileStatusActive,
	}
	
	err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		// 创建文件记录
		if err := s.repo.WithTx(tx).File.Create(ctx, file); err != nil {
			return err
		}
		
		// 更新上传会话状态
		return s.repo.WithTx(tx).Upload.CompleteSession(ctx, req.SessionID, file.ID)
	})
	
	if err != nil {
		return nil, NewServiceError("FileCreateFailed", "failed to create file record", err)
	}
	
	return &CompleteUploadResponse{
		File:    file,
		Message: "Upload completed successfully",
	}, nil
}

// GetFile 获取文件
func (s *fileService) GetFile(ctx context.Context, fileID uint, userID uint) (*models.File, error) {
	file, err := s.repo.File.GetByID(ctx, fileID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewServiceError("FileNotFound", "file not found", err)
		}
		return nil, NewServiceError("FileRetrieveFailed", "failed to retrieve file", err)
	}
	
	if file.UserID != userID {
		return nil, NewServiceError("PermissionDenied", "access denied to file", nil)
	}
	
	return file, nil
}

// GetFileByPath 根据路径获取文件
func (s *fileService) GetFileByPath(ctx context.Context, path string, userID uint) (*models.File, error) {
	// 这里简化实现，实际应该根据用户和路径查询
	file, err := s.repo.File.GetByPath(ctx, path)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewServiceError("FileNotFound", "file not found", err)
		}
		return nil, NewServiceError("FileRetrieveFailed", "failed to retrieve file", err)
	}
	
	if file.UserID != userID {
		return nil, NewServiceError("PermissionDenied", "access denied to file", nil)
	}
	
	return file, nil
}

// ListFiles 列举文件
func (s *fileService) ListFiles(ctx context.Context, req *ListFilesRequest) (*ListFilesResponse, error) {
	files, total, err := s.repo.File.List(ctx, req.UserID, req.FolderID, req.Offset, req.Limit, req.Filters)
	if err != nil {
		return nil, NewServiceError("FileListFailed", "failed to list files", err)
	}
	
	return &ListFilesResponse{
		Files: files,
		Total: total,
	}, nil
}

// SearchFiles 搜索文件
func (s *fileService) SearchFiles(ctx context.Context, req *SearchFilesRequest) (*SearchFilesResponse, error) {
	files, total, err := s.repo.File.Search(ctx, req.UserID, req.Keyword, req.Offset, req.Limit, req.Filters)
	if err != nil {
		return nil, NewServiceError("FileSearchFailed", "failed to search files", err)
	}
	
	return &SearchFilesResponse{
		Files: files,
		Total: total,
	}, nil
}

// GetRecentFiles 获取最近文件
func (s *fileService) GetRecentFiles(ctx context.Context, userID uint, days, limit int) ([]*models.File, error) {
	files, err := s.repo.File.GetRecentFiles(ctx, userID, days, limit)
	if err != nil {
		return nil, NewServiceError("RecentFilesFailed", "failed to get recent files", err)
	}
	
	return files, nil
}

// MoveFile 移动文件
func (s *fileService) MoveFile(ctx context.Context, fileID, folderID uint, userID uint) error {
	// 检查文件权限
	file, err := s.GetFile(ctx, fileID, userID)
	if err != nil {
		return err
	}
	
	// 检查目标文件夹权限
	if folderID != 0 {
		folder, err := s.repo.Folder.GetByID(ctx, folderID)
		if err != nil {
			return NewServiceError("FolderNotFound", "target folder not found", err)
		}
		if folder.UserID != userID {
			return NewServiceError("PermissionDenied", "access denied to target folder", nil)
		}
	}
	
	// 执行移动
	var targetFolderID *uint
	if folderID != 0 {
		targetFolderID = &folderID
	}
	
	file.FolderID = targetFolderID
	if err := s.repo.File.Update(ctx, file); err != nil {
		return NewServiceError("FileMoveeFailed", "failed to move file", err)
	}
	
	return nil
}

// CopyFile 复制文件
func (s *fileService) CopyFile(ctx context.Context, fileID uint, destFolderID uint, userID uint) (*models.File, error) {
	// 获取源文件
	sourceFile, err := s.GetFile(ctx, fileID, userID)
	if err != nil {
		return nil, err
	}
	
	// 检查目标文件夹权限
	if destFolderID != 0 {
		folder, err := s.repo.Folder.GetByID(ctx, destFolderID)
		if err != nil {
			return nil, NewServiceError("FolderNotFound", "target folder not found", err)
		}
		if folder.UserID != userID {
			return nil, NewServiceError("PermissionDenied", "access denied to target folder", nil)
		}
	}
	
	// 创建复制文件记录
	copiedFile := &models.File{
		Name:         sourceFile.Name,
		OriginalName: sourceFile.OriginalName,
		Hash:         sourceFile.Hash,
		Size:         sourceFile.Size,
		ContentType:  sourceFile.ContentType,
		Extension:    sourceFile.Extension,
		Path:         sourceFile.Path, // 共享存储路径
		UserID:       userID,
		Category:     sourceFile.Category,
		StorageTier:  sourceFile.StorageTier,
		Tags:         sourceFile.Tags,
		Description:  sourceFile.Description,
		Status:       models.FileStatusActive,
		Metadata:     sourceFile.Metadata,
	}
	
	if destFolderID != 0 {
		copiedFile.FolderID = &destFolderID
	}
	
	// 生成唯一名称（如果需要）
	copiedFile.Name = s.generateUniqueFileName(ctx, copiedFile.Name, copiedFile.FolderID, userID)
	
	if err := s.repo.File.Create(ctx, copiedFile); err != nil {
		return nil, NewServiceError("FileCopyFailed", "failed to copy file", err)
	}
	
	return copiedFile, nil
}

// RenameFile 重命名文件
func (s *fileService) RenameFile(ctx context.Context, fileID uint, newName string, userID uint) error {
	// 验证新文件名
	if err := s.utils.ValidateFileName(newName); err != nil {
		return err
	}
	
	// 获取文件
	file, err := s.GetFile(ctx, fileID, userID)
	if err != nil {
		return err
	}
	
	// 检查同文件夹下是否有重名文件
	existingFiles, _, err := s.repo.File.List(ctx, userID, file.FolderID, 0, 1, &repository.FileFilters{
		SortBy: "name",
	})
	if err != nil {
		return NewServiceError("FileCheckFailed", "failed to check existing files", err)
	}
	
	for _, existing := range existingFiles {
		if existing.ID != fileID && strings.EqualFold(existing.Name, newName) {
			return NewServiceError("DuplicateFileName", "file with same name already exists", nil)
		}
	}
	
	// 更新文件名
	file.Name = newName
	if err := s.repo.File.Update(ctx, file); err != nil {
		return NewServiceError("FileRenameFailed", "failed to rename file", err)
	}
	
	return nil
}

// DeleteFile 删除文件
func (s *fileService) DeleteFile(ctx context.Context, fileID uint, userID uint) error {
	// 检查文件权限
	file, err := s.GetFile(ctx, fileID, userID)
	if err != nil {
		return err
	}
	
	return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		
		// 软删除文件记录
		if err := txRepo.File.Delete(ctx, fileID); err != nil {
			return NewServiceError("FileDeleteFailed", "failed to delete file", err)
		}
		
		// 检查是否还有其他引用相同存储文件的记录
		duplicates, err := txRepo.File.GetFilesByHashes(ctx, []string{file.Hash}, userID)
		if err != nil {
			return NewServiceError("DuplicateCheckFailed", "failed to check file duplicates", err)
		}
		
		// 如果没有其他引用，删除存储文件
		if len(duplicates) == 0 {
			if err := s.storage.Delete(ctx, file.Path); err != nil {
				// 存储删除失败不影响数据库操作
				// TODO: 记录日志，后续清理
			}
		}
		
		return nil
	})
}

// RestoreFile 恢复文件
func (s *fileService) RestoreFile(ctx context.Context, fileID uint, userID uint) error {
	// 获取已删除的文件（需要使用Unscoped查询）
	var file models.File
	err := s.repo.GetDB().Unscoped().Where("id = ? AND user_id = ?", fileID, userID).First(&file).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return NewServiceError("FileNotFound", "deleted file not found", err)
		}
		return NewServiceError("FileRetrieveFailed", "failed to retrieve deleted file", err)
	}
	
	// 检查文件是否已删除
	if file.DeletedAt.Time.IsZero() {
		return NewServiceError("FileNotDeleted", "file is not deleted", nil)
	}
	
	// 恢复文件
	if err := s.repo.GetDB().Unscoped().Model(&file).Update("deleted_at", nil).Error; err != nil {
		return NewServiceError("FileRestoreFailed", "failed to restore file", err)
	}
	
	return nil
}

// DownloadFile 下载文件
func (s *fileService) DownloadFile(ctx context.Context, fileID uint, userID uint) (*DownloadFileResponse, error) {
	// 获取文件
	file, err := s.GetFile(ctx, fileID, userID)
	if err != nil {
		return nil, err
	}
	
	// 从存储下载
	reader, err := s.storage.Download(ctx, file.Path)
	if err != nil {
		return nil, NewServiceError("FileDownloadFailed", "failed to download file from storage", err)
	}
	
	return &DownloadFileResponse{
		File:   file,
		Reader: reader,
	}, nil
}

// GetFileURL 获取文件URL
func (s *fileService) GetFileURL(ctx context.Context, fileID uint, userID uint, expires int) (string, error) {
	// 获取文件
	file, err := s.GetFile(ctx, fileID, userID)
	if err != nil {
		return "", err
	}
	
	// 生成预签名URL
	expireTime := time.Duration(expires) * time.Second
	if expireTime == 0 {
		expireTime = 24 * time.Hour // 默认24小时
	}
	
	url, err := s.storage.GetURL(ctx, file.Path, expireTime)
	if err != nil {
		return "", NewServiceError("URLGenerationFailed", "failed to generate file URL", err)
	}
	
	return url, nil
}

// GetFileVersions 获取文件版本
func (s *fileService) GetFileVersions(ctx context.Context, fileID uint, userID uint) ([]*models.File, error) {
	// 检查文件权限
	_, err := s.GetFile(ctx, fileID, userID)
	if err != nil {
		return nil, err
	}
	
	versions, err := s.repo.File.GetVersions(ctx, fileID)
	if err != nil {
		return nil, NewServiceError("VersionRetrieveFailed", "failed to retrieve file versions", err)
	}
	
	return versions, nil
}

// CreateFileVersion 创建文件版本
func (s *fileService) CreateFileVersion(ctx context.Context, fileID uint, req *UploadFileRequest) (*models.File, error) {
	// 检查原文件权限
	_, err := s.GetFile(ctx, fileID, req.UserID)
	if err != nil {
		return nil, err
	}
	
	// 上传新版本
	uploadResp, err := s.UploadFile(ctx, req)
	if err != nil {
		return nil, err
	}
	
	// 创建版本关系
	newFile := uploadResp.File
	err = s.repo.File.CreateVersion(ctx, newFile, fileID)
	if err != nil {
		// 如果版本创建失败，删除刚上传的文件
		s.DeleteFile(ctx, newFile.ID, req.UserID)
		return nil, NewServiceError("VersionCreateFailed", "failed to create file version", err)
	}
	
	return newFile, nil
}

// RestoreFileVersion 恢复文件版本
func (s *fileService) RestoreFileVersion(ctx context.Context, versionID uint, userID uint) error {
	// 获取版本文件
	version, err := s.GetFile(ctx, versionID, userID)
	if err != nil {
		return err
	}
	
	if version.ParentID == nil {
		return NewServiceError("NotAVersion", "file is not a version", nil)
	}
	
	// 获取主文件
	mainFile, err := s.GetFile(ctx, *version.ParentID, userID)
	if err != nil {
		return err
	}
	
	return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		
		// 将当前主文件标记为非最新
		mainFile.IsLatest = false
		if err := txRepo.File.Update(ctx, mainFile); err != nil {
			return err
		}
		
		// 将版本文件标记为最新
		version.IsLatest = true
		return txRepo.File.Update(ctx, version)
	})
}

// 辅助方法

// calculateFileHash 计算文件哈希
func (s *fileService) calculateFileHash(reader io.Reader) (string, io.Reader, error) {
	hash := md5.New()
	teeReader := io.TeeReader(reader, hash)
	
	// 创建缓冲区来重新创建reader
	buf := make([]byte, 32*1024) // 32KB buffer
	var data []byte
	
	for {
		n, err := teeReader.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", nil, err
		}
	}
	
	hashStr := hex.EncodeToString(hash.Sum(nil))
	newReader := strings.NewReader(string(data))
	
	return hashStr, newReader, nil
}

// getContentType 获取内容类型
func (s *fileService) getContentType(provided, extension string) string {
	if provided != "" {
		return provided
	}
	return s.utils.GetMimeType(extension)
}

// extractFileMetadata 提取文件元数据
func (s *fileService) extractFileMetadata(req *UploadFileRequest) models.FileMetadata {
	metadata := models.FileMetadata{}
	
	if s.utils.IsImageFile(s.utils.ExtractExtension(req.FileName)) {
		// TODO: 提取EXIF信息，可以添加到Custom字段中
		if metadata.Custom == nil {
			metadata.Custom = make(map[string]interface{})
		}
		metadata.Custom["upload_time"] = time.Now()
	}
	
	return metadata
}

// generateUniqueFileName 生成唯一文件名
func (s *fileService) generateUniqueFileName(ctx context.Context, baseName string, folderID *uint, userID uint) string {
	name := baseName
	counter := 1
	
	for {
		// 检查文件名是否存在
		existing, _, _ := s.repo.File.List(ctx, userID, folderID, 0, 1, &repository.FileFilters{
			SortBy: "name",
		})
		
		nameExists := false
		for _, file := range existing {
			if strings.EqualFold(file.Name, name) {
				nameExists = true
				break
			}
		}
		
		if !nameExists {
			return name
		}
		
		// 生成新的名称
		ext := filepath.Ext(baseName)
		nameWithoutExt := strings.TrimSuffix(baseName, ext)
		name = fmt.Sprintf("%s (%d)%s", nameWithoutExt, counter, ext)
		counter++
		
		// 防止无限循环
		if counter > 1000 {
			name = fmt.Sprintf("%s_%d%s", nameWithoutExt, time.Now().Unix(), ext)
			break
		}
	}
	
	return name
}

// UpdateFile 更新文件信息
func (s *fileService) UpdateFile(ctx context.Context, req *UpdateFileRequest) error {
	// 获取文件并检查权限
	file, err := s.repo.File.GetByID(ctx, req.FileID)
	if err != nil {
		return NewServiceError("FileNotFound", "file not found", err)
	}
	
	if file.UserID != req.UserID {
		return NewServiceError("PermissionDenied", "access denied to file", nil)
	}
	
	// 更新字段
	if req.Name != "" && req.Name != file.Name {
		// 验证新名称
		if err := s.utils.ValidateFileName(req.Name); err != nil {
			return err
		}
		file.Name = req.Name
	}
	
	if req.Description != "" {
		file.Description = req.Description
	}
	
	// 更新标签
	if req.Tags != nil {
		file.Tags = req.Tags
	}
	
	if req.IsPublic != nil {
		file.IsPublic = *req.IsPublic
	}
	
	// 更新自定义元数据
	if req.Metadata != nil && len(req.Metadata) > 0 {
		// Convert map[string]string to map[string]interface{}
		customData := make(map[string]interface{})
		for k, v := range req.Metadata {
			customData[k] = v
		}
		file.Metadata.Custom = customData
	}
	
	// 保存更新
	if err := s.repo.File.Update(ctx, file); err != nil {
		return NewServiceError("FileUpdateFailed", "failed to update file", err)
	}
	
	return nil
}

// GetDownloadURL 获取下载URL
func (s *fileService) GetDownloadURL(ctx context.Context, fileID uint, userID uint) (string, error) {
	// 复用GetFileURL方法，使用默认1小时过期时间
	return s.GetFileURL(ctx, fileID, userID, 3600)
}