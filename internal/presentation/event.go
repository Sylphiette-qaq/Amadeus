package presentation

import "github.com/cloudwego/eino/schema"

type EventType string

const (
	EventTurnStarted      EventType = "turn_started"
	EventReasoningDelta   EventType = "reasoning_delta"
	EventAnswerDelta      EventType = "answer_delta"
	EventAssistantFinal   EventType = "assistant_final"
	EventToolCallStarted  EventType = "tool_call_started"
	EventToolCallFinished EventType = "tool_call_finished"
	EventTurnFailed       EventType = "turn_failed"
)

type Event struct {
	Type    EventType
	Turn    int
	Content string
	Error   error

	ToolCall schema.ToolCall
	ToolName string
	Success  bool
}
