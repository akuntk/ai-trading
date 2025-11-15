[根目录](../../CLAUDE.md) > **api**

# API模块 - RESTful API服务器

## 模块职责

API模块是NOFX系统的**对外服务窗口**，提供完整的RESTful API接口，支持前端Web界面、第三方集成和系统监控，是连接用户界面与后端核心功能的重要桥梁。

## 核心功能
- 🌐 **RESTful API设计**：标准化的HTTP接口
- 🔐 **JWT认证授权**：安全的用户身份验证
- 📊 **实时数据服务**：交易状态和性能监控
- ⚙️ **配置管理**：系统配置的完整CRUD操作

## 入口与启动

### 主入口文件
- **`server.go`** - API服务器核心实现
- **`utils.go`** - 工具函数和中间件
- **`crypto_handler.go`** - 加密相关接口

### 服务器结构
```go
type Server struct {
    config     *config.Config
    db         config.DatabaseInterface
    router     *gin.Engine
    jwtSecret  string
    betaMode   bool
}
```

## 对外接口

### 系统配置接口
```go
// 获取系统配置
GET /api/system/config

// 更新系统配置
PUT /api/system/config

// 获取系统状态
GET /api/system/status
```

### 用户认证接口
```go
// 用户登录
POST /api/auth/login

// 刷新Token
POST /api/auth/refresh

// 验证Token
GET /api/auth/verify

// 2FA相关
POST /api/auth/2fa/setup
POST /api/auth/2fa/verify
```

### AI模型管理接口
```go
// 获取AI模型列表
GET /api/ai-models

// 更新AI模型配置
PUT /api/ai-models/:id

// 测试AI模型连接
POST /api/ai-models/:id/test
```

### 交易所配置接口
```go
// 获取交易所列表
GET /api/exchanges

// 更新交易所配置
PUT /api/exchanges/:id

// 测试交易所连接
POST /api/exchanges/:id/test
```

### 交易员管理接口
```go
// 获取交易员列表
GET /api/traders

// 创建交易员
POST /api/traders

// 更新交易员配置
PUT /api/traders/:id

// 启动/停止交易员
POST /api/traders/:id/start
POST /api/traders/:id/stop

// 删除交易员
DELETE /api/traders/:id

// 获取交易员日志
GET /api/traders/:id/logs

// 获取交易员性能分析
GET /api/traders/:id/performance
```

## 关键依赖与配置

### 依赖模块
- `config` - 数据库和配置管理
- `crypto` - 加密服务
- `trader` - 交易执行（通过manager）
- `decision` - AI决策引擎
- `logger` - 日志记录

### 配置参数
```go
type Config struct {
    APIServerPort int `json:"api_server_port"`
    JWTSecret     string `json:"jwt_secret"`
    BetaMode      bool `json:"beta_mode"`
}
```

## 中间件系统

### JWT认证中间件
```go
func (s *Server) authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "Missing token"})
            c.Abort()
            return
        }

        // 验证Token逻辑
        claims, err := validateJWT(token, s.jwtSecret)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        c.Set("userID", claims.UserID)
        c.Next()
    }
}
```

### CORS中间件
- 跨域请求处理
- 预检请求支持
- 安全头设置

### 日志中间件
- 请求日志记录
- 响应时间统计
- 错误追踪

### 限流中间件
- API调用频率限制
- 用户级别限流
- IP级别限制

## 数据模型

### API响应格式
```go
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
    Message string      `json:"message,omitempty"`
}
```

### 用户模型
```go
type UserResponse struct {
    ID          string    `json:"id"`
    Email       string    `json:"email"`
    OTPVerified bool      `json:"otp_verified"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### AI模型配置
```go
type AIModelConfig struct {
    ID              string `json:"id"`
    Name            string `json:"name"`
    Provider        string `json:"provider"`
    Enabled         bool   `json:"enabled"`
    CustomAPIURL    string `json:"customApiUrl"`
    CustomModelName string `json:"customModelName"`
}
```

## 认证与授权

### JWT实现
```go
type JWTClaims struct {
    UserID  string `json:"user_id"`
    Email   string `json:"email"`
    IsAdmin bool   `json:"is_admin"`
    jwt.RegisteredClaims
}

func generateJWT(userID, email string, isAdmin bool, secret string, expiration time.Time) (string, error)
func validateJWT(tokenString, secret string) (*JWTClaims, error)
```

### 2FA集成
- TOTP算法支持
- QR码生成
- 备份恢复码

### 权限控制
- 基于角色的访问控制
- 资源级别权限检查
- 管理员特权接口

## 错误处理

### 统一错误格式
```go
type APIError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}
```

### 错误分类
- **400 Bad Request** - 请求参数错误
- **401 Unauthorized** - 认证失败
- **403 Forbidden** - 权限不足
- **404 Not Found** - 资源不存在
- **429 Too Many Requests** - 限流
- **500 Internal Server Error** - 服务器错误

## 数据验证

### 请求验证
- 参数类型检查
- 必填字段验证
- 数据范围验证
- 格式规范检查

### 输入过滤
- SQL注入防护
- XSS攻击防护
- 数据清理和转义

## 性能优化

### 响应优化
- JSON压缩
- 缓存控制头
- 条件请求支持

### 数据库优化
- 连接池管理
- 查询优化
- 批量操作

### 并发处理
- 协程池管理
- 请求队列
- 超时控制

## 安全机制

### HTTPS支持
- TLS证书配置
- 安全头设置
- HSTS策略

### API安全
- 密钥管理
- 请求签名
- 重放攻击防护

### 数据保护
- 敏感数据加密
- 日志脱敏
- 数据备份

## 监控与指标

### API监控
- 请求量统计
- 响应时间监控
- 错误率分析

### 系统监控
- 服务器资源使用
- 数据库性能
- 内存使用情况

### 告警机制
- 异常告警
- 性能告警
- 安全告警

## 测试与质量

### 单元测试
- 接口逻辑测试
- 中间件测试
- 工具函数测试

### 集成测试
- API端点测试
- 认证流程测试
- 数据流测试

### 性能测试
- 负载测试
- 压力测试
- 并发测试

## 版本管理

### API版本控制
- URL版本控制 `/api/v1/`
- 头部版本控制
- 向后兼容性保证

### 变更管理
- 版本发布说明
- 废弃接口通知
- 迁移指南

## 常见问题 (FAQ)

### Q: 如何处理API认证过期？
A: 实现了JWT自动刷新机制，前端可以在Token过期前使用refresh token获取新Token。

### Q: 如何保证API调用的安全性？
A: 多层安全机制：HTTPS + JWT + 限流 + 输入验证 + 权限控制。

### Q: 如何处理高并发请求？
A: 使用连接池、协程池、缓存和限流机制来保证系统稳定性。

## 相关文件清单

```
api/
├── server.go              # API服务器核心
├── utils.go               # 工具函数和中间件
├── crypto_handler.go      # 加密相关接口
├── handlers/              # 处理器目录
│   ├── auth.go           # 认证处理器
│   ├── system.go         # 系统配置处理器
│   ├── ai_models.go      # AI模型处理器
│   ├── exchanges.go      # 交易所处理器
│   └── traders.go        # 交易员处理器
└── CLAUDE.md             # 本文档
```

## 变更记录 (Changelog)

### 2025-11-15 06:49:04 - 模块文档创建
- ✅ 完成API接口设计分析
- ✅ 认证授权机制文档
- ✅ 安全和性能优化说明