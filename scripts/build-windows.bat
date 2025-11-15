@echo off
REM Windows 打包脚本

setlocal

set APP_NAME=FFmpeg-Binary
set BUILD_DIR=build\windows
set INSTALLER_NAME=FFmpeg-Binary-Setup.exe

echo === 开始构建 Windows 应用 ===

REM 清理旧的构建
if exist "%BUILD_DIR%" rmdir /s /q "%BUILD_DIR%"
mkdir "%BUILD_DIR%"

REM 构建可执行文件
echo === 编译 Windows 可执行文件 ===
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w -H windowsgui" -o "%BUILD_DIR%\ffmpeg-binary.exe" .

if errorlevel 1 (
    echo 编译失败!
    exit /b 1
)

echo === 编译成功 ===

REM 创建安装目录结构
mkdir "%BUILD_DIR%\installer"

REM 创建安装说明
echo FFmpeg Binary Service > "%BUILD_DIR%\installer\README.txt"
echo. >> "%BUILD_DIR%\installer\README.txt"
echo 安装步骤: >> "%BUILD_DIR%\installer\README.txt"
echo 1. 将 ffmpeg-binary.exe 复制到 C:\Program Files\FFmpeg-Binary\ >> "%BUILD_DIR%\installer\README.txt"
echo 2. 运行 ffmpeg-binary.exe install 安装自启动 >> "%BUILD_DIR%\installer\README.txt"
echo 3. 服务将自动启动 >> "%BUILD_DIR%\installer\README.txt"
echo. >> "%BUILD_DIR%\installer\README.txt"
echo 注意: 需要单独下载 FFmpeg.exe 并放置在同目录的 bin 文件夹中 >> "%BUILD_DIR%\installer\README.txt"
echo 下载地址: https://www.gyan.dev/ffmpeg/builds/ >> "%BUILD_DIR%\installer\README.txt"

REM 创建安装脚本
echo @echo off > "%BUILD_DIR%\installer\install.bat"
echo echo 正在安装 FFmpeg Binary 服务... >> "%BUILD_DIR%\installer\install.bat"
echo. >> "%BUILD_DIR%\installer\install.bat"
echo set INSTALL_DIR=C:\Program Files\FFmpeg-Binary >> "%BUILD_DIR%\installer\install.bat"
echo. >> "%BUILD_DIR%\installer\install.bat"
echo mkdir "%%INSTALL_DIR%%" >> "%BUILD_DIR%\installer\install.bat"
echo copy ffmpeg-binary.exe "%%INSTALL_DIR%%\" >> "%BUILD_DIR%\installer\install.bat"
echo. >> "%BUILD_DIR%\installer\install.bat"
echo cd /d "%%INSTALL_DIR%%" >> "%BUILD_DIR%\installer\install.bat"
echo ffmpeg-binary.exe install >> "%BUILD_DIR%\installer\install.bat"
echo. >> "%BUILD_DIR%\installer\install.bat"
echo echo 安装完成! >> "%BUILD_DIR%\installer\install.bat"
echo echo 服务已启动 >> "%BUILD_DIR%\installer\install.bat"
echo pause >> "%BUILD_DIR%\installer\install.bat"

REM 复制可执行文件到安装器目录
copy "%BUILD_DIR%\ffmpeg-binary.exe" "%BUILD_DIR%\installer\"

echo.
echo === 构建完成 ===
echo 可执行文件: %BUILD_DIR%\ffmpeg-binary.exe
echo 安装器目录: %BUILD_DIR%\installer\
echo.
echo 注意:
echo 1. 需要手动下载 FFmpeg for Windows
echo 2. 可以使用 NSIS 或 Inno Setup 创建专业安装包
echo 3. 当前提供简单的批处理安装脚本
echo.

endlocal