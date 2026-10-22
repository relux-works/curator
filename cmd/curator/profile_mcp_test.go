package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/mcpnative"
)

// The real dispatcher exposes a separate profile carrier and read-only status.
// No project scope or explicit CLI action confers ownership-action authority.
func TestProfileMCPCLI(t *testing.T) {
	source, home := profileHome(t)
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	configBody, _ := json.Marshal(map[string]any{"schema_version": 1, "projects": map[string]any{}, "skills_root": filepath.Join(home, "unused-skills")})
	if err := os.WriteFile(source.path, configBody, 0o600); err != nil {
		t.Fatal(err)
	}
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "sample", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("fixture install: %d %s", code, stderr)
	}
	for _, args := range [][]string{{"status", "sample"}, {"install", "sample", "--dry-run"}, {"repair", "sample", "--dry-run"}} {
		command := append([]string{"profile", "mcp"}, args...)
		code, stdout, stderr := runProfile(t, source, command...)
		if code != exitOK {
			t.Fatalf("%v: %d %s", args, code, stderr)
		}
		var result envprofile.ProfileMCPResult
		if err := json.Unmarshal([]byte(stdout), &result); err != nil || result.Capability != mcpnative.Capability || result.Profile != "sample" || result.Published || len(result.Rows) != 0 {
			t.Fatalf("carrier: %s %v", stdout, err)
		}
	}
	for _, action := range []string{"restore", "replace", "remove", "release"} {
		code, _, stderr := runProfile(t, source, "profile", "mcp", action, "sample")
		if code != exitFail || !strings.Contains(stderr, mcpnative.AuthorityUnavailable) {
			t.Fatalf("%s conferred authority: %d %s", action, code, stderr)
		}
	}
	code, _, stderr := runProfile(t, source, "profile", "mcp", "install", "sample", "--takeover")
	if code != exitFail || !strings.Contains(stderr, mcpnative.AuthorityUnavailable) {
		t.Fatalf("takeover: %d %s", code, stderr)
	}
	code, _, stderr = runProfile(t, source, "profile", "mcp", "install", "sample", "--scope", "project")
	if code != exitFail || !strings.Contains(stderr, "project_scope_shared_home") {
		t.Fatalf("scope: %d %s", code, stderr)
	}
	for _, args := range [][]string{{}, {"status"}, {"unknown", "sample"}, {"status", "sample", "--unknown"}} {
		code, _, _ := runProfile(t, source, append([]string{"profile", "mcp"}, args...)...)
		if code != exitUsage {
			t.Fatalf("invalid args %v: %d", args, code)
		}
	}
	if _, err := os.Lstat(filepath.Join(home, "mcp-native-v1")); !os.IsNotExist(err) {
		t.Fatalf("read-only/refused commands created ownership: %v", err)
	}
}
