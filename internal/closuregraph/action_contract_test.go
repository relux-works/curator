package closuregraph

import (
	"strings"
	"testing"
)

func TestActionTemplateGrammarClosesEveryDeclaredSlot(t *testing.T) {
	base := ActionPayload{Profile: "fixture-source-v1", ActionSubtype: "compiler", ExecutionDomain: ExecutionTarget, ArgvTemplate: []string{"$TOOL(cc)", "$READ(src)", "-o", "$WRITE(bin)"}, ToolSlotNames: []string{"cc"}, ReadSlotNames: []string{"src"}, WriteSlotNames: []string{"bin"}, EnvironmentPolicyID: "env-v1", ProcessPolicyID: "process-v1", Network: "none"}
	tests := []struct {
		name, want string
		mutate     func(ActionPayload) ActionPayload
	}{
		{name: "hidden command executable", want: "argv_template[0]", mutate: func(payload ActionPayload) ActionPayload {
			payload.ArgvTemplate = append([]string{"/usr/bin/hidden"}, payload.ArgvTemplate...)
			return payload
		}},
		{name: "undeclared placeholder", want: "undeclared tool slot", mutate: func(payload ActionPayload) ActionPayload {
			payload.ArgvTemplate[0] = "$TOOL(hidden)"
			return payload
		}},
		{name: "declared placeholder absent", want: "declared but absent", mutate: func(payload ActionPayload) ActionPayload {
			payload.ArgvTemplate[1] = "source.c"
			return payload
		}},
		{name: "unknown placeholder", want: "unsupported action placeholder", mutate: func(payload ActionPayload) ActionPayload {
			payload.ArgvTemplate[1] = "$INPUT(src)"
			return payload
		}},
		{name: "unterminated placeholder", want: "unterminated", mutate: func(payload ActionPayload) ActionPayload {
			payload.ArgvTemplate[3] = "$WRITE(bin"
			return payload
		}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			payload := base
			payload.ArgvTemplate = append([]string{}, base.ArgvTemplate...)
			payload = testCase.mutate(payload)
			node := Node{Kind: NodeAction, LogicalKey: "action:template", Payload: payload}
			if err := node.Validate(); err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("error = %v, want %q", err, testCase.want)
			}
		})
	}
}

func TestSelectedActionEdgesReconcileExactPathsAndClasses(t *testing.T) {
	tests := []struct {
		name, want string
		mutate     func(*testing.T, graphInputs) graphInputs
	}{
		{name: "tool executable path", want: "uses_tool.executable_relative_path", mutate: func(t *testing.T, input graphInputs) graphInputs {
			edges := append([]Edge{}, input.records.BindingEdges...)
			for index := range edges {
				if edges[index].Kind == EdgeUsesTool {
					payload := edges[index].Payload.(UsesToolPayload)
					payload.ExecutableRelativePath = "bin/other-cc"
					edges[index].Payload = payload
				}
			}
			return rebuildBinding(t, input, input.records.BindingNodes, edges)
		}},
		{name: "read path", want: "reads.path", mutate: func(t *testing.T, input graphInputs) graphInputs {
			edges := append([]Edge{}, input.records.CaptureEdges...)
			for index := range edges {
				if edges[index].Kind == EdgeReads {
					payload := edges[index].Payload.(ReadsPayload)
					payload.Path = "hidden.c"
					edges[index].Payload = payload
				}
			}
			return rebuildCapture(t, input, edges)
		}},
		{name: "produces path", want: "produces.path", mutate: func(t *testing.T, input graphInputs) graphInputs {
			edges := append([]Edge{}, input.records.CaptureEdges...)
			for index := range edges {
				if edges[index].Kind == EdgeProduces {
					payload := edges[index].Payload.(ProducesPayload)
					payload.Path = "bin/other"
					edges[index].Payload = payload
				}
			}
			return rebuildCapture(t, input, edges)
		}},
		{name: "produces class", want: "produces.write_class", mutate: func(t *testing.T, input graphInputs) graphInputs {
			edges := append([]Edge{}, input.records.CaptureEdges...)
			for index := range edges {
				if edges[index].Kind == EdgeProduces {
					payload := edges[index].Payload.(ProducesPayload)
					payload.WriteClass = "source.generated_text"
					edges[index].Payload = payload
				}
			}
			return rebuildCapture(t, input, edges)
		}},
		{name: "publication destination", want: "publishes.destination", mutate: func(t *testing.T, input graphInputs) graphInputs {
			edges := append([]Edge{}, input.records.CaptureEdges...)
			for index := range edges {
				if edges[index].Kind == EdgePublishes {
					payload := edges[index].Payload.(PublishesPayload)
					payload.Destination = "bin/other"
					edges[index].Payload = payload
				}
			}
			return rebuildCapture(t, input, edges)
		}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input := testCase.mutate(t, cgp10Inputs(t))
			_, err := projectInputs(t, input)
			requireReferenceIssue(t, err, testCase.want)
		})
	}
}

func TestObservationRejectsProducesWriteClassDrift(t *testing.T) {
	input := cgp10Inputs(t)
	records := input.records
	for index := range records.CaptureEdges {
		if records.CaptureEdges[index].Kind == EdgeProduces {
			payload := records.CaptureEdges[index].Payload.(ProducesPayload)
			payload.WriteClass = "source.generated_text"
			records.CaptureEdges[index].Payload = payload
			producesID := mustEdgeID(t, records.CaptureEdges[index])
			observation := mustGolden[ProducedArtifactObservation](t, loadGoldenRecords(t), "cgp10.observation.one")
			observation.ProducesEdgeID = producesID
			if err := observation.ValidateAgainst(records); err == nil || !strings.Contains(err.Error(), "artifact_local_output_drift") {
				t.Fatalf("error = %v, want output drift", err)
			}
			return
		}
	}
	t.Fatal("missing produces edge")
}

func rebuildCapture(t *testing.T, input graphInputs, edges []Edge) graphInputs {
	t.Helper()
	capture, err := NewCaptureGraph(input.capture.ProfileID, input.capture.PolicyIDs, input.capture.RootNodeIDs, input.records.CaptureNodes, edges, input.capture.ArtifactManifestIDs)
	if err != nil {
		t.Fatal(err)
	}
	captureID, _ := capture.ID()
	selectionID, _ := input.selection.ID()
	binding, err := NewSelectionBinding(captureID, selectionID, input.records.BindingNodes, input.records.BindingEdges)
	if err != nil {
		t.Fatal(err)
	}
	input.capture = capture
	input.binding = binding
	input.records = NewRecordTables(input.records.CaptureNodes, edges, input.records.BindingNodes, input.records.BindingEdges)
	return input
}
