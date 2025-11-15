[根目录](../../CLAUDE.md) > **trader**

# Trader模块 - 交易执行引擎

## 模块职责

Trader模块是NOFX系统的**交易执行层**，负责将AI决策转化为实际的交易操作，支持多个主流加密货币交易所的统一接入和执行。

## 核心功能
- 🔄 **多交易所统一接口**：Binance Futures、Hyperliquid、Aster DEX
- ⚡ **低延迟执行**：优化的API调用和错误重试机制
- 🛡️ **风险控制**：仓位管理、止损止盈、强平保护
- 📊 **实时监控**：持仓跟踪、盈亏计算、执行统计

## 入口与启动

### 主入口文件
- **`auto_trader.go`** - 自动交易器主实现
- **`interface.go`** - 交易接口定义和通用实现

### 核心结构体
```go
type AutoTrader struct {
    id             string
    exchangeType   string  // "binance", "hyperliquid", "aster"
    apiClient      *rest.Client
    wsClient       *websocket.Client
    mcpClient      *mcp.Client  // AI决策客户端
    // ... 其他字段
}
```

## 对外接口

### 交易执行接口
```go
// 核心交易执行方法
func (at *AutoTrader) ExecuteDecision(decision *decision.Decision) error

// 仓位管理方法
func (at *AutoTrader) GetPositions() ([]Position, error)
func (at *AutoTrader) ClosePosition(symbol, side string) error
func (at *AutoTrader) UpdateSLTP(position *Position) error

// 账户信息查询
func (at *AutoTrader) GetAccountInfo() (*Account, error)
```

### 交易所适配器
- **Binance Futures**: `binance_futures.go`
- **Hyperliquid**: `hyperliquid_trader.go`
- **Aster DEX**: `aster_trader.go`

## 关键依赖与配置

### 依赖模块
- `decision` - AI决策引擎
- `market` - 市场数据
- `config` - 数据库配置
- `crypto` - 加密服务

### 配置需求
- API密钥和权限配置
- 杠杆和保证金设置
- 风险参数配置
- 交易符号白名单

## 数据模型

### Position结构
```go
type Position struct {
    Symbol           string  `json:"symbol"`
    Side             string  `json:"side"` // "long" or "short"
    EntryPrice       float64 `json:"entry_price"`
    MarkPrice        float64 `json:"mark_price"`
    Quantity         float64 `json:"quantity"`
    Leverage         int     `json:"leverage"`
    UnrealizedPnL    float64 `json:"unrealized_pnl"`
    LiquidationPrice float64 `json:"liquidation_price"`
}
```

### Account结构
```go
type Account struct {
    TotalWalletBalance    float64 `json:"total_wallet_balance"`
    AvailableBalance      float64 `json:"available_balance"`
    TotalUnrealizedPnl    float64 `json:"total_unrealized_pnl"`
    TotalMarginBalance    float64 `json:"total_margin_balance"`
    MaintenanceMargin     float64 `json:"maintenance_margin"`
}
```

## 交易所特定实现

### Binance Futures
- 支持USDT永续合约
- 全仓和逐仓模式
- API限流管理
- 订单状态跟踪

### Hyperliquid
- 原生DEX集成
- Agent钱包模式
- 高频交易支持
- 低延迟执行

### Aster DEX
- Solana生态DEX
- 钱包签名集成
- 智能合约交互
- 流动性聚合

## 错误处理与重试

### 错误分类
- **网络错误**: 自动重试机制
- **API限流**: 指数退避策略
- **余额不足**: 立即停止交易
- **市场异常**: 安全模式切换

### 重试策略
```go
type RetryConfig struct {
    MaxRetries    int
    InitialDelay  time.Duration
    MaxDelay      time.Duration
    BackoffFactor float64
}
```

## 性能优化

### 并发控制
- 连接池管理
- 请求队列优化
- 内存复用机制

### 缓存策略
- 仓位信息缓存
- 账户状态缓存
- 市场数据缓存

## 监控与日志

### 关键指标
- 订单执行延迟
- 成交率统计
- 错误率监控
- 仓位变化跟踪

### 日志记录
- 交易决策执行
- API调用详情
- 错误和异常
- 性能统计数据

## 安全机制

### API安全
- 密钥加密存储
- 请求签名验证
- IP白名单支持
- 2FA认证集成

### 交易安全
- 仓位大小限制
- 频率限制保护
- 异常交易检测
- 紧急停止机制

## 测试与质量

### 单元测试覆盖
- 接口实现测试
- 错误处理测试
- 边界条件测试
- 并发安全测试

### 集成测试
- 交易所连接测试
- 端到端交易流程
- 压力测试和性能测试

## 常见问题 (FAQ)

### Q: 如何添加新的交易所支持？
A: 实现Trader接口，添加相应的适配器文件，并在工厂函数中注册。

### Q: 如何处理API限流？
A: 系统内置了智能限流处理，采用指数退避策略和请求队列管理。

### Q: 如何确保交易安全？
A: 多层安全机制：加密存储、签名验证、仓位限制、异常检测。

## 相关文件清单

```
trader/
├── auto_trader.go          # 主交易器实现
├── interface.go            # 交易接口定义
├── binance_futures.go      # Binance适配器
├── hyperliquid_trader.go   # Hyperliquid适配器
├── aster_trader.go         # Aster DEX适配器
├── CLAUDE.md              # 本文档
└── types.go               # 数据类型定义
```

## 变更记录 (Changelog)

### 2025-11-15 06:49:04 - 模块文档创建
- ✅ 完成模块结构和接口分析
- ✅ 生成详细的API文档
- 📋 待完成：具体实现细节分析