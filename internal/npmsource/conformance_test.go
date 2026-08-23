package npmsource

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

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

type npmFixture struct {
	root, work string
	request    ParseRequest
	tarballs   map[string]RawTarball
	payloads   map[string][]byte
	graph      Graph
}

func newNPMFixture(t *testing.T) npmFixture {
	t.Helper()
	root := filepath.Join(t.TempDir(), "project")
	work := filepath.Join(t.TempDir(), "work")
	if err := os.MkdirAll(filepath.Join(root, "packages", "workspace"), 0o700); err != nil {
		t.Fatal(err)
	}
	rootManifest := map[string]any{"name": "app", "version": "1.0.0", "workspaces": []string{"packages/*"}, "dependencies": map[string]string{"a": "1.0.0"}, "optionalDependencies": map[string]string{"opt": "1.0.0"}}
	workspaceManifest := map[string]any{"name": "workspace", "version": "1.0.0"}
	packages := map[string]fixturePackage{
		"node_modules/a":   {name: "a", version: "1.0.0", metadata: map[string]any{"dependencies": map[string]string{"b": "1.0.0"}}, files: map[string][]byte{"index.js": []byte("module.exports = require('b')\n")}},
		"node_modules/b":   {name: "b", version: "1.0.0", metadata: map[string]any{}, files: map[string][]byte{"index.js": []byte("module.exports = 'b'\n")}},
		"node_modules/opt": {name: "opt", version: "1.0.0", metadata: map[string]any{"os": []string{"linux"}}, files: map[string][]byte{"index.js": []byte("module.exports = 'optional'\n")}},
	}
	tarballs := map[string]RawTarball{}
	payloads := map[string][]byte{}
	lockPackages := map[string]any{
		"":                       map[string]any{"name": "app", "version": "1.0.0", "dependencies": map[string]string{"a": "1.0.0"}, "optionalDependencies": map[string]string{"opt": "1.0.0"}},
		"packages/workspace":     map[string]any{"name": "workspace", "version": "1.0.0"},
		"node_modules/workspace": map[string]any{"resolved": "packages/workspace", "link": true},
	}
	for installPath, pkg := range packages {
		payload := buildTGZ(t, pkg)
		payloads[installPath] = payload
		sum := sha512.Sum512(payload)
		integrity := "sha512-" + base64.StdEncoding.EncodeToString(sum[:])
		resolved := fmt.Sprintf("https://registry.npmjs.org/%s/-/%s-%s.tgz", pkg.name, pkg.name, pkg.version)
		entry := map[string]any{"name": pkg.name, "version": pkg.version, "resolved": resolved, "integrity": integrity}
		for key, value := range pkg.metadata {
			entry[key] = value
		}
		if installPath == "node_modules/opt" {
			entry["optional"] = true
		}
		lockPackages[installPath] = entry
		rawPath := filepath.Join(t.TempDir(), strings.ReplaceAll(pkg.name, "/", "-")+".tgz")
		if err := os.WriteFile(rawPath, payload, 0o600); err != nil {
			t.Fatal(err)
		}
		tarballs[installPath] = RawTarball{Path: rawPath}
	}
	lock := map[string]any{"name": "app", "version": "1.0.0", "lockfileVersion": 3, "packages": lockPackages}
	lockBytes := mustJSON(t, lock)
	rootBytes := mustJSON(t, rootManifest)
	workspaceBytes := mustJSON(t, workspaceManifest)
	if err := os.WriteFile(filepath.Join(root, "package.json"), rootBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "package-lock.json"), lockBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "index.js"), []byte("require('a')\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "packages", "workspace", "package.json"), workspaceBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	request := ParseRequest{LockName: "package-lock.json", LockBytes: lockBytes, Manifests: map[string][]byte{"package.json": rootBytes, "packages/workspace/package.json": workspaceBytes}, AllowedRegistryOrigins: []string{"https://registry.npmjs.org/"}, Target: Target{OS: "darwin", Architecture: "arm64", Libc: "none", IncludeDev: true}}
	graph, err := Parse(request)
	if err != nil {
		t.Fatal(err)
	}
	return npmFixture{root: root, work: work, request: request, tarballs: tarballs, payloads: payloads, graph: graph}
}

func TestS01N01ClosedGraphCaptureAndExactTarballs(t *testing.T) {
	fixture := newNPMFixture(t)
	if fixture.graph.LockfileVersion != 3 || len(fixture.graph.Packages) != 6 || len(fixture.graph.Workspaces) != 1 {
		t.Fatalf("unexpected npm graph: %+v", fixture.graph)
	}
	opt := fixture.graph.Packages[fixture.graph.packageIndex["node_modules/opt"]]
	if opt.Selected || opt.PruneReason != "os_pruned" {
		t.Fatalf("optional target decision not retained: %+v", opt)
	}
	optionalEdges := 0
	for _, edge := range fixture.graph.Edges {
		if edge.From == "" && edge.Name == "opt" {
			optionalEdges++
			if edge.Scope != "optional" {
				t.Fatalf("optional dependency did not override runtime scope: %+v", edge)
			}
		}
	}
	if optionalEdges != 1 {
		t.Fatalf("optional dependency produced %d edges, want one", optionalEdges)
	}
	capture := captureFixture(t, fixture)
	if len(capture.Evidence.Tarballs) != 3 || !capture.Evidence.NodeCaptureGraphID.Valid() {
		t.Fatalf("incomplete capture evidence: %+v", capture.Evidence)
	}
	for _, item := range capture.Evidence.Tarballs {
		if !item.ArtifactManifestID.Valid() || !item.IntakeReceiptID.Valid() || item.Integrity == "" || item.SHA256 == "" {
			t.Fatalf("unbound raw tarball: %+v", item)
		}
	}
	runner := &fakeRunner{graph: fixture.graph, capture: capture}
	authority := makeExecutionContext(t, capture, runner)
	if len(authority.BuildPlan.ActionNodeIDs) != 3 {
		t.Fatalf("npm C5 plan has %d actions, want cache/install/invoke", len(authority.BuildPlan.ActionNodeIDs))
	}
}

func TestS06CanonicalLockOrderAndN10TargetSelection(t *testing.T) {
	fixture := newNPMFixture(t)
	var value map[string]any
	if err := json.Unmarshal(fixture.request.LockBytes, &value); err != nil {
		t.Fatal(err)
	}
	permuted := []byte(`{"version":"1.0.0","packages":` + string(mustJSON(t, value["packages"])) + `,"name":"app","lockfileVersion":3}`)
	request := fixture.request
	request.LockBytes = permuted
	second, err := Parse(request)
	if err != nil {
		t.Fatal(err)
	}
	if second.LockDigest != fixture.graph.LockDigest || second.RawLockSHA256 == fixture.graph.RawLockSHA256 {
		t.Fatalf("semantic/raw lock identities not separated: first=%+v second=%+v", fixture.graph, second)
	}
	request = fixture.request
	request.Target = Target{OS: "linux", Architecture: "arm64", Libc: "glibc", IncludeDev: true}
	linux, err := Parse(request)
	if err != nil {
		t.Fatal(err)
	}
	if !linux.Packages[linux.packageIndex["node_modules/opt"]].Selected {
		t.Fatal("Linux optional package was not selected")
	}
}

func TestSupportedV2V3AndShrinkwrapSchemas(t *testing.T) {
	fixture := newNPMFixture(t)
	var lock map[string]any
	if err := json.Unmarshal(fixture.request.LockBytes, &lock); err != nil {
		t.Fatal(err)
	}
	entries := lock["packages"].(map[string]any)
	legacy := map[string]any{}
	for _, name := range []string{"a", "b", "opt"} {
		entry := entries["node_modules/"+name].(map[string]any)
		item := map[string]any{"version": entry["version"], "resolved": entry["resolved"], "integrity": entry["integrity"]}
		if optional, ok := entry["optional"]; ok {
			item["optional"] = optional
		}
		if dependencies, ok := entry["dependencies"]; ok {
			item["requires"] = dependencies
		}
		legacy[name] = item
	}
	lock["lockfileVersion"] = 2
	lock["dependencies"] = legacy
	request := fixture.request
	request.LockBytes = mustJSON(t, lock)
	graph, err := Parse(request)
	if err != nil || graph.LockfileVersion != 2 {
		t.Fatalf("supported package-lock v2 rejected: graph=%+v err=%v", graph, err)
	}
	request.LockName = "npm-shrinkwrap.json"
	if _, err = Parse(request); err != nil {
		t.Fatalf("supported npm-shrinkwrap v2 rejected: %v", err)
	}
	lock["lockfileVersion"] = 1
	request.LockBytes = mustJSON(t, lock)
	_, err = Parse(request)
	assertCode(t, err, CodeLockFormatUnsupported)
}

func TestV2LegacyTreeDriftAndDuplicateJSONFailClosed(t *testing.T) {
	fixture := newNPMFixture(t)
	var lock map[string]any
	_ = json.Unmarshal(fixture.request.LockBytes, &lock)
	entries := lock["packages"].(map[string]any)
	legacy := map[string]any{}
	for _, name := range []string{"a", "b", "opt"} {
		entry := entries["node_modules/"+name].(map[string]any)
		legacy[name] = map[string]any{"version": entry["version"], "resolved": entry["resolved"], "integrity": entry["integrity"]}
	}
	legacy["a"].(map[string]any)["version"] = "9.9.9"
	lock["lockfileVersion"] = 2
	lock["dependencies"] = legacy
	request := fixture.request
	request.LockBytes = mustJSON(t, lock)
	_, err := Parse(request)
	assertCode(t, err, CodeLockStale)
	request = fixture.request
	request.LockBytes = []byte(`{"name":"app","name":"shadow","lockfileVersion":3,"packages":{}}`)
	_, err = Parse(request)
	assertCode(t, err, CodeLockFormatUnsupported)
}

func TestN11MutableLocatorAndLockBundlingFailClosed(t *testing.T) {
	fixture := newNPMFixture(t)
	var lock map[string]any
	_ = json.Unmarshal(fixture.request.LockBytes, &lock)
	entry := lock["packages"].(map[string]any)["node_modules/a"].(map[string]any)
	entry["resolved"] = "git+https://example.invalid/a.git#main"
	request := fixture.request
	request.LockBytes = mustJSON(t, lock)
	_, err := Parse(request)
	assertCode(t, err, CodeOriginUnpinned)
	fixture = newNPMFixture(t)
	_ = json.Unmarshal(fixture.request.LockBytes, &lock)
	entry = lock["packages"].(map[string]any)["node_modules/a"].(map[string]any)
	entry["resolved"] = "https://unapproved.example/a-1.0.0.tgz"
	request = fixture.request
	request.LockBytes = mustJSON(t, lock)
	_, err = Parse(request)
	assertCode(t, err, CodeOriginUnpinned)
	fixture = newNPMFixture(t)
	_ = json.Unmarshal(fixture.request.LockBytes, &lock)
	entry = lock["packages"].(map[string]any)["node_modules/a"].(map[string]any)
	entry["inBundle"] = true
	request = fixture.request
	request.LockBytes = mustJSON(t, lock)
	_, err = Parse(request)
	assertCode(t, err, CodeBundledDependencyUnsupported)
}

func TestN02StaleLockN03IntegrityAndN09WorkspaceEscape(t *testing.T) {
	t.Run("N02 stale", func(t *testing.T) {
		fixture := newNPMFixture(t)
		var manifest map[string]any
		_ = json.Unmarshal(fixture.request.Manifests["package.json"], &manifest)
		manifest["dependencies"].(map[string]any)["a"] = "2.0.0"
		fixture.request.Manifests["package.json"] = mustJSON(t, manifest)
		_, err := Parse(fixture.request)
		assertCode(t, err, CodeLockStale)
	})
	t.Run("N03 integrity", func(t *testing.T) {
		fixture := newNPMFixture(t)
		payload := append([]byte(nil), fixture.payloads["node_modules/a"]...)
		payload[len(payload)-1] ^= 1
		if err := os.WriteFile(fixture.tarballs["node_modules/a"].Path, payload, 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := captureFixtureError(t, fixture)
		assertCode(t, err, CodeIntegrityMismatch)
	})
	t.Run("N09 escape", func(t *testing.T) {
		fixture := newNPMFixture(t)
		fixture.request.Manifests["../package.json"] = fixture.request.Manifests["package.json"]
		_, err := Parse(fixture.request)
		assertCode(t, err, CodeLocalPathEscape)
	})
}

func TestN04LifecycleN05BindingGYPAndN11BundledDependency(t *testing.T) {
	for _, testCase := range []struct {
		name, path string
		mutate     func(*fixturePackage)
		code       string
	}{
		{name: "N04 lifecycle", path: "node_modules/a", mutate: func(pkg *fixturePackage) {
			pkg.metadata["scripts"] = map[string]string{"install": "node install.js"}
			pkg.metadata["hasInstallScript"] = true
		}, code: CodeHookUndeclared},
		{name: "N05 binding", path: "node_modules/a", mutate: func(pkg *fixturePackage) { pkg.files["binding.gyp"] = []byte("{'targets': []}\n") }, code: CodeNativeBuildUnsupported},
		{name: "N11 bundle", path: "node_modules/a", mutate: func(pkg *fixturePackage) {
			pkg.metadata["bundleDependencies"] = []string{"b"}
			pkg.files["node_modules/b/index.js"] = []byte("module.exports=1\n")
		}, code: CodeBundledDependencyUnsupported},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newNPMFixture(t)
			pkg := fixturePackage{name: "a", version: "1.0.0", metadata: map[string]any{"dependencies": map[string]string{"b": "1.0.0"}}, files: map[string][]byte{"index.js": []byte("module.exports=1\n")}}
			testCase.mutate(&pkg)
			replaceTarballAndLock(t, &fixture, testCase.path, pkg)
			graph, err := Parse(fixture.request)
			if err != nil {
				if ErrorCode(err) == testCase.code {
					return
				}
				t.Fatal(err)
			}
			fixture.graph = graph
			_, err = captureFixtureError(t, fixture)
			assertCode(t, err, testCase.code)
		})
	}
}

func TestS05N06CompiledPayloadHasSharedPrimaryDiagnostic(t *testing.T) {
	fixture := newNPMFixture(t)
	pkg := fixturePackage{name: "a", version: "1.0.0", metadata: map[string]any{"dependencies": map[string]string{"b": "1.0.0"}}, files: map[string][]byte{"safe.js": []byte("module.exports=1\n"), "renamed.txt": []byte{0, 0x61, 0x73, 0x6d, 1, 0, 0, 0}}}
	replaceTarballAndLock(t, &fixture, "node_modules/a", pkg)
	graph, err := Parse(fixture.request)
	if err != nil {
		t.Fatal(err)
	}
	fixture.graph = graph
	_, err = captureFixtureError(t, fixture)
	if artifactpolicy.ErrorCode(err) != artifactpolicy.CodeCompiledDependency {
		t.Fatalf("got %v (%s), want shared compiled diagnostic", err, artifactpolicy.ErrorCode(err))
	}
}

func TestS02S03S04S07S08N12N13OfflineMaterializationGates(t *testing.T) {
	t.Run("S01 S08 N13 positive", func(t *testing.T) {
		fixture := newNPMFixture(t)
		if err := os.MkdirAll(filepath.Join(fixture.root, "node_modules", "poison"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(fixture.root, "node_modules", "poison", "package.json"), []byte(`{"name":"poison"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		capture := captureFixture(t, fixture)
		if !reflect.DeepEqual(capture.Evidence.DiscardedDerivedPaths, []string{"node_modules"}) {
			t.Fatalf("preseeded state not discarded: %v", capture.Evidence.DiscardedDerivedPaths)
		}
		runner := &fakeRunner{graph: fixture.graph}
		cache := deriveCacheFixture(t, capture, runner)
		materialized, err := Materialize(context.Background(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: filepath.Join(t.TempDir(), "materialize-work")}, runner.context)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(materialized.MaterializedPackages, []string{"node_modules/a", "node_modules/b", "node_modules/workspace"}) {
			t.Fatalf("unexpected installed graph: %v", materialized.MaterializedPackages)
		}
		if runner.lastCI.Environment["NPM_CONFIG_CACHE"] == "" || !containsAll(runner.lastCI.Argv, "ci", "--offline", "--ignore-scripts") {
			t.Fatalf("npm ci contract is incomplete: %+v", runner.lastCI)
		}
	})
	t.Run("S07 extra package", func(t *testing.T) {
		fixture := newNPMFixture(t)
		capture := captureFixture(t, fixture)
		runner := &fakeRunner{graph: fixture.graph, extraPackage: "node_modules/extra"}
		cache := deriveCacheFixture(t, capture, runner)
		_, err := Materialize(context.Background(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: filepath.Join(t.TempDir(), "work")}, runner.context)
		assertCode(t, err, CodeGraphIncomplete)
	})
	t.Run("N12 missing private cache", func(t *testing.T) {
		fixture := newNPMFixture(t)
		capture := captureFixture(t, fixture)
		runner := &fakeRunner{graph: fixture.graph}
		cache := deriveCacheFixture(t, capture, runner)
		if err := makeTreeWritable(cache.root); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(filepath.Join(cache.root, cache.Receipt.Files[0].Path)); err != nil {
			t.Fatal(err)
		}
		_, err := Materialize(context.Background(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: filepath.Join(t.TempDir(), "work")}, runner.context)
		assertCode(t, err, CodeOfflineInputMissing)
	})
	t.Run("S04 verified network attempt", func(t *testing.T) {
		fixture := newNPMFixture(t)
		capture := captureFixture(t, fixture)
		provider := newVerifiedFixtureProvider()
		recheck := &fakeRunner{graph: fixture.graph, capture: capture}
		authority := makeVerifiedExecutionContext(t, capture, provider, recheck)
		provider.allowExecution = true
		provider.networkAttempt = true
		destination := filepath.Join(t.TempDir(), "cache")
		_, err := DerivePrivateCache(t.Context(), capture, destination, filepath.Join(t.TempDir(), "work"), authority)
		var diagnostic *closureexec.DiagnosticError
		if !errors.As(err, &diagnostic) || diagnostic.Code != CodeNetworkAttempted {
			t.Fatalf("network attempt diagnostic=%v", err)
		}
		if provider.starts != 1 {
			t.Fatalf("network attempt starts=%d, want one denied declared action", provider.starts)
		}
		if _, statErr := os.Lstat(destination); !errors.Is(statErr, fs.ErrNotExist) {
			t.Fatalf("network failure published cache: %v", statErr)
		}
	})
}

func TestN07ShippedGeneratedTextAndOfflineInvocation(t *testing.T) {
	fixture := newNPMFixture(t)
	capture := captureFixture(t, fixture)
	runner := &fakeRunner{graph: fixture.graph}
	cache := deriveCacheFixture(t, capture, runner)
	materialized, err := Materialize(context.Background(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: filepath.Join(t.TempDir(), "work")}, runner.context)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := Invoke(context.Background(), materialized, "index.js", nil, runner.context)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.AssuranceMode != closureexec.AssurancePortable || receipt.Audit.Network != "not-observed" || runner.lastNode.Executable != "bin/node" {
		t.Fatalf("invocation did not retain offline Node authority: %+v", receipt)
	}
}

func TestMaterializedPackagesMustMatchAndReadmitExactTarballBytes(t *testing.T) {
	t.Run("substituted source", func(t *testing.T) {
		fixture := newNPMFixture(t)
		capture := captureFixture(t, fixture)
		runner := &fakeRunner{graph: fixture.graph, mutateInstall: func(root string) error {
			return os.WriteFile(filepath.Join(root, "node_modules", "a", "index.js"), []byte("module.exports='substituted'\n"), 0o600)
		}}
		cache := deriveCacheFixture(t, capture, runner)
		_, err := Materialize(t.Context(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: filepath.Join(t.TempDir(), "work")}, runner.context)
		assertCode(t, err, CodeIntegrityMismatch)
	})
	for _, testCase := range []struct {
		name   string
		code   string
		mutate func(string) error
	}{
		{name: "implicit binding gyp", code: CodeNativeBuildUnsupported, mutate: func(root string) error {
			return os.WriteFile(filepath.Join(root, "node_modules", "a", "binding.gyp"), []byte("{'targets': []}\n"), 0o600)
		}},
		{name: "bundled dependency tree", code: CodeBundledDependencyUnsupported, mutate: func(root string) error {
			target := filepath.Join(root, "node_modules", "a", "node_modules", "evil")
			if err := os.MkdirAll(target, 0o700); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(target, "package.json"), []byte(`{"name":"evil","version":"1.0.0"}`), 0o600)
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newNPMFixture(t)
			capture := captureFixture(t, fixture)
			runner := &fakeRunner{graph: fixture.graph, mutateInstall: testCase.mutate}
			cache := deriveCacheFixture(t, capture, runner)
			_, err := Materialize(t.Context(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: filepath.Join(t.TempDir(), "work")}, runner.context)
			assertCode(t, err, testCase.code)
		})
	}

	wasm := []byte{'\x00', 'a', 's', 'm', '\x01', '\x00', '\x00', '\x00'}
	nested := func() []byte {
		var payload bytes.Buffer
		writer := zip.NewWriter(&payload)
		member, err := writer.Create("nested/module.wasm")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = member.Write(wasm); err != nil {
			t.Fatal(err)
		}
		if err = writer.Close(); err != nil {
			t.Fatal(err)
		}
		return payload.Bytes()
	}()
	for _, testCase := range []struct {
		name, path string
		code       artifactpolicy.DiagnosticCode
		payload    []byte
	}{
		{name: "direct native addon", path: "addon.node", payload: wasm, code: artifactpolicy.CodeCompiledDependency},
		{name: "renamed compiled payload", path: "notes.txt", payload: wasm, code: artifactpolicy.CodeCompiledDependency},
		{name: "nested compiled payload", path: "nested.zip", payload: nested, code: artifactpolicy.CodeCompiledDependency},
		{name: "opaque payload", path: "payload.bin", payload: []byte{0x13, 0x37, 0x00, 0xff, 0x80, 0x01}, code: artifactpolicy.CodeOpaqueDependency},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newNPMFixture(t)
			capture := captureFixture(t, fixture)
			runner := &fakeRunner{graph: fixture.graph, mutateInstall: func(root string) error {
				return os.WriteFile(filepath.Join(root, "node_modules", "a", testCase.path), testCase.payload, 0o600)
			}}
			cache := deriveCacheFixture(t, capture, runner)
			_, err := Materialize(t.Context(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: filepath.Join(t.TempDir(), "work")}, runner.context)
			if got := artifactpolicy.ErrorCode(err); got != testCase.code {
				t.Fatalf("materialized payload error=%v code=%s, want %s", err, got, testCase.code)
			}
		})
	}
}

func TestN08GeneratedWriteSetUsesCommonNodeContract(t *testing.T) {
	fixture := newNPMFixture(t)
	capture := captureFixture(t, fixture)
	if !capture.Evidence.NodeCaptureGraphID.Valid() {
		t.Fatal("npm capture did not reach the common Node graph")
	}
	declaration := digestID([]byte("declared-output"))
	declared := []nodesource.GeneratedOutput{{Path: "dist/cli.js", Class: "source.generated_text", Grammar: "javascript-source-v1", Role: "runtime", DeclarationDigest: declaration}}
	observed := map[string]closuregraph.ID{"dist/cli.js": digestID([]byte("expected")), "dist/extra.js": digestID([]byte("undeclared"))}
	err := nodesource.ValidateObservedOutputs(declared, observed)
	if nodesource.ErrorCode(err) != nodesource.CodeGeneratedOutputDrift {
		t.Fatalf("npm/common generated-output gate = %v (%s)", err, nodesource.ErrorCode(err))
	}
}

func TestExecutionAuditDiagnosticsAndInvocationContainment(t *testing.T) {
	fixture := newNPMFixture(t)
	capture := captureFixture(t, fixture)
	runner := &fakeRunner{graph: fixture.graph}
	cache := deriveCacheFixture(t, capture, runner)
	materialized, err := Materialize(context.Background(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: filepath.Join(t.TempDir(), "work")}, runner.context)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Invoke(context.Background(), materialized, "../escape.js", nil, runner.context)
	assertCode(t, err, CodeLocalPathEscape)
	_, err = Invoke(context.Background(), materialized, "missing.js", nil, runner.context)
	assertCode(t, err, CodeInputUndeclared)
}

func TestPortableSuccessAndVerifiedProviderGates(t *testing.T) {
	fixture := newNPMFixture(t)
	capture := captureFixture(t, fixture)
	runner := &fakeRunner{graph: fixture.graph}
	portable := makeExecutionContext(t, capture, runner)
	runner.context = portable
	if err := validateExecutionContext(portable); err != nil {
		t.Fatalf("portable runner rejected: %v", err)
	}
	_, err := DerivePrivateCache(context.Background(), capture, filepath.Join(t.TempDir(), "cache"), filepath.Join(t.TempDir(), "work"), &ExecutionContext{})
	assertCode(t, err, CodeDerivationUnauthorized)
	if runner.starts != 0 {
		t.Fatalf("verified missing-provider gate started %d processes", runner.starts)
	}
}

func TestVerifiedAuthorityFailuresAreZeroStart(t *testing.T) {
	provider := newVerifiedFixtureProvider()
	config := provider.config()
	if _, err := closureexec.NewAssuredExecutor(config, nil, nil, "test-head"); err == nil || provider.starts != 0 {
		t.Fatalf("missing verified provider did not fail at zero start: %v starts=%d", err, provider.starts)
	}
	badConfig := config
	badConfig.ProviderVersion = "incompatible"
	if _, err := closureexec.NewAssuredExecutor(badConfig, nil, provider, "test-head"); err == nil || provider.starts != 0 {
		t.Fatalf("incompatible verified provider did not fail at zero start: %v starts=%d", err, provider.starts)
	}
	incomplete := config
	incomplete.ProviderBinarySHA256 = ""
	if _, err := closureexec.NewAssuredExecutor(incomplete, nil, provider, "test-head"); err == nil || provider.starts != 0 {
		t.Fatalf("incomplete verified identity did not fail at zero start: %v starts=%d", err, provider.starts)
	}

	fixture := newNPMFixture(t)
	capture := captureFixture(t, fixture)
	for _, testCase := range []struct {
		name   string
		mutate func(*verifiedFixtureProvider, *ExecutionContext)
	}{
		{name: "incomplete capabilities", mutate: func(provider *verifiedFixtureProvider, _ *ExecutionContext) { provider.incompleteCapabilities = true }},
		{name: "cross mode C5", mutate: func(_ *verifiedFixtureProvider, context *ExecutionContext) {
			portableRunner := &fakeRunner{graph: fixture.graph}
			context.BuildPlan = makeExecutionContext(t, capture, portableRunner).BuildPlan
		}},
		{name: "nonce receipt drift", mutate: func(provider *verifiedFixtureProvider, _ *ExecutionContext) { provider.driftCapabilities = true }},
		{name: "provider identity drift", mutate: func(provider *verifiedFixtureProvider, _ *ExecutionContext) { provider.driftIdentity = true }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			caseProvider := newVerifiedFixtureProvider()
			recheck := &fakeRunner{graph: fixture.graph, capture: capture}
			context := makeVerifiedExecutionContext(t, capture, caseProvider, recheck)
			testCase.mutate(caseProvider, context)
			_, err := DerivePrivateCache(t.Context(), capture, filepath.Join(t.TempDir(), "cache"), filepath.Join(t.TempDir(), "work"), context)
			if err == nil {
				t.Fatal("invalid verified authority was accepted")
			}
			if caseProvider.starts != 0 {
				t.Fatalf("invalid verified authority started %d processes", caseProvider.starts)
			}
		})
	}
}

func TestVerifiedProviderExecutesExactCacheCIAndNodePermits(t *testing.T) {
	fixture := newNPMFixture(t)
	capture := captureFixture(t, fixture)
	provider := newVerifiedFixtureProvider()
	recheck := &fakeRunner{graph: fixture.graph, capture: capture}
	authority := makeVerifiedExecutionContext(t, capture, provider, recheck)
	provider.allowExecution = true
	cache, err := DerivePrivateCache(t.Context(), capture, filepath.Join(t.TempDir(), "cache"), filepath.Join(t.TempDir(), "cache-work"), authority)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTreeWritable(cache.root) })
	materialized, err := Materialize(t.Context(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: filepath.Join(t.TempDir(), "materialize-work")}, authority)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Invoke(t.Context(), materialized, "index.js", nil, authority); err != nil {
		t.Fatal(err)
	}
	if provider.starts < 3 {
		t.Fatalf("verified provider did not execute the cache/ci/node operation family: starts=%d", provider.starts)
	}
}

func TestVerifiedProviderObservesRealNodeLaunchedNPMBoundary(t *testing.T) {
	if testing.Short() {
		t.Skip("real verified npm boundary is an integration test")
	}
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm unavailable")
	}
	fixture := newNPMFixture(t)
	capture := captureFixture(t, fixture)
	runner := newConcreteNPMRunner(t)
	provider := newVerifiedFixtureProvider()
	authority := makeVerifiedExecutionContext(t, capture, provider, runner)
	provider.allowExecution = true
	cache, err := DerivePrivateCache(t.Context(), capture, filepath.Join(t.TempDir(), "cache"), filepath.Join(t.TempDir(), "cache-work"), authority)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTreeWritable(cache.root) })
	materialized, err := Materialize(t.Context(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: filepath.Join(t.TempDir(), "materialize-work")}, authority)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Invoke(t.Context(), materialized, "index.js", nil, authority); err != nil {
		t.Fatal(err)
	}
	if len(provider.realLaunches) < 3 {
		t.Fatalf("verified provider observed %d real launches", len(provider.realLaunches))
	}
	managerLaunches := 0
	for _, launch := range provider.realLaunches {
		if launch.Executable != runner.nodePath {
			t.Fatalf("verified launch crossed an unbound interpreter: %+v", launch)
		}
		for _, entry := range launch.Environment {
			if strings.HasPrefix(entry, "PATH=") {
				t.Fatalf("verified launch retained PATH fallback: %+v", launch)
			}
		}
		if len(launch.Argv) > 0 && strings.HasSuffix(launch.Argv[0], "npm-cli.js") {
			managerLaunches++
			if filepath.Clean(filepath.Join(launch.CWD, filepath.FromSlash(launch.Argv[0]))) != runner.npmPath {
				t.Fatalf("verified npm entry point differs from C0 binding: %+v", launch)
			}
		}
	}
	if managerLaunches < 2 {
		t.Fatalf("verified provider observed %d Node-launched npm operations", managerLaunches)
	}
}

func TestVerifiedProviderRejectsUndeclaredProcessObservation(t *testing.T) {
	fixture := newNPMFixture(t)
	capture := captureFixture(t, fixture)
	provider := newVerifiedFixtureProvider()
	recheck := &fakeRunner{graph: fixture.graph, capture: capture}
	authority := makeVerifiedExecutionContext(t, capture, provider, recheck)
	provider.allowExecution = true
	provider.extraProcess = "ambient/downloader"
	_, err := DerivePrivateCache(t.Context(), capture, filepath.Join(t.TempDir(), "cache"), filepath.Join(t.TempDir(), "cache-work"), authority)
	if err == nil || !strings.Contains(err.Error(), "process set") {
		t.Fatalf("undeclared verified process observation was accepted: %v", err)
	}
}

func TestC5AndExactExecutableGatesAreZeroStart(t *testing.T) {
	fixture := newNPMFixture(t)
	capture := captureFixture(t, fixture)
	t.Run("action absent from C5", func(t *testing.T) {
		runner := &fakeRunner{graph: fixture.graph, capture: capture}
		context := makeExecutionContext(t, capture, runner)
		context.BuildPlan.ActionNodeIDs = context.BuildPlan.ActionNodeIDs[1:]
		context.BuildPlan.Waves = [][]closuregraph.ID{append([]closuregraph.ID(nil), context.BuildPlan.ActionNodeIDs...)}
		_, err := DerivePrivateCache(t.Context(), capture, filepath.Join(t.TempDir(), "cache"), filepath.Join(t.TempDir(), "work"), context)
		assertCode(t, err, CodeDerivationUnauthorized)
		if runner.starts != 0 {
			t.Fatalf("absent C5 action started %d processes", runner.starts)
		}
	})
	t.Run("unbound shebang interpreter", func(t *testing.T) {
		runner := &fakeRunner{graph: fixture.graph, capture: capture}
		context := makeExecutionContext(t, capture, runner)
		context.Runtime.Manager.ExecutableRelativePath = "bin/npm"
		_, err := DerivePrivateCache(t.Context(), capture, filepath.Join(t.TempDir(), "cache"), filepath.Join(t.TempDir(), "work"), context)
		assertCode(t, err, CodeDerivationUnauthorized)
		if runner.starts != 0 {
			t.Fatalf("unbound interpreter started %d processes", runner.starts)
		}
	})
	for _, testCase := range []struct {
		name   string
		mutate func(*invocation)
	}{
		{name: "remove offline", mutate: func(call *invocation) { call.Args = removeString(call.Args, "--offline") }},
		{name: "remove ignore scripts", mutate: func(call *invocation) { call.Args = removeString(call.Args, "--ignore-scripts") }},
		{name: "cwd substitution", mutate: func(call *invocation) { call.CWD = "work/substituted" }},
		{name: "PATH substitution", mutate: func(call *invocation) { call.Environment["PATH"] = "/ambient/bin" }},
		{name: "input substitution", mutate: func(call *invocation) {
			for id := range call.InputMounts {
				call.InputMounts[id] = "capture/substituted"
				break
			}
		}},
		{name: "cache output substitution", mutate: func(call *invocation) { call.Template["cache"] = []string{"../ambient-cache"} }},
		{name: "write root substitution", mutate: func(call *invocation) { call.WriteRoots = []string{"work/ambient-output"} }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			runner := &fakeRunner{graph: fixture.graph, capture: capture}
			authority := makeExecutionContext(t, capture, runner)
			runner.context = authority
			session, err := newRunnerSession(t.Context(), capture, authority)
			if err != nil {
				t.Fatal(err)
			}
			call := cacheInvocationForTest(capture, authority)
			testCase.mutate(&call)
			if _, err = session.run(t.Context(), call, "npm-cache"); ErrorCode(err) != CodeDerivationUnauthorized {
				t.Fatalf("mutated C5 action admitted: %v", err)
			}
			if runner.starts != 0 {
				t.Fatalf("mutated C5 action crossed process-start seam %d times", runner.starts)
			}
		})
	}
}

func TestPortableAuditDoesNotEncodeUnobservedLosslessZeros(t *testing.T) {
	fixture := newNPMFixture(t)
	capture := captureFixture(t, fixture)
	runner := &fakeRunner{graph: fixture.graph}
	cache := deriveCacheFixture(t, capture, runner)
	materialized, err := Materialize(t.Context(), cache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: filepath.Join(t.TempDir(), "work")}, runner.context)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(materialized.Receipt)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"resolver_requests", "ambient_cache_reads", "lifecycle_scripts", "undeclared_processes", "undeclared_reads", "undeclared_writes"} {
		if bytes.Contains(payload, []byte(forbidden)) {
			t.Fatalf("portable audit synthesized lossless field %q: %s", forbidden, payload)
		}
	}
}

func TestClosedParserAndMetadataVariants(t *testing.T) {
	fixture := newNPMFixture(t)
	for _, testCase := range []struct {
		name   string
		mutate func(*ParseRequest)
		code   string
	}{
		{"missing lock", func(r *ParseRequest) { r.LockName = ""; r.LockBytes = nil }, CodeLockMissing},
		{"unknown lock", func(r *ParseRequest) { r.LockName = "yarn.lock" }, CodeLockFormatUnsupported},
		{"missing target", func(r *ParseRequest) { r.Target.OS = "" }, CodeGraphIncomplete},
		{"manifest escape", func(r *ParseRequest) { r.Manifests["/package.json"] = r.Manifests["package.json"] }, CodeLocalPathEscape},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := fixture.request
			request.Manifests = map[string][]byte{}
			for k, v := range fixture.request.Manifests {
				request.Manifests[k] = v
			}
			testCase.mutate(&request)
			_, err := Parse(request)
			assertCode(t, err, testCase.code)
		})
	}
	var lock map[string]any
	_ = json.Unmarshal(fixture.request.LockBytes, &lock)
	entry := lock["packages"].(map[string]any)["node_modules/a"].(map[string]any)
	delete(entry, "integrity")
	request := fixture.request
	request.LockBytes = mustJSON(t, lock)
	_, err := Parse(request)
	assertCode(t, err, CodeIntegrityMissing)

	// Explicit gypfile=false disables npm's implicit node-gyp synthesis.
	pkg := fixturePackage{name: "a", version: "1.0.0", metadata: map[string]any{"dependencies": map[string]string{"b": "1.0.0"}, "gypfile": false}, files: map[string][]byte{"index.js": []byte("module.exports=1\n"), "binding.gyp": []byte("{'targets': []}\n")}}
	fixture = newNPMFixture(t)
	replaceTarballAndLock(t, &fixture, "node_modules/a", pkg)
	graph, err := Parse(fixture.request)
	if err != nil {
		t.Fatal(err)
	}
	fixture.graph = graph
	if _, err = captureFixtureError(t, fixture); err != nil {
		t.Fatalf("gypfile=false package rejected: %v", err)
	}
}

func TestMetadataReconciliationBranchesAndClosedHelpers(t *testing.T) {
	base := Package{InstallPath: "node_modules/a", Name: "a", Version: "1.0.0", Dependencies: map[string]string{"b": "1.0.0"}, OptionalDependencies: map[string]string{}, PeerDependencies: map[string]string{"p": "1.0.0"}, PeerOptional: map[string]bool{"p": true}, OS: []string{"darwin"}, CPU: []string{"arm64"}, Libc: []string{"none"}}
	manifest := packageManifest{Name: "a", Version: "1.0.0", Dependencies: map[string]string{"b": "1.0.0"}, OptionalDependencies: map[string]string{}, PeerDependencies: map[string]string{"p": "1.0.0"}, PeerDependenciesMeta: map[string]struct {
		Optional bool `json:"optional"`
	}{"p": {Optional: true}}, OS: []string{"darwin"}, CPU: []string{"arm64"}, Libc: []string{"none"}, Scripts: map[string]string{}}
	for _, testCase := range []struct {
		name   string
		mutate func(*Package, *packageManifest, *tarInspection)
		code   string
	}{
		{"identity", func(_ *Package, m *packageManifest, _ *tarInspection) { m.Version = "2.0.0" }, CodeMetadataMismatch},
		{"dependencies", func(_ *Package, m *packageManifest, _ *tarInspection) { m.Dependencies["b"] = "2.0.0" }, CodeMetadataMismatch},
		{"peer meta", func(_ *Package, m *packageManifest, _ *tarInspection) {
			m.PeerDependenciesMeta["p"] = struct {
				Optional bool `json:"optional"`
			}{Optional: false}
		}, CodeMetadataMismatch},
		{"platform", func(_ *Package, m *packageManifest, _ *tarInspection) { m.CPU = []string{"x64"} }, CodeMetadataMismatch},
		{"install marker", func(p *Package, m *packageManifest, _ *tarInspection) {
			p.HasInstallScript = true
			m.Scripts = map[string]string{}
		}, CodeMetadataMismatch},
		{"bundled bytes", func(_ *Package, _ *packageManifest, i *tarInspection) { i.bundled = true }, CodeBundledDependencyUnsupported},
		{"implicit gyp", func(_ *Package, _ *packageManifest, i *tarInspection) { i.bindingGYP = true }, CodeNativeBuildUnsupported},
		{"lifecycle", func(p *Package, m *packageManifest, _ *tarInspection) {
			p.HasInstallScript = true
			m.Scripts = map[string]string{"postinstall": "node x.js"}
		}, CodeHookUndeclared},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			pkg := base
			m := manifest
			m.Dependencies = cloneMap(manifest.Dependencies)
			m.PeerDependenciesMeta = map[string]struct {
				Optional bool `json:"optional"`
			}{"p": {Optional: true}}
			inspection := tarInspection{}
			testCase.mutate(&pkg, &m, &inspection)
			assertCode(t, reconcileEmbeddedMetadata(pkg, m, inspection), testCase.code)
		})
	}
	patterns, err := workspacePatterns([]byte(`{"packages":["packages/*"]}`))
	if err != nil || !reflect.DeepEqual(patterns, []string{"packages/*"}) {
		t.Fatalf("workspace object grammar failed: %v %v", patterns, err)
	}
	if _, err = workspacePatterns([]byte(`42`)); ErrorCode(err) != CodeLockFormatUnsupported {
		t.Fatalf("numeric workspace grammar admitted: %v", err)
	}
	if selectorMatches([]string{"!darwin"}, "darwin") || !selectorMatches([]string{"!linux"}, "darwin") {
		t.Fatal("negative target selector semantics drifted")
	}
	if (&Error{Code: CodeLockMissing}).Error() != CodeLockMissing || ErrorCode(fmt.Errorf("plain")) != "" {
		t.Fatal("diagnostic fallback semantics drifted")
	}
	if _, err = DerivePrivateCache(context.Background(), nil, "", "", nil); ErrorCode(err) != CodeInputUndeclared {
		t.Fatalf("nil cache derivation admitted: %v", err)
	}
	if _, err = Materialize(context.Background(), nil, MaterializeRequest{}, nil); ErrorCode(err) != CodeInputUndeclared {
		t.Fatalf("nil materialization admitted: %v", err)
	}
}

func TestN01RealNPMCIUsesOnlyDerivedPrivateCache(t *testing.T) {
	if testing.Short() {
		t.Skip("real npm offline materialization is an integration test")
	}
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm unavailable")
	}
	fixture := newNPMFixture(t)
	capture := captureFixture(t, fixture)
	runner := newConcreteNPMRunner(t)
	cache := deriveCacheFixture(t, capture, runner)
	secondCache := deriveCacheFixture(t, capture, runner)
	if cache.Receipt.ID != secondCache.Receipt.ID {
		t.Fatalf("private npm cache derivation is not deterministic: %s != %s\nfirst=%+v\nsecond=%+v", cache.Receipt.ID, secondCache.Receipt.ID, cache.Receipt.Files, secondCache.Receipt.Files)
	}
	materialized, err := Materialize(context.Background(), cache, MaterializeRequest{
		Destination: filepath.Join(t.TempDir(), "materialized"),
		WorkRoot:    filepath.Join(t.TempDir(), "materialize-work"),
	}, runner.context)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(materialized.MaterializedPackages, []string{"node_modules/a", "node_modules/b", "node_modules/workspace"}) {
		t.Fatalf("real npm ci materialized unexpected graph: %v", materialized.MaterializedPackages)
	}
	if _, err = Invoke(t.Context(), materialized, "index.js", nil, runner.context); err != nil {
		t.Fatalf("real C0-bound Node invocation failed: %v", err)
	}
	firstFiles, err := inventoryFiles(materialized.Root)
	if err != nil {
		t.Fatal(err)
	}
	ambient := filepath.Join(runner.executionRoot, "ambient-home", ".npm", "_cacache", "content-v2", "sha512", "poison")
	if err = os.MkdirAll(filepath.Dir(ambient), 0o700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(ambient, []byte("same package identity, different ambient bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	second, err := Materialize(context.Background(), secondCache, MaterializeRequest{Destination: filepath.Join(t.TempDir(), "materialized"), WorkRoot: filepath.Join(t.TempDir(), "materialize-work")}, runner.context)
	if err != nil {
		t.Fatal(err)
	}
	secondFiles, err := inventoryFiles(second.Root)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(secondFiles, firstFiles) || !reflect.DeepEqual(second.MaterializedPackages, materialized.MaterializedPackages) {
		t.Fatalf("S03/S08 poisoned ambient replay drifted: first=%+v second=%+v", firstFiles, secondFiles)
	}
	if second.Receipt.Audit.Network != "not-observed" {
		t.Fatalf("portable replay inflated network evidence: %+v", second.Receipt.Audit)
	}
	if len(runner.launches) == 0 {
		t.Fatal("real npm vector observed no process launches")
	}
	managerLaunches := 0
	for _, launch := range runner.launches {
		if launch.Executable != runner.nodePath || len(launch.Argv) == 0 {
			t.Fatalf("real launch bypassed exact Node boundary: %+v", launch)
		}
		for _, entry := range launch.Environment {
			if strings.HasPrefix(entry, "PATH=") {
				t.Fatalf("real launch retained ambient PATH fallback: %+v", launch)
			}
		}
		if strings.HasSuffix(launch.Argv[0], "npm-cli.js") {
			managerLaunches++
			if filepath.Clean(filepath.Join(launch.CWD, filepath.FromSlash(launch.Argv[0]))) != runner.npmPath {
				t.Fatalf("portable npm entry point differs from C0 binding: %+v", launch)
			}
		}
	}
	if managerLaunches < 2 {
		t.Fatalf("portable path observed %d Node-launched npm operations", managerLaunches)
	}
}

type concreteNPMRunner struct {
	*closureexec.ManagerProcessRunner
	context                   *ExecutionContext
	executionRoot             string
	npmPath, nodePath         string
	npmRelative, nodeRelative string
	npmDigest, nodeDigest     closuregraph.ID
	managerFingerprint        closuregraph.ID
	launches                  []closureexec.ProcessLaunch
}

func newConcreteNPMRunner(t *testing.T) *concreteNPMRunner {
	t.Helper()
	executionRoot := filepath.Join(t.TempDir(), "execution")
	if err := os.MkdirAll(filepath.Join(executionRoot, "bin"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(executionRoot, "work"), 0o700); err != nil {
		t.Fatal(err)
	}
	npmCommand, err := exec.LookPath("npm")
	if err != nil {
		t.Fatal(err)
	}
	npmPath, err := filepath.EvalSymlinks(npmCommand)
	if err != nil {
		t.Fatal(err)
	}
	npmRoot := filepath.Dir(filepath.Dir(npmPath))
	stagedNPMRoot := filepath.Join(executionRoot, "toolchain", "npm")
	if err = copyContainedTreeDereferencingLinks(npmRoot, npmRoot, stagedNPMRoot, map[string]bool{}); err != nil {
		t.Fatal(err)
	}
	npmLeaf, err := filepath.Rel(npmRoot, npmPath)
	if err != nil {
		t.Fatal(err)
	}
	stagedNPM := filepath.Join(stagedNPMRoot, npmLeaf)
	if err = os.Chmod(stagedNPM, 0o500); err != nil { // #nosec G302 -- exact selected CLI entry point must be executable.
		t.Fatal(err)
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
	stagedNodeRoot := filepath.Join(executionRoot, "toolchain", "node")
	if err = os.MkdirAll(filepath.Join(stagedNodeRoot, "bin"), 0o700); err != nil {
		t.Fatal(err)
	}
	nodeLeaf, err := filepath.Rel(nodeRoot, nodePath)
	if err != nil {
		t.Fatal(err)
	}
	nodePayload, err := os.ReadFile(nodePath) // #nosec G304 -- exact integration tool selected at C0.
	if err != nil {
		t.Fatal(err)
	}
	stagedNode := filepath.Join(stagedNodeRoot, nodeLeaf)
	if err = os.WriteFile(stagedNode, nodePayload, 0o500); err != nil {
		t.Fatal(err)
	}
	linkedNodeLibraries, err := filepath.Glob(filepath.Join(nodeRoot, "lib", "libnode*.dylib"))
	if err != nil {
		t.Fatal(err)
	}
	for _, library := range linkedNodeLibraries {
		payload, readErr := os.ReadFile(library) // #nosec G304 -- selected runtime's adjacent rpath dependency.
		if readErr != nil {
			t.Fatal(readErr)
		}
		target := filepath.Join(stagedNodeRoot, "lib", filepath.Base(library))
		if err = os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(target, payload, 0o400); err != nil {
			t.Fatal(err)
		}
	}
	npmPayload, err := os.ReadFile(stagedNPM) // #nosec G304 -- exact staged package-manager executable.
	if err != nil {
		t.Fatal(err)
	}
	manager, err := closureexec.NewManagerProcessRunner(executionRoot, filepath.Join(executionRoot, "output"))
	if err != nil {
		t.Fatal(err)
	}
	npmDigest := digestID(npmPayload)
	nodeDigest := digestID(nodePayload)
	managerFingerprint, err := closuregraph.DomainID("npm-interpreted-toolchain-v1", map[string]any{"entrypoint_sha256": string(npmDigest), "node_sha256": string(nodeDigest)})
	if err != nil {
		t.Fatal(err)
	}
	runner := &concreteNPMRunner{ManagerProcessRunner: manager, executionRoot: executionRoot, npmPath: stagedNPM, nodePath: stagedNode, npmRelative: filepath.ToSlash(filepath.Join("toolchain", "npm", npmLeaf)), nodeRelative: filepath.ToSlash(filepath.Join("toolchain", "node", nodeLeaf)), npmDigest: npmDigest, nodeDigest: nodeDigest, managerFingerprint: managerFingerprint}
	manager.ProcessLaunchObserver = func(launch closureexec.ProcessLaunch) { runner.launches = append(runner.launches, launch) }
	return runner
}

func (runner *concreteNPMRunner) RecheckTool(_ context.Context, tool nodesource.ToolIdentity) (closureexec.ToolchainIdentity, error) {
	nodePayload, err := os.ReadFile(runner.nodePath) // #nosec G304 -- exact selected runtime below the staged toolchain.
	if err != nil || digestID(nodePayload) != runner.nodeDigest {
		return closureexec.ToolchainIdentity{}, fmt.Errorf("selected tool changed: %w", err)
	}
	fingerprint := runner.nodeDigest
	if tool.Role == "package-manager" {
		npmPayload, readErr := os.ReadFile(runner.npmPath) // #nosec G304 -- exact selected npm CLI below the staged toolchain.
		if readErr != nil || digestID(npmPayload) != runner.npmDigest {
			return closureexec.ToolchainIdentity{}, fmt.Errorf("selected npm entry point changed: %w", readErr)
		}
		fingerprint = runner.managerFingerprint
	}
	return closureexec.ToolchainIdentity{Fingerprint: fingerprint, ExecutableSHA256: runner.nodeDigest}, nil
}

type fakeRunner struct {
	graph            Graph
	capture          *Capture
	context          *ExecutionContext
	starts           int
	extraPackage     string
	mutateInstall    func(string) error
	recheckErr       error
	lastCI, lastNode closureexec.DerivationPermit
}

func cacheInvocationForTest(capture *Capture, authority *ExecutionContext) invocation {
	keys := make([]string, 0, len(capture.tarballs))
	for key := range capture.tarballs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	item := capture.tarballs[keys[0]]
	workPath := "work/npm-cache-0000"
	tarballPath := "capture/npm-cache-tarball.tgz"
	relativeTarball := relativeLogical(workPath, tarballPath)
	return invocation{
		Tool: authority.Runtime.Manager, CWD: workPath,
		Args:        []string{relativeLogical(workPath, authority.Runtime.Manager.EntrypointRelativePath), "cache", "add", relativeTarball, "--offline", "--ignore-scripts", "--cache", ".cache", "--userconfig", ".home/npmrc", "--logs-dir", ".home/logs", "--no-audit", "--no-fund"},
		Environment: npmEnvironment(".home", ".cache", ".home/npmrc", ".home/logs"),
		Inputs:      map[closuregraph.ID]closureexec.AdmittedInput{capture.project.receiptID: capture.project.input, item.receiptID: item.input},
		InputMounts: map[closuregraph.ID]string{capture.project.receiptID: "capture/npm-cache-project", item.receiptID: tarballPath},
		ReadRoots:   append([]string{"capture/npm-cache-project", tarballPath}, authority.Runtime.Manager.ReadRoots...), WriteRoots: []string{workPath},
		WorkCopies: []closureexec.WorkCopy{{ReceiptID: capture.project.receiptID, Path: workPath, Retain: true}},
		Template:   map[string][]string{"manager_entrypoint": {relativeLogical(workPath, authority.Runtime.Manager.EntrypointRelativePath)}, "tarball": {relativeTarball}, "cache": {".cache"}, "userconfig": {".home/npmrc"}, "logs": {".home/logs"}, "work": {workPath}},
	}
}

func removeString(values []string, target string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != target {
			result = append(result, value)
		}
	}
	return result
}

func (runner *fakeRunner) RecheckTool(_ context.Context, tool nodesource.ToolIdentity) (closureexec.ToolchainIdentity, error) {
	if runner.recheckErr != nil {
		return closureexec.ToolchainIdentity{}, runner.recheckErr
	}
	return closureexec.ToolchainIdentity{Fingerprint: tool.Fingerprint, ExecutableSHA256: tool.ExecutableSHA256}, nil
}

func (runner *fakeRunner) Run(_ context.Context, request closureexec.ExecutionRequest) (closureexec.PortableRunResult, error) {
	runner.starts++
	permit := request.Permit
	executionRoot := runner.context.ExecutionRoot
	if err := prepareFakeExecution(request, executionRoot); err != nil {
		return closureexec.PortableRunResult{}, err
	}
	cwd := filepath.Join(executionRoot, filepath.FromSlash(permit.CWD))
	if len(permit.Argv) > 2 && permit.Argv[1] == "cache" {
		cache := filepath.Join(cwd, filepath.FromSlash(argAfter(permit.Argv, "--cache")))
		if err := os.MkdirAll(cache, 0o700); err != nil {
			return closureexec.PortableRunResult{}, err
		}
		payload := []byte(permit.Argv[3])
		name := fmt.Sprintf("content-%x", sha512.Sum512(payload))
		if err := os.WriteFile(filepath.Join(cache, name), payload, 0o600); err != nil {
			return closureexec.PortableRunResult{}, err
		}
	}
	if len(permit.Argv) > 1 && permit.Argv[1] == "ci" {
		runner.lastCI = permit
		for _, pkg := range runner.graph.Packages {
			if !pkg.Selected || pkg.InstallPath == "" {
				continue
			}
			target := filepath.Join(cwd, filepath.FromSlash(pkg.InstallPath))
			if pkg.Link {
				if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
					return closureexec.PortableRunResult{}, err
				}
				if err := os.Symlink(filepath.Join("..", filepath.FromSlash(runner.graph.linkTargets[pkg.InstallPath])), target); err != nil {
					return closureexec.PortableRunResult{}, err
				}
				continue
			}
			if !isExternalInstallPath(pkg.InstallPath) {
				continue
			}
			if err := os.MkdirAll(target, 0o700); err != nil {
				return closureexec.PortableRunResult{}, err
			}
			if runner.capture == nil {
				return closureexec.PortableRunResult{}, fmt.Errorf("fake npm runner lacks admitted tarball authority")
			}
			if err := extractFixtureTarball(runner.capture.tarballs[pkg.InstallPath], target); err != nil {
				return closureexec.PortableRunResult{}, err
			}
		}
		if runner.extraPackage != "" {
			target := filepath.Join(cwd, filepath.FromSlash(runner.extraPackage))
			if err := os.MkdirAll(target, 0o700); err != nil {
				return closureexec.PortableRunResult{}, err
			}
			if err := os.WriteFile(filepath.Join(target, "package.json"), []byte(`{"name":"extra"}`), 0o600); err != nil {
				return closureexec.PortableRunResult{}, err
			}
		}
		if runner.mutateInstall != nil {
			if err := runner.mutateInstall(cwd); err != nil {
				return closureexec.PortableRunResult{}, err
			}
		}
	}
	if len(permit.Argv) == 0 || !strings.HasSuffix(permit.Argv[0], "npm-cli.js") {
		runner.lastNode = permit
	}
	return portableResult(executionRoot, permit)
}

func prepareFakeExecution(request closureexec.ExecutionRequest, root string) error {
	inputs := map[closuregraph.ID]closureexec.ReplayInput{}
	for _, input := range request.Inputs {
		inputs[input.ReceiptID] = input
		path, err := input.ProtectedPath()
		if err != nil {
			return err
		}
		target := filepath.Join(root, filepath.FromSlash(input.MountPath))
		_ = os.RemoveAll(target)
		if input.IsTree() {
			if err = os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				return err
			}
			if err = copyWritableTree(path, target); err != nil {
				return err
			}
		} else {
			payload, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
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
		input := inputs[work.ReceiptID]
		source, err := input.ProtectedPath()
		if err != nil {
			return err
		}
		target := filepath.Join(root, filepath.FromSlash(work.Path))
		_ = os.RemoveAll(target)
		if err = os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		if err = copyWritableTree(source, target); err != nil {
			return err
		}
	}
	output := filepath.Join(root, "output")
	_ = os.RemoveAll(output)
	return os.MkdirAll(output, 0o700)
}

func makeExecutionContext(t *testing.T, capture *Capture, runner interface {
	closureexec.PortableProcessRunner
	RecheckTool(context.Context, nodesource.ToolIdentity) (closureexec.ToolchainIdentity, error)
}) *ExecutionContext {
	t.Helper()
	executionRoot := ""
	if concrete, ok := runner.(*concreteNPMRunner); ok {
		executionRoot = concrete.executionRoot
	} else {
		executionRoot = filepath.Join(t.TempDir(), "execution")
		if err := os.MkdirAll(filepath.Join(executionRoot, "bin"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(executionRoot, "work"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(executionRoot, "output"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	target := capture.Graph.Target
	platform := closuregraph.TargetPlatformPayload{OS: target.OS, Architecture: target.Architecture, ABI: "node", Libc: target.Libc, MinimumRuntime: "bound-by-node-runtime", SDKID: "none", TargetTriple: target.Architecture + "-" + target.OS, Runtime: "node", LanguageModes: map[string]string{"package_manager": "npm"}, Tuning: map[string]string{}}
	platformNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: platform}
	platformID, err := platformNode.ID()
	if err != nil {
		t.Fatal(err)
	}
	selection, err := closuregraph.NewSelectionContext(capture.NodeCapture.Graph.RootNodeIDs, map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID}, []string{}, target.IncludeDev, map[string]string{"os": target.OS, "cpu": target.Architecture, "libc": target.Libc}, map[string]string{}, []string{"npm-platform-v1"})
	if err != nil {
		t.Fatal(err)
	}
	tool := func(role, path, seed string) nodesource.ToolIdentity {
		return nodesource.ToolIdentity{Role: role, PolicySelector: role + "-v1", ExecutableRelativePath: path, VersionOutput: role + " fixture", PlatformABI: target.OS + "-" + target.Architecture, Fingerprint: digestID([]byte(seed + "-tree")), ExecutableSHA256: digestID([]byte(seed + "-executable")), ExecutionDomain: closuregraph.ExecutionTarget}
	}
	runtimeBinding := nodesource.RuntimeBinding{Platform: platform, Node: tool("node-runtime", "bin/node", "node"), Manager: tool("package-manager", "bin/node", "npm"), TargetNodeIDs: append([]closuregraph.ID(nil), capture.NodeCapture.Graph.RootNodeIDs...)}
	runtimeBinding.Node.ReadRoots = []string{"toolchain/node"}
	runtimeBinding.Manager.EntrypointRelativePath = "toolchain/npm/bin/npm-cli.js"
	runtimeBinding.Manager.ReadRoots = []string{"toolchain/node", "toolchain/npm"}
	runtimeBinding.Manager.ExecutableSHA256 = runtimeBinding.Node.ExecutableSHA256
	if concrete, ok := runner.(*concreteNPMRunner); ok {
		runtimeBinding.Manager.ExecutableRelativePath = concrete.nodeRelative
		runtimeBinding.Manager.EntrypointRelativePath = concrete.npmRelative
		runtimeBinding.Manager.ExecutableSHA256 = concrete.nodeDigest
		runtimeBinding.Manager.Fingerprint = concrete.managerFingerprint
		runtimeBinding.Node.ExecutableRelativePath = concrete.nodeRelative
		runtimeBinding.Node.ExecutableSHA256 = concrete.nodeDigest
		runtimeBinding.Node.Fingerprint = concrete.nodeDigest
	}
	c0, err := nodesource.NewC0Checkpoint(capture.NodeCapture, selection, runtimeBinding)
	if err != nil {
		t.Fatal(err)
	}
	runtimeBinding.C0Checkpoint = &c0
	evaluator := closuregraph.ConditionEvaluatorFunc{EvaluatorID: "npm-platform-v1", EvaluateFunc: evaluateNPMPlatformCondition}
	_, plan, err := nodesource.Close(capture.NodeCapture, selection, runtimeBinding, []closuregraph.ConditionEvaluator{evaluator}, closureexec.PortableExecutionPolicyID)
	if err != nil {
		t.Fatal(err)
	}
	executor, err := closureexec.NewAssuredExecutor(closureexec.DefaultAssuranceConfig(), runner, nil, "test-head")
	if err != nil {
		t.Fatal(err)
	}
	return &ExecutionContext{Executor: executor, Selection: selection, Runtime: runtimeBinding, BuildPlan: plan, Recheck: runner.RecheckTool, ExecutionRoot: executionRoot}
}

func makeVerifiedExecutionContext(t *testing.T, capture *Capture, provider *verifiedFixtureProvider, recheck interface {
	closureexec.PortableProcessRunner
	RecheckTool(context.Context, nodesource.ToolIdentity) (closureexec.ToolchainIdentity, error)
}) *ExecutionContext {
	t.Helper()
	base := makeExecutionContext(t, capture, recheck)
	if fake, ok := recheck.(*fakeRunner); ok {
		fake.context = base
	}
	provider.execution = recheck
	evaluator := closuregraph.ConditionEvaluatorFunc{EvaluatorID: "npm-platform-v1", EvaluateFunc: evaluateNPMPlatformCondition}
	_, plan, err := nodesource.Close(capture.NodeCapture, base.Selection, base.Runtime, []closuregraph.ConditionEvaluator{evaluator}, closureexec.VerifiedExecutionPolicyID)
	if err != nil {
		t.Fatal(err)
	}
	executor, err := closureexec.NewAssuredExecutor(provider.config(), nil, provider, "test-head")
	if err != nil {
		t.Fatal(err)
	}
	return &ExecutionContext{Executor: executor, Selection: base.Selection, Runtime: base.Runtime, BuildPlan: plan, Recheck: recheck.RecheckTool, ExecutionRoot: base.ExecutionRoot}
}
func captureFixture(t *testing.T, fixture npmFixture) *Capture {
	t.Helper()
	capture, err := captureFixtureError(t, fixture)
	if err != nil {
		t.Fatal(err)
	}
	return capture
}
func captureFixtureError(t *testing.T, fixture npmFixture) (*Capture, error) {
	t.Helper()
	storeRoot := filepath.Join(t.TempDir(), "capture-store")
	store, err := closureexec.NewCaptureStore(storeRoot)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTreeWritable(storeRoot) })
	return CaptureAndAdmit(context.Background(), CaptureRequest{Graph: fixture.graph, ProjectRoot: fixture.root, Tarballs: fixture.tarballs, WorkRoot: fixture.work, Store: store, Policy: artifactpolicy.NewService(), PreviousCausalHead: "test-head"})
}
func deriveCacheFixture(t *testing.T, capture *Capture, runner interface {
	closureexec.PortableProcessRunner
	RecheckTool(context.Context, nodesource.ToolIdentity) (closureexec.ToolchainIdentity, error)
}) *PrivateCache {
	t.Helper()
	if fake, ok := runner.(*fakeRunner); ok {
		fake.capture = capture
		if fake.context == nil {
			fake.context = makeExecutionContext(t, capture, fake)
		}
	}
	if concrete, ok := runner.(*concreteNPMRunner); ok && concrete.context == nil {
		concrete.context = makeExecutionContext(t, capture, concrete)
	}
	cacheRoot := filepath.Join(t.TempDir(), "cache")
	cache, err := DerivePrivateCache(context.Background(), capture, cacheRoot, filepath.Join(t.TempDir(), "cache-work"), runnerContext(runner))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeTreeWritable(cacheRoot) })
	return cache
}

func runnerContext(runner any) *ExecutionContext {
	switch typed := runner.(type) {
	case *fakeRunner:
		return typed.context
	case *concreteNPMRunner:
		return typed.context
	default:
		return nil
	}
}

func portableResult(executionRoot string, permit closureexec.DerivationPermit) (closureexec.PortableRunResult, error) {
	root := filepath.Join(executionRoot, filepath.FromSlash(permit.Environment["CURATOR_OUTPUT_ROOT"]))
	for _, expected := range permit.ExpectedEvidence {
		path := filepath.Join(root, filepath.FromSlash(expected.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return closureexec.PortableRunResult{}, err
		}
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			return closureexec.PortableRunResult{}, err
		}
	}
	return closureexec.PortableRunResult{ExitCode: 0, OutputRoot: root}, nil
}

type verifiedFixtureProvider struct {
	identity               closureexec.ProviderIdentity
	starts                 int
	receipts               map[string]closureexec.ProviderCapabilityReceipt
	incompleteCapabilities bool
	driftCapabilities      bool
	driftIdentity          bool
	negotiations           int
	allowExecution         bool
	execution              closureexec.PortableProcessRunner
	extraProcess           string
	networkAttempt         bool
	realLaunches           []closureexec.ProcessLaunch
}

func newVerifiedFixtureProvider() *verifiedFixtureProvider {
	return &verifiedFixtureProvider{identity: closureexec.ProviderIdentity{
		Contract: closureexec.VerifiedProviderContractID, ProviderID: "fixture.provider", Version: "1.0.0",
		BinarySHA256: digestID([]byte("fixture-provider-binary")), TrustEvidence: "fixture-trust-root",
	}, receipts: map[string]closureexec.ProviderCapabilityReceipt{}}
}

func (provider *verifiedFixtureProvider) config() closureexec.AssuranceConfig {
	return closureexec.AssuranceConfig{Mode: closureexec.AssuranceVerified, ProviderID: provider.identity.ProviderID, ProviderVersion: provider.identity.Version, ProviderBinarySHA256: provider.identity.BinarySHA256, ProviderTrustEvidence: provider.identity.TrustEvidence}
}
func (provider *verifiedFixtureProvider) Identity() closureexec.ProviderIdentity {
	if provider.driftIdentity && provider.negotiations > 0 {
		identity := provider.identity
		identity.ProviderID = "fixture.provider.drifted"
		return identity
	}
	return provider.identity
}
func (*verifiedFixtureProvider) LosslessObservation() bool { return true }
func (provider *verifiedFixtureProvider) Negotiate(_ context.Context, nonce string) (closureexec.ProviderCapabilityReceipt, error) {
	provider.negotiations++
	if receipt, ok := provider.receipts[nonce]; ok {
		if provider.driftCapabilities {
			receipt.Capabilities = append([]closureexec.CapabilityEvidence(nil), receipt.Capabilities...)
			receipt.Capabilities[0].Status = "drifted"
		}
		return receipt, nil
	}
	now := time.Now().UTC()
	receipt := closureexec.ProviderCapabilityReceipt{Provider: provider.identity, Health: "healthy", Nonce: nonce, ObservedAt: now.Add(-time.Second), ExpiresAt: now.Add(time.Minute), Capabilities: []closureexec.CapabilityEvidence{
		{CapabilityID: "total-network-denial-v1", Status: "established"},
		{CapabilityID: "read-only-source-and-toolchain-v1", Status: "established"},
		{CapabilityID: "exact-executable-allowlisting-v1", Status: "established"},
		{CapabilityID: "private-build-root-only-writes-v1", Status: "established"},
		{CapabilityID: "hard-aggregate-descendant-resource-bounds-v1", Status: "established"},
		{CapabilityID: "fail-closed-capability-preflight-v1", Status: "established"},
	}}
	if provider.incompleteCapabilities {
		receipt.Capabilities = receipt.Capabilities[:len(receipt.Capabilities)-1]
	}
	provider.receipts[nonce] = receipt
	return receipt, nil
}
func (provider *verifiedFixtureProvider) EnforceAndObserve(ctx context.Context, request closureexec.ExecutionRequest) (closureexec.Audit, error) {
	provider.starts++
	if !provider.allowExecution || provider.execution == nil {
		return closureexec.Audit{}, fmt.Errorf("unexpected verified process start")
	}
	result, err := provider.execution.Run(ctx, request)
	if err != nil {
		return closureexec.Audit{}, err
	}
	defer result.Release()
	var observed *closureexec.ProcessLaunch
	if concrete, ok := provider.execution.(*concreteNPMRunner); ok {
		if len(concrete.launches) == 0 {
			return closureexec.Audit{}, fmt.Errorf("real provider observed no launch")
		}
		launch := concrete.launches[len(concrete.launches)-1]
		observed = &launch
		provider.realLaunches = append(provider.realLaunches, launch)
	}
	outputs := make([]closureexec.DerivationOutput, len(request.Permit.ExpectedEvidence))
	evidence := make([]string, len(request.Permit.ExpectedEvidence))
	for index, expected := range request.Permit.ExpectedEvidence {
		payload, readErr := os.ReadFile(filepath.Join(result.OutputRoot, filepath.FromSlash(expected.Path))) // #nosec G304 -- exact provider-owned evidence path.
		if readErr != nil {
			return closureexec.Audit{}, readErr
		}
		evidence[index] = expected.Path
		outputs[index] = closureexec.DerivationOutput{Path: expected.Path, SchemaID: expected.SchemaID, ArtifactManifestID: expected.ArtifactManifestID, SHA256: digestID(payload), Size: int64(len(payload))}
	}
	processes := append([]string(nil), request.Permit.AllowedProcesses...)
	if provider.extraProcess != "" {
		processes = append(processes, provider.extraProcess)
		sort.Strings(processes)
	}
	audit := closureexec.Audit{Executable: request.Permit.Executable, CWD: request.Permit.CWD, Argv: append([]string(nil), request.Permit.Argv...), Environment: cloneStrings(request.Permit.Environment), Processes: processes, Reads: append([]string(nil), request.Permit.ReadRoots...), Writes: append([]string(nil), request.Permit.WriteRoots...), Evidence: evidence, Network: "none", ExitCode: 0, Outputs: outputs}
	if provider.networkAttempt {
		audit.Network = "attempted:tcp"
	}
	if observed != nil {
		expectedExecutable := filepath.Join(provider.execution.(*concreteNPMRunner).executionRoot, filepath.FromSlash(request.Permit.Executable))
		expectedCWD := filepath.Join(provider.execution.(*concreteNPMRunner).executionRoot, filepath.FromSlash(request.Permit.CWD))
		if observed.Executable != expectedExecutable || observed.CWD != expectedCWD || !reflect.DeepEqual(observed.Argv, request.Permit.Argv) {
			return closureexec.Audit{}, &closureexec.DiagnosticError{Code: CodeProcessUndeclared, Detail: "real launch differs from exact permit"}
		}
		observedEnvironment := map[string]string{}
		for _, item := range observed.Environment {
			parts := strings.SplitN(item, "=", 2)
			if len(parts) != 2 {
				return closureexec.Audit{}, fmt.Errorf("malformed observed environment")
			}
			observedEnvironment[parts[0]] = parts[1]
		}
		audit.Environment = observedEnvironment
	}
	return audit, nil
}

func extractFixtureTarball(item capturedInput, destination string) error {
	input, err := item.handle.Open()
	if err != nil {
		return err
	}
	defer func() { _ = input.Close() }()
	gz, err := gzip.NewReader(input)
	if err != nil {
		return err
	}
	defer func() { _ = gz.Close() }()
	reader := tar.NewReader(gz)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(filepath.ToSlash(filepath.Clean(header.Name)), "package/")
		if name == "package" || name == "" {
			continue
		}
		target := filepath.Join(destination, filepath.FromSlash(name))
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(target, 0o700); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		payload, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		mode := os.FileMode(0o600)
		if header.Mode&0o111 != 0 {
			mode = 0o700
		}
		if err := os.WriteFile(target, payload, mode); err != nil {
			return err
		}
	}
}
func replaceTarballAndLock(t *testing.T, fixture *npmFixture, installPath string, pkg fixturePackage) {
	t.Helper()
	payload := buildTGZ(t, pkg)
	if err := os.WriteFile(fixture.tarballs[installPath].Path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	var lock map[string]any
	if err := json.Unmarshal(fixture.request.LockBytes, &lock); err != nil {
		t.Fatal(err)
	}
	entries := lock["packages"].(map[string]any)
	entry := entries[installPath].(map[string]any)
	sum := sha512.Sum512(payload)
	entry["integrity"] = "sha512-" + base64.StdEncoding.EncodeToString(sum[:])
	for key := range entry {
		if key != "name" && key != "version" && key != "resolved" && key != "integrity" && key != "optional" {
			delete(entry, key)
		}
	}
	for key, value := range pkg.metadata {
		if key == "hasInstallScript" {
			entry[key] = value
			continue
		}
		if key == "scripts" || key == "bundleDependencies" {
			continue
		}
		entry[key] = value
	}
	fixture.request.LockBytes = mustJSON(t, lock)
	fixture.payloads[installPath] = payload
	if err := os.WriteFile(filepath.Join(fixture.root, "package-lock.json"), fixture.request.LockBytes, 0o600); err != nil {
		t.Fatal(err)
	}
}
func buildTGZ(t *testing.T, pkg fixturePackage) []byte {
	t.Helper()
	metadata := map[string]any{"name": pkg.name, "version": pkg.version}
	for key, value := range pkg.metadata {
		if key != "hasInstallScript" {
			metadata[key] = value
		}
	}
	files := map[string][]byte{"package/package.json": mustJSON(t, metadata)}
	for name, payload := range pkg.files {
		files["package/"+name] = payload
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	var output bytes.Buffer
	gz := gzip.NewWriter(&output)
	tw := tar.NewWriter(gz)
	for _, name := range names {
		payload := files[name]
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(payload))}); err != nil {
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
		t.Fatalf("got error %v code %q, want %q", err, ErrorCode(err), want)
	}
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
func argAfter(args []string, name string) string {
	for i, value := range args {
		if value == name && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}
func makeTreeWritable(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return os.Chmod(path, 0o700)
		}
		return os.Chmod(path, 0o600)
	})
}
