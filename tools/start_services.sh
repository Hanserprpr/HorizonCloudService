#!/bin/bash

# 启动所有后端服务的脚本
echo "🚀 启动集成测试环境"
echo "=================================="

# 设置颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="/mnt/d/暑期项目/HorizonCloudService-main"
cd "$PROJECT_ROOT"

# 环境变量
export JWT_SECRET="your-development-secret-key"
export JWT_EXPIRATION_HOURS=24

# 日志目录
LOG_DIR="$PROJECT_ROOT/logs"
mkdir -p "$LOG_DIR"

echo -e "${BLUE}📋 检查现有服务进程...${NC}"

# 停止现有的服务进程
echo "🛑 停止现有服务进程..."
pkill -f "go run.*main.go" 2>/dev/null || true
pkill -f "user-service" 2>/dev/null || true
pkill -f "file-service" 2>/dev/null || true

# 等待进程完全停止
sleep 2

# 清理端口
echo "🧹 清理端口占用..."
lsof -ti:8001 | xargs kill -9 2>/dev/null || true
lsof -ti:8002 | xargs kill -9 2>/dev/null || true
lsof -ti:3000 | xargs kill -9 2>/dev/null || true

sleep 2

echo -e "${YELLOW}🔧 启动后端服务...${NC}"

# 启动用户服务
echo "👤 启动用户服务 (端口 8001)..."
cd "$PROJECT_ROOT/services/user-service"
JWT_SECRET="$JWT_SECRET" JWT_EXPIRATION_HOURS="$JWT_EXPIRATION_HOURS" \
  go run cmd/main.go > "$LOG_DIR/user-service.log" 2>&1 &
USER_SERVICE_PID=$!
echo "   PID: $USER_SERVICE_PID"

# 等待用户服务启动
echo "   等待用户服务启动..."
sleep 3

# 检查用户服务是否启动成功
if curl -s http://localhost:8001/health > /dev/null; then
    echo -e "   ${GREEN}✅ 用户服务启动成功${NC}"
else
    echo -e "   ${RED}❌ 用户服务启动失败${NC}"
    echo "   日志位置: $LOG_DIR/user-service.log"
    tail -10 "$LOG_DIR/user-service.log"
fi

# 启动文件服务
echo "📁 启动文件服务 (端口 8002)..."
cd "$PROJECT_ROOT/services/file-service"
JWT_SECRET="$JWT_SECRET" JWT_EXPIRATION_HOURS="$JWT_EXPIRATION_HOURS" \
  go run cmd/main.go > "$LOG_DIR/file-service.log" 2>&1 &
FILE_SERVICE_PID=$!
echo "   PID: $FILE_SERVICE_PID"

# 等待文件服务启动
echo "   等待文件服务启动..."
sleep 3

# 检查文件服务是否启动成功
if curl -s http://localhost:8002/api/v1/health > /dev/null; then
    echo -e "   ${GREEN}✅ 文件服务启动成功${NC}"
else
    echo -e "   ${RED}❌ 文件服务启动失败${NC}"
    echo "   日志位置: $LOG_DIR/file-service.log"
    tail -10 "$LOG_DIR/file-service.log"
fi

echo -e "${BLUE}📊 服务状态检查${NC}"

# 服务状态检查
check_service() {
    local service_name="$1"
    local url="$2"
    local expected_code="$3"
    
    echo -n "   $service_name: "
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null)
    
    if [ "$response" = "$expected_code" ]; then
        echo -e "${GREEN}✅ 运行正常 ($response)${NC}"
        return 0
    else
        echo -e "${RED}❌ 状态异常 ($response)${NC}"
        return 1
    fi
}

echo "🔍 检查各服务状态..."
check_service "用户服务健康检查" "http://localhost:8001/health" "200"
check_service "文件服务健康检查" "http://localhost:8002/api/v1/health" "200"
check_service "文件服务就绪检查" "http://localhost:8002/api/v1/health/ready" "200"

echo ""
echo -e "${GREEN}🎯 服务启动完成！${NC}"
echo "=================================="
echo "📋 服务信息:"
echo "   👤 用户服务: http://localhost:8001"
echo "   📁 文件服务: http://localhost:8002"
echo "   📊 日志目录: $LOG_DIR"
echo ""
echo "📝 进程信息:"
echo "   用户服务 PID: $USER_SERVICE_PID"
echo "   文件服务 PID: $FILE_SERVICE_PID"
echo ""
echo "🔧 环境变量:"
echo "   JWT_SECRET: $JWT_SECRET"
echo "   JWT_EXPIRATION_HOURS: $JWT_EXPIRATION_HOURS"
echo ""
echo "📚 查看日志命令:"
echo "   tail -f $LOG_DIR/user-service.log"
echo "   tail -f $LOG_DIR/file-service.log"
echo ""
echo "🛑 停止服务命令:"
echo "   kill $USER_SERVICE_PID $FILE_SERVICE_PID"
echo "   或者运行: pkill -f 'go run.*main.go'"
echo ""
echo -e "${YELLOW}💡 提示: 服务已在后台运行，可以开始集成测试${NC}"
echo "=================================="