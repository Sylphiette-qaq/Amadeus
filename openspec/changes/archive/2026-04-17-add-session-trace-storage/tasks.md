## 1. Session Storage Foundation

- [x] 1.1 Add a session-scoped storage model under `checkpoints/sessions/<session_id>/` with metadata initialization for each CLI process run
- [x] 1.2 Introduce structured record types and append helpers for `conversation.jsonl` and `trace.jsonl`
- [x] 1.3 Thread the active `session_id` and storage handles through startup and orchestrator construction

## 2. Conversation Persistence And Restore

- [x] 2.1 Replace legacy `context.txt` writes with structured conversation writes for user messages and final assistant messages only
- [x] 2.2 Update history loading to rebuild context from `conversation.jsonl` and restore only `user_message` and `assistant_final` records
- [x] 2.3 Decide and implement transitional behavior for legacy `context.txt` reads, either as explicit fallback or as a clean cutover for new sessions

## 3. Raw Trace Recording

- [x] 3.1 Persist a `turn_request` trace record containing the full `state.Messages` snapshot before each model invocation
- [x] 3.2 Persist every streamed model chunk and the final concatenated assistant message as separate trace records for each turn
- [x] 3.3 Persist raw `tool_call`, `tool_result`, and `tool_message` records at their respective orchestrator lifecycle boundaries
- [x] 3.4 Persist turn-level failure records so unsuccessful runs remain auditable

## 4. Verification

- [x] 4.1 Add tests for structured conversation persistence and restore behavior, including exclusion of tool and intermediate assistant records
- [x] 4.2 Add tests for trace record writing across turn requests, stream chunks, tool lifecycle events, and turn failures
- [x] 4.3 Verify an end-to-end multi-turn tool-using session writes the expected session directory contents and still restores only user and final assistant history
