package server

import (
	"ffmpeg-binary/internal/config"
	"ffmpeg-binary/internal/converter"
	"ffmpeg-binary/internal/task"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Server HTTP 服务器
type Server struct {
	config    *config.Config
	converter *converter.Converter
	taskMgr   *task.Manager
	router    *gin.Engine
}

// New 创建服务器
func New(cfg *config.Config) *Server {
	gin.SetMode(gin.ReleaseMode)

	s := &Server{
		config:    cfg,
		converter: converter.New(cfg.FFmpegPath),
		taskMgr:   task.NewManager(),
		router:    gin.Default(),
	}

	s.setupRoutes()
	return s
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// CORS 中间件
	s.router.Use(corsMiddleware())

	// API 路由
	api := s.router.Group("/api/v1")
	{
		// 同步转换接口
		api.POST("/convert/sync", s.handleSyncConvert)

		// 异步转换接口
		api.POST("/convert/async", s.handleAsyncConvert)
		api.POST("/convert/async/:task_id/chunk", s.handleUploadChunk)

		// 任务管理接口
		api.GET("/task/:task_id", s.handleGetTask)
		api.GET("/task/:task_id/download", s.handleDownloadVideo)
		api.DELETE("/task/:task_id", s.handleDeleteTask)
		api.GET("/tasks", s.handleListTasks)
	}

	// 健康检查
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "port": s.config.Port})
	})
}

// Start 启动服务器
func (s *Server) Start() error {
	// 验证 FFmpeg
	if err := s.converter.Validate(); err != nil {
		return err
	}

	// 使用固定端口
	port := s.config.Port
	addr := fmt.Sprintf("%s:%d", s.config.Host, port)
	log.Printf("FFmpeg 服务启动成功: http://%s", addr)
	log.Printf("数据目录: %s", s.config.DataDir)

	// 启动 HTTP 服务
	srv := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Minute, // 大文件上传需要更长超时
		WriteTimeout: 30 * time.Minute,
	}

	return srv.ListenAndServe()
}

// corsMiddleware CORS 中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
