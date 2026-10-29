package envprofile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/transaction"
)

// Compatibility setters have no production CLI caller, but their public API
// still uses the common locked publisher. Healthy writes and literal/encoded
// scope cleanup complete; returned faults restore both record preimages and a
// fresh engine resumes genuine pending intent. This is not an OS crash test.
func TestProvisioningOriginPointerOwner(t *testing.T) {
	for _, scoped := range []bool{false, true} {
		for _, mode := range []string{"healthy", "returned-error", "restart"} {
			name := "current"
			if scoped {
				name = "scoped"
			}
			t.Run(name+"/"+mode, func(t *testing.T) {
				fx, _, _ := recordedOriginFixture(t, "a")
				path := CurrentFile(fx.home)
				legacy := ""
				if scoped {
					path = filepath.Join(ScopedDir(fx.home), scopeFileName("env:codex_cli"))
					legacy = filepath.Join(ScopedDir(fx.home), "env:codex_cli")
					if err := os.MkdirAll(ScopedDir(fx.home), 0700); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(legacy, []byte("old\n"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				prior, priorErr := os.ReadFile(path)
				hit := false
				injected := errors.New("pointer-owner-fault")
				saved := newOperationEngine
				t.Cleanup(func() { newOperationEngine = saved })
				newOperationEngine = func(h string, opts ...transaction.Option) (*transaction.Engine, error) {
					return saved(h, append(opts, transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
						if e.Point == transaction.PointAfterInstall && e.LivePath == path {
							hit = true
							if mode == "returned-error" {
								return injected
							}
							if mode == "restart" {
								panic("pointer-stop")
							}
						}
						return nil
					}}))...)
				}
				var err error
				func() {
					if mode == "restart" {
						defer func() {
							if recover() != "pointer-stop" {
								t.Fatal("wrong restart stop")
							}
						}()
					}
					if scoped {
						err = SetScoped(fx.home, "env:codex_cli", fx.profile, false)
					} else {
						err = SetCurrent(fx.home, fx.profile)
					}
				}()
				newOperationEngine = saved
				if !hit {
					t.Fatalf("actual pointer install not reached: %v", err)
				}
				if mode == "returned-error" {
					if !errors.Is(err, injected) {
						t.Fatalf("wrong fault: %v", err)
					}
					got, e := os.ReadFile(path)
					if string(got) != string(prior) || (e == nil) != (priorErr == nil) {
						t.Fatal("pointer preimage lost")
					}
					if scoped {
						got, e = os.ReadFile(legacy)
						if e != nil || string(got) != "old\n" {
							t.Fatal("legacy preimage lost")
						}
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if mode == "restart" {
					op, e := beginOperation(fx.home)
					if e != nil {
						t.Fatal(e)
					}
					if e = op.close(); e != nil {
						t.Fatal(e)
					}
				}
				got, e := os.ReadFile(path)
				if e != nil || strings.TrimSpace(string(got)) != fx.profile {
					t.Fatalf("healthy continuation: %q %v", got, e)
				}
				if scoped {
					if _, e = os.Lstat(legacy); !os.IsNotExist(e) {
						t.Fatalf("stale spelling resurrects: %v", e)
					}
					if e = SetScoped(fx.home, "env:codex_cli", "", true); e != nil {
						t.Fatal(e)
					}
					if _, e = os.Lstat(path); !os.IsNotExist(e) {
						t.Fatalf("clear continuation: %v", e)
					}
				}
			})
		}
	}
}

// Unsafe scope operands cannot escape the scope record directory on either
// write or clear. The healthy encoded-scope owner above prevents reject-all.
func TestProvisioningOriginPointerScopeBoundary(t *testing.T) {
	fx, _, _ := recordedOriginFixture(t, "a")
	before := nativeSnapshot(t, fx.home)
	for _, scope := range []string{"", ".", "..", "../source.json", "env/codex", "env\\codex"} {
		for _, clear := range []bool{false, true} {
			err := SetScoped(fx.home, scope, fx.profile, clear)
			if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid+": invalid scope key") {
				t.Fatalf("intended scope refusal %q clear=%t: %v", scope, clear, err)
			}
			if nativeSnapshot(t, fx.home) != before {
				t.Fatal("scope refusal changed manager state")
			}
		}
	}
}
