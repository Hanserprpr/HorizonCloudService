import { createBrowserRouter, Navigate } from 'react-router-dom';
import SimpleLoginPage from '@pages/auth/SimpleLoginPage';

// 简单的测试页面
const TestLoginPage = () => (
  <div style={{ 
    minHeight: '100vh', 
    display: 'flex', 
    alignItems: 'center', 
    justifyContent: 'center',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    color: 'white',
    fontSize: '24px',
    fontFamily: 'Arial, sans-serif'
  }}>
    <div style={{ textAlign: 'center' }}>
      <h1>登录页面</h1>
      <p>这是一个简化的测试页面</p>
      <button 
        onClick={() => alert('点击成功!')}
        style={{
          padding: '12px 24px',
          fontSize: '16px',
          background: '#fff',
          color: '#333',
          border: 'none',
          borderRadius: '6px',
          cursor: 'pointer'
        }}
      >
        测试按钮
      </button>
    </div>
  </div>
);

const SimpleDashboardPage = () => (
  <div style={{ 
    minHeight: '100vh', 
    display: 'flex', 
    alignItems: 'center', 
    justifyContent: 'center',
    background: '#f0f2f5',
    fontSize: '24px',
    fontFamily: 'Arial, sans-serif'
  }}>
    <div style={{ textAlign: 'center' }}>
      <h1>仪表盘页面</h1>
      <p>这是一个简化的测试页面</p>
    </div>
  </div>
);

// 简化路由配置
export const simpleRouter = createBrowserRouter([
  {
    path: '/',
    element: <Navigate to="/auth/login" replace />,
  },
  {
    path: '/auth/login',
    element: <SimpleLoginPage />,
  },
  {
    path: '/dashboard',
    element: <SimpleDashboardPage />,
  },
  {
    path: '*',
    element: <div style={{ padding: '20px' }}>404 - 页面不存在</div>,
  },
]);