package orchestrator

import (
	"Amadeus/internal/memory"
	internaltool "Amadeus/internal/tool"
	"context"
	"fmt"

	model "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type chatModel interface {
	BindTools([]*schema.ToolInfo) error
	Generate(context.Context, []*schema.Message, ...model.Option) (*schema.Message, error)
	Stream(context.Context, []*schema.Message, ...model.Option) (*schema.StreamReader[*schema.Message], error)
}

type toolExecutor interface {
	ToolInfos() []*schema.ToolInfo
	Execute(context.Context, string, string) (internaltool.Result, error)
}

type Orchestrator struct {
	model      chatModel
	executor   toolExecutor
	store      *memory.Store
	maxTurns   int
	systemText string
	stream     bool
}

func New(model chatModel, executor *internaltool.Executor, store *memory.Store, systemText string, stream bool) (*Orchestrator, error) {
	if store == nil {
		return nil, fmt.Errorf("memory store is required")
	}

	// 工具在编排器初始化时统一绑定到模型，避免请求过程中反复构建工具描述。
	if err := model.BindTools(executor.ToolInfos()); err != nil {
		return nil, fmt.Errorf("bind tools: %w", err)
	}

	return &Orchestrator{
		model:      model,
		executor:   executor,
		store:      store,
		maxTurns:   loadMaxTurns(),
		systemText: systemText,
		stream:     stream,
	}, nil
}

func (o *Orchestrator) HandleTurn(ctx context.Context, userQuestion string) error {
	return o.handleTurn(ctx, userQuestion)
}
