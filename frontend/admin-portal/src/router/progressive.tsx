import { createBrowserRouter, Navigate } from 'react-router-dom';
import { lazy, Suspense } from 'react';
import { Spin } from 'antd';
import SimpleAuthGuard from './SimpleAuthGuard';
import SimpleMainLayout from '@components/layout/SimpleMainLayout';
import SimpleLoginPage from '@pages/auth/SimpleLoginPage';

// 现有的简单页面组件
const DashboardPage = lazy(() => import('@pages/dashboard/DashboardPage'));
const FilesPage = lazy(() => import('@pages/files/FilesPage'));
const UsersPage = lazy(() => import('@pages/users/UsersPage'));
const SettingsPage = lazy(() => import('@pages/settings/SettingsPage'));
const ProfilePage = lazy(() => import('@pages/profile/ProfilePage'));

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

// 错误页面
const ErrorPage = () => (
  <div style={{
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    minHeight: '100vh',
    flexDirection: 'column',
  }}>
    <h1>出错了</h1>
    <p>页面加载失败，请刷新重试。</p>
    <button onClick={() => window.location.reload()}>刷新页面</button>
  </div>
);

// 渐进式路由配置
export const progressiveRouter = createBrowserRouter([
  {
    path: '/',
    element: <Navigate to="/dashboard" replace />,
    errorElement: <ErrorPage />,
  },
  {
    path: '/auth',
    children: [
      {
        path: 'login',
        element: <SimpleLoginPage />,
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
      <SimpleAuthGuard>
        <SimpleMainLayout />
      </SimpleAuthGuard>
    ),
    errorElement: <ErrorPage />,
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
    element: <ErrorPage />,
  },
]);