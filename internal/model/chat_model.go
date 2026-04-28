package model

import (
	"context"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
)

var SystemMessage = `你是一个人工智能助手，名称是Amadeus。你需要用语气平淡，内容简洁且专业的语气回答问题。`

func BuildSystemMessage(agentMarkdown string) string {
	if agentMarkdown == "" {
		return SystemMessage
	}

	return SystemMessage + `

以下是当前可用的 skills 列表。这里只提供名称和简介；当你确认某个 skill 适用时，再调用 load_skill 加载该 skill 的完整说明。

` + agentMarkdown
}

func GetChatModel(ctx context.Context) *openai.ChatModel {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("DEEPSEEK_API_KEY is required")
	}

	settings := ResolveChatModelSettings()

	// model 层只负责创建底层 ChatModel，不绑定工具也不承接业务编排。
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:          apiKey,
		Model:           settings.Model,
		BaseURL:         settings.BaseURL,
		ReasoningEffort: openai.ReasoningEffortLevel(settings.ReasoningEffort),
		ExtraFields:     buildExtraFields(settings),
	})
	if err != nil {
		log.Fatal(err)
	}

	return chatModel
}
