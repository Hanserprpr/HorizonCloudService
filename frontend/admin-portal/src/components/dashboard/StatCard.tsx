import React from 'react';
import { Card, Statistic, Row, Col } from 'antd';
import { 
  ArrowUpOutlined, 
  ArrowDownOutlined,
  MinusOutlined
} from '@ant-design/icons';
import type { ReactNode } from 'react';

export interface StatCardProps {
  title: string;
  value: string | number;
  prefix?: ReactNode;
  suffix?: ReactNode;
  trend?: {
    value: number;
    isPositive?: boolean;
    suffix?: string;
  };
  icon?: ReactNode;
  color?: string;
  loading?: boolean;
  onClick?: () => void;
}

const StatCard: React.FC<StatCardProps> = ({
  title,
  value,
  prefix,
  suffix,
  trend,
  icon,
  color = '#1677ff',
  loading = false,
  onClick,
}) => {
  const getTrendIcon = () => {
    if (!trend || trend.value === 0) return <MinusOutlined />;
    return trend.isPositive !== false ? <ArrowUpOutlined /> : <ArrowDownOutlined />;
  };

  const getTrendColor = () => {
    if (!trend || trend.value === 0) return '#8c8c8c';
    return trend.isPositive !== false ? '#52c41a' : '#ff4d4f';
  };

  return (
    <Card
      hoverable={!!onClick}
      loading={loading}
      onClick={onClick}
      style={{
        cursor: onClick ? 'pointer' : 'default',
        height: '100%',
      }}
      bodyStyle={{ padding: '20px' }}
    >
      <Row align="middle" justify="space-between">
        <Col flex="1">
          <div style={{ marginBottom: 8 }}>
            <span style={{ 
              color: '#8c8c8c', 
              fontSize: '14px',
              fontWeight: 500 
            }}>
              {title}
            </span>
          </div>
          
          <div style={{ marginBottom: trend ? 8 : 0 }}>
            <Statistic
              value={value}
              prefix={prefix}
              suffix={suffix}
              valueStyle={{
                fontSize: '24px',
                fontWeight: 600,
                color: '#262626',
                lineHeight: 1.2,
              }}
            />
          </div>

          {trend && (
            <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
              <span style={{ 
                color: getTrendColor(),
                fontSize: '12px',
                display: 'flex',
                alignItems: 'center',
                gap: 2,
              }}>
                {getTrendIcon()}
                {Math.abs(trend.value)}%{trend.suffix || ''}
              </span>
              <span style={{ 
                color: '#8c8c8c', 
                fontSize: '12px' 
              }}>
                较昨天
              </span>
            </div>
          )}
        </Col>

        {icon && (
          <Col>
            <div
              style={{
                width: 48,
                height: 48,
                borderRadius: 8,
                backgroundColor: `${color}1a`,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '20px',
                color: color,
              }}
            >
              {icon}
            </div>
          </Col>
        )}
      </Row>
    </Card>
  );
};

export default StatCard;