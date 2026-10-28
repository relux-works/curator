//go:build unix

package envprofile

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/transaction"
)

// Actual repair selects copies only for context/skill destination capability
// failures and journals their copy reasons with the unchanged origin. Returned
// faults restore all preimages; fresh recovery needs neither B nor the planning
// alternative. Auth links never acquire a copy alternative. This raw-syscall
// injection is not a real unsupported-filesystem or native-account witness.
func TestProvisioningOriginFreshLinkFallback(t *testing.T) {
	for _, state := range []string{"healthy", "returned-fault", "restart", "credential-link"} {
		t.Run(state, func(t *testing.T) {
			fx, req, origin := recordedOriginSkillsFixture(t, "a", true)
			target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			if err := os.Remove(filepath.Join(target, "AGENTS.md")); err != nil {
				t.Fatal(err)
			}
			before := nativeSnapshot(t, target)
			req.Repair = true
			req.originAccountLookup = func() (envorigin.Account, error) {
				t.Error("recorded repair rediscovered B")
				return envorigin.Account{}, fmt.Errorf("forbidden discovery")
			}
			t.Setenv("CODEX_HOME", t.TempDir())
			calls, faultReached := 0, false
			fault := errors.New("fixture fallback publication fault")
			req.transactionOptions = []transaction.Option{transaction.WithSymlink(func(old, next string) error {
				auth := strings.HasSuffix(old, string(filepath.Separator)+"auth.json")
				if auth == (state == "credential-link") {
					calls++
					return syscall.ENOTSUP
				}
				return os.Symlink(old, next)
			}), transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
				if e.Point == transaction.PointAfterInstall && e.Class == "owned-tree" && state != "healthy" && state != "credential-link" {
					faultReached = true
					if state == "restart" {
						panic("fallback unit stop")
					}
					return fault
				}
				return nil
			}})}
			var result *ResolveResult
			var err error
			func() {
				if state == "restart" {
					defer func() {
						if got := recover(); got != "fallback unit stop" {
							t.Fatalf("wrong stop: %v", got)
						}
					}()
				}
				result, err = Resolve(req)
			}()
			if calls == 0 {
				t.Fatalf("destination capability syscall not reached: %v", err)
			}
			if state == "credential-link" || state == "returned-fault" {
				if state == "credential-link" {
					if !errors.Is(err, syscall.ENOTSUP) || !strings.Contains(err.Error(), "symlink unavailable") {
						t.Fatalf("credential capability refusal missing: %v", err)
					}
				} else if !faultReached || !errors.Is(err, fault) {
					t.Fatalf("post-copy fault not reached: %v", err)
				}
				if result != nil && len(result.Document) > 0 {
					t.Fatal("fallback refusal emitted fragment")
				}
				if nativeSnapshot(t, target) != before {
					t.Fatal("fallback refusal changed managed preimages")
				}
				return
			}
			if state == "restart" {
				if !faultReached {
					t.Fatal("copy stop not reached")
				}
				req.transactionOptions = nil
				result, err = Resolve(req)
			}
			if err != nil || result == nil || len(result.Document) == 0 {
				t.Fatalf("fallback continuation: %v", err)
			}
			raw, err := os.ReadFile(filepath.Join(target, envorigin.Name))
			if err != nil || !bytes.Equal(raw, origin) {
				t.Fatal("fallback changed captured authority")
			}
			marker := readManagedMarker(t, fx, "codex_cli")
			for _, kind := range []string{envmarker.SurfaceRootContext, envmarker.SurfaceSkills, envmarker.SurfaceSystemPrompt, envmarker.SurfaceMCP} {
				surface := marker.Surfaces[kind]
				for _, rel := range surface.Paths {
					info, err := os.Lstat(filepath.Join(target, filepath.FromSlash(rel)))
					if err != nil || info.Mode()&os.ModeSymlink != 0 {
						t.Fatalf("projection was not copied: %s %v", rel, err)
					}
					found := false
					if surface.Copies != nil {
						for _, copy := range *surface.Copies {
							if copy.Path == rel && copy.Reason == envmarker.ReasonSymlinkFallback {
								found = true
							}
						}
					}
					if !found {
						t.Fatalf("copied projection lacks durable reason: %s", rel)
					}
				}
			}
			if text, err := os.Readlink(filepath.Join(target, "auth.json")); err != nil || text != filepath.Join(fx.native["codex_cli"], "auth.json") {
				t.Fatal("auth ceased to be recorded-A link")
			}
			req.Repair = false
			if next, e := Resolve(req); e != nil || next == nil || len(next.Document) == 0 {
				t.Fatalf("copied stable read: %v", e)
			}
		})
	}
}

// The native in-place publisher binds its own marker to the same durable copy
// choice. A later fault restores the native preimages; a fresh engine can finish
// the chosen copies after interruption without looking up today's caller.
func TestProvisioningOriginNativeLinkFallback(t *testing.T) {
	for _, state := range []string{"healthy", "returned-fault", "restart"} {
		t.Run(state, func(t *testing.T) {
			fx, req, _ := recordedOriginSkillsFixture(t, "a", true)
			native := fx.native["codex_cli"]
			// Missing owned entries need repair, but no old symlink needs a
			// lossless backup on the synthetically unsupported destination.
			prior, e := envmarker.Read(native)
			if e != nil || prior == nil {
				t.Fatalf("native fixture marker: %v", e)
			}
			for _, surface := range prior.Surfaces {
				for _, rel := range surface.Paths {
					if e := os.Remove(filepath.Join(native, filepath.FromSlash(rel))); e != nil {
						t.Fatal(e)
					}
				}
			}
			before := nativeSnapshot(t, native)
			calls, hit := 0, false
			fault := errors.New("fixture native fallback fault")
			op, err := beginOperation(fx.home, transaction.WithSymlink(func(old, next string) error {
				calls++
				return syscall.ENOTSUP
			}), transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
				if e.Point == transaction.PointAfterInstall && e.Class == "owned-tree" && state != "healthy" {
					hit = true
					if state == "restart" {
						panic("native fallback stop")
					}
					return fault
				}
				return nil
			}}))
			if err != nil {
				t.Fatal(err)
			}
			var rows []EntryResult
			func() {
				defer op.close()
				if state == "restart" {
					defer func() {
						if got := recover(); got != "native fallback stop" {
							t.Fatalf("wrong stop: %v", got)
						}
					}()
				}
				rows, err = materializeScopeWithNativeHome(op, fx.home, fx.profile, "codex_cli", Policy{}, req.NativeHomeOf)
				if err == nil {
					for _, row := range rows {
						if !row.OK {
							err = row.cause
							if err == nil {
								err = fmt.Errorf("native row refusal: %s", row.Detail)
							}
						}
					}
				}
			}()
			if calls == 0 {
				t.Fatalf("native destination syscall not reached: %+v %v", rows, err)
			}
			if state == "returned-fault" {
				if !hit || !errors.Is(err, fault) {
					t.Fatalf("native fault not reached: %+v %v", rows, err)
				}
				if nativeSnapshot(t, native) != before {
					t.Fatal("native fallback refusal changed preimages")
				}
				return
			}
			if state == "restart" {
				if !hit {
					t.Fatal("native interruption not reached")
				}
				t.Setenv("CODEX_HOME", t.TempDir())
				fresh, e := beginOperation(fx.home)
				if e != nil {
					t.Fatal(e)
				}
				fresh.close()
			} else if err != nil {
				t.Fatal(err)
			}
			marker, err := envmarker.Read(native)
			if err != nil || marker == nil {
				t.Fatalf("native marker absent: %v", err)
			}
			for _, kind := range []string{envmarker.SurfaceRootContext, envmarker.SurfaceSkills} {
				surface := marker.Surfaces[kind]
				if len(surface.Paths) == 0 {
					t.Fatalf("native surface not exercised: %s", kind)
				}
				for _, rel := range surface.Paths {
					info, e := os.Lstat(filepath.Join(native, filepath.FromSlash(rel)))
					if e != nil || info.Mode()&os.ModeSymlink != 0 {
						t.Fatalf("native copy absent: %s %v", rel, e)
					}
					found := false
					if surface.Copies != nil {
						for _, copy := range *surface.Copies {
							if copy.Path == rel && copy.Reason == envmarker.ReasonSymlinkFallback {
								found = true
							}
						}
					}
					if !found {
						t.Fatalf("native copy reason absent: %s", rel)
					}
				}
			}
		})
	}
}
