package handlers

import (
	"file-service/internal/services"
)

// Handlers 处理器容器
type Handlers struct {
	File      *FileHandler
	Folder    *FolderHandler
	Upload    *UploadHandler
	Thumbnail *ThumbnailHandler
	Health    *HealthHandler
}

// NewHandlers 创建处理器容器
func NewHandlers(services *services.Services) *Handlers {
	return &Handlers{
		File:      NewFileHandler(services),
		Folder:    NewFolderHandler(services),
		Upload:    NewUploadHandler(services),
		Thumbnail: NewThumbnailHandler(services),
		Health:    NewHealthHandler(services),
	}
}