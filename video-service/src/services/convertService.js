/**
 * 视频转换服务
 * 
 * 功能：
 * - webm转mp4
 * - 支持自定义编码参数
 * - 实时进度追踪
 */

const ffmpeg = require('fluent-ffmpeg');
const path = require('path');
const fs = require('fs-extra');
const { v4: uuidv4 } = require('uuid');

// 如果环境变量中指定了ffmpeg路径，则使用
if (process.env.FFMPEG_PATH) {
  ffmpeg.setFfmpegPath(process.env.FFMPEG_PATH);
}

// 内存存储（生产环境建议使用Redis或数据库）
const convertTasks = new Map();

// 质量预设配置
const qualityPresets = {
  low: {
    videoBitrate: '500k',
    audioBitrate: '64k',
    videoCodec: 'libx264',
    audioCodec: 'aac',
    preset: 'veryfast',
    crf: 28
  },
  medium: {
    videoBitrate: '1000k',
    audioBitrate: '128k',
    videoCodec: 'libx264',
    audioCodec: 'aac',
    preset: 'medium',
    crf: 23
  },
  high: {
    videoBitrate: '2000k',
    audioBitrate: '192k',
    videoCodec: 'libx264',
    audioCodec: 'aac',
    preset: 'slow',
    crf: 18
  }
};

/**
 * 开始视频转换
 */
async function startConversion(params) {
  const {
    inputPath,
    outputFormat = 'mp4',
    quality = 'medium',
    options = {},
    uploadId
  } = params;

  // 验证输入文件
  if (!await fs.pathExists(inputPath)) {
    throw new Error(`输入文件不存在: ${inputPath}`);
  }

  // 生成任务ID和输出路径
  const taskId = uuidv4();
  const outputDir = process.env.OUTPUT_DIR || './output';
  await fs.ensureDir(outputDir);

  const inputFileName = path.basename(inputPath, path.extname(inputPath));
  const outputFileName = `${inputFileName}_converted.${outputFormat}`;
  const outputPath = path.join(outputDir, `${taskId}_${outputFileName}`);

  // 创建转换任务
  const task = {
    taskId,
    uploadId,
    inputPath,
    outputPath,
    outputFormat,
    quality,
    options,
    status: 'pending', // pending, processing, completed, failed, cancelled
    progress: 0,
    error: null,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    startedAt: null,
    completedAt: null,
    duration: 0,
    size: 0
  };

  convertTasks.set(taskId, task);
  console.log(`创建转换任务: ${taskId}, 输入: ${inputPath}, 输出: ${outputPath}`);

  // 异步执行转换
  performConversion(task).catch(err => {
    console.error(`转换任务执行失败: ${taskId}`, err);
  });

  return task;
}

/**
 * 执行视频转换
 */
async function performConversion(task) {
  const { taskId, inputPath, outputPath, outputFormat, quality, options } = task;

  try {
    task.status = 'processing';
    task.startedAt = new Date().toISOString();
    task.updatedAt = new Date().toISOString();

    console.log(`开始转换: ${taskId}`);

    // 获取质量预设
    const preset = qualityPresets[quality] || qualityPresets.medium;

    // 创建ffmpeg命令
    const command = ffmpeg(inputPath);

    // 设置视频编码器
    command.videoCodec(options.videoCodec || preset.videoCodec);

    // 设置音频编码器
    command.audioCodec(options.audioCodec || preset.audioCodec);

    // 设置比特率
    if (options.videoBitrate || preset.videoBitrate) {
      command.videoBitrate(options.videoBitrate || preset.videoBitrate);
    }

    if (options.audioBitrate || preset.audioBitrate) {
      command.audioBitrate(options.audioBitrate || preset.audioBitrate);
    }

    // 设置帧率
    if (options.fps) {
      command.fps(options.fps);
    }

    // 设置输出格式
    command.format(outputFormat);

    // 对于MP4格式，添加额外的参数以确保兼容性
    if (outputFormat === 'mp4') {
      command.outputOptions([
        '-movflags', 'faststart', // 优化流媒体播放
        '-preset', options.preset || preset.preset,
        '-crf', String(options.crf || preset.crf)
      ]);
    }

    // 如果有额外的自定义选项
    if (options.customOptions && Array.isArray(options.customOptions)) {
      command.outputOptions(options.customOptions);
    }

    // 监听进度
    command.on('progress', (progress) => {
      if (progress.percent) {
        task.progress = Math.round(progress.percent);
        task.updatedAt = new Date().toISOString();
        console.log(`转换进度 ${taskId}: ${task.progress}%`);
      }
    });

    // 监听完成
    command.on('end', async () => {
      try {
        // 获取输出文件信息
        const stats = await fs.stat(outputPath);
        
        task.status = 'completed';
        task.progress = 100;
        task.completedAt = new Date().toISOString();
        task.updatedAt = new Date().toISOString();
        task.size = stats.size;

        console.log(`转换完成: ${taskId}, 输出文件大小: ${stats.size} bytes`);

        // 🔧 删除原始的WebM文件（转换完成后不再需要）
        if (await fs.pathExists(inputPath)) {
          try {
            await fs.remove(inputPath);
            console.log(`✅ 已删除原始WebM文件: ${inputPath}`);
          } catch (removeError) {
            console.error(`⚠️ 删除原始WebM文件失败: ${inputPath}`, removeError);
            // 不影响转换任务的状态，继续
          }
        }
      } catch (error) {
        console.error(`获取输出文件信息失败: ${taskId}`, error);
        task.status = 'failed';
        task.error = error.message;
        task.updatedAt = new Date().toISOString();
      }
    });

    // 监听错误
    command.on('error', (err) => {
      console.error(`转换失败: ${taskId}`, err);
      task.status = 'failed';
      task.error = err.message;
      task.updatedAt = new Date().toISOString();
    });

    // 设置输出路径并开始转换
    command.save(outputPath);

    // 保存命令引用，用于取消操作
    task.command = command;
  } catch (error) {
    console.error(`执行转换任务失败: ${taskId}`, error);
    task.status = 'failed';
    task.error = error.message;
    task.updatedAt = new Date().toISOString();
  }
}

/**
 * 获取转换任务
 */
async function getConvertTask(taskId) {
  const task = convertTasks.get(taskId);
  if (!task) {
    return null;
  }

  // 返回任务副本，不包含内部属性
  const { command, ...publicTask } = task;
  return publicTask;
}

/**
 * 取消转换任务
 */
async function cancelConversion(taskId) {
  const task = convertTasks.get(taskId);
  if (!task) {
    throw new Error(`转换任务不存在: ${taskId}`);
  }

  if (task.status === 'processing' && task.command) {
    // 杀死ffmpeg进程
    task.command.kill('SIGKILL');
  }

  task.status = 'cancelled';
  task.updatedAt = new Date().toISOString();

  // 删除未完成的输出文件
  if (await fs.pathExists(task.outputPath)) {
    await fs.remove(task.outputPath);
  }

  console.log(`转换任务已取消: ${taskId}`);
  return task;
}

/**
 * 获取转换任务列表
 */
async function listConvertTasks(filters = {}) {
  const { status, limit = 50 } = filters;

  let tasks = Array.from(convertTasks.values());

  // 过滤状态
  if (status) {
    tasks = tasks.filter(task => task.status === status);
  }

  // 按创建时间倒序排序
  tasks.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));

  // 限制数量
  tasks = tasks.slice(0, limit);

  // 移除内部属性
  return tasks.map(({ command, ...publicTask }) => publicTask);
}

/**
 * 清理过期任务
 */
async function cleanupExpiredTasks() {
  const retentionHours = parseInt(process.env.FILE_RETENTION_HOURS || '24');
  const expirationTime = Date.now() - retentionHours * 60 * 60 * 1000;

  for (const [taskId, task] of convertTasks.entries()) {
    const taskTime = new Date(task.createdAt).getTime();
    
    if (taskTime < expirationTime && ['completed', 'failed', 'cancelled'].includes(task.status)) {
      console.log(`清理过期转换任务: ${taskId}`);
      
      // 删除输出文件
      if (task.outputPath && await fs.pathExists(task.outputPath)) {
        await fs.remove(task.outputPath);
      }
      
      // 从内存中移除
      convertTasks.delete(taskId);
    }
  }
}

// 定时清理过期任务（每小时执行一次）
setInterval(cleanupExpiredTasks, 60 * 60 * 1000);

module.exports = {
  startConversion,
  getConvertTask,
  cancelConversion,
  listConvertTasks,
  cleanupExpiredTasks
};


