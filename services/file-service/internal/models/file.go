package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// File 文件模型
type File struct {
	BaseModel
	
	// 基本信息
	Name         string `gorm:"not null;size:255;index" json:"name"`                    // 文件名
	OriginalName string `gorm:"not null;size:255" json:"original_name"`                // 原始文件名
	Path         string `gorm:"not null;size:1000;unique;index" json:"path"`           // 存储路径
	Size         int64  `gorm:"not null;index" json:"size"`                            // 文件大小(字节)
	ContentType  string `gorm:"size:100" json:"content_type"`                          // MIME类型
	Extension    string `gorm:"size:10;index" json:"extension"`                        // 文件扩展名
	Description  string `gorm:"size:500" json:"description,omitempty"`                 // 文件描述
	
	// 文件标识和安全
	Hash         string `gorm:"not null;size:64;uniqueIndex" json:"hash"`              // SHA-256哈希值
	MD5          string `gorm:"size:32;index" json:"md5"`                              // MD5哈希值
	DownloadURL  string `gorm:"size:500" json:"download_url,omitempty"`                // 预签名下载URL
	
	// 所属关系
	UserID       uint   `gorm:"not null;index" json:"user_id"`                         // 所属用户ID
	FolderID     *uint  `gorm:"index" json:"folder_id,omitempty"`                      // 所属文件夹ID
	
	// 文件状态和分类
	Status       int    `gorm:"not null;default:1;index" json:"status"`                // 状态: 1=正常 2=处理中 3=已删除 4=损坏
	Category     string `gorm:"size:50;index" json:"category"`                         // 文件分类: image/video/audio/document/archive/other
	IsPublic     bool   `gorm:"default:false;index" json:"is_public"`                  // 是否公开
	
	// 版本控制
	Version      int    `gorm:"not null;default:1" json:"version"`                     // 版本号
	ParentID     *uint  `gorm:"index" json:"parent_id,omitempty"`                      // 父版本ID
	IsLatest     bool   `gorm:"default:true;index" json:"is_latest"`                   // 是否最新版本
	
	// 元数据
	Metadata     FileMetadata `gorm:"type:jsonb" json:"metadata,omitempty"`            // 文件元数据
	Tags         []string     `gorm:"type:text[]" json:"tags,omitempty"`               // 文件标签
	
	// 存储和访问统计
	StorageTier  string `gorm:"size:20;default:'hot';index" json:"storage_tier"`       // 存储层级: hot/warm/cold/archive
	DownloadCount int64 `gorm:"default:0" json:"download_count"`                       // 下载次数
	LastAccessed *int64 `json:"last_accessed,omitempty"`                              // 最后访问时间戳
	
	// 关联关系
	Folder       *Folder     `gorm:"foreignKey:FolderID;constraint:OnDelete:SET NULL" json:"folder,omitempty"`
	Versions     []File      `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE" json:"versions,omitempty"`
	Thumbnails   []Thumbnail `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"thumbnails,omitempty"`
	Shares       []Share     `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"shares,omitempty"`
}

// FileMetadata 文件元数据结构
type FileMetadata struct {
	// 图片元数据
	Width         int                    `json:"width,omitempty"`          // 图片宽度
	Height        int                    `json:"height,omitempty"`         // 图片高度
	ColorSpace    string                 `json:"color_space,omitempty"`    // 色彩空间
	
	// 视频元数据  
	Duration      float64                `json:"duration,omitempty"`       // 视频时长(秒)
	Bitrate       int64                  `json:"bitrate,omitempty"`        // 比特率
	FrameRate     float64                `json:"frame_rate,omitempty"`     // 帧率
	Resolution    string                 `json:"resolution,omitempty"`     // 分辨率
	
	// 音频元数据
	Artist        string                 `json:"artist,omitempty"`         // 艺术家
	Album         string                 `json:"album,omitempty"`          // 专辑
	Title         string                 `json:"title,omitempty"`          // 标题
	Genre         string                 `json:"genre,omitempty"`          // 流派
	
	// 地理位置信息
	Latitude      *float64               `json:"latitude,omitempty"`       // 纬度
	Longitude     *float64               `json:"longitude,omitempty"`      // 经度
	Location      string                 `json:"location,omitempty"`       // 位置描述
	
	// EXIF数据
	CameraModel   string                 `json:"camera_model,omitempty"`   // 相机型号
	LensModel     string                 `json:"lens_model,omitempty"`     // 镜头型号
	ISO           int                    `json:"iso,omitempty"`            // ISO感光度
	Aperture      string                 `json:"aperture,omitempty"`       // 光圈值
	ShutterSpeed  string                 `json:"shutter_speed,omitempty"`  // 快门速度
	FocalLength   string                 `json:"focal_length,omitempty"`   // 焦距
	
	// 文档元数据
	Author        string                 `json:"author,omitempty"`         // 作者
	Subject       string                 `json:"subject,omitempty"`        // 主题
	Keywords      []string               `json:"keywords,omitempty"`       // 关键词
	PageCount     int                    `json:"page_count,omitempty"`     // 页数
	WordCount     int                    `json:"word_count,omitempty"`     // 字数
	
	// 扩展字段
	Custom        map[string]interface{} `json:"custom,omitempty"`         // 自定义元数据
}

// 实现GORM的Valuer和Scanner接口，支持JSON存储
func (m FileMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *FileMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = FileMetadata{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	
	return json.Unmarshal(bytes, m)
}

// TableName 指定表名
func (File) TableName() string {
	return TablePrefix + "files"
}

// FileStatus 文件状态常量
const (
	FileStatusActive     = 1 // 活跃/正常 (别名: FileStatusNormal)
	FileStatusNormal     = 1 // 正常
	FileStatusProcessing = 2 // 处理中
	FileStatusDeleted    = 3 // 已删除
	FileStatusCorrupted  = 4 // 损坏
	FileStatusMerged     = 5 // 已合并(用于去重)
)

// FileCategory 文件分类常量
const (
	CategoryImage    = "image"
	CategoryVideo    = "video"  
	CategoryAudio    = "audio"
	CategoryDocument = "document"
	CategoryArchive  = "archive"
	CategoryOther    = "other"
)

// StorageTier 存储层级常量
const (
	StorageTierHot     = "hot"     // 热存储
	StorageTierWarm    = "warm"    // 温存储
	StorageTierCold    = "cold"    // 冷存储
	StorageTierArchive = "archive" // 归档存储
)

// GetCategoryByContentType 根据MIME类型获取文件分类
func GetCategoryByContentType(contentType string) string {
	switch {
	case contentType == "":
		return CategoryOther
	case contentType[:5] == "image":
		return CategoryImage
	case contentType[:5] == "video":
		return CategoryVideo
	case contentType[:5] == "audio":
		return CategoryAudio
	case contentType[:11] == "application":
		switch contentType {
		case "application/pdf", "application/msword", 
			 "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
			return CategoryDocument
		case "application/zip", "application/x-tar", "application/gzip":
			return CategoryArchive
		default:
			return CategoryDocument
		}
	default:
		return CategoryOther
	}
}

// IsImage 判断是否为图片文件
func (f *File) IsImage() bool {
	return f.Category == CategoryImage
}

// IsVideo 判断是否为视频文件
func (f *File) IsVideo() bool {
	return f.Category == CategoryVideo
}

// IsAudio 判断是否为音频文件
func (f *File) IsAudio() bool {
	return f.Category == CategoryAudio
}

// NeedsThumbnail 判断是否需要生成缩略图
func (f *File) NeedsThumbnail() bool {
	return f.IsImage() || f.IsVideo()
}

// GetDisplayName 获取显示名称
func (f *File) GetDisplayName() string {
	if f.Name != "" {
		return f.Name
	}
	return f.OriginalName
}

// GetSizeDisplay 获取可读的文件大小
func (f *File) GetSizeDisplay() string {
	size := float64(f.Size)
	units := []string{"B", "KB", "MB", "GB", "TB"}
	
	for i, unit := range units {
		if size < 1024 || i == len(units)-1 {
			return fmt.Sprintf("%.2f %s", size, unit)
		}
		size /= 1024
	}
	return fmt.Sprintf("%d B", f.Size)
}