package crossconformance

// Local skill + script + dependency end to end through the compiled CLI.
//
// The install package proves this shape through the Go API
// (TestDraftLocalRuntimeMaterializesFromFrozenSnapshot). This test proves
// the same provider/consumer pair — runnable scripts plus a skill-command
// dependency — through the production binary: `project resolve` binds the
// frozen lock, `install` publishes context, markers, adapter mirrors,
// protected runtimes, and shims, and the shims execute the frozen bytes.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/testcli"
)

// writeDraftScriptSkill writes one local draft package with a runnable
// script command under a declared runtime root, mirroring the install
// package's fixture shape (policy paragraph, schema-4 spec, unix/win
// paths, executable script).
func writeDraftScriptSkill(t *testing.T, dir, name, command, scriptBody string, mutateSpec func(map[string]any)) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	skillMD := "---\nname: " + name + "\ndescription: Test\n---\n# " + name + "\n" +
		"Resolve commands via project .agents/bin lookup, then manager global/bin fallback, " +
		"then a validated bare command: command -v on POSIX and Get-Command on PowerShell.\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(skillMD), 0o644); err != nil {
		t.Fatal(err)
	}
	spec := map[string]any{
		"schema_version": 4,
		"capabilities":   map[string]any{},
		"commands": map[string]any{
			command: map[string]any{"type": "script", "unix_path": "scripts/" + command + ".sh", "win_path": "scripts/" + command + ".sh"},
		},
		"dependencies":  map[string]any{"skills": map[string]any{}},
		"runtime_roots": []string{"scripts"},
	}
	if mutateSpec != nil {
		mutateSpec(spec)
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "agent-skill.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scripts", command+".sh"), []byte(scriptBody), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "references", "info.md"), []byte("context"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func draftCLIShimName(command string) string {
	if draftPlatform() == "windows" {
		return command + ".cmd"
	}
	return command
}

func assertDraftNoLiveLinks(t *testing.T, root string) {
	t.Helper()
	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			t.Fatalf("protected runtime tree holds a link (must be frozen bytes, never live links): %s", path)
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			t.Fatalf("protected runtime tree holds a special file: %s", path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestDraftSourcesCLILocalSkillScriptDependencies(t *testing.T) {
	root := t.TempDir()
	configPath, project, home := setupCLIProject(t, root)
	writeDraftScriptSkill(t, filepath.Join(project, "skills", "provider"), "provider", "ptool", "#!/bin/sh\necho provider-ok\n", nil)
	writeDraftScriptSkill(t, filepath.Join(project, "skills", "consumer"), "consumer", "ctool", "#!/bin/sh\necho consumer-ok\n",
		func(spec map[string]any) {
			spec["dependencies"] = map[string]any{
				"commands": map[string]any{
					"dep": map[string]any{"type": "skill", "skill": "provider", "command": "ptool"},
				},
				"skills": map[string]any{},
			}
		})
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["*"]}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := runCurator(t, home, configPath, nil, "project", "resolve", "app"); code != 0 {
		t.Fatalf("resolve = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := runCurator(t, home, configPath, nil, "install", "app"); code != 0 {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	tools := map[string]string{"provider": "ptool", "consumer": "ctool"}
	for _, name := range []string{"provider", "consumer"} {
		member, ok := lock.Find(name)
		if !ok {
			t.Fatalf("lock misses %s", name)
		}
		digest, err := member.Package.Digest()
		if err != nil {
			t.Fatal(err)
		}
		key, err := runtimestore.SourceV1Key(digest)
		if err != nil {
			t.Fatal(err)
		}
		script := filepath.Join(home, "runtime", name, key, "scripts", tools[name]+".sh")
		scriptPayload, err := os.ReadFile(script)
		if err != nil {
			t.Fatalf("runtime script missing for %s: %v", name, err)
		}
		if !strings.Contains(string(scriptPayload), "-ok") {
			t.Fatalf("runtime script for %s has unexpected bytes: %q", name, scriptPayload)
		}
		assertDraftNoLiveLinks(t, filepath.Join(home, "runtime", name, key))
		installed := filepath.Join(project, ".agents", "skills", name)
		installedPayload, err := os.ReadFile(filepath.Join(installed, "SKILL.md"))
		if err != nil {
			t.Fatalf("context missing for %s: %v", name, err)
		}
		if _, err := os.Lstat(filepath.Join(installed, "scripts")); !os.IsNotExist(err) {
			t.Fatalf("runtime root leaked into installed context for %s", name)
		}
		recorded := marker.Read(installed)
		if recorded == nil || recorded.SchemaVersion != marker.SchemaV5 {
			t.Fatalf("marker for %s is not schema 5: %+v", name, recorded)
		}
		if recorded.Package == nil || recorded.Package.Snapshot != member.Package.Snapshot ||
			recorded.LockSHA256 != lock.LockSHA256 {
			t.Fatalf("marker for %s does not bind the frozen package: %+v", name, recorded)
		}
		mirrorPayload, err := os.ReadFile(filepath.Join(project, adapters.AgentPaths["codex_cli"], name, "SKILL.md"))
		if err != nil {
			t.Fatalf("adapter mirror missing for %s: %v", name, err)
		}
		if string(mirrorPayload) != string(installedPayload) {
			t.Fatalf("adapter mirror for %s does not serve the installed context", name)
		}
		shimPayload, err := os.ReadFile(filepath.Join(project, ".agents", "bin", draftCLIShimName(tools[name])))
		if err != nil {
			t.Fatalf("shim missing for %s: %v", name, err)
		}
		if !strings.Contains(string(shimPayload), filepath.Join("runtime", name)) {
			t.Fatalf("shim for %s does not reach the protected runtime store:\n%s", name, shimPayload)
		}
		if strings.Contains(string(shimPayload), filepath.Join("skills", name)) {
			t.Fatalf("shim for %s points at the live source tree:\n%s", name, shimPayload)
		}
	}
	t.Run("shims execute the frozen runtime", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("executes POSIX skill commands")
		}
		for command, want := range map[string]string{"ptool": "provider-ok\n", "ctool": "consumer-ok\n"} {
			got := testcli.Output(t, "", nil, filepath.Join(project, ".agents", "bin", command))
			if got+"\n" != want {
				t.Fatalf("shim %s printed %q, want %q", command, got, want)
			}
		}
	})
}

func TestDraftSourcesCLIDirectoryManifestDependencyInstallRefreshAndAudit(t *testing.T) {
	root := t.TempDir()
	configPath, project, home := setupCLIProject(t, root)
	skillsRoot := filepath.Join(root, "skills-root")
	providerRepo := filepath.Join(skillsRoot, "backend")
	dependencyGit := "https://github.com/example/role-skills.git"
	writeManifestDependencySkill(t, filepath.Join(providerRepo, "skills", "backend"), "backend", 9, map[string]map[string]any{})
	writeManifestDependencySkill(t, filepath.Join(providerRepo, "skills", "frontend"), "frontend", 9, map[string]map[string]any{})
	writeManifestDependencySkill(t, filepath.Join(providerRepo, "skills", "shared"), "shared", 4, map[string]map[string]any{})
	runDraftGit(t, providerRepo, "init", "-q", "-b", "main")
	runDraftGit(t, providerRepo, "add", ".")
	runDraftGit(t, providerRepo, "commit", "-qm", "base role packages")
	baseCommit := strings.TrimSpace(draftGitOutput(t, providerRepo, "rev-parse", "HEAD"))
	sharedRequirement := map[string]map[string]any{
		"shared": {
			"git":       dependencyGit,
			"ref":       map[string]any{"kind": "revision", "value": baseCommit},
			"directory": "skills/shared",
		},
	}
	writeManifestDependencySkill(t, filepath.Join(providerRepo, "skills", "backend"), "backend", 9, sharedRequirement)
	writeManifestDependencySkill(t, filepath.Join(providerRepo, "skills", "frontend"), "frontend", 9, sharedRequirement)
	runDraftGit(t, providerRepo, "add", ".")
	runDraftGit(t, providerRepo, "commit", "-qm", "role dependencies")
	roleCommit := strings.TrimSpace(draftGitOutput(t, providerRepo, "rev-parse", "HEAD"))

	appDir := filepath.Join(project, "skills", "app")
	writeManifestDependencySkill(t, appDir, "app", 9, map[string]map[string]any{
		"backend": {
			"git":       dependencyGit,
			"ref":       map[string]any{"kind": "revision", "value": roleCommit},
			"directory": "skills/backend",
		},
		"frontend": {
			"git":       dependencyGit,
			"ref":       map[string]any{"kind": "revision", "value": roleCommit},
			"directory": "skills/frontend",
		},
	})
	skillfile := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"name":"app","from":"s","directory":"skills/app"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(skillfile), 0o644); err != nil {
		t.Fatal(err)
	}

	configPayload, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var configDocument map[string]any
	if err := json.Unmarshal(configPayload, &configDocument); err != nil {
		t.Fatal(err)
	}
	configDocument["audit"] = map[string]any{
		"enabled": true, "mode": "advisory", "backend": "null", "registry_policy": "advisory",
	}
	configPayload, err = json.MarshalIndent(configDocument, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, configPayload, 0o600); err != nil {
		t.Fatal(err)
	}

	if code, stdout, stderr := runCurator(t, home, configPath, nil, "project", "resolve", "app"); code != 0 {
		t.Fatalf("project resolve = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	lockPath := filepath.Join(project, "Skillfile.lock.json")
	lock, err := sourcelock.Read(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	assertDirectoryDependencyLock := func() {
		t.Helper()
		for _, name := range []string{"backend", "frontend", "shared"} {
			member, ok := lock.Find(name)
			if !ok || member.Package.Kind != sourcelock.KindNetworkGit || member.Directory != "skills/"+name || member.Package.Directory != "skills/"+name {
				t.Fatalf("locked dependency %s does not bind selected directory: %+v", name, member)
			}
		}
	}
	assertDirectoryDependencyLock()

	for _, operation := range [][]string{{"install", "app"}, {"project", "refresh", "app"}, {"install", "app"}} {
		if code, stdout, stderr := runCurator(t, home, configPath, nil, operation...); code != 0 {
			t.Fatalf("%v = %d\nstdout:\n%s\nstderr:\n%s", operation, code, stdout, stderr)
		}
		if operation[0] == "project" {
			lock, err = sourcelock.Read(lockPath)
			if err != nil {
				t.Fatal(err)
			}
			assertDirectoryDependencyLock()
		}
	}

	for _, name := range []string{"backend", "frontend", "shared"} {
		installedPath := filepath.Join(project, ".agents", "skills", name)
		installedSkill, err := os.ReadFile(filepath.Join(installedPath, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(installedSkill), name) {
			t.Fatalf("installed context for %s does not come from selected folder: %s", name, installedSkill)
		}
		installedMarker := marker.Read(installedPath)
		if installedMarker == nil || installedMarker.Package == nil || installedMarker.Package.Directory != "skills/"+name {
			t.Fatalf("install marker for %s lost package directory: %+v", name, installedMarker)
		}
	}

	objects, err := filepath.Glob(filepath.Join(home, "source-audit", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	auditedDirectories := map[string]bool{}
	for _, path := range objects {
		if strings.HasSuffix(path, ".report.json") {
			continue
		}
		payload, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		object, err := audit.ParseSourceAudit(payload)
		if err != nil {
			t.Fatalf("parse source audit %s: %v", path, err)
		}
		if object.Package.Kind == sourcelock.KindNetworkGit {
			auditedDirectories[object.Package.Directory] = true
		}
	}
	for _, directory := range []string{"skills/backend", "skills/frontend", "skills/shared"} {
		if !auditedDirectories[directory] {
			t.Fatalf("source-audit records omit selected package directory: %v", auditedDirectories)
		}
	}
}
