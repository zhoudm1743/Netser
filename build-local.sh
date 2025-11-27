#!/bin/bash

# Netser 本地构建脚本 - 仅构建当前平台
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

# 设置 wails 路径
WAILS_BIN="${GOPATH:-$HOME/go}/bin/wails"

echo -e "${BLUE}[信息]${NC} 开始构建 Netser v1.0.1 (仅当前平台)"
echo ""

# 清理旧的构建文件
if [ -d "build" ]; then
    rm -rf build
    echo -e "${GREEN}[完成]${NC} 已清理 build 目录"
fi

# 构建当前平台
echo ""
echo -e "${BLUE}[信息]${NC} 构建当前平台版本..."
$WAILS_BIN build

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}[完成]${NC} 构建成功！"
    echo ""
    echo "可执行文件位置："
    ls -lh build/bin/
else
    echo ""
    echo -e "\033[0;31m[错误]\033[0m 构建失败！"
    exit 1
fi
