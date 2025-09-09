# File Service - 文件管理服务

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Test Status](https://img.shields.io/badge/tests-passing-brightgreen.svg)](#测试)

文件管理服务是一个基于 Go 语言开发的高性能文件存储和管理微服务，支持大文件分片上传、多存储后端、智能缩略图生成、用户配额管理等企业级功能。

## ✨ 特性

### 🚀 核心功能
- **多种上传方式**: 支持简单上传和分片上传，适应不同文件大小
- **多存储后端**: 支持 MinIO、AWS S3 等对象存储，可扩展本地存储
- **文件去重**: 基于 SHA-256 哈希的智能文件去重
- **版本控制**: 完整的文件版本管理和历史记录
- **文件夹管理**: 支持层级文件夹结构和批量操作

### 🎯 企业级功能
- **用户隔离**: 完整的多用户数据隔离和权限控制
- **配额管理**: 灵活的存储配额和文件数量限制
- **缩略图服务**: 自动生成多尺寸图像缩略图
- **搜索功能**: 基于文件名、标签和元数据的快速搜索
- **批量操作**: 支持文件和文件夹的批量管理

### ⚡ 性能与可靠性
- **高并发**: 基于 Gin 框架的高性能 HTTP 服务
- **异步处理**: 缩略图生成和大文件处理的异步队列
- **健康检查**: 完整的服务健康监控和指标收集
- **优雅关闭**: 支持零停机时间的服务更新

## 🏗 架构设计

```
文件服务架构
├── API Gateway Layer          # API 网关层
│   ├── 认证中间件 (JWT)
│   ├── 配额检查中间件
│   └── 请求路由和负载均衡
├── Handler Layer             # HTTP 处理层
│   ├── FileHandler          # 文件操作 API
│   ├── FolderHandler        # 文件夹操作 API
│   ├── UploadHandler        # 上传处理 API
│   ├── ThumbnailHandler     # 缩略图 API
│   └── HealthHandler        # 健康检查 API
├── Service Layer            # 业务逻辑层
│   ├── FileService          # 文件管理服务
│   ├── UploadService        # 上传管理服务
│   ├── ThumbnailService     # 缩略图服务
│   └── UserServiceClient    # 用户服务客户端
├── Repository Layer         # 数据访问层
│   ├── FileRepository       # 文件数据仓库
│   ├── FolderRepository     # 文件夹数据仓库
│   └── UploadRepository     # 上传会话仓库
├── Storage Layer           # 存储抽象层
│   ├── MinIO Storage       # MinIO 对象存储
│   ├── S3 Storage          # AWS S3 存储
│   └── Memory Storage      # 内存存储(测试用)
└── Database Layer          # 数据库层
    ├── PostgreSQL          # 主数据库
    ├── Redis               # 缓存和会话
    └── 数据库迁移
```

## 🛠 技术栈

- **语言**: Go 1.21+
- **Web框架**: Gin
- **数据库**: PostgreSQL 15+ 
- **缓存**: Redis 7+
- **对象存储**: MinIO / AWS S3
- **ORM**: GORM v2
- **测试**: Testify, HTTPTest
- **文档**: Swagger
- **监控**: Prometheus + Grafana

## 📦 快速开始

### 环境要求

- Go 1.21 或更高版本
- PostgreSQL 15+
- Redis 7+
- MinIO 或 AWS S3 访问权限

### 安装依赖

```bash
# 克隆项目
git clone <repository-url>
cd file-service

# 安装依赖
make deps

# 安装开发工具（可选）
make install-tools
```

### 配置环境

创建 `.env` 文件：

```env
# 服务配置
SERVICE_NAME=file-service
SERVICE_PORT=8080
SERVICE_ENV=development

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_NAME=file_service
DB_USER=postgres
DB_PASSWORD=password
DB_SSL_MODE=disable

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=0

# 存储配置
STORAGE_TYPE=minio
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=files
MINIO_USE_SSL=false

# JWT 配置
JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h

# 用户服务配置
USER_SERVICE_URL=http://localhost:8081
```

### 运行服务

```bash
# 开发模式
make dev

# 或者构建后运行
make build
make run

# 使用 Docker
make docker-build
make docker-run
```

## 📋 API 文档

### 文件操作 API

#### 简单文件上传
```http
POST /api/v1/upload/simple
Content-Type: multipart/form-data

file: <file>
folder_id: <folder_id> (可选)
description: <description> (可选)
tags: <tags> (可选)
```

#### 获取文件列表
```http
GET /api/v1/files
Authorization: Bearer <token>

Parameters:
- page: 页码 (默认: 1)
- page_size: 每页大小 (默认: 20)
- folder_id: 文件夹ID (可选)
- category: 文件分类 (可选)
- sort_by: 排序字段 (name, size, created_at)
- sort_order: 排序方向 (asc, desc)
```

#### 获取文件信息
```http
GET /api/v1/files/{id}
Authorization: Bearer <token>
```

#### 下载文件
```http
GET /api/v1/files/{id}/download
Authorization: Bearer <token>
```

#### 搜索文件
```http
GET /api/v1/files/search?q=<query>
Authorization: Bearer <token>
```

### 分片上传 API

#### 初始化分片上传
```http
POST /api/v1/upload/initiate
Content-Type: application/json
Authorization: Bearer <token>

{
  "file_name": "large-file.zip",
  "file_size": 104857600,
  "content_type": "application/zip",
  "chunk_size": 1048576,
  "folder_id": 123
}
```

#### 上传分片
```http
POST /api/v1/upload/chunk
Content-Type: application/json
Authorization: Bearer <token>

{
  "session_id": "upload-session-id",
  "chunk_index": 0,
  "chunk_size": 1048576,
  "chunk_data": "<base64-encoded-data>"
}
```

#### 完成上传
```http
POST /api/v1/upload/complete
Content-Type: application/json
Authorization: Bearer <token>

{
  "session_id": "upload-session-id"
}
```

### 文件夹操作 API

#### 创建文件夹
```http
POST /api/v1/folders
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "新文件夹",
  "parent_id": 123,
  "description": "文件夹描述"
}
```

#### 获取文件夹内容
```http
GET /api/v1/folders/{id}/contents
Authorization: Bearer <token>
```

#### 获取文件夹树
```http
GET /api/v1/folders/tree
Authorization: Bearer <token>
```

### 缩略图 API

#### 获取缩略图
```http
GET /api/v1/thumbnails/{file_id}?size=medium
Authorization: Bearer <token>

Sizes: small(150x100), medium(300x200), large(600x400)
```

#### 生成缩略图
```http
POST /api/v1/thumbnails/{file_id}/generate
Content-Type: application/json
Authorization: Bearer <token>

{
  "sizes": ["small", "medium", "large"],
  "quality": 80
}
```

完整的 API 文档请访问: `/swagger/index.html`

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
make test

# 只运行单元测试
make test-unit

# 运行集成测试
make test-integration

# 运行基准测试
make test-benchmark

# 生成覆盖率报告
make test-coverage

# 运行完整测试套件
make test-all
```

### 测试覆盖率

项目目标是保持 **80%+** 的测试覆盖率。当前覆盖率报告会在 `test-results/coverage.html` 中生成。

### 测试环境变量

```bash
# 启用集成测试
export INTEGRATION_TESTS=1

# 启用基准测试
export BENCHMARK_ALL=1

# 启用压力测试
export STRESS_TESTS=1
```

## 🏗 部署

### Docker 部署

```bash
# 构建镜像
make docker-build

# 运行容器
docker run -d \
  --name file-service \
  -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e REDIS_HOST=host.docker.internal \
  -e MINIO_ENDPOINT=host.docker.internal:9000 \
  file-service:latest
```

### Docker Compose 部署

```bash
# 启动完整环境
make docker-compose-up

# 停止服务
make docker-compose-down
```

### Kubernetes 部署

参考 `deployments/kubernetes/` 目录下的配置文件。

## 📊 监控和指标

### 健康检查端点

```http
GET /health                    # 基本健康检查
GET /health/ready             # 就绪检查 (Kubernetes)
GET /health/live              # 存活检查 (Kubernetes)
GET /metrics                  # Prometheus 指标
```

### Prometheus 指标

服务暴露以下指标：

- `file_service_requests_total` - 总请求数
- `file_service_request_duration_seconds` - 请求延迟
- `file_service_uploads_total` - 上传次数
- `file_service_storage_bytes` - 存储使用量
- `file_service_active_uploads` - 活跃上传数

### 日志

使用结构化日志格式，支持以下日志级别：
- `DEBUG` - 调试信息
- `INFO` - 一般信息
- `WARN` - 警告信息  
- `ERROR` - 错误信息

## 🔧 配置

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `SERVICE_PORT` | 服务端口 | `8080` |
| `DB_HOST` | 数据库主机 | `localhost` |
| `DB_PORT` | 数据库端口 | `5432` |
| `REDIS_HOST` | Redis主机 | `localhost` |
| `STORAGE_TYPE` | 存储类型 | `minio` |
| `JWT_SECRET` | JWT密钥 | - |
| `LOG_LEVEL` | 日志级别 | `info` |

### 配置文件

支持 YAML 配置文件 `config.yaml`:

```yaml
server:
  port: 8080
  mode: release

database:
  host: localhost
  port: 5432
  name: file_service
  user: postgres
  password: password

storage:
  type: minio
  minio:
    endpoint: localhost:9000
    access_key: minioadmin
    secret_key: minioadmin
    bucket: files
    use_ssl: false

quota:
  default_storage: 1073741824  # 1GB
  default_files: 1000
  check_interval: 60s
```

## 🔒 安全

### 认证和授权

- 使用 JWT token 进行身份认证
- 支持基于角色的访问控制 (RBAC)
- 用户数据完全隔离

### 安全最佳实践

- 所有输入都进行严格验证
- SQL 注入防护
- XSS 攻击防护  
- 文件类型和大小限制
- 上传速率限制

### 安全扫描

```bash
# 运行安全扫描
make security-scan

# 检查依赖漏洞
make deps-check
```

## 🤝 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 开发规范

- 遵循 Go 代码规范
- 编写测试用例，保持测试覆盖率 > 80%
- 更新相关文档
- 提交信息使用约定式提交格式

### 代码质量检查

```bash
# 代码格式化
make fmt

# 代码检查
make lint

# 运行完整 CI 流水线
make ci
```

## 📈 性能

### 性能指标

在标准硬件配置下的性能基准：

- **吞吐量**: 1000+ 请求/秒
- **上传速度**: 100MB/s+（取决于网络）
- **响应时间**: P95 < 100ms
- **并发支持**: 1000+ 并发连接

### 性能优化

- 使用连接池减少数据库连接开销
- Redis 缓存热点数据
- 异步处理耗时操作
- 分片上传减少内存使用
- CDN 加速文件下载

## 📄 许可证

本项目采用 [MIT 许可证](LICENSE)。

## 🙋‍♂️ 支持

如果你有任何问题或建议，请：

1. 查看 [FAQ](docs/FAQ.md)
2. 搜索 [Issues](issues)
3. 创建新的 [Issue](issues/new)
4. 发送邮件到 [support@example.com](mailto:support@example.com)

## 🔄 更新日志

查看 [CHANGELOG.md](CHANGELOG.md) 了解版本更新详情。

## 🌟 致谢

感谢以下开源项目：

- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [GORM](https://github.com/go-gorm/gorm) - Go ORM 库
- [Testify](https://github.com/stretchr/testify) - 测试工具
- [MinIO](https://github.com/minio/minio) - 对象存储服务

---

<div align="center">
Made with ❤️ by File Service Team
</div>