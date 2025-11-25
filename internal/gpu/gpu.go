package gpu

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

// AccelerationType GPU 加速类型
type AccelerationType string

const (
	// 硬件加速类型
	AccelNone         AccelerationType = "none"         // 无硬件加速
	AccelNVIDIA       AccelerationType = "nvidia"       // NVIDIA CUDA/NVENC
	AccelAMD          AccelerationType = "amd"          // AMD VCE/AMF
	AccelIntel        AccelerationType = "intel"        // Intel Quick Sync
	AccelVideoToolbox AccelerationType = "videotoolbox" // macOS VideoToolbox
)

// Config GPU 配置
type Config struct {
	Enabled     bool             // 是否启用 GPU 加速
	AccelType   AccelerationType // 加速类型
	DecodeCodec string           // 解码器(如 h264_cuvid)
	EncodeCodec string           // 编码器(如 h264_nvenc)
	ExtraArgs   []string         // 额外的 FFmpeg 参数
	FallbackCPU bool             // 失败时回退到 CPU
}

// Detector GPU 检测器
type Detector struct {
	ffmpegPath string
}

// NewDetector 创建 GPU 检测器
func NewDetector(ffmpegPath string) *Detector {
	return &Detector{
		ffmpegPath: ffmpegPath,
	}
}

// DetectGPU 自动检测可用的 GPU 加速
func (d *Detector) DetectGPU() *Config {
	log.Println("🔍 开始检测 GPU 加速支持...")

	// 获取 FFmpeg 支持的编码器列表
	encoders, err := d.getEncoders()
	if err != nil {
		log.Printf("⚠️  无法获取编码器列表: %v", err)
		return &Config{Enabled: false, AccelType: AccelNone}
	}

	// 按优先级检测
	if runtime.GOOS == "darwin" {
		// macOS 优先使用 VideoToolbox
		if d.checkVideoToolbox(encoders) {
			return d.createVideoToolboxConfig()
		}
	}

	// NVIDIA GPU (跨平台)
	if d.checkNVIDIA(encoders) {
		return d.createNVIDIAConfig()
	}

	// AMD GPU
	if d.checkAMD(encoders) {
		return d.createAMDConfig()
	}

	// Intel Quick Sync
	if d.checkIntel(encoders) {
		return d.createIntelConfig()
	}

	log.Println("ℹ️  未检测到可用的 GPU 加速,将使用 CPU 编码")
	return &Config{Enabled: false, AccelType: AccelNone}
}

// getEncoders 获取 FFmpeg 支持的编码器列表
func (d *Detector) getEncoders() (string, error) {
	cmd := exec.Command(d.ffmpegPath, "-encoders")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// checkNVIDIA 检查 NVIDIA GPU 支持
func (d *Detector) checkNVIDIA(encoders string) bool {
	// 检查是否有 NVENC 编码器
	hasNvenc := strings.Contains(encoders, "h264_nvenc") ||
		strings.Contains(encoders, "hevc_nvenc")

	if !hasNvenc {
		return false
	}

	// 检查 nvidia-smi 是否可用(可选)
	if runtime.GOOS != "darwin" {
		cmd := exec.Command("nvidia-smi")
		if err := cmd.Run(); err == nil {
			log.Println("✅ 检测到 NVIDIA GPU")
			return true
		}
	}

	log.Println("⚠️  检测到 NVENC 编码器但无法确认 GPU 状态")
	return true
}

// checkAMD 检查 AMD GPU 支持
func (d *Detector) checkAMD(encoders string) bool {
	hasAMF := strings.Contains(encoders, "h264_amf") ||
		strings.Contains(encoders, "hevc_amf")

	if hasAMF {
		log.Println("✅ 检测到 AMD GPU (AMF)")
		return true
	}
	return false
}

// checkIntel 检查 Intel Quick Sync 支持
func (d *Detector) checkIntel(encoders string) bool {
	hasQSV := strings.Contains(encoders, "h264_qsv") ||
		strings.Contains(encoders, "hevc_qsv")

	if hasQSV {
		log.Println("✅ 检测到 Intel Quick Sync")
		return true
	}
	return false
}

// checkVideoToolbox 检查 macOS VideoToolbox 支持
func (d *Detector) checkVideoToolbox(encoders string) bool {
	hasVT := strings.Contains(encoders, "h264_videotoolbox") ||
		strings.Contains(encoders, "hevc_videotoolbox")

	if hasVT {
		log.Println("✅ 检测到 macOS VideoToolbox 硬件加速")
		return true
	}
	return false
}

// createNVIDIAConfig 创建 NVIDIA 配置
func (d *Detector) createNVIDIAConfig() *Config {
	log.Println("🎮 使用 NVIDIA GPU 加速 (NVENC)")
	return &Config{
		Enabled:     true,
		AccelType:   AccelNVIDIA,
		DecodeCodec: "h264_cuvid", // CUDA 解码
		EncodeCodec: "h264_nvenc", // NVENC 编码
		ExtraArgs: []string{
			"-hwaccel", "cuda",
			"-hwaccel_output_format", "cuda",
		},
		FallbackCPU: true,
	}
}

// createAMDConfig 创建 AMD 配置
func (d *Detector) createAMDConfig() *Config {
	log.Println("🎮 使用 AMD GPU 加速 (AMF)")
	return &Config{
		Enabled:     true,
		AccelType:   AccelAMD,
		DecodeCodec: "", // AMF 主要用于编码
		EncodeCodec: "h264_amf",
		ExtraArgs:   []string{},
		FallbackCPU: true,
	}
}

// createIntelConfig 创建 Intel 配置
func (d *Detector) createIntelConfig() *Config {
	log.Println("🎮 使用 Intel Quick Sync 加速")
	return &Config{
		Enabled:     true,
		AccelType:   AccelIntel,
		DecodeCodec: "h264_qsv",
		EncodeCodec: "h264_qsv",
		ExtraArgs: []string{
			"-hwaccel", "qsv",
		},
		FallbackCPU: true,
	}
}

// createVideoToolboxConfig 创建 VideoToolbox 配置
func (d *Detector) createVideoToolboxConfig() *Config {
	log.Println("🎮 使用 macOS VideoToolbox 硬件加速")
	return &Config{
		Enabled:     true,
		AccelType:   AccelVideoToolbox,
		DecodeCodec: "", // VideoToolbox 解码通过 -hwaccel 自动处理
		EncodeCodec: "h264_videotoolbox",
		ExtraArgs: []string{
			"-hwaccel", "videotoolbox",
			"-hwaccel_output_format", "videotoolbox_vld", // 保持硬件格式,避免 CPU-GPU 传输
		},
		FallbackCPU: true,
	}
}

// BuildFFmpegArgs 构建带 GPU 加速的 FFmpeg 参数
// inputArgs: -i 之前的参数, outputArgs: 输出相关的参数
func (cfg *Config) BuildFFmpegArgs(inputFile, outputFile string, outputArgs []string) []string {
	if !cfg.Enabled {
		// 无 GPU 加速,返回标准 CPU 编码参数
		args := []string{"-i", inputFile}
		args = append(args, outputArgs...)
		args = append(args, outputFile)
		return args
	}

	args := []string{}

	// 1. 添加硬件加速参数(在 -i 之前)
	args = append(args, cfg.ExtraArgs...)

	// 2. 如果有硬件解码器,添加解码参数
	if cfg.DecodeCodec != "" {
		args = append(args, "-c:v", cfg.DecodeCodec)
	}

	// 3. 输入文件
	args = append(args, "-i", inputFile)

	// 4. 添加输出参数,但替换编码器
	for i := 0; i < len(outputArgs); i++ {
		// 跳过 CPU 编码器,替换为 GPU 编码器
		if outputArgs[i] == "-c:v" && i+1 < len(outputArgs) {
			args = append(args, "-c:v", cfg.EncodeCodec)
			i++ // 跳过下一个参数(原来的编码器)
			continue
		}

		// 某些参数在 GPU 编码时不适用
		if outputArgs[i] == "-preset" && i+1 < len(outputArgs) {
			// GPU 编码器有自己的 preset
			switch cfg.AccelType {
			case AccelNVIDIA:
				args = append(args, "-preset", "p4") // NVENC preset (p1-p7)
			case AccelAMD:
				args = append(args, "-quality", "balanced") // AMF quality
			case AccelIntel:
				args = append(args, "-preset", "medium") // QSV preset
			case AccelVideoToolbox:
				// VideoToolbox 不需要 preset
			}
			i++ // 跳过原 preset 值
			continue
		}

		// CRF 在某些 GPU 编码器上需要转换为 qp
		if outputArgs[i] == "-crf" && i+1 < len(outputArgs) {
			switch cfg.AccelType {
			case AccelNVIDIA:
				args = append(args, "-cq", outputArgs[i+1]) // NVENC 使用 -cq
			case AccelAMD:
				args = append(args, "-rc", "cqp", "-qp", outputArgs[i+1]) // AMF 使用 qp
			case AccelVideoToolbox:
				args = append(args, "-q:v", outputArgs[i+1]) // VideoToolbox 使用 q:v
			default:
				args = append(args, outputArgs[i], outputArgs[i+1])
			}
			i++
			continue
		}

		args = append(args, outputArgs[i])
	}

	// 5. 输出文件
	args = append(args, outputFile)

	return args
}

// GetFallbackArgs 获取 CPU 回退参数(当 GPU 失败时)
func (cfg *Config) GetFallbackArgs(inputFile, outputFile string, outputArgs []string) []string {
	log.Println("⚠️  GPU 加速失败,回退到 CPU 编码...")

	args := []string{"-i", inputFile}

	// 使用标准 CPU 编码器
	for i := 0; i < len(outputArgs); i++ {
		if outputArgs[i] == "-c:v" && i+1 < len(outputArgs) {
			args = append(args, "-c:v", "libx264") // 回退到 libx264
			i++
			continue
		}
		args = append(args, outputArgs[i])
	}

	args = append(args, outputFile)
	return args
}

// Test 测试 GPU 配置是否可用
func (cfg *Config) Test(ffmpegPath string) error {
	if !cfg.Enabled {
		return nil
	}

	log.Printf("🧪 测试 %s GPU 加速...", cfg.AccelType)

	// 创建一个简单的测试命令
	args := []string{
		"-f", "lavfi",
		"-i", "testsrc=duration=1:size=320x240:rate=1",
	}

	// 添加硬件加速参数
	args = append(args, cfg.ExtraArgs...)

	// 添加编码器
	args = append(args, "-c:v", cfg.EncodeCodec)
	args = append(args, "-f", "null", "-")

	cmd := exec.Command(ffmpegPath, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GPU 测试失败: %v", err)
	}

	log.Println("✅ GPU 加速测试通过")
	return nil
}
