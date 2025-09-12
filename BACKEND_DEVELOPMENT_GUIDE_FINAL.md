# AI智能媒体云存储系统 - 后端开发指南

## 项目概览

### 系统定位
AI驱动的智能媒体云存储平台，专为企业级文件管理、智能分析和语义搜索设计。

- **管理端(admin-portal)**: 完整网盘文件管理系统，类似百度网盘、Google Drive
- **用户端(user-gallery)**: 只读图片视频画廊浏览系统，智能展示和搜索
- **AI增强**: 图像识别、语义搜索、智能推荐、自动标签

### 架构特征
- **技术栈**: Go 1.23+ + Gin + GORM + PostgreSQL + Redis + MinIO + Qdrant
- **架构模式**: 微服务架构 + 事件驱动 + API网关
- **部署方式**: Docker + Kubernetes + 完整监控栈
- **AI能力**: CLIP/BLIP模型 + 向量搜索 + 智能标签

### 微服务架构全景 (8个微服务)

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                               客户端层 (Client Layer)                                   │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│    Admin Portal (React)     │    User Gallery (React)    │   Mobile App    │  API Docs │
│        ✅ 90%              │        ❌ 待开发           │     ⏳ 计划     │ 📝 Swagger │
└─────────────────────────────────────────────────────────────────────────────────────────┘
                                              │
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                            API 网关层 (Gateway Layer)                                  │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│       🛡️ Gateway (端口: 8080) - 认证授权 │ 限流 │ 路由 │ 监控 │ 负载均衡               │
│                                🔄 基础版本开发中                                      │
└─────────────────────────────────────────────────────────────────────────────────────────┘
                                              │
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                           业务服务层 (Service Layer)                                    │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│ ┌───────────────┐ ┌───────────────┐ ┌───────────────┐ ┌───────────────┐                │
│ │ User Service  │ │ File Service  │ │System Service │ │ AI Service    │                │
│ │   ✅ 100%    │ │   ✅ 100%    │ │   ✅ 100%    │ │  ❌ 待开发    │                │
│ │   (8001)      │ │   (8083)      │ │   (8003)      │ │  (8084)       │                │
│ └───────────────┘ └───────────────┘ └───────────────┘ └───────────────┘                │
│ ┌───────────────┐ ┌───────────────┐ ┌───────────────┐ ┌───────────────┐                │
│ │Search Service │ │Permission Svc │ │ Share Service │ │Notification   │                │
│ │  ❌ 待开发    │ │  ❌ 待开发    │ │  ❌ 待开发    │ │  ❌ 待开发    │                │
│ │  (8086)       │ │  (8085)       │ │  (8087)       │ │  (8088)       │                │
│ └───────────────┘ └───────────────┘ └───────────────┘ └───────────────┘                │
└─────────────────────────────────────────────────────────────────────────────────────────┘
                                              │
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              数据层 (Data Layer)                                       │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│   PostgreSQL   │ Redis Cluster │  Vector DB   │    MinIO     │  Message Queue         │
│  (文件元数据)  │  (缓存/会话)  │  (Qdrant)    │ (对象存储)   │ (Redis Streams)        │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

## 快速上手 (30分钟)

### 环境准备
```bash
# 克隆项目
git clone <repository-url>
cd HorizonCloudService-main

# 环境要求检查
go version  # 需要: go1.23.0+
node --version  # 需要: v18+
docker --version  # 需要: 20.10+

# 安装Go依赖 (每个服务目录下)
cd services/user-service && go mod tidy
cd ../file-service && go mod tidy  
cd ../system-service && go mod tidy
cd ../gateway && go mod tidy
```

### 启动已完成的核心服务
```bash
# Terminal 1: 启动用户服务
cd services/user-service && go run cmd/main.go
# ✅ 用户服务启动: http://localhost:8001 (默认端口)

# Terminal 2: 启动文件服务  
cd services/file-service && go run test_main_simple.go
# ✅ 文件服务启动: http://localhost:8083

# Terminal 3: 启动系统服务
cd services/system-service && go run cmd/main.go
# ✅ 系统服务启动: http://localhost:8003 (默认端口)

# Terminal 4: 启动前端管理界面
cd frontend/admin-portal && npm install && npm run dev
# ✅ 管理界面启动: http://localhost:5173

# Terminal 5: (可选) 启动API网关
cd services/gateway && go run cmd/main.go
# 🔄 API网关启动: http://localhost:8080
```

### 验证系统状态
```bash
# 健康检查
curl http://localhost:8001/health  # 用户服务
curl http://localhost:8003/health  # 系统服务  
curl http://localhost:8083/health  # 文件服务

# 登录测试 (默认管理员: admin/password123)
curl -X POST http://localhost:8001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}'

# 前端界面验证
echo "访问管理后台: http://localhost:5173"
echo "使用 admin/password123 登录"
```

## 已完成服务详细分析

### **用户服务 (user-service)** - 100%完成

**技术栈**: Go 1.23 + Gin + GORM + bcrypt + JWT
**数据库**: PostgreSQL (生产) / 内存数据库 (测试)
**端口**: 8001 (默认) / 8081 (配置)

**核心功能**:
- JWT认证系统 (HS256签名, 24小时有效期, 刷新令牌机制)
- 用户CRUD管理 (注册、登录、资料管理)  
- 存储配额管理 (默认5GB配额, 管理员可调整)
- 异步活动日志记录 (不阻塞主流程, 完整审计追踪)
- bcrypt密码加密 (cost=12, 企业级安全标准)

**API端点总览** (15个):
```yaml
认证API (4个):
  POST /api/v1/auth/login      # 用户登录
  POST /api/v1/auth/register   # 用户注册  
  POST /api/v1/auth/refresh    # 刷新Token
  POST /api/v1/auth/logout     # 用户登出

用户管理API (4个):
  GET  /api/v1/users/profile   # 获取用户资料
  PUT  /api/v1/users/profile   # 更新用户资料
  PUT  /api/v1/users/password  # 修改密码
  GET  /api/v1/users/quota     # 获取配额信息

管理员API (7个):
  GET    /api/v1/admin/users              # 用户列表 (分页+搜索)
  GET    /api/v1/admin/users/:id          # 用户详情  
  PUT    /api/v1/admin/users/:id          # 更新用户
  DELETE /api/v1/admin/users/:id          # 删除用户
  PUT    /api/v1/admin/users/:id/quota    # 设置用户配额
  GET    /api/v1/admin/activity/:user_id/logs # 用户活动日志
  GET    /api/v1/admin/activity/system     # 系统活动日志
```

**代码架构**: 标准三层架构
- **Repository层**: 数据访问抽象 (支持Mock测试)
- **Service层**: 业务逻辑处理 (JWT生成、密码验证、配额管理)
- **Handler层**: HTTP请求处理 (参数验证、响应格式化)

**测试覆盖**: 
- **总计**: 45个测试用例，100%通过
- **Repository层**: 64.3%覆盖率 (13个测试)
- **Service层**: 55.7%覆盖率 (18个测试)  
- **Handler层**: 9.9%覆盖率 (14个测试)

**API请求示例**:

```bash
# 1. 用户注册
curl -X POST http://localhost:8001/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "confirm_password": "password123"
  }'

# 响应示例
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "role": "user",
    "status": "active",
    "storage_quota": 5368709120,
    "storage_used": 0,
    "created_at": "2025-01-09T10:00:00Z"
  }
}

# 2. 用户登录
curl -X POST http://localhost:8001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'

# 响应示例
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400,
    "token_type": "Bearer",
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "role": "user"
    }
  }
}

# 3. 获取用户资料 (需要JWT认证)
curl -X GET http://localhost:8001/api/v1/users/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 4. 管理员获取用户列表
curl -X GET "http://localhost:8001/api/v1/admin/users?page=1&page_size=10&search=test" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 响应示例
{
  "code": 200,
  "message": "获取成功",
  "data": [
    {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "role": "user",
      "status": "active",
      "storage_quota": 5368709120,
      "storage_used": 1024000,
      "created_at": "2025-01-09T10:00:00Z",
      "last_login": "2025-01-09T15:30:00Z"
    }
  ],
  "page": 1,
  "page_size": 10,
  "total": 1,
  "pages": 1
}
```

### **文件服务 (file-service)** - 100%完成

**技术栈**: Go 1.23 + Gin + GORM + MinIO + 图像处理
**存储**: MinIO (S3兼容) + Local Storage (开发)
**端口**: 8083 (默认) / 8002 (开发环境)

**核心功能**:
- **分片上传系统**: 5MB分片, 断点续传, 并发控制, 失败重试
- **多存储后端**: 统一存储接口, 支持本地存储和MinIO, 可扩展云存储
- **文件去重**: SHA-256指纹计算, 引用计数管理, 存储空间优化
- **缩略图服务**: 多尺寸异步生成 (small/medium/large), 格式转换
- **文件夹管理**: 层级结构, 路径索引, 权限继承, 统计信息
- **权限集成**: JWT认证集成, 用户配额检查, 管理员权限验证

**API端点总览** (29个):
```yaml
文件管理API (11个):
  GET    /api/v1/files              # 文件列表 (分页+筛选)
  POST   /api/v1/files              # 创建文件记录
  GET    /api/v1/files/:id          # 文件详情
  PUT    /api/v1/files/:id          # 更新文件元数据
  DELETE /api/v1/files/:id          # 删除文件 (软删除)
  GET    /api/v1/files/:id/download # 下载文件
  POST   /api/v1/files/:id/copy     # 复制文件
  PUT    /api/v1/files/:id/move     # 移动文件
  GET    /api/v1/files/search       # 搜索文件
  POST   /api/v1/files/batch        # 批量操作
  GET    /api/v1/files/stats        # 用户文件统计

文件夹API (7个):
  GET    /api/v1/folders              # 文件夹列表
  POST   /api/v1/folders              # 创建文件夹
  GET    /api/v1/folders/:id          # 文件夹详情
  PUT    /api/v1/folders/:id          # 更新文件夹
  DELETE /api/v1/folders/:id          # 删除文件夹
  GET    /api/v1/folders/:id/contents # 文件夹内容
  GET    /api/v1/folders/tree         # 文件夹树

分片上传API (7个):
  POST /api/v1/upload/simple          # 简单上传
  POST /api/v1/upload/initiate        # 初始化分片上传
  POST /api/v1/upload/chunk           # 上传分片
  POST /api/v1/upload/complete        # 完成上传
  POST /api/v1/upload/abort           # 取消上传
  GET  /api/v1/upload/status/:session_id # 上传进度
  GET  /api/v1/upload/sessions         # 上传会话列表

缩略图API (4个):
  GET    /api/v1/thumbnails/:file_id          # 获取缩略图
  POST   /api/v1/thumbnails/:file_id/generate # 生成缩略图
  GET    /api/v1/thumbnails/:file_id/list     # 缩略图列表
  DELETE /api/v1/thumbnails/:file_id          # 删除缩略图
```

**存储抽象架构**:
```go
// 统一存储接口设计
type Storage interface {
    Upload(ctx context.Context, key string, data io.Reader, size int64) error
    Download(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    GetURL(ctx context.Context, key string, expires time.Duration) (string, error)
}

// 支持的存储实现
type LocalStorage struct { /* 本地存储实现 */ }
type MinIOStorage struct { /* MinIO S3兼容实现 */ }
// 可扩展: AWSStorage, GCPStorage, AzureStorage
```

**API请求示例**:

```bash
# 1. 获取文件列表
curl -X GET "http://localhost:8083/api/v1/files?page=1&page_size=20&folder_id=1&type=image" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 响应示例
{
  "code": 200,
  "message": "获取成功",
  "data": [
    {
      "id": 1,
      "name": "photo.jpg",
      "size": 1024000,
      "content_type": "image/jpeg",
      "hash": "d41d8cd98f00b204e9800998ecf8427e",
      "folder_id": 1,
      "user_id": 1,
      "download_count": 0,
      "created_at": "2025-01-09T10:00:00Z",
      "updated_at": "2025-01-09T10:00:00Z",
      "thumbnail_url": "/api/v1/thumbnails/1?size=medium",
      "download_url": "/api/v1/files/1/download"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total": 1,
  "pages": 1
}

# 2. 初始化分片上传
curl -X POST http://localhost:8083/api/v1/upload/initiate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "file_name": "large_video.mp4",
    "size": 104857600,
    "content_type": "video/mp4",
    "folder_id": 1
  }'

# 响应示例
{
  "code": 200,
  "message": "初始化成功",
  "data": {
    "session_id": "upload_session_123456",
    "chunk_size": 5242880,
    "total_chunks": 20,
    "expires_at": "2025-01-09T11:00:00Z"
  }
}

# 3. 上传分片
curl -X POST http://localhost:8083/api/v1/upload/chunk \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -F "session_id=upload_session_123456" \
  -F "chunk_index=0" \
  -F "chunk=@chunk_0.bin"

# 4. 完成上传
curl -X POST http://localhost:8083/api/v1/upload/complete \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "session_id": "upload_session_123456"
  }'

# 5. 批量文件操作
curl -X POST http://localhost:8083/api/v1/files/batch \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "operation": "delete",
    "file_ids": [1, 2, 3]
  }'
```

### **系统服务 (system-service)** - 100%完成

**技术栈**: Go 1.23 + Gin + 系统监控
**端口**: 8003 (默认) / 8082 (配置)

**核心功能**:
- **系统统计信息**: 用户数量、文件统计、存储使用情况、活跃度分析
- **健康状态监控**: 服务状态检查、数据库连接、存储可用性验证
- **管理员仪表盘**: 综合统计数据、趋势分析、系统概览

**API端点总览** (6个):
```yaml
系统统计API:
  GET /api/v1/system/stats          # 系统总体统计
  GET /api/v1/system/health         # 系统健康状态
  GET /api/v1/system/version        # 系统版本信息

管理员统计API:
  GET /api/v1/admin/stats/overview  # 管理员仪表盘统计
  GET /api/v1/admin/stats/users     # 用户统计详情
  GET /api/v1/admin/stats/files     # 文件统计详情
```

**API请求示例**:

```bash
# 1. 获取系统总体统计
curl -X GET http://localhost:8003/api/v1/system/stats \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 响应示例
{
  "code": 200,
  "message": "获取成功", 
  "data": {
    "total_users": 1250,
    "total_files": 45672,
    "total_storage_used": "2.5TB",
    "total_storage_quota": "10TB",
    "active_users_24h": 89,
    "uploads_today": 234,
    "system_uptime": "15d 4h 32m",
    "cpu_usage": 23.5,
    "memory_usage": 67.2
  }
}

# 2. 获取系统健康状态
curl -X GET http://localhost:8003/api/v1/system/health \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 响应示例
{
  "code": 200,
  "message": "系统健康",
  "data": {
    "status": "healthy",
    "checks": {
      "database": {
        "status": "ok",
        "response_time": "2ms"
      },
      "redis": {
        "status": "ok",
        "response_time": "1ms"
      },
      "storage": {
        "status": "ok",
        "available_space": "7.5TB"
      }
    },
    "timestamp": "2025-01-09T10:00:00Z"
  }
}

# 3. 管理员仪表盘统计
curl -X GET http://localhost:8003/api/v1/admin/stats/overview \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### **API网关 (gateway)** - 基础版本开发中

**技术栈**: Go 1.23 + Gin + JWT中间件
**端口**: 8080

**当前实现**:
- Gin HTTP路由器和基础反向代理
- JWT认证中间件集成  
- CORS配置和请求头处理
- 基础健康检查和路由转发

**待完善功能**:
- Redis-based限流控制
- 请求监控和Prometheus指标收集
- 负载均衡策略 (轮询、权重、健康检查)
- API版本管理和向后兼容

## 待开发服务规划

### **优先级1: AI服务 (ai-service)** - 核心智能功能

**技术架构**: Go HTTP API + Python ML运行时 (混合架构)
**预估端口**: 8084
**开发工期**: 4周
**团队需求**: 1名Go开发者 + 1名AI工程师

**核心功能需求**:
- **图像内容识别**: 基于CLIP模型的多模态理解, 生成512维语义向量
- **图像质量评估**: 模糊检测、噪点分析、曝光评估、构图分析
- **自动标签生成**: 基于BLIP-2的图像描述生成和标签提取
- **向量嵌入管理**: 向量生成、存储、索引、相似度计算
- **批量处理队列**: Redis Streams异步任务处理, 支持优先级和重试

**API设计规划** (8个端点):
```yaml
图像分析API:
  POST /api/v1/ai/analyze/image        # 分析单张图像
  POST /api/v1/ai/batch-analyze        # 批量图像分析
  GET  /api/v1/ai/analysis/:file_id    # 获取分析结果

向量生成API:  
  POST /api/v1/ai/generate-embedding   # 生成图像向量
  PUT  /api/v1/ai/update-embedding     # 更新向量

标签管理API:
  GET  /api/v1/ai/tags/suggest         # 获取建议标签
  POST /api/v1/ai/tags/custom          # 训练自定义标签

队列管理API:
  GET  /api/v1/ai/queue/status         # 处理队列状态
```

**实现架构**:
```go
// AI服务核心结构
type AIService struct {
    pythonClient *http.Client      // Python ML服务客户端
    redisQueue   *redis.Client     // 异步任务队列
    vectorDB     *qdrant.Client    // 向量数据库连接
    fileService  FileServiceClient // 文件服务集成
}

// 图像分析请求结构
type AnalyzeImageRequest struct {
    FileID       uint     `json:"file_id" binding:"required"`
    ImageURL     string   `json:"image_url" binding:"required"`
    AnalysisType []string `json:"analysis_type"` // ["content", "quality", "embedding"]
    Priority     int      `json:"priority"`      // 1-10 (10最高)
}

// 分析结果响应结构
type AnalyzeImageResponse struct {
    FileID       uint      `json:"file_id"`
    ContentTags  []string  `json:"content_tags"`
    Description  string    `json:"description"`
    QualityScore float64   `json:"quality_score"`
    Embedding    []float32 `json:"embedding,omitempty"`
    ProcessedAt  time.Time `json:"processed_at"`
    Confidence   float64   `json:"confidence"`
}
```

### **优先级2: 搜索服务 (search-service)** - 智能搜索

**技术架构**: Go + Qdrant向量数据库 + Elasticsearch
**预估端口**: 8086
**开发工期**: 2周
**团队需求**: 1名Go开发者

**核心功能需求**:
- **向量相似度搜索**: 基于CLIP向量的图像语义搜索
- **自然语言搜索**: "找出所有日落风景照片" 类型的查询处理
- **以图搜图功能**: 上传图片查找视觉相似内容
- **智能推荐系统**: 基于用户行为和内容相似性的推荐
- **搜索建议**: 自动补全、拼写纠错、热门搜索

**API设计规划** (9个端点):
```yaml
搜索API:
  POST /api/v1/search/semantic         # 语义搜索
  POST /api/v1/search/visual           # 视觉搜索
  POST /api/v1/search/advanced         # 高级筛选搜索
  GET  /api/v1/search/suggestions      # 搜索建议

推荐API:
  GET  /api/v1/recommend/similar       # 相似内容推荐
  GET  /api/v1/recommend/trending      # 热门内容推荐

索引管理API (管理员):
  POST /api/v1/indexes/rebuild         # 重建搜索索引
  GET  /api/v1/indexes/status          # 索引状态
  DELETE /api/v1/indexes/clear         # 清空索引
```

### **优先级3-5: 其他业务服务**

**permission-service (权限服务)**:

- 技术栈: Go + Casbin RBAC引擎 + PostgreSQL
- 开发工期: 3周
- 核心功能: 细粒度权限控制、审计日志、角色管理

**share-service (分享服务)**:
- 技术栈: Go + 加密Token + 访问统计
- 开发工期: 2周  
- 核心功能: 分享链接、权限控制、协作功能

**notification-service (通知服务)**:
- 技术栈: Go + 多通道通知 + 模板引擎
- 开发工期: 1.5周
- 核心功能: 邮件/短信/推送通知、事件驱动

## 开发标准和最佳实践

### Go微服务标准目录结构
```
service-name/
├── cmd/                    # 应用入口点
│   └── main.go            # 服务启动文件
├── internal/              # 私有应用代码  
│   ├── config/            # 配置管理 (viper)
│   ├── handlers/          # HTTP处理器 (Gin)
│   ├── services/          # 业务逻辑层
│   ├── repository/        # 数据访问层 (GORM)
│   ├── models/           # 数据模型 (struct + validation)
│   ├── middleware/       # 中间件 (JWT, CORS, 限流)
│   └── routes/           # 路由配置
├── pkg/                  # 可导出库代码
├── migrations/           # 数据库迁移文件
├── tests/               # 测试文件
├── go.mod              # Go模块定义 (go 1.23)
├── go.sum              # 依赖校验文件  
├── Dockerfile          # 多阶段构建配置
├── docker-compose.yml  # 本地开发配置
└── README.md           # 服务文档
```

### 三层架构实现标准

**1. Handler层** (HTTP接口处理):
```go
type UserHandler struct {
    userService services.UserService
    logger      *zap.Logger
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
        return
    }
    
    user, err := h.userService.CreateUser(c.Request.Context(), &req)
    if err != nil {
        h.logger.Error("创建用户失败", zap.Error(err))
        ErrorResponse(c, http.StatusInternalServerError, "创建用户失败", err.Error())
        return
    }
    
    SuccessResponse(c, http.StatusCreated, "创建成功", user)
}
```

**2. Service层** (业务逻辑处理):
```go
type userService struct {
    userRepo repository.UserRepository
    logger   *zap.Logger
    jwtSecret string
}

func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // 1. 业务逻辑验证
    if err := s.validateCreateUser(req); err != nil {
        return nil, fmt.Errorf("用户验证失败: %w", err)
    }
    
    // 2. 密码加密
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("密码加密失败: %w", err)
    }
    
    // 3. 构造用户对象
    user := &models.User{
        Username:     req.Username,
        Email:        req.Email,  
        Password:     string(hashedPassword),
        Role:         "user",
        Status:       "active",
        StorageQuota: 5 * 1024 * 1024 * 1024, // 5GB默认配额
    }
    
    // 4. 保存到数据库
    return s.userRepo.Create(ctx, user)
}
```

**3. Repository层** (数据访问抽象):
```go
type userRepository struct {
    db *gorm.DB
}

func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
    if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
        if errors.Is(err, gorm.ErrDuplicatedKey) {
            return nil, fmt.Errorf("用户名或邮箱已存在")
        }
        return nil, fmt.Errorf("创建用户失败: %w", err)
    }
    return user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
    var user models.User
    if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("用户不存在")
        }
        return nil, fmt.Errorf("查询用户失败: %w", err)
    }
    return &user, nil
}
```

### API设计规范

**1. RESTful API标准**:
```yaml
资源命名: 使用名词复数形式 (/users, /files, /folders)
HTTP方法: GET(查询), POST(创建), PUT(更新), DELETE(删除)
状态码: 200(成功), 201(创建), 400(参数错误), 401(未授权), 404(不存在), 500(服务错误)
版本控制: URL路径版本 (/api/v1/, /api/v2/)
```

**2. 统一响应格式**:
```go
// 成功响应
type SuccessResponse struct {
    Code    int         `json:"code"`           // 业务状态码 (200表示成功)
    Message string      `json:"message"`        // 提示信息
    Data    interface{} `json:"data,omitempty"` // 响应数据
}

// 错误响应
type ErrorResponse struct {
    Code    int    `json:"code"`    // 业务错误码
    Message string `json:"message"` // 错误信息
    Detail  string `json:"detail,omitempty"`  // 详细错误 (仅开发环境)
}

// 分页响应
type PaginatedResponse struct {
    Code     int         `json:"code"`
    Message  string      `json:"message"`
    Data     interface{} `json:"data"`
    Page     int         `json:"page"`      // 当前页码
    PageSize int         `json:"page_size"` // 页面大小
    Total    int64       `json:"total"`     // 总记录数
    Pages    int64       `json:"pages"`     // 总页数
}

// 统一响应处理函数
func SuccessResponse(c *gin.Context, httpCode int, message string, data interface{}) {
    c.JSON(httpCode, SuccessResponse{
        Code:    200,  // 业务成功码固定为200
        Message: message,
        Data:    data,
    })
}

func ErrorResponse(c *gin.Context, httpCode int, message string, detail string) {
    response := ErrorResponse{
        Code:    httpCode,
        Message: message,
    }
    if gin.Mode() == gin.DebugMode {
        response.Detail = detail  // 仅开发环境显示详细错误
    }
    c.JSON(httpCode, response)
}
```

### 🧪 测试策略和质量保证

**1. 测试层次结构**:
- **单元测试**: >80% 代码覆盖率, 使用testify + mock
- **集成测试**: API端点完整测试, 数据库集成验证
- **E2E测试**: 主要用户流程端到端验证

**2. 单元测试模板**:
```go
func TestUserService_CreateUser(t *testing.T) {
    suite.Run(t, new(UserServiceTestSuite))
}

type UserServiceTestSuite struct {
    suite.Suite
    userService services.UserService
    mockRepo    *mocks.MockUserRepository
}

func (s *UserServiceTestSuite) SetupTest() {
    s.mockRepo = &mocks.MockUserRepository{}
    s.userService = services.NewUserService(s.mockRepo, zap.NewNop(), "test-secret")
}

func (s *UserServiceTestSuite) TestCreateUser_Success() {
    // 准备测试数据
    req := &services.CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
    }
    
    expectedUser := &models.User{
        ID:       1,
        Username: req.Username,
        Email:    req.Email,
        Role:     "user",
        Status:   "active",
    }
    
    // 设置mock期望
    s.mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
        return user.Username == req.Username && user.Email == req.Email
    })).Return(expectedUser, nil)
    
    // 执行测试
    user, err := s.userService.CreateUser(context.Background(), req)
    
    // 验证结果
    s.NoError(err)
    s.Equal(expectedUser.Username, user.Username)
    s.Equal(expectedUser.Email, user.Email)
    s.NotEmpty(user.Password) // 密码应该被加密
    s.mockRepo.AssertExpectations(s.T())
}
```

## 容器化和部署

### **多阶段Dockerfile模板**:
```dockerfile
# 构建阶段
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并构建
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o main cmd/main.go

# 运行阶段
FROM alpine:latest

# 安装必要的CA证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 从构建阶段复制可执行文件
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8081/health || exit 1

# 暴露端口
EXPOSE 8081

# 启动应用
CMD ["./main"]
```

### **Docker Compose开发环境**:
```yaml
version: '3.8'
services:
  # 用户服务
  user-service:
    build: ./services/user-service
    ports:
      - "8001:8001"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=horizon_cloud
      - DB_USER=postgres
      - DB_PASSWORD=password
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - JWT_SECRET=your-jwt-secret-key
    depends_on:
      - postgres
      - redis
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8081/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # 文件服务  
  file-service:
    build: ./services/file-service  
    ports:
      - "8083:8083"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - MINIO_ENDPOINT=minio:9000
      - MINIO_ACCESS_KEY=minioadmin
      - MINIO_SECRET_KEY=12345678
      - STORAGE_BACKEND=minio
    volumes:
      - ./storage:/app/storage
    depends_on:
      - postgres
      - redis
      - minio

  # 系统服务
  system-service:
    build: ./services/system-service
    ports:
      - "8003:8003"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis

  # API网关
  gateway:
    build: ./services/gateway
    ports:
      - "8080:8080"
    environment:
      - USER_SERVICE_URL=http://user-service:8081
      - FILE_SERVICE_URL=http://file-service:8083
      - SYSTEM_SERVICE_URL=http://system-service:8082
    depends_on:
      - user-service
      - file-service
      - system-service

  # 数据库
  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=horizon_cloud
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./shared/database/init.sql:/docker-entrypoint-initdb.d/init.sql

  # 缓存和队列
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  # 对象存储
  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    ports:
      - "9000:9000"  # API端口
      - "9001:9001"  # 管理控制台
    environment:
      - MINIO_ROOT_USER=minioadmin
      - MINIO_ROOT_PASSWORD=12345678
    volumes:
      - minio_data:/data

  # 前端管理界面
  admin-portal:
    build: ./frontend/admin-portal
    ports:
      - "5173:5173"
    environment:
      - VITE_API_BASE_URL=http://localhost:8080
    depends_on:
      - gateway

volumes:
  postgres_data:
  redis_data:
  minio_data:
```

## 环境变量和配置管理

### **核心服务环境变量**

**User Service (.env)**:
```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_NAME=horizon_cloud
DB_USER=postgres
DB_PASSWORD=password
DB_SSLMODE=disable

# Redis配置  
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT配置
JWT_SECRET=horizon-cloud-jwt-secret-key-min-32-chars
JWT_EXPIRES_HOURS=24
REFRESH_TOKEN_EXPIRES_DAYS=7

# 服务配置
PORT=8001
GIN_MODE=release  # debug, release, test
LOG_LEVEL=info
```

**File Service (.env)**:
```bash
# 服务配置
SERVER_PORT=8083
GIN_MODE=release

# 存储配置
STORAGE_BACKEND=local  # local, minio, s3
LOCAL_STORAGE_PATH=/tmp/horizoncloud

# MinIO配置
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=12345678
MINIO_BUCKET=horizoncloud
MINIO_USE_SSL=false

# 上传配置
MAX_FILE_SIZE=1073741824  # 1GB
CHUNK_SIZE=5242880        # 5MB
MAX_CONCURRENT_UPLOADS=10
SUPPORTED_FORMATS=jpg,jpeg,png,gif,mp4,avi,mov,pdf,doc,docx

# 缩略图配置
THUMBNAIL_SIZES=small:200x200,medium:400x400,large:800x800
THUMBNAIL_QUALITY=85

# 用户服务集成
USER_SERVICE_BASE_URL=http://localhost:8001
```

**System Service (.env)**:
```bash
# 服务配置
PORT=8003
GIN_MODE=release

# 系统监控配置
MONITORING_INTERVAL=30s
HEALTH_CHECK_TIMEOUT=10s
SYSTEM_STATS_CACHE_TTL=300s
```

### **Docker Compose 快速启动**

创建项目根目录 `.env` 文件：
```bash
# 数据库配置
POSTGRES_DB=horizon_cloud
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password

# Redis配置
REDIS_PASSWORD=

# MinIO配置
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=12345678

# JWT配置
JWT_SECRET=horizon-cloud-jwt-secret-key-2025
```

启动命令：
```bash
# 一键启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f user-service
docker-compose logs -f file-service
docker-compose logs -f system-service

# 重启特定服务
docker-compose restart user-service

# 停止所有服务
docker-compose down

# 清理数据卷（注意：会删除数据）
docker-compose down -v
```

## 常见问题和故障排除

### **编译和启动问题**

**问题 1: Go模块依赖错误**
```bash
# 症状
go: module example.com/module not found

# 解决方案
cd services/user-service
go mod tidy
go mod download
```

**问题 2: 端口冲突**
```bash
# 症状
bind: address already in use

# 解决方案
# 检查端口占用
lsof -i :8001
# 杀死进程或更改服务端口
export PORT=8011
```

**问题 3: 数据库连接失败**
```bash
# 症状
dial tcp [::1]:5432: connect: connection refused

# 解决方案
# 检查PostgreSQL是否启动
docker-compose ps postgres
# 检查数据库配置
docker-compose logs postgres
# 重启数据库服务
docker-compose restart postgres
```

### **API调用问题**

**问题 4: JWT认证失败**
```bash
# 症状
{"code":401,"message":"未授权访问"}

# 解决方案
# 1. 检查token是否正确
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8001/api/v1/users/profile

# 2. 重新登录获取新token
curl -X POST http://localhost:8001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}'
```

**问题 5: 文件上传失败**
```bash
# 症状
{"code":500,"message":"存储服务不可用"}

# 解决方案
# 检查存储服务状态
docker-compose ps minio
# 检查存储路径权限
ls -la /tmp/horizoncloud
# 检查MinIO配置
curl http://localhost:9001  # MinIO控制台
```

### **性能问题**

**问题 6: 响应时间过慢**
```bash
# 诊断步骤
# 1. 检查数据库连接数
docker-compose exec postgres psql -U postgres -d horizon_cloud -c "SELECT count(*) FROM pg_stat_activity;"

# 2. 检查Redis状态
docker-compose exec redis redis-cli info memory

# 3. 检查服务资源使用
docker stats
```
