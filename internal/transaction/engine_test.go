package transaction

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type testLock struct{ err error }

func (lock testLock) AssertHeld() error { return lock.err }

type pointerTestLock struct{}

func (*pointerTestLock) AssertHeld() error { return nil }

func TestPrepareCanonicalJournalReferencedKeysAndCommitCleanup(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	liveRoot := filepath.Join(root, "live")
	stageRoot := filepath.Join(root, "private-stage")
	mustMkdirAll(t, liveRoot)
	mustMkdirAll(t, stageRoot)

	targets := []Target{
		fileTarget(t, "shim", "文", filepath.Join(liveRoot, "three"), filepath.Join(stageRoot, "three"), "old-3", "new-3"),
		fileTarget(t, "runtime", "é", filepath.Join(liveRoot, "two"), filepath.Join(stageRoot, "two"), "old-2", "new-2"),
		fileTarget(t, "runtime", "z", filepath.Join(liveRoot, "one"), filepath.Join(stageRoot, "one"), "old-1", "new-1"),
	}
	engine := mustEngine(t, home)
	journal, err := engine.Prepare(testLock{}, Plan{
		TransactionID:       "txn-ordering",
		ProjectIdentity:     "/unrelated/project-b",
		Targets:             targets,
		ReferencedBuildKeys: []string{"key-z", "key-a", "key-é"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for index := 1; index < len(journal.Targets); index++ {
		if compareTargetRecords(journal.Targets[index-1], journal.Targets[index]) >= 0 {
			t.Fatalf("targets are not canonical: %+v", journal.Targets)
		}
	}
	if got, want := journal.ReferencedBuildKeys, []string{"key-a", "key-z", "key-é"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("referenced keys = %q, want %q", got, want)
	}
	payload, err := os.ReadFile(engine.journalPath(journal.TransactionID))
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := marshalJournal(journal)
	if err != nil || !bytes.Equal(payload, canonical) {
		t.Fatalf("journal is not canonical: %v\n%s", err, payload)
	}
	keys, err := engine.ReferencedBuildKeys(testLock{})
	if err != nil || !reflect.DeepEqual(keys, journal.ReferencedBuildKeys) {
		t.Fatalf("discover referenced keys = %q, %v", keys, err)
	}
	if err := engine.Commit(testLock{}, journal.TransactionID); err != nil {
		t.Fatal(err)
	}
	for _, target := range targets {
		if got := mustRead(t, target.LivePath); !strings.HasPrefix(got, "new-") {
			t.Fatalf("%s = %q after commit", target.LivePath, got)
		}
	}
	if _, err := os.Lstat(engine.journalPath(journal.TransactionID)); !os.IsNotExist(err) {
		t.Fatalf("successful journal remains: %v", err)
	}
	for _, target := range journal.Targets {
		for _, sidecar := range []string{target.BackupPath, target.RollbackPath, target.StagedPath} {
			if sidecar != "" {
				if _, err := os.Lstat(sidecar); !os.IsNotExist(err) {
					t.Fatalf("successful sidecar remains at %s: %v", sidecar, err)
				}
			}
		}
	}
	keys, err = engine.ReferencedBuildKeys(testLock{})
	if err != nil || len(keys) != 0 {
		t.Fatalf("keys after cleanup = %q, %v", keys, err)
	}
}

func TestFaultAtEveryTargetBoundaryRollsBackInReverse(t *testing.T) {
	points := []Point{PointBeforeBackup, PointAfterBackup, PointBeforeInstall, PointAfterInstall, PointTargetCommitted}
	for failIndex := 0; failIndex < 3; failIndex++ {
		for _, point := range points {
			name := fmt.Sprintf("target-%d/%s", failIndex, point)
			t.Run(name, func(t *testing.T) {
				root := t.TempDir()
				home := filepath.Join(root, "home")
				mustMkdirAll(t, filepath.Join(root, "live"))
				mustMkdirAll(t, filepath.Join(root, "stage"))
				var targets []Target
				for index := 0; index < 3; index++ {
					targets = append(targets, fileTarget(t, "class", fmt.Sprintf("%d", index), filepath.Join(root, "live", fmt.Sprintf("%d", index)), filepath.Join(root, "stage", fmt.Sprintf("%d", index)), fmt.Sprintf("old-%d", index), fmt.Sprintf("new-%d", index)))
				}
				var rollbackOrder []int
				injected := errors.New("injected target boundary")
				engine := mustEngine(t, home, WithHooks(Hooks{
					Fault: func(event Event) error {
						if event.TargetIndex == failIndex && event.Point == point {
							return injected
						}
						return nil
					},
					Observe: func(event Event) {
						if event.Point == PointTargetRolledBack {
							rollbackOrder = append(rollbackOrder, event.TargetIndex)
						}
					},
				}))
				journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-fault", ProjectIdentity: "/project", Targets: targets})
				if err != nil {
					t.Fatal(err)
				}
				err = engine.Commit(testLock{}, journal.TransactionID)
				if !errors.Is(err, injected) {
					t.Fatalf("commit error = %v, want injected", err)
				}
				for index, target := range targets {
					if got, want := mustRead(t, target.LivePath), fmt.Sprintf("old-%d", index); got != want {
						t.Fatalf("target %d = %q, want untouched/restored %q", index, got, want)
					}
				}
				lastTouched := failIndex
				if point == PointBeforeBackup {
					lastTouched--
				}
				var wantOrder []int
				for index := lastTouched; index >= 0; index-- {
					wantOrder = append(wantOrder, index)
				}
				if !reflect.DeepEqual(rollbackOrder, wantOrder) {
					t.Fatalf("rollback order = %v, want exact reverse %v", rollbackOrder, wantOrder)
				}
				if _, err := os.Lstat(engine.journalPath(journal.TransactionID)); !os.IsNotExist(err) {
					t.Fatalf("rolled-back journal remains: %v", err)
				}
			})
		}
	}
}

func TestJournalStateIsDurableBeforeEachNextSwap(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "live"))
	mustMkdirAll(t, filepath.Join(root, "stage"))
	targets := []Target{
		fileTarget(t, "class", "a", filepath.Join(root, "live", "a"), filepath.Join(root, "stage", "a"), "old-a", "new-a"),
		fileTarget(t, "class", "b", filepath.Join(root, "live", "b"), filepath.Join(root, "stage", "b"), "old-b", "new-b"),
	}
	var engine *Engine
	engine = mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{Observe: func(event Event) {
		if event.Point != PointBeforeInstall && (event.Point != PointBeforeBackup || event.TargetIndex != 1) {
			return
		}
		journal, err := engine.loadJournal(event.TransactionID)
		if err != nil {
			t.Fatalf("load durable boundary journal: %v", err)
		}
		if event.Point == PointBeforeInstall && journal.Targets[event.TargetIndex].State != StateBackedUp {
			t.Fatalf("state before install = %s, want backed_up", journal.Targets[event.TargetIndex].State)
		}
		if event.Point == PointBeforeBackup && journal.Targets[0].State != StateCommitted {
			t.Fatalf("prior target before next backup = %s, want committed", journal.Targets[0].State)
		}
	}}))
	journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-durability", ProjectIdentity: "/project", Targets: targets})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Commit(testLock{}, journal.TransactionID); err != nil {
		t.Fatal(err)
	}
}

func TestRollbackDesiredDigestMismatchPreservesUnknownBytes(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "live"))
	mustMkdirAll(t, filepath.Join(root, "stage"))
	targets := []Target{
		fileTarget(t, "class", "a", filepath.Join(root, "live", "a"), filepath.Join(root, "stage", "a"), "old-a", "new-a"),
		fileTarget(t, "class", "b", filepath.Join(root, "live", "b"), filepath.Join(root, "stage", "b"), "old-b", "new-b"),
		fileTarget(t, "class", "c", filepath.Join(root, "live", "c"), filepath.Join(root, "stage", "c"), "old-c", "new-c"),
	}
	injected := errors.New("trigger rollback")
	engine := mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{Fault: func(event Event) error {
		if event.Point == PointBeforeBackup && event.TargetIndex == 2 {
			mustWrite(t, targets[1].LivePath, "unknown-concurrent-state")
			return injected
		}
		return nil
	}}))
	journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-mismatch", ProjectIdentity: "/project", Targets: targets})
	if err != nil {
		t.Fatal(err)
	}
	err = engine.Commit(testLock{}, journal.TransactionID)
	if !errors.Is(err, injected) || !errors.Is(err, ErrImplementationCorruption) {
		t.Fatalf("commit error = %v, want injected plus implementation-corruption", err)
	}
	if got := mustRead(t, targets[1].LivePath); got != "unknown-concurrent-state" {
		t.Fatalf("unknown concurrent bytes overwritten: %q", got)
	}
	if got := mustRead(t, targets[2].LivePath); got != "old-c" {
		t.Fatalf("untouched later target changed: %q", got)
	}
	stored, err := engine.loadJournal(journal.TransactionID)
	if err != nil || stored.Phase != PhaseRollingBack {
		t.Fatalf("mismatch journal phase = %v, %v", stored, err)
	}
}

func TestDirectoryReplacementAndGenerationExpectation(t *testing.T) {
	root := t.TempDir()
	live := filepath.Join(root, "live", "runtime")
	source := filepath.Join(root, "stage", "runtime")
	mustMkdirAll(t, filepath.Join(live, "bin"))
	mustMkdirAll(t, filepath.Join(source, "bin"))
	mustWrite(t, filepath.Join(live, "generation"), "generation-1")
	mustWrite(t, filepath.Join(live, "bin", "tool"), "old")
	mustWrite(t, filepath.Join(source, "generation"), "generation-2")
	mustWrite(t, filepath.Join(source, "bin", "tool"), "new")
	engine := mustEngine(t, filepath.Join(root, "home"))
	journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-directory", ProjectIdentity: "/project", Targets: []Target{{
		Class: "runtime", Identifier: "tool", LivePath: live, StagedSource: source,
		ExpectedGeneration: "generation-1", GenerationPath: "generation",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if journal.Targets[0].ExpectedGeneration != "generation-1" || journal.Targets[0].PreimageDigest != "" {
		t.Fatalf("generation expectation not recorded: %+v", journal.Targets[0])
	}
	if err := engine.Commit(testLock{}, journal.TransactionID); err != nil {
		t.Fatal(err)
	}
	if got := mustRead(t, filepath.Join(live, "bin", "tool")); got != "new" {
		t.Fatalf("directory target = %q", got)
	}
}

func TestCreationAndRemovalRollbackRestoreExactPreimages(t *testing.T) {
	root := t.TempDir()
	liveRoot := filepath.Join(root, "live")
	stageRoot := filepath.Join(root, "stage")
	mustMkdirAll(t, liveRoot)
	mustMkdirAll(t, stageRoot)
	createdLive := filepath.Join(liveRoot, "created")
	createdSource := filepath.Join(stageRoot, "created")
	mustWrite(t, createdSource, "desired-created")
	removedLive := filepath.Join(liveRoot, "removed")
	mustWrite(t, removedLive, "preimage-removed")
	removedDigest, err := DigestPath(removedLive)
	if err != nil {
		t.Fatal(err)
	}
	targets := []Target{
		{Class: "class", Identifier: "a-create", LivePath: createdLive, StagedSource: createdSource, PreimageDigest: DigestAbsent},
		{Class: "class", Identifier: "b-remove", LivePath: removedLive, PreimageDigest: removedDigest},
	}
	injected := errors.New("rollback mixed creation/removal")
	engine := mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{Fault: func(event Event) error {
		if event.Point == PointTargetCommitted && event.TargetIndex == 1 {
			return injected
		}
		return nil
	}}))
	journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-create-remove", ProjectIdentity: "/project", Targets: targets})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Commit(testLock{}, journal.TransactionID); !errors.Is(err, injected) {
		t.Fatalf("commit error = %v, want injected", err)
	}
	if _, err := os.Lstat(createdLive); !os.IsNotExist(err) {
		t.Fatalf("newly created target survived rollback: %v", err)
	}
	if got := mustRead(t, removedLive); got != "preimage-removed" {
		t.Fatalf("removed target restored as %q", got)
	}
}

func TestRecoveryProcessesTransactionIDsAndRollbackStateDeterministically(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "live"))
	mustMkdirAll(t, filepath.Join(root, "stage"))
	engine := mustEngine(t, filepath.Join(root, "home"))
	for _, id := range []string{"txn-z", "txn-a"} {
		target := fileTarget(t, "class", id, filepath.Join(root, "live", id), filepath.Join(root, "stage", id), "old-"+id, "new-"+id)
		if _, err := engine.Prepare(testLock{}, Plan{TransactionID: id, ProjectIdentity: "/different/" + id, Targets: []Target{target}, ReferencedBuildKeys: []string{"key-" + id}}); err != nil {
			t.Fatal(err)
		}
	}
	var committed []string
	restarted := mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{Observe: func(event Event) {
		if event.Point == PointTargetCommitted {
			committed = append(committed, event.TransactionID)
		}
	}}))
	if err := restarted.Recover(testLock{}); err != nil {
		t.Fatal(err)
	}
	if want := []string{"txn-a", "txn-z"}; !reflect.DeepEqual(committed, want) {
		t.Fatalf("recovery order = %q, want transaction-id order %q", committed, want)
	}

	rollbackTarget := fileTarget(t, "class", "rollback", filepath.Join(root, "live", "rollback"), filepath.Join(root, "stage", "rollback"), "old", "new")
	injected := errors.New("crash after restoring preimage")
	pausing := mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{Fault: func(event Event) error {
		if event.Point == PointAfterRestore {
			return injected
		}
		if event.Point == PointTargetCommitted {
			return errors.New("start rollback")
		}
		return nil
	}}))
	journal, err := pausing.Prepare(testLock{}, Plan{TransactionID: "txn-rollback", ProjectIdentity: "/old/project", Targets: []Target{rollbackTarget}})
	if err != nil {
		t.Fatal(err)
	}
	err = pausing.Commit(testLock{}, journal.TransactionID)
	if !errors.Is(err, injected) {
		t.Fatalf("paused rollback error = %v", err)
	}
	if err := restarted.Recover(testLock{}); err != nil {
		t.Fatal(err)
	}
	if got := mustRead(t, rollbackTarget.LivePath); got != "old" {
		t.Fatalf("resumed rollback target = %q", got)
	}
}

func TestDurableRemovalCleansCrashTombWhenOriginalIsAbsent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "backup")
	tomb := path + ".delete"
	mustWrite(t, tomb, "retired")
	digest, err := DigestPath(tomb)
	if err != nil {
		t.Fatal(err)
	}
	if err := removeDurably(path, digest); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(tomb); !os.IsNotExist(err) {
		t.Fatalf("crash tomb remains: %v", err)
	}
}

func TestDurableRemovalPreservesUnprovenCrashTomb(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "backup")
	tomb := path + ".delete"
	mustWrite(t, tomb, "unknown")
	wantedPath := filepath.Join(root, "wanted")
	mustWrite(t, wantedPath, "recorded")
	wantedDigest, err := DigestPath(wantedPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := removeDurably(path, wantedDigest); !errors.Is(err, ErrImplementationCorruption) {
		t.Fatalf("remove error = %v, want implementation-corruption", err)
	}
	if got := mustRead(t, tomb); got != "unknown" {
		t.Fatalf("unknown tomb changed: %q", got)
	}
}

func TestAPIsRequireHeldHomeLock(t *testing.T) {
	engine := mustEngine(t, filepath.Join(t.TempDir(), "home"))
	var nilPointer *pointerTestLock
	locks := []HomeLock{nil, nilPointer, testLock{err: errors.New("released")}}
	for index, lock := range locks {
		if _, err := engine.Prepare(lock, Plan{}); err == nil {
			t.Fatalf("Prepare accepted lock %d", index)
		}
		if err := engine.Commit(lock, "txn"); err == nil {
			t.Fatalf("Commit accepted lock %d", index)
		}
		if err := engine.Recover(lock); err == nil {
			t.Fatalf("Recover accepted lock %d", index)
		}
		if _, err := engine.ReferencedBuildKeys(lock); err == nil {
			t.Fatalf("ReferencedBuildKeys accepted lock %d", index)
		}
	}
}

func TestConcurrentDigestAndReferencedKeyReads(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "live"))
	mustMkdirAll(t, filepath.Join(root, "stage"))
	target := fileTarget(t, "class", "race", filepath.Join(root, "live", "race"), filepath.Join(root, "stage", "race"), "old", "new")
	engine := mustEngine(t, filepath.Join(root, "home"))
	if _, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-race", ProjectIdentity: "/project", Targets: []Target{target}, ReferencedBuildKeys: []string{"key"}}); err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	for index := 0; index < 16; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for iteration := 0; iteration < 20; iteration++ {
				if _, err := DigestPath(target.LivePath); err != nil {
					t.Errorf("digest: %v", err)
				}
				if keys, err := engine.ReferencedBuildKeys(testLock{}); err != nil || !reflect.DeepEqual(keys, []string{"key"}) {
					t.Errorf("keys = %q, %v", keys, err)
				}
			}
		}()
	}
	wait.Wait()
}

func fileTarget(t *testing.T, class, identifier, live, source, old, desired string) Target {
	t.Helper()
	mustWrite(t, live, old)
	mustWrite(t, source, desired)
	digest, err := DigestPath(live)
	if err != nil {
		t.Fatal(err)
	}
	return Target{Class: class, Identifier: identifier, LivePath: live, StagedSource: source, PreimageDigest: digest}
}

func mustEngine(t *testing.T, home string, options ...Option) *Engine {
	t.Helper()
	engine, err := New(home, options...)
	if err != nil {
		t.Fatal(err)
	}
	return engine
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	mustMkdirAll(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}
