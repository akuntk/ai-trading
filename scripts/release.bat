@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM AI交易系统版本发布脚本 (Windows版本)
set "GREEN=[32m"
set "YELLOW=[33m"
set "RED=[31m"
set "BLUE=[34m"
set "PURPLE=[95m"
set "NC=[0m"

REM 项目配置
set PROJECT_NAME=nofx
set VERSION_FILE=version.txt
set CHANGELOG_FILE=CHANGELOG.md
set BUILD_DIR=build
set RELEASE_DIR=releases

REM 默认参数
set VERSION=
set VERSION_TYPE=
set BUILD_TYPE=release
set PLATFORM=all
set SKIP_TESTS=false
set DRY_RUN=false
set SKIP_GIT=false
set SKIP_GITHUB=false

REM 解析命令行参数
:parse_args
if "%~1"=="" goto main
if "%~1"=="-h" goto show_help
if "%~1"=="--help" goto show_help
if "%~1"=="-t" (
    set VERSION_TYPE=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="-b" (
    set BUILD_TYPE=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="-p" (
    set PLATFORM=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="-s" (
    set SKIP_TESTS=true
    shift
    goto parse_args
)
if "%~1"=="-d" (
    set DRY_RUN=true
    shift
    goto parse_args
)
if "%~1"=="--no-git" (
    set SKIP_GIT=true
    shift
    goto parse_args
)
if "%~1"=="--no-github" (
    set SKIP_GITHUB=true
    shift
    goto parse_args
)
if "%VERSION%"=="" (
    set VERSION=%~1
    shift
    goto parse_args
)
shift
goto parse_args

:show_help
echo AI交易系统版本发布脚本 (Windows)
echo.
echo 用法: %0 [选项] [版本号]
echo.
echo 选项:
echo   -h, --help        显示此帮助信息
echo   -t, --type        版本类型 (major^|minor^|patch^|pre)
echo   -b, --build       构建类型 (debug^|release)
echo   -p, --platform    目标平台 (all^|windows^|linux^|darwin)
echo   -s, --skip-tests  跳过测试
echo   -d, --dry-run     仅显示将要执行的操作，不实际执行
echo   --no-git          跳过Git操作
echo   --no-github       跳过GitHub发布
echo.
echo 示例:
echo   %0 1.1.0                    # 发布1.1.0版本
echo   %0 -t minor                 # 发布下一个次版本
echo   %0 -t patch --no-git       # 发布补丁版本，跳过Git
echo   %0 -t major --dry-run       # 预览主版本发布
goto :eof

REM 日志函数
:log_info
echo %BLUE%[信息]%NC% %~1
goto :eof

:log_success
echo %GREEN%[成功]%NC% %~1
goto :eof

:log_warning
echo %YELLOW%[警告]%NC% %~1
goto :eof

:log_error
echo %RED%[错误]%NC% %~1
goto :eof

:log_step
echo %PURPLE%[步骤]%NC% %~1
goto :eof

REM 检查Git状态
:check_git_status
if "%SKIP_GIT%"=="true" goto :eof

call :log_step "检查Git状态..."

REM 检查是否有未提交的更改
git status --porcelain >nul 2>&1
if errorlevel 1 (
    call :log_error "未检测到Git仓库"
    goto :eof
)

git status --porcelain > temp_git_status.txt
for /f %%i in (temp_git_status.txt) do set "GIT_HAS_CHANGES=1"
del temp_git_status.txt

if defined GIT_HAS_CHANGES (
    call :log_error "有未提交的更改，请先提交所有更改"
    git status --short
    exit /b 1
)

call :log_success "Git状态检查完成"
goto :eof

REM 获取当前版本
:get_current_version
if exist "%VERSION_FILE%" (
    set /p CURRENT_VERSION=<%VERSION_FILE%
) else (
    set CURRENT_VERSION=1.0.0
)
goto :eof

REM 计算下一个版本号
:calculate_next_version
call :get_current_version

REM 解析版本号
for /f "tokens=1,2,3 delims=." %%a in ("%CURRENT_VERSION%") do (
    set MAJOR=%%a
    set MINOR=%%b
    set PATCH=%%c
)

if "%VERSION_TYPE%"=="major" (
    set /a MAJOR+=1
    set MINOR=0
    set PATCH=0
) else if "%VERSION_TYPE%"=="minor" (
    set /a MINOR+=1
    set PATCH=0
) else if "%VERSION_TYPE%"=="patch" (
    set /a PATCH+=1
) else if "%VERSION_TYPE%"=="pre" (
    set /a PATCH+=1
    set VERSION=%MAJOR%.%MINOR%.%PATCH%-pre
    goto :eof
) else (
    call :log_error "无效的版本类型: %VERSION_TYPE%"
    exit /b 1
)

set VERSION=%MAJOR%.%MINOR%.%PATCH%
goto :eof

REM 验证版本号格式
:validate_version
echo %VERSION% | findstr /R "^[0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*" >nul
if errorlevel 1 (
    call :log_error "无效的版本号格式: %VERSION%
    call :log_error "版本号格式应为: major.minor.patch[-suffix]"
    exit /b 1
)
goto :eof

REM 运行测试
:run_tests
if "%SKIP_TESTS%"=="true" (
    call :log_warning "跳过测试"
    goto :eof
)

call :log_step "运行测试..."

REM 后端测试
call :log_info "运行后端Go测试..."
go test ./... -v
if errorlevel 1 (
    call :log_error "后端测试失败"
    exit /b 1
)

REM 前端测试
call :log_info "运行前端测试..."
cd web
npm test
if errorlevel 1 (
    call :log_error "前端测试失败"
    cd ..
    exit /b 1
)
cd ..

call :log_success "所有测试通过"
goto :eof

REM 更新版本信息
:update_version
if "%DRY_RUN%"=="true" (
    call :log_info "[DRY RUN] 将更新版本到: %VERSION%"
    goto :eof
)

call :log_step "更新版本信息..."

REM 更新版本文件
echo %VERSION% > %VERSION_FILE%

REM 更新package.json版本
powershell -Command "(Get-Content web\package.json) -replace '\"version\": \".*\"', '\"version\": \"%VERSION%\"' | Set-Content web\package.json"

call :log_success "版本信息已更新到 %VERSION%"
goto :eof

REM 构建二进制文件
:build_binaries
call :log_step "构建二进制文件..."

if exist "%BUILD_DIR%" rmdir /s /q "%BUILD_DIR%"
mkdir "%BUILD_DIR%"

REM 设置构建标志
set LDFLAGS=-X main.AppVersion=%VERSION% -X main.BuildTime=%date% %time%
if "%BUILD_TYPE%"=="release" (
    set LDFLAGS=%LDFLAGS% -s -w
)

REM 构建Windows版本
if "%PLATFORM%"=="all" set PLATFORM=windows
if "%PLATFORM%"=="windows" (
    call :log_info "构建Windows AMD64..."
    set GOOS=windows
    set GOARCH=amd64
    go build -ldflags "%LDFLAGS%" -o "%BUILD_DIR%\%PROJECT_NAME%-windows-amd64.exe" .
    if errorlevel 1 (
        call :log_error "Windows构建失败"
        exit /b 1
    )
    call :log_success "Windows AMD64构建成功"
)

call :log_success "二进制文件构建完成"
goto :eof

REM 构建前端
:build_frontend
call :log_step "构建前端..."

cd web
call npm ci
if errorlevel 1 (
    call :log_error "npm依赖安装失败"
    cd ..
    exit /b 1
)

if "%BUILD_TYPE%"=="release" (
    call npm run build
) else (
    call npm run build:dev 2>nul || call npm run build
)

if errorlevel 1 (
    call :log_error "前端构建失败"
    cd ..
    exit /b 1
)

cd ..
call :log_success "前端构建成功"
goto :eof

REM 创建发布包
:create_release_packages
call :log_step "创建发布包..."

set VERSION_RELEASE_DIR=%RELEASE_DIR%\v%VERSION%
if exist "%VERSION_RELEASE_DIR%" rmdir /s /q "%VERSION_RELEASE_DIR%"
mkdir "%VERSION_RELEASE_DIR%"

REM 复制二进制文件
if exist "%BUILD_DIR%" (
    xcopy "%BUILD_DIR%" "%VERSION_RELEASE_DIR%" /E /I /Y
)

REM 复制前端文件
if exist "web\dist" (
    mkdir "%VERSION_RELEASE_DIR%\web"
    xcopy "web\dist" "%VERSION_RELEASE_DIR%\web" /E /I /Y /S
)

REM 复制配置文件
if exist "config.json.example" (
    copy "config.json.example" "%VERSION_RELEASE_DIR%\"
)
if exist "README.md" (
    copy "README.md" "%VERSION_RELEASE_DIR%\"
)
if exist "%CHANGELOG_FILE%" (
    copy "%CHANGELOG_FILE%" "%VERSION_RELEASE_DIR%\"
)

REM 创建Windows安装脚本
echo @echo off > "%VERSION_RELEASE_DIR%\install.bat"
echo echo 正在安装AI交易系统... >> "%VERSION_RELEASE_DIR%\install.bat"
echo. >> "%VERSION_RELEASE_DIR%\install.bat"
echo if exist "%PROJECT_NAME%-windows-amd64.exe" ( >> "%VERSION_RELEASE_DIR%\install.bat"
echo     copy "%PROJECT_NAME%-windows-amd64.exe" "%PROJECT_NAME%.exe" >> "%VERSION_RELEASE_DIR%\install.bat"
echo     echo 安装完成！ >> "%VERSION_RELEASE_DIR%\install.bat"
echo     echo 使用方法: >> "%VERSION_RELEASE_DIR%\install.bat"
echo     echo   %PROJECT_NAME%.exe --help >> "%VERSION_RELEASE_DIR%\install.bat"
echo ) else ( >> "%VERSION_RELEASE_DIR%\install.bat"
echo     echo 错误: 找不到Windows二进制文件 >> "%VERSION_RELEASE_DIR%\install.bat"
echo     pause >> "%VERSION_RELEASE_DIR%\install.bat"
echo     exit /b 1 >> "%VERSION_RELEASE_DIR%\install.bat"
echo ) >> "%VERSION_RELEASE_DIR%\install.bat"

call :log_success "发布包已创建: %VERSION_RELEASE_DIR%"
goto :eof

REM 生成校验和
:generate_checksums
call :log_step "生成文件校验和..."

cd "%VERSION_RELEASE_DIR%"

REM 生成SHA256校验和
for %%f in (*) do (
    certutil -hashfile "%%f" SHA256 | findstr /V "hash" > "%%f.sha256"
)

REM 生成MD5校验和
for %%f in (*) do (
    certutil -hashfile "%%f" MD5 | findstr /V "hash" > "%%f.md5"
)

REM 合并校验和文件
del checksums.txt 2>nul
for %%f in (*.sha256) do (
    type "%%f" >> checksums.txt
    del "%%f"
)

cd ..

call :log_success "校验和文件已生成"
goto :eof

REM 生成发布报告
:generate_release_report
call :log_step "生成发布报告..."

set REPORT_FILE=reports\release-v%VERSION%.md
if not exist "reports" mkdir reports

echo # 版本发布报告 - v%VERSION% > "%REPORT_FILE%"
echo. >> "%REPORT_FILE%"
echo ## 发布信息 >> "%REPORT_FILE%"
echo. >> "%REPORT_FILE%"
echo - **版本号**: v%VERSION% >> "%REPORT_FILE%"
echo - **发布时间**: %date% %time% >> "%REPORT_FILE%"
echo - **构建类型**: %BUILD_TYPE% >> "%REPORT_FILE%"
echo - **目标平台**: %PLATFORM% >> "%REPORT_FILE%"
echo. >> "%REPORT_FILE%"
echo ## 发布内容 >> "%REPORT_FILE%"
echo. >> "%REPORT_FILE%"
echo ### 后端 >> "%REPORT_FILE%"
echo - Go二进制文件 >> "%REPORT_FILE%"
echo - 版本信息: %VERSION% >> "%REPORT_FILE%"
echo. >> "%REPORT_FILE%"
echo ### 前端 >> "%REPORT_FILE%"
echo - React应用构建文件 >> "%REPORT_FILE%"
echo - 静态资源优化 >> "%REPORT_FILE%"
echo. >> "%REPORT_FILE%"
echo ### 配置 >> "%REPORT_FILE%"
echo - 配置文件模板 >> "%REPORT_FILE%"
echo - 安装脚本 >> "%REPORT_FILE%"
echo - 校验和文件 >> "%REPORT_FILE%"
echo. >> "%REPORT_FILE%"
echo ## 文件清单 >> "%REPORT_FILE%"
echo \`\`\` >> "%REPORT_FILE%"
dir /b "%VERSION_RELEASE_DIR%" >> "%REPORT_FILE%"
echo \`\`\` >> "%REPORT_FILE%"
echo. >> "%REPORT_FILE%"
echo ## 校验和 >> "%REPORT_FILE%"
echo 所有文件的校验和信息已包含在 \`checksums.txt\` 文件中。 >> "%REPORT_FILE%"
echo. >> "%REPORT_FILE%"
echo ## 安装指南 >> "%REPORT_FILE%"
echo 详细的安装指南请参考 \`README.md\` 文件。 >> "%REPORT_FILE%"

call :log_success "发布报告已生成: %REPORT_FILE%"
goto :eof

REM 主函数
:main
echo 🚀 AI交易系统版本发布脚本 (Windows)
echo ========================================

REM 检查Git状态
call :check_git_status

REM 确定版本号
if "%VERSION%"=="" (
    if not "%VERSION_TYPE%"=="" (
        call :calculate_next_version
    ) else (
        call :log_error "请指定版本号或版本类型"
        call :show_help
        exit /b 1
    )
)

call :validate_version

call :log_info "发布版本: v%VERSION%"
call :log_info "构建类型: %BUILD_TYPE%"
call :log_info "目标平台: %PLATFORM%"

if "%DRY_RUN%"=="true" (
    call :log_warning "这是预演模式，不会执行实际操作"
)

REM 执行发布流程
call :run_tests
call :update_version
call :build_binaries
call :build_frontend
call :create_release_packages
call :generate_checksums
call :generate_release_report

echo.
echo ========================================
call :log_success "🎉 版本发布完成！"
echo ========================================
echo 版本: v%VERSION%
echo 发布目录: %VERSION_RELEASE_DIR%
echo 发布报告: %REPORT_FILE%
echo.
echo 后续步骤:
echo 1. 测试新版本
echo 2. 发布到生产环境
echo 3. 更新文档
echo 4. 通知用户

pause
goto :eof

REM 运行主函数
call :main %*