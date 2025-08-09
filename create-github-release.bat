@echo off
chcp 65001 > nul
title GitHub Release 创建指南

echo ==========================================
echo        GitHub Release 创建指南  
echo ==========================================
echo.

set VERSION=v1.0.0
set REPO_URL=https://github.com/zhoudm1743/Netser

echo [信息] 版本: %VERSION%
echo [信息] 仓库: %REPO_URL%
echo [信息] 发布文件位置: %cd%\releases\
echo.

echo ==========================================
echo              操作步骤
echo ==========================================
echo.

echo [步骤 1] 打开 GitHub 仓库
echo    🌐 在浏览器中打开: %REPO_URL%
echo.

echo [步骤 2] 进入 Releases 页面
echo    📦 点击仓库页面右侧的 "Releases"
echo    📝 点击 "Create a new release"
echo.

echo [步骤 3] 填写 Release 信息
echo    🏷️ Tag version: %VERSION%
echo    📄 Release title: Netser %VERSION% - 网络通信调试工具
echo.

echo [步骤 4] 复制以下内容到 Release Notes:
echo ==========================================
echo.
echo ## 🎉 Netser %VERSION% 首次发布
echo.
echo ### ✨ 功能特性
echo - 🌐 **TCP通信** - 支持TCP客户端和服务端
echo - 🔌 **串口通信** - 完整的串口设备支持
echo - 💾 **消息持久化** - 基于BoltDB的嵌入式数据库
echo - 🚀 **实时通信** - WebSocket实时消息推送  
echo - 🎨 **现代界面** - Vue3 + Element Plus
echo - 🎯 **十六进制支持** - 支持十六进制数据格式
echo.
echo ### 📦 下载文件
echo - **Windows x64**: `netser-windows-amd64-%VERSION%.zip`
echo - **Windows ARM64**: `netser-windows-arm64-%VERSION%.zip`
echo.
echo ### 🚀 快速开始
echo 1. 下载对应平台的压缩包
echo 2. 解压并运行 `netser-windows-*.exe`
echo 3. 开始使用网络调试功能！
echo.
echo ### 📋 系统要求
echo - Windows 10/11
echo - x64 或 ARM64 架构
echo.
echo ### 🛠️ 技术栈
echo - **后端**: Go + Wails + BoltDB + WebSocket
echo - **前端**: Vue 3 + Element Plus + Pinia
echo.
echo ==========================================
echo.

echo [步骤 5] 上传发布文件
echo    📁 将以下文件拖拽到 "Attach binaries" 区域:
echo       - netser-windows-amd64-%VERSION%.zip
echo       - netser-windows-arm64-%VERSION%.zip
echo       - netser-windows-amd64.exe
echo       - netser-windows-arm64.exe
echo.

echo [步骤 6] 发布
echo    ✅ 检查信息无误后，点击 "Publish release"
echo.

echo ==========================================
echo            快捷操作
echo ==========================================
echo.

echo [1] 打开 GitHub Release 页面
echo [2] 打开 releases 文件夹
echo [3] 打开项目仓库
echo [4] 退出
echo.

set /p choice=请选择操作 (1-4): 

if "%choice%"=="1" (
    start "" "%REPO_URL%/releases/new?tag=%VERSION%"
    echo [完成] 已打开 GitHub Release 创建页面
)

if "%choice%"=="2" (
    start "" "%cd%\releases"
    echo [完成] 已打开 releases 文件夹
)

if "%choice%"=="3" (
    start "" "%REPO_URL%"
    echo [完成] 已打开项目仓库
)

if "%choice%"=="4" (
    echo [退出] 再见！
    exit /b 0
)

echo.
echo 按任意键退出...
pause > nul 