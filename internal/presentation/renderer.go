package presentation

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

const (
	viewModeChat   = "chat"
	viewModeTrace  = "trace"
	ansiDimGray    = "\033[90m"
	ansiResetStyle = "\033[0m"
)

type Renderer struct {
	mu sync.Mutex

	mode string
	turn int

	toolsOpened   bool
	reasoningSeen bool
	answerSeen    bool
	atLineStart   bool
	toolCount     int
	startedOutput bool
	activeSection sectionKind
	activeStyle   string
}

var defaultRenderer = NewRenderer(loadViewMode())

type sectionKind string

const (
	sectionNone      sectionKind = ""
	sectionReasoning sectionKind = "reasoning"
	sectionAnswer    sectionKind = "answer"
)

func NewRenderer(mode string) *Renderer {
	if mode != viewModeTrace {
		mode = viewModeChat
	}

	return &Renderer{mode: mode}
}

func loadViewMode() string {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("AMADEUS_CLI_VIEW")))
	if raw == viewModeTrace {
		return viewModeTrace
	}
	return viewModeChat
}

func Emit(event Event) {
	defaultRenderer.Emit(event)
}

func PrintTurnError(err error) {
	if err == nil {
		return
	}
	Emit(Event{
		Type:  EventTurnFailed,
		Error: err,
	})
}

func (r *Renderer) Emit(event Event) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch event.Type {
	case EventTurnStarted:
		r.startTurn(event)
	case EventReasoningDelta:
		r.writeReasoning(event.Content)
	case EventAnswerDelta:
		r.writeAnswer(event.Content)
	case EventAssistantFinal:
		r.completeTurn()
	case EventToolCallStarted:
		r.startTool(event)
	case EventToolCallFinished:
		r.finishTool(event)
	case EventTurnFailed:
		r.failTurn(event.Error)
	}
}

func (r *Renderer) startTurn(event Event) {
	if event.Turn > 0 {
		r.turn = event.Turn
	} else {
		r.turn++
	}
	r.toolsOpened = false
	r.reasoningSeen = false
	r.answerSeen = false
	r.atLineStart = true
	r.toolCount = 0
	r.startedOutput = false
	r.activeSection = sectionNone
	r.activeStyle = ""
}

func (r *Renderer) writeReasoning(delta string) {
	if delta == "" {
		return
	}

	if r.activeSection != sectionReasoning {
		r.finishAssistant()
		if !r.startedOutput {
			fmt.Println()
			r.startedOutput = true
		} else if !r.reasoningSeen {
			fmt.Println()
		}
		r.reasoningSeen = true
		r.activeSection = sectionReasoning
		r.atLineStart = true
	}

	r.writeStreamDelta("> ", delta, ansiDimGray)
}

func (r *Renderer) writeAnswer(delta string) {
	if delta == "" {
		return
	}

	if r.activeSection != sectionAnswer {
		wasReasoning := r.activeSection == sectionReasoning
		r.finishAssistant()
		if !r.startedOutput {
			fmt.Println()
			r.startedOutput = true
		} else if wasReasoning {
			fmt.Println()
		}
		r.answerSeen = true
		r.activeSection = sectionAnswer
		r.atLineStart = true
	}

	r.writeStreamDelta("> ", delta, "")
}

func (r *Renderer) finishAssistant() {
	if r.activeStyle != "" {
		fmt.Print(ansiResetStyle)
		r.activeStyle = ""
	}
	if !r.atLineStart {
		fmt.Println()
	}
	r.atLineStart = true
	r.activeSection = sectionNone
}

func (r *Renderer) completeTurn() {
	r.finishAssistant()
}

func (r *Renderer) startTool(event Event) {
	r.finishAssistant()

	if !r.toolsOpened {
		if !r.startedOutput {
			fmt.Println()
			r.startedOutput = true
		}
		fmt.Println("> tools")
		r.toolsOpened = true
	}
	r.toolCount++

	if r.mode == viewModeTrace {
		fmt.Printf("> %d. %s\n", r.toolCount, event.ToolCall.Function.Name)
		r.writeBlock(">    args: ", event.ToolCall.Function.Arguments)
		fmt.Println(">    status: running")
		return
	}

	fmt.Printf("> %d. %s · running\n", r.toolCount, event.ToolCall.Function.Name)
}

func (r *Renderer) finishTool(event Event) {
	r.finishAssistant()

	summary := summarize(event.Content)
	if r.mode == viewModeTrace {
		fmt.Printf(">    result: %s · success=%t\n", event.ToolName, event.Success)
		r.writeBlock(">    body: ", event.Content)
		return
	}

	fmt.Printf(">    result: %s · success=%t\n", event.ToolName, event.Success)
	r.writeBlock(">    summary: ", summary)
}

func (r *Renderer) failTurn(err error) {
	r.finishAssistant()
	if !r.startedOutput {
		fmt.Println()
		r.startedOutput = true
	}
	fmt.Println("> error")
	r.writeBlock("> ", err.Error())
	r.completeTurn()
}

func summarize(content string) string {
	normalized := strings.Join(strings.Fields(content), " ")
	if normalized == "" {
		return "<empty>"
	}

	const maxLen = 140
	if len(normalized) <= maxLen {
		return normalized
	}

	return normalized[:maxLen-3] + "..."
}

func (r *Renderer) writeStreamDelta(prefix, delta, style string) {
	for _, ch := range delta {
		if style != "" && r.activeStyle != style {
			fmt.Print(style)
			r.activeStyle = style
		}
		if style == "" && r.activeStyle != "" {
			fmt.Print(ansiResetStyle)
			r.activeStyle = ""
		}

		if r.atLineStart {
			fmt.Print(prefix)
			r.atLineStart = false
		}

		fmt.Printf("%c", ch)
		if ch == '\n' {
			if r.activeStyle != "" {
				fmt.Print(ansiResetStyle)
				r.activeStyle = ""
			}
			r.atLineStart = true
		}
	}
}

func (r *Renderer) writeBlock(prefix, content string) {
	if content == "" {
		fmt.Printf("%s<empty>\n", prefix)
		return
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			fmt.Printf("%s\n", prefix)
			continue
		}
		fmt.Printf("%s%s\n", prefix, line)
	}
}
