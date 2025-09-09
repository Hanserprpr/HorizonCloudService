import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { message } from 'antd';
import { SystemService } from '@services/systemService';
import { queryKeys } from '@lib/queryClient';
import type { SystemStats } from '@types/index';

// 系统统计hook
export const useSystemStats = () => {
  return useQuery({
    queryKey: queryKeys.system.stats(),
    queryFn: SystemService.getSystemStats,
    staleTime: 30 * 1000, // 30秒
    refetchInterval: 60 * 1000, // 每分钟刷新
  });
};

// 系统健康状态hook
export const useHealthStatus = () => {
  return useQuery({
    queryKey: queryKeys.system.health(),
    queryFn: SystemService.getHealthStatus,
    staleTime: 15 * 1000, // 15秒
    refetchInterval: 30 * 1000, // 每30秒刷新
  });
};

// 系统配置hook
export const useSystemConfig = () => {
  return useQuery({
    queryKey: queryKeys.system.config(),
    queryFn: SystemService.getSystemConfig,
    staleTime: 5 * 60 * 1000, // 5分钟
  });
};

// 系统日志hook
export const useSystemLogs = (params?: {
  level?: 'debug' | 'info' | 'warn' | 'error';
  service?: string;
  start_date?: string;
  end_date?: string;
  page?: number;
  size?: number;
}, enabled: boolean = true) => {
  return useQuery({
    queryKey: ['system', 'logs', params],
    queryFn: () => SystemService.getSystemLogs(params),
    enabled: enabled,
    staleTime: 30 * 1000, // 30秒
  });
};

// 维护状态hook
export const useMaintenanceStatus = () => {
  return useQuery({
    queryKey: ['system', 'maintenance'],
    queryFn: SystemService.getMaintenanceStatus,
    staleTime: 60 * 1000, // 1分钟
  });
};

// 系统操作hooks
export const useSystemMutations = () => {
  const queryClient = useQueryClient();

  // 更新系统配置
  const updateConfig = useMutation({
    mutationFn: SystemService.updateSystemConfig,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.system.config() });
      message.success('系统配置更新成功');
    },
  });

  // 清理缓存
  const clearCache = useMutation({
    mutationFn: (cacheType?: 'all' | 'thumbnails' | 'sessions' | 'temp') =>
      SystemService.clearCache(cacheType),
    onSuccess: () => {
      message.success('缓存清理完成');
    },
  });

  // 设置维护模式
  const setMaintenanceMode = useMutation({
    mutationFn: ({ enabled, message: msg }: { enabled: boolean; message?: string }) =>
      SystemService.setMaintenanceMode(enabled, msg),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['system', 'maintenance'] });
      message.success('维护模式设置成功');
    },
  });

  return {
    updateConfig,
    clearCache,
    setMaintenanceMode,
  };
};