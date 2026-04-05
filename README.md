# Claude Code 历史记录查看器

非官方项目，与 Anthropic 无关联，也未获其认可。

这是一个使用 Go 实现的 Claude Code 历史记录查看器。

## 功能特性

- 📝 **历史记录管理** - 读取和管理 Claude Code 的历史对话记录
- 🔍 **搜索功能** - 支持关键词、项目、日期范围搜索
- 📊 **统计信息** - 提供会话和项目的统计信息
- 💰 **Token 统计** - 详细的 token 使用量统计（按会话、项目、全局）
- 🚀 **高性能** - 使用 Go 实现，提供更好的性能
- 📦 **一体化部署** - 前端静态资源已打包，可直接运行

## 服务端口

- **默认端口**: 3000
- **健康检查**: `http://localhost:3000/health`

## API 接口

### 健康检查
- `GET /health` - 服务健康状态

### 历史记录
- `GET /api/history` - 获取历史记录列表（分页）
- `GET /api/history/search` - 搜索历史记录
- `GET /api/history/stats` - 获取统计信息

### 会话
- `GET /api/sessions/:sessionId` - 获取会话详情
- `GET /api/sessions/:sessionId/messages` - 分页获取会话消息

### Token 统计
- `GET /api/tokens/session/:sessionId` - 获取会话 token 统计
- `GET /api/tokens/project/:projectPath` - 获取项目 token 统计
- `GET /api/tokens/global` - 获取全局 token 统计（按天聚合）

## 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `PORT` | 3000 | 服务端口 |
| `NODE_ENV` | production | 运行环境 |
| `CORS_ORIGIN` | http://localhost:5173 | CORS 允许的源 |
| `CLAUDE_HISTORY_PATH` | ~/.claude/history.jsonl | 历史记录文件路径 |
| `CLAUDE_PROJECTS_PATH` | ~/.claude/projects | 项目目录路径 |

## 开发

### 前置要求
- Go 1.23+
- Claude Code 历史记录目录

### 使用 go install 安装
```bash
go install github.com/marlin9993/claude-code-insight/cmd/standalone@latest
```

安装后可直接运行：

```bash
standalone
standalone -p 3001
```

### 安装依赖
```bash
go mod download
```

### 本地运行
```bash
make run
```

### 构建
```bash
make build
```

### 运行测试
```bash
make test
```

## 发布

### 使用 gh 手动创建 release
```bash
git tag v0.1.0
git push origin v0.1.0
gh release create v0.1.0 --verify-tag --generate-notes
```

### 使用 GitHub Actions 自动编译并发布
- 仓库已包含 CI 工作流：push 和 pull request 时自动执行 `go test ./...`
- 仓库已包含 release 工作流：推送 `v*` 标签时自动编译多平台二进制并上传到 GitHub Release

触发方式：

```bash
git tag v0.1.0
git push origin v0.1.0
```

自动发布的二进制目标平台：
- linux amd64
- darwin amd64
- darwin arm64
- windows amd64

## 项目结构

```
.
├── cmd/
│   └── standalone/
│       ├── dist/             # 前端构建产物
│       ├── main.go           # 一体化应用入口
│       └── README.md
├── internal/
│   ├── config/
│   │   └── config.go         # 配置管理
│   ├── controllers/
│   │   ├── history_controller.go
│   │   ├── session_controller.go
│   │   └── token_controller.go
│   ├── services/
│   │   ├── file_reader.go    # JSONL 文件读取
│   │   └── token_stats.go    # Token 统计
│   └── models/
│       └── models.go         # 数据模型
├── Makefile
├── go.mod
└── README.md
```

## 性能优势

- ✅ 更低的内存占用
- ✅ 更快的启动速度
- ✅ 更高效的并发处理
- ✅ 单一二进制文件，易于部署
- ✅ 原生支持高并发场景

## 许可证

MIT
