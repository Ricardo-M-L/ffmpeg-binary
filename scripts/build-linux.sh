#!/bin/bash
# Linux 打包脚本 - 生成 DEB/RPM 安装包

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

APP_NAME="ffmpeg-binary"
VERSION="1.0.0"
DIST_DIR="dist/linux"
ICON_FILE="assets/icons/icon.png"

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║           Linux 软件包打包工具                              ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# 清理旧文件
echo "==> 清理旧文件..."
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# 编译二进制
echo "==> 编译 Linux 二进制文件..."
echo "    架构: amd64"

GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "$DIST_DIR/$APP_NAME" .
chmod +x "$DIST_DIR/$APP_NAME"

echo "    ✅ 二进制文件已生成"

# 创建 DEB 包结构
echo "==> 创建 DEB 包..."
DEB_DIR="$DIST_DIR/deb"
mkdir -p "$DEB_DIR/DEBIAN"
mkdir -p "$DEB_DIR/usr/local/bin"
mkdir -p "$DEB_DIR/usr/share/applications"
mkdir -p "$DEB_DIR/usr/share/icons/hicolor/512x512/apps"
mkdir -p "$DEB_DIR/usr/share/doc/$APP_NAME"

# 复制二进制文件
cp "$DIST_DIR/$APP_NAME" "$DEB_DIR/usr/local/bin/"

# 复制图标
if [ -f "$ICON_FILE" ]; then
    cp "$ICON_FILE" "$DEB_DIR/usr/share/icons/hicolor/512x512/apps/$APP_NAME.png"
fi

# 创建 control 文件
cat > "$DEB_DIR/DEBIAN/control" << EOF
Package: $APP_NAME
Version: $VERSION
Section: utils
Priority: optional
Architecture: amd64
Depends: ffmpeg
Maintainer: FFmpeg Binary Team
Description: FFmpeg Video Conversion Service
 A local service for converting WebM videos to MP4 format.
 Provides both synchronous and asynchronous conversion APIs.
EOF

# 创建 postinst 脚本(安装后执行)
cat > "$DEB_DIR/DEBIAN/postinst" << 'EOF'
#!/bin/bash
set -e

echo "正在配置 FFmpeg Binary 服务..."

# 为当前用户安装自启动
if [ -n "$SUDO_USER" ]; then
    sudo -u $SUDO_USER /usr/local/bin/ffmpeg-binary install
    # 启动服务
    sudo -u $SUDO_USER nohup /usr/local/bin/ffmpeg-binary > /dev/null 2>&1 &
else
    /usr/local/bin/ffmpeg-binary install
    nohup /usr/local/bin/ffmpeg-binary > /dev/null 2>&1 &
fi

echo "FFmpeg Binary 服务已安装并启动"
echo "服务地址: http://127.0.0.1:28888 (端口可能不同,请查看日志)"

exit 0
EOF

chmod 755 "$DEB_DIR/DEBIAN/postinst"

# 创建 prerm 脚本(卸载前执行)
cat > "$DEB_DIR/DEBIAN/prerm" << 'EOF'
#!/bin/bash
set -e

echo "正在停止 FFmpeg Binary 服务..."

# 停止服务
pkill -f ffmpeg-binary || true

# 卸载自启动
/usr/local/bin/ffmpeg-binary uninstall || true

exit 0
EOF

chmod 755 "$DEB_DIR/DEBIAN/prerm"

# 创建 .desktop 文件
cat > "$DEB_DIR/usr/share/applications/$APP_NAME.desktop" << EOF
[Desktop Entry]
Version=1.0
Type=Application
Name=FFmpeg Binary Service
Comment=Video conversion service
Exec=/usr/local/bin/$APP_NAME
Icon=$APP_NAME
Terminal=false
Categories=Utility;
EOF

# 创建文档
cat > "$DEB_DIR/usr/share/doc/$APP_NAME/README" << 'EOF'
FFmpeg Binary Service
=====================

This service provides WebM to MP4 video conversion.

Usage:
  The service runs automatically after installation on port 28888.

  API Documentation:
  - POST /api/v1/convert/sync - Synchronous conversion
  - POST /api/v1/convert/async - Asynchronous conversion

Configuration:
  ~/.ffmpeg-binary/config.json

Logs:
  Check system logs or ~/.ffmpeg-binary/ directory

Uninstall:
  sudo apt remove ffmpeg-binary
EOF

# 构建 DEB 包
dpkg-deb --build "$DEB_DIR" "$DIST_DIR/${APP_NAME}_${VERSION}_amd64.deb"

echo "    ✅ DEB 包已创建"

# 创建 RPM 包(如果安装了 rpm 工具)
if command -v rpmbuild &> /dev/null; then
    echo "==> 创建 RPM 包..."

    RPM_DIR="$DIST_DIR/rpm"
    mkdir -p "$RPM_DIR"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

    # 创建 spec 文件
    cat > "$RPM_DIR/SPECS/$APP_NAME.spec" << SPEC_EOF
Name:           $APP_NAME
Version:        $VERSION
Release:        1%{?dist}
Summary:        FFmpeg Video Conversion Service
License:        MIT
Requires:       ffmpeg

%description
A local service for converting WebM videos to MP4 format.

%install
mkdir -p %{buildroot}/usr/local/bin
cp $DIST_DIR/$APP_NAME %{buildroot}/usr/local/bin/

%files
/usr/local/bin/$APP_NAME

%post
/usr/local/bin/$APP_NAME install
nohup /usr/local/bin/$APP_NAME > /dev/null 2>&1 &

%preun
pkill -f $APP_NAME || true
/usr/local/bin/$APP_NAME uninstall || true

SPEC_EOF

    rpmbuild -bb --define "_topdir $RPM_DIR" "$RPM_DIR/SPECS/$APP_NAME.spec"

    # 复制 RPM 到 dist 目录
    find "$RPM_DIR/RPMS" -name "*.rpm" -exec cp {} "$DIST_DIR/" \;

    echo "    ✅ RPM 包已创建"
else
    echo "    ℹ️  跳过 RPM 包(未安装 rpmbuild)"
fi

# 创建通用安装脚本
echo "==> 创建通用安装脚本..."
cat > "$DIST_DIR/install.sh" << 'INSTALL_EOF'
#!/bin/bash

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║         FFmpeg Binary 服务安装程序                          ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# 检查是否为 root
if [ "$EUID" -ne 0 ]; then
    echo "⚠️  请使用 root 权限运行"
    echo "   sudo ./install.sh"
    exit 1
fi

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# 检查 FFmpeg
if ! command -v ffmpeg &> /dev/null; then
    echo "⚠️  未检测到 FFmpeg"
    echo ""
    echo "请先安装 FFmpeg:"
    echo "  Ubuntu/Debian: sudo apt install ffmpeg"
    echo "  CentOS/RHEL:   sudo yum install ffmpeg"
    echo "  Fedora:        sudo dnf install ffmpeg"
    echo ""
    read -p "是否继续安装? (y/N): " choice
    if [ "$choice" != "y" ] && [ "$choice" != "Y" ]; then
        exit 0
    fi
fi

echo "==> 正在安装..."

# 复制二进制文件
echo "    • 复制程序文件"
cp "$SCRIPT_DIR/ffmpeg-binary" /usr/local/bin/
chmod +x /usr/local/bin/ffmpeg-binary

# 为当前用户安装自启动
echo "    • 配置开机自启动"
if [ -n "$SUDO_USER" ]; then
    sudo -u $SUDO_USER /usr/local/bin/ffmpeg-binary install
else
    /usr/local/bin/ffmpeg-binary install
fi

# 启动服务
echo "    • 启动服务"
if [ -n "$SUDO_USER" ]; then
    sudo -u $SUDO_USER nohup /usr/local/bin/ffmpeg-binary > /dev/null 2>&1 &
else
    nohup /usr/local/bin/ffmpeg-binary > /dev/null 2>&1 &
fi

sleep 2

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                 ✅ 安装成功!                                ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "服务已启动并设置为开机自启"
echo "默认端口: 28888"
echo ""
echo "查看日志: ~/.ffmpeg-binary/"
echo ""

INSTALL_EOF

chmod +x "$DIST_DIR/install.sh"

echo "    ✅ 安装脚本已创建"

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                 ✅ 打包完成!                                ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "📦 二进制文件: $DIST_DIR/$APP_NAME"
echo "📦 DEB 包: $DIST_DIR/${APP_NAME}_${VERSION}_amd64.deb"

if [ -f "$DIST_DIR"/*.rpm ]; then
    echo "📦 RPM 包: $DIST_DIR/*.rpm"
fi

echo "📜 通用安装脚本: $DIST_DIR/install.sh"
echo ""
echo "使用方法:"
echo "  Debian/Ubuntu: sudo dpkg -i ${APP_NAME}_${VERSION}_amd64.deb"
echo "  CentOS/RHEL:   sudo rpm -i ${APP_NAME}-${VERSION}-1.*.rpm"
echo "  通用方法:      sudo ./install.sh"
echo ""