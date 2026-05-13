## Why

Amadeus 的对话历史目前以 JSONL 文件按 session 存储，每次启动只能读取当前 session 的记录，跨 session 的内容对模型完全不可见。用户无法让 Amadeus 回忆起之前 session 中讨论过的内容。

通过引入向量语义检索，Amadeus 可以将所有历史对话索引到 Milvus，并通过 `search_memory` 工具让模型在需要时主动检索相关历史片段，实现跨 session 的长期记忆能力。

## What Changes

- 新增 `internal/memory/indexer.go`：封装 OpenAI Embedding + Milvus 写入，每轮对话结束后将 user/assistant 消息各自索引为独立向量记录。
- 新增 `internal/tool/basetools/search_memory.go`：`search_memory` 工具，模型主动调用时对查询文本 embedding 后从 Milvus 检索 Top-K 相关历史片段并返回。
- 修改 `internal/orchestrator/loop.go`：在 `handleTurn` 中对话持久化后异步触发索引写入。
- 修改 `internal/tool/basetools/load.go`：注册 `search_memory` 工具。
- 修改 `cmd/amadeus/main.go`：初始化 Indexer 并注入到 Orchestrator 和 Tool 层。
- 新增环境变量：`OPENAI_EMBEDDING_API_KEY`、`OPENAI_EMBEDDING_BASE_URL`、`OPENAI_EMBEDDING_MODEL`、`MILVUS_ADDRESS`、`MILVUS_COLLECTION`。
- Milvus 不可用时软降级：Indexer 写入失败只打印警告，`search_memory` 返回 `Success: false`，不中断主流程。

## Capabilities

### New Capabilities
- `memory-rag`：基于 Milvus 向量数据库的跨 session 对话历史语义检索能力。

### Modified Capabilities

## Impact

- 受影响代码：`internal/memory/`、`internal/tool/basetools/`、`internal/orchestrator/`、`cmd/amadeus/`。
- 新增外部依赖：`github.com/cloudwego/eino-ext/components/embedding/openai`、`github.com/cloudwego/eino-ext/components/retriever/milvus`、`github.com/cloudwego/eino-ext/components/indexer/milvus`。
- Milvus 服务为可选依赖，未配置或不可用时功能静默降级，不影响现有对话能力。
