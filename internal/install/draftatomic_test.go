package install

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/transaction"
)

// This suite proves the draft publication contract at the production entry
// (install.Project): lock, marker, runtime, and adapters publish as one
// recoverable transaction, and a fault at any publication step restores
// the exact prior state. The lock itself is never written by install —
// the sweep pins that too — so "prior state" always includes the
// generation the failed run consumed.

// draftSweepClasses are the deterministic classes a draft project upgrade
// commits: replaced and added context, new runtime leaves, canonical
// shims, env files, adapter mirrors and ledger, managed removals of the
// dropped skill, and the consumer ledger last.
var draftSweepClasses = []string{
	staging.ClassContext, staging.ClassRuntime, staging.ClassCanonicalShim,
	staging.ClassEnvFile, staging.ClassAdapterLedger, staging.ClassRemoval, staging.ClassConsumer,
}

// setupDraftSweep writes a two-script local collection and resolves it
// through the production refresh path (lock plus machine bindings).
func setupDraftSweep(t *testing.T) (project, home string) {
	t.Helper()
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["*"]}]}`
	project, home = draftSweepProject(t, payload)
	writeDraftScriptSkill(t, filepath.Join(project, "skills", "review"), "review", "rtool", "#!/bin/sh\necho review-ok\n", nil)
	writeDraftScriptSkill(t, filepath.Join(project, "skills", "scan"), "scan", "stool", "#!/bin/sh\necho scan-ok\n", nil)
	refreshDraftLocked(t, project, home, payload, nil)
	return project, home
}

func draftSweepProject(t *testing.T, payload string) (string, string) {
	t.Helper()
	project := t.TempDir()
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	testGit(t, project, "init", "-q")
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n.claude/skills/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return project, home
}

// refreshDraftLocked runs the explicit resolve/refresh attempt and
// publishes the lock plus machine bindings, exactly like the CLI.
func refreshDraftLocked(t *testing.T, project, home, payload string, gitRoots map[string]string) *closure.DraftPlan {
	t.Helper()
	m, err := manifest.ParseBytes([]byte(payload), filepath.Join(project, "Skillfile.json"))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := closure.RefreshDraft(closure.DraftResolveConfig{
		ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: []byte(payload),
		Expansion: manifest.ExpansionOptions{GitRoots: gitRoots},
	}, sourcelock.PathIn(project), DraftBindingsPath(home, project))
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func draftSweepInstall(t *testing.T, home, project string, opts Options) Result {
	t.Helper()
	cfg := draftTestConfig(home, t.TempDir())
	opts.Platform = installPlatform()
	return Project(cfg, project, "test", opts)
}

type draftSharedState map[string]string

func snapshotDraftState(t *testing.T, home, project string) draftSharedState {
	t.Helper()
	paths := map[string]string{
		"lock":             sourcelock.PathIn(project),
		"bindings":         DraftBindingsPath(home, project),
		"project/skills":   filepath.Join(project, ".agents", "skills"),
		"project/bin":      filepath.Join(project, ".agents", "bin"),
		"project/env.sh":   filepath.Join(project, ".agents", "env.sh"),
		"project/env.ps1":  filepath.Join(project, ".agents", "env.ps1"),
		"project/adapters": filepath.Join(project, filepath.FromSlash(adapters.AgentPaths["claude_code"])),
		"home/runtime":     filepath.Join(home, "runtime"),
		"home/consumers":   filepath.Join(home, scopes.ConsumersName),
	}
	state := draftSharedState{}
	for key, path := range paths {
		state[key] = draftEntryDigest(path)
	}
	return state
}

// draftEntryDigest reads one snapshot path exactly as it is. A symbolic
// link is digested by its destination rather than dereferenced, so a
// mirror that was replaced, re-pointed, or removed shows up as a change.
func draftEntryDigest(path string) string {
	info, err := os.Lstat(path)
	switch {
	case os.IsNotExist(err):
		return transaction.DigestAbsent
	case err != nil:
		return "unreadable:" + err.Error()
	case info.Mode()&os.ModeSymlink != 0:
		digest, err := transaction.DigestTarget(transaction.KindEntry, path)
		if err != nil {
			return "unreadable:" + err.Error()
		}
		return digest
	case !info.IsDir():
		digest, err := transaction.DigestPath(path)
		if err != nil {
			return "unreadable:" + err.Error()
		}
		return digest
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return "unreadable:" + err.Error()
	}
	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		parts = append(parts, entry.Name()+"="+draftEntryDigest(filepath.Join(path, entry.Name())))
	}
	sort.Strings(parts)
	return "dir[" + strings.Join(parts, ",") + "]"
}

func (state draftSharedState) diff(other draftSharedState) []string {
	var changed []string
	for key, digest := range state {
		if other[key] != digest {
			changed = append(changed, fmt.Sprintf("%s: %s -> %s", key, digest, other[key]))
		}
	}
	sort.Strings(changed)
	return changed
}

// draftCommitProbe records the ordered target boundaries a commit crosses
// and fails exactly one class at PointAfterBackup, which the engine emits
// once for every target — including a removal, which never reaches the
// install boundary.
type draftCommitProbe struct {
	mu        sync.Mutex
	committed []transaction.Event
	rolled    []transaction.Event
	preimage  map[int]bool
	failed    *transaction.Event
	failClass string
	failErr   error
}

func (probe *draftCommitProbe) hooks() transaction.Hooks {
	return transaction.Hooks{
		Observe: func(event transaction.Event) {
			probe.mu.Lock()
			defer probe.mu.Unlock()
			switch event.Point {
			case transaction.PointBeforeBackup:
				if probe.preimage == nil {
					probe.preimage = map[int]bool{}
				}
				_, err := os.Lstat(event.LivePath)
				probe.preimage[event.TargetIndex] = err == nil
			case transaction.PointTargetCommitted:
				probe.committed = append(probe.committed, event)
			case transaction.PointTargetRolledBack:
				probe.rolled = append(probe.rolled, event)
			}
		},
		Fault: func(event transaction.Event) error {
			probe.mu.Lock()
			defer probe.mu.Unlock()
			if probe.failErr == nil || probe.failed != nil ||
				event.Point != transaction.PointAfterBackup || event.Class != probe.failClass {
				return nil
			}
			probe.failed = &event
			return probe.failErr
		},
	}
}

func (probe *draftCommitProbe) committedClasses() []string {
	probe.mu.Lock()
	defer probe.mu.Unlock()
	classes := make([]string, 0, len(probe.committed))
	for _, event := range probe.committed {
		classes = append(classes, event.Class)
	}
	return classes
}

// assertReverseRollback proves the restore sequence is the failing target
// followed by every committed target in exact reverse commit order. A
// target with no prior state is correctly absent: there is nothing to
// put back.
func (probe *draftCommitProbe) assertReverseRollback(t *testing.T) {
	t.Helper()
	probe.mu.Lock()
	committed := append([]transaction.Event(nil), probe.committed...)
	rolled := append([]transaction.Event(nil), probe.rolled...)
	preimage := map[int]bool{}
	for index, existed := range probe.preimage {
		preimage[index] = existed
	}
	failed := probe.failed
	probe.mu.Unlock()

	if failed == nil {
		t.Fatal("the injected fault never fired")
	}
	var want []transaction.Event
	if preimage[failed.TargetIndex] {
		want = append(want, *failed)
	}
	for index := len(committed) - 1; index >= 0; index-- {
		want = append(want, committed[index])
	}
	if len(rolled) != len(want) {
		t.Fatalf("rolled back %d targets, want %d (the failing target plus %d committed)",
			len(rolled), len(want), len(committed))
	}
	for offset, event := range rolled {
		if event.Class != want[offset].Class || event.Identifier != want[offset].Identifier {
			t.Fatalf("rollback step %d restored %s/%s, want %s/%s (exact reverse order)",
				offset, event.Class, event.Identifier, want[offset].Class, want[offset].Identifier)
		}
	}
}

func assertNoDraftJournalRemains(t *testing.T, home string) {
	t.Helper()
	root := filepath.Join(home, "state", "transactions", "v1")
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) > 0 {
		t.Fatalf("a failed install left %d journals behind for recovery to finish", len(entries))
	}
}

// TestDraftFailureAtEveryTargetClassRestoresPriorState injects a failure
// at each journaled class of a draft upgrade and proves the complete
// prior lock, marker, runtime, shim, env, adapter, removal, and consumer
// state comes back, in exact reverse commit order, with no journal left
// behind. The upgrade replaces one member, adds one, and drops one, so
// the commit carries every class the draft lane publishes.
func TestDraftFailureAtEveryTargetClassRestoresPriorState(t *testing.T) {
	project, home := setupDraftSweep(t)
	if result := draftSweepInstall(t, home, project, Options{}); result.Status != "ok" {
		t.Fatalf("baseline install failed: %+v", result)
	}
	// Upgrade: review's runtime bytes change (replaced context, marker,
	// and runtime leaf), prov arrives (added state), and scan leaves
	// (managed removals of context, shim, and mirror).
	if err := os.WriteFile(filepath.Join(project, "skills", "review", "scripts", "rtool.sh"), []byte("#!/bin/sh\necho review-v2\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeDraftScriptSkill(t, filepath.Join(project, "skills", "prov"), "prov", "ptool", "#!/bin/sh\necho prov-ok\n", nil)
	if err := os.RemoveAll(filepath.Join(project, "skills", "scan")); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(filepath.Join(project, "Skillfile.json"))
	if err != nil {
		t.Fatal(err)
	}
	plan := refreshDraftLocked(t, project, home, string(payload), nil)
	before := snapshotDraftState(t, home, project)

	for _, class := range draftSweepClasses {
		t.Run(class, func(t *testing.T) {
			probe := &draftCommitProbe{
				failClass: class,
				failErr:   fmt.Errorf("injected failure in class %s", class),
			}
			result := draftSweepInstall(t, home, project, Options{
				Commit: CommitDeps{Hooks: probe.hooks(), MaxRestarts: 1},
			})
			if result.Status != "failed" {
				t.Fatalf("install status = %q, want failed in class %s: %+v", result.Status, class, result)
			}
			if changed := before.diff(snapshotDraftState(t, home, project)); len(changed) > 0 {
				t.Fatalf("rollback did not restore the prior state: %v", changed)
			}
			for _, committed := range probe.committedClasses() {
				if committed > class {
					t.Fatalf("class %s committed after the failure in %s", committed, class)
				}
			}
			probe.assertReverseRollback(t)
			assertNoDraftJournalRemains(t, home)
		})
	}

	// The same upgrade must still succeed after the whole sweep, and the
	// successful commit is what proves every swept class was really on
	// the table — in order — and binds the refreshed lock.
	probe := &draftCommitProbe{}
	result := draftSweepInstall(t, home, project, Options{Commit: CommitDeps{Hooks: probe.hooks()}})
	if result.Status != "ok" {
		t.Fatalf("install after the rollback sweep failed: %+v", result)
	}
	classes := probe.committedClasses()
	for _, want := range draftSweepClasses {
		if !containsDraftString(classes, want) {
			t.Fatalf("upgrade never commits class %q; classes = %v", want, classes)
		}
	}
	for index := 1; index < len(classes); index++ {
		if classes[index] < classes[index-1] {
			t.Fatalf("classes committed out of order: %v", classes)
		}
	}
	recorded := marker.Read(filepath.Join(project, ".agents", "skills", "review"))
	if recorded == nil || recorded.LockSHA256 != plan.Lock.LockSHA256 {
		t.Fatalf("upgraded marker binds %v, want lock %s", recorded, plan.Lock.LockSHA256)
	}
	if _, err := os.Lstat(filepath.Join(project, ".agents", "skills", "scan")); !os.IsNotExist(err) {
		t.Fatalf("dropped skill still installed: %v", err)
	}
	assertNoDraftJournalRemains(t, home)
}

func containsDraftString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
