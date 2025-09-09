import React, { useEffect } from 'react';
import {
  Form,
  Input,
  Select,
  InputNumber,
  Switch,
  Space,
  Button,
  Card,
  Divider,
  Alert,
  Typography,
} from 'antd';
import {
  UserOutlined,
  MailOutlined,
  LockOutlined,
  CrownOutlined,
  DatabaseOutlined,
} from '@ant-design/icons';
import { formatFileSize } from '@utils/index';
import type { User, CreateUserRequest, UpdateUserRequest } from '../../types';

const { Text } = Typography;
const { TextArea } = Input;

export interface UserFormProps {
  user?: User;
  loading?: boolean;
  onSubmit: (data: CreateUserRequest | UpdateUserRequest) => Promise<void>;
  onCancel?: () => void;
  mode?: 'create' | 'edit';
}

// 存储配额选项
const STORAGE_QUOTA_OPTIONS = [
  { value: 1 * 1024 * 1024 * 1024, label: '1 GB' },
  { value: 5 * 1024 * 1024 * 1024, label: '5 GB' },
  { value: 10 * 1024 * 1024 * 1024, label: '10 GB' },
  { value: 50 * 1024 * 1024 * 1024, label: '50 GB' },
  { value: 100 * 1024 * 1024 * 1024, label: '100 GB' },
  { value: 500 * 1024 * 1024 * 1024, label: '500 GB' },
  { value: 1024 * 1024 * 1024 * 1024, label: '1 TB' },
];

// 角色选项
const ROLE_OPTIONS = [
  { value: 'user', label: '普通用户', icon: <UserOutlined /> },
  { value: 'moderator', label: '版主', icon: <CrownOutlined /> },
  { value: 'admin', label: '管理员', icon: <CrownOutlined /> },
];

export const UserForm: React.FC<UserFormProps> = ({
  user,
  loading = false,
  onSubmit,
  onCancel,
  mode = 'create',
}) => {
  const [form] = Form.useForm();
  const isEditing = mode === 'edit';

  // 初始化表单数据
  useEffect(() => {
    if (user && isEditing) {
      form.setFieldsValue({
        username: user.username,
        email: user.email,
        display_name: user.display_name,
        role: user.role,
        storage_quota: user.storage_quota,
        status: user.status === 1,
      });
    } else if (!isEditing) {
      // 创建模式的默认值
      form.setFieldsValue({
        role: 'user',
        storage_quota: 10 * 1024 * 1024 * 1024, // 10GB
        status: true,
      });
    }
  }, [user, isEditing, form]);

  // 表单提交
  const handleSubmit = async (values: any) => {
    try {
      const formData = {
        ...values,
        status: values.status ? 1 : 0,
      };

      await onSubmit(formData);
      
      if (!isEditing) {
        form.resetFields();
      }
    } catch (error) {
      // 错误处理已在 mutation 中完成
      console.error('Form submission error:', error);
    }
  };

  // 自定义验证规则
  const validateUsername = async (rule: any, value: string) => {
    if (!value) return;
    
    if (value.length < 3) {
      throw new Error('用户名至少3个字符');
    }
    
    if (!/^[a-zA-Z0-9_-]+$/.test(value)) {
      throw new Error('用户名只能包含字母、数字、下划线和连字符');
    }
  };

  const validateEmail = async (rule: any, value: string) => {
    if (!value) return;
    
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(value)) {
      throw new Error('请输入有效的邮箱地址');
    }
  };

  const validatePassword = async (rule: any, value: string) => {
    if (!value && !isEditing) {
      throw new Error('密码不能为空');
    }
    
    if (value && value.length < 6) {
      throw new Error('密码至少6个字符');
    }
  };

  return (
    <div style={{ maxWidth: 800, margin: '0 auto' }}>
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        scrollToFirstError
        size="large"
      >
        {/* 基本信息 */}
        <Card title="基本信息" style={{ marginBottom: 16 }}>
          <Form.Item
            name="username"
            label="用户名"
            rules={[
              { required: true, message: '请输入用户名' },
              { validator: validateUsername },
            ]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="请输入用户名"
              disabled={isEditing} // 编辑模式下不允许修改用户名
            />
          </Form.Item>

          <Form.Item
            name="email"
            label="邮箱地址"
            rules={[
              { required: true, message: '请输入邮箱地址' },
              { validator: validateEmail },
            ]}
          >
            <Input
              prefix={<MailOutlined />}
              placeholder="请输入邮箱地址"
              type="email"
            />
          </Form.Item>

          <Form.Item
            name="display_name"
            label="显示名称"
            rules={[
              { max: 50, message: '显示名称不能超过50个字符' },
            ]}
          >
            <Input
              placeholder="请输入显示名称（可选）"
            />
          </Form.Item>

          {!isEditing && (
            <Form.Item
              name="password"
              label="密码"
              rules={[
                { validator: validatePassword },
              ]}
            >
              <Input.Password
                prefix={<LockOutlined />}
                placeholder="请输入密码"
                visibilityToggle
              />
            </Form.Item>
          )}

          {isEditing && (
            <Alert
              message="密码修改"
              description="如需修改密码，请使用重置密码功能。"
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />
          )}
        </Card>

        {/* 权限设置 */}
        <Card title="权限设置" style={{ marginBottom: 16 }}>
          <Form.Item
            name="role"
            label="用户角色"
            rules={[{ required: true, message: '请选择用户角色' }]}
          >
            <Select placeholder="请选择用户角色">
              {ROLE_OPTIONS.map(option => (
                <Select.Option key={option.value} value={option.value}>
                  <Space>
                    {option.icon}
                    {option.label}
                  </Space>
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="status"
            label="账户状态"
            valuePropName="checked"
          >
            <Switch
              checkedChildren="启用"
              unCheckedChildren="禁用"
            />
          </Form.Item>
        </Card>

        {/* 存储配额 */}
        <Card title="存储配额" style={{ marginBottom: 24 }}>
          <Form.Item
            name="storage_quota"
            label="存储空间"
            rules={[{ required: true, message: '请设置存储配额' }]}
          >
            <Select
              placeholder="请选择存储配额"
              optionLabelProp="label"
            >
              {STORAGE_QUOTA_OPTIONS.map(option => (
                <Select.Option 
                  key={option.value} 
                  value={option.value}
                  label={option.label}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span>{option.label}</span>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {formatFileSize(option.value)}
                    </Text>
                  </div>
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            shouldUpdate={(prevValues, currentValues) =>
              prevValues.storage_quota !== currentValues.storage_quota
            }
            style={{ marginBottom: 0 }}
          >
            {({ getFieldValue }) => {
              const quota = getFieldValue('storage_quota');
              return quota ? (
                <Alert
                  message={`已选择存储配额: ${formatFileSize(quota)}`}
                  type="info"
                  showIcon
                  style={{ marginTop: 8 }}
                />
              ) : null;
            }}
          </Form.Item>

          {isEditing && user && (
            <div style={{ marginTop: 16, padding: 16, background: '#fafafa', borderRadius: 6 }}>
              <Text strong>当前存储使用情况</Text>
              <div style={{ marginTop: 8 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                  <Text>已使用:</Text>
                  <Text>{formatFileSize(user.storage_used)}</Text>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Text>使用率:</Text>
                  <Text>
                    {user.storage_quota > 0 
                      ? Math.round((user.storage_used / user.storage_quota) * 100)
                      : 0
                    }%
                  </Text>
                </div>
              </div>
            </div>
          )}
        </Card>

        {/* 操作按钮 */}
        <div style={{ textAlign: 'right' }}>
          <Space>
            {onCancel && (
              <Button size="large" onClick={onCancel}>
                取消
              </Button>
            )}
            <Button
              type="primary"
              size="large"
              htmlType="submit"
              loading={loading}
              icon={isEditing ? <DatabaseOutlined /> : <UserOutlined />}
            >
              {isEditing ? '更新用户' : '创建用户'}
            </Button>
          </Space>
        </div>
      </Form>
    </div>
  );
};