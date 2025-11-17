/**
 * 进度查询路由
 * 
 * 功能：
 * - 统一查询上传和转换进度
 * - 支持WebSocket实时推送（可选）
 */

const express = require('express');
const chunkService = require('../services/chunkService');
const convertService = require('../services/convertService');

const router = express.Router();

/**
 * GET /api/progress/:id
 * 查询任务进度（自动识别上传任务或转换任务）
 */
router.get('/:id', async (req, res) => {
  try {
    const { id } = req.params;

    // 首先尝试作为上传任务查询
    let uploadTask = await chunkService.getUploadTask(id);
    if (uploadTask) {
      return res.json({
        success: true,
        data: {
          type: 'upload',
          taskId: id,
          status: uploadTask.status,
          progress: uploadTask.uploadedChunks / uploadTask.totalChunks * 100,
          uploadedChunks: uploadTask.uploadedChunks,
          totalChunks: uploadTask.totalChunks,
          fileName: uploadTask.fileName,
          fileSize: uploadTask.fileSize,
          createdAt: uploadTask.createdAt,
          updatedAt: uploadTask.updatedAt
        }
      });
    }

    // 尝试作为转换任务查询
    let convertTask = await convertService.getConvertTask(id);
    if (convertTask) {
      return res.json({
        success: true,
        data: {
          type: 'convert',
          taskId: id,
          status: convertTask.status,
          progress: convertTask.progress || 0,
          inputPath: convertTask.inputPath,
          outputPath: convertTask.outputPath,
          outputFormat: convertTask.outputFormat,
          quality: convertTask.quality,
          error: convertTask.error,
          createdAt: convertTask.createdAt,
          updatedAt: convertTask.updatedAt,
          completedAt: convertTask.completedAt
        }
      });
    }

    // 任务不存在
    return res.status(404).json({
      success: false,
      message: '任务不存在'
    });
  } catch (error) {
    console.error('查询进度失败:', error);
    res.status(500).json({
      success: false,
      message: '查询进度失败',
      error: error.message
    });
  }
});

/**
 * GET /api/progress/batch
 * 批量查询多个任务进度
 * 
 * Query参数：
 * - ids: 任务ID列表，逗号分隔
 */
router.get('/batch', async (req, res) => {
  try {
    const { ids } = req.query;

    if (!ids) {
      return res.status(400).json({
        success: false,
        message: '缺少ids参数'
      });
    }

    const idList = ids.split(',').map(id => id.trim());
    const results = [];

    for (const id of idList) {
      try {
        // 查询上传任务
        let uploadTask = await chunkService.getUploadTask(id);
        if (uploadTask) {
          results.push({
            type: 'upload',
            taskId: id,
            status: uploadTask.status,
            progress: uploadTask.uploadedChunks / uploadTask.totalChunks * 100,
            fileName: uploadTask.fileName
          });
          continue;
        }

        // 查询转换任务
        let convertTask = await convertService.getConvertTask(id);
        if (convertTask) {
          results.push({
            type: 'convert',
            taskId: id,
            status: convertTask.status,
            progress: convertTask.progress || 0,
            outputPath: convertTask.outputPath
          });
          continue;
        }

        // 任务不存在
        results.push({
          taskId: id,
          status: 'not_found'
        });
      } catch (err) {
        console.error(`查询任务${id}失败:`, err);
        results.push({
          taskId: id,
          status: 'error',
          error: err.message
        });
      }
    }

    res.json({
      success: true,
      data: {
        tasks: results,
        total: results.length
      }
    });
  } catch (error) {
    console.error('批量查询进度失败:', error);
    res.status(500).json({
      success: false,
      message: '批量查询进度失败',
      error: error.message
    });
  }
});

module.exports = router;


