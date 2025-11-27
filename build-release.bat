@echo off
chcp 65001 > nul
title Netser 自动构建发布脚本

echo ==========================================
echo    Netser 网络通信调试工具 - 构建脚本
echo ==========================================
echo.

:: 设置版本号
set VERSION=v1.0.1
set BUILD_DATE=%date:~0,10%

echo [信息] 开始构建版本: %VERSION%
echo [信息] 构建日期: %BUILD_DATE%
echo.

:: 清理旧的构建文件
echo [步骤 1/8] 清理旧构建文件...
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
echo [步骤 2/8] 构建 Windows amd64 版本...
wails build -platform windows/amd64 -o releases/netser-windows-amd64.exe
if %ERRORLEVEL% NEQ 0 (
    echo [错误] Windows amd64 构建失败！
    pause
    exit /b 1
)
echo [完成] Windows amd64 构建成功
echo.

:: 构建 Windows arm64 版本  
echo [步骤 3/8] 构建 Windows arm64 版本...
wails build -platform windows/arm64 -o releases/netser-windows-arm64.exe
if %ERRORLEVEL% NEQ 0 (
    echo [错误] Windows arm64 构建失败！
    pause
    exit /b 1
)
echo [完成] Windows arm64 构建成功
echo.

:: 构建 Linux amd64 版本
echo [步骤 4/8] 构建 Linux amd64 版本...
wails build -platform linux/amd64 -o releases/netser-linux-amd64
if %ERRORLEVEL% NEQ 0 (
    echo [错误] Linux amd64 构建失败！
    pause
    exit /b 1
)
echo [完成] Linux amd64 构建成功
echo.

:: 构建 Linux arm64 版本
echo [步骤 5/8] 构建 Linux arm64 版本...
wails build -platform linux/arm64 -o releases/netser-linux-arm64
if %ERRORLEVEL% NEQ 0 (
    echo [错误] Linux arm64 构建失败！
    pause
    exit /b 1
)
echo [完成] Linux arm64 构建成功
echo.

:: 构建 macOS universal 版本（支持 Intel 和 Apple Silicon）
echo [步骤 6/8] 构建 macOS universal 版本...
wails build -platform darwin/universal -o releases/Netser.app
if %ERRORLEVEL% NEQ 0 (
    echo [错误] macOS universal 构建失败！
    pause
    exit /b 1
)
echo [完成] macOS universal 构建成功
echo.

:: 复制构建文件到 releases 目录
echo [步骤 7/8] 复制构建文件...
if exist "build\bin\releases\netser-windows-amd64.exe" (
    copy "build\bin\releases\netser-windows-amd64.exe" "releases\"
    echo [完成] 复制 Windows amd64 可执行文件
)

if exist "build\bin\releases\netser-windows-arm64.exe" (
    copy "build\bin\releases\netser-windows-arm64.exe" "releases\"
    echo [完成] 复制 Windows arm64 可执行文件
)
echo.

:: 复制 Linux 和 macOS 构建文件
if exist "build\bin\releases\netser-linux-amd64" (
    copy "build\bin\releases\netser-linux-amd64" "releases\"
    echo [完成] 复制 Linux amd64 可执行文件
)

if exist "build\bin\releases\netser-linux-arm64" (
    copy "build\bin\releases\netser-linux-arm64" "releases\"
    echo [完成] 复制 Linux arm64 可执行文件
)

if exist "build\bin\releases\Netser.app" (
    xcopy /E /I /Y "build\bin\releases\Netser.app" "releases\Netser.app"
    echo [完成] 复制 macOS 应用程序包
)
echo.

:: 创建压缩包
echo [步骤 8/8] 创建压缩包...
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

:: 创建 Linux amd64 压缩包
if exist "netser-linux-amd64" (
    powershell -Command "Compress-Archive -Path 'netser-linux-amd64' -DestinationPath 'netser-linux-amd64-%VERSION%.tar.gz' -Force"
    echo [完成] 创建 netser-linux-amd64-%VERSION%.tar.gz
)

:: 创建 Linux arm64 压缩包
if exist "netser-linux-arm64" (
    powershell -Command "Compress-Archive -Path 'netser-linux-arm64' -DestinationPath 'netser-linux-arm64-%VERSION%.tar.gz' -Force"
    echo [完成] 创建 netser-linux-arm64-%VERSION%.tar.gz
)

:: 创建 macOS 压缩包
if exist "Netser.app" (
    powershell -Command "Compress-Archive -Path 'Netser.app' -DestinationPath 'netser-macos-universal-%VERSION%.zip' -Force"
    echo [完成] 创建 netser-macos-universal-%VERSION%.zip
)

:: 创建 README 文件
echo.
echo 创建发布说明...
(
echo # Netser %VERSION% 发布包
echo.
echo ## 文件说明
echo.
echo ### Windows
echo - `netser-windows-amd64.exe` - Windows x64 可执行文件
echo - `netser-windows-arm64.exe` - Windows ARM64 可执行文件  
echo - `netser-windows-amd64-%VERSION%.zip` - Windows x64 压缩包
echo - `netser-windows-arm64-%VERSION%.zip` - Windows ARM64 压缩包
echo.
echo ### Linux
echo - `netser-linux-amd64` - Linux x64 可执行文件
echo - `netser-linux-arm64` - Linux ARM64 可执行文件
echo - `netser-linux-amd64-%VERSION%.tar.gz` - Linux x64 压缩包
echo - `netser-linux-arm64-%VERSION%.tar.gz` - Linux ARM64 压缩包
echo.
echo ### macOS
echo - `Netser.app` - macOS 应用程序包^(支持 Intel 和 Apple Silicon^)
echo - `netser-macos-universal-%VERSION%.zip` - macOS 压缩包
echo.
echo ## 使用方法
echo.
echo ### Windows
echo 1. 下载对应架构的 .exe 文件或压缩包
echo 2. 如果下载的是压缩包，解压缩
echo 3. 双击运行 .exe 文件
echo.
echo ### Linux
echo 1. 下载对应架构的文件或压缩包
echo 2. 添加执行权限: `chmod +x netser-linux-amd64`
echo 3. 运行: `./netser-linux-amd64`
echo.
echo ### macOS
echo 1. 下载 .app 文件或压缩包
echo 2. 如果下载的是压缩包，解压缩
echo 3. 将 Netser.app 拖到应用程序文件夹
echo 4. 首次运行可能需要在系统设置中允许运行
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