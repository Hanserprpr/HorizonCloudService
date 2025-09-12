@echo off
echo 🚀 启动文件服务 (简化版)...

cd /d "services\file-service"

REM 设置环境变量
set USER_SERVICE_BASE_URL=http://localhost:8001
set JWT_SECRET=your-development-secret-key
set SERVER_PORT=8002

echo 🔧 环境变量:
echo    USER_SERVICE_BASE_URL=%USER_SERVICE_BASE_URL%
echo    JWT_SECRET=%JWT_SECRET%
echo    SERVER_PORT=%SERVER_PORT%

echo 🔄 启动文件服务...
file-service.exe

pause
