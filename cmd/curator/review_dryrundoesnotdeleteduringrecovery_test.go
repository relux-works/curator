package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/snapcache"
	"github.com/relux-works/curator/internal/transaction"
	"os"
	"path/filepath"
	"testing"
)

func TestReviewDryRunDoesNotDeleteDuringRecovery(t *testing.T) {
	home := gcHome(t)
	live := filepath.Join(home, "pending-uninstall-file")
	if err := os.WriteFile(live, []byte("keep me"), 0600); err != nil {
		t.Fatal(err)
	}
	manager, err := managerlock.New(home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := transaction.New(home)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := transaction.DigestPath(live)
	if err != nil {
		t.Fatal(err)
	}
	_, err = engine.Prepare(lock, transaction.Plan{TransactionID: "review-pending-uninstall", ProjectIdentity: home, Targets: []transaction.Target{{Class: "installed-file", Identifier: "file", Kind: transaction.KindBytes, LivePath: live, PreimageDigest: digest}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	journalPath := filepath.Join(home, "state", "transactions", "v1", "review-pending-uninstall.json")
	before, err := os.ReadFile(journalPath)
	if err != nil {
		t.Fatal(err)
	}
	code, out, stderr := runCache(t, home, "prune", "--dry-run", "--json")
	if _, err := os.Stat(live); err != nil {
		t.Errorf("dry-run deleted pending transaction target: %v; code=%d out=%s stderr=%s", err, code, out, stderr)
	}
	if code != exitFail {
		t.Fatalf("dry-run with recovery pending returned %d: %s %s", code, out, stderr)
	}
	var report snapcache.Report
	if err := json.Unmarshal([]byte(out), &report); err != nil || !report.DryRun {
		t.Fatalf("dry-run report = %s: %v", out, err)
	}
	after, err := os.ReadFile(journalPath)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("dry-run changed journal: %v", err)
	}
	code, out, stderr = runCache(t, home, "prune", "--json")
	if code != exitOK {
		t.Fatalf("real recovery: %d %s %s", code, out, stderr)
	}
	if _, err := os.Stat(live); !os.IsNotExist(err) {
		t.Fatalf("real recovery did not finish pending deletion: %v", err)
	}

}
