import React from 'react';
import {
  Drawer,
  List,
  Progress,
  Button,
  Typography,
  Space,
  Tag,
  Tooltip,
  Empty,
  Divider,
  Badge,
} from 'antd';
import {
  CloudUploadOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  CloseCircleOutlined,
  ReloadOutlined,
  DeleteOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons';
import { formatFileSize, formatRelativeTime } from '@utils/index';
import type { UploadTask, UploadStatus } from '@hooks/useChunkedUpload';

const { Text, Title } = Typography;

interface UploadQueueProps {
  visible: boolean;
  onClose: () => void;
  uploadTasks: UploadTask[];
  onPause: (taskId: string) => void;
  onCancel: (taskId: string) => void;
  onRetry: (taskId: string) => void;
  onClearCompleted: () => void;
  onClearAll: () => void;
  totalTasks: number;
  completedTasks: number;
  failedTasks: number;
  activeTasks: number;
}

// 状态图标映射
const StatusIcon: React.FC<{ status: UploadStatus }> = ({ status }) => {
  const iconMap = {
    pending: <ClockCircleOutlined style={{ color: '#faad14' }} />,
    uploading: <CloudUploadOutlined style={{ color: '#1677ff' }} />,
    paused: <PauseCircleOutlined style={{ color: '#faad14' }} />,
    completed: <CheckCircleOutlined style={{ color: '#52c41a' }} />,
    failed: <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />,
    cancelled: <CloseCircleOutlined style={{ color: '#8c8c8c' }} />,
  };

  return iconMap[status] || iconMap.pending;
};

// 状态标签映射
const StatusTag: React.FC<{ status: UploadStatus }> = ({ status }) => {
  const tagMap = {
    pending: <Tag color="warning">等待中</Tag>,
    uploading: <Tag color="processing">上传中</Tag>,
    paused: <Tag color="warning">已暂停</Tag>,
    completed: <Tag color="success">已完成</Tag>,
    failed: <Tag color="error">失败</Tag>,
    cancelled: <Tag color="default">已取消</Tag>,
  };

  return tagMap[status] || tagMap.pending;
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
  if (seconds <= 0 || !isFinite(seconds)) return '--';
  
  if (seconds < 60) return `${Math.round(seconds)}秒`;
  if (seconds < 3600) return `${Math.round(seconds / 60)}分钟`;
  return `${Math.round(seconds / 3600)}小时`;
};

export const UploadQueue: React.FC<UploadQueueProps> = ({
  visible,
  onClose,
  uploadTasks,
  onPause,
  onCancel,
  onRetry,
  onClearCompleted,
  onClearAll,
  totalTasks,
  completedTasks,
  failedTasks,
  activeTasks,
}) => {
  // 计算总体进度
  const overallProgress = totalTasks > 0 
    ? Math.round((completedTasks / totalTasks) * 100)
    : 0;

  // 渲染任务操作按钮
  const renderTaskActions = (task: UploadTask) => {
    const actions: React.ReactNode[] = [];

    switch (task.status) {
      case 'uploading':
        actions.push(
          <Tooltip title="暂停上传" key="pause">
            <Button
              type="text"
              size="small"
              icon={<PauseCircleOutlined />}
              onClick={() => onPause(task.id)}
            />
          </Tooltip>
        );
        break;

      case 'paused':
        actions.push(
          <Tooltip title="继续上传" key="resume">
            <Button
              type="text"
              size="small"
              icon={<PlayCircleOutlined />}
              onClick={() => onRetry(task.id)}
            />
          </Tooltip>
        );
        break;

      case 'failed':
        actions.push(
          <Tooltip title="重试上传" key="retry">
            <Button
              type="text"
              size="small"
              icon={<ReloadOutlined />}
              onClick={() => onRetry(task.id)}
            />
          </Tooltip>
        );
        break;
    }

    // 所有状态都可以取消/删除
    if (task.status !== 'completed') {
      actions.push(
        <Tooltip title="取消上传" key="cancel">
          <Button
            type="text"
            size="small"
            icon={<CloseCircleOutlined />}
            onClick={() => onCancel(task.id)}
            danger
          />
        </Tooltip>
      );
    }

    return <Space size={4}>{actions}</Space>;
  };

  // 渲染任务详情
  const renderTaskDetails = (task: UploadTask) => {
    const details: React.ReactNode[] = [];

    // 文件大小信息
    details.push(
      <Text type="secondary" key="size">
        {formatFileSize(task.uploadedBytes)} / {formatFileSize(task.totalBytes)}
      </Text>
    );

    // 上传速度 (仅在上传中显示)
    if (task.status === 'uploading' && task.speed > 0) {
      details.push(
        <Text type="secondary" key="speed">
          {formatSpeed(task.speed)}
        </Text>
      );
    }

    // 预估剩余时间 (仅在上传中显示)
    if (task.status === 'uploading' && task.estimatedTime) {
      details.push(
        <Text type="secondary" key="eta">
          剩余 {formatEstimatedTime(task.estimatedTime)}
        </Text>
      );
    }

    // 错误信息
    if (task.status === 'failed' && task.error) {
      details.push(
        <Text type="danger" key="error">
          {task.error}
        </Text>
      );
    }

    return <Space size={8} wrap>{details}</Space>;
  };

  return (
    <Drawer
      title={
        <Space>
          <CloudUploadOutlined />
          <span>上传队列</span>
          {totalTasks > 0 && (
            <Badge 
              count={activeTasks} 
              style={{ backgroundColor: '#1677ff' }}
            />
          )}
        </Space>
      }
      placement="right"
      width={480}
      open={visible}
      onClose={onClose}
      extra={
        <Space>
          {completedTasks > 0 && (
            <Button
              size="small"
              icon={<DeleteOutlined />}
              onClick={onClearCompleted}
            >
              清理已完成
            </Button>
          )}
          {totalTasks > 0 && (
            <Button
              size="small"
              danger
              onClick={onClearAll}
            >
              清空队列
            </Button>
          )}
        </Space>
      }
    >
      {/* 总体统计 */}
      {totalTasks > 0 && (
        <div style={{ marginBottom: 16 }}>
          <div style={{ marginBottom: 8 }}>
            <Text strong>总体进度</Text>
            <Text type="secondary" style={{ float: 'right' }}>
              {completedTasks}/{totalTasks} 个文件
            </Text>
          </div>
          <Progress 
            percent={overallProgress} 
            strokeColor="#1677ff"
            showInfo={false}
          />
          <div style={{ marginTop: 8, display: 'flex', justifyContent: 'space-between' }}>
            <Space size={16}>
              <Text type="secondary">
                <CloudUploadOutlined /> {activeTasks} 上传中
              </Text>
              <Text type="secondary">
                <CheckCircleOutlined /> {completedTasks} 已完成
              </Text>
              {failedTasks > 0 && (
                <Text type="secondary">
                  <ExclamationCircleOutlined /> {failedTasks} 失败
                </Text>
              )}
            </Space>
          </div>
        </div>
      )}

      <Divider />

      {/* 上传任务列表 */}
      {uploadTasks.length === 0 ? (
        <Empty
          description="没有上传任务"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
      ) : (
        <List
          dataSource={uploadTasks}
          renderItem={(task) => (
            <List.Item
              key={task.id}
              actions={[renderTaskActions(task)]}
              style={{ paddingLeft: 0, paddingRight: 0 }}
            >
              <div style={{ width: '100%' }}>
                {/* 文件名和状态 */}
                <div style={{ display: 'flex', alignItems: 'center', marginBottom: 8 }}>
                  <StatusIcon status={task.status} />
                  <Text 
                    strong 
                    style={{ 
                      marginLeft: 8, 
                      flex: 1,
                      wordBreak: 'break-all',
                    }}
                    title={task.file.name}
                  >
                    {task.file.name}
                  </Text>
                  <StatusTag status={task.status} />
                </div>

                {/* 进度条 */}
                {(task.status === 'uploading' || task.status === 'paused') && (
                  <div style={{ marginBottom: 8 }}>
                    <Progress
                      percent={task.progress}
                      size="small"
                      strokeColor={task.status === 'paused' ? '#faad14' : '#1677ff'}
                      showInfo={false}
                    />
                  </div>
                )}

                {/* 任务详情 */}
                <div>
                  {renderTaskDetails(task)}
                </div>
              </div>
            </List.Item>
          )}
        />
      )}
    </Drawer>
  );
};