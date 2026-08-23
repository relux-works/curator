package rustsource

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
)

func TestParseManifestAndLockRetainSelectionNeutralSuperset(t *testing.T) {
	manifest, err := ParseManifest("workspace/Cargo.toml", []byte(`[package]
name = "cli"
version = "0.1.0"
edition = "2024"

[features]
extra = ["dep:optional"]

[dependencies]
optional = { version = "1", optional = true, default-features = false, features = ["b", "a"] }

[target.'cfg(unix)'.dependencies]
unix-only = "2"

[target.'cfg(windows)'.dependencies]
windows-only = "3"

[[bin]]
name = "cli"
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Dependencies) != 3 || manifest.Dependencies[1].Target == "" || manifest.Dependencies[2].Target == "" {
		t.Fatalf("target declarations not retained: %#v", manifest.Dependencies)
	}
	if got := manifest.Dependencies[0].Features; !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("features = %#v", got)
	}
	lockPayload := []byte(`version = 4

[[package]]
name = "cli"
version = "0.1.0"
dependencies = ["optional 1.0.0 (registry+https://github.com/rust-lang/crates.io-index)", "unix-only 2.0.0 (registry+https://github.com/rust-lang/crates.io-index)", "windows-only 3.0.0 (registry+https://github.com/rust-lang/crates.io-index)"]

[[package]]
name = "optional"
version = "1.0.0"
source = "registry+https://github.com/rust-lang/crates.io-index"
checksum = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

[[package]]
name = "unix-only"
version = "2.0.0"
source = "registry+https://github.com/rust-lang/crates.io-index"
checksum = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

[[package]]
name = "windows-only"
version = "3.0.0"
source = "registry+https://github.com/rust-lang/crates.io-index"
checksum = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
`)
	lock, err := ParseLock(lockPayload)
	if err != nil {
		t.Fatal(err)
	}
	if len(lock.Packages) != 4 {
		t.Fatalf("lock superset = %d", len(lock.Packages))
	}
	graph, err := NewCaptureGraph(lock, []Manifest{manifest}, []string{"sha256:" + digest([]byte("manifest"))})
	if err != nil {
		t.Fatal(err)
	}
	if graph.ContainsSelectionFacts() {
		t.Fatal("selection facts contaminated capture")
	}
	graph2, err := NewCaptureGraph(lock, []Manifest{manifest}, []string{"sha256:" + digest([]byte("manifest"))})
	if err != nil {
		t.Fatal(err)
	}
	if graph.Identity != graph2.Identity {
		t.Fatal("feature selection changed capture identity")
	}
}

func TestParseManifestRejectsUnknownSecurityDeclaration(t *testing.T) {
	_, err := ParseManifest("Cargo.toml", []byte("[package]\nname='x'\nversion='1.0.0'\nmagic-runner='x'\n"))
	if ErrorCode(err) != CodeGraphIncomplete {
		t.Fatalf("code = %q, err=%v", ErrorCode(err), err)
	}
}

func TestParseManifestRejectsPackageProfileAsUntrustedConfig(t *testing.T) {
	_, err := ParseManifest("Cargo.toml", []byte("[package]\nname='x'\nversion='1.0.0'\n[profile.release]\nlto=true\n"))
	if ErrorCode(err) != CodeConfigUntrusted {
		t.Fatalf("code=%q err=%v", ErrorCode(err), err)
	}
}

func TestGitLockRequiresFullLowercaseCommit(t *testing.T) {
	base := func(source string) []byte {
		return []byte("version = 4\n[[package]]\nname='git_leaf'\nversion='0.1.0'\nsource='" + source + "'\n")
	}
	valid := "git+https://example.invalid/repo?branch=main#0123456789abcdef0123456789abcdef01234567"
	if _, err := ParseLock(base(valid)); err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"git+https://example.invalid/repo?branch=main", "git+https://example.invalid/repo#01234567", "git+https://example.invalid/repo#0123456789ABCDEF0123456789ABCDEF01234567"} {
		if _, err := ParseLock(base(source)); ErrorCode(err) != CodeGitIdentityInvalid {
			t.Fatalf("source=%s code=%q err=%v", source, ErrorCode(err), err)
		}
	}
}

func TestRegistryTransformUsesExactRootAndBasenameRules(t *testing.T) {
	key := PackageKey{Name: "edge", Version: "1.0.0", Source: "registry+https://example.invalid/index"}
	archive := crateBytes(t, key, map[string][]byte{".gitattributes": []byte("root"), ".gitignore": []byte("root"), ".cargo-ok": {}, "nested/.cargo-ok": {}, "nested/.gitignore": []byte("copy"), "Cargo.toml": []byte("[package]\nname='edge'\nversion='1.0.0'\n"), "src/lib.rs": []byte("pub fn x() {}\n")})
	checksum := digest(archive)
	index, _ := json.Marshal(map[string]any{"name": "edge", "vers": "1.0.0", "cksum": checksum, "deps": []any{}})
	result, err := deriveRegistryTransform(registryOrigin{Package: key, IndexRecord: index, Archive: archive, Checksum: checksum})
	if err != nil {
		t.Fatal(err)
	}
	dispositions := map[string]Disposition{}
	for _, entry := range result.Entries {
		dispositions[entry.OriginPath] = entry.Disposition
	}
	if dispositions[".gitignore"] != OmitReserved || dispositions["nested/.cargo-ok"] != OmitRegistryCargoOK || dispositions["nested/.gitignore"] != CopyIdentical {
		t.Fatalf("dispositions = %#v", dispositions)
	}
	var checksumObject map[string]any
	if err = json.Unmarshal(result.ChecksumBytes, &checksumObject); err != nil {
		t.Fatal(err)
	}
	files := checksumObject["files"].(map[string]any)
	if _, ok := files["nested/.gitignore"]; !ok {
		t.Fatal("nested .gitignore omitted")
	}
	if _, ok := files["nested/.cargo-ok"]; ok {
		t.Fatal("nested .cargo-ok copied")
	}
}

func TestGitTransformBranchesAndTamperVerification(t *testing.T) {
	manifest := []byte("[package]\nname = \"git_leaf\"\nversion = \"0.1.0\"\n")
	normalized := []byte("# normalized\n[package]\nname = \"git_leaf\"\nversion = \"0.1.0\"\n")
	leaf := func(path string, data []byte) OriginLeaf {
		return OriginLeaf{Path: path, Bytes: data, Size: int64(len(data)), SHA256: digest(data)}
	}
	origin := gitOrigin{Package: PackageKey{Name: "git_leaf", Version: "0.1.0", Source: "git+https://example.invalid/repo#0123456789012345678901234567890123456789"}, Commit: "0123456789012345678901234567890123456789", Tree: "tree", ManifestTracked: true, Leaves: []OriginLeaf{leaf(".gitignore", []byte("x")), leaf("Cargo.toml", manifest), leaf("src/lib.rs", []byte("pub fn x() {}\n")), leaf("target/tracked.txt", []byte("tracked")), leaf("outside.txt", []byte("outside"))}}
	derivation := gitDerivation{mode: ProjectionGitIndexNoInclude, selected: []string{".gitignore", "Cargo.toml", "src/lib.rs", "target/tracked.txt"}, normalizerID: NormalizerID, normalizerInputs: []string{"Cargo.toml", "src/lib.rs", "target/tracked.txt"}, normalizedManifest: normalized, receiptID: "issued-fixture", commit: "0123456789012345678901234567890123456789", tree: "tree", manifestTracked: true}
	seal, sealErr := gitDerivationSeal(derivation)
	if sealErr != nil {
		t.Fatal(sealErr)
	}
	derivation.seal = seal
	result, err := deriveGitTransform(origin, derivation)
	if err != nil {
		t.Fatal(err)
	}
	byOrigin := map[string]Disposition{}
	for _, entry := range result.Entries {
		byOrigin[entry.OriginPath] = entry.Disposition
	}
	if byOrigin["target/tracked.txt"] != CopyIdentical || byOrigin["outside.txt"] != OmitUnselected || byOrigin["Cargo.toml"] != ReplaceNormalizedManifest {
		t.Fatalf("dispositions = %#v", byOrigin)
	}
	root := t.TempDir()
	materializeVendor(t, root, []VendorPackage{result})
	if err = VerifyVendor(root, []VendorPackage{result}); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, result.Directory, "src", "lib.rs"), []byte("forged"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code := ErrorCode(VerifyVendor(root, []VendorPackage{result})); code != CodeGitIdentityInvalid {
		t.Fatalf("tamper code = %q", code)
	}
	forgedSelection := derivation
	forgedSelection.selected = []string{"Cargo.toml"}
	if _, transformErr := deriveGitTransform(origin, forgedSelection); ErrorCode(transformErr) != CodeGitIdentityInvalid {
		t.Fatalf("forged selection code=%q err=%v", ErrorCode(transformErr), transformErr)
	}
	forgedManifest := derivation
	forgedManifest.normalizedManifest = []byte("[package]\nname='forged'\n")
	if _, transformErr := deriveGitTransform(origin, forgedManifest); ErrorCode(transformErr) != CodeGitIdentityInvalid {
		t.Fatalf("forged manifest code=%q err=%v", ErrorCode(transformErr), transformErr)
	}
	derivation.mode = ProjectionFilesystemInclude
	origin.Include = []string{"Cargo.toml", "src/**", "target/**"}
	derivation.include = []string{"Cargo.toml", "src/**", "target/**"}
	derivation.selected = []string{"Cargo.toml", "src/lib.rs"}
	derivation.seal, err = gitDerivationSeal(derivation)
	if err != nil {
		t.Fatal(err)
	}
	result, err = deriveGitTransform(origin, derivation)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range result.Entries {
		if entry.OriginPath == "target/tracked.txt" && entry.Disposition != OmitUnselected {
			t.Fatalf("root target disposition = %q", entry.Disposition)
		}
	}
}

func TestGitTransformRejectsCallerWithoutSealedDerivation(t *testing.T) {
	manifest := []byte("[package]\nname='x'\nversion='1'\n")
	origin := gitOrigin{Package: PackageKey{Name: "x", Version: "1", Source: "git+https://example.invalid/repo#0123456789abcdef0123456789abcdef01234567"}, Commit: "0123456789abcdef0123456789abcdef01234567", Tree: "tree", ManifestTracked: true, Leaves: []OriginLeaf{{Path: "Cargo.toml", Bytes: manifest, Size: int64(len(manifest)), SHA256: digest(manifest)}}}
	if _, err := deriveGitTransform(origin, gitDerivation{}); ErrorCode(err) != CodeGitIdentityInvalid {
		t.Fatalf("code=%q err=%v", ErrorCode(err), err)
	}
}

func TestPinnedGitNormalizerMatchesCargo092SimpleManifest(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "src"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "Cargo.toml"), []byte("[package]\nname = \"git_leaf\"\nversion = \"0.1.0\"\nedition = \"2021\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "lib.rs"), []byte("pub fn x() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	payload, err := normalizeGitManifestV1(root, []string{"Cargo.toml", "src/lib.rs"})
	if err != nil {
		t.Fatal(err)
	}
	expected := cargoManifestPreamble + `
[package]
edition = "2021"
name = "git_leaf"
version = "0.1.0"
build = false
autolib = false
autobins = false
autoexamples = false
autotests = false
autobenches = false
readme = false

[lib]
name = "git_leaf"
path = "src/lib.rs"
`
	if string(payload) != expected {
		t.Fatalf("normalized manifest differs\nwant:\n%s\n got:\n%s", expected, payload)
	}
	if digest(payload) != "e9eafb4fa1a3b8328b3a85987990f810af4fe2e54d5786be34c680a5d24f35e3" {
		t.Fatalf("normalized digest = %s", digest(payload))
	}
}

func TestPinnedGitNormalizerResolvesWorkspaceAndExplicitBinary(t *testing.T) {
	root := t.TempDir()
	member := filepath.Join(root, "member")
	if err := os.MkdirAll(filepath.Join(member, "src"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "Cargo.toml"), []byte("[workspace]\nmembers=[\"member\"]\n[workspace.package]\nversion=\"0.2.0\"\nedition=\"2021\"\nlicense=\"MIT\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := "[package]\nname=\"probe\"\nversion.workspace=true\nedition.workspace=true\nlicense.workspace=true\n\n[[bin]]\nname=\"probe\"\npath=\"src/main.rs\"\n"
	if err := os.WriteFile(filepath.Join(member, "Cargo.toml"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(member, "src", "main.rs"), []byte("fn main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	payload, err := normalizeGitManifestV1(member, []string{"Cargo.toml", "src/main.rs"})
	if err != nil {
		t.Fatal(err)
	}
	expected := cargoManifestPreamble + `
[package]
edition = "2021"
name = "probe"
version = "0.2.0"
build = false
autolib = false
autobins = false
autoexamples = false
autotests = false
autobenches = false
readme = false
license = "MIT"
resolver = "1"

[[bin]]
name = "probe"
path = "src/main.rs"
`
	if string(payload) != expected {
		t.Fatalf("workspace normalized manifest differs\nwant:\n%s\n got:\n%s", expected, payload)
	}
}

func TestManagerRejectsCompiledGitBeforeOracleVendorOrPrivateCargoState(t *testing.T) {
	workspace := t.TempDir()
	manifestPath := filepath.Join(workspace, "Cargo.toml")
	lockPath := filepath.Join(workspace, "Cargo.lock")
	if err := os.WriteFile(manifestPath, []byte("[package]\nname='root'\nversion='0.1.0'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	commit := "0123456789abcdef0123456789abcdef01234567"
	lock := []byte("version = 4\n\n[[package]]\nname='root'\nversion='0.1.0'\n\n[[package]]\nname='bad'\nversion='1.0.0'\nsource='git+https://example.invalid/bad#" + commit + "'\n")
	if err := os.WriteFile(lockPath, lock, 0o600); err != nil {
		t.Fatal(err)
	}
	repository := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repository, "target"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "Cargo.toml"), []byte("[package]\nname='bad'\nversion='1.0.0'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "target", "renamed"), []byte{0, 'a', 's', 'm', 1, 0, 0, 0}, 0o600); err != nil {
		t.Fatal(err)
	}
	manager, err := NewManager(t.Context(), ManagerConfig{WorkRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	_, err = manager.Capture(t.Context(), RawCaptureRequest{Workspace: RawTree{Root: workspace}, Lock: RawFile{Path: lockPath}, Manifests: []RawManifest{{Path: "Cargo.toml", File: RawFile{Path: manifestPath}}}, Git: []RawGitOrigin{{DeclaredURL: "https://example.invalid/bad", LockedCommit: commit, Repository: RawTree{Root: repository}}}, Paths: []RawPathOrigin{{DeclaredPath: ".", Tree: RawTree{Root: workspace}}}})
	if artifactpolicy.ErrorCode(err) != artifactpolicy.CodeCompiledDependency {
		t.Fatalf("code=%q err=%v", artifactpolicy.ErrorCode(err), err)
	}
	if len(manager.state.oracleReceipts) != 0 {
		t.Fatalf("oracle receipts = %v", manager.state.oracleReceipts)
	}
	for _, forbidden := range []string{manager.state.cargoHome, manager.state.vendor} {
		if _, statErr := os.Lstat(forbidden); !os.IsNotExist(statErr) {
			t.Fatalf("forbidden pre-vendor path exists: %s (%v)", forbidden, statErr)
		}
	}
}

func TestGitProtectedSnapshotMutationPreventsOracleStart(t *testing.T) {
	repository := t.TempDir()
	if err := os.Mkdir(filepath.Join(repository, "src"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "Cargo.toml"), []byte("[package]\nname='leaf'\nversion='0.1.0'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "src", "lib.rs"), []byte("pub fn leaf() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitFixture(t, repository, "init", "-q")
	runGitFixture(t, repository, "config", "user.name", "Curator")
	runGitFixture(t, repository, "config", "user.email", "curator@example.invalid")
	runGitFixture(t, repository, "add", ".")
	runGitFixture(t, repository, "commit", "-qm", "fixture")
	commit := runGitFixture(t, repository, "rev-parse", "HEAD")
	declaredURL := "file://" + filepath.ToSlash(repository)
	lock, err := ParseLock([]byte("version = 4\n\n[[package]]\nname='leaf'\nversion='0.1.0'\nsource='git+" + declaredURL + "?rev=" + commit + "#" + commit + "'\n"))
	if err != nil {
		t.Fatal(err)
	}
	manager, err := NewManager(t.Context(), ManagerConfig{WorkRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	raw := RawGitOrigin{DeclaredURL: declaredURL, Selector: "rev=" + commit, LockedCommit: commit, Repository: RawTree{Root: repository}}
	if err = manager.preAdmitGitSourceTrees(t.Context(), lock, []RawGitOrigin{raw}); err != nil {
		t.Fatal(err)
	}
	root, _ := filepath.Abs(repository)
	protected, err := manager.state.gitInputs[root].adminInput.Tree.ProtectedPath()
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(protected, "Cargo.toml")
	if err = os.Chmod(manifest, 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(manifest, []byte("tampered\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err = manager.bindGit(t.Context(), lock, []RawGitOrigin{raw})
	var diagnostic *closureexec.DiagnosticError
	if !errors.As(err, &diagnostic) || diagnostic.Code != "closure_derivation_drift" {
		t.Fatalf("error = %v", err)
	}
	if len(manager.state.oracleReceipts) != 0 {
		t.Fatalf("oracle receipts = %v", manager.state.oracleReceipts)
	}
}

func runGitFixture(t *testing.T, root string, args ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, args...)...) // #nosec G204 -- fixed test fixture command.
	payload, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, payload)
	}
	return strings.TrimSpace(string(payload))
}

func TestCargo092StableURLShortHash(t *testing.T) {
	url := "file:///var/folders/cz/jqbthtks55zbkpcdkfyk4bl80000gp/T/tmp.Ir13lCGpQ5/repo"
	if got := cargoShortHash(url); got != "74903e34883e1a3b" {
		t.Fatalf("short hash = %s", got)
	}
}

func TestGitDerivationBindingRejectsAbsentManagerExecutor(t *testing.T) {
	if _, err := bindGitDerivation(nil, closureexec.DerivationReceipt{}, nil); ErrorCode(err) != CodeGitIdentityInvalid {
		t.Fatalf("code=%q err=%v", ErrorCode(err), err)
	}
}

type recordingRunner struct {
	commits, runs int
	expected      []VendorPackage
	destination   string
	invocation    vendorInvocation
	badPermit     bool
}

type metadataRecordingRunner struct {
	commits, runs int
	payload       []byte
}

func (runner *metadataRecordingRunner) CommitMetadata(_ context.Context, invocation metadataInvocation) (permit, error) {
	runner.commits++
	if err := invocation.validate(); err != nil {
		return permit{}, err
	}
	id, err := invocation.ID()
	if err != nil {
		return permit{}, err
	}
	return permit{ID: "sha256:" + digest([]byte("metadata-permit")), InvocationID: id}, nil
}
func (runner *metadataRecordingRunner) RunMetadata(_ context.Context, _ permit, _ metadataInvocation, recheck func() error) ([]byte, string, error) {
	if err := recheck(); err != nil {
		return nil, "", err
	}
	runner.runs++
	return append([]byte(nil), runner.payload...), "metadata-receipt", nil
}

func (runner *recordingRunner) CommitVendor(_ context.Context, invocation vendorInvocation) (permit, error) {
	runner.commits++
	runner.invocation = invocation
	if err := invocation.validate(); err != nil {
		return permit{}, err
	}
	id, err := invocation.ID()
	if err != nil {
		return permit{}, err
	}
	if runner.badPermit {
		id = "sha256:" + digest([]byte("widened"))
	}
	return permit{ID: "sha256:" + digest([]byte("permit")), InvocationID: id}, nil
}
func (runner *recordingRunner) RunVendor(_ context.Context, _ permit, invocation vendorInvocation, recheck func() error) (string, error) {
	if err := recheck(); err != nil {
		return "", err
	}
	runner.runs++
	if err := materializeVendorRunner(runner.destination, runner.expected); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(invocation.ConfigPath), 0o700); err != nil {
		return "", err
	}
	if err := os.WriteFile(invocation.ConfigPath, invocation.ConfigBytes, 0o600); err != nil {
		return "", err
	}
	return "receipt-1", nil
}

func TestCaptureAndVendorRejectsCompiledOriginBeforeTransformWithZeroSpawns(t *testing.T) {
	key := PackageKey{Name: "bad", Version: "1.0.0", Source: "registry+https://example.invalid/index"}
	archive := crateBytes(t, key, map[string][]byte{".cargo-checksum.json": []byte(`{"forged":true}`), "Cargo.toml": []byte("[package]\nname='bad'\nversion='1.0.0'\n"), "target/tool": {0, 'a', 's', 'm', 1, 0, 0, 0}})
	checksum := digest(archive)
	index, _ := json.Marshal(map[string]any{"name": "bad", "vers": "1.0.0", "cksum": checksum})
	lock := LockFile{Version: 4, Digest: digest([]byte("lock")), Packages: []LockPackage{{Key: key, Kind: SourceRegistry, Checksum: checksum}}}
	runner := &recordingRunner{}
	destination := filepath.Join(t.TempDir(), "vendor")
	home := filepath.Join(t.TempDir(), "cargo-home")
	configPath := filepath.Join(home, "config.toml")
	_, err := captureAndVendor(t.Context(), captureRequest{Lock: lock, Registry: []registryOrigin{{Package: key, IndexRecord: index, Archive: archive, Checksum: checksum}}, WorkspaceRoot: t.TempDir(), VendorDestination: destination, CargoHome: home, ConfigPath: configPath, StageCargoHome: stageTestCargoHome, Toolchain: testToolchain(), RecheckToolchain: func() (cargoToolchain, error) { return testToolchain(), nil }, Runner: runner})
	if err == nil {
		t.Fatal("compiled origin admitted")
	}
	if artifactpolicy.ErrorCode(err) != artifactpolicy.CodeCompiledDependency {
		t.Fatalf("admission did not precede transform: %v", err)
	}
	if runner.commits != 0 || runner.runs != 0 {
		t.Fatalf("Cargo activity commits=%d runs=%d", runner.commits, runner.runs)
	}
	if _, statErr := os.Stat(destination); !os.IsNotExist(statErr) {
		t.Fatalf("vendor destination exists: %v", statErr)
	}
	for _, path := range []string{home, configPath} {
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Fatalf("pre-admission output exists: %s", path)
		}
	}
}

func TestRustConformanceRH07CapturePreservesCompiledGitAndPathDiagnostics(t *testing.T) {
	compiled := []byte{0, 'a', 's', 'm', 1, 0, 0, 0}
	for _, testCase := range []struct {
		name string
		kind SourceKind
	}{
		{name: "git", kind: SourceGit},
		{name: "path", kind: SourcePath},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			workspace := t.TempDir()
			root := filepath.Join(workspace, "dependency")
			if err := os.MkdirAll(filepath.Join(root, "target"), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "target", "compiled"), compiled, 0o600); err != nil {
				t.Fatal(err)
			}
			key := PackageKey{Name: testCase.name, Version: "1.0.0"}
			request := captureRequest{WorkspaceRoot: workspace, VendorDestination: filepath.Join(t.TempDir(), "vendor"), CargoHome: filepath.Join(t.TempDir(), "cargo-home"), Toolchain: testToolchain(), RecheckToolchain: func() (cargoToolchain, error) { return testToolchain(), nil }, StageCargoHome: stageTestCargoHome}
			request.ConfigPath = filepath.Join(request.CargoHome, "config.toml")
			runner := &recordingRunner{}
			request.Runner = runner
			if testCase.kind == SourceGit {
				const commit = "0123456789abcdef0123456789abcdef01234567"
				key.Source = "git+https://example.invalid/repo#" + commit
				request.Lock = LockFile{Version: 4, Digest: digest([]byte("lock")), Packages: []LockPackage{{Key: key, Kind: SourceGit}}}
				request.Git = []gitOrigin{{Package: key, DeclaredURL: "https://example.invalid/repo", Commit: commit, Tree: "tree", Root: root, ManifestTracked: true, Leaves: []OriginLeaf{{Path: "target/compiled", SHA256: digest(compiled), Size: int64(len(compiled)), Bytes: compiled}}}}
			} else {
				request.Lock = LockFile{Version: 4, Digest: digest([]byte("lock")), Packages: []LockPackage{{Key: key, Kind: SourcePath}}}
				request.Paths = []pathOrigin{{Package: key, Root: root}}
			}
			_, err := captureAndVendor(t.Context(), request)
			if artifactpolicy.ErrorCode(err) != artifactpolicy.CodeCompiledDependency {
				t.Fatalf("diagnostic=%q err=%v", artifactpolicy.ErrorCode(err), err)
			}
			if runner.commits != 0 || runner.runs != 0 {
				t.Fatalf("Cargo activity commits=%d runs=%d", runner.commits, runner.runs)
			}
			if _, statErr := os.Lstat(request.VendorDestination); !os.IsNotExist(statErr) {
				t.Fatalf("vendor destination exists: %v", statErr)
			}
		})
	}
}

func TestCaptureAndVendorExactRegistryMapping(t *testing.T) {
	key := PackageKey{Name: "ok", Version: "1.0.0", Source: "registry+https://example.invalid/index"}
	archive := crateBytes(t, key, map[string][]byte{"Cargo.toml": []byte("[package]\nname='ok'\nversion='1.0.0'\n"), "src/lib.rs": []byte("pub fn ok() {}\n")})
	checksum := digest(archive)
	index, _ := json.Marshal(map[string]any{"name": "ok", "vers": "1.0.0", "cksum": checksum})
	origin := registryOrigin{Package: key, IndexRecord: index, Archive: archive, Checksum: checksum}
	expected, err := deriveRegistryTransform(origin)
	if err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(t.TempDir(), "vendor")
	runner := &recordingRunner{expected: []VendorPackage{expected}, destination: destination}
	lock := LockFile{Version: 4, Digest: digest([]byte("lock")), Packages: []LockPackage{{Key: key, Kind: SourceRegistry, Checksum: checksum}}}
	home := filepath.Join(t.TempDir(), "cargo-home")
	result, err := captureAndVendor(t.Context(), captureRequest{Lock: lock, Registry: []registryOrigin{origin}, WorkspaceRoot: t.TempDir(), VendorDestination: destination, CargoHome: home, ConfigPath: filepath.Join(home, "config.toml"), StageCargoHome: stageTestCargoHome, Toolchain: testToolchain(), RecheckToolchain: func() (cargoToolchain, error) { return testToolchain(), nil }, Runner: runner})
	if err != nil {
		t.Fatal(err)
	}
	if runner.commits != 1 || runner.runs != 1 || result.VendorReceipt == "" || len(result.ArtifactManifestIDs) != 3 {
		t.Fatalf("result=%#v commits=%d runs=%d", result, runner.commits, runner.runs)
	}
	if runner.invocation.CargoHomeDigest == "" || runner.invocation.ConfigSHA256 == "" || runner.invocation.Environment["CARGO_HOME"] == "" {
		t.Fatalf("permit omitted Cargo home/config binding: %#v", runner.invocation)
	}
	badDestination := filepath.Join(t.TempDir(), "vendor")
	badRunner := &recordingRunner{expected: []VendorPackage{expected}, destination: badDestination, badPermit: true}
	badHome := filepath.Join(t.TempDir(), "cargo-home")
	_, err = captureAndVendor(t.Context(), captureRequest{Lock: lock, Registry: []registryOrigin{origin}, WorkspaceRoot: t.TempDir(), VendorDestination: badDestination, CargoHome: badHome, ConfigPath: filepath.Join(badHome, "config.toml"), StageCargoHome: stageTestCargoHome, Toolchain: testToolchain(), RecheckToolchain: func() (cargoToolchain, error) { return testToolchain(), nil }, Runner: badRunner})
	if ErrorCode(err) != CodeVendorIncomplete {
		t.Fatalf("widened permit code=%q err=%v", ErrorCode(err), err)
	}
	if badRunner.runs != 0 {
		t.Fatalf("widened permit spawned Cargo %d times", badRunner.runs)
	}
	if _, statErr := os.Stat(badDestination); !os.IsNotExist(statErr) {
		t.Fatalf("widened permit created vendor destination")
	}
}

func TestCaptureAndVendorUsesOnlyPipelineOwnedGitDerivation(t *testing.T) {
	const commit = "0123456789abcdef0123456789abcdef01234567"
	manifest := []byte("[package]\nname = \"git_leaf\"\nversion = \"0.1.0\"\n")
	source := []byte("pub fn value() -> u8 { 7 }\n")
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "Cargo.toml"), manifest, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "lib.rs"), source, 0o600); err != nil {
		t.Fatal(err)
	}
	key := PackageKey{Name: "git_leaf", Version: "0.1.0", Source: "git+https://example.invalid/repo#" + commit}
	origin := gitOrigin{Package: key, DeclaredURL: "https://example.invalid/repo", Commit: commit, Tree: "tree", Root: root, ManifestTracked: true, Leaves: []OriginLeaf{{Path: "Cargo.toml", SHA256: digest(manifest), Size: int64(len(manifest)), Bytes: manifest}, {Path: "src/lib.rs", SHA256: digest(source), Size: int64(len(source)), Bytes: source}}}
	derivation := gitDerivation{mode: ProjectionGitIndexNoInclude, selected: []string{"Cargo.toml", "src/lib.rs"}, normalizerInputs: []string{"Cargo.toml", "src/lib.rs"}, normalizerID: NormalizerID, normalizedManifest: []byte("# normalized\n[package]\nname = \"git_leaf\"\nversion = \"0.1.0\"\n"), receiptID: "manager-receipt", commit: commit, tree: "tree", manifestTracked: true}
	var err error
	derivation.seal, err = gitDerivationSeal(derivation)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := deriveGitTransform(origin, derivation)
	if err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(t.TempDir(), "vendor")
	home := filepath.Join(t.TempDir(), "cargo-home")
	runner := &recordingRunner{expected: []VendorPackage{expected}, destination: destination}
	request := captureRequest{Lock: LockFile{Version: 4, Digest: digest([]byte("lock")), Packages: []LockPackage{{Key: key, Kind: SourceGit}}}, Git: []gitOrigin{origin}, WorkspaceRoot: t.TempDir(), VendorDestination: destination, CargoHome: home, ConfigPath: filepath.Join(home, "config.toml"), StageCargoHome: stageTestCargoHome, Toolchain: testToolchain(), RecheckToolchain: func() (cargoToolchain, error) { return testToolchain(), nil }, Runner: runner, gitDerivations: map[string]gitDerivation{key.String(): derivation}}
	result, err := captureAndVendor(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}
	if runner.commits != 1 || runner.runs != 1 || len(result.VendorPackages) != 1 {
		t.Fatalf("result=%#v commits=%d runs=%d", result, runner.commits, runner.runs)
	}
}

func TestSourceReplacementConfigCoversEveryRemoteLockSource(t *testing.T) {
	vendor := filepath.Join(t.TempDir(), "vendor")
	packages := []LockPackage{{Kind: SourceRegistry, Key: PackageKey{Name: "a", Version: "1", Source: "registry+https://github.com/rust-lang/crates.io-index"}}, {Kind: SourceRegistry, Key: PackageKey{Name: "b", Version: "1", Source: "registry+https://registry.example/index"}}, {Kind: SourceGit, Key: PackageKey{Name: "c", Version: "1", Source: "git+https://git.example/repo?branch=main#0123456789abcdef0123456789abcdef01234567"}}}
	payload, err := DeriveSourceReplacementConfig(vendor, packages)
	if err != nil {
		t.Fatal(err)
	}
	for _, needle := range []string{"[source.crates-io]", "registry+https://registry.example/index", "[source.\"git+https://git.example/repo?branch=main\"]", "[source.vendored-sources]", filepath.ToSlash(vendor)} {
		if !bytes.Contains(payload, []byte(needle)) {
			t.Fatalf("config missing %q: %s", needle, payload)
		}
	}
	if bytes.Contains(payload, []byte("#0123456789abcdef0123456789abcdef01234567")) {
		t.Fatalf("precise Git commit leaked into replacement key: %s", payload)
	}
	expected := "[source.crates-io]\nreplace-with = \"vendored-sources\"\n\n[source.\"git+https://git.example/repo?branch=main\"]\ngit = \"https://git.example/repo\"\nbranch = \"main\"\nreplace-with = \"vendored-sources\"\n\n[source.\"registry+https://registry.example/index\"]\nregistry = \"https://registry.example/index\"\nreplace-with = \"vendored-sources\"\n\n[source.vendored-sources]\ndirectory = \"" + filepath.ToSlash(vendor) + "\"\n"
	if string(payload) != expected {
		t.Fatalf("config bytes differ\nwant: %q\n got: %q", expected, payload)
	}
}

func TestMetadataToolchainDriftCausesZeroSpawns(t *testing.T) {
	runner := &metadataRecordingRunner{}
	expected := testToolchain()
	drifted := expected
	drifted.Fingerprint = "sha256:" + digest([]byte("drift"))
	home, homeDigest, configPath, configBytes := metadataInputs(t)
	_, err := runPermittedMetadata(t.Context(), metadataRequest{CaptureID: "sha256:" + digest([]byte("capture")), WorkspaceRoot: t.TempDir(), ManifestPath: filepath.Join(t.TempDir(), "Cargo.toml"), CargoHome: home, CargoHomeDigest: homeDigest, ConfigPath: configPath, ConfigBytes: configBytes, Selection: SelectionContext{Target: "native"}, Toolchain: expected, RecheckToolchain: func() (cargoToolchain, error) { return drifted, nil }, Runner: runner})
	if ErrorCode(err) != CodeVendorTransformUnsupported {
		t.Fatalf("code=%q err=%v", ErrorCode(err), err)
	}
	if runner.commits != 1 || runner.runs != 0 {
		t.Fatalf("commits=%d runs=%d", runner.commits, runner.runs)
	}
}

func TestPermittedMetadataUsesDistinctExactPermits(t *testing.T) {
	payload := []byte(`{"packages":[{"id":"cli 0.1.0 (path+file:///workspace)","name":"cli","version":"0.1.0","source":null,"manifest_path":"workspace/Cargo.toml","links":null,"targets":[{"name":"cli","kind":["bin"],"crate_types":["bin"],"src_path":"workspace/src/main.rs"}]}],"resolve":{"nodes":[{"id":"cli 0.1.0 (path+file:///workspace)","features":[],"dependencies":[],"deps":[]}]},"version":1,"workspace_root":"/workspace","target_directory":"/target","workspace_members":[],"workspace_default_members":[],"metadata":null}`)
	runner := &metadataRecordingRunner{payload: payload}
	toolchain := testToolchain()
	home, homeDigest, configPath, configBytes := metadataInputs(t)
	result, err := runPermittedMetadata(t.Context(), metadataRequest{CaptureID: "sha256:" + digest([]byte("capture")), WorkspaceRoot: t.TempDir(), ManifestPath: filepath.Join(t.TempDir(), "Cargo.toml"), CargoHome: home, CargoHomeDigest: homeDigest, ConfigPath: configPath, ConfigBytes: configBytes, Selection: SelectionContext{Target: "native", Features: []string{"z", "a"}}, Toolchain: toolchain, RecheckToolchain: func() (cargoToolchain, error) { return toolchain, nil }, Runner: runner})
	if err != nil {
		t.Fatal(err)
	}
	if runner.commits != 2 || runner.runs != 2 || result.UnfilteredReceipt == "" || result.ActiveReceipt == "" {
		t.Fatalf("result=%#v commits=%d runs=%d", result, runner.commits, runner.runs)
	}
}

func TestParseMetadataMalformedNestedValuesFailClosed(t *testing.T) {
	fixtures := [][]byte{[]byte(`{"packages":[],"resolve":{"nodes":[1]}}`), []byte(`{"packages":[],"resolve":{"nodes":[{"id":"x","features":[],"deps":[1]}]}}`), []byte(`{"packages":[{"id":"x","name":"x","version":"1","source":[],"manifest_path":"x/Cargo.toml","targets":[]}],"resolve":{"nodes":[]}}`), []byte(`{"packages":[{"id":"x","name":"x","version":"1","source":null,"manifest_path":"x/Cargo.toml","targets":[{"name":"x","src_path":"x/lib.rs","kind":["lib"],"crate_types":["lib"],"runner":"bad"}]}],"resolve":{"nodes":[]}}`)}
	for _, fixture := range fixtures {
		if _, err := ParseMetadata(fixture); ErrorCode(err) != CodeGraphIncomplete {
			t.Fatalf("code=%q err=%v", ErrorCode(err), err)
		}
	}
}

func TestMetadataConfigDriftCausesZeroPermitsAndSpawns(t *testing.T) {
	runner := &metadataRecordingRunner{}
	toolchain := testToolchain()
	home, homeDigest, configPath, configBytes := metadataInputs(t)
	configBytes = append(configBytes, '#')
	_, err := runPermittedMetadata(t.Context(), metadataRequest{CaptureID: "sha256:" + digest([]byte("capture")), WorkspaceRoot: t.TempDir(), ManifestPath: filepath.Join(t.TempDir(), "Cargo.toml"), CargoHome: home, CargoHomeDigest: homeDigest, ConfigPath: configPath, ConfigBytes: configBytes, Selection: SelectionContext{Target: "native"}, Toolchain: toolchain, RecheckToolchain: func() (cargoToolchain, error) { return toolchain, nil }, Runner: runner})
	if ErrorCode(err) != CodeConfigUntrusted {
		t.Fatalf("code=%q err=%v", ErrorCode(err), err)
	}
	if runner.commits != 0 || runner.runs != 0 {
		t.Fatalf("commits=%d runs=%d", runner.commits, runner.runs)
	}
}

func TestCGP05ReconcileReusesSelectionNeutralCaptureAndSeparatesTargetBinding(t *testing.T) {
	lock := LockFile{Version: 4, Digest: digest([]byte("lock")), Packages: []LockPackage{{Key: PackageKey{Name: "cli", Version: "0.1.0"}, Kind: SourcePath}}}
	manifest := Manifest{Path: "workspace/Cargo.toml", PackageName: "cli", PackageVersion: "0.1.0", Digest: digest([]byte("manifest"))}
	capture, err := NewCaptureGraph(lock, []Manifest{manifest}, []string{"sha256:" + digest([]byte("artifact"))})
	if err != nil {
		t.Fatal(err)
	}
	metadata := Metadata{Packages: []MetadataPackage{{ID: "cli 0.1.0 (path+file:///workspace)", Name: "cli", Version: "0.1.0", ManifestPath: "workspace/Cargo.toml", Targets: []MetadataTarget{{Name: "cli", Kinds: []string{"bin"}, CrateTypes: []string{"bin"}, SrcPath: "workspace/src/main.rs"}}}}, Resolve: []MetadataNode{{ID: "cli 0.1.0 (path+file:///workspace)", Features: []string{}}}}
	selection := SelectionContext{Package: "cli", Binary: "cli", Target: "aarch64-apple-darwin", ResolvedFeatures: map[string][]string{"cli 0.1.0 (path+file:///workspace)": {}}}
	darwin, err := Reconcile(capture, selection, metadata, selection.Target, "sha256:"+digest([]byte("cargo-darwin")))
	if err != nil {
		t.Fatal(err)
	}
	selection.Target = "x86_64-unknown-linux-gnu"
	linux, err := Reconcile(capture, selection, metadata, selection.Target, "sha256:"+digest([]byte("cargo-linux")))
	if err != nil {
		t.Fatal(err)
	}
	if darwin.CaptureID != linux.CaptureID || darwin.Identity == linux.Identity {
		t.Fatalf("capture=%s/%s active=%s/%s", darwin.CaptureID, linux.CaptureID, darwin.Identity, linux.Identity)
	}
}

func testToolchain() cargoToolchain {
	return cargoToolchain{CargoPath: "/toolchain/bin/cargo", Version: "1.91.0", ImplementationCommit: "ea2d97820c16195b0ca3fadb4319fe512c199a43", BinarySHA256: "sha256:" + digest([]byte("cargo")), Fingerprint: "sha256:" + digest([]byte("toolchain")), C0CheckpointID: "sha256:" + digest([]byte("c0"))}
}
func crateBytes(t *testing.T, key PackageKey, files map[string][]byte) []byte {
	t.Helper()
	var output bytes.Buffer
	gzipWriter := gzip.NewWriter(&output)
	tarWriter := tar.NewWriter(gzipWriter)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sortStrings(names)
	for _, name := range names {
		payload := files[name]
		if err := tarWriter.WriteHeader(&tar.Header{Name: key.Name + "-" + key.Version + "/" + name, Mode: 0o644, Size: int64(len(payload)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tarWriter.Write(payload); err != nil {
			t.Fatal(err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}
func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}
func materializeVendor(t *testing.T, root string, packages []VendorPackage) {
	t.Helper()
	if err := materializeVendorRunner(root, packages); err != nil {
		t.Fatal(err)
	}
}
func materializeVendorRunner(root string, packages []VendorPackage) error {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return err
	}
	for _, pkg := range packages {
		base := filepath.Join(root, pkg.Directory)
		for _, leaf := range pkg.Files {
			target := filepath.Join(base, filepath.FromSlash(leaf.Path))
			if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				return err
			}
			if err := os.WriteFile(target, leaf.Bytes, 0o600); err != nil {
				return err
			}
		}
	}
	return nil
}

func stageTestCargoHome(_ context.Context, root string, manifestIDs []string) error {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return err
	}
	payload, err := json.Marshal(manifestIDs)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "admitted-inputs.json"), payload, 0o600)
}
func metadataInputs(t *testing.T) (string, string, string, []byte) {
	t.Helper()
	home := filepath.Join(t.TempDir(), "cargo-home")
	if err := stageTestCargoHome(t.Context(), home, []string{"sha256:" + digest([]byte("input"))}); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(home, "config.toml")
	configBytes := []byte("[source.curator-vendor]\ndirectory = \"/vendor\"\n")
	if err := os.WriteFile(configPath, configBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	homeDigest, err := directoryDigest(home)
	if err != nil {
		t.Fatal(err)
	}
	return home, homeDigest, configPath, configBytes
}
