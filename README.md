# Amadeus

一个基于 Cloudwego Eino 框架的 AI 助手应用，支持流式对话和工具调用。

## 功能特性

- 流式对话输出
- 支持 MCP (Model Context Protocol) 工具集成
- 内置计算器工具
- 对话历史记录管理

## 技术栈

- Go 1.25
- Cloudwego Eino v0.7.32
- DeepSeek API
- MCP Go v0.43.2

## 安装

```bash
go mod download
```

## 运行

```bash
go run main.go
```

## 项目结构

```
Amadeus/
├── agent/          # Agent 配置和初始化
├── checkpoints/    # 对话历史存储
├── prompt/         # 提示词相关
├── tools/          # 工具集成（计算器、MCP客户端等）
└── utils/          # 工具函数（输入、输出、流式处理等）
```

## 配置

在 `agent/agent.go` 中配置 DeepSeek API 密钥和模型参数：

```go
ownerAPIKey   = "your-api-key"
modelURL      = "https://api.deepseek.com"
ModelType     = "deepseek-chat"
```

在 `tools/toolsConfig.json` 中配置 MCP 工具。

## 使用说明

启动程序后，输入问题即可与 Amadeus 进行对话。支持多轮对话，对话历史会自动保存。
