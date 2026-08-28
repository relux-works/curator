package crossconformance_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/relux-works/curator/internal/crossconformance"
	"github.com/relux-works/curator/internal/rustsource"
)

// rustTargets are the two exact native targets the same Cargo lock superset is
// reconciled against. rust-source-v1 admits one native target per build, so
// the two branches are two separate native hosts rather than a cross build.
var rustTargets = []struct{ label, target, toolchain string }{
	{label: "aarch64-apple-darwin", target: "aarch64-apple-darwin", toolchain: "sha256:" + repeat64('1')},
	{label: "x86_64-unknown-linux-gnu", target: "x86_64-unknown-linux-gnu", toolchain: "sha256:" + repeat64('2')},
}

func repeat64(character byte) string {
	value := make([]byte, 64)
	for index := range value {
		value[index] = character
	}
	return string(value)
}

func rustWorkspaceFiles() map[string]string {
	return map[string]string{
		"Cargo.toml":      "[workspace]\nmembers=[\"app\",\"dep\"]\nresolver=\"2\"\n",
		"Cargo.lock":      "version = 4\n\n[[package]]\nname=\"app\"\nversion=\"0.1.0\"\ndependencies=[\"dep\"]\n\n[[package]]\nname=\"dep\"\nversion=\"0.1.0\"\n",
		"app/Cargo.toml":  "[package]\nname=\"app\"\nversion=\"0.1.0\"\nedition=\"2021\"\n[dependencies]\ndep={path=\"../dep\"}\n[[bin]]\nname=\"app-bin\"\npath=\"src/main.rs\"\n",
		"app/src/main.rs": "fn main(){println!(\"{}\", dep::value());}\n",
		"dep/Cargo.toml":  "[package]\nname=\"dep\"\nversion=\"0.1.0\"\nedition=\"2021\"\n",
		"dep/src/lib.rs":  "pub fn value()->u8{5}\n",
	}
}

// rustMetadata is the closed Cargo metadata projection rust-source-v1 accepts
// for the workspace above. Its manifest paths are exactly the captured ones,
// which is what lets Reconcile prove containment.
const rustMetadata = `{"version":1,"packages":[` +
	`{"id":"app 0.1.0","name":"app","version":"0.1.0","source":null,"manifest_path":"app/Cargo.toml","links":null,` +
	`"targets":[{"name":"app-bin","src_path":"app/src/main.rs","kind":["bin"],"crate_types":["bin"]}]},` +
	`{"id":"dep 0.1.0","name":"dep","version":"0.1.0","source":null,"manifest_path":"dep/Cargo.toml","links":null,` +
	`"targets":[{"name":"dep","src_path":"dep/src/lib.rs","kind":["lib"],"crate_types":["lib"]}]}],` +
	`"resolve":{"root":"app 0.1.0","nodes":[` +
	`{"id":"app 0.1.0","features":[],"dependencies":["dep 0.1.0"],"deps":[{"pkg":"dep 0.1.0","name":"dep","dep_kinds":[{"kind":null,"target":null}]}]},` +
	`{"id":"dep 0.1.0","features":[],"dependencies":[],"deps":[]}]}}`

// rustCaptureGraph builds the selection-neutral rust-source-v1 lock superset
// from the same bytes the manager captures, using the production constructor.
func rustCaptureGraph(t *testing.T, artifactManifestIDs []string) rustsource.CaptureGraph {
	t.Helper()
	files := rustWorkspaceFiles()
	lock, err := rustsource.ParseLock([]byte(files["Cargo.lock"]))
	if err != nil {
		t.Fatal(err)
	}
	manifests := []rustsource.Manifest{}
	for _, name := range []string{"Cargo.toml", "app/Cargo.toml", "dep/Cargo.toml"} {
		manifest, parseErr := rustsource.ParseManifest(name, []byte(files[name]))
		if parseErr != nil {
			t.Fatalf("%s: %v", name, parseErr)
		}
		manifests = append(manifests, manifest)
	}
	capture, err := rustsource.NewCaptureGraph(lock, manifests, artifactManifestIDs)
	if err != nil {
		t.Fatal(err)
	}
	return capture
}

// projectRustPath reconciles one exact native target and toolchain against the
// shared lock superset and reports the cross-adapter projection.
func projectRustPath(t *testing.T, targetIndex int, artifactManifestIDs []string) crossconformance.TargetProjection {
	t.Helper()
	target := rustTargets[targetIndex]
	capture := rustCaptureGraph(t, artifactManifestIDs)
	metadata, err := rustsource.ParseMetadata([]byte(rustMetadata))
	if err != nil {
		t.Fatal(err)
	}
	selection := rustsource.SelectionContext{
		Package: "app", Binary: "app-bin", Target: target.target, DefaultFeatures: true,
		Features: []string{}, TargetCFG: []string{},
		ResolvedFeatures: map[string][]string{"app 0.1.0": {}, "dep 0.1.0": {}},
	}
	active, err := rustsource.Reconcile(capture, selection, metadata, target.target, target.toolchain)
	if err != nil {
		t.Fatalf("rust reconcile %s: %v", target.label, err)
	}
	return crossconformance.TargetProjection{
		Path:                   crossconformance.PathRust,
		TargetLabel:            target.label,
		CaptureIdentity:        capture.Identity,
		SelectionIdentity:      active.Identity,
		BindingIdentity:        active.Identity,
		ActiveIdentity:         active.Identity,
		TargetPlatformIdentity: target.target,
		EmitsBindingRecords:    false,
		ToolIdentities:         []string{target.toolchain},
	}
}

// rustCaptureText is the canonical text of the selection-neutral capture. The
// suite uses it to prove no bound target or tool identity is spelled inside.
func rustCaptureText(t *testing.T) string {
	t.Helper()
	capture := rustCaptureGraph(t, nil)
	text := capture.SchemaID + "\n" + capture.LockDigest + "\n" + capture.Identity + "\n"
	for _, pkg := range capture.Packages {
		text += pkg.Key.String() + "\n"
	}
	for _, declaration := range capture.Declarations {
		text += declaration.Name + " " + declaration.Kind + " " + declaration.Target + " " + declaration.Package + "\n"
	}
	for _, path := range capture.CapturedManifestPaths {
		text += path + "\n"
	}
	return text
}

// rustManagerEvidence runs the real rust-source-v1 manager so the cross suite
// consumes the same C0-bound Cargo registration, vendor transform receipt, and
// metadata receipts the accepted Rust suite does. rust-source-v1 pins one
// approved Cargo descriptor per native target, so this is the same environment
// requirement the delivered Rust adapter already carries.
type rustManagerEvidence struct {
	artifactManifestIDs []string
	receipts            []string
	tools               []string
	target              string
}

func runRustManager(t *testing.T) rustManagerEvidence {
	t.Helper()
	workspace := t.TempDir()
	files := rustWorkspaceFiles()
	for name, payload := range files {
		writeFile(t, filepath.Join(workspace, filepath.FromSlash(name)), []byte(payload))
	}
	manager, err := rustsource.NewManager(context.Background(), rustsource.ManagerConfig{WorkRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("rust manager (needs the approved pinned Cargo for %s/%s): %v", runtime.GOOS, runtime.GOARCH, err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	manifests := []rustsource.RawManifest{}
	for _, name := range []string{"Cargo.toml", "app/Cargo.toml", "dep/Cargo.toml"} {
		manifests = append(manifests, rustsource.RawManifest{Path: name, File: rustsource.RawFile{Path: filepath.Join(workspace, filepath.FromSlash(name))}})
	}
	paths := []rustsource.RawPathOrigin{}
	for _, name := range []string{"app", "dep"} {
		paths = append(paths, rustsource.RawPathOrigin{DeclaredPath: name, Tree: rustsource.RawTree{Root: filepath.Join(workspace, name)}})
	}
	capture, err := manager.Capture(context.Background(), rustsource.RawCaptureRequest{
		Workspace: rustsource.RawTree{Root: workspace},
		Lock:      rustsource.RawFile{Path: filepath.Join(workspace, "Cargo.lock")},
		Manifests: manifests, Paths: paths,
	})
	if err != nil {
		t.Fatalf("rust capture: %v", err)
	}
	toolchain, err := manager.BuildToolchain()
	if err != nil {
		t.Fatal(err)
	}
	tools := []string{}
	for _, tool := range toolchain {
		tools = append(tools, string(tool.ContentFingerprint))
	}
	evidence := rustManagerEvidence{
		artifactManifestIDs: capture.Evidence.ArtifactManifestIDs,
		receipts:            appendUnique(nil, capture.Evidence.VendorReceipt),
		tools:               tools,
	}
	native, ok := map[string]string{"darwin": "apple-darwin", "linux": "unknown-linux-gnu"}[runtime.GOOS]
	if !ok {
		t.Fatalf("rust-source-v1 declares no native target for %s", runtime.GOOS)
	}
	architecture := map[string]string{"amd64": "x86_64", "arm64": "aarch64"}[runtime.GOARCH]
	evidence.target = architecture + "-" + native
	metadata, err := manager.DeriveMetadata(context.Background(), capture, rustsource.SelectionContext{
		Package: "app", Binary: "app-bin", Target: evidence.target, DefaultFeatures: true,
		Features: []string{}, TargetCFG: []string{},
	})
	if err != nil {
		t.Fatalf("rust metadata derivation: %v", err)
	}
	evidence.receipts = appendUnique(evidence.receipts, metadata.UnfilteredReceipt, metadata.ActiveReceipt)
	for _, receipt := range capture.Evidence.GitProjectionReceipts {
		evidence.receipts = appendUnique(evidence.receipts, receipt)
	}
	return evidence
}
