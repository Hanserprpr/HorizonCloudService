package models

import "fmt"

// Thumbnail 缩略图模型
type Thumbnail struct {
	BaseModel
	
	// 关联信息
	FileID      uint   `gorm:"not null;index" json:"file_id"`                          // 关联文件ID
	
	// 缩略图规格
	Size        string `gorm:"not null;size:20;index" json:"size"`                     // 缩略图尺寸规格
	Width       int    `gorm:"not null" json:"width"`                                  // 实际宽度
	Height      int    `gorm:"not null" json:"height"`                                 // 实际高度
	Quality     int    `gorm:"default:80" json:"quality"`                              // 图片质量(1-100)
	
	// 存储信息
	Path        string `gorm:"not null;size:1000;uniqueIndex" json:"path"`             // 存储路径
	FileSize    int64  `gorm:"not null" json:"file_size"`                              // 缩略图文件大小
	ContentType string `gorm:"not null;size:100" json:"content_type"`                  // MIME类型
	ETag        string `gorm:"size:100" json:"etag,omitempty"`                         // 存储ETag
	
	// 状态
	Status      int    `gorm:"not null;default:1;index" json:"status"`                 // 状态: 1=正常 2=生成中 3=失败
	
	// 访问信息
	DownloadURL string `gorm:"size:500" json:"download_url,omitempty"`                 // 预签名下载URL
	
	// 关联关系
	File        *File  `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"file,omitempty"`
}

// TableName 指定表名
func (Thumbnail) TableName() string {
	return TablePrefix + "thumbnails"
}

// ThumbnailStatus 缩略图状态常量
const (
	ThumbnailStatusReady      = 1 // 正常
	ThumbnailStatusGenerating = 2 // 生成中
	ThumbnailStatusFailed     = 3 // 生成失败
)

// 缩略图尺寸规格常量
const (
	ThumbnailSizeSmall  = "small"   // 小尺寸: 150x150
	ThumbnailSizeMedium = "medium"  // 中尺寸: 300x300
	ThumbnailSizeLarge  = "large"   // 大尺寸: 600x600
)

// 缩略图规格配置
var ThumbnailSizes = map[string]struct {
	Width  int
	Height int
}{
	ThumbnailSizeSmall:  {Width: 150, Height: 150},
	ThumbnailSizeMedium: {Width: 300, Height: 300},
	ThumbnailSizeLarge:  {Width: 600, Height: 600},
}

// IsReady 检查缩略图是否准备就绪
func (t *Thumbnail) IsReady() bool {
	return t.Status == ThumbnailStatusReady
}

// IsGenerating 检查是否正在生成中
func (t *Thumbnail) IsGenerating() bool {
	return t.Status == ThumbnailStatusGenerating
}

// IsFailed 检查是否生成失败
func (t *Thumbnail) IsFailed() bool {
	return t.Status == ThumbnailStatusFailed
}

// MarkReady 标记为就绪
func (t *Thumbnail) MarkReady() {
	t.Status = ThumbnailStatusReady
}

// MarkGenerating 标记为生成中
func (t *Thumbnail) MarkGenerating() {
	t.Status = ThumbnailStatusGenerating
}

// MarkFailed 标记为失败
func (t *Thumbnail) MarkFailed() {
	t.Status = ThumbnailStatusFailed
}

// GetDisplaySize 获取显示尺寸
func (t *Thumbnail) GetDisplaySize() string {
	return t.Size
}

// GetSizeDisplay 获取可读的文件大小
func (t *Thumbnail) GetSizeDisplay() string {
	size := float64(t.FileSize)
	units := []string{"B", "KB", "MB"}
	
	for i, unit := range units {
		if size < 1024 || i == len(units)-1 {
			if size < 10 {
				return fmt.Sprintf("%.2f %s", size, unit)
			}
			return fmt.Sprintf("%.1f %s", size, unit)
		}
		size /= 1024
	}
	return fmt.Sprintf("%d B", t.FileSize)
}