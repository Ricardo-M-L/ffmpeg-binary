package server

import (
	"goalfy-mediaconverter/internal/split"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// handleSplitStart 处理视频切割开始请求
// POST /api/split/start
func (s *Server) handleSplitStart(c *gin.Context) {
	var req split.SplitRequest

	// 绑定JSON请求
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, split.SplitResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// 参数验证
	if req.TaskID == "" {
		c.JSON(http.StatusBadRequest, split.SplitResponse{
			Success: false,
			Error:   "缺少taskId参数",
		})
		return
	}

	if req.DeleteIntervals == nil {
		c.JSON(http.StatusBadRequest, split.SplitResponse{
			Success: false,
			Error:   "deleteIntervals必须是数组",
		})
		return
	}

	if req.VideoDuration <= 0 {
		c.JSON(http.StatusBadRequest, split.SplitResponse{
			Success: false,
			Error:   "无效的videoDuration",
		})
		return
	}

	// 执行切割
	result, err := s.splitter.SplitVideo(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, split.SplitResponse{
			Success: false,
			Error:   "切割失败: " + err.Error(),
		})
		return
	}

	// 返回结果
	if result.Success {
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusInternalServerError, result)
	}
}

// handleSplitDownload 处理片段下载请求
// GET /api/split/download/:taskId/:segmentIndex
func (s *Server) handleSplitDownload(c *gin.Context) {
	taskID := c.Param("taskId")
	segmentIndexStr := c.Param("segmentIndex")

	// 参数验证
	if taskID == "" || segmentIndexStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少必要参数",
		})
		return
	}

	// 解析片段索引
	segmentIndex, err := strconv.Atoi(segmentIndexStr)
	if err != nil || segmentIndex <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的segmentIndex",
		})
		return
	}

	// 查找文件
	filePath, err := s.splitter.FindSegmentFile(taskID, segmentIndex)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "未找到片段文件: " + taskID + " - part" + segmentIndexStr,
		})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "video/mp4")
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Disposition", "attachment; filename=\""+taskID+"_part"+segmentIndexStr+".mp4\"")

	// 流式传输文件
	c.File(filePath)
}

// handleSplitCleanup 处理清理切割文件请求
// DELETE /api/split/cleanup/:taskId
func (s *Server) handleSplitCleanup(c *gin.Context) {
	taskID := c.Param("taskId")

	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少taskId参数",
		})
		return
	}

	// 执行清理
	count, err := s.splitter.CleanupSplitFiles(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "清理失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "清理完成",
		"deleted": count,
	})
}
