@echo off
chcp 65001 > nul
title 复制发布文件到 releases 目录

echo ==========================================
echo        复制构建文件到 releases 目录
echo ==========================================
echo.

:: 设置版本号
set VERSION=v1.0.1

:: 清理并创建 releases 目录
if exist "releases" (
    rmdir /s /q "releases"
    echo [完成] 已清理 releases 目录
)
mkdir releases
echo [完成] 创建 releases 目录
echo.

:: 复制构建文件
echo [步骤 1/3] 复制构建文件...
if exist "build\bin\releases\netser-windows-amd64.exe" (
    copy "build\bin\releases\netser-windows-amd64.exe" "releases\"
    echo [完成] 复制 netser-windows-amd64.exe
) else (
    echo [错误] 找不到 netser-windows-amd64.exe
)

if exist "build\bin\releases\netser-windows-arm64.exe" (
    copy "build\bin\releases\netser-windows-arm64.exe" "releases\"
    echo [完成] 复制 netser-windows-arm64.exe
) else (
    echo [错误] 找不到 netser-windows-arm64.exe
)
echo.

:: 创建压缩包
echo [步骤 2/3] 创建压缩包...
cd releases

if exist "netser-windows-amd64.exe" (
    powershell -Command "Compress-Archive -Path 'netser-windows-amd64.exe' -DestinationPath 'netser-windows-amd64-%VERSION%.zip' -Force"
    echo [完成] 创建 netser-windows-amd64-%VERSION%.zip
)

if exist "netser-windows-arm64.exe" (
    powershell -Command "Compress-Archive -Path 'netser-windows-arm64.exe' -DestinationPath 'netser-windows-arm64-%VERSION%.zip' -Force"
    echo [完成] 创建 netser-windows-arm64-%VERSION%.zip
)

:: 创建发布说明
echo [步骤 3/3] 创建发布说明...
(
echo # Netser %VERSION% 发布包
echo.
echo ## 📦 文件列表
echo - netser-windows-amd64.exe ^(Windows x64 可执行文件^)
echo - netser-windows-arm64.exe ^(Windows ARM64 可执行文件^)
echo - netser-windows-amd64-%VERSION%.zip ^(Windows x64 压缩包^)
echo - netser-windows-arm64-%VERSION%.zip ^(Windows ARM64 压缩包^)
echo.
echo ## 🚀 使用方法
echo 1. 根据你的系统架构下载对应文件
echo 2. 解压压缩包 ^(如果下载的是 .zip 文件^)
echo 3. 双击运行 .exe 文件
echo 4. 开始使用 Netser 网络调试工具！
echo.
echo ## ✨ 功能特性
echo - 🌐 TCP 客户端/服务端通信
echo - 🔌 串口设备通信支持  
echo - 💾 消息持久化存储 ^(BoltDB^)
echo - 🚀 实时数据传输 ^(WebSocket^)
echo - 🎯 十六进制数据支持
echo - 🎨 现代化界面 ^(Vue3 + Element Plus^)
echo.
echo ## 📋 系统要求
echo - Windows 10/11
echo - x64 或 ARM64 架构
echo.
echo 构建时间: %date% %time%
echo 版本: %VERSION%
echo 项目地址: https://github.com/zhoudm1743/Netser
) > RELEASE-README.md

cd ..

echo.
echo ==========================================
echo            复制完成！
echo ==========================================
echo.

:: 显示结果
echo [📁 发布文件]
dir releases
echo.

echo [📊 文件大小统计]
for %%f in (releases\*) do (
    for %%s in (%%f) do (
        set /a "size_mb=%%~zs/1048576"
        echo   %%~nxf: %%~zs 字节 ^(!size_mb! MB^)
    )
)
echo.

echo [📂 发布包位置] %cd%\releases\
echo.
echo [✅ 下一步操作]
echo 1. 检查 releases 目录中的文件
echo 2. 上传到 GitHub Release 页面
echo 3. 或直接分发给用户使用
echo.
echo 按任意键退出...
pause > nul 