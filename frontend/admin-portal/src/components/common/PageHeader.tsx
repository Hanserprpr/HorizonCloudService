import React from 'react';
import { Typography, Space, Button, Divider } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

const { Title, Text } = Typography;

interface PageHeaderProps {
  title: string;
  subtitle?: string;
  showBack?: boolean;
  backPath?: string;
  extra?: React.ReactNode;
  children?: React.ReactNode;
}

const PageHeader: React.FC<PageHeaderProps> = ({
  title,
  subtitle,
  showBack = false,
  backPath,
  extra,
  children,
}) => {
  const navigate = useNavigate();

  const handleBack = () => {
    if (backPath) {
      navigate(backPath);
    } else {
      navigate(-1);
    }
  };

  return (
    <div className="page-header">
      <div className="page-header-content">
        <div className="page-header-title">
          <Space align="center">
            {showBack && (
              <Button
                type="text"
                icon={<ArrowLeftOutlined />}
                onClick={handleBack}
                style={{ marginRight: 8 }}
              />
            )}
            <div>
              <Title level={2} style={{ margin: 0 }}>
                {title}
              </Title>
              {subtitle && (
                <Text type="secondary" style={{ fontSize: '14px' }}>
                  {subtitle}
                </Text>
              )}
            </div>
          </Space>
        </div>

        {extra && (
          <div className="page-header-extra">
            {extra}
          </div>
        )}
      </div>

      {children && (
        <>
          <Divider style={{ margin: '16px 0' }} />
          <div className="page-header-content">
            {children}
          </div>
        </>
      )}

      <style jsx>{`
        .page-header {
          background: #fff;
          padding: 16px 24px;
          margin: -24px -24px 24px -24px;
          border-bottom: 1px solid #f0f0f0;
        }

        .page-header-content {
          display: flex;
          align-items: center;
          justify-content: space-between;
        }

        .page-header-title {
          flex: 1;
        }

        .page-header-extra {
          flex-shrink: 0;
        }

        @media (max-width: 768px) {
          .page-header {
            padding: 12px 16px;
            margin: -24px -24px 16px -24px;
          }
          
          .page-header-content {
            flex-direction: column;
            align-items: flex-start;
            gap: 12px;
          }

          .page-header-extra {
            width: 100%;
          }
        }
      `}</style>
    </div>
  );
};

export default PageHeader;