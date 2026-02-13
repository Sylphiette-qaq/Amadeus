package agent

import (
	"Amadeus/tools"
	"context"
	"log"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

var (
	ModelType                    = "deepseek-chat"
	ownerAPIKey                  = "sk-ab95814d25f54a02aaee43f062926e2c"
	modelURL                     = "https://api.deepseek.com"
	SystemMessageDefaultTemplate = `你是一个{role}。你需要用{style}的语气回答问题。`
	UserMessageDefaultTemplate   = `{question}`
)

func GetChatModel(ctx context.Context) *deepseek.ChatModel {
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  ownerAPIKey,
		Model:   ModelType,
		BaseURL: modelURL,
	})

	if err != nil {
		log.Fatal(err)
	}
	return chatModel
}

func GetAgent(ctx context.Context) *adk.ChatModelAgent {
	chatModel := GetChatModel(ctx)

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "Amadeus",
		Description: "一个人工智能助手，名称叫Amadeus",
		Instruction: SystemMessageDefaultTemplate,
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{
					tools.GetCalculatorTool(),
				},
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	return agent
}
