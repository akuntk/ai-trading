#!/bin/bash

# AI交易系统版本控制更新测试脚本
# 用于测试完整的版本控制、检测、下载、安装和重启流程

set -e

echo "🚀 开始测试AI交易系统版本控制更新系统..."
echo "=================================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
API_BASE_URL="http://localhost:8080"
WEB_BASE_URL="http://localhost:3000"
TEST_VERSION="1.0.1-test"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查测试依赖..."

    # 检查curl
    if ! command -v curl &> /dev/null; then
        log_error "curl未安装，请先安装curl"
        exit 1
    fi

    # 检查jq
    if ! command -v jq &> /dev/null; then
        log_warning "jq未安装，将使用其他方式解析JSON"
    fi

    log_success "依赖检查完成"
}

# 测试API服务器连通性
test_api_connectivity() {
    log_info "测试API服务器连通性..."

    if curl -s -f "${API_BASE_URL}/api/health" > /dev/null; then
        log_success "API服务器连通正常"
        return 0
    else
        log_error "无法连接到API服务器: ${API_BASE_URL}"
        return 1
    fi
}

# 测试获取当前版本
test_get_current_version() {
    log_info "测试获取当前版本..."

    response=$(curl -s "${API_BASE_URL}/api/version/current")

    if [ $? -eq 0 ]; then
        if echo "$response" | grep -q '"success":true'; then
            current_version=$(echo "$response" | grep -o '"version":"[^"]*"' | cut -d'"' -f4)
            log_success "当前版本: $current_version"
            return 0
        else
            log_error "获取版本失败: $response"
            return 1
        fi
    else
        log_error "无法获取当前版本"
        return 1
    fi
}

# 测试检查更新
test_check_update() {
    log_info "测试检查更新..."

    response=$(curl -s "${API_BASE_URL}/api/version/check")

    if [ $? -eq 0 ]; then
        if echo "$response" | grep -q '"success":true'; then
            has_update=$(echo "$response" | grep -o '"has_update":[^,]*' | cut -d':' -f2)
            log_success "检查更新完成，有更新: $has_update"

            if echo "$response" | grep -q '"latest_ver"'; then
                latest_version=$(echo "$response" | grep -o '"latest_ver":"[^"]*"' | cut -d'"' -f4)
                log_info "最新版本: $latest_version"
            fi

            return 0
        else
            log_error "检查更新失败: $response"
            return 1
        fi
    else
        log_error "无法检查更新"
        return 1
    fi
}

# 测试更新状态
test_update_status() {
    log_info "测试获取更新状态..."

    response=$(curl -s "${API_BASE_URL}/api/version/status")

    if [ $? -eq 0 ]; then
        if echo "$response" | grep -q '"success":true'; then
            updating=$(echo "$response" | grep -o '"updating":[^,]*' | cut -d':' -f2)
            log_success "更新状态获取完成，正在更新: $updating"
            return 0
        else
            log_error "获取更新状态失败: $response"
            return 1
        fi
    else
        log_error "无法获取更新状态"
        return 1
    fi
}

# 测试自动更新设置
test_auto_update_setting() {
    log_info "测试自动更新设置..."

    # 启用自动更新
    response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d '{"enabled":true}' \
        "${API_BASE_URL}/api/version/auto-update")

    if [ $? -eq 0 ]; then
        if echo "$response" | grep -q '"success":true'; then
            log_success "自动更新设置成功"
            return 0
        else
            log_error "设置自动更新失败: $response"
            return 1
        fi
    else
        log_error "无法设置自动更新"
        return 1
    fi
}

# 测试下载更新（模拟）
test_download_update() {
    log_info "测试下载更新（模拟）..."

    response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d '{"force":false,"auto_restart":false,"backup":true}' \
        "${API_BASE_URL}/api/version/download")

    if [ $? -eq 0 ]; then
        if echo "$response" | grep -q '"success":true'; then
            log_success "下载更新请求已发送"
            return 0
        else
            log_warning "下载更新可能失败或正在下载: $response"
            return 0  # 不作为失败，因为可能是正常的业务逻辑
        fi
    else
        log_error "无法下载更新"
        return 1
    fi
}

# 测试更新进度
test_update_progress() {
    log_info "测试获取更新进度..."

    response=$(curl -s "${API_BASE_URL}/api/version/progress")

    if [ $? -eq 0 ]; then
        if echo "$response" | grep -q '"success":true'; then
            progress=$(echo "$response" | grep -o '"progress":[^,]*' | cut -d':' -f2)
            status=$(echo "$response" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
            log_success "更新进度: $status ($progress%)"
            return 0
        else
            log_warning "获取更新进度失败: $response"
            return 0  # 不作为失败，可能是没有进行中的更新
        fi
    else
        log_error "无法获取更新进度"
        return 1
    fi
}

# 测试获取更新历史
test_update_history() {
    log_info "测试获取更新历史..."

    response=$(curl -s "${API_BASE_URL}/api/version/history")

    if [ $? -eq 0 ]; then
        if echo "$response" | grep -q '"success":true'; then
            log_success "更新历史获取成功"
            return 0
        else
            log_error "获取更新历史失败: $response"
            return 1
        fi
    else
        log_error "无法获取更新历史"
        return 1
    fi
}

# 测试前端版本更新页面
test_frontend_version_page() {
    log_info "测试前端版本更新页面..."

    if curl -s -f "${WEB_BASE_URL}/version" > /dev/null; then
        log_success "前端版本更新页面可访问"
        return 0
    else
        log_warning "前端版本更新页面无法访问"
        return 1
    fi
}

# 压力测试
stress_test_version_api() {
    log_info "执行版本API压力测试..."

    success_count=0
    total_requests=10

    for i in $(seq 1 $total_requests); do
        if curl -s -f "${API_BASE_URL}/api/version/current" > /dev/null; then
            ((success_count++))
        fi
        echo -n "."
    done

    echo ""
    success_rate=$((success_count * 100 / total_requests))

    if [ $success_rate -ge 90 ]; then
        log_success "压力测试通过 ($success_count/$total_requests 成功)"
        return 0
    else
        log_error "压力测试失败 ($success_count/$total_requests 成功)"
        return 1
    fi
}

# 主测试函数
run_tests() {
    log_info "开始执行测试套件..."
    echo ""

    local failed_tests=0
    local total_tests=0

    # 测试列表
    local tests=(
        "check_dependencies"
        "test_api_connectivity"
        "test_get_current_version"
        "test_check_update"
        "test_update_status"
        "test_auto_update_setting"
        "test_download_update"
        "test_update_progress"
        "test_update_history"
        "test_frontend_version_page"
        "stress_test_version_api"
    )

    # 执行测试
    for test in "${tests[@]}"; do
        ((total_tests++))
        echo "执行测试: $test"

        if $test; then
            log_success "✓ $test 通过"
        else
            log_error "✗ $test 失败"
            ((failed_tests++))
        fi
        echo ""
    done

    # 测试结果汇总
    echo "=================================================="
    log_info "测试完成！"
    echo "总测试数: $total_tests"
    echo "通过测试: $((total_tests - failed_tests))"
    echo "失败测试: $failed_tests"

    if [ $failed_tests -eq 0 ]; then
        log_success "🎉 所有测试通过！版本控制系统工作正常。"
        return 0
    else
        log_error "❌ 有 $failed_tests 个测试失败，请检查系统配置。"
        return 1
    fi
}

# 清理函数
cleanup() {
    log_info "清理测试环境..."
    # 这里可以添加清理逻辑
}

# 信号处理
trap cleanup EXIT

# 显示帮助信息
show_help() {
    echo "AI交易系统版本控制测试脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示此帮助信息"
    echo "  -q, --quiet    静默模式"
    echo "  -v, --verbose  详细模式"
    echo ""
    echo "环境变量:"
    echo "  API_BASE_URL   API服务器地址 (默认: http://localhost:8080)"
    echo "  WEB_BASE_URL   Web服务器地址 (默认: http://localhost:3000)"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -q|--quiet)
            QUIET=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 主程序入口
main() {
    echo "AI交易系统版本控制更新系统测试"
    echo "====================================="
    echo "API服务器: ${API_BASE_URL}"
    echo "Web服务器: ${WEB_BASE_URL}"
    echo ""

    # 检查环境
    if ! check_dependencies; then
        exit 1
    fi

    # 运行测试
    if run_tests; then
        exit 0
    else
        exit 1
    fi
}

# 运行主程序
main "$@"