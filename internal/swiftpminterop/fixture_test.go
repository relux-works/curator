package swiftpminterop

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

type fixture struct {
	t          *testing.T
	base       string
	root       string
	sdkRoot    string
	manifest   swiftpmsource.Manifest
	source     swiftpmsource.Config
	interop    Config
	files      map[string]string
	manifests  map[string]swiftpmsource.Manifest
	systemRoot string

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
	value := &fixture{t: t, base: base, root: filepath.Join(base, "root"), sdkRoot: filepath.Join(base, "sdk"), systemRoot: filepath.Join(base, "system")}
	value.files = map[string]string{
		"Package.swift":               "// swift-tools-version: 6.1\nimport PackageDescription\n",
		"Package.resolved":            `{"version":3,"pins":[]}`,
		"Sources/App/main.swift":      "import CLib\nprint(value())\n",
		"Sources/CLib/lib.c":          "#include \"CLib.h\"\n#include <stdio.h>\nint value(void) { return 1; }\n",
		"Sources/CLib/include/CLib.h": "#ifndef CLIB_H\n#define CLIB_H\nint value(void);\n#endif\n",
	}
	value.manifest = swiftpmsource.Manifest{
		PackageName: "root", ToolsVersion: "6.1",
		Products: []swiftpmsource.Product{{Name: "cli", Type: "executable", Targets: []string{"App"}}},
		Targets: []swiftpmsource.Target{
			{Name: "App", Type: "executable", Path: "Sources/App", Sources: []string{"Sources/App/main.swift"}, Dependencies: []swiftpmsource.TargetDependency{{Name: "CLib"}}},
			{Name: "CLib", Type: "regular", Path: "Sources/CLib", Sources: []string{"Sources/CLib/lib.c"}},
		},
	}
	value.writeExternal(map[string]string{
		"usr/include/stdio.h":  "int printf(const char *, ...);\n",
		"usr/include/stdint.h": "typedef int int32_t;\n",
	}, value.sdkRoot)
	return value
}

func (f *fixture) writeExternal(files map[string]string, root string) {
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

func (f *fixture) materialize() {
	f.t.Helper()
	for relative, payload := range f.files {
		full := filepath.Join(f.root, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
			f.t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(payload), 0o600); err != nil {
			f.t.Fatal(err)
		}
	}
	store, err := closureexec.NewCaptureStore(filepath.Join(f.base, "store"))
	if err != nil {
		f.t.Fatal(err)
	}
	tools := swiftpmsource.Toolchain{Swift: tool("swift", '1'), SwiftPM: tool("swiftpm", '2'), PackageDescription: tool("package-description", '3'), Git: tool("git", '4')}
	tools.Recheck = rechecker
	f.source = swiftpmsource.Config{
		Store: store, Policy: artifactpolicy.NewService(), Evaluator: &fakeEvaluator{manifest: f.manifest, byIdentity: f.manifests},
		Broker: fakeBroker{}, MirrorVerifier: fakeMirrorVerifier{}, OfflineRunner: fakeOfflineRunner{}, Toolchain: tools,
		Destination: swiftpmsource.Destination{Platform: darwinPlatform(), Markers: map[string]string{"platform": "macos", "configuration": "release", "architecture": "arm64"}},
		CausalHead:  "sha256:" + strings.Repeat("0", 64),
	}
	f.interop = Config{
		Clang:    tool("clang", '5'),
		ClangCXX: tool("clang++", '6'),
		SDK: ExternalComponent{
			Role: "macos-sdk", ExecutableRelativePath: "usr/bin/xcrun", PlatformABI: "darwin-arm64", PolicySelector: "apple-sdk-v1",
			VersionOutput: "MacOSX 26.5", Fingerprint: id('7'), SDKFactsDigest: id('8'), Roots: []string{filepath.Join(f.sdkRoot, "usr", "include")},
			Modules: []string{"Foundation"}, Frameworks: []string{"Foundation"}, Libraries: []string{"c", "System"},
		},
		Profile:   darwinProfile(),
		Assurance: closureexec.AssurancePortable,
		Recheck:   rechecker,
	}
	if f.materializeHook != nil {
		f.materializeHook()
	}
}

func (f *fixture) capture() (*swiftpmsource.Capture, error) {
	f.t.Helper()
	return swiftpmsource.CaptureAndClose(f.t.Context(), f.source, swiftpmsource.Request{Root: f.root, Product: "cli", Resolved: []byte(f.files["Package.resolved"])})
}

func (f *fixture) close() (*Result, error) {
	f.t.Helper()
	f.materialize()
	capture, err := f.capture()
	if err != nil {
		return nil, err
	}
	return Close(f.t.Context(), f.interop, capture)
}

func (f *fixture) mustClose() *Result {
	f.t.Helper()
	result, err := f.close()
	if err != nil {
		f.t.Fatalf("interop closure failed: %v", err)
	}
	return result
}

func darwinPlatform() closuregraph.TargetPlatformPayload {
	return closuregraph.TargetPlatformPayload{OS: "darwin", Architecture: "arm64", ABI: "darwin", Libc: "libSystem", MinimumRuntime: "macos-14", SDKID: "macos-sdk-v1", TargetTriple: "arm64-apple-macosx14.0", Runtime: "swift-6", LanguageModes: map[string]string{"swift": "6"}, Tuning: map[string]string{}}
}

func darwinProfile() PlatformProfile {
	return PlatformProfile{ID: "apple-swift-6.3.2-macos-26.5-v1", TargetTriples: []string{"arm64-apple-macosx14.0"}, CxxInterop: true, ObjectiveCRuntime: "objc4", CxxStandardModes: []string{"c++17"}}
}

func rechecker(_ context.Context, expected swiftpmsource.ToolIdentity) (closureexec.ToolchainIdentity, error) {
	return closureexec.ToolchainIdentity{Fingerprint: expected.Fingerprint, ExecutableSHA256: expected.ExecutableSHA256}, nil
}

var hexAlphabet = "0123456789abcdef"

func tool(role string, value byte) swiftpmsource.ToolIdentity {
	return swiftpmsource.ToolIdentity{Role: role, ExecutableRelativePath: "bin/" + strings.ReplaceAll(role, "+", "x"), VersionOutput: role + " 6.3.2 (Apple Swift version 6.3.2)", PlatformABI: "darwin-arm64", PolicySelector: "swift-toolchain-v1", Fingerprint: id(value), ExecutableSHA256: id(hexAlphabet[(strings.IndexByte(hexAlphabet, value)+8)%16])}
}

func id(value byte) closuregraph.ID {
	return closuregraph.ID("sha256:" + strings.Repeat(string(value), 64))
}

// fakeEvaluator answers one manifest per package identity so a fixture can
// declare an in-root `.package(path:)` dependency and exercise the
// cross-package edges the adapter really carries.
type fakeEvaluator struct {
	manifest   swiftpmsource.Manifest
	byIdentity map[string]swiftpmsource.Manifest
}

func (e *fakeEvaluator) Evaluate(_ context.Context, root string, permit swiftpmsource.ManifestPermit) (swiftpmsource.ManifestResult, error) {
	if permit.Network != "none" || !filepath.IsAbs(root) {
		return swiftpmsource.ManifestResult{}, errors.New("open manifest permit")
	}
	prebuiltsDisabled := false
	for _, argument := range permit.Argv {
		if argument == "--disable-experimental-prebuilts" {
			prebuiltsDisabled = true
		}
	}
	if !prebuiltsDisabled {
		return swiftpmsource.ManifestResult{}, errors.New("P05: SwiftPM permit omits --disable-experimental-prebuilts")
	}
	receipt, _ := closuregraph.DomainID("fake-swiftpm-manifest-receipt-v1", map[string]any{"package": permit.PackageIdentity, "permit": string(permit.ID)})
	manifest := e.manifest
	if selected, exists := e.byIdentity[permit.PackageIdentity]; exists {
		manifest = selected
	}
	return swiftpmsource.ManifestResult{Manifest: manifest, ReceiptID: receipt}, nil
}

type fakeBroker struct{}

func (fakeBroker) Acquire(context.Context, swiftpmsource.Pin) (swiftpmsource.Snapshot, error) {
	return swiftpmsource.Snapshot{}, errors.New("fixture has no source-control pins")
}

type fakeMirrorVerifier struct{}

func (fakeMirrorVerifier) Verify(context.Context, string, swiftpmsource.Pin, swiftpmsource.Snapshot) (swiftpmsource.GitVerificationEvidence, error) {
	return swiftpmsource.GitVerificationEvidence{}, errors.New("fixture has no mirrors")
}

type fakeOfflineRunner struct{}

func (fakeOfflineRunner) Replay(_ context.Context, capture *swiftpmsource.Capture) (swiftpmsource.OfflineMetadataResult, error) {
	identities := []string{}
	for _, pin := range capture.Lock.Pins {
		identities = append(identities, pin.Identity)
	}
	sort.Strings(identities)
	receipt, _ := closuregraph.DomainID("swiftpm-test-offline-receipt-v1", map[string]any{"identities": anyStrings(identities)})
	return swiftpmsource.OfflineMetadataResult{ReceiptID: receipt, PackageIdentities: identities}, nil
}

type fakeReadSets struct {
	reads    map[string][]ObservedRead
	observed bool
	calls    []string
}

func (r *fakeReadSets) ObserveReads(_ context.Context, request ReadSetRequest) (ReadSetResult, error) {
	r.calls = append(r.calls, request.Package+":"+request.Target)
	if !r.observed {
		return ReadSetResult{Observed: false}, nil
	}
	receipt, _ := closuregraph.DomainID("fake-read-set-receipt-v1", map[string]any{"target": request.Package + ":" + request.Target})
	return ReadSetResult{Observed: true, Reads: r.reads[request.Package+":"+request.Target], ReceiptID: receipt}, nil
}

func requireCode(t *testing.T, err error, want Code) {
	t.Helper()
	if ErrorCode(err) != want {
		t.Fatalf("error code = %q, want %q (err=%v)", ErrorCode(err), want, err)
	}
}

func (f *fixture) addFiles(files map[string]string) {
	for relative, payload := range files {
		f.files[relative] = payload
	}
}

func (f *fixture) target(name string) *swiftpmsource.Target {
	f.t.Helper()
	for index := range f.manifest.Targets {
		if f.manifest.Targets[index].Name == name {
			return &f.manifest.Targets[index]
		}
	}
	f.t.Fatalf("fixture has no target %s", name)
	return nil
}

func (f *fixture) addTarget(target swiftpmsource.Target) {
	f.manifest.Targets = append(f.manifest.Targets, target)
}

func linuxProfile() PlatformProfile {
	return PlatformProfile{ID: "swift-6.3.2-linux-x86_64-v1", TargetTriples: []string{"x86_64-unknown-linux-gnu"}}
}

func linuxPlatform() closuregraph.TargetPlatformPayload {
	return closuregraph.TargetPlatformPayload{OS: "linux", Architecture: "x86_64", ABI: "gnu", Libc: "glibc", MinimumRuntime: "glibc-2.31", SDKID: "swift-linux-sdk-v1", TargetTriple: "x86_64-unknown-linux-gnu", Runtime: "swift-6", LanguageModes: map[string]string{"swift": "6"}, Tuning: map[string]string{}}
}

// useLinuxDestination rebinds the fixture to the Linux destination profile.
// Only the exact selection changes; the admitted source closure does not.
func (f *fixture) useLinuxDestination() {
	f.materializeHook = func() {
		f.source.Destination = swiftpmsource.Destination{Platform: linuxPlatform(), Markers: map[string]string{"platform": "linux", "configuration": "release", "architecture": "x86_64"}}
		f.interop.Profile = linuxProfile()
		f.interop.SDK.PlatformABI = "linux-x86_64"
	}
}

func swiftpmCondition(expression string) *closuregraph.Condition {
	return &closuregraph.Condition{EvaluatorID: swiftpmsource.ConditionEvaluatorID, Expression: expression}
}
