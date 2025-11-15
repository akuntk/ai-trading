#!/bin/bash

# AI交易系统本地部署脚本
# 支持 Linux 和 macOS

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 检查当前目录
check_project_directory() {
    if [ ! -f "go.mod" ]; then
        log_error "请确保在项目根目录（AIZH）运行此脚本"
        exit 1
    fi
    log_success "检测到项目根目录，开始部署..."
}

# 检查Go环境
check_go_environment() {
    log_info "检查Go环境..."

    if ! command -v go &> /dev/null; then
        log_error "未检测到Go环境"
        echo "请从 https://golang.org/dl/ 下载并安装Go 1.25或更高版本"
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}')
    log_success "Go环境正常，版本：$GO_VERSION"

    # 检查Go版本是否满足要求
    GO_MAJOR=$(echo $GO_VERSION | sed 's/go//' | cut -d. -f1)
    GO_MINOR=$(echo $GO_VERSION | sed 's/go//' | cut -d. -f2)

    if [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 21 ]); then
        log_error "Go版本过低，需要Go 1.21或更高版本"
        exit 1
    fi
}

# 检查Node.js环境
check_node_environment() {
    log_info "检查Node.js环境..."

    if ! command -v node &> /dev/null; then
        log_error "未检测到Node.js环境"
        echo "请从 https://nodejs.org/ 下载并安装Node.js 18或更高版本"
        exit 1
    fi

    NODE_VERSION=$(node --version)
    log_success "Node.js环境正常，版本：$NODE_VERSION"

    # 检查Node.js版本是否满足要求
    NODE_MAJOR=$(echo $NODE_VERSION | sed 's/v//' | cut -d. -f1)

    if [ "$NODE_MAJOR" -lt 18 ]; then
        log_error "Node.js版本过低，需要Node.js 18或更高版本"
        exit 1
    fi
}

# 安装后端依赖
install_go_dependencies() {
    log_info "安装Go模块依赖..."

    go mod download
    if [ $? -ne 0 ]; then
        log_error "Go模块下载失败"
        echo "请检查网络连接或设置Go代理："
        echo "export GOPROXY=https://goproxy.cn,direct"
        exit 1
    fi

    log_success "Go依赖安装完成"
}

# 初始化配置文件
init_config_files() {
    log_info "初始化配置文件..."

    # 检查config.json
    if [ ! -f "config.json" ]; then
        log_info "创建默认配置文件..."
        cat > config.json << EOF
{
  "api_server_port": 8080,
  "beta_mode": false,
  "jwt_secret": "your-jwt-secret-key-change-in-production",
  "auto_update_enabled": true,
  "version_check_interval": "1h"
}
EOF
        log_success "配置文件已创建：config.json"
    else
        log_success "配置文件已存在：config.json"
    fi

    # 检查.env文件
    if [ ! -f ".env" ]; then
        log_info "创建环境变量文件..."
        cat > .env << EOF
# AI交易系统环境变量
APP_VERSION=1.0.0
BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')
JWT_SECRET=your-jwt-secret-key-change-in-production
BETA_MODE=false
API_PORT=8080

# 数据库配置
DB_PATH=./data/nofx.db

# 交易所API配置（请替换为实际密钥）
BINANCE_API_KEY=
BINANCE_SECRET_KEY=

# AI模型配置
DEEPSEEK_API_KEY=
QWEN_API_KEY=

# Telegram通知（可选）
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=
EOF
        log_success "环境变量文件已创建：.env"
        log_warning "⚠️  警告：请编辑.env文件，填入实际的API密钥"
    else
        log_success "环境变量文件已存在：.env"
    fi

    # 创建必要的目录
    mkdir -p data logs
    log_success "数据目录已创建"
}

# 安装前端依赖
install_node_dependencies() {
    log_info "安装前端依赖..."

    cd web
    npm install
    if [ $? -ne 0 ]; then
        log_error "前端依赖安装失败"
        echo "请检查Node.js版本和网络连接"
        cd ..
        exit 1
    fi
    log_success "前端依赖安装完成"
    cd ..
}

# 构建后端
build_backend() {
    log_info "编译后端程序..."

    # 设置编译目标
    OUTPUT_NAME="nofx"
    if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
        OUTPUT_NAME="nofx.exe"
    fi

    go build -o $OUTPUT_NAME .
    if [ $? -ne 0 ]; then
        log_error "后端编译失败"
        exit 1
    fi

    log_success "后端编译完成：$OUTPUT_NAME"
}

# 启动后端服务
start_backend() {
    log_info "启动后端API服务器..."

    # 设置可执行权限
    if [ -f "nofx" ]; then
        chmod +x nofx
    fi

    # 启动后端（后台运行）
    nohup ./nofx > logs/backend.log 2>&1 &
    BACKEND_PID=$!
    echo $BACKEND_PID > .backend_pid

    log_success "后端API服务器已启动，PID: $BACKEND_PID"

    # 等待后端启动
    log_info "等待后端服务启动..."
    sleep 5

    # 检查后端是否启动成功
    if curl -s -f http://localhost:8080/api/health > /dev/null 2>&1; then
        log_success "后端API服务器启动成功"
    else
        log_warning "⚠️  警告：后端服务可能未完全启动，请检查日志：logs/backend.log"
    fi
}

# 启动前端服务
start_frontend() {
    log_info "启动前端开发服务器..."

    cd web
    # 启动前端（后台运行）
    nohup npm run dev > ../logs/frontend.log 2>&1 &
    FRONTEND_PID=$!
    echo $FRONTEND_PID > ../.frontend_pid
    cd ..

    log_success "前端开发服务器已启动，PID: $FRONTEND_PID"

    # 等待前端启动
    sleep 3
}

# 显示部署结果
show_deployment_result() {
    echo ""
    echo "========================================="
    log_success "🎉 部署完成！"
    echo "========================================="
    echo -e "${BLUE}访问地址：${NC}"
    echo "  - 前端界面: http://localhost:3000"
    echo "  - API服务器: http://localhost:8080"
    echo "  - 版本管理: http://localhost:3000/version"
    echo ""
    echo -e "${YELLOW}默认账户信息：${NC}"
    echo "  - 管理员邮箱: admin@example.com"
    echo "  - 管理员密码: admin123"
    echo ""
    echo -e "${BLUE}重要提示：${NC}"
    echo "  1. 请编辑 .env 文件，填入实际的API密钥"
    echo "  2. 首次登录需要设置2FA验证"
    echo "  3. 在版本管理页面可以检查和安装更新"
    echo "  4. 使用 ./stop.sh 停止所有服务"
    echo "  5. 日志文件位置: logs/"
    echo ""

    # 询问是否打开浏览器
    read -p "是否自动打开浏览器？(y/n): " choice
    if [[ "$choice" == "y" || "$choice" == "Y" ]]; then
        log_info "正在打开浏览器..."
        if command -v open &> /dev/null; then
            open http://localhost:3000
        elif command -v xdg-open &> /dev/null; then
            xdg-open http://localhost:3000
        elif command -v gnome-open &> /dev/null; then
            gnome-open http://localhost:3000
        else
            log_warning "无法自动打开浏览器，请手动访问：http://localhost:3000"
        fi
    fi
}

# 显示帮助信息
show_help() {
    echo "AI交易系统本地部署脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示此帮助信息"
    echo "  -v, --verbose  详细输出"
    echo "  --dev          开发模式"
    echo "  --prod         生产模式"
    echo ""
    echo "环境变量:"
    echo "  NODE_ENV       运行环境 (development/production)"
    echo "  API_PORT       API服务器端口 (默认: 8080)"
    echo "  WEB_PORT       Web服务器端口 (默认: 3000)"
}

# 解析命令行参数
VERBOSE=false
MODE="development"

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --dev)
            MODE="development"
            shift
            ;;
        --prod)
            MODE="production"
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 主函数
main() {
    echo "🚀 AI交易系统本地部署脚本"
    echo "========================================"
    echo "运行模式: $MODE"
    echo ""

    # 执行部署步骤
    check_project_directory
    check_go_environment
    check_node_environment
    install_go_dependencies
    init_config_files
    install_node_dependencies
    build_backend
    start_backend
    start_frontend
    show_deployment_result

    log_success "部署脚本执行完成！"
}

# 错误处理
trap 'log_error "部署过程中发生错误"; exit 1' ERR

# 运行主函数
main "$@"