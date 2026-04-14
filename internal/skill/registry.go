package skill

import (
	"fmt"
	"os"
	"strings"
)

func LoadAgentMarkdown(cfg Config) (string, error) {
	data, err := os.ReadFile(cfg.AgentMDPath)
	if err != nil {
		return "", fmt.Errorf("read agent.md: %w", err)
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return "", fmt.Errorf("agent.md is empty: %s", cfg.AgentMDPath)
	}

	if !strings.Contains(content, "name:") || !strings.Contains(content, "desc:") {
		return "", fmt.Errorf("agent.md must contain at least one skill name and desc: %s", cfg.AgentMDPath)
	}

	return content, nil
}
