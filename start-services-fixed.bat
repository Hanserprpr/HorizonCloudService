@echo off
setlocal enabledelayedexpansion

:: Force UTF-8 output for console (optional, but keep ASCII only in this file)
chcp 65001 >nul

:: Move to script directory
pushd "%~dp0"

:: Create logs dir
if not exist "logs" mkdir "logs"

:: Global env
set "JWT_SECRET=your-development-secret-key"
set "JWT_EXPIRATION_HOURS=24"

echo === Starting User Service (port 8001) ===
if not exist "services\user-service" (
  echo [ERROR] Path not found: services\user-service
  goto :END
)
pushd "services\user-service"
start "" cmd /c "go run cmd/main.go > ..\..\logs\user-service.log 2>&1"
popd

echo ... wait for user-service
timeout /t 5 /nobreak >nul

echo === Starting File Service (port 8083) ===
if not exist "services\file-service" (
  echo [ERROR] Path not found: services\file-service
  goto :END
)
pushd "services\file-service"
set "USER_SERVICE_BASE_URL=http://localhost:8001"
set "SERVER_PORT=8083"
set "STORAGE_BACKEND=local"
set "LOCAL_STORAGE_ROOT=./uploads"
if not exist "uploads" mkdir "uploads"
start "" cmd /c "go run cmd/main.go > ..\..\logs\file-service.log 2>&1"
popd

echo ... wait for file-service
timeout /t 5 /nobreak >nul

echo === Starting System Service (port 8003) ===
if not exist "services\system-service" (
  echo [ERROR] Path not found: services\system-service
  goto :END
)
pushd "services\system-service"
set "USER_SERVICE_URL=http://localhost:8001"
set "FILE_SERVICE_URL=http://localhost:8083"
start "" cmd /c "go run cmd/main.go > ..\..\logs\system-service.log 2>&1"
popd

echo.
echo === All services started ===
echo User  : http://localhost:8001
echo File  : http://localhost:8083
echo System: http://localhost:8003
echo.
echo Logs:
echo   logs\user-service.log
echo   logs\file-service.log
echo   logs\system-service.log
echo.
echo Tip: use "type logs\user-service.log" to view logs.

:END
popd
pause
