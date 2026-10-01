package main

import (
	"github.com/relux-works/curator/internal/sourcelock"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReviewLinkedSourceLockParentCannotDelete(t *testing.T) {
	home := gcHome(t)
	stale := writeCacheEntry(t, home, "skill-a", strings.Repeat("2", 40), time.Now().Add(-40*24*time.Hour))
	target := t.TempDir()
	pkg, err := sourcelock.NetworkGitPackage("github.com/example/skill-b", sourcelock.Commit{ObjectFormat: "sha1", Hex: strings.Repeat("3", 40)}, ".")
	if err != nil {
		t.Fatal(err)
	}
	lock, err := sourcelock.New("sha256:"+strings.Repeat("a", 64), []sourcelock.Member{{Name: "skill-b", Directory: ".", Package: pkg, ContentSHA256: "sha256:" + strings.Repeat("b", 64)}})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.Write(sourcelock.PathIn(target), lock); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(home, "global")); err != nil {
		t.Fatal(err)
	}
	code, out, stderr := runCache(t, home, "prune", "--json")
	if _, err := os.Stat(stale); err != nil {
		t.Errorf("linked source-lock parent allowed deletion: %v", err)
	}
	if code != exitFail {
		t.Errorf("code=%d want 1; out=%s stderr=%s", code, out, stderr)
	}
}
