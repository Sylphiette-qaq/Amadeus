package main

import (
	"Amadeus/agent"
	"Amadeus/utils"
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()

	// 创建 Agent
	agent := agent.GetAgent(ctx)

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

		utils.StreamResponse(ctx, agent, userQuestion)
	}
}
