package envprofile

import (
	"errors"
	"github.com/relux-works/curator/internal/envorigin"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/transaction"
)

// Actual recorded-origin repair prunes a bounded, metadata-only backup
// generation with symlink children. A returned quarantine fault restores the
// exact old generation and all home surfaces; a foreign top-level entry refuses
// before any publication. These synthetic generations contain no credentials.
func TestProvisioningOriginBackupRetention(t *testing.T) {
	for _, mode := range []string{"healthy", "quarantine-fault", "foreign-entry", "forward-restart", "cleanup-restart", "substituted-quarantine"} {
		t.Run(mode, func(t *testing.T) {
			fx, req, _ := recordedOriginSkillsFixture(t, "a", true)
			native := fx.native["codex_cli"]
			op, err := beginOperation(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			results, err := materializeScopeWithNativeHome(op, fx.home, fx.profile, "codex_cli", Policy{}, req.NativeHomeOf)
			if err := op.close(); err != nil {
				t.Fatal(err)
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, result := range results {
				if !result.OK {
					t.Fatal(result.Detail)
				}
			}
			lock, _, err := readLock(fx.home, fx.profile)
			if err != nil {
				t.Fatal(err)
			}
			skills, err := skillsOf(fx.home, newGitManager(fx.home), lock)
			if err != nil {
				t.Fatal(err)
			}
			for _, skill := range skills {
				path := filepath.Join(native, "skills", skill.name)
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(skill.source, path); err != nil {
					t.Fatal(err)
				}
			}
			base := filepath.Join(native, ".agent-environment-backup")
			for i := 1; i <= BackupRetention; i++ {
				dir := filepath.Join(base, strconv.Itoa(i))
				if err := os.MkdirAll(filepath.Join(dir, "skills"), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(t.TempDir(), "unread-target"), filepath.Join(dir, "skills", "fixture-skill")); err != nil {
					t.Fatal(err)
				}
			}
			oldest := filepath.Join(base, "1")
			if mode == "foreign-entry" {
				if err := os.WriteFile(filepath.Join(oldest, "auth.json"), []byte("synthetic-unowned"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			if err := os.Remove(filepath.Join(target, "AGENTS.md")); err != nil {
				t.Fatal(err)
			}
			beforeNative, beforeManaged := nativeSnapshot(t, native), nativeSnapshot(t, target)
			hit, stopped := false, false
			unitID, quarantine := "", ""
			planted := errors.New("backup-retention-quarantine-fault")
			req.Repair = true
			req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
				if e.Point == transaction.PointAfterQuarantine && e.LivePath == oldest {
					hit = true
					unitID = e.TransactionID
					if mode == "quarantine-fault" {
						return planted
					}
					if mode == "forward-restart" || mode == "substituted-quarantine" {
						matches, er := filepath.Glob(filepath.Join(base, ".curator-quarantine-*"))
						if er != nil || len(matches) != 1 {
							t.Fatalf("quarantine fixture: %v %v", matches, er)
						}
						quarantine = matches[0]
						stopped = true
						panic("backup-unit-stop")
					}
				}
				if mode == "cleanup-restart" && unitID == e.TransactionID && e.Point == transaction.PointDuringQuarantineRemoval {
					stopped = true
					panic("backup-unit-stop")
				}
				return nil
			}})}
			var result *ResolveResult
			func() {
				if strings.HasSuffix(mode, "restart") || mode == "substituted-quarantine" {
					defer func() {
						if v := recover(); v != "backup-unit-stop" {
							t.Fatalf("wrong stop: %v", v)
						}
					}()
				}
				result, err = Resolve(req)
			}()
			if strings.HasSuffix(mode, "restart") || mode == "substituted-quarantine" {
				if !stopped {
					t.Fatal("actual quarantine stop not reached")
				}
				req.transactionOptions = nil
				req.originAccountLookup = func() (envorigin.Account, error) {
					t.Error("restart recaptured account")
					return envorigin.Account{}, errors.New("forbidden account lookup")
				}
				t.Setenv("CODEX_HOME", t.TempDir())
				if mode == "substituted-quarantine" {
					if err := os.Rename(quarantine, quarantine+".saved"); err != nil {
						t.Fatal(err)
					}
					if err := os.Mkdir(quarantine, 0700); err != nil {
						t.Fatal(err)
					}
					beforeNative, beforeManaged = nativeSnapshot(t, native), nativeSnapshot(t, target)
				}
				result, err = Resolve(req)
			}
			if mode == "healthy" || strings.HasSuffix(mode, "restart") {
				if err != nil || result == nil || len(result.Document) == 0 {
					t.Fatalf("actual backup pruning not completed: hit=%t %v", hit, err)
				}
				if _, err := os.Lstat(oldest); !os.IsNotExist(err) {
					t.Fatalf("backup retention left old generation: %v", err)
				}
				if !hit {
					t.Fatal("backup pruning skipped actual quarantine owner")
				}
				entries, err := os.ReadDir(base)
				if err != nil || len(entries) != BackupRetention {
					t.Fatalf("retention count %d: %v", len(entries), err)
				}
			} else {
				if mode == "quarantine-fault" && (!hit || !errors.Is(err, planted)) {
					t.Fatalf("intended fault not reached: %t %v", hit, err)
				}
				if mode == "substituted-quarantine" && (err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid)) {
					t.Fatalf("substituted quarantine not specifically refused: %v", err)
				}
				if mode == "foreign-entry" && (hit || err == nil || !strings.Contains(err.Error(), "backup contains an unowned slot")) {
					t.Fatalf("foreign generation accepted: %t %v", hit, err)
				}
				if result != nil && len(result.Document) > 0 {
					t.Fatal("refusal emitted fragment")
				}
				if nativeSnapshot(t, native) != beforeNative || nativeSnapshot(t, target) != beforeManaged {
					t.Fatal("backup refusal changed preimages")
				}
			}
		})
	}
}
