package memory

import (
	"strings"

	"github.com/cloudwego/eino/schema"
)

func parseContextLine(line string) *schema.Message {
	idx1 := strings.Index(line, "]")
	if idx1 == -1 {
		return nil
	}

	rest := strings.TrimSpace(line[idx1+1:])
	idx2 := strings.Index(rest, ":")
	if idx2 == -1 {
		return nil
	}

	roleStr := strings.TrimSpace(rest[:idx2])
	content := strings.TrimSpace(rest[idx2+1:])

	var role schema.RoleType
	switch roleStr {
	case "user":
		role = schema.User
	case "assistant":
		role = schema.Assistant
	case "system":
		role = schema.System
	default:
		return nil
	}

	return &schema.Message{
		Role:    role,
		Content: content,
	}
}
