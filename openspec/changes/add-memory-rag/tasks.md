## 1. 基础设施：Embedding + Milvus 初始化

- [x] 1.1 引入依赖：`eino-ext/components/embedding/openai`、`eino-ext/components/retriever/milvus`、`eino-ext/components/indexer/milvus`，运行 `go mod tidy`。
- [x] 1.2 新增 `internal/memory/indexer.go`：定义 `Indexer` 结构体，封装 OpenAI Embedding client 和 Milvus Indexer/Retriever 的初始化（含 Collection 自动建表），提供 `IndexMessages(sessionID, turn, userMsg, assistantMsg)` 和 `Search(query, topK)` 方法。
- [x] 1.3 Milvus 或 Embedding 初始化失败时返回 no-op `Indexer`，调用其方法时静默跳过并打印警告。

## 2. 写入路径集成

- [x] 2.1 在 `Orchestrator` 结构体中新增 `indexer *memory.Indexer` 字段。
- [x] 2.2 在 `handleTurn` 的 `AppendAssistantFinal` 之后，启动 goroutine 调用 `indexer.IndexMessages`，传入当前 sessionID、turn、user 消息、assistant 最终消息。

## 3. search_memory 工具

- [x] 3.1 新增 `internal/tool/basetools/search_memory.go`：实现 `search_memory` 工具，参数 `query`（必填）和 `top_k`（可选，默认 5，上限 20），调用 `Indexer.Search` 并将结果格式化为可读文本返回。
- [x] 3.2 在 `basetools.Load()` 中注册 `search_memory` 工具（仅当 Indexer 非 no-op 时注册，或始终注册但 no-op 时返回服务不可用）。

## 4. 配置与初始化

- [x] 4.1 新增 `internal/memory/indexer_config.go`（或在 indexer.go 内）：从环境变量读取 `OPENAI_EMBEDDING_API_KEY`、`OPENAI_EMBEDDING_BASE_URL`、`OPENAI_EMBEDDING_MODEL`、`MILVUS_ADDRESS`、`MILVUS_COLLECTION`。
- [x] 4.2 在 `cmd/amadeus/main.go` 中初始化 `memory.Indexer`，注入到 Orchestrator 和 Tool 注册。
- [x] 4.3 更新 `.env.example` 添加新环境变量及说明注释。
- [x] 4.4 更新 `AGENTS.md`：环境变量配置章节、架构包结构图。

## 5. 验证

- [x] 5.1 为 `memory.Indexer` 编写单元测试：mock Embedding 和 Milvus，验证 IndexMessages 正确拆分 user/assistant 并分别调用 Store，以及 Search 正确格式化返回结果。
- [x] 5.2 为 `search_memory` 工具编写单元测试：验证参数校验（top_k 范围）、no-op 降级行为。
- [x] 5.3 运行完整测试套件 `go test ./...`，确保无回归。
