## Context

Amadeus has already moved request orchestration into `internal/orchestrator`, which means the application now knows when a turn starts, when a tool call is emitted, when a tool result returns, and when a final assistant answer is available. The current CLI presentation layer does not model any of that structure. It still prints a prompt, raw tool call payloads, raw tool results, and a final assistant response with minimal formatting.

This creates three practical problems. First, users cannot easily distinguish normal conversation from execution trace data. Second, debugging tool-heavy turns requires scanning unstructured JSON-like output. Third, the presentation package is too thin to evolve into richer modes such as concise user-facing output versus developer-facing trace output.

There is also a model-level mismatch. `internal/model/chat_model.go` still defaults to `deepseek-chat`, and the current orchestrator path uses `Generate(...)`, which means the CLI only sees a completed answer after the model finishes. However, the current DeepSeek integration already exposes `Stream(...)`, and `schema.Message` already includes `ReasoningContent`, so the existing architecture can support streaming answer text plus streaming or final reasoning text without changing the product boundary.

The design needs to stay within the current CLI environment and should not require introducing a full TUI framework. It also needs to preserve the existing orchestrator ownership of control flow rather than moving state back into ad hoc print helpers.

## Goals / Non-Goals

**Goals:**
- Create a presentation model that renders a single request as a structured CLI interaction panel instead of a sequence of unrelated print statements.
- Separate user turn content, agent status, tool execution, errors, and final answers into distinct display sections.
- Stream model output into the panel as it is generated instead of waiting for a full final answer.
- Switch the default model to `deepseek-reasoner` and surface its `reasoning_content` in the interaction flow.
- Support a concise default view while keeping an explicit path to richer execution traces for debugging.
- Keep the design compatible with the current orchestrator loop and local terminal usage.

**Non-Goals:**
- Building a full-screen TUI with cursor movement, panes, or mouse interaction.
- Changing the core model, tool execution, or memory persistence behavior beyond what is needed to surface presentation events.
- Introducing remote telemetry, web dashboards, or external logging systems as part of this change.

## Decisions

### 1. Switch the default model to `deepseek-reasoner`

The model factory should default to `deepseek-reasoner` rather than `deepseek-chat`, while still allowing explicit override through environment configuration.

Rationale:
- The requested panel behavior depends on the model exposing reasoning content in addition to final answer text.
- Defaulting the application to the reasoning-capable model keeps the interaction contract aligned with the new panel requirements.
- Keeping environment override preserves operational flexibility.

Alternatives considered:
- Keep `deepseek-chat` as default and only support reasoning optionally. Rejected because it would make the new panel behavior inconsistent by default.
- Hardcode `deepseek-reasoner` without override. Rejected because environment-based model selection is already part of the current setup and remains useful.

### 2. Introduce presentation events instead of printing directly from orchestration

The presentation layer should receive typed events such as `TurnStarted`, `ToolCallStarted`, `ToolCallFinished`, `AssistantResponseReady`, and `TurnFailed` rather than raw data blobs and immediate print instructions.

Rationale:
- This keeps orchestration in charge of execution while letting presentation decide how to render each state.
- It avoids scattering display formatting across `main.go`, `orchestrator`, and utility code.
- It gives the project a stable seam for future streaming and trace output improvements.

Alternatives considered:
- Continue expanding `fmt.Printf` helpers in place. Rejected because formatting logic will stay tightly coupled to execution order and become harder to test.
- Move rendering entirely into `main.go`. Rejected because it would make the CLI entrypoint absorb presentation-specific complexity.

### 3. Use streaming model consumption as the primary orchestration path

The orchestrator should consume `ChatModel.Stream(...)` for assistant turns so the CLI can render answer tokens and reasoning state while the model is still generating. The orchestrator should accumulate the streamed result into a final assistant message that preserves both `Content` and `ReasoningContent` for downstream tool handling and persistence.

Rationale:
- Streaming is required to make the panel feel active rather than stalled during long generations.
- It aligns the actual execution behavior with the interaction panel's progress-oriented design.
- The model integration already exposes `Stream(...)`, so this is a natural extension of the current architecture.

Alternatives considered:
- Keep `Generate(...)` and fake progress with static placeholders. Rejected because it would still hide real model latency and would not expose incremental reasoning updates.
- Stream only the final answer text and ignore reasoning content. Rejected because the user explicitly wants the model thinking process surfaced.

### 4. Read and render `ReasoningContent` as a first-class channel

The presentation and orchestration layers should treat `schema.Message.ReasoningContent` as a first-class output channel that is separate from the assistant's answer text. In the default panel this reasoning should appear as a dedicated "thinking" or "reasoning" section, and in trace mode it can be rendered in fuller detail.

Rationale:
- `Content` and `ReasoningContent` have different semantic roles and should not be merged into one blob.
- Preserving the distinction supports cleaner UI hierarchy and clearer debugging.

Alternatives considered:
- Append reasoning text directly into the final answer body. Rejected because it conflates private thought process with user-facing answer text.
- Discard reasoning content after streaming. Rejected because the feature objective is to expose and inspect it.

### 5. Use a sectioned single-turn layout in plain terminal output

Each user request should render as a visually grouped block with sections for user input, agent status/progress, tool timeline, and final answer. The output should remain line-oriented and shell-friendly, without relying on terminal cursor control.

Rationale:
- This improves readability immediately with low implementation risk.
- It works in local terminals, logs, and redirected output without special capabilities.
- It avoids the complexity and fragility of adopting a full TUI framework at this stage.

Alternatives considered:
- Adopt Bubble Tea or another TUI library now. Rejected because the current scope is about structured output, not interactive terminal app behavior.
- Keep a flat chronological log only. Rejected because users need visual grouping, not just timestamps.

### 6. Default to chat view and make trace view explicit

The default rendering should emphasize user-visible conversation and concise progress summaries. Detailed tool arguments, full tool results, and lower-level execution detail should be reserved for an explicit trace/debug mode.

Rationale:
- Most runs should optimize for readability rather than maximum raw detail.
- Developers still need detailed traces when validating tool routing and failures.
- This creates a stable product distinction between “use the assistant” and “debug the assistant.”

Alternatives considered:
- Always show full raw details. Rejected because the current output already proves this becomes noisy quickly.
- Hide tool activity entirely. Rejected because tool execution is central to trust and debuggability.

### 7. Summarize tool activity by status and key outcome

Tool rendering should prefer summaries such as tool name, success/failure, and a short result synopsis over unbounded raw payload dumps. Full payload detail can remain available in trace mode.

Rationale:
- Raw arguments and JSON results are often the least readable form of information.
- A summary-first strategy keeps the panel compact while still communicating progress and outcomes.

Alternatives considered:
- Truncate raw payloads only. Rejected because truncation without semantic summarization still produces noisy output.

## Risks / Trade-offs

- [Risk] The event model may be too narrow for future streaming or partial output states. → Mitigation: define event types around lifecycle boundaries and allow event payloads to evolve without changing the public rendering contract.
- [Risk] Streamed reasoning content may arrive in chunks that are incomplete or visually noisy. → Mitigation: buffer by section and render reasoning updates with stable prefixes or grouped blocks instead of printing every fragment raw.
- [Risk] `deepseek-reasoner` may change response latency, token usage, or tool-calling behavior relative to `deepseek-chat`. → Mitigation: keep environment override support and validate the orchestrator flow against the new default model before removing fallbacks.
- [Risk] Maintaining both chat view and trace view could introduce duplicated rendering logic. → Mitigation: build both views on top of the same event stream and shared formatting helpers.
- [Risk] Summarizing tool results may hide details needed during debugging. → Mitigation: make trace mode explicit and preserve access to raw detail paths there.
- [Risk] Plain terminal formatting may still look limited compared with a true TUI. → Mitigation: optimize structure and hierarchy first; defer cursor-driven UI until the limitations become concrete.

## Migration Plan

1. Change the model default to `deepseek-reasoner` while preserving environment override behavior.
2. Add a presentation event model and adjust orchestrator output calls to emit structured lifecycle events from a streaming model path.
3. Implement streaming accumulation for assistant `Content` and `ReasoningContent`.
4. Implement a default CLI renderer for the new event stream with sectioned single-turn output, including a dedicated reasoning section.
5. Add trace-mode rendering for detailed tool arguments, tool results, and reasoning detail.
6. Remove or deprecate older direct print helpers once the new panel output fully covers current usage.

Rollback is straightforward because the change is presentation-local. The previous raw print helpers can remain available behind a fallback path until the new renderer is validated.

## Open Questions

- Should trace mode be selected by environment variable, CLI flag, or both?
- Do we want per-tool custom summarizers now, or should the first iteration use generic summary rules?
- Should reasoning content be persisted in the same history store as assistant answers, or treated as ephemeral debug output in the first iteration?
