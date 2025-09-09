import { apiClient } from './api';
import type { User, PaginatedResponse, PaginationParams, UserStats } from '../types';

export class UserService {
  // 获取用户列表
  static async getUsers(params?: PaginationParams & {
    keyword?: string;
    status?: number;
    role?: string;
  }): Promise<PaginatedResponse<User>> {
    const response = await apiClient.get<PaginatedResponse<User>>('/api/v1/admin/users', params);
    return response.data;
  }

  // 获取单个用户详情
  static async getUser(userId: number): Promise<User> {
    const response = await apiClient.get<User>(`/api/v1/admin/users/${userId}`);
    return response.data;
  }

  // 创建用户
  static async createUser(userData: {
    username: string;
    email: string;
    password: string;
    display_name?: string;
    role: string;
    storage_quota: number;
    status: number;
  }): Promise<User> {
    const response = await apiClient.post<User>('/api/v1/admin/users', userData);
    return response.data;
  }

  // 更新用户
  static async updateUser(userId: number, userData: Partial<User>): Promise<User> {
    const response = await apiClient.put<User>(`/api/v1/admin/users/${userId}`, userData);
    return response.data;
  }

  // 删除用户
  static async deleteUser(userId: number): Promise<void> {
    await apiClient.delete(`/api/v1/admin/users/${userId}`);
  }

  // 批量删除用户
  static async batchDeleteUsers(userIds: number[]): Promise<void> {
    await apiClient.post('/api/v1/admin/users/batch-delete', { user_ids: userIds });
  }

  // 更新用户状态
  static async updateUserStatus(userId: number, status: 'active' | 'inactive'): Promise<User> {
    const statusValue = status === 'active' ? 1 : 0;
    const response = await apiClient.patch<User>(`/api/v1/admin/users/${userId}/status`, { 
      status: statusValue 
    });
    return response.data;
  }

  // 重置用户密码
  static async resetPassword(userId: number, newPassword: string): Promise<void> {
    await apiClient.post(`/api/v1/admin/users/${userId}/reset-password`, {
      new_password: newPassword,
    });
  }

  // 更新用户配额
  static async updateUserQuota(userId: number, quota: number): Promise<User> {
    const response = await apiClient.patch<User>(`/api/v1/admin/users/${userId}/quota`, {
      storage_quota: quota,
    });
    return response.data;
  }

  // 获取用户活动日志
  static async getUserActivityLogs(userId: number, params?: PaginationParams): Promise<PaginatedResponse<any>> {
    const response = await apiClient.get<PaginatedResponse<any>>(
      `/api/v1/admin/users/${userId}/activity`,
      params
    );
    return response.data;
  }


  // 获取用户总体统计
  static async getUserStats(): Promise<{
    total_users: number;
    active_users: number;
    total_storage_used: number;
    average_storage_per_user: number;
  }> {
    const response = await apiClient.get('/api/v1/admin/users/stats');
    return response.data;
  }

  // 获取单个用户详细统计
  static async getUserDetailStats(userId: number): Promise<UserStats> {
    const response = await apiClient.get<UserStats>(`/api/v1/admin/users/${userId}/stats`);
    return response.data;
  }

  // 批量更新用户状态
  static async batchUpdateUserStatus(userIds: number[], status: 'active' | 'inactive'): Promise<void> {
    const statusValue = status === 'active' ? 1 : 0;
    await apiClient.post('/api/v1/admin/users/batch-status', {
      user_ids: userIds,
      status: statusValue
    });
  }

  // 批量更新用户配额
  static async batchUpdateUserQuota(userIds: number[], quota: number): Promise<void> {
    await apiClient.post('/api/v1/admin/users/batch-quota', {
      user_ids: userIds,
      storage_quota: quota
    });
  }
}

// 导出服务类作为默认导出
export const userService = UserService;