# Amadeus

一个基于 Cloudwego Eino 框架的 CLI AI 助手，支持流式对话、工具调用和会话追踪存储。

## 功能特性

- 流式对话输出
- 支持 MCP (Model Context Protocol) 工具集成
- Skill 注入与按需加载
- 会话历史与 trace 记录管理

## 技术栈

- Go 1.25
- Cloudwego Eino v0.7.32
- DeepSeek API
## 安装

```bash
go mod download
```

## 运行

```bash
go run ./cmd/amadeus
```

## 项目结构

```text
Amadeus/
├── cmd/amadeus/          # CLI 入口
├── internal/             # 核心模块（model/orchestrator/presentation/tool/memory/skill/session）
├── skills/               # Skill 注册与业务 skill 文档
├── tools/                # MCP 工具配置
├── checkpoints/sessions/ # 会话与 trace 存储
├── docs/                 # 设计与实现文档
└── openspec/             # 变更提案与规格
```

## 配置

- 通过环境变量配置模型：
  - `DEEPSEEK_API_KEY`
  - `DEEPSEEK_MODEL`，默认 `deepseek-reasoner`
  - `DEEPSEEK_BASE_URL`，默认 `https://api.deepseek.com`
- 启动时会自动从 `.env` 读取本地开发环境变量。

在 `tools/toolsConfig.json` 中配置 MCP 工具。

## 使用说明

启动程序后，在终端输入问题即可与 Amadeus 多轮对话。系统会在 `checkpoints/sessions/` 下保存当前会话的对话记录与 trace 数据。
