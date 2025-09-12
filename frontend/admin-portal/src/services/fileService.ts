import { apiClient } from './api';
import { FILE_SERVICE_URL } from '@constants/index';

// 创建专用的文件服务API客户端
import { ApiClient } from './api';
const fileApiClient = new ApiClient(FILE_SERVICE_URL);
import type { 
  FileItem, 
  FolderItem, 
  UploadSession, 
  PaginatedResponse, 
  PaginationParams,
  SearchParams,
  Thumbnail,
  BatchOperation
} from '../types';

export class FileService {
  // ============ 文件管理 ============
  
  // 获取文件列表
  static async getFiles(params?: PaginationParams & SearchParams & {
    folder_id?: number;
  }): Promise<PaginatedResponse<FileItem>> {
    const response = await fileApiClient.get<PaginatedResponse<FileItem>>('/api/v1/files', params);
    return response.data;
  }

  // 获取文件详情
  static async getFile(fileId: number): Promise<FileItem> {
    const response = await fileApiClient.get<FileItem>(`/api/v1/files/${fileId}`);
    return response.data;
  }

  // 更新文件信息
  static async updateFile(fileId: number, data: Partial<FileItem>): Promise<FileItem> {
    const response = await fileApiClient.put<FileItem>(`/api/v1/files/${fileId}`, data);
    return response.data;
  }

  // 删除文件
  static async deleteFile(fileId: number): Promise<void> {
    console.log('🌐 FileService.deleteFile 发起API调用，fileId:', fileId);
    console.log('🌐 API URL:', `/api/v1/files/${fileId}`);
    try {
      const response = await fileApiClient.delete(`/api/v1/files/${fileId}`);
      console.log('🌐 FileService.deleteFile API调用成功:', response);
      return response.data;
    } catch (error) {
      console.log('🌐 FileService.deleteFile API调用失败:', error);
      throw error;
    }
  }

  // 批量删除文件
  static async deleteFiles(fileIds: number[]): Promise<void> {
    await fileApiClient.post('/api/v1/files/batch-delete', { file_ids: fileIds });
  }

  // 移动文件
  static async moveFile(fileId: number, targetFolderId?: number): Promise<FileItem> {
    const response = await fileApiClient.post<FileItem>(`/api/v1/files/${fileId}/move`, {
      folder_id: targetFolderId,
    });
    return response.data;
  }

  // 复制文件
  static async copyFile(fileId: number, targetFolderId?: number, newName?: string): Promise<FileItem> {
    const requestData: any = {};
    if (targetFolderId !== undefined) {
      requestData.folder_id = targetFolderId;
    }
    if (newName) {
      requestData.new_name = newName;
    }
    
    const response = await fileApiClient.post<FileItem>(`/api/v1/files/${fileId}/copy`, requestData);
    return response.data;
  }

  // 批量操作
  static async batchOperation(operation: BatchOperation): Promise<void> {
    await fileApiClient.post('/api/v1/files/batch-operation', operation);
  }

  // 获取下载链接
  static async getDownloadUrl(fileId: number): Promise<{ download_url: string }> {
    const response = await fileApiClient.get<{ download_url: string }>(`/api/v1/files/${fileId}/download`);
    return response.data;
  }

  // 下载文件
  static async downloadFile(fileId: number, filename?: string): Promise<void> {
    return fileApiClient.download(`/api/v1/files/${fileId}/download`, filename);
  }

  // ============ 文件夹管理 ============

  // 获取文件夹列表
  static async getFolders(params?: {
    parent_id?: number;
    user_id?: number;
  } & PaginationParams): Promise<PaginatedResponse<FolderItem>> {
    const response = await fileApiClient.get<PaginatedResponse<FolderItem>>('/api/v1/folders', params);
    return response.data;
  }

  // 创建文件夹
  static async createFolder(data: {
    name: string;
    description?: string;
    parent_id?: number;
  }): Promise<FolderItem> {
    const response = await fileApiClient.post<FolderItem>('/api/v1/folders', data);
    return response.data;
  }

  // 更新文件夹
  static async updateFolder(folderId: number, data: Partial<FolderItem>): Promise<FolderItem> {
    const response = await fileApiClient.put<FolderItem>(`/api/v1/folders/${folderId}`, data);
    return response.data;
  }

  // 删除文件夹
  static async deleteFolder(folderId: number): Promise<void> {
    await fileApiClient.delete(`/api/v1/folders/${folderId}`);
  }

  // 获取单个文件夹详情
  static async getFolder(folderId: number): Promise<FolderItem> {
    const response = await fileApiClient.get<FolderItem>(`/api/v1/folders/${folderId}`);
    return response.data;
  }

  // 获取文件夹内容
  static async getFolderContents(folderId?: number, params?: PaginationParams): Promise<{
    files: FileItem[];
    folders: FolderItem[];
    total_files: number;
    total_folders: number;
  }> {
    const effectiveFolderId = folderId || 0;
    const response = await fileApiClient.get(`/api/v1/folders/${effectiveFolderId}/contents`, params);
    return response.data;
  }

  // 获取文件夹树
  static async getFolderTree(userId?: number): Promise<FolderItem[]> {
    const response = await fileApiClient.get<FolderItem[]>('/api/v1/folders/tree', { user_id: userId });
    return response.data;
  }

  // ============ 上传管理 ============

  // 简单上传
  static async uploadFile(file: File, folderId?: number, onProgress?: (progress: number) => void): Promise<FileItem> {
    const formData = new FormData();
    formData.append('file', file);
    if (folderId) {
      formData.append('folder_id', folderId.toString());
    }

    const response = await fileApiClient.upload<FileItem>('/api/v1/upload/simple', formData, onProgress);
    return response.data;
  }

  // 初始化分片上传
  static async initiateUpload(data: {
    file_name: string;
    size: number;
    content_type: string;
    folder_id?: number;
  }): Promise<UploadSession> {
    // 构造请求数据，如果folder_id未定义则不包含该字段
    const requestData: any = {
      file_name: data.file_name,
      size: data.size,
      content_type: data.content_type
    };
    
    // 只有当folder_id存在且不为undefined时才添加该字段
    if (data.folder_id !== undefined) {
      requestData.folder_id = data.folder_id;
    }
    
    const response = await fileApiClient.post<UploadSession>('/api/v1/upload/initiate', requestData);
    return response.data;
  }

  // 上传分片
  static async uploadChunk(data: {
    session_id: string;
    chunk_index: number;
    chunk_data: Blob;
  }, options?: {
    signal?: AbortSignal;
    onUploadProgress?: (progressEvent: any) => void;
  }): Promise<void> {
    const formData = new FormData();
    formData.append('session_id', data.session_id);
    formData.append('chunk_index', data.chunk_index.toString());
    formData.append('chunk', data.chunk_data);

    await fileApiClient.upload('/api/v1/upload/chunk', formData, undefined, {
      signal: options?.signal,
      onUploadProgress: options?.onUploadProgress,
    });
  }

  // 完成上传
  static async completeUpload(sessionId: string): Promise<FileItem> {
    const response = await fileApiClient.post<FileItem>(`/api/v1/upload/complete/${sessionId}`, {});
    return response.data;
  }

  // 取消上传
  static async cancelUpload(sessionId: string): Promise<void> {
    await fileApiClient.post(`/api/v1/upload/abort/${sessionId}`, {});
  }

  // 获取上传进度
  static async getUploadProgress(sessionId: string): Promise<UploadSession> {
    const response = await fileApiClient.get<UploadSession>(`/api/v1/upload/progress/${sessionId}`);
    return response.data;
  }

  // ============ 缩略图管理 ============

  // 获取单个文件的缩略图列表
  static async getThumbnails(fileId: number): Promise<Thumbnail[]> {
    const response = await fileApiClient.get<Thumbnail[]>(`/api/v1/thumbnails/${fileId}`);
    return response.data;
  }

  // 获取多个文件的缩略图列表
  static async getMultipleThumbnails(fileIds: number[]): Promise<Record<number, Thumbnail[]>> {
    const promises = fileIds.map(async (fileId) => {
      try {
        const thumbnails = await this.getThumbnails(fileId);
        return { fileId, thumbnails };
      } catch (error) {
        return { fileId, thumbnails: [] };
      }
    });
    
    const results = await Promise.all(promises);
    return results.reduce((acc, { fileId, thumbnails }) => {
      acc[fileId] = thumbnails;
      return acc;
    }, {} as Record<number, Thumbnail[]>);
  }

  // 生成缩略图
  static async generateThumbnail(fileId: number, sizes: string[]): Promise<Thumbnail[]> {
    const response = await fileApiClient.post<Thumbnail[]>(`/api/v1/thumbnails/${fileId}/generate`, {
      sizes,
    });
    return response.data;
  }

  // 删除缩略图
  static async deleteThumbnail(fileId: number): Promise<void> {
    await fileApiClient.delete(`/api/v1/thumbnails/${fileId}`);
  }

  // ============ 搜索功能 ============

  // 搜索文件
  static async searchFiles(params: SearchParams & PaginationParams): Promise<PaginatedResponse<FileItem>> {
    const response = await fileApiClient.get<PaginatedResponse<FileItem>>('/api/v1/files/search', params);
    return response.data;
  }

  // ============ 统计信息 ============

  // 获取文件统计
  static async getFileStats(): Promise<{
    total_files: number;
    total_size: number;
    files_by_type: Record<string, number>;
    storage_by_user: Array<{ user_id: number; username: string; storage_used: number }>;
  }> {
    const response = await fileApiClient.get('/api/v1/files/stats/storage');
    return response.data;
  }
}

// 导出服务类
export const fileService = FileService;