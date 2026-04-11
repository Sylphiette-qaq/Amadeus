# 分阶段实施与代码改造点

## 12. 实施方案

### 12.1 Phase 1：最小可运行版本

目标是先替换执行主链路，不追求一次性把所有能力做全。

实施内容：

1. 保留现有 DeepSeek ChatModel 初始化逻辑
2. 新增 `GetOrchestrator` 或 `orchestrator.Run`
3. 停止创建 `adk.ChatModelAgent`
4. 仅支持单工具串行调用
5. 通过 OpenAI `tools` + `tool_calls` 协议驱动工具调用
6. 继续沿用现有上下文文件，但增加基础结构标记

### 12.2 Phase 2：结构化工具与上下文

实施内容：

1. 引入 `ToolRegistry` 与 `ToolExecutor`
2. 将 MCP 工具转换为统一注册结构
3. 将上下文持久化升级为 JSON Lines
4. 增加最大轮次、超时、重试与错误包装

### 12.3 Phase 3：可观测性与策略控制

实施内容：

1. 增加每轮 trace 日志
2. 支持工具调用审计
3. 支持上下文压缩
4. 支持不同业务场景下的策略切换

## 13. 推荐代码改造点

### 13.1 `main.go`

当前：

- 初始化 Agent
- 调用 `utils.StreamResponse`

改造后：

- 初始化纯 ChatModel
- 初始化 ToolRegistry / ToolExecutor
- 初始化 Orchestrator
- 调用 `orchestrator.HandleTurn(ctx, userQuestion)`

### 13.2 `agent/agent.go`

建议拆分为：

- `GetChatModel(ctx)` 保留
- 删除或废弃 `GetAgent(ctx)`

该目录后续可重命名为 `model/` 或 `llm/`，避免名称继续误导为 Agent 层。

### 13.3 `utils/stream.go`

当前文件直接依赖 `*adk.ChatModelAgent`，不适合保留为现状。建议重构为：

- `RenderFinalStream`
- `RenderAssistantMessage`
- `RenderToolEvent`

把「编排」和「展示」解耦。

### 13.4 `utils/memory.go`

建议保留接口思路，但升级为更结构化的存储方式，至少支持工具消息类型。

### 13.5 `tools/`

建议保留现有工具实现，并补充：

- 工具元信息导出
- 参数校验入口
- 统一执行包装
