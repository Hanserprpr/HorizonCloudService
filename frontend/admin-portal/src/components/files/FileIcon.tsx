import React from 'react';
import {
  FileImageOutlined,
  FileOutlined,
  FilePdfOutlined,
  FileWordOutlined,
  FileExcelOutlined,
  FilePptOutlined,
  FileZipOutlined,
  FileTextOutlined,
  FolderOutlined,
  FolderOpenOutlined,
  CodeOutlined,
  PlayCircleOutlined,
  SoundOutlined,
} from '@ant-design/icons';

export interface FileIconProps {
  fileName?: string;
  contentType?: string;
  isFolder?: boolean;
  isOpen?: boolean;
  size?: number;
  style?: React.CSSProperties;
}

const FileIcon: React.FC<FileIconProps> = ({
  fileName,
  contentType,
  isFolder = false,
  isOpen = false,
  size = 16,
  style,
}) => {
  const iconStyle = {
    fontSize: size,
    ...style,
  };

  // 文件夹图标
  if (isFolder) {
    return isOpen ? (
      <FolderOpenOutlined style={{ ...iconStyle, color: '#1677ff' }} />
    ) : (
      <FolderOutlined style={{ ...iconStyle, color: '#1677ff' }} />
    );
  }

  // 根据文件类型返回图标
  const getFileIcon = () => {
    if (!contentType && !fileName) {
      return <FileOutlined style={{ ...iconStyle, color: '#8c8c8c' }} />;
    }

    // 优先使用 contentType 判断
    if (contentType) {
      if (contentType.startsWith('image/')) {
        return <FileImageOutlined style={{ ...iconStyle, color: '#52c41a' }} />;
      }
      if (contentType.startsWith('video/')) {
        return <PlayCircleOutlined style={{ ...iconStyle, color: '#722ed1' }} />;
      }
      if (contentType.startsWith('audio/')) {
        return <SoundOutlined style={{ ...iconStyle, color: '#fa8c16' }} />;
      }
      if (contentType === 'application/pdf') {
        return <FilePdfOutlined style={{ ...iconStyle, color: '#f5222d' }} />;
      }
      if (contentType.includes('word') || contentType.includes('document')) {
        return <FileWordOutlined style={{ ...iconStyle, color: '#1890ff' }} />;
      }
      if (contentType.includes('excel') || contentType.includes('spreadsheet')) {
        return <FileExcelOutlined style={{ ...iconStyle, color: '#52c41a' }} />;
      }
      if (contentType.includes('powerpoint') || contentType.includes('presentation')) {
        return <FilePptOutlined style={{ ...iconStyle, color: '#fa541c' }} />;
      }
      if (contentType.includes('zip') || contentType.includes('rar') || contentType.includes('7z')) {
        return <FileZipOutlined style={{ ...iconStyle, color: '#722ed1' }} />;
      }
      if (contentType.startsWith('text/')) {
        return <FileTextOutlined style={{ ...iconStyle, color: '#13c2c2' }} />;
      }
    }

    // 使用文件扩展名判断
    if (fileName) {
      const extension = fileName.toLowerCase().split('.').pop();
      
      switch (extension) {
        case 'jpg':
        case 'jpeg':
        case 'png':
        case 'gif':
        case 'bmp':
        case 'webp':
        case 'svg':
          return <FileImageOutlined style={{ ...iconStyle, color: '#52c41a' }} />;
        
        case 'mp4':
        case 'avi':
        case 'mov':
        case 'wmv':
        case 'flv':
        case 'webm':
          return <PlayCircleOutlined style={{ ...iconStyle, color: '#722ed1' }} />;
        
        case 'mp3':
        case 'wav':
        case 'flac':
        case 'aac':
        case 'ogg':
          return <SoundOutlined style={{ ...iconStyle, color: '#fa8c16' }} />;
        
        case 'pdf':
          return <FilePdfOutlined style={{ ...iconStyle, color: '#f5222d' }} />;
        
        case 'doc':
        case 'docx':
          return <FileWordOutlined style={{ ...iconStyle, color: '#1890ff' }} />;
        
        case 'xls':
        case 'xlsx':
          return <FileExcelOutlined style={{ ...iconStyle, color: '#52c41a' }} />;
        
        case 'ppt':
        case 'pptx':
          return <FilePptOutlined style={{ ...iconStyle, color: '#fa541c' }} />;
        
        case 'zip':
        case 'rar':
        case '7z':
        case 'tar':
        case 'gz':
          return <FileZipOutlined style={{ ...iconStyle, color: '#722ed1' }} />;
        
        case 'txt':
        case 'md':
        case 'log':
          return <FileTextOutlined style={{ ...iconStyle, color: '#13c2c2' }} />;
        
        case 'js':
        case 'ts':
        case 'jsx':
        case 'tsx':
        case 'html':
        case 'css':
        case 'json':
        case 'xml':
        case 'py':
        case 'java':
        case 'cpp':
        case 'c':
        case 'go':
        case 'rs':
        case 'php':
          return <CodeOutlined style={{ ...iconStyle, color: '#13c2c2' }} />;
        
        default:
          return <FileOutlined style={{ ...iconStyle, color: '#8c8c8c' }} />;
      }
    }

    return <FileOutlined style={{ ...iconStyle, color: '#8c8c8c' }} />;
  };

  return getFileIcon();
};

export default FileIcon;