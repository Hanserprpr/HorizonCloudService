package services

import (
	"context"
	"fmt"
	"user-service/internal/models"
)

// ListUsers 获取用户列表（管理员功能）
func (s *userService) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	// 计算偏移量
	offset := (req.Page - 1) * req.PageSize
	
	// 获取用户列表
	users, total, err := s.userRepo.List(ctx, offset, req.PageSize, req.Status)
	if err != nil {
		return nil, fmt.Errorf("获取用户列表失败: %w", err)
	}
	
	// 转换为响应格式
	userProfiles := make([]*UserProfile, len(users))
	for i, user := range users {
		userProfiles[i] = &UserProfile{
			ID:           user.ID,
			StudentID:    user.StudentID,
			NickName:     user.NickName,
			Email:        user.Email,
			Phone:        user.Phone,
			Avatar:       user.Avatar,
			RoleID:       user.RoleID,
			Status:       user.Status,
			StorageQuota: user.StorageQuota,
			StorageUsed:  user.StorageUsed,
			LastLoginAt:  user.LastLoginAt,
			LoginCount:   user.LoginCount,
			CreatedAt:    user.CreateTime(),
			UpdatedAt:    user.UpdateTime(),
		}
	}
	
	return &ListUsersResponse{
		Users:    userProfiles,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Pages:    (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}, nil
}

// SearchUsers 搜索用户（管理员功能）
func (s *userService) SearchUsers(ctx context.Context, req *SearchUsersRequest) (*SearchUsersResponse, error) {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	// 计算偏移量
	offset := (req.Page - 1) * req.PageSize
	
	// 搜索用户
	users, total, err := s.userRepo.Search(ctx, req.Keyword, offset, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("搜索用户失败: %w", err)
	}
	
	// 转换为响应格式
	userProfiles := make([]*UserProfile, len(users))
	for i, user := range users {
		userProfiles[i] = &UserProfile{
			ID:           user.ID,
			StudentID:    user.StudentID,
			NickName:     user.NickName,
			Email:        user.Email,
			Phone:        user.Phone,
			Avatar:       user.Avatar,
			RoleID:       user.RoleID,
			Status:       user.Status,
			StorageQuota: user.StorageQuota,
			StorageUsed:  user.StorageUsed,
			LastLoginAt:  user.LastLoginAt,
			LoginCount:   user.LoginCount,
			CreatedAt:    user.CreateTime(),
			UpdatedAt:    user.UpdateTime(),
		}
	}
	
	return &SearchUsersResponse{
		Users:    userProfiles,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Pages:    (total + int64(req.PageSize) - 1) / int64(req.PageSize),
		Keyword:  req.Keyword,
	}, nil
}

// GetUsersByRole 根据角色获取用户（管理员功能）
func (s *userService) GetUsersByRole(ctx context.Context, req *GetUsersByRoleRequest) (*GetUsersByRoleResponse, error) {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	// 计算偏移量
	offset := (req.Page - 1) * req.PageSize
	
	// 获取用户列表
	users, total, err := s.userRepo.GetUsersByRole(ctx, req.RoleID, offset, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("获取角色用户失败: %w", err)
	}
	
	// 转换为响应格式
	userProfiles := make([]*UserProfile, len(users))
	for i, user := range users {
		userProfiles[i] = &UserProfile{
			ID:           user.ID,
			StudentID:    user.StudentID,
			NickName:     user.NickName,
			Email:        user.Email,
			Phone:        user.Phone,
			Avatar:       user.Avatar,
			RoleID:       user.RoleID,
			Status:       user.Status,
			StorageQuota: user.StorageQuota,
			StorageUsed:  user.StorageUsed,
			LastLoginAt:  user.LastLoginAt,
			LoginCount:   user.LoginCount,
			CreatedAt:    user.CreateTime(),
			UpdatedAt:    user.UpdateTime(),
		}
	}
	
	return &GetUsersByRoleResponse{
		Users:    userProfiles,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Pages:    (total + int64(req.PageSize) - 1) / int64(req.PageSize),
		RoleID:   req.RoleID,
	}, nil
}

// UpdateUserStatus 更新用户状态（管理员功能）
func (s *userService) UpdateUserStatus(ctx context.Context, userID uint, status int) error {
	// 验证状态值
	if status != 1 && status != 2 {
		return fmt.Errorf("无效的状态值: %d", status)
	}
	
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}
	
	// 更新状态
	if err := s.userRepo.UpdateStatus(ctx, userID, status); err != nil {
		return fmt.Errorf("更新用户状态失败: %w", err)
	}
	
	// 记录活动
	action := "enable_user"
	detail := "启用用户"
	if status == 2 {
		action = "disable_user"
		detail = "禁用用户"
	}
	s.logUserActivity(ctx, userID, action, "user", detail, "", "")
	
	return nil
}

// DeleteUser 删除用户（管理员功能）
func (s *userService) DeleteUser(ctx context.Context, userID uint) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}
	
	// 软删除用户
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}
	
	// 记录活动
	s.logUserActivity(ctx, userID, "delete_user", "user", "删除用户", "", "")
	
	return nil
}

// GetUserQuota 获取用户存储配额
func (s *userService) GetUserQuota(ctx context.Context, userID uint) (*models.UserQuota, error) {
	quota, err := s.userRepo.GetUserQuota(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户配额失败: %w", err)
	}
	return quota, nil
}

// UpdateUserQuota 更新用户存储配额（管理员功能）
func (s *userService) UpdateUserQuota(ctx context.Context, userID uint, quota int64) error {
	// 验证配额值
	if quota <= 0 {
		return fmt.Errorf("配额值必须大于0")
	}
	
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}
	
	// 更新配额
	if err := s.userRepo.UpdateStorageQuota(ctx, userID, quota); err != nil {
		return fmt.Errorf("更新用户配额失败: %w", err)
	}
	
	// 记录活动
	s.logUserActivity(ctx, userID, "update_quota", "user", fmt.Sprintf("更新存储配额: %d bytes", quota), "", "")
	
	return nil
}

// UpdateStorageUsed 更新用户存储使用量
func (s *userService) UpdateStorageUsed(ctx context.Context, userID uint, used int64) error {
	// 验证使用量值
	if used < 0 {
		return fmt.Errorf("使用量不能为负数")
	}
	
	// 检查用户是否存在
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}
	
	// 检查是否超出配额
	if used > user.StorageQuota {
		return fmt.Errorf("存储使用量超出配额限制")
	}
	
	// 更新使用量
	if err := s.userRepo.UpdateStorageUsed(ctx, userID, used); err != nil {
		return fmt.Errorf("更新存储使用量失败: %w", err)
	}
	
	return nil
}

// LogActivity 记录用户活动
func (s *userService) LogActivity(ctx context.Context, req *LogActivityRequest) error {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		return err
	}
	
	log := &models.ActivityLog{
		UserID:    req.UserID,
		Action:    req.Action,
		Resource:  req.Resource,
		Detail:    req.Detail,
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
	}
	
	return s.userRepo.CreateActivityLog(ctx, log)
}

// GetUserActivityLogs 获取用户活动日志
func (s *userService) GetUserActivityLogs(ctx context.Context, req *GetActivityLogsRequest) (*GetActivityLogsResponse, error) {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	// 计算偏移量
	offset := (req.Page - 1) * req.PageSize
	
	// 获取活动日志
	logs, total, err := s.userRepo.GetUserActivityLogs(ctx, req.UserID, offset, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("获取活动日志失败: %w", err)
	}
	
	// 转换为响应格式
	activityLogs := make([]*ActivityLogItem, len(logs))
	for i, log := range logs {
		activityLogs[i] = &ActivityLogItem{
			ID:        log.ID,
			UserID:    log.UserID,
			Action:    log.Action,
			Resource:  log.Resource,
			Detail:    log.Detail,
			IPAddress: log.IPAddress,
			UserAgent: log.UserAgent,
			CreatedAt: log.CreateTime(),
		}
	}
	
	return &GetActivityLogsResponse{
		Logs:     activityLogs,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Pages:    (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}, nil
}