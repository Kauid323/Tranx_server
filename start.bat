@echo off
chcp 65001 > nul
echo ========================================
echo    TaruApp 社区服务器启动脚本
echo ========================================
echo.

echo [1/3] 检查 Go 环境...
go version
if %errorlevel% neq 0 (
    echo 错误: 未检测到 Go 环境，请先安装 Go 1.21+
    pause
    exit /b 1
)
echo.

echo [2/3] 下载依赖包...
go mod download
if %errorlevel% neq 0 (
    echo 错误: 依赖包下载失败
    pause
    exit /b 1
)
echo.

echo [3/3] 启动服务器...
echo.
echo ========================================
echo  TaruApp 服务器运行在: http://localhost:4999
echo  按 Ctrl+C 停止服务器
echo ========================================
echo.

go run main.go

pause

