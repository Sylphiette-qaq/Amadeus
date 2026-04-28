package model

import (
	"log"
	"os"
	"strconv"
)

const (
	defaultModelType       = "deepseek-v4-flash"
	defaultModelURL        = "https://api.deepseek.com"
	defaultThinkingType    = "enabled"
	defaultReasoningEffort = "medium"
	defaultStream          = true
)

type ChatModelSettings struct {
	Model           string
	BaseURL         string
	ThinkingType    string
	ReasoningEffort string
	Stream          bool
}

func ResolveChatModelSettings() ChatModelSettings {
	return ChatModelSettings{
		Model:           getenvDefault("DEEPSEEK_MODEL", defaultModelType),
		BaseURL:         getenvDefault("DEEPSEEK_BASE_URL", defaultModelURL),
		ThinkingType:    getenvDefault("DEEPSEEK_THINKING_TYPE", defaultThinkingType),
		ReasoningEffort: getenvDefault("DEEPSEEK_REASONING_EFFORT", defaultReasoningEffort),
		Stream:          getenvBoolDefault("DEEPSEEK_STREAM", defaultStream),
	}
}

func buildExtraFields(settings ChatModelSettings) map[string]any {
	if settings.ThinkingType == "" {
		return nil
	}

	return map[string]any{
		"thinking": map[string]any{
			"type": settings.ThinkingType,
		},
	}
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getenvBoolDefault(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		log.Fatalf("%s must be a boolean value: %v", key, err)
	}

	return parsed
}
