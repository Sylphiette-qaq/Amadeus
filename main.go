package main

import (
	"Amadeus/agent"
	"Amadeus/orchestrator"
	"Amadeus/tools"
	"Amadeus/utils"
	"context"
	"fmt"

	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()

	_ = godotenv.Load()

	chatModel := agent.GetChatModel(ctx)
	availableTools, err := tools.LoadInvokableTools(ctx, "./tools/toolsConfig.json")
	if err != nil {
		fmt.Println("初始化工具失败：", err)
		return
	}

	orch, err := orchestrator.New(ctx, chatModel, availableTools)
	if err != nil {
		fmt.Println("初始化编排器失败：", err)
		return
	}

	for {
		userQuestion, err := utils.ReadUserInput()
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
