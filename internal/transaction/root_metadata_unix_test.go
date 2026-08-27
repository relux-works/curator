//go:build unix

package transaction

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryDigestBindsRootMetadata(t *testing.T) {
	root := filepath.Join(t.TempDir(), "target")
	mustWrite(t, filepath.Join(root, "child"), "bytes")
	before, err := DigestPath(root)
	if err != nil {
		t.Fatal(err)
	}
	changedMode := changeDirectoryMode(t, root)
	after, err := DigestPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if before == after {
		t.Fatalf("directory digest ignored root mode %04o", changedMode)
	}
}

func TestRollbackRejectsRootDirectoryMetadataChangeAndPreservesTarget(t *testing.T) {
	root := t.TempDir()
	live, source, target := directoryTarget(t, root)
	var changedMode os.FileMode
	injected := errors.New("start rollback after root metadata change")
	engine := mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{Fault: func(event Event) error {
		if event.Point == PointTargetCommitted {
			changedMode = changeDirectoryMode(t, live)
			return injected
		}
		return nil
	}}))
	journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-root-mode-rollback", ProjectIdentity: "/project", Targets: []Target{target}})
	if err != nil {
		t.Fatal(err)
	}
	err = engine.Commit(testLock{}, journal.TransactionID)
	if !errors.Is(err, injected) || !errors.Is(err, ErrImplementationCorruption) {
		t.Fatalf("commit error = %v, want injected plus implementation-corruption", err)
	}
	assertDirectoryTargetPreserved(t, live, source, changedMode)
	if _, err := os.Lstat(journal.Targets[0].BackupPath); err != nil {
		t.Fatalf("rollback mismatch removed backup: %v", err)
	}
	if _, err := os.Lstat(engine.journalPath(journal.TransactionID)); err != nil {
		t.Fatalf("rollback mismatch removed journal: %v", err)
	}
}

func TestCleanupRejectsRootDirectoryMetadataChangeAndRetainsRecoveryState(t *testing.T) {
	root := t.TempDir()
	live, source, target := directoryTarget(t, root)
	var changedMode os.FileMode
	injected := errors.New("stop after root metadata change")
	engine := mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{Fault: func(event Event) error {
		if event.Point == PointBeforeCleanup {
			changedMode = changeDirectoryMode(t, live)
			return injected
		}
		return nil
	}}))
	journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-root-mode-cleanup", ProjectIdentity: "/project", Targets: []Target{target}})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Commit(testLock{}, journal.TransactionID); !errors.Is(err, injected) {
		t.Fatalf("commit error = %v, want injected", err)
	}
	engine.hooks = Hooks{}
	err = engine.Recover(testLock{})
	if !errors.Is(err, ErrImplementationCorruption) {
		t.Fatalf("cleanup recovery error = %v, want implementation-corruption", err)
	}
	assertDirectoryTargetPreserved(t, live, source, changedMode)
	if _, err := os.Lstat(journal.Targets[0].BackupPath); err != nil {
		t.Fatalf("cleanup mismatch removed backup: %v", err)
	}
	if _, err := os.Lstat(engine.journalPath(journal.TransactionID)); err != nil {
		t.Fatalf("cleanup mismatch removed journal: %v", err)
	}
}

func directoryTarget(t *testing.T, root string) (string, string, Target) {
	t.Helper()
	live := filepath.Join(root, "live", "runtime")
	source := filepath.Join(root, "stage", "runtime")
	mustWrite(t, filepath.Join(live, "bin", "tool"), "old")
	mustWrite(t, filepath.Join(source, "bin", "tool"), "new")
	digest, err := DigestPath(live)
	if err != nil {
		t.Fatal(err)
	}
	return live, source, Target{Class: "runtime", Identifier: "tool", LivePath: live, StagedSource: source, PreimageDigest: digest}
}

func changeDirectoryMode(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	changed := info.Mode().Perm() ^ 0o040
	if err := os.Chmod(path, changed); err != nil {
		t.Fatal(err)
	}
	return changed
}

func assertDirectoryTargetPreserved(t *testing.T, live, source string, changedMode os.FileMode) {
	t.Helper()
	if got := mustRead(t, filepath.Join(live, "bin", "tool")); got != "new" {
		t.Fatalf("current directory bytes overwritten: %q", got)
	}
	info, err := os.Lstat(live)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != changedMode {
		t.Fatalf("current directory mode = %04o, want preserved %04o", info.Mode().Perm(), changedMode)
	}
	if got := mustRead(t, filepath.Join(source, "bin", "tool")); got != "new" {
		t.Fatalf("source directory changed: %q", got)
	}
}
