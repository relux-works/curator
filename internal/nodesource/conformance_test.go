package nodesource

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/artifactpolicy/conformance"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
)

func nodeTool(role, path, seed string, domain closuregraph.ExecutionDomain) ToolIdentity {
	return ToolIdentity{Role: role, PolicySelector: role + "-v1", ExecutableRelativePath: path, VersionOutput: role + " 1.0", PlatformABI: "node", Fingerprint: id(seed), ExecutableSHA256: id(seed + "-executable"), ExecutionDomain: domain}
}

func generatedCapture(t *testing.T, manager ManagerProfile) (Capture, GeneratedAction) {
	t.Helper()
	packages := packageSet()
	packages = append(packages, PackageInstance{Key: "config@workspace", Name: "config", Version: "1.0.0", Origin: "workspace:config", Checksum: "sha256-config", WorkspacePath: "config", ArtifactManifestID: id("am-config"), SnapshotDigest: id("tree-config")})
	capture, err := BuildCapture(CaptureInput{Manager: manager, RootKeys: []string{"app@workspace"}, Packages: packages, PolicyIDs: []string{"artifact-policy-v1"}})
	if err != nil {
		t.Fatal(err)
	}
	spec := GeneratedAction{
		Name: "compile", Argv: []string{"--project", "$READ(config-0000)"}, WorkingDirectory: "workspace",
		Compiler: nodeTool("typescript-compiler", "bin/tsc", "tsc", closuregraph.ExecutionHost),
		Inputs: []GeneratedInput{
			{NodeID: capture.SourceNodeIDs["config@workspace"], Path: "workspace/tsconfig.json", Class: "source.tree", Role: "config"},
			{NodeID: capture.SourceNodeIDs["peer@1[p=react@18]"], Path: "workspace/plugin.js", Class: "source.tree", Role: "plugin"},
			{NodeID: capture.SourceNodeIDs["app@workspace"], Path: "workspace/src", Class: "source.tree", Role: "source"},
		},
		TargetNodeID: capture.ProductNodeIDs["app@workspace"], EnvironmentPolicyID: "node-env-v1", ProcessPolicyID: "node-process-v1",
		Outputs: []GeneratedOutput{{Path: "dist/cli.js", Class: "source.generated_text", Grammar: "javascript-v1", Role: "published_command"}},
	}
	capture, err = AddGeneratedActions(capture, []GeneratedAction{spec})
	if err != nil {
		t.Fatal(err)
	}
	return capture, spec
}

func closeGenerated(t *testing.T, capture Capture, osName string) (closuregraph.GraphBundle, closuregraph.BuildPlan, RuntimeBinding) {
	t.Helper()
	root := capture.ProductNodeIDs["app@workspace"]
	platform := closuregraph.TargetPlatformPayload{OS: osName, Architecture: "arm64", ABI: "node", Libc: "none", MinimumRuntime: "22", SDKID: "none", TargetTriple: "arm64-" + osName, Runtime: "node", LanguageModes: map[string]string{"module": "esm"}, Tuning: map[string]string{}}
	platformNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: platform}
	platformID, _ := platformNode.ID()
	selection, err := closuregraph.NewSelectionContext([]closuregraph.ID{root}, map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID}, []string{}, false, map[string]string{"os": osName}, map[string]string{"react": "18"}, []string{})
	if err != nil {
		t.Fatal(err)
	}
	exact := RuntimeBinding{Platform: platform, Node: nodeTool("node-runtime", "bin/node", "node-"+osName, closuregraph.ExecutionTarget), Manager: nodeTool("package-manager", "bin/manager", "manager-"+osName, closuregraph.ExecutionTarget), TargetNodeIDs: []closuregraph.ID{root}}
	c0, err := NewC0Checkpoint(capture, selection, exact)
	if err != nil {
		t.Fatal(err)
	}
	exact.C0Checkpoint = &c0
	bundle, plan, err := Close(capture, selection, exact, nil, "node-build-execution-v1")
	if err != nil {
		t.Fatal(err)
	}
	return bundle, plan, exact
}

func TestCGP10CGN15DeclaredGeneratorIsClosedBeforeC4AndPlanIsDeterministic(t *testing.T) {
	capture, spec := generatedCapture(t, ManagerNPM)
	bundle, plan, _ := closeGenerated(t, capture, "linux")
	if len(plan.ActionNodeIDs) != 1 || len(plan.DeclaredOutputNodeIDs) != 1 || len(plan.Waves) != 1 {
		t.Fatalf("plan does not contain exact declared action/output: %#v", plan)
	}
	if bundle.Authority.C0Checkpoint == nil || len(bundle.Authority.C4Selectors) != 1 || len(bundle.Authority.Toolchains) != 3 {
		t.Fatalf("Node/manager C0 and compiler C4 authority = %#v", bundle.Authority)
	}
	for _, authority := range bundle.Authority.Toolchains[:2] {
		if authority.FirstBound != closuregraph.ToolchainBoundAtC0 {
			t.Fatalf("runtime tool was not first-bound at C0: %#v", authority)
		}
	}
	for _, id := range append(append([]closuregraph.ID{}, plan.ActionNodeIDs...), plan.DeclaredOutputNodeIDs...) {
		found := false
		for _, captured := range capture.Graph.NodeIDs {
			found = found || captured == id
		}
		if !found {
			t.Fatalf("C5 grew graph with %s", id)
		}
	}
	var actionID, outputID, producesID closuregraph.ID
	for _, node := range bundle.Records.CaptureNodes {
		nodeID, _ := node.ID()
		if node.Kind == closuregraph.NodeAction {
			actionID = nodeID
		}
		if node.Kind == closuregraph.NodeOutputArtifact {
			outputID = nodeID
		}
	}
	for _, edge := range bundle.Records.CaptureEdges {
		if edge.Kind == closuregraph.EdgeProduces && edge.FromNodeID == actionID && edge.ToNodeID == outputID {
			producesID, _ = edge.ID()
		}
	}
	observation := closuregraph.ProducedArtifactObservation{Class: spec.Outputs[0].Class, ExpectedOutputNodeID: outputID, Path: spec.Outputs[0].Path, ProducerActionID: actionID, ProducesEdgeID: producesID, SHA256: id("observed-js"), Size: 3}
	if err := ValidateOutputObservations([]ObservedOutput{{Grammar: "javascript-v1", Observation: observation}}, bundle, plan); err != nil {
		t.Fatal(err)
	}
	captureID, _ := bundle.Capture.ID()
	selectionID, _ := bundle.Selection.ID()
	bindingID, _ := bundle.Binding.ID()
	activeID, _ := bundle.Active.ID()
	previous := id("c3-node-admission")
	c4 := closuregraph.Checkpoint{SchemaID: closuregraph.SchemaCheckpoint, Name: closuregraph.CheckpointC4, PreviousCheckpointID: &previous, Payload: closuregraph.C4ClosePayload{ActiveGraphID: activeID, CapturedGraphID: captureID, SelectionBindingID: bindingID, SelectionContextID: selectionID}, Decision: closuregraph.DecisionAdmit, Diagnostics: []closuregraph.Diagnostic{}}
	planID, _ := plan.ID()
	c5, err := closuregraph.NewCheckpoint(closuregraph.C5PlanPayload{BuildPlanID: planID}, &c4, nil)
	if err != nil {
		t.Fatal(err)
	}
	closure, err := closuregraph.NewSourceClosure(c5)
	if err != nil {
		t.Fatal(err)
	}
	closureID, _ := closure.ID()
	expected := closuregraph.ExpectedCacheInput{SchemaID: closuregraph.SchemaExpectedCacheInput, ClosureID: closureID, ExpectedOutputNodeIDs: plan.DeclaredOutputNodeIDs}
	observationID, _ := observation.ID()
	execution := closuregraph.ExecutionReceipt{SchemaID: closuregraph.SchemaExecutionReceipt, ActionOrder: plan.ActionNodeIDs, ClosureID: closureID, Decision: "success", Network: "none", ProducedObservationIDs: []closuregraph.ID{observationID}, ToolchainRechecks: "match", WriteSet: []string{spec.Outputs[0].Path}}
	publicationEvidence := closuregraph.PublicationEvidence{C4: c4, C5: c5, Graph: bundle, Plan: plan, Closure: closure}
	if err = publicationEvidence.ValidateForPublication(expected, execution, []closuregraph.ProducedArtifactObservation{observation}); err != nil {
		t.Fatalf("shared output/receipt contract rejected Node output: %v", err)
	}
	if ErrorCode(ValidateOutputObservations([]ObservedOutput{{Grammar: "typescript-v1", Observation: observation}}, bundle, plan)) != CodeGeneratedOutputDrift {
		t.Fatal("grammar differing from graph-bound declaration was accepted")
	}
	observation.SHA256 = "invalid"
	if ErrorCode(ValidateOutputObservations([]ObservedOutput{{Grammar: "javascript-v1", Observation: observation}}, bundle, plan)) != CodeGeneratedOutputDrift {
		t.Fatal("invalid observed content identity was accepted")
	}
}

func TestCGP05CGN09TargetChangesBindingActiveAndPlanButNotCapture(t *testing.T) {
	capture, _ := generatedCapture(t, ManagerNPM)
	darwinBundle, darwinPlan, _ := closeGenerated(t, capture, "darwin")
	linuxBundle, linuxPlan, _ := closeGenerated(t, capture, "linux")
	captureID, _ := capture.Graph.ID()
	if darwinBundle.Binding.CapturedGraphID != captureID || linuxBundle.Binding.CapturedGraphID != captureID {
		t.Fatal("target branch replaced capture")
	}
	darwinActive, _ := darwinBundle.Active.ID()
	linuxActive, _ := linuxBundle.Active.ID()
	darwinPlanID, _ := darwinPlan.ID()
	linuxPlanID, _ := linuxPlan.ID()
	if darwinActive == linuxActive || darwinPlanID == linuxPlanID {
		t.Fatal("target branch aliased active graph or plan")
	}
}

func TestN10ManagerPermutationAndExactConditionEvidence(t *testing.T) {
	packages := packageSet()
	packages[0].Dependencies = []Dependency{
		{PackageKey: "peer@1[p=react@18]", Scope: closuregraph.ScopePeer, DeclarationField: "peerDependencies.react"},
		{PackageKey: "peer@1[p=react@18]", Scope: closuregraph.ScopeOptional, DeclarationField: "optionalDependencies.peer", Condition: &closuregraph.Condition{EvaluatorID: "node-marker-v1", Expression: "os=linux"}},
	}
	reversed := packageSet()
	reversed[0].Dependencies = []Dependency{packages[0].Dependencies[1], packages[0].Dependencies[0]}
	var managerIndependent closuregraph.ID
	for _, profile := range []ManagerProfile{ManagerNPM, ManagerPNPM, ManagerYarnClassic, ManagerYarnModern} {
		left, err := BuildCapture(CaptureInput{Manager: profile, RootKeys: []string{"app@workspace"}, Packages: packages, PolicyIDs: []string{"p"}})
		if err != nil {
			t.Fatal(err)
		}
		right, err := BuildCapture(CaptureInput{Manager: profile, RootKeys: []string{"app@workspace"}, Packages: reversed, PolicyIDs: []string{"p"}})
		if err != nil {
			t.Fatal(err)
		}
		leftID, _ := left.Graph.ID()
		rightID, _ := right.Graph.ID()
		if leftID != rightID {
			t.Fatalf("%s dependency permutation changed capture", profile)
		}
		if managerIndependent == "" {
			managerIndependent = leftID
		} else if managerIndependent != leftID {
			t.Fatalf("%s changed the manager-independent canonical capture", profile)
		}
	}
	duplicate := packageSet()
	duplicate[0].Dependencies = []Dependency{packages[0].Dependencies[0], packages[0].Dependencies[0]}
	if _, err := BuildCapture(CaptureInput{Manager: ManagerNPM, RootKeys: []string{"app@workspace"}, Packages: duplicate, PolicyIDs: []string{"p"}}); ErrorCode(err) != string(closuregraph.CodeGraphReferenceInvalid) {
		t.Fatalf("duplicate declaration error = %v", err)
	}
	capture, err := BuildCapture(CaptureInput{Manager: ManagerNPM, RootKeys: []string{"app@workspace"}, Packages: packages, PolicyIDs: []string{"p"}})
	if err != nil {
		t.Fatal(err)
	}
	root := capture.ProductNodeIDs["app@workspace"]
	platform := closuregraph.TargetPlatformPayload{OS: "linux", Architecture: "arm64", ABI: "node", Libc: "none", MinimumRuntime: "22", SDKID: "none", TargetTriple: "arm64-linux", Runtime: "node", LanguageModes: map[string]string{}, Tuning: map[string]string{}}
	platformNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: platform}
	platformID, _ := platformNode.ID()
	selection, err := closuregraph.NewSelectionContext([]closuregraph.ID{root}, map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID}, []string{}, false, map[string]string{"os": "linux"}, map[string]string{"react": "18"}, []string{"node-marker-v1"})
	if err != nil {
		t.Fatal(err)
	}
	exact := RuntimeBinding{Platform: platform, Node: nodeTool("node-runtime", "bin/node", "node-linux", closuregraph.ExecutionTarget), Manager: nodeTool("package-manager", "bin/npm", "npm-linux", closuregraph.ExecutionTarget), TargetNodeIDs: []closuregraph.ID{root}}
	c0, err := NewC0Checkpoint(capture, selection, exact)
	if err != nil {
		t.Fatal(err)
	}
	exact.C0Checkpoint = &c0
	evaluator := closuregraph.ConditionEvaluatorFunc{EvaluatorID: "node-marker-v1", EvaluateFunc: func(condition closuregraph.Condition, input closuregraph.EvaluationInput) (bool, error) {
		return condition.Expression == "os="+input.Selection.Markers["os"], nil
	}}
	bundle, _, err := Close(capture, selection, exact, []closuregraph.ConditionEvaluator{evaluator}, "node-build-execution-v1")
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Active.EdgeActivations) != 1 || bundle.Active.EdgeActivations[0].State != closuregraph.ActivationSelected || bundle.Active.EdgeActivations[0].Reason != closuregraph.ReasonConditionTrue {
		t.Fatalf("conditional selection evidence = %#v", bundle.Active.EdgeActivations)
	}
	darwin := platform
	darwin.OS, darwin.TargetTriple = "darwin", "arm64-darwin"
	darwinNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: darwin}
	darwinID, _ := darwinNode.ID()
	prunedSelection, err := closuregraph.NewSelectionContext([]closuregraph.ID{root}, map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: darwinID}, []string{}, false, map[string]string{"os": "darwin"}, map[string]string{"react": "18"}, []string{"node-marker-v1"})
	if err != nil {
		t.Fatal(err)
	}
	prunedExact := RuntimeBinding{Platform: darwin, Node: nodeTool("node-runtime", "bin/node", "node-darwin", closuregraph.ExecutionTarget), Manager: nodeTool("package-manager", "bin/npm", "npm-darwin", closuregraph.ExecutionTarget), TargetNodeIDs: []closuregraph.ID{root}}
	prunedC0, err := NewC0Checkpoint(capture, prunedSelection, prunedExact)
	if err != nil {
		t.Fatal(err)
	}
	prunedExact.C0Checkpoint = &prunedC0
	pruned, _, err := Close(capture, prunedSelection, prunedExact, []closuregraph.ConditionEvaluator{evaluator}, "node-build-execution-v1")
	if err != nil {
		t.Fatal(err)
	}
	if len(pruned.Active.EdgeActivations) != 1 || pruned.Active.EdgeActivations[0].State != closuregraph.ActivationPruned || pruned.Active.EdgeActivations[0].Reason != closuregraph.ReasonConditionFalse {
		t.Fatalf("conditional prune evidence = %#v", pruned.Active.EdgeActivations)
	}
}

func TestFeaturePeerRuntimeAndManagerBindingsDriftIndependently(t *testing.T) {
	capture, _ := generatedCapture(t, ManagerNPM)
	root := capture.ProductNodeIDs["app@workspace"]
	platform := closuregraph.TargetPlatformPayload{OS: "linux", Architecture: "arm64", ABI: "node", Libc: "none", MinimumRuntime: "22", SDKID: "none", TargetTriple: "arm64-linux", Runtime: "node", LanguageModes: map[string]string{"module": "esm"}, Tuning: map[string]string{}}
	platformNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: platform}
	platformID, _ := platformNode.ID()
	type identities struct{ binding, active, plan closuregraph.ID }
	project := func(features []string, peers map[string]string, nodeSeed, managerSeed string) identities {
		selection, err := closuregraph.NewSelectionContext([]closuregraph.ID{root}, map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID}, features, false, map[string]string{"os": "linux"}, peers, []string{})
		if err != nil {
			t.Fatal(err)
		}
		exact := RuntimeBinding{Platform: platform, Node: nodeTool("node-runtime", "bin/node", nodeSeed, closuregraph.ExecutionTarget), Manager: nodeTool("package-manager", "bin/npm", managerSeed, closuregraph.ExecutionTarget), TargetNodeIDs: []closuregraph.ID{root}}
		c0, err := NewC0Checkpoint(capture, selection, exact)
		if err != nil {
			t.Fatal(err)
		}
		exact.C0Checkpoint = &c0
		bundle, plan, err := Close(capture, selection, exact, nil, "node-build-execution-v1")
		if err != nil {
			t.Fatal(err)
		}
		bindingID, _ := bundle.Binding.ID()
		activeID, _ := bundle.Active.ID()
		planID, _ := plan.ID()
		return identities{binding: bindingID, active: activeID, plan: planID}
	}
	base := project([]string{}, map[string]string{"react": "18"}, "node-a", "manager-a")
	for name, changed := range map[string]identities{
		"feature": project([]string{"debug"}, map[string]string{"react": "18"}, "node-a", "manager-a"),
		"peer":    project([]string{}, map[string]string{"react": "19"}, "node-a", "manager-a"),
		"runtime": project([]string{}, map[string]string{"react": "18"}, "node-b", "manager-a"),
		"manager": project([]string{}, map[string]string{"react": "18"}, "node-a", "manager-b"),
	} {
		if base.binding == changed.binding || base.active == changed.active || base.plan == changed.plan {
			t.Fatalf("%s-only drift aliased binding/active/plan: base=%#v changed=%#v", name, base, changed)
		}
	}
}

func TestN08UndeclaredGeneratorInputAndMissingC0HaveZeroActionStarts(t *testing.T) {
	capture, _ := generatedCapture(t, ManagerNPM)
	root := capture.ProductNodeIDs["app@workspace"]
	spec := GeneratedAction{Name: "hidden", Argv: []string{"--generate"}, WorkingDirectory: "workspace", Compiler: nodeTool("typescript-compiler", "bin/tsc", "tsc-hidden", closuregraph.ExecutionHost), Inputs: []GeneratedInput{{NodeID: id("ambient-input"), Path: "host/ambient.json", Class: "source.config", Role: "config"}}, TargetNodeID: root, EnvironmentPolicyID: "node-env-v1", ProcessPolicyID: "node-process-v1", Outputs: []GeneratedOutput{{Path: "dist/hidden.js", Class: "source.generated_text", Grammar: "javascript-v1"}}}
	if _, err := AddGeneratedActions(capture, []GeneratedAction{spec}); ErrorCode(err) != CodeBuildDependencyUnlocked {
		t.Fatalf("undeclared generator input error = %v", err)
	}
	platform := closuregraph.TargetPlatformPayload{OS: "linux", Architecture: "arm64", ABI: "node", Libc: "none", MinimumRuntime: "22", SDKID: "none", TargetTriple: "arm64-linux", Runtime: "node", LanguageModes: map[string]string{}, Tuning: map[string]string{}}
	platformNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: platform}
	platformID, _ := platformNode.ID()
	selection, err := closuregraph.NewSelectionContext([]closuregraph.ID{root}, map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID}, []string{}, false, map[string]string{}, map[string]string{}, []string{})
	if err != nil {
		t.Fatal(err)
	}
	exact := RuntimeBinding{Platform: platform, Node: nodeTool("node-runtime", "bin/node", "node", closuregraph.ExecutionTarget), Manager: nodeTool("package-manager", "bin/npm", "npm", closuregraph.ExecutionTarget), TargetNodeIDs: []closuregraph.ID{root}}
	if _, _, _, _, err = Bind(capture, selection, exact); ErrorCode(err) != string(closuregraph.CodeCheckpointInvalid) {
		t.Fatalf("missing C0 error = %v", err)
	}
}

func TestMissingExecutableEvidenceRejectsNodeManagerAndCompiler(t *testing.T) {
	capture, _ := generatedCapture(t, ManagerNPM)
	root := capture.ProductNodeIDs["app@workspace"]
	platform := closuregraph.TargetPlatformPayload{OS: "linux", Architecture: "arm64", ABI: "node", Libc: "none", MinimumRuntime: "22", SDKID: "none", TargetTriple: "arm64-linux", Runtime: "node", LanguageModes: map[string]string{}, Tuning: map[string]string{}}
	platformNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: platform}
	platformID, _ := platformNode.ID()
	selection, err := closuregraph.NewSelectionContext([]closuregraph.ID{root}, map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID}, []string{}, false, map[string]string{}, map[string]string{}, []string{})
	if err != nil {
		t.Fatal(err)
	}
	base := RuntimeBinding{Platform: platform, Node: nodeTool("node-runtime", "bin/node", "node", closuregraph.ExecutionTarget), Manager: nodeTool("package-manager", "bin/npm", "npm", closuregraph.ExecutionTarget), TargetNodeIDs: []closuregraph.ID{root}}
	for _, tc := range []struct {
		name   string
		mutate func(*RuntimeBinding)
	}{
		{name: "node", mutate: func(binding *RuntimeBinding) { binding.Node.ExecutableSHA256 = "" }},
		{name: "manager", mutate: func(binding *RuntimeBinding) { binding.Manager.ExecutableSHA256 = "" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			binding := base
			tc.mutate(&binding)
			if _, err := NewC0Checkpoint(capture, selection, binding); ErrorCode(err) != string(closuregraph.CodeGraphIncomplete) {
				t.Fatalf("missing executable evidence error = %v", err)
			}
		})
	}
	spec := GeneratedAction{Name: "missing-compiler", Argv: []string{"--project"}, WorkingDirectory: "workspace", Compiler: nodeTool("typescript-compiler", "bin/tsc", "tsc-missing", closuregraph.ExecutionHost), Inputs: []GeneratedInput{{NodeID: capture.SourceNodeIDs["app@workspace"], Path: "workspace/src", Class: "source.tree", Role: "source"}}, TargetNodeID: root, EnvironmentPolicyID: "node-env-v1", ProcessPolicyID: "node-process-v1", Outputs: []GeneratedOutput{{Path: "dist/missing.js", Class: "source.generated_text", Grammar: "javascript-v1"}}}
	spec.Compiler.ExecutableSHA256 = ""
	if _, _, _, err := BuildGeneratedAction(spec); ErrorCode(err) != string(closuregraph.CodeGraphIncomplete) {
		t.Fatalf("missing compiler executable evidence error = %v", err)
	}
}

func TestDeclaredGeneratorChainingIsPermutationInvariantAndCyclesReject(t *testing.T) {
	base, err := BuildCapture(CaptureInput{Manager: ManagerNPM, RootKeys: []string{"app@workspace"}, Packages: packageSet(), PolicyIDs: []string{"p"}})
	if err != nil {
		t.Fatal(err)
	}
	root := base.ProductNodeIDs["app@workspace"]
	first := GeneratedAction{Name: "first", Argv: []string{"--emit"}, WorkingDirectory: "workspace", Compiler: nodeTool("typescript-compiler", "bin/tsc", "tsc-chain", closuregraph.ExecutionHost), Inputs: []GeneratedInput{{NodeID: base.SourceNodeIDs["app@workspace"], Path: "workspace/src", Class: "source.tree", Role: "source"}}, TargetNodeID: root, EnvironmentPolicyID: "node-env-v1", ProcessPolicyID: "node-process-v1", Outputs: []GeneratedOutput{{Path: "gen/intermediate.js", Class: "source.generated_text", Grammar: "javascript-v1", Role: "intermediate", Intermediate: true}}}
	_, firstOutputs, _, err := BuildGeneratedAction(first)
	if err != nil {
		t.Fatal(err)
	}
	intermediateID, _ := firstOutputs[0].ID()
	second := GeneratedAction{Name: "second", Argv: []string{"--bundle"}, WorkingDirectory: "workspace", Compiler: nodeTool("typescript-compiler", "bin/tsc", "tsc-chain", closuregraph.ExecutionHost), Inputs: []GeneratedInput{{NodeID: intermediateID, Path: "gen/intermediate.js", Class: "source.generated_text", Role: "source"}}, TargetNodeID: root, EnvironmentPolicyID: "node-env-v1", ProcessPolicyID: "node-process-v1", Outputs: []GeneratedOutput{{Path: "dist/cli.js", Class: "source.generated_text", Grammar: "javascript-v1", Role: "published_command"}}}
	forward, err := AddGeneratedActions(base, []GeneratedAction{first, second})
	if err != nil {
		t.Fatal(err)
	}
	reverse, err := AddGeneratedActions(base, []GeneratedAction{second, first})
	if err != nil {
		t.Fatal(err)
	}
	forwardID, _ := forward.Graph.ID()
	reverseID, _ := reverse.Graph.ID()
	if forwardID != reverseID {
		t.Fatal("generator declaration order changed capture identity")
	}
	_, plan, _ := closeGenerated(t, forward, "linux")
	if len(plan.Waves) != 2 || len(plan.Waves[0]) != 1 || len(plan.Waves[1]) != 1 {
		t.Fatalf("chained generator waves = %#v", plan.Waves)
	}

	_, secondOutputs, _, err := BuildGeneratedAction(second)
	if err != nil {
		t.Fatal(err)
	}
	secondOutputID, _ := secondOutputs[0].ID()
	first.Inputs = []GeneratedInput{{NodeID: secondOutputID, Path: "dist/cli.js", Class: "source.generated_text", Role: "source"}}
	cyclic, err := AddGeneratedActions(base, []GeneratedAction{first, second})
	if err != nil {
		t.Fatal(err)
	}
	root = cyclic.ProductNodeIDs["app@workspace"]
	platform := closuregraph.TargetPlatformPayload{OS: "linux", Architecture: "arm64", ABI: "node", Libc: "none", MinimumRuntime: "22", SDKID: "none", TargetTriple: "arm64-linux", Runtime: "node", LanguageModes: map[string]string{}, Tuning: map[string]string{}}
	platformNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: platform}
	platformID, _ := platformNode.ID()
	selection, _ := closuregraph.NewSelectionContext([]closuregraph.ID{root}, map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID}, []string{}, false, map[string]string{}, map[string]string{}, []string{})
	exact := RuntimeBinding{Platform: platform, Node: nodeTool("node-runtime", "bin/node", "node-cycle", closuregraph.ExecutionTarget), Manager: nodeTool("package-manager", "bin/npm", "npm-cycle", closuregraph.ExecutionTarget), TargetNodeIDs: []closuregraph.ID{root}}
	c0, err := NewC0Checkpoint(cyclic, selection, exact)
	if err != nil {
		t.Fatal(err)
	}
	exact.C0Checkpoint = &c0
	if _, _, err = Close(cyclic, selection, exact, nil, "node-build-execution-v1"); closuregraph.ErrorCode(err) != closuregraph.CodeBuildCycle {
		t.Fatalf("generator cycle error = %v", err)
	}
}

func TestN07N09N13ShippedGeneratedWorkspaceAndFreshManagerState(t *testing.T) {
	input := CaptureInput{Manager: ManagerYarnModern, RootKeys: []string{"app@workspace"}, Packages: packageSet(), PolicyIDs: []string{"p"}, ShippedGenerated: []ShippedGeneratedText{{PackageKey: "app@workspace", Path: "dist/shipped.min.js", Grammar: "javascript-minified-v1", ArtifactManifestID: id("am-app"), TreeDigest: id("shipped-js")}}}
	capture, err := BuildCapture(input)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, node := range capture.Nodes {
		if node.Kind == closuregraph.NodeSourceSet && node.Payload.(closuregraph.SourceSetPayload).SourceClass == "source.generated_text" {
			found = true
		}
	}
	if !found {
		t.Fatal("shipped generated JavaScript was not retained as immutable source.generated_text")
	}
	escaping := packageSet()
	escaping[0].WorkspacePath = "../outside"
	if _, err = BuildCapture(CaptureInput{Manager: ManagerNPM, RootKeys: []string{"app@workspace"}, Packages: escaping, PolicyIDs: []string{"p"}}); ErrorCode(err) != "closure_local_path_escape" {
		t.Fatalf("workspace escape error = %v", err)
	}
	plan, err := PlanFreshMaterialization([]string{"node_modules", ".yarn/install-state.gz", ".pnp.cjs"})
	if err != nil {
		t.Fatal(err)
	}
	if plan.AmbientAuthority || plan.LifecycleMode != "disabled" || !plan.RequiresDerivationReceipt || len(plan.DiscardPaths) != 3 {
		t.Fatalf("fresh materialization plan = %#v", plan)
	}
	if err = ValidateDerivedMaterialization(plan, id("derived-manager-state-receipt")); err != nil {
		t.Fatal(err)
	}
	if ErrorCode(ValidateDerivedMaterialization(plan, "")) != CodeGeneratedOutputDrift {
		t.Fatal("unreceipted manager state was accepted")
	}
}

func TestN06CompiledNodeAddonAndWasmRejectBeforeAnyNodeProcess(t *testing.T) {
	for _, fixture := range []struct {
		path string
		data []byte
	}{
		{path: "prebuilds/addon.node", data: conformance.GNUSharedObject()},
		{path: "dist/module.wasm", data: conformance.Wasm()},
	} {
		t.Run(fixture.path, func(t *testing.T) {
			sum := sha256.Sum256(fixture.data)
			descriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileNodeV1, Manager: string(ManagerNPM), PackageName: "compiled-fixture", PackageVersion: "1.0.0", Origin: artifactpolicy.OriginEvidence{Locator: "fixture://compiled", ImmutableID: "compiled-r1", LockRecord: "lock-r1", ChecksumSHA256: "sha256:" + hex.EncodeToString(sum[:]), Verified: true}}
			_, err := artifactpolicy.NewService().AdmitDependency(t.Context(), artifactpolicy.DependencyRequest{Descriptor: descriptor, Payload: artifactpolicy.Payload{Path: fixture.path, Size: int64(len(fixture.data)), Reader: bytes.NewReader(fixture.data)}})
			if artifactpolicy.ErrorCode(err) != artifactpolicy.CodeCompiledDependency {
				t.Fatalf("compiled Node input error = %v", err)
			}
		})
	}
}

type metadataRunner struct {
	root   string
	starts int
}

func (runner *metadataRunner) Run(_ context.Context, _ closureexec.ExecutionRequest) (closureexec.PortableRunResult, error) {
	runner.starts++
	if err := os.MkdirAll(runner.root, 0o700); err != nil {
		return closureexec.PortableRunResult{}, err
	}
	if err := os.WriteFile(filepath.Join(runner.root, "metadata.json"), []byte("{}\n"), 0o600); err != nil {
		return closureexec.PortableRunResult{}, err
	}
	return closureexec.PortableRunResult{ExitCode: 0, OutputRoot: runner.root}, nil
}

func TestCGP11CGN16CGN17CGN18MetadataRequiresExactC0PermitRecheckAndReceipt(t *testing.T) {
	capture, _ := generatedCapture(t, ManagerNPM)
	bundle, _, exact := closeGenerated(t, capture, "linux")
	c0 := *exact.C0Checkpoint
	c0ID, _ := c0.ID()
	managerNodeID := closuregraph.ID("")
	for _, node := range bundle.Records.BindingNodes {
		if node.Kind == closuregraph.NodeToolchainComponent && node.Payload.(closuregraph.ToolchainComponentPayload).ComponentRole == "package-manager" {
			managerNodeID, _ = node.ID()
		}
	}
	store, err := closureexec.NewCaptureStore(filepath.Join(t.TempDir(), "capture"))
	if err != nil {
		t.Fatal(err)
	}
	handle, err := store.Capture("fixture://lock", 4, bytes.NewReader([]byte("lock")))
	if err != nil {
		t.Fatal(err)
	}
	intake, err := store.Admit(handle, "fixture://lock", closureexec.AdmissionEvidence{PreviousCausalHead: "head-0", ArtifactPolicyID: "policy-v1", SourceProfileID: ProfileID, DetectorRegistryID: "detectors-v1", LimitVectorID: "limits-v1", ArtifactManifestID: id("metadata-input-manifest")})
	if err != nil {
		t.Fatal(err)
	}
	intakeID, _ := intake.ID()
	input := closureexec.AdmittedInput{Receipt: intake, Handle: handle}
	requirement := closureexec.EvidenceRequirement{Path: "metadata.json", SchemaID: "node-metadata-v1", ArtifactManifestID: id("metadata-output-manifest")}
	evidenceID, _ := closuregraph.DomainID("curator-derivation-evidence-schema-v1", map[string]any{"requirements": []any{map[string]any{"artifact_manifest_id": string(requirement.ArtifactManifestID), "path": requirement.Path, "schema_id": requirement.SchemaID}}})
	limits := closureexec.ResourceLimits{OutputBytes: 1024, ReadBytes: 4096, WriteBytes: 2048, WallTimeMillis: 1000, ProcessCount: 1}
	limitID, _ := limits.ID()
	permit := closureexec.DerivationPermit{SchemaID: closureexec.SchemaDerivationPermit, PreviousCausalHead: "head-0", InvocationKey: "node-metadata-v1", InvocationSubtype: closureexec.DerivationMetadata, AdmittedInputReceiptIDs: []closuregraph.ID{intakeID}, InputMounts: []closureexec.InputMount{{ReceiptID: intakeID, Path: "capture/lock"}}, C0CheckpointID: c0ID, ToolchainNodeID: managerNodeID, ToolchainFingerprint: exact.Manager.Fingerprint, ExecutableSHA256: exact.Manager.ExecutableSHA256, Executable: "bin/manager", CWD: "work", Argv: []string{"metadata", "--offline"}, Environment: map[string]string{"HOME": "home"}, HostID: bundle.Selection.PlatformRoles[closuregraph.PlatformTarget], TargetID: bundle.Selection.PlatformRoles[closuregraph.PlatformTarget], AllowedProcesses: []string{"bin/manager"}, ReadRoots: []string{"capture/lock"}, WriteRoots: []string{"metadata.json"}, ExpectedEvidence: []closureexec.EvidenceRequirement{requirement}, Network: "none", RecheckRule: "immediate-exact-v1", ResourceLimits: limits, ResourceLimitID: limitID, EvidenceSchemaID: evidenceID}
	runner := &metadataRunner{root: filepath.Join(t.TempDir(), "output")}
	executor, err := closureexec.NewAssuredExecutor(closureexec.DefaultAssuranceConfig(), runner, nil, "head-0")
	if err != nil {
		t.Fatal(err)
	}
	drifted := permit
	drifted.C0CheckpointID = id("wrong-c0")
	if _, err = ExecuteMetadataDerivation(t.Context(), c0, MetadataDerivationRequest{Executor: executor, Permit: drifted, C0ToolNodes: bundle.Records.BindingNodes, Inputs: map[closuregraph.ID]closureexec.AdmittedInput{intakeID: input}, Recheck: func(context.Context) (closureexec.ToolchainIdentity, error) {
		return closureexec.ToolchainIdentity{Fingerprint: exact.Manager.Fingerprint, ExecutableSHA256: exact.Manager.ExecutableSHA256}, nil
	}}); ErrorCode(err) != string(closuregraph.CodeDerivationUnauthorized) || runner.starts != 0 {
		t.Fatalf("drifted C0 result=%v starts=%d", err, runner.starts)
	}
	for _, tc := range []struct {
		name        string
		mutate      func(*closureexec.DerivationPermit)
		mutateNodes func([]closuregraph.Node) []closuregraph.Node
	}{
		{name: "tool-node", mutate: func(value *closureexec.DerivationPermit) {
			for _, node := range bundle.Records.BindingNodes {
				if node.Kind == closuregraph.NodeToolchainComponent && node.Payload.(closuregraph.ToolchainComponentPayload).ComponentRole == "node-runtime" {
					value.ToolchainNodeID, _ = node.ID()
				}
			}
		}},
		{name: "fingerprint", mutate: func(value *closureexec.DerivationPermit) { value.ToolchainFingerprint = id("substitute-fingerprint") }},
		{name: "executable-sha", mutate: func(value *closureexec.DerivationPermit) { value.ExecutableSHA256 = id("substitute-executable") }},
		{name: "path", mutate: func(value *closureexec.DerivationPermit) { value.Executable = "bin/substitute-manager" }},
		{name: "recheck-rule", mutate: func(value *closureexec.DerivationPermit) { value.RecheckRule = "substitute-recheck-v1" }},
		{name: "host", mutate: func(value *closureexec.DerivationPermit) { value.HostID = id("substitute-host") }},
		{name: "target", mutate: func(value *closureexec.DerivationPermit) { value.TargetID = id("substitute-target") }},
		{name: "execution-domain", mutate: func(*closureexec.DerivationPermit) {}, mutateNodes: func(nodes []closuregraph.Node) []closuregraph.Node {
			changed := append([]closuregraph.Node(nil), nodes...)
			for index, node := range changed {
				if node.Kind != closuregraph.NodeToolchainComponent || node.Payload.(closuregraph.ToolchainComponentPayload).ComponentRole != "package-manager" {
					continue
				}
				payload := node.Payload.(closuregraph.ToolchainComponentPayload)
				payload.ExecutionDomain = closuregraph.ExecutionHost
				payload.PlatformRoleNames = []closuregraph.PlatformRole{closuregraph.PlatformHost}
				changed[index].Payload = payload
			}
			return changed
		}},
	} {
		t.Run("zero-start-"+tc.name, func(t *testing.T) {
			candidate := permit
			tc.mutate(&candidate)
			nodes := bundle.Records.BindingNodes
			if tc.mutateNodes != nil {
				nodes = tc.mutateNodes(nodes)
			}
			caseRunner := &metadataRunner{root: filepath.Join(t.TempDir(), "output")}
			caseExecutor, createErr := closureexec.NewAssuredExecutor(closureexec.DefaultAssuranceConfig(), caseRunner, nil, "head-0")
			if createErr != nil {
				t.Fatal(createErr)
			}
			_, deriveErr := ExecuteMetadataDerivation(t.Context(), c0, MetadataDerivationRequest{Executor: caseExecutor, Permit: candidate, C0ToolNodes: nodes, Inputs: map[closuregraph.ID]closureexec.AdmittedInput{intakeID: input}, Recheck: func(context.Context) (closureexec.ToolchainIdentity, error) {
				return closureexec.ToolchainIdentity{Fingerprint: candidate.ToolchainFingerprint, ExecutableSHA256: candidate.ExecutableSHA256}, nil
			}})
			if ErrorCode(deriveErr) != string(closuregraph.CodeDerivationUnauthorized) || caseRunner.starts != 0 {
				t.Fatalf("substitution result=%v starts=%d", deriveErr, caseRunner.starts)
			}
		})
	}
	driftRunner := &metadataRunner{root: filepath.Join(t.TempDir(), "drift-output")}
	driftExecutor, err := closureexec.NewAssuredExecutor(closureexec.DefaultAssuranceConfig(), driftRunner, nil, "head-0")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ExecuteMetadataDerivation(t.Context(), c0, MetadataDerivationRequest{Executor: driftExecutor, Permit: permit, C0ToolNodes: bundle.Records.BindingNodes, Inputs: map[closuregraph.ID]closureexec.AdmittedInput{intakeID: input}, Recheck: func(context.Context) (closureexec.ToolchainIdentity, error) {
		return closureexec.ToolchainIdentity{Fingerprint: id("drifted-manager"), ExecutableSHA256: exact.Manager.ExecutableSHA256}, nil
	}}); err == nil || driftRunner.starts != 0 {
		t.Fatalf("runtime drift result=%v starts=%d", err, driftRunner.starts)
	}
	receipt, err := ExecuteMetadataDerivation(t.Context(), c0, MetadataDerivationRequest{Executor: executor, Permit: permit, C0ToolNodes: bundle.Records.BindingNodes, Inputs: map[closuregraph.ID]closureexec.AdmittedInput{intakeID: input}, Recheck: func(context.Context) (closureexec.ToolchainIdentity, error) {
		return closureexec.ToolchainIdentity{Fingerprint: exact.Manager.Fingerprint, ExecutableSHA256: exact.Manager.ExecutableSHA256}, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if runner.starts != 1 || receipt.InvocationSubtype != closureexec.DerivationMetadata || receipt.Decision != "success" {
		t.Fatalf("metadata receipt/start evidence = %#v starts=%d", receipt, runner.starts)
	}
}
