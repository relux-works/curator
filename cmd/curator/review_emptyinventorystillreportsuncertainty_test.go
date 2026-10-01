package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReviewEmptyInventoryStillReportsUncertainty(t *testing.T) {
	home := gcHome(t)
	if err := os.MkdirAll(filepath.Join(home, "global"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "global", "Skillfile.lock.json"), []byte("{"), 0644); err != nil {
		t.Fatal(err)
	}
	code, out, stderr := runCache(t, home, "prune", "--json")
	if code != exitFail {
		t.Errorf("code=%d want 1 under unreadable state; out=%s stderr=%s", code, out, stderr)
	}
}
