# Amadeus 项目概览

## 1. 项目简介

Amadeus 是一个基于 [Cloudwego Eino](https://github.com/cloudwego/eino) 框架的 CLI AI 助手，采用**手动编排**（Manual Orchestration）模式——即只使用 Eino 的底层模型能力和工具接口，由业务代码自行控制推理轮次、工具调度、上下文拼装与终止判断。

### 核心能力

- **流式对话输出**：支持 streaming 和非 streaming 两种模式
- **工具调用**：集成 MCP (Model Context Protocol) 工具 + 内置基础工具
- **Skill 系统**：按需加载业务指令，注入对话上下文
- **会话持久化**：每次会话的对话记录、trace 数据均持久化到本地文件系统

### 技术栈

| 组件 | 选型 |
|------|------|
| 语言 | Go 1.25 |
| AI 框架 | Cloudwego Eino v0.7.32 |
| 模型 API | OpenAI-compatible（默认 DeepSeek） |
| MCP 客户端 | mark3labs/mcp-go v0.43.2 |
| 配置加载 | godotenv v1.5.1 |

---

## 2. 目录结构

```
Amadeus/
├── cmd/amadeus/               # CLI 入口
│   └── main.go
├── internal/                  # 核心模块
│   ├── model/                 # 模型层：ChatModel 创建与配置
│   ├── orchestrator/          # 编排层：多轮对话循环、工具调度
│   ├── tool/                  # 工具层：MCP 客户端、基础工具注册与执行
│   │   └── basetools/         # 内置基础工具
│   ├── memory/                # 记忆层：会话与 trace 的持久化存储
│   ├── session/               # 会话状态：消息历史与 Skill 激活状态
│   ├── skill/                 # Skill 配置加载与内容读取
│   └── presentation/          # 表现层：CLI 输入读取、事件输出与格式化渲染
├── skills/                    # Skill 注册表与业务 skill 文档
│   ├── agent.md               # Skill 注册表（name + desc 列表）
│   └── metadata-platform-cli/ # 示例 skill
│       └── SKILL.md
├── tools/
│   └── toolsConfig.json       # MCP 工具配置文件
├── checkpoints/sessions/      # 会话与 trace 存储目录
├── docs/                      # 设计与实现文档
└── openspec/                  # 变更提案与规格
```

---

## 3. 核心架构

### 3.1 整体流程

```
┌──────────┐     ┌──────────────┐     ┌────────────┐     ┌──────────┐
│  CLI 输入 │────▶│  Orchestrator │────▶│   ChatModel │────▶│   输出    │
│ (present) │     │  (loop.go)   │     │ (deepseek) │     │ (renderer)│
└──────────┘     └──────┬───────┘     └────────────┘     └──────────┘
                        │
                        ▼
                 ┌──────────────┐
                 │   Executor    │
                 │ (tool map)    │
                 └──────┬───────┘
                        │
          ┌─────────────┼─────────────┐
          ▼             ▼             ▼
   ┌──────────┐  ┌──────────┐  ┌──────────┐
   │  bash    │  │load_skill│  │ MCP 工具  │
   └──────────┘  └──────────┘  └──────────┘
```

### 3.2 启动流程（cmd/amadeus/main.go）

1. 加载 `.env` 环境变量
2. 加载 Skill 配置（`agent.md` 路径 + skill 根目录）
3. 解析 `agent.md` 作为 system prompt 中的 skill 注册表
4. 创建 ChatModel（DeepSeek / OpenAI-compatible）
5. 加载工具（内置工具 `bash`、`load_skill` + MCP 工具）
6. 创建 Tool Executor（工具名 → InvokableTool 映射）
7. 创建 Orchestrator（绑定模型与工具）
8. 进入 CLI 循环：读取用户输入 → `HandleTurn` → 渲染输出

---

## 4. 模块详解

### 4.1 模型层（internal/model/）

**chat_model.go**

- 通过环境变量配置模型 `DEEPSEEK_*` 系列参数
- `BuildSystemMessage()` 将 `agent.md` 内容拼接到 system prompt 末尾，使模型感知可用 skill
- `GetChatModel()` 创建 OpenAI-compatible ChatModel 实例

**config.go**

- `ChatModelSettings` 结构体承载模型配置
- 支持环境变量：`DEEPSEEK_MODEL`、`DEEPSEEK_BASE_URL`、`DEEPSEEK_THINKING_TYPE`、`DEEPSEEK_REASONING_EFFORT`、`DEEPSEEK_STREAM`

**reasoning_payload.go**

- 通过 `ReasoningPayloadOption()` 注入 `reasoning_content` 字段到请求 payload
- 确保 DeepSeek 等模型的推理内容（thinking）能在多轮对话中被正确保留和传递

### 4.2 编排层（internal/orchestrator/）

**orchestrator.go**

- `Orchestrator` 结构体聚合：`chatModel`（模型接口）、`executor`（工具执行器）、`store`（会话存储）
- `New()` 初始化时将工具绑定到模型
- `HandleTurn()` 作为外部入口，每次用户输入触发一轮完整处理

**loop.go** — 核心多轮循环

1. 从 `store` 加载历史对话和已加载的 skill
2. 重建 `session.State`（system message + 历史消息 + 当前问题）
3. 持久化用户消息
4. 调用 `run()` 进入多轮循环（最大轮次由 `AMADEUS_MAX_TURNS` 控制，默认 8）
5. 每轮通过 `streamModelTurn()` 调用模型，支持流式 / 非流式
6. 如果模型返回 tool_calls，依次执行工具并将结果回填到消息历史
7. 如果 `load_skill` 工具返回成功，将 skill 内容作为 system message 注入状态
8. 当模型返回不含 tool_calls 且 content 非空时，视为最终回复，结束循环

**parser.go**

- `ParseToolArguments()` 对 tool 参数做最小校验：非空且合法 JSON

**policy.go**

- `loadMaxTurns()` 从环境变量读取最大对话轮次，默认 8

### 4.3 工具层（internal/tool/）

**registry.go**

- `LoadInvokableTools()` 聚合基础工具和 MCP 工具
- 先加载 `basetools.Load()` 返回的内置工具，再遍历 MCP 配置创建客户端并获取工具

**executor.go**

- `Executor` 维护工具名 → `InvokableTool` 的映射
- `Execute()` 执行工具并返回标准化的 `Result` 结构（ToolName、Success、Data、Error）
- 工具执行失败时仍返回结构化结果，让编排器可以将错误回填为 tool message

**mcp.go**

- `MCPServerConfig`：定义 MCP 服务器的命令、参数和环境变量
- `ToolsConfig`：从 JSON 文件解析 `mcpServers` 配置
- `CreateMcpClientsFromConfig()`：为每个 MCP 服务器创建 stdio 客户端并初始化
- `resolveConfigEnv()`：从环境变量展开配置中的 `${VAR}` 占位符

**basetools/bash.go**

- 内置 bash 工具，支持设置 workdir 和 timeout_seconds
- 输出标准格式化：command、workdir、exit_code、stdout、stderr
- 输出截断保护（最大 32KB），超时保护（最长 60 秒）

**basetools/load_skill.go**

- 内置 `load_skill` 工具，按 skill 名称加载 `SKILL.md` 内容
- 参数：`name`（skill 名称）
- 返回：`skill.Document`（Name、Path、Content）

### 4.4 会话状态（internal/session/）

**state.go**

- `State` 结构体维护：消息列表、当前轮次、工具调用计数、完成标志、最后工具结果、已加载 skill
- `NewState()`：从历史记录和已加载 skill 构建完整的消息序列
- `ActivateSkill()`：将 skill 文档作为 system message 追加到消息列表，去重

### 4.5 记忆层（internal/memory/）

**store.go**

- `Store` 使用 JSONL 格式将每一轮对话和 trace 持久化到 `checkpoints/sessions/{session_id}/` 目录
- 每次启动自动生成 `session_id`（格式：`YYYYMMDD-HHMMSS-随机4字节hex`）
- 三个文件：
  - `meta.json`：会话元信息（session_id、started_at、model、base_url）
  - `conversation.jsonl`：用户消息和助手最终回复（用于重建对话历史）
  - `trace.jsonl`：详细 trace 记录（请求消息、模型响应、错误）
  - `loaded_skills.jsonl`：已加载的 skill 记录

### 4.6 Skill 系统（internal/skill/）

**config.go**

- `Config` 结构体：`AgentMDPath`（skill 注册表路径）、`SkillRootPath`（skill 文档根目录）
- 通过环境变量配置：`SKILL_AGENT_MD_REL` / `SKILL_AGENT_MD_ABS`、`SKILL_ROOT_REL` / `SKILL_ROOT_ABS`
- 支持相对路径和绝对路径两种配置方式

**registry.go**

- `LoadAgentMarkdown()`：读取 `agent.md` 文件内容并做基本校验（非空、包含 name 和 desc）

**loader.go**

- `LoadSkillContent()`：按 skill 名称验证并读取对应的 `SKILL.md` 文件
- 安全校验：路径逃逸检查，防止目录穿越

### 4.7 表现层（internal/presentation/）

**event.go**

- 定义事件类型常量：TurnStarted、ReasoningDelta、AnswerDelta、AssistantFinal、ToolCallStarted、ToolCallFinished、TurnFailed

**input.go**

- `ReadUserInput()`：从 stdin 读取一行用户输入

**output.go**

- `PrintToolCall()`、`PrintToolResult()`、`PrintAssistantResponse()`：通过 `Emit()` 发送事件

**renderer.go**

- `Renderer` 是 CLI 输出渲染的核心，支持两种视图模式：
  - **chat 模式**（默认）：简洁对话式输出，reasoning 用灰色 `•` 前缀，answer 用 `>` 前缀，工具调用灰色显示
  - **trace 模式**（`AMADEUS_CLI_VIEW=trace`）：详细展示每个工具调用的参数、状态、结果
- 支持终端宽度自适应（`COLUMNS` 环境变量）
- reasoning 内容用 ANSI 灰色 (`\033[90m`) 渲染，与 answer 区分
- 自动换行和前缀对齐

---

## 5. 配置参考

### 5.1 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DEEPSEEK_API_KEY` | — | DeepSeek API Key（必填） |
| `DEEPSEEK_MODEL` | `deepseek-v4-flash` | 模型名称 |
| `DEEPSEEK_BASE_URL` | `https://api.deepseek.com` | API 端点 |
| `DEEPSEEK_THINKING_TYPE` | `enabled` | 推理模式 |
| `DEEPSEEK_REASONING_EFFORT` | `medium` | 推理投入度 |
| `DEEPSEEK_STREAM` | `true` | 是否启用流式输出 |
| `AMADEUS_MAX_TURNS` | `8` | 最大对话轮次 |
| `AMADEUS_CLI_VIEW` | `chat` | CLI 视图模式（`chat` / `trace`） |
| `COLUMNS` | `100` | 终端宽度 |
| `SKILL_AGENT_MD_REL` | — | agent.md 相对路径 |
| `SKILL_AGENT_MD_ABS` | — | agent.md 绝对路径 |
| `SKILL_ROOT_REL` | — | skill 根目录相对路径 |
| `SKILL_ROOT_ABS` | — | skill 根目录绝对路径 |
| `AMAP_MAPS_API_KEY` | — | 高德地图 API Key（MCP 工具使用） |

### 5.2 工具配置（tools/toolsConfig.json）

```json
{
  "mcpServers": {
    "amap-maps": {
      "command": "npx",
      "args": ["-y", "@amap/amap-maps-mcp-server"],
      "env": {
        "AMAP_MAPS_API_KEY": "${AMAP_MAPS_API_KEY}"
      }
    }
  }
}
```

环境变量占位符 `${VAR}` 会在启动时自动展开。

### 5.3 Skill 注册表（skills/agent.md）

```markdown
# Skills

- name: metadata-platform-cli
  desc: 当用户需要使用 metadata-platform-service-cli 查询支持的 route_v1 API、查看指定 API 说明、拼装 invoke 参数或执行调用时，使用这个 skill。
```

每个 skill 对应 `skills/{name}/SKILL.md` 文件，由 `load_skill` 工具按需加载。

---

## 6. 核心数据流

### 6.1 一次完整的对话轮次

```
用户输入 "查询所有 API"
       │
       ▼
Orchestrator.HandleTurn()
       │
       ├─ 1. LoadConversation()  ← 从 JSONL 恢复历史
       ├─ 2. LoadLoadedSkills()  ← 恢复已加载 skill
       ├─ 3. NewState()          ← 构建消息序列
       ├─ 4. AppendUserMessage() ← 持久化用户消息
       │
       ▼
   run() 多轮循环（最大 maxTurns 轮）
       │
       ├─ streamModelTurn()
       │   ├─ 调用 ChatModel.Stream() / Generate()
       │   ├─ 逐 chunk 输出 reasoning / answer 事件
       │   └─ persist model response
       │
       ├─ 判断：有 tool_calls 吗？
       │   ├─ 是 → 执行工具
       │   │   ├─ parse arguments
       │   │   ├─ executor.Execute()  ← bash / load_skill / MCP tools
       │   │   ├─ 如果是 load_skill → 解析 Document
       │   │   ├─ 追加 tool message 到 state
       │   │   └─ 如果加载了 skill → ActivateSkill() + persist
       │   │
       │   └─ 否 → Finished = true，返回最终回复
       │
       └─ 超过 maxTurns → 返回错误
```

### 6.2 会话文件结构

```
checkpoints/sessions/20260428-120727-016ebc2d/
├── meta.json                # 会话元信息
├── conversation.jsonl       # 对话记录（user_message / assistant_final）
├── trace.jsonl              # 详细 trace（turn_request / model_response / turn_error）
└── loaded_skills.jsonl      # 已加载 skill 记录
```

---

## 7. 运行方式

### 7.1 构建

```bash
go build -o Amadeus ./cmd/amadeus
```

### 7.2 运行

```bash
# 前置条件：配置 .env 文件或设置环境变量
export DEEPSEEK_API_KEY=sk-xxx

# 直接运行
go run ./cmd/amadeus

# 或编译后运行
./Amadeus
```

### 7.3 查看模式

```bash
# 默认 chat 模式
AMADEUS_CLI_VIEW=chat ./Amadeus

# trace 模式（详细展示工具调用过程）
AMADEUS_CLI_VIEW=trace ./Amadeus
```

---

## 8. Skill 开发指南

编写一个新的 skill 只需两步：

### 8.1 在 `skills/agent.md` 注册

```markdown
- name: my-skill-name
  desc: 用一句话描述该 skill 的适用场景。
```

### 8.2 创建 skill 文档

`skills/my-skill-name/SKILL.md`，内容为 Markdown 格式，包含：

```markdown
---
name: my-skill-name
description: 完整描述
---

# Skill 标题

详细的行为指令、规则、流程和示例。
```

AI 助手在对话中会根据 `agent.md` 中的 desc 判断是否需要加载该 skill，确认后调用 `load_skill` 工具加载完整指令。

---

## 9. 设计原则

1. **手动编排**：推理轮次、工具调度、上下文拼装和终止判断全部由业务代码控制，不依赖框架内置 Agent
2. **最小校验前置**：工具参数在进入执行层前先做 JSON 合法性校验
3. **错误结构化**：工具执行失败仍返回结构化 Result，让模型可以理解失败原因
4. **持久化透明**：所有对话和 trace 数据以 JSONL 格式保存在本地，便于审计和调试
5. **Skill 按需加载**：只有确认需要时才加载完整 skill 指令，避免 system prompt 膨胀
6. **配置优先**：所有可调参数（模型、轮次、路径、视图）均通过环境变量暴露

---

## 10. 相关文档

- [手动编排改造方案（索引）](manual-orchestrator/index.md)
- [手动编排执行计划](manual-orchestrator-execution-plan.md)
- [Skill 技术方案](skill-technical-proposal.md)
- [Agent 实践学习笔记](agent-practice-learning.md)
