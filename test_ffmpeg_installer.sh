#!/bin/bash
# FFmpeg 自动安装功能测试脚本

echo "================================================"
echo "FFmpeg 自动安装功能测试"
echo "================================================"
echo ""

# 1. 检查当前 FFmpeg 状态
echo "1️⃣  检查当前 FFmpeg 状态..."
if command -v ffmpeg &> /dev/null; then
    echo "   ✅ FFmpeg 已安装: $(which ffmpeg)"
    ffmpeg -version | head -1
else
    echo "   ❌ FFmpeg 未安装"
fi
echo ""

# 2. 构建测试二进制
echo "2️⃣  构建测试二进制..."
go build -o ffmpeg-binary-test .
if [ $? -eq 0 ]; then
    echo "   ✅ 构建成功"
else
    echo "   ❌ 构建失败"
    exit 1
fi
echo ""

# 3. 测试 FFmpeg 检测和安装功能
echo "3️⃣  测试服务启动(包含 FFmpeg 自动检测)..."
echo "   提示: 如果 FFmpeg 未安装,程序会自动安装"
echo "   按 Ctrl+C 可以停止测试"
echo ""
./ffmpeg-binary-test

echo ""
echo "================================================"
echo "测试完成!"
echo "================================================"