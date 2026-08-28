package swiftpmbuild

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpminterop"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// The manager refuses an incomplete authority, non-absolute roots, and an
// overlapping protected store before any planning work happens.
func TestManagerAuthorityContract(t *testing.T) {
	if _, err := NewManager(Config{}); ErrorCode(err) != CodeDerivationUnauthorized {
		t.Fatalf("empty manager config = %v", err)
	}
	fixture := newFixture(t)
	fixture.materialize()
	if _, err := NewManager(fixture.build); err != nil {
		t.Fatalf("complete manager config = %v", err)
	}
	t.Run("relative roots", func(t *testing.T) {
		config := fixture.build
		config.StoreRoot = "protected"
		requireCode(t, mustFailManager(t, config), CodeDerivationUnauthorized)
	})
	t.Run("store inside execution root", func(t *testing.T) {
		config := fixture.build
		config.StoreRoot = filepath.Join(fixture.execRoot, "protected")
		requireCode(t, mustFailManager(t, config), CodeDerivationUnauthorized)
	})
	t.Run("output root outside execution root", func(t *testing.T) {
		config := fixture.build
		config.OutputRoot = filepath.Join(fixture.base, "elsewhere")
		requireCode(t, mustFailManager(t, config), CodeDerivationUnauthorized)
	})
	var absent *Manager
	if _, err := absent.Build(context.Background(), nil); ErrorCode(err) != CodeDerivationUnauthorized {
		t.Fatalf("absent manager = %v", err)
	}
}

func mustFailManager(t *testing.T, config Config) error {
	t.Helper()
	_, err := NewManager(config)
	if err == nil {
		t.Fatal("invalid manager configuration was accepted")
	}
	return err
}

// CGP05: the exact binding names the platform and every selected build tool
// identity, and the selection-neutral capture never carries one of them.
func TestCGP05BuildBindingNamesEveryExactSelectedIdentity(t *testing.T) {
	plan := newFixture(t).mustPlan()
	for _, slot := range []ToolSlot{SlotSwiftPM, SlotSwiftCompiler, SlotPackageDescription, SlotClang, SlotLinker, SlotSDK} {
		bound, present := plan.Binding.Slots[slot]
		if !present || !bound.NodeID.Valid() || bound.Payload.ContentFingerprint == "" {
			t.Fatalf("slot %q resolved to %#v", slot, bound)
		}
	}
	if !plan.Binding.PlatformNodeID.Valid() || plan.Binding.Platform.TargetTriple != fixtureTriple {
		t.Fatalf("platform binding = %#v", plan.Binding)
	}
	for _, node := range plan.Graph.Records.CaptureNodes {
		if node.Kind == closuregraph.NodeTargetPlatform || node.Kind == closuregraph.NodeToolchainComponent {
			t.Fatalf("capture graph contains an exact selection fact: %s", node.LogicalKey)
		}
	}
	found := false
	for _, node := range plan.Graph.Records.BindingNodes {
		if node.LogicalKey == "swiftpm.build.component.macos-linker" {
			found = true
		}
	}
	if !found {
		t.Fatal("selected linker is absent from the exact selection binding")
	}
}

// CGP05: every selected action binds exactly one target platform and resolves
// each declared tool slot exactly once.
func TestCGP05EveryActionTargetAndToolSlotResolvesOnce(t *testing.T) {
	plan := newFixture(t).mustPlan()
	if err := validateActionBindings(plan.Graph, plan.Binding); err != nil {
		t.Fatalf("single-resolution rule failed: %v", err)
	}
	slots := map[closuregraph.ID]map[string]int{}
	for _, edge := range plan.Graph.Records.BindingEdges {
		if edge.Kind != closuregraph.EdgeUsesTool {
			continue
		}
		if slots[edge.FromNodeID] == nil {
			slots[edge.FromNodeID] = map[string]int{}
		}
		slots[edge.FromNodeID][edge.Payload.(closuregraph.UsesToolPayload).ToolSlot]++
	}
	link := slots[plan.LinkActionNodeID]
	if link["build-driver"] != 1 || link["linker"] != 1 || len(link) != 2 {
		t.Fatalf("link action tool slots = %#v", link)
	}
}

// A missing, duplicate, dangling, wrong-kind, or unused build slot declaration
// fails closed before any process starts.
func TestBuildBindingRejectsMissingDuplicateAndWrongKindSlots(t *testing.T) {
	t.Run("missing slot role", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.materializeHook = func() { delete(fixture.build.Slots, SlotClang) }
		_, err := fixture.plan()
		requireCode(t, err, CodeGraphReferenceInvalid)
	})
	t.Run("unknown component role", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.materializeHook = func() { fixture.build.Slots[SlotClang] = "clang-absent" }
		_, err := fixture.plan()
		requireCode(t, err, CodeGraphReferenceInvalid)
	})
	t.Run("two slots resolve to one component", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.materializeHook = func() { fixture.build.Slots[SlotClang] = "swift" }
		_, err := fixture.plan()
		requireCode(t, err, CodeGraphReferenceInvalid)
	})
	t.Run("linker role duplicates an accepted binding", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.materializeHook = func() { fixture.build.Linker.Role = "clang" }
		_, err := fixture.plan()
		requireCode(t, err, CodeGraphReferenceInvalid)
	})
	t.Run("slot the selection does not use", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.materializeHook = func() { fixture.build.Slots[SlotClangCXX] = "clang++" }
		_, err := fixture.plan()
		requireCode(t, err, CodeGraphReferenceInvalid)
	})
	t.Run("incomplete linker identity", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.materializeHook = func() { fixture.build.Linker.VersionOutput = "" }
		_, err := fixture.plan()
		requireCode(t, err, CodeToolchainUntrusted)
	})
}

// A physical component whose identity drifted between binding and use is
// rejected at the immediate time-of-use boundary.
func TestDriftedToolBindingIsRejectedBeforePlanning(t *testing.T) {
	fixture := newFixture(t)
	fixture.materializeHook = func() { fixture.build.Recheck = driftedRecheck("clang") }
	_, err := fixture.plan()
	requireCode(t, err, CodeToolchainChanged)
}

// A read-set observation failure is a rejection, never a silent downgrade.
func TestRecheckFailureIsRejected(t *testing.T) {
	fixture := newFixture(t)
	fixture.materializeHook = func() { fixture.build.Recheck = failingRecheck }
	_, err := fixture.plan()
	requireCode(t, err, CodeToolchainChanged)
}

// Verified assurance without an observed compiler read set fails closed rather
// than silently accepting the portable reject-by-default verdict.
func TestVerifiedBuildRequiresObservedReadSet(t *testing.T) {
	fixture := newFixture(t)
	fixture.materializeHook = func() { fixture.build.Assurance = closureexec.AssuranceVerified }
	_, err := fixture.plan()
	requireCode(t, err, CodeHeaderInputUndeclared)
}

// The plan preserves every accepted capture and binding identity and only adds
// the product link action and its expected output.
func TestPlanPreservesAcceptedIdentitiesThroughC5(t *testing.T) {
	fixture := newFixture(t)
	fixture.materialize()
	capture, interop := fixture.closure()
	plan, err := NewPlan(t.Context(), fixture.build, capture, interop)
	if err != nil {
		t.Fatal(err)
	}
	if err = preservesAcceptedIdentities(interop, plan.Graph); err != nil {
		t.Fatalf("accepted identities were not preserved: %v", err)
	}
	interopCapture, _ := interop.Graph.ID()
	planCapture, _ := plan.Graph.Capture.ID()
	if interopCapture == planCapture {
		t.Fatal("build planning did not extend the accepted capture graph")
	}
	added := len(plan.Graph.Records.CaptureNodes) - len(interop.Records.CaptureNodes)
	if added != 2 {
		t.Fatalf("build planning added %d capture nodes, want the link action and its output", added)
	}
	if plan.C4.PreviousCheckpointID == nil {
		t.Fatal("build C4 does not chain from the accepted interop C4")
	}
	interopC4, _ := interop.C4.ID()
	if *plan.C4.PreviousCheckpointID != interopC4 {
		t.Fatal("build C4 chains from a foreign checkpoint")
	}
	c4ID, _ := plan.C4.ID()
	if plan.C5.PreviousCheckpointID == nil || *plan.C5.PreviousCheckpointID != c4ID || plan.C5.Name != closuregraph.CheckpointC5 {
		t.Fatalf("C5 checkpoint = %#v", plan.C5)
	}
}

// A foreign or unchained interop closure is rejected before binding.
func TestForeignInteropClosureIsRejected(t *testing.T) {
	fixture := newFixture(t)
	fixture.materialize()
	capture, interop := fixture.closure()
	other := newFixture(t)
	other.files["Sources/App/main.swift"] = "import CLib\nprint(value() + 1)\n"
	other.materialize()
	otherCapture, _ := other.closure()
	_, err := NewPlan(t.Context(), fixture.build, otherCapture, interop)
	requireCode(t, err, CodeCheckpointInvalid)
	if _, err = NewPlan(t.Context(), fixture.build, nil, interop); ErrorCode(err) != CodeDerivationUnauthorized {
		t.Fatalf("absent capture = %v", err)
	}
	if _, err = NewPlan(t.Context(), fixture.build, capture, nil); ErrorCode(err) != CodeDerivationUnauthorized {
		t.Fatalf("absent interop closure = %v", err)
	}
}

// The planned command is the exact offline native invocation: prebuilts
// disabled, resolution forced, isolated roots, and the selected product.
func TestPlannedCommandIsOfflineFrozenAndIsolated(t *testing.T) {
	plan := newFixture(t).mustPlan()
	argv := strings.Join(plan.Command.Argv, " ")
	for _, expected := range []string{
		"--disable-experimental-prebuilts", "--force-resolved-versions", "--disable-netrc",
		"--build-system native", "--configuration " + fixtureConfiguration,
		"--product " + fixtureProduct, "--triple " + fixtureTriple,
		"--cache-path .curator/cache", "--config-path .curator/config",
		"--security-path .curator/security", "--scratch-path .curator/scratch",
	} {
		if !strings.Contains(argv, expected) {
			t.Fatalf("planned argv %q omits %q", argv, expected)
		}
	}
	for key, value := range plan.Command.Environment {
		if strings.Contains(value, os.TempDir()) {
			t.Fatalf("planned environment %s leaks a temporary path: %q", key, value)
		}
	}
	if plan.OutputPath != path.Join(".curator", "scratch", fixtureScratchTriple, fixtureConfiguration, fixtureProduct) {
		t.Fatalf("planned output path = %q", plan.OutputPath)
	}
}

// The planned command identity excludes every temporary path so two isolated
// execution roots produce the same portable identity.
func TestPlannedCommandIdentityIsPortable(t *testing.T) {
	first := newFixture(t).mustPlan()
	second := newFixture(t).mustPlan()
	if first.CommandID != second.CommandID {
		t.Fatalf("command identity is not portable: %s != %s", first.CommandID, second.CommandID)
	}
}

// The full offline build publishes the product through the protected store and
// an independently derived expected input reuses it exactly.
func TestOfflineBuildPublishesAndReusesExactly(t *testing.T) {
	fixture := newFixture(t)
	plan := fixture.mustPlan()
	manager := fixture.manager()
	result, err := manager.Build(t.Context(), plan)
	if err != nil {
		t.Fatalf("offline build failed: %v", err)
	}
	if result.CacheHit || result.ArtifactPath == "" {
		t.Fatalf("first build result = %#v", result)
	}
	payload, err := os.ReadFile(result.ArtifactPath)
	if err != nil || string(payload) != "curator-product" {
		t.Fatalf("published product = %q (%v)", payload, err)
	}
	if len(result.Observations) != 3 {
		t.Fatalf("observations = %#v", result.Observations)
	}
	classes := map[string]int{}
	for _, observation := range result.Observations {
		classes[observation.Class]++
		if !observation.SHA256.Valid() || observation.Size == 0 {
			t.Fatalf("observation %#v", observation)
		}
	}
	if classes["native.executable"] != 1 || classes["native.object"] != 2 {
		t.Fatalf("observation classes = %#v", classes)
	}
	if result.Execution.Network != "none" || result.Execution.Decision != "success" {
		t.Fatalf("execution receipt = %#v", result.Execution)
	}
	if !sortedUnique(result.Execution.WriteSet) || len(result.Execution.WriteSet) != 3 {
		t.Fatalf("write set = %#v", result.Execution.WriteSet)
	}
	if result.C6.Name != closuregraph.CheckpointC6 || result.C7.Name != closuregraph.CheckpointC7 {
		t.Fatalf("C6/C7 = %v %v", result.C6.Name, result.C7.Name)
	}
	c6ID, _ := result.C6.ID()
	if result.C7.PreviousCheckpointID == nil || *result.C7.PreviousCheckpointID != c6ID {
		t.Fatal("C7 does not chain from C6")
	}
	if fixture.starts != 1 {
		t.Fatalf("offline build started %d processes, want exactly one", fixture.starts)
	}

	replan := newFixture(t)
	replan.base = fixture.base
	replan.root, replan.sdkRoot = fixture.root, fixture.sdkRoot
	replan.execRoot, replan.outputRoot, replan.storeRoot = fixture.execRoot, fixture.outputRoot, fixture.storeRoot
	replan.files, replan.manifest = fixture.files, fixture.manifest
	second := replan.mustPlan()
	hit, err := replan.manager().Build(t.Context(), second)
	if err != nil {
		t.Fatalf("cache reuse failed: %v", err)
	}
	if !hit.CacheHit || hit.ArtifactPath != result.ArtifactPath || hit.AssuredCacheInput != result.AssuredCacheInput {
		t.Fatalf("independent expected input did not reuse the exact entry: %#v", hit)
	}
	if replan.starts != 0 {
		t.Fatalf("cache hit started %d processes", replan.starts)
	}
}

// The offline build never launches anything but the exact committed argv.
func TestOfflineBuildLaunchesExactlyTheCommittedCommand(t *testing.T) {
	fixture := newFixture(t)
	plan := fixture.mustPlan()
	if _, err := fixture.manager().Build(t.Context(), plan); err != nil {
		t.Fatal(err)
	}
	if len(fixture.launches) != 1 {
		t.Fatalf("launches = %d", len(fixture.launches))
	}
	launch := fixture.launches[0]
	if !equalStrings(launch.Argv, plan.Command.Argv) {
		t.Fatalf("launched argv = %v, want %v", launch.Argv, plan.Command.Argv)
	}
	if filepath.Base(launch.Executable) != filepath.Base(filepath.FromSlash(stubExecutableRelative())) {
		t.Fatalf("launched executable = %q", launch.Executable)
	}
	for _, entry := range launch.Environment {
		if strings.HasPrefix(entry, "PATH=") && !strings.HasPrefix(entry, "PATH="+fixture.execRoot) {
			t.Fatalf("build PATH escaped the execution root: %q", entry)
		}
	}
}

// A declared write the build never produced fails closed without publication.
func TestMissingDeclaredObjectFailsClosedWithoutPublication(t *testing.T) {
	fixture := newFixture(t)
	fixture.stubExtra = []stubAction{{Op: "remove-all", Path: ".curator/scratch/" + fixtureScratchTriple + "/" + fixtureConfiguration + "/CLib.build"}}
	plan := fixture.mustPlan()
	_, err := fixture.manager().Build(t.Context(), plan)
	requireCode(t, err, CodeOutputUnreceipted)
	requireNoPublication(t, fixture)
}

// A build that never wrote the declared product fails closed.
func TestMissingProductFailsClosedWithoutPublication(t *testing.T) {
	fixture := newFixture(t)
	fixture.stubExtra = []stubAction{{Op: "remove", Path: ".curator/scratch/" + fixtureScratchTriple + "/" + fixtureConfiguration + "/" + fixtureProduct}}
	plan := fixture.mustPlan()
	_, err := fixture.manager().Build(t.Context(), plan)
	if err == nil {
		t.Fatal("absent product was published")
	}
	requireNoPublication(t, fixture)
}

// A non-empty output root before the permitted action is unreceipted local
// output and fails closed.
func TestPreExistingOutputRootFailsClosed(t *testing.T) {
	fixture := newFixture(t)
	plan := fixture.mustPlan()
	if err := os.MkdirAll(fixture.outputRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture.outputRoot, "stale"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := fixture.manager().Build(t.Context(), plan)
	requireCode(t, err, CodeOutputUnreceipted)
	requireNoPublication(t, fixture)
}

// The staged tree is exactly the declared write set: the observed product plus
// one reconciliation record per declared intermediate object slot.
func TestStagedWriteSetIsExactlyTheDeclaredSet(t *testing.T) {
	fixture := newFixture(t)
	plan := fixture.mustPlan()
	result, err := fixture.manager().Build(t.Context(), plan)
	if err != nil {
		t.Fatal(err)
	}
	declared := map[string]bool{plan.OutputPath: true}
	for _, slot := range plan.Objects {
		declared[slot.Path] = true
	}
	if len(declared) != len(result.Execution.WriteSet) {
		t.Fatalf("declared %d write paths, receipt has %d", len(declared), len(result.Execution.WriteSet))
	}
	for _, value := range result.Execution.WriteSet {
		if !declared[value] {
			t.Fatalf("receipt write set contains undeclared %q", value)
		}
	}
}

// The dependency-file grammar every C-family and Swift driver emits is parsed
// exactly, including continuations and escaped separators.
func TestDependencyFileGrammarIsParsedExactly(t *testing.T) {
	payload := "out.o : /a/main.swift \\\n  /b/with\\ space.h /c/x.modulemap\n" +
		"/d/second.o: /e/second.c\n"
	got := parseDependencyFile(payload)
	want := []string{"/a/main.swift", "/b/with space.h", "/c/x.modulemap", "/e/second.c"}
	if !equalStrings(got, want) {
		t.Fatalf("parsed %#v, want %#v", got, want)
	}
	if classifyRead("/x/y.h") != "header" || classifyRead("/x/y.modulemap") != "module-map" || classifyRead("/x/y.swiftinterface") != "swift-module" || classifyRead("/x/y.c") != "source" || classifyRead("/x/y") != "input" {
		t.Fatal("observed read classification is not exact")
	}
}

// SwiftPM names its build directory after the triple with the platform version
// removed; the planned output path must reconcile with that exactly.
func TestScratchDirectoryReproducesSwiftPMTripleNaming(t *testing.T) {
	for triple, want := range map[string]string{
		"arm64-apple-macosx14.0":        "arm64-apple-macosx",
		"x86_64-unknown-linux-gnu":      "x86_64-unknown-linux-gnu",
		"arm64-apple-ios17.0-simulator": "arm64-apple-ios-simulator",
	} {
		if got := unversionedTriple(triple); got != want {
			t.Fatalf("unversionedTriple(%q) = %q, want %q", triple, got, want)
		}
	}
}

// Portable assurance keeps the reject-by-default read verdict: the observer
// reports not-observed rather than claiming coverage it cannot have.
func TestPortableObserverReportsNotObserved(t *testing.T) {
	fixture := newFixture(t)
	fixture.materialize()
	capture, _ := fixture.closure()
	observer, err := NewReadSetObserver(fixture.build, capture, fixture.source.Toolchain.SwiftPM, fixtureTriple)
	if err != nil {
		t.Fatal(err)
	}
	result, err := observer.ObserveReads(t.Context(), swiftpminterop.ReadSetRequest{Package: "root", Target: "CLib"})
	if err != nil || result.Observed || len(result.Reads) != 0 {
		t.Fatalf("portable observation = %#v (%v)", result, err)
	}
	if fixture.starts != 0 {
		t.Fatalf("portable observation started %d processes", fixture.starts)
	}
}

// An absent observer authority fails closed instead of degrading.
func TestObserverAuthorityContract(t *testing.T) {
	fixture := newFixture(t)
	fixture.materialize()
	capture, _ := fixture.closure()
	if _, err := NewReadSetObserver(Config{}, capture, fixture.source.Toolchain.SwiftPM, fixtureTriple); ErrorCode(err) != CodeDerivationUnauthorized {
		t.Fatal("incomplete observer authority was accepted")
	}
	if _, err := NewReadSetObserver(fixture.build, nil, fixture.source.Toolchain.SwiftPM, fixtureTriple); ErrorCode(err) != CodeDerivationUnauthorized {
		t.Fatal("observer without a capture was accepted")
	}
	var absent *ReadSetObserver
	if _, err := absent.ObserveReads(context.Background(), swiftpminterop.ReadSetRequest{}); ErrorCode(err) != CodeDerivationUnauthorized {
		t.Fatal("absent observer was accepted")
	}
}

// Observed reads of derived build state are locally produced outputs of the
// same permitted derivation, but a dependency checkout is admitted source and
// is rewritten to that dependency's protected root. A checkout that matches no
// admitted package identity fails closed rather than being dropped.
func TestObservedReadMappingSeparatesBuildTreeFromAdmittedSource(t *testing.T) {
	base := t.TempDir()
	work := filepath.Join(base, "exec", "work", "package")
	admitted := filepath.Join(base, "store", "root")
	dependency := filepath.Join(base, "store", "dep")
	roots := map[string]string{"root": admitted, "dep": dependency}
	resolved, generated, err := mapObservedRead(filepath.Join(work, "Sources", "App", "main.swift"), work, "root", roots)
	if err != nil || generated || resolved != filepath.Join(admitted, "Sources", "App", "main.swift") {
		t.Fatalf("admitted source mapping = %q generated=%v (%v)", resolved, generated, err)
	}
	if _, generated, err = mapObservedRead(filepath.Join(work, ".curator", "scratch", "x", "module.modulemap"), work, "root", roots); err != nil || !generated {
		t.Fatalf("build-tree read was not separated: generated=%v (%v)", generated, err)
	}
	checkout := filepath.Join(work, ".curator", "scratch", "checkouts", "dep", "Sources", "CDep", "include", "CDep.h")
	if resolved, generated, err = mapObservedRead(checkout, work, "root", roots); err != nil || generated || resolved != filepath.Join(dependency, "Sources", "CDep", "include", "CDep.h") {
		t.Fatalf("dependency checkout mapping = %q generated=%v (%v)", resolved, generated, err)
	}
	unknown := filepath.Join(work, ".curator", "scratch", "checkouts", "smuggled", "evil.h")
	if _, _, err = mapObservedRead(unknown, work, "root", roots); ErrorCode(err) != CodeHeaderInputUndeclared {
		t.Fatalf("unadmitted checkout read error = %v", err)
	}
	if _, _, err = mapObservedRead(filepath.Join(work, ".curator", "scratch", "checkouts"), work, "root", roots); ErrorCode(err) != CodeHeaderInputUndeclared {
		t.Fatal("a bare checkouts read was not rejected")
	}
	external := filepath.Join(base, "sdk", "usr", "include", "stdio.h")
	if resolved, generated, err = mapObservedRead(external, work, "root", roots); err != nil || generated || resolved != external {
		t.Fatalf("external read mapping = %q generated=%v (%v)", resolved, generated, err)
	}
}

// The verified observation pass must cover every package, not only the root
// one: a dependency target's compiler reads are rewritten to that dependency's
// admitted protected root, genuinely derived build state is dropped, and an
// external toolchain read is kept verbatim for the interop resolver.
func TestHarvestedReadSetCoversEveryAdmittedPackage(t *testing.T) {
	fixture := newFixture(t)
	fixture.addSourceControlDependency("dep", "CDep", map[string]string{
		"Sources/CDep/dep.c":          "#include \"CDep.h\"\nint dep(void) { return 7; }\n",
		"Sources/CDep/include/CDep.h": "#ifndef CDEP_H\n#define CDEP_H\nint dep(void);\n#endif\n",
	})
	fixture.materialize()
	capture, _ := fixture.closure()
	observer, err := NewReadSetObserver(fixture.build, capture, fixture.source.Toolchain.SwiftPM, fixtureTriple)
	if err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(fixture.execRoot, filepath.FromSlash(buildWorkMount))
	scratch := filepath.Join(work, filepath.FromSlash(swiftpmScratchDirectory(fixtureTriple, fixtureConfiguration)))
	checkout := filepath.Join(work, ".curator", "scratch", "checkouts", "dep")
	sdkHeader := filepath.Join(fixture.sdkRoot, "usr", "include", "stdio.h")
	writeDependencyFile(t, filepath.Join(scratch, "CDep.build"), "dep.c.d", []string{
		filepath.Join(checkout, "Sources", "CDep", "dep.c"),
		filepath.Join(checkout, "Sources", "CDep", "include", "CDep.h"),
		filepath.Join(scratch, "CDep.build", "module.modulemap"),
		sdkHeader,
	})
	writeDependencyFile(t, filepath.Join(scratch, "CLib.build"), "lib.c.d", []string{
		filepath.Join(work, "Sources", "CLib", "lib.c"),
	})
	observed, err := harvestDependencyFiles(observer, scratch)
	if err != nil {
		t.Fatalf("harvest failed: %v", err)
	}
	admitted := map[string]string{}
	for _, pkg := range capture.Packages {
		root, rootErr := pkg.ProtectedRoot()
		if rootErr != nil {
			t.Fatal(rootErr)
		}
		admitted[pkg.Identity] = root
	}
	want := []swiftpminterop.ObservedRead{
		{Path: filepath.Join(admitted["dep"], "Sources", "CDep", "dep.c"), Class: "source"},
		{Path: filepath.Join(admitted["dep"], "Sources", "CDep", "include", "CDep.h"), Class: "header"},
		{Path: sdkHeader, Class: "header"},
	}
	sort.Slice(want, func(i, j int) bool { return want[i].Path < want[j].Path })
	if !reflect.DeepEqual(observed["CDep"], want) {
		t.Fatalf("dependency read set = %#v, want %#v", observed["CDep"], want)
	}
	rootReads := observed["CLib"]
	if len(rootReads) != 1 || rootReads[0].Path != filepath.Join(admitted["root"], "Sources", "CLib", "lib.c") {
		t.Fatalf("root read set = %#v", rootReads)
	}
}

// A compiler read of a dependency checkout that matches no admitted package
// identity is undeclared closure input; the harvest fails closed rather than
// silently discarding it as build-tree state.
func TestHarvestedReadSetFailsClosedOnUnadmittedCheckout(t *testing.T) {
	fixture := newFixture(t)
	fixture.materialize()
	capture, _ := fixture.closure()
	observer, err := NewReadSetObserver(fixture.build, capture, fixture.source.Toolchain.SwiftPM, fixtureTriple)
	if err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(fixture.execRoot, filepath.FromSlash(buildWorkMount))
	scratch := filepath.Join(work, filepath.FromSlash(swiftpmScratchDirectory(fixtureTriple, fixtureConfiguration)))
	writeDependencyFile(t, filepath.Join(scratch, "CLib.build"), "lib.c.d", []string{
		filepath.Join(work, ".curator", "scratch", "checkouts", "smuggled", "Sources", "Evil", "evil.h"),
	})
	if _, err = harvestDependencyFiles(observer, scratch); ErrorCode(err) != CodeHeaderInputUndeclared {
		t.Fatalf("unadmitted checkout harvest error = %v", err)
	}
}

func writeDependencyFile(t *testing.T, directory, name string, reads []string) {
	t.Helper()
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	slashed := make([]string, 0, len(reads))
	for _, read := range reads {
		slashed = append(slashed, filepath.ToSlash(read))
	}
	payload := filepath.ToSlash(filepath.Join(directory, "out.o")) + " : " + strings.Join(slashed, " ") + "\n"
	if err := os.WriteFile(filepath.Join(directory, name), []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
}

// Every declared per-source object output is published from the exact bytes
// the compiler produced, under its declared logical path.
func TestDeclaredObjectOutputsArePublishedFromObservedBytes(t *testing.T) {
	fixture := newFixture(t)
	plan := fixture.mustPlan()
	result, err := fixture.manager().Build(t.Context(), plan)
	if err != nil {
		t.Fatal(err)
	}
	for _, slot := range plan.Objects {
		found := false
		for _, observation := range result.Observations {
			if observation.Path == slot.Path && observation.ExpectedOutputNodeID == slot.NodeID {
				found = true
			}
		}
		if !found {
			t.Fatalf("declared object %q has no observation", slot.Path)
		}
	}
	if plan.Objects[0].Path != ".curator/objects/root/App/Sources/App/main.swift.o" && plan.Objects[1].Path != ".curator/objects/root/App/Sources/App/main.swift.o" {
		t.Fatalf("declared object paths = %v", []string{plan.Objects[0].Path, plan.Objects[1].Path})
	}
}

func requireNoPublication(t *testing.T, fixture *fixture) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(fixture.storeRoot, "blobs"))
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("rejected build published %d protected blobs", len(entries))
	}
}

func sortedUnique(values []string) bool {
	for index := 1; index < len(values); index++ {
		if values[index-1] >= values[index] {
			return false
		}
	}
	return true
}

// driftedRecheck answers with a changed content fingerprint for exactly one
// selected role so the time-of-use boundary sees real drift.
func driftedRecheck(role string) func(context.Context, swiftpmsource.ToolIdentity) (closureexec.ToolchainIdentity, error) {
	return func(_ context.Context, expected swiftpmsource.ToolIdentity) (closureexec.ToolchainIdentity, error) {
		if expected.Role == role {
			return closureexec.ToolchainIdentity{Fingerprint: id('e'), ExecutableSHA256: expected.ExecutableSHA256}, nil
		}
		return closureexec.ToolchainIdentity{Fingerprint: expected.Fingerprint, ExecutableSHA256: expected.ExecutableSHA256}, nil
	}
}

func failingRecheck(context.Context, swiftpmsource.ToolIdentity) (closureexec.ToolchainIdentity, error) {
	return closureexec.ToolchainIdentity{}, os.ErrNotExist
}
