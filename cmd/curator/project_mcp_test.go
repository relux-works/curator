package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/projectmcpstate"
	"github.com/relux-works/curator/internal/testgitenv"
)

func writeProjectMCP(t *testing.T, path string, value any) {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, payload, 0600); err != nil {
		t.Fatal(err)
	}
}
func readProjectMCP(t *testing.T, path string) []byte {
	t.Helper()
	p, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func patchMCPConfig(t *testing.T, path, key string, value any) {
	t.Helper()
	var obj map[string]any
	if err := json.Unmarshal(readProjectMCP(t, path), &obj); err != nil {
		t.Fatal(err)
	}
	obj[key] = value
	writeProjectMCP(t, path, obj)
}
func mcpCLISetup(t *testing.T, server map[string]any) (string, string, string) {
	t.Helper()
	root := testgitenv.ShortTempDir(t)
	cfg, project := setupCLIProject(t, root)
	for _, dir := range []string{filepath.Dir(cfg), project} {
		if err := privatedir.ProtectTree(dir); err != nil {
			t.Fatal(err)
		}
	}
	repo := filepath.Join(root, "package")
	if err := os.MkdirAll(repo, 0700); err != nil {
		t.Fatal(err)
	}
	writeProjectMCP(t, filepath.Join(repo, "agent-mcp.json"), map[string]any{"schema_version": 1, "name": "tool", "version": "1.0.0", "server": server})
	runGit(t, repo, "init", "-q", "-b", "main")
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-qm", "fixture")
	runGit(t, repo, "tag", "v1.0.0")
	gitcfg := filepath.Join(root, "gitconfig")
	remote := "file://" + filepath.ToSlash(repo)
	if filepath.VolumeName(repo) != "" {
		remote = "file:///" + filepath.ToSlash(repo)
	}
	if err := os.WriteFile(gitcfg, []byte("[url \""+remote+"\"]\n insteadOf = https://packages.example/tool\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", gitcfg)
	patchMCPConfig(t, cfg, "allowed_sources", []string{"packages.example/tool"})
	patchMCPConfig(t, cfg, "schema_version", 2)
	patchMCPConfig(t, cfg, "environments", map[string]any{"mcp_package_allowlist": []string{"packages.example/tool"}, "require_source_signers": false})
	writeProjectMCP(t, filepath.Join(project, "Skillfile.json"), map[string]any{"schema_version": 3, "agents": []string{"codex_cli"}, "skills": []any{}, "mcp_servers": map[string]any{"tool": map[string]any{"package": map[string]any{"git": "https://packages.example/tool", "tag": "v1.0.0"}, "install": map[string]any{"scope": "user", "approval": map[string]any{"approved": true}}}}})
	return cfg, project, repo
}

// The production CLI resolves a real fixture Git package, installs and previews
// the exact locked snapshot offline, reports every explicit target, and never
// starts a runtime or changes foreign native configuration. A runtime tripwire
// is invoked once as the positive control for the observation instrument.
func TestProjectMCPProductionReplay(t *testing.T) {
	cfg, project, repo := mcpCLISetup(t, map[string]any{"transport": "stdio", "command": "mcp-stage-a-tripwire", "args": []string{}, "environments": []string{"muse", "claude_code", "codex_cli", "pi", "opencode"}})
	bin := testgitenv.ShortTempDir(t)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(bin, "mcp-stage-a-tripwire")
	if filepath.Ext(exe) == ".exe" {
		target += ".exe"
	}
	if err := os.WriteFile(target, readProjectMCP(t, exe), 0700); err != nil {
		t.Fatal(err)
	}
	tripwire := filepath.Join(testgitenv.ShortTempDir(t), "invoked")
	t.Setenv("CURATOR_TEST_MCP_TRIPWIRE", tripwire)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	if err := exec.Command(target).Run(); err == nil {
		t.Fatal("tripwire must fail")
	}
	if _, err := os.Stat(tripwire); err != nil {
		t.Fatalf("tripwire not applied: %v", err)
	}
	if err := os.Remove(tripwire); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(project, ".mcp.json")
	if err := os.WriteFile(foreign, []byte(`{"foreign":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	code, out, stderr := capture(t, cfg, "project", "resolve", "app")
	if code != exitOK {
		t.Fatalf("resolve %d: %s %s", code, out, stderr)
	}
	if strings.Count(out, "path: unavailable (skipped: unsupported)") != 5 || !strings.Contains(out, `install option "scope": ignored`) {
		t.Fatal(out)
	}
	// CCJ identity ignores formatting while preserving the entire document.
	skillfile := filepath.Join(project, "Skillfile.json")
	if err := os.WriteFile(skillfile, append(readProjectMCP(t, skillfile), '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	before := readProjectMCP(t, envprofile.ProjectMCPLockPath(project))
	lock, err := projectmcpstate.DecodeLock(before)
	if err != nil {
		t.Fatal(err)
	}
	if lock.ProjectID != "app" || len(lock.Entries) != 1 || lock.Entries[0].Resolved.Version != "1.0.0" {
		t.Fatalf("lock: %+v", lock)
	}
	// Remote growth is deliberately unavailable: install must consume the pin.
	if err := os.Rename(repo, repo+"-offline"); err != nil {
		t.Fatal(err)
	}
	trace := filepath.Join(testgitenv.ShortTempDir(t), "git-trace")
	t.Setenv("GIT_TRACE", trace)
	for _, args := range [][]string{{"install", "app", "--dry-run"}, {"install", "app"}, {"status", "app"}} {
		homeBefore := snapshotHelpTree(t, filepath.Dir(cfg))
		projectBefore := snapshotHelpTree(t, project)
		code, out, stderr = capture(t, cfg, args...)
		if code != exitOK {
			t.Fatalf("%v: %d %s %s", args, code, out, stderr)
		}
		if (len(args) > 2 || args[0] == "status") && (!reflect.DeepEqual(homeBefore, snapshotHelpTree(t, filepath.Dir(cfg))) || !reflect.DeepEqual(projectBefore, snapshotHelpTree(t, project))) {
			t.Fatal("preview/status wrote manager or project state")
		}
		if args[0] == "install" && strings.Count(out, "path: unavailable (skipped: unsupported)") != 5 {
			t.Fatal(out)
		}
	}
	rawTrace := readProjectMCP(t, trace)
	for _, forbidden := range []string{"git fetch ", "git clone ", "git ls-remote "} {
		if strings.Contains(string(rawTrace), forbidden) {
			t.Fatalf("replay acquisition: %s", forbidden)
		}
	}
	if !bytes.Equal(before, readProjectMCP(t, envprofile.ProjectMCPLockPath(project))) {
		t.Fatal("replay rewrote lock")
	}
	if string(readProjectMCP(t, foreign)) != `{"foreign":true}` {
		t.Fatal("native config changed")
	}
	if _, err := os.Stat(tripwire); !os.IsNotExist(err) {
		t.Fatalf("runtime invoked: %v", err)
	}
	// Tightened machine policy refuses every independently known locked pair.
	patchMCPConfig(t, cfg, "environments", map[string]any{"mcp_package_allowlist": []string{"packages.example/other"}, "require_source_signers": false})
	homeBefore := snapshotHelpTree(t, filepath.Dir(cfg))
	code, out, stderr = capture(t, cfg, "install", "app", "--dry-run")
	if code != exitFail || strings.Count(out, "refused: mcp_package_not_allowed") != 5 || !strings.Contains(stderr, "mcp_package_not_allowed") {
		t.Fatalf("policy refusal %d %s %s", code, out, stderr)
	}
	if !reflect.DeepEqual(homeBefore, snapshotHelpTree(t, filepath.Dir(cfg))) {
		t.Fatal("policy refusal changed manager state")
	}
	patchMCPConfig(t, cfg, "environments", map[string]any{"mcp_package_allowlist": []string{"packages.example/tool"}, "require_source_signers": false})
	// A full-document semantic edit invalidates the source binding even when
	// normalized MCP declarations are unchanged.
	file := filepath.Join(project, "Skillfile.json")
	var edited map[string]any
	if err := json.Unmarshal(readProjectMCP(t, file), &edited); err != nil {
		t.Fatal(err)
	}
	edited["locale"] = "en"
	writeProjectMCP(t, file, edited)
	code, _, stderr = capture(t, cfg, "install", "app", "--dry-run")
	if code != exitFail || !strings.Contains(stderr, "source_lock_stale") {
		t.Fatalf("stale %d %s", code, stderr)
	}
	if !bytes.Equal(before, readProjectMCP(t, envprofile.ProjectMCPLockPath(project))) {
		t.Fatal("refusal changed lock")
	}
}

// Inline admission is refused before gitignore repair or source acquisition;
// ignored project-supplied approval and scope options cannot activate stage B.
func TestProjectMCPInlineRefusalBeforeWrites(t *testing.T) {
	cfg, project := setupCLIProject(t, testgitenv.ShortTempDir(t))
	writeProjectMCP(t, filepath.Join(project, "Skillfile.json"), map[string]any{"schema_version": 3, "skills": []any{}, "mcp_servers": map[string]any{"tool": map[string]any{"server": map[string]any{"transport": "http", "url": "https://service.example/mcp", "environments": []string{"codex_cli", "muse"}}, "install": map[string]any{"approval": true, "scope": "user"}}}})
	before := readProjectMCP(t, filepath.Join(project, ".gitignore"))
	for _, args := range [][]string{{"project", "resolve", "app"}, {"install", "app", "--dry-run"}, {"install", "app", "--fix-gitignore"}} {
		code, stdout, stderr := capture(t, cfg, args...)
		if strings.Count(stdout, "refused: mcp_inline_approval_unsupported") != 2 {
			t.Fatalf("missing inline pairs: %v: %s", args, stdout)
		}
		if code != exitFail || !strings.Contains(stderr, "mcp_inline_approval_unsupported") {
			t.Fatalf("%v %d %s", args, code, stderr)
		}
	}
	if !bytes.Equal(before, readProjectMCP(t, filepath.Join(project, ".gitignore"))) {
		t.Fatal("refusal changed gitignore")
	}
	if _, err := os.Stat(envprofile.ProjectMCPLockPath(project)); !os.IsNotExist(err) {
		t.Fatal("inline lock published")
	}
}

// Schema 1 remains accepted globally; schema 3 refuses through the real status
// and install entrypoints before a project-only MCP arm can become global.
func TestProjectMCPGlobalDispatch(t *testing.T) {
	home := globalScopeDeclaring(t, "")
	cfg := filepath.Join(home, "config.json")
	if code, _, stderr := capture(t, cfg, "global", "status"); code != exitOK {
		t.Fatalf("schema 1: %d %s", code, stderr)
	}
	writeFile(t, filepath.Join(install.GlobalRoot(home), manifest.Name), `{"schema_version":3,"skills":[],"mcp_servers":{}}`)
	for _, command := range []string{"status", "install"} {
		if code, _, stderr := capture(t, cfg, "global", command); code != exitFail || !strings.Contains(stderr, "global scope requires schema 1") {
			t.Fatalf("schema 3 %s: %d %s", command, code, stderr)
		}
	}
}

// An invalid declaration cannot hide independently parsed inline pairs on
// either side of its NAME. Resolve/preview write nothing; actual install only
// creates existing operation-lock bookkeeping. Consumer state stays unchanged,
// and malformed inputs and ignored option values never enter output.
func TestProjectMCPMalformedSiblingReporting(t *testing.T) {
	cfg, project := setupCLIProject(t, testgitenv.ShortTempDir(t))
	writeProjectMCP(t, filepath.Join(project, "Skillfile.json"), map[string]any{"schema_version": 3, "skills": []any{}, "mcp_servers": map[string]any{
		"a-valid": map[string]any{"server": map[string]any{"transport": "http", "url": "https://safe.example/mcp", "environments": []string{"codex_cli"}}},
		"bad":     map[string]any{"server": map[string]any{"transport": "http", "url": "https://safe.example/mcp", "private": "AKIA1234567890ABCDEF"}},
		"z-valid": map[string]any{"server": map[string]any{"transport": "http", "url": "https://safe.example/mcp", "environments": []string{"pi", "muse"}}},
	}})
	writeFile(t, filepath.Join(project, ".mcp.json"), `{"foreign":true}`)
	writeFile(t, envprofile.ProjectMCPLockPath(project), `{"prior":true}`)
	before := snapshotHelpTree(t, filepath.Dir(cfg))
	projectBefore := snapshotHelpTree(t, project)
	for _, args := range [][]string{{"project", "resolve", "app"}, {"install", "app", "--dry-run"}, {"install", "app", "--fix-gitignore"}} {
		code, stdout, stderr := capture(t, cfg, args...)
		if code != exitFail || !strings.Contains(stderr, "mcp_declaration_invalid") || strings.Count(stdout, "refused: mcp_inline_approval_unsupported") != 3 {
			t.Fatalf("sibling report %v: %d %s %s", args, code, stdout, stderr)
		}
		if args[0] == "project" || args[len(args)-1] == "--dry-run" {
			if !reflect.DeepEqual(before, snapshotHelpTree(t, filepath.Dir(cfg))) {
				t.Fatal("read-only refusal wrote manager state")
			}
		}
		if strings.Contains(stdout+stderr, "AKIA1234567890ABCDEF") || strings.Contains(stdout, "mcp bad ->") {
			t.Fatal("invalid declaration material escaped")
		}
	}
	// Real operations retain the existing manager lock footprint. All other
	// manager bytes and every project entry must retain the post-plant state.
	after := snapshotHelpTree(t, filepath.Dir(cfg))
	for name, value := range after {
		if name == "." || name == "state" || name == filepath.Join("state", "locks") || strings.HasPrefix(name, filepath.Join("state", "locks")+string(filepath.Separator)) {
			continue
		}
		if before[name] != value {
			t.Fatalf("changed manager consumer entry %s", name)
		}
	}
	for name, value := range before {
		if after[name] != value && name != "." && name != "state" {
			t.Fatalf("changed prior manager entry %s", name)
		}
	}
	if !reflect.DeepEqual(projectBefore, snapshotHelpTree(t, project)) {
		t.Fatal("refused sibling input changed project state")
	}
}
