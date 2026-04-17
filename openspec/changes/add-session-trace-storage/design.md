## Context

Amadeus currently persists only flattened text lines in `checkpoints/context.txt`. That file stores a timestamp, a role, and a content string, which is enough to replay a coarse user and assistant conversation but not enough to reconstruct the raw orchestration flow. The orchestrator already has richer state in memory: it rebuilds a `session.State` with full `schema.Message` arrays, streams model chunks, detects intermediate assistant messages with `tool_calls`, executes tools, and creates `tool` messages for the next turn. None of those objects are durably preserved today.

The requested change adds two distinct needs that must not be conflated. First, the application still needs a compact history to restore future turns. Second, operators need a lossless trace of what the orchestrator actually sent to and received from the model, including raw tool-calling structures in the model messages. The current text format cannot satisfy both purposes without either dropping fidelity or polluting restored context with internal trace data.

The design should stay local to the CLI process, avoid external dependencies, and reuse the current orchestrator ownership of control flow. It should also keep the trace records close to the business objects already in use, rather than serializing presentation output or ad hoc debug strings.

## Goals / Non-Goals

**Goals:**
- Introduce a durable `session_id` so each CLI run writes to its own session directory.
- Separate restorable conversation history from raw orchestration trace storage.
- Persist the exact `state.Messages` snapshot before every model turn.
- Persist the exact `schema.Message` array sent to the model for each turn and the final raw `schema.Message` returned by the model for that turn.
- Preserve tool-calling structures such as `tool_calls` inside the raw model message instead of splitting them into presentation-oriented records.
- Restore future model context only from user messages and final assistant answers, excluding tool and trace records.

**Non-Goals:**
- Capturing raw HTTP request and response payloads below the model SDK boundary.
- Reusing the trace log itself as the source of truth for context restoration.
- Changing model prompting, tool protocols, or terminal rendering behavior beyond the minimum needed to attach storage hooks.
- Introducing remote logging, log shipping, or external observability systems.

## Decisions

### 1. Use session-scoped directories with separate conversation and trace files

Each CLI run should generate a `session_id` and write under `checkpoints/sessions/<session_id>/`. The session directory should contain:

- `meta.json` for stable session metadata
- `conversation.jsonl` for restorable conversation history
- `trace.jsonl` for raw orchestration records

Rationale:
- This avoids mixing separate runs into a single append-only file.
- It makes session replay and manual inspection straightforward.
- It creates a clean boundary between business conversation history and internal execution trace.

Alternatives considered:
- Keep a single global JSONL file for all sessions. Rejected because session boundaries become implicit and recovery gets harder.
- Extend the existing `context.txt` format with more fields. Rejected because text parsing would become brittle and still encourage lossy flattening.

### 2. Restore context only from the conversation lane

`conversation.jsonl` should include only two record types:

- `user_message`
- `assistant_final`

Only those records should be loaded when rebuilding history for the next user turn. Intermediate assistant messages with `tool_calls`, `tool` messages, chunks, and errors should remain trace-only.

Rationale:
- The next model turn should see the user-visible conversation, not the entire execution trace.
- Intermediate assistant tool-routing messages are part of orchestration mechanics, not durable dialogue history.
- This preserves the simplicity of context restoration while allowing the trace lane to remain lossless.

Alternatives considered:
- Restore from all assistant messages. Rejected because assistant messages that only contain `tool_calls` would pollute future context.
- Restore from trace records. Rejected because trace includes internal execution details that are not appropriate for normal conversational memory.

### 3. Persist raw business objects as JSON instead of flattening strings

The storage layer should serialize the existing Go objects directly into JSON-friendly envelopes:

- `schema.Message` for `message`, `chunk`, and `messages`
- `schema.ToolCall` for tool call records
- `tool.Result` for tool results

Each record should also carry `session_id`, `turn`, `timestamp`, and a `type` discriminator.

Rationale:
- This keeps the persisted data aligned with actual orchestrator semantics.
- It avoids re-parsing display strings or reverse-engineering context from flattened text.
- It minimizes information loss and keeps future readers flexible.

Alternatives considered:
- Persist only `content` strings plus a few metadata fields. Rejected because it would immediately lose tool structure, message role details, and chunk boundaries.
- Persist renderer events. Rejected because presentation is already a transformed view and is not the system of record.

### 4. Record turn requests at the model call boundary

Before calling `o.model.Stream(ctx, state.Messages)`, the orchestrator should append a `turn_request` record that contains the full current `state.Messages` snapshot.

Rationale:
- This is the most faithful representation of what the model saw for that turn.
- It makes later debugging of tool-choice behavior possible without inference.
- It naturally captures the current system prompt, restored conversation, and any tool messages appended in prior turns.

Alternatives considered:
- Record only the final assistant output. Rejected because it omits the most valuable debugging input: the prompt context that produced the output.
- Record only diffs from the previous turn. Rejected because reconstructing complete state later would be more complex and error-prone.

### 5. Store final raw model messages instead of stream fragments

During `Stream(...)`, the orchestrator should still consume chunks for user-facing streaming output, but the audit trace should only persist the concatenated final assistant message for the turn.

Rationale:
- The audit objective is to preserve the effective model output structure, not the transport-level streaming fragments.
- Final assistant messages retain raw fields such as `tool_calls`, which are the relevant structures for debugging tool use.
- This keeps the trace aligned with the exact objects that drive orchestration decisions and reduces noise in the audit log.

Alternatives considered:
- Store every stream chunk. Rejected because chunk boundaries are implementation detail rather than durable conversational context.
- Store both chunks and final messages. Rejected because it duplicates the model output in two forms and makes audits harder to scan.

### 6. Let turn request snapshots carry tool context

The trace should rely on `turn_request` snapshots to preserve the actual inputs delivered to the model, including any `tool` messages produced by prior tool executions. Separate tool lifecycle records are not required for audit because the model-relevant context already appears in the next request snapshot, and model-originated tool instructions remain visible in the raw assistant message.

Rationale:
- The audit target is the actual model context, not every intermediate internal lifecycle event.
- `turn_request` plus `model_response` gives a complete before/after view at the model boundary.
- This preserves raw `tool_calls` and `tool` messages without duplicating them into parallel record types.

Alternatives considered:
- Store tool call, tool result, and tool message as separate trace entries. Rejected because they describe internal orchestration steps rather than the minimal model boundary needed for audit.
- Store only turn requests. Rejected because the raw model output structure would still be missing.

## Risks / Trade-offs

- [Risk] JSON serialization of `schema.Message` may include fields not needed for restoration or may evolve with upstream library changes. → Mitigation: keep restore logic narrowly scoped to the fields it actually needs and treat trace records as append-only observational data.
- [Risk] Excluding stream fragments means token-by-token generation order is no longer auditable. → Mitigation: define the audit scope around model-boundary inputs and outputs rather than transport-level streaming detail.
- [Risk] Two storage lanes can drift if writes are scattered across the orchestrator. → Mitigation: centralize conversation and trace append helpers behind a dedicated memory/session storage package.
- [Risk] Existing users may still have historical `context.txt` data they expect to load. → Mitigation: keep migration behavior explicit, either by treating the new format as applying to new sessions only or by preserving a compatibility loader if needed during rollout.
- [Risk] Session identifiers created per process run may not support continuing a prior session automatically. → Mitigation: define the first version around fresh-session storage and add explicit session-resume controls later if needed.

## Migration Plan

1. Introduce session metadata and a per-session directory layout under `checkpoints/sessions/`.
2. Add structured append helpers for conversation and trace records.
3. Update orchestrator entrypoints to create a session, write user and final assistant conversation records, and emit trace records at model and tool lifecycle boundaries.
4. Replace history loading so new runs read from `conversation.jsonl` rather than the legacy text file.
5. Keep the old `context.txt` path out of the write path for new sessions. If compatibility is needed, support it as a read-only fallback during transition.

Rollback is contained because the change is local persistence logic. Reverting to the old text history path only requires switching history loading and append calls back to the legacy file and ignoring the new session directories.

## Open Questions

- Should the application automatically mark one session as "latest" for resume, or should session resume remain manual for now?
- Do we want to include additional metadata such as configured model name and base URL in `meta.json` on day one, or keep metadata minimal?
- Should the first implementation read legacy `context.txt` as a fallback when no current session exists, or can the new session format start cleanly for fresh runs only?
