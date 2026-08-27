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
)

const (
	journalSchema = "curator-install-transaction-v1"
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

// Phase identifies the durable transaction-wide recovery action.
type Phase string

const (
	// PhasePreparing records a journal whose sibling staging may be incomplete.
	PhasePreparing Phase = "preparing"
	// PhasePrepared records fully durable staging before live mutation.
	PhasePrepared Phase = "prepared"
	// PhaseCommitting resumes ordered target replacement.
	PhaseCommitting Phase = "committing"
	// PhaseCleanup resumes backup deletion after every consumer is durable.
	PhaseCleanup Phase = "cleanup"
	// PhaseRollingBack resumes exact reverse target restoration.
	PhaseRollingBack Phase = "rolling_back"
	// PhaseRollbackCleanup resumes sidecar deletion after restoration.
	PhaseRollbackCleanup Phase = "rollback_cleanup"
)

// CommitState identifies one target's durable progress.
type CommitState string

const (
	// StatePending records a target whose live preimage has not been moved.
	StatePending CommitState = "pending"
	// StateBackedUp records a durable backup before the desired swap.
	StateBackedUp CommitState = "backed_up"
	// StateCommitted records a durable desired consumer target.
	StateCommitted CommitState = "committed"
	// StateRolledBack records a durably restored preimage.
	StateRolledBack CommitState = "rolled_back"
)

// TargetKind selects how the engine interprets one target's own path.
type TargetKind string

const (
	// KindBytes is the default target kind: one regular file or one link-free
	// directory tree, owned as bytes. Its path is fully resolved for namespace
	// independence, so two targets that alias one object through a link are
	// refused, and a link is never an acceptable live, staged, or backup state.
	KindBytes TargetKind = ""
	// KindEntry is one directory entry the manager owns as itself, which may be
	// a symbolic link — an agent adapter mirror in symlink mode is the reason
	// this kind exists. The final path component is never resolved, and a link
	// is digested, staged, backed up, restored, and removed as the link it is:
	// its complete content is its destination string. Entries inside a directory
	// tree are still refused, exactly as for KindBytes.
	KindEntry TargetKind = "entry"
)

func (kind TargetKind) valid() bool {
	switch kind {
	case KindBytes, KindEntry:
		return true
	default:
		return false
	}
}

// Target describes one desired replacement. Class and Identifier are the
// canonical ordering key. An empty StagedSource means the desired state is
// absence. Exactly one preimage expectation is required: PreimageDigest, or
// ExpectedGeneration read from GenerationPath (relative to LivePath, or the
// live file itself when empty). A KindEntry target must use PreimageDigest,
// because a link carries no generation file.
type Target struct {
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
	TransactionID       string
	ProjectIdentity     string
	Targets             []Target
	ReferencedBuildKeys []string
}

// Journal is the canonical durable recovery record. Paths are absolute so
// home-only restart recovery never depends on the invoking project.
type Journal struct {
	Schema              string         `json:"schema"`
	TransactionID       string         `json:"transaction_id"`
	ProjectIdentity     string         `json:"project_identity"`
	Phase               Phase          `json:"phase"`
	RemovalPath         string         `json:"removal_path,omitempty"`
	RemovalDigest       string         `json:"removal_digest,omitempty"`
	RemovalEntries      []RemovalEntry `json:"removal_entries,omitempty"`
	ReferencedBuildKeys []string       `json:"referenced_build_keys"`
	Targets             []TargetRecord `json:"targets"`
}

// RemovalEntry records one file, directory, or link that the transaction owns
// in an in-progress cleanup tomb. Missing entries are already durably removed;
// unrecorded or mismatched entries are unknown concurrent state.
//
// LinkTarget is the exact destination string of a "link" entry and is the whole
// of its content: recording it is what makes a link restorable and removable
// with the same certainty as a digested file.
type RemovalEntry struct {
	RelativePath string `json:"relative_path"`
	Kind         string `json:"kind"`
	Mode         uint32 `json:"mode"`
	Digest       string `json:"digest,omitempty"`
	LinkTarget   string `json:"link_target,omitempty"`
}

// TargetRecord is the complete path-backed recovery record for one target.
type TargetRecord struct {
	Class               string         `json:"class"`
	Identifier          string         `json:"identifier"`
	Kind                TargetKind     `json:"kind,omitempty"`
	LivePath            string         `json:"live_path"`
	StagedPath          string         `json:"staged_path,omitempty"`
	StagedSource        string         `json:"staged_source,omitempty"`
	StagingEntries      []RemovalEntry `json:"staging_entries,omitempty"`
	StagingIndex        int            `json:"staging_index,omitempty"`
	StagingActive       bool           `json:"staging_active,omitempty"`
	StagingCreated      bool           `json:"staging_created,omitempty"`
	StagingBytes        int64          `json:"staging_bytes,omitempty"`
	StagingPrefixDigest string         `json:"staging_prefix_digest,omitempty"`
	StagingWriteBytes   int64          `json:"staging_write_bytes,omitempty"`
	StagingWriteDigest  string         `json:"staging_write_digest,omitempty"`
	StagingDiscarded    bool           `json:"staging_discarded,omitempty"`
	PreimageDigest      string         `json:"preimage_digest,omitempty"`
	ExpectedGeneration  string         `json:"expected_generation,omitempty"`
	GenerationPath      string         `json:"generation_path,omitempty"`
	BackupPath          string         `json:"backup_path"`
	RollbackPath        string         `json:"rollback_path"`
	DesiredDigest       string         `json:"desired_digest"`
	BackupDigest        string         `json:"backup_digest,omitempty"`
	State               CommitState    `json:"state"`
}

// digest reads one of this target's own paths — live, staged, backup, or
// rollback — with the strictness its kind requires. Every state comparison in
// the engine goes through here so a target can never be read as a shape it does
// not own.
func (record *TargetRecord) digest(path string) (string, error) {
	return DigestTarget(record.Kind, path)
}

// Point identifies a fault-injection and observation boundary.
type Point string

const (
	// PointAfterStagingChunkSync follows a durable partial-file write while the
	// preparing journal still carries its write-ahead ownership record.
	PointAfterStagingChunkSync Point = "after_staging_chunk_sync"
	// PointDuringStagingCopy follows a durable partial-file write after the
	// preparing journal records its exact byte count and prefix digest.
	PointDuringStagingCopy Point = "during_staging_copy"
	// PointPrepared follows durable preparation.
	PointPrepared Point = "prepared"
	// PointBeforeBackup precedes the live-to-backup rename.
	PointBeforeBackup Point = "before_backup"
	// PointAfterBackup follows the durable rename but precedes its state write.
	PointAfterBackup Point = "after_backup"
	// PointBeforeInstall follows durable backed-up state.
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
