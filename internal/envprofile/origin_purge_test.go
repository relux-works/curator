package envprofile

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/transaction"
)

// Explicit later purge reaches the real absent-source owner using only its
// permanent removal receipt. Unreadable opaque session bytes and an external
// link are deleted as entries; their contents/targets are not read or copied.
// Substitution, a new namespace or partial installed profile refuses before
// moving any home. No native account or provider process is involved.
func TestProvisioningOriginRetainedPurgeOwner(t *testing.T) {
	for _, state := range []string{"healthy", "substituted-origin", "new-namespace", "partial-installed"} {
		t.Run(state, func(t *testing.T) {
			fx, _, _, _, _ := removedOriginFixture(t)
			target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			external := t.TempDir()
			leaf := filepath.Join(external, "unrelated.txt")
			const payload = "unrelated opaque fixture must survive"
			if err := os.WriteFile(leaf, []byte(payload), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(external, filepath.Join(target, "external-session-link")); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(target, "opaque-session"), []byte("unreadable owned fixture payload"), 0000); err != nil {
				t.Fatal(err)
			}
			switch state {
			case "substituted-origin":
				if err := os.WriteFile(filepath.Join(target, envorigin.Name), []byte("{}"), 0600); err != nil {
					t.Fatal(err)
				}
			case "new-namespace":
				if err := os.Mkdir(filepath.Join(EnvRoot(fx.home), fx.profile, "unexpected-adapter"), 0700); err != nil {
					t.Fatal(err)
				}
			case "partial-installed":
				if err := os.Mkdir(ProfileDir(fx.home, fx.profile), 0700); err != nil {
					t.Fatal(err)
				}
			}
			manager, err := envunit.ObserveManager(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			_, before, err := envlifecycle.Read(manager)
			if err != nil {
				t.Fatal(err)
			}
			err = Remove(fx.home, fx.profile, true)
			if state != "healthy" {
				if err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) {
					t.Fatalf("intended retained purge refusal: %v", err)
				}
				if _, err := os.Lstat(target); err != nil {
					t.Fatalf("refusal moved retained home: %v", err)
				}
				_, after, err := envlifecycle.Read(manager)
				if err != nil || string(before) != string(after) {
					t.Fatal("refusal changed permanent receipt")
				}
				return
			}
			if err != nil {
				t.Fatalf("actual later purge: %v", err)
			}
			if _, err := os.Lstat(target); !os.IsNotExist(err) {
				t.Fatalf("purge left live target: %v", err)
			}
			if raw, err := os.ReadFile(leaf); err != nil || string(raw) != payload {
				t.Fatal("purge followed external session link")
			}
			ledger, _, err := envlifecycle.Read(manager)
			if err != nil || len(ledger.Entries) != 1 || ledger.Entries[0].Purged == nil {
				t.Fatalf("terminal absence receipt missing: %v", err)
			}
			if err := Remove(fx.home, fx.profile, true); err != nil {
				t.Fatalf("read-only terminal purge retry: %v", err)
			}
		})
	}
}

// The actual purge owner restores the same home/receipt on precommit failure.
// Abrupt rename-window stops and committed cleanup failures remain durable and
// recover on a fresh engine; a cleanup failure is never reported as rollback.
// Stops are in-process panic models, not process/power-loss qualification.
func TestProvisioningOriginRetainedPurgeRecovery(t *testing.T) {
	for _, mode := range []string{"returned-before-commit", "restart-after-rename", "returned-cleanup", "restart-cleanup"} {
		t.Run(mode, func(t *testing.T) {
			fx, _, _, _, before := removedOriginFixture(t)
			target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			manager, err := envunit.ObserveManager(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			_, receipt, err := envlifecycle.Read(manager)
			if err != nil {
				t.Fatal(err)
			}
			reached := false
			failure := errors.New("fixture opaque purge fault")
			factory := newOperationEngine
			newOperationEngine = func(home string, options ...transaction.Option) (*transaction.Engine, error) {
				return factory(home, append(options, transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
					point := transaction.PointAfterQuarantine
					if strings.Contains(mode, "cleanup") {
						point = transaction.PointDuringQuarantineRemoval
					}
					if event.Point != point || reached {
						return nil
					}
					reached = true
					if strings.HasPrefix(mode, "restart") {
						panic("fixture opaque purge restart")
					}
					return failure
				}}))...)
			}
			t.Cleanup(func() { newOperationEngine = factory })
			if strings.HasPrefix(mode, "restart") {
				func() {
					defer func() {
						if value := recover(); value != "fixture opaque purge restart" {
							t.Fatalf("wrong stop: %v", value)
						}
					}()
					_ = Remove(fx.home, fx.profile, true)
				}()
			} else {
				err := Remove(fx.home, fx.profile, true)
				if !errors.Is(err, failure) {
					t.Fatalf("intended purge fault: %v", err)
				}
			}
			newOperationEngine = factory
			if !reached {
				t.Fatal("purge continuation fault was not reached")
			}
			if mode == "returned-before-commit" {
				_, after, err := envlifecycle.Read(manager)
				if err != nil || !bytes.Equal(receipt, after) || nativeSnapshot(t, target) != before {
					t.Fatal("precommit purge did not restore same home and receipt")
				}
			} else if strings.Contains(mode, "cleanup") {
				if _, err := os.Lstat(target); !os.IsNotExist(err) {
					t.Fatalf("committed cleanup was falsely rolled back: %v", err)
				}
				ledger, _, err := envlifecycle.Read(manager)
				if err != nil || ledger.Entries[0].Purged == nil {
					t.Fatalf("committed purge lacks terminal receipt: %v", err)
				}
			}
			if err := Remove(fx.home, fx.profile, true); err != nil {
				t.Fatalf("fresh-engine exact purge resume: %v", err)
			}
			entries, err := os.ReadDir(filepath.Join(fx.home, "state/transactions/v1"))
			if err != nil || len(entries) != 0 {
				t.Fatalf("purge recovery left journal: %v %v", entries, err)
			}
			if _, err := os.Lstat(target); !os.IsNotExist(err) {
				t.Fatalf("recovered purge left home: %v", err)
			}
		})
	}
}
