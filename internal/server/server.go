package server

import (
	"ffmpeg-binary/internal/config"
	"ffmpeg-binary/internal/converter"
	"ffmpeg-binary/internal/task"
	"ffmpeg-binary/internal/upload"
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
	uploadMgr *upload.Manager
	router    *gin.Engine
}

// New 创建服务器
func New(cfg *config.Config) *Server {
	gin.SetMode(gin.ReleaseMode)

	s := &Server{
		config:    cfg,
		converter: converter.New(cfg.FFmpegPath),
		taskMgr:   task.NewManager(),
		uploadMgr: upload.NewManager(cfg.TempDir, cfg.DataDir),
		router:    gin.Default(),
	}

	s.setupRoutes()
	return s
}

// setupRoutes 设置路由(完全兼容 video-service)
func (s *Server) setupRoutes() {
	// CORS 中间件
	s.router.Use(corsMiddleware())

	// API 路由组
	api := s.router.Group("/api")
	{
		// 上传模块
		upload := api.Group("/upload")
		{
			upload.POST("/init", s.handleUploadInit)
			upload.POST("/chunk", s.handleUploadChunk)
			upload.GET("/status/:uploadId", s.handleUploadStatus)
			upload.POST("/cancel/:uploadId", s.handleUploadCancel)
		}

		// 转换模块
		convert := api.Group("/convert")
		{
			convert.POST("/start", s.handleConvertStart)
			convert.GET("/status/:taskId", s.handleConvertStatus)
			convert.POST("/cancel/:taskId", s.handleConvertCancel)
			convert.GET("/list", s.handleConvertList)
			convert.GET("/download/:taskId", s.handleConvertDownload)
		}

		// 进度查询模块
		progress := api.Group("/progress")
		{
			progress.GET("/:id", s.handleProgress)
		}

		// 文件管理模块
		files := api.Group("/files")
		{
			files.POST("/delete", s.handleDeleteFiles)
		}
	}

	// 健康检查
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "ffmpeg-binary",
			"version":   "1.0.0",
		})
	})

	// 静态文件服务(下载输出文件)
	s.router.Static("/downloads", s.config.OutputDir)
}

// Start 启动服务器
func (s *Server) Start() error {
	// 使用固定端口
	port := s.config.Port
	addr := fmt.Sprintf("%s:%d", s.config.Host, port)

	log.Println("\n===========================================")
	log.Println("🚀 FFmpeg Binary 服务启动成功!")
	log.Println("===========================================")
	log.Printf("📡 服务地址: http://%s", addr)
	log.Printf("📝 健康检查: http://%s/health", addr)
	log.Printf("📂 数据目录: %s", s.config.DataDir)
	log.Printf("📂 临时目录: %s", s.config.TempDir)
	log.Printf("📂 输出目录: %s", s.config.OutputDir)
	log.Println("===========================================\n")

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
