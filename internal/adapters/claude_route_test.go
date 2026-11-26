package adapters

import "testing"

func TestClaudeSkillPathComesFromRegisteredPolicy(t *testing.T) {
	if AgentPaths["claude_code"] != ".claude/skills" {
		t.Fatalf("Claude project skill path = %q", AgentPaths["claude_code"])
	}
	if _, exists := AgentPaths[""]; exists {
		t.Fatal("empty agent key escaped initialization")
	}
	if !KnownAgents()["claude_code"] {
		t.Fatal("registered Claude adapter is not recognized")
	}
}
