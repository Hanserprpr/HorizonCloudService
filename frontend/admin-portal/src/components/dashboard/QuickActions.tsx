import React from 'react';
import { Card, Row, Col, Button, Typography, Space } from 'antd';
import { 
  CloudUploadOutlined,
  FolderAddOutlined,
  UserAddOutlined,
  SettingOutlined,
  FileSearchOutlined,
  BarChartOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { ROUTES } from '@constants/index';

const { Title, Text } = Typography;

interface QuickActionItem {
  key: string;
  title: string;
  description: string;
  icon: React.ReactNode;
  color: string;
  action: () => void;
}

const QuickActions: React.FC = () => {
  const navigate = useNavigate();

  const quickActions: QuickActionItem[] = [
    {
      key: 'upload',
      title: '上传文件',
      description: '快速上传文件到云端',
      icon: <CloudUploadOutlined />,
      color: '#52c41a',
      action: () => navigate(ROUTES.FILES),
    },
    {
      key: 'new-folder',
      title: '新建文件夹',
      description: '创建新的文件夹',
      icon: <FolderAddOutlined />,
      color: '#1677ff',
      action: () => navigate(ROUTES.FILES),
    },
    {
      key: 'add-user',
      title: '添加用户',
      description: '创建新的用户账户',
      icon: <UserAddOutlined />,
      color: '#722ed1',
      action: () => navigate(ROUTES.USERS),
    },
    {
      key: 'search-files',
      title: '搜索文件',
      description: '快速查找文件内容',
      icon: <FileSearchOutlined />,
      color: '#fa8c16',
      action: () => navigate(ROUTES.FILES),
    },
    {
      key: 'system-stats',
      title: '系统分析',
      description: '查看详细统计报告',
      icon: <BarChartOutlined />,
      color: '#13c2c2',
      action: () => navigate(ROUTES.SETTINGS),
    },
    {
      key: 'settings',
      title: '系统设置',
      description: '配置系统参数',
      icon: <SettingOutlined />,
      color: '#8c8c8c',
      action: () => navigate(ROUTES.SETTINGS),
    },
  ];

  return (
    <Card
      title={
        <Title level={5} style={{ margin: 0 }}>
          快速操作
        </Title>
      }
      bodyStyle={{ padding: '16px 24px' }}
    >
      <Row gutter={[12, 12]}>
        {quickActions.map((action) => (
          <Col xs={12} sm={8} lg={4} key={action.key}>
            <div
              onClick={action.action}
              style={{
                padding: '16px 12px',
                borderRadius: 8,
                border: '1px solid #f0f0f0',
                cursor: 'pointer',
                textAlign: 'center',
                transition: 'all 0.2s ease',
                backgroundColor: '#fff',
                ':hover': {
                  borderColor: action.color,
                  boxShadow: `0 2px 8px ${action.color}20`,
                },
              }}
              onMouseEnter={(e) => {
                const target = e.currentTarget;
                target.style.borderColor = action.color;
                target.style.boxShadow = `0 2px 8px ${action.color}20`;
                target.style.transform = 'translateY(-1px)';
              }}
              onMouseLeave={(e) => {
                const target = e.currentTarget;
                target.style.borderColor = '#f0f0f0';
                target.style.boxShadow = 'none';
                target.style.transform = 'translateY(0)';
              }}
            >
              <Space direction="vertical" size={8} style={{ width: '100%' }}>
                <div
                  style={{
                    width: 40,
                    height: 40,
                    borderRadius: 20,
                    backgroundColor: `${action.color}1a`,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    margin: '0 auto',
                    fontSize: '16px',
                    color: action.color,
                  }}
                >
                  {action.icon}
                </div>
                
                <div>
                  <Text
                    strong
                    style={{
                      fontSize: '12px',
                      display: 'block',
                      lineHeight: 1.2,
                      marginBottom: 2,
                    }}
                  >
                    {action.title}
                  </Text>
                  <Text
                    type="secondary"
                    style={{
                      fontSize: '10px',
                      lineHeight: 1.2,
                    }}
                  >
                    {action.description}
                  </Text>
                </div>
              </Space>
            </div>
          </Col>
        ))}
      </Row>
    </Card>
  );
};

export default QuickActions;