package closuregraph

import (
	"fmt"
	"sort"
)

// EdgeKind is the closed curator-edge-v1 kind discriminator.
type EdgeKind string

const (
	// EdgeDeclares and the related constants enumerate the eleven closed
	// relationship kinds.
	EdgeDeclares EdgeKind = "declares"
	// EdgeResolvesTo binds a declaration to an immutable resolved instance.
	EdgeResolvesTo EdgeKind = "resolves_to"
	// EdgeRequires records a typed dependency relationship.
	EdgeRequires EdgeKind = "requires"
	// EdgeReads binds a declared action read slot.
	EdgeReads EdgeKind = "reads"
	// EdgeUsesTool binds a declared action tool slot.
	EdgeUsesTool EdgeKind = "uses_tool"
	// EdgeTargets binds a declared platform role to an exact platform.
	EdgeTargets EdgeKind = "targets"
	// EdgeProduces binds a declared action write slot.
	EdgeProduces EdgeKind = "produces"
	// EdgeProvidesInterop attaches the provider side of an interop boundary.
	EdgeProvidesInterop EdgeKind = "provides_interop"
	// EdgeConsumesInterop attaches the consumer side of an interop boundary.
	EdgeConsumesInterop EdgeKind = "consumes_interop"
	// EdgeInvokes records a declared runtime process relationship.
	EdgeInvokes EdgeKind = "invokes"
	// EdgePublishes binds an expected artifact to a command product.
	EdgePublishes EdgeKind = "publishes"
)

// RequirementScope is the closed dependency scope of a requires edge.
type RequirementScope string

const (
	// ScopeRuntime and the related constants enumerate closed requires scopes.
	ScopeRuntime RequirementScope = "runtime"
	// ScopeBuild identifies a build-time dependency.
	ScopeBuild RequirementScope = "build"
	// ScopeDevelopment identifies an inactive development dependency.
	ScopeDevelopment RequirementScope = "development"
	// ScopePeer identifies a peer-context dependency.
	ScopePeer RequirementScope = "peer"
	// ScopeOptional identifies an optional dependency declaration.
	ScopeOptional RequirementScope = "optional"
	// ScopeWorkspace identifies a workspace-local dependency.
	ScopeWorkspace RequirementScope = "workspace"
	// ScopeTool identifies a locally produced tool dependency.
	ScopeTool RequirementScope = "tool"
	// ScopeToolchain identifies an external toolchain dependency.
	ScopeToolchain RequirementScope = "toolchain"
	// ScopePackageArtifact identifies an immutable package payload relationship.
	ScopePackageArtifact RequirementScope = "package_artifact"
)

// Condition is an unevaluated conditional declaration. Selection state and
// evaluation results live only in ActiveGraph.
type Condition struct {
	EvaluatorID string
	Expression  string
}

// Validate checks a closed unevaluated condition declaration.
func (condition Condition) Validate() error {
	if err := validatePortableText(condition.EvaluatorID, "condition evaluator_id", false); err != nil {
		return err
	}
	return validatePortableText(condition.Expression, "condition expression", false)
}

func (condition Condition) value() map[string]any {
	return map[string]any{"evaluator_id": condition.EvaluatorID, "expression": condition.Expression}
}

func decodeConditionObject(raw map[string]any) (Condition, error) {
	if err := exactFields(raw, "condition", []string{"evaluator_id", "expression"}, nil); err != nil {
		return Condition{}, err
	}
	evaluator, err := requiredString(raw, "evaluator_id", "condition")
	if err != nil {
		return Condition{}, err
	}
	expression, err := requiredString(raw, "expression", "condition")
	if err != nil {
		return Condition{}, err
	}
	condition := Condition{EvaluatorID: evaluator, Expression: expression}
	return condition, condition.Validate()
}

// EvidenceOrigin identifies the exact declaration field that supplied a
// semantic record. ManifestDigest is optional only for synthetic selection
// fields such as selection.platform_roles.target.
type EvidenceOrigin struct {
	Field          string
	ManifestDigest ID
}

func (origin EvidenceOrigin) validate() error {
	if err := validatePortableText(origin.Field, "origin field", false); err != nil {
		return err
	}
	if origin.ManifestDigest != "" {
		return validateID(origin.ManifestDigest, "origin manifest_digest")
	}
	return nil
}

func (origin EvidenceOrigin) value() map[string]any {
	result := map[string]any{"field": origin.Field}
	if origin.ManifestDigest != "" {
		result["manifest_digest"] = string(origin.ManifestDigest)
	}
	return result
}

func decodeOrigin(raw map[string]any, context string) (EvidenceOrigin, error) {
	if err := exactFields(raw, context, []string{"field"}, []string{"manifest_digest"}); err != nil {
		return EvidenceOrigin{}, err
	}
	field, err := requiredString(raw, "field", context)
	if err != nil {
		return EvidenceOrigin{}, err
	}
	digest, err := optionalString(raw, "manifest_digest", context)
	if err != nil {
		return EvidenceOrigin{}, err
	}
	origin := EvidenceOrigin{Field: field, ManifestDigest: ID(digest)}
	return origin, origin.validate()
}

// Edge is one typed relationship. Selection state is deliberately absent.
type Edge struct {
	Kind       EdgeKind
	EdgeKey    string
	FromNodeID ID
	ToNodeID   ID
	Payload    EdgePayload
}

// EdgePayload is implemented only by the eleven closed edge payloads.
type EdgePayload interface {
	edgePayload()
	kind() EdgeKind
	validate() error
	value() map[string]any
	condition() *Condition
}

// DeclaresPayload records a manifest declaration relationship.
type DeclaresPayload struct{ Origin EvidenceOrigin }

// ResolvesToPayload records immutable lock/origin/artifact mapping.
type ResolvesToPayload struct {
	LockField          string
	Origin             EvidenceOrigin
	Checksum           string
	ArtifactManifestID ID
}

// RequiresPayload records dependency scope and optional unevaluated condition.
type RequiresPayload struct {
	Scope          RequirementScope
	Condition      *Condition
	Origin         EvidenceOrigin
	DependencyKind string
}

// ReadsPayload binds an action or target read slot.
type ReadsPayload struct {
	Path       string
	ReadSlot   string
	ReadClass  string
	Projection []string
}

// UsesToolPayload binds an action's named executable slot.
type UsesToolPayload struct {
	ExecutableRelativePath string
	ToolSlot               string
	InvocationRole         string
}

// TargetsPayload binds a selected node to an exact platform role.
type TargetsPayload struct {
	BindingRole PlatformRole
	Origin      EvidenceOrigin
}

// ProducesPayload binds an action's named output slot.
type ProducesPayload struct {
	Path       string
	WriteSlot  string
	WriteClass string
}

// ProvidesInteropPayload attaches a provider to an explicit boundary.
type ProvidesInteropPayload struct {
	Origin      EvidenceOrigin
	EvidenceIDs []ID
	ExportRole  string
	LinkMode    string
}

// ConsumesInteropPayload attaches a consumer to an explicit boundary.
type ConsumesInteropPayload struct {
	Origin         EvidenceOrigin
	Use            string
	ABIExpectation string
}

// InvokesPayload records a runtime executable/protocol relationship.
type InvokesPayload struct {
	ProtocolSchema       string
	ExecutableResolution string
	ArgumentsContract    string
	EnvironmentContract  string
	WorkingDirectory     string
}

// PublishesPayload maps a product entry point to an expected output.
type PublishesPayload struct {
	Destination string
	EntryPoint  string
}

func (DeclaresPayload) edgePayload()        {}
func (ResolvesToPayload) edgePayload()      {}
func (RequiresPayload) edgePayload()        {}
func (ReadsPayload) edgePayload()           {}
func (UsesToolPayload) edgePayload()        {}
func (TargetsPayload) edgePayload()         {}
func (ProducesPayload) edgePayload()        {}
func (ProvidesInteropPayload) edgePayload() {}
func (ConsumesInteropPayload) edgePayload() {}
func (InvokesPayload) edgePayload()         {}
func (PublishesPayload) edgePayload()       {}

func (DeclaresPayload) kind() EdgeKind        { return EdgeDeclares }
func (ResolvesToPayload) kind() EdgeKind      { return EdgeResolvesTo }
func (RequiresPayload) kind() EdgeKind        { return EdgeRequires }
func (ReadsPayload) kind() EdgeKind           { return EdgeReads }
func (UsesToolPayload) kind() EdgeKind        { return EdgeUsesTool }
func (TargetsPayload) kind() EdgeKind         { return EdgeTargets }
func (ProducesPayload) kind() EdgeKind        { return EdgeProduces }
func (ProvidesInteropPayload) kind() EdgeKind { return EdgeProvidesInterop }
func (ConsumesInteropPayload) kind() EdgeKind { return EdgeConsumesInterop }
func (InvokesPayload) kind() EdgeKind         { return EdgeInvokes }
func (PublishesPayload) kind() EdgeKind       { return EdgePublishes }

func (DeclaresPayload) condition() *Condition         { return nil }
func (ResolvesToPayload) condition() *Condition       { return nil }
func (payload RequiresPayload) condition() *Condition { return payload.Condition }
func (ReadsPayload) condition() *Condition            { return nil }
func (UsesToolPayload) condition() *Condition         { return nil }
func (TargetsPayload) condition() *Condition          { return nil }
func (ProducesPayload) condition() *Condition         { return nil }
func (ProvidesInteropPayload) condition() *Condition  { return nil }
func (ConsumesInteropPayload) condition() *Condition  { return nil }
func (InvokesPayload) condition() *Condition          { return nil }
func (PublishesPayload) condition() *Condition        { return nil }

// Validate checks the closed edge schema independent of its endpoint table.
func (edge Edge) Validate() error {
	if !validEdgeKind(edge.Kind) {
		return fmt.Errorf("%s: unsupported edge kind %q", CodeGraphSchemaUnsupported, edge.Kind)
	}
	if err := validatePortableText(edge.EdgeKey, "edge edge_key", false); err != nil {
		return err
	}
	if err := validateID(edge.FromNodeID, "edge from_node_id"); err != nil {
		return err
	}
	if err := validateID(edge.ToNodeID, "edge to_node_id"); err != nil {
		return err
	}
	if edge.Payload == nil {
		return fmt.Errorf("edge payload is required")
	}
	payloadKind, canonical := edgePayloadValueKind(edge.Payload)
	if !canonical {
		return fmt.Errorf("%s: edge payload for kind %q must use its canonical value representation, got %T", CodeGraphSchemaUnsupported, edge.Kind, edge.Payload)
	}
	if payloadKind != edge.Kind {
		return fmt.Errorf("edge kind %q does not match payload kind %q", edge.Kind, payloadKind)
	}
	return edge.Payload.validate()
}

// edgePayloadValueKind recognizes only the exact value representations that
// DecodeEdge returns. Pointer forms satisfy EdgePayload through value-receiver
// methods, but are not a second supported canonical record representation.
func edgePayloadValueKind(payload EdgePayload) (EdgeKind, bool) {
	switch payload.(type) {
	case DeclaresPayload:
		return EdgeDeclares, true
	case ResolvesToPayload:
		return EdgeResolvesTo, true
	case RequiresPayload:
		return EdgeRequires, true
	case ReadsPayload:
		return EdgeReads, true
	case UsesToolPayload:
		return EdgeUsesTool, true
	case TargetsPayload:
		return EdgeTargets, true
	case ProducesPayload:
		return EdgeProduces, true
	case ProvidesInteropPayload:
		return EdgeProvidesInterop, true
	case ConsumesInteropPayload:
		return EdgeConsumesInterop, true
	case InvokesPayload:
		return EdgeInvokes, true
	case PublishesPayload:
		return EdgePublishes, true
	default:
		return "", false
	}
}

// CanonicalBytes returns exact curator-edge-v1 CCJ bytes.
func (edge Edge) CanonicalBytes() ([]byte, error) { return canonicalBytes(edge) }

// ID derives the domain-separated curator-edge-v1 identity.
func (edge Edge) ID() (ID, error) { return recordID(edge) }

func (edge Edge) domainLabel() string { return LabelEdge }
func (edge Edge) canonicalValue() map[string]any {
	return map[string]any{"edge_key": edge.EdgeKey, "from_node_id": string(edge.FromNodeID), "kind": string(edge.Kind), "payload": edge.Payload.value(), "to_node_id": string(edge.ToNodeID)}
}

func validEdgeKind(kind EdgeKind) bool {
	switch kind {
	case EdgeDeclares, EdgeResolvesTo, EdgeRequires, EdgeReads, EdgeUsesTool,
		EdgeTargets, EdgeProduces, EdgeProvidesInterop, EdgeConsumesInterop,
		EdgeInvokes, EdgePublishes:
		return true
	default:
		return false
	}
}

func validRequirementScope(scope RequirementScope) bool {
	switch scope {
	case ScopeRuntime, ScopeBuild, ScopeDevelopment, ScopePeer, ScopeOptional,
		ScopeWorkspace, ScopeTool, ScopeToolchain, ScopePackageArtifact:
		return true
	default:
		return false
	}
}

func (payload DeclaresPayload) validate() error { return payload.Origin.validate() }
func (payload DeclaresPayload) value() map[string]any {
	return map[string]any{"origin": payload.Origin.value()}
}

func (payload ResolvesToPayload) validate() error {
	if err := validatePortableText(payload.LockField, "lock_field", false); err != nil {
		return err
	}
	if err := payload.Origin.validate(); err != nil {
		return err
	}
	if err := validatePortableText(payload.Checksum, "checksum", false); err != nil {
		return err
	}
	return validateID(payload.ArtifactManifestID, "artifact_manifest_id")
}
func (payload ResolvesToPayload) value() map[string]any {
	return map[string]any{"lock_field": payload.LockField, "origin": payload.Origin.value(), "checksum": payload.Checksum, "artifact_manifest_id": string(payload.ArtifactManifestID)}
}

func (payload RequiresPayload) validate() error {
	if !validRequirementScope(payload.Scope) {
		return fmt.Errorf("unsupported requires scope %q", payload.Scope)
	}
	if payload.Condition != nil {
		if err := payload.Condition.Validate(); err != nil {
			return err
		}
	}
	if err := payload.Origin.validate(); err != nil {
		return err
	}
	if payload.DependencyKind != "" {
		return validatePortableText(payload.DependencyKind, "dependency_kind", false)
	}
	return nil
}
func (payload RequiresPayload) value() map[string]any {
	value := map[string]any{"scope": string(payload.Scope), "origin": payload.Origin.value()}
	if payload.Condition != nil {
		value["condition"] = payload.Condition.value()
	}
	if payload.DependencyKind != "" {
		value["dependency_kind"] = payload.DependencyKind
	}
	return value
}

func (payload ReadsPayload) validate() error {
	if err := validatePortablePath(payload.Path, "reads path"); err != nil {
		return err
	}
	if err := validatePortableText(payload.ReadSlot, "read_slot", false); err != nil {
		return err
	}
	if payload.ReadClass != "" {
		if err := validatePortableText(payload.ReadClass, "read_class", false); err != nil {
			return err
		}
	}
	if payload.Projection != nil {
		for index, path := range payload.Projection {
			if err := validatePortablePath(path, fmt.Sprintf("projection[%d]", index)); err != nil {
				return err
			}
		}
		return validateUniqueStrings(payload.Projection, "projection", true)
	}
	return nil
}
func (payload ReadsPayload) value() map[string]any {
	value := map[string]any{"path": payload.Path, "read_slot": payload.ReadSlot}
	if payload.ReadClass != "" {
		value["read_class"] = payload.ReadClass
	}
	if payload.Projection != nil {
		value["projection"] = stringsToAny(payload.Projection)
	}
	return value
}

func (payload UsesToolPayload) validate() error {
	if err := validatePortablePath(payload.ExecutableRelativePath, "executable_relative_path"); err != nil {
		return err
	}
	if err := validatePortableText(payload.ToolSlot, "tool_slot", false); err != nil {
		return err
	}
	if payload.InvocationRole != "" {
		return validatePortableText(payload.InvocationRole, "invocation_role", false)
	}
	return nil
}
func (payload UsesToolPayload) value() map[string]any {
	value := map[string]any{"executable_relative_path": payload.ExecutableRelativePath, "tool_slot": payload.ToolSlot}
	if payload.InvocationRole != "" {
		value["invocation_role"] = payload.InvocationRole
	}
	return value
}

func (payload TargetsPayload) validate() error {
	if payload.BindingRole != PlatformTarget && payload.BindingRole != PlatformHost {
		return fmt.Errorf("unsupported binding_role %q", payload.BindingRole)
	}
	return payload.Origin.validate()
}
func (payload TargetsPayload) value() map[string]any {
	return map[string]any{"binding_role": string(payload.BindingRole), "origin": payload.Origin.value()}
}

func (payload ProducesPayload) validate() error {
	if err := validatePortablePath(payload.Path, "produces path"); err != nil {
		return err
	}
	if err := validatePortableText(payload.WriteSlot, "write_slot", false); err != nil {
		return err
	}
	if payload.WriteClass != "" {
		return validatePortableText(payload.WriteClass, "write_class", false)
	}
	return nil
}
func (payload ProducesPayload) value() map[string]any {
	value := map[string]any{"path": payload.Path, "write_slot": payload.WriteSlot}
	if payload.WriteClass != "" {
		value["write_class"] = payload.WriteClass
	}
	return value
}

func (payload ProvidesInteropPayload) validate() error {
	if err := payload.Origin.validate(); err != nil {
		return err
	}
	if err := validateIDSlice(payload.EvidenceIDs, "evidence_ids", true); err != nil {
		return err
	}
	return validatePortableTextFields(map[string]string{"export_role": payload.ExportRole, "link_mode": payload.LinkMode}, false, false)
}
func (payload ProvidesInteropPayload) value() map[string]any {
	return map[string]any{"origin": payload.Origin.value(), "evidence_ids": idsToAny(payload.EvidenceIDs), "export_role": payload.ExportRole, "link_mode": payload.LinkMode}
}

func (payload ConsumesInteropPayload) validate() error {
	if err := payload.Origin.validate(); err != nil {
		return err
	}
	return validatePortableTextFields(map[string]string{"use": payload.Use, "abi_expectation": payload.ABIExpectation}, false, false)
}
func (payload ConsumesInteropPayload) value() map[string]any {
	return map[string]any{"origin": payload.Origin.value(), "use": payload.Use, "abi_expectation": payload.ABIExpectation}
}

func (payload InvokesPayload) validate() error {
	if err := validatePortableTextFields(map[string]string{"protocol_schema": payload.ProtocolSchema, "executable_resolution": payload.ExecutableResolution, "arguments_contract": payload.ArgumentsContract, "environment_contract": payload.EnvironmentContract}, false, false); err != nil {
		return err
	}
	if payload.WorkingDirectory != "" {
		return validatePortablePath(payload.WorkingDirectory, "working_directory")
	}
	return nil
}
func (payload InvokesPayload) value() map[string]any {
	value := map[string]any{"protocol_schema": payload.ProtocolSchema, "executable_resolution": payload.ExecutableResolution, "arguments_contract": payload.ArgumentsContract, "environment_contract": payload.EnvironmentContract}
	if payload.WorkingDirectory != "" {
		value["working_directory"] = payload.WorkingDirectory
	}
	return value
}

func (payload PublishesPayload) validate() error {
	if err := validatePortablePath(payload.Destination, "destination"); err != nil {
		return err
	}
	return validatePortableText(payload.EntryPoint, "entry_point", false)
}
func (payload PublishesPayload) value() map[string]any {
	return map[string]any{"destination": payload.Destination, "entry_point": payload.EntryPoint}
}

// DecodeEdge accepts exact canonical CCJ bytes and rejects unknown fields or
// unsupported payload kinds.
func DecodeEdge(payload []byte) (Edge, error) {
	raw, err := decodeCanonicalObject(payload, "edge")
	if err != nil {
		return Edge{}, err
	}
	if err := exactFields(raw, "edge", []string{"edge_key", "from_node_id", "kind", "payload", "to_node_id"}, nil); err != nil {
		return Edge{}, err
	}
	key, err := requiredString(raw, "edge_key", "edge")
	if err != nil {
		return Edge{}, err
	}
	from, err := requiredString(raw, "from_node_id", "edge")
	if err != nil {
		return Edge{}, err
	}
	kindValue, err := requiredString(raw, "kind", "edge")
	if err != nil {
		return Edge{}, err
	}
	to, err := requiredString(raw, "to_node_id", "edge")
	if err != nil {
		return Edge{}, err
	}
	payloadRaw, err := requiredObject(raw, "payload", "edge")
	if err != nil {
		return Edge{}, err
	}
	edgePayload, err := decodeEdgePayload(EdgeKind(kindValue), payloadRaw)
	if err != nil {
		return Edge{}, err
	}
	edge := Edge{Kind: EdgeKind(kindValue), EdgeKey: key, FromNodeID: ID(from), ToNodeID: ID(to), Payload: edgePayload}
	if err := edge.Validate(); err != nil {
		return Edge{}, err
	}
	if err := requireDecodedRecordRoundTrip(payload, edge); err != nil {
		return Edge{}, err
	}
	return edge, nil
}

func decodeEdgePayload(kind EdgeKind, raw map[string]any) (EdgePayload, error) {
	switch kind {
	case EdgeDeclares:
		if err := exactFields(raw, "declares payload", []string{"origin"}, nil); err != nil {
			return nil, err
		}
		originRaw, err := requiredObject(raw, "origin", "declares payload")
		if err != nil {
			return nil, err
		}
		origin, err := decodeOrigin(originRaw, "declares origin")
		return DeclaresPayload{Origin: origin}, err
	case EdgeResolvesTo:
		const context = "resolves_to payload"
		if err := exactFields(raw, context, []string{"lock_field", "origin", "checksum", "artifact_manifest_id"}, nil); err != nil {
			return nil, err
		}
		fields, err := decodeStringFields(raw, context, []string{"lock_field", "checksum", "artifact_manifest_id"}, nil)
		if err != nil {
			return nil, err
		}
		originRaw, err := requiredObject(raw, "origin", context)
		if err != nil {
			return nil, err
		}
		origin, err := decodeOrigin(originRaw, "resolves_to origin")
		if err != nil {
			return nil, err
		}
		return ResolvesToPayload{LockField: fields["lock_field"], Origin: origin, Checksum: fields["checksum"], ArtifactManifestID: ID(fields["artifact_manifest_id"])}, nil
	case EdgeRequires:
		const context = "requires payload"
		if err := exactFields(raw, context, []string{"scope", "origin"}, []string{"condition", "dependency_kind"}); err != nil {
			return nil, err
		}
		fields, err := decodeStringFields(raw, context, []string{"scope"}, []string{"dependency_kind"})
		if err != nil {
			return nil, err
		}
		originRaw, err := requiredObject(raw, "origin", context)
		if err != nil {
			return nil, err
		}
		origin, err := decodeOrigin(originRaw, "requires origin")
		if err != nil {
			return nil, err
		}
		p := RequiresPayload{Scope: RequirementScope(fields["scope"]), Origin: origin, DependencyKind: fields["dependency_kind"]}
		if conditionRaw, present, err := optionalObject(raw, "condition", context); err != nil {
			return nil, err
		} else if present {
			condition, err := decodeConditionObject(conditionRaw)
			if err != nil {
				return nil, err
			}
			p.Condition = &condition
		}
		return p, nil
	case EdgeReads:
		const context = "reads payload"
		if err := exactFields(raw, context, []string{"path", "read_slot"}, []string{"read_class", "projection"}); err != nil {
			return nil, err
		}
		fields, err := decodeStringFields(raw, context, []string{"path", "read_slot"}, []string{"read_class"})
		if err != nil {
			return nil, err
		}
		p := ReadsPayload{Path: fields["path"], ReadSlot: fields["read_slot"], ReadClass: fields["read_class"]}
		p.Projection, err = optionalStringSlice(raw, "projection", context)
		return p, err
	case EdgeUsesTool:
		const context = "uses_tool payload"
		if err := exactFields(raw, context, []string{"executable_relative_path", "tool_slot"}, []string{"invocation_role"}); err != nil {
			return nil, err
		}
		fields, err := decodeStringFields(raw, context, []string{"executable_relative_path", "tool_slot"}, []string{"invocation_role"})
		if err != nil {
			return nil, err
		}
		return UsesToolPayload{ExecutableRelativePath: fields["executable_relative_path"], ToolSlot: fields["tool_slot"], InvocationRole: fields["invocation_role"]}, nil
	case EdgeTargets:
		const context = "targets payload"
		if err := exactFields(raw, context, []string{"binding_role", "origin"}, nil); err != nil {
			return nil, err
		}
		fields, err := decodeStringFields(raw, context, []string{"binding_role"}, nil)
		if err != nil {
			return nil, err
		}
		originRaw, err := requiredObject(raw, "origin", context)
		if err != nil {
			return nil, err
		}
		origin, err := decodeOrigin(originRaw, "targets origin")
		return TargetsPayload{BindingRole: PlatformRole(fields["binding_role"]), Origin: origin}, err
	case EdgeProduces:
		const context = "produces payload"
		if err := exactFields(raw, context, []string{"path", "write_slot"}, []string{"write_class"}); err != nil {
			return nil, err
		}
		fields, err := decodeStringFields(raw, context, []string{"path", "write_slot"}, []string{"write_class"})
		if err != nil {
			return nil, err
		}
		return ProducesPayload{Path: fields["path"], WriteSlot: fields["write_slot"], WriteClass: fields["write_class"]}, nil
	case EdgeProvidesInterop:
		const context = "provides_interop payload"
		if err := exactFields(raw, context, []string{"origin", "evidence_ids", "export_role", "link_mode"}, nil); err != nil {
			return nil, err
		}
		fields, err := decodeStringFields(raw, context, []string{"export_role", "link_mode"}, nil)
		if err != nil {
			return nil, err
		}
		originRaw, err := requiredObject(raw, "origin", context)
		if err != nil {
			return nil, err
		}
		origin, err := decodeOrigin(originRaw, "provides_interop origin")
		if err != nil {
			return nil, err
		}
		ids, err := requiredIDSlice(raw, "evidence_ids", context)
		if err != nil {
			return nil, err
		}
		return ProvidesInteropPayload{Origin: origin, EvidenceIDs: ids, ExportRole: fields["export_role"], LinkMode: fields["link_mode"]}, nil
	case EdgeConsumesInterop:
		const context = "consumes_interop payload"
		if err := exactFields(raw, context, []string{"origin", "use", "abi_expectation"}, nil); err != nil {
			return nil, err
		}
		fields, err := decodeStringFields(raw, context, []string{"use", "abi_expectation"}, nil)
		if err != nil {
			return nil, err
		}
		originRaw, err := requiredObject(raw, "origin", context)
		if err != nil {
			return nil, err
		}
		origin, err := decodeOrigin(originRaw, "consumes_interop origin")
		if err != nil {
			return nil, err
		}
		return ConsumesInteropPayload{Origin: origin, Use: fields["use"], ABIExpectation: fields["abi_expectation"]}, nil
	case EdgeInvokes:
		const context = "invokes payload"
		if err := exactFields(raw, context, []string{"protocol_schema", "executable_resolution", "arguments_contract", "environment_contract"}, []string{"working_directory"}); err != nil {
			return nil, err
		}
		fields, err := decodeStringFields(raw, context, []string{"protocol_schema", "executable_resolution", "arguments_contract", "environment_contract"}, []string{"working_directory"})
		if err != nil {
			return nil, err
		}
		return InvokesPayload{ProtocolSchema: fields["protocol_schema"], ExecutableResolution: fields["executable_resolution"], ArgumentsContract: fields["arguments_contract"], EnvironmentContract: fields["environment_contract"], WorkingDirectory: fields["working_directory"]}, nil
	case EdgePublishes:
		const context = "publishes payload"
		if err := exactFields(raw, context, []string{"destination", "entry_point"}, nil); err != nil {
			return nil, err
		}
		fields, err := decodeStringFields(raw, context, []string{"destination", "entry_point"}, nil)
		if err != nil {
			return nil, err
		}
		return PublishesPayload{Destination: fields["destination"], EntryPoint: fields["entry_point"]}, nil
	default:
		return nil, fmt.Errorf("%s: unsupported edge kind %q", CodeGraphSchemaUnsupported, kind)
	}
}

func sortEdges(edges []Edge) []Edge {
	result := append([]Edge(nil), edges...)
	sort.Slice(result, func(i, j int) bool {
		leftID, _ := result[i].ID()
		rightID, _ := result[j].ID()
		if result[i].Kind != result[j].Kind {
			return result[i].Kind < result[j].Kind
		}
		if result[i].EdgeKey != result[j].EdgeKey {
			return result[i].EdgeKey < result[j].EdgeKey
		}
		return leftID < rightID
	})
	return result
}
