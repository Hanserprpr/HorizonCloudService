import React, { useState, useEffect } from 'react';
import { Row, Col, Typography, message, Modal, Spin, FloatButton } from 'antd';
import { useParams, useNavigate } from 'react-router-dom';
import { CloudUploadOutlined } from '@ant-design/icons';

// Components
import FileList from '@components/files/FileList';
import FolderNavigation from '@components/files/FolderNavigation';
import FileSearch from '@components/files/FileSearch';
import FileToolbar from '@components/files/FileToolbar';
import { UploadManager } from '@components/upload';

// Hooks
import {
  useFolderContents,
  useFolderTree,
  useCreateFolder,
  useDeleteFile,
  useDeleteFolder,
  useRenameFile,
  useRenameFolder,
  useMoveFile,
  useCopyFile,
  useBatchOperation,
} from '@hooks/useFileManager';
import { useFileStore } from '@stores/fileStore';

// Types
import type { ViewMode } from '@components/files/FileList';
import type { FileSearchParams } from '@components/files/FileSearch';
import type { BreadcrumbItem } from '@components/files/FolderNavigation';
import type { FileItem, FolderItem } from '../../types';

const { Title } = Typography;

const FilesPage: React.FC = () => {
  const { folderId } = useParams();
  const navigate = useNavigate();
  
  // 状态管理
  const {
    currentFolderId,
    viewMode,
    selectedFileIds,
    searchParams,
    breadcrumb,
    setCurrentFolderId,
    setViewMode,
    setSelectedFileIds,
    setSearchParams,
    setBreadcrumb,
  } = useFileStore();

  // 本地状态
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
  const [uploadManagerVisible, setUploadManagerVisible] = useState(false);
  const [uploadProgress, setUploadProgress] = useState<Record<string, number>>({});

  // 当前文件夹ID（从URL参数或store获取）
  const effectiveFolderId = folderId ? parseInt(folderId) : (currentFolderId || 0);

  // 数据获取
  const { 
    data: folderContents, 
    isLoading: contentsLoading, 
    refetch: refetchContents 
  } = useFolderContents(effectiveFolderId, searchParams);

  const { 
    data: folderTree, 
    isLoading: treeLoading 
  } = useFolderTree();

  // Mutations
  const createFolderMutation = useCreateFolder();
  const deleteFileMutation = useDeleteFile();
  const deleteFolderMutation = useDeleteFolder();
  const renameFileMutation = useRenameFile();
  const renameFolderMutation = useRenameFolder();
  const moveFileMutation = useMoveFile();
  const copyFileMutation = useCopyFile();
  const batchOperationMutation = useBatchOperation();

  // 初始化面包屑
  useEffect(() => {
    if (effectiveFolderId && !breadcrumb.length) {
      // TODO: 根据文件夹ID构建面包屑路径
      setBreadcrumb([]);
    }
  }, [effectiveFolderId, breadcrumb, setBreadcrumb]);

  // 同步文件夹ID
  useEffect(() => {
    if (effectiveFolderId !== currentFolderId) {
      setCurrentFolderId(effectiveFolderId);
    }
  }, [effectiveFolderId, currentFolderId, setCurrentFolderId]);

  // 处理视图模式切换
  const handleViewModeChange = (mode: ViewMode) => {
    setViewMode(mode);
  };

  // 处理选择变化
  const handleSelectionChange = (selectedIds: number[]) => {
    setSelectedFileIds(selectedIds);
  };

  // 处理项目点击
  const handleItemClick = (item: FileItem | FolderItem) => {
    console.log('点击项目:', item);
  };

  // 处理项目双击（进入文件夹）
  const handleItemDoubleClick = (item: FileItem | FolderItem) => {
    if ('folder_type' in item || !('content_type' in item)) {
      // 是文件夹，导航到该文件夹
      navigate(`/files/${item.id}`);
      setCurrentFolderId(item.id);
      setBreadcrumb([...breadcrumb, { 
        id: item.id, 
        name: item.name, 
        path: `/files/${item.id}` 
      }]);
    } else {
      // 是文件，执行预览或下载
      handlePreview(item as FileItem);
    }
  };

  // 处理搜索
  const handleSearch = (params: FileSearchParams) => {
    setSearchParams(params);
  };

  // 处理搜索参数变化
  const handleSearchParamsChange = (params: FileSearchParams) => {
    setSearchParams(params);
  };

  // 处理清空搜索
  const handleClearSearch = () => {
    setSearchParams({});
  };

  // 处理面包屑点击
  const handleBreadcrumbClick = (folderId: number) => {
    if (folderId === 0) {
      navigate('/files');
      setCurrentFolderId(0);
      setBreadcrumb([]);
    } else {
      navigate(`/files/${folderId}`);
      setCurrentFolderId(folderId);
      // TODO: 更新面包屑到对应层级
    }
  };

  // 处理文件夹树选择
  const handleTreeSelect = (selectedKeys: React.Key[]) => {
    if (selectedKeys.length > 0) {
      const folderId = selectedKeys[0] as number;
      navigate(`/files/${folderId}`);
      setCurrentFolderId(folderId);
    }
  };

  // 处理文件夹树展开
  const handleTreeExpand = (expandedKeys: React.Key[]) => {
    setExpandedKeys(expandedKeys);
  };

  // 处理文件上传 (打开上传管理器)
  const handleUpload = (files: File[]) => {
    if (files.length > 0) {
      setUploadManagerVisible(true);
      // 文件上传将由UploadManager组件处理
    }
  };

  // 处理上传完成
  const handleUploadComplete = () => {
    // 刷新文件列表
    refetchContents();
    message.success('文件上传完成');
  };

  // 打开上传管理器
  const openUploadManager = () => {
    setUploadManagerVisible(true);
  };

  // 关闭上传管理器
  const closeUploadManager = () => {
    setUploadManagerVisible(false);
  };

  // 处理创建文件夹
  const handleCreateFolder = (name: string) => {
    createFolderMutation.mutate({
      name,
      parent_id: effectiveFolderId || 0,
    });
  };

  // 处理文件下载
  const handleDownload = (file: FileItem) => {
    // TODO: 实现下载逻辑
    console.log('下载文件:', file);
    message.info('文件下载功能开发中');
  };

  // 处理重命名
  const handleRename = (item: FileItem | FolderItem, newName: string) => {
    if ('content_type' in item) {
      // 文件重命名
      renameFileMutation.mutate({
        fileId: item.id,
        newName,
      });
    } else {
      // 文件夹重命名
      renameFolderMutation.mutate({
        folderId: item.id,
        newName,
      });
    }
  };

  // 处理删除
  const handleDelete = (item: FileItem | FolderItem) => {
    if ('content_type' in item) {
      // 删除文件
      deleteFileMutation.mutate(item.id);
    } else {
      // 删除文件夹
      deleteFolderMutation.mutate(item.id);
    }
  };

  // 处理复制
  const handleCopy = (item: FileItem | FolderItem) => {
    // TODO: 实现复制逻辑
    console.log('复制项目:', item);
    message.info('复制功能开发中');
  };

  // 处理移动
  const handleMove = (item: FileItem | FolderItem) => {
    // TODO: 实现移动逻辑
    console.log('移动项目:', item);
    message.info('移动功能开发中');
  };

  // 处理预览
  const handlePreview = (file: FileItem) => {
    if (file.content_type?.startsWith('image/')) {
      Modal.info({
        title: file.file_name,
        content: (
          <div style={{ textAlign: 'center', padding: '20px 0' }}>
            <img
              src={file.download_url || '#'}
              alt={file.file_name}
              style={{ maxWidth: '100%', maxHeight: '400px' }}
              onError={(e) => {
                (e.target as HTMLImageElement).style.display = 'none';
                message.error('图片预览失败');
              }}
            />
          </div>
        ),
        width: 800,
        okText: '关闭',
      });
    } else {
      message.info('该文件类型暂不支持预览');
    }
  };

  // 处理显示详情
  const handleShowInfo = (item: FileItem | FolderItem) => {
    // TODO: 实现详情显示
    console.log('显示详情:', item);
    message.info('详情显示功能开发中');
  };

  // 处理批量删除
  const handleBatchDelete = (ids: number[]) => {
    batchOperationMutation.mutate({
      operation: 'delete',
      file_ids: ids.filter(id => 
        folderContents?.files?.find(f => f.id === id)
      ),
      folder_ids: ids.filter(id => 
        folderContents?.folders?.find(f => f.id === id)
      ),
    });
  };

  // 处理批量下载
  const handleBatchDownload = (ids: number[]) => {
    console.log('批量下载:', ids);
    message.info('批量下载功能开发中');
  };

  // 处理刷新
  const handleRefresh = () => {
    refetchContents();
  };

  // 处理全选
  const handleSelectAll = () => {
    if (folderContents) {
      const allIds = [
        ...(folderContents.files || []).map(f => f.id),
        ...(folderContents.folders || []).map(f => f.id),
      ];
      setSelectedFileIds(allIds);
    }
  };

  // 处理取消选择
  const handleSelectNone = () => {
    setSelectedFileIds([]);
  };

  // 合并文件和文件夹数据
  const allItems = [
    ...(folderContents?.folders || []),
    ...(folderContents?.files || []),
  ];

  const selectedItems = allItems.filter(item => selectedFileIds.includes(item.id));

  return (
    <div style={{ padding: '0 4px' }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0, marginBottom: 8 }}>
          文件管理
        </Title>
      </div>

      <Row gutter={[16, 16]}>
        {/* 左侧导航 */}
        <Col xs={24} lg={6}>
          <FolderNavigation
            currentPath={breadcrumb}
            folderTree={folderTree || []}
            expandedKeys={expandedKeys}
            selectedKeys={effectiveFolderId ? [effectiveFolderId] : []}
            onBreadcrumbClick={handleBreadcrumbClick}
            onTreeSelect={handleTreeSelect}
            onTreeExpand={handleTreeExpand}
            loading={treeLoading}
          />
        </Col>

        {/* 右侧内容 */}
        <Col xs={24} lg={18}>
          {/* 搜索栏 */}
          <div style={{ marginBottom: 16 }}>
            <FileSearch
              value={searchParams}
              onChange={handleSearchParamsChange}
              onSearch={handleSearch}
              onClear={handleClearSearch}
              loading={contentsLoading}
            />
          </div>

          {/* 工具栏 */}
          <FileToolbar
            selectedItems={selectedItems}
            onUpload={openUploadManager}
            onCreateFolder={handleCreateFolder}
            onBatchDelete={handleBatchDelete}
            onBatchDownload={handleBatchDownload}
            onRefresh={handleRefresh}
            onSelectAll={handleSelectAll}
            onSelectNone={handleSelectNone}
            uploadProgress={uploadProgress}
            loading={contentsLoading}
          />

          {/* 文件列表 */}
          <Spin spinning={contentsLoading}>
            <FileList
              files={allItems}
              loading={contentsLoading}
              viewMode={viewMode}
              selectedIds={selectedFileIds}
              onViewModeChange={handleViewModeChange}
              onSelectionChange={handleSelectionChange}
              onItemClick={handleItemClick}
              onItemDoubleClick={handleItemDoubleClick}
              onDownload={handleDownload}
              onRename={handleRename}
              onDelete={handleDelete}
              onCopy={handleCopy}
              onMove={handleMove}
              onPreview={handlePreview}
              onShowInfo={handleShowInfo}
            />
          </Spin>
        </Col>
      </Row>

      {/* 上传管理器 */}
      <UploadManager
        visible={uploadManagerVisible}
        onClose={closeUploadManager}
        currentFolderId={effectiveFolderId}
        onUploadComplete={handleUploadComplete}
      />

      {/* 悬浮上传按钮 */}
      <FloatButton
        icon={<CloudUploadOutlined />}
        type="primary"
        style={{ right: 24, bottom: 24 }}
        onClick={openUploadManager}
        tooltip="上传文件"
      />
    </div>
  );
};

export default FilesPage;