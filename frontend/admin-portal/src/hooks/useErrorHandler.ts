import { useCallback, createElement } from 'react';
import { message, notification, Modal } from 'antd';
import { 
  ExclamationCircleOutlined, 
  InfoCircleOutlined, 
  CheckCircleOutlined,
  CloseCircleOutlined,
  WarningOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { ROUTES } from '@constants/index';

interface ApiError {
  message: string;
  code?: string | number;
  status?: number;
  data?: any;
  timestamp?: string;
}

interface ErrorHandlerOptions {
  showNotification?: boolean;
  showMessage?: boolean;
  showModal?: boolean;
  autoRedirect?: boolean;
  notificationDuration?: number;
  messageDuration?: number;
}

interface FeedbackOptions {
  type?: 'success' | 'info' | 'warning' | 'error';
  title?: string;
  message: string;
  duration?: number;
  showNotification?: boolean;
  showMessage?: boolean;
  onClose?: () => void;
  onAction?: () => void;
  actionText?: string;
}

/**
 * 统一的错误处理和用户反馈 Hook
 * 提供错误处理、用户反馈、成功提示等功能
 */
export const useErrorHandler = () => {
  const navigate = useNavigate();

  // 错误处理主函数
  const handleError = useCallback((
    error: Error | ApiError | any,
    options: ErrorHandlerOptions = {}
  ) => {
    const {
      showNotification = true,
      showMessage = false,
      showModal = false,
      autoRedirect = true,
      notificationDuration = 6,
      messageDuration = 4,
    } = options;

    // 解析错误信息
    const errorInfo = parseError(error);
    
    console.error('Error handled:', {
      originalError: error,
      parsedError: errorInfo,
      options,
      timestamp: new Date().toISOString(),
      url: window.location.href,
    });

    // 根据错误类型执行不同的处理策略
    switch (errorInfo.type) {
      case 'network':
        handleNetworkError(errorInfo, { showNotification, notificationDuration });
        break;
        
      case 'auth':
        handleAuthError(errorInfo, { showNotification, autoRedirect, notificationDuration });
        break;
        
      case 'validation':
        handleValidationError(errorInfo, { showMessage, messageDuration });
        break;
        
      case 'permission':
        handlePermissionError(errorInfo, { showNotification, showModal, notificationDuration });
        break;
        
      case 'server':
        handleServerError(errorInfo, { showNotification, notificationDuration });
        break;
        
      default:
        handleGenericError(errorInfo, { 
          showNotification, 
          showMessage, 
          notificationDuration, 
          messageDuration 
        });
        break;
    }

    return errorInfo;
  }, [navigate]);

  // 用户反馈函数
  const showFeedback = useCallback((options: FeedbackOptions) => {
    const {
      type = 'info',
      title,
      message: msg,
      duration = 4,
      showNotification = true,
      showMessage = false,
      onClose,
      onAction,
      actionText,
    } = options;

    if (showNotification && title) {
      const iconMap = {
        success: createElement(CheckCircleOutlined, { style: { color: '#52c41a' } }),
        info: createElement(InfoCircleOutlined, { style: { color: '#1677ff' } }),
        warning: createElement(WarningOutlined, { style: { color: '#fa8c16' } }),
        error: createElement(CloseCircleOutlined, { style: { color: '#ff4d4f' } }),
      };

      notification[type]({
        message: title,
        description: msg,
        icon: iconMap[type],
        duration,
        placement: 'topRight',
        onClose,
        btn: onAction && actionText ? createElement('button', {
          onClick: () => {
            onAction();
            notification.destroy();
          },
          style: {
            background: 'none',
            border: 'none',
            color: type === 'error' ? '#ff4d4f' : '#1677ff',
            cursor: 'pointer',
            textDecoration: 'underline',
          },
        }, actionText) : undefined,
      });
    } else if (showMessage || !title) {
      message[type]({
        content: msg,
        duration,
        onClose,
      });
    }
  }, []);

  // 成功反馈
  const showSuccess = useCallback((msg: string, title?: string) => {
    showFeedback({ type: 'success', message: msg, title });
  }, [showFeedback]);

  // 信息反馈
  const showInfo = useCallback((msg: string, title?: string) => {
    showFeedback({ type: 'info', message: msg, title });
  }, [showFeedback]);

  // 警告反馈
  const showWarning = useCallback((msg: string, title?: string) => {
    showFeedback({ type: 'warning', message: msg, title });
  }, [showFeedback]);

  // 错误反馈
  const showError = useCallback((msg: string, title?: string) => {
    showFeedback({ type: 'error', message: msg, title });
  }, [showFeedback]);

  // 确认对话框
  const showConfirm = useCallback((
    title: string,
    content: string,
    onOk: () => void | Promise<void>,
    onCancel?: () => void
  ) => {
    Modal.confirm({
      title,
      content,
      icon: createElement(ExclamationCircleOutlined),
      okText: '确定',
      cancelText: '取消',
      onOk,
      onCancel,
      centered: true,
    });
  }, []);

  // 危险操作确认对话框
  const showDeleteConfirm = useCallback((
    title: string,
    content: string,
    onOk: () => void | Promise<void>,
    onCancel?: () => void
  ) => {
    Modal.confirm({
      title,
      content,
      icon: createElement(ExclamationCircleOutlined, { style: { color: '#ff4d4f' } }),
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk,
      onCancel,
      centered: true,
    });
  }, []);

  return {
    handleError,
    showFeedback,
    showSuccess,
    showInfo,
    showWarning,
    showError,
    showConfirm,
    showDeleteConfirm,
  };
};

// 错误解析函数
const parseError = (error: any): {
  type: 'network' | 'auth' | 'validation' | 'permission' | 'server' | 'generic';
  message: string;
  code?: string | number;
  status?: number;
  details?: any;
} => {
  // 网络错误
  if (
    error.name === 'NetworkError' ||
    error.message?.includes('Network Error') ||
    error.message?.includes('fetch') ||
    error.code === 'ERR_NETWORK'
  ) {
    return {
      type: 'network',
      message: '网络连接失败，请检查网络设置后重试',
      code: error.code,
    };
  }

  // HTTP 错误响应
  if (error.response || error.status) {
    const status = error.response?.status || error.status;
    const data = error.response?.data || error.data;
    const message = data?.message || error.message || '服务器错误';

    switch (status) {
      case 400:
        return {
          type: 'validation',
          message: message || '请求参数错误',
          status,
          details: data,
        };
        
      case 401:
        return {
          type: 'auth',
          message: '身份验证失败，请重新登录',
          status,
          details: data,
        };
        
      case 403:
        return {
          type: 'permission',
          message: '没有权限执行此操作',
          status,
          details: data,
        };
        
      case 404:
        return {
          type: 'generic',
          message: '请求的资源不存在',
          status,
          details: data,
        };
        
      case 429:
        return {
          type: 'generic',
          message: '请求太频繁，请稍后重试',
          status,
          details: data,
        };
        
      case 500:
      case 502:
      case 503:
      case 504:
        return {
          type: 'server',
          message: '服务器暂时不可用，请稍后重试',
          status,
          details: data,
        };
        
      default:
        return {
          type: 'generic',
          message: message || `服务器错误 (${status})`,
          status,
          details: data,
        };
    }
  }

  // 一般错误
  return {
    type: 'generic',
    message: error.message || error.toString() || '发生未知错误',
    details: error,
  };
};

// 特定类型错误处理函数
const handleNetworkError = (errorInfo: any, options: any) => {
  if (options.showNotification) {
    notification.error({
      message: '网络连接错误',
      description: errorInfo.message,
      duration: options.notificationDuration,
      placement: 'topRight',
    });
  }
};

const handleAuthError = (errorInfo: any, options: any) => {
  if (options.showNotification) {
    notification.warning({
      message: '身份验证失败',
      description: errorInfo.message,
      duration: options.notificationDuration,
      placement: 'topRight',
    });
  }

  if (options.autoRedirect) {
    setTimeout(() => {
      localStorage.clear();
      window.location.href = ROUTES.LOGIN;
    }, 2000);
  }
};

const handleValidationError = (errorInfo: any, options: any) => {
  if (options.showMessage) {
    message.error({
      content: errorInfo.message,
      duration: options.messageDuration,
    });
  }
};

const handlePermissionError = (errorInfo: any, options: any) => {
  if (options.showNotification) {
    notification.warning({
      message: '权限不足',
      description: errorInfo.message,
      duration: options.notificationDuration,
      placement: 'topRight',
    });
  }
};

const handleServerError = (errorInfo: any, options: any) => {
  if (options.showNotification) {
    notification.error({
      message: '服务器错误',
      description: errorInfo.message,
      duration: options.notificationDuration,
      placement: 'topRight',
    });
  }
};

const handleGenericError = (errorInfo: any, options: any) => {
  if (options.showNotification) {
    notification.error({
      message: '操作失败',
      description: errorInfo.message,
      duration: options.notificationDuration,
      placement: 'topRight',
    });
  } else if (options.showMessage) {
    message.error({
      content: errorInfo.message,
      duration: options.messageDuration,
    });
  }
};

export default useErrorHandler;