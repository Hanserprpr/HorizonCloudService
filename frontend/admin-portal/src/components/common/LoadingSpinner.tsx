import React from 'react';
import { Spin, Card, Skeleton, Space } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';

interface LoadingSpinnerProps {
  size?: 'small' | 'default' | 'large';
  tip?: string;
  spinning?: boolean;
  children?: React.ReactNode;
  style?: React.CSSProperties;
  className?: string;
  type?: 'spinner' | 'skeleton' | 'card' | 'overlay';
  rows?: number; // 用于 skeleton 类型
  height?: string | number; // 用于占位高度
}

/**
 * 统一的加载状态组件
 * 支持多种加载状态显示方式：旋转器、骨架屏、卡片式等
 */
export const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  size = 'default',
  tip = '加载中...',
  spinning = true,
  children,
  style,
  className,
  type = 'spinner',
  rows = 3,
  height,
}) => {
  const customIcon = <LoadingOutlined style={{ fontSize: 24, color: '#1677ff' }} spin />;

  // 根据类型渲染不同的加载状态
  const renderLoading = () => {
    switch (type) {
      case 'skeleton':
        return (
          <div style={style} className={className}>
            <Skeleton
              active
              paragraph={{ rows }}
              title={{ width: '60%' }}
              loading={spinning}
            >
              {children}
            </Skeleton>
          </div>
        );

      case 'card':
        return (
          <Card
            style={{ 
              ...style, 
              height: height || 200,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center'
            }}
            className={className}
            bodyStyle={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              height: '100%'
            }}
          >
            {spinning ? (
              <Space direction="vertical" align="center">
                <Spin indicator={customIcon} size={size} />
                <span style={{ color: '#666', fontSize: 14 }}>{tip}</span>
              </Space>
            ) : (
              children
            )}
          </Card>
        );

      case 'overlay':
        return (
          <div style={{ 
            position: 'relative', 
            minHeight: height || 100,
            ...style 
          }} className={className}>
            <Spin 
              spinning={spinning} 
              tip={tip} 
              size={size}
              indicator={customIcon}
              style={{
                maxHeight: 'none',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center'
              }}
            >
              <div style={{ 
                minHeight: height || 100,
                width: '100%',
                opacity: spinning ? 0.5 : 1,
                transition: 'opacity 0.3s ease'
              }}>
                {children}
              </div>
            </Spin>
          </div>
        );

      case 'spinner':
      default:
        if (children) {
          return (
            <Spin 
              spinning={spinning} 
              tip={tip} 
              size={size}
              indicator={customIcon}
              style={style}
              className={className}
            >
              {children}
            </Spin>
          );
        }

        return (
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              padding: '48px 24px',
              minHeight: height || 100,
              ...style
            }}
            className={className}
          >
            <Space direction="vertical" align="center">
              <Spin indicator={customIcon} size={size} />
              <span style={{ color: '#666', fontSize: 14, marginTop: 8 }}>{tip}</span>
            </Space>
          </div>
        );
    }
  };

  return renderLoading();
};

// 页面级加载组件
export const PageLoading: React.FC<{ tip?: string }> = ({ 
  tip = '页面加载中...' 
}) => (
  <LoadingSpinner
    type="overlay"
    tip={tip}
    size="large"
    height="60vh"
    style={{
      background: '#fafafa',
      borderRadius: 8
    }}
  />
);

// 内容区加载组件
export const ContentLoading: React.FC<{ 
  tip?: string; 
  height?: string | number;
}> = ({ 
  tip = '内容加载中...', 
  height = 300 
}) => (
  <LoadingSpinner
    type="card"
    tip={tip}
    height={height}
  />
);

// 表格数据加载组件
export const TableLoading: React.FC<{ rows?: number }> = ({ rows = 5 }) => (
  <div style={{ padding: '24px' }}>
    <Skeleton
      active
      title={{ width: '30%' }}
      paragraph={{ rows, width: ['100%', '100%', '80%', '60%', '40%'] }}
    />
  </div>
);

// 按钮加载状态组件
export const ButtonLoading: React.FC<{
  loading?: boolean;
  children: React.ReactNode;
  [key: string]: any;
}> = ({ loading, children, ...props }) => (
  <Spin spinning={loading} size="small">
    {React.cloneElement(children as React.ReactElement, {
      ...props,
      loading,
    })}
  </Spin>
);

// 数据获取加载包装器
export const DataLoadingWrapper: React.FC<{
  loading: boolean;
  error?: Error | null;
  empty?: boolean;
  emptyText?: string;
  children: React.ReactNode;
  loadingComponent?: React.ReactNode;
  errorComponent?: React.ReactNode;
  emptyComponent?: React.ReactNode;
}> = ({
  loading,
  error,
  empty,
  emptyText = '暂无数据',
  children,
  loadingComponent,
  errorComponent,
  emptyComponent,
}) => {
  if (loading) {
    return loadingComponent || <ContentLoading />;
  }

  if (error) {
    return errorComponent || (
      <div style={{ 
        textAlign: 'center', 
        padding: '48px 24px',
        color: '#ff4d4f'
      }}>
        加载失败: {error.message}
      </div>
    );
  }

  if (empty) {
    return emptyComponent || (
      <div style={{ 
        textAlign: 'center', 
        padding: '48px 24px',
        color: '#999'
      }}>
        {emptyText}
      </div>
    );
  }

  return <>{children}</>;
};

export default LoadingSpinner;