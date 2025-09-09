import React from 'react';
import { Card, Progress, Row, Col, Typography, Space, Divider } from 'antd';
import { 
  CloudOutlined, 
  DatabaseOutlined,
  LineChartOutlined,
} from '@ant-design/icons';

const { Title, Text } = Typography;

interface StorageData {
  used: string;
  total: string;
  usagePercent: number;
  trend: number;
}

interface StorageTrendData {
  date: string;
  used: number;
  uploaded: number;
}

interface StorageChartProps {
  data: StorageData;
  trendData: StorageTrendData[];
  loading?: boolean;
}

const StorageChart: React.FC<StorageChartProps> = ({ 
  data, 
  trendData, 
  loading = false 
}) => {
  const getUsageStatus = (percent: number) => {
    if (percent >= 90) return { status: 'exception' as const, color: '#ff4d4f' };
    if (percent >= 70) return { status: 'active' as const, color: '#fa8c16' };
    return { status: 'normal' as const, color: '#52c41a' };
  };

  const usageStatus = getUsageStatus(data.usagePercent);

  // 简化的趋势显示 - 最近7天的上传量
  const recentUploads = trendData.slice(-7);
  const totalUploadedThisWeek = recentUploads.reduce((sum, item) => sum + item.uploaded, 0);
  const avgUploadPerDay = totalUploadedThisWeek / recentUploads.length;

  return (
    <Card
      title={
        <Space align="center">
          <DatabaseOutlined style={{ color: '#1677ff' }} />
          <Title level={5} style={{ margin: 0 }}>
            存储使用情况
          </Title>
        </Space>
      }
      loading={loading}
      bodyStyle={{ padding: '20px 24px' }}
    >
      <Space direction="vertical" size={20} style={{ width: '100%' }}>
        {/* 主要存储指标 */}
        <Row gutter={16} align="middle">
          <Col span={16}>
            <Space direction="vertical" size={8} style={{ width: '100%' }}>
              <div style={{ marginBottom: 8 }}>
                <Text style={{ fontSize: '16px', fontWeight: 600 }}>
                  {data.used}
                </Text>
                <Text type="secondary" style={{ marginLeft: 8 }}>
                  / {data.total}
                </Text>
              </div>

              <Progress
                percent={data.usagePercent}
                status={usageStatus.status}
                strokeColor={usageStatus.color}
                trailColor="#f5f5f5"
                strokeWidth={8}
                showInfo={false}
                style={{ margin: 0 }}
              />

              <Row justify="space-between" align="middle">
                <Col>
                  <Text style={{ fontSize: '12px', color: '#8c8c8c' }}>
                    使用率: {data.usagePercent.toFixed(1)}%
                  </Text>
                </Col>
                <Col>
                  <Text 
                    style={{ 
                      fontSize: '12px', 
                      color: data.trend > 0 ? '#52c41a' : '#ff4d4f' 
                    }}
                  >
                    {data.trend > 0 ? '↑' : '↓'} {Math.abs(data.trend)}%
                  </Text>
                </Col>
              </Row>
            </Space>
          </Col>

          <Col span={8} style={{ textAlign: 'center' }}>
            <div
              style={{
                width: 60,
                height: 60,
                borderRadius: 30,
                backgroundColor: `${usageStatus.color}1a`,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                margin: '0 auto',
              }}
            >
              <CloudOutlined 
                style={{ 
                  fontSize: '24px', 
                  color: usageStatus.color 
                }} 
              />
            </div>
            <Text 
              style={{ 
                fontSize: '10px', 
                color: '#8c8c8c',
                marginTop: 4,
                display: 'block',
              }}
            >
              云存储
            </Text>
          </Col>
        </Row>

        <Divider style={{ margin: '8px 0' }} />

        {/* 存储分析 */}
        <Row gutter={16}>
          <Col span={12}>
            <Space direction="vertical" size={2}>
              <Text style={{ fontSize: '12px', color: '#8c8c8c' }}>
                本周上传
              </Text>
              <Text style={{ fontSize: '14px', fontWeight: 600 }}>
                {totalUploadedThisWeek.toFixed(1)} GB
              </Text>
              <Text style={{ fontSize: '10px', color: '#bfbfbf' }}>
                日均 {avgUploadPerDay.toFixed(1)} GB
              </Text>
            </Space>
          </Col>

          <Col span={12}>
            <Space direction="vertical" size={2}>
              <Text style={{ fontSize: '12px', color: '#8c8c8c' }}>
                剩余空间
              </Text>
              <Text style={{ fontSize: '14px', fontWeight: 600 }}>
                {(parseFloat(data.total) - parseFloat(data.used)).toFixed(1)} GB
              </Text>
              <Text style={{ fontSize: '10px', color: '#bfbfbf' }}>
                约可用 {Math.floor((parseFloat(data.total) - parseFloat(data.used)) / avgUploadPerDay)} 天
              </Text>
            </Space>
          </Col>
        </Row>

        {/* 简化的趋势图表示 */}
        <div>
          <Space align="center" style={{ marginBottom: 8 }}>
            <LineChartOutlined style={{ fontSize: '12px', color: '#1677ff' }} />
            <Text style={{ fontSize: '12px', color: '#8c8c8c' }}>
              最近7天趋势
            </Text>
          </Space>
          
          <div style={{ display: 'flex', gap: 2, height: 20, alignItems: 'end' }}>
            {recentUploads.map((item, index) => (
              <div
                key={index}
                style={{
                  flex: 1,
                  backgroundColor: '#1677ff',
                  opacity: 0.6 + (item.uploaded / Math.max(...recentUploads.map(d => d.uploaded))) * 0.4,
                  height: `${(item.uploaded / Math.max(...recentUploads.map(d => d.uploaded))) * 16 + 4}px`,
                  minHeight: '2px',
                  borderRadius: '1px',
                }}
                title={`${item.date}: ${item.uploaded}GB`}
              />
            ))}
          </div>
        </div>
      </Space>
    </Card>
  );
};

export default StorageChart;