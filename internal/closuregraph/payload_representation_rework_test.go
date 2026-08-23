package closuregraph

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestClosedNodePayloadPointerRepresentationsFailCanonically(t *testing.T) {
	tests := []struct {
		kind     NodeKind
		nonNil   NodePayload
		typedNil NodePayload
	}{
		{NodeCommandProduct, &CommandProductPayload{}, (*CommandProductPayload)(nil)},
		{NodePackageInstance, &PackageInstancePayload{}, (*PackageInstancePayload)(nil)},
		{NodeSourceSet, &SourceSetPayload{}, (*SourceSetPayload)(nil)},
		{NodeTargetUnit, &TargetUnitPayload{}, (*TargetUnitPayload)(nil)},
		{NodeAction, &ActionPayload{}, (*ActionPayload)(nil)},
		{NodeGeneratedArtifact, &GeneratedArtifactPayload{}, (*GeneratedArtifactPayload)(nil)},
		{NodeInteropBoundary, &InteropBoundaryPayload{}, (*InteropBoundaryPayload)(nil)},
		{NodeToolchainComponent, &ToolchainComponentPayload{}, (*ToolchainComponentPayload)(nil)},
		{NodeTargetPlatform, &TargetPlatformPayload{}, (*TargetPlatformPayload)(nil)},
		{NodeOutputArtifact, &OutputArtifactPayload{}, (*OutputArtifactPayload)(nil)},
	}
	for _, testCase := range tests {
		t.Run(string(testCase.kind), func(t *testing.T) {
			var diagnostics []string
			for _, representation := range []struct {
				name    string
				payload NodePayload
			}{{"non-nil", testCase.nonNil}, {"typed-nil", testCase.typedNil}} {
				t.Run(representation.name, func(t *testing.T) {
					node := Node{Kind: testCase.kind, LogicalKey: "node:pointer:" + string(testCase.kind), Payload: representation.payload}
					err := callWithoutPanic(t, node.Validate)
					requirePayloadRepresentationError(t, err, CodeGraphSchemaUnsupported, "node payload")
					diagnostics = append(diagnostics, err.Error())

					_, err = callValueWithoutPanic(t, node.CanonicalBytes)
					requirePayloadRepresentationError(t, err, CodeGraphSchemaUnsupported, "node payload")
					_, err = callValueWithoutPanic(t, node.ID)
					requirePayloadRepresentationError(t, err, CodeGraphSchemaUnsupported, "node payload")
				})
			}
			if diagnostics[0] != diagnostics[1] {
				t.Fatalf("%s pointer diagnostics differ by nilness:\nnon-nil: %s\ntyped-nil: %s", testCase.kind, diagnostics[0], diagnostics[1])
			}
		})
	}
}

func TestClosedEdgePayloadPointerRepresentationsFailCanonically(t *testing.T) {
	tests := []struct {
		kind     EdgeKind
		nonNil   EdgePayload
		typedNil EdgePayload
	}{
		{EdgeDeclares, &DeclaresPayload{}, (*DeclaresPayload)(nil)},
		{EdgeResolvesTo, &ResolvesToPayload{}, (*ResolvesToPayload)(nil)},
		{EdgeRequires, &RequiresPayload{}, (*RequiresPayload)(nil)},
		{EdgeReads, &ReadsPayload{}, (*ReadsPayload)(nil)},
		{EdgeUsesTool, &UsesToolPayload{}, (*UsesToolPayload)(nil)},
		{EdgeTargets, &TargetsPayload{}, (*TargetsPayload)(nil)},
		{EdgeProduces, &ProducesPayload{}, (*ProducesPayload)(nil)},
		{EdgeProvidesInterop, &ProvidesInteropPayload{}, (*ProvidesInteropPayload)(nil)},
		{EdgeConsumesInterop, &ConsumesInteropPayload{}, (*ConsumesInteropPayload)(nil)},
		{EdgeInvokes, &InvokesPayload{}, (*InvokesPayload)(nil)},
		{EdgePublishes, &PublishesPayload{}, (*PublishesPayload)(nil)},
	}
	for _, testCase := range tests {
		t.Run(string(testCase.kind), func(t *testing.T) {
			var diagnostics []string
			for _, representation := range []struct {
				name    string
				payload EdgePayload
			}{{"non-nil", testCase.nonNil}, {"typed-nil", testCase.typedNil}} {
				t.Run(representation.name, func(t *testing.T) {
					edge := Edge{
						Kind: testCase.kind, EdgeKey: "edge:pointer:" + string(testCase.kind),
						FromNodeID: testDigest('1'), ToNodeID: testDigest('2'), Payload: representation.payload,
					}
					err := callWithoutPanic(t, edge.Validate)
					requirePayloadRepresentationError(t, err, CodeGraphSchemaUnsupported, "edge payload")
					diagnostics = append(diagnostics, err.Error())

					_, err = callValueWithoutPanic(t, edge.CanonicalBytes)
					requirePayloadRepresentationError(t, err, CodeGraphSchemaUnsupported, "edge payload")
					_, err = callValueWithoutPanic(t, edge.ID)
					requirePayloadRepresentationError(t, err, CodeGraphSchemaUnsupported, "edge payload")
				})
			}
			if diagnostics[0] != diagnostics[1] {
				t.Fatalf("%s pointer diagnostics differ by nilness:\nnon-nil: %s\ntyped-nil: %s", testCase.kind, diagnostics[0], diagnostics[1])
			}
		})
	}
}

func TestClosedCheckpointPayloadPointerRepresentationsFailCanonically(t *testing.T) {
	tests := []struct {
		name     CheckpointName
		nonNil   CheckpointPayload
		typedNil CheckpointPayload
	}{
		{CheckpointC0, &C0ProfilePayload{}, (*C0ProfilePayload)(nil)},
		{CheckpointC1, &C1ResolvePayload{}, (*C1ResolvePayload)(nil)},
		{CheckpointC2, &C2CapturePayload{}, (*C2CapturePayload)(nil)},
		{CheckpointC3A, &C3AdmitPayload{Phase: "origin"}, (*C3AdmitPayload)(nil)},
		{CheckpointC3B, &C3AdmitPayload{Phase: "derived"}, (*C3AdmitPayload)(nil)},
		{CheckpointC3, &C3AdmitPayload{Phase: "main"}, (*C3AdmitPayload)(nil)},
		{CheckpointC4, &C4ClosePayload{}, (*C4ClosePayload)(nil)},
		{CheckpointC5, &C5PlanPayload{}, (*C5PlanPayload)(nil)},
		{CheckpointC6, &C6OfflinePayload{}, (*C6OfflinePayload)(nil)},
		{CheckpointC7, &C7PublishPayload{}, (*C7PublishPayload)(nil)},
	}
	for _, testCase := range tests {
		t.Run(string(testCase.name), func(t *testing.T) {
			var diagnostics []string
			for _, representation := range []struct {
				name    string
				payload CheckpointPayload
			}{{"non-nil", testCase.nonNil}, {"typed-nil", testCase.typedNil}} {
				t.Run(representation.name, func(t *testing.T) {
					checkpoint := checkpointWithPayloadRepresentation(testCase.name, representation.payload)
					err := callWithoutPanic(t, checkpoint.Validate)
					requirePayloadRepresentationError(t, err, CodeCheckpointInvalid, "checkpoint")
					diagnostics = append(diagnostics, err.Error())

					_, err = callValueWithoutPanic(t, checkpoint.CanonicalBytes)
					requirePayloadRepresentationError(t, err, CodeCheckpointInvalid, "checkpoint")
					_, err = callValueWithoutPanic(t, checkpoint.ID)
					requirePayloadRepresentationError(t, err, CodeCheckpointInvalid, "checkpoint")
					_, err = callValueWithoutPanic(t, func() (Checkpoint, error) {
						return NewCheckpoint(representation.payload, nil, []Diagnostic{})
					})
					requirePayloadRepresentationError(t, err, CodeCheckpointInvalid, "checkpoint payload")
				})
			}
			if diagnostics[0] != diagnostics[1] {
				t.Fatalf("%s pointer diagnostics differ by nilness:\nnon-nil: %s\ntyped-nil: %s", testCase.name, diagnostics[0], diagnostics[1])
			}
		})
	}
}

func TestPointerPayloadGraphDiagnosticsArePermutationIndependent(t *testing.T) {
	input := cgp10Inputs(t)
	nodes := append([]Node{}, input.records.CaptureNodes...)
	edges := append([]Edge{}, input.records.BindingEdges...)
	for index := range nodes {
		switch nodes[index].Kind {
		case NodeAction:
			payload := nodes[index].Payload.(ActionPayload)
			nodes[index].Payload = &payload
		case NodeSourceSet:
			nodes[index].Payload = (*SourceSetPayload)(nil)
		}
	}
	for index := range edges {
		switch edges[index].Kind {
		case EdgeUsesTool:
			edges[index].Payload = (*UsesToolPayload)(nil)
		case EdgeTargets:
			payload := edges[index].Payload.(TargetsPayload)
			edges[index].Payload = &payload
		}
	}

	validate := func(nodes []Node, edges []Edge) []Issue {
		records := NewRecordTables(nodes, input.records.CaptureEdges, input.records.BindingNodes, edges)
		_, err := callValueWithoutPanic(t, func() (GraphBundle, error) {
			return ProjectActive(input.capture, input.selection, input.binding, records, input.authority, nil)
		})
		var validation *ValidationError
		if !errors.As(err, &validation) {
			t.Fatalf("pointer graph error = %T %v, want ValidationError", err, err)
		}
		foundNode, foundEdge := false, false
		for _, issue := range validation.Issues {
			if strings.Contains(issue.Message, "node payload") && strings.Contains(issue.Message, "canonical value representation") {
				foundNode = true
			}
			if strings.Contains(issue.Message, "edge payload") && strings.Contains(issue.Message, "canonical value representation") {
				foundEdge = true
			}
		}
		if !foundNode || !foundEdge {
			t.Fatalf("pointer graph issues omit canonical node/edge representation errors: %#v", validation.Issues)
		}
		return validation.Issues
	}

	left := validate(nodes, edges)
	reverseNodes(nodes)
	reverseEdges(edges)
	right := validate(nodes, edges)
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("pointer diagnostics changed under permutation:\nleft:  %#v\nright: %#v", left, right)
	}
}

func checkpointWithPayloadRepresentation(name CheckpointName, payload CheckpointPayload) Checkpoint {
	checkpoint := Checkpoint{
		SchemaID: SchemaCheckpoint, Name: name, Payload: payload,
		Decision: decisionForCheckpoint(name), Diagnostics: []Diagnostic{},
	}
	if name != CheckpointC0 {
		previous := testDigest('f')
		checkpoint.PreviousCheckpointID = &previous
	}
	return checkpoint
}

func requirePayloadRepresentationError(t *testing.T, err error, code DiagnosticCode, subject string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), string(code)) ||
		!strings.Contains(err.Error(), subject) ||
		!strings.Contains(err.Error(), "canonical value representation") {
		t.Fatalf("error = %v, want %s %s canonical representation rejection", err, code, subject)
	}
}

func callWithoutPanic(t *testing.T, call func() error) (err error) {
	t.Helper()
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("validation panicked instead of returning a canonical diagnostic: %v", recovered)
		}
	}()
	return call()
}

func callValueWithoutPanic[T any](t *testing.T, call func() (T, error)) (value T, err error) {
	t.Helper()
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("validation panicked instead of returning a canonical diagnostic: %v", recovered)
		}
	}()
	return call()
}
