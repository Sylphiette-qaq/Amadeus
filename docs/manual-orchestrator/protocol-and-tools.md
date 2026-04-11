# 模型输出协议与工具体系

## 7. 模型输出协议设计

### 7.1 推荐不要依赖自然语言猜测工具调用

如果完全依赖模型自然语言输出，例如「我要调用天气工具查询」，后续解析会非常脆弱。建议直接采用 OpenAI Function Calling / Tool Calling 规范进行交互，不再设计项目私有的一套工具调用协议。

### 7.2 推荐采用 OpenAI 工具调用协议

主编排程序应按 OpenAI 协议向模型传入 `tools` 定义，每个工具使用 JSON Schema 描述参数。模型若决定调用工具，会在 assistant 消息中返回 `tool_calls`。

工具定义示例：

```json
{
  "type": "function",
  "function": {
    "name": "weather",
    "description": "查询指定城市天气",
    "parameters": {
      "type": "object",
      "properties": {
        "city": {
          "type": "string",
          "description": "城市名称"
        }
      },
      "required": ["city"]
    }
  }
}
```

模型返回工具调用示例：

```json
{
  "role": "assistant",
  "content": "",
  "tool_calls": [
    {
      "id": "call_123",
      "type": "function",
      "function": {
        "name": "weather",
        "arguments": "{\"city\":\"上海\"}"
      }
    }
  ]
}
```

应用执行工具后，需按 OpenAI 协议把工具结果作为 `role: "tool"` 的消息回填给模型：

```json
{
  "role": "tool",
  "tool_call_id": "call_123",
  "content": "上海今天天气为多云，最高温度 25 摄氏度。"
}
```

如果本轮没有 `tool_calls`，则可认为模型已经给出了最终回答，`content` 即面向用户输出的结果。

### 7.3 协议字段建议

编排器应重点处理以下 OpenAI 协议字段：

- `tools`：请求时提供的工具列表
- `tool_choice`：工具调用策略，建议支持 `auto` / `required` / 指定函数
- `assistant.tool_calls[]`：模型发起的工具调用列表
- `assistant.tool_calls[].id`：工具调用唯一标识
- `assistant.tool_calls[].function.name`：工具名称
- `assistant.tool_calls[].function.arguments`：JSON 字符串格式的参数
- `tool.tool_call_id`：工具回填消息与调用 ID 的关联字段
- `content`：模型最终自然语言回答

### 7.4 解析失败兜底

OpenAI 协议下，主风险不再是「整段输出不是项目约定 JSON」，而是「`function.arguments` 不是合法 JSON」或「参数与 Schema 不匹配」。主编排程序需要进入如下兜底逻辑：

1. 解析 `tool_calls[].function.arguments`
2. 若 JSON 解析失败，则记录异常日志并向模型补充纠偏消息
3. 若字段缺失或类型不匹配，则拒绝执行并返回工具参数错误
4. 若连续多轮参数非法，则结束并返回错误

### 7.5 最终回答判定

建议采用以下判定规则：

1. assistant 消息存在 `tool_calls`，则进入工具执行分支
2. assistant 消息不存在 `tool_calls`，且 `content` 非空，则视为最终回答
3. assistant 消息同时无 `tool_calls` 且 `content` 为空，则视为异常响应

## 8. 工具体系改造设计

### 8.1 改造原则

现有工具本身可以保留，但需要把「工具发现」和「工具执行」从 Agent 初始化阶段拆出来。

### 8.2 工具注册表

建议新增 `ToolRegistry`，统一维护：

- 工具名称
- 工具描述
- 参数 Schema
- 执行函数
- 工具来源（本地/MCP）
- 是否启用

其中参数 Schema 建议直接对齐 OpenAI `function.parameters` 所需的 JSON Schema 结构，避免项目内部维护两套协议。

### 8.3 工具执行器

建议新增 `ToolExecutor`，负责：

- 根据工具名查找工具
- 参数校验
- 工具调用
- 超时控制
- 错误包装
- 返回统一结构

统一结果格式建议如下：

```go
type ToolExecutionResult struct {
    ToolName   string
    Success    bool
    Output     string
    Error      string
    DurationMs int64
}
```

### 8.4 MCP 工具接入策略

当前 `tools/mcpClient.go` 的客户端初始化逻辑可以复用，但不再将 MCP 工具直接交给 `adk.ToolsConfig`，而是：

1. 启动时初始化 MCP 客户端
2. 拉取工具元数据
3. 转换为项目内部统一工具描述
4. 在主编排过程中按需调用对应工具

### 8.5 工具结果回填规范

工具返回结果后，应按 OpenAI 协议回填 `role: "tool"` 消息，而不是项目私有文本标记。建议结构如下：

```json
{
  "role": "tool",
  "tool_call_id": "call_123",
  "content": "{\"success\":true,\"data\":\"...\",\"error\":\"\"}"
}
```

`content` 可以是普通文本，也可以是 JSON 字符串。若后续系统希望支持更稳定的结果再利用，建议统一使用 JSON 字符串。
