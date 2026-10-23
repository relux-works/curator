package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/scopes"
)

// Hybrid schema 1 works through list/status/preview/install. Project-only
// schemas refuse those same entrypoints, preserve both manifests/gitignore,
// and cannot activate project state from the dropped hybrid MCP declaration.
func TestProjectMCPHybridDispatch(t *testing.T) {
	for _, version := range []int{1, 2, 3} {
		t.Run(fmt.Sprint(version), func(t *testing.T) {
			cfg, project := setupCLIProject(t, t.TempDir())
			writeProjectMCP(t, filepath.Join(project, "Skillfile.json"), map[string]any{"schema_version": 1, "skills": []any{}})
			hybrid := scopes.HybridManifestPath(filepath.Dir(cfg))
			if err := os.MkdirAll(filepath.Dir(hybrid), 0700); err != nil {
				t.Fatal(err)
			}
			value := map[string]any{"schema_version": version, "skills": []any{}}
			if version == 3 {
				value["mcp_servers"] = map[string]any{"inline": "https://service.example/mcp"}
			}
			writeProjectMCP(t, hybrid, value)
			beforeHybrid := readProjectMCP(t, hybrid)
			beforeProject := readProjectMCP(t, filepath.Join(project, "Skillfile.json"))
			beforeIgnore := readProjectMCP(t, filepath.Join(project, ".gitignore"))
			for _, args := range [][]string{{"hybrid", "list"}, {"status", "app"}, {"install", "app", "--dry-run"}, {"install", "app", "--fix-gitignore"}} {
				code, _, stderr := capture(t, cfg, args...)
				if version == 1 {
					if code != exitOK {
						t.Fatalf("valid %v: %d %s", args, code, stderr)
					}
					continue
				}
				if code != exitFail || !strings.Contains(stderr, "requires a newer tool") {
					t.Fatalf("hybrid project schema admitted by %v: %d %s", args, code, stderr)
				}
			}
			if version != 1 && (!bytes.Equal(beforeHybrid, readProjectMCP(t, hybrid)) || !bytes.Equal(beforeProject, readProjectMCP(t, filepath.Join(project, "Skillfile.json"))) || !bytes.Equal(beforeIgnore, readProjectMCP(t, filepath.Join(project, ".gitignore")))) {
				t.Fatal("hybrid refusal changed consumer inputs")
			}
		})
	}
}
