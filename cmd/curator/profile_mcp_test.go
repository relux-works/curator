package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
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

// MCP takeover help is syntactic admission only. True spellings refuse every
// real action with the authority diagnostic and no manager/native writes;
// false spellings preserve normal read-only status/preview controls.
func TestProfileMCPTakeoverFlagContract(t *testing.T) {
	source, home := profileHome(t)
	pinOperatorHome(t)
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	configBody, _ := json.Marshal(map[string]any{"schema_version": 1, "projects": map[string]any{}, "skills_root": filepath.Join(home, "unused-skills")})
	if err := os.WriteFile(source.path, configBody, 0o600); err != nil {
		t.Fatal(err)
	}
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "sample", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("fixture: %d %s", code, stderr)
	}
	nativeHome := os.Getenv("HOME")
	before, nativeBefore := snapshotHelpTree(t, home), snapshotHelpTree(t, nativeHome)
	for _, verb := range []string{"install", "status", "repair", "restore", "replace", "remove", "release"} {
		for _, flag := range []string{"--takeover", "--takeover=true", "--takeover=1", "--takeover=TRUE", "-takeover"} {
			args := []string{"profile", "mcp", verb, "sample", flag}
			if code, _, stderr := runProfile(t, source, append(append([]string{}, args...), "--help")...); code != exitOK {
				t.Fatalf("help %v: %d %s", args, code, stderr)
			}
			if code, _, stderr := runProfile(t, source, args...); code != exitFail || !strings.Contains(stderr, mcpnative.AuthorityUnavailable) {
				t.Fatalf("MCP_TAKEOVER_AUTHORITY: %v: %d %s", args, code, stderr)
			}
		}
	}
	if !reflect.DeepEqual(before, snapshotHelpTree(t, home)) || !reflect.DeepEqual(nativeBefore, snapshotHelpTree(t, nativeHome)) {
		t.Fatal("takeover/help changed state")
	}
	// The existing CLI fixture supplies successful profile status and preview;
	// false flags must behave exactly like their unflagged invocation.
	for _, verb := range []string{"status", "install", "repair"} {
		args := []string{"profile", "mcp", verb, "sample"}
		if verb != "status" {
			args = append(args, "--dry-run")
		}
		wantCode, wantOut, wantErr := runProfile(t, source, args...)
		if wantCode != exitOK {
			t.Fatalf("valid false-flag control: %d %s", wantCode, wantErr)
		}
		for _, flag := range []string{"--takeover=false", "--takeover=0"} {
			code, out, stderr := runProfile(t, source, append(append([]string{}, args...), flag)...)
			if code != wantCode || out != wantOut || stderr != wantErr {
				t.Fatalf("false flag changed dispatch: %v: %d %s", args, code, stderr)
			}
		}
	}
	if !reflect.DeepEqual(before, snapshotHelpTree(t, home)) || !reflect.DeepEqual(nativeBefore, snapshotHelpTree(t, nativeHome)) {
		t.Fatal("false flag changed state")
	}
}
