## Why

Amadeus 目前只能通过 CLI 与用户交互。接入 QQ 可以让 agent 服务覆盖日常使用场景，通过 NapCat（OneBot 11 协议）将 Amadeus 的对话能力暴露给 QQ 私聊和群聊。

## What Changes

- 新增 `cmd/agent-server`：将 orchestrator 包装为 HTTP Agent API，暴露 `POST /chat` 接口，按 `conversation_id` 维护独立会话
- 新增 `cmd/qq`：QQ 适配器，监听 NapCat webhook 事件（OneBot 11），响应私聊消息和群聊 @bot 消息，调用 Agent API 后将回复发回 QQ
- `internal/orchestrator` 新增 `HandleTurnWithResponse()` 方法，返回 assistant 最终回复文本
- 新增环境变量配置：`NAPCAT_API_URL`、`NAPCAT_TOKEN`、`QQ_BOT_ID`、`AGENT_SERVER_ADDR`、`AGENT_SERVER_URL`

## Capabilities

### New Capabilities

- `agent-http-api`：Amadeus orchestrator 的 HTTP 封装，提供 `POST /chat` 接口，按 conversation_id 管理独立会话
- `qq-adapter`：NapCat webhook 接收与 QQ 消息分发，支持私聊全响应和群聊 @mention 响应

### Modified Capabilities

（无）

## Impact

- `internal/orchestrator`：新增一个公开方法，不改变现有行为
- `cmd/amadeus`：不受影响
- 新增两个独立二进制入口，均通过 `.env` 配置，互相解耦
