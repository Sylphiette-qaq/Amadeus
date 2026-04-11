package session

import "github.com/cloudwego/eino/schema"

type State struct {
	Messages       []*schema.Message
	CurrentTurn    int
	ToolCallCount  int
	Finished       bool
	LastToolResult string
}

func NewState(history []*schema.Message, systemText, userQuestion string) *State {
	messages := make([]*schema.Message, 0, len(history)+2)
	messages = append(messages, schema.SystemMessage(systemText))
	messages = append(messages, history...)
	messages = append(messages, schema.UserMessage(userQuestion))

	return &State{
		Messages: messages,
	}
}

func (s *State) Append(message *schema.Message) {
	if message == nil {
		return
	}

	s.Messages = append(s.Messages, message)
}
