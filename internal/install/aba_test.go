package install

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/adapters"
	devsubpkg "github.com/relux-works/curator/internal/devsub"
	manifestpkg "github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/scopes"
)

// A one-way mutation of a declaration document is not the whole race. Every
// supported writer rewrites its manifest in place, so a declaration can also go
// A -> B -> A *around* the read: the closure is resolved from the transient B
// while both the state before and the state at commit time are byte-identical A.
// A generation taken from the path as a separate operation cannot see that — both
// samples are A — and the run commits context, shims, adapters, and consumer
// state for a declaration set that exists nowhere.
//
// Each case below drives exactly that sequence through the real writer of its
// document: the transient generation is applied inside the single read that feeds
// the parser, and the byte-identical restoration lands at the pre-home-lock
// boundary, where OnStaged fires after every private build succeeded and before
// the commit phase acquires the manager-home lock. Each then requires a closure
// restart and proves the transient declaration reached no shared state.

// duringDocumentRead applies change inside the one read of target that binds a
// generation to the bytes the parser consumes. It fires once, so the restarted
// attempt reads the settled document, and it fails the test if the window never
// opened — a case whose mutation missed the read would assert nothing.
func duringDocumentRead(t *testing.T, target string, change func() error) {
	t.Helper()
	fired := false
	afterDocumentOpen = func(path string) {
		if fired || path != target {
			return
		}
		fired = true
		if err := change(); err != nil {
			t.Errorf("apply the transient generation of %s: %v", path, err)
		}
	}
	t.Cleanup(func() {
		afterDocumentOpen = nil
		if !fired {
			t.Errorf("the read window never opened for %s, so no transient generation was parsed", target)
		}
	})
}

// settledBytes records the document a case starts from, after canonicalizing it
// through its real writer, so the restoration can be proven byte-identical
// instead of merely equivalent.
func settledBytes(t *testing.T, path string) []byte {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

// assertRestoredByteIdentically proves the case really exercised the ABA window
// rather than the one-way mutation the earlier cases already cover: if the
// restoration changed a single byte, the recheck would catch the difference for
// the wrong reason.
func assertRestoredByteIdentically(t *testing.T, path string, settled []byte) {
	t.Helper()
	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(current, settled) {
		t.Fatalf("the exercise did not restore %s byte-identically:\nsettled = %q\ncurrent = %q",
			path, settled, current)
	}
}

// TestProjectManifestABAAroundTheReadRestartsClosure adds a declaration through
// `curator add`'s writer inside the manifest read and removes it again through
// `curator remove`'s writer before the home lock. The manifest ends byte-identical
// to where it started, so only a generation bound to the bytes the parser
// consumed can tell that the planned closure was resolved from neither state.
func TestProjectManifestABAAroundTheReadRestartsClosure(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.skill("skill-b")
	e.declare("skill-a")
	// Canonicalize through the real writer, so removing skill-b restores exactly
	// these bytes rather than the fixture's own formatting.
	if err := manifestpkg.AddDecl(e.project, "skill-a", "tag", "v1", "", ""); err != nil {
		t.Fatal(err)
	}
	manifestPath := manifestpkg.PathIn(e.project)
	settled := settledBytes(t, manifestPath)

	duringDocumentRead(t, manifestPath, func() error {
		return manifestpkg.AddDecl(e.project, "skill-b", "tag", "v1", "", "")
	})
	result := e.install(Options{OnStaged: onStagedOnce(func() error {
		return manifestpkg.RemoveDecl(e.project, "skill-b")
	})})

	assertRestoredByteIdentically(t, manifestPath, settled)
	assertClosureRestart(t, result, projectManifestKey)
	assertEntriesAbsent(t, "project state of a transient declaration", projectEntries(e, "skill-b")...)
	assertEntriesPresent(t, "project state of the declared skill", projectEntries(e, "skill-a")...)
	assertNoJournalRemains(t, e.home)
}

// TestGlobalManifestABAAroundTheReadRestartsClosure is the machine-wide case,
// through `curator global add|remove`'s writer. Its consumer ledger is the
// machine-wide one, so a transient declaration that committed here would be
// visible to every project on the machine.
func TestGlobalManifestABAAroundTheReadRestartsClosure(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.skill("skill-b")
	e.globalDeclareAll("skill-a")
	userHome := t.TempDir()
	manifestPath := manifestpkg.PathIn(GlobalRoot(e.home))
	settled := settledBytes(t, manifestPath)

	duringDocumentRead(t, manifestPath, func() error {
		return manifestpkg.AddDecl(GlobalRoot(e.home), "skill-b", "tag", "v1", "", "")
	})
	result := Global(e.cfg, userHome, Options{Platform: "unix", OnStaged: onStagedOnce(func() error {
		return manifestpkg.RemoveDecl(GlobalRoot(e.home), "skill-b")
	})})

	assertRestoredByteIdentically(t, manifestPath, settled)
	assertClosureRestart(t, result, globalManifestKey)
	assertEntriesAbsent(t, "machine-wide state of a transient declaration", globalEntries(e, userHome, "skill-b")...)
	assertEntriesPresent(t, "machine-wide state of the declared skill", globalEntries(e, userHome, "skill-a")...)
	assertNoJournalRemains(t, e.home)
}

// TestSubstitutionsABAAroundTheReadRestartsClosure covers the document with no
// Curator writer at all: Skillfile.dev.json is edited by hand, so an editor that
// writes, saves again, and reverts is the ordinary case rather than a race. A
// transient substitution redirects the declaration at a local checkout, so
// committing the planned closure would install content the settled project never
// asked for and record it in the marker.
func TestSubstitutionsABAAroundTheReadRestartsClosure(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	local := filepath.Join(e.skillsRoot, "skill-a")
	empty := "{\"substitutions\": {}}\n"
	substituted := `{"substitutions": {"skill-a": {"path": ` + quoteJSONPath(local) + "}}}\n"
	e.write(e.project, devsubpkg.Name, empty)
	substitutionsPath := devsubpkg.PathIn(e.project)
	settled := settledBytes(t, substitutionsPath)

	duringDocumentRead(t, substitutionsPath, func() error {
		return os.WriteFile(substitutionsPath, []byte(substituted), 0o644)
	})
	result := e.install(Options{OnStaged: onStagedOnce(func() error {
		return os.WriteFile(substitutionsPath, []byte(empty), 0o644)
	})})

	assertRestoredByteIdentically(t, substitutionsPath, settled)
	assertClosureRestart(t, result, substitutionsKey)
	if hasMessageContaining(result.Messages, "SUBSTITUTION") {
		t.Fatalf("a transient substitution survived into the committed run; messages = %v", result.Messages)
	}
	assertEntriesPresent(t, "project state of the declared skill", projectEntries(e, "skill-a")...)
	if recorded := readMarkerFor(t, e, "skill-a").Substituted; recorded != "" {
		t.Fatalf("the committed marker records a transient substitution: %s", recorded)
	}
	assertNoJournalRemains(t, e.home)
}

// TestHybridActivationABAAroundTheReadRestartsClosure covers machine-home
// activation through `curator hybrid add`'s writer. The declaration exists in
// every state here — only its target list moves, and it moves back — so nothing
// short of re-evaluating activation against the bytes the closure was built from
// can drop the transiently activated skill.
func TestHybridActivationABAAroundTheReadRestartsClosure(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.skill("skill-h")
	e.declare("skill-a")
	if err := scopes.AddHybridDecl(e.home, "skill-h", "tag", "v1", "", []string{"some-other-project"}); err != nil {
		t.Fatal(err)
	}
	hybridPath := scopes.HybridManifestPath(e.home)
	settled := settledBytes(t, hybridPath)

	duringDocumentRead(t, hybridPath, func() error {
		return scopes.AddHybridDecl(e.home, "skill-h", "tag", "v1", "", []string{"test"})
	})
	result := e.install(Options{OnStaged: onStagedOnce(func() error {
		return scopes.AddHybridDecl(e.home, "skill-h", "tag", "v1", "", []string{"some-other-project"})
	})})

	assertRestoredByteIdentically(t, hybridPath, settled)
	assertClosureRestart(t, result, hybridActivationKey)
	assertEntriesAbsent(t, "hybrid state of a transiently activated skill",
		filepath.Join(scopes.HybridSkillsRoot(e.home), "skill-h"),
		filepath.Join(e.project, filepath.FromSlash(adapters.AgentPaths["claude_code"]), "skill-h"))
	assertEntriesPresent(t, "project state of the declared skill", projectEntries(e, "skill-a")...)
	assertNoJournalRemains(t, e.home)
}

// TestByteIdenticalRewriteDuringTheReadDoesNotRestart is the control every case
// above needs in the other direction. Binding the generation to the parsed bytes
// must not turn every concurrent save into a restart: a writer that rewrites a
// document with the same content selects nothing new, and the run has to commit
// its planned closure instead of looping until MaxRestarts.
func TestByteIdenticalRewriteDuringTheReadDoesNotRestart(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	manifestPath := manifestpkg.PathIn(e.project)
	settled := settledBytes(t, manifestPath)

	duringDocumentRead(t, manifestPath, func() error {
		return os.WriteFile(manifestPath, settled, 0o644)
	})
	result := e.install(Options{})

	assertRestoredByteIdentically(t, manifestPath, settled)
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	if hasMessageContaining(result.Messages, "restarting from") {
		t.Fatalf("a rewrite that changed no byte restarted the run; messages = %v", result.Messages)
	}
	assertEntriesPresent(t, "project state of the declared skill", projectEntries(e, "skill-a")...)
	assertNoJournalRemains(t, e.home)
}
