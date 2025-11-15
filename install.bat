@echo off
chcp 65001 > nul
echo.
echo =====================================
echo   NOFX AI交易系统 v1.0.1 安装向导
echo =====================================
echo.

:: 检查管理员权限
net session >nul 2>&1
if %errorLevel% == 0 (
    echo ✓ 检测到管理员权限
) else (
    echo ⚠️  建议以管理员权限运行此脚本
    echo.
)

:: 设置安装目录
set INSTALL_DIR=%~dp0
echo 📁 安装目录: %INSTALL_DIR%

:: 检查必要文件
echo.
echo 🔍 检查必要文件...

if not exist "%INSTALL_DIR%nofx.exe" (
    echo ❌ 错误: 找不到主程序 nofx.exe
    echo 请确保下载完整安装包
    pause
    exit /b 1
)

if not exist "%INSTALL_DIR%config.json.example" (
    echo ❌ 错误: 找不到配置文件模板
    pause
    exit /b 1
)

if not exist "%INSTALL_DIR%web\dist\index.html" (
    echo ❌ 错误: 找不到前端文件
    pause
    exit /b 1
)

echo ✓ 必要文件检查完成

:: 创建配置文件
echo.
echo ⚙️  配置系统...

if not exist "%INSTALL_DIR%config.json" (
    echo 创建配置文件...
    copy "%INSTALL_DIR%config.json.example" "%INSTALL_DIR%config.json" > nul
    echo ✓ 配置文件已创建: config.json
    echo.
    echo 📝 请编辑 config.json 文件配置您的交易参数
    echo 当前配置为默认示例配置
) else (
    echo ✓ 配置文件已存在
)

:: 创建必要目录
echo.
echo 📂 创建目录结构...

if not exist "%INSTALL_DIR%logs" mkdir "%INSTALL_DIR%logs"
if not exist "%INSTALL_DIR%backup" mkdir "%INSTALL_DIR%backup"
if not exist "%INSTALL_DIR%temp" mkdir "%INSTALL_DIR%temp"

echo ✓ 目录结构创建完成

:: 检查端口占用
echo.
echo 🔌 检查端口占用...

netstat -an | findstr ":8080" > nul
if %errorLevel% == 0 (
    echo ⚠️  端口 8080 已被占用，请修改配置文件中的端口设置
) else (
    echo ✓ 端口 8080 可用
)

netstat -an | findstr ":3000" > nul
if %errorLevel% == 0 (
    echo ⚠️  端口 3000 已被占用，请修改前端配置
) else (
    echo ✓ 端口 3000 可用
)

:: 设置防火墙规则
echo.
echo 🔥 配置防火墙...

netsh advfirewall firewall show rule name="NOFX API Server" > nul 2>&1
if %errorLevel% neq 0 (
    echo 添加防火墙规则...
    netsh advfirewall firewall add rule name="NOFX API Server" dir=in action=allow protocol=TCP localport=8080 > nul
    echo ✓ API服务器防火墙规则已添加
) else (
    echo ✓ 防火墙规则已存在
)

:: 创建快捷方式
echo.
echo 🎯 创建桌面快捷方式...

set SHORTCUT=%USERPROFILE%\Desktop\NOFX AI交易系统.lnk
powershell -command "$WshShell = New-Object -comObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('%SHORTCUT%'); $Shortcut.TargetPath = '%INSTALL_DIR%nofx.exe'; $Shortcut.WorkingDirectory = '%INSTALL_DIR%'; $Shortcut.Description = 'NOFX AI交易系统 v1.0.1'; $Shortcut.Save()"

if exist "%SHORTCUT%" (
    echo ✓ 桌面快捷方式已创建
) else (
    echo ⚠️  桌面快捷方式创建失败，请手动创建
)

:: 安装完成
echo.
echo =====================================
echo          安装完成!
echo =====================================
echo.
echo 📁 安装目录: %INSTALL_DIR%
echo 🎯 桌面快捷方式: NOFX AI交易系统
echo.
echo 🚀 启动方法:
echo   1. 双击桌面快捷方式
echo   2. 或在安装目录运行: nofx.exe
echo.
echo 🌐 访问地址:
echo   API服务器: http://localhost:8080
echo   Web界面:   http://localhost:3000
echo.
echo 📋 下一步操作:
echo   1. 编辑 config.json 配置交易参数
echo   2. 启动系统进行测试
echo   3. 查看 README.md 了解更多功能
echo.
echo 📞 技术支持: support@nofx.com
echo.

pause