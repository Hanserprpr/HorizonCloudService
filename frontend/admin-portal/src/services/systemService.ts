import { apiClient } from './api';
import type { SystemStats } from '../types';

export class SystemService {
  // 获取系统统计信息
  static async getSystemStats(): Promise<SystemStats> {
    const response = await apiClient.get<SystemStats>('/api/v1/system/stats');
    return response.data;
  }

  // 获取系统健康状态
  static async getHealthStatus(): Promise<{
    status: 'healthy' | 'unhealthy';
    services: Array<{
      name: string;
      status: 'online' | 'offline';
      response_time?: number;
      last_check: string;
    }>;
    uptime: number;
    memory_usage: number;
    cpu_usage: number;
    disk_usage: number;
  }> {
    const response = await apiClient.get('/api/v1/system/health');
    return response.data;
  }

  // 获取系统配置
  static async getSystemConfig(): Promise<{
    upload: {
      max_file_size: number;
      chunk_size: number;
      allowed_types: string[];
    };
    storage: {
      default_quota: number;
      max_quota: number;
    };
    security: {
      password_policy: {
        min_length: number;
        require_uppercase: boolean;
        require_lowercase: boolean;
        require_numbers: boolean;
        require_special_chars: boolean;
      };
      session_timeout: number;
    };
  }> {
    const response = await apiClient.get('/api/v1/system/config');
    return response.data;
  }

  // 更新系统配置
  static async updateSystemConfig(config: any): Promise<void> {
    await apiClient.put('/api/v1/system/config', config);
  }

  // 获取系统日志
  static async getSystemLogs(params?: {
    level?: 'debug' | 'info' | 'warn' | 'error';
    service?: string;
    start_date?: string;
    end_date?: string;
    page?: number;
    size?: number;
  }): Promise<{
    logs: Array<{
      timestamp: string;
      level: string;
      service: string;
      message: string;
      details?: any;
    }>;
    total: number;
  }> {
    const response = await apiClient.get('/api/v1/system/logs', params);
    return response.data;
  }

  // 清理系统缓存
  static async clearCache(cacheType?: 'all' | 'thumbnails' | 'sessions' | 'temp'): Promise<void> {
    await apiClient.post('/api/v1/system/cache/clear', { type: cacheType || 'all' });
  }

  // 系统维护模式
  static async setMaintenanceMode(enabled: boolean, message?: string): Promise<void> {
    await apiClient.post('/api/v1/system/maintenance', { 
      enabled, 
      message: message || '系统维护中，请稍后再试' 
    });
  }

  // 获取维护状态
  static async getMaintenanceStatus(): Promise<{
    enabled: boolean;
    message: string;
    started_at?: string;
  }> {
    const response = await apiClient.get('/api/v1/system/maintenance');
    return response.data;
  }
}

// 导出服务类
export const systemService = SystemService;