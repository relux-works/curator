package swiftpmsource

import (
	"context"
	"path"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
)

const (
	// ProfileID identifies the restricted source-only SwiftPM adapter.
	ProfileID = "swiftpm-source-v1"
	// ConditionEvaluatorID owns SwiftPM platform/configuration/trait predicates.
	ConditionEvaluatorID = "swiftpm-condition-v1"
	// ManifestSchemaID identifies normalized dump-package evidence.
	ManifestSchemaID = "swiftpm-manifest-evaluation-v1"
)

// ToolIdentity is one exact C0 evidence executable or runtime.
type ToolIdentity struct {
	Role, ExecutableRelativePath, VersionOutput, PlatformABI, PolicySelector string
	Fingerprint, ExecutableSHA256                                            closuregraph.ID
	ProcessFamily                                                            []ToolProcessIdentity
}

// ToolProcessIdentity binds an additional executable that the selected tool
// may start. Portable mode declares but does not claim observation of this
// family; a verified provider must enforce it.
type ToolProcessIdentity struct {
	ExecutableRelativePath string
	ExecutableSHA256       closuregraph.ID
}

// Toolchain binds every executable/runtime able to derive evidence before C5.
type Toolchain struct {
	Swift, SwiftPM, PackageDescription, Git ToolIdentity
	Recheck                                 func(context.Context, ToolIdentity) (closureexec.ToolchainIdentity, error)
}

// Destination is the exact selection binding. Abstract platform predicates
// remain in capture; these concrete facts occur only in SelectionBinding.
type Destination struct {
	Platform closuregraph.TargetPlatformPayload
	Markers  map[string]string
}

// ManifestDependency is one unevaluated package declaration.
type ManifestDependency struct {
	Identity, Location, Requirement, LocalPath string
	Kind                                       SourceKind
}

// TargetDependency is one target/product/package edge with an optional
// selection-neutral condition.
type TargetDependency struct {
	Name, Package, Product string
	Condition              *closuregraph.Condition
}

// BuildSetting retains conditions and rejects unsafe flags before planning.
type BuildSetting struct {
	Kind, Value string
	Condition   *closuregraph.Condition
	Unsafe      bool
}

// Target is the source-closure projection emitted by controlled dump-package.
// PublicHeadersPath retains SwiftPM's declared C-family public-header
// directory verbatim; an empty value means the target declared none and the
// consumer applies SwiftPM's documented default.
type Target struct {
	Name, Type, Path  string
	PublicHeadersPath string
	Sources           []string
	Dependencies      []TargetDependency
	Settings          []BuildSetting
}

// SourceRoot is the package-relative directory SwiftPM enumerates this
// target's sources from. It is the declared path when the manifest gives one
// and SwiftPM's documented convention default otherwise. SwiftPM's native
// build system mirrors the source tree below this root into the target build
// directory, so a consumer that must reconcile a produced object with its
// declared source resolves the source against exactly this root.
func (target Target) SourceRoot() string {
	return targetSourceRoot(target.Name, target.Type, target.Path)
}

// targetSourceRoot applies the documented default exactly once so manifest
// normalization and downstream reconciliation cannot disagree.
func targetSourceRoot(name, kind, declared string) string {
	if declared != "" {
		return declared
	}
	base := "Sources"
	if kind == "test" {
		base = "Tests"
	}
	return path.Join(base, name)
}

// Product is one declared package product.
type Product struct {
	Name, Type string
	Targets    []string
}

// Manifest is normalized executable Package.swift evidence. The adapter
// hashes this value independently; evaluators do not supply its authority.
type Manifest struct {
	PackageName, ToolsVersion, SelectedManifest string
	Dependencies                                []ManifestDependency
	Products                                    []Product
	Targets                                     []Target
	Traits                                      []string
}

// ManifestPermit is committed by the adapter before the evaluator seam.
type ManifestPermit struct {
	ID, C0CheckpointID, IntakeReceiptID closuregraph.ID
	PackageIdentity, SelectedManifest   string
	ToolchainFingerprint                closuregraph.ID
	Argv                                []string
	Environment                         map[string]string
	Network                             string
	input                               closureexec.AdmittedInput
	ToolchainNodeID, HostID, TargetID   closuregraph.ID
}

// ManifestEvaluator performs exactly one controlled manifest derivation.
// Root is always an admitted, immutable protected tree.
type ManifestEvaluator interface {
	Evaluate(context.Context, string, ManifestPermit) (ManifestResult, error)
}

// ManifestResult couples normalized evaluator output to the executor-issued
// receipt for the exact committed process.
type ManifestResult struct {
	Manifest  Manifest
	ReceiptID closuregraph.ID
}

// LockResolver may create a candidate top-level lock only after the complete
// root tree is admitted and its manifest permit succeeds.
type LockResolver interface {
	Resolve(context.Context, string, ResolutionPermit, Manifest) (ResolutionResult, error)
	VerifyResult(ResolutionPermit, ResolutionResult) error
}

// ResolutionResult couples generated lock bytes to the executor-issued
// derivation receipt that causally produced them.
type ResolutionResult struct {
	Lock                 []byte
	ReceiptID            closuregraph.ID
	JournalEntryIDs      []closuregraph.ID
	DerivationReceiptIDs []closuregraph.ID
	GitPermitIDs         []closuregraph.ID
	GitReceiptIDs        []closuregraph.ID
}

// ResolutionPermit binds the only manager resolution process allowed before
// the generated root lock is frozen. Network belongs exclusively to the
// acquisition broker named by this permit.
type ResolutionPermit struct {
	ID, C0CheckpointID, IntakeReceiptID closuregraph.ID
	ToolchainFingerprint                closuregraph.ID
	AlgorithmID                         string
	Environment                         map[string]string
	Network                             string
	input                               closureexec.AdmittedInput
	ToolchainNodeID, HostID, TargetID   closuregraph.ID
}

// Snapshot is one exact broker acquisition result. Root is the checked-out
// complete package tree; MirrorRoot is a same-kind source-control repository
// that contains Revision without needing the original origin.
type Snapshot struct {
	Identity, Root, MirrorRoot, Revision, GitTree             string
	Kind                                                      SourceKind
	BrokerReceiptID                                           closuregraph.ID
	AcquisitionReceipt                                        closureexec.SourceAcquisitionReceipt
	CommitObject                                              []byte
	BrokerPermitIDs, BrokerProcessReceiptIDs                  []closuregraph.ID
	UsesSubmodules, UsesLFS, UsesCheckoutFilter, RequiresHook bool
}

// AcquisitionBroker is the sole network-capable source-control boundary. It
// returns bytes but never evaluates manifests or package hooks.
type AcquisitionBroker interface {
	Acquire(context.Context, Pin) (Snapshot, error)
}

type acquisitionEvidenceVerifier interface {
	VerifySnapshot(Pin, Snapshot) error
}

// MirrorVerifier proves that a captured repository contains the exact pinned
// commit and tree without consulting its original origin.
type MirrorVerifier interface {
	Verify(context.Context, string, Pin, Snapshot) (GitVerificationEvidence, error)
}

// GitVerificationEvidence binds every mirror inspection subprocess to its
// committed permit and authority-issued receipt.
type GitVerificationEvidence struct {
	PermitIDs, ReceiptIDs []closuregraph.ID
}

// OfflineMetadataRunner performs the authoritative SwiftPM metadata replay
// from the admitted root and kind-preserving mirrors under network denial.
type OfflineMetadataRunner interface {
	Replay(context.Context, *Capture) (OfflineMetadataResult, error)
}

// OfflineMetadataResult is executor-issued evidence from show-dependencies.
type OfflineMetadataResult struct {
	ReceiptID         closuregraph.ID
	PackageIdentities []string
}

// Mirror preserves both the canonical original kind and local mirror kind.
type Mirror struct {
	Identity, Original, Local, Revision, GitTree  string
	OriginalKind, LocalKind                       SourceKind
	SnapshotDigest, MirrorDigest                  closuregraph.ID
	BrokerReceiptID                               closuregraph.ID
	MirrorIntakeReceiptID                         closuregraph.ID
	ArtifactManifestID                            closuregraph.ID
	CommitEvidenceIntakeReceiptID                 closuregraph.ID
	CommitEvidenceArtifactManifestID              closuregraph.ID
	MirrorDerivationPermitID                      closuregraph.ID
	MirrorDerivationReceiptID                     closuregraph.ID
	VerificationPermitIDs, VerificationReceiptIDs []closuregraph.ID
	AuthorizedOutputPath                          string
	authorization                                 artifactpolicy.SourceControlMirrorAuthorization
	input                                         closureexec.AdmittedInput
	commitInput                                   closureexec.AdmittedInput
}

// Config supplies only central services and exact selected evidence.
type Config struct {
	Store                *closureexec.CaptureStore
	Policy               *artifactpolicy.Service
	Evaluator            ManifestEvaluator
	Resolver             LockResolver
	Broker               AcquisitionBroker
	MirrorVerifier       MirrorVerifier
	OfflineRunner        OfflineMetadataRunner
	Toolchain            Toolchain
	GitExecutionRoot     string
	GitToolRoot          string
	Destination          Destination
	CausalHead           string
	ProcessStartObserver func(ManifestPermit)
}

// Request selects exactly one executable product and root package tree.
type Request struct {
	Root, Product string
	Resolved      []byte
}

// PackageEvidence is the immutable intake, manifest, origin, and mirror record.
type PackageEvidence struct {
	Identity, Origin, Revision, GitTree, SelectedManifest                 string
	Kind                                                                  SourceKind
	SnapshotDigest, ArtifactManifestID                                    closuregraph.ID
	IntakeReceiptID, ManifestPermitID, ManifestReceiptID, BrokerReceiptID closuregraph.ID
	ManifestDigest                                                        closuregraph.ID
	SourceInventoryDigest                                                 closuregraph.ID
	Mirror                                                                *Mirror
	Manifest                                                              Manifest
	BrokerPermitIDs, BrokerProcessReceiptIDs                              []closuregraph.ID
	input                                                                 closureexec.AdmittedInput
	evaluationSubpath                                                     string
}

// Capture is the selection-neutral source closure plus exact selected overlay.
type Capture struct {
	Lock, RootLock               Lock
	Packages                     []PackageEvidence
	Mirrors                      []Mirror
	Graph                        closuregraph.CaptureGraph
	Selection                    closuregraph.SelectionContext
	Binding                      closuregraph.SelectionBinding
	Active                       closuregraph.ActiveGraph
	Records                      closuregraph.RecordTables
	Authority                    closuregraph.BindingAuthority
	C0, C1, C2, C3, C4           closuregraph.Checkpoint
	ResolutionPermitID           closuregraph.ID
	ResolutionReceiptID          closuregraph.ID
	GraphDigest, InventoryDigest closuregraph.ID
	ProductNodeID                closuregraph.ID
	TargetNodeIDs                []closuregraph.ID
	CausalHead                   string
	rootInput                    closureexec.AdmittedInput
	config                       Config
}
