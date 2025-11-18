#!/bin/bash
# GUI 启动器脚本
# 当用户在启动台或应用程序文件夹点击应用时显示一个简单的状态窗口

# 检查服务是否正在运行
if pgrep -f "ffmpeg-binary" > /dev/null 2>&1; then
    # 服务正在运行,显示状态信息
    osascript <<EOF
tell application "System Events"
    activate
    display dialog "FFmpeg Binary 服务状态

🟢 服务运行中
📡 地址: http://127.0.0.1:28888
📊 健康检查: http://127.0.0.1:28888/health

此服务在后台运行,无需打开此窗口。
要卸载,只需将此应用拖到废纸篓即可。" buttons {"在浏览器中打开", "查看日志", "关闭"} default button "关闭" with icon POSIX file "/Applications/FFmpeg-Binary.app/Contents/Resources/icon.icns"

    set userChoice to button returned of result

    if userChoice is "在浏览器中打开" then
        do shell script "open http://127.0.0.1:28888/health"
    else if userChoice is "查看日志" then
        do shell script "open -a Console ~/Library/Logs/ffmpeg-binary.log"
    end if
end tell
EOF
else
    # 服务未运行,询问是否启动
    RESPONSE=$(osascript <<EOF
tell application "System Events"
    activate
    display dialog "FFmpeg Binary 服务状态

🔴 服务未运行

是否要启动服务?" buttons {"取消", "启动服务"} default button "启动服务" with icon POSIX file "/Applications/FFmpeg-Binary.app/Contents/Resources/icon.icns"
    button returned of result
end tell
EOF
)

    if [ "$RESPONSE" = "启动服务" ]; then
        # 启动服务
        export PATH="/opt/homebrew/bin:/usr/local/bin:$PATH"
        nohup /Applications/FFmpeg-Binary.app/Contents/MacOS/ffmpeg-binary-service > ~/Library/Logs/ffmpeg-binary.log 2>&1 &

        sleep 2

        # 检查是否启动成功
        if pgrep -f "ffmpeg-binary" > /dev/null 2>&1; then
            osascript -e 'display notification "服务已成功启动" with title "FFmpeg Binary"'
        else
            osascript -e 'display alert "启动失败" message "请查看日志文件 ~/Library/Logs/ffmpeg-binary.log"'
        fi
    fi
fi
