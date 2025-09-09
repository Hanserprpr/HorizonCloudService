import React, { useState } from 'react';
import {
  Space,
  Button,
  Input,
  Select,
  Dropdown,
  Typography,
  Popconfirm,
  Modal,
  Form,
  InputNumber,
  message,
} from 'antd';
import type { MenuProps } from 'antd';
import {
  PlusOutlined,
  SearchOutlined,
  FilterOutlined,
  DeleteOutlined,
  DownloadOutlined,
  ReloadOutlined,
  MoreOutlined,
  UserSwitchOutlined,
  SettingOutlined,
  KeyOutlined,
} from '@ant-design/icons';
import { formatFileSize } from '@utils/index';
import type { User } from '../../types';

const { Text } = Typography;
const { Search } = Input;

export interface UserToolbarProps {
  selectedUsers?: User[];
  onSearch?: (keyword: string) => void;
  onFilter?: (filters: UserFilterParams) => void;
  onRefresh?: () => void;
  onCreateUser?: () => void;
  onBatchDelete?: (userIds: number[]) => void;
  onBatchUpdateStatus?: (userIds: number[], status: 'active' | 'inactive') => void;
  onBatchUpdateQuota?: (userIds: number[], quota: number) => void;
  onExport?: () => void;
  loading?: boolean;
}

export interface UserFilterParams {
  role?: string;
  status?: 'active' | 'inactive';
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

// 角色选项
const ROLE_OPTIONS = [
  { label: '全部角色', value: '' },
  { label: '管理员', value: 'admin' },
  { label: '普通用户', value: 'user' },
  { label: '版主', value: 'moderator' },
  { label: 'VIP用户', value: 'vip' },
];

// 状态选项
const STATUS_OPTIONS = [
  { label: '全部状态', value: '' },
  { label: '正常用户', value: 'active' },
  { label: '禁用用户', value: 'inactive' },
];

// 排序选项
const SORT_OPTIONS = [
  { label: '按注册时间', value: 'created_at' },
  { label: '按最后登录', value: 'last_login_at' },
  { label: '按用户名', value: 'username' },
  { label: '按存储使用', value: 'storage_used' },
];

export const UserToolbar: React.FC<UserToolbarProps> = ({
  selectedUsers = [],
  onSearch,
  onFilter,
  onRefresh,
  onCreateUser,
  onBatchDelete,
  onBatchUpdateStatus,
  onBatchUpdateQuota,
  onExport,
  loading = false,
}) => {
  const [searchKeyword, setSearchKeyword] = useState('');
  const [currentFilters, setCurrentFilters] = useState<UserFilterParams>({});
  const [quotaModalVisible, setQuotaModalVisible] = useState(false);
  const [quotaForm] = Form.useForm();

  const hasSelection = selectedUsers.length > 0;
  const selectionCount = selectedUsers.length;

  // 搜索处理
  const handleSearch = (value: string) => {
    setSearchKeyword(value);
    onSearch?.(value);
  };

  // 筛选处理
  const handleFilterChange = (key: keyof UserFilterParams, value: any) => {
    const newFilters = { ...currentFilters, [key]: value || undefined };
    setCurrentFilters(newFilters);
    onFilter?.(newFilters);
  };

  // 批量删除确认
  const handleBatchDelete = () => {
    if (!hasSelection) return;
    
    const userIds = selectedUsers.map(user => user.id);
    onBatchDelete?.(userIds);
  };

  // 批量状态更新
  const handleBatchStatusUpdate = (status: 'active' | 'inactive') => {
    if (!hasSelection) return;
    
    const userIds = selectedUsers.map(user => user.id);
    onBatchUpdateStatus?.(userIds, status);
  };

  // 批量配额更新
  const handleQuotaModalOpen = () => {
    if (!hasSelection) return;
    
    setQuotaModalVisible(true);
    quotaForm.setFieldsValue({
      quota: 10 * 1024 * 1024 * 1024, // 默认10GB
    });
  };

  const handleQuotaSubmit = async (values: { quota: number }) => {
    try {
      const userIds = selectedUsers.map(user => user.id);
      await onBatchUpdateQuota?.(userIds, values.quota);
      setQuotaModalVisible(false);
      quotaForm.resetFields();
      message.success(`已为${selectionCount}个用户更新配额`);
    } catch (error) {
      // 错误处理已在上级组件完成
    }
  };

  // 批量操作菜单
  const batchActionMenu: MenuProps['items'] = [
    {
      key: 'enable',
      label: '批量启用',
      icon: <UserSwitchOutlined />,
      onClick: () => handleBatchStatusUpdate('active'),
      disabled: !hasSelection,
    },
    {
      key: 'disable',
      label: '批量禁用',
      icon: <UserSwitchOutlined />,
      onClick: () => handleBatchStatusUpdate('inactive'),
      disabled: !hasSelection,
    },
    {
      type: 'divider',
    },
    {
      key: 'quota',
      label: '批量设置配额',
      icon: <SettingOutlined />,
      onClick: handleQuotaModalOpen,
      disabled: !hasSelection,
    },
    {
      type: 'divider',
    },
    {
      key: 'delete',
      label: '批量删除',
      icon: <DeleteOutlined />,
      danger: true,
      disabled: !hasSelection,
      onClick: handleBatchDelete,
    },
  ];

  return (
    <div>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '16px 0',
        borderBottom: '1px solid #f0f0f0',
        marginBottom: 16,
      }}>
        {/* 左侧操作区 */}
        <Space size="middle">
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={onCreateUser}
          >
            新建用户
          </Button>

          <Button
            icon={<ReloadOutlined />}
            onClick={onRefresh}
            loading={loading}
          >
            刷新
          </Button>

          <Dropdown
            menu={{ items: batchActionMenu }}
            disabled={!hasSelection}
          >
            <Button icon={<MoreOutlined />}>
              批量操作
              {hasSelection && (
                <Text type="secondary">({selectionCount})</Text>
              )}
            </Button>
          </Dropdown>

          <Button
            icon={<DownloadOutlined />}
            onClick={onExport}
          >
            导出用户
          </Button>
        </Space>

        {/* 右侧搜索和筛选区 */}
        <Space size="middle">
          <Search
            placeholder="搜索用户名或邮箱"
            value={searchKeyword}
            onChange={(e) => setSearchKeyword(e.target.value)}
            onSearch={handleSearch}
            style={{ width: 250 }}
            allowClear
            enterButton={<SearchOutlined />}
          />

          <Select
            placeholder="角色筛选"
            value={currentFilters.role}
            onChange={(value) => handleFilterChange('role', value)}
            style={{ width: 120 }}
            allowClear
          >
            {ROLE_OPTIONS.map(option => (
              <Select.Option key={option.value} value={option.value}>
                {option.label}
              </Select.Option>
            ))}
          </Select>

          <Select
            placeholder="状态筛选"
            value={currentFilters.status}
            onChange={(value) => handleFilterChange('status', value)}
            style={{ width: 120 }}
            allowClear
          >
            {STATUS_OPTIONS.map(option => (
              <Select.Option key={option.value} value={option.value}>
                {option.label}
              </Select.Option>
            ))}
          </Select>

          <Select
            placeholder="排序方式"
            value={currentFilters.sort_by}
            onChange={(value) => handleFilterChange('sort_by', value)}
            style={{ width: 140 }}
            allowClear
          >
            {SORT_OPTIONS.map(option => (
              <Select.Option key={option.value} value={option.value}>
                {option.label}
              </Select.Option>
            ))}
          </Select>
        </Space>
      </div>

      {/* 批量操作提示 */}
      {hasSelection && (
        <div style={{
          padding: '8px 16px',
          background: '#e6f7ff',
          border: '1px solid #91d5ff',
          borderRadius: 6,
          marginBottom: 16,
        }}>
          <Space>
            <Text>
              已选择 <Text strong>{selectionCount}</Text> 个用户
            </Text>
            <Popconfirm
              title={`确定要删除选中的 ${selectionCount} 个用户吗？`}
              description="删除用户将同时删除其所有文件数据，此操作不可恢复。"
              onConfirm={handleBatchDelete}
              okText="确定删除"
              cancelText="取消"
              okType="danger"
            >
              <Button 
                type="text" 
                danger 
                size="small"
                icon={<DeleteOutlined />}
              >
                批量删除
              </Button>
            </Popconfirm>
          </Space>
        </div>
      )}

      {/* 批量配额设置模态框 */}
      <Modal
        title={`批量设置存储配额 (${selectionCount} 个用户)`}
        open={quotaModalVisible}
        onCancel={() => setQuotaModalVisible(false)}
        onOk={() => quotaForm.submit()}
        destroyOnClose
      >
        <Form
          form={quotaForm}
          layout="vertical"
          onFinish={handleQuotaSubmit}
        >
          <Form.Item
            name="quota"
            label="存储配额"
            rules={[
              { required: true, message: '请设置存储配额' },
              { type: 'number', min: 1, message: '配额必须大于0' },
            ]}
          >
            <Select placeholder="请选择存储配额">
              <Select.Option value={1 * 1024 * 1024 * 1024}>1 GB</Select.Option>
              <Select.Option value={5 * 1024 * 1024 * 1024}>5 GB</Select.Option>
              <Select.Option value={10 * 1024 * 1024 * 1024}>10 GB</Select.Option>
              <Select.Option value={50 * 1024 * 1024 * 1024}>50 GB</Select.Option>
              <Select.Option value={100 * 1024 * 1024 * 1024}>100 GB</Select.Option>
              <Select.Option value={500 * 1024 * 1024 * 1024}>500 GB</Select.Option>
              <Select.Option value={1024 * 1024 * 1024 * 1024}>1 TB</Select.Option>
            </Select>
          </Form.Item>
          
          <Text type="secondary">
            将为选中的 {selectionCount} 个用户设置相同的存储配额
          </Text>
        </Form>
      </Modal>
    </div>
  );
};