package tool

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/tool/mcp"
	einotool "github.com/cloudwego/eino/components/tool"
)

func LoadInvokableTools(ctx context.Context, configPath string) ([]einotool.InvokableTool, error) {
	availableTools := []einotool.InvokableTool{GetCalculatorTool()}

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
				return nil, fmt.Errorf("mcp tool %T is not invokable", baseTool)
			}

			availableTools = append(availableTools, invokableTool)
		}
	}

	return availableTools, nil
}
