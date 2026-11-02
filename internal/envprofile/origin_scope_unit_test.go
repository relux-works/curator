package envprofile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/transaction"
)

// Global LP replacement includes a same-profile scoped native adapter; a scope
// selecting another profile is the disjoint neighbor and must stay untouched.
// Both paths reach the actual global publisher, with only synthetic roots.
func TestProvisioningOriginGlobalScopedConsumers(t *testing.T) {
	for _, same := range []bool{true, false} {
		t.Run(map[bool]string{true: "same-profile", false: "other-profile"}[same], func(t *testing.T) {
			fx, resolver := recordedGlobalOriginFixture(t)
			selected := "disjoint-fixture"
			if same {
				selected = fx.profile
			}
			if err := SetScoped(fx.home, "env:codex_cli", selected, false); err != nil {
				t.Fatal(err)
			}
			native := fx.native[envregistry.CodexCLI]
			before := nativeSnapshot(t, native)
			info, err := GlobalInstall(fx.home, Policy{}, envregistry.DefaultMachineConfig(), resolver)
			if err != nil {
				t.Fatal(err)
			}
			if info.LockHash == fx.lockHash {
				t.Fatal("fixture did not reach changed global lock")
			}
			if same {
				marker, err := envmarker.Read(native)
				if err != nil || marker == nil || marker.Profile.LockSHA256 != strings.TrimPrefix(info.LockHash, "sha256:") {
					t.Fatalf("scoped consumer omitted: %+v %v", marker, err)
				}
			} else if nativeSnapshot(t, native) != before {
				t.Fatal("global rebind changed disjoint scoped native home")
			}
		})
	}
}

// Native use and activation share one durable scope unit. A fault after the
// current-pointer write restores every native preimage and the old pointer;
// the healthy neighbor must reach beta on all registered native adapters.
// Returned faults are exercised here; OS crashes are outside this fixture.
func TestProvisioningOriginUseWholeScope(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "healthy", true: "activation-fault"}[fail], func(t *testing.T) {
			enableV1WritersForTest(t)
			home := gitFixtureHome(t)
			pinOperatorHome(t, t.TempDir())
			homes := pinHomes(t)
			for _, name := range []string{"alpha", "beta"} {
				root := t.TempDir()
				writePackage(t, root, name, "1.0.0", name+"\n")
				if _, _, _, err := Install(home, InstallOptions{Operand: root}); err != nil {
					t.Fatal(err)
				}
			}
			before := map[string]string{}
			for id, root := range homes {
				before[string(id)] = nativeSnapshot(t, root)
			}
			pointer, err := os.ReadFile(CurrentFile(home))
			if err != nil {
				t.Fatal(err)
			}
			saved := newOperationEngine
			t.Cleanup(func() { newOperationEngine = saved })
			hit := false
			planted := errors.New("scope-activation-fault")
			newOperationEngine = func(h string, opts ...transaction.Option) (*transaction.Engine, error) {
				return saved(h, append(opts, transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
					if event.Point == transaction.PointAfterInstall && event.LivePath == CurrentFile(home) {
						hit = true
						if fail {
							return planted
						}
					}
					return nil
				}}))...)
			}
			rows, err := UseWithPolicy(home, "beta", "", "", false, Policy{})
			newOperationEngine = saved
			if !hit {
				t.Fatalf("actual activation owner not reached: %+v %v", rows, err)
			}
			if fail {
				if !errors.Is(err, planted) {
					t.Fatalf("wrong refusal: %v", err)
				}
				for id, root := range homes {
					if nativeSnapshot(t, root) != before[string(id)] {
						t.Fatalf("scope refusal changed %s", id)
					}
				}
				after, e := os.ReadFile(CurrentFile(home))
				if e != nil || string(after) != string(pointer) {
					t.Fatalf("scope pointer preimage: %q %v", after, e)
				}
			} else {
				if err != nil || len(rows) != len(Adapters) {
					t.Fatalf("healthy scope continuation: %+v %v", rows, err)
				}
				current, e := Current(home)
				if e != nil || current != "beta" {
					t.Fatalf("current %q %v", current, e)
				}
				for id, root := range homes {
					marker, e := envmarker.Read(root)
					if e != nil || marker == nil || marker.Profile.Name != "beta" {
						t.Fatalf("native %s was not activated: %+v %v", id, marker, e)
					}
				}
				if _, e := os.Lstat(filepath.Join(home, "profiles", "beta", "lock.json")); e != nil {
					t.Fatal(e)
				}
			}
		})
	}
}
