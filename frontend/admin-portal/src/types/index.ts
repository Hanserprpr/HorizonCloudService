// 基础类型定义
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
  timestamp: string;
}

// 分页相关类型
export interface PaginationParams {
  page?: number;
  size?: number;
  sort?: string;
  order?: 'asc' | 'desc';
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  size: number;
  pages: number;
}

// 用户相关类型
export interface User {
  id: number;
  student_id: string;
  email: string;
  display_name?: string;
  status: number;
  role: string;
  storage_quota: number;
  storage_used: number;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
}

export interface LoginForm {
  student_id: string;
  password: string;
  remember?: boolean;
}

export interface LoginResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface CreateUserRequest {
  student_id: string;
  email: string;
  password: string;
  display_name?: string;
  role?: string;
  storage_quota?: number;
  status?: number;
}

export interface UpdateUserRequest {
  student_id?: string;
  email?: string;
  display_name?: string;
  role?: string;
  storage_quota?: number;
  status?: number;
}

// 文件相关类型
export interface FileItem {
  id: number;
  name: string;
  original_name: string;
  path: string;
  size: number;
  content_type: string;
  extension: string;
  hash: string;
  user_id: number;
  folder_id?: number;
  status: number;
  category?: string;
  is_public: boolean;
  version: number;
  download_count: number;
  created_at: string;
  updated_at: string;
  thumbnail_url?: string;
  download_url?: string;
}

export interface FolderItem {
  id: number;
  name: string;
  path: string;
  description?: string;
  parent_id?: number;
  level: number;
  user_id: number;
  status: number;
  is_system: boolean;
  is_shared: boolean;
  file_count: number;
  folder_count: number;
  total_size: number;
  created_at: string;
  updated_at: string;
}

// 上传相关类型
export interface UploadSession {
  session_id: string;
  file_name: string;
  file_size: number;
  content_type: string;
  user_id: number;
  folder_id?: number;
  chunk_size: number;
  total_chunks: number;
  uploaded_chunks: number;
  status: number;
  progress: number;
  created_at: string;
  updated_at: string;
  expires_at: string;
}

export interface UploadConfig {
  chunkSize: number;
  maxConcurrent: number;
  retryTimes: number;
}

export const UploadStatus = {
  Pending: 1,
  Uploading: 2,
  Completed: 3,
  Failed: 4,
  Cancelled: 5,
} as const;

// 统计相关类型
export interface SystemStats {
  total_users: number;
  total_files: number;
  total_storage: number;
  storage_used: number;
  active_uploads: number;
}

export interface UserStats {
  file_count: number;
  folder_count: number;
  storage_used: number;
  storage_quota: number;
  recent_uploads: number;
}

// 缩略图相关类型
export interface Thumbnail {
  id: number;
  file_id: number;
  size: string;
  width: number;
  height: number;
  path: string;
  file_size: number;
  content_type: string;
  status: number;
  download_url?: string;
  created_at: string;
}

// 错误类型
export interface AppError {
  code: number;
  message: string;
  details?: any;
}

// 表单相关类型
export interface FormState<T> {
  data: T;
  loading: boolean;
  errors: Record<string, string>;
}

// UI状态类型
export interface UIState {
  sidebarCollapsed: boolean;
  theme: 'light' | 'dark';
  loading: boolean;
  currentPath: string[];
}

// 搜索和过滤类型
export interface SearchParams {
  keyword?: string;
  type?: string;
  extension?: string;
  size_min?: number;
  size_max?: number;
  date_from?: string;
  date_to?: string;
  folder_id?: number;
}

export interface BatchOperation {
  type: 'delete' | 'move' | 'copy';
  target_ids: number[];
  destination_folder_id?: number;
}

// 系统设置相关类型
export interface SystemSettings {
  // 基本设置
  system_name: string;
  system_description?: string;
  system_version: string;
  enable_registration: boolean;
  maintenance_mode: boolean;
  
  // 存储设置
  default_storage_quota: number;
  max_storage_quota: number;
  storage_backend: 'local' | 'minio' | 's3' | 'oss';
  enable_file_deduplication: boolean;
  enable_thumbnail_generation: boolean;
  thumbnail_quality: number;
  max_file_size: number;
  allowed_file_types: string[];
  
  // 安全设置
  require_email_verification: boolean;
  password_min_length: number;
  password_complexity: boolean;
  session_timeout: number;
  max_login_attempts: number;
  enable_logging: boolean;
  log_level: 'debug' | 'info' | 'warn' | 'error';
  
  // AI设置
  enable_ai_analysis: boolean;
  ai_analysis_queue_size: number;
  enable_semantic_search: boolean;
  auto_tagging_enabled: boolean;
  
  // 通知设置
  enable_email_notifications: boolean;
  notification_email: string;
  enable_system_alerts: boolean;
}

export interface SystemInfo {
  version: string;
  build_time: string;
  go_version: string;
  os: string;
  arch: string;
  uptime: number;
  
  // 运行时信息
  memory_usage: {
    total: number;
    used: number;
    available: number;
  };
  
  // 存储信息
  storage_info: {
    total_space: number;
    used_space: number;
    free_space: number;
    files_count: number;
  };
  
  // 数据库信息
  database_info: {
    type: string;
    version: string;
    size: number;
    connections: number;
  };
  
  // 服务状态
  services_status: {
    user_service: boolean;
    file_service: boolean;
    ai_service?: boolean;
    search_service?: boolean;
  };
}

export interface UpdateSettingsRequest {
  [key: string]: any;
}

// 健康检查类型
export interface HealthCheck {
  status: 'healthy' | 'unhealthy';
  checks: Record<string, boolean>;
  timestamp: string;
}

// 服务重启请求类型
export interface ServiceRestartRequest {
  service_name: string;
}

// 缓存清理请求类型
export interface ClearCacheRequest {
  cache_type?: 'all' | 'thumbnails' | 'sessions' | 'temp';
}

// 日志下载请求类型
export interface DownloadLogsRequest {
  service?: string;
  date?: string;
}

// 存储配置类型
export interface StorageConfig {
  backend: 'local' | 'minio' | 's3' | 'oss';
  
  // 本地存储
  path?: string;
  
  // MinIO配置
  endpoint?: string;
  access_key?: string;
  secret_key?: string;
  bucket?: string;
  region?: string;
  use_ssl?: boolean;
  path_style?: boolean;
  
  // S3配置（AWS）
  aws_region?: string;
  aws_access_key_id?: string;
  aws_secret_access_key?: string;
  aws_bucket?: string;
  
  // OSS配置（阿里云）
  oss_endpoint?: string;
  oss_access_key_id?: string;
  oss_access_key_secret?: string;
  oss_bucket?: string;
  oss_region?: string;
}

// 邮件配置类型
export interface EmailConfig {
  provider: 'smtp' | 'sendgrid' | 'mailgun' | 'ses';
  config: {
    // SMTP配置
    host?: string;
    port?: number;
    username?: string;
    password?: string;
    from_address?: string;
    use_tls?: boolean;
    
    // SendGrid配置
    sendgrid_api_key?: string;
    sendgrid_from?: string;
    
    // Mailgun配置
    mailgun_domain?: string;
    mailgun_api_key?: string;
    mailgun_from?: string;
    
    // AWS SES配置
    ses_region?: string;
    ses_access_key?: string;
    ses_secret_key?: string;
    ses_from?: string;
  };
}