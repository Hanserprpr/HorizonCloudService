import React, { Component, ErrorInfo, ReactNode } from 'react';
import { Button, Result, Card, Typography, Space, Collapse } from 'antd';
import { 
  ExclamationCircleOutlined, 
  ReloadOutlined, 
  BugOutlined,
  WarningOutlined 
} from '@ant-design/icons';

const { Text, Paragraph } = Typography;
const { Panel } = Collapse;

interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
  errorInfo?: ErrorInfo;
  errorId?: string;
}

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo, errorId: string) => void;
}

/**
 * React Error Boundary 组件
 * 用于捕获子组件中的 JavaScript 错误，记录错误并显示友好的错误界面
 */
export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
    };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    // 更新 state 使下一次渲染能够显示降级后的 UI
    const errorId = `error-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    return {
      hasError: true,
      error,
      errorId,
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    const errorId = this.state.errorId || 'unknown';
    
    // 记录错误到控制台
    console.error('ErrorBoundary caught an error:', {
      error,
      errorInfo,
      errorId,
      timestamp: new Date().toISOString(),
      userAgent: navigator.userAgent,
      url: window.location.href,
    });

    // 更新状态以包含错误信息
    this.setState({
      errorInfo,
    });

    // 调用外部错误处理函数
    if (this.props.onError) {
      this.props.onError(error, errorInfo, errorId);
    }

    // 在实际应用中，你可能想要将错误报告给错误监控服务
    // 例如: Sentry.captureException(error);
  }

  handleReload = () => {
    // 重新加载页面
    window.location.reload();
  };

  handleReset = () => {
    // 重置错误状态
    this.setState({
      hasError: false,
      error: undefined,
      errorInfo: undefined,
      errorId: undefined,
    });
  };

  render() {
    if (this.state.hasError) {
      // 如果提供了自定义 fallback，使用它
      if (this.props.fallback) {
        return this.props.fallback;
      }

      const { error, errorInfo, errorId } = this.state;
      const isDevelopment = process.env.NODE_ENV === 'development';

      return (
        <div style={{ 
          padding: '48px 24px', 
          minHeight: '50vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center'
        }}>
          <Card style={{ maxWidth: 600, width: '100%' }}>
            <Result
              status="error"
              icon={<ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />}
              title="页面出现错误"
              subTitle="很抱歉，页面遇到了一些问题。请尝试刷新页面或联系管理员。"
              extra={[
                <Space key="actions" direction="vertical" style={{ width: '100%' }}>
                  <Space>
                    <Button 
                      type="primary" 
                      icon={<ReloadOutlined />}
                      onClick={this.handleReload}
                    >
                      刷新页面
                    </Button>
                    <Button 
                      icon={<BugOutlined />}
                      onClick={this.handleReset}
                    >
                      重试
                    </Button>
                  </Space>
                  
                  {isDevelopment && (
                    <div style={{ width: '100%', textAlign: 'left', marginTop: 24 }}>
                      <Collapse ghost>
                        <Panel 
                          header={
                            <Space>
                              <WarningOutlined style={{ color: '#fa8c16' }} />
                              <Text strong>错误详情 (开发模式)</Text>
                            </Space>
                          }
                          key="error-details"
                        >
                          <div style={{ marginBottom: 16 }}>
                            <Text strong>错误ID：</Text>
                            <Text code copyable>{errorId}</Text>
                          </div>
                          
                          <div style={{ marginBottom: 16 }}>
                            <Text strong>错误信息：</Text>
                            <Paragraph style={{ marginTop: 8 }}>
                              <Text code style={{ 
                                display: 'block', 
                                padding: 12, 
                                background: '#fff2f0',
                                border: '1px solid #ffccc7',
                                borderRadius: 4,
                                wordBreak: 'break-all'
                              }}>
                                {error?.message || 'Unknown error'}
                              </Text>
                            </Paragraph>
                          </div>

                          <div style={{ marginBottom: 16 }}>
                            <Text strong>错误堆栈：</Text>
                            <Paragraph style={{ marginTop: 8 }}>
                              <Text code style={{ 
                                display: 'block', 
                                padding: 12, 
                                background: '#f6f6f6',
                                border: '1px solid #d9d9d9',
                                borderRadius: 4,
                                fontSize: 12,
                                fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
                                whiteSpace: 'pre-wrap',
                                wordBreak: 'break-all',
                                maxHeight: 200,
                                overflow: 'auto'
                              }}>
                                {error?.stack || 'No stack trace available'}
                              </Text>
                            </Paragraph>
                          </div>

                          {errorInfo && (
                            <div>
                              <Text strong>组件堆栈：</Text>
                              <Paragraph style={{ marginTop: 8 }}>
                                <Text code style={{ 
                                  display: 'block', 
                                  padding: 12, 
                                  background: '#f6f6f6',
                                  border: '1px solid #d9d9d9',
                                  borderRadius: 4,
                                  fontSize: 12,
                                  fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
                                  whiteSpace: 'pre-wrap',
                                  wordBreak: 'break-all',
                                  maxHeight: 200,
                                  overflow: 'auto'
                                }}>
                                  {errorInfo.componentStack}
                                </Text>
                              </Paragraph>
                            </div>
                          )}
                        </Panel>
                      </Collapse>
                    </div>
                  )}
                </Space>
              ]}
            />
          </Card>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;