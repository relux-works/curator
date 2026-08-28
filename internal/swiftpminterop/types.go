package swiftpminterop

import (
	"context"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// ProfileID reuses the SwiftPM source profile: the interop stage is a later
// phase of one adapter, not a second profile that could relax its rules.
const ProfileID = swiftpmsource.ProfileID

// InteropSchemaID identifies the canonical interop evidence record.
const InteropSchemaID = "swiftpm-interop-closure-v1"

// ExternalComponent is one Curator-selected external toolchain, SDK, sysroot,
// or system-library binding. It always lives outside the dependency closure
// and never gains trust from package metadata or a provider hint.
type ExternalComponent struct {
	Role                   string
	ExecutableRelativePath string
	PlatformABI            string
	PolicySelector         string
	VersionOutput          string
	Fingerprint            closuregraph.ID
	ExecutableSHA256       closuregraph.ID
	SDKFactsDigest         closuregraph.ID
	Roots                  []string
	Modules                []string
	Libraries              []string
	Frameworks             []string
}

// SystemLibrary binds one declared .systemLibrary target to exactly one
// selected external component and the module map already inside it.
type SystemLibrary struct {
	Package, Target string
	ModuleMapPath   string
	Component       ExternalComponent
}

// PlatformProfile is the accepted destination profile for restricted
// languages. An unlisted triple, or a language whose gate is false, has no
// conformance evidence and therefore no support.
type PlatformProfile struct {
	ID                string
	TargetTriples     []string
	CxxInterop        bool
	ObjectiveCRuntime string
	CxxStandardModes  []string
}

// ObservedRead is one compiler-observed header, module, framework, library, or
// SDK read supplied by a verified provider.
type ObservedRead struct{ Path, Class string }

// ReadSetRequest asks a provider for the exact read set of one capture target.
type ReadSetRequest struct {
	Package, Target  string
	Languages        []string
	Sources          []string
	PublicHeaderRoot string
	ModuleMap        string
}

// ReadSetResult is provider-issued read evidence. Observed must be false when
// the provider cannot actually observe reads; the adapter then records
// not-observed rather than claiming coverage it does not have.
type ReadSetResult struct {
	Observed  bool
	Reads     []ObservedRead
	ReceiptID closuregraph.ID
}

// ReadSetProvider is the manager-owned seam that returns compiler-observed
// reads. This package never starts a compiler itself.
type ReadSetProvider interface {
	ObserveReads(context.Context, ReadSetRequest) (ReadSetResult, error)
}

// ReadSetEvidence is the portable-honest aggregate read verdict.
type ReadSetEvidence struct {
	Mode        string
	ReceiptIDs  []closuregraph.ID
	Resolutions []Resolution
}

// Config supplies only exact selected external evidence plus the shared
// assurance mode. Every field is verified before any target is admitted.
type Config struct {
	Clang           swiftpmsource.ToolIdentity
	ClangCXX        swiftpmsource.ToolIdentity
	SDK             ExternalComponent
	Sysroot         *ExternalComponent
	SystemLibraries []SystemLibrary
	Profile         PlatformProfile
	Assurance       closureexec.AssuranceMode
	ReadSets        ReadSetProvider
	Recheck         func(context.Context, swiftpmsource.ToolIdentity) (closureexec.ToolchainIdentity, error)
}

// TargetInterop is the exact classification and evidence of one declared
// capture target. Selected records whether the exact destination binding keeps
// the target; capture holds every declared target either way.
type TargetInterop struct {
	Selected        bool
	Package, Target string
	Kind            TargetKind
	Languages       []Language
	Sources         []string
	// SourceRoot is the package-relative directory SwiftPM enumerates this
	// target's sources from, with the documented convention default already
	// applied. The native build system mirrors the tree below it into the
	// target build directory, so a consumer reconciling a produced object with
	// its declared source resolves the source against exactly this root.
	SourceRoot       string
	PublicHeaderRoot string
	Headers          []HeaderFile
	ModuleMap        *ModuleMapEvidence
	Includes         []IncludeReference
	// CxxInteropMode is the exact destination verdict for the opt-in.
	CxxInteropMode bool
	// CxxInteropDeclared is the condition-neutral declaration of the opt-in.
	CxxInteropDeclared bool
	NodeID             closuregraph.ID
	ActionNodeID       closuregraph.ID
	// ObjectNodeIDs are the declared per-source object outputs of this
	// target's compile action, in source order.
	ObjectNodeIDs   []closuregraph.ID
	SourceSetNodeID closuregraph.ID
	HeaderSetNodeID closuregraph.ID
	ToolNodeID      closuregraph.ID
}

// Boundary is one explicit language boundary between two capture targets.
// Condition is the selection-neutral predicate the declaring edge carries;
// Selected is the exact destination verdict the active projection confirms.
type Boundary struct {
	Condition          *closuregraph.Condition
	Selected           bool
	Mode               closuregraph.InteropMode
	Provider, Consumer string
	ProviderLanguages  []Language
	ConsumerLanguages  []Language
	ABI                string
	Runtime            string
	InterfaceContract  string
	CallingConvention  string
	LinkLoadSemantics  string
	ToolchainRole      string
	NodeID             closuregraph.ID
}

// Result is the complete interop closure: exact per-target classification,
// declared boundaries, module-map and read evidence, and the republished
// capture graph, selection binding, active projection, and C4 checkpoint.
type Result struct {
	Targets        []TargetInterop
	Boundaries     []Boundary
	ModuleMaps     []ModuleMapEvidence
	Reads          ReadSetEvidence
	Graph          closuregraph.CaptureGraph
	Selection      closuregraph.SelectionContext
	Binding        closuregraph.SelectionBinding
	Active         closuregraph.ActiveGraph
	Records        closuregraph.RecordTables
	Authority      closuregraph.BindingAuthority
	C4             closuregraph.Checkpoint
	GraphDigest    closuregraph.ID
	EvidenceDigest closuregraph.ID
}
