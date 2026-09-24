package scriptworker

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/godriver"
)

// workerStderrLimit bounds the worker's own diagnostic stream.
const workerStderrLimit = int64(64 * 1024)

// workerShutdownGrace bounds the join after the session completes.
const workerShutdownGrace = 5 * time.Second

// defaultOutputLimit bounds one captured interpreter output. Direct
// pass-through streaming stays permitted where the manager binds the streams
// itself; this bound covers only what the session captures.
const defaultOutputLimit = int64(16 * 1024 * 1024)

// launchHooks carries test-only seams. afterResolve runs after the initial
// identity resolution and before the launch-boundary recheck, so a test can
// replace the manager bytes in that window and prove the recheck refuses.
// Production passes nil. The hook cannot skip any check; it only makes the
// race under test deterministic.
type launchHooks struct {
	afterResolve func()
}

// runSession performs the manager half of one worker session: it resolves
// the interpreter identity, probes the native-control inventory for this
// invocation, resolves the manager identity, creates the operation-private
// runtime area, derives the containment profile from the declared
// capabilities, re-verifies identity at the launch boundary, starts exactly
// that executable in the fixed hidden mode, sends one canonical bounded
// request with a fresh nonce, requires the identity proof together with the
// capability-evidence record, permits the run only after validating both,
// and terminates the complete worker domain before returning. A host that
// cannot provide a mandatory control refuses with
// `script_execution_control_unavailable` at the probe, before any worker
// starts and before any private area is created.
func runSession(ctx context.Context, request LaunchRequest, hooks *launchHooks) (Result, error) {
	managerPath := request.ManagerPath
	if managerPath == "" {
		if path, err := os.Executable(); err == nil {
			managerPath = path
		} else {
			managerPath = os.Args[0]
		}
	}
	interpreter, err := ResolveInterpreter(request.InterpreterID, request.Interpreters, request.ForbiddenRoots)
	if err != nil {
		return Result{}, err
	}
	platform, probes, err := probeScriptInventory()
	if err != nil {
		return Result{}, err
	}
	identity, err := godriver.ResolveManagerIdentity(managerPath)
	if err != nil {
		return Result{}, workerDiagnostic(err)
	}
	if err := checkLaunchRequest(request); err != nil {
		return Result{}, err
	}
	private, err := createPrivateArea(request.PrivateBase, request.ForbiddenRoots)
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = os.RemoveAll(private.base) }()

	declared, err := ParseDeclaredCapabilities(request.CapabilitiesRaw)
	if err != nil {
		return Result{}, err
	}
	profile, err := DeriveProfile(DerivationInput{
		Declared:       declared,
		InterpreterID:  request.InterpreterID,
		Interpreter:    interpreter,
		HostEnv:        request.HostEnvironment,
		ExecSearchDirs: request.ExecSearchDirs,
		ForbiddenRoots: request.ForbiddenRoots,
		Private:        private,
		FarmParent:     private.base,
		ProjectRoot:    request.ProjectRoot,
	})
	if err != nil {
		return Result{}, err
	}

	nonce, err := sessionToken()
	if err != nil {
		return Result{}, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot derive a session nonce")
	}
	if hooks != nil && hooks.afterResolve != nil {
		hooks.afterResolve()
	}
	// Re-verify immediately before launch so a replacement race between the
	// first check and exec cannot widen the process graph.
	if err := identity.Verify(); err != nil {
		return Result{}, workerDiagnostic(err)
	}
	if err := VerifyInterpreter(interpreter); err != nil {
		return Result{}, err
	}
	for _, execIdentity := range profile.Exec {
		if err := VerifyExec(execIdentity); err != nil {
			return Result{}, err
		}
	}

	client := &workerClient{nonce: nonce}
	domain, err := prepareScriptDomain(probes)
	if err != nil {
		return Result{}, err
	}
	client.domain = domain
	defer client.teardown()

	command := exec.CommandContext(ctx, identity.Path, WorkerMode) // #nosec G204 -- identity-verified re-execution of this manager
	command.Dir = filepath.Dir(identity.Path)
	command.Env = scriptWorkerEnvironment()
	command.SysProcAttr = scriptWorkerSysProcAttr()
	command.Cancel = func() error {
		terminateScriptDomain(command, domain)
		return nil
	}
	command.WaitDelay = workerShutdownGrace
	client.stderr = &boundedBuffer{budget: &outputBudget{remaining: workerStderrLimit}}
	command.Stderr = client.stderr
	stdin, err := command.StdinPipe()
	if err != nil {
		return Result{}, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot open the worker request channel")
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return Result{}, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot open the worker response channel")
	}
	client.command, client.stdin, client.stdout = command, stdin, stdout
	if err := domain.launch(command); err != nil {
		return Result{}, err
	}

	inventory := make([]ScriptInventoryInput, 0, len(probes))
	for _, probe := range probes {
		inventory = append(inventory, ScriptInventoryInput{
			Name: probe.Name, Availability: probe.Availability, Present: probe.Present,
		})
	}
	cgroupPath, cgroupRoot := domain.cgroupParams()
	wire := workerRequest{
		Version:           protocolVersion,
		ExecutablePath:    identity.Path,
		ExecutableSHA256:  identity.SHA256,
		ExecutableSize:    identity.Size,
		InterpreterID:     interpreter.ID,
		InterpreterPath:   interpreter.Path,
		InterpreterSHA:    interpreter.SHA256,
		InterpreterSize:   interpreter.Size,
		RuntimeEntry:      request.RuntimeEntry,
		Args:              append([]string(nil), request.Args...),
		Environment:       append([]string(nil), profile.Environment...),
		WorkingDir:        profile.WorkingDir,
		PrivateBase:       private.base,
		PrivateTmp:        private.tmp,
		PrivateConfig:     private.config,
		PrivateCache:      private.cache,
		FarmDir:           profile.Path,
		NetworkOffline:    profile.NetworkOffline,
		ProjectRoot:       profile.ProjectRoot,
		Stdin:             append([]byte(nil), request.Stdin...),
		StdinIsNull:       request.Stdin == nil,
		OutputLimit:       defaultOutputLimit,
		InventoryPlatform: platform,
		Inventory:         inventory,
		WritePaths:        append([]string(nil), profile.WritePaths...),
		CgroupPath:        cgroupPath,
		CgroupRoot:        cgroupRoot,
		FileSizeBound:     domain.fileSizeParam(),
	}
	if err := client.send(workerMessage{Kind: kindRequest, Nonce: nonce, Request: &wire}); err != nil {
		return Result{}, err
	}
	ready, err := client.receive(kindReady)
	if err != nil {
		return Result{}, err
	}
	if ready.Ready == nil {
		return Result{}, diagnostic(CodeWorkerProtocolInvalid, "worker acknowledgement carries no identity proof")
	}
	if err := identity.MatchesExpectation(ready.Ready.ExecutablePath, ready.Ready.ExecutableSHA256, ready.Ready.ExecutableSize); err != nil {
		return Result{}, workerDiagnostic(err)
	}
	if err := interpreter.matchesExpectation(interpreter.Path, ready.Ready.InterpreterSHA, ready.Ready.InterpreterSize); err != nil {
		return Result{}, err
	}
	// The evidence gate: the parent validates the worker's
	// capability-evidence record against this invocation's own probes
	// before permitting the run. A missing record, a record that
	// contradicts the probe, or a record that claims a deferred guarantee
	// refuses here, and the interpreter never runs.
	if ready.Ready.Evidence == nil {
		return Result{}, diagnostic(CodeCapabilityEvidenceInvalid, "worker acknowledgement carries no capability evidence")
	}
	if err := validateScriptEvidence(*ready.Ready.Evidence, platform, probes); err != nil {
		return Result{}, err
	}
	parentRecord := buildScriptEvidence(platform, probes, installableScriptControls(probes))
	if !scriptEvidenceEqual(parentRecord, *ready.Ready.Evidence) {
		return Result{}, diagnostic(CodeCapabilityEvidenceInvalid, "worker capability evidence contradicts this invocation's probe")
	}
	// The permit gate: the worker starts the interpreter only after the
	// parent validates the identity proof and the evidence record and
	// sends this frame. A proof or record that fails validation never
	// earns a permit, so the interpreter never runs.
	if err := client.send(workerMessage{Kind: kindPermit, Nonce: nonce}); err != nil {
		return Result{}, err
	}
	message, err := client.receive(kindResult)
	if err != nil {
		return Result{}, err
	}
	if message.Result == nil {
		return Result{}, diagnostic(CodeWorkerProtocolInvalid, "worker result carries no invocation result")
	}
	result := message.Result
	if result.Evidence != nil {
		return Result{}, diagnostic(CodeCapabilityEvidenceInvalid, "the invocation returned a second capability evidence record")
	}
	if result.Started != 1 {
		return Result{}, diagnostic(CodeWorkerIdentityInvalid,
			"worker started %d programs in this session, want exactly 1", result.Started)
	}
	if result.Overflow {
		return Result{}, diagnostic(CodeWorkerProtocolInvalid, "interpreter output exceeded the capture bound")
	}
	return Result{Stdout: result.Stdout, Stderr: result.Stderr, ExitCode: result.ExitCode, Report: profile.Report, Evidence: parentRecord}, nil
}

// checkLaunchRequest fails fast on a malformed manager-side request before
// any worker exists. The worker re-validates everything; this check only
// moves the refusal earlier.
func checkLaunchRequest(request LaunchRequest) error {
	if request.RuntimeEntry == "" || !filepath.IsAbs(request.RuntimeEntry) {
		return diagnostic(CodePackageInfluenceForbidden, "runtime entry is not an absolute manager-derived path")
	}
	for _, arg := range request.Args {
		if strings.ContainsRune(arg, 0) {
			return diagnostic(CodeWorkerProtocolInvalid, "argument vector carries a NUL byte")
		}
	}
	if request.PrivateBase == "" || !filepath.IsAbs(request.PrivateBase) {
		return diagnostic(CodeWorkerProtocolInvalid, "private base is not absolute")
	}
	if request.ProjectRoot != "" && !filepath.IsAbs(request.ProjectRoot) {
		return diagnostic(CodeWorkerProtocolInvalid, "project root is not absolute")
	}
	return nil
}

// privateArea is the operation-private runtime area: temporary,
// configuration, and cache roots resolved independently of package data.
type privateArea struct {
	base   string
	tmp    string
	config string
	cache  string
}

// createPrivateArea creates the operation-private runtime area below an
// absolute manager-owned base outside every forbidden root.
func createPrivateArea(base string, forbiddenRoots []string) (privateArea, error) {
	physical, err := godriver.CanonicalPhysicalPath(base)
	if err != nil {
		return privateArea{}, diagnosticErr(CodeWorkerProtocolInvalid, err, "private base is unavailable")
	}
	info, err := os.Lstat(physical)
	if err != nil || !info.IsDir() {
		return privateArea{}, diagnosticErr(CodeWorkerProtocolInvalid, err, "private base is not a directory")
	}
	for _, root := range forbiddenRoots {
		resolved, resolveErr := godriver.CanonicalPhysicalPath(root)
		if resolveErr != nil {
			continue
		}
		if physical == resolved || isBelow(physical, resolved) {
			return privateArea{}, diagnostic(CodeWorkerProtocolInvalid, "private base is under a repository or runtime root")
		}
	}
	operation, err := os.MkdirTemp(physical, ".curator-script-")
	if err != nil {
		return privateArea{}, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot create the operation-private runtime area")
	}
	area := privateArea{base: operation}
	for _, leaf := range []struct {
		name   string
		target *string
	}{
		{"tmp", &area.tmp}, {"config", &area.config}, {"cache", &area.cache},
	} {
		path := filepath.Join(operation, leaf.name)
		if err := os.Mkdir(path, 0o700); err != nil {
			_ = os.RemoveAll(operation)
			return privateArea{}, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot create the operation-private %s root", leaf.name)
		}
		*leaf.target = path
	}
	return area, nil
}

// workerClient is one parent-side worker session.
type workerClient struct {
	command *exec.Cmd
	domain  *scriptDomain
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  *boundedBuffer

	nonce string

	finished bool
	drained  bool
}

func (client *workerClient) send(message workerMessage) error {
	if err := writeMessage(client.stdin, message); err != nil {
		if code := DiagnosticCode(err); code != "" {
			return err
		}
		return client.protocolFailure(err, "cannot send the %s message", message.Kind)
	}
	return nil
}

func (client *workerClient) receive(kind string) (workerMessage, error) {
	message, err := readMessage(client.stdout)
	if err != nil {
		return workerMessage{}, client.protocolFailure(err, "cannot read the %s message", kind)
	}
	if message.Kind == kindFailure {
		if message.Failure == nil || message.Failure.Code == "" {
			return workerMessage{}, diagnostic(CodeWorkerProtocolInvalid, "worker reported an unstructured failure")
		}
		return workerMessage{}, &Diagnostic{Code: message.Failure.Code, Detail: message.Failure.Detail}
	}
	if message.Kind != kind {
		return workerMessage{}, diagnostic(CodeWorkerProtocolInvalid, "worker sent %q, want %q", message.Kind, kind)
	}
	if message.Nonce != client.nonce {
		return workerMessage{}, diagnostic(CodeWorkerProtocolInvalid, "worker response carries an unknown nonce")
	}
	return message, nil
}

// protocolFailure prefers the worker's own stable diagnostic when the session
// channel closed because the worker rejected the session.
func (client *workerClient) protocolFailure(cause error, format string, args ...any) error {
	if detail := string(client.stderr.Bytes()); detail != "" {
		return diagnosticErr(CodeWorkerProtocolInvalid, cause, format+": %s", append(args, detail)...)
	}
	return diagnosticErr(CodeWorkerProtocolInvalid, cause, format, args...)
}

// teardown terminates and joins the complete worker domain and discards the
// session channel. It is safe to call more than once. The domain is
// terminated before the worker is reaped, so no interpreter descendant can
// outlive the invocation.
func (client *workerClient) teardown() {
	if client.finished {
		return
	}
	client.finished = true
	if client.stdin != nil {
		_ = writeMessage(client.stdin, workerMessage{Kind: kindShutdown, Nonce: client.nonce})
		_ = client.stdin.Close()
	}
	if client.command != nil && client.command.Process != nil {
		if !client.drained {
			drained := make(chan struct{})
			go func() {
				if client.stdout != nil {
					_, _ = io.Copy(io.Discard, client.stdout)
				}
				close(drained)
			}()
			select {
			case <-drained:
			case <-time.After(workerShutdownGrace):
			}
		}
		terminateScriptDomain(client.command, client.domain)
		_ = client.command.Wait()
	}
	if client.stdout != nil {
		_ = client.stdout.Close()
	}
	if client.domain != nil {
		client.domain.close()
	}
}

func sessionToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

// scriptWorkerEnvironment is the fixed bootstrap environment of the worker
// process itself. It carries only indispensable operating-system process
// variables.
func scriptWorkerEnvironment() []string {
	values := indispensableScriptEnvironment()
	if values == nil {
		values = map[string]string{}
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	environment := make([]string, 0, len(keys))
	for _, key := range keys {
		environment = append(environment, key+"="+values[key])
	}
	return environment
}
