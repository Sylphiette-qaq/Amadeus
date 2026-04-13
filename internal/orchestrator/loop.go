package orchestrator

import (
	"Amadeus/internal/memory"
	"Amadeus/internal/presentation"
	"Amadeus/internal/session"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
)

func (o *Orchestrator) handleTurn(ctx context.Context, userQuestion string) error {
	history := memory.LoadContext()
	memory.SaveMessage(schema.User, userQuestion)

	// 每次用户输入都从“系统消息 + 历史消息 + 当前问题”重建一次会话状态，
	// 这样后续接入摘要、裁剪或审计字段时有稳定的装配入口。
	state := session.NewState(history, o.systemText, userQuestion)
	finalMessage, err := o.run(ctx, state)
	if err != nil {
		return err
	}

	presentation.PrintAssistantResponse(finalMessage.Content)
	memory.SaveMessage(finalMessage.Role, finalMessage.Content)
	return nil
}

func (o *Orchestrator) run(ctx context.Context, state *session.State) (*schema.Message, error) {
	for turn := 1; turn <= o.maxTurns; turn++ {
		state.CurrentTurn = turn

		resp, err := o.model.Generate(ctx, state.Messages)
		if err != nil {
			return nil, fmt.Errorf("model generate failed on turn %d: %w", turn, err)
		}

		if len(resp.ToolCalls) == 0 {
			if strings.TrimSpace(resp.Content) == "" {
				return nil, fmt.Errorf("empty assistant response on turn %d", turn)
			}

			// 没有 tool_calls 且 content 非空时，视为本轮已经得到最终答复。
			state.Finished = true
			return resp, nil
		}

		// assistant 的工具调用消息必须先入历史，再把对应的 tool 消息逐条回填。
		state.Append(resp)

		for _, toolCall := range resp.ToolCalls {
			presentation.PrintToolCall(toolCall)

			toolMessage, toolErr := o.executeTool(ctx, toolCall, state)
			if toolErr != nil {
				return nil, toolErr
			}

			state.Append(toolMessage)
			state.ToolCallCount++
		}
	}

	return nil, fmt.Errorf("max turns exceeded: %d", o.maxTurns)
}

func (o *Orchestrator) executeTool(ctx context.Context, toolCall schema.ToolCall, state *session.State) (*schema.Message, error) {
	// 先校验 arguments 至少是合法 JSON，避免把明显坏输入直接交给工具执行层。
	if err := ParseToolArguments(toolCall.Function.Arguments); err != nil {
		return nil, fmt.Errorf("invalid tool arguments for %q: %w", toolCall.Function.Name, err)
	}

	result, err := o.executor.Execute(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
	if err != nil {
		return nil, err
	}

	toolContentBytes, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return nil, fmt.Errorf("marshal tool result for %q: %w", toolCall.Function.Name, marshalErr)
	}

	// 统一将工具结果包装为 JSON 字符串，便于下一轮模型稳定消费，也为后续结构化存储留接口。
	toolContent := string(toolContentBytes)
	state.LastToolResult = toolContent
	presentation.PrintToolResult(toolCall.Function.Name, result.Success, toolContent)

	toolCallID := toolCall.ID
	if toolCallID == "" {
		// 某些模型/实现可能不稳定返回 tool_call_id，这里保底回退到工具名，避免消息丢关联。
		toolCallID = toolCall.Function.Name
	}

	return schema.ToolMessage(toolContent, toolCallID, schema.WithToolName(toolCall.Function.Name)), nil
}
