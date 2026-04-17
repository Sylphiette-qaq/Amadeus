## ADDED Requirements

### Requirement: The CLI model path SHALL use streaming output from `deepseek-reasoner`
The system SHALL use `deepseek-reasoner` as the default model for the CLI interaction path and SHALL consume model output through a streaming interface so the panel can update while generation is in progress.

#### Scenario: Default model uses reasoner
- **WHEN** the CLI starts without an explicit model override
- **THEN** the model layer SHALL default to `deepseek-reasoner` for assistant turns

#### Scenario: Panel updates while model is generating
- **WHEN** the model is producing an answer for a user turn
- **THEN** the CLI interaction panel SHALL render incremental output updates before the final assistant message is fully complete

### Requirement: CLI turns SHALL be rendered as structured interaction blocks
The system SHALL render each user request as a structured interaction block that clearly separates user input, agent progress, and the final assistant answer instead of printing the exchange as undifferentiated terminal lines.

#### Scenario: Render a normal turn
- **WHEN** a user submits a prompt and the agent completes the turn without errors
- **THEN** the terminal output SHALL present the user input and assistant answer as distinct sections within the same turn block

#### Scenario: Render multiple turns
- **WHEN** a user completes multiple prompts in the same session
- **THEN** each prompt-response cycle SHALL be rendered as its own distinct turn block so operators can identify turn boundaries

### Requirement: The system SHALL surface model reasoning content from `reasoning_content`
The system SHALL read model reasoning text from `schema.Message.ReasoningContent` and present it as a dedicated output channel rather than dropping it or merging it into the final answer body.

#### Scenario: Render reasoning content during a turn
- **WHEN** the streamed or final assistant message includes `reasoning_content`
- **THEN** the interaction panel SHALL render that reasoning content in a dedicated reasoning section for the active turn

#### Scenario: Keep reasoning separate from final answer
- **WHEN** both answer content and reasoning content are available for the same assistant turn
- **THEN** the final answer section SHALL remain separate from the reasoning section in the rendered panel

### Requirement: Tool activity SHALL be presented as explicit execution status
The system SHALL present tool activity as explicit execution status information rather than raw, unstructured payload dumps.

#### Scenario: Show tool invocation progress
- **WHEN** the orchestrator emits one or more tool calls during a turn
- **THEN** the terminal output SHALL identify each tool by name and show whether it is running, succeeded, or failed

#### Scenario: Show concise tool outcomes in default view
- **WHEN** a tool call finishes in the default interaction view
- **THEN** the terminal output SHALL show a concise summary of the tool outcome instead of the full raw tool arguments and full raw result payload

### Requirement: The default panel SHALL prioritize readability over raw trace detail
The system SHALL provide a default interaction view that emphasizes conversation flow and concise execution summaries.

#### Scenario: Hide raw payloads in default view
- **WHEN** the system is running in the default interaction view
- **THEN** detailed raw tool arguments and raw tool result bodies SHALL not be printed inline as the primary output format

#### Scenario: Preserve final answer prominence
- **WHEN** a turn includes both tool activity and a final assistant answer
- **THEN** the final assistant answer SHALL be rendered in a dedicated answer section that remains visually prominent relative to tool trace data

#### Scenario: Preserve reasoning readability in default view
- **WHEN** the system is running in the default interaction view and reasoning content is available
- **THEN** the panel SHALL render reasoning in a readable dedicated section without replacing the final answer as the primary user-facing outcome

### Requirement: The system SHALL support a detailed trace view for debugging
The system SHALL provide a detailed trace-oriented rendering mode for developers who need to inspect tool execution details.

#### Scenario: Show detailed tool arguments in trace view
- **WHEN** the system is running in trace view and a tool call is emitted
- **THEN** the terminal output SHALL include the tool arguments associated with that tool call

#### Scenario: Show detailed tool results in trace view
- **WHEN** the system is running in trace view and a tool call finishes
- **THEN** the terminal output SHALL include the detailed tool result content associated with that tool call

#### Scenario: Show detailed reasoning in trace view
- **WHEN** the system is running in trace view and the assistant emits reasoning content
- **THEN** the terminal output SHALL include the detailed reasoning content associated with that assistant turn

### Requirement: Errors SHALL be rendered as first-class panel states
The system SHALL render execution failures and user-visible errors as explicit error states rather than mixing them into normal assistant output.

#### Scenario: Tool failure appears as error state
- **WHEN** a tool call fails during a turn
- **THEN** the panel SHALL render that failure as an error state associated with the tool entry for that turn

#### Scenario: Turn failure appears as turn-level error
- **WHEN** the orchestrator fails before producing a final assistant answer
- **THEN** the panel SHALL render a turn-level error section that distinguishes the failure from a normal assistant response
