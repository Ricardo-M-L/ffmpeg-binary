#!/bin/bash
# macOS PKG 安装包构建脚本 - 图形化安装界面

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

APP_NAME="FFmpeg Binary"
BUNDLE_ID="com.ffmpeg.binary"
VERSION="1.0.0"
INSTALL_LOCATION="/Applications/FFmpeg-Binary.app"
DIST_DIR="dist/macos"
PKG_NAME="FFmpeg-Binary-Installer.pkg"
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
APP_PATH="$DIST_DIR/pkg-root/Applications/FFmpeg-Binary.app"
mkdir -p "$APP_PATH/Contents/MacOS"
mkdir -p "$APP_PATH/Contents/Resources"

# 复制可执行文件
cp "$DIST_DIR/ffmpeg-binary" "$APP_PATH/Contents/MacOS/"
chmod +x "$APP_PATH/Contents/MacOS/ffmpeg-binary"

# 复制图标
if [ -f "$ICON_FILE" ]; then
    cp "$ICON_FILE" "$APP_PATH/Contents/Resources/"
    ICON_ENTRY="    <key>CFBundleIconFile</key>\n    <string>icon.icns</string>"
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
    <key>LSUIElement</key>
    <true/>
    <key>NSHighResolutionCapable</key>
    <true/>
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

echo "配置 FFmpeg Binary 服务..."

# 安装自启动 (作为当前用户)
sudo -u "$CURRENT_USER" /Applications/FFmpeg-Binary.app/Contents/MacOS/ffmpeg-binary install 2>/dev/null || true

# 启动服务 (作为当前用户)
sudo -u "$CURRENT_USER" nohup /Applications/FFmpeg-Binary.app/Contents/MacOS/ffmpeg-binary > "$USER_HOME/Library/Logs/ffmpeg-binary.log" 2>&1 &

# 等待服务启动
sleep 2

# 显示安装成功消息
cat > /tmp/ffmpeg-binary-install.txt << 'EOF'

╔══════════════════════════════════════════════════════════════╗
║              FFmpeg Binary 安装成功!                        ║
╚══════════════════════════════════════════════════════════════╝

✅ 服务已启动并设置为开机自启
📁 日志文件: ~/Library/Logs/ffmpeg-binary.log
🌐 服务地址: http://127.0.0.1:18888

服务将在后台运行,不会显示任何窗口。
EOF

# 如果有图形界面,显示通知
sudo -u "$CURRENT_USER" osascript -e 'display notification "FFmpeg Binary 服务已安装并启动" with title "安装成功"' 2>/dev/null || true

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
    <h1>欢迎安装 FFmpeg Binary 服务</h1>
    <p>FFmpeg Binary 是一个本地视频转换服务,提供 WebM 到 MP4 的转换功能。</p>

    <h3>主要功能:</h3>
    <div class="feature">✓ 同步视频流转换</div>
    <div class="feature">✓ 异步分块上传转换</div>
    <div class="feature">✓ 任务状态查询</div>
    <div class="feature">✓ 本地服务 (127.0.0.1)</div>
    <div class="feature">✓ 智能端口选择 (18888-28888)</div>
    <div class="feature">✓ 开机自动启动</div>

    <h3>系统要求:</h3>
    <p>• macOS 10.15 或更高版本<br>
       • FFmpeg 4.0+ (安装命令: <code>brew install ffmpeg</code>)</p>

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
        .info { background: #f5f5f5; padding: 15px; border-radius: 5px; margin: 10px 0; }
    </style>
</head>
<body>
    <h1>安装完成!</h1>
    <p>FFmpeg Binary 服务已成功安装。</p>

    <div class="info">
        <h3>服务信息:</h3>
        <p>🌐 服务地址: <strong>http://127.0.0.1:18888</strong><br>
           📊 健康检查: <strong>http://127.0.0.1:18888/health</strong><br>
           📁 日志文件: <strong>~/Library/Logs/ffmpeg-binary.log</strong></p>
    </div>

    <h3>使用方法:</h3>
    <p>服务已在后台启动,可以直接通过 API 使用。详细 API 文档请查看项目 README。</p>

    <h3>卸载方法:</h3>
    <p>1. 停止服务: <code>pkill -f ffmpeg-binary</code><br>
       2. 删除自启动: <code>/Applications/FFmpeg-Binary.app/Contents/MacOS/ffmpeg-binary uninstall</code><br>
       3. 删除应用: 在"应用程序"中删除 FFmpeg-Binary.app</p>
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
    <title>FFmpeg Binary</title>
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