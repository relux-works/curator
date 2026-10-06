// Rework-1 regressions for the Spec §8 version rule at the profile
// lock/store readers: a lock-declared v1 member identity is never
// computed or trusted over NUL bytes, while v2 treats NUL as data.
package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/opaquescan"
)

func writeStoreTestFiles(t *testing.T, root string, files map[string][]byte) {
	t.Helper()
	for rel, payload := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, payload, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func pinHex(t *testing.T, entry string, version hashing.Version) string {
	t.Helper()
	digest, err := hashing.ContentSHA256WithVersion(entry, map[string]bool{}, version)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimPrefix(digest, "sha256:")
}

// The strict member audit refuses a v1 NUL snapshot before hashing its
// revocation identity; the same snapshot under v2 audits clean.
func TestStrictAuditMemberRefusesV1NULBeforeHashing(t *testing.T) {
	home := t.TempDir()
	entry := filepath.Join(t.TempDir(), "entry")
	writeStoreTestFiles(t, entry, map[string][]byte{
		"assets/a.bin": []byte("x\x00y"),
	})
	manager := &gitManager{}
	v1 := contextresolve.Resolved{Kind: contextlock.KindSkill, Name: "skill-a", HashVersion: hashing.VersionV1}
	if _, err := strictAuditMember(home, manager, v1, entry, Policy{}); err == nil ||
		!strings.Contains(err.Error(), opaquescan.FindingNUL) {
		t.Fatalf("v1 strict member audit over NUL = %v, want the opaque refusal", err)
	}
	v2 := contextresolve.Resolved{Kind: contextlock.KindSkill, Name: "skill-a", HashVersion: hashing.VersionV2}
	if _, err := strictAuditMember(home, manager, v2, entry, Policy{}); err != nil {
		t.Fatalf("v2 strict member audit over NUL = %v, want admission", err)
	}
}

// The store pin check refuses a v1 NUL entry before recomputing its
// state pin; a matching v1 pin over clean bytes still verifies, and a
// v2 NUL entry with a matching v2 pin verifies as data.
func TestValidateNamedStorePinsRefusesV1NULBeforeHashing(t *testing.T) {
	home := t.TempDir()

	nulEntry := contextstore.EntryDir(home, contextlock.KindSkill, "nul-skill", "pending")
	writeStoreTestFiles(t, nulEntry, map[string][]byte{
		"assets/a.bin": []byte("x\x00y"),
	})
	v1Lock := &contextlock.Lock{HashVersion: 1, Members: []contextlock.Member{{
		Kind: contextlock.KindSkill, Name: "nul-skill", StateHash: pinHex(t, nulEntry, hashing.VersionV1),
	}}}
	// Rekey the entry under its true pin: the bytes match the recorded
	// v1 pin exactly, so only the pre-hash guard can refuse.
	pinned := contextstore.EntryDir(home, contextlock.KindSkill, "nul-skill", v1Lock.Members[0].StateHash)
	if err := os.MkdirAll(filepath.Dir(pinned), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(nulEntry, pinned); err != nil {
		t.Fatal(err)
	}
	failure := validateNamedStorePins(home, v1Lock)
	if failure == nil || failure.Check != diagPinHash || !strings.Contains(failure.Err.Error(), opaquescan.FindingNUL) {
		t.Fatalf("v1 pin check over NUL = %+v, want a pin_hash failure carrying the opaque refusal", failure)
	}

	cleanEntry := contextstore.EntryDir(home, contextlock.KindSkill, "clean-skill", "pending")
	writeStoreTestFiles(t, cleanEntry, map[string][]byte{
		"assets/a.bin": []byte("x"),
	})
	cleanLock := &contextlock.Lock{HashVersion: 1, Members: []contextlock.Member{{
		Kind: contextlock.KindSkill, Name: "clean-skill", StateHash: pinHex(t, cleanEntry, hashing.VersionV1),
	}}}
	pinnedClean := contextstore.EntryDir(home, contextlock.KindSkill, "clean-skill", cleanLock.Members[0].StateHash)
	if err := os.MkdirAll(filepath.Dir(pinnedClean), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(cleanEntry, pinnedClean); err != nil {
		t.Fatal(err)
	}
	if failure := validateNamedStorePins(home, cleanLock); failure != nil {
		t.Fatalf("v1 pin check over clean matching bytes = %+v, want nil", failure)
	}

	v2Lock := &contextlock.Lock{HashVersion: 2, Members: []contextlock.Member{{
		Kind: contextlock.KindSkill, Name: "nul-v2-skill", StateHash: pinHex(t, pinned, hashing.VersionV2),
	}}}
	v2Entry := contextstore.EntryDir(home, contextlock.KindSkill, "nul-v2-skill", v2Lock.Members[0].StateHash)
	if err := os.MkdirAll(filepath.Dir(v2Entry), 0o755); err != nil {
		t.Fatal(err)
	}
	writeStoreTestFiles(t, v2Entry, map[string][]byte{
		"assets/a.bin": []byte("x\x00y"),
	})
	if failure := validateNamedStorePins(home, v2Lock); failure != nil {
		t.Fatalf("v2 pin check over NUL = %+v, want nil (NUL is ordinary v2 data)", failure)
	}
}

// The managed skill loader refuses a lock-declared v1 member identity
// over NUL bytes; the v2 member loads with its v2 hash.
func TestSkillsOfRefusesV1NULBeforeHashing(t *testing.T) {
	home := t.TempDir()
	manager := &gitManager{}

	nulEntry := contextstore.EntryDir(home, contextlock.KindSkill, "nul-skill", "pending")
	writeStoreTestFiles(t, nulEntry, map[string][]byte{
		"assets/a.bin": []byte("x\x00y"),
	})
	v1Lock := &contextlock.Lock{HashVersion: 1, Members: []contextlock.Member{{
		Kind: contextlock.KindSkill, Name: "nul-skill", StateHash: pinHex(t, nulEntry, hashing.VersionV1),
	}}}
	pinned := contextstore.EntryDir(home, contextlock.KindSkill, "nul-skill", v1Lock.Members[0].StateHash)
	if err := os.MkdirAll(filepath.Dir(pinned), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(nulEntry, pinned); err != nil {
		t.Fatal(err)
	}
	if _, err := skillsOf(home, manager, v1Lock); err == nil ||
		!strings.Contains(err.Error(), DiagRepairFailed) ||
		!strings.Contains(err.Error(), opaquescan.FindingNUL) {
		t.Fatalf("v1 skillsOf over NUL = %v, want a repair failure carrying the opaque refusal", err)
	}

	v2Lock := &contextlock.Lock{HashVersion: 2, Members: []contextlock.Member{{
		Kind: contextlock.KindSkill, Name: "nul-v2-skill", StateHash: pinHex(t, pinned, hashing.VersionV2),
	}}}
	v2Entry := contextstore.EntryDir(home, contextlock.KindSkill, "nul-v2-skill", v2Lock.Members[0].StateHash)
	if err := os.MkdirAll(filepath.Dir(v2Entry), 0o755); err != nil {
		t.Fatal(err)
	}
	writeStoreTestFiles(t, v2Entry, map[string][]byte{
		"assets/a.bin": []byte("x\x00y"),
	})
	skills, err := skillsOf(home, manager, v2Lock)
	if err != nil {
		t.Fatalf("v2 skillsOf over NUL = %v, want loaded skills", err)
	}
	if len(skills) != 1 || skills[0].name != "nul-v2-skill" {
		t.Fatalf("v2 skills = %+v, want the one loaded member", skills)
	}
	wantV2, err := hashing.ContentSHA256WithVersion(v2Entry, map[string]bool{}, hashing.VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	if skills[0].hash != wantV2 {
		t.Fatalf("v2 skill hash = %s, want %s", skills[0].hash, wantV2)
	}
}

// Rework-2 regression for revision-3 finding F3 at the profile
// lock/store readers: each v1 opaque refusal is observed to precede
// hashing. Every clean control hashes, proving the seam is wired.
func TestGuardedReadersV1NULRefusalComputesNoV1Identity(t *testing.T) {
	t.Run("strictAuditMember", func(t *testing.T) {
		home := t.TempDir()
		nulEntry := filepath.Join(t.TempDir(), "entry")
		writeStoreTestFiles(t, nulEntry, map[string][]byte{"assets/a.bin": []byte("x\x00y")})
		cleanEntry := filepath.Join(t.TempDir(), "entry")
		writeStoreTestFiles(t, cleanEntry, map[string][]byte{"assets/a.bin": []byte("x")})
		manager := &gitManager{}
		v1 := contextresolve.Resolved{Kind: contextlock.KindSkill, Name: "skill-a", HashVersion: hashing.VersionV1}
		var auditErr error
		if calls := hashing.CountV1Hashes(func() {
			_, auditErr = strictAuditMember(home, manager, v1, nulEntry, Policy{})
		}); calls != 0 {
			t.Fatalf("v1 strict member audit over NUL computed %d v1 identities, want 0", calls)
		}
		if auditErr == nil || !strings.Contains(auditErr.Error(), opaquescan.FindingNUL) {
			t.Fatalf("v1 strict member audit over NUL = %v, want the opaque refusal", auditErr)
		}
		if calls := hashing.CountV1Hashes(func() {
			if _, err := strictAuditMember(home, manager, v1, cleanEntry, Policy{}); err != nil {
				t.Fatal(err)
			}
		}); calls == 0 {
			t.Fatal("clean v1 strict member audit observed no v1 hash; the seam is not wired")
		}
	})

	rekeyed := func(t *testing.T, home, name string, files map[string][]byte, version hashing.Version) *contextlock.Lock {
		t.Helper()
		staging := contextstore.EntryDir(home, contextlock.KindSkill, name, "pending")
		writeStoreTestFiles(t, staging, files)
		pin := pinHex(t, staging, version)
		lock := &contextlock.Lock{Members: []contextlock.Member{{
			Kind: contextlock.KindSkill, Name: name, StateHash: pin,
		}}}
		lock.HashVersion = int(version)
		pinned := contextstore.EntryDir(home, contextlock.KindSkill, name, pin)
		if err := os.MkdirAll(filepath.Dir(pinned), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(staging, pinned); err != nil {
			t.Fatal(err)
		}
		return lock
	}

	t.Run("validateNamedStorePins", func(t *testing.T) {
		home := t.TempDir()
		nulLock := rekeyed(t, home, "nul-skill",
			map[string][]byte{"assets/a.bin": []byte("x\x00y")}, hashing.VersionV1)
		var failure *storeEntryFailure
		if calls := hashing.CountV1Hashes(func() {
			failure = validateNamedStorePins(home, nulLock)
		}); calls != 0 {
			t.Fatalf("v1 pin check over NUL computed %d v1 identities, want 0", calls)
		}
		if failure == nil || failure.Check != diagPinHash ||
			!strings.Contains(failure.Err.Error(), opaquescan.FindingNUL) {
			t.Fatalf("v1 pin check over NUL = %+v, want a pin_hash opaque failure", failure)
		}
		cleanLock := rekeyed(t, home, "clean-skill",
			map[string][]byte{"assets/a.bin": []byte("x")}, hashing.VersionV1)
		if calls := hashing.CountV1Hashes(func() {
			if failure := validateNamedStorePins(home, cleanLock); failure != nil {
				t.Fatalf("clean v1 pin check = %+v, want nil", failure)
			}
		}); calls == 0 {
			t.Fatal("clean v1 pin check observed no v1 hash; the seam is not wired")
		}
	})

	t.Run("skillsOf", func(t *testing.T) {
		home := t.TempDir()
		manager := &gitManager{}
		nulLock := rekeyed(t, home, "nul-skill",
			map[string][]byte{"assets/a.bin": []byte("x\x00y")}, hashing.VersionV1)
		var skillsErr error
		if calls := hashing.CountV1Hashes(func() {
			_, skillsErr = skillsOf(home, manager, nulLock)
		}); calls != 0 {
			t.Fatalf("v1 skillsOf over NUL computed %d v1 identities, want 0", calls)
		}
		if skillsErr == nil || !strings.Contains(skillsErr.Error(), opaquescan.FindingNUL) {
			t.Fatalf("v1 skillsOf over NUL = %v, want the opaque refusal", skillsErr)
		}
		cleanLock := rekeyed(t, home, "clean-skill",
			map[string][]byte{"assets/a.bin": []byte("x")}, hashing.VersionV1)
		if calls := hashing.CountV1Hashes(func() {
			skills, err := skillsOf(home, manager, cleanLock)
			if err != nil {
				t.Fatal(err)
			}
			if len(skills) != 1 {
				t.Fatalf("clean v1 skills = %+v, want one member", skills)
			}
		}); calls == 0 {
			t.Fatal("clean v1 skillsOf observed no v1 hash; the seam is not wired")
		}
	})
}
