package godriver

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// linkLauncherForTest creates a launcher link named link that points at the
// installed file target, which is the shape a package manager publishes on
// PATH. On Windows a file symbolic link needs SeCreateSymbolicLinkPrivilege,
// which a plain local account may not hold; the directory-junction form of the
// same launch shape is covered without any privilege in
// identity_windows_test.go.
//
// The skip reason is the exact phrasing `.github/ci/skip-classes.tsv`
// classifies as `host-capability` -- the same one
// internal/buildcache/protection_windows_test.go uses for this situation. A
// Windows runner without the privilege must degrade on a recorded class, not
// fail the platform-case gate on an unrecognised reason, so reword it only
// together with that table.
func linkLauncherForTest(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("creating Windows symlink requires host support: %v", err)
		}
		t.Fatal(err)
	}
}

// writeTestExecutable writes a native-looking installed file so a resolved
// identity is exercised against the same byte shape a real manager has.
func writeTestExecutable(t *testing.T, path string, tail string) string {
	t.Helper()
	writeTestFile(t, path, append(testExecutableBytes(), []byte(tail)...), 0o755)
	return mustPhysical(t, path)
}

// TestExecutableIdentityResolvesALauncherLink pins the ordering the manager
// profile requires: the running executable is resolved to the canonical
// installed file first, and only then is substitution of that file rejected.
// The launch shape an operator used -- a shim symlink on PATH, a chain of them,
// a link to the install directory -- is not itself a fault, and every shape must
// produce the one identity of the one installed file.
func TestExecutableIdentityResolvesALauncherLink(t *testing.T) {
	install := t.TempDir()
	installed := writeTestExecutable(t, filepath.Join(install, "curator"), "installed manager bytes")
	direct, err := resolveExecutableIdentity(installed)
	if err != nil {
		t.Fatalf("the installed file was not accepted directly: %v", err)
	}

	for _, testCase := range []struct {
		name    string
		launch  func(t *testing.T, shim string) string
		skipWin bool
	}{
		{
			name: "shim symlink on PATH",
			launch: func(t *testing.T, shim string) string {
				link := filepath.Join(shim, "curator")
				linkLauncherForTest(t, installed, link)
				return link
			},
		},
		{
			name: "chain of shim symlinks",
			launch: func(t *testing.T, shim string) string {
				first := filepath.Join(shim, "curator-1")
				second := filepath.Join(shim, "curator-2")
				linkLauncherForTest(t, installed, first)
				linkLauncherForTest(t, first, second)
				return second
			},
		},
		{
			name: "symlinked install directory",
			launch: func(t *testing.T, shim string) string {
				link := filepath.Join(shim, "opt")
				linkLauncherForTest(t, install, link)
				return filepath.Join(link, "curator")
			},
		},
		{
			name: "unclean relative spelling of the link",
			launch: func(t *testing.T, shim string) string {
				link := filepath.Join(shim, "curator")
				linkLauncherForTest(t, installed, link)
				return filepath.Join(shim, "nested", "..", "curator")
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			shim := t.TempDir()
			identity, err := resolveExecutableIdentity(testCase.launch(t, shim))
			if err != nil {
				t.Fatalf("a manager launched through this shape was refused: %v", err)
			}
			if identity.Path != direct.Path || identity.SHA256 != direct.SHA256 || identity.Size != direct.Size {
				t.Fatalf("identity = %+v, want the installed file %+v", identity, direct)
			}
			// The recorded path is the physical file, so the worker the
			// manager re-executes resolves the same identity and the protocol
			// proof compares like with like.
			if err := identity.matches(direct.Path, direct.SHA256, direct.Size); err != nil {
				t.Fatalf("the worker proof did not match the manager identity: %v", err)
			}
			if err := identity.Verify(); err != nil {
				t.Fatalf("re-proving the resolved identity failed: %v", err)
			}
		})
	}
}

// TestExecutableIdentityRejectsSubstitutionThroughALauncherLink keeps the
// second half of the rule honest: resolving the launch shape is not trusting
// whatever it reaches. Every substitution of the installed file still fails
// closed with the one stable diagnostic, and an identity already recorded stays
// pinned to the physical file it named.
func TestExecutableIdentityRejectsSubstitutionThroughALauncherLink(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		build func(t *testing.T, directory string) string
	}{
		{
			name: "dangling launcher link",
			build: func(t *testing.T, directory string) string {
				link := filepath.Join(directory, "curator")
				linkLauncherForTest(t, filepath.Join(directory, "missing"), link)
				return link
			},
		},
		{
			name: "launcher link to a directory",
			build: func(t *testing.T, directory string) string {
				target := filepath.Join(directory, "elsewhere")
				if err := os.MkdirAll(target, 0o755); err != nil {
					t.Fatal(err)
				}
				link := filepath.Join(directory, "curator")
				linkLauncherForTest(t, target, link)
				return link
			},
		},
		{
			name: "launcher link to an empty file",
			build: func(t *testing.T, directory string) string {
				target := filepath.Join(directory, "empty")
				writeTestFile(t, target, nil, 0o755)
				link := filepath.Join(directory, "curator")
				linkLauncherForTest(t, target, link)
				return link
			},
		},
		{
			name: "launcher link loop",
			build: func(t *testing.T, directory string) string {
				first, second := filepath.Join(directory, "curator"), filepath.Join(directory, "curator-next")
				linkLauncherForTest(t, second, first)
				linkLauncherForTest(t, first, second)
				return first
			},
		},
		{
			name: "launcher link to a hard-linked file",
			build: func(t *testing.T, directory string) string {
				target := writeTestExecutable(t, filepath.Join(directory, "curator-real"), "hard linked")
				if err := os.Link(target, filepath.Join(directory, "curator-alias")); err != nil {
					t.Skipf("this host cannot create a hard link: %v", err)
				}
				link := filepath.Join(directory, "curator")
				linkLauncherForTest(t, target, link)
				return link
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			path := testCase.build(t, t.TempDir())
			if code := DiagnosticCode(mustFail(resolveExecutableIdentity(path))); code != CodeWorkerIdentityInvalid {
				t.Fatalf("substitution code = %q, want %q", code, CodeWorkerIdentityInvalid)
			}
		})
	}

	// A recorded identity anchors on the physical file. Retargeting the shim
	// afterwards cannot move it, and the file it now reaches cannot pass as the
	// same manager.
	directory := t.TempDir()
	installed := writeTestExecutable(t, filepath.Join(directory, "curator-real"), "installed manager bytes")
	other := writeTestExecutable(t, filepath.Join(directory, "curator-other"), "attacker bytes")
	link := filepath.Join(directory, "curator")
	linkLauncherForTest(t, installed, link)
	identity, err := resolveExecutableIdentity(link)
	if err != nil {
		t.Fatal(err)
	}
	if identity.Path != installed {
		t.Fatalf("identity path = %q, want the physical installed file %q", identity.Path, installed)
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	linkLauncherForTest(t, other, link)
	if err := identity.Verify(); err != nil {
		t.Fatalf("retargeting the shim moved a recorded identity: %v", err)
	}
	substituted, err := resolveExecutableIdentity(link)
	if err != nil {
		t.Fatal(err)
	}
	if code := DiagnosticCode(identity.matches(substituted.Path, substituted.SHA256, substituted.Size)); code != CodeWorkerIdentityInvalid {
		t.Fatalf("substituted proof code = %q, want %q", code, CodeWorkerIdentityInvalid)
	}
	// Replacing the bytes of the physical file is still caught by the re-proof
	// that runs at the launch boundary and again before publication.
	writeTestFile(t, installed, append(testExecutableBytes(), []byte("replaced bytes")...), 0o755)
	if code := DiagnosticCode(identity.Verify()); code != CodeWorkerIdentityInvalid {
		t.Fatalf("mutated executable code = %q, want %q", code, CodeWorkerIdentityInvalid)
	}

	// Swapping the physical file itself for a link is substitution of the
	// installed file rather than a launch shape, so the re-proof that runs at
	// the launch boundary must refuse it even though resolution reached the
	// original file through a link in the first place.
	swap := t.TempDir()
	physical := writeTestExecutable(t, filepath.Join(swap, "curator-real"), "installed manager bytes")
	shim := filepath.Join(swap, "curator")
	linkLauncherForTest(t, physical, shim)
	pinned, err := resolveExecutableIdentity(shim)
	if err != nil {
		t.Fatal(err)
	}
	elsewhere := writeTestExecutable(t, filepath.Join(swap, "curator-elsewhere"), "attacker bytes")
	if err := os.Remove(physical); err != nil {
		t.Fatal(err)
	}
	linkLauncherForTest(t, elsewhere, physical)
	if code := DiagnosticCode(pinned.Verify()); code != CodeWorkerIdentityInvalid {
		t.Fatalf("physical file swapped for a link code = %q, want %q", code, CodeWorkerIdentityInvalid)
	}
}

// TestExecutableIdentityResolvesARealSymlinkedProcessLaunch is the case the
// unit resolutions above cannot prove on their own: it starts this binary
// through a shim link and reads back what that process resolved for itself, so
// whatever os.Executable reports for this platform's launch shape is observed
// rather than assumed.
func TestExecutableIdentityResolvesARealSymlinkedProcessLaunch(t *testing.T) {
	self, err := resolveExecutableIdentity(managerExecutable())
	if err != nil {
		t.Fatalf("this test binary is not an acceptable manager executable: %v", err)
	}
	launcher := filepath.Join(t.TempDir(), "curator")
	if runtime.GOOS == "windows" {
		launcher += ".exe"
	}
	linkLauncherForTest(t, self.Path, launcher)

	probe := probeIdentityThrough(t, launcher)
	if probe.Error != "" {
		t.Fatalf("a manager started through %q refused its own identity: %s", launcher, probe.Error)
	}
	// Evidence for the platform, recorded either way: Linux reports the
	// already-resolved /proc/self/exe, while darwin and Windows report the path
	// the process was started with, so on those hosts canonicalization is what
	// makes the manager and the worker agree on one identity.
	if probe.Reported == self.Path {
		t.Logf("evidence: on %s os.Executable already reports the physical path %q for a launch through %q",
			runtime.GOOS, probe.Reported, launcher)
	} else {
		t.Logf("evidence: on %s os.Executable reports the launch path %q rather than the physical %q, so resolution is load-bearing here",
			runtime.GOOS, probe.Reported, self.Path)
	}
	if probe.Path != self.Path || probe.SHA256 != self.SHA256 || probe.Size != self.Size {
		t.Fatalf("identity resolved by the started process = %+v, want %+v", probe, self)
	}
}

// TestBuildAcceptsAManagerStartedThroughALauncherLink drives the whole
// identity handshake for a manager whose own path is a shim link: the manager
// resolves and hashes itself, re-executes the canonical file as the worker, and
// the worker proves the same identity back over the protocol.
func TestBuildAcceptsAManagerStartedThroughALauncherLink(t *testing.T) {
	self, err := resolveExecutableIdentity(managerExecutable())
	if err != nil {
		t.Fatal(err)
	}
	launcher := filepath.Join(t.TempDir(), "curator")
	if runtime.GOOS == "windows" {
		launcher += ".exe"
	}
	linkLauncherForTest(t, self.Path, launcher)

	restore := managerExecutable
	managerExecutable = func() string { return launcher }
	t.Cleanup(func() { managerExecutable = restore })

	fixture := newSnapshotFixture(t)
	fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
	result, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 60 * time.Second}))
	if err != nil {
		t.Fatalf("a build by a manager started through a shim link was refused: %v", err)
	}
	if result.Artifact.StagedPath == "" {
		t.Fatalf("result = %+v, want a staged artifact", result)
	}
}

// TestWorkerProvesTheInstalledIdentityWhenLaunchedThroughALink covers the half
// TestBuildAcceptsAManagerStartedThroughALauncherLink cannot reach: there the
// manager resolves a link but still re-executes the physical file, so the
// worker's own resolution never sees one. Here the worker process itself is
// started from a link and still has to prove the installed identity back over
// the protocol, which is the boundary at workerserver.go's accept.
func TestWorkerProvesTheInstalledIdentityWhenLaunchedThroughALink(t *testing.T) {
	scenario := newWorkerScenario(t)
	launcher := filepath.Join(t.TempDir(), "curator")
	if runtime.GOOS == "windows" {
		launcher += ".exe"
	}
	linkLauncherForTest(t, scenario.identity.Path, launcher)
	scenario.launchPath = launcher

	worker := scenario.start()
	message := worker.receive()
	if message.Kind != kindReady || message.Ready == nil {
		t.Fatalf("a worker started through %q sent %q, want ready", launcher, message.Kind)
	}
	if err := scenario.identity.matches(message.Ready.ExecutablePath, message.Ready.ExecutableSHA256, message.Ready.ExecutableSize); err != nil {
		t.Fatalf("a worker started through a launcher link did not prove the installed identity: %v", err)
	}
}

// probeIdentityThrough starts this binary at path in the test-only identity
// probe mode and returns what the started process resolved for itself.
func probeIdentityThrough(t *testing.T, path string) identityProbe {
	t.Helper()
	command := exec.Command(path, identityProbeMode) // #nosec G204 -- test-owned launcher link to this test binary
	command.Stderr = os.Stderr
	output, err := command.Output()
	if err != nil {
		t.Fatalf("cannot start the identity probe through %q: %v", path, err)
	}
	var probe identityProbe
	if err := json.Unmarshal(output, &probe); err != nil {
		t.Fatalf("cannot decode the identity probe output %q: %v", output, err)
	}
	return probe
}
