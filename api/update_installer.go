package api

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// UpdateInstaller 更新安装器
type UpdateInstaller struct {
	server         *Server
	downloadPath   string
	backupPath     string
	installPath    string
	tempDir        string
}

// NewUpdateInstaller 创建更新安装器
func NewUpdateInstaller(server *Server) *UpdateInstaller {
	tempDir := filepath.Join(os.TempDir(), "nofx-updates")

	return &UpdateInstaller{
		server:       server,
		downloadPath: filepath.Join(tempDir, "downloads"),
		backupPath:   filepath.Join(tempDir, "backups"),
		installPath:  tempDir, // 应该是当前应用目录
		tempDir:      tempDir,
	}
}

// DownloadUpdate 下载更新
func (ui *UpdateInstaller) DownloadUpdate(versionInfo *VersionInfo, progressCallback func(*UpdateProgress)) error {
	startTime := time.Now()

	// 确保下载目录存在
	if err := os.MkdirAll(ui.downloadPath, 0755); err != nil {
		return fmt.Errorf("创建下载目录失败: %v", err)
	}

	// 发送开始下载进度
	if progressCallback != nil {
		progressCallback(&UpdateProgress{
			Status:  "downloading",
			Message: "准备下载更新文件...",
			Progress: 0,
		})
	}

	// 构建下载文件路径
	filename := fmt.Sprintf("nofx-v%s-%s.zip", versionInfo.Version, getPlatformString())
	downloadFile := filepath.Join(ui.downloadPath, filename)

	// 检查文件是否已存在
	if _, err := os.Stat(downloadFile); err == nil {
		// 验证现有文件
		if err := ui.verifyDownloadedFile(downloadFile, versionInfo); err == nil {
			if progressCallback != nil {
				progressCallback(&UpdateProgress{
					Status:  "downloading",
					Message: "更新文件已存在，跳过下载",
					Progress: 100,
				})
			}
			log.Printf("✅ 更新文件已存在: %s", downloadFile)
			return nil
		} else {
			log.Printf("⚠️  现有文件验证失败，重新下载: %v", err)
			os.Remove(downloadFile)
		}
	}

	// 开始下载
	log.Printf("📥 开始下载更新: %s", versionInfo.DownloadURL)

	resp, err := http.Get(versionInfo.DownloadURL)
	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 创建文件
	file, err := os.Create(downloadFile)
	if err != nil {
		return fmt.Errorf("创建下载文件失败: %v", err)
	}
	defer file.Close()

	// 获取文件大小
	totalSize := resp.ContentLength
	if totalSize <= 0 {
		totalSize = versionInfo.UpdateSize
	}

	// 进度跟踪
	var downloaded int64
	lastProgress := 0
	buffer := make([]byte, 32*1024) // 32KB buffer

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			written, writeErr := file.Write(buffer[:n])
			if writeErr != nil {
				return fmt.Errorf("写入文件失败: %v", writeErr)
			}
			downloaded += int64(written)

			// 计算进度
			if totalSize > 0 {
				progress := float64(downloaded) / float64(totalSize) * 100
				if int(progress) > lastProgress {
					lastProgress = int(progress)

					// 计算下载速度和预计时间
					elapsed := time.Since(startTime).Seconds()
					speed := int64(float64(downloaded) / elapsed)
					var eta int64
					if speed > 0 {
						eta = int64((float64(totalSize-downloaded) / float64(speed)))
					}

					if progressCallback != nil {
						progressCallback(&UpdateProgress{
							Status:     "downloading",
							Message:    fmt.Sprintf("正在下载... %.1f%%", progress),
							Progress:   progress,
							TotalSize:  totalSize,
							Downloaded: downloaded,
							Speed:      speed,
							ETA:        eta,
						})
					}
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("下载中断: %v", err)
		}
	}

	// 验证下载的文件
	if progressCallback != nil {
		progressCallback(&UpdateProgress{
			Status:  "verifying",
			Message: "验证下载文件...",
			Progress: 95,
		})
	}

	if err := ui.verifyDownloadedFile(downloadFile, versionInfo); err != nil {
		os.Remove(downloadFile)
		return fmt.Errorf("文件验证失败: %v", err)
	}

	// 下载完成
	if progressCallback != nil {
		progressCallback(&UpdateProgress{
			Status:  "downloading",
			Message: "下载完成",
			Progress: 100,
		})
	}

	log.Printf("✅ 下载完成: %s (%.2f MB)", downloadFile, float64(downloaded)/1024/1024)
	return nil
}

// verifyDownloadedFile 验证下载的文件
func (ui *UpdateInstaller) verifyDownloadedFile(filePath string, versionInfo *VersionInfo) error {
	// 检查文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("无法获取文件信息: %v", err)
	}

	if versionInfo.UpdateSize > 0 && fileInfo.Size() != versionInfo.UpdateSize {
		return fmt.Errorf("文件大小不匹配: 期望 %d, 实际 %d", versionInfo.UpdateSize, fileInfo.Size())
	}

	// 验证校验和
	if versionInfo.Checksum != "" {
		fileChecksum, err := ui.calculateFileChecksum(filePath)
		if err != nil {
			return fmt.Errorf("计算文件校验和失败: %v", err)
		}

		if !strings.EqualFold(fileChecksum, versionInfo.Checksum) {
			return fmt.Errorf("文件校验和不匹配: 期望 %s, 实际 %s", versionInfo.Checksum, fileChecksum)
		}
	}

	// 验证ZIP文件
	if err := ui.validateZipFile(filePath); err != nil {
		return fmt.Errorf("ZIP文件验证失败: %v", err)
	}

	return nil
}

// calculateFileChecksum 计算文件校验和
func (ui *UpdateInstaller) calculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// validateZipFile 验证ZIP文件
func (ui *UpdateInstaller) validateZipFile(filePath string) error {
	file, err := zip.OpenReader(filePath)
	if err != nil {
		return fmt.Errorf("打开ZIP文件失败: %v", err)
	}
	defer file.Close()

	// 检查ZIP文件是否损坏
	for _, f := range file.File {
		if f.Method == zip.Store {
			// 存储方法，检查文件是否可读
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("ZIP文件中文件 %s 读取失败: %v", f.Name, err)
			}
			rc.Close()
		}
	}

	return nil
}

// BackupCurrentVersion 备份当前版本
func (ui *UpdateInstaller) BackupCurrentVersion(progressCallback func(*UpdateProgress)) error {
	if progressCallback != nil {
		progressCallback(&UpdateProgress{
			Status:  "backup",
			Message: "备份当前版本...",
			Progress: 5,
		})
	}

	// 确保备份目录存在
	if err := os.MkdirAll(ui.backupPath, 0755); err != nil {
		return fmt.Errorf("创建备份目录失败: %v", err)
	}

	// 创建备份文件名
	currentVersion := getAppVersion()
	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("nofx-v%s-%s-%s.zip", currentVersion, getPlatformString(), timestamp)
	backupFile := filepath.Join(ui.backupPath, backupName)

	// 获取当前应用目录（假设是执行文件的目录）
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取执行文件路径失败: %v", err)
	}
	appDir := filepath.Dir(execPath)

	log.Printf("📦 开始备份当前版本: %s -> %s", appDir, backupFile)

	// 创建ZIP备份文件
	if err := ui.createZipBackup(appDir, backupFile, progressCallback); err != nil {
		return fmt.Errorf("创建备份失败: %v", err)
	}

	if progressCallback != nil {
		progressCallback(&UpdateProgress{
			Status:  "backup",
			Message: "备份完成",
			Progress: 15,
		})
	}

	log.Printf("✅ 备份完成: %s", backupFile)
	return nil
}

// createZipBackup 创建ZIP备份
func (ui *UpdateInstaller) createZipBackup(sourceDir, backupFile string, progressCallback func(*UpdateProgress)) error {
	backupFileWriter, err := os.Create(backupFile)
	if err != nil {
		return err
	}
	defer backupFileWriter.Close()

	zipWriter := zip.NewWriter(backupFileWriter)
	defer zipWriter.Close()

	// 遍历源目录并添加到ZIP
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// 跳过临时文件和目录
		if strings.Contains(relPath, "temp") || strings.Contains(relPath, ".git") {
			return nil
		}

		// 创建ZIP文件头
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = relPath

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// 复制文件内容
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

// InstallUpdate 安装更新
func (ui *UpdateInstaller) InstallUpdate(versionInfo *VersionInfo, progressCallback func(*UpdateProgress)) error {
	if progressCallback != nil {
		progressCallback(&UpdateProgress{
			Status:  "installing",
			Message: "准备安装更新...",
			Progress: 85,
		})
	}

	// 找到下载的更新文件
	filename := fmt.Sprintf("nofx-v%s-%s.zip", versionInfo.Version, getPlatformString())
	updateFile := filepath.Join(ui.downloadPath, filename)

	// 检查文件是否存在
	if _, err := os.Stat(updateFile); os.IsNotExist(err) {
		return fmt.Errorf("更新文件不存在: %s", updateFile)
	}

	// 解压更新文件
	tempExtractDir := filepath.Join(ui.tempDir, "extract-"+versionInfo.Version)
	if err := ui.extractUpdate(updateFile, tempExtractDir, progressCallback); err != nil {
		return fmt.Errorf("解压更新文件失败: %v", err)
	}

	// 执行安装脚本或文件
	if err := ui.performInstallation(tempExtractDir, progressCallback); err != nil {
		return fmt.Errorf("执行安装失败: %v", err)
	}

	if progressCallback != nil {
		progressCallback(&UpdateProgress{
			Status:  "completed",
			Message: "安装完成，准备重启...",
			Progress: 100,
		})
	}

	log.Printf("✅ 更新安装完成: v%s", versionInfo.Version)
	return nil
}

// extractUpdate 解压更新文件
func (ui *UpdateInstaller) extractUpdate(updateFile, extractDir string, progressCallback func(*UpdateProgress)) error {
	if progressCallback != nil {
		progressCallback(&UpdateProgress{
			Status:  "installing",
			Message: "解压更新文件...",
			Progress: 87,
		})
	}

	// 确保解压目录存在
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("创建解压目录失败: %v", err)
	}

	// 打开ZIP文件
	reader, err := zip.OpenReader(updateFile)
	if err != nil {
		return fmt.Errorf("打开更新文件失败: %v", err)
	}
	defer reader.Close()

	// 解压文件
	for _, file := range reader.File {
		path := filepath.Join(extractDir, file.Name)

		// 确保路径在解压目录内（防止路径遍历攻击）
		if !strings.HasPrefix(path, extractDir+string(os.PathSeparator)) {
			return fmt.Errorf("非法文件路径: %s", file.Name)
		}

		// 创建目录
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.FileInfo().Mode())
			continue
		}

		// 创建文件
		fileReader, err := file.Open()
		if err != nil {
			return fmt.Errorf("打开文件 %s 失败: %v", file.Name, err)
		}

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			fileReader.Close()
			return fmt.Errorf("创建文件 %s 失败: %v", path, err)
		}

		_, err = io.Copy(targetFile, fileReader)
		fileReader.Close()
		targetFile.Close()

		if err != nil {
			return fmt.Errorf("复制文件 %s 失败: %v", file.Name, err)
		}
	}

	log.Printf("✅ 解压完成: %s", extractDir)
	return nil
}

// performInstallation 执行安装
func (ui *UpdateInstaller) performInstallation(extractDir string, progressCallback func(*UpdateProgress)) error {
	if progressCallback != nil {
		progressCallback(&UpdateProgress{
			Status:  "installing",
			Message: "正在安装文件...",
			Progress: 92,
		})
	}

	// 获取当前应用目录
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取执行文件路径失败: %v", err)
	}
	appDir := filepath.Dir(execPath)

	// 查找安装脚本
	installScript := filepath.Join(extractDir, "install.sh")
	if runtime.GOOS == "windows" {
		installScript = filepath.Join(extractDir, "install.bat")
	}

	// 如果有安装脚本，执行它
	if _, err := os.Stat(installScript); err == nil {
		log.Printf("🔧 执行安装脚本: %s", installScript)
		return ui.executeInstallScript(installScript, appDir, extractDir)
	}

	// 否则，执行文件替换
	return ui.performFileReplacement(extractDir, appDir, progressCallback)
}

// executeInstallScript 执行安装脚本
func (ui *UpdateInstaller) executeInstallScript(scriptPath, appDir, extractDir string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", scriptPath, appDir, extractDir)
	} else {
		cmd = exec.Command("bash", scriptPath, appDir, extractDir)
	}

	// 设置工作目录
	cmd.Dir = extractDir

	// 执行脚本
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("执行安装脚本失败: %v\n输出: %s", err, string(output))
	}

	log.Printf("✅ 安装脚本执行完成:\n%s", string(output))
	return nil
}

// performFileReplacement 执行文件替换
func (ui *UpdateInstaller) performFileReplacement(extractDir, appDir string, progressCallback func(*UpdateProgress)) error {
	if progressCallback != nil {
		progressCallback(&UpdateProgress{
			Status:  "installing",
			Message: "正在替换应用文件...",
			Progress: 95,
		})
	}

	// 遍历解压目录，复制文件到应用目录
	return filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(extractDir, path)
		if err != nil {
			return err
		}

		// 跳过安装脚本等特殊文件
		if strings.Contains(relPath, "install.") || strings.Contains(relPath, ".DS_Store") {
			return nil
		}

		targetPath := filepath.Join(appDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		// 复制文件
		return copyFile(path, targetPath)
	})
}

// copyFile 复制文件
func copyFile(src, dst string) error {
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

// Cleanup 清理临时文件
func (ui *UpdateInstaller) Cleanup() error {
	log.Printf("🧹 清理临时文件...")

	// 清理解压目录
	if err := os.RemoveAll(filepath.Join(ui.tempDir, "extract-")); err != nil {
		log.Printf("⚠️  清理解压目录失败: %v", err)
	}

	// 保留最近的几个备份文件，删除旧的
	return ui.cleanupOldBackups()
}

// cleanupOldBackups 清理旧备份
func (ui *UpdateInstaller) cleanupOldBackups() error {
	files, err := filepath.Glob(filepath.Join(ui.backupPath, "*.zip"))
	if err != nil {
		return err
	}

	// 按修改时间排序，保留最新的5个
	if len(files) > 5 {
		// 按修改时间排序
		for i := 0; i < len(files)-1; i++ {
			for j := i + 1; j < len(files); j++ {
				info1, _ := os.Stat(files[i])
				info2, _ := os.Stat(files[j])
				if info1.ModTime().Before(info2.ModTime()) {
					files[i], files[j] = files[j], files[i]
				}
			}
		}

		// 删除旧文件
		for i := 0; i < len(files)-5; i++ {
			if err := os.Remove(files[i]); err != nil {
				log.Printf("⚠️  删除旧备份失败 %s: %v", files[i], err)
			} else {
				log.Printf("🗑️  已删除旧备份: %s", files[i])
			}
		}
	}

	return nil
}