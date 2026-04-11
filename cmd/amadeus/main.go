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

	for {
		userQuestion, err := presentation.ReadUserInput()
		if err != nil {
			fmt.Println("读取输入时发生错误：", err)
			return
		}
		if userQuestion == "" {
			fmt.Println("没有输入任何内容")
			return
		}

		if err := orch.HandleTurn(ctx, userQuestion); err != nil {
			fmt.Println("处理请求失败：", err)
		}
	}
}
