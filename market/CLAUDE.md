[根目录](../../CLAUDE.md) > **market**

# Market模块 - 市场数据引擎

## 模块职责

Market模块是NOFX系统的**市场数据中心**，负责实时数据获取、技术指标计算、市场状态监控，为AI决策引擎提供全面、准确、及时的市场信息。

## 核心功能
- 📊 **实时市场数据**：K线、价格、成交量、持仓量
- 🔬 **技术指标计算**：EMA、MACD、RSI、ATR等50+指标
- 🌐 **多交易所聚合**：统一数据格式和标准化
- ⚡ **高性能缓存**：内存缓存 + 智能更新策略

## 入口与启动

### 主入口文件
- **`data.go`** - 市场数据处理核心逻辑
- **`websocket_client.go`** - WebSocket实时数据客户端
- **`api_client.go`** - REST API数据获取
- **`monitor.go`** - 市场监控和数据质量检查

### 核心结构体
```go
type Data struct {
    Symbol             string         `json:"symbol"`
    CurrentPrice       float64        `json:"current_price"`
    PriceChange1h      float64        `json:"price_change_1h"`
    CurrentEMA20       float64        `json:"current_ema20"`
    CurrentMACD        float64        `json:"current_macd"`
    CurrentRSI7        float64        `json:"current_rsi7"`
    OpenInterest       *OIData        `json:"open_interest"`
    FundingRate        float64        `json:"funding_rate"`
    IntradaySeries     *IntradayData  `json:"intraday_series"`
    LongerTermContext  *LongerTermData `json:"longer_term_context"`
}
```

## 对外接口

### 数据获取接口
```go
// 获取标准化市场数据
func Get(symbol string) (*Data, error)

// WebSocket监控客户端
func (wsm *WSMonitorClient) GetCurrentKlines(symbol, interval string) ([]Kline, error)

// API客户端数据获取
func (apiClient *APIClient) GetKlines(symbol, interval string, limit int) ([]Kline, error)
```

### 技术指标计算
```go
// EMA计算
func calculateEMA(klines []Kline, period int) float64

// MACD计算
func calculateMACD(klines []Kline) float64

// RSI计算
func calculateRSI(klines []Kline, period int) float64

// ATR计算
func calculateATR(klines []Kline, period int) float64
```

## 关键依赖与配置

### 依赖模块
- `config` - 系统配置
- `crypto` - 加密服务
- 外部API: Binance、Hyperliquid等

### 配置参数
```go
var (
    fundingRateMap sync.Map // 资金费率缓存
    frCacheTTL     = 1 * time.Hour // 缓存TTL
)
```

## 数据模型

### K线数据结构
```go
type Kline struct {
    OpenTime   int64   `json:"open_time"`
    Open       float64 `json:"open"`
    High       float64 `json:"high"`
    Low        float64 `json:"low"`
    Close      float64 `json:"close"`
    Volume     float64 `json:"volume"`
    CloseTime  int64   `json:"close_time"`
}
```

### 持仓量数据
```go
type OIData struct {
    Latest  float64 `json:"latest"`
    Average float64 `json:"average"`
}
```

### 日内数据
```go
type IntradayData struct {
    MidPrices   []float64 `json:"mid_prices"`
    EMA20Values []float64 `json:"ema20_values"`
    MACDValues  []float64 `json:"macd_values"`
    RSI7Values  []float64 `json:"rsi7_values"`
    RSI14Values []float64 `json:"rsi14_values"`
    Volume      []float64 `json:"volume"`
    ATR14       float64   `json:"atr14"`
}
```

## 技术指标实现

### EMA (指数移动平均线)
- 支持任意周期设置
- 平滑因子优化计算
- 实时更新机制

### MACD (异同移动平均线)
- 12/26周期EMA
- 信号线和柱状图
- 趋势识别算法

### RSI (相对强弱指数)
- 7/14周期双RSI
- Wilder平滑方法
- 超买超卖区间

### ATR (平均真实波幅)
- 真实波幅计算
- 3/14周期双ATR
- 波动率分析

## WebSocket实时数据

### 数据流处理
```go
type WSMonitorClient struct {
    conn        *websocket.Conn
    klineData   sync.Map // map[string][]Kline
    subscribeCh chan string
    errorCh     chan error
}
```

### 订阅管理
- 动态订阅/取消订阅
- 自动重连机制
- 数据质量检查

### 数据缓存策略
- 多级缓存架构
- 内存使用优化
- 过期数据清理

## API数据获取

### REST API客户端
```go
type APIClient struct {
    client      *http.Client
    baseURL     string
    apiKey      string
    secretKey   string
    rateLimiter *rate.Limiter
}
```

### 限流管理
- 请求频率控制
- 优先级队列
- 智能退避策略

### 错误处理
- 网络异常重试
- API限流处理
- 数据验证检查

## 数据标准化与格式化

### 符号标准化
```go
func Normalize(symbol string) string {
    symbol = strings.ToUpper(symbol)
    if strings.HasSuffix(symbol, "USDT") {
        return symbol
    }
    return symbol + "USDT"
}
```

### 价格动态精度
- 超低价币种（< 0.0001）：8位小数
- 低价币种（< 0.01）：6位小数
- 中价币种（< 100）：4位小数
- 高价币种（≥ 100）：2位小数

## 性能优化

### 内存管理
- 对象池复用
- 内存预分配
- 垃圾回收优化

### 并发处理
- 协程池管理
- 读写锁优化
- 无锁数据结构

### 计算优化
- 增量计算算法
- 批量处理机制
- SIMD指令利用

## 数据质量保证

### 数据验证
- 价格合理性检查
- 成交量一致性验证
- 时间序列完整性

### 异常检测
- 价格突变检测
- 数据延迟监控
- 来源交叉验证

## 监控与告警

### 关键指标
- 数据更新延迟
- API调用成功率
- 缓存命中率
- 错误率统计

### 告警机制
- 数据异常告警
- 连接中断告警
- 性能下降告警

## 测试与质量

### 单元测试
- 技术指标计算测试
- 数据格式化测试
- 缓存机制测试

### 集成测试
- WebSocket连接测试
- API接口测试
- 端到端数据流测试

### 性能测试
- 高并发数据获取
- 大量指标计算
- 内存使用分析

## 常见问题 (FAQ)

### Q: 如何处理不同交易所的数据格式差异？
A: 通过统一的数据模型和标准化接口，屏蔽底层数据源差异。

### Q: 技术指标计算的准确性如何保证？
A: 使用成熟的TA-Lib库算法，并定期与第三方数据源进行交叉验证。

### Q: 如何应对API限流和连接问题？
A: 实现了智能限流、自动重连和多数据源备份机制。

## 相关文件清单

```
market/
├── data.go               # 核心数据处理逻辑
├── websocket_client.go   # WebSocket实时数据
├── api_client.go         # REST API数据获取
├── monitor.go            # 市场监控
├── types.go              # 数据类型定义
└── CLAUDE.md            # 本文档
```

## 变更记录 (Changelog)

### 2025-11-15 06:49:04 - 模块文档创建
- ✅ 完成数据模型分析
- ✅ 技术指标实现文档
- 📋 待完成：WebSocket详细实现分析