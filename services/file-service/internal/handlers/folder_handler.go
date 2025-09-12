package handlers

import (
	"file-service/internal/services"

	"github.com/gin-gonic/gin"
)

// FolderHandler 文件夹处理器
type FolderHandler struct {
	*BaseHandler
}

// NewFolderHandler 创建文件夹处理器
func NewFolderHandler(services *services.Services) *FolderHandler {
	return &FolderHandler{
		BaseHandler: NewBaseHandler(services),
	}
}

// CreateFolderRequest 创建文件夹请求
type CreateFolderRequest struct {
	Name        string `json:"name" binding:"required"`
	ParentID    *uint  `json:"parent_id"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
}

// UpdateFolderRequest 更新文件夹请求
type UpdateFolderRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
}

// MoveFolderRequest 移动文件夹请求
type MoveFolderRequest struct {
	NewParentID uint `json:"new_parent_id" binding:"required"`
}

// GetFolderContentsRequest 获取文件夹内容请求
type GetFolderContentsRequest struct {
	PaginationRequest
	FolderID *uint `form:"folder_id"`
}

// CreateFolder 创建文件夹
func (h *FolderHandler) CreateFolder(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	var req CreateFolderRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	// 构建服务请求
	serviceReq := &services.CreateFolderRequest{
		Name:        req.Name,
		ParentID:    req.ParentID,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		UserID:      userID,
	}

	folder, err := h.services.Folder.CreateFolder(c.Request.Context(), serviceReq)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, folder)
}

// GetFolder 获取文件夹详情
func (h *FolderHandler) GetFolder(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	folderID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	folder, err := h.services.Folder.GetFolder(c.Request.Context(), folderID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, folder)
}

func (h *FolderHandler) GetFolderRecommend(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	folderID := uint(0)

	folder, err := h.services.Folder.GetFolder(c.Request.Context(), folderID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, folder)
}

// GetFolderByPath 根据路径获取文件夹
func (h *FolderHandler) GetFolderByPath(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	path := c.Query("path")
	if path == "" {
		h.BadRequest(c, "Missing path parameter", nil)
		return
	}

	folder, err := h.services.Folder.GetFolderByPath(c.Request.Context(), path, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, folder)
}

// UpdateFolder 更新文件夹
func (h *FolderHandler) UpdateFolder(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	folderID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	var req UpdateFolderRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	// 构建服务请求
	serviceReq := &services.UpdateFolderRequest{
		FolderID:    folderID,
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
	}

	err = h.services.Folder.UpdateFolder(c.Request.Context(), serviceReq)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, gin.H{"message": "Folder updated successfully"})
}

// DeleteFolder 删除文件夹
func (h *FolderHandler) DeleteFolder(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	folderID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	err = h.services.Folder.DeleteFolder(c.Request.Context(), folderID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, gin.H{"message": "Folder deleted successfully"})
}

// ListFolders 获取文件夹列表
func (h *FolderHandler) ListFolders(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	parentIDStr := c.Query("parent_id")
	var parentID *uint
	if parentIDStr != "" {
		if id, err := h.GetIDParam(c, "parent_id"); err == nil {
			parentID = &id
		}
	}

	folders, err := h.services.Folder.ListFolders(c.Request.Context(), userID, parentID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, folders)
}

// GetFolderTree 获取文件夹树
func (h *FolderHandler) GetFolderTree(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	rootIDStr := c.Query("root_id")
	var rootID *uint
	if rootIDStr != "" {
		if id, err := h.GetIDParam(c, "root_id"); err == nil {
			rootID = &id
		}
	}

	tree, err := h.services.Folder.GetFolderTree(c.Request.Context(), userID, rootID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, tree)
}

// GetFolderPath 获取文件夹路径
func (h *FolderHandler) GetFolderPath(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	folderID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	path, err := h.services.Folder.GetFolderPath(c.Request.Context(), folderID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, path)
}

// MoveFolder 移动文件夹
func (h *FolderHandler) MoveFolder(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	folderID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	var req MoveFolderRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	err = h.services.Folder.MoveFolder(c.Request.Context(), folderID, req.NewParentID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, gin.H{"message": "Folder moved successfully"})
}

// CopyFolder 复制文件夹
func (h *FolderHandler) CopyFolder(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	folderID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	var req MoveFolderRequest // 复用相同结构
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	copiedFolder, err := h.services.Folder.CopyFolder(c.Request.Context(), folderID, req.NewParentID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, copiedFolder)
}

// RenameFolder 重命名文件夹
func (h *FolderHandler) RenameFolder(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	folderID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	newName := c.PostForm("name")
	if newName == "" {
		h.BadRequest(c, "Missing new name", nil)
		return
	}

	err = h.services.Folder.RenameFolder(c.Request.Context(), folderID, newName, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, gin.H{"message": "Folder renamed successfully"})
}

// GetFolderContents 获取文件夹内容
func (h *FolderHandler) GetFolderContents(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	// 从URL路径参数获取文件夹ID
	folderIDStr := c.Param("id")
	var folderIDPtr *uint
	if folderIDStr != "" && folderIDStr != "0" {
		if folderID, err := h.GetIDParam(c, "id"); err == nil && folderID > 0 {
			folderIDPtr = &folderID
		}
	}

	var req GetFolderContentsRequest
	if err := h.BindQuery(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	// 构建服务请求
	serviceReq := &services.GetFolderContentsRequest{
		FolderID: folderIDPtr,
		UserID:   userID,
		Offset:   (req.Page - 1) * req.PageSize,
		Limit:    req.PageSize,
	}

	contents, err := h.services.Folder.GetFolderContents(c.Request.Context(), serviceReq)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, contents)
}

// GetFolderStats 获取文件夹统计
func (h *FolderHandler) GetFolderStats(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	folderID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	stats, err := h.services.Folder.GetFolderStats(c.Request.Context(), folderID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, stats)
}

// CreateSystemFolders 创建系统文件夹
func (h *FolderHandler) CreateSystemFolders(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	err := h.services.Folder.CreateSystemFolders(c.Request.Context(), userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, gin.H{"message": "System folders created successfully"})
}

// GetSystemFolder 获取系统文件夹
func (h *FolderHandler) GetSystemFolder(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	folderType := c.Param("type")
	if folderType == "" {
		h.BadRequest(c, "Missing folder type", nil)
		return
	}

	folder, err := h.services.Folder.GetSystemFolder(c.Request.Context(), userID, folderType)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, folder)
}

// SyncFolderStats 同步文件夹统计
func (h *FolderHandler) SyncFolderStats(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	folderID, err := h.GetIDParam(c, "id")
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	err = h.services.Folder.SyncFolderStats(c.Request.Context(), folderID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, gin.H{"message": "Folder statistics synchronized successfully"})
}

// RecalculateAllStats 重新计算用户所有文件夹统计
func (h *FolderHandler) RecalculateAllStats(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	err := h.services.Folder.RecalculateAllStats(c.Request.Context(), userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, gin.H{"message": "All folder statistics recalculated successfully"})
}

// SearchFolders 搜索文件夹
func (h *FolderHandler) SearchFolders(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	var req struct {
		PaginationRequest
		Keyword string `form:"q" binding:"required"`
	}

	if err := h.BindQuery(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	offset := (req.Page - 1) * req.PageSize
	folders, total, err := h.services.Folder.SearchFolders(c.Request.Context(), userID, req.Keyword, offset, req.PageSize)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Paginate(c, folders, total, &req.PaginationRequest)
}
