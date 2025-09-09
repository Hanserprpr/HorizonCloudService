import React, { useState } from 'react';
import {
  Table,
  Tag,
  Space,
  Button,
  Dropdown,
  Avatar,
  Typography,
  Progress,
  Tooltip,
  Popconfirm,
  Badge,
} from 'antd';
import type { ColumnsType, TableProps } from 'antd/es/table';
import {
  UserOutlined,
  EditOutlined,
  DeleteOutlined,
  MoreOutlined,
  EyeOutlined,
  KeyOutlined,
  StopOutlined,
  PlayCircleOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import { formatFileSize, formatDateTime } from '@utils/index';
import type { User } from '../../types';
import type { MenuProps } from 'antd';

const { Text } = Typography;

export interface UserTableProps {
  users: User[];
  loading?: boolean;
  selectedRowKeys?: React.Key[];
  onSelectionChange?: (selectedRowKeys: React.Key[], selectedRows: User[]) => void;
  onView?: (user: User) => void;
  onEdit?: (user: User) => void;
  onDelete?: (user: User) => void;
  onResetPassword?: (user: User) => void;
  onToggleStatus?: (user: User) => void;
  onManageQuota?: (user: User) => void;
  pagination?: TableProps<User>['pagination'];
  onChange?: TableProps<User>['onChange'];
}

// 状态标签组件
const StatusTag: React.FC<{ status: number }> = ({ status }) => {
  if (status === 1) {
    return <Tag color="success">正常</Tag>;
  }
  return <Tag color="error">禁用</Tag>;
};

// 角色标签组件
const RoleTag: React.FC<{ role: string }> = ({ role }) => {
  const roleConfig = {
    admin: { color: 'red', text: '管理员' },
    user: { color: 'blue', text: '普通用户' },
    moderator: { color: 'orange', text: '版主' },
    vip: { color: 'gold', text: 'VIP用户' },
  };

  const config = roleConfig[role as keyof typeof roleConfig] || { color: 'default', text: role };

  return <Tag color={config.color}>{config.text}</Tag>;
};

// 存储使用进度组件
const StorageProgress: React.FC<{ used: number; total: number }> = ({ used, total }) => {
  const percentage = total > 0 ? Math.round((used / total) * 100) : 0;
  const isNearLimit = percentage >= 80;
  const isOverLimit = percentage >= 100;

  let status: 'success' | 'exception' | 'active' | 'normal' = 'normal';
  if (isOverLimit) {
    status = 'exception';
  } else if (isNearLimit) {
    status = 'active';
  } else {
    status = 'success';
  }

  return (
    <div style={{ minWidth: 120 }}>
      <Progress
        percent={percentage}
        size="small"
        status={status}
        showInfo={false}
      />
      <Text type="secondary" style={{ fontSize: 12 }}>
        {formatFileSize(used)} / {formatFileSize(total)}
      </Text>
    </div>
  );
};

export const UserTable: React.FC<UserTableProps> = ({
  users,
  loading = false,
  selectedRowKeys = [],
  onSelectionChange,
  onView,
  onEdit,
  onDelete,
  onResetPassword,
  onToggleStatus,
  onManageQuota,
  pagination,
  onChange,
}) => {
  // 生成操作菜单
  const getActionMenu = (user: User): MenuProps['items'] => [
    {
      key: 'view',
      icon: <EyeOutlined />,
      label: '查看详情',
      onClick: () => onView?.(user),
    },
    {
      key: 'edit',
      icon: <EditOutlined />,
      label: '编辑用户',
      onClick: () => onEdit?.(user),
    },
    {
      type: 'divider',
    },
    {
      key: 'quota',
      icon: <SettingOutlined />,
      label: '配额管理',
      onClick: () => onManageQuota?.(user),
    },
    {
      key: 'reset-password',
      icon: <KeyOutlined />,
      label: '重置密码',
      onClick: () => onResetPassword?.(user),
    },
    {
      key: 'toggle-status',
      icon: user.status === 1 ? <StopOutlined /> : <PlayCircleOutlined />,
      label: user.status === 1 ? '禁用用户' : '启用用户',
      onClick: () => onToggleStatus?.(user),
    },
    {
      type: 'divider',
    },
    {
      key: 'delete',
      icon: <DeleteOutlined />,
      label: '删除用户',
      danger: true,
      onClick: () => onDelete?.(user),
    },
  ];

  // 表格列定义
  const columns: ColumnsType<User> = [
    {
      title: '用户信息',
      dataIndex: 'username',
      key: 'username',
      width: 200,
      render: (text: string, record: User) => (
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Avatar 
            size={40} 
            icon={<UserOutlined />}
            style={{ 
              backgroundColor: record.status === 1 ? '#1677ff' : '#8c8c8c',
              marginRight: 12,
            }}
          >
            {record.display_name?.charAt(0) || text.charAt(0).toUpperCase()}
          </Avatar>
          <div>
            <div>
              <Text strong>{record.display_name || text}</Text>
              {record.last_login_at && (
                <Badge 
                  status="success" 
                  style={{ marginLeft: 8 }}
                  title="最近活跃"
                />
              )}
            </div>
            <Text type="secondary" style={{ fontSize: 12 }}>
              @{text}
            </Text>
            <br />
            <Text type="secondary" style={{ fontSize: 12 }}>
              {record.email}
            </Text>
          </div>
        </div>
      ),
      sorter: true,
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      width: 100,
      render: (role: string) => <RoleTag role={role} />,
      filters: [
        { text: '管理员', value: 'admin' },
        { text: '普通用户', value: 'user' },
        { text: '版主', value: 'moderator' },
        { text: 'VIP用户', value: 'vip' },
      ],
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: number) => <StatusTag status={status} />,
      filters: [
        { text: '正常', value: 1 },
        { text: '禁用', value: 0 },
      ],
    },
    {
      title: '存储使用',
      key: 'storage',
      width: 180,
      render: (_, record: User) => (
        <StorageProgress 
          used={record.storage_used} 
          total={record.storage_quota} 
        />
      ),
      sorter: true,
    },
    {
      title: '注册时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: (date: string) => (
        <Tooltip title={formatDateTime(date)}>
          <Text type="secondary">
            {new Date(date).toLocaleDateString()}
          </Text>
        </Tooltip>
      ),
      sorter: true,
    },
    {
      title: '最后登录',
      dataIndex: 'last_login_at',
      key: 'last_login_at',
      width: 120,
      render: (date?: string) => {
        if (!date) {
          return <Text type="secondary">从未登录</Text>;
        }
        
        const loginDate = new Date(date);
        const now = new Date();
        const diffHours = Math.floor((now.getTime() - loginDate.getTime()) / (1000 * 60 * 60));
        
        if (diffHours < 1) {
          return <Text type="success">刚刚</Text>;
        } else if (diffHours < 24) {
          return <Text type="success">{diffHours}小时前</Text>;
        } else if (diffHours < 24 * 7) {
          return <Text>{Math.floor(diffHours / 24)}天前</Text>;
        } else {
          return (
            <Tooltip title={formatDateTime(date)}>
              <Text type="secondary">
                {loginDate.toLocaleDateString()}
              </Text>
            </Tooltip>
          );
        }
      },
      sorter: true,
    },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      fixed: 'right',
      render: (_, record: User) => (
        <Space size="small">
          <Button
            type="text"
            icon={<EyeOutlined />}
            size="small"
            onClick={() => onView?.(record)}
            title="查看详情"
          />
          <Button
            type="text"
            icon={<EditOutlined />}
            size="small"
            onClick={() => onEdit?.(record)}
            title="编辑用户"
          />
          <Dropdown
            menu={{ items: getActionMenu(record) }}
            trigger={['click']}
            placement="bottomRight"
          >
            <Button
              type="text"
              icon={<MoreOutlined />}
              size="small"
              title="更多操作"
            />
          </Dropdown>
        </Space>
      ),
    },
  ];

  const rowSelection = onSelectionChange ? {
    selectedRowKeys,
    onChange: onSelectionChange,
    preserveSelectedRowKeys: true,
  } : undefined;

  return (
    <Table<User>
      columns={columns}
      dataSource={users}
      rowKey="id"
      loading={loading}
      rowSelection={rowSelection}
      pagination={pagination}
      onChange={onChange}
      scroll={{ x: 1200 }}
      size="middle"
      showSorterTooltip={false}
    />
  );
};