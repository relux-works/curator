package scopes

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/sourcelock"
)

func TestCollectDryRunAndSweepProtectInstalledSnapshots(t *testing.T) {
	home := protectedTestHome(t)
	store := openProtectedStore(t, home)
	commit := strings.Repeat("a", 40)
	orphanCommit := strings.Repeat("b", 40)
	markerCommit := strings.Repeat("e", 40)
	previousCommit := strings.Repeat("f", 40)
	source := "example.com/team/skill"
	legacyLive := filepath.Join(contextstore.Root(home), "skill", "kept", commit)
	legacyOrphan := filepath.Join(contextstore.Root(home), "skill", "orphan", orphanCommit)
	sourceLive := filepath.Join(home, "cache", filepath.FromSlash(source), commit, "snapshot")
	sourceOrphan := filepath.Join(home, "cache", filepath.FromSlash(source), orphanCommit, "snapshot")
	legacyMarker := filepath.Join(contextstore.Root(home), "skill", "marker-live", markerCommit)
	sourceMarker := filepath.Join(home, "cache", "marker-live", markerCommit, "snapshot")
	legacyPrevious := filepath.Join(contextstore.Root(home), "skill", "previous", previousCommit)
	sourcePrevious := filepath.Join(home, "cache", filepath.FromSlash(source), previousCommit, "snapshot")
	for _, path := range []string{legacyLive, legacyOrphan, sourceLive, sourceOrphan, legacyMarker, sourceMarker, legacyPrevious, sourcePrevious} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "payload"), []byte("fixture\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	installBuildMarker(t, filepath.Join(home, "global", "skills"), "marker-live", markerCommit, "kept-tool", buildKey("2"))
	lockPath := filepath.Join(home, "profiles", "profile", "lock.json")
	if _, err := contextlock.Write(lockPath, &contextlock.Lock{
		SchemaVersion: 2,
		HashVersion:   2,
		Root:          "profile",
		Members: []contextlock.Member{
			{Kind: contextlock.KindContext, Name: "profile", Version: "1.0.0", StateHash: strings.Repeat("c", 64), RequiredBy: []string{}},
			{Kind: contextlock.KindSkill, Name: "kept", Source: source, Commit: commit, RequiredBy: []string{}},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := contextlock.Write(filepath.Join(home, "profiles", "profile", "lock.prev.json"), &contextlock.Lock{
		SchemaVersion: 2,
		HashVersion:   2,
		Root:          "profile",
		Members: []contextlock.Member{
			{Kind: contextlock.KindContext, Name: "profile", Version: "1.0.0", StateHash: strings.Repeat("d", 64), RequiredBy: []string{}},
			{Kind: contextlock.KindSkill, Name: "previous", Source: source, Commit: previousCommit, RequiredBy: []string{}},
		},
	}); err != nil {
		t.Fatal(err)
	}

	preview, err := collectWithTestHomeLock(t, MaintenanceRequest{Home: home, Cache: store, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.WouldRemoveSkillSnapshots) != 1 || preview.WouldRemoveSkillSnapshots[0] != "orphan/"+orphanCommit {
		t.Fatalf("legacy snapshot preview = %v", preview.WouldRemoveSkillSnapshots)
	}
	if len(preview.WouldRemoveSourceSnapshots) != 1 || preview.WouldRemoveSourceSnapshots[0] != source+"/"+orphanCommit {
		t.Fatalf("source snapshot preview = %v", preview.WouldRemoveSourceSnapshots)
	}
	for _, path := range []string{legacyLive, legacyOrphan, sourceLive, sourceOrphan, legacyMarker, sourceMarker, legacyPrevious, sourcePrevious} {
		if _, err := os.Stat(filepath.Join(path, "payload")); err != nil {
			t.Fatalf("dry-run changed %s: %v", path, err)
		}
	}

	result, err := collectWithTestHomeLock(t, MaintenanceRequest{Home: home, Cache: store})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.RemovedSkillSnapshots) != 1 || len(result.RemovedSourceSnapshots) != 1 {
		t.Fatalf("snapshot sweep result = %+v", result)
	}
	for _, path := range []string{legacyLive, sourceLive, legacyMarker, sourceMarker, legacyPrevious, sourcePrevious} {
		if _, err := os.Stat(filepath.Join(path, "payload")); err != nil {
			t.Fatalf("installed snapshot %s was removed: %v", path, err)
		}
	}
	for _, path := range []string{legacyOrphan, sourceOrphan} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("orphan snapshot %s remains or could not be inspected: %v", path, err)
		}
	}
}

func TestCollectKeepsSnapshotsReferencedBySourceLocks(t *testing.T) {
	home := protectedTestHome(t)
	store := openProtectedStore(t, home)
	project := t.TempDir()
	roots := []struct {
		name string
		path string
	}{
		{name: "project", path: project},
		{name: "global", path: filepath.Join(home, "global")},
		{name: "hybrid", path: filepath.Join(home, "hybrid")},
	}
	for index, root := range roots {
		source := "example.com/team/" + root.name + "-skill"
		commit := strings.Repeat(string(rune('a'+index)), 40)
		snapshot := filepath.Join(home, "cache", filepath.FromSlash(source), commit, "snapshot")
		if err := os.MkdirAll(snapshot, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(snapshot, "payload"), []byte("locked\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		pkg, err := sourcelock.NetworkGitPackage(source, sourcelock.Commit{ObjectFormat: "sha1", Hex: commit}, ".")
		if err != nil {
			t.Fatal(err)
		}
		lock, err := sourcelock.New("sha256:"+strings.Repeat("1", 64), []sourcelock.Member{{
			Name: "locked-skill", Directory: ".", Package: pkg, ContentSHA256: "sha256:" + strings.Repeat("2", 64),
		}})
		if err != nil {
			t.Fatal(err)
		}
		if err := sourcelock.Write(sourcelock.PathIn(root.path), lock); err != nil {
			t.Fatal(err)
		}
		if root.name == "project" {
			if err := RecordConsumer(home, root.path); err != nil {
				t.Fatal(err)
			}
		}
	}

	preview, err := collectWithTestHomeLock(t, MaintenanceRequest{Home: home, Cache: store, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.WouldRemoveSourceSnapshots) != 0 {
		t.Fatalf("source-lock snapshots preview as stale: %v", preview.WouldRemoveSourceSnapshots)
	}
	result, err := collectWithTestHomeLock(t, MaintenanceRequest{Home: home, Cache: store})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.RemovedSourceSnapshots) != 0 {
		t.Fatalf("source-lock snapshots removed: %v", result.RemovedSourceSnapshots)
	}
	for index, root := range roots {
		source := "example.com/team/" + root.name + "-skill"
		commit := strings.Repeat(string(rune('a'+index)), 40)
		payload := filepath.Join(home, "cache", filepath.FromSlash(source), commit, "snapshot", "payload")
		if _, err := os.Stat(payload); err != nil {
			t.Fatalf("source lock at %s did not retain %s: %v", root.name, payload, err)
		}
	}
}

func TestSourceSnapshotSweepRefusesOutsideStore(t *testing.T) {
	home := t.TempDir()
	outside := t.TempDir()
	commit := strings.Repeat("d", 40)
	snapshot := filepath.Join(outside, "example.com", "team", "skill", commit, "snapshot")
	if err := os.MkdirAll(snapshot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snapshot, "payload"), []byte("protected\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(home, "cache")); err != nil {
		t.Fatal(err)
	}
	_, _, err := sweepSourceSnapshots(home, map[string]bool{}, map[string]bool{}, maintenanceLock{check: func() error { return nil }}, false)
	if err == nil || !strings.Contains(err.Error(), "refuse source snapshot outside the cache store") {
		t.Fatalf("outside-store sweep error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(snapshot, "payload")); err != nil {
		t.Fatalf("outside snapshot was changed: %v", err)
	}
}

func TestSourceSnapshotSweepRetiresBeforePartialRemoval(t *testing.T) {
	home := t.TempDir()
	source := "example.com/team/skill"
	commit := strings.Repeat("d", 40)
	commitDir := filepath.Join(home, "cache", filepath.FromSlash(source), commit)
	snapshot := filepath.Join(commitDir, "snapshot")
	if err := os.MkdirAll(snapshot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snapshot, "payload"), []byte("immutable\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	injected := errors.New("injected partial removal")
	partialRemove := func(path string) error {
		if !strings.HasPrefix(filepath.Base(path), retiredSourceSnapshotPrefix) {
			t.Fatalf("remove called before retirement: %s", path)
		}
		if err := os.Remove(filepath.Join(path, "payload")); err != nil {
			t.Fatal(err)
		}
		return injected
	}
	_, _, err := sweepSourceSnapshotsWithRemove(
		home, map[string]bool{}, map[string]bool{}, maintenanceLock{check: func() error { return nil }}, false, partialRemove,
	)
	if !errors.Is(err, injected) {
		t.Fatalf("partial cleanup error = %v, want injected removal failure", err)
	}
	if _, err := os.Lstat(snapshot); !os.IsNotExist(err) {
		t.Fatalf("partially removed tree remains at immutable snapshot key: %v", err)
	}
	entries, err := os.ReadDir(commitDir)
	if err != nil {
		t.Fatal(err)
	}
	var retired string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), retiredSourceSnapshotPrefix) {
			retired = filepath.Join(commitDir, entry.Name())
			break
		}
	}
	if retired == "" {
		t.Fatal("failed cleanup did not leave a retired tree for a later sweep")
	}
	if _, err := os.Stat(filepath.Join(retired, "payload")); !os.IsNotExist(err) {
		t.Fatalf("injected partial deletion did not affect the retired tree: %v", err)
	}
	removed, _, err := sweepSourceSnapshotsWithRemove(
		home, map[string]bool{}, map[string]bool{}, maintenanceLock{check: func() error { return nil }}, false, os.RemoveAll,
	)
	if err != nil || len(removed) != 1 || !strings.HasPrefix(removed[0], "retired:") {
		t.Fatalf("retry did not clean the retired tree: removed=%v err=%v", removed, err)
	}
	if _, err := os.Lstat(retired); !os.IsNotExist(err) {
		t.Fatalf("retired tree remains after successful retry: %v", err)
	}
}

func TestLegacySkillSnapshotSweepRefusesOutsideStore(t *testing.T) {
	home := t.TempDir()
	outside := t.TempDir()
	commit := strings.Repeat("c", 40)
	snapshot := filepath.Join(outside, "orphan", commit)
	if err := os.MkdirAll(snapshot, 0o755); err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(snapshot, "payload")
	if err := os.WriteFile(payload, []byte("protected\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(contextstore.Root(home), "skill")
	if err := os.MkdirAll(filepath.Dir(root), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, root); err != nil {
		t.Fatal(err)
	}
	_, _, err := sweepLegacySkillSnapshots(home, map[string]bool{}, maintenanceLock{check: func() error { return nil }}, false)
	if err == nil || !strings.Contains(err.Error(), "refuse snapshot outside the skill store") {
		t.Fatalf("outside-store sweep error = %v", err)
	}
	if _, err := os.Stat(payload); err != nil {
		t.Fatalf("outside snapshot was changed: %v", err)
	}
}
