package scriptworker

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/godriver"
)

// rawWorker drives a real hidden-mode script worker process over the session
// channel with complete control over the framing, so every identity and
// protocol rejection is proved against the shipped worker rather than a mock.
type rawWorker struct {
	t       *testing.T
	command *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
}

func startRawWorker(t *testing.T, launchPath string) *rawWorker {
	t.Helper()
	command := exec.Command(launchPath, WorkerMode) // #nosec G204 -- identity-verified re-execution of the test binary
	command.Env = scriptWorkerEnvironment()
	command.SysProcAttr = scriptWorkerSysProcAttr()
	stdin, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	command.Stderr = os.Stderr
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	worker := &rawWorker{t: t, command: command, stdin: stdin, stdout: stdout}
	t.Cleanup(worker.close)
	return worker
}

func (worker *rawWorker) send(message workerMessage) {
	worker.t.Helper()
	if err := writeMessage(worker.stdin, message); err != nil {
		worker.t.Fatalf("cannot send %s: %v", message.Kind, err)
	}
}

func (worker *rawWorker) receive() workerMessage {
	worker.t.Helper()
	message, err := readMessage(worker.stdout)
	if err != nil {
		worker.t.Fatalf("cannot read a worker message: %v", err)
	}
	return message
}

// permit sends the parent's permit frame: the worker starts the
// interpreter only after it arrives.
func (worker *rawWorker) permit(nonce string) {
	worker.t.Helper()
	worker.send(workerMessage{Kind: kindPermit, Nonce: nonce})
}

// expectFailure requires the worker to reject with the given stable
// diagnostic before the interpreter runs.
func (worker *rawWorker) expectFailure(code string) {
	worker.t.Helper()
	message := worker.receive()
	if message.Kind != kindFailure || message.Failure == nil {
		worker.t.Fatalf("worker sent %q, want a failure", message.Kind)
	}
	if message.Failure.Code != code {
		worker.t.Fatalf("worker failure code = %q (%s), want %q", message.Failure.Code, message.Failure.Detail, code)
	}
}

// expectFailureDetail requires the worker to reject with the given stable
// code and a detail naming the expected cause, before the interpreter runs.
// The detail pins which gate refused, so a row cannot pass on the wrong
// refusal (for example the name or image gate firing before the tamper the
// row wants to prove).
func (worker *rawWorker) expectFailureDetail(code, wantDetail string) {
	worker.t.Helper()
	message := worker.receive()
	if message.Kind != kindFailure || message.Failure == nil {
		worker.t.Fatalf("worker sent %q, want a failure", message.Kind)
	}
	if message.Failure.Code != code {
		worker.t.Fatalf("worker failure code = %q (%s), want %q", message.Failure.Code, message.Failure.Detail, code)
	}
	if !strings.Contains(message.Failure.Detail, wantDetail) {
		worker.t.Fatalf("worker failure detail = %q, want it to name %q", message.Failure.Detail, wantDetail)
	}
}

func (worker *rawWorker) close() {
	_ = worker.stdin.Close()
	terminateScriptDomain(worker.command, nil)
	_ = worker.command.Wait()
}

// workerFixture is one valid session input: a stub interpreter, a runtime
// entry, and manager-owned directories.
type workerFixture struct {
	manager     godriver.ExecutableIdentity
	interpreter InterpreterIdentity
	entry       string
	workDir     string
	base        string
	tmp         string
	config      string
	cache       string
	farm        string
	environment []string
	nonce       string
}

func newWorkerFixture(t *testing.T, launchPath string) *workerFixture {
	t.Helper()
	root := mustPhysical(t, t.TempDir())
	entry := filepath.Join(root, "runtime", "tool")
	writeTestFile(t, entry, []byte("print('tool')\n"), 0o644)
	workDir := filepath.Join(root, "work")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}
	private := filepath.Join(root, "private")
	for _, leaf := range []string{"tmp", "config", "cache"} {
		if err := os.MkdirAll(filepath.Join(private, leaf), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	manager, err := godriver.ResolveManagerIdentity(launchPath)
	if err != nil {
		t.Fatal(err)
	}
	stub := mustPhysical(t, stubInterpreterBinary(t))
	interpreter, err := ResolveInterpreter("python3-v1", map[string]string{"python3-v1": stub}, nil)
	if err != nil {
		t.Fatal(err)
	}
	farm := filepath.Join(private, "path")
	if err := os.MkdirAll(farm, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := linkFarmEntry(farm, filepath.Base(interpreter.Path), interpreter.Path); err != nil {
		t.Fatal(err)
	}
	// The environment is production-shaped: the same builder the manager
	// runs, over an all-absent declaration and an empty host environment,
	// so the worker's revalidation accepts the untouched request.
	declared, err := ParseDeclaredCapabilities(nil)
	if err != nil {
		t.Fatal(err)
	}
	environment, _, err := buildSessionEnvironment(environmentRequest{
		declared:      declared,
		interpreterID: interpreter.ID,
		hostEnv:       []string{},
		platform:      runtime.GOOS,
		interpreter:   interpreter,
		private:       privateArea{base: private, tmp: filepath.Join(private, "tmp"), config: filepath.Join(private, "config"), cache: filepath.Join(private, "cache")},
		farmDir:       farm,
		offline:       true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return &workerFixture{
		manager: manager, interpreter: interpreter, entry: entry,
		workDir: workDir, base: private,
		tmp:         filepath.Join(private, "tmp"),
		config:      filepath.Join(private, "config"),
		cache:       filepath.Join(private, "cache"),
		farm:        farm,
		environment: environment,
		nonce:       strings.Repeat("ab", 32),
	}
}

func (fixture *workerFixture) request() *workerRequest {
	platform := ScriptInventoryPlatform(runtime.GOOS)
	inventory := make([]ScriptInventoryInput, 0, len(scriptInventoryOrder))
	for _, name := range scriptInventoryOrder {
		availability := scriptInventoryPlatforms[platform][name].Availability
		inventory = append(inventory, ScriptInventoryInput{
			Name: name, Availability: availability, Present: fixtureInventoryPresent(name),
		})
	}
	var bound uint64
	if fixtureInventoryPresent(ScriptControlPerFileSizeLimit) {
		if current, ok := currentScriptFileSizeLimit(); ok {
			bound = current
		}
	}
	return &workerRequest{
		Version:           protocolVersion,
		ExecutablePath:    fixture.manager.Path,
		ExecutableSHA256:  fixture.manager.SHA256,
		ExecutableSize:    fixture.manager.Size,
		InterpreterID:     fixture.interpreter.ID,
		InterpreterPath:   fixture.interpreter.Path,
		InterpreterSHA:    fixture.interpreter.SHA256,
		InterpreterSize:   fixture.interpreter.Size,
		RuntimeEntry:      fixture.entry,
		Environment:       append([]string(nil), fixture.environment...),
		WorkingDir:        fixture.workDir,
		PrivateBase:       fixture.base,
		PrivateTmp:        fixture.tmp,
		PrivateConfig:     fixture.config,
		PrivateCache:      fixture.cache,
		FarmDir:           fixture.farm,
		NetworkOffline:    true,
		StdinIsNull:       true,
		OutputLimit:       defaultOutputLimit,
		InventoryPlatform: platform,
		Inventory:         inventory,
		FileSizeBound:     bound,
	}
}

// fixtureInventoryPresent marks the controls a raw worker fixture claims
// installable: exactly the controls a directly started worker process can
// confirm. The raw worker leads its own session on unix and holds a
// close-on-exec session channel everywhere, and the per-file bound is
// parameterized with the bound it observes; job-backed and Linux-only
// controls stay unclaimed because the fixture starts no domain for them.
// Production sessions claim the full probed set through runSession instead.
func fixtureInventoryPresent(name string) bool {
	switch name {
	case ScriptControlDescendantDomainTermination:
		return runtime.GOOS != "windows"
	case ScriptControlPerFileSizeLimit:
		current, ok := currentScriptFileSizeLimit()
		return ok && current != 0
	case ScriptControlInheritedHandleRestriction:
		return true
	default:
		return false
	}
}

// stubReport decodes what the stub interpreter recorded about its launch.
type stubReport struct {
	Executable string            `json:"executable"`
	Argv       []string          `json:"argv"`
	Cwd        string            `json:"cwd"`
	Stdin      string            `json:"stdin"`
	Probe      string            `json:"probe"`
	Env        map[string]string `json:"env"`
	LookPath   map[string]string `json:"lookpath"`
	Write      map[string]string `json:"write"`
	Truncate   map[string]string `json:"truncate"`
	OTrunc     map[string]string `json:"otrunc"`
	Unlink     map[string]string `json:"unlink"`
	Rmdir      map[string]string `json:"rmdir"`
	Mkdir      map[string]string `json:"mkdir"`
	Mkfifo     map[string]string `json:"mkfifo"`
	Rename     map[string]string `json:"rename"`
	ExecTry    string            `json:"exec_try"`
	SpawnPid   int               `json:"spawn_pid"`
	SpawnErr   string            `json:"spawn_err"`
}

func decodeStubReport(t *testing.T, stdout []byte) stubReport {
	t.Helper()
	var report stubReport
	if err := json.Unmarshal(stdout, &report); err != nil {
		t.Fatalf("cannot decode the stub report %q: %v", stdout, err)
	}
	return report
}

// TestScriptWorkerHappyPath proves the fixed worker session at the real
// process boundary: identity proof, interpreter verification, runtime-entry
// execution with verbatim arguments, and the child's exit status.
func TestScriptWorkerHappyPath(t *testing.T) {
	worker := startRawWorker(t, os.Args[0])
	fixture := newWorkerFixture(t, os.Args[0])
	wire := fixture.request()
	wire.Args = []string{"--count", "3", "hello world", "ünicode", "-x"}
	wire.Environment = append(wire.Environment, "STUB_EXIT=0")

	worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
	ready := worker.receive()
	if ready.Kind != kindReady || ready.Ready == nil {
		t.Fatalf("worker sent %q, want ready", ready.Kind)
	}
	if ready.Nonce != fixture.nonce {
		t.Fatal("worker acknowledgement carries a foreign nonce")
	}
	if ready.Ready.ExecutableSHA256 != fixture.manager.SHA256 || ready.Ready.InterpreterSHA != fixture.interpreter.SHA256 {
		t.Fatalf("worker proof = %+v, want the session identities", ready.Ready)
	}
	worker.permit(fixture.nonce)
	result := worker.receive()
	if result.Kind != kindResult || result.Result == nil {
		t.Fatalf("worker sent %q, want result", result.Kind)
	}
	if result.Nonce != fixture.nonce {
		t.Fatal("worker result carries a foreign nonce")
	}
	if result.Result.ExitCode != 0 || result.Result.Started != 1 || result.Result.Overflow {
		t.Fatalf("result = %+v, want one clean run", result.Result)
	}
	report := decodeStubReport(t, result.Result.Stdout)
	wantArgv := append([]string{fixture.interpreter.Path, fixture.entry}, wire.Args...)
	if strings.Join(report.Argv, "\x00") != strings.Join(wantArgv, "\x00") {
		t.Fatalf("argv = %q, want %q", report.Argv, wantArgv)
	}
	if report.Cwd != fixture.workDir {
		t.Fatalf("cwd = %q, want %q", report.Cwd, fixture.workDir)
	}
	worker.send(workerMessage{Kind: kindShutdown, Nonce: fixture.nonce})
}

// TestScriptWorkerRechecksInterpreterBeforeExec narrows the interpreter
// identity gate at the worker boundary: replace the verified image after
// ready but before the manager permit, then require the launch-boundary
// proof to refuse before the interpreter can run.
func TestScriptWorkerRechecksInterpreterBeforeExec(t *testing.T) {
	worker := startRawWorker(t, os.Args[0])
	fixture := newWorkerFixture(t, os.Args[0])
	sharedInterpreter := fixture.interpreter
	isolatedPath := filepath.Join(filepath.Dir(fixture.entry), "worker-interpreter"+exeSuffix())
	copyTestFile(t, sharedInterpreter.Path, isolatedPath)
	interpreter, err := ResolveInterpreter("python3-v1", map[string]string{"python3-v1": isolatedPath}, nil)
	if err != nil {
		t.Fatal(err)
	}
	fixture.interpreter = interpreter
	wire := fixture.request()
	marker := filepath.Join(filepath.Dir(fixture.entry), "interpreter-ran")
	wire.Environment = append(wire.Environment, "STUB_MARKER="+marker)

	worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
	ready := worker.receive()
	if ready.Kind != kindReady || ready.Ready == nil {
		t.Fatalf("worker sent %q, want ready before the interpreter permit", ready.Kind)
	}

	image, err := os.ReadFile(wire.InterpreterPath) // #nosec G304 -- fixture-owned interpreter image
	if err != nil {
		t.Fatal(err)
	}
	if len(image) == 0 {
		t.Fatal("the fixture interpreter image is empty")
	}
	image[len(image)-1] ^= 0xff // keep the native image header; change the proved bytes
	if err := os.WriteFile(wire.InterpreterPath, image, 0o700); err != nil {
		t.Fatalf("cannot replace the accepted interpreter image: %v", err)
	}
	if err := VerifyInterpreter(sharedInterpreter); err != nil {
		t.Fatalf("the replacement test changed the package-shared stub image: %v", err)
	}

	worker.permit(fixture.nonce)
	worker.expectFailureDetail(CodeWorkerIdentityInvalid, "interpreter identity proof")
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the replaced interpreter ran after worker acceptance")
	} else if !os.IsNotExist(err) {
		t.Fatalf("cannot check the interpreter marker: %v", err)
	}
}

// TestScriptWorkerReturnsTheChildExitStatus proves the interpreter's exit
// status and stderr return without reinterpretation.
func TestScriptWorkerReturnsTheChildExitStatus(t *testing.T) {
	worker := startRawWorker(t, os.Args[0])
	fixture := newWorkerFixture(t, os.Args[0])
	wire := fixture.request()
	wire.Environment = append(wire.Environment, "STUB_EXIT=7", "STUB_STDERR=boom\n")

	worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
	if ready := worker.receive(); ready.Kind != kindReady {
		t.Fatalf("worker sent %q, want ready", ready.Kind)
	}
	worker.permit(fixture.nonce)
	result := worker.receive()
	if result.Kind != kindResult || result.Result == nil {
		t.Fatalf("worker sent %q, want result", result.Kind)
	}
	if result.Result.ExitCode != 7 {
		t.Fatalf("exit code = %d, want 7", result.Result.ExitCode)
	}
	if string(result.Result.Stderr) != "boom\n" {
		t.Fatalf("stderr = %q, want %q", result.Result.Stderr, "boom\n")
	}
	worker.send(workerMessage{Kind: kindShutdown, Nonce: fixture.nonce})
}

// TestScriptWorkerBindsStandardInput proves explicit stream binding: a
// request payload reaches the interpreter's standard input, and a null
// binding reads EOF rather than an inherited descriptor.
func TestScriptWorkerBindsStandardInput(t *testing.T) {
	worker := startRawWorker(t, os.Args[0])
	fixture := newWorkerFixture(t, os.Args[0])
	wire := fixture.request()
	wire.Stdin = []byte("payload-bytes\n")
	wire.StdinIsNull = false

	worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
	if ready := worker.receive(); ready.Kind != kindReady {
		t.Fatalf("worker sent %q, want ready", ready.Kind)
	}
	worker.permit(fixture.nonce)
	result := worker.receive()
	if result.Kind != kindResult || result.Result == nil {
		t.Fatalf("worker sent %q, want result", result.Kind)
	}
	if report := decodeStubReport(t, result.Result.Stdout); report.Stdin != "payload-bytes\n" {
		t.Fatalf("stdin = %q, want the request payload", report.Stdin)
	}
	worker.send(workerMessage{Kind: kindShutdown, Nonce: fixture.nonce})

	// The null arm: the stub reads EOF, which an inherited descriptor would
	// only produce by accident of the test host.
	second := startRawWorker(t, os.Args[0])
	other := newWorkerFixture(t, os.Args[0])
	nullWire := other.request()
	second.send(workerMessage{Kind: kindRequest, Nonce: other.nonce, Request: nullWire})
	if ready := second.receive(); ready.Kind != kindReady {
		t.Fatalf("worker sent %q, want ready", ready.Kind)
	}
	second.permit(other.nonce)
	nullResult := second.receive()
	if nullResult.Kind != kindResult || nullResult.Result == nil {
		t.Fatalf("worker sent %q, want result", nullResult.Kind)
	}
	if report := decodeStubReport(t, nullResult.Result.Stdout); report.Stdin != "" {
		t.Fatalf("null stdin read %q, want EOF", report.Stdin)
	}
	second.send(workerMessage{Kind: kindShutdown, Nonce: other.nonce})
}

// TestScriptWorkerIgnoresTheShebang proves the executed program is the
// resolved interpreter, never anything the script bytes name.
func TestScriptWorkerIgnoresTheShebang(t *testing.T) {
	worker := startRawWorker(t, os.Args[0])
	fixture := newWorkerFixture(t, os.Args[0])
	marker := filepath.Join(fixture.workDir, "ran")
	writeTestFile(t, fixture.entry, []byte("#!/bin/sh\nexit 99\n"), 0o755)
	wire := fixture.request()
	wire.Environment = append(wire.Environment, "STUB_MARKER="+marker)

	worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
	if ready := worker.receive(); ready.Kind != kindReady {
		t.Fatalf("worker sent %q, want ready", ready.Kind)
	}
	worker.permit(fixture.nonce)
	result := worker.receive()
	if result.Kind != kindResult || result.Result == nil {
		t.Fatalf("worker sent %q, want result", result.Kind)
	}
	if result.Result.ExitCode != 0 {
		t.Fatalf("exit code = %d, want the interpreter's 0, not the shebang's 99", result.Result.ExitCode)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("the configured interpreter never ran")
	}
	worker.send(workerMessage{Kind: kindShutdown, Nonce: fixture.nonce})
}

// TestScriptWorkerRejectsMalformedSessions proves framing, version, and
// nonce failures refuse before any identity proof and before the
// interpreter runs.
func TestScriptWorkerRejectsMalformedSessions(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "ran")
	build := func(t *testing.T, mutate func(*workerRequest, *string)) (*rawWorker, workerMessage) {
		worker := startRawWorker(t, os.Args[0])
		fixture := newWorkerFixture(t, os.Args[0])
		wire := fixture.request()
		wire.Environment = append(wire.Environment, "STUB_MARKER="+marker)
		nonce := fixture.nonce
		mutate(wire, &nonce)
		return worker, workerMessage{Kind: kindRequest, Nonce: nonce, Request: wire}
	}
	for _, testCase := range []struct {
		name   string
		mutate func(*workerRequest, *string)
	}{
		{"short-nonce", func(_ *workerRequest, nonce *string) { *nonce = "ab" }},
		{"non-hex-nonce", func(_ *workerRequest, nonce *string) { *nonce = strings.Repeat("zz", 32) }},
		{"empty-nonce", func(_ *workerRequest, nonce *string) { *nonce = "" }},
		{"wrong-version", func(wire *workerRequest, _ *string) { wire.Version = "curator-go-worker-v1" }},
		{"empty-version", func(wire *workerRequest, _ *string) { wire.Version = "" }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			worker, message := build(t, testCase.mutate)
			worker.send(message)
			worker.expectFailure(CodeWorkerProtocolInvalid)
		})
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the interpreter ran for a malformed session")
	}
}

// TestScriptWorkerRejectsForgedWorkerIdentity proves a copied or modified
// manager binary cannot pass as the expected worker: the proof binds path,
// hash, and size together.
func TestScriptWorkerRejectsForgedWorkerIdentity(t *testing.T) {
	// A copied binary launched in worker mode proves its own bytes, which
	// do not match the original's expectation.
	original := mustPhysical(t, os.Args[0])
	copied := filepath.Join(mustPhysical(t, t.TempDir()), executableFixtureName("curator-copy"))
	copyTestFile(t, original, copied)
	worker := startRawWorker(t, copied)
	fixture := newWorkerFixture(t, original)
	marker := filepath.Join(fixture.workDir, "ran")
	wire := fixture.request()
	wire.Environment = append(wire.Environment, "STUB_MARKER="+marker)

	worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
	worker.expectFailure(CodeWorkerIdentityInvalid)
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the interpreter ran for a forged worker identity")
	}

	// A modified binary fails the same way: the bytes no longer hash to
	// the recorded identity.
	modified := filepath.Join(mustPhysical(t, t.TempDir()), executableFixtureName("curator-modified"))
	copyTestFile(t, original, modified)
	payload, err := os.ReadFile(modified) // #nosec G304 -- test-owned copy
	if err != nil {
		t.Fatal(err)
	}
	payload[len(payload)-1] ^= 0xff
	if err := os.WriteFile(modified, payload, 0o755); err != nil {
		t.Fatal(err)
	}
	second := startRawWorker(t, modified)
	other := newWorkerFixture(t, original)
	otherWire := other.request()
	otherWire.Environment = append(otherWire.Environment, "STUB_MARKER="+marker)
	second.send(workerMessage{Kind: kindRequest, Nonce: other.nonce, Request: otherWire})
	second.expectFailure(CodeWorkerIdentityInvalid)
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the interpreter ran for a modified worker binary")
	}
}

// TestScriptWorkerRejectsSubstitutedManager proves a worker started through
// a launcher link still proves the installed identity, and a retargeted
// link cannot move it.
func TestScriptWorkerRejectsSubstitutedManager(t *testing.T) {
	directory := mustPhysical(t, t.TempDir())
	installed := filepath.Join(directory, executableFixtureName("curator-real"))
	copyTestFile(t, mustPhysical(t, os.Args[0]), installed)
	other := filepath.Join(directory, executableFixtureName("curator-other"))
	copyTestFile(t, installed, other)
	payload, err := os.ReadFile(other) // #nosec G304 -- test-owned copy
	if err != nil {
		t.Fatal(err)
	}
	payload[len(payload)-1] ^= 0xff
	if err := os.WriteFile(other, payload, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(directory, executableFixtureName("curator"))
	if err := os.Symlink(installed, link); err != nil {
		t.Skipf("this host forbids symlink creation: %v", err)
	}

	// Through the link at the installed file: the worker proves that file.
	worker := startRawWorker(t, link)
	fixture := newWorkerFixture(t, installed)
	wire := fixture.request()
	worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
	if ready := worker.receive(); ready.Kind != kindReady {
		t.Fatalf("worker launched through a link sent %q, want ready", ready.Kind)
	}
	worker.permit(fixture.nonce)
	result := worker.receive()
	if result.Kind != kindResult {
		t.Fatalf("worker launched through a link sent %q, want result", result.Kind)
	}
	worker.send(workerMessage{Kind: kindShutdown, Nonce: fixture.nonce})

	// Retarget the link at the attacker's bytes: the recorded identity no
	// longer matches what the worker proves.
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(other, link); err != nil {
		t.Fatal(err)
	}
	substituted := startRawWorker(t, link)
	second := newWorkerFixture(t, installed)
	secondWire := second.request()
	substituted.send(workerMessage{Kind: kindRequest, Nonce: second.nonce, Request: secondWire})
	substituted.expectFailure(CodeWorkerIdentityInvalid)
}

// TestScriptWorkerRejectsTamperedInterpreter proves the interpreter file is
// verified before it runs: a hash the bytes no longer match refuses, and a
// replacement of the file between sessions refuses too.
func TestScriptWorkerRejectsTamperedInterpreter(t *testing.T) {
	worker := startRawWorker(t, os.Args[0])
	fixture := newWorkerFixture(t, os.Args[0])
	marker := filepath.Join(fixture.workDir, "ran")
	wire := fixture.request()
	wire.Environment = append(wire.Environment, "STUB_MARKER="+marker)
	wire.InterpreterSHA = "sha256:" + strings.Repeat("0", 64)

	worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
	worker.expectFailureDetail(CodeWorkerIdentityInvalid, "interpreter identity proof")
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the interpreter ran for a tampered identity")
	}

	// A file replaced after the expectation was recorded refuses the same
	// way, even though the expectation itself is well-formed. The copy
	// names the native image itself (with the platform executable suffix
	// on Windows) so the pre-tamper resolution succeeds there; the tamper
	// then flips the last byte, keeping the image header intact, so the
	// refusal names the proof mismatch rather than the image gate.
	stub := stubInterpreterBinary(t)
	replaced := filepath.Join(mustPhysical(t, t.TempDir()), executableFixtureName("interp"))
	copyTestFile(t, stub, replaced)
	replaced = mustPhysical(t, replaced)
	recorded, err := ResolveInterpreter("node-v1", map[string]string{"node-v1": replaced}, nil)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(replaced) // #nosec G304 -- test-owned copy
	if err != nil {
		t.Fatal(err)
	}
	payload[len(payload)-1] ^= 0xff
	if err := os.WriteFile(replaced, payload, 0o755); err != nil {
		t.Fatal(err)
	}
	second := startRawWorker(t, os.Args[0])
	other := newWorkerFixture(t, os.Args[0])
	otherWire := other.request()
	otherWire.InterpreterID = recorded.ID
	otherWire.InterpreterPath = recorded.Path
	otherWire.InterpreterSHA = recorded.SHA256
	otherWire.InterpreterSize = recorded.Size
	otherWire.Environment = append(otherWire.Environment, "STUB_MARKER="+marker)
	second.send(workerMessage{Kind: kindRequest, Nonce: other.nonce, Request: otherWire})
	second.expectFailureDetail(CodeWorkerIdentityInvalid, "interpreter identity proof")
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the interpreter ran after its file was replaced")
	}
}

// TestScriptWorkerRejectsWrapperImage proves the worker refuses a non-native
// interpreter binding at the real process boundary — expectFailure with the
// interpreter-never-ran marker — even when the request carries the wrapper's
// own correct hash and size: an identity proof is not a license to run a
// wrapper. The POSIX arm is a working shebang wrapper that would exec the
// stub; the Windows arms are a .cmd wrapper, an extensionless native copy,
// and an .exe name over non-image bytes.
func TestScriptWorkerRejectsWrapperImage(t *testing.T) {
	directory := mustPhysical(t, t.TempDir())
	stub := mustPhysical(t, stubInterpreterBinary(t))
	marker := filepath.Join(directory, "ran")
	type binding struct {
		name string
		path string
	}
	var bindings []binding
	if runtime.GOOS == "windows" {
		cmdWrapper := filepath.Join(directory, "wrapper.cmd")
		writeTestFile(t, cmdWrapper, []byte("@echo off\r\necho ran>> \""+marker+"\"\r\n"), 0o644)
		bindings = append(bindings, binding{"cmd-wrapper", cmdWrapper})
		extensionless := filepath.Join(directory, "interp")
		copyTestFile(t, stub, extensionless)
		bindings = append(bindings, binding{"extensionless-native-image", extensionless})
		fake := filepath.Join(directory, "fake.exe")
		writeTestFile(t, fake, []byte("@echo off\r\necho ran>> \""+marker+"\"\r\n"), 0o644)
		bindings = append(bindings, binding{"exe-named-text", fake})
	} else {
		wrapper := filepath.Join(directory, "wrapper")
		writeTestFile(t, wrapper, []byte("#!/bin/sh\necho ran >> \""+marker+"\"\nexec \""+stub+"\" \"$@\"\n"), 0o755)
		bindings = append(bindings, binding{"shebang-wrapper", wrapper})
	}
	for _, testCase := range bindings {
		t.Run(testCase.name, func(t *testing.T) {
			path := mustPhysical(t, testCase.path)
			payload, err := os.ReadFile(path) // #nosec G304 -- test-owned path
			if err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(payload)
			worker := startRawWorker(t, os.Args[0])
			fixture := newWorkerFixture(t, os.Args[0])
			wire := fixture.request()
			wire.InterpreterPath = path
			wire.InterpreterSHA = "sha256:" + hex.EncodeToString(digest[:])
			wire.InterpreterSize = int64(len(payload))
			wire.Environment = append(wire.Environment, "STUB_MARKER="+marker)
			worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
			worker.expectFailure(CodeWorkerIdentityInvalid)
		})
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("a wrapper interpreter ran")
	}
}

// TestScriptWorkerRejectsPackageSelectedPrograms proves package-shaped
// values — a non-closed interpreter identifier, a non-absolute or
// non-regular runtime entry — refuse as package influence before the
// interpreter runs.
func TestScriptWorkerRejectsPackageSelectedPrograms(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "ran")
	for _, testCase := range []struct {
		name   string
		mutate func(t *testing.T, fixture *workerFixture, wire *workerRequest)
	}{
		{"shell-interpreter", func(_ *testing.T, _ *workerFixture, wire *workerRequest) {
			wire.InterpreterID = "bash-v1"
		}},
		{"empty-interpreter", func(_ *testing.T, _ *workerFixture, wire *workerRequest) {
			wire.InterpreterID = ""
		}},
		{"relative-runtime-entry", func(_ *testing.T, _ *workerFixture, wire *workerRequest) {
			wire.RuntimeEntry = "scripts/tool"
		}},
		{"empty-runtime-entry", func(_ *testing.T, _ *workerFixture, wire *workerRequest) {
			wire.RuntimeEntry = ""
		}},
		{"directory-runtime-entry", func(_ *testing.T, fixture *workerFixture, wire *workerRequest) {
			wire.RuntimeEntry = fixture.workDir
		}},
		{"symlink-runtime-entry", func(t *testing.T, fixture *workerFixture, wire *workerRequest) {
			link := filepath.Join(fixture.workDir, "tool-link")
			if err := os.Symlink(fixture.entry, link); err != nil {
				t.Skipf("this host forbids symlink creation: %v", err)
			}
			wire.RuntimeEntry = link
		}},
		{"relative-interpreter-path", func(_ *testing.T, _ *workerFixture, wire *workerRequest) {
			wire.InterpreterPath = "bin/python3"
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			worker := startRawWorker(t, os.Args[0])
			fixture := newWorkerFixture(t, os.Args[0])
			wire := fixture.request()
			wire.Environment = append(wire.Environment, "STUB_MARKER="+marker)
			testCase.mutate(t, fixture, wire)
			worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
			worker.expectFailure(CodePackageInfluenceForbidden)
		})
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the interpreter ran for a package-selected program")
	}
}

// TestWorkerRejectsMalformedInventory proves the worker validates the
// parent-probed inventory closed-form before applying anything: every
// malformed shape refuses with worker-protocol-invalid before the
// interpreter runs.
func TestWorkerRejectsMalformedInventory(t *testing.T) {
	host := ScriptInventoryPlatform(runtime.GOOS)
	valid := func() []ScriptInventoryInput {
		inputs := make([]ScriptInventoryInput, 0, len(scriptInventoryOrder))
		for _, name := range scriptInventoryOrder {
			inputs = append(inputs, ScriptInventoryInput{
				Name: name, Availability: scriptInventoryPlatforms[host][name].Availability,
			})
		}
		return inputs
	}
	for _, testCase := range []struct {
		name     string
		mutate   func(request *workerRequest)
		platform string
	}{
		{"unknown-platform", func(_ *workerRequest) {}, "plan9"},
		{"short-inventory", func(request *workerRequest) { request.Inventory = request.Inventory[:7] }, host},
		{"unknown-control", func(request *workerRequest) { request.Inventory[0].Name = "script-fast-start" }, host},
		{"duplicate-control", func(request *workerRequest) { request.Inventory[1].Name = request.Inventory[0].Name }, host},
		{"unknown-availability", func(request *workerRequest) { request.Inventory[0].Availability = "sometimes" }, host},
		{"cgroup-without-installable", func(request *workerRequest) { request.CgroupPath = "Delegated/child" }, host},
		{"bound-without-installable", func(request *workerRequest) { request.FileSizeBound = 1024 }, host},
		{"relative-write-path", func(request *workerRequest) { request.WritePaths = []string{"relative/path"} }, host},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			worker := startRawWorker(t, os.Args[0])
			fixture := newWorkerFixture(t, os.Args[0])
			wire := fixture.request()
			wire.Inventory = valid()
			wire.InventoryPlatform = testCase.platform
			wire.FileSizeBound = 0
			testCase.mutate(wire)
			worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
			worker.expectFailure(CodeWorkerProtocolInvalid)
		})
	}
}

// TestWorkerConfirmsFileSizeBound proves the worker confirms the installed
// per-file bound: with the observed bound the ready record reports the
// control applied.
func TestWorkerConfirmsFileSizeBound(t *testing.T) {
	if _, ok := currentScriptFileSizeLimit(); !ok {
		t.Skip("this case is exercised on the unix inventories, which alone observe RLIMIT_FSIZE")
	}
	worker := startRawWorker(t, os.Args[0])
	fixture := newWorkerFixture(t, os.Args[0])
	wire := fixture.request()
	wire.InventoryPlatform = ScriptInventoryPlatform(runtime.GOOS)
	for index := range wire.Inventory {
		if wire.Inventory[index].Name == ScriptControlPerFileSizeLimit {
			wire.Inventory[index].Availability = ScriptAvailabilityAvailable
			wire.Inventory[index].Present = true
		}
	}
	current, _ := currentScriptFileSizeLimit()
	wire.FileSizeBound = current
	worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
	ready := worker.receive()
	if ready.Kind != kindReady || ready.Ready == nil || ready.Ready.Evidence == nil {
		t.Fatalf("worker sent %q, want ready with evidence", ready.Kind)
	}
	for _, entry := range ready.Ready.Evidence.Controls {
		if entry.Name == ScriptControlPerFileSizeLimit && entry.Status != ScriptStatusApplied {
			t.Fatalf("per-file-size-limit evidence status = %q, want applied", entry.Status)
		}
	}
	worker.permit(fixture.nonce)
	if result := worker.receive(); result.Kind != kindResult {
		t.Fatalf("worker sent %q, want result", result.Kind)
	}
}

// TestWorkerRejectsWrongFileSizeBound proves a mismatched per-file bound
// refuses with capability-evidence-invalid: the confirmation compares the
// installed bound, it does not trust the request.
func TestWorkerRejectsWrongFileSizeBound(t *testing.T) {
	worker := startRawWorker(t, os.Args[0])
	fixture := newWorkerFixture(t, os.Args[0])
	wire := fixture.request()
	for index := range wire.Inventory {
		if wire.Inventory[index].Name == ScriptControlPerFileSizeLimit {
			wire.Inventory[index].Availability = ScriptAvailabilityAvailable
			wire.Inventory[index].Present = true
		}
	}
	wrong := uint64(1024)
	if current, ok := currentScriptFileSizeLimit(); ok {
		wrong = current + 1
		if wrong == 0 || wrong == current {
			wrong = current - 1
		}
	}
	if wrong == 0 {
		t.Fatal("no distinct wrong bound is expressible")
	}
	wire.FileSizeBound = wrong
	worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
	worker.expectFailure(CodeCapabilityEvidenceInvalid)
}

// TestScriptWorkerRejectsReplayedNonce proves the session binds every
// message to the fresh request nonce: a shutdown under another nonce
// tears the session down without a second run.
func TestScriptWorkerRejectsReplayedNonce(t *testing.T) {
	worker := startRawWorker(t, os.Args[0])
	fixture := newWorkerFixture(t, os.Args[0])
	wire := fixture.request()
	worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
	if ready := worker.receive(); ready.Kind != kindReady {
		t.Fatalf("worker sent %q, want ready", ready.Kind)
	}
	worker.permit(fixture.nonce)
	if result := worker.receive(); result.Kind != kindResult {
		t.Fatalf("worker sent %q, want result", result.Kind)
	}
	worker.send(workerMessage{Kind: kindShutdown, Nonce: strings.Repeat("00", 32)})
	worker.expectFailure(CodeWorkerProtocolInvalid)
}

// TestScriptWorkerModeIsNotReachableThroughPackageData proves the hidden
// mode is an argv-selected implementation boundary: no request field names
// it, and a request carrying it as data is still just data.
func TestScriptWorkerModeIsNotReachableThroughPackageData(t *testing.T) {
	worker := startRawWorker(t, os.Args[0])
	fixture := newWorkerFixture(t, os.Args[0])
	wire := fixture.request()
	wire.Args = []string{WorkerMode}
	wire.Environment = append(wire.Environment, "STUB_EXIT=0")
	worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
	if ready := worker.receive(); ready.Kind != kindReady {
		t.Fatalf("worker sent %q, want ready", ready.Kind)
	}
	worker.permit(fixture.nonce)
	result := worker.receive()
	if result.Kind != kindResult || result.Result == nil {
		t.Fatalf("worker sent %q, want result", result.Kind)
	}
	report := decodeStubReport(t, result.Result.Stdout)
	if len(report.Argv) == 0 || report.Argv[len(report.Argv)-1] != WorkerMode {
		t.Fatalf("argv = %q, want the mode carried as inert data", report.Argv)
	}
	worker.send(workerMessage{Kind: kindShutdown, Nonce: fixture.nonce})
}
