//go:build windows

package service

import (
	"fmt"
	"goalfy-mediaconverter/internal/config"
	"goalfy-mediaconverter/internal/installer"
	"goalfy-mediaconverter/internal/server"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	serviceName = "GoalfyMediaConverter"
	displayName = "Goalfy Media Converter Service"
	description = "视频转换服务 - 提供 WebM 到 MP4 转换功能"
)

// WindowsService 实现 Windows 服务接口
type WindowsService struct {
	server  *server.Server
	cfg     *config.Config
	logFile *os.File // 保持日志文件句柄不被关闭
}

// Execute 实现 svc.Handler 接口
func (ws *WindowsService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		log.Printf("加载配置失败: %v", err)
		changes <- svc.Status{State: svc.Stopped}
		return false, 1
	}

	// 设置日志文件输出
	logFile, err := setupLogFile(cfg)
	if err != nil {
		log.Printf("设置日志文件失败: %v", err)
		// 继续运行,使用默认日志输出
	} else {
		ws.logFile = logFile
		defer ws.logFile.Close() // 服务停止时关闭日志文件
	}

	// 查找 FFmpeg
	ffmpegInstaller := installer.NewFFmpegInstaller()
	ffmpegPath, err := ffmpegInstaller.FindFFmpeg()
	if err != nil {
		log.Printf("警告: FFmpeg 未找到: %v", err)
		ffmpegPath = ""
	} else {
		log.Printf("FFmpeg 已找到: %s", ffmpegPath)
	}
	cfg.FFmpegPath = ffmpegPath

	ws.cfg = cfg

	// 启动服务器
	ws.server = server.New(cfg)

	// 在 goroutine 中启动服务器
	go func() {
		if err := ws.server.Start(); err != nil {
			log.Printf("启动服务器失败: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(1 * time.Second)

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	log.Println("GoalfyMediaConverter 服务已启动")

	// 服务主循环
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				log.Println("收到停止信号,正在关闭服务...")
				break loop
			default:
				log.Printf("收到未知控制信号: %v", c)
			}
		}
	}

	changes <- svc.Status{State: svc.StopPending}
	// 这里可以添加清理逻辑
	log.Println("GoalfyMediaConverter 服务已停止")
	changes <- svc.Status{State: svc.Stopped}
	return false, 0
}

// RunService 运行 Windows 服务
func RunService(isDebug bool) error {
	if isDebug {
		return debug.Run(serviceName, &WindowsService{})
	}
	return svc.Run(serviceName, &WindowsService{})
}

// InstallService 安装 Windows 服务
func InstallService() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("连接服务管理器失败: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("服务 %s 已存在", serviceName)
	}

	s, err = m.CreateService(serviceName, exePath, mgr.Config{
		DisplayName: displayName,
		Description: description,
		StartType:   mgr.StartAutomatic, // 自动启动
	})
	if err != nil {
		return fmt.Errorf("创建服务失败: %w", err)
	}
	defer s.Close()

	// 设置事件日志
	err = eventlog.InstallAsEventCreate(serviceName, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("设置事件日志失败: %w", err)
	}

	return nil
}

// UninstallService 卸载 Windows 服务
func UninstallService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("连接服务管理器失败: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("打开服务失败: %w", err)
	}
	defer s.Close()

	// 停止服务
	status, err := s.Query()
	if err != nil {
		return fmt.Errorf("查询服务状态失败: %w", err)
	}

	if status.State != svc.Stopped {
		_, err = s.Control(svc.Stop)
		if err != nil {
			return fmt.Errorf("停止服务失败: %w", err)
		}

		// 等待服务停止
		timeout := time.Now().Add(10 * time.Second)
		for status.State != svc.Stopped {
			if time.Now().After(timeout) {
				return fmt.Errorf("等待服务停止超时")
			}
			time.Sleep(300 * time.Millisecond)
			status, err = s.Query()
			if err != nil {
				return fmt.Errorf("查询服务状态失败: %w", err)
			}
		}
	}

	// 删除服务
	err = s.Delete()
	if err != nil {
		return fmt.Errorf("删除服务失败: %w", err)
	}

	// 移除事件日志
	err = eventlog.Remove(serviceName)
	if err != nil {
		return fmt.Errorf("移除事件日志失败: %w", err)
	}

	return nil
}

// StartService 启动 Windows 服务
func StartService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("连接服务管理器失败: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("打开服务失败: %w", err)
	}
	defer s.Close()

	err = s.Start()
	if err != nil {
		return fmt.Errorf("启动服务失败: %w", err)
	}

	return nil
}

// StopService 停止 Windows 服务
func StopService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("连接服务管理器失败: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("打开服务失败: %w", err)
	}
	defer s.Close()

	_, err = s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("停止服务失败: %w", err)
	}

	return nil
}

// setupLogFile 设置日志文件输出
func setupLogFile(cfg *config.Config) (*os.File, error) {
	// 确定日志目录(使用程序所在目录下的 logs 文件夹)
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	exeDir := filepath.Dir(exePath)
	logDir := filepath.Join(exeDir, "logs")

	// 创建日志目录
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 日志文件路径:logs/service.log
	logFile := filepath.Join(logDir, "service.log")

	// 打开或创建日志文件(追加模式,带同步写入标志)
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}

	// 设置日志输出到文件(不输出到控制台,Windows 服务无控制台)
	log.SetOutput(file)

	// 设置日志格式
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("✅ 日志文件已设置: %s", logFile)
	log.Printf("======================================")
	log.Printf("  GoalfyMediaConverter 服务启动")
	log.Printf("======================================")

	return file, nil
}
