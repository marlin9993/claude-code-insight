# Claude Code Insight - Standalone 版本

这是一个一体化版本，前端静态资源已经包含在仓库中，并嵌入到 Go 二进制文件中，只需一个可执行文件即可运行完整服务。

## 特性

- **零依赖部署**：前端资源已嵌入二进制文件
- **单文件部署**：编译后只需一个可执行文件

## 编译

```bash
make build-standalone
```

## 使用 go install 安装

```bash
go install github.com/marlin9993/claude-code-insight/cmd/standalone@latest
```

安装后的命令名为 `standalone`：

```bash
standalone
standalone -p 3001
```

### 最小化版本（推荐）

使用 UPX 压缩生成更小的二进制文件：

```bash
make build-minimal
```

**体积对比：**
- 标准版本：~18MB
- 最小化版本：~3.4MB（减少 76%）

> **注意**：需要安装 UPX 才能使用最小化构建。在 Ubuntu/Debian 上安装：
> ```bash
> sudo apt-get install upx
> ```

## 运行

```bash
# 使用默认端口 3000
./bin/claude-insight

# 指定端口
./bin/claude-insight -p 8080
```

## 访问

启动成功后访问：

- **前端页面**：http://localhost:3000
- **API 基础路径**：/api/*
- **健康检查**：http://localhost:3000/health

## 配置

程序会读取 Claude Code 相关环境变量或默认路径，确保 Claude Code 数据目录可访问。

## 故障排除

### 端口被占用

使用 `-p` 参数指定其他端口：

```bash
./bin/claude-insight -p 3001
```
