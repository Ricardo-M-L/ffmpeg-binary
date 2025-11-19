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
// 注意: PKG安装时不应该调用此方法,应该直接调用 FindFFmpeg()
// 因为 PKG 的 postinstall 脚本会以 root 权限安装 FFmpeg
func (i *FFmpegInstaller) CheckAndInstall() (string, error) {
	// 1. 先尝试查找已安装的 FFmpeg
	ffmpegPath, err := i.FindFFmpeg()
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
	ffmpegPath, err = i.FindFFmpeg()
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

// FindFFmpeg 查找 FFmpeg 可执行文件 (公开方法,用于只查找不安装)
func (i *FFmpegInstaller) FindFFmpeg() (string, error) {
	return i.findFFmpeg()
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
		// 获取当前执行文件所在目录
		exePath, err := os.Executable()
		if err == nil {
			exeDir := strings.TrimSuffix(exePath, "\\ffmpeg-binary.exe")
			exeDir = strings.TrimSuffix(exeDir, "/ffmpeg-binary.exe")
			// 优先查找安装目录的 bin 文件夹
			paths = append(paths,
				exeDir+"\\bin\\ffmpeg.exe",
				exeDir+"/bin/ffmpeg.exe",
			)
		}
		// 添加常见的系统路径
		paths = append(paths,
			`C:\Program Files\ffmpeg\bin\ffmpeg.exe`,
			`C:\Program Files\GoalfyMediaConverter\bin\ffmpeg.exe`,
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
	fmt.Println("📦 正在下载 FFmpeg 静态编译版本...")

	// 使用 evermeet.cx 提供的 FFmpeg 静态编译版本 (ZIP格式)
	ffmpegURL := "https://evermeet.cx/ffmpeg/getrelease/zip"

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "ffmpeg-install-*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	zipPath := tmpDir + "/ffmpeg.zip"

	// 下载 FFmpeg ZIP 文件
	fmt.Println("下载 FFmpeg...")
	cmd := exec.Command("curl", "-L", "-o", zipPath, ffmpegURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("下载 FFmpeg 失败: %v", err)
	}

	// 解压 ZIP 文件
	fmt.Println("解压 FFmpeg...")
	cmd = exec.Command("unzip", "-q", zipPath, "-d", tmpDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("解压 FFmpeg 失败: %v", err)
	}

	// 安装到 /usr/local/bin
	ffmpegBinary := tmpDir + "/ffmpeg"
	installDir := "/usr/local/bin"

	if _, err := os.Stat(ffmpegBinary); os.IsNotExist(err) {
		return fmt.Errorf("解压后未找到 ffmpeg 二进制文件")
	}

	fmt.Printf("安装 FFmpeg 到 %s...\n", installDir)

	// 确保安装目录存在
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("创建安装目录失败: %v", err)
	}

	// 检查是否已经是 root 权限 (UID == 0)
	isRoot := os.Geteuid() == 0

	targetPath := installDir + "/ffmpeg"

	// 复制到系统目录
	if isRoot {
		// 已经是 root 权限,直接复制
		cmd = exec.Command("cp", "-f", ffmpegBinary, targetPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("复制 FFmpeg 失败: %v", err)
		}

		// 设置执行权限
		if err := os.Chmod(targetPath, 0755); err != nil {
			return fmt.Errorf("设置权限失败: %v", err)
		}
	} else {
		// 非 root 权限,尝试直接复制
		cmd = exec.Command("cp", "-f", ffmpegBinary, targetPath)
		if err := cmd.Run(); err != nil {
			// 如果复制失败,尝试使用 sudo
			fmt.Println("需要管理员权限安装 FFmpeg...")
			cmd = exec.Command("sudo", "cp", "-f", ffmpegBinary, targetPath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("安装 FFmpeg 失败: %v", err)
			}

			// 使用 sudo 设置权限
			cmd = exec.Command("sudo", "chmod", "+x", targetPath)
			cmd.Run() // 忽略错误
		} else {
			// 直接复制成功,设置权限
			os.Chmod(targetPath, 0755) // 忽略错误
		}
	}

	fmt.Println("✅ FFmpeg 安装成功")
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
