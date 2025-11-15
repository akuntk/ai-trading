@echo off
chcp 65001 >nul
echo 🔄 AI交易系统服务重启脚本
echo ========================================

REM 设置颜色
set "GREEN=[32m"
set "YELLOW=[33m"
set "RED=[31m"
set "BLUE=[34m"
set "NC=[0m"

echo %BLUE%[步骤1/4]%NC% 停止当前运行的服务...

REM 调用停止脚本
call stop.bat

echo.
echo %BLUE%[步骤2/4]%NC% 清理临时文件和缓存...

REM 清理构建文件
if exist "build" (
    echo %YELLOW%正在清理构建文件...%NC%
    rmdir /s /q build
    echo %GREEN%✓%NC% 构建文件已清理
)

REM 清理前端缓存
if exist "web\node_modules\.cache" (
    echo %YELLOW%正在清理前端缓存...%NC%
    rmdir /s /q "web\node_modules\.cache"
    echo %GREEN%✓%NC% 前端缓存已清理
)

REM 清理临时文件
if exist "*.tmp" del /q *.tmp 2>nul
if exist "temp_*.txt" del /q temp_*.txt 2>nul
if exist "temp_*.json" del /q temp_*.json 2>nul

REM 清理日志文件（可选）
echo.
set /p choice="是否清理日志文件？(y/n): "
if /i "%choice%"=="y" (
    if exist "logs" (
        echo %YELLOW%正在清理日志文件...%NC%
        del /q logs\*.log 2>nul
        echo %GREEN%✓%NC% 日志文件已清理
    )
)

echo.
echo %BLUE%[步骤3/4]%NC% 重新部署和启动服务...

REM 检查部署脚本是否存在
if not exist "deploy.bat" (
    echo %RED%错误：未找到部署脚本 deploy.bat%NC%
    echo 请确保在项目根目录运行此脚本
    pause
    exit /b 1
)

REM 重新运行部署脚本
echo %YELLOW%正在重新部署项目...%NC%
call deploy.bat

if errorlevel 1 (
    echo %RED%错误：部署失败%NC%
    pause
    exit /b 1
)

echo.
echo %BLUE%[步骤4/4]%NC% 验证服务启动状态...

REM 等待服务启动
echo %YELLOW%等待服务启动...%NC%
timeout /t 10 /nobreak >nul

REM 检查后端服务
echo %BLUE%检查后端API服务器...%NC%
curl -s -f http://localhost:8080/api/health >nul 2>&1
if errorlevel 1 (
    echo %YELLOW%⚠️  警告：后端服务可能未完全启动%NC%
    echo 请检查后端控制台输出或手动启动服务
    echo 后端命令: nofx.exe
) else (
    echo %GREEN%✓%NC% 后端API服务器运行正常
)

REM 检查前端服务
echo.
echo %BLUE%检查前端开发服务器...%NC%
curl -s -f http://localhost:3000 >nul 2>&1
if errorlevel 1 (
    echo %YELLOW%⚠️  警告：前端服务可能未完全启动%NC%
    echo 请检查前端控制台输出或手动启动服务
    echo 前端命令: cd web && npm run dev
) else (
    echo %GREEN%✓%NC% 前端开发服务器运行正常
)

echo.
echo ========================================
echo %GREEN%🎉 服务重启完成！%NC%
echo ========================================
echo.
echo 服务状态检查结果：
if not errorlevel 1 (
    echo ✅ 后端API服务器: http://localhost:8080
    echo ✅ 前端界面: http://localhost:3000
    echo ✅ 版本管理: http://localhost:3000/version
) else (
    echo ⚠️  请手动检查服务状态
    echo 后端API: http://localhost:8080/api/health
    echo 前端界面: http://localhost:3000
)
echo.
echo %BLUE%访问地址：%NC%
echo - 🏠 主界面: http://localhost:3000
echo - 📡 API服务: http://localhost:8080
echo - 🔧 版本管理: http://localhost:3000/version
echo.
echo %BLUE%常用命令：%NC%
echo - 停止服务: stop.bat
echo - 重启服务: restart.bat
echo - 查看日志: type logs\backend.log 或 logs\frontend.log
echo.
echo %GREEN%服务重启完成！如果遇到问题，请检查控制台输出。%NC%

REM 询问是否打开浏览器
set /p open_browser="是否自动打开浏览器？(y/n): "
if /i "%open_browser%"=="y" (
    echo 正在打开浏览器...
    start http://localhost:3000
)

echo.
pause