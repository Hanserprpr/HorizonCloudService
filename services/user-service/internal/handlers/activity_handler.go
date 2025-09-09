package handlers

import (
	"net/http"
	"strconv"
	"time"
	"user-service/internal/services"
	
	"github.com/gin-gonic/gin"
)

// ActivityHandler 活动日志处理器
type ActivityHandler struct {
	userService services.UserService
}

// NewActivityHandler 创建活动日志处理器
func NewActivityHandler(userService services.UserService) *ActivityHandler {
	return &ActivityHandler{
		userService: userService,
	}
}

// LogActivity 记录用户活动
// @Summary 记录用户活动
// @Description 记录用户的操作活动（内部服务调用）
// @Tags 活动日志
// @Accept json
// @Produce json
// @Param request body services.LogActivityRequest true "活动记录请求"
// @Success 201 {object} Response "记录成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/internal/activity/log [post]
func (h *ActivityHandler) LogActivity(c *gin.Context) {
	var req services.LogActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数格式错误", err.Error())
		return
	}
	
	// 如果没有提供IP地址和User-Agent，从请求中获取
	if req.IPAddress == "" {
		req.IPAddress = c.ClientIP()
	}
	if req.UserAgent == "" {
		req.UserAgent = c.GetHeader("User-Agent")
	}
	
	// 调用服务层
	err := h.userService.LogActivity(c.Request.Context(), &req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "记录活动失败", err.Error())
		return
	}
	
	SuccessResponse(c, http.StatusCreated, "记录成功", nil)
}

// GetUserActivityLogs 获取用户活动日志
// @Summary 获取用户活动日志
// @Description 获取指定用户的活动历史记录
// @Tags 活动日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "用户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param action query string false "操作类型筛选"
// @Param start_date query string false "开始日期 (YYYY-MM-DD)"
// @Param end_date query string false "结束日期 (YYYY-MM-DD)"
// @Success 200 {object} Response{data=services.GetActivityLogsResponse} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 403 {object} Response "权限不足"
// @Failure 404 {object} Response "用户不存在"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/admin/activity/{user_id}/logs [get]
func (h *ActivityHandler) GetUserActivityLogs(c *gin.Context) {
	// 解析路径参数
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误", "无效的用户ID")
		return
	}
	
	// 解析查询参数
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	
	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}
	
	req := &services.GetActivityLogsRequest{
		UserID:   uint(userID),
		Page:     page,
		PageSize: pageSize,
	}
	
	// 调用服务层
	result, err := h.userService.GetUserActivityLogs(c.Request.Context(), req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "获取活动日志失败", err.Error())
		return
	}
	
	SuccessResponse(c, http.StatusOK, "获取成功", result)
}

// GetSystemActivityLogs 获取系统活动日志
// @Summary 获取系统活动日志
// @Description 管理员获取系统整体活动日志
// @Tags 活动日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param action query string false "操作类型筛选"
// @Param user_id query int false "用户ID筛选"
// @Param start_date query string false "开始日期 (YYYY-MM-DD)"
// @Param end_date query string false "结束日期 (YYYY-MM-DD)"
// @Success 200 {object} Response{data=SystemActivityLogsResponse} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 403 {object} Response "权限不足"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/admin/activity/system [get]
func (h *ActivityHandler) GetSystemActivityLogs(c *gin.Context) {
	// 解析查询参数
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	
	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}
	
	var userID *uint
	if uid := c.Query("user_id"); uid != "" {
		if parsed, err := strconv.ParseUint(uid, 10, 32); err == nil {
			id := uint(parsed)
			userID = &id
		}
	}
	
	// 这里应该调用一个专门的系统日志查询服务
	// 暂时返回示例数据，实际实现需要在Service层添加相应方法
	// userID可以用于筛选特定用户的系统日志
	_ = userID // 避免未使用变量警告
	
	response := SystemActivityLogsResponse{
		Logs: []*SystemActivityLogItem{
			{
				ID:        1,
				UserID:    1,
				StudentID: "2021001",
				NickName:  "张三",
				Action:    "login",
				Resource:  "user",
				Detail:    "用户登录",
				IPAddress: "192.168.1.100",
				UserAgent: "Mozilla/5.0...",
				CreatedAt: time.Now().Unix(),
			},
			{
				ID:        2,
				UserID:    2,
				StudentID: "2021002",
				NickName:  "李四",
				Action:    "upload_file",
				Resource:  "file",
				Detail:    "上传文件: example.jpg",
				IPAddress: "192.168.1.101",
				UserAgent: "Mozilla/5.0...",
				CreatedAt: time.Now().Unix() - 3600,
			},
		},
		Total:    2,
		Page:     page,
		PageSize: pageSize,
		Pages:    1,
	}
	
	SuccessResponse(c, http.StatusOK, "获取成功", response)
}

// GetActivityStatistics 获取活动统计信息
// @Summary 获取活动统计信息
// @Description 管理员获取活动操作的统计分析
// @Tags 活动日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period query string false "统计周期" Enums(day, week, month) default(week)
// @Success 200 {object} Response{data=ActivityStatisticsResponse} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 403 {object} Response "权限不足"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/admin/activity/statistics [get]
func (h *ActivityHandler) GetActivityStatistics(c *gin.Context) {
	period := c.DefaultQuery("period", "week")
	
	// 验证统计周期
	if period != "day" && period != "week" && period != "month" {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误", "无效的统计周期")
		return
	}
	
	// 这里应该调用一个专门的统计服务
	// 暂时返回示例数据，实际实现需要在Service层添加相应的统计方法
	
	response := ActivityStatisticsResponse{
		Period:       period,
		TotalActions: 10250,
		UniqueUsers:  856,
		TopActions: []ActionStatistic{
			{Action: "login", Count: 3250, Percentage: 31.7},
			{Action: "upload_file", Count: 2180, Percentage: 21.3},
			{Action: "download_file", Count: 1950, Percentage: 19.0},
			{Action: "view_file", Count: 1420, Percentage: 13.9},
			{Action: "share_file", Count: 890, Percentage: 8.7},
			{Action: "delete_file", Count: 560, Percentage: 5.5},
		},
		HourlyDistribution: generateHourlyData(),
		DailyTrend:        generateDailyTrend(period),
	}
	
	SuccessResponse(c, http.StatusOK, "获取成功", response)
}

// BatchDeleteLogs 批量删除活动日志
// @Summary 批量删除活动日志
// @Description 管理员批量删除过期的活动日志
// @Tags 活动日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BatchDeleteLogsRequest true "批量删除请求"
// @Success 200 {object} Response{data=BatchDeleteLogsResponse} "删除成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 403 {object} Response "权限不足"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/admin/activity/batch-delete [post]
func (h *ActivityHandler) BatchDeleteLogs(c *gin.Context) {
	var req BatchDeleteLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数格式错误", err.Error())
		return
	}
	
	// 解析日期
	beforeDate, err := time.Parse("2006-01-02", req.BeforeDate)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误", "无效的日期格式")
		return
	}
	
	// 检查日期是否过于接近
	if time.Since(beforeDate).Hours() < 24*7 {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误", "不能删除7天内的日志")
		return
	}
	
	// 这里应该调用Service层的批量删除方法
	// 暂时返回示例响应
	deletedCount := 1250 // 实际应该从数据库删除操作获取
	
	response := BatchDeleteLogsResponse{
		DeletedCount: deletedCount,
		BeforeDate:   req.BeforeDate,
	}
	
	SuccessResponse(c, http.StatusOK, "批量删除成功", response)
}

// 响应结构体和辅助函数

// SystemActivityLogItem 系统活动日志项
type SystemActivityLogItem struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	StudentID string `json:"student_id"`
	NickName  string `json:"nick_name"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
	Detail    string `json:"detail"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	CreatedAt int64  `json:"created_at"`
}

// SystemActivityLogsResponse 系统活动日志响应
type SystemActivityLogsResponse struct {
	Logs     []*SystemActivityLogItem `json:"logs"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
	Pages    int64                    `json:"pages"`
}

// ActivityStatisticsResponse 活动统计响应
type ActivityStatisticsResponse struct {
	Period             string              `json:"period"`              // 统计周期
	TotalActions       int                 `json:"total_actions"`       // 总操作数
	UniqueUsers        int                 `json:"unique_users"`        // 活跃用户数
	TopActions         []ActionStatistic   `json:"top_actions"`         // 热门操作
	HourlyDistribution []HourlyData        `json:"hourly_distribution"` // 小时分布
	DailyTrend         []DailyData         `json:"daily_trend"`         // 日趋势
}

// ActionStatistic 操作统计
type ActionStatistic struct {
	Action     string  `json:"action"`     // 操作类型
	Count      int     `json:"count"`      // 次数
	Percentage float64 `json:"percentage"` // 占比
}

// HourlyData 小时数据
type HourlyData struct {
	Hour  int `json:"hour"`  // 小时 (0-23)
	Count int `json:"count"` // 操作次数
}

// DailyData 日数据
type DailyData struct {
	Date  string `json:"date"`  // 日期 (YYYY-MM-DD)
	Count int    `json:"count"` // 操作次数
}

// BatchDeleteLogsRequest 批量删除日志请求
type BatchDeleteLogsRequest struct {
	BeforeDate string   `json:"before_date" binding:"required"` // 删除此日期之前的日志 (YYYY-MM-DD)
	Actions    []string `json:"actions"`                        // 特定操作类型（可选）
}

// BatchDeleteLogsResponse 批量删除日志响应
type BatchDeleteLogsResponse struct {
	DeletedCount int    `json:"deleted_count"` // 删除的记录数
	BeforeDate   string `json:"before_date"`   // 删除的截止日期
}

// 辅助函数生成示例数据

func generateHourlyData() []HourlyData {
	data := make([]HourlyData, 24)
	for i := 0; i < 24; i++ {
		count := 200 + i*10 // 简单的示例数据
		if i >= 9 && i <= 17 { // 工作时间活跃度更高
			count = 400 + i*15
		}
		data[i] = HourlyData{
			Hour:  i,
			Count: count,
		}
	}
	return data
}

func generateDailyTrend(period string) []DailyData {
	var days int
	switch period {
	case "week":
		days = 7
	case "month":
		days = 30
	default:
		days = 1
	}
	
	data := make([]DailyData, days)
	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -days+i+1)
		data[i] = DailyData{
			Date:  date.Format("2006-01-02"),
			Count: 800 + i*20, // 简单的示例数据
		}
	}
	return data
}