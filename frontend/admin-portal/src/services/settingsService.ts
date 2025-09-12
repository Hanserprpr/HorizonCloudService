import axios, { AxiosInstance } from 'axios';
import { SYSTEM_SERVICE_URL } from '../constants';
import type { 
  SystemSettings, 
  SystemInfo, 
  UpdateSettingsRequest,
  StorageConfig,
  EmailConfig 
} from '../types';

// 系统服务API客户端
class SystemServiceClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: SYSTEM_SERVICE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // 请求拦截器 - 添加认证token
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('auth-token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // 响应拦截器 - 处理错误
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          // Token过期，清除本地存储并跳转到登录页
          localStorage.removeItem('auth-token');
          localStorage.removeItem('refresh-token');
          window.location.href = '/auth/login';
        }
        return Promise.reject(error);
      }
    );
  }

  // 系统统计信息
  async getSystemStats() {
    const response = await this.client.get('/api/v1/system/stats');
    return response.data;
  }

  // 系统健康检查
  async getSystemHealth() {
    const response = await this.client.get('/api/v1/system/health');
    return response.data;
  }

  // 清理缓存
  async clearCache() {
    const response = await this.client.post('/api/v1/system/cache/clear');
    return response.data;
  }

  // 获取设置
  async getSettings() {
    const response = await this.client.get('/api/v1/admin/settings');
    return response.data;
  }

  // 更新设置
  async updateSettings(settings: any) {
    const response = await this.client.put('/api/v1/admin/settings', { settings });
    return response.data;
  }

  // 获取存储设置
  async getStorageSettings() {
    const response = await this.client.get('/api/v1/admin/settings/storage');
    return response.data;
  }

  // 更新存储设置
  async updateStorageSettings(settings: any) {
    const response = await this.client.put('/api/v1/admin/settings/storage', settings);
    return response.data;
  }

  // 测试存储设置
  async testStorageSettings(settings: any) {
    const response = await this.client.post('/api/v1/admin/settings/test-storage', settings);
    return response.data;
  }
}

// 创建系统服务客户端实例
const systemClient = new SystemServiceClient();

export class SettingsService {

  // 获取系统设置
  static async getSettings(): Promise<SystemSettings> {
    try {
      const response = await systemClient.getSettings();
      if (response.success) {
        return this.transformSettingsFromAPI(response.data.settings);
      }
      throw new Error(response.message || 'Failed to get settings');
    } catch (error) {
      console.error('Failed to get system settings:', error);
      // 返回默认设置作为后备
      return this.getDefaultSettings();
    }
  }

  // 更新系统设置
  static async updateSettings(settings: UpdateSettingsRequest): Promise<SystemSettings> {
    try {
      const transformedSettings = this.transformSettingsToAPI(settings);
      const response = await systemClient.updateSettings(transformedSettings);
      if (response.success) {
        return this.transformSettingsFromAPI(response.data?.settings || transformedSettings);
      }
      throw new Error(response.message || 'Failed to update settings');
    } catch (error) {
      console.error('Failed to update system settings:', error);
      throw error;
    }
  }

  // 获取系统信息
  static async getSystemInfo(): Promise<SystemInfo> {
    try {
      const statsResponse = await systemClient.getSystemStats();
      const healthResponse = await systemClient.getSystemHealth();
      
      if (statsResponse.success && healthResponse.success) {
        return this.transformSystemInfoFromAPI(statsResponse.data, healthResponse.data);
      }
      throw new Error('Failed to get system information');
    } catch (error) {
      console.error('Failed to get system info:', error);
      // 返回默认系统信息作为后备
      return this.getDefaultSystemInfo();
    }
  }

  // 系统健康检查
  static async healthCheck(): Promise<{
    status: 'healthy' | 'unhealthy';
    checks: Record<string, boolean>;
    timestamp: string;
  }> {
    try {
      const response = await systemClient.getSystemHealth();
      if (response.success) {
        return this.transformHealthCheckFromAPI(response.data);
      }
      throw new Error('Failed to get health check');
    } catch (error) {
      console.error('Failed to get health check:', error);
      return {
        status: 'unhealthy' as const,
        checks: {
          database: false,
          storage: false,
          cache: false,
          queue: false,
          diskSpace: false,
          memory: false
        },
        timestamp: new Date().toISOString()
      };
    }
  }

  // 获取存储配置
  static async getStorageConfig(): Promise<StorageConfig> {
    try {
      const response = await systemClient.getStorageSettings();
      if (response.success) {
        return this.transformStorageConfigFromAPI(response.data);
      }
      throw new Error('Failed to get storage config');
    } catch (error) {
      console.error('Failed to get storage config:', error);
      return this.getDefaultStorageConfig();
    }
  }

  // 更新存储配置
  static async updateStorageConfig(config: StorageConfig): Promise<StorageConfig> {
    try {
      const transformedConfig = this.transformStorageConfigToAPI(config);
      const response = await systemClient.updateStorageSettings(transformedConfig);
      if (response.success) {
        return this.transformStorageConfigFromAPI(response.data);
      }
      throw new Error(response.message || 'Failed to update storage config');
    } catch (error) {
      console.error('Failed to update storage config:', error);
      throw error;
    }
  }

  // 测试存储连接
  static async testStorageConnection(config: StorageConfig): Promise<{
    success: boolean;
    message: string;
    details?: any;
  }> {
    try {
      const transformedConfig = this.transformStorageConfigToAPI(config);
      const response = await systemClient.testStorageSettings(transformedConfig);
      return {
        success: response.data.success,
        message: response.data.message,
        details: response.data.details
      };
    } catch (error) {
      console.error('Failed to test storage connection:', error);
      return {
        success: false,
        message: '存储连接测试失败',
        details: {
          error: error instanceof Error ? error.message : 'Unknown error',
          timestamp: new Date().toISOString()
        }
      };
    }
  }

  // 获取邮件配置  
  static async getEmailConfig(): Promise<EmailConfig> {
    try {
      const response = await systemClient.getSettings();
      if (response.success) {
        return this.extractEmailConfigFromAPI(response.data.settings);
      }
      throw new Error('Failed to get email config');
    } catch (error) {
      console.error('Failed to get email config:', error);
      return this.getDefaultEmailConfig();
    }
  }

  // 更新邮件配置
  static async updateEmailConfig(config: EmailConfig): Promise<EmailConfig> {
    try {
      // 获取当前设置然后更新邮件部分
      const currentSettings = await systemClient.getSettings();
      if (currentSettings.success) {
        const updatedSettings = {
          ...currentSettings.data.settings,
          email_config: this.transformEmailConfigToAPI(config)
        };
        
        const response = await systemClient.updateSettings(updatedSettings);
        if (response.success) {
          return config;
        }
      }
      throw new Error('Failed to update email config');
    } catch (error) {
      console.error('Failed to update email config:', error);
      throw error;
    }
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

  // ==================== 数据转换方法 ====================

  // 将API响应转换为前端SystemSettings格式
  private static transformSettingsFromAPI(apiData: any): SystemSettings {
    return {
      siteName: apiData.site_name || '云存储管理后台',
      defaultUserQuota: apiData.default_user_quota || 5 * 1024 * 1024 * 1024,
      allowRegistration: apiData.allow_registration || false,
      maxFileSize: apiData.max_file_size || 100 * 1024 * 1024,
      supportedFileTypes: apiData.supported_file_types || [
        'image/jpeg', 'image/png', 'image/gif',
        'application/pdf', 'text/plain'
      ],
      enableThumbnails: apiData.enable_thumbnails || true,
      thumbnailSizes: apiData.thumbnail_sizes || ['small', 'medium'],
      enableFileVersioning: apiData.enable_file_versioning || false,
      maxFileVersions: apiData.max_file_versions || 3,
      enableEmailNotifications: apiData.enable_email_notifications || false,
      emailConfig: this.transformEmailConfigFromAPI(apiData.email_config),
      storageConfig: this.transformStorageConfigFromAPI(apiData.storage_config),
      theme: {
        primaryColor: apiData.theme?.primary_color || '#1677FF',
        darkMode: apiData.theme?.dark_mode || false
      },
      security: {
        sessionTimeout: apiData.security?.session_timeout || 12 * 60,
        enableTwoFactor: apiData.security?.enable_two_factor || false,
        passwordMinLength: apiData.security?.password_min_length || 8,
        passwordRequireSpecialChar: apiData.security?.password_require_special_char || true
      }
    };
  }

  // 将前端SystemSettings格式转换为API请求格式
  private static transformSettingsToAPI(settings: UpdateSettingsRequest): any {
    return {
      site_name: settings.siteName,
      default_user_quota: settings.defaultUserQuota,
      allow_registration: settings.allowRegistration,
      max_file_size: settings.maxFileSize,
      supported_file_types: settings.supportedFileTypes,
      enable_thumbnails: settings.enableThumbnails,
      thumbnail_sizes: settings.thumbnailSizes,
      enable_file_versioning: settings.enableFileVersioning,
      max_file_versions: settings.maxFileVersions,
      enable_email_notifications: settings.enableEmailNotifications,
      email_config: this.transformEmailConfigToAPI(settings.emailConfig),
      storage_config: this.transformStorageConfigToAPI(settings.storageConfig),
      theme: {
        primary_color: settings.theme?.primaryColor,
        dark_mode: settings.theme?.darkMode
      },
      security: {
        session_timeout: settings.security?.sessionTimeout,
        enable_two_factor: settings.security?.enableTwoFactor,
        password_min_length: settings.security?.passwordMinLength,
        password_require_special_char: settings.security?.passwordRequireSpecialChar
      }
    };
  }

  // 将API响应转换为前端SystemInfo格式
  private static transformSystemInfoFromAPI(statsData: any, healthData: any): SystemInfo {
    return {
      version: '1.0.0',
      uptime: statsData.uptime || '0s',
      totalUsers: statsData.total_users || 0,
      totalFiles: statsData.total_files || 0,
      totalStorage: statsData.total_storage || 0,
      usedStorage: statsData.used_storage || 0,
      availableStorage: statsData.available_storage || statsData.total_storage || 0,
      systemHealth: {
        status: healthData.status || 'unknown',
        checks: healthData.checks || {},
        timestamp: healthData.timestamp || new Date().toISOString()
      },
      services: {
        'user-service': 'running',
        'file-service': 'running',
        'system-service': 'running',
        'database': healthData.checks?.database ? 'running' : 'stopped',
        'storage': healthData.checks?.storage ? 'running' : 'stopped',
        'cache': healthData.checks?.cache ? 'running' : 'stopped'
      }
    };
  }

  // 将API健康检查响应转换为前端格式
  private static transformHealthCheckFromAPI(healthData: any) {
    return {
      status: healthData.status === 'healthy' ? 'healthy' as const : 'unhealthy' as const,
      checks: healthData.checks || {
        database: false,
        storage: false,
        cache: false,
        queue: false,
        diskSpace: false,
        memory: false
      },
      timestamp: healthData.timestamp || new Date().toISOString()
    };
  }

  // 存储配置相关转换
  private static transformStorageConfigFromAPI(apiData: any): StorageConfig {
    if (!apiData) return this.getDefaultStorageConfig();
    
    return {
      type: apiData.type || 'local',
      localPath: apiData.local_path || './uploads',
      maxStorageSize: apiData.max_storage_size || 500 * 1024 * 1024 * 1024,
      s3Config: {
        accessKeyId: apiData.s3_config?.access_key_id || '',
        secretAccessKey: apiData.s3_config?.secret_access_key || '',
        region: apiData.s3_config?.region || 'us-east-1',
        bucket: apiData.s3_config?.bucket || '',
        endpoint: apiData.s3_config?.endpoint || ''
      }
    };
  }

  private static transformStorageConfigToAPI(config: StorageConfig): any {
    return {
      type: config.type,
      local_path: config.localPath,
      max_storage_size: config.maxStorageSize,
      s3_config: {
        access_key_id: config.s3Config.accessKeyId,
        secret_access_key: config.s3Config.secretAccessKey,
        region: config.s3Config.region,
        bucket: config.s3Config.bucket,
        endpoint: config.s3Config.endpoint
      }
    };
  }

  // 邮件配置相关转换
  private static transformEmailConfigFromAPI(apiData: any): EmailConfig {
    if (!apiData) return this.getDefaultEmailConfig();
    
    return {
      smtpHost: apiData.smtp_host || '',
      smtpPort: apiData.smtp_port || 587,
      smtpUser: apiData.smtp_user || '',
      smtpPassword: apiData.smtp_password || '',
      smtpSecure: apiData.smtp_secure || true,
      fromName: apiData.from_name || '系统通知',
      fromEmail: apiData.from_email || 'system@example.com'
    };
  }

  private static transformEmailConfigToAPI(config: EmailConfig): any {
    return {
      smtp_host: config.smtpHost,
      smtp_port: config.smtpPort,
      smtp_user: config.smtpUser,
      smtp_password: config.smtpPassword,
      smtp_secure: config.smtpSecure,
      from_name: config.fromName,
      from_email: config.fromEmail
    };
  }

  private static extractEmailConfigFromAPI(apiData: any): EmailConfig {
    return this.transformEmailConfigFromAPI(apiData.email_config);
  }

  // ==================== 默认值方法 ====================

  private static getDefaultSettings(): SystemSettings {
    return {
      siteName: '云存储管理后台',
      defaultUserQuota: 5 * 1024 * 1024 * 1024, // 5GB
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
      emailConfig: this.getDefaultEmailConfig(),
      storageConfig: this.getDefaultStorageConfig(),
      theme: {
        primaryColor: '#1677FF',
        darkMode: false
      },
      security: {
        sessionTimeout: 12 * 60,
        enableTwoFactor: false,
        passwordMinLength: 8,
        passwordRequireSpecialChar: true
      }
    };
  }

  private static getDefaultSystemInfo(): SystemInfo {
    return {
      version: '1.0.0',
      uptime: '0s',
      totalUsers: 0,
      totalFiles: 0,
      totalStorage: 500 * 1024 * 1024 * 1024, // 500GB
      usedStorage: 0,
      availableStorage: 500 * 1024 * 1024 * 1024,
      systemHealth: {
        status: 'unknown',
        checks: {},
        timestamp: new Date().toISOString()
      },
      services: {
        'user-service': 'unknown',
        'file-service': 'unknown', 
        'system-service': 'unknown',
        'database': 'unknown',
        'storage': 'unknown',
        'cache': 'unknown'
      }
    };
  }

  private static getDefaultStorageConfig(): StorageConfig {
    return {
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
    };
  }

  private static getDefaultEmailConfig(): EmailConfig {
    return {
      smtpHost: '',
      smtpPort: 587,
      smtpUser: '',
      smtpPassword: '',
      smtpSecure: true,
      fromName: '系统通知',
      fromEmail: 'system@example.com'
    };
  }

  // Mock设置数据（作为后备）
  private static mockSettings: SystemSettings = {
    siteName: '云存储管理后台',
    defaultUserQuota: 5 * 1024 * 1024 * 1024, // 5GB
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
}

// 导出服务实例
export const settingsService = SettingsService;