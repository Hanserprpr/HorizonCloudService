import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { message } from 'antd';
import { MockSettingsService } from '@services/mockSettingsService';
import { formatFileSize } from '@utils/index';
import type { 
  SystemSettings, 
  UpdateSettingsRequest,
  StorageConfig,
  SystemInfo,
  HealthCheck,
  ServiceRestartRequest,
  ClearCacheRequest,
  DownloadLogsRequest
} from '../types';

// 获取系统设置
export const useSystemSettings = () => {
  return useQuery({
    queryKey: ['system-settings'],
    queryFn: () => MockSettingsService.getSettings(),
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 2,
  });
};

// 更新系统设置
export const useUpdateSettings = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (settings: UpdateSettingsRequest) => MockSettingsService.updateSettings(settings),
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
    queryFn: () => MockSettingsService.getSystemInfo(),
    staleTime: 1 * 60 * 1000, // 1 minute
    retry: 2,
    refetchInterval: 30 * 1000, // 每30秒自动刷新
  });
};

// 系统健康检查
export const useHealthCheck = () => {
  return useQuery({
    queryKey: ['health-check'],
    queryFn: () => MockSettingsService.getHealthCheck(),
    staleTime: 30 * 1000, // 30 seconds
    retry: 2,
    refetchInterval: 15 * 1000, // 每15秒自动检查
  });
};

// 获取存储配置
export const useStorageConfig = () => {
  return useQuery({
    queryKey: ['storage-config'],
    queryFn: () => MockSettingsService.getStorageConfig(),
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 2,
  });
};

// 更新存储配置
export const useUpdateStorageConfig = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (config: StorageConfig) => MockSettingsService.updateStorageConfig(config),
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
    mutationFn: (config: StorageConfig) => MockSettingsService.testStorageConnection(config),
    onSuccess: (data) => {
      if (data.success) {
        message.success(`存储连接测试成功: ${data.message}`);
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

// 重启系统服务
export const useRestartService = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (serviceName: string) => MockSettingsService.restartService(serviceName),
    onSuccess: (data, serviceName) => {
      // 刷新相关数据
      queryClient.invalidateQueries({ queryKey: ['system-info'] });
      queryClient.invalidateQueries({ queryKey: ['health-check'] });
      
      if (data.success) {
        message.success(data.message);
      } else {
        message.error(data.message);
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
    mutationFn: (cacheType?: string) => MockSettingsService.clearCache(cacheType),
    onSuccess: (data) => {
      if (data.success) {
        let successMessage = data.message;
        if (data.cleared_size) {
          successMessage += ` (清理了 ${formatFileSize(data.cleared_size)})`;
        }
        message.success(successMessage);
      } else {
        message.error(data.message);
      }
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '缓存清理失败';
      message.error(errorMessage);
    },
  });
};

// 下载系统日志
export const useDownloadLogs = () => {
  return useMutation({
    mutationFn: (request: DownloadLogsRequest) => MockSettingsService.downloadLogs(request),
    onSuccess: (data) => {
      // 创建下载链接
      const link = document.createElement('a');
      link.href = data.download_url;
      link.download = data.download_url.split('/').pop() || 'system-logs.zip';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      message.success('日志下载开始，请检查下载文件夹');
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '日志下载失败';
      message.error(errorMessage);
    },
  });
};

// 导出系统设置
export const useExportSettings = () => {
  return useMutation({
    mutationFn: () => MockSettingsService.exportSettings(),
    onSuccess: (data) => {
      // 创建下载链接
      const link = document.createElement('a');
      link.href = data.download_url;
      link.download = `system-settings-${new Date().toISOString().split('T')[0]}.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
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
    mutationFn: (file: File) => MockSettingsService.importSettings(file),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['system-settings'] });
      if (data.success) {
        message.success(`系统设置导入成功，导入了 ${data.imported_count} 个配置项`);
      } else {
        message.error('系统设置导入失败');
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
    mutationFn: () => MockSettingsService.resetSettings(),
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

// 验证系统设置
export const useValidateSettings = () => {
  return useMutation({
    mutationFn: (settings: Partial<SystemSettings>) => MockSettingsService.validateSettings(settings),
    onSuccess: (data) => {
      if (data.valid) {
        message.success('设置验证通过');
      } else {
        const errorCount = Object.keys(data.errors).length;
        const warningCount = Object.keys(data.warnings).length;
        message.error(`设置验证失败: ${errorCount} 个错误, ${warningCount} 个警告`);
      }
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message || '设置验证失败';
      message.error(errorMessage);
    },
  });
};