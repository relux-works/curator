package swiftpmbuild

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpminterop"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// S01/S02: the accepted Swift-plus-C closure plans one object per source of
// every selected target and exactly one published product.
func TestS01S02SelectedTargetsDeclareOneObjectPerSource(t *testing.T) {
	plan := newFixture(t).mustPlan()
	if len(plan.Objects) != 2 {
		t.Fatalf("declared object slots = %#v", plan.Objects)
	}
	targets := map[string]string{}
	for _, slot := range plan.Objects {
		targets[slot.Target] = slot.Source
		if !strings.HasSuffix(slot.Path, slot.Source+".o") {
			t.Fatalf("object slot path %q does not name its source %q", slot.Path, slot.Source)
		}
	}
	if targets["App"] != "Sources/App/main.swift" || targets["CLib"] != "Sources/CLib/lib.c" {
		t.Fatalf("object slots = %#v", targets)
	}
}

// S03/S10: a multi-source C-family target declares and publishes one object
// per source; the accepted per-target declaration is never collapsed.
func TestS03MultiSourceTargetDeclaresEveryObject(t *testing.T) {
	fixture := newFixture(t)
	fixture.files["Sources/CLib/extra.c"] = "#include \"CLib.h\"\nint extra(void) { return 2; }\n"
	fixture.manifest.Targets[1].Sources = []string{"Sources/CLib/extra.c", "Sources/CLib/lib.c"}
	fixture.stubExtra = []stubAction{{Op: "write", Path: ".curator/scratch/" + fixtureScratchTriple + "/" + fixtureConfiguration + "/CLib.build/extra.c.o", Payload: "extra-object"}}
	plan := fixture.mustPlan()
	if len(plan.Objects) != 3 {
		t.Fatalf("declared object slots = %d, want one per source", len(plan.Objects))
	}
	result, err := fixture.manager().Build(t.Context(), plan)
	if err != nil {
		t.Fatalf("multi-source build failed: %v", err)
	}
	if len(result.Observations) != 4 || len(result.Execution.WriteSet) != 4 {
		t.Fatalf("observations=%d write set=%v", len(result.Observations), result.Execution.WriteSet)
	}
}

// S03/S10: one Clang target may carry two sources that share a base name in
// different subdirectories. SwiftPM mirrors the tree below the target's
// declared source root, so both objects must resolve and publish; matching on
// the base name alone cannot disambiguate them.
func TestS03SameBaseNameSourcesInDifferentDirectoriesResolve(t *testing.T) {
	fixture := newFixture(t)
	fixture.files["Sources/CLib/a/x.c"] = "#include \"CLib.h\"\nint left(void) { return 2; }\n"
	fixture.files["Sources/CLib/b/x.c"] = "#include \"CLib.h\"\nint right(void) { return 3; }\n"
	fixture.manifest.Targets[1].Sources = []string{"Sources/CLib/a/x.c", "Sources/CLib/b/x.c", "Sources/CLib/lib.c"}
	scratch := ".curator/scratch/" + fixtureScratchTriple + "/" + fixtureConfiguration
	fixture.stubExtra = []stubAction{
		{Op: "mkdir", Path: scratch + "/CLib.build/a"},
		{Op: "mkdir", Path: scratch + "/CLib.build/b"},
		{Op: "write", Path: scratch + "/CLib.build/a/x.c.o", Payload: "left-object"},
		{Op: "write", Path: scratch + "/CLib.build/b/x.c.o", Payload: "right-object"},
	}
	plan := fixture.mustPlan()
	if len(plan.Objects) != 4 {
		t.Fatalf("declared object slots = %d, want one per source", len(plan.Objects))
	}
	result, err := fixture.manager().Build(t.Context(), plan)
	if err != nil {
		t.Fatalf("same-base-name build failed: %v", err)
	}
	published := map[string]closuregraph.ID{}
	for _, observation := range result.Observations {
		published[observation.Path] = observation.SHA256
	}
	left, right := published[".curator/objects/root/CLib/Sources/CLib/a/x.c.o"], published[".curator/objects/root/CLib/Sources/CLib/b/x.c.o"]
	if !left.Valid() || !right.Valid() || left == right {
		t.Fatalf("same-base-name objects did not publish distinct bytes: %#v", published)
	}
	if len(result.Observations) != 5 || len(result.Execution.WriteSet) != 5 {
		t.Fatalf("observations=%d write set=%v", len(result.Observations), result.Execution.WriteSet)
	}
}

// A pruned target is captured but never compiled, so it contributes neither an
// action nor a declared object to the exact build plan.
func TestPrunedTargetContributesNoDeclaredOutput(t *testing.T) {
	fixture := newFixture(t)
	fixture.manifest.Targets[0].Dependencies = []swiftpmsource.TargetDependency{{Name: "CLib", Condition: &closuregraph.Condition{EvaluatorID: swiftpmsource.ConditionEvaluatorID, Expression: "platform=linux"}}}
	plan := fixture.mustPlan()
	for _, slot := range plan.Objects {
		if slot.Target == "CLib" {
			t.Fatalf("pruned target contributed a declared output: %#v", slot)
		}
	}
	if len(plan.Objects) != 1 {
		t.Fatalf("declared object slots = %#v", plan.Objects)
	}
}

// R01/R05/R07: a root that reaches a pinned source-control dependency mounts
// the admitted mirror read-only next to the admitted build root, denies
// network at the committed permit boundary, and generates a kind-preserving
// mirrors.json that maps the original origin onto that exact mount.
func TestR01R05OfflineBuildMountsAdmittedInputsWithNetworkDenied(t *testing.T) {
	fixture := newFixture(t)
	fixture.addSourceControlDependency("dep", "CDep", map[string]string{
		"Sources/CDep/dep.c":          "#include \"CDep.h\"\nint dep(void) { return 7; }\n",
		"Sources/CDep/include/CDep.h": "#ifndef CDEP_H\n#define CDEP_H\nint dep(void);\n#endif\n",
	})
	scratch := ".curator/scratch/" + fixtureScratchTriple + "/" + fixtureConfiguration
	fixture.stubExtra = []stubAction{
		{Op: "mkdir", Path: scratch + "/CDep.build"},
		{Op: "write", Path: scratch + "/CDep.build/dep.c.o", Payload: "dep-object"},
		{Op: "write", Path: scratch + "/CDep.build/dep.c.d",
			Payload: "{{PWD}}/" + scratch + "/CDep.build/dep.c.o : {{PWD}}/.curator/scratch/checkouts/dep/Sources/CDep/dep.c\n"},
	}
	fixture.materialize()
	capture, interop := fixture.closure()
	plan, err := NewPlan(t.Context(), fixture.build, capture, interop)
	if err != nil {
		t.Fatal(err)
	}
	mount := onlyMirrorMount(t, capture)
	result, err := fixture.manager().Build(t.Context(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if !equalStrings(result.ReadSet, []string{buildRootMount, mount}) {
		t.Fatalf("read set = %#v, want the admitted build root plus %q", result.ReadSet, mount)
	}
	if result.Execution.Network != "none" {
		t.Fatalf("execution network = %q", result.Execution.Network)
	}
	dependencyObject := false
	for _, observation := range result.Observations {
		if observation.Path == ".curator/objects/dep/CDep/Sources/CDep/dep.c.o" {
			dependencyObject = true
		}
	}
	if !dependencyObject {
		t.Fatalf("dependency object was not published: %#v", result.Observations)
	}
}

// The generated mirror configuration preserves the SwiftPM source-control kind
// of every admitted pin and points it at the exact isolated mount.
func TestGeneratedMirrorConfigurationPreservesSourceControlKind(t *testing.T) {
	fixture := newFixture(t)
	fixture.addSourceControlDependency("dep", "CDep", map[string]string{
		"Sources/CDep/dep.c":          "#include \"CDep.h\"\nint dep(void) { return 7; }\n",
		"Sources/CDep/include/CDep.h": "#ifndef CDEP_H\n#define CDEP_H\nint dep(void);\n#endif\n",
	})
	fixture.materialize()
	capture, _ := fixture.closure()
	root, cleanup, err := materializeCaptureRoot(fixture.build, capture)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	payload, err := os.ReadFile(filepath.Join(root, ".curator", "config", "mirrors.json"))
	if err != nil {
		t.Fatal(err)
	}
	want := "file://" + filepath.Join(fixture.execRoot, filepath.FromSlash(onlyMirrorMount(t, capture)))
	var generated struct {
		Object []struct{ Original, Mirror string } `json:"object"`
	}
	if err = json.Unmarshal(payload, &generated); err != nil {
		t.Fatal(err)
	}
	if len(generated.Object) != 1 || generated.Object[0].Original != "https://example.invalid/dep" || generated.Object[0].Mirror != want {
		t.Fatalf("generated mirrors = %#v, want the remote origin mapped onto %q", generated.Object, want)
	}
}

// Two admitted mirrors that share one intake receipt cannot both be mounted,
// so the offline build rejects the duplicate before any process starts.
func TestDuplicateMirrorReceiptFailsClosed(t *testing.T) {
	fixture := newFixture(t)
	fixture.addSourceControlDependency("dep", "CDep", map[string]string{
		"Sources/CDep/dep.c":          "#include \"CDep.h\"\nint dep(void) { return 7; }\n",
		"Sources/CDep/include/CDep.h": "#ifndef CDEP_H\n#define CDEP_H\nint dep(void);\n#endif\n",
	})
	fixture.materialize()
	capture, interop := fixture.closure()
	plan, err := NewPlan(t.Context(), fixture.build, capture, interop)
	if err != nil {
		t.Fatal(err)
	}
	capture.Mirrors = append(capture.Mirrors, capture.Mirrors[0])
	_, err = fixture.manager().Build(t.Context(), plan)
	requireCode(t, err, CodeMirrorMissing)
	if fixture.starts != 0 {
		t.Fatalf("duplicate mirror receipt started %d processes", fixture.starts)
	}
	requireNoPublication(t, fixture)
}

// R06: a root lock pin without a captured mirror fails closed before planning.
func TestR06MissingMirrorFailsClosed(t *testing.T) {
	fixture := newFixture(t)
	fixture.materialize()
	capture, interop := fixture.closure()
	capture.Lock.Pins = append(capture.Lock.Pins, swiftpmsource.Pin{Identity: "dependency", Kind: swiftpmsource.SourceRemote, Revision: strings.Repeat("a", 40)})
	_, err := NewPlan(t.Context(), fixture.build, capture, interop)
	requireCode(t, err, CodeMirrorMissing)
}

// A build requested before the root lock is frozen is rejected.
func TestUnfrozenResolutionFailsClosed(t *testing.T) {
	fixture := newFixture(t)
	fixture.materialize()
	capture, interop := fixture.closure()
	capture.Lock.Digest = ""
	_, err := NewPlan(t.Context(), fixture.build, capture, interop)
	requireCode(t, err, CodeResolutionUnfrozen)
}

// Graph drift between the accepted closure and the republished build closure
// is rejected before any process starts.
func TestGraphDriftIsRejectedBeforeExecution(t *testing.T) {
	fixture := newFixture(t)
	fixture.materialize()
	capture, interop := fixture.closure()
	dropped := interop.Records.CaptureNodes[:len(interop.Records.CaptureNodes)-1]
	interop.Records.CaptureNodes = append([]closuregraph.Node(nil), dropped...)
	_, err := NewPlan(t.Context(), fixture.build, capture, interop)
	if err == nil {
		t.Fatal("a drifted accepted closure was accepted")
	}
	requireNoPublication(t, fixture)
}

// An object a selected target produced that no declared slot claims is
// undeclared local generation and fails closed without publication. The
// declared slot itself still resolves, so this fails for exactly that reason
// rather than because the resolution was ambiguous.
func TestUndeclaredGeneratedObjectFailsClosed(t *testing.T) {
	scratch := ".curator/scratch/" + fixtureScratchTriple + "/" + fixtureConfiguration
	fixture := newFixture(t)
	fixture.stubExtra = []stubAction{
		{Op: "mkdir", Path: scratch + "/CLib.build/generated"},
		{Op: "write", Path: scratch + "/CLib.build/generated/smuggled.c.o", Payload: "undeclared"},
	}
	plan := fixture.mustPlan()
	slot := ObjectSlot{}
	for _, declared := range plan.Objects {
		if declared.Target == "CLib" {
			slot = declared
		}
	}
	_, err := fixture.manager().Build(t.Context(), plan)
	requireCode(t, err, CodeOutputUnreceipted)
	requireNoPublication(t, fixture)
	if match, resolveErr := resolveProducedObject(slot, []string{"lib.c.o", "generated/smuggled.c.o"}); resolveErr != nil || match != "lib.c.o" {
		t.Fatalf("declared slot resolution = %q (%v); the rejection must come from the unclaimed object", match, resolveErr)
	}
}

// The command reconciliation rejects a receipt that inflates portable
// assurance, drops the exact declared product, or claims a network attempt.
func TestCommandReconciliationRejectsInflatedOrDriftedEvidence(t *testing.T) {
	permit := closureexec.DerivationPermit{
		Executable: stubExecutableRelative(), CWD: "work/package", Argv: []string{"build"},
		ReadRoots: []string{"inputs/build-root"}, WriteRoots: []string{"work/package"},
		AllowedProcesses: []string{stubExecutableRelative()},
		ExpectedEvidence: []closureexec.EvidenceRequirement{{Path: "work/package/product", SchemaID: ExecutablePayloadSchemaID}},
	}
	base := closureexec.DerivationReceipt{
		AssuranceMode: closureexec.AssurancePortable, BeforeFingerprint: id('1'), AfterFingerprint: id('1'),
		Audit:   closureexec.Audit{Executable: stubExecutableRelative(), CWD: "work/package", Argv: []string{"build"}, Network: "not-observed"},
		Outputs: []closureexec.DerivationOutput{{Path: "work/package/product", SchemaID: ExecutablePayloadSchemaID}},
	}
	if err := reconcileCommand(permit, base); err != nil {
		t.Fatalf("honest portable receipt rejected: %v", err)
	}
	t.Run("inflated portable assurance", func(t *testing.T) {
		receipt := base
		receipt.Audit.Reads = []string{"inputs/build-root"}
		requireCode(t, reconcileCommand(permit, receipt), CodeOfflineRebuildFailed)
	})
	t.Run("drifted command", func(t *testing.T) {
		receipt := base
		receipt.Audit.Argv = []string{"build", "--extra"}
		requireCode(t, reconcileCommand(permit, receipt), CodeBuildGraphDrift)
	})
	t.Run("toolchain drift", func(t *testing.T) {
		receipt := base
		receipt.AfterFingerprint = id('2')
		requireCode(t, reconcileCommand(permit, receipt), CodeToolchainChanged)
	})
	t.Run("missing declared product", func(t *testing.T) {
		receipt := base
		receipt.Outputs = nil
		requireCode(t, reconcileCommand(permit, receipt), CodeOutputUnreceipted)
	})
	t.Run("verified network attempt", func(t *testing.T) {
		receipt := base
		receipt.AssuranceMode = closureexec.AssuranceVerified
		receipt.Audit.Network = "observed"
		requireCode(t, reconcileCommand(permit, receipt), CodeNetworkAttempted)
	})
	t.Run("verified undeclared read", func(t *testing.T) {
		receipt := base
		receipt.AssuranceMode = closureexec.AssuranceVerified
		receipt.Audit.Network = "none"
		receipt.Audit.Reads = []string{"inputs/build-root", "/etc"}
		requireCode(t, reconcileCommand(permit, receipt), CodeInputUndeclared)
	})
	t.Run("verified undeclared write", func(t *testing.T) {
		receipt := base
		receipt.AssuranceMode = closureexec.AssuranceVerified
		receipt.Audit.Network = "none"
		receipt.Audit.Reads = permit.ReadRoots
		receipt.Audit.Writes = []string{"/tmp"}
		requireCode(t, reconcileCommand(permit, receipt), CodeWriteUndeclared)
	})
	t.Run("verified undeclared process", func(t *testing.T) {
		receipt := base
		receipt.AssuranceMode = closureexec.AssuranceVerified
		receipt.Audit.Network = "none"
		receipt.Audit.Reads = permit.ReadRoots
		receipt.Audit.Writes = permit.WriteRoots
		receipt.Audit.Processes = []string{"bin/curl"}
		requireCode(t, reconcileCommand(permit, receipt), CodeProcessUndeclared)
	})
	t.Run("unsupported assurance mode", func(t *testing.T) {
		receipt := base
		receipt.AssuranceMode = "best-effort"
		requireCode(t, reconcileCommand(permit, receipt), CodeOfflineRebuildFailed)
	})
}

// Output drift between the observed bytes and the issued enforcement receipt
// is rejected before publication, and an absent declared product is
// unreceipted local output rather than a silent success.
func TestOutputDriftIsRejectedBeforePublication(t *testing.T) {
	fixture := newFixture(t)
	plan := fixture.mustPlan()
	if _, err := fixture.manager().Build(t.Context(), plan); err != nil {
		t.Fatal(err)
	}
	evidence := filepath.Join(buildWorkMount, plan.OutputPath)
	receipt := closureexec.DerivationReceipt{Outputs: []closureexec.DerivationOutput{{Path: evidence, SHA256: id('c'), Size: 3}}}
	_, err := observeProduct(fixture.build, plan, receipt, filepath.ToSlash(evidence), map[string][]byte{})
	requireCode(t, err, CodeOutputDrift)
	if _, err = observeProduct(fixture.build, plan, receipt, "work/package/absent", map[string][]byte{}); ErrorCode(err) != CodeOutputUnreceipted {
		t.Fatalf("absent product error = %v", err)
	}
}

// The build stage never mutates an accepted expected graph record while it
// publishes observed bytes.
func TestPublicationDoesNotMutateExpectedGraphRecords(t *testing.T) {
	fixture := newFixture(t)
	plan := fixture.mustPlan()
	before, err := plan.Graph.Capture.ID()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = fixture.manager().Build(t.Context(), plan); err != nil {
		t.Fatal(err)
	}
	after, err := plan.Graph.Capture.ID()
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("publication mutated the expected capture graph identity")
	}
	planID, _ := plan.BuildPlan.ID()
	c5 := plan.C5.Payload.(closuregraph.C5PlanPayload)
	if c5.BuildPlanID != planID {
		t.Fatal("C5 no longer names the exact build plan")
	}
}

var _ = swiftpminterop.ProfileID

// onlyMirrorMount returns the isolated mount of the single admitted mirror.
func onlyMirrorMount(t *testing.T, capture *swiftpmsource.Capture) string {
	t.Helper()
	mirrors, err := capture.OfflineMirrors()
	if err != nil {
		t.Fatal(err)
	}
	if len(mirrors) != 1 {
		t.Fatalf("admitted mirrors = %d, want exactly one", len(mirrors))
	}
	return mirrors[0].Mount
}
