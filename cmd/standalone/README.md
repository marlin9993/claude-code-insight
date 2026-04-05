# Claude Code Insight - Standalone 版本

这是一个一体化版本，将前端静态资源嵌入到 Go 二进制文件中，只需一个可执行文件即可运行完整服务。

## 特性

- **零依赖部署**：前端资源已嵌入二进制文件
- **自动构建前端**：首次运行时自动检测并构建前端
- **单文件部署**：编译后只需一个可执行文件

## 编译

### 标准版本

```bash
# 从 backend-go 目录编译
cd backend-go
make build-standalone
```

### 最小化版本（推荐）

使用 UPX 压缩生成更小的二进制文件：

```bash
cd backend-go
make build-minimal
```

**体积对比：**
- 标准版本：~18MB
- 最小化版本：~3.4MB（减少 76%）

> **注意**：需要安装 UPX 才能使用最小化构建。在 Ubuntu/Debian 上安装：
> ```bash
> sudo apt-get install upx
> ```

### 使用项目根目录的构建脚本

```bash
# 一键构建（最小化版本）
./build.sh
```

## 运行

```bash
# 使用默认端口 3000
./bin/claude-insight

# 指定端口
./bin/claude-insight -p 8080
```

## 首次运行

首次运行时，如果检测到前端未构建，程序会自动：

1. 安装前端依赖（`npm install`）
2. 构建前端（`npm run build`）
3. 复制构建产物到 `dist/` 目录

这个过程可能需要几分钟，请耐心等待。

## 访问

启动成功后访问：

- **前端页面**：http://localhost:3000
- **API 基础路径**：/api/*
- **健康检查**：http://localhost:3000/health

## 配置

程序会读取 `~/.claude-code-insight/config.json` 配置文件，确保配置中的 `claudeCodePath` 指向正确的 Claude Code 数据目录。

## 故障排除

### 前端构建失败

如果自动构建失败，请手动执行：

```bash
cd ../frontend
npm install
npm run build
cd ../backend-go
```

### 端口被占用

使用 `-p` 参数指定其他端口：

```bash
./bin/standalone -p 3001
```

### 找不到前端目录

程序会在当前目录的上级目录查找 `frontend/` 目录。如果路径不同，请修改 `buildFrontend()` 函数中的路径配置。
