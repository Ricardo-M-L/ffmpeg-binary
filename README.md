# FFmpeg Binary Service

一个本地运行的视频处理服务,支持文件切片上传、WebM 到 MP4 转换和实时进度查询。

## 🌟 功能特性

- ✅ **文件切片上传**: 支持大文件分片上传,自动合并
- ✅ **视频格式转换**: WebM → MP4 转换,支持多种质量选项
- ✅ **实时进度查询**: 统一的进度查询接口
- ✅ **固定端口**: 使用固定端口 28888
- ✅ **开机自启**: 支持 macOS/Windows/Linux 自启动
- ✅ **本地服务**: 仅监听 127.0.0.1,安全可靠
- ✅ **完全兼容**: 接口 100% 兼容 video-service (Node.js 版本)
- 🆕 **FFmpeg 自动安装**: 自动检测并安装 FFmpeg,无需手动配置

## 🚀 快速开始

> 🆕 **无需手动安装 FFmpeg!** 服务会自动检测并安装 FFmpeg,详见 [FFmpeg 自动安装说明](./FFMPEG_AUTO_INSTALL.md)

### 开发环境运行

```bash
# 安装依赖
go mod download

# 运行服务(会自动检测/安装 FFmpeg)
go run main.go

# 服务启动在 http://127.0.0.1:28888
```

**首次启动日志示例:**

```
✅ FFmpeg 已安装: /opt/homebrew/bin/ffmpeg
===========================================
🚀 FFmpeg Binary 服务启动成功!
===========================================
📡 服务地址: http://127.0.0.1:28888
...
```

或如果未安装:

```
⚠️  FFmpeg 未安装或不可用,正在自动安装...
📦 正在通过 Homebrew 安装 FFmpeg...
[安装进度...]
✅ FFmpeg 安装成功: /opt/homebrew/bin/ffmpeg
```

### 生产环境部署

#### macOS

```bash
# 构建 DMG 安装包
./scripts/build-macos-dmg.sh

# 安装
# 1. 打开 build/macos/FFmpeg-Binary-Installer.dmg
# 2. 将应用拖到 Applications 文件夹
# 3. 运行应用安装自启动:
/Applications/FFmpeg-Binary.app/Contents/MacOS/ffmpeg-binary install
```

#### Windows

```bash
# 构建 Windows 可执行文件
./scripts/build-windows.bat

# 安装
# 1. 复制 ffmpeg-binary.exe 到 C:\Program Files\FFmpeg-Binary\
# 2. 运行 install.bat 安装自启动
```

---

## 📡 API 接口文档

### 基础信息

- **基础URL**: `http://127.0.0.1:28888`
- **默认端口**: 28888
- **响应格式**: JSON

---

## 📤 上传模块 (`/api/upload`)

### 1. 初始化上传任务

**接口**: `POST /api/upload/init`

**请求体**:
```json
{
  "fileName": "video.webm",
  "fileSize": 10240000,
  "totalChunks": 10,
  "chunkSize": 1024000
}
```

**响应**:
```json
{
  "success": true,
  "message": "上传任务初始化成功",
  "data": {
    "uploadId": "550e8400-e29b-41d4-a716-446655440000",
    "fileName": "video.webm",
    "totalChunks": 10
  }
}
```

---

### 2. 上传文件切片

**接口**: `POST /api/upload/chunk`

**请求类型**: `multipart/form-data`

**FormData 字段**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file` | File | ✅ | 文件切片 |
| `uploadId` | String | ✅ | 上传任务ID |
| `chunkIndex` | Number | ✅ | 切片索引(从0开始) |

**响应**:
```json
{
  "success": true,
  "message": "切片上传成功",
  "data": {
    "uploadId": "550e8400-e29b-41d4-a716-446655440000",
    "chunkIndex": 5,
    "uploadedChunks": 6,
    "totalChunks": 10,
    "isComplete": false
  }
}
```

---

### 3. 查询上传状态

**接口**: `GET /api/upload/status/:uploadId`

**响应**:
```json
{
  "success": true,
  "data": {
    "uploadId": "550e8400-e29b-41d4-a716-446655440000",
    "fileName": "video.webm",
    "fileSize": 10240000,
    "totalChunks": 10,
    "uploadedChunks": 10,
    "status": "merged",
    "mergedPath": "/path/to/merged/file.webm",
    "createdAt": "2025-11-16T15:00:00Z",
    "updatedAt": "2025-11-16T15:05:00Z"
  }
}
```

**状态说明**:
- `uploading`: 正在上传中
- `merged`: 已合并完成
- `failed`: 失败

---

### 4. 取消上传任务

**接口**: `POST /api/upload/cancel/:uploadId`

**响应**:
```json
{
  "success": true,
  "message": "上传任务已取消"
}
```

---

## 🎬 转换模块 (`/api/convert`)

### 5. 开始视频转换

**接口**: `POST /api/convert/start`

**请求体**:
```json
{
  "uploadId": "550e8400-e29b-41d4-a716-446655440000",
  "outputFormat": "mp4",
  "quality": "medium"
}
```

**参数说明**:
- `uploadId` / `filePath`: 二选一
  - `uploadId`: 引用已上传的文件
  - `filePath`: 直接指定文件路径
- `outputFormat`: 输出格式,默认 `mp4`
- `quality`: 质量 `low`/`medium`/`high`,默认 `medium`

**响应**:
```json
{
  "success": true,
  "message": "转换任务已启动",
  "data": {
    "taskId": "task_1234567890",
    "inputPath": "/uploads/video.webm",
    "outputFormat": "mp4",
    "quality": "medium"
  }
}
```

---

### 6. 查询转换状态

**接口**: `GET /api/convert/status/:taskId`

**响应**:
```json
{
  "success": true,
  "data": {
    "taskId": "task_1234567890",
    "status": "processing",
    "progress": 65,
    "inputPath": "/uploads/video.webm",
    "outputPath": "/output/video.mp4",
    "outputFormat": "mp4",
    "quality": "medium",
    "error": null,
    "createdAt": "2025-11-16T15:10:00Z",
    "updatedAt": "2025-11-16T15:12:00Z",
    "completedAt": null
  }
}
```

**状态说明**:
- `pending`: 等待开始
- `processing`: 转换中
- `completed`: 转换完成
- `failed`: 转换失败

---

### 7. 取消转换任务

**接口**: `POST /api/convert/cancel/:taskId`

**响应**:
```json
{
  "success": true,
  "message": "转换任务已取消"
}
```

---

### 8. 获取转换任务列表

**接口**: `GET /api/convert/list?status=completed&limit=20`

**查询参数**:
- `status` (可选): 按状态筛选
- `limit` (可选): 返回数量限制,默认 50

**响应**:
```json
{
  "success": true,
  "data": {
    "tasks": [
      {
        "taskId": "task_xxx",
        "status": "completed",
        "progress": 100,
        "outputPath": "/output/video.mp4"
      }
    ],
    "total": 1
  }
}
```

---

### 9. 下载转换后的文件

**接口**: `GET /api/convert/download/:taskId`

**响应**: 视频文件流 (`video/mp4`)

**示例**:
```javascript
window.location.href = 'http://127.0.0.1:28888/api/convert/download/task_1234567890';
```

---

## 📊 进度查询模块 (`/api/progress`)

### 10. 统一进度查询

**接口**: `GET /api/progress/:id`

**说明**: 自动识别上传任务或转换任务

**上传任务响应**:
```json
{
  "success": true,
  "data": {
    "type": "upload",
    "taskId": "550e8400-e29b-41d4-a716-446655440000",
    "status": "uploading",
    "progress": 60,
    "uploadedChunks": 6,
    "totalChunks": 10,
    "fileName": "video.webm"
  }
}
```

**转换任务响应**:
```json
{
  "success": true,
  "data": {
    "type": "convert",
    "taskId": "task_1234567890",
    "status": "processing",
    "progress": 75,
    "outputPath": "/output/video.mp4"
  }
}
```

---

## 🏥 其他接口

### 健康检查

**接口**: `GET /health`

**响应**:
```json
{
  "status": "ok",
  "timestamp": "2025-11-16T15:30:00Z",
  "service": "ffmpeg-binary",
  "version": "1.0.0"
}
```

### 静态文件访问

**接口**: `GET /downloads/:filename`

**说明**: 直接访问输出目录中的文件

---

## 💻 使用示例

### 完整流程: 上传 → 转换 → 下载

```javascript
const API_BASE = 'http://127.0.0.1:28888/api';

// 1. 初始化上传
const initRes = await fetch(`${API_BASE}/upload/init`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    fileName: 'video.webm',
    fileSize: file.size,
    totalChunks: Math.ceil(file.size / chunkSize),
    chunkSize: chunkSize
  })
});
const { uploadId } = (await initRes.json()).data;

// 2. 上传切片
for (let i = 0; i < totalChunks; i++) {
  const chunk = file.slice(i * chunkSize, (i + 1) * chunkSize);
  const formData = new FormData();
  formData.append('file', chunk);
  formData.append('uploadId', uploadId);
  formData.append('chunkIndex', i);

  await fetch(`${API_BASE}/upload/chunk`, {
    method: 'POST',
    body: formData
  });
}

// 3. 等待合并
let merged = false;
while (!merged) {
  const statusRes = await fetch(`${API_BASE}/upload/status/${uploadId}`);
  const status = await statusRes.json();
  merged = status.data.status === 'merged';
  await new Promise(r => setTimeout(r, 1000));
}

// 4. 开始转换
const convertRes = await fetch(`${API_BASE}/convert/start`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    uploadId: uploadId,
    outputFormat: 'mp4',
    quality: 'medium'
  })
});
const { taskId } = (await convertRes.json()).data;

// 5. 轮询转换进度
let completed = false;
while (!completed) {
  const progressRes = await fetch(`${API_BASE}/progress/${taskId}`);
  const progress = await progressRes.json();
  console.log(`进度: ${progress.data.progress}%`);
  completed = progress.data.status === 'completed';
  await new Promise(r => setTimeout(r, 1000));
}

// 6. 下载文件
window.location.href = `${API_BASE}/convert/download/${taskId}`;
```

---

## 🔧 配置说明

### 环境变量

服务运行时会自动创建以下目录:

```bash
~/.ffmpeg-binary/
├── data/      # 合并后的文件
├── temp/      # 临时切片文件
├── output/    # 转换后的输出文件
└── config.json # 配置文件
```

### 配置文件

配置文件位置: `~/.ffmpeg-binary/config.json`

```json
{
  "port": 28888,
  "host": "127.0.0.1",
  "data_dir": "~/.ffmpeg-binary/data",
  "temp_dir": "~/.ffmpeg-binary/temp",
  "output_dir": "~/.ffmpeg-binary/output",
  "ffmpeg_path": "/usr/local/bin/ffmpeg"
}
```

---

## 📦 项目结构

```
ffmpeg-binary/
├── main.go                      # 入口文件
├── internal/
│   ├── config/                  # 配置管理
│   ├── converter/               # FFmpeg 转换器
│   ├── installer/               # 🆕 FFmpeg 自动安装器
│   ├── task/                    # 转换任务管理
│   ├── upload/                  # 上传任务管理
│   ├── server/                  # HTTP 服务器
│   │   ├── server.go           # 路由配置
│   │   └── handlers.go         # 接口处理器
│   └── autostart/              # 自启动管理
├── examples/
│   └── demo.html               # 前端示例
└── scripts/                     # 构建脚本
```

---

## 🔗 相关链接

- [FFmpeg 自动安装说明](./FFMPEG_AUTO_INSTALL.md) 🆕
- [快速开始指南](./QUICKSTART.md)
- [API 接口文档](./API.md)
- [接口测试示例](./examples/demo.html)

---

## 📝 许可证

MIT License

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request!
