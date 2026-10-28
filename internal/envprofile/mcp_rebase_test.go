package envprofile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/transaction"
)

// The actual protected publisher enlists writes and removals in one journal.
// A returned fault restores both preimages. A foreign replacement is preserved
// and leaves a specifically refused recovery unit, never callback compensation.
// This is an isolated publication/restart witness, not MCP provider execution.
func TestMCPProtectedPublicationCompensation(t *testing.T) {
	for _, state := range []string{"commit", "rollback", "operator-edit"} {
		t.Run(state, func(t *testing.T) {
			home := t.TempDir()
			write := filepath.Join(home, "write")
			remove := filepath.Join(home, "remove")
			for _, path := range []string{write, remove} {
				file, err := privatedir.Create(path, 0600)
				if err != nil {
					t.Fatal(err)
				}
				_, err = file.WriteString("old")
				closeErr := file.Close()
				if err != nil || closeErr != nil {
					t.Fatal(err, closeErr)
				}
			}
			hit := false
			fault := errors.New("protected-publisher-fault")
			op, err := beginOperation(home, transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
				if event.Point == transaction.PointAfterInstall && event.LivePath == write {
					hit = true
					if state == "operator-edit" {
						if err := os.WriteFile(write, []byte("operator"), 0600); err != nil {
							return err
						}
					}
					if state != "commit" {
						return fault
					}
				}
				return nil
			}}))
			if err != nil {
				t.Fatal(err)
			}
			err = op.publishProtected(map[string][]byte{write: []byte("new")}, nil, func() error { return nil }, remove)
			_ = op.close()
			if !hit {
				t.Fatalf("actual protected publisher not reached: %v", err)
			}
			if state == "operator-edit" {
				if !errors.Is(err, fault) || !errors.Is(err, transaction.ErrImplementationCorruption) {
					t.Fatalf("intended replacement refusal: %v", err)
				}
				if raw, e := os.ReadFile(write); e != nil || string(raw) != "operator" {
					t.Fatalf("foreign replacement changed: %q %v", raw, e)
				}
				if _, e := os.Lstat(remove); !os.IsNotExist(e) {
					t.Fatalf("refusal partially restored removal: %v", e)
				}
				pending := nativeSnapshot(t, filepath.Join(home, "state", "transactions", "v1"))
				if pending == "" {
					t.Fatal("refusal lost pending recovery unit")
				}
				fresh, e := beginOperation(home)
				if e == nil {
					_ = fresh.close()
					t.Fatal("substituted postimage admitted on restart")
				}
				if !errors.Is(e, transaction.ErrImplementationCorruption) || !strings.Contains(e.Error(), "rollback refused unknown current bytes") {
					t.Fatalf("wrong restart refusal: %v", e)
				}
				if nativeSnapshot(t, filepath.Join(home, "state", "transactions", "v1")) != pending {
					t.Fatal("restart refusal changed pending unit")
				}
				return
			}
			if state == "rollback" {
				if !errors.Is(err, fault) {
					t.Fatalf("wrong returned fault: %v", err)
				}
				for _, path := range []string{write, remove} {
					if raw, e := os.ReadFile(path); e != nil || string(raw) != "old" {
						t.Fatalf("lost rollback preimage: %q %v", raw, e)
					}
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if raw, e := os.ReadFile(write); e != nil || string(raw) != "new" {
					t.Fatalf("write did not commit: %q %v", raw, e)
				}
				if _, e := os.Lstat(remove); !os.IsNotExist(e) {
					t.Fatalf("removal did not commit: %v", e)
				}
			}
			entries, e := os.ReadDir(filepath.Join(home, "state", "transactions", "v1"))
			if e != nil || len(entries) != 0 {
				t.Fatalf("completed publication cleanup: %v %v", entries, e)
			}
		})
	}
}
