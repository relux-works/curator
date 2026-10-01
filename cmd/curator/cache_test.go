package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/snapcache"
)

// writeCacheEntry publishes one commit-keyed snapshot last used at lastUsed.
func writeCacheEntry(t *testing.T, home, source, commit string, lastUsed time.Time) string {
	t.Helper()
	entry := filepath.Join(home, "cache", filepath.FromSlash(source), commit)
	if err := os.MkdirAll(filepath.Join(entry, "snapshot"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(entry, "snapshot", "SKILL.md"), []byte("skill"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(entry, lastUsed, lastUsed); err != nil {
		t.Fatal(err)
	}
	return entry
}

func runCache(t *testing.T, home string, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run(append([]string{"cache"}, args...), fileConfigSource(filepath.Join(home, "config.json")), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestCachePruneJSONReportAndRemoval(t *testing.T) {
	t.Parallel()
	home := gcHome(t)
	// installTestMarker records commit 1111...: that snapshot is reachable.
	installTestMarker(t, filepath.Join(home, "global", "skills"), "skill-a")
	old := time.Now().Add(-40 * 24 * time.Hour)
	reachable := writeCacheEntry(t, home, "skill-a", strings.Repeat("1", 40), old)
	stale := writeCacheEntry(t, home, "skill-a", strings.Repeat("2", 40), old)
	recent := writeCacheEntry(t, home, "skill-a", strings.Repeat("3", 40), time.Now().Add(-3*24*time.Hour))

	code, stdout, stderr := runCache(t, home, "prune", "--older-than", "14d", "--dry-run", "--json")
	if code != exitOK {
		t.Fatalf("dry run = %d\nstderr:\n%s", code, stderr)
	}
	var report snapcache.Report
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatalf("report is not JSON: %v\n%s", err, stdout)
	}
	var got []string
	for _, entry := range report.Entries {
		got = append(got, entry.Commit[:1]+":"+entry.Action+":"+entry.Reason)
	}
	want := []string{"1:retain:reachable", "2:remove:unreachable", "3:retain:newer_than"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("plan = %q, want %q", got, want)
	}
	if !report.DryRun || report.Policy.OlderThanSeconds == nil || *report.Policy.OlderThanSeconds != 14*86400 || report.Policy.KeepLast != nil {
		t.Fatalf("report header = %+v", report)
	}
	for _, key := range []string{`"dry_run"`, `"size_measure": "allocated"`, `"keep_last": null`, `"last_used_at"`, `"remove_allocated_bytes"`, `"warnings": []`} {
		if !strings.Contains(stdout, key) {
			t.Fatalf("report lacks %s:\n%s", key, stdout)
		}
	}
	if _, err := os.Stat(stale); err != nil {
		t.Fatal("dry run removed an entry")
	}

	code, stdout, stderr = runCache(t, home, "prune", "--older-than", "14d")
	if code != exitOK {
		t.Fatalf("prune = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "cache prune: removed 1 of 3 snapshots") {
		t.Fatalf("summary missing:\n%s", stdout)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatal("stale entry survived prune")
	}
	for _, path := range []string{reachable, recent} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("retained entry %s is gone: %v", path, err)
		}
	}
}

func TestCachePruneRemovesNothingWhenAReferenceIsUntrusted(t *testing.T) {
	t.Parallel()
	home := gcHome(t)
	stale := writeCacheEntry(t, home, "skill-a", strings.Repeat("2", 40), time.Now().Add(-40*24*time.Hour))
	if err := os.MkdirAll(filepath.Join(home, "global"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "global", "Skillfile.lock.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runCache(t, home, "prune")
	if code != exitFail || !strings.Contains(stderr, "removed nothing") || !strings.Contains(stdout, "reference uncertain") {
		t.Fatalf("prune = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if _, err := os.Stat(stale); err != nil {
		t.Fatal("prune removed an entry under an untrusted reference set")
	}
}

func TestCachePruneRejectsInvalidFlags(t *testing.T) {
	t.Parallel()
	home := gcHome(t)
	for _, args := range [][]string{
		{"prune", "--keep-last", "-1"},
		{"prune", "--keep-last", "two"},
		{"prune", "--older-than", "-3h"},
		{"prune", "--older-than", "3w"},
		{"prune", "extra"},
		{"sweep"},
		{},
	} {
		if code, _, _ := runCache(t, home, args...); code != exitUsage {
			t.Errorf("cache %q = %d, want %d", args, code, exitUsage)
		}
	}
}

func TestParseRetentionDuration(t *testing.T) {
	t.Parallel()
	for value, want := range map[string]time.Duration{"14d": 14 * 24 * time.Hour, "0d": 0, "72h": 72 * time.Hour, "90m": 90 * time.Minute} {
		got, err := parseRetentionDuration(value)
		if err != nil || got != want {
			t.Errorf("parseRetentionDuration(%q) = %s, %v; want %s", value, got, err, want)
		}
	}
}
