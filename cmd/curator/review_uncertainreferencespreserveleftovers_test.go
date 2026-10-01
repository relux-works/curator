package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReviewUncertainReferencesPreserveLeftovers(t *testing.T) {
	home := gcHome(t)
	writeCacheEntry(t, home, "skill-a", strings.Repeat("2", 40), time.Now().Add(-40*24*time.Hour))
	leftover := filepath.Join(home, "cache", "skill-a", ".curator-prune-"+strings.Repeat("1", 40)+"-0011")
	if err := os.MkdirAll(filepath.Join(leftover, "snapshot"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, "global"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "global", "Skillfile.lock.json"), []byte("{"), 0644); err != nil {
		t.Fatal(err)
	}
	code, out, stderr := runCache(t, home, "prune", "--json")
	if _, err := os.Stat(leftover); err != nil {
		t.Errorf("untrusted state still deleted leftover: %v; code=%d out=%s stderr=%s", err, code, out, stderr)
	}
}
