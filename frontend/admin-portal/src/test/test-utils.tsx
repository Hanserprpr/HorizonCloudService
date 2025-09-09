import React from 'react';
import { render, RenderOptions } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';

// 创建测试用的QueryClient
const createTestQueryClient = () => new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
    mutations: {
      retry: false,
    },
  },
});

// 包装组件，提供必要的Provider
const AllTheProviders: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const queryClient = createTestQueryClient();

  return (
    <BrowserRouter>
      <QueryClientProvider client={queryClient}>
        <ConfigProvider locale={zhCN}>
          {children}
        </ConfigProvider>
      </QueryClientProvider>
    </BrowserRouter>
  );
};

// 自定义render函数
const customRender = (
  ui: React.ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) => render(ui, { wrapper: AllTheProviders, ...options });

// 重新导出所有testing-library内容
export * from '@testing-library/react';
export { customRender as render };

// 测试工具函数
export const mockLocalStorage = () => {
  const store: { [key: string]: string } = {};
  
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value.toString();
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      Object.keys(store).forEach(key => delete store[key]);
    },
  };
};

// Mock用户数据
export const mockUser = {
  id: 1,
  username: 'admin',
  email: 'admin@example.com',
  display_name: '管理员',
  role: 'admin',
  status: 1,
  storage_quota: 10737418240, // 10GB
  storage_used: 1073741824,   // 1GB
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

// Mock文件数据
export const mockFile = {
  id: 1,
  name: 'test-image.jpg',
  size: 1048576, // 1MB
  type: 'file',
  content_type: 'image/jpeg',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  folder_id: null,
  user_id: 1,
};