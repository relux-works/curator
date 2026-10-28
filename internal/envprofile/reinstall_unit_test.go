package envprofile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/transaction"
)

// Actual same-source Install owns changed-pin rebind and unchanged-pin --use
// activation with existing scopes. A fault at the later Codex native entry
// restores profile records, homes, native outputs and selection to the invocation
// baseline. Immutable store additions are deliberately outside this inventory.
func TestReinstallWholeInvocationUnit(t *testing.T) {
	for _, state := range []string{"changed-success", "changed-fault", "unchanged-success", "unchanged-fault"} {
		t.Run(state, func(t *testing.T) {
			home := gitFixtureHome(t)
			pinOperatorHome(t, t.TempDir())
			homes := pinHomes(t)
			source := t.TempDir()
			writePackage(t, source, "reinstall-unit", "1.0.0", "first\n")
			if _, _, _, err := Install(home, InstallOptions{Operand: source}); err != nil {
				t.Fatal(err)
			}
			fx := &managedFixture{home: home, profile: "reinstall-unit", native: homes, xdg: filepath.Dir(homes["opencode"]), launch: t.TempDir(), accountAnchor: t.TempDir()}
			for _, id := range []string{"claude_code", "codex_cli"} {
				req := fx.request(id)
				req.Repair = true
				if _, err := Resolve(req); err != nil {
					t.Fatal(err)
				}
			}
			if err := SetScoped(home, "env:codex_cli", fx.profile, false); err != nil {
				t.Fatal(err)
			}
			unchanged := strings.HasPrefix(state, "unchanged")
			if unchanged {
				if err := SetCurrent(home, DefaultProfile); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.WriteFile(filepath.Join(source, "context", "a.md"), []byte("second\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			snapshot := func() string {
				var b strings.Builder
				for _, path := range []string{ProfileDir(home, fx.profile), ManagedHomeDir(home, fx.profile, "claude_code"), ManagedHomeDir(home, fx.profile, "codex_cli"), filepath.Dir(CurrentFile(home)), homes["claude_code"], homes["codex_cli"], homes["pi"], homes["opencode"]} {
					b.WriteString(nativeSnapshot(t, path))
				}
				return b.String()
			}
			before := snapshot()
			hit := false
			saved := newOperationEngine
			if strings.HasSuffix(state, "fault") {
				newOperationEngine = func(h string, opts ...transaction.Option) (*transaction.Engine, error) {
					return saved(h, append(opts, transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
						if e.Point == transaction.PointAfterInstall && e.LivePath == filepath.Join(homes["codex_cli"], "AGENTS.md") {
							hit = true
							return errors.New("reinstall scoped publication fault")
						}
						return nil
					}}))...)
				}
			}
			_, activated, _, err := Install(home, InstallOptions{Operand: source, Use: unchanged})
			newOperationEngine = saved
			if strings.HasSuffix(state, "fault") {
				if !hit || err == nil || !strings.Contains(err.Error(), "reinstall scoped publication fault") {
					t.Fatalf("actual reinstall continuation not reached: hit=%t %v", hit, err)
				}
				if snapshot() != before {
					t.Fatal("reinstall refusal changed invocation preimages")
				}
				return
			}
			if err != nil || activated != unchanged {
				t.Fatalf("reinstall success: activated=%t err=%v", activated, err)
			}
			if current, err := Current(home); err != nil || current != fx.profile {
				t.Fatalf("reinstall selection: %s %v", current, err)
			}
			for _, id := range []string{"claude_code", "codex_cli"} {
				req := fx.request(id)
				if _, err := Resolve(req); err != nil {
					t.Fatalf("reinstall left stale managed home: %s %v", id, err)
				}
			}
		})
	}
}
