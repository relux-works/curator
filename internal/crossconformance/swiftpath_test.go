package crossconformance_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/crossconformance"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// swiftDestinations are the two exact SwiftPM destinations the cross suite
// projects the same captured Swift/C-family closure onto.
var swiftDestinations = []struct {
	label    string
	platform closuregraph.TargetPlatformPayload
	markers  map[string]string
}{
	{
		label: "darwin-arm64",
		platform: closuregraph.TargetPlatformPayload{
			OS: "darwin", Architecture: "arm64", ABI: "darwin", Libc: "libSystem", MinimumRuntime: "macos-14",
			SDKID: "macos-sdk-v1", TargetTriple: "arm64-apple-macosx14.0", Runtime: "swift-6",
			LanguageModes: map[string]string{"swift": "6"}, Tuning: map[string]string{},
		},
		markers: map[string]string{"platform": "macos", "configuration": "release", "architecture": "arm64"},
	},
	{
		label: "linux-x86_64",
		platform: closuregraph.TargetPlatformPayload{
			OS: "linux", Architecture: "x86_64", ABI: "gnu", Libc: "glibc", MinimumRuntime: "linux",
			SDKID: "swift-linux-sdk-v1", TargetTriple: "x86_64-unknown-linux-gnu", Runtime: "swift-6",
			LanguageModes: map[string]string{"swift": "6"}, Tuning: map[string]string{},
		},
		markers: map[string]string{"platform": "linux", "configuration": "release", "architecture": "x86_64"},
	},
}

type swiftFixture struct {
	root      string
	config    swiftpmsource.Config
	evaluator *swiftEvaluator
	broker    *swiftBroker
	platform  closuregraph.ID
}

func newSwiftFixture(t *testing.T, destinationIndex int, extra map[string][]byte) swiftFixture {
	t.Helper()
	base := t.TempDir()
	releaseTree(t, base)
	root := writeSwiftPackage(t, base, "root", extra)
	dependency := writeSwiftPackage(t, base, "dep", extra)
	writeFile(t, filepath.Join(root, "Package.resolved"), swiftLock())
	mirror := filepath.Join(base, "mirror-dep")
	if err := os.Mkdir(mirror, 0o700); err != nil {
		t.Fatal(err)
	}
	store, err := closureexec.NewCaptureStore(filepath.Join(base, "store"))
	if err != nil {
		t.Fatal(err)
	}
	evaluator := &swiftEvaluator{manifests: swiftManifests()}
	broker := &swiftBroker{snapshots: map[string]swiftpmsource.Snapshot{
		"dep": {Identity: "dep", Root: dependency, MirrorRoot: mirror, Revision: strings.Repeat("a", 40), GitTree: strings.Repeat("c", 40), Kind: swiftpmsource.SourceRemote, BrokerReceiptID: digestID([]byte("swift-broker-dep"))},
	}}
	destination := swiftDestinations[destinationIndex]
	markers := map[string]string{}
	for key, value := range destination.markers {
		markers[key] = value
	}
	tools := swiftpmsource.Toolchain{
		Swift:              swiftTool("swift", "Apple Swift version 6.3.2"),
		SwiftPM:            swiftTool("swiftpm", "Swift Package Manager 6.3.2"),
		PackageDescription: swiftTool("package-description", "PackageDescription 6.3.2"),
		Git:                swiftTool("git", "git version 2.50.1"),
	}
	tools.Recheck = func(_ context.Context, expected swiftpmsource.ToolIdentity) (closureexec.ToolchainIdentity, error) {
		return closureexec.ToolchainIdentity{Fingerprint: expected.Fingerprint, ExecutableSHA256: expected.ExecutableSHA256}, nil
	}
	config := swiftpmsource.Config{
		Store: store, Policy: artifactpolicy.NewService(), Evaluator: evaluator, Broker: broker,
		MirrorVerifier: swiftMirrorVerifier{}, OfflineRunner: swiftOfflineRunner{}, Toolchain: tools,
		Destination: swiftpmsource.Destination{Platform: destination.platform, Markers: markers},
		CausalHead:  "sha256:" + strings.Repeat("0", 64),
	}
	config.ProcessStartObserver = func(permit swiftpmsource.ManifestPermit) {
		evaluator.starts = append(evaluator.starts, permit.PackageIdentity)
	}
	platformNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "swiftpm.platform.target", Payload: destination.platform}
	platformID, err := platformNode.ID()
	if err != nil {
		t.Fatal(err)
	}
	return swiftFixture{root: root, config: config, evaluator: evaluator, broker: broker, platform: platformID}
}

func projectSwiftPath(t *testing.T, destinationIndex int) crossconformance.TargetProjection {
	t.Helper()
	fixture := newSwiftFixture(t, destinationIndex, nil)
	capture, err := swiftpmsource.CaptureAndClose(context.Background(), fixture.config, swiftpmsource.Request{Root: fixture.root, Product: "cli", Resolved: swiftLock()})
	if err != nil {
		t.Fatalf("SwiftPM capture: %v", err)
	}
	receipts := appendUnique(nil, string(capture.ResolutionReceiptID))
	for _, evidence := range capture.Packages {
		receipts = appendUnique(receipts, string(evidence.IntakeReceiptID), string(evidence.ManifestReceiptID))
	}
	checkpoints := []crossconformance.CheckpointLink{}
	for _, checkpoint := range []closuregraph.Checkpoint{capture.C0, capture.C1, capture.C2, capture.C3, capture.C4} {
		link := crossconformance.CheckpointLink{Name: string(checkpoint.Name), Identity: identityOf(t, checkpoint)}
		if checkpoint.PreviousCheckpointID != nil {
			link.Previous = string(*checkpoint.PreviousCheckpointID)
		}
		checkpoints = append(checkpoints, link)
	}
	tools := []string{}
	for _, tool := range []swiftpmsource.ToolIdentity{fixture.config.Toolchain.Swift, fixture.config.Toolchain.SwiftPM, fixture.config.Toolchain.PackageDescription, fixture.config.Toolchain.Git} {
		tools = append(tools, string(tool.Fingerprint))
	}
	return crossconformance.TargetProjection{
		Path:                   crossconformance.PathSwiftPM,
		TargetLabel:            swiftDestinations[destinationIndex].label,
		CaptureIdentity:        identityOf(t, capture.Graph),
		SelectionIdentity:      identityOf(t, capture.Selection),
		BindingIdentity:        identityOf(t, capture.Binding),
		ActiveIdentity:         identityOf(t, capture.Active),
		CaptureNodeKinds:       nodeKindCensus(t, capture.Records.CaptureNodes),
		CaptureEdgeKinds:       edgeKindCensus(t, capture.Records.CaptureEdges),
		BindingNodeKinds:       nodeKindCensus(t, capture.Records.BindingNodes),
		BindingEdgeKinds:       edgeKindCensus(t, capture.Records.BindingEdges),
		TargetPlatformIdentity: string(fixture.platform),
		ExplicitTargetEdges:    explicitTargetEdges(t, capture.Records.BindingEdges, fixture.platform),
		EmitsBindingRecords:    true,
		ToolIdentities:         tools,
		Checkpoints:            checkpoints,
		DerivationReceipts:     receipts,
	}
}

func appendUnique(values []string, additions ...string) []string {
	for _, addition := range additions {
		if addition == "" {
			continue
		}
		found := false
		for _, existing := range values {
			found = found || existing == addition
		}
		if !found {
			values = append(values, addition)
		}
	}
	return values
}

func writeSwiftPackage(t *testing.T, base, name string, extra map[string][]byte) string {
	t.Helper()
	root := filepath.Join(base, name)
	files := map[string][]byte{
		"Package.swift":                  []byte("// swift-tools-version: 6.1\nimport PackageDescription\n"),
		"Sources/App/main.swift":         []byte("public func value() -> Int { 1 }\n"),
		"Sources/CShim/shim.c":           []byte("int shim_value(void) { return 1; }\n"),
		"Sources/CShim/shim.h":           []byte("int shim_value(void);\n"),
		"Sources/CShim/module.modulemap": []byte("module CShim { header \"shim.h\" export * }\n"),
	}
	for relative, payload := range extra {
		files["Sources/App/"+relative] = payload
	}
	for relative, payload := range files {
		writeFile(t, filepath.Join(root, filepath.FromSlash(relative)), payload)
	}
	return root
}

func swiftLock() []byte {
	return []byte(`{"version":3,"pins":[{"identity":"dep","kind":"remoteSourceControl","location":"https://example.invalid/Dep","state":{"revision":"` + strings.Repeat("a", 40) + `","version":"1.0.0"}}]}`)
}

func swiftManifests() map[string]swiftpmsource.Manifest {
	return map[string]swiftpmsource.Manifest{
		"root": {
			PackageName: "root", ToolsVersion: "6.1",
			Products: []swiftpmsource.Product{{Name: "cli", Type: "executable", Targets: []string{"App"}}},
			Targets: []swiftpmsource.Target{
				{Name: "App", Type: "executable", Sources: []string{"Sources/App/main.swift"}, Dependencies: []swiftpmsource.TargetDependency{{Package: "dep", Product: "DepProd"}, {Name: "CShim"}}},
				{Name: "CShim", Type: "regular", Path: "Sources/CShim", PublicHeadersPath: ".", Sources: []string{"Sources/CShim/shim.c"}},
			},
			Dependencies: []swiftpmsource.ManifestDependency{{Identity: "dep", Kind: swiftpmsource.SourceRemote, Location: "https://example.invalid/Dep", Requirement: "exact:1.0.0"}},
		},
		"dep": {
			PackageName: "dep", ToolsVersion: "6.1",
			Products: []swiftpmsource.Product{{Name: "DepProd", Type: "library", Targets: []string{"DepTarget"}}},
			Targets:  []swiftpmsource.Target{{Name: "DepTarget", Type: "regular", Sources: []string{"Sources/App/main.swift"}}},
		},
	}
}

func swiftTool(role, version string) swiftpmsource.ToolIdentity {
	return swiftpmsource.ToolIdentity{
		Role: role, ExecutableRelativePath: "bin/" + role, VersionOutput: version,
		PlatformABI: "darwin-arm64", PolicySelector: "swift-toolchain-v1",
		Fingerprint: digestID([]byte("swift-tool-" + role)), ExecutableSHA256: digestID([]byte("swift-tool-executable-" + role)),
	}
}

type swiftEvaluator struct {
	manifests map[string]swiftpmsource.Manifest
	starts    []string
}

func (evaluator *swiftEvaluator) Evaluate(_ context.Context, root string, permit swiftpmsource.ManifestPermit) (swiftpmsource.ManifestResult, error) {
	if permit.Network != "none" {
		return swiftpmsource.ManifestResult{}, errors.New("manifest permit does not deny the network")
	}
	if !filepath.IsAbs(root) {
		return swiftpmsource.ManifestResult{}, errors.New("manifest root is not a protected absolute tree")
	}
	manifest, ok := evaluator.manifests[permit.PackageIdentity]
	if !ok {
		return swiftpmsource.ManifestResult{}, fmt.Errorf("no manifest for %s", permit.PackageIdentity)
	}
	receipt, err := closuregraph.DomainID("cross-conformance-swift-manifest-receipt-v1", map[string]any{
		"package": permit.PackageIdentity, "permit": string(permit.ID), "selected": permit.SelectedManifest,
	})
	if err != nil {
		return swiftpmsource.ManifestResult{}, err
	}
	return swiftpmsource.ManifestResult{Manifest: manifest, ReceiptID: receipt}, nil
}

type swiftBroker struct {
	snapshots map[string]swiftpmsource.Snapshot
	starts    []string
}

func (broker *swiftBroker) Acquire(_ context.Context, pin swiftpmsource.Pin) (swiftpmsource.Snapshot, error) {
	broker.starts = append(broker.starts, pin.Identity)
	snapshot, ok := broker.snapshots[pin.Identity]
	if !ok {
		return swiftpmsource.Snapshot{}, fmt.Errorf("no snapshot for %s", pin.Identity)
	}
	return snapshot, nil
}

type swiftMirrorVerifier struct{}

func (swiftMirrorVerifier) Verify(_ context.Context, _ string, pin swiftpmsource.Pin, snapshot swiftpmsource.Snapshot) (swiftpmsource.GitVerificationEvidence, error) {
	if snapshot.Identity != pin.Identity || snapshot.Kind != pin.Kind || snapshot.Revision != pin.Revision || snapshot.GitTree == "" {
		return swiftpmsource.GitVerificationEvidence{}, errors.New("mirror identity does not match the frozen pin")
	}
	return swiftpmsource.GitVerificationEvidence{}, nil
}

type swiftOfflineRunner struct{}

func (swiftOfflineRunner) Replay(_ context.Context, capture *swiftpmsource.Capture) (swiftpmsource.OfflineMetadataResult, error) {
	identities := make([]string, 0, len(capture.Lock.Pins))
	for _, pin := range capture.Lock.Pins {
		identities = append(identities, pin.Identity)
	}
	sort.Strings(identities)
	values := make([]any, len(identities))
	for index, identity := range identities {
		values[index] = identity
	}
	receipt, err := closuregraph.DomainID("cross-conformance-swift-offline-receipt-v1", map[string]any{"identities": values})
	if err != nil {
		return swiftpmsource.OfflineMetadataResult{}, err
	}
	return swiftpmsource.OfflineMetadataResult{ReceiptID: receipt, PackageIdentities: identities}, nil
}
