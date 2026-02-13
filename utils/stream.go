package utils

import (
	"Amadeus/prompt"
	"context"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// StreamResponse 流式处理对话响应
// 参数:
//
//	ctx: 上下文
//	runner: 运行器
//	query: 用户查询
//	isFirst: 是否首次对话
func StreamResponse(ctx context.Context, runner *adk.Runner, query string, isFirst bool) {

	// 加载上下文
	history := LoadContext()
	// 保存用户问题
	SaveMessage(schema.User, query)

	// 初始化消息列表
	var messages []*schema.Message

	// 首次对话使用系统提示词模板
	if isFirst {
		messages, _ = prompt.GetMessage(ctx, history, query)
	} else {
		// 非首次对话直接使用历史上下文
		messages = append(history, schema.UserMessage(query))
	}

	// 运行对话
	iter := runner.Run(ctx, messages, adk.WithCheckPointID("default"))

	// 初始化变量
	var lastMessage *schema.Message
	var toolCallOutput strings.Builder

	// 处理对话结果
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		// 处理错误
		if event.Err != nil {
			return
		}

		// 处理消息输出
		if event.Output != nil && event.Output.MessageOutput != nil {
			msgOutput := event.Output.MessageOutput

			// 检查是否为工具输出（通过ToolName字段判断）
			if msgOutput.ToolName != "" {
				// 处理工具输出
				handleToolOutput(msgOutput, &toolCallOutput)
			} else {
				// 处理普通AI输出
				handleAIOutput(msgOutput, &lastMessage)
			}
		}
	}

	// 保存助手回复
	if lastMessage != nil {
		SaveMessage(lastMessage.Role, lastMessage.Content)
	}
}
