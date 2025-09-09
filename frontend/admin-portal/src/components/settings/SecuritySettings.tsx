import React from 'react';
import {
  Form,
  Input,
  InputNumber,
  Select,
  Switch,
  Card,
  Button,
  Space,
  Typography,
  Divider,
  Row,
  Col,
  Alert,
  Slider,
  Checkbox,
} from 'antd';
import {
  SecurityScanOutlined,
  SaveOutlined,
  ReloadOutlined,
  LockOutlined,
  KeyOutlined,
  SafetyCertificateOutlined,
  UserOutlined,
} from '@ant-design/icons';
import type { SystemSettings } from '../../types';

const { Title, Text, Paragraph } = Typography;

export interface SecuritySettingsProps {
  settings?: SystemSettings;
  loading?: boolean;
  onSave: (values: Partial<SystemSettings>) => Promise<void>;
}

const PASSWORD_COMPLEXITY_OPTIONS = [
  { label: '包含大写字母', value: 'uppercase' },
  { label: '包含小写字母', value: 'lowercase' },
  { label: '包含数字', value: 'numbers' },
  { label: '包含特殊字符', value: 'symbols' },
];

const LOG_LEVELS = [
  { value: 'debug', label: 'Debug', color: '#1677ff' },
  { value: 'info', label: 'Info', color: '#52c41a' },
  { value: 'warn', label: 'Warning', color: '#faad14' },
  { value: 'error', label: 'Error', color: '#ff4d4f' },
];

export const SecuritySettings: React.FC<SecuritySettingsProps> = ({
  settings,
  loading,
  onSave,
}) => {
  const [form] = Form.useForm();

  React.useEffect(() => {
    if (settings) {
      form.setFieldsValue({
        password_min_length: settings.password_min_length,
        password_complexity: settings.password_complexity,
        session_timeout: settings.session_timeout / 60, // 转换为分钟
        max_login_attempts: settings.max_login_attempts,
        enable_logging: settings.enable_logging,
        log_level: settings.log_level,
      });
    }
  }, [settings, form]);

  const handleSubmit = async (values: any) => {
    // 转换session_timeout从分钟到秒
    const submissionValues = {
      ...values,
      session_timeout: values.session_timeout * 60,
    };
    await onSave(submissionValues);
  };

  const handleReset = () => {
    form.resetFields();
    if (settings) {
      form.setFieldsValue({
        ...settings,
        session_timeout: settings.session_timeout / 60,
      });
    }
  };

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      {/* 密码安全设置 */}
      <Card
        title={
          <Space>
            <LockOutlined />
            密码安全设置
          </Space>
        }
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={handleReset}>
              重置
            </Button>
            <Button 
              type="primary" 
              icon={<SaveOutlined />} 
              onClick={() => form.submit()}
              loading={loading}
            >
              保存设置
            </Button>
          </Space>
        }
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          size="large"
        >
          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                name="password_min_length"
                label="密码最小长度"
                rules={[
                  { required: true, message: '请设置密码最小长度' },
                  { type: 'number', min: 6, max: 50, message: '密码长度应在6-50字符之间' },
                ]}
              >
                <InputNumber
                  min={6}
                  max={50}
                  style={{ width: '100%' }}
                  addonAfter="字符"
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="password_complexity"
                label="密码复杂度"
                valuePropName="checked"
                tooltip="启用后要求密码包含大小写字母、数字和特殊字符"
              >
                <Switch
                  checkedChildren="严格"
                  unCheckedChildren="宽松"
                />
              </Form.Item>
            </Col>
          </Row>

          <Alert
            message="密码安全提示"
            description="强密码策略可以有效防止暴力破解攻击，建议启用密码复杂度要求。"
            type="info"
            showIcon
            style={{ marginBottom: 24 }}
          />
        </Form>
      </Card>

      {/* 会话安全设置 */}
      <Card
        title={
          <Space>
            <KeyOutlined />
            会话安全设置
          </Space>
        }
      >
        <Form form={form} layout="vertical" size="large">
          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                name="session_timeout"
                label="会话超时时间"
                rules={[
                  { required: true, message: '请设置会话超时时间' },
                  { type: 'number', min: 5, max: 1440, message: '会话时间应在5-1440分钟之间' },
                ]}
              >
                <InputNumber
                  min={5}
                  max={1440}
                  style={{ width: '100%' }}
                  addonAfter="分钟"
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="max_login_attempts"
                label="最大登录尝试次数"
                rules={[
                  { required: true, message: '请设置最大登录尝试次数' },
                  { type: 'number', min: 3, max: 20, message: '尝试次数应在3-20次之间' },
                ]}
              >
                <InputNumber
                  min={3}
                  max={20}
                  style={{ width: '100%' }}
                  addonAfter="次"
                />
              </Form.Item>
            </Col>
          </Row>

          <Alert
            message="会话安全说明"
            description="会话超时后用户需要重新登录。过多的登录失败尝试会导致账户暂时锁定。"
            type="warning"
            showIcon
            style={{ marginBottom: 24 }}
          />
        </Form>
      </Card>

      {/* 日志与监控设置 */}
      <Card
        title={
          <Space>
            <SafetyCertificateOutlined />
            日志与监控设置
          </Space>
        }
      >
        <Form form={form} layout="vertical" size="large">
          <Row gutter={24}>
            <Col span={12}>
              <Form.Item
                name="enable_logging"
                label="启用系统日志"
                valuePropName="checked"
                tooltip="启用后记录系统操作和安全事件"
              >
                <Switch
                  checkedChildren="启用"
                  unCheckedChildren="禁用"
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="log_level"
                label="日志级别"
                tooltip="选择要记录的日志最低级别"
              >
                <Select placeholder="请选择日志级别">
                  {LOG_LEVELS.map(level => (
                    <Select.Option key={level.value} value={level.value}>
                      <Space>
                        <div 
                          style={{ 
                            width: 8, 
                            height: 8, 
                            borderRadius: '50%', 
                            backgroundColor: level.color 
                          }} 
                        />
                        {level.label}
                      </Space>
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Card>

      {/* 安全建议 */}
      <Card
        title={
          <Space>
            <SecurityScanOutlined />
            安全建议
          </Space>
        }
      >
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Alert
            message="密码安全"
            description={
              <ul style={{ margin: '8px 0 0 0', paddingLeft: '16px' }}>
                <li>设置足够长度的密码要求（建议8位以上）</li>
                <li>启用密码复杂度要求，确保密码强度</li>
                <li>定期提醒用户更改密码</li>
              </ul>
            }
            type="info"
            showIcon
          />

          <Alert
            message="会话管理"
            description={
              <ul style={{ margin: '8px 0 0 0', paddingLeft: '16px' }}>
                <li>设置合适的会话超时时间（建议30-120分钟）</li>
                <li>限制登录尝试次数防止暴力破解</li>
                <li>监控异常登录行为</li>
              </ul>
            }
            type="warning"
            showIcon
          />

          <Alert
            message="系统监控"
            description={
              <ul style={{ margin: '8px 0 0 0', paddingLeft: '16px' }}>
                <li>启用系统日志记录重要操作</li>
                <li>设置适当的日志级别平衡安全与性能</li>
                <li>定期检查系统日志发现异常</li>
              </ul>
            }
            type="success"
            showIcon
          />
        </Space>
      </Card>
    </Space>
  );
};