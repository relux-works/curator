package swiftpmbuild

import (
	"context"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpminterop"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// ProfileID reuses the SwiftPM source profile: planning, building, and
// publishing are later phases of one adapter, not a second profile that could
// relax the accepted rules.
const ProfileID = swiftpmsource.ProfileID

// PlanSchemaID identifies the canonical SwiftPM build-plan command record.
const PlanSchemaID = "swiftpm-build-command-v1"

// ExecutablePayloadSchemaID identifies the exact declared product payload.
const ExecutablePayloadSchemaID = "swiftpm-native-executable-v1"

// ToolSlot is one exact build tool or component slot. Every slot must resolve
// exactly once against the accepted C4 binding overlay before planning.
type ToolSlot string

// The closed swiftpm-source-v1 build slot set.
const (
	// SlotSwiftPM is the SwiftPM build driver that plans and runs the build.
	SlotSwiftPM ToolSlot = "swiftpm"
	// SlotSwiftCompiler is the exact swiftc identity SwiftPM must invoke.
	SlotSwiftCompiler ToolSlot = "swiftc"
	// SlotPackageDescription is the manifest runtime bound at C0.
	SlotPackageDescription ToolSlot = "package-description"
	// SlotClang is the exact C-family driver identity.
	SlotClang ToolSlot = "clang"
	// SlotClangCXX is the exact C++-family driver identity. It is required
	// only when a selected target carries C++ or Objective-C++ source.
	SlotClangCXX ToolSlot = "clang++"
	// SlotLinker is the exact linker identity that produces the product.
	SlotLinker ToolSlot = "linker"
	// SlotSDK is the exact selected SDK component.
	SlotSDK ToolSlot = "sdk"
)

// requiredSlots is the closed set every accepted build binding must resolve.
var requiredSlots = []ToolSlot{SlotClang, SlotLinker, SlotPackageDescription, SlotSDK, SlotSwiftCompiler, SlotSwiftPM}

// SlotBinding is one exactly resolved binding node for a build slot.
type SlotBinding struct {
	Slot    ToolSlot
	Role    string
	NodeID  closuregraph.ID
	Payload closuregraph.ToolchainComponentPayload
}

// Binding is the exact C4 overlay this stage consumes. It repeats no tool
// identity: every physical component resolves from canonical selection-binding
// nodes and typed edges published by the accepted upstream stages.
type Binding struct {
	PlatformNodeID closuregraph.ID
	Platform       closuregraph.TargetPlatformPayload
	ProductNodeID  closuregraph.ID
	Slots          map[ToolSlot]SlotBinding
}

// ObjectSlot is one declared intermediate object write slot. The accepted
// interop closure declares exactly one slot per selected compile source; the
// offline build must materialize every one of them from real produced bytes.
type ObjectSlot struct {
	Package, Target string
	// Source is the package-relative source path this object is compiled from.
	Source string
	// SourceRoot is the target's package-relative source root. SwiftPM mirrors
	// the tree below it into the target build directory, so the produced
	// object is reconciled against the source path relative to this root.
	SourceRoot     string
	Kind           string
	Path           string
	NodeID         closuregraph.ID
	ActionNodeID   closuregraph.ID
	ProducesEdgeID closuregraph.ID
}

// Command is the exact committed SwiftPM build invocation.
type Command struct {
	Executable  string
	CWD         string
	Argv        []string
	Environment map[string]string
}

// Plan is the immutable C5 result: an extended C4 closure that adds only the
// product link action and its expected output, plus the exact deterministic
// action DAG, command, and publication authority derived from it.
type Plan struct {
	Binding          Binding
	Graph            closuregraph.GraphBundle
	C4, C5           closuregraph.Checkpoint
	BuildPlan        closuregraph.BuildPlan
	Closure          closuregraph.SourceClosure
	Expected         closuregraph.ExpectedCacheInput
	Publication      closuregraph.PublicationEvidence
	Command          Command
	CommandID        closuregraph.ID
	LinkActionNodeID closuregraph.ID
	OutputNodeID     closuregraph.ID
	OutputPath       string
	Objects          []ObjectSlot
	ScratchDirectory string
	capture          *swiftpmsource.Capture
	interop          *swiftpminterop.Result
}

// Result is the detached C6/C7 evidence plus the protected product path.
type Result struct {
	CacheHit          bool
	ArtifactPath      string
	ActiveGraphID     closuregraph.ID
	CommandID         closuregraph.ID
	Execution         closuregraph.ExecutionReceipt
	Publication       closuregraph.PublicationReceipt
	Observations      []closuregraph.ProducedArtifactObservation
	AssuredCacheInput closuregraph.ID
	ReadSet           []string
	WriteSet          []string
	C6, C7            closuregraph.Checkpoint
}

// Config supplies only central services and the exact selected linker and
// output identities. Every other physical identity is resolved from the
// accepted upstream capture and interop binding overlay.
type Config struct {
	Executor      *closureexec.Executor
	Store         *closureexec.CaptureStore
	Policy        *artifactpolicy.Service
	ExecutionRoot string
	OutputRoot    string
	StoreRoot     string
	// Linker is the exact selected linker identity. It is the only physical
	// component the build stage itself selects; every other slot must already
	// exist in the accepted C4 binding overlay.
	Linker swiftpminterop.ExternalComponent
	// Slots names the exact accepted binding role each build slot must resolve
	// to. It carries selection labels only: every physical identity is read
	// back from the accepted binding node, never restated here.
	Slots map[ToolSlot]string
	// Configuration is the exact SwiftPM build configuration.
	Configuration string
	// Assurance is the shared execution assurance mode.
	Assurance        closureexec.AssuranceMode
	AllowedProcesses []string
	Recheck          func(context.Context, swiftpmsource.ToolIdentity) (closureexec.ToolchainIdentity, error)
	CausalHead       string
}
