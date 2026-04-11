package orchestrator

import (
	internaltool "Amadeus/internal/tool"
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
)

type Orchestrator struct {
	model      *deepseek.ChatModel
	executor   *internaltool.Executor
	maxTurns   int
	systemText string
}

func New(model *deepseek.ChatModel, executor *internaltool.Executor, systemText string) (*Orchestrator, error) {
	if err := model.BindTools(executor.ToolInfos()); err != nil {
		return nil, fmt.Errorf("bind tools: %w", err)
	}

	return &Orchestrator{
		model:      model,
		executor:   executor,
		maxTurns:   loadMaxTurns(),
		systemText: systemText,
	}, nil
}

func (o *Orchestrator) HandleTurn(ctx context.Context, userQuestion string) error {
	return o.handleTurn(ctx, userQuestion)
}
