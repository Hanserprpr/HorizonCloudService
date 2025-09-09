package handlers

import (
	"net/http"
	"strconv"
	"user-service/internal/services"
	
	"github.com/gin-gonic/gin"
)

// QuotaHandler 配额管理处理器
type QuotaHandler struct {
	userService services.UserService
}

// NewQuotaHandler 创建配额管理处理器
func NewQuotaHandler(userService services.UserService) *QuotaHandler {
	return &QuotaHandler{
		userService: userService,
	}
}

// GetQuotaInfo 获取用户配额信息
// @Summary 获取用户配额信息
// @Description 获取指定用户的存储配额详情
// @Tags 配额管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "用户ID"
// @Success 200 {object} Response{data=models.UserQuota} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 403 {object} Response "权限不足"
// @Failure 404 {object} Response "用户不存在"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/admin/quota/{user_id} [get]
func (h *QuotaHandler) GetQuotaInfo(c *gin.Context) {
	// 解析路径参数
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误", "无效的用户ID")
		return
	}
	
	// 调用服务层
	quota, err := h.userService.GetUserQuota(c.Request.Context(), uint(userID))
	if err != nil {
		if err.Error() == "用户不存在" {
			ErrorResponse(c, http.StatusNotFound, "用户不存在", err.Error())
		} else {
			ErrorResponse(c, http.StatusInternalServerError, "获取配额信息失败", err.Error())
		}
		return
	}
	
	SuccessResponse(c, http.StatusOK, "获取成功", quota)
}

// UpdateStorageUsed 更新存储使用量
// @Summary 更新存储使用量
// @Description 更新用户的存储使用量（由文件服务调用）
// @Tags 配额管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "用户ID"
// @Param request body UpdateStorageUsedRequest true "存储使用量更新请求"
// @Success 200 {object} Response "更新成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "用户不存在"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/internal/quota/{user_id}/used [put]
func (h *QuotaHandler) UpdateStorageUsed(c *gin.Context) {
	// 解析路径参数
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误", "无效的用户ID")
		return
	}
	
	var req UpdateStorageUsedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数格式错误", err.Error())
		return
	}
	
	// 调用服务层
	err = h.userService.UpdateStorageUsed(c.Request.Context(), uint(userID), req.Used)
	if err != nil {
		if err.Error() == "用户不存在" {
			ErrorResponse(c, http.StatusNotFound, "用户不存在", err.Error())
		} else if err.Error() == "存储使用量超出配额限制" {
			ErrorResponse(c, http.StatusBadRequest, "配额不足", err.Error())
		} else {
			ErrorResponse(c, http.StatusInternalServerError, "更新存储使用量失败", err.Error())
		}
		return
	}
	
	SuccessResponse(c, http.StatusOK, "更新成功", nil)
}

// BatchUpdateQuota 批量更新用户配额
// @Summary 批量更新用户配额
// @Description 管理员批量更新多个用户的存储配额
// @Tags 配额管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BatchUpdateQuotaRequest true "批量配额更新请求"
// @Success 200 {object} Response{data=BatchUpdateQuotaResponse} "更新成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 403 {object} Response "权限不足"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/admin/quota/batch-update [post]
func (h *QuotaHandler) BatchUpdateQuota(c *gin.Context) {
	var req BatchUpdateQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数格式错误", err.Error())
		return
	}
	
	// 验证请求参数
	if len(req.Updates) == 0 {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误", "更新列表不能为空")
		return
	}
	
	if len(req.Updates) > 100 {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误", "一次最多更新100个用户")
		return
	}
	
	var successCount int
	var failedUsers []BatchUpdateResult
	
	// 逐个更新用户配额
	for _, update := range req.Updates {
		err := h.userService.UpdateUserQuota(c.Request.Context(), update.UserID, update.Quota)
		if err != nil {
			failedUsers = append(failedUsers, BatchUpdateResult{
				UserID: update.UserID,
				Error:  err.Error(),
			})
		} else {
			successCount++
		}
	}
	
	response := BatchUpdateQuotaResponse{
		Total:        len(req.Updates),
		Success:      successCount,
		Failed:       len(failedUsers),
		FailedUsers:  failedUsers,
	}
	
	if len(failedUsers) == 0 {
		SuccessResponse(c, http.StatusOK, "批量更新成功", response)
	} else {
		SuccessResponse(c, http.StatusOK, "批量更新完成，部分失败", response)
	}
}

// GetQuotaStatistics 获取配额统计信息
// @Summary 获取配额统计信息
// @Description 管理员获取系统整体配额使用统计
// @Tags 配额管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response{data=QuotaStatisticsResponse} "获取成功"
// @Failure 401 {object} Response "未授权"
// @Failure 403 {object} Response "权限不足"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/admin/quota/statistics [get]
func (h *QuotaHandler) GetQuotaStatistics(c *gin.Context) {
	// 这里应该调用一个专门的统计服务，暂时返回示例数据
	// 实际实现中需要在Service层添加相应的统计方法
	
	statistics := QuotaStatisticsResponse{
		TotalUsers:        1000,
		TotalQuota:        5368709120000,   // 5TB
		TotalUsed:         2684354560000,   // 2.5TB
		AverageUsage:      50.0,
		UsersOverLimit:    5,
		UsersNearLimit:    25,
		TopUsageRanges: []UsageRange{
			{Range: "0-20%", Count: 400},
			{Range: "21-40%", Count: 300},
			{Range: "41-60%", Count: 200},
			{Range: "61-80%", Count: 75},
			{Range: "81-100%", Count: 25},
		},
	}
	
	SuccessResponse(c, http.StatusOK, "获取成功", statistics)
}

// 请求和响应结构体

// UpdateStorageUsedRequest 更新存储使用量请求
type UpdateStorageUsedRequest struct {
	Used int64 `json:"used" binding:"required,min=0"` // 使用量，字节为单位
}

// BatchUpdateQuotaRequest 批量更新配额请求
type BatchUpdateQuotaRequest struct {
	Updates []QuotaUpdate `json:"updates" binding:"required,min=1,max=100"`
}

// QuotaUpdate 配额更新项
type QuotaUpdate struct {
	UserID uint  `json:"user_id" binding:"required,min=1"`
	Quota  int64 `json:"quota" binding:"required,min=1"`
}

// BatchUpdateQuotaResponse 批量更新配额响应
type BatchUpdateQuotaResponse struct {
	Total       int                  `json:"total"`        // 总数
	Success     int                  `json:"success"`      // 成功数
	Failed      int                  `json:"failed"`       // 失败数
	FailedUsers []BatchUpdateResult  `json:"failed_users"` // 失败的用户列表
}

// BatchUpdateResult 批量更新结果
type BatchUpdateResult struct {
	UserID uint   `json:"user_id"`
	Error  string `json:"error"`
}

// QuotaStatisticsResponse 配额统计响应
type QuotaStatisticsResponse struct {
	TotalUsers        int           `json:"total_users"`         // 总用户数
	TotalQuota        int64         `json:"total_quota"`         // 总配额
	TotalUsed         int64         `json:"total_used"`          // 总使用量
	AverageUsage      float64       `json:"average_usage"`       // 平均使用率(%)
	UsersOverLimit    int           `json:"users_over_limit"`    // 超限用户数
	UsersNearLimit    int           `json:"users_near_limit"`    // 接近限制用户数(>80%)
	TopUsageRanges    []UsageRange  `json:"top_usage_ranges"`    // 使用率分布
}

// UsageRange 使用率范围
type UsageRange struct {
	Range string `json:"range"` // 范围描述
	Count int    `json:"count"` // 用户数量
}