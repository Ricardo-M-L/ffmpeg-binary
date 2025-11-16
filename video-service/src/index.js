/**
 * Goalfy 视频处理服务 - 主入口文件
 * 
 * 功能：
 * 1. 文件切片上传
 * 2. 切片合并
 * 3. webm转mp4视频转换
 * 4. 转换进度查询
 */

const express = require('express');
const cors = require('cors');
const dotenv = require('dotenv');
const fs = require('fs-extra');
const path = require('path');

// 加载环境变量
dotenv.config();

// 导入路由
const uploadRoutes = require('./routes/upload');
const convertRoutes = require('./routes/convert');
const progressRoutes = require('./routes/progress');

// 创建Express应用
const app = express();
const PORT = process.env.PORT || 3000;

// 中间件配置
app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// 日志中间件
app.use((req, res, next) => {
  console.log(`[${new Date().toISOString()}] ${req.method} ${req.path}`);
  next();
});

// 确保必要的目录存在
const ensureDirectories = async () => {
  const dirs = [
    process.env.UPLOAD_DIR || './uploads',
    process.env.TEMP_DIR || './temp',
    process.env.OUTPUT_DIR || './output'
  ];

  for (const dir of dirs) {
    await fs.ensureDir(dir);
    console.log(`✓ 目录已创建/确认: ${dir}`);
  }
};

// 路由配置
app.use('/api/upload', uploadRoutes);
app.use('/api/convert', convertRoutes);
app.use('/api/progress', progressRoutes);

// 健康检查接口
app.get('/health', (req, res) => {
  res.json({
    status: 'ok',
    timestamp: new Date().toISOString(),
    service: 'goalfy-video-service',
    version: '1.0.0'
  });
});

// 静态文件服务 - 用于下载已转换的文件
app.use('/downloads', express.static(process.env.OUTPUT_DIR || './output'));

// 404处理
app.use((req, res) => {
  res.status(404).json({
    success: false,
    message: '接口不存在',
    path: req.path
  });
});

// 错误处理中间件
app.use((err, req, res, next) => {
  console.error('服务器错误:', err);
  res.status(500).json({
    success: false,
    message: '服务器内部错误',
    error: process.env.NODE_ENV === 'development' ? err.message : undefined
  });
});

// 启动服务
const startServer = async () => {
  try {
    // 创建必要的目录
    await ensureDirectories();

    // 启动HTTP服务器
    app.listen(PORT, () => {
      console.log('\n===========================================');
      console.log('🚀 Goalfy 视频处理服务启动成功！');
      console.log('===========================================');
      console.log(`📡 服务地址: http://localhost:${PORT}`);
      console.log(`📝 健康检查: http://localhost:${PORT}/health`);
      console.log(`📂 上传目录: ${process.env.UPLOAD_DIR || './uploads'}`);
      console.log(`📂 输出目录: ${process.env.OUTPUT_DIR || './output'}`);
      console.log('===========================================\n');
    });
  } catch (error) {
    console.error('❌ 服务启动失败:', error);
    process.exit(1);
  }
};

// 优雅关闭
process.on('SIGTERM', async () => {
  console.log('\n收到SIGTERM信号，正在优雅关闭...');
  process.exit(0);
});

process.on('SIGINT', async () => {
  console.log('\n收到SIGINT信号，正在优雅关闭...');
  process.exit(0);
});

// 启动服务
startServer();

module.exports = app;

