import React from 'react';
import {
  Card,
  Descriptions,
  Progress,
  Tag,
  Button,
  Space,
  Row,
  Col,
  Statistic,
  Typography,
  Alert,
  Divider,
  Modal,
  List,
  Badge,
} from 'antd';
import {
  InfoCircleOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  CloseCircleOutlined,
  WarningOutlined,
  PoweroffOutlined,
  ClearOutlined,
  DownloadOutlined,
} from '@ant-design/icons';
import { formatFileSize, formatDuration } from '@utils/index';
import type { SystemInfo } from '../../types';

const { Title, Text } = Typography;
const { confirm } = Modal;

export interface SystemInfoProps {
  systemInfo?: SystemInfo;
  healthCheck?: {
    status: 'healthy' | 'unhealthy';
    checks: Record<string, boolean>;
    timestamp: string;
  };
  loading?: boolean;
  onRefresh: () => void;
  onRestartService: (serviceName: string) => Promise<void>;
  onClearCache: (cacheType?: string) => Promise<void>;
  onDownloadLogs: (service?: string) => Promise<void>;
}

const SERVICE_NAMES: Record<string, string> = {
  user_service: '用户服务',
  file_service: '文件服务',
  ai_service: 'AI服务',
  search_service: '搜索服务',
};

const CACHE_TYPES = [
  { value: 'all', label: '所有缓存', description: '清理所有类型的缓存' },
  { value: 'thumbnails', label: '缩略图缓存', description: '清理图片缩略图缓存' },
  { value: 'sessions', label: '会话缓存', description: '清理用户会话缓存' },
  { value: 'temp', label: '临时文件', description: '清理临时文件缓存' },
];

export const SystemInfo: React.FC<SystemInfoProps> = ({
  systemInfo,
  healthCheck,
  loading,
  onRefresh,
  onRestartService,
  onClearCache,
  onDownloadLogs,
}) => {
  const handleRestartService = (serviceName: string) => {
    confirm({
      title: `重启 ${SERVICE_NAMES[serviceName] || serviceName}`,
      icon: <ExclamationCircleOutlined />,
      content: (
        <div>
          <p>确定要重启此服务吗？重启期间相关功能可能不可用。</p>
          <Alert
            message="注意"
            description="重启服务可能会中断当前的用户操作，请谨慎操作。"
            type="warning"
            showIcon
          />
        </div>
      ),
      okText: '确认重启',
      okType: 'danger',
      cancelText: '取消',
      onOk: () => onRestartService(serviceName),
    });
  };

  const handleClearCache = (cacheType?: string) => {
    const cacheInfo = CACHE_TYPES.find(t => t.value === cacheType);
    confirm({
      title: `清理${cacheInfo?.label || '缓存'}`,
      icon: <WarningOutlined />,
      content: `确定要清理${cacheInfo?.label || '缓存'}吗？这可能会暂时影响系统性能。`,
      okText: '确认清理',
      cancelText: '取消',
      onOk: () => onClearCache(cacheType),
    });
  };

  const renderServiceStatus = (status: boolean) => {
    if (status) {
      return <Tag color="success" icon={<CheckCircleOutlined />}>正常运行</Tag>;
    } else {
      return <Tag color="error" icon={<CloseCircleOutlined />}>服务异常</Tag>;
    }
  };

  const renderHealthStatus = () => {
    if (!healthCheck) return null;

    const isHealthy = healthCheck.status === 'healthy';
    const failedChecks = Object.entries(healthCheck.checks || {}).filter(([_, status]) => !status);

    return (
      <Alert
        message={
          <Space>
            <Text strong>系统健康状态</Text>
            <Tag color={isHealthy ? 'success' : 'error'} icon={isHealthy ? <CheckCircleOutlined /> : <ExclamationCircleOutlined />}>
              {isHealthy ? '健康' : '异常'}
            </Tag>
            <Text type="secondary">
              最后检查: {new Date(healthCheck.timestamp).toLocaleString()}
            </Text>
          </Space>
        }
        description={
          failedChecks.length > 0 && (
            <div style={{ marginTop: 8 }}>
              <Text type="secondary">异常项目:</Text>
              <List
                size="small"
                dataSource={failedChecks}
                renderItem={([name]) => (
                  <List.Item>
                    <Text type="danger">• {name}</Text>
                  </List.Item>
                )}
                style={{ marginTop: 4 }}
              />
            </div>
          )
        }
        type={isHealthy ? 'success' : 'error'}
        showIcon
        style={{ marginBottom: 24 }}
      />
    );
  };

  if (!systemInfo) {
    return (
      <Card loading={loading}>
        <Text type="secondary">正在加载系统信息...</Text>
      </Card>
    );
  }

  const memoryUsagePercent = systemInfo.memory_usage 
    ? Math.round((systemInfo.memory_usage.used / systemInfo.memory_usage.total) * 100)
    : 0;

  const storageUsagePercent = systemInfo.storage_info
    ? Math.round((systemInfo.storage_info.used_space / systemInfo.storage_info.total_space) * 100)
    : 0;

  return (
    <div>
      {renderHealthStatus()}

      <Row gutter={[16, 16]}>
        {/* 系统基本信息 */}
        <Col span={12}>
          <Card
            title={
              <Space>
                <InfoCircleOutlined />
                系统信息
              </Space>
            }
            extra={
              <Button 
                icon={<ReloadOutlined />} 
                size="small"
                loading={loading}
                onClick={onRefresh}
              >
                刷新
              </Button>
            }
          >
            <Descriptions column={1} size="small">
              <Descriptions.Item label="系统版本">
                <Tag color="blue">{systemInfo.version}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="构建时间">
                {new Date(systemInfo.build_time).toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="Go版本">
                {systemInfo.go_version}
              </Descriptions.Item>
              <Descriptions.Item label="操作系统">
                {systemInfo.os} / {systemInfo.arch}
              </Descriptions.Item>
              <Descriptions.Item label="运行时间">
                {formatDuration(systemInfo.uptime * 1000)}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>

        {/* 资源使用情况 */}
        <Col span={12}>
          <Card
            title={
              <Space>
                <WarningOutlined />
                资源使用
              </Space>
            }
          >
            <Row gutter={16}>
              <Col span={12}>
                <Statistic 
                  title="内存使用率" 
                  value={memoryUsagePercent} 
                  suffix="%" 
                />
                <Progress 
                  percent={memoryUsagePercent}
                  size="small"
                  status={memoryUsagePercent > 80 ? 'exception' : 'normal'}
                />
                <Text type="secondary" style={{ fontSize: 12 }}>
                  {formatFileSize(systemInfo.memory_usage.used)} / {formatFileSize(systemInfo.memory_usage.total)}
                </Text>
              </Col>
              <Col span={12}>
                <Statistic 
                  title="存储使用率" 
                  value={storageUsagePercent} 
                  suffix="%" 
                />
                <Progress 
                  percent={storageUsagePercent}
                  size="small"
                  status={storageUsagePercent > 90 ? 'exception' : 'normal'}
                />
                <Text type="secondary" style={{ fontSize: 12 }}>
                  {formatFileSize(systemInfo.storage_info.used_space)} / {formatFileSize(systemInfo.storage_info.total_space)}
                </Text>
              </Col>
            </Row>
          </Card>
        </Col>

        {/* 数据库信息 */}
        <Col span={12}>
          <Card title="数据库信息">
            <Descriptions column={1} size="small">
              <Descriptions.Item label="数据库类型">
                {systemInfo.database_info.type}
              </Descriptions.Item>
              <Descriptions.Item label="版本">
                {systemInfo.database_info.version}
              </Descriptions.Item>
              <Descriptions.Item label="数据库大小">
                {formatFileSize(systemInfo.database_info.size)}
              </Descriptions.Item>
              <Descriptions.Item label="连接数">
                {systemInfo.database_info.connections}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>

        {/* 服务状态 */}
        <Col span={12}>
          <Card
            title="服务状态"
            extra={
              <Space>
                <Button 
                  icon={<ClearOutlined />} 
                  size="small"
                  onClick={() => handleClearCache()}
                >
                  清理缓存
                </Button>
                <Button 
                  icon={<DownloadOutlined />} 
                  size="small"
                  onClick={() => onDownloadLogs()}
                >
                  下载日志
                </Button>
              </Space>
            }
          >
            <Space direction="vertical" style={{ width: '100%' }}>
              {Object.entries(systemInfo.services_status).map(([serviceName, status]) => (
                <div key={serviceName} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Space>
                    <Text>{SERVICE_NAMES[serviceName] || serviceName}</Text>
                    {renderServiceStatus(status)}
                  </Space>
                  <Button
                    size="small"
                    icon={<PoweroffOutlined />}
                    onClick={() => handleRestartService(serviceName)}
                    disabled={status === false}
                    type="text"
                  >
                    重启
                  </Button>
                </div>
              ))}
            </Space>
          </Card>
        </Col>

        {/* 文件统计 */}
        <Col span={24}>
          <Card title="存储统计">
            <Row gutter={16}>
              <Col span={6}>
                <Statistic
                  title="总文件数"
                  value={systemInfo.storage_info.files_count}
                  prefix={<InfoCircleOutlined />}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title="已用存储"
                  value={formatFileSize(systemInfo.storage_info.used_space)}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title="可用存储"
                  value={formatFileSize(systemInfo.storage_info.free_space)}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title="总存储容量"
                  value={formatFileSize(systemInfo.storage_info.total_space)}
                />
              </Col>
            </Row>
          </Card>
        </Col>

        {/* 缓存管理 */}
        <Col span={24}>
          <Card title="缓存管理">
            <Row gutter={16}>
              {CACHE_TYPES.map(cacheType => (
                <Col span={6} key={cacheType.value}>
                  <Card 
                    size="small" 
                    hoverable
                    onClick={() => handleClearCache(cacheType.value)}
                    style={{ textAlign: 'center', cursor: 'pointer' }}
                  >
                    <div>
                      <ClearOutlined style={{ fontSize: 24, color: '#1677ff', marginBottom: 8 }} />
                      <div style={{ fontWeight: 'bold' }}>{cacheType.label}</div>
                      <div style={{ fontSize: 12, color: '#666', marginTop: 4 }}>
                        {cacheType.description}
                      </div>
                    </div>
                  </Card>
                </Col>
              ))}
            </Row>
          </Card>
        </Col>
      </Row>
    </div>
  );
};