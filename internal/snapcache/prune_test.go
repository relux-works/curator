package snapcache

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

type heldLock struct{ err error }

func (lock heldLock) AssertHeld() error { return lock.err }

var testNow = time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

func commitOf(ch string) string { return strings.Repeat(ch, 40) }

// writeEntry publishes one snapshot entry with the given file bytes and sets
// its last-use time (the entry directory's modification time).
func writeEntry(t *testing.T, home, source, commit string, lastUsed time.Time, files map[string]string) string {
	t.Helper()
	entry := filepath.Join(CacheRoot(home), filepath.FromSlash(source), commit)
	for name, body := range files {
		path := filepath.Join(entry, "snapshot", filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chtimes(entry, lastUsed, lastUsed); err != nil {
		t.Fatal(err)
	}
	return entry
}

func exists(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Lstat(path)
	if err == nil {
		return true
	}
	if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return false
}

func TestReadInventoryFindsNestedSourcesAndSkipsOtherStores(t *testing.T) {
	home := t.TempDir()
	writeEntry(t, home, "github.com/example/skills", commitOf("a"), testNow, map[string]string{"SKILL.md": "a"})
	writeEntry(t, home, "project-management", strings.Repeat("b", 64), testNow, map[string]string{"SKILL.md": "bb"})
	// Other stores and non-entries under the cache root.
	for _, dir := range []string{
		filepath.Join("build", "go-v1", strings.Repeat("c", 64), "snapshot"),
		filepath.Join("registry", commitOf("d"), "snapshot"),
		filepath.Join("project-management", commitOf("e"), ".snapshot-123.tmp"),
		filepath.Join("project-management", "not-a-commit", "snapshot"),
	} {
		if err := os.MkdirAll(filepath.Join(CacheRoot(home), dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	inventory, err := ReadInventory(home)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, entry := range inventory.Entries {
		got = append(got, entry.Source+"@"+entry.Commit[:1])
	}
	want := []string{"github.com/example/skills@a", "project-management@b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("entries = %q, want %q", got, want)
	}
	if !inventory.Entries[0].LastUsed.Equal(testNow) {
		t.Fatalf("last use = %s, want %s", inventory.Entries[0].LastUsed, testNow)
	}
	if inventory.Entries[1].LogicalBytes != 2 {
		t.Fatalf("logical bytes = %d, want 2", inventory.Entries[1].LogicalBytes)
	}
}

func TestReadInventoryAbsentCacheIsEmpty(t *testing.T) {
	inventory, err := ReadInventory(t.TempDir())
	if err != nil || len(inventory.Entries) != 0 {
		t.Fatalf("inventory = %+v, %v", inventory, err)
	}
}

func TestReadInventoryRefusesLinkedCacheRoot(t *testing.T) {
	home := t.TempDir()
	target := t.TempDir()
	if err := os.Symlink(target, CacheRoot(home)); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := ReadInventory(home); err == nil {
		t.Fatal("a linked cache root was traversed")
	}
}

func TestReadInventoryNeverFollowsLinkedSources(t *testing.T) {
	home := t.TempDir()
	elsewhere := t.TempDir()
	writeEntry(t, elsewhere, "victim", commitOf("a"), testNow.Add(-72*time.Hour), map[string]string{"f": "x"})
	if err := os.MkdirAll(CacheRoot(home), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(CacheRoot(elsewhere), "victim"), filepath.Join(CacheRoot(home), "linked")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	report, err := Prune(Request{Home: home, Lock: heldLock{}, Policy: Policy{Grace: DefaultGrace}, Now: testNow})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Entries) != 0 || !exists(t, filepath.Join(CacheRoot(elsewhere), "victim", commitOf("a"))) {
		t.Fatalf("a linked source was traversed: %+v", report.Entries)
	}
	if len(report.Warnings) != 1 || !strings.Contains(report.Warnings[0], "ignored link") {
		t.Fatalf("warnings = %q", report.Warnings)
	}
}

func TestMeasureCountsHardLinksOnce(t *testing.T) {
	home := t.TempDir()
	entry := writeEntry(t, home, "s", commitOf("a"), testNow, map[string]string{"one": strings.Repeat("x", 5000)})
	singleAllocated, singleLogical, err := measure(entry)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Link(filepath.Join(entry, "snapshot", "one"), filepath.Join(entry, "snapshot", "two")); err != nil {
		t.Fatal(err)
	}
	allocated, logical, err := measure(entry)
	if err != nil {
		t.Fatal(err)
	}
	if logical != 5000 {
		t.Fatalf("logical = %d, want 5000", logical)
	}
	if allocated != singleAllocated || logical != singleLogical {
		t.Fatalf("two names measured %d/%d, one identity measured %d/%d", allocated, logical, singleAllocated, singleLogical)
	}
}

// pruneFixture is one store: a reachable old entry, an unreachable old
// entry, an unreachable entry inside the grace period, and a second source
// whose only entry is unreachable and old.
func pruneFixture(t *testing.T) (home string, paths map[string]string) {
	t.Helper()
	home = t.TempDir()
	old := testNow.Add(-30 * 24 * time.Hour)
	paths = map[string]string{
		"reachable": writeEntry(t, home, "github.com/example/skills", commitOf("a"), old, map[string]string{"f": "aaaa"}),
		"stale":     writeEntry(t, home, "github.com/example/skills", commitOf("b"), old.Add(time.Hour), map[string]string{"f": "bbbbbbbb"}),
		"young":     writeEntry(t, home, "github.com/example/skills", commitOf("c"), testNow.Add(-time.Hour), map[string]string{"f": "c"}),
		"other":     writeEntry(t, home, "other", commitOf("d"), old, map[string]string{"f": "dd"}),
	}
	return home, paths
}

func TestPruneRemovesOnlyUnreachableEntriesOutsideRetention(t *testing.T) {
	home, paths := pruneFixture(t)
	refs := References{Commits: map[string]bool{commitOf("a"): true}}
	report, err := Prune(Request{Home: home, Lock: heldLock{}, Policy: Policy{Grace: DefaultGrace}, References: refs, Now: testNow})
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]bool{"reachable": true, "stale": false, "young": true, "other": false} {
		if got := exists(t, paths[name]); got != want {
			t.Errorf("%s entry exists = %v, want %v", name, got, want)
		}
	}
	// The emptied source directory goes with its last entry; the cache root stays.
	if exists(t, filepath.Join(CacheRoot(home), "other")) || !exists(t, CacheRoot(home)) {
		t.Fatal("empty source directories were not tidied up to the cache root")
	}
	var reasons []string
	for _, entry := range report.Entries {
		reasons = append(reasons, entry.Commit[:1]+":"+entry.Reason+":"+map[bool]string{true: "removed", false: "kept"}[entry.Removed])
	}
	want := []string{"a:reachable:kept", "b:unreachable:removed", "c:grace:kept", "d:unreachable:removed"}
	if !reflect.DeepEqual(reasons, want) {
		t.Fatalf("report = %q, want %q", reasons, want)
	}
	if report.Totals.Entries != 4 || report.Totals.RemoveEntries != 2 || report.Totals.RemoveLogicalBytes != 10 || report.Totals.LogicalBytes != 15 {
		t.Fatalf("totals = %+v", report.Totals)
	}
	if report.SizeMeasure != "allocated" || report.DryRun || report.Policy.GraceSeconds != 86400 || report.Policy.KeepLast != nil || report.Policy.OlderThanSeconds != nil {
		t.Fatalf("report header = %+v", report)
	}
	if leftovers, _ := filepath.Glob(filepath.Join(CacheRoot(home), "*", "*", "*", prunePrefix+"*")); len(leftovers) != 0 {
		t.Fatalf("removal left %q behind", leftovers)
	}
}

func TestPruneDryRunPlansTheSameAndRemovesNothing(t *testing.T) {
	home, paths := pruneFixture(t)
	refs := References{Commits: map[string]bool{commitOf("a"): true}}
	keep := 1
	olderThan := 7 * 24 * time.Hour
	policy := Policy{KeepLast: &keep, OlderThan: &olderThan, Grace: DefaultGrace}
	dry, err := Prune(Request{Home: home, Lock: heldLock{}, Policy: policy, References: refs, Now: testNow, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	for name, path := range paths {
		if !exists(t, path) {
			t.Fatalf("dry run removed the %s entry", name)
		}
	}
	actual, err := Prune(Request{Home: home, Lock: heldLock{}, Policy: policy, References: refs, Now: testNow})
	if err != nil {
		t.Fatal(err)
	}
	if !dry.DryRun || actual.DryRun || *dry.Policy.KeepLast != 1 || *dry.Policy.OlderThanSeconds != 604800 {
		t.Fatalf("policy or dry-run flag not reported: %+v", dry)
	}
	// The reports differ only in dry_run and each entry's removed member.
	for index := range actual.Entries {
		plan := dry.Entries[index]
		plan.Removed = actual.Entries[index].Removed
		if plan != actual.Entries[index] {
			t.Fatalf("entry %d: dry run planned %+v, actual run did %+v", index, dry.Entries[index], actual.Entries[index])
		}
	}
	if dry.Totals != actual.Totals {
		t.Fatalf("totals differ: dry %+v actual %+v", dry.Totals, actual.Totals)
	}
	// keep-last 1 keeps the young entry (newest of its source) and "other"'s
	// only entry; the stale entry is outside both windows.
	if exists(t, paths["stale"]) || !exists(t, paths["other"]) || !exists(t, paths["young"]) {
		t.Fatal("actual run did not follow its plan")
	}
}

func TestPruneUnderUncertaintyRemovesNothing(t *testing.T) {
	home, paths := pruneFixture(t)
	report, err := Prune(Request{
		Home: home, Lock: heldLock{}, Policy: Policy{Grace: DefaultGrace}, Now: testNow,
		References: References{Commits: map[string]bool{}, Uncertain: []string{"install marker x is invalid"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for name, path := range paths {
		if !exists(t, path) {
			t.Fatalf("uncertain run removed the %s entry", name)
		}
	}
	if report.Certain() || report.Totals.RemoveEntries != 0 || len(report.Warnings) != 1 {
		t.Fatalf("report = %+v", report)
	}
}

func TestPruneFinishesAnInterruptedRemoval(t *testing.T) {
	home := t.TempDir()
	leftover := filepath.Join(CacheRoot(home), "s", prunePrefix+commitOf("a")+"-0011")
	if err := os.MkdirAll(filepath.Join(leftover, "snapshot"), 0o755); err != nil {
		t.Fatal(err)
	}
	dry, err := Prune(Request{Home: home, Lock: heldLock{}, Policy: Policy{Grace: DefaultGrace}, Now: testNow, DryRun: true})
	if err != nil || !exists(t, leftover) || len(dry.Entries) != 0 || len(dry.Warnings) != 1 {
		t.Fatalf("dry run: %+v, %v", dry, err)
	}
	if _, err := Prune(Request{Home: home, Lock: heldLock{}, Policy: Policy{Grace: DefaultGrace}, Now: testNow}); err != nil {
		t.Fatal(err)
	}
	if exists(t, leftover) {
		t.Fatal("interrupted removal was not finished")
	}
}

func TestPruneRequiresTheHeldHomeLock(t *testing.T) {
	home, paths := pruneFixture(t)
	for name, lock := range map[string]HomeLock{"nil": nil, "released": heldLock{err: errors.New("released")}} {
		if _, err := Prune(Request{Home: home, Lock: lock, Policy: Policy{Grace: DefaultGrace}, Now: testNow}); err == nil {
			t.Fatalf("%s lock: prune ran", name)
		}
	}
	if !exists(t, paths["stale"]) {
		t.Fatal("prune without the lock removed an entry")
	}
}

func TestPruneUncertainLeftoversAndDryRunWarnings(t *testing.T) {
	for _, uncertain := range []bool{false, true} {
		home := t.TempDir()
		leftover := filepath.Join(CacheRoot(home), "s", prunePrefix+commitOf("a")+"-0011")
		if err := os.MkdirAll(filepath.Join(leftover, "snapshot"), 0o755); err != nil {
			t.Fatal(err)
		}
		refs := References{}
		if uncertain {
			refs.Uncertain = []string{"unreadable source lock"}
		}
		req := Request{Home: home, Lock: heldLock{}, References: refs, Now: testNow, DryRun: true}
		dry, err := Prune(req)
		if err != nil {
			t.Fatal(err)
		}
		if !exists(t, leftover) {
			t.Fatal("dry run deleted a leftover")
		}
		req.DryRun = false
		actual, err := Prune(req)
		if err != nil {
			t.Fatal(err)
		}
		if exists(t, leftover) != uncertain {
			t.Fatalf("leftover existence does not match uncertainty %t", uncertain)
		}
		if !reflect.DeepEqual(dry.Warnings, actual.Warnings) {
			t.Fatalf("dry and real warnings differ: %q / %q", dry.Warnings, actual.Warnings)
		}
	}
}

func TestPruneCertaintyIndependentOfEntryReasons(t *testing.T) {
	for _, reachableOnly := range []bool{false, true} {
		home := t.TempDir()
		refs := References{Commits: map[string]bool{commitOf("a"): true}, Uncertain: []string{"bad source lock"}}
		if reachableOnly {
			writeEntry(t, home, "s", commitOf("a"), testNow.Add(-72*time.Hour), map[string]string{"f": "x"})
		}
		report, err := Prune(Request{Home: home, Lock: heldLock{}, References: refs, Now: testNow})
		if err != nil {
			t.Fatal(err)
		}
		if report.Certain() {
			t.Fatalf("uncertainty lost with reachableOnly=%t: %+v", reachableOnly, report)
		}
		if reachableOnly && report.Entries[0].Reason != ReasonReachable {
			t.Fatal("reachable reason precedence changed")
		}
	}
}
