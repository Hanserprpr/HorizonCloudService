import React from 'react';
import { Card, Statistic, Typography, Skeleton } from 'antd';
import { ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons';

const { Text } = Typography;

interface StatCardProps {
  title: string;
  value: number | string;
  suffix?: string;
  prefix?: React.ReactNode;
  precision?: number;
  loading?: boolean;
  trend?: {
    value: number;
    isPositive: boolean;
    text?: string;
  };
  color?: string;
  icon?: React.ReactNode;
  onClick?: () => void;
}

const StatCard: React.FC<StatCardProps> = ({
  title,
  value,
  suffix,
  prefix,
  precision,
  loading = false,
  trend,
  color = '#1677FF',
  icon,
  onClick,
}) => {
  if (loading) {
    return (
      <Card>
        <Skeleton active paragraph={{ rows: 2 }} />
      </Card>
    );
  }

  return (
    <Card 
      hoverable={!!onClick}
      onClick={onClick}
      styles={{
        body: { padding: '20px 24px' }
      }}
    >
      <div className="stat-card-content">
        <div className="stat-card-main">
          <Statistic
            title={title}
            value={value}
            suffix={suffix}
            prefix={prefix}
            precision={precision}
            valueStyle={{ 
              color,
              fontSize: '24px',
              fontWeight: 600
            }}
          />
          
          {trend && (
            <div className="stat-card-trend">
              <Text 
                type={trend.isPositive ? 'success' : 'danger'}
                style={{ 
                  fontSize: '12px',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '4px'
                }}
              >
                {trend.isPositive ? <ArrowUpOutlined /> : <ArrowDownOutlined />}
                {trend.value}%
                {trend.text && <span style={{ marginLeft: '4px' }}>{trend.text}</span>}
              </Text>
            </div>
          )}
        </div>
        
        {icon && (
          <div className="stat-card-icon">
            {icon}
          </div>
        )}
      </div>

      <style jsx>{`
        .stat-card-content {
          display: flex;
          justify-content: space-between;
          align-items: flex-start;
        }

        .stat-card-main {
          flex: 1;
        }

        .stat-card-trend {
          margin-top: 8px;
        }

        .stat-card-icon {
          font-size: 40px;
          color: ${color};
          opacity: 0.8;
          margin-left: 16px;
        }

        :global(.ant-card:hover) .stat-card-icon {
          transform: scale(1.1);
          transition: transform 0.2s;
        }
      `}</style>
    </Card>
  );
};

export default StatCard;