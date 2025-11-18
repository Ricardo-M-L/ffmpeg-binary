//go:build windows

package main

import (
	"fmt"
	"goalfy-mediaconverter/internal/service"
	"log"
	"os"

	"golang.org/x/sys/windows/svc"
)

func runAsWindowsService() {
	// 检查命令行参数
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install-service":
			// 安装 Windows 服务
			if err := service.InstallService(); err != nil {
				log.Fatalf("安装服务失败: %v", err)
			}
			fmt.Println("✅ Windows 服务安装成功")
			fmt.Println("使用 'sc start GoalfyMediaConverter' 启动服务")
			fmt.Println("或者重启计算机后自动启动")
			return

		case "uninstall-service":
			// 卸载 Windows 服务
			if err := service.UninstallService(); err != nil {
				log.Fatalf("卸载服务失败: %v", err)
			}
			fmt.Println("✅ Windows 服务卸载成功")
			return

		case "start-service":
			// 启动服务
			if err := service.StartService(); err != nil {
				log.Fatalf("启动服务失败: %v", err)
			}
			fmt.Println("✅ 服务启动成功")
			return

		case "stop-service":
			// 停止服务
			if err := service.StopService(); err != nil {
				log.Fatalf("停止服务失败: %v", err)
			}
			fmt.Println("✅ 服务停止成功")
			return

		case "debug":
			// 调试模式(在控制台运行,便于调试)
			runAsConsole()
			return
		}
	}

	// 检查是否在服务控制管理器中运行
	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("检查会话类型失败: %v", err)
	}

	if isIntSess {
		// 交互式会话(控制台),显示帮助信息
		fmt.Println("GoalfyMediaConverter - 视频转换服务")
		fmt.Println("")
		fmt.Println("用法:")
		fmt.Println("  ffmpeg-binary.exe install-service   - 安装为 Windows 服务")
		fmt.Println("  ffmpeg-binary.exe uninstall-service - 卸载 Windows 服务")
		fmt.Println("  ffmpeg-binary.exe start-service     - 启动服务")
		fmt.Println("  ffmpeg-binary.exe stop-service      - 停止服务")
		fmt.Println("  ffmpeg-binary.exe debug             - 调试模式(控制台运行)")
		fmt.Println("")
		fmt.Println("服务管理:")
		fmt.Println("  sc start GoalfyMediaConverter       - 启动服务")
		fmt.Println("  sc stop GoalfyMediaConverter        - 停止服务")
		fmt.Println("  sc query GoalfyMediaConverter       - 查看服务状态")
		return
	}

	// 作为 Windows 服务运行
	if err := service.RunService(false); err != nil {
		log.Fatalf("运行服务失败: %v", err)
	}
}
