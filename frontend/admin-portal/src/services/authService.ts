import { apiClient } from './api';
import type { LoginForm, LoginResponse, User } from '../types';

export class AuthService {
  // 登录
  static async login(credentials: LoginForm): Promise<LoginResponse> {
    const response = await apiClient.post<LoginResponse>('/api/v1/auth/login', credentials);
    return response.data;
  }

  // 登出
  static async logout(): Promise<void> {
    try {
      await apiClient.post('/api/v1/auth/logout');
    } catch (error) {
      // 登出失败也要清除本地状态
      console.warn('Logout request failed:', error);
    }
  }

  // 刷新token
  static async refreshToken(refreshToken: string): Promise<{ access_token: string }> {
    const response = await apiClient.post<{ access_token: string }>('/api/v1/auth/refresh', {
      refresh_token: refreshToken,
    });
    return response.data;
  }

  // 获取当前用户信息
  static async getCurrentUser(): Promise<User> {
    const response = await apiClient.get<User>('/api/v1/users/profile');
    return response.data;
  }

  // 更新用户资料
  static async updateProfile(userData: Partial<User>): Promise<User> {
    const response = await apiClient.put<User>('/api/v1/users/profile', userData);
    return response.data;
  }

  // 修改密码
  static async changePassword(data: {
    old_password: string;
    new_password: string;
    confirm_password: string;
  }): Promise<void> {
    await apiClient.post('/api/v1/users/change-password', data);
  }

  // 验证token有效性
  static async validateToken(): Promise<boolean> {
    try {
      await apiClient.get('/api/v1/auth/validate');
      return true;
    } catch {
      return false;
    }
  }
}