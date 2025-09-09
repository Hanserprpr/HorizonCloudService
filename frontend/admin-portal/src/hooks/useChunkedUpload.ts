import { useState, useCallback, useRef } from 'react';
import { message } from 'antd';
import { fileService } from '@services/index';
import { useFileStore } from '@stores/fileStore';
import { calculateFileHash } from '@utils/index';
import type { UploadSession } from '@types/index';

// 上传配置
const UPLOAD_CONFIG = {
  chunkSize: 5 * 1024 * 1024, // 5MB per chunk
  maxConcurrent: 3, // 最大并发上传数
  maxRetries: 3, // 最大重试次数
  retryDelay: 1000, // 重试延迟 (ms)
};

// 上传状态
export type UploadStatus = 'pending' | 'uploading' | 'paused' | 'completed' | 'failed' | 'cancelled';

// 上传任务
export interface UploadTask {
  id: string;
  file: File;
  sessionId?: string;
  status: UploadStatus;
  progress: number;
  speed: number; // bytes per second
  uploadedBytes: number;
  totalBytes: number;
  error?: string;
  startTime?: number;
  estimatedTime?: number; // 预估剩余时间(秒)
}

// 分片信息
interface ChunkInfo {
  index: number;
  start: number;
  end: number;
  size: number;
  uploaded: boolean;
  retries: number;
  blob?: Blob;
}

export const useChunkedUpload = () => {
  const [uploadTasks, setUploadTasks] = useState<UploadTask[]>([]);
  const [isUploading, setIsUploading] = useState(false);
  const uploadQueues = useRef<Map<string, ChunkInfo[]>>(new Map());
  const abortControllers = useRef<Map<string, AbortController>>(new Map());
  
  const {
    addUploadSession,
    updateUploadSession,
    removeUploadSession,
    setUploading,
    updateUploadProgress,
    triggerRefresh,
  } = useFileStore();

  // 生成任务ID
  const generateTaskId = () => Date.now().toString(36) + Math.random().toString(36).substr(2);

  // 计算文件分片
  const calculateChunks = (file: File): ChunkInfo[] => {
    const chunks: ChunkInfo[] = [];
    const totalChunks = Math.ceil(file.size / UPLOAD_CONFIG.chunkSize);

    for (let i = 0; i < totalChunks; i++) {
      const start = i * UPLOAD_CONFIG.chunkSize;
      const end = Math.min(start + UPLOAD_CONFIG.chunkSize, file.size);
      
      chunks.push({
        index: i,
        start,
        end,
        size: end - start,
        uploaded: false,
        retries: 0,
      });
    }

    return chunks;
  };

  // 更新任务状态
  const updateTaskStatus = useCallback((taskId: string, updates: Partial<UploadTask>) => {
    setUploadTasks(prev => 
      prev.map(task => 
        task.id === taskId 
          ? { ...task, ...updates }
          : task
      )
    );
  }, []);

  // 计算上传速度和预估时间
  const calculateMetrics = (task: UploadTask, uploadedBytes: number): Partial<UploadTask> => {
    if (!task.startTime) return {};

    const elapsed = (Date.now() - task.startTime) / 1000; // 秒
    const speed = uploadedBytes / elapsed; // bytes per second
    const progress = (uploadedBytes / task.totalBytes) * 100;
    const remainingBytes = task.totalBytes - uploadedBytes;
    const estimatedTime = speed > 0 ? remainingBytes / speed : 0;

    return {
      speed: Math.round(speed),
      progress: Math.min(progress, 100),
      uploadedBytes,
      estimatedTime: Math.round(estimatedTime),
    };
  };

  // 上传单个分片
  const uploadChunk = async (
    taskId: string, 
    sessionId: string, 
    chunk: ChunkInfo, 
    file: File,
    abortController: AbortController
  ): Promise<boolean> => {
    try {
      // 创建分片数据
      chunk.blob = file.slice(chunk.start, chunk.end);
      
      await fileService.uploadChunk({
        session_id: sessionId,
        chunk_index: chunk.index,
        chunk_data: chunk.blob,
      }, {
        signal: abortController.signal,
        onUploadProgress: (progressEvent) => {
          // 实时更新分片进度 (这里可以进一步优化)
        }
      });

      chunk.uploaded = true;
      return true;
    } catch (error: any) {
      if (error.name === 'AbortError') {
        throw error; // 重新抛出取消错误
      }

      chunk.retries++;
      console.warn(`Chunk ${chunk.index} upload failed (attempt ${chunk.retries}):`, error);

      if (chunk.retries < UPLOAD_CONFIG.maxRetries) {
        // 等待重试延迟
        await new Promise(resolve => setTimeout(resolve, UPLOAD_CONFIG.retryDelay * chunk.retries));
        return false; // 需要重试
      }

      throw new Error(`Chunk ${chunk.index} failed after ${UPLOAD_CONFIG.maxRetries} retries: ${error.message}`);
    }
  };

  // 并发上传分片队列
  const uploadChunksInParallel = async (
    taskId: string,
    sessionId: string,
    chunks: ChunkInfo[],
    file: File,
    abortController: AbortController
  ) => {
    const unuploadedChunks = chunks.filter(chunk => !chunk.uploaded);
    let concurrentCount = 0;
    let chunkIndex = 0;

    const uploadNextChunk = async (): Promise<void> => {
      while (chunkIndex < unuploadedChunks.length && concurrentCount < UPLOAD_CONFIG.maxConcurrent) {
        const chunk = unuploadedChunks[chunkIndex++];
        concurrentCount++;

        try {
          let success = false;
          while (!success && chunk.retries < UPLOAD_CONFIG.maxRetries) {
            success = await uploadChunk(taskId, sessionId, chunk, file, abortController);
            
            if (success) {
              // 更新总体进度
              const uploadedBytes = chunks
                .filter(c => c.uploaded)
                .reduce((sum, c) => sum + c.size, 0);

              const task = uploadTasks.find(t => t.id === taskId);
              if (task) {
                const metrics = calculateMetrics(task, uploadedBytes);
                updateTaskStatus(taskId, metrics);
                
                // 更新store进度
                if (task.sessionId) {
                  updateUploadProgress(task.sessionId, metrics.progress || 0);
                }
              }
            }
          }

          if (!success) {
            throw new Error(`Chunk ${chunk.index} failed after all retries`);
          }
        } catch (error) {
          concurrentCount--;
          throw error;
        }

        concurrentCount--;
      }

      // 如果还有未上传的分片，继续
      if (chunkIndex < unuploadedChunks.length) {
        await uploadNextChunk();
      }
    };

    // 启动并发上传
    const promises: Promise<void>[] = [];
    for (let i = 0; i < Math.min(UPLOAD_CONFIG.maxConcurrent, unuploadedChunks.length); i++) {
      promises.push(uploadNextChunk());
    }

    await Promise.all(promises);
  };

  // 开始上传文件
  const startUpload = useCallback(async (
    files: File[], 
    folderId?: number
  ): Promise<void> => {
    if (files.length === 0) return;

    setIsUploading(true);
    setUploading(true);

    const newTasks: UploadTask[] = files.map(file => ({
      id: generateTaskId(),
      file,
      status: 'pending' as UploadStatus,
      progress: 0,
      speed: 0,
      uploadedBytes: 0,
      totalBytes: file.size,
    }));

    setUploadTasks(prev => [...prev, ...newTasks]);

    // 并发处理多个文件
    const uploadPromises = newTasks.map(async (task) => {
      const abortController = new AbortController();
      abortControllers.current.set(task.id, abortController);

      try {
        updateTaskStatus(task.id, { 
          status: 'uploading',
          startTime: Date.now()
        });

        // 1. 初始化上传会话
        const session = await fileService.initiateUpload({
          file_name: task.file.name,
          size: task.file.size,
          content_type: task.file.type || 'application/octet-stream',
          folder_id: folderId,
        });

        // 更新任务和store
        updateTaskStatus(task.id, { sessionId: session.session_id });
        addUploadSession({
          session_id: session.session_id,
          file_name: task.file.name,
          size: task.file.size,
          uploaded_size: 0,
          status: 'uploading',
          created_at: new Date().toISOString(),
        });

        // 2. 计算分片
        const chunks = calculateChunks(task.file);
        uploadQueues.current.set(task.id, chunks);

        // 3. 并发上传分片
        await uploadChunksInParallel(
          task.id, 
          session.session_id, 
          chunks, 
          task.file,
          abortController
        );

        // 4. 完成上传
        const fileResult = await fileService.completeUpload(session.session_id);

        updateTaskStatus(task.id, { 
          status: 'completed',
          progress: 100,
          uploadedBytes: task.file.size,
        });

        updateUploadSession(session.session_id, {
          status: 'completed',
          uploaded_size: task.file.size,
        });

        message.success(`文件 "${task.file.name}" 上传成功`);

      } catch (error: any) {
        console.error('Upload failed for task:', task.id, error);

        if (error.name === 'AbortError') {
          updateTaskStatus(task.id, { status: 'cancelled' });
          message.info(`文件 "${task.file.name}" 上传已取消`);
        } else {
          updateTaskStatus(task.id, { 
            status: 'failed', 
            error: error.message 
          });
          message.error(`文件 "${task.file.name}" 上传失败: ${error.message}`);
        }

        // 清理会话
        if (task.sessionId) {
          updateUploadSession(task.sessionId, { status: 'failed' });
        }
      } finally {
        abortControllers.current.delete(task.id);
        uploadQueues.current.delete(task.id);
      }
    });

    try {
      await Promise.allSettled(uploadPromises);
      
      // 刷新文件列表
      triggerRefresh();
      
    } finally {
      setIsUploading(false);
      setUploading(false);
    }
  }, [updateTaskStatus, addUploadSession, updateUploadSession, updateUploadProgress, setUploading, triggerRefresh]);

  // 暂停上传
  const pauseUpload = useCallback((taskId: string) => {
    const abortController = abortControllers.current.get(taskId);
    if (abortController) {
      abortController.abort();
      updateTaskStatus(taskId, { status: 'paused' });
    }
  }, [updateTaskStatus]);

  // 取消上传
  const cancelUpload = useCallback((taskId: string) => {
    const abortController = abortControllers.current.get(taskId);
    if (abortController) {
      abortController.abort();
    }
    
    updateTaskStatus(taskId, { status: 'cancelled' });
    
    // 清理资源
    const task = uploadTasks.find(t => t.id === taskId);
    if (task?.sessionId) {
      removeUploadSession(task.sessionId);
    }
    
    uploadQueues.current.delete(taskId);
    abortControllers.current.delete(taskId);
  }, [updateTaskStatus, uploadTasks, removeUploadSession]);

  // 重试上传
  const retryUpload = useCallback((taskId: string) => {
    const task = uploadTasks.find(t => t.id === taskId);
    if (task && task.status === 'failed') {
      // 重新开始上传这个文件 - 使用当前文件夹而不是硬编码0
      // 这里可以通过props传入currentFolderId，或者从store获取
      // 暂时设为undefined以上传到根目录
      startUpload([task.file], undefined);
    }
  }, [uploadTasks, startUpload]);

  // 清理已完成的任务
  const clearCompletedTasks = useCallback(() => {
    setUploadTasks(prev => prev.filter(task => 
      task.status !== 'completed' && task.status !== 'cancelled'
    ));
  }, []);

  // 清理所有任务
  const clearAllTasks = useCallback(() => {
    // 取消所有正在进行的上传
    uploadTasks.forEach(task => {
      if (task.status === 'uploading') {
        cancelUpload(task.id);
      }
    });
    
    setUploadTasks([]);
  }, [uploadTasks, cancelUpload]);

  return {
    // 状态
    uploadTasks,
    isUploading,

    // 操作方法
    startUpload,
    pauseUpload,
    cancelUpload,
    retryUpload,
    clearCompletedTasks,
    clearAllTasks,

    // 统计信息
    totalTasks: uploadTasks.length,
    completedTasks: uploadTasks.filter(t => t.status === 'completed').length,
    failedTasks: uploadTasks.filter(t => t.status === 'failed').length,
    activeTasks: uploadTasks.filter(t => t.status === 'uploading').length,
  };
};