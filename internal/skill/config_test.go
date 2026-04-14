package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigPrefersAbsolutePath(t *testing.T) {
	tempDir := t.TempDir()
	agentPath := filepath.Join(tempDir, "agent.md")
	skillRoot := filepath.Join(tempDir, "skills")

	if err := os.WriteFile(agentPath, []byte("- name: x\n  desc: y\n"), 0644); err != nil {
		t.Fatalf("write agent.md: %v", err)
	}
	if err := os.Mkdir(skillRoot, 0755); err != nil {
		t.Fatalf("mkdir skills: %v", err)
	}

	t.Setenv(envAgentMDAbs, agentPath)
	t.Setenv(envAgentMDRel, "./should-not-be-used")
	t.Setenv(envSkillRootAbs, skillRoot)
	t.Setenv(envSkillRootRel, "./should-not-be-used")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.AgentMDPath != agentPath {
		t.Fatalf("AgentMDPath = %q, want %q", cfg.AgentMDPath, agentPath)
	}
	if cfg.SkillRootPath != skillRoot {
		t.Fatalf("SkillRootPath = %q, want %q", cfg.SkillRootPath, skillRoot)
	}
}
