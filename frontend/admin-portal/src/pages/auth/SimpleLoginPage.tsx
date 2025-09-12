import React, { useState } from 'react';
import { Form, Input, Button, Card, Typography, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { Navigate } from 'react-router-dom';
import { AuthService } from '@services/authService';

const { Title, Text } = Typography;

interface LoginForm {
  student_id: string;
  password: string;
}

const SimpleLoginPage: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  // 真实API登录处理
  const handleSubmit = async (values: LoginForm) => {
    setLoading(true);
    
    try {
      console.log('🚀 提交登录请求:', values);
      
      // 调用真实API
      const response = await AuthService.login(values);
      
      console.log('✅ 登录响应:', response);
      
      // 存储认证信息
      localStorage.setItem('auth-token', response.access_token);
      localStorage.setItem('refresh-token', response.refresh_token);
      localStorage.setItem('user-info', JSON.stringify(response.user));
      localStorage.setItem('simple-auth', 'true');
      
      message.success('登录成功');
      setIsAuthenticated(true);
      
    } catch (error: any) {
      console.error('❌ 登录错误:', error);
      
      const errorMessage = error.response?.data?.message || error.message || '登录失败';
      message.error(errorMessage);
      
      form.setFields([
        { name: 'student_id', errors: [''] },
        { name: 'password', errors: [errorMessage] },
      ]);
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
              name="student_id"
              rules={[
                { required: true, message: '请输入学生ID!' },
                { min: 3, message: '学生ID至少3个字符!' },
              ]}
            >
              <Input
                prefix={<UserOutlined />}
                placeholder="学生ID"
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