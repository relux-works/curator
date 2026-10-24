package envprofile

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/privatedir"
)

// Claim: pinned snapshots consumed by composition use a private staging root
// and private files, including nested paths, before a caller receives them.
// A missing pin refuses without leaving an owned staging tree behind. This
// lower-level contract runs on Windows without claiming native CIP support.
func TestPinnedCompositionSnapshotPrivacy(t *testing.T) {
	home := t.TempDir()
	source := "github.com/example/composition"
	commit := writeGitStoreFixture(t, home, "skill", "fixture", source, map[string]string{"SKILL.md": "inert fixture", "nested/data.txt": "retained"})
	parent := t.TempDir()
	for _, name := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(name, parent)
	}
	member := contextlock.Member{Kind: "skill", Name: "fixture", Source: source, Commit: commit}
	snapshot, cleanup, _, err := extractPinnedSnapshot(home, member)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if err = filepath.WalkDir(snapshot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return privatedir.Validate(path)
		}
		return privatedir.ValidateFile(path)
	}); err != nil {
		t.Fatalf("pinned composition snapshot is not private: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(snapshot, "nested", "data.txt"))
	if err != nil || string(raw) != "retained" {
		t.Fatalf("private capture changed bytes: %q %v", raw, err)
	}
	cleanup()
	member.Commit = strings.Repeat("f", 40)
	_, _, _, err = extractPinnedSnapshot(home, member)
	if err == nil || !strings.Contains(err.Error(), "extract pinned snapshot") {
		t.Fatalf("missing pin got wrong refusal: %v", err)
	}
	entries, err := os.ReadDir(parent)
	if err != nil || len(entries) != 0 {
		t.Fatalf("snapshot refusal leaked staging: %v %v", entries, err)
	}
}
