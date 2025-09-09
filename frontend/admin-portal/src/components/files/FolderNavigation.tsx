import React from 'react';
import { Breadcrumb, Tree, Card, Space, Typography, Button } from 'antd';
import { 
  FolderOutlined, 
  FolderOpenOutlined, 
  HomeOutlined,
  RightOutlined,
} from '@ant-design/icons';
import type { DataNode } from 'antd/es/tree';
import type { FolderItem } from '../../types';

const { Text } = Typography;

export interface BreadcrumbItem {
  id: number;
  name: string;
  path: string;
}

export interface FolderNavigationProps {
  currentPath: BreadcrumbItem[];
  folderTree?: FolderItem[];
  expandedKeys?: React.Key[];
  selectedKeys?: React.Key[];
  onBreadcrumbClick?: (folderId: number) => void;
  onTreeSelect?: (selectedKeys: React.Key[], info: any) => void;
  onTreeExpand?: (expandedKeys: React.Key[], info: any) => void;
  loading?: boolean;
}

const FolderNavigation: React.FC<FolderNavigationProps> = ({
  currentPath = [],
  folderTree = [],
  expandedKeys = [],
  selectedKeys = [],
  onBreadcrumbClick,
  onTreeSelect,
  onTreeExpand,
  loading = false,
}) => {
  
  // 转换文件夹数据为Tree组件需要的格式
  const convertToTreeData = (folders: FolderItem[]): DataNode[] => {
    return folders.map(folder => ({
      key: folder.id,
      title: (
        <Space size={4}>
          <Text ellipsis style={{ maxWidth: 150 }}>
            {folder.name}
          </Text>
          {folder.file_count !== undefined && folder.file_count > 0 && (
            <Text type="secondary" style={{ fontSize: '11px' }}>
              ({folder.file_count})
            </Text>
          )}
        </Space>
      ),
      icon: (props: any) => {
        const isExpanded = expandedKeys.includes(folder.id);
        return isExpanded ? (
          <FolderOpenOutlined style={{ color: '#1677ff' }} />
        ) : (
          <FolderOutlined style={{ color: '#1677ff' }} />
        );
      },
      children: folder.children ? convertToTreeData(folder.children) : undefined,
      isLeaf: !folder.children || folder.children.length === 0,
    }));
  };

  // 面包屑项目
  const breadcrumbItems = [
    {
      title: (
        <Space size={4}>
          <HomeOutlined />
          <span 
            style={{ cursor: 'pointer' }}
            onClick={() => onBreadcrumbClick?.(0)}
          >
            根目录
          </span>
        </Space>
      ),
    },
    ...currentPath.map((item, index) => ({
      title: (
        <span
          style={{ cursor: 'pointer' }}
          onClick={() => onBreadcrumbClick?.(item.id)}
        >
          {item.name}
        </span>
      ),
    })),
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      {/* 面包屑导航 */}
      <Card size="small" bodyStyle={{ padding: '8px 16px' }}>
        <Breadcrumb
          separator={<RightOutlined style={{ fontSize: 10 }} />}
          items={breadcrumbItems}
        />
      </Card>

      {/* 文件夹树 */}
      <Card 
        size="small" 
        title={
          <Space>
            <FolderOutlined />
            <Text strong>文件夹</Text>
          </Space>
        }
        bodyStyle={{ padding: '8px' }}
      >
        {folderTree.length > 0 ? (
          <Tree
            showIcon
            treeData={convertToTreeData(folderTree)}
            expandedKeys={expandedKeys}
            selectedKeys={selectedKeys}
            onSelect={onTreeSelect}
            onExpand={onTreeExpand}
            style={{ 
              background: 'transparent',
              fontSize: '13px',
            }}
            height={300}
            virtual={false}
          />
        ) : (
          <div style={{ 
            textAlign: 'center', 
            padding: '40px 20px',
            color: '#8c8c8c' 
          }}>
            {loading ? '加载中...' : '暂无文件夹'}
          </div>
        )}
      </Card>
    </Space>
  );
};

export default FolderNavigation;