#!/bin/bash

# 用户服务简单启动脚本
# 设置开发环境变量并启动服务

set -e

echo "=== 用户服务启动 ==="
echo ""

# 加载环境变量
if [ -f ".env.development" ]; then
    echo "📁 加载环境配置: .env.development"
    export $(cat .env.development | grep -v '^#' | xargs)
else
    echo "⚠️  未找到.env.development文件，使用默认配置"
    export JWT_SECRET=your-development-secret-key
    export SERVER_PORT=8001
fi

# 显示关键配置
echo "🔧 服务配置:"
echo "   - 服务名: user-service"
echo "   - 端口: ${SERVER_PORT:-8001}"
echo "   - JWT密钥: ${JWT_SECRET}"
echo ""

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ Go未安装或不在PATH中"
    exit 1
fi

echo "🚀 启动用户服务..."
echo "📍 服务地址: http://localhost:${SERVER_PORT:-8001}"
echo "🔗 健康检查: http://localhost:${SERVER_PORT:-8001}/health"
echo "📚 API基础路径: http://localhost:${SERVER_PORT:-8001}/api/v1"
echo ""

# 启动服务
go run cmd/main.go