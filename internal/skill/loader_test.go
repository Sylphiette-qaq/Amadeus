package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSkillContent(t *testing.T) {
	tempDir := t.TempDir()
	skillRoot := filepath.Join(tempDir, "skills")
	skillDir := filepath.Join(skillRoot, "demo")
	skillFile := filepath.Join(skillDir, "SKILL.md")

	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	if err := os.WriteFile(skillFile, []byte("# Demo"), 0644); err != nil {
		t.Fatalf("write skill file: %v", err)
	}

	doc, err := LoadSkillContent(Config{SkillRootPath: skillRoot}, "demo")
	if err != nil {
		t.Fatalf("LoadSkillContent() error = %v", err)
	}

	if doc.Name != "demo" {
		t.Fatalf("doc.Name = %q, want demo", doc.Name)
	}
	if doc.Path != skillFile {
		t.Fatalf("doc.Path = %q, want %q", doc.Path, skillFile)
	}
	if doc.Content != "# Demo" {
		t.Fatalf("doc.Content = %q, want %q", doc.Content, "# Demo")
	}
}

func TestLoadSkillContentRejectsInvalidName(t *testing.T) {
	_, err := LoadSkillContent(Config{SkillRootPath: t.TempDir()}, "../demo")
	if err == nil {
		t.Fatal("expected error for invalid skill name")
	}
}
