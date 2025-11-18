#!/bin/bash
# FFmpeg Binary 清理监控脚本
# 当应用被移到废纸篓时立即自动清理

APP_PATH="/Applications/FFmpeg-Binary.app"
LAUNCH_AGENT_PLIST="$HOME/Library/LaunchAgents/com.ffmpeg.binary.plist"
WATCHER_PLIST="$HOME/Library/LaunchAgents/com.ffmpeg.binary.watcher.plist"
WATCHER_SCRIPT="$HOME/Library/Application Support/FFmpeg-Binary/cleanup-watcher.sh"
DATA_DIR="$HOME/.ffmpeg-binary"

# 检查应用是否还在应用程序文件夹
if [ ! -d "$APP_PATH" ]; then
    # 应用不在应用程序文件夹了(可能被移到废纸篓或删除),立即清理
    echo "$(date): 检测到应用已被移除,开始清理..."

    # 停止主服务
    pkill -f ffmpeg-binary 2>/dev/null || true
    echo "$(date): 已停止服务进程"

    # 移除主服务的 LaunchAgent
    if [ -f "$LAUNCH_AGENT_PLIST" ]; then
        launchctl unload "$LAUNCH_AGENT_PLIST" 2>/dev/null || true
        rm -f "$LAUNCH_AGENT_PLIST"
        echo "$(date): 已移除主服务 LaunchAgent"
    fi

    # 清理数据目录
    if [ -d "$DATA_DIR" ]; then
        rm -rf "$DATA_DIR"
        echo "$(date): 已清理数据目录"
    fi

    # 移除自己的 LaunchAgent
    if [ -f "$WATCHER_PLIST" ]; then
        launchctl unload "$WATCHER_PLIST" 2>/dev/null || true
        rm -f "$WATCHER_PLIST"
        echo "$(date): 已移除监控服务 plist"
    fi

    # 删除监控脚本自身和目录
    if [ -f "$WATCHER_SCRIPT" ]; then
        rm -f "$WATCHER_SCRIPT"
        echo "$(date): 已删除监控脚本"
    fi

    # 删除 Application Support 目录
    SUPPORT_DIR="$HOME/Library/Application Support/FFmpeg-Binary"
    if [ -d "$SUPPORT_DIR" ]; then
        rm -rf "$SUPPORT_DIR"
        echo "$(date): 已删除 Application Support 目录"
    fi

    echo "$(date): 清理完成"
fi
