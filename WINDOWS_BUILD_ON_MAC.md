# 在 Mac 上构建 Windows 安装包

## 快速开始

### 前提条件

1. **安装 Docker Desktop**
   - 下载地址: https://www.docker.com/products/docker-desktop/
   - 安装后启动 Docker Desktop
   - 确保 Docker 正在运行 (菜单栏看到 Docker 图标)

### 一键打包

在 Mac 终端中运行:

```bash
# 进入项目目录
cd /Users/ricardo/Documents/jetbrains-projects/GolandProjects/ffmpeg-binary

# 执行打包脚本
./scripts/build-windows-on-mac.sh
```

**首次运行**: 需要构建 Docker 镜像,大约 5-10 分钟
**后续运行**: 只需要 1-2 分钟

### 输出

成功后会生成:

```
dist/windows/GoalfyMediaConverter-Setup.exe
```

这就是完整的 Windows 安装包,包含所有功能:
- ✅ 图形化安装界面
- ✅ 自动下载 FFmpeg
- ✅ 开机自启动选项
- ✅ 标准卸载程序

---

## 工作原理

### 技术方案

使用 **Docker + Wine + Inno Setup** 在 Mac 上构建 Windows 安装包:

1. **Docker 容器**: 创建一个 Debian 系统环境
2. **Wine**: 在 Linux 中运行 Windows 程序
3. **Go 交叉编译**: 编译 Windows 可执行文件 (.exe)
4. **Inno Setup**: 在 Wine 中运行,生成专业的 Windows 安装包

### 流程说明

```
Mac 电脑
  │
  ├─ Docker 容器启动
  │   │
  │   ├─ 编译 Windows 可执行文件 (Go 交叉编译)
  │   │   └─ build/windows/ffmpeg-binary.exe
  │   │
  │   ├─ 运行 Inno Setup (通过 Wine)
  │   │   └─ 读取 scripts/setup.iss
  │   │   └─ 生成安装包
  │   │
  │   └─ 输出最终文件
  │
  └─ dist/windows/GoalfyMediaConverter-Setup.exe
```

---

## 详细步骤

### 1. 检查 Docker 状态

```bash
# 确认 Docker 已安装
docker --version

# 确认 Docker 正在运行
docker info
```

如果 Docker 未运行,打开 Docker Desktop 应用。

### 2. 首次构建 (仅需一次)

第一次运行脚本时,会自动构建 Docker 镜像:

```bash
./scripts/build-windows-on-mac.sh
```

输出示例:

```
╔══════════════════════════════════════════════════════════════╗
║     Windows 安装包构建工具 (在 Mac 上运行)                 ║
╚══════════════════════════════════════════════════════════════╝

==> 清理旧文件...
==> 检查 Docker 镜像...
    Docker 镜像不存在,开始构建...
    (首次构建需要 5-10 分钟,请耐心等待)

    [Docker 构建进度...]

    ✅ Docker 镜像已构建
==> 编译 Windows 可执行文件...
    架构: amd64
    ✅ 可执行文件已生成
==> 创建安装包...
    ✅ 安装包已创建

╔══════════════════════════════════════════════════════════════╗
║                 ✅ 打包完成!                                ║
╚══════════════════════════════════════════════════════════════╝

📦 安装包: dist/windows/GoalfyMediaConverter-Setup.exe
```

### 3. 后续使用

Docker 镜像构建完成后,再次运行只需要 1-2 分钟:

```bash
./scripts/build-windows-on-mac.sh
```

---

## 测试安装包

### 方法 1: 在 Windows 虚拟机中测试

推荐使用:
- **Parallels Desktop** (Mac 专用,性能最好)
- **VMware Fusion**
- **VirtualBox** (免费)

安装 Windows 10/11 虚拟机后:

1. 将 `dist/windows/GoalfyMediaConverter-Setup.exe` 复制到虚拟机
2. 双击运行安装程序
3. 按照向导完成安装
4. 测试功能: http://127.0.0.1:28888

### 方法 2: 在物理 Windows 电脑上测试

直接复制安装包到 Windows 电脑,双击运行。

---

## 故障排查

### Docker 未安装

错误信息:

```
❌ 错误: 未找到 Docker
```

**解决方案**: 安装 Docker Desktop
https://www.docker.com/products/docker-desktop/

### Docker 未运行

错误信息:

```
❌ 错误: Docker 未运行
```

**解决方案**:
1. 打开 "Docker Desktop" 应用
2. 等待 Docker 图标出现在菜单栏
3. 重新运行脚本

### 编译失败

如果编译失败,检查:

1. **Go 依赖是否完整**:
   ```bash
   go mod download
   go mod vendor
   ```

2. **代码是否有语法错误**:
   ```bash
   go build .
   ```

### 镜像构建失败

如果 Docker 镜像构建失败:

1. **清理旧镜像重试**:
   ```bash
   docker rmi ffmpeg-binary-windows-builder
   ./scripts/build-windows-on-mac.sh
   ```

2. **检查网络连接** (需要下载 Go 和 Inno Setup)

3. **检查磁盘空间** (Docker 镜像约 2GB)

---

## 高级选项

### 重新构建 Docker 镜像

如果需要更新 Docker 镜像 (例如升级 Go 版本):

```bash
# 删除旧镜像
docker rmi ffmpeg-binary-windows-builder

# 重新构建
./scripts/build-windows-on-mac.sh
```

### 自定义 Docker 镜像

编辑 `scripts/Dockerfile.windows` 文件:

- 修改 Go 版本
- 修改 Inno Setup 版本
- 添加其他工具

然后重新构建镜像。

### 修改安装包配置

编辑 `scripts/setup.iss` 文件:

- 修改应用名称、版本号
- 修改安装路径
- 修改图标、许可证
- 添加/删除文件
- 修改安装选项

保存后重新运行打包脚本。

---

## 对比原方案

### 原方案 (仅在 Windows 上)

```
Windows 电脑
  ├─ 安装 Go
  ├─ 安装 Inno Setup
  └─ 运行 build-windows.bat
      └─ 生成安装包
```

**缺点**:
- 必须有 Windows 电脑或虚拟机
- 需要手动安装多个工具
- 跨平台开发不方便

### 新方案 (Mac 上一键完成)

```
Mac 电脑
  ├─ 安装 Docker (只需一次)
  └─ 运行 build-windows-on-mac.sh
      └─ 自动完成所有步骤
```

**优点**:
- ✅ 只需安装 Docker
- ✅ 一行命令完成
- ✅ 不需要 Windows 环境
- ✅ 可重复构建
- ✅ 干净的开发环境

---

## 性能优化

### Docker 镜像缓存

Docker 会自动缓存构建层,首次构建后,后续构建会很快。

### 并行编译

如果需要同时构建多个平台:

```bash
# 并行构建 macOS 和 Windows
./scripts/build-macos-pkg.sh &
./scripts/build-windows-on-mac.sh &
wait
```

---

## 常见问题 (FAQ)

**Q: 首次构建为什么这么慢?**
A: 需要下载 Debian 基础镜像、Go、Inno Setup 等,总共约 2GB。Docker 会缓存这些内容,后续构建会很快。

**Q: 生成的安装包能在所有 Windows 上运行吗?**
A: 支持 Windows 10/11 (64 位)。Inno Setup 生成的是标准 Windows 安装程序。

**Q: 我可以在 Linux 上使用这个方案吗?**
A: 可以!这个方案同样适用于 Linux。只需安装 Docker 后运行脚本即可。

**Q: Docker 镜像占用多少空间?**
A: 约 2GB。可以用 `docker images` 查看。

**Q: 如何删除 Docker 镜像?**
A: `docker rmi ffmpeg-binary-windows-builder`

**Q: 可以自动签名吗?**
A: 目前不支持代码签名。如需签名,需要在 Windows 上使用 `signtool.exe` 手动签名。

---

## 总结

现在你可以在 Mac 上一键生成 Windows 安装包了!

**使用场景**:
- 在 Mac 上开发,需要发布 Windows 版本
- CI/CD 自动化构建 (支持 macOS/Linux runners)
- 团队协作 (不需要每个人都有 Windows 环境)

**下一步**:
1. 测试生成的安装包
2. 发布到 GitHub Releases
3. 让用户下载使用

Happy Building! 🚀