package crossconformance_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/crossconformance"
	"github.com/relux-works/curator/internal/nodesource"
	"github.com/relux-works/curator/internal/npmsource"
	"github.com/relux-works/curator/internal/pnpmsource"
	"github.com/relux-works/curator/internal/yarnclassicsource"
	"github.com/relux-works/curator/internal/yarnmodernsource"
)

// nodeCapture is the manager-neutral result of running one Node manager's
// production lock parser and capture/admission over the shared fixture.
type nodeCapture struct {
	path     crossconformance.PathID
	capture  nodesource.Capture
	receipts []string
}

// nodeTargets are the two exact destinations every Node path is projected onto.
var nodeTargets = []struct {
	label, os, architecture string
}{
	{label: "darwin-arm64", os: "darwin", architecture: "arm64"},
	{label: "linux-x64", os: "linux", architecture: "x64"},
}

func nodePlatform(osName, architecture string) closuregraph.TargetPlatformPayload {
	return closuregraph.TargetPlatformPayload{
		OS: osName, Architecture: architecture, ABI: "node", Libc: "none", MinimumRuntime: "22",
		SDKID: "none", TargetTriple: architecture + "-" + osName, Runtime: "node",
		LanguageModes: map[string]string{"module": "commonjs"}, Tuning: map[string]string{},
	}
}

func nodeTool(role, executable, seed string, domain closuregraph.ExecutionDomain) nodesource.ToolIdentity {
	return nodesource.ToolIdentity{
		Role: role, PolicySelector: role + "-v1", ExecutableRelativePath: executable,
		VersionOutput: role + " 1.0", PlatformABI: "node",
		Fingerprint: digestID([]byte(seed)), ExecutableSHA256: digestID([]byte(seed + "-executable")),
		ExecutionDomain: domain,
	}
}

// projectNodePath closes one Node capture onto one exact destination through
// the common Node contract every manager funnels into, and reports the shared
// TargetProjection the normative suite consumes.
func projectNodePath(t *testing.T, source nodeCapture, targetIndex int) crossconformance.TargetProjection {
	t.Helper()
	target := nodeTargets[targetIndex]
	capture := source.capture
	roots := capture.Graph.RootNodeIDs
	if len(roots) != 1 {
		t.Fatalf("%s capture has %d roots, want one command product", source.path, len(roots))
	}
	platform := nodePlatform(target.os, target.architecture)
	platformNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: platform}
	platformID, err := platformNode.ID()
	if err != nil {
		t.Fatal(err)
	}
	selection, err := closuregraph.NewSelectionContext(
		[]closuregraph.ID{roots[0]},
		map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID},
		[]string{}, false, map[string]string{"os": target.os, "cpu": target.architecture}, map[string]string{}, []string{},
	)
	if err != nil {
		t.Fatal(err)
	}
	runtime := nodesource.RuntimeBinding{
		Platform:      platform,
		Node:          nodeTool("node-runtime", "bin/node", "node-"+target.label, closuregraph.ExecutionTarget),
		Manager:       nodeTool("package-manager", "bin/manager", string(source.path)+"-"+target.label, closuregraph.ExecutionTarget),
		TargetNodeIDs: []closuregraph.ID{roots[0]},
	}
	c0, err := nodesource.NewC0Checkpoint(capture, selection, runtime)
	if err != nil {
		t.Fatalf("%s C0: %v", source.path, err)
	}
	runtime.C0Checkpoint = &c0
	bundle, plan, err := nodesource.Close(capture, selection, runtime, nil, "cross-conformance-execution-v1")
	if err != nil {
		t.Fatalf("%s close: %v", source.path, err)
	}
	return nodeProjection(t, source, target.label, bundle, plan, c0, platformID, runtime)
}

func nodeProjection(t *testing.T, source nodeCapture, label string, bundle closuregraph.GraphBundle, plan closuregraph.BuildPlan, c0 closuregraph.Checkpoint, platformID closuregraph.ID, runtime nodesource.RuntimeBinding) crossconformance.TargetProjection {
	t.Helper()
	projection := crossconformance.TargetProjection{
		Path:                   source.path,
		TargetLabel:            label,
		CaptureIdentity:        identityOf(t, bundle.Capture),
		SelectionIdentity:      identityOf(t, bundle.Selection),
		BindingIdentity:        identityOf(t, bundle.Binding),
		ActiveIdentity:         identityOf(t, bundle.Active),
		PlanIdentity:           identityOf(t, plan),
		CaptureNodeKinds:       nodeKindCensus(t, bundle.Records.CaptureNodes),
		CaptureEdgeKinds:       edgeKindCensus(t, bundle.Records.CaptureEdges),
		BindingNodeKinds:       nodeKindCensus(t, bundle.Records.BindingNodes),
		BindingEdgeKinds:       edgeKindCensus(t, bundle.Records.BindingEdges),
		TargetPlatformIdentity: string(platformID),
		ExplicitTargetEdges:    explicitTargetEdges(t, bundle.Records.BindingEdges, platformID),
		EmitsBindingRecords:    true,
		ToolIdentities:         []string{string(runtime.Node.Fingerprint), string(runtime.Manager.Fingerprint)},
		Checkpoints:            []crossconformance.CheckpointLink{{Name: string(c0.Name), Identity: identityOf(t, c0)}},
		DerivationReceipts:     source.receipts,
	}
	return projection
}

// nodePathRecords returns one Node path's real closed graph so the rejection
// matrix can tamper with exactly the records the adapter produced.
func nodePathRecords(t *testing.T, source nodeCapture) pathRecords {
	t.Helper()
	target := nodeTargets[0]
	roots := source.capture.Graph.RootNodeIDs
	platform := nodePlatform(target.os, target.architecture)
	platformNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: platform}
	platformID, err := platformNode.ID()
	if err != nil {
		t.Fatal(err)
	}
	selection, err := closuregraph.NewSelectionContext(
		[]closuregraph.ID{roots[0]},
		map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID},
		[]string{}, false, map[string]string{"os": target.os, "cpu": target.architecture}, map[string]string{}, []string{},
	)
	if err != nil {
		t.Fatal(err)
	}
	runtime := nodesource.RuntimeBinding{
		Platform:      platform,
		Node:          nodeTool("node-runtime", "bin/node", "node-"+target.label, closuregraph.ExecutionTarget),
		Manager:       nodeTool("package-manager", "bin/manager", string(source.path)+"-"+target.label, closuregraph.ExecutionTarget),
		TargetNodeIDs: []closuregraph.ID{roots[0]},
	}
	c0, err := nodesource.NewC0Checkpoint(source.capture, selection, runtime)
	if err != nil {
		t.Fatal(err)
	}
	runtime.C0Checkpoint = &c0
	binding, bindingNodes, bindingEdges, authority, err := nodesource.Bind(source.capture, selection, runtime)
	if err != nil {
		t.Fatal(err)
	}
	return pathRecords{
		capture:   source.capture.Graph,
		selection: selection,
		binding:   binding,
		records:   closuregraph.NewRecordTables(source.capture.Nodes, source.capture.Edges, bindingNodes, bindingEdges),
		authority: authority,
	}
}

type identifiable interface {
	ID() (closuregraph.ID, error)
}

func identityOf(t *testing.T, record identifiable) string {
	t.Helper()
	id, err := record.ID()
	if err != nil {
		t.Fatal(err)
	}
	return string(id)
}

func nodeKindCensus(t *testing.T, nodes []closuregraph.Node) map[string]int {
	t.Helper()
	census := map[string]int{}
	for _, node := range nodes {
		census[string(node.Kind)]++
	}
	return census
}

func edgeKindCensus(t *testing.T, edges []closuregraph.Edge) map[string]int {
	t.Helper()
	census := map[string]int{}
	for _, edge := range edges {
		census[string(edge.Kind)]++
	}
	return census
}

func explicitTargetEdges(t *testing.T, edges []closuregraph.Edge, platformID closuregraph.ID) int {
	t.Helper()
	count := 0
	for _, edge := range edges {
		if edge.Kind == closuregraph.EdgeTargets && edge.ToNodeID == platformID {
			count++
		}
	}
	return count
}

func newCaptureStore(t *testing.T) *closureexec.CaptureStore {
	t.Helper()
	root := filepath.Join(t.TempDir(), "capture-store")
	store, err := closureexec.NewCaptureStore(root)
	if err != nil {
		t.Fatal(err)
	}
	releaseTree(t, root)
	return store
}

// --- npm ---------------------------------------------------------------------

type npmFixture struct {
	root, work string
	request    npmsource.ParseRequest
	tarballs   map[string]npmsource.RawTarball
	payloads   map[string][]byte
}

func newNPMFixture(t *testing.T, extra map[string][]byte) npmFixture {
	t.Helper()
	root := filepath.Join(t.TempDir(), "project")
	work := filepath.Join(t.TempDir(), "work")
	rootManifest := mustJSON(t, map[string]any{"name": "app", "version": "1.0.0", "dependencies": map[string]string{"a": "1.0.0"}})
	lockPackages := map[string]any{"": map[string]any{"name": "app", "version": "1.0.0", "dependencies": map[string]string{"a": "1.0.0"}}}
	tarballs := map[string]npmsource.RawTarball{}
	payloads := map[string][]byte{}
	for _, pkg := range crossPackages() {
		if pkg.name == "opt" {
			continue
		}
		payload := buildTGZ(t, withExtraFiles(pkg, extra))
		install := "node_modules/" + pkg.name
		payloads[install] = payload
		lockPackages[install] = map[string]any{
			"name": pkg.name, "version": pkg.version,
			"resolved":  fmt.Sprintf("https://registry.npmjs.org/%s/-/%s-%s.tgz", pkg.name, pkg.name, pkg.version),
			"integrity": sriSHA512(payload),
			"dependencies": func() map[string]string {
				if declared, ok := pkg.metadata["dependencies"].(map[string]string); ok {
					return declared
				}
				return nil
			}(),
		}
		tarballs[install] = npmsource.RawTarball{Path: writeTemp(t, pkg.name+".tgz", payload)}
	}
	lock := mustJSON(t, map[string]any{"name": "app", "version": "1.0.0", "lockfileVersion": 3, "packages": lockPackages})
	writeFile(t, filepath.Join(root, "package.json"), rootManifest)
	writeFile(t, filepath.Join(root, "package-lock.json"), lock)
	writeFile(t, filepath.Join(root, "index.js"), []byte("require('a')\n"))
	return npmFixture{
		root: root, work: work, tarballs: tarballs, payloads: payloads,
		request: npmsource.ParseRequest{
			LockName: "package-lock.json", LockBytes: lock,
			Manifests:              map[string][]byte{"package.json": rootManifest},
			AllowedRegistryOrigins: []string{"https://registry.npmjs.org/"},
			Target:                 npmsource.Target{OS: "darwin", Architecture: "arm64", Libc: "none", IncludeDev: true},
		},
	}
}

func captureNPM(t *testing.T, fixture npmFixture) (*npmsource.Capture, error) {
	t.Helper()
	return npmsource.CaptureAndAdmit(context.Background(), npmsource.CaptureRequest{
		Graph: mustNPMGraph(t, fixture), ProjectRoot: fixture.root, Tarballs: fixture.tarballs,
		WorkRoot: fixture.work, Store: newCaptureStore(t), Policy: artifactpolicy.NewService(),
		PreviousCausalHead: "cross-conformance-head",
	})
}

func mustNPMGraph(t *testing.T, fixture npmFixture) npmsource.Graph {
	t.Helper()
	graph, err := npmsource.Parse(fixture.request)
	if err != nil {
		t.Fatal(err)
	}
	return graph
}

func npmCapture(t *testing.T) nodeCapture {
	t.Helper()
	capture, err := captureNPM(t, newNPMFixture(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	return nodeCapture{path: crossconformance.PathNPM, capture: capture.NodeCapture, receipts: intakeReceipts(capture.Evidence.Project, capture.Evidence.Tarballs)}
}

func intakeReceipts(project npmsource.ArtifactEvidence, tarballs []npmsource.ArtifactEvidence) []string {
	receipts := []string{string(project.IntakeReceiptID)}
	for _, item := range tarballs {
		receipts = append(receipts, string(item.IntakeReceiptID))
	}
	return receipts
}

// --- pnpm --------------------------------------------------------------------

type pnpmFixture struct {
	root, work string
	request    pnpmsource.ParseRequest
	tarballs   map[string]pnpmsource.RawTarball
	payloads   map[string][]byte
}

func newPNPMFixture(t *testing.T, extra map[string][]byte) pnpmFixture {
	t.Helper()
	root := filepath.Join(t.TempDir(), "project")
	work := filepath.Join(t.TempDir(), "work")
	rootManifest := mustJSON(t, map[string]any{"name": "app", "version": "1.0.0", "dependencies": map[string]string{"a": "^1.0.0"}})
	tarballs := map[string]pnpmsource.RawTarball{}
	payloads := map[string][]byte{}
	integrity := map[string]string{}
	for _, pkg := range crossPackages() {
		if pkg.name == "opt" {
			continue
		}
		payload := buildTGZ(t, withExtraFiles(pkg, extra))
		key := pkg.name + "@" + pkg.version
		payloads[key] = payload
		integrity[key] = sriSHA512(payload)
		tarballs[key] = pnpmsource.RawTarball{Path: writeTemp(t, pkg.name+".tgz", payload)}
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
        version: 1.0.0
packages:
  a@1.0.0:
    resolution:
      integrity: %s
  b@1.0.0:
    resolution:
      integrity: %s
snapshots:
  a@1.0.0:
    dependencies:
      b: 1.0.0
  b@1.0.0: {}
`, integrity["a@1.0.0"], integrity["b@1.0.0"])
	npmrc := []byte("ignore-scripts=true\nside-effects-cache=false\n")
	writeFile(t, filepath.Join(root, "package.json"), rootManifest)
	writeFile(t, filepath.Join(root, "pnpm-lock.yaml"), []byte(lock))
	writeFile(t, filepath.Join(root, ".npmrc"), npmrc)
	writeFile(t, filepath.Join(root, "index.js"), []byte("require('a')\n"))
	return pnpmFixture{
		root: root, work: work, tarballs: tarballs, payloads: payloads,
		request: pnpmsource.ParseRequest{
			LockBytes: []byte(lock), Manifests: map[string][]byte{"package.json": rootManifest},
			ConfigFiles:            map[string][]byte{".npmrc": npmrc},
			AllowedRegistryOrigins: []string{"https://registry.npmjs.org"},
			Target:                 pnpmsource.Target{OS: "darwin", Architecture: "arm64", Libc: "none", IncludeDev: true},
		},
	}
}

func capturePNPM(t *testing.T, fixture pnpmFixture) (*pnpmsource.Capture, error) {
	t.Helper()
	graph, err := pnpmsource.Parse(fixture.request)
	if err != nil {
		t.Fatal(err)
	}
	return pnpmsource.CaptureAndAdmit(context.Background(), pnpmsource.CaptureRequest{
		Graph: graph, ProjectRoot: fixture.root, Tarballs: fixture.tarballs, WorkRoot: fixture.work,
		Store: newCaptureStore(t), Policy: artifactpolicy.NewService(), PreviousCausalHead: "cross-conformance-head",
	})
}

func pnpmCapture(t *testing.T) nodeCapture {
	t.Helper()
	capture, err := capturePNPM(t, newPNPMFixture(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	receipts := []string{}
	for _, group := range [][]pnpmsource.ArtifactEvidence{capture.Evidence.LocalRoots, capture.Evidence.Tarballs} {
		for _, item := range group {
			receipts = append(receipts, string(item.IntakeReceiptID))
		}
	}
	return nodeCapture{path: crossconformance.PathPNPM, capture: capture.NodeCapture, receipts: receipts}
}

// --- Yarn Classic ------------------------------------------------------------

type yarnClassicFixture struct {
	root, work string
	request    yarnclassicsource.ParseRequest
	tarballs   map[string]yarnclassicsource.RawTarball
	payloads   map[string][]byte
}

func newYarnClassicFixture(t *testing.T, extra map[string][]byte) yarnClassicFixture {
	t.Helper()
	root := filepath.Join(t.TempDir(), "project")
	work := filepath.Join(t.TempDir(), "work")
	rootManifest := mustJSON(t, map[string]any{"name": "app", "version": "1.0.0", "private": true, "dependencies": map[string]string{"a": "^1.0.0"}})
	entries := []string{}
	paths := map[string]string{}
	payloads := map[string][]byte{}
	for _, pkg := range crossPackages() {
		if pkg.name == "opt" {
			continue
		}
		payload := buildTGZ(t, withExtraFiles(pkg, extra))
		entry := fmt.Sprintf("%s@^1.0.0:\n  version \"%s\"\n  resolved \"https://registry.yarnpkg.com/%s/-/%s-%s.tgz#pinned\"\n  integrity %s\n",
			pkg.name, pkg.version, pkg.name, pkg.name, pkg.version, sriSHA512(payload))
		if declared, ok := pkg.metadata["dependencies"].(map[string]string); ok {
			entry += "  dependencies:\n"
			for _, name := range sortedKeys(declared) {
				entry += fmt.Sprintf("    %s \"%s\"\n", name, declared[name])
			}
		}
		entries = append(entries, entry)
		paths[pkg.name] = writeTemp(t, pkg.name+"-1.0.0.tgz", payload)
		payloads[pkg.name] = payload
	}
	lock := []byte("# THIS IS AN AUTOGENERATED FILE. DO NOT EDIT THIS FILE DIRECTLY.\n# yarn lockfile v1\n\n" + strings.Join(entries, "\n"))
	writeFile(t, filepath.Join(root, "package.json"), rootManifest)
	writeFile(t, filepath.Join(root, "yarn.lock"), lock)
	writeFile(t, filepath.Join(root, "index.js"), []byte("require('a')\n"))
	request := yarnclassicsource.ParseRequest{
		LockName: "yarn.lock", LockBytes: lock, Manifests: map[string][]byte{"package.json": rootManifest},
		YarnVersion: yarnclassicsource.SupportedYarnVersion, AllowedRegistryOrigins: []string{"https://registry.yarnpkg.com/"},
		Target: yarnclassicsource.Target{OS: "darwin", Architecture: "arm64", Libc: "none", IncludeDev: true},
	}
	graph, err := yarnclassicsource.Parse(request)
	if err != nil {
		t.Fatal(err)
	}
	tarballs := map[string]yarnclassicsource.RawTarball{}
	keyed := map[string][]byte{}
	for _, pkg := range graph.Packages {
		if pkg.Resolved == "" {
			continue
		}
		tarballs[pkg.Key] = yarnclassicsource.RawTarball{Path: paths[pkg.Name]}
		keyed[pkg.Key] = payloads[pkg.Name]
	}
	return yarnClassicFixture{root: root, work: work, request: request, tarballs: tarballs, payloads: keyed}
}

func captureYarnClassic(t *testing.T, fixture yarnClassicFixture) (*yarnclassicsource.Capture, error) {
	t.Helper()
	graph, err := yarnclassicsource.Parse(fixture.request)
	if err != nil {
		t.Fatal(err)
	}
	return yarnclassicsource.CaptureAndAdmit(context.Background(), yarnclassicsource.CaptureRequest{
		Graph: graph, ProjectRoot: fixture.root, Tarballs: fixture.tarballs, WorkRoot: fixture.work,
		Store: newCaptureStore(t), Policy: artifactpolicy.NewService(), PreviousCausalHead: "cross-conformance-head",
	})
}

func yarnClassicCapture(t *testing.T) nodeCapture {
	t.Helper()
	capture, err := captureYarnClassic(t, newYarnClassicFixture(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	receipts := []string{string(capture.Evidence.Project.IntakeReceiptID)}
	for _, item := range capture.Evidence.Tarballs {
		receipts = append(receipts, string(item.IntakeReceiptID))
	}
	return nodeCapture{path: crossconformance.PathYarnClassic, capture: capture.NodeCapture, receipts: receipts}
}

// --- modern Yarn -------------------------------------------------------------

type yarnModernFixture struct {
	root, work string
	request    yarnmodernsource.ParseRequest
	archives   map[string]yarnmodernsource.RawArchive
	payloads   map[string][]byte
}

func modernRC() []byte {
	return []byte("nodeLinker: pnp\ncompressionLevel: 0\ncacheFolder: .yarn/cache\nenableGlobalCache: false\nenableNetwork: false\nenableImmutableInstalls: true\nenableScripts: false\nchecksumBehavior: throw\npnpMode: strict\n")
}

func newYarnModernFixture(t *testing.T, extra map[string][]byte) yarnModernFixture {
	t.Helper()
	root := filepath.Join(t.TempDir(), "project")
	work := filepath.Join(t.TempDir(), "work")
	manifest := []byte(`{"name":"app","version":"1.0.0","private":true,"packageManager":"yarn@4.9.2","dependencies":{"a":"npm:^1.0.0"}}`)
	rc := modernRC()
	entries := []string{}
	archives := map[string]yarnmodernsource.RawArchive{}
	payloads := map[string][]byte{}
	details := map[string]struct {
		checksum, path string
	}{}
	for _, pkg := range crossPackages() {
		if pkg.name == "opt" {
			continue
		}
		// Modern Yarn reconciles the embedded manifest against the lock
		// verbatim, and a modern lock records dependency ranges with their
		// resolution protocol. The payload bytes are otherwise the shared
		// fixture; only this manager-specific spelling differs.
		payload := cacheZIP(t, withExtraFiles(withProtocolRanges(pkg), extra))
		checksum := yarnmodernsource.CacheChecksum(payload, "10c0")
		details[pkg.name] = struct{ checksum, path string }{checksum: checksum, path: writeTemp(t, pkg.name+".zip", payload)}
		payloads[pkg.name] = payload
		entry := fmt.Sprintf("\"%s@npm:^%s\":\n  version: %s\n  resolution: \"%s@npm:%s\"\n", pkg.name, pkg.version, pkg.version, pkg.name, pkg.version)
		if declared, ok := pkg.metadata["dependencies"].(map[string]string); ok {
			entry += "  dependencies:\n"
			for _, name := range sortedKeys(declared) {
				entry += fmt.Sprintf("    %s: \"npm:%s\"\n", name, declared[name])
			}
		}
		entry += "  checksum: \"" + checksum + "\"\n  languageName: node\n  linkType: hard\n"
		entries = append(entries, entry)
	}
	lock := "# This file is generated by running \"yarn install\" inside your project.\n# Manual changes might be lost - proceed with caution!\n\n__metadata:\n  version: 8\n  cacheKey: 10c0\n\n\"app@workspace:.\":\n  version: 0.0.0-use.local\n  resolution: \"app@workspace:.\"\n  dependencies:\n    a: \"npm:^1.0.0\"\n  languageName: unknown\n  linkType: soft\n\n" + strings.Join(entries, "\n")
	writeFile(t, filepath.Join(root, "package.json"), manifest)
	writeFile(t, filepath.Join(root, "yarn.lock"), []byte(lock))
	writeFile(t, filepath.Join(root, ".yarnrc.yml"), rc)
	writeFile(t, filepath.Join(root, "index.js"), []byte("require('a')\n"))
	request := yarnmodernsource.ParseRequest{
		LockName: "yarn.lock", LockBytes: []byte(lock), Manifests: map[string][]byte{"package.json": manifest},
		Configuration: map[string][]byte{".yarnrc.yml": rc}, YarnVersion: yarnmodernsource.SupportedYarnVersion,
		Target: yarnmodernsource.Target{OS: "darwin", Architecture: "arm64", Libc: "none", IncludeDev: true},
	}
	graph, err := yarnmodernsource.Parse(request)
	if err != nil {
		t.Fatal(err)
	}
	keyed := map[string][]byte{}
	for _, pkg := range graph.Packages {
		if pkg.BaseKey != "" || pkg.Resolution == "" || pkg.Checksum == "" {
			continue
		}
		detail, ok := details[pkg.Name]
		if !ok {
			t.Fatalf("modern Yarn graph names unknown package %q", pkg.Name)
		}
		archives[pkg.Key] = yarnmodernsource.RawArchive{Path: detail.path, Format: "zip", SHA256: string(digestID(payloads[pkg.Name])), YarnChecksum: detail.checksum}
		keyed[pkg.Key] = payloads[pkg.Name]
	}
	return yarnModernFixture{root: root, work: work, request: request, archives: archives, payloads: keyed}
}

func captureYarnModern(t *testing.T, fixture yarnModernFixture) (*yarnmodernsource.Capture, error) {
	t.Helper()
	graph, err := yarnmodernsource.Parse(fixture.request)
	if err != nil {
		t.Fatal(err)
	}
	return yarnmodernsource.CaptureAndAdmit(context.Background(), yarnmodernsource.CaptureRequest{
		Graph: graph, ProjectRoot: fixture.root, Archives: fixture.archives, WorkRoot: fixture.work,
		Store: newCaptureStore(t), Policy: artifactpolicy.NewService(), PreviousCausalHead: "cross-conformance-head",
	})
}

func yarnModernCapture(t *testing.T) nodeCapture {
	t.Helper()
	capture, err := captureYarnModern(t, newYarnModernFixture(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	receipts := []string{string(capture.Evidence.Project.IntakeReceiptID)}
	for _, item := range capture.Evidence.Tarballs {
		receipts = append(receipts, string(item.IntakeReceiptID))
	}
	return nodeCapture{path: crossconformance.PathYarnModern, capture: capture.NodeCapture, receipts: receipts}
}

// withProtocolRanges rewrites the shared package's declared ranges into the
// modern Yarn "npm:" descriptor spelling its lock and manifests both use.
func withProtocolRanges(pkg sourcePackage) sourcePackage {
	declared, ok := pkg.metadata["dependencies"].(map[string]string)
	if !ok {
		return pkg
	}
	rewritten := map[string]string{}
	for name, value := range declared {
		rewritten[name] = "npm:" + value
	}
	metadata := map[string]any{}
	for key, value := range pkg.metadata {
		metadata[key] = value
	}
	metadata["dependencies"] = rewritten
	return sourcePackage{name: pkg.name, version: pkg.version, metadata: metadata, files: pkg.files}
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	for index := 1; index < len(keys); index++ {
		for inner := index; inner > 0 && keys[inner] < keys[inner-1]; inner-- {
			keys[inner], keys[inner-1] = keys[inner-1], keys[inner]
		}
	}
	return keys
}

// nodeCycleOutcome builds two declared generators over the path's real capture
// where each consumes the other's output, and requires the shared execution
// projection to reject the cycle before any plan exists.
func nodeCycleOutcome(t *testing.T, source nodeCapture) crossconformance.RejectionOutcome {
	t.Helper()
	base := source.capture
	root := base.Graph.RootNodeIDs[0]
	sourceNode := anySourceNodeID(t, base)
	compiler := nodeTool("typescript-compiler", "bin/tsc", "tsc-"+string(source.path), closuregraph.ExecutionHost)
	first := nodesource.GeneratedAction{
		Name: "first", Argv: []string{"--emit"}, WorkingDirectory: "workspace", Compiler: compiler,
		Inputs:       []nodesource.GeneratedInput{{NodeID: sourceNode, Path: "workspace/src", Class: "source.tree", Role: "source"}},
		TargetNodeID: root, EnvironmentPolicyID: "node-env-v1", ProcessPolicyID: "node-process-v1",
		Outputs: []nodesource.GeneratedOutput{{Path: "gen/intermediate.js", Class: "source.generated_text", Grammar: "javascript-v1", Role: "intermediate", Intermediate: true}},
	}
	_, firstOutputs, _, err := nodesource.BuildGeneratedAction(first)
	if err != nil {
		t.Fatal(err)
	}
	intermediateID, err := firstOutputs[0].ID()
	if err != nil {
		t.Fatal(err)
	}
	second := nodesource.GeneratedAction{
		Name: "second", Argv: []string{"--bundle"}, WorkingDirectory: "workspace", Compiler: compiler,
		Inputs:       []nodesource.GeneratedInput{{NodeID: intermediateID, Path: "gen/intermediate.js", Class: "source.generated_text", Role: "source"}},
		TargetNodeID: root, EnvironmentPolicyID: "node-env-v1", ProcessPolicyID: "node-process-v1",
		Outputs: []nodesource.GeneratedOutput{{Path: "dist/cli.js", Class: "source.generated_text", Grammar: "javascript-v1", Role: "published_command"}},
	}
	_, secondOutputs, _, err := nodesource.BuildGeneratedAction(second)
	if err != nil {
		t.Fatal(err)
	}
	secondOutputID, err := secondOutputs[0].ID()
	if err != nil {
		t.Fatal(err)
	}
	first.Inputs = []nodesource.GeneratedInput{{NodeID: secondOutputID, Path: "dist/cli.js", Class: "source.generated_text", Role: "source"}}
	cyclic, err := nodesource.AddGeneratedActions(base, []nodesource.GeneratedAction{first, second})
	if err != nil {
		t.Fatal(err)
	}
	cycleSource := nodeCapture{path: source.path, capture: cyclic, receipts: source.receipts}
	err = closeNodeCaptureExpectingFailure(t, cycleSource)
	return crossconformance.RejectionOutcome{Vector: "build-cycle", Path: source.path, Err: err, Code: diagnosticCode(err)}
}

// nodeUndeclaredInputOutcome declares a generator input that is in no admitted
// capture record and requires the common Node contract to refuse it.
func nodeUndeclaredInputOutcome(t *testing.T, source nodeCapture) crossconformance.RejectionOutcome {
	t.Helper()
	root := source.capture.Graph.RootNodeIDs[0]
	spec := nodesource.GeneratedAction{
		Name: "hidden", Argv: []string{"--generate"}, WorkingDirectory: "workspace",
		Compiler:     nodeTool("typescript-compiler", "bin/tsc", "tsc-hidden", closuregraph.ExecutionHost),
		Inputs:       []nodesource.GeneratedInput{{NodeID: digestID([]byte("ambient-host-input")), Path: "host/ambient.json", Class: "source.config", Role: "config"}},
		TargetNodeID: root, EnvironmentPolicyID: "node-env-v1", ProcessPolicyID: "node-process-v1",
		Outputs: []nodesource.GeneratedOutput{{Path: "dist/hidden.js", Class: "source.generated_text", Grammar: "javascript-v1"}},
	}
	_, err := nodesource.AddGeneratedActions(source.capture, []nodesource.GeneratedAction{spec})
	return crossconformance.RejectionOutcome{Vector: "undeclared-input", Path: source.path, Err: err, Code: diagnosticCode(err)}
}

func closeNodeCaptureExpectingFailure(t *testing.T, source nodeCapture) error {
	t.Helper()
	target := nodeTargets[0]
	platform := nodePlatform(target.os, target.architecture)
	platformNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: platform}
	platformID, err := platformNode.ID()
	if err != nil {
		t.Fatal(err)
	}
	root := source.capture.Graph.RootNodeIDs[0]
	selection, err := closuregraph.NewSelectionContext(
		[]closuregraph.ID{root},
		map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID},
		[]string{}, false, map[string]string{"os": target.os, "cpu": target.architecture}, map[string]string{}, []string{},
	)
	if err != nil {
		t.Fatal(err)
	}
	runtime := nodesource.RuntimeBinding{
		Platform:      platform,
		Node:          nodeTool("node-runtime", "bin/node", "node-"+target.label, closuregraph.ExecutionTarget),
		Manager:       nodeTool("package-manager", "bin/manager", string(source.path)+"-"+target.label, closuregraph.ExecutionTarget),
		TargetNodeIDs: []closuregraph.ID{root},
	}
	c0, err := nodesource.NewC0Checkpoint(source.capture, selection, runtime)
	if err != nil {
		t.Fatal(err)
	}
	runtime.C0Checkpoint = &c0
	_, _, err = nodesource.Close(source.capture, selection, runtime, nil, "cross-conformance-execution-v1")
	return err
}

func anySourceNodeID(t *testing.T, capture nodesource.Capture) closuregraph.ID {
	t.Helper()
	keys := make([]string, 0, len(capture.SourceNodeIDs))
	for key := range capture.SourceNodeIDs {
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		t.Fatal("capture has no source node")
	}
	for index := 1; index < len(keys); index++ {
		for inner := index; inner > 0 && keys[inner] < keys[inner-1]; inner-- {
			keys[inner], keys[inner-1] = keys[inner-1], keys[inner]
		}
	}
	return capture.SourceNodeIDs[keys[0]]
}
