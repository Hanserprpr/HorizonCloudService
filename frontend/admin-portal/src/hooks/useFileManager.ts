import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { message } from 'antd';
import { fileService } from '@services/fileService';
import { useFileStore } from '@stores/fileStore';
import type { 
  FileItem, 
  FolderItem, 
  PaginationParams,
  SearchParams,
  BatchOperation 
} from '../types';

// 获取文件列表
export const useFiles = (params?: PaginationParams & SearchParams & { folder_id?: number }) => {
  return useQuery({
    queryKey: ['files', params],
    queryFn: () => fileService.getFiles(params),
    staleTime: 30 * 1000, // 30秒内数据不过期
    retry: 2,
  });
};

// 获取文件夹列表
export const useFolders = (parentId?: number) => {
  return useQuery({
    queryKey: ['folders', parentId],
    queryFn: () => fileService.getFolders({ parent_id: parentId }),
    staleTime: 60 * 1000, // 1分钟内数据不过期
    retry: 2,
  });
};

// 获取单个文件详情
export const useFile = (fileId: number, enabled = true) => {
  return useQuery({
    queryKey: ['file', fileId],
    queryFn: () => fileService.getFile(fileId),
    enabled: enabled && !!fileId,
    staleTime: 60 * 1000,
    retry: 1,
  });
};

// 获取单个文件夹详情
export const useFolder = (folderId: number, enabled = true) => {
  return useQuery({
    queryKey: ['folder', folderId],
    queryFn: () => fileService.getFolder(folderId),
    enabled: enabled && !!folderId,
    staleTime: 60 * 1000,
    retry: 1,
  });
};

// 获取文件夹内容（文件+子文件夹）
export const useFolderContents = (folderId?: number, params?: SearchParams) => {
  const { refreshKey } = useFileStore();
  
  return useQuery({
    queryKey: ['folder-contents', folderId, params, refreshKey],
    queryFn: () => fileService.getFolderContents(folderId, params),
    staleTime: 30 * 1000,
    retry: 2,
  });
};

// 获取文件夹树
export const useFolderTree = () => {
  return useQuery({
    queryKey: ['folder-tree'],
    queryFn: () => fileService.getFolderTree(),
    staleTime: 5 * 60 * 1000, // 5分钟
    retry: 2,
  });
};

// 创建文件夹
export const useCreateFolder = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: fileService.createFolder,
    onSuccess: (data, variables) => {
      message.success('文件夹创建成功');
      // 更新相关缓存
      queryClient.invalidateQueries({ queryKey: ['folders', variables.parent_id] });
      queryClient.invalidateQueries({ queryKey: ['folder-contents', variables.parent_id] });
      queryClient.invalidateQueries({ queryKey: ['folder-tree'] });
    },
    onError: (error: any) => {
      message.error(error.message || '创建文件夹失败');
    },
  });
};

// 重命名文件
export const useRenameFile = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ fileId, newName }: { fileId: number; newName: string }) =>
      fileService.updateFile(fileId, { file_name: newName }),
    onSuccess: (data) => {
      message.success('文件重命名成功');
      // 更新相关缓存
      queryClient.invalidateQueries({ queryKey: ['file', data.id] });
      queryClient.invalidateQueries({ queryKey: ['files'] });
      queryClient.invalidateQueries({ queryKey: ['folder-contents'] });
    },
    onError: (error: any) => {
      message.error(error.message || '重命名文件失败');
    },
  });
};

// 重命名文件夹
export const useRenameFolder = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ folderId, newName }: { folderId: number; newName: string }) =>
      fileService.updateFolder(folderId, { name: newName }),
    onSuccess: (data) => {
      message.success('文件夹重命名成功');
      // 更新相关缓存
      queryClient.invalidateQueries({ queryKey: ['folder', data.id] });
      queryClient.invalidateQueries({ queryKey: ['folders'] });
      queryClient.invalidateQueries({ queryKey: ['folder-contents'] });
      queryClient.invalidateQueries({ queryKey: ['folder-tree'] });
    },
    onError: (error: any) => {
      message.error(error.message || '重命名文件夹失败');
    },
  });
};

// 删除文件
export const useDeleteFile = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (fileId: number) => {
      console.log('🔥 useDeleteFile.mutationFn 被调用，fileId:', fileId);
      return fileService.deleteFile(fileId);
    },
    onSuccess: () => {
      console.log('✅ useDeleteFile.onSuccess 删除成功');
      message.success('文件删除成功');
      // 更新相关缓存
      queryClient.invalidateQueries({ queryKey: ['files'] });
      queryClient.invalidateQueries({ queryKey: ['folder-contents'] });
    },
    onError: (error: any) => {
      console.log('❌ useDeleteFile.onError 删除失败:', error);
      message.error(error.message || '删除文件失败');
    },
  });
};

// 删除文件夹
export const useDeleteFolder = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: fileService.deleteFolder,
    onSuccess: () => {
      message.success('文件夹删除成功');
      // 更新相关缓存
      queryClient.invalidateQueries({ queryKey: ['folders'] });
      queryClient.invalidateQueries({ queryKey: ['folder-contents'] });
      queryClient.invalidateQueries({ queryKey: ['folder-tree'] });
    },
    onError: (error: any) => {
      message.error(error.message || '删除文件夹失败');
    },
  });
};

// 移动文件
export const useMoveFile = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ fileId, targetFolderId }: { fileId: number; targetFolderId: number }) =>
      fileService.moveFile(fileId, targetFolderId),
    onSuccess: () => {
      message.success('文件移动成功');
      // 更新相关缓存
      queryClient.invalidateQueries({ queryKey: ['files'] });
      queryClient.invalidateQueries({ queryKey: ['folder-contents'] });
    },
    onError: (error: any) => {
      message.error(error.message || '移动文件失败');
    },
  });
};

// 复制文件
export const useCopyFile = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ fileId, targetFolderId, newName }: { 
      fileId: number; 
      targetFolderId?: number; 
      newName?: string;
    }) => fileService.copyFile(fileId, targetFolderId, newName),
    onSuccess: () => {
      message.success('文件复制成功');
      // 更新相关缓存
      queryClient.invalidateQueries({ queryKey: ['files'] });
      queryClient.invalidateQueries({ queryKey: ['folder-contents'] });
    },
    onError: (error: any) => {
      message.error(error.message || '复制文件失败');
    },
  });
};

// 批量操作
export const useBatchOperation = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: fileService.batchOperation,
    onSuccess: (data, variables) => {
      const operationNames = {
        delete: '删除',
        move: '移动',
        copy: '复制',
      };
      const operationName = operationNames[variables.operation] || '操作';
      message.success(`批量${operationName}成功`);
      
      // 更新相关缓存
      queryClient.invalidateQueries({ queryKey: ['files'] });
      queryClient.invalidateQueries({ queryKey: ['folder-contents'] });
      queryClient.invalidateQueries({ queryKey: ['folders'] });
      queryClient.invalidateQueries({ queryKey: ['folder-tree'] });
    },
    onError: (error: any) => {
      message.error(error.message || '批量操作失败');
    },
  });
};

// 获取缩略图
export const useThumbnails = (fileIds: number[]) => {
  return useQuery({
    queryKey: ['thumbnails', fileIds],
    queryFn: () => fileService.getMultipleThumbnails(fileIds),
    enabled: fileIds.length > 0,
    staleTime: 10 * 60 * 1000, // 10分钟
    retry: 1,
  });
};