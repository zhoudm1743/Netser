@echo off
chcp 65001 > nul
title Netser 自动构建发布脚本

echo ==========================================
echo    Netser 网络通信调试工具 - 构建脚本
echo ==========================================
echo.

:: 设置版本号
set VERSION=v1.0.0
set BUILD_DATE=%date:~0,10%

echo [信息] 开始构建版本: %VERSION%
echo [信息] 构建日期: %BUILD_DATE%
echo.

:: 清理旧的构建文件
echo [步骤 1/6] 清理旧构建文件...
if exist "build" (
    rmdir /s /q "build"
    echo [完成] 已清理 build 目录
) else (
    echo [跳过] build 目录不存在
)

if exist "releases" (
    rmdir /s /q "releases"
    echo [完成] 已清理 releases 目录
) else (
    echo [跳过] releases 目录不存在
)

mkdir releases
echo [完成] 创建 releases 目录
echo.

:: 构建 Windows amd64 版本
echo [步骤 2/6] 构建 Windows amd64 版本...
wails build -platform windows/amd64 -o releases/netser-windows-amd64.exe
if %ERRORLEVEL% NEQ 0 (
    echo [错误] Windows amd64 构建失败！
    pause
    exit /b 1
)
echo [完成] Windows amd64 构建成功
echo.

:: 构建 Windows arm64 版本  
echo [步骤 3/6] 构建 Windows arm64 版本...
wails build -platform windows/arm64 -o releases/netser-windows-arm64.exe
if %ERRORLEVEL% NEQ 0 (
    echo [错误] Windows arm64 构建失败！
    pause
    exit /b 1
)
echo [完成] Windows arm64 构建成功
echo.

:: 复制构建文件到 releases 目录
echo [步骤 4/6] 复制构建文件...
if exist "build\bin\releases\netser-windows-amd64.exe" (
    copy "build\bin\releases\netser-windows-amd64.exe" "releases\"
    echo [完成] 复制 Windows amd64 可执行文件
)

if exist "build\bin\releases\netser-windows-arm64.exe" (
    copy "build\bin\releases\netser-windows-arm64.exe" "releases\"
    echo [完成] 复制 Windows arm64 可执行文件
)
echo.

:: 创建压缩包
echo [步骤 5/6] 创建压缩包...
cd releases

:: 创建 Windows amd64 压缩包
if exist "netser-windows-amd64.exe" (
    powershell -Command "Compress-Archive -Path 'netser-windows-amd64.exe' -DestinationPath 'netser-windows-amd64-%VERSION%.zip' -Force"
    echo [完成] 创建 netser-windows-amd64-%VERSION%.zip
)

:: 创建 Windows arm64 压缩包
if exist "netser-windows-arm64.exe" (
    powershell -Command "Compress-Archive -Path 'netser-windows-arm64.exe' -DestinationPath 'netser-windows-arm64-%VERSION%.zip' -Force"
    echo [完成] 创建 netser-windows-arm64-%VERSION%.zip
)

:: 创建 README 文件
echo [步骤 6/6] 创建发布说明...
(
echo # Netser %VERSION% 发布包
echo.
echo ## 文件说明
echo - `netser-windows-amd64.exe` - Windows x64 可执行文件
echo - `netser-windows-arm64.exe` - Windows ARM64 可执行文件  
echo - `netser-windows-amd64-%VERSION%.zip` - Windows x64 压缩包
echo - `netser-windows-arm64-%VERSION%.zip` - Windows ARM64 压缩包
echo.
echo ## 使用方法
echo 1. 下载对应架构的文件
echo 2. 解压压缩包^(可选^)
echo 3. 双击运行 .exe 文件
echo 4. 开始使用网络调试功能！
echo.
echo ## 功能特性
echo - TCP 客户端/服务端通信
echo - 串口设备通信支持
echo - 消息持久化存储
echo - 实时数据传输
echo - 十六进制数据支持
echo.
echo 构建时间: %BUILD_DATE%
echo 版本: %VERSION%
) > RELEASE-README.txt

cd ..

echo.
echo ==========================================
echo              构建完成！
echo ==========================================
echo.

:: 显示构建结果
echo [构建结果]
dir releases /b
echo.

:: 显示文件大小
echo [文件大小]
for %%f in (releases\*) do (
    for %%s in (%%f) do echo %%~nxf: %%~zs 字节
)
echo.

echo [发布包位置] %cd%\releases\
echo.
echo [下一步操作]
echo 1. 检查 releases 目录中的文件
echo 2. 上传到 GitHub Release
echo 3. 或分发给用户使用
echo.
echo 按任意键退出...
pause > nul 