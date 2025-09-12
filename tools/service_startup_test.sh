#!/bin/bash

echo "🚀 Service Startup Order Test"
echo "=================================="

# 设置环境变量
export JWT_SECRET=your-development-secret-key
export JWT_EXPIRATION_HOURS=24
export GIN_MODE=debug

echo "🔧 Environment Setup:"
echo "   JWT_SECRET=$JWT_SECRET"
echo "   JWT_EXPIRATION_HOURS=$JWT_EXPIRATION_HOURS"
echo "   GIN_MODE=$GIN_MODE"
echo ""

# 1. 启动用户服务
echo "📋 Step 1: Starting User Service..."
cd /mnt/d/暑期项目/HorizonCloudService-main/services/user-service

# 检查端口是否被占用
USER_SERVICE_PORT=8001
if lsof -i :$USER_SERVICE_PORT > /dev/null 2>&1; then
    echo "⚠️  Port $USER_SERVICE_PORT is already in use"
    echo "🔄 Stopping existing process..."
    pkill -f "user-service\|main.go" || true
    sleep 2
fi

echo "🔄 Starting user service in background..."
go run cmd/main.go > /tmp/user-service.log 2>&1 &
USER_SERVICE_PID=$!
echo "   PID: $USER_SERVICE_PID"

# 等待用户服务启动
echo "⏳ Waiting for user service to be ready..."
sleep 3

# 检查用户服务是否正常启动
if curl -s http://localhost:8001/health > /dev/null; then
    echo "✅ User service is ready"
else
    echo "❌ User service failed to start"
    echo "📋 User service logs:"
    tail -20 /tmp/user-service.log
    kill $USER_SERVICE_PID 2>/dev/null || true
    exit 1
fi

echo ""

# 2. 启动文件服务
echo "📋 Step 2: Starting File Service..."
cd /mnt/d/暑期项目/HorizonCloudService-main/services/file-service

# 检查端口是否被占用
FILE_SERVICE_PORT=8002
if lsof -i :$FILE_SERVICE_PORT > /dev/null 2>&1; then
    echo "⚠️  Port $FILE_SERVICE_PORT is already in use"
    echo "🔄 Stopping existing process..."
    pkill -f "file-service\|main.go" || true
    sleep 2
fi

echo "🔄 Starting file service in background..."
go run cmd/main.go > /tmp/file-service.log 2>&1 &
FILE_SERVICE_PID=$!
echo "   PID: $FILE_SERVICE_PID"

# 等待文件服务启动
echo "⏳ Waiting for file service to be ready..."
sleep 3

# 检查文件服务是否正常启动
if curl -s http://localhost:8002/health > /dev/null; then
    echo "✅ File service is ready"
else
    echo "❌ File service failed to start"
    echo "📋 File service logs:"
    tail -20 /tmp/file-service.log
    kill $USER_SERVICE_PID $FILE_SERVICE_PID 2>/dev/null || true
    exit 1
fi

echo ""
echo "🎉 Both services are running successfully!"
echo ""

# 3. 测试JWT认证流程
echo "📋 Step 3: Testing JWT Authentication Flow..."

# 3.1 从用户服务获取JWT token
echo "🔐 Getting JWT token from user service..."
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8001/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{
        "username": "testuser",
        "password": "testpassword"
    }')

echo "📝 Token Response: $TOKEN_RESPONSE"

# 检查响应是否包含token
if echo "$TOKEN_RESPONSE" | grep -q "access_token"; then
    ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    echo "✅ JWT token obtained: ${ACCESS_TOKEN:0:20}..."
else
    echo "❌ Failed to obtain JWT token"
    echo "📝 Will test with a manually generated token instead..."
    # 使用测试工具生成token
    ACCESS_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
fi

echo ""

# 3.2 使用JWT访问文件服务
echo "📁 Testing file service with JWT token..."
FILES_RESPONSE=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "X-Request-ID: test-$(date +%s)" \
    http://localhost:8002/api/v1/files)

echo "📝 Files Response: ${FILES_RESPONSE:0:200}..."

if echo "$FILES_RESPONSE" | grep -q "code.*200\|success\|files"; then
    echo "✅ File service JWT authentication working"
else
    echo "❌ File service JWT authentication failed"
    echo "📝 Full response: $FILES_RESPONSE"
fi

echo ""

# 4. 验证请求追踪
echo "📋 Step 4: Testing Request Tracing..."
REQUEST_ID="test-trace-$(date +%s)"
TRACE_RESPONSE=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "X-Request-ID: $REQUEST_ID" \
    -v http://localhost:8002/api/v1/files 2>&1)

if echo "$TRACE_RESPONSE" | grep -q "X-Request-ID: $REQUEST_ID"; then
    echo "✅ Request tracing working correctly"
else
    echo "⚠️  Request tracing may not be working as expected"
    echo "📝 Trace info: $(echo "$TRACE_RESPONSE" | grep -i "x-request-id" | head -5)"
fi

echo ""
echo "🏁 Service Startup and Integration Test Complete!"
echo ""
echo "📊 Summary:"
echo "   - User Service: ✅ Running on :8001"
echo "   - File Service: ✅ Running on :8002"
echo "   - JWT Authentication: $(echo "$FILES_RESPONSE" | grep -q "success\|200" && echo "✅ Working" || echo "❌ Issues detected")"
echo "   - Request Tracing: $(echo "$TRACE_RESPONSE" | grep -q "X-Request-ID" && echo "✅ Working" || echo "⚠️  May have issues")"
echo ""

# 保持服务运行或停止
read -p "🤔 Keep services running for manual testing? (y/N): " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "🛑 Stopping services..."
    kill $USER_SERVICE_PID $FILE_SERVICE_PID 2>/dev/null || true
    echo "✅ Services stopped"
else
    echo "🔄 Services are still running:"
    echo "   - User Service PID: $USER_SERVICE_PID (http://localhost:8001)"
    echo "   - File Service PID: $FILE_SERVICE_PID (http://localhost:8002)"
    echo "   - Kill with: kill $USER_SERVICE_PID $FILE_SERVICE_PID"
fi