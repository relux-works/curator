package envprofile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/stateread"
)

type preserveLockVector struct {
	Name  string   `json:"name"`
	Argv  []string `json:"argv"`
	Input struct {
		HomeState       string `json:"home_state"`
		LockFixture     string `json:"lock_fixture"`
		LockReplacement string `json:"lock_replacement_before_repair_lock"`
		Held            bool   `json:"repair_lock_held_by_other_operation"`
	} `json:"input"`
	Expected struct {
		Diagnostic string   `json:"diagnostic"`
		Fragment   bool     `json:"fragment_emitted"`
		Unchanged  bool     `json:"lock_bytes_unchanged"`
		Acquired   bool     `json:"repair_lock_acquired"`
		Writes     []string `json:"writes"`
	} `json:"expected"`
}

// This family is published outside the pinned conformance/v1 corpus. Keep its
// exact 6648e18 vectors locally so hosted CI drives them without moving that pin.
// CURATOR_PRESERVE_LOCK_VECTORS optionally selects the upstream family directly.
func TestRepairPreserveLockVectors(t *testing.T) {
	path := os.Getenv("CURATOR_PRESERVE_LOCK_VECTORS")
	if path == "" {
		path = "testdata/environments-repair-preserve-lock.json"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		Cases []preserveLockVector `json:"cases"`
	}
	if err := json.Unmarshal(raw, &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors.Cases) != 10 {
		t.Fatalf("vector count = %d, want 10", len(vectors.Cases))
	}
	selected := 0
	for _, tc := range vectors.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			selected++
			drivePreserveLockVector(t, tc)
		})
	}
	t.Logf("%d of %d vector subtests selected; production sites: Resolve/StatusOf; CLI grammar: TestEnvResolvePreserveLockRequiresRepair", selected, len(vectors.Cases))
}

func preserveLockFixture(t *testing.T, version, state string) *managedFixture {
	t.Helper()
	if version == "v1" {
		enableV1WritersForTest(t)
	} else {
		enableV2WritersForTest(t)
	}
	fx := writeManagedFixture(t, "example-profile")
	seedLiveNativeCredentials(t, fx)
	if state != "unprovisioned" {
		provision(t, fx, "pi", fx.request("pi").Machine)
		if state == "stale" {
			marker := readManagedMarker(t, fx, "pi")
			path := marker.Surfaces[envmarker.SurfaceRootContext].Paths[0]
			if err := os.Remove(filepath.Join(ManagedHomeDir(fx.home, fx.profile, "pi"), filepath.FromSlash(path))); err != nil {
				t.Fatal(err)
			}
		}
	}
	hashing.EnableV2Writers = true
	return fx
}

// Snapshot the entire temporary home, including directory modes, symlink
// destinations, stores, staging and journals. No lock/journal allowlist.
func preserveSnapshot(t *testing.T, home string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(home, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		value := info.Mode().String()
		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			value += target
		} else if info.Mode().IsRegular() {
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			value += string(raw)
		}
		rel, err := filepath.Rel(home, path)
		if err != nil {
			return err
		}
		result[rel] = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func drivePreserveLockVector(t *testing.T, tc preserveLockVector) {
	t.Helper()
	fx := preserveLockFixture(t, tc.Input.LockFixture, tc.Input.HomeState)
	if len(tc.Argv) > 1 && tc.Argv[1] == "status" {
		hashing.EnableV2Writers = false
		for _, id := range []string{"claude_code", "codex_cli", "opencode"} {
			provision(t, fx, id, fx.request(id).Machine)
		}
		hashing.EnableV2Writers = true
	}
	req := fx.request("pi")
	req.Repair = slices.Contains(tc.Argv, "--repair")
	req.PreserveLock = slices.Contains(tc.Argv, "--preserve-lock")

	if !tc.Expected.Unchanged {
		// Candidate source bytes differ from the pinned snapshot. Migration
		// must retain the snapshot rather than resolve this candidate again.
		source, err := readSource(fx.home, fx.profile)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(source.Path, "context", "a.md"), []byte("unselected candidate source\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	before := preserveSnapshot(t, fx.home)
	oldBytes, err := os.ReadFile(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	oldLock, oldHash, err := contextlock.Read(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	oldSnapshots := map[string]map[string]string{}
	for _, member := range oldLock.Members {
		oldSnapshots[member.Kind+"/"+member.Name] = preserveSnapshot(t, contextstore.EntryDir(fx.home, member.Kind, member.Name, member.PinKey()))
	}
	attempted := false
	req.beforeRepairLock = func() {
		attempted = true
		if tc.Input.LockReplacement != "" {
			lock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
			if err != nil {
				t.Fatal(err)
			}
			lock.Members[0].Weight++ // Same framing and closure; a different lock hash.
			if _, err := contextlock.Write(lockPath(fx.home, fx.profile), lock); err != nil {
				t.Fatal(err)
			}
			before = preserveSnapshot(t, fx.home)
			oldBytes, err = os.ReadFile(lockPath(fx.home, fx.profile))
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	var held *managerlock.HomeLock
	if tc.Input.Held {
		manager, err := managerlock.New(fx.home)
		if err != nil {
			t.Fatal(err)
		}
		held, err = manager.AcquireHomeOnly(t.Context(), false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = held.Close() })
		old := lockTimeout
		lockTimeout = 0
		t.Cleanup(func() { lockTimeout = old })
	}
	if len(tc.Argv) > 1 && tc.Argv[1] == "status" {
		status, err := StatusOf(statusRequest(fx))
		if err != nil {
			t.Fatal(err)
		}
		if status.NonCurrent {
			t.Fatal("current v1 matrix unexpectedly fails --check")
		}
		found := false
		for _, row := range status.Homes {
			if row.Profile == fx.profile && row.Environment == "pi" {
				found = true
				if !row.Current {
					t.Fatalf("v1 currency = %+v", row)
				}
			}
		}
		if !found {
			t.Fatal("missing pi currency row")
		}
	} else {
		result, err := Resolve(req)
		if tc.Expected.Diagnostic != "" {
			if err == nil || !strings.Contains(err.Error(), tc.Expected.Diagnostic) {
				t.Fatalf("diagnostic = %v, want %s", err, tc.Expected.Diagnostic)
			}
			if result != nil && len(result.Document) > 0 {
				t.Fatal("refusal emitted fragment")
			}
			if tc.Expected.Diagnostic == DiagRepairNeedsLockMigration {
				command := "curator env resolve pi --profile example-profile --repair"
				if !strings.HasSuffix(err.Error(), command) {
					t.Fatalf("operator command: %v", err)
				}
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
			if result == nil || len(result.Document) == 0 {
				t.Fatal("missing fragment")
			}
			_, hash, err := contextlock.Read(lockPath(fx.home, fx.profile))
			if err != nil {
				t.Fatal(err)
			}
			var fragment struct {
				Profile struct {
					LockSHA256 string `json:"lock_sha256"`
				} `json:"profile"`
			}
			if err := json.Unmarshal(result.Document, &fragment); err != nil {
				t.Fatal(err)
			}
			if fragment.Profile.LockSHA256 != hashing.Normalize(hash) {
				t.Fatalf("fragment profile.lock_sha256 = %s, want %s", fragment.Profile.LockSHA256, hashing.Normalize(hash))
			}
			if !tc.Expected.Unchanged && hash == oldHash {
				t.Fatal("plain repair skipped identity migration")
			}
		}
	}
	// Windows forbids reading the held lock entry. Keep the complete snapshot
	// comparison on either side of acquisition; omit no files.
	if held != nil {
		if err := held.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if !tc.Expected.Unchanged {
		newLock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
		if err != nil {
			t.Fatal(err)
		}
		if newLock.ContentHashVersion() != hashing.VersionV2 || newLock.Root != oldLock.Root || len(newLock.Members) != len(oldLock.Members) {
			t.Fatal("migration changed closure or failed to select v2")
		}
		for i, member := range newLock.Members {
			old := oldLock.Members[i]
			if !reflect.DeepEqual(oldSnapshots[member.Kind+"/"+member.Name], preserveSnapshot(t, contextstore.EntryDir(fx.home, member.Kind, member.Name, member.PinKey()))) {
				t.Fatal("migration changed snapshot bytes")
			}
			member.StateHash = old.StateHash
			if !reflect.DeepEqual(old, member) {
				t.Fatal("migration re-resolved closure, versions, sources or commits")
			}
		}
	}
	if attempted != tc.Expected.Acquired {
		t.Fatalf("repair lock attempted = %v, want %v", attempted, tc.Expected.Acquired)
	}
	newBytes, err := os.ReadFile(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(oldBytes, newBytes) != tc.Expected.Unchanged {
		t.Fatal("lock-byte invariant violated")
	}
	if len(tc.Expected.Writes) == 0 && !reflect.DeepEqual(before, preserveSnapshot(t, fx.home)) {
		t.Fatal("read-only/refusal path wrote manager state")
	}
}

func TestPreserveLockRefusesBeforeStoreRebuild(t *testing.T) {
	for _, failure := range []string{"pin", "boundary"} {
		t.Run(failure, func(t *testing.T) {
			fx := preserveLockFixture(t, "v1", "current")
			lock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
			if err != nil {
				t.Fatal(err)
			}
			member, _ := lock.Find(contextlock.KindMCP, "figma-devmode")
			entry := contextstore.EntryDir(fx.home, member.Kind, member.Name, member.PinKey())
			if failure == "pin" {
				if err := os.WriteFile(filepath.Join(entry, "unused", "pin-only.txt"), []byte("untrusted\n"), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Chmod(entry, 0777); err != nil {
					t.Fatal(err)
				}
			}
			before := preserveSnapshot(t, fx.home)
			req := fx.request("pi")
			req.Repair = true
			req.PreserveLock = true
			req.beforeRepairLock = func() { t.Fatal("migration refusal attempted repair lock") }
			result, err := Resolve(req)
			if err == nil || !strings.Contains(err.Error(), DiagRepairNeedsLockMigration) || result != nil {
				t.Fatalf("Resolve = %v, %v", result, err)
			}
			if !reflect.DeepEqual(before, preserveSnapshot(t, fx.home)) {
				t.Fatal("refusal wrote/rebuilt store")
			}
		})
	}
}

func TestPreserveLockRaceRefusesBeforeJournalRecovery(t *testing.T) {
	fx := preserveLockFixture(t, "v2", "stale")
	req := fx.request("pi")
	req.Repair = true
	req.PreserveLock = true
	journal := filepath.Join(fx.home, "state", "transactions", "v1", "broken.json")
	if err := os.MkdirAll(filepath.Dir(journal), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(journal, []byte("malformed"), 0600); err != nil {
		t.Fatal(err)
	}
	var before map[string]string
	req.beforeRepairLock = func() {
		lock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
		if err != nil {
			t.Fatal(err)
		}
		lock.Members[0].Weight++
		if _, err := contextlock.Write(lockPath(fx.home, fx.profile), lock); err != nil {
			t.Fatal(err)
		}
		before = preserveSnapshot(t, fx.home)
	}
	result, err := Resolve(req)
	if err == nil || !strings.Contains(err.Error(), DiagRepairFailed+": profile lock changed before repair-lock acquisition") || result != nil {
		t.Fatalf("Resolve = %v, %v", result, err)
	}
	if !reflect.DeepEqual(before, preserveSnapshot(t, fx.home)) {
		t.Fatal("race refusal wrote state")
	}
}

// A busy real lock defeats a refusal moved after beginOperation, even if that
// acquisition bypasses the request's test seam and leaves stable lock bytes.
func TestPreserveLockMigrationRefusalPrecedesBusyLock(t *testing.T) {
	for _, state := range []string{"stale", "unprovisioned", "current"} {
		t.Run(state, func(t *testing.T) {
			fx := preserveLockFixture(t, "v1", state)
			manager, err := managerlock.New(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			// The external holder's lock entry is fixture state, including for an
			// unprovisioned home. Create it before the read-only baseline.
			prepared, err := manager.AcquireHomeOnly(t.Context(), false)
			if err != nil {
				t.Fatal(err)
			}
			if err := prepared.Close(); err != nil {
				t.Fatal(err)
			}
			before := preserveSnapshot(t, fx.home)
			held, err := manager.AcquireHomeOnly(t.Context(), false)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = held.Close() })
			previous := lockTimeout
			lockTimeout = 0
			t.Cleanup(func() { lockTimeout = previous })
			req := fx.request("pi")
			req.Repair = true
			req.PreserveLock = true
			result, err := Resolve(req)
			if err == nil || !strings.Contains(err.Error(), DiagRepairNeedsLockMigration) || result != nil {
				t.Fatalf("migration refusal = %v, %v", result, err)
			}
			if err := held.Close(); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, preserveSnapshot(t, fx.home)) {
				t.Fatal("migration refusal wrote state")
			}
		})
	}
}

// Observe the real held lock during the raced-lock read, rather than treating
// the pre-acquisition injection hook as evidence of acquisition.
func TestPreserveLockRevalidationHoldsRepairLock(t *testing.T) {
	fx := preserveLockFixture(t, "v2", "stale")
	req := fx.request("pi")
	req.Repair, req.PreserveLock = true, true
	reads, observedHeld := 0, false
	req.readRegularFile = func(path string) (stateread.File, error) {
		if path == lockPath(fx.home, fx.profile) {
			reads++
			if reads == 2 {
				manager, err := managerlock.New(fx.home)
				if err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
				defer cancel()
				competing, err := manager.AcquireHomeOnly(ctx, false)
				if err == nil {
					_ = competing.Close()
					t.Fatal("preserved-lock revalidation ran before repair-lock acquisition")
				}
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("competing lock: %v", err)
				}
				observedHeld = true
			}
		}
		return stateread.ReadRegularFile(path)
	}
	var before map[string]string
	req.beforeRepairLock = func() {
		lock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
		if err != nil {
			t.Fatal(err)
		}
		lock.Members[0].Weight++
		if _, err := contextlock.Write(lockPath(fx.home, fx.profile), lock); err != nil {
			t.Fatal(err)
		}
		before = preserveSnapshot(t, fx.home)
	}
	result, err := Resolve(req)
	if err == nil || !strings.Contains(err.Error(), DiagRepairFailed) || result != nil {
		t.Fatalf("Resolve = %v, %v", result, err)
	}
	if !observedHeld {
		t.Fatal("preserved lock was not reread under the repair lock")
	}
	if !reflect.DeepEqual(before, preserveSnapshot(t, fx.home)) {
		t.Fatal("race refusal wrote state")
	}
}

func TestPlainRepairMigrationKeepsLockPrivate(t *testing.T) {
	fx := preserveLockFixture(t, "v1", "current")
	req := fx.request("pi")
	req.Repair = true
	result, err := Resolve(req)
	if err != nil || result == nil || len(result.Document) == 0 {
		t.Fatalf("plain migration: %v, %v", result, err)
	}
	if err := privatedir.ValidateFile(lockPath(fx.home, fx.profile)); err != nil {
		t.Fatalf("migrated lock lost its private boundary: %v", err)
	}
}
