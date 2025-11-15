package api

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// RestartManager 重启管理器
type RestartManager struct {
	server          *Server
	restartTime     time.Time
	restartReason   string
	countdownTime   int
	countdownActive bool
	countdownMutex  sync.Mutex
	cancelChan      chan bool
}

// RestartRequest 重启请求
type RestartRequest struct {
	DelaySeconds int    `json:"delay_seconds"` // 延迟秒数
	Reason       string `json:"reason"`        // 重启原因
	Force        bool   `json:"force"`         // 强制重启
	AutoUpdate   bool   `json:"auto_update"`   // 自动更新后重启
}

// RestartStatus 重启状态
type RestartStatus struct {
	IsCountingDown bool        `json:"is_counting_down"`
	CountdownTime  int         `json:"countdown_time"`  // 倒计时秒数
	RestartTime    time.Time   `json:"restart_time"`    // 计划重启时间
	Reason         string      `json:"reason"`          // 重启原因
	CanCancel      bool        `json:"can_cancel"`      // 是否可以取消
	Message        string      `json:"message"`         // 状态消息
	LastRestart    *time.Time  `json:"last_restart"`    // 上次重启时间
}

// NewRestartManager 创建重启管理器
func NewRestartManager(server *Server) *RestartManager {
	return &RestartManager{
		server:     server,
		cancelChan: make(chan bool, 1),
	}
}

// ScheduleRestart 计划重启
func (rm *RestartManager) ScheduleRestart(request RestartRequest) error {
	rm.countdownMutex.Lock()
	defer rm.countdownMutex.Unlock()

	if rm.countdownActive && !request.Force {
		return fmt.Errorf("已有重启计划在进行中")
	}

	// 设置倒计时时间
	delayTime := request.DelaySeconds
	if delayTime <= 0 {
		delayTime = 10 // 默认10秒
	}

	if delayTime > 300 {
		delayTime = 300 // 最大5分钟
	}

	rm.countdownTime = delayTime
	rm.restartReason = request.Reason
	rm.restartTime = time.Now().Add(time.Duration(delayTime) * time.Second)
	rm.countdownActive = true
	rm.cancelChan = make(chan bool, 1)

	log.Printf("🔄 计划重启应用: %d秒后重启，原因: %s", delayTime, request.Reason)

	// 启动倒计时
	go rm.startCountdown()

	return nil
}

// startCountdown 开始倒计时
func (rm *RestartManager) startCountdown() {
	rm.countdownMutex.Lock()
	if !rm.countdownActive {
		rm.countdownMutex.Unlock()
		return
	}
	rm.countdownMutex.Unlock()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	remaining := rm.countdownTime

	for remaining > 0 {
		rm.countdownMutex.Lock()
		if !rm.countdownActive {
			rm.countdownMutex.Unlock()
			return
		}

		// 广播倒计时状态
		rm.broadcastCountdownStatus(remaining, rm.restartReason)
		rm.countdownMutex.Unlock()

		select {
		case <-ticker.C:
			remaining--
		case <-rm.cancelChan:
			log.Printf("⏹️  重启已取消")
			rm.countdownMutex.Lock()
			rm.countdownActive = false
			rm.countdownMutex.Unlock()
			return
		}
	}

	// 执行重启
	log.Printf("🚀 执行应用重启...")
	rm.countdownMutex.Lock()
	rm.countdownActive = false
	rm.countdownMutex.Unlock()

	rm.performRestart()
}

// CancelRestart 取消重启
func (rm *RestartManager) CancelRestart() error {
	rm.countdownMutex.Lock()
	defer rm.countdownMutex.Unlock()

	if !rm.countdownActive {
		return fmt.Errorf("没有正在进行的重启计划")
	}

	select {
	case rm.cancelChan <- true:
		log.Printf("✅ 重启计划已取消")
		return nil
	default:
		return fmt.Errorf("无法取消重启")
	}
}

// ForceRestart 立即重启
func (rm *RestartManager) ForceRestart(reason string) {
	log.Printf("🚀 立即重启应用: %s", reason)
	rm.performRestart()
}

// GetRestartStatus 获取重启状态
func (rm *RestartManager) GetRestartStatus() *RestartStatus {
	rm.countdownMutex.Lock()
	defer rm.countdownMutex.Unlock()

	status := &RestartStatus{
		IsCountingDown: rm.countdownActive,
		CountdownTime:  rm.countdownTime,
		RestartTime:    rm.restartTime,
		Reason:         rm.restartReason,
		CanCancel:      rm.countdownActive && rm.countdownTime > 5, // 5秒内不能取消
	}

	if rm.countdownActive {
		remaining := int(rm.restartTime.Sub(time.Now()).Seconds())
		if remaining > 0 {
			status.Message = fmt.Sprintf("应用将在 %d 秒后重启", remaining)
		} else {
			status.Message = "正在重启..."
		}
	} else {
		status.Message = "没有重启计划"
	}

	return status
}

// broadcastCountdownStatus 广播倒计时状态
func (rm *RestartManager) broadcastCountdownStatus(seconds int, reason string) {
	// 通过WebSocket广播倒计时状态
	_ = map[string]interface{}{
		"type":     "restart_countdown",
		"seconds":  seconds,
		"reason":   reason,
		"canCancel": seconds > 5,
		"timestamp": time.Now().Unix(),
	}

	// 这里应该通过WebSocket管理器广播消息
	// 例如: websocketManager.BroadcastToAll(message)

	if seconds <= 10 {
		log.Printf("⏰ 应用重启倒计时: %d秒", seconds)
	}
}

// performRestart 执行重启
func (rm *RestartManager) performRestart() {
	// 保存重启记录
	rm.saveRestartRecord()

	// 停止服务器
	if rm.server != nil {
		log.Printf("🛑 停止API服务器...")

		if err := rm.server.Shutdown(); err != nil {
			log.Printf("⚠️  服务器关闭失败: %v", err)
		}
	}

	// 优雅地重启应用
	rm.restartApplication()
}

// restartApplication 重启应用程序
func (rm *RestartManager) restartApplication() {
	log.Printf("🔄 重启应用程序...")

	// 根据平台执行不同的重启逻辑
	switch runtime.GOOS {
	case "windows":
		rm.restartWindows()
	case "linux", "darwin":
		rm.restartUnix()
	default:
		log.Printf("❌ 不支持的操作系统: %s", runtime.GOOS)
		os.Exit(0)
	}
}

// restartWindows Windows重启
func (rm *RestartManager) restartWindows() {
	// 获取当前执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("❌ 获取执行文件路径失败: %v", err)
		os.Exit(1)
	}

	// 使用cmd.exe启动新进程
	cmd := exec.Command("cmd", "/C", "start", "cmd", "/C", execPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 启动新进程
	if err := cmd.Start(); err != nil {
		log.Printf("❌ 启动新进程失败: %v", err)
		os.Exit(1)
	}

	// 给新进程一点启动时间
	time.Sleep(2 * time.Second)

	// 退出当前进程
	os.Exit(0)
}

// restartUnix Unix/Linux/macOS重启
func (rm *RestartManager) restartUnix() {
	// 获取当前执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("❌ 获取执行文件路径失败: %v", err)
		os.Exit(1)
	}

	// 获取执行文件目录
	execDir := filepath.Dir(execPath)

	// 使用nohup启动新进程（确保进程在后台运行）
	args := []string{execPath}
	if len(os.Args) > 1 {
		args = append(args, os.Args[1:]...)
	}

	cmd := exec.Command("nohup", args...)
	cmd.Dir = execDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 设置进程组 (Linux/Mac专用)
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	// 启动新进程
	if err := cmd.Start(); err != nil {
		log.Printf("❌ 启动新进程失败: %v", err)
		os.Exit(1)
	}

	log.Printf("✅ 新进程已启动，PID: %d", cmd.Process.Pid)

	// 给新进程一点启动时间
	time.Sleep(2 * time.Second)

	// 退出当前进程
	os.Exit(0)
}

// saveRestartRecord 保存重启记录
func (rm *RestartManager) saveRestartRecord() {
	// 保存重启记录到数据库或文件
	restartRecord := map[string]interface{}{
		"restart_time": rm.restartTime,
		"reason":       rm.restartReason,
		"version":      getAppVersion(),
		"platform":     getPlatformInfo(),
	}

	log.Printf("💾 保存重启记录: %+v", restartRecord)

	// 这里可以实现将记录保存到数据库的逻辑
	// 例如: rm.server.database.SaveRestartRecord(restartRecord)
}

// ScheduleRestartAfterUpdate 更新后重启
func (rm *RestartManager) ScheduleRestartAfterUpdate(versionInfo *VersionInfo) error {
	reason := fmt.Sprintf("自动更新到版本 %s", versionInfo.Version)

	// 如果是关键更新，立即重启
	if versionInfo.IsCriticalUpdate {
		return rm.ScheduleRestart(RestartRequest{
			DelaySeconds: 5, // 关键更新5秒后重启
			Reason:       reason,
			Force:        true,
			AutoUpdate:   true,
		})
	}

	// 普通更新，给用户更多时间
	return rm.ScheduleRestart(RestartRequest{
		DelaySeconds: 30, // 普通更新30秒后重启
		Reason:       reason,
		Force:        false,
		AutoUpdate:   true,
	})
}

// CheckAndHandleGracefulShutdown 检查并处理优雅关闭
func (rm *RestartManager) CheckAndHandleGracefulShutdown() {
	// 如果有重启计划，提前停止接受新请求
	rm.countdownMutex.Lock()
	if rm.countdownActive {
		remaining := int(rm.restartTime.Sub(time.Now()).Seconds())
		if remaining > 0 && remaining <= 10 {
			log.Printf("⏰ 即将重启，停止接受新请求...")
			// 这里可以设置服务器为维护模式
			// 例如: rm.server.SetMaintenanceMode(true)
		}
	}
	rm.countdownMutex.Unlock()
}

// GetRestartHistory 获取重启历史
func (rm *RestartManager) GetRestartHistory() ([]map[string]interface{}, error) {
	// 从数据库或文件读取重启历史
	// 这里返回模拟数据
	history := []map[string]interface{}{
		{
			"restart_time": time.Now().Add(-24 * time.Hour),
			"reason":       "定期维护重启",
			"version":      "1.0.0",
			"success":      true,
		},
		{
			"restart_time": time.Now().Add(-7 * 24 * time.Hour),
			"reason":       "版本更新到v1.0.0",
			"version":      "1.0.0",
			"success":      true,
		},
	}

	return history, nil
}