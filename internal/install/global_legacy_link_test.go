package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/globalbins"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/transaction"
)

// TestGlobalReconcilesOwnedLegacyForwardingLink drives install.Global against
// every user-bin shim state: an owned regular launcher and an owned legacy
// symlink must both reconcile, while foreign bytes — present at planning or
// swapped in after preparation — must be preserved and never adopted.
func TestGlobalReconcilesOwnedLegacyForwardingLink(t *testing.T) {
	for _, shape := range []string{"owned-regular", "foreign-regular", "owned-legacy-link", "foreign-after-prepare"} {
		t.Run(shape, func(t *testing.T) {
			e := newEnv(t)
			e.skill("legacy-skill")
			e.globalDeclareAll("legacy-skill")
			user := t.TempDir()
			bin := filepath.Join(user, "bin")
			if err := os.Mkdir(bin, 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv(globalbins.UserBinEnv, bin)
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			opts := Options{Platform: installPlatform()}
			first := Global(e.cfg, user, opts)
			if first.Status != "ok" {
				t.Fatalf("initial global install: %+v", first)
			}
			name := "legacy-skill-tool"
			published := filepath.Join(bin, shimName(name))
			switch shape {
			case "foreign-regular":
				if err := os.WriteFile(published, []byte("foreign bytes\n"), 0o700); err != nil {
					t.Fatal(err)
				}
			case "owned-legacy-link":
				if installPlatform() == "windows" {
					t.Skip("Unix legacy-link reconciliation is exercised on Linux and macOS")
				}
				messages := globalbins.Refresh(e.home, map[string]bool{name: true}, installPlatform(), nil, user)
				info, err := os.Lstat(published)
				if err != nil || info.Mode()&os.ModeSymlink == 0 {
					t.Fatalf("Refresh did not make legacy link: %v %v", err, messages)
				}
			case "foreign-after-prepare":
				opts.Commit.Hooks = transaction.Hooks{Fault: func(event transaction.Event) error {
					if event.Point == transaction.PointPrepared {
						return os.WriteFile(published, []byte("foreign bytes\n"), 0o700)
					}
					return nil
				}}
			}
			result := Global(e.cfg, user, opts)
			t.Logf("Global status=%s errors=%v", result.Status, result.Errors)
			if strings.HasPrefix(shape, "foreign") {
				got, err := os.ReadFile(published)
				if err != nil || string(got) != "foreign bytes\n" {
					t.Fatalf("foreign shim changed: %q %v", got, err)
				}
				if shape == "foreign-after-prepare" && result.Status == "ok" {
					t.Fatal("preimage drift must fail")
				}
			} else if result.Status != "ok" {
				t.Fatalf("manager-owned forwarding shim cannot be reconciled: %+v", result)
			}
			if shape == "owned-legacy-link" && result.Status == "ok" {
				info, err := os.Lstat(published)
				if err != nil {
					t.Fatal(err)
				}
				if info.Mode()&os.ModeSymlink != 0 {
					t.Fatal("reconciled forwarding shim is still a link")
				}
				canonical := filepath.Join(GlobalRoot(e.home), "bin", shimName(name))
				got, err := os.ReadFile(published)
				if err != nil || string(got) != runtimestore.UnixShimContent(canonical, nil) {
					t.Fatalf("reconciled shim bytes = %q, %v; want the current forwarding launcher", got, err)
				}
			}
		})
	}
}
