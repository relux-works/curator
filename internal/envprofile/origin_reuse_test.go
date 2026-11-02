package envprofile

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/relux-works/curator/internal/managerlock"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/transaction"
)

// Actual same-name installation cannot adopt a removed retained home. After
// explicit purge, fresh records and exact terminal receipt consumption commit
// together, before any new origin capture. A source-stage fault restores full
// installed absence and the terminal receipt. The final Resolve uses only a
// synthetic account lookup and empty credential fixture, not a real account.
func TestProvisioningOriginPurgedNameReuse(t *testing.T) {
	for _, mode := range []string{"retained-refusal", "healthy", "source-fault"} {
		t.Run(mode, func(t *testing.T) {
			fx, req, _, _, before := removedOriginFixture(t)
			target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			if mode != "retained-refusal" {
				if err := Remove(fx.home, fx.profile, true); err != nil {
					t.Fatal(err)
				}
			}
			identity, err := envunit.ObserveManager(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			_, prior, err := envlifecycle.Read(identity)
			if err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			writePackage(t, root, "fresh-reuse-context", "1.0.0", "fresh context\n")
			saved := newOperationEngine
			t.Cleanup(func() { newOperationEngine = saved })
			hit := false
			planted := errors.New("reuse-source-fault")
			if mode == "source-fault" {
				newOperationEngine = func(h string, opts ...transaction.Option) (*transaction.Engine, error) {
					return saved(h, append(opts, transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
						if event.Point == transaction.PointAfterInstall && event.LivePath == sourcePath(fx.home, fx.profile) {
							hit = true
							return planted
						}
						return nil
					}}))...)
				}
			}
			info, activated, _, err := Install(fx.home, InstallOptions{Operand: root, As: fx.profile})
			newOperationEngine = saved
			if mode != "healthy" {
				if mode == "source-fault" {
					if !hit || !errors.Is(err, planted) {
						t.Fatalf("actual reuse fault: %t %v", hit, err)
					}
				} else if err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) {
					t.Fatalf("retained same-name install was admitted: %v", err)
				}
				if _, e := os.Lstat(ProfileDir(fx.home, fx.profile)); !os.IsNotExist(e) {
					t.Fatalf("failed reuse left installed records: %v", e)
				}
				_, after, e := envlifecycle.Read(identity)
				if e != nil || string(after) != string(prior) {
					t.Fatalf("failed reuse changed receipt: %v", e)
				}
				if mode == "retained-refusal" && nativeSnapshot(t, target) != before {
					t.Fatal("refused install changed retained state")
				}
				return
			}
			if err != nil || activated || info.Name != fx.profile {
				t.Fatalf("actual absent-name reinstall: %+v %t %v", info, activated, err)
			}
			ledger, _, err := envlifecycle.Read(identity)
			if err != nil || len(ledger.Entries) != 0 {
				t.Fatalf("reuse did not consume terminal receipt: %+v %v", ledger, err)
			}
			if _, err := os.Lstat(target); !os.IsNotExist(err) {
				t.Fatalf("install recaptured home: %v", err)
			}
			b := t.TempDir()
			if err := os.WriteFile(filepath.Join(b, "auth.json"), nil, 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("CODEX_HOME", b)
			req.origin = nil
			req.Repair = true
			captures := 0
			req.originAccountLookup = func() (envorigin.Account, error) {
				captures++
				ledger, _, err := envlifecycle.Read(identity)
				if err != nil || len(ledger.Entries) != 0 {
					t.Fatalf("capture preceded receipt consumption: %v", err)
				}
				if _, err := os.Lstat(target); !os.IsNotExist(err) {
					t.Fatalf("mutation preceded fresh capture: %v", err)
				}
				identity, err := envorigin.LookupIdentity()
				identity.Home = b
				return identity, err
			}
			result, err := Resolve(req)
			if err != nil || result == nil || len(result.Document) == 0 || captures != 1 {
				t.Fatalf("genuine fresh provisioning: captures=%d %v", captures, err)
			}
			link, err := os.Readlink(filepath.Join(target, "auth.json"))
			if err != nil || link != filepath.Join(b, "auth.json") {
				t.Fatalf("fresh origin was not B: %q %v", link, err)
			}
		})
	}
}

// Fresh recovery resumes an explicitly saved forward reuse, restores a saved
// rollback, or finishes committed cleanup. Every stop is bound to the actual
// reuse unit, so an earlier store transaction cannot qualify this control.
// This models restart with a fresh engine, not an OS process/account witness.
func TestProvisioningOriginPurgedReuseRecovery(t *testing.T) {
	for _, mode := range []string{"forward", "rollback", "cleanup"} {
		t.Run(mode, func(t *testing.T) {
			fx, _, _, _, _ := removedOriginFixture(t)
			if err := Remove(fx.home, fx.profile, true); err != nil {
				t.Fatal(err)
			}
			identity, err := envunit.ObserveManager(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			_, prior, err := envlifecycle.Read(identity)
			if err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			writePackage(t, root, "reuse-recovery-context", "1.0.0", "fresh context\n")
			saved := newOperationEngine
			t.Cleanup(func() { newOperationEngine = saved })
			hit := false
			reuseID := ""
			newOperationEngine = func(h string, opts ...transaction.Option) (*transaction.Engine, error) {
				return saved(h, append(opts, transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
					if e.Point == transaction.PointPrepared {
						raw, err := os.ReadFile(filepath.Join(h, "state", "transactions", "v1", e.TransactionID+".json"))
						if err != nil {
							t.Fatal(err)
						}
						var j transaction.Journal
						if err := json.Unmarshal(raw, &j); err != nil {
							t.Fatal(err)
						}
						if j.Unit != nil && j.Unit.Action == envunit.Reuse {
							reuseID = e.TransactionID
						}
					}
					if e.TransactionID != reuseID || reuseID == "" {
						return nil
					}
					if e.Point == transaction.PointAfterInstall && e.LivePath == sourcePath(fx.home, fx.profile) {
						if mode == "rollback" {
							return errors.New("reuse-enter-rollback")
						}
						if mode == "forward" {
							hit = true
							panic("reuse-stop")
						}
					}
					if mode == "rollback" && e.Point == transaction.PointAfterRestore && e.LivePath == sourcePath(fx.home, fx.profile) || mode == "cleanup" && e.Point == transaction.PointBeforeCleanup {
						hit = true
						panic("reuse-stop")
					}
					return nil
				}}))...)
			}
			func() {
				defer func() {
					newOperationEngine = saved
					if got := recover(); got != "reuse-stop" {
						t.Fatalf("intended boundary: %v; err=%v", got, err)
					}
				}()
				_, _, _, err = Install(fx.home, InstallOptions{Operand: root, As: fx.profile})
			}()
			if !hit {
				t.Fatalf("actual reuse boundary not reached: %v", err)
			}
			engine, err := saved(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			lm, err := managerlock.New(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			held, err := lm.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			err = engine.Recover(held)
			_ = held.Close()
			if err != nil {
				t.Fatalf("fresh reuse recovery: %v", err)
			}
			ledger, after, err := envlifecycle.Read(identity)
			if err != nil {
				t.Fatal(err)
			}
			if mode != "rollback" {
				if len(ledger.Entries) != 0 {
					t.Fatal("committed reuse restored old receipt")
				}
				if _, err := readSource(fx.home, fx.profile); err != nil {
					t.Fatal(err)
				}
			} else {
				if string(after) != string(prior) {
					t.Fatal("rollback recovery changed terminal receipt")
				}
				if _, err := os.Lstat(ProfileDir(fx.home, fx.profile)); !os.IsNotExist(err) {
					t.Fatal("rollback recovery left installed namespace")
				}
			}
			if _, err := os.Lstat(ManagedHomeDir(fx.home, fx.profile, "codex_cli")); !os.IsNotExist(err) {
				t.Fatal("reuse recovery provisioned a home")
			}
		})
	}
}
