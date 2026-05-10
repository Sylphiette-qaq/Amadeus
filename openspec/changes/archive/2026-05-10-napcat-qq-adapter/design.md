## Context

Amadeus 当前只有 CLI 入口（`cmd/amadeus`），`orchestrator.HandleTurn` 将回复 emit 到 stdout。要接入 QQ，需要两件事：(1) 把 orchestrator 包装成可被 HTTP 调用的服务，(2) 一个独立的 QQ 适配器监听 NapCat 推来的 OneBot 11 事件。

NapCat 配置：Action API `http://127.0.0.1:3000`（Token: 环境变量），Webhook 推送到 `http://localhost:8080`。

## Goals / Non-Goals

**Goals:**
- `cmd/agent-server` 暴露 `POST /chat`，按 `conversation_id` 维护独立 orchestrator + store
- `cmd/qq` 接收 NapCat webhook，响应私聊和群聊 @bot 消息
- orchestrator 新增 `HandleTurnWithResponse()` 返回回复文本
- 两个进程完全解耦，仅通过 HTTP 通信

**Non-Goals:**
- 群聊非 @bot 消息不响应
- 不做鉴权（webhook token 验证）、限流、持久化会话恢复（重启后会话重置）
- 不支持图片、文件等富媒体消息（仅处理文本）

## Decisions

**D1：orchestrator 新增 `HandleTurnWithResponse`**
- `run()` 已返回 `*schema.Message`，`handleTurn` 丢弃了它。新方法直接把 `resp.Content` 透传给调用方。
- 不改现有 `HandleTurn`，保持 CLI 行为不变。

**D2：Agent Server session 管理用 `sync.Map` + per-conversation mutex**
- `sync.Map[conversationID → *convSession]`，`convSession` 内含 `sync.Mutex` 保证串行。
- 不引入任何外部依赖，内存存储，重启清零（MVP 可接受）。
- 每个 convSession 懒加载：首次请求时初始化 store + orchestrator。

**D3：QQ Adapter 不 import 任何 `internal/*` 包**
- 仅通过 HTTP 调用 Agent Server `/chat`，完全解耦。
- `cmd/qq` 自包含：OneBot 事件解析、NapCat 发消息客户端、env 配置。

**D4：@bot 检测方式**
- 群消息 `message` 字段包含 `[CQ:at,qq=<BOT_ID>]` 即视为 @bot。
- 发给 AI 的文本先剥离 CQ 码再传入。

## Risks / Trade-offs

- [重启会话丢失] 内存 session 重启清零 → MVP 接受，后续可接 memory.Store 持久化
- [并发串行阻塞] 同一会话排队处理，模型响应慢时后续消息等待 → 符合设计预期
- [无 webhook 鉴权] NapCat 推来的事件不验证 token → 仅本地部署，风险可接受
