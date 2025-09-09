import React, { useState, useRef, useCallback } from 'react';
import {
  Upload,
  Button,
  Space,
  Typography,
  Card,
  Alert,
  List,
  Tag,
  Progress,
} from 'antd';
import {
  CloudUploadOutlined,
  InboxOutlined,
  FileOutlined,
  DeleteOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import type { UploadProps } from 'antd';
import { formatFileSize, isImageType, isVideoType, isAudioType } from '@utils/index';

const { Dragger } = Upload;
const { Text, Title } = Typography;

interface UploadDropzoneProps {
  onUpload: (files: File[]) => void;
  accept?: string;
  maxFileSize?: number; // bytes
  maxFiles?: number;
  multiple?: boolean;
  disabled?: boolean;
  className?: string;
  style?: React.CSSProperties;
}

interface FilePreview {
  file: File;
  id: string;
  preview?: string;
}

// 默认配置
const DEFAULT_CONFIG = {
  maxFileSize: 10 * 1024 * 1024 * 1024, // 10GB
  maxFiles: 100,
  accept: '*',
};

// 获取文件类型标签
const getFileTypeTag = (file: File): React.ReactNode => {
  const { type, name } = file;
  
  if (isImageType(type)) {
    return <Tag color="green">图片</Tag>;
  }
  
  if (isVideoType(type)) {
    return <Tag color="purple">视频</Tag>;
  }
  
  if (isAudioType(type)) {
    return <Tag color="orange">音频</Tag>;
  }
  
  if (type.includes('pdf')) {
    return <Tag color="red">PDF</Tag>;
  }
  
  if (type.includes('word') || name.endsWith('.docx') || name.endsWith('.doc')) {
    return <Tag color="blue">Word</Tag>;
  }
  
  if (type.includes('excel') || name.endsWith('.xlsx') || name.endsWith('.xls')) {
    return <Tag color="green">Excel</Tag>;
  }
  
  if (type.includes('zip') || type.includes('rar') || type.includes('7z')) {
    return <Tag color="volcano">压缩包</Tag>;
  }
  
  return <Tag color="default">文件</Tag>;
};

// 生成文件预览
const generateFilePreview = (file: File): Promise<string | undefined> => {
  return new Promise((resolve) => {
    if (isImageType(file.type)) {
      const reader = new FileReader();
      reader.onload = (e) => {
        resolve(e.target?.result as string);
      };
      reader.onerror = () => resolve(undefined);
      reader.readAsDataURL(file);
    } else {
      resolve(undefined);
    }
  });
};

export const UploadDropzone: React.FC<UploadDropzoneProps> = ({
  onUpload,
  accept = DEFAULT_CONFIG.accept,
  maxFileSize = DEFAULT_CONFIG.maxFileSize,
  maxFiles = DEFAULT_CONFIG.maxFiles,
  multiple = true,
  disabled = false,
  className,
  style,
}) => {
  const [selectedFiles, setSelectedFiles] = useState<FilePreview[]>([]);
  const [dragOver, setDragOver] = useState(false);
  const [validationErrors, setValidationErrors] = useState<string[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // 生成文件ID
  const generateFileId = () => Date.now().toString(36) + Math.random().toString(36).substr(2);

  // 验证文件
  const validateFiles = useCallback((files: FileList | File[]): { valid: File[]; errors: string[] } => {
    const fileArray = Array.from(files);
    const errors: string[] = [];
    const valid: File[] = [];

    // 检查文件数量
    if (selectedFiles.length + fileArray.length > maxFiles) {
      errors.push(`最多只能选择 ${maxFiles} 个文件`);
      return { valid, errors };
    }

    for (const file of fileArray) {
      // 检查文件大小
      if (file.size > maxFileSize) {
        errors.push(`文件 "${file.name}" 大小超过限制 (${formatFileSize(maxFileSize)})`);
        continue;
      }

      // 检查文件类型 (如果指定了accept)
      if (accept !== '*' && accept !== '') {
        const acceptTypes = accept.split(',').map(type => type.trim());
        const isAcceptable = acceptTypes.some(acceptType => {
          if (acceptType.startsWith('.')) {
            return file.name.toLowerCase().endsWith(acceptType.toLowerCase());
          } else if (acceptType.includes('*')) {
            const mimeBase = acceptType.split('*')[0];
            return file.type.startsWith(mimeBase);
          } else {
            return file.type === acceptType;
          }
        });

        if (!isAcceptable) {
          errors.push(`文件 "${file.name}" 类型不受支持`);
          continue;
        }
      }

      // 检查重复文件
      const isDuplicate = selectedFiles.some(selected => 
        selected.file.name === file.name && 
        selected.file.size === file.size &&
        selected.file.lastModified === file.lastModified
      );

      if (isDuplicate) {
        errors.push(`文件 "${file.name}" 已经存在`);
        continue;
      }

      valid.push(file);
    }

    return { valid, errors };
  }, [selectedFiles, maxFileSize, maxFiles, accept]);

  // 处理文件选择
  const handleFileSelect = useCallback(async (files: FileList | File[]) => {
    const { valid, errors } = validateFiles(files);
    
    setValidationErrors(errors);
    
    if (valid.length > 0) {
      // 生成文件预览
      const previews = await Promise.all(
        valid.map(async (file) => {
          const preview = await generateFilePreview(file);
          return {
            file,
            id: generateFileId(),
            preview,
          };
        })
      );

      setSelectedFiles(prev => [...prev, ...previews]);
    }
  }, [validateFiles]);

  // 拖拽处理
  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (!disabled) {
      setDragOver(true);
    }
  }, [disabled]);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragOver(false);
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragOver(false);
    
    if (disabled) return;

    const files = e.dataTransfer.files;
    if (files.length > 0) {
      handleFileSelect(files);
    }
  }, [disabled, handleFileSelect]);

  // 文件输入处理
  const handleFileInputChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      handleFileSelect(files);
    }
    // 清空input值以允许选择相同文件
    e.target.value = '';
  }, [handleFileSelect]);

  // 移除文件
  const removeFile = useCallback((fileId: string) => {
    setSelectedFiles(prev => prev.filter(file => file.id !== fileId));
    // 如果移除文件后没有错误文件了，清空错误信息
    setValidationErrors([]);
  }, []);

  // 清空所有文件
  const clearAllFiles = useCallback(() => {
    setSelectedFiles([]);
    setValidationErrors([]);
  }, []);

  // 开始上传
  const startUpload = useCallback(() => {
    if (selectedFiles.length === 0) return;
    
    const files = selectedFiles.map(fp => fp.file);
    onUpload(files);
    
    // 清空已选择的文件
    setSelectedFiles([]);
    setValidationErrors([]);
  }, [selectedFiles, onUpload]);

  // 点击选择文件
  const selectFiles = useCallback(() => {
    fileInputRef.current?.click();
  }, []);

  const uploadProps: UploadProps = {
    name: 'file',
    multiple,
    accept: accept === '*' ? undefined : accept,
    disabled,
    showUploadList: false,
    beforeUpload: () => false, // 阻止自动上传
    onChange: (info) => {
      if (info.fileList.length > 0) {
        const files = info.fileList.map(fileItem => fileItem.originFileObj).filter(Boolean) as File[];
        handleFileSelect(files);
      }
    },
  };

  return (
    <div className={className} style={style}>
      {/* 主上传区域 */}
      <Card>
        <Dragger
          {...uploadProps}
          className={dragOver ? 'dragover' : ''}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
          style={{
            backgroundColor: dragOver ? '#f0f8ff' : '#fafafa',
            borderColor: dragOver ? '#1677ff' : '#d9d9d9',
          }}
        >
          <p className="ant-upload-drag-icon">
            <InboxOutlined style={{ color: dragOver ? '#1677ff' : '#8c8c8c' }} />
          </p>
          <p className="ant-upload-text">
            点击或将文件拖拽到此区域上传
          </p>
          <p className="ant-upload-hint">
            支持单个或批量上传。{maxFileSize !== DEFAULT_CONFIG.maxFileSize && `单文件大小限制：${formatFileSize(maxFileSize)}`}
          </p>
        </Dragger>

        {/* 隐藏的文件输入 */}
        <input
          ref={fileInputRef}
          type="file"
          accept={accept === '*' ? undefined : accept}
          multiple={multiple}
          style={{ display: 'none' }}
          onChange={handleFileInputChange}
        />

        {/* 错误信息 */}
        {validationErrors.length > 0 && (
          <Alert
            type="warning"
            style={{ marginTop: 16 }}
            message="文件选择警告"
            description={
              <ul style={{ margin: 0, paddingLeft: 20 }}>
                {validationErrors.map((error, index) => (
                  <li key={index}>{error}</li>
                ))}
              </ul>
            }
            closable
            onClose={() => setValidationErrors([])}
          />
        )}
      </Card>

      {/* 已选择的文件列表 */}
      {selectedFiles.length > 0 && (
        <Card 
          title={
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span>已选择文件 ({selectedFiles.length})</span>
              <Button
                size="small"
                type="text"
                icon={<DeleteOutlined />}
                onClick={clearAllFiles}
              >
                清空
              </Button>
            </div>
          }
          style={{ marginTop: 16 }}
        >
          <List
            dataSource={selectedFiles}
            renderItem={(filePreview) => (
              <List.Item
                key={filePreview.id}
                actions={[
                  <Button
                    key="remove"
                    type="text"
                    size="small"
                    icon={<DeleteOutlined />}
                    onClick={() => removeFile(filePreview.id)}
                    danger
                  />,
                ]}
              >
                <List.Item.Meta
                  avatar={
                    filePreview.preview ? (
                      <img
                        src={filePreview.preview}
                        alt={filePreview.file.name}
                        style={{
                          width: 48,
                          height: 48,
                          objectFit: 'cover',
                          borderRadius: 4,
                          border: '1px solid #d9d9d9',
                        }}
                      />
                    ) : (
                      <FileOutlined style={{ fontSize: 24, color: '#8c8c8c' }} />
                    )
                  }
                  title={
                    <div>
                      <Text strong title={filePreview.file.name}>
                        {filePreview.file.name}
                      </Text>
                      <div style={{ marginTop: 4 }}>
                        {getFileTypeTag(filePreview.file)}
                      </div>
                    </div>
                  }
                  description={
                    <Text type="secondary">
                      {formatFileSize(filePreview.file.size)}
                    </Text>
                  }
                />
              </List.Item>
            )}
          />

          {/* 上传按钮 */}
          <div style={{ textAlign: 'center', marginTop: 16 }}>
            <Space>
              <Button
                type="primary"
                size="large"
                icon={<CloudUploadOutlined />}
                onClick={startUpload}
                disabled={selectedFiles.length === 0}
              >
                上传 {selectedFiles.length} 个文件
              </Button>
              <Button
                size="large"
                icon={<PlusOutlined />}
                onClick={selectFiles}
              >
                继续选择
              </Button>
            </Space>
          </div>
        </Card>
      )}
    </div>
  );
};