# FFmpeg Binary Service

一个本地运行的 WebM 到 MP4 视频转换服务,支持同步和异步转换模式。

## 功能特性

- ✅ **同步转换**: 直接接收 WebM 流,实时返回 MP4 流
- ✅ **异步转换**: 分片上传大文件,后台处理,支持进度查询
- ✅ **智能端口**: 自动选择 18888-28888 范围内的可用端口
- ✅ **开机自启**: 支持 macOS/Windows/Linux 自启动
- ✅ **本地服务**: 仅监听 127.0.0.1,安全可靠

## 快速开始

### 开发环境运行

```bash
# 安装依赖
go mod download

# 运行服务
go run main.go

# 服务会自动选择可用端口并启动
# 查看日志获取实际端口号
```

### 生产环境部署

#### macOS

```bash
# 构建 DMG 安装包
./build-macos.sh

# 安装
# 1. 打开 build/macos/FFmpeg-Binary-Installer.dmg
# 2. 将应用拖到 Applications 文件夹
# 3. 运行应用安装自启动:
/Applications/FFmpeg-Binary.app/Contents/MacOS/ffmpeg-binary install
```

#### Windows

```bash
# 构建 Windows 可执行文件
build-windows.bat

# 安装
# 1. 复制 ffmpeg-binary.exe 到 C:\Program Files\FFmpeg-Binary\
# 2. 运行 install.bat 安装自启动
```

## API 接口文档

### 1. 同步转换接口

直接将 WebM 流转换为 MP4 流返回。

**接口**: `POST /api/v1/convert/sync`

**请求**:
- Content-Type: `video/webm`
- Body: WebM 视频流

**响应**:
- Content-Type: `video/mp4`
- Body: MP4 视频流

**示例**:
```bash
curl -X POST http://127.0.0.1:18888/api/v1/convert/sync \
  -H "Content-Type: video/webm" \
  --data-binary @input.webm \
  -o output.mp4
```

**前端示例**:
```javascript
const response = await fetch('http://127.0.0.1:18888/api/v1/convert/sync', {
  method: 'POST',
  headers: {
    'Content-Type': 'video/webm'
  },
  body: webmBlob
});

const mp4Blob = await response.blob();
```

---

### 2. 创建异步转换任务

创建转换任务并返回任务 ID。

**接口**: `POST /api/v1/convert/async`

**响应**:
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "message": "任务已创建,请上传视频分片",
  "upload_url": "/api/v1/convert/async/550e8400-e29b-41d4-a716-446655440000/chunk",
  "status_url": "/api/v1/task/550e8400-e29b-41d4-a716-446655440000",
  "download_url": "/api/v1/task/550e8400-e29b-41d4-a716-446655440000/download"
}
```

---

### 3. 上传视频分片

分片上传 WebM 视频数据。

**接口**: `POST /api/v1/convert/async/:task_id/chunk`

**请求头**:
- `X-Last-Chunk: true` (最后一个分片时设置)

**请求体**: 视频分片数据

**响应**:
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "written": 1048576,
  "is_last": false
}
```

**前端分片上传示例**:
```javascript
// 创建任务
const createResp = await fetch('http://127.0.0.1:18888/api/v1/convert/async', {
  method: 'POST'
});
const { task_id, upload_url } = await createResp.json();

// 分片上传
const chunkSize = 1024 * 1024; // 1MB
const totalChunks = Math.ceil(file.size / chunkSize);

for (let i = 0; i < totalChunks; i++) {
  const start = i * chunkSize;
  const end = Math.min(start + chunkSize, file.size);
  const chunk = file.slice(start, end);
  const isLast = i === totalChunks - 1;

  await fetch(`http://127.0.0.1:18888${upload_url}`, {
    method: 'POST',
    headers: {
      'X-Last-Chunk': isLast ? 'true' : 'false'
    },
    body: chunk
  });
}
```

---

### 4. 查询任务状态

查询转换任务的状态和进度。

**接口**: `GET /api/v1/task/:task_id`

**响应**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing",
  "progress": 45,
  "error": "",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:30Z"
}
```

**状态说明**:
- `pending`: 等待处理
- `processing`: 转换中
- `completed`: 转换完成
- `failed`: 转换失败

**前端轮询示例**:
```javascript
async function pollTaskStatus(taskId) {
  const checkStatus = async () => {
    const resp = await fetch(`http://127.0.0.1:18888/api/v1/task/${taskId}`);
    const data = await resp.json();

    console.log(`进度: ${data.progress}%`);

    if (data.status === 'completed') {
      return data;
    } else if (data.status === 'failed') {
      throw new Error(data.error);
    }

    // 继续轮询
    await new Promise(resolve => setTimeout(resolve, 1000));
    return checkStatus();
  };

  return checkStatus();
}
```

---

### 5. 下载转换后的视频

下载转换完成的 MP4 视频。

**接口**: `GET /api/v1/task/:task_id/download`

**响应**:
- Content-Type: `video/mp4`
- Body: MP4 视频流

**示例**:
```bash
curl -o output.mp4 http://127.0.0.1:18888/api/v1/task/550e8400-e29b-41d4-a716-446655440000/download
```

**前端下载示例**:
```javascript
const resp = await fetch(`http://127.0.0.1:18888/api/v1/task/${taskId}/download`);
const blob = await resp.blob();
const url = URL.createObjectURL(blob);

// 触发下载
const a = document.createElement('a');
a.href = url;
a.download = 'converted.mp4';
a.click();
```

---

### 6. 删除任务

删除任务及相关文件。

**接口**: `DELETE /api/v1/task/:task_id`

**响应**:
```json
{
  "message": "任务已删除"
}
```

---

### 7. 列出所有任务

获取所有转换任务列表。

**接口**: `GET /api/v1/tasks`

**响应**:
```json
{
  "tasks": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "status": "completed",
      "progress": 100,
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:01:00Z"
    }
  ],
  "total": 1
}
```

---

### 8. 健康检查

检查服务是否正常运行。

**接口**: `GET /health`

**响应**:
```json
{
  "status": "ok",
  "port": 18888
}
```

## 完整前端使用示例

```javascript
class FFmpegConverter {
  constructor(baseUrl = 'http://127.0.0.1:18888') {
    this.baseUrl = baseUrl;
  }

  // 同步转换
  async convertSync(webmBlob) {
    const resp = await fetch(`${this.baseUrl}/api/v1/convert/sync`, {
      method: 'POST',
      headers: { 'Content-Type': 'video/webm' },
      body: webmBlob
    });
    return await resp.blob();
  }

  // 异步转换(分片上传)
  async convertAsync(file, onProgress) {
    // 创建任务
    const createResp = await fetch(`${this.baseUrl}/api/v1/convert/async`, {
      method: 'POST'
    });
    const { task_id, upload_url } = await createResp.json();

    // 分片上传
    const chunkSize = 1024 * 1024;
    const totalChunks = Math.ceil(file.size / chunkSize);

    for (let i = 0; i < totalChunks; i++) {
      const chunk = file.slice(i * chunkSize, (i + 1) * chunkSize);
      const isLast = i === totalChunks - 1;

      await fetch(`${this.baseUrl}${upload_url}`, {
        method: 'POST',
        headers: { 'X-Last-Chunk': isLast ? 'true' : 'false' },
        body: chunk
      });

      onProgress?.({ uploaded: i + 1, total: totalChunks });
    }

    // 轮询状态
    while (true) {
      const statusResp = await fetch(`${this.baseUrl}/api/v1/task/${task_id}`);
      const status = await statusResp.json();

      onProgress?.({ status: status.status, progress: status.progress });

      if (status.status === 'completed') {
        // 下载结果
        const downloadResp = await fetch(
          `${this.baseUrl}/api/v1/task/${task_id}/download`
        );
        return await downloadResp.blob();
      } else if (status.status === 'failed') {
        throw new Error(status.error);
      }

      await new Promise(resolve => setTimeout(resolve, 1000));
    }
  }
}

// 使用示例
const converter = new FFmpegConverter();

// 同步转换
const mp4Blob = await converter.convertSync(webmBlob);

// 异步转换
const mp4Blob = await converter.convertAsync(file, ({ progress, status }) => {
  console.log(`${status}: ${progress}%`);
});
```

## 配置文件

配置文件位置: `~/.ffmpeg-binary/config.json`

```json
{
  "port": 18888,
  "host": "127.0.0.1",
  "data_dir": "~/.ffmpeg-binary/data",
  "ffmpeg_path": "/usr/local/bin/ffmpeg"
}
```

## 自启动管理

```bash
# 安装自启动
./ffmpeg-binary install

# 卸载自启动
./ffmpeg-binary uninstall
```

## 系统要求

- Go 1.23+
- FFmpeg 4.0+ (需单独安装)

### FFmpeg 安装

**macOS**:
```bash
brew install ffmpeg
```

**Windows**:
下载地址: https://www.gyan.dev/ffmpeg/builds/

**Linux**:
```bash
sudo apt install ffmpeg  # Ubuntu/Debian
sudo yum install ffmpeg  # CentOS/RHEL
```

## 开发

```bash
# 运行测试
go test ./...

# 构建
go build -o ffmpeg-binary .

# 交叉编译
GOOS=darwin GOARCH=amd64 go build -o ffmpeg-binary-darwin .
GOOS=windows GOARCH=amd64 go build -o ffmpeg-binary.exe .
GOOS=linux GOARCH=amd64 go build -o ffmpeg-binary-linux .
```

## 许可证

MIT License