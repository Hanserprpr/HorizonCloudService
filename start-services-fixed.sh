#!/bin/bash

# 修复后的服务启动脚本
echo "🚀 启动修复后的服务..."

# 设置颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查端口是否被占用的函数
check_port() {
    local port=$1
    if lsof -i :$port > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  端口 $port 已被占用，正在停止现有进程...${NC}"
        pkill -f ":$port" || true
        sleep 2
    fi
}

# 等待服务启动的函数
wait_for_service() {
    local url=$1
    local service_name=$2
    local max_attempts=30
    local attempt=1
    
    echo "⏳ 等待 $service_name 启动..."
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$url" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ $service_name 启动成功${NC}"
            return 0
        fi
        echo "   尝试 $attempt/$max_attempts..."
        sleep 2
        ((attempt++))
    done
    
    echo -e "${RED}❌ $service_name 启动失败${NC}"
    return 1
}

# 创建日志目录
mkdir -p logs

# 1. 启动用户服务
echo "📋 启动用户服务 (端口 8001)..."
check_port 8001

cd services/user-service
export JWT_SECRET="your-development-secret-key"
export JWT_EXPIRATION_HOURS="24"
go run cmd/main.go > ../../logs/user-service.log 2>&1 &
USER_SERVICE_PID=$!
echo "   PID: $USER_SERVICE_PID"
cd ../..

# 等待用户服务启动
if ! wait_for_service "http://localhost:8001/health" "用户服务"; then
    echo "用户服务日志:"
    tail -20 logs/user-service.log
    exit 1
fi

# 2. 启动文件服务
echo "📁 启动文件服务 (端口 8083)..."
check_port 8083

cd services/file-service
export JWT_SECRET="your-development-secret-key"
export JWT_EXPIRATION_HOURS="24"
export USER_SERVICE_BASE_URL="http://localhost:8001"
export SERVER_PORT="8083"
export STORAGE_BACKEND="local"
export LOCAL_STORAGE_ROOT="./uploads"
go run cmd/main.go > ../../logs/file-service.log 2>&1 &
FILE_SERVICE_PID=$!
echo "   PID: $FILE_SERVICE_PID"
cd ../..

# 等待文件服务启动
if ! wait_for_service "http://localhost:8083/api/v1/health" "文件服务"; then
    echo "文件服务日志:"
    tail -20 logs/file-service.log
    kill $USER_SERVICE_PID 2>/dev/null || true
    exit 1
fi

# 3. 启动系统服务
echo "⚙️  启动系统服务 (端口 8003)..."
check_port 8003

cd services/system-service
export JWT_SECRET="your-development-secret-key"
export USER_SERVICE_URL="http://localhost:8001"
export FILE_SERVICE_URL="http://localhost:8083"
go run cmd/main.go > ../../logs/system-service.log 2>&1 &
SYSTEM_SERVICE_PID=$!
echo "   PID: $SYSTEM_SERVICE_PID"
cd ../..

# 等待系统服务启动
if ! wait_for_service "http://localhost:8003/health" "系统服务"; then
    echo "系统服务日志:"
    tail -20 logs/system-service.log
    kill $USER_SERVICE_PID $FILE_SERVICE_PID 2>/dev/null || true
    exit 1
fi

echo ""
echo -e "${GREEN}🎉 所有服务启动成功！${NC}"
echo "📋 服务信息:"
echo "   用户服务: http://localhost:8001 (PID: $USER_SERVICE_PID)"
echo "   文件服务: http://localhost:8083 (PID: $FILE_SERVICE_PID)"
echo "   系统服务: http://localhost:8003 (PID: $SYSTEM_SERVICE_PID)"
echo ""
echo "📝 日志文件位置:"
echo "   用户服务: logs/user-service.log"
echo "   文件服务: logs/file-service.log"
echo "   系统服务: logs/system-service.log"
echo ""
echo "🧪 运行测试:"
echo "   node test-config.js"
echo "   node comprehensive-api-test.js"
echo ""
echo "🛑 停止服务:"
echo "   kill $USER_SERVICE_PID $FILE_SERVICE_PID $SYSTEM_SERVICE_PID"

# 保存PID到文件
echo "$USER_SERVICE_PID $FILE_SERVICE_PID $SYSTEM_SERVICE_PID" > .service_pids

echo ""
echo "按 Ctrl+C 停止所有服务"

# 等待中断信号
trap 'echo ""; echo "🛑 停止所有服务..."; kill $USER_SERVICE_PID $FILE_SERVICE_PID $SYSTEM_SERVICE_PID 2>/dev/null || true; rm -f .service_pids; exit 0' INT

# 保持脚本运行
wait
