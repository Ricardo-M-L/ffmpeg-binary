# ✅ Windows 服务版本 - 无黑窗口,后台运行

## 🎉 问题已解决!

新版本已将程序修改为 **Windows 服务**,安装后:
- ✅ **无黑色控制台窗口**
- ✅ **后台静默运行**
- ✅ **开机自动启动**
- ✅ **像系统服务一样运行**

---

## 📦 生成的文件

```
dist/windows/GoalfyMediaConverter-Setup.exe (3.6MB)
build/windows/ffmpeg-binary.exe (8.8MB - 包含服务功能)
```

---

## 🚀 安装体验

### 用户安装步骤

1. **双击安装包** `GoalfyMediaConverter-Setup.exe`
2. **按照向导操作**
   - 欢迎页面
   - 选择安装位置 (默认: `C:\Program Files\GoalfyMediaConverter\`)
   - 安装进度
3. **自动完成**:
   - ✅ 安装程序文件
   - ✅ 安装为 Windows 服务
   - ✅ 设置开机自启动
   - ✅ 启动服务
4. **完成!**
   - ❌ 没有黑窗口
   - ✅ 服务在后台运行
   - ✅ 访问: http://127.0.0.1:28888

---

## 🎯 服务管理

### 开始菜单快捷方式

安装后,在开始菜单 "GoalfyMediaConverter" 文件夹中有:

1. **启动服务.bat** - 手动启动服务
2. **停止服务.bat** - 停止服务
3. **查看服务状态.bat** - 查看服务运行状态
4. **打开服务管理器.bat** - 打开 Windows 服务管理器
5. **打开Web界面.url** - 打开 http://127.0.0.1:28888
6. **卸载** - 卸载程序

### 使用 Windows 服务管理器

1. 按 `Win + R`,输入 `services.msc`,回车
2. 找到 "**Goalfy Media Converter Service**"
3. 可以:
   - 启动/停止服务
   - 设置启动类型 (已默认设置为"自动")
   - 查看服务状态

### 命令行管理

打开**管理员**命令提示符:

```batch
# 查看服务状态
sc query GoalfyMediaConverter

# 启动服务
sc start GoalfyMediaConverter

# 停止服务
sc stop GoalfyMediaConverter

# 设置自动启动
sc config GoalfyMediaConverter start= auto
```

---

## 🛠️ 技术实现

### 代码改动

1. **新增 Windows 服务模块**
   `internal/service/windows_service.go`
   - 实现 Windows 服务接口
   - 支持安装/卸载/启动/停止

2. **新增 Windows 入口**
   `main_windows.go`
   - 使用 Go build tags (仅在 Windows 编译)
   - 检测运行模式 (服务 vs 控制台)
   - 处理服务命令

3. **更新主程序**
   `main.go`
   - 区分 Windows 和其他平台
   - Windows 自动使用服务模式

4. **更新安装脚本**
   `scripts/setup.nsi`
   - 调用 `install-service` 安装服务
   - 调用 `sc start` 启动服务
   - 创建服务管理快捷方式

### 服务特性

- **服务名称**: `GoalfyMediaConverter`
- **显示名称**: `Goalfy Media Converter Service`
- **描述**: `视频转换服务 - 提供 WebM 到 MP4 转换功能`
- **启动类型**: 自动 (开机启动)
- **运行账户**: Local System
- **状态**: 运行中

---

## 📋 卸载

### 方法 1: 使用安装包卸载

1. 在开始菜单找到 "GoalfyMediaConverter"
2. 点击 "卸载"
3. 按照向导操作

### 方法 2: 控制面板卸载

1. 打开 "设置" → "应用"
2. 找到 "GoalfyMediaConverter"
3. 点击卸载

卸载过程会自动:
- 停止服务
- 卸载服务
- 删除所有文件
- 清理注册表

---

## 🔍 故障排查

### 问题 1: 服务无法启动

**症状**: 安装完成后服务没有运行

**解决方法**:
1. 打开服务管理器 (`services.msc`)
2. 找到 "Goalfy Media Converter Service"
3. 右键 → 启动
4. 如果启动失败,查看事件查看器 (Event Viewer)

### 问题 2: 端口被占用

**症状**: 服务启动但无法访问 28888 端口

**解决方法**:
```batch
# 检查端口占用
netstat -ano | findstr :28888

# 如果被占用,停止占用进程或更换端口
```

### 问题 3: FFmpeg 未找到

**症状**: 服务运行但转换功能不可用

**解决方法**:
1. 下载 FFmpeg: https://www.gyan.dev/ffmpeg/builds/
2. 将 `ffmpeg.exe` 复制到: `C:\Program Files\GoalfyMediaConverter\bin\`
3. 重启服务

### 问题 4: 手动安装服务

如果安装程序未能自动安装服务:

```batch
# 打开管理员命令提示符
cd "C:\Program Files\GoalfyMediaConverter"

# 安装服务
ffmpeg-binary.exe install-service

# 启动服务
sc start GoalfyMediaConverter
```

---

## 📊 对比旧版本

| 功能 | 旧版本 (控制台) | 新版本 (服务) |
|------|--------------|-------------|
| 运行方式 | 控制台程序 | Windows 服务 |
| 黑色窗口 | ✗ 有 | ✅ 无 |
| 后台运行 | ✗ 窗口必须保持打开 | ✅ 完全后台 |
| 开机启动 | ✗ 手动启动 | ✅ 自动启动 |
| 服务管理 | ✗ 无 | ✅ Windows 服务管理器 |
| 用户体验 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

---

## 🎓 开发者信息

### 本地调试

在开发环境,如果想在控制台查看日志:

```batch
# Windows
ffmpeg-binary.exe debug

# 或直接运行(会显示帮助信息)
ffmpeg-binary.exe
```

### 手动操作服务

```batch
# 安装服务
ffmpeg-binary.exe install-service

# 卸载服务
ffmpeg-binary.exe uninstall-service

# 启动服务
ffmpeg-binary.exe start-service

# 停止服务
ffmpeg-binary.exe stop-service
```

### Go 依赖

新增依赖:
```go
golang.org/x/sys/windows/svc        // Windows 服务接口
golang.org/x/sys/windows/svc/mgr    // 服务管理
golang.org/x/sys/windows/svc/eventlog // 事件日志
```

### 构建标签

使用 Go build tags 实现平台特定代码:
- `main_windows.go` - 仅在 Windows 编译
- `main.go` - 所有平台通用

---

## ✅ 总结

现在 Windows 安装包:

1. ✅ **完全静默运行** - 无黑窗口
2. ✅ **开机自动启动** - 作为 Windows 服务
3. ✅ **专业的服务管理** - 可通过服务管理器控制
4. ✅ **一键安装** - 用户体验极佳
5. ✅ **干净卸载** - 自动清理所有内容

就像 **QQ**、**微信**、**网易云音乐** 等专业软件一样运行! 🎉

---

**测试建议**:
在 Windows 虚拟机或实体机上测试新的安装包,确认:
- 安装过程顺利
- 服务成功安装并启动
- 没有黑色窗口出现
- Web 界面可以访问
- 重启后服务自动启动
- 卸载干净彻底
