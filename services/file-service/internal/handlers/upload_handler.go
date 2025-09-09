package handlers

import (
	"file-service/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UploadHandler 上传处理器
type UploadHandler struct {
	*BaseHandler
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(services *services.Services) *UploadHandler {
	return &UploadHandler{
		BaseHandler: NewBaseHandler(services),
	}
}

// InitiateUploadRequest 初始化上传请求
type InitiateUploadRequest struct {
	FileName    string            `json:"file_name" binding:"required"`
	Size        int64             `json:"size" binding:"required,min=1"`
	ContentType string            `json:"content_type"`
	FolderID    *uint             `json:"folder_id"`
	ChunkSize   int64             `json:"chunk_size,omitempty"`
	Metadata    map[string]string `json:"metadata"`
}

// UploadChunkRequest 上传分片请求
type UploadChunkRequest struct {
	SessionID  string `form:"session_id" binding:"required"`
	ChunkIndex int    `form:"chunk_index" binding:"min=0"`
	ChunkHash  string `form:"chunk_hash"`
}

// BatchInitiateRequest 批量初始化上传请求
type BatchInitiateRequest struct {
	Files []*InitiateUploadRequest `json:"files" binding:"required"`
}

// ResumeUploadRequest 断点续传请求
type ResumeUploadRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// InitiateUpload 初始化上传
func (h *UploadHandler) InitiateUpload(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	var req InitiateUploadRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	// 构建服务请求
	serviceReq := &services.InitiateUploadRequest{
		FileName:    req.FileName,
		Size:        req.Size,
		ContentType: req.ContentType,
		FolderID:    req.FolderID,
		ChunkSize:   req.ChunkSize,
		UserID:      userID,
		Metadata:    req.Metadata,
	}

	result, err := h.services.Upload.InitiateUpload(c.Request.Context(), serviceReq)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, result)
}

// UploadChunk 上传分片
func (h *UploadHandler) UploadChunk(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	var req UploadChunkRequest
	if err := c.ShouldBind(&req); err != nil {
		h.BadRequest(c, "Invalid request parameters", err)
		return
	}

	// 获取文件内容
	file, header, err := c.Request.FormFile("chunk")
	if err != nil {
		h.BadRequest(c, "Missing chunk data", err)
		return
	}
	defer file.Close()

	// 构建服务请求
	serviceReq := &services.UploadChunkRequest{
		SessionID:  req.SessionID,
		ChunkIndex: req.ChunkIndex,
		ChunkSize:  header.Size,
		ChunkData:  file,
		ChunkHash:  req.ChunkHash,
		UserID:     userID,
	}

	result, err := h.services.Upload.UploadChunk(c.Request.Context(), serviceReq)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, result)
}

// CompleteUpload 完成上传
func (h *UploadHandler) CompleteUpload(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		h.BadRequest(c, "Missing session_id parameter", nil)
		return
	}

	result, err := h.services.Upload.CompleteUpload(c.Request.Context(), sessionID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, result)
}

// AbortUpload 中止上传
func (h *UploadHandler) AbortUpload(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		h.BadRequest(c, "Missing session_id parameter", nil)
		return
	}

	err := h.services.Upload.AbortUpload(c.Request.Context(), sessionID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, gin.H{"message": "Upload aborted successfully"})
}

// GetUploadSession 获取上传会话信息
func (h *UploadHandler) GetUploadSession(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		h.BadRequest(c, "Missing session_id parameter", nil)
		return
	}

	session, err := h.services.Upload.GetUploadSession(c.Request.Context(), sessionID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, session)
}

// ListUploadSessions 获取上传会话列表
func (h *UploadHandler) ListUploadSessions(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	var req PaginationRequest
	if err := h.BindQuery(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	status := h.GetQueryParamInt(c, "status", 0)
	var statusPtr *int
	if status != 0 {
		statusPtr = &status
	}

	offset := (req.Page - 1) * req.PageSize
	sessions, total, err := h.services.Upload.ListUploadSessions(c.Request.Context(), userID, statusPtr, offset, req.PageSize)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Paginate(c, sessions, total, &req)
}

// GetUploadProgress 获取上传进度
func (h *UploadHandler) GetUploadProgress(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		h.BadRequest(c, "Missing session_id parameter", nil)
		return
	}

	progress, err := h.services.Upload.GetUploadProgress(c.Request.Context(), sessionID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, progress)
}

// BatchInitiateUpload 批量初始化上传
func (h *UploadHandler) BatchInitiateUpload(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	var req BatchInitiateRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleServiceError(c, err)
		return
	}

	// 转换请求
	serviceFiles := make([]*services.InitiateUploadRequest, len(req.Files))
	for i, fileReq := range req.Files {
		serviceFiles[i] = &services.InitiateUploadRequest{
			FileName:    fileReq.FileName,
			Size:        fileReq.Size,
			ContentType: fileReq.ContentType,
			FolderID:    fileReq.FolderID,
			ChunkSize:   fileReq.ChunkSize,
			UserID:      userID,
			Metadata:    fileReq.Metadata,
		}
	}

	serviceReq := &services.BatchInitiateUploadRequest{
		Files:  serviceFiles,
		UserID: userID,
	}

	result, err := h.services.Upload.BatchInitiateUpload(c.Request.Context(), serviceReq)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, result)
}

// ResumeUpload 断点续传
func (h *UploadHandler) ResumeUpload(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		h.BadRequest(c, "Missing session_id parameter", nil)
		return
	}

	result, err := h.services.Upload.ResumeUpload(c.Request.Context(), sessionID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, result)
}

// PauseUpload 暂停上传
func (h *UploadHandler) PauseUpload(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		h.BadRequest(c, "Missing session_id parameter", nil)
		return
	}

	err := h.services.Upload.PauseUpload(c.Request.Context(), sessionID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, gin.H{"message": "Upload paused successfully"})
}

// ResumeUploadFromPause 从暂停状态恢复上传
func (h *UploadHandler) ResumeUploadFromPause(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		h.BadRequest(c, "Missing session_id parameter", nil)
		return
	}

	err := h.services.Upload.ResumeUploadFromPause(c.Request.Context(), sessionID, userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, gin.H{"message": "Upload resumed successfully"})
}

// GetUploadStatistics 获取上传统计
func (h *UploadHandler) GetUploadStatistics(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	stats, err := h.services.Upload.GetUploadStatistics(c.Request.Context(), userID)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, stats)
}

// SimpleUpload 简单上传接口（兼容性）
func (h *UploadHandler) SimpleUpload(c *gin.Context) {
	userID := h.RequireAuth(c)
	if userID == 0 {
		return
	}

	// 获取表单数据
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.BadRequest(c, "Missing file", err)
		return
	}
	defer file.Close()

	folderID := c.PostForm("folder_id")
	var folderIDPtr *uint
	if folderID != "" {
		if id, err := strconv.ParseUint(folderID, 10, 32); err == nil {
			folderIDUint := uint(id)
			folderIDPtr = &folderIDUint
		}
	}

	// 构建上传请求
	serviceReq := &services.UploadFileRequest{
		FileName:    header.Filename,
		Size:        header.Size,
		ContentType: header.Header.Get("Content-Type"),
		FolderID:    folderIDPtr,
		UserID:      userID,
		Reader:      file,
	}

	result, err := h.services.File.UploadFile(c.Request.Context(), serviceReq)
	if err != nil {
		h.HandleServiceError(c, err)
		return
	}

	h.Success(c, result)
}