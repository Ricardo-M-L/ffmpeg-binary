package main

import (
	"ffmpeg-binary/internal/autostart"
	"ffmpeg-binary/internal/cleanup"
	"ffmpeg-binary/internal/config"
	"ffmpeg-binary/internal/installer"
	"ffmpeg-binary/internal/server"
	"fmt"
	"log"
	"os"
)

func main() {
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
		}
	}

	// 启动自清理监控(每10秒检查应用包是否存在)
	cleanupWatcher := cleanup.NewWatcher()
	cleanupWatcher.Start()
	log.Println("✓ 自清理监控已启动")

	// 启动服务器
	srv := server.New(cfg)
	if err := srv.Start(); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
