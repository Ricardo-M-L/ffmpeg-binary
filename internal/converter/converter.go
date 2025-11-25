package converter

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"

	"goalfy-mediaconverter/internal/gpu"
)

// Converter FFmpeg 转换器
type Converter struct {
	ffmpegPath string
	gpuConfig  *gpu.Config
}

// New 创建转换器
func New(ffmpegPath string) *Converter {
	// 自动检测 GPU 加速
	detector := gpu.NewDetector(ffmpegPath)
	gpuConfig := detector.DetectGPU()

	// 测试 GPU 配置
	if gpuConfig.Enabled {
		if err := gpuConfig.Test(ffmpegPath); err != nil {
			log.Printf("⚠️  GPU 测试失败: %v, 将使用 CPU 编码", err)
			gpuConfig.Enabled = false
		}
	}

	return &Converter{
		ffmpegPath: ffmpegPath,
		gpuConfig:  gpuConfig,
	}
}

// ConvertStream 同步转换视频流 (WebM -> MP4)
func (c *Converter) ConvertStream(ctx context.Context, input io.Reader, output io.Writer) error {
	var args []string

	if c.gpuConfig.Enabled {
		// GPU 加速模式
		log.Printf("🎮 使用 %s GPU 加速进行流转换", c.gpuConfig.AccelType)

		// 添加硬件加速参数
		args = append(args, c.gpuConfig.ExtraArgs...)

		// 如果有硬件解码器
		if c.gpuConfig.DecodeCodec != "" {
			args = append(args, "-c:v", c.gpuConfig.DecodeCodec)
		}

		// 输入
		args = append(args, "-i", "pipe:0")

		// GPU 编码器
		args = append(args, "-c:v", c.gpuConfig.EncodeCodec)
		args = append(args, "-c:a", "aac")

		// 根据 GPU 类型设置参数
		switch c.gpuConfig.AccelType {
		case gpu.AccelNVIDIA:
			args = append(args, "-preset", "p4", "-cq", "23")
		case gpu.AccelAMD:
			args = append(args, "-rc", "cqp", "-qp", "23")
		case gpu.AccelIntel:
			args = append(args, "-global_quality", "23")
		case gpu.AccelVideoToolbox:
			args = append(args,
				"-b:v", "0",
				"-q:v", "65",
				"-realtime", "1",
				"-allow_sw", "1",
			)
		}

		args = append(args, "-movflags", "frag_keyframe+empty_moov")
		args = append(args, "-f", "mp4", "pipe:1")
	} else {
		// CPU 模式
		log.Println("💻 使用 CPU 编码进行流转换")
		args = []string{
			"-i", "pipe:0",
			"-c:v", "libx264",
			"-c:a", "aac",
			"-movflags", "frag_keyframe+empty_moov",
			"-f", "mp4",
			"pipe:1",
		}
	}

	cmd := exec.CommandContext(ctx, c.ffmpegPath, args...)
	cmd.Stdin = input
	cmd.Stdout = output

	// 捕获 stderr 以便记录详细错误信息
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建 stderr 管道失败: %v", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 FFmpeg 失败: %v", err)
	}

	// 读取 stderr 输出
	stderrBytes, _ := io.ReadAll(stderrPipe)

	// 等待命令完成
	err = cmd.Wait()

	// GPU 失败时回退到 CPU
	if err != nil && c.gpuConfig.Enabled && c.gpuConfig.FallbackCPU {
		log.Printf("⚠️  GPU 编码失败: %v", err)
		log.Println("🔄 尝试使用 CPU 编码...")

		// 重新构建 CPU 命令
		cpuArgs := []string{
			"-i", "pipe:0",
			"-c:v", "libx264",
			"-c:a", "aac",
			"-movflags", "frag_keyframe+empty_moov",
			"-f", "mp4",
			"pipe:1",
		}

		cmd = exec.CommandContext(ctx, c.ffmpegPath, cpuArgs...)
		cmd.Stdin = input
		cmd.Stdout = output
		err = cmd.Run()
	}

	if err != nil {
		return fmt.Errorf("FFmpeg 转换失败: %v\nFFmpeg 输出:\n%s", err, string(stderrBytes))
	}

	return nil
}

// ConvertFile 异步转换视频文件 (WebM -> MP4)
func (c *Converter) ConvertFile(ctx context.Context, inputPath, outputPath string, progress chan<- int) error {
	defer close(progress)

	var args []string

	if c.gpuConfig.Enabled {
		// GPU 加速模式
		log.Printf("🎮 使用 %s GPU 加速进行文件转换", c.gpuConfig.AccelType)

		// 添加硬件加速参数
		args = append(args, c.gpuConfig.ExtraArgs...)

		// 如果有硬件解码器
		if c.gpuConfig.DecodeCodec != "" {
			args = append(args, "-c:v", c.gpuConfig.DecodeCodec)
		}

		// 输入
		args = append(args, "-i", inputPath)

		// GPU 编码器
		args = append(args, "-c:v", c.gpuConfig.EncodeCodec)
		args = append(args, "-c:a", "aac")

		// 根据 GPU 类型设置参数
		switch c.gpuConfig.AccelType {
		case gpu.AccelNVIDIA:
			args = append(args, "-preset", "p4", "-cq", "23")
		case gpu.AccelAMD:
			args = append(args, "-rc", "cqp", "-qp", "23")
		case gpu.AccelIntel:
			args = append(args, "-preset", "medium", "-global_quality", "23")
		case gpu.AccelVideoToolbox:
			args = append(args,
				"-b:v", "0",
				"-q:v", "65",
				"-realtime", "1",
				"-allow_sw", "1",
			)
		}

		args = append(args, "-y", outputPath)
	} else {
		// CPU 模式
		log.Println("💻 使用 CPU 编码进行文件转换")
		args = []string{
			"-i", inputPath,
			"-c:v", "libx264",
			"-c:a", "aac",
			"-preset", "medium",
			"-crf", "23",
			"-y",
			outputPath,
		}
	}

	cmd := exec.CommandContext(ctx, c.ffmpegPath, args...)
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return err
	}

	// 读取 stderr 输出(可以解析进度信息)
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if err != nil {
				break
			}
			_ = n // 这里可以解析 FFmpeg 输出来更新进度
			// 简化版:每次读取都报告进度增加
			select {
			case progress <- 10: // 简化的进度更新
			default:
			}
		}
	}()

	err := cmd.Wait()

	// GPU 失败时回退到 CPU
	if err != nil && c.gpuConfig.Enabled && c.gpuConfig.FallbackCPU {
		log.Printf("⚠️  GPU 编码失败: %v", err)
		log.Println("🔄 尝试使用 CPU 编码...")

		// CPU 回退
		cpuArgs := []string{
			"-i", inputPath,
			"-c:v", "libx264",
			"-c:a", "aac",
			"-preset", "medium",
			"-crf", "23",
			"-y",
			outputPath,
		}

		cmd = exec.CommandContext(ctx, c.ffmpegPath, cpuArgs...)
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			return err
		}

		// 读取 stderr 输出
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := stderr.Read(buf)
				if err != nil {
					break
				}
				_ = n
				select {
				case progress <- 10:
				default:
				}
			}
		}()

		err = cmd.Wait()
	}

	return err
}

// Validate 验证 FFmpeg 是否可用
func (c *Converter) Validate() error {
	cmd := exec.Command(c.ffmpegPath, "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("FFmpeg 不可用: %v", err)
	}
	return nil
}
