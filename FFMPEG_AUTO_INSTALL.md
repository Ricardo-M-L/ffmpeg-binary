# FFmpeg 自动安装功能说明

## 🎯 功能概述

FFmpeg Binary 服务现在支持 **自动检测和安装 FFmpeg**!

- ✅ **自动检测**: 启动时自动检查 FFmpeg 是否已安装
- ✅ **自动安装**: 如果未安装,自动通过包管理器安装
- ✅ **跨平台支持**: macOS、Linux、Windows 全平台支持
- ✅ **零配置**: 用户无需手动安装 FFmpeg

## 📋 工作流程

```
启动服务
   ↓
检查 FFmpeg 是否存在
   ↓
┌──────────────┐
│ 已安装?     │
└──────────────┘
   ↓           ↓
  是          否
   ↓           ↓
使用现有    自动安装
   ↓           ↓
   └───────────┘
        ↓
   启动服务
```

## 🔍 检测逻辑

### 1. 查找已安装的 FFmpeg

程序会按以下顺序查找:

**macOS:**
- `ffmpeg` (PATH 中)
- `/opt/homebrew/bin/ffmpeg` (Apple Silicon Homebrew)
- `/usr/local/bin/ffmpeg` (Intel Homebrew)

**Linux:**
- `ffmpeg` (PATH 中)
- `/usr/bin/ffmpeg`
- `/usr/local/bin/ffmpeg`

**Windows:**
- `ffmpeg` (PATH 中)
- `C:\Program Files\ffmpeg\bin\ffmpeg.exe`
- `C:\ffmpeg\bin\ffmpeg.exe`

### 2. 验证可用性

找到后会执行 `ffmpeg -version` 验证是否可用。

## 🛠️ 自动安装策略

### macOS

1. 检查 Homebrew 是否安装
2. 如果未安装,先安装 Homebrew
3. 执行 `brew install ffmpeg`

```bash
# 自动执行的命令
brew install ffmpeg
```

### Linux

根据发行版自动选择包管理器:

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install -y ffmpeg
```

**Fedora/RHEL/CentOS:**
```bash
sudo dnf install -y ffmpeg
```

**Arch Linux:**
```bash
sudo pacman -S --noconfirm ffmpeg
```

### Windows

通过 Chocolatey 安装:

```bash
choco install ffmpeg -y
```

> ⚠️ **注意**: Windows 用户需要先安装 Chocolatey

## 📝 日志输出

### FFmpeg 已安装

```
✅ FFmpeg 已安装: /opt/homebrew/bin/ffmpeg
===========================================
🚀 FFmpeg Binary 服务启动成功!
===========================================
📡 服务地址: http://127.0.0.1:28888
...
```

### FFmpeg 未安装 - 自动安装

```
⚠️  FFmpeg 未安装或不可用,正在自动安装...
📦 正在通过 Homebrew 安装 FFmpeg...
[安装进度输出...]
✅ FFmpeg 安装成功: /opt/homebrew/bin/ffmpeg
===========================================
🚀 FFmpeg Binary 服务启动成功!
===========================================
```

### 安装失败

```
FFmpeg 检查/安装失败: 安装 FFmpeg 失败: brew install ffmpeg 失败: ...
```

## 🧪 测试方法

### 方法 1: 使用测试脚本

```bash
./test_ffmpeg_installer.sh
```

### 方法 2: 手动测试

```bash
# 1. 构建
go build -o ffmpeg-binary-test .

# 2. 运行(会自动检测/安装 FFmpeg)
./ffmpeg-binary-test
```

### 方法 3: 模拟 FFmpeg 未安装

```bash
# 临时重命名 ffmpeg(仅用于测试)
sudo mv /opt/homebrew/bin/ffmpeg /opt/homebrew/bin/ffmpeg.backup

# 运行服务(会自动安装)
./ffmpeg-binary-test

# 恢复(如果需要)
sudo mv /opt/homebrew/bin/ffmpeg.backup /opt/homebrew/bin/ffmpeg
```

## 🎯 用户体验

### 之前

```
❌ 问题: 用户需要手动安装 FFmpeg
用户: "为什么服务启动不了?"
开发: "你需要先运行 brew install ffmpeg"
用户: "什么是 brew?"
```

### 现在

```
✅ 改进: 完全自动化
用户: 双击安装包
系统: 自动检测并安装 FFmpeg
用户: 直接使用,无需任何配置
```

## 📦 打包说明

新功能已集成到打包流程中,无需修改打包脚本。

### PKG 安装包

```bash
./scripts/build-macos-pkg.sh
```

安装包会:
1. 安装应用到 `/Applications/`
2. 启动服务
3. **自动检测并安装 FFmpeg**(新增)
4. 配置自启动

## 🔧 代码结构

```
internal/installer/
└── ffmpeg_installer.go    # FFmpeg 自动安装器

修改的文件:
- main.go                   # 集成自动安装
- internal/server/server.go # 移除手动验证
```

## 🚨 注意事项

1. **首次安装可能较慢**: FFmpeg 安装包较大(~100MB),首次安装需要几分钟
2. **需要管理员权限**: 安装 FFmpeg 可能需要 sudo 密码
3. **网络连接**: 需要网络连接下载 FFmpeg
4. **Windows 依赖**: Windows 用户需要先安装 Chocolatey

## 💡 优化建议

### 未来改进方向

1. **预下载 FFmpeg**: 将 FFmpeg 打包到安装包中,完全离线安装
2. **进度提示**: 显示安装进度百分比
3. **静默安装**: 提供静默安装选项,无需用户交互
4. **降级支持**: 如果自动安装失败,提示用户手动安装

## 🐛 故障排查

### 问题: Homebrew 安装失败

```bash
# 手动安装 Homebrew
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

### 问题: 权限不足

```bash
# 确保有 sudo 权限
sudo -v
```

### 问题: 网络问题

```bash
# 检查网络连接
curl -I https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh
```

## 📚 相关文档

- [FFmpeg 官网](https://ffmpeg.org/)
- [Homebrew 官网](https://brew.sh/)
- [Chocolatey 官网](https://chocolatey.org/)
