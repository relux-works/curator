package envprofile

import (
	"bytes"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/transaction"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestReviewLegacyRefusal retains the upstream R1 witness through real Resolve.
// Full subtree metadata excludes atime/ctime; valid neighbors must migrate the
// lock and repair context. POSIX unreadability has the established skip class.
func TestReviewLegacyRefusal(t *testing.T) {
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		for _, variant := range []string{"malformed", "array", "null", "scalar", "symlink", "directory", "unreadable", "trailing-object", "trailing-garbage", "valid", "valid-whitespace"} {
			t.Run(form+"/"+variant, func(t *testing.T) {
				if variant == "unreadable" && (runtime.GOOS == "windows" || os.Geteuid() == 0) {
					t.Skip("POSIX mode-bit unreadability")
				}
				prior := hashing.EnableV2Writers
				defer func() { hashing.EnableV2Writers = prior }()
				hashing.EnableV2Writers = false
				fx, req, path := claudeRepairFixture(t, form)
				req.PreserveLock = false
				root := ManagedParent(fx.home, fx.profile, envregistry.ClaudeCode)
				contextPath := filepath.Join(filepath.Dir(path), "CLAUDE.md")
				if err := os.WriteFile(contextPath, []byte("deliberate context drift\n"), 0600); err != nil {
					t.Fatal(err)
				}
				payload := []byte(`{"unknown":1e-7}`)
				diagnostic := envregistry.DiagSeedUnreadable
				switch variant {
				case "trailing-object":
					payload = []byte(`{} {}`)
				case "trailing-garbage":
					payload = []byte(`{}x`)
				case "valid-whitespace":
					payload = []byte("{\"unknown\":1e-7} \t\r\n")
				case "malformed":
					payload = []byte(`{"broken":`)
				case "array":
					payload = []byte(`[]`)
				case "null":
					payload = []byte(`null`)
				case "scalar":
					payload = []byte(`42`)
				}
				if err := os.WriteFile(path, payload, 0600); err != nil {
					t.Fatal(err)
				}
				if variant == "symlink" || variant == "directory" {
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
					if variant == "symlink" {
						target := filepath.Join(t.TempDir(), "state.json")
						if err := os.WriteFile(target, payload, 0600); err != nil {
							t.Fatal(err)
						}
						if err := os.Symlink(target, path); err != nil {
							t.Fatal(err)
						}
						diagnostic = DiagWriteWouldFollowLink
					} else if err := os.Mkdir(path, 0700); err != nil {
						t.Fatal(err)
					}
				}
				if variant == "unreadable" {
					if err := os.Chmod(path, 0); err != nil {
						t.Fatal(err)
					}
					defer func() { _ = os.Chmod(path, 0600) }()
				}
				before := claudeRefusalSnapshot(t, root)
				lockBefore := readClaudeState(t, lockPath(fx.home, fx.profile))
				hashing.EnableV2Writers = true
				recoveries := 0
				req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{ObserveRecovery: func() { recoveries++ }})}
				_, err := Resolve(req)
				if strings.HasPrefix(variant, "valid") {
					if err != nil {
						t.Fatal(err)
					}
					if bytes.Equal(readClaudeState(t, contextPath), []byte("deliberate context drift\n")) {
						t.Fatal("valid control did not repair context")
					}
					req.Repair = false
					req.PreserveLock = false
					if _, err := Resolve(req); err != nil {
						t.Fatalf("verify valid repair: %v", err)
					}
				} else {
					if recoveries != 0 {
						t.Error("seed refusal entered journal recovery")
					}
					if err == nil || !strings.Contains(err.Error(), diagnostic) {
						t.Errorf("Resolve = %v, want %s refusal", err, diagnostic)
					}
					assertClaudeRefusalSnapshot(t, before, claudeRefusalSnapshot(t, root))
					if variant == "unreadable" {
						if err := os.Chmod(path, 0600); err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(readClaudeState(t, path), payload) {
							t.Error("refusal changed unreadable bytes")
						}
					}
				}
				if strings.HasPrefix(variant, "valid") {
					if bytes.Equal(readClaudeState(t, lockPath(fx.home, fx.profile)), lockBefore) {
						t.Error("valid legacy control did not migrate lock")
					}
				} else if !bytes.Equal(readClaudeState(t, lockPath(fx.home, fx.profile)), lockBefore) {
					t.Error("repair changed preserved lock")
				}
			})
		}
	}
}

// A seed changed at the lock-acquisition seam must be validated after that
// seam, before recovery. This proves ordering, not protection from external writers.
func TestClaudeRefusalRevalidatesUnderRepairLock(t *testing.T) {
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		t.Run(form, func(t *testing.T) {
			fx, req, path := claudeRepairFixture(t, form)
			if err := os.WriteFile(filepath.Join(filepath.Dir(path), "CLAUDE.md"), []byte("drift"), 0600); err != nil {
				t.Fatal(err)
			}
			var before map[string]claudeRefusalEntry
			root := ManagedParent(fx.home, fx.profile, envregistry.ClaudeCode)
			req.beforeRepairLock = func() {
				if err := os.WriteFile(path, []byte(`{"broken":`), 0600); err != nil {
					t.Fatal(err)
				}
				before = claudeRefusalSnapshot(t, root)
			}
			recoveries := 0
			req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{ObserveRecovery: func() { recoveries++ }})}
			_, err := Resolve(req)
			if before == nil {
				t.Fatal("lock seam not reached")
			}
			if err == nil || !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable) {
				t.Fatalf("Resolve = %v, want seed refusal", err)
			}
			if recoveries != 0 {
				t.Error("seed refusal entered journal recovery")
			}
			assertClaudeRefusalSnapshot(t, before, claudeRefusalSnapshot(t, root))
		})
	}
}
