@echo off
chcp 65001 >nul
echo 📁 启动文件服务...

cd services\file-service

set JWT_SECRET=your-development-secret-key
set USER_SERVICE_BASE_URL=http://localhost:8001
set SERVER_PORT=8002
set STORAGE_BACKEND=local
set LOCAL_STORAGE_ROOT=./uploads

echo 🔧 环境变量设置:
echo    JWT_SECRET=%JWT_SECRET%
echo    USER_SERVICE_BASE_URL=%USER_SERVICE_BASE_URL%
echo    SERVER_PORT=%SERVER_PORT%

echo 🚀 启动文件服务...
go run cmd/main.go
