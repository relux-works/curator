//go:build unix

package transaction

import (
	"errors"
	"github.com/relux-works/curator/internal/staging"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// The actual destination staging syscall selects an explicitly prepared copy
// only for a capability error. Permission failures and ordinary links keep
// their own semantics; fresh recovery needs no planning source or callback.
func TestLinkFallbackPublication(t *testing.T) {
	for _, state := range []string{"linked", "unsupported", "permission", "restart"} {
		t.Run(state, func(t *testing.T) {
			root := t.TempDir()
			live := filepath.Join(root, "live")
			source := filepath.Join(root, "source-link")
			copy := filepath.Join(root, "copy")
			mustWrite(t, live, "old")
			mustWrite(t, copy, "new")
			mustSymlink(t, "canonical", source)
			mustMkdirAll(t, filepath.Join(root, "manager"))
			calls := 0
			engine, err := New(filepath.Join(root, "manager"), WithSymlink(func(old, next string) error {
				calls++
				if state == "permission" {
					return syscall.EACCES
				}
				if state != "linked" {
					return syscall.ENOTSUP
				}
				return os.Symlink(old, next)
			}))
			if err != nil {
				t.Fatal(err)
			}
			target := entryTarget(t, staging.ClassAdapterLedger, "fixture", live, source)
			target.LinkFallback = &LinkFallback{Source: copy, Class: target.Class}
			journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "link-fallback", ProjectIdentity: "fixture", Targets: []Target{target}})
			if calls != 1 {
				t.Fatalf("raw destination link calls = %d, want 1; preparation: %v", calls, err)
			}
			if state == "permission" {
				if !errors.Is(err, syscall.EACCES) {
					t.Fatalf("permission was converted to copy: %v", err)
				}
				if got := mustRead(t, live); got != "old" {
					t.Fatalf("live changed before commit: %q", got)
				}
				journals, readErr := engine.PendingRecovery(testLock{})
				if readErr != nil || journals {
					t.Fatalf("failed preparation left pending state: %v %v", journals, readErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := mustRead(t, live); got != "old" {
				t.Fatalf("live changed before commit: %q", got)
			}
			if state == "restart" {
				if err := os.Remove(source); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(copy); err != nil {
					t.Fatal(err)
				}
				fresh, err := New(filepath.Join(root, "manager"))
				if err != nil {
					t.Fatal(err)
				}
				if err := fresh.Recover(testLock{}); err != nil {
					t.Fatal(err)
				}
			} else if err := engine.Commit(testLock{}, journal.TransactionID); err != nil {
				t.Fatal(err)
			}
			if state == "linked" {
				if got := mustReadlink(t, live); got != "canonical" {
					t.Fatalf("link = %q", got)
				}
			} else {
				if got := mustRead(t, live); got != "new" {
					t.Fatalf("copy = %q", got)
				}
			}
		})
	}
}

// A cleanup conflict, changing alternative or undeclared auth alternative must
// not turn a capability error into a new forward plan. Cleanup may remain
// pending, but its fresh-engine continuation restores the original live entry.
func TestLinkFallbackRefusal(t *testing.T) {
	for _, state := range []string{"cleanup-refusal", "changed-alternative", "undeclared", "credential-entry"} {
		t.Run(state, func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "manager")
			mustMkdirAll(t, home)
			live, source, copy := filepath.Join(root, "live"), filepath.Join(root, "link"), filepath.Join(root, "copy")
			mustWrite(t, live, "old")
			mustWrite(t, copy, "new")
			mustSymlink(t, "canonical", source)
			fault := errors.New("fixture preparing cleanup refusal")
			calls := 0
			engine := mustEngine(t, home, WithSymlink(func(old, next string) error {
				calls++
				if state == "changed-alternative" {
					mustWrite(t, copy, "substituted")
				}
				return syscall.ENOTSUP
			}), WithHooks(Hooks{Fault: func(e Event) error {
				if state == "cleanup-refusal" && e.Point == PointAfterCleanupRename {
					return fault
				}
				return nil
			}}))
			target := entryTarget(t, staging.ClassAdapterLedger, "fixture", live, source)
			target.LinkFallback = &LinkFallback{Source: copy, Class: target.Class}
			if state == "undeclared" {
				target.LinkFallback = nil
			}
			if state == "credential-entry" {
				target.Class = "credential-entry"
			}
			targets := []Target{target}
			if state == "cleanup-refusal" {
				prefixLive, prefixSource := filepath.Join(root, "prefix-live"), filepath.Join(root, "prefix-source")
				mustWrite(t, prefixLive, "prefix-old")
				mustWrite(t, prefixSource, "prefix-new")
				targets = append(targets, fileTarget(t, "00-prefix", "prefix", prefixLive, prefixSource, "prefix-old", "prefix-new"))
			}
			_, err := engine.Prepare(testLock{}, Plan{TransactionID: "refused-fallback", ProjectIdentity: "fixture", Targets: targets})
			switch state {
			case "cleanup-refusal":
				if !errors.Is(err, fault) {
					t.Fatalf("cleanup refusal was retried: %v", err)
				}
			case "changed-alternative":
				if err == nil || !strings.Contains(err.Error(), "prepared copy alternative changed") {
					t.Fatalf("changed copy admitted: %v", err)
				}
			case "undeclared":
				if !errors.Is(err, syscall.ENOTSUP) {
					t.Fatalf("undeclared fallback admitted: %v", err)
				}
			case "credential-entry":
				if calls != 0 || err == nil || !strings.Contains(err.Error(), "outside declared projection") {
					t.Fatalf("credential fallback reached raw syscall: %d %v", calls, err)
				}
			}
			if got := mustRead(t, live); got != "old" {
				t.Fatalf("refusal changed live: %q", got)
			}
			pending, e := engine.PendingRecovery(testLock{})
			if e != nil || pending != (state == "cleanup-refusal") {
				t.Fatalf("unexpected recovery state: %v %v", pending, e)
			}
			if state == "cleanup-refusal" {
				fresh := mustEngine(t, home)
				if e := fresh.Recover(testLock{}); e != nil {
					t.Fatal(e)
				}
				if pending, e := fresh.PendingRecovery(testLock{}); e != nil || pending {
					t.Fatalf("cleanup did not finish: %v %v", pending, e)
				}
				if got := mustRead(t, live); got != "old" {
					t.Fatalf("cleanup forwarded refused copy: %q", got)
				}
			}
		})
	}
}
