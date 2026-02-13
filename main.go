package main

import (
	"Amadeus/agent"
	"Amadeus/utils"
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
)

func main() {
	ctx := context.Background()

	// 创建 Agent 和 Runner
	agent := agent.GetAgent(ctx)
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: agent,
	})

	isFirst := true

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

		utils.StreamResponse(ctx, runner, userQuestion, isFirst)
		isFirst = false
	}
}
