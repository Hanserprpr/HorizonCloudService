import { QueryClient } from '@tanstack/react-query';
import { message } from 'antd';
import { HTTP_STATUS, ERROR_MESSAGES } from '@constants/index';

// 创建查询客户端
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // 缓存时间：5分钟
      staleTime: 5 * 60 * 1000,
      // 缓存数据保留时间：10分钟
      gcTime: 10 * 60 * 1000,
      // 重试配置
      retry: (failureCount, error: any) => {
        // 认证错误不重试
        if (error?.response?.status === HTTP_STATUS.UNAUTHORIZED) {
          return false;
        }
        // 最多重试2次
        return failureCount < 2;
      },
      // 重试延迟
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      // 网络重连时重新获取数据
      refetchOnReconnect: true,
      // 窗口重新获得焦点时不自动重新获取数据
      refetchOnWindowFocus: false,
    },
    mutations: {
      // 错误处理
      onError: (error: any) => {
        console.error('Mutation error:', error);
        
        let errorMessage: string = ERROR_MESSAGES.SERVER_ERROR;
        
        if (error?.response?.status) {
          switch (error.response.status) {
            case HTTP_STATUS.UNAUTHORIZED:
              errorMessage = ERROR_MESSAGES.UNAUTHORIZED;
              break;
            case HTTP_STATUS.FORBIDDEN:
              errorMessage = ERROR_MESSAGES.FORBIDDEN;
              break;
            case HTTP_STATUS.NOT_FOUND:
              errorMessage = ERROR_MESSAGES.NOT_FOUND;
              break;
            case HTTP_STATUS.UNPROCESSABLE_ENTITY:
              errorMessage = ERROR_MESSAGES.FORM_VALIDATION_ERROR;
              break;
            default:
              errorMessage = error?.response?.data?.message || ERROR_MESSAGES.SERVER_ERROR;
          }
        } else if (!navigator.onLine) {
          errorMessage = ERROR_MESSAGES.NETWORK_ERROR;
        }
        
        message.error(errorMessage);
      },
    },
  },
});

// 查询键工厂
export const queryKeys = {
  // 用户相关
  users: {
    all: ['users'] as const,
    lists: () => [...queryKeys.users.all, 'list'] as const,
    list: (params: Record<string, any>) => [...queryKeys.users.lists(), params] as const,
    details: () => [...queryKeys.users.all, 'detail'] as const,
    detail: (id: number) => [...queryKeys.users.details(), id] as const,
    profile: () => [...queryKeys.users.all, 'profile'] as const,
    stats: () => [...queryKeys.users.all, 'stats'] as const,
  },

  // 文件相关
  files: {
    all: ['files'] as const,
    lists: () => [...queryKeys.files.all, 'list'] as const,
    list: (params: Record<string, any>) => [...queryKeys.files.lists(), params] as const,
    details: () => [...queryKeys.files.all, 'detail'] as const,
    detail: (id: number) => [...queryKeys.files.details(), id] as const,
    folderContents: (folderId?: number) => [...queryKeys.files.all, 'folder', folderId] as const,
    search: (params: Record<string, any>) => [...queryKeys.files.all, 'search', params] as const,
    stats: () => [...queryKeys.files.all, 'stats'] as const,
  },

  // 文件夹相关
  folders: {
    all: ['folders'] as const,
    lists: () => [...queryKeys.folders.all, 'list'] as const,
    list: (params: Record<string, any>) => [...queryKeys.folders.lists(), params] as const,
    details: () => [...queryKeys.folders.all, 'detail'] as const,
    detail: (id: number) => [...queryKeys.folders.details(), id] as const,
    tree: () => [...queryKeys.folders.all, 'tree'] as const,
    breadcrumb: (path: string) => [...queryKeys.folders.all, 'breadcrumb', path] as const,
  },

  // 上传相关
  uploads: {
    all: ['uploads'] as const,
    sessions: () => [...queryKeys.uploads.all, 'sessions'] as const,
    session: (sessionId: string) => [...queryKeys.uploads.sessions(), sessionId] as const,
    progress: (sessionId: string) => [...queryKeys.uploads.all, 'progress', sessionId] as const,
  },

  // 缩略图相关
  thumbnails: {
    all: ['thumbnails'] as const,
    lists: () => [...queryKeys.thumbnails.all, 'list'] as const,
    list: (fileId: number) => [...queryKeys.thumbnails.lists(), fileId] as const,
    detail: (id: number) => [...queryKeys.thumbnails.all, 'detail', id] as const,
  },

  // 系统相关
  system: {
    all: ['system'] as const,
    stats: () => [...queryKeys.system.all, 'stats'] as const,
    health: () => [...queryKeys.system.all, 'health'] as const,
    config: () => [...queryKeys.system.all, 'config'] as const,
  },
} as const;

// 类型推断辅助
export type QueryKeys = typeof queryKeys;