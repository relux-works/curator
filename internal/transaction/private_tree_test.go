package transaction

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
)

// Only enumerated build-artifact/snapshot trees use this flag. Privacy is
// durable across a fresh engine; a substituted link or foreign-read grant
// refuses before changing the old live entry. This is not a Windows witness.
func TestPrivateTreePublication(t *testing.T) {
	for _, restart := range []bool{false, true} {
		for _, plant := range []string{"valid", "source-link", "staged-grant"} {
			t.Run(map[bool]string{false: "commit", true: "recovery"}[restart]+"/"+plant, func(t *testing.T) {
				root := t.TempDir()
				live := filepath.Join(root, "live")
				source := filepath.Join(root, "source")
				if err := privatedir.Make(source); err != nil {
					t.Fatal(err)
				}
				f, err := privatedir.Create(filepath.Join(source, "artifact"), 0600)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = f.WriteString("synthetic build bytes"); err != nil {
					t.Fatal(err)
				}
				if err = f.Close(); err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(live, []byte("old entry"), 0600); err != nil {
					t.Fatal(err)
				}
				pre, err := DigestTarget(KindBytes, live)
				if err != nil {
					t.Fatal(err)
				}
				if plant == "source-link" {
					if err = os.Symlink(filepath.Join(root, "unrelated"), filepath.Join(source, "link")); err != nil {
						t.Fatal(err)
					}
				}
				plan := Plan{TransactionID: "private-tree", ProjectIdentity: "fixture", Targets: []Target{{PrivateTree: true, Class: "05-external-cache", Identifier: live, Kind: KindBytes, LivePath: live, StagedSource: source, PreimageDigest: pre}}}
				if err = privatedir.Make(filepath.Join(root, "manager")); err != nil {
					t.Fatal(err)
				}
				engine := mustEngine(t, filepath.Join(root, "manager"))
				journal, err := engine.Prepare(testLock{}, plan)
				if plant == "source-link" {
					if err == nil || !strings.Contains(err.Error(), "private transaction tree source") {
						t.Fatalf("link source admitted: %v", err)
					}
					if got := mustRead(t, live); got != "old entry" {
						t.Fatalf("source refusal changed live: %q", got)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if !journal.Targets[0].PrivateTree {
					t.Fatal("tree protection lost from journal")
				}
				sidecar := journal.Targets[0].StagedPath
				if err = validatePrivateTree(sidecar); err != nil {
					t.Fatalf("sidecar not private: %v", err)
				}
				if plant == "staged-grant" {
					broadenPrivateFileForTest(t, filepath.Join(sidecar, "artifact"))
				}
				if restart {
					err = mustEngine(t, filepath.Join(root, "manager")).Recover(testLock{})
				} else {
					err = engine.Commit(testLock{}, journal.TransactionID)
				}
				if plant == "staged-grant" {
					if err == nil || !strings.Contains(err.Error(), "private transaction tree replacement") {
						t.Fatalf("shared tree admitted: %v", err)
					}
					if got := mustRead(t, live); got != "old entry" {
						t.Fatalf("tree refusal changed live: %q", got)
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					if got := mustRead(t, filepath.Join(live, "artifact")); got != "synthetic build bytes" {
						t.Fatalf("tree not published: %q", got)
					}
					if err = validatePrivateTree(live); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
	}
}
