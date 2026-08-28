package swiftpmbuild

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpminterop"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

const (
	fixtureProduct       = "cli"
	fixtureTriple        = "arm64-apple-macosx14.0"
	fixtureScratchTriple = "arm64-apple-macosx"
	fixtureConfiguration = "release"
)

type fixture struct {
	t          *testing.T
	base       string
	root       string
	sdkRoot    string
	execRoot   string
	outputRoot string
	storeRoot  string

	files      map[string]string
	manifest   swiftpmsource.Manifest
	dependency map[string]swiftpmsource.Manifest
	snapshots  map[string]swiftpmsource.Snapshot
	source     swiftpmsource.Config
	interop    swiftpminterop.Config
	build      Config
	runner     *closureexec.ManagerProcessRunner
	launches   []closureexec.ProcessLaunch
	starts     int
	stubExtra  []stubAction

	materializeHook func()
}

func newFixture(t *testing.T) *fixture {
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
	value := &fixture{
		t: t, base: base, root: filepath.Join(base, "root"), sdkRoot: filepath.Join(base, "sdk"),
		execRoot: filepath.Join(base, "execution"), outputRoot: filepath.Join(base, "execution", "output"),
		storeRoot: filepath.Join(base, "protected"),
	}
	value.dependency = map[string]swiftpmsource.Manifest{}
	value.snapshots = map[string]swiftpmsource.Snapshot{}
	value.files = map[string]string{
		"Package.swift":               "// swift-tools-version: 6.1\nimport PackageDescription\n",
		"Package.resolved":            `{"version":3,"pins":[]}`,
		"Sources/App/main.swift":      "import CLib\nprint(value())\n",
		"Sources/CLib/lib.c":          "#include \"CLib.h\"\nint value(void) { return 1; }\n",
		"Sources/CLib/include/CLib.h": "#ifndef CLIB_H\n#define CLIB_H\nint value(void);\n#endif\n",
	}
	value.manifest = swiftpmsource.Manifest{
		PackageName: "root", ToolsVersion: "6.1",
		Products: []swiftpmsource.Product{{Name: fixtureProduct, Type: "executable", Targets: []string{"App"}}},
		Targets: []swiftpmsource.Target{
			{Name: "App", Type: "executable", Path: "Sources/App", Sources: []string{"Sources/App/main.swift"}, Dependencies: []swiftpmsource.TargetDependency{{Name: "CLib"}}},
			{Name: "CLib", Type: "regular", Path: "Sources/CLib", Sources: []string{"Sources/CLib/lib.c"}},
		},
	}
	value.writeTree(map[string]string{"usr/include/stdio.h": "int printf(const char *, ...);\n"}, value.sdkRoot)
	return value
}

// addSourceControlDependency binds one exact remote source-control dependency
// to the fixture: an admitted dependency tree, a captured local mirror, a
// frozen root lock pin, and the manifest edge the selected product reaches. It
// is what makes the mirror mount, the generated mirrors.json kind mapping, and
// the dependency read set exercisable end to end.
func (f *fixture) addSourceControlDependency(identity, target string, sources map[string]string) string {
	f.t.Helper()
	root := filepath.Join(f.base, "dependency-"+identity)
	files := map[string]string{"Package.swift": "// swift-tools-version: 6.1\nimport PackageDescription\n"}
	for relative, payload := range sources {
		files[relative] = payload
	}
	f.writeTree(files, root)
	mirror := filepath.Join(f.base, "mirror-"+identity)
	if err := os.MkdirAll(mirror, 0o700); err != nil {
		f.t.Fatal(err)
	}
	revision := strings.Repeat("a", 40)
	f.snapshots[identity] = swiftpmsource.Snapshot{
		Identity: identity, Root: root, MirrorRoot: mirror, Revision: revision,
		GitTree: strings.Repeat("b", 40), Kind: swiftpmsource.SourceRemote, BrokerReceiptID: id('b'),
	}
	location := "https://example.invalid/" + identity
	f.files["Package.resolved"] = `{"version":3,"pins":[{"identity":"` + identity + `","kind":"remoteSourceControl","location":"` + location + `","state":{"revision":"` + revision + `","version":"1.0.0"}}]}`
	f.manifest.Dependencies = append(f.manifest.Dependencies, swiftpmsource.ManifestDependency{Identity: identity, Kind: swiftpmsource.SourceRemote, Location: location, Requirement: "exact:1.0.0"})
	f.manifest.Targets[0].Dependencies = append(f.manifest.Targets[0].Dependencies, swiftpmsource.TargetDependency{Package: identity, Product: target + "Prod"})
	declared := []string{}
	for relative := range sources {
		if strings.HasSuffix(relative, ".c") || strings.HasSuffix(relative, ".swift") {
			declared = append(declared, relative)
		}
	}
	sort.Strings(declared)
	f.dependency[identity] = swiftpmsource.Manifest{
		PackageName: identity, ToolsVersion: "6.1",
		Products: []swiftpmsource.Product{{Name: target + "Prod", Type: "library", Targets: []string{target}}},
		Targets:  []swiftpmsource.Target{{Name: target, Type: "regular", Path: "Sources/" + target, Sources: declared}},
	}
	return root
}

func (f *fixture) writeTree(files map[string]string, root string) {
	f.t.Helper()
	for relative, payload := range files {
		full := filepath.Join(root, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
			f.t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(payload), 0o600); err != nil {
			f.t.Fatal(err)
		}
	}
}

var (
	swiftStubOnce  sync.Once
	swiftStubBytes []byte
	swiftStubErr   error
)

// builtSwiftStub compiles the scenario-driven SwiftPM stand-in once per test
// process and returns its executable bytes; every fixture installs its own
// copy below its private execution root.
func builtSwiftStub(t *testing.T) []byte {
	t.Helper()
	swiftStubOnce.Do(func() {
		output := filepath.Join(os.TempDir(), "curator-swiftpm-stub-"+strconv.Itoa(os.Getpid()))
		build := exec.Command("go", "build", "-o", output, "./testdata/swiftpm_stub")
		build.Env = os.Environ()
		if combined, err := build.CombinedOutput(); err != nil {
			swiftStubErr = fmt.Errorf("build swiftpm stub: %v\n%s", err, combined)
			return
		}
		swiftStubBytes, swiftStubErr = os.ReadFile(output)
		_ = os.Remove(output)
	})
	if swiftStubErr != nil {
		t.Fatal(swiftStubErr)
	}
	return swiftStubBytes
}

// stubAction is one declarative step of the SwiftPM driver stand-in. The
// stand-in was previously a POSIX shell script, which Windows cannot execute;
// the compiled scenario runner in testdata/swiftpm_stub reproduces the same
// observable build layout on every platform.
type stubAction struct {
	Op      string `json:"op"`
	Path    string `json:"path"`
	Payload string `json:"payload,omitempty"`
}

// swiftStubActions is the deterministic stand-in for the selected SwiftPM
// driver. It reproduces the exact native build layout: one product, one build
// directory per target, one object per source, and one compiler dependency
// file. {{PWD}} expands to the process working directory inside the stub.
func (f *fixture) swiftStubActions() []stubAction {
	scratch := ".curator/scratch/" + fixtureScratchTriple + "/" + fixtureConfiguration
	return append([]stubAction{
		{Op: "mkdir", Path: scratch + "/App.build"},
		{Op: "mkdir", Path: scratch + "/CLib.build"},
		{Op: "write", Path: scratch + "/" + fixtureProduct, Payload: "curator-product"},
		{Op: "chmod-readonly-exec", Path: scratch + "/" + fixtureProduct},
		{Op: "write", Path: scratch + "/App.build/main.swift.o", Payload: "app-object"},
		{Op: "write", Path: scratch + "/CLib.build/lib.c.o", Payload: "clib-object"},
		{Op: "write", Path: scratch + "/App.build/App.d",
			Payload: "{{PWD}}/" + scratch + "/App.build/main.swift.o : {{PWD}}/Sources/App/main.swift {{PWD}}/" + scratch + "/CLib.build/module.modulemap\n"},
		{Op: "write", Path: scratch + "/CLib.build/lib.c.d",
			Payload: "{{PWD}}/" + scratch + "/CLib.build/lib.c.o : {{PWD}}/Sources/CLib/lib.c {{PWD}}/Sources/CLib/include/CLib.h\n"},
		{Op: "write", Path: scratch + "/description.json", Payload: "{}"},
	}, f.stubExtra...)
}

func (f *fixture) materialize() {
	f.t.Helper()
	f.writeTree(f.files, f.root)
	if err := os.MkdirAll(filepath.Join(f.execRoot, "bin"), 0o700); err != nil {
		f.t.Fatal(err)
	}
	stubPath := filepath.Join(f.execRoot, filepath.FromSlash(stubExecutableRelative()))
	if err := os.RemoveAll(stubPath); err != nil {
		f.t.Fatal(err)
	}
	stub := builtSwiftStub(f.t)
	if err := os.WriteFile(stubPath, stub, 0o500); err != nil {
		f.t.Fatal(err)
	}
	scenario, err := json.Marshal(f.swiftStubActions())
	if err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(f.execRoot, "bin", "swiftpm-scenario.json"), scenario, 0o600); err != nil {
		f.t.Fatal(err)
	}
	sum := sha256.Sum256(stub)
	stubDigest := closuregraph.ID("sha256:" + hex.EncodeToString(sum[:]))

	runner, err := closureexec.NewManagerProcessRunner(f.execRoot, f.outputRoot)
	if err != nil {
		f.t.Fatal(err)
	}
	runner.ProcessStartObserver = func(closureexec.DerivationPermit) { f.starts++ }
	runner.ProcessLaunchObserver = func(launch closureexec.ProcessLaunch) { f.launches = append(f.launches, launch) }
	f.runner = runner
	executor, err := closureexec.NewAssuredExecutor(closureexec.DefaultAssuranceConfig(), runner, nil, "swiftpm-build-fixture")
	if err != nil {
		f.t.Fatal(err)
	}
	store, err := closureexec.NewCaptureStore(filepath.Join(f.base, "store"))
	if err != nil {
		f.t.Fatal(err)
	}
	driver := tool("swiftpm", '2')
	driver.Fingerprint, driver.ExecutableSHA256 = stubDigest, stubDigest
	tools := swiftpmsource.Toolchain{Swift: tool("swift", '1'), SwiftPM: driver, PackageDescription: tool("package-description", '3'), Git: tool("git", '4')}
	tools.Recheck = rechecker
	f.source = swiftpmsource.Config{
		Store: store, Policy: artifactpolicy.NewService(), Evaluator: &fakeEvaluator{root: f.manifest, dependency: f.dependency},
		Broker: fakeBroker{snapshots: f.snapshots}, MirrorVerifier: fakeMirrorVerifier{}, OfflineRunner: fakeOfflineRunner{}, Toolchain: tools,
		Destination: swiftpmsource.Destination{Platform: darwinPlatform(), Markers: map[string]string{"platform": "macos", "configuration": "release", "architecture": "arm64"}},
		CausalHead:  "sha256:" + strings.Repeat("0", 64),
	}
	f.interop = swiftpminterop.Config{
		Clang: tool("clang", '5'), ClangCXX: tool("clang++", '6'),
		SDK: swiftpminterop.ExternalComponent{
			Role: "macos-sdk", ExecutableRelativePath: "usr/bin/xcrun", PlatformABI: "darwin-arm64", PolicySelector: "apple-sdk-v1",
			VersionOutput: "MacOSX 26.5", Fingerprint: id('7'), SDKFactsDigest: id('8'), Roots: []string{filepath.Join(f.sdkRoot, "usr", "include")},
			Modules: []string{"Foundation"}, Frameworks: []string{"Foundation"}, Libraries: []string{"c", "System"},
		},
		Profile:   darwinProfile(),
		Assurance: closureexec.AssurancePortable,
		Recheck:   rechecker,
	}
	f.build = Config{
		Executor: executor, Store: store, Policy: artifactpolicy.NewService(),
		ExecutionRoot: f.execRoot, OutputRoot: f.outputRoot, StoreRoot: f.storeRoot,
		Linker: swiftpminterop.ExternalComponent{
			Role: "macos-linker", ExecutableRelativePath: "usr/bin/ld", PlatformABI: "darwin-arm64",
			PolicySelector: "apple-linker-v1", VersionOutput: "ld 1234", Fingerprint: id('9'), ExecutableSHA256: id('a'),
		},
		Configuration: fixtureConfiguration, Assurance: closureexec.AssurancePortable,
		AllowedProcesses: []string{stubExecutableRelative()},
		Slots: map[ToolSlot]string{
			SlotSwiftPM: "swiftpm", SlotSwiftCompiler: "swift", SlotPackageDescription: "package-description",
			SlotClang: "clang", SlotSDK: "macos-sdk",
		},
		Recheck: rechecker, CausalHead: "sha256:" + strings.Repeat("0", 64),
	}
	if f.materializeHook != nil {
		f.materializeHook()
	}
}

func (f *fixture) closure() (*swiftpmsource.Capture, *swiftpminterop.Result) {
	f.t.Helper()
	capture, err := swiftpmsource.CaptureAndClose(f.t.Context(), f.source, swiftpmsource.Request{Root: f.root, Product: fixtureProduct, Resolved: []byte(f.files["Package.resolved"])})
	if err != nil {
		f.t.Fatalf("source closure failed: %v", err)
	}
	interop, err := swiftpminterop.Close(f.t.Context(), f.interop, capture)
	if err != nil {
		f.t.Fatalf("interop closure failed: %v", err)
	}
	return capture, interop
}

func (f *fixture) plan() (*Plan, error) {
	f.t.Helper()
	f.materialize()
	capture, interop := f.closure()
	return NewPlan(f.t.Context(), f.build, capture, interop)
}

func (f *fixture) mustPlan() *Plan {
	f.t.Helper()
	plan, err := f.plan()
	if err != nil {
		f.t.Fatalf("build planning failed: %v", err)
	}
	return plan
}

func (f *fixture) manager() *Manager {
	f.t.Helper()
	manager, err := NewManager(f.build)
	if err != nil {
		f.t.Fatal(err)
	}
	return manager
}

func darwinPlatform() closuregraph.TargetPlatformPayload {
	return closuregraph.TargetPlatformPayload{OS: "darwin", Architecture: "arm64", ABI: "darwin", Libc: "libSystem", MinimumRuntime: "macos-14", SDKID: "macos-sdk-v1", TargetTriple: fixtureTriple, Runtime: "swift-6", LanguageModes: map[string]string{"swift": "6"}, Tuning: map[string]string{}}
}

func darwinProfile() swiftpminterop.PlatformProfile {
	return swiftpminterop.PlatformProfile{ID: "apple-swift-6.3.2-macos-26.5-v1", TargetTriples: []string{fixtureTriple}, CxxInterop: true, ObjectiveCRuntime: "objc4", CxxStandardModes: []string{"c++17"}}
}

func rechecker(_ context.Context, expected swiftpmsource.ToolIdentity) (closureexec.ToolchainIdentity, error) {
	return closureexec.ToolchainIdentity{Fingerprint: expected.Fingerprint, ExecutableSHA256: expected.ExecutableSHA256}, nil
}

var hexAlphabet = "0123456789abcdef"

// stubExecutableRelative is the driver stub's permit-relative path. Windows
// CreateProcess resolution in os/exec requires a PATHEXT extension, so the
// same fixture identity carries the platform's real executable spelling.
func stubExecutableRelative() string {
	if runtime.GOOS == "windows" {
		return "bin/swiftpm.exe"
	}
	return "bin/swiftpm"
}

func tool(role string, value byte) swiftpmsource.ToolIdentity {
	relative := "bin/" + strings.ReplaceAll(role, "+", "x")
	if role == "swiftpm" {
		relative = stubExecutableRelative()
	}
	return swiftpmsource.ToolIdentity{
		Role: role, ExecutableRelativePath: relative,
		VersionOutput: role + " 6.3.2 (Apple Swift version 6.3.2)", PlatformABI: "darwin-arm64", PolicySelector: "swift-toolchain-v1",
		Fingerprint: id(value), ExecutableSHA256: id(hexAlphabet[(strings.IndexByte(hexAlphabet, value)+8)%16]),
	}
}

func id(value byte) closuregraph.ID {
	return closuregraph.ID("sha256:" + strings.Repeat(string(value), 64))
}

type fakeEvaluator struct {
	root       swiftpmsource.Manifest
	dependency map[string]swiftpmsource.Manifest
}

func (e *fakeEvaluator) Evaluate(_ context.Context, root string, permit swiftpmsource.ManifestPermit) (swiftpmsource.ManifestResult, error) {
	if permit.Network != "none" || !filepath.IsAbs(root) {
		return swiftpmsource.ManifestResult{}, errors.New("open manifest permit")
	}
	prebuilts := false
	for _, argument := range permit.Argv {
		if argument == "--disable-experimental-prebuilts" {
			prebuilts = true
		}
	}
	if !prebuilts {
		return swiftpmsource.ManifestResult{}, errors.New("SwiftPM permit omits --disable-experimental-prebuilts")
	}
	manifest := e.root
	if declared, present := e.dependency[permit.PackageIdentity]; present {
		manifest = declared
	}
	receipt, _ := closuregraph.DomainID("fake-swiftpm-manifest-receipt-v1", map[string]any{"package": permit.PackageIdentity, "permit": string(permit.ID)})
	return swiftpmsource.ManifestResult{Manifest: manifest, ReceiptID: receipt}, nil
}

type fakeBroker struct {
	snapshots map[string]swiftpmsource.Snapshot
}

func (b fakeBroker) Acquire(_ context.Context, pin swiftpmsource.Pin) (swiftpmsource.Snapshot, error) {
	snapshot, present := b.snapshots[pin.Identity]
	if !present {
		return swiftpmsource.Snapshot{}, errors.New("fixture has no snapshot for " + pin.Identity)
	}
	return snapshot, nil
}

type fakeMirrorVerifier struct{}

func (fakeMirrorVerifier) Verify(_ context.Context, _ string, pin swiftpmsource.Pin, snapshot swiftpmsource.Snapshot) (swiftpmsource.GitVerificationEvidence, error) {
	if snapshot.Identity != pin.Identity || snapshot.Kind != pin.Kind || snapshot.Revision != pin.Revision || snapshot.GitTree == "" {
		return swiftpmsource.GitVerificationEvidence{}, errors.New("mirror identity mismatch")
	}
	return swiftpmsource.GitVerificationEvidence{}, nil
}

type fakeOfflineRunner struct{}

func (fakeOfflineRunner) Replay(_ context.Context, capture *swiftpmsource.Capture) (swiftpmsource.OfflineMetadataResult, error) {
	identities := []string{}
	for _, pin := range capture.Lock.Pins {
		identities = append(identities, pin.Identity)
	}
	sort.Strings(identities)
	values := make([]any, len(identities))
	for index, identity := range identities {
		values[index] = identity
	}
	receipt, _ := closuregraph.DomainID("swiftpm-test-offline-receipt-v1", map[string]any{"identities": values})
	return swiftpmsource.OfflineMetadataResult{ReceiptID: receipt, PackageIdentities: identities}, nil
}

func requireCode(t *testing.T, err error, want Code) {
	t.Helper()
	if ErrorCode(err) != want {
		t.Fatalf("error code = %q, want %q (err=%v)", ErrorCode(err), want, err)
	}
}
