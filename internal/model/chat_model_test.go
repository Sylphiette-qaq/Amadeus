package model

import (
	"testing"
)

func TestResolveChatModelSettingsReadsManualConfig(t *testing.T) {
	t.Setenv("DEEPSEEK_MODEL", "deepseek-v4-pro")
	t.Setenv("DEEPSEEK_BASE_URL", "https://api.deepseek.com")
	t.Setenv("DEEPSEEK_THINKING_TYPE", "enabled")
	t.Setenv("DEEPSEEK_REASONING_EFFORT", "high")
	t.Setenv("DEEPSEEK_STREAM", "false")

	settings := ResolveChatModelSettings()
	if settings.Model != "deepseek-v4-pro" {
		t.Fatalf("Model = %q, want deepseek-v4-pro", settings.Model)
	}
	if settings.BaseURL != "https://api.deepseek.com" {
		t.Fatalf("BaseURL = %q, want https://api.deepseek.com", settings.BaseURL)
	}
	if settings.ThinkingType != "enabled" {
		t.Fatalf("ThinkingType = %q, want enabled", settings.ThinkingType)
	}
	if settings.ReasoningEffort != "high" {
		t.Fatalf("ReasoningEffort = %q, want high", settings.ReasoningEffort)
	}
	if settings.Stream {
		t.Fatal("Stream = true, want false")
	}
}

func TestResolveChatModelSettingsUsesFallbacks(t *testing.T) {
	t.Setenv("DEEPSEEK_MODEL", "")
	t.Setenv("DEEPSEEK_BASE_URL", "")
	t.Setenv("DEEPSEEK_THINKING_TYPE", "")
	t.Setenv("DEEPSEEK_REASONING_EFFORT", "")
	t.Setenv("DEEPSEEK_STREAM", "")

	settings := ResolveChatModelSettings()
	if settings.Model != defaultModelType {
		t.Fatalf("Model = %q, want %q", settings.Model, defaultModelType)
	}
	if settings.BaseURL != defaultModelURL {
		t.Fatalf("BaseURL = %q, want %q", settings.BaseURL, defaultModelURL)
	}
	if settings.ThinkingType != defaultThinkingType {
		t.Fatalf("ThinkingType = %q, want %q", settings.ThinkingType, defaultThinkingType)
	}
	if settings.ReasoningEffort != defaultReasoningEffort {
		t.Fatalf("ReasoningEffort = %q, want %q", settings.ReasoningEffort, defaultReasoningEffort)
	}
	if settings.Stream != defaultStream {
		t.Fatalf("Stream = %v, want %v", settings.Stream, defaultStream)
	}
}

func TestBuildExtraFieldsAddsThinkingConfig(t *testing.T) {
	fields := buildExtraFields(ChatModelSettings{ThinkingType: "enabled"})
	if fields == nil {
		t.Fatal("fields = nil, want thinking config")
	}

	thinking, ok := fields["thinking"].(map[string]any)
	if !ok {
		t.Fatalf("thinking field = %#v, want map[string]any", fields["thinking"])
	}
	if thinking["type"] != "enabled" {
		t.Fatalf("thinking type = %q, want enabled", thinking["type"])
	}
}
