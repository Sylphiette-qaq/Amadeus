## ADDED Requirements

### Requirement: 对话历史实时向量索引
每轮对话成功结束后，Amadeus SHALL 将本轮 user 消息和 assistant 消息分别 embedding 后写入 Milvus，每条消息存为独立记录（含 session_id、turn、role、text、vector 字段）。

#### Scenario: 正常轮次索引
- **WHEN** 一轮对话成功结束（assistant 最终消息已持久化）
- **THEN** user 消息和 assistant 消息各自 embedding 并写入 Milvus，写入在后台异步执行不阻塞响应

#### Scenario: Milvus 写入失败
- **WHEN** Milvus 连接不可用或写入返回错误
- **THEN** 以 `log.Printf` 记录警告，主对话流程不受影响，不向用户报错

#### Scenario: Embedding API 不可用
- **WHEN** `OPENAI_EMBEDDING_API_KEY` 未配置或 API 返回错误
- **THEN** 索引写入静默跳过，主对话流程继续正常运行

### Requirement: search_memory 工具
Amadeus SHALL 提供 `search_memory` 工具，使模型可主动检索跨 session 的语义相关历史对话片段。

#### Scenario: 检索到相关记录
- **WHEN** 模型调用 `search_memory(query="上次关于 pgvector 的讨论", top_k=5)`
- **THEN** 返回最多 5 条按语义相似度排序的历史消息片段，每条包含 session_id、turn、role 和原始文本

#### Scenario: 未找到相关记录
- **WHEN** query 与所有历史记录相似度均低于阈值，或 Milvus 中暂无记录
- **THEN** 返回 `Result{Success: true, Data: "未找到相关历史记录"}`

#### Scenario: 记忆服务不可用（降级）
- **WHEN** Milvus 未启动或连接失败
- **THEN** 返回 `Result{Success: false, Error: "记忆服务不可用"}`，编排器继续运行不中断

#### Scenario: top_k 超出范围
- **WHEN** 模型传入 `top_k=0` 或 `top_k=25`（超出 1–20 范围）
- **THEN** 自动 clamp 到合法范围（最小 1，最大 20）后继续执行

### Requirement: 软降级启动
当 RAG 相关环境变量未配置时，Amadeus SHALL 正常启动并保持完整对话能力，仅 RAG 功能不可用。

#### Scenario: 环境变量缺失时启动
- **WHEN** `OPENAI_EMBEDDING_API_KEY` 或 `MILVUS_ADDRESS` 未在 `.env` 中配置
- **THEN** Amadeus 正常启动，打印一条警告提示 RAG 功能已禁用，`search_memory` 工具注册但始终返回服务不可用
