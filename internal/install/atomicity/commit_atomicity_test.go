package atomicity

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/transaction"
)

// sweepScenario is one complete installation shape the fault sweep walks. Each
// builds a baseline, then upgrades it so the second commit carries every
// deterministic class the scope can produce.
type sweepScenario struct {
	name string
	mode string
	// classes are the classes this scenario must actually commit; a scenario
	// that stops producing one silently would otherwise weaken the sweep.
	classes []string
	// injectClasses selects the classes this entry injects a fault at. It is a
	// selector over classes, never a replacement for it: injections inside one
	// entry are a serial chain, because each has to start from the state the
	// previous rollback restored, so the sweep is partitioned one injected
	// class per entry. The post-sweep success assertion keeps checking the
	// full classes list, so narrowing this narrows no coverage.
	injectClasses []string
	// baseline installs the prior state the rollback has to restore.
	baseline func(t *testing.T, e *env)
	// upgrade changes the declarations so the second install replaces context,
	// adds new state, and drops managed state that must be removed.
	upgrade func(t *testing.T, e *env)
	// install runs one installation of this scenario's scope.
	install func(e *env, opts install.Options) install.Result
	// userHome is the PATH-visible user home the parent prepared for this
	// entry. Only the machine-wide scope mirrors into one, and two scenarios
	// must never share it: their snapshots would overlap.
	userHome string
}

func projectSweepScenario(name, mode string, injectClasses ...string) sweepScenario {
	return sweepScenario{
		name:          name,
		mode:          mode,
		classes:       projectSweepClasses,
		injectClasses: injectClasses,
		baseline: func(t *testing.T, e *env) {
			t.Helper()
			e.skill("skill-a")
			e.skill("skill-b")
			e.skill("skill-h")
			e.skill("skill-h2")
			e.declareWithAgents([]string{"claude_code"}, "skill-a")
			e.hybridDeclare("skill-h")
			if result := e.install(install.Options{}); result.Status != "ok" {
				t.Fatalf("baseline install failed: %+v", result)
			}
			if _, err := os.Stat(filepath.Join(scopes.HybridSkillsRoot(e.home), "skill-h")); err != nil {
				t.Fatalf("the baseline installed no hybrid context: %v", err)
			}
		},
		upgrade: func(t *testing.T, e *env) {
			t.Helper()
			// skill-b arrives with a runtime tree and shim of its own, and the
			// active hybrid skill is swapped, so the commit carries project and
			// hybrid context, mirror entries in two adapter roots, and managed
			// removals of the shim and mirror of the hybrid skill that left.
			// The extra agent changes every project marker, so the surviving
			// context is genuinely replaced over a preimage rather than found
			// already current, and a second adapter root joins the commit.
			e.declareWithAgents([]string{"claude_code", "codex_cli"}, "skill-a", "skill-b")
			e.hybridDeclare("skill-h2")
		},
		install: func(e *env, opts install.Options) install.Result { return e.install(opts) },
	}
}

var projectSweepClasses = []string{
	staging.ClassContext, staging.ClassRuntime, staging.ClassCanonicalShim,
	staging.ClassEnvFile, staging.ClassAdapterLedger, staging.ClassRemoval, staging.ClassConsumer,
}

// globalSweepClasses are the classes the machine-wide scope is swept over. The
// two it alone can produce — the PATH-visible forwarding shims and the user-bin
// mirror ledger — are the reason this scenario exists; context, the adapter
// ledger, and managed removals are included because their global instances live
// in the user home rather than a checkout, which is a different rollback
// surface. The classes a global commit shares byte-for-byte with a project one
// are swept in the project scenario, which keeps the suite inside the
// per-package test timeout.
var globalSweepClasses = []string{
	staging.ClassContext, staging.ClassForwardingShim,
	staging.ClassAdapterLedger, staging.ClassMirrorLedger, staging.ClassRemoval,
}

func globalSweepScenario(name, mode, userHome string, injectClasses ...string) sweepScenario {
	return sweepScenario{
		name:          name,
		mode:          mode,
		classes:       globalSweepClasses,
		injectClasses: injectClasses,
		userHome:      userHome,
		baseline: func(t *testing.T, e *env) {
			t.Helper()
			e.skill("skill-a")
			e.skill("skill-b")
			e.skill("skill-drop")
			e.globalDeclareAll([]string{"claude_code"}, "skill-a", "skill-drop")
			if result := e.installGlobal(install.Options{}); result.Status != "ok" {
				t.Fatalf("baseline global install failed: %+v", result)
			}
			if _, err := os.Lstat(filepath.Join(e.userHome, ".local", "bin", "skill-a-tool")); err != nil {
				t.Fatalf("the baseline published no user-bin forwarding shim: %v", err)
			}
		},
		upgrade: func(t *testing.T, e *env) {
			t.Helper()
			e.globalDeclareAll([]string{"claude_code", "codex_cli"}, "skill-a", "skill-b")
		},
		install: func(e *env, opts install.Options) install.Result { return e.installGlobal(opts) },
	}
}

// TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder injects a
// failure at each target of a second installation and proves the complete prior
// project, global, hybrid, runtime, shim, env, adapter, mirror, removal, and
// consumer state comes back, in exact reverse commit order.
//
// The project scenarios run in the production-default adapter mode, so the
// mirror entries they commit and restore are the symbolic links a real machine
// gets; the global scenarios add the forwarding shims and the user-bin mirror
// ledger that only the machine-wide scope produces.
//
// Every injected class is still covered, and each injection still asserts the
// whole-state digest, the committed-class cutoff, the exact reverse rollback,
// and that no journal survives — followed by a full successful install over
// every class the scope produces. One property is deliberately retired: the
// sweep used to chain all of a scope's injections onto one baseline, which
// additionally exercised "a rollback from class X leaves nothing that trips
// class Y" for every ordered pair. That chaining was defence in depth behind
// the per-injection whole-state comparison, and it is what made the package
// exceed its race-build timeout, so it is traded for the partition below.
func TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder(t *testing.T) {
	// The machine-wide scope only publishes forwarding shims into a user bin
	// that is already on PATH, so the sweep gives every global entry one of its
	// own — the entries below must not share a user home, or their snapshots
	// would overlap. PATH is set here, in the sequential parent, because the
	// scenarios below run in parallel; every prepared bin goes on it in this
	// one call. Each entry still resolves to its own, because globalbins.Select
	// prefers <userHome>/.local/bin and its fallback scan refuses any directory
	// outside the user home it was handed.
	globalUserHomes := make([]string, len(globalSweepClasses))
	for index := range globalUserHomes {
		globalUserHomes[index] = t.TempDir()
	}
	if runtime.GOOS != "windows" {
		search := os.Getenv("PATH")
		for _, userHome := range globalUserHomes {
			userBin := filepath.Join(userHome, ".local", "bin")
			if err := os.MkdirAll(userBin, 0o755); err != nil {
				t.Fatal(err)
			}
			search = userBin + string(os.PathListSeparator) + search
		}
		t.Setenv("PATH", search)
	}

	// The sweep is partitioned one injected class per entry. Injections within
	// an entry are inherently serial — each starts from the state the previous
	// rollback restored — so a seven-long chain costs seven installs of wall
	// time that no parallelism can shorten, while seven one-long chains run at
	// once. Every entry keeps its own baseline, its own per-injection
	// whole-state comparison, and its own post-sweep success over the full
	// class list.
	scenarios := make([]sweepScenario, 0, len(projectSweepClasses)+len(globalSweepClasses))
	for _, class := range projectSweepClasses {
		scenarios = append(scenarios,
			projectSweepScenario("project-hybrid-auto/"+class, "auto", class))
	}
	for index, class := range globalSweepClasses {
		scenarios = append(scenarios,
			globalSweepScenario("global-auto/"+class, "auto", globalUserHomes[index], class))
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			// This entry's own baseline serves its own injections. A correct
			// rollback returns the machine to exactly the prior state, which
			// every injection asserts directly against this snapshot.
			e := newEnv(t)
			e.cfg.AdapterMode = scenario.mode
			if scenario.userHome != "" {
				e.userHome = scenario.userHome
			}
			scenario.baseline(t, e)
			before := snapshotState(t, e)
			scenario.upgrade(t, e)

			for _, class := range scenario.injectClasses {
				passed := t.Run(class, func(t *testing.T) {
					probe := &commitProbe{
						failClass: class,
						failErr:   fmt.Errorf("injected failure in class %s", class),
					}
					result := scenario.install(e, install.Options{
						Commit: install.CommitDeps{Hooks: probe.hooks(), MaxRestarts: 1},
					})
					// A class the scenario never produces cannot be failed, so
					// the install would succeed here: coverage of the injected
					// class is enforced rather than assumed, and the post-sweep
					// block below enforces the full scenario.classes list.
					if result.Status != "failed" {
						t.Fatalf("install status = %q, want failed in class %s: %+v", result.Status, class, result)
					}
					if changed := before.diff(snapshotState(t, e)); len(changed) > 0 {
						t.Fatalf("rollback did not restore the prior state: %v", changed)
					}
					for _, committed := range probe.committedClasses() {
						if committed > class {
							t.Fatalf("class %s committed after the failure in %s", committed, class)
						}
					}
					probe.assertReverseRollback(t)
					assertNoJournalRemains(t, e.home)
				})
				if !passed {
					// Once one rollback is wrong this entry's baseline is no
					// longer trustworthy, so any later injection on the same
					// chain would report noise rather than new information.
					break
				}
			}

			// The same operation must still succeed after the whole sweep, and
			// the successful commit is what proves every swept class was really
			// on the table.
			probe := &commitProbe{}
			result := scenario.install(e, install.Options{Commit: install.CommitDeps{Hooks: probe.hooks()}})
			if result.Status != "ok" {
				t.Fatalf("install after the rollback sweep failed: %+v", result)
			}
			classes := probe.committedClasses()
			for _, want := range scenario.classes {
				if !containsString(classes, want) {
					t.Fatalf("scenario %s never commits class %q; classes = %v", scenario.name, want, classes)
				}
			}
			for index := 1; index < len(classes); index++ {
				if classes[index] < classes[index-1] {
					t.Fatalf("classes committed out of order: %v", classes)
				}
			}
		})
	}
}

// TestAdapterMirrorLinksAreJournaledAndRestoredExactly is the symlink half of
// the atomicity contract: in the production-default auto mode and in explicit
// symlink mode, a mirror entry that a commit re-points is restored to its exact
// prior destination when a later class fails, and no prepared journal survives
// the failure for a later recovery to commit.
//
// The re-pointing is real rather than synthetic: skill-h starts in the hybrid
// store and becomes project-declared, so its canonical root — and therefore its
// mirror destination — changes.
func TestAdapterMirrorLinksAreJournaledAndRestoredExactly(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"auto", "symlink"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			e := newEnv(t)
			e.cfg.AdapterMode = mode
			e.skill("skill-a")
			e.skill("skill-h")
			e.declare("skill-a")
			e.hybridDeclare("skill-h")
			if result := e.install(install.Options{}); result.Status != "ok" {
				t.Fatalf("baseline install failed: %+v", result)
			}

			mirror := filepath.Join(e.project, ".claude", "skills", "skill-h")
			before := adapterEntryState(t, mirror)
			if !strings.HasPrefix(before, "link:") {
				t.Skipf("%s mode did not produce a link on this filesystem: %s", mode, before)
			}
			if !strings.Contains(before, "hybrid") {
				t.Fatalf("the hybrid mirror does not point at the hybrid store: %s", before)
			}
			priorState := snapshotState(t, e)

			// skill-h becomes project-declared, so the mirror has to be
			// re-pointed at the project store; then the commit fails.
			e.declare("skill-a", "skill-h")
			e.hybridDeclare()
			probe := &commitProbe{failClass: staging.ClassConsumer, failErr: fmt.Errorf("injected consumer-class failure")}
			result := e.install(install.Options{Commit: install.CommitDeps{Hooks: probe.hooks(), MaxRestarts: 1}})
			if result.Status != "failed" {
				t.Fatalf("install status = %q, want failed: %+v", result.Status, result)
			}

			if got := adapterEntryState(t, mirror); got != before {
				t.Fatalf("mirror entry = %s, want its exact prior link %s", got, before)
			}
			if changed := priorState.diff(snapshotState(t, e)); len(changed) > 0 {
				t.Fatalf("rollback did not restore the prior state: %v", changed)
			}
			probe.assertReverseRollback(t)
			assertNoJournalRemains(t, e.home)

			// A recovery pass after the failure must find nothing to finish.
			// The failed operation returned without leaving a transaction that
			// a later run would commit on its behalf.
			afterFailure := snapshotState(t, e)
			runRecovery(t, e.home)
			if changed := afterFailure.diff(snapshotState(t, e)); len(changed) > 0 {
				t.Fatalf("recovery applied state from a failed operation: %v", changed)
			}
			if got := adapterEntryState(t, mirror); got != before {
				t.Fatalf("recovery re-pointed the mirror to %s after the operation failed", got)
			}

			// The same operation must still succeed on a retry.
			if result := e.install(install.Options{}); result.Status != "ok" {
				t.Fatalf("retry after rollback failed: %+v", result)
			}
			retried := adapterEntryState(t, mirror)
			if !strings.HasPrefix(retried, "link:") || strings.Contains(retried, "hybrid") {
				t.Fatalf("mirror entry after a successful retry = %s, want a link into the project store", retried)
			}
		})
	}
}

// runRecovery finishes every incomplete transaction under one manager home,
// which is what the next installation would do before mutating anything.
func runRecovery(t *testing.T, home string) {
	t.Helper()
	engine, err := transaction.New(home)
	if err != nil {
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
	recoverErr := engine.Recover(lock)
	if closeErr := lock.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if recoverErr != nil {
		t.Fatal(recoverErr)
	}
}

// TestStaleAdapterEntryIsRemovedBeforeTheConsumerLedger closes the ordering hole
// that made stale mirror pruning a post-commit sweep: a managed entry the next
// ledger no longer claims is a journaled removal, and it commits before the
// machine-wide consumer ledger, in every adapter mode.
func TestStaleAdapterEntryIsRemovedBeforeTheConsumerLedger(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"auto", "symlink", "copy"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			e := newEnv(t)
			e.cfg.AdapterMode = mode
			e.skill("skill-a")
			e.skill("skill-drop")
			e.declare("skill-a", "skill-drop")
			if result := e.install(install.Options{}); result.Status != "ok" {
				t.Fatalf("baseline install failed: %+v", result)
			}
			stale := filepath.Join(e.project, ".claude", "skills", "skill-drop")
			if adapterEntryState(t, stale) == "absent" {
				t.Fatalf("the baseline install published no mirror for skill-drop")
			}

			e.declare("skill-a")
			probe := &commitProbe{}
			if result := e.install(install.Options{Commit: install.CommitDeps{Hooks: probe.hooks()}}); result.Status != "ok" {
				t.Fatalf("second install failed: %+v", result)
			}
			if got := adapterEntryState(t, stale); got != "absent" {
				t.Fatalf("stale mirror entry survived: %s", got)
			}

			probe.mu.Lock()
			defer probe.mu.Unlock()
			removalIndex, consumerIndex := -1, -1
			for index, event := range probe.committed {
				if event.Class == staging.ClassRemoval && strings.Contains(event.Identifier, "adapter/") {
					removalIndex = index
				}
				if event.Class == staging.ClassConsumer {
					consumerIndex = index
				}
			}
			if removalIndex < 0 {
				t.Fatalf("the stale mirror was not a journaled removal: %+v", probe.committed)
			}
			if consumerIndex < 0 {
				t.Fatal("the project install committed no consumer ledger")
			}
			if removalIndex > consumerIndex {
				t.Fatal("the consumer ledger committed before the stale mirror was removed")
			}
			if last := probe.committed[len(probe.committed)-1]; last.Class != staging.ClassConsumer {
				t.Fatalf("last committed class = %q, want the consumer ledger", last.Class)
			}
		})
	}
}

// TestStaleAdapterRemovalRollsBackToTheExactPriorEntry proves the removal class
// is restorable, which is the whole reason stale pruning moved into the journal.
func TestStaleAdapterRemovalRollsBackToTheExactPriorEntry(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.cfg.AdapterMode = "auto"
	e.skill("skill-a")
	e.skill("skill-drop")
	e.declare("skill-a", "skill-drop")
	if result := e.install(install.Options{}); result.Status != "ok" {
		t.Fatalf("baseline install failed: %+v", result)
	}
	stale := filepath.Join(e.project, ".claude", "skills", "skill-drop")
	before := adapterEntryState(t, stale)
	if before == "absent" {
		t.Fatal("the baseline install published no mirror for skill-drop")
	}
	priorState := snapshotState(t, e)

	e.declare("skill-a")
	probe := &commitProbe{failClass: staging.ClassConsumer, failErr: fmt.Errorf("injected consumer-class failure")}
	result := e.install(install.Options{Commit: install.CommitDeps{Hooks: probe.hooks(), MaxRestarts: 1}})
	if result.Status != "failed" {
		t.Fatalf("install status = %q, want failed: %+v", result.Status, result)
	}
	if got := adapterEntryState(t, stale); got != before {
		t.Fatalf("removed mirror entry came back as %s, want its exact prior state %s", got, before)
	}
	if changed := priorState.diff(snapshotState(t, e)); len(changed) > 0 {
		t.Fatalf("rollback did not restore the prior state: %v", changed)
	}
	assertNoJournalRemains(t, e.home)
}
