#!/bin/bash

# AI交易系统服务重启脚本 (Linux/macOS)
# 停止当前服务并重新部署启动

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

# 显示帮助信息
show_help() {
    echo "AI交易系统服务重启脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help        显示此帮助信息"
    echo "  -v, --verbose    详细输出"
    echo "  -f, --force      强制重启，跳过确认"
    echo "  --no-clean      跳过清理缓存"
    echo "  --no-backup     跳过备份"
    echo "  --service-only  仅重启服务，不重新部署"
    echo ""
    echo "示例:"
    echo "  $0              # 正常重启"
    echo "  $0 --force      # 强制重启"
    echo "  $0 --no-clean    # 跳过缓存清理"
    echo "  $0 --service-only # 仅重启服务"
}

# 解析命令行参数
VERBOSE=false
FORCE=false
NO_CLEAN=false
NO_BACKUP=false
SERVICE_ONLY=false

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
        -f|--force)
            FORCE=true
            shift
            ;;
        --no-clean)
            NO_CLEAN=true
            shift
            ;;
        --no-backup)
            NO_BACKUP=true
            shift
            ;;
        --service-only)
            SERVICE_ONLY=true
            shift
            ;;
        -*)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 停止当前服务
stop_services() {
    log_info "停止当前运行的服务..."

    # 调用停止脚本
    if [[ -f "stop.sh" ]]; then
        ./stop.sh
        if [[ $? -eq 0 ]]; then
            log_success "当前服务已停止"
        else
            log_error "停止服务时出错"
        fi
    else
        log_warning "未找到停止脚本，手动停止服务..."

        # 手动停止后端服务
        pkill -f "nofx" 2>/dev/null || true
        pkill -f "nofx.exe" 2>/dev/null || true

        # 手动停止前端服务
        pkill -f "npm run dev" 2>/dev/null || true
        pkill -f "vite" 2>/dev/null || true

        sleep 2
        log_info "手动停止完成"
    fi

    # 等待进程完全停止
    sleep 3
}

# 清理缓存和临时文件
cleanup_files() {
    if [[ "$NO_CLEAN" == "true" ]]; then
        log_info "跳过缓存清理"
        return 0
    fi

    log_info "清理缓存和临时文件..."

    # 清理构建文件
    if [[ -d "build" ]]; then
        if [[ "$VERBOSE" == "true" ]]; then
            echo "删除构建目录: $(du -sh build | cut -f1)"
        fi
        rm -rf build/
        log_success "构建文件已清理"
    fi

    # 清理前端缓存
    if [[ -d "web/node_modules/.cache" ]]; then
        if [[ "$VERBOSE" == "true" ]]; then
            echo "删除前端缓存: $(du -sh web/node_modules/.cache | cut -f1)"
        fi
        rm -rf web/node_modules/.cache
        log_success "前端缓存已清理"
    fi

    # 清理临时文件
    find . -name "*.tmp" -delete 2>/dev/null || true
    find . -name "temp_*" -delete 2>/dev/null || true
    find . -name ".DS_Store" -delete 2>/dev/null || true

    # 清理日志文件（可选）
    if [[ -d "logs" ]]; then
        read -p "是否清理日志文件？(y/n): " clean_logs
        if [[ "$clean_logs" == "y" || "$clean_logs" == "Y" ]]; then
            if [[ "$VERBOSE" == "true" ]]; then
                echo "清理日志文件: $(du -sh logs | cut -f1)"
            fi
            find logs -name "*.log" -delete 2>/dev/null || true
            log_success "日志文件已清理"
        fi
    fi

    # 清理Go缓存
    if command -v go &> /dev/null; then
        go clean -cache
        log_success "Go缓存已清理"
    fi

    log_success "缓存和临时文件清理完成"
}

# 创建备份
create_backup() {
    if [[ "$NO_BACKUP" == "true" ]]; then
        log_info "跳过备份创建"
        return 0
    fi

    log_info "创建备份..."

    # 创建备份目录
    BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$BACKUP_DIR"

    # 备份配置文件
    if [[ -f "config.json" ]]; then
        cp config.json "$BACKUP_DIR/"
        log_success "配置文件已备份"
    fi

    # 备份环境变量文件
    if [[ -f ".env" ]]; then
        cp .env "$BACKUP_DIR/"
        log_success "环境变量文件已备份"
    fi

    # 备份数据库文件
    if [[ -d "data" && -n "$(ls -A data 2>/dev/null)" ]]; then
        cp -r data "$BACKUP_DIR/"
        log_success "数据文件已备份"
    fi

    log_success "备份已创建: $BACKUP_DIR"
}

# 重新部署服务
deploy_services() {
    if [[ "$SERVICE_ONLY" == "true" ]]; then
        log_info "跳过重新部署，仅重启服务"
        restart_services_only
        return 0
    fi

    log_info "重新部署和启动服务..."

    # 检查部署脚本
    if [[ ! -f "deploy.sh" ]]; then
        log_error "未找到部署脚本 deploy.sh"
        echo "请确保在项目根目录运行此脚本"
        exit 1
    fi

    # 给脚本执行权限
    chmod +x deploy.sh

    # 检查强制重启
    if [[ "$FORCE" == "true" ]]; then
        log_info "强制重新部署..."
        ./deploy.sh
    else
        # 正常部署
        log_info "开始重新部署..."
        ./deploy.sh
    fi

    if [[ $? -eq 0 ]]; then
        log_success "重新部署完成"
    else
        log_error "重新部署失败"
        exit 1
    fi
}

# 仅重启服务
restart_services_only() {
    log_info "仅重启服务..."

    # 启动后端
    log_info "启动后端服务..."
    if [[ -f "nofx" ]]; then
        chmod +x nofx 2>/dev/null
        nohup ./nofx > logs/backend.log 2>&1 &
        BACKEND_PID=$!
        echo $BACKEND_PID > .backend_pid
        log_success "后端服务已启动，PID: $BACKEND_PID"
    elif [[ -f "nofx.exe" ]]; then
        nohup ./nofx.exe > logs/backend.log 2>&1 &
        BACKEND_PID=$!
        echo $BACKEND_PID > .backend_pid
        log_success "后端服务已启动，PID: $BACKEND_PID"
    else
        log_error "未找到后端可执行文件"
        return 1
    fi

    # 启动前端
    log_info "启动前端服务..."
    cd web
    nohup npm run dev > ../logs/frontend.log 2>&1 &
    FRONTEND_PID=$!
    echo $FRONTEND_PID > ../.frontend_pid
    cd ..

    if [[ $? -eq 0 ]]; then
        log_success "前端服务已启动，PID: $FRONTEND_PID"
    else
        log_error "前端服务启动失败"
        cd ..
        return 1
    fi

    # 等待服务启动
    log_info "等待服务启动..."
    sleep 8
}

# 验证服务状态
verify_services() {
    log_info "验证服务启动状态..."

    local backend_ok=false
    local frontend_ok=false

    # 检查后端服务
    log_info "检查后端API服务器..."
    if curl -s -f http://localhost:8080/api/health > /dev/null; then
        log_success "✓ 后端API服务器运行正常"
        backend_ok=true
    else
        log_warning "⚠️  后端API服务器可能未启动"
        echo "后端地址: http://localhost:8080/api/health"
        echo "后端命令: ./nofx"
        backend_ok=false
    fi

    # 检查前端服务
    log_info "检查前端开发服务器..."
    if curl -s -f http://localhost:3000 > /dev/null; then
        log_success "✓ 前端开发服务器运行正常"
        frontend_ok=true
    else
        log_warning "⚠️  前端开发服务器可能未启动"
        echo "前端地址: http://localhost:3000"
        echo "前端命令: cd web && npm run dev"
        frontend_ok=false
    fi

    # 显示验证结果
    echo ""
    echo "======================================"
    echo "🔍 服务状态验证结果"
    echo "======================================"

    if [[ "$backend_ok" == "true" ]]; then
        echo "✅ 后端API服务器: http://localhost:8080"
    else
        echo "❌ 后端API服务器: 未响应"
    fi

    if [[ "$frontend_ok" == "true" ]]; then
        echo "✅ 前端界面: http://localhost:3000"
        echo "✅ 版本管理: http://localhost:3000/version"
    else
        echo "❌ 前端界面: 未响应"
    fi

    echo ""

    # 如果服务未正常启动，显示帮助信息
    if [[ "$backend_ok" == "false" || "$frontend_ok" == "false" ]]; then
        echo ""
        echo "🔧 故障排除建议："
        echo ""
        if [[ "$backend_ok" == "false" ]]; then
            echo "后端服务："
            echo " 1. 检查后端日志: tail -f logs/backend.log"
            echo " 2. 手动启动后端: ./nofx"
            echo " 3. 检查端口占用: lsof -i :8080"
            echo ""
        fi

        if [[ "$frontend_ok" == "false" ]]; then
            echo "前端服务："
            echo "1. 检查前端日志: tail -f logs/frontend.log"
            echo "2. 手动启动前端: cd web && npm run dev"
            echo "3. 检查端口占用: lsof -i :3000"
            echo "4. 检查Node.js版本: node --version"
            echo ""
        fi

        echo "💡 提示：可以使用以下命令查看详细日志"
        echo "   tail -f logs/backend.log  # 后端日志"
        echo "   tail -f logs/frontend.log # 前端日志"
    fi

    # 返回状态
    if [[ "$backend_ok" == "true" && "$frontend_ok" == "true" ]]; then
        return 0
    else
        return 1
    fi
}

# 显示重启结果
show_restart_result() {
    echo ""
    echo "========================================="
    echo "🎉 服务重启完成！"
    echo "========================================="
    echo ""
    echo "🌐 访问地址："
    echo "- 🏠 主界面: http://localhost:3000"
    echo "- 📡 API服务: http://localhost:8080"
    echo "- 🔧 版本管理: http://localhost:3000/version"
    echo ""
    echo "📋 当前服务状态："

    # 显示进程信息
    if [[ -f ".backend_pid" ]]; then
        local backend_pid=$(cat .backend_pid)
        if ps -p $backend_pid > /dev/null 2>&1; then
            echo "- ✅ 后端进程运行中 (PID: $backend_pid)"
        else
            echo "- ❌ 后端进程未运行"
        fi
    fi

    if [[ -f ".frontend_pid" ]]; then
        local frontend_pid=$(cat .frontend_pid)
        if ps -p $frontend_pid > /dev/null 2>&1; then
            echo "- ✅ 前端进程运行中 (PID: $frontend_pid)"
        else
            echo "- ❌ 前端进程未运行"
        fi
    fi

    echo ""
    echo "🛠️ 管理命令："
    echo "- 停止服务: ./stop.sh"
    echo "- 重启服务: ./restart.sh"
    echo "- 查看后端日志: tail -f logs/backend.log"
    echo "- 查看前端日志: tail -f logs/frontend.log"
    echo "- 检查端口占用: netstat -tlnp | grep ':8080\|:3000'"
    echo ""

    # 询问是否打开浏览器
    read -p "是否自动打开浏览器？(y/n): " open_browser
    if [[ "$open_browser" == "y" || "$open_browser" == "Y" ]]; then
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

# 主函数
main() {
    echo "🔄 AI交易系统服务重启脚本"
    echo "======================================="
    echo "模式: $(if [[ "$SERVICE_ONLY" == "true" ]]; then echo "仅重启服务"; else echo "完整重启（停止+清理+重新部署）"; fi)"
    echo ""

    # 如果不是强制模式，询问确认
    if [[ "$FORCE" != "true" ]]; then
        read -p "确认重启所有服务？(y/n): " confirm
        if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
            log_info "用户取消操作"
            exit 0
        fi
    fi

    # 执行重启流程
    stop_services
    cleanup_files
    create_backup
    deploy_services

    # 验证服务状态
    if verify_services; then
        show_restart_result
    else
        log_error "服务重启失败，请检查错误日志"
        exit 1
    fi
}

# 错误处理
trap 'log_error "重启过程中发生错误"; exit 1' ERR

# 运行主函数
main "$@"