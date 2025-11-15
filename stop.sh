#!/bin/bash

# AI交易系统服务停止脚本

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

# 停止后端服务
stop_backend() {
    log_info "停止后端API服务器..."

    if [ -f ".backend_pid" ]; then
        BACKEND_PID=$(cat .backend_pid)
        if ps -p $BACKEND_PID > /dev/null 2>&1; then
            kill $BACKEND_PID
            sleep 2

            # 强制杀死如果还在运行
            if ps -p $BACKEND_PID > /dev/null 2>&1; then
                kill -9 $BACKEND_PID
            fi

            log_success "后端服务已停止 (PID: $BACKEND_PID)"
        else
            log_warning "后端服务进程不存在"
        fi
        rm -f .backend_pid
    else
        log_warning "未找到后端服务PID文件"
    fi

    # 查找并停止可能的残留进程
    PIDS=$(pgrep -f "nofx" || true)
    if [ ! -z "$PIDS" ]; then
        echo $PIDS | xargs kill -9 2>/dev/null || true
        log_success "已清理残留的后端进程"
    fi
}

# 停止前端服务
stop_frontend() {
    log_info "停止前端开发服务器..."

    if [ -f ".frontend_pid" ]; then
        FRONTEND_PID=$(cat .frontend_pid)
        if ps -p $FRONTEND_PID > /dev/null 2>&1; then
            kill $FRONTEND_PID
            sleep 2

            # 强制杀死如果还在运行
            if ps -p $FRONTEND_PID > /dev/null 2>&1; then
                kill -9 $FRONTEND_PID
            fi

            log_success "前端服务已停止 (PID: $FRONTEND_PID)"
        else
            log_warning "前端服务进程不存在"
        fi
        rm -f .frontend_pid
    else
        log_warning "未找到前端服务PID文件"
    fi

    # 查找并停止可能的残留进程
    PIDS=$(pgrep -f "npm run dev" || true)
    if [ ! -z "$PIDS" ]; then
        echo $PIDS | xargs kill -9 2>/dev/null || true
        log_success "已清理残留的前端进程"
    fi

    # 停止可能存在的vite进程
    PIDS=$(pgrep -f "vite" || true)
    if [ ! -z "$PIDS" ]; then
        echo $PIDS | xargs kill -9 2>/dev/null || true
        log_success "已清理vite进程"
    fi
}

# 检查端口占用
check_ports() {
    log_info "检查端口占用情况..."

    # 检查8080端口
    if lsof -i :8080 > /dev/null 2>&1; then
        log_warning "端口8080仍被占用"
        lsof -i :8080
    else
        log_success "端口8080已释放"
    fi

    # 检查3000端口
    if lsof -i :3000 > /dev/null 2>&1; then
        log_warning "端口3000仍被占用"
        lsof -i :3000
    else
        log_success "端口3000已释放"
    fi
}

# 主函数
main() {
    echo "🛑 AI交易系统服务停止脚本"
    echo "========================================"

    stop_backend
    stop_frontend
    check_ports

    echo ""
    echo "========================================="
    log_success "✅ 所有服务已停止！"
    echo "========================================="
}

# 运行主函数
main "$@"