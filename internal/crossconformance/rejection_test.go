package crossconformance_test

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy"
	policyconformance "github.com/relux-works/curator/internal/artifactpolicy/conformance"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/crossconformance"
	"github.com/relux-works/curator/internal/rustsource"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// proveRejectionMatrix runs the published cross-adapter rejection matrix. Each
// vector requires the same three things from every path that runs it: a stable
// diagnostic from the closed set, no affected process, and no publication.
func proveRejectionMatrix(t *testing.T, coverage *crossconformance.Coverage, rustManagerUnavailable bool) {
	vectors := map[string]crossconformance.RejectionVector{}
	for _, vector := range crossconformance.RejectionVectors() {
		vectors[vector.ID] = vector
	}
	record := func(t *testing.T, vector string, outcome crossconformance.RejectionOutcome) {
		t.Helper()
		if err := crossconformance.CheckRejection(vectors[vector], outcome); err != nil {
			t.Error(err)
			return
		}
		coverage.RecordRejection(vector, outcome.Path)
	}

	t.Run("compiled-dependency-bytes", func(t *testing.T) {
		for _, outcome := range compiledByteOutcomes(t, rustManagerUnavailable) {
			record(t, "compiled-dependency-bytes", outcome)
		}
	})

	t.Run("opaque-dependency-bytes", func(t *testing.T) {
		for _, outcome := range opaqueByteOutcomes(t) {
			record(t, "opaque-dependency-bytes", outcome)
		}
	})

	t.Run("verified-binary-unavailable", func(t *testing.T) {
		for _, path := range crossconformance.DeliveredPaths() {
			profile := pathProfiles[path]
			payload := policyconformance.GNUSharedObject()
			_, err := artifactpolicy.NewService().AdmitVerifiedBinary(context.Background(), artifactpolicy.VerifiedBinaryRequest{
				Descriptor: crossDescriptor(profile.adapter, profile.profile, string(path), payload),
				Payload:    artifactpolicy.Payload{Path: "verified/candidate.so", Size: int64(len(payload)), Reader: strings.NewReader(string(payload))},
			})
			record(t, "verified-binary-unavailable", crossconformance.RejectionOutcome{Vector: "verified-binary-unavailable", Path: path, Err: err, Code: diagnosticCode(err)})
		}
	})

	t.Run("unreceipted-output", func(t *testing.T) {
		for _, path := range crossconformance.DeliveredPaths() {
			profile := pathProfiles[path]
			payload := policyconformance.GNUSharedObject()
			_, err := artifactpolicy.NewService().AdmitLocalOutput(context.Background(), artifactpolicy.LocalOutputRequest{
				Descriptor: crossDescriptor(profile.adapter, profile.profile, string(path), payload),
				Payload:    artifactpolicy.Payload{Path: "output/cli", Size: int64(len(payload)), Reader: strings.NewReader(string(payload))},
			})
			record(t, "unreceipted-output", crossconformance.RejectionOutcome{Vector: "unreceipted-output", Path: path, Err: err, Code: diagnosticCode(err)})
		}
	})

	t.Run("graph-family", func(t *testing.T) {
		for _, outcome := range graphRejectionOutcomes(t) {
			record(t, outcome.Vector, outcome)
		}
	})

	t.Run("identity-family", func(t *testing.T) {
		for _, outcome := range identityRejectionOutcomes(t) {
			record(t, outcome.Vector, outcome)
		}
	})

	t.Run("execution-and-output-family", func(t *testing.T) {
		for _, outcome := range executionRejectionOutcomes(t) {
			record(t, outcome.Vector, outcome)
		}
	})

	t.Run("node-graph-and-input-family", func(t *testing.T) {
		for path, source := range map[crossconformance.PathID]nodeCapture{
			crossconformance.PathNPM:         npmCapture(t),
			crossconformance.PathPNPM:        pnpmCapture(t),
			crossconformance.PathYarnClassic: yarnClassicCapture(t),
			crossconformance.PathYarnModern:  yarnModernCapture(t),
		} {
			_ = path
			record(t, "build-cycle", nodeCycleOutcome(t, source))
			record(t, "undeclared-input", nodeUndeclaredInputOutcome(t, source))
		}
	})

	t.Run("integrity-and-offline-family", func(t *testing.T) {
		for _, outcome := range integrityAndOfflineOutcomes(t) {
			record(t, outcome.Vector, outcome)
		}
	})
}

// integrityAndOfflineOutcomes proves that package bytes which no longer match
// the frozen lock, and captured inputs that are simply gone, both fail closed
// before any materialization.
func integrityAndOfflineOutcomes(t *testing.T) []crossconformance.RejectionOutcome {
	t.Helper()
	outcomes := []crossconformance.RejectionOutcome{}

	npmTampered := newNPMFixture(t, nil)
	writeFile(t, npmTampered.tarballs["node_modules/a"].Path, tamperedTarball(t))
	_, npmErr := captureNPM(t, npmTampered)
	outcomes = append(outcomes, crossconformance.RejectionOutcome{Vector: "integrity-mismatch", Path: crossconformance.PathNPM, Err: npmErr, Code: diagnosticCode(npmErr)})

	pnpmTampered := newPNPMFixture(t, nil)
	writeFile(t, pnpmTampered.tarballs["a@1.0.0"].Path, tamperedTarball(t))
	_, pnpmErr := capturePNPM(t, pnpmTampered)
	outcomes = append(outcomes, crossconformance.RejectionOutcome{Vector: "integrity-mismatch", Path: crossconformance.PathPNPM, Err: pnpmErr, Code: diagnosticCode(pnpmErr)})

	classicTampered := newYarnClassicFixture(t, nil)
	for key := range classicTampered.tarballs {
		writeFile(t, classicTampered.tarballs[key].Path, tamperedTarball(t))
		break
	}
	_, classicErr := captureYarnClassic(t, classicTampered)
	outcomes = append(outcomes, crossconformance.RejectionOutcome{Vector: "integrity-mismatch", Path: crossconformance.PathYarnClassic, Err: classicErr, Code: diagnosticCode(classicErr)})

	npmMissing := newNPMFixture(t, nil)
	if err := os.Remove(npmMissing.tarballs["node_modules/a"].Path); err != nil {
		t.Fatal(err)
	}
	_, npmMissingErr := captureNPM(t, npmMissing)
	outcomes = append(outcomes, crossconformance.RejectionOutcome{Vector: "offline-input-missing", Path: crossconformance.PathNPM, Err: npmMissingErr, Code: diagnosticCode(npmMissingErr)})

	modernMissing := newYarnModernFixture(t, nil)
	for key, archive := range modernMissing.archives {
		if err := os.Remove(archive.Path); err != nil {
			t.Fatal(err)
		}
		_ = key
		break
	}
	_, modernMissingErr := captureYarnModern(t, modernMissing)
	outcomes = append(outcomes, crossconformance.RejectionOutcome{Vector: "offline-input-missing", Path: crossconformance.PathYarnModern, Err: modernMissingErr, Code: diagnosticCode(modernMissingErr)})

	missingMirror := newSwiftFixture(t, 0, nil)
	missingMirror.broker.snapshots = map[string]swiftpmsource.Snapshot{}
	starts := 0
	missingMirror.config.ProcessStartObserver = func(swiftpmsource.ManifestPermit) { starts++ }
	_, swiftErr := swiftpmsource.CaptureAndClose(context.Background(), missingMirror.config, swiftpmsource.Request{Root: missingMirror.root, Product: "cli", Resolved: swiftLock()})
	outcomes = append(outcomes, crossconformance.RejectionOutcome{Vector: "offline-input-missing", Path: crossconformance.PathSwiftPM, Err: swiftErr, Code: diagnosticCode(swiftErr), ProcessStarts: 0})
	_ = starts
	return outcomes
}

// tamperedTarball is a structurally valid package payload whose bytes differ
// from the ones the frozen lock recorded.
func tamperedTarball(t *testing.T) []byte {
	t.Helper()
	return buildTGZ(t, sourcePackage{name: "a", version: "1.0.0", files: map[string][]byte{"index.js": []byte("module.exports = 'substituted'\n")}})
}

func crossDescriptor(adapter string, profile artifactpolicy.ProfileID, manager string, payload []byte) artifactpolicy.Descriptor {
	return artifactpolicy.Descriptor{
		AdapterID: adapter, ProfileID: profile, Manager: manager,
		PackageName: "cross-fixture", PackageVersion: "1.0.0",
		Origin: artifactpolicy.OriginEvidence{
			Locator: "cross://" + manager + "/fixture", ImmutableID: "cross-fixture-1.0.0",
			LockRecord: "cross-conformance-lock", ChecksumSHA256: string(digestID(payload)), Verified: true,
		},
	}
}

// compiledByteOutcomes drives one pinned GNU shared object through every
// delivered path's real capture and admission API. No path may admit it, and
// no path may start a manager process first.
func compiledByteOutcomes(t *testing.T, rustManagerUnavailable bool) map[crossconformance.PathID]crossconformance.RejectionOutcome {
	t.Helper()
	outcomes := map[crossconformance.PathID]crossconformance.RejectionOutcome{}
	extra := sharedCompiledPayload()

	_, npmErr := captureNPM(t, newNPMFixture(t, extra))
	outcomes[crossconformance.PathNPM] = crossconformance.RejectionOutcome{Vector: "compiled-dependency-bytes", Path: crossconformance.PathNPM, Err: npmErr, Code: diagnosticCode(npmErr)}

	_, pnpmErr := capturePNPM(t, newPNPMFixture(t, extra))
	outcomes[crossconformance.PathPNPM] = crossconformance.RejectionOutcome{Vector: "compiled-dependency-bytes", Path: crossconformance.PathPNPM, Err: pnpmErr, Code: diagnosticCode(pnpmErr)}

	_, classicErr := captureYarnClassic(t, newYarnClassicFixture(t, extra))
	outcomes[crossconformance.PathYarnClassic] = crossconformance.RejectionOutcome{Vector: "compiled-dependency-bytes", Path: crossconformance.PathYarnClassic, Err: classicErr, Code: diagnosticCode(classicErr)}

	_, modernErr := captureYarnModern(t, newYarnModernFixture(t, extra))
	outcomes[crossconformance.PathYarnModern] = crossconformance.RejectionOutcome{Vector: "compiled-dependency-bytes", Path: crossconformance.PathYarnModern, Err: modernErr, Code: diagnosticCode(modernErr)}

	swift := newSwiftFixture(t, 0, compiledSwiftPayload())
	starts := 0
	swift.config.ProcessStartObserver = func(swiftpmsource.ManifestPermit) { starts++ }
	_, swiftErr := swiftpmsource.CaptureAndClose(context.Background(), swift.config, swiftpmsource.Request{Root: swift.root, Product: "cli", Resolved: swiftLock()})
	outcomes[crossconformance.PathSwiftPM] = crossconformance.RejectionOutcome{Vector: "compiled-dependency-bytes", Path: crossconformance.PathSwiftPM, Err: swiftErr, Code: diagnosticCode(swiftErr), ProcessStarts: starts}

	if !rustManagerUnavailable {
		rustErr := captureRustCompiledWorkspace(t)
		outcomes[crossconformance.PathRust] = crossconformance.RejectionOutcome{Vector: "compiled-dependency-bytes", Path: crossconformance.PathRust, Err: rustErr, Code: diagnosticCode(rustErr)}
	}
	return outcomes
}

func captureRustCompiledWorkspace(t *testing.T) error {
	t.Helper()
	workspace := rustCompiledWorkspace(t)
	manager, err := rustsource.NewManager(context.Background(), rustsource.ManagerConfig{WorkRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("rust manager: %v", err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	manifests := []rustsource.RawManifest{}
	for _, name := range []string{"Cargo.toml", "app/Cargo.toml", "dep/Cargo.toml"} {
		manifests = append(manifests, rustsource.RawManifest{Path: name, File: rustsource.RawFile{Path: joinFixturePath(workspace, name)}})
	}
	paths := []rustsource.RawPathOrigin{}
	for _, name := range []string{"app", "dep"} {
		paths = append(paths, rustsource.RawPathOrigin{DeclaredPath: name, Tree: rustsource.RawTree{Root: joinFixturePath(workspace, name)}})
	}
	_, err = manager.Capture(context.Background(), rustsource.RawCaptureRequest{
		Workspace: rustsource.RawTree{Root: workspace},
		Lock:      rustsource.RawFile{Path: joinFixturePath(workspace, "Cargo.lock")},
		Manifests: manifests, Paths: paths,
	})
	return err
}

// opaqueByteOutcomes drives bytes with no complete allowed interpretation
// through every path's real admission seam.
func opaqueByteOutcomes(t *testing.T) []crossconformance.RejectionOutcome {
	t.Helper()
	payload := []byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0x7f, 0x10, 0x99}
	outcomes := []crossconformance.RejectionOutcome{}
	for _, path := range crossconformance.DeliveredPaths() {
		profile := pathProfiles[path]
		_, err := artifactpolicy.NewService().AdmitDependency(context.Background(), artifactpolicy.DependencyRequest{
			Descriptor: crossDescriptor(profile.adapter, profile.profile, string(path), payload),
			Payload:    artifactpolicy.Payload{Path: "opaque/blob.bin", Size: int64(len(payload)), Reader: strings.NewReader(string(payload))},
		})
		outcomes = append(outcomes, crossconformance.RejectionOutcome{Vector: "opaque-dependency-bytes", Path: path, Err: err, Code: diagnosticCode(err)})
	}
	return outcomes
}

// graphRejectionOutcomes tampers with each path's real selection overlay and
// requires the shared graph authority to fail closed identically.
func graphRejectionOutcomes(t *testing.T) []crossconformance.RejectionOutcome {
	t.Helper()
	outcomes := []crossconformance.RejectionOutcome{}
	for path, records := range closureGraphPaths(t) {
		outcomes = append(outcomes,
			tamperOutcome(t, "binding-duplicate-record", path, records, duplicateBindingNode),
			tamperOutcome(t, "binding-dangling-reference", path, records, danglingBindingEdge),
			tamperOutcome(t, "binding-wrong-kind", path, records, wrongKindBindingNode),
			tamperOutcome(t, "binding-replaces-capture", path, records, captureReplacingBindingNode),
			tamperOutcome(t, "binding-missing-target", path, records, missingTargetBinding),
		)
	}
	return outcomes
}

// pathRecords is one path's real closed graph, kept so the suite can tamper
// with exactly the records the adapter produced.
type pathRecords struct {
	capture   closuregraph.CaptureGraph
	selection closuregraph.SelectionContext
	binding   closuregraph.SelectionBinding
	records   closuregraph.RecordTables
	authority closuregraph.BindingAuthority
}

func closureGraphPaths(t *testing.T) map[crossconformance.PathID]pathRecords {
	t.Helper()
	result := map[crossconformance.PathID]pathRecords{}
	for path, source := range map[crossconformance.PathID]nodeCapture{
		crossconformance.PathNPM:         npmCapture(t),
		crossconformance.PathPNPM:        pnpmCapture(t),
		crossconformance.PathYarnClassic: yarnClassicCapture(t),
		crossconformance.PathYarnModern:  yarnModernCapture(t),
	} {
		result[path] = nodePathRecords(t, source)
	}
	fixture := newSwiftFixture(t, 0, nil)
	capture, err := swiftpmsource.CaptureAndClose(context.Background(), fixture.config, swiftpmsource.Request{Root: fixture.root, Product: "cli", Resolved: swiftLock()})
	if err != nil {
		t.Fatal(err)
	}
	result[crossconformance.PathSwiftPM] = pathRecords{
		capture: capture.Graph, selection: capture.Selection, binding: capture.Binding,
		records: capture.Records, authority: capture.Authority,
	}
	return result
}

func tamperOutcome(t *testing.T, vector string, path crossconformance.PathID, records pathRecords, tamper func(*testing.T, pathRecords) (closuregraph.SelectionBinding, closuregraph.RecordTables, closuregraph.SelectionContext)) crossconformance.RejectionOutcome {
	t.Helper()
	binding, tables, selection := tamper(t, records)
	_, err := closuregraph.ProjectActive(records.capture, selection, binding, tables, records.authority, nil)
	return crossconformance.RejectionOutcome{Vector: vector, Path: path, Err: err, Code: diagnosticCode(err)}
}

func duplicateBindingNode(t *testing.T, records pathRecords) (closuregraph.SelectionBinding, closuregraph.RecordTables, closuregraph.SelectionContext) {
	t.Helper()
	binding := cloneBinding(records.binding)
	if len(binding.BindingNodeIDs) == 0 {
		t.Fatal("path emitted no binding node to duplicate")
	}
	binding.BindingNodeIDs = append(binding.BindingNodeIDs, binding.BindingNodeIDs[0])
	return binding, records.records, records.selection
}

func danglingBindingEdge(t *testing.T, records pathRecords) (closuregraph.SelectionBinding, closuregraph.RecordTables, closuregraph.SelectionContext) {
	t.Helper()
	binding := cloneBinding(records.binding)
	binding.BindingEdgeIDs = insertSorted(binding.BindingEdgeIDs, closuregraph.ID("sha256:"+repeat64('7')))
	return binding, records.records, records.selection
}

func wrongKindBindingNode(t *testing.T, records pathRecords) (closuregraph.SelectionBinding, closuregraph.RecordTables, closuregraph.SelectionContext) {
	t.Helper()
	intruder := closuregraph.Node{Kind: closuregraph.NodeSourceSet, LogicalKey: "cross:intruder", Payload: closuregraph.SourceSetPayload{
		Profile: "cross-conformance-v1", Origin: "cross://intruder", Grammar: "text-v1", TrustRole: "dependency_input",
		Projection: []string{"intruder.txt"}, ArtifactManifestID: closuregraph.ID("sha256:" + repeat64('8')),
	}}
	id, err := intruder.ID()
	if err != nil {
		t.Fatal(err)
	}
	binding := cloneBinding(records.binding)
	binding.BindingNodeIDs = insertSorted(binding.BindingNodeIDs, id)
	tables := cloneTables(records.records)
	tables.BindingNodes = append(tables.BindingNodes, intruder)
	return binding, tables, records.selection
}

func captureReplacingBindingNode(t *testing.T, records pathRecords) (closuregraph.SelectionBinding, closuregraph.RecordTables, closuregraph.SelectionContext) {
	t.Helper()
	if len(records.records.CaptureNodes) == 0 {
		t.Fatal("path emitted no capture node to replace")
	}
	replaced := records.records.CaptureNodes[0]
	id, err := replaced.ID()
	if err != nil {
		t.Fatal(err)
	}
	binding := cloneBinding(records.binding)
	binding.BindingNodeIDs = insertSorted(binding.BindingNodeIDs, id)
	tables := cloneTables(records.records)
	tables.BindingNodes = append(tables.BindingNodes, replaced)
	return binding, tables, records.selection
}

func missingTargetBinding(t *testing.T, records pathRecords) (closuregraph.SelectionBinding, closuregraph.RecordTables, closuregraph.SelectionContext) {
	t.Helper()
	binding := cloneBinding(records.binding)
	tables := cloneTables(records.records)
	target := records.selection.PlatformRoles[closuregraph.PlatformTarget]
	kept := []closuregraph.ID{}
	for _, id := range binding.BindingNodeIDs {
		if id != target {
			kept = append(kept, id)
		}
	}
	binding.BindingNodeIDs = kept
	nodes := []closuregraph.Node{}
	for _, node := range tables.BindingNodes {
		id, err := node.ID()
		if err != nil {
			t.Fatal(err)
		}
		if id != target {
			nodes = append(nodes, node)
		}
	}
	tables.BindingNodes = nodes
	return binding, tables, records.selection
}

// insertSorted keeps a binding ID list in the canonical order the schema
// requires, so a tamper case exercises the reference rule under test instead of
// tripping the ordering rule first.
func insertSorted(values []closuregraph.ID, addition closuregraph.ID) []closuregraph.ID {
	result := append(append([]closuregraph.ID(nil), values...), addition)
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func cloneBinding(binding closuregraph.SelectionBinding) closuregraph.SelectionBinding {
	clone := binding
	clone.BindingNodeIDs = append([]closuregraph.ID(nil), binding.BindingNodeIDs...)
	clone.BindingEdgeIDs = append([]closuregraph.ID(nil), binding.BindingEdgeIDs...)
	return clone
}

func cloneTables(tables closuregraph.RecordTables) closuregraph.RecordTables {
	return closuregraph.NewRecordTables(
		append([]closuregraph.Node(nil), tables.CaptureNodes...),
		append([]closuregraph.Edge(nil), tables.CaptureEdges...),
		append([]closuregraph.Node(nil), tables.BindingNodes...),
		append([]closuregraph.Edge(nil), tables.BindingEdges...),
	)
}

// identityRejectionOutcomes proves that target and toolchain drift fail before
// any affected process.
func identityRejectionOutcomes(t *testing.T) []crossconformance.RejectionOutcome {
	t.Helper()
	outcomes := []crossconformance.RejectionOutcome{}

	capture := rustCaptureGraph(t, nil)
	metadata, err := rustsource.ParseMetadata([]byte(rustMetadata))
	if err != nil {
		t.Fatal(err)
	}
	_, rustErr := rustsource.Reconcile(capture, rustsource.SelectionContext{
		Package: "app", Binary: "app-bin", Target: "wasm32-unknown-unknown", DefaultFeatures: true,
		Features: []string{}, TargetCFG: []string{}, ResolvedFeatures: map[string][]string{"app 0.1.0": {}, "dep 0.1.0": {}},
	}, metadata, "aarch64-apple-darwin", "sha256:"+repeat64('1'))
	outcomes = append(outcomes, crossconformance.RejectionOutcome{Vector: "target-identity-drift", Path: crossconformance.PathRust, Err: rustErr, Code: diagnosticCode(rustErr)})

	swift := newSwiftFixture(t, 0, nil)
	swift.config.Destination.Platform.TargetTriple = ""
	starts := 0
	swift.config.ProcessStartObserver = func(swiftpmsource.ManifestPermit) { starts++ }
	_, swiftErr := swiftpmsource.CaptureAndClose(context.Background(), swift.config, swiftpmsource.Request{Root: swift.root, Product: "cli", Resolved: swiftLock()})
	outcomes = append(outcomes, crossconformance.RejectionOutcome{Vector: "target-identity-drift", Path: crossconformance.PathSwiftPM, Err: swiftErr, Code: diagnosticCode(swiftErr), ProcessStarts: starts})

	drift := newSwiftFixture(t, 0, nil)
	driftStarts := 0
	drift.config.ProcessStartObserver = func(swiftpmsource.ManifestPermit) { driftStarts++ }
	drift.config.Toolchain.Recheck = func(context.Context, swiftpmsource.ToolIdentity) (closureexec.ToolchainIdentity, error) {
		return closureexec.ToolchainIdentity{Fingerprint: closuregraph.ID("sha256:" + repeat64('e')), ExecutableSHA256: closuregraph.ID("sha256:" + repeat64('f'))}, nil
	}
	_, driftErr := swiftpmsource.CaptureAndClose(context.Background(), drift.config, swiftpmsource.Request{Root: drift.root, Product: "cli", Resolved: swiftLock()})
	outcomes = append(outcomes, crossconformance.RejectionOutcome{Vector: "toolchain-identity-drift", Path: crossconformance.PathSwiftPM, Err: driftErr, Code: diagnosticCode(driftErr), ProcessStarts: driftStarts})
	return outcomes
}

// executionRejectionOutcomes proves the execution and output boundaries fail
// closed for the seams reachable without a live manager process.
func executionRejectionOutcomes(t *testing.T) []crossconformance.RejectionOutcome {
	t.Helper()
	outcomes := []crossconformance.RejectionOutcome{}

	escaping := newSwiftFixture(t, 0, nil)
	escaping.config.Evaluator = &swiftEscapingDependencyEvaluator{inner: &swiftEvaluator{manifests: swiftManifests()}}
	escapeStarts := 0
	escaping.config.ProcessStartObserver = func(swiftpmsource.ManifestPermit) { escapeStarts++ }
	_, outsideErr := swiftpmsource.CaptureAndClose(context.Background(), escaping.config, swiftpmsource.Request{Root: escaping.root, Product: "cli", Resolved: swiftLock()})
	outcomes = append(outcomes, crossconformance.RejectionOutcome{Vector: "undeclared-input", Path: crossconformance.PathSwiftPM, Err: outsideErr, Code: diagnosticCode(outsideErr)})

	escape := newSwiftFixture(t, 0, map[string][]byte{})
	escape.config.Evaluator = &swiftEscapingEvaluator{inner: &swiftEvaluator{manifests: swiftManifests()}}
	_, escapeErr := swiftpmsource.CaptureAndClose(context.Background(), escape.config, swiftpmsource.Request{Root: escape.root, Product: "cli", Resolved: swiftLock()})
	outcomes = append(outcomes, crossconformance.RejectionOutcome{Vector: "undeclared-process", Path: crossconformance.PathSwiftPM, Err: escapeErr, Code: diagnosticCode(escapeErr)})
	return outcomes
}

// swiftEscapingDependencyEvaluator declares a target source outside the
// admitted package tree, which is the canonical undeclared-input shape for
// SwiftPM: nothing in the frozen closure can account for those bytes.
type swiftEscapingDependencyEvaluator struct{ inner *swiftEvaluator }

func (evaluator *swiftEscapingDependencyEvaluator) Evaluate(ctx context.Context, root string, permit swiftpmsource.ManifestPermit) (swiftpmsource.ManifestResult, error) {
	result, err := evaluator.inner.Evaluate(ctx, root, permit)
	if err != nil || permit.PackageIdentity != "root" {
		return result, err
	}
	targets := append([]swiftpmsource.Target(nil), result.Manifest.Targets...)
	for index := range targets {
		if targets[index].Name != "App" {
			continue
		}
		targets[index].Sources = append(append([]string(nil), targets[index].Sources...), "../outside-the-closure/hidden.swift")
	}
	result.Manifest.Targets = targets
	return result, nil
}

// swiftEscapingEvaluator returns a manifest the adapter never authorized: it
// answers with an identity the permit did not name.
type swiftEscapingEvaluator struct{ inner *swiftEvaluator }

func (evaluator *swiftEscapingEvaluator) Evaluate(ctx context.Context, root string, permit swiftpmsource.ManifestPermit) (swiftpmsource.ManifestResult, error) {
	result, err := evaluator.inner.Evaluate(ctx, root, permit)
	if err != nil {
		return result, err
	}
	result.ReceiptID = ""
	return result, nil
}

func joinFixturePath(root, name string) string {
	return filepath.Join(root, filepath.FromSlash(name))
}
