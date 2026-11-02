package envprofile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/transaction"
)

// Healthy five-link repair reaches the real native owner; a native target fault
// must restore all five native links and the managed home as a single unit.
// Restart uses a fresh engine and another caller selector without recapture.
// This fixture models a process stop with panic; it is not OS crash evidence.
func TestProvisioningOriginNativeWholeUnit(t *testing.T) {
	for _, mode := range []string{"healthy", "returned-error", "restart"} {
		t.Run(mode, func(t *testing.T) {
			fx, req, _ := recordedOriginSkillsFixture(t, "a", true)
			native := fx.native["codex_cli"]
			op, e := beginOperation(fx.home)
			if e != nil {
				t.Fatal(e)
			}
			results, e := materializeScopeWithNativeHome(op, fx.home, fx.profile, "codex_cli", Policy{}, req.NativeHomeOf)
			op.close()
			if e != nil {
				t.Fatal(e)
			}
			for _, result := range results {
				if !result.OK {
					t.Fatalf("initial native owner: %s", result.Detail)
				}
			}
			lock, _, e := readLock(fx.home, fx.profile)
			if e != nil {
				t.Fatal(e)
			}
			skills, e := skillsOf(fx.home, newGitManager(fx.home), lock)
			if e != nil || len(skills) != 5 {
				t.Fatalf("five-link setup: %v", e)
			}
			for _, skill := range skills {
				path := filepath.Join(native, "skills", skill.name)
				if e := os.Remove(path); e != nil {
					t.Fatal(e)
				}
				if e := os.Symlink(skill.source, path); e != nil {
					t.Fatal(e)
				}
			}
			target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			if e := os.Remove(filepath.Join(target, "AGENTS.md")); e != nil {
				t.Fatal(e)
			}
			beforeNative, beforeManaged := nativeSnapshot(t, native), nativeSnapshot(t, target)
			beforeLock, e := os.ReadFile(lockPath(fx.home, fx.profile))
			if e != nil {
				t.Fatal(e)
			}
			reached := false
			fault := errors.New("native-unit-publication-fault")
			req.Repair = true
			if mode != "healthy" {
				req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
					if event.Point == transaction.PointAfterInstall && event.LivePath == filepath.Join(native, "skills", skills[len(skills)-1].name) {
						reached = true
						if mode == "restart" {
							panic("native-unit-stop")
						}
						return fault
					}
					return nil
				}})}
			}
			var result *ResolveResult
			func() {
				if mode == "restart" {
					defer func() {
						if v := recover(); v != "native-unit-stop" {
							t.Fatalf("wrong stop: %v", v)
						}
					}()
				}
				result, e = Resolve(req)
			}()
			if mode == "returned-error" {
				if !reached || !errors.Is(e, fault) || result != nil && len(result.Document) > 0 {
					t.Fatalf("intended publication refusal/fragment: reached=%v error=%v", reached, e)
				}
				if nativeSnapshot(t, native) != beforeNative || nativeSnapshot(t, target) != beforeManaged {
					t.Fatal("whole-unit refusal changed five links or managed preimages")
				}
			} else {
				if mode == "restart" {
					if !reached {
						t.Fatal("native stop not reached")
					}
					req.transactionOptions = nil
					t.Setenv("CODEX_HOME", t.TempDir())
					req.originAccountLookup = func() (envorigin.Account, error) {
						t.Error("recorded restart rediscovered account")
						return envorigin.Account{}, fmt.Errorf("forbidden discovery")
					}
					result, e = Resolve(req)
				}
				if e != nil || result == nil || len(result.Document) == 0 {
					t.Fatalf("healthy whole-unit continuation: %v", e)
				}
				for _, skill := range skills {
					text, e := os.Readlink(filepath.Join(native, "skills", skill.name))
					if e != nil || text != skill.root {
						t.Fatal("whole-unit repair did not converge all five links")
					}
				}
			}
			after, e := os.ReadFile(lockPath(fx.home, fx.profile))
			if e != nil || contextlock.HashBytes(beforeLock) != contextlock.HashBytes(after) {
				t.Fatal("maintenance changed lock")
			}
		})
	}
}
