package closureexec

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/protocoljson"
)

func xid(c byte) closuregraph.ID {
	return closuregraph.ID("sha256:" + string(bytes.Repeat([]byte{c}, 64)))
}

func TestVerifyReplayTreeUsesCanonicalFullPathOrder(t *testing.T) {
	root := filepath.Join(t.TempDir(), "replay")
	t.Cleanup(func() {
		_ = filepath.WalkDir(root, func(current string, entry os.DirEntry, err error) error {
			if err == nil {
				if entry.IsDir() {
					_ = os.Chmod(current, 0o700)
				} else {
					_ = os.Chmod(current, 0o600)
				}
			}
			return nil
		})
	})
	virtual := filepath.Join(root, "node_modules", ".pnpm", "a")
	if err := os.MkdirAll(virtual, 0o700); err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{
		"node_modules/.pnpm/a/index.js":              []byte("a\n"),
		"node_modules/.pnpm-workspace-state-v1.json": []byte("{}\n"),
	}
	for name, payload := range files {
		target := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, payload, 0o400); err != nil {
			t.Fatal(err)
		}
	}
	if err := filepath.WalkDir(root, func(current string, entry os.DirEntry, err error) error {
		if err == nil && entry.IsDir() {
			return os.Chmod(current, 0o500) // #nosec G302 -- immutable replay fixture directory.
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	expected := make([]SnapshotFile, 0, len(files))
	for name, payload := range files {
		expected = append(expected, SnapshotFile{Path: name, SHA256: digestBytes(payload), Size: int64(len(payload))})
	}
	sort.Slice(expected, func(i, j int) bool { return expected[i].Path < expected[j].Path })
	if err := verifyReplayTree(root, expected); err != nil {
		t.Fatal(err)
	}
}

func goldenPayload(t *testing.T, name string) []byte {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join("..", "closuregraph", "testdata", "canonical-goldens.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range strings.Split(string(payload), "\n\n") {
		lines := strings.Split(block, "\n")
		if len(lines) >= 3 && lines[0] == "name="+name {
			return []byte(lines[2])
		}
	}
	t.Fatalf("golden %s not found", name)
	return nil
}

func publicationAuthorityFixture(t *testing.T) closuregraph.PublicationEvidence {
	t.Helper()
	decodeNode := func(name string) closuregraph.Node {
		record, err := closuregraph.DecodeNode(goldenPayload(t, name))
		if err != nil {
			t.Fatal(err)
		}
		return record
	}
	decodeEdge := func(name string) closuregraph.Edge {
		record, err := closuregraph.DecodeEdge(goldenPayload(t, name))
		if err != nil {
			t.Fatal(err)
		}
		return record
	}
	capture, err := closuregraph.DecodeCaptureGraph(goldenPayload(t, "cgp10.capture"))
	if err != nil {
		t.Fatal(err)
	}
	selection, err := closuregraph.DecodeSelectionContext(goldenPayload(t, "cgp10.selection"))
	if err != nil {
		t.Fatal(err)
	}
	binding, err := closuregraph.DecodeSelectionBinding(goldenPayload(t, "cgp10.binding"))
	if err != nil {
		t.Fatal(err)
	}
	active, err := closuregraph.DecodeActiveGraph(goldenPayload(t, "cgp10.active"))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := closuregraph.DecodeBuildPlan(goldenPayload(t, "cgp10.plan"))
	if err != nil {
		t.Fatal(err)
	}
	c4, err := closuregraph.DecodeCheckpoint(goldenPayload(t, "cgp10.c4"))
	if err != nil {
		t.Fatal(err)
	}
	c5, err := closuregraph.DecodeCheckpoint(goldenPayload(t, "cgp10.c5"))
	if err != nil {
		t.Fatal(err)
	}
	closure, err := closuregraph.DecodeSourceClosure(goldenPayload(t, "cgp10.closure"))
	if err != nil {
		t.Fatal(err)
	}
	captureNodes := []closuregraph.Node{decodeNode("cgp10.product"), decodeNode("cgp10.source"), decodeNode("cgp10.action"), decodeNode("cgp10.output")}
	captureEdges := []closuregraph.Edge{decodeEdge("cgp10.declares"), decodeEdge("cgp10.reads"), decodeEdge("cgp10.produces"), decodeEdge("cgp10.publishes")}
	platform, toolchain := decodeNode("cgp10.platform"), decodeNode("cgp10.toolchain")
	bindingNodes := []closuregraph.Node{platform, toolchain}
	bindingEdges := []closuregraph.Edge{decodeEdge("cgp10.uses-tool"), decodeEdge("cgp10.targets.product"), decodeEdge("cgp10.targets.action"), decodeEdge("cgp10.targets.toolchain"), decodeEdge("cgp10.targets.output")}
	selector, err := closuregraph.NewToolchainSelector(toolchain)
	if err != nil {
		t.Fatal(err)
	}
	selectorID, _ := selector.ID()
	toolchainID, _ := toolchain.ID()
	authority := closuregraph.BindingAuthority{Toolchains: []closuregraph.ToolchainBindingEvidence{{NodeID: toolchainID, FirstBound: closuregraph.ToolchainBoundAtC4, EvidenceID: selectorID}}, C4Selectors: []closuregraph.ToolchainSelector{selector}}
	graph := closuregraph.GraphBundle{Capture: capture, Selection: selection, Binding: binding, Active: active, Records: closuregraph.NewRecordTables(captureNodes, captureEdges, bindingNodes, bindingEdges), Authority: authority}
	return closuregraph.PublicationEvidence{C4: c4, C5: c5, Graph: graph, Plan: plan, Closure: closure}
}

func multiOutputPublicationFixture(t *testing.T) (closuregraph.PublicationEvidence, closuregraph.ExpectedCacheInput, closuregraph.ExecutionReceipt, []closuregraph.ProducedArtifactObservation) {
	t.Helper()
	base := publicationAuthorityFixture(t)
	captureNodes := append([]closuregraph.Node{}, base.Graph.Records.CaptureNodes...)
	captureEdges := append([]closuregraph.Edge{}, base.Graph.Records.CaptureEdges...)
	bindingNodes := append([]closuregraph.Node{}, base.Graph.Records.BindingNodes...)
	bindingEdges := append([]closuregraph.Edge{}, base.Graph.Records.BindingEdges...)
	var oldActionID, originalOutputID, platformID closuregraph.ID
	var originalOutput closuregraph.Node
	for i, node := range captureNodes {
		id, _ := node.ID()
		switch node.Kind {
		case closuregraph.NodeAction:
			oldActionID = id
			payload := node.Payload.(closuregraph.ActionPayload)
			payload.ArgvTemplate = append(payload.ArgvTemplate, "$WRITE(cli2)")
			payload.WriteSlotNames = []string{"cli", "cli2"}
			node.Payload = payload
			captureNodes[i] = node
		case closuregraph.NodeOutputArtifact:
			originalOutput, originalOutputID = node, id
		}
	}
	newActionID := closuregraph.ID("")
	for _, node := range captureNodes {
		if node.Kind == closuregraph.NodeAction {
			newActionID, _ = node.ID()
		}
	}
	for i, edge := range captureEdges {
		if edge.FromNodeID == oldActionID {
			edge.FromNodeID = newActionID
		}
		if edge.ToNodeID == oldActionID {
			edge.ToNodeID = newActionID
		}
		captureEdges[i] = edge
	}
	for _, node := range bindingNodes {
		if node.Kind == closuregraph.NodeTargetPlatform {
			platformID, _ = node.ID()
		}
	}
	for i, edge := range bindingEdges {
		if edge.FromNodeID == oldActionID {
			edge.FromNodeID = newActionID
		}
		bindingEdges[i] = edge
	}
	secondOutput := originalOutput
	secondOutput.LogicalKey = "output:second"
	secondPayload := secondOutput.Payload.(closuregraph.OutputArtifactPayload)
	secondPayload.LogicalPath = "bin/second"
	secondOutput.Payload = secondPayload
	secondOutputID, _ := secondOutput.ID()
	captureNodes = append(captureNodes, secondOutput)
	secondProduces := closuregraph.Edge{Kind: closuregraph.EdgeProduces, EdgeKey: "edge:compile-produces-second", FromNodeID: newActionID, ToNodeID: secondOutputID, Payload: closuregraph.ProducesPayload{Path: "bin/second", WriteSlot: "cli2"}}
	secondProducesID, _ := secondProduces.ID()
	captureEdges = append(captureEdges, secondProduces)
	bindingEdges = append(bindingEdges, closuregraph.Edge{Kind: closuregraph.EdgeTargets, EdgeKey: "edge:second-targets-platform", FromNodeID: secondOutputID, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.target"}}})
	capture, err := closuregraph.NewCaptureGraph(base.Graph.Capture.ProfileID, base.Graph.Capture.PolicyIDs, base.Graph.Capture.RootNodeIDs, captureNodes, captureEdges, base.Graph.Capture.ArtifactManifestIDs)
	if err != nil {
		t.Fatal(err)
	}
	captureID, _ := capture.ID()
	selectionID, _ := base.Graph.Selection.ID()
	binding, err := closuregraph.NewSelectionBinding(captureID, selectionID, bindingNodes, bindingEdges)
	if err != nil {
		t.Fatal(err)
	}
	records := closuregraph.NewRecordTables(captureNodes, captureEdges, bindingNodes, bindingEdges)
	graph, err := closuregraph.ProjectActive(capture, base.Graph.Selection, binding, records, base.Graph.Authority, nil)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := closuregraph.DeriveBuildPlan(graph, closuregraph.PlanOptions{ExecutionPolicyID: base.Plan.ExecutionPolicyID})
	if err != nil {
		t.Fatal(err)
	}
	activeID, _ := graph.Active.ID()
	bindingID, _ := binding.ID()
	previous := xid('9')
	c4 := closuregraph.Checkpoint{SchemaID: closuregraph.SchemaCheckpoint, Name: closuregraph.CheckpointC4, PreviousCheckpointID: &previous, Payload: closuregraph.C4ClosePayload{ActiveGraphID: activeID, CapturedGraphID: captureID, SelectionBindingID: bindingID, SelectionContextID: selectionID}, Decision: closuregraph.DecisionAdmit, Diagnostics: []closuregraph.Diagnostic{}}
	if err = c4.Validate(); err != nil {
		t.Fatal(err)
	}
	planID, _ := plan.ID()
	c5, err := closuregraph.NewCheckpoint(closuregraph.C5PlanPayload{BuildPlanID: planID}, &c4, []closuregraph.Diagnostic{})
	if err != nil {
		t.Fatal(err)
	}
	closure, err := closuregraph.NewSourceClosure(c5)
	if err != nil {
		t.Fatal(err)
	}
	closureID, _ := closure.ID()
	expected := closuregraph.ExpectedCacheInput{SchemaID: closuregraph.SchemaExpectedCacheInput, ClosureID: closureID, ExpectedOutputNodeIDs: append([]closuregraph.ID{}, plan.DeclaredOutputNodeIDs...)}
	originalProducesID := closuregraph.ID("")
	for _, edge := range captureEdges {
		if edge.Kind == closuregraph.EdgeProduces && edge.ToNodeID == originalOutputID {
			originalProducesID, _ = edge.ID()
		}
	}
	observations := []closuregraph.ProducedArtifactObservation{
		{Class: "native.executable", ExpectedOutputNodeID: originalOutputID, Path: "bin/cli", ProducerActionID: newActionID, ProducesEdgeID: originalProducesID, SHA256: digestBytes([]byte("one")), Size: 3},
		{Class: "native.executable", ExpectedOutputNodeID: secondOutputID, Path: "bin/second", ProducerActionID: newActionID, ProducesEdgeID: secondProducesID, SHA256: digestBytes([]byte("two")), Size: 3},
	}
	sort.Slice(observations, func(i, j int) bool {
		left, _ := observations[i].ID()
		right, _ := observations[j].ID()
		return left < right
	})
	observationIDs := make([]closuregraph.ID, len(observations))
	for i, observation := range observations {
		observationIDs[i], _ = observation.ID()
	}
	execution := closuregraph.ExecutionReceipt{SchemaID: closuregraph.SchemaExecutionReceipt, ActionOrder: append([]closuregraph.ID{}, plan.ActionNodeIDs...), ClosureID: closureID, Decision: "success", Network: "none", ProducedObservationIDs: observationIDs, ToolchainRechecks: "match", WriteSet: []string{"bin/cli", "bin/second"}}
	authority := closuregraph.PublicationEvidence{C4: c4, C5: c5, Graph: graph, Plan: plan, Closure: closure}
	return authority, expected, execution, observations
}

type fakeBoundary struct {
	starts           int
	negotiations     int
	audit            Audit
	err              error
	lossless         bool
	request          func(ExecutionRequest) error
	identity         ProviderIdentity
	negotiated       map[string]ProviderCapabilityReceipt
	negotiateErr     error
	capabilityMutate func(*ProviderCapabilityReceipt)
}

type fakePortableRunner struct {
	starts int
	root   string
	err    error
	extra  bool
}

func (runner *fakePortableRunner) Run(_ context.Context, _ ExecutionRequest) (PortableRunResult, error) {
	runner.starts++
	if runner.err != nil {
		return PortableRunResult{}, runner.err
	}
	if err := os.MkdirAll(filepath.Join(runner.root, "evidence"), 0o700); err != nil {
		return PortableRunResult{}, err
	}
	if err := os.WriteFile(filepath.Join(runner.root, "evidence", "manifest.json"), []byte("src"), 0o600); err != nil {
		return PortableRunResult{}, err
	}
	if runner.extra {
		if err := os.WriteFile(filepath.Join(runner.root, "extra"), []byte("bad"), 0o600); err != nil {
			return PortableRunResult{}, err
		}
	}
	return PortableRunResult{ExitCode: 0, OutputRoot: runner.root}, nil
}

func (b *fakeBoundary) LosslessObservation() bool { return b.lossless }
func (b *fakeBoundary) EnforceAndObserve(_ context.Context, request ExecutionRequest) (Audit, error) {
	b.starts++
	if b.request != nil {
		if err := b.request(request); err != nil {
			return Audit{}, err
		}
	}
	return b.audit, b.err
}
func (b *fakeBoundary) Identity() ProviderIdentity {
	if b.identity.ProviderID == "" {
		return ProviderIdentity{Contract: VerifiedProviderContractID, ProviderID: "fixture.provider", Version: "1.0.0", BinarySHA256: xid('b'), TrustEvidence: "fixture-signature"}
	}
	return b.identity
}
func (b *fakeBoundary) Negotiate(_ context.Context, nonce string) (ProviderCapabilityReceipt, error) {
	b.negotiations++
	if b.negotiateErr != nil {
		return ProviderCapabilityReceipt{}, b.negotiateErr
	}
	if b.negotiated == nil {
		b.negotiated = map[string]ProviderCapabilityReceipt{}
	}
	if receipt, ok := b.negotiated[nonce]; ok {
		return receipt, nil
	}
	now := timeNow()
	receipt := ProviderCapabilityReceipt{Provider: b.Identity(), Health: "healthy", Nonce: nonce, ObservedAt: now.Add(-time.Second), ExpiresAt: now.Add(time.Hour)}
	for _, capability := range verifiedCapabilities {
		receipt.Capabilities = append(receipt.Capabilities, CapabilityEvidence{CapabilityID: capability, Status: "established"})
	}
	if b.capabilityMutate != nil {
		b.capabilityMutate(&receipt)
	}
	b.negotiated[nonce] = receipt
	return receipt, nil
}

func publishCacheFixture(t *testing.T, store *ProtectedStore, input AssuredCacheInput) {
	t.Helper()
	expected := input.Expected
	execution, err := closuregraph.DecodeExecutionReceipt(goldenPayload(t, "cgp10.execution.one"))
	if err != nil {
		t.Fatal(err)
	}
	observation, err := closuregraph.DecodeProducedArtifactObservation(goldenPayload(t, "cgp10.observation.one"))
	if err != nil {
		t.Fatal(err)
	}
	staging := t.TempDir()
	if err = os.Mkdir(filepath.Join(staging, "bin"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(staging, "bin", "cli"), []byte("one"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Publish(publicationAuthorityFixture(t), AssuredCacheInput{Expected: expected, Binding: input.Binding}, execution, []closuregraph.ProducedArtifactObservation{observation}, staging); err != nil {
		t.Fatal(err)
	}
}

func TestImmutableCaptureRechecksAndRejectsMutation(t *testing.T) {
	store, err := NewCaptureStore(filepath.Join(t.TempDir(), "captures"))
	if err != nil {
		t.Fatal(err)
	}
	h, err := store.Capture("fixture://source", 3, bytes.NewReader([]byte("src")))
	if err != nil {
		t.Fatal(err)
	}
	if err = h.Recheck(); err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(h.path, 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(h.path, []byte("bad"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = h.Recheck(); err == nil {
		t.Fatal("mutated capture passed recheck")
	}
}

func permitFixture(receiptID closuregraph.ID) DerivationPermit {
	limits := ResourceLimits{OutputBytes: 1024, ReadBytes: 4096, WriteBytes: 2048, WallTimeMillis: 1000, ProcessCount: 1}
	limitID, _ := limits.ID()
	evidence := []EvidenceRequirement{{Path: "evidence/manifest.json", SchemaID: "fixture-manifest-v1", ArtifactManifestID: xid('6')}}
	evidenceID, _ := evidenceSchemaID(evidence)
	return DerivationPermit{SchemaID: SchemaDerivationPermit, PreviousCausalHead: "head-0", InvocationKey: "pkg:manifest", InvocationSubtype: DerivationManifest, AdmittedInputReceiptIDs: []closuregraph.ID{receiptID}, InputMounts: []InputMount{{ReceiptID: receiptID, Path: "capture/source"}}, C0CheckpointID: xid('1'), ToolchainNodeID: xid('2'), ToolchainFingerprint: xid('3'), ExecutableSHA256: xid('8'), Executable: "bin/tool", CWD: "work", Argv: []string{"--offline"}, Environment: map[string]string{"HOME": "home"}, HostID: xid('4'), TargetID: xid('5'), AllowedProcesses: []string{"bin/tool"}, ReadRoots: []string{"capture/source"}, WriteRoots: []string{"evidence/manifest.json"}, ExpectedEvidence: evidence, Network: "none", RecheckRule: "immediate-exact-v1", ResourceLimits: limits, ResourceLimitID: limitID, EvidenceSchemaID: evidenceID}
}

func portablePermitFixture(receiptID closuregraph.ID) DerivationPermit {
	return permitFixture(receiptID)
}
func verifiedConfigFixture() AssuranceConfig {
	return AssuranceConfig{Mode: AssuranceVerified, ProviderID: "fixture.provider", ProviderVersion: "1.0.0", ProviderBinarySHA256: xid('b'), ProviderTrustEvidence: "fixture-signature"}
}
func portableCacheInput(t *testing.T, expected closuregraph.ExpectedCacheInput) AssuredCacheInput {
	t.Helper()
	input := AssuredCacheInput{Expected: expected, Binding: portableAssuranceBinding()}
	if _, err := input.ID(); err != nil {
		t.Fatal(err)
	}
	return input
}
func toolIdentity(fingerprint byte) ToolchainIdentity {
	return ToolchainIdentity{Fingerprint: xid(fingerprint), ExecutableSHA256: xid('8')}
}
func auditFixture() Audit {
	return Audit{Executable: "bin/tool", CWD: "work", Argv: []string{"--offline"}, Environment: map[string]string{"HOME": "home"}, Processes: []string{"bin/tool"}, Reads: []string{"capture/source"}, Writes: []string{"evidence/manifest.json"}, Evidence: []string{"evidence/manifest.json"}, Network: "none", ExitCode: 0, Outputs: []DerivationOutput{{Path: "evidence/manifest.json", SchemaID: "fixture-manifest-v1", ArtifactManifestID: xid('6'), SHA256: xid('a'), Size: 3}}}
}

func admittedFixture(t *testing.T) (*CaptureStore, AdmittedInput, closuregraph.ID) {
	t.Helper()
	store, err := NewCaptureStore(filepath.Join(t.TempDir(), "captures"))
	if err != nil {
		t.Fatal(err)
	}
	handle, err := store.Capture("fixture://source", 3, bytes.NewReader([]byte("src")))
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := store.Admit(handle, "fixture://source", AdmissionEvidence{PreviousCausalHead: "head-0", ArtifactPolicyID: "policy-v1", SourceProfileID: "source-v1", DetectorRegistryID: "detectors-v1", LimitVectorID: "limits-v1", ArtifactManifestID: xid('7')})
	if err != nil {
		t.Fatal(err)
	}
	id, err := receipt.ID()
	if err != nil {
		t.Fatal(err)
	}
	return store, AdmittedInput{Receipt: receipt, Handle: handle}, id
}

func admittedTreeFixture(t *testing.T) (AdmittedInput, closuregraph.ID, string) {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.Mkdir(source, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "main.txt"), []byte("admitted"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(source, "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "nested", "leaf.txt"), []byte("leaf"), 0o600); err != nil {
		t.Fatal(err)
	}
	store, err := NewCaptureStore(filepath.Join(root, "captures"))
	if err != nil {
		t.Fatal(err)
	}
	tree, err := store.CaptureTree("fixture://tree", source)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = filepath.WalkDir(store.trees, func(path string, entry os.DirEntry, _ error) error {
			if entry == nil || entry.Type()&os.ModeSymlink != 0 {
				return nil
			}
			if entry.IsDir() {
				_ = os.Chmod(path, 0o700)
			} else {
				_ = os.Chmod(path, 0o600)
			}
			return nil
		})
	})
	receipt, err := store.AdmitTree(tree, "fixture://tree", AdmissionEvidence{PreviousCausalHead: "head-0", ArtifactPolicyID: "policy-v1", SourceProfileID: "source-v1", DetectorRegistryID: "detectors-v1", LimitVectorID: "limits-v1", ArtifactManifestID: xid('7')})
	if err != nil {
		t.Fatal(err)
	}
	id, err := receipt.ID()
	if err != nil {
		t.Fatal(err)
	}
	return AdmittedInput{Receipt: receipt, Tree: tree}, id, source
}

func TestDerivationRequiresCommittedPermitAndImmediateToolchainMatch(t *testing.T) {
	_, input, receiptID := admittedFixture(t)
	boundary := &fakeBoundary{audit: auditFixture(), lossless: true}
	executor, err := NewExecutor(boundary, "head-0")
	if err != nil {
		t.Fatal(err)
	}
	permit := permitFixture(receiptID)
	permitID, err := executor.Commit(permit)
	if err != nil {
		t.Fatal(err)
	}
	_, err = executor.Execute(context.Background(), permitID, func(context.Context) (ToolchainIdentity, error) { return toolIdentity('9'), nil }, map[closuregraph.ID]AdmittedInput{receiptID: input})
	if err == nil || boundary.starts != 0 {
		t.Fatalf("drift err=%v starts=%d", err, boundary.starts)
	}
	permitID, err = executor.Commit(permit)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := executor.Execute(context.Background(), permitID, func(context.Context) (ToolchainIdentity, error) { return toolIdentity('3'), nil }, map[closuregraph.ID]AdmittedInput{receiptID: input})
	if err != nil {
		t.Fatal(err)
	}
	if boundary.starts != 1 || receipt.Decision != "success" {
		t.Fatalf("receipt=%+v starts=%d", receipt, boundary.starts)
	}
	if err = executor.VerifyIssuedDerivationReceipt(receipt); err != nil {
		t.Fatalf("issued receipt rejected: %v", err)
	}
	foreign, err := NewExecutor(&fakeBoundary{audit: auditFixture(), lossless: true}, "head-0")
	if err != nil {
		t.Fatal(err)
	}
	if err = foreign.VerifyIssuedDerivationReceipt(receipt); err == nil {
		t.Fatal("foreign executor accepted reconstructed receipt")
	}
}

func TestUnavailableObservationProvidersFailClosed(t *testing.T) {
	if _, err := NewExecutor(nil, "head-0"); err == nil {
		t.Fatal("nil observation provider was accepted")
	}
	if _, err := NewExecutor(&fakeBoundary{}, "head-0"); err == nil {
		t.Fatal("enforcement-only provider was accepted")
	}
	if _, err := NewOSBoundary(t.TempDir(), t.TempDir()); err == nil {
		t.Fatal("unsupported built-in lossless provider was synthesized")
	}
}

func TestPortableIsDefaultAndEmitsOnlyEstablishedEvidence(t *testing.T) {
	_, input, receiptID := admittedFixture(t)
	runner := &fakePortableRunner{root: t.TempDir()}
	executor, err := NewAssuredExecutor(AssuranceConfig{}, runner, nil, "head-0")
	if err != nil {
		t.Fatal(err)
	}
	permitID, err := executor.Commit(portablePermitFixture(receiptID))
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := executor.Execute(context.Background(), permitID, func(context.Context) (ToolchainIdentity, error) { return toolIdentity('3'), nil }, map[closuregraph.ID]AdmittedInput{receiptID: input})
	if err != nil {
		t.Fatal(err)
	}
	if runner.starts != 1 || receipt.AssuranceMode != AssurancePortable || receipt.PolicyID != PortablePolicyID || receipt.Provider != nil || receipt.CapabilityReceiptID != nil {
		t.Fatalf("portable receipt binding = %+v, starts=%d", receipt, runner.starts)
	}
	if receipt.Audit.Network != "not-observed" || len(receipt.Audit.Processes) != 0 || len(receipt.Audit.Reads) != 0 || len(receipt.Audit.Writes) != 0 || !reflect.DeepEqual(receipt.ActualCapabilities, portableCapabilities) {
		t.Fatalf("portable receipt inflated host evidence: %+v", receipt)
	}
}

func TestAssuranceSelectionFailsClosedWithStableDiagnostics(t *testing.T) {
	tests := []struct {
		name     string
		config   AssuranceConfig
		runner   PortableProcessRunner
		provider VerifiedProvider
		code     string
	}{
		{name: "unknown mode", config: AssuranceConfig{Mode: "hardened"}, code: "execution_mode_unknown"},
		{name: "missing verified provider id", config: AssuranceConfig{Mode: AssuranceVerified}, code: "verified_provider_missing"},
		{name: "missing verified provider", config: verifiedConfigFixture(), code: "verified_provider_missing"},
		{name: "portable provider alias", config: DefaultAssuranceConfig(), runner: &fakePortableRunner{root: t.TempDir()}, provider: &fakeBoundary{lossless: true}, code: "assurance_evidence_mismatch"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := NewAssuredExecutor(testCase.config, testCase.runner, testCase.provider, "head-0")
			var diagnostic *DiagnosticError
			if !errors.As(err, &diagnostic) || diagnostic.Code != testCase.code {
				t.Fatalf("err=%v, want %s", err, testCase.code)
			}
		})
	}
}

func TestVerifiedProviderIdentityAndCapabilitiesRecheckedBeforeStart(t *testing.T) {
	_, input, receiptID := admittedFixture(t)
	provider := &fakeBoundary{audit: auditFixture(), lossless: true}
	executor, err := NewAssuredExecutor(verifiedConfigFixture(), nil, provider, "head-0")
	if err != nil {
		t.Fatal(err)
	}
	permitID, err := executor.Commit(permitFixture(receiptID))
	if err != nil {
		t.Fatal(err)
	}
	provider.identity = ProviderIdentity{Contract: VerifiedProviderContractID, ProviderID: "other.provider", Version: "1.0.0", BinarySHA256: xid('b'), TrustEvidence: "fixture-signature"}
	_, err = executor.Execute(context.Background(), permitID, func(context.Context) (ToolchainIdentity, error) { return toolIdentity('3'), nil }, map[closuregraph.ID]AdmittedInput{receiptID: input})
	var diagnostic *DiagnosticError
	if !errors.As(err, &diagnostic) || diagnostic.Code != "verified_provider_identity_invalid" || provider.starts != 0 {
		t.Fatalf("err=%v starts=%d", err, provider.starts)
	}
}

func TestVerifiedCapabilityDriftStartsNoProcess(t *testing.T) {
	_, input, receiptID := admittedFixture(t)
	provider := &fakeBoundary{audit: auditFixture(), lossless: true}
	executor, err := NewAssuredExecutor(verifiedConfigFixture(), nil, provider, "head-0")
	if err != nil {
		t.Fatal(err)
	}
	permitID, err := executor.Commit(permitFixture(receiptID))
	if err != nil {
		t.Fatal(err)
	}
	for nonce, receipt := range provider.negotiated {
		receipt.Capabilities[0].Status = "unavailable"
		provider.negotiated[nonce] = receipt
	}
	_, err = executor.Execute(context.Background(), permitID, func(context.Context) (ToolchainIdentity, error) { return toolIdentity('3'), nil }, map[closuregraph.ID]AdmittedInput{receiptID: input})
	var diagnostic *DiagnosticError
	if !errors.As(err, &diagnostic) || diagnostic.Code != "verified_capabilities_unsatisfied" || provider.starts != 0 {
		t.Fatalf("err=%v starts=%d", err, provider.starts)
	}
}

func TestPortableClaimInflationAndVerifiedReuseRejected(t *testing.T) {
	_, input, receiptID := admittedFixture(t)
	runner := &fakePortableRunner{root: t.TempDir()}
	executor, _ := NewAssuredExecutor(DefaultAssuranceConfig(), runner, nil, "head-0")
	permitID, _ := executor.Commit(portablePermitFixture(receiptID))
	receipt, err := executor.Execute(context.Background(), permitID, func(context.Context) (ToolchainIdentity, error) { return toolIdentity('3'), nil }, map[closuregraph.ID]AdmittedInput{receiptID: input})
	if err != nil {
		t.Fatal(err)
	}
	inflated := receipt
	inflated.ActualCapabilities = append(append([]CapabilityEvidence(nil), receipt.ActualCapabilities...), CapabilityEvidence{CapabilityID: verifiedCapabilities[0], Status: "established"})
	if err = inflated.Validate(); err == nil || !strings.Contains(err.Error(), "assurance_evidence_mismatch") {
		t.Fatalf("inflated portable receipt err=%v", err)
	}
	verified := receipt
	verified.AssuranceMode, verified.PolicyID, verified.ExecutionPolicyID = AssuranceVerified, VerifiedPolicyID, VerifiedExecutionPolicyID
	if err = verified.Validate(); err == nil || !strings.Contains(err.Error(), "assurance_evidence_mismatch") {
		t.Fatalf("portable receipt relabeled verified err=%v", err)
	}
	provider := ProviderIdentity{Contract: VerifiedProviderContractID, ProviderID: "fixture.provider", Version: "1.0.0", BinarySHA256: xid('b'), TrustEvidence: "fixture-signature"}
	if err = receipt.ValidateFor(verifiedConfigFixture(), &provider); err == nil || !strings.Contains(err.Error(), "assurance_evidence_mismatch") {
		t.Fatalf("portable receipt accepted for verified policy err=%v", err)
	}
}

func TestCheckpointIdentitiesCannotAliasAcrossModes(t *testing.T) {
	portable := ExecutionCheckpointIdentity{AssuranceMode: AssurancePortable, PolicyID: PortablePolicyID, ExecutionPolicyID: PortableExecutionPolicyID, ActualCapabilities: append([]CapabilityEvidence(nil), portableCapabilities...), OperationID: xid('d')}
	portableID, err := portable.ID()
	if err != nil {
		t.Fatal(err)
	}
	contract, capabilityID := VerifiedProviderContractID, xid('c')
	provider := ProviderIdentity{Contract: contract, ProviderID: "fixture.provider", Version: "1.0.0", BinarySHA256: xid('b'), TrustEvidence: "fixture-signature"}
	capabilities := make([]CapabilityEvidence, len(verifiedCapabilities))
	for index, capability := range verifiedCapabilities {
		capabilities[index] = CapabilityEvidence{CapabilityID: capability, Status: "established"}
	}
	verified := ExecutionCheckpointIdentity{AssuranceMode: AssuranceVerified, PolicyID: VerifiedPolicyID, ExecutionPolicyID: VerifiedExecutionPolicyID, ProviderContract: &contract, Provider: &provider, CapabilityReceiptID: &capabilityID, ActualCapabilities: capabilities, OperationID: xid('d')}
	verifiedID, err := verified.ID()
	if err != nil {
		t.Fatal(err)
	}
	if portableID == verifiedID {
		t.Fatal("portable and verified checkpoints aliased")
	}
}

func TestPortableDeclaredOutputSetRejectsUndeclaredFile(t *testing.T) {
	_, input, receiptID := admittedFixture(t)
	runner := &fakePortableRunner{root: t.TempDir(), extra: true}
	executor, _ := NewAssuredExecutor(DefaultAssuranceConfig(), runner, nil, "head-0")
	permitID, _ := executor.Commit(portablePermitFixture(receiptID))
	_, err := executor.Execute(context.Background(), permitID, func(context.Context) (ToolchainIdentity, error) { return toolIdentity('3'), nil }, map[closuregraph.ID]AdmittedInput{receiptID: input})
	var diagnostic *DiagnosticError
	if !errors.As(err, &diagnostic) || diagnostic.Code != "closure_write_undeclared" || len(executor.receipts) != 0 {
		t.Fatalf("err=%v receipts=%d", err, len(executor.receipts))
	}
}

func TestCompetingStalePermitStartsZeroProcessesAndIssuesNoReceipt(t *testing.T) {
	_, input, receiptID := admittedFixture(t)
	boundary := &fakeBoundary{audit: auditFixture(), lossless: true}
	executor, err := NewExecutor(boundary, "head-0")
	if err != nil {
		t.Fatal(err)
	}
	first := permitFixture(receiptID)
	second := first
	second.InvocationKey = "pkg:manifest:competitor"
	firstID, err := executor.Commit(first)
	if err != nil {
		t.Fatal(err)
	}
	secondID, err := executor.Commit(second)
	if err != nil {
		t.Fatal(err)
	}
	recheck := func(context.Context) (ToolchainIdentity, error) { return toolIdentity('3'), nil }
	if _, err = executor.Execute(context.Background(), firstID, recheck, map[closuregraph.ID]AdmittedInput{receiptID: input}); err != nil {
		t.Fatal(err)
	}
	starts := boundary.starts
	if _, err = executor.Execute(context.Background(), secondID, recheck, map[closuregraph.ID]AdmittedInput{receiptID: input}); err == nil {
		t.Fatal("stale competing permit executed")
	}
	if boundary.starts != starts || len(executor.receipts) != 1 {
		t.Fatalf("stale permit starts=%d receipts=%d", boundary.starts-starts, len(executor.receipts))
	}
}

func TestDerivationDriftCannotEnterReceiptChain(t *testing.T) {
	_, input, receiptID := admittedFixture(t)
	audit := auditFixture()
	audit.Network = "attempted"
	boundary := &fakeBoundary{audit: audit, lossless: true}
	executor, _ := NewExecutor(boundary, "head-0")
	permitID, _ := executor.Commit(permitFixture(receiptID))
	_, err := executor.Execute(context.Background(), permitID, func(context.Context) (ToolchainIdentity, error) { return toolIdentity('3'), nil }, map[closuregraph.ID]AdmittedInput{receiptID: input})
	var diagnostic *DiagnosticError
	if !errors.As(err, &diagnostic) || diagnostic.Code != "closure_network_attempted" {
		t.Fatalf("err=%v", err)
	}
	if len(executor.receipts) != 0 {
		t.Fatal("drifted derivation produced a receipt")
	}
}

func TestAuthoritativeProviderContractRejectsUndeclaredObservedOperations(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Audit)
	}{
		{name: "child process", mutate: func(audit *Audit) { audit.Processes = append(audit.Processes, "bin/helper") }},
		{name: "read", mutate: func(audit *Audit) { audit.Reads = append(audit.Reads, "ambient/secret") }},
		{name: "write", mutate: func(audit *Audit) { audit.Writes = append(audit.Writes, "ambient/output") }},
		{name: "network", mutate: func(audit *Audit) { audit.Network = "attempted" }},
		{name: "evidence", mutate: func(audit *Audit) { audit.Evidence = append(audit.Evidence, "evidence/extra.json") }},
		{name: "output record", mutate: func(audit *Audit) { audit.Outputs[0].SchemaID = "substituted-v1" }},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, input, receiptID := admittedFixture(t)
			audit := auditFixture()
			testCase.mutate(&audit)
			boundary := &fakeBoundary{audit: audit, lossless: true}
			executor, _ := NewExecutor(boundary, "head-0")
			permitID, _ := executor.Commit(permitFixture(receiptID))
			_, err := executor.Execute(context.Background(), permitID, func(context.Context) (ToolchainIdentity, error) { return toolIdentity('3'), nil }, map[closuregraph.ID]AdmittedInput{receiptID: input})
			if err == nil || boundary.starts != 1 || len(executor.receipts) != 0 {
				t.Fatalf("err=%v starts=%d receipts=%d", err, boundary.starts, len(executor.receipts))
			}
		})
	}
}

func TestCanonicalDerivationPermitAndReceiptRoundTripForEverySubtype(t *testing.T) {
	kinds := []DerivationKind{DerivationManifest, DerivationVendor, DerivationMirror, DerivationMetadata}
	for _, kind := range kinds {
		t.Run(string(kind), func(t *testing.T) {
			_, input, receiptID := admittedFixture(t)
			permit := permitFixture(receiptID)
			permit.InvocationSubtype = kind
			permit.InvocationKey = "pkg:" + string(kind)
			boundary := &fakeBoundary{audit: auditFixture(), lossless: true}
			executor, _ := NewExecutor(boundary, "head-0")
			permitID, err := executor.Commit(permit)
			if err != nil {
				t.Fatal(err)
			}
			permit = executor.committed[permitID]
			permitBytes, err := permit.CanonicalBytes()
			if err != nil {
				t.Fatal(err)
			}
			decodedPermit, err := DecodeDerivationPermit(permitBytes)
			if err != nil || !reflect.DeepEqual(decodedPermit, permit) {
				t.Fatalf("permit round trip = %#v, %v", decodedPermit, err)
			}
			receipt, err := executor.Execute(context.Background(), permitID, func(context.Context) (ToolchainIdentity, error) { return toolIdentity('3'), nil }, map[closuregraph.ID]AdmittedInput{receiptID: input})
			if err != nil {
				t.Fatal(err)
			}
			receiptBytes, err := receipt.CanonicalBytes()
			if err != nil {
				t.Fatal(err)
			}
			decodedReceipt, err := DecodeDerivationReceipt(receiptBytes)
			if err != nil || !reflect.DeepEqual(decodedReceipt, receipt) {
				t.Fatalf("receipt round trip = %#v, %v", decodedReceipt, err)
			}
		})
	}
}

func TestCanonicalDerivationEvidenceDriftFailsClosed(t *testing.T) {
	_, input, receiptID := admittedFixture(t)
	permit := permitFixture(receiptID)
	boundary := &fakeBoundary{audit: auditFixture(), lossless: true}
	executor, _ := NewExecutor(boundary, "head-0")
	permitID, _ := executor.Commit(permit)
	receipt, err := executor.Execute(context.Background(), permitID, func(context.Context) (ToolchainIdentity, error) { return toolIdentity('3'), nil }, map[closuregraph.ID]AdmittedInput{receiptID: input})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		permit  *DerivationPermit
		receipt *DerivationReceipt
	}{
		{name: "resource identity", permit: func() *DerivationPermit { value := permit; value.ResourceLimitID = xid('f'); return &value }()},
		{name: "manifest identity", permit: func() *DerivationPermit {
			value := permit
			value.ExpectedEvidence = append([]EvidenceRequirement{}, permit.ExpectedEvidence...)
			value.ExpectedEvidence[0].ArtifactManifestID = xid('f')
			return &value
		}()},
		{name: "receipt resource identity", receipt: func() *DerivationReceipt {
			value := receipt
			value.ResourceLimits.OutputBytes++
			return &value
		}()},
		{name: "receipt manifest identity", receipt: func() *DerivationReceipt {
			value := receipt
			value.ExpectedEvidence = append([]EvidenceRequirement{}, receipt.ExpectedEvidence...)
			value.ExpectedEvidence[0].ArtifactManifestID = xid('f')
			return &value
		}()},
		{name: "output digest", receipt: func() *DerivationReceipt {
			value := receipt
			value.Outputs = append([]DerivationOutput{}, receipt.Outputs...)
			value.Outputs[0].SHA256 = xid('f')
			return &value
		}()},
		{name: "next causal head", receipt: func() *DerivationReceipt { value := receipt; value.NextCausalHead = xid('f'); return &value }()},
		{name: "missing diagnostics", receipt: func() *DerivationReceipt { value := receipt; value.Diagnostics = nil; return &value }()},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.permit != nil {
				if _, err := testCase.permit.CanonicalBytes(); err == nil {
					t.Fatal("drifted permit encoded")
				}
				return
			}
			if _, err := testCase.receipt.CanonicalBytes(); err == nil {
				t.Fatal("drifted receipt encoded")
			}
		})
	}
}

func TestCGN16AndCGN18RejectedInputStartsNoProcess(t *testing.T) {
	_, _, receiptID := admittedFixture(t)
	boundary := &fakeBoundary{audit: auditFixture(), lossless: true}
	executor, err := NewExecutor(boundary, "head-0")
	if err != nil {
		t.Fatal(err)
	}
	permitID, err := executor.Commit(permitFixture(receiptID))
	if err != nil {
		t.Fatal(err)
	}
	_, err = executor.Execute(context.Background(), permitID, func(context.Context) (ToolchainIdentity, error) { return toolIdentity('3'), nil }, map[closuregraph.ID]AdmittedInput{})
	if err == nil || boundary.starts != 0 {
		t.Fatalf("missing admission err=%v starts=%d", err, boundary.starts)
	}
}

func TestCGN17ExecutableDriftStartsNoProcess(t *testing.T) {
	_, input, receiptID := admittedFixture(t)
	boundary := &fakeBoundary{audit: auditFixture(), lossless: true}
	executor, err := NewExecutor(boundary, "head-0")
	if err != nil {
		t.Fatal(err)
	}
	permitID, err := executor.Commit(permitFixture(receiptID))
	if err != nil {
		t.Fatal(err)
	}
	drift := ToolchainIdentity{Fingerprint: xid('3'), ExecutableSHA256: xid('9')}
	_, err = executor.Execute(context.Background(), permitID, func(context.Context) (ToolchainIdentity, error) { return drift, nil }, map[closuregraph.ID]AdmittedInput{receiptID: input})
	if err == nil || boundary.starts != 0 {
		t.Fatalf("executable drift err=%v starts=%d", err, boundary.starts)
	}
}

func TestImmutableAdmittedTreeReplayAndTimeOfUseRechecks(t *testing.T) {
	input, receiptID, source := admittedTreeFixture(t)
	if err := os.WriteFile(filepath.Join(source, "main.txt"), []byte("ambient"), 0o600); err != nil {
		t.Fatal(err)
	}
	boundary := &fakeBoundary{audit: auditFixture(), lossless: true, request: func(request ExecutionRequest) error {
		if len(request.Inputs) != 1 || !request.Inputs[0].IsTree() || request.Inputs[0].MountPath != "capture/source" {
			return errors.New("provider received an ambient or unbound input")
		}
		protected, err := request.Inputs[0].ProtectedPath()
		if err != nil {
			return err
		}
		payload, err := os.ReadFile(filepath.Join(protected, "main.txt"))
		if err != nil {
			return err
		}
		if string(payload) != "admitted" {
			return errors.New("provider observed ambient source substitution")
		}
		if err := os.WriteFile(filepath.Join(protected, "main.txt"), []byte("write"), 0o600); err == nil {
			return errors.New("protected replay tree was writable")
		}
		return nil
	}}
	executor, _ := NewExecutor(boundary, "head-0")
	permitID, err := executor.Commit(permitFixture(receiptID))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = executor.Execute(context.Background(), permitID, func(context.Context) (ToolchainIdentity, error) { return toolIdentity('3'), nil }, map[closuregraph.ID]AdmittedInput{receiptID: input}); err != nil {
		t.Fatal(err)
	}
}

func TestAdmittedTreeRejectsMissingMutatedLinkedWritableAndSubstitutedInputs(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, AdmittedInput)
	}{
		{name: "missing", mutate: func(t *testing.T, input AdmittedInput) {
			if err := os.Chmod(input.Tree.path, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(filepath.Join(input.Tree.path, "main.txt")); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(input.Tree.path, 0o500); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "mutated", mutate: func(t *testing.T, input AdmittedInput) {
			path := filepath.Join(input.Tree.path, "main.txt")
			if err := os.Chmod(path, 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("mutated"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(path, 0o400); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "linked", mutate: func(t *testing.T, input AdmittedInput) {
			if err := os.Chmod(input.Tree.path, 0o700); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(input.Tree.path, "main.txt")
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("elsewhere", path); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(input.Tree.path, 0o500); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "writable", mutate: func(t *testing.T, input AdmittedInput) {
			if err := os.Chmod(filepath.Join(input.Tree.path, "main.txt"), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "substituted", mutate: func(_ *testing.T, input AdmittedInput) {
			input.Tree.id = string(xid('f'))
		}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input, receiptID, _ := admittedTreeFixture(t)
			testCase.mutate(t, input)
			boundary := &fakeBoundary{audit: auditFixture(), lossless: true}
			executor, _ := NewExecutor(boundary, "head-0")
			permitID, err := executor.Commit(permitFixture(receiptID))
			if err != nil {
				t.Fatal(err)
			}
			_, err = executor.Execute(context.Background(), permitID, func(context.Context) (ToolchainIdentity, error) { return toolIdentity('3'), nil }, map[closuregraph.ID]AdmittedInput{receiptID: input})
			if err == nil || boundary.starts != 0 || len(executor.receipts) != 0 {
				t.Fatalf("err=%v starts=%d receipts=%d", err, boundary.starts, len(executor.receipts))
			}
		})
	}
}

func TestWorkspaceStartsEmptyAndRejectsAmbientCache(t *testing.T) {
	root := filepath.Join(t.TempDir(), "task")
	w, err := PrepareWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = w.RecheckEmptyWritableRoots(); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(w.Cache, "poison"), []byte("ambient"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = w.RecheckEmptyWritableRoots(); err == nil {
		t.Fatal("poisoned ambient cache accepted")
	}
}

func TestTaskPrivateDerivedManagerCacheReceipt(t *testing.T) {
	root := filepath.Join(t.TempDir(), "task")
	w, err := PrepareWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(filepath.Join(w.Cache, "pkg"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(w.Cache, "pkg", "index"), []byte("derived"), 0o600); err != nil {
		t.Fatal(err)
	}
	limits := ResourceLimits{OutputBytes: 1024, ReadBytes: 4096, WriteBytes: 2048, WallTimeMillis: 1000, ProcessCount: 1}
	limitID, _ := limits.ID()
	expectedEvidence := []EvidenceRequirement{{Path: "evidence/cache.json", SchemaID: "fixture-cache-v1", ArtifactManifestID: xid('6')}}
	evidenceID, _ := evidenceSchemaID(expectedEvidence)
	outputs := []DerivationOutput{{Path: "evidence/cache.json", SchemaID: "fixture-cache-v1", ArtifactManifestID: xid('6'), SHA256: xid('7'), Size: 7}}
	receipt := DerivationReceipt{SchemaID: SchemaDerivationReceipt, AssuranceMode: AssurancePortable, PolicyID: PortablePolicyID, ExecutionPolicyID: PortableExecutionPolicyID, ActualCapabilities: append([]CapabilityEvidence(nil), portableCapabilities...), PermitID: xid('1'), BeforeFingerprint: xid('2'), AfterFingerprint: xid('2'), Audit: Audit{Executable: "bin/tool", CWD: "work", Argv: []string{}, Environment: map[string]string{}, Processes: []string{}, Reads: []string{}, Writes: []string{}, Evidence: []string{"evidence/cache.json"}, Network: "not-observed", ExitCode: 0, Outputs: outputs}, InvocationSubtype: DerivationMetadata, ResourceLimits: limits, ResourceLimitID: limitID, ExpectedEvidence: expectedEvidence, EvidenceSchemaID: evidenceID, Outputs: outputs, Diagnostics: []DerivationDiagnostic{}, Decision: "success"}
	receipt.NextCausalHead, err = receipt.deriveNextCausalHead()
	if err != nil {
		t.Fatal(err)
	}
	cacheReceipt, err := w.ObserveDerivedCache(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if len(cacheReceipt.Files) != 1 || cacheReceipt.Files[0].Path != "pkg/index" {
		t.Fatalf("cache receipt = %+v", cacheReceipt)
	}
	provider := ProviderIdentity{Contract: VerifiedProviderContractID, ProviderID: "fixture.provider", Version: "1.0.0", BinarySHA256: xid('b'), TrustEvidence: "fixture-signature"}
	if err = cacheReceipt.ValidateFor(verifiedConfigFixture(), &provider); err == nil || !strings.Contains(err.Error(), "assurance_evidence_mismatch") {
		t.Fatalf("portable cache accepted for verified policy: %v", err)
	}
}

func TestProtectedPublicationExactHitAndPoisonRejection(t *testing.T) {
	staging := t.TempDir()
	if err := os.Mkdir(filepath.Join(staging, "bin"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staging, "bin", "cli"), []byte("one"), 0o600); err != nil {
		t.Fatal(err)
	}
	expected, err := closuregraph.DecodeExpectedCacheInput(goldenPayload(t, "cgp10.expected-cache-input"))
	if err != nil {
		t.Fatal(err)
	}
	observation, err := closuregraph.DecodeProducedArtifactObservation(goldenPayload(t, "cgp10.observation.one"))
	if err != nil {
		t.Fatal(err)
	}
	execution, err := closuregraph.DecodeExecutionReceipt(goldenPayload(t, "cgp10.execution.one"))
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewProtectedStore(filepath.Join(t.TempDir(), "protected"))
	if err != nil {
		t.Fatal(err)
	}
	cacheInput := portableCacheInput(t, expected)
	receipt, err := store.Publish(publicationAuthorityFixture(t), cacheInput, execution, []closuregraph.ProducedArtifactObservation{observation}, staging)
	if err != nil {
		t.Fatal(err)
	}
	hit, err := store.Inspect(cacheInput)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(hit.Publication, receipt) || len(hit.Paths) != 1 {
		t.Fatalf("hit=%+v", hit)
	}
	blob := hit.Paths["bin/cli"]
	if err = os.Chmod(blob, 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(blob, []byte("bad"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Inspect(cacheInput); err == nil {
		t.Fatal("poisoned protected cache reused")
	}
}

func TestProtectedCacheIdentityRejectsCrossModeProviderAndCapabilityReuse(t *testing.T) {
	expected, err := closuregraph.DecodeExpectedCacheInput(goldenPayload(t, "cgp10.expected-cache-input"))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("portable entry under verified policy", func(t *testing.T) {
		store, err := NewProtectedStore(filepath.Join(t.TempDir(), "protected"))
		if err != nil {
			t.Fatal(err)
		}
		publishCacheFixture(t, store, portableCacheInput(t, expected))
		provider := &fakeBoundary{lossless: true}
		executor, err := NewAssuredExecutor(verifiedConfigFixture(), nil, provider, "head-0")
		if err != nil {
			t.Fatal(err)
		}
		operation, err := executor.Preflight(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		verifiedInput, err := operation.CacheInput(expected)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = store.Inspect(verifiedInput); !os.IsNotExist(err) {
			t.Fatalf("portable cache did not produce a verified miss: %v", err)
		}
	})

	t.Run("provider identity and capability receipt drift", func(t *testing.T) {
		store, err := NewProtectedStore(filepath.Join(t.TempDir(), "protected"))
		if err != nil {
			t.Fatal(err)
		}
		firstProvider := &fakeBoundary{lossless: true}
		firstExecutor, _ := NewAssuredExecutor(verifiedConfigFixture(), nil, firstProvider, "head-0")
		firstOperation, err := firstExecutor.Preflight(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		firstInput, err := firstOperation.CacheInput(expected)
		if err != nil {
			t.Fatal(err)
		}
		publishCacheFixture(t, store, firstInput)

		secondOperation, err := firstExecutor.Preflight(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		secondInput, _ := secondOperation.CacheInput(expected)
		if _, err = store.Inspect(secondInput); !os.IsNotExist(err) {
			t.Fatalf("fresh capability receipt reused prior entry: %v", err)
		}

		identity := firstProvider.Identity()
		identity.ProviderID = "other.provider"
		otherConfig := verifiedConfigFixture()
		otherConfig.ProviderID = identity.ProviderID
		otherProvider := &fakeBoundary{lossless: true, identity: identity}
		otherExecutor, err := NewAssuredExecutor(otherConfig, nil, otherProvider, "head-0")
		if err != nil {
			t.Fatal(err)
		}
		otherOperation, err := otherExecutor.Preflight(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		otherInput, _ := otherOperation.CacheInput(expected)
		if _, err = store.Inspect(otherInput); !os.IsNotExist(err) {
			t.Fatalf("cross-provider cache entry reused: %v", err)
		}
	})
}

func TestVerifiedOperationPreflightDominatesCacheAndProcessStart(t *testing.T) {
	provider := &fakeBoundary{lossless: true, negotiateErr: errors.New("unhealthy")}
	executor, err := NewAssuredExecutor(verifiedConfigFixture(), nil, provider, "head-0")
	if err != nil {
		t.Fatal(err)
	}
	operation, err := executor.Preflight(context.Background())
	var diagnostic *DiagnosticError
	if operation != nil || !errors.As(err, &diagnostic) || diagnostic.Code != "verified_provider_unavailable" || provider.starts != 0 || provider.negotiations != 1 {
		t.Fatalf("operation=%v err=%v starts=%d negotiations=%d", operation, err, provider.starts, provider.negotiations)
	}
}

func TestMultiOutputPublicationReconcilesEveryObservation(t *testing.T) {
	authority, expected, execution, observations := multiOutputPublicationFixture(t)
	staging := t.TempDir()
	if err := os.Mkdir(filepath.Join(staging, "bin"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staging, "bin", "cli"), []byte("one"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staging, "bin", "second"), []byte("two"), 0o600); err != nil {
		t.Fatal(err)
	}
	store, err := NewProtectedStore(filepath.Join(t.TempDir(), "protected"))
	if err != nil {
		t.Fatal(err)
	}
	cacheInput := portableCacheInput(t, expected)
	publication, err := store.Publish(authority, cacheInput, execution, observations, staging)
	if err != nil {
		t.Fatal(err)
	}
	hit, err := store.Inspect(cacheInput)
	if err != nil {
		t.Fatal(err)
	}
	if len(hit.Paths) != 2 || !reflect.DeepEqual(hit.Publication, publication) {
		t.Fatalf("multi-output hit = %#v", hit)
	}
}

func TestPublicationRejectsUndeclaredOutput(t *testing.T) {
	staging := t.TempDir()
	if err := os.WriteFile(filepath.Join(staging, "extra"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	store, _ := NewProtectedStore(filepath.Join(t.TempDir(), "protected"))
	expected, _ := closuregraph.DecodeExpectedCacheInput(goldenPayload(t, "cgp10.expected-cache-input"))
	execution, _ := closuregraph.DecodeExecutionReceipt(goldenPayload(t, "cgp10.execution.one"))
	observation, _ := closuregraph.DecodeProducedArtifactObservation(goldenPayload(t, "cgp10.observation.one"))
	if _, err := store.Publish(publicationAuthorityFixture(t), portableCacheInput(t, expected), execution, []closuregraph.ProducedArtifactObservation{observation}, staging); err == nil {
		t.Fatal("undeclared output published")
	}
}

func TestPublicationRejectsPoisonedGraphClosureTargetAndToolReferences(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*closuregraph.PublicationEvidence, *closuregraph.ExecutionReceipt, *closuregraph.ProducedArtifactObservation)
	}{
		{name: "wrong action kind", mutate: func(_ *closuregraph.PublicationEvidence, execution *closuregraph.ExecutionReceipt, observation *closuregraph.ProducedArtifactObservation) {
			observation.ProducerActionID = observation.ExpectedOutputNodeID
			id, _ := observation.ID()
			execution.ProducedObservationIDs = []closuregraph.ID{id}
		}},
		{name: "wrong produces edge", mutate: func(_ *closuregraph.PublicationEvidence, execution *closuregraph.ExecutionReceipt, observation *closuregraph.ProducedArtifactObservation) {
			edge, _ := closuregraph.DecodeEdge(goldenPayload(t, "cgp10.declares"))
			observation.ProducesEdgeID, _ = edge.ID()
			id, _ := observation.ID()
			execution.ProducedObservationIDs = []closuregraph.ID{id}
		}},
		{name: "wrong path", mutate: func(_ *closuregraph.PublicationEvidence, execution *closuregraph.ExecutionReceipt, observation *closuregraph.ProducedArtifactObservation) {
			observation.Path = "bin/other"
			id, _ := observation.ID()
			execution.ProducedObservationIDs = []closuregraph.ID{id}
		}},
		{name: "wrong class", mutate: func(_ *closuregraph.PublicationEvidence, execution *closuregraph.ExecutionReceipt, observation *closuregraph.ProducedArtifactObservation) {
			observation.Class = "archive"
			id, _ := observation.ID()
			execution.ProducedObservationIDs = []closuregraph.ID{id}
		}},
		{name: "wrong closure", mutate: func(authority *closuregraph.PublicationEvidence, _ *closuregraph.ExecutionReceipt, _ *closuregraph.ProducedArtifactObservation) {
			authority.Closure.C5CheckpointID = xid('f')
		}},
		{name: "wrong C4", mutate: func(authority *closuregraph.PublicationEvidence, _ *closuregraph.ExecutionReceipt, _ *closuregraph.ProducedArtifactObservation) {
			payload := authority.C4.Payload.(closuregraph.C4ClosePayload)
			payload.ActiveGraphID = xid('f')
			authority.C4.Payload = payload
		}},
		{name: "wrong C5", mutate: func(authority *closuregraph.PublicationEvidence, _ *closuregraph.ExecutionReceipt, _ *closuregraph.ProducedArtifactObservation) {
			payload := authority.C5.Payload.(closuregraph.C5PlanPayload)
			payload.BuildPlanID = xid('f')
			authority.C5.Payload = payload
		}},
		{name: "tool record drift", mutate: func(authority *closuregraph.PublicationEvidence, _ *closuregraph.ExecutionReceipt, _ *closuregraph.ProducedArtifactObservation) {
			for i, node := range authority.Graph.Records.BindingNodes {
				if node.Kind == closuregraph.NodeToolchainComponent {
					payload := node.Payload.(closuregraph.ToolchainComponentPayload)
					payload.VersionOutput = "substituted"
					node.Payload = payload
					authority.Graph.Records.BindingNodes[i] = node
				}
			}
		}},
		{name: "target record drift", mutate: func(authority *closuregraph.PublicationEvidence, _ *closuregraph.ExecutionReceipt, _ *closuregraph.ProducedArtifactObservation) {
			for i, node := range authority.Graph.Records.BindingNodes {
				if node.Kind == closuregraph.NodeTargetPlatform {
					payload := node.Payload.(closuregraph.TargetPlatformPayload)
					payload.Architecture = "substituted"
					node.Payload = payload
					authority.Graph.Records.BindingNodes[i] = node
				}
			}
		}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			authority := publicationAuthorityFixture(t)
			expected, _ := closuregraph.DecodeExpectedCacheInput(goldenPayload(t, "cgp10.expected-cache-input"))
			execution, _ := closuregraph.DecodeExecutionReceipt(goldenPayload(t, "cgp10.execution.one"))
			observation, _ := closuregraph.DecodeProducedArtifactObservation(goldenPayload(t, "cgp10.observation.one"))
			testCase.mutate(&authority, &execution, &observation)
			staging := t.TempDir()
			if err := os.Mkdir(filepath.Join(staging, "bin"), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(staging, "bin", "cli"), []byte("one"), 0o600); err != nil {
				t.Fatal(err)
			}
			store, err := NewProtectedStore(filepath.Join(t.TempDir(), "protected"))
			if err != nil {
				t.Fatal(err)
			}
			cacheInput := portableCacheInput(t, expected)
			if _, err = store.Publish(authority, cacheInput, execution, []closuregraph.ProducedArtifactObservation{observation}, staging); err == nil {
				t.Fatal("poisoned publication was accepted")
			}
			cacheInputID, _ := cacheInput.ID()
			entry := filepath.Join(store.receipts, strings.TrimPrefix(string(cacheInputID), "sha256:")+".ccj.json")
			if _, statErr := os.Stat(entry); !os.IsNotExist(statErr) {
				t.Fatalf("poisoned publication created entry: %v", statErr)
			}
		})
	}
}

func TestProtectedStoreRejectsPoisonedPreexistingRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "protected")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := NewProtectedStore(root); err == nil {
		t.Fatal("unprotected preexisting root was permission-repaired")
	}
}

func TestProtectedHitRejectsReceiptAndOutputSetTampering(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*storedEntry)
	}{
		{name: "receipt", mutate: func(entry *storedEntry) {
			publication, _ := closuregraph.DecodePublicationReceipt(goldenPayload(t, "cgp10.publication.two"))
			wire, _ := publication.CanonicalBytes()
			entry.PublicationBytes = string(wire)
		}},
		{name: "substituted observation", mutate: func(entry *storedEntry) {
			entry.Outputs[0].ObservationBytes = string(goldenPayload(t, "cgp10.observation.two"))
		}},
		{name: "duplicate path", mutate: func(entry *storedEntry) {
			duplicate := entry.Outputs[0]
			duplicate.ObservationID = xid('f')
			entry.Outputs = append(entry.Outputs, duplicate)
		}},
		{name: "missing output", mutate: func(entry *storedEntry) { entry.Outputs = []storedOutput{} }},
		{name: "extra output", mutate: func(entry *storedEntry) {
			extra := entry.Outputs[0]
			extra.Path = "bin/extra"
			extra.ObservationID = xid('f')
			entry.Outputs = append(entry.Outputs, extra)
		}},
		{name: "digest", mutate: func(entry *storedEntry) { entry.Outputs[0].SHA256 = xid('f') }},
		{name: "size", mutate: func(entry *storedEntry) { entry.Outputs[0].Size++ }},
		{name: "execution reference", mutate: func(entry *storedEntry) { entry.ExecutionReceiptID = xid('f') }},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			expected, _ := closuregraph.DecodeExpectedCacheInput(goldenPayload(t, "cgp10.expected-cache-input"))
			execution, _ := closuregraph.DecodeExecutionReceipt(goldenPayload(t, "cgp10.execution.one"))
			observation, _ := closuregraph.DecodeProducedArtifactObservation(goldenPayload(t, "cgp10.observation.one"))
			staging := t.TempDir()
			if err := os.Mkdir(filepath.Join(staging, "bin"), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(staging, "bin", "cli"), []byte("one"), 0o600); err != nil {
				t.Fatal(err)
			}
			store, err := NewProtectedStore(filepath.Join(t.TempDir(), "protected"))
			if err != nil {
				t.Fatal(err)
			}
			cacheInput := portableCacheInput(t, expected)
			if _, err = store.Publish(publicationAuthorityFixture(t), cacheInput, execution, []closuregraph.ProducedArtifactObservation{observation}, staging); err != nil {
				t.Fatal(err)
			}
			cacheInputID, _ := cacheInput.ID()
			entryPath := filepath.Join(store.receipts, strings.TrimPrefix(string(cacheInputID), "sha256:")+".ccj.json")
			payload, err := os.ReadFile(entryPath)
			if err != nil {
				t.Fatal(err)
			}
			var entry storedEntry
			if err = protocoljson.UnmarshalCanonical(payload, &entry); err != nil {
				t.Fatal(err)
			}
			testCase.mutate(&entry)
			payload, err = protocoljson.MarshalCanonical(entryValue(entry))
			if err != nil {
				t.Fatal(err)
			}
			if err = os.Chmod(entryPath, 0o600); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(entryPath, payload, 0o600); err != nil {
				t.Fatal(err)
			}
			if err = os.Chmod(entryPath, 0o400); err != nil {
				t.Fatal(err)
			}
			if _, err = store.Inspect(cacheInput); err == nil {
				t.Fatal("tampered protected entry was reused")
			}
		})
	}
}

func TestCGP10ExactPublicationBranches(t *testing.T) {
	expected, err := closuregraph.DecodeExpectedCacheInput([]byte(`{"closure_id":"sha256:c66440c54e898549b510fb6e6415c8918cb44899a92ce06a98d671f6928f1c9d","expected_output_node_ids":["sha256:a0502b262aa5a9be018eb43910c58d7d6d6634b87301526727ef31a0ba180691"],"schema_id":"closure-expected-cache-input-v1"}`))
	if err != nil {
		t.Fatal(err)
	}
	branches := []struct{ name, bytesValue, observation, execution, publicationID string }{
		{name: "one", bytesValue: "one", observation: `{"class":"native.executable","expected_output_node_id":"sha256:a0502b262aa5a9be018eb43910c58d7d6d6634b87301526727ef31a0ba180691","path":"bin/cli","producer_action_id":"sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e","produces_edge_id":"sha256:30704c3d7a0d0f937e27de4d3996411e1d88b3261f126150ab287230ac0a45b3","sha256":"sha256:7692c3ad3540bb803c020b3aee66cd8887123234ea0c6e7143c0add73ff431ed","size":3}`, execution: `{"action_order":["sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e"],"closure_id":"sha256:c66440c54e898549b510fb6e6415c8918cb44899a92ce06a98d671f6928f1c9d","decision":"success","network":"none","produced_observation_ids":["sha256:5c7837de3e32a78c9c51c6a199d963ae2d3d7fb46ebfb24c062ae7f67bd065e9"],"schema_id":"closure-execution-receipt-v1","toolchain_rechecks":"match","write_set":["bin/cli"]}`, publicationID: "sha256:be40450ce12e9d10fa27a040040d79a55717ab58f7b8bf357f9fb8be76dfcd08"},
		{name: "two", bytesValue: "two", observation: `{"class":"native.executable","expected_output_node_id":"sha256:a0502b262aa5a9be018eb43910c58d7d6d6634b87301526727ef31a0ba180691","path":"bin/cli","producer_action_id":"sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e","produces_edge_id":"sha256:30704c3d7a0d0f937e27de4d3996411e1d88b3261f126150ab287230ac0a45b3","sha256":"sha256:3fc4ccfe745870e2c0d99f71f30ff0656c8dedd41cc1d7d3d376b0dbe685e2f3","size":3}`, execution: `{"action_order":["sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e"],"closure_id":"sha256:c66440c54e898549b510fb6e6415c8918cb44899a92ce06a98d671f6928f1c9d","decision":"success","network":"none","produced_observation_ids":["sha256:4ce66fd6765802f5692cc81b7229fff3d6ae5a85442da0af37192a0b15ce057a"],"schema_id":"closure-execution-receipt-v1","toolchain_rechecks":"match","write_set":["bin/cli"]}`, publicationID: "sha256:39f8595568f4d5ecad1d46b07ea5f0319b8a001c6029158f90afd28aaa8bc60d"},
	}
	for _, branch := range branches {
		t.Run(branch.name, func(t *testing.T) {
			observation, err := closuregraph.DecodeProducedArtifactObservation([]byte(branch.observation))
			if err != nil {
				t.Fatal(err)
			}
			execution, err := closuregraph.DecodeExecutionReceipt([]byte(branch.execution))
			if err != nil {
				t.Fatal(err)
			}
			staging := t.TempDir()
			if err = os.Mkdir(filepath.Join(staging, "bin"), 0o700); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(staging, "bin", "cli"), []byte(branch.bytesValue), 0o600); err != nil {
				t.Fatal(err)
			}
			store, err := NewProtectedStore(filepath.Join(t.TempDir(), "protected"))
			if err != nil {
				t.Fatal(err)
			}
			publication, err := store.Publish(publicationAuthorityFixture(t), portableCacheInput(t, expected), execution, []closuregraph.ProducedArtifactObservation{observation}, staging)
			if err != nil {
				t.Fatal(err)
			}
			id, err := publication.ID()
			if err != nil {
				t.Fatal(err)
			}
			if string(id) != branch.publicationID {
				t.Fatalf("publication id=%s want %s", id, branch.publicationID)
			}
		})
	}
}
