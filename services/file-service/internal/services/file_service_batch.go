package services

import (
	"context"
	"file-service/internal/models"
	"file-service/internal/repository"
	"fmt"
	"sync"
)

// BatchMoveFiles 批量移动文件
func (s *fileService) BatchMoveFiles(ctx context.Context, fileIDs []uint, folderID uint, userID uint) (*BatchOperationResponse, error) {
	if len(fileIDs) == 0 {
		return &BatchOperationResponse{
			Total:   0,
			Message: "No files to move",
		}, nil
	}
	
	// 检查目标文件夹权限
	if folderID != 0 {
		folder, err := s.repo.Folder.GetByID(ctx, folderID)
		if err != nil {
			return nil, NewServiceError("FolderNotFound", "target folder not found", err)
		}
		if folder.UserID != userID {
			return nil, NewServiceError("PermissionDenied", "access denied to target folder", nil)
		}
	}
	
	var success []uint
	var failed []BatchError
	var mu sync.Mutex
	var wg sync.WaitGroup
	
	// 并发处理，但限制并发数
	semaphore := make(chan struct{}, 10) // 最多10个并发
	
	for _, fileID := range fileIDs {
		wg.Add(1)
		go func(id uint) {
			defer wg.Done()
			semaphore <- struct{}{} // 获取信号量
			defer func() { <-semaphore }() // 释放信号量
			
			err := s.MoveFile(ctx, id, folderID, userID)
			
			mu.Lock()
			if err != nil {
				failed = append(failed, BatchError{
					ID:      id,
					Message: err.Error(),
				})
			} else {
				success = append(success, id)
			}
			mu.Unlock()
		}(fileID)
	}
	
	wg.Wait()
	
	message := "Batch move completed"
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

// BatchDeleteFiles 批量删除文件
func (s *fileService) BatchDeleteFiles(ctx context.Context, fileIDs []uint, userID uint) (*BatchOperationResponse, error) {
	if len(fileIDs) == 0 {
		return &BatchOperationResponse{
			Total:   0,
			Message: "No files to delete",
		}, nil
	}
	
	var success []uint
	var failed []BatchError
	var mu sync.Mutex
	var wg sync.WaitGroup
	
	// 限制并发数
	semaphore := make(chan struct{}, 10)
	
	for _, fileID := range fileIDs {
		wg.Add(1)
		go func(id uint) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			err := s.DeleteFile(ctx, id, userID)
			
			mu.Lock()
			if err != nil {
				failed = append(failed, BatchError{
					ID:      id,
					Message: err.Error(),
				})
			} else {
				success = append(success, id)
			}
			mu.Unlock()
		}(fileID)
	}
	
	wg.Wait()
	
	message := "Batch delete completed"
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

// BatchCopyFiles 批量复制文件
func (s *fileService) BatchCopyFiles(ctx context.Context, fileIDs []uint, folderID uint, userID uint) (*BatchOperationResponse, error) {
	if len(fileIDs) == 0 {
		return &BatchOperationResponse{
			Total:   0,
			Message: "No files to copy",
		}, nil
	}
	
	// 检查目标文件夹权限
	if folderID != 0 {
		folder, err := s.repo.Folder.GetByID(ctx, folderID)
		if err != nil {
			return nil, NewServiceError("FolderNotFound", "target folder not found", err)
		}
		if folder.UserID != userID {
			return nil, NewServiceError("PermissionDenied", "access denied to target folder", nil)
		}
	}
	
	var success []uint
	var failed []BatchError
	var mu sync.Mutex
	var wg sync.WaitGroup
	
	// 限制并发数
	semaphore := make(chan struct{}, 5) // 复制操作更消耗资源，限制为5个并发
	
	for _, fileID := range fileIDs {
		wg.Add(1)
		go func(id uint) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			copiedFile, err := s.CopyFile(ctx, id, folderID, userID)
			
			mu.Lock()
			if err != nil {
				failed = append(failed, BatchError{
					ID:      id,
					Message: err.Error(),
				})
			} else {
				success = append(success, copiedFile.ID)
			}
			mu.Unlock()
		}(fileID)
	}
	
	wg.Wait()
	
	message := "Batch copy completed"
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

// GetUserStats 获取用户统计
func (s *fileService) GetUserStats(ctx context.Context, userID uint) (*UserFileStats, error) {
	// 获取文件数量和存储使用量
	fileCount, err := s.repo.File.GetUserFileCount(ctx, userID)
	if err != nil {
		return nil, NewServiceError("StatsRetrieveFailed", "failed to get file count", err)
	}
	
	storageUsed, err := s.repo.File.GetUserStorageUsed(ctx, userID)
	if err != nil {
		return nil, NewServiceError("StatsRetrieveFailed", "failed to get storage used", err)
	}
	
	// 获取文件夹数量
	folders, err := s.repo.Folder.GetByUserID(ctx, userID)
	if err != nil {
		return nil, NewServiceError("StatsRetrieveFailed", "failed to get folder count", err)
	}
	
	// TODO: 从用户服务获取存储配额
	storageQuota := int64(10 * 1024 * 1024 * 1024) // 默认10GB
	
	var usagePercent float64
	if storageQuota > 0 {
		usagePercent = float64(storageUsed) / float64(storageQuota) * 100
	}
	
	return &UserFileStats{
		TotalFiles:   fileCount,
		TotalSize:    storageUsed,
		TotalFolders: int64(len(folders)),
		StorageUsed:  storageUsed,
		StorageQuota: storageQuota,
		UsagePercent: usagePercent,
	}, nil
}

// GetCategoryStats 获取分类统计
func (s *fileService) GetCategoryStats(ctx context.Context, userID uint) (map[string]int64, error) {
	stats, err := s.repo.File.GetCategoryStats(ctx, userID)
	if err != nil {
		return nil, NewServiceError("StatsRetrieveFailed", "failed to get category stats", err)
	}
	
	return stats, nil
}

// GetStorageStats 获取存储统计
func (s *fileService) GetStorageStats(ctx context.Context, userID uint) (*StorageStats, error) {
	// 获取用户存储使用量
	usedSize, err := s.repo.File.GetUserStorageUsed(ctx, userID)
	if err != nil {
		return nil, NewServiceError("StatsRetrieveFailed", "failed to get storage used", err)
	}
	
	// 获取文件数量
	fileCount, err := s.repo.File.GetUserFileCount(ctx, userID)
	if err != nil {
		return nil, NewServiceError("StatsRetrieveFailed", "failed to get file count", err)
	}
	
	// 获取分类统计
	categoryStats, err := s.repo.File.GetCategoryStats(ctx, userID)
	if err != nil {
		return nil, NewServiceError("StatsRetrieveFailed", "failed to get category stats", err)
	}
	
	// 获取存储层级统计
	tierStats, err := s.repo.File.GetStorageTierStats(ctx, userID)
	if err != nil {
		return nil, NewServiceError("StatsRetrieveFailed", "failed to get tier stats", err)
	}
	
	// TODO: 从用户服务获取总配额
	totalSize := int64(10 * 1024 * 1024 * 1024) // 默认10GB
	availableSize := totalSize - usedSize
	if availableSize < 0 {
		availableSize = 0
	}
	
	return &StorageStats{
		TotalSize:     totalSize,
		UsedSize:      usedSize,
		AvailableSize: availableSize,
		FileCount:     fileCount,
		CategoryStats: categoryStats,
		TierStats:     tierStats,
	}, nil
}

// 高级文件操作

// DuplicateFiles 查找重复文件
func (s *fileService) DuplicateFiles(ctx context.Context, userID uint) (map[string][]*models.File, error) {
	// 获取用户所有文件
	allFiles, _, err := s.repo.File.List(ctx, userID, nil, 0, 10000, nil)
	if err != nil {
		return nil, NewServiceError("FileRetrieveFailed", "failed to retrieve user files", err)
	}
	
	// 按哈希分组
	hashGroups := make(map[string][]*models.File)
	for _, file := range allFiles {
		hashGroups[file.Hash] = append(hashGroups[file.Hash], file)
	}
	
	// 过滤出有重复的文件组
	duplicates := make(map[string][]*models.File)
	for hash, files := range hashGroups {
		if len(files) > 1 {
			duplicates[hash] = files
		}
	}
	
	return duplicates, nil
}

// CleanupDuplicates 清理重复文件
func (s *fileService) CleanupDuplicates(ctx context.Context, userID uint, keepLatest bool) (*BatchOperationResponse, error) {
	duplicates, err := s.DuplicateFiles(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	var filesToDelete []uint
	
	for _, files := range duplicates {
		if len(files) <= 1 {
			continue
		}
		
		// 找到要保留的文件
		var keepFile *models.File
		if keepLatest {
			// 保留最新的
			for _, file := range files {
				if keepFile == nil || file.CreatedAt.After(keepFile.CreatedAt) {
					keepFile = file
				}
			}
		} else {
			// 保留最旧的
			for _, file := range files {
				if keepFile == nil || file.CreatedAt.Before(keepFile.CreatedAt) {
					keepFile = file
				}
			}
		}
		
		// 标记其他文件为删除
		for _, file := range files {
			if file.ID != keepFile.ID {
				filesToDelete = append(filesToDelete, file.ID)
			}
		}
	}
	
	if len(filesToDelete) == 0 {
		return &BatchOperationResponse{
			Total:   0,
			Message: "No duplicate files to clean up",
		}, nil
	}
	
	// 批量删除重复文件
	return s.BatchDeleteFiles(ctx, filesToDelete, userID)
}

// ArchiveOldFiles 归档旧文件
func (s *fileService) ArchiveOldFiles(ctx context.Context, userID uint, daysOld int) (*BatchOperationResponse, error) {
	// 获取旧文件
	filters := &repository.FileFilters{
		DateTo:      stringPtr(fmt.Sprintf("%d days ago", daysOld)),
		StorageTier: "hot", // 只归档热存储的文件
	}
	
	oldFiles, _, err := s.repo.File.List(ctx, userID, nil, 0, 10000, filters)
	if err != nil {
		return nil, NewServiceError("FileRetrieveFailed", "failed to retrieve old files", err)
	}
	
	if len(oldFiles) == 0 {
		return &BatchOperationResponse{
			Total:   0,
			Message: "No old files to archive",
		}, nil
	}
	
	var success []uint
	var failed []BatchError
	
	// 批量更新存储层级
	for _, file := range oldFiles {
		file.StorageTier = "cold"
		if err := s.repo.File.Update(ctx, file); err != nil {
			failed = append(failed, BatchError{
				ID:      file.ID,
				Message: err.Error(),
			})
		} else {
			success = append(success, file.ID)
		}
	}
	
	message := "Archive completed"
	if len(failed) > 0 {
		message += " with some errors"
	}
	
	return &BatchOperationResponse{
		Success: success,
		Failed:  failed,
		Total:   len(oldFiles),
		Message: message,
	}, nil
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}