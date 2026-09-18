package scopes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/runtimestore"
)

// TestCollectRetainsDraftSourceV1Runtime proves the mark phase follows the
// draft runtime key: a local-snapshot marker keeps its source-v1 package
// tree while an unreferenced leaf is still swept. Without the v5 branch
// the sweep would mark a bare-commit leaf that does not exist and delete
// the live draft tree.
func TestCollectRetainsDraftSourceV1Runtime(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	skillsDir := filepath.Join(project, ".agents", "skills", "draft-skill")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	snapshot := "sha256:" + strings.Repeat("ab", 32)
	pkg := &marker.Package{Kind: "local-snapshot", Snapshot: snapshot}
	digest, err := pkg.Digest()
	if err != nil {
		t.Fatal(err)
	}
	key, err := runtimestore.SourceV1Key(digest)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := hashing.ContentSHA256(skillsDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := marker.Write(skillsDir, &marker.Marker{
		Name: "draft-skill", Package: pkg,
		LockSHA256:    "sha256:" + strings.Repeat("cd", 32),
		ContentSHA256: hash,
		Agents:        []string{}, Commands: []string{}, Dependencies: []string{},
		SkillSchemaVersion: 4,
		InstalledAt:        "2026-09-18T00:00:00Z",
		Activation:         &marker.Activation{Context: true, Commands: []string{}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}
	live := filepath.Join(home, "runtime", "draft-skill", key)
	stale := filepath.Join(home, "runtime", "draft-skill", strings.Repeat("f", 40))
	for _, dir := range []string{live, stale} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	cache := &recordingCache{}
	result, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}, Cache: cache})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(live); err != nil {
		t.Fatalf("a live draft runtime tree was swept: %v", err)
	}
	if len(result.RemovedRuntime) != 1 || result.RemovedRuntime[0] != "draft-skill/"+strings.Repeat("f", 40) {
		t.Fatalf("removed runtime = %v, want only the unreferenced leaf", result.RemovedRuntime)
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("a clean pass warned: %v", result.Warnings)
	}
}
