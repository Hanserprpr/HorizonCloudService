import { createBrowserRouter, Navigate } from 'react-router-dom';
import { lazy, Suspense } from 'react';
import { Spin } from 'antd';
import { ROUTES } from '@constants/index';
import AuthGuard from './AuthGuard';
import MainLayout from '@components/layout/MainLayout';
import ErrorPage from '@pages/error/ErrorPage';

// 页面懒加载
const LoginPage = lazy(() => import('@pages/auth/LoginPage'));
const DashboardPage = lazy(() => import('@pages/dashboard/DashboardPage'));
const FilesPage = lazy(() => import('@pages/files/FilesPage'));
const UsersPage = lazy(() => import('@pages/users/UsersPage'));
const SettingsPage = lazy(() => import('@pages/settings/SettingsPage'));
const ProfilePage = lazy(() => import('@pages/profile/ProfilePage'));

// 加载中组件
const PageLoadingSpinner = () => (
  <div className="loading-container">
    <Spin size="large" tip="页面加载中..." />
  </div>
);

// 懒加载包装器
const LazyWrapper = ({ children }: { children: React.ReactNode }) => (
  <Suspense fallback={<PageLoadingSpinner />}>
    {children}
  </Suspense>
);

// 路由配置
export const router = createBrowserRouter([
  {
    path: '/',
    element: <Navigate to={ROUTES.DASHBOARD} replace />,
    errorElement: <ErrorPage />,
  },
  {
    path: '/auth',
    children: [
      {
        path: 'login',
        element: (
          <LazyWrapper>
            <LoginPage />
          </LazyWrapper>
        ),
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
      <AuthGuard>
        <MainLayout />
      </AuthGuard>
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