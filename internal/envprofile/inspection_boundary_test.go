package envprofile

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/skillintegrity"
)

// A newly extracted inspection snapshot has the same private boundary as a
// published snapshot, yields valid evidence and leaves manager state unchanged.
func TestInspectionSnapshotPrivateBoundary(t *testing.T) {
	home, candidate := verificationCandidate(t)
	member, ok := candidate.Lock.Find(contextlock.KindSkill, "verified-skill")
	if !ok {
		t.Fatal("missing fixture skill")
	}
	manager := newGitManager(home)
	entry := manager.entryPath(home, resolvedOf(member))
	if err := os.RemoveAll(entry); err != nil {
		t.Fatal(err)
	}
	before := hashTreeForTest(t, home)
	snapshot, cleanup, err := manager.inspectEntry(home, resolvedOf(member))
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if snapshot == entry {
		t.Fatal("inspection published the absent store entry")
	}
	if err := privatedir.Validate(snapshot); err != nil {
		t.Fatalf("inspection root is not private: %v", err)
	}
	if err := pathboundary.Validate(snapshot); err != nil {
		t.Fatalf("inspection tree boundary: %v", err)
	}
	if _, err := skillintegrity.Build(context.Background(), snapshot, contextlock.SkillIdentity(member), nil); err != nil {
		t.Fatalf("private inspection evidence: %v", err)
	}
	cleanup()
	if _, err := os.Lstat(snapshot); !os.IsNotExist(err) {
		t.Fatalf("inspection cleanup: %v", err)
	}
	if after := hashTreeForTest(t, home); after != before {
		t.Fatal("inspection changed manager state")
	}
}

// An existing snapshot with a foreign mutation grant remains refused after
// inspection. Inspection must never normalize its permissions or change bytes.
// Restoring that exact grant fixture makes the nearby valid control pass.
func TestInspectionPreservesExistingBoundaryRefusal(t *testing.T) {
	home, candidate := verificationCandidate(t)
	member, _ := candidate.Lock.Find(contextlock.KindSkill, "verified-skill")
	manager := newGitManager(home)
	entry := manager.entryPath(home, resolvedOf(member))
	before, err := os.ReadFile(filepath.Join(entry, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	restore, err := makeWorldWritableSingleDirectoryForTest(entry)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := restore(); err != nil {
			t.Error(err)
		}
	})
	inspected, cleanup, err := manager.inspectEntry(home, resolvedOf(member))
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if inspected != entry {
		t.Fatal("inspection replaced the existing entry")
	}
	if _, err := skillintegrity.Build(context.Background(), inspected, contextlock.SkillIdentity(member), nil); !skillintegrity.IsCode(err, "skill_integrity_boundary_invalid") {
		t.Fatalf("foreign mutation grant was admitted or repaired: %v", err)
	}
	after, err := os.ReadFile(filepath.Join(entry, "SKILL.md"))
	if err != nil || string(after) != string(before) {
		t.Fatal("inspection changed existing payload bytes")
	}
	if err := restore(); err != nil {
		t.Fatal(err)
	}
	if _, err := skillintegrity.Build(context.Background(), entry, contextlock.SkillIdentity(member), nil); err != nil {
		t.Fatalf("restored valid snapshot refused: %v", err)
	}
}
