package main

import (
	"Amadeus/internal/memory"
	"Amadeus/internal/model"
	"Amadeus/internal/orchestrator"
	"Amadeus/internal/skill"
	internaltool "Amadeus/internal/tool"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// convSession 持有单个对话的 orchestrator 和互斥锁，保证同一会话串行处理。
type convSession struct {
	mu   sync.Mutex
	orch *orchestrator.Orchestrator
}

type sessionManager struct {
	sessions sync.Map // map[conversationID string] -> *convSession

	// 共享资源：executor 和 system prompt 在所有会话间共享，避免重复启动 MCP 进程。
	ctx        context.Context
	systemText string
	executor   *internaltool.Executor
	settings   model.ChatModelSettings
}

func (sm *sessionManager) getOrCreate(conversationID string) (*convSession, error) {
	if v, ok := sm.sessions.Load(conversationID); ok {
		return v.(*convSession), nil
	}

	store, err := memory.NewStore(memory.Config{
		SessionID: sessionIDFromConversationID(conversationID),
		Model:     sm.settings.Model,
		BaseURL:   sm.settings.BaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create store for %q: %w", conversationID, err)
	}

	// 每个会话使用独立的 chatModel 实例（BindTools 会修改 model 状态）。
	chatModel := model.GetChatModel(sm.ctx)
	orch, err := orchestrator.New(chatModel, sm.executor, store, sm.systemText, sm.settings.Stream)
	if err != nil {
		return nil, fmt.Errorf("create orchestrator for %q: %w", conversationID, err)
	}

	sess := &convSession{orch: orch}
	actual, _ := sm.sessions.LoadOrStore(conversationID, sess)
	return actual.(*convSession), nil
}

func sessionIDFromConversationID(conversationID string) string {
	trimmed := strings.TrimSpace(conversationID)
	hash := sha256.Sum256([]byte(trimmed))
	hashSuffix := hex.EncodeToString(hash[:])[:12]

	var safe strings.Builder
	for _, r := range trimmed {
		switch {
		case r >= 'a' && r <= 'z':
			safe.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			safe.WriteRune(r)
		case r >= '0' && r <= '9':
			safe.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			safe.WriteRune(r)
		default:
			safe.WriteByte('_')
		}
	}

	safeName := strings.Trim(safe.String(), "._")
	if safeName == "" {
		safeName = "conversation"
	}
	if len(safeName) > 80 {
		safeName = safeName[:80]
	}

	return fmt.Sprintf("conversation-%s-%s", safeName, hashSuffix)
}

type chatRequest struct {
	ConversationID string `json:"conversation_id"`
	Message        string `json:"message"`
}

type chatResponse struct {
	Reply string `json:"reply"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	_ = godotenv.Load()

	ctx := context.Background()

	skillConfig, err := skill.LoadConfig()
	if err != nil {
		log.Fatal("初始化 skill 配置失败：", err)
	}
	agentMarkdown, err := skill.LoadAgentMarkdown(skillConfig)
	if err != nil {
		log.Fatal("加载 agent.md 失败：", err)
	}

	settings := model.ResolveChatModelSettings()

	initCtx, initCancel := context.WithTimeout(ctx, 30*time.Second)
	availableTools, err := internaltool.LoadInvokableTools(initCtx, "./tools/toolsConfig.json", skillConfig)
	initCancel()
	if err != nil {
		log.Fatal("初始化工具失败：", err)
	}
	executor, err := internaltool.NewExecutor(ctx, availableTools)
	if err != nil {
		log.Fatal("初始化工具执行器失败：", err)
	}

	sm := &sessionManager{
		ctx:        ctx,
		systemText: model.BuildSystemMessage(agentMarkdown),
		executor:   executor,
		settings:   settings,
	}

	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
			return
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON"})
			return
		}
		if strings.TrimSpace(req.ConversationID) == "" || strings.TrimSpace(req.Message) == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "conversation_id and message are required"})
			return
		}

		sess, err := sm.getOrCreate(req.ConversationID)
		if err != nil {
			log.Printf("session init error for %q: %v", req.ConversationID, err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "session init failed"})
			return
		}

		sess.mu.Lock()
		log.Printf("[%s] 开始处理: %q", req.ConversationID, req.Message)
		reply, err := sess.orch.HandleTurnWithResponse(r.Context(), req.Message)
		sess.mu.Unlock()

		if err != nil {
			log.Printf("[%s] turn error: %v", req.ConversationID, err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
			return
		}

		log.Printf("[%s] 回复完成: %q", req.ConversationID, reply)
		writeJSON(w, http.StatusOK, chatResponse{Reply: reply})
	})

	addr := os.Getenv("AGENT_SERVER_ADDR")
	if addr == "" {
		addr = ":9000"
	}
	log.Printf("Agent Server 启动，监听 %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
