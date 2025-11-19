//go:build windows

package config

import (
	"os"
	"path/filepath"
)

// getBaseDir 获取基础目录 (Windows 版本)
// Windows 上优先使用安装目录,这样文件都在 C:\Program Files\GoalfyMediaConverter\ 下
// 用户可以方便地找到转换后的视频文件
func getBaseDir() string {
	// 获取程序所在目录
	exePath, err := os.Executable()
	if err != nil {
		// 获取失败,回退到用户主目录
		return getUserHomeDir()
	}

	exeDir := filepath.Dir(exePath)

	// 测试是否有写权限
	// 在安装目录尝试创建测试文件
	if hasWritePermission(exeDir) {
		// 有写权限,使用安装目录
		// 例如: C:\Program Files\GoalfyMediaConverter\
		return exeDir
	}

	// 没有写权限(例如 Program Files 需要管理员权限)
	// 回退到用户主目录
	return getUserHomeDir()
}

// hasWritePermission 检测目录是否有写权限
func hasWritePermission(dir string) bool {
	testFile := filepath.Join(dir, ".writable_test")
	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(testFile)
	return true
}

// getUserHomeDir 获取用户主目录下的配置目录
func getUserHomeDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".goalfy-mediaconverter")
}
