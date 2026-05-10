## 1. Orchestrator 扩展

- [x] 1.1 在 `internal/orchestrator/orchestrator.go` 新增 `HandleTurnWithResponse(ctx, question) (string, error)`，复用 `handleTurn` 逻辑并返回 assistant 最终回复文本

## 2. Agent Server

- [x] 2.1 创建 `cmd/agent-server/main.go`，初始化共享的 skill/model/tool 配置
- [x] 2.2 实现 `convSession` 结构体（含 `sync.Mutex`、`orchestrator`、`store`）和 `SessionManager`（`sync.Map`），支持懒加载
- [x] 2.3 实现 `POST /chat` handler：解析请求、获取或创建 session、调用 `HandleTurnWithResponse`、返回 `{"reply":"..."}`
- [x] 2.4 新增环境变量 `AGENT_SERVER_ADDR`（默认 `:9000`），server 监听该地址

## 3. QQ Adapter

- [x] 3.1 创建 `cmd/qq/main.go`，从环境变量读取 `NAPCAT_API_URL`、`NAPCAT_TOKEN`、`QQ_BOT_ID`、`AGENT_SERVER_URL`
- [x] 3.2 实现 OneBot 11 事件结构体解析（`post_type`、`message_type`、`user_id`、`group_id`、`message`）
- [x] 3.3 实现 webhook handler：私聊直接处理，群聊检测 `[CQ:at,qq=<BOT_ID>]`，其他忽略
- [x] 3.4 实现剥离 CQ 码函数，将群消息中的 CQ 码去掉后再传给 Agent API
- [x] 3.5 实现调用 `POST /chat` 的 HTTP 客户端函数
- [x] 3.6 实现调用 NapCat `send_private_msg` / `send_group_msg` 的函数（带 Bearer token header）
- [x] 3.7 HTTP server 监听 `:8080`

## 4. 验证

- [x] 4.1 启动 agent-server，curl 测试 `POST /chat` 能返回回复
- [x] 4.2 启动 qq adapter，发私聊消息验证能收到 AI 回复
- [x] 4.3 在群里 @bot 验证能收到回复，不 @不回复
