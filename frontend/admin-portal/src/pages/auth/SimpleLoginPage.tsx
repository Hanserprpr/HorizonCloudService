import React, { useState } from 'react';
import { Form, Input, Button, Card, Typography, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { Navigate } from 'react-router-dom';

const { Title, Text } = Typography;

interface LoginForm {
  username: string;
  password: string;
}

const SimpleLoginPage: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  // 简化的登录处理
  const handleSubmit = async (values: LoginForm) => {
    setLoading(true);
    
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      if (values.username === 'admin' && values.password === 'admin123') {
        message.success('登录成功');
        localStorage.setItem('simple-auth', 'true');
        setIsAuthenticated(true);
      } else {
        message.error('用户名或密码错误');
        form.setFields([
          { name: 'username', errors: [''] },
          { name: 'password', errors: ['用户名或密码错误'] },
        ]);
      }
    } catch (error) {
      message.error('登录失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  // 登录成功后重定向
  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  return (
    <div style={{
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      padding: '20px',
    }}>
      <div style={{ width: '100%', maxWidth: '400px' }}>
        <Card style={{ borderRadius: '8px', boxShadow: '0 4px 24px rgba(0, 0, 0, 0.1)' }}>
          <div style={{ textAlign: 'center', marginBottom: '32px' }}>
            <Title level={2} style={{ marginBottom: '8px', color: '#1a1a1a' }}>
              云存储管理后台
            </Title>
            <Text type="secondary" style={{ fontSize: '14px' }}>
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

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                block
                style={{
                  height: '44px',
                  fontSize: '16px',
                  marginTop: '8px',
                }}
              >
                登录
              </Button>
            </Form.Item>
          </Form>

          <div style={{
            textAlign: 'center',
            marginTop: '16px',
            paddingTop: '16px',
            borderTop: '1px solid #f0f0f0',
          }}>
            <Text type="secondary" style={{ fontSize: '12px' }}>
              演示账户: admin / admin123
            </Text>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default SimpleLoginPage;