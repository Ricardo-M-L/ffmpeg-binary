# Goalfy 视频处理服务

## 📖 简介

这是一个基于Node.js的视频处理服务，专为Goalfy项目设计，提供以下核心功能：

1. **文件切片上传** - 支持大文件分片上传，提高上传稳定性
2. **切片自动合并** - 所有切片上传完成后自动合并为完整文件
3. **视频格式转换** - 使用FFmpeg将WebM格式转换为MP4格式
4. **实时进度查询** - 统一查询上传和转换进度

## 🚀 快速开始

### 环境要求

- **Node.js**: v14.0 或更高版本
- **FFmpeg**: 需要在系统中安装FFmpeg
  - macOS: `brew install ffmpeg`
  - Ubuntu/Debian: `sudo apt-get install ffmpeg`
  - Windows: 下载并安装 [FFmpeg](https://ffmpeg.org/download.html)

### 安装步骤

1. **安装依赖**
```bash
cd video-service
npm install
```

2. **配置环境变量**
```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置（可选）
nano .env
```

3. **启动服务**
```bash
# 开发模式（支持热重载）
npm run dev

# 生产模式
npm start
```

服务默认运行在 `http://localhost:3000`

## 📋 API接口文档

### 1. 健康检查

检查服务是否正常运行。

**接口**: `GET /health`

**响应示例**:
```json
{
  "status": "ok",
  "timestamp": "2025-01-15T10:00:00.000Z",
  "service": "goalfy-video-service",
  "version": "1.0.0"
}
```

---

### 2. 文件上传相关

#### 2.1 初始化上传任务

创建一个新的上传任务，获取上传ID。

**接口**: `POST /api/upload/init`

**请求体**:
```json
{
  "fileName": "recording.webm",
  "fileSize": 10485760,
  "totalChunks": 10,
  "chunkSize": 1048576
}
```

**参数说明**:
- `fileName`: 原始文件名
- `fileSize`: 文件总大小（字节）
- `totalChunks`: 切片总数
- `chunkSize`: 每个切片大小（字节）

**响应示例**:
```json
{
  "success": true,
  "message": "上传任务初始化成功",
  "data": {
    "uploadId": "550e8400-e29b-41d4-a716-446655440000",
    "fileName": "recording.webm",
    "totalChunks": 10
  }
}
```

#### 2.2 上传切片

上传单个文件切片。

**接口**: `POST /api/upload/chunk`

**请求类型**: `multipart/form-data`

**表单字段**:
- `file`: 文件切片（必需）
- `uploadId`: 上传任务ID（必需）
- `chunkIndex`: 切片索引，从0开始（必需）
- `chunkHash`: 切片MD5值（可选，用于校验）

**响应示例**:
```json
{
  "success": true,
  "message": "切片上传成功",
  "data": {
    "uploadId": "550e8400-e29b-41d4-a716-446655440000",
    "chunkIndex": 0,
    "uploadedChunks": 1,
    "totalChunks": 10,
    "isComplete": false
  }
}
```

**说明**: 当所有切片上传完成后（`isComplete: true`），服务器会自动开始合并文件。

#### 2.3 查询上传状态

查询上传任务的当前状态。

**接口**: `GET /api/upload/status/:uploadId`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "uploadId": "550e8400-e29b-41d4-a716-446655440000",
    "fileName": "recording.webm",
    "fileSize": 10485760,
    "totalChunks": 10,
    "uploadedChunks": 10,
    "status": "merged",
    "mergedPath": "/path/to/merged/file.webm",
    "createdAt": "2025-01-15T10:00:00.000Z",
    "updatedAt": "2025-01-15T10:05:00.000Z"
  }
}
```

**状态说明**:
- `uploading`: 正在上传切片
- `merging`: 正在合并切片
- `merged`: 合并完成
- `failed`: 失败
- `cancelled`: 已取消

#### 2.4 取消上传

取消上传任务并清理临时文件。

**接口**: `POST /api/upload/cancel/:uploadId`

**响应示例**:
```json
{
  "success": true,
  "message": "上传任务已取消"
}
```

---

### 3. 视频转换相关

#### 3.1 开始转换

启动视频格式转换任务。

**接口**: `POST /api/convert/start`

**请求体**:
```json
{
  "uploadId": "550e8400-e29b-41d4-a716-446655440000",
  "outputFormat": "mp4",
  "quality": "medium",
  "options": {
    "videoBitrate": "1000k",
    "audioBitrate": "128k",
    "fps": 30
  }
}
```

**参数说明**:
- `uploadId`: 上传任务ID（与`filePath`二选一）
- `filePath`: 直接指定文件路径（与`uploadId`二选一）
- `outputFormat`: 输出格式，默认`mp4`
- `quality`: 质量预设，可选`low`/`medium`/`high`，默认`medium`
- `options`: 可选的自定义FFmpeg参数

**质量预设说明**:

| 质量 | 视频比特率 | 音频比特率 | 编码速度 | CRF |
|-----|----------|----------|---------|-----|
| low | 500k | 64k | veryfast | 28 |
| medium | 1000k | 128k | medium | 23 |
| high | 2000k | 192k | slow | 18 |

**响应示例**:
```json
{
  "success": true,
  "message": "转换任务已启动",
  "data": {
    "taskId": "660e8400-e29b-41d4-a716-446655440001",
    "inputPath": "/path/to/input.webm",
    "outputFormat": "mp4",
    "quality": "medium"
  }
}
```

#### 3.2 查询转换状态

查询转换任务的进度和状态。

**接口**: `GET /api/convert/status/:taskId`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "taskId": "660e8400-e29b-41d4-a716-446655440001",
    "status": "processing",
    "progress": 65,
    "inputPath": "/path/to/input.webm",
    "outputPath": "/path/to/output.mp4",
    "outputFormat": "mp4",
    "quality": "medium",
    "createdAt": "2025-01-15T10:05:00.000Z",
    "updatedAt": "2025-01-15T10:07:30.000Z"
  }
}
```

**状态说明**:
- `pending`: 等待处理
- `processing`: 正在转换
- `completed`: 转换完成
- `failed`: 转换失败
- `cancelled`: 已取消

#### 3.3 取消转换

取消正在进行的转换任务。

**接口**: `POST /api/convert/cancel/:taskId`

**响应示例**:
```json
{
  "success": true,
  "message": "转换任务已取消"
}
```

#### 3.4 获取转换任务列表

查询所有转换任务。

**接口**: `GET /api/convert/list?status=completed&limit=50`

**查询参数**:
- `status`: 过滤状态（可选）
- `limit`: 返回数量限制，默认50

**响应示例**:
```json
{
  "success": true,
  "data": {
    "tasks": [
      {
        "taskId": "660e8400-e29b-41d4-a716-446655440001",
        "status": "completed",
        "progress": 100,
        "outputPath": "/path/to/output.mp4"
      }
    ],
    "total": 1
  }
}
```

---

### 4. 进度查询

#### 4.1 统一查询进度

自动识别任务类型（上传或转换）并返回进度。

**接口**: `GET /api/progress/:id`

**响应示例（上传任务）**:
```json
{
  "success": true,
  "data": {
    "type": "upload",
    "taskId": "550e8400-e29b-41d4-a716-446655440000",
    "status": "uploading",
    "progress": 70,
    "uploadedChunks": 7,
    "totalChunks": 10,
    "fileName": "recording.webm"
  }
}
```

**响应示例（转换任务）**:
```json
{
  "success": true,
  "data": {
    "type": "convert",
    "taskId": "660e8400-e29b-41d4-a716-446655440001",
    "status": "processing",
    "progress": 45,
    "outputPath": "/path/to/output.mp4"
  }
}
```

#### 4.2 批量查询进度

一次查询多个任务的进度。

**接口**: `GET /api/progress/batch?ids=id1,id2,id3`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "tasks": [
      {
        "type": "upload",
        "taskId": "id1",
        "status": "merged",
        "progress": 100
      },
      {
        "type": "convert",
        "taskId": "id2",
        "status": "processing",
        "progress": 30
      }
    ],
    "total": 2
  }
}
```

---

## 💡 使用示例

### JavaScript/Node.js客户端示例

完整的客户端实现请查看 `examples/client-example.js`

```javascript
const fs = require('fs');
const FormData = require('form-data');
const axios = require('axios');

const API_BASE = 'http://localhost:3000/api';

// 1. 初始化上传
async function uploadFile(filePath) {
  const fileSize = fs.statSync(filePath).size;
  const chunkSize = 1024 * 1024; // 1MB per chunk
  const totalChunks = Math.ceil(fileSize / chunkSize);

  // 初始化上传任务
  const initRes = await axios.post(`${API_BASE}/upload/init`, {
    fileName: path.basename(filePath),
    fileSize,
    totalChunks,
    chunkSize
  });

  const uploadId = initRes.data.data.uploadId;

  // 上传每个切片
  for (let i = 0; i < totalChunks; i++) {
    const start = i * chunkSize;
    const end = Math.min(start + chunkSize, fileSize);
    const chunk = fs.createReadStream(filePath, { start, end: end - 1 });

    const formData = new FormData();
    formData.append('file', chunk);
    formData.append('uploadId', uploadId);
    formData.append('chunkIndex', i);

    await axios.post(`${API_BASE}/upload/chunk`, formData, {
      headers: formData.getHeaders()
    });

    console.log(`已上传切片 ${i + 1}/${totalChunks}`);
  }

  return uploadId;
}

// 2. 开始转换
async function convertVideo(uploadId) {
  const res = await axios.post(`${API_BASE}/convert/start`, {
    uploadId,
    outputFormat: 'mp4',
    quality: 'medium'
  });

  return res.data.data.taskId;
}

// 3. 查询进度
async function checkProgress(taskId) {
  const res = await axios.get(`${API_BASE}/progress/${taskId}`);
  return res.data.data;
}

// 完整流程
async function main() {
  // 上传文件
  const uploadId = await uploadFile('./video.webm');
  console.log('上传完成:', uploadId);

  // 等待合并
  await new Promise(resolve => setTimeout(resolve, 2000));

  // 开始转换
  const taskId = await convertVideo(uploadId);
  console.log('转换已启动:', taskId);

  // 轮询进度
  while (true) {
    const progress = await checkProgress(taskId);
    console.log(`进度: ${progress.progress}%`);

    if (progress.status === 'completed') {
      console.log('转换完成!', progress.outputPath);
      break;
    }

    await new Promise(resolve => setTimeout(resolve, 1000));
  }
}

main();
```

### 浏览器端示例

```html
<!DOCTYPE html>
<html>
<head>
  <title>视频上传转换</title>
</head>
<body>
  <input type="file" id="fileInput" accept="video/webm">
  <button onclick="handleUpload()">上传并转换</button>
  <div id="progress"></div>

  <script>
    const API_BASE = 'http://localhost:3000/api';

    async function handleUpload() {
      const file = document.getElementById('fileInput').files[0];
      if (!file) return;

      const chunkSize = 1024 * 1024; // 1MB
      const totalChunks = Math.ceil(file.size / chunkSize);

      // 1. 初始化上传
      const initRes = await fetch(`${API_BASE}/upload/init`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          fileName: file.name,
          fileSize: file.size,
          totalChunks,
          chunkSize
        })
      });

      const { uploadId } = (await initRes.json()).data;

      // 2. 上传切片
      for (let i = 0; i < totalChunks; i++) {
        const start = i * chunkSize;
        const end = Math.min(start + chunkSize, file.size);
        const chunk = file.slice(start, end);

        const formData = new FormData();
        formData.append('file', chunk);
        formData.append('uploadId', uploadId);
        formData.append('chunkIndex', i);

        await fetch(`${API_BASE}/upload/chunk`, {
          method: 'POST',
          body: formData
        });

        updateProgress(`上传进度: ${Math.round((i + 1) / totalChunks * 100)}%`);
      }

      // 3. 等待合并
      updateProgress('文件合并中...');
      await new Promise(resolve => setTimeout(resolve, 2000));

      // 4. 开始转换
      const convertRes = await fetch(`${API_BASE}/convert/start`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          uploadId,
          outputFormat: 'mp4',
          quality: 'medium'
        })
      });

      const { taskId } = (await convertRes.json()).data;

      // 5. 轮询转换进度
      const interval = setInterval(async () => {
        const progressRes = await fetch(`${API_BASE}/progress/${taskId}`);
        const progress = (await progressRes.json()).data;

        updateProgress(`转换进度: ${progress.progress}%`);

        if (progress.status === 'completed') {
          clearInterval(interval);
          updateProgress('转换完成！');
          
          // 下载文件
          const downloadUrl = `http://localhost:3000/downloads/${taskId}_converted.mp4`;
          window.open(downloadUrl);
        }
      }, 1000);
    }

    function updateProgress(text) {
      document.getElementById('progress').textContent = text;
    }
  </script>
</body>
</html>
```

---

## ⚙️ 配置说明

### 环境变量

在 `.env` 文件中配置：

```bash
# 服务端口
PORT=3000

# 文件存储路径
UPLOAD_DIR=./uploads      # 合并后的文件存储目录
TEMP_DIR=./temp           # 临时切片存储目录
OUTPUT_DIR=./output       # 转换后的文件输出目录

# 单个切片最大大小（字节）
MAX_CHUNK_SIZE=10485760   # 默认10MB

# 文件保留时间（小时）
FILE_RETENTION_HOURS=24   # 24小时后自动清理

# FFmpeg路径（可选）
# 如果FFmpeg不在系统PATH中，请指定完整路径
# FFMPEG_PATH=/usr/local/bin/ffmpeg
```

### 自定义转换参数

在调用转换接口时，可以传入自定义的FFmpeg参数：

```json
{
  "uploadId": "xxx",
  "outputFormat": "mp4",
  "options": {
    "videoCodec": "libx264",
    "audioCodec": "aac",
    "videoBitrate": "2000k",
    "audioBitrate": "192k",
    "fps": 30,
    "preset": "slow",
    "crf": 18,
    "customOptions": [
      "-profile:v", "high",
      "-level", "4.0"
    ]
  }
}
```

---

## 🔧 故障排除

### 1. FFmpeg未找到

**错误**: `Error: ffmpeg was killed with signal SIGKILL`

**解决方案**:
```bash
# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt-get update
sudo apt-get install ffmpeg

# 验证安装
ffmpeg -version
```

### 2. 端口被占用

**错误**: `Error: listen EADDRINUSE: address already in use :::3000`

**解决方案**:
```bash
# 修改.env文件中的PORT配置
PORT=3001
```

### 3. 磁盘空间不足

定期清理过期文件，或减少 `FILE_RETENTION_HOURS` 的值。

### 4. 上传大文件失败

增加 `MAX_CHUNK_SIZE` 或使用更小的切片大小。

---

## 📊 性能建议

1. **切片大小**: 建议1-5MB，太小会增加请求数，太大可能导致内存问题
2. **质量选择**: 
   - `low`: 适合预览或临时使用
   - `medium`: 平衡质量和文件大小（推荐）
   - `high`: 高质量输出，文件较大
3. **并发控制**: 避免同时处理过多转换任务，可能导致系统资源耗尽
4. **存储清理**: 生产环境建议设置自动清理任务

---

## 🔒 安全建议

1. **添加认证**: 在生产环境中添加JWT或API Key认证
2. **文件类型验证**: 验证上传的文件类型和大小
3. **速率限制**: 使用express-rate-limit防止滥用
4. **CORS配置**: 限制允许访问的域名

---

## 📝 许可证

MIT License

---

## 📞 技术支持

如有问题或建议，请联系：
- 邮箱: support@goalfylearning.com
- GitHub Issues: [提交问题](https://github.com/your-repo/issues)

---

*最后更新: 2025年1月*


