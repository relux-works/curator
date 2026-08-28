package rustsource

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
)

func TestBuildToolchainRegistrationStartsNoProcessBeforeC0(t *testing.T) {
	requireNativeCargoDescriptor(t)
	var starts []closureexec.DerivationPermit
	manager, err := NewManager(t.Context(), ManagerConfig{WorkRoot: t.TempDir(), ProcessStartObserver: func(permit closureexec.DerivationPermit) {
		starts = append(starts, permit)
	}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	if len(starts) != 0 {
		t.Fatalf("manager construction/assurance/C0 registration started %d processes", len(starts))
	}
	if _, err = manager.BuildToolchain(); err != nil {
		t.Fatal(err)
	}
	if operation, preflightErr := manager.state.executor.Preflight(t.Context()); preflightErr != nil {
		t.Fatal(preflightErr)
	} else if operation == nil || len(starts) != 0 {
		t.Fatalf("assurance preflight crossed process-start seam: starts=%d", len(starts))
	}
	productionFiles, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	closureFiles, err := filepath.Glob(filepath.Join("..", "closureexec", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	productionFiles = append(productionFiles, closureFiles...)
	sharedProcessSeams := map[string]bool{"acquisition.go": true, "portable_runner.go": true}
	for _, name := range productionFiles {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		payload, readErr := os.ReadFile(name)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if strings.Contains(string(payload), "exec.Command") && !sharedProcessSeams[filepath.Base(name)] {
			t.Fatalf("production discovery/build source %s bypasses the instrumented process boundary", name)
		}
	}
	workspace := t.TempDir()
	writeRustFixtureFiles(t, workspace, map[string]string{
		"Cargo.toml":  "[package]\nname='pre_c0_fixture'\nversion='0.1.0'\nedition='2021'\n",
		"Cargo.lock":  "version = 4\n\n[[package]]\nname='pre_c0_fixture'\nversion='0.1.0'\n",
		"src/main.rs": "fn main(){}\n",
	})
	_, err = manager.Capture(t.Context(), RawCaptureRequest{Workspace: RawTree{Root: workspace}, Lock: RawFile{Path: filepath.Join(workspace, "Cargo.lock")}, Manifests: []RawManifest{{Path: "Cargo.toml", File: RawFile{Path: filepath.Join(workspace, "Cargo.toml")}}}, Paths: []RawPathOrigin{{DeclaredPath: ".", Tree: RawTree{Root: workspace}}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(starts) != 1 || starts[0].C0CheckpointID == "" || starts[0].AssuranceMode != closureexec.AssurancePortable || starts[0].PreviousCausalHead == "" {
		t.Fatalf("first process did not cross the committed permit seam: %#v", starts)
	}
}

func TestCargoRegistrationBindsApprovedExecutableBytes(t *testing.T) {
	requireNativeCargoDescriptor(t)
	registration := registerCargoAtC0(t.Context())
	if registration.err != nil {
		t.Fatal(registration.err)
	}
	if registration.descriptor.ExecutableSHA256 != registration.executableSHA256 || registration.descriptor.Version != "1.91.0" || registration.descriptor.ImplementationCommit != "ea2d97820c16195b0ca3fadb4319fe512c199a43" {
		t.Fatalf("unproved Cargo descriptor: %#v registration=%s", registration.descriptor, registration.executableSHA256)
	}
	registration.descriptor.ExecutableSHA256 = string(testBuildID('f'))
	if _, err := registration.recheck(t.Context()); ErrorCode(err) != CodeVendorTransformUnsupported {
		t.Fatalf("unapproved descriptor error=%v", err)
	}
}

func TestValidateBuildUnitsRejectsEveryUnsupportedActiveUnit(t *testing.T) {
	selection := SelectionContext{Package: "app", Binary: "app"}
	base := Metadata{
		Packages: []MetadataPackage{{ID: "path+file:///app#0.1.0", Name: "app", Version: "0.1.0", Targets: []MetadataTarget{{Name: "app", Kinds: []string{"bin"}, CrateTypes: []string{"bin"}}}}},
		Resolve:  []MetadataNode{{ID: "path+file:///app#0.1.0"}},
	}
	if err := validateBuildUnits(base, selection); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		code Code
		edit func(*Metadata)
	}{
		{name: "RH01_build_script", code: CodeBuildScriptUnsupported, edit: func(value *Metadata) {
			value.Packages[0].Targets = append(value.Packages[0].Targets, MetadataTarget{Name: "build-script-build", Kinds: []string{"custom-build"}})
		}},
		{name: "RH02_build_dependency", code: CodeBuildScriptUnsupported, edit: func(value *Metadata) {
			value.Resolve[0].Dependencies = []MetadataDependency{{ID: value.Resolve[0].ID, Kind: "build"}}
		}},
		{name: "RH03_proc_macro", code: CodeProcMacroUnsupported, edit: func(value *Metadata) {
			value.Packages[0].Targets = append(value.Packages[0].Targets, MetadataTarget{Name: "derive", Kinds: []string{"proc-macro"}})
		}},
		{name: "RH04_native_links", code: CodeNativeLinkUnsupported, edit: func(value *Metadata) { value.Packages[0].Links = "native" }},
		{name: "ambiguous-bin", code: CodeGraphIncomplete, edit: func(value *Metadata) {
			value.Packages[0].Targets = append(value.Packages[0].Targets, value.Packages[0].Targets[0])
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := base
			value.Packages = append([]MetadataPackage(nil), base.Packages...)
			value.Packages[0].Targets = append([]MetadataTarget(nil), base.Packages[0].Targets...)
			value.Resolve = append([]MetadataNode(nil), base.Resolve...)
			test.edit(&value)
			if err := validateBuildUnits(value, selection); ErrorCode(err) != test.code {
				t.Fatalf("code=%q err=%v", ErrorCode(err), err)
			}
		})
	}
}

func TestValidateCargoEventsAcceptsOneSelectedExecutable(t *testing.T) {
	target := t.TempDir()
	executable := filepath.Join(target, "aarch64-apple-darwin", "curator", "app")
	if err := os.MkdirAll(filepath.Dir(executable), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(executable, []byte("locally-built"), 0o500); err != nil {
		t.Fatal(err)
	}
	lines := []map[string]any{
		{"reason": "compiler-artifact", "package_id": "path+file:///dep#0.1.0", "target": map[string]any{"name": "dep", "kind": []string{"lib"}}, "executable": nil},
		{"reason": "compiler-artifact", "package_id": "path+file:///app#app@0.1.0", "target": map[string]any{"name": "app", "kind": []string{"bin"}}, "executable": executable},
		{"reason": "build-finished", "success": true},
	}
	payload := []byte{}
	for _, line := range lines {
		encoded, _ := json.Marshal(line)
		payload = append(payload, append(encoded, '\n')...)
	}
	events, got, err := validateCargoEvents(payload, SelectionContext{Package: "app", Binary: "app"}, MetadataPackage{ID: "path+workspace#app@0.1.0", Name: "app", Version: "0.1.0"}, target, "")
	if err != nil {
		t.Fatal(err)
	}
	if got != executable || len(events) != 3 || !events[1].Executable {
		t.Fatalf("events=%#v executable=%q", events, got)
	}
}

func TestValidateCargoEventsRejectsHooksDuplicatesAndEscapes(t *testing.T) {
	target := t.TempDir()
	outside := filepath.Join(t.TempDir(), "app")
	if err := os.WriteFile(outside, []byte("x"), 0o500); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		lines []map[string]any
		code  Code
	}{
		{name: "hook", lines: []map[string]any{{"reason": "build-script-executed"}}, code: CodeBuildScriptUnsupported},
		{name: "escape", lines: []map[string]any{{"reason": "compiler-artifact", "package_id": "app", "target": map[string]any{"name": "app", "kind": []string{"bin"}}, "executable": outside}}, code: CodeGraphIncomplete},
		{name: "unknown", lines: []map[string]any{{"reason": "future-event"}}, code: CodeGraphIncomplete},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := []byte{}
			for _, line := range test.lines {
				encoded, _ := json.Marshal(line)
				payload = append(payload, append(encoded, '\n')...)
			}
			if _, _, err := validateCargoEvents(payload, SelectionContext{Binary: "app"}, MetadataPackage{ID: "app", Name: "app", Version: "0.1.0"}, target, ""); ErrorCode(err) != test.code {
				t.Fatalf("code=%q err=%v", ErrorCode(err), err)
			}
		})
	}
}

func TestBuildBindingRequiresUniqueNativePlatformAndEveryPhysicalTool(t *testing.T) {
	binding, publication, selection, tools := buildBindingFixture(t)
	if err := validateBuildBinding(binding, publication, selection, tools); err != nil {
		t.Fatal(err)
	}

	t.Run("missing-tool-node", func(t *testing.T) {
		value := binding
		value.Nodes = append([]closuregraph.Node(nil), binding.Nodes[1:]...)
		if code := ErrorCode(validateBuildBinding(value, publication, selection, tools)); code != CodeGraphReferenceInvalid {
			t.Fatalf("code=%q", code)
		}
	})
	t.Run("duplicate-platform", func(t *testing.T) {
		value := binding
		value.Nodes = append(append([]closuregraph.Node(nil), binding.Nodes...), binding.Nodes[len(binding.Nodes)-1])
		if code := ErrorCode(validateBuildBinding(value, publication, selection, tools)); code != CodeGraphReferenceInvalid {
			t.Fatalf("code=%q", code)
		}
	})
	t.Run("fingerprint-drift", func(t *testing.T) {
		value := binding
		value.Nodes = append([]closuregraph.Node(nil), binding.Nodes...)
		for index, node := range value.Nodes {
			if node.Kind == closuregraph.NodeToolchainComponent {
				payload := node.Payload.(closuregraph.ToolchainComponentPayload)
				payload.ContentFingerprint = testBuildID('f')
				node.Payload = payload
				value.Nodes[index] = node
				break
			}
		}
		if code := ErrorCode(validateBuildBinding(value, publication, selection, tools)); code != CodeGraphReferenceInvalid {
			t.Fatalf("code=%q", code)
		}
	})
	t.Run("cross-target", func(t *testing.T) {
		value := selection
		value.Target = "x86_64-unknown-linux-gnu"
		if code := ErrorCode(validateBuildBinding(binding, publication, value, tools)); code != CodeTargetUnsupported {
			t.Fatalf("code=%q", code)
		}
	})
}

func TestRustConformanceR01R08R09RH05RH06RH08AndProtectedCache(t *testing.T) {
	requireNativeCargoDescriptor(t)
	workspace := t.TempDir()
	if err := os.Mkdir(filepath.Join(workspace, "src"), 0o700); err != nil {
		t.Fatal(err)
	}
	manifest := []byte("[package]\nname = \"offline_app\"\nversion = \"0.1.0\"\nedition = \"2021\"\n\n[[bin]]\nname = \"offline-app\"\npath = \"src/main.rs\"\n\n[dependencies]\noffline_dep = \"1.0.0\"\n")
	registrySource := "registry+https://github.com/rust-lang/crates.io-index"
	depKey := PackageKey{Name: "offline_dep", Version: "1.0.0", Source: registrySource}
	leafKey := PackageKey{Name: "offline_leaf", Version: "1.0.0", Source: registrySource}
	archive := crateBytes(t, depKey, map[string][]byte{"Cargo.toml": []byte("[package]\nname=\"offline_dep\"\nversion=\"1.0.0\"\nedition=\"2021\"\n[dependencies]\noffline_leaf=\"1.0.0\"\n"), "src/lib.rs": []byte("pub fn value() -> u8 { offline_leaf::value() }\n")})
	leafArchive := crateBytes(t, leafKey, map[string][]byte{"Cargo.toml": []byte("[package]\nname=\"offline_leaf\"\nversion=\"1.0.0\"\nedition=\"2021\"\n"), "src/lib.rs": []byte("pub fn value() -> u8 { 7 }\n")})
	checksum := digest(archive)
	leafChecksum := digest(leafArchive)
	index, _ := json.Marshal(map[string]any{"name": "offline_dep", "vers": "1.0.0", "cksum": checksum, "deps": []any{map[string]any{"name": "offline_leaf", "req": "^1.0.0"}}})
	leafIndex, _ := json.Marshal(map[string]any{"name": "offline_leaf", "vers": "1.0.0", "cksum": leafChecksum, "deps": []any{}})
	origins := t.TempDir()
	cratePath, indexPath := filepath.Join(origins, "offline_dep.crate"), filepath.Join(origins, "offline_dep.index.json")
	leafCratePath, leafIndexPath := filepath.Join(origins, "offline_leaf.crate"), filepath.Join(origins, "offline_leaf.index.json")
	if err := os.WriteFile(cratePath, archive, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(indexPath, index, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(leafCratePath, leafArchive, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(leafIndexPath, leafIndex, 0o600); err != nil {
		t.Fatal(err)
	}
	lock := []byte("version = 4\n\n[[package]]\nname = \"offline_app\"\nversion = \"0.1.0\"\ndependencies = [\"offline_dep\"]\n\n[[package]]\nname = \"offline_dep\"\nversion = \"1.0.0\"\nsource = \"" + registrySource + "\"\nchecksum = \"" + checksum + "\"\ndependencies = [\"offline_leaf\"]\n\n[[package]]\nname = \"offline_leaf\"\nversion = \"1.0.0\"\nsource = \"" + registrySource + "\"\nchecksum = \"" + leafChecksum + "\"\n")
	for path, payload := range map[string][]byte{
		filepath.Join(workspace, "Cargo.toml"):     manifest,
		filepath.Join(workspace, "Cargo.lock"):     lock,
		filepath.Join(workspace, "src", "main.rs"): []byte("fn main() { println!(\"offline-ok-{}\", offline_dep::value()); }\n"),
	} {
		if err := os.WriteFile(path, payload, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	manager, err := NewManager(t.Context(), ManagerConfig{WorkRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	captureRequest := RawCaptureRequest{Workspace: RawTree{Root: workspace}, Lock: RawFile{Path: filepath.Join(workspace, "Cargo.lock")}, Manifests: []RawManifest{{Path: "Cargo.toml", File: RawFile{Path: filepath.Join(workspace, "Cargo.toml")}}}, Registry: []RawRegistryOrigin{{SourceLocator: registrySource, IndexRecord: RawFile{Path: indexPath}, CrateArchive: RawFile{Path: cratePath}}, {SourceLocator: registrySource, IndexRecord: RawFile{Path: leafIndexPath}, CrateArchive: RawFile{Path: leafCratePath}}}, Paths: []RawPathOrigin{{DeclaredPath: ".", Tree: RawTree{Root: workspace}}}}
	capture, err := manager.Capture(t.Context(), captureRequest)
	if err != nil {
		t.Fatal(err)
	}
	selection := SelectionContext{Package: "offline_app", Binary: "offline-app", Target: manager.state.buildTools.target, DefaultFeatures: true, Features: []string{}, TargetCFG: []string{}}
	metadata, err := manager.DeriveMetadata(t.Context(), capture, selection)
	if err != nil {
		t.Fatal(err)
	}
	selection = metadata.ResolvedSelection()
	publication, binding := rustPublicationFixture(t, manager.state.buildTools, selection)
	processStarts := 0
	manager.state.processRunner.ProcessStartObserver = func(permit closureexec.DerivationPermit) {
		if permit.C0CheckpointID == "" || permit.AssuranceMode == "" || permit.PreviousCausalHead == "" {
			t.Fatalf("process crossed launch seam without committed C0/assurance authority: %#v", permit)
		}
		processStarts++
	}
	storeRoot := filepath.Join(t.TempDir(), "protected")
	planted := filepath.Join(manager.state.session, "build-metadata-output")
	if err := os.MkdirAll(planted, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planted, "planted-binary"), []byte("compiled"), 0o500); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Build(t.Context(), BuildRequest{Capture: capture, Metadata: metadata, Selection: selection, Binding: binding, Publication: publication, StoreRoot: storeRoot}); ErrorCode(err) != CodeLocalOutputUnreceipted {
		t.Fatalf("RH08 planted output error=%v", err)
	}
	if processStarts != 0 {
		t.Fatalf("RH08 planted output crossed process seam: %d", processStarts)
	}
	if entries, readErr := os.ReadDir(filepath.Join(storeRoot, "receipts")); readErr != nil || len(entries) != 0 {
		t.Fatalf("RH08 planted output published: entries=%d err=%v", len(entries), readErr)
	}
	if err := os.RemoveAll(planted); err != nil {
		t.Fatal(err)
	}
	result, err := manager.Build(t.Context(), BuildRequest{Capture: capture, Metadata: metadata, Selection: selection, Binding: binding, Publication: publication, StoreRoot: storeRoot})
	if err != nil {
		t.Fatal(err)
	}
	if result.CacheHit || processStarts != 2 || result.ArtifactPath == "" || result.Publication.Decision != "published" {
		t.Fatalf("result=%#v starts=%d", result, processStarts)
	}
	payload, err := exec.Command(result.ArtifactPath).Output() // #nosec G204 -- protected test artifact built from the fixture above.
	if err != nil || string(payload) != "offline-ok-7\n" {
		t.Fatalf("execute: %v %q", err, payload)
	}
	second, err := manager.Build(t.Context(), BuildRequest{Capture: capture, Metadata: metadata, Selection: selection, Binding: binding, Publication: publication, StoreRoot: storeRoot})
	if err != nil {
		t.Fatal(err)
	}
	if !second.CacheHit || processStarts != 2 || second.ArtifactPath != result.ArtifactPath {
		t.Fatalf("cache result=%#v starts=%d", second, processStarts)
	}
	secondManager, err := NewManager(t.Context(), ManagerConfig{WorkRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = secondManager.Close() })
	secondCapture, err := secondManager.Capture(t.Context(), captureRequest)
	if err != nil {
		t.Fatal(err)
	}
	secondMetadata, err := secondManager.DeriveMetadata(t.Context(), secondCapture, SelectionContext{Package: "offline_app", Binary: "offline-app", Target: secondManager.state.buildTools.target, DefaultFeatures: true, Features: []string{}, TargetCFG: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	secondSelection := secondMetadata.ResolvedSelection()
	secondPublication, secondBinding := rustPublicationFixture(t, secondManager.state.buildTools, secondSelection)
	cleanRebuild, err := secondManager.Build(t.Context(), BuildRequest{Capture: secondCapture, Metadata: secondMetadata, Selection: secondSelection, Binding: secondBinding, Publication: secondPublication, StoreRoot: filepath.Join(t.TempDir(), "second-protected")})
	if err != nil {
		t.Fatal(err)
	}
	firstBytes, err := os.ReadFile(result.ArtifactPath)
	if err != nil {
		t.Fatal(err)
	}
	secondBytes, err := os.ReadFile(cleanRebuild.ArtifactPath)
	if err != nil {
		t.Fatal(err)
	}
	if capture.state.graph.Identity != secondCapture.state.graph.Identity || result.ActiveGraphID != cleanRebuild.ActiveGraphID || result.CommandID != cleanRebuild.CommandID || !bytes.Equal(firstBytes, secondBytes) {
		t.Fatalf("R09 clean rebuild drift: capture=%s/%s active=%s/%s command=%s/%s bytes_equal=%v", capture.state.graph.Identity, secondCapture.state.graph.Identity, result.ActiveGraphID, cleanRebuild.ActiveGraphID, result.CommandID, cleanRebuild.CommandID, bytes.Equal(firstBytes, secondBytes))
	}
	missingStarts := processStarts
	_, err = manager.state.executor.Execute(t.Context(), testBuildID('7'), nil, nil)
	var missing *closureexec.DiagnosticError
	if !errors.As(err, &missing) || missing.Code != "closure_derivation_unauthorized" || processStarts != missingStarts {
		t.Fatalf("missing permit error=%v starts=%d", err, processStarts)
	}
	if err := os.Chmod(result.ArtifactPath, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Build(t.Context(), BuildRequest{Capture: capture, Metadata: metadata, Selection: selection, Binding: binding, Publication: publication, StoreRoot: storeRoot}); err == nil {
		t.Fatal("tampered protected cache entry was reused or overwritten")
	}
	if processStarts != 2 {
		t.Fatalf("Cargo restarted after protected-cache drift: %d", processStarts)
	}
	provider := newRustBuildProvider()
	enableVerifiedBuild(t, manager, provider)
	for _, vector := range []struct{ id, attempt string }{{"RH05_include_escape", "read"}, {"RH06_hook_network", "network"}, {"RH06_hook_child", "child"}, {"RH06_hook_write", "write"}} {
		t.Run(vector.id, func(t *testing.T) {
			provider.attempt = vector.attempt
			root := filepath.Join(t.TempDir(), "rejected-"+vector.attempt)
			_, buildErr := manager.Build(t.Context(), BuildRequest{Capture: capture, Metadata: metadata, Selection: selection, Binding: binding, Publication: publication, StoreRoot: root})
			if ErrorCode(buildErr) != CodeUndeclaredInput {
				t.Fatalf("%s enforcement error=%v", vector.attempt, buildErr)
			}
			if len(provider.attempts) == 0 || provider.attempts[len(provider.attempts)-1] != vector.attempt {
				t.Fatalf("%s attempt did not reach verified enforcement seam: %#v", vector.attempt, provider.attempts)
			}
			entries, readErr := os.ReadDir(filepath.Join(root, "receipts"))
			if readErr != nil || len(entries) != 0 {
				t.Fatalf("%s published before enforcement rejection: entries=%d err=%v", vector.attempt, len(entries), readErr)
			}
		})
	}
	provider.attempt = ""
	provider.mutateAudit = "evidence"
	auditRoot := filepath.Join(t.TempDir(), "rejected-audit-mismatch")
	_, err = manager.Build(t.Context(), BuildRequest{Capture: capture, Metadata: metadata, Selection: selection, Binding: binding, Publication: publication, StoreRoot: auditRoot})
	var auditMismatch *closureexec.DiagnosticError
	if !errors.As(err, &auditMismatch) || auditMismatch.Code != "closure_derivation_drift" {
		t.Fatalf("audit mismatch error=%v", err)
	}
	provider.mutateAudit = ""
	provider.mutateInput = true
	root := filepath.Join(t.TempDir(), "rejected-input-mutation")
	_, err = manager.Build(t.Context(), BuildRequest{Capture: capture, Metadata: metadata, Selection: selection, Binding: binding, Publication: publication, StoreRoot: root})
	var mutation *closureexec.DiagnosticError
	if !errors.As(err, &mutation) || mutation.Code != "closure_derivation_drift" {
		t.Fatalf("input mutation error=%v", err)
	}
	entries, readErr := os.ReadDir(filepath.Join(root, "receipts"))
	if readErr != nil || len(entries) != 0 {
		t.Fatalf("input mutation published: entries=%d err=%v", len(entries), readErr)
	}
}

func TestVerifiedManagerWithoutProviderFailsBeforeSessionOrProcess(t *testing.T) {
	root := t.TempDir()
	identity := newRustBuildProvider().Identity()
	_, err := NewManager(t.Context(), ManagerConfig{WorkRoot: root, Assurance: verifiedRustBuildConfig(identity)})
	var diagnostic *closureexec.DiagnosticError
	if !errors.As(err, &diagnostic) || diagnostic.Code != "verified_provider_missing" {
		t.Fatalf("verified unavailable error=%v", err)
	}
	entries, readErr := os.ReadDir(root)
	if readErr != nil || len(entries) != 0 {
		t.Fatalf("verified-unavailable manager created session/process state: entries=%d err=%v", len(entries), readErr)
	}
}

func TestPathOnlyClosureDerivesFreshFrozenMetadata(t *testing.T) {
	requireNativeCargoDescriptor(t)
	workspace := t.TempDir()
	if err := os.Mkdir(filepath.Join(workspace, "src"), 0o700); err != nil {
		t.Fatal(err)
	}
	for path, payload := range map[string][]byte{
		filepath.Join(workspace, "Cargo.toml"):     []byte("[package]\nname=\"path_only\"\nversion=\"0.1.0\"\nedition=\"2021\"\n"),
		filepath.Join(workspace, "Cargo.lock"):     []byte("version = 4\n\n[[package]]\nname = \"path_only\"\nversion = \"0.1.0\"\n"),
		filepath.Join(workspace, "src", "main.rs"): []byte("fn main() {}\n"),
	} {
		if err := os.WriteFile(path, payload, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	manager, err := NewManager(t.Context(), ManagerConfig{WorkRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	capture, err := manager.Capture(t.Context(), RawCaptureRequest{Workspace: RawTree{Root: workspace}, Lock: RawFile{Path: filepath.Join(workspace, "Cargo.lock")}, Manifests: []RawManifest{{Path: "Cargo.toml", File: RawFile{Path: filepath.Join(workspace, "Cargo.toml")}}}, Paths: []RawPathOrigin{{DeclaredPath: ".", Tree: RawTree{Root: workspace}}}})
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := manager.DeriveMetadata(t.Context(), capture, SelectionContext{Package: "path_only", Binary: "path_only", Target: manager.state.buildTools.target, DefaultFeatures: true, Features: []string{}, TargetCFG: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(metadata.Active.Packages) != 1 || metadata.Active.Packages[0].ManifestPath != "workspace/Cargo.toml" {
		t.Fatalf("metadata=%#v", metadata.Active)
	}
}

type rustBuildProvider struct {
	identity           closureexec.ProviderIdentity
	starts             int
	processStarts      int
	negotiations       int
	driftOnNegotiation int
	driftExecutable    string
	attempt            string
	attempts           []string
	mutateAudit        string
	mutateInput        bool
	observedAt         time.Time
	requests           []closureexec.ExecutionRequest
}

func newRustBuildProvider() *rustBuildProvider {
	return &rustBuildProvider{observedAt: time.Now().Add(-time.Second), identity: closureexec.ProviderIdentity{
		Contract: closureexec.VerifiedProviderContractID, ProviderID: "rust-build-test-provider", Version: "1.0.0",
		BinarySHA256: testBuildID('8'), TrustEvidence: "test-provider-identity",
	}}
}

func (provider *rustBuildProvider) Identity() closureexec.ProviderIdentity { return provider.identity }
func (*rustBuildProvider) LosslessObservation() bool                       { return true }
func (provider *rustBuildProvider) Negotiate(_ context.Context, nonce string) (closureexec.ProviderCapabilityReceipt, error) {
	provider.negotiations++
	if provider.driftOnNegotiation == provider.negotiations {
		if err := os.Chmod(provider.driftExecutable, 0o600); err != nil {
			return closureexec.ProviderCapabilityReceipt{}, err
		}
		if err := os.WriteFile(provider.driftExecutable, []byte("drifted-cargo"), 0o600); err != nil {
			return closureexec.ProviderCapabilityReceipt{}, err
		}
	}
	return closureexec.ProviderCapabilityReceipt{Provider: provider.identity, Health: "healthy", Nonce: nonce, ObservedAt: provider.observedAt, ExpiresAt: provider.observedAt.Add(time.Hour), Capabilities: verifiedRustCapabilities()}, nil
}

func (provider *rustBuildProvider) EnforceAndObserve(ctx context.Context, request closureexec.ExecutionRequest) (closureexec.Audit, error) {
	provider.starts++
	provider.requests = append(provider.requests, request)
	if provider.attempt != "" {
		provider.attempts = append(provider.attempts, provider.attempt)
		return closureexec.Audit{}, verifiedAttemptDiagnostic(provider.attempt)
	}
	root := request.Permit.Environment["CURATOR_EXECUTION_ROOT"]
	outputRoot := request.Permit.Environment["CURATOR_OUTPUT_ROOT"]
	evidenceRoot := request.Permit.Environment["CURATOR_EVIDENCE_ROOT"]
	if root == "" || outputRoot == "" || evidenceRoot != root {
		return closureexec.Audit{}, fmt.Errorf("test provider roots are absent")
	}
	executable := filepath.Join(root, filepath.FromSlash(request.Permit.Executable))
	executableBytes, err := os.ReadFile(executable) // #nosec G304 -- exact committed executable below the test execution root.
	if err != nil {
		return closureexec.Audit{}, err
	}
	executableSum := sha256.Sum256(executableBytes)
	if closuregraph.ID("sha256:"+hex.EncodeToString(executableSum[:])) != request.Permit.ExecutableSHA256 {
		return closureexec.Audit{}, &closureexec.DiagnosticError{Code: "artifact_toolchain_identity_changed", Detail: "verified fixture observed executable drift before process start"}
	}
	for _, input := range request.Inputs {
		source, err := input.ProtectedPath()
		if err != nil {
			return closureexec.Audit{}, err
		}
		if err = copyProviderTree(source, filepath.Join(root, filepath.FromSlash(input.MountPath))); err != nil {
			return closureexec.Audit{}, err
		}
	}
	for _, work := range request.Permit.WorkCopies {
		var source string
		for _, input := range request.Inputs {
			if input.ReceiptID == work.ReceiptID {
				source = filepath.Join(root, filepath.FromSlash(input.MountPath))
				break
			}
		}
		if source == "" {
			return closureexec.Audit{}, fmt.Errorf("work-copy source is absent")
		}
		if err := copyProviderTree(source, filepath.Join(root, filepath.FromSlash(work.Path))); err != nil {
			return closureexec.Audit{}, err
		}
	}
	if err := os.MkdirAll(outputRoot, 0o700); err != nil {
		return closureexec.Audit{}, err
	}
	command := exec.CommandContext(ctx, executable, request.Permit.Argv...) // #nosec G204 -- test executes the committed manager permit.
	command.Dir = filepath.Join(root, filepath.FromSlash(request.Permit.CWD))
	keys := make([]string, 0, len(request.Permit.Environment))
	for key := range request.Permit.Environment {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		command.Env = append(command.Env, key+"="+request.Permit.Environment[key])
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	provider.processStarts++
	if err := command.Run(); err != nil {
		return closureexec.Audit{}, fmt.Errorf("cargo test provider: %w: %s", err, stderr.String())
	}
	if request.Permit.StdoutEvidencePath != "" {
		path := filepath.Join(evidenceRoot, filepath.FromSlash(request.Permit.StdoutEvidencePath))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return closureexec.Audit{}, err
		}
		if err := os.WriteFile(path, stdout.Bytes(), 0o600); err != nil {
			return closureexec.Audit{}, err
		}
	}
	outputs := make([]closureexec.DerivationOutput, len(request.Permit.ExpectedEvidence))
	for index, expected := range request.Permit.ExpectedEvidence {
		path := filepath.Join(evidenceRoot, filepath.FromSlash(expected.Path))
		payload, err := os.ReadFile(path) // #nosec G304 -- exact committed evidence path in a private test root.
		if err != nil {
			return closureexec.Audit{}, err
		}
		escrow := filepath.Join(outputRoot, filepath.FromSlash(expected.Path))
		if err = os.MkdirAll(filepath.Dir(escrow), 0o700); err != nil {
			return closureexec.Audit{}, err
		}
		if err = os.WriteFile(escrow, payload, 0o600); err != nil {
			return closureexec.Audit{}, err
		}
		sum := sha256.Sum256(payload)
		outputs[index] = closureexec.DerivationOutput{Path: expected.Path, SchemaID: expected.SchemaID, ArtifactManifestID: expected.ArtifactManifestID, SHA256: closuregraph.ID("sha256:" + hex.EncodeToString(sum[:])), Size: int64(len(payload))}
	}
	if provider.mutateInput && len(request.Inputs) > 0 {
		mutated := false
		for _, input := range request.Inputs {
			path, _ := input.ProtectedPath()
			_ = filepath.WalkDir(path, func(current string, entry fs.DirEntry, walkErr error) error {
				if walkErr == nil && !entry.IsDir() {
					_ = os.Chmod(current, 0o600)
					_ = os.WriteFile(current, []byte("mutated"), 0o600)
					mutated = true
					return fs.SkipAll
				}
				return nil
			})
			if mutated {
				break
			}
		}
	}
	audit := closureexec.Audit{Executable: request.Permit.Executable, CWD: request.Permit.CWD, Argv: append([]string(nil), request.Permit.Argv...), Environment: cloneTestMap(request.Permit.Environment), Processes: append([]string(nil), request.Permit.AllowedProcesses...), Reads: append([]string(nil), request.Permit.ReadRoots...), Writes: append([]string(nil), request.Permit.WriteRoots...), Evidence: make([]string, len(request.Permit.ExpectedEvidence)), Network: "none", ExitCode: 0, Outputs: outputs}
	for index := range request.Permit.ExpectedEvidence {
		audit.Evidence[index] = request.Permit.ExpectedEvidence[index].Path
	}
	switch provider.mutateAudit {
	case "network":
		audit.Network = "attempted"
	case "child":
		audit.Processes = append(audit.Processes, "undeclared/child")
	case "read":
		audit.Reads = append(audit.Reads, "undeclared/read")
	case "write":
		audit.Writes = append(audit.Writes, "undeclared/write")
	case "evidence":
		audit.Evidence = append(audit.Evidence, "undeclared/evidence")
	}
	return audit, nil
}

func verifiedAttemptDiagnostic(operation string) error {
	// The injected verified provider models the syscall/event intercepted by a
	// lossless boundary before it can take effect. Unlike audit mutation after a
	// successful Cargo run, each record is the attempted operation itself.
	switch operation {
	case "network":
		return &closureexec.DiagnosticError{Code: "closure_network_attempted", Detail: "connect denied by verified fixture boundary"}
	case "child":
		return &closureexec.DiagnosticError{Code: "closure_process_undeclared", Detail: "undeclared exec denied by verified fixture boundary"}
	case "read":
		return &closureexec.DiagnosticError{Code: "closure_input_undeclared", Detail: "host read denied by verified fixture boundary"}
	case "write":
		return &closureexec.DiagnosticError{Code: "closure_write_undeclared", Detail: "out-of-root write denied by verified fixture boundary"}
	default:
		return fmt.Errorf("unknown verified fixture operation %q", operation)
	}
}

func enableVerifiedBuild(t *testing.T, manager *Manager, provider *rustBuildProvider) {
	t.Helper()
	executor, err := closureexec.NewAssuredExecutor(verifiedRustBuildConfig(provider.Identity()), nil, provider, manager.state.causalHead)
	if err != nil {
		t.Fatal(err)
	}
	manager.state.executor = executor
	manager.state.buildOperation = nil
}

func verifiedRustBuildConfig(identity closureexec.ProviderIdentity) closureexec.AssuranceConfig {
	return closureexec.AssuranceConfig{Mode: closureexec.AssuranceVerified, ProviderID: identity.ProviderID, ProviderVersion: identity.Version, ProviderBinarySHA256: identity.BinarySHA256, ProviderTrustEvidence: identity.TrustEvidence}
}

func verifiedRustCapabilities() []closureexec.CapabilityEvidence {
	return []closureexec.CapabilityEvidence{{CapabilityID: "total-network-denial-v1", Status: "established"}, {CapabilityID: "read-only-source-and-toolchain-v1", Status: "established"}, {CapabilityID: "exact-executable-allowlisting-v1", Status: "established"}, {CapabilityID: "private-build-root-only-writes-v1", Status: "established"}, {CapabilityID: "hard-aggregate-descendant-resource-bounds-v1", Status: "established"}, {CapabilityID: "fail-closed-capability-preflight-v1", Status: "established"}}
}

func copyProviderTree(source, target string) error {
	if err := os.RemoveAll(target); err != nil {
		return err
	}
	return filepath.WalkDir(source, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, current)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, rel)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o700)
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir member below protected test input.
		if err != nil {
			return err
		}
		return os.WriteFile(destination, payload, 0o600)
	})
}

func cloneTestMap(input map[string]string) map[string]string {
	result := make(map[string]string, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func rustPublicationFixture(t *testing.T, tools rustBuildToolchain, rustSelection SelectionContext) (closuregraph.PublicationEvidence, BuildBinding) {
	t.Helper()
	profile := ProfileID
	product := closuregraph.Node{Kind: closuregraph.NodeCommandProduct, LogicalKey: "product:offline-app", Payload: closuregraph.CommandProductPayload{Profile: profile, SkillKey: "fixture", CommandKey: rustSelection.Binary, EntryPointContract: "native_command", DeclarationDigest: testBuildID('a')}}
	source := closuregraph.Node{Kind: closuregraph.NodeSourceSet, LogicalKey: "source:workspace", Payload: closuregraph.SourceSetPayload{Profile: profile, Origin: "fixture://workspace", TrustRole: "dependency_input", ArtifactManifestID: testBuildID('1'), Grammar: "rust-source-v1", Projection: []string{"workspace"}}}
	toolSlots := make([]string, len(requiredBuildToolRoles))
	argv := []string{}
	for index, role := range requiredBuildToolRoles {
		toolSlots[index] = string(role)
		argv = append(argv, "$TOOL("+string(role)+")")
	}
	argv = append(argv, "$READ(source)", "$WRITE(binary)")
	action := closuregraph.Node{Kind: closuregraph.NodeAction, LogicalKey: "action:cargo-build", Payload: closuregraph.ActionPayload{Profile: profile, ActionSubtype: "compiler", ExecutionDomain: closuregraph.ExecutionTarget, ArgvTemplate: argv, EnvironmentPolicyID: "rust-build-environment-v1", ProcessPolicyID: "rust-build-process-v1", Network: "none", ToolSlotNames: toolSlots, ReadSlotNames: []string{"source"}, WriteSlotNames: []string{"binary"}}}
	output := closuregraph.Node{Kind: closuregraph.NodeOutputArtifact, LogicalKey: "output:offline-app", Payload: closuregraph.OutputArtifactPayload{Profile: profile, LogicalPath: "bin/" + rustSelection.Binary, ExpectedClass: "native.executable", OutputRole: "published_command", PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformTarget}}}
	productID, sourceID, actionID, outputID := mustNodeID(product), mustNodeID(source), mustNodeID(action), mustNodeID(output)
	captureNodes := []closuregraph.Node{product, source, action, output}
	captureEdges := []closuregraph.Edge{
		{Kind: closuregraph.EdgeDeclares, EdgeKey: "declares:build", FromNodeID: productID, ToNodeID: actionID, Payload: closuregraph.DeclaresPayload{Origin: closuregraph.EvidenceOrigin{Field: "fixture.build"}}},
		{Kind: closuregraph.EdgeReads, EdgeKey: "reads:source", FromNodeID: actionID, ToNodeID: sourceID, Payload: closuregraph.ReadsPayload{Path: "workspace", ReadSlot: "source", ReadClass: "rust-source"}},
		{Kind: closuregraph.EdgeProduces, EdgeKey: "produces:binary", FromNodeID: actionID, ToNodeID: outputID, Payload: closuregraph.ProducesPayload{Path: "bin/" + rustSelection.Binary, WriteSlot: "binary", WriteClass: "native.executable"}},
		{Kind: closuregraph.EdgePublishes, EdgeKey: "publishes:binary", FromNodeID: productID, ToNodeID: outputID, Payload: closuregraph.PublishesPayload{Destination: "bin/" + rustSelection.Binary, EntryPoint: rustSelection.Binary}},
	}
	capture, err := closuregraph.NewCaptureGraph(profile, []string{"rust-source-policy-v1"}, []closuregraph.ID{productID}, captureNodes, captureEdges, []closuregraph.ID{testBuildID('1')})
	if err != nil {
		t.Fatal(err)
	}
	platform := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "platform:" + tools.target, Payload: nativePlatformPayload(tools.target)}
	platformID := mustNodeID(platform)
	selection, err := closuregraph.NewSelectionContext([]closuregraph.ID{productID}, map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID}, rustSelection.Features, rustSelection.DefaultFeatures, map[string]string{}, map[string]string{}, []string{})
	if err != nil {
		t.Fatal(err)
	}
	bindingNodes := []closuregraph.Node{platform}
	bindingEdges := []closuregraph.Edge{}
	authority := closuregraph.BindingAuthority{}
	origin := closuregraph.EvidenceOrigin{Field: "selection.platform_roles.target"}
	for _, evidence := range tools.evidence() {
		node := closuregraph.Node{Kind: closuregraph.NodeToolchainComponent, LogicalKey: "tool:" + string(evidence.Role), Payload: closuregraph.ToolchainComponentPayload{ComponentRole: string(evidence.Role), ContentFingerprint: evidence.ContentFingerprint, ExecutableRelativePath: evidence.ExecutableRelativePath, PlatformABI: tools.target, PolicySelector: profile, VersionOutput: evidence.VersionOutput, TimeOfUseRecheckRule: "immediate-exact-v1", ExecutionDomain: closuregraph.ExecutionTarget, PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformTarget}}}
		id := mustNodeID(node)
		bindingNodes = append(bindingNodes, node)
		bindingEdges = append(bindingEdges,
			closuregraph.Edge{Kind: closuregraph.EdgeUsesTool, EdgeKey: "uses:" + string(evidence.Role), FromNodeID: actionID, ToNodeID: id, Payload: closuregraph.UsesToolPayload{ExecutableRelativePath: evidence.ExecutableRelativePath, ToolSlot: string(evidence.Role), InvocationRole: "build"}},
			closuregraph.Edge{Kind: closuregraph.EdgeTargets, EdgeKey: "targets:tool:" + string(evidence.Role), FromNodeID: id, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: origin}},
		)
		selector, err := closuregraph.NewToolchainSelector(node)
		if err != nil {
			t.Fatal(err)
		}
		selectorID, _ := selector.ID()
		authority.Toolchains = append(authority.Toolchains, closuregraph.ToolchainBindingEvidence{NodeID: id, FirstBound: closuregraph.ToolchainBoundAtC4, EvidenceID: selectorID})
		authority.C4Selectors = append(authority.C4Selectors, selector)
	}
	for key, id := range map[string]closuregraph.ID{"action": actionID, "output": outputID, "product": productID} {
		bindingEdges = append(bindingEdges, closuregraph.Edge{Kind: closuregraph.EdgeTargets, EdgeKey: "targets:" + key, FromNodeID: id, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: origin}})
	}
	captureID, _ := capture.ID()
	selectionID, _ := selection.ID()
	binding, err := closuregraph.NewSelectionBinding(captureID, selectionID, bindingNodes, bindingEdges)
	if err != nil {
		t.Fatal(err)
	}
	records := closuregraph.NewRecordTables(captureNodes, captureEdges, bindingNodes, bindingEdges)
	graph, err := closuregraph.ProjectActive(capture, selection, binding, records, authority, nil)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := closuregraph.DeriveBuildPlan(graph, closuregraph.PlanOptions{ExecutionPolicyID: "rust-source-manager-worker-v1"})
	if err != nil {
		t.Fatal(err)
	}
	activeID, _ := graph.Active.ID()
	bindingID, _ := binding.ID()
	previous := testBuildID('9')
	c4 := closuregraph.Checkpoint{SchemaID: closuregraph.SchemaCheckpoint, Name: closuregraph.CheckpointC4, PreviousCheckpointID: &previous, Payload: closuregraph.C4ClosePayload{ActiveGraphID: activeID, CapturedGraphID: captureID, SelectionBindingID: bindingID, SelectionContextID: selectionID}, Decision: closuregraph.DecisionAdmit, Diagnostics: []closuregraph.Diagnostic{}}
	if err = c4.Validate(); err != nil {
		t.Fatal(err)
	}
	planID, _ := plan.ID()
	c5, err := closuregraph.NewCheckpoint(closuregraph.C5PlanPayload{BuildPlanID: planID}, &c4, []closuregraph.Diagnostic{})
	if err != nil {
		t.Fatal(err)
	}
	closure, err := closuregraph.NewSourceClosure(c5)
	if err != nil {
		t.Fatal(err)
	}
	publication := closuregraph.PublicationEvidence{C4: c4, C5: c5, Graph: graph, Plan: plan, Closure: closure}
	return publication, BuildBinding{C4: c4, Selection: binding, Nodes: bindingNodes, Edges: bindingEdges, ProductNodeID: productID, ActionNodeID: actionID}
}

func nativePlatformPayload(target string) closuregraph.TargetPlatformPayload {
	parts := strings.Split(target, "-")
	architecture := parts[0]
	osName, abi, libc, sdk := "linux", "gnu", "glibc", "linux-sysroot"
	if strings.Contains(target, "apple-darwin") {
		osName, abi, libc, sdk = "darwin", "darwin", "system", "macos-sdk"
	}
	return closuregraph.TargetPlatformPayload{OS: osName, Architecture: architecture, ABI: abi, Libc: libc, MinimumRuntime: "native", SDKID: sdk, TargetTriple: target, Runtime: "native", LanguageModes: map[string]string{"rust": "1.91.0"}, Tuning: map[string]string{}}
}

func buildBindingFixture(t *testing.T) (BuildBinding, closuregraph.PublicationEvidence, SelectionContext, rustBuildToolchain) {
	t.Helper()
	selection := SelectionContext{Package: "app", Binary: "app", Target: "aarch64-apple-darwin", DefaultFeatures: true}
	tools := rustBuildToolchain{target: selection.Target, items: map[BuildToolRole]BuildToolEvidence{}}
	nodes := []closuregraph.Node{}
	for index, role := range requiredBuildToolRoles {
		evidence := BuildToolEvidence{Role: role, PhysicalPath: "/tool/" + string(role), ExecutableRelativePath: "tool/" + string(role), ContentFingerprint: testBuildID(byte('1' + index)), VersionOutput: string(role) + " 1"}
		tools.items[role] = evidence
		nodes = append(nodes, closuregraph.Node{Kind: closuregraph.NodeToolchainComponent, LogicalKey: "tool:" + string(role), Payload: closuregraph.ToolchainComponentPayload{ComponentRole: string(role), ContentFingerprint: evidence.ContentFingerprint, ExecutableRelativePath: evidence.ExecutableRelativePath, PlatformABI: "native", PolicySelector: ProfileID, VersionOutput: evidence.VersionOutput, TimeOfUseRecheckRule: "immediate-exact-v1"}})
	}
	platform := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "platform:native", Payload: closuregraph.TargetPlatformPayload{OS: "darwin", Architecture: "aarch64", ABI: "darwin", Libc: "system", MinimumRuntime: "native", SDKID: "macos-sdk", TargetTriple: selection.Target}}
	nodes = append(nodes, platform)
	platformID := mustNodeID(platform)
	productID, actionID := testBuildID('a'), testBuildID('b')
	edges := []closuregraph.Edge{}
	origin := closuregraph.EvidenceOrigin{Field: "selection.platform_roles.target"}
	for _, node := range nodes[:len(nodes)-1] {
		id := mustNodeID(node)
		payload := node.Payload.(closuregraph.ToolchainComponentPayload)
		role := BuildToolRole(payload.ComponentRole)
		edges = append(edges,
			closuregraph.Edge{Kind: closuregraph.EdgeUsesTool, EdgeKey: "uses:" + string(role), FromNodeID: actionID, ToNodeID: id, Payload: closuregraph.UsesToolPayload{ExecutableRelativePath: payload.ExecutableRelativePath, ToolSlot: string(role)}},
			closuregraph.Edge{Kind: closuregraph.EdgeTargets, EdgeKey: "targets:" + string(role), FromNodeID: id, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: origin}},
		)
	}
	edges = append(edges,
		closuregraph.Edge{Kind: closuregraph.EdgeTargets, EdgeKey: "targets:product", FromNodeID: productID, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: origin}},
		closuregraph.Edge{Kind: closuregraph.EdgeTargets, EdgeKey: "targets:action", FromNodeID: actionID, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: origin}},
	)
	captureID, selectionID, activeID := testBuildID('c'), testBuildID('d'), testBuildID('e')
	selectionBinding, err := closuregraph.NewSelectionBinding(captureID, selectionID, nodes, edges)
	if err != nil {
		t.Fatal(err)
	}
	selectionBindingID, _ := selectionBinding.ID()
	previous := testBuildID('9')
	c4 := closuregraph.Checkpoint{SchemaID: closuregraph.SchemaCheckpoint, Name: closuregraph.CheckpointC4, PreviousCheckpointID: &previous, Payload: closuregraph.C4ClosePayload{ActiveGraphID: activeID, CapturedGraphID: captureID, SelectionBindingID: selectionBindingID, SelectionContextID: selectionID}, Decision: closuregraph.DecisionAdmit, Diagnostics: []closuregraph.Diagnostic{}}
	if err := c4.Validate(); err != nil {
		t.Fatal(err)
	}
	binding := BuildBinding{C4: c4, Selection: selectionBinding, Nodes: nodes, Edges: edges, ProductNodeID: productID, ActionNodeID: actionID}
	return binding, closuregraph.PublicationEvidence{C4: c4}, selection, tools
}

func testBuildID(value byte) closuregraph.ID {
	return closuregraph.ID("sha256:" + strings.Repeat(string(value), 64))
}
