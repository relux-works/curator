//go:build windows

package transaction

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildJournalRejectsCaseInsensitiveLiveAliases(t *testing.T) {
	root := t.TempDir()
	sourceA := filepath.Join(root, "stage", "a")
	sourceB := filepath.Join(root, "stage", "b")
	mustWrite(t, sourceA, "new-a")
	mustWrite(t, sourceB, "new-b")
	upper := filepath.Join(root, "live", "Target")
	lower := filepath.Join(root, "live", "target")
	plan := Plan{TransactionID: "txn-windows-case-alias", ProjectIdentity: "/project", Targets: []Target{
		{Class: "a", Identifier: "upper", LivePath: upper, StagedSource: sourceA, PreimageDigest: DigestAbsent},
		{Class: "b", Identifier: "lower", LivePath: lower, StagedSource: sourceB, PreimageDigest: DigestAbsent},
	}}
	assertPlanRejectedBeforeJournal(t, filepath.Join(root, "home"), plan)
}

func TestWindowsRegularAndTreeDurability(t *testing.T) {
	root := t.TempDir()
	regular := filepath.Join(root, "regular")
	mustWrite(t, regular, "durable")
	if err := syncRegular(regular); err != nil {
		t.Fatalf("sync regular file with write-capable handle: %v", err)
	}

	tree := filepath.Join(root, "tree")
	mustWrite(t, filepath.Join(tree, "nested", "file"), "durable tree")
	if err := syncTree(tree); err != nil {
		t.Fatalf("sync regular files and directories in tree: %v", err)
	}
}

func TestWindowsJournalRemovalUsesPersistedModeAndPreservesChangedBytes(t *testing.T) {
	newJournal := func(t *testing.T, transactionID string) (*Engine, *Journal) {
		t.Helper()
		root := t.TempDir()
		engine := mustEngine(t, filepath.Join(root, "home"))
		journal, _, err := engine.buildJournal(Plan{
			TransactionID:   transactionID,
			ProjectIdentity: "/project",
			Targets: []Target{{
				Class:          "runtime",
				Identifier:     "journal",
				LivePath:       filepath.Join(root, "live", "target"),
				PreimageDigest: DigestAbsent,
			}},
		})
		if err != nil {
			t.Fatalf("build journal: %v", err)
		}
		if err := engine.saveJournal(journal); err != nil {
			t.Fatalf("save journal: %v", err)
		}
		return engine, journal
	}

	t.Run("canonical", func(t *testing.T) {
		engine, journal := newJournal(t, "txn-windows-journal-remove")
		path := engine.journalPath(journal.TransactionID)
		info, err := os.Lstat(path)
		if err != nil {
			t.Fatalf("stat journal: %v", err)
		}
		if got := info.Mode().Perm(); got != journalMode() {
			t.Fatalf("persisted journal mode = %#o, want %#o", got, journalMode())
		}
		if err := engine.removeJournalDurably(journal); err != nil {
			t.Fatalf("remove canonical journal durably: %v", err)
		}
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("canonical journal remains after removal: %v", err)
		}
	})

	t.Run("changed bytes", func(t *testing.T) {
		engine, journal := newJournal(t, "txn-windows-journal-changed")
		path := engine.journalPath(journal.TransactionID)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0) // #nosec G304 -- test-owned journal corruption
		if err != nil {
			t.Fatalf("open journal for corruption: %v", err)
		}
		if _, err := file.WriteString("foreign"); err != nil {
			_ = file.Close()
			t.Fatalf("change journal bytes: %v", err)
		}
		if err := errors.Join(file.Sync(), file.Close()); err != nil {
			t.Fatalf("persist changed journal bytes: %v", err)
		}
		if err := engine.removeJournalDurably(journal); !errors.Is(err, ErrImplementationCorruption) {
			t.Fatalf("remove changed journal error = %v, want implementation-corruption", err)
		}
		got, err := os.ReadFile(path)
		if err != nil || !bytes.HasSuffix(got, []byte("foreign")) {
			t.Fatalf("changed journal was not preserved: bytes=%q read_error=%v", got, err)
		}
	})
}
