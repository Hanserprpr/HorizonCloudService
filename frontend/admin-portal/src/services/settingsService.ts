import { apiClient } from './api';
import type { 
  SystemSettings, 
  SystemInfo, 
  UpdateSettingsRequest,
  StorageConfig,
  EmailConfig 
} from '../types';

export class SettingsService {
  // Mock数据存储
  private static mockSettings: SystemSettings = {
    siteName: '云存储管理后台',
    defaultUserQuota: 10 * 1024 * 1024 * 1024, // 10GB
    allowRegistration: true,
    maxFileSize: 500 * 1024 * 1024, // 500MB
    supportedFileTypes: [
      'image/jpeg', 'image/png', 'image/gif', 'image/webp',
      'video/mp4', 'video/avi', 'video/mov',
      'application/pdf', 'text/plain'
    ],
    enableThumbnails: true,
    thumbnailSizes: ['small', 'medium', 'large'],
    enableFileVersioning: true,
    maxFileVersions: 5,
    enableEmailNotifications: true,
    emailConfig: {
      smtpHost: 'smtp.example.com',
      smtpPort: 587,
      smtpUser: 'noreply@example.com',
      smtpPassword: '',
      smtpSecure: true,
      fromName: '云存储系统',
      fromEmail: 'noreply@example.com'
    },
    storageConfig: {
      type: 'local',
      localPath: './uploads',
      maxStorageSize: 1024 * 1024 * 1024 * 1024, // 1TB
      s3Config: {
        accessKeyId: '',
        secretAccessKey: '',
        region: 'us-west-2',
        bucket: '',
        endpoint: ''
      }
    },
    theme: {
      primaryColor: '#1677FF',
      darkMode: false
    },
    security: {
      sessionTimeout: 24 * 60, // 24小时（分钟）
      enableTwoFactor: false,
      passwordMinLength: 6,
      passwordRequireSpecialChar: false
    }
  };

  // 获取系统设置
  static async getSettings(): Promise<SystemSettings> {
    // 模拟网络延迟
    await new Promise(resolve => setTimeout(resolve, 200));
    return Promise.resolve(this.mockSettings);
  }

  // 更新系统设置
  static async updateSettings(settings: UpdateSettingsRequest): Promise<SystemSettings> {
    // 模拟网络延迟
    await new Promise(resolve => setTimeout(resolve, 300));
    
    // 更新Mock数据
    this.mockSettings = {
      ...this.mockSettings,
      ...settings
    };
    
    return Promise.resolve(this.mockSettings);
  }

  // 获取系统信息
  static async getSystemInfo(): Promise<SystemInfo> {
    // Mock系统信息
    await new Promise(resolve => setTimeout(resolve, 200));
    return Promise.resolve({
      version: '1.0.0',
      buildTime: '2025-01-09T10:00:00Z',
      startTime: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
      serverTime: new Date().toISOString(),
      uptime: 24 * 60 * 60, // 1天
      platform: 'linux',
      nodeVersion: '18.17.0',
      memoryUsage: {
        used: 512 * 1024 * 1024, // 512MB
        total: 2 * 1024 * 1024 * 1024, // 2GB
        free: 1.5 * 1024 * 1024 * 1024 // 1.5GB
      },
      diskUsage: {
        used: 50 * 1024 * 1024 * 1024, // 50GB
        total: 500 * 1024 * 1024 * 1024, // 500GB
        free: 450 * 1024 * 1024 * 1024 // 450GB
      },
      services: {
        database: 'running',
        storage: 'running',
        cache: 'running',
        queue: 'running'
      }
    });
  }

  // 系统健康检查
  static async healthCheck(): Promise<{
    status: 'healthy' | 'unhealthy';
    checks: Record<string, boolean>;
    timestamp: string;
  }> {
    await new Promise(resolve => setTimeout(resolve, 150));
    return Promise.resolve({
      status: 'healthy' as const,
      checks: {
        database: true,
        storage: true,
        cache: true,
        queue: true,
        diskSpace: true,
        memory: true
      },
      timestamp: new Date().toISOString()
    });
  }

  // 获取存储配置
  static async getStorageConfig(): Promise<StorageConfig> {
    await new Promise(resolve => setTimeout(resolve, 200));
    return Promise.resolve(this.mockSettings.storageConfig);
  }

  // 更新存储配置
  static async updateStorageConfig(config: StorageConfig): Promise<StorageConfig> {
    await new Promise(resolve => setTimeout(resolve, 300));
    this.mockSettings.storageConfig = config;
    return Promise.resolve(config);
  }

  // 测试存储连接
  static async testStorageConnection(config: StorageConfig): Promise<{
    success: boolean;
    message: string;
    details?: any;
  }> {
    await new Promise(resolve => setTimeout(resolve, 1000)); // 模拟测试延迟
    return Promise.resolve({
      success: true,
      message: '存储连接测试成功',
      details: {
        type: config.type,
        timestamp: new Date().toISOString()
      }
    });
  }

  // 获取邮件配置
  static async getEmailConfig(): Promise<EmailConfig> {
    await new Promise(resolve => setTimeout(resolve, 200));
    return Promise.resolve(this.mockSettings.emailConfig);
  }

  // 更新邮件配置
  static async updateEmailConfig(config: EmailConfig): Promise<EmailConfig> {
    await new Promise(resolve => setTimeout(resolve, 300));
    this.mockSettings.emailConfig = config;
    return Promise.resolve(config);
  }

  // 测试邮件发送
  static async testEmail(config: EmailConfig, testEmail: string): Promise<{
    success: boolean;
    message: string;
  }> {
    await new Promise(resolve => setTimeout(resolve, 2000)); // 模拟邮件发送延迟
    return Promise.resolve({
      success: true,
      message: `测试邮件已发送到 ${testEmail}`
    });
  }

  // 重启系统服务
  static async restartService(serviceName: string): Promise<{
    success: boolean;
    message: string;
  }> {
    await new Promise(resolve => setTimeout(resolve, 1500));
    return Promise.resolve({
      success: true,
      message: `服务 ${serviceName} 重启成功`
    });
  }

  // 清理系统缓存
  static async clearCache(cacheType?: 'all' | 'thumbnails' | 'sessions' | 'temp'): Promise<{
    success: boolean;
    message: string;
    cleared_items: number;
  }> {
    await new Promise(resolve => setTimeout(resolve, 800));
    const itemCount = Math.floor(Math.random() * 1000) + 100;
    return Promise.resolve({
      success: true,
      message: `${cacheType || 'all'} 缓存清理完成`,
      cleared_items: itemCount
    });
  }

  // 导出系统设置
  static async exportSettings(): Promise<Blob> {
    await new Promise(resolve => setTimeout(resolve, 500)); // 模拟网络延迟
    
    // 创建Mock配置文件内容
    const settingsData = {
      exportedAt: new Date().toISOString(),
      version: '1.0.0',
      settings: this.mockSettings
    };
    
    const jsonString = JSON.stringify(settingsData, null, 2);
    const blob = new Blob([jsonString], { type: 'application/json' });
    return Promise.resolve(blob);
  }

  // 导入系统设置
  static async importSettings(file: File): Promise<{
    success: boolean;
    message: string;
    imported_settings: number;
  }> {
    await new Promise(resolve => setTimeout(resolve, 800)); // 模拟文件处理时间
    
    try {
      // 模拟文件读取和解析
      const settingsCount = Math.floor(Math.random() * 20) + 10; // 10-30个设置项
      
      // 模拟成功导入
      return Promise.resolve({
        success: true,
        message: `成功导入 ${settingsCount} 个配置项`,
        imported_settings: settingsCount
      });
    } catch (error) {
      return Promise.resolve({
        success: false,
        message: '配置文件格式错误',
        imported_settings: 0
      });
    }
  }

  // 重置设置到默认值
  static async resetToDefaults(): Promise<SystemSettings> {
    await new Promise(resolve => setTimeout(resolve, 600)); // 模拟重置延迟
    
    // 重置到默认配置
    const defaultSettings: SystemSettings = {
      siteName: '云存储管理后台',
      defaultUserQuota: 5 * 1024 * 1024 * 1024, // 5GB 默认配额
      allowRegistration: false,
      maxFileSize: 100 * 1024 * 1024, // 100MB
      supportedFileTypes: [
        'image/jpeg', 'image/png', 'image/gif',
        'application/pdf', 'text/plain'
      ],
      enableThumbnails: true,
      thumbnailSizes: ['small', 'medium'],
      enableFileVersioning: false,
      maxFileVersions: 3,
      enableEmailNotifications: false,
      emailConfig: {
        smtpHost: '',
        smtpPort: 587,
        smtpUser: '',
        smtpPassword: '',
        smtpSecure: true,
        fromName: '系统通知',
        fromEmail: 'system@example.com'
      },
      storageConfig: {
        type: 'local',
        localPath: './uploads',
        maxStorageSize: 500 * 1024 * 1024 * 1024, // 500GB
        s3Config: {
          accessKeyId: '',
          secretAccessKey: '',
          region: 'us-east-1',
          bucket: '',
          endpoint: ''
        }
      },
      theme: {
        primaryColor: '#1677FF',
        darkMode: false
      },
      security: {
        sessionTimeout: 12 * 60, // 12小时
        enableTwoFactor: false,
        passwordMinLength: 8,
        passwordRequireSpecialChar: true
      }
    };
    
    // 更新Mock数据
    this.mockSettings = defaultSettings;
    return Promise.resolve(defaultSettings);
  }

  // 获取系统日志
  static async getSystemLogs(params?: {
    level?: string;
    service?: string;
    limit?: number;
    offset?: number;
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
    await new Promise(resolve => setTimeout(resolve, 400)); // 模拟日志查询时间
    
    // Mock日志数据
    const services = ['user-service', 'file-service', 'api-gateway', 'auth-service'];
    const levels = ['info', 'warn', 'error', 'debug'];
    const messages = [
      'User login successful',
      'File upload completed',
      'Authentication token refreshed',
      'Database connection established',
      'Cache cleared successfully',
      'Backup process completed',
      'System health check passed',
      'Configuration updated',
      'Service restart completed'
    ];
    
    const limit = params?.limit || 50;
    const offset = params?.offset || 0;
    const mockLogs = [];
    
    // 生成Mock日志条目
    for (let i = 0; i < limit; i++) {
      const timestamp = new Date(Date.now() - (i + offset) * 60000); // 每分钟一条日志
      mockLogs.push({
        timestamp: timestamp.toISOString(),
        level: params?.level || levels[Math.floor(Math.random() * levels.length)],
        service: params?.service || services[Math.floor(Math.random() * services.length)],
        message: messages[Math.floor(Math.random() * messages.length)],
        details: {
          user_id: Math.floor(Math.random() * 100) + 1,
          ip_address: `192.168.1.${Math.floor(Math.random() * 255)}`,
          request_id: `req_${Math.random().toString(36).substring(7)}`
        }
      });
    }
    
    return Promise.resolve({
      logs: mockLogs,
      total: 1500 + Math.floor(Math.random() * 500) // 模拟总数
    });
  }

  // 下载系统日志
  static async downloadLogs(service?: string, date?: string): Promise<Blob> {
    await new Promise(resolve => setTimeout(resolve, 1000)); // 模拟日志打包时间
    
    // 生成Mock日志内容
    const logLines = [];
    logLines.push(`# System Logs Export`);
    logLines.push(`# Generated: ${new Date().toISOString()}`);
    if (service) logLines.push(`# Service: ${service}`);
    if (date) logLines.push(`# Date: ${date}`);
    logLines.push('# Format: [timestamp] [level] [service] message');
    logLines.push('');
    
    // 添加一些示例日志行
    for (let i = 0; i < 100; i++) {
      const timestamp = new Date(Date.now() - i * 60000).toISOString();
      const level = ['INFO', 'WARN', 'ERROR'][Math.floor(Math.random() * 3)];
      const svc = service || ['user-service', 'file-service'][Math.floor(Math.random() * 2)];
      const message = [
        'Request processed successfully',
        'File operation completed',
        'User authentication verified',
        'Cache invalidation triggered'
      ][Math.floor(Math.random() * 4)];
      
      logLines.push(`[${timestamp}] [${level}] [${svc}] ${message}`);
    }
    
    const logContent = logLines.join('\n');
    const blob = new Blob([logContent], { type: 'text/plain' });
    return Promise.resolve(blob);
  }
}

// 导出服务实例
export const settingsService = SettingsService;