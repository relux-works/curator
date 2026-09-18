package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envprofile"
)

// These tests drive the §2.3 surfacing rows and the §2.2 empty-allowlist
// warning through the production run() entry point.

func writeGitRepoFile(t *testing.T, repo, name, content string) {
	t.Helper()
	full := filepath.Join(repo, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runGitRepo(t *testing.T, repo string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func commitGitRepo(t *testing.T, repo, tag string) {
	t.Helper()
	runGitRepo(t, repo, "add", ".")
	runGitRepo(t, repo, "commit", "-m", tag)
	runGitRepo(t, repo, "tag", tag)
}

// serveGitRepos maps https identities onto local fixture repositories
// through one insteadOf gitconfig.
func serveGitRepos(t *testing.T, repos map[string]string) {
	t.Helper()
	var rewrite strings.Builder
	for identity, repo := range repos {
		rewrite.WriteString("[url \"" + gitFileURL(repo) + "\"]\n\tinsteadOf = " + identity + "\n")
	}
	gitconfig := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(gitconfig, []byte(rewrite.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", gitconfig)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
}

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git on PATH")
	}
}

const surfacingToolManifest = `{"schema_version": 1, "name": "tool", "version": "1.0.0",` +
	`"server": {"transport": "stdio", "command": "npx", "args": ["-y", "tool"], "env_names": ["TOOL_TOKEN"]}}` + "\n"

const surfacingToolRow = `mcp-declaration tool 1.0.0 stdio command=npx args=["-y","tool"] env_names=["TOOL_TOKEN"]`

// TestProfileInstallListsStdioCommands drives profile install on a root
// requiring a stdio declaration: stdout lists the exact §2.3 row before
// the installed report, stderr carries mcp_package_allowlist_empty, and
// env status repeats the row with the S4 posture.
func TestProfileInstallListsStdioCommands(t *testing.T) {
	requireGit(t)
	source, _ := profileHome(t)
	tool := t.TempDir()
	writeGitRepoFile(t, tool, "agent-mcp.json", surfacingToolManifest)
	runGitRepo(t, tool, "init")
	commitGitRepo(t, tool, "v1.0.0")
	root := t.TempDir()
	writeGitRepoFile(t, root, "agent-context.json", `{"schema_version": 1, "name": "withmcp", "version": "1.0.0",`+
		`"requires": {"mcp": {"tool": {"git": "https://example.com/mcp-tool", "range": "*"}}}}`+"\n")
	runGitRepo(t, root, "init")
	commitGitRepo(t, root, "v1.0.0")
	serveGitRepos(t, map[string]string{"https://example.com/mcp-tool": tool, "https://example.com/withmcp": root})
	code, stdout, stderr := runProfile(t, source, "profile", "install", "https://example.com/withmcp")
	if code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	rowAt := strings.Index(stdout, surfacingToolRow)
	if rowAt < 0 {
		t.Fatalf("install stdout lists no stdio row:\n%s", stdout)
	}
	installedAt := strings.Index(stdout, "installed")
	if installedAt < 0 || rowAt > installedAt {
		t.Fatalf("the surfacing row must precede the installed report:\n%s", stdout)
	}
	if !strings.Contains(stderr, "mcp_package_allowlist_empty") {
		t.Fatalf("install stderr warns no empty allowlist:\n%s", stderr)
	}
	if !strings.Contains(stderr, "every declaration package in the closure is admitted") {
		t.Fatalf("the warning states no admission:\n%s", stderr)
	}
	code, stdout, _ = runProfile(t, source, "env", "status")
	if code != exitOK {
		t.Fatalf("status = %d\n%s", code, stdout)
	}
	if !strings.Contains(stdout, surfacingToolRow) {
		t.Fatalf("env status repeats no surfacing row:\n%s", stdout)
	}
	if !strings.Contains(stdout, "s4_profile: s4-warn, passable_env_names: unbounded") {
		t.Fatalf("env status postures no S4 profile:\n%s", stdout)
	}
}

// TestProfileUpdateListsNewDeclaration drives profile update onto a root
// whose new revision adds a declaration: the update moves and stdout
// lists the candidate set's row.
func TestProfileUpdateListsNewDeclaration(t *testing.T) {
	requireGit(t)
	source, _ := profileHome(t)
	tool := t.TempDir()
	writeGitRepoFile(t, tool, "agent-mcp.json", surfacingToolManifest)
	runGitRepo(t, tool, "init")
	commitGitRepo(t, tool, "v1.0.0")
	root := t.TempDir()
	writeGitRepoFile(t, root, "agent-context.json", `{"schema_version": 1, "name": "grows", "version": "1.0.0"}`+"\n")
	runGitRepo(t, root, "init")
	commitGitRepo(t, root, "v1.0.0")
	serveGitRepos(t, map[string]string{"https://example.com/mcp-tool": tool, "https://example.com/grows": root})
	if code, _, stderr := runProfile(t, source, "profile", "install", "https://example.com/grows"); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	writeGitRepoFile(t, root, "agent-context.json", `{"schema_version": 1, "name": "grows", "version": "1.0.1",`+
		`"requires": {"mcp": {"tool": {"git": "https://example.com/mcp-tool", "range": "*"}}}}`+"\n")
	commitGitRepo(t, root, "v1.0.1")
	code, stdout, stderr := runProfile(t, source, "profile", "update", "grows")
	if code != exitOK {
		t.Fatalf("update = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, surfacingToolRow) {
		t.Fatalf("update stdout lists no declaration row:\n%s", stdout)
	}
	if !strings.Contains(stdout, "updated") {
		t.Fatalf("update stdout reports no move:\n%s", stdout)
	}
	if !strings.Contains(stderr, "mcp_package_allowlist_empty") {
		t.Fatalf("update stderr warns no empty allowlist:\n%s", stderr)
	}
}

// TestProfileInstallWarnsEmptyAllowlist narrows the warning gate without
// git: a declaration-free install warns mcp_package_allowlist_empty and
// prints no surfacing rows.
func TestProfileInstallWarnsEmptyAllowlist(t *testing.T) {
	source, _ := profileHome(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	code, stdout, stderr := runProfile(t, source, "profile", "install", pkg)
	if code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stderr, "mcp_package_allowlist_empty") {
		t.Fatalf("install stderr warns no empty allowlist:\n%s", stderr)
	}
	if strings.Contains(stdout, "mcp-declaration") {
		t.Fatalf("an empty MCP set must print no rows:\n%s", stdout)
	}
}

// publicationWriter records stdout and snapshots whether the profile lock
// exists when the first mcp-declaration row arrives. Each row arrives in
// its own Write, so the check fires exactly at the emission point.
type publicationWriter struct {
	buffer         bytes.Buffer
	lock           string
	observed       bool
	publishedAtRow bool
}

func (w *publicationWriter) Write(p []byte) (int, error) {
	if !w.observed && bytes.Contains(p, []byte("mcp-declaration ")) {
		w.observed = true
		_, err := os.Stat(w.lock)
		w.publishedAtRow = err == nil
	}
	return w.buffer.Write(p)
}

// TestProfileInstallSurfacesBeforePublication drives profile install
// through the production run() entry point with a stdout writer that
// stats the profile lock when the first mcp-declaration row arrives:
// the row prints before the lock is published, exactly once, and before
// the installed report.
func TestProfileInstallSurfacesBeforePublication(t *testing.T) {
	requireGit(t)
	source, home := profileHome(t)
	tool := t.TempDir()
	writeGitRepoFile(t, tool, "agent-mcp.json", surfacingToolManifest)
	runGitRepo(t, tool, "init")
	commitGitRepo(t, tool, "v1.0.0")
	root := t.TempDir()
	writeGitRepoFile(t, root, "agent-context.json", `{"schema_version": 1, "name": "withmcp", "version": "1.0.0",`+
		`"requires": {"mcp": {"tool": {"git": "https://example.com/mcp-tool", "range": "*"}}}}`+"\n")
	runGitRepo(t, root, "init")
	commitGitRepo(t, root, "v1.0.0")
	serveGitRepos(t, map[string]string{"https://example.com/mcp-tool": tool, "https://example.com/withmcp": root})
	writer := &publicationWriter{lock: filepath.Join(envprofile.ProfileDir(home, "withmcp"), "lock.json")}
	var stderr strings.Builder
	code := run([]string{"profile", "install", "https://example.com/withmcp"}, source, writer, &stderr)
	stdout := writer.buffer.String()
	if !writer.observed {
		t.Fatal("no mcp-declaration row reached stdout")
	}
	if writer.publishedAtRow {
		t.Fatalf("lock.json already exists when first mcp-declaration row is written: %s", writer.lock)
	}
	if code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr.String())
	}
	if got := strings.Count(stdout, surfacingToolRow); got != 1 {
		t.Fatalf("the surfacing row prints %d times, want exactly once:\n%s", got, stdout)
	}
	if rowAt, installedAt := strings.Index(stdout, surfacingToolRow), strings.Index(stdout, "installed"); installedAt < 0 || rowAt > installedAt {
		t.Fatalf("the surfacing row must precede the installed report:\n%s", stdout)
	}
	if !strings.Contains(stderr.String(), "mcp_package_allowlist_empty") {
		t.Fatalf("install stderr warns no empty allowlist:\n%s", stderr.String())
	}
}

// updatePublicationWriter records stdout and snapshots the profile lock's
// exact bytes when the first mcp-declaration row arrives.
type updatePublicationWriter struct {
	buffer   bytes.Buffer
	lock     string
	old      []byte
	observed bool
	existed  bool
	matched  bool
}

func (w *updatePublicationWriter) Write(p []byte) (int, error) {
	if !w.observed && bytes.Contains(p, []byte("mcp-declaration ")) {
		w.observed = true
		payload, err := os.ReadFile(w.lock) // #nosec G304 -- test lock path
		w.existed = err == nil
		w.matched = bytes.Equal(payload, w.old)
	}
	return w.buffer.Write(p)
}

// TestProfileUpdateSurfacesBeforePublication drives profile update onto a
// root whose new revision adds a declaration, through the production
// run() entry point: the update moves, the row prints while lock.json
// still holds the old lock bytes, exactly once.
func TestProfileUpdateSurfacesBeforePublication(t *testing.T) {
	requireGit(t)
	source, home := profileHome(t)
	tool := t.TempDir()
	writeGitRepoFile(t, tool, "agent-mcp.json", surfacingToolManifest)
	runGitRepo(t, tool, "init")
	commitGitRepo(t, tool, "v1.0.0")
	root := t.TempDir()
	writeGitRepoFile(t, root, "agent-context.json", `{"schema_version": 1, "name": "grows", "version": "1.0.0"}`+"\n")
	runGitRepo(t, root, "init")
	commitGitRepo(t, root, "v1.0.0")
	serveGitRepos(t, map[string]string{"https://example.com/mcp-tool": tool, "https://example.com/grows": root})
	if code, _, stderr := runProfile(t, source, "profile", "install", "https://example.com/grows"); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	lock := filepath.Join(envprofile.ProfileDir(home, "grows"), "lock.json")
	oldBytes, err := os.ReadFile(lock) // #nosec G304 -- test lock path
	if err != nil {
		t.Fatal(err)
	}
	writeGitRepoFile(t, root, "agent-context.json", `{"schema_version": 1, "name": "grows", "version": "1.0.1",`+
		`"requires": {"mcp": {"tool": {"git": "https://example.com/mcp-tool", "range": "*"}}}}`+"\n")
	commitGitRepo(t, root, "v1.0.1")
	writer := &updatePublicationWriter{lock: lock, old: oldBytes}
	var stderr strings.Builder
	code := run([]string{"profile", "update", "grows"}, source, writer, &stderr)
	stdout := writer.buffer.String()
	if !writer.observed {
		t.Fatal("no mcp-declaration row reached stdout")
	}
	if !writer.existed {
		t.Fatal("lock.json is absent when the update emits: the old lock must still stand")
	}
	if !writer.matched {
		t.Fatal("lock.json at the first row is not the old lock: the row must precede publication")
	}
	if code != exitOK {
		t.Fatalf("update = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr.String())
	}
	if got := strings.Count(stdout, surfacingToolRow); got != 1 {
		t.Fatalf("the surfacing row prints %d times, want exactly once:\n%s", got, stdout)
	}
	if !strings.Contains(stdout, "updated") {
		t.Fatalf("update stdout reports no move:\n%s", stdout)
	}
	if !strings.Contains(stderr.String(), "mcp_package_allowlist_empty") {
		t.Fatalf("update stderr warns no empty allowlist:\n%s", stderr.String())
	}
}
