## ADDED Requirements

### Requirement: POST /chat 接口
Agent Server SHALL 在 `:9000` 监听并提供 `POST /chat` 接口，接受 `conversation_id` 和 `message`，返回 `reply` 文本。

#### Scenario: 正常对话请求
- **WHEN** 客户端 POST `{"conversation_id":"private:123","message":"你好"}` 到 `/chat`
- **THEN** 返回 `{"reply":"<assistant回复文本>"}` 且 HTTP 200

#### Scenario: 缺少必填字段
- **WHEN** 请求体缺少 `message` 或 `conversation_id`
- **THEN** 返回 HTTP 400

### Requirement: 按 conversation_id 隔离会话
Agent Server SHALL 为每个唯一 `conversation_id` 维护独立的 orchestrator 和 memory store，不同会话上下文互不干扰。

#### Scenario: 两个不同会话独立上下文
- **WHEN** `private:A` 和 `group:B` 各自发送消息
- **THEN** 两者分别维护各自的对话历史

### Requirement: 同一会话串行处理
同一 `conversation_id` 的并发请求 SHALL 串行排队执行，不得并发调用同一 orchestrator。

#### Scenario: 同一会话快速连发
- **WHEN** 同一 `conversation_id` 在第一条处理中收到第二条消息
- **THEN** 第二条等待第一条完成后再处理
