//go:build darwin || linux

package projectcontext

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/stateread"
)

// Claim: protected atomic writes create an absent leaf and replace a regular
// leaf, but preserve linked targets and classify a non-directory ancestor as
// unreadable rather than absent. Refusal leaves no staged write behind.
func TestProtectedAtomicWriteAbsence(t *testing.T) {
	root := t.TempDir()
	for _, content := range []string{"first", "second"} {
		if err := writeAtomic(root, "record.json", []byte(content)); err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(root, "record.json"))
		if err != nil || string(b) != content {
			t.Fatalf("regular record = %q, %v", b, err)
		}
	}
	write(t, filepath.Join(root, "blocked"), "ancestor sentinel")
	err := writeAtomic(root, "blocked/child.json", []byte("forbidden"))
	var unreadable *stateread.Error
	if !errors.As(err, &unreadable) || unreadable.Kind != stateread.KindUnreadable {
		t.Fatalf("non-directory ancestor = %v, want unreadable state", err)
	}
	b, err := os.ReadFile(filepath.Join(root, "blocked"))
	if err != nil || string(b) != "ancestor sentinel" {
		t.Fatalf("ancestor changed: %q, %v", b, err)
	}
	if err := os.Symlink("record.json", filepath.Join(root, "linked.json")); err != nil {
		t.Fatal(err)
	}
	required(t, writeAtomic(root, "linked.json", []byte("forbidden")), "project_context_drift")
	target, err := os.Readlink(filepath.Join(root, "linked.json"))
	if err != nil || target != "record.json" {
		t.Fatalf("link changed: %q, %v", target, err)
	}
	b, err = os.ReadFile(filepath.Join(root, "record.json"))
	if err != nil || string(b) != "second" {
		t.Fatalf("linked record changed: %q, %v", b, err)
	}
	staged, err := filepath.Glob(filepath.Join(root, ".write-*"))
	if err != nil || len(staged) != 0 {
		t.Fatalf("staged writes remain: %v, %v", staged, err)
	}
}

// Claim: an absent snapshot is published and an intact existing snapshot is
// reused, while a corrupt existing manifest refuses without replacement or
// an extra staged snapshot. Existing content is never treated as absence.
func TestProtectedSnapshotAbsence(t *testing.T) {
	op := &operation{root: t.TempDir()}
	c := fromFiles(map[string][]byte{"guide.md": []byte("retained")})
	ref, hash, err := op.put(c)
	if err != nil {
		t.Fatal(err)
	}
	nextRef, nextHash, err := op.put(c)
	if err != nil || nextRef != ref || nextHash != hash {
		t.Fatalf("snapshot reuse = %q, %q, %v", nextRef, nextHash, err)
	}
	manifest := filepath.Join(op.root, "snapshots", hash, "manifest.json")
	write(t, manifest, "{")
	_, _, err = op.put(c)
	required(t, err, "project_context_invalid")
	b, err := os.ReadFile(manifest)
	if err != nil || string(b) != "{" {
		t.Fatalf("corrupt manifest replaced: %q, %v", b, err)
	}
	entries, err := os.ReadDir(filepath.Join(op.root, "snapshots"))
	if err != nil || len(entries) != 1 || entries[0].Name() != hash {
		t.Fatalf("snapshot namespace changed: %v, %v", entries, err)
	}
}
