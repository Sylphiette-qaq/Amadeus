package model

import (
	"context"
	"encoding/json"

	openaiacl "github.com/cloudwego/eino-ext/libs/acl/openai"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func ReasoningPayloadOption() einomodel.Option {
	return openaiacl.WithRequestPayloadModifier(injectReasoningContentPayload)
}

func injectReasoningContentPayload(_ context.Context, messages []*schema.Message, rawBody []byte) ([]byte, error) {
	var payload map[string]any
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return nil, err
	}

	rawMessages, ok := payload["messages"].([]any)
	if !ok {
		return rawBody, nil
	}

	for i, message := range messages {
		if i >= len(rawMessages) {
			break
		}
		if message == nil || message.Role != schema.Assistant || message.ReasoningContent == "" {
			continue
		}

		rawMessage, ok := rawMessages[i].(map[string]any)
		if !ok {
			continue
		}
		rawMessage["reasoning_content"] = message.ReasoningContent
	}

	return json.Marshal(payload)
}
