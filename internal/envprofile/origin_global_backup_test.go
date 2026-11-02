package envprofile

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/transaction"
)

// The actual global group owns backup pruning too. Healthy L1 prunes; an
// ordinary later marker failure restores L0 and all generations; the narrowly
// typed conflict retains L1 while leaving generations/native/managed preimages
// untouched through journal cleanup and an ordinary restart admission.
func TestProvisioningOriginGlobalBackupRetention(t *testing.T) {
	for _, mode := range []string{"healthy", "later-fault", "retained"} {
		t.Run(mode, func(t *testing.T) {
			fx, resolver := recordedGlobalOriginFixture(t)
			op, e := beginOperation(fx.home)
			if e != nil {
				t.Fatal(e)
			}
			rows, e := materializeScopeWithNativeHome(op, fx.home, fx.profile, "codex_cli", Policy{}, resolver)
			op.close()
			if e != nil {
				t.Fatal(e)
			}
			for _, row := range rows {
				if !row.OK {
					t.Fatal(row.Detail)
				}
			}
			native := fx.native["codex_cli"]
			base := filepath.Join(native, ".agent-environment-backup")
			for i := 1; i <= BackupRetention; i++ {
				dir := filepath.Join(base, strconv.Itoa(i))
				if err := os.MkdirAll(dir, 0700); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(dir, "AGENTS.md")
				if _, err := os.Lstat(path); os.IsNotExist(err) {
					if err = os.Symlink("unread-synthetic-target", path); err != nil {
						t.Fatal(err)
					}
				} else if err != nil {
					t.Fatal(err)
				}
			}
			if mode == "retained" {
				plantGlobalUnmanaged(t, fx)
			}
			beforeNative := nativeSnapshot(t, native)
			managed := filepath.Join(EnvRoot(fx.home), fx.profile)
			beforeManaged := nativeSnapshot(t, managed)
			_, old, err := readLock(fx.home, fx.profile)
			if err != nil {
				t.Fatal(err)
			}
			hit := false
			planted := errors.New("global-backup-later-fault")
			saved := newOperationEngine
			if mode == "later-fault" {
				newOperationEngine = func(home string, options ...transaction.Option) (*transaction.Engine, error) {
					return saved(home, append(options, transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
						if e.Point == transaction.PointAfterInstall && e.Class == "marker" {
							hit = true
							if _, err := os.Lstat(filepath.Join(base, "1")); !os.IsNotExist(err) {
								t.Fatalf("later fault preceded actual backup quarantine: %v", err)
							}
							return planted
						}
						return nil
					}}))...)
				}
				t.Cleanup(func() { newOperationEngine = saved })
			}
			info, err := GlobalInstall(fx.home, Policy{}, envregistry.DefaultMachineConfig(), resolver)
			newOperationEngine = saved
			if mode == "retained" {
				var retained *transaction.RetainedGlobalConflict
				if !errors.As(err, &retained) || retained.Cleanup != nil {
					t.Fatalf("handled conflict not retained: %v", err)
				}
			} else if mode == "later-fault" {
				if !hit || !errors.Is(err, planted) {
					t.Fatalf("later fault not reached: %t %v", hit, err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			_, next, err := readLock(fx.home, fx.profile)
			if err != nil {
				t.Fatal(err)
			}
			if mode == "later-fault" {
				if next != old {
					t.Fatal("ordinary failure retained L1")
				}
			} else if next == old || info.LockHash != next {
				t.Fatal("healthy/handled lock disposition differs")
			}
			if err := envauthority.Validate(fx.home); err != nil {
				t.Fatal(err)
			}
			if mode != "healthy" {
				if nativeSnapshot(t, native) != beforeNative || nativeSnapshot(t, managed) != beforeManaged {
					t.Fatal("failed/retained global mutated backup/home preimages")
				}
			} else {
				if _, err := os.Lstat(filepath.Join(base, "1")); !os.IsNotExist(err) {
					t.Fatalf("global group failed to prune: %v", err)
				}
				entries, err := os.ReadDir(base)
				if err != nil || len(entries) != BackupRetention {
					t.Fatalf("global backup retention count %d: %v", len(entries), err)
				}
			}
		})
	}
}
