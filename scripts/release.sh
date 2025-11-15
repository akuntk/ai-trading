#!/bin/bash

# AI交易系统版本发布脚本
# 用于自动化构建、打包和发布新版本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 项目信息
PROJECT_NAME="nofx"
GITHUB_REPO="your-org/nofx"  # 替换为实际的GitHub仓库
VERSION_FILE="version.txt"
CHANGELOG_FILE="CHANGELOG.md"

# 日志函数
log_info() {
    echo -e "${BLUE}[信息]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[成功]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[警告]${NC} $1"
}

log_error() {
    echo -e "${RED}[错误]${NC} $1"
}

log_step() {
    echo -e "${PURPLE}[步骤]${NC} $1"
}

# 显示帮助信息
show_help() {
    echo "AI交易系统版本发布脚本"
    echo ""
    echo "用法: $0 [选项] [版本号]"
    echo ""
    echo "选项:"
    echo "  -h, --help        显示此帮助信息"
    echo "  -t, --type        版本类型 (major|minor|patch|pre)"
    echo "  -b, --build       构建类型 (debug|release)"
    echo "  -p, --platform    目标平台 (all|windows|linux|darwin)"
    echo "  -s, --skip-tests  跳过测试"
    echo "  -d, --dry-run     仅显示将要执行的操作，不实际执行"
    echo "  --no-git          跳过Git操作"
    echo "  --no-github       跳过GitHub发布"
    echo ""
    echo "示例:"
    echo "  $0 1.1.0                    # 发布1.1.0版本"
    echo "  $0 -t minor                 # 发布下一个次版本"
    echo "  $0 -t patch --no-git       # 发布补丁版本，跳过Git"
    echo "  $0 -t major --dry-run       # 预览主版本发布"
}

# 解析命令行参数
VERSION=""
VERSION_TYPE=""
BUILD_TYPE="release"
PLATFORM="all"
SKIP_TESTS=false
DRY_RUN=false
SKIP_GIT=false
SKIP_GITHUB=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -t|--type)
            VERSION_TYPE="$2"
            shift 2
            ;;
        -b|--build)
            BUILD_TYPE="$2"
            shift 2
            ;;
        -p|--platform)
            PLATFORM="$2"
            shift 2
            ;;
        -s|--skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        --no-git)
            SKIP_GIT=true
            shift
            ;;
        --no-github)
            SKIP_GITHUB=true
            shift
            ;;
        -*)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
        *)
            if [[ -z "$VERSION" ]]; then
                VERSION="$1"
            else
                log_error "只能指定一个版本号"
                exit 1
            fi
            shift
            ;;
    esac
done

# 检查Git状态
check_git_status() {
    if [[ "$SKIP_GIT" == "true" ]]; then
        return 0
    fi

    log_step "检查Git状态..."

    # 检查是否有未提交的更改
    if [[ -n $(git status --porcelain) ]]; then
        log_error "有未提交的更改，请先提交所有更改"
        git status --short
        exit 1
    fi

    # 检查当前分支
    CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
    if [[ "$CURRENT_BRANCH" != "main" && "$CURRENT_BRANCH" != "master" ]]; then
        log_warning "当前分支为 $CURRENT_BRANCH，建议在main或master分支发布"
        read -p "是否继续？(y/n): " choice
        if [[ "$choice" != "y" && "$choice" != "Y" ]]; then
            exit 1
        fi
    fi

    # 获取最新远程更新
    log_info "获取最新远程更新..."
    git fetch origin
    if [[ $(git rev-parse HEAD) != $(git rev-parse origin/$CURRENT_BRANCH) ]]; then
        log_error "本地分支与远程分支不同步，请先pull最新更改"
        exit 1
    fi

    log_success "Git状态检查完成"
}

# 获取当前版本
get_current_version() {
    if [[ -f "$VERSION_FILE" ]]; then
        cat "$VERSION_FILE"
    else
        echo "1.0.0"
    fi
}

# 计算下一个版本号
calculate_next_version() {
    local current_version=$(get_current_version)
    local major=$(echo $current_version | cut -d. -f1)
    local minor=$(echo $current_version | cut -d. -f2)
    local patch=$(echo $current_version | cut -d. -f3)

    case "$VERSION_TYPE" in
        "major")
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        "minor")
            minor=$((minor + 1))
            patch=0
            ;;
        "patch")
            patch=$((patch + 1))
            ;;
        "pre")
            patch=$((patch + 1))
            VERSION="${major}.${minor}.${patch}-pre"
            return
            ;;
        *)
            log_error "无效的版本类型: $VERSION_TYPE"
            exit 1
            ;;
    esac

    VERSION="${major}.${minor}.${patch}"
}

# 验证版本号格式
validate_version() {
    if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
        log_error "无效的版本号格式: $VERSION"
        log_error "版本号格式应为: major.minor.patch[-suffix]"
        exit 1
    fi
}

# 运行测试
run_tests() {
    if [[ "$SKIP_TESTS" == "true" ]]; then
        log_warning "跳过测试"
        return 0
    fi

    log_step "运行测试..."

    # 后端测试
    log_info "运行后端Go测试..."
    if ! go test ./... -v; then
        log_error "后端测试失败"
        exit 1
    fi

    # 前端测试
    log_info "运行前端测试..."
    cd web
    if ! npm test; then
        log_error "前端测试失败"
        cd ..
        exit 1
    fi
    cd ..

    log_success "所有测试通过"
}

# 更新版本信息
update_version() {
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 将更新版本到: $VERSION"
        return 0
    fi

    log_step "更新版本信息..."

    # 更新版本文件
    echo "$VERSION" > "$VERSION_FILE"

    # 更新package.json版本
    sed -i.bak "s/\"version\": \".*\"/\"version\": \"$VERSION\"/" web/package.json

    # 更新前端版本信息（如果需要）
    if [[ -f "web/src/version.ts" ]]; then
        sed -i.bak "s/export const APP_VERSION = \".*\"/export const APP_VERSION = \"$VERSION\"/" web/src/version.ts
    fi

    # 删除备份文件
    rm -f web/package.json.bak web/src/version.ts.bak 2>/dev/null

    log_success "版本信息已更新到 $VERSION"
}

# 创建Git标签和提交
create_git_release() {
    if [[ "$SKIP_GIT" == "true" ]]; then
        return 0
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 将创建Git标签: v$VERSION"
        return 0
    fi

    log_step "创建Git发布..."

    # 提交版本更改
    git add .
    git commit -m "release: version $VERSION"

    # 创建标签
    git tag -a "v$VERSION" -m "Release version $VERSION"

    # 推送到远程
    git push origin HEAD
    git push origin "v$VERSION"

    log_success "Git发布完成: v$VERSION"
}

# 构建二进制文件
build_binaries() {
    log_step "构建二进制文件..."

    BUILD_DIR="build"
    DIST_DIR="dist/v$VERSION"
    rm -rf "$BUILD_DIR" "$DIST_DIR"
    mkdir -p "$DIST_DIR"

    # 设置构建标志
    local ldflags="-X main.AppVersion=$VERSION -X main.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')"

    if [[ "$BUILD_TYPE" == "release" ]]; then
        ldflags="$ldflags -s -w"
    fi

    # 构建不同平台的二进制文件
    case "$PLATFORM" in
        "all"|"windows")
            log_info "构建Windows AMD64..."
            GOOS=windows GOARCH=amd64 go build -ldflags "$ldflags" -o "$BUILD_DIR/${PROJECT_NAME}-windows-amd64.exe" .
            if [[ $? -eq 0 ]]; then
                log_success "Windows AMD64构建成功"
            fi
            ;;
    esac

    case "$PLATFORM" in
        "all"|"linux")
            log_info "构建Linux AMD64..."
            GOOS=linux GOARCH=amd64 go build -ldflags "$ldflags" -o "$BUILD_DIR/${PROJECT_NAME}-linux-amd64" .
            if [[ $? -eq 0 ]]; then
                log_success "Linux AMD64构建成功"
            fi
            ;;
    esac

    case "$PLATFORM" in
        "all"|"darwin")
            log_info "构建macOS AMD64..."
            GOOS=darwin GOARCH=amd64 go build -ldflags "$ldflags" -o "$BUILD_DIR/${PROJECT_NAME}-darwin-amd64" .
            if [[ $? -eq 0 ]]; then
                log_success "macOS AMD64构建成功"
            fi

            log_info "构建macOS ARM64..."
            GOOS=darwin GOARCH=arm64 go build -ldflags "$ldflags" -o "$BUILD_DIR/${PROJECT_NAME}-darwin-arm64" .
            if [[ $? -eq 0 ]]; then
                log_success "macOS ARM64构建成功"
            fi
            ;;
    esac

    log_success "二进制文件构建完成"
}

# 构建前端
build_frontend() {
    log_step "构建前端..."

    cd web

    # 安装依赖
    npm ci

    # 构建生产版本
    if [[ "$BUILD_TYPE" == "release" ]]; then
        npm run build
    else
        npm run build:dev 2>/dev/null || npm run build
    fi

    if [[ $? -eq 0 ]]; then
        log_success "前端构建成功"
    else
        log_error "前端构建失败"
        cd ..
        exit 1
    fi

    cd ..
}

# 创建发布包
create_release_packages() {
    log_step "创建发布包..."

    RELEASE_DIR="releases/v$VERSION"
    rm -rf "$RELEASE_DIR"
    mkdir -p "$RELEASE_DIR"

    # 复制二进制文件
    if [[ -d "build" ]]; then
        cp -r build/* "$RELEASE_DIR/"
    fi

    # 复制前端构建文件
    if [[ -d "web/dist" ]]; then
        mkdir -p "$RELEASE_DIR/web"
        cp -r web/dist/* "$RELEASE_DIR/web/"
    fi

    # 复制配置文件和文档
    cp config.json.example "$RELEASE_DIR/"
    cp README.md "$RELEASE_DIR/"
    cp CHANGELOG.md "$RELEASE_DIR/"

    # 创建安装脚本
    cat > "$RELEASE_DIR/install.sh" << 'EOF'
#!/bin/bash
# AI交易系统安装脚本

echo "正在安装AI交易系统..."

# 根据平台选择合适的二进制文件
OS=$(uname -s)
ARCH=$(uname -m)

BINARY_NAME=""
case "$OS" in
    "Linux")
        case "$ARCH" in
            "x86_64") BINARY_NAME="nofx-linux-amd64" ;;
            *) echo "不支持的架构: $ARCH"; exit 1 ;;
        esac
        ;;
    "Darwin")
        case "$ARCH" in
            "x86_64") BINARY_NAME="nofx-darwin-amd64" ;;
            "arm64") BINARY_NAME="nofx-darwin-arm64" ;;
            *) echo "不支持的架构: $ARCH"; exit 1 ;;
        esac
        ;;
    *) echo "不支持的操作系统: $OS"; exit 1 ;;
esac

if [[ ! -f "$BINARY_NAME" ]]; then
    echo "错误: 找不到适合的二进制文件: $BINARY_NAME"
    exit 1
fi

# 设置执行权限
chmod +x "$BINARY_NAME"

# 创建符号链接
ln -sf "$BINARY_NAME" nofx

echo "安装完成！"
echo "使用方法:"
echo "  ./nofx --help"
EOF

    chmod +x "$RELEASE_DIR/install.sh"

    # 创建Windows安装脚本
    cat > "$RELEASE_DIR/install.bat" << 'EOF'
@echo off
echo 正在安装AI交易系统...

REM 检查平台
if exist "nofx-windows-amd64.exe" (
    copy "nofx-windows-amd64.exe" "nofx.exe"
    echo 安装完成！
    echo 使用方法:
    echo   nofx.exe --help
) else (
    echo 错误: 找不到适合的Windows二进制文件
    pause
    exit /b 1
)
EOF

    log_success "发布包已创建: $RELEASE_DIR"
}

# 生成校验和
generate_checksums() {
    log_step "生成文件校验和..."

    cd "$RELEASE_DIR"

    # 生成SHA256校验和
    sha256sum * > checksums.txt

    # 生成MD5校验和
    md5sum * > checksums.md5 2>/dev/null || openssl md5 * > checksums.md5 2>/dev/null

    cd ..

    log_success "校验和文件已生成"
}

# 创建GitHub发布
create_github_release() {
    if [[ "$SKIP_GITHUB" == "true" ]]; then
        return 0
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 将创建GitHub发布: v$VERSION"
        return 0
    fi

    log_step "创建GitHub发布..."

    # 检查gh CLI是否安装
    if ! command -v gh &> /dev/null; then
        log_warning "GitHub CLI (gh) 未安装，跳过GitHub发布"
        return 0
    fi

    # 获取发布说明
    local release_notes=""
    if [[ -f "$CHANGELOG_FILE" ]]; then
        # 从CHANGELOG中提取最新版本的说明
        release_notes=$(awk "/^## \[$VERSION\]/{f=1; if(/^## \[/)exit}" "$CHANGELOG_FILE" | tail -n +2)
    fi

    if [[ -z "$release_notes" ]]; then
        release_notes="版本 $VERSION 发布更新"
    fi

    # 创建发布
    cd "$RELEASE_DIR"
    gh release create "v$VERSION" \
        --title "Release v$VERSION" \
        --notes "$release_notes" \
        --draft=false \
        --prerelease=false \
        ./*

    cd ..

    log_success "GitHub发布已创建: v$VERSION"
}

# 更新版本服务器信息
update_version_server() {
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 将更新版本服务器信息"
        return 0
    fi

    log_step "更新版本服务器信息..."

    # 这里可以添加更新版本服务器的逻辑
    # 例如：调用API更新版本数据库或配置文件

    log_success "版本服务器信息已更新"
}

# 生成发布报告
generate_release_report() {
    log_step "生成发布报告..."

    REPORT_FILE="reports/release-v$VERSION.md"
    mkdir -p reports

    cat > "$REPORT_FILE" << EOF
# 版本发布报告 - v$VERSION

## 发布信息

- **版本号**: v$VERSION
- **发布时间**: $(date -u '+%Y-%m-%d %H:%M:%S UTC')
- **构建类型**: $BUILD_TYPE
- **目标平台**: $PLATFORM

## 发布内容

### 后端
- Go二进制文件
- 版本信息: $VERSION

### 前端
- React应用构建文件
- 静态资源优化

### 配置
- 配置文件模板
- 安装脚本
- 校验和文件

## 文件清单

\`\`\`
EOF

    if [[ -d "releases/v$VERSION" ]]; then
        ls -la "releases/v$VERSION" >> "$REPORT_FILE"
    fi

    cat >> "$REPORT_FILE" << EOF
\`\`\`

## 校验和

所有文件的校验和信息已包含在 \`checksums.txt\` 文件中。

## 安装指南

详细的安装指南请参考 \`README.md\` 文件。

## 变更日志

$(if [[ -f "$CHANGELOG_FILE" ]]; then awk "/^## \[$VERSION\]/{f=1; if(/^## \[/)exit}" "$CHANGELOG_FILE" | tail -n +2; else echo "暂无变更日志"; fi)

EOF

    log_success "发布报告已生成: $REPORT_FILE"
}

# 主函数
main() {
    echo "🚀 AI交易系统版本发布脚本"
    echo "========================================"

    # 检查Git状态
    check_git_status

    # 确定版本号
    if [[ -z "$VERSION" ]]; then
        if [[ -n "$VERSION_TYPE" ]]; then
            calculate_next_version
        else
            log_error "请指定版本号或版本类型"
            show_help
            exit 1
        fi
    fi

    validate_version

    log_info "发布版本: v$VERSION"
    log_info "构建类型: $BUILD_TYPE"
    log_info "目标平台: $PLATFORM"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_warning "这是预演模式，不会执行实际操作"
    fi

    # 执行发布流程
    run_tests
    update_version
    create_git_release
    build_binaries
    build_frontend
    create_release_packages
    generate_checksums
    create_github_release
    update_version_server
    generate_release_report

    echo ""
    echo "========================================="
    log_success "🎉 版本发布完成！"
    echo "========================================="
    echo "版本: v$VERSION"
    echo "发布目录: releases/v$VERSION"
    echo "发布报告: reports/release-v$VERSION.md"
    echo ""
    echo "后续步骤:"
    echo "1. 测试新版本"
    echo "2. 发布到生产环境"
    echo "3. 更新文档"
    echo "4. 通知用户"
}

# 错误处理
trap 'log_error "发布过程中发生错误"; exit 1' ERR

# 运行主函数
main "$@"