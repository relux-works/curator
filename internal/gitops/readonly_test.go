package gitops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Cached exact objects remain usable without temporary config writes; missing
// promisor objects refuse without executing the configured acquisition helper.
func TestPinnedReadOnlyNoAcquisition(t *testing.T) {
	repo, commit := directoryRepo(t, nil)
	scratch := t.TempDir()
	t.Setenv("TMPDIR", scratch)
	marker := filepath.Join(scratch, "unexpected-acquisition")
	gitRun(t, repo, "config", "remote.origin.url", "ext::touch "+filepath.ToSlash(marker))
	gitRun(t, repo, "config", "remote.origin.promisor", "true")
	gitRun(t, repo, "config", "extensions.partialClone", "origin")
	gitRun(t, repo, "config", "protocol.ext.allow", "always")
	// Ambient acquisition/config overrides must not replace the local read policy.
	t.Setenv("GIT_ALLOW_PROTOCOL", "ext")
	t.Setenv("GIT_NO_LAZY_FETCH", "0")
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "core.hooksPath")
	t.Setenv("GIT_CONFIG_VALUE_0", scratch)
	before, err := os.Stat(scratch)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := PinnedDirectorySnapshotTree(repo, commit, "agents/skills/demo"); err != nil {
		t.Fatalf("existing local pin: %v", err)
	}
	if _, _, err := PinnedCommitTree(repo, strings.Repeat("1", 40)); err == nil {
		t.Fatal("missing exact object admitted")
	}
	entries, err := os.ReadDir(scratch)
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(scratch)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(before, after) || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("local read created transient configuration or acquisition state")
	}
	if len(entries) != 0 {
		t.Fatalf("local read created files or invoked acquisition: %v", entries)
	}
}
