# 前后端API集成指南

## 🚨 当前状态：需要修复才能正常使用

前端和后端API **不完全匹配**，需要进行以下修复才能启动后端并正常使用完整功能。

## 🔍 发现的问题

### 1. **端口配置不匹配**
- **前端默认配置**: `http://localhost:8083`
- **后端实际端口**:
  - 用户服务: `8001`
  - 文件服务: 需要配置HTTP服务器

### 2. **微服务架构 vs 统一API**
- **前端设计**: 期望统一的API网关
- **后端实现**: 独立的微服务

### 3. **API响应格式**
- 需要验证响应数据结构是否一致

## 🛠 修复步骤

### 步骤1: 配置前端环境变量

已创建 `.env.development` 文件：
```env
VITE_API_BASE_URL=http://localhost:8001
VITE_USER_SERVICE_URL=http://localhost:8001
VITE_FILE_SERVICE_URL=http://localhost:8002
```

### 步骤2: 启动后端服务

#### 用户服务 (端口 8001)
```bash
cd services/user-service
go run cmd/main_simple.go
```

#### 文件服务 (需要修复)
文件服务缺少完整的HTTP服务器，需要补充。

### 步骤3: 创建简单的文件服务HTTP服务器

创建 `services/file-service/cmd/main.go`:

```go
package main

import (
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()
	
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "file-service",
		})
	})
	
	// 基础API路由
	api := router.Group("/api/v1")
	{
		// 文件路由 - 暂时返回Mock数据
		api.GET("/files", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"data": []interface{}{},
				"total": 0,
			})
		})
	}
	
	fmt.Println("🚀 File Service started on port 8002")
	router.Run(":8002")
}
```

### 步骤4: API对应关系

#### 认证API ✅ 已匹配
| 前端调用 | 后端路由 | 状态 |
|---------|---------|------|
| `POST /api/v1/auth/login` | `POST /api/v1/auth/login` | ✅ 匹配 |
| `POST /api/v1/auth/logout` | `POST /api/v1/auth/logout` | ✅ 匹配 |
| `POST /api/v1/auth/refresh` | `POST /api/v1/auth/refresh` | ✅ 匹配 |

#### 用户管理API ✅ 已匹配
| 前端调用 | 后端路由 | 状态 |
|---------|---------|------|
| `GET /api/v1/users/profile` | `GET /api/v1/users/profile` | ✅ 匹配 |
| `PUT /api/v1/users/profile` | `PUT /api/v1/users/profile` | ✅ 匹配 |
| `GET /api/v1/admin/users` | `GET /api/v1/admin/users` | ✅ 匹配 |

#### 文件管理API ❌ 需要实现
| 前端期望 | 后端状态 | 需要操作 |
|---------|---------|---------|
| `GET /api/v1/files` | ❌ 缺失 | 需要实现 |
| `POST /api/v1/upload/initiate` | ❌ 缺失 | 需要实现 |
| `POST /api/v1/upload/chunk` | ❌ 缺失 | 需要实现 |

## 🚀 启动指南

### 快速启动（最小可用版本）

1. **启动用户服务**:
```bash
cd services/user-service
go run cmd/main_simple.go
```

2. **启动前端**:
```bash
cd frontend/admin-portal
npm run dev
```

3. **验证连接**:
   - 访问: `http://localhost:3001`
   - 测试登录功能（只有用户认证可用）

### 完整功能启动

1. **实现文件服务HTTP服务器**
2. **启动所有后端服务**
3. **配置API网关**（可选）

## 🧪 测试API连接

### 测试用户服务连接
```bash
curl http://localhost:8001/health
curl -X POST http://localhost:8001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

### 测试前端API调用
在浏览器开发者工具中查看网络请求，验证API调用是否成功。

## 📋 功能可用性矩阵

| 功能模块 | 前端实现 | 后端API | 集成状态 |
|---------|---------|---------|---------|
| 用户登录 | ✅ | ✅ | 🟡 需要配置端口 |
| 用户管理 | ✅ | ✅ | 🟡 需要配置端口 |
| 文件管理 | ✅ | ❌ | ❌ 需要实现后端 |
| 文件上传 | ✅ | ❌ | ❌ 需要实现后端 |
| 系统设置 | ✅ | ❌ | ❌ 需要实现后端 |

## 🔧 推荐的集成方案

### 方案A: 最小可用集成（推荐）
1. 修复端口配置
2. 只启动用户服务
3. 其他功能使用Mock数据

### 方案B: 完整功能集成
1. 完善文件服务HTTP服务器
2. 实现所有API端点
3. 配置API网关

### 方案C: API网关集成
1. 使用Nginx或Kong作为API网关
2. 统一路由到各个微服务
3. 前端只需要连接网关

## 🎯 结论

**当前状态**: 前端已完全实现，后端部分实现
**可用功能**: 用户认证和管理（需要端口配置修复）
**需要补充**: 文件管理、系统设置等功能的后端API

建议先使用**方案A**进行最小可用集成，验证基础功能后再逐步完善其他模块。