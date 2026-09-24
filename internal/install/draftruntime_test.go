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
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/scriptworker"
	"github.com/relux-works/curator/internal/sourcelock"
)

// This suite exercises the draft local runtime lane at the production
// entry (install.Project): scripts and dependency runtimes materialize
// from immutable local snapshots under source-v1 keys with the existing
// command, capability, protected-store and shim semantics, and refresh
// replaces from a new frozen snapshot instead of adopting live bytes.

const draftLocalCollectionPayload = `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["*"]}]}`

// writeDraftScriptSkill writes one local draft package with a runnable
// script command under a declared runtime root. mutateSpec adjusts the
// agent-skill.json object (extra dependencies, policies) per case.
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

func draftLocalInstall(t *testing.T, cfgHome, project string, dryRun bool) Result {
	t.Helper()
	cfg := draftTestConfig(cfgHome, t.TempDir())
	return Project(cfg, project, "test", Options{DryRun: dryRun, DraftSourcesV1: true, Platform: installPlatform()})
}

func draftLockedRuntimeKey(t *testing.T, project, name string) (lock *sourcelock.Lock, key string) {
	t.Helper()
	var err error
	lock, err = sourcelock.Read(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find(name)
	if !ok {
		t.Fatalf("lock misses %s", name)
	}
	digest, err := member.Package.Digest()
	if err != nil {
		t.Fatal(err)
	}
	key, err = runtimestore.SourceV1Key(digest)
	if err != nil {
		t.Fatal(err)
	}
	return lock, key
}

func assertNoLiveLinks(t *testing.T, root string) {
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

// TestDraftLocalRuntimeMaterializesFromFrozenSnapshot is the acceptance
// core: a provider skill with a runnable script and a consumer skill with
// a runnable script plus a skill dependency install for real from one
// immutable local snapshot, and every published artifact is verified.
func TestDraftLocalRuntimeMaterializesFromFrozenSnapshot(t *testing.T) {
	project, home, _ := draftProject(t, draftLocalCollectionPayload, nil)
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
	resolveDraftForInstall(t, project, home, draftLocalCollectionPayload)

	result := draftLocalInstall(t, home, project, false)
	if result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}

	lock, err := sourcelock.Read(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"provider", "consumer"} {
		member, ok := lock.Find(name)
		if !ok {
			t.Fatalf("lock misses %s", name)
		}
		_, key := draftLockedRuntimeKey(t, project, name)

		// The declared runtime root materialized under the frozen
		// package key with the exact snapshot bytes.
		var script string
		if name == "provider" {
			script = "ptool.sh"
		} else {
			script = "ctool.sh"
		}
		runtimeScript := filepath.Join(home, "runtime", name, key, "scripts", script)
		payload, err := os.ReadFile(runtimeScript)
		if err != nil {
			t.Fatalf("runtime script missing: %v", err)
		}
		if !strings.Contains(string(payload), "-ok") {
			t.Fatalf("runtime script has unexpected bytes: %q", payload)
		}

		// The protected tree is copied bytes, never live links.
		assertNoLiveLinks(t, filepath.Join(home, "runtime", name, key))

		// Context installed without the runtime root; the marker binds
		// the frozen package and lock generation.
		installed := filepath.Join(project, ".agents", "skills", name)
		if _, err := os.Stat(filepath.Join(installed, "SKILL.md")); err != nil {
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

		// The adapter mirror serves the installed context, not the live source.
		mirrorSkill := filepath.Join(project, ".claude", "skills", name, "SKILL.md")
		mirrorPayload, err := os.ReadFile(mirrorSkill)
		if err != nil {
			t.Fatalf("adapter mirror missing for %s: %v", name, err)
		}
		installedPayload, err := os.ReadFile(filepath.Join(installed, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		if string(mirrorPayload) != string(installedPayload) {
			t.Fatalf("adapter mirror for %s does not serve the installed context", name)
		}

		// The canonical shim reaches the frozen runtime, never the live tree.
		var shim string
		if name == "provider" {
			shim = filepath.Join(project, ".agents", "bin", shimName("ptool"))
		} else {
			shim = filepath.Join(project, ".agents", "bin", shimName("ctool"))
		}
		shimPayload, err := os.ReadFile(shim)
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

	if runtime.GOOS == "windows" {
		t.Skip("executes POSIX skill commands")
	}
	for command, want := range map[string]string{"ptool": "provider-ok\n", "ctool": "consumer-ok\n"} {
		out, err := exec.Command(filepath.Join(project, ".agents", "bin", command)).Output()
		if err != nil {
			t.Fatalf("shim %s did not run: %v", command, err)
		}
		if string(out) != want {
			t.Fatalf("shim %s output = %q, want %q", command, out, want)
		}
	}

	// A second install without changes is current through the v5 marker.
	again := draftLocalInstall(t, home, project, false)
	if again.Status != "ok" {
		t.Fatalf("reinstall = %+v", again)
	}
	if !strings.Contains(strings.Join(again.Messages, "\n"), "up-to-date") {
		t.Fatalf("reinstall did not report up-to-date installations: %q", again.Messages)
	}
}

// TestDraftLocalRefreshReplacesFrozenRuntime proves refresh semantics:
// live edits stay pinned until an explicit refresh reselects, and the
// refresh publishes a new frozen tree under a new package key instead of
// adopting live bytes or linking them.
func TestDraftLocalRefreshReplacesFrozenRuntime(t *testing.T) {
	project, home, _ := draftProject(t, draftLocalCollectionPayload, nil)
	script := filepath.Join(project, "skills", "provider", "scripts", "ptool.sh")
	writeDraftScriptSkill(t, filepath.Join(project, "skills", "provider"), "provider", "ptool", "#!/bin/sh\necho v1\n", nil)
	resolveDraftForInstall(t, project, home, draftLocalCollectionPayload)
	if result := draftLocalInstall(t, home, project, false); result.Status != "ok" {
		t.Fatalf("install v1 = %+v", result)
	}
	_, key1 := draftLockedRuntimeKey(t, project, "provider")
	runtimeScript1 := filepath.Join(home, "runtime", "provider", key1, "scripts", "ptool.sh")
	if payload, err := os.ReadFile(runtimeScript1); err != nil || !strings.Contains(string(payload), "echo v1") {
		t.Fatalf("runtime v1 = %q, %v", payload, err)
	}

	// A live edit without refresh stays pinned: the frozen install keeps
	// serving the locked bytes and the package key does not move.
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho v2-live\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if result := draftLocalInstall(t, home, project, false); result.Status != "ok" {
		t.Fatalf("pinned reinstall = %+v", result)
	}
	if payload, err := os.ReadFile(runtimeScript1); err != nil || !strings.Contains(string(payload), "echo v1") {
		t.Fatalf("pinned reinstall adopted live bytes: %q, %v", payload, err)
	}
	if _, key := draftLockedRuntimeKey(t, project, "provider"); key != key1 {
		t.Fatalf("pinned reinstall moved the package key %s -> %s", key1, key)
	}

	// Explicit refresh reselects: the lock generation moves and the next
	// install publishes the new frozen tree under a new key.
	resolveDraftForInstall(t, project, home, draftLocalCollectionPayload)
	lock, key2 := draftLockedRuntimeKey(t, project, "provider")
	_ = lock
	if key2 == key1 {
		t.Fatalf("refresh left the package key unchanged after a runtime-only edit")
	}
	if result := draftLocalInstall(t, home, project, false); result.Status != "ok" {
		t.Fatalf("install v2 = %+v", result)
	}
	runtimeScript2 := filepath.Join(home, "runtime", "provider", key2, "scripts", "ptool.sh")
	payload, err := os.ReadFile(runtimeScript2)
	if err != nil || !strings.Contains(string(payload), "echo v2-live") {
		t.Fatalf("refreshed runtime = %q, %v, want the reselected frozen bytes", payload, err)
	}
	assertNoLiveLinks(t, filepath.Join(home, "runtime", "provider", key2))
	// The refreshed install rebound the marker: it names the new frozen
	// package and lock generation, so status observes the runtime-only
	// change even though the projected context stayed identical.
	recorded := marker.Read(filepath.Join(project, ".agents", "skills", "provider"))
	if recorded == nil || recorded.SchemaVersion != marker.SchemaV5 {
		t.Fatalf("refreshed marker is not schema 5: %+v", recorded)
	}
	if recorded.LockSHA256 != lock.LockSHA256 {
		t.Fatalf("refreshed marker binds lock %s, want %s", recorded.LockSHA256, lock.LockSHA256)
	}
	member, ok := lock.Find("provider")
	if !ok {
		t.Fatalf("refreshed lock misses provider")
	}
	if recorded.Package == nil || recorded.Package.Snapshot != member.Package.Snapshot {
		t.Fatalf("refreshed marker does not bind the reselected package: %+v", recorded.Package)
	}
}

// TestDraftLocalTamperedFrozenRuntimeRefuses proves full-inventory
// authentication covers runtime roots on the mutating path: a tampered
// script fails source_snapshot_changed before any publication, even though
// the projected context hash excludes runtime bytes.
func TestDraftLocalTamperedFrozenRuntimeRefuses(t *testing.T) {
	project, home, _ := draftProject(t, draftLocalCollectionPayload, nil)
	writeDraftScriptSkill(t, filepath.Join(project, "skills", "provider"), "provider", "ptool", "#!/bin/sh\necho v1\n", nil)
	plan := resolveDraftForInstall(t, project, home, draftLocalCollectionPayload)
	if err := os.WriteFile(filepath.Join(plan.Frozen["provider"], "scripts", "ptool.sh"), []byte("#!/bin/sh\necho TAMPERED\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	result := draftLocalInstall(t, home, project, false)
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_snapshot_changed") {
		t.Fatalf("result = %+v, want source_snapshot_changed", result)
	}
	assertNoInstallSideEffects(t, project, home)
}

// TestDraftLocalMissingSnapshotRefuses proves a missing locked snapshot
// fails unavailable on the mutating path without publishing partial state.
func TestDraftLocalMissingSnapshotRefuses(t *testing.T) {
	project, home, _ := draftProject(t, draftLocalCollectionPayload, nil)
	writeDraftScriptSkill(t, filepath.Join(project, "skills", "provider"), "provider", "ptool", "#!/bin/sh\necho v1\n", nil)
	plan := resolveDraftForInstall(t, project, home, draftLocalCollectionPayload)
	for _, tree := range plan.Frozen {
		if err := os.RemoveAll(filepath.Dir(filepath.Dir(tree))); err != nil {
			t.Fatal(err)
		}
	}
	result := draftLocalInstall(t, home, project, false)
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_snapshot_unavailable") {
		t.Fatalf("result = %+v, want source_snapshot_unavailable", result)
	}
	assertNoInstallSideEffects(t, project, home)
}

func assertNoInstallSideEffects(t *testing.T, project, home string) {
	t.Helper()
	if _, err := os.Lstat(filepath.Join(project, ".agents")); !os.IsNotExist(err) {
		t.Fatalf("refused install published project state under .agents")
	}
	if _, err := os.Lstat(filepath.Join(home, "runtime")); !os.IsNotExist(err) {
		t.Fatalf("refused install published protected runtime state")
	}
}

// TestDraftLocalEnforcedCommandRefused proves script-worker enforcement
// stays mandatory for draft locals on the mutating path: when this host
// cannot provide a mandatory control, an enforced command fails before
// any publication. With a host that provides the controls the same draft
// installs its enforced command as a native launcher — the R3 contract
// change from universal refusal to host-conditional preflight.
func TestDraftLocalEnforcedCommandRefused(t *testing.T) {
	// Sequential by construction: the refusal phase forces the
	// process-global probe fault. Never add t.Parallel here.
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	setup := func(t *testing.T) (project, home string) {
		t.Helper()
		project, home, _ = draftProject(t, payload, map[string]string{"skills/review": "review"})
		dir := filepath.Join(project, "skills", "review")
		if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "scripts", "tool"), []byte("#!/bin/sh\necho tool\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		spec := map[string]any{
			"schema_version": 8,
			"capabilities":   map[string]any{},
			"runtime_roots":  []string{"scripts"},
			"commands": map[string]any{
				"tool": map[string]any{
					"type": "script", "unix_path": "scripts/tool", "win_path": "scripts/tool",
					"execution_policy": "script-worker-v1", "interpreter": "python3-v1",
				},
			},
		}
		raw, _ := json.Marshal(spec)
		if err := os.WriteFile(filepath.Join(dir, "agent-skill.json"), raw, 0o644); err != nil {
			t.Fatal(err)
		}
		resolveDraftForInstall(t, project, home, payload)
		return project, home
	}
	// Phase 1: the host cannot provide a mandatory control, so the
	// draft install refuses on the mutating path before any
	// publication.
	restore := scriptworker.OverrideScriptProbeFaultForTest(func(control string) error {
		if control == scriptworker.ScriptControlDescendantDomainTermination {
			return errInjectedProbeFailure
		}
		return nil
	})
	project, home := setup(t)
	result := draftLocalInstall(t, home, project, false)
	restore()
	if result.Status != "failed" ||
		!strings.Contains(strings.Join(result.Errors, ";"), "script_execution_control_unavailable") {
		t.Fatalf("result = %+v, want scriptpolicy refusal", result)
	}
	assertNoInstallSideEffects(t, project, home)
	// Phase 2: with a host that provides the controls, the same draft
	// installs its enforced command as a native launcher.
	project, home = setup(t)
	result = draftLocalInstall(t, home, project, false)
	if result.Status != "ok" {
		t.Fatalf("result = %+v, want a staged draft install", result)
	}
	sidecars, err := filepath.Glob(filepath.Join(project, ".agents", "bin", "*"+scriptworker.ShimSidecarSuffix))
	if err != nil || len(sidecars) == 0 {
		t.Fatalf("the draft install published no native launcher sidecar: %v", err)
	}
}

// TestDraftLocalMissingSystemCommandRefused proves system-command
// readiness stays mandatory for draft locals.
func TestDraftLocalMissingSystemCommandRefused(t *testing.T) {
	project, home, _ := draftProject(t, draftLocalCollectionPayload, nil)
	writeDraftScriptSkill(t, filepath.Join(project, "skills", "provider"), "provider", "ptool", "#!/bin/sh\necho v1\n", nil)
	writeDraftScriptSkill(t, filepath.Join(project, "skills", "consumer"), "consumer", "ctool", "#!/bin/sh\necho v1\n",
		func(spec map[string]any) {
			spec["dependencies"] = map[string]any{
				"commands": map[string]any{
					"sys": map[string]any{"type": "system", "command": "definitely-not-a-real-binary-xyz"},
				},
				"skills": map[string]any{},
			}
		})
	resolveDraftForInstall(t, project, home, draftLocalCollectionPayload)
	result := draftLocalInstall(t, home, project, false)
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "missing system command") {
		t.Fatalf("result = %+v, want missing system command refusal", result)
	}
	assertNoInstallSideEffects(t, project, home)
}

// TestDraftLocalSkillDependencyRefused proves skill dependency closure
// stays mandatory for draft locals: an absent provider and a provider
// without the required script command both fail before publication.
func TestDraftLocalSkillDependencyRefused(t *testing.T) {
	skillDepOn := func(skill, command string) func(map[string]any) {
		return func(spec map[string]any) {
			spec["dependencies"] = map[string]any{
				"commands": map[string]any{
					"dep": map[string]any{"type": "skill", "skill": skill, "command": command},
				},
				"skills": map[string]any{},
			}
		}
	}
	t.Run("absent-provider", func(t *testing.T) {
		project, home, _ := draftProject(t, draftLocalCollectionPayload, nil)
		writeDraftScriptSkill(t, filepath.Join(project, "skills", "consumer"), "consumer", "ctool", "#!/bin/sh\necho v1\n",
			skillDepOn("absent", "ptool"))
		resolveDraftForInstall(t, project, home, draftLocalCollectionPayload)
		result := draftLocalInstall(t, home, project, false)
		if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "missing skill dependency") {
			t.Fatalf("result = %+v, want missing skill dependency refusal", result)
		}
		assertNoInstallSideEffects(t, project, home)
	})
	t.Run("provider-without-script-command", func(t *testing.T) {
		project, home, _ := draftProject(t, draftLocalCollectionPayload, nil)
		writeDraftPackage(t, filepath.Join(project, "skills", "provider"), "provider")
		writeDraftScriptSkill(t, filepath.Join(project, "skills", "consumer"), "consumer", "ctool", "#!/bin/sh\necho v1\n",
			skillDepOn("provider", "ptool"))
		resolveDraftForInstall(t, project, home, draftLocalCollectionPayload)
		result := draftLocalInstall(t, home, project, false)
		if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "does not export a script command") {
			t.Fatalf("result = %+v, want script-command refusal", result)
		}
		assertNoInstallSideEffects(t, project, home)
	})
}

// TestDraftLocalInvalidCapabilitiesRefuseAtResolve proves capability
// declarations stay validated for draft locals: an unknown capability
// field fails the explicit resolve with source_member_invalid.
func TestDraftLocalInvalidCapabilitiesRefuseAtResolve(t *testing.T) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home, raw := draftProject(t, payload, map[string]string{"skills/review": "review"})
	spec := map[string]any{
		"schema_version": 4,
		"capabilities":   map[string]any{"bogus": true},
		"commands":       map[string]any{},
		"dependencies":   map[string]any{"skills": map[string]any{}},
	}
	encoded, _ := json.Marshal(spec)
	if err := os.WriteFile(filepath.Join(project, "skills", "review", "agent-skill.json"), encoded, 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := manifest.ParseBytesWithOptions(raw, filepath.Join(project, "Skillfile.json"), manifest.ParseOptions{DraftSourcesV1: true})
	if err != nil {
		t.Fatal(err)
	}
	_, err = closure.ResolveDraft(closure.DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: raw})
	if err == nil || !strings.Contains(err.Error(), "source_member_invalid") {
		t.Fatalf("resolve err = %v, want source_member_invalid", err)
	}
}

// TestDraftMarkerPackageDigestMatchesLockDigest pins the two package
// preimage implementations together: the marker v5 package and the lock
// member hash identically, so runtime store, status, repair, rollback and
// GC follow one key.
func TestDraftMarkerPackageDigestMatchesLockDigest(t *testing.T) {
	project, home, _ := draftProject(t, draftLocalCollectionPayload, nil)
	writeDraftScriptSkill(t, filepath.Join(project, "skills", "provider"), "provider", "ptool", "#!/bin/sh\necho v1\n", nil)
	resolveDraftForInstall(t, project, home, draftLocalCollectionPayload)
	lock, err := sourcelock.Read(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("provider")
	if !ok {
		t.Fatalf("lock misses provider")
	}
	lockDigest, err := member.Package.Digest()
	if err != nil {
		t.Fatal(err)
	}
	markerDigest, err := draftMarkerPackage(member.Package).Digest()
	if err != nil {
		t.Fatal(err)
	}
	if lockDigest != markerDigest {
		t.Fatalf("marker digest %s != lock digest %s", markerDigest, lockDigest)
	}
	if key, err := runtimestore.SourceV1Key(lockDigest); err != nil {
		t.Fatalf("lock digest does not name a runtime key: %v", err)
	} else if keys, err := draftRuntimeKeys(lock); err != nil || keys["provider"] != key {
		t.Fatalf("runtime keys = %v, %v, want provider=%s", keys, err, key)
	}
}

// TestDraftRuntimeKeysCoverAllArms proves the migrated boundary: every
// locked draft member takes a source-v1 key derived from its frozen
// package digest, never a commit-keyed leaf. The marker and the runtime
// key migrated together (marker schema 5 for every draft installation),
// so GC marks every draft runtime leaf from the recorded package.
func TestDraftRuntimeKeysCoverAllArms(t *testing.T) {
	localPkg, err := sourcelock.LocalPackage("sha256:" + strings.Repeat("ab", 32))
	if err != nil {
		t.Fatal(err)
	}
	gitPkg, err := sourcelock.NetworkGitPackage("example.org/kit",
		sourcelock.Commit{ObjectFormat: "sha1", Hex: strings.Repeat("cd", 20)}, "skills/review")
	if err != nil {
		t.Fatal(err)
	}
	configuredPkg, err := sourcelock.ConfiguredGitPackage("vendor/kit",
		sourcelock.Commit{ObjectFormat: "sha1", Hex: strings.Repeat("ef", 20)})
	if err != nil {
		t.Fatal(err)
	}
	lock := &sourcelock.Lock{Members: []sourcelock.Member{
		{Name: "local", Package: localPkg},
		{Name: "review", Package: gitPkg},
		{Name: "vendored", Package: configuredPkg},
	}}
	keys, err := draftRuntimeKeys(lock)
	if err != nil {
		t.Fatal(err)
	}
	for _, member := range lock.Members {
		digest, err := member.Package.Digest()
		if err != nil {
			t.Fatal(err)
		}
		want, err := runtimestore.SourceV1Key(digest)
		if err != nil {
			t.Fatal(err)
		}
		if keys[member.Name] != want {
			t.Fatalf("keys[%s] = %q, want source-v1 key %q", member.Name, keys[member.Name], want)
		}
	}
}
