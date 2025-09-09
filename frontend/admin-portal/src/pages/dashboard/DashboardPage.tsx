import React from 'react';
import { Row, Col, Typography, Space } from 'antd';
import { useErrorHandler } from '@hooks/useErrorHandler';
import { LoadingSpinner, DataLoadingWrapper } from '@components/common/LoadingSpinner';
import { 
  UserOutlined, 
  FileOutlined, 
  CloudUploadOutlined, 
  DashboardOutlined,
  TeamOutlined,
  HddOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

// Components
import StatCard from '@components/dashboard/StatCard';
import RecentActivity from '@components/dashboard/RecentActivity';
import QuickActions from '@components/dashboard/QuickActions';
import StorageChart from '@components/dashboard/StorageChart';

// Hooks & Services
import { 
  useDashboardStats, 
  useRecentActivities, 
  useStorageTrend 
} from '@hooks/useDashboard';
import { ROUTES } from '@constants/index';

const { Title } = Typography;

const DashboardPage: React.FC = () => {
  const navigate = useNavigate();
  const { handleError, showError } = useErrorHandler();
  
  // 数据获取
  const { 
    data: stats, 
    isLoading: statsLoading, 
    error: statsError 
  } = useDashboardStats();
  
  const { 
    data: activities = [], 
    isLoading: activitiesLoading,
    error: activitiesError
  } = useRecentActivities();
  
  const { 
    data: storageTrend = [], 
    isLoading: trendLoading,
    error: trendError
  } = useStorageTrend();

  // 错误处理
  React.useEffect(() => {
    if (statsError) {
      handleError(statsError, {
        showNotification: true,
        notificationDuration: 8,
      });
    }
  }, [statsError, handleError]);

  React.useEffect(() => {
    if (activitiesError) {
      handleError(activitiesError, { showMessage: true });
    }
  }, [activitiesError, handleError]);

  React.useEffect(() => {
    if (trendError) {
      handleError(trendError, { showMessage: true });
    }
  }, [trendError, handleError]);

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0, marginBottom: 8 }}>
          仪表盘
        </Title>
        <Typography.Text type="secondary">
          欢迎回来！这里是您的系统概览
        </Typography.Text>
      </div>

      <Space direction="vertical" size={24} style={{ width: '100%' }}>
        {/* 统计卡片区域 */}
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} lg={6}>
            <StatCard
              title="总用户数"
              value={stats?.users.total || 0}
              prefix={<UserOutlined />}
              icon={<TeamOutlined />}
              color="#1677ff"
              trend={{
                value: stats?.users.trend || 0,
                isPositive: (stats?.users.trend || 0) > 0,
              }}
              loading={statsLoading}
              onClick={() => navigate(ROUTES.USERS)}
            />
          </Col>

          <Col xs={24} sm={12} lg={6}>
            <StatCard
              title="总文件数"
              value={stats?.files.total || 0}
              prefix={<FileOutlined />}
              icon={<HddOutlined />}
              color="#52c41a"
              trend={{
                value: stats?.files.trend || 0,
                isPositive: (stats?.files.trend || 0) > 0,
              }}
              loading={statsLoading}
              onClick={() => navigate(ROUTES.FILES)}
            />
          </Col>

          <Col xs={24} sm={12} lg={6}>
            <StatCard
              title="今日上传"
              value={stats?.files.uploadedToday || 0}
              suffix="个文件"
              icon={<CloudUploadOutlined />}
              color="#722ed1"
              trend={{
                value: 15.6,
                isPositive: true,
                suffix: ` (${stats?.files.sizeToday || '0 GB'})`
              }}
              loading={statsLoading}
              onClick={() => navigate(ROUTES.FILES)}
            />
          </Col>

          <Col xs={24} sm={12} lg={6}>
            <StatCard
              title="存储使用率"
              value={stats?.storage.usagePercent || 0}
              suffix="%"
              icon={<DashboardOutlined />}
              color={
                (stats?.storage.usagePercent || 0) > 80 
                  ? '#ff4d4f' 
                  : (stats?.storage.usagePercent || 0) > 60 
                    ? '#fa8c16' 
                    : '#13c2c2'
              }
              trend={{
                value: stats?.storage.trend || 0,
                isPositive: false, // 存储使用率增长显示为红色
              }}
              loading={statsLoading}
            />
          </Col>
        </Row>

        {/* 主要内容区域 */}
        <Row gutter={[16, 16]}>
          {/* 左侧内容 */}
          <Col xs={24} lg={16}>
            <Space direction="vertical" size={16} style={{ width: '100%' }}>
              {/* 快速操作 */}
              <QuickActions />
              
              {/* 存储图表 */}
              {stats && (
                <StorageChart
                  data={stats.storage}
                  trendData={storageTrend}
                  loading={trendLoading}
                />
              )}
            </Space>
          </Col>

          {/* 右侧内容 */}
          <Col xs={24} lg={8}>
            <DataLoadingWrapper
              loading={activitiesLoading}
              error={activitiesError}
              empty={!activities || activities.length === 0}
              emptyText="暂无最近活动"
            >
              <RecentActivity
                activities={activities}
                loading={false} // 已由wrapper处理
              />
            </DataLoadingWrapper>
          </Col>
        </Row>
      </Space>
    </div>
  );
};

export default DashboardPage;