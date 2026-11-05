package envprofile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

// Public Resolve schedules the external write at an existing caller-provided
// native-home lookup, only after CLAUDE.md has actually been repaired. This is
// a deterministic late window, not an atomic-CAS assertion.
func TestReviewFinalPublicationUpdate(t *testing.T) {
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		for _, variant := range []string{"changed", "created", "deleted", "invalid", "array", "tail", "unchanged", "absent", "satisfied"} {
			t.Run(form+"/"+variant, func(t *testing.T) {
				fx, req, seed := claudeRepairFixture(t, form)
				initial := []byte(`{"counter":1}`)
				if variant == "satisfied" {
					initial = readClaudeState(t, seed)
				}
				context := filepath.Join(filepath.Dir(seed), "CLAUDE.md")
				if err := os.WriteFile(seed, initial, 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(context, []byte("review drift"), 0600); err != nil {
					t.Fatal(err)
				}
				if variant == "created" || variant == "absent" {
					if err := os.Remove(seed); err != nil {
						t.Fatal(err)
					}
				}
				seedBefore, _ := os.Stat(seed)
				activeLock := claudeFreshnessLockPath(t, fx.home)
				lockBytes := readClaudeState(t, activeLock)
				var observed map[string]claudeRefusalEntry
				original := req.NativeHomeOf
				calls := 0
				var writeErr error
				req.NativeHomeOf = func(id string) (string, error) {
					current, err := os.ReadFile(context)
					if err == nil && !bytes.Equal(current, []byte("review drift")) && calls == 0 {
						calls++
						switch variant {
						case "changed", "created":
							writeErr = os.WriteFile(seed, []byte(`{"counter":2,"tool_update":"must survive","number":1e999}`), 0600)
						case "deleted":
							writeErr = os.Remove(seed)
						case "invalid":
							writeErr = os.WriteFile(seed, []byte(`{"broken":`), 0600)
						case "array":
							writeErr = os.WriteFile(seed, []byte(`[]`), 0600)
						case "tail":
							writeErr = os.WriteFile(seed, []byte(`{} {}`), 0600)
						}
						if writeErr == nil {
							observed, writeErr = claudeFreshnessSnapshot(fx.home, activeLock)
						}
					}
					return original(id)
				}
				_, err := Resolve(req)
				if writeErr != nil {
					t.Fatal(writeErr)
				}
				if calls != 1 {
					t.Fatalf("late write seam not reached: %d", calls)
				}
				if !bytes.Equal(lockBytes, readClaudeState(t, activeLock)) {
					t.Error("active lock bytes changed")
				}
				if variant == "invalid" || variant == "array" || variant == "tail" {
					if err == nil || !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable) {
						t.Fatalf("wrong late refusal: %v", err)
					}
					after, snapErr := claudeFreshnessSnapshot(fx.home, activeLock)
					if snapErr != nil {
						t.Fatal(snapErr)
					}
					// Staging can change the seed directory mtime; all payloads,
					// inventory, lock metadata and regular-file identities must stay.
					for key, old := range observed {
						if old.info.IsDir() {
							now, ok := after[key]
							if !ok || !os.SameFile(old.info, now.info) || old.info.Mode() != now.info.Mode() {
								t.Errorf("late refusal changed directory identity/mode: %s", key)
								continue
							}
							old.info = now.info
							observed[key] = old
						}
					}
					assertClaudeRefusalSnapshot(t, observed, after)
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				got := claudeNumberObject(t, readClaudeState(t, seed))
				if variant == "changed" || variant == "created" {
					if got["counter"] != json.Number("2") || got["tool_update"] != "must survive" || got["number"] != json.Number("1e999") {
						t.Errorf("late fresh tool update lost: %v", got)
					}
				} else if variant == "deleted" || variant == "absent" {
					if _, exists := got["counter"]; exists {
						t.Error("late deletion resurrected stale field")
					}
				} else if variant == "satisfied" {
					after, statErr := os.Stat(seed)
					if statErr != nil {
						t.Fatal(statErr)
					}
					if !bytes.Equal(initial, readClaudeState(t, seed)) || !os.SameFile(seedBefore, after) || !seedBefore.ModTime().Equal(after.ModTime()) {
						t.Error("satisfied seed changed bytes or metadata")
					}
				} else if got["counter"] != json.Number("1") {
					t.Errorf("unchanged control lost: %v", got)
				}
			})
		}
	}
}
