package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var skillNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type Document struct {
	Name    string `json:"skill_name"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

func LoadSkillContent(cfg Config, name string) (Document, error) {
	name = strings.TrimSpace(name)
	if !skillNamePattern.MatchString(name) {
		return Document{}, fmt.Errorf("invalid skill name: %q", name)
	}

	skillPath := filepath.Join(cfg.SkillRootPath, name, "SKILL.md")
	cleanPath := filepath.Clean(skillPath)
	relPath, err := filepath.Rel(cfg.SkillRootPath, cleanPath)
	if err != nil {
		return Document{}, fmt.Errorf("resolve skill path: %w", err)
	}
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return Document{}, fmt.Errorf("skill path escapes root: %q", name)
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return Document{}, fmt.Errorf("read skill file: %w", err)
	}

	return Document{
		Name:    name,
		Path:    cleanPath,
		Content: string(data),
	}, nil
}
