package main

import (
	"context"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
)

func main() {
	ctx := context.Background()

	// 初始化嵌入器
	embedder, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey:  "your-api-key",
		Model:   "text-embedding-ada-002",
		Timeout: 30 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	// 生成文本向量
	texts := []string{
		"这是第一段示例文本",
		"这是第二段示例文本",
	}

	embeddings, err := embedder.EmbedStrings(ctx, texts)
	if err != nil {
		panic(err)
	}

	// 使用生成的向量
	for i, embedding := range embeddings {
		println("文本", i+1, "的向量维度:", len(embedding))
	}
}
