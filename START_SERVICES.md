# 🚀 服务启动指南

本文档提供完整的前后端服务启动步骤，用于测试网盘管理系统的核心功能。

## 📋 系统架构概述

当前实现了以下核心服务：
- **用户服务 (8001端口)**: 处理用户认证、用户管理、配额管理
- **文件服务 (8002端口)**: 处理文件上传、下载、文件夹管理、缩略图
- **管理前端 (5173端口)**: React管理界面，文件管理、用户管理

## 🛠️ 环境要求

### 后端要求
- Go 1.21+ 
- SQLite (自动创建数据库文件)

### 前端要求  
- Node.js 18+
- npm 或 yarn

## ⚡ 快速启动

### 1. 启动用户服务 (8001端口)

```bash
# 进入用户服务目录
cd services/user-service

# 安装依赖
go mod tidy

# 启动服务 (使用简化配置)
go run cmd/main_simple.go
```

**预期输出**:
```
🚀 User Service started successfully on port 8001
📖 API Documentation: http://localhost:8001/api/v1
💓 Health Check: http://localhost:8001/health
🔐 Authentication: POST http://localhost:8001/api/v1/auth/register
```

### 2. 启动文件服务 (8002端口)

**新开终端窗口**:
```bash
# 进入文件服务目录
cd services/file-service

# 安装依赖
go mod tidy

# 启动服务 (使用简化配置)
go run cmd/main_simple.go
```

**预期输出**:
```
🎉 File Service HTTP Server started successfully!
📍 Server running on port :8002
🔗 Health check: http://localhost:8002/health
🔗 Ready check: http://localhost:8002/health/ready
📁 Upload directory: ./uploads
🗄️  Database: file-service.db
🌐 API Base: http://localhost:8002/api/v1
```

### 3. 启动前端管理界面 (5173端口)

**新开终端窗口**:
```bash
# 进入前端目录
cd frontend/admin-portal

# 安装依赖 (首次运行)
npm install

# 启动开发服务器
npm run dev
```

**预期输出**:
```
  VITE v5.0.x ready in xxx ms

  ➜  Local:   http://localhost:5173/
  ➜  Network: use --host to expose
```

## 🔧 验证服务状态

### 检查服务健康状态

```bash
# 检查用户服务
curl http://localhost:8001/health

# 检查文件服务  
curl http://localhost:8002/health
```

**预期响应**:
```json
{
  "status": "healthy",
  "service": "user-service", // 或 "file-service"
  "timestamp": 1640995200,
  "version": "1.0.0"
}
```

### 访问管理界面

1. 打开浏览器访问: http://localhost:5173
2. 应该看到登录页面
3. 注册新用户或使用测试账号登录

## 📊 功能测试清单

### ✅ 用户认证功能
- [ ] 用户注册 - POST `/api/v1/auth/register`
- [ ] 用户登录 - POST `/api/v1/auth/login`  
- [ ] JWT Token认证
- [ ] 自动登录状态保持

### ✅ 文件管理功能
- [ ] 文件夹浏览 - GET `/api/v1/folders/tree`
- [ ] 文件夹内容 - GET `/api/v1/folders/{id}/contents`
- [ ] 简单文件上传 - POST `/api/v1/upload/simple`
- [ ] 分片文件上传 - POST `/api/v1/upload/initiate`
- [ ] 文件下载 - GET `/api/v1/files/{id}/download`
- [ ] 文件删除 - DELETE `/api/v1/files/{id}`
- [ ] 缩略图生成 - POST `/api/v1/thumbnails/{file_id}/generate`

### ✅ 管理界面功能
- [ ] 仪表盘统计信息
- [ ] 文件列表和搜索
- [ ] 文件上传进度显示
- [ ] 用户管理界面
- [ ] 系统设置页面

## 🐛 常见问题排查

### 1. 端口占用错误
```bash
# 检查端口占用
netstat -an | grep :8001  # Windows
lsof -i :8001            # macOS/Linux

# 终止占用进程
kill -9 <PID>
```

### 2. 数据库连接错误
- 确保服务目录有写入权限
- SQLite数据库文件会自动创建在服务根目录

### 3. CORS跨域错误
- 确认前端访问地址在CORS允许列表中
- 检查浏览器控制台是否有CORS错误

### 4. 文件上传失败
- 检查文件服务是否正常启动 (8002端口)
- 验证./uploads目录是否创建成功
- 检查文件大小是否超过限制

### 5. JWT认证失败
- 确认登录后token正确存储
- 检查API请求头是否包含Authorization
- 验证token未过期

## 🔍 API端点测试

### 用户服务API (8001端口)
```bash
# 注册用户
curl -X POST http://localhost:8001/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","email":"admin@example.com","password":"123456"}'

# 用户登录
curl -X POST http://localhost:8001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}'
```

### 文件服务API (8002端口)
```bash
# 获取文件夹树 (需要JWT token)
curl -X GET http://localhost:8002/api/v1/folders/tree \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 简单文件上传 (需要JWT token)
curl -X POST http://localhost:8002/api/v1/upload/simple \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@/path/to/your/file.jpg"
```

## 📝 开发说明

### 数据库文件位置
- 用户服务: `services/user-service/user-service.db`
- 文件服务: `services/file-service/file-service.db`

### 上传文件存储
- 本地存储路径: `services/file-service/uploads/`
- 缩略图路径: `services/file-service/uploads/thumbnails/`

### 日志查看
- 服务启动时会显示详细的API端点信息
- 开发环境会在控制台输出详细的请求日志
- 前端API调用会在浏览器开发者工具中显示

## 🎯 下一步开发

1. **权限服务** - 实现RBAC权限控制
2. **AI服务** - 添加智能图片分析
3. **搜索服务** - 实现语义搜索功能
4. **API网关** - 统一API入口和认证
5. **生产部署** - Docker容器化部署

---

🎉 **服务启动成功后，访问 http://localhost:5173 开始体验完整的网盘管理系统！**