package main

import (
	"strings"
	"testing"
)

func TestSessionIDFromConversationIDIsStableAndPathSafe(t *testing.T) {
	first := sessionIDFromConversationID(" group:123 ")
	second := sessionIDFromConversationID("group:123")

	if first != second {
		t.Fatalf("expected stable session ID for trimmed conversation ID, got %q and %q", first, second)
	}
	if !strings.HasPrefix(first, "conversation-group_123-") {
		t.Fatalf("expected readable normalized prefix, got %q", first)
	}

	for _, invalid := range []string{`<`, `>`, `:`, `"`, `/`, `\`, `|`, `?`, `*`} {
		if strings.Contains(first, invalid) {
			t.Fatalf("session ID %q contains invalid path character %q", first, invalid)
		}
	}
}

func TestSessionIDFromConversationIDAvoidsNormalizedCollisions(t *testing.T) {
	colonID := sessionIDFromConversationID("group:123")
	slashID := sessionIDFromConversationID("group/123")

	if colonID == slashID {
		t.Fatalf("expected different session IDs for distinct conversation IDs, got %q", colonID)
	}
}

func TestSessionIDFromConversationIDHandlesOnlyInvalidCharacters(t *testing.T) {
	sessionID := sessionIDFromConversationID("::")

	if !strings.HasPrefix(sessionID, "conversation-conversation-") {
		t.Fatalf("expected fallback readable name, got %q", sessionID)
	}
}
