package handlers

import (
	"file-service/internal/models"
	"file-service/internal/repository"
	"file-service/internal/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// FileHandler 文件处理器
type FileHandler struct {
	*BaseHandler
}

// NewFileHandler 创建文件处理器
func NewFileHandler(services *services.Services) *FileHandler {
	return &FileHandler{
		BaseHandler: NewBaseHandler(services),
	}
}

// UploadFileRequest 文件上传请求
type UploadFileRequest struct {
	FileName    string            `json:"file_name" binding:"required"`
	Size        int64             `json:"size" binding:"required,min=1"`
	ContentType string            `json:"content_type"`
	FolderID    *uint             `json:"folder_id"`
	Metadata    map[string]string `json:"metadata"`
}

// FileListRequest 文件列表请求
type FileListRequest struct {
	PaginationRequest
	FolderID    *uint  `form:"folder_id"`
	Category    string `form:"category"`
	Keywords    string `form:"keywords"`
	SortBy      string `form:"sort_by,default=created_at"`
	SortOrder   string `form:"sort_order,default=desc"`
	StorageTier string `form:"storage_tier"`
	Status      int    `form:"status"`
}

// FileSearchRequest 文件搜索请求
type FileSearchRequest struct {
	PaginationRequest
	Query       string `form:"q" binding:"required"`
	Category    string `form:"category"`
	StorageTier string `form:"storage_tier"`
	DateFrom    string `form:"date_from"`
	DateTo      string `form:"date_to"`
	SizeMin     int64  `form:"size_min"`
	SizeMax     int64  `form:"size_max"`
}

// FileUpdateRequest 文件更新请求
type FileUpdateRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	IsPublic    *bool             `json:"is_public"`
	Metadata    map[string]string `json:"metadata"`
}

// FileMoveRequest 文件移动请求
type FileMoveRequest struct {
	FolderID uint `json:"folder_id" binding:"required"`
}

// FileBatchRequest 批量操作请求
type FileBatchRequest struct {
	FileIDs []uint `json:"file_ids" binding:"required"`
	Action  string `json:"action" binding:"required,oneof=move delete copy"`
	Target  uint   `json:"target"` // 目标文件夹ID（移动和复制时使用）
}

// UploadFile 上传文件
func (h *FileHandler) UploadFile(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	var req UploadFileRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	// 调用服务层上传文件
	serviceReq := &services.UploadFileRequest{
		FileName:    req.FileName,
		Size:        req.Size,
		ContentType: req.ContentType,
		FolderID:    req.FolderID,
		UserID:      userID,
		Metadata:    req.Metadata,
		Reader:      c.Request.Body,
	}

	result, err := h.services.File.UploadFile(c.Request.Context(), serviceReq)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, result)
}

// GetFile 获取文件详情
func (h *FileHandler) GetFile(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	fileID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	file, err := h.services.File.GetFile(c.Request.Context(), fileID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, file)
}

// ListFiles 获取文件列表
func (h *FileHandler) ListFiles(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	var req FileListRequest
	if err := h.BindQuery(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	// 构建过滤器
	status := models.FileStatusActive
	filters := &repository.FileFilters{
		Category:    req.Category,
		StorageTier: req.StorageTier,
		Status:      &status,
	}

	if req.Status != 0 {
		filters.Status = &req.Status
	}

	offset := (req.Page - 1) * req.PageSize

	var files []*models.File
	var total int64
	var err error

	if req.Keywords != "" {
		// 搜索模式
		searchReq := &services.SearchFilesRequest{
			UserID:  userID,
			Keyword: req.Keywords,
			Offset:  offset,
			Limit:   req.PageSize,
			Filters: filters,
		}
		searchResp, err := h.services.File.SearchFiles(c.Request.Context(), searchReq)
		if err == nil {
			files = searchResp.Files
			total = searchResp.Total
		}
	} else {
		// 列表模式
		listReq := &services.ListFilesRequest{
			UserID:   userID,
			FolderID: req.FolderID,
			Offset:   offset,
			Limit:    req.PageSize,
			Filters:  filters,
		}
		listResp, err := h.services.File.ListFiles(c.Request.Context(), listReq)
		if err == nil {
			files = listResp.Files
			total = listResp.Total
		}
	}

	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Paginate(c, files, total, &req.PaginationRequest)
}

// SearchFiles 搜索文件
func (h *FileHandler) SearchFiles(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	var req FileSearchRequest
	if err := h.BindQuery(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	// 构建过滤器
	status := models.FileStatusActive
	filters := &repository.FileFilters{
		Category:    req.Category,
		StorageTier: req.StorageTier,
		DateFrom:    &req.DateFrom,
		DateTo:      &req.DateTo,
		MinSize:     &req.SizeMin,
		MaxSize:     &req.SizeMax,
		Status:      &status,
	}

	offset := (req.Page - 1) * req.PageSize
	searchReq := &services.SearchFilesRequest{
		UserID:  userID,
		Keyword: req.Query,
		Offset:  offset,
		Limit:   req.PageSize,
		Filters: filters,
	}
	searchResp, err := h.services.File.SearchFiles(c.Request.Context(), searchReq)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	files := searchResp.Files
	total := searchResp.Total

	h.Paginate(c, files, total, &req.PaginationRequest)
}

// UpdateFile 更新文件信息
func (h *FileHandler) UpdateFile(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	fileID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	var req FileUpdateRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	// 构建更新请求
	serviceReq := &services.UpdateFileRequest{
		FileID:      fileID,
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Tags:        req.Tags,
		IsPublic:    req.IsPublic,
		Metadata:    req.Metadata,
	}

	err = h.services.File.UpdateFile(c.Request.Context(), serviceReq)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, gin.H{"message": "File updated successfully"})
}

// DeleteFile 删除文件
func (h *FileHandler) DeleteFile(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	fileID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	err = h.services.File.DeleteFile(c.Request.Context(), fileID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, gin.H{"message": "File deleted successfully"})
}

// DownloadFile 下载文件
func (h *FileHandler) DownloadFile(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	fileID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	// 获取下载URL或直接流式传输
	downloadURL, err := h.services.File.GetDownloadURL(c.Request.Context(), fileID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	// 重定向到预签名URL
	c.Redirect(http.StatusFound, downloadURL)
}

// MoveFile 移动文件
func (h *FileHandler) MoveFile(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	fileID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	var req FileMoveRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	err = h.services.File.MoveFile(c.Request.Context(), fileID, req.FolderID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, gin.H{"message": "File moved successfully"})
}

// CopyFile 复制文件
func (h *FileHandler) CopyFile(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	fileID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	var req FileMoveRequest // 复用相同结构
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	copiedFile, err := h.services.File.CopyFile(c.Request.Context(), fileID, req.FolderID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, copiedFile)
}

// BatchOperation 批量操作
func (h *FileHandler) BatchOperation(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	var req FileBatchRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	var result *services.BatchOperationResponse
	var err error

	switch req.Action {
	case "move":
		result, err = h.services.File.BatchMoveFiles(c.Request.Context(), req.FileIDs, req.Target, userID)
	case "delete":
		result, err = h.services.File.BatchDeleteFiles(c.Request.Context(), req.FileIDs, userID)
	case "copy":
		result, err = h.services.File.BatchCopyFiles(c.Request.Context(), req.FileIDs, req.Target, userID)
	default:
		h.BadRequest(c, "Unsupported batch action", nil)
		return
	}

	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, result)
}

// GetFileVersions 获取文件版本列表
func (h *FileHandler) GetFileVersions(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	fileID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	versions, err := h.services.Version.GetFileVersions(c.Request.Context(), fileID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, versions)
}

// RestoreFileVersion 恢复文件版本
func (h *FileHandler) RestoreFileVersion(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	versionID, err := h.GetIDParam(c, "version_id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	restoredFile, err := h.services.Version.RestoreVersion(c.Request.Context(), versionID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, restoredFile)
}

// GetDuplicateFiles 获取重复文件
func (h *FileHandler) GetDuplicateFiles(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	folderID := h.GetQueryParam(c, "folder_id")
	var folderIDPtr *uint
	if folderID != "" {
		if id, err := strconv.ParseUint(folderID, 10, 32); err == nil {
			folderIDUint := uint(id)
			folderIDPtr = &folderIDUint
		}
	}

	duplicates, err := h.services.Version.FindDuplicateFiles(c.Request.Context(), userID, folderIDPtr)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, duplicates)
}

// CleanupDuplicates 清理重复文件
func (h *FileHandler) CleanupDuplicates(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	strategyParam := h.GetQueryParam(c, "strategy")
	strategy := services.DuplicateCleanupKeepNewest // 默认策略

	switch strings.ToLower(strategyParam) {
	case "oldest":
		strategy = services.DuplicateCleanupKeepOldest
	case "newest":
		strategy = services.DuplicateCleanupKeepNewest
	case "largest_name":
		strategy = services.DuplicateCleanupKeepLargestName
	}

	result, err := h.services.Version.CleanupDuplicates(c.Request.Context(), userID, strategy)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, result)
}

// GetUserStats 获取用户文件统计
func (h *FileHandler) GetUserStats(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	stats, err := h.services.File.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, stats)
}

// GetStorageStats 获取存储统计
func (h *FileHandler) GetStorageStats(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	stats, err := h.services.File.GetStorageStats(c.Request.Context(), userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, stats)
}

// GetAdminFileStats 获取管理员级别的文件统计（所有用户）
func (h *FileHandler) GetAdminFileStats(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	// 这里应该检查管理员权限，但目前简化处理
	// TODO: 添加管理员权限验证

	stats, err := h.services.File.GetAdminFileStats(c.Request.Context())
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, stats)
}