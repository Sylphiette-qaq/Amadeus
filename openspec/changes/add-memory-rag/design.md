## Context

Amadeus 当前的记忆系统：每次启动生成新 Session ID，`memory.Store` 将 user/assistant 消息追加到 `conversation.jsonl`，`LoadConversation()` 只读取当前 session 文件。历史 session 的内容对模型完全不可见。

本 change 在不改变现有 JSONL 存储行为的前提下，新增一条写入路径（JSONL → Embedding → Milvus）和一条读取路径（模型工具调用 → Embedding → Milvus 检索）。

## Goals / Non-Goals

**Goals:**
- 每轮对话结束后将 user、assistant 消息分别 embedding 并写入 Milvus。
- 提供 `search_memory` 工具，让模型主动检索跨 session 的历史语义相关内容。
- Milvus / Embedding 服务不可用时软降级，不中断主对话流程。
- 索引粒度为单条消息（user 和 assistant 各一条），保证检索语义精度。

**Non-Goals:**
- 不替换现有 JSONL 存储，两者并存。
- 不实现自动（无工具调用）的隐式上下文注入。
- 不支持对 Milvus 中已有记录的更新或删除。
- 不实现批量历史迁移工具（历史 session 的 JSONL 不自动补索引）。

## Decisions

- **索引粒度：user/assistant 各一条。** 整轮合并为一条会导致长对话向量语义模糊，拆分后召回精度更高，且元数据可区分来源角色。

- **读取路径：工具调用（非自动注入）。** 与现有 `load_skill`、`bash`、`cmd` 工具的编排风格一致；让模型自己决定何时需要检索，避免每轮都附加无关历史片段增加 token 消耗。

- **Milvus 软降级。** Indexer 写入失败只 `log.Printf` 警告，`search_memory` 工具返回 `Result{Success: false, Error: "..."}` 而非 panic 或中断编排。符合项目现有工具执行的错误处理惯例。

- **Embedding 使用 OpenAI API（独立配置）。** 当前 DeepSeek 模型无 Embedding 端点，OpenAI Embedding 通过 `eino-ext/components/embedding/openai` 接入，与 ChatModel 使用不同的 API Key 和 Base URL，互不干扰。

- **Milvus 使用 `eino-ext/components/retriever/milvus` + `indexer/milvus`。** 官方支持，接口标准，无需引入社区包。

## Milvus Collection Schema

```
Collection: amadeus_memory（可通过 MILVUS_COLLECTION 配置）

字段：
  id          int64   主键，自增
  session_id  varchar 来源 session
  turn        int32   轮次编号
  role        varchar "user" | "assistant"
  text        varchar 原始消息文本（用于返回给模型展示）
  vector      float[] embedding 向量（维度由 OPENAI_EMBEDDING_MODEL 决定）
```

## Data Flow

```
写入路径：
  handleTurn 结束
    → memory.Indexer.IndexMessages(sessionID, turn, userMsg, assistantMsg)
        → OpenAI Embedding(userMsg.Content)   → Milvus.Store(user record)
        → OpenAI Embedding(assistantMsg.Content) → Milvus.Store(assistant record)
        （失败时 log.Printf，不返回 error）

读取路径：
  模型调用 search_memory(query="...", top_k=5)
    → basetools.SearchMemoryTool
        → OpenAI Embedding(query)
        → Milvus.Retrieve(vector, topK=top_k)
        → 格式化为可读文本片段列表
        → Result{Success: true, Data: "..."}
```

## Risks / Trade-offs

- **Embedding API 延迟** → 写入路径在 `handleTurn` 返回后异步执行（goroutine），不阻塞用户等待响应。
- **Milvus 首次启动需建 Collection** → Indexer 初始化时自动 CreateCollection（若不存在），失败则软降级。
- **向量维度依赖模型** → `text-embedding-3-small` 为 1536 维，切换模型需重建 collection，文档中注明。
