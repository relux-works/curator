package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The exact-source CLI rejects the r2 context-module witness before publishing
// source/lock metadata or contact bytes in either native root document. Clean
// and reserved-host modules install and appear in both documents. Context
// manifests have no capability declaration; declared skill hosts are covered by
// TestProfileInstallMarkdownScope and TestProfileInstallMarkdownCaseAliasesCLI.
func TestProfileInstallMarkdownContextCLI(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "curator")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, "./cmd/curator")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build exact-source CLI: %v\n%s", err, output)
	}
	for _, tc := range []struct {
		name, body string
		allowed    bool
	}{
		{"undeclared", "Read https://api.example.net/guide\n", false},
		{"clean", "# Clean context\n", true},
		{"reserved", "Read https://api.example/guide\n", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := gitFixtureHome(t)
			pinOperatorHome(t, home)
			homes := pinHomes(t)
			profile := t.TempDir()
			writePackage(t, profile, "review-context", "1.0.0", tc.body)
			configPath := filepath.Join(home, "config.json")
			payload, err := json.Marshal(map[string]any{"schema_version": 1, "skills_root": filepath.Join(home, "skills"), "projects": map[string]any{}, "audit": map[string]any{"enabled": true, "mode": "strict", "fail_on": "high", "backend": "null"}})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(configPath, payload, 0600); err != nil {
				t.Fatal(err)
			}
			systemPath := filepath.Join(home, "system.json")
			if err := os.WriteFile(systemPath, []byte(`{"schema_version":1}`), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("CURATOR_CONFIG", configPath)
			t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
			command := exec.Command(binary, "profile", "install", profile)
			output, err := command.CombinedOutput()
			native := []string{filepath.Join(homes["claude_code"], "CLAUDE.md"), filepath.Join(homes["codex_cli"], "AGENTS.md")}
			if !tc.allowed {
				exit, ok := err.(*exec.ExitError)
				if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "audit.capability.network-undeclared") || !strings.Contains(string(output), "context/a.md") || !strings.Contains(string(output), "contacts api.example.net,") {
					t.Fatalf("CLI context not specifically refused: %v\n%s", err, output)
				}
				for _, path := range []string{sourcePath(home, "review-context"), lockPath(home, "review-context")} {
					if _, err := os.Stat(path); !os.IsNotExist(err) {
						t.Fatalf("refused context published profile metadata: %v", err)
					}
				}
				for _, path := range native {
					payload, err := os.ReadFile(path)
					if err != nil && !os.IsNotExist(err) {
						t.Fatal(err)
					}
					if strings.Contains(string(payload), "https://api.example.net/guide") {
						t.Fatal("refused context published native URL")
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("context positive control refused: %v\n%s", err, output)
			}
			for _, path := range []string{sourcePath(home, "review-context"), lockPath(home, "review-context")} {
				if _, err := os.Stat(path); err != nil {
					t.Fatalf("successful context omitted metadata: %v", err)
				}
			}
			for _, path := range native {
				payload, err := os.ReadFile(path)
				if err != nil || !strings.Contains(string(payload), tc.body) {
					t.Fatalf("successful context omitted native bytes: %v %q", err, payload)
				}
			}
		})
	}
}

// MCP members publish server configuration rather than Markdown, but retain
// the base source audit over all Markdown. A contact in an unlisted README
// refuses the whole profile before publication; clean and reserved controls
// still install their MCP member.
func TestProfileInstallMarkdownMCP(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		allowed    bool
	}{
		{"undeclared", "Read https://api.example.net/guide\n", false},
		{"clean", "# Clean\n", true},
		{"reserved", "Read https://api.example/guide\n", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := gitFixtureHome(t)
			pinOperatorHome(t, home)
			pinHomes(t)
			ids := newGitIdentities(t)
			mcp := gitRepo(t, map[string]string{
				"agent-mcp.json":     `{"schema_version":1,"name":"tool","version":"1.0.0","server":{"transport":"stdio","command":"tool","args":[],"env_names":[]}}`,
				"unlisted/README.md": tc.body,
			}, "v1.0.0")
			operand := ids.serve(mcp, "https://example.com/mcp/tool")
			profile := t.TempDir()
			writeManifestPackage(t, profile, `{"schema_version":1,"name":"mcp-profile","version":"1.0.0","requires":{"mcp":{"tool":{"git":"`+operand+`","range":"*"}}}}`, nil)
			info, _, _, err := Install(home, InstallOptions{Operand: profile})
			if !tc.allowed {
				if err == nil || !strings.Contains(err.Error(), "audit.capability.network-undeclared") || !strings.Contains(err.Error(), "unlisted/README.md") {
					t.Fatalf("MCP contact not specifically refused: %v", err)
				}
				for _, path := range []string{sourcePath(home, "mcp-profile"), lockPath(home, "mcp-profile")} {
					if _, err := os.Stat(path); !os.IsNotExist(err) {
						t.Fatalf("refused MCP published profile metadata: %v", err)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("MCP control refused: %v", err)
			}
			if _, ok := info.Lock.Find("mcp", "tool"); !ok {
				t.Fatal("successful MCP control omitted member")
			}
		})
	}
}
