@echo off
chcp 65001 > nul
title Netser 发布助手

:main
cls
echo ==========================================
echo           Netser 发布助手
echo ==========================================
echo.
echo 🚀 Netser 网络通信调试工具发布管理
echo.
echo [1] 🔨 完整构建 (清理 + 构建 + 打包)
echo [2] 📦 快速打包 (复制现有构建文件)
echo [3] 📋 创建 GitHub Release
echo [4] 📁 打开 releases 文件夹
echo [5] 🌐 打开 GitHub 仓库
echo [6] ❌ 退出
echo.
echo ==========================================

set /p choice=请选择操作 (1-6): 

if "%choice%"=="1" goto build_all
if "%choice%"=="2" goto copy_files
if "%choice%"=="3" goto github_release
if "%choice%"=="4" goto open_folder
if "%choice%"=="5" goto open_repo
if "%choice%"=="6" goto exit
goto main

:build_all
echo.
echo [选择] 完整构建流程
echo ==========================================
call build-release.bat
echo.
echo 按任意键返回主菜单...
pause > nul
goto main

:copy_files
echo.
echo [选择] 快速打包流程
echo ==========================================
call copy-releases.bat
echo.
echo 按任意键返回主菜单...
pause > nul
goto main

:github_release
echo.
echo [选择] GitHub Release 创建
echo ==========================================
call create-github-release.bat
echo.
echo 按任意键返回主菜单...
pause > nul
goto main

:open_folder
echo.
echo [选择] 打开 releases 文件夹
if exist "releases" (
    start "" "%cd%\releases"
    echo [完成] 已打开 releases 文件夹
) else (
    echo [错误] releases 文件夹不存在，请先运行构建或打包
)
echo.
echo 按任意键返回主菜单...
pause > nul
goto main

:open_repo
echo.
echo [选择] 打开 GitHub 仓库
start "" "https://github.com/zhoudm1743/Netser"
echo [完成] 已打开 GitHub 仓库
echo.
echo 按任意键返回主菜单...
pause > nul
goto main

:exit
echo.
echo 👋 感谢使用 Netser 发布助手！
echo.
exit /b 0 