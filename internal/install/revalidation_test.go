package install

import (
	"os"
	"path/filepath"
	"testing"

	manifestpkg "github.com/relux-works/curator/internal/manifest"
)

// The declaration inputs a closure is resolved from are read in the window
// before the manager-home lock, and the commands that rewrite them take none of
// this run's locks:
//
//   - `curator add|remove` rewrites the project Skillfile through
//     manifest.AddDecl / manifest.RemoveDecl.
//   - `curator global add|remove` rewrites the machine-wide Skillfile the same
//     way; living under GlobalRoot(home) does not serialize it against the
//     global operation lock, because that writer never takes it.
//   - Skillfile.dev.json has no Curator writer at all and is edited by hand.
//
// Every case below mutates one of those inputs through its real writer at
// exactly the pre-home-lock boundary — OnStaged fires after every private build
// succeeded and before the serialized commit phase acquires the home lock — and
// requires the operation to restart closure resolution instead of committing
// the declarations it planned against.

// onStagedOnce returns an OnStaged hook that applies change during the first
// attempt only, so the restarted attempt sees the settled state.
func onStagedOnce(change func() error) func(Staged) error {
	applied := false
	return func(Staged) error {
		if applied {
			return nil
		}
		applied = true
		return change()
	}
}

// assertClosureRestart proves the run reported a closure-resolution restart and
// named the exact observation that moved.
func assertClosureRestart(t *testing.T, result Result, key documentKey) {
	t.Helper()
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	if !hasMessageContaining(result.Messages, "restarting from closure resolution") {
		t.Fatalf("a %s change did not restart closure resolution; messages = %v", key, result.Messages)
	}
	if !hasMessageContaining(result.Messages, "generation of "+string(key)+" changed") {
		t.Fatalf("the restart did not name %s; messages = %v", key, result.Messages)
	}
}

// assertEntriesAbsent proves nothing of a dropped declaration reached shared
// state. Every path is checked with Lstat, so a mirror link counts as present
// instead of resolving into the tree behind it.
func assertEntriesAbsent(t *testing.T, what string, paths ...string) {
	t.Helper()
	for _, path := range paths {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("stale %s committed at %s: %v", what, path, err)
		}
	}
}

func assertEntriesPresent(t *testing.T, what string, paths ...string) {
	t.Helper()
	for _, path := range paths {
		if _, err := os.Lstat(path); err != nil {
			t.Fatalf("%s missing at %s: %v", what, path, err)
		}
	}
}

// projectEntries is the context, shim, and adapter mirror one project skill
// owns: the three classes a stale closure would wrongly commit or omit.
func projectEntries(e *env, name string) []string {
	return []string{
		filepath.Join(e.project, ".agents", "skills", name),
		filepath.Join(e.project, ".agents", "bin", shimName(name+"-tool")),
		filepath.Join(e.project, ".claude", "skills", name),
	}
}

// globalEntries is the machine-wide equivalent, whose adapter mirror lands in
// the user home rather than the checkout.
func globalEntries(e *env, userHome, name string) []string {
	return []string{
		filepath.Join(GlobalRoot(e.home), "skills", name),
		filepath.Join(GlobalRoot(e.home), "bin", shimName(name+"-tool")),
		filepath.Join(userHome, ".claude", "skills", name),
	}
}

// assertNoJournalRemains proves the run left no durable transaction behind. A
// prepared or committing journal is one recovery would finish, so a restarted
// operation must not leave the discarded attempt's journal for the next run.
func assertNoJournalRemains(t *testing.T, home string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(home, "state", "transactions", "v1"))
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	var remaining []string
	for _, entry := range entries {
		remaining = append(remaining, entry.Name())
	}
	if len(remaining) > 0 {
		t.Fatalf("a restarted installation left transaction journals behind: %v", remaining)
	}
}

// globalDeclareAll initializes the machine-wide scope and declares every named
// skill through the same writer `curator global add` uses.
func (e *env) globalDeclareAll(names ...string) {
	e.t.Helper()
	if _, err := GlobalInit(e.home); err != nil {
		e.t.Fatal(err)
	}
	for _, name := range names {
		if err := manifestAddGlobal(e, name); err != nil {
			e.t.Fatal(err)
		}
	}
}

// TestStableDeclarationInputsCommitWithoutRestarting is the control the
// mutation cases below need. Observing the manifests and the substitution
// manifest must only restart a run when one of them actually moved; if every
// install restarted, the mutation assertions would prove nothing about the
// mutation.
func TestStableDeclarationInputsCommitWithoutRestarting(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")

	project := e.install(Options{})
	if project.Status != "ok" {
		t.Fatalf("project install failed: %+v", project)
	}
	if hasMessageContaining(project.Messages, "restarting from") {
		t.Fatalf("unchanged declaration inputs restarted a project run; messages = %v", project.Messages)
	}

	e.skill("skill-g")
	e.globalDeclareAll("skill-g")
	global := Global(e.cfg, t.TempDir(), Options{Platform: installPlatform()})
	if global.Status != "ok" {
		t.Fatalf("global install failed: %+v", global)
	}
	if hasMessageContaining(global.Messages, "restarting from") {
		t.Fatalf("an unchanged global manifest restarted a machine-wide run; messages = %v", global.Messages)
	}
}

// TestProjectDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState
// runs `curator remove` against the project Skillfile in the window between
// private build staging and the manager-home lock. Applying the planned closure
// would install a context, a shim, and an adapter mirror for a skill that is no
// longer declared.
func TestProjectDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.skill("skill-a")
	e.skill("skill-b")
	e.declare("skill-a", "skill-b")

	result := e.install(Options{OnStaged: onStagedOnce(func() error {
		return manifestpkg.RemoveDecl(e.project, "skill-b")
	})})

	assertClosureRestart(t, result, projectManifestKey)
	assertEntriesAbsent(t, "project state of an undeclared skill", projectEntries(e, "skill-b")...)
	assertEntriesPresent(t, "project state of the surviving skill", projectEntries(e, "skill-a")...)
	assertNoJournalRemains(t, e.home)
}

// TestProjectDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure is
// the opposite direction: a declaration that appears during staging must join
// the committed closure instead of being silently omitted.
func TestProjectDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.skill("skill-a")
	e.skill("skill-b")
	e.declare("skill-a")

	result := e.install(Options{OnStaged: onStagedOnce(func() error {
		return manifestpkg.AddDecl(e.project, "skill-b", "tag", "v1", "", "")
	})})

	assertClosureRestart(t, result, projectManifestKey)
	assertEntriesPresent(t, "project state of the added skill", projectEntries(e, "skill-b")...)
	assertEntriesPresent(t, "project state of the original skill", projectEntries(e, "skill-a")...)
	assertNoJournalRemains(t, e.home)
}

// TestGlobalDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState is
// the machine-wide removal case. The global manifest sits under the scope's own
// operation identity, which is exactly why it looked stable and was not: the
// writer takes no lock.
func TestGlobalDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.skill("skill-a")
	e.skill("skill-b")
	e.globalDeclareAll("skill-a", "skill-b")
	userHome := t.TempDir()

	result := Global(e.cfg, userHome, Options{Platform: installPlatform(), OnStaged: onStagedOnce(func() error {
		return manifestpkg.RemoveDecl(GlobalRoot(e.home), "skill-b")
	})})

	assertClosureRestart(t, result, globalManifestKey)
	assertEntriesAbsent(t, "machine-wide state of an undeclared skill", globalEntries(e, userHome, "skill-b")...)
	assertEntriesPresent(t, "machine-wide state of the surviving skill", globalEntries(e, userHome, "skill-a")...)
	assertNoJournalRemains(t, e.home)
}

// TestGlobalDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure is
// the machine-wide opposite direction.
func TestGlobalDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.skill("skill-a")
	e.skill("skill-b")
	e.globalDeclareAll("skill-a")
	userHome := t.TempDir()

	result := Global(e.cfg, userHome, Options{Platform: installPlatform(), OnStaged: onStagedOnce(func() error {
		return manifestpkg.AddDecl(GlobalRoot(e.home), "skill-b", "tag", "v1", "", "")
	})})

	assertClosureRestart(t, result, globalManifestKey)
	assertEntriesPresent(t, "machine-wide state of the added skill", globalEntries(e, userHome, "skill-b")...)
	assertEntriesPresent(t, "machine-wide state of the original skill", globalEntries(e, userHome, "skill-a")...)
	assertNoJournalRemains(t, e.home)
}

// TestDevSubstitutionAppearingBeforeHomeLockRestartsClosureResolution covers
// the third declaration input. Nothing serializes Skillfile.dev.json against an
// installation, and a substitution redirects a declaration at a local checkout,
// so a run that committed the closure it planned would install the pinned tag
// while reporting a project whose declarations now say otherwise.
func TestDevSubstitutionAppearingBeforeHomeLockRestartsClosureResolution(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	local := filepath.Join(e.skillsRoot, "skill-a")

	result := e.install(Options{OnStaged: onStagedOnce(func() error {
		e.write(e.project, "Skillfile.dev.json",
			`{"substitutions": {"skill-a": {"path": `+quoteJSONPath(local)+`}}}`)
		return nil
	})})

	assertClosureRestart(t, result, substitutionsKey)
	if !hasMessageContaining(result.Messages, "SUBSTITUTION skill-a -> path "+local) {
		t.Fatalf("the restarted attempt did not pick up the substitution; messages = %v", result.Messages)
	}
	assertEntriesPresent(t, "project state of the substituted skill", projectEntries(e, "skill-a")...)
	assertNoJournalRemains(t, e.home)
}

// quoteJSONPath renders a filesystem path as a JSON string literal, so a
// Windows separator cannot turn into an escape sequence.
func quoteJSONPath(path string) string {
	quoted := make([]rune, 0, len(path)+2)
	quoted = append(quoted, '"')
	for _, char := range path {
		if char == '\\' || char == '"' {
			quoted = append(quoted, '\\')
		}
		quoted = append(quoted, char)
	}
	return string(append(quoted, '"'))
}
