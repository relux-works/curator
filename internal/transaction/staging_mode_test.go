package transaction

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// The actual staged writer preserves four source modes. Both final-chmod
// fault boundaries restore old bytes/mode on return and after a fresh engine
// recovers a preparing journal. Panic models loss of callbacks, not OS crash.
func TestStagingFinalModeBoundaries(t *testing.T) {
	for _, mode := range []os.FileMode{0400, 0600, 0640, 0644} {
		for _, phase := range []string{"healthy", "before-return", "after-return", "before-restart", "after-restart"} {
			t.Run(fmt.Sprintf("%04o/%s", mode, phase), func(t *testing.T) {
				root := t.TempDir()
				home := filepath.Join(root, "manager")
				mustMkdirAll(t, home)
				target := fileTarget(t, "fixture", "mode", filepath.Join(root, "live"), filepath.Join(root, "source"), "old", "new bytes")
				if err := os.Chmod(target.StagedSource, mode); err != nil {
					t.Fatal(err)
				}
				injected := errors.New("final-mode fault")
				hit := false
				engine := mustEngine(t, home, WithHooks(Hooks{Fault: func(e Event) error {
					point := PointBeforeStagingFinalMode
					if phase == "after-return" || phase == "after-restart" {
						point = PointAfterStagingFinalMode
					}
					if phase != "healthy" && e.Point == point {
						hit = true
						if phase == "before-restart" || phase == "after-restart" {
							panic(injected)
						}
						return injected
					}
					return nil
				}}))
				var journal *Journal
				var err error
				var interrupted any
				func() {
					defer func() { interrupted = recover() }()
					journal, err = engine.Prepare(testLock{}, Plan{TransactionID: "mode-boundary", ProjectIdentity: "fixture", Targets: []Target{target}})
				}()
				if phase == "healthy" {
					if err != nil {
						t.Fatal(err)
					}
					if err = engine.Commit(testLock{}, journal.TransactionID); err != nil {
						t.Fatal(err)
					}
				} else {
					if !hit {
						t.Fatalf("final chmod point missed: %v", err)
					}
					if interrupted != injected && !errors.Is(err, injected) {
						t.Fatalf("wrong failure: %v %v", err, interrupted)
					}
					if err = mustEngine(t, home).Recover(testLock{}); err != nil {
						t.Fatal(err)
					}
				}
				want, wantMode := "old", os.FileMode(0600)
				if phase == "healthy" {
					want, wantMode = "new bytes", mode
				}
				info, err := os.Stat(target.LivePath)
				if err != nil {
					t.Fatal(err)
				}
				if got := mustRead(t, target.LivePath); got != want || info.Mode().Perm() != wantMode {
					t.Fatalf("bytes/mode = %q/%04o, want %q/%04o", got, info.Mode().Perm(), want, wantMode)
				}
				if pending, err := mustEngine(t, home).PendingRecovery(testLock{}); err != nil || pending {
					t.Fatalf("unfinished recovery: %t %v", pending, err)
				}
			})
		}
	}
}
