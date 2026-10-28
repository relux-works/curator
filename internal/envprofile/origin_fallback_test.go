package envprofile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/transaction"
)

// Actual repair replaces a previously recorded copy as a complete owned tree.
// A returned fault restores the original tree and marker, including excluded
// fixture bytes; success removes stale excluded bytes. It never snapshots the
// home/session tree and does not prove platform symlink-unavailable transport.
func TestProvisioningOriginCopiedSkillUnit(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "healthy", true: "returned-fault"}[fail], func(t *testing.T) {
			fx, req, _ := recordedOriginFixture(t, "a")
			home := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			marker := readManagedMarker(t, fx, "codex_cli")
			surface := marker.Surfaces[envmarker.SurfaceSkills]
			if len(surface.Paths) != 1 {
				t.Fatalf("expected one skill: %+v", surface)
			}
			rel := surface.Paths[0]
			live := filepath.Join(home, filepath.FromSlash(rel))
			source, err := os.Readlink(live)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(live); err != nil {
				t.Fatal(err)
			}
			if err := copyTree(source, home, rel); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(live, ".research"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(live, ".research", "old.txt"), []byte("owned fixture stale projection"), 0600); err != nil {
				t.Fatal(err)
			}
			copies := append([]envmarker.Copy(nil), (*surface.Copies)...)
			copies = append(copies, envmarker.Copy{Path: rel, Reason: envmarker.ReasonSymlinkFallback})
			surface.Copies = &copies
			marker.Surfaces[envmarker.SurfaceSkills] = surface
			raw, err := marker.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(home, envmarker.Name), raw, 0600); err != nil {
				t.Fatal(err)
			}
			before := nativeSnapshot(t, home)
			fault := errors.New("fixture copied-skill unit fault")
			reached := false
			if fail {
				req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
					if event.Point == transaction.PointAfterInstall && event.Class == "owned-tree" {
						reached = true
						return fault
					}
					return nil
				}})}
			}
			req.Repair = true
			result, err := Resolve(req)
			if fail {
				if !reached || !errors.Is(err, fault) {
					t.Fatalf("intended owned-tree fault not reached: %v", err)
				}
				if result != nil && len(result.Document) != 0 {
					t.Fatal("fault emitted fragment")
				}
				if nativeSnapshot(t, home) != before {
					t.Fatal("copied-skill refusal changed managed preimages")
				}
				return
			}
			if err != nil || result == nil || len(result.Document) == 0 {
				t.Fatalf("copied-skill continuation: %v", err)
			}
			if _, err := os.Lstat(filepath.Join(live, ".research")); !os.IsNotExist(err) {
				t.Fatalf("retained excluded projection bytes: %v", err)
			}
			updated := readManagedMarker(t, fx, "codex_cli")
			if !recordedSkillFallback(updated, strings.ReplaceAll(rel, string(filepath.Separator), "/")) {
				t.Fatal("lost recorded copy reason")
			}
		})
	}
}
