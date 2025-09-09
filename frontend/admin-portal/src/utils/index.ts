import { FILE_SIZE_UNITS, FILE_TYPE_ICONS } from '@constants/index';

// 格式化文件大小
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  
  const k = 1024;
  const sizes = FILE_SIZE_UNITS;
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

// 格式化日期时间
export const formatDateTime = (dateStr: string): string => {
  const date = new Date(dateStr);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
};

// 格式化日期（通用版本）
export const formatDate = (dateStr: string, format?: string): string => {
  const date = new Date(dateStr);
  
  if (format === 'MM-DD') {
    return date.toLocaleDateString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
    });
  }
  
  if (format === 'YYYY-MM-DD') {
    return date.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    });
  }
  
  // 默认格式：YYYY-MM-DD HH:mm
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
};

// 格式化相对时间
export const formatRelativeTime = (dateStr: string): string => {
  const date = new Date(dateStr);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  
  const minutes = Math.floor(diff / (1000 * 60));
  const hours = Math.floor(diff / (1000 * 60 * 60));
  const days = Math.floor(diff / (1000 * 60 * 60 * 24));
  
  if (minutes < 1) return '刚刚';
  if (minutes < 60) return `${minutes}分钟前`;
  if (hours < 24) return `${hours}小时前`;
  if (days < 7) return `${days}天前`;
  
  return formatDateTime(dateStr);
};

// 获取文件类型图标
export const getFileTypeIcon = (contentType: string): string => {
  return FILE_TYPE_ICONS[contentType as keyof typeof FILE_TYPE_ICONS] || FILE_TYPE_ICONS.default;
};

// 获取文件扩展名
export const getFileExtension = (filename: string): string => {
  return filename.slice((filename.lastIndexOf('.') - 1 >>> 0) + 2);
};

// 检查文件类型是否为图片
export const isImageType = (contentType: string): boolean => {
  return contentType.startsWith('image/');
};

// 检查文件类型是否为视频
export const isVideoType = (contentType: string): boolean => {
  return contentType.startsWith('video/');
};

// 检查文件类型是否为音频
export const isAudioType = (contentType: string): boolean => {
  return contentType.startsWith('audio/');
};

// 生成随机字符串
export const generateRandomString = (length: number = 8): string => {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
};

// 深拷贝对象
export const deepClone = <T>(obj: T): T => {
  if (obj === null || typeof obj !== 'object') return obj;
  if (obj instanceof Date) return new Date(obj.getTime()) as any;
  if (obj instanceof Array) return obj.map(item => deepClone(item)) as any;
  
  const cloned = {} as T;
  Object.keys(obj).forEach(key => {
    (cloned as any)[key] = deepClone((obj as any)[key]);
  });
  
  return cloned;
};

// 防抖函数
export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  delay: number
): ((...args: Parameters<T>) => void) => {
  let timeoutId: NodeJS.Timeout;
  
  return (...args: Parameters<T>) => {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => func(...args), delay);
  };
};

// 节流函数
export const throttle = <T extends (...args: any[]) => any>(
  func: T,
  delay: number
): ((...args: Parameters<T>) => void) => {
  let lastCall = 0;
  
  return (...args: Parameters<T>) => {
    const now = Date.now();
    if (now - lastCall >= delay) {
      lastCall = now;
      func(...args);
    }
  };
};

// 计算文件hash值(用于上传)
export const calculateFileHash = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = async (e) => {
      try {
        const arrayBuffer = e.target?.result as ArrayBuffer;
        const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer);
        const hashArray = Array.from(new Uint8Array(hashBuffer));
        const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
        resolve(hashHex);
      } catch (error) {
        reject(error);
      }
    };
    reader.onerror = reject;
    reader.readAsArrayBuffer(file);
  });
};

// 验证邮箱格式
export const isValidEmail = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

// 验证用户名格式
export const isValidUsername = (username: string): boolean => {
  const usernameRegex = /^[a-zA-Z0-9_-]{3,20}$/;
  return usernameRegex.test(username);
};

// 验证密码强度
export const validatePassword = (password: string): {
  isValid: boolean;
  messages: string[];
} => {
  const messages: string[] = [];
  let isValid = true;
  
  if (password.length < 8) {
    messages.push('密码长度至少8位');
    isValid = false;
  }
  
  if (!/[a-z]/.test(password)) {
    messages.push('密码必须包含小写字母');
    isValid = false;
  }
  
  if (!/[A-Z]/.test(password)) {
    messages.push('密码必须包含大写字母');
    isValid = false;
  }
  
  if (!/\d/.test(password)) {
    messages.push('密码必须包含数字');
    isValid = false;
  }
  
  if (!/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
    messages.push('密码必须包含特殊字符');
    isValid = false;
  }
  
  return { isValid, messages };
};

// 下载文件
export const downloadFile = (url: string, filename: string): void => {
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
};

// 复制到剪贴板
export const copyToClipboard = (text: string): Promise<void> => {
  if (navigator.clipboard) {
    return navigator.clipboard.writeText(text);
  } else {
    // 备用方案
    const textArea = document.createElement('textarea');
    textArea.value = text;
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    document.execCommand('copy');
    document.body.removeChild(textArea);
    return Promise.resolve();
  }
};

// 获取URL查询参数
export const getQueryParams = (): Record<string, string> => {
  const params: Record<string, string> = {};
  const urlParams = new URLSearchParams(window.location.search);
  
  urlParams.forEach((value, key) => {
    params[key] = value;
  });
  
  return params;
};

// 设置URL查询参数
export const setQueryParams = (params: Record<string, string | number | undefined>): void => {
  const url = new URL(window.location.href);
  
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined) {
      url.searchParams.set(key, String(value));
    } else {
      url.searchParams.delete(key);
    }
  });
  
  window.history.replaceState({}, '', url.toString());
};

// 格式化百分比
export const formatPercentage = (value: number, total: number): string => {
  if (total === 0) return '0%';
  const percentage = (value / total * 100).toFixed(1);
  return `${percentage}%`;
};

// 生成颜色
export const generateColor = (str: string): string => {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
  }
  
  const hue = hash % 360;
  return `hsl(${hue}, 70%, 60%)`;
};

// 休眠函数
export const sleep = (ms: number): Promise<void> => {
  return new Promise(resolve => setTimeout(resolve, ms));
};

// 格式化持续时间（毫秒）
export const formatDuration = (ms: number): string => {
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);
  
  if (days > 0) {
    return `${days}天${hours % 24}小时`;
  } else if (hours > 0) {
    return `${hours}小时${minutes % 60}分钟`;
  } else if (minutes > 0) {
    return `${minutes}分钟${seconds % 60}秒`;
  } else {
    return `${seconds}秒`;
  }
};