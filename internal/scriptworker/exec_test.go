package scriptworker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestDefaultExecSearchDirsUsesManagerSystemRoot proves Windows executable
// lookup reads the captured manager environment, including its
// case-insensitive SYSTEMROOT name, and searches System32 before the root.
func TestDefaultExecSearchDirsUsesManagerSystemRoot(t *testing.T) {
	root := t.TempDir()
	dirs := defaultExecSearchDirs("windows", []string{"sYsTeMrOoT=" + root})
	want := []string{filepath.Join(root, "System32"), root}
	if len(dirs) != len(want) {
		t.Fatalf("Windows exec search dirs = %q, want %q", dirs, want)
	}
	for i := range want {
		if dirs[i] != want[i] {
			t.Fatalf("Windows exec search dirs = %q, want %q", dirs, want)
		}
	}
}

// TestDeriveProfileUsesDefaultExecSearchDirsAndBuildsDeclaredExecFarm
// exercises the same nil ExecSearchDirs path as Launch on Windows, with an
// injected platform and manager SYSTEMROOT so it also runs on Darwin/Linux.
// The extra NTFS-style hard link models the Windows component-store files
// that the production System32 exception must copy into the private farm.
func TestDeriveProfileUsesDefaultExecSearchDirsAndBuildsDeclaredExecFarm(t *testing.T) {
	fixture := newLaunchFixture(t, os.Args[0])
	root := mustPhysical(t, t.TempDir())
	system32 := filepath.Join(root, "System32")
	execPath := filepath.Join(system32, "cmd.exe")
	copyTestFile(t, fixture.stub, execPath)
	winsxs := filepath.Join(root, "WinSxS")
	if err := os.MkdirAll(winsxs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(execPath, filepath.Join(winsxs, "cmd.exe")); err != nil {
		t.Fatalf("create Windows component-store hard-link fixture: %v", err)
	}

	private, err := createPrivateArea(fixture.request.PrivateBase, fixture.request.ForbiddenRoots)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(private.base) }()
	interpreter, err := ResolveInterpreter("python3-v1", fixture.request.Interpreters, fixture.request.ForbiddenRoots)
	if err != nil {
		t.Fatal(err)
	}
	declared, err := ParseDeclaredCapabilities(json.RawMessage(`{"exec":["cmd.exe"]}`))
	if err != nil {
		t.Fatal(err)
	}
	profile, err := deriveProfileForPlatform(DerivationInput{
		Declared: declared, InterpreterID: "python3-v1", Interpreter: interpreter,
		HostEnv: []string{"sYsTeMrOoT=" + root, "PATH=/caller/path"},
		// Leave ExecSearchDirs nil: this is the behavior under qualification.
		ForbiddenRoots: fixture.request.ForbiddenRoots,
		Private:        private,
		FarmParent:     private.base,
	}, "windows")
	if err != nil {
		t.Fatalf("derive Windows profile from default manager dirs: %v", err)
	}

	wantResolved := mustPhysical(t, execPath)
	if got := profile.Report.ResolvedExec["cmd.exe"]; got != wantResolved {
		t.Fatalf("resolved cmd.exe = %q, want manager System32 file %q", got, wantResolved)
	}
	wantFarm := []string{filepath.Base(interpreter.Path), "cmd.exe"}
	sort.Strings(wantFarm)
	gotFarm := append([]string(nil), profile.Report.FarmEntries...)
	sort.Strings(gotFarm)
	if strings.Join(gotFarm, ",") != strings.Join(wantFarm, ",") {
		t.Fatalf("PATH farm entries = %q, want interpreter and declared exec %q", gotFarm, wantFarm)
	}

	// The hard-link exception is tied to the manager's default System32
	// search path; an explicit directory does not inherit that trust.
	explicit := DerivationInput{
		Declared: declared, InterpreterID: "python3-v1", Interpreter: interpreter,
		HostEnv: []string{"SYSTEMROOT=" + root}, ExecSearchDirs: []string{system32},
		ForbiddenRoots: fixture.request.ForbiddenRoots, Private: private, FarmParent: private.base,
	}
	explicitProfile, err := deriveProfileForPlatform(explicit, "windows")
	if err != nil {
		t.Fatalf("explicit search directory derivation refused: %v", err)
	}
	if got := strings.Join(explicitProfile.Report.UnresolvedExec, ","); got != "cmd.exe" {
		t.Fatalf("explicit search unresolved exec = %q, want cmd.exe because the hard-link allowance is default-list-only", got)
	}
	if _, found := explicitProfile.Report.ResolvedExec["cmd.exe"]; found {
		t.Fatalf("explicit search directory resolved the multiple-link cmd.exe: %q", explicitProfile.Report.ResolvedExec["cmd.exe"])
	}
}
