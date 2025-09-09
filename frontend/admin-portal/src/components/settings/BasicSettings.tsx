import React from 'react';
import {
  Form,
  Input,
  InputNumber,
  Switch,
  Card,
  Button,
  Space,
  Typography,
  Divider,
  Row,
  Col,
} from 'antd';
import {
  SettingOutlined,
  SaveOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import type { SystemSettings } from '../../types';

const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;

export interface BasicSettingsProps {
  settings?: SystemSettings;
  loading?: boolean;
  onSave: (values: Partial<SystemSettings>) => Promise<void>;
}

export const BasicSettings: React.FC<BasicSettingsProps> = ({
  settings,
  loading,
  onSave,
}) => {
  const [form] = Form.useForm();

  React.useEffect(() => {
    if (settings) {
      form.setFieldsValue({
        system_name: settings.system_name,
        system_description: settings.system_description,
        enable_registration: settings.enable_registration,
        require_email_verification: settings.require_email_verification,
        maintenance_mode: settings.maintenance_mode,
        maintenance_message: settings.maintenance_message,
      });
    }
  }, [settings, form]);

  const handleSubmit = async (values: any) => {
    await onSave(values);
  };

  const handleReset = () => {
    form.resetFields();
    if (settings) {
      form.setFieldsValue(settings);
    }
  };

  return (
    <Card
      title={
        <Space>
          <SettingOutlined />
          基本设置
        </Space>
      }
      extra={
        <Space>
          <Button 
            icon={<ReloadOutlined />} 
            onClick={handleReset}
          >
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
      >
        <Row gutter={[24, 16]}>
          <Col span={12}>
            <Form.Item
              name="system_name"
              label="系统名称"
              rules={[
                { required: true, message: '请输入系统名称' },
                { max: 100, message: '系统名称不能超过100个字符' },
              ]}
            >
              <Input placeholder="请输入系统名称" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              name="system_version"
              label="系统版本"
            >
              <Input 
                value={settings?.system_version} 
                disabled 
                placeholder="系统版本（只读）"
              />
            </Form.Item>
          </Col>
        </Row>

        <Form.Item
          name="system_description"
          label="系统描述"
          rules={[
            { max: 500, message: '系统描述不能超过500个字符' },
          ]}
        >
          <TextArea 
            rows={3}
            placeholder="请输入系统描述（可选）"
            showCount
            maxLength={500}
          />
        </Form.Item>

        <Divider orientation="left">
          <Text strong>用户注册设置</Text>
        </Divider>

        <Row gutter={[24, 16]}>
          <Col span={12}>
            <Form.Item
              name="enable_registration"
              label="开放用户注册"
              valuePropName="checked"
              tooltip="开启后允许新用户自行注册账户"
            >
              <Switch
                checkedChildren="开启"
                unCheckedChildren="关闭"
              />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              name="require_email_verification"
              label="邮箱验证"
              valuePropName="checked"
              tooltip="开启后新用户注册需要验证邮箱"
            >
              <Switch
                checkedChildren="必须"
                unCheckedChildren="可选"
              />
            </Form.Item>
          </Col>
        </Row>

        <Divider orientation="left">
          <Text strong>系统维护</Text>
        </Divider>

        <Form.Item
          name="maintenance_mode"
          label="维护模式"
          valuePropName="checked"
          tooltip="开启维护模式后，普通用户无法访问系统"
        >
          <Switch
            checkedChildren="维护中"
            unCheckedChildren="正常运行"
          />
        </Form.Item>

        <Form.Item
          shouldUpdate={(prevValues, currentValues) =>
            prevValues.maintenance_mode !== currentValues.maintenance_mode
          }
        >
          {({ getFieldValue }) => {
            return getFieldValue('maintenance_mode') ? (
              <Form.Item
                name="maintenance_message"
                label="维护提示信息"
                rules={[
                  { required: true, message: '请输入维护提示信息' },
                  { max: 200, message: '维护提示信息不能超过200个字符' },
                ]}
              >
                <TextArea 
                  rows={2}
                  placeholder="请输入维护期间显示给用户的提示信息"
                  showCount
                  maxLength={200}
                />
              </Form.Item>
            ) : null;
          }}
        </Form.Item>

        <div style={{ marginTop: 24, padding: 16, background: '#f5f5f5', borderRadius: 6 }}>
          <Paragraph type="secondary" style={{ margin: 0, fontSize: 12 }}>
            <Text strong>提示：</Text>
            修改基本设置后，部分配置需要重启相关服务才能生效。
            请在系统信息页面查看服务状态并按需重启。
          </Paragraph>
        </div>
      </Form>
    </Card>
  );
};