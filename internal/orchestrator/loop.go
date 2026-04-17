package orchestrator

import (
	"Amadeus/internal/presentation"
	"Amadeus/internal/session"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
)

func (o *Orchestrator) handleTurn(ctx context.Context, userQuestion string) error {
	history, err := o.store.LoadConversation()
	if err != nil {
		return fmt.Errorf("load conversation history: %w", err)
	}

	// 每次用户输入都从“系统消息 + 历史消息 + 当前问题”重建一次会话状态，
	// 这样后续接入摘要、裁剪或审计字段时有稳定的装配入口。
	state := session.NewState(history, o.systemText, userQuestion)
	if err := o.store.AppendUserMessage(0, schema.UserMessage(userQuestion)); err != nil {
		return fmt.Errorf("persist user message: %w", err)
	}
	presentation.Emit(presentation.Event{
		Type:    presentation.EventTurnStarted,
		Content: userQuestion,
	})
	finalMessage, err := o.run(ctx, state)
	if err != nil {
		if traceErr := o.store.AppendTurnError(state.CurrentTurn, err); traceErr != nil {
			return fmt.Errorf("persist turn error: %w", traceErr)
		}
		return err
	}

	if err := o.store.AppendAssistantFinal(state.CurrentTurn, finalMessage); err != nil {
		return fmt.Errorf("persist final assistant message: %w", err)
	}

	return nil
}

func (o *Orchestrator) run(ctx context.Context, state *session.State) (*schema.Message, error) {
	for turn := 1; turn <= o.maxTurns; turn++ {
		state.CurrentTurn = turn

		resp, err := o.streamModelTurn(ctx, state)
		if err != nil {
			return nil, fmt.Errorf("model stream failed on turn %d: %w", turn, err)
		}

		if len(resp.ToolCalls) == 0 {
			if strings.TrimSpace(resp.Content) == "" {
				return nil, fmt.Errorf("empty assistant response on turn %d", turn)
			}

			// 没有 tool_calls 且 content 非空时，视为本轮已经得到最终答复。
			state.Finished = true
			presentation.Emit(presentation.Event{Type: presentation.EventAssistantFinal})
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

func (o *Orchestrator) streamModelTurn(ctx context.Context, state *session.State) (*schema.Message, error) {
	if err := o.store.AppendTurnRequest(state.CurrentTurn, state.Messages); err != nil {
		return nil, fmt.Errorf("persist turn request: %w", err)
	}

	stream, err := o.model.Stream(ctx, state.Messages)
	if err != nil {
		return nil, err
	}

	var chunks []*schema.Message
	for {
		chunk, recvErr := stream.Recv()
		if recvErr == io.EOF {
			break
		}
		if recvErr != nil {
			return nil, recvErr
		}
		if chunk == nil {
			continue
		}

		chunks = append(chunks, chunk)
		if chunk.ReasoningContent == "" {
			if extracted, ok := deepseek.GetReasoningContent(chunk); ok {
				chunk.ReasoningContent = extracted
			}
		}
		if chunk.ReasoningContent != "" {
			presentation.Emit(presentation.Event{
				Type:    presentation.EventReasoningDelta,
				Content: chunk.ReasoningContent,
			})
		}
		if chunk.Content != "" {
			presentation.Emit(presentation.Event{
				Type:    presentation.EventAnswerDelta,
				Content: chunk.Content,
			})
		}
	}

	if len(chunks) == 0 {
		return nil, fmt.Errorf("empty stream response")
	}

	resp, err := schema.ConcatMessages(chunks)
	if err != nil {
		return nil, fmt.Errorf("concat stream messages: %w", err)
	}

	if resp.ReasoningContent == "" {
		if extracted, ok := deepseek.GetReasoningContent(resp); ok {
			resp.ReasoningContent = extracted
		}
	}
	if err := o.store.AppendModelResponse(state.CurrentTurn, resp); err != nil {
		return nil, fmt.Errorf("persist model response: %w", err)
	}

	return resp, nil
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

	toolMessage := schema.ToolMessage(toolContent, toolCallID, schema.WithToolName(toolCall.Function.Name))
	return toolMessage, nil
}
