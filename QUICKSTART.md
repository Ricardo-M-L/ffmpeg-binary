# 🚀 快速开始指南

这是一个本地运行的视频处理服务,提供文件上传、WebM 到 MP4 转换和实时进度查询功能。

---

## 第一步: 安装 FFmpeg

### macOS
```bash
brew install ffmpeg
```

### Windows
1. 访问 [FFmpeg 官网](https://ffmpeg.org/download.html)
2. 下载 "ffmpeg-release-essentials.zip"
3. 解压到 `C:\ffmpeg\`
4. 添加 `C:\ffmpeg\bin` 到系统 PATH

### Linux (Ubuntu/Debian)
```bash
sudo apt update
sudo apt install ffmpeg
```

验证安装:
```bash
ffmpeg -version
```

---

## 第二步: 运行服务

### 开发模式
```bash
cd /path/to/ffmpeg-binary
go run main.go
```

### 生产模式
```bash
# 编译
go build -o ffmpeg-binary

# 运行
./ffmpeg-binary
```

服务启动后会显示:
```
===========================================
🚀 FFmpeg Binary 服务启动成功!
===========================================
📡 服务地址: http://127.0.0.1:28888
📝 健康检查: http://127.0.0.1:28888/health
📂 数据目录: ~/.ffmpeg-binary/data
📂 临时目录: ~/.ffmpeg-binary/temp
📂 输出目录: ~/.ffmpeg-binary/output
===========================================
```

---

## 第三步: 测试服务

### 健康检查
```bash
curl http://127.0.0.1:28888/health
```

预期输出:
```json
{
  "status": "ok",
  "timestamp": "2025-11-16T15:30:00Z",
  "service": "ffmpeg-binary",
  "version": "1.0.0"
}
```

### 使用前端示例
```bash
# 在浏览器打开
open examples/demo.html
```

---

## 🎬 快速使用示例

### 方式一: 小文件上传转换 (< 10MB)

适用于小视频文件,简单快速。

```bash
# 1. 初始化上传
curl -X POST http://127.0.0.1:28888/api/upload/init \
  -H "Content-Type: application/json" \
  -d '{
    "fileName": "video.webm",
    "fileSize": 5242880,
    "totalChunks": 1,
    "chunkSize": 5242880
  }'

# 响应示例:
# {"success":true,"data":{"uploadId":"550e8400-e29b-41d4-a716-446655440000","fileName":"video.webm","totalChunks":1}}

# 2. 上传文件(单个切片)
curl -X POST http://127.0.0.1:28888/api/upload/chunk \
  -F "file=@video.webm" \
  -F "uploadId=550e8400-e29b-41d4-a716-446655440000" \
  -F "chunkIndex=0"

# 3. 等待合并完成(自动后台进行)
sleep 2

# 4. 开始转换
curl -X POST http://127.0.0.1:28888/api/convert/start \
  -H "Content-Type: application/json" \
  -d '{
    "uploadId": "550e8400-e29b-41d4-a716-446655440000",
    "outputFormat": "mp4",
    "quality": "medium"
  }'

# 响应示例:
# {"success":true,"data":{"taskId":"task_1234567890","inputPath":"/path/to/file","quality":"medium"}}

# 5. 查询进度
curl http://127.0.0.1:28888/api/progress/task_1234567890

# 6. 下载转换后的文件
curl http://127.0.0.1:28888/api/convert/download/task_1234567890 -o output.mp4
```

### 方式二: 大文件分片上传 (> 10MB)

适用于大视频文件,分片上传更可靠。

```javascript
const API_BASE = 'http://127.0.0.1:28888/api';
const chunkSize = 1024 * 1024; // 1MB 每片

// 1. 初始化上传
const initRes = await fetch(`${API_BASE}/upload/init`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    fileName: file.name,
    fileSize: file.size,
    totalChunks: Math.ceil(file.size / chunkSize),
    chunkSize: chunkSize
  })
});
const { uploadId } = (await initRes.json()).data;

// 2. 分片上传
const totalChunks = Math.ceil(file.size / chunkSize);
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

  console.log(`上传进度: ${((i + 1) / totalChunks * 100).toFixed(1)}%`);
}

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
  console.log(`转换进度: ${progress.data.progress}%`);

  completed = progress.data.status === 'completed';
  if (progress.data.status === 'failed') {
    console.error('转换失败');
    break;
  }

  if (!completed) await new Promise(r => setTimeout(r, 1000));
}

// 6. 下载文件
window.location.href = `${API_BASE}/convert/download/${taskId}`;
```

---

## 📦 打包部署

### macOS DMG
```bash
./scripts/build-macos-dmg.sh

# 输出: build/macos/FFmpeg-Binary-Installer.dmg
# 1. 打开 DMG 文件
# 2. 将应用拖到 Applications 文件夹
# 3. 运行应用安装自启动
/Applications/FFmpeg-Binary.app/Contents/MacOS/ffmpeg-binary install
```

### Windows 安装包
```bash
# 在 Windows 上运行
./scripts/build-windows.bat

# 输出: build/windows/ffmpeg-binary.exe
# 1. 复制 exe 到 C:\Program Files\FFmpeg-Binary\
# 2. 运行 install.bat 安装自启动
```

---

## 🔍 故障排查

### 问题 1: 服务无法启动

**检查 FFmpeg 是否安装**:
```bash
ffmpeg -version
```

**检查端口 28888 是否被占用**:
```bash
# macOS/Linux
lsof -i :28888

# Windows
netstat -ano | findstr :28888
```

**解决方案**:
- 如果端口被占用,关闭占用进程或修改配置文件 `~/.ffmpeg-binary/config.json` 中的 `port` 字段

### 问题 2: 上传失败

**常见原因**:
1. 文件太大超过服务器限制
2. 磁盘空间不足
3. 权限问题

**检查磁盘空间**:
```bash
df -h ~/.ffmpeg-binary
```

**查看日志**:
- 服务运行窗口会显示详细日志
- 检查错误信息中的具体原因

### 问题 3: 转换失败

**排查步骤**:
1. 确认上传的文件是 WebM 格式
2. 检查文件是否完整(未损坏)
3. 查看任务状态获取详细错误信息:
   ```bash
   curl http://127.0.0.1:28888/api/convert/status/YOUR_TASK_ID
   ```

**常见错误**:
- `文件尚未合并完成`: 需要等待上传的所有切片合并完成
- `输入文件不存在`: 上传任务可能已被清理,需要重新上传
- `FFmpeg 转换失败`: 检查 FFmpeg 安装和文件格式

### 问题 4: 前端跨域错误

服务已启用 CORS,如果仍有问题:

1. **确认使用正确的地址**: 使用 `http://127.0.0.1:28888` 而不是 `localhost`
2. **检查浏览器控制台**: 查看具体的 CORS 错误信息
3. **清除浏览器缓存**: 硬刷新页面 (`Cmd + Shift + R` 或 `Ctrl + Shift + R`)

### 问题 5: 文件下载失败

**检查任务状态**:
```bash
curl http://127.0.0.1:28888/api/convert/status/YOUR_TASK_ID
```

确保:
- 任务状态为 `completed`
- `outputPath` 字段有值

**使用进度查询接口**:
```bash
curl http://127.0.0.1:28888/api/progress/YOUR_TASK_ID
```

---

## 📚 API 接口快速参考

| 功能 | 接口 | 方法 |
|------|------|------|
| 初始化上传 | `/api/upload/init` | POST |
| 上传切片 | `/api/upload/chunk` | POST |
| 查询上传状态 | `/api/upload/status/:uploadId` | GET |
| 取消上传 | `/api/upload/cancel/:uploadId` | POST |
| 开始转换 | `/api/convert/start` | POST |
| 查询转换状态 | `/api/convert/status/:taskId` | GET |
| 取消转换 | `/api/convert/cancel/:taskId` | POST |
| 获取任务列表 | `/api/convert/list` | GET |
| 下载文件 | `/api/convert/download/:taskId` | GET |
| 统一进度查询 | `/api/progress/:id` | GET |

完整 API 文档请查看 [README.md](./README.md)

---

## 💡 使用建议

### 质量选择
- `low`: 快速转换,文件较小,质量一般 - 适合预览
- `medium`: 平衡质量和速度 - **推荐**
- `high`: 高质量,转换较慢,文件较大 - 适合最终输出

### 文件大小建议
- **< 10MB**: 使用单个切片上传即可
- **10MB - 100MB**: 使用 1MB 切片大小
- **> 100MB**: 使用 2-5MB 切片大小

### 生产环境建议
1. 设置固定端口(默认 28888)
2. 定期清理输出目录中的旧文件
3. 根据服务器 CPU 调整转换质量参数
4. 监控磁盘空间,避免空间不足

### 安全建议
- 服务仅监听 `127.0.0.1`,只允许本地访问
- 不要将服务暴露到公网
- 定期更新 FFmpeg 到最新版本

---

## 🔗 更多资源

- **完整 API 文档**: [README.md](./README.md)
- **前端示例代码**: [examples/demo.html](./examples/demo.html)
- **构建文档**: [docs/BUILD.md](./docs/BUILD.md)
- **接口测试**: 使用前端示例或 Postman/curl 测试

---

## ⚡ 常用命令

```bash
# 启动服务
go run main.go

# 编译
go build -o ffmpeg-binary

# 健康检查
curl http://127.0.0.1:28888/health

# 查看任务列表
curl http://127.0.0.1:28888/api/convert/list

# 查看上传状态
curl http://127.0.0.1:28888/api/upload/status/YOUR_UPLOAD_ID

# 查看转换进度
curl http://127.0.0.1:28888/api/progress/YOUR_TASK_ID

# 下载转换后的文件
curl http://127.0.0.1:28888/api/convert/download/YOUR_TASK_ID -o output.mp4
```

---

**需要帮助?**

- 查看完整文档: [README.md](./README.md)
- 查看示例代码: [examples/demo.html](./examples/demo.html)
- 提交问题: [GitHub Issues](https://github.com/your-repo/ffmpeg-binary/issues)
