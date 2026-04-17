## Why

The current Amadeus CLI interaction surface is still a thin text prompt plus raw print output, which makes multi-turn agent execution hard to follow once tool calls and intermediate states appear. This is more limiting now because the current loop still waits for a full model response before printing, and it does not surface the model's reasoning stream even though the message schema already supports `reasoning_content`.

This change is needed now because the project has already moved orchestration into application code, making it possible to control streaming, model selection, and presentation behavior in one place. Upgrading the panel without first exposing streaming output and reasoning state would leave the interaction feeling static and opaque.

## What Changes

- Redesign the CLI interaction panel around clear conversation sections for user input, agent response, tool activity, and execution status.
- Switch the default model from `deepseek-chat` to `deepseek-reasoner` so the CLI can consume both answer content and model reasoning content.
- Replace one-shot final answer rendering with streaming model output so the panel updates while the request is running.
- Introduce a structured presentation model so tool calls, tool results, errors, and final answers are rendered with distinct visual treatments instead of raw `fmt.Printf` output.
- Surface the model's reasoning text from `schema.Message.ReasoningContent` and render it as a first-class panel section instead of discarding it.
- Add turn-aware and progress-aware output patterns so users can understand what the agent is doing during a request, not only the final answer.
- Define a lightweight view mode strategy that supports a concise chat-oriented view by default while preserving access to detailed execution traces for debugging.

## Capabilities

### New Capabilities
- `cli-agent-interaction-panel`: Defines the requirements for presenting user turns, agent progress, tool execution, and final responses in a readable CLI panel.

### Modified Capabilities

## Impact

- Affected code will primarily include `cmd/amadeus/main.go`, `internal/model/chat_model.go`, `internal/presentation/*`, and the orchestration flow in `internal/orchestrator/*` where execution events are surfaced.
- The change may require introducing presentation-specific view models or event types to avoid leaking raw tool and model payloads directly to the terminal.
- The change will alter default model behavior by preferring `deepseek-reasoner`, and will require the orchestration loop to consume `Stream(...)` output and read `reasoning_content` from streamed or final model messages.
- No external API changes are expected, but terminal output behavior will change materially for local users and developers operating the CLI.
