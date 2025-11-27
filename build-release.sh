#!/bin/bash

# Netser 跨平台构建脚本 (Linux/macOS)
# 设置颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 设置 wails 路径
WAILS_BIN="${GOPATH:-$HOME/go}/bin/wails"

# 检查 wails 是否存在
if [ ! -f "$WAILS_BIN" ]; then
    echo -e "${RED}[错误]${NC} 未找到 wails 命令: $WAILS_BIN"
    echo -e "${YELLOW}[提示]${NC} 请先安装 Wails: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    exit 1
fi

echo -e "${BLUE}[信息]${NC} 使用 Wails: $WAILS_BIN"

echo "=========================================="
echo "   Netser 网络通信调试工具 - 构建脚本"
echo "=========================================="
echo ""

# 设置版本号
VERSION="v1.0.1"
BUILD_DATE=$(date '+%Y-%m-%d')

echo -e "${BLUE}[信息]${NC} 开始构建版本: ${VERSION}"
echo -e "${BLUE}[信息]${NC} 构建日期: ${BUILD_DATE}"
echo ""

# 清理旧的构建文件
echo -e "${YELLOW}[步骤 1/8]${NC} 清理旧构建文件..."
if [ -d "build" ]; then
    rm -rf build
    echo -e "${GREEN}[完成]${NC} 已清理 build 目录"
else
    echo -e "${BLUE}[跳过]${NC} build 目录不存在"
fi

if [ -d "releases" ]; then
    rm -rf releases
    echo -e "${GREEN}[完成]${NC} 已清理 releases 目录"
else
    echo -e "${BLUE}[跳过]${NC} releases 目录不存在"
fi

mkdir -p releases
echo -e "${GREEN}[完成]${NC} 创建 releases 目录"
echo ""

# 构建 Windows amd64 版本
echo -e "${YELLOW}[步骤 2/8]${NC} 构建 Windows amd64 版本..."
if $WAILS_BIN build -platform windows/amd64 -o releases/netser-windows-amd64.exe; then
    echo -e "${GREEN}[完成]${NC} Windows amd64 构建成功"
else
    echo -e "${RED}[错误]${NC} Windows amd64 构建失败！"
    exit 1
fi
echo ""

# 构建 Windows arm64 版本
echo -e "${YELLOW}[步骤 3/8]${NC} 构建 Windows arm64 版本..."
if $WAILS_BIN build -platform windows/arm64 -o releases/netser-windows-arm64.exe; then
    echo -e "${GREEN}[完成]${NC} Windows arm64 构建成功"
else
    echo -e "${RED}[错误]${NC} Windows arm64 构建失败！"
    exit 1
fi
echo ""

# 构建 Linux amd64 版本
echo -e "${YELLOW}[步骤 4/8]${NC} 构建 Linux amd64 版本..."
if $WAILS_BIN build -platform linux/amd64 -o releases/netser-linux-amd64; then
    echo -e "${GREEN}[完成]${NC} Linux amd64 构建成功"
else
    echo -e "${RED}[错误]${NC} Linux amd64 构建失败！"
    exit 1
fi
echo ""

# 构建 Linux arm64 版本
echo -e "${YELLOW}[步骤 5/8]${NC} 构建 Linux arm64 版本..."
if $WAILS_BIN build -platform linux/arm64 -o releases/netser-linux-arm64; then
    echo -e "${GREEN}[完成]${NC} Linux arm64 构建成功"
else
    echo -e "${RED}[错误]${NC} Linux arm64 构建失败！"
    exit 1
fi
echo ""

# 构建 macOS universal 版本
echo -e "${YELLOW}[步骤 6/8]${NC} 构建 macOS universal 版本..."
if $WAILS_BIN build -platform darwin/universal -o releases/Netser.app; then
    echo -e "${GREEN}[完成]${NC} macOS universal 构建成功"
else
    echo -e "${RED}[错误]${NC} macOS universal 构建失败！"
    exit 1
fi
echo ""

# 复制构建文件
echo -e "${YELLOW}[步骤 7/8]${NC} 复制构建文件..."

if [ -f "build/bin/netser-windows-amd64.exe" ]; then
    cp build/bin/netser-windows-amd64.exe releases/
    echo -e "${GREEN}[完成]${NC} 复制 Windows amd64 可执行文件"
fi

if [ -f "build/bin/netser-windows-arm64.exe" ]; then
    cp build/bin/netser-windows-arm64.exe releases/
    echo -e "${GREEN}[完成]${NC} 复制 Windows arm64 可执行文件"
fi

if [ -f "build/bin/netser-linux-amd64" ]; then
    cp build/bin/netser-linux-amd64 releases/
    chmod +x releases/netser-linux-amd64
    echo -e "${GREEN}[完成]${NC} 复制 Linux amd64 可执行文件"
fi

if [ -f "build/bin/netser-linux-arm64" ]; then
    cp build/bin/netser-linux-arm64 releases/
    chmod +x releases/netser-linux-arm64
    echo -e "${GREEN}[完成]${NC} 复制 Linux arm64 可执行文件"
fi

if [ -d "build/bin/Netser.app" ]; then
    cp -R build/bin/Netser.app releases/
    echo -e "${GREEN}[完成]${NC} 复制 macOS 应用程序包"
fi
echo ""

# 创建压缩包
echo -e "${YELLOW}[步骤 8/8]${NC} 创建压缩包..."
cd releases

# Windows 压缩包
if [ -f "netser-windows-amd64.exe" ]; then
    zip -q "netser-windows-amd64-${VERSION}.zip" netser-windows-amd64.exe
    echo -e "${GREEN}[完成]${NC} 创建 netser-windows-amd64-${VERSION}.zip"
fi

if [ -f "netser-windows-arm64.exe" ]; then
    zip -q "netser-windows-arm64-${VERSION}.zip" netser-windows-arm64.exe
    echo -e "${GREEN}[完成]${NC} 创建 netser-windows-arm64-${VERSION}.zip"
fi

# Linux 压缩包
if [ -f "netser-linux-amd64" ]; then
    tar -czf "netser-linux-amd64-${VERSION}.tar.gz" netser-linux-amd64
    echo -e "${GREEN}[完成]${NC} 创建 netser-linux-amd64-${VERSION}.tar.gz"
fi

if [ -f "netser-linux-arm64" ]; then
    tar -czf "netser-linux-arm64-${VERSION}.tar.gz" netser-linux-arm64
    echo -e "${GREEN}[完成]${NC} 创建 netser-linux-arm64-${VERSION}.tar.gz"
fi

# macOS 压缩包
if [ -d "Netser.app" ]; then
    zip -qr "netser-macos-universal-${VERSION}.zip" Netser.app
    echo -e "${GREEN}[完成]${NC} 创建 netser-macos-universal-${VERSION}.zip"
fi

# 创建 README 文件
echo ""
echo "创建发布说明..."
cat > RELEASE-README.md << EOF
# Netser ${VERSION} 发布包

## 文件说明

### Windows
- \`netser-windows-amd64.exe\` - Windows x64 可执行文件
- \`netser-windows-arm64.exe\` - Windows ARM64 可执行文件
- \`netser-windows-amd64-${VERSION}.zip\` - Windows x64 压缩包
- \`netser-windows-arm64-${VERSION}.zip\` - Windows ARM64 压缩包

### Linux
- \`netser-linux-amd64\` - Linux x64 可执行文件
- \`netser-linux-arm64\` - Linux ARM64 可执行文件
- \`netser-linux-amd64-${VERSION}.tar.gz\` - Linux x64 压缩包
- \`netser-linux-arm64-${VERSION}.tar.gz\` - Linux ARM64 压缩包

### macOS
- \`Netser.app\` - macOS 应用程序包（支持 Intel 和 Apple Silicon）
- \`netser-macos-universal-${VERSION}.zip\` - macOS 压缩包

## 使用方法

### Windows
1. 下载对应架构的 .exe 文件或压缩包
2. 如果下载的是压缩包，解压缩
3. 双击运行 .exe 文件

### Linux
1. 下载对应架构的文件或压缩包
2. 添加执行权限: \`chmod +x netser-linux-amd64\`
3. 运行: \`./netser-linux-amd64\`

### macOS
1. 下载 .app 文件或压缩包
2. 如果下载的是压缩包，解压缩
3. 将 Netser.app 拖到应用程序文件夹
4. 首次运行可能需要在系统设置中允许运行

## 功能特性
- TCP 客户端/服务端通信
- 串口设备通信支持
- 消息持久化存储
- 实时数据传输
- 十六进制数据支持
- 跨平台支持（Windows、Linux、macOS）

构建时间: ${BUILD_DATE}
版本: ${VERSION}
EOF

cd ..

echo ""
echo "=========================================="
echo "              构建完成！"
echo "=========================================="
echo ""

# 显示构建结果
echo -e "${BLUE}[构建结果]${NC}"
ls -lh releases/ | grep -v "^total" | grep -v "^d"
echo ""

echo -e "${GREEN}[发布包位置]${NC} $(pwd)/releases/"
echo ""
echo -e "${YELLOW}[下一步操作]${NC}"
echo "1. 检查 releases 目录中的文件"
echo "2. 上传到 GitHub Release"
echo "3. 或分发给用户使用"
echo ""
