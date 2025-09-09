import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { message } from 'antd';
import { SettingsService } from '@services/index';
import type { 
  SystemSettings, 
  SystemInfo, 
  UpdateSettingsRequest, 
  StorageConfig,
  EmailConfig 
} from '../types';

// 获取系统设置
export const useSystemSettings = () => {
  return useQuery({
    queryKey: ['system-settings'],
    queryFn: () => SettingsService.getSettings(),
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 2,
  });
};

// 更新系统设置
export const useUpdateSettings = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (settings: UpdateSettingsRequest) => SettingsService.updateSettings(settings),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['system-settings'] });
      message.success('系统设置更新成功');
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '更新系统设置失败';
      message.error(errorMessage);
    },
  });
};

// 获取系统信息
export const useSystemInfo = () => {
  return useQuery({
    queryKey: ['system-info'],
    queryFn: () => SettingsService.getSystemInfo(),
    staleTime: 1 * 60 * 1000, // 1 minute
    retry: 2,
    refetchInterval: 30 * 1000, // 每30秒自动刷新
  });
};

// 系统健康检查
export const useHealthCheck = () => {
  return useQuery({
    queryKey: ['health-check'],
    queryFn: () => SettingsService.healthCheck(),
    staleTime: 30 * 1000, // 30 seconds
    retry: 2,
    refetchInterval: 15 * 1000, // 每15秒自动检查
  });
};

// 获取存储配置
export const useStorageConfig = () => {
  return useQuery({
    queryKey: ['storage-config'],
    queryFn: () => SettingsService.getStorageConfig(),
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 2,
  });
};

// 更新存储配置
export const useUpdateStorageConfig = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (config: StorageConfig) => SettingsService.updateStorageConfig(config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['storage-config'] });
      message.success('存储配置更新成功');
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '更新存储配置失败';
      message.error(errorMessage);
    },
  });
};

// 测试存储连接
export const useTestStorageConnection = () => {
  return useMutation({
    mutationFn: (config: StorageConfig) => SettingsService.testStorageConnection(config),
    onSuccess: (data) => {
      if (data.success) {
        message.success('存储连接测试成功');
      } else {
        message.error(`存储连接测试失败: ${data.message}`);
      }
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '存储连接测试失败';
      message.error(errorMessage);
    },
  });
};

// 获取邮件配置
export const useEmailConfig = () => {
  return useQuery({
    queryKey: ['email-config'],
    queryFn: () => SettingsService.getEmailConfig(),
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 2,
  });
};

// 更新邮件配置
export const useUpdateEmailConfig = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (config: EmailConfig) => SettingsService.updateEmailConfig(config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['email-config'] });
      message.success('邮件配置更新成功');
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '更新邮件配置失败';
      message.error(errorMessage);
    },
  });
};

// 测试邮件发送
export const useTestEmail = () => {
  return useMutation({
    mutationFn: ({ config, testEmail }: { config: EmailConfig; testEmail: string }) =>
      SettingsService.testEmail(config, testEmail),
    onSuccess: (data) => {
      if (data.success) {
        message.success('测试邮件发送成功');
      } else {
        message.error(`测试邮件发送失败: ${data.message}`);
      }
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '测试邮件发送失败';
      message.error(errorMessage);
    },
  });
};

// 重启系统服务
export const useRestartService = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (serviceName: string) => SettingsService.restartService(serviceName),
    onSuccess: (data, serviceName) => {
      queryClient.invalidateQueries({ queryKey: ['system-info'] });
      queryClient.invalidateQueries({ queryKey: ['health-check'] });
      if (data.success) {
        message.success(`服务 ${serviceName} 重启成功`);
      } else {
        message.error(`服务 ${serviceName} 重启失败: ${data.message}`);
      }
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '服务重启失败';
      message.error(errorMessage);
    },
  });
};

// 清理系统缓存
export const useClearCache = () => {
  return useMutation({
    mutationFn: (cacheType?: 'all' | 'thumbnails' | 'sessions' | 'temp') =>
      SettingsService.clearCache(cacheType),
    onSuccess: (data) => {
      if (data.success) {
        message.success(`缓存清理成功，清理了 ${data.cleared_items} 个项目`);
      } else {
        message.error(`缓存清理失败: ${data.message}`);
      }
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '缓存清理失败';
      message.error(errorMessage);
    },
  });
};

// 导出系统设置
export const useExportSettings = () => {
  return useMutation({
    mutationFn: () => SettingsService.exportSettings(),
    onSuccess: (blob) => {
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `system-settings-${new Date().toISOString().split('T')[0]}.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
      message.success('系统设置导出成功');
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '系统设置导出失败';
      message.error(errorMessage);
    },
  });
};

// 导入系统设置
export const useImportSettings = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (file: File) => SettingsService.importSettings(file),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['system-settings'] });
      if (data.success) {
        message.success(`系统设置导入成功，导入了 ${data.imported_settings} 个配置项`);
      } else {
        message.error(`系统设置导入失败: ${data.message}`);
      }
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '系统设置导入失败';
      message.error(errorMessage);
    },
  });
};

// 重置设置到默认值
export const useResetSettings = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: () => SettingsService.resetToDefaults(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['system-settings'] });
      message.success('系统设置已重置为默认值');
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '重置系统设置失败';
      message.error(errorMessage);
    },
  });
};

// 获取系统日志
export const useSystemLogs = (params?: {
  level?: string;
  service?: string;
  limit?: number;
  offset?: number;
}) => {
  return useQuery({
    queryKey: ['system-logs', params],
    queryFn: () => SettingsService.getSystemLogs(params),
    staleTime: 30 * 1000, // 30 seconds
    retry: 2,
  });
};

// 下载系统日志
export const useDownloadLogs = () => {
  return useMutation({
    mutationFn: ({ service, date }: { service?: string; date?: string }) =>
      SettingsService.downloadLogs(service, date),
    onSuccess: (blob, variables) => {
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      const filename = `logs-${variables.service || 'all'}-${variables.date || new Date().toISOString().split('T')[0]}.log`;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
      message.success('日志下载成功');
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '日志下载失败';
      message.error(errorMessage);
    },
  });
};