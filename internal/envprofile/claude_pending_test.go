package envprofile

import (
	"bytes"
	"context"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

// Real durable prepared intent, real manager lock, and public Resolve. No
// fabricated recovery inventory. atime/ctime and recovery-directory mtime are
// not asserted; untouched regular entries retain identity/mode/mtime/bytes.
func TestReviewPendingJournal(t *testing.T) {
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		for _, variant := range []string{"early-invalid", "late-invalid", "late-array", "late-tail", "unchanged"} {
			t.Run(form+"/"+variant, func(t *testing.T) {
				invalidPayload := []byte(`{"broken":`)
				if variant == "late-array" {
					invalidPayload = []byte(`[]`)
				}
				if variant == "late-tail" {
					invalidPayload = []byte(`{} {}`)
				}
				fx, req, seed := claudeRepairFixture(t, form)
				contextPath := filepath.Join(filepath.Dir(seed), "CLAUDE.md")
				desired := readClaudeState(t, contextPath)
				if err := os.WriteFile(contextPath, []byte("pending old context"), 0600); err != nil {
					t.Fatal(err)
				}
				// Existing valid seed is satisfied, so successful recovery should leave
				// all other payloads alone and not cause a new repair publication.
				if variant == "early-invalid" {
					if err := os.WriteFile(seed, []byte(`{"broken":`), 0600); err != nil {
						t.Fatal(err)
					}
				}
				baseline := claudeRefusalSnapshot(t, fx.home)
				mgr, err := managerlock.New(fx.home)
				if err != nil {
					t.Fatal(err)
				}
				lock, err := mgr.AcquireHomeOnly(context.Background(), false)
				if err != nil {
					t.Fatal(err)
				}
				engine, err := transaction.New(fx.home)
				if err != nil {
					t.Fatal(err)
				}
				source := filepath.Join(t.TempDir(), "target")
				if err := os.WriteFile(source, desired, 0600); err != nil {
					t.Fatal(err)
				}
				digest, err := transaction.DigestTarget(transaction.KindBytes, contextPath)
				if err != nil {
					t.Fatal(err)
				}
				_, err = engine.Prepare(lock, transaction.Plan{TransactionID: "review-pending", ProjectIdentity: "profiles", Targets: []transaction.Target{{Class: "profile", Identifier: contextPath, Kind: transaction.KindBytes, LivePath: contextPath, StagedSource: source, PreimageDigest: digest}}})
				if err != nil {
					t.Fatal(err)
				}
				pending, err := engine.PendingRecovery(lock)
				if err != nil || !pending {
					t.Fatalf("pending setup: %v %v", pending, err)
				}
				if err := lock.Close(); err != nil {
					t.Fatal(err)
				}
				before := claudeRefusalSnapshot(t, fx.home)
				calls, installs := 0, 0
				var hookErr error
				req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{
					ObserveRecovery: func() { calls++ },
					Observe: func(event transaction.Event) {
						if event.Point == transaction.PointAfterInstall && event.LivePath == contextPath {
							installs++
							if strings.HasPrefix(variant, "late-") {
								hookErr = os.WriteFile(seed, invalidPayload, 0600)
							}
						}
					},
				})}
				result, resolveErr := Resolve(req)
				if hookErr != nil {
					t.Fatal(hookErr)
				}
				after := claudeRefusalSnapshot(t, fx.home)
				if variant == "early-invalid" {
					if resolveErr == nil || !strings.Contains(resolveErr.Error(), envregistry.DiagSeedUnreadable) {
						t.Errorf("wrong early refusal: %v", resolveErr)
					}
					if resolveErr != nil && strings.Contains(resolveErr.Error(), "transaction recovery completed") {
						t.Error("early refusal falsely reported recovery")
					}
					if calls != 0 || installs != 0 {
						t.Errorf("early refusal ran recovery: %d/%d", calls, installs)
					}
					assertClaudeRefusalSnapshot(t, before, after)
					return
				}
				if calls != 1 || installs != 1 {
					t.Fatalf("recovery not real: calls=%d installs=%d", calls, installs)
				}
				if !bytes.Equal(readClaudeState(t, contextPath), desired) {
					t.Error("journal target not recovered exactly")
				}
				// Baseline has no pending sidecars. Exactly its inventory must remain.
				if len(after) != len(baseline) {
					t.Errorf("recovered inventory count %d want %d", len(after), len(baseline))
				}
				seedKey, _ := filepath.Rel(fx.home, seed)
				contextKey, _ := filepath.Rel(fx.home, contextPath)
				for key, old := range baseline {
					now, ok := after[key]
					if !ok {
						t.Errorf("recovery lost %s", key)
						continue
					}
					if key == contextKey {
						continue
					}
					if key == seedKey && strings.HasPrefix(variant, "late-") {
						if !bytes.Equal(now.data, invalidPayload) {
							t.Error("late input overwritten")
						}
						continue
					}
					if old.info.IsDir() {
						continue
					} // recovery legitimately changes ancestor mtimes
					assertClaudeRefusalSnapshot(t, map[string]claudeRefusalEntry{key: old}, map[string]claudeRefusalEntry{key: now})
				}
				for key := range after {
					if _, ok := baseline[key]; !ok {
						t.Errorf("new operation entry: %s", key)
					}
				}
				if variant == "unchanged" {
					if resolveErr != nil {
						t.Fatal(resolveErr)
					}
					return
				}
				if resolveErr == nil || !strings.Contains(resolveErr.Error(), envregistry.DiagSeedUnreadable) {
					t.Fatalf("wrong late refusal: %v", resolveErr)
				}
				t.Logf("real recovery completed, public result=%+v error=%v", result, resolveErr)
				report := strings.ToLower(resolveErr.Error())
				if result != nil {
					report += " " + strings.ToLower(result.Notice+strings.Join(result.Warnings, " "))
				}
				if !strings.Contains(report, "previously pending transaction recovery completed") {
					t.Error("completed recovery is absent from public refusal")
				}
			})
		}
	}
}
