package closuregraph

import "context"

// CaptureAdapter emits only selection-neutral capture records. Exact platform
// and external-toolchain records are intentionally outside this interface.
type CaptureAdapter interface {
	Capture(context.Context) (CaptureGraph, []Node, []Edge, error)
}

// SelectionAdapter owns ecosystem-specific selection evaluation and emits the
// only permitted binding overlay plus its versioned condition evaluators.
type SelectionAdapter interface {
	Bind(context.Context, CaptureGraph, SelectionContext) (SelectionBinding, []Node, []Edge, BindingAuthority, []ConditionEvaluator, error)
}

// CheckpointResult is one successful stage payload and its deterministic
// diagnostics before the shared chain codec assigns a checkpoint identity.
type CheckpointResult[P CheckpointPayload] struct {
	Payload     P
	Diagnostics []Diagnostic
}

// C0Profiler binds policy, requested selection, platforms, capabilities, and
// all evidence-derivation tools before any process may run.
type C0Profiler interface {
	Profile(context.Context) (CheckpointResult[C0ProfilePayload], error)
}

// C1Resolver produces frozen declaration, lock, condition, candidate, and
// intake/derivation-journal evidence.
type C1Resolver interface {
	Resolve(context.Context, Checkpoint) (CheckpointResult[C1ResolvePayload], error)
}

// C2Capturer produces the complete immutable capture aggregate.
type C2Capturer interface {
	CaptureInputs(context.Context, Checkpoint) (CheckpointResult[C2CapturePayload], error)
}

// C3Admitter produces recursive artifact-admission evidence. Cargo adapters
// may additionally emit origin and derived C3 payloads through the same type.
type C3Admitter interface {
	Admit(context.Context, Checkpoint) (CheckpointResult[C3AdmitPayload], error)
}

// C4Closer reconciles capture, selection, binding, and active graph evidence.
type C4Closer interface {
	Close(context.Context, Checkpoint) (GraphBundle, CheckpointResult[C4ClosePayload], error)
}

// C5Planner derives the immutable acyclic build plan.
type C5Planner interface {
	Plan(context.Context, Checkpoint, GraphBundle) (BuildPlan, CheckpointResult[C5PlanPayload], error)
}

// C6OfflineExecutor supplies the separately receipted offline observation
// result. Sandbox implementation belongs to the protected executor task.
type C6OfflineExecutor interface {
	ExecuteOffline(context.Context, Checkpoint, BuildPlan) (ExecutionReceipt, []ProducedArtifactObservation, CheckpointResult[C6OfflinePayload], error)
}

// C7Publisher supplies the protected publication result. Storage and atomic
// publication behavior remain outside this schema package.
type C7Publisher interface {
	Publish(context.Context, Checkpoint, ExecutionReceipt) (PublicationReceipt, CheckpointResult[C7PublishPayload], error)
}
