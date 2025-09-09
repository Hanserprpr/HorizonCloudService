# 微服务架构框架重构

```
HorizonCloudService/
├── services/                          # Go微服务架构 (8个独立服务)
│   ├── gateway/                      # API网关服务 (Kong-based)
│   │   ├── cmd/main.go              # 服务启动入口
│   │   ├── internal/                # 私有应用代码
│   │   │   ├── auth/                # JWT认证中间件
│   │   │   ├── ratelimit/           # Redis-based限流控制  
│   │   │   ├── routing/             # 智能路由管理
│   │   │   └── monitoring/          # 请求监控和指标
│   │   ├── pkg/                     # 可导出的库代码
│   │   ├── config/                  # 配置文件
│   │   ├── go.mod                   # Go模块定义
│   │   └── Dockerfile
│   │
│   ├── user-service/                # 用户管理服务 (Admin Domain)
│   │   ├── cmd/main.go             # 服务启动入口
│   │   ├── internal/               # 私有业务逻辑
│   │   │   ├── handlers/           # HTTP处理器
│   │   │   ├── models/             # 数据模型和结构体
│   │   │   ├── services/           # 业务逻辑层
│   │   │   ├── repository/         # 数据访问层
│   │   │   └── auth/               # 认证相关逻辑
│   │   ├── pkg/                    # 可复用包
│   │   ├── config/                 # 配置文件
│   │   ├── migrations/             # 数据库迁移文件
│   │   ├── go.mod                  # 模块依赖
│   │   └── tests/
│   │
│   ├── permission-service/          # 权限管理服务 (Admin Domain)
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── rbac/               # RBAC权限控制
│   │   │   ├── policies/           # Casbin策略引擎
│   │   │   ├── audit/              # 审计日志
│   │   │   └── access_control/     # 资源访问控制
│   │   ├── config/
│   │   ├── go.mod
│   │   └── tests/
│   │
│   ├── file-service/                # 文件管理服务 (Business Domain)
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── handlers/           # RESTful文件操作API
│   │   │   ├── storage/            # 多存储后端抽象层
│   │   │   ├── upload/             # 分片上传和断点续传
│   │   │   ├── thumbnail/          # 多分辨率缩略图生成
│   │   │   ├── versioning/         # 文件版本控制
│   │   │   └── deduplication/      # SHA-256文件去重
│   │   ├── config/
│   │   ├── go.mod
│   │   └── tests/
│   │
│   ├── ai-service/                  # AI处理服务 (AI Domain)
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── handlers/           # HTTP API处理器
│   │   │   ├── models/             # CLIP/BLIP模型管理
│   │   │   ├── processors/         # 图像分析处理器
│   │   │   ├── embeddings/         # 向量生成和管理
│   │   │   ├── analysis/           # 内容识别和质量评估
│   │   │   ├── tagging/            # 自动标签生成
│   │   │   └── queue/              # Redis异步任务队列
│   │   ├── python-runtime/         # Python ML运行时
│   │   │   ├── models/
│   │   │   ├── inference/
│   │   │   └── requirements.txt
│   │   ├── config/
│   │   ├── go.mod
│   │   └── tests/
│   │
│   ├── model-service/               # 模型管理服务 (AI Domain)
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── registry/           # 模型版本注册
│   │   │   ├── lifecycle/          # 模型生命周期管理
│   │   │   ├── inference/          # ONNX优化推理
│   │   │   ├── monitoring/         # 模型性能监控
│   │   │   └── scaling/            # 动态资源调度
│   │   ├── config/
│   │   ├── go.mod
│   │   └── tests/
│   │
│   ├── search-service/              # 搜索服务 (Business Domain)
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── handlers/           # 搜索API处理器
│   │   │   ├── vector/             # Qdrant向量搜索
│   │   │   ├── semantic/           # 自然语言语义搜索
│   │   │   ├── visual/             # 以图搜图功能
│   │   │   ├── recommend/          # 智能推荐算法
│   │   │   ├── indexing/           # 搜索索引管理
│   │   │   └── suggestions/        # 搜索建议和自动补全
│   │   ├── config/
│   │   ├── go.mod
│   │   └── tests/
│   │
│   ├── share-service/               # 分享服务 (Business Domain)
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── handlers/           # 分享API处理器
│   │   │   ├── links/              # 分享链接管理
│   │   │   ├── permissions/        # 分享权限控制
│   │   │   ├── expiration/         # 过期时间管理
│   │   │   ├── access_logs/        # 分享访问统计
│   │   │   └── collaboration/      # 协作功能
│   │   ├── config/
│   │   ├── go.mod
│   │   └── tests/
│   │
│   └── notification-service/        # 通知服务 (Business Domain)
│       ├── cmd/main.go             # 服务入口
│       ├── internal/
│       │   ├── handlers/           # HTTP处理器  
│       │   ├── services/           # 通知业务逻辑
│       │   ├── channels/           # 多通道通知实现
│       │   ├── templates/          # 消息模板管理
│       │   └── events/             # 事件处理
│       ├── config/
│       ├── go.mod
│       └── tests/
│
├── frontend/                        # 前端应用 (React生态)
│   ├── admin-portal/               # 管理后台 (React + Ant Design)
│   │   ├── src/
│   │   │   ├── components/         # 可复用组件
│   │   │   ├── pages/             # 页面组件
│   │   │   ├── hooks/             # 自定义React hooks
│   │   │   ├── stores/            # Zustand状态管理
│   │   │   ├── services/          # API调用服务
│   │   │   └── utils/             # 工具函数
│   │   ├── package.json           # npm依赖配置
│   │   └── vite.config.ts         # Vite构建配置
│   │
│   └── user-gallery/              # 用户画廊 (React + Ant Design)
│       ├── src/
│       │   ├── components/        # UI组件库
│       │   ├── pages/            # 页面路由
│       │   ├── stores/           # 全局状态
│       │   ├── hooks/            # 业务hooks
│       │   └── services/         # HTTP客户端
│       └── package.json
│
├── shared/                          # 共享组件
│   ├── database/                   # 数据库相关
│   │   ├── migrations/             # 数据库迁移
│   │   ├── models/                # 共享数据模型
│   │   └── init.sql               # 初始化脚本
│   ├── events/                     # 事件定义
│   ├── proto/                      # gRPC协议定义
│   ├── pkg/                        # 共享Go包
│   │   ├── auth/                  # 认证组件
│   │   ├── cache/                 # 缓存组件
│   │   ├── config/                # 配置管理
│   │   ├── database/              # 数据库连接
│   │   ├── logger/                # 日志组件
│   │   ├── middleware/            # 中间件
│   │   ├── response/              # 响应格式
│   │   └── utils/                 # 通用工具
│   └── monitoring/                 # 监控配置
│
├── infrastructure/                 # 基础设施配置
│   ├── docker/                    # Docker配置
│   ├── kubernetes/                # K8s部署文件
│   ├── nginx/                     # 反向代理配置
│   ├── monitoring/                # 监控部署
│   │   ├── prometheus/
│   │   ├── grafana/
│   │   └── elk/
│   └── storage/                   # 存储配置
│
├── docker-compose.yml             # 本地开发环境
├── docker-compose.prod.yml        # 生产环境
├── Makefile                       # 便捷命令
└── go.work                        # Go workspace配置
```

