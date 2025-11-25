#!/bin/bash

# 本地开发调试脚本
# 直接运行服务,无需安装

set -e

echo "======================================"
echo "  本地开发模式 - GoalfyMediaConverter"
echo "======================================"
echo ""

# 获取脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# 创建必要的目录
DATA_DIR="$HOME/Library/Application Support/GoalfyMediaConverter-Dev"
TEMP_DIR="$HOME/Library/Caches/GoalfyMediaConverter-Dev/temp"
OUTPUT_DIR="$HOME/Library/Caches/GoalfyMediaConverter-Dev/output"
LOG_DIR="$HOME/Library/Logs"

echo "📁 创建开发目录..."
mkdir -p "$DATA_DIR"
mkdir -p "$TEMP_DIR"
mkdir -p "$OUTPUT_DIR"
mkdir -p "$LOG_DIR"

# 检查 FFmpeg
if ! command -v ffmpeg &> /dev/null; then
    echo "⚠️  未找到 ffmpeg,请先安装:"
    echo "   brew install ffmpeg"
    exit 1
fi

FFMPEG_PATH=$(which ffmpeg)
echo "✅ FFmpeg 路径: $FFMPEG_PATH"

# 编译
echo ""
echo "🔨 编译项目..."
go build -o ffmpeg-binary-dev

if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi

echo "✅ 编译成功"
echo ""

# 设置环境变量
export GOALFY_DATA_DIR="$DATA_DIR"
export GOALFY_TEMP_DIR="$TEMP_DIR"
export GOALFY_OUTPUT_DIR="$OUTPUT_DIR"
export GOALFY_FFMPEG_PATH="$FFMPEG_PATH"
export GOALFY_PORT="8080"
export GOALFY_HOST="0.0.0.0"
export GOALFY_DEV_MODE="true"  # 开发模式,跳过自清理监控

# 显示配置
echo "======================================"
echo "  配置信息"
echo "======================================"
echo "数据目录: $DATA_DIR"
echo "临时目录: $TEMP_DIR"
echo "输出目录: $OUTPUT_DIR"
echo "FFmpeg:   $FFMPEG_PATH"
echo "端口:     8080"
echo "日志:     $LOG_DIR/goalfy-mediaconverter-dev.log"
echo "======================================"
echo ""

# 运行服务
echo "🚀 启动服务..."
echo ""
echo "💡 提示:"
echo "   - 日志输出到: $LOG_DIR/goalfy-mediaconverter-dev.log"
echo "   - 实时查看日志: tail -f $LOG_DIR/goalfy-mediaconverter-dev.log"
echo "   - 按 Ctrl+C 停止服务"
echo ""

# 启动服务,日志输出到文件和控制台
./ffmpeg-binary-dev 2>&1 | tee "$LOG_DIR/goalfy-mediaconverter-dev.log"