//go:build unix

package envprofile

import (
	"errors"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/transaction"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Actual Resolve reports postcommit scratch cleanup separately from rollback.
// Returned preparation faults restore the whole home even if scratch cleanup
// also fails. No credential/session payload is used or copied by this control.
func TestProvisioningOriginHomeCleanup(t *testing.T) {
	for _, mode := range []string{"healthy", "committed-cleanup", "refused-cleanup"} {
		t.Run(mode, func(t *testing.T) {
			fx, req, origin := recordedOriginFixture(t, "a")
			target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			if err := os.Remove(filepath.Join(target, "AGENTS.md")); err != nil {
				t.Fatal(err)
			}
			before := nativeSnapshot(t, target)
			tmp := t.TempDir()
			t.Setenv("TMPDIR", tmp)
			req.Repair = true
			var stages []string
			hit := false
			fault := errors.New("home cleanup prepared fault")
			req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
				if e.Point == transaction.PointPrepared {
					hit = true
					var err error
					stages, err = filepath.Glob(filepath.Join(tmp, "curator-home-unit-*"))
					if err != nil || len(stages) == 0 {
						t.Fatalf("actual home staging not reached: %v %v", stages, err)
					}
					if mode != "healthy" {
						for _, path := range stages {
							// Transaction preparation has consumed its staged payloads.
							// Keep one disposable entry so cleanup must unlink inside
							// this protected directory rather than remove an empty one.
							if err := os.WriteFile(filepath.Join(path, "cleanup-probe"), []byte("fixture"), 0600); err != nil {
								t.Fatal(err)
							}
							if err := os.Chmod(path, 0500); err != nil {
								t.Fatal(err)
							}
						}
					}
					if mode == "refused-cleanup" {
						return fault
					}
				}
				return nil
			}})}
			result, err := Resolve(req)
			for _, stage := range stages {
				if mode != "healthy" {
					if e := os.Chmod(stage, 0700); e != nil {
						t.Error(e)
					}
					if e := os.RemoveAll(stage); e != nil {
						t.Error(e)
					}
				}
			}
			if !hit {
				t.Fatalf("actual publication missed: %v", err)
			}
			if mode == "healthy" {
				if err != nil || result == nil || len(result.Document) == 0 {
					t.Fatalf("healthy publication: %v", err)
				}
			} else {
				if !errors.Is(err, fs.ErrPermission) || !strings.Contains(err.Error(), "cleanup") || errors.Is(err, fault) != (mode == "refused-cleanup") {
					t.Fatalf("cleanup phase lost: %v", err)
				}
				if result != nil && len(result.Document) > 0 {
					t.Fatal("cleanup refusal emitted fragment")
				}
				if mode == "refused-cleanup" && nativeSnapshot(t, target) != before {
					t.Fatal("precommit failure did not restore old home")
				}
			}
			if string(nativeRead(t, filepath.Join(target, envorigin.Name))) != string(origin) {
				t.Fatal("cleanup recaptured origin")
			}
			if mode == "committed-cleanup" {
				req.transactionOptions = nil
				req.Repair = false
				if _, err := Resolve(req); err != nil {
					t.Fatalf("committed state was reported as rollback: %v", err)
				}
			}
		})
	}
}
