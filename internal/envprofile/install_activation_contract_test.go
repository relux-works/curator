package envprofile

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/testruntimes"
)

type activationTreeEntry struct {
	Kind    string
	Mode    uint32
	ModTime int64
	Data    []byte
	Link    string
}

// snapshotActivationTree records bytes, links, modes, and mtimes without
// following links, so unchanged homes also prove that no rewrite occurred.
func snapshotActivationTree(t *testing.T, root string) map[string]activationTreeEntry {
	t.Helper()
	if _, err := os.Lstat(root); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		t.Fatalf("inspect snapshot root %s: %v", root, err)
	}
	entries := map[string]activationTreeEntry{}
	err := filepath.WalkDir(root, func(path string, _ os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		entry := activationTreeEntry{Mode: uint32(info.Mode()), ModTime: info.ModTime().UnixNano()}
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			entry.Kind = "symlink"
			entry.Link, err = os.Readlink(path)
		case info.IsDir():
			entry.Kind = "directory"
		case info.Mode().IsRegular():
			entry.Kind = "file"
			entry.Data, err = os.ReadFile(path)
		default:
			entry.Kind = "other"
		}
		if err != nil {
			return err
		}
		entries[rel] = entry
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot tree %s: %v", root, err)
	}
	return entries
}

func snapshotActivationHomes(t *testing.T, homes map[string]string) map[string]map[string]activationTreeEntry {
	t.Helper()
	result := make(map[string]map[string]activationTreeEntry, len(homes))
	for id, home := range homes {
		result[id] = snapshotActivationTree(t, home)
	}
	return result
}

// activationTreeStateDifferences compares restored content, links, modes, and
// entry sets. Compensating writes may update timestamps; the no-use tests use
// full snapshots to prove that no write occurred at all.
func activationTreeStateDifferences(before, after map[string]activationTreeEntry) []string {
	paths := make([]string, 0, len(before)+len(after))
	seen := map[string]bool{}
	for path := range before {
		seen[path] = true
		paths = append(paths, path)
	}
	for path := range after {
		if !seen[path] {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	var differences []string
	for _, path := range paths {
		oldEntry, oldOK := before[path]
		newEntry, newOK := after[path]
		if !oldOK || !newOK || oldEntry.Kind != newEntry.Kind || oldEntry.Mode != newEntry.Mode ||
			oldEntry.Link != newEntry.Link || string(oldEntry.Data) != string(newEntry.Data) {
			differences = append(differences, path)
		}
	}
	return differences
}

// setActivationRuntimes puts the shared zero-exit runtime stand-ins ahead of
// the package's TestMain PATH (activationRuntimeFallbackPath) for the rest of
// the test, so ordinary test tools stay reachable.
func setActivationRuntimes(t *testing.T, runtimes ...string) string {
	t.Helper()
	return testruntimes.Pin(t, activationRuntimeFallbackPath, runtimes...)
}

// setActivationRuntimesWithoutFallback isolates runtime-presence tests from
// adapter binaries installed on the host. Use this when absence is the
// contract under test; ordinary tools are not needed after fixture setup.
func setActivationRuntimesWithoutFallback(t *testing.T, runtimes ...string) string {
	t.Helper()
	return testruntimes.Pin(t, "", runtimes...)
}

func allActivationRuntimes() []string { return testruntimes.All() }

// This proves a no-use registration writes its source and lock while leaving
// native homes and current markers byte-for-byte unchanged, with and without
// a prior current. Missing runtimes stay warnings naming only active adapters.
func TestInstallWithoutUseLeavesHomesAndCurrentUntouched(t *testing.T) {
	for _, seededCurrent := range []bool{false, true} {
		name := "no-current"
		if seededCurrent {
			name = "existing-current"
		}
		t.Run(name, func(t *testing.T) {
			home := gitFixtureHome(t)
			homes := pinHomes(t)
			if seededCurrent {
				setActivationRuntimes(t, allActivationRuntimes()...)
				alpha := t.TempDir()
				writePackage(t, alpha, "alpha", "1.0.0", "alpha\n")
				if _, activated, _, err := Install(home, InstallOptions{Operand: alpha, Use: true}); err != nil || !activated {
					t.Fatalf("seed alpha activation = %t, %v", activated, err)
				}
			}
			beforeHomes := snapshotActivationHomes(t, homes)
			beforeCurrent := snapshotActivationTree(t, CurrentFile(home))
			beforeScoped := snapshotActivationTree(t, ScopedDir(home))
			emptyBin := t.TempDir()
			t.Setenv("PATH", emptyBin)
			source := t.TempDir()
			writePackage(t, source, "beta", "1.0.0", "beta\n")
			info, activated, _, err := Install(home, InstallOptions{Operand: source})
			if err != nil || activated {
				t.Fatalf("install without --use activated=%t, err=%v", activated, err)
			}
			if got := snapshotActivationHomes(t, homes); !reflect.DeepEqual(got, beforeHomes) {
				t.Fatal("install without --use changed an adapter home")
			}
			if got := snapshotActivationTree(t, CurrentFile(home)); !reflect.DeepEqual(got, beforeCurrent) {
				t.Fatal("install without --use changed the machine current marker")
			}
			if got := snapshotActivationTree(t, ScopedDir(home)); !reflect.DeepEqual(got, beforeScoped) {
				t.Fatal("install without --use changed scoped current markers")
			}
			if len(info.Warnings) == 0 {
				t.Fatal("missing scoped runtimes produced no warning")
			}
			for _, id := range []string{"claude_code", "codex_cli", "opencode", "pi"} {
				if !strings.Contains(strings.Join(info.Warnings, "\n"), id) {
					t.Fatalf("missing runtime warning does not name scoped adapter %s: %v", id, info.Warnings)
				}
			}
			// beta is current nowhere, so no stale-home warning belongs here.
			if strings.Contains(strings.Join(info.Warnings, "\n"), DiagHomeStale) {
				t.Fatalf("registration of a non-current profile warned about stale homes: %v", info.Warnings)
			}
			if _, err := readSource(home, "beta"); err != nil {
				t.Fatalf("registration missing after no-use install: %v", err)
			}
			if _, _, err := readLock(home, "beta"); err != nil {
				t.Fatalf("lock missing after no-use install: %v", err)
			}
		})
	}
}

// This proves an explicit path reinstall can refresh its lock without
// resynchronizing the profile's current adapter homes, and that it then warns
// with environment_home_stale that the current homes lag the new lock; a
// reinstall that leaves the lock unchanged warns nothing (negative control).
func TestInstallReinstallWithoutUseLeavesCurrentHomesUntouched(t *testing.T) {
	home := gitFixtureHome(t)
	homes := pinHomes(t)
	setActivationRuntimes(t, allActivationRuntimes()...)
	source := t.TempDir()
	writePackage(t, source, "alpha", "1.0.0", "before\n")
	if _, activated, _, err := Install(home, InstallOptions{Operand: source, Use: true}); err != nil || !activated {
		t.Fatalf("seed install activated=%t, err=%v", activated, err)
	}
	beforeHomes := snapshotActivationHomes(t, homes)
	beforeCurrent := snapshotActivationTree(t, CurrentFile(home))
	beforeScoped := snapshotActivationTree(t, ScopedDir(home))
	oldLock, oldHash, err := readLock(home, "alpha")
	if err != nil || oldLock == nil {
		t.Fatalf("read original lock: %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, "context", "a.md"), []byte("after\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, activated, updated, err := Install(home, InstallOptions{Operand: source})
	if err != nil || activated || !updated {
		t.Fatalf("reinstall activated=%t updated=%t, err=%v", activated, updated, err)
	}
	if info.LockHash == oldHash {
		t.Fatal("changed source did not update the persisted lock")
	}
	if got := snapshotActivationHomes(t, homes); !reflect.DeepEqual(got, beforeHomes) {
		t.Fatal("reinstall without --use changed an adapter home")
	}
	if got := snapshotActivationTree(t, CurrentFile(home)); !reflect.DeepEqual(got, beforeCurrent) {
		t.Fatal("reinstall without --use changed the machine current marker")
	}
	if got := snapshotActivationTree(t, ScopedDir(home)); !reflect.DeepEqual(got, beforeScoped) {
		t.Fatal("reinstall without --use changed scoped current markers")
	}
	stale := strings.Join(info.Warnings, "\n")
	if !strings.Contains(stale, DiagHomeStale) || !strings.Contains(stale, "alpha") || !strings.Contains(stale, "profile use alpha") {
		t.Fatalf("moved lock of the current profile did not warn that its homes are stale: %v", info.Warnings)
	}
	again, _, updated, err := Install(home, InstallOptions{Operand: source})
	if err != nil || !updated {
		t.Fatalf("unchanged reinstall updated=%t, err=%v", updated, err)
	}
	if again.LockHash != info.LockHash {
		t.Fatal("unchanged source moved the lock")
	}
	if strings.Contains(strings.Join(again.Warnings, "\n"), DiagHomeStale) {
		t.Fatalf("unchanged lock must not warn about stale homes: %v", again.Warnings)
	}
}

// This proves reinstall --use changes only machine-scope adapter homes and
// does not require or rewrite an adapter already assigned to an environment
// scope.
func TestInstallUseLeavesEnvironmentScopedAdapterUntouched(t *testing.T) {
	home := gitFixtureHome(t)
	homes := pinHomes(t)
	setActivationRuntimes(t, allActivationRuntimes()...)
	source := t.TempDir()
	writePackage(t, source, "alpha", "1.0.0", "before\n")
	if _, activated, _, err := Install(home, InstallOptions{Operand: source, Use: true}); err != nil || !activated {
		t.Fatalf("seed install activated=%t, err=%v", activated, err)
	}
	scopedProfile := t.TempDir()
	writePackage(t, scopedProfile, "scoped", "1.0.0", "scoped\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: scopedProfile}); err != nil {
		t.Fatalf("install scoped profile: %v", err)
	}
	if _, err := UseWithPolicy(home, "scoped", "codex_cli", "", false, Policy{}); err != nil {
		t.Fatalf("assign codex_cli scope: %v", err)
	}
	codexBefore := snapshotActivationTree(t, homes["codex_cli"])
	if err := os.WriteFile(filepath.Join(source, "context", "a.md"), []byte("after\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	setActivationRuntimes(t, "claude", "opencode", "pi")
	if _, activated, updated, err := Install(home, InstallOptions{Operand: source, Use: true}); err != nil || !activated || !updated {
		t.Fatalf("machine-scope reinstall activated=%t updated=%t, err=%v", activated, updated, err)
	}
	if got := snapshotActivationTree(t, homes["codex_cli"]); !reflect.DeepEqual(got, codexBefore) {
		t.Fatal("machine-scope reinstall changed the environment-scoped codex home")
	}
	scoped, err := ScopedCurrents(home)
	if err != nil || scoped["env:codex_cli"] != "scoped" {
		t.Fatalf("codex scoped current=%q, err=%v; want scoped", scoped["env:codex_cli"], err)
	}
	claude, err := os.ReadFile(filepath.Join(homes["claude_code"], "CLAUDE.md"))
	if err != nil || !strings.Contains(string(claude), "after") {
		t.Fatalf("machine-scope home was not updated: %s err=%v", claude, err)
	}
}

// This proves runtime refusal precedes every profile write, while selected
// environment scope checks only its adapter and machine scope omits adapters
// already owned by a scoped current.
func TestUseRuntimePreflightIsScopedAndReadOnlyOnRefusal(t *testing.T) {
	home := gitFixtureHome(t)
	homes := pinHomes(t)
	source := t.TempDir()
	writePackage(t, source, "alpha", "1.0.0", "alpha\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: source}); err != nil {
		t.Fatal(err)
	}
	beforeHomes := snapshotActivationHomes(t, homes)
	beforeManager := snapshotActivationTree(t, home)
	runtimeBin := setActivationRuntimesWithoutFallback(t, "claude", "opencode", "pi")
	if got := os.Getenv("PATH"); got != runtimeBin {
		t.Fatalf("runtime test PATH = %q, want only the fixture directory %q", got, runtimeBin)
	}
	for _, runtime := range []string{"claude", "opencode", "pi"} {
		if _, err := exec.LookPath(runtime); err != nil {
			t.Fatalf("positive runtime fixture %q is unavailable: %v", runtime, err)
		}
	}
	if _, err := exec.LookPath("codex"); err == nil {
		t.Fatal("missing-runtime fixture unexpectedly finds codex on its isolated PATH")
	}
	beta := t.TempDir()
	writePackage(t, beta, "beta", "1.0.0", "beta\n")
	if info, activated, updated, err := Install(home, InstallOptions{Operand: beta, Use: true}); err == nil ||
		!strings.Contains(err.Error(), DiagAdapterRuntimeMissing) || activated || updated || info.Name != "" {
		t.Fatalf("install --use with a missing scoped runtime = info=%+v activated=%t updated=%t err=%v", info, activated, updated, err)
	}
	if got := snapshotActivationTree(t, home); !reflect.DeepEqual(got, beforeManager) {
		t.Fatal("refused install --use changed manager state before writing registration or lock")
	}
	if _, err := readSource(home, "beta"); !isStateAbsent(err) {
		t.Fatalf("refused install --use left a registration: %v", err)
	}
	if _, _, err := readLock(home, "beta"); !isStateAbsent(err) {
		t.Fatalf("refused install --use left a lock: %v", err)
	}
	if _, err := UseWithPolicy(home, "alpha", "codex_cli", "", false, Policy{}); err == nil ||
		!strings.Contains(err.Error(), DiagAdapterRuntimeMissing) ||
		!strings.Contains(err.Error(), "codex_cli") || !strings.Contains(err.Error(), "codex") {
		t.Fatalf("missing scoped runtime error = %v", err)
	}
	if got := snapshotActivationHomes(t, homes); !reflect.DeepEqual(got, beforeHomes) {
		t.Fatal("refused scoped use changed an adapter home")
	}
	if got := snapshotActivationTree(t, home); !reflect.DeepEqual(got, beforeManager) {
		t.Fatal("refused scoped use wrote manager state")
	}
	setActivationRuntimesWithoutFallback(t, "codex")
	if _, err := UseWithPolicy(home, "alpha", "codex_cli", "", false, Policy{}); err != nil {
		t.Fatalf("codex-only scoped use failed with its runtime present: %v", err)
	}
	setActivationRuntimesWithoutFallback(t, "claude", "opencode", "pi")
	if _, err := UseWithPolicy(home, "alpha", "", "", false, Policy{}); err != nil {
		t.Fatalf("machine use refused a missing runtime outside machine scope: %v", err)
	}
}

// This proves a later adapter failure rolls back already-written homes and
// current markers while keeping the new install record and lock available.
// The positive control is the per-adapter result set: claude_code (ordered
// before the broken codex home), opencode, and pi report OK, so homes were
// written in this attempt and the byte-equal homes below are a rollback, not
// a mutant that writes nothing once any adapter is broken.
func TestInstallUseFailureRollsBackHomesButKeepsRegistrationAndLock(t *testing.T) {
	home := gitFixtureHome(t)
	homes := pinHomes(t)
	setActivationRuntimes(t, allActivationRuntimes()...)
	alpha := t.TempDir()
	writePackage(t, alpha, "alpha", "1.0.0", "alpha\n")
	if _, activated, _, err := Install(home, InstallOptions{Operand: alpha, Use: true}); err != nil || !activated {
		t.Fatalf("seed install activated=%t, err=%v", activated, err)
	}
	beforeHomes := snapshotActivationHomes(t, homes)
	beforeCurrent := snapshotActivationTree(t, CurrentFile(home))
	beforeScoped := snapshotActivationTree(t, ScopedDir(home))
	blocker := filepath.Join(t.TempDir(), "codex-home-is-a-file")
	if err := os.WriteFile(blocker, []byte("fixture blocker"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_HOME", blocker)
	beta := t.TempDir()
	writePackage(t, beta, "beta", "1.0.0", "beta\n")
	info, activated, _, err := Install(home, InstallOptions{Operand: beta, Use: true})
	if err == nil || !strings.Contains(err.Error(), DiagUsePartial) {
		t.Fatalf("install --use error = %v, want a surfaced activation failure", err)
	}
	if activated || info.Name != "beta" {
		t.Fatalf("failed activation result name=%q activated=%t", info.Name, activated)
	}
	written := map[string]bool{}
	for _, result := range info.Activation {
		written[result.Adapter] = result.OK
	}
	for _, id := range []string{"claude_code", "opencode", "pi"} {
		if !written[id] {
			t.Fatalf("positive control: adapter %s was not written before the failure was surfaced: %+v", id, info.Activation)
		}
	}
	if ok, attempted := written["codex_cli"]; !attempted || ok {
		t.Fatalf("codex_cli must fail on its regular-file home: %+v", info.Activation)
	}
	afterHomes := snapshotActivationHomes(t, homes)
	for id, before := range beforeHomes {
		if differences := activationTreeStateDifferences(before, afterHomes[id]); len(differences) != 0 {
			t.Fatalf("failed install --use did not restore adapter %s state: %v", id, differences)
		}
	}
	if got := snapshotActivationTree(t, CurrentFile(home)); !reflect.DeepEqual(got, beforeCurrent) {
		t.Fatal("failed install --use changed the machine current marker")
	}
	if got := snapshotActivationTree(t, ScopedDir(home)); !reflect.DeepEqual(got, beforeScoped) {
		t.Fatal("failed install --use changed scoped current markers")
	}
	if current, err := Current(home); err != nil || current != "alpha" {
		t.Fatalf("current after failed activation = %q, %v; want alpha", current, err)
	}
	if _, err := readSource(home, "beta"); err != nil {
		t.Fatalf("registration did not survive activation failure: %v", err)
	}
	if _, _, err := readLock(home, "beta"); err != nil {
		t.Fatalf("lock did not survive activation failure: %v", err)
	}
}

// This proves a first install --use on a clean machine, whose adapter homes
// did not exist before, rolls a late adapter failure back to absent homes:
// claude_code, codex_cli, and opencode were written (their OK rows are the
// positive control), the directories the attempt created are gone afterwards,
// no current marker appears, and the registration and lock remain. The
// regular-file blocker standing in for the pi home is left as it was.
func TestInstallUseFailureOnCleanMachineRemovesCreatedHomes(t *testing.T) {
	home := gitFixtureHome(t)
	homes := pinHomes(t)
	for id, dir := range homes {
		if _, err := os.Lstat(dir); !os.IsNotExist(err) {
			t.Fatalf("fixture home %s exists before the install: %v", id, err)
		}
	}
	blocker := filepath.Join(t.TempDir(), "pi-home-is-a-file")
	if err := os.WriteFile(blocker, []byte("fixture blocker"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PI_CODING_AGENT_DIR", blocker)
	blockerBefore := snapshotActivationTree(t, blocker)
	source := t.TempDir()
	writePackage(t, source, "alpha", "1.0.0", "alpha\n")
	info, activated, _, err := Install(home, InstallOptions{Operand: source, Use: true})
	if err == nil || !strings.Contains(err.Error(), DiagUsePartial) {
		t.Fatalf("install --use error = %v, want a surfaced activation failure", err)
	}
	if activated || info.Name != "alpha" {
		t.Fatalf("failed activation result name=%q activated=%t", info.Name, activated)
	}
	written := map[string]bool{}
	for _, result := range info.Activation {
		written[result.Adapter] = result.OK
	}
	for _, id := range []string{"claude_code", "codex_cli", "opencode"} {
		if !written[id] {
			t.Fatalf("positive control: adapter %s was not written before pi failed: %+v", id, info.Activation)
		}
	}
	if ok, attempted := written["pi"]; !attempted || ok {
		t.Fatalf("pi must fail on its regular-file home: %+v", info.Activation)
	}
	for _, id := range []string{"claude_code", "codex_cli", "opencode"} {
		if _, err := os.Lstat(homes[id]); !os.IsNotExist(err) {
			t.Fatalf("rollback left the created %s home in place: %v", id, err)
		}
	}
	if differences := activationTreeStateDifferences(blockerBefore, snapshotActivationTree(t, blocker)); len(differences) != 0 {
		t.Fatalf("rollback changed the pi blocker: %v", differences)
	}
	if current, err := Current(home); err != nil || current != "" {
		t.Fatalf("current after failed first activation = %q, %v; want none", current, err)
	}
	if got := snapshotActivationTree(t, ScopedDir(home)); got != nil {
		t.Fatalf("failed first activation wrote scoped current markers: %v", got)
	}
	if _, err := readSource(home, "alpha"); err != nil {
		t.Fatalf("registration did not survive activation failure: %v", err)
	}
	if _, _, err := readLock(home, "alpha"); err != nil {
		t.Fatalf("lock did not survive activation failure: %v", err)
	}
}
