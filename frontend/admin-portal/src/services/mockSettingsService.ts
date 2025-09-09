// Mock设置服务 - 用于前端演示和开发
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
import { sleep } from '@utils/index';

// Mock系统设置数据
const mockSystemSettings: SystemSettings = {
  // 基本设置
  system_name: 'HorizonCloud 媒体云存储系统',
  system_description: '基于AI驱动的企业级媒体云存储和管理平台',
  system_version: 'v1.0.0',
  enable_registration: true,
  maintenance_mode: false,
  
  // 存储设置
  default_storage_quota: 10 * 1024 * 1024 * 1024, // 10GB
  max_storage_quota: 100 * 1024 * 1024 * 1024, // 100GB
  storage_backend: 'minio',
  enable_file_deduplication: true,
  enable_thumbnail_generation: true,
  thumbnail_quality: 80,
  max_file_size: 5 * 1024 * 1024 * 1024, // 5GB
  allowed_file_types: ['image/jpeg', 'image/png', 'image/gif', 'video/mp4', 'application/pdf'],
  
  // 安全设置
  require_email_verification: true,
  password_min_length: 8,
  password_complexity: true,
  session_timeout: 3600, // 1小时
  max_login_attempts: 5,
  enable_logging: true,
  log_level: 'info',
  
  // AI设置
  enable_ai_analysis: true,
  ai_analysis_queue_size: 100,
  enable_semantic_search: true,
  auto_tagging_enabled: true,
  
  // 通知设置
  enable_email_notifications: true,
  notification_email: 'admin@horizon.cloud',
  enable_system_alerts: true,
};

// Mock存储配置
const mockStorageConfig: StorageConfig = {
  backend: 'minio',
  endpoint: 'http://localhost:9000',
  access_key: 'minioadmin',
  secret_key: 'minioadmin',
  bucket: 'media-storage',
  region: 'us-east-1',
  use_ssl: false,
  path_style: true,
};

// Mock系统信息
const mockSystemInfo: SystemInfo = {
  version: 'v1.0.0-beta',
  build_time: '2024-01-15T10:30:00Z',
  go_version: 'go1.21.5',
  os: 'Linux',
  arch: 'amd64',
  uptime: 86400, // 24小时（秒）
  
  memory_usage: {
    total: 8 * 1024 * 1024 * 1024, // 8GB
    used: 3.2 * 1024 * 1024 * 1024, // 3.2GB
    available: 4.8 * 1024 * 1024 * 1024, // 4.8GB
  },
  
  storage_info: {
    total_space: 1024 * 1024 * 1024 * 1024, // 1TB
    used_space: 256 * 1024 * 1024 * 1024, // 256GB
    free_space: 768 * 1024 * 1024 * 1024, // 768GB
    files_count: 12567,
  },
  
  database_info: {
    type: 'PostgreSQL',
    version: '15.3',
    size: 2.5 * 1024 * 1024 * 1024, // 2.5GB
    connections: 25,
  },
  
  services_status: {
    user_service: true,
    file_service: true,
    ai_service: true,
    search_service: false, // 模拟一个服务异常
  },
};

// Mock健康检查
const mockHealthCheck: HealthCheck = {
  status: 'unhealthy', // 模拟系统异常状态
  checks: {
    'database_connection': true,
    'storage_backend': true,
    'redis_cache': true,
    'search_service': false, // 模拟搜索服务异常
    'ai_service': true,
  },
  timestamp: new Date().toISOString(),
};

export class MockSettingsService {
  // 获取系统设置
  static async getSettings(): Promise<SystemSettings> {
    await sleep(800); // 模拟网络延迟
    return { ...mockSystemSettings };
  }

  // 更新系统设置
  static async updateSettings(settings: UpdateSettingsRequest): Promise<SystemSettings> {
    await sleep(1200);
    
    // 模拟更新逻辑
    Object.assign(mockSystemSettings, settings);
    
    return { ...mockSystemSettings };
  }

  // 重置系统设置
  static async resetSettings(): Promise<SystemSettings> {
    await sleep(1000);
    return { ...mockSystemSettings };
  }

  // 获取存储配置
  static async getStorageConfig(): Promise<StorageConfig> {
    await sleep(600);
    return { ...mockStorageConfig };
  }

  // 更新存储配置
  static async updateStorageConfig(config: StorageConfig): Promise<StorageConfig> {
    await sleep(1500);
    
    // 模拟更新逻辑
    Object.assign(mockStorageConfig, config);
    
    return { ...mockStorageConfig };
  }

  // 测试存储连接
  static async testStorageConnection(config: StorageConfig): Promise<{
    success: boolean;
    message: string;
    details?: any;
  }> {
    await sleep(2000); // 模拟连接测试时间
    
    // 模拟连接测试逻辑
    const isValidEndpoint = config.endpoint && config.endpoint.startsWith('http');
    const hasCredentials = config.access_key && config.secret_key;
    
    if (!isValidEndpoint) {
      return {
        success: false,
        message: '存储端点地址无效',
        details: { error: 'Invalid endpoint format' },
      };
    }
    
    if (!hasCredentials) {
      return {
        success: false,
        message: '访问密钥配置不完整',
        details: { error: 'Missing access credentials' },
      };
    }
    
    // 模拟成功连接
    return {
      success: true,
      message: '存储连接测试成功',
      details: {
        latency: Math.floor(Math.random() * 100 + 50) + 'ms',
        region: config.region,
        bucket_exists: true,
      },
    };
  }

  // 获取系统信息
  static async getSystemInfo(): Promise<SystemInfo> {
    await sleep(1000);
    
    // 模拟动态数据更新
    const currentTime = Date.now();
    mockSystemInfo.uptime = Math.floor((currentTime - new Date('2024-01-15T10:30:00Z').getTime()) / 1000);
    
    // 模拟内存使用变化
    const baseUsed = 3.2 * 1024 * 1024 * 1024;
    const variation = (Math.random() - 0.5) * 0.5 * 1024 * 1024 * 1024; // ±0.5GB变化
    mockSystemInfo.memory_usage.used = Math.max(0, baseUsed + variation);
    mockSystemInfo.memory_usage.available = mockSystemInfo.memory_usage.total - mockSystemInfo.memory_usage.used;
    
    return { ...mockSystemInfo };
  }

  // 获取健康检查
  static async getHealthCheck(): Promise<HealthCheck> {
    await sleep(500);
    
    // 模拟健康状态变化
    const checks = { ...mockHealthCheck.checks };
    
    // 随机模拟服务状态变化
    if (Math.random() > 0.8) {
      checks.search_service = !checks.search_service;
    }
    
    const hasFailures = Object.values(checks).some(status => !status);
    
    return {
      status: hasFailures ? 'unhealthy' : 'healthy',
      checks,
      timestamp: new Date().toISOString(),
    };
  }

  // 重启服务
  static async restartService(serviceName: string): Promise<{ success: boolean; message: string }> {
    await sleep(3000); // 模拟重启时间
    
    // 更新服务状态
    if (mockSystemInfo.services_status.hasOwnProperty(serviceName)) {
      mockSystemInfo.services_status[serviceName as keyof typeof mockSystemInfo.services_status] = true;
    }
    
    return {
      success: true,
      message: `${serviceName} 重启成功`,
    };
  }

  // 清理缓存
  static async clearCache(cacheType?: string): Promise<{ success: boolean; message: string; cleared_size?: number }> {
    await sleep(2000);
    
    // 模拟清理的缓存大小
    const clearedSize = Math.floor(Math.random() * 500 + 100) * 1024 * 1024; // 100-600MB
    
    const cacheTypeLabels: Record<string, string> = {
      'all': '所有缓存',
      'thumbnails': '缩略图缓存',
      'sessions': '会话缓存',
      'temp': '临时文件',
    };
    
    const label = cacheTypeLabels[cacheType || 'all'] || '缓存';
    
    return {
      success: true,
      message: `${label}清理完成`,
      cleared_size: clearedSize,
    };
  }

  // 下载日志
  static async downloadLogs(request: DownloadLogsRequest): Promise<{ download_url: string; expires_in: number }> {
    await sleep(1500);
    
    // 模拟生成下载链接
    const logFileName = request.service 
      ? `${request.service}-logs-${new Date().toISOString().split('T')[0]}.zip`
      : `system-logs-${new Date().toISOString().split('T')[0]}.zip`;
    
    return {
      download_url: `/api/v1/admin/logs/download/${logFileName}?token=mock-download-token`,
      expires_in: 3600, // 1小时过期
    };
  }

  // 导出系统设置
  static async exportSettings(): Promise<{ download_url: string }> {
    await sleep(1000);
    
    return {
      download_url: `/api/v1/admin/settings/export?token=mock-export-token`,
    };
  }

  // 导入系统设置
  static async importSettings(file: File): Promise<{ success: boolean; imported_count: number }> {
    await sleep(2000);
    
    return {
      success: true,
      imported_count: Math.floor(Math.random() * 20 + 10), // 10-30个设置项
    };
  }

  // 验证设置
  static async validateSettings(settings: Partial<SystemSettings>): Promise<{
    valid: boolean;
    errors: Record<string, string[]>;
    warnings: Record<string, string[]>;
  }> {
    await sleep(800);
    
    const errors: Record<string, string[]> = {};
    const warnings: Record<string, string[]> = {};
    
    // 模拟验证逻辑
    if (settings.password_min_length && settings.password_min_length < 6) {
      errors.password_min_length = ['密码长度至少为6位'];
    }
    
    if (settings.session_timeout && settings.session_timeout < 300) {
      warnings.session_timeout = ['会话超时时间过短，建议至少5分钟'];
    }
    
    if (settings.default_storage_quota && settings.max_storage_quota && 
        settings.default_storage_quota > settings.max_storage_quota) {
      errors.default_storage_quota = ['默认配额不能超过最大配额'];
    }
    
    return {
      valid: Object.keys(errors).length === 0,
      errors,
      warnings,
    };
  }
}

// 默认导出
export const mockSettingsService = MockSettingsService;