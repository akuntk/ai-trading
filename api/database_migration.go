package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"nofx/config"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DatabaseMigration 数据库迁移结构
type DatabaseMigration struct {
	Version     string    `json:"version"`     // 迁移版本号
	Name        string    `json:"name"`        // 迁移名称
	Description string    `json:"description"` // 迁移描述
	UpSQL       string    `json:"up_sql"`      // 升级SQL
	DownSQL     string    `json:"down_sql"`    // 回滚SQL
	Author      string    `json:"author"`      // 作者
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
	IsCritical  bool      `json:"is_critical"` // 是否为关键迁移
}

// MigrationManager 迁移管理器
type MigrationManager struct {
	database     *config.Database
	migrations   []DatabaseMigration
	migrationDir string
}

// NewMigrationManager 创建迁移管理器
func NewMigrationManager(database *config.Database) *MigrationManager {
	return &MigrationManager{
		database:     database,
		migrationDir: "./migrations",
		migrations:   make([]DatabaseMigration, 0),
	}
}

// LoadMigrations 加载所有迁移文件
func (m *MigrationManager) LoadMigrations() error {
	// 创建迁移目录
	if err := os.MkdirAll(m.migrationDir, 0755); err != nil {
		return fmt.Errorf("创建迁移目录失败: %v", err)
	}

	// 加载内置迁移
	m.loadBuiltInMigrations()

	// 加载外部迁移文件
	files, err := filepath.Glob(filepath.Join(m.migrationDir, "*.json"))
	if err != nil {
		return fmt.Errorf("读取迁移文件失败: %v", err)
	}

	for _, file := range files {
		if err := m.loadMigrationFile(file); err != nil {
			log.Printf("⚠️  加载迁移文件 %s 失败: %v", file, err)
		}
	}

	// 按版本号排序
	sort.Slice(m.migrations, func(i, j int) bool {
		return compareVersions(m.migrations[i].Version, m.migrations[j].Version) < 0
	})

	log.Printf("✅ 已加载 %d 个数据库迁移", len(m.migrations))
	return nil
}

// loadBuiltInMigrations 加载内置迁移
func (m *MigrationManager) loadBuiltInMigrations() {
	// 初始化迁移表的迁移
	initMigration := DatabaseMigration{
		Version:     "1.0.0",
		Name:        "init_migration_system",
		Description: "初始化迁移系统",
		UpSQL: `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    checksum VARCHAR(64)
);

CREATE TABLE IF NOT EXISTS migration_backups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version VARCHAR(50) NOT NULL,
    backup_data TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`,
		DownSQL: `DROP TABLE IF EXISTS schema_migrations; DROP TABLE IF EXISTS migration_backups;`,
		Author:     "System",
		CreatedAt:  time.Now(),
		IsCritical: false,
	}

	m.migrations = append(m.migrations, initMigration)
}

// loadMigrationFile 加载单个迁移文件
func (m *MigrationManager) loadMigrationFile(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	var migration DatabaseMigration
	if err := json.Unmarshal(data, &migration); err != nil {
		return err
	}

	m.migrations = append(m.migrations, migration)
	return nil
}

// GetCurrentDBVersion 获取当前数据库版本
func (m *MigrationManager) GetCurrentDBVersion() (string, error) {
	// 确保迁移表存在
	if err := m.ensureMigrationTable(); err != nil {
		return "", err
	}

	var version sql.NullString
	err := m.database.DB().QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return "1.0.0", nil // 默认版本
		}
		return "", err
	}

	if !version.Valid || version.String == "" {
		return "1.0.0", nil
	}
	return version.String, nil
}

// GetPendingMigrations 获取待执行的迁移
func (m *MigrationManager) GetPendingMigrations() ([]DatabaseMigration, error) {
	currentVersion, err := m.GetCurrentDBVersion()
	if err != nil {
		return nil, err
	}

	var pending []DatabaseMigration
	for _, migration := range m.migrations {
		if compareVersions(migration.Version, currentVersion) > 0 {
			pending = append(pending, migration)
		}
	}

	return pending, nil
}

// ExecuteMigration 执行迁移
func (m *MigrationManager) ExecuteMigration(migration DatabaseMigration, autoBackup bool) error {
	log.Printf("🔄 开始执行数据库迁移: %s (%s)", migration.Version, migration.Name)

	// 1. 检查迁移是否已执行
	if m.isMigrationApplied(migration.Version) {
		log.Printf("⚠️  迁移 %s 已经执行过，跳过", migration.Version)
		return nil
	}

	// 2. 自动备份
	if autoBackup {
		if err := m.createDatabaseBackup(migration.Version); err != nil {
			return fmt.Errorf("数据库备份失败: %v", err)
		}
	}

	// 3. 验证SQL
	if err := m.validateMigrationSQL(migration); err != nil {
		return fmt.Errorf("SQL验证失败: %v", err)
	}

	// 4. 执行事务
	tx, err := m.database.DB().Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	// 5. 执行迁移SQL
	if migration.UpSQL != "" {
		if _, err := tx.Exec(migration.UpSQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("执行迁移SQL失败: %v", err)
		}
	}

	// 6. 记录迁移
	if err := m.recordMigration(tx, migration); err != nil {
		tx.Rollback()
		return fmt.Errorf("记录迁移失败: %v", err)
	}

	// 7. 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	log.Printf("✅ 数据库迁移完成: %s", migration.Version)
	return nil
}

// RollbackMigration 回滚迁移
func (m *MigrationManager) RollbackMigration(targetVersion string) error {
	log.Printf("🔄 开始回滚数据库到版本: %s", targetVersion)

	currentVersion, err := m.GetCurrentDBVersion()
	if err != nil {
		return err
	}

	if compareVersions(targetVersion, currentVersion) >= 0 {
		return fmt.Errorf("目标版本 %s 不小于当前版本 %s", targetVersion, currentVersion)
	}

	// 找到需要回滚的迁移（从新到旧）
	var migrationsToRollback []DatabaseMigration
	for i := len(m.migrations) - 1; i >= 0; i-- {
		migration := m.migrations[i]
		if compareVersions(migration.Version, currentVersion) <= 0 &&
			compareVersions(migration.Version, targetVersion) > 0 {
			if migration.DownSQL != "" {
				migrationsToRollback = append(migrationsToRollback, migration)
			}
		}
	}

	// 执行回滚
	for _, migration := range migrationsToRollback {
		if err := m.executeRollback(migration); err != nil {
			return fmt.Errorf("回滚迁移 %s 失败: %v", migration.Version, err)
		}
	}

	log.Printf("✅ 数据库回滚完成到版本: %s", targetVersion)
	return nil
}

// executeRollback 执行单个回滚
func (m *MigrationManager) executeRollback(migration DatabaseMigration) error {
	tx, err := m.database.DB().Begin()
	if err != nil {
		return err
	}

	// 执行回滚SQL
	if _, err := tx.Exec(migration.DownSQL); err != nil {
		tx.Rollback()
		return err
	}

	// 删除迁移记录
	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", migration.Version); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// createDatabaseBackup 创建数据库备份
func (m *MigrationManager) createDatabaseBackup(version string) error {
	backupPath := filepath.Join("backup", fmt.Sprintf("database-backup-v%s-%s.db", version, time.Now().Format("20060102-150405")))

	// 确保备份目录存在
	if err := os.MkdirAll("backup", 0755); err != nil {
		return err
	}

	// 执行数据库备份
	return m.database.Backup(backupPath)
}

// isMigrationApplied 检查迁移是否已应用
func (m *MigrationManager) isMigrationApplied(version string) bool {
	var count int
	err := m.database.DB().QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// recordMigration 记录已执行的迁移
func (m *MigrationManager) recordMigration(tx *sql.Tx, migration DatabaseMigration) error {
	_, err := tx.Exec(`
		INSERT INTO schema_migrations (version, name, description)
		VALUES (?, ?, ?)
	`, migration.Version, migration.Name, migration.Description)
	return err
}

// validateMigrationSQL 验证迁移SQL
func (m *MigrationManager) validateMigrationSQL(migration DatabaseMigration) error {
	// 可以添加SQL语法检查、危险操作检查等
	// 这里只做基本检查
	if migration.UpSQL == "" && migration.DownSQL == "" {
		return fmt.Errorf("迁移SQL不能为空")
	}
	return nil
}

// ensureMigrationTable 确保迁移表存在
func (m *MigrationManager) ensureMigrationTable() error {
	_, err := m.database.DB().Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(50) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			checksum VARCHAR(64)
		)
	`)
	return err
}

// GetMigrationStatus 获取迁移状态
func (m *MigrationManager) GetMigrationStatus() (map[string]interface{}, error) {
	currentVersion, err := m.GetCurrentDBVersion()
	if err != nil {
		return nil, err
	}

	pending, err := m.GetPendingMigrations()
	if err != nil {
		return nil, err
	}

	status := map[string]interface{}{
		"current_version":    currentVersion,
		"pending_migrations": len(pending),
		"total_migrations":   len(m.migrations),
		"needs_migration":    len(pending) > 0,
	}

	// 添加待迁移详情
	if len(pending) > 0 {
		var pendingDetails []map[string]interface{}
		for _, p := range pending {
			pendingDetails = append(pendingDetails, map[string]interface{}{
				"version":     p.Version,
				"name":        p.Name,
				"description": p.Description,
				"is_critical": p.IsCritical,
			})
		}
		status["pending_details"] = pendingDetails
	}

	return status, nil
}

// CreateMigrationFile 创建新的迁移文件
func (m *MigrationManager) CreateMigrationFile(version, name, description string) error {
	migration := DatabaseMigration{
		Version:     version,
		Name:        name,
		Description: description,
		UpSQL:       "-- 在此添加升级SQL",
		DownSQL:     "-- 在此添加回滚SQL",
		Author:      "User",
		CreatedAt:   time.Now(),
		IsCritical:  false,
	}

	// 生成文件名
	fileName := fmt.Sprintf("%s_%s.json", strings.ReplaceAll(version, ".", "_"), name)
	filePath := filepath.Join(m.migrationDir, fileName)

	// 写入文件
	data, err := json.MarshalIndent(migration, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}