import React, { useEffect } from 'react';
import { message, notification } from 'antd';
import { 
  ExclamationCircleOutlined, 
  CloseCircleOutlined,
  WarningOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';

interface ErrorEvent extends Event {
  error?: Error;
  message?: string;
  filename?: string;
  lineno?: number;
  colno?: number;
}

interface PromiseRejectionEvent extends Event {
  reason?: any;
  promise?: Promise<any>;
}

/**
 * 全局错误处理组件
 * 监听全局 JavaScript 错误和未处理的 Promise 拒绝
 */
export const GlobalErrorHandler: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  useEffect(() => {
    // 处理全局 JavaScript 错误
    const handleError = (event: ErrorEvent) => {
      const error = event.error || new Error(event.message || 'Unknown error');
      
      console.error('Global JavaScript Error:', {
        error,
        message: event.message,
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno,
        timestamp: new Date().toISOString(),
        userAgent: navigator.userAgent,
        url: window.location.href,
      });

      // 根据错误类型决定通知方式
      if (isNetworkError(error)) {
        handleNetworkError(error);
      } else if (isAuthError(error)) {
        handleAuthError(error);
      } else {
        handleGenericError(error);
      }
    };

    // 处理未捕获的 Promise 拒绝
    const handleUnhandledRejection = (event: PromiseRejectionEvent) => {
      const reason = event.reason;
      
      console.error('Unhandled Promise Rejection:', {
        reason,
        timestamp: new Date().toISOString(),
        userAgent: navigator.userAgent,
        url: window.location.href,
      });

      // 阻止默认的控制台错误输出
      event.preventDefault();

      // 根据拒绝原因决定处理方式
      if (reason instanceof Error) {
        if (isNetworkError(reason)) {
          handleNetworkError(reason);
        } else if (isAuthError(reason)) {
          handleAuthError(reason);
        } else {
          handleGenericError(reason);
        }
      } else if (typeof reason === 'string') {
        handleGenericError(new Error(reason));
      } else {
        handleGenericError(new Error('Unknown promise rejection'));
      }
    };

    // 注册事件监听器
    window.addEventListener('error', handleError);
    window.addEventListener('unhandledrejection', handleUnhandledRejection);

    // 清理函数
    return () => {
      window.removeEventListener('error', handleError);
      window.removeEventListener('unhandledrejection', handleUnhandledRejection);
    };
  }, []);

  return <>{children}</>;
};

// 错误类型判断函数
const isNetworkError = (error: Error): boolean => {
  const networkErrorMessages = [
    'Network Error',
    'NetworkError',
    'Failed to fetch',
    'fetch',
    'ECONNREFUSED',
    'ENOTFOUND',
    'timeout',
    'ERR_NETWORK',
    'ERR_INTERNET_DISCONNECTED',
  ];
  
  return networkErrorMessages.some(msg => 
    error.message?.toLowerCase().includes(msg.toLowerCase()) ||
    error.name?.toLowerCase().includes(msg.toLowerCase())
  );
};

const isAuthError = (error: Error): boolean => {
  const authErrorMessages = [
    'Unauthorized',
    'Authentication',
    'Token expired',
    'Invalid token',
    'Login required',
    '401',
    '403',
  ];
  
  return authErrorMessages.some(msg => 
    error.message?.toLowerCase().includes(msg.toLowerCase())
  );
};

// 错误处理函数
const handleNetworkError = (error: Error) => {
  notification.error({
    message: '网络连接错误',
    description: '请检查您的网络连接并重试。如果问题持续存在，请联系管理员。',
    icon: <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />,
    duration: 6,
    key: 'network-error', // 防止重复显示相同的网络错误
    placement: 'topRight',
  });
};

const handleAuthError = (error: Error) => {
  notification.warning({
    message: '身份验证失败',
    description: '您的登录状态已过期，请重新登录。',
    icon: <WarningOutlined style={{ color: '#fa8c16' }} />,
    duration: 8,
    key: 'auth-error',
    placement: 'topRight',
    btn: (
      <button
        onClick={() => {
          // 清除本地存储的认证信息
          localStorage.removeItem('auth-token');
          localStorage.removeItem('refresh-token');
          localStorage.removeItem('user');
          
          // 跳转到登录页面
          window.location.href = '/login';
          
          // 关闭通知
          notification.destroy('auth-error');
        }}
        style={{
          background: 'none',
          border: 'none',
          color: '#fa8c16',
          cursor: 'pointer',
          textDecoration: 'underline',
        }}
      >
        立即登录
      </button>
    ),
  });
};

const handleGenericError = (error: Error) => {
  const isDevelopment = process.env.NODE_ENV === 'development';
  
  // 过滤掉一些常见的非关键错误
  const ignoredErrors = [
    'ResizeObserver loop limit exceeded',
    'Non-Error promise rejection captured',
    'Script error',
    'Loading chunk',
    'Loading CSS chunk',
  ];
  
  const shouldIgnore = ignoredErrors.some(ignored => 
    error.message?.includes(ignored)
  );
  
  if (shouldIgnore) {
    return;
  }
  
  if (isDevelopment) {
    // 开发环境显示详细错误信息
    notification.error({
      message: '页面错误 (开发模式)',
      description: (
        <div>
          <div>{error.message || 'Unknown error'}</div>
          <div style={{ 
            marginTop: 8, 
            fontSize: 12, 
            color: '#666',
            fontFamily: 'Monaco, Menlo, monospace' 
          }}>
            {error.stack?.split('\n')[1]?.trim() || 'No stack trace'}
          </div>
        </div>
      ),
      icon: <CloseCircleOutlined style={{ color: '#ff4d4f' }} />,
      duration: 10,
      placement: 'topRight',
    });
  } else {
    // 生产环境显示友好的错误提示
    message.error({
      content: '页面遇到了一些问题，请刷新页面重试',
      duration: 5,
      key: 'generic-error',
    });
  }
};

export default GlobalErrorHandler;