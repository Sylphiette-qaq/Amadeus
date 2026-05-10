## ADDED Requirements

### Requirement: 接收 NapCat webhook 事件
QQ Adapter SHALL 在 `:8080` 监听 HTTP POST，解析 OneBot 11 格式的事件 JSON。

#### Scenario: 收到私聊消息事件
- **WHEN** NapCat POST `{"post_type":"message","message_type":"private","user_id":123,"message":"你好"}`
- **THEN** Adapter 解析成功并触发处理流程

#### Scenario: 收到非消息事件
- **WHEN** `post_type` 不为 `message`
- **THEN** 直接返回 HTTP 200，不做任何处理

### Requirement: 响应私聊消息
QQ Adapter SHALL 对所有私聊消息（`message_type=private`）调用 Agent API 并将回复发回给发送者。

#### Scenario: 私聊消息正常处理
- **WHEN** 收到 `message_type=private`，`user_id=123`，`message="你好"`
- **THEN** 以 `conversation_id=private:123` 调用 `/chat`，再调用 NapCat `send_private_msg` 发回回复

### Requirement: 响应群聊 @bot 消息
QQ Adapter SHALL 仅对包含 `[CQ:at,qq=<BOT_ID>]` 的群消息响应，其余群消息忽略。

#### Scenario: 群消息含 @bot
- **WHEN** 收到 `message_type=group`，message 含 `[CQ:at,qq=BOT_ID]`
- **THEN** 剥离 CQ 码后以 `conversation_id=group:<group_id>` 调用 `/chat`，再 `send_group_msg` 回复

#### Scenario: 群消息不含 @bot
- **WHEN** 收到 `message_type=group`，message 不含 `[CQ:at,qq=BOT_ID]`
- **THEN** 忽略，返回 HTTP 200 不回复
