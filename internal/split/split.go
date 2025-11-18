package split

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// TimeInterval 时间区间
type TimeInterval struct {
	Start float64 `json:"start"` // 开始时间(秒)
	End   float64 `json:"end"`   // 结束时间(秒)
}

// SplitRequest 切割请求
type SplitRequest struct {
	TaskID          string         `json:"taskId" binding:"required"`          // 任务ID
	DeleteIntervals []TimeInterval `json:"deleteIntervals" binding:"required"` // 要删除的时间区间
	VideoDuration   float64        `json:"videoDuration" binding:"required"`   // 视频总时长(秒)
	InputPath       string         `json:"inputPath"`                          // 输入文件路径(由服务端设置,不从JSON接收)
}

// SegmentResult 片段结果
type SegmentResult struct {
	Success       bool    `json:"success"`
	OutputPath    string  `json:"outputPath"`
	Size          int64   `json:"size"`
	Duration      float64 `json:"duration"`
	StartTime     float64 `json:"startTime"`
	EndTime       float64 `json:"endTime"`
	SegmentIndex  int     `json:"segmentIndex"`
	FileName      string  `json:"fileName"`
	OriginalStart float64 `json:"originalStart"`
	OriginalEnd   float64 `json:"originalEnd"`
}

// SplitResponse 切割响应
type SplitResponse struct {
	Success       bool            `json:"success"`
	TaskID        string          `json:"taskId,omitempty"`
	TotalSegments int             `json:"totalSegments,omitempty"`
	Segments      []SegmentResult `json:"segments,omitempty"`
	Error         string          `json:"error,omitempty"`
}

// Splitter 视频切割器
type Splitter struct {
	ffmpegPath string
	outputDir  string
}

// New 创建切割器
func New(ffmpegPath, outputDir string) *Splitter {
	return &Splitter{
		ffmpegPath: ffmpegPath,
		outputDir:  outputDir,
	}
}

// calculateRetainedSegments 计算保留的视频片段
func calculateRetainedSegments(videoDuration float64, deleteIntervals []TimeInterval) []TimeInterval {
	// 如果没有删除区间,保留整个视频
	if len(deleteIntervals) == 0 {
		return []TimeInterval{{Start: 0, End: videoDuration}}
	}

	// 按开始时间排序删除区间
	sortedIntervals := make([]TimeInterval, len(deleteIntervals))
	copy(sortedIntervals, deleteIntervals)
	sort.Slice(sortedIntervals, func(i, j int) bool {
		return sortedIntervals[i].Start < sortedIntervals[j].Start
	})

	// 计算保留的片段
	retained := []TimeInterval{}
	currentTime := 0.0

	for _, interval := range sortedIntervals {
		// 添加删除区间之前的保留片段
		if currentTime < interval.Start {
			retained = append(retained, TimeInterval{
				Start: currentTime,
				End:   interval.Start,
			})
		}
		currentTime = interval.End
	}

	// 添加最后一个保留片段
	if currentTime < videoDuration {
		retained = append(retained, TimeInterval{
			Start: currentTime,
			End:   videoDuration,
		})
	}

	// 过滤无效片段
	validRetained := []TimeInterval{}
	for _, segment := range retained {
		if segment.End > segment.Start {
			validRetained = append(validRetained, segment)
		}
	}

	return validRetained
}

// splitSegment 切割单个视频片段
func (s *Splitter) splitSegment(inputPath, outputPath string, startTime, duration float64) error {
	// 构建FFmpeg命令
	// 使用重新编码以获得精确的切割(不使用 -c copy)
	// ffmpeg -ss 开始时间 -i input.mp4 -t 时长 -c:v libx264 -c:a aac output.mp4
	args := []string{
		"-ss", fmt.Sprintf("%.3f", startTime), // -ss 放在 -i 之前,更快速定位
		"-i", inputPath,
		"-t", fmt.Sprintf("%.3f", duration),
		"-c:v", "libx264", // 重新编码视频以获得精确切割
		"-c:a", "aac", // 重新编码音频
		"-preset", "ultrafast", // 使用最快编码速度
		"-crf", "23", // 质量控制(18-28,越小质量越好)
		"-f", "mp4",
		"-movflags", "+faststart", // 优化流媒体播放
		"-y", // 覆盖输出文件
		outputPath,
	}

	log.Printf("🎬 FFmpeg 命令: %s %s", s.ffmpegPath, strings.Join(args, " "))

	cmd := exec.Command(s.ffmpegPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("FFmpeg 执行失败: %v", err)
	}

	return nil
}

// SplitVideo 执行视频切割任务
func (s *Splitter) SplitVideo(req SplitRequest) (*SplitResponse, error) {
	log.Printf("📹 开始视频切割任务: %s", req.TaskID)

	// 1. 使用传入的文件路径
	inputPath := req.InputPath
	if inputPath == "" {
		return &SplitResponse{
			Success: false,
			Error:   "未提供输入文件路径",
		}, nil
	}

	// 检查文件是否存在
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return &SplitResponse{
			Success: false,
			Error:   fmt.Sprintf("视频文件不存在: %s", inputPath),
		}, nil
	}

	log.Printf("✅ 找到源文件: %s", inputPath)

	// 2. 计算保留片段
	retainedSegments := calculateRetainedSegments(req.VideoDuration, req.DeleteIntervals)
	if len(retainedSegments) == 0 {
		return &SplitResponse{
			Success: false,
			Error:   "没有要保留的视频片段",
		}, nil
	}

	log.Printf("📊 计算出 %d 个保留片段", len(retainedSegments))

	// 3. 切割每个片段
	segments := []SegmentResult{}
	// 使用任务ID作为基础文件名
	baseFileName := req.TaskID

	for i, segment := range retainedSegments {
		segmentIndex := i + 1
		duration := segment.End - segment.Start

		// 输出文件名: taskId_part1.mp4, taskId_part2.mp4, ...
		outputFileName := fmt.Sprintf("%s_part%d.mp4", baseFileName, segmentIndex)
		outputPath := filepath.Join(s.outputDir, outputFileName)

		log.Printf("🔪 切割片段 %d/%d: %.2fs - %.2fs (时长: %.2fs)",
			segmentIndex, len(retainedSegments), segment.Start, segment.End, duration)

		// 执行切割
		err := s.splitSegment(inputPath, outputPath, segment.Start, duration)
		if err != nil {
			log.Printf("❌ 片段 %d 切割失败: %v", segmentIndex, err)
			segments = append(segments, SegmentResult{
				Success:      false,
				SegmentIndex: segmentIndex,
			})
			continue
		}

		// 获取文件信息
		fileInfo, err := os.Stat(outputPath)
		var fileSize int64 = 0
		if err == nil {
			fileSize = fileInfo.Size()
		}

		log.Printf("✅ 片段 %d 切割成功: %s (%.2f MB)",
			segmentIndex, outputFileName, float64(fileSize)/(1024*1024))

		segments = append(segments, SegmentResult{
			Success:       true,
			OutputPath:    outputPath,
			Size:          fileSize,
			Duration:      duration,
			StartTime:     segment.Start,
			EndTime:       segment.End,
			SegmentIndex:  segmentIndex,
			FileName:      outputFileName,
			OriginalStart: segment.Start,
			OriginalEnd:   segment.End,
		})
	}

	// 4. 删除原始完整文件(节省空间)
	if _, err := os.Stat(inputPath); err == nil {
		if err := os.Remove(inputPath); err != nil {
			log.Printf("⚠️  删除原始文件失败: %v", err)
		} else {
			log.Printf("🗑️  已删除原始完整MP4文件")
		}
	}

	log.Printf("🎉 视频切割任务完成: %d 个片段", len(segments))

	return &SplitResponse{
		Success:       true,
		TaskID:        req.TaskID,
		TotalSegments: len(segments),
		Segments:      segments,
	}, nil
}

// FindSegmentFile 查找片段文件
func (s *Splitter) FindSegmentFile(taskID string, segmentIndex int) (string, error) {
	files, err := os.ReadDir(s.outputDir)
	if err != nil {
		return "", fmt.Errorf("读取输出目录失败: %v", err)
	}

	targetPattern := fmt.Sprintf("_part%d.mp4", segmentIndex)
	for _, file := range files {
		if strings.Contains(file.Name(), taskID) && strings.Contains(file.Name(), targetPattern) {
			return filepath.Join(s.outputDir, file.Name()), nil
		}
	}

	return "", fmt.Errorf("未找到片段文件")
}

// CleanupSplitFiles 清理切割文件
func (s *Splitter) CleanupSplitFiles(taskID string) (int, error) {
	files, err := os.ReadDir(s.outputDir)
	if err != nil {
		return 0, fmt.Errorf("读取输出目录失败: %v", err)
	}

	count := 0
	for _, file := range files {
		if strings.Contains(file.Name(), taskID) && strings.Contains(file.Name(), "_part") && strings.HasSuffix(file.Name(), ".mp4") {
			filePath := filepath.Join(s.outputDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				log.Printf("⚠️  删除文件失败 %s: %v", file.Name(), err)
			} else {
				count++
				log.Printf("🗑️  已删除: %s", file.Name())
			}
		}
	}

	return count, nil
}
