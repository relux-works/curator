package closuregraph

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestEveryClosedNodeAndEdgeKindHasStrictRoundTripCodec(t *testing.T) {
	from, to := testDigest('1'), testDigest('2')
	product := fixtureProduct("codec")
	productPayload := product.Payload.(CommandProductPayload)
	productPayload.PlatformRoleNames = []PlatformRole{PlatformTarget}
	product.Payload = productPayload
	nodes := []Node{
		product,
		{Kind: NodePackageInstance, LogicalKey: "package:codec", Payload: PackageInstancePayload{Profile: "fixture-source-v1", Ecosystem: "rust", Manager: "cargo", NormalizedSourceID: "registry.codec", Origin: "registry://codec/1.0.0", LockInstanceKey: "codec@1.0.0", Name: "codec", Version: "1.0.0", ArtifactManifestID: testDigest('3'), TrustRole: TrustDependencyInput}},
		{Kind: NodeSourceSet, LogicalKey: "source:codec", Payload: SourceSetPayload{Profile: "fixture-source-v1", Origin: "fixture://codec", ArtifactManifestID: testDigest('4'), Projection: []string{"src/main.go"}, Grammar: "go-source-v1", TrustRole: TrustDependencyInput, SourceClass: "source.text", TreeDigest: testDigest('5')}},
		fixtureTarget("codec", "go"),
		fixtureAction("codec", []string{}, []string{}),
		fixtureGenerated("codec.go"),
		{Kind: NodeInteropBoundary, LogicalKey: "interop:codec", Payload: InteropBoundaryPayload{Profile: "fixture-source-v1", Mode: InteropSubprocessProtocol, ProviderLanguageClasses: []string{"go"}, ConsumerLanguageClasses: []string{"swift"}, ContractDigest: testDigest('6'), ProtocolSchema: "json-v1", InterfaceContract: "stdin-stdout-v1", LinkLoadSemantics: "runtime-invoke"}},
		{Kind: NodeToolchainComponent, LogicalKey: "toolchain:codec", Payload: ToolchainComponentPayload{ComponentRole: "compiler", ContentFingerprint: testDigest('7'), ExecutableRelativePath: "bin/compiler", PlatformABI: "linux-x86_64", PolicySelector: "toolchain-v1", VersionOutput: "compiler 1.0", LinkFingerprintIDs: []ID{}, SDKFactsDigest: testDigest('8'), TimeOfUseRecheckRule: "exact-content-v1", ExecutionDomain: ExecutionTarget}},
		{Kind: NodeTargetPlatform, LogicalKey: "platform:codec", Payload: TargetPlatformPayload{OS: "linux", Architecture: "x86_64", ABI: "gnu", Libc: "glibc", MinimumRuntime: "glibc-2.31", SDKID: "sdk-v1", TargetTriple: "x86_64-unknown-linux-gnu", Runtime: "native", LanguageModes: map[string]string{}, Tuning: map[string]string{}}},
		{Kind: NodeOutputArtifact, LogicalKey: "output:codec", Payload: OutputArtifactPayload{Profile: "fixture-source-v1", LogicalPath: "bin/codec", ExpectedClass: "native.executable", OutputRole: "published_command", CompatibilityPredicate: "target-match-v1", DeclarationDigest: testDigest('9')}},
	}
	seenNodes := map[NodeKind]bool{}
	for _, node := range nodes {
		node := node
		t.Run("node "+string(node.Kind), func(t *testing.T) {
			node.Payload.nodePayload()
			payload, err := node.CanonicalBytes()
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeNode(payload)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded, node) {
				t.Fatalf("decoded = %#v, want %#v", decoded, node)
			}
			seenNodes[node.Kind] = true
		})
	}
	if len(seenNodes) != 10 {
		t.Fatalf("covered %d node kinds", len(seenNodes))
	}

	origin := fixtureOrigin("codec.field")
	condition := Condition{EvaluatorID: "fixture-evaluator-v1", Expression: "feature.codec"}
	edges := []Edge{
		{Kind: EdgeDeclares, EdgeKey: "edge:declares", FromNodeID: from, ToNodeID: to, Payload: DeclaresPayload{Origin: origin}},
		{Kind: EdgeResolvesTo, EdgeKey: "edge:resolves", FromNodeID: from, ToNodeID: to, Payload: ResolvesToPayload{LockField: "lock.codec", Origin: origin, Checksum: "sha512-codec", ArtifactManifestID: testDigest('a')}},
		{Kind: EdgeRequires, EdgeKey: "edge:requires", FromNodeID: from, ToNodeID: to, Payload: RequiresPayload{Scope: ScopeOptional, Condition: &condition, Origin: origin, DependencyKind: "package"}},
		{Kind: EdgeReads, EdgeKey: "edge:reads", FromNodeID: from, ToNodeID: to, Payload: ReadsPayload{Path: "src/codec.go", ReadSlot: "source", ReadClass: "source", Projection: []string{"src/codec.go"}}},
		{Kind: EdgeUsesTool, EdgeKey: "edge:uses", FromNodeID: from, ToNodeID: to, Payload: UsesToolPayload{ExecutableRelativePath: "bin/compiler", ToolSlot: "compiler", InvocationRole: "compile"}},
		{Kind: EdgeTargets, EdgeKey: "edge:targets", FromNodeID: from, ToNodeID: to, Payload: TargetsPayload{BindingRole: PlatformTarget, Origin: EvidenceOrigin{Field: "selection.platform_roles.target"}}},
		{Kind: EdgeProduces, EdgeKey: "edge:produces", FromNodeID: from, ToNodeID: to, Payload: ProducesPayload{Path: "bin/codec", WriteSlot: "output", WriteClass: "native.executable"}},
		{Kind: EdgeProvidesInterop, EdgeKey: "edge:provides", FromNodeID: from, ToNodeID: to, Payload: ProvidesInteropPayload{Origin: origin, EvidenceIDs: []ID{testDigest('b')}, ExportRole: "headers", LinkMode: "static"}},
		{Kind: EdgeConsumesInterop, EdgeKey: "edge:consumes", FromNodeID: from, ToNodeID: to, Payload: ConsumesInteropPayload{Origin: origin, Use: "compile", ABIExpectation: "c-abi-v1"}},
		{Kind: EdgeInvokes, EdgeKey: "edge:invokes", FromNodeID: from, ToNodeID: to, Payload: InvokesPayload{ProtocolSchema: "json-v1", ExecutableResolution: "bundle-relative-v1", ArgumentsContract: "argv-v1", EnvironmentContract: "env-v1", WorkingDirectory: "work"}},
		{Kind: EdgePublishes, EdgeKey: "edge:publishes", FromNodeID: from, ToNodeID: to, Payload: PublishesPayload{Destination: "bin/codec", EntryPoint: "codec"}},
	}
	seenEdges := map[EdgeKind]bool{}
	for _, edge := range edges {
		edge := edge
		t.Run("edge "+string(edge.Kind), func(t *testing.T) {
			edge.Payload.edgePayload()
			_ = edge.Payload.condition()
			payload, err := edge.CanonicalBytes()
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeEdge(payload)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded, edge) {
				t.Fatalf("decoded = %#v, want %#v", decoded, edge)
			}
			seenEdges[edge.Kind] = true
		})
	}
	if len(seenEdges) != 11 {
		t.Fatalf("covered %d edge kinds", len(seenEdges))
	}
}

func TestNodeAndEdgeCodecsRejectEveryWrongTypedOptionalField(t *testing.T) {
	product := fixtureProduct("strict-optionals")
	product = nodeWithPlatformRoles(product, []PlatformRole{PlatformTarget})
	packageNode := Node{Kind: NodePackageInstance, LogicalKey: "package:strict-optionals", Payload: PackageInstancePayload{Profile: "fixture-source-v1", Ecosystem: "rust", Manager: "cargo", NormalizedSourceID: "registry.strict-optionals", Origin: "registry://strict-optionals/1.0.0", LockInstanceKey: "strict-optionals@1.0.0", Name: "strict-optionals", Version: "1.0.0", ArtifactManifestID: testDigest('1'), SnapshotDigest: testDigest('2'), WorkspacePath: "workspace/strict-optionals", TrustRole: TrustDependencyInput}}
	source := Node{Kind: NodeSourceSet, LogicalKey: "source:strict-optionals", Payload: SourceSetPayload{Profile: "fixture-source-v1", Origin: "fixture://strict-optionals", ArtifactManifestID: testDigest('3'), Projection: []string{"src/main.go"}, Grammar: "go-source-v1", TrustRole: TrustDependencyInput, SourceClass: "source.text", TreeDigest: testDigest('4')}}
	target := nodeWithPlatformRoles(fixtureTarget("strict-optionals", "go"), []PlatformRole{PlatformTarget})
	action := fixtureAction("strict-optionals", []string{}, []string{})
	actionPayload := action.Payload.(ActionPayload)
	actionPayload.WorkingDirectoryTemplate = "work"
	actionPayload.PlatformRoleNames = []PlatformRole{PlatformTarget}
	action.Payload = actionPayload
	interop := nodeWithPlatformRoles(roleInteropFixture("strict-optionals", InteropCXX), []PlatformRole{PlatformTarget})
	toolchain := Node{Kind: NodeToolchainComponent, LogicalKey: "toolchain:strict-optionals", Payload: ToolchainComponentPayload{ComponentRole: "compiler", ContentFingerprint: testDigest('5'), ExecutableRelativePath: "bin/compiler", PlatformABI: "linux-x86_64", PolicySelector: "toolchain-v1", VersionOutput: "compiler 1.0", LinkFingerprintIDs: []ID{testDigest('6')}, SDKFactsDigest: testDigest('7'), TimeOfUseRecheckRule: "exact-content-v1", ExecutionDomain: ExecutionTarget, PlatformRoleNames: []PlatformRole{PlatformTarget}}}
	platform := Node{Kind: NodeTargetPlatform, LogicalKey: "platform:strict-optionals", Payload: TargetPlatformPayload{OS: "linux", Architecture: "x86_64", ABI: "gnu", Libc: "glibc", MinimumRuntime: "glibc-2.31", SDKID: "sdk-v1", TargetTriple: "x86_64-unknown-linux-gnu", Runtime: "native", LanguageModes: map[string]string{"go": "go1.25"}, Tuning: map[string]string{"cpu": "generic"}}}
	output := Node{Kind: NodeOutputArtifact, LogicalKey: "output:strict-optionals", Payload: OutputArtifactPayload{Profile: "fixture-source-v1", LogicalPath: "bin/strict-optionals", ExpectedClass: "native.executable", OutputRole: "published_command", CompatibilityPredicate: "target-match-v1", DeclarationDigest: testDigest('8'), PlatformRoleNames: []PlatformRole{PlatformTarget}}}
	nodeCases := []struct {
		node   Node
		fields []string
	}{
		{node: product, fields: []string{"platform_role_names"}},
		{node: packageNode, fields: []string{"manager", "normalized_source_id", "artifact_manifest_id", "snapshot_digest", "workspace_path"}},
		{node: source, fields: []string{"source_class", "tree_digest"}},
		{node: target, fields: []string{"platform_role_names"}},
		{node: action, fields: []string{"working_directory_template", "platform_role_names"}},
		{node: interop, fields: []string{"abi", "runtime", "protocol_schema", "interface_contract", "calling_convention", "link_load_semantics", "platform_role_names"}},
		{node: toolchain, fields: []string{"link_fingerprint_ids", "sdk_facts_digest", "time_of_use_recheck_rule", "execution_domain", "platform_role_names"}},
		{node: platform, fields: []string{"runtime", "language_modes", "tuning"}},
		{node: output, fields: []string{"compatibility_predicate", "declaration_digest", "platform_role_names"}},
	}
	for _, testCase := range nodeCases {
		for _, field := range testCase.fields {
			t.Run("node/"+string(testCase.node.Kind)+"/"+field, func(t *testing.T) {
				canonical, err := testCase.node.CanonicalBytes()
				if err != nil {
					t.Fatal(err)
				}
				mutated := mutateCanonicalFieldType(t, canonical, "payload", field)
				if _, err := DecodeNode(mutated); err == nil {
					t.Fatalf("DecodeNode accepted wrong-typed optional field %s", field)
				}
			})
		}
	}

	from, to := testDigest('9'), testDigest('a')
	origin := fixtureOrigin("strict.optionals")
	condition := Condition{EvaluatorID: "fixture-evaluator-v1", Expression: "feature.strict"}
	edgeCases := []struct {
		edge  Edge
		paths [][]string
	}{
		{edge: Edge{Kind: EdgeDeclares, EdgeKey: "edge:strict-declares", FromNodeID: from, ToNodeID: to, Payload: DeclaresPayload{Origin: origin}}, paths: [][]string{{"payload", "origin", "manifest_digest"}}},
		{edge: Edge{Kind: EdgeRequires, EdgeKey: "edge:strict-requires", FromNodeID: from, ToNodeID: to, Payload: RequiresPayload{Scope: ScopeOptional, Condition: &condition, Origin: origin, DependencyKind: "package"}}, paths: [][]string{{"payload", "condition"}, {"payload", "dependency_kind"}}},
		{edge: Edge{Kind: EdgeReads, EdgeKey: "edge:strict-reads", FromNodeID: from, ToNodeID: to, Payload: ReadsPayload{Path: "src/main.go", ReadSlot: "source", ReadClass: "source", Projection: []string{"src/main.go"}}}, paths: [][]string{{"payload", "read_class"}, {"payload", "projection"}}},
		{edge: Edge{Kind: EdgeUsesTool, EdgeKey: "edge:strict-uses", FromNodeID: from, ToNodeID: to, Payload: UsesToolPayload{ExecutableRelativePath: "bin/compiler", ToolSlot: "compiler", InvocationRole: "compile"}}, paths: [][]string{{"payload", "invocation_role"}}},
		{edge: Edge{Kind: EdgeProduces, EdgeKey: "edge:strict-produces", FromNodeID: from, ToNodeID: to, Payload: ProducesPayload{Path: "bin/output", WriteSlot: "output", WriteClass: "native.executable"}}, paths: [][]string{{"payload", "write_class"}}},
		{edge: Edge{Kind: EdgeInvokes, EdgeKey: "edge:strict-invokes", FromNodeID: from, ToNodeID: to, Payload: InvokesPayload{ProtocolSchema: "json-v1", ExecutableResolution: "bundle-relative-v1", ArgumentsContract: "argv-v1", EnvironmentContract: "env-v1", WorkingDirectory: "work"}}, paths: [][]string{{"payload", "working_directory"}}},
	}
	for _, testCase := range edgeCases {
		for _, path := range testCase.paths {
			t.Run("edge/"+string(testCase.edge.Kind)+"/"+strings.Join(path[1:], "."), func(t *testing.T) {
				canonical, err := testCase.edge.CanonicalBytes()
				if err != nil {
					t.Fatal(err)
				}
				mutated := mutateCanonicalFieldType(t, canonical, path...)
				if _, err := DecodeEdge(mutated); err == nil {
					t.Fatalf("DecodeEdge accepted wrong-typed optional field %s", strings.Join(path, "."))
				}
			})
		}
	}
}

func mutateCanonicalFieldType(t *testing.T, canonical []byte, path ...string) []byte {
	t.Helper()
	raw, err := decodeCanonicalObject(canonical, "test record")
	if err != nil {
		t.Fatal(err)
	}
	cursor := raw
	for _, field := range path[:len(path)-1] {
		next, ok := cursor[field].(map[string]any)
		if !ok {
			t.Fatalf("mutation path %s is not an object", strings.Join(path, "."))
		}
		cursor = next
	}
	leaf := path[len(path)-1]
	if _, exists := cursor[leaf]; !exists {
		t.Fatalf("mutation field %s is absent", strings.Join(path, "."))
	}
	cursor[leaf] = int64(1)
	mutated, err := canonicalMapBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	return mutated
}

func TestC0ThroughC7CheckpointCodecsAndChain(t *testing.T) {
	platformID, selectionID := testDigest('1'), testDigest('2')
	payloads := []CheckpointPayload{
		C0ProfilePayload{AdapterProfileID: "fixture-source-v1", SchemaIDs: []string{"closure-v1"}, ArtifactPolicyID: "artifact-policy-v1", DetectorRegistryID: "detectors-v1", SourceGrammarIDs: []string{"source-v1"}, LimitVectorID: "limits-v1", SelectionContextID: selectionID, PlatformNodeIDs: []ID{platformID}, PlatformRoles: map[PlatformRole]ID{PlatformTarget: platformID}, ManagerSchemaIDs: []string{"manager-v1"}, ConfigurationPolicyID: "configuration-v1", CapabilityIDs: []string{}, EvidenceToolchainNodeIDs: []ID{}},
		C1ResolvePayload{RootDeclarationIDs: []ID{testDigest('3')}, WorkspaceDeclarationIDs: []ID{}, LockCandidateID: testDigest('4'), ConditionEdgeIDs: []ID{}, ParserEvaluatorIDs: []string{"parser-v1"}, CandidateNodeIDs: []ID{testDigest('5')}, CandidateEdgeIDs: []ID{}, SelectionContextID: selectionID, JournalEntryIDs: []ID{testDigest('6')}},
		C2CapturePayload{IntakeReceiptIDs: []ID{testDigest('7')}, OriginIDs: []ID{testDigest('8')}, ProtectedHandleIDs: []ID{testDigest('9')}, BrokerReceiptIDs: []ID{}},
		C3AdmitPayload{Phase: "main", IntakeReceiptIDs: []ID{testDigest('7')}, ArtifactManifestIDs: []ID{testDigest('a')}, DerivationReceiptIDs: []ID{}},
		C4ClosePayload{ActiveGraphID: testDigest('b'), CapturedGraphID: testDigest('c'), SelectionBindingID: testDigest('d'), SelectionContextID: selectionID},
		C5PlanPayload{BuildPlanID: testDigest('e')},
		C6OfflinePayload{ExecutionReceiptID: testDigest('f')},
		C7PublishPayload{PublicationReceiptID: testDigest('0')},
	}
	chain := make([]Checkpoint, 0, len(payloads))
	var previous *Checkpoint
	for _, payload := range payloads {
		payload.checkpointPayload()
		checkpoint, err := NewCheckpoint(payload, previous, []Diagnostic{})
		if err != nil {
			t.Fatal(err)
		}
		canonical, err := checkpoint.CanonicalBytes()
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := DecodeCheckpoint(canonical)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(decoded, checkpoint) {
			t.Fatalf("%s round trip mismatch", checkpoint.Name)
		}
		chain = append(chain, checkpoint)
		previous = &chain[len(chain)-1]
	}
	if err := validateCheckpointSequence(chain); err != nil {
		t.Fatal(err)
	}
	broken := append([]Checkpoint{}, chain...)
	stale := testDigest('9')
	broken[5].PreviousCheckpointID = &stale
	if err := validateCheckpointSequence(broken); err == nil || !bytes.Contains([]byte(err.Error()), []byte(string(CodeCheckpointInvalid))) {
		t.Fatalf("broken chain error = %v", err)
	}
}

func TestCargoC3IntermediateCheckpointChain(t *testing.T) {
	base := checkpointPayloadFixtures()
	payloads := []CheckpointPayload{base[0], base[1], base[2], C3AdmitPayload{Phase: "origin", IntakeReceiptIDs: []ID{testDigest('1')}, ArtifactManifestIDs: []ID{testDigest('2')}, DerivationReceiptIDs: []ID{}}, C3AdmitPayload{Phase: "derived", IntakeReceiptIDs: []ID{testDigest('1')}, ArtifactManifestIDs: []ID{testDigest('3')}, DerivationReceiptIDs: []ID{testDigest('4')}}, base[3], base[4], base[5], base[6], base[7]}
	chain := []Checkpoint{}
	var previous *Checkpoint
	for _, payload := range payloads {
		checkpoint, err := NewCheckpoint(payload, previous, []Diagnostic{})
		if err != nil {
			t.Fatal(err)
		}
		chain = append(chain, checkpoint)
		previous = &chain[len(chain)-1]
	}
	if err := validateCheckpointSequence(chain); err != nil {
		t.Fatal(err)
	}
}

func checkpointPayloadFixtures() []CheckpointPayload {
	platformID, selectionID := testDigest('1'), testDigest('2')
	return []CheckpointPayload{C0ProfilePayload{AdapterProfileID: "fixture-v1", SchemaIDs: []string{"schema-v1"}, ArtifactPolicyID: "policy-v1", DetectorRegistryID: "detector-v1", SourceGrammarIDs: []string{"grammar-v1"}, LimitVectorID: "limits-v1", SelectionContextID: selectionID, PlatformNodeIDs: []ID{platformID}, PlatformRoles: map[PlatformRole]ID{PlatformTarget: platformID}, ManagerSchemaIDs: []string{}, ConfigurationPolicyID: "config-v1", CapabilityIDs: []string{}, EvidenceToolchainNodeIDs: []ID{}}, C1ResolvePayload{RootDeclarationIDs: []ID{}, WorkspaceDeclarationIDs: []ID{}, LockCandidateID: testDigest('3'), ConditionEdgeIDs: []ID{}, ParserEvaluatorIDs: []string{}, CandidateNodeIDs: []ID{}, CandidateEdgeIDs: []ID{}, SelectionContextID: selectionID, JournalEntryIDs: []ID{}}, C2CapturePayload{IntakeReceiptIDs: []ID{}, OriginIDs: []ID{}, ProtectedHandleIDs: []ID{}, BrokerReceiptIDs: []ID{}}, C3AdmitPayload{Phase: "main", IntakeReceiptIDs: []ID{}, ArtifactManifestIDs: []ID{}, DerivationReceiptIDs: []ID{}}, C4ClosePayload{ActiveGraphID: testDigest('4'), CapturedGraphID: testDigest('5'), SelectionBindingID: testDigest('6'), SelectionContextID: selectionID}, C5PlanPayload{BuildPlanID: testDigest('7')}, C6OfflinePayload{ExecutionReceiptID: testDigest('8')}, C7PublishPayload{PublicationReceiptID: testDigest('9')}}
}

type adapterStub struct{}

func (adapterStub) Capture(context.Context) (CaptureGraph, []Node, []Edge, error) {
	return CaptureGraph{}, nil, nil, nil
}
func (adapterStub) Bind(context.Context, CaptureGraph, SelectionContext) (SelectionBinding, []Node, []Edge, BindingAuthority, []ConditionEvaluator, error) {
	return SelectionBinding{}, nil, nil, BindingAuthority{}, nil, nil
}
func (adapterStub) Profile(context.Context) (CheckpointResult[C0ProfilePayload], error) {
	return CheckpointResult[C0ProfilePayload]{}, nil
}
func (adapterStub) Resolve(context.Context, Checkpoint) (CheckpointResult[C1ResolvePayload], error) {
	return CheckpointResult[C1ResolvePayload]{}, nil
}
func (adapterStub) CaptureInputs(context.Context, Checkpoint) (CheckpointResult[C2CapturePayload], error) {
	return CheckpointResult[C2CapturePayload]{}, nil
}
func (adapterStub) Admit(context.Context, Checkpoint) (CheckpointResult[C3AdmitPayload], error) {
	return CheckpointResult[C3AdmitPayload]{}, nil
}
func (adapterStub) Close(context.Context, Checkpoint) (GraphBundle, CheckpointResult[C4ClosePayload], error) {
	return GraphBundle{}, CheckpointResult[C4ClosePayload]{}, nil
}
func (adapterStub) Plan(context.Context, Checkpoint, GraphBundle) (BuildPlan, CheckpointResult[C5PlanPayload], error) {
	return BuildPlan{}, CheckpointResult[C5PlanPayload]{}, nil
}
func (adapterStub) ExecuteOffline(context.Context, Checkpoint, BuildPlan) (ExecutionReceipt, []ProducedArtifactObservation, CheckpointResult[C6OfflinePayload], error) {
	return ExecutionReceipt{}, nil, CheckpointResult[C6OfflinePayload]{}, nil
}
func (adapterStub) Publish(context.Context, Checkpoint, ExecutionReceipt) (PublicationReceipt, CheckpointResult[C7PublishPayload], error) {
	return PublicationReceipt{}, CheckpointResult[C7PublishPayload]{}, nil
}

var _ CaptureAdapter = adapterStub{}
var _ SelectionAdapter = adapterStub{}
var _ C0Profiler = adapterStub{}
var _ C1Resolver = adapterStub{}
var _ C2Capturer = adapterStub{}
var _ C3Admitter = adapterStub{}
var _ C4Closer = adapterStub{}
var _ C5Planner = adapterStub{}
var _ C6OfflineExecutor = adapterStub{}
var _ C7Publisher = adapterStub{}
