package orchestrator

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ParseToolArguments(arguments string) error {
	// M1 先做最小校验：保证 arguments 至少是非空 JSON。
	// 更严格的字段级 Schema 校验放到后续 ToolExecutor/策略层。
	if strings.TrimSpace(arguments) == "" {
		return fmt.Errorf("empty arguments")
	}

	if !json.Valid([]byte(arguments)) {
		return fmt.Errorf("arguments is not valid JSON")
	}

	return nil
}
