package install

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/transaction"
)

// Production data-flow proof for the local-source boundary guard: every
// test here drives the real staging functions (stageProjectTargets,
// stageGlobalTargets), the real serialized publisher (runCommit), or the
// real public entries (Project, Global). No test injects canned
// scopeTargets or boundary records.

// boundaryFixture is one context-only closure node with a real snapshot
// directory: enough for the production stagers to plan context, mirror,
// ledger, and env targets without any compiler or network.
type boundaryFixture struct {
	home        string
	project     string
	userHome    string
	skillsDir   string
	hybridStore string
	binDir      string
	snap        string
	node        *closure.Node
	cfg         *config.Config
}

func newBoundaryFixture(t *testing.T) *boundaryFixture {
	t.Helper()
	home := t.TempDir()
	project := t.TempDir()
	snap := t.TempDir()
	writeBoundarySkill(t, snap, "skill-a", "v1")
	spec, err := skillspec.Load(snap)
	if err != nil {
		t.Fatal(err)
	}
	fixture := &boundaryFixture{
		home:        home,
		project:     project,
		userHome:    t.TempDir(),
		skillsDir:   filepath.Join(home, "skills"),
		hybridStore: filepath.Join(home, "hybrid"),
		binDir:      filepath.Join(project, ".agents", "bin"),
		snap:        snap,
		node: &closure.Node{
			Name:     "skill-a",
			Decl:     manifest.Decl{Name: "skill-a", Source: "skill-a", Ref: manifest.Ref{Kind: "tag", Value: "v1"}},
			Snapshot: snap,
			Spec:     spec,
			Resolved: gitops.ResolvedRef{Kind: "tag", Ref: "v1", Commit: "0123456789abcdef0123456789abcdef01234567"},
			Edges:    []closure.Edge{{Consumer: closure.ProjectEdge, Mode: "full"}},
		},
		cfg: &config.Config{
			Path:        filepath.Join(home, "config.json"),
			SkillsRoot:  t.TempDir(),
			AdapterMode: "copy",
		},
	}
	return fixture
}

func writeBoundarySkill(t *testing.T, snap, name, body string) {
	t.Helper()
	payload := "---\nname: " + name + "\ndescription: d\n---\n# " + name + "\n" + body + "\n"
	if err := os.WriteFile(filepath.Join(snap, "SKILL.md"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	// A context-only skill still carries a valid manifest: marker
	// validation rejects a manifest-less spec.
	manifest := `{"schema_version": 4, "capabilities": {}, "commands": {}}` + "\n"
	if err := os.WriteFile(filepath.Join(snap, "csk-skill.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (fixture *boundaryFixture) projectRequest(stageRoot string) projectTargetRequest {
	return projectTargetRequest{
		cfg: fixture.cfg, alias: "test", projectRoot: fixture.project,
		platform: installPlatform(), nodes: []*closure.Node{fixture.node},
		agents: []string{"claude_code"}, effectiveLocale: "en",
		skillsDir: fixture.skillsDir, binDir: fixture.binDir,
		hybridStore: fixture.hybridStore,
		deps:        BuildDeps{Clock: fixedClock{at: time.Unix(1_700_000_000, 0).UTC()}},
		scoped:      scopeCommit{stageRoot: stageRoot},
	}
}

func (fixture *boundaryFixture) globalRequest(stageRoot string) globalTargetRequest {
	return globalTargetRequest{
		cfg: fixture.cfg, home: fixture.home, userHome: fixture.userHome,
		platform: installPlatform(), nodes: []*closure.Node{fixture.node},
		agents: []string{"claude_code"}, effectiveLocale: "en",
		skillsDir: fixture.skillsDir, binDir: filepath.Join(fixture.home, "global", "bin"),
		deps:   BuildDeps{Clock: fixedClock{at: time.Unix(1_700_000_000, 0).UTC()}},
		scoped: scopeCommit{stageRoot: stageRoot},
	}
}

// adapterDirOf returns the live adapter root directory for one agent below
// a scope root (project or user home), derived from the same table the
// planner stages into.
func adapterDirOf(t *testing.T, scopeRoot, agent string) string {
	t.Helper()
	rel, known := adapters.AgentPaths[agent]
	if !known {
		t.Fatalf("unknown agent %s", agent)
	}
	return filepath.Dir(filepath.Join(scopeRoot, filepath.FromSlash(rel)))
}

func findLive(t *testing.T, targets []staging.Target, live string) staging.Target {
	t.Helper()
	for _, target := range targets {
		if target.LivePath == live {
			return target
		}
	}
	t.Fatalf("no planned target for %s", live)
	return staging.Target{}
}

func TestStageProjectTargetsPopulatesBoundaries(t *testing.T) {
	t.Parallel()
	fixture := newBoundaryFixture(t)
	targets, err := stageProjectTargets(fixture.projectRequest(t.TempDir()))
	if err != nil {
		t.Fatalf("stageProjectTargets: %v", err)
	}
	if len(targets.plan.Targets) == 0 {
		t.Fatal("production staging planned no targets")
	}
	if targets.boundaries == nil {
		t.Fatal("production staging left scopeTargets.boundaries nil; the publication guard is dead")
	}
	admitted := targets.boundaries.Admitted
	if len(admitted) != 1 || admitted[0] != fixture.snap {
		t.Fatalf("boundaries.Admitted = %q, want [%s]", admitted, fixture.snap)
	}
	// The planned mirror really covers the skill, and the record the
	// publisher will recheck is self-consistent right now.
	mirrorLive := filepath.Join(fixture.project, filepath.FromSlash(adapters.AgentPaths["claude_code"]), "skill-a")
	findLive(t, targets.plan.Targets, mirrorLive)
	if err := targets.plan.Recheck(targets.boundaries.Snapshot, admitted); err != nil {
		t.Fatalf("fresh production record fails its own recheck: %v", err)
	}
	if len(targets.boundaries.Durable.Targets) == 0 {
		t.Fatal("production staging left the durable proof empty; restart recovery would fail closed")
	}
}

func TestStageProjectTargetsRefusesMirrorOverSnapshot(t *testing.T) {
	t.Parallel()
	fixture := newBoundaryFixture(t)
	// Point the live adapter root into the admitted snapshot: the mirror
	// destination then resolves inside the install's own content input.
	adapterDir := adapterDirOf(t, fixture.project, "claude_code")
	if err := os.Symlink(fixture.snap, adapterDir); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	_, err := stageProjectTargets(fixture.projectRequest(t.TempDir()))
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("stageProjectTargets err = %v, want source_output_overlap", err)
	}
	// The refusal precedes every live write: the snapshot holds exactly
	// its own files and the link itself is untouched.
	entries, err := os.ReadDir(fixture.snap)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if len(names) != 2 {
		t.Fatalf("snapshot holds %q after refusal, want only SKILL.md and csk-skill.json", names)
	}
	if info, err := os.Lstat(adapterDir); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("adapter root link disturbed by refused staging: %v", err)
	}
}

// TestStageProjectTargetsRefusesProjectInsideSnapshot proves production
// staging refuses a project rooted inside its own admitted snapshot: every
// project-joined adapter destination then resolves within admitted inputs.
func TestStageProjectTargetsRefusesProjectInsideSnapshot(t *testing.T) {
	t.Parallel()
	fixture := newBoundaryFixture(t)
	nested := filepath.Join(fixture.snap, "sub", "project")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	fixture.project = nested
	fixture.binDir = filepath.Join(nested, ".agents", "bin")
	_, err := stageProjectTargets(fixture.projectRequest(t.TempDir()))
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("stageProjectTargets err = %v, want source_output_overlap", err)
	}
}

func TestStageProjectTargetsRefusesMissingSnapshot(t *testing.T) {
	t.Parallel()
	fixture := newBoundaryFixture(t)
	fixture.node.Snapshot = ""
	_, err := stageProjectTargets(fixture.projectRequest(t.TempDir()))
	if err == nil || !strings.Contains(err.Error(), "source_member_missing") {
		t.Fatalf("staging without a snapshot err = %v, want source_member_missing", err)
	}
}

func TestStageGlobalTargetsPopulatesBoundaries(t *testing.T) {
	t.Parallel()
	fixture := newBoundaryFixture(t)
	targets, err := stageGlobalTargets(fixture.globalRequest(t.TempDir()))
	if err != nil {
		t.Fatalf("stageGlobalTargets: %v", err)
	}
	if len(targets.plan.Targets) == 0 {
		t.Fatal("global production staging planned no targets")
	}
	if targets.boundaries == nil {
		t.Fatal("global production staging left scopeTargets.boundaries nil; the publication guard is dead")
	}
	admitted := targets.boundaries.Admitted
	if len(admitted) != 1 || admitted[0] != fixture.snap {
		t.Fatalf("global boundaries.Admitted = %q, want [%s]", admitted, fixture.snap)
	}
	mirrorLive := filepath.Join(fixture.userHome, filepath.FromSlash(adapters.AgentPaths["claude_code"]), "skill-a")
	findLive(t, targets.plan.Targets, mirrorLive)
	if err := targets.plan.Recheck(targets.boundaries.Snapshot, admitted); err != nil {
		t.Fatalf("fresh global production record fails its own recheck: %v", err)
	}
	if len(targets.boundaries.Durable.Targets) == 0 {
		t.Fatal("global production staging left the durable proof empty; restart recovery would fail closed")
	}
}

func TestStageGlobalTargetsRefusesMirrorOverSnapshot(t *testing.T) {
	t.Parallel()
	fixture := newBoundaryFixture(t)
	adapterDir := adapterDirOf(t, fixture.userHome, "claude_code")
	if err := os.Symlink(fixture.snap, adapterDir); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	_, err := stageGlobalTargets(fixture.globalRequest(t.TempDir()))
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("stageGlobalTargets err = %v, want source_output_overlap", err)
	}
	entries, err := os.ReadDir(fixture.snap)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("snapshot disturbed by refused global staging: %d entries", len(entries))
	}
}

// realStagingCommit drives runCommit with the production staging function
// itself (not a canned scopeTargets) and the journal implementation the
// caller chooses.
func realStagingCommit(t *testing.T, fixture *boundaryFixture, journal TargetJournal) (commitOutcome, error) {
	t.Helper()
	return runCommit(context.Background(), commitRequest{
		scope: "test", home: fixture.home, projectRoot: fixture.project,
		commit: CommitDeps{
			Locks:            realLocks(t, fixture.home),
			Journal:          journal,
			NewTransactionID: func() (string, error) { return "test-real-staging-txn", nil },
		},
		stageTargets: func(scoped scopeCommit) (scopeTargets, error) {
			request := fixture.projectRequest(scoped.stageRoot)
			request.scoped = scoped
			return stageProjectTargets(request)
		},
	})
}

func TestRunCommitWithRealStagingJournalsOneTransaction(t *testing.T) {
	t.Parallel()
	fixture := newBoundaryFixture(t)
	journal := &stubJournal{}
	outcome, err := realStagingCommit(t, fixture, journal)
	if err != nil {
		t.Fatalf("runCommit over production staging: %v", err)
	}
	if journal.prepared != 1 || journal.commits != 1 {
		t.Fatalf("prepared=%d commits=%d, want one journaled commit", journal.prepared, journal.commits)
	}
	if len(outcome.messages) == 0 {
		t.Fatal("production staging reported no messages")
	}
}

func TestRunCommitWithRealStagingRefusesOverlapBeforeJournaling(t *testing.T) {
	t.Parallel()
	fixture := newBoundaryFixture(t)
	adapterDir := adapterDirOf(t, fixture.project, "claude_code")
	if err := os.Symlink(fixture.snap, adapterDir); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	journal := &stubJournal{}
	_, err := realStagingCommit(t, fixture, journal)
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("runCommit err = %v, want source_output_overlap", err)
	}
	if journal.prepared != 0 || journal.commits != 0 {
		t.Fatalf("refused commit reached journaling: prepared=%d commits=%d", journal.prepared, journal.commits)
	}
}

// TestPerWriteGuardRefusesSwappedParentAndRollsBack is the publication
// regression: with v1 committed, a v2 commit swaps a destination parent
// after planning and before a later write. The per-write guard must
// refuse with the boundary diagnostic, restore the already-published
// earlier target to v1, and leave the refused target at v1.
//
// The swap preserves the target's own bytes (the original child is moved
// back under the new parent), so every content-based check — the engine
// preimage verification included — passes, and only the physical-identity
// guard can refuse. That isolation is the point: without the per-write
// identity comparison the commit would succeed.
func TestPerWriteGuardRefusesSwappedParentAndRollsBack(t *testing.T) {
	fixture := newBoundaryFixture(t)
	newEngine := func(t *testing.T, hooks transaction.Hooks) *transaction.Engine {
		t.Helper()
		engine, err := transaction.New(fixture.home, transaction.WithHooks(hooks))
		if err != nil {
			t.Fatal(err)
		}
		return engine
	}
	stage := func(scoped scopeCommit) (scopeTargets, error) {
		request := fixture.projectRequest(scoped.stageRoot)
		request.scoped = scoped
		return stageProjectTargets(request)
	}
	commit := func(t *testing.T, journal TargetJournal, id string) error {
		t.Helper()
		_, err := runCommit(context.Background(), commitRequest{
			scope: "test", home: fixture.home, projectRoot: fixture.project,
			commit: CommitDeps{
				Locks:            realLocks(t, fixture.home),
				Journal:          journal,
				NewTransactionID: func() (string, error) { return id, nil },
			},
			stageTargets: stage,
		})
		return err
	}

	// Commit 1 installs v1 through the real journal with no fault.
	if err := commit(t, newEngine(t, transaction.Hooks{}), "test-perwrite-first"); err != nil {
		t.Fatalf("first commit: %v", err)
	}
	contextLive := filepath.Join(fixture.skillsDir, "skill-a", "SKILL.md")
	mirrorLive := filepath.Join(fixture.project, filepath.FromSlash(adapters.AgentPaths["claude_code"]), "skill-a", "SKILL.md")
	for _, path := range []string{contextLive, mirrorLive} {
		payload, err := os.ReadFile(path)
		if err != nil || !strings.Contains(string(payload), "\nv1\n") {
			t.Fatalf("%s after first commit = %q, %v; want v1 bytes", path, payload, err)
		}
	}

	// v2 forces every content target to re-stage on the next commit.
	writeBoundarySkill(t, fixture.snap, "skill-a", "v2")

	var hookErr error
	hookRan := false
	swapParent := func(event transaction.Event) error {
		if event.Point != transaction.PointBeforeBackup || event.Class != staging.ClassAdapterLedger || !strings.Contains(event.Identifier, "/entry/") {
			return nil
		}
		hookRan = true
		parent := filepath.Dir(event.LivePath)
		aged := parent + "-fault-old"
		if err := os.Rename(parent, aged); err != nil {
			hookErr = err
			return nil
		}
		if err := os.Mkdir(parent, 0o755); err != nil {
			hookErr = err
			return nil
		}
		// Move the original child back so its bytes are identical and
		// only the parent identity changed.
		if err := os.Rename(filepath.Join(aged, filepath.Base(event.LivePath)), event.LivePath); err != nil {
			hookErr = err
			return nil
		}
		return nil
	}
	err := commit(t, newEngine(t, transaction.Hooks{Fault: swapParent}), "test-perwrite-second")
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("second commit err = %v, want source_output_overlap", err)
	}
	if hookErr != nil {
		t.Fatalf("fault hook failed: %v", hookErr)
	}
	if !hookRan {
		t.Fatal("fault hook never fired; the refusal proves nothing about a mid-commit swap")
	}
	// The earlier context target published v2 and was rolled back to v1;
	// the refused mirror never published v2 at all.
	for _, path := range []string{contextLive, mirrorLive} {
		payload, err := os.ReadFile(path)
		if err != nil || !strings.Contains(string(payload), "\nv1\n") || strings.Contains(string(payload), "\nv2\n") {
			t.Fatalf("%s after refused commit = %q, %v; want exact v1 bytes", path, payload, err)
		}
	}
	entries, err := os.ReadDir(filepath.Join(fixture.home, "state", "transactions", "v1"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("refused commit left %d journal entries", len(entries))
	}
}

// discoverCacheSnapshot locates the extracted snapshot of one skill below
// the manager home without assuming the cache layout: it is the only
// directory whose SKILL.md carries that skill's name marker.
func discoverCacheSnapshot(t *testing.T, home, name string) string {
	t.Helper()
	var found string
	root := filepath.Join(home, "cache")
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || entry.Name() != "SKILL.md" {
			return nil
		}
		payload, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if strings.Contains(string(payload), "name: "+name+"\n") {
			found = filepath.Dir(path)
		}
		return nil
	})
	if err != nil || found == "" {
		t.Fatalf("no cache snapshot for %s below %s: %v", name, root, err)
	}
	return found
}

// TestProjectInstallRefusesAdapterDestinationOverSnapshot drives the
// public install entry end to end: after a successful install, one live
// mirror entry is relinked into the install's own content snapshot, and
// the reinstall must refuse with the boundary diagnostic instead of
// publishing mirrors over admitted inputs.
//
// The ledger from the first install still claims the entry under its
// exact spelling, so the unmanaged-takeover check passes it deliberately;
// only the destination-separation guard can refuse this shape.
func TestProjectInstallRefusesAdapterDestinationOverSnapshot(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	first := e.install(Options{})
	if first.Status != "ok" {
		t.Fatalf("first install failed: %+v", first)
	}
	snap := discoverCacheSnapshot(t, e.home, "skill-a")
	before, err := os.ReadFile(filepath.Join(snap, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	mirrorLive := filepath.Join(e.project, filepath.FromSlash(adapters.AgentPaths["claude_code"]), "skill-a")
	if err := os.RemoveAll(mirrorLive); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(snap, mirrorLive); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	second := e.install(Options{})
	if second.Status != "failed" {
		t.Fatalf("reinstall status = %q, want failed", second.Status)
	}
	joined := strings.Join(second.Errors, "\n")
	if !strings.Contains(joined, "source_output_overlap") {
		t.Fatalf("reinstall errors = %q, want source_output_overlap", joined)
	}
	after, err := os.ReadFile(filepath.Join(snap, "SKILL.md"))
	if err != nil || string(after) != string(before) {
		t.Fatalf("admitted snapshot changed by refused install: %q vs %q, %v", after, before, err)
	}
	entries, err := os.ReadDir(snap)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		switch entry.Name() {
		case "SKILL.md", "references", "scripts", "csk-skill.json", "README.md":
		default:
			t.Fatalf("snapshot gained %q from the refused install", entry.Name())
		}
	}
}

// TestGlobalInstallRefusesAdapterDestinationOverSnapshot is the global
// scope counterpart: the home-level adapter root relinked into the
// install's own content snapshot refuses the reinstall the same way.
func TestGlobalInstallRefusesAdapterDestinationOverSnapshot(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-g")
	if _, err := GlobalInit(e.home); err != nil {
		t.Fatal(err)
	}
	if err := manifestAddGlobal(e, "skill-g"); err != nil {
		t.Fatal(err)
	}
	userHome := t.TempDir()
	first := Global(e.cfg, userHome, Options{Platform: installPlatform()})
	if first.Status != "ok" {
		t.Fatalf("first global install failed: %+v", first)
	}
	snap := discoverCacheSnapshot(t, e.home, "skill-g")
	before, err := os.ReadFile(filepath.Join(snap, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	mirrorLive := filepath.Join(userHome, filepath.FromSlash(adapters.AgentPaths["claude_code"]), "skill-g")
	if err := os.RemoveAll(mirrorLive); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(snap, mirrorLive); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	second := Global(e.cfg, userHome, Options{Platform: installPlatform()})
	if second.Status != "failed" {
		t.Fatalf("global reinstall status = %q, want failed", second.Status)
	}
	joined := strings.Join(second.Errors, "\n")
	if !strings.Contains(joined, "source_output_overlap") {
		t.Fatalf("global reinstall errors = %q, want source_output_overlap", joined)
	}
	after, err := os.ReadFile(filepath.Join(snap, "SKILL.md"))
	if err != nil || string(after) != string(before) {
		t.Fatalf("admitted snapshot changed by refused global install: %q vs %q, %v", after, before, err)
	}
}
