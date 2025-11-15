# 📦 NOFX AI交易系统 - 发布文件清单

> **版本**: v1.0.1
> **更新日期**: 2025-11-16
> **适用场景**: 用户首次安装、系统部署、发布打包

## 🎯 用户首次安装所需文件清单

### 📋 核心必需文件

#### 1. 可执行文件
```
├── nofx.exe              # Windows主程序
├── nofx                  # Linux/macOS主程序
└── start.sh              # 启动脚本（可选）
```

#### 2. 配置文件模板
```
├── config.json.example   # 配置文件模板
├── .env.example         # 环境变量模板
└── .dockerignore        # Docker部署配置
```

#### 3. 数据库相关
```
├── migrations/          # 数据库迁移文件目录
│   ├── 1_0_0_init_database.json
│   └── 1_1_0_add_trader_performance_index.json
└── config.db           # 初始空数据库（SQLite）
```

#### 4. 前端资源
```
├── web/
│   ├── dist/           # 编译后的前端文件
│   │   ├── index.html
│   │   ├── assets/
│   │   │   ├── main.js
│   │   │   └── main.css
│   │   └── favicon.ico
│   └── package.json    # 前端依赖信息
```

#### 5. 文档说明
```
├── README.md           # 项目说明文档
├── CHANGELOG.zh-CN.md  # 中文更新日志
├── VERSION_SYSTEM_README.md  # 版本系统说明
├── LOCAL_DEPLOYMENT_GUIDE.md # 本地部署指南
└── DEPLOYMENT.md       # 部署文档
```

#### 6. 脚本和工具
```
├── scripts/
│   ├── install.sh      # Linux/macOS安装脚本
│   ├── install.bat     # Windows安装脚本
│   └── health_check.sh # 健康检查脚本
└── deploy.sh           # 快速部署脚本
```

#### 7. 许可证和法律文件
```
├── LICENSE             # 开源许可证
├── SECURITY.md         # 安全说明
└── TERMS_OF_SERVICE.md # 服务条款
```

## 📦 发布包建议结构

### 完整版安装包
```
nofx-v1.0.1-complete/
├── README_INSTALLATION.md          # 📖 安装指南
├── QUICK_START.md                  # 🚀 快速开始
├── config.json.example            # ⚙️ 配置模板
├── install.bat                    # 🪟 Windows安装
├── install.sh                     # 🐧 Linux/macOS安装
├── nofx.exe                       # 🎯 Windows主程序
├── nofx                           # 🎯 Linux/macOS主程序
├── web/                           # 🌐 前端文件
│   └── dist/
│       ├── index.html
│       └── assets/
├── migrations/                    # 🗄️ 数据库迁移
│   └── 1_0_0_init_database.json
├── docs/                          # 📚 文档目录
│   ├── USER_GUIDE.md
│   ├── CONFIGURATION.md
│   └── TROUBLESHOOTING.md
└── tools/                         # 🔧 辅助工具
    ├── health_check.bat
    └── backup_config.sh
```

### 精简版安装包
```
nofx-v1.0.1-minimal/
├── nofx.exe                  # 主程序
├── config.json.example      # 配置模板
├── web/dist/                 # 前端文件
├── migrations/               # 数据库迁移
└── INSTALL.md               # 安装说明
```

## 🚀 快速安装流程

### 用户安装步骤
1. **下载安装包**
2. **解压到目标目录**
3. **运行安装脚本**：
   - Windows: `install.bat`
   - Linux/macOS: `chmod +x install.sh && ./install.sh`
4. **配置系统**：
   - 复制 `config.json.example` 为 `config.json`
   - 修改必要的配置项
5. **启动系统**：
   - Windows: `nofx.exe`
   - Linux/macOS: `./nofx`

## 📋 发布文件清单

### GitHub Release 文件
- `nofx-v1.0.1-windows-complete.zip` - Windows完整版
- `nofx-v1.0.1-windows-minimal.zip` - Windows精简版
- `nofx-v1.0.1-linux.tar.gz` - Linux版本
- `nofx-v1.0.1-macos.dmg` - macOS版本

### 打包命令参考
```bash
# Windows完整版
7z a -tzip nofx-v1.0.1-windows-complete.zip ^
  README_INSTALLATION.md QUICK_START.md ^
  config.json.example install.bat install.sh ^
  nofx.exe nofx ^
  web/dist/ migrations/ docs/ tools/

# Windows精简版
7z a -tzip nofx-v1.0.1-windows-minimal.zip ^
  nofx.exe config.json.example ^
  web/dist/ migrations/ INSTALL.md

# Linux版本
tar -czf nofx-v1.0.1-linux.tar.gz ^
  README_INSTALLATION.md QUICK_START.md ^
  config.json.example install.sh nofx ^
  web/dist/ migrations/ docs/ tools/
```

## ⚙️ 配置要求

### 最低系统要求
- **操作系统**: Windows 10+, Ubuntu 18.04+, macOS 10.15+
- **内存**: 最小 4GB，推荐 8GB+
- **磁盘空间**: 最小 500MB，推荐 2GB+
- **网络**: 稳定的互联网连接

### 依赖项
- **Go 1.21+** (源码编译)
- **Node.js 18+** (前端开发)
- **SQLite** (数据库，内置)
- **TA-Lib** (技术指标库，可选)

## 📝 安装后检查清单

- [ ] 可执行文件权限正确
- [ ] 配置文件已创建并配置
- [ ] 数据库文件存在且可访问
- [ ] 前端资源文件完整
- [ ] 网络端口未被占用
- [ ] 日志目录可写入
- [ ] 系统服务可正常启动

## 🔄 更新流程

1. **停止当前运行的服务**
2. **备份现有配置和数据**
3. **下载新版本安装包**
4. **替换可执行文件**
5. **运行数据库迁移**（如需要）
6. **重启服务**
7. **验证功能正常**

## 🆘 故障排除

### 常见问题
- **端口占用**: 修改配置文件中的端口设置
- **权限问题**: 检查文件和目录权限
- **数据库错误**: 确认SQLite文件权限
- **前端无法访问**: 检查dist目录是否完整

### 技术支持
- 📧 邮箱: support@nofx.com
- 🐛 问题反馈: GitHub Issues
- 📖 文档: [项目文档](https://docs.nofx.com)

---

**注意**:
- 本文档随版本更新，请确认使用对应版本的说明
- 建议在生产环境部署前先在测试环境验证
- 定期检查GitHub Releases获取最新版本

🤖 Generated with [Claude Code](https://claude.com/claude-code)