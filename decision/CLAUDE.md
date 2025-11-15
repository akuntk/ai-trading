[根目录](../../CLAUDE.md) > **decision**

# Decision模块 - AI决策引擎

## 模块职责

Decision模块是NOFX系统的**智能大脑**，负责市场分析、交易决策生成、多AI模型竞争，是整个系统实现自动化交易的核心决策中心。

## 核心功能
- 🤖 **多AI模型支持**：DeepSeek、Qwen、OpenAI兼容API
- 🧠 **智能决策生成**：基于市场数据的交易信号分析
- 📝 **提示词工程**：模块化提示词模板和自定义策略
- 🔄 **自我博弈进化**：多模型竞争和性能评估

## 入口与启动

### 主入口文件
- **`engine.go`** - AI决策引擎核心实现
- **`prompt_manager.go`** - 提示词模板管理器

### 核心结构体
```go
type FullDecision struct {
    SystemPrompt         string     `json:"system_prompt"`
    UserPrompt           string     `json:"user_prompt"`
    CoTTrace             string     `json:"cot_trace"`
    Decisions            []Decision `json:"decisions"`
    Timestamp            time.Time  `json:"timestamp"`
    AIRequestDurationMs  int64      `json:"ai_request_duration_ms"`
}

type Decision struct {
    Symbol          string  `json:"symbol"`
    Action          string  `json:"action"`
    Leverage        int     `json:"leverage,omitempty"`
    PositionSizeUSD float64 `json:"position_size_usd,omitempty"`
    StopLoss        float64 `json:"stop_loss,omitempty"`
    TakeProfit      float64 `json:"take_profit,omitempty"`
    Confidence      int     `json:"confidence,omitempty"`
    Reasoning       string  `json:"reasoning"`
}
```

## 对外接口

### 决策生成接口
```go
// 获取完整AI决策
func GetFullDecision(ctx *Context, mcpClient *mcp.Client) (*FullDecision, error)

// 使用自定义提示词获取决策
func GetFullDecisionWithCustomPrompt(
    ctx *Context,
    mcpClient *mcp.Client,
    customPrompt string,
    overrideBase bool,
    templateName string
) (*FullDecision, error)
```

### 上下文构建接口
```go
// 构建系统提示词
func buildSystemPrompt(accountEquity float64, btcEthLeverage, altcoinLeverage int, templateName string) string

// 构建用户输入提示词
func buildUserPrompt(ctx *Context) string

// 获取市场数据用于决策
func fetchMarketDataForContext(ctx *Context) error
```

## 关键依赖与配置

### 依赖模块
- `market` - 市场数据和技术指标
- `trader` - 交易执行和账户信息
- `mcp` - AI模型通信协议
- `config` - 系统配置和提示词模板

### AI模型配置
- DeepSeek API集成
- Qwen API集成
- 自定义OpenAI兼容API
- 模型竞争和选择机制

## 决策上下文 (Context)

### 交易上下文结构
```go
type Context struct {
    CurrentTime      string                  `json:"current_time"`
    RuntimeMinutes   int                     `json:"runtime_minutes"`
    CallCount        int                     `json:"call_count"`
    Account          AccountInfo             `json:"account"`
    Positions        []PositionInfo          `json:"positions"`
    CandidateCoins   []CandidateCoin         `json:"candidate_coins"`
    MarketDataMap    map[string]*market.Data `json:"-"`
    OITopDataMap     map[string]*OITopData   `json:"-"`
    Performance      interface{}             `json:"-"`
    BTCETHLeverage   int                     `json:"-"`
    AltcoinLeverage  int                     `json:"-"`
}
```

### 候选币种管理
- AI500币种池信号
- OI Top持仓增长信号
- 流动性过滤机制（15M USD持仓价值门槛）
- 动态候选数量调整

## 提示词工程

### 模板系统
- **default** - 基础交易策略模板
- **自定义模板** - 用户个性化策略
- **动态组合** - 基础 + 自定义混合模式

### 硬约束规则
- 风险回报比 ≥ 1:3
- 最多持仓3个币种
- 单币仓位限制（山寨币1.5倍净值，BTC/ETH 10倍净值）
- 杠杆限制（配置可调）
- 最小开仓金额 ≥ 12 USDT

### 输出格式标准化
- XML标签分离：`<reasoning>` 和 `<decision>`
- JSON决策数组
- 严格的字段验证
- 安全回退机制

## AI响应解析

### 多层解析策略
1. **思维链提取**：优先使用`<reasoning>`标签
2. **决策提取**：从`<decision>`标签获取JSON
3. **格式修复**：处理全角字符和格式错误
4. **安全回退**：失败时生成等待决策

### 错误处理机制
```go
func parseFullDecisionResponse(aiResponse string, accountEquity float64, btcEthLeverage, altcoinLeverage int) (*FullDecision, error) {
    // 1. 提取思维链
    cotTrace := extractCoTTrace(aiResponse)

    // 2. 提取JSON决策
    decisions, err := extractDecisions(aiResponse)

    // 3. 验证决策
    if err := validateDecisions(decisions, accountEquity, btcEthLeverage, altcoinLeverage); err != nil {
        // 生成安全回退决策
    }

    return &FullDecision{CoTTrace: cotTrace, Decisions: decisions}, nil
}
```

## 决策验证系统

### 风险控制验证
- 杠杆上限检查
- 仓位大小验证
- 最小开仓金额检查
- 风险回报比计算（必须 ≥ 3.0）

### 动态调整机制
```go
// 杠杆超限自动修正
if d.Leverage > maxLeverage {
    log.Printf("⚠️ 杠杆超限，自动调整为上限值 %dx", maxLeverage)
    d.Leverage = maxLeverage
}
```

### 交易操作类型
- `open_long/open_short` - 开仓操作
- `close_long/close_short` - 平仓操作
- `update_stop_loss/update_take_profit` - 动态调整
- `partial_close` - 部分平仓
- `hold/wait` - 观望状态

## 市场数据处理

### 技术指标集成
- EMA20/50 趋势分析
- MACD 动量指标
- RSI7/14 超买超卖
- ATR3/14 波动率分析
- 成交量和持仓量分析

### 多时间框架分析
- 3分钟K线：日内短期分析
- 4小时K线：中长期趋势判断
- 实时价格：执行时机选择

### 流动性过滤
- OI持仓价值门槛：15M USD
- 市场深度评估
- 滑点风险控制

## 性能优化

### 并发处理
- 市场数据并发获取
- AI模型并发调用
- 结果聚合和排序

### 缓存机制
- 市场数据缓存（1分钟TTL）
- 提示词模板缓存
- 决策结果缓存

### 响应时间优化
- API调用超时控制
- 数据预加载
- 批量处理机制

## 安全机制

### 输入验证
- AI响应格式验证
- 数值范围检查
- 恶意输入过滤

### 异常处理
- API调用失败处理
- 数据异常检测
- 决策冲突解决

## 监控与日志

### 决策质量监控
- 决策成功率统计
- 盈亏表现分析
- 模型性能对比

### 性能监控
- AI响应时间
- 决策生成延迟
- 资源使用情况

## 测试与质量

### 单元测试
- 决策解析测试
- 验证逻辑测试
- 边界条件测试

### 集成测试
- 端到端决策流程
- 多模型竞争测试
- 异常场景处理

### 回测验证
- 历史数据回测
- 策略效果评估
- 参数优化分析

## 常见问题 (FAQ)

### Q: 如何处理AI模型响应格式不一致？
A: 实现了多层解析策略，包括格式修复、字符转换和安全回退机制。

### Q: 多AI模型竞争如何实现？
A: 通过Context中的Performance字段记录历史表现，动态选择最优模型。

### Q: 如何确保决策的安全性？
A: 多层验证机制：硬约束检查、风险控制验证、动态调整限制。

## 相关文件清单

```
decision/
├── engine.go              # AI决策引擎核心
├── prompt_manager.go      # 提示词模板管理
├── validate_test.go       # 验证逻辑测试
├── templates/             # 提示词模板目录
│   ├── default.txt        # 默认交易策略模板
│   └── ...                # 其他模板文件
└── CLAUDE.md             # 本文档
```

## 变更记录 (Changelog)

### 2025-11-15 06:49:04 - 模块文档创建
- ✅ 完成决策引擎架构分析
- ✅ 提示词工程文档
- ✅ 验证和安全机制说明