package converter

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

// Converter FFmpeg 转换器
type Converter struct {
	ffmpegPath string
}

// New 创建转换器
func New(ffmpegPath string) *Converter {
	return &Converter{
		ffmpegPath: ffmpegPath,
	}
}

// ConvertStream 同步转换视频流 (WebM -> MP4)
func (c *Converter) ConvertStream(ctx context.Context, input io.Reader, output io.Writer) error {
	// FFmpeg 命令: 从 stdin 读取 WebM,输出 MP4 到 stdout
	cmd := exec.CommandContext(ctx, c.ffmpegPath,
		"-i", "pipe:0", // 从 stdin 读取
		"-c:v", "libx264", // 视频编码器
		"-c:a", "aac", // 音频编码器
		"-movflags", "frag_keyframe+empty_moov", // MP4 流式输出
		"-f", "mp4", // 输出格式
		"pipe:1", // 输出到 stdout
	)

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
	if err := cmd.Wait(); err != nil {
		// 返回详细的错误信息,包含 FFmpeg 的 stderr 输出
		return fmt.Errorf("FFmpeg 转换失败: %v\nFFmpeg 输出:\n%s", err, string(stderrBytes))
	}

	return nil
}

// ConvertFile 异步转换视频文件 (WebM -> MP4)
func (c *Converter) ConvertFile(ctx context.Context, inputPath, outputPath string, progress chan<- int) error {
	defer close(progress)

	// FFmpeg 命令: 文件转文件
	cmd := exec.CommandContext(ctx, c.ffmpegPath,
		"-i", inputPath,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-preset", "medium", // 编码速度/质量平衡
		"-crf", "23", // 质量控制
		"-y", // 覆盖输出文件
		outputPath,
	)

	// 获取视频时长用于计算进度(简化版,实际可通过 ffprobe 获取)
	// 这里只是示例,实际可以解析 stderr 输出来更新进度
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

	return cmd.Wait()
}

// Validate 验证 FFmpeg 是否可用
func (c *Converter) Validate() error {
	cmd := exec.Command(c.ffmpegPath, "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("FFmpeg 不可用: %v", err)
	}
	return nil
}
