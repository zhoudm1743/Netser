# Netser 构建指南

## 本地构建

### Linux/macOS

**快速构建（仅当前平台）：**
```bash
chmod +x build-local.sh
./build-local.sh
```

**跨平台构建（需要额外工具链）：**
```bash
chmod +x build-release.sh
./build-release.sh
```

### Windows

**快速构建（仅当前平台）：**
```cmd
wails build
```

**跨平台构建：**
```cmd
build-release.bat
```

## GitHub Actions 自动构建

### 触发方式

#### 方式 1：创建版本标签（推荐）

```bash
# 创建标签
git tag -a v1.0.1 -m "Release v1.0.1"

# 推送标签（触发自动构建）
git push origin v1.0.1
```

#### 方式 2：手动触发

1. 进入 GitHub 仓库页面
2. 点击 **Actions** 标签
3. 选择 **Release Build** 工作流
4. 点击 **Run workflow** 按钮
5. 选择分支，点击 **Run workflow**

### 构建产物

构建完成后，自动构建会生成以下文件：

**Windows:**
- `netser-windows-amd64.exe.zip`
- `netser-windows-arm64.exe.zip`

**Linux:**
- `netser-linux-amd64.tar.gz`
- `netser-linux-arm64.tar.gz`

**macOS:**
- `Netser.app.zip` (Universal - 支持 Intel 和 Apple Silicon)

### 自动发布

如果是通过标签触发的构建，GitHub Actions 会自动：

1. ✅ 构建所有平台的版本
2. ✅ 创建 GitHub Release
3. ✅ 上传所有构建产物
4. ✅ 生成 Release Notes

你只需要访问：`https://github.com/zhoudm1743/Netser/releases` 就能看到发布包。

## 版本号管理

需要更新版本号时，修改以下文件：

1. `build-release.sh` - 第 29 行
2. `build-release.bat` - 第 11 行
3. `build-local.sh` - 第 11 行
4. `frontend/package.json` - version 字段

建议使用语义化版本号：
- `v1.0.0` - 主版本.次版本.修订号
- `v1.0.1` - Bug 修复
- `v1.1.0` - 新功能
- `v2.0.0` - 重大更新

## 构建要求

### 本地环境

- Go 1.21+
- Node.js 18+
- Wails CLI v2.11.0+

**Linux 额外依赖：**
```bash
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev
```

**macOS 额外依赖：**
```bash
xcode-select --install
```

### GitHub Actions

无需任何配置，GitHub 会自动提供所需的构建环境。

## 常见问题

### 1. 交叉编译失败

**问题：** 在 Linux x86_64 上构建 ARM64 版本失败

**解决：**
- 本地只构建当前平台：使用 `build-local.sh`
- 需要多平台：使用 GitHub Actions 自动构建

### 2. 前端依赖安装失败

**问题：** npm install 失败

**解决：**
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

### 3. Wails 命令找不到

**问题：** `wails: command not found`

**解决：**
```bash
# 安装 Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 添加到 PATH（添加到 ~/.bashrc 或 ~/.zshrc）
export PATH=$PATH:$HOME/go/bin
```

## 快速开始

```bash
# 1. 克隆仓库
git clone git@github.com:zhoudm1743/Netser.git
cd Netser

# 2. 安装依赖
cd frontend && npm install && cd ..

# 3. 本地开发
wails dev

# 4. 本地构建
./build-local.sh

# 5. 发布版本
git tag v1.0.1
git push origin v1.0.1
# 然后访问 GitHub Releases 下载构建产物
```

## 发布流程

1. **更新版本号** - 修改上述 4 个文件
2. **测试构建** - 运行 `./build-local.sh` 确保能正常构建
3. **提交代码** - `git add . && git commit -m "版本更新"`
4. **创建标签** - `git tag -a v1.0.1 -m "Release v1.0.1"`
5. **推送标签** - `git push origin v1.0.1`
6. **等待构建** - 访问 GitHub Actions 查看进度（约 10-15 分钟）
7. **验证发布** - 访问 Releases 页面确认所有文件已上传

## 技术栈

- **框架:** Wails v2
- **后端:** Go 1.21+
- **前端:** Vue 3 + Vite + Element Plus
- **状态管理:** Pinia
- **样式:** SCSS
- **构建:** Wails CLI + GitHub Actions

---

更多信息请参考：
- Wails 文档: https://wails.io
- GitHub Actions 文档: https://docs.github.com/actions
