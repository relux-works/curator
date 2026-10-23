package envprofile

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/projectmcp"
	"github.com/relux-works/curator/internal/stateread"
)

// Warm caches and policy switches cannot substitute another source's commit,
// even for identities colliding under separator replacement or truncation.
// Cold acquisitions and safe repeated acquisitions retain the exact fixture OID.
func TestProjectMCPCacheSourceBinding(t *testing.T) {
	for _, pair := range [][2]string{
		{"packages.example/a_b/tool", "packages.example/a/b_tool"},
		{"packages.example/a/" + strings.Repeat("segment/", 22) + "tool", "packages.example/b/" + strings.Repeat("segment/", 22) + "tool"},
	} {
		t.Run(pair[0], func(t *testing.T) {
			ids := newGitIdentities(t)
			home := t.TempDir()
			for index, source := range pair {
				payload, _ := json.Marshal(map[string]any{"schema_version": 1, "name": "tool", "version": "1.0.0", "server": map[string]any{"transport": "stdio", "command": "runtime", "args": []string{source}}})
				repo := gitRepo(t, map[string]string{"agent-mcp.json": string(payload)}, "v1.0.0")
				ids.serve(repo, "https://"+source)
				output, err := gitOutput(repo, "rev-parse", "HEAD")
				if err != nil {
					t.Fatal(err)
				}
				expected := strings.TrimSpace(output)
				decl, err := projectmcp.ParseDeclaration(map[string]any{"package": map[string]any{"git": "https://" + source, "tag": "v1.0.0"}})
				if err != nil {
					t.Fatal(err)
				}
				policy := Policy{AllowedSources: []string{source}, MCPAllowlist: []string{source}}
				for _, cache := range []string{home, t.TempDir()} {
					entries, _, err := ResolveProjectMCP(cache, map[string]projectmcp.Declaration{"tool": decl}, policy)
					if err != nil || len(entries) != 1 {
						t.Fatalf("source %d: %v", index, err)
					}
					if entries[0].Resolved.Commit != expected {
						t.Fatalf("wrong-source commit admitted: got %s want %s", entries[0].Resolved.Commit, expected)
					}
				}
			}
		})
	}
}

// A cache whose origin was changed is refused before fetch and during offline
// replay, despite an intact protected lock/snapshot. The unmodified sibling
// replays successfully and neither rejection rewrites the portable lock.
func TestProjectMCPCacheOriginRefused(t *testing.T) {
	home, root, lock, decls, policy := lockedProjectMCP(t, Policy{}, false)
	if _, _, err := ReplayProjectMCP(home, root, "project", lock.ManifestSHA256, decls, policy); err != nil {
		t.Fatal(err)
	}
	cache := ""
	err := filepath.WalkDir(home, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			cache = filepath.Dir(path)
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil || cache == "" {
		t.Fatalf("cache fixture: %v", err)
	}
	gitStatusCommand(t, cache, "remote", "set-url", "origin", "https://packages.example/other")
	before, err := os.ReadFile(filepath.Join(root, "Skillfile.mcp-lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := ResolveProjectMCP(home, decls, policy); err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("wrong origin acquired: %v", err)
	}
	if _, _, err := ReplayProjectMCP(home, root, "project", lock.ManifestSHA256, decls, policy); err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("wrong origin replayed: %v", err)
	}
	after, err := os.ReadFile(filepath.Join(root, "Skillfile.mcp-lock.json"))
	if err != nil || string(after) != string(before) {
		t.Fatalf("lock changed: %v", err)
	}
}

// A genuinely absent descendant is creatable, whereas an untraversable parent
// is a typed unreadable route and cannot masquerade as absence.
func TestProjectMCPRouteReadBoundary(t *testing.T) {
	root := t.TempDir()
	if err := validateMCPRoute(root, filepath.Join(root, "missing", "entry")); err != nil {
		t.Fatal(err)
	}
	parent := filepath.Join(root, "not-directory")
	if err := os.WriteFile(parent, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := validateMCPRoute(root, filepath.Join(parent, "entry")); err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) {
		t.Fatalf("unreadable route: %v", err)
	}
}

// Legacy lossy profile caches are never promoted to project MCP authority.
// A fresh source-bound acquisition succeeds despite a colliding legacy clone,
// and the legacy repository remains untouched for its existing consumers.
func TestProjectMCPLegacyCacheIsNotAuthority(t *testing.T) {
	ids := newGitIdentities(t)
	home := t.TempDir()
	first := "packages.example/a_b/tool"
	second := "packages.example/a/b_tool"
	for _, source := range []string{first, second} {
		payload, err := json.Marshal(map[string]any{"schema_version": 1, "name": "tool", "version": "1.0.0", "server": map[string]any{"transport": "stdio", "command": "runtime", "args": []string{source}}})
		if err != nil {
			t.Fatal(err)
		}
		ids.serve(gitRepo(t, map[string]string{"agent-mcp.json": string(payload)}, "v1.0.0"), "https://"+source)
	}
	legacy := newGitManager(home)
	legacy.recordRaw(first, "https://"+first)
	repo, err := legacy.ensureRepo(first)
	if err != nil {
		t.Fatal(err)
	}
	before, err := gitOutput(repo, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	decl, err := projectmcp.ParseDeclaration(map[string]any{"package": map[string]any{"git": "https://" + second, "tag": "v1.0.0"}})
	if err != nil {
		t.Fatal(err)
	}
	entries, _, err := ResolveProjectMCP(home, map[string]projectmcp.Declaration{"tool": decl}, Policy{AllowedSources: []string{second}, MCPAllowlist: []string{second}})
	if err != nil || len(entries) != 1 {
		t.Fatalf("legacy cache blocked independent acquisition: %v", err)
	}
	args := entries[0].Resolved.Server["args"].([]string)
	if len(args) != 1 || args[0] != second {
		t.Fatalf("legacy bytes became authority: %v", args)
	}
	after, err := gitOutput(repo, "rev-parse", "HEAD")
	if err != nil || after != before {
		t.Fatalf("legacy cache changed: %v", err)
	}
}
