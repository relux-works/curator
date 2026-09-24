package scriptworker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/scriptpolicy"
)

// TestLaunchRefusesAtPreflight proves the production entry refuses before
// anything starts when this host cannot provide a mandatory control,
// naming it. The fault injector fails the termination probe through the
// production probe path; the refusal names the control and starts no
// worker.
func TestLaunchRefusesAtPreflight(t *testing.T) {
	// Sequential by construction: the test forces the process-global probe
	// fault. Never add t.Parallel here.
	restore := OverrideScriptProbeFaultForTest(func(control string) error {
		if control == ScriptControlDescendantDomainTermination {
			return errors.New("injected probe failure")
		}
		return nil
	})
	defer restore()
	fixture := newLaunchFixture(t, os.Args[0])
	result, err := Launch(context.Background(), fixture.request)
	if DiagnosticCode(err) != scriptpolicy.ControlUnavailable {
		t.Fatalf("Launch error = %v, want %s", err, scriptpolicy.ControlUnavailable)
	}
	if len(result.Stdout) != 0 || len(result.Stderr) != 0 || result.ExitCode != 0 {
		t.Fatalf("Launch result = %+v, want zero", result)
	}
	var refusal *scriptpolicy.Error
	if !errors.As(err, &refusal) {
		t.Fatalf("Launch error %T does not carry a *scriptpolicy.Error: %v", err, err)
	}
	if !strings.Contains(refusal.Detail, ScriptControlDescendantDomainTermination) {
		t.Fatalf("refusal detail %q does not name %q", refusal.Detail, ScriptControlDescendantDomainTermination)
	}
}

// launchFixture is one manager-side session input over the stub interpreter.
type launchFixture struct {
	request LaunchRequest
	stub    string
	entry   string
	root    string
}

func newLaunchFixture(t *testing.T, managerPath string) *launchFixture {
	t.Helper()
	root := mustPhysical(t, t.TempDir())
	entry := filepath.Join(root, "store", "skill", "commit", "scripts", "tool")
	writeTestFile(t, entry, []byte("print('tool')\n"), 0o644)
	privateBase := filepath.Join(root, "private")
	if err := os.MkdirAll(privateBase, 0o755); err != nil {
		t.Fatal(err)
	}
	execDir := filepath.Join(root, "exec")
	if err := os.MkdirAll(execDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stub := mustPhysical(t, stubInterpreterBinary(t))
	return &launchFixture{
		request: LaunchRequest{
			ManagerPath:     managerPath,
			Interpreters:    map[string]string{"python3-v1": stub, "node-v1": stub},
			ForbiddenRoots:  []string{filepath.Join(root, "repo")},
			InterpreterID:   "python3-v1",
			RuntimeEntry:    entry,
			CapabilitiesRaw: json.RawMessage(`{}`),
			HostEnvironment: []string{},
			ExecSearchDirs:  []string{execDir},
			PrivateBase:     privateBase,
		},
		stub:  stub,
		entry: entry,
		root:  root,
	}
}

// declareCapabilities replaces the fixture's declared capabilities object.
func (fixture *launchFixture) declareCapabilities(t *testing.T, capabilities string) {
	t.Helper()
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(capabilities), &raw); err != nil {
		t.Fatal(err)
	}
	fixture.request.CapabilitiesRaw = raw
}

// TestRunSessionHappyPath proves the manager half of the fixed session at
// the real process boundary: launch-boundary recheck, fresh nonce,
// identity proof, verbatim arguments, exit status, the derived environment
// without inheritance, and private-area cleanup.
func TestRunSessionHappyPath(t *testing.T) {
	fixture := newLaunchFixture(t, os.Args[0])
	fixture.request.Args = []string{"--verbose", "two words"}
	fixture.declareCapabilities(t, `{"env_read": ["STUB_EXIT"]}`)
	fixture.request.HostEnvironment = []string{"STUB_EXIT=0", "STUB_INHERIT_PROBE=leaked"}
	fixture.request.Stdin = []byte("hello\n")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	result, err := runSession(ctx, fixture.request, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr %q)", result.ExitCode, result.Stderr)
	}
	var report stubReport
	if err := json.Unmarshal(result.Stdout, &report); err != nil {
		t.Fatalf("cannot decode the stub report %q: %v", result.Stdout, err)
	}
	wantArgv := []string{fixture.stub, fixture.entry, "--verbose", "two words"}
	if strings.Join(report.Argv, "\x00") != strings.Join(wantArgv, "\x00") {
		t.Fatalf("argv = %q, want %q", report.Argv, wantArgv)
	}
	if report.Executable != fixture.stub {
		t.Fatalf("executable = %q, want the resolved interpreter %q", report.Executable, fixture.stub)
	}
	if report.Stdin != "hello\n" {
		t.Fatalf("stdin = %q, want the explicit payload", report.Stdin)
	}
	if report.Probe != "" {
		t.Fatalf("probe = %q, want no inherited environment", report.Probe)
	}
	if report.Env["STUB_EXIT"] != "0" {
		t.Fatalf("STUB_EXIT = %q, want the declared passthrough value", report.Env["STUB_EXIT"])
	}
	if _, present := report.Env["STUB_INHERIT_PROBE"]; present {
		t.Fatal("the undeclared host variable reached the interpreter")
	}
	// The operation-private area is removed before the invocation returns.
	entries, err := os.ReadDir(fixture.request.PrivateBase)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range entries {
		if strings.HasPrefix(item.Name(), ".curator-script-") {
			t.Fatalf("operation-private area %q outlived the invocation", item.Name())
		}
	}
}

// TestRunSessionRechecksIdentityAtLaunchBoundary proves a manager binary
// replaced between the first check and exec never runs: the recheck refuses
// with worker-identity-invalid, and the replacement's bytes are not what the
// session speaks to. Dropping the recheck turns this refusal into a
// protocol failure against the replacement's output.
func TestRunSessionRechecksIdentityAtLaunchBoundary(t *testing.T) {
	managerCopy := filepath.Join(mustPhysical(t, t.TempDir()), executableFixtureName("curator-manager"))
	copyTestFile(t, mustPhysical(t, os.Args[0]), managerCopy)
	fixture := newLaunchFixture(t, managerCopy)

	replacement := stubInterpreterBinary(t)
	hooks := &launchHooks{afterResolve: func() {
		payload, err := os.ReadFile(replacement) // #nosec G304 -- test-owned stub path
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(managerCopy, payload, 0o755); err != nil {
			panic(err)
		}
	}}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, err := runSession(ctx, fixture.request, hooks)
	if DiagnosticCode(err) != CodeWorkerIdentityInvalid {
		t.Fatalf("replaced-manager error = %v, want %s", err, CodeWorkerIdentityInvalid)
	}
}

// TestClientRejectsUnknownNonce proves the parent binds every worker message
// to the fresh session nonce. Dropping the comparison accepts a replayed
// recording as the live worker's proof.
func TestClientRejectsUnknownNonce(t *testing.T) {
	nonce := strings.Repeat("ab", 32)
	frame := func(kind, frameNonce string) []byte {
		var buffer bytes.Buffer
		if err := writeMessage(&buffer, workerMessage{Kind: kind, Nonce: frameNonce, Ready: &workerReady{}}); err != nil {
			t.Fatal(err)
		}
		return buffer.Bytes()
	}
	client := &workerClient{nonce: nonce, stdout: io.NopCloser(bytes.NewReader(frame(kindReady, strings.Repeat("00", 32))))}
	if _, err := client.receive(kindReady); DiagnosticCode(err) != CodeWorkerProtocolInvalid {
		t.Fatalf("foreign-nonce error = %v, want %s", err, CodeWorkerProtocolInvalid)
	}
	// The control: the bound nonce is accepted.
	bound := &workerClient{nonce: nonce, stdout: io.NopCloser(bytes.NewReader(frame(kindReady, nonce)))}
	message, err := bound.receive(kindReady)
	if err != nil {
		t.Fatalf("bound nonce was refused: %v", err)
	}
	if message.Nonce != nonce {
		t.Fatalf("nonce = %q, want the session nonce", message.Nonce)
	}
	// A worker failure still surfaces the worker's own stable code.
	var failures bytes.Buffer
	if err := writeMessage(&failures, workerMessage{
		Kind: kindFailure, Nonce: nonce,
		Failure: &workerFailure{Code: CodeWorkerIdentityInvalid, Detail: "proof mismatch"},
	}); err != nil {
		t.Fatal(err)
	}
	failing := &workerClient{nonce: nonce, stdout: io.NopCloser(bytes.NewReader(failures.Bytes()))}
	if _, err := failing.receive(kindReady); DiagnosticCode(err) != CodeWorkerIdentityInvalid {
		t.Fatalf("failure error = %v, want %s", err, CodeWorkerIdentityInvalid)
	}
}

// TestRunSessionTerminatesDescendants proves the worker-domain teardown
// reaps an interpreter descendant that outlives the interpreter itself.
// Skipping the teardown leaves the descendant alive past the return.
// The descendant publishes its identity through the stdout protocol
// report, never through a side file (the write confinement grants the
// derived path set plus the null device, nothing else); the spawned
// program is the granted interpreter itself, so the exec denial allows
// the spawn wherever it enforces.
func TestRunSessionTerminatesDescendants(t *testing.T) {
	fixture := newLaunchFixture(t, os.Args[0])
	fixture.request.ProjectRoot = fixture.root
	fixture.declareCapabilities(t, `{"filesystem": "repo", "env_read": ["STUB_SPAWN_SLEEP", "STUB_SPAWN_SECONDS"]}`)
	fixture.request.HostEnvironment = []string{
		"STUB_SPAWN_SLEEP=1", "STUB_SPAWN_SECONDS=120"}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	result, err := runSession(ctx, fixture.request, nil)
	if err != nil {
		t.Fatal(err)
	}
	report := decodeStubReport(t, result.Stdout)
	if report.SpawnErr != "" {
		t.Fatalf("the interpreter descendant did not start: %s", report.SpawnErr)
	}
	if report.SpawnPid <= 0 {
		t.Fatal("the interpreter descendant published no identity")
	}
	pid := report.SpawnPid
	deadline := time.Now().Add(15 * time.Second)
	for processAlive(pid) && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if processAlive(pid) {
		t.Fatalf("descendant %d outlived the worker domain teardown", pid)
	}
}

// TestRunSessionRejectsPATHInterpreter proves the launch never consults the
// user PATH: without a binding the session refuses even when PATH names a
// working interpreter, and with a binding the configured file is what runs.
// Falling back to a PATH lookup turns the first refusal into a launch.
func TestRunSessionRejectsPATHInterpreter(t *testing.T) {
	decoyDir := mustPhysical(t, t.TempDir())
	decoy := filepath.Join(decoyDir, "python3")
	if runtime.GOOS == "windows" {
		decoy += ".exe"
	}
	copyTestFile(t, stubInterpreterBinary(t), decoy)
	t.Setenv("PATH", decoyDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	fixture := newLaunchFixture(t, os.Args[0])
	fixture.request.Interpreters = nil
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if _, err := runSession(ctx, fixture.request, nil); DiagnosticCode(err) != scriptpolicy.ControlUnavailable {
		t.Fatalf("unconfigured launch with PATH decoy = %v, want %s", err, scriptpolicy.ControlUnavailable)
	}

	bound := newLaunchFixture(t, os.Args[0])
	result, err := runSession(ctx, bound.request, nil)
	if err != nil {
		t.Fatal(err)
	}
	var report stubReport
	if err := json.Unmarshal(result.Stdout, &report); err != nil {
		t.Fatal(err)
	}
	if report.Executable != bound.stub {
		t.Fatalf("executable = %q, want the configured file %q, not the PATH decoy", report.Executable, bound.stub)
	}
}

// TestBuiltCuratorWorkerHandshake proves the production binary wires the
// fixed hidden mode: the built curator answers the manager's session
// protocol with an identity proof and runs one invocation, and the mode is
// selected by exactly that one argument.
func TestBuiltCuratorWorkerHandshake(t *testing.T) {
	curator := mustPhysical(t, builtCuratorBinary(t))
	identity, err := godriver.ResolveManagerIdentity(curator)
	if err != nil {
		t.Fatal(err)
	}
	worker := startRawWorker(t, curator)
	fixture := newWorkerFixture(t, curator)
	if fixture.manager.SHA256 != identity.SHA256 {
		t.Fatalf("fixture identity does not match the built binary")
	}
	wire := fixture.request()
	wire.Args = []string{"production"}
	worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
	ready := worker.receive()
	if ready.Kind != kindReady || ready.Ready == nil {
		t.Fatalf("built curator sent %q, want ready", ready.Kind)
	}
	if ready.Ready.ExecutableSHA256 != identity.SHA256 {
		t.Fatalf("built curator proved %s, want %s", ready.Ready.ExecutableSHA256, identity.SHA256)
	}
	worker.permit(fixture.nonce)
	result := worker.receive()
	if result.Kind != kindResult || result.Result == nil || result.Result.ExitCode != 0 {
		t.Fatalf("built curator result = %+v, want one clean run", result.Result)
	}
	report := decodeStubReport(t, result.Result.Stdout)
	if len(report.Argv) != 3 || report.Argv[2] != "production" {
		t.Fatalf("argv = %q, want the verbatim argument", report.Argv)
	}
	worker.send(workerMessage{Kind: kindShutdown, Nonce: fixture.nonce})

	// The mode takes exactly one argument: anything else never reaches the
	// worker.
	extra := exec.Command(curator, WorkerMode, "extra")
	extra.Env = os.Environ()
	output, err := extra.CombinedOutput()
	if err == nil {
		t.Fatalf("extra-argument worker invocation exited 0: %s", output)
	}
	if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() == 0 {
		t.Fatalf("extra-argument worker invocation error = %v", err)
	}
}
