//go:build windows

package godriver

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestSelectResolvesAGoRootReachedThroughADirectoryJunction pins the selection
// of a Go installation the host publishes behind a directory junction, which is
// how the GitHub Actions tool cache publishes one on Windows.
//
// A junction is a link, but since Go 1.23 os.Lstat reports it as ModeIrregular
// rather than ModeSymlink -- it is a reparse-tag name surrogate, so the ModeDir
// branch is skipped, and its tag is not IO_REPARSE_TAG_SYMLINK, so it falls to
// ModeIrregular. filepath.EvalSymlinks follows only ModeSymlink and therefore
// returned the junction untouched, after which "is the resolved root a real
// directory?" answered no about an ordinary directory and this boundary refused
// the host's own working toolchain with go_toolchain_missing.
//
// The case also pins the property that makes accepting it safe: what comes back
// is the physical directory, not the junction. The fingerprint and the
// verifySelectedRoot identity check anchor there, so retargeting the junction
// afterwards cannot move the trusted toolchain under this session.
func TestSelectResolvesAGoRootReachedThroughADirectoryJunction(t *testing.T) {
	physical := testToolchain(t)
	junction := filepath.Join(t.TempDir(), "x64")
	makeTestJunction(t, junction, physical)

	// The junction really is the shape that used to be refused.
	info, err := os.Lstat(junction)
	if err != nil {
		t.Fatal(err)
	}
	if info.IsDir() {
		t.Skip("this host cannot create a junction os.Lstat reports as a non-directory")
	}

	executable, root, rootInfo, err := selectToolchain(Config{GOROOT: junction}, "")
	if err != nil {
		t.Fatalf("select a GOROOT reached through a directory junction: %v", err)
	}
	wantRoot := mustPhysical(t, physical)
	if root != wantRoot {
		t.Fatalf("selected GOROOT = %q, want the physical target %q", root, wantRoot)
	}
	if executable != filepath.Join(wantRoot, "bin", platformGoName) {
		t.Fatalf("selected launcher = %q, want it under the physical root", executable)
	}
	if rootInfo == nil || !rootInfo.IsDir() {
		t.Fatalf("selected root info = %+v, want a directory", rootInfo)
	}
	// The identity the session re-checks before every child must hold for the
	// physical root it was given.
	if err := verifySelectedRoot(root, rootInfo, executable); err != nil {
		t.Fatalf("verify the selected physical root: %v", err)
	}
}

// TestSelectStillRefusesAJunctionToANonDirectory keeps the type check honest:
// resolving a link is not the same as trusting whatever it points at.
func TestSelectStillRefusesAJunctionToANonDirectory(t *testing.T) {
	physical := testToolchain(t)
	junction := filepath.Join(t.TempDir(), "x64")
	makeTestJunction(t, junction, physical)
	if err := os.RemoveAll(physical); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := selectToolchain(Config{GOROOT: junction}, ""); err == nil {
		t.Fatal("a junction whose target is gone was accepted as a trusted GOROOT")
	}
}

func makeTestJunction(t *testing.T, link, target string) {
	t.Helper()
	// mklink /J is the unprivileged way to create a junction; the symlink forms
	// need SeCreateSymbolicLinkPrivilege, which a hosted runner may not grant.
	output, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput()
	if err != nil {
		t.Skipf("this host cannot create a directory junction: %v: %s", err, output)
	}
}
