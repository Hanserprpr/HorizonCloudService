import axios, { type AxiosInstance, AxiosError, type AxiosResponse } from 'axios';
import { API_BASE_URL, STORAGE_KEYS, HTTP_STATUS, ERROR_MESSAGES } from '@constants/index';
import type { ApiResponse } from '../types';

// 创建axios实例
class ApiClient {
  private instance: AxiosInstance;
  private isRefreshing = false;
  private failedQueue: Array<{
    resolve: (token: string) => void;
    reject: (error: any) => void;
  }> = [];

  constructor(baseURL: string = API_BASE_URL) {
    this.instance = axios.create({
      baseURL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.setupRequestInterceptor();
    this.setupResponseInterceptor();
  }

  // 请求拦截器
  private setupRequestInterceptor(): void {
    this.instance.interceptors.request.use(
      (config) => {
        // 添加JWT token
        const token = localStorage.getItem(STORAGE_KEYS.AUTH_TOKEN);
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }

        // 添加请求ID用于追踪
        config.headers['X-Request-ID'] = this.generateRequestId();

        // 开发环境日志
        if (import.meta.env.DEV) {
          console.log(`🚀 API Request: ${config.method?.toUpperCase()} ${config.url}`, {
            headers: config.headers,
            data: config.data,
            params: config.params,
          });
        }

        return config;
      },
      (error) => {
        console.error('Request interceptor error:', error);
        return Promise.reject(error);
      }
    );
  }

  // 响应拦截器
  private setupResponseInterceptor(): void {
    this.instance.interceptors.response.use(
      (response: AxiosResponse) => {
        // 开发环境日志
        if (import.meta.env.DEV) {
          console.log(`✅ API Response: ${response.config.method?.toUpperCase()} ${response.config.url}`, {
            status: response.status,
            data: response.data,
          });
        }

        return response;
      },
      async (error: AxiosError) => {
        const originalRequest = error.config as any;

        // 开发环境错误日志
        if (import.meta.env.DEV) {
          console.error(`❌ API Error: ${originalRequest?.method?.toUpperCase()} ${originalRequest?.url}`, {
            status: error.response?.status,
            data: error.response?.data,
            message: error.message,
          });
        }

        // 处理401错误 - token过期
        if (error.response?.status === HTTP_STATUS.UNAUTHORIZED && !originalRequest._retry) {
          if (this.isRefreshing) {
            // 如果正在刷新token，将请求加入队列
            return new Promise((resolve, reject) => {
              this.failedQueue.push({ resolve, reject });
            }).then(token => {
              originalRequest.headers.Authorization = `Bearer ${token}`;
              return this.instance(originalRequest);
            }).catch(err => {
              return Promise.reject(err);
            });
          }

          originalRequest._retry = true;
          this.isRefreshing = true;

          try {
            const refreshToken = localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN);
            if (!refreshToken) {
              throw new Error('No refresh token');
            }

            // 刷新token
            const newToken = await this.refreshToken(refreshToken);
            
            // 更新token
            localStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, newToken);
            
            // 处理队列中的请求
            this.processQueue(null, newToken);
            
            // 重新发起原请求
            originalRequest.headers.Authorization = `Bearer ${newToken}`;
            return this.instance(originalRequest);

          } catch (refreshError) {
            // 刷新token失败，清除认证状态
            this.processQueue(refreshError, null);
            this.clearAuthData();
            this.redirectToLogin();
            return Promise.reject(refreshError);
          } finally {
            this.isRefreshing = false;
          }
        }

        // 处理其他错误
        this.handleError(error);
        return Promise.reject(error);
      }
    );
  }

  // 刷新token
  private async refreshToken(refreshToken: string): Promise<string> {
    const response = await axios.post(
      `${API_BASE_URL}/api/v1/auth/refresh`,
      { refresh_token: refreshToken },
      { timeout: 10000 }
    );
    return response.data.data.access_token;
  }

  // 处理请求队列
  private processQueue(error: any, token: string | null): void {
    this.failedQueue.forEach(({ resolve, reject }) => {
      if (error) {
        reject(error);
      } else {
        resolve(token!);
      }
    });
    
    this.failedQueue = [];
  }

  // 清除认证数据
  private clearAuthData(): void {
    localStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN);
    localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN);
    localStorage.removeItem(STORAGE_KEYS.USER_INFO);
  }

  // 重定向到登录页
  private redirectToLogin(): void {
    if (typeof window !== 'undefined') {
      window.location.href = '/auth/login';
    }
  }

  // 生成请求ID
  private generateRequestId(): string {
    return `req-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  // 错误处理 - 现在只记录错误，不显示消息（由GlobalErrorHandler处理）
  private handleError(error: AxiosError): void {
    // 增强错误对象，添加更多上下文信息
    const enhancedError = {
      ...error,
      timestamp: new Date().toISOString(),
      requestId: error.config?.headers?.['X-Request-ID'],
      url: error.config?.url,
      method: error.config?.method?.toUpperCase(),
      isNetworkError: !navigator.onLine || error.code === 'ECONNABORTED',
      status: error.response?.status,
      data: error.response?.data,
    };

    // 记录错误信息（用于调试和监控）
    console.error('API Error Details:', enhancedError);

    // 可以在这里添加错误监控服务的调用
    // 例如: Sentry.captureException(enhancedError);
  }

  // 公共方法
  public get<T = any>(url: string, params?: any): Promise<ApiResponse<T>> {
    return this.instance.get(url, { params }).then(response => response.data);
  }

  public post<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
    return this.instance.post(url, data).then(response => response.data);
  }

  public put<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
    return this.instance.put(url, data).then(response => response.data);
  }

  public patch<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
    return this.instance.patch(url, data).then(response => response.data);
  }

  public delete<T = any>(url: string): Promise<ApiResponse<T>> {
    return this.instance.delete(url).then(response => response.data);
  }

  // 文件上传
  public upload<T = any>(
    url: string, 
    formData: FormData, 
    onProgress?: (progress: number) => void,
    options?: {
      signal?: AbortSignal;
      onUploadProgress?: (progressEvent: any) => void;
    }
  ): Promise<ApiResponse<T>> {
    return this.instance.post(url, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      signal: options?.signal,
      onUploadProgress: (progressEvent) => {
        // 使用传入的自定义进度回调，否则使用默认的
        if (options?.onUploadProgress) {
          options.onUploadProgress(progressEvent);
        } else if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          onProgress(progress);
        }
      },
    }).then(response => response.data);
  }

  // 下载文件
  public download(url: string, filename?: string): Promise<void> {
    return this.instance.get(url, {
      responseType: 'blob',
    }).then(response => {
      const blob = new Blob([response.data]);
      const downloadUrl = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = downloadUrl;
      link.download = filename || 'download';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(downloadUrl);
    });
  }

  // 获取原始axios实例（用于特殊需求）
  public getInstance(): AxiosInstance {
    return this.instance;
  }
}

// 创建默认API客户端实例
export const apiClient = new ApiClient();

// 导出类用于创建其他实例
export { ApiClient };
export default apiClient;