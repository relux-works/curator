package transaction

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/relux-works/curator/internal/staging"
)

// boundaryCall records one per-write guard invocation.
type boundaryCall struct {
	index    int
	livePath string
}

// recordingGuard is a scriptable BoundaryCheck: it records every call and
// refuses when refuse says so.
type recordingGuard struct {
	mu     sync.Mutex
	calls  []boundaryCall
	refuse func(index int) error
}

func (guard *recordingGuard) check(index int, livePath string) error {
	guard.mu.Lock()
	defer guard.mu.Unlock()
	guard.calls = append(guard.calls, boundaryCall{index: index, livePath: livePath})
	if guard.refuse != nil {
		return guard.refuse(index)
	}
	return nil
}

func (guard *recordingGuard) called(index int) int {
	guard.mu.Lock()
	defer guard.mu.Unlock()
	count := 0
	for _, call := range guard.calls {
		if call.index == index {
			count++
		}
	}
	return count
}

// TestCommitRunsBoundaryCheckBeforeEachWrite proves the engine consults the
// plan's guard before the live mutations of every target: the backup
// rename and the install rename each fire it with the journal index and
// live path, and a passing guard commits normally.
func TestCommitRunsBoundaryCheckBeforeEachWrite(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	liveRoot := filepath.Join(root, "live")
	stageRoot := filepath.Join(root, "private-stage")
	mustMkdirAll(t, liveRoot)
	mustMkdirAll(t, stageRoot)

	targets := []Target{
		fileTarget(t, "a", "one", filepath.Join(liveRoot, "one"), filepath.Join(stageRoot, "one"), "old-1", "new-1"),
		fileTarget(t, "a", "two", filepath.Join(liveRoot, "two"), filepath.Join(stageRoot, "two"), "old-2", "new-2"),
	}
	guard := &recordingGuard{}
	engine := mustEngine(t, home)
	journal, err := engine.Prepare(testLock{}, Plan{
		TransactionID:   "txn-boundary-pass",
		ProjectIdentity: "/test/project",
		Targets:         targets,
		BoundaryCheck:   guard.check,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Commit(testLock{}, journal.TransactionID); err != nil {
		t.Fatal(err)
	}
	// Journal order is canonical by (class, identifier): one, then two.
	// Each target fires the guard at least at its backup and at its
	// install; asserting per-index liveness (not an exact count) keeps
	// the test about the property, not the call tally.
	for index, want := range []string{
		filepath.Join(liveRoot, "one"),
		filepath.Join(liveRoot, "two"),
	} {
		if got := guard.called(index); got < 2 {
			t.Fatalf("target %d guard calls = %d, want at least 2 (backup and install)", index, got)
		}
		guard.mu.Lock()
		var paths []string
		for _, call := range guard.calls {
			if call.index == index {
				paths = append(paths, call.livePath)
			}
		}
		guard.mu.Unlock()
		for _, path := range paths {
			if path != want {
				t.Fatalf("target %d guard live path = %q, want %q", index, path, want)
			}
		}
	}
	if got := mustRead(t, filepath.Join(liveRoot, "one")); got != "new-1" {
		t.Fatalf("one = %q after commit, want new-1", got)
	}
	if got := mustRead(t, filepath.Join(liveRoot, "two")); got != "new-2" {
		t.Fatalf("two = %q after commit, want new-2", got)
	}
	if _, err := os.Lstat(engine.journalPath(journal.TransactionID)); !os.IsNotExist(err) {
		t.Fatalf("committed journal remains: %v", err)
	}
	engine.mu.Lock()
	remaining := len(engine.guards)
	engine.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("retained guards after terminal commit = %d, want 0", remaining)
	}
}

// TestCommitBoundaryRefusalAtBackupRollsBackPublishedTargets proves a
// refusal before the second target's backup rename restores the already
// committed first target and never touches the refused one: the live
// tree is the exact pre-commit state and the journal is gone.
func TestCommitBoundaryRefusalAtBackupRollsBackPublishedTargets(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	liveRoot := filepath.Join(root, "live")
	stageRoot := filepath.Join(root, "private-stage")
	mustMkdirAll(t, liveRoot)
	mustMkdirAll(t, stageRoot)

	targets := []Target{
		fileTarget(t, "a", "one", filepath.Join(liveRoot, "one"), filepath.Join(stageRoot, "one"), "old-1", "new-1"),
		fileTarget(t, "a", "two", filepath.Join(liveRoot, "two"), filepath.Join(stageRoot, "two"), "old-2", "new-2"),
	}
	guard := &recordingGuard{refuse: func(index int) error {
		if index == 1 {
			return fmt.Errorf("source_output_overlap: test refusal before backup")
		}
		return nil
	}}
	var observedMu sync.Mutex
	var backedUp []int
	engine := mustEngine(t, home, WithHooks(Hooks{Observe: func(event Event) {
		if event.Point == PointAfterBackup {
			observedMu.Lock()
			defer observedMu.Unlock()
			backedUp = append(backedUp, event.TargetIndex)
		}
	}}))
	journal, err := engine.Prepare(testLock{}, Plan{
		TransactionID:   "txn-boundary-refuse-backup",
		ProjectIdentity: "/test/project",
		Targets:         targets,
		BoundaryCheck:   guard.check,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = engine.Commit(testLock{}, journal.TransactionID)
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("commit err = %v, want source_output_overlap", err)
	}
	// The first target backed up and committed before the refusal; the
	// refused target never reached its backup rename, which is what
	// proves the backup-site check fired rather than the install-site
	// one catching it later.
	observedMu.Lock()
	defer observedMu.Unlock()
	for _, index := range backedUp {
		if index == 1 {
			t.Fatalf("refused target reached its backup rename; backedUp = %v", backedUp)
		}
	}
	found := false
	for _, index := range backedUp {
		if index == 0 {
			found = true
		}
	}
	if !found {
		t.Fatalf("first target never backed up; backedUp = %v", backedUp)
	}
	// The first target committed and was rolled back; the second was
	// refused before its backup rename and is untouched.
	if got := mustRead(t, filepath.Join(liveRoot, "one")); got != "old-1" {
		t.Fatalf("one = %q after refused commit, want old-1 (rolled back)", got)
	}
	if got := mustRead(t, filepath.Join(liveRoot, "two")); got != "old-2" {
		t.Fatalf("two = %q after refused commit, want old-2 (untouched)", got)
	}
	if _, err := os.Lstat(engine.journalPath(journal.TransactionID)); !os.IsNotExist(err) {
		t.Fatalf("refused journal remains: %v", err)
	}
	engine.mu.Lock()
	remaining := len(engine.guards)
	engine.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("retained guards after rolled-back commit = %d, want 0", remaining)
	}
}

// TestCommitBoundaryRefusalAtInstallRollsBackPublishedTargets proves the
// install-rename gate: a target with no live bytes to back up is still
// rechecked before its install rename, and a refusal there rolls back
// the earlier published target.
func TestCommitBoundaryRefusalAtInstallRollsBackPublishedTargets(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	liveRoot := filepath.Join(root, "live")
	stageRoot := filepath.Join(root, "private-stage")
	mustMkdirAll(t, liveRoot)
	mustMkdirAll(t, stageRoot)

	first := fileTarget(t, "a", "one", filepath.Join(liveRoot, "one"), filepath.Join(stageRoot, "one"), "old-1", "new-1")
	// The second target is a fresh install: no live bytes, so no backup
	// rename — only the install rename, which is still guarded.
	mustWrite(t, filepath.Join(stageRoot, "two"), "new-2")
	second := Target{
		Class: "a", Identifier: "two",
		LivePath:       filepath.Join(liveRoot, "two"),
		StagedSource:   filepath.Join(stageRoot, "two"),
		PreimageDigest: DigestAbsent,
	}
	guard := &recordingGuard{refuse: func(index int) error {
		if index == 1 {
			return fmt.Errorf("source_output_overlap: test refusal before install")
		}
		return nil
	}}
	engine := mustEngine(t, home)
	journal, err := engine.Prepare(testLock{}, Plan{
		TransactionID:   "txn-boundary-refuse-install",
		ProjectIdentity: "/test/project",
		Targets:         []Target{first, second},
		BoundaryCheck:   guard.check,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = engine.Commit(testLock{}, journal.TransactionID)
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("commit err = %v, want source_output_overlap", err)
	}
	if got := guard.called(1); got == 0 {
		t.Fatal("fresh-install target never consulted the guard before its install rename")
	}
	if got := mustRead(t, filepath.Join(liveRoot, "one")); got != "old-1" {
		t.Fatalf("one = %q after refused commit, want old-1 (rolled back)", got)
	}
	if _, err := os.Lstat(filepath.Join(liveRoot, "two")); !os.IsNotExist(err) {
		t.Fatalf("refused fresh-install target present after rollback: %v", err)
	}
	if _, err := os.Lstat(engine.journalPath(journal.TransactionID)); !os.IsNotExist(err) {
		t.Fatalf("refused journal remains: %v", err)
	}
}

// TestCommitRollbackNeverConsultsTheBoundaryGuard proves rollback is not
// itself gated: a guard that refuses everything still lets the refused
// commit's rollback complete, so the live tree returns to its preimage
// instead of stranding a journal.
func TestCommitRollbackNeverConsultsTheBoundaryGuard(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	liveRoot := filepath.Join(root, "live")
	stageRoot := filepath.Join(root, "private-stage")
	mustMkdirAll(t, liveRoot)
	mustMkdirAll(t, stageRoot)

	targets := []Target{
		fileTarget(t, "a", "one", filepath.Join(liveRoot, "one"), filepath.Join(stageRoot, "one"), "old-1", "new-1"),
	}
	guard := &recordingGuard{refuse: func(int) error {
		return fmt.Errorf("source_output_overlap: always refuse")
	}}
	engine := mustEngine(t, home)
	journal, err := engine.Prepare(testLock{}, Plan{
		TransactionID:   "txn-boundary-always-refuse",
		ProjectIdentity: "/test/project",
		Targets:         targets,
		BoundaryCheck:   guard.check,
	})
	if err != nil {
		t.Fatal(err)
	}
	// The only target is refused before its backup; there is nothing to
	// restore, and the rollback path must not consult the guard again
	// (it would refuse the restoration the same way).
	before := guard.called(0)
	if err := engine.Commit(testLock{}, journal.TransactionID); err == nil {
		t.Fatal("commit with an always-refusing guard succeeded")
	}
	// Exactly one call: the backup refusal aborts before the install,
	// and the rollback path consults nothing.
	if got := guard.called(0); got != before+1 {
		t.Fatalf("guard calls = %d, want exactly one forward refusal (rollback consults nothing)", got)
	}
	if got := mustRead(t, filepath.Join(liveRoot, "one")); got != "old-1" {
		t.Fatalf("one = %q after refused commit, want old-1", got)
	}
	if _, err := os.Lstat(engine.journalPath(journal.TransactionID)); !os.IsNotExist(err) {
		t.Fatalf("refused journal remains: %v", err)
	}
}

// TestRecoveryMustRecheckPhysicalBoundary proves restart recovery
// re-verifies the durable boundary before each remaining write: after a
// same-spelling parent swap (bytes and names identical, only the parent
// identity changed), a fresh Engine's Recover refuses with the boundary
// diagnostic and leaves the live bytes at the preimage.
func TestRecoveryMustRecheckPhysicalBoundary(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	parent := filepath.Join(root, "live")
	stage := filepath.Join(root, "stage")
	mustMkdirAll(t, parent)
	mustMkdirAll(t, stage)
	live := filepath.Join(parent, "one")
	target := fileTarget(t, "a", "one", live, filepath.Join(stage, "one"), "old", "new")
	st := staging.Target{LivePath: live}
	snapshot, err := (staging.Plan{Targets: []staging.Target{st}}).Snapshot(nil)
	if err != nil {
		t.Fatal(err)
	}
	durable, err := snapshot.Durable()
	if err != nil {
		t.Fatal(err)
	}
	guard := func(_ int, _ string) error { return staging.RecheckOne(st, snapshot, nil) }
	engine := mustEngine(t, home)
	if _, err := engine.Prepare(testLock{}, Plan{TransactionID: "recovery-boundary", ProjectIdentity: "/test/project", Targets: []Target{target}, BoundaryCheck: guard, BoundaryProof: &durable}); err != nil {
		t.Fatal(err)
	}
	aged := parent + "-old"
	if err := os.Rename(parent, aged); err != nil {
		t.Fatal(err)
	}
	mustMkdirAll(t, parent)
	entries, err := os.ReadDir(aged)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if err := os.Rename(filepath.Join(aged, entry.Name()), filepath.Join(parent, entry.Name())); err != nil {
			t.Fatal(err)
		}
	}
	if err := guard(0, live); err == nil {
		t.Fatal("control: physical guard did not detect replacement")
	} else {
		t.Logf("control guard refuses: %v", err)
	}
	restarted := mustEngine(t, home)
	err = restarted.Recover(testLock{})
	if err == nil {
		t.Fatalf("recovery published across replaced parent: live=%q, want refusal and old bytes", mustRead(t, live))
	}
	if !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("recovery err = %v, want source_output_overlap", err)
	}
	if got := mustRead(t, live); got != "old" {
		t.Fatalf("live = %q after refused recovery, want old", got)
	}
	if _, err := os.Lstat(restarted.journalPath("recovery-boundary")); !os.IsNotExist(err) {
		t.Fatalf("refused journal remains: %v", err)
	}
}

// TestRecoveryWithUnchangedBoundaryPublishes proves a protected journal
// with an intact boundary still commits after a restart: the durable
// proof restores and verifies, and the live bytes advance.
func TestRecoveryWithUnchangedBoundaryPublishes(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	parent := filepath.Join(root, "live")
	stage := filepath.Join(root, "stage")
	mustMkdirAll(t, parent)
	mustMkdirAll(t, stage)
	live := filepath.Join(parent, "one")
	target := fileTarget(t, "a", "one", live, filepath.Join(stage, "one"), "old", "new")
	st := staging.Target{LivePath: live}
	snapshot, err := (staging.Plan{Targets: []staging.Target{st}}).Snapshot(nil)
	if err != nil {
		t.Fatal(err)
	}
	durable, err := snapshot.Durable()
	if err != nil {
		t.Fatal(err)
	}
	guard := func(_ int, _ string) error { return staging.RecheckOne(st, snapshot, nil) }
	engine := mustEngine(t, home)
	if _, err := engine.Prepare(testLock{}, Plan{TransactionID: "recovery-unchanged", ProjectIdentity: "/test/project", Targets: []Target{target}, BoundaryCheck: guard, BoundaryProof: &durable}); err != nil {
		t.Fatal(err)
	}
	restarted := mustEngine(t, home)
	if err := restarted.Recover(testLock{}); err != nil {
		t.Fatalf("recovery with unchanged boundary: %v", err)
	}
	if got := mustRead(t, live); got != "new" {
		t.Fatalf("live = %q after recovery, want new", got)
	}
	if _, err := os.Lstat(restarted.journalPath("recovery-unchanged")); !os.IsNotExist(err) {
		t.Fatalf("committed journal remains: %v", err)
	}
}

// TestRecoveryUsesOriginalSnapshotIdentity proves the durable proof carries
// the planning-time identity even when the parent is swapped between
// Snapshot and Durable conversion: Durable serializes the stored pin, the
// pre-journal control passes after the original is restored, and a fresh
// Engine's Recover still refuses the replacement. Committed twin of the
// rev7 reviewer regression TestReviewRecoveryUsesOriginalSnapshotIdentity
// (same sequence through real Engine.Prepare and fresh Engine.Recover).
func TestRecoveryUsesOriginalSnapshotIdentity(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	parent := filepath.Join(root, "live")
	stage := filepath.Join(root, "stage")
	mustMkdirAll(t, parent)
	mustMkdirAll(t, stage)
	live := filepath.Join(parent, "one")
	target := fileTarget(t, "a", "one", live, filepath.Join(stage, "one"), "old", "new")
	st := staging.Target{LivePath: live}
	snapshot, err := (staging.Plan{Targets: []staging.Target{st}}).Snapshot(nil)
	if err != nil {
		t.Fatal(err)
	}
	guard := func(_ int, _ string) error { return staging.RecheckOne(st, snapshot, nil) }
	aged := parent + "-old"
	if err := os.Rename(parent, aged); err != nil {
		t.Fatal(err)
	}
	mustMkdirAll(t, parent)
	entries, err := os.ReadDir(aged)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if err := os.Rename(filepath.Join(aged, entry.Name()), filepath.Join(parent, entry.Name())); err != nil {
			t.Fatal(err)
		}
	}
	if err := guard(0, live); err == nil {
		t.Fatal("control: physical guard did not detect replacement")
	} else {
		t.Logf("control guard refuses: %v", err)
	}
	durable, err := snapshot.Durable()
	if err != nil {
		t.Fatal(err)
	}
	moveChildren := func(from, to string) {
		t.Helper()
		moved, err := os.ReadDir(from)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range moved {
			if err := os.Rename(filepath.Join(from, entry.Name()), filepath.Join(to, entry.Name())); err != nil {
				t.Fatal(err)
			}
		}
	}
	replacement := parent + "-replacement"
	if err := os.Rename(parent, replacement); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(aged, parent); err != nil {
		t.Fatal(err)
	}
	moveChildren(replacement, parent)
	if err := guard(0, live); err != nil {
		t.Fatalf("pre-journal control: %v", err)
	}
	engine := mustEngine(t, home)
	if _, err := engine.Prepare(testLock{}, Plan{TransactionID: "recovery-boundary", ProjectIdentity: "/test/project", Targets: []Target{target}, BoundaryCheck: guard, BoundaryProof: &durable}); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(parent, aged); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(replacement, parent); err != nil {
		t.Fatal(err)
	}
	moveChildren(aged, parent)
	if err := guard(0, live); err == nil {
		t.Fatal("same-process guard must refuse")
	}
	restarted := mustEngine(t, home)
	err = restarted.Recover(testLock{})
	if err == nil {
		t.Fatalf("recovery published across replaced parent: live=%q, want refusal and old bytes", mustRead(t, live))
	}
	if !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("recovery err = %v, want source_output_overlap", err)
	}
	if got := mustRead(t, live); got != "old" {
		t.Fatalf("live = %q after refused recovery, want old", got)
	}
	if _, err := os.Lstat(restarted.journalPath("recovery-boundary")); !os.IsNotExist(err) {
		t.Fatalf("refused journal remains: %v", err)
	}
}

// TestRecoveryCaptureWindowAncestorRefuses proves a swap injected inside
// Snapshot — between the pin inspection and the token derivation —
// cannot plant an inconsistent pin: the committed twin of the rev8
// reviewer probe TestReviewWindowsCaptureCoherence for destination
// ancestors. Capture pins the pre-swap parent, the pre-journal control
// passes after the original is restored, Prepare persists a protected
// journal, and a fresh Engine's Recover refuses the replacement with the
// boundary diagnostic and leaves the live bytes at the preimage.
func TestRecoveryCaptureWindowAncestorRefuses(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	parent := filepath.Join(root, "live")
	stage := filepath.Join(root, "stage")
	mustMkdirAll(t, parent)
	mustMkdirAll(t, stage)
	live := filepath.Join(parent, "one")
	target := fileTarget(t, "a", "one", live, filepath.Join(stage, "one"), "old", "new")
	st := staging.Target{LivePath: live}
	canonicalParent, err := staging.Canonicalize(parent)
	if err != nil {
		t.Fatal(err)
	}
	aged := parent + "-old"
	fired := false
	var hookErr error
	staging.SetCaptureHook(func(canonical string) {
		if fired || hookErr != nil || canonical != canonicalParent {
			return
		}
		fired = true
		if err := os.Rename(parent, aged); err != nil {
			hookErr = err
			return
		}
		if err := os.Mkdir(parent, 0o755); err != nil {
			hookErr = err
			return
		}
		entries, err := os.ReadDir(aged)
		if err != nil {
			hookErr = err
			return
		}
		for _, entry := range entries {
			if err := os.Rename(filepath.Join(aged, entry.Name()), filepath.Join(parent, entry.Name())); err != nil {
				hookErr = err
				return
			}
		}
	})
	t.Cleanup(func() { staging.SetCaptureHook(nil) })
	snapshot, err := (staging.Plan{Targets: []staging.Target{st}}).Snapshot(nil)
	staging.SetCaptureHook(nil)
	if err != nil {
		t.Fatal(err)
	}
	if hookErr != nil {
		t.Fatal(hookErr)
	}
	if !fired {
		t.Fatal("capture hook never fired for the planned parent")
	}
	guard := func(_ int, _ string) error { return staging.RecheckOne(st, snapshot, nil) }
	if err := guard(0, live); err == nil {
		t.Fatal("control: physical guard did not detect replacement")
	} else {
		t.Logf("control guard refuses: %v", err)
	}
	durable, err := snapshot.Durable()
	if err != nil {
		t.Fatal(err)
	}
	moveChildren := func(from, to string) {
		t.Helper()
		moved, err := os.ReadDir(from)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range moved {
			if err := os.Rename(filepath.Join(from, entry.Name()), filepath.Join(to, entry.Name())); err != nil {
				t.Fatal(err)
			}
		}
	}
	replacement := parent + "-replacement"
	if err := os.Rename(parent, replacement); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(aged, parent); err != nil {
		t.Fatal(err)
	}
	moveChildren(replacement, parent)
	if err := guard(0, live); err != nil {
		t.Fatalf("pre-journal control: %v", err)
	}
	engine := mustEngine(t, home)
	if _, err := engine.Prepare(testLock{}, Plan{TransactionID: "recovery-capture-ancestor", ProjectIdentity: "/test/project", Targets: []Target{target}, BoundaryCheck: guard, BoundaryProof: &durable}); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(parent, aged); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(replacement, parent); err != nil {
		t.Fatal(err)
	}
	moveChildren(aged, parent)
	if err := guard(0, live); err == nil {
		t.Fatal("same-process guard must refuse")
	}
	restarted := mustEngine(t, home)
	err = restarted.Recover(testLock{})
	if err == nil {
		t.Fatalf("recovery published across replaced parent: live=%q, want refusal and old bytes", mustRead(t, live))
	}
	if !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("recovery err = %v, want source_output_overlap", err)
	}
	if got := mustRead(t, live); got != "old" {
		t.Fatalf("live = %q after refused recovery, want old", got)
	}
	if _, err := os.Lstat(restarted.journalPath("recovery-capture-ancestor")); !os.IsNotExist(err) {
		t.Fatalf("refused journal remains: %v", err)
	}
}

// TestRecoveryCaptureWindowAdmittedRefuses proves the admitted-input half
// of the capture window through the real journal: a swap injected inside
// Snapshot while the admitted input is pinned cannot plant the
// replacement into the durable proof. The destination parent is never
// touched, so only the admitted-identity gate can refuse at recovery.
func TestRecoveryCaptureWindowAdmittedRefuses(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	parent := filepath.Join(root, "live")
	stage := filepath.Join(root, "stage")
	admitted := filepath.Join(root, "authored")
	mustMkdirAll(t, parent)
	mustMkdirAll(t, stage)
	mustMkdirAll(t, admitted)
	live := filepath.Join(parent, "one")
	target := fileTarget(t, "a", "one", live, filepath.Join(stage, "one"), "old", "new")
	st := staging.Target{LivePath: live}
	canonicalAdmitted, err := staging.Canonicalize(admitted)
	if err != nil {
		t.Fatal(err)
	}
	aged := admitted + "-old"
	fired := false
	var hookErr error
	staging.SetCaptureHook(func(canonical string) {
		if fired || hookErr != nil || canonical != canonicalAdmitted {
			return
		}
		fired = true
		if err := os.Rename(admitted, aged); err != nil {
			hookErr = err
			return
		}
		if err := os.Mkdir(admitted, 0o755); err != nil {
			hookErr = err
		}
	})
	t.Cleanup(func() { staging.SetCaptureHook(nil) })
	snapshot, err := (staging.Plan{Targets: []staging.Target{st}}).Snapshot([]string{admitted})
	staging.SetCaptureHook(nil)
	if err != nil {
		t.Fatal(err)
	}
	if hookErr != nil {
		t.Fatal(hookErr)
	}
	if !fired {
		t.Fatal("capture hook never fired for the admitted input")
	}
	guard := func(_ int, _ string) error { return staging.RecheckOne(st, snapshot, []string{admitted}) }
	if err := guard(0, live); err == nil {
		t.Fatal("control: physical guard did not detect replacement")
	} else {
		t.Logf("control guard refuses: %v", err)
	}
	durable, err := snapshot.Durable()
	if err != nil {
		t.Fatal(err)
	}
	replacement := admitted + "-replacement"
	if err := os.Rename(admitted, replacement); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(aged, admitted); err != nil {
		t.Fatal(err)
	}
	if err := guard(0, live); err != nil {
		t.Fatalf("pre-journal control: %v", err)
	}
	engine := mustEngine(t, home)
	if _, err := engine.Prepare(testLock{}, Plan{TransactionID: "recovery-capture-admitted", ProjectIdentity: "/test/project", Targets: []Target{target}, BoundaryCheck: guard, BoundaryProof: &durable}); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(admitted, aged); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(replacement, admitted); err != nil {
		t.Fatal(err)
	}
	if err := guard(0, live); err == nil {
		t.Fatal("same-process guard must refuse")
	}
	restarted := mustEngine(t, home)
	err = restarted.Recover(testLock{})
	if err == nil {
		t.Fatalf("recovery published across replaced input: live=%q, want refusal and old bytes", mustRead(t, live))
	}
	if !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("recovery err = %v, want source_output_overlap", err)
	}
	if !strings.Contains(err.Error(), "admitted input") {
		t.Fatalf("recovery err = %v, want the admitted-identity gate", err)
	}
	if got := mustRead(t, live); got != "old" {
		t.Fatalf("live = %q after refused recovery, want old", got)
	}
	if _, err := os.Lstat(restarted.journalPath("recovery-capture-admitted")); !os.IsNotExist(err) {
		t.Fatalf("refused journal remains: %v", err)
	}
}

// TestRecoveryLegacyJournalSkipsGuard proves journals without a boundary
// proof keep their legacy behavior across a restart: no guard, normal
// publication, even across a same-spelling swap the guard would refuse.
func TestRecoveryLegacyJournalSkipsGuard(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	parent := filepath.Join(root, "live")
	stage := filepath.Join(root, "stage")
	mustMkdirAll(t, parent)
	mustMkdirAll(t, stage)
	live := filepath.Join(parent, "one")
	target := fileTarget(t, "a", "one", live, filepath.Join(stage, "one"), "old", "new")
	engine := mustEngine(t, home)
	if _, err := engine.Prepare(testLock{}, Plan{TransactionID: "recovery-legacy", ProjectIdentity: "/test/project", Targets: []Target{target}}); err != nil {
		t.Fatal(err)
	}
	aged := parent + "-old"
	if err := os.Rename(parent, aged); err != nil {
		t.Fatal(err)
	}
	mustMkdirAll(t, parent)
	entries, err := os.ReadDir(aged)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if err := os.Rename(filepath.Join(aged, entry.Name()), filepath.Join(parent, entry.Name())); err != nil {
			t.Fatal(err)
		}
	}
	restarted := mustEngine(t, home)
	if err := restarted.Recover(testLock{}); err != nil {
		t.Fatalf("legacy recovery: %v", err)
	}
	if got := mustRead(t, live); got != "new" {
		t.Fatalf("live = %q after legacy recovery, want new", got)
	}
}
