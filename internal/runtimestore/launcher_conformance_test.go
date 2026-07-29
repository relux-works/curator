package runtimestore

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/skillspec"
)

// authoritativeLauncherCase is the published launcher contract. Only the field
// names are mirrored here; every expected value is read from the authoritative
// document at run time, so this file holds no private copy of the suite.
type authoritativeLauncherCase struct {
	Name                  string   `json:"name"`
	Platforms             []string `json:"platforms"`
	ForwardArguments      bool     `json:"forward_arguments"`
	PreserveExitStatus    bool     `json:"preserve_exit_status"`
	PreserveInheritedPath bool     `json:"preserve_inherited_path"`
	RequiredPathRoles     []string `json:"required_path_roles"`
}

func authoritativeLauncherCases(t *testing.T) []authoritativeLauncherCase {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "manager-lifecycle.json")) // #nosec G304 -- explicit authoritative conformance input
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		LauncherCases []authoritativeLauncherCase `json:"launcher_cases"`
	}
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	if len(document.LauncherCases) == 0 {
		t.Fatal("the authoritative suite publishes no launcher case to bind")
	}
	return document.LauncherCases
}

// launcherBinding is the executable fixture bound to one published launcher
// case: the dependency kind the installed command actually resolves at run
// time, and the exit status its implementation returns.
type launcherBinding struct {
	// dependency names the output token the case is really about, so the two
	// published cases cannot pass on each other's evidence.
	dependency string
	exitStatus int
}

// launcherBindings maps every published launcher case to the executable
// fixture that proves it. A published case with no entry fails the binding
// test rather than being silently skipped.
var launcherBindings = map[string]launcherBinding{
	"skill-command-without-shell-activation":  {dependency: "skill-command", exitStatus: 23},
	"declared-system-command-without-profile": {dependency: "system-command", exitStatus: 61},
}

type launcherFixture struct {
	binDir      string
	shim        string
	runtimePath string
	storeEntry  string
	systemDir   string
	inherited   string
	pathEntries []string
}

func writeLauncherFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil { // #nosec G306 -- launcher fixtures must be executable
		t.Fatal(err)
	}
}

// launcherImplementation is the skill command body. It resolves one dependency
// per published path role by bare name, reports the PATH it received, reports
// the executable it was replaced by, echoes its exact argument vector, and
// exits with a fixed status. Resolving the inherited entry is tolerated so the
// same implementation covers both the inherited-PATH run and the run that
// carries no inherited PATH at all.
func launcherImplementation(exitStatus int) string {
	return "#!/bin/sh\n" +
		"sibling-tool\n" +
		"declared-helper\n" +
		"inherited-helper 2>/dev/null || printf 'no-inherited\\n'\n" +
		"printf 'PATH=<%s>\\n' \"${PATH-}\"\n" +
		"printf 'SELF=<%s>\\n' \"$0\"\n" +
		"printf 'ARGS='\n" +
		"for arg do printf '<%s>' \"$arg\"; done\n" +
		"printf '\\n'\n" +
		"exit " + strconv.Itoa(exitStatus) + "\n"
}

// newLauncherFixture installs one skill runtime root into the store and writes
// the managed launcher over it, carrying exactly the three published path
// roles: the command directory, the implementation runtime the launcher execs,
// and the directory of the declared system dependency.
func newLauncherFixture(t *testing.T, binding launcherBinding) launcherFixture {
	t.Helper()
	root := t.TempDir()
	home := filepath.Join(root, "manager home")
	snapshot := filepath.Join(root, "snapshot's dir")
	systemDir := filepath.Join(root, "system dependencies")
	inherited := filepath.Join(root, "inherited path")
	binDir := filepath.Join(root, "command directory")

	writeLauncherFile(t, filepath.Join(systemDir, "declared-helper"), "#!/bin/sh\nprintf 'system-command\\n'\n")
	writeLauncherFile(t, filepath.Join(inherited, "inherited-helper"), "#!/bin/sh\nprintf 'inherited-path\\n'\n")
	writeLauncherFile(t, filepath.Join(snapshot, "scripts", "sibling-tool"), "#!/bin/sh\nprintf 'skill-command\\n'\n")
	writeLauncherFile(t, filepath.Join(snapshot, "scripts", "launcher-tool"), launcherImplementation(binding.exitStatus))

	commit := "0123456789abcdef0123456789abcdef01234567"
	if _, err := InstallRuntimeRoots(home, "launcher-skill", commit, snapshot, []string{"scripts"}); err != nil {
		t.Fatal(err)
	}
	runtimePath, err := RuntimeCommandPath(home, "launcher-skill", commit,
		skillspec.Command{Name: "launcher-tool", Type: "script", UnixPath: "scripts/launcher-tool"}, "unix")
	if err != nil {
		t.Fatal(err)
	}
	siblingPath, err := RuntimeCommandPath(home, "launcher-skill", commit,
		skillspec.Command{Name: "sibling-tool", Type: "script", UnixPath: "scripts/sibling-tool"}, "unix")
	if err != nil {
		t.Fatal(err)
	}

	absoluteBin, err := filepath.Abs(binDir)
	if err != nil {
		t.Fatal(err)
	}
	// The same two roles the manager derives for a real installation: the
	// command directory first, then the directory of the declared system
	// command. The implementation-runtime role is the exec target itself.
	entries := []string{absoluteBin, systemDir}
	if _, err := WriteBinShim(binDir, "sibling-tool", siblingPath, "unix", entries); err != nil {
		t.Fatal(err)
	}
	shim, err := WriteBinShim(binDir, "launcher-tool", runtimePath, "unix", entries)
	if err != nil {
		t.Fatal(err)
	}
	return launcherFixture{
		binDir: absoluteBin, shim: shim, runtimePath: runtimePath,
		storeEntry: Dir(home, "launcher-skill", commit),
		systemDir:  systemDir, inherited: inherited, pathEntries: entries,
	}
}

// launcherArguments is one argument vector chosen so any shell re-splitting,
// quoting, or dropping of an empty argument is visible in the echo.
var launcherArguments = []string{"space value", `quote"value`, "percent%value", "$NOT_EXPANDED", "Юникод", ""}

func wantLauncherArgs() string {
	rendered := "ARGS="
	for _, argument := range launcherArguments {
		rendered += "<" + argument + ">"
	}
	return rendered
}

// TestAuthoritativeLauncherCasesForwardArgvPathRolesAndExitStatus binds every
// published launcher case to a real process launch of an installed command.
// The published booleans drive the assertions, and each published path role
// must have its own executable proof.
func TestAuthoritativeLauncherCasesForwardArgvPathRolesAndExitStatus(t *testing.T) {
	cases := authoritativeLauncherCases(t)
	for _, published := range cases {
		published := published
		binding, bound := launcherBindings[published.Name]
		if !bound {
			t.Fatalf("published launcher case %q has no executable binding", published.Name)
		}
		if len(published.Platforms) == 0 {
			t.Fatalf("published launcher case %q names no platform", published.Name)
		}
		t.Run(published.Name, func(t *testing.T) {
			for _, platform := range published.Platforms {
				switch platform {
				case "unix":
					if runtime.GOOS == "windows" {
						t.Log("the POSIX launcher is not executable on this host; the Windows form is asserted below")
						continue
					}
					runUnixLauncherCase(t, published, binding)
				case "windows":
					assertWindowsLauncherForm(t, published)
				default:
					t.Fatalf("published launcher case %q names an unbound platform %q", published.Name, platform)
				}
			}
		})
	}
}

func runUnixLauncherCase(t *testing.T, published authoritativeLauncherCase, binding launcherBinding) {
	t.Helper()
	fixture := newLauncherFixture(t, binding)

	// Every published path role needs its own executable proof; an unknown role
	// must fail rather than pass by omission.
	for _, role := range published.RequiredPathRoles {
		switch role {
		case "command_directory":
			if filepath.Dir(fixture.shim) != fixture.binDir {
				t.Fatalf("launcher %s is not published into the command directory %s", fixture.shim, fixture.binDir)
			}
		case "implementation_runtime":
			if !strings.HasPrefix(fixture.runtimePath, fixture.storeEntry+string(os.PathSeparator)) {
				t.Fatalf("implementation %s is not inside the runtime store entry %s",
					fixture.runtimePath, fixture.storeEntry)
			}
		case "system_dependencies":
			if !containsPath(fixture.pathEntries, fixture.systemDir) {
				t.Fatalf("path entries %v omit the declared system dependency directory %s",
					fixture.pathEntries, fixture.systemDir)
			}
		default:
			t.Fatalf("published path role %q has no executable binding", role)
		}
	}

	// Run one: an inherited PATH is present, which is the only run that can
	// prove the launcher preserves it instead of replacing it.
	inheritedOutput, inheritedCode := runLauncher(t, fixture.shim, []string{"PATH=" + fixture.inherited})
	if published.PreserveExitStatus && inheritedCode != binding.exitStatus {
		t.Fatalf("inherited-PATH launch exit = %d, want the implementation's %d\n%s",
			inheritedCode, binding.exitStatus, inheritedOutput)
	}
	if published.PreserveInheritedPath {
		wantPath := "PATH=<" + strings.Join(fixture.pathEntries, ":") + ":" + fixture.inherited + ">"
		if !strings.Contains(inheritedOutput, wantPath) {
			t.Fatalf("launcher did not prefix its roles onto the inherited PATH; want %q:\n%s",
				wantPath, inheritedOutput)
		}
		if !strings.Contains(inheritedOutput, "inherited-path") {
			t.Fatalf("an inherited PATH entry stopped resolving under the launcher:\n%s", inheritedOutput)
		}
	}
	if published.ForwardArguments && !strings.Contains(inheritedOutput, wantLauncherArgs()) {
		t.Fatalf("argument vector was not forwarded exactly; want %q:\n%s", wantLauncherArgs(), inheritedOutput)
	}
	if !strings.Contains(inheritedOutput, "SELF=<"+fixture.runtimePath+">") {
		t.Fatalf("the launcher did not exec the installed implementation %s:\n%s",
			fixture.runtimePath, inheritedOutput)
	}
	if !strings.Contains(inheritedOutput, binding.dependency) {
		t.Fatalf("the %s dependency this case is about did not resolve:\n%s", binding.dependency, inheritedOutput)
	}

	// Run two: no inherited PATH and no profile of any kind. Both published
	// dependency kinds must still resolve, because the launcher carries them.
	bareOutput, bareCode := runLauncher(t, fixture.shim, []string{})
	if published.PreserveExitStatus && bareCode != binding.exitStatus {
		t.Fatalf("no-activation launch exit = %d, want %d\n%s", bareCode, binding.exitStatus, bareOutput)
	}
	if published.ForwardArguments && !strings.Contains(bareOutput, wantLauncherArgs()) {
		t.Fatalf("argument vector was not forwarded without an inherited PATH:\n%s", bareOutput)
	}
	for _, resolved := range []string{"skill-command", "system-command"} {
		if !strings.Contains(bareOutput, resolved) {
			t.Fatalf("%s did not resolve without shell activation or a profile:\n%s", resolved, bareOutput)
		}
	}
	if !strings.Contains(bareOutput, "no-inherited") {
		t.Fatalf("the run without an inherited PATH still reached an inherited entry:\n%s", bareOutput)
	}
	// Without an inherited PATH the shell supplies its own default, so the
	// contract is that the published roles lead it, not that they are alone.
	if !strings.Contains(bareOutput, "PATH=<"+strings.Join(fixture.pathEntries, ":")) {
		t.Fatalf("the launcher did not lead with its own roles without an inherited PATH:\n%s", bareOutput)
	}
	if strings.Contains(bareOutput, fixture.inherited) {
		t.Fatalf("a PATH entry that was never inherited reached the implementation:\n%s", bareOutput)
	}
}

// assertWindowsLauncherForm asserts the published contract against the exact
// managed Windows launcher bytes. Those bytes are testable on any host; only
// their execution needs a Windows machine, which the report records.
func assertWindowsLauncherForm(t *testing.T, published authoritativeLauncherCase) {
	t.Helper()
	runtimePath := `C:\manager home\runtime\launcher-skill\0123\scripts\launcher-tool.cmd`
	systemDir := `C:\system dependencies`
	binDir := `C:\project\.agents\bin`
	content := WindowsShimContent(runtimePath, []string{binDir, systemDir})
	for _, role := range published.RequiredPathRoles {
		switch role {
		case "command_directory":
			if !strings.Contains(content, `set "PATH=`+binDir+`;`) {
				t.Fatalf("Windows launcher omits the command directory:\n%s", content)
			}
		case "implementation_runtime":
			if !strings.Contains(content, `call "`+runtimePath+`" %*`) {
				t.Fatalf("Windows launcher does not call the installed implementation:\n%s", content)
			}
		case "system_dependencies":
			if !strings.Contains(content, systemDir+`;%PATH%`) {
				t.Fatalf("Windows launcher omits the declared system dependency directory:\n%s", content)
			}
		default:
			t.Fatalf("published path role %q has no Windows binding", role)
		}
	}
	if published.ForwardArguments && !strings.Contains(content, `%*`) {
		t.Fatalf("Windows launcher does not forward its argument vector:\n%s", content)
	}
	if published.PreserveExitStatus && !strings.Contains(content, "exit /b %ERRORLEVEL%") {
		t.Fatalf("Windows launcher does not return the implementation's exit status:\n%s", content)
	}
	if published.PreserveInheritedPath && !strings.Contains(content, `;%PATH%`) {
		t.Fatalf("Windows launcher replaces the inherited PATH:\n%s", content)
	}
}

func runLauncher(t *testing.T, shim string, environment []string) (string, int) {
	t.Helper()
	command := exec.Command(shim, launcherArguments...) // #nosec G204 -- the launcher under test
	command.Env = environment
	output, runErr := command.CombinedOutput()
	if runErr == nil {
		return string(output), 0
	}
	var exitErr *exec.ExitError
	if !errors.As(runErr, &exitErr) {
		t.Fatalf("launching %s failed before it could report a status: %v\n%s", shim, runErr, output)
	}
	return string(output), exitErr.ExitCode()
}

func containsPath(entries []string, want string) bool {
	for _, entry := range entries {
		if entry == want {
			return true
		}
	}
	return false
}
