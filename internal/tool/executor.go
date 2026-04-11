package tool

import (
	"context"
	"fmt"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type Result struct {
	ToolName string `json:"tool_name"`
	Success  bool   `json:"success"`
	Data     string `json:"data,omitempty"`
	Error    string `json:"error,omitempty"`
}

type Executor struct {
	tools     map[string]einotool.InvokableTool
	toolInfos []*schema.ToolInfo
}

func NewExecutor(ctx context.Context, availableTools []einotool.InvokableTool) (*Executor, error) {
	toolInfos := make([]*schema.ToolInfo, 0, len(availableTools))
	toolMap := make(map[string]einotool.InvokableTool, len(availableTools))

	for _, availableTool := range availableTools {
		info, err := availableTool.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("load tool info: %w", err)
		}

		toolInfos = append(toolInfos, info)
		toolMap[info.Name] = availableTool
	}

	return &Executor{
		tools:     toolMap,
		toolInfos: toolInfos,
	}, nil
}

func (e *Executor) ToolInfos() []*schema.ToolInfo {
	return e.toolInfos
}

func (e *Executor) Execute(ctx context.Context, toolName, arguments string) (Result, error) {
	invokableTool, ok := e.tools[toolName]
	if !ok {
		return Result{}, fmt.Errorf("tool %q not found", toolName)
	}

	output, err := invokableTool.InvokableRun(ctx, arguments)
	result := Result{
		ToolName: toolName,
		Success:  err == nil,
	}

	if err != nil {
		result.Error = err.Error()
		return result, nil
	}

	result.Data = output
	return result, nil
}
