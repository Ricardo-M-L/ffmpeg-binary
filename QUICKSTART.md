# 🚀 快速开始指南

## 第一步: 安装 FFmpeg

### macOS
```bash
brew install ffmpeg
```

### Windows
1. 访问
2. 下载 "ffmpeg-release-essentials.zip"
3. 解压到 `C:\ffmpeg\`
4. 添加 `C:\ffmpeg\bin` 到系统 PATH

### Linux (Ubuntu/Debian)
```bash
sudo apt update
sudo apt install ffmpeg
```

## 第二步: 运行服务

### 开发模式
```bash
cd /Users/ricardo/Documents/jetbrains-projects/GolandProjects/ffmpeg-binary
go run main.go
```

### 生产模式
```bash
# 编译
go build -o ffmpeg-binary

# 运行
./ffmpeg-binary

# 服务默认在端口 28888 运行
# 查看输出确认端口号
```

## 第三步: 测试服务

### 健康检查
```bash
curl http://127.0.0.1:28888/health
# 预期输出: {"status":"ok","port":28888}
```

### 使用前端示例
```bash
# 在浏览器打开
open examples/demo.html
```

## 第四步: 安装自启动(可选)

```bash
# 安装
./ffmpeg-binary install

# 卸载
./ffmpeg-binary uninstall
```

## 🎬 使用示例

### 1. 命令行测试(同步转换)

```bash
# 假设你有一个 test.webm 文件
curl -X POST http://127.0.0.1:28888/api/v1/convert/sync \
  -H "Content-Type: video/webm" \
  --data-binary @test.webm \
  -o output.mp4
```

### 2. JavaScript 示例(同步)

```javascript
// 选择文件
const fileInput = document.querySelector('input[type="file"]');
const file = fileInput.files[0];

// 转换
const response = await fetch('http://127.0.0.1:28888/api/v1/convert/sync', {
  method: 'POST',
  headers: { 'Content-Type': 'video/webm' },
  body: file
});

const mp4Blob = await response.blob();
const url = URL.createObjectURL(mp4Blob);

// 播放或下载
const video = document.createElement('video');
video.src = url;
video.controls = true;
document.body.appendChild(video);
```

### 3. JavaScript 示例(异步 - 大文件)

```javascript
// 1. 创建任务
const createResp = await fetch('http://127.0.0.1:28888/api/v1/convert/async', {
  method: 'POST'
});
const { task_id, upload_url } = await createResp.json();

// 2. 分片上传
const chunkSize = 1024 * 1024; // 1MB
const totalChunks = Math.ceil(file.size / chunkSize);

for (let i = 0; i < totalChunks; i++) {
  const chunk = file.slice(i * chunkSize, (i + 1) * chunkSize);
  const isLast = i === totalChunks - 1;

  await fetch(`http://127.0.0.1:28888${upload_url}`, {
    method: 'POST',
    headers: { 'X-Last-Chunk': isLast ? 'true' : 'false' },
    body: chunk
  });

  console.log(`上传进度: ${i + 1}/${totalChunks}`);
}

// 3. 轮询状态
while (true) {
  await new Promise(resolve => setTimeout(resolve, 1000));

  const statusResp = await fetch(`http://127.0.0.1:28888/api/v1/task/${task_id}`);
  const status = await statusResp.json();

  console.log(`转换进度: ${status.progress}%`);

  if (status.status === 'completed') {
    // 4. 下载结果
    const downloadResp = await fetch(
      `http://127.0.0.1:28888/api/v1/task/${task_id}/download`
    );
    const mp4Blob = await downloadResp.blob();

    // 使用 blob
    const url = URL.createObjectURL(mp4Blob);
    console.log('转换完成!', url);
    break;
  } else if (status.status === 'failed') {
    console.error('转换失败:', status.error);
    break;
  }
}
```

## 📦 打包部署

### macOS DMG
```bash
./build-macos.sh

# 输出: build/macos/FFmpeg-Binary-Installer.dmg
# 双击 DMG,拖拽到 Applications 即可安装
```

### Windows EXE
```bash
# 在 Windows 上运行
build-windows.bat

# 输出: build\windows\ffmpeg-binary.exe
# 运行 installer\install.bat 完成安装
```

## 🔍 故障排查

### 问题: 服务无法启动

**检查 FFmpeg**:
```bash
ffmpeg -version
```

**检查端口占用**:
```bash
# macOS/Linux
lsof -i :28888

# Windows
netstat -ano | findstr :28888
```

### 问题: 转换失败

1. 确认文件是 WebM 格式
2. 检查磁盘空间
3. 查看日志: `~/Library/Logs/ffmpeg-binary.log` (macOS)

### 问题: 前端跨域错误

服务已启用 CORS,如果仍有问题:
1. 确认服务地址为 `http://127.0.0.1:28888`
2. 不要使用 `localhost`,使用 `127.0.0.1`

## 📚 更多信息

- 完整 API 文档: 查看 `README.md`
- 前端示例: `examples/demo.html`
- 项目总结: `PROJECT_SUMMARY.md`

## 💡 提示

1. **小文件(< 10MB)**: 使用同步转换接口
2. **大文件(> 10MB)**: 使用异步转换接口
3. **生产环境**: 建议设置固定端口(修改配置文件)
4. **性能优化**: 根据 CPU 调整 FFmpeg 编码参数

---

**需要帮助?** 查看完整文档或提交 Issue