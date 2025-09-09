import React, { useState } from 'react';
import {
  Typography,
  Tabs,
  Space,
  Breadcrumb,
  Alert,
  Spin,
} from 'antd';
import {
  SettingOutlined,
  DatabaseOutlined,
  SecurityScanOutlined,
  InfoCircleOutlined,
  HomeOutlined,
} from '@ant-design/icons';
import {
  BasicSettings,
  StorageSettings,
  SecuritySettings,
  SystemInfo,
} from '@components/settings';
import {
  useSystemSettings,
  useUpdateSettings,
  useSystemInfo,
  useHealthCheck,
  useStorageConfig,
  useUpdateStorageConfig,
  useTestStorageConnection,
  useRestartService,
  useClearCache,
  useDownloadLogs,
} from '@hooks/useSettingsMock';

const { Title, Text } = Typography;

export const SettingsPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState('basic');

  // API Hooks
  const {
    data: systemSettings,
    isLoading: settingsLoading,
    error: settingsError,
  } = useSystemSettings();

  const {
    data: systemInfo,
    isLoading: infoLoading,
    refetch: refetchSystemInfo,
  } = useSystemInfo();

  const {
    data: healthCheck,
    isLoading: healthLoading,
  } = useHealthCheck();

  const {
    data: storageConfig,
    isLoading: storageConfigLoading,
  } = useStorageConfig();

  const updateSettingsMutation = useUpdateSettings();
  const updateStorageConfigMutation = useUpdateStorageConfig();
  const testStorageConnectionMutation = useTestStorageConnection();
  const restartServiceMutation = useRestartService();
  const clearCacheMutation = useClearCache();
  const downloadLogsMutation = useDownloadLogs();

  // 处理基本设置保存
  const handleSaveBasicSettings = async (values: any) => {
    await updateSettingsMutation.mutateAsync(values);
  };

  // 处理安全设置保存
  const handleSaveSecuritySettings = async (values: any) => {
    await updateSettingsMutation.mutateAsync(values);
  };

  // 处理存储设置保存
  const handleSaveStorageSettings = async (values: any) => {
    await updateSettingsMutation.mutateAsync(values);
  };

  // 处理存储配置保存
  const handleSaveStorageConfig = async (config: any) => {
    await updateStorageConfigMutation.mutateAsync(config);
  };

  // 处理存储连接测试
  const handleTestStorageConnection = async (config: any) => {
    await testStorageConnectionMutation.mutateAsync(config);
  };

  // 处理服务重启
  const handleRestartService = async (serviceName: string) => {
    await restartServiceMutation.mutateAsync(serviceName);
  };

  // 处理缓存清理
  const handleClearCache = async (cacheType?: string) => {
    await clearCacheMutation.mutateAsync(cacheType);
  };

  // 处理日志下载
  const handleDownloadLogs = async (service?: string) => {
    await downloadLogsMutation.mutateAsync({ service });
  };

  // 刷新系统信息
  const handleRefreshSystemInfo = () => {
    refetchSystemInfo();
  };

  // 标签页配置
  const tabItems = [
    {
      key: 'basic',
      label: (
        <Space>
          <SettingOutlined />
          基本设置
        </Space>
      ),
      children: (
        <BasicSettings
          settings={systemSettings}
          loading={updateSettingsMutation.isPending}
          onSave={handleSaveBasicSettings}
        />
      ),
    },
    {
      key: 'storage',
      label: (
        <Space>
          <DatabaseOutlined />
          存储设置
        </Space>
      ),
      children: (
        <StorageSettings
          settings={systemSettings}
          storageConfig={storageConfig}
          loading={updateSettingsMutation.isPending || updateStorageConfigMutation.isPending}
          onSaveSettings={handleSaveStorageSettings}
          onSaveStorageConfig={handleSaveStorageConfig}
          onTestConnection={handleTestStorageConnection}
        />
      ),
    },
    {
      key: 'security',
      label: (
        <Space>
          <SecurityScanOutlined />
          安全设置
        </Space>
      ),
      children: (
        <SecuritySettings
          settings={systemSettings}
          loading={updateSettingsMutation.isPending}
          onSave={handleSaveSecuritySettings}
        />
      ),
    },
    {
      key: 'system',
      label: (
        <Space>
          <InfoCircleOutlined />
          系统信息
        </Space>
      ),
      children: (
        <SystemInfo
          systemInfo={systemInfo}
          healthCheck={healthCheck}
          loading={infoLoading || healthLoading}
          onRefresh={handleRefreshSystemInfo}
          onRestartService={handleRestartService}
          onClearCache={handleClearCache}
          onDownloadLogs={handleDownloadLogs}
        />
      ),
    },
  ];

  // 错误处理
  if (settingsError) {
    return (
      <div>
        <Breadcrumb
          items={[
            { href: '/', title: <HomeOutlined /> },
            { href: '/dashboard', title: '控制台' },
            { title: (<><SettingOutlined /><span>系统设置</span></>) },
          ]}
          style={{ marginBottom: 16 }}
        />
        
        <Alert
          message="加载设置失败"
          description="无法加载系统设置，请检查网络连接或联系管理员。"
          type="error"
          showIcon
          action={
            <Space>
              <button onClick={() => window.location.reload()}>重试</button>
            </Space>
          }
        />
      </div>
    );
  }

  return (
    <div className="settings-page" style={{ padding: '24px' }}>
      {/* 页面头部 */}
      <div style={{ marginBottom: 24 }}>
        <Breadcrumb
          items={[
            { href: '/', title: <HomeOutlined /> },
            { href: '/dashboard', title: '控制台' },
            { title: (<><SettingOutlined /><span>系统设置</span></>) },
          ]}
          style={{ marginBottom: 16 }}
        />
        
        <Title level={2} style={{ margin: 0 }}>
          <Space>
            <SettingOutlined />
            系统设置
          </Space>
        </Title>
        <Text 
          type="secondary"
          style={{ 
            display: 'block',
            marginTop: 8
          }}
        >
          管理系统配置、存储设置、安全参数和查看系统信息
        </Text>
      </div>

      {/* 系统健康状态提醒 */}
      {healthCheck && healthCheck.status === 'unhealthy' && (
        <Alert
          message="系统健康状态异常"
          description="检测到部分服务运行异常，建议检查系统信息页面。"
          type="warning"
          showIcon
          closable
          style={{ marginBottom: 24 }}
        />
      )}

      {/* 设置标签页 */}
      <Spin spinning={settingsLoading || storageConfigLoading} tip="正在加载设置...">
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={tabItems}
          style={{ 
            minHeight: 600,
            marginTop: 16,
          }}
          tabBarStyle={{
            marginBottom: 24,
          }}
        />
      </Spin>
    </div>
  );
};

export default SettingsPage;