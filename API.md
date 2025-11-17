# FFmpeg Binary - 完整 API 接口文档

## 📋 目录

- [基础信息](#基础信息)
- [上传模块](#上传模块)
- [转换模块](#转换模块)
- [视频切割模块](#视频切割模块)
- [进度查询模块](#进度查询模块)
- [文件管理模块](#文件管理模块)
- [其他接口](#其他接口)
- [错误码说明](#错误码说明)

---

## 基础信息

### 服务配置

- **基础 URL**: `http://127.0.0.1:28888`
- **默认端口**: 28888
- **响应格式**: JSON
- **字符编码**: UTF-8

### 通用响应格式

所有接口统一返回以下格式:

```json
{
  "success": true,          // 布尔值,表示请求是否成功
  "message": "操作成功",     // 可选,操作描述信息
  "data": {                 // 可选,返回的数据对象
    // ... 具体数据
  }
}
```

---

## 上传模块

### 1. 初始化上传任务

初始化一个文件上传任务,获取 uploadId 用于后续上传切片。

**接口**: `POST /api/upload/init`

**请求头**:
```
Content-Type: application/json
```

**请求体**:
```json
{
  "fileName": "video.webm",      // 必填,文件名
  "fileSize": 10240000,          // 必填,文件总大小(字节)
  "totalChunks": 10,             // 必填,切片总数
  "chunkSize": 1024000           // 可选,每个切片大小(字节)
}
```

**响应示例**:
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

上传单个文件切片。

**接口**: `POST /api/upload/chunk`

**请求类型**: `multipart/form-data`

**FormData 字段**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file` | File | ✅ | 文件切片数据 |
| `uploadId` | String | ✅ | 上传任务 ID |
| `chunkIndex` | Number | ✅ | 切片索引(从 0 开始) |

**响应示例**:
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

**说明**:
- 当 `isComplete` 为 `true` 时,服务器会自动在后台合并所有切片
- 合并过程是异步的,需要通过状态查询接口检查合并进度

---

### 3. 查询上传状态

查询上传任务的当前状态。

**接口**: `GET /api/upload/status/:uploadId`

**URL 参数**:
- `uploadId`: 上传任务 ID

**响应示例**:
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
    "mergedPath": "/Users/ricardo/.goalfy-mediaconverter/data/550e8400-e29b-41d4-a716-446655440000.webm",
    "createdAt": "2025-11-17T10:00:00+08:00",
    "updatedAt": "2025-11-17T10:05:00+08:00"
  }
}
```

**状态说明**:
- `uploading`: 正在上传中
- `merged`: 已合并完成
- `failed`: 失败

---

### 4. 取消上传任务

取消一个上传任务并清理临时文件。

**接口**: `POST /api/upload/cancel/:uploadId`

**URL 参数**:
- `uploadId`: 上传任务 ID

**响应示例**:
```json
{
  "success": true,
  "message": "上传任务已取消"
}
```

---

## 转换模块

### 5. 开始视频转换

启动一个视频转换任务。

**接口**: `POST /api/convert/start`

**请求头**:
```
Content-Type: application/json
```

**请求体**:
```json
{
  "uploadId": "550e8400-e29b-41d4-a716-446655440000",  // uploadId 和 filePath 二选一
  "filePath": "/path/to/video.webm",                  // uploadId 和 filePath 二选一
  "outputFormat": "mp4",                              // 可选,默认 mp4
  "quality": "medium"                                 // 可选,low/medium/high,默认 medium
}
```

**参数说明**:
- `uploadId` / `filePath`: 二选一
  - `uploadId`: 引用已上传的文件
  - `filePath`: 直接指定文件路径
- `outputFormat`: 输出格式,目前支持 `mp4`
- `quality`: 转换质量
  - `low`: 快速转换,文件较小
  - `medium`: 平衡质量和速度(推荐)
  - `high`: 高质量,转换较慢

**响应示例**:
```json
{
  "success": true,
  "message": "转换任务已启动",
  "data": {
    "taskId": "task_1234567890",
    "inputPath": "/Users/ricardo/.goalfy-mediaconverter/data/video.webm",
    "outputFormat": "mp4",
    "quality": "medium"
  }
}
```

---

### 6. 查询转换状态

查询转换任务的详细状态和进度。

**接口**: `GET /api/convert/status/:taskId`

**URL 参数**:
- `taskId`: 转换任务 ID

**响应示例**:
```json
{
  "success": true,
  "data": {
    "taskId": "task_1234567890",
    "status": "processing",
    "progress": 65,
    "inputPath": "/Users/ricardo/.goalfy-mediaconverter/data/video.webm",
    "outputPath": "/Users/ricardo/.goalfy-mediaconverter/output/task_1234567890.mp4",
    "outputFormat": "mp4",
    "quality": "medium",
    "error": null,
    "createdAt": "2025-11-17T10:10:00+08:00",
    "updatedAt": "2025-11-17T10:12:00+08:00",
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

取消一个正在进行或等待中的转换任务。

**接口**: `POST /api/convert/cancel/:taskId`

**URL 参数**:
- `taskId`: 转换任务 ID

**响应示例**:
```json
{
  "success": true,
  "message": "转换任务已取消"
}
```

---

### 8. 获取转换任务列表

获取所有转换任务的列表,支持筛选和分页。

**接口**: `GET /api/convert/list`

**查询参数**:
- `status` (可选): 按状态筛选,可选值: `pending`/`processing`/`completed`/`failed`
- `limit` (可选): 返回数量限制,默认 50

**请求示例**:
```
GET /api/convert/list?status=completed&limit=20
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "tasks": [
      {
        "taskId": "task_1234567890",
        "status": "completed",
        "progress": 100,
        "inputPath": "/Users/ricardo/.goalfy-mediaconverter/data/video.webm",
        "outputPath": "/Users/ricardo/.goalfy-mediaconverter/output/task_1234567890.mp4",
        "outputFormat": "mp4",
        "quality": "medium",
        "createdAt": "2025-11-17T10:10:00+08:00",
        "completedAt": "2025-11-17T10:15:00+08:00"
      }
    ],
    "total": 1
  }
}
```

---

### 9. 下载转换后的文件

下载已完成转换的视频文件。

**接口**: `GET /api/convert/download/:taskId`

**URL 参数**:
- `taskId`: 转换任务 ID

**响应**:
- 成功时返回视频文件流 (`video/mp4`)
- 失败时返回 JSON 错误信息

**响应头**:
```
Content-Type: video/mp4
Content-Disposition: attachment; filename="task_1234567890.mp4"
```

**使用示例**:
```javascript
// 浏览器直接下载
window.location.href = 'http://127.0.0.1:28888/api/convert/download/task_1234567890';

// 或使用 fetch
const response = await fetch('http://127.0.0.1:28888/api/convert/download/task_1234567890');
const blob = await response.blob();
const url = URL.createObjectURL(blob);
```

---

## 视频切割模块

### 10. 开始视频切割

根据删除区间切割视频,生成多个片段文件。

**接口**: `POST /api/split/start`

**请求头**:
```
Content-Type: application/json
```

**请求体**:
```json
{
  "taskId": "task_1234567890",           // 必填,已转换的视频任务ID
  "deleteIntervals": [                   // 必填,要删除的时间区间数组
    { "start": 10, "end": 15 },         // 删除10-15秒
    { "start": 30, "end": 45 }          // 删除30-45秒
  ],
  "videoDuration": 60                    // 必填,视频总时长(秒)
}
```

**参数说明**:
- `taskId`: 转换任务ID,系统会查找对应的 `_converted.mp4` 文件
- `deleteIntervals`: 时间区间数组
  - `start`: 删除开始时间(秒)
  - `end`: 删除结束时间(秒)
- `videoDuration`: 视频总时长,用于计算最后保留片段

**响应示例**:
```json
{
  "success": true,
  "taskId": "task_1234567890",
  "totalSegments": 3,
  "segments": [
    {
      "success": true,
      "outputPath": "/Users/ricardo/.goalfy-mediaconverter/output/task_1234567890_part1.mp4",
      "size": 2048000,
      "duration": 10,
      "startTime": 0,
      "endTime": 10,
      "segmentIndex": 1,
      "fileName": "task_1234567890_part1.mp4",
      "originalStart": 0,
      "originalEnd": 10
    },
    {
      "success": true,
      "outputPath": "/Users/ricardo/.goalfy-mediaconverter/output/task_1234567890_part2.mp4",
      "size": 3072000,
      "duration": 15,
      "startTime": 20,
      "endTime": 35,
      "segmentIndex": 2,
      "fileName": "task_1234567890_part2.mp4",
      "originalStart": 20,
      "originalEnd": 35
    },
    {
      "success": true,
      "outputPath": "/Users/ricardo/.goalfy-mediaconverter/output/task_1234567890_part3.mp4",
      "size": 3584000,
      "duration": 15,
      "startTime": 45,
      "endTime": 60,
      "segmentIndex": 3,
      "fileName": "task_1234567890_part3.mp4",
      "originalStart": 45,
      "originalEnd": 60
    }
  ]
}
```

**说明**:
- 切割采用无损复制模式(`-c copy`),不重新编码,速度快
- 原始的 `_converted.mp4` 文件会被自动删除以节省空间
- 片段文件命名规则: `{taskId}_part{序号}.mp4`
- 支持 HTTP 流媒体播放(使用 `-movflags +faststart` 优化)

**错误响应**:
```json
{
  "success": false,
  "error": "未找到已转换的视频文件: task_xxx"
}
```

---

### 11. 下载视频片段

下载指定的视频片段文件。

**接口**: `GET /api/split/download/:taskId/:segmentIndex`

**URL 参数**:
- `taskId`: 任务ID
- `segmentIndex`: 片段索引(从1开始)

**请求示例**:
```
GET /api/split/download/task_1234567890/1
```

**响应**:
- 成功时返回视频文件流 (`video/mp4`)
- 失败时返回 JSON 错误信息

**响应头**:
```
Content-Type: video/mp4
Content-Disposition: attachment; filename="task_1234567890_part1.mp4"
Accept-Ranges: bytes
```

**说明**:
- 支持HTTP断点续传(`Accept-Ranges: bytes`)
- 可在浏览器中直接播放
- 使用流式传输,适合大文件

**使用示例**:
```javascript
// 浏览器直接下载
window.location.href = 'http://127.0.0.1:28888/api/split/download/task_1234567890/1';

// 或使用 fetch
const response = await fetch('http://127.0.0.1:28888/api/split/download/task_1234567890/1');
const blob = await response.blob();
const url = URL.createObjectURL(blob);
```

**错误响应**:
```json
{
  "success": false,
  "error": "未找到片段文件: task_xxx - part1"
}
```

---

### 12. 清理切割文件

删除指定任务的所有切割片段文件。

**接口**: `DELETE /api/split/cleanup/:taskId`

**URL 参数**:
- `taskId`: 任务ID

**请求示例**:
```
DELETE /api/split/cleanup/task_1234567890
```

**响应示例**:
```json
{
  "success": true,
  "message": "清理完成",
  "deleted": 3
}
```

**参数说明**:
- `deleted`: 成功删除的文件数量

**说明**:
- 删除所有匹配 `{taskId}_part*.mp4` 的文件
- 不会删除原始的转换文件
- 文件不存在不会报错,只返回删除数量

---

## 进度查询模块

### 13. 统一进度查询

自动识别上传任务或转换任务,返回对应的进度信息。

**接口**: `GET /api/progress/:id`

**URL 参数**:
- `id`: 任务 ID (可以是 uploadId 或 taskId)

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
    "fileName": "video.webm",
    "fileSize": 10240000,
    "createdAt": "2025-11-17T10:00:00+08:00",
    "updatedAt": "2025-11-17T10:03:00+08:00"
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
    "inputPath": "/Users/ricardo/.goalfy-mediaconverter/data/video.webm",
    "outputPath": "/Users/ricardo/.goalfy-mediaconverter/output/task_1234567890.mp4",
    "outputFormat": "mp4",
    "quality": "medium",
    "createdAt": "2025-11-17T10:10:00+08:00",
    "updatedAt": "2025-11-17T10:13:00+08:00"
  }
}
```

**说明**:
- 响应中的 `type` 字段标识任务类型 (`upload` 或 `convert`)
- 根据 `type` 字段,数据结构会有所不同

---

## 文件管理模块

### 11. 批量删除本地文件

批量删除服务器上的文件,支持删除转换后的 MP4 文件和临时文件。

**接口**: `POST /api/files/delete`

**请求头**:
```
Content-Type: application/json
```

**请求体**:
```json
{
  "filePaths": [
    "/Users/ricardo/.goalfy-mediaconverter/output/task_1234567890.mp4",
    "/Users/ricardo/.goalfy-mediaconverter/output/task_9876543210.mp4"
  ]
}
```

**参数说明**:
- `filePaths`: 文件路径数组,必须是绝对路径
- 仅允许删除以下目录下的文件:
  - `~/.goalfy-mediaconverter/output/` (转换后的文件)
  - `~/.goalfy-mediaconverter/data/` (合并后的上传文件)
  - `~/.goalfy-mediaconverter/temp/` (临时切片文件)

**响应示例**:
```json
{
  "success": true,
  "message": "处理完成: 成功 2 个,失败 0 个",
  "data": {
    "total": 2,
    "successCount": 2,
    "failCount": 0,
    "results": [
      {
        "filePath": "/Users/ricardo/.goalfy-mediaconverter/output/task_1234567890.mp4",
        "success": true,
        "message": "删除成功"
      },
      {
        "filePath": "/Users/ricardo/.goalfy-mediaconverter/output/task_9876543210.mp4",
        "success": true,
        "message": "删除成功"
      }
    ]
  }
}
```

**错误示例**:
```json
{
  "success": true,
  "message": "处理完成: 成功 1 个,失败 1 个",
  "data": {
    "total": 2,
    "successCount": 1,
    "failCount": 1,
    "results": [
      {
        "filePath": "/Users/ricardo/.goalfy-mediaconverter/output/task_1234567890.mp4",
        "success": true,
        "message": "删除成功"
      },
      {
        "filePath": "/Users/ricardo/.goalfy-mediaconverter/output/not_exist.mp4",
        "success": false,
        "message": "文件不存在"
      }
    ]
  }
}
```

**安全限制**:
- 只能删除服务配置目录 (`output`/`data`/`temp`) 下的文件
- 尝试删除其他目录的文件会返回权限错误
- 每个文件的删除结果都会单独返回

---

## 其他接口

### 12. 健康检查

检查服务是否正常运行。

**接口**: `GET /health`

**响应示例**:
```json
{
  "status": "ok",
  "timestamp": "2025-11-17T10:30:00+08:00",
  "service": "goalfy-mediaconverter",
  "version": "1.0.0"
}
```

---

### 13. 静态文件访问

直接访问输出目录中的文件。

**接口**: `GET /downloads/:filename`

**URL 参数**:
- `filename`: 文件名

**使用示例**:
```
http://127.0.0.1:28888/downloads/task_1234567890.mp4
```

**说明**:
- 直接返回文件内容
- 适用于在浏览器中预览文件
- 建议使用 `/api/convert/download/:taskId` 接口下载文件

---

## 错误码说明

### HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

### 业务错误信息

所有业务错误都会在响应的 `message` 字段中说明,常见错误如下:

**上传模块**:
- `缺少必要参数: fileName, fileSize, totalChunks` - 初始化上传时参数不完整
- `上传任务不存在` - 使用了无效的 uploadId
- `文件尚未合并完成` - 尝试在合并完成前开始转换

**转换模块**:
- `必须提供uploadId或filePath` - 开始转换时两个参数都没提供
- `输入文件不存在` - 指定的文件路径不存在
- `转换任务不存在` - 使用了无效的 taskId
- `文件尚未转换完成` - 尝试下载未完成的任务

**文件管理**:
- `缺少必要参数: filePaths` - 删除请求缺少文件路径数组
- `filePaths 不能为空` - 文件路径数组为空
- `文件不存在` - 尝试删除不存在的文件
- `无权限删除此文件` - 尝试删除不在允许目录中的文件

---

## 完整使用流程示例

### 小文件上传转换流程

```javascript
const API_BASE = 'http://127.0.0.1:28888/api';

// 1. 初始化上传
const initRes = await fetch(`${API_BASE}/upload/init`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    fileName: 'video.webm',
    fileSize: file.size,
    totalChunks: 1,
    chunkSize: file.size
  })
});
const { uploadId } = (await initRes.json()).data;

// 2. 上传文件(单个切片)
const formData = new FormData();
formData.append('file', file);
formData.append('uploadId', uploadId);
formData.append('chunkIndex', '0');

await fetch(`${API_BASE}/upload/chunk`, {
  method: 'POST',
  body: formData
});

// 3. 等待合并完成
let merged = false;
while (!merged) {
  const statusRes = await fetch(`${API_BASE}/upload/status/${uploadId}`);
  const status = await statusRes.json();
  merged = status.data.status === 'merged';
  if (!merged) await new Promise(r => setTimeout(r, 1000));
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
  if (!completed) await new Promise(r => setTimeout(r, 1000));
}

// 6. 下载文件
window.location.href = `${API_BASE}/convert/download/${taskId}`;

// 7. (可选)删除文件
await fetch(`${API_BASE}/files/delete`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    filePaths: ['/Users/ricardo/.goalfy-mediaconverter/output/task_xxx.mp4']
  })
});
```

### 大文件分片上传流程

```javascript
const API_BASE = 'http://127.0.0.1:28888/api';
const chunkSize = 1024 * 1024; // 1MB

// 1. 初始化上传
const totalChunks = Math.ceil(file.size / chunkSize);
const initRes = await fetch(`${API_BASE}/upload/init`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    fileName: file.name,
    fileSize: file.size,
    totalChunks: totalChunks,
    chunkSize: chunkSize
  })
});
const { uploadId } = (await initRes.json()).data;

// 2. 分片上传
for (let i = 0; i < totalChunks; i++) {
  const chunk = file.slice(i * chunkSize, (i + 1) * chunkSize);
  const formData = new FormData();
  formData.append('file', chunk);
  formData.append('uploadId', uploadId);
  formData.append('chunkIndex', i.toString());

  await fetch(`${API_BASE}/upload/chunk`, {
    method: 'POST',
    body: formData
  });

  console.log(`上传进度: ${((i + 1) / totalChunks * 100).toFixed(1)}%`);
}

// 3-6. 后续步骤与小文件流程相同...
```

---

## 接口总览

| 序号 | 模块 | 接口 | 方法 | 说明 |
|------|------|------|------|------|
| 1 | 上传 | `/api/upload/init` | POST | 初始化上传任务 |
| 2 | 上传 | `/api/upload/chunk` | POST | 上传文件切片 |
| 3 | 上传 | `/api/upload/status/:uploadId` | GET | 查询上传状态 |
| 4 | 上传 | `/api/upload/cancel/:uploadId` | POST | 取消上传任务 |
| 5 | 转换 | `/api/convert/start` | POST | 开始视频转换 |
| 6 | 转换 | `/api/convert/status/:taskId` | GET | 查询转换状态 |
| 7 | 转换 | `/api/convert/cancel/:taskId` | POST | 取消转换任务 |
| 8 | 转换 | `/api/convert/list` | GET | 获取转换任务列表 |
| 9 | 转换 | `/api/convert/download/:taskId` | GET | 下载转换后的文件 |
| 10 | 切割 | `/api/split/start` | POST | 开始视频切割 |
| 11 | 切割 | `/api/split/download/:taskId/:segmentIndex` | GET | 下载视频片段 |
| 12 | 切割 | `/api/split/cleanup/:taskId` | DELETE | 清理切割文件 |
| 13 | 进度 | `/api/progress/:id` | GET | 统一进度查询 |
| 14 | 文件 | `/api/files/delete` | POST | 批量删除本地文件 |
| 15 | 其他 | `/health` | GET | 健康检查 |
| 16 | 其他 | `/downloads/:filename` | GET | 静态文件访问 |

---

## 版本信息

- **当前版本**: 1.0.0
- **最后更新**: 2025-11-17
- **兼容性**: 100% 兼容 video-service (Node.js 版本)

---

## 技术支持

如有问题或建议,请提交 Issue 或查看完整文档:
- [README.md](./README.md) - 项目说明
- [QUICKSTART.md](./QUICKSTART.md) - 快速开始指南
- [examples/demo.html](./examples/demo.html) - 前端示例代码
