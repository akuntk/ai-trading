#!/bin/bash

# NOFX AI交易系统 v1.0.1 安装脚本
# 适用于 Linux 和 macOS 系统

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_header() {
    echo -e "${BLUE}"
    echo "====================================="
    echo "   NOFX AI交易系统 v1.0.1 安装向导"
    echo "====================================="
    echo -e "${NC}"
}

# 检查操作系统
check_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="Linux"
        DISTRO=$(lsb_release -si 2>/dev/null || echo "Unknown")
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macOS"
    else
        print_error "不支持的操作系统: $OSTYPE"
        exit 1
    fi

    print_info "检测到操作系统: $OS"
    if [[ "$OS" == "Linux" ]]; then
        print_info "发行版: $DISTRO"
    fi
}

# 检查依赖
check_dependencies() {
    print_info "检查系统依赖..."

    # 检查基本命令
    local missing_deps=()

    for cmd in curl wget tar; do
        if ! command -v $cmd &> /dev/null; then
            missing_deps+=($cmd)
        fi
    done

    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        print_error "缺少依赖: ${missing_deps[*]}"
        print_info "请安装缺少的依赖后重新运行"
        if [[ "$OS" == "Linux" ]]; then
            print_info "Ubuntu/Debian: sudo apt-get install ${missing_deps[*]}"
            print_info "CentOS/RHEL: sudo yum install ${missing_deps[*]}"
        fi
        exit 1
    fi

    print_success "依赖检查通过"
}

# 设置安装目录
setup_install_dir() {
    INSTALL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    print_info "安装目录: $INSTALL_DIR"

    # 检查目录权限
    if [[ ! -w "$INSTALL_DIR" ]]; then
        print_error "安装目录权限不足: $INSTALL_DIR"
        print_info "请使用合适的权限运行此脚本"
        exit 1
    fi
}

# 检查必要文件
check_required_files() {
    print_info "检查必要文件..."

    local required_files=(
        "nofx"
        "config.json.example"
        "web/dist/index.html"
    )

    local missing_files=()

    for file in "${required_files[@]}"; do
        if [[ ! -f "$INSTALL_DIR/$file" ]]; then
            missing_files+=($file)
        fi
    done

    if [[ ${#missing_files[@]} -gt 0 ]]; then
        print_error "缺少必要文件:"
        for file in "${missing_files[@]}"; do
            echo "  - $file"
        done
        print_error "请确保下载完整安装包"
        exit 1
    fi

    print_success "必要文件检查通过"
}

# 配置系统
configure_system() {
    print_info "配置系统..."

    # 创建配置文件
    if [[ ! -f "$INSTALL_DIR/config.json" ]]; then
        cp "$INSTALL_DIR/config.json.example" "$INSTALL_DIR/config.json"
        print_success "配置文件已创建: config.json"
        print_warning "请编辑 config.json 文件配置您的交易参数"
    else
        print_success "配置文件已存在"
    fi

    # 设置执行权限
    chmod +x "$INSTALL_DIR/nofx" 2>/dev/null || true
    print_success "主程序执行权限已设置"
}

# 创建目录结构
create_directories() {
    print_info "创建目录结构..."

    local directories=(
        "logs"
        "backup"
        "temp"
    )

    for dir in "${directories[@]}"; do
        if [[ ! -d "$INSTALL_DIR/$dir" ]]; then
            mkdir -p "$INSTALL_DIR/$dir"
            print_success "目录已创建: $dir"
        fi
    done
}

# 检查端口占用
check_ports() {
    print_info "检查端口占用..."

    local ports=("8080" "3000")

    for port in "${ports[@]}"; do
        if command -v netstat &> /dev/null; then
            if netstat -tuln 2>/dev/null | grep -q ":$port "; then
                print_warning "端口 $port 已被占用"
                print_warning "请修改配置文件中的端口设置"
            else
                print_success "端口 $port 可用"
            fi
        elif command -v ss &> /dev/null; then
            if ss -tuln 2>/dev/null | grep -q ":$port "; then
                print_warning "端口 $port 已被占用"
            else
                print_success "端口 $port 可用"
            fi
        else
            print_warning "无法检查端口占用情况（缺少netstat或ss命令）"
        fi
    done
}

# 配置防火墙（Linux）
configure_firewall() {
    if [[ "$OS" != "Linux" ]]; then
        return
    fi

    print_info "配置防火墙规则..."

    # 检查防火墙管理工具
    if command -v ufw &> /dev/null; then
        # Ubuntu/Debian UFW
        if ! ufw status | grep -q "8080/tcp"; then
            sudo ufw allow 8080/tcp comment "NOFX API Server" 2>/dev/null || print_warning "无法配置防火墙规则，请手动配置"
            print_success "UFW防火墙规则已添加"
        else
            print_success "UFW防火墙规则已存在"
        fi
    elif command -v firewall-cmd &> /dev/null; then
        # CentOS/RHEL firewalld
        if ! sudo firewall-cmd --list-ports | grep -q "8080/tcp"; then
            sudo firewall-cmd --add-port=8080/tcp --permanent 2>/dev/null || print_warning "无法配置防火墙规则，请手动配置"
            sudo firewall-cmd --reload 2>/dev/null || true
            print_success "firewalld防火墙规则已添加"
        else
            print_success "firewalld防火墙规则已存在"
        fi
    else
        print_warning "未检测到防火墙管理工具，请手动配置"
    fi
}

# 创建启动脚本
create_startup_script() {
    print_info "创建启动脚本..."

    local startup_script="$INSTALL_DIR/start.sh"

    cat > "$startup_script" << 'EOF'
#!/bin/bash

# NOFX AI交易系统启动脚本

INSTALL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$INSTALL_DIR"

echo "🚀 启动 NOFX AI交易系统..."

# 检查配置文件
if [[ ! -f "config.json" ]]; then
    echo "❌ 错误: 配置文件不存在，请先运行安装脚本"
    exit 1
fi

# 启动主程序
./nofx

EOF

    chmod +x "$startup_script"
    print_success "启动脚本已创建: start.sh"
}

# 创建系统服务（可选）
create_systemd_service() {
    if [[ "$OS" != "Linux" ]]; then
        return
    fi

    if ! command -v systemctl &> /dev/null; then
        return
    fi

    print_info "是否创建系统服务? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        return
    fi

    local service_file="/etc/systemd/system/nofx.service"

    if sudo tee "$service_file" > /dev/null << EOF
[Unit]
Description=NOFX AI Trading System
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/nofx
Restart=always
RestartSec=10
Environment=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

[Install]
WantedBy=multi-user.target
EOF
    then
        sudo systemctl daemon-reload
        print_success "系统服务已创建: nofx.service"
        print_info "使用以下命令管理服务:"
        print_info "  启动: sudo systemctl start nofx"
        print_info "  停止: sudo systemctl stop nofx"
        print_info "  开机自启: sudo systemctl enable nofx"
    else
        print_warning "系统服务创建失败"
    fi
}

# 安装完成
installation_complete() {
    print_header
    echo -e "${GREEN}          安装完成!${NC}"
    echo "====================================="
    echo
    echo -e "${BLUE}📁 安装目录:${NC} $INSTALL_DIR"
    echo -e "${BLUE}🎯 启动方法:${NC}"
    echo "   1. 运行启动脚本: ./start.sh"
    echo "   2. 直接运行程序: ./nofx"
    echo
    echo -e "${BLUE}🌐 访问地址:${NC}"
    echo "   API服务器: http://localhost:8080"
    echo "   Web界面:   http://localhost:3000"
    echo
    echo -e "${BLUE}📋 下一步操作:${NC}"
    echo "   1. 编辑 config.json 配置交易参数"
    echo "   2. 启动系统进行测试"
    echo "   3. 查看 README.md 了解更多功能"
    echo
    echo -e "${BLUE}📞 技术支持:${NC} support@nofx.com"
    echo
}

# 主函数
main() {
    print_header

    # 检查是否为root用户
    if [[ $EUID -eq 0 ]]; then
        print_warning "不建议以root用户运行安装脚本"
        read -p "是否继续? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi

    check_os
    check_dependencies
    setup_install_dir
    check_required_files
    configure_system
    create_directories
    check_ports
    configure_firewall
    create_startup_script
    create_systemd_service
    installation_complete
}

# 运行主函数
main "$@"