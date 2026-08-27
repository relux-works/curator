package transaction

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNamespaceIdentityIsReadOnceWithinOneValidationPass pins the memoization
// that keeps the pairwise independence sweep at O(P) filesystem reads. A path
// takes part in P-1 pairs; the object it names must be read for the first of
// them and answered from the pass snapshot for the rest.
func TestNamespaceIdentityIsReadOnceWithinOneValidationPass(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "live", "target")
	mustWrite(t, path, "present")

	resolved := resolveNamespacePath(targetNamespacePath{owner: "target 0", kind: "live", path: path, key: path})
	first, firstErr := resolved.identity()
	if firstErr != nil {
		t.Fatalf("first identity read failed: %v", firstErr)
	}

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	second, secondErr := resolved.identity()
	if secondErr != nil {
		t.Fatalf("second identity read consulted the filesystem again: %v", secondErr)
	}
	if !os.SameFile(first, second) {
		t.Fatal("second identity read did not reuse the pass snapshot")
	}
}

// TestNamespaceIdentitySnapshotDoesNotOutliveItsPass is the other half of the
// contract: the snapshot above must be scoped to one pass, so a later pass —
// and therefore a later saveJournal call — reads the filesystem as it is now.
func TestNamespaceIdentitySnapshotDoesNotOutliveItsPass(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "live", "target")
	mustWrite(t, path, "present")

	candidate := targetNamespacePath{owner: "target 0", kind: "live", path: path, key: path}
	firstPass := resolveNamespacePath(candidate)
	if _, err := firstPass.identity(); err != nil {
		t.Fatalf("first pass identity read failed: %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	secondPass := resolveNamespacePath(candidate)
	if _, err := secondPass.identity(); !os.IsNotExist(err) {
		t.Fatalf("second pass identity error = %v, want not-exist", err)
	}
}

// TestValidateIndependentTargetNamespacesRejectsMalformedPaths keeps the
// resolution stage fail-closed: a path that is not valid absolute filesystem
// text is rejected outright rather than compared as an opaque string.
func TestValidateIndependentTargetNamespacesRejectsMalformedPaths(t *testing.T) {
	root := t.TempDir()
	valid := namespaceTargetRecord(0, "class", "id", filepath.Join(root, "live", "target"))
	tests := []struct {
		name     string
		mutate   func(*TargetRecord)
		reserved []targetNamespacePath
	}{
		{name: "relative live", mutate: func(value *TargetRecord) { value.LivePath = "live/target" }},
		{name: "nul live", mutate: func(value *TargetRecord) { value.LivePath = filepath.Join(root, "live\x00target") }},
		{name: "invalid utf8 live", mutate: func(value *TargetRecord) { value.LivePath = filepath.Join(root, "live\xff") }},
		{name: "relative staged", mutate: func(value *TargetRecord) { value.StagedPath = "stage/target" }},
		{name: "relative backup", mutate: func(value *TargetRecord) { value.BackupPath = "backup" }},
		{name: "nul rollback", mutate: func(value *TargetRecord) { value.RollbackPath = filepath.Join(root, "roll\x00back") }},
		{name: "relative entry live", mutate: func(value *TargetRecord) {
			value.Kind = KindEntry
			value.LivePath = "live/entry"
		}},
		{
			name:     "relative reserved",
			mutate:   func(*TargetRecord) {},
			reserved: []targetNamespacePath{{owner: "engine", kind: "journal namespace", path: "home/state"}},
		},
		{
			name:     "nul reserved",
			mutate:   func(*TargetRecord) {},
			reserved: []targetNamespacePath{{owner: "engine", kind: "journal namespace", path: filepath.Join(root, "home\x00state")}},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			target := valid
			testCase.mutate(&target)
			err := validateIndependentTargetNamespaces([]TargetRecord{target}, testCase.reserved...)
			if err == nil {
				t.Fatal("malformed target namespace was accepted")
			}
			if !strings.Contains(err.Error(), "path is not valid absolute filesystem text") {
				t.Fatalf("error = %v, want a malformed-path rejection", err)
			}
		})
	}
}

// TestValidateIndependentTargetNamespacesRejectsOverlappingPaths sweeps the ways
// two declared paths can name the same object. Containment, exact repetition and
// filesystem aliasing all have to fail, and so does a target that reaches into a
// namespace the manager reserves for itself.
func TestValidateIndependentTargetNamespacesRejectsOverlappingPaths(t *testing.T) {
	tests := []struct {
		name  string
		build func(t *testing.T, root string) ([]TargetRecord, []targetNamespacePath)
	}{
		{
			name: "nested live paths",
			build: func(_ *testing.T, root string) ([]TargetRecord, []targetNamespacePath) {
				first := namespaceTargetRecord(0, "a", "outer", filepath.Join(root, "live", "target"))
				second := namespaceTargetRecord(1, "b", "inner", filepath.Join(root, "live", "target", "nested"))
				return []TargetRecord{first, second}, nil
			},
		},
		{
			name: "repeated live path",
			build: func(_ *testing.T, root string) ([]TargetRecord, []targetNamespacePath) {
				live := filepath.Join(root, "live", "target")
				first := namespaceTargetRecord(0, "a", "first", live)
				second := namespaceTargetRecord(1, "b", "second", live)
				return []TargetRecord{first, second}, nil
			},
		},
		{
			name: "live path is another target's backup sidecar",
			build: func(_ *testing.T, root string) ([]TargetRecord, []targetNamespacePath) {
				first := namespaceTargetRecord(0, "a", "first", filepath.Join(root, "live", "target"))
				second := namespaceTargetRecord(1, "b", "second", first.BackupPath)
				return []TargetRecord{first, second}, nil
			},
		},
		{
			name: "live path is another target's cleanup tomb",
			build: func(_ *testing.T, root string) ([]TargetRecord, []targetNamespacePath) {
				first := namespaceTargetRecord(0, "a", "first", filepath.Join(root, "live", "target"))
				second := namespaceTargetRecord(1, "b", "second", first.RollbackPath+".delete")
				return []TargetRecord{first, second}, nil
			},
		},
		{
			name: "hard link alias between live paths",
			build: func(t *testing.T, root string) ([]TargetRecord, []targetNamespacePath) {
				first := namespaceTargetRecord(0, "a", "first", filepath.Join(root, "live-a", "target"))
				second := namespaceTargetRecord(1, "b", "second", filepath.Join(root, "live-b", "target"))
				mustWrite(t, first.LivePath, "shared")
				mustMkdirAll(t, filepath.Dir(second.LivePath))
				if err := os.Link(first.LivePath, second.LivePath); err != nil {
					t.Skipf("filesystem does not support hard links: %v", err)
				}
				return []TargetRecord{first, second}, nil
			},
		},
		{
			name: "symbolic link alias between live parents",
			build: func(t *testing.T, root string) ([]TargetRecord, []targetNamespacePath) {
				first := namespaceTargetRecord(0, "a", "first", filepath.Join(root, "live-a", "target"))
				second := namespaceTargetRecord(1, "b", "second", filepath.Join(root, "live-b", "target"))
				mustWrite(t, first.LivePath, "present")
				if err := os.Symlink(filepath.Dir(first.LivePath), filepath.Dir(second.LivePath)); err != nil {
					t.Skipf("filesystem does not support symbolic links: %v", err)
				}
				return []TargetRecord{first, second}, nil
			},
		},
		{
			name: "target reaches into the reserved namespace",
			build: func(t *testing.T, root string) ([]TargetRecord, []targetNamespacePath) {
				reserved := filepath.Join(root, "home", "state")
				mustMkdirAll(t, reserved)
				target := namespaceTargetRecord(0, "a", "first", filepath.Join(reserved, "target"))
				return []TargetRecord{target}, []targetNamespacePath{{owner: "engine", kind: "journal namespace", path: reserved}}
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			targets, reserved := testCase.build(t, root)
			err := validateIndependentTargetNamespaces(targets, reserved...)
			if err == nil {
				t.Fatal("overlapping target namespaces were accepted")
			}
			if !strings.Contains(err.Error(), "overlaps") {
				t.Fatalf("error = %v, want an overlap rejection", err)
			}
		})
	}
}

// TestValidateIndependentTargetNamespacesAcceptsDisjointPaths guards the sweep
// against becoming vacuously strict: the memoized snapshot must still let a
// genuinely disjoint graph through, including one with many paths.
func TestValidateIndependentTargetNamespacesAcceptsDisjointPaths(t *testing.T) {
	root := t.TempDir()
	targets := make([]TargetRecord, 0, 24)
	for index := 0; index < 24; index++ {
		live := filepath.Join(root, "live", "target-"+string(rune('a'+index)))
		mustWrite(t, live, "present")
		targets = append(targets, namespaceTargetRecord(index, "class", "id-"+string(rune('a'+index)), live))
	}
	reserved := filepath.Join(root, "home", "state")
	mustMkdirAll(t, reserved)
	if err := validateIndependentTargetNamespaces(targets, targetNamespacePath{owner: "engine", kind: "journal namespace", path: reserved}); err != nil {
		t.Fatalf("disjoint target namespaces were rejected: %v", err)
	}
}

// TestSaveJournalRejectsNamespaceAliasIntroducedBetweenSaves is the fail-closed
// case the per-pass snapshot exists to preserve. The first save resolves a
// disjoint graph; the filesystem then aliases two of its paths onto one object;
// the second save has to see that and refuse before touching the journal.
func TestSaveJournalRejectsNamespaceAliasIntroducedBetweenSaves(t *testing.T) {
	tests := []struct {
		name  string
		alias func(t *testing.T, first, second string)
	}{
		{
			name: "hard link alias",
			alias: func(t *testing.T, first, second string) {
				if err := os.Remove(second); err != nil {
					t.Fatal(err)
				}
				if err := os.Link(first, second); err != nil {
					t.Skipf("filesystem does not support hard links: %v", err)
				}
			},
		},
		{
			name: "symbolic link alias on the live parent",
			alias: func(t *testing.T, first, second string) {
				parent := filepath.Dir(second)
				if err := os.RemoveAll(parent); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Dir(first), parent); err != nil {
					t.Skipf("filesystem does not support symbolic links: %v", err)
				}
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			// The two live paths share a base name so that aliasing either the
			// files or their parent directories makes the resolved keys collide.
			firstLive := filepath.Join(root, "live-a", "shared")
			secondLive := filepath.Join(root, "live-b", "shared")
			first := fileTarget(t, "a", "first", firstLive, filepath.Join(root, "stage", "a"), "old-a", "new-a")
			second := fileTarget(t, "b", "second", secondLive, filepath.Join(root, "stage", "b"), "old-b", "new-b")
			engine := mustEngine(t, filepath.Join(root, "home"))
			journal, _, err := engine.buildJournal(Plan{TransactionID: "txn-between-save-alias", ProjectIdentity: "/project", Targets: []Target{first, second}})
			if err != nil {
				t.Fatal(err)
			}
			if err := engine.saveJournal(journal); err != nil {
				t.Fatalf("disjoint graph rejected on the first save: %v", err)
			}
			saved := mustRead(t, engine.journalPath(journal.TransactionID))

			testCase.alias(t, firstLive, secondLive)

			journal.Targets[0].State = StateBackedUp
			if err := engine.saveJournal(journal); err == nil {
				t.Fatal("second save accepted a graph the filesystem had aliased")
			} else if !errors.Is(err, errInvalidJournal) {
				t.Fatalf("second save error = %v, want a journal rejection", err)
			}
			if got := mustRead(t, engine.journalPath(journal.TransactionID)); got != saved {
				t.Fatal("rejected save mutated the journal on disk")
			}
		})
	}
}

// TestRecoverRejectsDecodedTargetNamespacesAliasedWhileStopped covers the
// externally supplied graph: a journal that was valid when written is decoded
// again after the filesystem changed underneath it, and recovery has to
// revalidate it before resuming any mutation.
func TestRecoverRejectsDecodedTargetNamespacesAliasedWhileStopped(t *testing.T) {
	root := t.TempDir()
	firstLive := filepath.Join(root, "live-a", "shared")
	secondLive := filepath.Join(root, "live-b", "shared")
	injected := errors.New("stop after prepare")
	first := fileTarget(t, "a", "first", firstLive, filepath.Join(root, "stage", "a"), "old-a", "new-a")
	second := fileTarget(t, "b", "second", secondLive, filepath.Join(root, "stage", "b"), "old-b", "new-b")
	home := filepath.Join(root, "home")
	engine := mustEngine(t, home, WithHooks(Hooks{Fault: func(event Event) error {
		if event.Point == PointBeforeCleanup {
			return injected
		}
		return nil
	}}))
	journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-recover-alias", ProjectIdentity: "/project", Targets: []Target{first, second}})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Commit(testLock{}, journal.TransactionID); !errors.Is(err, injected) {
		t.Fatalf("commit error = %v", err)
	}

	if err := os.RemoveAll(filepath.Dir(secondLive)); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Dir(firstLive), filepath.Dir(secondLive)); err != nil {
		t.Skipf("filesystem does not support symbolic links: %v", err)
	}
	before := mustRead(t, firstLive)

	restarted := mustEngine(t, home)
	if err := restarted.Recover(testLock{}); err == nil {
		t.Fatal("recovery accepted a decoded graph the filesystem had aliased")
	} else if !errors.Is(err, errInvalidJournal) {
		t.Fatalf("recovery error = %v, want a journal rejection", err)
	}
	if got := mustRead(t, firstLive); got != before {
		t.Fatalf("rejected recovery mutated a live target: %q", got)
	}
}

// namespaceTargetRecord builds the record shape the namespace sweep reads, with
// the canonical sidecars the rest of journal validation would have enforced.
// The paths need not exist: resolution falls back to the closest existing
// ancestor, which is what a not-yet-created sidecar looks like in practice.
func namespaceTargetRecord(index int, class, identifier, live string) TargetRecord {
	return TargetRecord{
		Class:        class,
		Identifier:   identifier,
		Kind:         KindBytes,
		LivePath:     live,
		StagedPath:   sidecarPath(live, "txn-namespace", index, "desired"),
		BackupPath:   sidecarPath(live, "txn-namespace", index, "backup"),
		RollbackPath: sidecarPath(live, "txn-namespace", index, "rollback"),
	}
}

// BenchmarkValidateIndependentTargetNamespaces measures one validation pass over
// a disjoint graph as the declared path count P grows. The pairwise sweep visits
// O(P^2) pairs either way; what this separates is whether each pair costs a
// filesystem round trip. Per-pass reuse should leave the cost dominated by
// in-memory comparison, so the time per pair stops growing with P.
func BenchmarkValidateIndependentTargetNamespaces(b *testing.B) {
	for _, count := range []int{8, 16, 32, 64} {
		b.Run(fmt.Sprintf("targets=%d", count), func(b *testing.B) {
			root := b.TempDir()
			targets := namespaceBenchmarkTargets(b, root, count)
			reserved := filepath.Join(root, "home", "state")
			if err := os.MkdirAll(reserved, 0o700); err != nil {
				b.Fatal(err)
			}
			journalNamespace := targetNamespacePath{owner: "engine", kind: "journal namespace", path: reserved}
			b.ReportMetric(float64(len(targets)*7+1), "paths")
			b.ResetTimer()
			for iteration := 0; iteration < b.N; iteration++ {
				if err := validateIndependentTargetNamespaces(targets, journalNamespace); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func namespaceBenchmarkTargets(tb testing.TB, root string, count int) []TargetRecord {
	tb.Helper()
	live := filepath.Join(root, "live")
	if err := os.MkdirAll(live, 0o700); err != nil {
		tb.Fatal(err)
	}
	targets := make([]TargetRecord, 0, count)
	for index := 0; index < count; index++ {
		path := filepath.Join(live, fmt.Sprintf("target-%03d", index))
		if err := os.WriteFile(path, []byte("present"), 0o600); err != nil {
			tb.Fatal(err)
		}
		targets = append(targets, namespaceTargetRecord(index, "class", fmt.Sprintf("id-%03d", index), path))
	}
	return targets
}
