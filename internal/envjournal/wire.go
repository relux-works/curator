// Package envjournal defines the actual durable transaction wire. Sharing the
// data type does not admit an operation; readers and the executor validate the
// phase, complete target set and fresh filesystem evidence independently.
package envjournal

import (
	"github.com/relux-works/curator/internal/entrydigest"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/staging"
)

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
	// StateBackedUp records a durable backup before the desired swap. For
	// KindExchange, BackupPath holds the desired partner until that swap.
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
	// KindExchange owns the same bytes as KindBytes, but replaces a present
	// preimage with a present staged target using an atomic name exchange.
	// It is supported on Darwin and Linux. The old object remains at the
	// backup name until durable commit cleanup; rollback also exchanges names.
	KindExchange TargetKind = "exchange"
)

func (kind TargetKind) Valid() bool {
	switch kind {
	case KindBytes, KindEntry, KindExchange:
		return true
	default:
		return false
	}
}

// GlobalRetention is a potential, pinned profile-phase unmanaged conflict.
// Only retaining records the post-lock confirmation. Recovery before that point
// compensates; it never forwards home writes or infers takeover from a restart.
type GlobalRetention struct {
	Healthy  bool                         `json:"healthy"`
	Conflict envlifecycle.ConflictReceipt `json:"conflict"`
	State    string                       `json:"state"`
}

type Journal struct {
	Global              *GlobalRetention         `json:"global_retention,omitempty"`
	Opaque              []OpaqueRecord           `json:"opaque_quarantines,omitempty"`
	Unit                *envunit.Unit            `json:"environment_unit,omitempty"`
	Directories         []DirectoryRecord        `json:"directories,omitempty"`
	ReadInputs          map[string]string        `json:"read_inputs,omitempty"`
	Schema              string                   `json:"schema"`
	TransactionID       string                   `json:"transaction_id"`
	ProjectIdentity     string                   `json:"project_identity"`
	Phase               Phase                    `json:"phase"`
	RemovalPath         string                   `json:"removal_path,omitempty"`
	RemovalDigest       string                   `json:"removal_digest,omitempty"`
	RemovalEntries      []RemovalEntry           `json:"removal_entries,omitempty"`
	ReferencedBuildKeys []string                 `json:"referenced_build_keys"`
	Targets             []TargetRecord           `json:"targets"`
	BoundaryProof       *staging.DurableSnapshot `json:"boundary_proof,omitempty"`
}

// OpaqueRecord pins a whole home entry without a recursive manifest/digest.
// Explicit purge owns its removal; no credential/session bytes enter a journal.
type OpaqueRecord struct {
	Class          string        `json:"class,omitempty"`
	LivePath       string        `json:"live_path"`
	QuarantinePath string        `json:"quarantine_path"`
	Before         envunit.Image `json:"before"`
	State          CommitState   `json:"state"`
}

type DirectoryRecord struct {
	EmptyMuseHome string        `json:"empty_muse_home,omitempty"`
	LivePath      string        `json:"live_path"`
	Before        envunit.Image `json:"before"`
	After         envunit.Image `json:"after"`
	BasePath      string        `json:"base_path"`
	Base          envunit.Image `json:"base"`
	StagedPath    string        `json:"staged_path"`
	State         CommitState   `json:"state"`
}

type RemovalEntry struct {
	RelativePath string `json:"relative_path"`
	Kind         string `json:"kind"`
	Mode         uint32 `json:"mode"`
	Digest       string `json:"digest,omitempty"`
	LinkTarget   string `json:"link_target,omitempty"`
}

type TargetRecord struct {
	PrivateTree         bool           `json:"private_tree,omitempty"`
	PrivateFile         bool           `json:"private_file,omitempty"`
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
func (record *TargetRecord) Digest(path string) (string, error) {
	return entrydigest.DigestTarget(string(record.Kind), path)
}
