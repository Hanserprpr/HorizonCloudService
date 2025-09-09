import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { message } from 'antd';
import { UserService } from '@services/userService';
import { queryKeys } from '@lib/queryClient';
import { SUCCESS_MESSAGES } from '@constants/index';
import type { User, PaginationParams } from '../types';

// 用户列表hook
export const useUsers = (params?: PaginationParams & {
  keyword?: string;
  status?: number;
  role?: string;
}) => {
  return useQuery({
    queryKey: queryKeys.users.list(params || {}),
    queryFn: () => UserService.getUsers(params),
    staleTime: 2 * 60 * 1000, // 2分钟
  });
};

// 单个用户详情hook
export const useUser = (userId: number, enabled: boolean = true) => {
  return useQuery({
    queryKey: queryKeys.users.detail(userId),
    queryFn: () => UserService.getUser(userId),
    enabled: enabled && !!userId,
  });
};

// 用户统计hook
export const useUserStats = (userId: number, enabled: boolean = true) => {
  return useQuery({
    queryKey: [...queryKeys.users.detail(userId), 'stats'],
    queryFn: () => UserService.getUserStats(userId),
    enabled: enabled && !!userId,
  });
};

// 用户整体统计hook
export const useUsersOverallStats = () => {
  return useQuery({
    queryKey: queryKeys.users.stats(),
    queryFn: () => UserService.getUsersOverallStats(),
    staleTime: 5 * 60 * 1000, // 5分钟
  });
};

// 用户操作hooks
export const useUserMutations = () => {
  const queryClient = useQueryClient();

  // 创建用户
  const createUser = useMutation({
    mutationFn: UserService.createUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.lists() });
      queryClient.invalidateQueries({ queryKey: queryKeys.users.stats() });
      message.success(SUCCESS_MESSAGES.CREATE_SUCCESS);
    },
  });

  // 更新用户
  const updateUser = useMutation({
    mutationFn: ({ userId, userData }: { userId: number; userData: Partial<User> }) =>
      UserService.updateUser(userId, userData),
    onSuccess: (updatedUser) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.lists() });
      queryClient.setQueryData(queryKeys.users.detail(updatedUser.id), updatedUser);
      message.success(SUCCESS_MESSAGES.UPDATE_SUCCESS);
    },
  });

  // 删除用户
  const deleteUser = useMutation({
    mutationFn: UserService.deleteUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.lists() });
      queryClient.invalidateQueries({ queryKey: queryKeys.users.stats() });
      message.success(SUCCESS_MESSAGES.DELETE_SUCCESS);
    },
  });

  // 批量删除用户
  const deleteUsers = useMutation({
    mutationFn: UserService.deleteUsers,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.lists() });
      queryClient.invalidateQueries({ queryKey: queryKeys.users.stats() });
      message.success(SUCCESS_MESSAGES.DELETE_SUCCESS);
    },
  });

  // 更新用户状态
  const updateUserStatus = useMutation({
    mutationFn: ({ userId, status }: { userId: number; status: number }) =>
      UserService.updateUserStatus(userId, status),
    onSuccess: (updatedUser) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.lists() });
      queryClient.setQueryData(queryKeys.users.detail(updatedUser.id), updatedUser);
      message.success('用户状态更新成功');
    },
  });

  // 重置用户密码
  const resetUserPassword = useMutation({
    mutationFn: ({ userId, newPassword }: { userId: number; newPassword: string }) =>
      UserService.resetUserPassword(userId, newPassword),
    onSuccess: () => {
      message.success('密码重置成功');
    },
  });

  // 更新用户配额
  const updateUserQuota = useMutation({
    mutationFn: ({ userId, quota }: { userId: number; quota: number }) =>
      UserService.updateUserQuota(userId, quota),
    onSuccess: (updatedUser) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.lists() });
      queryClient.setQueryData(queryKeys.users.detail(updatedUser.id), updatedUser);
      message.success('用户配额更新成功');
    },
  });

  return {
    createUser,
    updateUser,
    deleteUser,
    deleteUsers,
    updateUserStatus,
    resetUserPassword,
    updateUserQuota,
  };
};