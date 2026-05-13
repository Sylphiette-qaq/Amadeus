package basetools_test

import (
	"context"
	"testing"

	"Amadeus/internal/memory"
	"Amadeus/internal/tool/basetools"
)

func noopIndexer(t *testing.T) *memory.Indexer {
	t.Helper()
	idx, err := memory.NewIndexer(context.Background(), memory.IndexerConfig{EmbeddingAPIKey: ""})
	if err != nil {
		t.Fatalf("NewIndexer: %v", err)
	}
	return idx
}

func TestGetSearchMemoryTool_ToolInfo(t *testing.T) {
	ctx := context.Background()
	tool := basetools.GetSearchMemoryTool(noopIndexer(t))

	info, err := tool.Info(ctx)
	if err != nil {
		t.Fatalf("tool.Info() error: %v", err)
	}
	if info.Name != "search_memory" {
		t.Errorf("tool name = %q, want search_memory", info.Name)
	}
}

func TestGetSearchMemoryTool_NoopReturnsError(t *testing.T) {
	ctx := context.Background()
	tool := basetools.GetSearchMemoryTool(noopIndexer(t))

	// With noop indexer, Search returns an error; InvokableRun should propagate it.
	_, err := tool.InvokableRun(ctx, `{"query":"pgvector"}`)
	if err == nil {
		t.Fatal("expected error from noop search, got nil")
	}
}

func TestGetSearchMemoryTool_MissingQuery(t *testing.T) {
	ctx := context.Background()
	tool := basetools.GetSearchMemoryTool(noopIndexer(t))

	_, err := tool.InvokableRun(ctx, `{}`)
	if err == nil {
		t.Fatal("expected error for missing required param 'query'")
	}
}

func TestGetSearchMemoryTool_NilIndexer(t *testing.T) {
	ctx := context.Background()
	tool := basetools.GetSearchMemoryTool(nil)

	// Nil indexer treated same as noop; should return an error.
	_, err := tool.InvokableRun(ctx, `{"query":"test"}`)
	if err == nil {
		t.Fatal("expected error from nil indexer search, got nil")
	}
}
