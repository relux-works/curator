package envprofile

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

// Real Resolve schedules a tool write after early validation. Successful merges
// use the fresh input (including absence); invalid input preserves the snapshot
// taken after the tool write. This is a deterministic window, not an atomic CAS
// claim for external writers or a rollback claim for pending journal recovery.
func TestClaudeRepairFreshInput(t *testing.T) {
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		for _, variant := range []string{"changed", "created", "deleted", "invalid", "legacy-invalid", "store-invalid", "unchanged", "absent", "satisfied"} {
			t.Run(form+"/"+variant, func(t *testing.T) {
				prior := hashing.EnableV2Writers
				defer func() { hashing.EnableV2Writers = prior }()
				if variant == "legacy-invalid" {
					hashing.EnableV2Writers = false
				}
				fx, req, path := claudeRepairFixture(t, form)
				hashing.EnableV2Writers = prior
				root := fx.home
				activeLock := claudeFreshnessLockPath(t, root)
				lockBytes := readClaudeState(t, activeLock)
				lockKey, err := filepath.Rel(root, activeLock)
				if err != nil {
					t.Fatal(err)
				}
				if variant == "store-invalid" {
					lock, _, err := readLock(fx.home, fx.profile)
					if err != nil {
						t.Fatal(err)
					}
					member, ok := lock.Find(contextlock.KindMCP, "figma-devmode")
					if !ok {
						t.Fatal("fixture missing pinned member")
					}
					entry := contextstore.EntryDir(fx.home, member.Kind, member.Name, member.PinKey())
					if err := os.WriteFile(filepath.Join(entry, "unused", "pin-only.txt"), []byte("corrupt pin"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				initial := []byte(`{"counter":1,"stale_field":"old"}`)
				if variant == "satisfied" {
					initial = readClaudeState(t, path)
				}
				if err := os.WriteFile(path, initial, 0600); err != nil {
					t.Fatal(err)
				}
				if variant == "created" || variant == "absent" {
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
				}
				context := filepath.Join(filepath.Dir(path), "CLAUDE.md")
				if err := os.WriteFile(context, []byte("drift"), 0600); err != nil {
					t.Fatal(err)
				}
				var afterTool map[string]claudeRefusalEntry
				var observationErr error
				calls := 0
				req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{ObserveRecovery: func() {
					calls++
					var err error
					switch variant {
					case "changed", "created":
						err = os.WriteFile(path, []byte(`{"counter":2,"tool_update":"must survive","number":1e999}`), 0600)
					case "deleted":
						err = os.Remove(path)
					case "invalid", "legacy-invalid", "store-invalid":
						err = os.WriteFile(path, []byte(`{"broken":`), 0600)
					}
					if err != nil {
						observationErr = err
						return
					}
					afterTool, observationErr = claudeFreshnessSnapshot(root, activeLock)
				}})}
				_, err = Resolve(req)
				// Never Fatal inside ObserveRecovery: let production release its lock first.
				if observationErr != nil {
					t.Fatalf("recovery observation: %v", observationErr)
				}
				if calls != 1 || afterTool == nil {
					t.Fatalf("recovery plant calls = %d, want 1", calls)
				}
				after, snapshotErr := claudeFreshnessSnapshot(root, activeLock)
				if snapshotErr != nil {
					t.Fatal(snapshotErr)
				}
				assertClaudeRefusalSnapshot(t, map[string]claudeRefusalEntry{lockKey: afterTool[lockKey]}, map[string]claudeRefusalEntry{lockKey: after[lockKey]})
				if afterTool[lockKey].info.Size() != after[lockKey].info.Size() || !bytes.Equal(lockBytes, readClaudeState(t, activeLock)) {
					t.Error("repair changed active lock size or bytes")
				}
				if strings.HasSuffix(variant, "invalid") {
					if err != nil && strings.Contains(err.Error(), "transaction recovery completed") {
						t.Error("empty recovery falsely reported completion")
					}
					if err == nil || !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable) {
						t.Errorf("Resolve = %v, want environment_seed_unreadable refusal", err)
					}
					assertClaudeRefusalSnapshot(t, afterTool, after)
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				got := claudeNumberObject(t, readClaudeState(t, path))
				switch variant {
				case "changed", "created":
					if got["counter"] != json.Number("2") || got["tool_update"] != "must survive" || got["number"] != json.Number("1e999") {
						t.Errorf("fresh tool update lost: %v", got)
					}
					fallthrough
				case "deleted", "absent":
					if _, exists := got["stale_field"]; exists {
						t.Error("resurrected stale user field")
					}
					if (variant == "deleted" || variant == "absent") && (got["counter"] != nil || got["hasCompletedOnboarding"] != true) {
						t.Error("fresh absence did not use the provisioning seed")
					}
				case "unchanged":
					if got["counter"] != json.Number("1") || got["stale_field"] != "old" {
						t.Errorf("unchanged input lost: %v", got)
					}
				case "satisfied":
					key, err := filepath.Rel(root, path)
					if err != nil {
						t.Fatal(err)
					}
					after := claudeRefusalSnapshot(t, root)
					assertClaudeRefusalSnapshot(t, map[string]claudeRefusalEntry{key: afterTool[key]}, map[string]claudeRefusalEntry{key: after[key]})
				}
				if bytes.Equal(readClaudeState(t, context), []byte("drift")) {
					t.Error("valid neighbor did not repair context")
				}
				req.transactionOptions = nil
				req.Repair = false
				beforeNoOp := claudeRefusalSnapshot(t, root)
				if _, err := Resolve(req); err != nil {
					t.Fatalf("verify fresh repair: %v", err)
				}
				assertClaudeRefusalSnapshot(t, beforeNoOp, claudeRefusalSnapshot(t, root))
			})
		}
	}
}

// Match managerlock's manager-home lock name using its canonical home identity.
// No directory-wide exception: project/build/other home lock bytes remain observed.
func claudeFreshnessLockPath(t *testing.T, home string) string {
	t.Helper()
	manager, err := managerlock.New(home)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte("manager-home\x00" + manager.Home()))
	return filepath.Join(home, "state", "locks", "v1", fmt.Sprintf("%x.lock", digest))
}

// Snapshot the full inventory, excluding only the exact active lock's bytes.
// Windows forbids reading its locked region. Presence, regular type, identity,
// mode, size and mtime remain observed; callers compare lock bytes after release.
// Unlike the shared refusal helper, any other unreadable regular file is an error.
func claudeFreshnessSnapshot(root, activeLock string) (map[string]claudeRefusalEntry, error) {
	result := map[string]claudeRefusalEntry{}
	foundLock := false
	err := filepath.WalkDir(root, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		entry := claudeRefusalEntry{info: info}
		switch {
		case path == activeLock:
			if !info.Mode().IsRegular() {
				return fmt.Errorf("active lock is not regular")
			}
			foundLock = true
		case info.Mode()&os.ModeSymlink != 0:
			entry.link, err = os.Readlink(path)
		case info.Mode().IsRegular():
			entry.data, err = os.ReadFile(path)
		}
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		result[rel] = entry
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !foundLock {
		return nil, fmt.Errorf("active lock missing from inventory")
	}
	return result, nil
}

// The observation exception is one exact file, not all locks or state files.
// Missing/nonregular active locks refuse; sibling bytes and new entries survive
// the inventory, so narrowing its byte coverage is detectable on every OS.
func TestClaudeFreshnessSnapshotExactLockOnly(t *testing.T) {
	root := t.TempDir()
	active := filepath.Join(root, "active.lock")
	sibling := filepath.Join(root, "other.lock")
	for _, path := range []string{active, sibling} {
		if err := os.WriteFile(path, []byte("lock bytes"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	snapshot, err := claudeFreshnessSnapshot(root, active)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot) != 3 || snapshot["active.lock"].data != nil || snapshot["active.lock"].info.Size() != 10 || string(snapshot["other.lock"].data) != "lock bytes" {
		t.Fatal("exact-lock observation lost inventory, metadata or sibling bytes")
	}
	if err := os.WriteFile(sibling, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	after, err := claudeFreshnessSnapshot(root, active)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(snapshot["other.lock"].data, after["other.lock"].data) {
		t.Fatal("sibling lock byte change hidden")
	}
	if err := os.Remove(active); err != nil {
		t.Fatal(err)
	}
	if _, err := claudeFreshnessSnapshot(root, active); err == nil {
		t.Fatal("missing active lock accepted")
	}
	if err := os.Mkdir(active, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := claudeFreshnessSnapshot(root, active); err == nil {
		t.Fatal("nonregular active lock accepted")
	}
}
