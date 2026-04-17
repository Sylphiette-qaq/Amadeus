package presentation

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	defer reader.Close()

	os.Stdout = writer
	defer func() {
		os.Stdout = originalStdout
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	return string(output)
}

func stripANSI(text string) string {
	replacer := strings.NewReplacer(ansiDimGray, "", ansiResetStyle, "")
	return replacer.Replace(text)
}

func TestToolOutputUsesDimGrayStyle(t *testing.T) {
	renderer := NewRenderer(viewModeChat)
	renderer.lineWidth = 80

	output := captureStdout(t, func() {
		renderer.Emit(Event{Type: EventTurnStarted})
		renderer.Emit(Event{
			Type: EventToolCallStarted,
			ToolCall: schema.ToolCall{
				Function: schema.FunctionCall{Name: "search_docs"},
			},
		})
		renderer.Emit(Event{
			Type:     EventToolCallFinished,
			ToolName: "search_docs",
			Success:  true,
			Content:  "matched one result",
		})
	})
	plain := stripANSI(output)

	if !strings.Contains(plain, "• 1. search_docs · running\n") {
		t.Fatalf("expected tool status to use dim style, output=%q", output)
	}
	if strings.Contains(plain, "• tools\n") {
		t.Fatalf("expected tools header to be removed, output=%q", output)
	}
	if !strings.Contains(plain, "  summary: matched one result\n") {
		t.Fatalf("expected tool summary to use dim style, output=%q", output)
	}
}

func TestReasoningAndToolOutputUseSingleLeadingMarker(t *testing.T) {
	renderer := NewRenderer(viewModeChat)
	renderer.lineWidth = 24

	output := captureStdout(t, func() {
		renderer.Emit(Event{Type: EventTurnStarted})
		renderer.Emit(Event{
			Type:    EventReasoningDelta,
			Content: "abcdefghijklmnopqrstuvw",
		})
		renderer.Emit(Event{Type: EventAssistantFinal})
		renderer.Emit(Event{
			Type:    EventAnswerDelta,
			Content: "abcdefghijklmnopqrstuvw",
		})
		renderer.Emit(Event{Type: EventAssistantFinal})
		renderer.Emit(Event{
			Type: EventToolCallStarted,
			ToolCall: schema.ToolCall{
				Function: schema.FunctionCall{Name: "lookup"},
			},
		})
		renderer.Emit(Event{
			Type:     EventToolCallFinished,
			ToolName: "lookup",
			Success:  true,
			Content:  strings.Repeat("x", 40),
		})
	})
	plain := stripANSI(output)

	if !strings.Contains(plain, "• abcdefghijklmnopqrstuv\n  w\n") {
		t.Fatalf("expected reasoning output to wrap with a single bullet, output=%q", output)
	}
	if !strings.Contains(plain, "> abcdefghijklmnopqrstuv\n  w\n") {
		t.Fatalf("expected answer output to wrap with only the first line using >, output=%q", output)
	}
	if !strings.Contains(plain, "  summary: xxxxxxxxxxxxxxxxxxxx\n           xxxxxxxxxxxxxxxxxxxx") {
		t.Fatalf("expected tool summary output to wrap without repeating bullet, output=%q", output)
	}
	if strings.Contains(plain, "\n> w") || strings.Contains(plain, "\n• w") {
		t.Fatalf("expected continuation lines to omit repeated leading markers, output=%q", output)
	}
}
