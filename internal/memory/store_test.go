package memory

import (
	"Amadeus/internal/skill"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
)

func TestStoreInitializesSessionAndLoadsConversation(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	store, err := NewStore(Config{
		RootDir:   t.TempDir(),
		SessionID: "session-test",
		Model:     "deepseek-reasoner",
		BaseURL:   "https://api.deepseek.com",
		Now:       func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(store.SessionDir(), "meta.json")); err != nil {
		t.Fatalf("meta.json missing: %v", err)
	}

	if err := store.AppendUserMessage(0, schema.UserMessage("hello")); err != nil {
		t.Fatalf("AppendUserMessage() error = %v", err)
	}
	if err := store.AppendAssistantFinal(1, schema.AssistantMessage("hi", nil)); err != nil {
		t.Fatalf("AppendAssistantFinal() error = %v", err)
	}

	messages, err := store.LoadConversation()
	if err != nil {
		t.Fatalf("LoadConversation() error = %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("LoadConversation() count = %d, want 2", len(messages))
	}
	if messages[0].Role != schema.User || messages[0].Content != "hello" {
		t.Fatalf("unexpected first message: %+v", messages[0])
	}
	if messages[1].Role != schema.Assistant || messages[1].Content != "hi" {
		t.Fatalf("unexpected second message: %+v", messages[1])
	}
}

func TestStoreWritesTraceRecords(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	store, err := NewStore(Config{
		RootDir:   t.TempDir(),
		SessionID: "trace-test",
		Now:       func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	if err := store.AppendTurnRequest(1, []*schema.Message{schema.SystemMessage("system"), schema.UserMessage("question")}); err != nil {
		t.Fatalf("AppendTurnRequest() error = %v", err)
	}
	if err := store.AppendModelResponse(1, schema.AssistantMessage("", []schema.ToolCall{{
		ID: "call-1",
		Function: schema.FunctionCall{
			Name:      "calculator",
			Arguments: `{"expression":"1+1"}`,
		},
	}})); err != nil {
		t.Fatalf("AppendModelResponse() error = %v", err)
	}
	if err := store.AppendTurnError(1, assertErr("boom")); err != nil {
		t.Fatalf("AppendTurnError() error = %v", err)
	}

	lines := readJSONLines(t, store.TracePath())
	if len(lines) != 3 {
		t.Fatalf("trace record count = %d, want 3", len(lines))
	}

	var first TraceRecord
	if err := json.Unmarshal(lines[0], &first); err != nil {
		t.Fatalf("unmarshal first trace record: %v", err)
	}
	if first.Type != RecordTypeTurnRequest || len(first.Messages) != 2 {
		t.Fatalf("unexpected first trace record: %+v", first)
	}

	var middle TraceRecord
	if err := json.Unmarshal(lines[1], &middle); err != nil {
		t.Fatalf("unmarshal middle trace record: %v", err)
	}
	if middle.Type != RecordTypeModelResponse || middle.Message == nil || len(middle.Message.ToolCalls) != 1 {
		t.Fatalf("unexpected middle trace record: %+v", middle)
	}

	var last TraceRecord
	if err := json.Unmarshal(lines[2], &last); err != nil {
		t.Fatalf("unmarshal last trace record: %v", err)
	}
	if last.Type != RecordTypeTurnError || last.Error != "boom" {
		t.Fatalf("unexpected last trace record: %+v", last)
	}
}

func TestStorePersistsLoadedSkillsSeparately(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	store, err := NewStore(Config{
		RootDir:   t.TempDir(),
		SessionID: "skill-test",
		Now:       func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	doc := skill.Document{
		Name:    "openspec-explore",
		Path:    "/tmp/openspec-explore/SKILL.md",
		Content: "# Explore",
	}
	if err := store.AppendLoadedSkill(2, doc); err != nil {
		t.Fatalf("AppendLoadedSkill() error = %v", err)
	}
	if err := store.AppendLoadedSkill(3, doc); err != nil {
		t.Fatalf("AppendLoadedSkill() duplicate error = %v", err)
	}

	loaded, err := store.LoadLoadedSkills()
	if err != nil {
		t.Fatalf("LoadLoadedSkills() error = %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("LoadLoadedSkills() count = %d, want 1", len(loaded))
	}
	if loaded[0] != doc {
		t.Fatalf("unexpected loaded skill: %+v", loaded[0])
	}

	lines := readJSONLines(t, store.LoadedSkillsPath())
	if len(lines) != 2 {
		t.Fatalf("loaded skill record count = %d, want 2", len(lines))
	}

	var first LoadedSkillRecord
	if err := json.Unmarshal(lines[0], &first); err != nil {
		t.Fatalf("unmarshal loaded skill record: %v", err)
	}
	if first.Type != RecordTypeLoadedSkill || first.SkillName != doc.Name || first.Content != doc.Content {
		t.Fatalf("unexpected loaded skill record: %+v", first)
	}
}

type assertErr string

func (e assertErr) Error() string {
	return string(e)
}

func readJSONLines(t *testing.T, path string) [][]byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	rawLines := bytesSplitLines(data)
	lines := make([][]byte, 0, len(rawLines))
	for _, line := range rawLines {
		if len(line) == 0 {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func bytesSplitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
