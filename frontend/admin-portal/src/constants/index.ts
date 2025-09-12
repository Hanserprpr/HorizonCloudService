// API相关常量
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8001';
export const USER_SERVICE_URL = import.meta.env.VITE_USER_SERVICE_URL || 'http://localhost:8001';
export const FILE_SERVICE_URL = import.meta.env.VITE_FILE_SERVICE_URL || 'http://localhost:8002';
export const SYSTEM_SERVICE_URL = import.meta.env.VITE_SYSTEM_SERVICE_URL || 'http://localhost:8003';

// 应用配置常量
export const APP_NAME = import.meta.env.VITE_APP_NAME || '云存储管理后台';
export const APP_VERSION = import.meta.env.VITE_APP_VERSION || '1.0.0';

// 上传相关常量
export const UPLOAD_CONFIG = {
  CHUNK_SIZE: parseInt(import.meta.env.VITE_UPLOAD_CHUNK_SIZE) || 5 * 1024 * 1024, // 5MB
  MAX_CONCURRENT: parseInt(import.meta.env.VITE_UPLOAD_MAX_CONCURRENT) || 3,
  RETRY_TIMES: parseInt(import.meta.env.VITE_UPLOAD_RETRY_TIMES) || 3,
  SUPPORTED_TYPES: [
    'image/jpeg',
    'image/png',
    'image/gif',
    'image/webp',
    'image/svg+xml',
    'video/mp4',
    'video/avi',
    'video/mov',
    'video/wmv',
    'audio/mp3',
    'audio/wav',
    'audio/flac',
    'application/pdf',
    'text/plain',
    'application/msword',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    'application/vnd.ms-excel',
    'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  ],
  MAX_FILE_SIZE: 2 * 1024 * 1024 * 1024, // 2GB
};

// 文件类型映射
export const FILE_TYPE_ICONS = {
  // 图片
  'image/jpeg': 'file-image',
  'image/png': 'file-image',
  'image/gif': 'file-image',
  'image/webp': 'file-image',
  'image/svg+xml': 'file-image',
  
  // 视频
  'video/mp4': 'video-camera',
  'video/avi': 'video-camera',
  'video/mov': 'video-camera',
  'video/wmv': 'video-camera',
  
  // 音频
  'audio/mp3': 'audio',
  'audio/wav': 'audio',
  'audio/flac': 'audio',
  
  // 文档
  'application/pdf': 'file-pdf',
  'text/plain': 'file-text',
  'application/msword': 'file-word',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document': 'file-word',
  'application/vnd.ms-excel': 'file-excel',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': 'file-excel',
  
  // 默认
  'default': 'file',
};

// 文件大小格式化
export const FILE_SIZE_UNITS = ['B', 'KB', 'MB', 'GB', 'TB'];

// 状态常量
export const USER_STATUS = {
  INACTIVE: 0,
  ACTIVE: 1,
  SUSPENDED: 2,
  DELETED: 3,
} as const;

export const FILE_STATUS = {
  DELETED: 0,
  ACTIVE: 1,
  PROCESSING: 2,
  FAILED: 3,
} as const;

export const UPLOAD_STATUS = {
  PENDING: 1,
  UPLOADING: 2,
  COMPLETED: 3,
  FAILED: 4,
  CANCELLED: 5,
} as const;

// 用户角色
export const USER_ROLES = {
  ADMIN: 'admin',
  USER: 'user',
  GUEST: 'guest',
} as const;

// 分页配置
export const PAGINATION = {
  DEFAULT_PAGE_SIZE: 20,
  PAGE_SIZE_OPTIONS: ['10', '20', '50', '100'],
  SHOW_SIZE_CHANGER: true,
  SHOW_QUICK_JUMPER: true,
};

// 主题配置
export const THEME_CONFIG = {
  PRIMARY_COLOR: '#1677FF',
  SUCCESS_COLOR: '#52C41A',
  WARNING_COLOR: '#FAAD14',
  ERROR_COLOR: '#FF4D4F',
  INFO_COLOR: '#1677FF',
};

// 布局配置
export const LAYOUT_CONFIG = {
  SIDEBAR_WIDTH: 240,
  SIDEBAR_COLLAPSED_WIDTH: 80,
  HEADER_HEIGHT: 64,
  FOOTER_HEIGHT: 48,
};

// 路由路径常量
export const ROUTES = {
  LOGIN: '/auth/login',
  DASHBOARD: '/dashboard',
  FILES: '/files',
  USERS: '/users',
  SETTINGS: '/settings',
  PROFILE: '/profile',
  AUTH: {
    LOGIN: '/auth/login',
  },
} as const;

// 本地存储键名
export const STORAGE_KEYS = {
  AUTH_TOKEN: 'auth-token',
  REFRESH_TOKEN: 'refresh-token',
  USER_INFO: 'user-info',
  THEME: 'theme',
  SIDEBAR_COLLAPSED: 'sidebar-collapsed',
  LANGUAGE: 'language',
} as const;

// HTTP状态码
export const HTTP_STATUS = {
  OK: 200,
  CREATED: 201,
  NO_CONTENT: 204,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  UNPROCESSABLE_ENTITY: 422,
  INTERNAL_SERVER_ERROR: 500,
} as const;

// 错误消息
export const ERROR_MESSAGES = {
  NETWORK_ERROR: '网络连接失败，请检查网络设置',
  UNAUTHORIZED: '登录已过期，请重新登录',
  FORBIDDEN: '没有权限执行此操作',
  NOT_FOUND: '请求的资源不存在',
  SERVER_ERROR: '服务器内部错误，请稍后重试',
  UPLOAD_FAILED: '文件上传失败',
  INVALID_FILE_TYPE: '不支持的文件类型',
  FILE_TOO_LARGE: '文件大小超过限制',
  FORM_VALIDATION_ERROR: '表单验证失败，请检查输入内容',
} as const;

// 成功消息
export const SUCCESS_MESSAGES = {
  LOGIN_SUCCESS: '登录成功',
  LOGOUT_SUCCESS: '已安全退出',
  SAVE_SUCCESS: '保存成功',
  DELETE_SUCCESS: '删除成功',
  UPLOAD_SUCCESS: '上传完成',
  UPDATE_SUCCESS: '更新成功',
  CREATE_SUCCESS: '创建成功',
} as const;