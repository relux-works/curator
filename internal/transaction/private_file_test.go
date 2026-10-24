package transaction

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
)

// Private replacements are private while staged and after commit or restart.
// A foreign-read grant on either the source or prepared sidecar is refused,
// preserving the old live bytes; a fresh engine must retain the same gate.
func TestPrivateFilePublication(t *testing.T) {
	for _, restart := range []bool{false, true} {
		for _, plant := range []string{"valid", "source", "staged"} {
			t.Run(map[bool]string{false: "commit", true: "recovery"}[restart]+"/"+plant, func(t *testing.T) {
				root := t.TempDir()
				home := filepath.Join(root, "home")
				target := fileTarget(t, "private", "one", filepath.Join(root, "live", "one"), filepath.Join(root, "source", "one"), "old", "new")
				target.PrivateFile = true
				if err := privatedir.ProtectFile(target.StagedSource); err != nil {
					t.Fatal(err)
				}
				if plant == "source" {
					broadenPrivateFileForTest(t, target.StagedSource)
				}
				engine := mustEngine(t, home)
				journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "private-file", ProjectIdentity: "profile", Targets: []Target{target}})
				if plant == "source" {
					if err == nil || !strings.Contains(err.Error(), "private transaction source") {
						t.Fatalf("shared source accepted: %v", err)
					}
					if got := mustRead(t, target.LivePath); got != "old" {
						t.Fatalf("source refusal changed live bytes: %q", got)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if !journal.Targets[0].PrivateFile {
					t.Fatal("private contract absent from journal")
				}
				staged := journal.Targets[0].StagedPath
				if err := privatedir.ValidateFile(staged); err != nil {
					t.Fatalf("staged file not private: %v", err)
				}
				if plant == "staged" {
					broadenPrivateFileForTest(t, staged)
				}
				if restart {
					err = mustEngine(t, home).Recover(testLock{})
				} else {
					err = engine.Commit(testLock{}, journal.TransactionID)
				}
				if plant == "staged" {
					if err == nil || !strings.Contains(err.Error(), "private transaction replacement") {
						t.Fatalf("shared staged file accepted: %v", err)
					}
					if got := mustRead(t, target.LivePath); got != "old" {
						t.Fatalf("staged refusal changed live bytes: %q", got)
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					if got := mustRead(t, target.LivePath); got != "new" {
						t.Fatalf("private publication = %q", got)
					}
					if err := privatedir.ValidateFile(target.LivePath); err != nil {
						t.Fatalf("published file not private: %v", err)
					}
				}
			})
		}
	}
}

// A private-file contract cannot silently become a deletion or directory-tree
// publication. Existing live bytes are untouched by both invalid plans.
func TestPrivateFileRejectsOtherShapes(t *testing.T) {
	for _, shape := range []string{"absent", "directory"} {
		t.Run(shape, func(t *testing.T) {
			root := t.TempDir()
			target := fileTarget(t, "private", "one", filepath.Join(root, "live", "one"), filepath.Join(root, "source", "one"), "old", "new")
			target.PrivateFile = true
			if shape == "absent" {
				target.StagedSource = ""
			} else {
				if err := os.Remove(target.StagedSource); err != nil {
					t.Fatal(err)
				}
				if err := privatedir.Make(target.StagedSource); err != nil {
					t.Fatal(err)
				}
			}
			_, err := mustEngine(t, filepath.Join(root, "home")).Prepare(testLock{}, Plan{TransactionID: "private-shape", ProjectIdentity: "profile", Targets: []Target{target}})
			if err == nil || !strings.Contains(err.Error(), "private file requires a present byte replacement") && !strings.Contains(err.Error(), "private transaction source") {
				t.Fatalf("non-file private plan refusal = %v", err)
			}
			if got := mustRead(t, target.LivePath); got != "old" {
				t.Fatalf("invalid private plan changed bytes: %q", got)
			}
		})
	}
}

// A grant planted by the last caller boundary callback must still refuse the
// publication. This distinguishes the final privacy recheck from planning and
// the earlier staged digest check, while retaining the original live bytes.
func TestPrivateFileBoundaryMutation(t *testing.T) {
	root := t.TempDir()
	target := fileTarget(t, "private", "one", filepath.Join(root, "live", "one"), filepath.Join(root, "source", "one"), "old", "new")
	target.PrivateFile = true
	if err := privatedir.ProtectFile(target.StagedSource); err != nil {
		t.Fatal(err)
	}
	var staged string
	planted := false
	engine := mustEngine(t, filepath.Join(root, "home"))
	journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "private-boundary", ProjectIdentity: "profile", Targets: []Target{target}, BoundaryCheck: func(_ int, path string) error {
		if _, err := os.Lstat(path); os.IsNotExist(err) {
			broadenPrivateFileForTest(t, staged)
			planted = true
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	staged = journal.Targets[0].StagedPath
	err = engine.Commit(testLock{}, journal.TransactionID)
	if !planted || err == nil || !strings.Contains(err.Error(), "private transaction replacement") {
		t.Fatalf("final boundary privacy refusal = %v; applied=%v", err, planted)
	}
	if got := mustRead(t, target.LivePath); got != "old" {
		t.Fatalf("boundary refusal changed live bytes: %q", got)
	}
}
