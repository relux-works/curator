package install

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/transaction"
)

// sharedState is a digest of every class of state one installation commits.
// A rollback is only correct if the complete map returns to its prior value,
// so the tests below compare whole snapshots rather than individual paths.
type sharedState map[string]string

func snapshotState(t *testing.T, e *env) sharedState {
	t.Helper()
	paths := map[string]string{
		"project/skills":   filepath.Join(e.project, ".agents", "skills"),
		"project/bin":      filepath.Join(e.project, ".agents", "bin"),
		"project/env.sh":   filepath.Join(e.project, ".agents", "env.sh"),
		"project/env.ps1":  filepath.Join(e.project, ".agents", "env.ps1"),
		"project/adapters": filepath.Join(e.project, ".claude", "skills"),
		"project/adapters-codex": filepath.Join(e.project,
			filepath.FromSlash(adapters.AgentPaths["codex_cli"])),
		"hybrid/store":   scopes.HybridSkillsRoot(e.home),
		"home/runtime":   filepath.Join(e.home, "runtime"),
		"home/consumers": filepath.Join(e.home, scopes.ConsumersName),
		"global/skills":  filepath.Join(e.home, "global", "skills"),
		"global/bin":     filepath.Join(e.home, "global", "bin"),
		"global/env.sh":  filepath.Join(e.home, "global", "env.sh"),
	}
	state := sharedState{}
	for key, path := range paths {
		state[key] = entryDigest(path)
	}
	return state
}

// entryDigest reads one snapshot path exactly as it is. A symbolic link is
// digested by its destination rather than dereferenced, so a mirror that was
// replaced, re-pointed, or removed shows up as a change instead of resolving to
// the same canonical bytes.
func entryDigest(path string) string {
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
	// A directory may legitimately hold mirror links, which DigestPath refuses
	// inside a tree, so the tree is summarized entry by entry.
	entries, err := os.ReadDir(path)
	if err != nil {
		return "unreadable:" + err.Error()
	}
	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		parts = append(parts, entry.Name()+"="+entryDigest(filepath.Join(path, entry.Name())))
	}
	sort.Strings(parts)
	return "dir[" + strings.Join(parts, ",") + "]"
}

func (state sharedState) diff(other sharedState) []string {
	var changed []string
	for key, digest := range state {
		if other[key] != digest {
			changed = append(changed, fmt.Sprintf("%s: %s -> %s", key, digest, other[key]))
		}
	}
	sort.Strings(changed)
	return changed
}

// commitProbe records the ordered target boundaries a commit crosses and can
// fail exactly one of them.
//
// The fault fires at PointAfterBackup, which the engine emits once for every
// target — including a removal, whose desired state is already present once its
// preimage has been moved aside and which therefore never reaches the install
// boundary. Faulting there is what makes a sweep able to reach every class.
type commitProbe struct {
	mu        sync.Mutex
	committed []transaction.Event
	rolled    []transaction.Event
	// preimage records, per target index, whether the live path held anything
	// before the commit touched it. A target with nothing to restore is
	// correctly absent from the rollback sequence.
	preimage map[int]bool
	failed   *transaction.Event
	failAt   int
	failErr  error
}

func (probe *commitProbe) hooks() transaction.Hooks {
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
				event.Point != transaction.PointAfterBackup || event.TargetIndex != probe.failAt {
				return nil
			}
			probe.failed = &event
			return probe.failErr
		},
	}
}

func (probe *commitProbe) committedClasses() []string {
	probe.mu.Lock()
	defer probe.mu.Unlock()
	classes := make([]string, 0, len(probe.committed))
	for _, event := range probe.committed {
		classes = append(classes, event.Class)
	}
	return classes
}

// hookedLocks wraps the real broker so a test can act in the window between
// private staging and the manager-home lock.
type hookedLocks struct {
	inner      LockBroker
	beforeHome func(attempt int)
	attempts   int
}

func (locks *hookedLocks) AcquireProjects(ctx context.Context, projects ...string) (ProjectLocks, error) {
	return locks.inner.AcquireProjects(ctx, projects...)
}

func (locks *hookedLocks) AcquireHome(ctx context.Context) (HomeLock, error) {
	locks.attempts++
	if locks.beforeHome != nil {
		locks.beforeHome(locks.attempts)
	}
	return locks.inner.AcquireHome(ctx)
}

func realLocks(t *testing.T, home string) LockBroker {
	t.Helper()
	manager, err := managerlock.New(home)
	if err != nil {
		t.Fatal(err)
	}
	return &managerLocks{manager: manager}
}

// TestCommitOrdersTargetClassesWithConsumerLast proves the commit walks the
// normative classes in order and updates the machine-wide consumer ledger last.
func TestCommitOrdersTargetClassesWithConsumerLast(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	probe := &commitProbe{}

	result := e.install(Options{Commit: CommitDeps{Hooks: probe.hooks()}})
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}

	classes := probe.committedClasses()
	if len(classes) == 0 {
		t.Fatal("the commit journaled no target")
	}
	for index := 1; index < len(classes); index++ {
		if classes[index] < classes[index-1] {
			t.Fatalf("classes committed out of order: %v", classes)
		}
	}
	if got := classes[len(classes)-1]; got != staging.ClassConsumer {
		t.Fatalf("last committed class = %q, want the consumer ledger %q", got, staging.ClassConsumer)
	}
	for _, want := range []string{staging.ClassContext, staging.ClassRuntime, staging.ClassCanonicalShim, staging.ClassEnvFile, staging.ClassAdapterLedger} {
		if !containsString(classes, want) {
			t.Fatalf("class %q never committed; classes = %v", want, classes)
		}
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// TestConsumerLedgerIsAbsentAfterAFailedFirstInstallAndCommitsLastOnSuccess
// pins the two halves of the consumer-ledger rule on a machine that has never
// installed anything.
func TestConsumerLedgerIsAbsentAfterAFailedFirstInstallAndCommitsLastOnSuccess(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	consumers := filepath.Join(e.home, scopes.ConsumersName)

	probe := &commitProbe{failAt: 0, failErr: fmt.Errorf("injected first-target failure")}
	result := e.install(Options{Commit: CommitDeps{Hooks: probe.hooks(), MaxRestarts: 1}})
	if result.Status != "failed" {
		t.Fatalf("install status = %q, want failed: %+v", result.Status, result)
	}
	if _, err := os.Lstat(consumers); !os.IsNotExist(err) {
		t.Fatalf("consumer ledger exists after a failed first install: %v", err)
	}

	if result := e.install(Options{}); result.Status != "ok" {
		t.Fatalf("second install failed: %+v", result)
	}
	registered := scopes.LoadConsumers(e.home)
	if len(registered) != 1 || registered[0] != e.project {
		t.Fatalf("consumers = %v, want exactly the installed checkout %q", registered, e.project)
	}
}

// TestCachePublicationFailureLeavesInstallationAndProtectedCacheUntouched
// proves a publication fault aborts before the first target swap and preserves
// every pre-existing immutable cache entry.
func TestCachePublicationFailureLeavesInstallationAndProtectedCacheUntouched(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha", "beta")
	e.declare("build-skill")
	deps, _, cache, _ := newFakeDeps(t)
	cache.seedHit("alpha")

	if result := e.install(Options{Build: deps}); result.Status != "ok" {
		t.Fatalf("baseline install failed: %+v", result)
	}
	before := snapshotState(t, e)
	// Immutable entries are additive: a rollback must preserve what was already
	// protected, not withdraw what a later publication legitimately added.
	protectedBefore := map[string]string{}
	for command, result := range cache.byCommand {
		digest, err := transaction.DigestPath(result.ArtifactPath)
		if err != nil {
			t.Fatal(err)
		}
		protectedBefore[command] = digest
	}

	// Force a rebuild of both commands, then fail publication of one of them.
	e.skill("skill-a")
	e.declare("build-skill", "skill-a")
	deps2, _, cache2, _ := newFakeDeps(t)
	cache2.root = cache.root
	cache2.publishErr["beta"] = fmt.Errorf("injected publication failure")

	result := e.install(Options{Build: deps2, Commit: CommitDeps{MaxRestarts: 1}})
	if result.Status != "failed" {
		t.Fatalf("install status = %q, want failed: %+v", result.Status, result)
	}
	if changed := before.diff(snapshotState(t, e)); len(changed) > 0 {
		t.Fatalf("a publication failure changed installed state: %v", changed)
	}
	for command, want := range protectedBefore {
		got, err := transaction.DigestPath(cache.byCommand[command].ArtifactPath)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("a publication failure changed the pre-existing immutable entry for %q", command)
		}
	}
}

// Environment contract of the concurrent-install helper process.
const (
	helperInstallEnv = "CURATOR_INSTALL_HELPER_MODE"
	helperHomeEnv    = "CURATOR_INSTALL_HELPER_HOME"
	helperSkillsEnv  = "CURATOR_INSTALL_HELPER_SKILLS"
	helperProjectEnv = "CURATOR_INSTALL_HELPER_PROJECT"
)

// TestConcurrentProjectInstallsPreserveBothConsumers proves the consumer ledger
// is merged under the manager-home lock rather than staged before it: two
// checkouts installing at once must both end up registered, and neither may
// overwrite the other's record.
//
// The two installs run as separate processes because the manager lock layer
// deliberately serializes lock-class acquisition per process. Sharing one
// process would test that guard rather than the ledger merge this asserts.
func TestConcurrentProjectInstallsPreserveBothConsumers(t *testing.T) {
	home := t.TempDir()
	skillsRoot := t.TempDir()
	first := newEnvIn(t, home, skillsRoot)
	second := newEnvIn(t, home, skillsRoot)
	first.skill("skill-a")
	first.declare("skill-a")
	second.declare("skill-a")

	var group sync.WaitGroup
	failures := make([]error, 2)
	for index, scope := range []*env{first, second} {
		group.Add(1)
		go func(index int, project string) {
			defer group.Done()
			command := exec.Command(os.Args[0], "-test.run=^TestInstallHelperProcess$")
			command.Env = append(os.Environ(),
				helperInstallEnv+"=1",
				helperHomeEnv+"="+home,
				helperSkillsEnv+"="+skillsRoot,
				helperProjectEnv+"="+project,
			)
			if output, err := command.CombinedOutput(); err != nil {
				failures[index] = fmt.Errorf("%w\n%s", err, output)
			}
		}(index, scope.project)
	}
	group.Wait()

	for index, err := range failures {
		if err != nil {
			t.Fatalf("concurrent install %d failed: %v", index, err)
		}
	}
	registered := scopes.LoadConsumers(home)
	for _, want := range []string{first.project, second.project} {
		if !containsString(registered, want) {
			t.Fatalf("consumers = %v, want both checkouts including %q", registered, want)
		}
	}
	for _, scope := range []*env{first, second} {
		if _, err := os.Stat(filepath.Join(scope.project, ".agents", "skills", "skill-a", "SKILL.md")); err != nil {
			t.Fatalf("checkout %s is not installed after the concurrent run: %v", scope.project, err)
		}
	}
}

// TestInstallHelperProcess installs one checkout and is only a test by name; it
// is the child half of the concurrency assertion above.
func TestInstallHelperProcess(t *testing.T) {
	if os.Getenv(helperInstallEnv) != "1" {
		t.Skip("helper process")
	}
	home := os.Getenv(helperHomeEnv)
	project := os.Getenv(helperProjectEnv)
	cfg := &config.Config{
		Path:          filepath.Join(home, "config.json"),
		SkillsRoot:    os.Getenv(helperSkillsEnv),
		DefaultAgents: []string{"claude_code"},
		AdapterMode:   "auto",
	}
	result := Project(cfg, project, filepath.Base(project), Options{Platform: "unix"})
	if result.Status != "ok" {
		t.Fatalf("helper install failed: %+v", result)
	}
}

// TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets drives a
// second project to a full commit inside the window where the first project has
// already staged privately but has not yet taken the manager-home lock. The
// first project then fails and rolls back; the second project's committed
// shared state must survive, because the rolled-back preimage was read under
// the same lock and therefore cannot predate that commit.
func TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets(t *testing.T) {
	home := t.TempDir()
	skillsRoot := t.TempDir()
	failing := newEnvIn(t, home, skillsRoot)
	winner := newEnvIn(t, home, skillsRoot)
	failing.skill("skill-a")
	failing.declare("skill-a")
	winner.declare("skill-a")

	var once sync.Once
	locks := &hookedLocks{inner: realLocks(t, home)}
	locks.beforeHome = func(int) {
		once.Do(func() {
			if result := winner.install(Options{}); result.Status != "ok" {
				t.Errorf("the winning install failed: %+v", result)
			}
		})
	}
	probe := &commitProbe{failAt: 0, failErr: fmt.Errorf("injected failure after another project committed")}

	result := failing.install(Options{Commit: CommitDeps{
		Locks: locks, Hooks: probe.hooks(), MaxRestarts: 1,
	}})
	if result.Status != "failed" {
		t.Fatalf("install status = %q, want failed: %+v", result.Status, result)
	}

	registered := scopes.LoadConsumers(home)
	if !containsString(registered, winner.project) {
		t.Fatalf("consumers = %v, want the committed checkout %q to survive the other project's rollback",
			registered, winner.project)
	}
	if containsString(registered, failing.project) {
		t.Fatalf("consumers = %v, want no record of the rolled-back checkout %q", registered, failing.project)
	}
	if _, err := os.Stat(filepath.Join(winner.project, ".agents", "skills", "skill-a", "SKILL.md")); err != nil {
		t.Fatalf("the committed project context did not survive the other project's rollback: %v", err)
	}
}

// TestRecoveryCompletesBeforeAnyNewMutation leaves a prepared transaction on
// the machine and proves the next installation finishes it before it changes
// anything of its own.
func TestRecoveryCompletesBeforeAnyNewMutation(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")

	// A prepared-but-uncommitted transaction that installs one manager file.
	engine, err := transaction.New(e.home)
	if err != nil {
		t.Fatal(err)
	}
	manager, err := managerlock.New(e.home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	stage := t.TempDir()
	staged := filepath.Join(stage, "recovered.json")
	if err := os.WriteFile(staged, []byte("{\"recovered\":true}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	live := filepath.Join(e.home, "recovered.json")
	if _, err := engine.Prepare(lock, transaction.Plan{
		TransactionID:   "pending-recovery",
		ProjectIdentity: e.project,
		Targets: []transaction.Target{{
			Class: staging.ClassContext, Identifier: "recovered",
			LivePath: live, StagedSource: staged, PreimageDigest: transaction.DigestAbsent,
		}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(live); !os.IsNotExist(err) {
		t.Fatalf("preparation already published the target: %v", err)
	}

	recoveredBeforeInstall := false
	locks := &hookedLocks{inner: realLocks(t, e.home)}
	probe := &commitProbe{}
	hooks := probe.hooks()
	observe := hooks.Observe
	hooks.Observe = func(event transaction.Event) {
		if event.Point == transaction.PointBeforeInstall && !recoveredBeforeInstall {
			if _, err := os.Lstat(live); err == nil {
				recoveredBeforeInstall = true
			}
		}
		observe(event)
	}

	result := e.install(Options{Commit: CommitDeps{Locks: locks, Hooks: hooks}})
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	if _, err := os.Stat(live); err != nil {
		t.Fatalf("the pending transaction was never recovered: %v", err)
	}
	if !recoveredBeforeInstall {
		t.Fatal("recovery did not complete before the new installation mutated its first target")
	}
}

// TestStaleInstalledGenerationRestartsInsteadOfApplyingTheOldPlan mutates an
// installed marker in the window before the manager-home lock. The operation
// must discard its staging and restart rather than commit against the
// generation it observed.
func TestStaleInstalledGenerationRestartsInsteadOfApplyingTheOldPlan(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	if result := e.install(Options{}); result.Status != "ok" {
		t.Fatalf("baseline install failed: %+v", result)
	}

	installed := filepath.Join(e.project, ".agents", "skills", "skill-a", ".csk-install.json")
	locks := &hookedLocks{inner: realLocks(t, e.home)}
	locks.beforeHome = func(attempt int) {
		// The first acquisition is journal recovery; the second is this
		// operation's commit. Move the generation just before that one.
		if attempt != 2 {
			return
		}
		payload, err := os.ReadFile(installed) // #nosec G304 -- test-owned installation
		if err != nil {
			t.Error(err)
			return
		}
		var marker map[string]any
		if err := json.Unmarshal(payload, &marker); err != nil {
			t.Error(err)
			return
		}
		marker["installed_at"] = "2020-01-01T00:00:00Z"
		mutated, err := json.MarshalIndent(marker, "", "  ")
		if err != nil {
			t.Error(err)
			return
		}
		if err := os.WriteFile(installed, append(mutated, '\n'), 0o644); err != nil {
			t.Error(err)
		}
	}

	result := e.install(Options{Commit: CommitDeps{Locks: locks}})
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	if !hasMessageContaining(result.Messages, "restarting from") {
		t.Fatalf("no restart was reported; messages = %v", result.Messages)
	}
	if !hasMessageContaining(result.Messages, "generation of marker/project/skill-a changed") {
		t.Fatalf("the restart did not name the stale generation; messages = %v", result.Messages)
	}
}

// TestMaintenanceFailureAfterCommitIsAWarning proves post-commit garbage
// collection cannot revert a durable installation.
func TestMaintenanceFailureAfterCommitIsAWarning(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")

	result := e.install(Options{Commit: CommitDeps{
		Collect: func(scopes.MaintenanceRequest) (scopes.MaintenanceResult, error) {
			return scopes.MaintenanceResult{}, fmt.Errorf("injected maintenance failure")
		},
	}})
	if result.Status != "ok" {
		t.Fatalf("a maintenance failure reverted the installation: %+v", result)
	}
	if !hasMessageContaining(result.Messages, "post-install maintenance failed") {
		t.Fatalf("the maintenance failure was not reported; messages = %v", result.Messages)
	}
	if _, err := os.Stat(filepath.Join(e.project, ".agents", "skills", "skill-a", "SKILL.md")); err != nil {
		t.Fatalf("the installation is not durable after a maintenance failure: %v", err)
	}
}

// TestNoSharedTargetChangesBeforeTheManagerHomeLock proves every persistent
// change of an installation happens inside the serialized commit phase.
func TestNoSharedTargetChangesBeforeTheManagerHomeLock(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	before := snapshotState(t, e)

	locks := &hookedLocks{inner: realLocks(t, e.home)}
	var atLock []sharedState
	locks.beforeHome = func(int) { atLock = append(atLock, snapshotState(t, e)) }

	result := e.install(Options{Commit: CommitDeps{Locks: locks}})
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	if len(atLock) < 2 {
		t.Fatalf("the operation took the home lock %d times, want recovery and commit", len(atLock))
	}
	for index, state := range atLock {
		if changed := before.diff(state); len(changed) > 0 {
			t.Fatalf("shared state changed before home-lock acquisition %d: %v", index+1, changed)
		}
	}
	if changed := before.diff(snapshotState(t, e)); len(changed) == 0 {
		t.Fatal("the installation committed nothing, so the assertion above proves nothing")
	}
}

// TestGlobalCommitCarriesNoConsumerLedger pins the one class difference between
// the two scopes: the machine-wide scope registers no runtime consumer.
func TestGlobalCommitCarriesNoConsumerLedger(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.globalDeclare("skill-a")
	userHome := t.TempDir()
	probe := &commitProbe{}

	result := Global(e.cfg, userHome, Options{Platform: "unix", Commit: CommitDeps{Hooks: probe.hooks()}})
	if result.Status != "ok" {
		t.Fatalf("global install failed: %+v", result)
	}
	for _, class := range probe.committedClasses() {
		if class == staging.ClassConsumer {
			t.Fatal("the global scope committed a consumer ledger")
		}
	}
	// Post-commit maintenance may still normalize the registry file; what the
	// global scope must never do is register a checkout in it.
	if registered := scopes.LoadConsumers(e.home); len(registered) != 0 {
		t.Fatalf("the global scope registered consumers %v", registered)
	}
}

// TestStagedPlanRejectsTwoProducersClaimingOneLivePath keeps a producer defect
// from reaching the journal as an ambiguous target set.
func TestStagedPlanRejectsTwoProducersClaimingOneLivePath(t *testing.T) {
	var plan staging.Plan
	plan.Replace(staging.ClassContext, "project/skill-a", "/live/skill-a", "/staged/one")
	plan.Replace(staging.ClassAdapterLedger, "adapter/skill-a", "/live/skill-a", "/staged/two")
	if err := plan.Validate(); err == nil {
		t.Fatal("two targets claiming one live path were accepted")
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

// newEnvIn builds a scope that shares one manager home and skill root with
// another, which is what makes the concurrency assertions meaningful.
func newEnvIn(t *testing.T, home, skillsRoot string) *env {
	t.Helper()
	e := &env{t: t, skillsRoot: skillsRoot, home: home, project: t.TempDir()}
	e.git(e.project, "init", "-q")
	e.cfg = &config.Config{
		Path:          filepath.Join(home, "config.json"),
		SkillsRoot:    skillsRoot,
		DefaultAgents: []string{"claude_code"},
		AdapterMode:   "auto",
	}
	return e
}

// TestAdapterLedgerCommitsAfterTheMirrorsItClaims pins the intra-class order
// that makes an adapter ledger meaningful: a ledger must never claim an entry
// that is not durable yet.
func TestAdapterLedgerCommitsAfterTheMirrorsItClaims(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	probe := &commitProbe{}

	if result := e.install(Options{Commit: CommitDeps{Hooks: probe.hooks()}}); result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}

	probe.mu.Lock()
	defer probe.mu.Unlock()
	entryIndex, ledgerIndex := -1, -1
	for index, event := range probe.committed {
		if event.Class != staging.ClassAdapterLedger {
			continue
		}
		if strings.Contains(event.Identifier, "/entry/") {
			entryIndex = index
		}
		if strings.HasSuffix(event.Identifier, "/ledger") {
			ledgerIndex = index
		}
	}
	if entryIndex < 0 || ledgerIndex < 0 {
		t.Fatalf("the adapter class committed no entry and ledger pair: %+v", probe.committed)
	}
	if entryIndex > ledgerIndex {
		t.Fatal("the adapter ledger committed before the mirror entry it claims")
	}
	if _, err := os.Stat(filepath.Join(e.project, ".claude", "skills", adapters.LedgerName)); err != nil {
		t.Fatalf("the adapter ledger is not durable: %v", err)
	}
}
