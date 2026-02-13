package agent

import (
	"Amadeus/tools"
	"context"
	"log"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
)

var (
	ModelType     = "deepseek-chat"
	ownerAPIKey   = "sk-ab95814d25f54a02aaee43f062926e2c"
	modelURL      = "https://api.deepseek.com"
	SystemMessage = `你是一个人工智能助手，名称是Amadeus。你需要用语气平淡，内容简洁且专业的语气回答问题。`
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

func GetAgent(ctx context.Context) *react.Agent {
	chatModel := GetChatModel(ctx)

	// 从配置文件创建MCP客户端
	clients, err := tools.CreateMcpClientsFromConfig(ctx, "./tools/toolsConfig.json")
	if err != nil {
		log.Fatalf("创建MCP客户端失败: %v", err)
	}

	// 从所有MCP客户端获取工具
	var allTools []tool.BaseTool
	for _, cli := range clients {
		tools, err := mcp.GetTools(ctx, &mcp.Config{Cli: cli})
		if err != nil {
			log.Printf("获取工具失败: %v", err)
			continue
		}
		allTools = append(allTools, tools...)
	}

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ModelNodeName:    "Amadeus",
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: allTools,
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	return agent
}
