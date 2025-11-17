package installer

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// FFmpegInstaller FFmpeg 安装器
type FFmpegInstaller struct{}

// NewFFmpegInstaller 创建安装器
func NewFFmpegInstaller() *FFmpegInstaller {
	return &FFmpegInstaller{}
}

// CheckAndInstall 检查 FFmpeg 是否存在,不存在则自动安装
func (i *FFmpegInstaller) CheckAndInstall() (string, error) {
	// 1. 先尝试查找已安装的 FFmpeg
	ffmpegPath, err := i.findFFmpeg()
	if err == nil && ffmpegPath != "" {
		// FFmpeg 已存在,验证是否可用
		if i.validateFFmpeg(ffmpegPath) {
			fmt.Printf("✅ FFmpeg 已安装: %s\n", ffmpegPath)
			return ffmpegPath, nil
		}
	}

	// 2. FFmpeg 不存在或不可用,自动安装
	fmt.Println("⚠️  FFmpeg 未安装或不可用,正在自动安装...")

	if err := i.installFFmpeg(); err != nil {
		return "", fmt.Errorf("安装 FFmpeg 失败: %v", err)
	}

	// 3. 重新查找安装后的 FFmpeg
	ffmpegPath, err = i.findFFmpeg()
	if err != nil {
		return "", fmt.Errorf("安装后未找到 FFmpeg: %v", err)
	}

	// 4. 验证安装是否成功
	if !i.validateFFmpeg(ffmpegPath) {
		return "", fmt.Errorf("FFmpeg 安装后验证失败")
	}

	fmt.Printf("✅ FFmpeg 安装成功: %s\n", ffmpegPath)
	return ffmpegPath, nil
}

// findFFmpeg 查找 FFmpeg 可执行文件
func (i *FFmpegInstaller) findFFmpeg() (string, error) {
	// 尝试常见路径
	paths := []string{
		"ffmpeg", // 在 PATH 中查找
	}

	// 根据操作系统添加额外路径
	switch runtime.GOOS {
	case "darwin":
		paths = append(paths,
			"/opt/homebrew/bin/ffmpeg", // Apple Silicon Homebrew
			"/usr/local/bin/ffmpeg",    // Intel Homebrew
		)
	case "linux":
		paths = append(paths,
			"/usr/bin/ffmpeg",
			"/usr/local/bin/ffmpeg",
		)
	case "windows":
		paths = append(paths,
			`C:\Program Files\ffmpeg\bin\ffmpeg.exe`,
			`C:\ffmpeg\bin\ffmpeg.exe`,
		)
	}

	// 尝试每个路径
	for _, path := range paths {
		if resolvedPath, err := exec.LookPath(path); err == nil {
			return resolvedPath, nil
		}
	}

	return "", fmt.Errorf("未找到 FFmpeg")
}

// validateFFmpeg 验证 FFmpeg 是否可用
func (i *FFmpegInstaller) validateFFmpeg(path string) bool {
	cmd := exec.Command(path, "-version")
	return cmd.Run() == nil
}

// installFFmpeg 根据操作系统自动安装 FFmpeg
func (i *FFmpegInstaller) installFFmpeg() error {
	switch runtime.GOOS {
	case "darwin":
		return i.installOnMacOS()
	case "linux":
		return i.installOnLinux()
	case "windows":
		return i.installOnWindows()
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// installOnMacOS 在 macOS 上安装 FFmpeg
func (i *FFmpegInstaller) installOnMacOS() error {
	// 检查 Homebrew 是否安装
	if !i.isHomebrewInstalled() {
		fmt.Println("⚠️  Homebrew 未安装,正在安装 Homebrew...")
		if err := i.installHomebrew(); err != nil {
			return fmt.Errorf("安装 Homebrew 失败: %v", err)
		}
	}

	fmt.Println("📦 正在通过 Homebrew 安装 FFmpeg...")
	cmd := exec.Command("brew", "install", "ffmpeg")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("brew install ffmpeg 失败: %v", err)
	}

	return nil
}

// isHomebrewInstalled 检查 Homebrew 是否已安装
func (i *FFmpegInstaller) isHomebrewInstalled() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

// installHomebrew 安装 Homebrew
func (i *FFmpegInstaller) installHomebrew() error {
	// Homebrew 官方安装脚本
	installScript := `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`

	cmd := exec.Command("bash", "-c", installScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin // 需要用户交互

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行 Homebrew 安装脚本失败: %v", err)
	}

	return nil
}

// installOnLinux 在 Linux 上安装 FFmpeg
func (i *FFmpegInstaller) installOnLinux() error {
	// 检测 Linux 发行版
	distro := i.detectLinuxDistro()

	fmt.Printf("📦 正在在 %s 上安装 FFmpeg...\n", distro)

	var cmd *exec.Cmd
	switch distro {
	case "ubuntu", "debian":
		cmd = exec.Command("sudo", "apt-get", "update")
		cmd.Run() // 忽略更新错误
		cmd = exec.Command("sudo", "apt-get", "install", "-y", "ffmpeg")
	case "fedora", "rhel", "centos":
		cmd = exec.Command("sudo", "dnf", "install", "-y", "ffmpeg")
	case "arch":
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", "ffmpeg")
	default:
		return fmt.Errorf("不支持的 Linux 发行版: %s", distro)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("安装 FFmpeg 失败: %v", err)
	}

	return nil
}

// detectLinuxDistro 检测 Linux 发行版
func (i *FFmpegInstaller) detectLinuxDistro() string {
	// 读取 /etc/os-release
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "unknown"
	}

	content := string(data)
	if strings.Contains(content, "Ubuntu") {
		return "ubuntu"
	} else if strings.Contains(content, "Debian") {
		return "debian"
	} else if strings.Contains(content, "Fedora") {
		return "fedora"
	} else if strings.Contains(content, "CentOS") || strings.Contains(content, "Red Hat") {
		return "rhel"
	} else if strings.Contains(content, "Arch") {
		return "arch"
	}

	return "unknown"
}

// installOnWindows 在 Windows 上安装 FFmpeg
func (i *FFmpegInstaller) installOnWindows() error {
	// Windows 上通过 Chocolatey 安装
	if !i.isChocolateyInstalled() {
		return fmt.Errorf("请先安装 Chocolatey: https://chocolatey.org/install\n然后运行: choco install ffmpeg")
	}

	fmt.Println("📦 正在通过 Chocolatey 安装 FFmpeg...")
	cmd := exec.Command("choco", "install", "ffmpeg", "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("choco install ffmpeg 失败: %v", err)
	}

	return nil
}

// isChocolateyInstalled 检查 Chocolatey 是否已安装
func (i *FFmpegInstaller) isChocolateyInstalled() bool {
	_, err := exec.LookPath("choco")
	return err == nil
}
