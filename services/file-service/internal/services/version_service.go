package services

import (
	"context"
	"file-service/internal/models"
	"file-service/internal/repository"
	"file-service/internal/storage"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// VersionService 文件版本服务接口
type VersionService interface {
	// 版本管理
	CreateVersion(ctx context.Context, req *CreateVersionRequest) (*models.File, error)
	GetFileVersions(ctx context.Context, fileID uint, userID uint) ([]*models.File, error)
	GetVersion(ctx context.Context, versionID uint, userID uint) (*models.File, error)
	RestoreVersion(ctx context.Context, versionID uint, userID uint) (*models.File, error)
	DeleteVersion(ctx context.Context, versionID uint, userID uint) error
	
	// 版本比较
	CompareVersions(ctx context.Context, version1ID, version2ID uint, userID uint) (*VersionCompareResponse, error)
	
	// 版本清理
	CleanupOldVersions(ctx context.Context, fileID uint, keepCount int, userID uint) (*BatchOperationResponse, error)
	
	// 去重功能
	FindDuplicateFiles(ctx context.Context, userID uint, folderID *uint) (map[string][]*models.File, error)
	MergeDuplicateFiles(ctx context.Context, sourceFileIDs []uint, targetFileID uint, userID uint) error
	CleanupDuplicates(ctx context.Context, userID uint, strategy DuplicateCleanupStrategy) (*BatchOperationResponse, error)
	
	// 统计功能
	GetVersionStats(ctx context.Context, fileID uint, userID uint) (*VersionStats, error)
	GetDeduplicationStats(ctx context.Context, userID uint) (*DeduplicationStats, error)
}

// versionService 版本服务实现
type versionService struct {
	repo    *repository.Repository
	storage storage.Storage
	utils   PathUtils
}

// NewVersionService 创建版本服务实例
func NewVersionService(repo *repository.Repository, storage storage.Storage) VersionService {
	return &versionService{
		repo:    repo,
		storage: storage,
		utils:   PathUtils{},
	}
}

// CreateVersion 创建文件版本
func (s *versionService) CreateVersion(ctx context.Context, req *CreateVersionRequest) (*models.File, error) {
	// 获取原始文件
	originalFile, err := s.repo.File.GetByID(ctx, req.FileID)
	if err != nil {
		return nil, NewServiceError("FileNotFound", "original file not found", err)
	}

	if originalFile.UserID != req.UserID {
		return nil, NewServiceError("PermissionDenied", "access denied to file", nil)
	}

	// Create the new version
	newVersion := &models.File{
		Name:         req.FileName,
		OriginalName: req.FileName,
		Hash:         req.Hash,
		Size:         req.Size,
		ContentType:  req.ContentType,
		Extension:    s.utils.ExtractExtension(req.FileName),
		Path:         req.StoragePath,
		ParentID:     &originalFile.ID, // Point to original file
		UserID:       req.UserID,
		FolderID:     originalFile.FolderID,
		Category:     originalFile.Category,
		StorageTier:  "hot", // New versions default to hot storage
		IsLatest:     true,
		Version:      1, // Will be updated by repository layer
		Status:       models.FileStatusActive,
	}

	err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		// Create new version
		if err := txRepo.File.Create(ctx, newVersion); err != nil {
			return NewServiceError("VersionCreateFailed", "failed to create new version", err)
		}

		// If original file is latest, mark as non-latest
		if originalFile.IsLatest {
			originalFile.IsLatest = false
			if err := txRepo.File.Update(ctx, originalFile); err != nil {
				return NewServiceError("VersionUpdateFailed", "failed to update original file version", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return newVersion, nil

}

// GetFileVersions 获取文件版本列表
func (s *versionService) GetFileVersions(ctx context.Context, fileID uint, userID uint) ([]*models.File, error) {
	// 检查文件权限
	file, err := s.repo.File.GetByID(ctx, fileID)
	if err != nil {
		return nil, NewServiceError("FileNotFound", "file not found", err)
	}

	if file.UserID != userID {
		return nil, NewServiceError("PermissionDenied", "access denied to file", nil)
	}

	return s.repo.File.GetVersions(ctx, fileID)
}

// GetVersion 获取特定版本
func (s *versionService) GetVersion(ctx context.Context, versionID uint, userID uint) (*models.File, error) {
	version, err := s.repo.File.GetByID(ctx, versionID)
	if err != nil {
		return nil, NewServiceError("VersionNotFound", "version not found", err)
	}

	if version.UserID != userID {
		return nil, NewServiceError("PermissionDenied", "access denied to version", nil)
	}

	return version, nil
}

// RestoreVersion 恢复到指定版本
func (s *versionService) RestoreVersion(ctx context.Context, versionID uint, userID uint) (*models.File, error) {
	// 获取要恢复的版本
	version, err := s.GetVersion(ctx, versionID, userID)
	if err != nil {
		return nil, err
	}

	if version.ParentID == nil {
		return nil, NewServiceError("InvalidVersion", "cannot restore main file", nil)
	}

	// 获取主文件
	mainFile, err := s.repo.File.GetByID(ctx, *version.ParentID)
	if err != nil {
		return nil, NewServiceError("MainFileNotFound", "main file not found", err)
	}

	err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		// 将当前最新版本标记为非最新
		currentVersions, err := txRepo.File.GetVersions(ctx, mainFile.ID)
		if err != nil {
			return NewServiceError("VersionRetrieveFailed", "failed to get current versions", err)
		}

		for _, v := range currentVersions {
			if v.IsLatest {
				v.IsLatest = false
				if err := txRepo.File.Update(ctx, v); err != nil {
					return NewServiceError("VersionUpdateFailed", "failed to update version", err)
				}
			}
		}

		// 将恢复的版本标记为最新
		version.IsLatest = true
		if err := txRepo.File.Update(ctx, version); err != nil {
			return NewServiceError("VersionRestoreFailed", "failed to restore version", err)
		}

		// 更新主文件信息
		mainFile.Hash = version.Hash
		mainFile.Size = version.Size
		mainFile.Path = version.Path
		mainFile.ContentType = version.ContentType

		if err := txRepo.File.Update(ctx, mainFile); err != nil {
			return NewServiceError("MainFileUpdateFailed", "failed to update main file", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return mainFile, nil
}

// DeleteVersion 删除版本
func (s *versionService) DeleteVersion(ctx context.Context, versionID uint, userID uint) error {
	version, err := s.GetVersion(ctx, versionID, userID)
	if err != nil {
		return err
	}

	if version.ParentID == nil {
		return NewServiceError("CannotDeleteMainFile", "cannot delete main file version", nil)
	}

	if version.IsLatest {
		return NewServiceError("CannotDeleteLatestVersion", "cannot delete latest version", nil)
	}

	return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		// 删除数据库记录
		if err := txRepo.File.Delete(ctx, versionID); err != nil {
			return NewServiceError("VersionDeleteFailed", "failed to delete version", err)
		}

		// 异步删除存储文件
		go func() {
			if err := s.storage.Delete(context.Background(), version.Path); err != nil {
				// TODO: 记录日志，可能需要重试机制
			}
		}()

		return nil
	})
}

// CompareVersions 比较两个版本
func (s *versionService) CompareVersions(ctx context.Context, version1ID, version2ID uint, userID uint) (*VersionCompareResponse, error) {
	// 获取两个版本
	version1, err := s.GetVersion(ctx, version1ID, userID)
	if err != nil {
		return nil, err
	}

	version2, err := s.GetVersion(ctx, version2ID, userID)
	if err != nil {
		return nil, err
	}

	// 比较基本信息
	differences := []string{}

	if version1.Hash != version2.Hash {
		differences = append(differences, "Content hash differs")
	}

	if version1.Size != version2.Size {
		differences = append(differences, fmt.Sprintf("Size differs: %d vs %d bytes", version1.Size, version2.Size))
	}

	if version1.ContentType != version2.ContentType {
		differences = append(differences, fmt.Sprintf("Content type differs: %s vs %s", version1.ContentType, version2.ContentType))
	}

	timeDiff := version2.CreatedAt.Sub(version1.CreatedAt)

	return &VersionCompareResponse{
		Version1:       version1,
		Version2:       version2,
		Differences:    differences,
		TimeDifference: timeDiff,
		AreIdentical:   len(differences) == 0,
	}, nil
}

// CleanupOldVersions 清理旧版本
func (s *versionService) CleanupOldVersions(ctx context.Context, fileID uint, keepCount int, userID uint) (*BatchOperationResponse, error) {
	// 检查文件权限
	file, err := s.repo.File.GetByID(ctx, fileID)
	if err != nil {
		return nil, NewServiceError("FileNotFound", "file not found", err)
	}

	if file.UserID != userID {
		return nil, NewServiceError("PermissionDenied", "access denied to file", nil)
	}

	// 获取所有版本，按创建时间倒序
	versions, err := s.repo.File.GetVersions(ctx, fileID)
	if err != nil {
		return nil, NewServiceError("VersionRetrieveFailed", "failed to get file versions", err)
	}

	if len(versions) <= keepCount {
		return &BatchOperationResponse{
			Total:   0,
			Message: "No versions to cleanup",
		}, nil
	}

	// 确定要删除的版本（保留最新的N个版本）
	versionsToDelete := versions[keepCount:]
	var success []uint
	var failed []BatchError

	for _, version := range versionsToDelete {
		if version.IsLatest {
			continue // 跳过最新版本
		}

		if err := s.DeleteVersion(ctx, version.ID, userID); err != nil {
			failed = append(failed, BatchError{
				ID:      version.ID,
				Message: err.Error(),
			})
		} else {
			success = append(success, version.ID)
		}
	}

	message := "Version cleanup completed"
	if len(failed) > 0 {
		message += " with some errors"
	}

	return &BatchOperationResponse{
		Success: success,
		Failed:  failed,
		Total:   len(versionsToDelete),
		Message: message,
	}, nil
}

// FindDuplicateFiles 查找重复文件
func (s *versionService) FindDuplicateFiles(ctx context.Context, userID uint, folderID *uint) (map[string][]*models.File, error) {
	// 获取指定范围内的文件
	status := models.FileStatusActive
	files, _, err := s.repo.File.List(ctx, userID, folderID, 0, 10000, &repository.FileFilters{
		Status: &status,
	})
	if err != nil {
		return nil, NewServiceError("FileRetrieveFailed", "failed to retrieve files", err)
	}

	// 按哈希分组
	hashGroups := make(map[string][]*models.File)
	for _, file := range files {
		if file.ParentID == nil { // 只考虑主文件，不包括版本
			hashGroups[file.Hash] = append(hashGroups[file.Hash], file)
		}
	}

	// 过滤出有重复的组
	duplicates := make(map[string][]*models.File)
	for hash, fileGroup := range hashGroups {
		if len(fileGroup) > 1 {
			duplicates[hash] = fileGroup
		}
	}

	return duplicates, nil
}

// MergeDuplicateFiles 合并重复文件
func (s *versionService) MergeDuplicateFiles(ctx context.Context, sourceFileIDs []uint, targetFileID uint, userID uint) error {
	// 检查目标文件权限
	targetFile, err := s.repo.File.GetByID(ctx, targetFileID)
	if err != nil {
		return NewServiceError("TargetFileNotFound", "target file not found", err)
	}

	if targetFile.UserID != userID {
		return NewServiceError("PermissionDenied", "access denied to target file", nil)
	}

	return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		for _, sourceID := range sourceFileIDs {
			if sourceID == targetFileID {
				continue // 跳过目标文件自身
			}

			// 获取源文件
			sourceFile, err := txRepo.File.GetByID(ctx, sourceID)
			if err != nil {
				continue // 跳过不存在的文件
			}

			if sourceFile.UserID != userID {
				continue // 跳过没有权限的文件
			}

			// 检查文件哈希是否相同
			if sourceFile.Hash != targetFile.Hash {
				continue // 跳过不重复的文件
			}

			// 更新源文件的引用，将其指向目标文件
			sourceFile.ParentID = &targetFileID
			sourceFile.IsLatest = false
			sourceFile.Status = models.FileStatusMerged

			if err := txRepo.File.Update(ctx, sourceFile); err != nil {
				return NewServiceError("FileMergeFailed", "failed to merge source file", err)
			}

			// 异步删除源文件的存储（如果路径不同）
			if sourceFile.Path != targetFile.Path {
				go func(path string) {
					if err := s.storage.Delete(context.Background(), path); err != nil {
						// TODO: 记录日志
					}
				}(sourceFile.Path)
			}
		}

		return nil
	})
}

// CleanupDuplicates 清理重复文件
func (s *versionService) CleanupDuplicates(ctx context.Context, userID uint, strategy DuplicateCleanupStrategy) (*BatchOperationResponse, error) {
	duplicates, err := s.FindDuplicateFiles(ctx, userID, nil)
	if err != nil {
		return nil, err
	}

	var success []uint
	var failed []BatchError

	for _, fileGroup := range duplicates {
		if len(fileGroup) <= 1 {
			continue
		}

		// 确定保留的文件
		var keepFile *models.File
		var filesToMerge []uint

		switch strategy {
		case DuplicateCleanupKeepOldest:
			for _, file := range fileGroup {
				if keepFile == nil || file.CreatedAt.Before(keepFile.CreatedAt) {
					keepFile = file
				}
			}
		case DuplicateCleanupKeepNewest:
			for _, file := range fileGroup {
				if keepFile == nil || file.CreatedAt.After(keepFile.CreatedAt) {
					keepFile = file
				}
			}
		case DuplicateCleanupKeepLargestName:
			for _, file := range fileGroup {
				if keepFile == nil || len(file.Name) > len(keepFile.Name) {
					keepFile = file
				}
			}
		}

		// 收集要合并的文件ID
		for _, file := range fileGroup {
			if file.ID != keepFile.ID {
				filesToMerge = append(filesToMerge, file.ID)
			}
		}

		// 执行合并
		if err := s.MergeDuplicateFiles(ctx, filesToMerge, keepFile.ID, userID); err != nil {
			for _, id := range filesToMerge {
				failed = append(failed, BatchError{
					ID:      id,
					Message: err.Error(),
				})
			}
		} else {
			success = append(success, filesToMerge...)
		}
	}

	message := "Duplicate cleanup completed"
	if len(failed) > 0 {
		message += " with some errors"
	}

	return &BatchOperationResponse{
		Success: success,
		Failed:  failed,
		Total:   len(success) + len(failed),
		Message: message,
	}, nil
}

// GetVersionStats 获取版本统计
func (s *versionService) GetVersionStats(ctx context.Context, fileID uint, userID uint) (*VersionStats, error) {
	// 检查文件权限
	file, err := s.repo.File.GetByID(ctx, fileID)
	if err != nil {
		return nil, NewServiceError("FileNotFound", "file not found", err)
	}

	if file.UserID != userID {
		return nil, NewServiceError("PermissionDenied", "access denied to file", nil)
	}

	versions, err := s.repo.File.GetVersions(ctx, fileID)
	if err != nil {
		return nil, NewServiceError("VersionRetrieveFailed", "failed to get file versions", err)
	}

	var totalSize int64
	var currentVersion *models.File

	for _, version := range versions {
		totalSize += version.Size
		if version.IsLatest {
			currentVersion = version
		}
	}

	return &VersionStats{
		TotalVersions:  len(versions),
		CurrentVersion: currentVersion.Version,
		TotalSize:      totalSize,
		CreatedAt:      file.CreatedAt,
		LastModified:   file.UpdatedAt,
	}, nil
}

// GetDeduplicationStats 获取去重统计
func (s *versionService) GetDeduplicationStats(ctx context.Context, userID uint) (*DeduplicationStats, error) {
	duplicates, err := s.FindDuplicateFiles(ctx, userID, nil)
	if err != nil {
		return nil, err
	}

	var totalDuplicateFiles int
	var potentialSavings int64
	var duplicateGroups int = len(duplicates)

	for _, fileGroup := range duplicates {
		totalDuplicateFiles += len(fileGroup)
		
		// 计算可节省的空间（保留一个文件，删除其余的）
		if len(fileGroup) > 1 {
			fileSize := fileGroup[0].Size // 所有文件大小相同
			potentialSavings += fileSize * int64(len(fileGroup)-1)
		}
	}

	return &DeduplicationStats{
		DuplicateGroups:   duplicateGroups,
		DuplicateFiles:    totalDuplicateFiles,
		PotentialSavings:  potentialSavings,
		SavingsPercent:    0, // TODO: 计算相对于总存储的百分比
	}, nil
}

// 请求和响应结构体

type CreateVersionRequest struct {
	FileID      uint   `json:"file_id"`
	FileName    string `json:"file_name"`
	Hash        string `json:"hash"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	StoragePath string `json:"storage_path"`
	VersionNote string `json:"version_note"`
	UserID      uint   `json:"user_id"`
}

type VersionCompareResponse struct {
	Version1       *models.File  `json:"version1"`
	Version2       *models.File  `json:"version2"`
	Differences    []string      `json:"differences"`
	TimeDifference time.Duration `json:"time_difference"`
	AreIdentical   bool          `json:"are_identical"`
}

type VersionStats struct {
	TotalVersions  int          `json:"total_versions"`
	CurrentVersion int          `json:"current_version"`
	TotalSize      int64        `json:"total_size"`
	CreatedAt      time.Time    `json:"created_at"`
	LastModified   time.Time    `json:"last_modified"`
}

type DeduplicationStats struct {
	DuplicateGroups  int     `json:"duplicate_groups"`
	DuplicateFiles   int     `json:"duplicate_files"`
	PotentialSavings int64   `json:"potential_savings"`
	SavingsPercent   float64 `json:"savings_percent"`
}

// 去重策略枚举
type DuplicateCleanupStrategy int

const (
	DuplicateCleanupKeepOldest DuplicateCleanupStrategy = iota
	DuplicateCleanupKeepNewest
	DuplicateCleanupKeepLargestName
)