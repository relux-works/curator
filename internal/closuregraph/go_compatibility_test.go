package closuregraph

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
)

func TestGoV1CompatibilityGoldenMapsIntoCanonicalGraph(t *testing.T) {
	input := buildmeta.Input{SchemaVersion: buildmeta.SchemaVersion, Driver: buildmeta.DriverGoV1, BuildSource: buildsource.Identity{Algorithm: buildsource.Algorithm, ContentSHA256: "sha256:" + strings.Repeat("b", 64)}, BuildRoot: "build", Command: "golden-tool", SourceDir: "build/cmd/golden-tool", Target: buildmeta.Target{GOOS: "darwin", GOARCH: "arm64", Tuning: map[string]string{"GOARM64": "v8.0"}}, Toolchain: buildmeta.Toolchain{Algorithm: buildmeta.ToolchainAlgorithm, GoRelpath: buildmeta.ToolchainGoRelpath, GoVersion: "go version go1.26.1 darwin/arm64", ContentSHA256: "sha256:" + strings.Repeat("c", 64)}, Policy: buildmeta.FixedPolicy()}
	key, err := input.CacheKey()
	if err != nil {
		t.Fatal(err)
	}
	const wantLegacyCacheKey = "sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b"
	if string(key) != wantLegacyCacheKey {
		t.Fatalf("legacy cache key = %s, want %s", key, wantLegacyCacheKey)
	}
	artifact := buildmeta.Artifact{Path: "bin/golden-tool", SHA256: "sha256:" + strings.Repeat("d", 64), Size: 1234567}
	legacyReceipt, err := buildmeta.NewReceipt(input, artifact)
	if err != nil {
		t.Fatal(err)
	}
	legacyBytes, err := legacyReceipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	legacyHash, err := buildmeta.HashReceiptBytes(legacyBytes)
	if err != nil {
		t.Fatal(err)
	}
	const wantLegacyReceiptHash = "sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd"
	if string(legacyHash) != wantLegacyReceiptHash {
		t.Fatalf("legacy receipt hash = %s, want %s", legacyHash, wantLegacyReceiptHash)
	}

	manifestID := ID(input.BuildSource.ContentSHA256)
	product := Node{Kind: NodeCommandProduct, LogicalKey: "product:go-golden-tool", Payload: CommandProductPayload{Profile: "go-v1", SkillKey: "go-golden", CommandKey: input.Command, EntryPointContract: "native_command", DeclarationDigest: testDigest('a')}}
	source := Node{Kind: NodeSourceSet, LogicalKey: "source:go-golden-tool", Payload: SourceSetPayload{Profile: "go-v1", Origin: "curator-build-source-v1://golden", ArtifactManifestID: manifestID, Projection: []string{input.SourceDir}, Grammar: "go-source-v1", TrustRole: TrustDependencyInput}}
	action := Node{Kind: NodeAction, LogicalKey: "action:go-build-golden-tool", Payload: ActionPayload{Profile: "go-v1", ActionSubtype: "compiler", ExecutionDomain: ExecutionTarget, ArgvTemplate: []string{"$TOOL(go)", "build", "-o", "$WRITE(binary)", "$READ(source)"}, ToolSlotNames: []string{"go"}, ReadSlotNames: []string{"source"}, WriteSlotNames: []string{"binary"}, EnvironmentPolicyID: "go-v1-env", ProcessPolicyID: buildmeta.ExecutionPolicy, Network: "none"}}
	output := Node{Kind: NodeOutputArtifact, LogicalKey: "output:go-golden-tool", Payload: OutputArtifactPayload{Profile: "go-v1", LogicalPath: artifact.Path, ExpectedClass: "native.executable", OutputRole: "published_command"}}
	productID, sourceID, actionID, outputID := mustNodeID(t, product), mustNodeID(t, source), mustNodeID(t, action), mustNodeID(t, output)
	captureEdges := []Edge{{Kind: EdgeDeclares, EdgeKey: "edge:go-product-declares-build", FromNodeID: productID, ToNodeID: actionID, Payload: DeclaresPayload{Origin: fixtureOrigin("commands.golden-tool.build")}}, {Kind: EdgeReads, EdgeKey: "edge:go-build-reads-source", FromNodeID: actionID, ToNodeID: sourceID, Payload: ReadsPayload{Path: input.SourceDir, ReadSlot: "source"}}, {Kind: EdgeProduces, EdgeKey: "edge:go-build-produces-binary", FromNodeID: actionID, ToNodeID: outputID, Payload: ProducesPayload{Path: artifact.Path, WriteSlot: "binary"}}, {Kind: EdgePublishes, EdgeKey: "edge:go-product-publishes-binary", FromNodeID: productID, ToNodeID: outputID, Payload: PublishesPayload{Destination: artifact.Path, EntryPoint: input.Command}}}
	capture, err := NewCaptureGraph("go-v1", []string{"curator-artifact-policy-v1"}, []ID{productID}, []Node{product, source, action, output}, captureEdges, []ID{manifestID})
	if err != nil {
		t.Fatal(err)
	}
	platform := Node{Kind: NodeTargetPlatform, LogicalKey: "platform:go-darwin-arm64", Payload: TargetPlatformPayload{OS: input.Target.GOOS, Architecture: input.Target.GOARCH, ABI: "darwin", Libc: "libSystem", MinimumRuntime: "macos-native", SDKID: "go-darwin-sdk-v1", TargetTriple: "arm64-apple-darwin", LanguageModes: map[string]string{"go": "go1.26"}, Tuning: input.Target.Tuning}}
	toolchain := Node{Kind: NodeToolchainComponent, LogicalKey: "toolchain:go-v1", Payload: ToolchainComponentPayload{ComponentRole: "go_compiler", ContentFingerprint: ID(input.Toolchain.ContentSHA256), ExecutableRelativePath: input.Toolchain.GoRelpath, PlatformABI: input.Target.GOOS + "-" + input.Target.GOARCH, PolicySelector: input.Toolchain.Algorithm, VersionOutput: input.Toolchain.GoVersion, TimeOfUseRecheckRule: "exact-content-v1", ExecutionDomain: ExecutionTarget}}
	platformID, toolchainID := mustNodeID(t, platform), mustNodeID(t, toolchain)
	selection, err := NewSelectionContext([]ID{productID}, map[PlatformRole]ID{PlatformTarget: platformID}, []string{}, false, map[string]string{}, map[string]string{}, []string{})
	if err != nil {
		t.Fatal(err)
	}
	targets := func(key string, from ID) Edge {
		return Edge{Kind: EdgeTargets, EdgeKey: key, FromNodeID: from, ToNodeID: platformID, Payload: TargetsPayload{BindingRole: PlatformTarget, Origin: EvidenceOrigin{Field: "selection.platform_roles.target"}}}
	}
	bindingEdges := []Edge{{Kind: EdgeUsesTool, EdgeKey: "edge:go-build-uses-go", FromNodeID: actionID, ToNodeID: toolchainID, Payload: UsesToolPayload{ExecutableRelativePath: input.Toolchain.GoRelpath, ToolSlot: "go"}}, targets("edge:go-product-targets", productID), targets("edge:go-action-targets", actionID), targets("edge:go-output-targets", outputID), targets("edge:go-toolchain-targets", toolchainID)}
	captureID, _ := capture.ID()
	selectionID, _ := selection.ID()
	binding, err := NewSelectionBinding(captureID, selectionID, []Node{platform, toolchain}, bindingEdges)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := ProjectActive(capture, selection, binding, NewRecordTables([]Node{product, source, action, output}, captureEdges, []Node{platform, toolchain}, bindingEdges), c4AuthorityForNode(t, toolchain), nil)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: buildmeta.ExecutionPolicy})
	if err != nil {
		t.Fatal(err)
	}
	activeID, _ := bundle.Active.ID()
	bindingID, _ := binding.ID()
	planID, _ := plan.ID()
	goldens := map[string]struct {
		got  ID
		want ID
	}{
		"capture":   {captureID, "sha256:4ee6856acb64e411bd45ef3e49bdf4e03c13a247275701216e092ff5c5c100d4"},
		"selection": {selectionID, "sha256:a4362e2d38e508ab151b6ff9a3bb4ab45c3bdba4e5471eb43c81e8ea4f572572"},
		"binding":   {bindingID, "sha256:09ca88fa14f3b750ef388109a3927e12f9c81d2c3a549c2fea5c04462944bab1"},
		"active":    {activeID, "sha256:bd2a2969c322e16d8487094dd40cf655af65d17bd7f7b5a11c7d1b86e63e7971"},
		"plan":      {planID, "sha256:1927a1d0aa060a2813fb110f9a2dc6b2f1d9ab9f96f4e1e7be1ce90d215b01cb"},
	}
	for name, golden := range goldens {
		if golden.got != golden.want {
			t.Errorf("%s ID = %s, want %s", name, golden.got, golden.want)
		}
	}
}
