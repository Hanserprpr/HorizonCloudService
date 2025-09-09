import React from 'react';
import { Result, Button } from 'antd';
import { useNavigate, useRouteError } from 'react-router-dom';
import { ROUTES } from '@constants/index';

const ErrorPage: React.FC = () => {
  const navigate = useNavigate();
  const error = useRouteError() as any;

  const getErrorInfo = () => {
    if (error?.status === 404) {
      return {
        status: '404' as const,
        title: '页面不存在',
        subTitle: '抱歉，您访问的页面不存在。',
      };
    }

    if (error?.status === 403) {
      return {
        status: '403' as const,
        title: '访问被拒绝',
        subTitle: '抱歉，您没有权限访问此页面。',
      };
    }

    return {
      status: '500' as const,
      title: '服务器错误',
      subTitle: '抱歉，服务器出现了一些问题。',
    };
  };

  const { status, title, subTitle } = getErrorInfo();

  const handleGoHome = () => {
    navigate(ROUTES.DASHBOARD, { replace: true });
  };

  const handleGoBack = () => {
    navigate(-1);
  };

  return (
    <div className="error-page">
      <Result
        status={status}
        title={title}
        subTitle={subTitle}
        extra={[
          <Button type="primary" key="home" onClick={handleGoHome}>
            返回首页
          </Button>,
          <Button key="back" onClick={handleGoBack}>
            返回上一页
          </Button>,
        ]}
      />
      
      {/* 开发环境显示错误详情 */}
      {import.meta.env.DEV && error && (
        <div style={{ margin: '20px', padding: '20px', backgroundColor: '#f5f5f5', borderRadius: '4px' }}>
          <h4>错误详情 (仅开发环境显示):</h4>
          <pre style={{ fontSize: '12px', overflow: 'auto' }}>
            {error.message || JSON.stringify(error, null, 2)}
          </pre>
        </div>
      )}

      <style jsx>{`
        .error-page {
          min-height: 100vh;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 20px;
        }
      `}</style>
    </div>
  );
};

export default ErrorPage;