package session

import (
	"fmt"

	"Amadeus/internal/skill"
	"github.com/cloudwego/eino/schema"
)

type State struct {
	Messages       []*schema.Message
	CurrentTurn    int
	ToolCallCount  int
	Finished       bool
	LastToolResult string
	LoadedSkills   map[string]skill.Document
}

func NewState(history []*schema.Message, loadedSkills []skill.Document, systemText, userQuestion string) *State {
	messages := make([]*schema.Message, 0, len(history)+len(loadedSkills)+2)
	messages = append(messages, schema.SystemMessage(systemText))
	loaded := make(map[string]skill.Document, len(loadedSkills))
	for _, doc := range loadedSkills {
		if doc.Name == "" {
			continue
		}
		loaded[doc.Name] = doc
		messages = append(messages, schema.SystemMessage(BuildLoadedSkillMessage(doc)))
	}
	messages = append(messages, history...)
	messages = append(messages, schema.UserMessage(userQuestion))

	return &State{
		Messages:     messages,
		LoadedSkills: loaded,
	}
}

func (s *State) Append(message *schema.Message) {
	if message == nil {
		return
	}

	s.Messages = append(s.Messages, message)
}

func (s *State) HasLoadedSkill(name string) bool {
	if s == nil || s.LoadedSkills == nil {
		return false
	}

	_, ok := s.LoadedSkills[name]
	return ok
}

func (s *State) ActivateSkill(doc skill.Document) bool {
	if doc.Name == "" {
		return false
	}
	if s.LoadedSkills == nil {
		s.LoadedSkills = make(map[string]skill.Document)
	}
	if _, ok := s.LoadedSkills[doc.Name]; ok {
		return false
	}

	s.LoadedSkills[doc.Name] = doc
	s.Messages = append(s.Messages, schema.SystemMessage(BuildLoadedSkillMessage(doc)))
	return true
}

func BuildLoadedSkillMessage(doc skill.Document) string {
	return fmt.Sprintf("[Loaded Skill: %s]\n%s", doc.Name, doc.Content)
}
