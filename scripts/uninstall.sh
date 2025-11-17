#!/bin/bash
# FFmpeg Binary 完整卸载脚本
# 用途: 完全清理 FFmpeg Binary 服务及其所有配置

set -e

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║         FFmpeg Binary 卸载工具                             ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# 获取当前用户
CURRENT_USER="${USER}"
if [ -z "$CURRENT_USER" ]; then
    CURRENT_USER=$(stat -f "%Su" /dev/console)
fi

USER_HOME=$(eval echo ~$CURRENT_USER)
PLIST_PATH="$USER_HOME/Library/LaunchAgents/com.ffmpeg.binary.plist"
APP_PATH="/Applications/FFmpeg-Binary.app"
CONFIG_DIR="$USER_HOME/.ffmpeg-binary"

echo "检测到用户: $CURRENT_USER"
echo ""

# 步骤 1: 停止并卸载 launchd 服务
echo "1️⃣  停止 launchd 服务..."
if [ -f "$PLIST_PATH" ]; then
    sudo -u "$CURRENT_USER" launchctl unload "$PLIST_PATH" 2>/dev/null || echo "   服务未运行,已跳过"
    rm -f "$PLIST_PATH"
    echo "   ✅ launchd 服务已卸载"
else
    echo "   ⚠️  未找到 launchd 配置,已跳过"
fi
echo ""

# 步骤 2: 停止正在运行的进程
echo "2️⃣  停止服务进程..."
if pgrep -f "ffmpeg-binary" > /dev/null 2>&1; then
    pkill -f "ffmpeg-binary" || true
    sleep 2
    # 强制杀死残留进程
    pkill -9 -f "ffmpeg-binary" 2>/dev/null || true
    echo "   ✅ 服务进程已停止"
else
    echo "   ⚠️  未找到运行中的进程,已跳过"
fi
echo ""

# 步骤 3: 删除应用程序
echo "3️⃣  删除应用程序..."
if [ -d "$APP_PATH" ]; then
    rm -rf "$APP_PATH"
    echo "   ✅ 应用程序已删除: $APP_PATH"
else
    echo "   ⚠️  应用程序不存在,已跳过"
fi
echo ""

# 步骤 4: 清理配置和数据(可选)
echo "4️⃣  清理配置和数据..."
read -p "   是否删除用户数据和配置? (包括已转换的视频) [y/N]: " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -d "$CONFIG_DIR" ]; then
        rm -rf "$CONFIG_DIR"
        echo "   ✅ 配置和数据已删除: $CONFIG_DIR"
    else
        echo "   ⚠️  配置目录不存在,已跳过"
    fi
else
    echo "   ⏭️  保留用户数据"
fi
echo ""

# 步骤 5: 清理日志文件(可选)
echo "5️⃣  清理日志文件..."
LOG_FILES=(
    "$USER_HOME/Library/Logs/ffmpeg-binary.log"
    "$USER_HOME/Library/Logs/ffmpeg-binary.err"
)

for log_file in "${LOG_FILES[@]}"; do
    if [ -f "$log_file" ]; then
        rm -f "$log_file"
        echo "   ✅ 已删除: $log_file"
    fi
done
echo ""

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║               ✅ 卸载完成!                                  ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "已清理的项目:"
echo "  ✓ launchd 自启动服务"
echo "  ✓ 运行中的进程"
echo "  ✓ 应用程序文件"
echo "  ✓ 日志文件"
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "  ✓ 用户配置和数据"
fi
echo ""
echo "感谢使用 FFmpeg Binary!"
echo ""
