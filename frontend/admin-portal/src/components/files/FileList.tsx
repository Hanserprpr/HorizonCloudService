import React, { useState } from 'react';
import {
  Table,
  Card,
  Row,
  Col,
  Button,
  Space,
  Dropdown,
  Checkbox,
  Typography,
  Tag,
  Tooltip,
  Modal,
  Input,
  message,
  Progress,
  App,
} from 'antd';
import {
  MoreOutlined,
  DownloadOutlined,
  EditOutlined,
  DeleteOutlined,
  CopyOutlined,
  ScissorOutlined,
  EyeOutlined,
  InfoCircleOutlined,
  UnorderedListOutlined,
  AppstoreOutlined,
  ProfileOutlined,
} from '@ant-design/icons';
import type { ColumnsType, TableProps } from 'antd/es/table';
import { formatFileSize, formatDate } from '@utils/index';
import FileIcon from './FileIcon';
import type { FileItem, FolderItem } from '../../types';

const { Text, Title } = Typography;

export type ViewMode = 'list' | 'grid' | 'detail';

export interface FileListProps {
  files: (FileItem | FolderItem)[];
  loading?: boolean;
  viewMode?: ViewMode;
  selectedIds?: number[];
  onViewModeChange?: (mode: ViewMode) => void;
  onSelectionChange?: (selectedIds: number[]) => void;
  onItemClick?: (item: FileItem | FolderItem) => void;
  onItemDoubleClick?: (item: FileItem | FolderItem) => void;
  onDownload?: (item: FileItem) => void;
  onRename?: (item: FileItem | FolderItem, newName: string) => void;
  onDelete?: (item: FileItem | FolderItem) => void;
  onCopy?: (item: FileItem | FolderItem) => void;
  onMove?: (item: FileItem | FolderItem) => void;
  onPreview?: (item: FileItem) => void;
  onShowInfo?: (item: FileItem | FolderItem) => void;
}

const FileList: React.FC<FileListProps> = ({
  files,
  loading = false,
  viewMode = 'list',
  selectedIds = [],
  onViewModeChange,
  onSelectionChange,
  onItemClick,
  onItemDoubleClick,
  onDownload,
  onRename,
  onDelete,
  onCopy,
  onMove,
  onPreview,
  onShowInfo,
}) => {
  const [renameModalVisible, setRenameModalVisible] = useState(false);
  const [renameItem, setRenameItem] = useState<FileItem | FolderItem | null>(null);
  const [newName, setNewName] = useState('');
  const [deleteModalVisible, setDeleteModalVisible] = useState(false);
  const [deleteItem, setDeleteItem] = useState<FileItem | FolderItem | null>(null);
  
  // 使用Ant Design hooks来访问modal（保留作为备用方案）
  const { modal } = App.useApp();

  // 判断是否为文件夹
  const isFolder = (item: FileItem | FolderItem): item is FolderItem => {
    return 'folder_type' in item || !('content_type' in item);
  };

  // 处理菜单点击
  const handleMenuClick = (item: FileItem | FolderItem, key: string) => {
    console.log('🔍 handleMenuClick 被调用，key:', key, 'item:', item);
    
    switch (key) {
      case 'rename':
        setRenameItem(item);
        setNewName(isFolder(item) ? item.name : (item as FileItem).file_name);
        setRenameModalVisible(true);
        break;
      case 'copy':
        onCopy?.(item);
        break;
      case 'move':
        onMove?.(item);
        break;
      case 'info':
        onShowInfo?.(item);
        break;
      case 'delete':
        console.log('🗑️ 删除菜单项点击，准备显示确认对话框');
        setDeleteItem(item);
        setDeleteModalVisible(true);
        break;
      default:
        console.log('❌ 未知菜单项:', key);
    }
  };

  // 获取操作菜单
  const getActionMenu = (item: FileItem | FolderItem) => {
    const commonActions = [
      {
        key: 'rename',
        icon: <EditOutlined />,
        label: '重命名',
      },
      {
        key: 'copy',
        icon: <CopyOutlined />,
        label: '复制',
      },
      {
        key: 'move',
        icon: <ScissorOutlined />,
        label: '移动',
      },
      {
        key: 'info',
        icon: <InfoCircleOutlined />,
        label: '详细信息',
      },
      {
        type: 'divider' as const,
      },
      {
        key: 'delete',
        icon: <DeleteOutlined />,
        label: '删除',
        danger: true,
      },
    ];

    // 文件特有操作
    if (!isFolder(item)) {
      const fileItem = item as FileItem;
      return [
        {
          key: 'download',
          icon: <DownloadOutlined />,
          label: '下载',
          onClick: () => onDownload?.(fileItem),
        },
        {
          key: 'preview',
          icon: <EyeOutlined />,
          label: '预览',
          onClick: () => onPreview?.(fileItem),
          disabled: !fileItem.content_type?.startsWith('image/'),
        },
        ...commonActions,
      ];
    }

    return commonActions;
  };

  // 处理重命名
  const handleRename = () => {
    if (!renameItem || !newName.trim()) {
      message.error('请输入有效的名称');
      return;
    }

    onRename?.(renameItem, newName.trim());
    setRenameModalVisible(false);
    setRenameItem(null);
    setNewName('');
  };

  // 处理删除确认
  const handleDeleteConfirm = () => {
    console.log('✅ 确认删除按钮点击');
    if (deleteItem) {
      onDelete?.(deleteItem);
    }
    setDeleteModalVisible(false);
    setDeleteItem(null);
  };

  // 处理删除取消
  const handleDeleteCancel = () => {
    console.log('❌ 取消删除操作');
    setDeleteModalVisible(false);
    setDeleteItem(null);
  };

  // 列表视图的列定义
  const tableColumns: ColumnsType<FileItem | FolderItem> = [
    {
      title: '',
      width: 40,
      render: (_, item) => (
        <Checkbox
          checked={selectedIds.includes(item.id)}
          onChange={(e) => {
            const newSelectedIds = e.target.checked
              ? [...selectedIds, item.id]
              : selectedIds.filter(id => id !== item.id);
            onSelectionChange?.(newSelectedIds);
          }}
        />
      ),
    },
    {
      title: '名称',
      key: 'name',
      ellipsis: true,
      render: (_, item) => (
        <Space>
          <FileIcon
            fileName={isFolder(item) ? undefined : (item as FileItem).file_name}
            contentType={isFolder(item) ? undefined : (item as FileItem).content_type}
            isFolder={isFolder(item)}
            size={18}
          />
          <Text
            style={{ cursor: 'pointer' }}
            onClick={() => onItemClick?.(item)}
            onDoubleClick={() => onItemDoubleClick?.(item)}
          >
            {isFolder(item) ? item.name : (item as FileItem).file_name}
          </Text>
          {!isFolder(item) && (item as FileItem).is_favorite && (
            <Tag color="orange" size="small">收藏</Tag>
          )}
        </Space>
      ),
    },
    {
      title: '大小',
      key: 'size',
      width: 100,
      align: 'right',
      render: (_, item) => {
        if (isFolder(item)) {
          return <Text type="secondary">-</Text>;
        }
        return <Text type="secondary">{formatFileSize((item as FileItem).size)}</Text>;
      },
    },
    {
      title: '类型',
      key: 'type',
      width: 120,
      render: (_, item) => {
        if (isFolder(item)) {
          return <Tag color="blue">文件夹</Tag>;
        }
        const fileItem = item as FileItem;
        const extension = fileItem.file_name?.split('.').pop()?.toLowerCase();
        return <Tag color="default">{extension || '未知'}</Tag>;
      },
    },
    {
      title: '修改时间',
      key: 'updated_at',
      width: 160,
      render: (_, item) => (
        <Text type="secondary">{formatDate(item.updated_at)}</Text>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 60,
      align: 'center',
      render: (_, item) => (
        <Dropdown
          menu={{ 
            items: getActionMenu(item),
            onClick: ({ key }) => handleMenuClick(item, key)
          }}
          trigger={['click']}
          placement="bottomRight"
        >
          <Button type="text" icon={<MoreOutlined />} size="small" />
        </Dropdown>
      ),
    },
  ];

  // 网格视图
  const renderGridView = () => (
    <Row gutter={[16, 16]}>
      {files.map((item) => (
        <Col xs={12} sm={8} md={6} lg={4} xl={3} key={`grid-item-${item.id}`}>
          <Card
            hoverable
            size="small"
            className={selectedIds.includes(item.id) ? 'file-card-selected' : ''}
            onClick={() => onItemClick?.(item)}
            onDoubleClick={() => onItemDoubleClick?.(item)}
            cover={
              <div style={{ 
                padding: '20px', 
                textAlign: 'center',
                backgroundColor: '#fafafa',
                height: 80,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}>
                <FileIcon
                  fileName={isFolder(item) ? undefined : (item as FileItem).file_name}
                  contentType={isFolder(item) ? undefined : (item as FileItem).content_type}
                  isFolder={isFolder(item)}
                  size={32}
                />
              </div>
            }
            actions={[
              <Checkbox
                key={`grid-select-${item.id}`}
                checked={selectedIds.includes(item.id)}
                onChange={(e) => {
                  const newSelectedIds = e.target.checked
                    ? [...selectedIds, item.id]
                    : selectedIds.filter(id => id !== item.id);
                  onSelectionChange?.(newSelectedIds);
                }}
                onClick={(e) => e.stopPropagation()}
              />,
              <Dropdown
                key={`grid-more-${item.id}`}
                menu={{ 
                  items: getActionMenu(item),
                  onClick: ({ key }) => handleMenuClick(item, key)
                }}
                trigger={['click']}
                placement="bottomRight"
              >
                <Button 
                  type="text" 
                  icon={<MoreOutlined />} 
                  size="small"
                  onClick={(e) => e.stopPropagation()}
                />
              </Dropdown>,
            ]}
          >
            <Card.Meta
              title={
                <Tooltip title={isFolder(item) ? item.name : (item as FileItem).file_name}>
                  <div style={{ 
                    overflow: 'hidden', 
                    textOverflow: 'ellipsis',
                    fontSize: '12px',
                  }}>
                    {isFolder(item) ? item.name : (item as FileItem).file_name}
                  </div>
                </Tooltip>
              }
              description={
                <Space direction="vertical" size={2}>
                  {!isFolder(item) && (
                    <Text type="secondary" style={{ fontSize: '11px' }}>
                      {formatFileSize((item as FileItem).size)}
                    </Text>
                  )}
                  <Text type="secondary" style={{ fontSize: '11px' }}>
                    {formatDate(item.updated_at, 'MM-DD')}
                  </Text>
                </Space>
              }
            />
          </Card>
        </Col>
      ))}
    </Row>
  );

  // 详情视图
  const renderDetailView = () => (
    <div>
      {files.map((item) => (
        <Card
          key={`detail-item-${item.id}`}
          size="small"
          style={{ marginBottom: 8 }}
          className={selectedIds.includes(item.id) ? 'file-card-selected' : ''}
        >
          <Row align="middle" gutter={16}>
            <Col flex="none">
              <Checkbox
                checked={selectedIds.includes(item.id)}
                onChange={(e) => {
                  const newSelectedIds = e.target.checked
                    ? [...selectedIds, item.id]
                    : selectedIds.filter(id => id !== item.id);
                  onSelectionChange?.(newSelectedIds);
                }}
              />
            </Col>
            <Col flex="none">
              <FileIcon
                fileName={isFolder(item) ? undefined : (item as FileItem).file_name}
                contentType={isFolder(item) ? undefined : (item as FileItem).content_type}
                isFolder={isFolder(item)}
                size={24}
              />
            </Col>
            <Col flex="1">
              <Space direction="vertical" size={2}>
                <Text
                  strong
                  style={{ cursor: 'pointer' }}
                  onClick={() => onItemClick?.(item)}
                  onDoubleClick={() => onItemDoubleClick?.(item)}
                >
                  {isFolder(item) ? item.name : (item as FileItem).file_name}
                </Text>
                <Space>
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    {isFolder(item) 
                      ? '文件夹' 
                      : `${formatFileSize((item as FileItem).size)} • ${(item as FileItem).content_type || '未知类型'}`
                    }
                  </Text>
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    修改时间: {formatDate(item.updated_at)}
                  </Text>
                </Space>
              </Space>
            </Col>
            <Col flex="none">
              <Dropdown
                menu={{ 
                  items: getActionMenu(item),
                  onClick: ({ key }) => handleMenuClick(item, key)
                }}
                trigger={['click']}
                placement="bottomRight"
              >
                <Button type="text" icon={<MoreOutlined />} />
              </Dropdown>
            </Col>
          </Row>
        </Card>
      ))}
    </div>
  );

  return (
    <>
      <div style={{ marginBottom: 16 }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Space>
              <Text type="secondary">
                共 {files.length} 项
              </Text>
              {selectedIds.length > 0 && (
                <Text type="secondary">
                  已选择 {selectedIds.length} 项
                </Text>
              )}
            </Space>
          </Col>
          <Col>
            <Space>
              <Button.Group>
                <Button
                  icon={<UnorderedListOutlined />}
                  type={viewMode === 'list' ? 'primary' : 'default'}
                  onClick={() => onViewModeChange?.('list')}
                />
                <Button
                  icon={<AppstoreOutlined />}
                  type={viewMode === 'grid' ? 'primary' : 'default'}
                  onClick={() => onViewModeChange?.('grid')}
                />
                <Button
                  icon={<ProfileOutlined />}
                  type={viewMode === 'detail' ? 'primary' : 'default'}
                  onClick={() => onViewModeChange?.('detail')}
                />
              </Button.Group>
            </Space>
          </Col>
        </Row>
      </div>

      {viewMode === 'list' && (
        <Table
          columns={tableColumns}
          dataSource={files}
          rowKey="id"
          loading={loading}
          pagination={false}
          size="small"
        />
      )}
      
      {viewMode === 'grid' && renderGridView()}
      
      {viewMode === 'detail' && renderDetailView()}

      {/* 重命名模态框 */}
      <Modal
        title={`重命名${renameItem && isFolder(renameItem) ? '文件夹' : '文件'}`}
        open={renameModalVisible}
        onOk={handleRename}
        onCancel={() => {
          setRenameModalVisible(false);
          setRenameItem(null);
          setNewName('');
        }}
        okText="确认"
        cancelText="取消"
      >
        <Input
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          placeholder={`请输入新的${renameItem && isFolder(renameItem) ? '文件夹' : '文件'}名称`}
          onPressEnter={handleRename}
        />
      </Modal>

      {/* 删除确认模态框 */}
      <Modal
        title={`确认删除${deleteItem && isFolder(deleteItem) ? '文件夹' : '文件'}?`}
        open={deleteModalVisible}
        onOk={handleDeleteConfirm}
        onCancel={handleDeleteCancel}
        okText="确认删除"
        okType="danger"
        cancelText="取消"
        width={400}
      >
        <p>
          确定要删除 "{deleteItem && (isFolder(deleteItem) ? deleteItem.name : (deleteItem as FileItem).file_name)}" 吗？
        </p>
        <p style={{ color: '#ff4d4f', fontSize: '14px', marginTop: '8px' }}>
          此操作不可撤销。
        </p>
      </Modal>

      <style jsx>{`
        :global(.file-card-selected) {
          border-color: #1677ff;
          box-shadow: 0 0 0 2px rgba(22, 119, 255, 0.2);
        }
        
        :global(.ant-card-body) {
          padding: 12px;
        }
        
        :global(.ant-card-meta-title) {
          margin-bottom: 4px;
        }
      `}</style>
    </>
  );
};

export default FileList;