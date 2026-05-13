# Amadeus — Agent 使用指南

## 概述

Amadeus 是一个基于 [Cloudwego Eino](https://github.com/cloudwego/eino) 框架构建的 Go 语言 CLI AI 助手，通过 OpenAI 兼容接口接入 DeepSeek 模型（默认 `deepseek-v4-flash`），支持流式/非流式对话、工具调用、Skill 按需加载以及会话 trace 持久化。

核心能力：

- **多轮对话**：会话历史跨轮次持久化，重启后自动恢复上下文。
- **工具调用**：内置 `bash`、`cmd`、`load_skill`、`search_memory` 工具；支持通过 MCP 协议挂载外部工具（配置于 `tools/toolsConfig.json`）。
- **Skill 系统**：模型启动时仅获知 skill 目录（`skills/agent.md`），在需要时通过 `load_skill` 按需加载完整的 `SKILL.md` 内容，并在当前会话中持久激活。
- **RAG 记忆**：每轮对话结束后异步将 user/assistant 消息 embedding 写入 Milvus；模型可通过 `search_memory` 工具跨 session 检索语义相关的历史对话片段。Milvus 不可用时自动降级。
- **会话存储**：每次启动生成唯一 Session ID，在 `checkpoints/sessions/<session-id>/` 下写入 `conversation.jsonl`、`trace.jsonl`、`loaded_skills.jsonl`、`meta.json`。

---

## 启动命令

```bash
# 安装依赖
go mod download

# 运行程序
go run ./cmd/amadeus

# 运行所有测试
go test ./...

# 运行单个包的测试
go test ./internal/orchestrator/...

# 运行单个测试函数
go test ./internal/orchestrator/... -run TestHandleTurnPersistsTraceAndRestoresConversationOnly
```

> 启动前需确保 `.env` 文件存在（或环境变量已设置）。

---

## 架构

```
cmd/amadeus/main.go          # CLI 入口：初始化各层，进入读取-响应循环
│
├── internal/model/          # 模型层：通过 OpenAI 兼容接口创建 ChatModel，解析模型配置（config.go），构造 system prompt
├── internal/skill/          # Skill 层：解析 agent.md、按名称加载 SKILL.md
├── internal/tool/           # 工具层：聚合 basetools + MCP 工具，构建 Executor
│   └── basetools/           #   内置工具：bash、cmd、load_skill、search_memory
├── internal/memory/         # 存储层：JSONL 格式会话 & trace 写入/读取；RAG Indexer（Milvus）
├── internal/session/        # 会话状态：单次请求内的 messages 组装与 skill 激活
├── internal/orchestrator/   # 编排层：驱动模型流式推理 + 工具调用循环
└── internal/presentation/   # 展示层：流式输出、工具调用打印、用户输入读取
```

### 关键数据流

```
用户输入
  → orchestrator.HandleTurn
    → memory.Store.LoadConversation + LoadLoadedSkills   # 恢复历史
    → session.NewState                                    # 组装 messages
    → orchestrator.run（最多 AMADEUS_MAX_TURNS 轮）
        → DEEPSEEK_STREAM=true  → model.Stream → 流式输出 reasoning/content   # 流式模式（默认）
        → DEEPSEEK_STREAM=false → model.Generate → 一次性输出                  # 非流式模式
        → 有 tool_calls → executor.Execute               # 执行工具
            → load_skill 成功 → state.ActivateSkill      # 激活 skill
        → 无 tool_calls → 返回最终回答
    → memory.Store.AppendAssistantFinal                  # 持久化对话
    → memory.Indexer.IndexMessages（goroutine）          # 异步 embedding → Milvus
```

### 会话文件结构

每次运行在 `checkpoints/sessions/<session-id>/` 下生成：

| 文件 | 内容 |
|------|------|
| `meta.json` | Session ID、启动时间、模型名、BaseURL |
| `conversation.jsonl` | 仅含 `user_message` / `assistant_final` 记录，用于下轮恢复 |
| `trace.jsonl` | 完整 trace：turn_request、model_response、turn_error |
| `loaded_skills.jsonl` | 已激活的 skill 文档，跨轮次恢复用 |

---

## 本地开发和启动流程

### 环境变量配置

在项目根目录创建 `.env` 文件（启动时自动加载）：

```dotenv
# 必填
DEEPSEEK_API_KEY=<your-api-key>

# 可选，有默认值
DEEPSEEK_MODEL=deepseek-v4-flash          # 默认模型
DEEPSEEK_BASE_URL=https://api.deepseek.com
DEEPSEEK_THINKING_TYPE=enabled            # 思考模式：enabled / disabled
DEEPSEEK_REASONING_EFFORT=medium          # 推理强度：low / medium / high
DEEPSEEK_STREAM=true                      # 是否使用流式输出：true / false
AMADEUS_MAX_TURNS=8                       # 单次用户输入的最大 tool-call 循环轮次
AMADEUS_HISTORY_LIMIT=100                 # 加载到上下文的最大历史消息条数（按 Message 计，默认 100）

# Skill 路径（相对路径优先，绝对路径兜底）
SKILL_AGENT_MD_REL=./skills/agent.md
SKILL_AGENT_MD_ABS=
SKILL_ROOT_REL=./skills
SKILL_ROOT_ABS=

# RAG 记忆（可选，不配置时 search_memory 工具静默降级）
OPENAI_EMBEDDING_API_KEY=               # OpenAI Embedding API Key
OPENAI_EMBEDDING_BASE_URL=https://api.openai.com/v1
OPENAI_EMBEDDING_MODEL=text-embedding-3-small
MILVUS_ADDRESS=localhost:19530          # Milvus gRPC 地址（docker-compose.milvus.yml 启动）
MILVUS_COLLECTION=amadeus_memory        # Collection 名称
```

### 添加新 Skill

1. 在 `skills/<skill-name>/` 目录下创建 `SKILL.md`。
2. 在 `skills/agent.md` 中新增一条记录：
   ```markdown
   - name: <skill-name>
     desc: <一句话描述，告知模型何时应调用此 skill>
   ```
   `agent.md` 必须包含至少一个 `name:` 和 `desc:` 字段，否则启动失败。
3. Skill 名称只允许 `[a-zA-Z0-9_-]`，与目录名保持一致。

### 添加 MCP 工具

编辑 `tools/toolsConfig.json`，在 `mcpServers` 下增加服务项：
f
```json
{
  "mcpServers": {
    "<tool-name>": {
      "command": "npx",
      "args": ["-y", "<mcp-package>"],
      "env": { "API_KEY": "${YOUR_ENV_VAR}" }
    }
  }
}
```

### 代码规范

- **import 顺序**：标准库 → 第三方库 → 项目内部包（`Amadeus/internal/...`）。
- **错误处理**：初始化失败用 `log.Fatal()`；运行时非致命错误用 `log.Printf()` 并向上返回 `error`。
- **工具执行**：工具错误不直接中断编排，`Executor.Execute` 会将错误包装为 `Result{Success: false, Error: ...}` 回填给模型。
- **上下文**：所有网络调用和长耗时函数均接收 `context.Context`，以 `context.Background()` 为根。

---

## 文档导航

| 文档 | 路径 | 内容 |
|------|------|------|
| 项目简介 | `README.md` | 功能特性、技术栈、配置项快速参考 |
| Skill 技术方案 | `docs/skill-technical-proposal.md` | Skill 系统的设计目标、约束与实现思路 |
| 手工编排执行计划 | `docs/manual-orchestrator-execution-plan.md` | Orchestrator 手工编排的设计记录 |
| 变更提案目录 | `openspec/changes/` | 各功能点的提案、设计与任务拆分（openspec 格式） |
| openspec 配置 | `openspec/config.yaml` | openspec schema 与项目上下文配置 |
| Skill 注册表 | `skills/agent.md` | 当前已注册的所有 skill 名称与说明 |
| MCP 工具配置 | `tools/toolsConfig.json` | 外部 MCP 工具的启动命令与环境变量 |
| NapCat 接口文档索引 | `docs/NapCat 接口文档/README.md` | NapCat OneBot 11 HTTP POST 接口总览，含 19 份按章节拆分的子文档 |

---

## 维护约定

> **当项目发生以下变动时，必须同步更新本文件（`AGENTS.md`）：**

| 变动类型 | 需要更新的章节 |
|----------|----------------|
| 新增 / 删除 `internal/` 子包 | 架构 → 包结构图 |
| 新增 / 修改环境变量或默认值 | 本地开发和启动流程 → 环境变量配置 |
| 新增 / 删除内置工具（`basetools/`） | 概述 → 核心能力、架构 → 包结构图 |
| 默认模型或底层 SDK 变更 | 概述、架构 → 模型层注释 |
| 新增 / 删除文档或目录 | 文档导航 |
| 启动流程发生变化 | 启动命令、本地开发和启动流程 |
| Skill 系统设计发生变化 | 架构 → 关键数据流、本地开发和启动流程 → 添加新 Skill |
