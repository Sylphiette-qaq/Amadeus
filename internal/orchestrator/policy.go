package orchestrator

import (
	"os"
	"strconv"
)

const defaultMaxTurns = 8

func loadMaxTurns() int {
	raw := os.Getenv("AMADEUS_MAX_TURNS")
	if raw == "" {
		return defaultMaxTurns
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return defaultMaxTurns
	}

	return value
}
