/**
 * 视频转换路由
 * 
 * 功能：
 * - webm转mp4
 * - 支持自定义编码参数
 */

const express = require('express');
const path = require('path');
const fs = require('fs-extra');
const convertService = require('../services/convertService');

const router = express.Router();

/**
 * POST /api/convert/start
 * 开始视频转换任务
 * 
 * 请求体：
 * {
 *   "uploadId": "xxx",           // 上传任务ID（二选一）
 *   "filePath": "/path/to/file", // 或直接指定文件路径（二选一）
 *   "outputFormat": "mp4",       // 输出格式，默认mp4
 *   "quality": "high",           // 质量: low/medium/high，默认medium
 *   "options": {                 // 可选的ffmpeg参数
 *     "videoBitrate": "1000k",
 *     "audioBitrate": "128k",
 *     "fps": 30
 *   }
 * }
 */
router.post('/start', async (req, res) => {
  try {
    const {
      uploadId,
      filePath,
      outputFormat = 'mp4',
      quality = 'medium',
      options = {}
    } = req.body;

    // 参数验证
    if (!uploadId && !filePath) {
      return res.status(400).json({
        success: false,
        message: '必须提供uploadId或filePath'
      });
    }

    // 获取输入文件路径
    let inputPath;
    if (uploadId) {
      const chunkService = require('../services/chunkService');
      const uploadTask = await chunkService.getUploadTask(uploadId);
      
      if (!uploadTask) {
        return res.status(404).json({
          success: false,
          message: '上传任务不存在'
        });
      }

      if (uploadTask.status !== 'merged') {
        return res.status(400).json({
          success: false,
          message: `文件尚未合并完成，当前状态: ${uploadTask.status}`
        });
      }

      inputPath = uploadTask.mergedPath;
    } else {
      inputPath = filePath;
    }

    // 检查输入文件是否存在
    if (!await fs.pathExists(inputPath)) {
      return res.status(404).json({
        success: false,
        message: '输入文件不存在'
      });
    }

    // 开始转换
    const convertTask = await convertService.startConversion({
      inputPath,
      outputFormat,
      quality,
      options,
      uploadId
    });

    res.json({
      success: true,
      message: '转换任务已启动',
      data: {
        taskId: convertTask.taskId,
        inputPath,
        outputFormat,
        quality
      }
    });
  } catch (error) {
    console.error('启动转换任务失败:', error);
    res.status(500).json({
      success: false,
      message: '启动转换任务失败',
      error: error.message
    });
  }
});

/**
 * GET /api/convert/status/:taskId
 * 查询转换状态
 */
router.get('/status/:taskId', async (req, res) => {
  try {
    const { taskId } = req.params;

    const task = await convertService.getConvertTask(taskId);

    if (!task) {
      return res.status(404).json({
        success: false,
        message: '转换任务不存在'
      });
    }

    res.json({
      success: true,
      data: task
    });
  } catch (error) {
    console.error('查询转换状态失败:', error);
    res.status(500).json({
      success: false,
      message: '查询转换状态失败',
      error: error.message
    });
  }
});

/**
 * POST /api/convert/cancel/:taskId
 * 取消转换任务
 */
router.post('/cancel/:taskId', async (req, res) => {
  try {
    const { taskId } = req.params;

    await convertService.cancelConversion(taskId);

    res.json({
      success: true,
      message: '转换任务已取消'
    });
  } catch (error) {
    console.error('取消转换任务失败:', error);
    res.status(500).json({
      success: false,
      message: '取消转换任务失败',
      error: error.message
    });
  }
});

/**
 * GET /api/convert/list
 * 获取所有转换任务列表
 */
router.get('/list', async (req, res) => {
  try {
    const { status, limit = 50 } = req.query;

    const tasks = await convertService.listConvertTasks({
      status,
      limit: parseInt(limit)
    });

    res.json({
      success: true,
      data: {
        tasks,
        total: tasks.length
      }
    });
  } catch (error) {
    console.error('获取转换任务列表失败:', error);
    res.status(500).json({
      success: false,
      message: '获取转换任务列表失败',
      error: error.message
    });
  }
});

/**
 * GET /api/convert/download/:taskId
 * 下载转换后的文件（二进制流）
 */
router.get('/download/:taskId', async (req, res) => {
  try {
    const { taskId } = req.params;
    const fs = require('fs-extra');

    const task = await convertService.getConvertTask(taskId);

    if (!task) {
      return res.status(404).json({
        success: false,
        message: '转换任务不存在'
      });
    }

    if (task.status !== 'completed') {
      return res.status(400).json({
        success: false,
        message: `文件尚未转换完成，当前状态: ${task.status}`
      });
    }

    if (!await fs.pathExists(task.outputPath)) {
      return res.status(404).json({
        success: false,
        message: '输出文件不存在'
      });
    }

    // 设置响应头
    const fileName = task.outputPath.split('/').pop();
    res.setHeader('Content-Type', 'video/mp4');
    res.setHeader('Content-Disposition', `attachment; filename="${fileName}"`);

    // 流式传输文件
    const fileStream = fs.createReadStream(task.outputPath);
    fileStream.pipe(res);

    fileStream.on('error', (error) => {
      console.error('文件流传输失败:', error);
      if (!res.headersSent) {
        res.status(500).json({
          success: false,
          message: '文件传输失败'
        });
      }
    });
  } catch (error) {
    console.error('下载文件失败:', error);
    res.status(500).json({
      success: false,
      message: '下载文件失败',
      error: error.message
    });
  }
});

module.exports = router;

