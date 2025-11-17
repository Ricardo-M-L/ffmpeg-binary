package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"goalfy-mediaconverter/internal/task"
	"goalfy-mediaconverter/internal/upload"

	"github.com/gin-gonic/gin"
)

// ==================== 上传模块 ====================

// handleUploadInit 初始化上传任务
// POST /api/upload/init
func (s *Server) handleUploadInit(c *gin.Context) {
	var req struct {
		FileName    string `json:"fileName" binding:"required"`
		FileSize    int64  `json:"fileSize" binding:"required"`
		TotalChunks int    `json:"totalChunks" binding:"required"`
		ChunkSize   int64  `json:"chunkSize"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "缺少必要参数: fileName, fileSize, totalChunks",
		})
		return
	}

	uploadTask, err := s.uploadMgr.CreateUploadTask(req.FileName, req.FileSize, req.TotalChunks, req.ChunkSize)
	if err != nil {
		log.Printf("创建上传任务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "创建上传任务失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "上传任务初始化成功",
		"data": gin.H{
			"uploadId":    uploadTask.UploadID,
			"fileName":    uploadTask.FileName,
			"totalChunks": uploadTask.TotalChunks,
		},
	})
}

// handleUploadChunk 上传文件切片
// POST /api/upload/chunk
func (s *Server) handleUploadChunk(c *gin.Context) {
	// 获取表单参数
	uploadID := c.PostForm("uploadId")
	chunkIndexStr := c.PostForm("chunkIndex")

	if uploadID == "" || chunkIndexStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "缺少必要参数: uploadId, chunkIndex, file",
		})
		return
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "chunkIndex 必须是数字",
		})
		return
	}

	// 获取上传任务
	uploadTask, err := s.uploadMgr.GetUploadTask(uploadID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 获取文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "缺少文件",
		})
		return
	}

	// 保存切片
	chunkPath := uploadTask.GetChunkPath(chunkIndex)
	if err := c.SaveUploadedFile(file, chunkPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "保存切片失败",
			"error":   err.Error(),
		})
		return
	}

	// 记录切片上传
	if err := s.uploadMgr.RecordChunk(uploadID, chunkIndex); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "记录切片失败",
			"error":   err.Error(),
		})
		return
	}

	// 刷新任务状态
	uploadTask, _ = s.uploadMgr.GetUploadTask(uploadID)
	isComplete := uploadTask.IsComplete()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "切片上传成功",
		"data": gin.H{
			"uploadId":       uploadID,
			"chunkIndex":     chunkIndex,
			"uploadedChunks": uploadTask.UploadedChunks,
			"totalChunks":    uploadTask.TotalChunks,
			"isComplete":     isComplete,
		},
	})

	// 如果所有切片都已上传,开始合并
	if isComplete {
		log.Printf("所有切片上传完成,开始合并文件: %s", uploadID)
		go func() {
			if err := s.uploadMgr.MergeChunks(uploadID); err != nil {
				log.Printf("合并切片失败: %v", err)
			} else {
				log.Printf("文件合并完成: %s", uploadID)
			}
		}()
	}
}

// handleUploadStatus 查询上传状态
// GET /api/upload/status/:uploadId
func (s *Server) handleUploadStatus(c *gin.Context) {
	uploadID := c.Param("uploadId")

	uploadTask, err := s.uploadMgr.GetUploadTask(uploadID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "上传任务不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    uploadTask,
	})
}

// handleUploadCancel 取消上传任务
// POST /api/upload/cancel/:uploadId
func (s *Server) handleUploadCancel(c *gin.Context) {
	uploadID := c.Param("uploadId")

	if err := s.uploadMgr.CancelUpload(uploadID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "取消上传任务失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "上传任务已取消",
	})
}

// ==================== 转换模块 ====================

// handleConvertStart 开始视频转换任务
// POST /api/convert/start
func (s *Server) handleConvertStart(c *gin.Context) {
	var req struct {
		UploadID     string                 `json:"uploadId"`
		FilePath     string                 `json:"filePath"`
		OutputFormat string                 `json:"outputFormat"`
		Quality      string                 `json:"quality"`
		Options      map[string]interface{} `json:"options"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	// 验证输入源
	if req.UploadID == "" && req.FilePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "必须提供uploadId或filePath",
		})
		return
	}

	// 获取输入文件路径
	var inputPath string
	if req.UploadID != "" {
		uploadTask, err := s.uploadMgr.GetUploadTask(req.UploadID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "上传任务不存在",
			})
			return
		}

		if uploadTask.Status != upload.UploadStatusMerged {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": fmt.Sprintf("文件尚未合并完成,当前状态: %s", uploadTask.Status),
			})
			return
		}

		inputPath = uploadTask.MergedPath
	} else {
		inputPath = req.FilePath
	}

	// 检查输入文件是否存在
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "输入文件不存在",
		})
		return
	}

	// 设置默认值
	if req.OutputFormat == "" {
		req.OutputFormat = "mp4"
	}
	if req.Quality == "" {
		req.Quality = "medium"
	}

	// 生成输出文件路径
	outputPath := filepath.Join(s.config.OutputDir, fmt.Sprintf("%s.%s", generateTaskID(), req.OutputFormat))

	// 创建转换任务
	convertTask := s.taskMgr.CreateWithOptions(inputPath, outputPath, req.OutputFormat, req.Quality, req.UploadID)

	// 异步执行转换
	go s.processConvertTask(convertTask)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "转换任务已启动",
		"data": gin.H{
			"taskId":       convertTask.ID,
			"inputPath":    inputPath,
			"outputFormat": req.OutputFormat,
			"quality":      req.Quality,
		},
	})
}

// handleConvertStatus 查询转换状态
// GET /api/convert/status/:taskId
func (s *Server) handleConvertStatus(c *gin.Context) {
	taskID := c.Param("taskId")

	convertTask, err := s.taskMgr.Get(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "转换任务不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    convertTask,
	})
}

// handleConvertCancel 取消转换任务
// POST /api/convert/cancel/:taskId
func (s *Server) handleConvertCancel(c *gin.Context) {
	taskID := c.Param("taskId")

	if err := s.taskMgr.Delete(taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "取消转换任务失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "转换任务已取消",
	})
}

// handleConvertList 获取转换任务列表
// GET /api/convert/list
func (s *Server) handleConvertList(c *gin.Context) {
	statusFilter := c.Query("status")
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	tasks := s.taskMgr.List()

	// 过滤
	var filtered []*task.Task
	for _, t := range tasks {
		if statusFilter == "" || string(t.Status) == statusFilter {
			filtered = append(filtered, t)
		}
	}

	// 限制数量
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tasks": filtered,
			"total": len(filtered),
		},
	})
}

// handleConvertDownload 下载转换后的文件
// GET /api/convert/download/:taskId
func (s *Server) handleConvertDownload(c *gin.Context) {
	taskID := c.Param("taskId")

	convertTask, err := s.taskMgr.Get(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "转换任务不存在",
		})
		return
	}

	if convertTask.Status != task.StatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("文件尚未转换完成,当前状态: %s", convertTask.Status),
		})
		return
	}

	if _, err := os.Stat(convertTask.OutputPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "输出文件不存在",
		})
		return
	}

	// 设置响应头
	fileName := filepath.Base(convertTask.OutputPath)
	c.Header("Content-Type", "video/mp4")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))

	// 流式传输文件
	file, err := os.Open(convertTask.OutputPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "打开文件失败",
		})
		return
	}
	defer file.Close()

	io.Copy(c.Writer, file)
}

// ==================== 进度查询模块 ====================

// handleProgress 统一进度查询
// GET /api/progress/:id
func (s *Server) handleProgress(c *gin.Context) {
	id := c.Param("id")

	// 首先尝试作为上传任务查询
	if uploadTask, err := s.uploadMgr.GetUploadTask(id); err == nil {
		progress := 0.0
		if uploadTask.TotalChunks > 0 {
			progress = float64(uploadTask.UploadedChunks) / float64(uploadTask.TotalChunks) * 100
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"type":           "upload",
				"taskId":         id,
				"status":         uploadTask.Status,
				"progress":       int(progress),
				"uploadedChunks": uploadTask.UploadedChunks,
				"totalChunks":    uploadTask.TotalChunks,
				"fileName":       uploadTask.FileName,
				"fileSize":       uploadTask.FileSize,
				"createdAt":      uploadTask.CreatedAt,
				"updatedAt":      uploadTask.UpdatedAt,
			},
		})
		return
	}

	// 尝试作为转换任务查询
	if convertTask, err := s.taskMgr.Get(id); err == nil {
		// 如果任务状态为 completed,删除 inputPath 文件
		if convertTask.Status == task.StatusCompleted && convertTask.InputPath != "" {
			if _, err := os.Stat(convertTask.InputPath); err == nil {
				// 文件存在,尝试删除
				if err := os.Remove(convertTask.InputPath); err != nil {
					log.Printf("删除 inputPath 文件失败: %s, 错误: %v", convertTask.InputPath, err)
				} else {
					log.Printf("✓ 已删除 inputPath 文件: %s (任务 %s 已完成)", convertTask.InputPath, id)
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"type":         "convert",
				"taskId":       id,
				"status":       convertTask.Status,
				"progress":     convertTask.Progress,
				"inputPath":    convertTask.InputPath,
				"outputPath":   convertTask.OutputPath,
				"outputFormat": convertTask.OutputFormat,
				"quality":      convertTask.Quality,
				"error":        convertTask.Error,
				"createdAt":    convertTask.CreatedAt,
				"updatedAt":    convertTask.UpdatedAt,
				"completedAt":  convertTask.CompletedAt,
			},
		})
		return
	}

	// 任务不存在
	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"message": "任务不存在",
	})
}

// ==================== 辅助函数 ====================

// processConvertTask 处理转换任务
func (s *Server) processConvertTask(t *task.Task) {
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
		s.taskMgr.MarkCompleted(t.ID)
		log.Printf("任务 %s 转换完成", t.ID)

		// 删除输入文件(如果是上传的临时文件)
		if t.UploadID != "" {
			os.Remove(t.InputPath)
		}
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

// ==================== 文件管理模块 ====================

// handleDeleteFiles 批量删除本地文件
// POST /api/files/delete
func (s *Server) handleDeleteFiles(c *gin.Context) {
	var req struct {
		FilePaths []string `json:"filePaths" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "缺少必要参数: filePaths",
		})
		return
	}

	if len(req.FilePaths) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "filePaths 不能为空",
		})
		return
	}

	// 删除结果
	results := make([]gin.H, 0, len(req.FilePaths))
	successCount := 0
	failCount := 0

	for _, filePath := range req.FilePaths {
		// 检查文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			results = append(results, gin.H{
				"filePath": filePath,
				"success":  false,
				"message":  "文件不存在",
			})
			failCount++
			continue
		}

		// 安全检查:只允许删除 output 目录下的文件
		if !filepath.HasPrefix(filePath, s.config.OutputDir) &&
			!filepath.HasPrefix(filePath, s.config.DataDir) &&
			!filepath.HasPrefix(filePath, s.config.TempDir) {
			results = append(results, gin.H{
				"filePath": filePath,
				"success":  false,
				"message":  "无权限删除此文件(仅允许删除 output/data/temp 目录下的文件)",
			})
			failCount++
			continue
		}

		// 删除文件
		if err := os.Remove(filePath); err != nil {
			results = append(results, gin.H{
				"filePath": filePath,
				"success":  false,
				"message":  fmt.Sprintf("删除失败: %v", err),
			})
			failCount++
			log.Printf("删除文件失败: %s, 错误: %v", filePath, err)
		} else {
			results = append(results, gin.H{
				"filePath": filePath,
				"success":  true,
				"message":  "删除成功",
			})
			successCount++
			log.Printf("删除文件成功: %s", filePath)

			// 🔧 删除对应的任务记录
			// 遍历所有任务,找到 OutputPath 或 InputPath 匹配的任务并删除
			tasks := s.taskMgr.List()
			for _, task := range tasks {
				if task.OutputPath == filePath || task.InputPath == filePath {
					if err := s.taskMgr.Delete(task.ID); err == nil {
						log.Printf("已删除任务记录: %s (文件: %s)", task.ID, filePath)
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("处理完成: 成功 %d 个,失败 %d 个", successCount, failCount),
		"data": gin.H{
			"total":        len(req.FilePaths),
			"successCount": successCount,
			"failCount":    failCount,
			"results":      results,
		},
	})
}

// ==================== 辅助函数 ====================

// generateTaskID 生成任务ID
func generateTaskID() string {
	return fmt.Sprintf("task_%d", timeNow().UnixNano())
}

// timeNow 获取当前时间(便于测试)
var timeNow = func() time.Time {
	return time.Now()
}
