# 🚀 NOFX AI交易系统 - 快速开始指南

> **版本**: v1.0.1
> **最后更新**: 2025-11-16

## 📋 安装前准备

### 系统要求
- **操作系统**: Windows 10+, Ubuntu 18.04+, macOS 10.15+
- **内存**: 最小 4GB，推荐 8GB+
- **磁盘空间**: 最小 500MB，推荐 2GB+
- **网络**: 稳定的互联网连接

### 依赖项
- **Windows**: 无额外依赖
- **Linux**: curl, wget, tar
- **macOS**: Xcode Command Line Tools

## 🎯 快速安装

### Windows 用户
```cmd
# 1. 下载并解压安装包
# 2. 双击运行安装脚本
install.bat

# 3. 或手动执行
copy config.json.example config.json
nofx.exe
```

### Linux/macOS 用户
```bash
# 1. 下载并解压安装包
tar -xzf nofx-v1.0.1-linux.tar.gz
cd nofx/

# 2. 运行安装脚本
chmod +x install.sh
./install.sh

# 3. 启动系统
./start.sh
```

## ⚙️ 基础配置

### 1. 编辑配置文件
```bash
# 复制配置模板
cp config.json.example config.json

# 编辑配置 (重要!)
nano config.json
```

### 2. 最小配置示例
```json
{
  "api_server_port": 8080,
  "web_port": 3000,
  "database": {
    "type": "sqlite",
    "path": "config.db"
  },
  "exchanges": {
    "binance": {
      "enabled": false,
      "api_key": "",
      "secret": ""
    }
  },
  "ai_models": {
    "default_provider": "deepseek",
    "deepseek": {
      "enabled": false,
      "api_key": ""
    }
  }
}
```

### 3. 配置说明
- **api_server_port**: API服务器端口 (默认8080)
- **web_port**: Web界面端口 (默认3000)
- **database**: 数据库配置，SQLite无需额外配置
- **exchanges**: 交易所配置，需要申请API密钥
- **ai_models**: AI模型配置，需要申请API密钥

## 🚀 启动系统

### 方法1: 使用安装脚本
```bash
# Windows
install.bat

# Linux/macOS
./install.sh
```

### 方法2: 直接启动
```bash
# Windows
nofx.exe

# Linux/macOS
./nofx
```

### 方法3: 使用启动脚本
```bash
./start.sh
```

## 🌐 访问系统

启动成功后，可以通过以下地址访问：

- **API服务器**: http://localhost:8080
- **Web界面**: http://localhost:3000
- **API文档**: http://localhost:8080/api/docs

## 📊 首次使用

### 1. 检查系统状态
访问 http://localhost:8080/api/system/status 查看系统状态

### 2. 配置交易所
1. 获取交易所API密钥
2. 在Web界面中配置交易所
3. 测试连接

### 3. 配置AI模型
1. 获取AI服务API密钥
2. 在Web界面中配置AI模型
3. 测试连接

### 4. 创建交易员
1. 在Web界面中创建交易员
2. 配置交易策略
3. 启动交易员

## 🔧 常用命令

### 系统管理
```bash
# 启动系统
./nofx

# 后台启动 (Linux/macOS)
nohup ./nofx > logs/app.log 2>&1 &

# 查看日志
tail -f logs/app.log

# 停止系统
pkill nofx
```

### 数据库管理
```bash
# 备份数据库
cp config.db backup/config_backup_$(date +%Y%m%d_%H%M%S).db

# 查看数据库信息
./nofx --db-info

# 运行数据库迁移
./nofx --migrate
```

### 配置管理
```bash
# 验证配置文件
./nofx --validate-config

# 重载配置
./nofx --reload-config

# 查看配置
./nofx --show-config
```

## 🆘 常见问题

### Q: 启动时提示端口被占用
```bash
# 查看端口占用
netstat -tuln | grep 8080  # Linux/macOS
netstat -an | findstr 8080  # Windows

# 修改配置文件中的端口
nano config.json
```

### Q: 无法访问Web界面
1. 检查防火墙设置
2. 确认端口是否正确
3. 查看系统日志

### Q: 数据库连接失败
1. 检查数据库文件权限
2. 确认磁盘空间充足
3. 查看错误日志

### Q: API密钥配置失败
1. 确认API密钥格式正确
2. 检查网络连接
3. 验证API权限

## 📚 更多资源

### 文档链接
- [完整文档](./README.md)
- [配置指南](./LOCAL_DEPLOYMENT_GUIDE.md)
- [API文档](./docs/API.md)
- [故障排除](./docs/TROUBLESHOOTING.md)

### 示例配置
- [配置示例](./config.json.example)
- [环境变量示例](./.env.example)
- [Docker配置](./docker-compose.yml)

### 社区支持
- 📧 邮箱: support@nofx.com
- 🐛 问题反馈: GitHub Issues
- 💬 讨论: GitHub Discussions

## 🔄 更新系统

### 自动更新
```bash
# 检查更新
./nofx --check-update

# 下载并安装更新
./nofx --update
```

### 手动更新
```bash
# 1. 备份数据
cp config.db backup/config_backup_$(date +%Y%m%d_%H%M%S).db

# 2. 下载新版本
wget https://github.com/akuntk/ai-trading/releases/latest/download/nofx-linux.tar.gz

# 3. 替换文件
tar -xzf nofx-linux.tar.gz
cp nofx-new ./nofx

# 4. 运行迁移
./nofx --migrate

# 5. 重启系统
./nofx
```

## 🛡️ 安全建议

1. **API密钥安全**
   - 不要在代码中硬编码API密钥
   - 定期更换API密钥
   - 使用最小权限原则

2. **网络安全**
   - 在生产环境中使用HTTPS
   - 配置防火墙规则
   - 限制API访问

3. **数据安全**
   - 定期备份数据库
   - 加密敏感配置信息
   - 监控系统日志

---

🎉 **恭喜！** 您已成功安装并启动NOFX AI交易系统

如需更多帮助，请查看完整文档或联系技术支持。