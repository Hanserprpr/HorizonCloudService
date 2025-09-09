import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { STORAGE_KEYS } from '@constants/index';

interface UIState {
  // 侧边栏状态
  sidebarCollapsed: boolean;
  setSidebarCollapsed: (collapsed: boolean) => void;
  toggleSidebar: () => void;
  
  // 主题设置
  theme: 'light' | 'dark';
  setTheme: (theme: 'light' | 'dark') => void;
  
  // 全局加载状态
  loading: boolean;
  setLoading: (loading: boolean) => void;
  
  // 当前路径面包屑
  breadcrumb: Array<{ title: string; path?: string }>;
  setBreadcrumb: (breadcrumb: Array<{ title: string; path?: string }>) => void;
  
  // 移动端响应式
  isMobile: boolean;
  setIsMobile: (isMobile: boolean) => void;
  
  // 语言设置
  language: string;
  setLanguage: (language: string) => void;
  
  // 通知数量
  notificationCount: number;
  setNotificationCount: (count: number) => void;
  
  // 刷新触发器
  refreshTrigger: number;
  triggerRefresh: () => void;
}

export const useUIStore = create<UIState>()(
  persist(
    (set, get) => ({
      // 初始状态
      sidebarCollapsed: false,
      theme: 'light',
      loading: false,
      breadcrumb: [],
      isMobile: false,
      language: 'zh-CN',
      notificationCount: 0,
      refreshTrigger: 0,

      // 侧边栏控制
      setSidebarCollapsed: (collapsed: boolean) => {
        set({ sidebarCollapsed: collapsed });
      },

      toggleSidebar: () => {
        set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed }));
      },

      // 主题控制
      setTheme: (theme: 'light' | 'dark') => {
        set({ theme });
      },

      // 加载状态控制
      setLoading: (loading: boolean) => {
        set({ loading });
      },

      // 面包屑控制
      setBreadcrumb: (breadcrumb: Array<{ title: string; path?: string }>) => {
        set({ breadcrumb });
      },

      // 移动端控制
      setIsMobile: (isMobile: boolean) => {
        set({ isMobile });
        // 移动端自动收起侧边栏
        if (isMobile) {
          set({ sidebarCollapsed: true });
        }
      },

      // 语言设置
      setLanguage: (language: string) => {
        set({ language });
      },

      // 通知数量
      setNotificationCount: (count: number) => {
        set({ notificationCount: count });
      },

      // 刷新触发
      triggerRefresh: () => {
        set((state) => ({ refreshTrigger: state.refreshTrigger + 1 }));
      },
    }),
    {
      name: 'ui-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        sidebarCollapsed: state.sidebarCollapsed,
        theme: state.theme,
        language: state.language,
      }),
    }
  )
);