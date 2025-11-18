#!/bin/bash
# Windows 打包脚本 - 在 Mac 上运行 (使用 NSIS)
# NSIS 是原生 Linux 工具,支持 ARM64,无需 Wine

set -e

APP_NAME="GoalfyMediaConverter"
BUILD_DIR="build/windows"
DIST_DIR="dist/windows"
DOCKER_IMAGE="ffmpeg-binary-windows-builder-light"
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║     Windows 安装包构建工具 (Mac - 使用 NSIS)               ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ 错误: 未找到 Docker"
    echo ""
    echo "请安装 Docker Desktop:"
    echo "  https://www.docker.com/products/docker-desktop/"
    echo ""
    exit 1
fi

# 检查 Docker 是否运行
if ! docker info &> /dev/null; then
    echo "❌ 错误: Docker 未运行"
    echo ""
    echo "请启动 Docker Desktop,然后重新运行此脚本"
    echo ""
    exit 1
fi

# 清理旧的构建
echo "==> 清理旧文件..."
rm -rf "$PROJECT_ROOT/$BUILD_DIR" "$PROJECT_ROOT/$DIST_DIR"
mkdir -p "$PROJECT_ROOT/$BUILD_DIR" "$PROJECT_ROOT/$DIST_DIR"

# 检查 Docker 镜像是否存在
echo "==> 检查 Docker 镜像..."
if ! docker image inspect "$DOCKER_IMAGE" &> /dev/null; then
    echo "    Docker 镜像不存在,开始构建..."
    echo "    (首次构建需要 2-3 分钟)"
    echo ""

    docker build -t "$DOCKER_IMAGE" -f "$PROJECT_ROOT/scripts/Dockerfile.windows-light" "$PROJECT_ROOT/scripts"

    if [ $? -ne 0 ]; then
        echo ""
        echo "    ❌ Docker 镜像构建失败!"
        exit 1
    fi
    echo ""
    echo "    ✅ Docker 镜像已构建"
else
    echo "    ✅ Docker 镜像已存在"
fi

# 使用 Docker 容器编译 Windows 可执行文件
echo "==> 编译 Windows 可执行文件..."
echo "    架构: amd64"

docker run --rm \
    -v "$PROJECT_ROOT:/workspace" \
    -w /workspace \
    -e CGO_ENABLED=0 \
    -e GOOS=windows \
    -e GOARCH=amd64 \
    "$DOCKER_IMAGE" \
    go build -ldflags="-s -w" -o "$BUILD_DIR/ffmpeg-binary.exe" .

if [ $? -ne 0 ]; then
    echo "    ❌ 编译失败!"
    exit 1
fi
echo "    ✅ 可执行文件已生成"

# 使用 Docker 容器运行 NSIS 构建安装包
echo "==> 创建安装包 (使用 NSIS)..."

docker run --rm \
    -v "$PROJECT_ROOT:/workspace" \
    -w /workspace/scripts \
    "$DOCKER_IMAGE" \
    makensis -V2 setup.nsi

if [ $? -ne 0 ]; then
    echo "    ❌ 安装包创建失败!"
    exit 1
fi

echo "    ✅ 安装包已创建"

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                 ✅ 打包完成!                                ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "📦 安装包: $DIST_DIR/GoalfyMediaConverter-Setup.exe"
echo ""
echo "功能特性:"
echo "  ✅ 图形化安装界面 (NSIS)"
echo "  ✅ 自动下载并安装 FFmpeg"
echo "  ✅ 开机自启动 (可选)"
echo "  ✅ 自动启动服务"
echo "  ✅ 标准卸载程序"
echo "  ✅ 添加到\"添加/删除程序\""
echo ""
echo "用户安装体验:"
echo "  1. 双击 GoalfyMediaConverter-Setup.exe"
echo "  2. 按照安装向导操作"
echo "  3. 安装程序自动下载 FFmpeg"
echo "  4. 服务自动启动"
echo "  5. 访问 http://127.0.0.1:28888"
echo ""
echo "💡 提示: 可以在 Windows 虚拟机或实体机上测试安装包"
echo ""