package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	embeddingOpenAI "github.com/cloudwego/eino-ext/components/embedding/openai"
	milvusindexer "github.com/cloudwego/eino-ext/components/indexer/milvus"
	milvusretriever "github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	milvusclient "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const (
	defaultEmbeddingBaseURL = "https://open.bigmodel.cn/api/paas/v4"
	defaultEmbeddingModel   = "embedding-3"
	defaultMilvusAddress    = "localhost:19530"
	defaultMilvusCollection = "amadeus_memory"
	defaultEmbeddingDim     = 2048
	maxContentLength        = 8192
)

// sessionIDKey is the context key used to propagate the current session ID to tools.
type sessionIDKey struct{}

// WithSessionID returns a new context carrying the given session ID.
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey{}, sessionID)
}

// SessionIDFromContext retrieves the session ID injected by WithSessionID.
// Returns an empty string if no session ID is present.
func SessionIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(sessionIDKey{}).(string)
	return v
}

// IndexerConfig holds configuration for the RAG memory indexer.
type IndexerConfig struct {
	EmbeddingAPIKey  string
	EmbeddingBaseURL string
	EmbeddingModel   string
	EmbeddingDim     int
	MilvusAddress    string
	Collection       string
}

// LoadIndexerConfig reads RAG configuration from environment variables.
// Requires OPENAI_EMBEDDING_API_KEY to be explicitly set; if absent, RAG is disabled.
func LoadIndexerConfig() IndexerConfig {
	apiKey := os.Getenv("OPENAI_EMBEDDING_API_KEY")
	baseURL := os.Getenv("OPENAI_EMBEDDING_BASE_URL")
	if baseURL == "" {
		baseURL = defaultEmbeddingBaseURL
	}
	model := os.Getenv("OPENAI_EMBEDDING_MODEL")
	if model == "" {
		model = defaultEmbeddingModel
	}
	dim := defaultEmbeddingDim
	if v := os.Getenv("OPENAI_EMBEDDING_DIMENSIONS"); v != "" {
		if n, err := fmt.Sscanf(v, "%d", &dim); n != 1 || err != nil {
			dim = defaultEmbeddingDim
		}
	}
	addr := os.Getenv("MILVUS_ADDRESS")
	if addr == "" {
		addr = defaultMilvusAddress
	}
	col := os.Getenv("MILVUS_COLLECTION")
	if col == "" {
		col = defaultMilvusCollection
	}
	return IndexerConfig{
		EmbeddingAPIKey:  apiKey,
		EmbeddingBaseURL: baseURL,
		EmbeddingModel:   model,
		EmbeddingDim:     dim,
		MilvusAddress:    addr,
		Collection:       col,
	}
}

// Indexer provides RAG memory: indexes conversation turns into Milvus and retrieves
// semantically similar historical messages. When noop is true, all operations are silent no-ops.
type Indexer struct {
	noop      bool
	milvusIdx *milvusindexer.Indexer
	milvusRet *milvusretriever.Retriever
}

// memoryRow is the custom row struct for Milvus InsertRows, using FloatVector for OpenAI embeddings.
type memoryRow struct {
	ID       string    `milvus:"name:id"`
	Vector   []float32 `milvus:"name:vector"`
	Content  string    `milvus:"name:content"`
	Metadata []byte    `milvus:"name:metadata"`
}

// memoryFields defines the Milvus collection schema for conversation memory.
// Uses FloatVector (IP metric) instead of the default BinaryVector (HAMMING).
func memoryFields(dim int64) []*entity.Field {
	return []*entity.Field{
		entity.NewField().
			WithName("id").
			WithIsPrimaryKey(true).
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(255),
		entity.NewField().
			WithName("vector").
			WithDataType(entity.FieldTypeFloatVector).
			WithDim(dim),
		entity.NewField().
			WithName("content").
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(maxContentLength),
		entity.NewField().
			WithName("metadata").
			WithDataType(entity.FieldTypeJSON),
	}
}

// memoryDocumentConverter converts eino Documents with float64 vectors to Milvus row structs.
func memoryDocumentConverter(_ context.Context, docs []*schema.Document, vectors [][]float64) ([]interface{}, error) {
	rows := make([]interface{}, 0, len(docs))
	for i, doc := range docs {
		meta, err := json.Marshal(doc.MetaData)
		if err != nil {
			return nil, fmt.Errorf("marshal metadata: %w", err)
		}
		fv := make([]float32, len(vectors[i]))
		for j, v := range vectors[i] {
			fv[j] = float32(v)
		}
		content := doc.Content
		if len(content) > maxContentLength {
			content = content[:maxContentLength]
		}
		rows = append(rows, &memoryRow{
			ID:       doc.ID,
			Content:  content,
			Vector:   fv,
			Metadata: meta,
		})
	}
	return rows, nil
}

// memoryVectorConverter converts float64 query vectors to FloatVector entities for Milvus search.
func memoryVectorConverter(_ context.Context, vectors [][]float64) ([]entity.Vector, error) {
	result := make([]entity.Vector, len(vectors))
	for i, v := range vectors {
		fv := make(entity.FloatVector, len(v))
		for j, f := range v {
			fv[j] = float32(f)
		}
		result[i] = fv
	}
	return result, nil
}

// NewIndexer initializes the RAG memory indexer with Milvus and OpenAI Embedding.
// Returns a no-op Indexer (and nil error) if any dependency is unavailable.
func NewIndexer(ctx context.Context, cfg IndexerConfig) (*Indexer, error) {
	if cfg.EmbeddingAPIKey == "" {
		log.Printf("[memory.Indexer] OPENAI_EMBEDDING_API_KEY not set, RAG memory disabled")
		return &Indexer{noop: true}, nil
	}

	dim := cfg.EmbeddingDim
	if dim <= 0 {
		dim = defaultEmbeddingDim
	}
	dimPtr := dim

	embedder, err := embeddingOpenAI.NewEmbedder(ctx, &embeddingOpenAI.EmbeddingConfig{
		APIKey:     cfg.EmbeddingAPIKey,
		BaseURL:    cfg.EmbeddingBaseURL,
		Model:      cfg.EmbeddingModel,
		Dimensions: &dimPtr,
	})
	if err != nil {
		log.Printf("[memory.Indexer] failed to create embedder: %v, RAG memory disabled", err)
		return &Indexer{noop: true}, nil
	}

	milvusConn, err := milvusclient.NewGrpcClient(ctx, cfg.MilvusAddress)
	if err != nil {
		log.Printf("[memory.Indexer] failed to connect to Milvus at %s: %v, RAG memory disabled", cfg.MilvusAddress, err)
		return &Indexer{noop: true}, nil
	}

	idx, err := milvusindexer.NewIndexer(ctx, &milvusindexer.IndexerConfig{
		Client:            milvusConn,
		Collection:        cfg.Collection,
		Fields:            memoryFields(int64(dim)),
		MetricType:        milvusindexer.IP,
		DocumentConverter: memoryDocumentConverter,
		Embedding:         embedder,
	})
	if err != nil {
		log.Printf("[memory.Indexer] failed to init Milvus indexer: %v, RAG memory disabled", err)
		return &Indexer{noop: true}, nil
	}

	ret, err := milvusretriever.NewRetriever(ctx, &milvusretriever.RetrieverConfig{
		Client:          milvusConn,
		Collection:      cfg.Collection,
		VectorField:     "vector",
		OutputFields:    []string{"content", "metadata"},
		MetricType:      entity.IP,
		TopK:            5,
		VectorConverter: memoryVectorConverter,
		Embedding:       embedder,
	})
	if err != nil {
		log.Printf("[memory.Indexer] failed to init Milvus retriever: %v, RAG memory disabled", err)
		return &Indexer{noop: true}, nil
	}

	return &Indexer{
		noop:      false,
		milvusIdx: idx,
		milvusRet: ret,
	}, nil
}

// IsNoop returns true when the indexer is running in degraded mode (no Milvus connection).
func (ix *Indexer) IsNoop() bool {
	return ix.noop
}

// IndexMessages indexes a user message and an assistant message as separate vector records.
// Intended to be called in a goroutine after each conversation turn; logs errors silently.
func (ix *Indexer) IndexMessages(ctx context.Context, sessionID string, turn int, userMsg, assistantMsg *schema.Message) {
	if ix == nil || ix.noop {
		return
	}

	docs := []*schema.Document{
		{
			ID:      fmt.Sprintf("%s-t%d-user", sessionID, turn),
			Content: userMsg.Content,
			MetaData: map[string]any{
				"session_id": sessionID,
				"turn":       turn,
				"role":       "user",
			},
		},
		{
			ID:      fmt.Sprintf("%s-t%d-assistant", sessionID, turn),
			Content: assistantMsg.Content,
			MetaData: map[string]any{
				"session_id": sessionID,
				"turn":       turn,
				"role":       "assistant",
			},
		},
	}

	if _, err := ix.milvusIdx.Store(ctx, docs); err != nil {
		log.Printf("[memory.Indexer] failed to index turn %d for session %s: %v", turn, sessionID, err)
	}
}

// Search retrieves the top-k most semantically similar historical messages for the given query.
// If a session ID is present in ctx (via WithSessionID), results are scoped to that session only.
// Returns a formatted multi-line string, or an error if the service is unavailable.
func (ix *Indexer) Search(ctx context.Context, query string, topK int) (string, error) {
	if ix == nil || ix.noop {
		return "", fmt.Errorf("记忆服务不可用")
	}

	opts := []retriever.Option{retriever.WithTopK(topK)}
	if sessionID := SessionIDFromContext(ctx); sessionID != "" {
		// Filter to the current conversation only using Milvus JSON field syntax.
		opts = append(opts, milvusretriever.WithFilter(fmt.Sprintf(`metadata["session_id"] == "%s"`, sessionID)))
	}

	docs, err := ix.milvusRet.Retrieve(ctx, query, opts...)
	if err != nil {
		return "", fmt.Errorf("检索失败: %w", err)
	}

	if len(docs) == 0 {
		return "未找到相关历史记录", nil
	}

	var sb strings.Builder
	for _, doc := range docs {
		sessionID, _ := doc.MetaData["session_id"].(string)
		role, _ := doc.MetaData["role"].(string)
		// JSON numbers deserialize as float64 in Go
		turnFloat, _ := doc.MetaData["turn"].(float64)
		turn := int(turnFloat)
		sb.WriteString(fmt.Sprintf("[Session: %s, Turn %d, %s]\n%s\n\n", sessionID, turn, role, doc.Content))
	}
	return strings.TrimSpace(sb.String()), nil
}
