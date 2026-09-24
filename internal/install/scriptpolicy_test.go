package install

import (
	"bytes"
	"encoding/json"
	"errors"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/scriptpolicy"
	"github.com/relux-works/curator/internal/scriptworker"
	"github.com/relux-works/curator/internal/skillspec"
)

// errInjectedProbeFailure fails one native-control probe through the
// production probe path, so install tests prove the host preflight
// refuses without starting any worker.
var errInjectedProbeFailure = errors.New("injected probe failure")

// nativeLauncherName is the enforced native launcher filename: extensionless
// on unix, `.exe` on Windows — never the ordinary `.cmd` shim name, which
// the enforced policy forbids.
func nativeLauncherName(command string) string {
	if runtime.GOOS == "windows" {
		return command + ".exe"
	}
	return command
}

// schema8ScriptSkill creates a tagged schema-8 skill whose single script
// command is enforced or declared-only. It is the install-side twin of the
// suite's valid-script-worker-enforced schema case. The enforced arm binds
// the given closed interpreter identifier.
func (e *env) schema8ScriptSkill(name string, enforced bool, interpreter ...string) {
	e.t.Helper()
	dir := filepath.Join(e.skillsRoot, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		e.t.Fatal(err)
	}
	e.git(dir, "init", "-q", "-b", "main")
	e.write(dir, "SKILL.md", "---\nname: "+name+"\ndescription: d\n---\n# "+name+"\n")
	e.write(dir, "scripts/"+name+"-tool", "#!/bin/sh\necho "+name+"\n")
	command := map[string]any{
		"type":      "script",
		"unix_path": "scripts/" + name + "-tool",
		"win_path":  "scripts/" + name + "-tool",
	}
	if enforced {
		bound := "python3-v1"
		if len(interpreter) != 0 {
			bound = interpreter[0]
		}
		command["execution_policy"] = skillspec.ScriptExecutionPolicy
		command["interpreter"] = bound
	}
	payload, _ := json.MarshalIndent(map[string]any{
		"schema_version": 8,
		"capabilities": map[string]any{
			"env_read": []string{}, "exec": "none", "filesystem": "repo",
			"network": "none", "secrets": "none",
		},
		"runtime_roots": []string{"scripts"},
		"commands":      map[string]any{name + "-tool": command},
	}, "", "  ")
	e.write(dir, "agent-skill.json", string(payload))
	e.git(dir, "add", ".")
	e.git(dir, "commit", "-qm", "init")
	e.git(dir, "tag", "v1")
}

// TestMain removes the shared enforced-fixture build directory after the
// run, so the ~20 MB production-binary fixtures do not leak into the
// system temp dir on every package run.
func TestMain(m *testing.M) {
	code := m.Run()
	if enforcedOnce.directory != "" {
		_ = os.RemoveAll(enforcedOnce.directory)
	}
	os.Exit(code)
}

// An enforced command must not reach a shim when this host cannot provide
// a mandatory control. Admission proceeds (the control table is complete),
// then the install-time host preflight refuses with
// `script_execution_control_unavailable` naming the unavailable control,
// for each closed interpreter identifier — and publishes nothing.
func TestEnforcedScriptCommandIsRefusedAtInstall(t *testing.T) {
	// Sequential by construction: the test forces the process-global probe
	// fault. Never add t.Parallel here.
	defer scriptworker.OverrideScriptProbeFaultForTest(func(control string) error {
		if control == scriptworker.ScriptControlDescendantDomainTermination {
			return errInjectedProbeFailure
		}
		return nil
	})()
	for _, interpreter := range []string{"node-v1", "python3-v1"} {
		t.Run(interpreter, func(t *testing.T) {
			e := newEnv(t)
			e.schema8ScriptSkill("enforced-skill", true, interpreter)
			e.declare("enforced-skill")

			result := e.install(Options{})
			if result.Status == "ok" {
				t.Fatalf("an enforced script command installed: %+v", result)
			}
			reported := strings.Join(append(result.Errors, result.Messages...), "\n")
			if !strings.Contains(reported, scriptpolicy.ControlUnavailable) {
				t.Fatalf("install did not report %s: %s", scriptpolicy.ControlUnavailable, reported)
			}
			if !strings.Contains(reported, scriptworker.ScriptControlDescendantDomainTermination) {
				t.Fatalf("install refusal does not name the unavailable control: %s", reported)
			}
			// The forbidden outcome is the shim, not just the exit status: a partial
			// install that still published the launcher would run the package code
			// uncontained even though the run reported a failure.
			shim := filepath.Join(e.project, ".agents", "bin", shimName("enforced-skill-tool"))
			if _, err := os.Lstat(shim); err == nil {
				t.Fatal("a shim was published for an enforced script command")
			}
			if runtime.GOOS == "windows" {
				native := filepath.Join(e.project, ".agents", "bin", nativeLauncherName("enforced-skill-tool"))
				if _, err := os.Lstat(native); err == nil {
					t.Fatal("a native launcher was published for a refused enforced script command")
				}
				if _, err := os.Lstat(native + scriptworker.ShimSidecarSuffix); err == nil {
					t.Fatal("a sidecar was published for a refused enforced script command")
				}
			}
		})
	}
}

// The control: schema 8 by itself changes nothing for a command that did not
// opt in, so a declared-only script command installs exactly as before.
func TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.schema8ScriptSkill("plain-skill", false)
	e.declare("plain-skill")

	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("a declared-only schema-8 skill failed to install: %+v", result)
	}
	shim := filepath.Join(e.project, ".agents", "bin", shimName("plain-skill-tool"))
	if _, err := os.Lstat(shim); err != nil {
		t.Fatalf("shim missing for a declared-only schema-8 command: %v", err)
	}
	if _, err := os.Stat(filepath.Join(e.project, ".agents", "skills", "plain-skill", "SKILL.md")); err != nil {
		t.Fatalf("context missing: %v", err)
	}
}

// The shim writer is the last layer that could turn an enforced command into
// an uncontained launcher: the declared-only writer never sees an enforced
// command at all, and the enforced writer lists it for native launcher
// staging once admission and the host preflight pass.
func TestActiveScriptCommandWritersSplitDeclaredOnlyAndEnforced(t *testing.T) {
	t.Parallel()
	node := &closure.Node{
		Name: "skill", Spec: &skillspec.Spec{Commands: map[string]skillspec.Command{
			"plain": {Name: "plain", Type: "script", UnixPath: "scripts/plain"},
			"enforced": {
				Name: "enforced", Type: "script", UnixPath: "scripts/enforced",
				ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "node-v1",
			},
		}},
		Edges: []closure.Edge{{Consumer: closure.ProjectEdge, Mode: "full"}},
	}
	// The declared-only writer skips enforced commands: they stage as
	// native launchers, never as ordinary shims.
	declared, err := activeScriptCommands(node, node.ActiveCommands())
	if err != nil {
		t.Fatalf("declared-only writer refused: %v", err)
	}
	if len(declared) != 1 || declared[0].Name != "plain" {
		t.Fatalf("declared commands = %+v, want only plain", declared)
	}
	// The enforced writer lists the command for native launcher staging.
	enforced, err := activeEnforcedScriptCommands(node, node.ActiveCommands())
	if err != nil {
		t.Fatalf("enforced writer refused: %v", err)
	}
	if len(enforced) != 1 || enforced[0].Name != "enforced" {
		t.Fatalf("enforced commands = %+v, want the enforced command", enforced)
	}

	// An unknown policy keeps the original refusal at this layer too.
	node.Spec.Commands["enforced"] = skillspec.Command{
		Name: "enforced", Type: "script", UnixPath: "scripts/enforced",
		ExecutionPolicy: "script-worker-v9", Interpreter: "node-v1",
	}
	_, err = activeEnforcedScriptCommands(node, node.ActiveCommands())
	if err == nil {
		t.Fatal("the enforced writer accepted an unknown execution policy")
	}
	if scriptpolicy.Code(err) != scriptpolicy.PolicyUnsupported {
		t.Fatalf("Code = %q, want %q", scriptpolicy.Code(err), scriptpolicy.PolicyUnsupported)
	}
	if !strings.Contains(err.Error(), "skill.enforced") {
		t.Fatalf("refusal does not name the command: %v", err)
	}
	if strings.Contains(err.Error(), "%!w(") {
		t.Fatalf("refusal formats a nil admission: %v", err)
	}

	// An enforced command that no edge activates is not staged at all, so
	// the remaining declared-only command still reaches the writer normally.
	node.Spec.Commands["enforced"] = skillspec.Command{
		Name: "enforced", Type: "script", UnixPath: "scripts/enforced",
		ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "node-v1",
	}
	node.Edges = []closure.Edge{{Consumer: closure.ProjectEdge, Mode: "runtime", Commands: []string{"plain"}}}
	commands, err := activeScriptCommands(node, node.ActiveCommands())
	if err != nil {
		t.Fatalf("declared-only command refused: %v", err)
	}
	if len(commands) != 1 || commands[0].Name != "plain" {
		t.Fatalf("commands = %+v", commands)
	}
	if enforced, err := activeEnforcedScriptCommands(node, node.ActiveCommands()); err != nil || len(enforced) != 0 {
		t.Fatalf("inactive enforced commands staged: %+v, %v", enforced, err)
	}
}

// TestActiveEnforcedScriptCommandsAdmits proves the enforced writer
// proceeds once admission and the host preflight succeed: the guard
// returns the commands for native launcher staging instead of an error,
// so it can never degrade to a wrapped nil.
func TestActiveEnforcedScriptCommandsAdmits(t *testing.T) {
	t.Parallel()
	node := &closure.Node{
		Name: "skill", Spec: &skillspec.Spec{Commands: map[string]skillspec.Command{
			"enforced": {
				Name: "enforced", Type: "script", UnixPath: "scripts/enforced",
				ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "node-v1",
			},
		}},
		Edges: []closure.Edge{{Consumer: closure.ProjectEdge, Mode: "full"}},
	}
	commands, err := activeEnforcedScriptCommands(node, node.ActiveCommands())
	if err != nil {
		t.Fatalf("enforced writer refused: %v", err)
	}
	if len(commands) != 1 || commands[0].Name != "enforced" {
		t.Fatalf("commands = %+v, want the enforced command", commands)
	}
}

// TestActiveEnforcedScriptCommandsRefusesUnavailableHost proves the
// enforced writer refuses with control-unavailable when this host cannot
// provide a mandatory control, naming the node and the control.
func TestActiveEnforcedScriptCommandsRefusesUnavailableHost(t *testing.T) {
	// Sequential by construction: the test forces the process-global probe
	// fault. Never add t.Parallel here.
	defer scriptworker.OverrideScriptProbeFaultForTest(func(control string) error {
		if control == scriptworker.ScriptControlDescendantDomainTermination {
			return errInjectedProbeFailure
		}
		return nil
	})()
	node := &closure.Node{
		Name: "skill", Spec: &skillspec.Spec{Commands: map[string]skillspec.Command{
			"enforced": {
				Name: "enforced", Type: "script", UnixPath: "scripts/enforced",
				ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "node-v1",
			},
		}},
		Edges: []closure.Edge{{Consumer: closure.ProjectEdge, Mode: "full"}},
	}
	_, err := activeEnforcedScriptCommands(node, node.ActiveCommands())
	if err == nil {
		t.Fatal("the enforced writer accepted a command the host cannot run")
	}
	if scriptpolicy.Code(err) != scriptpolicy.ControlUnavailable {
		t.Fatalf("Code = %q, want %q", scriptpolicy.Code(err), scriptpolicy.ControlUnavailable)
	}
	if !strings.Contains(err.Error(), "skill") ||
		!strings.Contains(err.Error(), scriptworker.ScriptControlDescendantDomainTermination) {
		t.Fatalf("refusal does not name the node and the control: %v", err)
	}
}

// schema8EnforcedSkillWithCaps creates a tagged schema-8 skill whose single
// script command is enforced with the given declared capabilities object.
func (e *env) schema8EnforcedSkillWithCaps(name string, capabilities map[string]any) {
	e.t.Helper()
	dir := filepath.Join(e.skillsRoot, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		e.t.Fatal(err)
	}
	e.git(dir, "init", "-q", "-b", "main")
	e.write(dir, "SKILL.md", "---\nname: "+name+"\ndescription: d\n---\n# "+name+"\n")
	e.write(dir, "scripts/"+name+"-tool", "print('tool')\n")
	payload, _ := json.MarshalIndent(map[string]any{
		"schema_version": 8,
		"capabilities":   capabilities,
		"runtime_roots":  []string{"scripts"},
		"commands": map[string]any{name + "-tool": map[string]any{
			"type": "script", "unix_path": "scripts/" + name + "-tool", "win_path": "scripts/" + name + "-tool",
			"execution_policy": skillspec.ScriptExecutionPolicy, "interpreter": "python3-v1",
		}},
	}, "", "  ")
	e.write(dir, "agent-skill.json", string(payload))
	e.git(dir, "add", ".")
	e.git(dir, "commit", "-qm", "init")
	e.git(dir, "tag", "v1")
}

var enforcedOnce struct {
	sync.Once
	curator   string
	stub      string
	directory string
	err       error
}

// enforcedBinaries builds the production curator and the stub interpreter
// once per package test run: the enforced install test publishes a real
// native launcher copied from the production binary and runs it against
// the stub.
func enforcedBinaries(t *testing.T) (curator, stub string) {
	t.Helper()
	enforcedOnce.Do(func() {
		directory, err := os.MkdirTemp("", "curator-enforced-")
		if err != nil {
			enforcedOnce.err = err
			return
		}
		enforcedOnce.directory = directory
		suffix := ""
		if runtime.GOOS == "windows" {
			suffix = ".exe"
		}
		goBinary := filepath.Join(build.Default.GOROOT, "bin", "go"+suffix)
		curatorOut := filepath.Join(directory, "curator"+suffix)
		buildCurator := exec.Command(goBinary, "build", "-o", curatorOut, "../../cmd/curator") // #nosec G204 -- fixed test-only build of the production binary under test
		buildCurator.Env = os.Environ()
		if combined, err := buildCurator.CombinedOutput(); err != nil {
			enforcedOnce.err = err
			t.Logf("curator build output: %s", combined)
			return
		}
		stubOut := filepath.Join(directory, "stubinterp"+suffix)
		buildStub := exec.Command(goBinary, "build", "-o", stubOut, "../scriptworker/testdata/stubinterp") // #nosec G204 -- fixed test-only build of the local test double
		buildStub.Env = os.Environ()
		if combined, err := buildStub.CombinedOutput(); err != nil {
			enforcedOnce.err = err
			t.Logf("stub build output: %s", combined)
			return
		}
		enforcedOnce.curator, enforcedOnce.stub = curatorOut, stubOut
	})
	if enforcedOnce.err != nil {
		t.Fatalf("cannot build the enforced fixtures: %v", enforcedOnce.err)
	}
	return enforcedOnce.curator, enforcedOnce.stub
}

// TestEnforcedScriptCommandInstallsNativeLauncher proves the production
// install path end to end: install publishes a native launcher plus its
// sidecar contract (no shell, `.cmd`, or symlink shim), records the
// derivation, and the published launcher runs the derived profile through
// the real worker.
func TestEnforcedScriptCommandInstallsNativeLauncher(t *testing.T) {
	curator, stub := enforcedBinaries(t)
	t.Cleanup(runtimestore.SetManagerBinaryForTest(curator))

	e := newEnv(t)
	e.schema8EnforcedSkillWithCaps("enforced-skill", map[string]any{
		"network":    []string{"api.example.com"},
		"exec":       []string{"no-such-tool-zzz-cur"},
		"secrets":    []string{"release-token"},
		"env_read":   []string{"LD_PRELOAD", "HTTP_PROXY", "STUB_EXIT"},
		"filesystem": "repo",
	})
	e.declare("enforced-skill")

	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("enforced install failed: %+v", result)
	}
	reported := strings.Join(append(result.Errors, result.Messages...), "\n")
	for _, want := range []string{
		"is enforced (script-worker-v1)",
		"withholds manager-owned env_read entries: HTTP_PROXY, LD_PRELOAD",
		"declares network hosts, recorded reporting-only without filtering: api.example.com",
		"omits unresolvable exec names from PATH: no-such-tool-zzz-cur",
	} {
		if !strings.Contains(reported, want) {
			t.Fatalf("install record does not report %q: %s", want, reported)
		}
	}

	binDir := filepath.Join(e.project, ".agents", "bin")
	launcher := filepath.Join(binDir, "enforced-skill-tool")
	if runtime.GOOS == "windows" {
		launcher += ".exe"
	}
	info, err := os.Lstat(launcher)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("native launcher is not a regular file: %v", err)
	}
	header, err := os.ReadFile(launcher) // #nosec G304 -- staged launcher under test
	if err != nil || len(header) < 8 {
		t.Fatalf("cannot read the staged launcher: %v", err)
	}
	if !godriver.NativeExecutableHeader(header[:8]) {
		t.Fatal("the enforced launcher is not a native executable image")
	}
	if runtime.GOOS == "windows" {
		if _, err := os.Lstat(filepath.Join(binDir, "enforced-skill-tool.cmd")); err == nil {
			t.Fatal("a .cmd wrapper was published for an enforced command")
		}
	}
	sidecarPath := launcher + scriptworker.ShimSidecarSuffix
	sidecar, err := scriptworker.LoadShimSidecar(sidecarPath)
	if err != nil {
		t.Fatalf("cannot load the published sidecar: %v", err)
	}
	if sidecar.Skill != "enforced-skill" || sidecar.Command != "enforced-skill-tool" ||
		sidecar.Interpreter != "python3-v1" || sidecar.SchemaVersion != 8 {
		t.Fatalf("sidecar identity = %+v", sidecar)
	}

	// The published launcher runs the derived profile through the real
	// worker: invoked in-process, the session the launcher starts is the
	// production session over the installed artifacts.
	var stdout, stderr bytes.Buffer
	code := scriptworker.RunShim(scriptworker.ShimRequest{
		ExePath: launcher,
		Args:    []string{"hello"},
		Stdin:   strings.NewReader("piped-input\n"),
		Stdout:  &stdout,
		Stderr:  &stderr,
		Environ: []string{
			"STUB_EXIT=0",
			"LD_PRELOAD=/host/evil.so",
			"HTTP_PROXY=http://host-proxy:8080",
			"PATH=/host/bin",
			"HOME=/host/home",
		},
		Interpreters: map[string]string{"python3-v1": stub},
	})
	if code != 0 {
		t.Fatalf("launcher exit = %d, stderr %q", code, stderr.String())
	}
	var observed struct {
		Argv  []string          `json:"argv"`
		Cwd   string            `json:"cwd"`
		Stdin string            `json:"stdin"`
		Env   map[string]string `json:"env"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &observed); err != nil {
		t.Fatalf("cannot decode the stub report %q: %v", stdout.String(), err)
	}
	if len(observed.Argv) != 3 || observed.Argv[2] != "hello" {
		t.Fatalf("argv = %q, want the verbatim argument", observed.Argv)
	}
	if observed.Stdin != "piped-input\n" {
		t.Fatalf("stdin = %q, want the piped payload", observed.Stdin)
	}
	physicalProject, err := godriver.CanonicalPhysicalPath(e.project)
	if err != nil {
		t.Fatal(err)
	}
	if observed.Cwd != physicalProject {
		t.Fatalf("cwd = %q, want the project root %q", observed.Cwd, physicalProject)
	}
	if observed.Env["CSK_PROJECT_ROOT"] != physicalProject {
		t.Fatalf("CSK_PROJECT_ROOT = %q, want the project root", observed.Env["CSK_PROJECT_ROOT"])
	}
	if _, present := observed.Env["LD_PRELOAD"]; present {
		t.Fatal("a withheld reserved name reached the interpreter")
	}
	if _, present := observed.Env["HTTP_PROXY"]; present {
		t.Fatal("a withheld proxy value reached the interpreter")
	}
	if observed.Env["STUB_EXIT"] != "0" {
		t.Fatal("the declared passthrough did not reach the interpreter")
	}
	if strings.Contains(observed.Env["PATH"], string(os.PathListSeparator)) {
		t.Fatalf("PATH = %q, want exactly the farm", observed.Env["PATH"])
	}

	// The same artifact in a fresh process without an operator-bound
	// interpreter refuses control-unavailable: resolution cannot be
	// applied on this host.
	launched := exec.Command(launcher)
	launched.Env = append(os.Environ(), "CURATOR_CONFIG="+filepath.Join(t.TempDir(), "config.json"))
	launched.Dir = t.TempDir()
	shimOutput, shimErr := launched.CombinedOutput()
	if shimErr == nil {
		t.Fatalf("the installed launcher ran past the production preflight: %s", shimOutput)
	}
	if !strings.Contains(string(shimOutput), scriptpolicy.ControlUnavailable) {
		t.Fatalf("installed launcher output does not name %s: %s", scriptpolicy.ControlUnavailable, shimOutput)
	}
}

// TestScriptOptInCasesAtInstallEntry drives the six published opt-in cases
// through the production install entry: accepted shapes install — a native
// launcher for the enforced shape, an ordinary shim for declared-only
// shapes — and rejected shapes refuse before anything is published.
func TestScriptOptInCasesAtInstallEntry(t *testing.T) {
	t.Parallel()
	writeOptInSkill := func(e *env, name string, schema int, command map[string]any) {
		e.t.Helper()
		dir := filepath.Join(e.skillsRoot, name)
		if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
			e.t.Fatal(err)
		}
		e.git(dir, "init", "-q", "-b", "main")
		e.write(dir, "SKILL.md", "---\nname: "+name+"\ndescription: d\n---\n# "+name+"\n")
		e.write(dir, "scripts/tool", "print('tool')\n")
		spec := map[string]any{
			"schema_version": schema,
			"capabilities":   map[string]any{},
			"runtime_roots":  []string{"scripts"},
			"commands":       map[string]any{name + "-tool": command},
		}
		payload, err := json.MarshalIndent(spec, "", "  ")
		if err != nil {
			e.t.Fatal(err)
		}
		e.write(dir, "agent-skill.json", string(payload))
		e.git(dir, "add", ".")
		e.git(dir, "commit", "-qm", "init")
		e.git(dir, "tag", "v1")
	}
	scriptCommand := func(extra map[string]any) map[string]any {
		command := map[string]any{"type": "script", "unix_path": "scripts/tool", "win_path": "scripts/tool"}
		for key, value := range extra {
			command[key] = value
		}
		return command
	}
	for _, testCase := range []struct {
		name    string
		schema  int
		command map[string]any
		// wantLauncher selects the native launcher; otherwise the
		// ordinary shim. wantCode names the refusal the rejected shapes
		// report: every rejected shape refuses at manifest parse time,
		// naming its field.
		wantLauncher bool
		wantCode     string
	}{
		{name: "schema8-explicit-opt-in", schema: 8,
			command:      scriptCommand(map[string]any{"execution_policy": "script-worker-v1", "interpreter": "python3-v1"}),
			wantLauncher: true},
		{name: "schema8-absent-policy", schema: 8, command: scriptCommand(nil)},
		{name: "legacy-schema7-script", schema: 7, command: scriptCommand(nil)},
		{name: "interpreter-without-policy", schema: 8,
			command:  scriptCommand(map[string]any{"interpreter": "python3-v1"}),
			wantCode: "interpreter"},
		{name: "policy-without-interpreter", schema: 8,
			command:  scriptCommand(map[string]any{"execution_policy": "script-worker-v1"}),
			wantCode: "interpreter"},
		{name: "unknown-policy", schema: 8,
			command:  scriptCommand(map[string]any{"execution_policy": "script-worker-v9", "interpreter": "python3-v1"}),
			wantCode: "execution_policy"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			skill := "optin-" + strings.ReplaceAll(testCase.name, "-opt-in", "")
			e := newEnv(t)
			writeOptInSkill(e, skill, testCase.schema, testCase.command)
			e.declare(skill)

			result := e.install(Options{})
			reported := strings.Join(append(result.Errors, result.Messages...), "\n")
			if testCase.wantCode != "" {
				if result.Status == "ok" {
					t.Fatalf("a rejected opt-in shape installed: %+v", result)
				}
				if !strings.Contains(reported, testCase.wantCode) {
					t.Fatalf("install did not report %s: %s", testCase.wantCode, reported)
				}
				return
			}
			if result.Status != "ok" {
				t.Fatalf("an accepted opt-in shape refused: %+v", result)
			}
			name := shimName(skill + "-tool")
			if testCase.wantLauncher {
				name = nativeLauncherName(skill + "-tool")
			}
			launcher := filepath.Join(e.project, ".agents", "bin", name)
			if _, err := os.Lstat(launcher); err != nil {
				t.Fatalf("no launcher was published: %v", err)
			}
			if testCase.wantLauncher {
				info, err := os.Lstat(launcher)
				if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
					t.Fatalf("the enforced shape did not publish a native launcher: %v", err)
				}
				if _, err := os.Lstat(launcher + scriptworker.ShimSidecarSuffix); err != nil {
					t.Fatalf("the enforced shape published no sidecar: %v", err)
				}
			} else if _, err := os.Lstat(launcher + scriptworker.ShimSidecarSuffix); err == nil {
				t.Fatal("a declared-only shape published a native launcher sidecar")
			}
		})
	}
}

// TestGlobalEnforcedInstallExcludesForwardingMirror proves the
// forwarding-mirror exclusion for global scope: an enforced command
// installs its canonical native launcher into the global bin, and no
// shell-wrapper forwarding entry for it appears in the user bin.
func TestGlobalEnforcedInstallExcludesForwardingMirror(t *testing.T) {
	e := newEnv(t)
	e.schema8EnforcedSkillWithCaps("global-enforced", map[string]any{})
	if _, err := GlobalInit(e.home); err != nil {
		t.Fatal(err)
	}
	if err := manifestAddGlobal(e, "global-enforced"); err != nil {
		t.Fatal(err)
	}
	userHome := t.TempDir()
	userBin := filepath.Join(userHome, ".local", "bin")
	if runtime.GOOS != "windows" {
		if err := os.MkdirAll(userBin, 0o755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", userBin+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
	result := Global(e.cfg, userHome, Options{Platform: installPlatform()})
	if result.Status != "ok" {
		t.Fatalf("global install: %+v", result)
	}
	launcher := filepath.Join(e.home, "global", "bin", "global-enforced-tool")
	if runtime.GOOS == "windows" {
		launcher += ".exe"
	}
	info, err := os.Lstat(launcher)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("global native launcher is not a regular file: %v", err)
	}
	if _, err := os.Lstat(launcher + scriptworker.ShimSidecarSuffix); err != nil {
		t.Fatalf("global launcher sidecar missing: %v", err)
	}
	if runtime.GOOS != "windows" {
		if _, err := os.Lstat(filepath.Join(userBin, "global-enforced-tool")); err == nil {
			t.Fatal("a forwarding wrapper was published for an enforced command")
		}
	}
}

// TestEnforcedInstallAndLaunchAtCLIEntry drives the enforced shape
// through the real command-line entry: the built curator binary installs
// the project by argv dispatch, and the installed native launcher executes
// the script end to end with the operator-bound interpreter.
func TestEnforcedInstallAndLaunchAtCLIEntry(t *testing.T) {
	curator, stub := enforcedBinaries(t)
	e := newEnv(t)
	e.schema8EnforcedSkillWithCaps("cli-enforced", map[string]any{})
	e.declare("cli-enforced")

	cliConfig, err := json.Marshal(map[string]any{
		"schema_version":      1,
		"skills_root":         e.skillsRoot,
		"projects":            map[string]any{"cli": map[string]any{"path": e.project}},
		"script_interpreters": map[string]string{"python3-v1": stub},
	})
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, cliConfig, 0o644); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	cliEnv := func() []string {
		return append(os.Environ(),
			"HOME="+home,
			"CURATOR_CONFIG="+configPath,
		)
	}
	installCLI := exec.Command(curator, "install", e.project)
	installCLI.Env = cliEnv()
	installCLI.Dir = t.TempDir()
	if output, err := installCLI.CombinedOutput(); err != nil {
		t.Fatalf("curator install refused: %v: %s", err, output)
	}
	launcher := filepath.Join(e.project, ".agents", "bin", nativeLauncherName("cli-enforced-tool"))
	if _, err := os.Lstat(launcher + scriptworker.ShimSidecarSuffix); err != nil {
		t.Fatalf("the CLI install published no native launcher sidecar: %v", err)
	}
	launched := exec.Command(launcher, "hello")
	launched.Env = cliEnv()
	launched.Dir = t.TempDir()
	launched.Stdin = strings.NewReader("")
	output, err := launched.CombinedOutput()
	if err != nil {
		t.Fatalf("the installed launcher refused: %v: %s", err, output)
	}
	var report struct {
		Argv []string `json:"argv"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(output), &report); err != nil {
		t.Fatalf("the installed launcher did not run the script: %v: %s", err, output)
	}
	if len(report.Argv) == 0 || report.Argv[len(report.Argv)-1] != "hello" {
		t.Fatalf("stub argv = %q, want the forwarded argument", report.Argv)
	}
}

// TestEnforcedToDeclaredOnlyFlipRemovesNativeLauncher proves launcher
// ownership across an enforcement flip: when a command stops being
// enforced, the native launcher and its sidecar are removed and the
// ordinary shim returns.
func TestEnforcedToDeclaredOnlyFlipRemovesNativeLauncher(t *testing.T) {
	curator, _ := enforcedBinaries(t)
	t.Cleanup(runtimestore.SetManagerBinaryForTest(curator))

	e := newEnv(t)
	e.schema8EnforcedSkillWithCaps("flip-skill", map[string]any{})
	e.declare("flip-skill")
	if result := e.install(Options{}); result.Status != "ok" {
		t.Fatalf("enforced install failed: %+v", result)
	}
	binDir := filepath.Join(e.project, ".agents", "bin")
	launcher := filepath.Join(binDir, "flip-skill-tool")
	if runtime.GOOS == "windows" {
		launcher += ".exe"
	}
	if _, err := os.Lstat(launcher); err != nil {
		t.Fatalf("native launcher missing after enforced install: %v", err)
	}
	if _, err := os.Lstat(launcher + scriptworker.ShimSidecarSuffix); err != nil {
		t.Fatalf("sidecar missing after enforced install: %v", err)
	}

	// Flip the command to declared-only and reinstall the moved tag.
	skillDir := filepath.Join(e.skillsRoot, "flip-skill")
	payload, _ := json.MarshalIndent(map[string]any{
		"schema_version": 8,
		"capabilities":   map[string]any{},
		"runtime_roots":  []string{"scripts"},
		"commands": map[string]any{"flip-skill-tool": map[string]any{
			"type": "script", "unix_path": "scripts/flip-skill-tool", "win_path": "scripts/flip-skill-tool",
		}},
	}, "", "  ")
	e.write(skillDir, "agent-skill.json", string(payload))
	e.git(skillDir, "add", ".")
	e.git(skillDir, "commit", "-qm", "declared-only")
	e.git(skillDir, "tag", "-f", "v1")
	if result := e.install(Options{}); result.Status != "ok" {
		t.Fatalf("declared-only reinstall failed: %+v", result)
	}
	// On unix the ordinary shim takes the launcher's path, so the flip
	// replaces native bytes with the shell shim; on Windows the `.exe`
	// launcher is gone and the `.cmd` shim returns. Either way the
	// sidecar is removed.
	if runtime.GOOS == "windows" {
		if _, err := os.Lstat(launcher); err == nil {
			t.Fatal("the native launcher survived the flip to declared-only")
		}
	} else {
		payload, err := os.ReadFile(launcher) // #nosec G304 -- staged launcher under test
		if err != nil || len(payload) < 8 {
			t.Fatalf("cannot read the flipped launcher: %v", err)
		}
		if godriver.NativeExecutableHeader(payload[:8]) {
			t.Fatal("the launcher is still a native image after the flip to declared-only")
		}
	}
	if _, err := os.Lstat(launcher + scriptworker.ShimSidecarSuffix); err == nil {
		t.Fatal("the sidecar survived the flip to declared-only")
	}
	if _, err := os.Lstat(filepath.Join(binDir, shimName("flip-skill-tool"))); err != nil {
		t.Fatalf("ordinary shim missing after the flip: %v", err)
	}
}
