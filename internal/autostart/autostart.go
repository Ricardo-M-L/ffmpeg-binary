package autostart

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Install 安装自启动
func Install() error {
	switch runtime.GOOS {
	case "darwin":
		return installMacOS()
	case "windows":
		return installWindows()
	case "linux":
		return installLinux()
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// Uninstall 卸载自启动
func Uninstall() error {
	switch runtime.GOOS {
	case "darwin":
		return uninstallMacOS()
	case "windows":
		return uninstallWindows()
	case "linux":
		return uninstallLinux()
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// installMacOS 安装 macOS 自启动 (launchd)
func installMacOS() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// LaunchAgents 目录
	launchAgentsDir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		return err
	}

	// plist 文件路径
	plistPath := filepath.Join(launchAgentsDir, "com.ffmpeg.binary.plist")

	// 获取当前可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	// plist 内容
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.ffmpeg.binary</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>EnvironmentVariables</key>
	<dict>
		<key>PATH</key>
		<string>/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
	</dict>
	<key>StandardOutPath</key>
	<string>%s/Library/Logs/ffmpeg-binary.log</string>
	<key>StandardErrorPath</key>
	<string>%s/Library/Logs/ffmpeg-binary.err</string>
</dict>
</plist>`, exePath, home, home)

	// 写入 plist 文件
	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return err
	}

	// 加载服务
	// launchctl load 命令在较新的 macOS 版本中已废弃,这里提供说明
	fmt.Println("macOS 自启动配置文件已创建:", plistPath)
	fmt.Println("请注意: 服务将在下次登录时自动启动")
	fmt.Println("如需立即启动,请运行: launchctl load", plistPath)

	return nil
}

// uninstallMacOS 卸载 macOS 自启动
func uninstallMacOS() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.ffmpeg.binary.plist")

	// 1. 先停止并卸载 launchd 服务
	fmt.Println("正在停止 launchd 服务...")
	if err := execCommand("launchctl", "unload", plistPath); err != nil {
		fmt.Printf("警告: launchctl unload 失败 (服务可能未运行): %v\n", err)
		// 继续执行,不返回错误
	}

	// 2. 删除 plist 文件
	fmt.Println("正在删除自启动配置文件...")
	if err := os.Remove(plistPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("自启动配置文件不存在,已跳过")
		} else {
			return fmt.Errorf("删除配置文件失败: %v", err)
		}
	} else {
		fmt.Println("✅ 自启动配置已删除")
	}

	// 3. 停止正在运行的进程
	fmt.Println("正在停止服务进程...")
	if err := execCommand("pkill", "-f", "ffmpeg-binary"); err != nil {
		fmt.Println("警告: 未找到运行中的服务进程")
	} else {
		fmt.Println("✅ 服务进程已停止")
	}

	fmt.Println("\n✅ 卸载完成!")
	return nil
}

// execCommand 执行系统命令
func execCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// installWindows 安装 Windows 自启动 (注册表)
func installWindows() error {
	// Windows 通过注册表实现自启动
	// HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	// 创建快捷方式到启动文件夹
	startupDir := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	shortcutPath := filepath.Join(startupDir, "FFmpeg Binary.lnk")

	// 这里需要使用 Windows API 或第三方库创建快捷方式
	// 简化版: 提示用户手动创建
	fmt.Printf("请将以下程序添加到启动项:\n%s\n", exePath)
	fmt.Printf("或手动创建快捷方式到: %s\n", shortcutPath)

	return nil
}

// uninstallWindows 卸载 Windows 自启动
func uninstallWindows() error {
	startupDir := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	shortcutPath := filepath.Join(startupDir, "FFmpeg Binary.lnk")

	return os.Remove(shortcutPath)
}

// installLinux 安装 Linux 自启动 (systemd)
func installLinux() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// systemd 用户服务目录
	systemdDir := filepath.Join(home, ".config", "systemd", "user")
	if err := os.MkdirAll(systemdDir, 0755); err != nil {
		return err
	}

	servicePath := filepath.Join(systemdDir, "ffmpeg-binary.service")

	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	serviceContent := fmt.Sprintf(`[Unit]
Description=FFmpeg Binary Service
After=network.target

[Service]
Type=simple
ExecStart=%s
Restart=always
RestartSec=10

[Install]
WantedBy=default.target
`, exePath)

	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return err
	}

	fmt.Println("systemd 服务文件已创建:", servicePath)
	fmt.Println("启用服务: systemctl --user enable ffmpeg-binary.service")
	fmt.Println("启动服务: systemctl --user start ffmpeg-binary.service")

	return nil
}

// uninstallLinux 卸载 Linux 自启动
func uninstallLinux() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	servicePath := filepath.Join(home, ".config", "systemd", "user", "ffmpeg-binary.service")

	fmt.Println("停止服务: systemctl --user stop ffmpeg-binary.service")
	fmt.Println("禁用服务: systemctl --user disable ffmpeg-binary.service")

	return os.Remove(servicePath)
}
