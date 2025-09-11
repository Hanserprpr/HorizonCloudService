package handlers

import (
	"user-service/internal/services"
)

// InternalHandler 内部服务处理器
type InternalHandler struct {
	userService services.UserService
}

// NewInternalHandler 创建内部服务处理器
func NewInternalHandler(userService services.UserService) *InternalHandler {
	return &InternalHandler{
		userService: userService,
	}
}