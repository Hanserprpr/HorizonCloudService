#!/bin/bash

# 文件服务启动脚本 - 简化版本
echo "🚀 Starting File Service HTTP Server..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed or not in PATH"
    exit 1
fi

# 进入文件服务目录
cd "$(dirname "$0")"
echo "📂 Working directory: $(pwd)"

# 安装依赖 (如果需要)
echo "🔧 Installing dependencies..."
go mod tidy

# 创建必要的目录
mkdir -p ./uploads
mkdir -p ./uploads/thumbnails

# 设置环境变量
export SERVER_PORT=8002
export STORAGE_BACKEND=local
export LOCAL_STORAGE_ROOT=./uploads
export JWT_SECRET=your-development-secret-key

echo "⚙️  Environment Configuration:"
echo "   - Port: 8002"
echo "   - Storage: Local (./uploads)"
echo "   - Database: SQLite (file-service.db)"
echo "   - JWT Secret: Development key"

echo ""
echo "🔄 Starting server..."
echo "📍 Server will be available at: http://localhost:8002"
echo "🔗 Health check: http://localhost:8002/health"
echo "📚 API Documentation: http://localhost:8002/api/v1"
echo ""
echo "Press Ctrl+C to stop the server"
echo "============================================"

# 启动服务器
go run cmd/main_simple.go