## ADDED Requirements

### Requirement: The system SHALL create session-scoped storage for each CLI run
The system SHALL create a unique `session_id` for each CLI process run and SHALL write persistent session artifacts under a session-specific storage location instead of appending all runs into a single shared text file.

#### Scenario: Create storage for a new session
- **WHEN** the CLI starts a new process run
- **THEN** the system SHALL generate a `session_id`
- **THEN** the system SHALL create storage for that session under `checkpoints/sessions/<session_id>/`

#### Scenario: Persist session metadata
- **WHEN** a new session is initialized
- **THEN** the system SHALL persist session metadata that includes the `session_id`

### Requirement: The system SHALL persist restorable conversation history separately from trace data
The system SHALL maintain a conversation history lane that is distinct from raw orchestration trace storage and SHALL use that conversation lane as the only source for restoring future context.

#### Scenario: Store user input in conversation history
- **WHEN** a user submits a prompt
- **THEN** the system SHALL append that user message to the session conversation history

#### Scenario: Store only final assistant replies in conversation history
- **WHEN** the orchestrator completes a user turn with a final assistant answer that has no further `tool_calls`
- **THEN** the system SHALL append that final assistant message to the session conversation history

#### Scenario: Exclude internal orchestration records from restored history
- **WHEN** the system rebuilds conversation history for a future user turn
- **THEN** it SHALL load only user messages and final assistant answers from the conversation history
- **THEN** it SHALL exclude streamed chunks, intermediate assistant tool-routing messages, tool messages, tool results, and errors from restored context

### Requirement: The system SHALL persist the exact model request context for every turn
For every orchestrator turn that invokes the model, the system SHALL persist the full message array that is sent to the model without summarization or presentation-oriented transformation.

#### Scenario: Record turn request before model invocation
- **WHEN** the orchestrator is about to invoke the model for a turn
- **THEN** the system SHALL append a trace record that contains the turn number and the full `state.Messages` snapshot used for that model request

#### Scenario: Preserve system and tool context in trace requests
- **WHEN** a turn request contains system messages, restored conversation history, or prior tool messages
- **THEN** the persisted trace record SHALL retain those messages in the same request snapshot

### Requirement: The system SHALL persist streamed model output in raw form
The system SHALL persist the final raw assistant message for each turn in the same structured form used by orchestration decisions, including any tool-calling fields returned by the model.

#### Scenario: Record the final assistant message for the turn
- **WHEN** the model stream completes and the orchestrator concatenates the chunks into a final assistant message
- **THEN** the system SHALL append a `model_response` trace record containing that concatenated assistant message

#### Scenario: Preserve raw tool-calling structure in assistant messages
- **WHEN** the model returns a structured assistant message that includes `tool_calls`
- **THEN** the persisted assistant trace record SHALL retain those `tool_calls` in the raw assistant message structure

### Requirement: The system SHALL audit model-boundary context rather than stream fragments
The system SHALL structure its trace around the actual input and output objects at the model boundary instead of persisting transport-level stream fragments or duplicated internal lifecycle records.

#### Scenario: Preserve tool messages in the next model request snapshot
- **WHEN** a prior turn appends one or more `tool` messages before the next model invocation
- **THEN** the next persisted `turn_request` trace record SHALL retain those `tool` messages in the full request snapshot

#### Scenario: Do not emit chunk-level trace records
- **WHEN** the orchestrator receives streamed chunks while building a final assistant message
- **THEN** the audit trace SHALL not create separate per-chunk trace records for those fragments

### Requirement: The system SHALL persist turn failures as trace records
The system SHALL persist turn-level orchestration failures so that unsuccessful runs remain auditable.

#### Scenario: Record turn failure
- **WHEN** the orchestrator fails before producing a final assistant answer for a user turn
- **THEN** the system SHALL append a trace record for that turn failure including the associated session and turn identity
