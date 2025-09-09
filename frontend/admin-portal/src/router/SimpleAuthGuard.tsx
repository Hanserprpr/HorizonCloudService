import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';

interface SimpleAuthGuardProps {
  children: React.ReactNode;
}

const SimpleAuthGuard: React.FC<SimpleAuthGuardProps> = ({ children }) => {
  const location = useLocation();
  
  // 简化的认证检查 - 检查localStorage中是否有模拟的认证状态
  const isAuthenticated = localStorage.getItem('simple-auth') === 'true';

  // 未认证用户重定向到登录页
  if (!isAuthenticated) {
    return (
      <Navigate
        to="/auth/login"
        state={{ from: location.pathname }}
        replace
      />
    );
  }

  // 已认证用户可以访问受保护的路由
  return <>{children}</>;
};

export default SimpleAuthGuard;