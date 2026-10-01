package main

// Native lifecycle black box: the real curator binary, built once, drives
// install → shim run → cache-hit reinstall → remove of a schema-7 skill
// whose command builds from an external build repository, after a refused
// install over an ordinary checkout. The repository is a local git checkout
// selected through Skillfile.dev.json, so the run needs git and a Go
// toolchain but no network. The build completes only on hosts
// rc5-native-control-inventory-v1 covers (macOS, Windows); elsewhere the
// case skips with the inventory reason.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/testcli"
)

func TestNativeBlackboxExternalBuildLifecycle(t *testing.T) {
	requireNativeControlInventoryPlatform(t)
	testcli.RequireGit(t)
	bin := testcli.Binary(t)
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	skillsRoot := filepath.Join(root, "skills")
	project := filepath.Join(root, "project")
	home := filepath.Join(root, "home")
	if err := privatedir.Make(home); err != nil {
		t.Fatal(err)
	}
	if err := privatedir.Validate(home); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{filepath.Join(skillsRoot, "skill-a"), project} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	work := filepath.Join(project, "tools-src")
	draftCLIGit(t, "", "init", "--quiet", "-b", "main", "--object-format=sha1", work)
	writeBlackboxFiles(t, work, map[string]string{
		"skill-build.json":       `{"schema_version":1,"targets":{"tool":{"driver":"go-repository-v1","build_root":"tools","source_dir":"tools/cmd/tool"}}}`,
		"tools/go.mod":           "module example.test/tool\n\ngo 1.21\n",
		"tools/cmd/tool/main.go": "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"blackbox-tool-ok\") }\n",
	})
	draftCLIGit(t, work, "add", ".")
	draftCLIGit(t, work, "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "tool")
	lock := draftCLIGit(t, work, "rev-parse", "HEAD")

	skill := filepath.Join(skillsRoot, "skill-a")
	var payload []byte
	payload, err = json.Marshal(map[string]any{
		"schema_version": 7,
		"capabilities":   map[string]any{},
		"build_repositories": map[string]any{
			"tools": map[string]any{"git": "https://fixture.test/tools.git", "locked_commit": map[string]any{"object_format": "sha1", "hex": lock}},
		},
		"commands": map[string]any{
			"bbtool": map[string]any{"type": "build", "driver": "go-repository-v1", "repository": "tools", "target": "tool"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	writeBlackboxFiles(t, skill, map[string]string{
		"agent-skill.json": string(payload),
		"SKILL.md":         "---\nname: skill-a\n---\n# Skill\n",
	})
	draftCLIGit(t, "", "init", "--quiet", "-b", "main", skill)
	draftCLIGit(t, skill, "add", ".")
	draftCLIGit(t, skill, "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "skill")
	draftCLIGit(t, skill, "tag", "v1")

	writeBlackboxFiles(t, project, map[string]string{
		manifest.Name:        `{"schema_version":1,"agents":["codex_cli"],"skills":[{"name":"skill-a","tag":"v1"}]}`,
		"Skillfile.dev.json": `{"schema_version":2,"substitutions":{},"build_repository_substitutions":{"skill-a":{"tools":{"path":"tools-src"}}}}`,
		".gitignore":         ".agents/\n.codex/skills/\nSkillfile.dev.json\ntools-src/\n",
	})
	draftCLIGit(t, "", "init", "--quiet", "-b", "main", project)
	configPath := filepath.Join(home, "config.json")
	if err := config.Bootstrap(configPath, skillsRoot, "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	if err := config.AddProject(configPath, "app", project, []string{"codex_cli"}); err != nil {
		t.Fatal(err)
	}
	env := []string{"CURATOR_CONFIG=" + configPath, "HOME=" + home, "USERPROFILE=" + home}

	run := func(args ...string) string {
		t.Helper()
		code, stdout, stderr := testcli.Run(t, project, env, "", bin, args...)
		if code != 0 {
			t.Fatalf("curator %v exit=%d\nstdout:\n%s\nstderr:\n%s", args, code, stdout, stderr)
		}
		t.Logf("curator %v exit=0\n%s", args, stdout)
		if strings.Contains(stderr, "build cache sweep skipped") {
			t.Fatalf("cache cleanup could not prove the fixture's private home: %s", stderr)
		}
		return stdout
	}
	shim := filepath.Join(project, ".agents", "bin", "bbtool")
	if runtime.GOOS == "windows" {
		shim += ".cmd"
	}
	installedSkill := filepath.Join(project, ".agents", "skills", "skill-a")
	artifacts := filepath.Join(home, "external-build-cache", "artifacts")

	// The ordinary checkout still carries the hooks, info, and logs git
	// init/commit wrote, which local admission refuses: the install fails
	// before any build and publishes nothing.
	refusedCode, refusedOut, refusedErr := testcli.Run(t, project, env, "", bin, "install", "app")
	if refusedCode == 0 || !strings.Contains(refusedOut+refusedErr, "build_repository_source_unavailable") {
		t.Fatalf("install over an ordinary checkout: exit=%d\nstdout:\n%s\nstderr:\n%s", refusedCode, refusedOut, refusedErr)
	}
	t.Logf("curator [install app] over an ordinary checkout exit=%d\n%s%s", refusedCode, refusedOut, refusedErr)
	if _, err := os.Lstat(shim); !os.IsNotExist(err) {
		t.Fatalf("refused install published %s: %v", shim, err)
	}
	entries, err := os.ReadDir(artifacts)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("read %s: %v", artifacts, err)
	}
	if len(entries) != 0 {
		t.Fatalf("refused install published build roots: %v", entries)
	}
	pruneBlackboxGitDir(t, work)

	// Install: admission, audit, build into the cache, shim publication.
	stdout := run("install", "app")
	if !strings.Contains(stdout, "BUILD REPOSITORY SUBSTITUTION skill-a.tools") {
		t.Fatalf("install did not report the local substitution:\n%s", stdout)
	}
	key := blackboxOutcome(t, stdout, "would-preflight-and-build")
	buildRoot := filepath.Join(artifacts, key)
	receipt := filepath.Join(buildRoot, "receipt.json")
	firstReceipt := blackboxRead(t, receipt)
	artifactPath := filepath.Join(buildRoot, buildrepo.CacheArtifactName(runtime.GOOS))
	firstInfo, err := os.Stat(artifactPath)
	if err != nil {
		t.Fatalf("build root has no artifact after install: %v\n%s", err, stdout)
	}
	for _, path := range []string{shim, installedSkill} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("install did not publish %s: %v", path, err)
		}
	}

	// The shim resolves to the cached artifact and runs it.
	shimCommand := shim
	var shimArgs []string
	if runtime.GOOS == "windows" {
		// Batch shims need the native command interpreter; CreateProcess
		// cannot launch a .cmd file directly.
		shimCommand = "cmd.exe"
		shimArgs = []string{"/d", "/c", "call", shim}
	}
	code, shimOut, shimErr := testcli.Run(t, project, env, "", shimCommand, shimArgs...)
	if code != 0 || strings.TrimSpace(shimOut) != "blackbox-tool-ok" {
		t.Fatalf("shim exit=%d stdout=%q stderr=%q", code, shimOut, shimErr)
	}
	t.Logf("shim %s exit=0 stdout=%q", filepath.Base(shim), shimOut)

	// Reinstall: same key, cache hit, the artifact and receipt untouched.
	stdout = run("install", "app")
	if again := blackboxOutcome(t, stdout, "cache-hit"); again != key {
		t.Fatalf("reinstall key %s, first install key %s", again, key)
	}
	againInfo, err := os.Stat(artifactPath)
	if err != nil || !againInfo.ModTime().Equal(firstInfo.ModTime()) {
		t.Fatalf("cache hit rewrote the artifact: %v (%v -> %v)", err, firstInfo.ModTime(), againInfo)
	}
	if blackboxRead(t, receipt) != firstReceipt {
		t.Fatal("cache hit rewrote the receipt")
	}

	// Remove the declaration and reconcile: shim, skill, and build root go.
	if stdout := run("remove", "skill-a"); !strings.Contains(stdout, "removed skill-a") {
		t.Fatalf("remove stdout:\n%s", stdout)
	}
	if stdout := run("install", "app"); strings.Contains(stdout, "external build key=") {
		t.Fatalf("reconcile after remove still plans a build:\n%s", stdout)
	}
	for _, path := range []string{shim, installedSkill, buildRoot} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("%s survived remove: %v", path, err)
		}
	}
}

var blackboxKeyPattern = regexp.MustCompile(`skill-a\.bbtool external build key=sha256:([0-9a-f]{64}) outcome=(\S+)`)

// blackboxOutcome returns the external build key of the one bbtool row and
// fails unless its outcome is want.
func blackboxOutcome(t *testing.T, stdout, want string) string {
	t.Helper()
	rows := blackboxKeyPattern.FindAllStringSubmatch(stdout, -1)
	if len(rows) != 1 || rows[0][2] != want {
		t.Fatalf("want one bbtool build row with outcome=%s, got %q\nstdout:\n%s", want, rows, stdout)
	}
	return rows[0][1]
}

func blackboxRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// pruneBlackboxGitDir leaves only the administration children local
// admission accepts (HEAD, config, index, objects, refs, packed-refs);
// git init/commit also write hooks, info, logs, and description.
func pruneBlackboxGitDir(t *testing.T, work string) {
	t.Helper()
	allowed := map[string]bool{"HEAD": true, "config": true, "index": true, "objects": true, "refs": true, "packed-refs": true}
	gitDir := filepath.Join(work, ".git")
	entries, err := os.ReadDir(gitDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !allowed[entry.Name()] {
			if err := os.RemoveAll(filepath.Join(gitDir, entry.Name())); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func writeBlackboxFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
