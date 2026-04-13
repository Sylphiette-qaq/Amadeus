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
	// 执行器同时维护“运行时调用映射”和“模型可见的工具描述”，
	// 避免注册表和执行层各自重复扫描工具。
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

	// 工具自身报错时仍然返回结构化 Result，让编排器可以把失败信息作为 tool message 回填给模型。
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
