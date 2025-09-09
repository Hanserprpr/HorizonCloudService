import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { message } from 'antd';
import { fileService as FileService } from '@services/fileService';
import { queryKeys } from '@lib/queryClient';
import { SUCCESS_MESSAGES } from '@constants/index';
import { useFileStore } from '@stores/fileStore';
import type { FileItem, FolderItem, PaginationParams, SearchParams, BatchOperation } from '@types/index';

// 文件列表hook
export const useFiles = (params?: PaginationParams & SearchParams & {
  folder_id?: number;
}) => {
  return useQuery({
    queryKey: queryKeys.files.list(params || {}),
    queryFn: () => FileService.getFiles(params),
    staleTime: 30 * 1000, // 30秒
  });
};

// 单个文件详情hook
export const useFile = (fileId: number, enabled: boolean = true) => {
  return useQuery({
    queryKey: queryKeys.files.detail(fileId),
    queryFn: () => FileService.getFile(fileId),
    enabled: enabled && !!fileId,
  });
};

// 文件夹列表hook
export const useFolders = (params?: {
  parent_id?: number;
  user_id?: number;
} & PaginationParams) => {
  return useQuery({
    queryKey: queryKeys.folders.list(params || {}),
    queryFn: () => FileService.getFolders(params),
    staleTime: 2 * 60 * 1000, // 2分钟
  });
};

// 文件夹内容hook
export const useFolderContents = (folderId: number, params?: PaginationParams, enabled: boolean = true) => {
  return useQuery({
    queryKey: queryKeys.files.folderContents(folderId),
    queryFn: () => FileService.getFolderContents(folderId, params),
    enabled: enabled && !!folderId,
    staleTime: 30 * 1000,
  });
};

// 文件夹树hook
export const useFolderTree = (userId?: number) => {
  return useQuery({
    queryKey: queryKeys.folders.tree(),
    queryFn: () => FileService.getFolderTree(userId),
    staleTime: 5 * 60 * 1000, // 5分钟
  });
};

// 文件搜索hook
export const useSearchFiles = (params: SearchParams & PaginationParams, enabled: boolean = true) => {
  return useQuery({
    queryKey: queryKeys.files.search(params),
    queryFn: () => FileService.searchFiles(params),
    enabled: enabled && !!(params.keyword || params.type || params.extension),
    staleTime: 60 * 1000, // 1分钟
  });
};

// 文件统计hook
export const useFileStats = () => {
  return useQuery({
    queryKey: queryKeys.files.stats(),
    queryFn: () => FileService.getFileStats(),
    staleTime: 5 * 60 * 1000, // 5分钟
  });
};

// 缩略图hook
export const useThumbnails = (fileId: number, enabled: boolean = true) => {
  return useQuery({
    queryKey: queryKeys.thumbnails.list(fileId),
    queryFn: () => FileService.getThumbnails(fileId),
    enabled: enabled && !!fileId,
    staleTime: 10 * 60 * 1000, // 10分钟
  });
};

// 文件操作hooks
export const useFileMutations = () => {
  const queryClient = useQueryClient();
  const { triggerRefresh, addUploadSession, updateUploadSession, removeUploadSession } = useFileStore();

  // 简单文件上传
  const uploadFile = useMutation({
    mutationFn: ({ file, folderId, onProgress }: { 
      file: File; 
      folderId?: number; 
      onProgress?: (progress: number) => void 
    }) => FileService.uploadFile(file, folderId, onProgress),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.files.lists() });
      queryClient.invalidateQueries({ queryKey: queryKeys.files.stats() });
      message.success(SUCCESS_MESSAGES.UPLOAD_SUCCESS);
      triggerRefresh();
    },
  });

  // 分片上传 - 初始化
  const initiateUpload = useMutation({
    mutationFn: FileService.initiateUpload,
    onSuccess: (session) => {
      addUploadSession(session);
    },
  });

  // 分片上传 - 上传分片
  const uploadChunk = useMutation({
    mutationFn: FileService.uploadChunk,
    onSuccess: (_, variables) => {
      // 更新上传进度（这里需要计算进度）
      const progress = Math.round(((variables.chunk_index + 1) / 10) * 100); // 假设10个分片
      updateUploadSession(variables.session_id, { progress });
    },
  });

  // 分片上传 - 完成上传
  const completeUpload = useMutation({
    mutationFn: FileService.completeUpload,
    onSuccess: (file, sessionId) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.files.lists() });
      queryClient.invalidateQueries({ queryKey: queryKeys.files.stats() });
      removeUploadSession(sessionId);
      message.success(SUCCESS_MESSAGES.UPLOAD_SUCCESS);
      triggerRefresh();
    },
  });

  // 取消上传
  const cancelUpload = useMutation({
    mutationFn: FileService.cancelUpload,
    onSuccess: (_, sessionId) => {
      removeUploadSession(sessionId);
      message.info('上传已取消');
    },
  });

  // 更新文件
  const updateFile = useMutation({
    mutationFn: ({ fileId, data }: { fileId: number; data: Partial<FileItem> }) =>
      FileService.updateFile(fileId, data),
    onSuccess: (updatedFile) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.files.lists() });
      queryClient.setQueryData(queryKeys.files.detail(updatedFile.id), updatedFile);
      message.success(SUCCESS_MESSAGES.UPDATE_SUCCESS);
      triggerRefresh();
    },
  });

  // 删除文件
  const deleteFile = useMutation({
    mutationFn: FileService.deleteFile,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.files.lists() });
      queryClient.invalidateQueries({ queryKey: queryKeys.files.stats() });
      message.success(SUCCESS_MESSAGES.DELETE_SUCCESS);
      triggerRefresh();
    },
  });

  // 批量删除文件
  const deleteFiles = useMutation({
    mutationFn: FileService.deleteFiles,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.files.lists() });
      queryClient.invalidateQueries({ queryKey: queryKeys.files.stats() });
      message.success(SUCCESS_MESSAGES.DELETE_SUCCESS);
      triggerRefresh();
    },
  });

  // 移动文件
  const moveFile = useMutation({
    mutationFn: ({ fileId, targetFolderId }: { fileId: number; targetFolderId?: number }) =>
      FileService.moveFile(fileId, targetFolderId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.files.lists() });
      message.success('文件移动成功');
      triggerRefresh();
    },
  });

  // 复制文件
  const copyFile = useMutation({
    mutationFn: ({ fileId, targetFolderId }: { fileId: number; targetFolderId?: number }) =>
      FileService.copyFile(fileId, targetFolderId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.files.lists() });
      message.success('文件复制成功');
      triggerRefresh();
    },
  });

  // 批量操作
  const batchOperation = useMutation({
    mutationFn: FileService.batchOperation,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.files.lists() });
      message.success('批量操作完成');
      triggerRefresh();
    },
  });

  // 创建文件夹
  const createFolder = useMutation({
    mutationFn: FileService.createFolder,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.folders.lists() });
      queryClient.invalidateQueries({ queryKey: queryKeys.folders.tree() });
      message.success('文件夹创建成功');
      triggerRefresh();
    },
  });

  // 更新文件夹
  const updateFolder = useMutation({
    mutationFn: ({ folderId, data }: { folderId: number; data: Partial<FolderItem> }) =>
      FileService.updateFolder(folderId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.folders.lists() });
      queryClient.invalidateQueries({ queryKey: queryKeys.folders.tree() });
      message.success('文件夹更新成功');
      triggerRefresh();
    },
  });

  // 删除文件夹
  const deleteFolder = useMutation({
    mutationFn: FileService.deleteFolder,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.folders.lists() });
      queryClient.invalidateQueries({ queryKey: queryKeys.folders.tree() });
      message.success('文件夹删除成功');
      triggerRefresh();
    },
  });

  // 生成缩略图
  const generateThumbnail = useMutation({
    mutationFn: ({ fileId, sizes }: { fileId: number; sizes: string[] }) =>
      FileService.generateThumbnail(fileId, sizes),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.thumbnails.list(variables.fileId) });
      message.success('缩略图生成成功');
    },
  });

  return {
    uploadFile,
    initiateUpload,
    uploadChunk,
    completeUpload,
    cancelUpload,
    updateFile,
    deleteFile,
    deleteFiles,
    moveFile,
    copyFile,
    batchOperation,
    createFolder,
    updateFolder,
    deleteFolder,
    generateThumbnail,
  };
};