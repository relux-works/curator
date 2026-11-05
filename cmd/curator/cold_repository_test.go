//go:build curator_cold_repository && linux

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/testcli"
)

// A full real-binary install admits a large vendored Git fixture, compiles it
// with the repository driver, publishes its receipt/shim, and runs it from PATH.
// An ordinary unpruned checkout is the paired refusal and publishes nothing.
func TestColdRepositoryInstallOnPATH(t *testing.T) {
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
	source := os.Getenv("CURATOR_COLD_REPOSITORY")
	if source == "" {
		t.Fatal("CURATOR_COLD_REPOSITORY must name a vendored fixture")
	}
	command := os.Getenv("CURATOR_COLD_COMMAND")
	if command == "" {
		command = "tool"
	}
	exportName := "coldtool"
	if command == "a2a" {
		exportName = "a2a"
	}
	if err := os.CopyFS(filepath.Join(work, "tools"), os.DirFS(source)); err != nil {
		t.Fatal(err)
	}
	writeBlackboxFiles(t, work, map[string]string{
		"skill-build.json": `{"schema_version":1,"targets":{"tool":{"driver":"go-repository-v1","build_root":"tools","source_dir":"tools/cmd/` + command + `"}}}`,
	})
	var count int
	var size int64
	if err := filepath.WalkDir(filepath.Join(work, "tools"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			count++
			size += info.Size()
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	t.Logf("real repository files=%d bytes=%d", count, size)
	if count < 700 || size < 12*1024*1024 {
		t.Fatalf("fixture is smaller than the cold repository witness: files=%d bytes=%d", count, size)
	}

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
			exportName: map[string]any{"type": "build", "driver": "go-repository-v1", "repository": "tools", "target": "tool"},
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
	shim := filepath.Join(project, ".agents", "bin", exportName)
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
	var key string
	prefix := "skill-a." + exportName + " external build key=sha256:"
	for _, line := range strings.Split(stdout, "\n") {
		if at := strings.Index(line, prefix); at >= 0 {
			fields := strings.Fields(line[at+len(prefix):])
			if len(fields) > 1 && fields[1] == "outcome=would-preflight-and-build" {
				key = fields[0]
			}
		}
	}
	if len(key) != 64 {
		t.Fatal(fmt.Sprintf("install did not publish a real build outcome: %s", stdout))
	}

	buildRoot := filepath.Join(artifacts, key)
	receipt := filepath.Join(buildRoot, "receipt.json")
	_ = blackboxRead(t, receipt)
	artifactPath := filepath.Join(buildRoot, buildrepo.CacheArtifactName(runtime.GOOS))
	_, err = os.Stat(artifactPath)
	if err != nil {
		t.Fatalf("build root has no artifact after install: %v\n%s", err, stdout)
	}
	for _, path := range []string{shim, installedSkill} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("install did not publish %s: %v", path, err)
		}
	}

	// The shim resolves to the cached artifact and runs it.
	shimCommand := exportName
	var shimArgs []string
	if runtime.GOOS == "windows" {
		// Batch shims need the native command interpreter; CreateProcess
		// cannot launch a .cmd file directly.
		shimCommand = "cmd.exe"
		shimArgs = []string{"/d", "/c", "call", shim}
	}
	t.Setenv("PATH", filepath.Join(project, ".agents", "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
	resolved, err := exec.LookPath(exportName)
	if err != nil || resolved != shim {
		t.Fatalf("installed command is not on PATH: %s %v", resolved, err)
	}
	if command == "a2a" {
		shimArgs = append(shimArgs, "--help")
	}
	code, shimOut, shimErr := testcli.Run(t, project, env, "", shimCommand, shimArgs...)
	if code != 0 || strings.TrimSpace(shimOut) == "" {
		t.Fatalf("shim exit=%d stdout=%q stderr=%q", code, shimOut, shimErr)
	}
	t.Logf("shim %s exit=0 stdout=%q", filepath.Base(shim), shimOut)

}
