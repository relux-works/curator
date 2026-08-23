package rustsource

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
)

func TestRustConformanceR03R05R06R07PathWorkspaceBuild(t *testing.T) {
	workspace := t.TempDir()
	files := map[string]string{
		"Cargo.toml":             "[workspace]\nmembers=[\"app\",\"dep1\",\"dep2\"]\nexclude=[\"unix_dep\",\"windows_dep\"]\nresolver=\"2\"\n",
		"Cargo.lock":             "version = 4\n\n[[package]]\nname=\"app\"\nversion=\"0.1.0\"\ndependencies=[\"dep1\",\"platform_unix\",\"platform_windows\"]\n\n[[package]]\nname=\"dep1\"\nversion=\"0.1.0\"\ndependencies=[\"dep2\"]\n\n[[package]]\nname=\"dep2\"\nversion=\"0.1.0\"\n\n[[package]]\nname=\"platform_unix\"\nversion=\"0.1.0\"\n\n[[package]]\nname=\"platform_windows\"\nversion=\"0.1.0\"\n",
		"app/Cargo.toml":         "[package]\nname=\"app\"\nversion=\"0.1.0\"\nedition=\"2021\"\n[dependencies]\ndep1={path=\"../dep1\"}\n[target.'cfg(unix)'.dependencies]\nplatform_unix={path=\"../unix_dep\"}\n[target.'cfg(windows)'.dependencies]\nplatform_windows={path=\"../windows_dep\"}\n[[bin]]\nname=\"selected-bin\"\npath=\"src/main.rs\"\n",
		"app/src/main.rs":        "#[cfg(unix)] use platform_unix as platform; #[cfg(windows)] use platform_windows as platform; include!(\"fragment.rs\"); const TEXT:&str=include_str!(\"message.txt\"); const BYTES:&[u8]=include_bytes!(\"bytes.txt\"); fn main(){println!(\"{}-{}-{}-{}\",TEXT.trim(),dep1::value()+platform::value(),BYTES.len(),included_value());}\n",
		"app/src/fragment.rs":    "fn included_value()->u8{13}\n",
		"app/src/message.txt":    "closure\n",
		"app/src/bytes.txt":      "abc",
		"dep1/Cargo.toml":        "[package]\nname=\"dep1\"\nversion=\"0.1.0\"\nedition=\"2021\"\n[dependencies]\ndep2={path=\"../dep2\"}\n",
		"dep1/src/lib.rs":        "pub fn value()->u8{dep2::value()}\n",
		"dep2/Cargo.toml":        "[package]\nname=\"dep2\"\nversion=\"0.1.0\"\nedition=\"2021\"\n",
		"dep2/src/lib.rs":        "pub fn value()->u8{5}\n",
		"unix_dep/Cargo.toml":    "[package]\nname=\"platform_unix\"\nversion=\"0.1.0\"\nedition=\"2021\"\n",
		"unix_dep/src/lib.rs":    "pub fn value()->u8{2}\n",
		"windows_dep/Cargo.toml": "[package]\nname=\"platform_windows\"\nversion=\"0.1.0\"\nedition=\"2021\"\n",
		"windows_dep/src/lib.rs": "pub fn value()->u8{9}\n",
	}
	writeRustFixtureFiles(t, workspace, files)

	manifestNames := []string{"Cargo.toml", "app/Cargo.toml", "dep1/Cargo.toml", "dep2/Cargo.toml", "unix_dep/Cargo.toml", "windows_dep/Cargo.toml"}
	manifests := make([]RawManifest, 0, len(manifestNames))
	for _, name := range manifestNames {
		manifests = append(manifests, RawManifest{Path: name, File: RawFile{Path: filepath.Join(workspace, filepath.FromSlash(name))}})
	}
	paths := []RawPathOrigin{}
	for _, name := range []string{"app", "dep1", "dep2", "unix_dep", "windows_dep"} {
		paths = append(paths, RawPathOrigin{DeclaredPath: name, Tree: RawTree{Root: filepath.Join(workspace, name)}})
	}
	manager, err := NewManager(t.Context(), ManagerConfig{WorkRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	capture, err := manager.Capture(t.Context(), RawCaptureRequest{Workspace: RawTree{Root: workspace}, Lock: RawFile{Path: filepath.Join(workspace, "Cargo.lock")}, Manifests: manifests, Paths: paths})
	if err != nil {
		t.Fatal(err)
	}
	if len(capture.state.graph.Packages) != 5 {
		t.Fatalf("R03/R05 lock superset packages=%d", len(capture.state.graph.Packages))
	}
	selection := SelectionContext{Package: "app", Binary: "selected-bin", Target: manager.state.buildTools.target, DefaultFeatures: true, Features: []string{}, TargetCFG: []string{}}
	metadata, err := manager.DeriveMetadata(t.Context(), capture, selection)
	if err != nil {
		t.Fatal(err)
	}
	selection = metadata.ResolvedSelection()
	names := []string{}
	for _, pkg := range metadata.Active.Packages {
		names = append(names, pkg.Name)
	}
	sort.Strings(names)
	present := strings.Join(names, ",")
	want, pruned := "platform_unix", "platform_windows"
	if runtime.GOOS == "windows" {
		want, pruned = pruned, want
	}
	if !strings.Contains(present, want) || strings.Contains(present, pruned) {
		t.Fatalf("R05 target pruning active=%v want=%s pruned=%s", names, want, pruned)
	}
	protectedWorkspace, err := manager.state.workspaceInput.Tree.ProtectedPath()
	if err != nil {
		t.Fatal(err)
	}
	fragment, err := os.ReadFile(filepath.Join(protectedWorkspace, "app", "src", "fragment.rs"))
	if err != nil || string(fragment) != files["app/src/fragment.rs"] {
		t.Fatalf("R07 include! leaf missing from protected closure: payload=%q err=%v", fragment, err)
	}
	workspaceReceiptID, err := manager.state.workspaceInput.Receipt.ID()
	if err != nil || workspaceReceiptID != manager.state.workspaceID {
		t.Fatalf("R07 workspace closure receipt invalid: id=%s want=%s err=%v", workspaceReceiptID, manager.state.workspaceID, err)
	}
	publication, binding := rustPublicationFixture(t, manager.state.buildTools, selection)
	var buildPermits []closureexec.DerivationPermit
	manager.state.processRunner.ProcessStartObserver = func(permit closureexec.DerivationPermit) {
		buildPermits = append(buildPermits, permit)
	}
	result, err := manager.Build(t.Context(), BuildRequest{Capture: capture, Metadata: metadata, Selection: selection, Binding: binding, Publication: publication, StoreRoot: filepath.Join(t.TempDir(), "protected")})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(result.ArtifactPath)
	if err != nil || len(payload) == 0 || result.Publication.Decision != "published" {
		t.Fatalf("R03/R06/R07 build payload=%d publication=%q err=%v", len(payload), result.Publication.Decision, err)
	}
	if len(buildPermits) != 2 {
		t.Fatalf("R07 build permit count=%d", len(buildPermits))
	}
	for _, permit := range buildPermits {
		found := false
		for _, receiptID := range permit.AdmittedInputReceiptIDs {
			found = found || receiptID == manager.state.workspaceID
		}
		if !found {
			t.Fatalf("R07 include! closure receipt %s absent from %s permit inputs: %#v", manager.state.workspaceID, permit.InvocationKey, permit.AdmittedInputReceiptIDs)
		}
	}
}

func TestRustConformanceR02GitBuildWithoutOriginalRemoteOrCache(t *testing.T) {
	submodule := t.TempDir()
	writeRustFixtureFiles(t, submodule, map[string]string{"nested.txt": "submodule evidence\n"})
	runGitFixture(t, submodule, "init", "-q")
	runGitFixture(t, submodule, "config", "user.name", "Curator Test")
	runGitFixture(t, submodule, "config", "user.email", "curator@example.invalid")
	runGitFixture(t, submodule, "add", "nested.txt")
	runGitFixture(t, submodule, "commit", "-qm", "submodule fixture")

	repository := t.TempDir()
	writeRustFixtureFiles(t, repository, map[string]string{
		"Cargo.toml": "[package]\nname=\"git_leaf\"\nversion=\"0.1.0\"\nedition=\"2021\"\n",
		"src/lib.rs": "pub fn value()->u8{11}\n",
	})
	runGitFixture(t, repository, "init", "-q")
	runGitFixture(t, repository, "config", "user.name", "Curator Test")
	runGitFixture(t, repository, "config", "user.email", "curator@example.invalid")
	runGitFixture(t, repository, "-c", "protocol.file.allow=always", "submodule", "add", "-q", "file://"+filepath.ToSlash(submodule), "deps/sub")
	runGitFixture(t, repository, "add", "Cargo.toml", "src/lib.rs")
	runGitFixture(t, repository, "commit", "-qm", "fixture")
	commit := runGitFixture(t, repository, "rev-parse", "HEAD")
	declaredURL := "file://" + filepath.ToSlash(repository)

	workspace := t.TempDir()
	rootManifest := "[package]\nname=\"git_app\"\nversion=\"0.1.0\"\nedition=\"2021\"\n[dependencies]\ngit_leaf={git=\"" + declaredURL + "\",rev=\"" + commit + "\"}\n[[bin]]\nname=\"git-app\"\npath=\"src/main.rs\"\n"
	source := "git+" + declaredURL + "?rev=" + commit + "#" + commit
	writeRustFixtureFiles(t, workspace, map[string]string{
		"Cargo.toml":  rootManifest,
		"Cargo.lock":  "version = 4\n\n[[package]]\nname=\"git_app\"\nversion=\"0.1.0\"\ndependencies=[\"git_leaf\"]\n\n[[package]]\nname=\"git_leaf\"\nversion=\"0.1.0\"\nsource=\"" + source + "\"\n",
		"src/main.rs": "fn main(){println!(\"git-{}\",git_leaf::value());}\n",
	})
	manager, err := NewManager(t.Context(), ManagerConfig{WorkRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	capture, err := manager.Capture(t.Context(), RawCaptureRequest{Workspace: RawTree{Root: workspace}, Lock: RawFile{Path: filepath.Join(workspace, "Cargo.lock")}, Manifests: []RawManifest{{Path: "Cargo.toml", File: RawFile{Path: filepath.Join(workspace, "Cargo.toml")}}}, Git: []RawGitOrigin{{DeclaredURL: declaredURL, Selector: "rev=" + commit, LockedCommit: commit, Repository: RawTree{Root: repository}}}, Paths: []RawPathOrigin{{DeclaredPath: ".", Tree: RawTree{Root: workspace}}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(capture.Evidence.GitObjectReceipts) != 1 || len(capture.Evidence.GitProjectionReceipts) != 1 {
		t.Fatalf("R02 Git receipts=%#v/%#v", capture.Evidence.GitObjectReceipts, capture.Evidence.GitProjectionReceipts)
	}
	unavailable := repository + "-unavailable"
	if err = os.Rename(repository, unavailable); err != nil {
		t.Fatal(err)
	}
	selection := SelectionContext{Package: "git_app", Binary: "git-app", Target: manager.state.buildTools.target, DefaultFeatures: true, Features: []string{}, TargetCFG: []string{}}
	metadata, err := manager.DeriveMetadata(t.Context(), capture, selection)
	if err != nil {
		t.Fatal(err)
	}
	selection = metadata.ResolvedSelection()
	publication, binding := rustPublicationFixture(t, manager.state.buildTools, selection)
	result, err := manager.Build(t.Context(), BuildRequest{Capture: capture, Metadata: metadata, Selection: selection, Binding: binding, Publication: publication, StoreRoot: filepath.Join(t.TempDir(), "protected")})
	if err != nil {
		t.Fatal(err)
	}
	if result.ArtifactPath == "" || result.Publication.Decision != "published" {
		t.Fatalf("R02 build=%#v", result)
	}
}

func TestRustConformanceRF12ClosedTargetAndUnstableArtifactInputs(t *testing.T) {
	_, _, selection, tools := buildBindingFixture(t)
	for _, target := range []string{"custom-target.json", "wasm32-unknown-unknown", selection.Target + "," + selection.Target} {
		value := selection
		value.Target = target
		if _, err := Reconcile(CaptureGraph{}, value, Metadata{}, tools.target, string(testBuildID('1'))); ErrorCode(err) != CodeTargetUnsupported {
			t.Fatalf("RF12 target=%q error=%v", target, err)
		}
	}
	for name, manifest := range map[string][]byte{
		"unstable-Z-config":   []byte("[package]\nname='app'\nversion='0.1.0'\n[unstable]\nbuild-std=true\n"),
		"artifact-dependency": []byte("[package]\nname='app'\nversion='0.1.0'\n[dependencies]\ntool={version='1',artifact='bin'}\n"),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseManifest("Cargo.toml", manifest); ErrorCode(err) != CodeGraphIncomplete {
				t.Fatalf("RF12 error=%v", err)
			}
		})
	}
}

func TestRustConformanceRF09RF10RF11FailBeforeCompilation(t *testing.T) {
	t.Run("RF09_unknown_target_kind", func(t *testing.T) {
		payload := []byte(`{"packages":[{"id":"path+file:///workspace#app@0.1.0","name":"app","version":"0.1.0","source":null,"manifest_path":"workspace/Cargo.toml","links":null,"targets":[{"name":"app","kind":["future-kind"],"crate_types":["bin"],"src_path":"workspace/src/main.rs"}]}],"resolve":{"nodes":[{"id":"path+file:///workspace#app@0.1.0","features":[],"dependencies":[],"deps":[]}]}}`)
		if _, err := ParseMetadata(payload); ErrorCode(err) != CodeGraphIncomplete {
			t.Fatalf("RF09 error=%v", err)
		}
	})
	t.Run("RF09_source_or_path_outside_capture", func(t *testing.T) {
		lock := LockFile{Version: 4, Digest: digest([]byte("lock")), Packages: []LockPackage{{Key: PackageKey{Name: "app", Version: "0.1.0"}, Kind: SourcePath}}}
		manifest := Manifest{Path: "workspace/Cargo.toml", PackageName: "app", PackageVersion: "0.1.0", Digest: digest([]byte("manifest"))}
		capture, err := NewCaptureGraph(lock, []Manifest{manifest}, []string{"sha256:" + digest([]byte("artifact"))})
		if err != nil {
			t.Fatal(err)
		}
		id := "path+file:///workspace#app@0.1.0"
		metadata := Metadata{Packages: []MetadataPackage{{ID: id, Name: "app", Version: "0.1.0", ManifestPath: "/host/outside/Cargo.toml", Targets: []MetadataTarget{{Name: "app", Kinds: []string{"bin"}, CrateTypes: []string{"bin"}, SrcPath: "/host/outside/main.rs"}}}}, Resolve: []MetadataNode{{ID: id, Features: []string{}}}}
		selection := SelectionContext{Package: "app", Binary: "app", Target: "aarch64-apple-darwin", ResolvedFeatures: map[string][]string{id: {}}}
		if _, err = Reconcile(capture, selection, metadata, selection.Target, string(testBuildID('1'))); ErrorCode(err) != CodeGraphIncomplete {
			t.Fatalf("RF09 error=%v", err)
		}
	})
	t.Run("RF10_feature_and_target_drift", func(t *testing.T) {
		selection := SelectionContext{Target: "native", Features: []string{"same", "same"}}
		if _, err := Reconcile(CaptureGraph{}, selection, Metadata{}, selection.Target, string(testBuildID('1'))); ErrorCode(err) != CodeFeatureProfileMismatch {
			t.Fatalf("RF10 feature error=%v", err)
		}
		selection.Features = nil
		if _, err := Reconcile(CaptureGraph{}, selection, Metadata{}, "different-native", string(testBuildID('1'))); ErrorCode(err) != CodeTargetUnsupported {
			t.Fatalf("RF10 target error=%v", err)
		}
	})
	t.Run("RF11_config_and_wrapper_rejection", func(t *testing.T) {
		workspace := t.TempDir()
		writeRustFixtureFiles(t, workspace, map[string]string{".cargo/config.toml": "[build]\nrustc-wrapper='host-wrapper'\n"})
		if err := rejectBuildConfiguration(workspace); ErrorCode(err) != CodeConfigUntrusted {
			t.Fatalf("RF11 config error=%v", err)
		}
		environment := buildEnvironment(rustBuildToolchain{target: "aarch64-apple-darwin", items: map[BuildToolRole]BuildToolEvidence{BuildToolRustc: {PhysicalPath: "/tool/rustc"}, BuildToolLinker: {PhysicalPath: "/tool/cc"}, BuildToolSDK: {PhysicalPath: "/sdk"}, BuildToolSysroot: {PhysicalPath: "/sysroot"}}}, filepath.Join(t.TempDir(), "cargo"), filepath.Join(t.TempDir(), "target"), filepath.Join(t.TempDir(), "tmp"), t.TempDir())
		for _, key := range []string{"RUSTC_WRAPPER", "RUSTC_WORKSPACE_WRAPPER", "CARGO_BUILD_RUSTC_WRAPPER", "CARGO_TARGET_DIR_FROM_HOST"} {
			if _, ok := environment[key]; ok {
				t.Fatalf("RF11 ambient wrapper %s propagated", key)
			}
		}
	})
}

func TestRustConformanceRH09ToolchainCopyIsDependencyArtifact(t *testing.T) {
	workspace := t.TempDir()
	root := filepath.Join(workspace, "dependency")
	writeRustFixtureFiles(t, root, map[string]string{"Cargo.toml": "[package]\nname='bad'\nversion='1.0.0'\n", "vendor/toolchain/cargo": string([]byte{0, 'a', 's', 'm', 1, 0, 0, 0})})
	key := PackageKey{Name: "bad", Version: "1.0.0"}
	runner := &recordingRunner{}
	home := filepath.Join(t.TempDir(), "cargo-home")
	request := captureRequest{Lock: LockFile{Version: 4, Digest: digest([]byte("lock")), Packages: []LockPackage{{Key: key, Kind: SourcePath}}}, Paths: []pathOrigin{{Package: key, Root: root}}, WorkspaceRoot: workspace, VendorDestination: filepath.Join(t.TempDir(), "vendor"), CargoHome: home, ConfigPath: filepath.Join(home, "config.toml"), Toolchain: testToolchain(), RecheckToolchain: func() (cargoToolchain, error) { return testToolchain(), nil }, StageCargoHome: stageTestCargoHome, Runner: runner}
	_, err := captureAndVendor(t.Context(), request)
	if artifactpolicy.ErrorCode(err) != artifactpolicy.CodeCompiledDependency || runner.commits != 0 || runner.runs != 0 {
		t.Fatalf("RH09 diagnostic=%q commits=%d runs=%d err=%v", artifactpolicy.ErrorCode(err), runner.commits, runner.runs, err)
	}
}

func TestRustConformanceRH10BuildRejectsToolchainDriftBeforeProcessOrPublication(t *testing.T) {
	workspace := t.TempDir()
	writeRustFixtureFiles(t, workspace, map[string]string{
		"Cargo.toml":  "[package]\nname='rh10_app'\nversion='0.1.0'\nedition='2021'\n[[bin]]\nname='rh10-app'\npath='src/main.rs'\n",
		"Cargo.lock":  "version = 4\n\n[[package]]\nname='rh10_app'\nversion='0.1.0'\n",
		"src/main.rs": "fn main(){println!(\"rh10\");}\n",
	})
	manager, err := NewManager(t.Context(), ManagerConfig{WorkRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	capture, err := manager.Capture(t.Context(), RawCaptureRequest{
		Workspace: RawTree{Root: workspace},
		Lock:      RawFile{Path: filepath.Join(workspace, "Cargo.lock")},
		Manifests: []RawManifest{{Path: "Cargo.toml", File: RawFile{Path: filepath.Join(workspace, "Cargo.toml")}}},
		Paths:     []RawPathOrigin{{DeclaredPath: ".", Tree: RawTree{Root: workspace}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	selection := SelectionContext{Package: "rh10_app", Binary: "rh10-app", Target: manager.state.buildTools.target, DefaultFeatures: true, Features: []string{}, TargetCFG: []string{}}
	metadata, err := manager.DeriveMetadata(t.Context(), capture, selection)
	if err != nil {
		t.Fatal(err)
	}
	selection = metadata.ResolvedSelection()
	publication, binding := rustPublicationFixture(t, manager.state.buildTools, selection)
	provider := newRustBuildProvider()
	enableVerifiedBuild(t, manager, provider)
	operation, err := manager.state.executor.Preflight(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	manager.state.buildOperation = operation
	provider.driftExecutable = filepath.Join(manager.state.execRoot, "bin", "cargo")
	provider.driftOnNegotiation = provider.negotiations + 2 // Build revalidation, then the committed Executor start seam.
	storeRoot := filepath.Join(t.TempDir(), "protected")
	result, err := manager.Build(t.Context(), BuildRequest{Capture: capture, Metadata: metadata, Selection: selection, Binding: binding, Publication: publication, StoreRoot: storeRoot})
	if ErrorCode(err) != CodeToolchainIdentityChanged {
		t.Fatalf("RH10 diagnostic=%q result=%#v err=%v", ErrorCode(err), result, err)
	}
	if provider.processStarts != 0 || result.Execution.SchemaID != "" || result.Publication.SchemaID != "" || result.ArtifactPath != "" {
		t.Fatalf("RH10 crossed process/receipt/publication boundary: starts=%d result=%#v", provider.processStarts, result)
	}
	for _, directory := range []string{"receipts", "blobs"} {
		entries, readErr := os.ReadDir(filepath.Join(storeRoot, directory))
		if readErr != nil || len(entries) != 0 {
			t.Fatalf("RH10 protected %s not empty: entries=%d err=%v", directory, len(entries), readErr)
		}
	}
}

func writeRustFixtureFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, contents := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRustConformanceR04FeatureSelectionKeepsCaptureNeutral(t *testing.T) {
	lock := LockFile{Version: 4, Digest: digest([]byte("lock")), Packages: []LockPackage{{Key: PackageKey{Name: "app", Version: "0.1.0"}, Kind: SourcePath}, {Key: PackageKey{Name: "optional_dep", Version: "1.0.0", Source: "registry+https://example.invalid/index"}, Kind: SourceRegistry}}}
	manifest := Manifest{Path: "workspace/Cargo.toml", PackageName: "app", PackageVersion: "0.1.0", Features: map[string][]string{"extra": {"dep:optional_dep"}}, Dependencies: []DependencyDeclaration{{Name: "optional_dep", Version: "1", Optional: true, DefaultFeatures: true}}, Digest: digest([]byte("manifest"))}
	capture, err := NewCaptureGraph(lock, []Manifest{manifest}, []string{"sha256:" + digest([]byte("artifact"))})
	if err != nil {
		t.Fatal(err)
	}
	rootID := "path+file:///workspace#app@0.1.0"
	depID := "registry+https://example.invalid/index#optional_dep@1.0.0"
	rootPackage := MetadataPackage{ID: rootID, Name: "app", Version: "0.1.0", ManifestPath: "workspace/Cargo.toml", Targets: []MetadataTarget{{Name: "app", Kinds: []string{"bin"}, CrateTypes: []string{"bin"}, SrcPath: "workspace/src/main.rs"}}}
	disabledMetadata := Metadata{Packages: []MetadataPackage{rootPackage}, Resolve: []MetadataNode{{ID: rootID, Features: []string{}}}}
	enabledMetadata := Metadata{Packages: []MetadataPackage{rootPackage, {ID: depID, Name: "optional_dep", Version: "1.0.0", Source: "registry+https://example.invalid/index", ManifestPath: "vendor/optional_dep-1.0.0/Cargo.toml", Targets: []MetadataTarget{{Name: "optional_dep", Kinds: []string{"lib"}, CrateTypes: []string{"lib"}, SrcPath: "vendor/optional_dep-1.0.0/src/lib.rs"}}}}, Resolve: []MetadataNode{{ID: rootID, Features: []string{"extra"}, Dependencies: []MetadataDependency{{ID: depID, Name: "optional_dep", Kind: "normal"}}}, {ID: depID, Features: []string{}}}}
	disabledSelection := SelectionContext{Package: "app", Binary: "app", Target: "aarch64-apple-darwin", DefaultFeatures: true, Features: []string{}, ResolvedFeatures: map[string][]string{rootID: {}}}
	enabledSelection := SelectionContext{Package: "app", Binary: "app", Target: disabledSelection.Target, DefaultFeatures: true, Features: []string{"extra"}, ResolvedFeatures: map[string][]string{rootID: {"extra"}, depID: {}}}
	disabled, err := Reconcile(capture, disabledSelection, disabledMetadata, disabledSelection.Target, string(testBuildID('1')))
	if err != nil {
		t.Fatal(err)
	}
	enabled, err := Reconcile(capture, enabledSelection, enabledMetadata, enabledSelection.Target, string(testBuildID('1')))
	if err != nil {
		t.Fatal(err)
	}
	if disabled.CaptureID != enabled.CaptureID || disabled.Identity == enabled.Identity || bytes.Equal([]byte(disabled.Identity), []byte(enabled.Identity)) {
		t.Fatalf("R04 capture=%s/%s active=%s/%s", disabled.CaptureID, enabled.CaptureID, disabled.Identity, enabled.Identity)
	}
}

func TestRustConformanceR04FeatureSelectedBuildsHaveDistinctIdentities(t *testing.T) {
	workspace := t.TempDir()
	writeRustFixtureFiles(t, workspace, map[string]string{
		"Cargo.toml":              "[package]\nname=\"feature_app\"\nversion=\"0.1.0\"\nedition=\"2021\"\n[dependencies]\noptional_dep={path=\"optional_dep\",optional=true}\n[features]\nextra=[\"dep:optional_dep\"]\n[[bin]]\nname=\"feature-app\"\npath=\"src/main.rs\"\n",
		"Cargo.lock":              "version = 4\n\n[[package]]\nname=\"feature_app\"\nversion=\"0.1.0\"\ndependencies=[\"optional_dep\"]\n\n[[package]]\nname=\"optional_dep\"\nversion=\"0.1.0\"\n",
		"src/main.rs":             "#[cfg(feature=\"extra\")] fn value()->u8{optional_dep::value()} #[cfg(not(feature=\"extra\"))] fn value()->u8{1} fn main(){println!(\"{}\",value());}\n",
		"optional_dep/Cargo.toml": "[package]\nname=\"optional_dep\"\nversion=\"0.1.0\"\nedition=\"2021\"\n",
		"optional_dep/src/lib.rs": "pub fn value()->u8{9}\n",
	})
	manager, err := NewManager(t.Context(), ManagerConfig{WorkRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	capture, err := manager.Capture(t.Context(), RawCaptureRequest{Workspace: RawTree{Root: workspace}, Lock: RawFile{Path: filepath.Join(workspace, "Cargo.lock")}, Manifests: []RawManifest{{Path: "Cargo.toml", File: RawFile{Path: filepath.Join(workspace, "Cargo.toml")}}, {Path: "optional_dep/Cargo.toml", File: RawFile{Path: filepath.Join(workspace, "optional_dep", "Cargo.toml")}}}, Paths: []RawPathOrigin{{DeclaredPath: ".", Tree: RawTree{Root: workspace}}, {DeclaredPath: "optional_dep", Tree: RawTree{Root: filepath.Join(workspace, "optional_dep")}}}})
	if err != nil {
		t.Fatal(err)
	}
	build := func(name string, features []string) (BuildResult, []byte) {
		t.Helper()
		selection := SelectionContext{Package: "feature_app", Binary: "feature-app", Target: manager.state.buildTools.target, DefaultFeatures: true, Features: features, TargetCFG: []string{}}
		metadata, metadataErr := manager.DeriveMetadata(t.Context(), capture, selection)
		if metadataErr != nil {
			t.Fatal(metadataErr)
		}
		selection = metadata.ResolvedSelection()
		publication, binding := rustPublicationFixture(t, manager.state.buildTools, selection)
		result, buildErr := manager.Build(t.Context(), BuildRequest{Capture: capture, Metadata: metadata, Selection: selection, Binding: binding, Publication: publication, StoreRoot: filepath.Join(t.TempDir(), name)})
		if buildErr != nil {
			t.Fatal(buildErr)
		}
		payload, readErr := os.ReadFile(result.ArtifactPath)
		if readErr != nil {
			t.Fatal(readErr)
		}
		return result, payload
	}
	disabled, disabledBytes := build("disabled", []string{})
	enabled, enabledBytes := build("enabled", []string{"extra"})
	if disabled.ActiveGraphID == enabled.ActiveGraphID || disabled.CommandID == enabled.CommandID || bytes.Equal(disabledBytes, enabledBytes) {
		t.Fatalf("R04 identities active=%s/%s command=%s/%s bytes_equal=%v", disabled.ActiveGraphID, enabled.ActiveGraphID, disabled.CommandID, enabled.CommandID, bytes.Equal(disabledBytes, enabledBytes))
	}
}
