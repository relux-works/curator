package pnpmsource

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/nodesource"
)

type fixturePackage struct {
	name, version string
	metadata      map[string]any
	files         map[string][]byte
}
type pnpmFixture struct {
	root, work string
	request    ParseRequest
	graph      Graph
	tarballs   map[string]RawTarball
	payloads   map[string][]byte
}

func newPNPMFixture(t *testing.T) pnpmFixture {
	t.Helper()
	root := filepath.Join(t.TempDir(), "project")
	work := filepath.Join(t.TempDir(), "work")
	if err := os.MkdirAll(filepath.Join(root, "packages", "cli"), 0o700); err != nil {
		t.Fatal(err)
	}
	rootManifest := mustJSON(t, map[string]any{"name": "app", "version": "1.0.0", "dependencies": map[string]string{"a": "^1.0.0", "cli": "workspace:*"}, "optionalDependencies": map[string]string{"optional": "1.0.0"}})
	workspaceManifest := mustJSON(t, map[string]any{"name": "cli", "version": "1.0.0", "dependencies": map[string]string{"b": "1.0.0"}})
	packages := []fixturePackage{
		{name: "a", version: "1.0.0", metadata: map[string]any{"dependencies": map[string]string{"b": "1.0.0", "peer": "2.0.0"}, "peerDependencies": map[string]string{"peer": "^2.0.0"}}, files: map[string][]byte{"index.js": []byte("module.exports = require('b')\n")}},
		{name: "b", version: "1.0.0", metadata: map[string]any{}, files: map[string][]byte{"index.js": []byte("module.exports = 'b'\n")}},
		{name: "peer", version: "2.0.0", metadata: map[string]any{}, files: map[string][]byte{"index.js": []byte("module.exports = 'peer'\n")}},
		{name: "optional", version: "1.0.0", metadata: map[string]any{"os": []string{"linux"}}, files: map[string][]byte{"index.js": []byte("module.exports = 'optional'\n")}},
	}
	tarballs := map[string]RawTarball{}
	payloads := map[string][]byte{}
	integrities := map[string]string{}
	for _, pkg := range packages {
		payload := buildTGZ(t, pkg)
		key := pkg.name + "@" + pkg.version
		payloads[key] = payload
		sum := sha512.Sum512(payload)
		integrities[key] = "sha512-" + base64.StdEncoding.EncodeToString(sum[:])
		file := filepath.Join(t.TempDir(), pkg.name+".tgz")
		if err := os.WriteFile(file, payload, 0o600); err != nil {
			t.Fatal(err)
		}
		tarballs[key] = RawTarball{Path: file}
	}
	lock := fmt.Sprintf(`lockfileVersion: '9.0'
settings:
  autoInstallPeers: true
  excludeLinksFromLockfile: false
importers:
  .:
    dependencies:
      a:
        specifier: ^1.0.0
        version: 1.0.0(peer@2.0.0)
      cli:
        specifier: workspace:*
        version: link:packages/cli
    optionalDependencies:
      optional:
        specifier: 1.0.0
        version: 1.0.0
  packages/cli:
    dependencies:
      b:
        specifier: 1.0.0
        version: 1.0.0
packages:
  a@1.0.0:
    resolution:
      integrity: %s
    peerDependencies:
      peer: ^2.0.0
  b@1.0.0:
    resolution:
      integrity: %s
  optional@1.0.0:
    resolution:
      integrity: %s
    os: [linux]
  peer@2.0.0:
    resolution:
      integrity: %s
snapshots:
  a@1.0.0(peer@2.0.0):
    dependencies:
      b: 1.0.0
      peer: 2.0.0
  b@1.0.0: {}
  optional@1.0.0: {}
  peer@2.0.0: {}
`, integrities["a@1.0.0"], integrities["b@1.0.0"], integrities["optional@1.0.0"], integrities["peer@2.0.0"])
	workspace := []byte("packages:\n  - packages/*\n")
	request := ParseRequest{LockBytes: []byte(lock), Manifests: map[string][]byte{"package.json": rootManifest, "packages/cli/package.json": workspaceManifest}, ConfigFiles: map[string][]byte{".npmrc": []byte("ignore-scripts=true\nside-effects-cache=false\n"), "pnpm-workspace.yaml": workspace}, AllowedRegistryOrigins: []string{"https://registry.npmjs.org"}, Target: Target{OS: "darwin", Architecture: "arm64", Libc: "none", IncludeDev: true}}
	for name, payload := range map[string][]byte{"package.json": rootManifest, "pnpm-lock.yaml": []byte(lock), ".npmrc": request.ConfigFiles[".npmrc"], "pnpm-workspace.yaml": workspace, "index.js": []byte("require('a')\n"), "packages/cli/package.json": workspaceManifest, "packages/cli/index.js": []byte("require('b')\n")} {
		target := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, payload, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	graph, err := Parse(request)
	if err != nil {
		t.Fatal(err)
	}
	return pnpmFixture{root: root, work: work, request: request, graph: graph, tarballs: tarballs, payloads: payloads}
}

func newTwoPeerContextFixture(t *testing.T) pnpmFixture {
	t.Helper()
	fixture := newPNPMFixture(t)
	peer := fixturePackage{name: "peer", version: "2.1.0", metadata: map[string]any{}, files: map[string][]byte{"index.js": []byte("module.exports = 'peer-2.1'\n")}}
	payload := buildTGZ(t, peer)
	sum := sha512.Sum512(payload)
	integrity := "sha512-" + base64.StdEncoding.EncodeToString(sum[:])
	lock := string(fixture.request.LockBytes)
	lock = strings.Replace(lock, `  packages/cli:
    dependencies:
      b:
        specifier: 1.0.0
        version: 1.0.0
`, `  packages/cli:
    dependencies:
      a:
        specifier: ^1.0.0
        version: 1.0.0(peer@2.1.0)
`, 1)
	lock = strings.Replace(lock, "  peer@2.0.0:\n    resolution:\n      integrity: "+fixture.graph.Packages[fixture.graph.packageIndex["peer@2.0.0"]].Integrity+"\nsnapshots:\n", "  peer@2.0.0:\n    resolution:\n      integrity: "+fixture.graph.Packages[fixture.graph.packageIndex["peer@2.0.0"]].Integrity+"\n  peer@2.1.0:\n    resolution:\n      integrity: "+integrity+"\nsnapshots:\n", 1)
	lock = strings.Replace(lock, "  b@1.0.0: {}\n", "  a@1.0.0(peer@2.1.0):\n    dependencies:\n      b: 1.0.0\n      peer: 2.1.0\n  b@1.0.0: {}\n", 1)
	lock = strings.Replace(lock, "  peer@2.0.0: {}\n", "  peer@2.0.0: {}\n  peer@2.1.0: {}\n", 1)
	workspaceManifest := mustJSON(t, map[string]any{"name": "cli", "version": "1.0.0", "dependencies": map[string]string{"a": "^1.0.0"}})
	fixture.request.LockBytes = []byte(lock)
	fixture.request.Manifests = cloneBytesMap(fixture.request.Manifests)
	fixture.request.Manifests["packages/cli/package.json"] = workspaceManifest
	fixture.graph = mustParse(t, fixture.request)
	file := filepath.Join(t.TempDir(), "peer-2.1.0.tgz")
	if err := os.WriteFile(file, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	fixture.tarballs = cloneTarballs(fixture.tarballs)
	fixture.tarballs["peer@2.1.0"] = RawTarball{Path: file}
	fixture.payloads["peer@2.1.0"] = payload
	if err := os.WriteFile(filepath.Join(fixture.root, "pnpm-lock.yaml"), fixture.request.LockBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture.root, "packages", "cli", "package.json"), workspaceManifest, 0o600); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func newHostPNPMFixture(t *testing.T) pnpmFixture {
	t.Helper()
	fixture := newPNPMFixture(t)
	pkg := fixturePackage{name: "optional", version: "1.0.0", metadata: map[string]any{"os": []string{runtime.GOOS}}, files: map[string][]byte{"index.js": []byte("module.exports = 'optional'\n")}}
	payload := buildTGZ(t, pkg)
	sum := sha512.Sum512(payload)
	integrity := "sha512-" + base64.StdEncoding.EncodeToString(sum[:])
	oldIntegrity := fixture.graph.Packages[fixture.graph.packageIndex["optional@1.0.0"]].Integrity
	lock := strings.Replace(string(fixture.request.LockBytes), oldIntegrity, integrity, 1)
	lock = strings.Replace(lock, "    os: [linux]\n", "    os: ["+runtime.GOOS+"]\n", 1)
	fixture.request.LockBytes = []byte(lock)
	fixture.request.Target.OS = runtime.GOOS
	fixture.request.Target.Architecture = runtime.GOARCH
	fixture.graph = mustParse(t, fixture.request)
	file := filepath.Join(t.TempDir(), "optional-host.tgz")
	if err := os.WriteFile(file, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	fixture.tarballs = cloneTarballs(fixture.tarballs)
	fixture.tarballs["optional@1.0.0"] = RawTarball{Path: file}
	fixture.payloads["optional@1.0.0"] = payload
	if err := os.WriteFile(filepath.Join(fixture.root, "pnpm-lock.yaml"), fixture.request.LockBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func newTargetPrunedSnapshotDependencyFixture(t *testing.T) pnpmFixture {
	t.Helper()
	fixture := newPNPMFixture(t)
	unsupportedOS := "linux"
	if runtime.GOOS == "linux" {
		unsupportedOS = "darwin"
	}
	optionalPackage := fixturePackage{
		name:     "optional",
		version:  "1.0.0",
		metadata: map[string]any{"dependencies": map[string]string{"b": "1.0.0"}, "os": []string{unsupportedOS}},
		files:    map[string][]byte{"index.js": []byte("module.exports = require('b')\n")},
	}
	aPackage := fixturePackage{
		name:    "a",
		version: "1.0.0",
		metadata: map[string]any{
			"dependencies":         map[string]string{"b": "1.0.0", "peer": "2.0.0"},
			"optionalDependencies": map[string]string{"optional": "1.0.0"},
			"peerDependencies":     map[string]string{"peer": "^2.0.0"},
		},
		files: map[string][]byte{"index.js": []byte("module.exports = require('b')\n")},
	}
	optionalPayload := buildTGZ(t, optionalPackage)
	aPayload := buildTGZ(t, aPackage)
	optionalSum := sha512.Sum512(optionalPayload)
	aSum := sha512.Sum512(aPayload)
	optionalIntegrity := "sha512-" + base64.StdEncoding.EncodeToString(optionalSum[:])
	aIntegrity := "sha512-" + base64.StdEncoding.EncodeToString(aSum[:])
	oldOptionalIntegrity := fixture.graph.Packages[fixture.graph.packageIndex["optional@1.0.0"]].Integrity
	oldAIntegrity := fixture.graph.Packages[fixture.graph.packageIndex["a@1.0.0"]].Integrity
	lock := strings.Replace(string(fixture.request.LockBytes), oldOptionalIntegrity, optionalIntegrity, 1)
	lock = strings.Replace(lock, oldAIntegrity, aIntegrity, 1)
	lock = strings.Replace(lock, "    os: [linux]\n", "    os: ["+unsupportedOS+"]\n", 1)
	lock = strings.Replace(lock, "    optionalDependencies:\n      optional:\n        specifier: 1.0.0\n        version: 1.0.0\n", "", 1)
	lock = strings.Replace(lock, "      peer: 2.0.0\n", "      peer: 2.0.0\n    optionalDependencies:\n      optional: 1.0.0\n", 1)
	lock = strings.Replace(lock, "  optional@1.0.0: {}\n", "  optional@1.0.0:\n    dependencies:\n      b: 1.0.0\n", 1)
	rootManifest := mustJSON(t, map[string]any{"name": "app", "version": "1.0.0", "dependencies": map[string]string{"a": "^1.0.0", "cli": "workspace:*"}})
	fixture.request.LockBytes = []byte(lock)
	fixture.request.Manifests = cloneBytesMap(fixture.request.Manifests)
	fixture.request.Manifests["package.json"] = rootManifest
	fixture.request.Target.OS = runtime.GOOS
	fixture.request.Target.Architecture = runtime.GOARCH
	fixture.graph = mustParse(t, fixture.request)
	optionalFile := filepath.Join(t.TempDir(), "optional-pruned.tgz")
	if err := os.WriteFile(optionalFile, optionalPayload, 0o600); err != nil {
		t.Fatal(err)
	}
	aFile := filepath.Join(t.TempDir(), "a-with-optional.tgz")
	if err := os.WriteFile(aFile, aPayload, 0o600); err != nil {
		t.Fatal(err)
	}
	fixture.tarballs = cloneTarballs(fixture.tarballs)
	fixture.tarballs["optional@1.0.0"] = RawTarball{Path: optionalFile}
	fixture.tarballs["a@1.0.0"] = RawTarball{Path: aFile}
	fixture.payloads["optional@1.0.0"] = optionalPayload
	fixture.payloads["a@1.0.0"] = aPayload
	if err := os.WriteFile(filepath.Join(fixture.root, "pnpm-lock.yaml"), fixture.request.LockBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture.root, "package.json"), rootManifest, 0o600); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func newUnreachableSnapshotDependencyFixture(t *testing.T) pnpmFixture {
	return newUnreachableSnapshotDependencyTargetFixture(t, "", "")
}

func newUnreachableSnapshotDependencyTargetFixture(t *testing.T, selector, unsupported string) pnpmFixture {
	t.Helper()
	fixture := newPNPMFixture(t)
	metadata := map[string]any{"dependencies": map[string]string{"b": "1.0.0"}}
	packageSelector := ""
	if selector != "" {
		metadata[selector] = []string{unsupported}
		packageSelector = "\n    " + selector + ": [" + unsupported + "]"
	}
	pkg := fixturePackage{
		name:     "dormant",
		version:  "1.0.0",
		metadata: metadata,
		files:    map[string][]byte{"index.js": []byte("module.exports = require('b')\n")},
	}
	payload := buildTGZ(t, pkg)
	sum := sha512.Sum512(payload)
	integrity := "sha512-" + base64.StdEncoding.EncodeToString(sum[:])
	lock := strings.Replace(string(fixture.request.LockBytes), "snapshots:\n", "  dormant@1.0.0:\n    resolution:\n      integrity: "+integrity+packageSelector+"\nsnapshots:\n", 1)
	lock = strings.Replace(lock, "  peer@2.0.0: {}\n", "  peer@2.0.0: {}\n  dormant@1.0.0:\n    dependencies:\n      b: 1.0.0\n", 1)
	fixture.request.LockBytes = []byte(lock)
	fixture.graph = mustParse(t, fixture.request)
	file := filepath.Join(t.TempDir(), "dormant.tgz")
	if err := os.WriteFile(file, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	fixture.tarballs = cloneTarballs(fixture.tarballs)
	fixture.tarballs["dormant@1.0.0"] = RawTarball{Path: file}
	fixture.payloads["dormant@1.0.0"] = payload
	if err := os.WriteFile(filepath.Join(fixture.root, "pnpm-lock.yaml"), fixture.request.LockBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func materializeFixture(t *testing.T, capture *Capture, mutate func(string) error) (*Materialization, error) {
	t.Helper()
	runner := &fakePNPMRunner{capture: capture, mutateInstall: mutate}
	authority := makeExecutionContext(t, capture, runner)
	runner.authority = authority
	storeRoot := filepath.Join(t.TempDir(), "store")
	store, err := DerivePrivateStore(context.Background(), capture, storeRoot, authority)
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() { _ = makeTreeWritable(storeRoot) })
	return Materialize(context.Background(), store, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: t.TempDir()}, authority)
}

func snapshotDependencyLink(root, snapshot, dependency string) string {
	return filepath.Join(root, "node_modules", ".pnpm", pnpmSnapshotDirectory(snapshot), "node_modules", filepath.FromSlash(dependency))
}

func TestS01N01ClosedLockImporterPeerWorkspaceAndTargetGraph(t *testing.T) {
	fixture := newPNPMFixture(t)
	if fixture.graph.LockfileVersion != "9.0" || len(fixture.graph.Importers) != 2 || len(fixture.graph.Snapshots) != 4 || len(fixture.graph.LocalRoots) != 2 {
		t.Fatalf("unexpected graph: %+v", fixture.graph)
	}
	a := fixture.graph.Snapshots[fixture.graph.snapshotIndex["a@1.0.0(peer@2.0.0)"]]
	if a.PeerContext != "(peer@2.0.0)" || !a.Selected {
		t.Fatalf("peer context not retained: %+v", a)
	}
	optional := fixture.graph.Snapshots[fixture.graph.snapshotIndex["optional@1.0.0"]]
	if optional.Selected || optional.PruneReason != "os_mismatch" {
		t.Fatalf("target pruning not retained: %+v", optional)
	}
	workspaceEdge := false
	workspaceDependencyEdge := false
	peerEdge := false
	for _, edge := range fixture.graph.Edges {
		workspaceEdge = workspaceEdge || (edge.From == "local:." && edge.To == "local:packages/cli")
		workspaceDependencyEdge = workspaceDependencyEdge || (edge.From == "local:packages/cli" && edge.Name == "b" && edge.Selected)
		peerEdge = peerEdge || (edge.From == a.Key && edge.Scope == "peer" && edge.To == "peer@2.0.0")
	}
	if !workspaceEdge || !workspaceDependencyEdge || !peerEdge {
		t.Fatalf("workspace/peer edges missing: %+v", fixture.graph.Edges)
	}
	capture := captureFixture(t, fixture)
	workspaceID := capture.NodeCapture.PackageNodeIDs["local:packages/cli"]
	bID := capture.NodeCapture.PackageNodeIDs["b@1.0.0"]
	commonEdge := false
	for _, edge := range capture.NodeCapture.Edges {
		commonEdge = commonEdge || (edge.Kind == closuregraph.EdgeRequires && edge.FromNodeID == workspaceID && edge.ToNodeID == bID)
	}
	if !commonEdge {
		t.Fatal("selected workspace dependency disappeared from common Node capture")
	}
	runner := &fakePNPMRunner{capture: capture}
	authority := makeExecutionContext(t, capture, runner)
	runner.authority = authority
	evaluator := closuregraph.ConditionEvaluatorFunc{EvaluatorID: "pnpm-platform-selector-v1", EvaluateFunc: evaluatePlatform}
	bundle, _, err := nodesource.Close(capture.NodeCapture, authority.Selection, authority.Runtime, []closuregraph.ConditionEvaluator{evaluator}, closureexec.PortableExecutionPolicyID)
	if err != nil {
		t.Fatal(err)
	}
	activeWorkspaceDependency := false
	for _, edge := range bundle.Records.CaptureEdges {
		if edge.Kind != closuregraph.EdgeRequires || edge.FromNodeID != workspaceID || edge.ToNodeID != bID {
			continue
		}
		workspaceSelected, dependencySelected := false, false
		for _, activation := range bundle.Active.NodeActivations {
			workspaceSelected = workspaceSelected || (activation.NodeID == workspaceID && activation.State == closuregraph.ActivationSelected)
			dependencySelected = dependencySelected || (activation.NodeID == bID && activation.State == closuregraph.ActivationSelected)
		}
		activeWorkspaceDependency = workspaceSelected && dependencySelected
	}
	if !activeWorkspaceDependency {
		t.Fatal("selected workspace dependency disappeared from common active graph")
	}
}

func TestPinnedPNPMManagerPatchHashNormalizesCRLF(t *testing.T) {
	hash, err := managerPatchHash([]byte("a\r\nb\r\n"))
	if err != nil {
		t.Fatal(err)
	}
	const expected = "911169ddaaf146aff539f58c26c489af3b892dff0fe283c1c264c65ae5aa59a2"
	if hash != expected {
		t.Fatalf("manager patch hash = %q, want %q", hash, expected)
	}
}

func TestS06CanonicalLockOrderAndN10TargetSelection(t *testing.T) {
	fixture := newPNPMFixture(t)
	request := fixture.request
	request.LockBytes = []byte(strings.ReplaceAll(string(request.LockBytes), "settings:\n  autoInstallPeers: true\n  excludeLinksFromLockfile: false", "settings:\n  excludeLinksFromLockfile: false\n  autoInstallPeers: true"))
	second, err := Parse(request)
	if err != nil {
		t.Fatal(err)
	}
	if second.LockDigest != fixture.graph.LockDigest || second.RawLockSHA256 == fixture.graph.RawLockSHA256 {
		t.Fatalf("semantic/raw identities not separated")
	}
	request = fixture.request
	request.Target = Target{OS: "linux", Architecture: "arm64", Libc: "glibc", IncludeDev: true}
	linux, err := Parse(request)
	if err != nil {
		t.Fatal(err)
	}
	if !linux.Snapshots[linux.snapshotIndex["optional@1.0.0"]].Selected {
		t.Fatal("linux optional snapshot not selected")
	}
}

func TestN02LockSchemaManifestAndPeerDriftFailClosed(t *testing.T) {
	fixture := newPNPMFixture(t)
	variants := []struct {
		name   string
		mutate func(*ParseRequest)
		code   string
	}{{"missing", func(r *ParseRequest) { r.LockBytes = nil }, CodeLockMissing}, {"schema", func(r *ParseRequest) {
		r.LockBytes = []byte(strings.Replace(string(r.LockBytes), "'9.0'", "'10.0'", 1))
	}, CodeLockFormatUnsupported}, {"manifest", func(r *ParseRequest) {
		r.Manifests = cloneBytesMap(r.Manifests)
		r.Manifests["package.json"] = mustJSON(t, map[string]any{"name": "app", "version": "1.0.0"})
	}, CodeLockStale}, {"peer", func(r *ParseRequest) {
		r.LockBytes = []byte(strings.Replace(string(r.LockBytes), "a@1.0.0(peer@2.0.0):", "a@1.0.0(peer@3.0.0):", 1))
	}, CodeGraphIncomplete}}
	for _, variant := range variants {
		t.Run(variant.name, func(t *testing.T) {
			request := fixture.request
			variant.mutate(&request)
			_, err := Parse(request)
			assertCode(t, err, variant.code)
		})
	}
}

func TestN09N11ExtensionsConfigPatchesAndLocalRootsFailClosed(t *testing.T) {
	fixture := newPNPMFixture(t)
	t.Run("pnpmfile", func(t *testing.T) {
		request := fixture.request
		request.ConfigFiles = cloneBytesMap(request.ConfigFiles)
		request.ConfigFiles[".pnpmfile.cjs"] = []byte("module.exports = {}\n")
		_, err := Parse(request)
		assertCode(t, err, CodeManagerPluginUndeclared)
	})
	t.Run("custom resolver", func(t *testing.T) {
		request := fixture.request
		request.ConfigFiles = cloneBytesMap(request.ConfigFiles)
		request.ConfigFiles[".npmrc"] = []byte("resolve-peers-from-workspace-root=true\ncustom-fetcher=./fetch.js\n")
		_, err := Parse(request)
		assertCode(t, err, CodeManagerPluginUndeclared)
	})
	t.Run("side effects", func(t *testing.T) {
		request := fixture.request
		request.ConfigFiles = cloneBytesMap(request.ConfigFiles)
		request.ConfigFiles[".npmrc"] = []byte("side-effects-cache=true\n")
		_, err := Parse(request)
		assertCode(t, err, CodeHookUndeclared)
	})
	t.Run("undeclared patch", func(t *testing.T) {
		request := fixture.request
		request.PatchFiles = map[string][]byte{"patches/a.patch": []byte("diff --git a/a b/a\n")}
		_, err := Parse(request)
		assertCode(t, err, CodeManagerPluginUndeclared)
	})
	t.Run("file root", func(t *testing.T) {
		request := fixture.request
		request.LockBytes = []byte(strings.Replace(string(request.LockBytes), "version: link:packages/cli", "version: file:../outside", 1))
		_, err := Parse(request)
		assertCode(t, err, CodeLocalPathEscape)
	})
}

func TestContainedFileDependencyUsesIndependentLocalRoot(t *testing.T) {
	fixture := newPNPMFixture(t)
	fixture.request.LockBytes = []byte(strings.ReplaceAll(string(fixture.request.LockBytes), "specifier: workspace:*\n        version: link:packages/cli", "specifier: file:packages/cli\n        version: file:packages/cli"))
	fixture.request.Manifests = cloneBytesMap(fixture.request.Manifests)
	fixture.request.Manifests["package.json"] = mustJSON(t, map[string]any{"name": "app", "version": "1.0.0", "dependencies": map[string]string{"a": "^1.0.0", "cli": "file:packages/cli"}, "optionalDependencies": map[string]string{"optional": "1.0.0"}})
	if err := os.WriteFile(filepath.Join(fixture.root, "package.json"), fixture.request.Manifests["package.json"], 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture.root, "pnpm-lock.yaml"), fixture.request.LockBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	fixture.graph = mustParse(t, fixture.request)
	capture := captureFixture(t, fixture)
	if len(capture.Evidence.LocalRoots) != 2 {
		t.Fatalf("file dependency was not independently captured: %+v", capture.Evidence.LocalRoots)
	}
	found := false
	for _, edge := range fixture.graph.Edges {
		found = found || (edge.From == "local:." && edge.To == "local:packages/cli")
	}
	if !found {
		t.Fatal("file dependency did not resolve to its exact contained local root")
	}
}

func TestCraftedDependencyPathFailsBeforeMaterialization(t *testing.T) {
	fixture := newPNPMFixture(t)
	request := fixture.request
	request.LockBytes = []byte(strings.Replace(string(request.LockBytes), "b@1.0.0:\n    resolution:", "../../../../tmp/pwned@1.0.0:\n    resolution:", 1))
	_, err := Parse(request)
	assertCode(t, err, CodeLockFormatUnsupported)
}

func TestClosedParserAndInvocationBoundaryVariants(t *testing.T) {
	fixture := newPNPMFixture(t)
	for _, testCase := range []struct{ name, lock, code string }{
		{"duplicate YAML key", "lockfileVersion: '9.0'\nlockfileVersion: '9.0'\n", CodeLockFormatUnsupported},
		{"trailing YAML document", string(fixture.request.LockBytes) + "---\nlockfileVersion: '10.0'\n", CodeLockFormatUnsupported},
		{"trailing YAML content", string(fixture.request.LockBytes) + "---\nconflicting: true\n", CodeLockFormatUnsupported},
		{"unknown snapshot context", strings.Replace(string(fixture.request.LockBytes), "a@1.0.0(peer@2.0.0):", "a@1.0.0(hidden-context):", 1), CodeLockFormatUnsupported},
		{"nested peer context", strings.Replace(string(fixture.request.LockBytes), "a@1.0.0(peer@2.0.0):", "a@1.0.0(peer@2.0.0(other@1.0.0)):", 1), CodeLockFormatUnsupported},
		{"missing integrity", strings.Replace(string(fixture.request.LockBytes), "integrity: "+fixture.graph.Packages[0].Integrity, "integrity: ''", 1), CodeIntegrityMissing},
		{"unsafe origin", strings.Replace(string(fixture.request.LockBytes), "resolution:\n      integrity:", "resolution:\n      tarball: http://example.invalid/a.tgz\n      integrity:", 1), CodeOriginUnpinned},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := fixture.request
			request.LockBytes = []byte(testCase.lock)
			_, err := Parse(request)
			assertCode(t, err, testCase.code)
		})
	}
	t.Run("workspace trailing document", func(t *testing.T) {
		request := fixture.request
		request.ConfigFiles = cloneBytesMap(request.ConfigFiles)
		request.ConfigFiles["pnpm-workspace.yaml"] = append(request.ConfigFiles["pnpm-workspace.yaml"], []byte("---\npackages:\n  - other/*\n")...)
		_, err := Parse(request)
		assertCode(t, err, CodeLockFormatUnsupported)
	})
	capture := captureFixture(t, fixture)
	runner := &fakePNPMRunner{capture: capture}
	authority := makeExecutionContext(t, capture, runner)
	runner.authority = authority
	storeRoot := filepath.Join(t.TempDir(), "store")
	store, err := DerivePrivateStore(context.Background(), capture, storeRoot, authority)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTreeWritable(storeRoot) })
	materialized, err := Materialize(context.Background(), store, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "out"), WorkRoot: t.TempDir()}, authority)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range []string{"../escape.js", "missing.js"} {
		_, err = Invoke(context.Background(), materialized, entry, nil, authority)
		if ErrorCode(err) == "" {
			t.Fatalf("entry %q did not fail with adapter diagnostic: %v", entry, err)
		}
	}
}

func TestOverridesAndDeclaredPatchAreBoundAndAdmitted(t *testing.T) {
	fixture := newPNPMFixture(t)
	patchPath := "patches/a.patch"
	patchBytes := []byte("diff --git a/index.js b/index.js\n--- a/index.js\n+++ b/index.js\n@@ -1 +1 @@\n-module.exports = require('b')\n+module.exports = 'patched-a'\n")
	managerHash, err := managerPatchHash(patchBytes)
	if err != nil {
		t.Fatal(err)
	}
	fixture.request.LockBytes = []byte(strings.Replace(string(fixture.request.LockBytes), "importers:\n", fmt.Sprintf("overrides:\n  b: 1.0.0\npatchedDependencies:\n  a@1.0.0:\n    hash: %s\n    path: patches/a.patch\nimporters:\n", managerHash), 1))
	fixture.request.LockBytes = []byte(strings.ReplaceAll(string(fixture.request.LockBytes), "1.0.0(peer@2.0.0)", "1.0.0(patch_hash="+managerHash+")(peer@2.0.0)"))
	fixture.request.PatchFiles = map[string][]byte{patchPath: patchBytes}
	fixture.request.Manifests = cloneBytesMap(fixture.request.Manifests)
	fixture.request.Manifests["package.json"] = mustJSON(t, map[string]any{"name": "app", "version": "1.0.0", "dependencies": map[string]string{"a": "^1.0.0", "cli": "workspace:*"}, "optionalDependencies": map[string]string{"optional": "1.0.0"}, "pnpm": map[string]any{"overrides": map[string]string{"b": "1.0.0"}, "patchedDependencies": map[string]string{"a@1.0.0": patchPath}}})
	if err := os.MkdirAll(filepath.Join(fixture.root, "patches"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture.root, patchPath), patchBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture.root, "package.json"), fixture.request.Manifests["package.json"], 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture.root, "pnpm-lock.yaml"), fixture.request.LockBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	fixture.graph = mustParse(t, fixture.request)
	if fixture.graph.Overrides["b"] != "1.0.0" || len(fixture.graph.Patches) != 1 || !strings.HasPrefix(fixture.graph.Patches[0].SHA256, "sha256:") {
		t.Fatalf("override/patch not bound: %+v %+v", fixture.graph.Overrides, fixture.graph.Patches)
	}
	capture := captureFixture(t, fixture)
	if len(capture.Evidence.Patches) != 1 || !capture.Evidence.Patches[0].ArtifactManifestID.Valid() || len(capture.Evidence.PatchTransforms) != 1 || !capture.Evidence.PatchTransforms[0].ReceiptID.Valid() {
		t.Fatalf("patch not independently admitted/transformed: patches=%+v transforms=%+v", capture.Evidence.Patches, capture.Evidence.PatchTransforms)
	}
	runner := &fakePNPMRunner{capture: capture}
	authority := makeExecutionContext(t, capture, runner)
	runner.authority = authority
	storeRoot := filepath.Join(t.TempDir(), "store")
	store, err := DerivePrivateStore(context.Background(), capture, storeRoot, authority)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTreeWritable(storeRoot) })
	materialized, err := Materialize(context.Background(), store, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: t.TempDir()}, authority)
	if err != nil {
		t.Fatal(err)
	}
	patchedRoot := filepath.Join(materialized.Root, "node_modules", ".pnpm", pnpmSnapshotDirectory("a@1.0.0(patch_hash="+managerHash+")(peer@2.0.0)"), "node_modules", "a", "index.js")
	patched, err := os.ReadFile(patchedRoot)
	if err != nil || string(patched) != "module.exports = 'patched-a'\n" {
		t.Fatalf("patched materialization mismatch: payload=%q err=%v", patched, err)
	}
	for _, variant := range []struct {
		name   string
		mutate func(*ParseRequest)
	}{
		{"stale manager hash", func(request *ParseRequest) {
			request.LockBytes = []byte(strings.ReplaceAll(string(request.LockBytes), managerHash, strings.Repeat("0", 64)))
		}},
		{"patch content drift", func(request *ParseRequest) {
			request.PatchFiles = cloneBytesMap(request.PatchFiles)
			request.PatchFiles[patchPath] = append(append([]byte(nil), request.PatchFiles[patchPath]...), []byte("# drift\n")...)
		}},
	} {
		t.Run(variant.name, func(t *testing.T) {
			request := fixture.request
			variant.mutate(&request)
			_, err := Parse(request)
			assertCode(t, err, CodeIntegrityMismatch)
		})
	}
}

func TestCaptureAdmitsTarballsAndLocalRootsIndependently(t *testing.T) {
	fixture := newPNPMFixture(t)
	capture := captureFixture(t, fixture)
	if len(capture.Evidence.Tarballs) != 4 || len(capture.Evidence.LocalRoots) != 2 || !capture.Evidence.NodeCaptureGraphID.Valid() {
		t.Fatalf("incomplete capture evidence: %+v", capture.Evidence)
	}
	if capture.Evidence.LocalRoots[0].IntakeReceiptID == capture.Evidence.LocalRoots[1].IntakeReceiptID {
		t.Fatal("local roots were not captured independently")
	}
}

func TestS02N06N12IntegrityNativeAndAmbientStoreVectors(t *testing.T) {
	t.Run("missing raw input", func(t *testing.T) {
		fixture := newPNPMFixture(t)
		variant := fixture
		variant.tarballs = cloneTarballs(fixture.tarballs)
		delete(variant.tarballs, "b@1.0.0")
		_, err := captureFixtureError(t, variant)
		assertCode(t, err, CodeOfflineInputMissing)
	})
	t.Run("integrity", func(t *testing.T) {
		fixture := newPNPMFixture(t)
		variant := fixture
		variant.tarballs = cloneTarballs(fixture.tarballs)
		file := filepath.Join(t.TempDir(), "bad.tgz")
		_ = os.WriteFile(file, []byte("tampered"), 0o600)
		variant.tarballs["b@1.0.0"] = RawTarball{Path: file}
		_, err := captureFixtureError(t, variant)
		assertCode(t, err, CodeIntegrityMismatch)
	})
	t.Run("compiled", func(t *testing.T) {
		fixture := newPNPMFixture(t)
		variant := fixture
		replacePackage(t, &variant, "b@1.0.0", fixturePackage{name: "b", version: "1.0.0", files: map[string][]byte{"renamed.txt": {0, 0x61, 0x73, 0x6d, 1, 0, 0, 0}}})
		_, err := captureFixtureError(t, variant)
		if artifactpolicy.ErrorCode(err) != artifactpolicy.CodeCompiledDependency {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("ambient ignored", func(t *testing.T) {
		fixture := newPNPMFixture(t)
		ambient := filepath.Join(fixture.root, "node_modules", "a")
		_ = os.MkdirAll(ambient, 0o700)
		_ = os.WriteFile(filepath.Join(ambient, "poison.js"), []byte("poison"), 0o600)
		capture := captureFixture(t, fixture)
		if len(capture.Evidence.DiscardedAmbientPaths) != 1 || capture.Evidence.DiscardedAmbientPaths[0] != "node_modules" {
			t.Fatalf("ambient tree became authority: %+v", capture.Evidence.DiscardedAmbientPaths)
		}
	})
	t.Run("side effects rejected", func(t *testing.T) {
		fixture := newPNPMFixture(t)
		store := filepath.Join(fixture.root, ".pnpm-store")
		_ = os.MkdirAll(store, 0o700)
		_ = os.WriteFile(filepath.Join(store, "pkg-side-effects.json"), []byte("{}"), 0o600)
		_, err := captureFixtureError(t, fixture)
		assertCode(t, err, CodeHookUndeclared)
	})
}

func TestS03S08N01PrivateStoreAndFrozenOfflineMaterialization(t *testing.T) {
	fixture := newPNPMFixture(t)
	capture := captureFixture(t, fixture)
	runner := &fakePNPMRunner{capture: capture}
	authority := makeExecutionContext(t, capture, runner)
	runner.authority = authority
	storeRoot := filepath.Join(t.TempDir(), "private-store")
	store, err := DerivePrivateStore(context.Background(), capture, storeRoot, authority)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTreeWritable(storeRoot) })
	if runner.starts != 1 || !store.Receipt.ID.Valid() || len(store.Receipt.Files) == 0 {
		t.Fatalf("private store evidence missing: starts=%d receipt=%+v", runner.starts, store.Receipt)
	}
	if !containsAll(runner.storePermit.Argv, "store", "add", "--config.side-effects-cache=false") || runner.storePermit.Environment["NPM_CONFIG_OFFLINE"] != "true" || runner.storePermit.Environment["NPM_CONFIG_IGNORE_SCRIPTS"] != "true" {
		t.Fatalf("store argv is not closed: %v", runner.storePermit.Argv)
	}
	materializedRoot := filepath.Join(t.TempDir(), "materialized")
	materialized, err := Materialize(context.Background(), store, MaterializeRequest{Destination: materializedRoot, WorkRoot: t.TempDir()}, authority)
	if err != nil {
		t.Fatal(err)
	}
	if runner.starts != 2 || len(materialized.MaterializedPackages) != 3 {
		t.Fatalf("unexpected replay: starts=%d packages=%v", runner.starts, materialized.MaterializedPackages)
	}
	if !containsAll(runner.installPermit.Argv, "install", "--frozen-lockfile", "--offline", "--ignore-scripts", "--config.side-effects-cache=false", "--package-import-method=copy") {
		t.Fatalf("install argv is not closed: %v", runner.installPermit.Argv)
	}
	storeMount := argAfter(runner.installPermit.Argv, "--store-dir")
	if storeMount == "" {
		t.Fatal("install omitted private store")
	}
	for _, write := range runner.installPermit.WriteRoots {
		if strings.Contains(write, "capture/pnpm-private-store") {
			t.Fatalf("read-only store appeared in write roots: %v", runner.installPermit.WriteRoots)
		}
	}
	if !containsAll(runner.installPermit.WriteRoots, "work/pnpm-install-project", "work/pnpm-install-store") {
		t.Fatalf("install omitted declared writable project/store overlay: %v", runner.installPermit.WriteRoots)
	}
	invokeReceipt, err := Invoke(context.Background(), materialized, "index.js", []string{"--version"}, authority)
	if err != nil {
		t.Fatal(err)
	}
	if invokeReceipt.Decision != "success" || runner.starts != 3 || runner.nodePermit.Network != "none" || !containsAll(runner.nodePermit.Argv, "index.js", "--version") {
		t.Fatalf("offline invocation evidence is incomplete: starts=%d permit=%+v receipt=%+v", runner.starts, runner.nodePermit, invokeReceipt)
	}
}

func TestLockSupersetSnapshotDependencyLinksAreReconciledIndependentlyOfSelection(t *testing.T) {
	fixture := newTargetPrunedSnapshotDependencyFixture(t)
	snapshot := fixture.graph.Snapshots[fixture.graph.snapshotIndex["optional@1.0.0"]]
	if snapshot.Selected || !snapshot.Reachable || snapshot.PruneReason != "os_mismatch" {
		t.Fatalf("unexpected lock-superset selection: %+v", snapshot)
	}
	capture := captureFixture(t, fixture)
	materialized, err := materializeFixture(t, capture, nil)
	if err != nil {
		t.Fatal(err)
	}
	if containsAll(materialized.MaterializedPackages, snapshot.Key) {
		t.Fatalf("lock-superset snapshot leaked into active package set: %v", materialized.MaterializedPackages)
	}

	t.Run("missing link", func(t *testing.T) {
		fixture := newTargetPrunedSnapshotDependencyFixture(t)
		capture := captureFixture(t, fixture)
		_, err := materializeFixture(t, capture, func(root string) error {
			return os.Remove(snapshotDependencyLink(root, "optional@1.0.0", "b"))
		})
		assertCode(t, err, CodeGraphIncomplete)
	})

	t.Run("target pruned swapped link", func(t *testing.T) {
		fixture := newTargetPrunedSnapshotDependencyFixture(t)
		capture := captureFixture(t, fixture)
		_, err := materializeFixture(t, capture, func(root string) error {
			link := snapshotDependencyLink(root, "optional@1.0.0", "b")
			if err := os.Remove(link); err != nil {
				return err
			}
			wrong := filepath.Join(root, "node_modules", ".pnpm", pnpmSnapshotDirectory("peer@2.0.0"), "node_modules", "peer")
			return os.Symlink(wrong, link)
		})
		assertCode(t, err, CodeGraphIncomplete)
	})

	t.Run("target pruned unclaimed link", func(t *testing.T) {
		fixture := newTargetPrunedSnapshotDependencyFixture(t)
		capture := captureFixture(t, fixture)
		_, err := materializeFixture(t, capture, func(root string) error {
			target := filepath.Join(root, "node_modules", ".pnpm", pnpmSnapshotDirectory("peer@2.0.0"), "node_modules", "peer")
			return os.Symlink(target, snapshotDependencyLink(root, "optional@1.0.0", "rogue"))
		})
		assertCode(t, err, CodeGraphIncomplete)
	})
}

func TestUnreachableLockSupersetSnapshotRejectsBeforeInstall(t *testing.T) {
	fixture := newUnreachableSnapshotDependencyFixture(t)
	capture := captureFixture(t, fixture)
	runner := &fakePNPMRunner{capture: capture}
	authority := makeExecutionContext(t, capture, runner)
	runner.authority = authority
	storeRoot := filepath.Join(t.TempDir(), "store")
	store, err := DerivePrivateStore(t.Context(), capture, storeRoot, authority)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTreeWritable(storeRoot) })
	_, err = Materialize(t.Context(), store, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: t.TempDir()}, authority)
	assertCode(t, err, CodeGraphIncomplete)
	if runner.starts != 1 {
		t.Fatalf("unreachable snapshot started install: starts=%d", runner.starts)
	}
}

func TestTargetPrunedUnreachableSnapshotRejectsBeforeInstall(t *testing.T) {
	for _, variant := range []struct {
		selector    string
		unsupported string
		reason      string
	}{
		{selector: "os", unsupported: "unsupported-os", reason: "os_mismatch"},
		{selector: "cpu", unsupported: "unsupported-cpu", reason: "cpu_mismatch"},
		{selector: "libc", unsupported: "unsupported-libc", reason: "libc_mismatch"},
	} {
		t.Run(variant.selector, func(t *testing.T) {
			fixture := newUnreachableSnapshotDependencyTargetFixture(t, variant.selector, variant.unsupported)
			snapshot := fixture.graph.Snapshots[fixture.graph.snapshotIndex["dormant@1.0.0"]]
			if snapshot.Selected || snapshot.PruneReason != variant.reason || snapshot.Reachable {
				t.Fatalf("unexpected unreachable target-pruned snapshot: %+v", snapshot)
			}
			capture := captureFixture(t, fixture)
			runner := &fakePNPMRunner{capture: capture}
			authority := makeExecutionContext(t, capture, runner)
			runner.authority = authority
			storeRoot := filepath.Join(t.TempDir(), "store")
			store, err := DerivePrivateStore(t.Context(), capture, storeRoot, authority)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = makeTreeWritable(storeRoot) })
			_, err = Materialize(t.Context(), store, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: t.TempDir()}, authority)
			assertCode(t, err, CodeGraphIncomplete)
			if runner.starts != 1 {
				t.Fatalf("target-pruned unreachable snapshot started install: starts=%d", runner.starts)
			}
		})
	}
}

func TestRealPinnedPNPMTargetPrunedUnreachableRejectsBeforeInstall(t *testing.T) {
	if testing.Short() {
		t.Skip("real pinned pnpm materialization is an integration test")
	}
	fixture := newUnreachableSnapshotDependencyTargetFixture(t, "os", "unsupported-os")
	capture := captureFixture(t, fixture)
	runner := newConcretePNPMRunner(t)
	runner.context = makeExecutionContext(t, capture, runner)
	storeRoot := filepath.Join(t.TempDir(), "store")
	store, err := DerivePrivateStore(t.Context(), capture, storeRoot, runner.context)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTreeWritable(storeRoot) })
	launchesBeforeInstall := len(runner.launches)
	_, err = Materialize(t.Context(), store, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: filepath.Join(t.TempDir(), "work")}, runner.context)
	assertCode(t, err, CodeGraphIncomplete)
	if len(runner.launches) != launchesBeforeInstall {
		t.Fatalf("target-pruned unreachable snapshot started install: launches=%d before=%d", len(runner.launches), launchesBeforeInstall)
	}
}

func TestRealPinnedPNPMLockSupersetSnapshotDependencies(t *testing.T) {
	if testing.Short() {
		t.Skip("real pinned pnpm materialization is an integration test")
	}
	fixture := newTargetPrunedSnapshotDependencyFixture(t)
	capture := captureFixture(t, fixture)
	runner := newConcretePNPMRunner(t)
	runner.context = makeExecutionContext(t, capture, runner)
	storeRoot := filepath.Join(t.TempDir(), "store")
	store, err := DerivePrivateStore(t.Context(), capture, storeRoot, runner.context)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTreeWritable(storeRoot) })
	materialized, err := Materialize(t.Context(), store, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: filepath.Join(t.TempDir(), "work")}, runner.context)
	if err != nil {
		if diagnostic, ok := err.(*Error); ok {
			t.Fatalf("%v fields=%v", err, diagnostic.Fields)
		}
		t.Fatal(err)
	}
	if containsAll(materialized.MaterializedPackages, "optional@1.0.0") {
		t.Fatalf("target-pruned snapshot leaked into active package set: %v", materialized.MaterializedPackages)
	}
}

func TestN04N11N12MaterializationSideEffectsMissingStoreAndAmbientFallbackFailClosed(t *testing.T) {
	t.Run("missing store member", func(t *testing.T) {
		fixture := newPNPMFixture(t)
		capture := captureFixture(t, fixture)
		runner := &fakePNPMRunner{capture: capture}
		authority := makeExecutionContext(t, capture, runner)
		runner.authority = authority
		storeRoot := filepath.Join(t.TempDir(), "store")
		store, err := DerivePrivateStore(context.Background(), capture, storeRoot, authority)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = makeTreeWritable(storeRoot) })
		_ = makeTreeWritable(storeRoot)
		if err = os.Remove(filepath.Join(storeRoot, filepath.FromSlash(store.Receipt.Files[0].Path))); err != nil {
			t.Fatal(err)
		}
		_, err = Materialize(context.Background(), store, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "out"), WorkRoot: t.TempDir()}, authority)
		assertCode(t, err, CodeIntegrityMismatch)
		if runner.starts != 1 {
			t.Fatalf("install started after store drift: %d", runner.starts)
		}
	})
	t.Run("side effect output", func(t *testing.T) {
		fixture := newPNPMFixture(t)
		capture := captureFixture(t, fixture)
		runner := &fakePNPMRunner{capture: capture, emitSideEffects: true}
		authority := makeExecutionContext(t, capture, runner)
		runner.authority = authority
		storeRoot := filepath.Join(t.TempDir(), "store")
		_, err := DerivePrivateStore(context.Background(), capture, storeRoot, authority)
		assertCode(t, err, CodeHookUndeclared)
	})
	t.Run("ambient cannot satisfy", func(t *testing.T) {
		fixture := newPNPMFixture(t)
		capture := captureFixture(t, fixture)
		runner := &fakePNPMRunner{capture: capture}
		authority := makeExecutionContext(t, capture, runner)
		runner.authority = authority
		storeRoot := filepath.Join(t.TempDir(), "store")
		store, err := DerivePrivateStore(context.Background(), capture, storeRoot, authority)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = makeTreeWritable(storeRoot) })
		runner.omitPackage = "b@1.0.0"
		ambient := filepath.Join(t.TempDir(), "ambient-store")
		_ = os.MkdirAll(ambient, 0o700)
		_ = os.WriteFile(filepath.Join(ambient, "b"), fixture.payloads["b@1.0.0"], 0o600)
		_, err = Materialize(context.Background(), store, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "out"), WorkRoot: t.TempDir()}, authority)
		assertCode(t, err, CodeGraphIncomplete)
	})
	t.Run("local root drift", func(t *testing.T) {
		fixture := newPNPMFixture(t)
		capture := captureFixture(t, fixture)
		runner := &fakePNPMRunner{capture: capture, mutateInstall: func(root string) error {
			return os.WriteFile(filepath.Join(root, "packages", "cli", "index.js"), []byte("drift\n"), 0o600)
		}}
		authority := makeExecutionContext(t, capture, runner)
		runner.authority = authority
		storeRoot := filepath.Join(t.TempDir(), "store")
		store, err := DerivePrivateStore(context.Background(), capture, storeRoot, authority)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = makeTreeWritable(storeRoot) })
		_, err = Materialize(context.Background(), store, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "out"), WorkRoot: t.TempDir()}, authority)
		assertCode(t, err, CodeIntegrityMismatch)
	})
}

func TestMaterializedVirtualStoreRejectsEveryUnclaimedOrInvalidPackageRoot(t *testing.T) {
	variants := []struct {
		name   string
		mutate func(string) error
		code   string
	}{
		{"metadata-less expected package", func(root string) error {
			return os.Remove(filepath.Join(root, "node_modules", ".pnpm", pnpmSnapshotDirectory("b@1.0.0"), "node_modules", "b", "package.json"))
		}, CodeGraphIncomplete},
		{"malformed expected metadata", func(root string) error {
			return os.WriteFile(filepath.Join(root, "node_modules", ".pnpm", pnpmSnapshotDirectory("b@1.0.0"), "node_modules", "b", "package.json"), []byte("{"), 0o600)
		}, CodeMetadataMismatch},
		{"unclaimed metadata-less entry", func(root string) error {
			path := filepath.Join(root, "node_modules", ".pnpm", "unclaimed@1.0.0", "node_modules", "unclaimed")
			if err := os.MkdirAll(path, 0o700); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(path, "index.js"), []byte("module.exports = 1\n"), 0o600)
		}, CodeGraphIncomplete},
		{"unclaimed valid package entry", func(root string) error {
			path := filepath.Join(root, "node_modules", ".pnpm", "rogue@1.0.0", "node_modules", "rogue")
			if err := os.MkdirAll(path, 0o700); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(path, "package.json"), []byte(`{"name":"rogue","version":"1.0.0"}`), 0o600)
		}, CodeGraphIncomplete},
	}
	for _, variant := range variants {
		t.Run(variant.name, func(t *testing.T) {
			fixture := newPNPMFixture(t)
			capture := captureFixture(t, fixture)
			runner := &fakePNPMRunner{capture: capture, mutateInstall: variant.mutate}
			authority := makeExecutionContext(t, capture, runner)
			runner.authority = authority
			storeRoot := filepath.Join(t.TempDir(), "store")
			store, err := DerivePrivateStore(context.Background(), capture, storeRoot, authority)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = makeTreeWritable(storeRoot) })
			_, err = Materialize(context.Background(), store, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "out"), WorkRoot: t.TempDir()}, authority)
			assertCode(t, err, variant.code)
		})
	}
}

func TestMaterializedImporterLayoutsRejectEveryUnclaimedOrMiswiredMember(t *testing.T) {
	variants := []struct {
		name   string
		mutate func(string) error
	}{
		{"metadata-less root package", func(root string) error {
			return os.Mkdir(filepath.Join(root, "node_modules", "rogue"), 0o700)
		}},
		{"valid root package", func(root string) error {
			path := filepath.Join(root, "node_modules", "rogue")
			if err := os.Mkdir(path, 0o700); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(path, "package.json"), []byte(`{"name":"rogue","version":"1.0.0"}`), 0o600)
		}},
		{"regular root member", func(root string) error {
			return os.WriteFile(filepath.Join(root, "node_modules", "rogue"), []byte("not a link\n"), 0o600)
		}},
		{"workspace unclaimed member", func(root string) error {
			return os.WriteFile(filepath.Join(root, "packages", "cli", "node_modules", "rogue"), []byte("not a link\n"), 0o600)
		}},
		{"missing workspace direct link", func(root string) error {
			return os.Remove(filepath.Join(root, "packages", "cli", "node_modules", "b"))
		}},
		{"wrong root direct link", func(root string) error {
			link := filepath.Join(root, "node_modules", "a")
			if err := os.Remove(link); err != nil {
				return err
			}
			wrong := filepath.Join(root, "node_modules", ".pnpm", pnpmSnapshotDirectory("b@1.0.0"), "node_modules", "b")
			return os.Symlink(wrong, link)
		}},
		{"malformed manager metadata", func(root string) error {
			return os.WriteFile(filepath.Join(root, "node_modules", ".modules.yaml"), []byte("---\nlayoutVersion: 5\n---\nrogue: true\n"), 0o600)
		}},
	}
	for _, variant := range variants {
		t.Run(variant.name, func(t *testing.T) {
			fixture := newPNPMFixture(t)
			capture := captureFixture(t, fixture)
			_, err := materializeFixture(t, capture, variant.mutate)
			assertCode(t, err, CodeGraphIncomplete)
		})
	}
}

func TestUnsupportedPNPMManagerVersionIsZeroStart(t *testing.T) {
	fixture := newPNPMFixture(t)
	capture := captureFixture(t, fixture)
	runner := &fakePNPMRunner{capture: capture}
	authority := makeExecutionContext(t, capture, runner)
	runner.authority = authority
	authority.Runtime.Manager.VersionOutput = "10.32.1"
	_, err := DerivePrivateStore(t.Context(), capture, filepath.Join(t.TempDir(), "store"), authority)
	assertCode(t, err, CodeRuntimeIdentityChanged)
	if runner.starts != 0 {
		t.Fatalf("unsupported pnpm release started %d processes", runner.starts)
	}
}

func TestWritableStoreOverlayAllowsOnlyExactProjectRegistration(t *testing.T) {
	makeFixture := func(t *testing.T) (string, string, []StoreFile) {
		t.Helper()
		store := filepath.Join(t.TempDir(), "store")
		project := filepath.Join(t.TempDir(), "project")
		if err := os.MkdirAll(filepath.Join(store, "v10", "files"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(project, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(store, "v10", "files", "content"), []byte("source\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		expected, err := inventoryStore(store)
		if err != nil {
			t.Fatal(err)
		}
		registry := filepath.Join(store, "v10", "projects")
		if err = os.MkdirAll(registry, 0o700); err != nil {
			t.Fatal(err)
		}
		if err = os.Symlink(project, filepath.Join(registry, "project-hash")); err != nil {
			t.Fatal(err)
		}
		return store, project, expected
	}
	t.Run("exact registration", func(t *testing.T) {
		store, project, expected := makeFixture(t)
		if err := reconcileWritableStoreOverlay(store, project, expected); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("unclaimed registry member", func(t *testing.T) {
		store, project, expected := makeFixture(t)
		registry := filepath.Join(store, "v10", "projects")
		if err := os.WriteFile(filepath.Join(registry, "rogue"), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		assertCode(t, reconcileWritableStoreOverlay(store, project, expected), CodeInputUndeclared)
	})
	t.Run("frozen content drift", func(t *testing.T) {
		store, project, expected := makeFixture(t)
		if err := os.WriteFile(filepath.Join(store, "v10", "files", "content"), []byte("drift\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		assertCode(t, reconcileWritableStoreOverlay(store, project, expected), CodeIntegrityMismatch)
	})
}

func TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization(t *testing.T) {
	if testing.Short() {
		t.Skip("real pinned pnpm materialization is an integration test")
	}
	fixture := newHostPNPMFixture(t)
	capture := captureFixture(t, fixture)
	runner := newConcretePNPMRunner(t)
	runner.context = makeExecutionContext(t, capture, runner)
	storeRoot := filepath.Join(t.TempDir(), "store")
	store, err := DerivePrivateStore(t.Context(), capture, storeRoot, runner.context)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTreeWritable(storeRoot) })
	materialized, err := Materialize(t.Context(), store, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: filepath.Join(t.TempDir(), "work")}, runner.context)
	if err != nil {
		if diagnostic, ok := err.(*Error); ok {
			t.Fatalf("%v fields=%v\nlaunches=%+v", err, diagnostic.Fields, runner.launches)
		}
		t.Fatalf("%v\nlaunches=%+v", err, runner.launches)
	}
	if !containsAll(materialized.MaterializedPackages, "a@1.0.0(peer@2.0.0)", "b@1.0.0", "optional@1.0.0", "peer@2.0.0") || len(materialized.MaterializedPackages) != 4 {
		t.Fatalf("real pinned pnpm materialized unexpected snapshots: %v", materialized.MaterializedPackages)
	}
	if _, err = Invoke(t.Context(), materialized, "index.js", nil, runner.context); err != nil {
		t.Fatal(err)
	}
	if len(runner.launches) < 3 {
		t.Fatalf("real pinned pnpm path observed %d process launches", len(runner.launches))
	}
	managerLaunches := 0
	for _, launch := range runner.launches {
		if launch.Executable != runner.nodePath {
			t.Fatalf("real pnpm launch bypassed exact C0 Node runtime: %+v", launch)
		}
		for _, entry := range launch.Environment {
			if strings.HasPrefix(entry, "PATH=") {
				t.Fatalf("real pnpm launch retained ambient PATH fallback: %+v", launch)
			}
		}
		if len(launch.Argv) > 0 && strings.HasSuffix(launch.Argv[0], "pnpm.cjs") {
			managerLaunches++
			if filepath.Clean(filepath.Join(launch.CWD, filepath.FromSlash(launch.Argv[0]))) != runner.pnpmPath {
				t.Fatalf("real pnpm entry point differs from C0 binding: %+v", launch)
			}
		}
	}
	if managerLaunches != 2 {
		t.Fatalf("real pinned path observed %d pnpm operations, want 2", managerLaunches)
	}
}

func TestExactPeerContextInstancesAndLinksAreReconciled(t *testing.T) {
	t.Run("two contexts", func(t *testing.T) {
		fixture := newTwoPeerContextFixture(t)
		capture := captureFixture(t, fixture)
		materialized, _ := materializeFixture(t, capture, nil)
		if len(materialized.MaterializedPackages) != 5 || !containsAll(materialized.MaterializedPackages, "a@1.0.0(peer@2.0.0)", "a@1.0.0(peer@2.1.0)") {
			t.Fatalf("exact peer contexts were not retained: %v", materialized.MaterializedPackages)
		}
	})
	for _, variant := range []struct {
		name   string
		mutate func(string) error
	}{
		{"missing peer link", func(root string) error {
			return os.Remove(filepath.Join(root, "node_modules", ".pnpm", pnpmSnapshotDirectory("a@1.0.0(peer@2.1.0)"), "node_modules", "peer"))
		}},
		{"swapped peer link", func(root string) error {
			link := filepath.Join(root, "node_modules", ".pnpm", pnpmSnapshotDirectory("a@1.0.0(peer@2.1.0)"), "node_modules", "peer")
			if err := os.Remove(link); err != nil {
				return err
			}
			wrong := filepath.Join(root, "node_modules", ".pnpm", pnpmSnapshotDirectory("peer@2.0.0"), "node_modules", "peer")
			return os.Symlink(wrong, link)
		}},
	} {
		t.Run(variant.name, func(t *testing.T) {
			fixture := newTwoPeerContextFixture(t)
			capture := captureFixture(t, fixture)
			_, err := materializeFixture(t, capture, variant.mutate)
			assertCode(t, err, CodeGraphIncomplete)
		})
	}
}

type fakePNPMRunner struct {
	capture                                *Capture
	authority                              *ExecutionContext
	starts                                 int
	emitSideEffects                        bool
	omitPackage                            string
	mutateInstall                          func(string) error
	storePermit, installPermit, nodePermit closureexec.DerivationPermit
}

type concretePNPMRunner struct {
	*closureexec.ManagerProcessRunner
	context                    *ExecutionContext
	executionRoot              string
	pnpmPath, nodePath         string
	pnpmRelative, nodeRelative string
	pnpmDigest, nodeDigest     closuregraph.ID
	managerFingerprint         closuregraph.ID
	managerFiles               []StoreFile
	launches                   []closureexec.ProcessLaunch
}

func newConcretePNPMRunner(t *testing.T) *concretePNPMRunner {
	t.Helper()
	pnpmCommand, err := exec.LookPath("pnpm")
	if err != nil {
		t.Skip("pinned pnpm executable unavailable")
	}
	versionOutput, err := exec.Command("node", pnpmCommand, "--version").CombinedOutput() // #nosec G204 -- integration test resolves the explicitly selected task-local pnpm path.
	if err != nil {
		t.Fatalf("read pnpm version: %v: %s", err, versionOutput)
	}
	if strings.TrimSpace(string(versionOutput)) != SupportedPNPMVersion {
		t.Skipf("pnpm %s is outside pinned profile %s", strings.TrimSpace(string(versionOutput)), SupportedPNPMVersion)
	}
	pnpmPath, err := filepath.EvalSymlinks(pnpmCommand)
	if err != nil {
		t.Fatal(err)
	}
	pnpmRoot := filepath.Dir(filepath.Dir(pnpmPath))
	executionRoot := filepath.Join(t.TempDir(), "execution")
	for _, dir := range []string{"bin", "work", "output"} {
		if err = os.MkdirAll(filepath.Join(executionRoot, dir), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	stagedPNPMRoot := filepath.Join(executionRoot, "toolchain", "pnpm")
	if err = copyContainedTree(pnpmRoot, stagedPNPMRoot); err != nil {
		t.Fatal(err)
	}
	pnpmLeaf, err := filepath.Rel(pnpmRoot, pnpmPath)
	if err != nil {
		t.Fatal(err)
	}
	stagedPNPM := filepath.Join(stagedPNPMRoot, pnpmLeaf)
	pnpmPayload, err := os.ReadFile(stagedPNPM) // #nosec G304 -- exact staged manager entry point selected for C0.
	if err != nil {
		t.Fatal(err)
	}
	managerFiles, err := inventoryStore(stagedPNPMRoot)
	if err != nil {
		t.Fatal(err)
	}
	managerFileValues := make([]any, len(managerFiles))
	for index, file := range managerFiles {
		managerFileValues[index] = map[string]any{"path": file.Path, "sha256": string(file.SHA256), "size": file.Size}
	}
	nodeCommand, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	nodePath, err := filepath.EvalSymlinks(nodeCommand)
	if err != nil {
		t.Fatal(err)
	}
	nodeRoot := filepath.Dir(filepath.Dir(nodePath))
	nodeLeaf, err := filepath.Rel(nodeRoot, nodePath)
	if err != nil {
		t.Fatal(err)
	}
	nodePayload, err := os.ReadFile(nodePath) // #nosec G304 -- exact selected Node runtime for integration evidence.
	if err != nil {
		t.Fatal(err)
	}
	stagedNode := filepath.Join(executionRoot, "toolchain", "node", nodeLeaf)
	if err = os.MkdirAll(filepath.Dir(stagedNode), 0o700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(stagedNode, nodePayload, 0o500); err != nil {
		t.Fatal(err)
	}
	linkedNodeLibraries, err := filepath.Glob(filepath.Join(nodeRoot, "lib", "libnode*.dylib"))
	if err != nil {
		t.Fatal(err)
	}
	for _, library := range linkedNodeLibraries {
		payload, readErr := os.ReadFile(library) // #nosec G304 -- adjacent selected Node runtime library.
		if readErr != nil {
			t.Fatal(readErr)
		}
		target := filepath.Join(executionRoot, "toolchain", "node", "lib", filepath.Base(library))
		if err = os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(target, payload, 0o400); err != nil {
			t.Fatal(err)
		}
	}
	pnpmDigest := digestID(pnpmPayload)
	nodeDigest := digestID(nodePayload)
	managerFingerprint, err := closuregraph.DomainID("pnpm-interpreted-toolchain-v1", map[string]any{"manager_version": SupportedPNPMVersion, "entrypoint_sha256": string(pnpmDigest), "node_sha256": string(nodeDigest), "manager_files": managerFileValues})
	if err != nil {
		t.Fatal(err)
	}
	manager, err := closureexec.NewManagerProcessRunner(executionRoot, filepath.Join(executionRoot, "output"))
	if err != nil {
		t.Fatal(err)
	}
	runner := &concretePNPMRunner{ManagerProcessRunner: manager, executionRoot: executionRoot, pnpmPath: stagedPNPM, nodePath: stagedNode, pnpmRelative: filepath.ToSlash(filepath.Join("toolchain", "pnpm", pnpmLeaf)), nodeRelative: filepath.ToSlash(filepath.Join("toolchain", "node", nodeLeaf)), pnpmDigest: pnpmDigest, nodeDigest: nodeDigest, managerFingerprint: managerFingerprint, managerFiles: managerFiles}
	manager.ProcessLaunchObserver = func(launch closureexec.ProcessLaunch) { runner.launches = append(runner.launches, launch) }
	return runner
}

func (runner *concretePNPMRunner) RecheckTool(_ context.Context, tool nodesource.ToolIdentity) (closureexec.ToolchainIdentity, error) {
	nodePayload, err := os.ReadFile(runner.nodePath) // #nosec G304 -- exact staged Node runtime below the task toolchain.
	if err != nil || digestID(nodePayload) != runner.nodeDigest {
		return closureexec.ToolchainIdentity{}, fmt.Errorf("selected Node runtime changed: %w", err)
	}
	fingerprint := runner.nodeDigest
	if tool.Role == "package-manager" {
		pnpmPayload, readErr := os.ReadFile(runner.pnpmPath) // #nosec G304 -- exact staged pnpm entry point below the task toolchain.
		files, inventoryErr := inventoryStore(filepath.Dir(filepath.Dir(runner.pnpmPath)))
		if readErr != nil || inventoryErr != nil || digestID(pnpmPayload) != runner.pnpmDigest || !equalStoreFiles(files, runner.managerFiles) {
			return closureexec.ToolchainIdentity{}, fmt.Errorf("selected pnpm toolchain changed")
		}
		fingerprint = runner.managerFingerprint
	}
	return closureexec.ToolchainIdentity{Fingerprint: fingerprint, ExecutableSHA256: runner.nodeDigest}, nil
}

func (runner *concretePNPMRunner) Run(ctx context.Context, request closureexec.ExecutionRequest) (closureexec.PortableRunResult, error) {
	// Each manager operation owns a fresh evidence root. The shared portable
	// runner intentionally retains the prior receipt bytes after observation,
	// so this integration harness rotates that exact task-local root between
	// already-receipted operations.
	if err := os.RemoveAll(runner.OutputRoot); err != nil {
		return closureexec.PortableRunResult{}, err
	}
	if err := os.MkdirAll(runner.OutputRoot, 0o700); err != nil {
		return closureexec.PortableRunResult{}, err
	}
	return runner.ManagerProcessRunner.Run(ctx, request)
}

func (runner *fakePNPMRunner) RecheckTool(_ context.Context, tool nodesource.ToolIdentity) (closureexec.ToolchainIdentity, error) {
	return closureexec.ToolchainIdentity{Fingerprint: tool.Fingerprint, ExecutableSHA256: tool.ExecutableSHA256}, nil
}
func (runner *fakePNPMRunner) Run(_ context.Context, request closureexec.ExecutionRequest) (closureexec.PortableRunResult, error) {
	runner.starts++
	if err := prepareFakeExecution(request, runner.authority.ExecutionRoot); err != nil {
		return closureexec.PortableRunResult{}, err
	}
	permit := request.Permit
	cwd := filepath.Join(runner.authority.ExecutionRoot, filepath.FromSlash(permit.CWD))
	if len(permit.Argv) > 5 && permit.Argv[4] == "store" && permit.Argv[5] == "add" {
		runner.storePermit = permit
		store := filepath.Clean(filepath.Join(cwd, filepath.FromSlash(argAfter(permit.Argv, "--store-dir"))))
		if err := os.MkdirAll(filepath.Join(store, "v10", "files"), 0o700); err != nil {
			return closureexec.PortableRunResult{}, err
		}
		for _, pkg := range runner.capture.Graph.Packages {
			payload := runner.capture.tarballs[pkg.Key].digest
			if err := os.WriteFile(filepath.Join(store, "v10", "files", strings.ReplaceAll(pkg.Key, "/", "_")+".json"), []byte(payload), 0o600); err != nil {
				return closureexec.PortableRunResult{}, err
			}
			integrity, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(pkg.Integrity, "sha512-"))
			if err != nil || len(integrity) != sha512.Size {
				return closureexec.PortableRunResult{}, fmt.Errorf("invalid fixture integrity for %s", pkg.Key)
			}
			hexDigest := fmt.Sprintf("%x", integrity)[:64]
			index := filepath.Join(store, "v10", "index", hexDigest[:2], hexDigest[2:]+"-file+fixture.json")
			if err := os.MkdirAll(filepath.Dir(index), 0o700); err != nil {
				return closureexec.PortableRunResult{}, err
			}
			if err := os.WriteFile(index, []byte(`{"files":{}}`), 0o600); err != nil {
				return closureexec.PortableRunResult{}, err
			}
		}
		if runner.emitSideEffects {
			if err := os.WriteFile(filepath.Join(store, "v10", "files", "pkg-side-effects.json"), []byte("{}"), 0o600); err != nil {
				return closureexec.PortableRunResult{}, err
			}
		}
	}
	if len(permit.Argv) > 1 && permit.Argv[1] == "install" {
		runner.installPermit = permit
		roots := map[string]string{}
		for _, snapshot := range runner.capture.Graph.Snapshots {
			if snapshot.PackageKey == runner.omitPackage {
				continue
			}
			target := filepath.Join(cwd, "node_modules", ".pnpm", pnpmSnapshotDirectory(snapshot.Key), "node_modules", filepath.FromSlash(snapshot.Name))
			roots[snapshot.Key] = target
			if err := writeExpectedPackage(runner.capture.tarballs[snapshot.PackageKey], target); err != nil {
				return closureexec.PortableRunResult{}, err
			}
		}
		for _, edge := range runner.capture.Graph.Edges {
			if edge.To == "" || roots[edge.From] == "" {
				continue
			}
			target := roots[edge.To]
			if strings.HasPrefix(edge.To, "local:") {
				target = filepath.Join(cwd, filepath.FromSlash(strings.TrimPrefix(edge.To, "local:")))
			}
			if target == "" {
				continue
			}
			link := filepath.Join(filepath.Dir(roots[edge.From]), filepath.FromSlash(edge.Name))
			if _, err := os.Lstat(link); err == nil {
				continue
			}
			if err := os.MkdirAll(filepath.Dir(link), 0o700); err != nil {
				return closureexec.PortableRunResult{}, err
			}
			if err := os.Symlink(target, link); err != nil {
				return closureexec.PortableRunResult{}, err
			}
		}
		if err := os.WriteFile(filepath.Join(cwd, "node_modules", ".modules.yaml"), []byte("layoutVersion: 5\n"), 0o600); err != nil {
			return closureexec.PortableRunResult{}, err
		}
		if err := os.WriteFile(filepath.Join(cwd, "node_modules", ".pnpm-workspace-state-v1.json"), []byte("{}\n"), 0o600); err != nil {
			return closureexec.PortableRunResult{}, err
		}
		for _, local := range runner.capture.Graph.LocalRoots {
			owner := cwd
			if local.Path != "." {
				owner = filepath.Join(cwd, filepath.FromSlash(local.Path))
			}
			for _, edge := range runner.capture.Graph.Edges {
				if edge.From != localRootKey(local.Path) || !edge.Selected {
					continue
				}
				target := roots[edge.To]
				if strings.HasPrefix(edge.To, "local:") {
					path := strings.TrimPrefix(edge.To, "local:")
					target = cwd
					if path != "." {
						target = filepath.Join(cwd, filepath.FromSlash(path))
					}
				}
				link := filepath.Join(owner, "node_modules", filepath.FromSlash(edge.Name))
				if err := os.MkdirAll(filepath.Dir(link), 0o700); err != nil {
					return closureexec.PortableRunResult{}, err
				}
				if err := os.Symlink(target, link); err != nil {
					return closureexec.PortableRunResult{}, err
				}
			}
		}
		if runner.mutateInstall != nil {
			if err := runner.mutateInstall(cwd); err != nil {
				return closureexec.PortableRunResult{}, err
			}
		}
	}
	if len(permit.Argv) > 0 && permit.Argv[0] == "index.js" {
		runner.nodePermit = permit
	}
	return portableResult(runner.authority.ExecutionRoot, permit)
}

func writeExpectedPackage(item capturedBlob, destination string) error {
	if err := os.MkdirAll(destination, 0o700); err != nil {
		return err
	}
	executable := map[string]bool{}
	for _, file := range item.files {
		executable[file.Path] = file.Executable
	}
	for name, payload := range item.contents {
		target := filepath.Join(destination, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		mode := os.FileMode(0o600)
		if executable[name] {
			mode = 0o700
		}
		if err := os.WriteFile(target, payload, mode); err != nil {
			return err
		}
	}
	return nil
}

func makeExecutionContext(t *testing.T, capture *Capture, runner interface {
	closureexec.PortableProcessRunner
	RecheckTool(context.Context, nodesource.ToolIdentity) (closureexec.ToolchainIdentity, error)
}) *ExecutionContext {
	t.Helper()
	executionRoot := ""
	if concrete, ok := runner.(*concretePNPMRunner); ok {
		executionRoot = concrete.executionRoot
	} else {
		executionRoot = filepath.Join(t.TempDir(), "execution")
		for _, dir := range []string{"bin", "work", "output"} {
			if err := os.MkdirAll(filepath.Join(executionRoot, dir), 0o700); err != nil {
				t.Fatal(err)
			}
		}
	}
	target := capture.Graph.Target
	platform := closuregraph.TargetPlatformPayload{OS: target.OS, Architecture: target.Architecture, ABI: "node", Libc: target.Libc, MinimumRuntime: "bound", SDKID: "none", TargetTriple: target.Architecture + "-" + target.OS, Runtime: "node", LanguageModes: map[string]string{"package_manager": "pnpm"}, Tuning: map[string]string{}}
	platformNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: platform}
	platformID, err := platformNode.ID()
	if err != nil {
		t.Fatal(err)
	}
	selection, err := closuregraph.NewSelectionContext(capture.NodeCapture.Graph.RootNodeIDs, map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID}, []string{}, target.IncludeDev, map[string]string{"os": target.OS, "cpu": target.Architecture, "libc": target.Libc}, map[string]string{}, []string{"pnpm-platform-selector-v1"})
	if err != nil {
		t.Fatal(err)
	}
	tool := func(role, seed string) nodesource.ToolIdentity {
		version := role + " fixture"
		if role == "package-manager" {
			version = SupportedPNPMVersion
		}
		return nodesource.ToolIdentity{Role: role, PolicySelector: role + "-v1", ExecutableRelativePath: "bin/node", VersionOutput: version, PlatformABI: target.OS + "-" + target.Architecture, Fingerprint: digestID([]byte(seed + "-tree")), ExecutableSHA256: digestID([]byte("node-executable")), ExecutionDomain: closuregraph.ExecutionTarget}
	}
	runtime := nodesource.RuntimeBinding{Platform: platform, Node: tool("node-runtime", "node"), Manager: tool("package-manager", "pnpm"), TargetNodeIDs: append([]closuregraph.ID(nil), capture.NodeCapture.Graph.RootNodeIDs...)}
	runtime.Node.ReadRoots = []string{"toolchain/node"}
	runtime.Manager.EntrypointRelativePath = "toolchain/pnpm/bin/pnpm.cjs"
	runtime.Manager.ReadRoots = []string{"toolchain/node", "toolchain/pnpm"}
	if concrete, ok := runner.(*concretePNPMRunner); ok {
		runtime.Node.ExecutableRelativePath = concrete.nodeRelative
		runtime.Node.ExecutableSHA256 = concrete.nodeDigest
		runtime.Node.Fingerprint = concrete.nodeDigest
		runtime.Manager.ExecutableRelativePath = concrete.nodeRelative
		runtime.Manager.ExecutableSHA256 = concrete.nodeDigest
		runtime.Manager.EntrypointRelativePath = concrete.pnpmRelative
		runtime.Manager.Fingerprint = concrete.managerFingerprint
		runtime.Manager.VersionOutput = SupportedPNPMVersion
	}
	c0, err := nodesource.NewC0Checkpoint(capture.NodeCapture, selection, runtime)
	if err != nil {
		t.Fatal(err)
	}
	runtime.C0Checkpoint = &c0
	evaluator := closuregraph.ConditionEvaluatorFunc{EvaluatorID: "pnpm-platform-selector-v1", EvaluateFunc: evaluatePlatform}
	_, plan, err := nodesource.Close(capture.NodeCapture, selection, runtime, []closuregraph.ConditionEvaluator{evaluator}, closureexec.PortableExecutionPolicyID)
	if err != nil {
		t.Fatal(err)
	}
	executor, err := closureexec.NewAssuredExecutor(closureexec.DefaultAssuranceConfig(), runner, nil, "test-head")
	if err != nil {
		t.Fatal(err)
	}
	return &ExecutionContext{Executor: executor, Selection: selection, Runtime: runtime, BuildPlan: plan, Recheck: runner.RecheckTool, ExecutionRoot: executionRoot}
}
func prepareFakeExecution(request closureexec.ExecutionRequest, root string) error {
	inputs := map[closuregraph.ID]closureexec.ReplayInput{}
	for _, input := range request.Inputs {
		inputs[input.ReceiptID] = input
		source, err := input.ProtectedPath()
		if err != nil {
			return err
		}
		target := filepath.Join(root, filepath.FromSlash(input.MountPath))
		_ = os.RemoveAll(target)
		if input.IsTree() {
			if err = copyWritableTree(source, target); err != nil {
				return err
			}
		} else {
			payload, err := os.ReadFile(source)
			if err != nil {
				return err
			}
			if err = os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				return err
			}
			if err = os.WriteFile(target, payload, 0o600); err != nil {
				return err
			}
		}
	}
	for _, work := range request.Permit.WorkCopies {
		source, err := inputs[work.ReceiptID].ProtectedPath()
		if err != nil {
			return err
		}
		target := filepath.Join(root, filepath.FromSlash(work.Path))
		_ = os.RemoveAll(target)
		if err = copyWritableTree(source, target); err != nil {
			return err
		}
	}
	output := filepath.Join(root, "output")
	_ = os.RemoveAll(output)
	return os.MkdirAll(output, 0o700)
}
func portableResult(root string, permit closureexec.DerivationPermit) (closureexec.PortableRunResult, error) {
	output := filepath.Join(root, filepath.FromSlash(permit.Environment["CURATOR_OUTPUT_ROOT"]))
	for _, expected := range permit.ExpectedEvidence {
		target := filepath.Join(output, filepath.FromSlash(expected.Path))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return closureexec.PortableRunResult{}, err
		}
		if err := os.WriteFile(target, nil, 0o600); err != nil {
			return closureexec.PortableRunResult{}, err
		}
	}
	return closureexec.PortableRunResult{ExitCode: 0, OutputRoot: output}, nil
}
func containsAll(values []string, wants ...string) bool {
	for _, want := range wants {
		found := false
		for _, value := range values {
			found = found || value == want
		}
		if !found {
			return false
		}
	}
	return true
}
func argAfter(values []string, name string) string {
	for index, value := range values {
		if value == name && index+1 < len(values) {
			return values[index+1]
		}
	}
	return ""
}

func captureFixture(t *testing.T, fixture pnpmFixture) *Capture {
	t.Helper()
	capture, err := captureFixtureError(t, fixture)
	if err != nil {
		t.Fatal(err)
	}
	return capture
}
func captureFixtureError(t *testing.T, fixture pnpmFixture) (*Capture, error) {
	t.Helper()
	storeRoot := filepath.Join(t.TempDir(), "capture-store")
	store, err := closureexec.NewCaptureStore(storeRoot)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTreeWritable(storeRoot) })
	return CaptureAndAdmit(context.Background(), CaptureRequest{Graph: fixture.graph, ProjectRoot: fixture.root, Tarballs: fixture.tarballs, WorkRoot: fixture.work, Store: store, Policy: artifactpolicy.NewService(), PreviousCausalHead: "test-head"})
}
func replacePackage(t *testing.T, fixture *pnpmFixture, key string, pkg fixturePackage) {
	t.Helper()
	payload := buildTGZ(t, pkg)
	sum := sha512.Sum512(payload)
	integrity := "sha512-" + base64.StdEncoding.EncodeToString(sum[:])
	fixture.request.LockBytes = []byte(strings.Replace(string(fixture.request.LockBytes), fixture.graph.Packages[fixture.graph.packageIndex[key]].Integrity, integrity, 1))
	fixture.graph = mustParse(t, fixture.request)
	file := filepath.Join(t.TempDir(), "replacement.tgz")
	_ = os.WriteFile(file, payload, 0o600)
	fixture.tarballs = cloneTarballs(fixture.tarballs)
	fixture.tarballs[key] = RawTarball{Path: file}
	_ = os.WriteFile(filepath.Join(fixture.root, "pnpm-lock.yaml"), fixture.request.LockBytes, 0o600)
}
func mustParse(t *testing.T, request ParseRequest) Graph {
	t.Helper()
	graph, err := Parse(request)
	if err != nil {
		t.Fatal(err)
	}
	return graph
}
func buildTGZ(t *testing.T, pkg fixturePackage) []byte {
	t.Helper()
	metadata := map[string]any{"name": pkg.name, "version": pkg.version}
	for key, value := range pkg.metadata {
		metadata[key] = value
	}
	files := map[string][]byte{"package.json": mustJSON(t, metadata)}
	for key, value := range pkg.files {
		files[key] = value
	}
	var output bytes.Buffer
	gz := gzip.NewWriter(&output)
	tw := tar.NewWriter(gz)
	keys := sortedKeys(files)
	for _, name := range keys {
		payload := files[name]
		if err := tw.WriteHeader(&tar.Header{Name: "package/" + name, Mode: 0o644, Size: int64(len(payload))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(payload); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}
func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}
func assertCode(t *testing.T, err error, want string) {
	t.Helper()
	if ErrorCode(err) != want {
		t.Fatalf("code=%q want=%q err=%v", ErrorCode(err), want, err)
	}
}
func cloneBytesMap(values map[string][]byte) map[string][]byte {
	result := map[string][]byte{}
	for key, value := range values {
		result[key] = append([]byte(nil), value...)
	}
	return result
}
func cloneTarballs(values map[string]RawTarball) map[string]RawTarball {
	result := map[string]RawTarball{}
	for key, value := range values {
		result[key] = value
	}
	return result
}
func makeTreeWritable(root string) error {
	return filepath.WalkDir(root, func(current string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.Chmod(current, 0o700)
		}
		return os.Chmod(current, 0o600)
	})
}
