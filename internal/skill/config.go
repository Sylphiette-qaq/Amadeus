package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	envAgentMDRel   = "SKILL_AGENT_MD_REL"
	envAgentMDAbs   = "SKILL_AGENT_MD_ABS"
	envSkillRootRel = "SKILL_ROOT_REL"
	envSkillRootAbs = "SKILL_ROOT_ABS"
)

type Config struct {
	AgentMDPath   string
	SkillRootPath string
}

func LoadConfig() (Config, error) {
	agentMDPath, err := resolvePath(
		os.Getenv(envAgentMDAbs),
		os.Getenv(envAgentMDRel),
		"agent.md",
	)
	if err != nil {
		return Config{}, fmt.Errorf("resolve agent.md path: %w", err)
	}

	skillRootPath, err := resolvePath(
		os.Getenv(envSkillRootAbs),
		os.Getenv(envSkillRootRel),
		"skill root",
	)
	if err != nil {
		return Config{}, fmt.Errorf("resolve skill root path: %w", err)
	}

	if err := validateFile(agentMDPath, "agent.md"); err != nil {
		return Config{}, err
	}

	if err := validateDir(skillRootPath, "skill root"); err != nil {
		return Config{}, err
	}

	return Config{
		AgentMDPath:   agentMDPath,
		SkillRootPath: skillRootPath,
	}, nil
}

func resolvePath(absPath, relPath, label string) (string, error) {
	if strings.TrimSpace(absPath) != "" {
		return filepath.Abs(absPath)
	}

	if strings.TrimSpace(relPath) != "" {
		return filepath.Abs(relPath)
	}

	return "", fmt.Errorf("%s path is not configured; set one of the env vars for abs/rel path", label)
}

func validateFile(path, label string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s does not exist: %w", label, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s must be a file: %s", label, path)
	}

	return nil
}

func validateDir(path, label string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s does not exist: %w", label, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s must be a directory: %s", label, path)
	}

	return nil
}
