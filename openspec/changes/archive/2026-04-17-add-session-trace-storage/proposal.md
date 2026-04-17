## Why

The current CLI persistence layer only appends flattened text lines to `checkpoints/context.txt`, which loses tool messages, intermediate turns, streaming chunks, and the full message arrays that the orchestrator sends to the model. That makes it impossible to audit a session faithfully or inspect the exact raw context that produced a model decision.

This change is needed now because the project already has a manual orchestrator with explicit turn boundaries, tool execution, and streamed model output. Those seams are the right place to introduce durable, raw session records without relying on rendered terminal output or lossy text reconstruction.

## What Changes

- Introduce session-scoped storage with a generated `session_id` and per-session files under `checkpoints/sessions/<session_id>/`.
- Split persistence into two lanes:
- `conversation.jsonl` stores only user messages and final assistant answers for future context restoration.
- `trace.jsonl` stores raw orchestration events, including full turn request message arrays, streamed model chunks, tool calls, tool results, tool messages, and turn errors.
- Persist the exact `state.Messages` snapshot before each model turn so every model request can be inspected later with its full original context.
- Persist every streamed model chunk before concatenation, then persist the concatenated assistant message for the turn.
- Persist raw tool call, tool execution result, and generated tool message records without summarization or display-oriented formatting.
- Update context loading to rebuild history only from the conversation lane, restoring user messages and final assistant answers while excluding tool and trace data.

## Capabilities

### New Capabilities
- `session-trace-storage`: Defines session-scoped persistence for raw orchestration traces and a separate conversation history used for context restoration.

### Modified Capabilities

## Impact

- Affected code will include `cmd/amadeus/main.go`, `internal/orchestrator/*`, `internal/session/*`, and `internal/memory/*`.
- The current text-based `checkpoints/context.txt` format will be superseded by structured session storage for new runs.
- The change introduces a persistent `session_id` concept and additional JSON Lines files that must be written during each turn.
- No model or tool protocol changes are required, but the orchestrator will gain trace-writing responsibilities at model and tool lifecycle boundaries.
