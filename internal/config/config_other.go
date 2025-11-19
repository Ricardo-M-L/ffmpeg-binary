//go:build !windows

package config

import (
	"os"
	"path/filepath"
)

// getBaseDir 获取基础目录 (macOS/Linux 版本)
// macOS 和 Linux 上始终使用用户主目录
// 因为应用程序通常安装在系统目录(如 /Applications),没有写权限
func getBaseDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".goalfy-mediaconverter")
}
