package model

import (
	"context"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
)

const (
	defaultModelType = "deepseek-chat"
	defaultModelURL  = "https://api.deepseek.com"
)

var SystemMessage = `你是一个人工智能助手，名称是Amadeus。你需要用语气平淡，内容简洁且专业的语气回答问题。`

func GetChatModel(ctx context.Context) *deepseek.ChatModel {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("DEEPSEEK_API_KEY is required")
	}

	// model 层只负责创建底层 ChatModel，不绑定工具也不承接业务编排。
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  apiKey,
		Model:   getenvDefault("DEEPSEEK_MODEL", defaultModelType),
		BaseURL: getenvDefault("DEEPSEEK_BASE_URL", defaultModelURL),
	})
	if err != nil {
		log.Fatal(err)
	}

	return chatModel
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
