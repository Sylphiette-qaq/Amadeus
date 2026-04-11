package utils

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// StreamResponse 流式处理对话响应
// 参数:
//
//	ctx: 上下文
//	agent: AI代理
//	query: 用户查询
func StreamResponse(ctx context.Context, ag *adk.ChatModelAgent, query string) {
	// 加载上下文
	history := LoadContext()
	// 保存用户问题
	SaveMessage(schema.User, query)

	// 初始化消息列表
	var agentInput *adk.AgentInput

	var messages []*schema.Message

	// 非首次对话直接使用历史上下文
	messages = append(history, schema.UserMessage(query))

	agentInput = &adk.AgentInput{
		messages,
		false,
	}

	// 运行对话
	msgReader := ag.Run(ctx, agentInput)

	// 初始化变量
	var lastMessage *schema.Message

	// 处理对话结果
	for {
		event, ok := msgReader.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			fmt.Printf("对话出错: %v\n", event.Err)
			break
		}

		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}

		msgOutput := event.Output.MessageOutput

		// 处理消息输出
		HandleMessageOutput(msgOutput, &lastMessage)
	}

	// 保存助手回复
	if lastMessage != nil {
		SaveMessage(lastMessage.Role, lastMessage.Content)
	}
}
