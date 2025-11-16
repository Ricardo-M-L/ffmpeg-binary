/**
 * 文件切片上传路由
 * 
 * 功能：
 * - 接收文件切片
 * - 存储切片信息
 * - 完成后自动合并切片
 */

const express = require('express');
const multer = require('multer');
const path = require('path');
const fs = require('fs-extra');
const { v4: uuidv4 } = require('uuid');
const chunkService = require('../services/chunkService');

const router = express.Router();

// 使用内存存储，避免multer的body解析时序问题
const upload = multer({
  storage: multer.memoryStorage(),
  limits: {
    fileSize: parseInt(process.env.MAX_CHUNK_SIZE || '10485760') // 默认10MB
  }
});

/**
 * POST /api/upload/init
 * 初始化上传任务
 * 
 * 请求体：
 * {
 *   "fileName": "video.webm",
 *   "fileSize": 1024000,
 *   "totalChunks": 10,
 *   "chunkSize": 102400
 * }
 */
router.post('/init', async (req, res) => {
  try {
    const { fileName, fileSize, totalChunks, chunkSize } = req.body;

    // 参数验证
    if (!fileName || !fileSize || !totalChunks) {
      return res.status(400).json({
        success: false,
        message: '缺少必要参数: fileName, fileSize, totalChunks'
      });
    }

    // 生成唯一的上传ID
    const uploadId = uuidv4();

    // 创建上传任务记录
    const uploadTask = await chunkService.createUploadTask({
      uploadId,
      fileName,
      fileSize,
      totalChunks,
      chunkSize
    });

    res.json({
      success: true,
      message: '上传任务初始化成功',
      data: {
        uploadId,
        fileName,
        totalChunks
      }
    });
  } catch (error) {
    console.error('初始化上传任务失败:', error);
    res.status(500).json({
      success: false,
      message: '初始化上传任务失败',
      error: error.message
    });
  }
});

/**
 * POST /api/upload/chunk
 * 上传单个切片
 * 
 * FormData:
 * - file: 文件切片
 * - uploadId: 上传任务ID
 * - chunkIndex: 切片索引（从0开始）
 * - chunkHash: 切片MD5（可选，用于校验）
 */
router.post('/chunk', upload.single('file'), async (req, res) => {
  try {
    const { uploadId, chunkIndex, chunkHash } = req.body;
    const file = req.file;

    // 参数验证
    if (!uploadId || chunkIndex === undefined || !file) {
      return res.status(400).json({
        success: false,
        message: '缺少必要参数: uploadId, chunkIndex, file'
      });
    }

    // 手动保存文件到磁盘
    const chunkDir = path.join(process.env.TEMP_DIR || './temp', uploadId);
    await fs.ensureDir(chunkDir);
    const chunkPath = path.join(chunkDir, `chunk_${chunkIndex}`);
    await fs.writeFile(chunkPath, file.buffer);

    // 记录切片上传
    await chunkService.recordChunk({
      uploadId,
      chunkIndex: parseInt(chunkIndex),
      chunkPath,
      chunkSize: file.size,
      chunkHash
    });

    // 检查是否所有切片都已上传
    const uploadTask = await chunkService.getUploadTask(uploadId);
    const isComplete = uploadTask.uploadedChunks === uploadTask.totalChunks;

    res.json({
      success: true,
      message: '切片上传成功',
      data: {
        uploadId,
        chunkIndex: parseInt(chunkIndex),
        uploadedChunks: uploadTask.uploadedChunks,
        totalChunks: uploadTask.totalChunks,
        isComplete
      }
    });

    // 如果所有切片都已上传，开始合并
    if (isComplete) {
      console.log(`所有切片上传完成，开始合并文件: ${uploadId}`);
      // 异步执行合并操作
      chunkService.mergeChunks(uploadId).catch(err => {
        console.error('合并切片失败:', err);
      });
    }
  } catch (error) {
    console.error('上传切片失败:', error);
    res.status(500).json({
      success: false,
      message: '上传切片失败',
      error: error.message
    });
  }
});

/**
 * GET /api/upload/status/:uploadId
 * 查询上传状态
 */
router.get('/status/:uploadId', async (req, res) => {
  try {
    const { uploadId } = req.params;

    const uploadTask = await chunkService.getUploadTask(uploadId);

    if (!uploadTask) {
      return res.status(404).json({
        success: false,
        message: '上传任务不存在'
      });
    }

    res.json({
      success: true,
      data: uploadTask
    });
  } catch (error) {
    console.error('查询上传状态失败:', error);
    res.status(500).json({
      success: false,
      message: '查询上传状态失败',
      error: error.message
    });
  }
});

/**
 * POST /api/upload/cancel/:uploadId
 * 取消上传任务
 */
router.post('/cancel/:uploadId', async (req, res) => {
  try {
    const { uploadId } = req.params;

    await chunkService.cancelUpload(uploadId);

    res.json({
      success: true,
      message: '上传任务已取消'
    });
  } catch (error) {
    console.error('取消上传任务失败:', error);
    res.status(500).json({
      success: false,
      message: '取消上传任务失败',
      error: error.message
    });
  }
});

module.exports = router;

