import React from 'react';
import { Spin, Alert } from 'antd';

interface LoadingWrapperProps {
  loading?: boolean;
  error?: Error | null;
  empty?: boolean;
  emptyText?: string;
  children: React.ReactNode;
  size?: 'small' | 'default' | 'large';
  tip?: string;
  minHeight?: number;
}

const LoadingWrapper: React.FC<LoadingWrapperProps> = ({
  loading = false,
  error = null,
  empty = false,
  emptyText = '暂无数据',
  children,
  size = 'large',
  tip = '加载中...',
  minHeight = 200,
}) => {
  // 错误状态
  if (error) {
    return (
      <div style={{ minHeight, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Alert
          message="加载失败"
          description={error.message || '数据加载失败，请稍后重试'}
          type="error"
          showIcon
          style={{ maxWidth: 400 }}
        />
      </div>
    );
  }

  // 加载状态
  if (loading) {
    return (
      <div style={{ minHeight, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Spin size={size} tip={tip} />
      </div>
    );
  }

  // 空状态
  if (empty) {
    return (
      <div style={{ minHeight, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <div style={{ textAlign: 'center', color: '#999' }}>
          <div style={{ fontSize: '48px', marginBottom: '16px' }}>📄</div>
          <div>{emptyText}</div>
        </div>
      </div>
    );
  }

  // 正常渲染
  return <>{children}</>;
};

export default LoadingWrapper;