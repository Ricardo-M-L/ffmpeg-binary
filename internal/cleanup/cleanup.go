package cleanup

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	appPath          = "/Applications/FFmpeg-Binary.app"
	checkInterval    = 10 * time.Second // 每10秒检查一次
	launchAgentPath  = "Library/LaunchAgents/com.ffmpeg.binary.plist"
	watcherAgentPath = "Library/LaunchAgents/com.ffmpeg.binary.watcher.plist"
	dataDir          = ".ffmpeg-binary"
	supportDir       = "Library/Application Support/FFmpeg-Binary"
)

// Watcher 自清理监控器
type Watcher struct {
	stopCh chan struct{}
}

// NewWatcher 创建新的监控器
func NewWatcher() *Watcher {
	return &Watcher{
		stopCh: make(chan struct{}),
	}
}

// Start 启动监控
func (w *Watcher) Start() {
	go w.watch()
}

// Stop 停止监控
func (w *Watcher) Stop() {
	close(w.stopCh)
}

// watch 监控应用包是否存在
func (w *Watcher) watch() {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.checkAndCleanup()
		case <-w.stopCh:
			return
		}
	}
}

// checkAndCleanup 检查应用是否存在,如果不存在则执行清理
func (w *Watcher) checkAndCleanup() {
	// 检查应用包是否还在 /Applications/
	_, err := os.Stat(appPath)
	appExists := err == nil

	if appExists {
		// 应用还在,什么都不做
		return
	}

	// 应用不在 /Applications/ 了,检查是否在废纸篓
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("获取用户目录失败: %v", err)
		return
	}

	trashPath := filepath.Join(homeDir, ".Trash", "FFmpeg-Binary.app")
	_, err = os.Stat(trashPath)
	inTrash := err == nil

	if !inTrash {
		// 既不在 /Applications/ 也不在废纸篓
		// 可能是被彻底删除了,或者正在移动过程中
		// 为了安全起见,再等一个周期确认
		return
	}

	// 确认应用在废纸篓中,执行清理
	log.Println("检测到应用已被移到废纸篓,开始执行自清理...")
	w.performCleanup()

	// 清理完成后退出程序
	log.Println("清理完成,程序即将退出")
	time.Sleep(1 * time.Second)
	os.Exit(0)
}

// performCleanup 执行清理操作
func (w *Watcher) performCleanup() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("获取用户目录失败: %v", err)
		return
	}

	// 1. 卸载主服务 LaunchAgent
	launchAgentFile := filepath.Join(homeDir, launchAgentPath)
	if err := w.unloadLaunchAgent(launchAgentFile); err != nil {
		log.Printf("卸载主服务 LaunchAgent 失败: %v", err)
	} else {
		log.Println("✓ 已卸载主服务 LaunchAgent")
	}

	// 2. 卸载监控服务 LaunchAgent (如果存在)
	watcherAgentFile := filepath.Join(homeDir, watcherAgentPath)
	if err := w.unloadLaunchAgent(watcherAgentFile); err != nil {
		log.Printf("卸载监控服务 LaunchAgent 失败: %v", err)
	} else {
		log.Println("✓ 已卸载监控服务 LaunchAgent")
	}

	// 3. 删除数据目录
	dataDirPath := filepath.Join(homeDir, dataDir)
	if err := os.RemoveAll(dataDirPath); err != nil {
		log.Printf("删除数据目录失败: %v", err)
	} else {
		log.Printf("✓ 已删除数据目录: %s", dataDirPath)
	}

	// 4. 删除 Application Support 目录
	supportDirPath := filepath.Join(homeDir, supportDir)
	if err := os.RemoveAll(supportDirPath); err != nil {
		log.Printf("删除 Application Support 目录失败: %v", err)
	} else {
		log.Printf("✓ 已删除 Application Support 目录: %s", supportDirPath)
	}

	log.Println("所有清理操作已完成")
}

// unloadLaunchAgent 卸载并删除 LaunchAgent
func (w *Watcher) unloadLaunchAgent(plistPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return nil // 文件不存在,无需卸载
	}

	// 卸载 LaunchAgent
	cmd := exec.Command("launchctl", "unload", plistPath)
	if err := cmd.Run(); err != nil {
		log.Printf("launchctl unload 失败: %v", err)
		// 继续删除文件
	}

	// 删除 plist 文件
	if err := os.Remove(plistPath); err != nil {
		return err
	}

	return nil
}
