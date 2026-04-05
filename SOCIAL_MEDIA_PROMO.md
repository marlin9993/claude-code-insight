# Claude Code Insight 社媒宣传稿

这份文案用于在 X、即刻、少数派、Telegram、微信群或朋友圈发布项目介绍。

注意：
- 不要直接使用真实 Claude 会话内容截图
- 不要暴露本机用户名、项目路径、session id、私有仓库名、邮件地址
- 推荐先准备一份演示用数据，再截图

## 一句话版本

我做了一个 `Claude Code Insight`：把本地 Claude Code 历史记录变成可搜索、可统计、可浏览的独立 Web 界面，单个 Go 二进制即可运行。

## 短帖版本

最近做了个小工具：`Claude Code Insight`。

它可以把本地 Claude Code 历史记录整理成一个独立 Web 界面，支持：
- 历史会话浏览
- 关键词搜索
- 项目维度查看
- Token 统计
- 单文件部署

项目本身是 Go 写的，前端静态资源已经打包进程序里，启动后就能直接打开用。

安装：

```bash
go install github.com/marlin9993/claude-code-insight/cmd/standalone@latest
```

源码和 release：
- GitHub: https://github.com/marlin9993/claude-code-insight
- Releases: https://github.com/marlin9993/claude-code-insight/releases

## 长帖版本

我把自己平时看 Claude Code 历史记录时最麻烦的几件事，做成了一个独立工具：`Claude Code Insight`。

它解决的是几个很实际的问题：
- 历史对话是本地文件，但原始格式不适合直接翻
- 会话一多之后，很难按项目、关键词、时间快速定位
- 想看 token 使用趋势、项目使用情况时，原始数据不够直观

现在这个工具可以直接读取本地 Claude Code 历史记录，并提供一个更适合日常使用的界面：
- 浏览历史会话
- 搜索对话内容
- 按项目查看记录
- 看会话、项目、全局 token 统计
- 一个二进制直接运行

技术上也尽量做得简单一些：
- Go 实现
- standalone 方式分发
- 前端资源已打包
- GitHub Actions 自动构建 release

安装方式：

```bash
go install github.com/marlin9993/claude-code-insight/cmd/standalone@latest
```

项目地址：
- GitHub: https://github.com/marlin9993/claude-code-insight
- Releases: https://github.com/marlin9993/claude-code-insight/releases

如果你也经常用 Claude Code，欢迎试试，也欢迎提 issue。

## X / Twitter 版本

Built a small tool for browsing local Claude Code history:

- search conversations
- browse by project
- inspect token usage
- standalone Go binary

`go install github.com/marlin9993/claude-code-insight/cmd/standalone@latest`

GitHub:
https://github.com/marlin9993/claude-code-insight

## 配图建议

推荐只发 3 张图，信息密度够了：

1. 首页总览
- 展示项目列表 / 最近会话 / 总体统计
- 不显示真实项目名

2. 搜索页
- 展示关键词检索结果
- 关键词使用演示词，例如 `refactor`、`bugfix`、`release`

3. 统计页
- 展示 token 趋势图和模型分布
- 这是最容易发且最不泄露正文的图

## 截图脱敏清单

截图前逐项检查：

- 用户名是否出现
- 本机绝对路径是否出现
- 仓库名是否包含私有代号
- 对话正文是否包含客户名、域名、IP、邮箱、密钥
- session id 是否完整暴露
- 浏览器标签页是否含敏感站点
- 系统菜单栏是否含个人头像、邮箱、VPN、公司名称

## 演示数据建议

如果要专门做截图，建议准备一份假的展示数据：

- 项目名：`demo-app`、`cli-tool`、`blog-engine`
- 搜索词：`search`、`cache`、`release`
- 会话摘要：只保留泛化后的技术描述
- token 数据：用聚合图，不放原始正文

## 可直接复制的配文

做了个 `Claude Code Insight`，把本地 Claude Code 历史记录变成了一个更好搜索和浏览的界面。

支持历史会话、项目视图、搜索、token 统计，Go 实现，单个二进制直接运行。

```bash
go install github.com/marlin9993/claude-code-insight/cmd/standalone@latest
```

https://github.com/marlin9993/claude-code-insight

## 标签建议

可选标签：
- `#golang`
- `#opensource`
- `#claudecode`
- `#productivity`
- `#indiehacker`
