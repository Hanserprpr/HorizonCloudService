import React from 'react';
import {
  Card,
  Progress,
  Typography,
  Space,
  Row,
  Col,
  Statistic,
  Button,
  Tag,
  List,
  Avatar,
} from 'antd';
import {
  CloudUploadOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ClockCircleOutlined,
  FileOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import { formatFileSize } from '@utils/index';
import type { UploadTask, UploadStatus } from '@hooks/useChunkedUpload';

const { Text, Title } = Typography;

interface UploadProgressProps {
  uploadTasks: UploadTask[];
  totalTasks: number;
  completedTasks: number;
  failedTasks: number;
  activeTasks: number;
  onPause: (taskId: string) => void;
  onResume: (taskId: string) => void;
  onCancel: (taskId: string) => void;
  onRetry: (taskId: string) => void;
  style?: React.CSSProperties;
  className?: string;
}

// 状态颜色映射
const getStatusColor = (status: UploadStatus): string => {
  const colorMap = {
    pending: '#faad14',
    uploading: '#1677ff',
    paused: '#faad14',
    completed: '#52c41a',
    failed: '#ff4d4f',
    cancelled: '#8c8c8c',
  };
  return colorMap[status] || colorMap.pending;
};

// 状态文本映射
const getStatusText = (status: UploadStatus): string => {
  const textMap = {
    pending: '等待中',
    uploading: '上传中',
    paused: '已暂停',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消',
  };
  return textMap[status] || textMap.pending;
};

// 格式化上传速度
const formatSpeed = (bytesPerSecond: number): string => {
  if (bytesPerSecond === 0) return '0 B/s';
  
  const units = ['B/s', 'KB/s', 'MB/s', 'GB/s'];
  const k = 1024;
  const i = Math.floor(Math.log(bytesPerSecond) / Math.log(k));
  
  return parseFloat((bytesPerSecond / Math.pow(k, i)).toFixed(1)) + ' ' + units[i];
};

// 格式化剩余时间
const formatEstimatedTime = (seconds: number): string => {
  if (seconds <= 0 || !isFinite(seconds)) return '计算中...';
  
  if (seconds < 60) return `${Math.round(seconds)}秒`;
  if (seconds < 3600) return `${Math.round(seconds / 60)}分钟`;
  return `${Math.round(seconds / 3600)}小时`;
};

export const UploadProgress: React.FC<UploadProgressProps> = ({
  uploadTasks,
  totalTasks,
  completedTasks,
  failedTasks,
  activeTasks,
  onPause,
  onResume,
  onCancel,
  onRetry,
  style,
  className,
}) => {
  // 计算总体进度
  const overallProgress = totalTasks > 0 
    ? Math.round((completedTasks / totalTasks) * 100)
    : 0;

  // 计算总体传输速度
  const totalSpeed = uploadTasks
    .filter(task => task.status === 'uploading')
    .reduce((sum, task) => sum + task.speed, 0);

  // 计算总传输量
  const totalBytes = uploadTasks.reduce((sum, task) => sum + task.totalBytes, 0);
  const uploadedBytes = uploadTasks.reduce((sum, task) => sum + task.uploadedBytes, 0);

  // 渲染任务操作按钮
  const renderTaskActions = (task: UploadTask) => {
    const actions: React.ReactNode[] = [];

    switch (task.status) {
      case 'uploading':
        actions.push(
          <Button
            key="pause"
            type="text"
            size="small"
            icon={<PauseCircleOutlined />}
            onClick={() => onPause(task.id)}
          />
        );
        break;

      case 'paused':
        actions.push(
          <Button
            key="resume"
            type="text"
            size="small"
            icon={<PlayCircleOutlined />}
            onClick={() => onResume(task.id)}
          />
        );
        break;

      case 'failed':
        actions.push(
          <Button
            key="retry"
            type="text"
            size="small"
            icon={<CloudUploadOutlined />}
            onClick={() => onRetry(task.id)}
          />
        );
        break;
    }

    if (task.status !== 'completed') {
      actions.push(
        <Button
          key="cancel"
          type="text"
          size="small"
          icon={<CloseCircleOutlined />}
          onClick={() => onCancel(task.id)}
          danger
        />
      );
    }

    return actions;
  };

  if (totalTasks === 0) {
    return null;
  }

  return (
    <div className={className} style={style}>
      {/* 总体进度卡片 */}
      <Card title="上传进度" style={{ marginBottom: 16 }}>
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={6}>
            <Statistic
              title="总进度"
              value={overallProgress}
              suffix="%"
              prefix={<CloudUploadOutlined />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="已完成"
              value={completedTasks}
              suffix={`/ ${totalTasks}`}
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="上传中"
              value={activeTasks}
              prefix={<ClockCircleOutlined style={{ color: '#1677ff' }} />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="上传速度"
              value={formatSpeed(totalSpeed)}
              prefix={<CloudUploadOutlined />}
            />
          </Col>
        </Row>

        <Progress
          percent={overallProgress}
          strokeColor={{
            '0%': '#108ee9',
            '100%': '#52c41a',
          }}
          style={{ marginBottom: 8 }}
        />

        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
          <Text type="secondary">
            {formatFileSize(uploadedBytes)} / {formatFileSize(totalBytes)}
          </Text>
          {failedTasks > 0 && (
            <Text type="danger">
              <ExclamationCircleOutlined /> {failedTasks} 个文件失败
            </Text>
          )}
        </div>
      </Card>

      {/* 详细任务列表 */}
      <Card title="任务详情">
        <List
          dataSource={uploadTasks}
          renderItem={(task) => (
            <List.Item
              key={task.id}
              actions={renderTaskActions(task)}
            >
              <List.Item.Meta
                avatar={
                  <Avatar 
                    icon={<FileOutlined />} 
                    style={{ backgroundColor: getStatusColor(task.status) }}
                  />
                }
                title={
                  <div>
                    <Text strong style={{ marginRight: 8 }}>
                      {task.file.name}
                    </Text>
                    <Tag color={getStatusColor(task.status)}>
                      {getStatusText(task.status)}
                    </Tag>
                  </div>
                }
                description={
                  <div>
                    <div style={{ marginBottom: 4 }}>
                      <Text type="secondary">
                        {formatFileSize(task.uploadedBytes)} / {formatFileSize(task.totalBytes)}
                        {task.status === 'uploading' && task.speed > 0 && (
                          <span style={{ marginLeft: 8 }}>
                            {formatSpeed(task.speed)}
                          </span>
                        )}
                      </Text>
                    </div>

                    {/* 进度条 */}
                    {(task.status === 'uploading' || task.status === 'paused') && (
                      <Progress
                        percent={task.progress}
                        size="small"
                        strokeColor={getStatusColor(task.status)}
                        showInfo={false}
                        style={{ marginBottom: 4 }}
                      />
                    )}

                    {/* 额外信息 */}
                    <Space size={16} wrap>
                      {task.status === 'uploading' && task.estimatedTime && (
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          剩余时间: {formatEstimatedTime(task.estimatedTime)}
                        </Text>
                      )}

                      {task.status === 'failed' && task.error && (
                        <Text type="danger" style={{ fontSize: 12 }}>
                          错误: {task.error}
                        </Text>
                      )}

                      {task.status === 'completed' && (
                        <Text type="success" style={{ fontSize: 12 }}>
                          <CheckCircleOutlined /> 上传完成
                        </Text>
                      )}
                    </Space>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      </Card>
    </div>
  );
};