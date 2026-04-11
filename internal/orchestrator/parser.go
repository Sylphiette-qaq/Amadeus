package orchestrator

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ParseToolArguments(arguments string) error {
	if strings.TrimSpace(arguments) == "" {
		return fmt.Errorf("empty arguments")
	}

	if !json.Valid([]byte(arguments)) {
		return fmt.Errorf("arguments is not valid JSON")
	}

	return nil
}
