import { create } from 'zustand';
import type { FileItem, FolderItem, SearchParams, UploadSession } from '@types/index';

interface FileState {
  // 当前文件夹信息
  currentFolder: FolderItem | null;
  currentFolderId: number;
  setCurrentFolder: (folder: FolderItem | null) => void;
  setCurrentFolderId: (folderId: number) => void;
  
  // 当前路径
  currentPath: string;
  setCurrentPath: (path: string) => void;
  
  // 面包屑导航
  breadcrumb: Array<{ id: number; name: string; path: string }>;
  setBreadcrumb: (breadcrumb: Array<{ id: number; name: string; path: string }>) => void;
  
  // 选中的文件和文件夹（统一为ID数组）
  selectedFiles: FileItem[];
  selectedFolders: FolderItem[];
  selectedFileIds: number[];
  setSelectedFiles: (files: FileItem[]) => void;
  setSelectedFolders: (folders: FolderItem[]) => void;
  setSelectedFileIds: (ids: number[]) => void;
  addSelectedFile: (file: FileItem) => void;
  removeSelectedFile: (fileId: number) => void;
  clearSelection: () => void;
  
  // 视图模式
  viewMode: 'list' | 'grid' | 'detail';
  setViewMode: (mode: 'list' | 'grid' | 'detail') => void;
  
  // 搜索参数
  searchParams: SearchParams;
  setSearchParams: (params: SearchParams) => void;
  clearSearch: () => void;
  
  // 排序参数
  sortBy: string;
  sortOrder: 'asc' | 'desc';
  setSorting: (sortBy: string, order: 'asc' | 'desc') => void;
  
  // 上传会话管理
  uploadSessions: UploadSession[];
  addUploadSession: (session: UploadSession) => void;
  updateUploadSession: (sessionId: string, updates: Partial<UploadSession>) => void;
  removeUploadSession: (sessionId: string) => void;
  clearUploadSessions: () => void;
  
  // 上传队列状态
  isUploading: boolean;
  setUploading: (uploading: boolean) => void;
  uploadProgress: Record<string, number>;
  updateUploadProgress: (sessionId: string, progress: number) => void;
  
  // 文件操作状态
  operationLoading: boolean;
  setOperationLoading: (loading: boolean) => void;
  
  // 预览文件
  previewFile: FileItem | null;
  setPreviewFile: (file: FileItem | null) => void;
  
  // 刷新列表
  refreshKey: number;
  triggerRefresh: () => void;
}

export const useFileStore = create<FileState>((set, get) => ({
  // 初始状态
  currentFolder: null,
  currentFolderId: undefined,
  currentPath: '/',
  breadcrumb: [],
  selectedFiles: [],
  selectedFolders: [],
  selectedFileIds: [],
  viewMode: 'list',
  searchParams: {},
  sortBy: 'name',
  sortOrder: 'asc',
  uploadSessions: [],
  isUploading: false,
  uploadProgress: {},
  operationLoading: false,
  previewFile: null,
  refreshKey: 0,

  // 当前文件夹
  setCurrentFolder: (folder: FolderItem | null) => {
    set({ currentFolder: folder });
  },

  setCurrentFolderId: (folderId: number) => {
    set({ currentFolderId: folderId });
  },

  // 当前路径
  setCurrentPath: (path: string) => {
    set({ currentPath: path });
  },

  // 面包屑导航
  setBreadcrumb: (breadcrumb: Array<{ id: number; name: string; path: string }>) => {
    set({ breadcrumb });
  },

  // 选中文件管理
  setSelectedFiles: (files: FileItem[]) => {
    set({ selectedFiles: files });
  },

  setSelectedFolders: (folders: FolderItem[]) => {
    set({ selectedFolders: folders });
  },

  setSelectedFileIds: (ids: number[]) => {
    set({ selectedFileIds: ids });
  },

  addSelectedFile: (file: FileItem) => {
    const { selectedFiles } = get();
    if (!selectedFiles.find(f => f.id === file.id)) {
      set({ selectedFiles: [...selectedFiles, file] });
    }
  },

  removeSelectedFile: (fileId: number) => {
    const { selectedFiles } = get();
    set({ selectedFiles: selectedFiles.filter(f => f.id !== fileId) });
  },

  clearSelection: () => {
    set({ selectedFiles: [], selectedFolders: [] });
  },

  // 视图模式
  setViewMode: (mode: 'list' | 'grid' | 'detail') => {
    set({ viewMode: mode });
  },

  // 搜索参数
  setSearchParams: (params: SearchParams) => {
    set({ searchParams: params });
  },

  clearSearch: () => {
    set({ searchParams: {} });
  },

  // 排序
  setSorting: (sortBy: string, order: 'asc' | 'desc') => {
    set({ sortBy, sortOrder: order });
  },

  // 上传会话管理
  addUploadSession: (session: UploadSession) => {
    const { uploadSessions } = get();
    set({ uploadSessions: [...uploadSessions, session] });
  },

  updateUploadSession: (sessionId: string, updates: Partial<UploadSession>) => {
    const { uploadSessions } = get();
    set({
      uploadSessions: uploadSessions.map(session =>
        session.session_id === sessionId ? { ...session, ...updates } : session
      ),
    });
  },

  removeUploadSession: (sessionId: string) => {
    const { uploadSessions, uploadProgress } = get();
    const newUploadProgress = { ...uploadProgress };
    delete newUploadProgress[sessionId];
    
    set({
      uploadSessions: uploadSessions.filter(s => s.session_id !== sessionId),
      uploadProgress: newUploadProgress,
    });
  },

  clearUploadSessions: () => {
    set({ uploadSessions: [], uploadProgress: {} });
  },

  // 上传状态
  setUploading: (uploading: boolean) => {
    set({ isUploading: uploading });
  },

  updateUploadProgress: (sessionId: string, progress: number) => {
    const { uploadProgress } = get();
    set({
      uploadProgress: {
        ...uploadProgress,
        [sessionId]: progress,
      },
    });
  },

  // 操作加载状态
  setOperationLoading: (loading: boolean) => {
    set({ operationLoading: loading });
  },

  // 预览文件
  setPreviewFile: (file: FileItem | null) => {
    set({ previewFile: file });
  },

  // 刷新列表
  triggerRefresh: () => {
    set((state) => ({ refreshKey: state.refreshKey + 1 }));
  },
}));