#!/bin/bash
# macOS PKG 安装包构建脚本 - 图形化安装界面

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

APP_NAME="GoalfyMediaConverter"
BUNDLE_ID="com.goalfy.mediaconverter"
VERSION="1.0.0"
INSTALL_LOCATION="/Applications/GoalfyMediaConverter.app"
DIST_DIR="dist/macos"
PKG_NAME="GoalfyMediaConverter-Installer.pkg"
ICON_FILE="assets/icons/icon.icns"

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║           macOS PKG 安装包构建工具                         ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# 清理旧构建
echo "==> 清理旧文件..."
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR/pkg-root"
mkdir -p "$DIST_DIR/scripts"
mkdir -p "$DIST_DIR/resources"

# 编译 Universal Binary
echo "==> 编译 macOS Universal Binary..."
echo "    架构: amd64 + arm64"

GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "$DIST_DIR/ffmpeg-binary-amd64" .
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "$DIST_DIR/ffmpeg-binary-arm64" .

lipo -create -output "$DIST_DIR/ffmpeg-binary" \
    "$DIST_DIR/ffmpeg-binary-amd64" \
    "$DIST_DIR/ffmpeg-binary-arm64"

rm "$DIST_DIR/ffmpeg-binary-amd64" "$DIST_DIR/ffmpeg-binary-arm64"
echo "    ✅ Universal Binary 已生成"

# 创建 .app 包结构
echo "==> 创建 .app 包..."
APP_PATH="$DIST_DIR/pkg-root/Applications/GoalfyMediaConverter.app"
mkdir -p "$APP_PATH/Contents/MacOS"
mkdir -p "$APP_PATH/Contents/Resources"

# 复制可执行文件和 GUI 启动器
cp "$DIST_DIR/ffmpeg-binary" "$APP_PATH/Contents/MacOS/ffmpeg-binary-service"
chmod +x "$APP_PATH/Contents/MacOS/ffmpeg-binary-service"

# 复制 GUI 启动器作为主可执行文件
cp "scripts/gui-launcher.sh" "$APP_PATH/Contents/MacOS/ffmpeg-binary"
chmod +x "$APP_PATH/Contents/MacOS/ffmpeg-binary"

# 复制卸载脚本到 Resources
cp "scripts/uninstall.sh" "$APP_PATH/Contents/Resources/"
chmod +x "$APP_PATH/Contents/Resources/uninstall.sh"

# 复制图标
if [ -f "$ICON_FILE" ]; then
    cp "$ICON_FILE" "$APP_PATH/Contents/Resources/"
    ICON_ENTRY="    <key>CFBundleIconFile</key>
    <string>icon.icns</string>"
else
    echo "    ⚠️  图标文件不存在"
    ICON_ENTRY=""
fi

# 创建 Info.plist
cat > "$APP_PATH/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>ffmpeg-binary</string>
    <key>CFBundleIdentifier</key>
    <string>$BUNDLE_ID</string>
    <key>CFBundleName</key>
    <string>$APP_NAME</string>
    <key>CFBundleDisplayName</key>
    <string>$APP_NAME</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>$VERSION</string>
    <key>CFBundleVersion</key>
    <string>1</string>
$ICON_ENTRY
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>NSHighResolutionCapable</key>
    <string>true</string>
    <key>LSApplicationCategoryType</key>
    <string>public.app-category.utilities</string>
</dict>
</plist>
EOF

echo "    ✅ .app 包已创建"

# 创建安装后脚本 (postinstall)
echo "==> 创建安装脚本..."
cat > "$DIST_DIR/scripts/postinstall" << 'POSTINSTALL'
#!/bin/bash

# 获取当前用户
CURRENT_USER="${USER}"
if [ -z "$CURRENT_USER" ]; then
    CURRENT_USER=$(stat -f "%Su" /dev/console)
fi

USER_HOME=$(eval echo ~$CURRENT_USER)
FFMPEG_INSTALL_DIR="/usr/local/bin"
LOG_FILE="$USER_HOME/Library/Logs/goalfy-mediaconverter-install.log"

# 确保日志目录存在
mkdir -p "$USER_HOME/Library/Logs"

# 重定向所有输出到日志文件
exec > >(tee -a "$LOG_FILE") 2>&1

echo "========================================="
echo "GoalfyMediaConverter 安装脚本"
echo "开始时间: $(date)"
echo "========================================="
echo ""

echo "配置 GoalfyMediaConverter 服务..."

# 1. 检查并安装 FFmpeg (在启动服务前)
echo "检查 FFmpeg 是否已安装..."
if ! command -v ffmpeg &> /dev/null; then
    echo "⚠️  FFmpeg 未安装,正在下载静态编译版本..."

    # 检测 CPU 架构
    ARCH=$(uname -m)
    echo "检测到 CPU 架构: $ARCH"

    # evermeet.cx 只提供 x86_64 版本,需要 Rosetta 2 在 Apple Silicon 上运行
    # 先检查是否有 Rosetta 2 (对于 Apple Silicon Mac)
    if [ "$ARCH" = "arm64" ]; then
        echo "检测到 Apple Silicon Mac"
        if ! /usr/bin/pgrep -q oahd; then
            echo "⚠️  未检测到 Rosetta 2,正在安装..."
            # 静默安装 Rosetta 2
            /usr/sbin/softwareupdate --install-rosetta --agree-to-license 2>&1 | tee -a "$LOG_FILE" || true
            sleep 2
        else
            echo "✓ Rosetta 2 已安装"
        fi
    fi

    # 使用 evermeet.cx 的 x86_64 版本 (会通过 Rosetta 2 在 Apple Silicon 上运行)
    FFMPEG_URL="https://evermeet.cx/ffmpeg/getrelease/zip"
    echo "下载 FFmpeg (x86_64 版本,支持所有 Mac 通过 Rosetta 2)"

    # 下载到临时目录
    TMP_DIR=$(mktemp -d)
    echo "下载 FFmpeg (可能需要几分钟,取决于网络速度)..."

    # 添加超时参数防止卡死:
    # --connect-timeout 30: 连接超时 30 秒
    # --max-time 300: 总下载时间不超过 5 分钟
    # -S: 显示错误信息
    # --retry 2: 失败时重试 2 次
    # --retry-delay 3: 重试间隔 3 秒
    if ! curl -L -S --connect-timeout 30 --max-time 300 --retry 2 --retry-delay 3 -o "$TMP_DIR/ffmpeg.zip" "$FFMPEG_URL"; then
        echo "❌ FFmpeg 下载失败 (可能是网络问题或下载超时)"
        echo "   您可以稍后手动安装 FFmpeg: brew install ffmpeg"
        echo "   或从 https://evermeet.cx/ffmpeg/ 下载安装"
        rm -rf "$TMP_DIR"
        # 下载失败不阻止安装继续,让用户可以手动安装 FFmpeg
        echo "⚠️ 跳过 FFmpeg 安装,继续配置服务..."
    else
        echo "✓ FFmpeg 下载完成"

        # 解压 ZIP 文件 (macOS 自带 unzip)
        echo "解压 FFmpeg..."
        cd "$TMP_DIR"
        unzip -q ffmpeg.zip

        # 安装到系统目录 (postinstall 已经是 root 权限,可以直接复制)
        if [ -f "ffmpeg" ]; then
            echo "安装 FFmpeg 到 $FFMPEG_INSTALL_DIR..."

            # 确保目录存在
            mkdir -p "$FFMPEG_INSTALL_DIR"

            # 直接复制 (已经是 root 权限)
            cp -f ffmpeg "$FFMPEG_INSTALL_DIR/ffmpeg"

            # 设置执行权限
            chmod 755 "$FFMPEG_INSTALL_DIR/ffmpeg"

            echo "✓ FFmpeg 安装成功"
        else
            echo "❌ FFmpeg 解压失败"
            rm -rf "$TMP_DIR"
            echo "⚠️ 跳过 FFmpeg 安装,继续配置服务..."
        fi

        # 清理临时文件
        rm -rf "$TMP_DIR"
    fi
else
    echo "✓ FFmpeg 已安装: $(which ffmpeg)"
fi

# 2. 安装自启动配置 (只安装,不立即启动)
echo ""
echo "配置自启动..."
if sudo -u "$CURRENT_USER" /Applications/GoalfyMediaConverter.app/Contents/MacOS/ffmpeg-binary-service install 2>&1 | tee -a "$LOG_FILE"; then
    echo "✓ 自启动配置已安装"
else
    echo "⚠️ 自启动配置失败,请查看日志"
fi

# 3. 加载 LaunchAgent 立即启动服务
echo ""
echo "启动 GoalfyMediaConverter 服务..."
PLIST_PATH="$USER_HOME/Library/LaunchAgents/com.ffmpeg.binary.plist"
if [ -f "$PLIST_PATH" ]; then
    # 使用 launchctl 加载服务,让 launchd 负责启动
    # 这样不会阻塞 postinstall 脚本
    sudo -u "$CURRENT_USER" launchctl load "$PLIST_PATH" 2>&1 | tee -a "$LOG_FILE" || true
    echo "✓ 服务配置已加载,将在后台启动"
    echo "  提示: 服务将在几秒钟内启动完成"
else
    echo "⚠️ 未找到 LaunchAgent 配置文件"
fi

# 4. 修改应用包的所有权为当前用户,避免删除时需要密码
chown -R "$CURRENT_USER:staff" /Applications/GoalfyMediaConverter.app
echo "✓ 已设置应用包权限"

# 5. 显示安装成功通知
echo ""
echo "========================================="
echo "安装完成!"
echo "结束时间: $(date)"
echo "========================================="
echo ""
echo "服务信息:"
echo "  • 地址: http://127.0.0.1:28888"
echo "  • 日志: $USER_HOME/Library/Logs/goalfy-mediaconverter.log"
echo "  • 安装日志: $LOG_FILE"
echo ""
echo "提示:"
echo "  • 服务已在后台启动 (可能需要几秒钟)"
echo "  • 如果服务未启动,请查看日志文件"
echo "  • 卸载方法: 直接将应用拖到废纸篓即可"
echo ""

sudo -u "$CURRENT_USER" osascript -e 'display notification "GoalfyMediaConverter 已安装,拖到废纸篓即可自动卸载" with title "安装成功"' 2>/dev/null || true

exit 0
POSTINSTALL

chmod +x "$DIST_DIR/scripts/postinstall"

# 创建卸载前脚本 (preinstall) - 如果已安装则先停止
cat > "$DIST_DIR/scripts/preinstall" << 'PREINSTALL'
#!/bin/bash

# 如果服务正在运行,先停止
pkill -f ffmpeg-binary 2>/dev/null || true

# 等待进程完全停止
sleep 1

exit 0
PREINSTALL

chmod +x "$DIST_DIR/scripts/preinstall"

echo "    ✅ 安装脚本已创建"

# 创建欢迎信息
echo "==> 创建安装界面文本..."
cat > "$DIST_DIR/resources/welcome.html" << 'WELCOME'
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; padding: 20px; }
        h1 { color: #667eea; }
        .feature { margin: 10px 0; }
        .icon { color: #667eea; font-size: 20px; }
    </style>
</head>
<body>
    <h1>欢迎安装 GoalfyMediaConverter</h1>
    <p>GoalfyMediaConverter 是一个本地视频转换服务,提供 WebM 到 MP4 的转换功能。</p>

    <h3>主要功能:</h3>
    <div class="feature">✓ 同步视频流转换</div>
    <div class="feature">✓ 异步分块上传转换</div>
    <div class="feature">✓ 任务状态查询</div>
    <div class="feature">✓ 本地服务 (127.0.0.1)</div>
    <div class="feature">✓ 智能端口选择 (28888)</div>
    <div class="feature">✓ 开机自动启动</div>

    <h3>系统要求:</h3>
    <p>• macOS 10.15 或更高版本<br>
       • 无需手动安装任何依赖,FFmpeg 将自动下载安装</p>

    <p><strong>注意:</strong> 服务将在后台静默运行,不会显示任何窗口。</p>
</body>
</html>
WELCOME

# 创建结束信息
cat > "$DIST_DIR/resources/conclusion.html" << 'CONCLUSION'
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; padding: 20px; }
        h1 { color: #4CAF50; }
        .info {
            background: #2d2d2d;
            color: #ffffff;
            padding: 15px;
            border-radius: 5px;
            margin: 10px 0;
            border: 1px solid #4a4a4a;
        }
        .info h3 {
            color: #ffffff;
            margin-top: 0;
        }
        .info strong {
            color: #ffd700;
        }
        code {
            background: #1a1a1a;
            color: #00ff00;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: Monaco, Consolas, monospace;
        }
    </style>
</head>
<body>
    <h1>安装完成!</h1>
    <p>GoalfyMediaConverter 已成功安装。</p>

    <div class="info">
        <h3>服务信息:</h3>
        <p>🌐 服务地址: <strong>http://127.0.0.1:28888</strong><br>
           📊 健康检查: <strong>http://127.0.0.1:28888/health</strong><br>
           📁 日志文件: <strong>~/Library/Logs/goalfy-mediaconverter.log</strong></p>
    </div>

    <h3>使用方法:</h3>
    <p>服务已在后台启动,可以直接通过 API 使用。详细 API 文档请查看项目 README。</p>

    <h3>卸载方法:</h3>
    <p><strong>只需拖到废纸篓即可!</strong></p>
    <p>直接将 GoalfyMediaConverter.app 从"应用程序"或启动台拖到废纸篓,系统会在 1 分钟内自动清理所有相关文件和服务,包括:</p>
    <ul>
        <li>✓ 停止运行中的服务进程</li>
        <li>✓ 移除自启动配置</li>
        <li>✓ 清理数据目录 (~/.goalfy-mediaconverter)</li>
    </ul>
    <p><small>💡 提示:拖到废纸篓后约 1 分钟内自动清理完成,无需清空废纸篓</small></p>
</body>
</html>
CONCLUSION

echo "    ✅ 安装界面文本已创建"

# 构建组件包
echo "==> 构建组件包..."
pkgbuild --root "$DIST_DIR/pkg-root" \
         --scripts "$DIST_DIR/scripts" \
         --identifier "$BUNDLE_ID" \
         --version "$VERSION" \
         --install-location "/" \
         "$DIST_DIR/component.pkg"

echo "    ✅ 组件包已创建"

# 创建 Distribution 定义
echo "==> 创建 Distribution 定义..."
cat > "$DIST_DIR/distribution.xml" << EOF
<?xml version="1.0" encoding="utf-8"?>
<installer-gui-script minSpecVersion="1">
    <title>GoalfyMediaConverter</title>
    <background file="background.png" alignment="bottomleft" scaling="proportional"/>
    <welcome file="welcome.html"/>
    <conclusion file="conclusion.html"/>
    <pkg-ref id="$BUNDLE_ID"/>
    <options customize="never" require-scripts="false" hostArchitectures="x86_64,arm64"/>
    <choices-outline>
        <line choice="default">
            <line choice="$BUNDLE_ID"/>
        </line>
    </choices-outline>
    <choice id="default"/>
    <choice id="$BUNDLE_ID" visible="false">
        <pkg-ref id="$BUNDLE_ID"/>
    </choice>
    <pkg-ref id="$BUNDLE_ID" version="$VERSION" onConclusion="none">component.pkg</pkg-ref>
</installer-gui-script>
EOF

# 复制背景图片 (如果有)
if [ -f "assets/pkg-background.png" ]; then
    cp "assets/pkg-background.png" "$DIST_DIR/resources/background.png"
fi

# 构建最终的产品包
echo "==> 构建最终安装包..."
productbuild --distribution "$DIST_DIR/distribution.xml" \
             --resources "$DIST_DIR/resources" \
             --package-path "$DIST_DIR" \
             "$DIST_DIR/$PKG_NAME"

echo "    ✅ 安装包已创建"

# 设置包图标 (可选)
if [ -f "$ICON_FILE" ] && command -v fileicon &> /dev/null; then
    echo "==> 设置安装包图标..."
    fileicon set "$DIST_DIR/$PKG_NAME" "$ICON_FILE"
    echo "    ✅ 图标已设置"
fi

# 清理临时文件
echo "==> 清理临时文件..."
rm -rf "$DIST_DIR/pkg-root"
rm -rf "$DIST_DIR/scripts"
rm -rf "$DIST_DIR/resources"
rm -f "$DIST_DIR/component.pkg"
rm -f "$DIST_DIR/distribution.xml"
rm -f "$DIST_DIR/ffmpeg-binary"

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                 ✅ 打包完成!                                ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "📦 安装包: $DIST_DIR/$PKG_NAME"
echo ""
echo "使用方法:"
echo "  1. 双击 PKG 文件"
echo "  2. 按照图形化界面提示完成安装"
echo "  3. 服务将自动在后台启动,无需任何窗口操作"
echo ""
echo "特点:"
echo "  ✓ 标准的 macOS 图形化安装界面"
echo "  ✓ 自动安装到 /Applications/"
echo "  ✓ 自动配置开机自启动"
echo "  ✓ 自动启动后台服务"
echo "  ✓ 无终端窗口,静默运行"
echo ""