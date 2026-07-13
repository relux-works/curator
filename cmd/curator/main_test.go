package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
)

func TestUsageEnumeratesDocumentedCommands(t *testing.T) {
	for _, command := range []string{
		"bootstrap", "init", "add", "remove", "install", "update", "upgrade",
		"status", "list", "project", "config", "skill", "global", "hybrid",
		"audit", "gc", "shell-init", "ui",
	} {
		if !strings.Contains(usage, command) {
			t.Fatalf("usage does not enumerate %q", command)
		}
	}
}

func TestRunVersionExitsZero(t *testing.T) {
	if code := run([]string{"--version"}); code != 0 {
		t.Fatalf("run(--version) = %d, want 0", code)
	}
}

func TestRunNoArgsPrintsUsage(t *testing.T) {
	if code := run(nil); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	if code := run([]string{"frobnicate"}); code != 2 {
		t.Fatalf("run(frobnicate) = %d, want 2", code)
	}
}

func TestShellInitPrintsHooks(t *testing.T) {
	for _, shellName := range []string{"zsh", "bash", "powershell"} {
		if code := run([]string{"shell-init", shellName}); code != 0 {
			t.Fatalf("shell-init %s = %d", shellName, code)
		}
	}
	if code := run([]string{"shell-init", "fish"}); code != 2 {
		t.Fatalf("unsupported shell must be usage error")
	}
}

func TestSkillCheckOnTempDir(t *testing.T) {
	// an empty directory fails validation (missing SKILL.md)
	dir := t.TempDir()
	if code := run([]string{"skill", "check", dir}); code != 1 {
		t.Fatalf("skill check on empty dir = %d, want 1", code)
	}
	if code := run([]string{"skill", "check", dir, "--json"}); code != 1 {
		t.Fatalf("skill check with trailing --json = %d, want 1", code)
	}
}

func TestInstallFlagsAcceptTrailingOptions(t *testing.T) {
	opts, positional, all, auditMode, err := installFlags([]string{"project-a", "--dry-run", "--strict-tags", "--audit", "strict"})
	if err != nil {
		t.Fatal(err)
	}
	if len(positional) != 1 || positional[0] != "project-a" {
		t.Fatalf("positional = %v", positional)
	}
	if !opts.DryRun || !opts.StrictTags || all || auditMode != "strict" {
		t.Fatalf("parsed flags: opts=%+v all=%v audit=%q", opts, all, auditMode)
	}
}

func TestInstallAuditFlagAcceptsOptionalMode(t *testing.T) {
	_, positional, _, auditMode, err := installFlags([]string{"--audit", "project-a"})
	if err != nil || auditMode != "advisory" || len(positional) != 1 || positional[0] != "project-a" {
		t.Fatalf("bare --audit: positional=%v mode=%q err=%v", positional, auditMode, err)
	}
	_, positional, _, auditMode, err = installFlags([]string{"project-a", "--audit", "strict"})
	if err != nil || auditMode != "strict" || len(positional) != 1 || positional[0] != "project-a" {
		t.Fatalf("strict --audit: positional=%v mode=%q err=%v", positional, auditMode, err)
	}
}

func TestSelectProjectTargetsUsesAliasesAndStableAllOrder(t *testing.T) {
	cfg := &config.Config{Projects: map[string]config.Project{
		"zeta":  {Path: "/work/zeta"},
		"alpha": {Path: "/work/alpha"},
	}}
	targets, err := selectProjectTargets(cfg, []string{"alpha"}, false)
	if err != nil || len(targets) != 1 || targets[0].Root != "/work/alpha" || targets[0].Alias != "alpha" {
		t.Fatalf("alias targets = %+v, %v", targets, err)
	}
	targets, err = selectProjectTargets(cfg, nil, true)
	if err != nil || len(targets) != 2 || targets[0].Alias != "alpha" || targets[1].Alias != "zeta" {
		t.Fatalf("all targets = %+v, %v", targets, err)
	}
}

func TestStatusDriftDetectsContentTampering(t *testing.T) {
	project := t.TempDir()
	skillsRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(
		`{"schema_version":1,"skills":[{"name":"skill-a","tag":"v1"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	installed := filepath.Join(project, ".agents", "skills", "skill-a")
	if err := os.MkdirAll(installed, 0o755); err != nil {
		t.Fatal(err)
	}
	skillPath := filepath.Join(installed, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	hash, err := hashing.ContentSHA256(installed, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := marker.Write(installed, &marker.Marker{
		Name: "skill-a", Source: "skill-a", RefKind: "tag", Ref: "v1",
		Commit: "0123456789abcdef0123456789abcdef01234567", ContentSHA256: hash,
		InstalledAt: "2026-07-13T00:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skillPath, []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	drift := statusDrift(&config.Config{SkillsRoot: skillsRoot}, project)
	if drift["skill-a"] != "content-drift" {
		t.Fatalf("drift = %v", drift)
	}
}

func TestBootstrapAndProjectCommands(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("CURATOR_CONFIG", configPath)
	skillsRoot := t.TempDir()
	if code := run([]string{"bootstrap", "--non-interactive", "--skills-root", skillsRoot}); code != exitOK {
		t.Fatalf("bootstrap = %d", code)
	}
	project := t.TempDir()
	if code := run([]string{"project", "add", "app", project, "--agents", "codex_cli"}); code != exitOK {
		t.Fatalf("project add = %d", code)
	}
	if _, err := os.Stat(filepath.Join(project, manifest.Name)); err != nil {
		t.Fatalf("project manifest missing: %v", err)
	}
	if code := run([]string{"project", "resolve", "app"}); code != exitOK {
		t.Fatalf("project resolve = %d", code)
	}
	cfg, err := config.Load(configPath, nil)
	if err != nil || cfg.Projects["app"].Path != project {
		t.Fatalf("saved project: cfg=%+v err=%v", cfg, err)
	}
}

func TestCLIEndToEndInstallStatusAndTamperCheck(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "home", "config.json")
	t.Setenv("CURATOR_CONFIG", configPath)
	skillsRoot := filepath.Join(root, "skills")
	project := filepath.Join(root, "project")
	if err := os.MkdirAll(filepath.Join(skillsRoot, "skill-a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}

	skillRepo := filepath.Join(skillsRoot, "skill-a")
	runGit(t, skillRepo, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(skillRepo, "SKILL.md"), []byte("---\nname: skill-a\n---\n# Skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, skillRepo, "add", ".")
	runGit(t, skillRepo, "commit", "-qm", "initial skill")
	runGit(t, skillRepo, "tag", "v1")

	runGit(t, project, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n.codex/skills/\nSkillfile.dev.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, manifest.Name), []byte(
		`{"schema_version":1,"agents":["codex_cli"],"skills":[{"name":"skill-a","tag":"v1"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := config.Bootstrap(configPath, skillsRoot, "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	if err := config.AddProject(configPath, "app", project, []string{"codex_cli"}); err != nil {
		t.Fatal(err)
	}

	if code := run([]string{"install", "app"}); code != exitOK {
		t.Fatalf("install = %d", code)
	}
	installedSkill := filepath.Join(project, ".agents", "skills", "skill-a", "SKILL.md")
	if _, err := os.Stat(installedSkill); err != nil {
		t.Fatalf("installed skill missing: %v", err)
	}
	if code := run([]string{"status", "app", "--check"}); code != exitOK {
		t.Fatalf("clean status = %d", code)
	}
	if err := os.WriteFile(installedSkill, []byte("tampered\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"status", "app", "--check"}); code != exitFail {
		t.Fatalf("tampered status = %d, want %d", code, exitFail)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	command.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}
