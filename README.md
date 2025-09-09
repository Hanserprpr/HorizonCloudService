# 智能媒体云存储系统

基于微服务架构的AI驱动媒体云存储平台，提供智能文件管理、语义搜索、自动标签和协作分享功能。

## 🏗️ 架构概览

### 微服务架构
```
┌─────────────────────────────────────────────────────────────┐
│                    前端层 (Frontend Layer)                  │
├─────────────────────────────────────────────────────────────┤
│  管理后台      │  用户画廊      │  移动端     │  API文档      │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   API网关层 (Gateway Layer)                 │
├─────────────────────────────────────────────────────────────┤
│  认证授权 │ 限流防护 │ 路由转发 │ 负载均衡 │ 监控日志         │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   微服务层 (Service Layer)                  │
├─────────────────────────────────────────────────────────────┤
│ 用户服务 │ 文件服务 │ AI服务 │ 搜索服务 │ 分享服务 │ 通知服务  │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    数据层 (Data Layer)                      │
├─────────────────────────────────────────────────────────────┤
│ MySQL │ PostgreSQL │ Redis │ Qdrant │ MinIO │ Prometheus     │
└─────────────────────────────────────────────────────────────┘
```

### 服务域划分
- **管理域** (Admin): user-service, permission-service
- **业务域** (Business): file-service, search-service, share-service, notification-service  
- **智能域** (AI): ai-service, model-service

## 快速开始

### 环境要求
- Docker 20.10+
- Docker Compose 2.0+
- Go 1.21+ (开发)
- Node.js 18+ (前端开发)

### 启动开发环境
```bash
# 克隆项目
git clone <repository-url>
cd HorizonCloudService-main

# 启动所有服务
make dev-up

# 检查服务状态
make status

# 查看日志
make logs
```

### 服务端点
| 服务 | 端口 | 描述 | 健康检查 |
|------|------|------|----------|
| API网关 | 8080 | 统一入口 | http://localhost:8080/health |
| 用户服务 | 8081 | 用户管理 | http://localhost:8081/health |
| 文件服务 | 8083 | 文件管理 | http://localhost:8083/health |
| AI服务 | 8084 | 智能分析 | http://localhost:8084/health |
| 搜索服务 | 8086 | 语义搜索 | http://localhost:8086/health |

### 管理界面
- **Grafana监控**: http://localhost:3000 (admin/admin123)
- **MinIO控制台**: http://localhost:9001 (minioadmin/12345678)  
- **Prometheus指标**: http://localhost:9090

## 📦 项目结构

```
HorizonCloudService/
├── services/                    # 微服务
│   ├── gateway/                # API网关
│   ├── user-service/           # 用户管理
│   ├── file-service/           # 文件管理
│   ├── ai-service/             # AI处理
│   ├── search-service/         # 搜索服务
│   └── ...                     # 其他服务
├── shared/                     # 共享组件
│   ├── pkg/                   # 共享Go包
│   ├── database/              # 数据库迁移
│   └── monitoring/            # 监控配置
├── frontend/                  # 前端应用
│   ├── admin-portal/          # 管理后台(React)
│   └── user-gallery/          # 用户画廊(React)
├── infrastructure/            # 基础设施
│   ├── kubernetes/            # K8s部署
│   ├── monitoring/            # 监控配置
│   └── nginx/                 # 反向代理
├── docker-compose.yml         # 开发环境
├── Makefile                   # 开发命令
└── go.work                    # Go工作空间
```

## 🛠️ 开发命令

### 环境管理
```bash
make dev-up          # 启动开发环境
make dev-down        # 停止开发环境  
make dev-restart     # 重启开发环境
make dev-logs        # 查看日志
```

### 构建和测试
```bash
make build-all       # 构建所有服务
make test           # 运行所有测试
make lint           # 代码检查
make format         # 代码格式化
```

### 数据库管理
```bash
make db-migrate     # 数据库迁移
make db-seed        # 填充测试数据
make db-backup      # 数据库备份
```

## 🔧 技术栈

### 后端服务
- **语言**: Go 1.21+
- **框架**: Gin (HTTP), GORM (ORM)
- **数据库**: MySQL 8.0, PostgreSQL 15+, Redis 7+
- **存储**: MinIO (S3兼容)
- **向量数据库**: Qdrant
- **消息队列**: Redis Streams

### AI/ML栈
- **运行时**: Python 3.11+ + Go协调
- **模型**: CLIP, BLIP-2, YOLOv8
- **框架**: PyTorch, HuggingFace Transformers
- **推理**: ONNX Runtime

### 前端技术
- **框架**: React 18+, TypeScript
- **UI库**: Ant Design
- **构建**: Vite 4+
- **状态**: Zustand
- **HTTP**: TanStack Query + Axios

### 基础设施
- **容器**: Docker, Kubernetes
- **监控**: Prometheus + Grafana + ELK
- **网关**: Kong (计划), 当前使用Gin
- **CI/CD**: GitHub Actions

## 📊 核心功能

### 智能文件管理
- 🚀 **分片上传**: 支持大文件断点续传
- 🎯 **智能去重**: SHA-256哈希去重
- 📱 **多分辨率缩略图**: 自适应缩略图生成
- 🗂️ **版本控制**: 文件版本历史管理
- ⚡ **多层存储**: 热/温/冷存储自动分层

### AI驱动功能
- 🧠 **图像理解**: CLIP+BLIP多模态分析
- 🔍 **语义搜索**: "找出所有日落风景照片"  
- 👁️ **视觉搜索**: 以图搜图相似度匹配
- 🏷️ **智能标签**: 自动生成描述性标签
- 💡 **智能推荐**: 基于内容和行为推荐

### 协作和分享
- 🔗 **智能分享**: 灵活的权限控制
- ⏰ **过期管理**: 自动过期和续期
- 📊 **访问统计**: 详细的分享分析
- 👥 **团队协作**: 多用户协作工具

## 🔐 安全特性

- **认证**: JWT + OAuth2 + LDAP集成
- **授权**: 基于RBAC的细粒度权限控制  
- **传输**: 全链路HTTPS加密
- **存储**: 文件加密存储
- **审计**: 完整的操作审计日志
- **防护**: API限流、DDoS防护

## 📈 监控和运维

- **指标收集**: Prometheus自定义指标
- **可视化**: Grafana仪表盘
- **日志聚合**: ELK栈集中日志
- **链路追踪**: Jaeger分布式追踪
- **健康检查**: 多级健康检查
- **告警**: 智能告警和故障自愈

## 🚢 部署

### 开发环境
```bash
make dev-up    # Docker Compose一键启动
```

### 生产环境
```bash
make k8s-deploy    # Kubernetes部署
```

### 环境配置
- **开发**: docker-compose.yml
- **测试**: docker-compose.test.yml  
- **生产**: Kubernetes + Helm
