#!/usr/bin/env node

// 临时文件服务模拟器 - 用于测试API修复
import express from 'express';
import cors from 'cors';
import multer from 'multer';

const app = express();
const port = 8002;

// 中间件
app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// 配置multer用于文件上传
const upload = multer({ dest: 'uploads/' });

console.log('🚀 启动临时文件服务模拟器...');
console.log('📍 端口: 8002');
console.log('🔧 这是一个临时服务，用于测试API修复');

// 健康检查端点
app.get('/api/v1/health', (req, res) => {
  res.json({
    status: 'healthy',
    service: 'mock-file-service',
    message: '临时文件服务模拟器运行正常'
  });
});

app.get('/health', (req, res) => {
  res.json({ status: 'healthy' });
});

app.get('/api/v1/health/ready', (req, res) => {
  res.json({ status: 'ready' });
});

app.get('/api/v1/health/metrics', (req, res) => {
  res.json({
    metrics: 'ok',
    uptime: process.uptime(),
    memory: process.memoryUsage(),
    timestamp: new Date().toISOString()
  });
});

app.get('/api/v1/health/stats', (req, res) => {
  res.json({
    stats: 'ok',
    active_connections: 1,
    total_requests: 100,
    service_status: 'healthy'
  });
});

// 认证中间件模拟
const mockAuth = (req, res, next) => {
  const auth = req.headers.authorization;
  if (!auth || !auth.startsWith('Bearer ')) {
    return res.status(401).json({ error: 'Unauthorized' });
  }
  // 模拟用户ID
  req.user = { id: 1 };
  next();
};

// 文件服务端点
app.get('/api/v1/auth/me', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: { id: req.user.id, name: 'Test User' }
  });
});

app.get('/api/v1/files', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: []
  });
});

app.get('/api/v1/folders', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: []
  });
});

app.get('/api/v1/quota/status', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: {
      used_storage: 0,
      total_storage: 5368709120,
      used_files: 0,
      total_files: 10000,
      usage_percentage: 0
    }
  });
});

app.get('/api/v1/files/search', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: { files: [], total: 0 }
  });
});

app.get('/api/v1/files/stats', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: { total_files: 0, total_size: 0 }
  });
});

app.get('/api/v1/files/storage-stats', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: { used: 0, available: 5368709120 }
  });
});

app.get('/api/v1/files/duplicates', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: []
  });
});

app.post('/api/v1/files/batch', mockAuth, (req, res) => {
  res.json({
    success: true,
    message: 'Batch operation completed',
    data: { processed: 0 }
  });
});

app.post('/api/v1/folders', mockAuth, (req, res) => {
  const timestamp = Date.now();
  res.json({
    success: true,
    data: {
      id: timestamp,
      name: req.body.name,
      created_at: new Date().toISOString()
    }
  });
});

app.get('/api/v1/folders/tree', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: []
  });
});

app.get('/api/v1/folders/search', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: []
  });
});

app.get('/api/v1/folders/by-path', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: { id: 1, name: 'root', path: '/' }
  });
});

app.post('/api/v1/folders/system/create', mockAuth, (req, res) => {
  res.json({
    success: true,
    message: 'System folders created',
    data: {
      created_folders: ['Documents', 'Images', 'Videos', 'Music']
    }
  });
});

// 文件夹详细操作端点
app.get('/api/v1/folders/:id', mockAuth, (req, res) => {
  const folderId = req.params.id;
  res.json({
    success: true,
    data: {
      id: parseInt(folderId),
      name: `测试文件夹_${folderId}`,
      path: `/测试文件夹_${folderId}`,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      size: 0,
      file_count: 0
    }
  });
});

app.get('/api/v1/folders/:id/contents', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: {
      folders: [],
      files: [],
      total_folders: 0,
      total_files: 0
    }
  });
});

app.get('/api/v1/folders/:id/path', mockAuth, (req, res) => {
  const folderId = req.params.id;
  res.json({
    success: true,
    data: {
      path: `/测试文件夹_${folderId}`,
      breadcrumbs: [
        { id: 0, name: 'root', path: '/' },
        { id: parseInt(folderId), name: `测试文件夹_${folderId}`, path: `/测试文件夹_${folderId}` }
      ]
    }
  });
});

app.get('/api/v1/folders/:id/stats', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: {
      total_size: 0,
      file_count: 0,
      folder_count: 0,
      last_modified: new Date().toISOString()
    }
  });
});

app.put('/api/v1/folders/:id', mockAuth, (req, res) => {
  const folderId = req.params.id;
  res.json({
    success: true,
    message: 'Folder updated successfully',
    data: {
      id: parseInt(folderId),
      name: req.body.name || `更新的文件夹_${folderId}`,
      description: req.body.description || '更新的描述',
      updated_at: new Date().toISOString()
    }
  });
});

app.delete('/api/v1/folders/:id', mockAuth, (req, res) => {
  const folderId = req.params.id;
  res.json({
    success: true,
    message: `Folder ${folderId} deleted successfully`,
    data: {
      deleted_folder_id: parseInt(folderId),
      deleted_at: new Date().toISOString()
    }
  });
});

app.get('/api/v1/upload/sessions', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: []
  });
});

app.post('/api/v1/files/upload', mockAuth, upload.single('file'), (req, res) => {
  res.json({
    success: true,
    message: 'File uploaded successfully',
    data: {
      id: Date.now(),
      name: req.file ? req.file.originalname : 'test-file.txt',
      size: req.file ? req.file.size : 100
    }
  });
});

// 简单文件上传端点
app.post('/api/v1/upload/simple', mockAuth, upload.single('file'), (req, res) => {
  res.json({
    success: true,
    message: 'Simple file upload successful',
    data: {
      id: Date.now(),
      name: req.file ? req.file.originalname : 'simple-upload.txt',
      size: req.file ? req.file.size : 100,
      upload_type: 'simple'
    }
  });
});

// 分片上传初始化
app.post('/api/v1/upload/initiate', mockAuth, (req, res) => {
  const sessionId = `session_${Date.now()}`;
  res.json({
    success: true,
    message: 'Chunked upload initiated',
    data: {
      session_id: sessionId,
      chunk_size: 1024 * 1024, // 1MB
      total_chunks: Math.ceil((req.body.file_size || 5000000) / (1024 * 1024))
    }
  });
});

app.post('/api/v1/upload/chunk', mockAuth, upload.single('chunk'), (req, res) => {
  res.json({
    success: true,
    message: 'Chunk uploaded successfully',
    data: {
      chunk_number: req.body.chunk_number || 1,
      session_id: req.body.session_id || 'session_default'
    }
  });
});

app.get('/api/v1/upload/stats', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: { active_uploads: 0, completed_uploads: 0 }
  });
});

// 上传统计信息
app.get('/api/v1/upload/statistics', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: {
      total_uploads: 10,
      successful_uploads: 9,
      failed_uploads: 1,
      total_bytes_uploaded: 1024 * 1024 * 50, // 50MB
      average_upload_speed: 1024 * 1024 * 2 // 2MB/s
    }
  });
});

// 分片上传完成端点
app.post('/api/v1/upload/:sessionId/complete', mockAuth, (req, res) => {
  const sessionId = req.params.sessionId;
  res.json({
    success: true,
    message: 'Chunked upload completed successfully',
    data: {
      session_id: sessionId,
      file_id: Date.now(),
      file_name: req.body.file_name || 'chunked-upload.bin',
      file_size: req.body.file_size || 5000000,
      completed_at: new Date().toISOString()
    }
  });
});

app.get('/api/v1/upload/progress/:sessionId', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: { progress: 100, status: 'completed' }
  });
});

// 文件管理端点
app.get('/api/v1/files/:id', mockAuth, (req, res) => {
  const fileId = req.params.id;
  res.json({
    success: true,
    data: {
      id: parseInt(fileId),
      name: `测试文件_${fileId}.txt`,
      size: 750,
      mime_type: 'text/plain',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      folder_id: null,
      path: `/测试文件_${fileId}.txt`,
      hash: `hash_${fileId}`,
      version: 1
    }
  });
});

app.put('/api/v1/files/:id', mockAuth, (req, res) => {
  const fileId = req.params.id;
  res.json({
    success: true,
    message: 'File updated successfully',
    data: {
      id: parseInt(fileId),
      name: req.body.name || `更新的文件_${fileId}.txt`,
      description: req.body.description || '更新的描述',
      updated_at: new Date().toISOString()
    }
  });
});

app.post('/api/v1/files/:id/copy', mockAuth, (req, res) => {
  const fileId = req.params.id;
  const newFileId = Date.now();
  res.json({
    success: true,
    message: 'File copied successfully',
    data: {
      original_id: parseInt(fileId),
      new_id: newFileId,
      new_name: `副本_测试文件_${fileId}.txt`,
      created_at: new Date().toISOString()
    }
  });
});

app.get('/api/v1/files/:id/versions', mockAuth, (req, res) => {
  const fileId = req.params.id;
  res.json({
    success: true,
    data: {
      versions: [
        {
          version: 1,
          created_at: new Date().toISOString(),
          size: 750,
          hash: `hash_${fileId}_v1`
        }
      ],
      current_version: 1,
      total_versions: 1
    }
  });
});

app.put('/api/v1/files/:id/move', mockAuth, (req, res) => {
  const fileId = req.params.id;
  res.json({
    success: true,
    message: 'File moved successfully',
    data: {
      id: parseInt(fileId),
      old_folder_id: req.body.old_folder_id || null,
      new_folder_id: req.body.new_folder_id || 1,
      moved_at: new Date().toISOString()
    }
  });
});

app.get('/api/v1/files/:id/download', mockAuth, (req, res) => {
  const fileId = req.params.id;
  res.json({
    success: true,
    data: {
      download_url: `http://localhost:8002/download/${fileId}`,
      expires_at: new Date(Date.now() + 3600000).toISOString(), // 1小时后过期
      file_name: `测试文件_${fileId}.txt`
    }
  });
});

app.delete('/api/v1/files/:id', mockAuth, (req, res) => {
  const fileId = req.params.id;
  res.json({
    success: true,
    message: 'File deleted successfully',
    data: {
      deleted_file_id: parseInt(fileId),
      deleted_at: new Date().toISOString()
    }
  });
});

// 缩略图相关端点
app.post('/api/v1/thumbnails/files/:id/generate', mockAuth, (req, res) => {
  const fileId = req.params.id;
  res.json({
    success: true,
    message: 'Thumbnails generated successfully',
    data: {
      file_id: parseInt(fileId),
      thumbnails: [
        { size: 'small', width: 150, height: 150, url: `/thumbnails/${fileId}_small.jpg` },
        { size: 'medium', width: 300, height: 300, url: `/thumbnails/${fileId}_medium.jpg` },
        { size: 'large', width: 600, height: 600, url: `/thumbnails/${fileId}_large.jpg` }
      ],
      generated_at: new Date().toISOString()
    }
  });
});

app.get('/api/v1/thumbnails/files/:id', mockAuth, (req, res) => {
  const fileId = req.params.id;
  res.json({
    success: true,
    data: {
      file_id: parseInt(fileId),
      thumbnails: [
        { size: 'small', width: 150, height: 150, url: `/thumbnails/${fileId}_small.jpg` },
        { size: 'medium', width: 300, height: 300, url: `/thumbnails/${fileId}_medium.jpg` },
        { size: 'large', width: 600, height: 600, url: `/thumbnails/${fileId}_large.jpg` }
      ],
      total_thumbnails: 3
    }
  });
});

app.get('/api/v1/thumbnails/files/:id/info', mockAuth, (req, res) => {
  const fileId = req.params.id;
  res.json({
    success: true,
    data: {
      file_id: parseInt(fileId),
      has_thumbnails: true,
      thumbnail_count: 3,
      supported_sizes: ['small', 'medium', 'large'],
      last_generated: new Date().toISOString()
    }
  });
});

app.get('/api/v1/thumbnails/files/:id/url/:size', mockAuth, (req, res) => {
  const fileId = req.params.id;
  const size = req.params.size;
  res.json({
    success: true,
    data: {
      thumbnail_url: `http://localhost:8002/thumbnails/${fileId}_${size}.jpg`,
      size: size,
      expires_at: new Date(Date.now() + 3600000).toISOString()
    }
  });
});

app.post('/api/v1/thumbnails/files/:id/refresh', mockAuth, (req, res) => {
  const fileId = req.params.id;
  res.json({
    success: true,
    message: 'Thumbnails refreshed successfully',
    data: {
      file_id: parseInt(fileId),
      refreshed_at: new Date().toISOString(),
      thumbnails_count: 3
    }
  });
});

app.get('/api/v1/thumbnails/stats', mockAuth, (req, res) => {
  res.json({
    success: true,
    data: {
      total_thumbnails: 15,
      total_files_with_thumbnails: 5,
      storage_used: 1024 * 1024 * 2, // 2MB
      last_generation: new Date().toISOString()
    }
  });
});

// 启动服务器
app.listen(port, () => {
  console.log(`✅ 临时文件服务模拟器启动成功！`);
  console.log(`🌐 服务地址: http://localhost:${port}`);
  console.log(`🔗 健康检查: http://localhost:${port}/api/v1/health`);
  console.log(`📋 这个模拟器提供了所有必要的API端点来测试修复效果`);
  console.log(`🧪 现在可以运行: node comprehensive-api-test.js`);
});
