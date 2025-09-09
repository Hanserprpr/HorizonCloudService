package models

import (
	"fmt"
	"strings"
)

// Folder 文件夹模型
type Folder struct {
	BaseModel
	
	// 基本信息
	Name        string `gorm:"not null;size:255;index" json:"name"`                     // 文件夹名称
	Path        string `gorm:"not null;size:1000;index" json:"path"`                   // 文件夹路径
	Description string `gorm:"size:500" json:"description,omitempty"`                  // 描述
	
	// 层级关系
	ParentID    *uint   `gorm:"index" json:"parent_id,omitempty"`                       // 父文件夹ID
	Level       int     `gorm:"not null;default:0;index" json:"level"`                  // 层级深度
	MaterializedPath string `gorm:"size:2000;index" json:"materialized_path"`          // 物化路径，用于快速查询
	
	// 所属关系
	UserID      uint   `gorm:"not null;index" json:"user_id"`                          // 所属用户ID
	
	// 状态和权限
	Status      int    `gorm:"not null;default:1;index" json:"status"`                 // 状态: 1=正常 2=已删除
	IsSystem    bool   `gorm:"default:false" json:"is_system"`                         // 是否系统文件夹
	IsShared    bool   `gorm:"default:false;index" json:"is_shared"`                   // 是否共享
	
	// 统计信息
	FileCount   int64  `gorm:"default:0" json:"file_count"`                            // 文件数量
	FolderCount int64  `gorm:"default:0" json:"folder_count"`                          // 子文件夹数量
	TotalSize   int64  `gorm:"default:0" json:"total_size"`                            // 总大小(字节)
	
	// 显示和排序
	Color       string `gorm:"size:7" json:"color,omitempty"`                          // 显示颜色
	Icon        string `gorm:"size:50" json:"icon,omitempty"`                          // 图标
	SortOrder   int    `gorm:"default:0" json:"sort_order"`                            // 排序权重
	
	// 关联关系
	Parent      *Folder  `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE" json:"parent,omitempty"`
	Children    []Folder `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE" json:"children,omitempty"`
	Files       []File   `gorm:"foreignKey:FolderID;constraint:OnDelete:CASCADE" json:"files,omitempty"`
	Shares      []Share  `gorm:"foreignKey:FolderID;constraint:OnDelete:CASCADE" json:"shares,omitempty"`
}

// TableName 指定表名
func (Folder) TableName() string {
	return TablePrefix + "folders"
}

// FolderStatus 文件夹状态常量
const (
	FolderStatusActive  = 1 // 活跃/正常 (别名: FolderStatusNormal)
	FolderStatusNormal  = 1 // 正常
	FolderStatusDeleted = 2 // 已删除
)

// 系统文件夹名称常量
const (
	SystemFolderRoot      = "/"           // 根目录
	SystemFolderImages    = "Images"      // 图片文件夹
	SystemFolderVideos    = "Videos"      // 视频文件夹
	SystemFolderAudios    = "Audios"      // 音频文件夹
	SystemFolderDocuments = "Documents"   // 文档文件夹
	SystemFolderArchives  = "Archives"    // 压缩包文件夹
	SystemFolderTrash     = "Trash"       // 回收站
)

// GeneratePath 生成文件夹完整路径
func (f *Folder) GeneratePath() string {
	if f.ParentID == nil {
		return "/" + f.Name
	}
	return f.Path + "/" + f.Name
}

// UpdateMaterializedPath 更新物化路径
func (f *Folder) UpdateMaterializedPath(parentPath string) {
	if parentPath == "" {
		f.MaterializedPath = fmt.Sprintf("/%d", f.ID)
	} else {
		f.MaterializedPath = fmt.Sprintf("%s/%d", parentPath, f.ID)
	}
}

// GetPathComponents 获取路径组件
func (f *Folder) GetPathComponents() []string {
	if f.MaterializedPath == "" {
		return []string{}
	}
	
	components := strings.Split(strings.Trim(f.MaterializedPath, "/"), "/")
	var result []string
	for _, comp := range components {
		if comp != "" {
			result = append(result, comp)
		}
	}
	return result
}

// IsChildOf 检查是否为指定文件夹的子文件夹
func (f *Folder) IsChildOf(parentID uint) bool {
	path := f.MaterializedPath
	parentPath := fmt.Sprintf("/%d/", parentID)
	return strings.Contains(path, parentPath) || 
		   strings.HasSuffix(path, fmt.Sprintf("/%d", parentID))
}

// IsRoot 检查是否为根文件夹
func (f *Folder) IsRoot() bool {
	return f.ParentID == nil || f.Level == 0
}

// GetDisplayName 获取显示名称
func (f *Folder) GetDisplayName() string {
	if f.Name == SystemFolderRoot {
		return "根目录"
	}
	return f.Name
}

// AddFile 增加文件统计
func (f *Folder) AddFile(size int64) {
	f.FileCount++
	f.TotalSize += size
}

// RemoveFile 减少文件统计  
func (f *Folder) RemoveFile(size int64) {
	if f.FileCount > 0 {
		f.FileCount--
	}
	if f.TotalSize >= size {
		f.TotalSize -= size
	}
}

// AddFolder 增加文件夹统计
func (f *Folder) AddFolder() {
	f.FolderCount++
}

// RemoveFolder 减少文件夹统计
func (f *Folder) RemoveFolder() {
	if f.FolderCount > 0 {
		f.FolderCount--
	}
}

// GetSizeDisplay 获取可读的大小显示
func (f *Folder) GetSizeDisplay() string {
	size := float64(f.TotalSize)
	units := []string{"B", "KB", "MB", "GB", "TB"}
	
	for i, unit := range units {
		if size < 1024 || i == len(units)-1 {
			return fmt.Sprintf("%.2f %s", size, unit)
		}
		size /= 1024
	}
	return fmt.Sprintf("%d B", f.TotalSize)
}