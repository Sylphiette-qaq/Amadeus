package memory_test

import (
	"context"
	"testing"

	"Amadeus/internal/memory"

	"github.com/cloudwego/eino/schema"
)

func TestIndexer_NoopWhenAPIKeyMissing(t *testing.T) {
	ctx := context.Background()
	cfg := memory.IndexerConfig{
		EmbeddingAPIKey: "", // no key → noop
		MilvusAddress:   "localhost:19530",
		Collection:      "test_collection",
	}

	idx, err := memory.NewIndexer(ctx, cfg)
	if err != nil {
		t.Fatalf("NewIndexer() unexpected error: %v", err)
	}
	if !idx.IsNoop() {
		t.Fatal("expected noop indexer when APIKey is empty")
	}
}

func TestIndexer_IndexMessages_NoopSilent(t *testing.T) {
	ctx := context.Background()
	idx, _ := memory.NewIndexer(ctx, memory.IndexerConfig{EmbeddingAPIKey: ""})

	// Should not panic and return silently.
	userMsg := schema.UserMessage("hello")
	assistantMsg := schema.AssistantMessage("hi there", nil)
	idx.IndexMessages(ctx, "session-001", 1, userMsg, assistantMsg)
}

func TestIndexer_IndexMessages_NilReceiver(t *testing.T) {
	ctx := context.Background()
	var idx *memory.Indexer // nil receiver

	userMsg := schema.UserMessage("hello")
	assistantMsg := schema.AssistantMessage("hi there", nil)
	// Must not panic.
	idx.IndexMessages(ctx, "session-001", 1, userMsg, assistantMsg)
}

func TestIndexer_Search_NoopReturnsError(t *testing.T) {
	ctx := context.Background()
	idx, _ := memory.NewIndexer(ctx, memory.IndexerConfig{EmbeddingAPIKey: ""})

	_, err := idx.Search(ctx, "some query", 5)
	if err == nil {
		t.Fatal("expected error from noop Search, got nil")
	}
}

func TestIndexer_Search_NilReceiverReturnsError(t *testing.T) {
	ctx := context.Background()
	var idx *memory.Indexer

	_, err := idx.Search(ctx, "some query", 5)
	if err == nil {
		t.Fatal("expected error from nil receiver Search, got nil")
	}
}

func TestLoadIndexerConfig_Defaults(t *testing.T) {
	// Unset all env vars to test defaults.
	t.Setenv("OPENAI_EMBEDDING_API_KEY", "")
	t.Setenv("OPENAI_EMBEDDING_BASE_URL", "")
	t.Setenv("OPENAI_EMBEDDING_MODEL", "")
	t.Setenv("MILVUS_ADDRESS", "")
	t.Setenv("MILVUS_COLLECTION", "")

	cfg := memory.LoadIndexerConfig()

	if cfg.EmbeddingBaseURL != "https://api.openai.com/v1" {
		t.Errorf("EmbeddingBaseURL = %q, want default", cfg.EmbeddingBaseURL)
	}
	if cfg.EmbeddingModel != "text-embedding-3-small" {
		t.Errorf("EmbeddingModel = %q, want default", cfg.EmbeddingModel)
	}
	if cfg.MilvusAddress != "localhost:19530" {
		t.Errorf("MilvusAddress = %q, want default", cfg.MilvusAddress)
	}
	if cfg.Collection != "amadeus_memory" {
		t.Errorf("Collection = %q, want default", cfg.Collection)
	}
}

func TestLoadIndexerConfig_EnvOverride(t *testing.T) {
	t.Setenv("OPENAI_EMBEDDING_API_KEY", "sk-test-key")
	t.Setenv("OPENAI_EMBEDDING_BASE_URL", "https://custom.api/v1")
	t.Setenv("OPENAI_EMBEDDING_MODEL", "text-embedding-ada-002")
	t.Setenv("MILVUS_ADDRESS", "192.168.1.100:19530")
	t.Setenv("MILVUS_COLLECTION", "my_collection")

	cfg := memory.LoadIndexerConfig()

	if cfg.EmbeddingAPIKey != "sk-test-key" {
		t.Errorf("EmbeddingAPIKey = %q, want overridden value", cfg.EmbeddingAPIKey)
	}
	if cfg.EmbeddingBaseURL != "https://custom.api/v1" {
		t.Errorf("EmbeddingBaseURL = %q, want overridden value", cfg.EmbeddingBaseURL)
	}
	if cfg.MilvusAddress != "192.168.1.100:19530" {
		t.Errorf("MilvusAddress = %q, want overridden value", cfg.MilvusAddress)
	}
	if cfg.Collection != "my_collection" {
		t.Errorf("Collection = %q, want overridden value", cfg.Collection)
	}
}
