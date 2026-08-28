package swiftpmbuild

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpminterop"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// executionPolicyID is the exact C5 execution policy this stage plans under.
const executionPolicyID = "swiftpm-native-build-v1"

// executionRootPlaceholder keeps the planned command portable. The concrete
// task-private execution root is substituted only when the permit is
// committed, so temporary paths never enter the plan identity.
const executionRootPlaceholder = "{execution-root}"

// scratchRoot is the isolated build root inside the private work copy. Every
// SwiftPM cache, config, security, scratch, and output path lives below it, so
// no ambient home or user configuration can reach the build.
const scratchRoot = ".curator"

// NewPlan binds the exact build overlay, extends the accepted interop closure
// with exactly one product link action and its expected output, and derives
// the immutable C5 plan, command, closure, and publication authority. It adds
// no process and reads no observed byte.
func NewPlan(ctx context.Context, config Config, capture *swiftpmsource.Capture, interop *swiftpminterop.Result) (*Plan, error) {
	if capture == nil || interop == nil {
		return nil, fail(CodeDerivationUnauthorized, "SwiftPM build requires an accepted capture and interop closure")
	}
	if config.Recheck == nil || config.Configuration == "" || config.Linker.Role == "" {
		return nil, fail(CodeDerivationUnauthorized, "SwiftPM build authority is incomplete")
	}
	if err := assuranceReads(interop, config.Assurance); err != nil {
		return nil, err
	}
	if err := validateAcceptedChain(capture, interop); err != nil {
		return nil, err
	}
	index, err := indexBinding(interop)
	if err != nil {
		return nil, err
	}
	platformID, ok := interop.Selection.PlatformRoles[closuregraph.PlatformTarget]
	if !ok || !platformID.Valid() {
		return nil, fail(CodeGraphIncomplete, "accepted selection binds no exact target platform")
	}
	platformNode, present := index.nodesByID[platformID]
	if !present || platformNode.Kind != closuregraph.NodeTargetPlatform {
		return nil, fail(CodeGraphReferenceInvalid, "selected target platform is dangling or wrong-kind")
	}
	linkerNode, linkerID, err := linkerBindingNode(config.Linker)
	if err != nil {
		return nil, err
	}
	if _, duplicate := index.nodesByID[linkerID]; duplicate {
		return nil, failFields(CodeGraphReferenceInvalid, map[string]string{"role": config.Linker.Role}, "selected linker duplicates an accepted binding node")
	}
	if _, duplicate := index.idsByRole[config.Linker.Role]; duplicate {
		return nil, failFields(CodeGraphReferenceInvalid, map[string]string{"role": config.Linker.Role}, "selected linker role is already bound")
	}
	slots, err := resolveSlots(config, index, requiresCXXDriver(interop), linkerNode, linkerID)
	if err != nil {
		return nil, err
	}
	binding := Binding{PlatformNodeID: platformID, Platform: platformNode.Payload.(closuregraph.TargetPlatformPayload), ProductNodeID: capture.ProductNodeID, Slots: slots}
	if err = recheckSlots(ctx, config, binding); err != nil {
		return nil, err
	}

	objects, err := selectedObjectSlots(interop)
	if err != nil {
		return nil, err
	}
	outputPath := productOutputPath(binding.Platform.TargetTriple, config.Configuration, capture.SelectionProduct())
	linkNodes, linkEdges, linkBindingEdges, linkAction, output, err := linkRecords(capture, binding, objects, outputPath)
	if err != nil {
		return nil, err
	}

	nodes := append(append([]closuregraph.Node(nil), interop.Records.CaptureNodes...), linkNodes...)
	edges := append(append([]closuregraph.Edge(nil), interop.Records.CaptureEdges...), linkEdges...)
	bindingNodes := append(append([]closuregraph.Node(nil), interop.Records.BindingNodes...), linkerNode)
	bindingEdges := append(append([]closuregraph.Edge(nil), interop.Records.BindingEdges...), linkBindingEdges...)

	graph, err := closuregraph.NewCaptureGraph(ProfileID, interop.Graph.PolicyIDs, interop.Graph.RootNodeIDs, nodes, edges, interop.Graph.ArtifactManifestIDs)
	if err != nil {
		return nil, err
	}
	graphID, err := graph.ID()
	if err != nil {
		return nil, err
	}
	selectionID, err := interop.Selection.ID()
	if err != nil {
		return nil, err
	}
	selectionBinding, err := closuregraph.NewSelectionBinding(graphID, selectionID, bindingNodes, bindingEdges)
	if err != nil {
		return nil, err
	}
	authority, err := buildAuthority(interop, linkerNode, linkerID)
	if err != nil {
		return nil, err
	}
	records := closuregraph.NewRecordTables(nodes, edges, bindingNodes, bindingEdges)
	bundle, err := closuregraph.ProjectActive(graph, interop.Selection, selectionBinding, records, authority, []closuregraph.ConditionEvaluator{destinationEvaluator(capture)})
	if err != nil {
		return nil, err
	}
	if err = preservesAcceptedIdentities(interop, bundle); err != nil {
		return nil, err
	}
	if err = validateActionBindings(bundle, binding); err != nil {
		return nil, err
	}

	activeID, err := bundle.Active.ID()
	if err != nil {
		return nil, err
	}
	bindingID, err := selectionBinding.ID()
	if err != nil {
		return nil, err
	}
	c4, err := closuregraph.NewCheckpoint(closuregraph.C4ClosePayload{ActiveGraphID: activeID, CapturedGraphID: graphID, SelectionBindingID: bindingID, SelectionContextID: selectionID}, &interop.C4, nil)
	if err != nil {
		return nil, err
	}
	c4ID, err := c4.ID()
	if err != nil {
		return nil, err
	}
	buildPlan, err := closuregraph.DeriveBuildPlan(bundle, closuregraph.PlanOptions{ExecutionPolicyID: executionPolicyID, LastCheckpointID: c4ID})
	if err != nil {
		return nil, err
	}
	planID, err := buildPlan.ID()
	if err != nil {
		return nil, err
	}
	if err = validateDeclaredOutputs(buildPlan, objects, output); err != nil {
		return nil, err
	}
	c5, err := closuregraph.NewCheckpoint(closuregraph.C5PlanPayload{BuildPlanID: planID}, &c4, nil)
	if err != nil {
		return nil, err
	}
	closure, err := closuregraph.NewSourceClosure(c5)
	if err != nil {
		return nil, err
	}
	closureID, err := closure.ID()
	if err != nil {
		return nil, err
	}
	expected := closuregraph.ExpectedCacheInput{SchemaID: closuregraph.SchemaExpectedCacheInput, ClosureID: closureID, ExpectedOutputNodeIDs: append([]closuregraph.ID(nil), buildPlan.DeclaredOutputNodeIDs...)}
	if err = expected.Validate(); err != nil {
		return nil, err
	}
	command := buildCommand(config, capture, binding)
	commandID, err := commandIdentity(command, binding)
	if err != nil {
		return nil, err
	}
	return &Plan{
		Binding: binding, Graph: bundle, C4: c4, C5: c5, BuildPlan: buildPlan, Closure: closure, Expected: expected,
		Publication: closuregraph.PublicationEvidence{C4: c4, C5: c5, Graph: bundle, Plan: buildPlan, Closure: closure},
		Command:     command, CommandID: commandID, LinkActionNodeID: linkAction, OutputNodeID: output,
		OutputPath: outputPath, Objects: objects, ScratchDirectory: swiftpmScratchDirectory(binding.Platform.TargetTriple, config.Configuration),
		capture: capture, interop: interop,
	}, nil
}

// validateDeclaredOutputs proves the C5 plan declares exactly the selected
// product plus one object per selected compile source, and nothing else.
func validateDeclaredOutputs(plan closuregraph.BuildPlan, objects []ObjectSlot, product closuregraph.ID) error {
	declared := map[closuregraph.ID]bool{}
	for _, id := range plan.DeclaredOutputNodeIDs {
		declared[id] = true
	}
	if len(declared) != len(objects)+1 || !declared[product] {
		return fail(CodeBuildGraphDrift, "C5 plan declares %d outputs, want the product plus %d objects", len(declared), len(objects))
	}
	for _, object := range objects {
		if !declared[object.NodeID] {
			return failFields(CodeBuildGraphDrift, map[string]string{"path": object.Path}, "C5 plan omits a declared object output")
		}
	}
	return nil
}

// validateAcceptedChain proves the interop closure this stage consumes is the
// exact republication of the accepted source capture, not a foreign record.
func validateAcceptedChain(capture *swiftpmsource.Capture, interop *swiftpminterop.Result) error {
	if err := interop.C4.Validate(); err != nil || interop.C4.Name != closuregraph.CheckpointC4 {
		return fail(CodeCheckpointInvalid, "accepted interop C4 checkpoint is absent or invalid")
	}
	captureC4ID, err := capture.C4.ID()
	if err != nil {
		return err
	}
	if interop.C4.PreviousCheckpointID == nil || *interop.C4.PreviousCheckpointID != captureC4ID {
		return fail(CodeCheckpointInvalid, "accepted interop closure does not chain from the supplied source capture")
	}
	captureSelectionID, err := capture.Selection.ID()
	if err != nil {
		return err
	}
	interopSelectionID, err := interop.Selection.ID()
	if err != nil {
		return err
	}
	if captureSelectionID != interopSelectionID {
		return fail(CodeCheckpointInvalid, "accepted interop closure binds another selection context")
	}
	if len(capture.Lock.Pins) != len(capture.Mirrors) {
		return fail(CodeMirrorMissing, "root lock pins and captured mirrors are not bijective")
	}
	if !capture.Lock.Digest.Valid() {
		return fail(CodeResolutionUnfrozen, "build requires a frozen root lock")
	}
	return nil
}

// preservesAcceptedIdentities proves this stage only added records: every
// accepted capture and binding identity survives unchanged through C5.
func preservesAcceptedIdentities(interop *swiftpminterop.Result, bundle closuregraph.GraphBundle) error {
	present := map[closuregraph.ID]bool{}
	for _, node := range append(append([]closuregraph.Node(nil), bundle.Records.CaptureNodes...), bundle.Records.BindingNodes...) {
		id, err := node.ID()
		if err != nil {
			return err
		}
		present[id] = true
	}
	for _, edge := range append(append([]closuregraph.Edge(nil), bundle.Records.CaptureEdges...), bundle.Records.BindingEdges...) {
		id, err := edge.ID()
		if err != nil {
			return err
		}
		present[id] = true
	}
	for _, node := range append(append([]closuregraph.Node(nil), interop.Records.CaptureNodes...), interop.Records.BindingNodes...) {
		id, err := node.ID()
		if err != nil {
			return err
		}
		if !present[id] {
			return failFields(CodeBuildGraphDrift, map[string]string{"logical_key": node.LogicalKey}, "build planning replaced an accepted graph node")
		}
	}
	for _, edge := range append(append([]closuregraph.Edge(nil), interop.Records.CaptureEdges...), interop.Records.BindingEdges...) {
		id, err := edge.ID()
		if err != nil {
			return err
		}
		if !present[id] {
			return failFields(CodeBuildGraphDrift, map[string]string{"edge_key": edge.EdgeKey}, "build planning replaced an accepted graph edge")
		}
	}
	return nil
}

// selectedObjectSlots resolves the exact intermediate object slot each
// selected compile action declares. The link action reads exactly this set and
// the offline build must materialize every one of them.
func selectedObjectSlots(interop *swiftpminterop.Result) ([]ObjectSlot, error) {
	nodes := map[closuregraph.ID]closuregraph.Node{}
	for _, node := range interop.Records.CaptureNodes {
		id, err := node.ID()
		if err != nil {
			return nil, err
		}
		nodes[id] = node
	}
	produces := map[closuregraph.ID]closuregraph.ID{}
	for _, edge := range interop.Records.CaptureEdges {
		if edge.Kind != closuregraph.EdgeProduces {
			continue
		}
		if _, duplicate := produces[edge.ToNodeID]; duplicate {
			return nil, fail(CodeGraphReferenceInvalid, "one declared object output has two producers")
		}
		id, err := edge.ID()
		if err != nil {
			return nil, err
		}
		produces[edge.ToNodeID] = id
	}
	objects := []ObjectSlot{}
	seen := map[closuregraph.ID]bool{}
	for _, target := range interop.Targets {
		if !target.Selected || target.Kind == swiftpminterop.KindSystem || !target.ActionNodeID.Valid() {
			continue
		}
		if len(target.ObjectNodeIDs) != len(target.Sources) || len(target.Sources) == 0 {
			return nil, failFields(CodeGraphReferenceInvalid, map[string]string{"target": target.Package + ":" + target.Target}, "selected compile action declares %d object outputs for %d sources", len(target.ObjectNodeIDs), len(target.Sources))
		}
		for index, nodeID := range target.ObjectNodeIDs {
			node, resolved := nodes[nodeID]
			producesID, produced := produces[nodeID]
			if !resolved || node.Kind != closuregraph.NodeOutputArtifact || !produced || seen[nodeID] {
				return nil, failFields(CodeGraphReferenceInvalid, map[string]string{"target": target.Package + ":" + target.Target}, "declared object output is dangling, wrong-kind, unproduced, or duplicated")
			}
			seen[nodeID] = true
			objects = append(objects, ObjectSlot{
				Package: target.Package, Target: target.Target, Source: target.Sources[index], SourceRoot: target.SourceRoot, Kind: string(target.Kind),
				NodeID: nodeID, ActionNodeID: target.ActionNodeID, ProducesEdgeID: producesID,
				Path: node.Payload.(closuregraph.OutputArtifactPayload).LogicalPath,
			})
		}
	}
	if len(objects) == 0 {
		return nil, fail(CodeGraphIncomplete, "selected product compiles no target")
	}
	sort.Slice(objects, func(i, j int) bool { return objects[i].NodeID < objects[j].NodeID })
	return objects, nil
}

// linkRecords declares the single product link action, its expected immutable
// output, and the exact tool and platform binding overlay it consumes.
func linkRecords(capture *swiftpmsource.Capture, binding Binding, objects []ObjectSlot, outputPath string) ([]closuregraph.Node, []closuregraph.Edge, []closuregraph.Edge, closuregraph.ID, closuregraph.ID, error) {
	product := capture.SelectionProduct()
	objectIDs := make([]closuregraph.ID, len(objects))
	for index, object := range objects {
		objectIDs[index] = object.NodeID
	}
	declaration, err := closuregraph.DomainID("swiftpm-build-link-action-v1", map[string]any{
		"objects": idValues(objectIDs), "output": outputPath, "product": product, "target_triple": binding.Platform.TargetTriple,
	})
	if err != nil {
		return nil, nil, nil, "", "", err
	}
	readSlots := make([]string, len(objects))
	for index := range objects {
		readSlots[index] = objectReadSlot(index)
	}
	argv := []string{"$TOOL(build-driver)", "$TOOL(linker)"}
	for _, slot := range readSlots {
		argv = append(argv, "$READ("+slot+")")
	}
	argv = append(argv, "$WRITE(product)")
	action := closuregraph.Node{Kind: closuregraph.NodeAction, LogicalKey: "swiftpm.build.link." + product, Payload: closuregraph.ActionPayload{
		Profile: ProfileID, ActionSubtype: "swiftpm-link", ExecutionDomain: closuregraph.ExecutionTarget,
		ArgvTemplate:  argv,
		ToolSlotNames: []string{"build-driver", "linker"}, ReadSlotNames: readSlots, WriteSlotNames: []string{"product"},
		EnvironmentPolicyID: "swiftpm-build-environment-v1", ProcessPolicyID: "swiftpm-build-process-v1", Network: "none",
		PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformTarget},
	}}
	if err = action.Validate(); err != nil {
		return nil, nil, nil, "", "", err
	}
	actionID, err := action.ID()
	if err != nil {
		return nil, nil, nil, "", "", err
	}
	output := closuregraph.Node{Kind: closuregraph.NodeOutputArtifact, LogicalKey: "swiftpm.build.product." + product, Payload: closuregraph.OutputArtifactPayload{
		Profile: ProfileID, LogicalPath: outputPath, ExpectedClass: "native.executable", OutputRole: "command",
		CompatibilityPredicate: binding.Platform.TargetTriple, DeclarationDigest: declaration,
		PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformTarget},
	}}
	if err = output.Validate(); err != nil {
		return nil, nil, nil, "", "", err
	}
	outputID, err := output.ID()
	if err != nil {
		return nil, nil, nil, "", "", err
	}
	edges := []closuregraph.Edge{
		{Kind: closuregraph.EdgeDeclares, EdgeKey: "swiftpm.build.declares-link." + product, FromNodeID: binding.ProductNodeID, ToNodeID: actionID, Payload: closuregraph.DeclaresPayload{Origin: closuregraph.EvidenceOrigin{Field: "products." + product + ".link"}}},
		{Kind: closuregraph.EdgeProduces, EdgeKey: "swiftpm.build.produces-product." + product, FromNodeID: actionID, ToNodeID: outputID, Payload: closuregraph.ProducesPayload{Path: outputPath, WriteSlot: "product", WriteClass: "native.executable"}},
		{Kind: closuregraph.EdgePublishes, EdgeKey: "swiftpm.build.publishes." + product, FromNodeID: binding.ProductNodeID, ToNodeID: outputID, Payload: closuregraph.PublishesPayload{Destination: outputPath, EntryPoint: product}},
	}
	for index, object := range objects {
		edges = append(edges, closuregraph.Edge{
			Kind: closuregraph.EdgeReads, EdgeKey: linkReadKey(product, index), FromNodeID: actionID, ToNodeID: object.NodeID,
			Payload: closuregraph.ReadsPayload{Path: object.Path, ReadSlot: objectReadSlot(index), ReadClass: "native.object"},
		})
	}
	bindingEdges := []closuregraph.Edge{
		{Kind: closuregraph.EdgeUsesTool, EdgeKey: "swiftpm.build.uses-driver." + product, FromNodeID: actionID, ToNodeID: binding.Slots[SlotSwiftPM].NodeID, Payload: closuregraph.UsesToolPayload{ExecutableRelativePath: binding.Slots[SlotSwiftPM].Payload.ExecutableRelativePath, ToolSlot: "build-driver", InvocationRole: "swiftpm-build"}},
		{Kind: closuregraph.EdgeUsesTool, EdgeKey: "swiftpm.build.uses-linker." + product, FromNodeID: actionID, ToNodeID: binding.Slots[SlotLinker].NodeID, Payload: closuregraph.UsesToolPayload{ExecutableRelativePath: binding.Slots[SlotLinker].Payload.ExecutableRelativePath, ToolSlot: "linker", InvocationRole: "swiftpm-link"}},
		{Kind: closuregraph.EdgeTargets, EdgeKey: "swiftpm.build.link-target." + product, FromNodeID: actionID, ToNodeID: binding.PlatformNodeID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.target"}}},
		{Kind: closuregraph.EdgeTargets, EdgeKey: "swiftpm.build.product-target." + product, FromNodeID: outputID, ToNodeID: binding.PlatformNodeID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.target"}}},
		{Kind: closuregraph.EdgeTargets, EdgeKey: "swiftpm.build.linker-target." + product, FromNodeID: binding.Slots[SlotLinker].NodeID, ToNodeID: binding.PlatformNodeID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.target"}}},
		{Kind: closuregraph.EdgeRequires, EdgeKey: "swiftpm.build.linker-requires." + product, FromNodeID: binding.ProductNodeID, ToNodeID: binding.Slots[SlotLinker].NodeID, Payload: closuregraph.RequiresPayload{Scope: closuregraph.ScopeToolchain, Origin: closuregraph.EvidenceOrigin{Field: "selection.build_components." + binding.Slots[SlotLinker].Role}, DependencyKind: "external-toolchain"}},
	}
	return []closuregraph.Node{action, output}, edges, bindingEdges, actionID, outputID, nil
}

// linkerBindingNode publishes the exact selected linker as a binding node. The
// linker is the only physical component this stage selects itself.
func linkerBindingNode(component swiftpminterop.ExternalComponent) (closuregraph.Node, closuregraph.ID, error) {
	if component.Fingerprint == "" || !component.Fingerprint.Valid() || component.ExecutableRelativePath == "" || component.PlatformABI == "" || component.PolicySelector == "" || component.VersionOutput == "" {
		return closuregraph.Node{}, "", failFields(CodeToolchainUntrusted, map[string]string{"role": component.Role}, "selected linker identity is incomplete")
	}
	payload := closuregraph.ToolchainComponentPayload{
		ComponentRole: component.Role, ContentFingerprint: component.Fingerprint, ExecutableRelativePath: component.ExecutableRelativePath,
		PlatformABI: component.PlatformABI, PolicySelector: component.PolicySelector, VersionOutput: component.VersionOutput,
		SDKFactsDigest: component.SDKFactsDigest, TimeOfUseRecheckRule: "immediate-exact-v1",
		ExecutionDomain: closuregraph.ExecutionTarget, PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformTarget},
	}
	if component.ExecutableSHA256.Valid() {
		payload.LinkFingerprintIDs = []closuregraph.ID{component.ExecutableSHA256}
	}
	node := closuregraph.Node{Kind: closuregraph.NodeToolchainComponent, LogicalKey: "swiftpm.build.component." + component.Role, Payload: payload}
	if err := node.Validate(); err != nil {
		return closuregraph.Node{}, "", failFields(CodeToolchainUntrusted, map[string]string{"role": component.Role}, "selected linker is not a valid binding node: %v", err)
	}
	id, err := node.ID()
	return node, id, err
}

// buildAuthority extends the accepted C4 binding authority with the exact
// linker selector. Every earlier binding evidence record is preserved.
func buildAuthority(interop *swiftpminterop.Result, linker closuregraph.Node, linkerID closuregraph.ID) (closuregraph.BindingAuthority, error) {
	authority := closuregraph.BindingAuthority{
		C0Checkpoint: interop.Authority.C0Checkpoint,
		Toolchains:   append([]closuregraph.ToolchainBindingEvidence(nil), interop.Authority.Toolchains...),
		C4Selectors:  append([]closuregraph.ToolchainSelector(nil), interop.Authority.C4Selectors...),
	}
	selector, err := closuregraph.NewToolchainSelector(linker)
	if err != nil {
		return closuregraph.BindingAuthority{}, err
	}
	selectorID, err := selector.ID()
	if err != nil {
		return closuregraph.BindingAuthority{}, err
	}
	authority.C4Selectors = append(authority.C4Selectors, selector)
	authority.Toolchains = append(authority.Toolchains, closuregraph.ToolchainBindingEvidence{NodeID: linkerID, FirstBound: closuregraph.ToolchainBoundAtC4, EvidenceID: selectorID})
	sort.Slice(authority.Toolchains, func(i, j int) bool { return authority.Toolchains[i].NodeID < authority.Toolchains[j].NodeID })
	sort.Slice(authority.C4Selectors, func(i, j int) bool { return authority.C4Selectors[i].NodeID < authority.C4Selectors[j].NodeID })
	return authority, nil
}

// destinationEvaluator rebuilds the exact SwiftPM condition evaluator from the
// accepted destination markers. The build stage cannot invent a verdict.
func destinationEvaluator(capture *swiftpmsource.Capture) closuregraph.ConditionEvaluator {
	markers := capture.Destination().Markers
	return closuregraph.ConditionEvaluatorFunc{EvaluatorID: swiftpmsource.ConditionEvaluatorID, EvaluateFunc: func(condition closuregraph.Condition, _ closuregraph.EvaluationInput) (bool, error) {
		if condition.EvaluatorID != swiftpmsource.ConditionEvaluatorID {
			return false, fail(CodeGraphReferenceInvalid, "wrong SwiftPM condition evaluator")
		}
		parts := strings.Split(condition.Expression, "=")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return false, fail(CodeGraphReferenceInvalid, "unsupported SwiftPM condition %q", condition.Expression)
		}
		switch parts[0] {
		case "platform", "configuration", "architecture":
			return strings.EqualFold(markers[parts[0]], parts[1]), nil
		case "trait":
			return markers["trait:"+parts[1]] == "true", nil
		default:
			return false, fail(CodeGraphReferenceInvalid, "unsupported SwiftPM condition key %q", parts[0])
		}
	}}
}

// buildCommand is the exact native SwiftPM invocation. Experimental prebuilts
// are disabled, resolution is forced to the frozen lock, network is denied,
// and every cache, config, security, scratch, and output root is isolated.
func buildCommand(config Config, capture *swiftpmsource.Capture, binding Binding) Command {
	work := "work/package"
	argv := plannedBuildArgv(config.Configuration, binding.Platform.TargetTriple, capture.SelectionProduct())
	isolated := executionRootPlaceholder + "/" + work + "/" + scratchRoot
	environment := map[string]string{
		"HOME":                 isolated + "/home",
		"SWIFTPM_CACHE_DIR":    isolated + "/cache",
		"SWIFTPM_CONFIG_DIR":   isolated + "/config",
		"SWIFTPM_SCRATCH_DIR":  isolated + "/scratch",
		"SWIFTPM_SECURITY_DIR": isolated + "/security",
		"TZ":                   "UTC",
	}
	return Command{Executable: binding.Slots[SlotSwiftPM].Payload.ExecutableRelativePath, CWD: work, Argv: argv, Environment: environment}
}

// plannedBuildArgv is the exact native SwiftPM build invocation: experimental
// prebuilts disabled, resolution forced to the frozen lock, every cache,
// config, security, and scratch root isolated inside the private work copy,
// and exactly one selected product on exactly one destination.
func plannedBuildArgv(configuration, triple, product string) []string {
	return []string{
		"build", "--package-path", ".",
		"--cache-path", path.Join(scratchRoot, "cache"),
		"--config-path", path.Join(scratchRoot, "config"),
		"--security-path", path.Join(scratchRoot, "security"),
		"--scratch-path", path.Join(scratchRoot, "scratch"),
		"--disable-netrc", "--disable-experimental-prebuilts", "--force-resolved-versions",
		"--build-system", "native",
		"--configuration", configuration,
		"--triple", triple,
		"--product", product,
	}
}

// commandIdentity is the portable identity of the planned command. Temporary
// paths, timestamps, and discovery order are excluded by construction.
func commandIdentity(command Command, binding Binding) (closuregraph.ID, error) {
	slots := make([]ToolSlot, 0, len(binding.Slots))
	for slot := range binding.Slots {
		slots = append(slots, slot)
	}
	sort.Slice(slots, func(i, j int) bool { return slots[i] < slots[j] })
	tools := make([]any, len(slots))
	for index, slot := range slots {
		bound := binding.Slots[slot]
		tools[index] = map[string]any{"fingerprint": string(bound.Payload.ContentFingerprint), "relative_path": bound.Payload.ExecutableRelativePath, "role": bound.Role, "slot": string(slot), "version": bound.Payload.VersionOutput}
	}
	return closuregraph.DomainID(PlanSchemaID, map[string]any{
		"argv": stringValues(command.Argv), "cwd": command.CWD, "environment": environmentValues(command.Environment),
		"executable": command.Executable, "network": "none", "target": binding.Platform.TargetTriple, "tools": tools,
	})
}

// productOutputPath is the exact logical path SwiftPM writes the product to
// inside the isolated scratch root.
func productOutputPath(triple, configuration, product string) string {
	return path.Join(swiftpmScratchDirectory(triple, configuration), product)
}

// swiftpmScratchDirectory reproduces SwiftPM's build subdirectory. SwiftPM
// names it after the target triple with the platform version removed, so the
// exact selected triple and the observed output path stay reconciled.
func swiftpmScratchDirectory(triple, configuration string) string {
	return path.Join(scratchRoot, "scratch", unversionedTriple(triple), configuration)
}

// unversionedTriple strips the trailing platform version from the operating
// system component, matching SwiftPM's own build-directory naming.
func unversionedTriple(triple string) string {
	parts := strings.Split(triple, "-")
	if len(parts) < 3 {
		return triple
	}
	parts[2] = strings.TrimRight(parts[2], "0123456789.")
	return strings.Join(parts, "-")
}

// objectReadSlot names one exact per-target object read slot. Every declared
// read slot must resolve exactly once, so the link action cannot collapse the
// selected targets into one anonymous slot. The name is the canonical position
// in the object set ordered by node identity, which keeps the closed slot
// grammar and stays independent of package and target spelling.
func objectReadSlot(index int) string {
	return fmt.Sprintf("objects.%04d", index)
}

func linkReadKey(product string, index int) string {
	return fmt.Sprintf("swiftpm.build.reads-objects.%s.%04d", product, index)
}

func idValues(values []closuregraph.ID) []any {
	result := make([]any, len(values))
	for index, value := range values {
		result[index] = string(value)
	}
	return result
}

func stringValues(values []string) []any {
	result := make([]any, len(values))
	for index, value := range values {
		result[index] = value
	}
	return result
}

func environmentValues(values map[string]string) map[string]any {
	result := make(map[string]any, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}
