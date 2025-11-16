/**
 * 文件切片服务
 * 
 * 功能：
 * - 管理上传任务
 * - 合并文件切片
 * - 清理临时文件
 */

const fs = require('fs-extra');
const path = require('path');

// 内存存储（生产环境建议使用Redis或数据库）
const uploadTasks = new Map();

/**
 * 创建上传任务
 */
async function createUploadTask(taskData) {
  const { uploadId, fileName, fileSize, totalChunks, chunkSize } = taskData;

  const task = {
    uploadId,
    fileName,
    fileSize,
    totalChunks,
    chunkSize,
    uploadedChunks: 0,
    chunks: new Array(totalChunks).fill(null),
    status: 'uploading', // uploading, merging, merged, failed, cancelled
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    mergedPath: null
  };

  uploadTasks.set(uploadId, task);
  console.log(`创建上传任务: ${uploadId}, 文件: ${fileName}, 切片数: ${totalChunks}`);

  return task;
}

/**
 * 记录切片上传
 */
async function recordChunk(chunkData) {
  const { uploadId, chunkIndex, chunkPath, chunkSize, chunkHash } = chunkData;

  const task = uploadTasks.get(uploadId);
  if (!task) {
    throw new Error(`上传任务不存在: ${uploadId}`);
  }

  // 记录切片信息
  task.chunks[chunkIndex] = {
    index: chunkIndex,
    path: chunkPath,
    size: chunkSize,
    hash: chunkHash,
    uploadedAt: new Date().toISOString()
  };

  task.uploadedChunks++;
  task.updatedAt = new Date().toISOString();

  console.log(`切片上传完成: ${uploadId}, 索引: ${chunkIndex}, 进度: ${task.uploadedChunks}/${task.totalChunks}`);

  return task;
}

/**
 * 获取上传任务
 */
async function getUploadTask(uploadId) {
  return uploadTasks.get(uploadId) || null;
}

/**
 * 合并文件切片
 */
async function mergeChunks(uploadId) {
  const task = uploadTasks.get(uploadId);
  if (!task) {
    throw new Error(`上传任务不存在: ${uploadId}`);
  }

  // 检查所有切片是否都已上传
  if (task.uploadedChunks !== task.totalChunks) {
    throw new Error(`切片未上传完整: ${task.uploadedChunks}/${task.totalChunks}`);
  }

  // 检查切片是否有缺失
  const missingChunks = [];
  for (let i = 0; i < task.totalChunks; i++) {
    if (!task.chunks[i]) {
      missingChunks.push(i);
    }
  }

  if (missingChunks.length > 0) {
    throw new Error(`切片缺失: ${missingChunks.join(', ')}`);
  }

  console.log(`开始合并切片: ${uploadId}`);
  task.status = 'merging';
  task.updatedAt = new Date().toISOString();

  try {
    // 创建输出目录
    const uploadDir = process.env.UPLOAD_DIR || './uploads';
    await fs.ensureDir(uploadDir);

    // 生成输出文件路径
    const outputPath = path.join(uploadDir, `${uploadId}_${task.fileName}`);
    
    // 创建写入流
    const writeStream = fs.createWriteStream(outputPath);

    // 按顺序合并切片
    for (let i = 0; i < task.totalChunks; i++) {
      const chunk = task.chunks[i];
      const chunkData = await fs.readFile(chunk.path);
      
      await new Promise((resolve, reject) => {
        writeStream.write(chunkData, (err) => {
          if (err) reject(err);
          else resolve();
        });
      });

      console.log(`合并切片 ${i + 1}/${task.totalChunks}`);
    }

    // 关闭写入流
    await new Promise((resolve, reject) => {
      writeStream.end((err) => {
        if (err) reject(err);
        else resolve();
      });
    });

    // 验证文件大小
    const stats = await fs.stat(outputPath);
    console.log(`合并完成，文件大小: ${stats.size} bytes，预期: ${task.fileSize} bytes`);

    // 更新任务状态
    task.status = 'merged';
    task.mergedPath = outputPath;
    task.mergedSize = stats.size;
    task.updatedAt = new Date().toISOString();

    // 清理临时切片文件
    await cleanupChunks(uploadId);

    console.log(`文件合并成功: ${outputPath}`);
    return task;
  } catch (error) {
    console.error(`合并切片失败: ${uploadId}`, error);
    task.status = 'failed';
    task.error = error.message;
    task.updatedAt = new Date().toISOString();
    throw error;
  }
}

/**
 * 清理切片文件
 */
async function cleanupChunks(uploadId) {
  try {
    const tempDir = process.env.TEMP_DIR || './temp';
    const chunkDir = path.join(tempDir, uploadId);

    if (await fs.pathExists(chunkDir)) {
      await fs.remove(chunkDir);
      console.log(`清理临时文件: ${chunkDir}`);
    }
  } catch (error) {
    console.error(`清理临时文件失败: ${uploadId}`, error);
  }
}

/**
 * 取消上传任务
 */
async function cancelUpload(uploadId) {
  const task = uploadTasks.get(uploadId);
  if (!task) {
    throw new Error(`上传任务不存在: ${uploadId}`);
  }

  task.status = 'cancelled';
  task.updatedAt = new Date().toISOString();

  // 清理临时文件
  await cleanupChunks(uploadId);

  console.log(`上传任务已取消: ${uploadId}`);
  return task;
}

/**
 * 清理过期任务
 */
async function cleanupExpiredTasks() {
  const retentionHours = parseInt(process.env.FILE_RETENTION_HOURS || '24');
  const expirationTime = Date.now() - retentionHours * 60 * 60 * 1000;

  for (const [uploadId, task] of uploadTasks.entries()) {
    const taskTime = new Date(task.createdAt).getTime();
    
    if (taskTime < expirationTime) {
      console.log(`清理过期任务: ${uploadId}`);
      
      // 删除合并后的文件
      if (task.mergedPath && await fs.pathExists(task.mergedPath)) {
        await fs.remove(task.mergedPath);
      }
      
      // 删除临时文件
      await cleanupChunks(uploadId);
      
      // 从内存中移除
      uploadTasks.delete(uploadId);
    }
  }
}

// 定时清理过期任务（每小时执行一次）
setInterval(cleanupExpiredTasks, 60 * 60 * 1000);

module.exports = {
  createUploadTask,
  recordChunk,
  getUploadTask,
  mergeChunks,
  cancelUpload,
  cleanupChunks,
  cleanupExpiredTasks
};

