package models

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

// Share 文件分享模型
type Share struct {
	BaseModel
	
	// 基本信息
	Token       string `gorm:"not null;uniqueIndex;size:64" json:"token"`              // 分享令牌
	Name        string `gorm:"size:255" json:"name,omitempty"`                         // 自定义分享名称
	Description string `gorm:"size:500" json:"description,omitempty"`                  // 分享描述
	
	// 关联资源
	FileID      *uint  `gorm:"index" json:"file_id,omitempty"`                         // 分享的文件ID
	FolderID    *uint  `gorm:"index" json:"folder_id,omitempty"`                       // 分享的文件夹ID
	
	// 所属用户
	UserID      uint   `gorm:"not null;index" json:"user_id"`                          // 创建者用户ID
	
	// 访问控制
	Password    string `gorm:"size:255" json:"password,omitempty"`                     // 访问密码(加密存储)
	MaxDownloads *int  `json:"max_downloads,omitempty"`                               // 最大下载次数
	AllowPreview bool  `gorm:"default:true" json:"allow_preview"`                     // 允许预览
	AllowDownload bool `gorm:"default:true" json:"allow_download"`                    // 允许下载
	
	// 时间控制
	ExpiresAt   *time.Time `gorm:"index" json:"expires_at,omitempty"`                  // 过期时间
	
	// 状态
	Status      int    `gorm:"not null;default:1;index" json:"status"`                 // 状态: 1=活跃 2=禁用 3=过期
	IsPublic    bool   `gorm:"default:false;index" json:"is_public"`                   // 是否公开(搜索引擎可索引)
	
	// 统计信息
	ViewCount   int64  `gorm:"default:0" json:"view_count"`                            // 浏览次数
	DownloadCount int64 `gorm:"default:0" json:"download_count"`                       // 下载次数
	LastAccessed *time.Time `json:"last_accessed,omitempty"`                          // 最后访问时间
	
	// 关联关系
	File        *File   `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"file,omitempty"`
	Folder      *Folder `gorm:"foreignKey:FolderID;constraint:OnDelete:CASCADE" json:"folder,omitempty"`
	AccessLogs  []ShareAccessLog `gorm:"foreignKey:ShareID;constraint:OnDelete:CASCADE" json:"access_logs,omitempty"`
}

// ShareAccessLog 分享访问日志
type ShareAccessLog struct {
	BaseModel
	
	// 关联信息
	ShareID     uint   `gorm:"not null;index" json:"share_id"`                         // 分享ID
	
	// 访问信息
	IPAddress   string `gorm:"not null;size:45;index" json:"ip_address"`               // 访问IP
	UserAgent   string `gorm:"size:500" json:"user_agent,omitempty"`                   // User Agent
	Referer     string `gorm:"size:500" json:"referer,omitempty"`                      // 来源页面
	
	// 操作信息
	Action      string `gorm:"not null;size:20;index" json:"action"`                   // 操作类型: view/download/preview
	Success     bool   `gorm:"not null;index" json:"success"`                          // 操作是否成功
	ErrorMessage string `gorm:"size:500" json:"error_message,omitempty"`               // 错误消息
	
	// 地理位置(可选)
	Country     string `gorm:"size:50" json:"country,omitempty"`                       // 国家
	Region      string `gorm:"size:50" json:"region,omitempty"`                        // 地区
	City        string `gorm:"size:50" json:"city,omitempty"`                          // 城市
	
	// 关联关系
	Share       *Share `gorm:"foreignKey:ShareID;constraint:OnDelete:CASCADE" json:"share,omitempty"`
}

// TableName 指定表名
func (Share) TableName() string {
	return TablePrefix + "shares"
}

// TableName 指定表名
func (ShareAccessLog) TableName() string {
	return TablePrefix + "share_access_logs"
}

// ShareStatus 分享状态常量
const (
	ShareStatusActive   = 1 // 活跃
	ShareStatusDisabled = 2 // 禁用
	ShareStatusExpired  = 3 // 过期
)

// ShareAction 分享操作常量
const (
	ShareActionView     = "view"     // 查看
	ShareActionDownload = "download" // 下载
	ShareActionPreview  = "preview"  // 预览
)

// GenerateToken 生成分享令牌
func GenerateShareToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// IsExpired 检查分享是否过期
func (s *Share) IsExpired() bool {
	if s.Status == ShareStatusExpired {
		return true
	}
	if s.ExpiresAt != nil && time.Now().After(*s.ExpiresAt) {
		return true
	}
	return false
}

// IsActive 检查分享是否活跃
func (s *Share) IsActive() bool {
	return s.Status == ShareStatusActive && !s.IsExpired()
}

// HasPassword 检查是否设置了密码
func (s *Share) HasPassword() bool {
	return s.Password != ""
}

// IsDownloadLimitReached 检查下载限制是否达到
func (s *Share) IsDownloadLimitReached() bool {
	if s.MaxDownloads == nil {
		return false
	}
	return s.DownloadCount >= int64(*s.MaxDownloads)
}

// CanDownload 检查是否可以下载
func (s *Share) CanDownload() bool {
	return s.IsActive() && 
		   s.AllowDownload && 
		   !s.IsDownloadLimitReached()
}

// CanPreview 检查是否可以预览
func (s *Share) CanPreview() bool {
	return s.IsActive() && s.AllowPreview
}

// IncrementView 增加浏览次数
func (s *Share) IncrementView() {
	s.ViewCount++
	now := time.Now()
	s.LastAccessed = &now
}

// IncrementDownload 增加下载次数
func (s *Share) IncrementDownload() {
	s.DownloadCount++
	now := time.Now()
	s.LastAccessed = &now
}

// GetShareURL 获取分享URL
func (s *Share) GetShareURL(baseURL string) string {
	return baseURL + "/s/" + s.Token
}

// GetResourceType 获取资源类型
func (s *Share) GetResourceType() string {
	if s.FileID != nil {
		return "file"
	}
	if s.FolderID != nil {
		return "folder"
	}
	return "unknown"
}

// GetDisplayName 获取显示名称
func (s *Share) GetDisplayName() string {
	if s.Name != "" {
		return s.Name
	}
	if s.File != nil {
		return s.File.GetDisplayName()
	}
	if s.Folder != nil {
		return s.Folder.GetDisplayName()
	}
	return "未命名分享"
}