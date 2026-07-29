package install

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/transaction"
)

// TestPostCommitMaintenanceRunsUnderTheHeldHomeLock proves the maintenance
// sweep is part of the serialized commit phase: it receives the same held lock
// the transaction used, so consumer pruning, runtime collection, and the
// protected build-cache sweep cannot race another manager-home mutation.
func TestPostCommitMaintenanceRunsUnderTheHeldHomeLock(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")

	var seen scopes.MaintenanceRequest
	calls := 0
	result := e.install(Options{Commit: CommitDeps{
		Collect: func(request scopes.MaintenanceRequest) (scopes.MaintenanceResult, error) {
			calls++
			seen = request
			return scopes.MaintenanceResult{}, nil
		},
	}})
	if result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	if calls != 1 {
		t.Fatalf("maintenance ran %d times", calls)
	}
	if seen.Home != e.home {
		t.Fatalf("maintenance home = %q, want %q", seen.Home, e.home)
	}
	if seen.Lock == nil {
		t.Fatal("maintenance ran without a lock witness")
	}
	// The commit phase released the lock on return, so the witness the sweep
	// was handed must have been a live one at the time it ran.
	if err := seen.Lock.AssertHeld(); err == nil {
		t.Fatal("the maintenance lock witness outlived the commit phase")
	}
}

// TestPostCommitMaintenanceWarningsReachTheResult proves a retained-but-unproven
// cache entry is reported to the operator instead of being swallowed, and that
// it never reverts the durable installation.
func TestPostCommitMaintenanceWarningsReachTheResult(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")

	result := e.install(Options{Commit: CommitDeps{
		Collect: func(scopes.MaintenanceRequest) (scopes.MaintenanceResult, error) {
			return scopes.MaintenanceResult{
				Warnings: []string{"build cache sweep: retained sha256:deadbeef: invalid receipt"},
			}, nil
		},
	}})
	if result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	if !hasMessageContaining(result.Messages, "retained sha256:deadbeef") {
		t.Fatalf("the maintenance warning was not reported; messages = %v", result.Messages)
	}
	if _, err := os.Stat(filepath.Join(e.project, ".agents", "skills", "skill-a", "SKILL.md")); err != nil {
		t.Fatalf("a maintenance warning reverted the installation: %v", err)
	}
}

// TestPostCommitMaintenanceMarksInFlightJournals proves an in-flight
// transaction's build references reach the sweep, so an artifact another
// operation still depends on is retained.
func TestPostCommitMaintenanceMarksInFlightJournals(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")

	engine, err := transaction.New(e.home)
	if err != nil {
		t.Fatal(err)
	}
	var seen []string
	result := e.install(Options{Commit: CommitDeps{
		Journal: &journalKeyStub{Engine: engine, keys: []string{"sha256:" + repeatHex("a")}},
		Collect: func(request scopes.MaintenanceRequest) (scopes.MaintenanceResult, error) {
			seen = request.JournalKeys
			return scopes.MaintenanceResult{}, nil
		},
	}})
	if result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	if len(seen) != 1 || seen[0] != "sha256:"+repeatHex("a") {
		t.Fatalf("journal keys = %v", seen)
	}
}

// journalKeyStub is the real transaction engine with one fixed answer for the
// in-flight reference query, so the wiring can be observed without racing a
// second concurrent installation.
type journalKeyStub struct {
	*transaction.Engine
	keys []string
}

func (stub *journalKeyStub) ReferencedBuildKeys(transaction.HomeLock) ([]string, error) {
	return stub.keys, nil
}

func repeatHex(seed string) string {
	out := ""
	for len(out) < 64 {
		out += seed
	}
	return out[:64]
}
