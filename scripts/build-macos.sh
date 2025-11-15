#!/bin/bash
# macOS 打包脚本 - 生成可自动安装的 DMG

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

APP_NAME="FFmpeg Binary"
APP_BUNDLE="FFmpeg-Binary.app"
DMG_NAME="FFmpeg-Binary-Installer.dmg"
DIST_DIR="dist/macos"
ICON_FILE="assets/icons/icon.icns"

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║           macOS DMG 打包工具                                ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# 清理旧的构建
echo "==> 清理旧文件..."
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# 编译二进制
echo "==> 编译 macOS Universal Binary..."
echo "    架构: amd64 + arm64"

GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "$DIST_DIR/ffmpeg-binary-amd64" .
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "$DIST_DIR/ffmpeg-binary-arm64" .

# 合并为 Universal Binary
lipo -create -output "$DIST_DIR/ffmpeg-binary" \
    "$DIST_DIR/ffmpeg-binary-amd64" \
    "$DIST_DIR/ffmpeg-binary-arm64"

rm "$DIST_DIR/ffmpeg-binary-amd64" "$DIST_DIR/ffmpeg-binary-arm64"
echo "    ✅ Universal Binary 已生成"

# 创建 .app 包结构
echo "==> 创建 .app 包..."
APP_PATH="$DIST_DIR/$APP_BUNDLE"
mkdir -p "$APP_PATH/Contents/MacOS"
mkdir -p "$APP_PATH/Contents/Resources"

# 复制可执行文件
cp "$DIST_DIR/ffmpeg-binary" "$APP_PATH/Contents/MacOS/"
chmod +x "$APP_PATH/Contents/MacOS/ffmpeg-binary"

# 复制图标(如果存在)
if [ -f "$ICON_FILE" ]; then
    cp "$ICON_FILE" "$APP_PATH/Contents/Resources/"
    ICON_ENTRY="    <key>CFBundleIconFile</key>\n    <string>icon.icns</string>"
else
    echo "    ⚠️  图标文件不存在,运行 scripts/generate-icons.sh 生成"
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
    <string>com.ffmpeg.binary</string>
    <key>CFBundleName</key>
    <string>$APP_NAME</string>
    <key>CFBundleDisplayName</key>
    <string>$APP_NAME</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
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

# 创建自动安装脚本
echo "==> 创建安装脚本..."
mkdir -p "$DIST_DIR/dmg-content"

cat > "$DIST_DIR/dmg-content/安装.command" << 'INSTALL_SCRIPT'
#!/bin/bash

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║         FFmpeg Binary 服务安装程序                          ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
APP_PATH="$SCRIPT_DIR/FFmpeg-Binary.app"

# 检查是否有系统 FFmpeg
if ! command -v ffmpeg &> /dev/null; then
    echo "⚠️  检测到系统未安装 FFmpeg"
    echo ""
    echo "请选择:"
    echo "  1) 继续安装(服务需要 FFmpeg 才能工作)"
    echo "  2) 取消安装,先安装 FFmpeg"
    echo ""
    read -p "请输入选择 [1/2]: " choice

    if [ "$choice" != "1" ]; then
        echo ""
        echo "已取消安装"
        echo "请先安装 FFmpeg:"
        echo "  brew install ffmpeg"
        echo ""
        echo "然后重新运行此安装程序"
        exit 0
    fi
fi

echo "==> 正在安装..."

# 复制应用到 Applications
echo "    • 复制应用到 /Applications/"
cp -R "$APP_PATH" "/Applications/"

# 安装自启动
echo "    • 配置开机自启动"
/Applications/FFmpeg-Binary.app/Contents/MacOS/ffmpeg-binary install

# 启动服务
echo "    • 启动服务"
nohup /Applications/FFmpeg-Binary.app/Contents/MacOS/ffmpeg-binary > ~/Library/Logs/ffmpeg-binary.log 2>&1 &

sleep 2

# 检查服务状态
if pgrep -f "ffmpeg-binary" > /dev/null; then
    echo ""
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                 ✅ 安装成功!                                ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo ""
    echo "服务已启动并设置为开机自启"
    echo "日志文件: ~/Library/Logs/ffmpeg-binary.log"
    echo ""

    # 尝试获取端口
    sleep 1
    if [ -f ~/.ffmpeg-binary/config.json ]; then
        PORT=$(grep -o '"port":[0-9]*' ~/.ffmpeg-binary/config.json | grep -o '[0-9]*')
        if [ -n "$PORT" ]; then
            echo "服务地址: http://127.0.0.1:$PORT"
            echo "健康检查: http://127.0.0.1:$PORT/health"
        fi
    fi
else
    echo ""
    echo "⚠️  服务未能自动启动"
    echo "请手动运行: /Applications/FFmpeg-Binary.app/Contents/MacOS/ffmpeg-binary"
fi

echo ""
read -p "按回车键关闭..."

INSTALL_SCRIPT

chmod +x "$DIST_DIR/dmg-content/安装.command"

# 复制 app 到 dmg 内容目录
cp -R "$APP_PATH" "$DIST_DIR/dmg-content/"

# 创建 README
cat > "$DIST_DIR/dmg-content/使用说明.txt" << 'README'
FFmpeg Binary 服务
==================

安装方式 1 (推荐):
-----------------
双击 "安装.command" 文件,按提示操作

安装方式 2 (手动):
-----------------
1. 将 FFmpeg-Binary.app 拖到 /Applications/ 文件夹
2. 打开终端,运行:
   /Applications/FFmpeg-Binary.app/Contents/MacOS/ffmpeg-binary install
3. 服务将自动启动并开机自启

注意事项:
--------
• 服务需要 FFmpeg 才能工作,请先安装:
  brew install ffmpeg

• 服务只监听 127.0.0.1,确保安全

• 默认端口范围: 18888-28888

查看日志:
--------
~/Library/Logs/ffmpeg-binary.log

卸载:
----
1. 停止服务:
   pkill -f ffmpeg-binary
2. 删除自启动:
   /Applications/FFmpeg-Binary.app/Contents/MacOS/ffmpeg-binary uninstall
3. 删除应用:
   rm -rf /Applications/FFmpeg-Binary.app
README

echo "    ✅ 安装脚本已创建"

# 创建 DMG
echo "==> 创建 DMG 安装包..."

if command -v create-dmg &> /dev/null; then
    # 使用 create-dmg (更美观)
    create-dmg \
        --volname "$APP_NAME" \
        --volicon "$ICON_FILE" \
        --window-pos 200 120 \
        --window-size 800 500 \
        --icon-size 100 \
        --icon "FFmpeg-Binary.app" 200 200 \
        --hide-extension "FFmpeg-Binary.app" \
        --icon "安装.command" 500 200 \
        --background assets/dmg-background.png \
        "$DIST_DIR/$DMG_NAME" \
        "$DIST_DIR/dmg-content" \
        2>/dev/null || {
            # 如果失败,使用简单模式
            create-dmg \
                --volname "$APP_NAME" \
                --window-pos 200 120 \
                --window-size 800 500 \
                "$DIST_DIR/$DMG_NAME" \
                "$DIST_DIR/dmg-content"
        }
    echo "    ✅ 使用 create-dmg 创建"
else
    # 使用 hdiutil (系统自带)
    echo "    使用 hdiutil 创建(安装 create-dmg 可获得更好效果)"
    hdiutil create -volname "$APP_NAME" \
                   -srcfolder "$DIST_DIR/dmg-content" \
                   -ov -format UDZO \
                   "$DIST_DIR/$DMG_NAME"
    echo "    💡 安装 create-dmg 获得更好效果: brew install create-dmg"
fi

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                 ✅ 打包完成!                                ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "📦 DMG 文件: $DIST_DIR/$DMG_NAME"
echo "📂 应用包: $DIST_DIR/$APP_BUNDLE"
echo ""
echo "使用方法:"
echo "  1. 打开 DMG 文件"
echo "  2. 双击 '安装.command'"
echo "  3. 服务自动启动并配置开机自启"
echo ""