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
	TempDir    string `json:"temp_dir"`    // 临时文件目录
	OutputDir  string `json:"output_dir"`  // 输出文件目录
	FFmpegPath string `json:"ffmpeg_path"` // FFmpeg 可执行文件路径
}

// Load 加载配置
func Load() (*Config, error) {
	baseDir := getBaseDir()

	cfg := &Config{
		Port:      28888, // 固定端口 28888
		Host:      "127.0.0.1",
		DataDir:   filepath.Join(baseDir, "data"),
		TempDir:   filepath.Join(baseDir, "temp"),
		OutputDir: filepath.Join(baseDir, "output"),
	}

	// 尝试从配置文件加载
	configPath := getConfigPath()
	if data, err := os.ReadFile(configPath); err == nil {
		_ = json.Unmarshal(data, cfg)
	}

	// 确保所有目录存在
	dirs := []string{cfg.DataDir, cfg.TempDir, cfg.OutputDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	// 查找 FFmpeg
	cfg.FFmpegPath = findFFmpeg()

	return cfg, nil
}

// Save 保存配置到文件
func (c *Config) Save() error {
	configPath := getConfigPath()
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// getConfigPath 获取配置文件路径
func getConfigPath() string {
	baseDir := getBaseDir()
	os.MkdirAll(baseDir, 0755)
	return filepath.Join(baseDir, "config.json")
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
