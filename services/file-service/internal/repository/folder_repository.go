package repository

import (
	"context"
	"file-service/internal/models"
	"fmt"

	"gorm.io/gorm"
)

// FolderRepository 文件夹仓库接口
type FolderRepository interface {
	// 基本CRUD操作
	Create(ctx context.Context, folder *models.Folder) error
	GetByID(ctx context.Context, id uint) (*models.Folder, error)
	GetByPath(ctx context.Context, userID uint, path string) (*models.Folder, error)
	Update(ctx context.Context, folder *models.Folder) error
	Delete(ctx context.Context, id uint) error
	
	// 层级操作
	GetChildren(ctx context.Context, parentID uint) ([]*models.Folder, error)
	GetByUserID(ctx context.Context, userID uint) ([]*models.Folder, error)
	GetRootFolders(ctx context.Context, userID uint) ([]*models.Folder, error)
	GetFolderTree(ctx context.Context, userID uint, rootID *uint) ([]*models.Folder, error)
	
	// 路径操作
	GetFolderPath(ctx context.Context, folderID uint) ([]models.Folder, error)
	UpdateMaterializedPaths(ctx context.Context, folderID uint, newPath string) error
	
	// 移动操作
	Move(ctx context.Context, folderID, newParentID uint) error
	CanMoveTo(ctx context.Context, folderID, targetParentID uint) (bool, error)
	
	// 统计操作
	UpdateStatistics(ctx context.Context, folderID uint) error
	GetStatistics(ctx context.Context, folderID uint) (*FolderStats, error)
	RecalculateStatistics(ctx context.Context, userID uint) error
	
	// 查询操作
	Search(ctx context.Context, userID uint, keyword string, offset, limit int) ([]*models.Folder, int64, error)
	GetByName(ctx context.Context, userID uint, parentID *uint, name string) (*models.Folder, error)
	
	// 批量操作
	BatchDelete(ctx context.Context, ids []uint) error
	BatchUpdateStatus(ctx context.Context, ids []uint, status int) error
	
	// 系统文件夹
	CreateSystemFolders(ctx context.Context, userID uint) error
	GetSystemFolder(ctx context.Context, userID uint, folderName string) (*models.Folder, error)
}

// FolderStats 文件夹统计信息
type FolderStats struct {
	FolderCount int64 `json:"folder_count"`
	FileCount   int64 `json:"file_count"`
	TotalSize   int64 `json:"total_size"`
}

// folderRepository 文件夹仓库实现
type folderRepository struct {
	db *gorm.DB
}

// NewFolderRepository 创建文件夹仓库实例
func NewFolderRepository(db *gorm.DB) FolderRepository {
	return &folderRepository{db: db}
}

// Create 创建文件夹
func (r *folderRepository) Create(ctx context.Context, folder *models.Folder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建文件夹
		if err := tx.Create(folder).Error; err != nil {
			return err
		}
		
		// 生成物化路径
		if folder.ParentID != nil {
			var parent models.Folder
			if err := tx.First(&parent, *folder.ParentID).Error; err != nil {
				return err
			}
			folder.UpdateMaterializedPath(parent.MaterializedPath)
			folder.Level = parent.Level + 1
		} else {
			folder.UpdateMaterializedPath("")
			folder.Level = 0
		}
		
		// 更新物化路径
		return tx.Save(folder).Error
	})
}

// GetByID 根据ID获取文件夹
func (r *folderRepository) GetByID(ctx context.Context, id uint) (*models.Folder, error) {
	var folder models.Folder
	err := r.db.WithContext(ctx).
		Preload("Parent").
		First(&folder, id).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

// GetByPath 根据路径获取文件夹
func (r *folderRepository) GetByPath(ctx context.Context, userID uint, path string) (*models.Folder, error) {
	var folder models.Folder
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND path = ?", userID, path).
		First(&folder).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

// Update 更新文件夹
func (r *folderRepository) Update(ctx context.Context, folder *models.Folder) error {
	return r.db.WithContext(ctx).Save(folder).Error
}

// Delete 删除文件夹(软删除)
func (r *folderRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查是否有子文件夹或文件
		var childCount int64
		if err := tx.Model(&models.Folder{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
			return err
		}
		if childCount > 0 {
			return fmt.Errorf("文件夹不为空，无法删除")
		}
		
		var fileCount int64
		if err := tx.Model(&models.File{}).Where("folder_id = ?", id).Count(&fileCount).Error; err != nil {
			return err
		}
		if fileCount > 0 {
			return fmt.Errorf("文件夹不为空，无法删除")
		}
		
		// 执行删除
		return tx.Delete(&models.Folder{}, id).Error
	})
}

// GetChildren 获取子文件夹
func (r *folderRepository) GetChildren(ctx context.Context, parentID uint) ([]*models.Folder, error) {
	var folders []*models.Folder
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("name ASC").
		Find(&folders).Error
	return folders, err
}

// GetByUserID 获取用户所有文件夹
func (r *folderRepository) GetByUserID(ctx context.Context, userID uint) ([]*models.Folder, error) {
	var folders []*models.Folder
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("level ASC, name ASC").
		Find(&folders).Error
	return folders, err
}

// GetRootFolders 获取根文件夹
func (r *folderRepository) GetRootFolders(ctx context.Context, userID uint) ([]*models.Folder, error) {
	var folders []*models.Folder
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND parent_id IS NULL", userID).
		Order("sort_order ASC, name ASC").
		Find(&folders).Error
	return folders, err
}

// GetFolderTree 获取文件夹树
func (r *folderRepository) GetFolderTree(ctx context.Context, userID uint, rootID *uint) ([]*models.Folder, error) {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	
	if rootID != nil {
		// 获取指定根目录下的所有文件夹
		var root models.Folder
		if err := r.db.First(&root, *rootID).Error; err != nil {
			return nil, err
		}
		query = query.Where("materialized_path LIKE ?", root.MaterializedPath+"/%")
	}
	
	var folders []*models.Folder
	err := query.Order("level ASC, name ASC").Find(&folders).Error
	return folders, err
}

// GetFolderPath 获取文件夹路径
func (r *folderRepository) GetFolderPath(ctx context.Context, folderID uint) ([]models.Folder, error) {
	var folder models.Folder
	if err := r.db.WithContext(ctx).First(&folder, folderID).Error; err != nil {
		return nil, err
	}
	
	// 解析物化路径获取所有父文件夹ID
	pathComponents := folder.GetPathComponents()
	if len(pathComponents) == 0 {
		return []models.Folder{folder}, nil
	}
	
	var folders []models.Folder
	err := r.db.WithContext(ctx).
		Where("id IN ?", pathComponents).
		Order("level ASC").
		Find(&folders).Error
	
	return folders, err
}

// UpdateMaterializedPaths 更新物化路径
func (r *folderRepository) UpdateMaterializedPaths(ctx context.Context, folderID uint, newPath string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 获取当前文件夹信息
		var folder models.Folder
		if err := tx.First(&folder, folderID).Error; err != nil {
			return err
		}
		
		oldPath := folder.MaterializedPath
		folder.MaterializedPath = newPath
		
		// 更新当前文件夹
		if err := tx.Save(&folder).Error; err != nil {
			return err
		}
		
		// 更新所有子文件夹的路径
		return tx.Model(&models.Folder{}).
			Where("materialized_path LIKE ?", oldPath+"/%").
			Update("materialized_path", 
				gorm.Expr("REPLACE(materialized_path, ?, ?)", oldPath, newPath)).Error
	})
}

// Move 移动文件夹
func (r *folderRepository) Move(ctx context.Context, folderID, newParentID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查是否可以移动
		canMove, err := r.canMoveToTx(tx, folderID, newParentID)
		if err != nil {
			return err
		}
		if !canMove {
			return fmt.Errorf("无法移动文件夹：会形成循环引用")
		}
		
		// 获取新父文件夹信息
		var newParent models.Folder
		if err := tx.First(&newParent, newParentID).Error; err != nil {
			return err
		}
		
		// 更新父文件夹ID和层级
		var folder models.Folder
		if err := tx.First(&folder, folderID).Error; err != nil {
			return err
		}
		
		folder.ParentID = &newParentID
		folder.Level = newParent.Level + 1
		
		// 生成新的物化路径
		newPath := fmt.Sprintf("%s/%d", newParent.MaterializedPath, folderID)
		
		// 更新物化路径
		return r.updateMaterializedPathsTx(tx, folderID, folder.MaterializedPath, newPath)
	})
}

// CanMoveTo 检查是否可以移动到指定位置
func (r *folderRepository) CanMoveTo(ctx context.Context, folderID, targetParentID uint) (bool, error) {
	return r.canMoveToTx(r.db.WithContext(ctx), folderID, targetParentID)
}

// canMoveToTx 事务版本的移动检查
func (r *folderRepository) canMoveToTx(tx *gorm.DB, folderID, targetParentID uint) (bool, error) {
	// 不能移动到自己
	if folderID == targetParentID {
		return false, nil
	}
	
	// 获取目标父文件夹的路径
	var targetParent models.Folder
	if err := tx.First(&targetParent, targetParentID).Error; err != nil {
		return false, err
	}
	
	// 检查目标是否是当前文件夹的子文件夹
	return !targetParent.IsChildOf(folderID), nil
}

// updateMaterializedPathsTx 事务版本的路径更新
func (r *folderRepository) updateMaterializedPathsTx(tx *gorm.DB, folderID uint, oldPath, newPath string) error {
	// 更新当前文件夹
	if err := tx.Model(&models.Folder{}).
		Where("id = ?", folderID).
		Update("materialized_path", newPath).Error; err != nil {
		return err
	}
	
	// 更新所有子文件夹
	return tx.Model(&models.Folder{}).
		Where("materialized_path LIKE ?", oldPath+"/%").
		Update("materialized_path", 
			gorm.Expr("REPLACE(materialized_path, ?, ?)", oldPath, newPath)).Error
}

// UpdateStatistics 更新文件夹统计信息
func (r *folderRepository) UpdateStatistics(ctx context.Context, folderID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 统计直接子文件夹数量
		var folderCount int64
		if err := tx.Model(&models.Folder{}).
			Where("parent_id = ?", folderID).
			Count(&folderCount).Error; err != nil {
			return err
		}
		
		// 统计文件数量和总大小
		var fileStats struct {
			FileCount int64
			TotalSize int64
		}
		if err := tx.Model(&models.File{}).
			Where("folder_id = ?", folderID).
			Select("COUNT(*) as file_count, COALESCE(SUM(size), 0) as total_size").
			Scan(&fileStats).Error; err != nil {
			return err
		}
		
		// 更新统计信息
		return tx.Model(&models.Folder{}).
			Where("id = ?", folderID).
			Updates(map[string]interface{}{
				"folder_count": folderCount,
				"file_count":   fileStats.FileCount,
				"total_size":   fileStats.TotalSize,
			}).Error
	})
}

// GetStatistics 获取文件夹统计信息
func (r *folderRepository) GetStatistics(ctx context.Context, folderID uint) (*FolderStats, error) {
	var folder models.Folder
	err := r.db.WithContext(ctx).First(&folder, folderID).Error
	if err != nil {
		return nil, err
	}
	
	return &FolderStats{
		FolderCount: folder.FolderCount,
		FileCount:   folder.FileCount,
		TotalSize:   folder.TotalSize,
	}, nil
}

// RecalculateStatistics 重新计算用户所有文件夹统计信息
func (r *folderRepository) RecalculateStatistics(ctx context.Context, userID uint) error {
	var folders []*models.Folder
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("level DESC"). // 从最深层开始计算
		Find(&folders).Error; err != nil {
		return err
	}
	
	for _, folder := range folders {
		if err := r.UpdateStatistics(ctx, folder.ID); err != nil {
			return err
		}
	}
	
	return nil
}

// Search 搜索文件夹
func (r *folderRepository) Search(ctx context.Context, userID uint, keyword string, offset, limit int) ([]*models.Folder, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Folder{}).Where("user_id = ?", userID)
	
	if keyword != "" {
		// 使用LIKE搜索结合LOWER函数实现不区分大小写搜索（SQLite兼容）
		query = query.Where(
			"LOWER(name) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?)", 
			"%"+keyword+"%", "%"+keyword+"%")
	}
	
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	var folders []*models.Folder
	err := query.Preload("Parent").
		Order("level ASC, name ASC").
		Offset(offset).
		Limit(limit).
		Find(&folders).Error
	
	return folders, total, err
}

// GetByName 根据名称获取文件夹
func (r *folderRepository) GetByName(ctx context.Context, userID uint, parentID *uint, name string) (*models.Folder, error) {
	query := r.db.WithContext(ctx).Where("user_id = ? AND name = ?", userID, name)
	
	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	
	var folder models.Folder
	err := query.First(&folder).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

// BatchDelete 批量删除文件夹
func (r *folderRepository) BatchDelete(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Delete(&models.Folder{}, ids).Error
}

// BatchUpdateStatus 批量更新状态
func (r *folderRepository) BatchUpdateStatus(ctx context.Context, ids []uint, status int) error {
	return r.db.WithContext(ctx).
		Model(&models.Folder{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}

// CreateSystemFolders 创建系统文件夹
func (r *folderRepository) CreateSystemFolders(ctx context.Context, userID uint) error {
	systemFolders := []models.Folder{
		{
			Name:     models.SystemFolderImages,
			Path:     "/" + models.SystemFolderImages,
			UserID:   userID,
			IsSystem: true,
			Color:    "#4CAF50",
			Icon:     "image",
		},
		{
			Name:     models.SystemFolderVideos,
			Path:     "/" + models.SystemFolderVideos,
			UserID:   userID,
			IsSystem: true,
			Color:    "#FF5722",
			Icon:     "video",
		},
		{
			Name:     models.SystemFolderAudios,
			Path:     "/" + models.SystemFolderAudios,
			UserID:   userID,
			IsSystem: true,
			Color:    "#9C27B0",
			Icon:     "audio",
		},
		{
			Name:     models.SystemFolderDocuments,
			Path:     "/" + models.SystemFolderDocuments,
			UserID:   userID,
			IsSystem: true,
			Color:    "#2196F3",
			Icon:     "document",
		},
		{
			Name:     models.SystemFolderArchives,
			Path:     "/" + models.SystemFolderArchives,
			UserID:   userID,
			IsSystem: true,
			Color:    "#FF9800",
			Icon:     "archive",
		},
	}
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, folder := range systemFolders {
			if err := r.Create(ctx, &folder); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetSystemFolder 获取系统文件夹
func (r *folderRepository) GetSystemFolder(ctx context.Context, userID uint, folderName string) (*models.Folder, error) {
	var folder models.Folder
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND name = ? AND is_system = true", userID, folderName).
		First(&folder).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}