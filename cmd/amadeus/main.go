package main

import (
	"Amadeus/internal/model"
	"Amadeus/internal/orchestrator"
	"Amadeus/internal/presentation"
	internaltool "Amadeus/internal/tool"
	"context"
	"fmt"

	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()

	// 本地开发默认从 .env 读取密钥和运行参数，生产环境仍可直接依赖外部环境变量。
	_ = godotenv.Load()

	chatModel := model.GetChatModel(ctx)
	availableTools, err := internaltool.LoadInvokableTools(ctx, "./tools/toolsConfig.json")
	if err != nil {
		fmt.Println("初始化工具失败：", err)
		return
	}

	executor, err := internaltool.NewExecutor(ctx, availableTools)
	if err != nil {
		fmt.Println("初始化工具执行器失败：", err)
		return
	}

	orch, err := orchestrator.New(chatModel, executor, model.SystemMessage)
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
			fmt.Println("处理请求失败：", err)
		}
	}
}
