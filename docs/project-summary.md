# Amadeus 项目技术总结

## 1. 项目概述

**Amadeus** 是一个基于 Cloudwego Eino 框架的 CLI AI 助手，采用**手动编排**（Manual Orchestration）模式，自主控制推理轮次、工具调度、上下文拼装与终止判断。使用 Go 1.25 编写，默认接入 DeepSeek API。

---

## 2. 项目结构

```
Amadeus/
├── cmd/amadeus/main.go        # CLI 入口
├── internal/
│   ├── model/                 # 模型层
│   ├── orchestrator/          # 编排层（核心）
│   ├── tool/                  # 工具层
│   │   └── basetools/         # 内置工具（bash, load_skill）
│   ├── memory/                # 会话持久化
│   ├── session/               # 会话状态管理
│   ├── skill/                 # Skill 系统
│   └── presentation/          # CLI 表现层
├── skills/                    # Skill 注册表 + 文档
├── tools/                     # MCP 工具配置
├── checkpoints/sessions/      # 会话数据存储
└── docs/                      # 文档
```

---

## 3. 模块职责与交互

### 3.1 模型层（internal/model/）

- **`chat_model.go`**：创建 OpenAI-compatible ChatModel，拼接 system message（含 `agent.md` skill 注册表）
- **`config.go`**：从 `DEEPSEEK_*` 环境变量解析模型配置
- **`reasoning_payload.go`**：通过 Payload Modifier 注入 `reasoning_content` 字段，保留多轮推理内容

### 3.2 编排层（internal/orchestrator/）— 核心

| 文件 | 职责 |
|------|------|
| `orchestrator.go` | 结构体定义，`New()` 绑定工具，`HandleTurn()` 外部入口 |
| `loop.go` | 多轮对话循环：调用模型 → 执行工具 → 回填结果 → 判断终止 |
| `parser.go` | 工具参数最小校验（JSON 合法性） |
| `policy.go` | 从环境变量 `AMADEUS_MAX_TURNS` 读取最大轮次，默认 8 |

**核心循环逻辑（loop.go `run()`）：**

1. 调用 `streamModelTurn()` 获取模型输出（流式/非流式）
2. 若模型返回 `tool_calls` → 遍历执行工具 → 回填 `tool message`
3. 若 `load_skill` 成功 → 注入 skill 内容为 system message
4. 若模型返回无 `tool_calls` 且 content 非空 → 终止循环
5. 超 maxTurns 返回错误

### 3.3 工具层（internal/tool/）

- **`registry.go`**：聚合基础工具（bash + load_skill）和 MCP 工具
- **`executor.go`**：工具名 → InvokableTool 映射，执行后返回标准化 `Result` 结构
- **`mcp.go`**：解析 `tools/toolsConfig.json`，为每个 MCP 服务器创建 stdio 客户端，支持 `${VAR}` 环境变量占位符展开
- **`basetools/bash.go`**：内置 bash 工具，支持 workdir、timeout（默认 15s，最大 60s），输出截断保护（32KB）
- **`basetools/load_skill.go`**：按名称加载 `skills/{name}/SKILL.md` 内容

### 3.4 会话状态（internal/session/state.go）

- `State` 结构体：消息列表、当前轮次、工具调用计数、已加载 skill
- `NewState()`：从历史 + 已加载 skill 构建完整消息序列
- `ActivateSkill()`：去重注入 skill 文档为 system message

### 3.5 记忆层（internal/memory/store.go）

- 使用 **JSONL** 格式持久化到 `checkpoints/sessions/{session_id}/`
- 每次启动自动生成 session_id（`YYYYMMDD-HHMMSS-随机4字节hex`）
- 4 个文件：
  - `meta.json` — 会话元信息
  - `conversation.jsonl` — 对话记录（重建历史用）
  - `trace.jsonl` — 详细 trace（请求、响应、错误）
  - `loaded_skills.jsonl` — 已加载 skill 记录

### 3.6 Skill 系统（internal/skill/）

- `config.go`：通过 `SKILL_AGENT_MD_*` / `SKILL_ROOT_*` 环境变量配置路径
- `registry.go`：读取 `agent.md` 注册表并校验格式
- `loader.go`：读取 skill 文档，含路径逃逸安全检查

### 3.7 表现层（internal/presentation/）

- **事件驱动输出**：定义 7 种事件类型，`Renderer` 统一消费
- **双视图模式**：
  - `chat` 模式（默认）：简洁对话式，reasoning 灰色，answer 正常色
  - `trace` 模式（`AMADEUS_CLI_VIEW=trace`）：详细展示工具调用过程
- 支持终端宽度自适应（`COLUMNS` 环境变量），自动换行和前缀对齐

---

## 4. 数据流（一次完整对话轮次）

```
用户输入
   ↓
1. LoadConversation()    ← 从 JSONL 恢复历史
2. LoadLoadedSkills()    ← 恢复已加载 skill
3. NewState()            ← 构建消息序列（system + history + user）
4. AppendUserMessage()   ← 持久化用户消息
   ↓
run() 多轮循环（maxTurns）
   ↓
streamModelTurn()
   ├── 流式：逐 chunk 输出 reasoning/answer 事件
   └── 非流式：一次返回完整响应
   ↓
有 tool_calls？
   ├── 是 → 执行工具（bash / load_skill / MCP）
   │        → 回填 tool message
   │        → load_skill 成功则注入 skill system message
   │        → 继续下一轮
   └── 否 → Finished=true，返回最终回复
```

---

## 5. Skill 系统工作机制

```
agent.md（注册表，仅 name + desc）
       │ 模型根据 desc 判断是否需要
       ▼
load_skill("xxx") 工具调用
       │
       ▼
skills/xxx/SKILL.md 完整指令
       │
       ▼
注入为 system message → 模型后续推理携带该指令
```

---

## 6. 关键设计决策

| 决策 | 说明 |
|------|------|
| **手动编排** | 不依赖 Eino 内置 Agent，自行控制循环与终止 |
| **Skill 按需加载** | 避免 system prompt 膨胀，仅确认需要时加载 |
| **JSONL 持久化** | 追加写入，方便审计、调试和恢复 |
| **错误结构化** | 工具失败仍返回 Result，让模型理解失败原因 |
| **参数前置校验** | 工具参数进入执行层前先做 JSON 合法性检查 |
| **MCP 集成** | 通过 stdio MCP 协议接入外部工具（如高德地图） |

---

## 7. 配置一览

### 模型配置

| 变量 | 默认值 |
|------|--------|
| `DEEPSEEK_API_KEY` | — |
| `DEEPSEEK_MODEL` | `deepseek-v4-flash` |
| `DEEPSEEK_BASE_URL` | `https://api.deepseek.com` |
| `DEEPSEEK_THINKING_TYPE` | `enabled` |
| `DEEPSEEK_REASONING_EFFORT` | `high` |
| `DEEPSEEK_STREAM` | `true` |

### 系统配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `AMADEUS_MAX_TURNS` | `8` | 最大对话轮次 |
| `AMADEUS_CLI_VIEW` | `chat` | 视图模式 |
| `COLUMNS` | `100` | 终端宽度 |
| `SKILL_AGENT_MD_REL` | — | agent.md 相对路径 |
| `SKILL_ROOT_REL` | — | skill 根目录相对路径 |

---

## 8. 依赖关系图

```
cmd/amadeus/main.go
  ├── internal/model/     (ChatModel)
  ├── internal/orchestrator/  (循环控制)
  │   ├── internal/session/   (状态管理)
  │   ├── internal/memory/    (持久化)
  │   └── internal/presentation/ (渲染输出)
  ├── internal/tool/      (工具执行)
  │   ├── internal/tool/basetools/ (内置工具)
  │   └── MCP clients
  └── internal/skill/     (Skill 加载)
```

---

## 9. 可扩展点

1. **新增 Skill**：在 `agent.md` 注册 + 创建 `skills/{name}/SKILL.md`
2. **新增 MCP 工具**：在 `tools/toolsConfig.json` 添加 MCP 服务器配置
3. **新增内置工具**：在 `internal/tool/basetools/` 实现并注册到 `Load()`
4. **切换模型**：修改 `DEEPSEEK_*` 环境变量指向任意 OpenAI-compatible API
5. **更换存储**：实现 `memory.Store` 接口对接数据库
