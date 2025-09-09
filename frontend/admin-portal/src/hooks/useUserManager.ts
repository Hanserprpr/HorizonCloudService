import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { message } from 'antd';
import { UserService } from '@services/index';
import type { 
  User, 
  CreateUserRequest, 
  UpdateUserRequest, 
  PaginatedResponse, 
  PaginationParams 
} from '../types';

// 用户查询参数
export interface UserQueryParams extends PaginationParams {
  keyword?: string; // 与userService保持一致
  search?: string;
  status?: 'active' | 'inactive';
  role?: string;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

// 获取用户列表
export const useUsers = (params?: UserQueryParams) => {
  return useQuery({
    queryKey: ['users', params],
    queryFn: () => UserService.getUsers(params),
    staleTime: 30 * 1000, // 30 seconds
    retry: 2,
  });
};

// 获取单个用户详情
export const useUser = (userId: number, enabled = true) => {
  return useQuery({
    queryKey: ['user', userId],
    queryFn: () => UserService.getUser(userId),
    enabled,
    staleTime: 30 * 1000,
    retry: 2,
  });
};

// 创建用户
export const useCreateUser = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (data: CreateUserRequest) => UserService.createUser(data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      message.success(`用户 "${data.data.username}" 创建成功`);
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '创建用户失败';
      message.error(errorMessage);
    },
  });
};

// 更新用户
export const useUpdateUser = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ userId, data }: { userId: number; data: UpdateUserRequest }) => 
      UserService.updateUser(userId, data),
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      queryClient.invalidateQueries({ queryKey: ['user', variables.userId] });
      message.success(`用户信息更新成功`);
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '更新用户失败';
      message.error(errorMessage);
    },
  });
};

// 删除用户
export const useDeleteUser = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (userId: number) => UserService.deleteUser(userId),
    onSuccess: (_, userId) => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      queryClient.removeQueries({ queryKey: ['user', userId] });
      message.success('用户删除成功');
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '删除用户失败';
      message.error(errorMessage);
    },
  });
};

// 批量删除用户
export const useBatchDeleteUsers = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (userIds: number[]) => UserService.batchDeleteUsers(userIds),
    onSuccess: (_, userIds) => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      userIds.forEach(id => {
        queryClient.removeQueries({ queryKey: ['user', id] });
      });
      message.success(`成功删除 ${userIds.length} 个用户`);
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '批量删除用户失败';
      message.error(errorMessage);
    },
  });
};

// 更新用户状态
export const useUpdateUserStatus = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ userId, status }: { userId: number; status: 'active' | 'inactive' }) =>
      UserService.updateUserStatus(userId, status),
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      queryClient.invalidateQueries({ queryKey: ['user', variables.userId] });
      const statusText = variables.status === 'active' ? '启用' : '禁用';
      message.success(`用户${statusText}成功`);
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '更新用户状态失败';
      message.error(errorMessage);
    },
  });
};

// 重置用户密码
export const useResetUserPassword = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ userId, newPassword }: { userId: number; newPassword: string }) =>
      UserService.resetPassword(userId, newPassword),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      message.success('密码重置成功');
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '重置密码失败';
      message.error(errorMessage);
    },
  });
};

// 更新用户配额
export const useUpdateUserQuota = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ userId, quota }: { userId: number; quota: number }) =>
      UserService.updateUserQuota(userId, quota),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      queryClient.invalidateQueries({ queryKey: ['user', variables.userId] });
      message.success('用户配额更新成功');
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '更新用户配额失败';
      message.error(errorMessage);
    },
  });
};

// 获取用户活动日志
export const useUserActivityLogs = (userId: number, params?: PaginationParams) => {
  return useQuery({
    queryKey: ['user-activity-logs', userId, params],
    queryFn: () => UserService.getUserActivityLogs(userId, params),
    enabled: !!userId,
    staleTime: 30 * 1000,
    retry: 2,
  });
};

// 获取用户统计信息
export const useUserStats = () => {
  return useQuery({
    queryKey: ['user-stats'],
    queryFn: () => UserService.getUserStats(),
    staleTime: 60 * 1000, // 1 minute
    retry: 2,
  });
};

// 获取用户详细统计信息
export const useUserDetailStats = (userId: number, enabled = true) => {
  return useQuery({
    queryKey: ['user-detail-stats', userId],
    queryFn: () => UserService.getUserStats(), // 可以扩展为特定用户的统计
    enabled: enabled && !!userId,
    staleTime: 60 * 1000, // 1 minute
    retry: 2,
  });
};

// 批量更新用户状态
export const useBatchUpdateUserStatus = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ userIds, status }: { userIds: number[]; status: 'active' | 'inactive' }) =>
      UserService.batchUpdateUserStatus(userIds, status),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      const statusText = variables.status === 'active' ? '启用' : '禁用';
      message.success(`批量${statusText}成功`);
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '批量更新用户状态失败';
      message.error(errorMessage);
    },
  });
};

// 批量更新用户配额
export const useBatchUpdateUserQuota = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ userIds, quota }: { userIds: number[]; quota: number }) =>
      UserService.batchUpdateUserQuota(userIds, quota),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      message.success(`批量设置配额成功`);
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '批量设置配额失败';
      message.error(errorMessage);
    },
  });
};

// 重置密码的别名，保持一致性
export const useResetPassword = useResetUserPassword;