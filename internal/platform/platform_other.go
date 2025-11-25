//go:build !windows

package platform

import (
	"fmt"
	"goalfy-mediaconverter/internal/autostart"
	"goalfy-mediaconverter/internal/cleanup"
	"goalfy-mediaconverter/internal/config"
	"goalfy-mediaconverter/internal/installer"
	"goalfy-mediaconverter/internal/server"
	"log"
	"os"
)

// start 是非 Windows 平台的实现
// 直接以控制台模式运行
func start() {
	runAsConsole()
}

// runAsConsole 控制台模式运行
func runAsConsole() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 只查找 FFmpeg,不自动安装 (安装由 PKG 的 postinstall 脚本处理)
	ffmpegInstaller := installer.NewFFmpegInstaller()
	ffmpegPath, err := ffmpegInstaller.FindFFmpeg()
	if err != nil {
		log.Printf("⚠️  警告: FFmpeg 未找到: %v", err)
		log.Printf("提示: 如果您使用 PKG 安装,请重新安装 PKG;如果开发环境,请手动安装 FFmpeg")
		// 继续运行,但 FFmpeg 功能将不可用
		ffmpegPath = ""
	} else {
		log.Printf("✅ FFmpeg 已找到: %s", ffmpegPath)
	}

	// 更新配置中的 FFmpeg 路径
	cfg.FFmpegPath = ffmpegPath

	// 检查命令行参数
	devMode := false
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			// 安装自启动
			if err := autostart.Install(); err != nil {
				log.Fatalf("安装自启动失败: %v", err)
			}
			fmt.Println("自启动安装成功")
			return
		case "uninstall":
			// 卸载自启动
			if err := autostart.Uninstall(); err != nil {
				log.Fatalf("卸载自启动失败: %v", err)
			}
			fmt.Println("自启动卸载成功")
			return
		case "dev":
			// 开发模式:跳过自清理监控
			devMode = true
			log.Println("🔧 开发模式已启用,跳过自清理监控")
		}
	}

	// 检查环境变量 (用于判断是否为开发模式)
	if os.Getenv("GOALFY_DEV_MODE") == "true" {
		devMode = true
		log.Println("🔧 开发模式已启用 (通过环境变量),跳过自清理监控")
	}

	// 只在非开发模式下启动自清理监控
	if !devMode {
		cleanupWatcher := cleanup.NewWatcher()
		cleanupWatcher.Start()
		log.Println("✓ 自清理监控已启动")
	}

	// 启动服务器
	srv := server.New(cfg)
	if err := srv.Start(); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
