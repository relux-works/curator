#!/usr/bin/env python3
"""Reconstruct the pre-patch cmd/curator/status_test.go by reversing every edit.

The file is untracked in this worktree, so `git diff` has no baseline for it.
Reversing the exact edit pairs and checking the reconstruction against the
recorded pre-patch digest gives one, and the resulting file is only used to
render a diff.
"""
import hashlib
import pathlib
import sys

AFTER = pathlib.Path(sys.argv[1])
OUT = pathlib.Path(sys.argv[2])
WANT_BEFORE = "4c2339dd36854cd51baf28c92c480b4341fa398c10118f41d44b8d29e845afbe"

# (new, old) pairs, in the order they were applied.
EDITS = []

EDITS.append((r'''// compiledProjectFixture is one real compiled installation: the project the
// operator declared, the manager home that owns the protected build cache, and
// the installed skill whose marker records the compiled state.
type compiledProjectFixture struct {
	project   string
	home      string
	installed string
}

// newInstalledCompiledProject builds that fixture and installs it once, through
// the real command path.
//
// Every surface below starts from the same thing: one installed compiled
// command whose protected entry and install marker agree. Deriving it costs a
// real compilation of the vendored module, so it is derived once and handed to
// each case rather than rebuilt per test.
func newInstalledCompiledProject(t *testing.T) compiledProjectFixture {
	t.Helper()
	project, home := compiledProject(t)
	if code, stdout, stderr := capture(t, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	return compiledProjectFixture{
		project:   project,
		home:      home,
		installed: filepath.Join(project, ".agents", "skills", "build-skill"),
	}
}

// TestCompiledProjectStatusRepairRollbackRecovery drives the whole compiled
// project lifecycle over one installed fixture: currentness reporting, corrupt
// cache repair by install and by upgrade, the rollback and recovery of a commit
// that failed after a real publication, protected cache state that moves during
// a check, and the repair of cache bytes outside a provable boundary.
//
// The cases run sequentially and share live state deliberately, so each one
// re-reads the marker, cache, and plan evidence it asserts against instead of
// carrying a baseline in, and each returns the fixture to a current state
// before the next one opens its own window.
func TestCompiledProjectStatusRepairRollbackRecovery(t *testing.T) {
	fixture := newInstalledCompiledProject(t)

	t.Run("status reports compiled currentness and fails check", func(t *testing.T) {
		assertCompiledCurrentnessAndFailedCheck(t, fixture)
	})
	t.Run("install and upgrade repair corrupt compiled state", func(t *testing.T) {
		assertCorruptCompiledStateIsRepaired(t, fixture)
	})
	t.Run("install and upgrade restore the cache when the commit fails", func(t *testing.T) {
		assertTheCacheIsRestoredWhenTheCommitFails(t, fixture)
	})
	t.Run("status reports protected cache state that moved during the check", func(t *testing.T) {
		assertProtectedCacheStateThatMovedDuringTheCheck(t, fixture)
	})
	t.Run("install repairs untrusted compiled state and preserves the old install", func(t *testing.T) {
		assertUntrustedCompiledStateIsRepaired(t, fixture)
	})
}

// assertCompiledCurrentnessAndFailedCheck is the end-to-end proof of the
// compiled currentness surface: a real installation reports every planned
// command as current, and each independent way that state can stop being
// current produces its own stable code and a non-zero `status --check`.
func assertCompiledCurrentnessAndFailedCheck(t *testing.T, fixture compiledProjectFixture) {
	project, home, installed := fixture.project, fixture.home, fixture.installed

	code, stdout, stderr := capture(t, "status", "app", "--json")''',
r'''// TestStatusReportsCompiledCurrentnessAndFailsCheck is the end-to-end proof of
// the compiled currentness surface: a real installation reports every planned
// command as current, and each independent way that state can stop being
// current produces its own stable code and a non-zero `status --check`.
func TestStatusReportsCompiledCurrentnessAndFailsCheck(t *testing.T) {
	project, home := compiledProject(t)
	if code, stdout, stderr := capture(t, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	installed := filepath.Join(project, ".agents", "skills", "build-skill")

	code, stdout, stderr := capture(t, "status", "app", "--json")'''))

EDITS.append((r'''// assertCorruptCompiledStateIsRepaired proves the reconciliation path for a
// protected entry that cannot be interpreted at all: corrupt receipt bytes and
// corrupt artifact bytes are rebuilt by install and by upgrade, only after
// every gate has passed, and a run that fails leaves the previous installation
// and the live cache exactly as they were.
func assertCorruptCompiledStateIsRepaired(t *testing.T, fixture compiledProjectFixture) {
	home, installed := fixture.home, fixture.installed

	for _, command := range []string{"install", "upgrade"} {
		t.Run(command, func(t *testing.T) {
			for _, corruption := range []struct {
				name    string
				corrupt func(t *testing.T, home string)
			}{
				{"receipt bytes", corruptCacheReceipt},
				{"artifact bytes", corruptCacheArtifact},
			} {
				t.Run(corruption.name, func(t *testing.T) {
					// The shared fixture is current when this case starts, and a
					// preceding repair legitimately republished the entry the marker
					// names, so the compiled state every assertion below compares
					// against is re-read here instead of being carried in.
					before := marker.Read(installed)
					if before == nil || len(before.Builds) != 1 {
						t.Fatalf("the installation this case starts from records no compiled state: %+v", before)
					}

					corruption.corrupt(t, home)''',
r'''// TestInstallAndUpgradeRepairCorruptCompiledState proves the reconciliation
// path for a protected entry that cannot be interpreted at all: corrupt
// receipt bytes and corrupt artifact bytes are rebuilt by install and by
// upgrade, only after every gate has passed, and a run that fails leaves the
// previous installation and the live cache exactly as they were.
func TestInstallAndUpgradeRepairCorruptCompiledState(t *testing.T) {
	for _, command := range []string{"install", "upgrade"} {
		t.Run(command, func(t *testing.T) {
			project, home := compiledProject(t)
			if code, stdout, stderr := capture(t, "install", "app"); code != exitOK {
				t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
			}
			installed := filepath.Join(project, ".agents", "skills", "build-skill")
			before := marker.Read(installed)
			if before == nil || len(before.Builds) != 1 {
				t.Fatalf("first install did not record compiled state: %+v", before)
			}

			for _, corruption := range []struct {
				name    string
				corrupt func(t *testing.T, home string)
			}{
				{"receipt bytes", corruptCacheReceipt},
				{"artifact bytes", corruptCacheArtifact},
			} {
				t.Run(corruption.name, func(t *testing.T) {
					corruption.corrupt(t, home)'''))

EDITS.append((r'''// assertTheCacheIsRestoredWhenTheCommitFails is the other half of the repair
// contract, driven end to end through the real commands.
//''',
r'''// TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails is the other half of
// the repair contract, driven end to end through the real commands.
//'''))

EDITS.append((r'''func assertTheCacheIsRestoredWhenTheCommitFails(t *testing.T, fixture compiledProjectFixture) {
	project, home, installed := fixture.project, fixture.home, fixture.installed

	for _, command := range []string{"install", "upgrade"} {
		t.Run(command, func(t *testing.T) {
			// The shared fixture is current when this case starts, and a preceding
			// repair legitimately republished the entry the marker names, so the
			// compiled state this case must find unchanged is re-read here.
			before := marker.Read(installed)
			if before == nil || len(before.Builds) != 1 {
				t.Fatalf("the installation this case starts from records no compiled state: %+v", before)
			}

			corruptCacheArtifact(t, home)
			// A privileged process ignores the directory mode this case fails the
			// commit with, and skips below with the cache already corrupted. The
			// next case is handed a current fixture on that path too.
			t.Cleanup(func() {
				if t.Skipped() {
					reinstall(t)
				}
			})
			if code, stdout, _ := capture(t, "status", "app", "--json"); code != exitOK ||
				decodeStatus(t, stdout).Builds[0].State != buildCorruptCache {
				t.Fatalf("status did not report corrupt compiled state:\n%s", stdout)
			}
			corruptArtifact := liveArtifactBytes(t, home)
			installedBefore := installedFingerprint(t, project, home)
			// Withdrawn entries are never deleted, only collected by the ordinary
			// sweep, so a shared cache root still holds what earlier cases withdrew.
			// The count this case owns is therefore measured against what it found.
			withdrawnBefore := quarantinedEntries(t, home)

			store := filepath.Join(project, ".agents", "skills")''',
r'''func TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails(t *testing.T) {
	for _, command := range []string{"install", "upgrade"} {
		t.Run(command, func(t *testing.T) {
			project, home := compiledProject(t)
			if code, stdout, stderr := capture(t, "install", "app"); code != exitOK {
				t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
			}
			installed := filepath.Join(project, ".agents", "skills", "build-skill")
			before := marker.Read(installed)
			if before == nil || len(before.Builds) != 1 {
				t.Fatalf("first install did not record compiled state: %+v", before)
			}

			corruptCacheArtifact(t, home)
			if code, stdout, _ := capture(t, "status", "app", "--json"); code != exitOK ||
				decodeStatus(t, stdout).Builds[0].State != buildCorruptCache {
				t.Fatalf("status did not report corrupt compiled state:\n%s", stdout)
			}
			corruptArtifact := liveArtifactBytes(t, home)
			installedBefore := installedFingerprint(t, project, home)

			store := filepath.Join(project, ".agents", "skills")'''))

EDITS.append((r'''			if withdrawn := quarantinedEntries(t, home) - withdrawnBefore; withdrawn != 1 {''',
r'''			if withdrawn := quarantinedEntries(t, home); withdrawn != 1 {'''))

EDITS.append((r'''// assertProtectedCacheStateThatMovedDuringTheCheck closes the other half of the
// classification window.
//''',
r'''// TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck closes the other
// half of the classification window.
//'''))

EDITS.append((r'''func assertProtectedCacheStateThatMovedDuringTheCheck(t *testing.T, fixture compiledProjectFixture) {
	project, home := fixture.project, fixture.home

	cfg, code := loadConfig()''',
r'''func TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck(t *testing.T) {
	project, home := compiledProject(t)
	if code, stdout, stderr := capture(t, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	cfg, code := loadConfig()'''))

EDITS.append((r'''		t.Run(name, func(t *testing.T) {
			// Each case opens its own window over the state it is about to move:
			// the case before it put every protected cache byte back from its own
			// snapshot, so the production plan and the marker fingerprint this
			// window is taken from are acquired again here rather than carried in.
			planned := install.Project(cfg, project, "app", install.Options{DryRun: true})''',
r'''		t.Run(name, func(t *testing.T) {
			// Each case opens its own window: a preceding case restored the
			// installation through a real reinstall, which legitimately rewrites
			// the marker this fingerprint has to be taken from.
			planned := install.Project(cfg, project, "app", install.Options{DryRun: true})'''))

EDITS.append((r'''			// Every byte and permission bit this case moves is copied aside before
			// it moves, and put back when the case ends. Repairing the fixture by
			// reinstalling compiles the command again and proves nothing this test
			// owns; install, repair, and rollback keep their own dedicated cases.
			snapshotBuildCacheAfter(t, home)
			move(t, planned.Builds[0].Expectation().Input)''',
r'''			move(t, planned.Builds[0].Expectation().Input)
			t.Cleanup(func() { reinstall(t) })'''))

EDITS.append((r'''// assertUntrustedCompiledStateIsRepaired proves the documented reconciliation
// path: install rebuilds candidate bytes that are outside a provable protected
// boundary instead of adopting them.
func assertUntrustedCompiledStateIsRepaired(t *testing.T, fixture compiledProjectFixture) {
	home, installed := fixture.home, fixture.installed

	// The shared fixture is current when this case starts, and a preceding
	// repair legitimately republished the entry the marker names, so the
	// compiled state this case must find preserved is re-read here.
	before := marker.Read(installed)
	if before == nil || len(before.Builds) != 1 {
		t.Fatalf("the installation this case starts from records no compiled state: %+v", before)
	}
	for _, entry := range cacheEntries(t, home) {''',
r'''// TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall proves the
// documented reconciliation path: install rebuilds candidate bytes that are
// outside a provable protected boundary instead of adopting them.
func TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall(t *testing.T) {
	project, home := compiledProject(t)
	if code, stdout, stderr := capture(t, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	installed := filepath.Join(project, ".agents", "skills", "build-skill")
	before := marker.Read(installed)
	if before == nil || len(before.Builds) != 1 {
		t.Fatalf("first install did not record compiled state: %+v", before)
	}
	for _, entry := range cacheEntries(t, home) {'''))

text = AFTER.read_text()
for index, (new, old) in enumerate(EDITS):
    if text.count(new) != 1:
        sys.exit(f"edit {index}: found {text.count(new)} occurrences of the applied text, want 1")
    text = text.replace(new, old, 1)

digest = hashlib.sha256(text.encode()).hexdigest()
OUT.write_text(text)
print(f"reconstructed={digest}")
print(f"expected     ={WANT_BEFORE}")
if digest != WANT_BEFORE:
    sys.exit("reconstruction does not match the recorded pre-patch digest")
print("OK: reconstruction is byte-identical to the pre-patch file")
