# 改造目标、架构与主编排程序

## 4. 改造目标

本次改造后的目标架构如下：

1. 保留 Eino/DeepSeek 作为底层 ChatModel 能力提供方
2. 不再使用 `adk.NewChatModelAgent`
3. 新增项目自有主编排程序，显式控制每一轮模型调用
4. 工具注册、工具发现、工具执行由项目代码统一管理
5. 每轮模型输出都进入业务侧处理逻辑，再决定下一步
6. 支持流式展示最终文本，同时保留结构化执行日志

需要强调的是，本次方案是「去 Agent 化」，不是「去工具化」。工具仍可继续使用，但由项目手动调度。

## 5. 目标架构设计

### 5.1 总体架构

建议将系统拆分为以下层次：

1. **`model` 层**  
   负责初始化 ChatModel，仅提供 `Generate` 或 `Stream` 能力，不做决策。

2. **`orchestrator` 层**  
   负责一次完整用户请求的执行编排，包括消息构建、模型调用、工具识别、工具执行、结果回填、终止判断。

3. **`tool registry` 层**  
   负责统一管理本地工具与 MCP 工具的元数据、调用入口和参数规范。

4. **`memory` 层**  
   负责会话历史、工具结果、中间状态、摘要信息的持久化与读取。

5. **`presentation` 层**  
   负责终端打印、流式展示、调试日志和错误输出。

### 5.2 推荐目录调整

建议目录演进为：

```text
Amadeus/
├── cmd/
│   └── amadeus/
│       └── main.go
├── internal/
│   ├── model/
│   │   └── chat_model.go
│   ├── orchestrator/
│   │   ├── orchestrator.go
│   │   ├── loop.go
│   │   ├── parser.go
│   │   └── policy.go
│   ├── tool/
│   │   ├── registry.go
│   │   ├── executor.go
│   │   ├── mcp.go
│   │   └── calculator.go
│   ├── memory/
│   │   ├── store.go
│   │   └── serializer.go
│   └── session/
│       └── state.go
├── docs/
├── checkpoints/
└── tools/
```

如果当前阶段不希望大规模调整目录，也可以先在现有结构上最小改造，待稳定后再迁移到 `internal/`。

## 6. 主编排程序设计

### 6.1 核心职责

主编排程序应显式实现以下逻辑：

1. 读取历史上下文
2. 组装系统消息、历史消息、当前用户消息
3. 调用模型获取本轮输出
4. 解析输出内容，判断是否存在工具调用意图
5. 若需要调用工具，则执行工具并生成工具结果消息
6. 将工具结果追加到上下文，再进入下一轮模型调用
7. 若模型产出最终回答，则结束本次请求
8. 保存完整会话记录和结构化日志

### 6.2 推荐执行循环

建议采用如下伪代码：

```go
for turn := 1; turn <= maxTurns; turn++ {
    messages := buildMessages(sessionState)

    resp, err := chatModel.Generate(ctx, messages)
    if err != nil {
        return fail(err)
    }

    sessionState.Append(resp)

    if len(resp.ToolCalls) == 0 {
        output(resp)
        persist(sessionState)
        return nil
    }

    for _, tc := range resp.ToolCalls {
        args, err := parser.ParseToolArguments(tc.Function.Arguments)
        if err != nil {
            sessionState.Append(toolErrorMessage(tc.ID, tc.Function.Name, err))
            continue
        }

        toolResult, err := toolExecutor.Execute(ctx, tc.Function.Name, args)
        sessionState.Append(toolResultMessage(tc.ID, tc.Function.Name, toolResult, err))
    }
}

return fail(maxTurnsExceeded)
```

### 6.3 轮次状态对象

建议引入 `SessionState` 或 `ConversationState`，统一承载：

- 会话 ID
- 历史消息
- 当前轮次编号
- 工具调用次数
- 最近一次工具调用结果
- 是否已结束
- 调试事件日志

这样可以避免编排逻辑散落在多个工具函数中。
