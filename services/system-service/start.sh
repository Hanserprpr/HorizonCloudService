#!/bin/bash

# System Service Startup Script

set -e

echo "🎯 Starting System Service..."
echo "📍 Port: 8003"
echo "🔗 Health check: http://localhost:8003/health"

# 设置开发环境
export GIN_MODE=debug
export PORT=8003
export SKIP_AUTH=true
export JWT_SECRET=your-secret-key

# 确保依赖已安装
if [ ! -f "go.mod" ]; then
    echo "❌ go.mod not found. Please run from the system-service directory."
    exit 1
fi

echo "📦 Installing dependencies..."
go mod tidy

echo "🚀 Starting system service..."
go run cmd/main.go