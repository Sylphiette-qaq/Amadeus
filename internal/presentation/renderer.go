package presentation

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	viewModeChat         = "chat"
	viewModeTrace        = "trace"
	ansiDimGray          = "\033[90m"
	ansiResetStyle       = "\033[0m"
	defaultTerminalWidth = 100
	minContentWidth      = 20
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
	lineWidth     int
	lineVisible   int
	firstPrefix   string
	nextPrefix    string
	prefixPending bool
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

	return &Renderer{
		mode:      mode,
		lineWidth: loadTerminalWidth(),
	}
}

func loadViewMode() string {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("AMADEUS_CLI_VIEW")))
	if raw == viewModeTrace {
		return viewModeTrace
	}
	return viewModeChat
}

func loadTerminalWidth() int {
	raw := strings.TrimSpace(os.Getenv("COLUMNS"))
	if raw == "" {
		return defaultTerminalWidth
	}

	width, err := strconv.Atoi(raw)
	if err != nil || width <= 0 {
		return defaultTerminalWidth
	}

	return width
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
	r.lineVisible = 0
	r.firstPrefix = ""
	r.nextPrefix = ""
	r.prefixPending = false
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
		r.firstPrefix = "• "
		r.nextPrefix = "  "
		r.prefixPending = true
	}

	r.writeStreamDelta(delta, ansiDimGray)
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
		r.firstPrefix = "> "
		r.nextPrefix = "  "
		r.prefixPending = true
	}

	r.writeStreamDelta(delta, "")
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
	r.lineVisible = 0
	r.firstPrefix = ""
	r.nextPrefix = ""
	r.prefixPending = false
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
		r.toolsOpened = true
	}
	r.toolCount++

	if r.mode == viewModeTrace {
		r.writeWrappedBlock("• ", "  ", fmt.Sprintf("%d. %s", r.toolCount, event.ToolCall.Function.Name), ansiDimGray)
		r.writeWrappedBlock("  args: ", "        ", event.ToolCall.Function.Arguments, ansiDimGray)
		r.writeWrappedBlock("  status: ", "          ", "running", ansiDimGray)
		return
	}

	r.writeWrappedBlock("• ", "  ", fmt.Sprintf("%d. %s · running", r.toolCount, event.ToolCall.Function.Name), ansiDimGray)
}

func (r *Renderer) finishTool(event Event) {
	r.finishAssistant()

	summary := summarize(event.Content)
	if r.mode == viewModeTrace {
		r.writeWrappedBlock("  result: ", "          ", fmt.Sprintf("%s · success=%t", event.ToolName, event.Success), ansiDimGray)
		r.writeWrappedBlock("  body: ", "        ", event.Content, ansiDimGray)
		return
	}

	r.writeWrappedBlock("  result: ", "          ", fmt.Sprintf("%s · success=%t", event.ToolName, event.Success), ansiDimGray)
	r.writeWrappedBlock("  summary: ", "           ", summary, ansiDimGray)
}

func (r *Renderer) failTurn(err error) {
	r.finishAssistant()
	if !r.startedOutput {
		fmt.Println()
		r.startedOutput = true
	}
	r.writeWrappedBlock("> ", "  ", "error", "")
	r.writeWrappedBlock("  ", "  ", err.Error(), "")
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

func (r *Renderer) writeStreamDelta(delta, style string) {
	for _, ch := range delta {
		if ch == '\n' {
			r.ensureStyle(style)
			if r.atLineStart {
				fmt.Print(r.currentPrefix())
				r.markPrefixUsed()
			}
			fmt.Printf("%c", ch)
			if r.activeStyle != "" {
				fmt.Print(ansiResetStyle)
				r.activeStyle = ""
			}
			r.atLineStart = true
			r.lineVisible = 0
			continue
		}

		r.ensureWrappedPrefix(style)
		fmt.Printf("%c", ch)
		r.lineVisible++
	}
}

func (r *Renderer) ensureStyle(style string) {
	if style != "" && r.activeStyle != style {
		fmt.Print(style)
		r.activeStyle = style
	}
	if style == "" && r.activeStyle != "" {
		fmt.Print(ansiResetStyle)
		r.activeStyle = ""
	}
}

func (r *Renderer) ensureWrappedPrefix(style string) {
	availableWidth := r.availableWidth(r.currentPrefix())
	if !r.atLineStart && r.lineVisible >= availableWidth {
		if r.activeStyle != "" {
			fmt.Print(ansiResetStyle)
			r.activeStyle = ""
		}
		fmt.Println()
		r.atLineStart = true
		r.lineVisible = 0
	}

	r.ensureStyle(style)
	if r.atLineStart {
		fmt.Print(r.currentPrefix())
		r.atLineStart = false
		r.lineVisible = 0
		r.markPrefixUsed()
	}
}

func (r *Renderer) currentPrefix() string {
	if r.prefixPending {
		return r.firstPrefix
	}
	return r.nextPrefix
}

func (r *Renderer) markPrefixUsed() {
	r.prefixPending = false
}

func (r *Renderer) availableWidth(prefix string) int {
	width := r.lineWidth - len([]rune(prefix))
	if width < minContentWidth {
		return minContentWidth
	}
	return width
}

func (r *Renderer) writeWrappedBlock(firstPrefix, continuationPrefix, content, style string) {
	if content == "" {
		content = "<empty>"
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			r.writeWrappedLine(firstPrefix, continuationPrefix, "", style)
			continue
		}
		r.writeWrappedLine(firstPrefix, continuationPrefix, line, style)
		firstPrefix = continuationPrefix
	}
}

func (r *Renderer) writeWrappedLine(firstPrefix, continuationPrefix, line, style string) {
	if line == "" {
		r.ensureStyle(style)
		fmt.Printf("%s\n", firstPrefix)
		if r.activeStyle != "" {
			fmt.Print(ansiResetStyle)
			r.activeStyle = ""
		}
		r.atLineStart = true
		r.lineVisible = 0
		return
	}

	runes := []rune(line)
	currentPrefix := firstPrefix
	for len(runes) > 0 {
		availableWidth := r.availableWidth(currentPrefix)
		chunkLen := availableWidth
		if chunkLen > len(runes) {
			chunkLen = len(runes)
		}

		r.ensureStyle(style)
		fmt.Printf("%s%s\n", currentPrefix, string(runes[:chunkLen]))
		if r.activeStyle != "" {
			fmt.Print(ansiResetStyle)
			r.activeStyle = ""
		}
		runes = runes[chunkLen:]
		currentPrefix = continuationPrefix
	}

	r.atLineStart = true
	r.lineVisible = 0
}
