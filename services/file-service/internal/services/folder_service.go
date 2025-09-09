package services

import (
	"context"
	"file-service/internal/models"
	"file-service/internal/repository"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// folderService 文件夹服务实现
type folderService struct {
	repo  *repository.Repository
	utils PathUtils
}

// NewFolderService 创建文件夹服务实例
func NewFolderService(repo *repository.Repository) FolderService {
	return &folderService{
		repo:  repo,
		utils: PathUtils{},
	}
}

// CreateFolder 创建文件夹
func (s *folderService) CreateFolder(ctx context.Context, req *CreateFolderRequest) (*models.Folder, error) {
	// 验证文件夹名
	if err := s.utils.ValidateFileName(req.Name); err != nil {
		return nil, err
	}
	
	// 检查父文件夹权限 (跳过根目录 parent_id = 0)
	if req.ParentID != nil && *req.ParentID != 0 {
		parent, err := s.repo.Folder.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, NewServiceError("ParentFolderNotFound", "parent folder not found", err)
		}
		if parent.UserID != req.UserID {
			return nil, NewServiceError("PermissionDenied", "access denied to parent folder", nil)
		}
	}
	
	// 检查同级文件夹名称冲突
	existing, err := s.repo.Folder.GetByName(ctx, req.UserID, req.ParentID, req.Name)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, NewServiceError("FolderCheckFailed", "failed to check existing folder", err)
	}
	if existing != nil {
		return nil, NewServiceError("DuplicateFolderName", "folder with same name already exists", nil)
	}
	
	// 生成路径
	var path string
	if req.ParentID != nil && *req.ParentID != 0 {
		parent, err := s.repo.Folder.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, NewServiceError("ParentFolderNotFound", "parent folder not found for path generation", err)
		}
		path = parent.Path + "/" + req.Name
	} else {
		path = "/" + req.Name
	}
	
	// 创建文件夹
	folder := &models.Folder{
		Name:        req.Name,
		Path:        path,
		ParentID:    req.ParentID,
		UserID:      req.UserID,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		IsSystem:    false,
		Status:      models.FolderStatusActive,
	}
	
	if err := s.repo.Folder.Create(ctx, folder); err != nil {
		return nil, NewServiceError("FolderCreateFailed", "failed to create folder", err)
	}
	
	return folder, nil
}

// GetFolder 获取文件夹
func (s *folderService) GetFolder(ctx context.Context, folderID uint, userID uint) (*models.Folder, error) {
	folder, err := s.repo.Folder.GetByID(ctx, folderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewServiceError("FolderNotFound", "folder not found", err)
		}
		return nil, NewServiceError("FolderRetrieveFailed", "failed to retrieve folder", err)
	}
	
	if folder.UserID != userID {
		return nil, NewServiceError("PermissionDenied", "access denied to folder", nil)
	}
	
	return folder, nil
}

// GetFolderByPath 根据路径获取文件夹
func (s *folderService) GetFolderByPath(ctx context.Context, path string, userID uint) (*models.Folder, error) {
	folder, err := s.repo.Folder.GetByPath(ctx, userID, path)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewServiceError("FolderNotFound", "folder not found", err)
		}
		return nil, NewServiceError("FolderRetrieveFailed", "failed to retrieve folder", err)
	}
	
	return folder, nil
}

// UpdateFolder 更新文件夹
func (s *folderService) UpdateFolder(ctx context.Context, req *UpdateFolderRequest) error {
	// 获取文件夹
	folder, err := s.GetFolder(ctx, req.FolderID, req.UserID)
	if err != nil {
		return err
	}
	
	// 更新字段
	if req.Name != "" && req.Name != folder.Name {
		// 验证新名称
		if err := s.utils.ValidateFileName(req.Name); err != nil {
			return err
		}
		
		// 检查同级名称冲突
		existing, err := s.repo.Folder.GetByName(ctx, req.UserID, folder.ParentID, req.Name)
		if err != nil && err != gorm.ErrRecordNotFound {
			return NewServiceError("FolderCheckFailed", "failed to check existing folder", err)
		}
		if existing != nil && existing.ID != req.FolderID {
			return NewServiceError("DuplicateFolderName", "folder with same name already exists", nil)
		}
		
		folder.Name = req.Name
		// 更新路径
		if folder.ParentID != nil && *folder.ParentID != 0 {
			parent, err := s.repo.Folder.GetByID(ctx, *folder.ParentID)
			if err != nil {
				return NewServiceError("ParentFolderNotFound", "parent folder not found for path update", err)
			}
			folder.Path = parent.Path + "/" + req.Name
		} else {
			folder.Path = "/" + req.Name
		}
	}
	
	if req.Description != "" {
		folder.Description = req.Description
	}
	if req.Color != "" {
		folder.Color = req.Color
	}
	if req.Icon != "" {
		folder.Icon = req.Icon
	}
	
	if err := s.repo.Folder.Update(ctx, folder); err != nil {
		return NewServiceError("FolderUpdateFailed", "failed to update folder", err)
	}
	
	return nil
}

// DeleteFolder 删除文件夹
func (s *folderService) DeleteFolder(ctx context.Context, folderID uint, userID uint) error {
	// 检查文件夹权限
	folder, err := s.GetFolder(ctx, folderID, userID)
	if err != nil {
		return err
	}
	
	// 检查是否为系统文件夹
	if folder.IsSystem {
		return NewServiceError("CannotDeleteSystemFolder", "cannot delete system folder", nil)
	}
	
	// 删除文件夹（Repository层会检查是否为空）
	if err := s.repo.Folder.Delete(ctx, folderID); err != nil {
		return NewServiceError("FolderDeleteFailed", "failed to delete folder", err)
	}
	
	return nil
}

// ListFolders 列举文件夹
func (s *folderService) ListFolders(ctx context.Context, userID uint, parentID *uint) ([]*models.Folder, error) {
	if parentID != nil {
		// 检查父文件夹权限
		_, err := s.GetFolder(ctx, *parentID, userID)
		if err != nil {
			return nil, err
		}
		
		return s.repo.Folder.GetChildren(ctx, *parentID)
	}
	
	// 获取根文件夹
	return s.repo.Folder.GetRootFolders(ctx, userID)
}

// GetFolderTree 获取文件夹树
func (s *folderService) GetFolderTree(ctx context.Context, userID uint, rootID *uint) ([]*models.Folder, error) {
	if rootID != nil {
		// 检查根文件夹权限
		_, err := s.GetFolder(ctx, *rootID, userID)
		if err != nil {
			return nil, err
		}
	}
	
	return s.repo.Folder.GetFolderTree(ctx, userID, rootID)
}

// GetFolderPath 获取文件夹路径
func (s *folderService) GetFolderPath(ctx context.Context, folderID uint, userID uint) ([]models.Folder, error) {
	// 检查文件夹权限
	_, err := s.GetFolder(ctx, folderID, userID)
	if err != nil {
		return nil, err
	}
	
	return s.repo.Folder.GetFolderPath(ctx, folderID)
}

// MoveFolder 移动文件夹
func (s *folderService) MoveFolder(ctx context.Context, folderID, newParentID uint, userID uint) error {
	// 检查源文件夹权限
	folder, err := s.GetFolder(ctx, folderID, userID)
	if err != nil {
		return err
	}
	
	// 检查是否为系统文件夹
	if folder.IsSystem {
		return NewServiceError("CannotMoveSystemFolder", "cannot move system folder", nil)
	}
	
	// 检查目标父文件夹权限
	if newParentID != 0 {
		_, err := s.GetFolder(ctx, newParentID, userID)
		if err != nil {
			return err
		}
		
		// 检查是否可以移动（防止循环引用）
		canMove, err := s.repo.Folder.CanMoveTo(ctx, folderID, newParentID)
		if err != nil {
			return NewServiceError("MoveCheckFailed", "failed to check move possibility", err)
		}
		if !canMove {
			return NewServiceError("InvalidMove", "cannot move folder: would create circular reference", nil)
		}
	}
	
	// 执行移动
	if err := s.repo.Folder.Move(ctx, folderID, newParentID); err != nil {
		return NewServiceError("FolderMoveFailed", "failed to move folder", err)
	}
	
	return nil
}

// CopyFolder 复制文件夹
func (s *folderService) CopyFolder(ctx context.Context, folderID, destParentID uint, userID uint) (*models.Folder, error) {
	// 获取源文件夹
	sourceFolder, err := s.GetFolder(ctx, folderID, userID)
	if err != nil {
		return nil, err
	}
	
	// 检查目标父文件夹权限
	if destParentID != 0 {
		_, err := s.GetFolder(ctx, destParentID, userID)
		if err != nil {
			return nil, err
		}
	}
	
	// 生成唯一名称
	newName := s.generateUniqueFolderName(ctx, sourceFolder.Name, destParentID, userID)
	
	// 创建复制文件夹
	var parentID *uint
	if destParentID != 0 {
		parentID = &destParentID
	}
	
	copiedFolder := &models.Folder{
		Name:        newName,
		Path:        s.generateFolderPath(ctx, newName, parentID),
		ParentID:    parentID,
		UserID:      userID,
		Description: sourceFolder.Description,
		Color:       sourceFolder.Color,
		Icon:        sourceFolder.Icon,
		IsSystem:    false,
		Status:      models.FolderStatusActive,
	}
	
	if err := s.repo.Folder.Create(ctx, copiedFolder); err != nil {
		return nil, NewServiceError("FolderCopyFailed", "failed to copy folder", err)
	}
	
	// TODO: 递归复制子文件夹和文件
	
	return copiedFolder, nil
}

// RenameFolder 重命名文件夹
func (s *folderService) RenameFolder(ctx context.Context, folderID uint, newName string, userID uint) error {
	// 验证新名称
	if err := s.utils.ValidateFileName(newName); err != nil {
		return err
	}
	
	// 获取文件夹
	folder, err := s.GetFolder(ctx, folderID, userID)
	if err != nil {
		return err
	}
	
	// 检查是否为系统文件夹
	if folder.IsSystem {
		return NewServiceError("CannotRenameSystemFolder", "cannot rename system folder", nil)
	}
	
	// 检查同级名称冲突
	existing, err := s.repo.Folder.GetByName(ctx, userID, folder.ParentID, newName)
	if err != nil && err != gorm.ErrRecordNotFound {
		return NewServiceError("FolderCheckFailed", "failed to check existing folder", err)
	}
	if existing != nil && existing.ID != folderID {
		return NewServiceError("DuplicateFolderName", "folder with same name already exists", nil)
	}
	
	// 更新名称和路径
	folder.Name = newName
	if folder.ParentID != nil && *folder.ParentID != 0 {
		parent, err := s.repo.Folder.GetByID(ctx, *folder.ParentID)
		if err != nil {
			return NewServiceError("ParentFolderNotFound", "parent folder not found for path update", err)
		}
		folder.Path = parent.Path + "/" + newName
	} else {
		folder.Path = "/" + newName
	}
	
	if err := s.repo.Folder.Update(ctx, folder); err != nil {
		return NewServiceError("FolderRenameFailed", "failed to rename folder", err)
	}
	
	return nil
}

// GetFolderContents 获取文件夹内容
func (s *folderService) GetFolderContents(ctx context.Context, req *GetFolderContentsRequest) (*GetFolderContentsResponse, error) {
	// 检查文件夹权限
	if req.FolderID != nil {
		_, err := s.GetFolder(ctx, *req.FolderID, req.UserID)
		if err != nil {
			return nil, err
		}
	}
	
	// 获取子文件夹
	var folders []*models.Folder
	var err error
	
	if req.FolderID != nil {
		folders, err = s.repo.Folder.GetChildren(ctx, *req.FolderID)
	} else {
		folders, err = s.repo.Folder.GetRootFolders(ctx, req.UserID)
	}
	
	if err != nil {
		return nil, NewServiceError("FolderListFailed", "failed to list folders", err)
	}
	
	// 获取文件
	files, _, err := s.repo.File.List(ctx, req.UserID, req.FolderID, req.Offset, req.Limit, nil)
	if err != nil {
		return nil, NewServiceError("FileListFailed", "failed to list files", err)
	}
	
	total := int64(len(folders) + len(files))
	
	return &GetFolderContentsResponse{
		Folders: folders,
		Files:   files,
		Total:   total,
	}, nil
}

// GetFolderStats 获取文件夹统计
func (s *folderService) GetFolderStats(ctx context.Context, folderID uint, userID uint) (*repository.FolderStats, error) {
	// 检查文件夹权限
	_, err := s.GetFolder(ctx, folderID, userID)
	if err != nil {
		return nil, err
	}
	
	return s.repo.Folder.GetStatistics(ctx, folderID)
}

// CreateSystemFolders 创建系统文件夹
func (s *folderService) CreateSystemFolders(ctx context.Context, userID uint) error {
	return s.repo.Folder.CreateSystemFolders(ctx, userID)
}

// GetSystemFolder 获取系统文件夹
func (s *folderService) GetSystemFolder(ctx context.Context, userID uint, folderType string) (*models.Folder, error) {
	return s.repo.Folder.GetSystemFolder(ctx, userID, folderType)
}

// 辅助方法

// generateUniqueFolderName 生成唯一文件夹名
func (s *folderService) generateUniqueFolderName(ctx context.Context, baseName string, parentID uint, userID uint) string {
	name := baseName
	counter := 1
	
	var parentIDPtr *uint
	if parentID != 0 {
		parentIDPtr = &parentID
	}
	
	for {
		// 检查名称是否存在
		existing, err := s.repo.Folder.GetByName(ctx, userID, parentIDPtr, name)
		if err != nil || existing == nil {
			return name
		}
		
		// 生成新名称
		name = fmt.Sprintf("%s (%d)", baseName, counter)
		counter++
		
		// 防止无限循环
		if counter > 1000 {
			name = fmt.Sprintf("%s_%d", baseName, time.Now().Unix())
			break
		}
	}
	
	return name
}

// generateFolderPath 生成文件夹路径
func (s *folderService) generateFolderPath(ctx context.Context, name string, parentID *uint) string {
	if parentID != nil && *parentID != 0 {
		parent, err := s.repo.Folder.GetByID(ctx, *parentID)
		if err == nil {
			return parent.Path + "/" + name
		}
	}
	return "/" + name
}

// 高级文件夹操作

// SyncFolderStats 同步文件夹统计
func (s *folderService) SyncFolderStats(ctx context.Context, folderID uint, userID uint) error {
	// 检查文件夹权限
	_, err := s.GetFolder(ctx, folderID, userID)
	if err != nil {
		return err
	}
	
	return s.repo.Folder.UpdateStatistics(ctx, folderID)
}

// RecalculateAllStats 重新计算用户所有文件夹统计
func (s *folderService) RecalculateAllStats(ctx context.Context, userID uint) error {
	return s.repo.Folder.RecalculateStatistics(ctx, userID)
}

// SearchFolders 搜索文件夹
func (s *folderService) SearchFolders(ctx context.Context, userID uint, keyword string, offset, limit int) ([]*models.Folder, int64, error) {
	return s.repo.Folder.Search(ctx, userID, keyword, offset, limit)
}