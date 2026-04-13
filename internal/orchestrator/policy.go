package orchestrator

import (
	"os"
	"strconv"
)

const defaultMaxTurns = 8

func loadMaxTurns() int {
	raw := os.Getenv("AMADEUS_MAX_TURNS")
	if raw == "" {
		// 没有显式配置时采用保守默认值，防止工具调用异常导致无限循环。
		return defaultMaxTurns
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return defaultMaxTurns
	}

	return value
}
