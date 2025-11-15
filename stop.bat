@echo off
chcp 65001 >nul
echo 🛑 AI交易系统服务停止脚本
echo =========================================

REM 设置颜色
set "GREEN=[32m"
set "YELLOW=[33m"
set "RED=[31m"
set "BLUE=[34m"
set "NC=[0m"

REM 停止后端服务
echo %YELLOW%[步骤1/2]%NC% 停止后端API服务器...

REM 查找并停止nofx.exe进程
tasklist /FI "IMAGENAME eq nofx.exe" 2>NUL | find /I "nofx.exe" >NUL
if %ERRORLEVEL% EQU 0 (
    echo %BLUE%[信息]%NC% 发现运行中的后端进程，正在停止...
    taskkill /F /IM nofx.exe >NUL 2>&1
    if %ERRORLEVEL% EQU 0 (
        echo %GREEN%✓%NC% 后端服务已停止
    ) else (
        echo %YELLOW%⚠️  警告：停止后端服务时出现问题%NC%
    )
) else (
    echo %GREEN%✓%NC% 后端服务未运行
)

REM 停止可能的Go进程
tasklist /FI "IMAGENAME eq go.exe" 2>NUL | find /I "go.exe" >NUL
if %ERRORLEVEL% EQU 0 (
    for /f "tokens=2" %%i in ('tasklist /FI "IMAGENAME eq go.exe" /FO LIST ^| findstr "go.exe"') do (
        taskkill /F /PID %%i >NUL 2>&1
    )
    echo %GREEN%✓%NC% 已清理可能的Go进程
)

REM 停止前端服务
echo.
echo %YELLOW%[步骤2/2]%NC% 停止前端开发服务器...

REM 查找并停止Node.js进程（npm run dev）
for /f "tokens=2" %%i in ('tasklist /FI "WINDOWTITLE eq AI交易系统前端*" /FO LIST ^| findstr /C:"node.exe" 2^>NUL') do (
    taskkill /F /PID %%i >NUL 2>&1
    if %ERRORLEVEL% EQU 0 (
        echo %GREEN%✓%NC% 前端服务已停止
        goto :frontend_stopped
    )
)

REM 备用方法：查找所有与vite相关的Node.js进程
for /f "tokens=2" %%i in ('tasklist /FI "IMAGENAME eq node.exe" /FO LIST ^| findstr "node.exe"') do (
    tasklist /FI "PID eq %%i" /FO LIST ^| findstr /I "vite\|dev" >NUL
    if %ERRORLEVEL% EQU 0 (
        taskkill /F /PID %%i >NUL 2>&1
        echo %GREEN%✓%NC% 前端服务已停止
        goto :frontend_stopped
    )
)

:frontend_stopped

REM 停止可能残留的npm进程
tasklist /FI "IMAGENAME eq npm.exe" 2>NUL | find /I "npm.exe" >NUL
if %ERRORLEVEL% EQU 0 (
    taskkill /F /IM npm.exe >NUL 2>&1
    echo %GREEN%✓%NC% 已清理npm进程
)

REM 检查端口占用
echo.
echo %BLUE%[信息]%NC% 检查端口占用情况...

REM 检查8080端口
netstat -ano | findstr :8080 >NUL
if %ERRORLEVEL% EQU 0 (
    echo %YELLOW%⚠️  警告：端口8080仍被占用%NC%
    netstat -ano | findstr :8080
) else (
    echo %GREEN%✓%NC% 端口8080已释放
)

REM 检查3000端口
netstat -ano | findstr :3000 >NUL
if %ERRORLEVEL% EQU 0 (
    echo %YELLOW%⚠️  警告：端口3000仍被占用%NC%
    netstat -ano | findstr :3000
) else (
    echo %GREEN%✓%NC% 端口3000已释放
)

echo.
echo =========================================
echo %GREEN%✅ 所有服务已停止！%NC%
echo =========================================
echo.

REM 询问是否删除PID文件
set /p choice="是否删除进程ID文件？(y/n): "
if /i "%choice%"=="y" (
    if exist ".backend_pid" del .backend_pid >NUL 2>&1
    if exist ".frontend_pid" del .frontend_pid >NUL 2>&1
    echo %GREEN%✓%NC% PID文件已删除
)

pause