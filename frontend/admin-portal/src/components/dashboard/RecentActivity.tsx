import React from 'react';
import { Card, List, Avatar, Typography, Space, Divider } from 'antd';
import { 
  FileImageOutlined,
  UserAddOutlined,
  FolderOutlined,
  LoginOutlined,
  DeleteOutlined,
  CloudUploadOutlined,
} from '@ant-design/icons';
import type { RecentActivity } from '@hooks/useDashboard';

const { Text, Title } = Typography;

interface RecentActivityProps {
  activities: RecentActivity[];
  loading?: boolean;
}

const RecentActivityComponent: React.FC<RecentActivityProps> = ({ 
  activities, 
  loading = false 
}) => {
  const getActivityIcon = (type: RecentActivity['type'], color?: string) => {
    const iconProps = {
      style: { color: color || '#1677ff', fontSize: '16px' }
    };

    switch (type) {
      case 'file_upload':
        return <CloudUploadOutlined {...iconProps} />;
      case 'user_register':
        return <UserAddOutlined {...iconProps} />;
      case 'file_delete':
        return <DeleteOutlined {...iconProps} />;
      case 'user_login':
        return <LoginOutlined {...iconProps} />;
      default:
        return <FileImageOutlined {...iconProps} />;
    }
  };

  const getActivityTypeText = (type: RecentActivity['type']) => {
    const typeMap = {
      file_upload: '文件上传',
      user_register: '用户注册',
      file_delete: '文件删除',
      user_login: '用户登录',
    };
    return typeMap[type] || '未知活动';
  };

  return (
    <Card
      title={
        <Space align="center">
          <Title level={5} style={{ margin: 0 }}>
            最近活动
          </Title>
          <Text type="secondary" style={{ fontSize: '12px' }}>
            最新 {activities.length} 条
          </Text>
        </Space>
      }
      extra={
        <Text 
          type="secondary" 
          style={{ fontSize: '12px', cursor: 'pointer' }}
          onClick={() => console.log('查看全部活动')}
        >
          查看全部
        </Text>
      }
      loading={loading}
      bodyStyle={{ padding: '12px 24px' }}
    >
      <List
        size="small"
        dataSource={activities}
        split={false}
        renderItem={(item, index) => (
          <>
            <List.Item
              style={{ 
                padding: '12px 0',
                border: 'none',
              }}
            >
              <List.Item.Meta
                avatar={
                  <Avatar
                    size={32}
                    style={{
                      backgroundColor: `${item.color}1a`,
                      border: `1px solid ${item.color}33`,
                    }}
                    icon={getActivityIcon(item.type, item.color)}
                  />
                }
                title={
                  <div style={{ marginBottom: 2 }}>
                    <Text strong style={{ fontSize: '13px' }}>
                      {item.title}
                    </Text>
                    <Text 
                      type="secondary" 
                      style={{ 
                        fontSize: '11px', 
                        marginLeft: 8,
                        padding: '1px 6px',
                        backgroundColor: '#f5f5f5',
                        borderRadius: 10,
                      }}
                    >
                      {getActivityTypeText(item.type)}
                    </Text>
                  </div>
                }
                description={
                  <Space direction="vertical" size={2} style={{ width: '100%' }}>
                    <Text 
                      type="secondary" 
                      style={{ fontSize: '12px', lineHeight: 1.4 }}
                    >
                      {item.description}
                    </Text>
                    <Space size={8}>
                      <Text style={{ fontSize: '11px', color: '#8c8c8c' }}>
                        {item.user}
                      </Text>
                      <Text style={{ fontSize: '11px', color: '#bfbfbf' }}>
                        •
                      </Text>
                      <Text style={{ fontSize: '11px', color: '#8c8c8c' }}>
                        {item.timestamp}
                      </Text>
                    </Space>
                  </Space>
                }
              />
            </List.Item>
            {index < activities.length - 1 && (
              <Divider style={{ margin: '4px 0' }} />
            )}
          </>
        )}
      />
    </Card>
  );
};

export default RecentActivityComponent;