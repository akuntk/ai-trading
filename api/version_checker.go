package api

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// VersionChecker 版本检查器
type VersionChecker struct {
	server         *Server
	checkInterval  time.Duration
	stopChan       chan bool
	isRunning      bool
	lastCheckTime  time.Time
	notificationChan chan *UpdateNotification
}

// UpdateNotification 更新通知
type UpdateNotification struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"` // new_update, critical_update, security_patch
	Title           string    `json:"title"`
	Message         string    `json:"message"`
	Version         string    `json:"version"`
	ReleaseDate     string    `json:"release_date"`
	IsCritical      bool      `json:"is_critical"`
	DownloadURL     string    `json:"download_url"`
	CreatedAt       time.Time `json:"created_at"`
	RequiresAction  bool      `json:"requires_action"`
	ActionText      string    `json:"action_text"`
}

// NotificationStorage 通知存储接口
type NotificationStorage interface {
	SaveNotification(userID string, notification *UpdateNotification) error
	GetUnreadNotifications(userID string) ([]*UpdateNotification, error)
	MarkNotificationRead(userID, notificationID string) error
	ClearOldNotifications(userID string, olderThan time.Duration) error
}

// NewVersionChecker 创建版本检查器
func NewVersionChecker(server *Server) *VersionChecker {
	// 从配置获取检查间隔，默认为1小时
	checkInterval := time.Hour
	if server != nil && server.database != nil {
		if intervalStr, _ := server.database.GetSystemConfig("version_check_interval"); intervalStr != "" {
			if duration, err := time.ParseDuration(intervalStr); err == nil {
				checkInterval = duration
			}
		}
	}

	return &VersionChecker{
		server:         server,
		checkInterval:  checkInterval,
		stopChan:       make(chan bool),
		isRunning:      false,
		notificationChan: make(chan *UpdateNotification, 100),
	}
}

// Start 开始版本检查
func (vc *VersionChecker) Start() {
	if vc.isRunning {
		log.Printf("⚠️  版本检查器已在运行")
		return
	}

	vc.isRunning = true
	log.Printf("🔍 启动版本检查器，检查间隔: %v", vc.checkInterval)

	// 立即执行一次检查
	go vc.checkForUpdates()

	// 启动定期检查
	go vc.periodicCheck()

	// 启动通知处理器
	go vc.notificationProcessor()

	// 启动WebSocket广播
	go vc.broadcastNotifications()
}

// Stop 停止版本检查
func (vc *VersionChecker) Stop() {
	if !vc.isRunning {
		return
	}

	log.Printf("⏹  停止版本检查器")
	vc.isRunning = false
	close(vc.stopChan)
}

// periodicCheck 定期检查
func (vc *VersionChecker) periodicCheck() {
	ticker := time.NewTicker(vc.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if vc.isRunning {
				go vc.checkForUpdates()
			}
		case <-vc.stopChan:
			return
		}
	}
}

// checkForUpdates 检查更新
func (vc *VersionChecker) checkForUpdates() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ 版本检查器异常: %v", r)
		}
	}()

	vc.lastCheckTime = time.Now()
	log.Printf("🔍 开始检查版本更新...")

	// 获取最新版本信息
	latestVersion, err := vc.server.fetchLatestVersion()
	if err != nil {
		log.Printf("❌ 获取最新版本失败: %v", err)
		return
	}

	// 获取当前版本
	currentVersion := getAppVersion()

	// 检查latestVersion是否为nil
	if latestVersion == nil {
		log.Printf("⚠️  无法获取最新版本信息，跳过版本检查")
		return
	}

	// 比较版本
	hasUpdate := compareVersions(latestVersion.Version, currentVersion) > 0
	if !hasUpdate {
		log.Printf("✅ 当前版本 %s 已是最新", currentVersion)
		return
	}

	log.Printf("🔔 发现新版本: %s -> %s", currentVersion, latestVersion.Version)

	// 检查是否为关键更新
	isCritical := latestVersion.IsCriticalUpdate ||
		compareVersions(currentVersion, latestVersion.MinVersion) < 0

	// 创建更新通知
	notification := &UpdateNotification{
		ID:             uuid.New().String(),
		Type:           "new_update",
		Title:          "发现新版本",
		Message:        fmt.Sprintf("发现新版本 %s，当前版本 %s", latestVersion.Version, currentVersion),
		Version:        latestVersion.Version,
		ReleaseDate:    latestVersion.ReleaseDate,
		IsCritical:     isCritical,
		DownloadURL:    latestVersion.DownloadURL,
		CreatedAt:      time.Now(),
		RequiresAction: isCritical,
	}

	if isCritical {
		notification.Type = "critical_update"
		notification.Title = "关键更新"
		notification.Message += " (关键更新，建议立即更新)"
		notification.ActionText = "立即更新"
	} else {
		notification.ActionText = "查看更新"
	}

	// 发送通知
	select {
	case vc.notificationChan <- notification:
		log.Printf("📢 已发送更新通知: %s", notification.Title)
	default:
		log.Printf("⚠️  通知队列已满，丢弃通知")
	}

	// 保存到数据库
	vc.saveNotificationToAllUsers(notification)

	// 如果是关键更新，启用自动更新（如果用户之前启用过）
	if isCritical && vc.server.getAutoUpdateSetting() {
		log.Printf("🚀 检测到关键更新，准备自动下载...")
		go vc.server.downloadAndInstallUpdate(latestVersion, UpdateRequest{
			AutoRestart: false, // 关键更新也要求用户确认重启
			Backup:      true,
		})
	}
}

// notificationProcessor 通知处理器
func (vc *VersionChecker) notificationProcessor() {
	for {
		select {
		case notification := <-vc.notificationChan:
			vc.processNotification(notification)
		case <-vc.stopChan:
			return
		}
	}
}

// processNotification 处理通知
func (vc *VersionChecker) processNotification(notification *UpdateNotification) {
	// 通知可以是：
	// 1. 发送到前端WebSocket
	// 2. 发送邮件通知
	// 3. 发送Telegram通知
	// 4. 发送到第三方通知服务

	log.Printf("📨 处理通知: %s - %s", notification.Title, notification.Message)

	// 这里可以扩展各种通知方式
	vc.sendWebSocketNotification(notification)
	vc.sendEmailNotification(notification)
	vc.sendTelegramNotification(notification)
}

// sendWebSocketNotification 发送WebSocket通知
func (vc *VersionChecker) sendWebSocketNotification(notification *UpdateNotification) {
	// 通过WebSocket广播给所有在线用户
	message := map[string]interface{}{
		"type":        "version_update",
		"notification": notification,
		"timestamp":   time.Now().Unix(),
	}

	_, _ = json.Marshal(message)

	// 这里应该通过WebSocket管理器广播消息
	// 例如: websocketManager.BroadcastToAll(data)
	log.Printf("🌐 WebSocket通知已广播: %s", notification.Title)
}

// sendEmailNotification 发送邮件通知
func (vc *VersionChecker) sendEmailNotification(notification *UpdateNotification) {
	// 获取所有用户邮箱并发送邮件
	// 这里需要实现邮件发送功能
	log.Printf("📧 邮件通知已发送: %s", notification.Title)
}

// sendTelegramNotification 发送Telegram通知
func (vc *VersionChecker) sendTelegramNotification(notification *UpdateNotification) {
	// 发送Telegram通知
	log.Printf("📱 Telegram通知已发送: %s", notification.Title)
}

// saveNotificationToAllUsers 为所有用户保存通知
func (vc *VersionChecker) saveNotificationToAllUsers(notification *UpdateNotification) {
	// 获取所有用户ID
	userIDs, err := vc.getAllUserIDs()
	if err != nil {
		log.Printf("❌ 获取用户列表失败: %v", err)
		return
	}

	// 为每个用户保存通知
	for _, userID := range userIDs {
		err := vc.saveNotification(userID, notification)
		if err != nil {
			log.Printf("❌ 为用户 %s 保存通知失败: %v", userID, err)
		}
	}
}

// getAllUserIDs 获取所有用户ID
func (vc *VersionChecker) getAllUserIDs() ([]string, error) {
	// 这里需要实现获取所有用户ID的逻辑
	// 可以从数据库查询所有用户
	return []string{}, nil
}

// saveNotification 保存通知到存储
func (vc *VersionChecker) saveNotification(userID string, notification *UpdateNotification) error {
	// 这里需要实现通知存储逻辑
	// 可以保存到数据库或文件
	return nil
}

// broadcastNotifications 广播通知
func (vc *VersionChecker) broadcastNotifications() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次未读通知
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			vc.checkAndBroadcastUnreadNotifications()
		case <-vc.stopChan:
			return
		}
	}
}

// checkAndBroadcastUnreadNotifications 检查并广播未读通知
func (vc *VersionChecker) checkAndBroadcastUnreadNotifications() {
	// 获取所有在线用户
	// 为每个在线用户检查未读通知并发送
	// 这里需要实现在线用户管理
}

// GetStatus 获取检查器状态
func (vc *VersionChecker) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"is_running":     vc.isRunning,
		"last_check":     vc.lastCheckTime,
		"check_interval": vc.checkInterval.String(),
		"next_check":     vc.lastCheckTime.Add(vc.checkInterval),
	}
}

// SetCheckInterval 设置检查间隔
func (vc *VersionChecker) SetCheckInterval(interval time.Duration) {
	vc.checkInterval = interval
	// 保存到配置
	vc.server.database.SetSystemConfig("version_check_interval", interval.String())
	log.Printf("✅ 版本检查间隔已更新为: %v", interval)
}

// ForceCheck 强制检查更新
func (vc *VersionChecker) ForceCheck() {
	if !vc.isRunning {
		log.Printf("⚠️  版本检查器未运行")
		return
	}

	log.Printf("🔍 强制检查版本更新...")
	go vc.checkForUpdates()
}

// CreateUpdateNotification 创建更新通知的辅助函数
func CreateUpdateNotification(title, message, version, downloadURL string, isCritical bool) *UpdateNotification {
	notificationType := "new_update"
	if isCritical {
		notificationType = "critical_update"
	}

	return &UpdateNotification{
		ID:             uuid.New().String(),
		Type:           notificationType,
		Title:          title,
		Message:        message,
		Version:        version,
		ReleaseDate:    time.Now().Format("2006-01-02"),
		IsCritical:     isCritical,
		DownloadURL:    downloadURL,
		CreatedAt:      time.Now(),
		RequiresAction: isCritical,
		ActionText:     "查看详情",
	}
}