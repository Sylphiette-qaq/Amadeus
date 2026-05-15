# RAG 记忆系统

## 概述

Amadeus 内置基于向量检索的长期记忆能力，将每轮对话的 user/assistant 消息向量化后存入 Milvus，并通过 `search_memory` 工具在后续对话中按语义相似度检索历史内容。

未配置 `OPENAI_EMBEDDING_API_KEY` 时，RAG 自动降级为 noop 模式，`search_memory` 工具返回"记忆服务不可用"，其余功能不受影响。

---

## 架构

```
用户输入
  → orchestrator.HandleTurn
      → （对话结束后）go memory.Indexer.IndexMessages   # 异步写入，不阻塞响应
          → OpenAI Embedding API → 向量化
          → Milvus gRPC (19530)  → 写入 FloatVector Collection

模型调用 search_memory 工具
  → basetools.GetSearchMemoryTool
      → memory.Indexer.Search(query, topK)
          → OpenAI Embedding API → 查询向量化
          → Milvus COSINE 检索 → 返回 top-K 历史片段
```

---

## 配置项

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `OPENAI_EMBEDDING_API_KEY` | （空，必填） | 不设置则 RAG 禁用 |
| `OPENAI_EMBEDDING_BASE_URL` | `https://open.bigmodel.cn/api/paas/v4` | 兼容 OpenAI 格式的 Embedding 接口 |
| `OPENAI_EMBEDDING_MODEL` | `embedding-3` | 向量模型名称 |
| `OPENAI_EMBEDDING_DIMENSIONS` | `2048` | 向量维度，需与模型一致 |
| `MILVUS_ADDRESS` | `localhost:19530` | Milvus gRPC 地址；Docker 环境改为 `milvus:19530` |
| `MILVUS_COLLECTION` | `amadeus_memory` | Collection 名称，首次启动自动创建 |

---

## Milvus Collection Schema

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | VarChar(255) | 主键，格式 `<session_id>-t<turn>-<role>` |
| `vector` | FloatVector(dim) | COSINE 度量的浮点向量 |
| `content` | VarChar(8192) | 消息原文（超长截断） |
| `metadata` | JSON | `session_id`, `turn`, `role` |

索引：AUTOINDEX，度量类型 COSINE。

---

## search_memory 工具

模型在需要回忆历史内容时自主调用，参数如下：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `query` | string | 是 | 语义检索查询文本 |
| `top_k` | integer | 否 | 返回条数，范围 1–20，默认 5 |

**Session 隔离**：当 context 中携带 `session_id`（agent-server 模式下自动注入）时，检索自动限定在当前会话内；CLI 模式下跨所有 session 检索。

**返回格式**：
```
[Session: <id>, Turn 3, user]
用户说的内容...

[Session: <id>, Turn 3, assistant]
助手的回复...
```

---

## 降级行为

以下情况下 `Indexer` 以 noop 模式启动，不报错、不影响对话：

1. `OPENAI_EMBEDDING_API_KEY` 为空
2. OpenAI Embedding API 初始化失败
3. Milvus 连接超时（15 秒，见 `cmd/agent-server/main.go`）
4. Milvus Indexer/Retriever 初始化失败

noop 模式下 `IndexMessages` 静默跳过，`Search` 返回 `"记忆服务不可用"` 错误字符串。

---

## 已知边界情况

- **Collection 为空时检索报错**：Milvus 在 collection 无数据时返回 `"extra output fields found"` 错误，代码已识别并转换为"未找到相关历史记录"。
- **内容长度限制**：单条消息超过 8192 字符时截断入库，原始对话历史不受影响。
- **向量维度不匹配**：若更换 Embedding 模型需同步修改 `OPENAI_EMBEDDING_DIMENSIONS`，并删除旧 Collection（或新建 Collection 名）。

---

## 相关代码

| 路径 | 职责 |
|------|------|
| `internal/memory/indexer.go` | Indexer 核心：连接 Milvus、向量化、读写 |
| `internal/tool/basetools/search_memory.go` | search_memory 工具定义 |
| `cmd/agent-server/main.go` | 带 15s 超时的 NewIndexer 调用，降级处理 |
