package nodesource

import (
	"fmt"
	"testing"

	"github.com/relux-works/curator/internal/closuregraph"
)

func twoRootCapture(t *testing.T, manager ManagerProfile, roots []string) Capture {
	t.Helper()
	packages := []PackageInstance{
		{Key: "alpha@workspace", Name: "alpha", Version: "1.0.0", Origin: "workspace:alpha", Checksum: "sha256-alpha", WorkspacePath: "alpha", ArtifactManifestID: id("am-alpha"), SnapshotDigest: id("tree-alpha")},
		{Key: "beta@workspace", Name: "beta", Version: "1.0.0", Origin: "workspace:beta", Checksum: "sha256-beta", WorkspacePath: "beta", ArtifactManifestID: id("am-beta"), SnapshotDigest: id("tree-beta")},
	}
	capture, err := BuildCapture(CaptureInput{Manager: manager, RootKeys: roots, Packages: packages, PolicyIDs: []string{"artifact-policy-v1"}})
	if err != nil {
		t.Fatal(err)
	}
	rootSet := map[string]bool{}
	for _, root := range roots {
		rootSet[root] = true
	}
	specs := make([]GeneratedAction, 0, len(roots))
	for _, pkg := range packages {
		if !rootSet[pkg.Key] {
			continue
		}
		specs = append(specs, GeneratedAction{
			Name: pkg.Name, Argv: []string{"--build", pkg.Name}, WorkingDirectory: pkg.WorkspacePath,
			Compiler:     nodeTool("typescript-compiler", "bin/tsc", "tsc-shared", closuregraph.ExecutionHost),
			Inputs:       []GeneratedInput{{NodeID: capture.SourceNodeIDs[pkg.Key], Path: pkg.WorkspacePath + "/src", Class: "source.tree", Role: "source"}},
			TargetNodeID: capture.ProductNodeIDs[pkg.Key], EnvironmentPolicyID: "node-env-v1", ProcessPolicyID: "node-process-v1",
			Outputs: []GeneratedOutput{{Path: "dist/" + pkg.Name + ".js", Class: "source.generated_text", Grammar: "javascript-v1", Role: "published_command"}},
		})
	}
	capture, err = AddGeneratedActions(capture, specs)
	if err != nil {
		t.Fatal(err)
	}
	return capture
}

func closeRoots(t *testing.T, capture Capture, selected, targets []closuregraph.ID, evaluators []closuregraph.ConditionEvaluator) (closuregraph.GraphBundle, closuregraph.BuildPlan) {
	t.Helper()
	platform := closuregraph.TargetPlatformPayload{OS: "linux", Architecture: "arm64", ABI: "node", Libc: "none", MinimumRuntime: "22", SDKID: "none", TargetTriple: "arm64-linux", Runtime: "node", LanguageModes: map[string]string{"module": "esm"}, Tuning: map[string]string{}}
	platformNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: platform}
	platformID, _ := platformNode.ID()
	evaluatorIDs := make([]string, len(evaluators))
	for index, evaluator := range evaluators {
		evaluatorIDs[index] = evaluator.ID()
	}
	selection, err := closuregraph.NewSelectionContext(selected, map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID}, []string{}, false, map[string]string{"os": "linux"}, map[string]string{}, evaluatorIDs)
	if err != nil {
		t.Fatal(err)
	}
	exact := RuntimeBinding{Platform: platform, Node: nodeTool("node-runtime", "bin/node", "node-multi", closuregraph.ExecutionTarget), Manager: nodeTool("package-manager", "bin/manager", "manager-multi", closuregraph.ExecutionTarget), TargetNodeIDs: targets}
	c0, err := NewC0Checkpoint(capture, selection, exact)
	if err != nil {
		t.Fatal(err)
	}
	exact.C0Checkpoint = &c0
	bundle, plan, err := Close(capture, selection, exact, evaluators, "node-build-execution-v1")
	if err != nil {
		t.Fatal(err)
	}
	return bundle, plan
}

func TestAllManagersCanonicalizeTwoRootsAndMultipleTargets(t *testing.T) {
	profiles := []ManagerProfile{ManagerNPM, ManagerPNPM, ManagerYarnClassic, ManagerYarnModern}
	for _, profile := range profiles {
		t.Run(string(profile), func(t *testing.T) {
			left := twoRootCapture(t, profile, []string{"alpha@workspace", "beta@workspace"})
			right := twoRootCapture(t, profile, []string{"beta@workspace", "alpha@workspace"})
			leftCaptureID, _ := left.Graph.ID()
			rightCaptureID, _ := right.Graph.ID()
			if leftCaptureID != rightCaptureID {
				t.Fatalf("root permutation changed capture: %s != %s", leftCaptureID, rightCaptureID)
			}
			products := []closuregraph.ID{left.ProductNodeIDs["alpha@workspace"], left.ProductNodeIDs["beta@workspace"]}
			leftBundle, leftPlan := closeRoots(t, left, products, products, nil)
			rightBundle, rightPlan := closeRoots(t, right, []closuregraph.ID{products[1], products[0]}, []closuregraph.ID{products[1], products[0]}, nil)
			leftBindingID, _ := leftBundle.Binding.ID()
			rightBindingID, _ := rightBundle.Binding.ID()
			leftActiveID, _ := leftBundle.Active.ID()
			rightActiveID, _ := rightBundle.Active.ID()
			leftPlanID, _ := leftPlan.ID()
			rightPlanID, _ := rightPlan.ID()
			if leftBindingID != rightBindingID || leftActiveID != rightActiveID || leftPlanID != rightPlanID {
				t.Fatalf("semantic set permutation changed identities: binding %s/%s active %s/%s plan %s/%s", leftBindingID, rightBindingID, leftActiveID, rightActiveID, leftPlanID, rightPlanID)
			}
		})
	}
}

func observationForOutput(t *testing.T, bundle closuregraph.GraphBundle, outputID closuregraph.ID, seed string) ObservedOutput {
	t.Helper()
	var output closuregraph.OutputArtifactPayload
	var actionID, producesID closuregraph.ID
	for _, node := range bundle.Records.CaptureNodes {
		nodeID, _ := node.ID()
		if nodeID == outputID {
			output = node.Payload.(closuregraph.OutputArtifactPayload)
		}
	}
	for _, edge := range bundle.Records.CaptureEdges {
		if edge.Kind == closuregraph.EdgeProduces && edge.ToNodeID == outputID {
			actionID = edge.FromNodeID
			producesID, _ = edge.ID()
		}
	}
	if output.LogicalPath == "" || !actionID.Valid() || !producesID.Valid() {
		t.Fatalf("output %s has incomplete lineage", outputID)
	}
	return ObservedOutput{Grammar: "javascript-v1", Observation: closuregraph.ProducedArtifactObservation{Class: output.ExpectedClass, ExpectedOutputNodeID: outputID, Path: output.LogicalPath, ProducerActionID: actionID, ProducesEdgeID: producesID, SHA256: id(seed), Size: 17}}
}

func TestOutputObservationsUseExactActivePlanSet(t *testing.T) {
	capture := twoRootCapture(t, ManagerNPM, []string{"alpha@workspace", "beta@workspace"})
	alpha := capture.ProductNodeIDs["alpha@workspace"]
	beta := capture.ProductNodeIDs["beta@workspace"]

	oneBundle, onePlan := closeRoots(t, capture, []closuregraph.ID{alpha}, []closuregraph.ID{alpha}, nil)
	if len(onePlan.DeclaredOutputNodeIDs) != 1 {
		t.Fatalf("one-product plan outputs = %v", onePlan.DeclaredOutputNodeIDs)
	}
	activeObservation := observationForOutput(t, oneBundle, onePlan.DeclaredOutputNodeIDs[0], "alpha-bytes")
	if err := ValidateOutputObservations([]ObservedOutput{activeObservation}, oneBundle, onePlan); err != nil {
		t.Fatalf("inactive output absence rejected: %v", err)
	}
	var inactiveID closuregraph.ID
	activeSet := map[closuregraph.ID]bool{onePlan.DeclaredOutputNodeIDs[0]: true}
	for _, node := range capture.Nodes {
		if node.Kind == closuregraph.NodeOutputArtifact {
			nodeID, _ := node.ID()
			if !activeSet[nodeID] {
				inactiveID = nodeID
			}
		}
	}
	if ErrorCode(ValidateOutputObservations([]ObservedOutput{activeObservation, observationForOutput(t, oneBundle, inactiveID, "inactive-bytes")}, oneBundle, onePlan)) != CodeGeneratedOutputDrift {
		t.Fatal("inactive output observation was accepted")
	}

	twoBundle, twoPlan := closeRoots(t, capture, []closuregraph.ID{alpha, beta}, []closuregraph.ID{alpha, beta}, nil)
	observed := make([]ObservedOutput, len(twoPlan.DeclaredOutputNodeIDs))
	for index, outputID := range twoPlan.DeclaredOutputNodeIDs {
		observed[index] = observationForOutput(t, twoBundle, outputID, fmt.Sprintf("both-%d", index))
	}
	if err := ValidateOutputObservations(observed, twoBundle, twoPlan); err != nil {
		t.Fatalf("two-product exact outputs rejected: %v", err)
	}
}

func TestOutputObservationsRejectForgedBuildPlans(t *testing.T) {
	capture := twoRootCapture(t, ManagerNPM, []string{"alpha@workspace", "beta@workspace"})
	alpha := capture.ProductNodeIDs["alpha@workspace"]
	beta := capture.ProductNodeIDs["beta@workspace"]
	bundle, plan := closeRoots(t, capture, []closuregraph.ID{alpha, beta}, []closuregraph.ID{alpha, beta}, nil)
	observed := make([]ObservedOutput, len(plan.DeclaredOutputNodeIDs))
	for index, outputID := range plan.DeclaredOutputNodeIDs {
		observed[index] = observationForOutput(t, bundle, outputID, fmt.Sprintf("exact-%d", index))
	}

	for name, mutate := range map[string]func(*closuregraph.BuildPlan){
		"zero outputs": func(forged *closuregraph.BuildPlan) {
			forged.DeclaredOutputNodeIDs = []closuregraph.ID{}
		},
		"subset outputs": func(forged *closuregraph.BuildPlan) {
			forged.DeclaredOutputNodeIDs = append([]closuregraph.ID(nil), forged.DeclaredOutputNodeIDs[:1]...)
		},
		"action set": func(forged *closuregraph.BuildPlan) {
			forged.ActionNodeIDs = append([]closuregraph.ID(nil), forged.ActionNodeIDs[:1]...)
			forged.OrderingEdges = []closuregraph.OrderingEdge{}
			forged.Waves = [][]closuregraph.ID{append([]closuregraph.ID(nil), forged.ActionNodeIDs...)}
		},
		"ordering": func(forged *closuregraph.BuildPlan) {
			forged.OrderingEdges = []closuregraph.OrderingEdge{{
				FromActionID: forged.ActionNodeIDs[0], ToActionID: forged.ActionNodeIDs[1],
				Reason: closuregraph.OrderGeneratedInput, SourceEdgeIDs: []closuregraph.ID{id("forged-ordering-source")},
			}}
			forged.Waves = [][]closuregraph.ID{{forged.ActionNodeIDs[0]}, {forged.ActionNodeIDs[1]}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			forged := plan
			forged.ActionNodeIDs = append([]closuregraph.ID(nil), plan.ActionNodeIDs...)
			forged.DeclaredOutputNodeIDs = append([]closuregraph.ID(nil), plan.DeclaredOutputNodeIDs...)
			forged.OrderingEdges = append([]closuregraph.OrderingEdge(nil), plan.OrderingEdges...)
			forged.Waves = make([][]closuregraph.ID, len(plan.Waves))
			for index := range plan.Waves {
				forged.Waves[index] = append([]closuregraph.ID(nil), plan.Waves[index]...)
			}
			mutate(&forged)
			if ErrorCode(ValidateOutputObservations(observed, bundle, forged)) != CodeGeneratedOutputDrift {
				t.Fatalf("forged %s plan was accepted", name)
			}
		})
	}
}

func TestConditionPrunedOutputIsAbsentAndCannotBeObserved(t *testing.T) {
	condition := &closuregraph.Condition{EvaluatorID: "node-marker-v1", Expression: "feature=disabled"}
	capture := twoRootCapture(t, ManagerNPM, []string{"alpha@workspace", "beta@workspace"})
	var betaOutput closuregraph.ID
	for _, node := range capture.Nodes {
		if node.Kind == closuregraph.NodeOutputArtifact && node.Payload.(closuregraph.OutputArtifactPayload).LogicalPath == "dist/beta.js" {
			betaOutput, _ = node.ID()
		}
	}
	conditional := closuregraph.Edge{Kind: closuregraph.EdgeRequires, EdgeKey: "node.conditional.beta-output", FromNodeID: capture.ProductNodeIDs["alpha@workspace"], ToNodeID: betaOutput, Payload: closuregraph.RequiresPayload{Scope: closuregraph.ScopeOptional, Condition: condition, Origin: closuregraph.EvidenceOrigin{Field: "features.beta"}, DependencyKind: "generated-output"}}
	capture.Edges = append(capture.Edges, conditional)
	graph, err := closuregraph.NewCaptureGraph(ProfileID, capture.Graph.PolicyIDs, capture.Graph.RootNodeIDs, capture.Nodes, capture.Edges, capture.Graph.ArtifactManifestIDs)
	if err != nil {
		t.Fatal(err)
	}
	capture.Graph = graph
	evaluator := closuregraph.ConditionEvaluatorFunc{EvaluatorID: "node-marker-v1", EvaluateFunc: func(closuregraph.Condition, closuregraph.EvaluationInput) (bool, error) { return false, nil }}
	root := capture.ProductNodeIDs["alpha@workspace"]
	bundle, plan := closeRoots(t, capture, []closuregraph.ID{root}, []closuregraph.ID{root}, []closuregraph.ConditionEvaluator{evaluator})
	if len(plan.DeclaredOutputNodeIDs) != 1 {
		t.Fatalf("condition-pruned plan outputs = %v", plan.DeclaredOutputNodeIDs)
	}
	active := observationForOutput(t, bundle, plan.DeclaredOutputNodeIDs[0], "active-condition")
	if err := ValidateOutputObservations([]ObservedOutput{active}, bundle, plan); err != nil {
		t.Fatalf("condition-pruned output absence rejected: %v", err)
	}
	if ErrorCode(ValidateOutputObservations([]ObservedOutput{active, observationForOutput(t, bundle, betaOutput, "pruned")}, bundle, plan)) != CodeGeneratedOutputDrift {
		t.Fatal("condition-pruned output observation was accepted")
	}
}

func TestDuplicateRootsAndTargetsReject(t *testing.T) {
	if _, err := BuildCapture(CaptureInput{Manager: ManagerNPM, RootKeys: []string{"alpha@workspace", "alpha@workspace"}, Packages: []PackageInstance{{Key: "alpha@workspace", Name: "alpha", Version: "1", Origin: "workspace:alpha", Checksum: "sha256-alpha", WorkspacePath: "alpha", ArtifactManifestID: id("am-alpha"), SnapshotDigest: id("tree-alpha")}}, PolicyIDs: []string{"p"}}); ErrorCode(err) != string(closuregraph.CodeGraphReferenceInvalid) {
		t.Fatalf("duplicate roots error = %v", err)
	}
	capture := twoRootCapture(t, ManagerNPM, []string{"alpha@workspace"})
	root := capture.ProductNodeIDs["alpha@workspace"]
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("duplicate target should reject, not panic: %v", recovered)
		}
	}()
	platform := closuregraph.TargetPlatformPayload{OS: "linux", Architecture: "arm64", ABI: "node", Libc: "none", MinimumRuntime: "22", SDKID: "none", TargetTriple: "arm64-linux", Runtime: "node", LanguageModes: map[string]string{}, Tuning: map[string]string{}}
	platformNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: platform}
	platformID, _ := platformNode.ID()
	selection, _ := closuregraph.NewSelectionContext([]closuregraph.ID{root}, map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID}, []string{}, false, map[string]string{}, map[string]string{}, []string{})
	exact := RuntimeBinding{Platform: platform, Node: nodeTool("node-runtime", "bin/node", "node-duplicate", closuregraph.ExecutionTarget), Manager: nodeTool("package-manager", "bin/manager", "manager-duplicate", closuregraph.ExecutionTarget), TargetNodeIDs: []closuregraph.ID{root, root}}
	c0, _ := NewC0Checkpoint(capture, selection, exact)
	exact.C0Checkpoint = &c0
	if _, _, _, _, err := Bind(capture, selection, exact); ErrorCode(err) != string(closuregraph.CodeGraphReferenceInvalid) {
		t.Fatalf("duplicate targets error = %v", err)
	}
}
