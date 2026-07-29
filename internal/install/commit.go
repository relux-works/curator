package install

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/transaction"
)

// HomeLock is the exclusive manager-home mutation witness. It is held from
// revalidation through commit, rollback, and post-commit maintenance, so no
// other project can publish a cache entry or swap a shared target in between.
type HomeLock interface {
	AssertHeld() error
	Close() error
}

// ProjectLocks is the canonical project operation lock set of one run. It is
// acquired before planning and released only after the whole handoff.
type ProjectLocks interface {
	Close() error
}

// LockBroker is the narrow locking boundary of a real installation.
//
// AcquireHome is deliberately allowed more than once per run: an operation
// takes the home lock briefly for journal recovery, releases it before any
// network, audit, fingerprinting, or compilation work, and takes it again for
// the serialized commit. A restart adds further acquisitions.
type LockBroker interface {
	AcquireProjects(ctx context.Context, projects ...string) (ProjectLocks, error)
	AcquireHome(ctx context.Context) (HomeLock, error)
}

// TargetJournal is the durable transaction boundary that owns every live swap.
type TargetJournal interface {
	Recover(lock transaction.HomeLock) error
	Prepare(lock transaction.HomeLock, plan transaction.Plan) (*transaction.Journal, error)
	Commit(lock transaction.HomeLock, transactionID string) error
	// ReferencedBuildKeys names every build key a still-live journal depends
	// on, so post-commit maintenance retains in-flight artifacts.
	ReferencedBuildKeys(lock transaction.HomeLock) ([]string, error)
}

// CachePublisher publishes verified private builds into the protected
// immutable build cache and re-reads the result. It never adopts or repairs
// foreign bytes.
//
// Publication and verification are deliberately the same boundary. The planning
// inspector runs before the home lock against whatever the cache looked like
// then; only the store that just published can say whether the entry a launcher
// is about to point at is protected right now.
type CachePublisher interface {
	Inspect(expect buildcache.Expectation) buildcache.Result
	// Publish compensates its own mutations. An error means the protected cache
	// root is what the call found, so this phase records nothing to put back for
	// it — unless the error satisfies buildcache.StateChanged, which says the
	// compensation itself did not complete.
	Publish(publication buildcache.Publication, lock buildcache.HomeLock) (buildcache.PublicationResult, error)
	// Revert puts back what one publication of this run displaced. It is the
	// compensating half of Publish: the protected entry a launcher points at has
	// to be live before the targets that reference it can be staged, so a run
	// that fails in between owns restoring the shared cache itself. Every error
	// it returns means the cache is not back at what the run found.
	Revert(key buildmeta.CacheKey, published buildcache.PublicationResult, lock buildcache.HomeLock) error
}

// CommitDeps injects the serialized publication and commit boundaries. The
// zero value resolves the real manager locks, transaction engine, protected
// cache publisher, and runtime collector.
type CommitDeps struct {
	Locks     LockBroker
	Journal   TargetJournal
	Publisher CachePublisher
	// Collect runs the post-commit maintenance sweep under the still-held home
	// lock. A failure is a reported warning: the installation is already
	// durable, and retaining an unreferenced entry is always safe.
	Collect func(request scopes.MaintenanceRequest) (scopes.MaintenanceResult, error)
	// Hooks inject faults at transaction boundaries so a test can prove that a
	// failure at any target class restores the prior state in reverse order.
	Hooks transaction.Hooks
	// PublishFault injects a failure at cache publication. It obeys the same
	// contract as the store: a plain error models a publication that put the
	// cache back itself, and a buildcache.StateChangedError models one that
	// could not.
	PublishFault func(buildmeta.CacheKey) error
	// NewTransactionID names one journal; the default is random.
	NewTransactionID func() (string, error)
	// MaxRestarts bounds the revalidation restart loop. Zero means the default.
	MaxRestarts int
}

// defaultMaxRestarts bounds how often concurrent shared-state changes may push
// one operation back to an earlier planning step before it reports the churn
// instead of looping. Restarting is always preferable to applying a stale plan,
// but an unbounded loop would hide a livelock.
const defaultMaxRestarts = 3

// resolve fills every unset boundary with its real implementation.
//
// inspector is the protected-cache boundary planning already read through. When
// it can publish — the real store can — it becomes the publisher too, so the
// commit phase publishes and re-reads through exactly one cache boundary
// instead of two that could disagree about the same home.
func (deps CommitDeps) resolve(home string, inspector CacheInspector) (CommitDeps, error) {
	if deps.Locks == nil {
		manager, err := managerlock.New(home)
		if err != nil {
			return CommitDeps{}, fmt.Errorf("open manager locks: %w", err)
		}
		deps.Locks = &managerLocks{manager: manager}
	}
	if deps.Journal == nil {
		engine, err := transaction.New(home, transaction.WithHooks(deps.Hooks))
		if err != nil {
			return CommitDeps{}, fmt.Errorf("open the install transaction journal: %w", err)
		}
		deps.Journal = engine
	}
	if deps.Publisher == nil {
		if publisher, capable := inspector.(CachePublisher); capable && publisher != nil {
			deps.Publisher = publisher
		} else {
			store, err := buildcache.New(home)
			if err != nil {
				return CommitDeps{}, fmt.Errorf("open protected build cache: %w", err)
			}
			deps.Publisher = store
		}
	}
	if deps.Collect == nil {
		deps.Collect = scopes.Collect
	}
	if deps.NewTransactionID == nil {
		deps.NewTransactionID = randomTransactionID
	}
	if deps.MaxRestarts <= 0 {
		deps.MaxRestarts = defaultMaxRestarts
	}
	return deps, nil
}

// managerLocks adapts the cross-platform manager locks to the narrow broker.
//
// Project locks come from one ordered operation that lives for the whole run.
// Home locks go through the manager's recovery-path acquisition instead of the
// operation, because an operation may hold the home lock only once and this
// lifecycle needs it for recovery, for the commit, and again after a restart.
// The class ordering rule still holds: the project locks are already held, and
// no project or key lock is ever taken while the home lock is held.
type managerLocks struct {
	manager   *managerlock.Manager
	operation *managerlock.Operation
}

func (locks *managerLocks) AcquireProjects(ctx context.Context, projects ...string) (ProjectLocks, error) {
	operation := locks.manager.NewOperation(false)
	if _, err := operation.AcquireProjects(ctx, projects...); err != nil {
		return nil, errors.Join(err, operation.Close())
	}
	locks.operation = operation
	return operation, nil
}

func (locks *managerLocks) AcquireHome(ctx context.Context) (HomeLock, error) {
	return locks.manager.AcquireHomeOnly(ctx, false)
}

func randomTransactionID() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate transaction id: %w", err)
	}
	return "install-" + hex.EncodeToString(buffer), nil
}

// restartStep names the earliest phase a stale observation invalidates. Larger
// values are earlier phases, so combining reasons takes the maximum.
type restartStep int

const (
	restartNone restartStep = iota
	// restartTargets re-derives the staged target set: a target owner or
	// expected preimage moved under us.
	restartTargets
	// restartPlan re-derives the build plan: cache trust or a required build
	// key changed, so the planned reuse verdict no longer holds.
	restartPlan
	// restartClosure re-resolves the closure: an installed generation changed,
	// so closure or activation input may differ.
	restartClosure
)

func (step restartStep) String() string {
	switch step {
	case restartTargets:
		return "target staging"
	case restartPlan:
		return "build planning"
	case restartClosure:
		return "closure resolution"
	default:
		return "none"
	}
}

// restartError reports that shared state changed while the operation held no
// home lock. The staged plan is discarded rather than applied.
type restartError struct {
	step    restartStep
	reasons []string
}

func (err *restartError) Error() string {
	return fmt.Sprintf("shared state changed; restarting from %s: %s", err.step, strings.Join(err.reasons, "; "))
}

// buildPhaseError marks a commit-phase failure that originated at the protected
// build cache boundary.
//
// Publication, receipt encoding, and reversal all carry filesystem-, toolchain-,
// or manager-private detail — cache and staging locations, receipt bytes, driver
// text — exactly like the planning and staging phases do. The marker is what
// lets a scope route those failures through the one bounded, path-redacted
// rendering while an ordinary commit failure keeps naming the operator's own
// paths, which are the actionable part of a journal or target error.
type buildPhaseError struct{ err error }

func (failure *buildPhaseError) Error() string { return failure.err.Error() }

func (failure *buildPhaseError) Unwrap() error { return failure.err }

// failedBuildBoundary marks one protected build-cache failure of the commit
// phase, so the scope renders it through the bounded, redacted path instead of
// printing manager-private cache and artifact locations verbatim.
func failedBuildBoundary(format string, args ...any) error {
	return &buildPhaseError{err: fmt.Errorf(format, args...)}
}

// documentKey is the observation key of a declaration document: a file whose
// bytes select installed content. Its own type is the structural half of the
// binding invariant — a document can only enter the observation set through
// observeDocument, which takes the generation of the bytes a parser consumed,
// and never through observe, which digests a path as a separate operation and
// would reopen the A -> B -> A window (see generation.go).
type documentKey string

// The optimistic observation keys of the declaration inputs a closure is
// resolved from. Each names a file that decides *what* an operation installs and
// that some supported Curator command rewrites while holding none of this run's
// locks, so a change to any of them restarts closure resolution rather than
// committing the closure that was planned.
const (
	// hybridActivationKey is the machine-home hybrid activation manifest,
	// rewritten by `curator hybrid add|rm`.
	hybridActivationKey documentKey = "activation/hybrid-manifest"
	// projectManifestKey is the project Skillfile, rewritten by
	// `curator add|remove`.
	projectManifestKey documentKey = "manifest/project"
	// globalManifestKey is the machine-wide Skillfile under GlobalRoot,
	// rewritten by `curator global add|remove`.
	globalManifestKey documentKey = "manifest/global"
	// substitutionsKey is the project development substitution manifest. No
	// Curator command writes it — it is edited by hand — but it redirects a
	// declaration at a local checkout or another ref, so it selects installed
	// content exactly like the manifest does.
	substitutionsKey documentKey = "substitutions/project"
)

// observations are the optimistic generation digests captured while the
// operation held no home lock (Spec §6.1 step 6). They are not commit
// authority: every one is re-read under the home lock in step 10, and a
// difference restarts the earliest affected step instead of applying the plan.
//
// The set has to cover every closure or activation input a phase read outside
// the home lock and could not otherwise prove stable. Each input below is
// classified by who can actually write it, not by which directory it sits in:
// the canonical project operation lock is a witness only against writers that
// take it, and the manifest-editing commands do not.
//
// The four declaration documents are observed at the generation of the exact
// bytes their parser consumed, taken by the single read in readDocument, so no
// separate digest of the path can disagree with the closure that was built.
// Anything else is observed by its path (see observe and observeDocument).
//
//   - The project manifest is observed (projectManifestKey). It lives in the
//     checkout, but `curator add|remove` rewrites it through manifest.AddDecl
//     and manifest.RemoveDecl without taking any operation lock, so being in
//     the locked checkout proves nothing about it.
//   - The machine-wide manifest is observed (globalManifestKey). Same reason:
//     it lives under GlobalRoot(home), which is the global scope's own
//     operation identity, but `curator global add|remove` writes it lock-free.
//   - The project development substitution manifest is observed
//     (substitutionsKey). It has no Curator writer at all — it is edited by
//     hand — and it redirects a declaration at a local checkout or another ref,
//     so it selects installed content and also drives the strict-audit refusal.
//   - The machine-home hybrid activation manifest is observed
//     (hybridActivationKey): `curator hybrid add|rm` rewrites it under no lock
//     at all, so it can move at any moment during an install.
//   - Installed markers of every closure node, in the project, hybrid, and
//     global stores, are observed one key each.
//   - Protected build-cache verdicts are observed per planned build as
//     outcomes, and re-derived through the publishing store itself.
//   - Every staged target's preimage is digested under the home lock inside
//     journalPlan, so target owners and preimages need no optimistic key.
//   - The managed .gitignore is deliberately *not* observed, and no lock is
//     claimed over it either: `curator init` appends to it and this operation
//     rewrites it itself when --fix-gitignore is set. It is a hygiene
//     precondition rather than a closure input — its content selects nothing
//     that gets installed — so a concurrent edit cannot make this run commit
//     the wrong closure, and the next run re-runs the gate. Observing a file
//     the run itself writes would only manufacture restart loops.
//   - Skill snapshots are commit-addressed and written once, so the tree the
//     closure resolved cannot change underneath the run. A tag that moves after
//     resolution leaves this operation committing the commit it pinned, which
//     is a complete and self-consistent installation; the next run resolves the
//     new commit and the moved-tag gate reports it. This covers a development
//     substitution too: a local checkout is resolved to a concrete HEAD commit
//     and snapshotted under that commit like any other source, so only the
//     substitution manifest above — which one it points at — needs observing.
//   - Registry attestations and MCP verification results are resolved before
//     the home lock and recorded into the marker this run commits. They select
//     nothing: the closure is already resolved when they run, and each is
//     evidence of what this operation found. A registry or agent configuration
//     that changes afterwards leaves this run committing the evidence it
//     pinned, exactly like a snapshot, and the next run records the new state.
//   - The audit trust state under the manager home is a gate on the resolved
//     closure, not an input to it: it can refuse an installation but cannot
//     change which skills it contains. No lock is claimed over it — `curator
//     audit pin` writes an approval outside any operation — and an approval
//     revoked after the gate ran leaves a complete installation the next run's
//     gate refuses. The gate memoizes its own verdicts beside that state; that
//     cache is derived from content hashes and is not an installation target.
//   - The user configuration is loaded once by the process entrypoint and
//     passed through as one immutable value, so every phase of the run — and
//     every restart attempt — reads the same snapshot and the run is internally
//     consistent. Observing it would not help: a restart re-reads no
//     configuration, so a concurrent edit could only turn into a restart loop.
//     Picking up a mid-run configuration edit needs per-attempt reloading,
//     which is a separate decision from this revalidation contract.
type observations struct {
	// entries maps a canonical shared-state key to the generation observed for
	// it, the path that revalidates it, and the reader that produced it. One
	// record holds all three, so an observation cannot exist without the location
	// and the reader its recheck needs.
	entries map[string]observation
	// outcomes maps one planned build to the protected-cache verdict planning
	// took for it.
	outcomes map[string]BuildOutcome
}

// observation is one optimistic generation and everything needed to revalidate
// it under the home lock.
type observation struct {
	// generation is what the reader below saw when the input was consulted.
	generation string
	// path is the location to re-read under the home lock.
	path string
	// bytesBound marks a declaration document, whose generation covers the exact
	// bytes a parser consumed rather than whatever the path held at some separate
	// moment. It selects the recheck reader too: an observation must always be
	// rechecked by the reader that recorded it, or two readers that disagree
	// about a path would restart every attempt forever.
	bytesBound bool
}

// current returns the generation of this observation's path now, read by the
// same reader that recorded it.
func (entry observation) current() string {
	if entry.bytesBound {
		return documentGeneration(entry.path)
	}
	digest, err := transaction.DigestPath(entry.path)
	if err != nil {
		// An unreadable or unsafe preimage is itself a stable observation: the
		// recheck reproduces the same marker and only a change restarts.
		return "unreadable:" + err.Error()
	}
	return digest
}

func newObservations() *observations {
	return &observations{
		entries:  map[string]observation{},
		outcomes: map[string]BuildOutcome{},
	}
}

// observe records the current digest of one consulted shared path. It is for
// state whose *path* is the observation — an installed marker — and deliberately
// cannot take a documentKey: a declaration document read separately from its
// digest is the ABA defect this contract exists to prevent.
func (observed *observations) observe(key, path string) {
	if observed == nil {
		return
	}
	entry := observation{path: path}
	entry.generation = entry.current()
	observed.entries[key] = entry
}

// observeDocument records one declaration document at the generation of the
// exact bytes its parser consumed, as returned by the single read in
// readDocument. The caller passes that generation rather than a path, so no
// second read can slip between the bytes the closure was built from and the
// generation this run revalidates.
func (observed *observations) observeDocument(key documentKey, path, generation string) {
	if observed == nil {
		return
	}
	observed.entries[string(key)] = observation{generation: generation, path: path, bytesBound: true}
}

// recheck re-reads every observation and reports the earliest affected step.
func (observed *observations) recheck(plan BuildPlan, inspect CacheInspector) *restartError {
	if observed == nil {
		return nil
	}
	step := restartNone
	var reasons []string
	keys := make([]string, 0, len(observed.entries))
	for key := range observed.entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		entry := observed.entries[key]
		if entry.current() != entry.generation {
			step = maxRestart(step, restartClosure)
			reasons = append(reasons, fmt.Sprintf("generation of %s changed", key))
		}
	}
	for _, build := range plan.builds {
		recorded, seen := observed.outcomes[build.skill+"."+build.command]
		if !seen {
			continue
		}
		current := BuildOutcome(inspect.Inspect(buildcache.Expectation{Input: build.input}).DryRunOutcome())
		// A miss that became a hit is another publisher's identical winner and
		// is safe to adopt. Any other transition invalidates the planned reuse.
		if current == recorded || (recorded.buildable() && current == BuildCacheHit) {
			continue
		}
		step = maxRestart(step, restartPlan)
		reasons = append(reasons, fmt.Sprintf("cache trust for %s.%s changed from %s to %s",
			build.skill, build.command, recorded, current))
	}
	if step == restartNone {
		return nil
	}
	return &restartError{step: step, reasons: reasons}
}

func maxRestart(left, right restartStep) restartStep {
	if right > left {
		return right
	}
	return left
}

// commitRequest is everything the serialized publication and commit phase needs
// from one scope.
type commitRequest struct {
	scope string
	home  string
	// projectRoot is the checkout this run registers as a runtime consumer.
	// The global scope leaves it empty and writes no consumer ledger.
	projectRoot string
	deps        BuildDeps
	commit      CommitDeps
	plan        BuildPlan
	staged      Staged
	observed    *observations
	// stageTargets derives the complete staged target set of the scope. It runs
	// under the home lock, after recovery, revalidation, and cache publication,
	// so every shared ledger it merges is read from committed state.
	stageTargets func(scopeCommit) (scopeTargets, error)
}

// scopeCommit is the home-locked context a scope stages its targets in.
type scopeCommit struct {
	stageRoot string
	// artifacts maps every planned build key to its protected cache entry.
	artifacts map[buildmeta.CacheKey]buildcache.Result
}

// scopeTargets is the staged result of one scope. Every live change of the
// scope is in the plan: there is deliberately no side channel for state the
// journal cannot own, so nothing mutates between preparation and commit and an
// ordinary failure always enters durable rollback.
type scopeTargets struct {
	plan     staging.Plan
	messages []string
	// referencedKeys are the build keys the committed markers depend on, so GC
	// retains them while this transaction is in flight.
	referencedKeys []string
}

// commitOutcome reports what the serialized phase changed.
type commitOutcome struct {
	messages []string
	warnings []string
	// retainedBuilds marks a failed run that could not put the shared build
	// cache back, so a presentation layer never tells an operator the live cache
	// is unchanged when it is not.
	retainedBuilds bool
}

// recoverJournals takes the home lock briefly, resumes every incomplete
// transaction, and releases it again before any network, audit, fingerprinting,
// or compilation work (Spec §6.1 step 1).
func recoverJournals(ctx context.Context, deps CommitDeps) error {
	lock, err := deps.Locks.AcquireHome(ctx)
	if err != nil {
		return fmt.Errorf("acquire the manager-home lock for recovery: %w", err)
	}
	recoverErr := deps.Journal.Recover(lock)
	closeErr := lock.Close()
	if recoverErr != nil {
		return fmt.Errorf("recover incomplete install transactions: %w", errors.Join(recoverErr, closeErr))
	}
	if closeErr != nil {
		return fmt.Errorf("release the manager-home lock after recovery: %w", closeErr)
	}
	return nil
}

// runCommit is the serialized publication and commit phase (Spec §6.1 steps
// 10-14). It holds the manager-home mutation lock from revalidation through
// rollback and post-commit maintenance.
func runCommit(ctx context.Context, request commitRequest) (outcome commitOutcome, err error) {
	lock, err := request.commit.Locks.AcquireHome(ctx)
	if err != nil {
		return commitOutcome{}, fmt.Errorf("acquire the manager-home lock: %w", err)
	}
	defer func() {
		if closeErr := lock.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("release the manager-home lock: %w", closeErr)
		}
	}()

	// Recovery completes before this operation mutates anything.
	if err := request.commit.Journal.Recover(lock); err != nil {
		return commitOutcome{}, fmt.Errorf("recover incomplete install transactions: %w", err)
	}

	// Revalidate every optimistic observation. A stale closure, activation,
	// cache trust, target owner, preimage, or required key restarts.
	if restart := request.observed.recheck(request.plan, request.deps.Cache); restart != nil {
		return commitOutcome{}, restart
	}

	stageRoot, err := os.MkdirTemp("", operationPrivatePrefix+"commit-")
	if err != nil {
		return commitOutcome{}, fmt.Errorf("create the commit staging root: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(stageRoot); removeErr != nil {
			outcome.warnings = append(outcome.warnings,
				fmt.Sprintf("%s: could not remove the commit staging root: %v", request.scope, removeErr))
		}
	}()

	builds, err := publishWinners(request, lock)
	// journaled is the durable plan this run recorded preimages for, and stays
	// empty until preparation has something to compare against.
	var journaled transaction.Plan
	// A failed run puts the shared build cache back before it releases the home
	// lock, so an operation that committed nothing also displaced nothing. The
	// reversal is deliberately conditional: see reversiblePublication.
	//
	// Every way that can fail ends in the same place — retainedBuilds — because
	// the presentation layer has exactly one decision to make: whether it may
	// still tell the operator the live build cache is unchanged.
	defer func() {
		if err == nil {
			return
		}
		if builds.retained {
			outcome.retainedBuilds = true
			outcome.warnings = append(outcome.warnings, fmt.Sprintf(
				"%s: warning: the rebuilt build cache entries of this run were kept: "+
					"a failed publication could not put the protected cache back", request.scope))
		}
		if len(builds.entries) == 0 {
			return
		}
		reverted, reason := reversiblePublication(request, journaled, builds.entries, lock)
		if !reverted {
			outcome.retainedBuilds = true
			outcome.warnings = append(outcome.warnings, fmt.Sprintf(
				"%s: warning: the rebuilt build cache entries of this run were kept: %s", request.scope, reason))
			return
		}
		if revertErr := revertPublications(request, builds.entries, lock); revertErr != nil {
			outcome.retainedBuilds = true
			outcome.warnings = append(outcome.warnings, fmt.Sprintf(
				"%s: warning: the rebuilt build cache entries of this run were kept: "+
					"the reversal did not complete", request.scope))
			err = errors.Join(err, revertErr)
		}
	}()
	if err != nil {
		return commitOutcome{}, err
	}
	artifacts := builds.artifacts
	outcome.messages = append(outcome.messages, builds.messages...)

	targets, err := request.stageTargets(scopeCommit{stageRoot: stageRoot, artifacts: artifacts})
	if err != nil {
		return commitOutcome{}, err
	}
	outcome.messages = append(outcome.messages, targets.messages...)

	if request.projectRoot != "" {
		consumer, err := scopes.StageConsumer(stageRoot, request.home, request.projectRoot)
		if err != nil {
			return commitOutcome{}, err
		}
		targets.plan.Merge(consumer)
	}
	if err := targets.plan.Validate(); err != nil {
		return commitOutcome{}, err
	}
	if targets.plan.Empty() {
		outcome.warnings = append(outcome.warnings, collectAfterCommit(request, lock)...)
		return outcome, nil
	}

	// A target below a directory that does not exist yet needs that directory
	// before the journal can place its sidecars. Those directories are not
	// journaled targets, so an abandoned commit has to unwind them itself;
	// otherwise a rolled-back installation would leave an empty runtime or
	// skills tree behind and the restored state would not be the prior one.
	journalPlan, created, planErr := journalPlan(request, targets)
	journaled = journalPlan
	committed := false
	defer func() {
		if !committed {
			removeCreatedDirectories(created)
		}
	}()
	if planErr != nil {
		return commitOutcome{}, planErr
	}
	// Preparation is immediately followed by the commit. Nothing may run in
	// between: a prepared journal is a transaction recovery will finish, so an
	// operation that returned failed must never leave one behind.
	if _, err := request.commit.Journal.Prepare(lock, journalPlan); err != nil {
		return commitOutcome{}, fmt.Errorf("prepare the install transaction: %w", err)
	}
	if err := request.commit.Journal.Commit(lock, journalPlan.TransactionID); err != nil {
		return commitOutcome{}, fmt.Errorf("commit the install transaction: %w", err)
	}
	committed = true
	outcome.warnings = append(outcome.warnings, collectAfterCommit(request, lock)...)
	return outcome, nil
}

// collectAfterCommit runs the maintenance sweep while the home lock is still
// held, so consumer pruning, runtime collection, and the protected build-cache
// sweep serialize with every other manager-home mutation. A failure is reported
// and never reverts the durable installation (Spec §6.1 step 14).
func collectAfterCommit(request commitRequest, lock HomeLock) []string {
	if request.commit.Collect == nil {
		return nil
	}
	warn := func(format string, args ...any) []string {
		return []string{fmt.Sprintf("%s: warning: %s", request.scope, fmt.Sprintf(format, args...))}
	}
	journalKeys, err := request.commit.Journal.ReferencedBuildKeys(lock)
	if err != nil {
		return warn("post-install maintenance failed: read in-flight build references: %v", err)
	}
	result, err := request.commit.Collect(scopes.MaintenanceRequest{
		Home:        request.home,
		Lock:        lock,
		JournalKeys: journalKeys,
	})
	var warnings []string
	for _, warning := range result.Warnings {
		warnings = append(warnings, fmt.Sprintf("%s: warning: %s", request.scope, warning))
	}
	if err != nil {
		warnings = append(warnings, warn("post-install maintenance failed: %v", err)...)
	}
	return warnings
}

// publication is one cache entry this run selected, and everything needed to
// put back what it displaced.
type publication struct {
	skill     string
	command   string
	key       buildmeta.CacheKey
	published buildcache.PublicationResult
}

// publishedBuilds is everything the publication phase produced, including what
// it left behind on a failing path.
type publishedBuilds struct {
	artifacts map[buildmeta.CacheKey]buildcache.Result
	// entries are the publications this run selected and therefore owns putting
	// back. They are reported on every path, including a failing one: a run that
	// displaced a cache entry and then failed still owes the restoration.
	entries  []publication
	messages []string
	// retained marks a publication that failed and could not put the protected
	// cache root back by itself. There is nothing here for the caller to revert
	// — the store compensates its own mutations — but the live cache is no
	// longer what this run found, and reporting it as unchanged would be false.
	retained bool
}

// publishWinners publishes every verified staged build and re-reads the
// protected entry of every planned command. A published duplicate of an
// identical existing winner is discarded rather than merged, and a key that
// still does not resolve after publication restarts build planning.
func publishWinners(request commitRequest, lock HomeLock) (publishedBuilds, error) {
	var builds publishedBuilds
	for _, build := range request.staged.builds {
		if request.commit.PublishFault != nil {
			if err := request.commit.PublishFault(build.key); err != nil {
				builds.retained = buildcache.StateChanged(err)
				return builds, failedBuildBoundary("publish %s.%s: %w", build.skill, build.command, err)
			}
		}
		receiptBytes, err := build.receipt.CanonicalBytes()
		if err != nil {
			return builds, failedBuildBoundary(
				"encode the receipt of %s.%s: %w", build.skill, build.command, err)
		}
		result, err := request.commit.Publisher.Publish(buildcache.Publication{
			Input:          build.receipt.Input,
			ReceiptBytes:   receiptBytes,
			ArtifactSource: build.path,
		}, lock)
		if err != nil {
			// A publication that failed without leaving the cache changed owes
			// nothing and is not recorded: adding it would make the caller revert
			// a mutation the store already undid.
			builds.retained = buildcache.StateChanged(err)
			return builds, failedBuildBoundary("publish %s.%s: %w", build.skill, build.command, err)
		}
		builds.entries = append(builds.entries, publication{
			skill: build.skill, command: build.command, key: build.key, published: result,
		})
		builds.messages = append(builds.messages, fmt.Sprintf("%s: %s.%s cache %s key=%s",
			request.scope, build.skill, build.command, result.Status, build.key))
	}

	artifacts := map[buildmeta.CacheKey]buildcache.Result{}
	var missing []string
	for _, build := range request.plan.builds {
		inspection := request.commit.Publisher.Inspect(buildcache.Expectation{Input: build.input})
		if inspection.Status != buildcache.Hit {
			missing = append(missing, fmt.Sprintf("%s.%s (%s)", build.skill, build.command, inspection.Status))
			continue
		}
		artifacts[build.key] = inspection
	}
	if len(missing) > 0 {
		builds.messages = nil
		return builds, &restartError{
			step:    restartPlan,
			reasons: []string{"required build keys are not protected after publication: " + strings.Join(missing, ", ")},
		}
	}
	builds.artifacts = artifacts
	return builds, nil
}

// reversiblePublication decides whether a failed run may put the shared build
// cache back, and says why when it may not.
//
// Two independent facts have to hold. Every target this run journaled has to be
// back at the exact preimage it recorded under the home lock, so the
// installation really did return to its prior state; and no journal may still
// reference the build keys this run published, so no durable transaction is
// left whose completion recovery — not this process — still owns.
//
// Either one unmet keeps the published entries. That asymmetry is deliberate:
// a valid entry no installation references is always safe and the ordinary
// sweep collects it, while restoring an unusable predecessor over the entry a
// recovered commit is about to point at would turn a reported failure into a
// broken installation.
func reversiblePublication(
	request commitRequest,
	journaled transaction.Plan,
	published []publication,
	lock HomeLock,
) (bool, string) {
	for _, target := range journaled.Targets {
		digest, err := transaction.DigestTarget(target.Kind, target.LivePath)
		if err != nil {
			return false, fmt.Sprintf("%s/%s could not be read back", target.Class, target.Identifier)
		}
		if digest != target.PreimageDigest {
			return false, fmt.Sprintf("%s/%s is no longer at the state this run found", target.Class, target.Identifier)
		}
	}
	inflight, err := request.commit.Journal.ReferencedBuildKeys(lock)
	if err != nil {
		return false, "in-flight build references could not be read"
	}
	referenced := map[string]bool{}
	for _, key := range inflight {
		referenced[key] = true
	}
	for _, entry := range published {
		if referenced[string(entry.key)] {
			return false, fmt.Sprintf("an incomplete transaction still references %s.%s", entry.skill, entry.command)
		}
	}
	return true, ""
}

// revertPublications puts every entry this run selected back, newest first, so
// the cache passes through the same states in reverse order.
func revertPublications(request commitRequest, published []publication, lock HomeLock) error {
	var errs []error
	for index := len(published) - 1; index >= 0; index-- {
		entry := published[index]
		if err := request.commit.Publisher.Revert(entry.key, entry.published, lock); err != nil {
			errs = append(errs, failedBuildBoundary(
				"restore the build cache entry of %s.%s: %w", entry.skill, entry.command, err))
		}
	}
	return errors.Join(errs...)
}

// journalPlan converts the staged target set into one durable transaction. The
// preimage digest of every target is read here, under the home lock, so no
// other process can move a target between the expectation and the swap.
// It returns the parent directories it had to create, deepest last, so an
// abandoned commit can unwind exactly what it added.
func journalPlan(request commitRequest, targets scopeTargets) (transaction.Plan, []string, error) {
	var created []string
	id, err := request.commit.NewTransactionID()
	if err != nil {
		return transaction.Plan{}, created, err
	}
	identity := request.projectRoot
	if identity == "" {
		identity = "global:" + request.home
	}
	plan := transaction.Plan{
		TransactionID:       id,
		ProjectIdentity:     identity,
		ReferencedBuildKeys: uniqueSorted(targets.referencedKeys),
	}
	for _, target := range targets.Sorted() {
		kind, err := journalKind(target.Kind)
		if err != nil {
			return transaction.Plan{}, created, fmt.Errorf("stage %s/%s: %w", target.Class, target.Identifier, err)
		}
		if target.StagedPath != "" {
			added, err := makeMissingDirectories(filepath.Dir(target.LivePath))
			if err != nil {
				return transaction.Plan{}, created, fmt.Errorf("prepare the parent of %s: %w", target.LivePath, err)
			}
			created = append(created, added...)
		}
		preimage, err := transaction.DigestTarget(kind, target.LivePath)
		if err != nil {
			return transaction.Plan{}, created, fmt.Errorf("read the current state of %s: %w", target.LivePath, err)
		}
		plan.Targets = append(plan.Targets, transaction.Target{
			Class:          target.Class,
			Identifier:     target.Identifier,
			Kind:           kind,
			LivePath:       target.LivePath,
			StagedSource:   target.StagedPath,
			PreimageDigest: preimage,
		})
	}
	return plan, created, nil
}

// journalKind maps a staged target kind onto the durable transaction kind. The
// two vocabularies are deliberately separate: internal/staging stays
// dependency-free, so the translation is explicit rather than assumed.
func journalKind(kind string) (transaction.TargetKind, error) {
	switch kind {
	case staging.KindBytes:
		return transaction.KindBytes, nil
	case staging.KindEntry:
		return transaction.KindEntry, nil
	default:
		return "", fmt.Errorf("unknown staged target kind %q", kind)
	}
}

// makeMissingDirectories creates path and every missing ancestor, returning the
// ones it actually created, shallowest first.
func makeMissingDirectories(path string) ([]string, error) {
	if _, err := os.Lstat(path); err == nil {
		return nil, nil
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	parent := filepath.Dir(path)
	var created []string
	if parent != path {
		added, err := makeMissingDirectories(parent)
		created = append(created, added...)
		if err != nil {
			return created, err
		}
	}
	if err := os.Mkdir(path, 0o755); err != nil {
		if os.IsExist(err) {
			return created, nil
		}
		return created, err
	}
	return append(created, path), nil
}

// removeCreatedDirectories drops the directories a commit added, deepest first.
// A directory that is no longer empty belongs to someone else by now and is
// left alone: unwinding our own scaffolding must never delete live state.
func removeCreatedDirectories(created []string) {
	for index := len(created) - 1; index >= 0; index-- {
		_ = os.Remove(created[index])
	}
}

// Sorted exposes the staged targets in commit order for the journal.
func (targets scopeTargets) Sorted() []staging.Target { return targets.plan.Sorted() }

func uniqueSorted(values []string) []string {
	seen := map[string]bool{}
	var unique []string
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	sort.Strings(unique)
	return unique
}
