import React from 'react';
import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query';
import { message } from 'antd';
import { useAuthStore } from '@stores/authStore';
import { AuthService } from '@services/authService';
import { queryKeys } from '@lib/queryClient';
import { SUCCESS_MESSAGES } from '@constants/index';
import type { LoginForm, LoginResponse } from '../types';

// 模拟API调用（开发阶段使用）
const mockApiCall = {
  login: async (credentials: LoginForm): Promise<LoginResponse> => {
    // 模拟API延迟
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    // 模拟登录验证
    if (credentials.student_id === 'admin' && credentials.password === 'admin123') {
      return {
        user: {
          id: 1,
          student_id: 'admin',
          email: 'admin@example.com',
          display_name: '系统管理员',
          status: 1,
          role: 'admin',
          storage_quota: 100 * 1024 * 1024 * 1024, // 100GB
          storage_used: 25 * 1024 * 1024 * 1024,   // 25GB
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
          last_login_at: new Date().toISOString(),
        },
        access_token: 'mock-jwt-token-12345',
        refresh_token: 'mock-refresh-token-67890',
        expires_in: 86400, // 24小时
      };
    } else {
      throw new Error('用户名或密码错误');
    }
  },

  getCurrentUser: async () => {
    await new Promise(resolve => setTimeout(resolve, 500));
    return {
      id: 1,
      student_id: 'admin',
      email: 'admin@example.com',
      display_name: '系统管理员',
      status: 1,
      role: 'admin',
      storage_quota: 100 * 1024 * 1024 * 1024,
      storage_used: 25 * 1024 * 1024 * 1024,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
      last_login_at: new Date().toISOString(),
    };
  },

  refreshToken: async (refreshToken: string): Promise<{ access_token: string }> => {
    await new Promise(resolve => setTimeout(resolve, 500));
    return {
      access_token: 'new-mock-jwt-token-' + Date.now(),
    };
  },
};

export const useAuth = () => {
  const queryClient = useQueryClient();
  const authStore = useAuthStore();

  // 使用模拟API（后续可切换到真实API）
  const useRealApi = import.meta.env.VITE_USE_REAL_API === 'true';
  
  // 获取当前用户信息的query - 添加更多安全检查
  const userQuery = useQuery({
    queryKey: queryKeys.users.profile(),
    queryFn: useRealApi ? AuthService.getCurrentUser : mockApiCall.getCurrentUser,
    enabled: Boolean(authStore.isAuthenticated && authStore.token),
    staleTime: 5 * 60 * 1000, // 5分钟
    retry: 1,
    retryOnMount: false,
    refetchOnWindowFocus: false,
  });

  // 登录mutation
  const loginMutation = useMutation({
    mutationFn: useRealApi ? AuthService.login : mockApiCall.login,
    onMutate: () => {
      authStore.setLoading(true);
    },
    onSuccess: (data) => {
      authStore.login(data);
      message.success(SUCCESS_MESSAGES.LOGIN_SUCCESS);
      
      // 登录成功后立即获取用户信息
      queryClient.invalidateQueries({ queryKey: queryKeys.users.profile() });
    },
    onError: (error: Error) => {
      console.error('Login error:', error);
      message.error(error.message || '登录失败，请重试');
    },
    onSettled: () => {
      authStore.setLoading(false);
    },
  });

  // 登出mutation
  const logoutMutation = useMutation({
    mutationFn: useRealApi ? AuthService.logout : async () => {},
    onSuccess: () => {
      authStore.logout();
      queryClient.clear(); // 清除所有缓存
      message.success(SUCCESS_MESSAGES.LOGOUT_SUCCESS);
    },
    onError: () => {
      // 即使API调用失败也要清除本地状态
      authStore.logout();
      queryClient.clear();
      message.success(SUCCESS_MESSAGES.LOGOUT_SUCCESS);
    },
  });

  // 刷新token mutation
  const refreshTokenMutation = useMutation({
    mutationFn: useRealApi 
      ? (refreshToken: string) => AuthService.refreshToken(refreshToken)
      : mockApiCall.refreshToken,
    onSuccess: (data) => {
      authStore.setToken(data.access_token);
    },
    onError: () => {
      // 刷新失败，清除认证状态
      authStore.logout();
      queryClient.clear();
    },
  });

  // 更新用户资料mutation
  const updateProfileMutation = useMutation({
    mutationFn: useRealApi 
      ? (userData: any) => AuthService.updateProfile(userData)
      : async (userData: any) => ({ ...authStore.user!, ...userData }),
    onSuccess: (updatedUser) => {
      authStore.updateUser(updatedUser);
      queryClient.setQueryData(queryKeys.users.profile(), updatedUser);
      message.success('资料更新成功');
    },
  });

  // 修改密码mutation
  const changePasswordMutation = useMutation({
    mutationFn: useRealApi 
      ? (data: any) => AuthService.changePassword(data)
      : async () => {},
    onSuccess: () => {
      message.success('密码修改成功');
    },
  });

  // 登录
  const login = async (credentials: LoginForm) => {
    return loginMutation.mutateAsync(credentials);
  };

  // 登出
  const logout = () => {
    logoutMutation.mutate();
  };

  // 刷新token
  const refreshToken = () => {
    const refreshToken = authStore.refreshToken;
    if (refreshToken) {
      return refreshTokenMutation.mutateAsync(refreshToken);
    }
    return Promise.reject(new Error('No refresh token'));
  };

  // 更新用户资料
  const updateProfile = (userData: any) => {
    return updateProfileMutation.mutateAsync(userData);
  };

  // 修改密码
  const changePassword = (data: any) => {
    return changePasswordMutation.mutateAsync(data);
  };

  // 检查认证状态 - 简化版本
  const checkAuth = React.useCallback(() => {
    try {
      const token = authStore.token;
      const user = authStore.user;
      
      if (!token || !user) {
        authStore.logout();
        return false;
      }
      
      return true;
    } catch (error) {
      console.error('checkAuth error:', error);
      authStore.logout();
      return false;
    }
  }, [authStore]);

  // 验证token有效性
  const validateToken = async () => {
    if (!useRealApi) return true;
    
    try {
      return await AuthService.validateToken();
    } catch {
      return false;
    }
  };

  return {
    // 状态
    user: userQuery.data || authStore.user,
    token: authStore.token,
    isAuthenticated: authStore.isAuthenticated,
    loading: authStore.loading || loginMutation.isPending || userQuery.isLoading,
    
    // 方法
    login,
    logout,
    refreshToken,
    updateProfile,
    changePassword,
    checkAuth,
    validateToken,
    
    // 权限检查
    isAdmin: authStore.isAdmin,
    hasPermission: authStore.hasPermission,
    
    // mutation状态
    loginError: loginMutation.error,
    isLoggingIn: loginMutation.isPending,
    isLoggingOut: logoutMutation.isPending,
    isUpdatingProfile: updateProfileMutation.isPending,
    isChangingPassword: changePasswordMutation.isPending,
    
    // query状态
    isLoadingUser: userQuery.isLoading,
    userError: userQuery.error,
  };
};