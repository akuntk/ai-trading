@echo off
REM AI交易系统版本控制更新测试脚本 (Windows版本)
REM 用于测试完整的版本控制、检测、下载、安装和重启流程

setlocal enabledelayedexpansion

echo 🚀 开始测试AI交易系统版本控制更新系统...
echo ==================================================

REM 配置
set API_BASE_URL=http://localhost:8080
set WEB_BASE_URL=http://localhost:3000
set TEST_VERSION=1.0.1-test

REM 日志函数
:log_info
echo [INFO] %~1
goto :eof

:log_success
echo [SUCCESS] %~1
goto :eof

:log_warning
echo [WARNING] %~1
goto :eof

:log_error
echo [ERROR] %~1
goto :eof

REM 检查依赖
:check_dependencies
call :log_info "检查测试依赖..."

REM 检查curl
curl --version >nul 2>&1
if errorlevel 1 (
    call :log_error "curl未安装，请先安装curl"
    exit /b 1
)

call :log_success "依赖检查完成"
goto :eof

REM 测试API服务器连通性
:test_api_connectivity
call :log_info "测试API服务器连通性..."

curl -s -f "%API_BASE_URL%/api/health" >nul 2>&1
if errorlevel 1 (
    call :log_error "无法连接到API服务器: %API_BASE_URL%"
    exit /b 1
)

call :log_success "API服务器连通正常"
goto :eof

REM 测试获取当前版本
:test_get_current_version
call :log_info "测试获取当前版本..."

curl -s "%API_BASE_URL%/api/version/current" >temp_response.txt 2>&1
if errorlevel 1 (
    call :log_error "无法获取当前版本"
    goto :error
)

findstr "success" temp_response.txt >nul
if errorlevel 1 (
    call :log_error "获取版本失败"
    type temp_response.txt
    goto :error
)

call :log_success "当前版本获取成功"
del temp_response.txt
goto :eof

REM 测试检查更新
:test_check_update
call :log_info "测试检查更新..."

curl -s "%API_BASE_URL%/api/version/check" >temp_response.txt 2>&1
if errorlevel 1 (
    call :log_error "无法检查更新"
    goto :error
)

findstr "success" temp_response.txt >nul
if errorlevel 1 (
    call :log_error "检查更新失败"
    type temp_response.txt
    goto :error
)

call :log_success "检查更新完成"
del temp_response.txt
goto :eof

REM 测试更新状态
:test_update_status
call :log_info "测试获取更新状态..."

curl -s "%API_BASE_URL%/api/version/status" >temp_response.txt 2>&1
if errorlevel 1 (
    call :log_error "无法获取更新状态"
    goto :error
)

findstr "success" temp_response.txt >nul
if errorlevel 1 (
    call :log_error "获取更新状态失败"
    type temp_response.txt
    goto :error
)

call :log_success "更新状态获取成功"
del temp_response.txt
goto :eof

REM 测试自动更新设置
:test_auto_update_setting
call :log_info "测试自动更新设置..."

echo {"enabled":true} >temp_request.json
curl -s -X POST -H "Content-Type: application/json" -d @temp_request.json "%API_BASE_URL%/api/version/auto-update" >temp_response.txt 2>&1

findstr "success" temp_response.txt >nul
if errorlevel 1 (
    call :log_warning "设置自动更新可能失败"
    type temp_response.txt
) else (
    call :log_success "自动更新设置成功"
)

del temp_request.json temp_response.txt 2>nul
goto :eof

REM 测试下载更新（模拟）
:test_download_update
call :log_info "测试下载更新（模拟）..."

echo {"force":false,"auto_restart":false,"backup":true} >temp_request.json
curl -s -X POST -H "Content-Type: application/json" -d @temp_request.json "%API_BASE_URL%/api/version/download" >temp_response.txt 2>&1

findstr "success" temp_response.txt >nul
if errorlevel 1 (
    call :log_warning "下载更新可能失败或正在下载"
) else (
    call :log_success "下载更新请求已发送"
)

del temp_request.json temp_response.txt 2>nul
goto :eof

REM 测试更新进度
:test_update_progress
call :log_info "测试获取更新进度..."

curl -s "%API_BASE_URL%/api/version/progress" >temp_response.txt 2>&1
if errorlevel 1 (
    call :log_warning "无法获取更新进度"
    goto :eof
)

findstr "success" temp_response.txt >nul
if errorlevel 1 (
    call :log_warning "获取更新进度失败"
) else (
    call :log_success "更新进度获取成功"
)

del temp_response.txt 2>nul
goto :eof

REM 测试获取更新历史
:test_update_history
call :log_info "测试获取更新历史..."

curl -s "%API_BASE_URL%/api/version/history" >temp_response.txt 2>&1
if errorlevel 1 (
    call :log_error "无法获取更新历史"
    goto :error
)

findstr "success" temp_response.txt >nul
if errorlevel 1 (
    call :log_error "获取更新历史失败"
    type temp_response.txt
    goto :error
)

call :log_success "更新历史获取成功"
del temp_response.txt
goto :eof

REM 测试前端版本更新页面
:test_frontend_version_page
call :log_info "测试前端版本更新页面..."

curl -s -f "%WEB_BASE_URL%/version" >nul 2>&1
if errorlevel 1 (
    call :log_warning "前端版本更新页面无法访问"
) else (
    call :log_success "前端版本更新页面可访问"
)
goto :eof

REM 压力测试
:stress_test_version_api
call :log_info "执行版本API压力测试..."

set /a success_count=0
set /a total_requests=10

for /l %%i in (1,1,%total_requests%) do (
    curl -s -f "%API_BASE_URL%/api/version/current" >nul 2>&1
    if not errorlevel 1 (
        set /a success_count+=1
    )
    set /p "=." <nul
)

echo.
set /a success_rate=!success_count! * 100 / %total_requests%

if !success_rate! ge 90 (
    call :log_success "压力测试通过 (!success_count!/%total_requests% 成功)"
) else (
    call :log_error "压力测试失败 (!success_count!/%total_requests% 成功)"
)
goto :eof

REM 主测试函数
:run_tests
call :log_info "开始执行测试套件..."
echo.

set /a failed_tests=0
set /a total_tests=0

REM 测试列表
call :test_check_dependencies
if errorlevel 1 (
    set /a failed_tests+=1
)
set /a total_tests+=1

call :test_api_connectivity
if errorlevel 1 (
    set /a failed_tests+=1
)
set /a total_tests+=1

call :test_get_current_version
if errorlevel 1 (
    set /a failed_tests+=1
)
set /a total_tests+=1

call :test_check_update
if errorlevel 1 (
    set /a failed_tests+=1
)
set /a total_tests+=1

call :test_update_status
if errorlevel 1 (
    set /a failed_tests+=1
)
set /a total_tests+=1

call :test_auto_update_setting
set /a total_tests+=1

call :test_download_update
set /a total_tests+=1

call :test_update_progress
set /a total_tests+=1

call :test_update_history
if errorlevel 1 (
    set /a failed_tests+=1
)
set /a total_tests+=1

call :test_frontend_version_page
set /a total_tests+=1

call :stress_test_version_api
if errorlevel 1 (
    set /a failed_tests+=1
)
set /a total_tests+=1

REM 测试结果汇总
echo ==================================================
call :log_info "测试完成！"
echo 总测试数: %total_tests%
echo 通过测试: %total_tests% - %failed_tests%
echo 失败测试: %failed_tests%

if %failed_tests% equ 0 (
    call :log_success "🎉 所有测试通过！版本控制系统工作正常。"
    exit /b 0
) else (
    call :log_error "❌ 有 %failed_tests% 个测试失败，请检查系统配置。"
    exit /b 1
)

REM 错误处理
:error
del temp_*.txt 2>nul
del temp_*.json 2>nul
exit /b 1

REM 主程序入口
:main
echo AI交易系统版本控制更新系统测试
echo ======================================
echo API服务器: %API_BASE_URL%
echo Web服务器: %WEB_BASE_URL%
echo.

REM 检查环境
call :check_dependencies
if errorlevel 1 (
    exit /b 1
)

REM 运行测试
call :run_tests

REM 清理临时文件
del temp_*.txt 2>nul
del temp_*.json 2>nul

goto :eof

REM 运行主程序
call :main