package orchestrator

import (
	"Amadeus/agent"
	"Amadeus/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const defaultMaxTurns = 8

type Orchestrator struct {
	model      *deepseek.ChatModel
	tools      map[string]einotool.InvokableTool
	maxTurns   int
	systemText string
}

type toolResultPayload struct {
	Success bool   `json:"success"`
	Data    string `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func New(ctx context.Context, model *deepseek.ChatModel, availableTools []einotool.InvokableTool) (*Orchestrator, error) {
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

	if err := model.BindTools(toolInfos); err != nil {
		return nil, fmt.Errorf("bind tools: %w", err)
	}

	return &Orchestrator{
		model:      model,
		tools:      toolMap,
		maxTurns:   loadMaxTurns(),
		systemText: agent.SystemMessage,
	}, nil
}

func (o *Orchestrator) HandleTurn(ctx context.Context, userQuestion string) error {
	history := utils.LoadContext()
	utils.SaveMessage(schema.User, userQuestion)

	messages := make([]*schema.Message, 0, len(history)+2)
	messages = append(messages, schema.SystemMessage(o.systemText))
	messages = append(messages, history...)
	messages = append(messages, schema.UserMessage(userQuestion))

	finalMessage, err := o.run(ctx, messages)
	if err != nil {
		return err
	}

	utils.PrintAssistantResponse(finalMessage.Content)
	utils.SaveMessage(finalMessage.Role, finalMessage.Content)

	return nil
}

func (o *Orchestrator) run(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	for turn := 1; turn <= o.maxTurns; turn++ {
		resp, err := o.model.Generate(ctx, messages)
		if err != nil {
			return nil, fmt.Errorf("model generate failed on turn %d: %w", turn, err)
		}

		if len(resp.ToolCalls) == 0 {
			if strings.TrimSpace(resp.Content) == "" {
				return nil, fmt.Errorf("empty assistant response on turn %d", turn)
			}

			return resp, nil
		}

		messages = append(messages, resp)

		for _, toolCall := range resp.ToolCalls {
			utils.PrintToolCall(toolCall)

			toolMessage, toolErr := o.executeTool(ctx, toolCall)
			if toolErr != nil {
				return nil, toolErr
			}

			messages = append(messages, toolMessage)
		}
	}

	return nil, fmt.Errorf("max turns exceeded: %d", o.maxTurns)
}

func (o *Orchestrator) executeTool(ctx context.Context, toolCall schema.ToolCall) (*schema.Message, error) {
	invokableTool, ok := o.tools[toolCall.Function.Name]
	if !ok {
		return nil, fmt.Errorf("tool %q not found", toolCall.Function.Name)
	}

	output, err := invokableTool.InvokableRun(ctx, toolCall.Function.Arguments)
	payload := toolResultPayload{Success: err == nil}
	if err != nil {
		payload.Error = err.Error()
	} else {
		payload.Data = output
	}

	toolContentBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil, fmt.Errorf("marshal tool result for %q: %w", toolCall.Function.Name, marshalErr)
	}

	utils.PrintToolResult(toolCall.Function.Name, payload.Success, string(toolContentBytes))

	toolCallID := toolCall.ID
	if toolCallID == "" {
		toolCallID = toolCall.Function.Name
	}

	return schema.ToolMessage(string(toolContentBytes), toolCallID, schema.WithToolName(toolCall.Function.Name)), nil
}

func loadMaxTurns() int {
	raw := os.Getenv("AMADEUS_MAX_TURNS")
	if raw == "" {
		return defaultMaxTurns
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return defaultMaxTurns
	}

	return value
}
