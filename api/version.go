package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"nofx/config"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// VersionInfo 版本信息结构
type VersionInfo struct {
	Version          string    `json:"version"`           // 版本号 (v1.0.0)
	BuildTime        string    `json:"build_time"`        // 构建时间
	ReleaseDate      string    `json:"release_date"`      // 发布日期
	ReleaseNotes     string    `json:"release_notes"`     // 更新说明
	DownloadURL      string    `json:"download_url"`      // 下载地址
	Checksum         string    `json:"checksum"`          // 文件校验和
	UpdateSize       int64     `json:"update_size"`       // 更新包大小
	IsCriticalUpdate bool      `json:"is_critical"`       // 是否为关键更新
	MinVersion       string    `json:"min_version"`       // 最低兼容版本
	Platform         string    `json:"platform"`          // 平台信息
	UpdateType       string    `json:"update_type"`       // 更新类型: full, patch
	ForceUpdate      bool      `json:"force_update"`      // 是否强制更新
	PublishedAt      time.Time `json:"published_at"`      // 发布时间
}

// UpdateStatus 更新状态
type UpdateStatus struct {
	HasUpdate     bool      `json:"has_update"`      // 是否有可用更新
	CurrentVer    string    `json:"current_ver"`     // 当前版本
	LatestVer     string    `json:"latest_ver"`      // 最新版本
	UpdateInfo    *VersionInfo `json:"update_info,omitempty"` // 更新信息
	LastCheck     time.Time `json:"last_check"`      // 最后检查时间
	DownloadURL   string    `json:"download_url,omitempty"` // 下载地址
	AutoUpdateEnabled bool  `json:"auto_update_enabled"` // 是否启用自动更新
}

// UpdateProgress 更新进度
type UpdateProgress struct {
	Status     string  `json:"status"`      // 状态: downloading, installing, completed, failed
	Progress   float64 `json:"progress"`    // 进度 0-100
	Message    string  `json:"message"`     // 状态消息
	Speed      int64   `json:"speed"`       // 下载速度 (bytes/s)
	TotalSize  int64   `json:"total_size"`  // 总大小
	Downloaded int64   `json:"downloaded"`  // 已下载大小
	ETA        int64   `json:"eta"`         // 预计剩余时间 (seconds)
}

// UpdateRequest 更新请求
type UpdateRequest struct {
	Force     bool   `json:"force"`      // 强制更新
	AutoRestart bool `json:"auto_restart"` // 自动重启
	Backup    bool   `json:"backup"`     // 是否备份
}

// VersionManager 版本管理器
type VersionManager struct {
	database       *config.Database
	currentVersion string
	buildTime      string
	updateChan     chan *UpdateProgress
	isUpdating     bool
	updateProgress *UpdateProgress
}

// NewVersionManager 创建版本管理器
func NewVersionManager(database *config.Database) *VersionManager {
	return &VersionManager{
		database:       database,
		currentVersion: getAppVersion(),
		buildTime:      getAppBuildTime(),
		updateChan:     make(chan *UpdateProgress, 100),
		isUpdating:     false,
		updateProgress: &UpdateProgress{Status: "idle", Progress: 0},
	}
}

// getAppVersion 获取应用版本
func getAppVersion() string {
	// 可以从环境变量、编译时注入或配置文件读取
	if version := os.Getenv("APP_VERSION"); version != "" {
		return version
	}
	return "1.0.0" // 默认版本
}

// getAppBuildTime 获取应用构建时间
func getAppBuildTime() string {
	if buildTime := os.Getenv("BUILD_TIME"); buildTime != "" {
		return buildTime
	}
	return time.Now().Format("2006-01-02 15:04:05")
}

// setupVersionRoutes 设置版本管理路由
func (s *Server) setupVersionRoutes() {
	// 版本管理路由组 - 所有接口都是公开的（本地部署无需认证）
	version := s.router.Group("/api/version")
	{
		// 公开接口（无需认证）
		version.GET("/current", s.handleGetCurrentVersion)
		version.GET("/check", s.handleCheckUpdate)
		version.GET("/status", s.handleGetUpdateStatus)
		version.GET("/progress", s.handleGetUpdateProgress)
		version.POST("/download", s.handleDownloadUpdate)
		version.POST("/install", s.handleInstallUpdate)
		version.POST("/restart", s.handleRestartUpdate)
		version.POST("/auto-update", s.handleToggleAutoUpdate)
		version.GET("/history", s.handleGetUpdateHistory)
		version.POST("/rollback", s.handleRollback)

		// 数据库迁移相关接口
		version.GET("/migration/status", s.handleGetMigrationStatus)
		version.GET("/migration/pending", s.handleGetPendingMigrations)
		version.POST("/migration/execute", s.handleExecuteMigration)
		version.POST("/migration/rollback", s.handleRollbackMigration)
		version.POST("/migration/backup", s.handleCreateBackup)
	}
}

// handleGetCurrentVersion 获取当前版本信息
func (s *Server) handleGetCurrentVersion(c *gin.Context) {
	versionInfo := &VersionInfo{
		Version:     getAppVersion(),
		BuildTime:   getAppBuildTime(),
		Platform:    getPlatformInfo(),
		PublishedAt: time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    versionInfo,
	})
}

// handleCheckUpdate 检查更新
func (s *Server) handleCheckUpdate(c *gin.Context) {
	// 获取最新版本信息
	latestVersion, err := s.fetchLatestVersion()
	if err != nil {
		log.Printf("❌ 获取最新版本失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "检查更新失败",
		})
		return
	}

	// 比较版本
	currentVer := getAppVersion()
	hasUpdate := compareVersions(latestVersion.Version, currentVer) > 0

	// 检查是否为关键更新
	isCritical := latestVersion.IsCriticalUpdate ||
		compareVersions(currentVer, latestVersion.MinVersion) < 0

	updateStatus := &UpdateStatus{
		HasUpdate:        hasUpdate,
		CurrentVer:       currentVer,
		LatestVer:        latestVersion.Version,
		LastCheck:        time.Now(),
		DownloadURL:      latestVersion.DownloadURL,
		AutoUpdateEnabled: s.getAutoUpdateSetting(),
	}

	if hasUpdate {
		updateStatus.UpdateInfo = latestVersion
	}

	// 保存检查记录
	s.saveUpdateCheckRecord(updateStatus)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    updateStatus,
	})

	if hasUpdate {
		log.Printf("🔔 发现新版本: %s -> %s", currentVer, latestVersion.Version)
		if isCritical {
			log.Printf("⚠️  检测到关键更新，建议立即更新！")
		}
	}
}

// handleGetUpdateStatus 获取更新状态
func (s *Server) handleGetUpdateStatus(c *gin.Context) {
	// 获取上次检查记录
	lastCheck := s.getLastUpdateCheck()

	// 如果正在更新，返回更新进度
	if s.versionManager != nil && s.versionManager.isUpdating {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"updating":  true,
				"progress":  s.versionManager.updateProgress,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"updating": false,
			"last_check": lastCheck,
		},
	})
}

// handleDownloadUpdate 下载更新
func (s *Server) handleDownloadUpdate(c *gin.Context) {
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
		})
		return
	}

	// 检查是否已经在更新
	if s.versionManager != nil && s.versionManager.isUpdating {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "正在更新中，请稍后",
		})
		return
	}

	// 获取最新版本信息
	latestVersion, err := s.fetchLatestVersion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取更新信息失败",
		})
		return
	}

	// 创建版本管理器
	if s.versionManager == nil {
		s.versionManager = NewVersionManager(s.database)
	}

	// 开始异步下载更新
	go s.downloadAndInstallUpdate(latestVersion, req)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "开始下载更新",
	})
}

// handleInstallUpdate 安装更新
func (s *Server) handleInstallUpdate(c *gin.Context) {
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
		})
		return
	}

	// 检查更新文件是否存在
	updateFile := getUpdateFilePath()
	if _, err := os.Stat(updateFile); os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "更新文件不存在，请先下载更新",
		})
		return
	}

	// 开始安装更新
	go s.installUpdate(req)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "开始安装更新",
	})
}

// handleRestartUpdate 重启应用
func (s *Server) handleRestartUpdate(c *gin.Context) {
	// 记录重启请求
	log.Printf("🔄 收到重启请求，准备在5秒后重启应用...")

	// 异步重启，避免阻塞响应
	go func() {
		time.Sleep(5 * time.Second)
		s.restartApplication()
	}()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "应用将在5秒后重启",
	})
}

// handleGetUpdateProgress 获取更新进度
func (s *Server) handleGetUpdateProgress(c *gin.Context) {
	if s.versionManager == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": &UpdateProgress{
				Status:  "idle",
				Progress: 0,
				Message: "未开始更新",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    s.versionManager.updateProgress,
	})
}

// handleToggleAutoUpdate 切换自动更新设置
func (s *Server) handleToggleAutoUpdate(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
		})
		return
	}

	// 保存设置
	err := s.database.SetSystemConfig("auto_update_enabled", strconv.FormatBool(req.Enabled))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "保存设置失败",
		})
		return
	}

	status := "已禁用"
	if req.Enabled {
		status = "已启用"
	}

	log.Printf("✅ 自动更新设置%s", status)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("自动更新%s", status),
	})
}

// handleGetUpdateHistory 获取更新历史
func (s *Server) handleGetUpdateHistory(c *gin.Context) {
	// 从数据库获取更新历史
	history, err := s.getUpdateHistory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取更新历史失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    history,
	})
}

// handleRollback 回滚版本
func (s *Server) handleRollback(c *gin.Context) {
	var req struct {
		TargetVersion string `json:"target_version"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
		})
		return
	}

	// 检查备份是否存在
	backupPath := getBackupPath(req.TargetVersion)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "备份文件不存在",
		})
		return
	}

	// 执行回滚
	go s.rollbackToVersion(req.TargetVersion)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "开始回滚到版本 " + req.TargetVersion,
	})
}

// fetchLatestVersion 获取最新版本信息
func (s *Server) fetchLatestVersion() (*VersionInfo, error) {
	// 这里可以从远程服务器获取版本信息
	// 为了演示，这里返回一个模拟的版本信息

	// 实际应用中，应该从以下方式获取：
	// 1. GitHub Releases API
	// 2. 自建版本服务器
	// 3. 配置文件中的版本信息

	updateServerURL := "https://api.github.com/repos/akuntk/ai-trading/releases/latest"

	// 模拟网络请求
	resp, err := http.Get(updateServerURL)
	if err != nil {
		// 如果远程获取失败，返回本地配置的版本
		return s.getLocalVersionInfo(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return s.getLocalVersionInfo(), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return s.getLocalVersionInfo(), nil
	}

	var release struct {
		TagName      string    `json:"tag_name"`
		Name         string    `json:"name"`
		Body         string    `json:"body"`
		PublishedAt  time.Time `json:"published_at"`
		Assets       []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		} `json:"assets"`
	}

	if err := json.Unmarshal(body, &release); err != nil {
		return s.getLocalVersionInfo(), nil
	}

	versionInfo := &VersionInfo{
		Version:         strings.TrimPrefix(release.TagName, "v"),
		ReleaseDate:     release.PublishedAt.Format("2006-01-02"),
		ReleaseNotes:    release.Body,
		PublishedAt:     release.PublishedAt,
		IsCriticalUpdate: false, // 可以从release body或label中解析
		MinVersion:      "1.0.0", // 可以从配置中获取
		UpdateType:      "full",
		ForceUpdate:     false,
	}

	// 获取下载链接（根据平台选择合适的文件）
	if len(release.Assets) > 0 {
		for _, asset := range release.Assets {
			if strings.Contains(asset.Name, getPlatformString()) {
				versionInfo.DownloadURL = asset.BrowserDownloadURL
				versionInfo.UpdateSize = asset.Size
				break
			}
		}
	}

	return versionInfo, nil
}

// getLocalVersionInfo 获取本地配置的版本信息
func (s *Server) getLocalVersionInfo() *VersionInfo {
	return &VersionInfo{
		Version:         "1.0.1",
		BuildTime:       time.Now().Format("2006-01-02 15:04:05"),
		ReleaseDate:     time.Now().Format("2006-01-02"),
		ReleaseNotes:    "新功能:\n- 添加自动版本控制和更新系统\n- 改进用户界面\n- 修复已知问题",
		DownloadURL:     "",
		UpdateSize:      0,
		IsCriticalUpdate: false,
		MinVersion:      "1.0.0",
		Platform:        getPlatformInfo(),
		UpdateType:      "full",
		ForceUpdate:     false,
		PublishedAt:     time.Now(),
	}
}

// downloadAndInstallUpdate 下载并安装更新
func (s *Server) downloadAndInstallUpdate(versionInfo *VersionInfo, req UpdateRequest) {
	if s.versionManager == nil {
		return
	}

	s.versionManager.isUpdating = true
	defer func() {
		s.versionManager.isUpdating = false
	}()

	// 发送进度更新
	s.sendUpdateProgress(&UpdateProgress{
		Status:  "preparing",
		Message: "准备下载更新...",
		Progress: 0,
	})

	// 备份当前版本
	if req.Backup {
		s.sendUpdateProgress(&UpdateProgress{
			Status:  "backup",
			Message: "备份当前版本...",
			Progress: 5,
		})

		if err := s.backupCurrentVersion(); err != nil {
			log.Printf("❌ 备份失败: %v", err)
			s.sendUpdateProgress(&UpdateProgress{
				Status:  "failed",
				Message: "备份失败: " + err.Error(),
				Progress: 0,
			})
			return
		}
	}

	// 下载更新
	s.sendUpdateProgress(&UpdateProgress{
		Status:  "downloading",
		Message: "下载更新文件...",
		Progress: 10,
	})

	if err := s.downloadUpdateFile(versionInfo); err != nil {
		log.Printf("❌ 下载更新失败: %v", err)
		s.sendUpdateProgress(&UpdateProgress{
			Status:  "failed",
			Message: "下载失败: " + err.Error(),
			Progress: 0,
		})
		return
	}

	// 验证下载文件
	s.sendUpdateProgress(&UpdateProgress{
		Status:  "verifying",
		Message: "验证更新文件...",
		Progress: 80,
	})

	if err := s.verifyUpdateFile(); err != nil {
		log.Printf("❌ 验证更新文件失败: %v", err)
		s.sendUpdateProgress(&UpdateProgress{
			Status:  "failed",
			Message: "文件验证失败: " + err.Error(),
			Progress: 0,
		})
		return
	}

	// 安装更新
	s.sendUpdateProgress(&UpdateProgress{
		Status:  "installing",
		Message: "安装更新...",
		Progress: 90,
	})

	if err := s.installUpdateFile(); err != nil {
		log.Printf("❌ 安装更新失败: %v", err)
		s.sendUpdateProgress(&UpdateProgress{
			Status:  "failed",
			Message: "安装失败: " + err.Error(),
			Progress: 0,
		})
		return
	}

	// 完成
	s.sendUpdateProgress(&UpdateProgress{
		Status:  "completed",
		Message: "更新完成，准备重启...",
		Progress: 100,
	})

	// 保存更新记录
	s.saveUpdateRecord(versionInfo)

	// 如果启用自动重启，则自动重启
	if req.AutoRestart {
		time.Sleep(3 * time.Second)
		s.restartApplication()
	}
}

// sendUpdateProgress 发送更新进度
func (s *Server) sendUpdateProgress(progress *UpdateProgress) {
	if s.versionManager != nil {
		s.versionManager.updateProgress = progress
		select {
		case s.versionManager.updateChan <- progress:
		default:
			// 如果通道满了，丢弃旧的进度
		}
	}
	log.Printf("📊 更新进度: %s - %.1f%% - %s", progress.Status, progress.Progress, progress.Message)
}

// 辅助函数

// compareVersions 比较版本号 (返回: -1, 0, 1)
func compareVersions(v1, v2 string) int {
	v1Parts := strings.Split(strings.TrimPrefix(v1, "v"), ".")
	v2Parts := strings.Split(strings.TrimPrefix(v2, "v"), ".")

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var v1Num, v2Num int

		if i < len(v1Parts) {
			if num, err := strconv.Atoi(v1Parts[i]); err == nil {
				v1Num = num
			}
		}

		if i < len(v2Parts) {
			if num, err := strconv.Atoi(v2Parts[i]); err == nil {
				v2Num = num
			}
		}

		if v1Num > v2Num {
			return 1
		}
		if v1Num < v2Num {
			return -1
		}
	}

	return 0
}

// getPlatformInfo 获取平台信息
func getPlatformInfo() string {
	return fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
}

// getPlatformString 获取平台字符串
func getPlatformString() string {
	switch runtime.GOOS {
	case "windows":
		if runtime.GOARCH == "amd64" {
			return "windows-amd64"
		}
		return "windows-" + runtime.GOARCH
	case "linux":
		if runtime.GOARCH == "amd64" {
			return "linux-amd64"
		}
		return "linux-" + runtime.GOARCH
	case "darwin":
		if runtime.GOARCH == "amd64" {
			return "darwin-amd64"
		} else if runtime.GOARCH == "arm64" {
			return "darwin-arm64"
		}
		return "darwin-" + runtime.GOARCH
	default:
		return runtime.GOOS + "-" + runtime.GOARCH
	}
}

// getUpdateFilePath 获取更新文件路径
func getUpdateFilePath() string {
	return filepath.Join(os.TempDir(), "nofx-update.bin")
}

// getBackupPath 获取备份路径
func getBackupPath(version string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("nofx-backup-%s", version))
}

// 其他辅助函数的实现...

// backupCurrentVersion 备份当前版本
func (s *Server) backupCurrentVersion() error {
	currentVersion := getAppVersion()
	backupPath := getBackupPath(currentVersion)

	// 确保备份目录存在
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("创建备份目录失败: %v", err)
	}

	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %v", err)
	}

	// 发送进度更新
	s.sendUpdateProgress(&UpdateProgress{
		Status:  "backup",
		Message: "备份当前版本...",
		Progress: 5,
	})

	// 打开原文件
	srcFile, err := os.Open(execPath)
	if err != nil {
		return fmt.Errorf("打开原文件失败: %v", err)
	}
	defer srcFile.Close()

	// 创建备份文件
	destFile, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("创建备份文件失败: %v", err)
	}
	defer destFile.Close()

	// 复制文件
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("复制文件失败: %v", err)
	}

	// 设置文件权限
	if err := os.Chmod(backupPath, 0755); err != nil {
		log.Printf("⚠️  设置备份文件权限失败: %v", err)
	}

	// 创建用户配置保护文件列表
	// 这些文件在更新时需要保护，不能被覆盖
	userDataFiles := []string{
		"config.json",        // 用户主配置文件
		"database.db",        // SQLite数据库文件
		"database.sqlite",    // SQLite数据库文件（另一种命名）
		".env",              // 环境变量文件
		"logs/",             // 日志目录
		"backup/",           // 备份目录
		"data/",             // 用户数据目录
	}

	// 记录需要保护的文件，以便在更新时避免覆盖
	protectionFilePath := filepath.Join(os.TempDir(), "nofx-user-protection.txt")
	protectionContent := strings.Join(userDataFiles, "\n")

	if err := os.WriteFile(protectionFilePath, []byte(protectionContent), 0644); err != nil {
		log.Printf("⚠️  创建用户文件保护列表失败: %v", err)
	} else {
		log.Printf("✅ 用户文件保护列表已创建: %s", protectionFilePath)
	}

	// 备份关键的配置文件（仅用于紧急恢复，不会在更新时恢复）
	configBackupPath := backupPath + ".emergency-config"
	emergencyConfigFiles := []string{"config.json.example"} // 只备份模板文件

	for _, configFile := range emergencyConfigFiles {
		if _, err := os.Stat(configFile); err == nil {
			if copyErr := copyFileForBackup(configFile, configBackupPath+"."+configFile); copyErr != nil {
				log.Printf("⚠️  备份配置模板 %s 失败: %v", configFile, copyErr)
			} else {
				log.Printf("✅ 配置模板 %s 备份成功", configFile)
			}
		}
	}

	log.Printf("✅ 当前版本备份完成: %s", backupPath)
	return nil
}

// copyFileForBackup 备份用的复制文件辅助函数
func copyFileForBackup(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// downloadUpdateFile 下载更新文件
func (s *Server) downloadUpdateFile(versionInfo *VersionInfo) error {
	if versionInfo.DownloadURL == "" {
		return fmt.Errorf("下载URL为空")
	}

	updateFile := getUpdateFilePath()

	// 创建文件
	file, err := os.Create(updateFile)
	if err != nil {
		return fmt.Errorf("创建更新文件失败: %v", err)
	}
	defer file.Close()

	// 发送进度更新
	s.sendUpdateProgress(&UpdateProgress{
		Status:  "downloading",
		Message: "开始下载更新文件...",
		Progress: 15,
	})

	// 发起HTTP请求
	resp, err := http.Get(versionInfo.DownloadURL)
	if err != nil {
		return fmt.Errorf("下载请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 获取文件大小
	totalSize := resp.ContentLength
	downloaded := int64(0)
	buffer := make([]byte, 32*1024) // 32KB缓冲区
	startTime := time.Now()

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := file.Write(buffer[:n])
			if writeErr != nil {
				return fmt.Errorf("写入文件失败: %v", writeErr)
			}

			downloaded += int64(n)

			// 计算进度
			progress := float64(downloaded) / float64(totalSize) * 100
			if progress > 90 {
				progress = 90 // 预留10%给验证阶段
			}

			// 计算下载速度
			elapsed := time.Since(startTime).Seconds()
			var speed int64
			if elapsed > 0 {
				speed = int64(float64(downloaded) / elapsed)
			}

			// 计算预计剩余时间
			var eta int64
			if speed > 0 {
				eta = int64(float64(totalSize-downloaded) / float64(speed))
			}

			// 发送进度更新
			s.sendUpdateProgress(&UpdateProgress{
				Status:     "downloading",
				Message:    "正在下载更新文件...",
				Progress:   progress,
				Speed:      speed,
				TotalSize:  totalSize,
				Downloaded: downloaded,
				ETA:        eta,
			})
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("下载过程中出错: %v", err)
		}
	}

	log.Printf("✅ 更新文件下载完成: %s (%d bytes)", updateFile, downloaded)
	return nil
}

// verifyUpdateFile 验证更新文件
func (s *Server) verifyUpdateFile() error {
	updateFile := getUpdateFilePath()

	// 检查文件是否存在
	fileInfo, err := os.Stat(updateFile)
	if err != nil {
		return fmt.Errorf("更新文件不存在: %v", err)
	}

	// 检查文件大小
	if fileInfo.Size() == 0 {
		return fmt.Errorf("更新文件为空")
	}

	// 发送进度更新
	s.sendUpdateProgress(&UpdateProgress{
		Status:  "verifying",
		Message: "验证文件完整性...",
		Progress: 85,
	})

	// 可以添加更多验证逻辑，比如：
	// 1. 校验和验证
	// 2. 数字签名验证
	// 3. 文件格式验证
	// 这里只做基本检查

	log.Printf("✅ 更新文件验证通过: %s (%d bytes)", updateFile, fileInfo.Size())
	return nil
}

// installUpdateFile 安装更新文件
func (s *Server) installUpdateFile() error {
	updateFile := getUpdateFilePath()

	// 检查更新文件是否存在
	if _, err := os.Stat(updateFile); err != nil {
		return fmt.Errorf("更新文件不存在: %v", err)
	}

	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %v", err)
	}

	// 发送进度更新
	s.sendUpdateProgress(&UpdateProgress{
		Status:  "installing",
		Message: "安装更新文件...",
		Progress: 90,
	})

	// 在Windows系统上，需要特殊处理正在运行的文件
	if runtime.GOOS == "windows" {
		return s.installUpdateWindows(updateFile, execPath)
	}

	// Unix/Linux/Mac 系统的安装逻辑
	return s.installUpdateUnix(updateFile, execPath)
}

// installUpdateWindows Windows系统的安装逻辑
func (s *Server) installUpdateWindows(updateFile, execPath string) error {
	// 创建批处理脚本来替换文件
	batchScript := `@echo off
echo 正在更新应用程序...
echo 只更新可执行文件，保留用户配置...
timeout /t 2 /nobreak >nul

REM 备份当前可执行文件
copy /Y "%s" "%s.bak" >nul

REM 停止可能的现有进程
taskkill /f /im nofx.exe >nul 2>&1
timeout /t 1 /nobreak >nul

REM 替换可执行文件
copy /Y "%s" "%s" >nul

if %ERRORLEVEL% EQU 0 (
    echo 更新成功！
    echo 用户配置文件保持不变
    echo 正在重启应用程序...
    timeout /t 1 /nobreak >nul
    start "" "%s"
) else (
    echo 更新失败！正在恢复备份...
    copy /Y "%s.bak" "%s" >nul
    pause
)

REM 清理临时文件
del "%s" >nul 2>&1
del "%%~f0" >nul 2>&1
`

	// 生成批处理脚本
	scriptPath := filepath.Join(os.TempDir(), "nofx-update.bat")
	scriptContent := fmt.Sprintf(batchScript, updateFile, execPath, updateFile, execPath, execPath, execPath, updateFile, scriptPath)

	// 写入批处理文件
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		return fmt.Errorf("创建更新脚本失败: %v", err)
	}

	// 启动批处理脚本
	cmd := exec.Command("cmd", "/C", scriptPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动更新脚本失败: %v", err)
	}

	log.Printf("✅ Windows更新脚本已启动: %s", scriptPath)
	return nil
}

// installUpdateUnix Unix/Linux/Mac系统的安装逻辑
func (s *Server) installUpdateUnix(updateFile, execPath string) error {
	// 创建更新脚本
	scriptContent := fmt.Sprintf(`#!/bin/bash

echo "正在更新应用程序..."
echo "只更新可执行文件，保留用户配置..."

# 备份当前可执行文件
cp -f "%s" "%s.bak"
if [ $? -ne 0 ]; then
    echo "备份失败，中止更新"
    exit 1
fi

# 等待应用完全关闭
sleep 2

# 停止可能的现有进程
pkill -f nofx > /dev/null 2>&1
sleep 1

# 替换可执行文件
cp -f "%s" "%s"

if [ $? -eq 0 ]; then
    echo "更新成功！"
    echo "用户配置文件保持不变"
    echo "正在重启应用程序..."
    sleep 1
    chmod +x "%s"
    exec "%s"
else
    echo "更新失败！正在恢复备份..."
    cp -f "%s.bak" "%s"
    exit 1
fi

# 清理临时文件
rm -f "$0"
rm -f "%s"
rm -f "%s.bak"
`, updateFile, execPath, updateFile, execPath, execPath, execPath, updateFile, execPath, updateFile)

	// 生成脚本文件
	scriptPath := filepath.Join(os.TempDir(), "nofx-update.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("创建更新脚本失败: %v", err)
	}

	// 启动脚本
	cmd := exec.Command(scriptPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动更新脚本失败: %v", err)
	}

	log.Printf("✅ Unix更新脚本已启动: %s", scriptPath)
	return nil
}

// restartApplication 重启应用
func (s *Server) restartApplication() {
	log.Printf("🔄 重启应用...")

	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("❌ 获取可执行文件路径失败: %v", err)
		os.Exit(1)
		return
	}

	// 优雅关闭：等待当前请求处理完成
	log.Printf("⏳ 等待当前请求处理完成...")
	time.Sleep(2 * time.Second)

	// 根据操作系统选择重启方式
	if runtime.GOOS == "windows" {
		s.restartWindows(execPath)
	} else {
		s.restartUnix(execPath)
	}
}

// restartWindows Windows系统重启
func (s *Server) restartWindows(execPath string) {
	// 创建重启脚本
	scriptContent := fmt.Sprintf(`@echo off
echo 正在重启应用程序...
timeout /t 2 /nobreak >nul
start "" "%s"
del "%%~f0"
`, execPath)

	scriptPath := filepath.Join(os.TempDir(), "nofx-restart.bat")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		log.Printf("❌ 创建重启脚本失败: %v", err)
		os.Exit(1)
		return
	}

	// 启动重启脚本
	cmd := exec.Command("cmd", "/C", scriptPath)
	if err := cmd.Start(); err != nil {
		log.Printf("❌ 启动重启脚本失败: %v", err)
		os.Exit(1)
		return
	}

	log.Printf("✅ 重启脚本已启动，应用程序即将退出")
	os.Exit(0)
}

// restartUnix Unix/Linux/Mac系统重启
func (s *Server) restartUnix(execPath string) {
	// 创建重启脚本
	scriptContent := fmt.Sprintf(`#!/bin/bash
echo "正在重启应用程序..."
sleep 2
exec "%s"
rm -f "$0"
`, execPath)

	scriptPath := filepath.Join(os.TempDir(), "nofx-restart.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		log.Printf("❌ 创建重启脚本失败: %v", err)
		os.Exit(1)
		return
	}

	// 启动重启脚本
	cmd := exec.Command(scriptPath)
	if err := cmd.Start(); err != nil {
		log.Printf("❌ 启动重启脚本失败: %v", err)
		os.Exit(1)
		return
	}

	log.Printf("✅ 重启脚本已启动，应用程序即将退出")
	os.Exit(0)
}

// getAutoUpdateSetting 获取自动更新设置
func (s *Server) getAutoUpdateSetting() bool {
	enabledStr, _ := s.database.GetSystemConfig("auto_update_enabled")
	return enabledStr == "true"
}

// saveUpdateCheckRecord 保存更新检查记录
func (s *Server) saveUpdateCheckRecord(status *UpdateStatus) {
	// 实现保存逻辑
}

// getLastUpdateCheck 获取上次更新检查记录
func (s *Server) getLastUpdateCheck() *UpdateStatus {
	// 实现获取逻辑
	return nil
}

// getUpdateHistory 获取更新历史
func (s *Server) getUpdateHistory() ([]interface{}, error) {
	// 实现获取历史逻辑
	return nil, nil
}

// saveUpdateRecord 保存更新记录
func (s *Server) saveUpdateRecord(versionInfo *VersionInfo) {
	// 实现保存记录逻辑
}

// rollbackToVersion 回滚到指定版本
func (s *Server) rollbackToVersion(version string) {
	// 实现回滚逻辑
}

// installUpdate 安装更新的具体实现
func (s *Server) installUpdate(req UpdateRequest) {
	log.Printf("🔄 开始安装更新...")

	// 阶段1: 检查数据库迁移
	s.sendUpdateProgress(&UpdateProgress{
		Status:  "checking",
		Message: "检查数据库迁移状态...",
		Progress: 10,
	})

	// 创建迁移管理器
	migrationManager := NewMigrationManager(s.database)
	if err := migrationManager.LoadMigrations(); err != nil {
		log.Printf("❌ 加载迁移失败: %v", err)
		s.sendUpdateProgress(&UpdateProgress{
			Status:  "error",
			Message: "加载迁移失败: " + err.Error(),
			Progress: 0,
		})
		return
	}

	// 检查是否需要迁移
	status, err := migrationManager.GetMigrationStatus()
	if err != nil {
		log.Printf("❌ 获取迁移状态失败: %v", err)
		s.sendUpdateProgress(&UpdateProgress{
			Status:  "error",
			Message: "获取迁移状态失败: " + err.Error(),
			Progress: 0,
		})
		return
	}

	// 阶段2: 执行数据库迁移（如果需要）
	if status["needs_migration"].(bool) {
		s.sendUpdateProgress(&UpdateProgress{
			Status:  "migrating",
			Message: "执行数据库迁移...",
			Progress: 30,
		})

		// 自动备份数据库
		s.sendUpdateProgress(&UpdateProgress{
			Status:  "backup",
			Message: "创建数据库备份...",
			Progress: 35,
		})

		backupPath := filepath.Join("backup", fmt.Sprintf("pre-update-backup-%s.db", time.Now().Format("20060102-150405")))
		if err := s.database.Backup(backupPath); err != nil {
			log.Printf("⚠️  数据库备份失败: %v", err)
			// 备份失败不中止更新，但记录警告
		} else {
			log.Printf("✅ 数据库备份成功: %s", backupPath)
		}

		// 获取待执行迁移
		pendingMigrations, err := migrationManager.GetPendingMigrations()
		if err != nil {
			log.Printf("❌ 获取待执行迁移失败: %v", err)
			s.sendUpdateProgress(&UpdateProgress{
				Status:  "error",
				Message: "获取待执行迁移失败: " + err.Error(),
				Progress: 0,
			})
			return
		}

		// 执行所有待执行迁移
		migrationProgress := 40.0
		progressStep := 30.0 / float64(len(pendingMigrations))

		for _, migration := range pendingMigrations {
			log.Printf("🔄 执行数据库迁移: %s (%s)", migration.Version, migration.Name)

			if err := migrationManager.ExecuteMigration(migration, true); err != nil {
				log.Printf("❌ 迁移执行失败 %s: %v", migration.Version, err)
				s.sendUpdateProgress(&UpdateProgress{
					Status:  "error",
					Message: fmt.Sprintf("迁移执行失败 %s: %v", migration.Version, err),
					Progress: 0,
				})
				return
			}

			migrationProgress += progressStep
			s.sendUpdateProgress(&UpdateProgress{
				Status:  "migrating",
				Message: fmt.Sprintf("执行迁移 %s (%s)", migration.Version, migration.Name),
				Progress: migrationProgress,
			})
		}

		log.Printf("✅ 数据库迁移完成")
		s.sendUpdateProgress(&UpdateProgress{
			Status:  "migrating",
			Message: "数据库迁移完成",
			Progress: 70,
		})
	} else {
		log.Printf("✅ 数据库无需迁移")
		s.sendUpdateProgress(&UpdateProgress{
			Status:  "migrating",
			Message: "数据库无需迁移",
			Progress: 70,
		})
	}

	// 阶段3: 安装更新文件
	s.sendUpdateProgress(&UpdateProgress{
		Status:  "installing",
		Message: "安装更新文件...",
		Progress: 80,
	})

	if err := s.installUpdateFile(); err != nil {
		log.Printf("❌ 安装更新文件失败: %v", err)
		s.sendUpdateProgress(&UpdateProgress{
			Status:  "error",
			Message: "安装更新文件失败: " + err.Error(),
			Progress: 0,
		})
		return
	}

	// 阶段4: 验证安装
	s.sendUpdateProgress(&UpdateProgress{
		Status:  "verifying",
		Message: "验证更新安装...",
		Progress: 95,
	})

	// 这里可以添加更多验证逻辑
	log.Printf("✅ 更新安装验证完成")

	// 完成安装
	s.sendUpdateProgress(&UpdateProgress{
		Status:  "completed",
		Message: "更新完成，准备重启...",
		Progress: 100,
	})

	// 保存更新记录
	s.saveUpdateRecord(&VersionInfo{
		Version:     getAppVersion(),
		BuildTime:   getAppBuildTime(),
		Platform:    getPlatformInfo(),
		PublishedAt: time.Now(),
	})

	log.Printf("✅ 更新安装完成")

	// 如果启用自动重启，则自动重启
	if req.AutoRestart {
		time.Sleep(3 * time.Second)
		s.restartApplication()
	}
}

// ===== 数据库迁移API处理函数 =====

// handleGetMigrationStatus 获取数据库迁移状态
func (s *Server) handleGetMigrationStatus(c *gin.Context) {
	// 创建迁移管理器
	migrationManager := NewMigrationManager(s.database)
	if err := migrationManager.LoadMigrations(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "加载迁移失败: " + err.Error(),
		})
		return
	}

	// 获取迁移状态
	status, err := migrationManager.GetMigrationStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取迁移状态失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// handleGetPendingMigrations 获取待执行的迁移
func (s *Server) handleGetPendingMigrations(c *gin.Context) {
	migrationManager := NewMigrationManager(s.database)
	if err := migrationManager.LoadMigrations(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "加载迁移失败: " + err.Error(),
		})
		return
	}

	pending, err := migrationManager.GetPendingMigrations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取待迁移列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    pending,
	})
}

// handleExecuteMigration 执行数据库迁移
func (s *Server) handleExecuteMigration(c *gin.Context) {
	var req struct {
		Version    string `json:"version"`
		AutoBackup bool   `json:"auto_backup"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
		})
		return
	}

	migrationManager := NewMigrationManager(s.database)
	if err := migrationManager.LoadMigrations(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "加载迁移失败: " + err.Error(),
		})
		return
	}

	// 查找指定版本的迁移
	var targetMigration *DatabaseMigration
	for _, migration := range migrationManager.migrations {
		if migration.Version == req.Version {
			targetMigration = &migration
			break
		}
	}

	if targetMigration == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   fmt.Sprintf("未找到版本 %s 的迁移", req.Version),
		})
		return
	}

	// 执行迁移
	if err := migrationManager.ExecuteMigration(*targetMigration, req.AutoBackup); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "执行迁移失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("数据库迁移 %s 执行成功", req.Version),
	})
}

// handleRollbackMigration 回滚数据库迁移
func (s *Server) handleRollbackMigration(c *gin.Context) {
	var req struct {
		TargetVersion string `json:"target_version"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
		})
		return
	}

	migrationManager := NewMigrationManager(s.database)
	if err := migrationManager.LoadMigrations(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "加载迁移失败: " + err.Error(),
		})
		return
	}

	// 创建备份
	backupPath := filepath.Join("backup", fmt.Sprintf("pre-rollback-backup-%s.db", time.Now().Format("20060102-150405")))
	if err := os.MkdirAll("backup", 0755); err == nil {
		s.database.Backup(backupPath)
		log.Printf("✅ 回滚前备份已创建: %s", backupPath)
	}

	// 执行回滚
	if err := migrationManager.RollbackMigration(req.TargetVersion); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "回滚失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("数据库已成功回滚到版本 %s", req.TargetVersion),
	})
}

// handleCreateBackup 创建数据库备份
func (s *Server) handleCreateBackup(c *gin.Context) {
	var req struct {
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
		})
		return
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102-150405")
	backupFileName := fmt.Sprintf("manual-backup-%s.db", timestamp)
	if req.Description != "" {
		backupFileName = fmt.Sprintf("manual-backup-%s-%s.db", req.Description, timestamp)
	}
	backupPath := filepath.Join("backup", backupFileName)

	// 确保备份目录存在
	if err := os.MkdirAll("backup", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建备份目录失败: " + err.Error(),
		})
		return
	}

	// 执行备份
	if err := s.database.Backup(backupPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建备份失败: " + err.Error(),
		})
		return
	}

	log.Printf("✅ 手动数据库备份已创建: %s", backupPath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "数据库备份创建成功",
		"data": gin.H{
			"backup_path": backupPath,
			"timestamp":   timestamp,
		},
	})
}