package envprofile

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/transaction"
)

// Actual Sync plans machine and scoped profiles into one publication. A fault
// in the scoped destination restores every adapter's exact preimages; the
// healthy neighbor updates all adapters to their independently selected profile.
func TestProvisioningOriginSyncWholeScope(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "healthy", true: "scoped-fault"}[fail], func(t *testing.T) {
			enableV1WritersForTest(t)
			home := gitFixtureHome(t)
			pinOperatorHome(t, t.TempDir())
			homes := pinHomes(t)
			scopeFixture(t, home)
			before := map[string]string{}
			for id, root := range homes {
				before[id] = nativeSnapshot(t, root)
			}
			saved := newOperationEngine
			t.Cleanup(func() { newOperationEngine = saved })
			hit := false
			fault := errors.New("fixture scoped sync publication fault")
			newOperationEngine = func(h string, opts ...transaction.Option) (*transaction.Engine, error) {
				return saved(h, append(opts, transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
					if e.Point == transaction.PointAfterInstall && e.LivePath == filepath.Join(homes["codex_cli"], "AGENTS.md") {
						hit = true
						if fail {
							return fault
						}
					}
					return nil
				}}))...)
			}
			rows, err := SyncWithPolicy(home, Policy{})
			newOperationEngine = saved
			if !hit {
				t.Fatalf("actual scoped sync owner not reached: %+v %v", rows, err)
			}
			if fail {
				if !errors.Is(err, fault) {
					t.Fatalf("wrong sync refusal: %v", err)
				}
				for id, root := range homes {
					if nativeSnapshot(t, root) != before[id] {
						t.Fatalf("scoped refusal changed %s preimages", id)
					}
				}
				return
			}
			if err != nil || len(rows) != len(Adapters) {
				t.Fatalf("healthy scoped sync: %+v %v", rows, err)
			}
			for id, root := range homes {
				want := "alpha"
				if id == "codex_cli" || id == "opencode" {
					want = "beta"
				}
				marker, e := envmarker.Read(root)
				if e != nil || marker == nil || marker.Profile.Name != want {
					t.Fatalf("sync selected wrong profile for %s: %+v %v", id, marker, e)
				}
			}
		})
	}
}
