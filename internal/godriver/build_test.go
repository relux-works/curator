package godriver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"go/build"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
)

func wantBuildArgv(artifact string) []string {
	return append(append([]string(nil), buildArgumentPrefix...), artifact, ".")
}

func TestBuildRunsExactlyOneFixedListAndBuildInsideTheWorker(t *testing.T) {
	fixture := newSnapshotFixture(t)
	fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "verified executable"})

	limits := ResourceLimits{Timeout: 30 * time.Second, OutputBytes: 64 * 1024, ArtifactBytes: 1024, DiskBytes: 8 * 1024 * 1024, MemoryBytes: 64 * 1024 * 1024, Processes: 8}
	result, err := Build(context.Background(), fixture.request(limits))
	if err != nil {
		t.Fatal(err)
	}
	calls := fixture.sourceAwareCalls()
	if len(calls) != 2 {
		t.Fatalf("source-aware calls = %d, want exactly one list and one build", len(calls))
	}
	if !reflect.DeepEqual(calls[0].Argv, listArguments) {
		t.Fatalf("list argv = %q", calls[0].Argv)
	}
	if !reflect.DeepEqual(calls[1].Argv, wantBuildArgv(result.Artifact.StagedPath)) {
		t.Fatalf("build argv = %q", calls[1].Argv)
	}
	for _, call := range calls {
		if mustPhysical(t, call.Dir) != fixture.sourceDir {
			t.Fatalf("Go child cwd = %q, want the canonical source directory %q", call.Dir, fixture.sourceDir)
		}
		assertFixedBuildEnvironment(t, call.Environment, fixture.session)
	}
	digest := sha256.Sum256([]byte("verified executable"))
	wantPath, err := buildmeta.ArtifactPath("golden-tool", fixture.session.Target().GOOS)
	if err != nil {
		t.Fatal(err)
	}
	if result.Artifact.Metadata.Path != wantPath ||
		result.Artifact.Metadata.Size != int64(len("verified executable")) ||
		result.Artifact.Metadata.SHA256 != "sha256:"+hex.EncodeToString(digest[:]) {
		t.Fatalf("artifact metadata = %+v", result.Artifact.Metadata)
	}
	info, err := os.Lstat(result.Artifact.StagedPath)
	if err != nil || !info.Mode().IsRegular() {
		t.Fatalf("staged artifact = %v, %v", info, err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("manager executable permissions were not applied: %v", info.Mode())
	}
	// The stub records every invocation. Exactly the three package-independent
	// probes plus the two source-aware vectors ran: the built output was never
	// started for validation, version discovery, or any other reason.
	if all := fixture.calls(); len(all) != 5 {
		t.Fatalf("recorded launcher calls = %d, want three probes plus list and build", len(all))
	}
}

func TestBuildEmitsExactlyOneClosedCapabilityEvidenceRecord(t *testing.T) {
	fixture := newSnapshotFixture(t)
	fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
	result, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 30 * time.Second}))
	if err != nil {
		t.Fatal(err)
	}
	evidence := result.Evidence
	if evidence.RecordVersion != CapabilityEvidenceVersion || evidence.ExecutionPolicy != ExecutionPolicy {
		t.Fatalf("evidence identity = %+v", evidence)
	}
	if evidence.Platform != InventoryPlatform(runtime.GOOS) {
		t.Fatalf("evidence platform = %q", evidence.Platform)
	}
	if len(evidence.Controls) != len(nativeControlInventory) {
		t.Fatalf("evidence entries = %d, want exactly one per inventory control", len(evidence.Controls))
	}
	records := nativeControlPlatforms[evidence.Platform]
	for index, entry := range evidence.Controls {
		if entry.Name != nativeControlInventory[index] || entry.ProbedAt != ProbeTiming {
			t.Fatalf("entry %d = %+v", index, entry)
		}
		want := records[entry.Name].Availability
		if entry.Availability != want {
			t.Fatalf("entry %q availability = %q, want the inventory value %q", entry.Name, entry.Availability, want)
		}
		if want == AvailabilityAvailable && entry.Status != StatusApplied {
			t.Fatalf("available control %q was not applied", entry.Name)
		}
		if want == AvailabilityUnavailable && entry.Status != StatusUnavailable {
			t.Fatalf("unavailable control %q reports %q", entry.Name, entry.Status)
		}
		if isDeferredHardenedGuarantee(entry.Name) {
			t.Fatalf("evidence claims deferred hardened guarantee %q", entry.Name)
		}
	}
}

func TestBuildStopsBeforeBuildForEveryPreflightRejectionClass(t *testing.T) {
	tests := []struct {
		name string
		code string
		edit func(*packageJSON)
	}{
		{name: "non main", code: "build_package_not_main", edit: func(item *packageJSON) { item.Name = "library" }},
		{name: "incomplete", code: "go_list_incomplete", edit: func(item *packageJSON) { item.Incomplete = true }},
		{name: "package error", code: "go_list_incomplete", edit: func(item *packageJSON) { item.Error = &packageError{Err: "bad"} }},
		{name: "dependency error", code: "go_list_incomplete", edit: func(item *packageJSON) { item.DepsErrors = []*packageError{{Err: "bad"}} }},
		{name: "test package", code: "go_test_input_forbidden", edit: func(item *packageJSON) { item.ForTest = "x" }},
		{name: "cgo", code: "cgo_required", edit: func(item *packageJSON) { item.CgoFiles = []string{"cgo.go"} }},
		{name: "native", code: "go_native_input_forbidden", edit: func(item *packageJSON) { item.CFiles = []string{"x.c"} }},
		{name: "assembly", code: "go_assembly_forbidden", edit: func(item *packageJSON) { item.SFiles = []string{"x.s"} }},
		{name: "syso", code: "go_syso_forbidden", edit: func(item *packageJSON) { item.SysoFiles = []string{"x.syso"} }},
		{name: "escaped source", code: "go_source_input_escape", edit: func(item *packageJSON) { item.GoFiles = []string{"../../../../outside.go"} }},
		{name: "escaped embed", code: "go_embed_input_escape", edit: func(item *packageJSON) { item.EmbedFiles = []string{"../../../../outside.txt"} }},
		{name: "replaced module", code: "vendor_metadata_inconsistent", edit: func(item *packageJSON) { item.Module.Replace = &moduleJSON{Path: "evil"} }},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newSnapshotFixture(t)
			item := fixture.rootPackage()
			testCase.edit(&item)
			fixture.start(stubScript{ListStdout: string(encodePackages(t, item)), Artifact: "artifact"})
			_, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 30 * time.Second}))
			if DiagnosticCode(err) != testCase.code {
				t.Fatalf("error = %v, want %s", err, testCase.code)
			}
			if calls := fixture.sourceAwareCalls(); len(calls) != 1 || calls[0].Argv[0] != "list" {
				t.Fatalf("calls = %+v, the compiler must not start after a preflight rejection", calls)
			}
		})
	}
}

func TestBuildRejectsDirectivesPGOAndMultipleRootsBeforeBuild(t *testing.T) {
	for _, testCase := range []struct {
		name, content, code string
		multiple            bool
	}{
		{name: "cgo import dynamic", content: "package main\n//go:cgo_import_dynamic x y z\nfunc main() {}\n", code: "go_forbidden_compiler_directive"},
		{name: "generator", content: "package main\n//go:generate forbidden\nfunc main() {}\n", code: "go_generator_forbidden"},
		{name: "pgo", content: "package main\nfunc main() {}\n", code: "go_pgo_forbidden"},
		{name: "multiple roots", content: "package main\nfunc main() {}\n", code: "build_package_ambiguous", multiple: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newSnapshotFixture(t)
			writeTestFile(t, filepath.Join(fixture.sourceDir, "main.go"), []byte(testCase.content), 0o644)
			if testCase.name == "pgo" {
				writeTestFile(t, filepath.Join(fixture.sourceDir, "default.pgo"), []byte("attempt"), 0o644)
			}
			item := fixture.rootPackage()
			items := []packageJSON{item}
			if testCase.multiple {
				second := item
				second.ImportPath = "example.test/other"
				items = append(items, second)
			}
			fixture.start(stubScript{ListStdout: string(encodePackages(t, items...)), Artifact: "artifact"})
			_, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 30 * time.Second}))
			if DiagnosticCode(err) != testCase.code {
				t.Fatalf("error = %v, want %s", err, testCase.code)
			}
			if calls := fixture.sourceAwareCalls(); len(calls) != 1 {
				t.Fatalf("calls = %d", len(calls))
			}
		})
	}
}

func TestBuildRejectsArtifactLinksOversizeAndUnexpectedOutputs(t *testing.T) {
	for _, testCase := range []struct {
		name, code string
		script     func(*workerFixture, stubScript) stubScript
	}{
		{name: "symlink", code: "artifact_special_file", script: func(_ *workerFixture, script stubScript) stubScript {
			script.ArtifactMode = "symlink"
			return script
		}},
		{name: "oversize", code: "artifact_size_limit", script: func(_ *workerFixture, script stubScript) stubScript {
			script.ArtifactPad = 4096
			return script
		}},
		{name: "extra output", code: "artifact_output_invalid", script: func(_ *workerFixture, script stubScript) stubScript {
			script.ExtraOutput = "extra"
			return script
		}},
		{name: "missing output", code: "artifact_output_invalid", script: func(_ *workerFixture, script stubScript) stubScript {
			script.ArtifactMode = "none"
			return script
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newSnapshotFixture(t)
			script := stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"}
			fixture.start(testCase.script(fixture, script))
			limits := ResourceLimits{Timeout: 30 * time.Second, ArtifactBytes: 64, DiskBytes: 8 * 1024 * 1024}
			_, err := Build(context.Background(), fixture.request(limits))
			if DiagnosticCode(err) != testCase.code {
				t.Fatalf("error = %v, want %s", err, testCase.code)
			}
			if calls := fixture.sourceAwareCalls(); len(calls) != 2 {
				t.Fatalf("calls = %d, want list and build", len(calls))
			}
		})
	}
}

func TestBuildRejectsHardLinkedArtifact(t *testing.T) {
	fixture := newSnapshotFixture(t)
	outside := filepath.Join(t.TempDir(), "outside-artifact")
	writeTestFile(t, outside, []byte("linked"), 0o600)
	fixture.start(stubScript{
		ListStdout: string(encodePackages(t, fixture.rootPackage())),
		// The manager staging root and the link source share one filesystem
		// only when the temporary directories do; skip otherwise.
		ArtifactMode: "hardlink", HardlinkSource: outside,
	})
	_, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 30 * time.Second}))
	switch DiagnosticCode(err) {
	case "artifact_link":
	case "go_build_failed":
		t.Skipf("hard links across the staging filesystem are unavailable: %v", err)
	default:
		t.Fatalf("error = %v, want artifact_link", err)
	}
}

func TestBuildEnforcesOutputDeadlineDiskAndNonzeroExit(t *testing.T) {
	for _, testCase := range []struct {
		name, code string
		script     func(stubScript) stubScript
		limits     ResourceLimits
	}{
		{name: "combined output", code: "process_output_limit", script: func(script stubScript) stubScript {
			script.ListPadBytes = 8192
			return script
		}, limits: ResourceLimits{Timeout: 30 * time.Second, OutputBytes: 1024}},
		{name: "deadline", code: "process_timeout", script: func(script stubScript) stubScript {
			script.ListSleepMS = 20000
			return script
		}, limits: ResourceLimits{Timeout: time.Second}},
		{name: "nonzero exit", code: "go_list_failed", script: func(script stubScript) stubScript {
			script.ListExit = 1
			script.ListStderr = "compiler failed"
			return script
		}, limits: ResourceLimits{Timeout: 30 * time.Second}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newSnapshotFixture(t)
			script := stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"}
			fixture.start(testCase.script(script))
			operation := fixture.session.operation
			_, err := Build(context.Background(), fixture.request(testCase.limits))
			if DiagnosticCode(err) != testCase.code {
				t.Fatalf("error = %v, want %s", err, testCase.code)
			}
			entries, readErr := os.ReadDir(operation)
			if readErr != nil {
				t.Fatal(readErr)
			}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".curator-go-build-") {
					t.Fatalf("failed private staging was not removed: %s", entry.Name())
				}
			}
		})
	}
}

func TestBuildRejectsOversizedPrivateState(t *testing.T) {
	fixture := newSnapshotFixture(t)
	fixture.start(stubScript{
		ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact",
		PrivateFileName: "oversized", PrivateFileBytes: 4 * 1024 * 1024,
	})
	limits := ResourceLimits{Timeout: 30 * time.Second, ArtifactBytes: 1024, FileBytes: 8 * 1024 * 1024, DiskBytes: 2 * 1024 * 1024}
	_, err := Build(context.Background(), fixture.request(limits))
	if DiagnosticCode(err) != "process_disk_limit" {
		t.Fatalf("error = %v, want process_disk_limit", err)
	}
}

func TestBuildClassifiesOfflineVendorAndToolchainFailures(t *testing.T) {
	for _, testCase := range []struct{ name, stderr, code string }{
		{name: "missing vendor", stderr: "cannot find module providing package example.test/missing: import lookup disabled by -mod=vendor", code: "vendor_dependency_missing"},
		{name: "inconsistent vendor", stderr: "go: inconsistent vendoring in /source", code: "vendor_metadata_inconsistent"},
		{name: "toolchain", stderr: "go.mod requires go >= 1.99.0 (running go 1.25.5; GOTOOLCHAIN=local)", code: "toolchain_switch_forbidden"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newSnapshotFixture(t)
			fixture.start(stubScript{ListExit: 1, ListStderr: testCase.stderr})
			_, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 30 * time.Second}))
			if DiagnosticCode(err) != testCase.code {
				t.Fatalf("error = %v, want %s", err, testCase.code)
			}
			if calls := fixture.sourceAwareCalls(); len(calls) != 1 {
				t.Fatalf("calls = %d", len(calls))
			}
		})
	}
}

func TestBuildRejectsWorkspaceAndPackageSelectedToolchainBeforeGo(t *testing.T) {
	for _, testCase := range []struct {
		name, code string
		write      func(*workerFixture)
	}{
		{name: "workspace", code: "workspace_dependency_forbidden", write: func(fixture *workerFixture) {
			writeTestFile(fixture.t, filepath.Join(fixture.buildRoot, "go.work"), []byte("go 1.23\nuse .\n"), 0o644)
		}},
		{name: "toolchain", code: "toolchain_switch_forbidden", write: func(fixture *workerFixture) {
			writeTestFile(fixture.t, filepath.Join(fixture.buildRoot, "go.mod"), []byte("module example.test/build\n\ngo 1.23\ntoolchain\tgo1.99.0\n"), 0o644)
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newSnapshotFixture(t)
			testCase.write(fixture)
			fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
			_, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 30 * time.Second}))
			if DiagnosticCode(err) != testCase.code {
				t.Fatalf("error = %v, want %s", err, testCase.code)
			}
			if calls := fixture.sourceAwareCalls(); len(calls) != 0 {
				t.Fatalf("calls = %d, no Go child may start", len(calls))
			}
		})
	}
}

func TestBuildDetectsSourceMutationThroughTheLastChild(t *testing.T) {
	fixture := newSnapshotFixture(t)
	fixture.start(stubScript{
		ListStdout:          string(encodePackages(t, fixture.rootPackage())),
		Artifact:            "artifact",
		MutateSourcePath:    filepath.Join(fixture.sourceDir, "main.go"),
		MutateSourceContent: "package main\nfunc main() { panic(\"mutated\") }\n",
	})
	_, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 30 * time.Second}))
	if DiagnosticCode(err) != "source_mutated" {
		t.Fatalf("error = %v, want source_mutated", err)
	}
	if calls := fixture.sourceAwareCalls(); len(calls) != 1 {
		t.Fatalf("calls = %d, the compiler must not start after source mutation", len(calls))
	}
}

func TestBuildRejectsEveryPackageInfluenceSurface(t *testing.T) {
	for _, surface := range []string{
		"executable", "args", "env", "output", "flags", "hooks", "plugins", "generate",
	} {
		t.Run(surface, func(t *testing.T) {
			fixture := newSnapshotFixture(t)
			fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
			request := fixture.request(ResourceLimits{Timeout: 30 * time.Second})
			request.CommandObject[surface] = "package-selected"
			_, err := Build(context.Background(), request)
			if DiagnosticCode(err) != CodePackageInfluenceForbidden {
				t.Fatalf("error = %v, want %s", err, CodePackageInfluenceForbidden)
			}
			if calls := fixture.sourceAwareCalls(); len(calls) != 0 {
				t.Fatalf("calls = %d, package influence must fail before the worker", len(calls))
			}
		})
	}
}

func TestBuildRequiresThePackageCommandSurface(t *testing.T) {
	fixture := newSnapshotFixture(t)
	fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
	request := fixture.request(ResourceLimits{Timeout: 30 * time.Second})
	request.CommandObject = nil
	if _, err := Build(context.Background(), request); DiagnosticCode(err) != CodePackageInfluenceForbidden {
		t.Fatalf("error = %v, want %s", err, CodePackageInfluenceForbidden)
	}
	request = fixture.request(ResourceLimits{Timeout: 30 * time.Second})
	request.CommandObject["driver"] = "go-repository-v1"
	if _, err := Build(context.Background(), request); DiagnosticCode(err) != CodePackageInfluenceForbidden {
		t.Fatalf("error = %v, want %s", err, CodePackageInfluenceForbidden)
	}
}

func TestBuildTerminatesTheCompleteWorkerDomain(t *testing.T) {
	fixture := newSnapshotFixture(t)
	pidFile := filepath.Join(t.TempDir(), "descendants")
	fixture.start(stubScript{
		ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact",
		SpawnChildren: 2, SpawnPidFile: pidFile, SpawnSeconds: 120,
	})
	if _, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 60 * time.Second})); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(pidFile) // #nosec G304 -- test-owned path
	if err != nil {
		t.Fatalf("tool children did not publish their identities: %v", err)
	}
	pids := strings.Fields(string(payload))
	if len(pids) != 2 {
		t.Fatalf("descendant pids = %v, want two", pids)
	}
	deadline := time.Now().Add(15 * time.Second)
	for _, raw := range pids {
		pid, convErr := strconv.Atoi(raw)
		if convErr != nil {
			t.Fatal(convErr)
		}
		for processAlive(pid) && time.Now().Before(deadline) {
			time.Sleep(50 * time.Millisecond)
		}
		if processAlive(pid) {
			t.Fatalf("descendant %d outlived the worker domain teardown", pid)
		}
	}
}

// TestPerFileSizeLimitIsReallyApplied proves the macOS inventory control is a
// real kernel limit inherited by the compiler, not a declaration: a private
// write above the bound fails inside the Go child.
func TestPerFileSizeLimitIsReallyApplied(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("per-file-size-limit is available only on macOS in rc5-native-control-inventory-v1")
	}
	fixture := newSnapshotFixture(t)
	fixture.start(stubScript{
		ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact",
		PrivateFileName: "oversized", PrivateFileBytes: 4 * 1024 * 1024,
	})
	limits := ResourceLimits{Timeout: 30 * time.Second, ArtifactBytes: 1024, FileBytes: 1024 * 1024, DiskBytes: defaultDiskLimit}
	_, err := Build(context.Background(), fixture.request(limits))
	if DiagnosticCode(err) != "go_build_failed" {
		t.Fatalf("error = %v, want the compiler to fail under RLIMIT_FSIZE", err)
	}
}

func TestRealGoV1VendoredBuildIsBoundedAndNotLaunched(t *testing.T) {
	if os.Getenv("CURATOR_REAL_GO_BUILD_TEST") != "1" {
		t.Skip("set CURATOR_REAL_GO_BUILD_TEST=1 for the bounded native go-v1 integration")
	}
	snapshot := filepath.Join(t.TempDir(), "snapshot")
	fixtureRoot := "testdata/realbuild"
	buildRootRel := "build"
	sourceDirRel := "build/cmd/golden-tool"
	if candidate := os.Getenv("CURATOR_GO_BUILD_FIXTURE"); candidate != "" {
		fixtureRoot = candidate
	}
	if err := os.CopyFS(snapshot, os.DirFS(fixtureRoot)); err != nil {
		t.Fatal(err)
	}
	snapshot = mustPhysical(t, snapshot)
	token, err := buildsource.Validate(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = token.Close() }()

	config := ConfigFromEnvironment(t.TempDir(), snapshot)
	config.CuratorGo = filepath.Join(build.Default.GOROOT, "bin", platformGoName)
	session, err := Establish(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = session.Close() }()

	result, err := Build(context.Background(), BuildRequest{
		Session: session, Source: token,
		CommandObject: BuildCommand{"type": "build", "driver": "go-v1", "source_dir": sourceDirRel},
		BuildRoot:     buildRootRel, SourceDir: sourceDirRel, Command: "golden-tool",
		Limits: ResourceLimits{
			Timeout: 90 * time.Second, OutputBytes: 2 * 1024 * 1024, ArtifactBytes: 64 * 1024 * 1024,
			FileBytes: defaultFileLimit, DiskBytes: defaultDiskLimit, MemoryBytes: defaultMemoryLimit, Processes: defaultProcessLimit,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Artifact.Metadata.Size <= 0 || result.Artifact.Metadata.SHA256 == "" {
		t.Fatalf("artifact = %+v", result.Artifact)
	}
	if result.Evidence.ExecutionPolicy != ExecutionPolicy || len(result.Evidence.Controls) != len(nativeControlInventory) {
		t.Fatalf("evidence = %+v", result.Evidence)
	}
	// Deliberately do not execute result.Artifact.StagedPath.
}

func assertFixedBuildEnvironment(t *testing.T, environment []string, session *Session) {
	t.Helper()
	if !reflect.DeepEqual(environment, session.Environment()) {
		t.Fatalf("environment changed\n got: %q\nwant: %q", environment, session.Environment())
	}
	values := environmentMap(environment)
	for key, want := range map[string]string{
		"GOPROXY": "off", "GOSUMDB": "off", "GOVCS": "*:off", "GOWORK": "off",
		"GOTOOLCHAIN": "local", "CGO_ENABLED": "0", "GO_EXTLINK_ENABLED": "0", "GOFLAGS": "", "GOENV": "off",
	} {
		if values[key] != want {
			t.Fatalf("%s = %q, want %q", key, values[key], want)
		}
	}
	if entries, err := os.ReadDir(values["PATH"]); err != nil || len(entries) != 0 {
		t.Fatalf("PATH is not an empty manager-owned directory: %v %v", entries, err)
	}
}
