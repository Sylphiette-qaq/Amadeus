package memory

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

const defaultSessionsRoot = "./checkpoints/sessions"

const (
	RecordTypeUserMessage    = "user_message"
	RecordTypeAssistantFinal = "assistant_final"
	RecordTypeTurnRequest    = "turn_request"
	RecordTypeModelResponse  = "model_response"
	RecordTypeTurnError      = "turn_error"
)

type Config struct {
	RootDir   string
	SessionID string
	Model     string
	BaseURL   string
	Now       func() time.Time
}

type Meta struct {
	SessionID string `json:"session_id"`
	StartedAt string `json:"started_at"`
	Model     string `json:"model,omitempty"`
	BaseURL   string `json:"base_url,omitempty"`
}

type ConversationRecord struct {
	Type      string          `json:"type"`
	SessionID string          `json:"session_id"`
	Turn      int             `json:"turn"`
	Timestamp string          `json:"timestamp"`
	Message   *schema.Message `json:"message"`
}

type TraceRecord struct {
	Type      string            `json:"type"`
	SessionID string            `json:"session_id"`
	Turn      int               `json:"turn"`
	Timestamp string            `json:"timestamp"`
	Messages  []*schema.Message `json:"messages,omitempty"`
	Message   *schema.Message   `json:"message,omitempty"`
	Error     string            `json:"error,omitempty"`
}

type Store struct {
	mu               sync.Mutex
	now              func() time.Time
	rootDir          string
	sessionID        string
	sessionDir       string
	metaPath         string
	conversationPath string
	tracePath        string
}

func NewStore(cfg Config) (*Store, error) {
	now := cfg.Now
	if now == nil {
		now = time.Now
	}

	rootDir := cfg.RootDir
	if rootDir == "" {
		rootDir = defaultSessionsRoot
	}

	sessionID := cfg.SessionID
	if sessionID == "" {
		sessionID = generateSessionID(now)
	}

	sessionDir := filepath.Join(rootDir, sessionID)
	store := &Store{
		now:              now,
		rootDir:          rootDir,
		sessionID:        sessionID,
		sessionDir:       sessionDir,
		metaPath:         filepath.Join(sessionDir, "meta.json"),
		conversationPath: filepath.Join(sessionDir, "conversation.jsonl"),
		tracePath:        filepath.Join(sessionDir, "trace.jsonl"),
	}

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("create session dir: %w", err)
	}

	meta := Meta{
		SessionID: sessionID,
		StartedAt: now().Format(time.RFC3339Nano),
		Model:     cfg.Model,
		BaseURL:   cfg.BaseURL,
	}

	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal session meta: %w", err)
	}

	if err := os.WriteFile(store.metaPath, append(metaBytes, '\n'), 0644); err != nil {
		return nil, fmt.Errorf("write session meta: %w", err)
	}

	return store, nil
}

func (s *Store) SessionID() string {
	return s.sessionID
}

func (s *Store) SessionDir() string {
	return s.sessionDir
}

func (s *Store) ConversationPath() string {
	return s.conversationPath
}

func (s *Store) TracePath() string {
	return s.tracePath
}

func (s *Store) AppendUserMessage(turn int, message *schema.Message) error {
	return s.appendConversation(RecordTypeUserMessage, turn, message)
}

func (s *Store) AppendAssistantFinal(turn int, message *schema.Message) error {
	return s.appendConversation(RecordTypeAssistantFinal, turn, message)
}

func (s *Store) LoadConversation() ([]*schema.Message, error) {
	file, err := os.Open(s.conversationPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open conversation history: %w", err)
	}
	defer file.Close()

	var messages []*schema.Message
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var record ConversationRecord
		if err := json.Unmarshal(line, &record); err != nil {
			return nil, fmt.Errorf("parse conversation record: %w", err)
		}

		if record.Type != RecordTypeUserMessage && record.Type != RecordTypeAssistantFinal {
			continue
		}
		if record.Message == nil {
			continue
		}

		messages = append(messages, record.Message)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan conversation history: %w", err)
	}

	return messages, nil
}

func (s *Store) AppendTurnRequest(turn int, messages []*schema.Message) error {
	return s.appendTrace(TraceRecord{
		Type:      RecordTypeTurnRequest,
		SessionID: s.sessionID,
		Turn:      turn,
		Timestamp: s.timestamp(),
		Messages:  messages,
	})
}

func (s *Store) AppendModelResponse(turn int, message *schema.Message) error {
	return s.appendTrace(TraceRecord{
		Type:      RecordTypeModelResponse,
		SessionID: s.sessionID,
		Turn:      turn,
		Timestamp: s.timestamp(),
		Message:   message,
	})
}

func (s *Store) AppendTurnError(turn int, err error) error {
	if err == nil {
		return nil
	}

	return s.appendTrace(TraceRecord{
		Type:      RecordTypeTurnError,
		SessionID: s.sessionID,
		Turn:      turn,
		Timestamp: s.timestamp(),
		Error:     err.Error(),
	})
}

func (s *Store) appendConversation(recordType string, turn int, message *schema.Message) error {
	if message == nil {
		return nil
	}

	return s.appendJSONL(s.conversationPath, ConversationRecord{
		Type:      recordType,
		SessionID: s.sessionID,
		Turn:      turn,
		Timestamp: s.timestamp(),
		Message:   message,
	})
}

func (s *Store) appendTrace(record TraceRecord) error {
	return s.appendJSONL(s.tracePath, record)
}

func (s *Store) appendJSONL(path string, record any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open jsonl file: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal record: %w", err)
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write jsonl record: %w", err)
	}

	return nil
}

func (s *Store) timestamp() string {
	return s.now().Format(time.RFC3339Nano)
}

func generateSessionID(now func() time.Time) string {
	suffix := make([]byte, 4)
	if _, err := rand.Read(suffix); err != nil {
		return fmt.Sprintf("%s-%d", now().Format("20060102-150405"), now().UnixNano())
	}

	return fmt.Sprintf("%s-%s", now().Format("20060102-150405"), hex.EncodeToString(suffix))
}
