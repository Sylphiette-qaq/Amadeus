package orchestrator

import (
	"Amadeus/internal/memory"
	internaltool "Amadeus/internal/tool"
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	model "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func TestHandleTurnPersistsTraceAndRestoresConversationOnly(t *testing.T) {
	store, err := memory.NewStore(memory.Config{
		RootDir:   t.TempDir(),
		SessionID: "session-e2e",
		Now:       func() time.Time { return time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	model := &fakeModel{
		streams: []streamPlan{
			{
				chunks: []*schema.Message{
					schema.AssistantMessage("", []schema.ToolCall{
						{
							ID: "call-1",
							Function: schema.FunctionCall{
								Name:      "calculator",
								Arguments: `{"expression":"1+1"}`,
							},
						},
					}),
				},
			},
			{
				chunks: []*schema.Message{
					schema.AssistantMessage("The answer ", nil),
					schema.AssistantMessage("is 2", nil),
				},
			},
			{
				chunks: []*schema.Message{
					schema.AssistantMessage("You are welcome", nil),
				},
			},
		},
	}
	executor := &fakeExecutor{
		results: map[string]internaltool.Result{
			"calculator": {
				ToolName: "calculator",
				Success:  true,
				Data:     "2",
			},
		},
	}

	orch := &Orchestrator{
		model:      model,
		executor:   executor,
		store:      store,
		maxTurns:   4,
		systemText: "system prompt",
	}

	if err := orch.HandleTurn(context.Background(), "what is 1+1"); err != nil {
		t.Fatalf("HandleTurn(first) error = %v", err)
	}
	if err := orch.HandleTurn(context.Background(), "thanks"); err != nil {
		t.Fatalf("HandleTurn(second) error = %v", err)
	}

	history, err := store.LoadConversation()
	if err != nil {
		t.Fatalf("LoadConversation() error = %v", err)
	}
	if len(history) != 4 {
		t.Fatalf("restored history count = %d, want 4", len(history))
	}
	if history[0].Content != "what is 1+1" || history[1].Content != "The answer is 2" {
		t.Fatalf("unexpected first turn history: %+v", history[:2])
	}
	if history[2].Content != "thanks" || history[3].Content != "You are welcome" {
		t.Fatalf("unexpected second turn history: %+v", history[2:])
	}

	if len(model.calls) != 3 {
		t.Fatalf("model Stream() call count = %d, want 3", len(model.calls))
	}
	if got := countRole(model.calls[2], schema.Tool); got != 0 {
		t.Fatalf("restored context unexpectedly included tool messages: %d", got)
	}
	if len(model.calls[2]) != 4 {
		t.Fatalf("third model call message count = %d, want 4", len(model.calls[2]))
	}
	if model.calls[2][1].Content != "what is 1+1" || model.calls[2][2].Content != "The answer is 2" || model.calls[2][3].Content != "thanks" {
		t.Fatalf("unexpected restored context in third model call: %+v", model.calls[2])
	}

	traceLines := readLines(t, store.TracePath())
	if len(traceLines) != 6 {
		t.Fatalf("trace line count = %d, want 6", len(traceLines))
	}

	var traceRecords []memory.TraceRecord
	for _, line := range traceLines {
		var record memory.TraceRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("unmarshal trace record: %v", err)
		}
		traceRecords = append(traceRecords, record)
	}

	if traceRecords[0].Type != memory.RecordTypeTurnRequest {
		t.Fatalf("first trace record type = %q, want %q", traceRecords[0].Type, memory.RecordTypeTurnRequest)
	}
	if traceRecords[1].Type != memory.RecordTypeModelResponse {
		t.Fatalf("expected model response record at index 1, got %q", traceRecords[1].Type)
	}
	if len(traceRecords[1].Message.ToolCalls) != 1 {
		t.Fatalf("expected assistant message to retain tool_calls, got %+v", traceRecords[1].Message)
	}
	if traceRecords[2].Type != memory.RecordTypeTurnRequest {
		t.Fatalf("expected follow-up turn request at index 2, got %q", traceRecords[2].Type)
	}
	if got := countRole(traceRecords[2].Messages, schema.Tool); got != 1 {
		t.Fatalf("expected follow-up turn request to include one tool message, got %d", got)
	}
	if traceRecords[4].Type != memory.RecordTypeTurnRequest {
		t.Fatalf("expected second user-turn request at index 4, got %q", traceRecords[4].Type)
	}
}

func TestHandleTurnPersistsTurnError(t *testing.T) {
	store, err := memory.NewStore(memory.Config{
		RootDir:   t.TempDir(),
		SessionID: "session-error",
	})
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	orch := &Orchestrator{
		model: &fakeModel{
			streams: []streamPlan{
				{err: errors.New("stream failed")},
			},
		},
		executor:   &fakeExecutor{},
		store:      store,
		maxTurns:   2,
		systemText: "system prompt",
	}

	err = orch.HandleTurn(context.Background(), "hello")
	if err == nil {
		t.Fatal("HandleTurn() expected error, got nil")
	}

	traceLines := readLines(t, store.TracePath())
	if len(traceLines) != 2 {
		t.Fatalf("trace line count = %d, want 2", len(traceLines))
	}

	var last memory.TraceRecord
	if err := json.Unmarshal([]byte(traceLines[len(traceLines)-1]), &last); err != nil {
		t.Fatalf("unmarshal last trace record: %v", err)
	}
	if last.Type != memory.RecordTypeTurnError {
		t.Fatalf("last trace record type = %q, want %q", last.Type, memory.RecordTypeTurnError)
	}
}

type streamPlan struct {
	chunks []*schema.Message
	err    error
}

type fakeModel struct {
	streams []streamPlan
	calls   [][]*schema.Message
}

func (m *fakeModel) BindTools(_ []*schema.ToolInfo) error {
	return nil
}

func (m *fakeModel) Stream(_ context.Context, messages []*schema.Message, _ ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	m.calls = append(m.calls, cloneMessages(messages))
	if len(m.streams) == 0 {
		return nil, errors.New("no stream plan available")
	}

	next := m.streams[0]
	m.streams = m.streams[1:]
	if next.err != nil {
		return nil, next.err
	}

	return schema.StreamReaderFromArray(next.chunks), nil
}

type fakeExecutor struct {
	results map[string]internaltool.Result
}

func (e *fakeExecutor) ToolInfos() []*schema.ToolInfo {
	return nil
}

func (e *fakeExecutor) Execute(_ context.Context, toolName, _ string) (internaltool.Result, error) {
	if result, ok := e.results[toolName]; ok {
		return result, nil
	}
	return internaltool.Result{}, errors.New("unknown tool")
}

func cloneMessages(messages []*schema.Message) []*schema.Message {
	cloned := make([]*schema.Message, 0, len(messages))
	for _, message := range messages {
		if message == nil {
			cloned = append(cloned, nil)
			continue
		}

		copy := *message
		cloned = append(cloned, &copy)
	}
	return cloned
}

func countRole(messages []*schema.Message, role schema.RoleType) int {
	count := 0
	for _, message := range messages {
		if message != nil && message.Role == role {
			count++
		}
	}
	return count
}

func readLines(t *testing.T, path string) []string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	var lines []string
	start := 0
	for i, b := range data {
		if b == '\n' {
			if i > start {
				lines = append(lines, string(data[start:i]))
			}
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, string(data[start:]))
	}
	return lines
}
