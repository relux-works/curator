// Package artifactpolicy implements Curator's adapter-independent, fail-closed
// artifact admission policy. It classifies immutable dependency bytes before
// any package-manager execution and keeps toolchain and locally produced output
// trust roles causally separate from dependency input.
package artifactpolicy

import (
	"context"
	"fmt"
	"io"
	"slices"
)

// SchemaID is the only canonical artifact-manifest schema implemented here.
const SchemaID = "artifact-manifest-v1"

// PolicyID is the immutable policy and default-limit identity.
const PolicyID = "curator-artifact-policy-v1"

// PolicyVersion is the closed policy major version.
const PolicyVersion = 1

// DetectorRegistryID identifies the complete detector set used by policy v1.
const DetectorRegistryID = "curator-artifact-detectors-v1"

// LimitVectorID identifies the immutable policy-v1 resource limits.
const LimitVectorID = "curator-artifact-limits-v1"

// ArtifactClass is a closed, stable artifact taxonomy value.
type ArtifactClass string

// Closed artifact classes. Adapters may narrow admitted source grammars but
// cannot add another class or change a class decision.
const (
	ClassSourceAuthoredText   ArtifactClass = "source.authored_text"
	ClassSourceGeneratedText  ArtifactClass = "source.generated_text"
	ClassTextMetadata         ArtifactClass = "text.metadata"
	ClassDataKnownInert       ArtifactClass = "data.known_inert"
	ClassNativeExecutable     ArtifactClass = "native.executable"
	ClassNativeObject         ArtifactClass = "native.object"
	ClassNativeLibraryStatic  ArtifactClass = "native.library.static"
	ClassNativeLibraryDynamic ArtifactClass = "native.library.dynamic"
	ClassELFETDYNAmbiguous    ArtifactClass = "native.elf.et_dyn_ambiguous"
	ClassAppleFramework       ArtifactClass = "apple.framework"
	ClassAppleXCFramework     ArtifactClass = "apple.xcframework"
	ClassNodeExtension        ArtifactClass = "native.extension.node"
	ClassPythonExtension      ArtifactClass = "native.extension.python"
	ClassJVMBytecode          ArtifactClass = "vm.jvm_bytecode"
	ClassPythonBytecode       ArtifactClass = "vm.python_bytecode"
	ClassJavaScriptCodeCache  ArtifactClass = "vm.javascript_code_cache"
	ClassWebAssembly          ArtifactClass = "ir.webassembly"
	ClassCompilerSerialized   ArtifactClass = "ir.compiler_serialized"
	ClassArchive              ArtifactClass = "container.archive"
	ClassCompressedStream     ArtifactClass = "container.compressed_stream"
	ClassDirectory            ArtifactClass = "fs.directory"
	ClassLink                 ArtifactClass = "fs.symlink_or_hardlink"
	ClassSpecial              ArtifactClass = "fs.special"
	ClassOpaqueUnknown        ArtifactClass = "opaque.unknown"
)

var artifactClasses = closedSet(
	ClassSourceAuthoredText, ClassSourceGeneratedText, ClassTextMetadata,
	ClassDataKnownInert, ClassNativeExecutable, ClassNativeObject,
	ClassNativeLibraryStatic, ClassNativeLibraryDynamic, ClassELFETDYNAmbiguous,
	ClassAppleFramework, ClassAppleXCFramework, ClassNodeExtension,
	ClassPythonExtension, ClassJVMBytecode, ClassPythonBytecode,
	ClassJavaScriptCodeCache, ClassWebAssembly, ClassCompilerSerialized,
	ClassArchive, ClassCompressedStream, ClassDirectory, ClassLink, ClassSpecial,
	ClassOpaqueUnknown,
)

// TrustRole records the causal origin of trust. Names, paths, and package
// metadata never establish a role.
type TrustRole string

// Closed trust roles.
const (
	RoleDependencyInput         TrustRole = "dependency_input"
	RoleExternalToolchain       TrustRole = "external_toolchain"
	RoleLocalBuildOutput        TrustRole = "local_build_output"
	RoleVerifiedBinaryCandidate TrustRole = "verified_binary_candidate"
)

var trustRoles = closedSet(
	RoleDependencyInput, RoleExternalToolchain, RoleLocalBuildOutput,
	RoleVerifiedBinaryCandidate,
)

// Decision is a closed artifact decision. DESCEND is structural and is never
// sufficient by itself to authorize execution or publication.
type Decision string

// Closed decisions.
const (
	DecisionAdmitInput     Decision = "ADMIT_INPUT"
	DecisionDescend        Decision = "DESCEND"
	DecisionReject         Decision = "REJECT"
	DecisionAllowToolchain Decision = "ALLOW_TOOLCHAIN"
	DecisionAllowOutput    Decision = "ALLOW_OUTPUT"
)

var decisions = closedSet(
	DecisionAdmitInput, DecisionDescend, DecisionReject,
	DecisionAllowToolchain, DecisionAllowOutput,
)

// DiagnosticCode is a stable machine-readable policy-v1 failure code.
type DiagnosticCode string

// Closed policy-v1 diagnostics.
const (
	CodeOriginUnverified           DiagnosticCode = "artifact_origin_unverified"
	CodeCompiledDependency         DiagnosticCode = "artifact_compiled_dependency_forbidden"
	CodeBinaryAdmissionUnavailable DiagnosticCode = "artifact_binary_admission_unavailable"
	CodeTypeAmbiguous              DiagnosticCode = "artifact_type_ambiguous"
	CodeOpaqueDependency           DiagnosticCode = "artifact_opaque_dependency_forbidden"
	CodeArchiveInvalid             DiagnosticCode = "artifact_archive_invalid"
	CodeArchiveUnsupported         DiagnosticCode = "artifact_archive_unsupported"
	CodeArchiveEncrypted           DiagnosticCode = "artifact_archive_encrypted"
	CodeArchiveUnsafePath          DiagnosticCode = "artifact_archive_unsafe_path"
	CodeArchiveUnsafeEntry         DiagnosticCode = "artifact_archive_unsafe_entry"
	CodeInspectionLimitExceeded    DiagnosticCode = "artifact_inspection_limit_exceeded"
	CodeInspectionUnavailable      DiagnosticCode = "artifact_inspection_unavailable"
	CodeGeneratedInputUndeclared   DiagnosticCode = "artifact_generated_input_undeclared"
	CodeToolchainUntrusted         DiagnosticCode = "artifact_toolchain_untrusted"
	CodeToolchainIdentityChanged   DiagnosticCode = "artifact_toolchain_identity_changed"
	CodeLocalOutputUnreceipted     DiagnosticCode = "artifact_local_output_unreceipted"
	CodeLocalOutputDrift           DiagnosticCode = "artifact_local_output_drift"
	CodePolicyInternalError        DiagnosticCode = "artifact_policy_internal_error"
)

var diagnosticCodes = closedSet(
	CodeOriginUnverified, CodeCompiledDependency, CodeBinaryAdmissionUnavailable,
	CodeTypeAmbiguous, CodeOpaqueDependency, CodeArchiveInvalid,
	CodeArchiveUnsupported, CodeArchiveEncrypted, CodeArchiveUnsafePath,
	CodeArchiveUnsafeEntry, CodeInspectionLimitExceeded,
	CodeInspectionUnavailable, CodeGeneratedInputUndeclared,
	CodeToolchainUntrusted, CodeToolchainIdentityChanged,
	CodeLocalOutputUnreceipted, CodeLocalOutputDrift, CodePolicyInternalError,
)

// NodeKind identifies a logical manifest node.
type NodeKind string

// Closed node kinds.
const (
	NodeRegularFile      NodeKind = "regular_file"
	NodeDirectory        NodeKind = "directory"
	NodeArchive          NodeKind = "archive"
	NodeCompressedStream NodeKind = "compressed_stream"
	NodeLink             NodeKind = "link"
	NodeSpecial          NodeKind = "special"
)

var nodeKinds = closedSet(
	NodeRegularFile, NodeDirectory, NodeArchive, NodeCompressedStream,
	NodeLink, NodeSpecial,
)

// ProfileID identifies a central, closed source grammar profile. The profiles
// share detectors and differ only by removing eligible source grammars.
type ProfileID string

// Closed source profiles used by the current adapter delivery scope.
const (
	ProfileCommonV1       ProfileID = "common-source-v1"
	ProfileGoV1           ProfileID = "go-source-v1"
	ProfileRustV1         ProfileID = "rust-source-v1"
	ProfileNodeV1         ProfileID = "node-source-v1"
	ProfilePythonSourceV1 ProfileID = "python-source-container-v1"
	ProfileSwiftPMV1      ProfileID = "swiftpm-source-v1"
)

var profileIDs = closedSet(
	ProfileCommonV1, ProfileGoV1, ProfileRustV1, ProfileNodeV1,
	ProfilePythonSourceV1, ProfileSwiftPMV1,
)

// GrammarID identifies one closed text parser or lexer.
type GrammarID string

// Closed source and metadata grammars.
const (
	GrammarPlain          GrammarID = "plain-text-v1"
	GrammarMarkdown       GrammarID = "markdown-text-v1"
	GrammarJSON           GrammarID = "json-v1"
	GrammarTOML           GrammarID = "toml-v1"
	GrammarYAML           GrammarID = "yaml-text-v1"
	GrammarGo             GrammarID = "go-v1"
	GrammarRust           GrammarID = "rust-lexer-v1"
	GrammarSwift          GrammarID = "swift-lexer-v1"
	GrammarC              GrammarID = "c-lexer-v1"
	GrammarCXX            GrammarID = "cxx-lexer-v1"
	GrammarObjectiveC     GrammarID = "objective-c-lexer-v1"
	GrammarJavaScript     GrammarID = "javascript-lexer-v1"
	GrammarTypeScript     GrammarID = "typescript-lexer-v1"
	GrammarPython         GrammarID = "python-lexer-v1"
	GrammarShell          GrammarID = "shell-lexer-v1"
	GrammarAssembly       GrammarID = "assembly-text-v1"
	GrammarModuleMap      GrammarID = "clang-modulemap-lexer-v1"
	GrammarSwiftInterface GrammarID = "swiftinterface-lexer-v1"
	GrammarSourceMap      GrammarID = "source-map-json-v1"
)

var grammarIDs = closedSet(
	GrammarPlain, GrammarMarkdown, GrammarJSON, GrammarTOML, GrammarYAML,
	GrammarGo, GrammarRust, GrammarSwift, GrammarC, GrammarCXX,
	GrammarObjectiveC, GrammarJavaScript, GrammarTypeScript, GrammarPython,
	GrammarShell, GrammarAssembly, GrammarModuleMap, GrammarSwiftInterface,
	GrammarSourceMap,
)

// UseKind is a manager-resolved use edge that can disambiguate structurally
// valid ELF ET_DYN payloads. A filename alone is not a use edge.
type UseKind string

// Closed manager-resolved use kinds.
const (
	UseExecute    UseKind = "execute"
	UseLinkOrLoad UseKind = "link_or_load"
)

var useKinds = closedSet(UseExecute, UseLinkOrLoad)

// Fact is one canonical detector or diagnostic fact. Fact slices are sorted
// by key and value before manifest encoding.
type Fact struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// UseEdge records a manager-resolved use and the canonical graph field that
// established it.
type UseEdge struct {
	Kind   UseKind `json:"kind"`
	Origin string  `json:"origin"`
}

// TextDeclaration binds a virtual path to one closed grammar and source class.
// GeneratedLineage is required when the adapter profile says generated inputs
// must carry declared lineage.
type TextDeclaration struct {
	Grammar          GrammarID     `json:"grammar"`
	Class            ArtifactClass `json:"class"`
	GeneratedLineage string        `json:"generated_lineage"`
}

// LimitVector is the complete immutable policy-v1 resource-limit vector.
type LimitVector struct {
	ID                   string `json:"id"`
	MaxRawPayloadBytes   int64  `json:"max_raw_payload_bytes"`
	MaxSingleLeafBytes   int64  `json:"max_single_leaf_bytes"`
	MaxTotalEmittedBytes int64  `json:"max_total_emitted_bytes"`
	MaxArchiveDepth      int64  `json:"max_archive_depth"`
	MaxContainerCount    int64  `json:"max_container_count"`
	MaxEntryCount        int64  `json:"max_entry_count"`
	MaxExpansionRatio    int64  `json:"max_expansion_ratio"`
	MaxPathBytes         int64  `json:"max_path_bytes"`
	MaxComponentBytes    int64  `json:"max_component_bytes"`
	MaxRecordedFindings  int64  `json:"max_recorded_findings"`
}

// DefaultLimits returns the exact immutable curator-artifact-policy-v1 limits.
func DefaultLimits() LimitVector {
	return LimitVector{
		ID:                   LimitVectorID,
		MaxRawPayloadBytes:   512 << 20,
		MaxSingleLeafBytes:   256 << 20,
		MaxTotalEmittedBytes: 2 << 30,
		MaxArchiveDepth:      8,
		MaxContainerCount:    1_024,
		MaxEntryCount:        100_000,
		MaxExpansionRatio:    200,
		MaxPathBytes:         4_096,
		MaxComponentBytes:    255,
		MaxRecordedFindings:  1_000,
	}
}

// DetectorIdentity binds a detector name to its implementation version.
type DetectorIdentity struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// OriginEvidence proves immutable capture before dependency admission.
type OriginEvidence struct {
	Locator        string `json:"locator"`
	ImmutableID    string `json:"immutable_id"`
	LockRecord     string `json:"lock_record"`
	ChecksumSHA256 string `json:"checksum_sha256"`
	Verified       bool   `json:"verified"`
}

// Descriptor supplies adapter-neutral audit metadata. DeclaredText and
// ResolvedUses are keyed by canonical virtual path.
type Descriptor struct {
	AdapterID               string
	ProfileID               ProfileID
	Manager                 string
	PackageName             string
	PackageVersion          string
	Origin                  OriginEvidence
	DeclaredText            map[string]TextDeclaration
	ResolvedUses            map[string][]UseEdge
	RequireGeneratedLineage bool
}

// Payload is one immutable regular-file byte sequence. Size must be exact;
// readers that end early or produce trailing bytes fail closed.
type Payload struct {
	Path   string
	Size   int64
	Reader io.Reader
}

// ArtifactExpectation is the independently derived expected output identity.
type ArtifactExpectation struct {
	Path   string
	Class  ArtifactClass
	SHA256 string
	Size   int64
}

// ToolchainAuthorization is a sealed central-manager checkpoint capability.
// Adapter packages cannot implement it, construct it from caller-selected
// roots, or populate its evidence. This package intentionally exposes no
// caller-configurable issuer. The manager integration may supply a value only
// after central selection, complete dependency-boundary exclusion, contained
// root fingerprinting, environment resolution, and time-of-use checks pass.
type ToolchainAuthorization interface {
	artifactPolicyToolchainAuthorization() toolchainAuthorizationRecord
}

// LocalOutputAuthorization is an opaque protected-executor receipt. Adapter
// packages cannot implement it or turn caller assertions, a fresh directory,
// or an expected digest into local-output trust. This package intentionally
// exposes no issuer. A future protected-executor integration may supply a
// value only after binding the declared action and toolchain, complete inputs
// and reads, observed process and writes, and exact protected publication.
type LocalOutputAuthorization interface {
	artifactPolicyLocalOutputAuthorization() localOutputAuthorizationRecord
}

// DependencyRequest inspects immutable package/dependency bytes.
type DependencyRequest struct {
	Descriptor Descriptor
	Payload    Payload
}

// ToolchainRequest inspects one independently selected toolchain component.
type ToolchainRequest struct {
	Descriptor    Descriptor
	Payload       Payload
	Authorization ToolchainAuthorization
}

// LocalOutputRequest inspects one causally produced protected output. A nil or
// foreign Authorization is fail-closed; staging bytes alone are never proof of
// production.
type LocalOutputRequest struct {
	Descriptor    Descriptor
	Payload       Payload
	Authorization LocalOutputAuthorization
}

// VerifiedBinaryRequest is the explicit unavailable capability seam. It never
// falls back to dependency source admission.
type VerifiedBinaryRequest struct {
	Descriptor Descriptor
	Payload    Payload
}

// DirectoryRequest inspects an immutable captured directory without following
// links. Root is operational and is excluded from portable manifest identity;
// VirtualRoot is the portable manifest root.
type DirectoryRequest struct {
	Descriptor  Descriptor
	Root        string
	VirtualRoot string
}

// Observation is a normalized detector result.
type Observation struct {
	DetectorID string `json:"detector_id"`
	Result     string `json:"result"`
	Facts      []Fact `json:"facts"`
}

// ManifestNode is one canonical artifact, member, synthetic directory, link,
// or special node.
type ManifestNode struct {
	Path               string        `json:"path"`
	OriginalNameBase64 string        `json:"original_name_base64"`
	CollisionKey       string        `json:"collision_key"`
	Kind               NodeKind      `json:"kind"`
	Parent             string        `json:"parent"`
	ContainerChain     []string      `json:"container_chain"`
	Size               int64         `json:"size"`
	SHA256             string        `json:"sha256"`
	Mode               int64         `json:"mode"`
	DeclaredUses       []UseEdge     `json:"declared_uses"`
	Observations       []Observation `json:"observations"`
	SelectedDetectorID string        `json:"selected_detector_id"`
	Class              ArtifactClass `json:"class"`
	Variant            string        `json:"variant"`
	Decision           Decision      `json:"decision"`
	Rule               string        `json:"rule"`
	InspectionComplete bool          `json:"inspection_complete"`
}

// Diagnostic is one stable structured finding. Human rendering is explicitly
// not the machine interface.
type Diagnostic struct {
	Code               DiagnosticCode `json:"code"`
	Path               string         `json:"path"`
	OriginalNameBase64 string         `json:"original_name_base64"`
	CollisionKey       string         `json:"collision_key"`
	Class              ArtifactClass  `json:"class"`
	Variant            string         `json:"variant"`
	DetectorID         string         `json:"detector_id"`
	Reason             string         `json:"reason"`
	ContainerChain     []string       `json:"container_chain"`
	SHA256             string         `json:"sha256"`
	Size               int64          `json:"size"`
	LimitName          string         `json:"limit_name"`
	Limit              int64          `json:"limit"`
	Observed           int64          `json:"observed"`
	Details            []Fact         `json:"details"`
}

// FindingEvidence is the compact, canonical semantic projection of one
// diagnostic. Evidence retains every finding even when the display-oriented
// Diagnostics slice is capped. DiagnosticSHA256 binds the exact full
// diagnostic (including Details); DetailsSHA256 makes that omitted detail
// projection explicit instead of leaving an unverifiable hash-only gap.
type FindingEvidence struct {
	DiagnosticSHA256   string         `json:"diagnostic_sha256"`
	Code               DiagnosticCode `json:"code"`
	Path               string         `json:"path"`
	OriginalNameBase64 string         `json:"original_name_base64"`
	CollisionKey       string         `json:"collision_key"`
	Class              ArtifactClass  `json:"class"`
	Variant            string         `json:"variant"`
	DetectorID         string         `json:"detector_id"`
	Reason             string         `json:"reason"`
	ContainerChain     []string       `json:"container_chain"`
	SHA256             string         `json:"sha256"`
	Size               int64          `json:"size"`
	LimitName          string         `json:"limit_name"`
	Limit              int64          `json:"limit"`
	Observed           int64          `json:"observed"`
	Details            []Fact         `json:"details"`
	DetailsSHA256      string         `json:"details_sha256"`
}

// FindingsSummary binds the complete canonical diagnostic set even when the
// display-oriented Diagnostics slice is capped. Evidence is always the full
// canonical sequence; Recorded is exactly min(Total, max_recorded_findings).
type FindingsSummary struct {
	Algorithm string            `json:"algorithm"`
	Total     int64             `json:"total"`
	Recorded  int64             `json:"recorded"`
	Evidence  []FindingEvidence `json:"evidence"`
	SHA256    string            `json:"sha256"`
}

// TraversalAccounting records the actual bounded work performed by the
// recursive inspector. Ratios are exact integer ceiling ratios so canonical
// evidence never depends on floating-point rendering.
type TraversalAccounting struct {
	RawPayloadBytes          int64 `json:"raw_payload_bytes"`
	TotalEmittedBytes        int64 `json:"total_emitted_bytes"`
	ManifestedEmittedBytes   int64 `json:"manifested_emitted_bytes"`
	UnmanifestedEmittedBytes int64 `json:"unmanifested_emitted_bytes"`
	ContainerCount           int64 `json:"container_count"`
	EntryCount               int64 `json:"entry_count"`
	ManifestedEntryCount     int64 `json:"manifested_entry_count"`
	UnmanifestedEntryCount   int64 `json:"unmanifested_entry_count"`
	MaxObservedArchiveDepth  int64 `json:"max_observed_archive_depth"`
	MaxObservedLeafBytes     int64 `json:"max_observed_leaf_bytes"`
	MaxManifestedLeafBytes   int64 `json:"max_manifested_leaf_bytes"`
	MaxUnmanifestedLeafBytes int64 `json:"max_unmanifested_leaf_bytes"`
	MaxStreamInputBytes      int64 `json:"max_stream_input_bytes"`
	MaxStreamEmittedBytes    int64 `json:"max_stream_emitted_bytes"`
	MaxStreamExpansionRatio  int64 `json:"max_stream_expansion_ratio"`
	AggregateExpansionRatio  int64 `json:"aggregate_expansion_ratio"`
}

// TraversalFailureEvidence binds work that could not become manifest nodes to
// the exact structural diagnostic that stopped traversal. The all-zero value
// is required when no unmanifested work occurred.
type TraversalFailureEvidence struct {
	Code                       DiagnosticCode `json:"code"`
	Path                       string         `json:"path"`
	Reason                     string         `json:"reason"`
	UnmanifestedEntryCount     int64          `json:"unmanifested_entry_count"`
	UnmanifestedEmittedBytes   int64          `json:"unmanifested_emitted_bytes"`
	MaxUnmanifestedLeafBytes   int64          `json:"max_unmanifested_leaf_bytes"`
	MaxObservedStreamInput     int64          `json:"max_observed_stream_input"`
	MaxObservedStreamEmitted   int64          `json:"max_observed_stream_emitted"`
	MaxObservedStreamExpansion int64          `json:"max_observed_stream_expansion"`
}

// RawPayloadEvidence identifies the exact captured root bytes or canonical
// captured tree.
type RawPayloadEvidence struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
	Kind   string `json:"kind"`
}

// Manifest is canonical artifact-manifest-v1 evidence. ManifestDigest is
// SHA-256 over the same manifest with manifest_digest set to the empty string,
// removing any self-reference while binding every other field.
type Manifest struct {
	SchemaID           string                   `json:"schema_id"`
	PolicyID           string                   `json:"policy_id"`
	PolicyVersion      int                      `json:"policy_version"`
	LimitVector        LimitVector              `json:"limit_vector"`
	DetectorRegistryID string                   `json:"detector_registry_id"`
	Detectors          []DetectorIdentity       `json:"detectors"`
	AdapterID          string                   `json:"adapter_id"`
	ProfileID          ProfileID                `json:"profile_id"`
	Manager            string                   `json:"manager"`
	PackageName        string                   `json:"package_name"`
	PackageVersion     string                   `json:"package_version"`
	Origin             OriginEvidence           `json:"origin"`
	TrustRole          TrustRole                `json:"trust_role"`
	RoleEvidence       []Fact                   `json:"role_evidence"`
	RawPayload         RawPayloadEvidence       `json:"raw_payload"`
	Accounting         TraversalAccounting      `json:"accounting"`
	TraversalFailure   TraversalFailureEvidence `json:"traversal_failure"`
	Nodes              []ManifestNode           `json:"nodes"`
	Diagnostics        []Diagnostic             `json:"diagnostics"`
	Findings           FindingsSummary          `json:"findings"`
	Decision           Decision                 `json:"decision"`
	ManifestDigest     string                   `json:"manifest_digest"`
}

// Result contains reusable canonical evidence for both admitted and rejected
// inputs. Admission returns a non-nil token only for a role-valid allow result.
type Result struct {
	Manifest       Manifest
	CanonicalBytes []byte
	Admission      *Admission
}

// Admission is an unforgeable-in-practice role token: its authorizing fields
// are unexported and are set only after canonical policy evidence is sealed.
// Its zero value never authorizes an action.
type Admission struct {
	role      TrustRole
	decision  Decision
	digest    string
	schema    string
	policy    string
	seal      *authorizationSeal
	toolchain *selectedToolchainState
}

// Role returns the causal role authorized by this token.
func (admission *Admission) Role() TrustRole {
	if admission == nil {
		return ""
	}
	return admission.role
}

// ManifestDigest returns the exact canonical evidence identity.
func (admission *Admission) ManifestDigest() string {
	if admission == nil {
		return ""
	}
	return admission.digest
}

// PolicyError reports a policy rejection while Result retains its canonical
// manifest evidence.
type PolicyError struct {
	Primary Diagnostic
}

func (err *PolicyError) Error() string {
	if err == nil {
		return ""
	}
	if err.Primary.Path == "" {
		return string(err.Primary.Code)
	}
	return fmt.Sprintf("%s: %s", err.Primary.Code, err.Primary.Path)
}

// ErrorCode returns a stable diagnostic code from a policy error.
func ErrorCode(err error) DiagnosticCode {
	policyErr, ok := err.(*PolicyError)
	if !ok || policyErr == nil {
		return ""
	}
	return policyErr.Primary.Code
}

// Service owns the immutable v1 classifier, detector registry, and limits.
// It has no adapter callbacks that could widen the central policy.
type Service struct {
	limits            LimitVector
	detectors         []DetectorIdentity
	authorizationSeal *authorizationSeal
}

type authorizationSeal struct {
	marker uint64
}

var managerAuthorizationSeal = &authorizationSeal{marker: 0x63757261746f7221}

type toolchainAuthorizationRecord struct {
	seal                        *authorizationSeal
	policySelector              string
	resolvedRoot                string
	executableRelativePath      string
	environmentSearchResolution string
	version                     string
	platform                    string
	fingerprintAlgorithm        string
	checkpointFingerprintSHA256 string
	timeOfUseFingerprintSHA256  string
	payloadPath                 string
	payloadSHA256               string
	payloadSize                 int64
	outsideDependencyClosure    bool
	containedLinksValidated     bool
	ordinaryNodesValidated      bool
}

type sealedToolchainAuthorization struct {
	record toolchainAuthorizationRecord
}

func (authorization sealedToolchainAuthorization) artifactPolicyToolchainAuthorization() toolchainAuthorizationRecord {
	return authorization.record
}

type localOutputAuthorizationRecord struct {
	seal                            *authorizationSeal
	sourceClosureDigest             string
	artifactManifestDigest          string
	buildPlanDigest                 string
	declaredActionID                string
	executionReceiptSHA256          string
	protectedReceiptSHA256          string
	protectedStoreIdentity          string
	stagingRootIdentity             string
	payloadPath                     string
	payloadSHA256                   string
	payloadSize                     int64
	expectation                     ArtifactExpectation
	stagingStartedEmpty             bool
	observedProduction              bool
	writeSetMatched                 bool
	preexistingInputExcluded        bool
	hardlinkSourceExcluded          bool
	expectationIndependentlyDerived bool
	completeInputMatched            bool
	protectedPublicationValidated   bool
}

type sealedLocalOutputAuthorization struct {
	record localOutputAuthorizationRecord
}

func (authorization sealedLocalOutputAuthorization) artifactPolicyLocalOutputAuthorization() localOutputAuthorizationRecord {
	return authorization.record
}

// NewService constructs the sole production policy-v1 service. Its closed
// selector can derive external-toolchain authority; local-output authority
// remains unavailable until the protected-executor integration exists.
func NewService() *Service {
	return &Service{
		limits: DefaultLimits(), detectors: detectorIdentities(),
		authorizationSeal: managerAuthorizationSeal,
	}
}

func configuredService(service *Service) (*Service, error) {
	if service == nil || (service.limits == (LimitVector{}) && len(service.detectors) == 0) {
		return NewService(), nil
	}
	if service.limits != DefaultLimits() {
		return nil, fmt.Errorf("artifact policy service has a non-production limit vector")
	}
	if !slices.Equal(service.detectors, detectorIdentities()) {
		return nil, fmt.Errorf("artifact policy service has a non-production detector registry")
	}
	if service.authorizationSeal == nil {
		return nil, fmt.Errorf("artifact policy service has no manager authorization authority")
	}
	return service, nil
}

// AuthorizeAdapterExecution verifies that every dependency token was admitted
// as input and the independently selected toolchain token was allowed. Callers
// invoke this immediately before any manager, hook, generator, or compiler.
func AuthorizeAdapterExecution(dependencies []*Admission, toolchain *Admission) error {
	if toolchain != nil && toolchain.toolchain != nil {
		_, err := authorizeSelectedToolchainExecution(context.Background(), toolchain.toolchain, dependencies, toolchain)
		return err
	}
	return authorizeAdmissionRoles(dependencies, toolchain)
}

func authorizeAdmissionRoles(dependencies []*Admission, toolchain *Admission) error {
	if toolchain == nil || toolchain.seal == nil {
		return fmt.Errorf("external toolchain admission is absent or invalid")
	}
	for index, dependency := range dependencies {
		if !validAdmission(dependency, RoleDependencyInput, DecisionAdmitInput) || dependency.seal != toolchain.seal {
			return fmt.Errorf("dependency admission %d is absent or invalid", index)
		}
	}
	if !validAdmission(toolchain, RoleExternalToolchain, DecisionAllowToolchain) {
		return fmt.Errorf("external toolchain admission is absent or invalid")
	}
	return nil
}

// AuthorizeCachePublication permits publication only for a causally receipted
// local output token.
func AuthorizeCachePublication(output *Admission) error {
	if !validAdmission(output, RoleLocalBuildOutput, DecisionAllowOutput) {
		return fmt.Errorf("local output admission is absent or invalid")
	}
	return nil
}

func validAdmission(admission *Admission, role TrustRole, decision Decision) bool {
	return admission != nil && admission.role == role && admission.decision == decision &&
		admission.schema == SchemaID && admission.policy == PolicyID && admission.digest != "" && admission.seal != nil
}

func closedSet[T ~string](values ...T) map[T]struct{} {
	result := make(map[T]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func contextError(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}
