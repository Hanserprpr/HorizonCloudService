@echo off
chcp 65001 >nul
echo 🚀 启动修复后的文件服务...

cd /d "services\file-service"
echo 📂 工作目录: %CD%

REM 设置环境变量
set JWT_SECRET=your-development-secret-key
set USER_SERVICE_BASE_URL=http://localhost:8001
set SERVER_PORT=8002
set STORAGE_BACKEND=local
set LOCAL_STORAGE_ROOT=./uploads

echo 🔧 环境变量设置:
echo    JWT_SECRET=%JWT_SECRET%
echo    USER_SERVICE_BASE_URL=%USER_SERVICE_BASE_URL%
echo    SERVER_PORT=%SERVER_PORT%

REM 创建必要的目录
if not exist "uploads" mkdir uploads
if not exist "uploads\thumbnails" mkdir uploads\thumbnails

echo 🔄 启动文件服务...
echo 📍 服务将在 http://localhost:8002 上运行
echo 🔗 健康检查: http://localhost:8002/api/v1/health
echo 按 Ctrl+C 停止服务
echo ============================================

REM 启动预编译的文件服务
file-service.exe
