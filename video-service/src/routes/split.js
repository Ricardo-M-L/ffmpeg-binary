const express = require('express');
const router = express.Router();
const { splitVideoTask, readVideoFile, cleanupSplitFiles } = require('../services/splitService');

/**
 * POST /api/split/start
 * 开始视频切割任务
 * 
 * Body:
 * {
 *   taskId: string,          // 已转换的视频任务ID
 *   deleteIntervals: Array,  // 要删除的时间区间 [{start, end}, ...]
 *   videoDuration: number    // 视频总时长（秒）
 * }
 */
router.post('/start', async (req, res) => {
  try {
    const { taskId, deleteIntervals, videoDuration } = req.body;
    
    // 参数验证
    if (!taskId) {
      return res.status(400).json({
        success: false,
        error: '缺少taskId参数'
      });
    }
    
    if (!Array.isArray(deleteIntervals)) {
      return res.status(400).json({
        success: false,
        error: 'deleteIntervals必须是数组'
      });
    }
    
    if (!videoDuration || videoDuration <= 0) {
      return res.status(400).json({
        success: false,
        error: '无效的videoDuration'
      });
    }
    
    console.log('📥 [Split API] 收到切割请求:', {
      taskId,
      deleteIntervals: deleteIntervals.length,
      videoDuration
    });
    
    // 执行切割任务
    const result = await splitVideoTask(taskId, deleteIntervals, videoDuration);
    
    res.json(result);
    
  } catch (error) {
    console.error('❌ [Split API] 切割失败:', error);
    res.status(500).json({
      success: false,
      error: error.message || '切割失败'
    });
  }
});

/**
 * GET /api/split/download/:taskId/:segmentIndex
 * 下载指定的切割片段（返回视频流）
 * 
 * Params:
 * - taskId: 任务ID
 * - segmentIndex: 片段索引（1, 2, 3, ...）
 */
router.get('/download/:taskId/:segmentIndex', async (req, res) => {
  try {
    const { taskId, segmentIndex } = req.params;
    
    if (!taskId || !segmentIndex) {
      return res.status(400).json({
        success: false,
        error: '缺少必要参数'
      });
    }
    
    console.log(`📥 [Split API] 下载片段请求: ${taskId} - 片段${segmentIndex}`);
    
    // 构建文件路径
    const path = require('path');
    const OUTPUT_DIR = path.join(__dirname, '../../output');
    const fs = require('fs-extra');
    
    // 查找文件
    const files = await fs.readdir(OUTPUT_DIR);
    const targetFile = files.find(f => 
      f.includes(taskId) && 
      f.includes(`_part${segmentIndex}.mp4`)
    );
    
    if (!targetFile) {
      return res.status(404).json({
        success: false,
        error: `未找到片段文件: ${taskId} - part${segmentIndex}`
      });
    }
    
    const filePath = path.join(OUTPUT_DIR, targetFile);
    const stats = await fs.stat(filePath);
    
    // 设置响应头
    res.setHeader('Content-Type', 'video/mp4');
    res.setHeader('Content-Length', stats.size);
    res.setHeader('Content-Disposition', `attachment; filename="${targetFile}"`);
    res.setHeader('Accept-Ranges', 'bytes');
    
    // 流式传输文件
    const fileStream = fs.createReadStream(filePath);
    
    fileStream.on('error', (error) => {
      console.error(`❌ [Split API] 文件流错误:`, error);
      if (!res.headersSent) {
        res.status(500).json({
          success: false,
          error: '文件读取失败'
        });
      }
    });
    
    fileStream.pipe(res);
    
    console.log(`✅ [Split API] 开始传输片段: ${targetFile} (${(stats.size / 1024 / 1024).toFixed(2)}MB)`);
    
  } catch (error) {
    console.error('❌ [Split API] 下载失败:', error);
    if (!res.headersSent) {
      res.status(500).json({
        success: false,
        error: error.message || '下载失败'
      });
    }
  }
});

/**
 * DELETE /api/split/cleanup/:taskId
 * 清理指定任务的切割文件
 */
router.delete('/cleanup/:taskId', async (req, res) => {
  try {
    const { taskId } = req.params;
    
    if (!taskId) {
      return res.status(400).json({
        success: false,
        error: '缺少taskId参数'
      });
    }
    
    console.log(`🗑️ [Split API] 清理切割文件: ${taskId}`);
    
    await cleanupSplitFiles(taskId);
    
    res.json({
      success: true,
      message: '清理完成'
    });
    
  } catch (error) {
    console.error('❌ [Split API] 清理失败:', error);
    res.status(500).json({
      success: false,
      error: error.message || '清理失败'
    });
  }
});

module.exports = router;

