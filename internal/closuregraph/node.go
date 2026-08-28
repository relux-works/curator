package closuregraph

import (
	"fmt"
	"sort"
)

// NodeKind is the closed curator-node-v1 kind discriminator.
type NodeKind string

const (
	// NodeCommandProduct and the related constants enumerate the ten closed
	// intrinsic node kinds.
	NodeCommandProduct NodeKind = "command_product"
	// NodePackageInstance identifies immutable package instances.
	NodePackageInstance NodeKind = "package_instance"
	// NodeSourceSet identifies immutable source inventories.
	NodeSourceSet NodeKind = "source_set"
	// NodeTargetUnit identifies language-specific compilation targets.
	NodeTargetUnit NodeKind = "target_unit"
	// NodeAction identifies declared build or generation actions.
	NodeAction NodeKind = "action"
	// NodeGeneratedArtifact identifies declared generated intermediates.
	NodeGeneratedArtifact NodeKind = "generated_artifact"
	// NodeInteropBoundary identifies explicit language or process boundaries.
	NodeInteropBoundary NodeKind = "interop_boundary"
	// NodeToolchainComponent identifies selected external toolchain components.
	NodeToolchainComponent NodeKind = "toolchain_component"
	// NodeTargetPlatform identifies exact concrete host or target platforms.
	NodeTargetPlatform NodeKind = "target_platform"
	// NodeOutputArtifact identifies immutable expected output declarations.
	NodeOutputArtifact NodeKind = "output_artifact"
)

// TrustRole is the causal origin role of source or tool bytes.
type TrustRole string

const (
	// TrustDependencyInput and the related constants enumerate the causal
	// trust roles recognized by the closure model.
	TrustDependencyInput TrustRole = "dependency_input"
	// TrustExternalToolchain marks Curator-selected toolchain bytes.
	TrustExternalToolchain TrustRole = "external_toolchain"
	// TrustLocalBuildOutput marks causally produced protected build outputs.
	TrustLocalBuildOutput TrustRole = "local_build_output"
	// TrustVerifiedBinaryCandidate marks the unavailable binary-admission seam.
	TrustVerifiedBinaryCandidate TrustRole = "verified_binary_candidate"
)

// ExecutionDomain identifies whether a declared unit executes for the host or
// for the selected target.
type ExecutionDomain string

const (
	// ExecutionHost and ExecutionTarget are the two execution domains.
	ExecutionHost ExecutionDomain = "host"
	// ExecutionTarget declares execution for the selected target platform.
	ExecutionTarget ExecutionDomain = "target"
)

// InteropMode is the closed language/process boundary mode.
type InteropMode string

const (
	// InteropCABI and the related constants enumerate the closed interop modes.
	InteropCABI InteropMode = "c_abi"
	// InteropCXX identifies direct C++ interoperation.
	InteropCXX InteropMode = "cxx_interop"
	// InteropObjCRuntime identifies Objective-C runtime interoperation.
	InteropObjCRuntime InteropMode = "objc_runtime"
	// InteropNativeLink identifies a native link boundary.
	InteropNativeLink InteropMode = "native_link"
	// InteropDynamicLoad identifies a dynamic-load boundary.
	InteropDynamicLoad InteropMode = "dynamic_load"
	// InteropHostExtension identifies a host extension boundary.
	InteropHostExtension InteropMode = "host_extension"
	// InteropSubprocessProtocol identifies a versioned subprocess protocol.
	InteropSubprocessProtocol InteropMode = "subprocess_protocol"
)

// PlatformRole is the closed role a concrete target platform fulfills.
type PlatformRole string

const (
	// PlatformTarget and PlatformHost are the concrete platform roles.
	PlatformTarget PlatformRole = "target"
	// PlatformHost identifies the platform on which evidence or actions execute.
	PlatformHost PlatformRole = "host"
)

// Node is one intrinsic graph record. Payload is a sealed kind-specific type;
// endpoints and other relationships cannot be represented inside it.
type Node struct {
	Kind       NodeKind
	LogicalKey string
	Payload    NodePayload
}

// NodePayload is implemented only by the ten closed intrinsic payload types
// in this package.
type NodePayload interface {
	nodePayload()
	kind() NodeKind
	validate() error
	value() map[string]any
	declaredPlatformRoles() []PlatformRole
}

// CommandProductPayload identifies one declared skill-facing command.
type CommandProductPayload struct {
	Profile            string
	SkillKey           string
	CommandKey         string
	EntryPointContract string
	DeclarationDigest  ID
	PlatformRoleNames  []PlatformRole
}

// PackageInstancePayload identifies one authoritative lock instance.
type PackageInstancePayload struct {
	Profile            string
	Ecosystem          string
	Manager            string
	NormalizedSourceID string
	Origin             string
	LockInstanceKey    string
	Name               string
	Version            string
	ArtifactManifestID ID
	SnapshotDigest     ID
	WorkspacePath      string
	TrustRole          TrustRole
}

// SourceSetPayload identifies an immutable admitted source projection.
type SourceSetPayload struct {
	Profile            string
	Origin             string
	ArtifactManifestID ID
	Projection         []string
	Grammar            string
	TrustRole          TrustRole
	SourceClass        string
	TreeDigest         ID
}

// TargetUnitPayload identifies one atomic compiler or manager unit.
// ConditionExpressions is retained in the closed wire schema but v1 requires
// it to be empty: selectable target predicates must be conditional edges so
// ActiveGraph can record exactly one selected/pruned result for each one.
type TargetUnitPayload struct {
	Profile              string
	TargetName           string
	TargetKind           string
	DeclarationDigest    ID
	Languages            []string
	ExecutionDomain      ExecutionDomain
	ConditionExpressions []Condition
	ExpectedOutputClass  string
	PlatformRoleNames    []PlatformRole
}

// ActionPayload identifies one declared executable action and its abstract
// slots. The executable and slot bindings remain typed edges.
type ActionPayload struct {
	Profile                  string
	ActionSubtype            string
	ExecutionDomain          ExecutionDomain
	ArgvTemplate             []string
	WorkingDirectoryTemplate string
	ToolSlotNames            []string
	ReadSlotNames            []string
	WriteSlotNames           []string
	EnvironmentPolicyID      string
	ProcessPolicyID          string
	Network                  string
	PlatformRoleNames        []PlatformRole
}

// GeneratedArtifactPayload identifies a declared locally generated input.
type GeneratedArtifactPayload struct {
	Profile           string
	LogicalPath       string
	Slot              string
	ExpectedClass     string
	Grammar           string
	Role              string
	DeclarationDigest ID
}

// InteropBoundaryPayload identifies an explicit language or process contract.
type InteropBoundaryPayload struct {
	Profile                 string
	Mode                    InteropMode
	ProviderLanguageClasses []string
	ConsumerLanguageClasses []string
	ContractDigest          ID
	ABI                     string
	Runtime                 string
	ProtocolSchema          string
	InterfaceContract       string
	CallingConvention       string
	LinkLoadSemantics       string
	PlatformRoleNames       []PlatformRole
}

// ToolchainComponentPayload identifies a selected external toolchain
// component. Its C0/C4 causal authority is kept in ToolchainAuthority.
type ToolchainComponentPayload struct {
	ComponentRole          string
	ContentFingerprint     ID
	ExecutableRelativePath string
	PlatformABI            string
	PolicySelector         string
	VersionOutput          string
	LinkFingerprintIDs     []ID
	SDKFactsDigest         ID
	TimeOfUseRecheckRule   string
	ExecutionDomain        ExecutionDomain
	PlatformRoleNames      []PlatformRole
}

// TargetPlatformPayload identifies one exact host or target destination.
type TargetPlatformPayload struct {
	OS             string
	Architecture   string
	ABI            string
	Libc           string
	MinimumRuntime string
	SDKID          string
	TargetTriple   string
	Runtime        string
	LanguageModes  map[string]string
	Tuning         map[string]string
}

// OutputArtifactPayload is an immutable expected-output declaration. Observed
// C6 bytes are represented only by ProducedArtifactObservation.
type OutputArtifactPayload struct {
	Profile                string
	LogicalPath            string
	ExpectedClass          string
	OutputRole             string
	CompatibilityPredicate string
	DeclarationDigest      ID
	PlatformRoleNames      []PlatformRole
}

func (CommandProductPayload) nodePayload()     {}
func (PackageInstancePayload) nodePayload()    {}
func (SourceSetPayload) nodePayload()          {}
func (TargetUnitPayload) nodePayload()         {}
func (ActionPayload) nodePayload()             {}
func (GeneratedArtifactPayload) nodePayload()  {}
func (InteropBoundaryPayload) nodePayload()    {}
func (ToolchainComponentPayload) nodePayload() {}
func (TargetPlatformPayload) nodePayload()     {}
func (OutputArtifactPayload) nodePayload()     {}

func (CommandProductPayload) kind() NodeKind     { return NodeCommandProduct }
func (PackageInstancePayload) kind() NodeKind    { return NodePackageInstance }
func (SourceSetPayload) kind() NodeKind          { return NodeSourceSet }
func (TargetUnitPayload) kind() NodeKind         { return NodeTargetUnit }
func (ActionPayload) kind() NodeKind             { return NodeAction }
func (GeneratedArtifactPayload) kind() NodeKind  { return NodeGeneratedArtifact }
func (InteropBoundaryPayload) kind() NodeKind    { return NodeInteropBoundary }
func (ToolchainComponentPayload) kind() NodeKind { return NodeToolchainComponent }
func (TargetPlatformPayload) kind() NodeKind     { return NodeTargetPlatform }
func (OutputArtifactPayload) kind() NodeKind     { return NodeOutputArtifact }

// Validate checks the closed node schema and intrinsic payload.
func (node Node) Validate() error {
	if !validNodeKind(node.Kind) {
		return fmt.Errorf("%s: unsupported node kind %q", CodeGraphSchemaUnsupported, node.Kind)
	}
	if err := validatePortableText(node.LogicalKey, "node logical_key", false); err != nil {
		return err
	}
	if node.Payload == nil {
		return fmt.Errorf("node payload is required")
	}
	payloadKind, canonical := nodePayloadValueKind(node.Payload)
	if !canonical {
		return fmt.Errorf("%s: node payload for kind %q must use its canonical value representation, got %T", CodeGraphSchemaUnsupported, node.Kind, node.Payload)
	}
	if payloadKind != node.Kind {
		return fmt.Errorf("node kind %q does not match payload kind %q", node.Kind, payloadKind)
	}
	return node.Payload.validate()
}

// nodePayloadValueKind recognizes only the exact value representations that
// the canonical node codec emits. The sealed interface's value-receiver
// methods also place pointers in its method set; accepting those alternate
// dynamic representations would create a second in-memory schema and permits
// typed-nil method panics before a canonical diagnostic can be returned.
func nodePayloadValueKind(payload NodePayload) (NodeKind, bool) {
	switch payload.(type) {
	case CommandProductPayload:
		return NodeCommandProduct, true
	case PackageInstancePayload:
		return NodePackageInstance, true
	case SourceSetPayload:
		return NodeSourceSet, true
	case TargetUnitPayload:
		return NodeTargetUnit, true
	case ActionPayload:
		return NodeAction, true
	case GeneratedArtifactPayload:
		return NodeGeneratedArtifact, true
	case InteropBoundaryPayload:
		return NodeInteropBoundary, true
	case ToolchainComponentPayload:
		return NodeToolchainComponent, true
	case TargetPlatformPayload:
		return NodeTargetPlatform, true
	case OutputArtifactPayload:
		return NodeOutputArtifact, true
	default:
		return "", false
	}
}

// CanonicalBytes returns exact curator-node-v1 CCJ bytes.
func (node Node) CanonicalBytes() ([]byte, error) { return canonicalBytes(node) }

// ID derives the domain-separated curator-node-v1 identity.
func (node Node) ID() (ID, error) { return recordID(node) }

func (node Node) domainLabel() string { return LabelNode }

func (node Node) canonicalValue() map[string]any {
	return map[string]any{
		"kind": string(node.Kind), "logical_key": node.LogicalKey, "payload": node.Payload.value(),
	}
}

func validNodeKind(kind NodeKind) bool {
	switch kind {
	case NodeCommandProduct, NodePackageInstance, NodeSourceSet, NodeTargetUnit,
		NodeAction, NodeGeneratedArtifact, NodeInteropBoundary,
		NodeToolchainComponent, NodeTargetPlatform, NodeOutputArtifact:
		return true
	default:
		return false
	}
}

func validExecutionDomain(domain ExecutionDomain) bool {
	return domain == ExecutionHost || domain == ExecutionTarget
}

func validInteropMode(mode InteropMode) bool {
	switch mode {
	case InteropCABI, InteropCXX, InteropObjCRuntime, InteropNativeLink,
		InteropDynamicLoad, InteropHostExtension, InteropSubprocessProtocol:
		return true
	default:
		return false
	}
}

func validateProfile(value string) error { return validatePortableText(value, "profile", false) }

func validateRoles(roles []PlatformRole, field string) error {
	if roles == nil {
		return nil
	}
	values := make([]string, len(roles))
	for index, role := range roles {
		if role != PlatformTarget && role != PlatformHost {
			return fmt.Errorf("%s contains unsupported platform role %q", field, role)
		}
		values[index] = string(role)
	}
	return validateUniqueStrings(values, field, true)
}

func validateRequiredRole(roles []PlatformRole, field string, required PlatformRole) error {
	if err := validateRoles(roles, field); err != nil {
		return err
	}
	if roles == nil {
		return nil
	}
	if len(roles) == 0 {
		return fmt.Errorf("%s must not be explicitly empty", field)
	}
	for _, role := range roles {
		if role == required {
			return nil
		}
	}
	return fmt.Errorf("%s must include required platform role %q", field, required)
}

func requiredRoleForDomain(domain ExecutionDomain) PlatformRole {
	if domain == ExecutionHost {
		return PlatformHost
	}
	return PlatformTarget
}

func roleValues(roles []PlatformRole) []any {
	values := make([]any, len(roles))
	for index, role := range roles {
		values[index] = string(role)
	}
	return values
}

func defaultTargetRoles(explicit []PlatformRole) []PlatformRole {
	if explicit != nil {
		return explicit
	}
	return []PlatformRole{PlatformTarget}
}

func (payload CommandProductPayload) validate() error {
	if err := validateProfile(payload.Profile); err != nil {
		return err
	}
	if err := validatePortableTextFields(map[string]string{"skill_key": payload.SkillKey, "command_key": payload.CommandKey, "entry_point_contract": payload.EntryPointContract}, false, false); err != nil {
		return err
	}
	if err := validateID(payload.DeclarationDigest, "declaration_digest"); err != nil {
		return err
	}
	return validateRequiredRole(payload.PlatformRoleNames, "platform_role_names", PlatformTarget)
}

func (payload CommandProductPayload) value() map[string]any {
	value := map[string]any{"profile": payload.Profile, "skill_key": payload.SkillKey, "command_key": payload.CommandKey, "entry_point_contract": payload.EntryPointContract, "declaration_digest": string(payload.DeclarationDigest)}
	if payload.PlatformRoleNames != nil {
		value["platform_role_names"] = roleValues(payload.PlatformRoleNames)
	}
	return value
}
func (payload CommandProductPayload) declaredPlatformRoles() []PlatformRole {
	return defaultTargetRoles(payload.PlatformRoleNames)
}

func (payload PackageInstancePayload) validate() error {
	if err := validateProfile(payload.Profile); err != nil {
		return err
	}
	if err := validatePortableTextFields(map[string]string{"ecosystem": payload.Ecosystem, "origin": payload.Origin, "lock_instance_key": payload.LockInstanceKey, "name": payload.Name, "version": payload.Version}, false, false); err != nil {
		return err
	}
	if err := validatePortableTextFields(map[string]string{"manager": payload.Manager, "normalized_source_id": payload.NormalizedSourceID}, false, true); err != nil {
		return err
	}
	if payload.ArtifactManifestID == "" && payload.SnapshotDigest == "" {
		return fmt.Errorf("package instance requires artifact_manifest_id or snapshot_digest")
	}
	if payload.ArtifactManifestID != "" {
		if err := validateID(payload.ArtifactManifestID, "artifact_manifest_id"); err != nil {
			return err
		}
	}
	if payload.SnapshotDigest != "" {
		if err := validateID(payload.SnapshotDigest, "snapshot_digest"); err != nil {
			return err
		}
	}
	if payload.WorkspacePath != "" {
		if err := validatePortablePath(payload.WorkspacePath, "workspace_path"); err != nil {
			return err
		}
	}
	if payload.TrustRole != TrustDependencyInput {
		return fmt.Errorf("package instance trust_role must be %q", TrustDependencyInput)
	}
	return nil
}

func (payload PackageInstancePayload) value() map[string]any {
	value := map[string]any{"profile": payload.Profile, "ecosystem": payload.Ecosystem, "origin": payload.Origin, "lock_instance_key": payload.LockInstanceKey, "name": payload.Name, "version": payload.Version, "trust_role": string(payload.TrustRole)}
	if payload.Manager != "" {
		value["manager"] = payload.Manager
	}
	if payload.NormalizedSourceID != "" {
		value["normalized_source_id"] = payload.NormalizedSourceID
	}
	if payload.ArtifactManifestID != "" {
		value["artifact_manifest_id"] = string(payload.ArtifactManifestID)
	}
	if payload.SnapshotDigest != "" {
		value["snapshot_digest"] = string(payload.SnapshotDigest)
	}
	if payload.WorkspacePath != "" {
		value["workspace_path"] = payload.WorkspacePath
	}
	return value
}
func (PackageInstancePayload) declaredPlatformRoles() []PlatformRole { return nil }

func (payload SourceSetPayload) validate() error {
	if err := validateProfile(payload.Profile); err != nil {
		return err
	}
	if err := validatePortableText(payload.Origin, "origin", false); err != nil {
		return err
	}
	if err := validateID(payload.ArtifactManifestID, "artifact_manifest_id"); err != nil {
		return err
	}
	if payload.Projection == nil {
		return fmt.Errorf("projection must be an explicit array")
	}
	for index, path := range payload.Projection {
		if err := validatePortablePath(path, fmt.Sprintf("projection[%d]", index)); err != nil {
			return err
		}
	}
	if err := validateUniqueStrings(payload.Projection, "projection", true); err != nil {
		return err
	}
	if err := validatePortableText(payload.Grammar, "grammar", false); err != nil {
		return err
	}
	if payload.TrustRole != TrustDependencyInput {
		return fmt.Errorf("source set trust_role must be %q", TrustDependencyInput)
	}
	if payload.SourceClass != "" {
		if err := validatePortableText(payload.SourceClass, "source_class", false); err != nil {
			return err
		}
	}
	if payload.TreeDigest != "" {
		if err := validateID(payload.TreeDigest, "tree_digest"); err != nil {
			return err
		}
	}
	return nil
}

func (payload SourceSetPayload) value() map[string]any {
	value := map[string]any{"profile": payload.Profile, "origin": payload.Origin, "artifact_manifest_id": string(payload.ArtifactManifestID), "projection": stringsToAny(payload.Projection), "grammar": payload.Grammar, "trust_role": string(payload.TrustRole)}
	if payload.SourceClass != "" {
		value["source_class"] = payload.SourceClass
	}
	if payload.TreeDigest != "" {
		value["tree_digest"] = string(payload.TreeDigest)
	}
	return value
}
func (SourceSetPayload) declaredPlatformRoles() []PlatformRole { return nil }

func (payload TargetUnitPayload) validate() error {
	if err := validateProfile(payload.Profile); err != nil {
		return err
	}
	if err := validatePortableTextFields(map[string]string{"target_name": payload.TargetName, "target_kind": payload.TargetKind, "expected_output_class": payload.ExpectedOutputClass}, false, false); err != nil {
		return err
	}
	if err := validateID(payload.DeclarationDigest, "declaration_digest"); err != nil {
		return err
	}
	if err := validateStringSlice(payload.Languages, "languages", true); err != nil {
		return err
	}
	if !validExecutionDomain(payload.ExecutionDomain) {
		return fmt.Errorf("unsupported execution_domain %q", payload.ExecutionDomain)
	}
	if payload.ConditionExpressions == nil {
		return fmt.Errorf("condition_expressions must be an explicit array")
	}
	if len(payload.ConditionExpressions) != 0 {
		return fmt.Errorf("condition_expressions on target_unit are unsupported until the active schema can record their exact activation; use conditional capture edges")
	}
	return validateRequiredRole(payload.PlatformRoleNames, "platform_role_names", requiredRoleForDomain(payload.ExecutionDomain))
}

func (payload TargetUnitPayload) value() map[string]any {
	conditions := make([]any, len(payload.ConditionExpressions))
	for i, condition := range payload.ConditionExpressions {
		conditions[i] = condition.value()
	}
	value := map[string]any{"profile": payload.Profile, "target_name": payload.TargetName, "target_kind": payload.TargetKind, "declaration_digest": string(payload.DeclarationDigest), "languages": stringsToAny(payload.Languages), "execution_domain": string(payload.ExecutionDomain), "condition_expressions": conditions, "expected_output_class": payload.ExpectedOutputClass}
	if payload.PlatformRoleNames != nil {
		value["platform_role_names"] = roleValues(payload.PlatformRoleNames)
	}
	return value
}
func (payload TargetUnitPayload) declaredPlatformRoles() []PlatformRole {
	if payload.PlatformRoleNames != nil {
		return payload.PlatformRoleNames
	}
	if payload.ExecutionDomain == ExecutionHost {
		return []PlatformRole{PlatformHost}
	}
	return []PlatformRole{PlatformTarget}
}

func (payload ActionPayload) validate() error {
	if err := validateProfile(payload.Profile); err != nil {
		return err
	}
	if err := validatePortableText(payload.ActionSubtype, "action_subtype", false); err != nil {
		return err
	}
	if !validExecutionDomain(payload.ExecutionDomain) {
		return fmt.Errorf("unsupported execution_domain %q", payload.ExecutionDomain)
	}
	if len(payload.ArgvTemplate) == 0 {
		return fmt.Errorf("argv_template must be a non-empty explicit array")
	}
	for index, value := range payload.ArgvTemplate {
		if err := validatePortableText(value, fmt.Sprintf("argv_template[%d]", index), false); err != nil {
			return err
		}
	}
	if err := validateStringSlice(payload.ToolSlotNames, "tool_slot_names", true); err != nil {
		return err
	}
	if err := validateStringSlice(payload.ReadSlotNames, "read_slot_names", true); err != nil {
		return err
	}
	if err := validateStringSlice(payload.WriteSlotNames, "write_slot_names", true); err != nil {
		return err
	}
	if err := validateActionTemplate(payload); err != nil {
		return err
	}
	if err := validatePortableTextFields(map[string]string{"environment_policy_id": payload.EnvironmentPolicyID, "process_policy_id": payload.ProcessPolicyID, "network": payload.Network}, false, false); err != nil {
		return err
	}
	return validateRequiredRole(payload.PlatformRoleNames, "platform_role_names", requiredRoleForDomain(payload.ExecutionDomain))
}

func (payload ActionPayload) value() map[string]any {
	value := map[string]any{"profile": payload.Profile, "action_subtype": payload.ActionSubtype, "execution_domain": string(payload.ExecutionDomain), "argv_template": stringsToAny(payload.ArgvTemplate), "tool_slot_names": stringsToAny(payload.ToolSlotNames), "read_slot_names": stringsToAny(payload.ReadSlotNames), "write_slot_names": stringsToAny(payload.WriteSlotNames), "environment_policy_id": payload.EnvironmentPolicyID, "process_policy_id": payload.ProcessPolicyID, "network": payload.Network}
	if payload.WorkingDirectoryTemplate != "" {
		value["working_directory_template"] = payload.WorkingDirectoryTemplate
	}
	if payload.PlatformRoleNames != nil {
		value["platform_role_names"] = roleValues(payload.PlatformRoleNames)
	}
	return value
}
func (payload ActionPayload) declaredPlatformRoles() []PlatformRole {
	if payload.PlatformRoleNames != nil {
		return payload.PlatformRoleNames
	}
	if payload.ExecutionDomain == ExecutionHost {
		return []PlatformRole{PlatformHost}
	}
	return []PlatformRole{PlatformTarget}
}

func (payload GeneratedArtifactPayload) validate() error {
	if err := validateProfile(payload.Profile); err != nil {
		return err
	}
	if err := validatePortablePath(payload.LogicalPath, "logical_path"); err != nil {
		return err
	}
	if err := validatePortableTextFields(map[string]string{"slot": payload.Slot, "expected_class": payload.ExpectedClass, "grammar": payload.Grammar, "role": payload.Role}, false, false); err != nil {
		return err
	}
	return validateID(payload.DeclarationDigest, "declaration_digest")
}
func (payload GeneratedArtifactPayload) value() map[string]any {
	return map[string]any{"profile": payload.Profile, "logical_path": payload.LogicalPath, "slot": payload.Slot, "expected_class": payload.ExpectedClass, "grammar": payload.Grammar, "role": payload.Role, "declaration_digest": string(payload.DeclarationDigest)}
}
func (GeneratedArtifactPayload) declaredPlatformRoles() []PlatformRole { return nil }

func (payload InteropBoundaryPayload) validate() error {
	if err := validateProfile(payload.Profile); err != nil {
		return err
	}
	if !validInteropMode(payload.Mode) {
		return fmt.Errorf("unsupported interop mode %q", payload.Mode)
	}
	if err := validateStringSlice(payload.ProviderLanguageClasses, "provider_language_classes", true); err != nil {
		return err
	}
	if err := validateStringSlice(payload.ConsumerLanguageClasses, "consumer_language_classes", true); err != nil {
		return err
	}
	if len(payload.ProviderLanguageClasses) == 0 || len(payload.ConsumerLanguageClasses) == 0 {
		return fmt.Errorf("interop boundary requires non-empty provider and consumer language classes")
	}
	if err := validateID(payload.ContractDigest, "contract_digest"); err != nil {
		return err
	}
	if err := validatePortableTextFields(map[string]string{"abi": payload.ABI, "runtime": payload.Runtime, "protocol_schema": payload.ProtocolSchema, "interface_contract": payload.InterfaceContract, "calling_convention": payload.CallingConvention, "link_load_semantics": payload.LinkLoadSemantics}, false, true); err != nil {
		return err
	}
	type requiredInteropField struct{ name, value string }
	requiredFields := []requiredInteropField{}
	switch payload.Mode {
	case InteropCABI:
		requiredFields = []requiredInteropField{{"abi", payload.ABI}, {"calling_convention", payload.CallingConvention}, {"interface_contract", payload.InterfaceContract}, {"link_load_semantics", payload.LinkLoadSemantics}}
	case InteropCXX:
		requiredFields = []requiredInteropField{{"abi", payload.ABI}, {"calling_convention", payload.CallingConvention}, {"interface_contract", payload.InterfaceContract}, {"link_load_semantics", payload.LinkLoadSemantics}, {"runtime", payload.Runtime}}
	case InteropObjCRuntime:
		requiredFields = []requiredInteropField{{"calling_convention", payload.CallingConvention}, {"interface_contract", payload.InterfaceContract}, {"link_load_semantics", payload.LinkLoadSemantics}, {"runtime", payload.Runtime}}
	case InteropNativeLink:
		requiredFields = []requiredInteropField{{"abi", payload.ABI}, {"interface_contract", payload.InterfaceContract}, {"link_load_semantics", payload.LinkLoadSemantics}}
	case InteropDynamicLoad:
		requiredFields = []requiredInteropField{{"abi", payload.ABI}, {"interface_contract", payload.InterfaceContract}, {"link_load_semantics", payload.LinkLoadSemantics}, {"runtime", payload.Runtime}}
	case InteropHostExtension:
		requiredFields = []requiredInteropField{{"interface_contract", payload.InterfaceContract}, {"link_load_semantics", payload.LinkLoadSemantics}, {"runtime", payload.Runtime}}
	case InteropSubprocessProtocol:
		requiredFields = []requiredInteropField{{"interface_contract", payload.InterfaceContract}, {"link_load_semantics", payload.LinkLoadSemantics}, {"protocol_schema", payload.ProtocolSchema}}
	}
	for _, field := range requiredFields {
		if field.value == "" {
			return fmt.Errorf("interop mode %q requires %s", payload.Mode, field.name)
		}
	}
	requiredRole := PlatformTarget
	if payload.Mode == InteropHostExtension {
		requiredRole = PlatformHost
	}
	return validateRequiredRole(payload.PlatformRoleNames, "platform_role_names", requiredRole)
}
func (payload InteropBoundaryPayload) value() map[string]any {
	value := map[string]any{"profile": payload.Profile, "mode": string(payload.Mode), "provider_language_classes": stringsToAny(payload.ProviderLanguageClasses), "consumer_language_classes": stringsToAny(payload.ConsumerLanguageClasses), "contract_digest": string(payload.ContractDigest)}
	optional := map[string]string{"abi": payload.ABI, "runtime": payload.Runtime, "protocol_schema": payload.ProtocolSchema, "interface_contract": payload.InterfaceContract, "calling_convention": payload.CallingConvention, "link_load_semantics": payload.LinkLoadSemantics}
	for _, field := range sortedMapKeys(optional) {
		item := optional[field]
		if item != "" {
			value[field] = item
		}
	}
	if payload.PlatformRoleNames != nil {
		value["platform_role_names"] = roleValues(payload.PlatformRoleNames)
	}
	return value
}
func (payload InteropBoundaryPayload) declaredPlatformRoles() []PlatformRole {
	if payload.PlatformRoleNames != nil {
		return payload.PlatformRoleNames
	}
	if payload.Mode == InteropHostExtension {
		return []PlatformRole{PlatformHost}
	}
	return []PlatformRole{PlatformTarget}
}

func (payload ToolchainComponentPayload) validate() error {
	if err := validatePortableTextFields(map[string]string{"component_role": payload.ComponentRole, "platform_abi": payload.PlatformABI, "policy_selector": payload.PolicySelector, "version_output": payload.VersionOutput}, false, false); err != nil {
		return err
	}
	if err := validateID(payload.ContentFingerprint, "content_fingerprint"); err != nil {
		return err
	}
	if err := validatePortablePath(payload.ExecutableRelativePath, "executable_relative_path"); err != nil {
		return err
	}
	if payload.LinkFingerprintIDs != nil {
		if err := validateIDSlice(payload.LinkFingerprintIDs, "link_fingerprint_ids", true); err != nil {
			return err
		}
	}
	if payload.SDKFactsDigest != "" {
		if err := validateID(payload.SDKFactsDigest, "sdk_facts_digest"); err != nil {
			return err
		}
	}
	if payload.TimeOfUseRecheckRule != "" {
		if err := validatePortableText(payload.TimeOfUseRecheckRule, "time_of_use_recheck_rule", false); err != nil {
			return err
		}
	}
	if payload.ExecutionDomain != "" && !validExecutionDomain(payload.ExecutionDomain) {
		return fmt.Errorf("unsupported execution_domain %q", payload.ExecutionDomain)
	}
	return validateRequiredRole(payload.PlatformRoleNames, "platform_role_names", requiredRoleForDomain(payload.ExecutionDomain))
}
func (payload ToolchainComponentPayload) value() map[string]any {
	value := map[string]any{"component_role": payload.ComponentRole, "content_fingerprint": string(payload.ContentFingerprint), "executable_relative_path": payload.ExecutableRelativePath, "platform_abi": payload.PlatformABI, "policy_selector": payload.PolicySelector, "version_output": payload.VersionOutput}
	if payload.LinkFingerprintIDs != nil {
		value["link_fingerprint_ids"] = idsToAny(payload.LinkFingerprintIDs)
	}
	if payload.SDKFactsDigest != "" {
		value["sdk_facts_digest"] = string(payload.SDKFactsDigest)
	}
	if payload.TimeOfUseRecheckRule != "" {
		value["time_of_use_recheck_rule"] = payload.TimeOfUseRecheckRule
	}
	if payload.ExecutionDomain != "" {
		value["execution_domain"] = string(payload.ExecutionDomain)
	}
	if payload.PlatformRoleNames != nil {
		value["platform_role_names"] = roleValues(payload.PlatformRoleNames)
	}
	return value
}
func (payload ToolchainComponentPayload) declaredPlatformRoles() []PlatformRole {
	if payload.PlatformRoleNames != nil {
		return payload.PlatformRoleNames
	}
	if payload.ExecutionDomain == ExecutionHost {
		return []PlatformRole{PlatformHost}
	}
	return []PlatformRole{PlatformTarget}
}

func (payload TargetPlatformPayload) validate() error {
	if err := validatePortableTextFields(map[string]string{"os": payload.OS, "architecture": payload.Architecture, "abi": payload.ABI, "libc": payload.Libc, "minimum_runtime": payload.MinimumRuntime, "sdk_id": payload.SDKID, "target_triple": payload.TargetTriple}, false, false); err != nil {
		return err
	}
	if payload.Runtime != "" {
		if err := validatePortableText(payload.Runtime, "runtime", false); err != nil {
			return err
		}
	}
	if payload.LanguageModes != nil {
		if err := validatePortableStringMap(payload.LanguageModes, "language_modes", false); err != nil {
			return err
		}
	}
	if payload.Tuning != nil {
		if err := validatePortableStringMap(payload.Tuning, "tuning", false); err != nil {
			return err
		}
	}
	return nil
}
func (payload TargetPlatformPayload) value() map[string]any {
	value := map[string]any{"os": payload.OS, "architecture": payload.Architecture, "abi": payload.ABI, "libc": payload.Libc, "minimum_runtime": payload.MinimumRuntime, "sdk_id": payload.SDKID, "target_triple": payload.TargetTriple}
	if payload.Runtime != "" {
		value["runtime"] = payload.Runtime
	}
	if payload.LanguageModes != nil {
		value["language_modes"] = stringMapToAny(payload.LanguageModes)
	}
	if payload.Tuning != nil {
		value["tuning"] = stringMapToAny(payload.Tuning)
	}
	return value
}
func (TargetPlatformPayload) declaredPlatformRoles() []PlatformRole { return nil }

func (payload OutputArtifactPayload) validate() error {
	if err := validateProfile(payload.Profile); err != nil {
		return err
	}
	if err := validatePortablePath(payload.LogicalPath, "logical_path"); err != nil {
		return err
	}
	if err := validatePortableTextFields(map[string]string{"expected_class": payload.ExpectedClass, "output_role": payload.OutputRole}, false, false); err != nil {
		return err
	}
	if payload.CompatibilityPredicate != "" {
		if err := validatePortableText(payload.CompatibilityPredicate, "compatibility_predicate", false); err != nil {
			return err
		}
	}
	if payload.DeclarationDigest != "" {
		if err := validateID(payload.DeclarationDigest, "declaration_digest"); err != nil {
			return err
		}
	}
	return validateRequiredRole(payload.PlatformRoleNames, "platform_role_names", PlatformTarget)
}
func (payload OutputArtifactPayload) value() map[string]any {
	value := map[string]any{"profile": payload.Profile, "logical_path": payload.LogicalPath, "expected_class": payload.ExpectedClass, "output_role": payload.OutputRole}
	if payload.CompatibilityPredicate != "" {
		value["compatibility_predicate"] = payload.CompatibilityPredicate
	}
	if payload.DeclarationDigest != "" {
		value["declaration_digest"] = string(payload.DeclarationDigest)
	}
	if payload.PlatformRoleNames != nil {
		value["platform_role_names"] = roleValues(payload.PlatformRoleNames)
	}
	return value
}
func (payload OutputArtifactPayload) declaredPlatformRoles() []PlatformRole {
	return defaultTargetRoles(payload.PlatformRoleNames)
}

// DecodeNode accepts exact canonical CCJ bytes and rejects unknown fields or
// unsupported payload kinds.
func DecodeNode(payload []byte) (Node, error) {
	raw, err := decodeCanonicalObject(payload, "node")
	if err != nil {
		return Node{}, err
	}
	if err := exactFields(raw, "node", []string{"kind", "logical_key", "payload"}, nil); err != nil {
		return Node{}, err
	}
	kindValue, err := requiredString(raw, "kind", "node")
	if err != nil {
		return Node{}, err
	}
	logicalKey, err := requiredString(raw, "logical_key", "node")
	if err != nil {
		return Node{}, err
	}
	payloadRaw, err := requiredObject(raw, "payload", "node")
	if err != nil {
		return Node{}, err
	}
	nodePayload, err := decodeNodePayload(NodeKind(kindValue), payloadRaw)
	if err != nil {
		return Node{}, err
	}
	node := Node{Kind: NodeKind(kindValue), LogicalKey: logicalKey, Payload: nodePayload}
	if err := node.Validate(); err != nil {
		return Node{}, err
	}
	if err := requireDecodedRecordRoundTrip(payload, node); err != nil {
		return Node{}, err
	}
	return node, nil
}

func decodeNodePayload(kind NodeKind, raw map[string]any) (NodePayload, error) {
	switch kind {
	case NodeCommandProduct:
		return decodeCommandProduct(raw)
	case NodePackageInstance:
		return decodePackageInstance(raw)
	case NodeSourceSet:
		return decodeSourceSet(raw)
	case NodeTargetUnit:
		return decodeTargetUnit(raw)
	case NodeAction:
		return decodeAction(raw)
	case NodeGeneratedArtifact:
		return decodeGeneratedArtifact(raw)
	case NodeInteropBoundary:
		return decodeInteropBoundary(raw)
	case NodeToolchainComponent:
		return decodeToolchainComponent(raw)
	case NodeTargetPlatform:
		return decodeTargetPlatform(raw)
	case NodeOutputArtifact:
		return decodeOutputArtifact(raw)
	default:
		return nil, fmt.Errorf("%s: unsupported node kind %q", CodeGraphSchemaUnsupported, kind)
	}
}

func parseRoles(raw map[string]any, context string) ([]PlatformRole, error) {
	values, err := optionalStringSlice(raw, "platform_role_names", context)
	if err != nil || values == nil {
		return nil, err
	}
	roles := make([]PlatformRole, len(values))
	for i, value := range values {
		roles[i] = PlatformRole(value)
	}
	return roles, nil
}

func decodeCommandProduct(raw map[string]any) (NodePayload, error) {
	const context = "command_product payload"
	if err := exactFields(raw, context, []string{"profile", "skill_key", "command_key", "entry_point_contract", "declaration_digest"}, []string{"platform_role_names"}); err != nil {
		return nil, err
	}
	fields, err := decodeStringFields(raw, context, []string{"profile", "skill_key", "command_key", "entry_point_contract", "declaration_digest"}, nil)
	if err != nil {
		return nil, err
	}
	roles, err := parseRoles(raw, context)
	if err != nil {
		return nil, err
	}
	return CommandProductPayload{Profile: fields["profile"], SkillKey: fields["skill_key"], CommandKey: fields["command_key"], EntryPointContract: fields["entry_point_contract"], DeclarationDigest: ID(fields["declaration_digest"]), PlatformRoleNames: roles}, nil
}

func decodePackageInstance(raw map[string]any) (NodePayload, error) {
	const context = "package_instance payload"
	required := []string{"profile", "ecosystem", "origin", "lock_instance_key", "name", "version", "trust_role"}
	optional := []string{"manager", "normalized_source_id", "artifact_manifest_id", "snapshot_digest", "workspace_path"}
	if err := exactFields(raw, context, required, optional); err != nil {
		return nil, err
	}
	fields, err := decodeStringFields(raw, context, required, optional)
	if err != nil {
		return nil, err
	}
	return PackageInstancePayload{
		Profile: fields["profile"], Ecosystem: fields["ecosystem"], Manager: fields["manager"],
		NormalizedSourceID: fields["normalized_source_id"], Origin: fields["origin"],
		LockInstanceKey: fields["lock_instance_key"], Name: fields["name"], Version: fields["version"],
		ArtifactManifestID: ID(fields["artifact_manifest_id"]), SnapshotDigest: ID(fields["snapshot_digest"]),
		WorkspacePath: fields["workspace_path"], TrustRole: TrustRole(fields["trust_role"]),
	}, nil
}

func decodeSourceSet(raw map[string]any) (NodePayload, error) {
	const context = "source_set payload"
	if err := exactFields(raw, context, []string{"profile", "origin", "artifact_manifest_id", "projection", "grammar", "trust_role"}, []string{"source_class", "tree_digest"}); err != nil {
		return nil, err
	}
	fields, err := decodeStringFields(raw, context, []string{"profile", "origin", "artifact_manifest_id", "grammar", "trust_role"}, []string{"source_class", "tree_digest"})
	if err != nil {
		return nil, err
	}
	projection, err := requiredStringSlice(raw, "projection", context)
	if err != nil {
		return nil, err
	}
	return SourceSetPayload{Profile: fields["profile"], Origin: fields["origin"], ArtifactManifestID: ID(fields["artifact_manifest_id"]), Projection: projection, Grammar: fields["grammar"], TrustRole: TrustRole(fields["trust_role"]), SourceClass: fields["source_class"], TreeDigest: ID(fields["tree_digest"])}, nil
}

func decodeTargetUnit(raw map[string]any) (NodePayload, error) {
	const context = "target_unit payload"
	if err := exactFields(raw, context, []string{"profile", "target_name", "target_kind", "declaration_digest", "languages", "execution_domain", "condition_expressions", "expected_output_class"}, []string{"platform_role_names"}); err != nil {
		return nil, err
	}
	fields, err := decodeStringFields(raw, context, []string{"profile", "target_name", "target_kind", "declaration_digest", "execution_domain", "expected_output_class"}, nil)
	if err != nil {
		return nil, err
	}
	p := TargetUnitPayload{Profile: fields["profile"], TargetName: fields["target_name"], TargetKind: fields["target_kind"], DeclarationDigest: ID(fields["declaration_digest"]), ExecutionDomain: ExecutionDomain(fields["execution_domain"]), ExpectedOutputClass: fields["expected_output_class"]}
	languages, err := requiredStringSlice(raw, "languages", context)
	if err != nil {
		return nil, err
	}
	p.Languages = languages
	conditionsRaw, ok := raw["condition_expressions"].([]any)
	if !ok {
		return nil, fmt.Errorf("target_unit payload field condition_expressions must be an array")
	}
	p.ConditionExpressions = make([]Condition, len(conditionsRaw))
	for i, item := range conditionsRaw {
		object, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("condition_expressions[%d] must be an object", i)
		}
		condition, err := decodeConditionObject(object)
		if err != nil {
			return nil, err
		}
		p.ConditionExpressions[i] = condition
	}
	p.PlatformRoleNames, err = parseRoles(raw, context)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func decodeAction(raw map[string]any) (NodePayload, error) {
	const context = "action payload"
	if err := exactFields(raw, context, []string{"profile", "action_subtype", "execution_domain", "argv_template", "tool_slot_names", "read_slot_names", "write_slot_names", "environment_policy_id", "process_policy_id", "network"}, []string{"working_directory_template", "platform_role_names"}); err != nil {
		return nil, err
	}
	fields, err := decodeStringFields(raw, context, []string{"profile", "action_subtype", "execution_domain", "environment_policy_id", "process_policy_id", "network"}, []string{"working_directory_template"})
	if err != nil {
		return nil, err
	}
	p := ActionPayload{Profile: fields["profile"], ActionSubtype: fields["action_subtype"], ExecutionDomain: ExecutionDomain(fields["execution_domain"]), WorkingDirectoryTemplate: fields["working_directory_template"], EnvironmentPolicyID: fields["environment_policy_id"], ProcessPolicyID: fields["process_policy_id"], Network: fields["network"]}
	p.ArgvTemplate, err = requiredStringSlice(raw, "argv_template", context)
	if err != nil {
		return nil, err
	}
	p.ToolSlotNames, err = requiredStringSlice(raw, "tool_slot_names", context)
	if err != nil {
		return nil, err
	}
	p.ReadSlotNames, err = requiredStringSlice(raw, "read_slot_names", context)
	if err != nil {
		return nil, err
	}
	p.WriteSlotNames, err = requiredStringSlice(raw, "write_slot_names", context)
	if err != nil {
		return nil, err
	}
	p.PlatformRoleNames, err = parseRoles(raw, context)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func decodeGeneratedArtifact(raw map[string]any) (NodePayload, error) {
	const context = "generated_artifact payload"
	required := []string{"profile", "logical_path", "slot", "expected_class", "grammar", "role", "declaration_digest"}
	if err := exactFields(raw, context, required, nil); err != nil {
		return nil, err
	}
	fields, err := decodeStringFields(raw, context, required, nil)
	if err != nil {
		return nil, err
	}
	return GeneratedArtifactPayload{Profile: fields["profile"], LogicalPath: fields["logical_path"], Slot: fields["slot"], ExpectedClass: fields["expected_class"], Grammar: fields["grammar"], Role: fields["role"], DeclarationDigest: ID(fields["declaration_digest"])}, nil
}

func decodeInteropBoundary(raw map[string]any) (NodePayload, error) {
	const context = "interop_boundary payload"
	if err := exactFields(raw, context, []string{"profile", "mode", "provider_language_classes", "consumer_language_classes", "contract_digest"}, []string{"abi", "runtime", "protocol_schema", "interface_contract", "calling_convention", "link_load_semantics", "platform_role_names"}); err != nil {
		return nil, err
	}
	fields, err := decodeStringFields(raw, context, []string{"profile", "mode", "contract_digest"}, []string{"abi", "runtime", "protocol_schema", "interface_contract", "calling_convention", "link_load_semantics"})
	if err != nil {
		return nil, err
	}
	p := InteropBoundaryPayload{Profile: fields["profile"], Mode: InteropMode(fields["mode"]), ContractDigest: ID(fields["contract_digest"]), ABI: fields["abi"], Runtime: fields["runtime"], ProtocolSchema: fields["protocol_schema"], InterfaceContract: fields["interface_contract"], CallingConvention: fields["calling_convention"], LinkLoadSemantics: fields["link_load_semantics"]}
	p.ProviderLanguageClasses, err = requiredStringSlice(raw, "provider_language_classes", context)
	if err != nil {
		return nil, err
	}
	p.ConsumerLanguageClasses, err = requiredStringSlice(raw, "consumer_language_classes", context)
	if err != nil {
		return nil, err
	}
	p.PlatformRoleNames, err = parseRoles(raw, context)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func decodeToolchainComponent(raw map[string]any) (NodePayload, error) {
	const context = "toolchain_component payload"
	if err := exactFields(raw, context, []string{"component_role", "content_fingerprint", "executable_relative_path", "platform_abi", "policy_selector", "version_output"}, []string{"link_fingerprint_ids", "sdk_facts_digest", "time_of_use_recheck_rule", "execution_domain", "platform_role_names"}); err != nil {
		return nil, err
	}
	fields, err := decodeStringFields(raw, context, []string{"component_role", "content_fingerprint", "executable_relative_path", "platform_abi", "policy_selector", "version_output"}, []string{"sdk_facts_digest", "time_of_use_recheck_rule", "execution_domain"})
	if err != nil {
		return nil, err
	}
	p := ToolchainComponentPayload{ComponentRole: fields["component_role"], ContentFingerprint: ID(fields["content_fingerprint"]), ExecutableRelativePath: fields["executable_relative_path"], PlatformABI: fields["platform_abi"], PolicySelector: fields["policy_selector"], VersionOutput: fields["version_output"], SDKFactsDigest: ID(fields["sdk_facts_digest"]), TimeOfUseRecheckRule: fields["time_of_use_recheck_rule"], ExecutionDomain: ExecutionDomain(fields["execution_domain"])}
	if values, err := optionalStringSlice(raw, "link_fingerprint_ids", context); err != nil {
		return nil, err
	} else if values != nil {
		p.LinkFingerprintIDs = make([]ID, len(values))
		for i, value := range values {
			p.LinkFingerprintIDs[i] = ID(value)
		}
	}
	roles, err := parseRoles(raw, context)
	if err != nil {
		return nil, err
	}
	p.PlatformRoleNames = roles
	return p, nil
}

func decodeTargetPlatform(raw map[string]any) (NodePayload, error) {
	const context = "target_platform payload"
	if err := exactFields(raw, context, []string{"os", "architecture", "abi", "libc", "minimum_runtime", "sdk_id", "target_triple"}, []string{"runtime", "language_modes", "tuning"}); err != nil {
		return nil, err
	}
	fields, err := decodeStringFields(raw, context, []string{"os", "architecture", "abi", "libc", "minimum_runtime", "sdk_id", "target_triple"}, []string{"runtime"})
	if err != nil {
		return nil, err
	}
	p := TargetPlatformPayload{OS: fields["os"], Architecture: fields["architecture"], ABI: fields["abi"], Libc: fields["libc"], MinimumRuntime: fields["minimum_runtime"], SDKID: fields["sdk_id"], TargetTriple: fields["target_triple"], Runtime: fields["runtime"]}
	if _, present := raw["language_modes"]; present {
		p.LanguageModes, err = requiredStringMap(raw, "language_modes", context)
		if err != nil {
			return nil, err
		}
	}
	if _, present := raw["tuning"]; present {
		p.Tuning, err = requiredStringMap(raw, "tuning", context)
		if err != nil {
			return nil, err
		}
	}
	return p, nil
}

func decodeOutputArtifact(raw map[string]any) (NodePayload, error) {
	const context = "output_artifact payload"
	if err := exactFields(raw, context, []string{"profile", "logical_path", "expected_class", "output_role"}, []string{"compatibility_predicate", "declaration_digest", "platform_role_names"}); err != nil {
		return nil, err
	}
	fields, err := decodeStringFields(raw, context, []string{"profile", "logical_path", "expected_class", "output_role"}, []string{"compatibility_predicate", "declaration_digest"})
	if err != nil {
		return nil, err
	}
	p := OutputArtifactPayload{Profile: fields["profile"], LogicalPath: fields["logical_path"], ExpectedClass: fields["expected_class"], OutputRole: fields["output_role"], CompatibilityPredicate: fields["compatibility_predicate"], DeclarationDigest: ID(fields["declaration_digest"])}
	roles, err := parseRoles(raw, context)
	if err != nil {
		return nil, err
	}
	p.PlatformRoleNames = roles
	return p, nil
}

func sortNodes(nodes []Node) []Node {
	result := append([]Node(nil), nodes...)
	sort.Slice(result, func(i, j int) bool {
		leftID, _ := result[i].ID()
		rightID, _ := result[j].ID()
		if result[i].Kind != result[j].Kind {
			return result[i].Kind < result[j].Kind
		}
		if result[i].LogicalKey != result[j].LogicalKey {
			return result[i].LogicalKey < result[j].LogicalKey
		}
		return leftID < rightID
	})
	return result
}
