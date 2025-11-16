package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 应用配置
type Config struct {
	Port       int    `json:"port"`        // 服务端口
	Host       string `json:"host"`        // 服务地址
	DataDir    string `json:"data_dir"`    // 数据存储目录
	FFmpegPath string `json:"ffmpeg_path"` // FFmpeg 可执行文件路径
}

// Load 加载配置
func Load() (*Config, error) {
	cfg := &Config{
		Port:    28888, // 固定端口 28888
		Host:    "127.0.0.1",
		DataDir: getDefaultDataDir(),
	}

	// 尝试从配置文件加载
	configPath := getConfigPath()
	if data, err := os.ReadFile(configPath); err == nil {
		_ = json.Unmarshal(data, cfg)
	}

	// 确保数据目录存在
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, err
	}

	// 查找 FFmpeg
	cfg.FFmpegPath = findFFmpeg()

	return cfg, nil
}

// Save 保存配置
func (c *Config) Save() error {
	configPath := getConfigPath()
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// getDefaultDataDir 获取默认数据目录
func getDefaultDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ffmpeg-binary", "data")
}

// getConfigPath 获取配置文件路径
func getConfigPath() string {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".ffmpeg-binary")
	os.MkdirAll(configDir, 0755)
	return filepath.Join(configDir, "config.json")
}

// findFFmpeg 查找 FFmpeg 可执行文件
func findFFmpeg() string {
	// 优先使用打包的 FFmpeg
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	possiblePaths := []string{
		filepath.Join(exeDir, "ffmpeg"),
		filepath.Join(exeDir, "ffmpeg.exe"),
		filepath.Join(exeDir, "bin", "ffmpeg"),
		filepath.Join(exeDir, "bin", "ffmpeg.exe"),
		"ffmpeg", // 系统 PATH
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "ffmpeg" // 默认使用系统 PATH
}
