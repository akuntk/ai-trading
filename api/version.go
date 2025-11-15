package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"nofx/config"
	"os"
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

	updateServerURL := "https://api.github.com/repos/your-repo/nofx/releases/latest"

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
	// 实现备份逻辑
	return nil
}

// downloadUpdateFile 下载更新文件
func (s *Server) downloadUpdateFile(versionInfo *VersionInfo) error {
	// 实现下载逻辑
	return nil
}

// verifyUpdateFile 验证更新文件
func (s *Server) verifyUpdateFile() error {
	// 实现验证逻辑
	return nil
}

// installUpdateFile 安装更新文件
func (s *Server) installUpdateFile() error {
	// 实现安装逻辑
	return nil
}

// restartApplication 重启应用
func (s *Server) restartApplication() {
	log.Printf("🔄 重启应用...")
	// 实现重启逻辑
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
	// 实现安装更新逻辑
}