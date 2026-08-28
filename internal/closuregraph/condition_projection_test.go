package closuregraph

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestTargetUnitConditionsFailClosedWithoutActivationEvidence(t *testing.T) {
	target := fixtureTarget("conditional", "rust")
	payload := target.Payload.(TargetUnitPayload)
	payload.ConditionExpressions = []Condition{{EvaluatorID: "fixture-target-v1", Expression: "target.os == linux"}}
	target.Payload = payload
	if err := target.Validate(); err == nil || !strings.Contains(err.Error(), "conditional capture edges") {
		t.Fatalf("error = %v, want unsupported target-unit condition placement", err)
	}
}

func TestConditionalEvaluationOrderIsCanonicalBeforeAnyEvaluatorCall(t *testing.T) {
	capture, selection, binding, tables, wantExpressions := multiConditionFixture(t, false)
	calls := []string{}
	evaluator := ConditionEvaluatorFunc{EvaluatorID: "fixture-target-v1", EvaluateFunc: func(condition Condition, _ EvaluationInput) (bool, error) {
		calls = append(calls, condition.Expression)
		return false, nil
	}}
	if _, err := ProjectActive(capture, selection, binding, tables, BindingAuthority{}, []ConditionEvaluator{evaluator}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(calls, wantExpressions) {
		t.Fatalf("evaluation calls = %v, want canonical order %v", calls, wantExpressions)
	}

	capture, selection, binding, tables, permutedWant := multiConditionFixture(t, true)
	calls = nil
	if _, err := ProjectActive(capture, selection, binding, tables, BindingAuthority{}, []ConditionEvaluator{evaluator}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(calls, permutedWant) || !reflect.DeepEqual(calls, wantExpressions) {
		t.Fatalf("permuted calls = %v, want %v", calls, wantExpressions)
	}
}

func TestConditionalFailurePrimaryIsPermutationStable(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		capture, selection, binding, tables, wantExpressions := multiConditionFixture(t, reverse)
		calls := []string{}
		evaluator := ConditionEvaluatorFunc{EvaluatorID: "fixture-target-v1", EvaluateFunc: func(condition Condition, _ EvaluationInput) (bool, error) {
			calls = append(calls, condition.Expression)
			return false, errors.New("failure:" + condition.Expression)
		}}
		_, err := ProjectActive(capture, selection, binding, tables, BindingAuthority{}, []ConditionEvaluator{evaluator})
		if err == nil || !strings.Contains(err.Error(), "failure:"+wantExpressions[0]) {
			t.Fatalf("reverse=%v error = %v, want first canonical expression %q", reverse, err, wantExpressions[0])
		}
		if !reflect.DeepEqual(calls, []string{wantExpressions[0]}) {
			t.Fatalf("reverse=%v calls = %v, want one canonical first call", reverse, calls)
		}
	}
}

func multiConditionFixture(t *testing.T, reverse bool) (CaptureGraph, SelectionContext, SelectionBinding, RecordTables, []string) {
	t.Helper()
	product := fixtureProduct("conditional")
	packageNode := func(name string, manifest ID) Node {
		return Node{Kind: NodePackageInstance, LogicalKey: "package:" + name, Payload: PackageInstancePayload{Profile: "fixture-source-v1", Ecosystem: "node", Manager: "npm", Origin: "registry://" + name + "/1.0.0", LockInstanceKey: name + "@1.0.0", Name: name, Version: "1.0.0", ArtifactManifestID: manifest, TrustRole: TrustDependencyInput}}
	}
	left := packageNode("left", testDigest('1'))
	right := packageNode("right", testDigest('2'))
	productID, leftID, rightID := mustNodeID(t, product), mustNodeID(t, left), mustNodeID(t, right)
	edges := []Edge{
		{Kind: EdgeRequires, EdgeKey: "edge:conditional-left", FromNodeID: productID, ToNodeID: leftID, Payload: RequiresPayload{Scope: ScopeOptional, Condition: &Condition{EvaluatorID: "fixture-target-v1", Expression: "feature.left"}, Origin: fixtureOrigin("dependencies.left")}},
		{Kind: EdgeRequires, EdgeKey: "edge:conditional-right", FromNodeID: productID, ToNodeID: rightID, Payload: RequiresPayload{Scope: ScopeOptional, Condition: &Condition{EvaluatorID: "fixture-target-v1", Expression: "feature.right"}, Origin: fixtureOrigin("dependencies.right")}},
	}
	if reverse {
		reverseEdges(edges)
	}
	nodes := []Node{product, left, right}
	capture, err := NewCaptureGraph("fixture-source-v1", []string{"fixture-policy-v1"}, []ID{productID}, nodes, edges, []ID{testDigest('1'), testDigest('2')})
	if err != nil {
		t.Fatal(err)
	}
	platform := Node{Kind: NodeTargetPlatform, LogicalKey: "platform:conditional", Payload: TargetPlatformPayload{OS: "linux", Architecture: "x86_64", ABI: "gnu", Libc: "glibc", MinimumRuntime: "glibc-2.31", SDKID: "fixture-sdk-v1", TargetTriple: "x86_64-unknown-linux-gnu"}}
	platformID := mustNodeID(t, platform)
	selection, err := NewSelectionContext([]ID{productID}, map[PlatformRole]ID{PlatformTarget: platformID}, []string{}, false, map[string]string{}, map[string]string{}, []string{"fixture-target-v1"})
	if err != nil {
		t.Fatal(err)
	}
	target := Edge{Kind: EdgeTargets, EdgeKey: "edge:conditional-target", FromNodeID: productID, ToNodeID: platformID, Payload: TargetsPayload{BindingRole: PlatformTarget, Origin: EvidenceOrigin{Field: "selection.platform_roles.target"}}}
	captureID, _ := capture.ID()
	selectionID, _ := selection.ID()
	binding, err := NewSelectionBinding(captureID, selectionID, []Node{platform}, []Edge{target})
	if err != nil {
		t.Fatal(err)
	}
	type conditionOrder struct {
		id         ID
		expression string
	}
	ordered := make([]conditionOrder, len(edges))
	for index, edge := range edges {
		ordered[index] = conditionOrder{id: mustEdgeID(t, edge), expression: edge.Payload.condition().Expression}
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].id < ordered[j].id })
	want := make([]string, len(ordered))
	for index := range ordered {
		want[index] = ordered[index].expression
	}
	return capture, selection, binding, NewRecordTables(nodes, edges, []Node{platform}, []Edge{target}), want
}

// A conditional consumer side is the only way an adapter can declare an
// interop boundary neutrally and still let the projector prune it: without the
// condition the consumer action would keep the boundary — and, through the
// reverse provides_interop traversal, its provider — reachable on every
// destination.
func TestConsumesInteropConditionIsProjected(t *testing.T) {
	payload := ConsumesInteropPayload{Origin: fixtureOrigin("targets.App.dependencies"), Use: "compile", ABIExpectation: "c-abi-v1"}
	if payload.condition() != nil {
		t.Fatal("an unconditional consumer side reported a condition")
	}
	condition := Condition{EvaluatorID: "fixture-target-v1", Expression: "platform=macos"}
	payload.Condition = &condition
	if got := payload.condition(); got == nil || got.Expression != condition.Expression {
		t.Fatalf("projected condition = %#v", got)
	}
	if _, present := payload.value()["condition"]; !present {
		t.Fatal("canonical payload omitted the declared condition")
	}
	edge := Edge{Kind: EdgeConsumesInterop, EdgeKey: "edge:conditional-consumes", FromNodeID: testDigest('1'), ToNodeID: testDigest('2'), Payload: payload}
	if err := edge.Validate(); err != nil {
		t.Fatalf("valid conditional consumer side rejected: %v", err)
	}
	edge.Payload = ConsumesInteropPayload{Origin: payload.Origin, Use: payload.Use, ABIExpectation: payload.ABIExpectation, Condition: &Condition{EvaluatorID: "", Expression: "platform=macos"}}
	if err := edge.Validate(); err == nil {
		t.Fatal("a malformed consumer-side condition was accepted")
	}
}
