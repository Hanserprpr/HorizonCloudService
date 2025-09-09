import React, { useState } from 'react';
import {
  Form,
  Input,
  InputNumber,
  Select,
  Card,
  Button,
  Space,
  Typography,
  Divider,
  Row,
  Col,
  Alert,
  Modal,
  Spin,
  Tag,
} from 'antd';
import {
  DatabaseOutlined,
  SaveOutlined,
  ReloadOutlined,
  LinkOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import { formatFileSize } from '@utils/index';
import type { SystemSettings, StorageConfig } from '../../types';

const { Title, Text, Paragraph } = Typography;
const { confirm } = Modal;

export interface StorageSettingsProps {
  settings?: SystemSettings;
  storageConfig?: StorageConfig;
  loading?: boolean;
  onSaveSettings: (values: Partial<SystemSettings>) => Promise<void>;
  onSaveStorageConfig: (config: StorageConfig) => Promise<void>;
  onTestConnection: (config: StorageConfig) => Promise<void>;
}

// 存储后端选项
const STORAGE_BACKENDS = [
  { value: 'local', label: '本地存储', icon: '🗂️' },
  { value: 'minio', label: 'MinIO', icon: '🪣' },
  { value: 's3', label: 'AWS S3', icon: '☁️' },
  { value: 'oss', label: '阿里云OSS', icon: '🌥️' },
];

// 预定义的存储配额选项
const STORAGE_QUOTA_OPTIONS = [
  { value: 1 * 1024 * 1024 * 1024, label: '1 GB' },
  { value: 5 * 1024 * 1024 * 1024, label: '5 GB' },
  { value: 10 * 1024 * 1024 * 1024, label: '10 GB' },
  { value: 50 * 1024 * 1024 * 1024, label: '50 GB' },
  { value: 100 * 1024 * 1024 * 1024, label: '100 GB' },
  { value: 500 * 1024 * 1024 * 1024, label: '500 GB' },
  { value: 1024 * 1024 * 1024 * 1024, label: '1 TB' },
];

export const StorageSettings: React.FC<StorageSettingsProps> = ({
  settings,
  storageConfig,
  loading,
  onSaveSettings,
  onSaveStorageConfig,
  onTestConnection,
}) => {
  const [settingsForm] = Form.useForm();
  const [storageForm] = Form.useForm();
  const [testingConnection, setTestingConnection] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<'success' | 'error' | null>(null);

  React.useEffect(() => {
    if (settings) {
      settingsForm.setFieldsValue({
        default_storage_quota: settings.default_storage_quota,
        max_storage_quota: settings.max_storage_quota,
        storage_backend: settings.storage_backend,
      });
    }
  }, [settings, settingsForm]);

  React.useEffect(() => {
    if (storageConfig) {
      storageForm.setFieldsValue({
        backend: storageConfig.backend,
        ...storageConfig.config,
      });
    }
  }, [storageConfig, storageForm]);

  const handleSettingsSubmit = async (values: any) => {
    await onSaveSettings(values);
  };

  const handleStorageConfigSubmit = async (values: any) => {
    const { backend, ...config } = values;
    
    confirm({
      title: '确认更新存储配置',
      icon: <ExclamationCircleOutlined />,
      content: '更新存储配置可能会影响文件访问，确定要继续吗？',
      okText: '确认更新',
      cancelText: '取消',
      onOk: async () => {
        await onSaveStorageConfig({
          backend,
          config,
        });
      },
    });
  };

  const handleTestConnection = async () => {
    const values = storageForm.getFieldsValue();
    const { backend, ...config } = values;
    
    setTestingConnection(true);
    setConnectionStatus(null);
    
    try {
      await onTestConnection({
        backend,
        config,
      });
      setConnectionStatus('success');
    } catch (error) {
      setConnectionStatus('error');
    } finally {
      setTestingConnection(false);
    }
  };

  const renderStorageConfigForm = (backend: string) => {
    switch (backend) {
      case 'local':
        return (
          <Form.Item
            name="path"
            label="存储路径"
            rules={[{ required: true, message: '请输入本地存储路径' }]}
          >
            <Input 
              placeholder="/var/lib/app/storage" 
              addonBefore="Path"
            />
          </Form.Item>
        );

      case 'minio':
        return (
          <>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="endpoint"
                  label="MinIO端点"
                  rules={[{ required: true, message: '请输入MinIO端点地址' }]}
                >
                  <Input placeholder="localhost:9000" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name="bucket"
                  label="存储桶名称"
                  rules={[{ required: true, message: '请输入存储桶名称' }]}
                >
                  <Input placeholder="media-storage" />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="access_key"
                  label="访问密钥ID"
                  rules={[{ required: true, message: '请输入访问密钥ID' }]}
                >
                  <Input placeholder="minioadmin" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name="secret_key"
                  label="访问密钥密码"
                  rules={[{ required: true, message: '请输入访问密钥密码' }]}
                >
                  <Input.Password placeholder="minioadmin" />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="region"
                  label="区域"
                >
                  <Input placeholder="us-east-1" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name="use_ssl"
                  label="使用SSL"
                  valuePropName="checked"
                >
                  <Select defaultValue={false}>
                    <Select.Option value={true}>是</Select.Option>
                    <Select.Option value={false}>否</Select.Option>
                  </Select>
                </Form.Item>
              </Col>
            </Row>
          </>
        );

      case 's3':
        return (
          <>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="aws_bucket"
                  label="S3存储桶"
                  rules={[{ required: true, message: '请输入S3存储桶名称' }]}
                >
                  <Input placeholder="my-bucket" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name="aws_region"
                  label="AWS区域"
                  rules={[{ required: true, message: '请选择AWS区域' }]}
                >
                  <Select placeholder="请选择区域">
                    <Select.Option value="us-east-1">US East (N. Virginia)</Select.Option>
                    <Select.Option value="us-west-2">US West (Oregon)</Select.Option>
                    <Select.Option value="eu-west-1">Europe (Ireland)</Select.Option>
                    <Select.Option value="ap-southeast-1">Asia Pacific (Singapore)</Select.Option>
                    <Select.Option value="ap-northeast-1">Asia Pacific (Tokyo)</Select.Option>
                  </Select>
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="aws_access_key_id"
                  label="Access Key ID"
                  rules={[{ required: true, message: '请输入Access Key ID' }]}
                >
                  <Input placeholder="AKIA..." />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name="aws_secret_access_key"
                  label="Secret Access Key"
                  rules={[{ required: true, message: '请输入Secret Access Key' }]}
                >
                  <Input.Password placeholder="..." />
                </Form.Item>
              </Col>
            </Row>
          </>
        );

      case 'oss':
        return (
          <>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="oss_bucket"
                  label="OSS存储桶"
                  rules={[{ required: true, message: '请输入OSS存储桶名称' }]}
                >
                  <Input placeholder="my-bucket" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name="oss_region"
                  label="OSS区域"
                  rules={[{ required: true, message: '请选择OSS区域' }]}
                >
                  <Select placeholder="请选择区域">
                    <Select.Option value="oss-cn-hangzhou">华东1（杭州）</Select.Option>
                    <Select.Option value="oss-cn-shanghai">华东2（上海）</Select.Option>
                    <Select.Option value="oss-cn-beijing">华北2（北京）</Select.Option>
                    <Select.Option value="oss-cn-shenzhen">华南1（深圳）</Select.Option>
                    <Select.Option value="oss-cn-hongkong">中国（香港）</Select.Option>
                  </Select>
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={8}>
                <Form.Item
                  name="oss_endpoint"
                  label="Endpoint"
                  rules={[{ required: true, message: '请输入Endpoint' }]}
                >
                  <Input placeholder="oss-cn-hangzhou.aliyuncs.com" />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item
                  name="oss_access_key_id"
                  label="AccessKey ID"
                  rules={[{ required: true, message: '请输入AccessKey ID' }]}
                >
                  <Input placeholder="LTAI..." />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item
                  name="oss_access_key_secret"
                  label="AccessKey Secret"
                  rules={[{ required: true, message: '请输入AccessKey Secret' }]}
                >
                  <Input.Password placeholder="..." />
                </Form.Item>
              </Col>
            </Row>
          </>
        );

      default:
        return null;
    }
  };

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      {/* 存储配额设置 */}
      <Card
        title={
          <Space>
            <DatabaseOutlined />
            存储配额设置
          </Space>
        }
        extra={
          <Space>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={() => settingsForm.resetFields()}
            >
              重置
            </Button>
            <Button 
              type="primary" 
              icon={<SaveOutlined />} 
              onClick={() => settingsForm.submit()}
              loading={loading}
            >
              保存设置
            </Button>
          </Space>
        }
      >
        <Form
          form={settingsForm}
          layout="vertical"
          onFinish={handleSettingsSubmit}
          size="large"
        >
          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                name="default_storage_quota"
                label="默认用户配额"
                rules={[{ required: true, message: '请设置默认用户配额' }]}
              >
                <Select placeholder="请选择默认配额">
                  {STORAGE_QUOTA_OPTIONS.map(option => (
                    <Select.Option key={option.value} value={option.value}>
                      {option.label} ({formatFileSize(option.value)})
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="max_storage_quota"
                label="最大用户配额"
                rules={[{ required: true, message: '请设置最大用户配额' }]}
              >
                <Select placeholder="请选择最大配额">
                  {STORAGE_QUOTA_OPTIONS.map(option => (
                    <Select.Option key={option.value} value={option.value}>
                      {option.label} ({formatFileSize(option.value)})
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Alert
            message="配额设置说明"
            description="默认配额会应用于新注册用户，最大配额限制了管理员可为用户设置的配额上限。"
            type="info"
            showIcon
            style={{ marginTop: 16 }}
          />
        </Form>
      </Card>

      {/* 存储后端配置 */}
      <Card
        title={
          <Space>
            <DatabaseOutlined />
            存储后端配置
            {connectionStatus === 'success' && (
              <Tag color="success" icon={<CheckCircleOutlined />}>连接正常</Tag>
            )}
            {connectionStatus === 'error' && (
              <Tag color="error" icon={<ExclamationCircleOutlined />}>连接失败</Tag>
            )}
          </Space>
        }
        extra={
          <Space>
            <Button 
              icon={<LinkOutlined />} 
              onClick={handleTestConnection}
              loading={testingConnection}
            >
              测试连接
            </Button>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={() => storageForm.resetFields()}
            >
              重置
            </Button>
            <Button 
              type="primary" 
              icon={<SaveOutlined />} 
              onClick={() => storageForm.submit()}
              loading={loading}
            >
              保存配置
            </Button>
          </Space>
        }
      >
        <Form
          form={storageForm}
          layout="vertical"
          onFinish={handleStorageConfigSubmit}
          size="large"
        >
          <Form.Item
            name="backend"
            label="存储后端"
            rules={[{ required: true, message: '请选择存储后端' }]}
          >
            <Select placeholder="请选择存储后端">
              {STORAGE_BACKENDS.map(backend => (
                <Select.Option key={backend.value} value={backend.value}>
                  {backend.icon} {backend.label}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            shouldUpdate={(prevValues, currentValues) =>
              prevValues.backend !== currentValues.backend
            }
          >
            {({ getFieldValue }) => {
              const backend = getFieldValue('backend');
              return backend ? (
                <div style={{ marginTop: 16 }}>
                  <Divider orientation="left">
                    <Text strong>{STORAGE_BACKENDS.find(b => b.value === backend)?.label} 配置</Text>
                  </Divider>
                  {renderStorageConfigForm(backend)}
                </div>
              ) : null;
            }}
          </Form.Item>
        </Form>

        <Alert
          message="重要提示"
          description="更改存储后端配置需要谨慎操作，建议先使用测试连接确认配置正确。更改后可能需要迁移现有文件数据。"
          type="warning"
          showIcon
          style={{ marginTop: 24 }}
        />
      </Card>
    </Space>
  );
};