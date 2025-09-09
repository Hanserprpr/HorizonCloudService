import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { STORAGE_KEYS } from '@constants/index';
import type { User, LoginResponse } from '@types/index';

interface AuthState {
  // 状态
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  
  // 动作
  login: (loginData: LoginResponse) => void;
  logout: () => void;
  updateUser: (user: User) => void;
  setLoading: (loading: boolean) => void;
  setToken: (token: string) => void;
  
  // 计算属性
  getUser: () => User | null;
  isAdmin: () => boolean;
  hasPermission: (permission: string) => boolean;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      // 初始状态
      user: null,
      token: null,
      refreshToken: null,
      isAuthenticated: false,
      loading: false,

      // 登录
      login: (loginData: LoginResponse) => {
        const { user, access_token, refresh_token } = loginData;
        
        set({
          user,
          token: access_token,
          refreshToken: refresh_token,
          isAuthenticated: true,
          loading: false,
        });
        
        // 存储到localStorage
        localStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, access_token);
        localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, refresh_token);
        localStorage.setItem(STORAGE_KEYS.USER_INFO, JSON.stringify(user));
      },

      // 退出登录
      logout: () => {
        set({
          user: null,
          token: null,
          refreshToken: null,
          isAuthenticated: false,
          loading: false,
        });
        
        // 清除localStorage
        localStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN);
        localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN);
        localStorage.removeItem(STORAGE_KEYS.USER_INFO);
      },

      // 更新用户信息
      updateUser: (user: User) => {
        set({ user });
        localStorage.setItem(STORAGE_KEYS.USER_INFO, JSON.stringify(user));
      },

      // 设置加载状态
      setLoading: (loading: boolean) => {
        set({ loading });
      },

      // 设置token
      setToken: (token: string) => {
        set({ token });
        localStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, token);
      },

      // 获取用户信息
      getUser: () => {
        return get().user;
      },

      // 检查是否为管理员
      isAdmin: () => {
        const user = get().user;
        return user?.role === 'admin';
      },

      // 检查权限
      hasPermission: (permission: string) => {
        const user = get().user;
        if (!user) return false;
        
        // 管理员拥有所有权限
        if (user.role === 'admin') return true;
        
        // 根据角色检查权限（可扩展）
        switch (permission) {
          case 'file.read':
          case 'file.upload':
            return true;
          case 'file.delete':
          case 'file.manage':
            return user.role === 'admin';
          case 'user.manage':
          case 'system.config':
            return user.role === 'admin';
          default:
            return false;
        }
      },
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);