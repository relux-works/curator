package envprofile

import (
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

func TestClaudeHomeAndNativeSkillRoutingUsesRegistry(t *testing.T) {
	adapter, ok := adapterByID(envregistry.ClaudeCode)
	if !ok || adapter.EnvVar != "CLAUDE_CONFIG_DIR" || adapter.DefaultDir != ".claude" || adapter.Target != "CLAUDE.md" {
		t.Fatalf("Claude switch adapter = %#v, found %v", adapter, ok)
	}
	if got := inPlaceSkillsDir(envregistry.ClaudeCode); got != "skills" {
		t.Fatalf("Claude native skills path = %q", got)
	}
	if got := inPlaceSkillsDir(envregistry.CodexCLI); got != "skills" {
		t.Fatalf("Codex native skills path = %q", got)
	}
	if got := inPlaceSkillsDir(envregistry.Pi); got != "skills" {
		t.Fatalf("Pi native skills path = %q", got)
	}
	for _, id := range []string{envregistry.OpenCode, envregistry.Muse, "cursor"} {
		if got := inPlaceSkillsDir(id); got != profileSkillsSurfaceDir {
			t.Fatalf("nonparticipant %s native skills path = %q", id, got)
		}
	}
}
