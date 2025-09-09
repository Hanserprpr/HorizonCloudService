import React, { useState, useCallback } from 'react';
import {
  Card,
  message,
  Modal,
  Form,
  Input,
  Button,
  Space,
  Typography,
  Divider,
  Breadcrumb,
} from 'antd';
import {
  UserOutlined,
  HomeOutlined,
  TeamOutlined,
  KeyOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import {
  UserTable,
  UserForm,
  UserDetailModal,
  UserToolbar,
  type UserFilterParams,
} from '@components/users';
import {
  useUsers,
  useCreateUser,
  useUpdateUser,
  useDeleteUser,
  useBatchDeleteUsers,
  useUpdateUserStatus,
  useBatchUpdateUserStatus,
  useBatchUpdateUserQuota,
  useResetPassword,
} from '@hooks/useUserManager';
import type { User, CreateUserRequest, UpdateUserRequest } from '@/types';

const { Title, Text } = Typography;
const { confirm } = Modal;

interface UsersPageState {
  // 搜索和筛选
  searchKeyword: string;
  filterParams: UserFilterParams;
  
  // 分页参数
  currentPage: number;
  pageSize: number;
  
  // 模态框状态
  createModalVisible: boolean;
  editModalVisible: boolean;
  detailModalVisible: boolean;
  resetPasswordModalVisible: boolean;
  
  // 选中的数据
  selectedUsers: User[];
  selectedRowKeys: React.Key[];
  currentUser: User | null;
}

export const UsersPage: React.FC = () => {
  const [resetPasswordForm] = Form.useForm();
  
  // 页面状态
  const [state, setState] = useState<UsersPageState>({
    searchKeyword: '',
    filterParams: {},
    currentPage: 1,
    pageSize: 20,
    createModalVisible: false,
    editModalVisible: false,
    detailModalVisible: false,
    resetPasswordModalVisible: false,
    selectedUsers: [],
    selectedRowKeys: [],
    currentUser: null,
  });

  // API Hooks
  const {
    data: usersData,
    isLoading: usersLoading,
    refetch: refetchUsers,
  } = useUsers({
    page: state.currentPage,
    size: state.pageSize,
    keyword: state.searchKeyword || undefined,
    ...state.filterParams,
  });

  const createUserMutation = useCreateUser();
  const updateUserMutation = useUpdateUser();
  const deleteUserMutation = useDeleteUser();
  const batchDeleteMutation = useBatchDeleteUsers();
  const updateStatusMutation = useUpdateUserStatus();
  const batchUpdateStatusMutation = useBatchUpdateUserStatus();
  const batchUpdateQuotaMutation = useBatchUpdateUserQuota();
  const resetPasswordMutation = useResetPassword();

  // 状态更新辅助函数
  const updateState = useCallback((updates: Partial<UsersPageState>) => {
    setState(prev => ({ ...prev, ...updates }));
  }, []);

  // 搜索处理
  const handleSearch = useCallback((keyword: string) => {
    updateState({
      searchKeyword: keyword,
      currentPage: 1, // 搜索时重置到第一页
    });
  }, [updateState]);

  // 筛选处理
  const handleFilter = useCallback((filters: UserFilterParams) => {
    updateState({
      filterParams: filters,
      currentPage: 1, // 筛选时重置到第一页
    });
  }, [updateState]);

  // 表格变化处理（分页、排序、筛选）
  const handleTableChange = useCallback((pagination: any, filters: any, sorter: any) => {
    const newFilterParams: UserFilterParams = {};
    
    // 处理角色筛选
    if (filters.role && filters.role.length > 0) {
      newFilterParams.role = filters.role[0];
    }
    
    // 处理状态筛选
    if (filters.status && filters.status.length > 0) {
      newFilterParams.status = filters.status[0] === 1 ? 'active' : 'inactive';
    }
    
    // 处理排序
    if (sorter && sorter.field) {
      newFilterParams.sort_by = sorter.field;
      newFilterParams.sort_order = sorter.order === 'ascend' ? 'asc' : 'desc';
    }

    updateState({
      currentPage: pagination?.current || 1,
      pageSize: pagination?.pageSize || 20,
      filterParams: newFilterParams,
    });
  }, [updateState]);

  // 选择变化处理
  const handleSelectionChange = useCallback((selectedRowKeys: React.Key[], selectedRows: User[]) => {
    updateState({
      selectedRowKeys,
      selectedUsers: selectedRows,
    });
  }, [updateState]);

  // 刷新数据
  const handleRefresh = useCallback(() => {
    refetchUsers();
    message.success('数据已刷新');
  }, [refetchUsers]);

  // 创建用户
  const handleCreateUser = useCallback(() => {
    updateState({ createModalVisible: true });
  }, [updateState]);

  const handleCreateSubmit = useCallback(async (data: CreateUserRequest) => {
    try {
      await createUserMutation.mutateAsync(data);
      updateState({ createModalVisible: false });
      message.success('用户创建成功');
      refetchUsers();
    } catch (error: any) {
      message.error(error?.message || '创建用户失败');
      throw error; // 让表单组件处理错误状态
    }
  }, [createUserMutation, updateState, refetchUsers]);

  // 编辑用户
  const handleEditUser = useCallback((user: User) => {
    updateState({
      currentUser: user,
      editModalVisible: true,
    });
  }, [updateState]);

  const handleEditSubmit = useCallback(async (data: UpdateUserRequest) => {
    if (!state.currentUser) return;
    
    try {
      await updateUserMutation.mutateAsync({
        userId: state.currentUser.id,
        data,
      });
      updateState({ editModalVisible: false, currentUser: null });
      message.success('用户信息更新成功');
      refetchUsers();
    } catch (error: any) {
      message.error(error?.message || '更新用户失败');
      throw error;
    }
  }, [state.currentUser, updateUserMutation, updateState, refetchUsers]);

  // 查看用户详情
  const handleViewUser = useCallback((user: User) => {
    updateState({
      currentUser: user,
      detailModalVisible: true,
    });
  }, [updateState]);

  // 删除用户
  const handleDeleteUser = useCallback((user: User) => {
    confirm({
      title: '删除用户确认',
      icon: <ExclamationCircleOutlined />,
      content: (
        <div>
          <p>确定要删除用户 <Text strong>{user.username}</Text> 吗？</p>
          <p style={{ color: '#ff4d4f', fontSize: '12px' }}>
            ⚠️ 删除用户将同时删除其所有文件数据，此操作不可恢复！
          </p>
        </div>
      ),
      okText: '确定删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await deleteUserMutation.mutateAsync(user.id);
          message.success('用户删除成功');
          refetchUsers();
          // 如果当前选中的用户被删除，清除选中状态
          updateState({
            selectedUsers: state.selectedUsers.filter(u => u.id !== user.id),
            selectedRowKeys: state.selectedRowKeys.filter(key => key !== user.id),
          });
        } catch (error: any) {
          message.error(error?.message || '删除用户失败');
        }
      },
    });
  }, [deleteUserMutation, refetchUsers, state.selectedUsers, state.selectedRowKeys, updateState]);

  // 批量删除用户
  const handleBatchDelete = useCallback((userIds: number[]) => {
    const userNames = state.selectedUsers
      .filter(user => userIds.includes(user.id))
      .map(user => user.username);
    
    confirm({
      title: '批量删除用户确认',
      icon: <ExclamationCircleOutlined />,
      content: (
        <div>
          <p>确定要删除以下 {userIds.length} 个用户吗？</p>
          <div style={{ maxHeight: 200, overflow: 'auto', marginTop: 8 }}>
            {userNames.map(name => (
              <div key={name}>• {name}</div>
            ))}
          </div>
          <p style={{ color: '#ff4d4f', fontSize: '12px', marginTop: 12 }}>
            ⚠️ 批量删除用户将同时删除其所有文件数据，此操作不可恢复！
          </p>
        </div>
      ),
      okText: '确定删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await batchDeleteMutation.mutateAsync(userIds);
          message.success(`成功删除 ${userIds.length} 个用户`);
          refetchUsers();
          // 清除选中状态
          updateState({
            selectedUsers: [],
            selectedRowKeys: [],
          });
        } catch (error: any) {
          message.error(error?.message || '批量删除失败');
        }
      },
    });
  }, [state.selectedUsers, batchDeleteMutation, refetchUsers, updateState]);

  // 切换用户状态
  const handleToggleStatus = useCallback(async (user: User) => {
    const newStatus = user.status === 1 ? 'inactive' : 'active';
    const actionText = newStatus === 'active' ? '启用' : '禁用';
    
    try {
      await updateStatusMutation.mutateAsync({
        userId: user.id,
        status: newStatus,
      });
      message.success(`用户${actionText}成功`);
      refetchUsers();
    } catch (error: any) {
      message.error(error?.message || `${actionText}用户失败`);
    }
  }, [updateStatusMutation, refetchUsers]);

  // 批量更新用户状态
  const handleBatchUpdateStatus = useCallback(async (userIds: number[], status: 'active' | 'inactive') => {
    const actionText = status === 'active' ? '启用' : '禁用';
    
    try {
      await batchUpdateStatusMutation.mutateAsync({ userIds, status });
      message.success(`批量${actionText}成功`);
      refetchUsers();
    } catch (error: any) {
      message.error(error?.message || `批量${actionText}失败`);
    }
  }, [batchUpdateStatusMutation, refetchUsers]);

  // 批量更新用户配额
  const handleBatchUpdateQuota = useCallback(async (userIds: number[], quota: number) => {
    try {
      await batchUpdateQuotaMutation.mutateAsync({ userIds, quota });
      message.success(`批量设置配额成功`);
      refetchUsers();
    } catch (error: any) {
      message.error(error?.message || '批量设置配额失败');
    }
  }, [batchUpdateQuotaMutation, refetchUsers]);

  // 重置密码
  const handleResetPassword = useCallback((user: User) => {
    updateState({
      currentUser: user,
      resetPasswordModalVisible: true,
    });
    resetPasswordForm.setFieldsValue({
      username: user.username,
      new_password: '',
    });
  }, [updateState, resetPasswordForm]);

  const handleResetPasswordSubmit = useCallback(async (values: { new_password: string }) => {
    if (!state.currentUser) return;
    
    try {
      await resetPasswordMutation.mutateAsync({
        userId: state.currentUser.id,
        newPassword: values.new_password,
      });
      updateState({ resetPasswordModalVisible: false, currentUser: null });
      resetPasswordForm.resetFields();
      message.success('密码重置成功');
    } catch (error: any) {
      message.error(error?.message || '重置密码失败');
    }
  }, [state.currentUser, resetPasswordMutation, updateState, resetPasswordForm]);

  // 配额管理（这里简化为显示消息，实际可以打开配额管理模态框）
  const handleManageQuota = useCallback((user: User) => {
    // 这里可以实现更详细的配额管理功能
    message.info(`配额管理功能：用户 ${user.username}`);
  }, []);

  // 导出用户数据
  const handleExport = useCallback(() => {
    // 这里可以实现用户数据导出功能
    message.info('用户数据导出功能开发中...');
  }, []);

  return (
    <div className="users-page">
      {/* 页面头部 */}
      <div style={{ marginBottom: 24 }}>
        <Breadcrumb
          items={[
            { href: '/', title: <HomeOutlined />, },
            { href: '/dashboard', title: '控制台' },
            { title: (<><TeamOutlined /><span>用户管理</span></>) },
          ]}
          style={{ marginBottom: 16 }}
        />
        
        <Title level={2} style={{ margin: 0 }}>
          <Space>
            <TeamOutlined />
            用户管理
          </Space>
        </Title>
        <Text type="secondary">
          管理系统用户账户、权限和存储配额
        </Text>
      </div>

      {/* 用户管理卡片 */}
      <Card>
        {/* 工具栏 */}
        <UserToolbar
          selectedUsers={state.selectedUsers}
          onSearch={handleSearch}
          onFilter={handleFilter}
          onRefresh={handleRefresh}
          onCreateUser={handleCreateUser}
          onBatchDelete={handleBatchDelete}
          onBatchUpdateStatus={handleBatchUpdateStatus}
          onBatchUpdateQuota={handleBatchUpdateQuota}
          onExport={handleExport}
          loading={usersLoading}
        />

        <Divider />

        {/* 用户表格 */}
        <UserTable
          users={usersData?.items || []}
          loading={usersLoading}
          selectedRowKeys={state.selectedRowKeys}
          onSelectionChange={handleSelectionChange}
          onView={handleViewUser}
          onEdit={handleEditUser}
          onDelete={handleDeleteUser}
          onResetPassword={handleResetPassword}
          onToggleStatus={handleToggleStatus}
          onManageQuota={handleManageQuota}
          pagination={{
            current: state.currentPage,
            pageSize: state.pageSize,
            total: usersData?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => 
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
            pageSizeOptions: ['10', '20', '50', '100'],
          }}
          onChange={handleTableChange}
        />
      </Card>

      {/* 创建用户模态框 */}
      <Modal
        title={
          <Space>
            <UserOutlined />
            创建新用户
          </Space>
        }
        open={state.createModalVisible}
        onCancel={() => updateState({ createModalVisible: false })}
        width={900}
        footer={null}
        destroyOnClose
      >
        <UserForm
          mode="create"
          loading={createUserMutation.isPending}
          onSubmit={handleCreateSubmit}
          onCancel={() => updateState({ createModalVisible: false })}
        />
      </Modal>

      {/* 编辑用户模态框 */}
      <Modal
        title={
          <Space>
            <UserOutlined />
            编辑用户信息
          </Space>
        }
        open={state.editModalVisible}
        onCancel={() => updateState({ editModalVisible: false, currentUser: null })}
        width={900}
        footer={null}
        destroyOnClose
      >
        <UserForm
          user={state.currentUser || undefined}
          mode="edit"
          loading={updateUserMutation.isPending}
          onSubmit={handleEditSubmit}
          onCancel={() => updateState({ editModalVisible: false, currentUser: null })}
        />
      </Modal>

      {/* 用户详情模态框 */}
      <UserDetailModal
        user={state.currentUser}
        visible={state.detailModalVisible}
        onClose={() => updateState({ detailModalVisible: false, currentUser: null })}
      />

      {/* 重置密码模态框 */}
      <Modal
        title={
          <Space>
            <KeyOutlined />
            重置用户密码
          </Space>
        }
        open={state.resetPasswordModalVisible}
        onCancel={() => {
          updateState({ resetPasswordModalVisible: false, currentUser: null });
          resetPasswordForm.resetFields();
        }}
        onOk={() => resetPasswordForm.submit()}
        confirmLoading={resetPasswordMutation.isPending}
        destroyOnClose
      >
        <Form
          form={resetPasswordForm}
          layout="vertical"
          onFinish={handleResetPasswordSubmit}
        >
          <Form.Item label="用户名">
            <Input disabled />
          </Form.Item>
          
          <Form.Item
            name="new_password"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 6, message: '密码至少6个字符' },
            ]}
          >
            <Input.Password
              placeholder="请输入新密码"
              visibilityToggle
            />
          </Form.Item>
          
          <div style={{ color: '#8c8c8c', fontSize: 12, marginTop: 8 }}>
            重置后用户需要使用新密码登录系统
          </div>
        </Form>
      </Modal>
    </div>
  );
};

export default UsersPage;