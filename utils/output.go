package utils

import (
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// OutputType 输出类型枚举
type OutputType int

const (
	// OutputTypeNormal 普通输出
	OutputTypeNormal OutputType = iota
	// OutputTypeToolCall 工具调用
	OutputTypeToolCall
	// OutputTypeToolResult 工具结果
	OutputTypeToolResult
	// OutputTypeThinking 思考过程
	OutputTypeThinking
)

// FormattedOutput 格式化输出结构
type FormattedOutput struct {
	Type    OutputType
	Content string
}

// FormatOutput 格式化输出内容
// 参数:
//
//	output: 格式化输出结构
//
// 返回:
//
//	格式化后的字符串
func FormatOutput(output FormattedOutput) string {
	switch output.Type {
	case OutputTypeToolCall:
		return fmt.Sprintf("\n[工具调用] %s\n", output.Content)
	case OutputTypeToolResult:
		return fmt.Sprintf("[工具结果] %s\n", output.Content)
	case OutputTypeThinking:
		return fmt.Sprintf("[思考中] %s\n", output.Content)
	case OutputTypeNormal:
		return output.Content + "\n\n"
	default:
		return output.Content
	}
}

// handleToolOutput 处理工具输出
// 参数:
//
//	msgOutput: 消息输出结构
//	output: 输出缓冲区
func handleToolOutput(msgOutput *adk.MessageVariant, output *strings.Builder) {
	// 处理流式输出
	if msgOutput.IsStreaming && msgOutput.MessageStream != nil {
		for {
			msg, err := msgOutput.MessageStream.Recv()
			if err != nil {
				break
			}
			output.WriteString(msg.Content)
		}
	} else if msgOutput.Message != nil {
		// 处理非流式输出
		output.WriteString(msgOutput.Message.Content)
	}

	// 格式化并输出工具结果
	if output.Len() > 0 {
		result := FormatOutput(FormattedOutput{
			Type:    OutputTypeToolResult,
			Content: output.String(),
		})
		fmt.Print(result)
	}
}

// handleAIOutput 处理AI输出
// 参数:
//
//	msgOutput: 消息输出结构
//	lastMessage: 最后一条消息指针
func handleAIOutput(msgOutput *adk.MessageVariant, lastMessage **schema.Message) {
	// 处理流式输出
	if msgOutput.IsStreaming && msgOutput.MessageStream != nil {
		for {
			msg, err := msgOutput.MessageStream.Recv()
			if err != nil {
				break
			}

			// 检查是否有工具调用
			if len(msg.ToolCalls) > 0 {
				// 输出工具调用信息
				printToolCalls(msg.ToolCalls)
			} else {
				// 普通文本输出
				fmt.Print(msg.Content)
			}

			*lastMessage = msg
		}
		// 流式输出结束后添加换行
		fmt.Println()
	} else if msgOutput.Message != nil {
		// 处理非流式输出
		msg := msgOutput.Message

		// 检查是否有工具调用
		if len(msg.ToolCalls) > 0 {
			printToolCalls(msg.ToolCalls)
		} else {
			// 普通文本输出
			output := FormatOutput(FormattedOutput{
				Type:    OutputTypeNormal,
				Content: msg.Content,
			})
			fmt.Print(output)
		}

		*lastMessage = msg
	}
}

// printToolCalls 输出工具调用信息
// 参数:
//
//	toolCalls: 工具调用列表
func printToolCalls(toolCalls []schema.ToolCall) {
	for _, tc := range toolCalls {
		output := FormatOutput(FormattedOutput{
			Type:    OutputTypeToolCall,
			Content: fmt.Sprintf("%s(%s)", tc.Function.Name, tc.Function.Arguments),
		})
		fmt.Print(output)
	}
}

// HandleMessageOutput 处理消息输出
// 参数:
//
//	msgOutput: 消息输出结构
//	lastMessage: 最后一条消息指针
func HandleMessageOutput(msgOutput *adk.MessageVariant, lastMessage **schema.Message) {
	// 初始化工具输出缓冲区
	var toolCallOutput strings.Builder

	// 检查是否为工具输出（通过ToolName字段判断）
	if msgOutput.ToolName != "" {
		// 处理工具输出
		handleToolOutput(msgOutput, &toolCallOutput)
	} else {
		// 处理普通AI输出
		handleAIOutput(msgOutput, lastMessage)
	}
}
