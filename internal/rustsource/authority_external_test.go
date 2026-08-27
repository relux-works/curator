package rustsource_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/rustsource"
)

func TestRawCaptureAPIContainsNoAuthorityOrDerivedSeam(t *testing.T) {
	forbiddenNames := []string{"runner", "executor", "provider", "permit", "receipt", "toolchain", "cargo", "config", "destination", "selected", "normalized", "leaves", "submodules", "tree_digest"}
	seen := map[reflect.Type]bool{}
	var inspect func(reflect.Type)
	inspect = func(value reflect.Type) {
		if seen[value] {
			return
		}
		seen[value] = true
		switch value.Kind() {
		case reflect.Func, reflect.Interface, reflect.Chan, reflect.UnsafePointer:
			t.Fatalf("reachable authority kind %s (%s)", value.Kind(), value)
		case reflect.Pointer, reflect.Slice, reflect.Array:
			inspect(value.Elem())
		case reflect.Struct:
			for i := 0; i < value.NumField(); i++ {
				field := value.Field(i)
				name := strings.ToLower(field.Name)
				for _, forbidden := range forbiddenNames {
					if strings.Contains(name, forbidden) {
						t.Fatalf("authority/derived field %s reachable through %s", field.Name, value)
					}
				}
				inspect(field.Type)
			}
		}
	}
	inspect(reflect.TypeOf(rustsource.RawCaptureRequest{}))
	if reflect.TypeOf(rustsource.Manager{}).Kind() != reflect.Struct {
		t.Fatal("Manager is not concrete")
	}
}

func TestProductionManagerCapturesRegistryFromRawPaths(t *testing.T) {
	requireNativeCargoDescriptor(t)
	workspace := t.TempDir()
	if err := os.Mkdir(filepath.Join(workspace, "src"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "src", "main.rs"), []byte("fn main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := []byte("[package]\nname = \"root\"\nversion = \"0.1.0\"\n[dependencies]\ndep = \"1\"\n")
	manifestPath := filepath.Join(workspace, "Cargo.toml")
	if err := os.WriteFile(manifestPath, manifest, 0o600); err != nil {
		t.Fatal(err)
	}
	archive := crate(t, "dep", "1.0.0", map[string][]byte{"Cargo.toml": []byte("[package]\nname='dep'\nversion='1.0.0'\n"), "src/lib.rs": []byte("pub fn value() -> u8 { 7 }\n")})
	sum := sha256.Sum256(archive)
	checksum := hex.EncodeToString(sum[:])
	index, err := json.Marshal(map[string]any{"name": "dep", "vers": "1.0.0", "cksum": checksum, "deps": []any{}})
	if err != nil {
		t.Fatal(err)
	}
	lock := []byte("version = 4\n\n[[package]]\nname = \"root\"\nversion = \"0.1.0\"\ndependencies = [\"dep\"]\n\n[[package]]\nname = \"dep\"\nversion = \"1.0.0\"\nsource = \"registry+https://github.com/rust-lang/crates.io-index\"\nchecksum = \"" + checksum + "\"\n")
	lockPath := filepath.Join(workspace, "Cargo.lock")
	indexPath := filepath.Join(workspace, "dep.index.json")
	cratePath := filepath.Join(workspace, "dep.crate")
	for path, payload := range map[string][]byte{lockPath: lock, indexPath: index, cratePath: archive} {
		if err := os.WriteFile(path, payload, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	manager, err := rustsource.NewManager(t.Context(), rustsource.ManagerConfig{WorkRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	capture, err := manager.Capture(t.Context(), rustsource.RawCaptureRequest{Workspace: rustsource.RawTree{Root: workspace}, Lock: rustsource.RawFile{Path: lockPath}, Manifests: []rustsource.RawManifest{{Path: "Cargo.toml", File: rustsource.RawFile{Path: manifestPath}}}, Registry: []rustsource.RawRegistryOrigin{{SourceLocator: "registry+https://github.com/rust-lang/crates.io-index", IndexRecord: rustsource.RawFile{Path: indexPath}, CrateArchive: rustsource.RawFile{Path: cratePath}}}, Paths: []rustsource.RawPathOrigin{{DeclaredPath: ".", Tree: rustsource.RawTree{Root: workspace}}}})
	if err != nil {
		t.Fatal(err)
	}
	if capture.Evidence.TransformID != rustsource.TransformID || capture.Evidence.VendorReceipt == "" || len(capture.Evidence.VendorPackages) != 1 {
		t.Fatalf("capture evidence = %#v", capture.Evidence)
	}
	metadata, err := manager.DeriveMetadata(t.Context(), capture, rustsource.SelectionContext{Target: rustHost(t), DefaultFeatures: true})
	if err != nil {
		t.Fatalf("metadata: %#v", err)
	}
	if metadata.UnfilteredReceipt == "" || metadata.ActiveReceipt == "" || len(metadata.Active.Packages) == 0 {
		t.Fatalf("metadata evidence = %#v", metadata)
	}
}

func TestZeroManagerAndForeignCaptureFailClosed(t *testing.T) {
	requireNativeCargoDescriptor(t)
	var zero rustsource.Manager
	if _, err := zero.Capture(t.Context(), rustsource.RawCaptureRequest{}); err == nil {
		t.Fatal("zero manager authorized capture")
	}
	first, err := rustsource.NewManager(t.Context(), rustsource.ManagerConfig{WorkRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	second, err := rustsource.NewManager(t.Context(), rustsource.ManagerConfig{WorkRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = first.Close(); _ = second.Close() })
	if _, err = second.DeriveMetadata(t.Context(), &rustsource.Capture{}, rustsource.SelectionContext{}); err == nil {
		t.Fatal("foreign/empty capture authorized metadata")
	}
}

func TestProductionManagerCapturesGitWithoutCallerProjection(t *testing.T) {
	requireNativeCargoDescriptor(t)
	submodule := t.TempDir()
	if err := os.WriteFile(filepath.Join(submodule, "nested.txt"), []byte("submodule evidence\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git(t, submodule, "init", "-q")
	git(t, submodule, "config", "user.name", "Curator Test")
	git(t, submodule, "config", "user.email", "curator@example.invalid")
	git(t, submodule, "add", "nested.txt")
	git(t, submodule, "commit", "-qm", "submodule fixture")
	repository := t.TempDir()
	gitManifest := []byte("[package]\nname = \"git_leaf\"\nversion = \"0.1.0\"\nedition = \"2021\"\n")
	if err := os.Mkdir(filepath.Join(repository, "src"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "Cargo.toml"), gitManifest, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "src", "lib.rs"), []byte("pub fn value() -> u8 { 7 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git(t, repository, "init", "-q")
	git(t, repository, "config", "user.name", "Curator Test")
	git(t, repository, "config", "user.email", "curator@example.invalid")
	git(t, repository, "-c", "protocol.file.allow=always", "submodule", "add", "-q", "file://"+filepath.ToSlash(submodule), "deps/sub")
	git(t, repository, "add", "Cargo.toml", "src/lib.rs")
	git(t, repository, "commit", "-qm", "fixture")
	commit := git(t, repository, "rev-parse", "HEAD")
	git(t, repository, "gc", "--aggressive", "--prune=now")

	workspace := t.TempDir()
	if err := os.Mkdir(filepath.Join(workspace, "src"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "src", "main.rs"), []byte("fn main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	rootManifest := []byte("[package]\nname = \"root\"\nversion = \"0.1.0\"\n[dependencies]\ngit_leaf = { git = \"file://" + filepath.ToSlash(repository) + "\", rev = \"" + commit + "\" }\n")
	manifestPath := filepath.Join(workspace, "Cargo.toml")
	lockPath := filepath.Join(workspace, "Cargo.lock")
	if err := os.WriteFile(manifestPath, rootManifest, 0o600); err != nil {
		t.Fatal(err)
	}
	source := "git+file://" + filepath.ToSlash(repository) + "?rev=" + commit + "#" + commit
	lock := []byte("version = 4\n\n[[package]]\nname = \"root\"\nversion = \"0.1.0\"\ndependencies = [\"git_leaf\"]\n\n[[package]]\nname = \"git_leaf\"\nversion = \"0.1.0\"\nsource = \"" + source + "\"\n")
	if err := os.WriteFile(lockPath, lock, 0o600); err != nil {
		t.Fatal(err)
	}
	manager, err := rustsource.NewManager(t.Context(), rustsource.ManagerConfig{WorkRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	capture, err := manager.Capture(t.Context(), rustsource.RawCaptureRequest{Workspace: rustsource.RawTree{Root: workspace}, Lock: rustsource.RawFile{Path: lockPath}, Manifests: []rustsource.RawManifest{{Path: "Cargo.toml", File: rustsource.RawFile{Path: manifestPath}}}, Git: []rustsource.RawGitOrigin{{DeclaredURL: "file://" + filepath.ToSlash(repository), Selector: "rev=" + commit, LockedCommit: commit, Repository: rustsource.RawTree{Root: repository}}}, Paths: []rustsource.RawPathOrigin{{DeclaredPath: ".", Tree: rustsource.RawTree{Root: workspace}}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(capture.Evidence.VendorPackages) != 1 || capture.Evidence.VendorPackages[0].Kind != rustsource.SourceGit || len(capture.Evidence.GitObjectReceipts) != 1 || len(capture.Evidence.GitProjectionReceipts) != 1 {
		t.Fatalf("Git vendor evidence = %#v", capture.Evidence.VendorPackages)
	}
	var normalized bool
	var copiedSubmodule bool
	for _, entry := range capture.Evidence.VendorPackages[0].Entries {
		if entry.OriginPath == "Cargo.toml" && entry.Disposition == rustsource.ReplaceNormalizedManifest {
			normalized = true
		}
		if entry.OriginPath == "deps/sub/nested.txt" && entry.Disposition == rustsource.CopyIdentical {
			copiedSubmodule = true
		}
	}
	if !normalized {
		t.Fatal("Git manifest was not independently normalized")
	}
	if !copiedSubmodule {
		t.Fatal("recursive Git submodule leaf was not bound into the transform")
	}
	metadata, err := manager.DeriveMetadata(t.Context(), capture, rustsource.SelectionContext{Target: rustHost(t), DefaultFeatures: true})
	if err != nil {
		t.Fatalf("Git metadata: %#v", err)
	}
	if metadata.UnfilteredReceipt == "" || metadata.ActiveReceipt == "" {
		t.Fatal("Git metadata receipts are absent")
	}
}

func requireNativeCargoDescriptor(t *testing.T) {
	t.Helper()
	target, approved := rustsource.NativeCargoDescriptorAvailable()
	if target != "" && !approved {
		t.Skipf("no operator-approved Cargo descriptor for native target %s", target)
	}
}

func git(t *testing.T, root string, args ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, args...)...) // #nosec G204 -- fixed test fixture command.
	payload, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, payload)
	}
	return strings.TrimSpace(string(payload))
}

func rustHost(t *testing.T) string {
	t.Helper()
	payload, err := exec.Command("rustc", "-vV").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(payload), "\n") {
		if strings.HasPrefix(line, "host: ") {
			return strings.TrimPrefix(line, "host: ")
		}
	}
	t.Fatal("rustc host is absent")
	return ""
}

func crate(t *testing.T, name, version string, files map[string][]byte) []byte {
	t.Helper()
	var result bytes.Buffer
	gz := gzip.NewWriter(&result)
	tarWriter := tar.NewWriter(gz)
	keys := make([]string, 0, len(files))
	for key := range files {
		keys = append(keys, key)
	}
	for _, key := range keys {
		payload := files[key]
		if err := tarWriter.WriteHeader(&tar.Header{Name: name + "-" + version + "/" + key, Mode: 0o600, Size: int64(len(payload)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tarWriter.Write(payload); err != nil {
			t.Fatal(err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return result.Bytes()
}
