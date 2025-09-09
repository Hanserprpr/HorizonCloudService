import React, { useState } from 'react';
import {
  Modal,
  Tabs,
  Button,
  Space,
  Typography,
  Badge,
  FloatButton,
  notification,
} from 'antd';
import {
  CloudUploadOutlined,
  UploadOutlined,
  BarsOutlined,
  SettingOutlined,
  CloseOutlined,
} from '@ant-design/icons';
import { useChunkedUpload } from '@hooks/useChunkedUpload';
import { UploadDropzone } from './UploadDropzone';
import { UploadQueue } from './UploadQueue';
import { UploadProgress } from './UploadProgress';

const { Title } = Typography;

interface UploadManagerProps {
  visible: boolean;
  onClose: () => void;
  currentFolderId?: number;
  onUploadComplete?: () => void;
}

// 上传配置
interface UploadSettings {
  maxFileSize: number;
  maxFiles: number;
  autoStart: boolean;
  concurrentUploads: number;
  chunkSize: number;
}

const DEFAULT_SETTINGS: UploadSettings = {
  maxFileSize: 10 * 1024 * 1024 * 1024, // 10GB
  maxFiles: 100,
  autoStart: true,
  concurrentUploads: 3,
  chunkSize: 5 * 1024 * 1024, // 5MB
};

export const UploadManager: React.FC<UploadManagerProps> = ({
  visible,
  onClose,
  currentFolderId,
  onUploadComplete,
}) => {
  const [activeTab, setActiveTab] = useState('dropzone');
  const [settings, setSettings] = useState<UploadSettings>(DEFAULT_SETTINGS);

  // 使用分片上传Hook
  const {
    uploadTasks,
    isUploading,
    startUpload,
    pauseUpload,
    cancelUpload,
    retryUpload,
    clearCompletedTasks,
    clearAllTasks,
    totalTasks,
    completedTasks,
    failedTasks,
    activeTasks,
  } = useChunkedUpload();

  // 处理文件上传
  const handleUpload = async (files: File[]) => {
    if (files.length === 0) return;

    try {
      // 显示上传通知
      notification.info({
        message: '开始上传',
        description: `开始上传 ${files.length} 个文件`,
        duration: 3,
      });

      // 切换到进度标签页
      setActiveTab('progress');

      // 开始上传
      await startUpload(files, currentFolderId);

      // 上传完成回调
      onUploadComplete?.();

      // 显示完成通知
      if (completedTasks > 0) {
        notification.success({
          message: '上传完成',
          description: `成功上传 ${completedTasks} 个文件`,
          duration: 5,
        });
      }

    } catch (error: any) {
      console.error('Upload failed:', error);
      notification.error({
        message: '上传失败',
        description: error.message || '上传过程中发生错误',
        duration: 8,
      });
    }
  };

  // 处理恢复上传
  const handleResumeUpload = (taskId: string) => {
    // 恢复上传实际上是重新开始上传该文件
    retryUpload(taskId);
  };

  // 渲染标签页项目
  const tabItems = [
    {
      key: 'dropzone',
      label: (
        <Space>
          <UploadOutlined />
          选择文件
        </Space>
      ),
      children: (
        <UploadDropzone
          onUpload={handleUpload}
          maxFileSize={settings.maxFileSize}
          maxFiles={settings.maxFiles}
          multiple={true}
          disabled={isUploading}
        />
      ),
    },
    {
      key: 'progress',
      label: (
        <Space>
          <BarsOutlined />
          上传进度
          {activeTasks > 0 && (
            <Badge count={activeTasks} size="small" />
          )}
        </Space>
      ),
      children: totalTasks > 0 ? (
        <UploadProgress
          uploadTasks={uploadTasks}
          totalTasks={totalTasks}
          completedTasks={completedTasks}
          failedTasks={failedTasks}
          activeTasks={activeTasks}
          onPause={pauseUpload}
          onResume={handleResumeUpload}
          onCancel={cancelUpload}
          onRetry={retryUpload}
        />
      ) : (
        <div style={{ textAlign: 'center', padding: '60px 0' }}>
          <CloudUploadOutlined style={{ fontSize: 48, color: '#d9d9d9' }} />
          <div style={{ marginTop: 16, color: '#8c8c8c' }}>
            没有上传任务
          </div>
        </div>
      ),
    },
    {
      key: 'queue',
      label: (
        <Space>
          <BarsOutlined />
          上传队列
          {totalTasks > 0 && (
            <Badge count={totalTasks} size="small" />
          )}
        </Space>
      ),
      children: (
        <div style={{ minHeight: 400 }}>
          <UploadQueue
            visible={true}
            onClose={() => {}}
            uploadTasks={uploadTasks}
            onPause={pauseUpload}
            onCancel={cancelUpload}
            onRetry={retryUpload}
            onClearCompleted={clearCompletedTasks}
            onClearAll={clearAllTasks}
            totalTasks={totalTasks}
            completedTasks={completedTasks}
            failedTasks={failedTasks}
            activeTasks={activeTasks}
          />
        </div>
      ),
    },
  ];

  return (
    <>
      <Modal
        title={
          <Space>
            <CloudUploadOutlined />
            <span>文件上传管理器</span>
            {activeTasks > 0 && (
              <Badge 
                count={activeTasks} 
                style={{ backgroundColor: '#1677ff' }}
              />
            )}
          </Space>
        }
        open={visible}
        onCancel={onClose}
        width={800}
        footer={null}
        destroyOnClose={false}
        maskClosable={!isUploading}
        closable={!isUploading}
      >
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={tabItems}
          style={{ minHeight: 500 }}
        />

        {/* 底部操作栏 */}
        {totalTasks > 0 && (
          <div
            style={{
              position: 'sticky',
              bottom: 0,
              background: '#fff',
              padding: '16px 0',
              borderTop: '1px solid #f0f0f0',
              marginTop: 16,
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <Space size={16}>
                  <span>
                    总计: <strong>{totalTasks}</strong> 个文件
                  </span>
                  <span>
                    已完成: <strong style={{ color: '#52c41a' }}>{completedTasks}</strong>
                  </span>
                  {failedTasks > 0 && (
                    <span>
                      失败: <strong style={{ color: '#ff4d4f' }}>{failedTasks}</strong>
                    </span>
                  )}
                  {activeTasks > 0 && (
                    <span>
                      上传中: <strong style={{ color: '#1677ff' }}>{activeTasks}</strong>
                    </span>
                  )}
                </Space>
              </div>
              <Space>
                {completedTasks > 0 && (
                  <Button size="small" onClick={clearCompletedTasks}>
                    清理已完成
                  </Button>
                )}
                <Button size="small" danger onClick={clearAllTasks}>
                  清空队列
                </Button>
              </Space>
            </div>
          </div>
        )}
      </Modal>

      {/* 悬浮上传按钮 (当模态框关闭时显示) */}
      {!visible && activeTasks > 0 && (
        <FloatButton
          icon={<CloudUploadOutlined />}
          badge={{ count: activeTasks }}
          type="primary"
          style={{ right: 24, bottom: 80 }}
          onClick={() => setActiveTab('progress')}
        />
      )}
    </>
  );
};

// 导出轻量级上传触发器
export const UploadTrigger: React.FC<{
  onUpload: (files: File[]) => void;
  disabled?: boolean;
  children?: React.ReactNode;
}> = ({ onUpload, disabled = false, children }) => {
  const fileInputRef = React.useRef<HTMLInputElement>(null);

  const handleClick = () => {
    if (!disabled) {
      fileInputRef.current?.click();
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      onUpload(Array.from(files));
    }
    // 清空input值
    e.target.value = '';
  };

  return (
    <>
      <div onClick={handleClick} style={{ cursor: disabled ? 'not-allowed' : 'pointer' }}>
        {children || (
          <Button 
            type="primary" 
            icon={<UploadOutlined />} 
            disabled={disabled}
          >
            上传文件
          </Button>
        )}
      </div>
      <input
        ref={fileInputRef}
        type="file"
        multiple
        style={{ display: 'none' }}
        onChange={handleFileChange}
      />
    </>
  );
};