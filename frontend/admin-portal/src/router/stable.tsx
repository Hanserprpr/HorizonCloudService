import { createBrowserRouter, Navigate } from 'react-router-dom';
import { lazy, Suspense } from 'react';
import { Spin } from 'antd';
import { ROUTES } from '@constants/index';
import SimpleAuthGuard from './SimpleAuthGuard';
import StableMainLayout from '@components/layout/StableMainLayout';  // 使用稳定MainLayout
import SimpleLoginPage from '@pages/auth/SimpleLoginPage'; // 使用简化登录

// 现有的页面组件
const DashboardPage = lazy(() => import('@pages/dashboard/DashboardPage'));
const FilesPage = lazy(() => import('@pages/files/FilesPage'));
const UsersPage = lazy(() => import('@pages/users/UsersPage'));
const SettingsPage = lazy(() => import('@pages/settings/SettingsPage'));
const ProfilePage = lazy(() => import('@pages/profile/ProfilePage'));
const ErrorPage = lazy(() => import('@pages/error/ErrorPage'));

// 加载中组件
const PageLoadingSpinner = () => (
  <div style={{
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    minHeight: '200px',
    flexDirection: 'column',
  }}>
    <Spin size="large" tip="页面加载中..." />
  </div>
);

// 懒加载包装器
const LazyWrapper = ({ children }: { children: React.ReactNode }) => (
  <Suspense fallback={<PageLoadingSpinner />}>
    {children}
  </Suspense>
);

// 稳定版路由配置 - 混合简化和完整组件
export const stableRouter = createBrowserRouter([
  {
    path: '/',
    element: <Navigate to={ROUTES.DASHBOARD} replace />,
    errorElement: <LazyWrapper><ErrorPage /></LazyWrapper>,
  },
  {
    path: '/auth',
    children: [
      {
        path: 'login',
        element: <SimpleLoginPage />, // 使用简化登录页面
      },
      {
        index: true,
        element: <Navigate to="/auth/login" replace />,
      },
    ],
  },
  {
    path: '/',
    element: (
      <SimpleAuthGuard> {/* 使用简化认证守卫 */}
        <StableMainLayout />   {/* 使用稳定的MainLayout */}
      </SimpleAuthGuard>
    ),
    errorElement: <LazyWrapper><ErrorPage /></LazyWrapper>,
    children: [
      {
        path: 'dashboard',
        element: (
          <LazyWrapper>
            <DashboardPage />
          </LazyWrapper>
        ),
      },
      {
        path: 'files',
        children: [
          {
            index: true,
            element: (
              <LazyWrapper>
                <FilesPage />
              </LazyWrapper>
            ),
          },
          {
            path: ':folderId',
            element: (
              <LazyWrapper>
                <FilesPage />
              </LazyWrapper>
            ),
          },
        ],
      },
      {
        path: 'users',
        children: [
          {
            index: true,
            element: (
              <LazyWrapper>
                <UsersPage />
              </LazyWrapper>
            ),
          },
          {
            path: ':userId',
            element: (
              <LazyWrapper>
                <UsersPage />
              </LazyWrapper>
            ),
          },
        ],
      },
      {
        path: 'settings',
        element: (
          <LazyWrapper>
            <SettingsPage />
          </LazyWrapper>
        ),
      },
      {
        path: 'profile',
        element: (
          <LazyWrapper>
            <ProfilePage />
          </LazyWrapper>
        ),
      },
    ],
  },
  {
    path: '*',
    element: <LazyWrapper><ErrorPage /></LazyWrapper>,
  },
]);