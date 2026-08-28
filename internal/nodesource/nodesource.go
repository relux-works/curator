// Package nodesource implements the manager-independent, pure-source Node and
// TypeScript graph bridge. Package-manager adapters provide normalized lock
// instances; this package owns the common capture, selection, binding and
// declared-generation policy.
package nodesource

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
)

// ProfileID is the manager-independent pure-source Node graph profile.
const ProfileID = "node-source-v1"

// Stable common-profile diagnostic codes.
const (
	CodeHookUndeclared          = "closure_hook_undeclared"
	CodeNativeBuildUnsupported  = "closure_native_build_unsupported"
	CodeBuildDependencyUnlocked = "closure_build_dependency_unlocked"
	CodeGeneratedOutputDrift    = "closure_generated_output_drift"
	CodeRuntimeIdentityChanged  = "closure_runtime_identity_changed"
	CodeManagerPluginUndeclared = "closure_manager_plugin_undeclared"
)

// ManagerProfile identifies a parser/materializer profile without changing
// the common canonical graph semantics.
type ManagerProfile string

// Supported manager schema profiles normalize into the same capture graph.
const (
	ManagerNPM         ManagerProfile = "npm-v1"
	ManagerPNPM        ManagerProfile = "pnpm-v1"
	ManagerYarnClassic ManagerProfile = "yarn-classic-v1"
	ManagerYarnModern  ManagerProfile = "yarn-modern-v1"
)

// Failure carries the stable common diagnostic code.
type Failure struct{ Code, Detail string }

func (e *Failure) Error() string { return e.Code + ": " + e.Detail }
func fail(code any, format string, args ...any) error {
	return &Failure{Code: fmt.Sprint(code), Detail: fmt.Sprintf(format, args...)}
}

// ErrorCode returns the stable common-profile diagnostic carried by err.
func ErrorCode(err error) string {
	if value, ok := err.(*Failure); ok {
		return value.Code
	}
	return ""
}

// Dependency is one normalized lock edge. Conditions are retained unevaluated
// in capture and evaluated only while projecting an active graph.
type Dependency struct {
	PackageKey       string
	Scope            closuregraph.RequirementScope
	Condition        *closuregraph.Condition
	DeclarationField string
}

// PackageInstance is the common lock-instance projection. PeerKey and
// WorkspacePath distinguish instances while target/runtime facts do not.
type PackageInstance struct {
	Key, Name, Version, Origin, Checksum, PeerKey, WorkspacePath string
	ArtifactManifestID, SnapshotDigest                           closuregraph.ID
	Dependencies                                                 []Dependency
	LifecycleScripts                                             []string
	BindingGYP, NativeBuild, BundledDependencies                 bool
	ManagerExtensions                                            []string
}

// CaptureInput is normalized by a manager-specific parser before it reaches
// the common bridge.
type CaptureInput struct {
	Manager          ManagerProfile
	RootKeys         []string
	Packages         []PackageInstance
	ShippedGenerated []ShippedGeneratedText
	PolicyIDs        []string
}

// ShippedGeneratedText is immutable generated JavaScript already present in
// admitted package bytes. It is source evidence, never a local build output.
type ShippedGeneratedText struct {
	PackageKey, Path, Grammar string
	ArtifactManifestID        closuregraph.ID
	TreeDigest                closuregraph.ID
}

// Capture is the shared selection-neutral graph and its record tables.
type Capture struct {
	Graph          closuregraph.CaptureGraph
	Nodes          []closuregraph.Node
	Edges          []closuregraph.Edge
	PackageNodeIDs map[string]closuregraph.ID
	ProductNodeIDs map[string]closuregraph.ID
	SourceNodeIDs  map[string]closuregraph.ID
	Manager        ManagerProfile
	generated      []generatedBinding
	runtimeActions []runtimeActionBinding
}

type generatedBinding struct {
	ActionID closuregraph.ID
	OwnerID  closuregraph.ID
	Tool     ToolIdentity
}

type runtimeActionBinding struct {
	ActionID closuregraph.ID
	OwnerID  closuregraph.ID
	ToolRole string
}

// RuntimeAction is an install or invocation operation that must occupy an
// immutable C5 action slot. Unlike a generator it declares no new source or
// output node; the executor permit owns its concrete read/write evidence.
type RuntimeAction struct {
	Name, Subtype, WorkingDirectory, EnvironmentPolicyID, ProcessPolicyID string
	OwnerNodeID                                                           closuregraph.ID
	ToolRole                                                              string
	ArgvTemplate                                                          []string
}

// BuildCapture maps package, peer, workspace and conditional declarations to
// one canonical graph. Manager identity is deliberately normalized to "node"
// so equivalent manager profiles produce the same capture identity.
func BuildCapture(input CaptureInput) (Capture, error) {
	if !validManager(input.Manager) {
		return Capture{}, fail(closuregraph.CodeGraphSchemaUnsupported, "unsupported manager profile %q", input.Manager)
	}
	packages := append([]PackageInstance(nil), input.Packages...)
	sort.Slice(packages, func(i, j int) bool { return packages[i].Key < packages[j].Key })
	nodes := make([]closuregraph.Node, 0, len(packages))
	ids := make(map[string]closuregraph.ID, len(packages))
	sourceIDs := make(map[string]closuregraph.ID, len(packages))
	manifests := make([]closuregraph.ID, 0, len(packages))
	manifestSet := make(map[closuregraph.ID]bool, len(packages))
	for i, pkg := range packages {
		if i > 0 && packages[i-1].Key == pkg.Key {
			return Capture{}, fail(closuregraph.CodeGraphReferenceInvalid, "duplicate package key %q", pkg.Key)
		}
		if err := validatePackage(pkg); err != nil {
			return Capture{}, err
		}
		node := closuregraph.Node{Kind: closuregraph.NodePackageInstance, LogicalKey: "node.package." + pkg.Key, Payload: closuregraph.PackageInstancePayload{
			Profile: ProfileID, Ecosystem: "node", Manager: "common", NormalizedSourceID: normalizedSource(pkg), Origin: pkg.Origin,
			LockInstanceKey: pkg.Key, Name: pkg.Name, Version: pkg.Version, ArtifactManifestID: pkg.ArtifactManifestID,
			SnapshotDigest: pkg.SnapshotDigest, WorkspacePath: pkg.WorkspacePath, TrustRole: closuregraph.TrustDependencyInput,
		}}
		id, err := node.ID()
		if err != nil {
			return Capture{}, err
		}
		nodes, ids[pkg.Key] = append(nodes, node), id
		if !manifestSet[pkg.ArtifactManifestID] {
			manifests = append(manifests, pkg.ArtifactManifestID)
			manifestSet[pkg.ArtifactManifestID] = true
		}
		source := closuregraph.Node{Kind: closuregraph.NodeSourceSet, LogicalKey: "node.source." + pkg.Key, Payload: closuregraph.SourceSetPayload{Profile: ProfileID, Origin: pkg.Origin, ArtifactManifestID: pkg.ArtifactManifestID, Projection: []string{}, Grammar: "node-package-source-v1", TrustRole: closuregraph.TrustDependencyInput, SourceClass: "source.tree", TreeDigest: pkg.SnapshotDigest}}
		sourceID, sourceErr := source.ID()
		if sourceErr != nil {
			return Capture{}, sourceErr
		}
		nodes, sourceIDs[pkg.Key] = append(nodes, source), sourceID
	}
	edges := []closuregraph.Edge{}
	for _, pkg := range packages {
		edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgeResolvesTo, EdgeKey: "node.resolves." + pkg.Key, FromNodeID: ids[pkg.Key], ToNodeID: sourceIDs[pkg.Key], Payload: closuregraph.ResolvesToPayload{LockField: "packages." + pkg.Key, Origin: closuregraph.EvidenceOrigin{Field: "packages." + pkg.Key + ".resolved"}, Checksum: pkg.Checksum, ArtifactManifestID: pkg.ArtifactManifestID}})
	}
	shipped := append([]ShippedGeneratedText(nil), input.ShippedGenerated...)
	sort.Slice(shipped, func(i, j int) bool {
		return shipped[i].PackageKey+"\x00"+shipped[i].Path < shipped[j].PackageKey+"\x00"+shipped[j].Path
	})
	for index, item := range shipped {
		packageID, ok := ids[item.PackageKey]
		if !ok || item.Path == "" || item.Grammar == "" || !item.ArtifactManifestID.Valid() || !item.TreeDigest.Valid() {
			return Capture{}, fail(closuregraph.CodeGraphIncomplete, "shipped generated text declaration is incomplete")
		}
		if index > 0 && shipped[index-1].PackageKey == item.PackageKey && shipped[index-1].Path == item.Path {
			return Capture{}, fail(closuregraph.CodeGraphReferenceInvalid, "duplicate shipped generated text %s", item.Path)
		}
		node := closuregraph.Node{Kind: closuregraph.NodeSourceSet, LogicalKey: fmt.Sprintf("node.shipped-generated.%s.%04d", item.PackageKey, index), Payload: closuregraph.SourceSetPayload{Profile: ProfileID, Origin: "package:" + item.PackageKey, ArtifactManifestID: item.ArtifactManifestID, Projection: []string{item.Path}, Grammar: item.Grammar, TrustRole: closuregraph.TrustDependencyInput, SourceClass: "source.generated_text", TreeDigest: item.TreeDigest}}
		nodeID, nodeErr := node.ID()
		if nodeErr != nil {
			return Capture{}, nodeErr
		}
		nodes = append(nodes, node)
		edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgeDeclares, EdgeKey: fmt.Sprintf("node.shipped-generated.%04d", index), FromNodeID: packageID, ToNodeID: nodeID, Payload: closuregraph.DeclaresPayload{Origin: closuregraph.EvidenceOrigin{Field: "files." + item.Path}}})
		present := false
		for _, manifestID := range manifests {
			present = present || manifestID == item.ArtifactManifestID
		}
		if !present {
			manifests = append(manifests, item.ArtifactManifestID)
			manifestSet[item.ArtifactManifestID] = true
		}
	}
	for _, pkg := range packages {
		from := ids[pkg.Key]
		dependencies, err := canonicalDependencies(pkg)
		if err != nil {
			return Capture{}, err
		}
		for index, dep := range dependencies {
			to, ok := ids[dep.PackageKey]
			if !ok {
				return Capture{}, fail(closuregraph.CodeGraphIncomplete, "%s requires missing package %s", pkg.Key, dep.PackageKey)
			}
			field := dep.DeclarationField
			if field == "" {
				field = "dependencies"
			}
			edge := closuregraph.Edge{Kind: closuregraph.EdgeRequires, EdgeKey: fmt.Sprintf("node.requires.%s.%04d", pkg.Key, index), FromNodeID: from, ToNodeID: to, Payload: closuregraph.RequiresPayload{Scope: dep.Scope, Condition: dep.Condition, Origin: closuregraph.EvidenceOrigin{Field: field}, DependencyKind: dependencyKind(pkg, dep)}}
			if err := edge.Validate(); err != nil {
				return Capture{}, err
			}
			edges = append(edges, edge)
		}
	}
	rootKeys := append([]string(nil), input.RootKeys...)
	sort.Strings(rootKeys)
	for i, key := range rootKeys {
		if key == "" || (i > 0 && rootKeys[i-1] == key) {
			return Capture{}, fail(closuregraph.CodeGraphReferenceInvalid, "root package keys must be non-empty and unique")
		}
	}
	roots := make([]closuregraph.ID, len(rootKeys))
	products := make(map[string]closuregraph.ID, len(rootKeys))
	for i, key := range rootKeys {
		packageID, ok := ids[key]
		if !ok {
			return Capture{}, fail(closuregraph.CodeGraphIncomplete, "root package %q is missing", key)
		}
		declaration, _ := closuregraph.DomainID("node-command-declaration-v1", map[string]any{"package_key": key})
		product := closuregraph.Node{Kind: closuregraph.NodeCommandProduct, LogicalKey: "node.product." + key, Payload: closuregraph.CommandProductPayload{Profile: ProfileID, SkillKey: key, CommandKey: key, EntryPointContract: "node-runtime-entry-v1", DeclarationDigest: declaration, PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformTarget}}}
		productID, productErr := product.ID()
		if productErr != nil {
			return Capture{}, productErr
		}
		nodes = append(nodes, product)
		products[key], roots[i] = productID, productID
		edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgeRequires, EdgeKey: fmt.Sprintf("node.product.requires.%04d", i), FromNodeID: productID, ToNodeID: packageID, Payload: closuregraph.RequiresPayload{Scope: closuregraph.ScopeRuntime, Origin: closuregraph.EvidenceOrigin{Field: "root.package"}, DependencyKind: "entry-package"}})
	}
	graph, err := closuregraph.NewCaptureGraph(ProfileID, input.PolicyIDs, roots, nodes, edges, manifests)
	if err != nil {
		return Capture{}, err
	}
	return Capture{Graph: graph, Nodes: nodes, Edges: edges, PackageNodeIDs: ids, ProductNodeIDs: products, SourceNodeIDs: sourceIDs, Manager: input.Manager}, nil
}

func canonicalDependencies(pkg PackageInstance) ([]Dependency, error) {
	values := append([]Dependency(nil), pkg.Dependencies...)
	key := func(dep Dependency) (string, error) {
		field := dep.DeclarationField
		if field == "" {
			field = "dependencies"
		}
		condition := any(nil)
		if dep.Condition != nil {
			condition = map[string]any{"evaluator_id": dep.Condition.EvaluatorID, "expression": dep.Condition.Expression}
		}
		encoded, err := closuregraph.DomainID("node-dependency-declaration-v1", map[string]any{
			"condition": condition, "declaration_field": field, "dependency_kind": dependencyKind(pkg, dep),
			"package_key": dep.PackageKey, "scope": string(dep.Scope),
		})
		return string(encoded), err
	}
	for index := range values {
		if _, err := key(values[index]); err != nil {
			return nil, err
		}
	}
	sort.Slice(values, func(i, j int) bool {
		left, _ := key(values[i])
		right, _ := key(values[j])
		return left < right
	})
	previous := ""
	for index, dep := range values {
		semantic, err := key(dep)
		if err != nil {
			return nil, err
		}
		if index > 0 && semantic == previous {
			return nil, fail(closuregraph.CodeGraphReferenceInvalid, "package %s has a duplicate dependency declaration for %s", pkg.Key, dep.PackageKey)
		}
		previous = semantic
	}
	return values, nil
}

func validManager(value ManagerProfile) bool {
	switch value {
	case ManagerNPM, ManagerPNPM, ManagerYarnClassic, ManagerYarnModern:
		return true
	}
	return false
}
func normalizedSource(pkg PackageInstance) string {
	if pkg.WorkspacePath != "" {
		return "workspace:" + pkg.WorkspacePath
	}
	return pkg.Origin + "#" + pkg.Checksum
}
func dependencyKind(pkg PackageInstance, dep Dependency) string {
	if dep.Scope == closuregraph.ScopePeer {
		return "peer:" + pkg.PeerKey
	}
	if dep.Scope == closuregraph.ScopeWorkspace {
		return "workspace"
	}
	return "package"
}

func validatePackage(pkg PackageInstance) error {
	for _, value := range []string{pkg.Key, pkg.Name, pkg.Version, pkg.Origin, pkg.Checksum} {
		if value == "" || strings.ContainsAny(value, "\x00\r\n") {
			return fail(closuregraph.CodeGraphIncomplete, "package identity field is missing or non-portable")
		}
	}
	if !pkg.ArtifactManifestID.Valid() || !pkg.SnapshotDigest.Valid() {
		return fail(closuregraph.CodeGraphIncomplete, "package %s lacks immutable admission evidence", pkg.Key)
	}
	if pkg.WorkspacePath != "" && (path.IsAbs(pkg.WorkspacePath) || path.Clean(pkg.WorkspacePath) != pkg.WorkspacePath || pkg.WorkspacePath == ".." || strings.HasPrefix(pkg.WorkspacePath, "../")) {
		return fail("closure_local_path_escape", "workspace %s escapes the capture root", pkg.WorkspacePath)
	}
	if pkg.BundledDependencies {
		return fail("closure_bundled_dependency_unsupported", "package %s declares bundled dependencies", pkg.Key)
	}
	if len(pkg.ManagerExtensions) != 0 {
		return fail(CodeManagerPluginUndeclared, "package %s uses manager extension %s", pkg.Key, pkg.ManagerExtensions[0])
	}
	if pkg.BindingGYP || pkg.NativeBuild {
		return fail(CodeNativeBuildUnsupported, "package %s requires a native build", pkg.Key)
	}
	if len(pkg.LifecycleScripts) != 0 {
		return fail(CodeHookUndeclared, "dependency %s declares lifecycle script %s", pkg.Key, pkg.LifecycleScripts[0])
	}
	return nil
}

// ToolIdentity is a full C0/C4 fingerprinted Node or manager executable.
type ToolIdentity struct {
	Role, PolicySelector, ExecutableRelativePath, VersionOutput, PlatformABI string
	Fingerprint                                                              closuregraph.ID
	ExecutableSHA256                                                         closuregraph.ID
	ExecutionDomain                                                          closuregraph.ExecutionDomain
	// EntrypointRelativePath identifies an interpreted tool entry point, such
	// as npm-cli.js, below the selected external toolchain root. ReadRoots are
	// the complete toolchain roots that the invocation may consume.
	EntrypointRelativePath string
	ReadRoots              []string
}

// RuntimeBinding is the exact selection-specific platform/runtime overlay.
type RuntimeBinding struct {
	Platform      closuregraph.TargetPlatformPayload
	Node, Manager ToolIdentity
	TargetNodeIDs []closuregraph.ID
	C0Checkpoint  *closuregraph.Checkpoint
}

// NewC0Checkpoint binds the selected Node and manager executables before any
// executable resolution or metadata derivation may run.
func NewC0Checkpoint(capture Capture, selection closuregraph.SelectionContext, exact RuntimeBinding) (closuregraph.Checkpoint, error) {
	selectionID, err := selection.ID()
	if err != nil {
		return closuregraph.Checkpoint{}, err
	}
	platform := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: exact.Platform}
	platformID, err := platform.ID()
	if err != nil {
		return closuregraph.Checkpoint{}, err
	}
	if selected, ok := selection.PlatformRoles[closuregraph.PlatformTarget]; !ok || selected != platformID {
		return closuregraph.Checkpoint{}, fail(closuregraph.CodeGraphReferenceInvalid, "selection target does not match C0 platform")
	}
	nodeTool, err := toolNode("node", exact.Node)
	if err != nil {
		return closuregraph.Checkpoint{}, err
	}
	managerTool, err := toolNode("manager", exact.Manager)
	if err != nil {
		return closuregraph.Checkpoint{}, err
	}
	nodeID, _ := nodeTool.ID()
	managerID, _ := managerTool.ID()
	evidenceToolIDs := []closuregraph.ID{nodeID, managerID}
	sort.Slice(evidenceToolIDs, func(i, j int) bool { return evidenceToolIDs[i] < evidenceToolIDs[j] })
	payload := closuregraph.C0ProfilePayload{
		AdapterProfileID: ProfileID, SchemaIDs: []string{"closure-graph-v1", "node-runtime-v1"},
		ArtifactPolicyID: "artifact-policy-v1", DetectorRegistryID: "artifact-detector-registry-v1",
		SourceGrammarIDs: []string{"javascript-v1", "json-v1", "typescript-v1"}, LimitVectorID: "artifact-limits-v1",
		SelectionContextID: selectionID, PlatformNodeIDs: []closuregraph.ID{platformID},
		PlatformRoles:    map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID},
		ManagerSchemaIDs: []string{string(capture.Manager)}, ConfigurationPolicyID: "node-configuration-v1",
		CapabilityIDs: []string{"lifecycle-suppressed", "network-denied", "pure-source"}, EvidenceToolchainNodeIDs: evidenceToolIDs,
	}
	if host, ok := selection.PlatformRoles[closuregraph.PlatformHost]; ok {
		if host != platformID {
			return closuregraph.Checkpoint{}, fail(closuregraph.CodeGraphReferenceInvalid, "common Node profile requires one executable host/target platform")
		}
		payload.PlatformRoles[closuregraph.PlatformHost] = host
	}
	return closuregraph.NewCheckpoint(payload, nil, nil)
}

// Bind creates only target-platform/toolchain nodes and targets/uses_tool
// edges, preserving capture identity.
func Bind(capture Capture, selection closuregraph.SelectionContext, exact RuntimeBinding) (closuregraph.SelectionBinding, []closuregraph.Node, []closuregraph.Edge, closuregraph.BindingAuthority, error) {
	if exact.C0Checkpoint == nil {
		return closuregraph.SelectionBinding{}, nil, nil, closuregraph.BindingAuthority{}, fail(closuregraph.CodeCheckpointInvalid, "Node and manager require exact C0 authority")
	}
	targetNodeIDs := append([]closuregraph.ID(nil), exact.TargetNodeIDs...)
	sort.Slice(targetNodeIDs, func(i, j int) bool { return targetNodeIDs[i] < targetNodeIDs[j] })
	if len(targetNodeIDs) == 0 {
		return closuregraph.SelectionBinding{}, nil, nil, closuregraph.BindingAuthority{}, fail(closuregraph.CodeGraphIncomplete, "runtime binding has no target nodes")
	}
	for i, id := range targetNodeIDs {
		if !id.Valid() || (i > 0 && targetNodeIDs[i-1] == id) {
			return closuregraph.SelectionBinding{}, nil, nil, closuregraph.BindingAuthority{}, fail(closuregraph.CodeGraphReferenceInvalid, "runtime target node IDs must be valid and unique")
		}
	}
	captureID, err := capture.Graph.ID()
	if err != nil {
		return closuregraph.SelectionBinding{}, nil, nil, closuregraph.BindingAuthority{}, err
	}
	selectionID, err := selection.ID()
	if err != nil {
		return closuregraph.SelectionBinding{}, nil, nil, closuregraph.BindingAuthority{}, err
	}
	platform := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: exact.Platform}
	nodeTool, err := toolNode("node", exact.Node)
	if err != nil {
		return closuregraph.SelectionBinding{}, nil, nil, closuregraph.BindingAuthority{}, err
	}
	managerTool, err := toolNode("manager", exact.Manager)
	if err != nil {
		return closuregraph.SelectionBinding{}, nil, nil, closuregraph.BindingAuthority{}, err
	}
	platformID, err := platform.ID()
	if err != nil {
		return closuregraph.SelectionBinding{}, nil, nil, closuregraph.BindingAuthority{}, err
	}
	if selection.PlatformRoles[closuregraph.PlatformTarget] != platformID {
		return closuregraph.SelectionBinding{}, nil, nil, closuregraph.BindingAuthority{}, fail(closuregraph.CodeGraphReferenceInvalid, "selection target does not match exact binding")
	}
	nodes := []closuregraph.Node{platform, nodeTool, managerTool}
	nodeID, _ := nodeTool.ID()
	managerID, _ := managerTool.ID()
	edges := []closuregraph.Edge{}
	for i, id := range targetNodeIDs {
		edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgeTargets, EdgeKey: fmt.Sprintf("node.targets.%04d", i), FromNodeID: id, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.target"}}})
	}
	for index, tool := range nodes[1:] {
		toolID, _ := tool.ID()
		edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgeRequires, EdgeKey: fmt.Sprintf("node.requires-tool.%04d", index), FromNodeID: targetNodeIDs[0], ToNodeID: toolID, Payload: closuregraph.RequiresPayload{Scope: closuregraph.ScopeToolchain, Origin: closuregraph.EvidenceOrigin{Field: []string{"runtime.node", "runtime.manager"}[index]}, DependencyKind: []string{"node-runtime", "package-manager"}[index]}})
		edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgeTargets, EdgeKey: fmt.Sprintf("node.targets-tool.%04d", index), FromNodeID: toolID, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.target"}}})
	}
	for index, generated := range capture.generated {
		selectedOwner := false
		for _, targetNodeID := range targetNodeIDs {
			selectedOwner = selectedOwner || targetNodeID == generated.OwnerID
		}
		if !selectedOwner {
			continue
		}
		tool, toolErr := toolNode("compiler."+string(generated.ActionID), generated.Tool)
		if toolErr != nil {
			return closuregraph.SelectionBinding{}, nil, nil, closuregraph.BindingAuthority{}, toolErr
		}
		toolID, _ := tool.ID()
		nodes = append(nodes, tool)
		edges = append(edges,
			closuregraph.Edge{Kind: closuregraph.EdgeUsesTool, EdgeKey: fmt.Sprintf("node.generator.tool.%04d", index), FromNodeID: generated.ActionID, ToNodeID: toolID, Payload: closuregraph.UsesToolPayload{ExecutableRelativePath: generated.Tool.ExecutableRelativePath, ToolSlot: "compiler", InvocationRole: "typescript-compiler"}},
			closuregraph.Edge{Kind: closuregraph.EdgeTargets, EdgeKey: fmt.Sprintf("node.generator.tool-target.%04d", index), FromNodeID: toolID, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformHost, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.host"}}},
			closuregraph.Edge{Kind: closuregraph.EdgeTargets, EdgeKey: fmt.Sprintf("node.generator.action-target.%04d", index), FromNodeID: generated.ActionID, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformHost, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.host"}}},
		)
		outputIndex := 0
		for _, edge := range capture.Edges {
			if edge.Kind != closuregraph.EdgeProduces || edge.FromNodeID != generated.ActionID {
				continue
			}
			for _, node := range capture.Nodes {
				nodeID, _ := node.ID()
				if nodeID == edge.ToNodeID && node.Kind == closuregraph.NodeOutputArtifact {
					edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgeTargets, EdgeKey: fmt.Sprintf("node.generator.output-target.%04d.%04d", index, outputIndex), FromNodeID: nodeID, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.target"}}})
					outputIndex++
				}
			}
		}
	}
	for index, action := range capture.runtimeActions {
		selectedOwner := false
		for _, targetNodeID := range targetNodeIDs {
			selectedOwner = selectedOwner || targetNodeID == action.OwnerID
		}
		if !selectedOwner {
			continue
		}
		toolID := managerID
		invocationRole := "package-manager"
		executable := exact.Manager.ExecutableRelativePath
		if action.ToolRole == "node-runtime" {
			toolID = nodeID
			invocationRole = "node-runtime"
			executable = exact.Node.ExecutableRelativePath
		}
		edges = append(edges,
			closuregraph.Edge{Kind: closuregraph.EdgeUsesTool, EdgeKey: fmt.Sprintf("node.runtime-action.tool.%04d", index), FromNodeID: action.ActionID, ToNodeID: toolID, Payload: closuregraph.UsesToolPayload{ExecutableRelativePath: executable, ToolSlot: "executor", InvocationRole: invocationRole}},
			closuregraph.Edge{Kind: closuregraph.EdgeTargets, EdgeKey: fmt.Sprintf("node.runtime-action.target.%04d", index), FromNodeID: action.ActionID, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.target"}}},
		)
	}
	binding, err := closuregraph.NewSelectionBinding(captureID, selectionID, nodes, edges)
	if err != nil {
		return closuregraph.SelectionBinding{}, nil, nil, closuregraph.BindingAuthority{}, err
	}
	selectors := make([]closuregraph.ToolchainSelector, 0, len(nodes)-3)
	authorities := make([]closuregraph.ToolchainBindingEvidence, 0, len(nodes)-1)
	c0ID, err := exact.C0Checkpoint.ID()
	if err != nil {
		return closuregraph.SelectionBinding{}, nil, nil, closuregraph.BindingAuthority{}, err
	}
	for index, node := range nodes[1:] {
		nodeID, _ := node.ID()
		if index < 2 {
			authorities = append(authorities, closuregraph.ToolchainBindingEvidence{NodeID: nodeID, FirstBound: closuregraph.ToolchainBoundAtC0, EvidenceID: c0ID})
			continue
		}
		selector, selectorErr := closuregraph.NewToolchainSelector(node)
		if selectorErr != nil {
			return closuregraph.SelectionBinding{}, nil, nil, closuregraph.BindingAuthority{}, selectorErr
		}
		selectors = append(selectors, selector)
		selectorID, selectorErr := selector.ID()
		if selectorErr != nil {
			return closuregraph.SelectionBinding{}, nil, nil, closuregraph.BindingAuthority{}, selectorErr
		}
		authorities = append(authorities, closuregraph.ToolchainBindingEvidence{NodeID: selector.NodeID, FirstBound: closuregraph.ToolchainBoundAtC4, EvidenceID: selectorID})
	}
	return binding, nodes, edges, closuregraph.BindingAuthority{Toolchains: authorities, C0Checkpoint: exact.C0Checkpoint, C4Selectors: selectors}, nil
}

func toolNode(key string, tool ToolIdentity) (closuregraph.Node, error) {
	if !tool.Fingerprint.Valid() || !tool.ExecutableSHA256.Valid() {
		return closuregraph.Node{}, fail(closuregraph.CodeGraphIncomplete, "tool %s lacks exact tree or executable content evidence", key)
	}
	domain := tool.ExecutionDomain
	if domain == "" {
		domain = closuregraph.ExecutionTarget
	}
	role := closuregraph.PlatformTarget
	if domain == closuregraph.ExecutionHost {
		role = closuregraph.PlatformHost
	}
	readRoots := append([]string(nil), tool.ReadRoots...)
	sort.Strings(readRoots)
	for index, root := range readRoots {
		if path.IsAbs(root) || path.Clean(root) != root || root == "." || root == ".." || strings.HasPrefix(root, "../") || (index > 0 && readRoots[index-1] == root) {
			return closuregraph.Node{}, fail(closuregraph.CodeGraphIncomplete, "tool %s has invalid external read roots", key)
		}
	}
	var invocationFacts closuregraph.ID
	if tool.EntrypointRelativePath != "" || len(readRoots) != 0 {
		if tool.EntrypointRelativePath != "" && (path.IsAbs(tool.EntrypointRelativePath) || path.Clean(tool.EntrypointRelativePath) != tool.EntrypointRelativePath || strings.HasPrefix(tool.EntrypointRelativePath, "../")) {
			return closuregraph.Node{}, fail(closuregraph.CodeGraphIncomplete, "tool %s has invalid interpreted entry point", key)
		}
		readRootValues := make([]any, len(readRoots))
		for index, root := range readRoots {
			readRootValues[index] = root
		}
		var err error
		invocationFacts, err = closuregraph.DomainID("node-tool-invocation-facts-v1", map[string]any{"entrypoint_relative_path": tool.EntrypointRelativePath, "read_roots": readRootValues})
		if err != nil {
			return closuregraph.Node{}, err
		}
	}
	node := closuregraph.Node{Kind: closuregraph.NodeToolchainComponent, LogicalKey: "node.tool." + key, Payload: closuregraph.ToolchainComponentPayload{ComponentRole: tool.Role, ContentFingerprint: tool.Fingerprint, ExecutableRelativePath: tool.ExecutableRelativePath, PlatformABI: tool.PlatformABI, PolicySelector: tool.PolicySelector, VersionOutput: tool.VersionOutput, LinkFingerprintIDs: []closuregraph.ID{tool.ExecutableSHA256}, SDKFactsDigest: invocationFacts, TimeOfUseRecheckRule: "immediate-exact-v1", ExecutionDomain: domain, PlatformRoleNames: []closuregraph.PlatformRole{role}}}
	return node, node.Validate()
}

// GeneratedAction is a fully declared C5 TypeScript/generator lineage.
type GeneratedAction struct {
	Name                                 string
	Argv                                 []string
	WorkingDirectory                     string
	Compiler                             ToolIdentity
	Inputs                               []GeneratedInput
	EnvironmentPolicyID, ProcessPolicyID string
	TargetNodeID                         closuregraph.ID
	Outputs                              []GeneratedOutput
}

// GeneratedInput binds one exact admitted source/config/plugin read.
type GeneratedInput struct {
	NodeID            closuregraph.ID
	Path, Class, Role string
}

// GeneratedOutput declares one immutable intermediate or published output.
type GeneratedOutput struct {
	Path, Class, Grammar, Role string
	DeclarationDigest          closuregraph.ID
	Intermediate               bool
}

func normalizeGeneratedOutput(output GeneratedOutput) (GeneratedOutput, error) {
	if output.Role == "" {
		output.Role = "runtime_output"
	}
	if output.Path == "" || output.Class == "" || output.Grammar == "" || output.Role == "" {
		return GeneratedOutput{}, fail(CodeGeneratedOutputDrift, "generated output lacks complete path/class/grammar/role evidence")
	}
	digest, err := closuregraph.DomainID("node-generated-output-declaration-v1", map[string]any{
		"class": output.Class, "grammar": output.Grammar, "intermediate": output.Intermediate,
		"path": output.Path, "role": output.Role,
	})
	if err != nil {
		return GeneratedOutput{}, err
	}
	if output.DeclarationDigest != "" && output.DeclarationDigest != digest {
		return GeneratedOutput{}, fail(CodeGeneratedOutputDrift, "generated output %s declaration identity differs", output.Path)
	}
	output.DeclarationDigest = digest
	return output, nil
}

// BuildGeneratedAction rejects incomplete lineage and returns immutable action,
// generated-output and typed slot records suitable for C5 planning.
func BuildGeneratedAction(spec GeneratedAction) (closuregraph.Node, []closuregraph.Node, []closuregraph.Edge, error) {
	if len(spec.Inputs) == 0 || len(spec.Outputs) == 0 || !spec.TargetNodeID.Valid() {
		return closuregraph.Node{}, nil, nil, fail(CodeBuildDependencyUnlocked, "generator %s has incomplete config/source/target/output lineage", spec.Name)
	}
	if len(spec.Argv) == 0 {
		return closuregraph.Node{}, nil, nil, fail(CodeHookUndeclared, "generator %s has no declared argv", spec.Name)
	}
	if _, err := toolNode("compiler."+spec.Name, spec.Compiler); err != nil {
		return closuregraph.Node{}, nil, nil, err
	}
	inputs := append([]GeneratedInput(nil), spec.Inputs...)
	sort.Slice(inputs, func(i, j int) bool {
		left := inputs[i].Role + "\x00" + inputs[i].Path + "\x00" + string(inputs[i].NodeID)
		right := inputs[j].Role + "\x00" + inputs[j].Path + "\x00" + string(inputs[j].NodeID)
		return left < right
	})
	readSlots := make([]string, len(inputs))
	argv := append([]string{"$TOOL(compiler)"}, spec.Argv...)
	seenInputs := map[closuregraph.ID]bool{}
	for i := range inputs {
		if !inputs[i].NodeID.Valid() || inputs[i].Path == "" || inputs[i].Class == "" || inputs[i].Role == "" || seenInputs[inputs[i].NodeID] {
			return closuregraph.Node{}, nil, nil, fail(CodeBuildDependencyUnlocked, "generator %s has an invalid or duplicate declared input", spec.Name)
		}
		seenInputs[inputs[i].NodeID] = true
		readSlots[i] = fmt.Sprintf("%s-%04d", inputs[i].Role, i)
		argv = append(argv, "$READ("+readSlots[i]+")")
	}
	outputs := append([]GeneratedOutput(nil), spec.Outputs...)
	writeSlots := make([]string, len(outputs))
	seenOutputPaths := map[string]bool{}
	for i := range outputs {
		var err error
		outputs[i], err = normalizeGeneratedOutput(outputs[i])
		if err != nil {
			return closuregraph.Node{}, nil, nil, err
		}
		if seenOutputPaths[outputs[i].Path] {
			return closuregraph.Node{}, nil, nil, fail(CodeGeneratedOutputDrift, "generator %s has duplicate output path %s", spec.Name, outputs[i].Path)
		}
		seenOutputPaths[outputs[i].Path] = true
		writeSlots[i] = fmt.Sprintf("output-%04d", i)
		argv = append(argv, "$WRITE("+writeSlots[i]+")")
	}
	action := closuregraph.Node{Kind: closuregraph.NodeAction, LogicalKey: "node.action." + spec.Name, Payload: closuregraph.ActionPayload{Profile: ProfileID, ActionSubtype: "typescript-generator", ExecutionDomain: closuregraph.ExecutionHost, ArgvTemplate: argv, WorkingDirectoryTemplate: spec.WorkingDirectory, ToolSlotNames: []string{"compiler"}, ReadSlotNames: readSlots, WriteSlotNames: writeSlots, EnvironmentPolicyID: spec.EnvironmentPolicyID, ProcessPolicyID: spec.ProcessPolicyID, Network: "none", PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformHost}}}
	if err := action.Validate(); err != nil {
		return closuregraph.Node{}, nil, nil, err
	}
	actionID, _ := action.ID()
	edges := []closuregraph.Edge{{Kind: closuregraph.EdgeDeclares, EdgeKey: "node.generator.declares." + spec.Name, FromNodeID: spec.TargetNodeID, ToNodeID: actionID, Payload: closuregraph.DeclaresPayload{Origin: closuregraph.EvidenceOrigin{Field: "generators." + spec.Name}}}}
	nodes := []closuregraph.Node{}
	for i, input := range inputs {
		edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgeReads, EdgeKey: fmt.Sprintf("node.generator.read.%s.%04d", spec.Name, i), FromNodeID: actionID, ToNodeID: input.NodeID, Payload: closuregraph.ReadsPayload{Path: input.Path, ReadSlot: readSlots[i], ReadClass: input.Class}})
	}
	for i, output := range outputs {
		var node closuregraph.Node
		if output.Intermediate {
			node = closuregraph.Node{Kind: closuregraph.NodeGeneratedArtifact, LogicalKey: fmt.Sprintf("node.generated.%s.%04d", spec.Name, i), Payload: closuregraph.GeneratedArtifactPayload{Profile: ProfileID, LogicalPath: output.Path, Slot: writeSlots[i], ExpectedClass: output.Class, Grammar: output.Grammar, Role: output.Role, DeclarationDigest: output.DeclarationDigest}}
		} else {
			node = closuregraph.Node{Kind: closuregraph.NodeOutputArtifact, LogicalKey: fmt.Sprintf("node.output.%s.%04d", spec.Name, i), Payload: closuregraph.OutputArtifactPayload{Profile: ProfileID, LogicalPath: output.Path, ExpectedClass: output.Class, OutputRole: output.Role, CompatibilityPredicate: "target-match-v1", DeclarationDigest: output.DeclarationDigest, PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformTarget}}}
		}
		if err := node.Validate(); err != nil {
			return closuregraph.Node{}, nil, nil, err
		}
		id, _ := node.ID()
		nodes = append(nodes, node)
		edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgeProduces, EdgeKey: fmt.Sprintf("node.generator.produces.%s.%04d", spec.Name, i), FromNodeID: actionID, ToNodeID: id, Payload: closuregraph.ProducesPayload{Path: output.Path, WriteSlot: writeSlots[i], WriteClass: output.Class}})
		if !output.Intermediate {
			edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgePublishes, EdgeKey: fmt.Sprintf("node.generator.publishes.%s.%04d", spec.Name, i), FromNodeID: spec.TargetNodeID, ToNodeID: id, Payload: closuregraph.PublishesPayload{Destination: output.Path, EntryPoint: output.Path}})
		}
	}
	return action, nodes, edges, nil
}

// AddGeneratedActions closes declared TypeScript/generator lineage into the
// selection-neutral graph. It never adds a concrete tool or target record.
func AddGeneratedActions(capture Capture, specs []GeneratedAction) (Capture, error) {
	nodes := append([]closuregraph.Node(nil), capture.Nodes...)
	edges := append([]closuregraph.Edge(nil), capture.Edges...)
	generated := append([]generatedBinding(nil), capture.generated...)
	known := map[closuregraph.ID]bool{}
	for _, node := range nodes {
		id, err := node.ID()
		if err != nil {
			return Capture{}, err
		}
		known[id] = true
	}
	ordered := append([]GeneratedAction(nil), specs...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Name < ordered[j].Name })
	type declaredAction struct {
		spec    GeneratedAction
		action  closuregraph.Node
		outputs []closuregraph.Node
		edges   []closuregraph.Edge
	}
	declared := make([]declaredAction, 0, len(ordered))
	outputPaths := map[string]bool{}
	for index, spec := range ordered {
		if index > 0 && ordered[index-1].Name == spec.Name {
			return Capture{}, fail(closuregraph.CodeGraphReferenceInvalid, "duplicate generator declaration %q", spec.Name)
		}
		action, outputs, actionEdges, err := BuildGeneratedAction(spec)
		if err != nil {
			return Capture{}, err
		}
		actionID, _ := action.ID()
		known[actionID] = true
		for _, output := range outputs {
			outputID, idErr := output.ID()
			if idErr != nil {
				return Capture{}, idErr
			}
			var logicalPath string
			switch payload := output.Payload.(type) {
			case closuregraph.GeneratedArtifactPayload:
				logicalPath = payload.LogicalPath
			case closuregraph.OutputArtifactPayload:
				logicalPath = payload.LogicalPath
			}
			if outputPaths[logicalPath] {
				return Capture{}, fail(CodeGeneratedOutputDrift, "duplicate generated output path %s", logicalPath)
			}
			outputPaths[logicalPath] = true
			known[outputID] = true
		}
		declared = append(declared, declaredAction{spec: spec, action: action, outputs: outputs, edges: actionEdges})
	}
	// All declared actions and outputs are indexed before references are checked,
	// so later actions may consume earlier generated artifacts independent of
	// caller declaration order. The shared planner remains the cycle authority.
	for _, item := range declared {
		if !known[item.spec.TargetNodeID] {
			return Capture{}, fail(closuregraph.CodeGraphReferenceInvalid, "generator %s target is absent from capture", item.spec.Name)
		}
		for _, input := range item.spec.Inputs {
			if !known[input.NodeID] {
				return Capture{}, fail(CodeBuildDependencyUnlocked, "generator %s input is absent from declared closure", item.spec.Name)
			}
		}
		actionID, _ := item.action.ID()
		nodes = append(nodes, item.action)
		nodes = append(nodes, item.outputs...)
		edges = append(edges, item.edges...)
		generated = append(generated, generatedBinding{ActionID: actionID, OwnerID: item.spec.TargetNodeID, Tool: item.spec.Compiler})
	}
	graph, err := closuregraph.NewCaptureGraph(ProfileID, capture.Graph.PolicyIDs, capture.Graph.RootNodeIDs, nodes, edges, capture.Graph.ArtifactManifestIDs)
	if err != nil {
		return Capture{}, err
	}
	capture.Graph, capture.Nodes, capture.Edges, capture.generated = graph, nodes, edges, generated
	return capture, nil
}

// AddRuntimeActions adds manager/runtime operations before C4 so C5, rather
// than an adapter-local digest, is the authority for every later process.
func AddRuntimeActions(capture Capture, specs []RuntimeAction) (Capture, error) {
	nodes := append([]closuregraph.Node(nil), capture.Nodes...)
	edges := append([]closuregraph.Edge(nil), capture.Edges...)
	bindings := append([]runtimeActionBinding(nil), capture.runtimeActions...)
	ordered := append([]RuntimeAction(nil), specs...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Name < ordered[j].Name })
	known := map[closuregraph.ID]bool{}
	for _, node := range nodes {
		id, err := node.ID()
		if err != nil {
			return Capture{}, err
		}
		known[id] = true
	}
	for index, spec := range ordered {
		if spec.Name == "" || spec.Subtype == "" || !known[spec.OwnerNodeID] || len(spec.ArgvTemplate) == 0 ||
			spec.WorkingDirectory == "" || spec.EnvironmentPolicyID == "" || spec.ProcessPolicyID == "" {
			return Capture{}, fail(CodeHookUndeclared, "runtime action %q is incomplete", spec.Name)
		}
		if index > 0 && ordered[index-1].Name == spec.Name {
			return Capture{}, fail(closuregraph.CodeGraphReferenceInvalid, "duplicate runtime action %q", spec.Name)
		}
		if spec.ToolRole != "package-manager" && spec.ToolRole != "node-runtime" {
			return Capture{}, fail(CodeHookUndeclared, "runtime action %q has unsupported tool role %q", spec.Name, spec.ToolRole)
		}
		action := closuregraph.Node{Kind: closuregraph.NodeAction, LogicalKey: "node.runtime-action." + spec.Name, Payload: closuregraph.ActionPayload{
			Profile: ProfileID, ActionSubtype: spec.Subtype, ExecutionDomain: closuregraph.ExecutionTarget,
			ArgvTemplate: append([]string{"$TOOL(executor)"}, spec.ArgvTemplate...), WorkingDirectoryTemplate: spec.WorkingDirectory,
			ToolSlotNames: []string{"executor"}, ReadSlotNames: []string{}, WriteSlotNames: []string{},
			EnvironmentPolicyID: spec.EnvironmentPolicyID, ProcessPolicyID: spec.ProcessPolicyID, Network: "none",
			PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformTarget},
		}}
		if err := action.Validate(); err != nil {
			return Capture{}, err
		}
		actionID, _ := action.ID()
		nodes = append(nodes, action)
		edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgeDeclares, EdgeKey: fmt.Sprintf("node.runtime-action.declares.%04d", index), FromNodeID: spec.OwnerNodeID, ToNodeID: actionID, Payload: closuregraph.DeclaresPayload{Origin: closuregraph.EvidenceOrigin{Field: "runtime_actions." + spec.Name}}})
		bindings = append(bindings, runtimeActionBinding{ActionID: actionID, OwnerID: spec.OwnerNodeID, ToolRole: spec.ToolRole})
	}
	graph, err := closuregraph.NewCaptureGraph(ProfileID, capture.Graph.PolicyIDs, capture.Graph.RootNodeIDs, nodes, edges, capture.Graph.ArtifactManifestIDs)
	if err != nil {
		return Capture{}, err
	}
	capture.Graph, capture.Nodes, capture.Edges, capture.runtimeActions = graph, nodes, edges, bindings
	return capture, nil
}

// Close projects the exact binding and derives the immutable C5 plan without
// adding graph records after C4.
func Close(capture Capture, selection closuregraph.SelectionContext, exact RuntimeBinding, evaluators []closuregraph.ConditionEvaluator, executionPolicyID string) (closuregraph.GraphBundle, closuregraph.BuildPlan, error) {
	binding, bindingNodes, bindingEdges, authority, err := Bind(capture, selection, exact)
	if err != nil {
		return closuregraph.GraphBundle{}, closuregraph.BuildPlan{}, err
	}
	records := closuregraph.NewRecordTables(capture.Nodes, capture.Edges, bindingNodes, bindingEdges)
	bundle, err := closuregraph.ProjectActive(capture.Graph, selection, binding, records, authority, evaluators)
	if err != nil {
		return closuregraph.GraphBundle{}, closuregraph.BuildPlan{}, err
	}
	plan, err := closuregraph.DeriveBuildPlan(bundle, closuregraph.PlanOptions{ExecutionPolicyID: executionPolicyID})
	return bundle, plan, err
}

// MetadataDerivationRequest is the only common-profile seam for executable
// C1-C4 metadata. The caller must supply an already admitted input set and an
// exact permit that names the C0-bound Node or manager tool.
type MetadataDerivationRequest struct {
	Executor    *closureexec.Executor
	Permit      closureexec.DerivationPermit
	C0ToolNodes []closuregraph.Node
	Inputs      map[closuregraph.ID]closureexec.AdmittedInput
	Recheck     func(context.Context) (closureexec.ToolchainIdentity, error)
}

// ExecuteMetadataDerivation commits before launch, performs the immediate
// tool/input rechecks, and accepts only an executor-issued causal receipt.
func ExecuteMetadataDerivation(ctx context.Context, c0 closuregraph.Checkpoint, request MetadataDerivationRequest) (closureexec.DerivationReceipt, error) {
	if request.Executor == nil || request.Recheck == nil {
		return closureexec.DerivationReceipt{}, fail(closuregraph.CodeDerivationUnauthorized, "metadata executor or recheck is absent")
	}
	if c0.Name != closuregraph.CheckpointC0 {
		return closureexec.DerivationReceipt{}, fail(closuregraph.CodeCheckpointInvalid, "metadata authority is not C0.profile")
	}
	c0ID, err := c0.ID()
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	toolIDs := map[closuregraph.ID]bool{}
	for _, id := range c0.Payload.(closuregraph.C0ProfilePayload).EvidenceToolchainNodeIDs {
		toolIDs[id] = true
	}
	if request.Permit.C0CheckpointID != c0ID || !toolIDs[request.Permit.ToolchainNodeID] || request.Permit.InvocationSubtype != closureexec.DerivationMetadata {
		return closureexec.DerivationReceipt{}, fail(closuregraph.CodeDerivationUnauthorized, "metadata permit is not bound to the exact C0 tool table")
	}
	var bound *closuregraph.ToolchainComponentPayload
	for _, node := range request.C0ToolNodes {
		if node.Kind != closuregraph.NodeToolchainComponent {
			continue
		}
		nodeID, nodeErr := node.ID()
		if nodeErr != nil {
			return closureexec.DerivationReceipt{}, nodeErr
		}
		if !toolIDs[nodeID] {
			continue
		}
		if nodeID == request.Permit.ToolchainNodeID {
			if bound != nil {
				return closureexec.DerivationReceipt{}, fail(closuregraph.CodeDerivationUnauthorized, "metadata tool record is duplicated")
			}
			payload := node.Payload.(closuregraph.ToolchainComponentPayload)
			bound = &payload
		}
	}
	if bound == nil || bound.ContentFingerprint != request.Permit.ToolchainFingerprint ||
		len(bound.LinkFingerprintIDs) != 1 || bound.LinkFingerprintIDs[0] != request.Permit.ExecutableSHA256 ||
		bound.ExecutableRelativePath != request.Permit.Executable || bound.TimeOfUseRecheckRule != request.Permit.RecheckRule {
		return closureexec.DerivationReceipt{}, fail(closuregraph.CodeDerivationUnauthorized, "metadata permit executable evidence differs from the exact C0 tool record")
	}
	payload := c0.Payload.(closuregraph.C0ProfilePayload)
	hostID, hostOK := payload.PlatformRoles[closuregraph.PlatformHost]
	if !hostOK {
		hostID, hostOK = payload.PlatformRoles[closuregraph.PlatformTarget]
	}
	targetID, targetOK := payload.PlatformRoles[closuregraph.PlatformTarget]
	if !hostOK || !targetOK || request.Permit.HostID != hostID || request.Permit.TargetID != targetID {
		return closureexec.DerivationReceipt{}, fail(closuregraph.CodeDerivationUnauthorized, "metadata permit host or target differs from C0")
	}
	requiredRole := closuregraph.PlatformTarget
	if bound.ExecutionDomain == closuregraph.ExecutionHost {
		requiredRole = closuregraph.PlatformHost
	}
	if len(bound.PlatformRoleNames) != 1 || bound.PlatformRoleNames[0] != requiredRole {
		return closureexec.DerivationReceipt{}, fail(closuregraph.CodeDerivationUnauthorized, "metadata permit execution domain differs from C0")
	}
	permitID, err := request.Executor.Commit(request.Permit)
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	receipt, err := request.Executor.Execute(ctx, permitID, request.Recheck, request.Inputs)
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	if err = request.Executor.VerifyIssuedDerivationReceipt(receipt); err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	return receipt, nil
}

// ObservedOutput adds the Node grammar contract to the shared immutable
// output/action/edge observation.
type ObservedOutput struct {
	Grammar     string
	Observation closuregraph.ProducedArtifactObservation
}

// ValidateOutputObservations reconciles one unique observation per graph-bound
// published output. Grammar is verified through the declaration digest derived
// from the complete path/class/grammar/role contract before C4.
func ValidateOutputObservations(observed []ObservedOutput, bundle closuregraph.GraphBundle, plan closuregraph.BuildPlan) error {
	if err := bundle.Validate(); err != nil {
		return fail(CodeGeneratedOutputDrift, "active output authority is invalid: %v", err)
	}
	if err := plan.Validate(); err != nil {
		return fail(CodeGeneratedOutputDrift, "build plan output authority is invalid: %v", err)
	}
	derivedPlan, err := closuregraph.DeriveBuildPlan(bundle, closuregraph.PlanOptions{ExecutionPolicyID: plan.ExecutionPolicyID})
	if err != nil {
		return fail(CodeGeneratedOutputDrift, "canonical build plan derivation failed: %v", err)
	}
	derivedPlanID, err := derivedPlan.ID()
	if err != nil {
		return fail(CodeGeneratedOutputDrift, "canonical build plan identity is invalid: %v", err)
	}
	suppliedPlanID, err := plan.ID()
	if err != nil || suppliedPlanID != derivedPlanID {
		return fail(CodeGeneratedOutputDrift, "build plan differs from the exact active graph projection")
	}
	activeID, err := bundle.Active.ID()
	if err != nil || plan.ActiveGraphID != activeID {
		return fail(CodeGeneratedOutputDrift, "build plan does not belong to the active graph")
	}
	records := bundle.Records
	type declaration struct {
		id      closuregraph.ID
		payload closuregraph.OutputArtifactPayload
	}
	declarations := map[string]declaration{}
	nodesByID := make(map[closuregraph.ID]closuregraph.Node, len(records.CaptureNodes))
	for _, node := range records.CaptureNodes {
		nodeID, nodeErr := node.ID()
		if nodeErr != nil {
			return fail(CodeGeneratedOutputDrift, "declared output identity is invalid: %v", nodeErr)
		}
		nodesByID[nodeID] = node
	}
	for _, nodeID := range plan.DeclaredOutputNodeIDs {
		node, ok := nodesByID[nodeID]
		if !ok || node.Kind != closuregraph.NodeOutputArtifact {
			return fail(CodeGeneratedOutputDrift, "active plan output %s is missing or wrong-kind", nodeID)
		}
		payload := node.Payload.(closuregraph.OutputArtifactPayload)
		if _, duplicate := declarations[payload.LogicalPath]; duplicate {
			return fail(CodeGeneratedOutputDrift, "graph has duplicate declared output path %s", payload.LogicalPath)
		}
		declarations[payload.LogicalPath] = declaration{id: nodeID, payload: payload}
	}
	if len(declarations) != len(observed) {
		return fail(CodeGeneratedOutputDrift, "declared and observed output counts differ")
	}
	seen := map[string]bool{}
	for _, output := range observed {
		declaration, ok := declarations[output.Observation.Path]
		if !ok || seen[output.Observation.Path] {
			return fail(CodeGeneratedOutputDrift, "observed output path %s is extra or duplicate", output.Observation.Path)
		}
		seen[output.Observation.Path] = true
		contract, err := normalizeGeneratedOutput(GeneratedOutput{Path: declaration.payload.LogicalPath, Class: declaration.payload.ExpectedClass, Grammar: output.Grammar, Role: declaration.payload.OutputRole})
		if err != nil || contract.DeclarationDigest != declaration.payload.DeclarationDigest || output.Observation.ExpectedOutputNodeID != declaration.id || output.Observation.Class != declaration.payload.ExpectedClass || !output.Observation.SHA256.Valid() {
			return fail(CodeGeneratedOutputDrift, "observed output %s class/grammar/content identity differs", output.Observation.Path)
		}
		if err := output.Observation.ValidateAgainst(records); err != nil {
			return fail(CodeGeneratedOutputDrift, "observed output %s does not match shared output contract: %v", output.Observation.Path, err)
		}
	}
	return nil
}

// ValidateObservedOutputs enforces the exact declared write set before publication.
func ValidateObservedOutputs(declared []GeneratedOutput, observed map[string]closuregraph.ID) error {
	paths := make(map[string]bool, len(declared))
	for _, original := range declared {
		item, err := normalizeGeneratedOutput(original)
		if err != nil || paths[item.Path] {
			return fail(CodeGeneratedOutputDrift, "declared outputs contain invalid or duplicate path/class/grammar/identity")
		}
		paths[item.Path] = true
	}
	if len(paths) != len(observed) {
		return fail(CodeGeneratedOutputDrift, "declared and observed output counts differ")
	}
	for _, item := range declared {
		digest, ok := observed[item.Path]
		if !ok {
			return fail(CodeGeneratedOutputDrift, "declared output %s is missing", item.Path)
		}
		if !digest.Valid() {
			return fail(CodeGeneratedOutputDrift, "observed output %s has invalid content identity", item.Path)
		}
	}
	return nil
}

// MaterializationPlan is the common N13 contract: installed trees, PnP state,
// and manager caches found before replay are discarded, never admitted as
// source authority, and the replacement requires a derivation receipt.
type MaterializationPlan struct {
	DiscardPaths              []string
	LifecycleMode             string
	AmbientAuthority          bool
	RequiresDerivationReceipt bool
}

// PlanFreshMaterialization canonicalizes preseeded derived state and forces a
// scripts-disabled, receipt-bearing rebuild from admitted raw artifacts.
func PlanFreshMaterialization(preseeded []string) (MaterializationPlan, error) {
	paths := append([]string(nil), preseeded...)
	sort.Strings(paths)
	for index, value := range paths {
		if value == "" || path.IsAbs(value) || path.Clean(value) != value || value == ".." || strings.HasPrefix(value, "../") || (index > 0 && paths[index-1] == value) {
			return MaterializationPlan{}, fail(CodeGeneratedOutputDrift, "preseeded manager state path is invalid or duplicate")
		}
	}
	return MaterializationPlan{DiscardPaths: paths, LifecycleMode: "disabled", AmbientAuthority: false, RequiresDerivationReceipt: true}, nil
}

// ValidateDerivedMaterialization rejects any regenerated manager state that
// lacks the protected derivation receipt required by the plan.
func ValidateDerivedMaterialization(plan MaterializationPlan, receiptID closuregraph.ID) error {
	if plan.LifecycleMode != "disabled" || plan.AmbientAuthority || !plan.RequiresDerivationReceipt || !receiptID.Valid() {
		return fail(CodeGeneratedOutputDrift, "derived manager state lacks lifecycle suppression or receipt authority")
	}
	return nil
}
