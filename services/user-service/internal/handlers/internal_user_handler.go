package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// InternalUserInfo 内部API用户信息结构体（用于微服务间通信）
type InternalUserInfo struct {
	ID               uint                   `json:"id"`
	Username         string                 `json:"username"`
	Email            string                 `json:"email"`
	Role             string                 `json:"role"`
	Status           int                    `json:"status"`
	StorageQuota     int64                  `json:"storage_quota"`
	FileCountQuota   int64                  `json:"file_count_quota"`
	StorageUsed      int64                  `json:"storage_used"`
	FileCount        int64                  `json:"file_count"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	LastLoginAt      *time.Time             `json:"last_login_at"`
	Settings         map[string]interface{} `json:"settings"`
}

// GetUser 获取用户信息（内部服务调用）
// @Summary 获取用户信息（内部服务调用）
// @Description 供其他微服务调用获取用户信息
// @Tags 内部服务
// @Accept json
// @Produce json
// @Param user_id path int true "用户ID"
// @Success 200 {object} Response{data=services.UserInfo} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 404 {object} Response "用户不存在"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/internal/users/{user_id} [get]
func (h *InternalHandler) GetUser(c *gin.Context) {
	// 解析路径参数
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误", "无效的用户ID")
		return
	}

	// 调用服务层获取用户信息
	user, err := h.userService.GetUserProfile(c.Request.Context(), uint(userID))
	if err != nil {
		if err.Error() == "用户不存在" {
			ErrorResponse(c, http.StatusNotFound, "用户不存在", err.Error())
		} else {
			ErrorResponse(c, http.StatusInternalServerError, "获取用户信息失败", err.Error())
		}
		return
	}

	// 转换为内部API格式，确保时间字段正确序列化
	internalUser := &InternalUserInfo{
		ID:             user.ID,
		Username:       user.NickName, // 使用昵称作为用户名
		Email:          user.Email,
		Role:           getUserRole(user.RoleID), // 将角色ID转换为角色名
		Status:         user.Status,
		StorageQuota:   user.StorageQuota,
		FileCountQuota: 10000, // 默认文件数量配额
		StorageUsed:    user.StorageUsed,
		FileCount:      0, // 需要从文件服务获取，暂时设为0
		CreatedAt:      time.Unix(user.CreatedAt/1000, 0),      // 转换Unix毫秒时间戳为time.Time
		UpdatedAt:      time.Unix(user.UpdatedAt/1000, 0),      // 转换Unix毫秒时间戳为time.Time
		LastLoginAt:    user.LastLoginAt,      // 直接使用*time.Time，已经是正确格式
		Settings:       make(map[string]interface{}),           // 初始化设置map
	}

	SuccessResponse(c, http.StatusOK, "获取成功", internalUser)
}

// getUserRole 根据角色ID获取角色名
func getUserRole(roleID int) string {
	switch roleID {
	case 1:
		return "user"
	case 2:
		return "admin"
	case 3:
		return "super_admin"
	default:
		return "user"
	}
}

