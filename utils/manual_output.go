package utils

import (
	"fmt"

	"github.com/cloudwego/eino/schema"
)

func PrintToolCall(toolCall schema.ToolCall) {
	fmt.Printf("\n[工具调用] %s(%s)\n", toolCall.Function.Name, toolCall.Function.Arguments)
}

func PrintToolResult(toolName string, success bool, content string) {
	fmt.Printf("[工具结果] %s success=%t %s\n", toolName, success, content)
}

func PrintAssistantResponse(content string) {
	fmt.Printf("%s\n\n", content)
}
