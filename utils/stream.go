package utils

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

// StreamResponse 流式处理对话响应
// 参数:
//
//	ctx: 上下文
//	agent: React Agent
//	query: 用户查询
func StreamResponse(ctx context.Context, agent *react.Agent, query string) {

	// 加载上下文
	history := LoadContext()
	// 保存用户问题
	SaveMessage(schema.User, query)

	// 初始化消息列表
	var messages []*schema.Message

	// 非首次对话直接使用历史上下文
	messages = append(history, schema.UserMessage(query))

	// 运行对话 - 使用Stream方法
	stream, err := agent.Stream(ctx, messages)
	if err != nil {
		fmt.Printf("运行Agent失败: %v\n", err)
		return
	}

	// 读取流式输出
	var result *schema.Message
	for {
		msg, err := stream.Recv()
		if err != nil {
			break
		}
		if msg != nil {
			result = msg
			fmt.Print(msg.Content)
		}
	}

	fmt.Println()

	// 保存助手回复
	if result != nil {
		SaveMessage(schema.Assistant, result.Content)
	}
}
