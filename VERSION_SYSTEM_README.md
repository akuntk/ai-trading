# AI交易系统 - 版本控制与自动更新系统

## 概述

本系统为AI交易系统(NOFX)提供完整的版本控制、自动检测、下载、安装和重启功能。支持跨平台部署，具备安全验证、进度监控和用户友好的操作界面。

## 🌟 核心特性

### 🔧 版本管理
- **自动版本检测**: 定期检查远程版本服务器
- **版本比较**: 智能比较版本号和兼容性
- **安全验证**: 文件校验和验证，防止恶意更新
- **多平台支持**: Windows、Linux、macOS全平台支持

### 📥 自动更新
- **断点续传**: 支持大文件下载和断点续传
- **进度监控**: 实时显示下载进度和速度
- **备份机制**: 更新前自动备份当前版本
- **回滚支持**: 支持版本回滚到历史版本

### 🔄 重启管理
- **优雅重启**: 安全重启应用，保证数据完整性
- **倒计时提示**: 用户可取消的重启倒计时
- **自动重启**: 更新完成后自动重启应用
- **状态监控**: 实时监控重启状态

### 📱 用户界面
- **现代化UI**: 基于React的响应式界面
- **实时更新**: WebSocket实时推送更新状态
- **多语言支持**: 中英文界面切换
- **操作友好**: 简单直观的操作流程

## 🏗️ 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端界面       │    │   API服务器      │    │   版本服务器     │
│                 │    │                 │    │                 │
│ VersionUpdate   │◄──►│ VersionManager  │◄──►│ GitHub Releases │
│ 进度显示        │    │ 路由处理        │    │ 自建服务器      │
│ 用户操作        │    │ 认证授权        │    │ 版本信息        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   核心组件       │
                       │                 │
                       │ VersionChecker  │
                       │ UpdateInstaller │
                       │ RestartManager  │
                       └─────────────────┘
```

## 📁 文件结构

```
AIZH/
├── api/
│   ├── version.go              # 版本管理核心服务
│   ├── version_checker.go      # 版本检测和通知
│   ├── update_installer.go     # 更新下载和安装
│   ├── restart_manager.go      # 重启管理
│   └── server.go               # 集成到API服务器
├── web/
│   └── src/
│       ├── components/
│       │   ├── VersionUpdate.tsx    # 版本更新主界面
│       │   └── ui/                 # UI组件
│       └── routes/
│           └── index.tsx           # 路由配置
└── scripts/
    ├── test_version_system.sh   # Linux/macOS测试脚本
    └── test_version_system.bat  # Windows测试脚本
```

## 🚀 快速开始

### 1. 启动系统

```bash
# 启动后端API服务器
go run ./

# 启动前端开发服务器
cd web && npm run dev
```

### 2. 访问版本管理界面

打开浏览器访问: `http://localhost:3000/version`

### 3. 配置版本服务器

编辑配置文件，设置版本检查间隔和远程服务器地址：

```json
{
  "version_check_interval": "1h",
  "version_server_url": "https://api.github.com/repos/your-repo/nofx",
  "auto_update_enabled": true
}
```

## 📖 API接口文档

### 公开接口（无需认证）

#### 获取当前版本
```http
GET /api/version/current
```

响应示例：
```json
{
  "success": true,
  "data": {
    "version": "1.0.0",
    "build_time": "2025-11-15 06:49:04",
    "platform": "windows-amd64",
    "release_date": "2025-11-15"
  }
}
```

#### 检查更新
```http
GET /api/version/check
```

响应示例：
```json
{
  "success": true,
  "data": {
    "has_update": true,
    "current_ver": "1.0.0",
    "latest_ver": "1.0.1",
    "update_info": {
      "version": "1.0.1",
      "release_notes": "修复了已知问题...",
      "download_url": "https://releases.example.com/nofx-v1.0.1.zip",
      "checksum": "sha256:abc123...",
      "is_critical": false
    },
    "last_check": "2025-11-15T06:49:04Z"
  }
}
```

### 需要认证的接口

#### 下载更新
```http
POST /api/version/download
Content-Type: application/json

{
  "force": false,
  "auto_restart": false,
  "backup": true
}
```

#### 获取更新进度
```http
GET /api/version/progress
```

响应示例：
```json
{
  "success": true,
  "data": {
    "status": "downloading",
    "progress": 65.5,
    "message": "正在下载...",
    "speed": 1048576,
    "total_size": 52428800,
    "downloaded": 34359738,
    "eta": 15
  }
}
```

#### 重启应用
```http
POST /api/version/restart
Content-Type: application/json

{
  "delay_seconds": 10,
  "reason": "版本更新完成",
  "force": false
}
```

#### 设置自动更新
```http
POST /api/version/auto-update
Content-Type: application/json

{
  "enabled": true
}
```

## 🎨 前端界面功能

### 总览页面
- **当前版本信息**: 显示版本号、构建时间、平台信息
- **更新状态**: 显示是否有可用更新和自动更新设置
- **更新进度**: 实时显示下载/安装进度
- **操作按钮**: 下载、安装、重启等操作

### 历史记录页面
- **更新历史**: 显示所有更新记录
- **重启历史**: 显示应用重启记录
- **操作日志**: 显示详细操作日志

### 设置页面
- **自动更新**: 开启/关闭自动更新
- **检查间隔**: 设置版本检查间隔
- **手动重启**: 手动触发应用重启

## 🔧 配置说明

### 系统配置项

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `version_check_interval` | `1h` | 版本检查间隔 |
| `auto_update_enabled` | `false` | 是否启用自动更新 |
| `version_server_url` | - | 版本服务器URL |
| `backup_retention_days` | `7` | 备份保留天数 |
| `max_download_retries` | `3` | 最大下载重试次数 |

### 环境变量

| 变量名 | 说明 |
|--------|------|
| `APP_VERSION` | 应用版本号 |
| `BUILD_TIME` | 构建时间 |
| `VERSION_SERVER_URL` | 版本服务器地址 |
| `AUTO_UPDATE_ENABLED` | 是否启用自动更新 |

## 🧪 测试

### 运行测试脚本

#### Linux/macOS
```bash
chmod +x scripts/test_version_system.sh
./scripts/test_version_system.sh
```

#### Windows
```cmd
scripts\test_version_system.bat
```

### 测试覆盖范围

- ✅ API服务器连通性
- ✅ 版本信息获取
- ✅ 更新检查功能
- ✅ 更新状态监控
- ✅ 下载进度追踪
- ✅ 自动更新设置
- ✅ 历史记录查询
- ✅ 前端界面访问
- ✅ 压力测试

## 🔒 安全机制

### 文件验证
- **校验和验证**: SHA256文件完整性检查
- **数字签名**: 可选的GPG签名验证
- **文件大小**: 预期文件大小匹配检查

### 权限控制
- **用户认证**: 基于JWT的用户认证
- **操作授权**: 关键操作需要管理员权限
- **操作日志**: 记录所有版本操作日志

### 备份机制
- **自动备份**: 更新前自动创建完整备份
- **备份验证**: 验证备份文件完整性
- **回滚支持**: 支持一键回滚到历史版本

## 🚨 注意事项

### 生产环境部署
1. **备份策略**: 确保有完整的数据备份策略
2. **测试环境**: 先在测试环境验证更新流程
3. **监控告警**: 设置更新失败的监控告警
4. **权限管理**: 严格控制更新操作权限

### 版本发布流程
1. **代码审查**: 确保代码质量
2. **测试验证**: 完整的自动化测试
3. **发布准备**: 准备更新包和发布说明
4. **分批发布**: 建议分批进行版本更新

## 🐛 故障排除

### 常见问题

#### 1. 下载失败
- 检查网络连接
- 验证下载URL有效性
- 查看防火墙设置

#### 2. 安装失败
- 检查磁盘空间
- 验证文件权限
- 查看安装日志

#### 3. 重启失败
- 检查进程权限
- 验证执行文件路径
- 手动重启服务

### 日志位置

- **应用日志**: `logs/application.log`
- **更新日志**: `logs/update.log`
- **错误日志**: `logs/error.log`

## 📞 支持与反馈

如有问题或建议，请通过以下方式联系：

- 📧 邮箱: support@nofx.com
- 🐛 问题反馈: GitHub Issues
- 📖 文档: [项目文档](https://docs.nofx.com)

---

**注意**: 本版本控制系统仍在持续开发中，部分功能可能需要额外配置。请在生产环境使用前充分测试。