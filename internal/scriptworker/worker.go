package scriptworker

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/relux-works/curator/internal/godriver"
)

// RunWorker is the fixed hidden script-worker mode of the installed manager.
// It reads one length-bounded session from input, proves its own executable
// identity, verifies the resolved interpreter file's identity, binds the
// standard streams explicitly, runs the interpreter exactly once against the
// manager-derived runtime entry, and starts no other program.
//
// It returns the process exit code; 0 means the session completed as
// specified. The interpreter's own exit status travels in the result message,
// never in this code.
func RunWorker(input io.Reader, output io.Writer) int {
	session := &workerSession{input: input, output: output}
	if err := session.run(); err != nil {
		if errors.Is(err, errPermitDeclined) {
			return 0
		}
		code := DiagnosticCode(err)
		if code == "" {
			code = CodeWorkerProtocolInvalid
		}
		_ = writeMessage(output, workerMessage{
			Kind: kindFailure, Nonce: session.nonce,
			Failure: &workerFailure{Code: code, Detail: err.Error()},
		})
		return exitWorkerFailure
	}
	return 0
}

const exitWorkerFailure = 3

type workerSession struct {
	input  io.Reader
	output io.Writer

	nonce   string
	request *workerRequest
	started int
	// nullInput is the pre-bound null-device handle for an explicit null
	// standard input. It opens before the inventory controls install
	// (Landlock never restricts an already-open descriptor), so the
	// interpreter spawn itself needs no open inside the enforced domain;
	// opens from inside the domain are covered by the write
	// confinement's file-typed null-device rule instead.
	nullInput *os.File
}

func (session *workerSession) run() error {
	if err := session.accept(); err != nil {
		return err
	}
	if err := session.awaitPermit(); err != nil {
		return err
	}
	result, err := session.serveRun()
	if err != nil {
		return err
	}
	if err := writeMessage(session.output, workerMessage{Kind: kindResult, Nonce: session.nonce, Result: result}); err != nil {
		return err
	}
	return session.awaitShutdown()
}

// awaitPermit waits for the parent's permit before the interpreter may
// start. The parent sends it only after validating the identity proof the
// worker just sent (and, from R3, the evidence record), so a refused proof
// never becomes a run. A shutdown here is a clean early end: the parent
// declined the run, and the interpreter never starts.
func (session *workerSession) awaitPermit() error {
	message, err := readMessage(session.input)
	if err != nil {
		return err
	}
	if message.Kind == kindShutdown && message.Nonce == session.nonce {
		return errPermitDeclined
	}
	if message.Kind != kindPermit {
		return diagnostic(CodeWorkerProtocolInvalid, "session expected %q, got %q", kindPermit, message.Kind)
	}
	if message.Nonce != session.nonce {
		return diagnostic(CodeWorkerProtocolInvalid, "session message carries a replayed or unknown nonce")
	}
	return nil
}

// errPermitDeclined ends the session cleanly when the parent declines the
// run after the identity proof. RunWorker maps it to a clean exit with no
// failure frame: the parent already holds its own refusal.
var errPermitDeclined = errors.New("the parent declined the run before the permit")

// accept validates the single request, proves the worker's own executable
// identity, verifies the interpreter file, and acknowledges the nonce.
func (session *workerSession) accept() error {
	message, err := readMessage(session.input)
	if err != nil {
		return err
	}
	if message.Kind != kindRequest || message.Request == nil {
		return diagnostic(CodeWorkerProtocolInvalid, "session must open with a request, got %q", message.Kind)
	}
	request := message.Request
	if request.Version != protocolVersion {
		return diagnostic(CodeWorkerProtocolInvalid, "unsupported worker protocol version %q", request.Version)
	}
	if len(message.Nonce) != sessionNonceLength || !isHex(message.Nonce) {
		return diagnostic(CodeWorkerProtocolInvalid, "session nonce is malformed")
	}
	session.nonce = message.Nonce
	session.request = request

	identity, err := godriver.ResolveManagerIdentity(workerExecutable())
	if err != nil {
		return workerDiagnostic(err)
	}
	if err := identity.MatchesExpectation(request.ExecutablePath, request.ExecutableSHA256, request.ExecutableSize); err != nil {
		return workerDiagnostic(err)
	}
	if err := session.validateRequest(); err != nil {
		return err
	}
	observed, err := verifyInterpreterFile(request)
	if err != nil {
		return err
	}
	// The null standard input binds here, before any control installs:
	// Landlock never restricts an already-open descriptor, so the bound
	// handle survives the spawn-time domain and the interpreter spawn
	// itself needs no open inside the domain. Opens of the null device
	// from inside the domain — redirected descendant stdio — are
	// covered by the file-typed null-device rule the write confinement
	// adds (only directory-typed rights on it would return EINVAL).
	if request.StdinIsNull || request.Stdin == nil {
		null, err := os.Open(os.DevNull)
		if err != nil {
			return diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot bind standard input to the null device")
		}
		session.nullInput = null
	}
	// The inventory controls install and confirm here, before the
	// readiness proof: the record the worker returns with its
	// acknowledgement reports exactly the mechanisms already in effect,
	// plus the Landlock controls whose enforcement the spawn will
	// perform (their construction is confirmed here; see
	// runInterpreter). Network isolation installs at the interpreter
	// spawn instead and confirms there before any result is returned.
	applied, err := applyScriptControls(request)
	if err != nil {
		return err
	}
	record := buildScriptEvidence(request.InventoryPlatform, scriptProbesFromInput(request.Inventory), applied)
	return writeMessage(session.output, workerMessage{
		Kind: kindReady, Nonce: session.nonce,
		Ready: &workerReady{
			ExecutablePath: identity.Path, ExecutableSHA256: identity.SHA256, ExecutableSize: identity.Size,
			InterpreterSHA: observed.SHA256, InterpreterSize: observed.Size,
			Evidence: &record,
		},
	})
}

func (session *workerSession) serveRun() (*workerResult, error) {
	// The launch-boundary recheck: the interpreter file is re-verified
	// immediately before exec so a replacement between accept and exec cannot
	// widen the graph. Accept already verified it once; this second proof is
	// the one the exec depends on.
	if _, err := verifyInterpreterFile(session.request); err != nil {
		return nil, err
	}
	return session.runInterpreter()
}

func (session *workerSession) awaitShutdown() error {
	message, err := readMessage(session.input)
	if err != nil {
		return err
	}
	if message.Kind != kindShutdown {
		return diagnostic(CodeWorkerProtocolInvalid, "session expected %q, got %q", kindShutdown, message.Kind)
	}
	if message.Nonce != session.nonce {
		return diagnostic(CodeWorkerProtocolInvalid, "session message carries a replayed or unknown nonce")
	}
	return nil
}

// workerExecutable resolves the running worker binary from the process itself.
func workerExecutable() string {
	if path, err := os.Executable(); err == nil {
		return path
	}
	return os.Args[0]
}

// workerDiagnostic re-codes a godriver identity failure at the script
// boundary. The check is shared; the diagnostic names the policy that
// refused.
func workerDiagnostic(err error) error {
	if godriver.DiagnosticCode(err) == "" {
		return err
	}
	return &Diagnostic{Code: CodeWorkerIdentityInvalid, Detail: err.Error(), Err: err}
}

// validateRequest is the worker's own guard over everything it was asked to
// do. A malformed version, path, argument vector, environment, root, or limit
// is rejected before the interpreter starts. Package-shaped values — a
// non-closed interpreter identifier, a non-absolute runtime entry, a runtime
// entry that is not a regular file — are refused as package influence,
// because the only thing that may select the executed program is the
// manager's own derivation.
func (session *workerSession) validateRequest() error {
	request := session.request
	if !closedInterpreters[request.InterpreterID] {
		return diagnostic(CodePackageInfluenceForbidden,
			"request selects interpreter identifier %q outside the closed set", request.InterpreterID)
	}
	if request.InterpreterPath == "" || !filepath.IsAbs(request.InterpreterPath) {
		return diagnostic(CodePackageInfluenceForbidden, "request interpreter path is not an absolute manager-derived path")
	}
	if !isDigest(request.InterpreterSHA) || request.InterpreterSize <= 0 {
		return diagnostic(CodeWorkerProtocolInvalid, "request carries a malformed interpreter expectation")
	}
	if request.RuntimeEntry == "" || !filepath.IsAbs(request.RuntimeEntry) {
		return diagnostic(CodePackageInfluenceForbidden, "request runtime entry is not an absolute manager-derived path")
	}
	if strings.ContainsRune(request.RuntimeEntry, 0) {
		return diagnostic(CodePackageInfluenceForbidden, "request runtime entry carries a NUL byte")
	}
	info, err := os.Lstat(request.RuntimeEntry)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return diagnostic(CodePackageInfluenceForbidden,
			"request runtime entry is not a regular manager-staged file")
	}
	for _, arg := range request.Args {
		if strings.ContainsRune(arg, 0) {
			return diagnostic(CodeWorkerProtocolInvalid, "request argument vector carries a NUL byte")
		}
	}
	if err := validateSessionEnvironment(request.Environment); err != nil {
		return err
	}
	if request.WorkingDir == "" || !filepath.IsAbs(request.WorkingDir) {
		return diagnostic(CodeWorkerProtocolInvalid, "request working directory is not absolute")
	}
	if info, err := os.Stat(request.WorkingDir); err != nil || !info.IsDir() {
		return diagnostic(CodeWorkerProtocolInvalid, "request working directory is not a directory")
	}
	if request.PrivateBase == "" || !filepath.IsAbs(request.PrivateBase) {
		return diagnostic(CodeWorkerProtocolInvalid, "request private runtime base is not absolute")
	}
	if info, err := os.Stat(request.PrivateBase); err != nil || !info.IsDir() {
		return diagnostic(CodeWorkerProtocolInvalid, "request private runtime base is not a directory")
	}
	for _, root := range []string{request.PrivateTmp, request.PrivateConfig, request.PrivateCache} {
		if root == "" || !filepath.IsAbs(root) {
			return diagnostic(CodeWorkerProtocolInvalid, "request private runtime area is not absolute")
		}
		if info, err := os.Stat(root); err != nil || !info.IsDir() {
			return diagnostic(CodeWorkerProtocolInvalid, "request private runtime root %q is not a directory", root)
		}
	}
	if request.FarmDir == "" || !filepath.IsAbs(request.FarmDir) {
		return diagnostic(CodeWorkerProtocolInvalid, "request PATH farm directory is not absolute")
	}
	if info, err := os.Stat(request.FarmDir); err != nil || !info.IsDir() {
		return diagnostic(CodeWorkerProtocolInvalid, "request PATH farm directory is not a directory")
	}
	if request.FarmDir != request.PrivateBase && !isBelow(request.FarmDir, request.PrivateBase) {
		return diagnostic(CodeWorkerProtocolInvalid, "request PATH farm directory is outside the private runtime area")
	}
	if request.ProjectRoot != "" {
		if !filepath.IsAbs(request.ProjectRoot) {
			return diagnostic(CodeWorkerProtocolInvalid, "request project root is not absolute")
		}
		if info, err := os.Stat(request.ProjectRoot); err != nil || !info.IsDir() {
			return diagnostic(CodeWorkerProtocolInvalid, "request project root is not a directory")
		}
	}
	// The worker re-applies the same derived-environment check the manager
	// ran as a self-check: PATH is exactly the farm, the private roots
	// are bound, and an offline derivation carries no proxy or resolver
	// configuration. A manager that mis-derived refuses here, before the
	// interpreter starts.
	if err := validateDerivedEnvironment(request.Environment, privateArea{
		base:   request.PrivateBase,
		tmp:    request.PrivateTmp,
		config: request.PrivateConfig,
		cache:  request.PrivateCache,
	}, request.FarmDir, request.ProjectRoot, request.NetworkOffline, runtime.GOOS); err != nil {
		return err
	}
	if err := validateRequestInventory(request); err != nil {
		return err
	}
	if request.OutputLimit <= 0 || request.OutputLimit > maxProtocolFrame {
		return diagnostic(CodeWorkerProtocolInvalid, "request output bound %d is outside the session bound", request.OutputLimit)
	}
	if int64(len(request.Stdin)) > maxProtocolFrame {
		return diagnostic(CodeWorkerProtocolInvalid, "request standard-input payload exceeds the session bound")
	}
	return nil
}

// validateRequestInventory checks the closed form of the parent-probed
// inventory the worker applies: exactly one entry per inventory control
// with a known availability, a coherent cgroup and file-bound
// parameterization, and absolute clean write paths.
func validateRequestInventory(request *workerRequest) error {
	switch request.InventoryPlatform {
	case ScriptPlatformLinux, ScriptPlatformMacOS, ScriptPlatformWindows:
	default:
		return diagnostic(CodeWorkerProtocolInvalid,
			"request inventory platform %q is not a known platform", request.InventoryPlatform)
	}
	if len(request.Inventory) != len(scriptInventoryOrder) {
		return diagnostic(CodeWorkerProtocolInvalid,
			"request inventory carries %d controls, want exactly one per inventory control", len(request.Inventory))
	}
	seen := make(map[string]bool, len(request.Inventory))
	installable := map[string]bool{}
	for _, input := range request.Inventory {
		if !inScriptInventory(input.Name) {
			return diagnostic(CodeWorkerProtocolInvalid,
				"request inventory names control %q outside the inventory", input.Name)
		}
		if seen[input.Name] {
			return diagnostic(CodeWorkerProtocolInvalid,
				"request inventory duplicates control %q", input.Name)
		}
		seen[input.Name] = true
		switch input.Availability {
		case ScriptAvailabilityAvailable, ScriptAvailabilityHostConditional, ScriptAvailabilityUnavailable:
		default:
			return diagnostic(CodeWorkerProtocolInvalid,
				"request inventory control %q reports availability %q", input.Name, input.Availability)
		}
		installable[input.Name] = input.Present &&
			(input.Availability == ScriptAvailabilityAvailable ||
				input.Availability == ScriptAvailabilityHostConditional)
	}
	for _, path := range request.WritePaths {
		if path == "" || !filepath.IsAbs(path) || filepath.Clean(path) != path {
			return diagnostic(CodeWorkerProtocolInvalid, "request write path %q is not an absolute clean path", path)
		}
		if strings.ContainsRune(path, 0) {
			return diagnostic(CodeWorkerProtocolInvalid, "request write path carries a NUL byte")
		}
	}
	// Only the Linux inventory installs aggregate limits through a cgroup;
	// the Windows inventory carries them on the private job object and
	// needs no request parameterization.
	cgroupWanted := request.InventoryPlatform == ScriptPlatformLinux &&
		(installable[ScriptControlActiveProcessCountLimit] || installable[ScriptControlAggregateMemoryLimit])
	if cgroupWanted && request.CgroupPath == "" {
		return diagnostic(CodeWorkerProtocolInvalid, "request installs cgroup limits but names no invocation cgroup")
	}
	if !cgroupWanted && request.CgroupPath != "" {
		return diagnostic(CodeWorkerProtocolInvalid, "request names an invocation cgroup no installable control uses")
	}
	if request.CgroupPath != "" && (!filepath.IsLocal(request.CgroupPath) || filepath.Clean(request.CgroupPath) != request.CgroupPath) {
		return diagnostic(CodeWorkerProtocolInvalid, "request invocation cgroup %q is not a clean relative path", request.CgroupPath)
	}
	if request.CgroupRoot != "" && (!filepath.IsAbs(request.CgroupRoot) || filepath.Clean(request.CgroupRoot) != request.CgroupRoot) {
		return diagnostic(CodeWorkerProtocolInvalid, "request cgroup root %q is not an absolute clean path", request.CgroupRoot)
	}
	if installable[ScriptControlPerFileSizeLimit] && request.FileSizeBound == 0 {
		return diagnostic(CodeWorkerProtocolInvalid, "request installs the per-file limit but names no byte bound")
	}
	if !installable[ScriptControlPerFileSizeLimit] && request.FileSizeBound != 0 {
		return diagnostic(CodeWorkerProtocolInvalid, "request names a per-file byte bound no installable control uses")
	}
	return nil
}

// validateSessionEnvironment checks the closed form of the session
// environment: KEY=value entries, no NUL, no repeated key. R1 applies what
// the manager sends; the declaration-derived closed value set is R2.
func validateSessionEnvironment(environment []string) error {
	seen := make(map[string]bool, len(environment))
	for _, item := range environment {
		key, _, present := strings.Cut(item, "=")
		if !present || key == "" {
			return diagnostic(CodeWorkerProtocolInvalid, "session environment contains a malformed entry")
		}
		if strings.ContainsRune(item, 0) {
			return diagnostic(CodeWorkerProtocolInvalid, "session environment contains a NUL byte")
		}
		if seen[key] {
			return diagnostic(CodeWorkerProtocolInvalid, "session environment repeats %s", key)
		}
		seen[key] = true
	}
	return nil
}

// verifyInterpreterFile proves the resolved interpreter file's identity: a
// canonical regular native executable image, single-linked, whose bytes hash
// to the manager's expectation. A wrapper — a POSIX shebang shim, a Windows
// batch file — would interpose another program between the worker and the
// interpreter, and an extensionless Windows path would execute a different
// file than the verified one, so wrappers are refused and the path below is
// the executed program. (The runtime entry's own bytes stay inert: the entry
// travels as an argument to the interpreter, never as an executed program.)
func verifyInterpreterFile(request *workerRequest) (InterpreterIdentity, error) {
	canonical, err := godriver.CanonicalPhysicalPath(request.InterpreterPath)
	if err != nil || canonical != request.InterpreterPath {
		return InterpreterIdentity{}, diagnosticErr(CodeWorkerIdentityInvalid, err,
			"the resolved interpreter path is not canonical and link-free")
	}
	observed, err := readInterpreterIdentity(request.InterpreterID, canonical)
	if err != nil {
		return InterpreterIdentity{}, err
	}
	expected := InterpreterIdentity{
		ID: request.InterpreterID, Path: request.InterpreterPath,
		SHA256: request.InterpreterSHA, Size: request.InterpreterSize,
	}
	if err := expected.matchesExpectation(observed.Path, observed.SHA256, observed.Size); err != nil {
		return InterpreterIdentity{}, err
	}
	return observed, nil
}

// runInterpreter starts exactly the verified interpreter against the
// manager-derived runtime entry. It is the worker's only process-creation
// site. Standard input is bound explicitly — to the request payload or to
// the platform null device — and the child's exit status is returned
// without reinterpretation. When the Landlock controls are installable
// the domain is enforced here, on the locked spawn thread, immediately
// before the fork, so the child inherits it. When network isolation is
// installable the child starts in a fresh user and network namespace,
// and the spawn confirms the fresh namespace before the child can
// produce a result.
func (session *workerSession) runInterpreter() (*workerResult, error) {
	request := session.request
	session.started++
	argv := append([]string{request.RuntimeEntry}, request.Args...)

	command := exec.Command(request.InterpreterPath, argv...) // #nosec G204 -- verified interpreter against a verified runtime entry, both manager-derived
	command.Dir = request.WorkingDir
	command.Env = append([]string(nil), request.Environment...)
	if request.StdinIsNull || request.Stdin == nil {
		if session.nullInput == nil {
			return nil, diagnostic(CodeWorkerProtocolInvalid, "null standard input was not bound before the controls installed")
		}
		defer func() { _ = session.nullInput.Close() }()
		command.Stdin = session.nullInput
	} else {
		command.Stdin = bytes.NewReader(request.Stdin)
	}
	command.SysProcAttr = interpreterSysProcAttr()
	isolated := requestNetNSInstallable(request)
	if isolated {
		command.SysProcAttr = scriptNetNSAttr()
		if command.SysProcAttr == nil {
			return nil, diagnostic(CodeCapabilityEvidenceInvalid, "network isolation is installable only on Linux")
		}
	}
	landlockExec, landlockWrite := requestLandlockInstallable(request)
	if landlockExec || landlockWrite {
		// The spawn-time Landlock enforcement: the privilege change,
		// the restriction, and the fork below must all happen on one
		// thread, so this goroutine locks to its thread first. The
		// lock is never released: returning a restricted thread to
		// the runtime pool would sandbox unrelated goroutines, and
		// the worker exits after this one session. An enforcement
		// failure refuses here and the interpreter never starts.
		runtime.LockOSThread()
		if err := enforceScriptLandlock(landlockExec, landlockWrite, landlockGrantsForRequest(request)); err != nil {
			return nil, diagnosticErr(CodeWorkerProtocolInvalid, err,
				"cannot enforce the Landlock domain for the installable controls")
		}
	}
	budget := &outputBudget{remaining: request.OutputLimit}
	stdout := &boundedBuffer{budget: budget}
	stderr := &boundedBuffer{budget: budget}
	command.Stdout = stdout
	command.Stderr = stderr

	if !isolated {
		return session.finishInterpreter(command, stdout, stderr, budget)
	}
	if err := command.Start(); err != nil {
		return nil, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot start the verified interpreter")
	}
	if command.Process == nil {
		return nil, diagnostic(CodeWorkerProtocolInvalid, "the started interpreter has no process identity")
	}
	if err := confirmScriptNetNS(command.Process.Pid); err != nil {
		_ = command.Process.Kill()
		_ = command.Wait()
		return nil, err
	}
	waitErr := command.Wait()
	return session.collectInterpreter(waitErr, command, stdout, stderr, budget)
}

// requestNetNSInstallable reports whether the request inventory installs
// network isolation for the interpreter child.
func requestNetNSInstallable(request *workerRequest) bool {
	for _, input := range request.Inventory {
		if input.Name != ScriptControlNetworkIsolationDomain {
			continue
		}
		return input.Present && input.Availability == ScriptAvailabilityHostConditional
	}
	return false
}

func (session *workerSession) finishInterpreter(command *exec.Cmd, stdout, stderr *boundedBuffer, budget *outputBudget) (*workerResult, error) {
	runErr := command.Run()
	return session.collectInterpreter(runErr, command, stdout, stderr, budget)
}

func (session *workerSession) collectInterpreter(runErr error, command *exec.Cmd, stdout, stderr *boundedBuffer, budget *outputBudget) (*workerResult, error) {
	result := &workerResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), Started: session.started}
	if errors.Is(stdout.err, errOutputLimit) || errors.Is(stderr.err, errOutputLimit) || errors.Is(budget.err, errOutputLimit) {
		result.Overflow = true
	}
	if runErr != nil {
		if state := command.ProcessState; state != nil {
			result.ExitCode = state.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}
	return result, nil
}

func isHex(value string) bool {
	if len(value) == 0 {
		return false
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdefABCDEF", character) {
			return false
		}
	}
	return true
}

func isDigest(value string) bool {
	digest, ok := strings.CutPrefix(value, "sha256:")
	if !ok || len(digest) != 64 {
		return false
	}
	return isHex(digest)
}
