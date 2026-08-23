//go:build darwin || linux

package swiftpmsource

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
)

// TestRealSwiftPMManifestVector keeps one real tool vector beside the semantic
// fakes. The adapter tests own policy; this smoke only proves the pinned argv
// shape still reaches SwiftPM and produces dump-package JSON.
func TestRealSwiftPMManifestVector(t *testing.T) {
	swift, err := exec.LookPath("swift")
	if err != nil {
		t.Skip("swift toolchain is not installed")
	}
	root := t.TempDir()
	if err = os.MkdirAll(filepath.Join(root, "Sources", "Fixture"), 0o700); err != nil {
		t.Fatal(err)
	}
	manifest := `// swift-tools-version: 5.9
import PackageDescription
let package = Package(name: "Fixture", products: [.executable(name: "fixture", targets: ["Fixture"])], targets: [.executableTarget(name: "Fixture")])
`
	if err = os.WriteFile(filepath.Join(root, "Package.swift"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "Sources", "Fixture", "main.swift"), []byte("print(1)\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.CommandContext(t.Context(), swift, manifestArgv()...) // #nosec G204 -- resolved test tool and production permit argv.
	command.Dir = root
	command.Env = []string{"HOME=" + filepath.Join(root, "empty-home"), "PATH=" + filepath.Dir(swift), "TZ=UTC"}
	output, err := command.Output()
	if err != nil {
		t.Fatalf("real SwiftPM dump-package: %v", err)
	}
	if !bytes.Contains(output, []byte(`"name" : "Fixture"`)) && !bytes.Contains(output, []byte(`"name":"Fixture"`)) {
		t.Fatalf("unexpected dump-package output: %s", output)
	}
}

func TestRealSwiftPMManifestRunsThroughProductionManagerAndExecutor(t *testing.T) {
	swift, err := exec.LookPath("swift")
	if err != nil {
		t.Skip("swift toolchain is not installed")
	}
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
	root := filepath.Join(base, "package")
	if err = os.MkdirAll(filepath.Join(root, "Sources", "Fixture"), 0o700); err != nil {
		t.Fatal(err)
	}
	dependency := filepath.Join(base, "dependency")
	if err = os.MkdirAll(filepath.Join(dependency, "Sources", "ALib"), 0o700); err != nil {
		t.Fatal(err)
	}
	dependencyManifest := "// swift-tools-version: 5.9\nimport PackageDescription\nlet package = Package(name: \"Dependency\", products: [.library(name: \"ALib\", targets: [\"ALib\"])], targets: [.target(name: \"ALib\")])\n"
	if err = os.WriteFile(filepath.Join(dependency, "Package.swift"), []byte(dependencyManifest), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dependency, "Sources", "ALib", "value.swift"), []byte("public let value = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(filepath.Join(dependency, "Sources", "ALib", "value.swift"), 0o700); err != nil {
		t.Fatal(err)
	}
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not installed")
	}
	gitRun := func(args ...string) string {
		command := exec.CommandContext(t.Context(), git, args...) // #nosec G204 -- fixed Git and test-owned arguments.
		command.Dir = dependency
		command.Env = []string{"GIT_CONFIG_NOSYSTEM=1", "HOME=" + filepath.Join(base, "git-home"), "LANG=C", "LC_ALL=C", "PATH=" + filepath.Dir(git), "TZ=UTC"}
		output, runErr := command.CombinedOutput()
		if runErr != nil {
			t.Fatalf("git %v: %v: %s", args, runErr, output)
		}
		return strings.TrimSpace(string(output))
	}
	gitRun("init", "-q")
	gitRun("add", "Package.swift", "Sources/ALib/value.swift")
	gitRun("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-q", "-m", "dependency")
	gitRun("tag", "1.0.0")
	revision := gitRun("rev-parse", "HEAD")
	origin := "file://" + dependency
	manifest := "// swift-tools-version: 5.9\nimport PackageDescription\nlet package = Package(name: \"Fixture\", products: [.executable(name: \"fixture\", targets: [\"Fixture\"])], dependencies: [.package(url: \"" + origin + "\", exact: \"1.0.0\")], targets: [.executableTarget(name: \"Fixture\", dependencies: [.product(name: \"ALib\", package: \"dependency\")])])\n"
	resolved := []byte(`{"version":3,"pins":[{"identity":"dependency","kind":"localSourceControl","location":"` + dependency + `","state":{"revision":"` + revision + `","version":"1.0.0"}}]}`)
	for name, payload := range map[string][]byte{
		"Package.swift":              []byte(manifest),
		"Package.resolved":           resolved,
		"Sources/Fixture/main.swift": []byte("print(1)\n"),
	} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err = os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, payload, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	executionRoot := filepath.Join(base, "execution")
	outputRoot := filepath.Join(executionRoot, "output")
	if err = os.MkdirAll(filepath.Join(executionRoot, "bin"), 0o700); err != nil {
		t.Fatal(err)
	}
	shim := []byte("#!/bin/sh\nexec \"" + swift + "\" \"$@\"\n")
	stagedSwift := filepath.Join(executionRoot, "bin", "swift-wrapper")
	if err = os.WriteFile(stagedSwift, shim, 0o500); err != nil {
		t.Fatal(err)
	}
	git, _, gitToolRoot, gitTool, gitRecheck := exactGitTestTool(t, git)
	stagedGit := filepath.Join(executionRoot, "bin", "git")
	if err = os.Symlink(git, stagedGit); err != nil {
		t.Fatal(err)
	}
	if len(gitTool.ProcessFamily) > 0 {
		helper := filepath.Join(gitToolRoot, filepath.FromSlash(gitTool.ProcessFamily[0].ExecutableRelativePath))
		if err = os.Symlink(helper, filepath.Join(executionRoot, "bin", "git-upload-pack")); err != nil {
			t.Fatal(err)
		}
	}
	digest := sha256.Sum256(shim)
	executableID := closuregraph.ID("sha256:" + hex.EncodeToString(digest[:]))
	runner, err := closureexec.NewManagerProcessRunner(executionRoot, outputRoot)
	if err != nil {
		t.Fatal(err)
	}
	starts := 0
	var launches []closureexec.ProcessLaunch
	runner.ProcessStartObserver = func(closureexec.DerivationPermit) { starts++ }
	runner.ProcessLaunchObserver = func(value closureexec.ProcessLaunch) { launches = append(launches, value) }
	executor, err := closureexec.NewAssuredExecutor(closureexec.DefaultAssuranceConfig(), runner, nil, "swiftpm-real-fixture")
	if err != nil {
		t.Fatal(err)
	}
	store, err := closureexec.NewCaptureStore(filepath.Join(base, "store"))
	if err != nil {
		t.Fatal(err)
	}
	version := "Apple Swift version 6.3.2"
	tool := ToolIdentity{Role: "swiftpm", ExecutableRelativePath: "bin/swift-wrapper", VersionOutput: version, PlatformABI: "darwin-arm64", PolicySelector: "swift-toolchain-v1", Fingerprint: executableID, ExecutableSHA256: executableID}
	gitTool.PlatformABI = "darwin-arm64"
	runtime := &ExecutorSwiftPM{Executor: executor, ExecutionRoot: executionRoot, OutputRoot: outputRoot, Tool: tool, AllowedProcesses: []string{"bin/git", "bin/git-upload-pack", "bin/swift-wrapper"}, Recheck: func(_ context.Context, expected ToolIdentity) (closureexec.ToolchainIdentity, error) {
		payload, readErr := os.ReadFile(filepath.Join(executionRoot, filepath.FromSlash(expected.ExecutableRelativePath)))
		if readErr != nil {
			return closureexec.ToolchainIdentity{}, readErr
		}
		observed := sha256.Sum256(payload)
		observedID := closuregraph.ID("sha256:" + hex.EncodeToString(observed[:]))
		return closureexec.ToolchainIdentity{Fingerprint: observedID, ExecutableSHA256: observedID}, nil
	}}
	tools := Toolchain{Swift: tool, SwiftPM: tool, PackageDescription: tool, Git: gitTool, Recheck: func(ctx context.Context, expected ToolIdentity) (closureexec.ToolchainIdentity, error) {
		if expected.Role == "git" {
			return gitRecheck(ctx, expected)
		}
		return runtime.Recheck(ctx, expected)
	}}
	tools.Swift.Role, tools.PackageDescription.Role = "swift", "package-description"
	gitStarts := 0
	var gitLaunches []closureexec.ProcessLaunch
	observeGitLaunch := func(launch closureexec.ProcessLaunch) { gitLaunches = append(gitLaunches, launch) }
	gitBroker := &GitBroker{WorkRoot: filepath.Join(executionRoot, "broker"), GitExecutable: git, ProcessStartObserver: func([]string) { gitStarts++ }, ProcessLaunchObserver: observeGitLaunch}
	verifier := &GitMirrorVerifier{GitExecutable: git, EnvironmentRoot: filepath.Join(executionRoot, "verify"), ProcessStartObserver: func([]string) { gitStarts++ }, ProcessLaunchObserver: observeGitLaunch}
	config := Config{Store: store, Policy: artifactpolicy.NewService(), Evaluator: runtime, Broker: gitBroker, MirrorVerifier: verifier, OfflineRunner: runtime, Toolchain: tools, GitExecutionRoot: executionRoot, GitToolRoot: gitToolRoot, Destination: Destination{Platform: closuregraph.TargetPlatformPayload{OS: "darwin", Architecture: "arm64", ABI: "darwin", Libc: "libSystem", MinimumRuntime: "macos-14", SDKID: "macos-sdk-v1", TargetTriple: "arm64-apple-macosx14.0", Runtime: "swift-6", LanguageModes: map[string]string{"swift": "6"}, Tuning: map[string]string{}}, Markers: map[string]string{"platform": "macos", "configuration": "release", "architecture": "arm64"}}, CausalHead: "sha256:" + strings.Repeat("0", 64)}
	manager, err := NewManager(config)
	if err != nil {
		t.Fatal(err)
	}
	capture, err := manager.CaptureAndClose(t.Context(), Request{Root: root, Product: "fixture", Resolved: resolved})
	if err != nil {
		t.Fatal(err)
	}
	if starts != 2 || capture.SelectionProduct() != "fixture" || len(capture.Packages) != 2 || len(capture.Mirrors) != 1 {
		t.Fatalf("production path evidence: starts=%d product=%q packages=%d", starts, capture.SelectionProduct(), len(capture.Packages))
	}
	if gitStarts == 0 || capture.Mirrors[0].ArtifactManifestID == capture.Packages[1].ArtifactManifestID {
		t.Fatalf("Git acquisition evidence missing or mirror reused checkout manifest: starts=%d mirror=%s package=%s", gitStarts, capture.Mirrors[0].ArtifactManifestID, capture.Packages[1].ArtifactManifestID)
	}
	if len(gitLaunches) == 0 {
		t.Fatal("production Git process launches were not instrumented")
	}
	for _, launch := range gitLaunches {
		if launch.Executable != git {
			t.Fatalf("Git work launched %q instead of exact C0 executable %q", launch.Executable, git)
		}
	}
	mirror := capture.Mirrors[0]
	if !mirror.CommitEvidenceIntakeReceiptID.Valid() || !mirror.CommitEvidenceArtifactManifestID.Valid() || !mirror.MirrorDerivationPermitID.Valid() || !mirror.MirrorDerivationReceiptID.Valid() || mirror.AuthorizedOutputPath == "" || mirror.authorization == nil {
		t.Fatalf("mirror causal authorization chain is incomplete: %+v", mirror)
	}
	c1 := capture.C1.Payload.(closuregraph.C1ResolvePayload)
	c3 := capture.C3.Payload.(closuregraph.C3AdmitPayload)
	if !containsID(c1.JournalEntryIDs, mirror.MirrorDerivationPermitID) || !containsID(c1.JournalEntryIDs, mirror.MirrorDerivationReceiptID) || !containsID(c3.IntakeReceiptIDs, mirror.CommitEvidenceIntakeReceiptID) || !containsID(c3.ArtifactManifestIDs, mirror.CommitEvidenceArtifactManifestID) || !containsID(c3.DerivationReceiptIDs, mirror.MirrorDerivationReceiptID) {
		t.Fatal("C1/C3 do not resolve the acquisition-to-mirror causal chain")
	}
	if got, want := launches[0].Argv, manifestArgv(); !reflect.DeepEqual(got, want) || contains(got, "--manifest-path") {
		t.Fatalf("launched manifest argv = %v, want exact permit argv %v", got, want)
	}
	if err = capture.ReplayOffline(t.Context()); err != nil {
		t.Fatal(err)
	}
	if starts != 5 || len(launches) != 5 {
		t.Fatalf("production replay starts=%d launches=%d, want manifest replay plus offline metadata", starts, len(launches))
	}
	if got, want := launches[4].Argv, offlineMetadataArgv(); !reflect.DeepEqual(got, want) {
		t.Fatalf("launched offline metadata argv = %v, want exact permit argv %v", got, want)
	}
	gitStartsBeforeTamper := gitStarts
	capture.Mirrors[0].AuthorizedOutputPath = "substituted/mirror.git"
	if err = capture.ReplayOffline(t.Context()); ErrorCode(err) != CodeDerivationUnauthorized {
		t.Fatalf("tampered mirror authorization error = %v", err)
	}
	if gitStarts != gitStartsBeforeTamper {
		t.Fatalf("tampered mirror authorization started Git: before=%d after=%d", gitStartsBeforeTamper, gitStarts)
	}
}

func TestGitBrokerCapturesExactLocalRevisionTreeAndMirrorR01R07R08R09(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not installed")
	}
	git, executionRoot, gitToolRoot, gitTool, gitRecheck := exactGitTestTool(t, git)
	base := t.TempDir()
	repository := filepath.Join(base, "origin")
	if err = os.Mkdir(repository, 0o700); err != nil {
		t.Fatal(err)
	}
	runGit := func(args ...string) string {
		t.Helper()
		command := exec.CommandContext(t.Context(), git, args...) // #nosec G204 -- fixed test Git with test-owned arguments.
		command.Dir = repository
		command.Env = []string{"GIT_CONFIG_NOSYSTEM=1", "HOME=" + filepath.Join(base, "home"), "LANG=C", "LC_ALL=C", "PATH=" + filepath.Dir(git), "TZ=UTC"}
		output, runErr := command.CombinedOutput()
		if runErr != nil {
			t.Fatalf("git %v: %v: %s", args, runErr, output)
		}
		return strings.TrimSpace(string(output))
	}
	runGit("init", "-q")
	runGit("config", "user.name", "Fixture")
	runGit("config", "user.email", "fixture@example.invalid")
	if err = os.MkdirAll(filepath.Join(repository, "Sources", "Fixture"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(repository, "Package.swift"), []byte("// swift-tools-version: 5.9\nimport PackageDescription\nlet package = Package(name: \"Fixture\")\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(repository, "Sources", "Fixture", "main.swift"), []byte("print(1)\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit("add", "Package.swift", "Sources/Fixture/main.swift")
	runGit("commit", "-q", "-m", "fixture")
	runGit("tag", "1.0.0")
	runGit("tag", "v1.5.0")
	runGit("tag", "2.0.0")
	runGit("branch", "stable")
	revision := runGit("rev-parse", "HEAD")
	tree := runGit("show", "-s", "--format=%T", revision)
	starts := 0
	broker := &GitBroker{WorkRoot: filepath.Join(executionRoot, "broker"), GitExecutable: git, ProcessStartObserver: func([]string) { starts++ }}
	authority, err := newSharedGitAuthority(id('c'), Toolchain{Git: gitTool, Recheck: gitRecheck}, executionRoot, gitToolRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err = broker.bindGitAuthority(authority); err != nil {
		t.Fatal(err)
	}
	for requirement, want := range map[string][2]string{
		"exact:1.0.0":         {"1.0.0", ""},
		"range:1.0.0..<2.0.0": {"1.5.0", ""},
		"branch:stable":       {"", "stable"},
	} {
		resolved, resolveErr := broker.ResolvePin(t.Context(), ManifestDependency{Identity: "fixture", Kind: SourceLocal, Location: repository, Requirement: requirement})
		if resolveErr != nil || resolved.Revision != revision || resolved.Version != want[0] || resolved.Branch != want[1] {
			t.Fatalf("resolve %s = %+v, %v", requirement, resolved, resolveErr)
		}
	}
	resolvedRevision, err := broker.ResolvePin(t.Context(), ManifestDependency{Identity: "fixture", Kind: SourceLocal, Location: repository, Requirement: "revision:" + revision})
	if err != nil || resolvedRevision.Revision != revision {
		t.Fatalf("revision resolve = %+v, %v", resolvedRevision, err)
	}
	pin := Pin{Identity: "fixture", Kind: SourceLocal, RawLocation: repository, CanonicalLocation: repository, Revision: revision, Version: "1.0.0"}
	snapshot, err := broker.Acquire(t.Context(), pin)
	if err != nil {
		t.Fatal(err)
	}
	if starts == 0 || snapshot.Revision != revision || snapshot.GitTree != tree || snapshot.Kind != SourceLocal || !snapshot.BrokerReceiptID.Valid() {
		t.Fatalf("broker evidence = %+v starts=%d", snapshot, starts)
	}
	if _, err = os.Stat(filepath.Join(snapshot.Root, ".git")); !os.IsNotExist(err) {
		t.Fatalf("snapshot retained Git administration state: %v", err)
	}
	verifier := &GitMirrorVerifier{GitExecutable: git, EnvironmentRoot: filepath.Join(base, "verify")}
	if err = verifier.bindGitAuthority(authority); err != nil {
		t.Fatal(err)
	}
	if _, err = verifier.Verify(t.Context(), snapshot.MirrorRoot, pin, snapshot); err != nil {
		t.Fatal(err)
	}
	drifted := snapshot
	drifted.GitTree = strings.Repeat("0", len(tree))
	if _, err = verifier.Verify(t.Context(), snapshot.MirrorRoot, pin, drifted); err == nil {
		t.Fatal("wrong mirror tree was accepted")
	}
}

func TestGitC0ExecutableMismatchRejectsBeforeAnyProcessStart(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not installed")
	}
	git, executionRoot, gitToolRoot, tool, recheck := exactGitTestTool(t, git)
	other := filepath.Join(t.TempDir(), "other-git")
	payload, err := os.ReadFile(git)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(other, payload, 0o500); err != nil {
		t.Fatal(err)
	}
	starts := 0
	broker := &GitBroker{WorkRoot: t.TempDir(), GitExecutable: other, ProcessStartObserver: func([]string) { starts++ }}
	authority, err := newSharedGitAuthority(id('c'), Toolchain{Git: tool, Recheck: recheck}, executionRoot, gitToolRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err = broker.bindGitAuthority(authority); ErrorCode(err) != CodeDerivationUnauthorized {
		t.Fatalf("C0 executable mismatch error = %v", err)
	}
	if starts != 0 {
		t.Fatalf("Git starts=%d, want zero before exact C0 binding", starts)
	}
}

func TestGitC0RelativeAndSymlinkEscapeRejectBeforeAnyProcessStart(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "git")
	if err := os.WriteFile(outside, []byte("#!/bin/sh\nexit 0\n"), 0o500); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "git-link")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	tool := ToolIdentity{Role: "git", ExecutableRelativePath: "git-link", VersionOutput: "git", PlatformABI: "test-host", PolicySelector: "swift-toolchain-v1", Fingerprint: id('a'), ExecutableSHA256: id('a')}
	starts := 0
	toolchain := Toolchain{Git: tool, Recheck: func(context.Context, ToolIdentity) (closureexec.ToolchainIdentity, error) {
		starts++
		return closureexec.ToolchainIdentity{}, nil
	}}
	if _, err := newSharedGitAuthority(id('c'), toolchain, root, root); ErrorCode(err) != CodeDerivationUnauthorized {
		t.Fatalf("symlink escape error = %v", err)
	}
	tool.ExecutableRelativePath = filepath.ToSlash(filepath.Join("..", filepath.Base(outside)))
	toolchain.Git = tool
	if _, err := newSharedGitAuthority(id('c'), toolchain, root, root); ErrorCode(err) != CodeDerivationUnauthorized {
		t.Fatalf("relative escape error = %v", err)
	}
	if starts != 0 {
		t.Fatalf("escaping C0 Git path started %d rechecks/processes", starts)
	}
}

func TestGitC0ProcessFamilyDriftRejectsBeforeAnyProcessStart(t *testing.T) {
	toolRoot := t.TempDir()
	executionRoot := t.TempDir()
	primaryPayload := []byte("#!/bin/sh\nexit 99\n")
	helperPayload := []byte("#!/bin/sh\nexit 98\n")
	if err := os.WriteFile(filepath.Join(toolRoot, "git"), primaryPayload, 0o500); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(toolRoot, "git-upload-pack"), helperPayload, 0o500); err != nil {
		t.Fatal(err)
	}
	primaryDigest, helperDigest := sha256.Sum256(primaryPayload), sha256.Sum256(helperPayload)
	primaryID := closuregraph.ID("sha256:" + hex.EncodeToString(primaryDigest[:]))
	helperID := closuregraph.ID("sha256:" + hex.EncodeToString(helperDigest[:]))
	tool := ToolIdentity{Role: "git", ExecutableRelativePath: "git", VersionOutput: "git", PlatformABI: "test-host", PolicySelector: "swift-toolchain-v1", ExecutableSHA256: primaryID, ProcessFamily: []ToolProcessIdentity{{ExecutableRelativePath: "git-upload-pack", ExecutableSHA256: helperID}}}
	var err error
	tool.Fingerprint, err = toolProcessFamilyFingerprint(tool)
	if err != nil {
		t.Fatal(err)
	}
	recheck := func(context.Context, ToolIdentity) (closureexec.ToolchainIdentity, error) {
		return closureexec.ToolchainIdentity{Fingerprint: tool.Fingerprint, ExecutableSHA256: primaryID}, nil
	}
	authority, err := newSharedGitAuthority(id('c'), Toolchain{Git: tool, Recheck: recheck}, executionRoot, toolRoot)
	if err != nil {
		t.Fatal(err)
	}
	starts := 0
	broker := &GitBroker{WorkRoot: filepath.Join(executionRoot, "broker"), GitExecutable: filepath.Join(toolRoot, "git"), ProcessStartObserver: func([]string) { starts++ }}
	if err = broker.bindGitAuthority(authority); err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(filepath.Join(toolRoot, "git-upload-pack"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(toolRoot, "git-upload-pack"), []byte("#!/bin/sh\nexit 97\n"), 0o500); err != nil {
		t.Fatal(err)
	}
	_, err = broker.ResolvePin(t.Context(), ManifestDependency{Identity: "fixture", Kind: SourceRemote, Location: "https://example.invalid/fixture.git", Requirement: "branch:main"})
	if err == nil {
		t.Fatal("drifted C0 Git process family was accepted")
	}
	if starts != 0 {
		t.Fatalf("drifted C0 Git process family started %d processes", starts)
	}
}

func TestBrokeredResolverGeneratesTransitiveLockBeforeMainCaptureR02(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not installed")
	}
	git, gitExecutionRoot, gitToolRoot, gitTool, gitRecheck := exactGitTestTool(t, git)
	fixture := newFixture(t)
	if err = os.Remove(filepath.Join(fixture.root, "Package.resolved")); err != nil {
		t.Fatal(err)
	}
	dependencyRoot := filepath.Join(filepath.Dir(fixture.root), "a")
	transitiveRoot := filepath.Join(filepath.Dir(fixture.root), "b")
	command := func(repository string, args ...string) {
		t.Helper()
		cmd := exec.CommandContext(t.Context(), git, args...) // #nosec G204 -- fixed test Git and test-owned arguments.
		cmd.Dir = repository
		cmd.Env = []string{"GIT_CONFIG_NOSYSTEM=1", "HOME=" + filepath.Join(filepath.Dir(fixture.root), "git-home"), "LANG=C", "LC_ALL=C", "PATH=" + filepath.Dir(git), "TZ=UTC"}
		if output, runErr := cmd.CombinedOutput(); runErr != nil {
			t.Fatalf("git %v: %v: %s", args, runErr, output)
		}
	}
	for _, repository := range []string{dependencyRoot, transitiveRoot} {
		command(repository, "init", "-q")
		command(repository, "add", "Package.swift", "Sources/App/main.swift")
		command(repository, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-q", "-m", "fixture")
		command(repository, "tag", "1.0.0")
	}
	rootManifest := fixture.evaluator.manifests["root"]
	rootManifest.Dependencies = []ManifestDependency{{Identity: "a", Kind: SourceLocal, Location: dependencyRoot, Requirement: "exact:1.0.0"}}
	rootManifest.Targets[0].Dependencies = []TargetDependency{{Package: "a", Product: "AProd"}}
	fixture.evaluator.manifests["root"] = rootManifest
	aManifest := fixture.evaluator.manifests["a"]
	aManifest.Dependencies = []ManifestDependency{{Identity: "b", Kind: SourceLocal, Location: transitiveRoot, Requirement: "exact:1.0.0"}}
	aManifest.Targets[0].Dependencies = []TargetDependency{{Package: "b", Product: "BProd"}}
	fixture.evaluator.manifests["a"] = aManifest
	broker := &GitBroker{WorkRoot: filepath.Join(gitExecutionRoot, "brokered-resolution"), GitExecutable: git}
	verifier := &GitMirrorVerifier{GitExecutable: git, EnvironmentRoot: filepath.Join(gitExecutionRoot, "verify")}
	fixture.config.Toolchain.Git = gitTool
	previousRecheck := fixture.config.Toolchain.Recheck
	fixture.config.Toolchain.Recheck = func(ctx context.Context, expected ToolIdentity) (closureexec.ToolchainIdentity, error) {
		if expected.Role == "git" {
			return gitRecheck(ctx, expected)
		}
		return previousRecheck(ctx, expected)
	}
	fixture.config.GitExecutionRoot = gitExecutionRoot
	fixture.config.GitToolRoot = gitToolRoot
	resolver := &BrokeredResolver{Store: fixture.config.Store, Policy: fixture.config.Policy, Evaluator: fixture.evaluator, Broker: broker, Toolchain: fixture.config.Toolchain, Destination: fixture.config.Destination, CausalHead: fixture.config.CausalHead, ProcessStartObserver: fixture.config.ProcessStartObserver}
	fixture.config.Broker = broker
	fixture.config.MirrorVerifier = verifier
	fixture.config.Resolver = resolver
	capture, err := CaptureAndClose(t.Context(), fixture.config, Request{Root: fixture.root, Product: "cli"})
	if err != nil {
		t.Fatal(err)
	}
	if len(capture.Lock.Pins) != 2 || capture.Lock.Pins[0].Identity != "a" || capture.Lock.Pins[1].Identity != "b" || capture.Lock.Pins[0].Version != "1.0.0" || !capture.ResolutionReceiptID.Valid() {
		t.Fatalf("generated lock evidence = %+v receipt=%s", capture.Lock, capture.ResolutionReceiptID)
	}
	resolver.mu.Lock()
	issued := resolver.issued[capture.ResolutionReceiptID]
	resolver.mu.Unlock()
	c0ID, _ := capture.C0.ID()
	permit := resolutionPermit(c0ID, capture.rootInput, fixture.config.Toolchain.Git, fixture.config.Destination)
	if err = resolver.VerifyResult(permit, issued); err != nil {
		t.Fatalf("generated-lock receipt does not resolve to exact permit: %v", err)
	}
	if len(issued.GitPermitIDs) == 0 || len(issued.GitPermitIDs) != len(issued.GitReceiptIDs) {
		t.Fatalf("generated-lock Git journal is incomplete: permits=%d receipts=%d", len(issued.GitPermitIDs), len(issued.GitReceiptIDs))
	}
	c1 := capture.C1.Payload.(closuregraph.C1ResolvePayload)
	c3 := capture.C3.Payload.(closuregraph.C3AdmitPayload)
	for index := range issued.GitPermitIDs {
		if !containsID(c1.JournalEntryIDs, issued.GitPermitIDs[index]) || !containsID(c1.JournalEntryIDs, issued.GitReceiptIDs[index]) || !containsID(c3.DerivationReceiptIDs, issued.GitReceiptIDs[index]) {
			t.Fatalf("Git permit/receipt pair is absent from C1/C3: %s %s", issued.GitPermitIDs[index], issued.GitReceiptIDs[index])
		}
	}
	if got, want := fixture.evaluator.starts, []string{"root", "a", "b", "a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("resolver/main manifest starts = %v, want %v", got, want)
	}
}

func exactGitTestTool(t *testing.T, executable string) (string, string, string, ToolIdentity, func(context.Context, ToolIdentity) (closureexec.ToolchainIdentity, error)) {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		t.Fatal(err)
	}
	toolRoot := filepath.Dir(filepath.Dir(resolved))
	relative, err := filepath.Rel(toolRoot, resolved)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(resolved)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(payload)
	id := closuregraph.ID("sha256:" + hex.EncodeToString(digest[:]))
	tool := ToolIdentity{Role: "git", ExecutableRelativePath: filepath.ToSlash(relative), VersionOutput: "git", PlatformABI: "test-host", PolicySelector: "swift-toolchain-v1", Fingerprint: id, ExecutableSHA256: id}
	if helper := filepath.Join(filepath.Dir(resolved), "git-upload-pack"); helper != resolved {
		if helperPayload, helperErr := os.ReadFile(helper); helperErr == nil {
			helperDigest := sha256.Sum256(helperPayload)
			helperID := closuregraph.ID("sha256:" + hex.EncodeToString(helperDigest[:]))
			helperRelative, _ := filepath.Rel(toolRoot, helper)
			tool.ProcessFamily = []ToolProcessIdentity{{ExecutableRelativePath: filepath.ToSlash(helperRelative), ExecutableSHA256: helperID}}
		}
	}
	if len(tool.ProcessFamily) > 0 {
		tool.Fingerprint, err = toolProcessFamilyFingerprint(tool)
		if err != nil {
			t.Fatal(err)
		}
	}
	recheck := func(_ context.Context, expected ToolIdentity) (closureexec.ToolchainIdentity, error) {
		current, readErr := os.ReadFile(filepath.Join(toolRoot, filepath.FromSlash(expected.ExecutableRelativePath)))
		if readErr != nil {
			return closureexec.ToolchainIdentity{}, readErr
		}
		sum := sha256.Sum256(current)
		observed := closuregraph.ID("sha256:" + hex.EncodeToString(sum[:]))
		if observed != expected.ExecutableSHA256 {
			return closureexec.ToolchainIdentity{Fingerprint: observed, ExecutableSHA256: observed}, nil
		}
		return closureexec.ToolchainIdentity{Fingerprint: expected.Fingerprint, ExecutableSHA256: observed}, nil
	}
	return resolved, t.TempDir(), toolRoot, tool, recheck
}
