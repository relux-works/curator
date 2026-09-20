//go:build windows

package pnpmsource

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func makeTestJunction(t *testing.T, link, target string) {
	t.Helper()
	// mklink /J is the unprivileged way to create a junction; the symlink forms
	// need SeCreateSymbolicLinkPrivilege, which a hosted runner may not grant.
	output, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput()
	if err != nil {
		t.Skipf("this host cannot create a directory junction: %v: %s", err, output)
	}
}

func makeJunctionStoreFixture(t *testing.T) (store, project string, expected []StoreFile) {
	t.Helper()
	store = filepath.Join(t.TempDir(), "store")
	project = filepath.Join(t.TempDir(), "project")
	if err := os.MkdirAll(filepath.Join(store, "v10", "files"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(project, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store, "v10", "files", "content"), []byte("source\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var err error
	expected, err = inventoryStore(store)
	if err != nil {
		t.Fatal(err)
	}
	return store, project, expected
}

// TestWritableStoreOverlayAdmitsWindowsJunctionRegistration feeds the exact
// member shape pnpm 10.33.0 writes on Windows to the registry: registerProject
// links the project through symlink-dir, which creates a junction (Go reports
// ModeIrregular, and filepath.EvalSymlinks does not follow it), named with the
// 32-hex createShortHash of the project directory.
func TestWritableStoreOverlayAdmitsWindowsJunctionRegistration(t *testing.T) {
	store, project, expected := makeJunctionStoreFixture(t)
	registry := filepath.Join(store, "v10", "projects")
	if err := os.MkdirAll(registry, 0o700); err != nil {
		t.Fatal(err)
	}
	makeTestJunction(t, filepath.Join(registry, "c8805e97311ac3a64fd7c507d26b47dc"), project)
	if err := reconcileWritableStoreOverlay(store, project, expected); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(registry); !os.IsNotExist(err) {
		t.Fatalf("registry was not removed: %v", err)
	}
}

func TestWritableStoreOverlayRefusesMisdirectedWindowsJunction(t *testing.T) {
	store, project, expected := makeJunctionStoreFixture(t)
	registry := filepath.Join(store, "v10", "projects")
	if err := os.MkdirAll(registry, 0o700); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(t.TempDir(), "other")
	if err := os.MkdirAll(other, 0o700); err != nil {
		t.Fatal(err)
	}
	makeTestJunction(t, filepath.Join(registry, "4990f83e93c63b5b58405c5ac8c8ca6a"), other)
	assertCode(t, reconcileWritableStoreOverlay(store, project, expected), CodeInputUndeclared)
}

func TestAdmittedLinkCoversWindowsJunctionShape(t *testing.T) {
	target := filepath.Join(t.TempDir(), "target")
	if err := os.MkdirAll(target, 0o700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "link")
	makeTestJunction(t, link, target)
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if !admittedLink(info, link) {
		t.Fatal("junction is not an admitted link")
	}
	resolved, err := normalizeTreeLink(link, info)
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(target)
	if err != nil {
		t.Fatal(err)
	}
	got, err := filepath.EvalSymlinks(resolved)
	if err != nil || got != want {
		t.Fatalf("junction resolved to %q (%v), want %q", got, err, want)
	}
	plain, err := os.Lstat(target)
	if err != nil {
		t.Fatal(err)
	}
	if admittedLink(plain, target) {
		t.Fatal("plain directory is an admitted link")
	}
	if resolved, err := normalizeTreeLink(target, plain); err != nil || resolved != target {
		t.Fatalf("plain directory normalized to %q (%v), want identity", resolved, err)
	}
}

func TestDirectNodeModulesAdmitWindowsJunctionLinks(t *testing.T) {
	root := t.TempDir()
	nodeModules := filepath.Join(root, "node_modules")
	target := filepath.Join(root, "target-pkg")
	if err := os.MkdirAll(filepath.Join(nodeModules, ".pnpm"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(target, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nodeModules, ".modules.yaml"), []byte("layoutVersion: 10\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nodeModules, ".pnpm-workspace-state-v1.json"), []byte(`{"settings":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	makeTestJunction(t, filepath.Join(nodeModules, "dep"), target)
	if err := validateDirectNodeModules(nodeModules, true, map[string]string{"dep": target}); err != nil {
		t.Fatal(err)
	}
}

// TestCopyContainedTreeDereferencesWindowsJunction pins the junction spelling
// of the dereferencing-copy contract: pnpm's install tree is full of
// junctions, and the copy must resolve each one to its target before
// filepath.EvalSymlinks sees the path, the way the npm sibling does. Without
// that normalization the recursive copy descends through the junction
// textually and EvalSymlinks fails on the junction-carrying path, which is
// the bare "The system cannot find the path specified." the two real-pnpm
// install cases hit on windows-latest (gate run 35481906193).
func TestCopyContainedTreeDereferencesWindowsJunction(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "pkg"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pkg", "index.js"), []byte("module.exports = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "node_modules"), 0o700); err != nil {
		t.Fatal(err)
	}
	makeTestJunction(t, filepath.Join(root, "node_modules", "pkg"), filepath.Join(root, "pkg"))
	destination := filepath.Join(t.TempDir(), "copied")
	if err := copyContainedTree(root, destination); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(filepath.Join(destination, "node_modules", "pkg", "index.js"))
	if err != nil || string(payload) != "module.exports = 1\n" {
		t.Fatalf("through-junction content was not copied: %q %v", payload, err)
	}
	if err := filepath.WalkDir(destination, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == destination {
			return nil
		}
		if entry.Type() != 0 && entry.Type() != fs.ModeDir {
			t.Errorf("through-junction copy left a non-regular member %q: %v", current, entry.Type())
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestCopyContainedTreeRefusesWindowsJunctionEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret"), []byte("outside\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	makeTestJunction(t, filepath.Join(root, "escape"), outside)
	assertCode(t, copyContainedTree(root, filepath.Join(t.TempDir(), "copied")), CodeLocalPathEscape)
}

func TestCopyContainedTreeRefusesWindowsJunctionCycle(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "dir"), 0o700); err != nil {
		t.Fatal(err)
	}
	makeTestJunction(t, filepath.Join(root, "dir", "cycle"), root)
	assertCode(t, copyContainedTree(root, filepath.Join(t.TempDir(), "copied")), CodeInputUndeclared)
}
