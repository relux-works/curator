package atomicity

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/scopes"
)

// The machine-level hybrid manifest decides which declarations join a project's
// effective closure, and `curator hybrid add|rm` rewrites it while holding no
// project operation lock. It is therefore read in the window before the
// manager-home lock and can move under a run that is still staging builds.
//
// These cases inject exactly that mutation at the pre-home-lock boundary —
// OnStaged fires after every private build succeeded and before the serialized
// commit phase acquires the home lock — and require the operation to restart
// closure resolution instead of committing the activation it planned against.

// mutateOnce returns an OnStaged hook that applies change during the first
// attempt only, so the restarted attempt sees the settled state.
func mutateOnce(change func()) func(install.Staged) error {
	applied := false
	return func(install.Staged) error {
		if !applied {
			applied = true
			change()
		}
		return nil
	}
}

func hasMessageContaining(messages []string, fragment string) bool {
	for _, message := range messages {
		if strings.Contains(message, fragment) {
			return true
		}
	}
	return false
}

// assertHybridActivationRestart proves the run reported a closure-resolution
// restart and named the hybrid manifest as the state that moved.
func assertHybridActivationRestart(t *testing.T, result install.Result) {
	t.Helper()
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	if !hasMessageContaining(result.Messages, "restarting from closure resolution") {
		t.Fatalf("a hybrid activation change did not restart closure resolution; messages = %v", result.Messages)
	}
	if !hasMessageContaining(result.Messages, "generation of activation/hybrid-manifest changed") {
		t.Fatalf("the restart did not name the hybrid activation manifest; messages = %v", result.Messages)
	}
}

// assertHybridSkillAbsent proves nothing of a hybrid skill reached shared state:
// neither its machine-level context nor the project adapter mirror that would
// point at it.
func assertHybridSkillAbsent(t *testing.T, e *env, name string) {
	t.Helper()
	context := filepath.Join(scopes.HybridSkillsRoot(e.home), name)
	if _, err := os.Lstat(context); !os.IsNotExist(err) {
		t.Fatalf("stale hybrid context committed at %s: %v", context, err)
	}
	mirror := filepath.Join(e.project, filepath.FromSlash(adapters.AgentPaths["claude_code"]), name)
	if state := adapterEntryState(t, mirror); state != "absent" {
		t.Fatalf("stale hybrid adapter mirror committed at %s: %s", mirror, state)
	}
}

// assertHybridSkillInstalled proves the hybrid skill is live in the machine
// store and mirrored into the project adapter root.
func assertHybridSkillInstalled(t *testing.T, e *env, name string) {
	t.Helper()
	context := filepath.Join(scopes.HybridSkillsRoot(e.home), name, "SKILL.md")
	if _, err := os.Stat(context); err != nil {
		t.Fatalf("hybrid context missing at %s: %v", context, err)
	}
	mirror := filepath.Join(e.project, filepath.FromSlash(adapters.AgentPaths["claude_code"]), name)
	if state := adapterEntryState(t, mirror); state == "absent" {
		t.Fatalf("hybrid adapter mirror missing at %s", mirror)
	}
}

// assertProjectSkillInstalled proves the restart re-derived a complete closure
// rather than dropping the work: the project's own declaration still installs.
func assertProjectSkillInstalled(t *testing.T, e *env, name string) {
	t.Helper()
	context := filepath.Join(e.project, ".agents", "skills", name, "SKILL.md")
	if _, err := os.Stat(context); err != nil {
		t.Fatalf("project context missing at %s: %v", context, err)
	}
}

// TestStableHybridActivationCommitsWithoutRestarting is the control the two
// mutation cases below need. Observing the hybrid manifest must only restart a
// run when the manifest actually moved; if every install restarted, the
// mutation assertions would prove nothing about the mutation.
func TestStableHybridActivationCommitsWithoutRestarting(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.skill("skill-a")
	e.skill("skill-h")
	e.declare("skill-a")
	e.hybridDeclare("skill-h")

	result := e.install(install.Options{})
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	if hasMessageContaining(result.Messages, "restarting from") {
		t.Fatalf("an unchanged hybrid manifest restarted the run; messages = %v", result.Messages)
	}
	assertHybridSkillInstalled(t, e, "skill-h")
	assertProjectSkillInstalled(t, e, "skill-a")
	assertNoJournalRemains(t, e.home)
}

// TestHybridDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleContext
// removes a hybrid declaration in the window between private build staging and
// the manager-home lock. Applying the planned closure would install a context
// and an adapter mirror for a skill that is no longer declared anywhere.
func TestHybridDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleContext(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.skill("skill-a")
	e.skill("skill-h")
	e.declare("skill-a")
	e.hybridDeclare("skill-h")

	result := e.install(install.Options{OnStaged: mutateOnce(func() {
		e.hybridDeclareTargeting([]string{"test"})
	})})

	assertHybridActivationRestart(t, result)
	assertHybridSkillAbsent(t, e, "skill-h")
	assertProjectSkillInstalled(t, e, "skill-a")
	assertNoJournalRemains(t, e.home)
}

// TestHybridDeclarationRetargetedBeforeHomeLockRestartsAndCommitsNoStaleContext
// keeps the declaration but points it at another project. The declaration still
// exists, so only a genuine re-evaluation of activation — not a mere
// "is it still declared" check — drops it from this project's closure.
func TestHybridDeclarationRetargetedBeforeHomeLockRestartsAndCommitsNoStaleContext(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.skill("skill-a")
	e.skill("skill-h")
	e.declare("skill-a")
	e.hybridDeclare("skill-h")

	result := e.install(install.Options{OnStaged: mutateOnce(func() {
		e.hybridDeclareTargeting([]string{"some-other-project"}, "skill-h")
	})})

	assertHybridActivationRestart(t, result)
	assertHybridSkillAbsent(t, e, "skill-h")
	assertProjectSkillInstalled(t, e, "skill-a")
	assertNoJournalRemains(t, e.home)
}

// TestHybridDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure is
// the opposite direction: a declaration that becomes applicable during staging
// must join the committed closure instead of being silently omitted. It also
// covers the absent-to-present transition of the manifest itself.
func TestHybridDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.skill("skill-a")
	e.skill("skill-h")
	e.declare("skill-a")
	if _, err := os.Lstat(scopes.HybridManifestPath(e.home)); !os.IsNotExist(err) {
		t.Fatalf("the fixture starts with a hybrid manifest, so the absent-to-present case is not covered: %v", err)
	}

	result := e.install(install.Options{OnStaged: mutateOnce(func() {
		e.hybridDeclare("skill-h")
	})})

	assertHybridActivationRestart(t, result)
	assertHybridSkillInstalled(t, e, "skill-h")
	assertProjectSkillInstalled(t, e, "skill-a")
	assertNoJournalRemains(t, e.home)
}
