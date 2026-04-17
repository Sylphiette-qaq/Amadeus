## 1. Presentation Event Model

- [x] 1.1 Change `internal/model/chat_model.go` to default to `deepseek-reasoner` while preserving environment override support
- [x] 1.2 Define presentation event types for turn lifecycle, tool lifecycle, streamed answer updates, reasoning updates, final answers, and turn-level failures under `internal/presentation`
- [x] 1.3 Refactor `internal/orchestrator` to consume model streaming output and emit structured presentation events instead of directly printing raw tool and assistant output
- [x] 1.4 Remove or isolate legacy direct print paths so the new renderer becomes the primary output flow

## 2. Default CLI Interaction Panel

- [x] 2.1 Implement streaming accumulation for assistant `Content` and `ReasoningContent` so each turn preserves both channels through completion
- [x] 2.2 Implement a default renderer that groups each request into a structured turn block with user, progress, reasoning, tool, and answer sections
- [x] 2.3 Add concise tool status and tool outcome summaries for the default view without printing raw payloads inline
- [x] 2.4 Update input and turn boundary presentation so multiple turns remain visually distinct during one CLI session

## 3. Trace View And Error States

- [x] 3.1 Add a trace-oriented rendering mode that includes detailed tool arguments, detailed tool result content, and detailed reasoning content
- [x] 3.2 Render tool failures and turn-level failures as explicit error states distinct from normal assistant output
- [x] 3.3 Verify the new streaming presentation flow against the interaction panel spec, including `reasoning_content` handling and default model selection
