package mcp

import (
	"reflect"
	"testing"
)

func TestClaudeNativeMCPPathsComeFromRegisteredPolicy(t *testing.T) {
	if !reflect.DeepEqual(projectSources["claude_code"], []string{".mcp.json"}) {
		t.Fatalf("Claude project MCP sources = %q", projectSources["claude_code"])
	}
	if !reflect.DeepEqual(homeSources["claude_code"], []string{".claude.json"}) {
		t.Fatalf("Claude home MCP sources = %q", homeSources["claude_code"])
	}
	if !reflect.DeepEqual(claudeSettings, []string{".claude/settings.json", ".claude/settings.local.json"}) {
		t.Fatalf("Claude settings sources = %q", claudeSettings)
	}
	if len(projectSources[""]) != 0 || len(homeSources[""]) != 0 {
		t.Fatal("temporary empty path key escaped initialization")
	}
}
