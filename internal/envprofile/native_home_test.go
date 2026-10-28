package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envorigin"
)

// Native context discovery retains its separate selector contract. Its actual
// owner refuses physical managed aliases before a native write, rather than
// borrowing a managed origin or substituting a guessed home. No OS account or
// provider witness is claimed; all paths and HOME are owned synthetic fixtures.
func TestProvisioningOriginNativeCustodyBoundary(t *testing.T) {
	root := t.TempDir()
	manager := filepath.Join(root, "manager")
	if e := os.Mkdir(manager, 0700); e != nil {
		t.Fatal(e)
	}
	t.Setenv("HOME", filepath.Join(root, "synthetic-operator"))
	for _, adapter := range Adapters {
		t.Run(adapter.ID, func(t *testing.T) {
			native := filepath.Join(root, adapter.ID)
			if e := os.Mkdir(native, 0700); e != nil {
				t.Fatal(e)
			}
			t.Setenv(adapter.EnvVar, native)
			want := native
			if adapter.ID == "opencode" {
				want = filepath.Join(native, "opencode")
			}
			got, e := resolveNativeHome(manager, adapter, nil)
			if e != nil || got != want {
				t.Fatalf("external positive: got=%q want=%q error=%v", got, want, e)
			}
			alias := filepath.Join(root, "alias-"+adapter.ID)
			if e := os.Symlink(manager, alias); e != nil {
				t.Fatal(e)
			}
			selector := filepath.Join(alias, "environments", "p", adapter.ID)
			t.Setenv(adapter.EnvVar, selector)
			_, e = resolveNativeHome(manager, adapter, nil)
			if e == nil || !strings.Contains(e.Error(), envorigin.DiagInvalid) {
				t.Fatalf("native owner admitted managed alias: %v", e)
			}
			if current, ok := os.LookupEnv(adapter.EnvVar); !ok || current != selector {
				t.Fatal("native guard changed caller selector")
			}
			if _, e := os.Lstat(filepath.Join(manager, "environments")); !os.IsNotExist(e) {
				t.Fatal("native guard materialized managed root")
			}
		})
	}
}
