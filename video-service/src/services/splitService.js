const fs = require('fs-extra');
const path = require('path');
const ffmpeg = require('fluent-ffmpeg');
const ffmpegStatic = require('ffmpeg-static');

// 设置FFmpeg路径
ffmpeg.setFfmpegPath(ffmpegStatic);

const OUTPUT_DIR = path.join(__dirname, '../../output');

/**
 * 根据删除区间计算保留的视频片段
 * 
 * @param {number} videoDuration - 视频总时长（秒）
 * @param {Array} deleteSegments - 要删除的片段 [{start, end}, ...]
 * @returns {Array} 保留的片段 [{start, end}, ...]
 */
function calculateRetainedSegments(videoDuration, deleteSegments) {
  if (!deleteSegments || deleteSegments.length === 0) {
    return [{ start: 0, end: videoDuration }];
  }
  
  // 按start时间排序
  const sortedDeletes = [...deleteSegments].sort((a, b) => a.start - b.start);
  
  const retained = [];
  let currentPos = 0;
  
  for (const segment of sortedDeletes) {
    if (currentPos < segment.start) {
      // 添加删除片段之前的保留部分
      retained.push({
        start: currentPos,
        end: segment.start
      });
    }
    currentPos = Math.max(currentPos, segment.end);
  }
  
  // 添加最后一个片段
  if (currentPos < videoDuration) {
    retained.push({
      start: currentPos,
      end: videoDuration
    });
  }
  
  return retained.filter(seg => seg.end > seg.start); // 过滤掉无效片段
}

/**
 * 使用FFmpeg切割视频片段（无损复制，不重新编码）
 * 
 * @param {string} inputPath - 输入视频路径
 * @param {number} startTime - 开始时间（秒）
 * @param {number} endTime - 结束时间（秒）
 * @param {string} outputPath - 输出路径
 * @returns {Promise<Object>} 切割结果
 */
function splitVideoSegment(inputPath, startTime, endTime, outputPath) {
  return new Promise((resolve, reject) => {
    console.log(`🎬 [Split] 切割片段: ${startTime}s - ${endTime}s`);
    
    const duration = endTime - startTime;
    
    ffmpeg(inputPath)
      .setStartTime(startTime)
      .setDuration(duration)
      // -c copy 表示无损复制，不重新编码（保留原质量和音频）
      .outputOptions([
        '-c copy',           // 无损复制音视频流
        '-f mp4',            // 明确指定输出格式为MP4
        '-movflags +faststart',  // MP4优化：将moov atom移到文件开头，支持流媒体播放
        '-avoid_negative_ts 1'   // 避免负时间戳问题
      ])
      .output(outputPath)
      .on('start', (commandLine) => {
        console.log(`📹 [Split] FFmpeg命令: ${commandLine}`);
      })
      .on('progress', (progress) => {
        if (progress.percent) {
          console.log(`⏳ [Split] 进度: ${progress.percent.toFixed(1)}%`);
        }
      })
      .on('end', () => {
        console.log(`✅ [Split] 切割完成: ${outputPath}`);
        
        // 获取文件大小和格式信息
        const stats = fs.statSync(outputPath);
        const fileExt = path.extname(outputPath);
        console.log(`📄 [Split] 输出文件格式: ${fileExt} (MP4)`);
        console.log(`📊 [Split] 输出文件大小: ${(stats.size / 1024 / 1024).toFixed(2)}MB`);
        
        resolve({
          success: true,
          outputPath: outputPath,
          size: stats.size,
          duration: duration,
          startTime: startTime,
          endTime: endTime
        });
      })
      .on('error', (err) => {
        console.error(`❌ [Split] 切割失败:`, err);
        reject(err);
      })
      .run();
  });
}

/**
 * 切割视频任务
 * 
 * @param {string} taskId - 转换任务ID（已转换的MP4文件对应的taskId）
 * @param {Array} deleteIntervals - 要删除的时间区间
 * @param {number} videoDuration - 视频总时长
 * @returns {Promise<Object>} 切割结果
 */
async function splitVideoTask(taskId, deleteIntervals, videoDuration) {
  try {
    console.log(`🎬 [Split] 开始切割任务:`, { taskId, deleteIntervals, videoDuration });
    
    // 查找已转换的MP4文件
    const files = await fs.readdir(OUTPUT_DIR);
    const convertedFile = files.find(f => 
      f.includes(taskId) && f.endsWith('_converted.mp4')
    );
    
    if (!convertedFile) {
      throw new Error(`未找到已转换的视频文件: ${taskId}`);
    }
    
    const inputPath = path.join(OUTPUT_DIR, convertedFile);
    console.log(`📂 [Split] 输入文件: ${inputPath}`);
    
    // 验证输入文件是MP4格式
    const inputStats = await fs.stat(inputPath);
    console.log(`📊 [Split] 输入文件大小: ${(inputStats.size / 1024 / 1024).toFixed(2)}MB`);
    console.log(`📄 [Split] 输入文件格式: MP4 (${convertedFile})`);
    
    // 计算保留的片段
    const retainedSegments = calculateRetainedSegments(videoDuration, deleteIntervals);
    console.log(`📊 [Split] 将切割为 ${retainedSegments.length} 个片段:`, retainedSegments);
    
    if (retainedSegments.length === 0) {
      throw new Error('没有要保留的视频片段');
    }
    
    // 提取基础文件名
    const baseFileName = convertedFile.replace('_converted.mp4', '');
    
    // 切割所有片段
    const results = [];
    
    for (let i = 0; i < retainedSegments.length; i++) {
      const segment = retainedSegments[i];
      const segmentIndex = i + 1;
      const outputFileName = `${baseFileName}_part${segmentIndex}.mp4`;
      const outputPath = path.join(OUTPUT_DIR, outputFileName);
      
      console.log(`🎬 [Split] 切割片段 ${segmentIndex}/${retainedSegments.length}`);
      
      // 使用FFmpeg切割
      const result = await splitVideoSegment(
        inputPath,
        segment.start,
        segment.end,
        outputPath
      );
      
      results.push({
        ...result,
        segmentIndex: segmentIndex,
        fileName: outputFileName,
        originalStart: segment.start,
        originalEnd: segment.end
      });
    }
    
    console.log(`🎉 [Split] 切割完成! 共 ${results.length} 个片段`);
    
    // 🔧 删除原始的完整MP4文件（切割完成后不再需要）
    if (await fs.pathExists(inputPath)) {
      try {
        await fs.remove(inputPath);
        console.log(`✅ 已删除原始完整MP4文件: ${inputPath}`);
      } catch (removeError) {
        console.error(`⚠️ 删除原始MP4文件失败: ${inputPath}`, removeError);
        // 不影响切割任务的状态，继续
      }
    }
    
    return {
      success: true,
      taskId: taskId,
      totalSegments: results.length,
      segments: results
    };
    
  } catch (error) {
    console.error(`❌ [Split] 切割任务失败:`, error);
    throw error;
  }
}

/**
 * 读取视频文件为Buffer
 * 
 * @param {string} filePath - 文件路径
 * @returns {Promise<Buffer>} 文件Buffer
 */
async function readVideoFile(filePath) {
  return await fs.readFile(filePath);
}

/**
 * 清理切割后的临时文件
 * 
 * @param {string} taskId - 任务ID
 * @returns {Promise<void>}
 */
async function cleanupSplitFiles(taskId) {
  try {
    const files = await fs.readdir(OUTPUT_DIR);
    const splitFiles = files.filter(f => 
      f.includes(taskId) && f.includes('_part') && f.endsWith('.mp4')
    );
    
    for (const file of splitFiles) {
      const filePath = path.join(OUTPUT_DIR, file);
      await fs.remove(filePath);
      console.log(`🗑️ [Split] 已清理: ${file}`);
    }
  } catch (error) {
    console.error(`❌ [Split] 清理文件失败:`, error);
  }
}

module.exports = {
  splitVideoTask,
  readVideoFile,
  cleanupSplitFiles,
  calculateRetainedSegments
};

