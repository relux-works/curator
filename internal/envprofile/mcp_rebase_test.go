package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
)

// Protected publication retains the landed compensation capture for writes and
// removals. A later operator edit refuses compensation without touching either
// file; unchanged postimages restore the exact old bytes and removed record.
func TestMCPProtectedPublicationCompensation(t *testing.T) {
	for _, edited := range []bool{false, true} {
		t.Run(map[bool]string{false: "unchanged", true: "operator-edit"}[edited], func(t *testing.T) {
			home := t.TempDir()
			write := filepath.Join(home, "write")
			remove := filepath.Join(home, "remove")
			for _, path := range []string{write, remove} {
				file, err := privatedir.Create(path, 0o600)
				if err != nil {
					t.Fatal(err)
				}
				_, err = file.WriteString("old")
				closeErr := file.Close()
				if err != nil || closeErr != nil {
					t.Fatal(err, closeErr)
				}
			}
			op, err := beginOperation(home)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = op.close() }()
			op.undo = &globalUndo{managerHome: home, entries: map[string]*globalImage{}}
			if err := op.publishProtected(map[string][]byte{write: []byte("new")}, nil, func() error { return nil }, remove); err != nil {
				t.Fatal(err)
			}
			for _, path := range []string{write, remove} {
				image := op.undo.entries[path]
				if image == nil || string(image.data) != "old" {
					t.Fatal("protected publication lost compensation preimage")
				}
			}
			if edited {
				if err := os.WriteFile(write, []byte("operator"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			err = op.undo.restore()
			if edited {
				if err == nil || !strings.Contains(err.Error(), "changed") {
					t.Fatalf("operator edit compensation refusal = %v", err)
				}
				if raw, e := os.ReadFile(write); e != nil || string(raw) != "operator" {
					t.Fatalf("operator bytes changed: %q %v", raw, e)
				}
				if _, e := os.Lstat(remove); !os.IsNotExist(e) {
					t.Fatalf("refusal partially restored removal: %v", e)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				for _, path := range []string{write, remove} {
					if raw, e := os.ReadFile(path); e != nil || string(raw) != "old" {
						t.Fatalf("compensation bytes = %q %v", raw, e)
					}
				}
			}
		})
	}
}
