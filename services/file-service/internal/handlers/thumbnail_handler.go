package handlers

import (
	"file-service/internal/models"
	"file-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ThumbnailHandler 缩略图处理器
type ThumbnailHandler struct {
	*BaseHandler
}

// NewThumbnailHandler 创建缩略图处理器
func NewThumbnailHandler(services *services.Services) *ThumbnailHandler {
	return &ThumbnailHandler{
		BaseHandler: NewBaseHandler(services),
	}
}

// GenerateThumbnailRequest 生成缩略图请求
type GenerateThumbnailRequest struct {
	FileID uint   `json:"file_id" binding:"required"`
	Size   string `json:"size" binding:"required,oneof=small medium large"`
}

// BatchGenerateThumbnailsRequest 批量生成缩略图请求
type BatchGenerateThumbnailsRequest struct {
	FileIDs []uint `json:"file_ids" binding:"required"`
}

// GenerateThumbnail 生成单个缩略图
func (h *ThumbnailHandler) GenerateThumbnail(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	var req GenerateThumbnailRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	// 构建服务请求
	serviceReq := &services.GenerateThumbnailRequest{
		FileID: req.FileID,
		Size:   req.Size,
		UserID: userID,
	}

	thumbnail, err := h.services.Thumbnail.GenerateThumbnail(c.Request.Context(), serviceReq)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, thumbnail)
}

// GenerateThumbnails 为文件生成所有尺寸的缩略图
func (h *ThumbnailHandler) GenerateThumbnails(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	fileID, err := h.GetIDParam(c, "file_id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	thumbnails, err := h.services.Thumbnail.GenerateThumbnails(c.Request.Context(), fileID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, thumbnails)
}

// GetThumbnail 获取缩略图详情
func (h *ThumbnailHandler) GetThumbnail(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	thumbnailID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	thumbnail, err := h.services.Thumbnail.GetThumbnail(c.Request.Context(), thumbnailID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, thumbnail)
}

// GetFileThumbnails 获取文件的所有缩略图
func (h *ThumbnailHandler) GetFileThumbnails(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	fileID, err := h.GetIDParam(c, "file_id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	thumbnails, err := h.services.Thumbnail.GetFileThumbnails(c.Request.Context(), fileID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, thumbnails)
}

// DeleteThumbnail 删除缩略图
func (h *ThumbnailHandler) DeleteThumbnail(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	thumbnailID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	err = h.services.Thumbnail.DeleteThumbnail(c.Request.Context(), thumbnailID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, gin.H{"message": "Thumbnail deleted successfully"})
}

// GetThumbnailURL 获取缩略图访问URL
func (h *ThumbnailHandler) GetThumbnailURL(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	fileID, err := h.GetIDParam(c, "file_id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	size := c.Param("size")
	if size == "" {
		size = "medium" // 默认中等尺寸
	}

	url, err := h.services.Thumbnail.GetThumbnailURL(c.Request.Context(), fileID, size, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, gin.H{"url": url})
}

// DownloadThumbnail 下载缩略图
func (h *ThumbnailHandler) DownloadThumbnail(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	fileID, err := h.GetIDParam(c, "file_id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	size := c.Param("size")
	if size == "" {
		size = "medium" // 默认中等尺寸
	}

	reader, contentType, err := h.services.Thumbnail.DownloadThumbnail(c.Request.Context(), fileID, size, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}
	defer reader.Close()

	// 设置响应头
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=86400") // 缓存24小时

	// 流式传输缩略图数据
	c.DataFromReader(http.StatusOK, -1, contentType, reader, nil)
}

// BatchGenerateThumbnails 批量生成缩略图
func (h *ThumbnailHandler) BatchGenerateThumbnails(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	var req BatchGenerateThumbnailsRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	result, err := h.services.Thumbnail.BatchGenerateThumbnails(c.Request.Context(), req.FileIDs, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, result)
}

// RefreshThumbnails 刷新缩略图
func (h *ThumbnailHandler) RefreshThumbnails(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	fileID, err := h.GetIDParam(c, "file_id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	thumbnails, err := h.services.Thumbnail.RefreshThumbnails(c.Request.Context(), fileID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, thumbnails)
}

// GetThumbnailStats 获取缩略图统计
func (h *ThumbnailHandler) GetThumbnailStats(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	stats, err := h.services.Thumbnail.GetThumbnailStats(c.Request.Context(), userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, stats)
}

// ServeThumbnail 直接服务缩略图（用于在线查看）
func (h *ThumbnailHandler) ServeThumbnail(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	fileID, err := h.GetIDParam(c, "file_id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	size := c.DefaultQuery("size", "medium")

	// 直接重定向到缩略图URL
	url, err := h.services.Thumbnail.GetThumbnailURL(c.Request.Context(), fileID, size, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	c.Redirect(http.StatusFound, url)
}

// PreviewThumbnail 预览缩略图（内联显示）
func (h *ThumbnailHandler) PreviewThumbnail(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	fileID, err := h.GetIDParam(c, "file_id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	size := c.DefaultQuery("size", "small") // 预览使用小尺寸

	reader, contentType, err := h.services.Thumbnail.DownloadThumbnail(c.Request.Context(), fileID, size, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}
	defer reader.Close()

	// 设置内联显示的响应头
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "inline")
	c.Header("Cache-Control", "public, max-age=86400")

	// 流式传输缩略图数据
	c.DataFromReader(http.StatusOK, -1, contentType, reader, nil)
}

// GetThumbnailInfo 获取缩略图信息（不下载内容）
func (h *ThumbnailHandler) GetThumbnailInfo(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	fileID, err := h.GetIDParam(c, "file_id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	size := c.DefaultQuery("size", "medium")

	thumbnails, err := h.services.Thumbnail.GetFileThumbnails(c.Request.Context(), fileID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	// 查找指定尺寸的缩略图
	var targetThumbnail *models.Thumbnail
	for _, thumbnail := range thumbnails {
		if thumbnail.Size == size {
			// 这里需要转换类型，实际实现时需要调整
			break
		}
	}

	if targetThumbnail == nil {
		h.NotFound(c, "Thumbnail not found for specified size", nil)
		return
	}

	h.Success(c, targetThumbnail)
}