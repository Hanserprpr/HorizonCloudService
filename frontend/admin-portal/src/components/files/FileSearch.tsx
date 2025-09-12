import React, { useState } from 'react';
import {
  Input,
  Select,
  Space,
  Button,
  Card,
  Row,
  Col,
  DatePicker,
  Slider,
  Typography,
  Tag,
  Collapse,
} from 'antd';
import {
  SearchOutlined,
  FilterOutlined,
  ClearOutlined,
  CalendarOutlined,
  FileOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import type { RangePickerProps } from 'antd/es/date-picker';
import { formatFileSize } from '@utils/index';

const { RangePicker } = DatePicker;
const { Text } = Typography;
const { Panel } = Collapse;

export interface FileSearchParams {
  keyword?: string;
  file_type?: string;
  content_type?: string;
  min_size?: number;
  max_size?: number;
  date_from?: string;
  date_to?: string;
  is_favorite?: boolean;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface FileSearchProps {
  value?: FileSearchParams;
  onChange?: (params: FileSearchParams) => void;
  onSearch?: (params: FileSearchParams) => void;
  onClear?: () => void;
  loading?: boolean;
}

const FileSearch: React.FC<FileSearchProps> = ({
  value = {},
  onChange,
  onSearch,
  onClear,
  loading = false,
}) => {
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [sizeRange, setSizeRange] = useState<[number, number]>([0, 1000]);

  // 文件类型选项
  const fileTypeOptions = [
    { label: '全部类型', value: '' },
    { label: '图片', value: 'image' },
    { label: '视频', value: 'video' },
    { label: '音频', value: 'audio' },
    { label: '文档', value: 'document' },
    { label: '压缩包', value: 'archive' },
    { label: '代码', value: 'code' },
    { label: '其他', value: 'other' },
  ];

  // 排序选项
  const sortOptions = [
    { label: '修改时间（新到旧）', value: 'updated_at_desc' },
    { label: '修改时间（旧到新）', value: 'updated_at_asc' },
    { label: '文件名（A-Z）', value: 'name_asc' },
    { label: '文件名（Z-A）', value: 'name_desc' },
    { label: '文件大小（小到大）', value: 'size_asc' },
    { label: '文件大小（大到小）', value: 'size_desc' },
  ];

  // 更新搜索参数
  const updateParams = (newParams: Partial<FileSearchParams>) => {
    const updated = { ...value, ...newParams };
    onChange?.(updated);
  };

  // 处理搜索
  const handleSearch = () => {
    onSearch?.(value);
  };

  // 处理清空
  const handleClear = () => {
    onChange?.({});
    onClear?.();
  };

  // 处理日期范围变化
  const handleDateRangeChange: RangePickerProps['onChange'] = (dates) => {
    if (dates && dates[0] && dates[1]) {
      updateParams({
        date_from: dates[0].format('YYYY-MM-DD'),
        date_to: dates[1].format('YYYY-MM-DD'),
      });
    } else {
      updateParams({
        date_from: undefined,
        date_to: undefined,
      });
    }
  };

  // 处理文件大小范围变化
  const handleSizeRangeChange = (range: [number, number]) => {
    setSizeRange(range);
    updateParams({
      min_size: range[0] * 1024 * 1024, // 转换为字节
      max_size: range[1] * 1024 * 1024,
    });
  };

  // 处理排序变化
  const handleSortChange = (sortValue: string) => {
    if (!sortValue || typeof sortValue !== 'string') return;
    const [sort_by, sort_order] = sortValue.split('_').slice(-2);
    updateParams({
      sort_by: sortValue.replace(`_${sort_order}`, ''),
      sort_order: sort_order as 'asc' | 'desc',
    });
  };

  return (
    <Card size="small">
      {/* 基本搜索 */}
      <Row gutter={[16, 16]} align="middle">
        <Col flex="1">
          <Input
            placeholder="搜索文件名..."
            prefix={<SearchOutlined />}
            value={value.keyword}
            onChange={(e) => updateParams({ keyword: e.target.value })}
            onPressEnter={handleSearch}
            allowClear
          />
        </Col>
        
        <Col>
          <Select
            style={{ width: 120 }}
            placeholder="文件类型"
            value={value.file_type}
            onChange={(fileType) => updateParams({ file_type: fileType })}
            options={fileTypeOptions}
            allowClear
          />
        </Col>

        <Col>
          <Select
            style={{ width: 140 }}
            placeholder="排序方式"
            value={value.sort_by && value.sort_order ? `${value.sort_by}_${value.sort_order}` : undefined}
            onChange={handleSortChange}
            options={sortOptions}
          />
        </Col>

        <Col>
          <Space>
            <Button
              type="primary"
              icon={<SearchOutlined />}
              onClick={handleSearch}
              loading={loading}
            >
              搜索
            </Button>
            
            <Button
              icon={<FilterOutlined />}
              onClick={() => setShowAdvanced(!showAdvanced)}
            >
              高级
            </Button>
            
            <Button
              icon={<ClearOutlined />}
              onClick={handleClear}
            >
              清空
            </Button>
          </Space>
        </Col>
      </Row>

      {/* 高级搜索 */}
      {showAdvanced && (
        <div style={{ marginTop: 16, padding: '16px 0', borderTop: '1px solid #f0f0f0' }}>
          <Collapse ghost>
            <Panel 
              header={
                <Space>
                  <FilterOutlined />
                  <Text strong>高级筛选</Text>
                </Space>
              } 
              key="advanced"
            >
              <Row gutter={[16, 16]}>
                {/* 日期范围 */}
                <Col xs={24} md={12}>
                  <Space direction="vertical" size={4} style={{ width: '100%' }}>
                    <Text strong>修改时间</Text>
                    <RangePicker
                      style={{ width: '100%' }}
                      onChange={handleDateRangeChange}
                      placeholder={['开始日期', '结束日期']}
                      value={
                        value.date_from && value.date_to
                          ? [dayjs(value.date_from), dayjs(value.date_to)]
                          : null
                      }
                    />
                  </Space>
                </Col>

                {/* 文件大小 */}
                <Col xs={24} md={12}>
                  <Space direction="vertical" size={4} style={{ width: '100%' }}>
                    <Text strong>文件大小 (MB)</Text>
                    <Slider
                      range
                      min={0}
                      max={1000}
                      value={sizeRange}
                      onChange={handleSizeRangeChange}
                      marks={{
                        0: '0MB',
                        100: '100MB',
                        500: '500MB',
                        1000: '1GB',
                      }}
                      tooltip={{
                        formatter: (value) => `${value}MB`,
                      }}
                    />
                    <Row justify="space-between">
                      <Text type="secondary" style={{ fontSize: '12px' }}>
                        最小: {formatFileSize(sizeRange[0] * 1024 * 1024)}
                      </Text>
                      <Text type="secondary" style={{ fontSize: '12px' }}>
                        最大: {formatFileSize(sizeRange[1] * 1024 * 1024)}
                      </Text>
                    </Row>
                  </Space>
                </Col>

                {/* 内容类型 */}
                <Col xs={24} md={12}>
                  <Space direction="vertical" size={4} style={{ width: '100%' }}>
                    <Text strong>内容类型</Text>
                    <Select
                      style={{ width: '100%' }}
                      placeholder="选择具体类型"
                      value={value.content_type}
                      onChange={(contentType) => updateParams({ content_type: contentType })}
                      allowClear
                      showSearch
                      options={[
                        { label: 'image/jpeg', value: 'image/jpeg' },
                        { label: 'image/png', value: 'image/png' },
                        { label: 'image/gif', value: 'image/gif' },
                        { label: 'video/mp4', value: 'video/mp4' },
                        { label: 'video/avi', value: 'video/avi' },
                        { label: 'audio/mp3', value: 'audio/mp3' },
                        { label: 'audio/wav', value: 'audio/wav' },
                        { label: 'application/pdf', value: 'application/pdf' },
                        { label: 'text/plain', value: 'text/plain' },
                      ]}
                    />
                  </Space>
                </Col>

                {/* 收藏状态 */}
                <Col xs={24} md={12}>
                  <Space direction="vertical" size={4} style={{ width: '100%' }}>
                    <Text strong>收藏状态</Text>
                    <Select
                      style={{ width: '100%' }}
                      placeholder="选择收藏状态"
                      value={value.is_favorite}
                      onChange={(isFavorite) => updateParams({ is_favorite: isFavorite })}
                      allowClear
                      options={[
                        { label: '全部', value: undefined },
                        { label: '已收藏', value: true },
                        { label: '未收藏', value: false },
                      ]}
                    />
                  </Space>
                </Col>
              </Row>

              <Row style={{ marginTop: 16 }}>
                <Col span={24}>
                  <Space>
                    <Button type="primary" onClick={handleSearch} loading={loading}>
                      应用筛选
                    </Button>
                    <Button onClick={handleClear}>
                      重置筛选
                    </Button>
                  </Space>
                </Col>
              </Row>
            </Panel>
          </Collapse>
        </div>
      )}

      {/* 当前筛选条件显示 */}
      {(value.keyword || value.file_type || value.date_from || value.min_size) && (
        <div style={{ marginTop: 12, padding: '8px 0', borderTop: '1px solid #f0f0f0' }}>
          <Space wrap>
            <Text type="secondary" style={{ fontSize: '12px' }}>当前筛选:</Text>
            
            {value.keyword && (
              <Tag 
                closable 
                onClose={() => updateParams({ keyword: undefined })}
                icon={<SearchOutlined />}
              >
                关键词: {value.keyword}
              </Tag>
            )}
            
            {value.file_type && (
              <Tag 
                closable 
                onClose={() => updateParams({ file_type: undefined })}
                icon={<FileOutlined />}
              >
                类型: {fileTypeOptions.find(opt => opt.value === value.file_type)?.label}
              </Tag>
            )}
            
            {value.date_from && value.date_to && (
              <Tag 
                closable 
                onClose={() => updateParams({ date_from: undefined, date_to: undefined })}
                icon={<CalendarOutlined />}
              >
                时间: {value.date_from} ~ {value.date_to}
              </Tag>
            )}
            
            {value.min_size && value.max_size && (
              <Tag 
                closable 
                onClose={() => updateParams({ min_size: undefined, max_size: undefined })}
              >
                大小: {formatFileSize(value.min_size)} ~ {formatFileSize(value.max_size)}
              </Tag>
            )}
          </Space>
        </div>
      )}
    </Card>
  );
};

export default FileSearch;