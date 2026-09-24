package scriptworker

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/scriptpolicy"
)

// TestResolveInterpreterFromOperatorConfigOnly proves the two closed
// identifiers resolve from the operator-trusted mapping and nowhere else:
// the returned identity is the configured file's canonical path, hash, and
// size.
func TestResolveInterpreterFromOperatorConfigOnly(t *testing.T) {
	stub := mustPhysical(t, stubInterpreterBinary(t))
	for _, identifier := range []string{"node-v1", "python3-v1"} {
		t.Run(identifier, func(t *testing.T) {
			mapping := map[string]string{identifier: stub}
			identity, err := ResolveInterpreter(identifier, mapping, nil)
			if err != nil {
				t.Fatal(err)
			}
			if identity.ID != identifier || identity.Path != stub {
				t.Fatalf("identity = %+v, want id %q path %q", identity, identifier, stub)
			}
			if !strings.HasPrefix(identity.SHA256, "sha256:") || len(identity.SHA256) != len("sha256:")+64 {
				t.Fatalf("identity digest = %q, want a canonical sha256 identity", identity.SHA256)
			}
			if identity.Size <= 0 {
				t.Fatalf("identity size = %d, want positive", identity.Size)
			}
			// The mapping is per identifier: configuring one does not
			// resolve the other.
			other := "node-v1"
			if identifier == "node-v1" {
				other = "python3-v1"
			}
			if _, err := ResolveInterpreter(other, mapping, nil); DiagnosticCode(err) != scriptpolicy.ControlUnavailable {
				t.Fatalf("unconfigured %s error = %v, want %s", other, err, scriptpolicy.ControlUnavailable)
			}
		})
	}
}

// TestResolveInterpreterRejectsUnknownIdentifier proves resolution outside
// the closed set is refused as package influence without consulting any
// other source: even a configured binding for it cannot exist.
func TestResolveInterpreterRejectsUnknownIdentifier(t *testing.T) {
	stub := mustPhysical(t, stubInterpreterBinary(t))
	for _, identifier := range []string{"bash-v1", "powershell-v1", "node-v2", "", "python3-v1 "} {
		if _, err := ResolveInterpreter(identifier, map[string]string{identifier: stub}, nil); DiagnosticCode(err) != CodePackageInfluenceForbidden {
			t.Fatalf("identifier %q error = %v, want %s", identifier, err, CodePackageInfluenceForbidden)
		}
	}
}

// TestResolveInterpreterIgnoresUserPATH proves the user PATH never selects
// the executable. With no configured binding the resolution refuses even
// when PATH names a working interpreter; with a binding the configured
// file wins over anything on PATH.
func TestResolveInterpreterIgnoresUserPATH(t *testing.T) {
	stub := mustPhysical(t, stubInterpreterBinary(t))
	decoyDir := t.TempDir()
	decoy := filepath.Join(decoyDir, "python3")
	if runtime.GOOS == "windows" {
		decoy += ".exe"
	}
	copyTestFile(t, stub, decoy)
	t.Setenv("PATH", decoyDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	if _, err := ResolveInterpreter("python3-v1", nil, nil); DiagnosticCode(err) != scriptpolicy.ControlUnavailable {
		t.Fatalf("unconfigured resolution with PATH decoy = %v, want %s", err, scriptpolicy.ControlUnavailable)
	}
	identity, err := ResolveInterpreter("python3-v1", map[string]string{"python3-v1": stub}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if identity.Path != stub {
		t.Fatalf("resolved path = %q, want the configured file %q, not the PATH decoy", identity.Path, stub)
	}
}

// TestResolveInterpreterRejectsManifestSelection proves a manifest value
// cannot select the executable. The resolution signature takes only the
// closed identifier and the operator mapping; this row pins that a manifest
// naming a different path changes nothing.
func TestResolveInterpreterRejectsManifestSelection(t *testing.T) {
	stub := mustPhysical(t, stubInterpreterBinary(t))
	manifestDir := t.TempDir()
	manifestPath := filepath.Join(manifestDir, "agent-skill.json")
	writeTestFile(t, manifestPath, []byte(`{"commands":{"tool":{"interpreter_path":`+strconvQuote(filepath.Join(manifestDir, "evil"))+`}}}`), 0o644)
	decoy := filepath.Join(manifestDir, "evil")
	copyTestFile(t, stub, decoy)

	// The manifest-named path is never consulted: without a binding the
	// resolution refuses, with a binding it returns the configured file.
	if _, err := ResolveInterpreter("node-v1", nil, nil); DiagnosticCode(err) != scriptpolicy.ControlUnavailable {
		t.Fatalf("unconfigured resolution with manifest decoy = %v, want %s", err, scriptpolicy.ControlUnavailable)
	}
	identity, err := ResolveInterpreter("node-v1", map[string]string{"node-v1": stub}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if identity.Path != stub {
		t.Fatalf("resolved path = %q, want the configured file %q", identity.Path, stub)
	}
}

func strconvQuote(value string) string {
	return "\"" + strings.ReplaceAll(value, `"`, `\"`) + "\""
}

// TestResolveInterpreterRejectsRuntimeRoot proves an interpreter resolving
// under a repository or runtime root is refused as package influence, even
// when the operator mapping names it: configuration cannot bless a
// package-controlled location.
func TestResolveInterpreterRejectsRuntimeRoot(t *testing.T) {
	stub := stubInterpreterBinary(t)
	runtimeRoot := mustPhysical(t, t.TempDir())
	nested := filepath.Join(runtimeRoot, "bin", filepath.Base(stub))
	copyTestFile(t, stub, nested)
	nested = mustPhysical(t, nested)

	for _, root := range []string{runtimeRoot, filepath.Dir(nested)} {
		mapping := map[string]string{"python3-v1": nested}
		if _, err := ResolveInterpreter("python3-v1", mapping, []string{root}); DiagnosticCode(err) != CodePackageInfluenceForbidden {
			t.Fatalf("runtime-root interpreter under %q error = %v, want %s", root, err, CodePackageInfluenceForbidden)
		}
	}
	// The same file outside every forbidden root resolves normally.
	identity, err := ResolveInterpreter("python3-v1", map[string]string{"python3-v1": nested}, []string{mustPhysical(t, t.TempDir())})
	if err != nil {
		t.Fatal(err)
	}
	if identity.Path != nested {
		t.Fatalf("resolved path = %q, want %q", identity.Path, nested)
	}
}

// TestResolveInterpreterResolvesABindingThroughALink proves a binding that
// names the interpreter through a launcher link resolves to the physical
// file, exactly like the manager's own launcher-link handling: what the
// checks reject is substitution of that file, not the link.
func TestResolveInterpreterResolvesABindingThroughALink(t *testing.T) {
	stub := mustPhysical(t, stubInterpreterBinary(t))
	link := filepath.Join(mustPhysical(t, t.TempDir()), "interp-link")
	if err := os.Symlink(stub, link); err != nil {
		t.Skipf("this host forbids symlink creation: %v", err)
	}
	identity, err := ResolveInterpreter("node-v1", map[string]string{"node-v1": link}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if identity.Path != stub {
		t.Fatalf("resolved path = %q, want the physical file %q", identity.Path, stub)
	}
}

// TestResolveInterpreterRejectsSubstitution proves hard-link and non-regular
// bindings are refused before any hash is trusted, including through a
// launcher link: resolving the link does not trust what it reaches.
func TestResolveInterpreterRejectsSubstitution(t *testing.T) {
	stub := mustPhysical(t, stubInterpreterBinary(t))
	directory := mustPhysical(t, t.TempDir())

	// The target names the native image itself (with the platform
	// executable suffix on Windows) so the refusal below proves the
	// link-count gate, not the name gate.
	hardTarget := filepath.Join(directory, executableFixtureName("interp-hard"))
	copyTestFile(t, stub, hardTarget)
	if err := os.Link(hardTarget, filepath.Join(directory, executableFixtureName("interp-hard-alias"))); err != nil {
		t.Logf("this host cannot create a hard link, skipping that row: %v", err)
	} else {
		if _, err := ResolveInterpreter("node-v1", map[string]string{"node-v1": hardTarget}, nil); DiagnosticCode(err) != CodeWorkerIdentityInvalid {
			t.Fatalf("hard-link binding error = %v, want %s", err, CodeWorkerIdentityInvalid)
		}
		link := filepath.Join(directory, "interp-hard-link")
		if err := os.Symlink(hardTarget, link); err != nil {
			t.Skipf("this host forbids symlink creation: %v", err)
		}
		if _, err := ResolveInterpreter("node-v1", map[string]string{"node-v1": link}, nil); DiagnosticCode(err) != CodeWorkerIdentityInvalid {
			t.Fatalf("link-to-hard-link binding error = %v, want %s", err, CodeWorkerIdentityInvalid)
		}
	}

	if _, err := ResolveInterpreter("node-v1", map[string]string{"node-v1": directory}, nil); DiagnosticCode(err) != CodeWorkerIdentityInvalid {
		t.Fatalf("directory binding error = %v, want %s", err, CodeWorkerIdentityInvalid)
	}
	if _, err := ResolveInterpreter("node-v1", map[string]string{"node-v1": "relative/interp"}, nil); DiagnosticCode(err) != CodeWorkerIdentityInvalid {
		t.Fatalf("relative binding error = %v, want %s", err, CodeWorkerIdentityInvalid)
	}
	if _, err := ResolveInterpreter("node-v1", map[string]string{"node-v1": filepath.Join(directory, "absent")}, nil); DiagnosticCode(err) != CodeWorkerIdentityInvalid {
		t.Fatalf("absent binding error = %v, want %s", err, CodeWorkerIdentityInvalid)
	}
}

// TestVerifyInterpreterDetectsReplacement proves the launch-boundary
// re-proof catches bytes that changed after resolution.
func TestVerifyInterpreterDetectsReplacement(t *testing.T) {
	stub := stubInterpreterBinary(t)
	target := filepath.Join(mustPhysical(t, t.TempDir()), filepath.Base(stub))
	copyTestFile(t, stub, target)
	target = mustPhysical(t, target)
	identity, err := ResolveInterpreter("python3-v1", map[string]string{"python3-v1": target}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyInterpreter(identity); err != nil {
		t.Fatalf("unchanged interpreter failed re-proof: %v", err)
	}
	payload, err := os.ReadFile(target) // #nosec G304 -- test-owned path
	if err != nil {
		t.Fatal(err)
	}
	payload[0] ^= 0xff
	if err := os.WriteFile(target, payload, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := VerifyInterpreter(identity); DiagnosticCode(err) != CodeWorkerIdentityInvalid {
		t.Fatalf("replaced interpreter error = %v, want %s", err, CodeWorkerIdentityInvalid)
	}
}

// TestInterpreterIdentityReusesTheManagerPrimitives proves the interpreter
// canonicalization agrees with the go-v1 worker's on the same host: a path
// through a launcher link resolves to the same physical file both workers
// would hash.
func TestInterpreterIdentityReusesTheManagerPrimitives(t *testing.T) {
	stub := mustPhysical(t, stubInterpreterBinary(t))
	directory := mustPhysical(t, t.TempDir())
	link := filepath.Join(directory, "launcher")
	if err := os.Symlink(stub, link); err != nil {
		t.Skipf("this host forbids symlink creation: %v", err)
	}
	physical, err := godriver.CanonicalPhysicalPath(link)
	if err != nil {
		t.Fatal(err)
	}
	if physical != stub {
		t.Fatalf("canonical path = %q, want the physical file %q", physical, stub)
	}
}

// TestResolveInterpreterRejectsWrapperImage proves a binding that is not a
// native executable image is refused at resolution, before any worker
// exists. A POSIX shebang wrapper (the pyenv/asdf/volta shim shape) would
// interpose another program between the worker and the interpreter; on
// Windows a .cmd/.bat binding would run through cmd.exe, and an
// extensionless path would execute a different file than the verified one,
// so the binding must name the native .exe image itself.
func TestResolveInterpreterRejectsWrapperImage(t *testing.T) {
	directory := mustPhysical(t, t.TempDir())
	stub := mustPhysical(t, stubInterpreterBinary(t))
	type binding struct {
		name string
		path string
	}
	var bindings []binding
	if runtime.GOOS == "windows" {
		cmdWrapper := filepath.Join(directory, "interp.cmd")
		writeTestFile(t, cmdWrapper, []byte("@echo off\r\necho wrapper\r\n"), 0o644)
		bindings = append(bindings, binding{"cmd-wrapper", cmdWrapper})
		batWrapper := filepath.Join(directory, "interp.bat")
		writeTestFile(t, batWrapper, []byte("@echo off\r\necho wrapper\r\n"), 0o644)
		bindings = append(bindings, binding{"bat-wrapper", batWrapper})
		// A native image without the .exe extension passes the header
		// gate and fails the name gate: Go would execute a neighboring
		// PATHEXT sibling instead of this file.
		extensionless := filepath.Join(directory, "interp")
		copyTestFile(t, stub, extensionless)
		bindings = append(bindings, binding{"extensionless-native-image", extensionless})
		// A .exe name over non-image bytes fails the header gate: this is
		// the arm a dropped header check would admit.
		fake := filepath.Join(directory, "fake.exe")
		writeTestFile(t, fake, []byte("@echo off\r\necho wrapper\r\n"), 0o644)
		bindings = append(bindings, binding{"exe-named-text", fake})
	} else {
		wrapper := filepath.Join(directory, "interp")
		writeTestFile(t, wrapper, []byte("#!/bin/sh\nexec \""+stub+"\" \"$@\"\n"), 0o755)
		bindings = append(bindings, binding{"shebang-wrapper", wrapper})
		plain := filepath.Join(directory, "plain")
		writeTestFile(t, plain, []byte("not an executable image\n"), 0o755)
		bindings = append(bindings, binding{"plain-text", plain})
	}
	for _, testCase := range bindings {
		t.Run(testCase.name, func(t *testing.T) {
			path := mustPhysical(t, testCase.path)
			if _, err := ResolveInterpreter("python3-v1", map[string]string{"python3-v1": path}, nil); DiagnosticCode(err) != CodeWorkerIdentityInvalid {
				t.Fatalf("wrapper binding %q error = %v, want %s", testCase.name, err, CodeWorkerIdentityInvalid)
			}
		})
	}
}

// TestHasWindowsExecutableExtension pins the Windows name gate the
// resolution and worker boundaries share: only a binding that names the
// .exe itself is returned unchanged by Go's lookExtensions. It runs on
// every OS; the production-boundary arms run on windows-latest.
func TestHasWindowsExecutableExtension(t *testing.T) {
	for _, testCase := range []struct {
		path string
		want bool
	}{
		{`C:\tools\node.exe`, true},
		{`C:\tools\NODE.EXE`, true},
		{`C:\tools\Node.Exe`, true},
		{`\\server\share\python.exe`, true},
		{`C:\tools\node`, false},
		{`C:\tools\node.cmd`, false},
		{`C:\tools\node.bat`, false},
		{`C:\tools\node.com`, false},
		{`C:\tools\node.exe.backup`, false},
		{`C:\tools\nodeexe`, false},
	} {
		if got := hasWindowsExecutableExtension(testCase.path); got != testCase.want {
			t.Errorf("hasWindowsExecutableExtension(%q) = %v, want %v", testCase.path, got, testCase.want)
		}
	}
}
