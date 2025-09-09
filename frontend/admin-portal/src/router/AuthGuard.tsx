import React, { useEffect } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { Spin } from 'antd';
import { useAuth } from '@hooks/useAuth';
import { ROUTES } from '@constants/index';

interface AuthGuardProps {
  children: React.ReactNode;
}

const AuthGuard: React.FC<AuthGuardProps> = ({ children }) => {
  const { isAuthenticated, loading, checkAuth, user } = useAuth();
  const location = useLocation();

  useEffect(() => {
    // 组件挂载时检查认证状态
    checkAuth();
  }, [checkAuth]);

  // 显示加载状态
  if (loading) {
    return (
      <div
        style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          minHeight: '100vh',
          flexDirection: 'column',
        }}
      >
        <Spin size="large" tip="验证身份中..." />
      </div>
    );
  }

  // 未认证用户重定向到登录页
  if (!isAuthenticated || !user) {
    return (
      <Navigate
        to={ROUTES.AUTH.LOGIN}
        state={{ from: location.pathname }}
        replace
      />
    );
  }

  // 已认证用户可以访问受保护的路由
  return <>{children}</>;
};

export default AuthGuard;