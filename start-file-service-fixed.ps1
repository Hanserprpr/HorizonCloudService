# PowerShell脚本启动文件服务
Write-Host "🚀 启动修复后的文件服务..." -ForegroundColor Green

# 设置工作目录
Set-Location "services\file-service"
Write-Host "📂 工作目录: $(Get-Location)" -ForegroundColor Yellow

# 设置环境变量
$env:JWT_SECRET = "your-development-secret-key"
$env:USER_SERVICE_BASE_URL = "http://localhost:8001"
$env:SERVER_PORT = "8002"
$env:STORAGE_BACKEND = "local"
$env:LOCAL_STORAGE_ROOT = "./uploads"

Write-Host "🔧 环境变量设置:" -ForegroundColor Cyan
Write-Host "   JWT_SECRET: $env:JWT_SECRET" -ForegroundColor White
Write-Host "   USER_SERVICE_BASE_URL: $env:USER_SERVICE_BASE_URL" -ForegroundColor White
Write-Host "   SERVER_PORT: $env:SERVER_PORT" -ForegroundColor White

# 检查Go环境
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Go未安装或不在PATH中" -ForegroundColor Red
    exit 1
}

# 创建必要的目录
New-Item -ItemType Directory -Force -Path "./uploads" | Out-Null
New-Item -ItemType Directory -Force -Path "./uploads/thumbnails" | Out-Null

Write-Host "🔄 启动文件服务..." -ForegroundColor Green
Write-Host "📍 服务将在 http://localhost:8002 上运行" -ForegroundColor Yellow
Write-Host "🔗 健康检查: http://localhost:8002/api/v1/health" -ForegroundColor Yellow
Write-Host "按 Ctrl+C 停止服务" -ForegroundColor Yellow
Write-Host "============================================" -ForegroundColor Cyan

# 启动服务
try {
    go run cmd/main.go
} catch {
    Write-Host "❌ 启动失败: $_" -ForegroundColor Red
    exit 1
}
