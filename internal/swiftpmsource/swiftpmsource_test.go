package swiftpmsource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
)

func TestParseResolvedR01R03R07R09(t *testing.T) {
	lock, err := ParseResolved(lockFixture())
	if err != nil {
		t.Fatal(err)
	}
	if lock.Schema != 3 || len(lock.Pins) != 2 || lock.Pins[0].Identity != "a" || lock.Pins[1].Identity != "b" {
		t.Fatalf("unexpected lock: %+v", lock)
	}
	if lock.Pins[0].RawLocation != "https://example.invalid/A" || lock.Pins[0].CanonicalLocation != "https://example.invalid/A" || lock.Pins[0].Revision != revision('a') {
		t.Fatalf("pin evidence lost: %+v", lock.Pins[0])
	}
	dependencyLock := []byte(`{"version":3,"pins":[]}`)
	if _, err = ParseResolved(dependencyLock); err != nil {
		t.Fatalf("dependency lock text must remain independently parseable: %v", err)
	}
}

func TestParseResolvedRejectsR04R08R11R13(t *testing.T) {
	tests := []struct {
		name, payload string
		code          Code
	}{
		{"mutable revision", `{"version":3,"pins":[{"identity":"a","kind":"remoteSourceControl","location":"https://example.invalid/A","state":{"branch":"main"}}]}`, CodeResolutionUnfrozen},
		{"registry", `{"version":3,"pins":[{"identity":"a","kind":"registry","location":"https://registry.invalid/a","state":{"revision":"` + revision('a') + `"}}]}`, CodeResolutionUnfrozen},
		{"duplicate", `{"version":3,"pins":[{"identity":"a","kind":"remoteSourceControl","location":"https://example.invalid/A","state":{"revision":"` + revision('a') + `"}},{"identity":"A","kind":"remoteSourceControl","location":"https://example.invalid/A","state":{"revision":"` + revision('a') + `"}}]}`, CodeDependencyPinMismatch},
		{"unknown field", `{"version":3,"pins":[],"objectVersion":1}`, CodeResolutionUnfrozen},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) { _, err := ParseResolved([]byte(test.payload)); requireCode(t, err, test.code) })
	}
}

func TestSelectManifestR12(t *testing.T) {
	files := []string{"Package.swift", "Package@swift-5.swift", "Package@swift-5.10.swift", "Package@swift-6.2.swift", "Package@swift-7.swift"}
	selected, err := SelectManifest(files, "Apple Swift version 6.3.2")
	if err != nil {
		t.Fatal(err)
	}
	if selected != "Package@swift-6.2.swift" {
		t.Fatalf("selected %q", selected)
	}
	selected, err = SelectManifest([]string{"Package.swift", "Package@swift-7.swift"}, "Swift version 6.3")
	if err != nil || selected != "Package.swift" {
		t.Fatalf("fallback = %q, %v", selected, err)
	}
}

func TestDecodeDumpPackageExactLocalSourceControl(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "Sources", "Root"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "Sources", "Root", "main.swift"), []byte("print(1)\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"dependencies":[{"sourceControl":[{"identity":"dep","location":{"local":["/tmp/dep"]},"productFilter":null,"requirement":{"exact":["1.0.0"]},"traits":[]}]}],"name":"Root","products":[{"name":"root","targets":["Root"],"type":{"executable":null}}],"targets":[{"dependencies":[{"product":["Dep","dep",null,null]}],"exclude":[],"name":"Root","settings":[],"type":"executable"}],"toolsVersion":{"_version":"5.9.0"},"traits":[]}`)
	manifest, err := decodeDumpPackage(payload, root)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Dependencies) != 1 || manifest.Dependencies[0].Kind != SourceLocal || manifest.Dependencies[0].Location != "/tmp/dep" || manifest.Dependencies[0].Requirement != "exact:1.0.0" {
		t.Fatalf("dependency = %+v", manifest.Dependencies)
	}
	if len(manifest.Targets) != 1 || len(manifest.Targets[0].Dependencies) != 1 || manifest.Targets[0].Dependencies[0].Product != "Dep" || manifest.Targets[0].Dependencies[0].Package != "dep" || !reflect.DeepEqual(manifest.Targets[0].Sources, []string{"Sources/Root/main.swift"}) {
		t.Fatalf("target = %+v", manifest.Targets)
	}
}

func TestCaptureAndOfflineReplayR01R05R06CGP05CGP11(t *testing.T) {
	fixture := newFixture(t)
	capture, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := fixture.evaluator.starts, []string{"root", "a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("manifest starts = %v, want %v", got, want)
	}
	if len(capture.Packages) != 3 || len(capture.Mirrors) != 2 {
		t.Fatalf("closure sizes = packages %d mirrors %d", len(capture.Packages), len(capture.Mirrors))
	}
	for _, node := range capture.Records.CaptureNodes {
		if node.Kind == closuregraph.NodeTargetPlatform || node.Kind == closuregraph.NodeToolchainComponent {
			t.Fatalf("selection-specific node leaked into capture: %s", node.Kind)
		}
	}
	for _, mirror := range capture.Mirrors {
		if mirror.OriginalKind != mirror.LocalKind {
			t.Fatalf("kind-changing mirror: %+v", mirror)
		}
	}
	if err = capture.ReplayOffline(t.Context()); err != nil {
		t.Fatal(err)
	}
	if got, want := fixture.evaluator.starts, []string{"root", "a", "b", "root", "a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("offline manifest replay order = %v, want %v", got, want)
	}
	if got, want := fixture.broker.starts, []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("offline replay reacquired origins: %v, want initial acquisition %v", got, want)
	}

	linux := newFixture(t)
	linux.config.Destination.Platform.OS = "linux"
	linux.config.Destination.Platform.ABI = "gnu"
	linux.config.Destination.Platform.Libc = "glibc"
	linux.config.Destination.Platform.SDKID = "swift-linux-sdk-v1"
	linux.config.Destination.Platform.TargetTriple = "x86_64-unknown-linux-gnu"
	linux.config.Destination.Markers["platform"] = "linux"
	other, err := CaptureAndClose(t.Context(), linux.config, Request{Root: linux.root, Product: "cli", Resolved: lockFixture()})
	if err != nil {
		t.Fatal(err)
	}
	if capture.GraphDigest != other.GraphDigest {
		t.Fatalf("destination changed selection-neutral capture: %s != %s", capture.GraphDigest, other.GraphDigest)
	}
	firstBinding, _ := capture.Binding.ID()
	secondBinding, _ := other.Binding.ID()
	if firstBinding == secondBinding {
		t.Fatal("destination did not change exact selection binding")
	}
}

func TestMirrorAdmissionRejectsCheckoutManifestSubstitution(t *testing.T) {
	fixture := newFixture(t)
	capture, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
	if err != nil {
		t.Fatal(err)
	}
	mirror := capture.Mirrors[0]
	pin := capture.Lock.Pins[0]
	protected, err := mirror.input.Tree.ProtectedPath()
	if err != nil {
		t.Fatal(err)
	}
	nodes, digest, err := inventoryMirrorTree(protected)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := Snapshot{Identity: mirror.Identity, MirrorRoot: protected, Revision: mirror.Revision, GitTree: mirror.GitTree, Kind: mirror.LocalKind, BrokerReceiptID: mirror.BrokerReceiptID}
	checkoutManifest := capture.Packages[1].ArtifactManifestID
	_, err = admitMirrorTree(fixture.config, mirror.input.Tree, pin, snapshot, GitVerificationEvidence{}, nodes, digest, checkoutManifest)
	requireCode(t, err, CodeDerivationUnauthorized)
	if checkoutManifest == mirror.ArtifactManifestID {
		t.Fatal("mirror and checkout unexpectedly share artifact evidence")
	}
}

func TestNoAffectedManifestOnIntakeToolPinAndMirrorRejectionCGN16CGN18(t *testing.T) {
	t.Run("root compiled payload", func(t *testing.T) {
		fixture := newFixture(t)
		if err := os.WriteFile(filepath.Join(fixture.root, "payload.wasm"), []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}, 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
		if artifactpolicy.ErrorCode(err) != artifactpolicy.CodeCompiledDependency {
			t.Fatalf("error = %v", err)
		}
		if len(fixture.evaluator.starts) != 0 {
			t.Fatalf("manifest started after intake rejection: %v", fixture.evaluator.starts)
		}
	})
	t.Run("C0 tool drift", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.config.Toolchain.Recheck = func(context.Context, ToolIdentity) (closureexec.ToolchainIdentity, error) {
			return closureexec.ToolchainIdentity{Fingerprint: id('f'), ExecutableSHA256: id('e')}, nil
		}
		_, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
		requireCode(t, err, CodeToolchainChanged)
		if len(fixture.evaluator.starts) != 0 {
			t.Fatalf("manifest started after tool drift: %v", fixture.evaluator.starts)
		}
	})
	t.Run("root origin drift before dependency", func(t *testing.T) {
		fixture := newFixture(t)
		root := fixture.evaluator.manifests["root"]
		root.Dependencies[0].Location = "https://evil.invalid/A"
		fixture.evaluator.manifests["root"] = root
		_, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
		requireCode(t, err, CodeDependencyPinMismatch)
		if got := fixture.evaluator.starts; !reflect.DeepEqual(got, []string{"root"}) {
			t.Fatalf("affected dependency manifest started: %v", got)
		}
	})
	t.Run("R04 stale exact requirement before dependency", func(t *testing.T) {
		fixture := newFixture(t)
		root := fixture.evaluator.manifests["root"]
		root.Dependencies[0].Requirement = "exact:2.0.0"
		fixture.evaluator.manifests["root"] = root
		_, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
		requireCode(t, err, CodeResolvedFileOutOfDate)
		if got := fixture.evaluator.starts; !reflect.DeepEqual(got, []string{"root"}) {
			t.Fatalf("stale dependency manifest started: %v", got)
		}
		if len(fixture.broker.starts) != 0 {
			t.Fatalf("stale dependency reached acquisition: %v", fixture.broker.starts)
		}
	})
	t.Run("R08 acquired revision mismatch before dependency", func(t *testing.T) {
		fixture := newFixture(t)
		snapshot := fixture.broker.snapshots["a"]
		snapshot.Revision = revision('f')
		fixture.broker.snapshots["a"] = snapshot
		_, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
		requireCode(t, err, CodeDependencyPinMismatch)
		if got := fixture.evaluator.starts; !reflect.DeepEqual(got, []string{"root"}) {
			t.Fatalf("pin-mismatched dependency manifest started: %v", got)
		}
	})
	t.Run("R13 submodule metadata before dependency", func(t *testing.T) {
		fixture := newFixture(t)
		if err := os.WriteFile(filepath.Join(fixture.broker.snapshots["a"].Root, ".gitmodules"), []byte("[submodule \"x\"]\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
		requireCode(t, err, CodeDependencyOriginUnsupported)
		if got := fixture.evaluator.starts; !reflect.DeepEqual(got, []string{"root"}) {
			t.Fatalf("submodule dependency manifest started: %v", got)
		}
	})
	t.Run("P07 compiled transitive payload before dependency", func(t *testing.T) {
		fixture := newFixture(t)
		if err := os.WriteFile(filepath.Join(fixture.broker.snapshots["a"].Root, "payload.wasm"), []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}, 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
		if artifactpolicy.ErrorCode(err) != artifactpolicy.CodeCompiledDependency {
			t.Fatalf("error = %v", err)
		}
		if got := fixture.evaluator.starts; !reflect.DeepEqual(got, []string{"root"}) {
			t.Fatalf("compiled dependency manifest started: %v", got)
		}
	})
	t.Run("transitive mirror missing", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.broker.failIdentity = "b"
		_, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
		requireCode(t, err, CodeDependencyMirrorMissing)
		if got := fixture.evaluator.starts; !reflect.DeepEqual(got, []string{"root", "a"}) {
			t.Fatalf("missing dependency manifest started: %v", got)
		}
	})
}

func TestExtensionReachabilityAndBinaryPolicyP01P08(t *testing.T) {
	t.Run("dormant plugin", func(t *testing.T) {
		fixture := newFixture(t)
		root := fixture.evaluator.manifests["root"]
		root.Targets = append(root.Targets, Target{Name: "Dormant", Type: "plugin", Sources: []string{}})
		fixture.evaluator.manifests["root"] = root
		if _, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()}); err != nil {
			t.Fatal(err)
		}
	})
	for _, kind := range []string{"plugin", "macro"} {
		t.Run("active "+kind, func(t *testing.T) {
			fixture := newFixture(t)
			root := fixture.evaluator.manifests["root"]
			root.Targets = append(root.Targets, Target{Name: "Extension", Type: kind, Sources: []string{"Sources/App/main.swift"}})
			root.Products[0].Targets = []string{"Extension"}
			fixture.evaluator.manifests["root"] = root
			_, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
			if kind == "plugin" {
				requireCode(t, err, CodePluginUnsupported)
			} else {
				requireCode(t, err, CodeMacroUnsupported)
			}
		})
	}
	t.Run("binary dominates plugin", func(t *testing.T) {
		fixture := newFixture(t)
		root := fixture.evaluator.manifests["root"]
		root.Targets = []Target{{Name: "Plugin", Type: "plugin"}, {Name: "Binary", Type: "binary"}}
		root.Products[0].Targets = []string{"Plugin"}
		fixture.evaluator.manifests["root"] = root
		_, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
		requireCode(t, err, CodeBinaryUnavailable)
		if got := fixture.evaluator.starts; !reflect.DeepEqual(got, []string{"root"}) {
			t.Fatalf("unexpected starts: %v", got)
		}
	})
}

func TestReplayDetectsManifestSourceAndGraphDrift(t *testing.T) {
	fixture := newFixture(t)
	capture, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
	if err != nil {
		t.Fatal(err)
	}
	root := fixture.evaluator.manifests["root"]
	root.ToolsVersion = "6.2"
	fixture.evaluator.manifests["root"] = root
	requireCode(t, capture.ReplayOffline(t.Context()), CodeManifestReplayDrift)
	root = fixture.evaluator.manifests["root"]
	root.ToolsVersion = "6.1"
	fixture.evaluator.manifests["root"] = root
	capture.GraphDigest = id('9')
	requireCode(t, capture.ReplayOffline(t.Context()), CodeBuildGraphDrift)
}

func TestLocalPathContainmentR10(t *testing.T) {
	manifest := fixtureManifests()["root"]
	manifest.SelectedManifest = "Package.swift"
	manifest.Dependencies = []ManifestDependency{{Identity: "local", Kind: SourcePath, LocalPath: "../escape"}}
	_, err := normalizeManifest(manifest)
	requireCode(t, err, CodeLocalDependencyOutside)
	manifest.Dependencies[0].LocalPath = "Packages/Local"
	if _, err = normalizeManifest(manifest); err != nil {
		t.Fatal(err)
	}
}

func TestInRootLocalPackageIsAdmittedBeforeItsManifestR10(t *testing.T) {
	fixture := newFixture(t)
	local := writePackage(t, filepath.Join(fixture.root, "Packages"), "Local")
	_ = local
	root := fixture.evaluator.manifests["root"]
	root.Dependencies = append(root.Dependencies, ManifestDependency{Identity: "local", Kind: SourcePath, LocalPath: "Packages/Local"})
	root.Targets[0].Dependencies = append(root.Targets[0].Dependencies, TargetDependency{Package: "local", Product: "LocalProd"})
	fixture.evaluator.manifests["root"] = root
	fixture.evaluator.manifests["local"] = Manifest{PackageName: "local", ToolsVersion: "6.1", Products: []Product{{Name: "LocalProd", Type: "library", Targets: []string{"LocalTarget"}}}, Targets: []Target{{Name: "LocalTarget", Type: "regular", Sources: []string{"Sources/App/main.swift"}}}}
	capture, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := fixture.evaluator.starts, []string{"root", "local", "a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("manifest order = %v, want %v", got, want)
	}
	found := false
	for _, pkg := range capture.Packages {
		if pkg.Identity == "local" {
			found = true
			if pkg.Kind != SourcePath || pkg.evaluationSubpath != "Packages/Local" {
				t.Fatalf("local evidence = %+v", pkg)
			}
		}
	}
	if !found {
		t.Fatal("local package absent from capture")
	}
}

func TestGeneratedRootLockIsFrozenBeforeDependenciesR02(t *testing.T) {
	fixture := newFixture(t)
	if err := os.Remove(filepath.Join(fixture.root, "Package.resolved")); err != nil {
		t.Fatal(err)
	}
	resolver := &fakeResolver{payload: lockFixture()}
	fixture.config.Resolver = resolver
	capture, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli"})
	if err != nil {
		t.Fatal(err)
	}
	if resolver.calls != 1 || capture.Lock.Digest == "" {
		t.Fatalf("generated lock was not frozen: calls=%d lock=%+v", resolver.calls, capture.Lock)
	}
	if !capture.ResolutionPermitID.Valid() || !capture.ResolutionReceiptID.Valid() {
		t.Fatalf("generated resolution journal is incomplete: permit=%s receipt=%s", capture.ResolutionPermitID, capture.ResolutionReceiptID)
	}
	c1 := capture.C1.Payload.(closuregraph.C1ResolvePayload)
	c3 := capture.C3.Payload.(closuregraph.C3AdmitPayload)
	if !containsID(c1.JournalEntryIDs, capture.ResolutionPermitID) || !containsID(c1.JournalEntryIDs, capture.ResolutionReceiptID) || !containsID(c3.DerivationReceiptIDs, capture.ResolutionReceiptID) {
		t.Fatalf("resolution evidence is absent from checkpoints: C1=%v C3=%v", c1.JournalEntryIDs, c3.DerivationReceiptIDs)
	}
	if got := fixture.evaluator.starts; !reflect.DeepEqual(got, []string{"root", "a", "b"}) {
		t.Fatalf("manifest order = %v", got)
	}
}

func TestDanglingLockPinStartsNoAffectedBrokerOrManifestR03R04(t *testing.T) {
	fixture := newFixture(t)
	fixture.broker.snapshots["c"] = fixture.broker.snapshots["b"]
	lock := []byte(`{"version":3,"pins":[{"identity":"a","kind":"remoteSourceControl","location":"https://example.invalid/A","state":{"revision":"` + revision('a') + `","version":"1.0.0"}},{"identity":"b","kind":"remoteSourceControl","location":"https://example.invalid/B","state":{"revision":"` + revision('b') + `","version":"1.0.0"}},{"identity":"c","kind":"remoteSourceControl","location":"https://example.invalid/C","state":{"revision":"` + revision('c') + `","version":"1.0.0"}}]}`)
	if err := os.WriteFile(filepath.Join(fixture.root, "Package.resolved"), lock, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lock})
	requireCode(t, err, CodeResolvedFileOutOfDate)
	if got, want := fixture.evaluator.starts, []string{"root", "a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("manifest starts = %v, want %v", got, want)
	}
	if contains(fixture.broker.starts, "c") {
		t.Fatalf("dangling pin reached broker: %v", fixture.broker.starts)
	}
}

func TestReplayRejectsCapturedMirrorByteDriftR08(t *testing.T) {
	fixture := newFixture(t)
	capture, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
	if err != nil {
		t.Fatal(err)
	}
	mirror := capture.Mirrors[0]
	if err = os.Chmod(mirror.Local, 0o700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(mirror.Local, "drift"), []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	requireCode(t, capture.ReplayOffline(t.Context()), CodeDependencyPinMismatch)
}

func TestSnapshotMirrorAndConditionFailuresR08P02(t *testing.T) {
	fixture := newFixture(t)
	pin := Pin{Identity: "a", Kind: SourceRemote, Revision: revision('a')}
	snapshot := fixture.broker.snapshots["a"]
	snapshot.Kind = SourceLocal
	requireCode(t, validateSnapshot(pin, snapshot), CodeDependencyPinMismatch)
	snapshot = fixture.broker.snapshots["a"]
	snapshot.MirrorRoot = filepath.Join(t.TempDir(), "missing")
	requireCode(t, validateSnapshot(pin, snapshot), CodeDependencyMirrorMissing)
	snapshot = fixture.broker.snapshots["a"]
	if err := os.WriteFile(filepath.Join(snapshot.Root, ".gitmodules"), []byte("[submodule \"x\"]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	requireCode(t, validateSnapshot(pin, snapshot), CodeDependencyOriginUnsupported)
	if err := os.Remove(filepath.Join(snapshot.Root, ".gitmodules")); err != nil {
		t.Fatal(err)
	}
	manifest := fixture.evaluator.manifests["root"]
	manifest.Targets[0].Settings = []BuildSetting{{Kind: "swiftSettings.unsafeFlags", Value: "-I/tmp", Unsafe: true, Condition: &closuregraph.Condition{EvaluatorID: ConditionEvaluatorID, Expression: "platform=macos"}}}
	fixture.evaluator.manifests["root"] = manifest
	_, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
	requireCode(t, err, CodeUnsafeSettingForbidden)
}

func TestOfflineReplayRejectsMissingMirrorInventoryAndNetworkFailure(t *testing.T) {
	fixture := newFixture(t)
	capture, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
	if err != nil {
		t.Fatal(err)
	}
	capture.Mirrors[0].LocalKind = SourceLocal
	requireCode(t, capture.ReplayOffline(t.Context()), CodeDependencyPinMismatch)
	capture.Mirrors[0].LocalKind = SourceRemote
	capture.InventoryDigest = id('8')
	requireCode(t, capture.ReplayOffline(t.Context()), CodeSourceInventoryDrift)
}

func TestClosedValidationBranches(t *testing.T) {
	if ErrorCode(errors.New("plain")) != "" {
		t.Fatal("plain error acquired an adapter code")
	}
	if _, err := SelectManifest([]string{"README"}, "Swift version 6.3"); ErrorCode(err) != CodeManifestReplayDrift {
		t.Fatalf("missing manifest error = %v", err)
	}
	if _, err := SelectManifest([]string{"Package.swift"}, "unknown tool"); ErrorCode(err) != CodeTargetPlatformUnsupported {
		t.Fatalf("unknown tool error = %v", err)
	}
	if selected, err := evaluateCondition("trait=Logging", map[string]string{"trait:Logging": "true"}); err != nil || !selected {
		t.Fatalf("trait condition = %t, %v", selected, err)
	}
	if _, err := evaluateCondition("sdk=macos", map[string]string{}); ErrorCode(err) != CodeGraphReferenceInvalid {
		t.Fatalf("unknown condition error = %v", err)
	}
	fixture := newFixture(t)
	fixture.config.Destination.Markers = nil
	_, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
	requireCode(t, err, CodeTargetPlatformUnsupported)
	fixture = newFixture(t)
	fixture.config.Toolchain.Swift.Fingerprint = ""
	_, err = CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli", Resolved: lockFixture()})
	requireCode(t, err, CodeTargetPlatformUnsupported)
	fixture = newFixture(t)
	_, err = CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "missing", Resolved: lockFixture()})
	requireCode(t, err, CodeGraphIncomplete)
}

func TestClosedRequirementDecoderAndLanguageBranches(t *testing.T) {
	pin := Pin{Version: "1.5.0", Revision: revision('a'), Branch: "stable"}
	for requirement, want := range map[string]bool{
		"exact:1.5.0": true, "exact:2.0.0": false,
		"revision:" + revision('a'): true, "revision:" + revision('b'): false,
		"branch:stable": true, "branch:other": false,
		"range:1.0.0..<2.0.0": true, "range:1.6.0..<2.0.0": false,
		"range:broken": false, "unknown:value": false, "missing": false,
	} {
		if got := requirementMatchesPin(requirement, pin); got != want {
			t.Fatalf("requirement %q = %t, want %t", requirement, got, want)
		}
	}
	for _, test := range []struct {
		raw  any
		want string
	}{
		{map[string]any{"exact": []any{"1.0.0"}}, "exact:1.0.0"},
		{map[string]any{"revision": []any{revision('a')}}, "revision:" + revision('a')},
		{map[string]any{"branch": []any{"main"}}, "branch:main"},
		{map[string]any{"range": []any{"1.0.0", "2.0.0"}}, "range:1.0.0..<2.0.0"},
	} {
		got, err := decodeRequirement(test.raw)
		if err != nil || got != test.want {
			t.Fatalf("decode requirement = %q, %v; want %q", got, err, test.want)
		}
	}
	for _, raw := range []any{nil, map[string]any{}, map[string]any{"exact": []any{}}, map[string]any{"range": []any{"1.0.0"}}, map[string]any{"unknown": []any{"x"}}, map[string]any{"exact": []any{1}}} {
		if _, err := decodeRequirement(raw); err == nil {
			t.Fatalf("invalid requirement accepted: %#v", raw)
		}
	}
	root := t.TempDir()
	inside := filepath.Join(root, "Packages", "Local")
	if err := os.MkdirAll(inside, 0o700); err != nil {
		t.Fatal(err)
	}
	dependency, err := decodeManifestDependency(map[string]any{"fileSystem": []any{map[string]any{"identity": "Local", "path": inside}}}, root)
	if err != nil || dependency.Kind != SourcePath || dependency.LocalPath != "Packages/Local" {
		t.Fatalf("file-system dependency = %+v, %v", dependency, err)
	}
	if _, err = decodeManifestDependency(map[string]any{"fileSystem": []any{map[string]any{"identity": "Escape", "path": filepath.Dir(root)}}}, root); ErrorCode(err) != CodeLocalDependencyOutside {
		t.Fatalf("escaping file-system dependency = %v", err)
	}
	for _, raw := range []map[string]json.RawMessage{
		{"target": json.RawMessage(`["Core",null]`)},
		{"byName": json.RawMessage(`["Core",null]`)},
		{"product": json.RawMessage(`["Core","dep",null,null]`)},
	} {
		if _, err = decodeTargetDependency(raw); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = decodeTargetDependency(map[string]json.RawMessage{"unknown": json.RawMessage(`["x"]`)}); err == nil {
		t.Fatal("unknown target dependency accepted")
	}
	setting, err := decodeBuildSetting(map[string]json.RawMessage{"unsafeFlags": json.RawMessage(`[["-I/tmp"],null]`)})
	if err != nil || !setting.Unsafe || settingKey(setting) == "" {
		t.Fatalf("unsafe setting = %+v, %v", setting, err)
	}
	languages := targetLanguages(Target{Sources: []string{"a.swift", "a.c", "a.cpp", "a.m", "a.mm", "ignored.txt"}})
	if want := []string{"c", "c++", "objective-c", "objective-c++", "swift"}; !reflect.DeepEqual(languages, want) {
		t.Fatalf("languages = %v, want %v", languages, want)
	}
	if (&Failure{Code: CodeGraphIncomplete, Detail: "x"}).Error() != "closure_graph_incomplete: x" || (&Failure{Code: CodeGraphIncomplete}).Error() != "closure_graph_incomplete" {
		t.Fatal("stable failure rendering drifted")
	}
	if mirrorRootFromArgs([]string{"--git-dir", "/tmp/mirror"}) != "/tmp/mirror" || mirrorRootFromArgs(nil) != "." {
		t.Fatal("mirror root argument parsing drifted")
	}
	for extension, want := range map[string]bool{".swift": true, ".c": true, ".cc": true, ".cpp": true, ".cxx": true, ".m": true, ".mm": true, ".s": true, ".txt": false} {
		if got := swiftPMSourceExtension(extension); got != want {
			t.Fatalf("source extension %s = %t, want %t", extension, got, want)
		}
	}
	if _, err := singletonKey(map[string]json.RawMessage{}); err == nil {
		t.Fatal("empty product type accepted")
	}
	if _, err := singletonKey(map[string]json.RawMessage{"a": nil, "b": nil}); err == nil {
		t.Fatal("multi-kind product type accepted")
	}
	remote, err := decodeManifestDependency(map[string]any{"sourceControl": []any{map[string]any{"identity": "Remote", "location": map[string]any{"remote": []any{"https://example.invalid/remote"}}, "requirement": map[string]any{"branch": []any{"main"}}}}}, root)
	if err != nil || remote.Kind != SourceRemote || remote.Requirement != "branch:main" {
		t.Fatalf("remote dependency = %+v, %v", remote, err)
	}
	if _, err = NewManager(Config{}); ErrorCode(err) != CodeDerivationUnauthorized {
		t.Fatalf("empty manager config = %v", err)
	}
	var manager *Manager
	if _, err = manager.CaptureAndClose(t.Context(), Request{}); ErrorCode(err) != CodeDerivationUnauthorized {
		t.Fatalf("nil manager = %v", err)
	}
	if err = validateRuntimeRoots(root, root); err == nil {
		t.Fatal("overlapping runtime roots accepted")
	}
}

func TestOfflineMetadataRejectsInvalidMountAndGraphEvidence(t *testing.T) {
	if _, err := mirrorMount("../escape"); ErrorCode(err) != CodeDependencyPinMismatch {
		t.Fatalf("invalid mirror mount error = %v", err)
	}
	if _, err := decodeDependencyIdentities([]byte(`{"identity":"root","dependencies":[{"identity":"a","dependencies":[]},{"identity":"a","dependencies":[]}]}`)); ErrorCode(err) != CodeBuildGraphDrift {
		t.Fatalf("duplicate metadata identity error = %v", err)
	}
	if _, err := decodeDependencyIdentities([]byte(`{"dependencies":[]}`)); ErrorCode(err) != CodeBuildGraphDrift {
		t.Fatalf("missing metadata root error = %v", err)
	}
	if equalStrings([]string{"a"}, []string{"a", "b"}) || equalStrings([]string{"a"}, []string{"b"}) {
		t.Fatal("unequal metadata identity sets compared equal")
	}
	a := closuregraph.ID("sha256:" + strings.Repeat("a", 64))
	b := closuregraph.ID("sha256:" + strings.Repeat("b", 64))
	if equalIDLists([]closuregraph.ID{a}, []closuregraph.ID{a, b}) || equalIDLists([]closuregraph.ID{a}, []closuregraph.ID{b}) {
		t.Fatal("unequal graph identity sets compared equal")
	}
}

type fixtureState struct {
	root      string
	config    Config
	evaluator *fakeEvaluator
	broker    *fakeBroker
}

func newFixture(t *testing.T) fixtureState {
	t.Helper()
	base := t.TempDir()
	t.Cleanup(func() {
		_ = filepath.WalkDir(base, func(current string, entry os.DirEntry, walkErr error) error {
			if walkErr == nil {
				if entry.IsDir() {
					_ = os.Chmod(current, 0o700)
				} else {
					_ = os.Chmod(current, 0o600)
				}
			}
			return nil
		})
	})
	root := writePackage(t, base, "root")
	a := writePackage(t, base, "a")
	b := writePackage(t, base, "b")
	if err := os.WriteFile(filepath.Join(root, "Package.resolved"), lockFixture(), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(a, "Package.resolved"), []byte(`{"version":3,"pins":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	mirrorA := filepath.Join(base, "mirror-a")
	mirrorB := filepath.Join(base, "mirror-b")
	for _, value := range []string{mirrorA, mirrorB} {
		if err := os.Mkdir(value, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	store, err := closureexec.NewCaptureStore(filepath.Join(base, "store"))
	if err != nil {
		t.Fatal(err)
	}
	evaluator := &fakeEvaluator{manifests: fixtureManifests()}
	broker := &fakeBroker{snapshots: map[string]Snapshot{"a": {Identity: "a", Root: a, MirrorRoot: mirrorA, Revision: revision('a'), GitTree: revision('c'), Kind: SourceRemote, BrokerReceiptID: id('a')}, "b": {Identity: "b", Root: b, MirrorRoot: mirrorB, Revision: revision('b'), GitTree: revision('d'), Kind: SourceRemote, BrokerReceiptID: id('b')}}}
	tools := Toolchain{Swift: tool("swift", '1'), SwiftPM: tool("swiftpm", '2'), PackageDescription: tool("package-description", '3'), Git: tool("git", '4')}
	tools.Recheck = func(_ context.Context, expected ToolIdentity) (closureexec.ToolchainIdentity, error) {
		return closureexec.ToolchainIdentity{Fingerprint: expected.Fingerprint, ExecutableSHA256: expected.ExecutableSHA256}, nil
	}
	config := Config{Store: store, Policy: artifactpolicy.NewService(), Evaluator: evaluator, Broker: broker, MirrorVerifier: fakeMirrorVerifier{}, OfflineRunner: fakeOfflineRunner{}, Toolchain: tools, Destination: Destination{Platform: closuregraph.TargetPlatformPayload{OS: "darwin", Architecture: "arm64", ABI: "darwin", Libc: "libSystem", MinimumRuntime: "macos-14", SDKID: "macos-sdk-v1", TargetTriple: "arm64-apple-macosx14.0", Runtime: "swift-6", LanguageModes: map[string]string{"swift": "6"}, Tuning: map[string]string{}}, Markers: map[string]string{"platform": "macos", "configuration": "release", "architecture": "arm64"}}, CausalHead: "sha256:" + strings.Repeat("0", 64)}
	config.ProcessStartObserver = func(permit ManifestPermit) { evaluator.starts = append(evaluator.starts, permit.PackageIdentity) }
	return fixtureState{root: root, config: config, evaluator: evaluator, broker: broker}
}

func fixtureManifests() map[string]Manifest {
	return map[string]Manifest{
		"root": {PackageName: "root", ToolsVersion: "6.1", Products: []Product{{Name: "cli", Type: "executable", Targets: []string{"App"}}}, Targets: []Target{{Name: "App", Type: "executable", Sources: []string{"Sources/App/main.swift"}, Dependencies: []TargetDependency{{Package: "a", Product: "AProd"}}}}, Dependencies: []ManifestDependency{{Identity: "a", Kind: SourceRemote, Location: "https://example.invalid/A", Requirement: "exact:1.0.0"}}},
		"a":    {PackageName: "a", ToolsVersion: "6.1", Products: []Product{{Name: "AProd", Type: "library", Targets: []string{"ATarget"}}}, Targets: []Target{{Name: "ATarget", Type: "regular", Sources: []string{"Sources/App/main.swift"}, Dependencies: []TargetDependency{{Package: "b", Product: "BProd"}}}}, Dependencies: []ManifestDependency{{Identity: "b", Kind: SourceRemote, Location: "https://example.invalid/B", Requirement: "exact:1.0.0"}}},
		"b":    {PackageName: "b", ToolsVersion: "6.1", Products: []Product{{Name: "BProd", Type: "library", Targets: []string{"BTarget"}}}, Targets: []Target{{Name: "BTarget", Type: "regular", Sources: []string{"Sources/App/main.swift"}}}}}
}

type fakeEvaluator struct {
	manifests map[string]Manifest
	starts    []string
}

func (e *fakeEvaluator) Evaluate(_ context.Context, root string, permit ManifestPermit) (ManifestResult, error) {
	if permit.Network != "none" || !contains(permit.Argv, "--disable-experimental-prebuilts") || contains(permit.Argv, "--manifest-path") {
		return ManifestResult{}, errors.New("open manifest permit")
	}
	if !filepath.IsAbs(root) {
		return ManifestResult{}, errors.New("root is not protected")
	}
	manifest, ok := e.manifests[permit.PackageIdentity]
	if !ok {
		return ManifestResult{}, fmt.Errorf("missing manifest %s", permit.PackageIdentity)
	}
	receipt, _ := closuregraph.DomainID("fake-swiftpm-manifest-receipt-v1", map[string]any{"package": permit.PackageIdentity, "permit": string(permit.ID), "selected": permit.SelectedManifest})
	return ManifestResult{Manifest: manifest, ReceiptID: receipt}, nil
}

type fakeBroker struct {
	snapshots    map[string]Snapshot
	failIdentity string
	starts       []string
}

func (b *fakeBroker) Acquire(_ context.Context, pin Pin) (Snapshot, error) {
	b.starts = append(b.starts, pin.Identity)
	if pin.Identity == b.failIdentity {
		return Snapshot{}, errors.New("missing mirror")
	}
	value, ok := b.snapshots[pin.Identity]
	if !ok {
		return Snapshot{}, errors.New("missing snapshot")
	}
	return value, nil
}

type fakeResolver struct {
	payload []byte
	calls   int
	issued  ResolutionResult
}

func (r *fakeResolver) Resolve(_ context.Context, root string, permit ResolutionPermit, _ Manifest) (ResolutionResult, error) {
	r.calls++
	if !filepath.IsAbs(root) || permit.Network != "broker-only" || permit.AlgorithmID != "swiftpm-brokered-resolution-v1" {
		return ResolutionResult{}, errors.New("uncontrolled resolver")
	}
	r.issued = ResolutionResult{Lock: append([]byte(nil), r.payload...), ReceiptID: id('7')}
	return r.issued, nil
}

func (r *fakeResolver) VerifyResult(_ ResolutionPermit, result ResolutionResult) error {
	if !reflect.DeepEqual(result, r.issued) {
		return errors.New("resolution result was not issued")
	}
	return nil
}

type fakeMirrorVerifier struct{}

func (fakeMirrorVerifier) Verify(_ context.Context, _ string, pin Pin, snapshot Snapshot) (GitVerificationEvidence, error) {
	if snapshot.Identity != pin.Identity || snapshot.Kind != pin.Kind || snapshot.Revision != pin.Revision || snapshot.GitTree == "" {
		return GitVerificationEvidence{}, errors.New("mirror identity mismatch")
	}
	return GitVerificationEvidence{}, nil
}

type fakeOfflineRunner struct{}

func (fakeOfflineRunner) Replay(_ context.Context, capture *Capture) (OfflineMetadataResult, error) {
	identities := make([]string, len(capture.Lock.Pins))
	for index, pin := range capture.Lock.Pins {
		identities[index] = pin.Identity
	}
	sort.Strings(identities)
	id, _ := closuregraph.DomainID("swiftpm-test-offline-receipt-v1", map[string]any{"identities": stringsAny(identities)})
	return OfflineMetadataResult{ReceiptID: id, PackageIdentities: identities}, nil
}

func writePackage(t *testing.T, base, name string) string {
	t.Helper()
	root := filepath.Join(base, name)
	if err := os.MkdirAll(filepath.Join(root, "Sources", "App"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "Package.swift"), []byte("// swift-tools-version: 6.1\nimport PackageDescription\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "Sources", "App", "main.swift"), []byte("public func value() -> Int { 1 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}
func lockFixture() []byte {
	return []byte(`{"version":3,"pins":[{"identity":"a","kind":"remoteSourceControl","location":"https://example.invalid/A","state":{"revision":"` + revision('a') + `","version":"1.0.0"}},{"identity":"b","kind":"remoteSourceControl","location":"https://example.invalid/B","state":{"revision":"` + revision('b') + `","version":"1.0.0"}}]}`)
}
func revision(value byte) string { return strings.Repeat(string(value), 40) }
func id(value byte) closuregraph.ID {
	return closuregraph.ID("sha256:" + strings.Repeat(string(value), 64))
}
func tool(role string, value byte) ToolIdentity {
	return ToolIdentity{Role: role, ExecutableRelativePath: "bin/" + role, VersionOutput: map[string]string{"swift": "Apple Swift version 6.3.2", "swiftpm": "Swift Package Manager 6.3.2", "package-description": "PackageDescription 6.3.2", "git": "git version 2.50.1"}[role], PlatformABI: "darwin-arm64", PolicySelector: "swift-toolchain-v1", Fingerprint: id(value), ExecutableSHA256: id(value + 3)}
}
func requireCode(t *testing.T, err error, want Code) {
	t.Helper()
	if ErrorCode(err) != want {
		t.Fatalf("error code = %q, want %q (err=%v)", ErrorCode(err), want, err)
	}
}
func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsID(values []closuregraph.ID, want closuregraph.ID) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
