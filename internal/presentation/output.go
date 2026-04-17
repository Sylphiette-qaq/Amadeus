package presentation

import "github.com/cloudwego/eino/schema"

func PrintToolCall(toolCall schema.ToolCall) {
	Emit(Event{
		Type:     EventToolCallStarted,
		ToolCall: toolCall,
	})
}

func PrintToolResult(toolName string, success bool, content string) {
	Emit(Event{
		Type:     EventToolCallFinished,
		ToolName: toolName,
		Success:  success,
		Content:  content,
	})
}

func PrintAssistantResponse(content string) {
	if content == "" {
		return
	}

	Emit(Event{
		Type:    EventAnswerDelta,
		Content: content,
	})
	Emit(Event{Type: EventAssistantFinal})
}
