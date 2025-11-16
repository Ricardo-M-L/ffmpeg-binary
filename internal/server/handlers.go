package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"ffmpeg-binary/internal/task"

	"github.com/gin-gonic/gin"
)

// handleSyncConvert 同步转换接口
// POST /api/v1/convert/sync
// 直接接收 WebM 流,返回 MP4 流
func (s *Server) handleSyncConvert(c *gin.Context) {
	// 检查请求体是否为空
	if c.Request.ContentLength == 0 {
		log.Printf("同步转换失败: 请求体为空")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体不能为空,请上传 WebM 视频文件"})
		return
	}

	log.Printf("开始同步转换,内容长度: %d 字节", c.Request.ContentLength)

	// 设置响应头(在确认有数据后再设置)
	c.Header("Content-Type", "video/mp4")
	c.Header("Transfer-Encoding", "chunked")

	// 从请求体读取 WebM 流
	// 直接将转换后的 MP4 流写入响应
	if err := s.converter.ConvertStream(c.Request.Context(), c.Request.Body, c.Writer); err != nil {
		log.Printf("同步转换失败: %v", err)
		// 注意:此时响应头已发送,无法返回 JSON 错误
		// 只能记录日志,客户端会收到不完整的响应
		return
	}

	log.Printf("同步转换完成")
}

// handleAsyncConvert 创建异步转换任务
// POST /api/v1/convert/async
// 返回任务 ID,准备接收分片上传
func (s *Server) handleAsyncConvert(c *gin.Context) {
	// 创建临时文件存储上传的 WebM
	inputPath := filepath.Join(s.config.DataDir, fmt.Sprintf("input_%d.webm", os.Getpid()))
	outputPath := filepath.Join(s.config.DataDir, fmt.Sprintf("output_%d.mp4", os.Getpid()))

	// 创建任务
	t := s.taskMgr.Create(inputPath, outputPath)

	// 创建输入文件
	if _, err := os.Create(inputPath); err != nil {
		s.taskMgr.UpdateError(t.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":      t.ID,
		"status":       t.Status,
		"message":      "任务已创建,请上传视频分片",
		"upload_url":   fmt.Sprintf("/api/v1/convert/async/%s/chunk", t.ID),
		"status_url":   fmt.Sprintf("/api/v1/task/%s", t.ID),
		"download_url": fmt.Sprintf("/api/v1/task/%s/download", t.ID),
	})
}

// handleUploadChunk 上传视频分片
// POST /api/v1/convert/async/:task_id/chunk
func (s *Server) handleUploadChunk(c *gin.Context) {
	taskID := c.Param("task_id")

	t, err := s.taskMgr.Get(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 打开输入文件(追加模式)
	file, err := os.OpenFile(t.InputPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	// 写入分片数据
	written, err := io.Copy(file, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 检查是否是最后一个分片
	isLast := c.GetHeader("X-Last-Chunk") == "true"

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"written": written,
		"is_last": isLast,
	})

	// 如果是最后一个分片,开始转换
	if isLast {
		go s.processTask(t)
	}
}

// processTask 处理转换任务
func (s *Server) processTask(t *task.Task) {
	// 更新状态为处理中
	s.taskMgr.UpdateStatus(t.ID, task.StatusProcessing, 0)

	// 进度通道
	progress := make(chan int, 10)

	// 启动转换
	go func() {
		err := s.converter.ConvertFile(t.Context(), t.InputPath, t.OutputPath, progress)
		if err != nil {
			log.Printf("任务 %s 转换失败: %v", t.ID, err)
			s.taskMgr.UpdateError(t.ID, err)
			return
		}

		// 转换完成
		s.taskMgr.UpdateStatus(t.ID, task.StatusCompleted, 100)
		log.Printf("任务 %s 转换完成", t.ID)

		// 删除输入文件
		os.Remove(t.InputPath)
	}()

	// 更新进度
	currentProgress := 0
	for p := range progress {
		currentProgress += p
		if currentProgress > 100 {
			currentProgress = 100
		}
		s.taskMgr.UpdateStatus(t.ID, task.StatusProcessing, currentProgress)
	}
}

// handleGetTask 查询任务状态
// GET /api/v1/task/:task_id
func (s *Server) handleGetTask(c *gin.Context) {
	taskID := c.Param("task_id")

	t, err := s.taskMgr.Get(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         t.ID,
		"status":     t.Status,
		"progress":   t.Progress,
		"error":      t.Error,
		"created_at": t.CreatedAt,
		"updated_at": t.UpdatedAt,
	})
}

// handleDownloadVideo 下载转换后的视频
// GET /api/v1/task/:task_id/download
func (s *Server) handleDownloadVideo(c *gin.Context) {
	taskID := c.Param("task_id")

	t, err := s.taskMgr.Get(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 检查任务状态
	if t.Status != task.StatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    "任务未完成",
			"status":   t.Status,
			"progress": t.Progress,
		})
		return
	}

	// 检查输出文件是否存在
	if _, err := os.Stat(t.OutputPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "视频文件不存在"})
		return
	}

	// 返回文件
	c.Header("Content-Type", "video/mp4")
	c.File(t.OutputPath)
}

// handleDeleteTask 删除任务
// DELETE /api/v1/task/:task_id
func (s *Server) handleDeleteTask(c *gin.Context) {
	taskID := c.Param("task_id")

	t, err := s.taskMgr.Get(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 删除文件
	os.Remove(t.InputPath)
	os.Remove(t.OutputPath)

	// 删除任务
	s.taskMgr.Delete(taskID)

	c.JSON(http.StatusOK, gin.H{"message": "任务已删除"})
}

// handleListTasks 列出所有任务
// GET /api/v1/tasks
func (s *Server) handleListTasks(c *gin.Context) {
	tasks := s.taskMgr.List()

	result := make([]gin.H, len(tasks))
	for i, t := range tasks {
		result[i] = gin.H{
			"id":         t.ID,
			"status":     t.Status,
			"progress":   t.Progress,
			"created_at": t.CreatedAt,
			"updated_at": t.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": result,
		"total": len(tasks),
	})
}
