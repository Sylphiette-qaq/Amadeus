package main

import (
	"Amadeus/internal/memory"
	"Amadeus/internal/model"
	"Amadeus/internal/orchestrator"
	"Amadeus/internal/presentation"
	"Amadeus/internal/skill"
	internaltool "Amadeus/internal/tool"
	"context"
	"fmt"

	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()

	// 本地开发默认从 .env 读取密钥和运行参数，生产环境仍可直接依赖外部环境变量。
	_ = godotenv.Load()

	skillConfig, err := skill.LoadConfig()
	if err != nil {
		fmt.Println("初始化 skill 配置失败：", err)
		return
	}

	agentMarkdown, err := skill.LoadAgentMarkdown(skillConfig)
	if err != nil {
		fmt.Println("加载 agent.md 失败：", err)
		return
	}

	settings := model.ResolveChatModelSettings()
	store, err := memory.NewStore(memory.Config{
		Model:   settings.Model,
		BaseURL: settings.BaseURL,
	})
	if err != nil {
		fmt.Println("初始化会话存储失败：", err)
		return
	}

	chatModel := model.GetChatModel(ctx)
	availableTools, err := internaltool.LoadInvokableTools(ctx, "./tools/toolsConfig.json", skillConfig)
	if err != nil {
		fmt.Println("初始化工具失败：", err)
		return
	}

	executor, err := internaltool.NewExecutor(ctx, availableTools)
	if err != nil {
		fmt.Println("初始化工具执行器失败：", err)
		return
	}

	orch, err := orchestrator.New(chatModel, executor, store, model.BuildSystemMessage(agentMarkdown))
	if err != nil {
		fmt.Println("初始化编排器失败：", err)
		return
	}

	// CLI 层只负责读取输入和展示错误，不参与任何编排决策。
	for {
		userQuestion, err := presentation.ReadUserInput()
		if err != nil {
			fmt.Println("读取输入时发生错误：", err)
			return
		}
		if userQuestion == "" {
			fmt.Println("没有输入任何内容")
			continue
		}

		if err := orch.HandleTurn(ctx, userQuestion); err != nil {
			presentation.PrintTurnError(err)
		}
	}
}
