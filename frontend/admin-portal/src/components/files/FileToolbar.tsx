import React, { useState } from 'react';
import {
  Space,
  Button,
  Dropdown,
  Modal,
  Input,
  message,
  Typography,
} from 'antd';
import {
  CloudUploadOutlined,
  FolderAddOutlined,
  DeleteOutlined,
  ScissorOutlined,
  CopyOutlined,
  DownloadOutlined,
  MoreOutlined,
  ReloadOutlined,
  SelectOutlined,
} from '@ant-design/icons';
import type { MenuProps } from 'antd';
import type { FileItem, FolderItem } from '../../types';

const { Text } = Typography;

export interface FileToolbarProps {
  selectedItems?: (FileItem | FolderItem)[];
  onUpload?: () => void; // 改为不接收参数，只是打开上传管理器
  onCreateFolder?: (name: string) => void;
  onBatchDelete?: (ids: number[]) => void;
  onBatchMove?: (ids: number[], targetFolderId: number) => void;
  onBatchCopy?: (ids: number[], targetFolderId: number) => void;
  onBatchDownload?: (ids: number[]) => void;
  onRefresh?: () => void;
  onSelectAll?: () => void;
  onSelectNone?: () => void;
  uploadProgress?: Record<string, number>;
  loading?: boolean;
}

const FileToolbar: React.FC<FileToolbarProps> = ({
  selectedItems = [],
  onUpload,
  onCreateFolder,
  onBatchDelete,
  onBatchMove,
  onBatchCopy,
  onBatchDownload,
  onRefresh,
  onSelectAll,
  onSelectNone,
  uploadProgress = {},
  loading = false,
}) => {
  const [createFolderVisible, setCreateFolderVisible] = useState(false);
  const [folderName, setFolderName] = useState('');

  const hasSelection = selectedItems.length > 0;
  const selectionCount = selectedItems.length;

  // 处理文件上传（现在只是打开上传管理器）
  const handleUploadClick = () => {
    if (onUpload) {
      onUpload();
    }
  };

  // 创建文件夹
  const handleCreateFolder = () => {
    if (!folderName.trim()) {
      message.error('请输入文件夹名称');
      return;
    }
    
    onCreateFolder?.(folderName.trim());
    setCreateFolderVisible(false);
    setFolderName('');
  };

  // 批量操作菜单
  const batchMenuItems: MenuProps['items'] = [
    {
      key: 'download',
      icon: <DownloadOutlined />,
      label: '批量下载',
      onClick: () => {
        const fileIds = selectedItems
          .filter(item => 'content_type' in item)
          .map(item => item.id);
        if (fileIds.length === 0) {
          message.warning('请选择要下载的文件');
          return;
        }
        onBatchDownload?.(fileIds);
      },
    },
    {
      key: 'copy',
      icon: <CopyOutlined />,
      label: '批量复制',
      onClick: () => {
        // TODO: 实现选择目标文件夹的逻辑
        message.info('请选择目标文件夹（功能开发中）');
      },
    },
    {
      key: 'move',
      icon: <ScissorOutlined />,
      label: '批量移动',
      onClick: () => {
        // TODO: 实现选择目标文件夹的逻辑
        message.info('请选择目标文件夹（功能开发中）');
      },
    },
    {
      type: 'divider',
    },
    {
      key: 'delete',
      icon: <DeleteOutlined />,
      label: '批量删除',
      danger: true,
      onClick: () => {
        Modal.confirm({
          title: '确认批量删除?',
          content: `确定要删除选中的 ${selectionCount} 个项目吗？此操作不可撤销。`,
          okText: '确认删除',
          okType: 'danger',
          cancelText: '取消',
          onOk: () => {
            const ids = selectedItems.map(item => item.id);
            onBatchDelete?.(ids);
          },
        });
      },
    },
  ];

  // 选择菜单
  const selectMenuItems: MenuProps['items'] = [
    {
      key: 'all',
      label: '全选',
      onClick: () => onSelectAll?.(),
    },
    {
      key: 'none',
      label: '取消选择',
      onClick: () => onSelectNone?.(),
    },
  ];

  return (
    <>
      <div style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center',
        padding: '12px 0',
        borderBottom: '1px solid #f0f0f0',
        marginBottom: 16,
      }}>
        {/* 左侧操作按钮 */}
        <Space>
          <Button 
            type="primary" 
            icon={<CloudUploadOutlined />}
            onClick={handleUploadClick}
            loading={Object.keys(uploadProgress).length > 0}
          >
            上传文件
          </Button>

          <Button
            icon={<FolderAddOutlined />}
            onClick={() => setCreateFolderVisible(true)}
          >
            新建文件夹
          </Button>

          <Button
            icon={<ReloadOutlined />}
            onClick={onRefresh}
            loading={loading}
          >
            刷新
          </Button>
        </Space>

        {/* 右侧选择和批量操作 */}
        <Space>
          {hasSelection && (
            <>
              <Text type="secondary">
                已选择 {selectionCount} 项
              </Text>
              
              <Dropdown
                menu={{ items: batchMenuItems }}
                trigger={['click']}
                placement="bottomRight"
              >
                <Button icon={<MoreOutlined />}>
                  批量操作
                </Button>
              </Dropdown>
            </>
          )}

          <Dropdown
            menu={{ items: selectMenuItems }}
            trigger={['click']}
            placement="bottomRight"
          >
            <Button icon={<SelectOutlined />}>
              选择
            </Button>
          </Dropdown>
        </Space>
      </div>

      {/* 上传进度显示 */}
      {Object.keys(uploadProgress).length > 0 && (
        <div style={{ 
          marginBottom: 16,
          padding: 12,
          backgroundColor: '#f9f9f9',
          borderRadius: 6,
        }}>
          <Space direction="vertical" size={8} style={{ width: '100%' }}>
            <Text strong>文件上传中...</Text>
            {Object.entries(uploadProgress).map(([fileName, progress]) => (
              <div key={fileName}>
                <div style={{ 
                  display: 'flex', 
                  justifyContent: 'space-between',
                  marginBottom: 4,
                }}>
                  <Text ellipsis style={{ maxWidth: 200 }}>
                    {fileName}
                  </Text>
                  <Text type="secondary">
                    {progress}%
                  </Text>
                </div>
                <div style={{
                  height: 4,
                  backgroundColor: '#e6f7ff',
                  borderRadius: 2,
                  overflow: 'hidden',
                }}>
                  <div
                    style={{
                      height: '100%',
                      backgroundColor: '#1677ff',
                      width: `${progress}%`,
                      transition: 'width 0.3s ease',
                    }}
                  />
                </div>
              </div>
            ))}
          </Space>
        </div>
      )}

      {/* 创建文件夹模态框 */}
      <Modal
        title="创建文件夹"
        open={createFolderVisible}
        onOk={handleCreateFolder}
        onCancel={() => {
          setCreateFolderVisible(false);
          setFolderName('');
        }}
        okText="创建"
        cancelText="取消"
      >
        <Input
          placeholder="请输入文件夹名称"
          value={folderName}
          onChange={(e) => setFolderName(e.target.value)}
          onPressEnter={handleCreateFolder}
          maxLength={50}
        />
      </Modal>
    </>
  );
};

export default FileToolbar;