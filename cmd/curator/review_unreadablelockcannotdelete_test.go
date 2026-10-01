package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReviewUnreadableLockCannotDelete(t *testing.T) {
	home := gcHome(t)
	stale := writeCacheEntry(t, home, "skill-a", strings.Repeat("2", 40), time.Now().Add(-40*24*time.Hour))
	if err := os.MkdirAll(filepath.Join(home, "global", "Skillfile.lock.json"), 0755); err != nil {
		t.Fatal(err)
	}
	code, out, stderr := runCache(t, home, "prune", "--json")
	if _, err := os.Stat(stale); err != nil {
		t.Errorf("unreadable lock allowed snapshot deletion: %v", err)
	}
	if code != exitFail {
		t.Errorf("code=%d want 1; out=%s stderr=%s", code, out, stderr)
	}
}
