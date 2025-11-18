#!/bin/bash
# 图标生成脚本 - 从 SVG 生成 macOS ICNS 和 Windows ICO

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
ASSETS_DIR="$PROJECT_DIR/assets/icons"
SVG_FILE="$ASSETS_DIR/icon.svg"

echo "==> 生成应用图标..."

# 检查依赖
if ! command -v convert &> /dev/null && ! command -v sips &> /dev/null; then
    echo "❌ 需要 ImageMagick (convert) 或 macOS (sips)"
    echo "   安装: brew install imagemagick"
    exit 1
fi

# 创建临时目录
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

echo "==> 从 SVG 生成 PNG 图标..."

# 生成各种尺寸的 PNG (macOS 需要)
sizes=(16 32 64 128 256 512 1024)

for size in "${sizes[@]}"; do
    if command -v rsvg-convert &> /dev/null; then
        # 使用 rsvg-convert (更好的质量)
        rsvg-convert -w $size -h $size "$SVG_FILE" -o "$TMP_DIR/icon_${size}x${size}.png"
    elif command -v convert &> /dev/null; then
        # 使用 ImageMagick
        convert -background none -resize ${size}x${size} "$SVG_FILE" "$TMP_DIR/icon_${size}x${size}.png"
    elif command -v sips &> /dev/null; then
        # macOS 自带工具(需要先转换 SVG)
        qlmanage -t -s $size -o "$TMP_DIR" "$SVG_FILE" 2>/dev/null
        mv "$TMP_DIR/$(basename "$SVG_FILE").png" "$TMP_DIR/icon_${size}x${size}.png" 2>/dev/null || true
    fi
    echo "   生成 ${size}x${size}"
done

# 生成 macOS ICNS
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "==> 生成 macOS ICNS 图标..."
    ICONSET_DIR="$TMP_DIR/icon.iconset"
    mkdir -p "$ICONSET_DIR"

    # 复制并重命名为 macOS iconset 格式
    cp "$TMP_DIR/icon_16x16.png" "$ICONSET_DIR/icon_16x16.png"
    cp "$TMP_DIR/icon_32x32.png" "$ICONSET_DIR/icon_16x16@2x.png"
    cp "$TMP_DIR/icon_32x32.png" "$ICONSET_DIR/icon_32x32.png"
    cp "$TMP_DIR/icon_64x64.png" "$ICONSET_DIR/icon_32x32@2x.png"
    cp "$TMP_DIR/icon_128x128.png" "$ICONSET_DIR/icon_128x128.png"
    cp "$TMP_DIR/icon_256x256.png" "$ICONSET_DIR/icon_128x128@2x.png"
    cp "$TMP_DIR/icon_256x256.png" "$ICONSET_DIR/icon_256x256.png"
    cp "$TMP_DIR/icon_512x512.png" "$ICONSET_DIR/icon_256x256@2x.png"
    cp "$TMP_DIR/icon_512x512.png" "$ICONSET_DIR/icon_512x512.png"
    cp "$TMP_DIR/icon_1024x1024.png" "$ICONSET_DIR/icon_512x512@2x.png"

    # 生成 ICNS
    iconutil -c icns "$ICONSET_DIR" -o "$ASSETS_DIR/icon.icns"
    echo "   ✅ 生成 icon.icns"
fi

# 生成 Windows ICO
echo "==> 生成 Windows ICO 图标..."
if command -v convert &> /dev/null; then
    convert "$TMP_DIR/icon_16x16.png" \
            "$TMP_DIR/icon_32x32.png" \
            "$TMP_DIR/icon_64x64.png" \
            "$TMP_DIR/icon_128x128.png" \
            "$TMP_DIR/icon_256x256.png" \
            "$ASSETS_DIR/icon.ico"
    echo "   ✅ 生成 icon.ico"
else
    echo "   ⚠️  需要 ImageMagick 生成 ICO 文件"
    echo "      安装: brew install imagemagick"
fi

# 保存高分辨率 PNG 用于 Linux
cp "$TMP_DIR/icon_512x512.png" "$ASSETS_DIR/icon.png"
echo "   ✅ 生成 icon.png (512x512)"

echo ""
echo "=== 图标生成完成 ==="
echo "macOS: $ASSETS_DIR/icon.icns"
echo "Windows: $ASSETS_DIR/icon.ico"
echo "Linux: $ASSETS_DIR/icon.png"