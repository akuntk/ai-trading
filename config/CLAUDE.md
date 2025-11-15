[根目录](../../CLAUDE.md) > **config**

# Config模块 - 数据库配置中心

## 模块职责

Config模块是NOFX系统的**数据持久化层**，负责所有配置数据的存储、管理、加密，支持多用户系统、敏感数据加密、数据库迁移和配置同步。

## 核心功能
- 🗄️ **数据库管理**：SQLite WAL模式，高性能并发
- 🔐 **数据加密存储**：API密钥、私钥等敏感信息加密
- 👥 **多用户支持**：用户隔离和权限管理
- ⚙️ **配置中心**：系统配置、AI模型、交易所、交易员配置

## 入口与启动

### 主入口文件
- **`database.go`** - 数据库核心实现和接口定义
- **`config.go`** - 配置文件加载和结构定义

### 核心接口
```go
type DatabaseInterface interface {
    SetCryptoService(cs *crypto.CryptoService)
    CreateUser(user *User) error
    GetUserByEmail(email string) (*User, error)
    GetAIModels(userID string) ([]*AIModelConfig, error)
    UpdateAIModel(userID, id string, enabled bool, apiKey, customAPIURL, customModelName string) error
    GetExchanges(userID string) ([]*ExchangeConfig, error)
    UpdateExchange(userID, id string, enabled bool, apiKey, secretKey string, testnet bool, ...) error
    // ... 更多接口
}
```

## 数据库架构

### 表结构设计
```sql
-- 用户表
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    otp_secret TEXT,
    otp_verified BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- AI模型配置表
CREATE TABLE ai_models (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL DEFAULT 'default',
    name TEXT NOT NULL,
    provider TEXT NOT NULL,
    enabled BOOLEAN DEFAULT 0,
    api_key TEXT DEFAULT '',
    custom_api_url TEXT DEFAULT '',
    custom_model_name TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 交易所配置表
CREATE TABLE exchanges (
    id TEXT NOT NULL,
    user_id TEXT NOT NULL DEFAULT 'default',
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    enabled BOOLEAN DEFAULT 0,
    api_key TEXT DEFAULT '',
    secret_key TEXT DEFAULT '',
    testnet BOOLEAN DEFAULT 0,
    hyperliquid_wallet_addr TEXT DEFAULT '',
    aster_user TEXT DEFAULT '',
    aster_signer TEXT DEFAULT '',
    aster_private_key TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 交易员配置表
CREATE TABLE traders (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL DEFAULT 'default',
    name TEXT NOT NULL,
    ai_model_id TEXT NOT NULL,
    exchange_id TEXT NOT NULL,
    initial_balance REAL NOT NULL,
    scan_interval_minutes INTEGER DEFAULT 3,
    is_running BOOLEAN DEFAULT 0,
    btc_eth_leverage INTEGER DEFAULT 5,
    altcoin_leverage INTEGER DEFAULT 5,
    trading_symbols TEXT DEFAULT '',
    use_coin_pool BOOLEAN DEFAULT 0,
    use_oi_top BOOLEAN DEFAULT 0,
    custom_prompt TEXT DEFAULT '',
    override_base_prompt BOOLEAN DEFAULT 0,
    system_prompt_template TEXT DEFAULT 'default',
    is_cross_margin BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (ai_model_id) REFERENCES ai_models(id),
    FOREIGN KEY (exchange_id) REFERENCES exchanges(id)
);
```

## 数据模型

### 用户模型
```go
type User struct {
    ID           string    `json:"id"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"` // 不返回到前端
    OTPSecret    string    `json:"-"` // 不返回到前端
    OTPVerified  bool      `json:"otp_verified"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

### AI模型配置
```go
type AIModelConfig struct {
    ID              string    `json:"id"`
    UserID          string    `json:"user_id"`
    Name            string    `json:"name"`
    Provider        string    `json:"provider"`
    Enabled         bool      `json:"enabled"`
    APIKey          string    `json:"apiKey"`
    CustomAPIURL    string    `json:"customApiUrl"`
    CustomModelName string    `json:"customModelName"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
```

### 交易所配置
```go
type ExchangeConfig struct {
    ID                    string    `json:"id"`
    UserID                string    `json:"user_id"`
    Name                  string    `json:"name"`
    Type                  string    `json:"type"` // "cex" or "dex"
    Enabled               bool      `json:"enabled"`
    APIKey                string    `json:"apiKey"`
    SecretKey             string    `json:"secretKey"`
    Testnet               bool      `json:"testnet"`
    HyperliquidWalletAddr string    `json:"hyperliquidWalletAddr"`
    AsterUser             string    `json:"asterUser"`
    AsterSigner           string    `json:"asterSigner"`
    AsterPrivateKey       string    `json:"asterPrivateKey"`
    CreatedAt             time.Time `json:"created_at"`
    UpdatedAt             time.Time `json:"updated_at"`
}
```

### 交易员配置
```go
type TraderRecord struct {
    ID                   string    `json:"id"`
    UserID               string    `json:"user_id"`
    Name                 string    `json:"name"`
    AIModelID            string    `json:"ai_model_id"`
    ExchangeID           string    `json:"exchange_id"`
    InitialBalance       float64   `json:"initial_balance"`
    ScanIntervalMinutes  int       `json:"scan_interval_minutes"`
    IsRunning            bool      `json:"is_running"`
    BTCETHLeverage       int       `json:"btc_eth_leverage"`
    AltcoinLeverage      int       `json:"altcoin_leverage"`
    TradingSymbols       string    `json:"trading_symbols"`
    UseCoinPool          bool      `json:"use_coin_pool"`
    UseOITop             bool      `json:"use_oi_top"`
    CustomPrompt         string    `json:"custom_prompt"`
    OverrideBasePrompt   bool      `json:"override_base_prompt"`
    SystemPromptTemplate string    `json:"system_prompt_template"`
    IsCrossMargin        bool      `json:"is_cross_margin"`
    CreatedAt            time.Time `json:"created_at"`
    UpdatedAt            time.Time `json:"updated_at"`
}
```

## 数据库初始化

### 数据库连接配置
```go
func NewDatabase(dbPath string) (*Database, error) {
    db, err := sql.Open("sqlite", dbPath)
    if err != nil {
        return nil, fmt.Errorf("打开数据库失败: %w", err)
    }

    // 启用 WAL 模式，提高并发性能和崩溃恢复能力
    if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
        db.Close()
        return nil, fmt.Errorf("启用WAL模式失败: %w", err)
    }

    // 设置 synchronous=FULL 确保数据持久性
    if _, err := db.Exec("PRAGMA synchronous=FULL"); err != nil {
        db.Close()
        return nil, fmt.Errorf("设置synchronous失败: %w", err)
    }

    database := &Database{db: db}
    if err := database.createTables(); err != nil {
        return nil, fmt.Errorf("创建表失败: %w", err)
    }

    if err := database.initDefaultData(); err != nil {
        return nil, fmt.Errorf("初始化默认数据失败: %w", err)
    }

    return database, nil
}
```

### WAL模式优势
- **更好的并发性能**：读操作不会被写操作阻塞
- **崩溃安全**：即使在断电或强制终止时也能保证数据完整性
- **更快的写入**：不需要每次都写入主数据库文件

## 数据加密

### 加密服务集成
```go
type Database struct {
    db            *sql.DB
    cryptoService *crypto.CryptoService
}

func (d *Database) SetCryptoService(cs *crypto.CryptoService) {
    d.cryptoService = cs
}

func (d *Database) encryptSensitiveData(plaintext string) string {
    if d.cryptoService == nil || plaintext == "" {
        return plaintext
    }

    encrypted, err := d.cryptoService.EncryptForStorage(plaintext)
    if err != nil {
        log.Printf("⚠️ 加密失败: %v", err)
        return plaintext // 返回明文作为降级处理
    }

    return encrypted
}
```

### 敏感字段加密
- API密钥和私钥
- 用户密码哈希
- OTP密钥
- 交易所配置信息

## 多用户支持

### 用户隔离
- 每个配置表都包含`user_id`字段
- 数据查询自动过滤用户数据
- 默认用户系统支持

### 配置继承
```go
// 用户特定配置优先，不存在时使用default用户配置
func (d *Database) GetAIModels(userID string) ([]*AIModelConfig, error) {
    rows, err := d.db.Query(`
        SELECT id, user_id, name, provider, enabled, api_key,
               COALESCE(custom_api_url, '') as custom_api_url,
               COALESCE(custom_model_name, '') as custom_model_name,
               created_at, updated_at
        FROM ai_models WHERE user_id = ? ORDER BY id
    `, userID)
    // ...
}
```

## 数据迁移

### 表结构演进
```go
// 为现有数据库添加新字段（向后兼容）
alterQueries := []string{
    `ALTER TABLE exchanges ADD COLUMN hyperliquid_wallet_addr TEXT DEFAULT ''`,
    `ALTER TABLE exchanges ADD COLUMN aster_user TEXT DEFAULT ''`,
    `ALTER TABLE exchanges ADD COLUMN custom_prompt TEXT DEFAULT ''`,
    // ... 更多ALTER语句
}

for _, query := range alterQueries {
    // 忽略已存在字段的错误
    d.db.Exec(query)
}
```

### 交易所表迁移
```go
func (d *Database) migrateExchangesTable() error {
    // 创建新的exchanges表，使用复合主键
    _, err = d.db.Exec(`
        CREATE TABLE exchanges_new (
            id TEXT NOT NULL,
            user_id TEXT NOT NULL DEFAULT 'default',
            -- ... 其他字段
            PRIMARY KEY (id, user_id),
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        )
    `)

    // 复制数据到新表，删除旧表，重命名新表
    // ...
}
```

## 配置管理

### 系统配置表
```go
type SystemConfig struct {
    Key       string `json:"key"`
    Value     string `json:"value"`
    UpdatedAt time.Time `json:"updated_at"`
}

// 配置项示例
systemConfigs := map[string]string{
    "beta_mode":            "false",
    "api_server_port":      "8080",
    "use_default_coins":    "true",
    "default_coins":        `["BTCUSDT","ETHUSDT","SOLUSDT","BNBUSDT","XRPUSDT","DOGEUSDT","ADAUSDT","HYPEUSDT"]`,
    "max_daily_loss":       "10.0",
    "max_drawdown":         "20.0",
    "stop_trading_minutes": "60",
    "btc_eth_leverage":     "5",
    "altcoin_leverage":     "5",
    "jwt_secret":           "",
}
```

### 配置文件同步
```go
func LoadConfig(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("读取配置文件失败: %w", err)
    }

    var configFile Config
    if err := json.Unmarshal(data, &configFile); err != nil {
        return nil, fmt.Errorf("解析配置文件失败: %w", err)
    }

    return &configFile, nil
}
```

## 触发器系统

### 自动更新时间戳
```sql
CREATE TRIGGER IF NOT EXISTS update_users_updated_at
    AFTER UPDATE ON users
    BEGIN
        UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_ai_models_updated_at
    AFTER UPDATE ON ai_models
    BEGIN
        UPDATE ai_models SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

-- ... 其他表的触发器
```

## 性能优化

### 索引策略
- 主键自动索引
- 外键索引
- 查询字段复合索引

### 连接池管理
```go
// 数据库连接池配置
db.SetMaxOpenConns(100)        // 最大打开连接数
db.SetMaxIdleConns(10)         // 最大空闲连接数
db.SetConnMaxLifetime(time.Hour) // 连接最大生存时间
```

### 查询优化
- 预编译语句
- 批量操作
- 事务管理

## 备份与恢复

### 数据备份
- 定期SQLite文件备份
- 增量备份策略
- 云存储同步

### 灾难恢复
- 备份文件验证
- 快速恢复流程
- 数据一致性检查

## 监控与维护

### 性能监控
- 查询执行时间
- 数据库大小监控
- 连接池状态

### 维护任务
- 数据清理
- 索引重建
- 统计信息更新

## 测试与质量

### 单元测试
- CRUD操作测试
- 加密解密测试
- 迁移脚本测试

### 集成测试
- 多用户场景测试
- 并发访问测试
- 数据一致性测试

## 常见问题 (FAQ)

### Q: 如何处理数据库并发访问？
A: 使用WAL模式和适当的锁机制，SQLite在WAL模式下支持很好的并发读操作。

### Q: 敏感数据如何安全存储？
A: 集成加密服务，所有API密钥、私钥等敏感信息都经过加密后存储。

### Q: 如何支持配置热更新？
A: 通过系统配置表和配置文件同步机制，支持运行时配置更新。

## 相关文件清单

```
config/
├── database.go           # 数据库核心实现
├── config.go             # 配置文件加载
├── migrations/           # 数据库迁移脚本
├── seeds/               # 初始数据脚本
└── CLAUDE.md            # 本文档
```

## 变更记录 (Changelog)

### 2025-11-15 06:49:04 - 模块文档创建
- ✅ 完成数据库架构分析
- ✅ 数据模型和加密机制文档
- ✅ 多用户支持和迁移策略说明