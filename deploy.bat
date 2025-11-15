@echo off
chcp 65001 >nul
echo 🚀 AI交易系统本地部署脚本
echo =========================================

REM 设置颜色
set "GREEN=[32m"
set "YELLOW=[33m"
set "RED=[31m"
set "BLUE=[34m"
set "NC=[0m"

REM 检查当前目录
if not exist "go.mod" (
    echo %RED%错误：请确保在项目根目录（AIZH）运行此脚本%NC%
    pause
    exit /b 1
)

echo %BLUE%[信息]%NC% 检测到项目根目录，开始部署...

REM 步骤1：检查Go环境
echo.
echo %YELLOW%[步骤1/7]%NC% 检查Go环境...
go version >nul 2>&1
if errorlevel 1 (
    echo %RED%错误：未检测到Go环境%NC%
    echo 请从 https://golang.org/dl/ 下载并安装Go 1.25或更高版本
    pause
    exit /b 1
)

for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
echo %GREEN%✓%NC% Go环境正常，版本：%GO_VERSION%

REM 步骤2：检查Node.js环境
echo.
echo %YELLOW%[步骤2/7]%NC% 检查Node.js环境...
node --version >nul 2>&1
if errorlevel 1 (
    echo %RED%错误：未检测到Node.js环境%NC%
    echo 请从 https://nodejs.org/ 下载并安装Node.js 18或更高版本
    pause
    exit /b 1
)

for /f "tokens=*" %%i in ('node --version') do set NODE_VERSION=%%i
echo %GREEN%✓%NC% Node.js环境正常，版本：%NODE_VERSION%

REM 步骤3：安装后端依赖
echo.
echo %YELLOW%[步骤3/7]%NC% 安装Go模块依赖...
go mod download
if errorlevel 1 (
    echo %RED%错误：Go模块下载失败%NC%
    echo 请检查网络连接或设置Go代理
    pause
    exit /b 1
)
echo %GREEN%✓%NC% Go依赖安装完成

REM 步骤4：检查并创建配置文件
echo.
echo %YELLOW%[步骤4/7]%NC% 初始化配置文件...

REM 检查config.json
if not exist "config.json" (
    echo %BLUE%[信息]%NC% 创建默认配置文件...
    echo {> config.json
    echo   "api_server_port": 8080,>> config.json
    echo   "beta_mode": false,>> config.json
    echo   "jwt_secret": "your-jwt-secret-key-change-in-production",>> config.json
    echo   "auto_update_enabled": true,>> config.json
    echo   "version_check_interval": "1h">> config.json
    echo }>> config.json
    echo %GREEN%✓%NC% 配置文件已创建：config.json
) else (
    echo %GREEN%✓%NC% 配置文件已存在：config.json
)

REM 检查.env文件
if not exist ".env" (
    echo %BLUE%[信息]%NC% 创建环境变量文件...
    echo # AI交易系统环境变量> .env
    echo APP_VERSION=1.0.0>> .env
    echo BUILD_TIME=%date% %time%>> .env
    echo JWT_SECRET=your-jwt-secret-key-change-in-production>> .env
    echo BETA_MODE=false>> .env
    echo API_PORT=8080>> .env
    echo # 数据库配置>> .env
    echo DB_PATH=./data/nofx.db>> .env
    echo # 交易所API配置（请替换为实际密钥）>> .env
    echo BINANCE_API_KEY=>> .env
    echo BINANCE_SECRET_KEY=>> .env
    echo # AI模型配置>> .env
    echo DEEPSEEK_API_KEY=>> .env
    echo QWEN_API_KEY=>> .env
    echo %GREEN%✓%NC% 环境变量文件已创建：.env
    echo %YELLOW%⚠️  警告：请编辑.env文件，填入实际的API密钥%NC%
) else (
    echo %GREEN%✓%NC% 环境变量文件已存在：.env
)

REM 创建数据目录
if not exist "data" mkdir data
if not exist "logs" mkdir logs
echo %GREEN%✓%NC% 数据目录已创建

REM 步骤5：安装前端依赖
echo.
echo %YELLOW%[步骤5/7]%NC% 安装前端依赖...
cd web
call npm install
if errorlevel 1 (
    echo %RED%错误：前端依赖安装失败%NC%
    echo 请检查Node.js版本和网络连接
    cd ..
    pause
    exit /b 1
)
echo %GREEN%✓%NC% 前端依赖安装完成
cd ..

REM 步骤6：启动后端服务
echo.
echo %YELLOW%[步骤6/7]%NC% 启动后端API服务器...
echo %BLUE%[信息]%NC% 正在编译后端程序...
go build -o nofx.exe .
if errorlevel 1 (
    echo %RED%错误：后端编译失败%NC%
    pause
    exit /b 1
)

echo %GREEN%✓%NC% 后端编译完成
echo %BLUE%[信息]%NC% 启动API服务器（端口8080）...
start "AI交易系统后端" cmd /k "nofx.exe"

REM 等待后端启动
echo %BLUE%[信息]%NC% 等待后端服务启动...
timeout /t 5 /nobreak >nul

REM 检查后端是否启动成功
curl -s -f http://localhost:8080/api/health >nul 2>&1
if errorlevel 1 (
    echo %YELLOW%⚠️  警告：后端服务可能未完全启动，请检查控制台输出%NC%
) else (
    echo %GREEN%✓%NC% 后端API服务器启动成功
)

REM 步骤7：启动前端服务
echo.
echo %YELLOW%[步骤7/7]%NC% 启动前端开发服务器...
cd web
echo %BLUE%[信息]%NC% 启动前端服务器（端口3000）...
start "AI交易系统前端" cmd /k "npm run dev"
cd ..

REM 等待前端启动
timeout /t 3 /nobreak >nul

echo.
echo =========================================
echo %GREEN%🎉 部署完成！%NC%
echo =========================================
echo %BLUE%访问地址：%NC%
echo   - 前端界面: http://localhost:3000
echo   - API服务器: http://localhost:8080
echo   - 版本管理: http://localhost:3000/version
echo.
echo %YELLOW%默认账户信息：%NC%
echo   - 管理员邮箱: admin@example.com
echo   - 管理员密码: admin123
echo.
echo %BLUE%重要提示：%NC%
echo   1. 请编辑 .env 文件，填入实际的API密钥
echo   2. 首次登录需要设置2FA验证
echo   3. 在版本管理页面可以检查和安装更新
echo   4. 按Ctrl+C停止服务器
echo.

REM 询问是否打开浏览器
set /p choice="是否自动打开浏览器？(y/n): "
if /i "%choice%"=="y" (
    echo 正在打开浏览器...
    start http://localhost:3000
)

echo %GREEN%部署脚本执行完成！%NC%
pause