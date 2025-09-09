import React, { useState } from 'react';
import {
  Modal,
  Descriptions,
  Tag,
  Avatar,
  Progress,
  Tabs,
  Table,
  Space,
  Button,
  Statistic,
  Card,
  Row,
  Col,
  Typography,
  Empty,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  UserOutlined,
  MailOutlined,
  CalendarOutlined,
  DatabaseOutlined,
  FileOutlined,
  HistoryOutlined,
  BarChartOutlined,
} from '@ant-design/icons';
import { formatFileSize, formatDateTime } from '@utils/index';
import { useUserActivityLogs, useUserDetailStats } from '@hooks/useUserManager';
import type { User } from '../../types';

const { Text, Title } = Typography;

export interface UserDetailModalProps {
  user: User | null;
  visible: boolean;
  onClose: () => void;
}

// 活动日志表格列定义
const activityColumns: ColumnsType<any> = [
  {
    title: '时间',
    dataIndex: 'created_at',
    key: 'created_at',
    width: 120,
    render: (date: string) => (
      <Text type="secondary">
        {new Date(date).toLocaleString()}
      </Text>
    ),
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 120,
    render: (action: string) => {
      const actionMap: Record<string, { color: string; text: string }> = {
        login: { color: 'success', text: '登录' },
        logout: { color: 'default', text: '登出' },
        upload: { color: 'processing', text: '上传文件' },
        download: { color: 'blue', text: '下载文件' },
        delete: { color: 'error', text: '删除文件' },
        create_folder: { color: 'cyan', text: '创建文件夹' },
      };
      
      const config = actionMap[action] || { color: 'default', text: action };
      return <Tag color={config.color}>{config.text}</Tag>;
    },
  },
  {
    title: '详情',
    dataIndex: 'description',
    key: 'description',
    ellipsis: true,
  },
  {
    title: 'IP地址',
    dataIndex: 'ip_address',
    key: 'ip_address',
    width: 120,
    render: (ip: string) => (
      <Text code style={{ fontSize: 12 }}>
        {ip || '-'}
      </Text>
    ),
  },
];

export const UserDetailModal: React.FC<UserDetailModalProps> = ({
  user,
  visible,
  onClose,
}) => {
  const [activeTab, setActiveTab] = useState('basic');

  // 获取用户活动日志
  const {
    data: activityLogs,
    isLoading: activityLoading,
  } = useUserActivityLogs(user?.id || 0, { page: 1, size: 50 });

  // 获取用户详细统计
  const {
    data: userStats,
    isLoading: statsLoading,
  } = useUserDetailStats(user?.id || 0);

  if (!user) return null;

  // 计算存储使用率
  const storageUsagePercent = user.storage_quota > 0 
    ? Math.round((user.storage_used / user.storage_quota) * 100) 
    : 0;

  // 状态标签
  const statusTag = user.status === 1 
    ? <Tag color="success">正常</Tag> 
    : <Tag color="error">禁用</Tag>;

  // 角色标签
  const roleMap: Record<string, { color: string; text: string }> = {
    admin: { color: 'red', text: '管理员' },
    user: { color: 'blue', text: '普通用户' },
    moderator: { color: 'orange', text: '版主' },
    vip: { color: 'gold', text: 'VIP用户' },
  };
  const roleConfig = roleMap[user.role] || { color: 'default', text: user.role };

  // 标签页内容
  const tabItems = [
    {
      key: 'basic',
      label: (
        <Space>
          <UserOutlined />
          基本信息
        </Space>
      ),
      children: (
        <div>
          <Row gutter={16}>
            <Col span={8}>
              <div style={{ textAlign: 'center', marginBottom: 24 }}>
                <Avatar 
                  size={80} 
                  icon={<UserOutlined />}
                  style={{ 
                    backgroundColor: user.status === 1 ? '#1677ff' : '#8c8c8c',
                    marginBottom: 16,
                  }}
                >
                  {user.display_name?.charAt(0) || user.username.charAt(0).toUpperCase()}
                </Avatar>
                <div>
                  <Title level={4} style={{ margin: 0 }}>
                    {user.display_name || user.username}
                  </Title>
                  <Text type="secondary">@{user.username}</Text>
                </div>
              </div>
            </Col>
            <Col span={16}>
              <Descriptions column={1} size="small">
                <Descriptions.Item label="用户ID">
                  {user.id}
                </Descriptions.Item>
                <Descriptions.Item label="邮箱">
                  <Space>
                    <MailOutlined />
                    {user.email}
                  </Space>
                </Descriptions.Item>
                <Descriptions.Item label="角色">
                  <Tag color={roleConfig.color}>{roleConfig.text}</Tag>
                </Descriptions.Item>
                <Descriptions.Item label="状态">
                  {statusTag}
                </Descriptions.Item>
                <Descriptions.Item label="注册时间">
                  <Space>
                    <CalendarOutlined />
                    {formatDateTime(user.created_at)}
                  </Space>
                </Descriptions.Item>
                <Descriptions.Item label="最后登录">
                  {user.last_login_at ? (
                    <Space>
                      <CalendarOutlined />
                      {formatDateTime(user.last_login_at)}
                    </Space>
                  ) : (
                    <Text type="secondary">从未登录</Text>
                  )}
                </Descriptions.Item>
              </Descriptions>
            </Col>
          </Row>
        </div>
      ),
    },
    {
      key: 'storage',
      label: (
        <Space>
          <DatabaseOutlined />
          存储信息
        </Space>
      ),
      children: (
        <div>
          <Row gutter={16}>
            <Col span={12}>
              <Card title="存储使用情况" size="small">
                <div style={{ marginBottom: 16 }}>
                  <Progress
                    type="circle"
                    percent={storageUsagePercent}
                    size={120}
                    status={storageUsagePercent >= 90 ? 'exception' : 'normal'}
                    format={() => `${storageUsagePercent}%`}
                  />
                </div>
                <Descriptions size="small" column={1}>
                  <Descriptions.Item label="已使用">
                    {formatFileSize(user.storage_used)}
                  </Descriptions.Item>
                  <Descriptions.Item label="总配额">
                    {formatFileSize(user.storage_quota)}
                  </Descriptions.Item>
                  <Descriptions.Item label="剩余空间">
                    {formatFileSize(user.storage_quota - user.storage_used)}
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            </Col>
            <Col span={12}>
              {userStats && (
                <Card title="文件统计" size="small" loading={statsLoading}>
                  <Row gutter={8}>
                    <Col span={12}>
                      <Statistic
                        title="总文件数"
                        value={userStats.total_files}
                        prefix={<FileOutlined />}
                      />
                    </Col>
                    <Col span={12}>
                      <Statistic
                        title="文件夹数"
                        value={userStats.total_folders || 0}
                        prefix={<DatabaseOutlined />}
                      />
                    </Col>
                  </Row>
                </Card>
              )}
            </Col>
          </Row>
        </div>
      ),
    },
    {
      key: 'activity',
      label: (
        <Space>
          <HistoryOutlined />
          活动日志
        </Space>
      ),
      children: (
        <div>
          {activityLogs && activityLogs.items.length > 0 ? (
            <Table
              columns={activityColumns}
              dataSource={activityLogs.items}
              rowKey="id"
              size="small"
              loading={activityLoading}
              pagination={{
                total: activityLogs.total,
                pageSize: 50,
                showSizeChanger: false,
                showQuickJumper: true,
              }}
            />
          ) : (
            <Empty 
              description="暂无活动记录"
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            />
          )}
        </div>
      ),
    },
  ];

  return (
    <Modal
      title={
        <Space>
          <UserOutlined />
          用户详情
        </Space>
      }
      open={visible}
      onCancel={onClose}
      width={900}
      footer={
        <Button type="primary" onClick={onClose}>
          关闭
        </Button>
      }
      destroyOnClose
    >
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={tabItems}
        size="small"
      />
    </Modal>
  );
};