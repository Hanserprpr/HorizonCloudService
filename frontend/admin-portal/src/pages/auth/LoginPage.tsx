import React from 'react';
import { Form, Input, Button, Card, Checkbox, message, Typography } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '@hooks/useAuth';
import { useResponsive } from '@hooks/useResponsive';
import { useErrorHandler } from '@hooks/useErrorHandler';
import { ROUTES, APP_NAME } from '@constants/index';
import type { LoginForm } from '@types/index';

const { Title, Text } = Typography;

const LoginPage: React.FC = () => {
  const { login, isAuthenticated, loading, loginError } = useAuth();
  const { isMobile } = useResponsive();
  const { handleError, showError } = useErrorHandler();
  const location = useLocation();
  const [form] = Form.useForm();

  // 已登录用户重定向
  if (isAuthenticated) {
    const from = (location.state as any)?.from || ROUTES.DASHBOARD;
    return <Navigate to={from} replace />;
  }

  const handleSubmit = async (values: LoginForm) => {
    try {
      await login(values);
      // 登录成功后会自动重定向
    } catch (error) {
      // 使用新的错误处理系统
      const errorInfo = handleError(error, {
        showMessage: true,
        showNotification: false, // 登录错误使用message而不是notification
      });
      
      // 如果是认证错误，高亮表单字段
      if (errorInfo.type === 'auth' || errorInfo.status === 401) {
        form.setFields([
          { name: 'username', errors: [''] },
          { name: 'password', errors: [errorInfo.message] },
        ]);
      } else if (errorInfo.type === 'validation') {
        // 处理表单验证错误
        showError(errorInfo.message);
      }
    }
  };

  return (
    <div className="login-container">
      <div className="login-wrapper">
        <Card className="login-card" bordered={false}>
          <div className="login-header">
            <Title level={2} className="login-title">
              {APP_NAME}
            </Title>
            <Text type="secondary" className="login-subtitle">
              管理员登录
            </Text>
          </div>

          <Form
            form={form}
            name="login"
            onFinish={handleSubmit}
            autoComplete="off"
            size="large"
            layout="vertical"
          >
            <Form.Item
              name="username"
              rules={[
                { required: true, message: '请输入用户名!' },
                { min: 3, message: '用户名至少3个字符!' },
              ]}
            >
              <Input
                prefix={<UserOutlined />}
                placeholder="用户名"
                autoComplete="username"
              />
            </Form.Item>

            <Form.Item
              name="password"
              rules={[
                { required: true, message: '请输入密码!' },
                { min: 6, message: '密码至少6个字符!' },
              ]}
            >
              <Input.Password
                prefix={<LockOutlined />}
                placeholder="密码"
                autoComplete="current-password"
              />
            </Form.Item>

            <Form.Item name="remember" valuePropName="checked">
              <Checkbox>记住登录状态</Checkbox>
            </Form.Item>

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                className="login-button"
                loading={loading}
                block
              >
                登录
              </Button>
            </Form.Item>
          </Form>

          <div className="login-demo-info">
            <Text type="secondary" style={{ fontSize: '12px' }}>
              演示账户: admin / admin123
            </Text>
          </div>
        </Card>
      </div>

      <style jsx>{`
        .login-container {
          min-height: 100vh;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 20px;
        }

        .login-wrapper {
          width: 100%;
          max-width: ${isMobile ? '100%' : '400px'};
        }

        .login-card {
          box-shadow: 0 4px 24px rgba(0, 0, 0, 0.1);
          border-radius: 8px;
        }

        .login-header {
          text-align: center;
          margin-bottom: 32px;
        }

        .login-title {
          margin-bottom: 8px !important;
          color: #1a1a1a;
        }

        .login-subtitle {
          font-size: 14px;
        }

        .login-button {
          height: 44px;
          font-size: 16px;
          margin-top: 8px;
        }

        .login-demo-info {
          text-align: center;
          margin-top: 16px;
          padding-top: 16px;
          border-top: 1px solid #f0f0f0;
        }
      `}</style>
    </div>
  );
};

export default LoginPage;