package main

import (
	"ffmpeg-binary/internal/autostart"
	"ffmpeg-binary/internal/config"
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

	// 启动服务器
	srv := server.New(cfg)
	if err := srv.Start(); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
