package tool

import (
	"Amadeus/internal/tool/basetools"
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/tool/mcp"
	einotool "github.com/cloudwego/eino/components/tool"
)

func LoadInvokableTools(ctx context.Context, configPath string) ([]einotool.InvokableTool, error) {
	availableTools := basetools.Load()

	// 先加载每次都应可用的基础工具，再把 MCP 工具拉平到同一列表。
	clients, err := CreateMcpClientsFromConfig(ctx, configPath)
	if err != nil {
		return nil, err
	}

	for _, cli := range clients {
		baseTools, err := mcp.GetTools(ctx, &mcp.Config{Cli: cli})
		if err != nil {
			return nil, fmt.Errorf("get mcp tools: %w", err)
		}

		for _, baseTool := range baseTools {
			invokableTool, ok := baseTool.(einotool.InvokableTool)
			if !ok {
				// 手动编排阶段必须能直接执行工具；不能执行的工具宁可提前失败，也不要静默忽略。
				return nil, fmt.Errorf("mcp tool %T is not invokable", baseTool)
			}

			availableTools = append(availableTools, invokableTool)
		}
	}

	return availableTools, nil
}
