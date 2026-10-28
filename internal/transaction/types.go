// Package transaction provides durable, recoverable replacement transactions
// for manager-owned install targets. Callers stage private build/runtime output
// before entering this package, then hold the manager-home lock while preparing,
// committing, recovering, or enumerating journal references.
package transaction

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/relux-works/curator/internal/envjournal"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/staging"
)

const (
	// PointBeforeStagingFinalMode observes a private writable sidecar before final protection.
	PointBeforeStagingFinalMode Point = "before_staging_final_mode"
	// PointAfterStagingFinalMode observes final protection with the original writer still open.
	PointAfterStagingFinalMode Point = "after_staging_final_mode"
	journalSchema                    = "curator-install-transaction-v1"
	// DigestAbsent is the canonical digest token for an absent target.
	DigestAbsent = "absent"
)

var (
	// ErrImplementationCorruption reports filesystem state that cannot be a
	// valid transition of the durable journal. Unknown bytes are preserved.
	ErrImplementationCorruption = errors.New("transaction implementation-corruption")
	errInvalidJournal           = errors.New("invalid transaction journal")
	transactionIDPattern        = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)
)

// HomeLock is the caller-owned witness for the exclusive manager-home lock.
// Implementations must fail AssertHeld after release.
type HomeLock interface {
	AssertHeld() error
}

// Journal wire identities are shared with the independent admission reader.
type Phase = envjournal.Phase
type CommitState = envjournal.CommitState
type TargetKind = envjournal.TargetKind

const (
	PhasePreparing       = envjournal.PhasePreparing
	PhasePrepared        = envjournal.PhasePrepared
	PhaseCommitting      = envjournal.PhaseCommitting
	PhaseCleanup         = envjournal.PhaseCleanup
	PhaseRollingBack     = envjournal.PhaseRollingBack
	PhaseRollbackCleanup = envjournal.PhaseRollbackCleanup
	StatePending         = envjournal.StatePending
	StateBackedUp        = envjournal.StateBackedUp
	StateCommitted       = envjournal.StateCommitted
	StateRolledBack      = envjournal.StateRolledBack
	KindBytes            = envjournal.KindBytes
	KindEntry            = envjournal.KindEntry
	KindExchange         = envjournal.KindExchange
)

// Target describes one desired replacement. Class and Identifier are the
// canonical ordering key. An empty StagedSource means the desired state is
// absence. Exactly one preimage expectation is required: PreimageDigest, or
// ExpectedGeneration read from GenerationPath (relative to LivePath, or the
// live file itself when empty). A KindEntry target must use PreimageDigest,
// because a link carries no generation file. A KindExchange target requires
// a present PreimageDigest and present StagedSource, with no generation fields.
// LinkFallback is an operation-private alternative for a noncredential
// projection. Source must be a pinned, link-free prepared file/tree. Marker
// fields bind a home copy reason to its exact declared surface; legacy adapter
// mirrors have no home marker. No alternate or callback is needed after Prepare.
type LinkFallback struct {
	Source       string
	Class        string
	MarkerPath   string
	Surface      string
	RelativePath string
}

type Target struct {
	LinkFallback *LinkFallback
	// PrivateFile requires a regular, owner-only replacement, created privately
	// and revalidated before publication, including after process restart.
	PrivateTree        bool
	PrivateFile        bool
	Class              string
	Identifier         string
	Kind               TargetKind
	LivePath           string
	StagedSource       string
	PreimageDigest     string
	ExpectedGeneration string
	GenerationPath     string
}

// Plan binds one project identity, ordered target set, and referenced builds.
type Plan struct {
	// Legacy is joined by the global profile owner before Prepare; it is not
	// authority for an independent legacy journal.
	Legacy    *envunit.LegacyPublication
	Opaque    []OpaqueTarget
	Lifecycle *envlifecycle.Change
	// Unit binds a complete home mutation to its exact origin and installed lock.
	Unit *envunit.Unit
	// Directories enlists parent existence and protection separately from file
	// bytes. Prepare creates only private sidecars; live parents change at Commit.
	Directories []DirectoryAction
	// ReadInputs pins non-mutating input bytes across both commit and restart.
	// Callers must capture these under the mutation lock before planning.
	ReadInputs          map[string]string
	TransactionID       string
	ProjectIdentity     string
	Targets             []Target
	ReferencedBuildKeys []string
	// BoundaryCheck, when non-nil, re-validates destination separation
	// and physical identity immediately before each publication write
	// (see BoundaryCheck). It is in-memory only: Prepare retains it for
	// Commit in the same process. Nil skips the in-memory gate; the
	// durable proof below still guards restart recovery.
	BoundaryCheck BoundaryCheck
	// BoundaryProof, when non-nil, is the durable boundary record
	// persisted in the journal: the same canonical spellings and
	// filesystem identities the in-memory check pins, in journal-stable
	// form. A nil proof is a legacy unguarded journal; a non-nil proof
	// is protected and is re-verified before each remaining write after
	// a restart, or the commit fails closed. A plan with a check but no
	// proof (test scaffolding) journals an empty protected marker that
	// fails closed on recovery, never legacy.
	BoundaryProof *staging.DurableSnapshot
}

type OpaqueTarget struct {
	Class    string
	LivePath string
	Before   envunit.Image
}
type OpaqueRecord = envjournal.OpaqueRecord

// Journal is the canonical durable recovery record. Paths are absolute so
// home-only restart recovery never depends on the invoking project.
type Journal = envjournal.Journal

// DirectoryAction owns one no-follow directory entry's metadata, never its
// children. Missing parents must each have their own action. Mode is private.
type DirectoryAction struct {
	// EmptyMuseHome binds an intentionally empty runtime directory to the
	// enclosing unit's exact Muse home. It grants no arbitrary parent action.
	EmptyMuseHome string
	LivePath      string
	Before        envunit.Image
	Mode          uint32
}

// DirectoryRecord pins both the original entry and the exact prepared inode.
// Base is the existing staging ancestor, independently guarded after restart.
type DirectoryRecord = envjournal.DirectoryRecord

// RemovalEntry records one file, directory, or link that the transaction owns
// in an in-progress cleanup tomb. Missing entries are already durably removed;
// unrecorded or mismatched entries are unknown concurrent state.
//
// LinkTarget is the exact destination string of a "link" entry and is the whole
// of its content: recording it is what makes a link restorable and removable
// with the same certainty as a digested file.
type RemovalEntry = envjournal.RemovalEntry

// TargetRecord is the complete path-backed recovery record for one target.
type TargetRecord = envjournal.TargetRecord

// Point identifies a fault-injection and observation boundary.
type Point string

const (
	// PointGlobalRetained follows durable handled-conflict confirmation, before receipt publication.
	PointGlobalRetained          Point = "global_retained"
	PointAfterQuarantine         Point = "after_quarantine"
	PointAfterQuarantineRestore  Point = "after_quarantine_restore"
	PointDuringQuarantineRemoval Point = "during_quarantine_removal"
	// PointAfterDirectoryInstall follows metadata publication before its state
	// write. Recovery must prove the exact prepared inode, not path existence.
	PointAfterDirectoryInstall Point = "after_directory_install"
	// PointAfterStagingChunkSync follows a durable partial-file write while the
	// preparing journal still carries its write-ahead ownership record.
	PointAfterStagingChunkSync Point = "after_staging_chunk_sync"
	// PointDuringStagingCopy follows a durable partial-file write after the
	// preparing journal records its exact byte count and prefix digest.
	PointDuringStagingCopy Point = "during_staging_copy"
	// PointPrepared follows durable preparation.
	PointPrepared Point = "prepared"
	// PointBeforeBackup precedes the live-to-backup rename, or the desired
	// staging-to-backup rename for KindExchange (whose live name stays put).
	PointBeforeBackup Point = "before_backup"
	// PointAfterBackup follows the durable rename but precedes its state write.
	PointAfterBackup Point = "after_backup"
	// PointBeforeInstall follows durable backed-up state. KindExchange retains
	// the old tree at the live name until the atomic install exchange.
	PointBeforeInstall Point = "before_install"
	// PointAfterInstall follows the durable desired rename but precedes its state write.
	PointAfterInstall Point = "after_install"
	// PointTargetCommitted follows durable committed state.
	PointTargetCommitted Point = "target_committed"
	// PointAfterRestore follows durable restoration but precedes its state write.
	PointAfterRestore Point = "after_restore"
	// PointTargetRolledBack follows durable rolled-back state.
	PointTargetRolledBack Point = "target_rolled_back"
	// PointBeforeCleanup follows the durable cleanup phase.
	PointBeforeCleanup Point = "before_cleanup"
	// PointAfterCleanupRename follows a durable sidecar-to-tomb rename. The
	// journal already records ownership of the removal at this boundary.
	PointAfterCleanupRename Point = "after_cleanup_rename"
	// PointDuringCleanupRemoval follows one durable entry removal from an
	// owned cleanup tomb. A directory tomb may be only partially present.
	PointDuringCleanupRemoval Point = "during_cleanup_removal"
)

// Event describes one fault-injection or observation boundary. LivePath is the
// path the boundary concerns, so an ordering proof can inspect the state a
// target actually had without reaching into the journal.
type Event struct {
	Point         Point
	TransactionID string
	TargetIndex   int
	Class         string
	Identifier    string
	LivePath      string
}

// Hooks are deliberately fault-injectable. Commit-boundary faults drive
// durable rollback; cleanup faults retain durable recovery state. Observe is
// notification-only and is useful for proving ordering.
type Hooks struct {
	Fault   func(Event) error
	Observe func(Event)
	// ObserveRecovery reports entry to Recover, including an empty inventory.
	ObserveRecovery func()
}

// Option configures an Engine.
type Option func(*Engine)

// WithHooks installs fault-injection and observation hooks.
func WithHooks(hooks Hooks) Option {
	return func(engine *Engine) { engine.hooks = hooks }
}

func corruptionf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrImplementationCorruption, fmt.Sprintf(format, args...))
}

func journalf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", errInvalidJournal, fmt.Sprintf(format, args...))
}

func sidecarPath(livePath, transactionID string, index int, suffix string) string {
	name := fmt.Sprintf(".curator-txn-%s-%03d.%s", shortIdentity(transactionID), index, suffix)
	return filepath.Join(filepath.Dir(livePath), name)
}

// CommitCheck revalidates caller-owned dependencies before forward publication,
// including restart recovery. The journal is an independent snapshot. A refusal
// rolls back through the usual journal protocol; rollback never calls the check.
// Callers must configure the same policy on every engine that may recover their
// transactions. The borrowed lock belongs to the current commit/recovery call.
type CommitCheck func(HomeLock, Journal) error

// WithCommitCheck installs the caller's commit and recovery admission policy.
// A nil check preserves the existing transaction behavior.
func WithCommitCheck(check CommitCheck) Option {
	return func(engine *Engine) { engine.commitCheck = check }
}
