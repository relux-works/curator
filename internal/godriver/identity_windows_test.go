//go:build windows

package godriver

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExecutableIdentityResolvesALauncherReachedThroughADirectoryJunction pins
// the Windows launch shape that has no unix counterpart: the manager is
// installed under a directory the host publishes as a junction, which is how a
// tool cache and several package managers put an install directory on PATH.
//
// A junction is a link, but since Go 1.23 os.Lstat reports it as ModeIrregular
// rather than ModeSymlink, so filepath.EvalSymlinks does not follow it, and for
// a junction that is not the last component walkSymlinks refuses to descend at
// all. Canonicalizing the running executable with EvalSymlinks therefore refused
// the operator's own working install with worker_identity_invalid before any
// substitution check ran. physicalPath resolves every link kind in one call, so
// resolution comes first and the substitution checks below apply to the physical
// file.
func TestExecutableIdentityResolvesALauncherReachedThroughADirectoryJunction(t *testing.T) {
	install := t.TempDir()
	installed := writeTestExecutable(t, filepath.Join(install, "curator.exe"), "installed manager bytes")
	direct, err := resolveExecutableIdentity(installed)
	if err != nil {
		t.Fatalf("the installed file was not accepted directly: %v", err)
	}

	junction := filepath.Join(t.TempDir(), "current")
	makeTestJunction(t, junction, install)
	info, err := os.Lstat(junction)
	if err != nil {
		t.Fatal(err)
	}
	if info.IsDir() {
		t.Skip("this host cannot create a junction os.Lstat reports as a non-directory")
	}

	launcher := filepath.Join(junction, "curator.exe")
	// The shape really is the one EvalSymlinks alone cannot canonicalize.
	if resolved, evalErr := filepath.EvalSymlinks(launcher); evalErr == nil && resolved == direct.Path {
		t.Logf("this host's EvalSymlinks already resolved %q; the junction case is covered by physicalPath either way", launcher)
	} else {
		t.Logf("EvalSymlinks(%q) = %q, %v", launcher, resolved, evalErr)
	}

	identity, err := resolveExecutableIdentity(launcher)
	if err != nil {
		t.Fatalf("a manager launched through a directory junction was refused: %v", err)
	}
	if identity.Path != direct.Path || identity.SHA256 != direct.SHA256 || identity.Size != direct.Size {
		t.Fatalf("identity = %+v, want the installed file %+v", identity, direct)
	}
	if err := identity.matches(direct.Path, direct.SHA256, direct.Size); err != nil {
		t.Fatalf("the worker proof did not match the manager identity: %v", err)
	}
	if err := identity.Verify(); err != nil {
		t.Fatalf("re-proving the resolved identity failed: %v", err)
	}
}

// TestExecutableIdentityStillRejectsSubstitutionBehindAJunction keeps the
// rejection half honest for the junction shape: resolving the link is not
// trusting what it reaches, and a recorded identity stays anchored on the
// physical file rather than on the junction.
func TestExecutableIdentityStillRejectsSubstitutionBehindAJunction(t *testing.T) {
	install := t.TempDir()
	installed := writeTestExecutable(t, filepath.Join(install, "curator.exe"), "installed manager bytes")
	junction := filepath.Join(t.TempDir(), "current")
	makeTestJunction(t, junction, install)

	launcher := filepath.Join(junction, "curator.exe")
	identity, err := resolveExecutableIdentity(launcher)
	if err != nil {
		t.Fatalf("a manager launched through a directory junction was refused: %v", err)
	}
	if identity.Path != mustPhysical(t, installed) {
		t.Fatalf("identity path = %q, want the physical installed file %q", identity.Path, installed)
	}

	// A junction retargeted at another install cannot move the recorded
	// identity, and the file it now reaches cannot pass as the same manager.
	other := t.TempDir()
	substituted := writeTestExecutable(t, filepath.Join(other, "curator.exe"), "attacker bytes")
	if err := os.Remove(junction); err != nil {
		t.Fatal(err)
	}
	makeTestJunction(t, junction, other)
	if err := identity.Verify(); err != nil {
		t.Fatalf("retargeting the junction moved a recorded identity: %v", err)
	}
	reached, err := resolveExecutableIdentity(launcher)
	if err != nil {
		t.Fatal(err)
	}
	if reached.Path != mustPhysical(t, substituted) {
		t.Fatalf("retargeted launcher resolved to %q, want %q", reached.Path, substituted)
	}
	if code := DiagnosticCode(identity.matches(reached.Path, reached.SHA256, reached.Size)); code != CodeWorkerIdentityInvalid {
		t.Fatalf("substituted proof code = %q, want %q", code, CodeWorkerIdentityInvalid)
	}

	// A junction whose target is gone is a launch that names nothing, and a
	// hard-linked file behind a junction is still a substitution.
	if err := os.RemoveAll(other); err != nil {
		t.Fatal(err)
	}
	if code := DiagnosticCode(mustFail(resolveExecutableIdentity(launcher))); code != CodeWorkerIdentityInvalid {
		t.Fatalf("dangling junction code = %q, want %q", code, CodeWorkerIdentityInvalid)
	}
	if err := os.Link(installed, filepath.Join(install, "curator-alias.exe")); err != nil {
		t.Skipf("this host cannot create a hard link: %v", err)
	}
	if err := os.Remove(junction); err != nil {
		t.Fatal(err)
	}
	makeTestJunction(t, junction, install)
	if code := DiagnosticCode(mustFail(resolveExecutableIdentity(launcher))); code != CodeWorkerIdentityInvalid {
		t.Fatalf("hard-linked executable behind a junction code = %q, want %q", code, CodeWorkerIdentityInvalid)
	}
}
