//go:build windows

package transaction

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
)

// Claim: the real transaction stages and publishes bytes and nested trees
// with protected owner-only ACLs; journal files/directories have the same shape.
// This checks the actual copies, not merely the private staging sources.
func TestWindowsTransactionPrivatePublication(t *testing.T) {
	for _, tree := range []bool{false, true} {
		t.Run(map[bool]string{false: "bytes", true: "tree"}[tree], func(t *testing.T) {
			root := t.TempDir()
			source := filepath.Join(root, "source")
			live := filepath.Join(root, "live")
			if tree {
				mustWrite(t, filepath.Join(source, "bin", "dispatch"), "private dispatch")
			} else {
				mustWrite(t, source, "private agent")
			}
			engine := mustEngine(t, filepath.Join(root, "home"))
			target := Target{Class: "generated", Identifier: "private", LivePath: live, StagedSource: source, PreimageDigest: DigestAbsent}
			journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "private-publication", ProjectIdentity: "private", Targets: []Target{target}})
			if err != nil {
				t.Fatal(err)
			}
			for _, dir := range []string{filepath.Join(root, "home"), filepath.Dir(engine.journalRoot), engine.journalRoot} {
				if err := privatedir.Validate(dir); err != nil {
					t.Fatalf("journal directory lost private DACL: %v", err)
				}
			}
			if err := privatedir.ValidateFile(engine.journalPath(journal.TransactionID)); err != nil {
				t.Fatalf("journal file lost private DACL: %v", err)
			}
			assertPrivatePublication(t, journal.Targets[0].StagedPath)
			if err := engine.Commit(testLock{}, journal.TransactionID); err != nil {
				t.Fatal(err)
			}
			assertPrivatePublication(t, live)
			path := live
			if tree {
				path = filepath.Join(live, "bin", "dispatch")
			}
			if raw, err := os.ReadFile(path); err != nil || len(raw) == 0 {
				t.Fatalf("private publication lost payload: %q %v", raw, err)
			}
		})
	}
}

func assertPrivatePublication(t *testing.T, root string) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if err := privatedir.Validate(path); err != nil {
				t.Fatalf("published directory lost private DACL: %v", err)
			}
		} else if err := privatedir.ValidateFile(path); err != nil {
			t.Fatalf("published file lost private DACL: %v", err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
