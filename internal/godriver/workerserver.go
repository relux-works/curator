package godriver

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// RunWorker is the fixed hidden worker mode of the installed manager. It reads
// one length-bounded session from input, confirms that every inventory control
// the manager installed before releasing it is really in effect, runs exactly
// one fixed go list and, after one authenticated permit, exactly one fixed go
// build, and starts no other program. It never executes the built artifact.
//
// The worker makes no control-availability decision: that determination and
// every installation belong to the manager parent and are complete before the
// worker executes. A control the worker cannot confirm is an evidence fault.
//
// It returns the process exit code; 0 means the session completed as specified.
func RunWorker(input io.Reader, output io.Writer) int {
	session := &workerSession{input: input, output: output}
	if err := session.run(); err != nil {
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
}

func (session *workerSession) run() error {
	if err := session.accept(); err != nil {
		return err
	}
	if err := session.serveList(); err != nil {
		return err
	}
	if err := session.serveBuild(); err != nil {
		return err
	}
	return session.awaitShutdown()
}

// accept performs steps 5 and 6 of the worker session: it validates the single
// request, proves its own executable identity, confirms the installed controls,
// and acknowledges the nonce together with the capability-evidence record.
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
	if len(message.Nonce) != sessionNonceLength || len(request.Secret) != sessionSecretLength {
		return diagnostic(CodeWorkerProtocolInvalid, "session nonce or secret is malformed")
	}
	session.nonce = message.Nonce
	session.request = request

	identity, err := resolveExecutableIdentity(workerExecutable())
	if err != nil {
		return err
	}
	if err := identity.matches(request.ExecutablePath, request.ExecutableSHA256, request.ExecutableSize); err != nil {
		return err
	}
	if err := session.validateRequest(); err != nil {
		return err
	}
	confirmed, err := observeNativeControls(session.limits(), request.Probes, session.protocolFiles())
	if err != nil {
		return err
	}
	evidence := evidenceFromApplied(request.Platform, request.Probes, confirmed)
	if err := validateCapabilityEvidence(evidence, request.Platform, request.Probes); err != nil {
		return err
	}
	return writeMessage(session.output, workerMessage{
		Kind: kindReady, Nonce: session.nonce,
		Ready: &workerReady{
			ExecutablePath: identity.Path, ExecutableSHA256: identity.SHA256, ExecutableSize: identity.Size,
			Applied: confirmed, Evidence: evidence,
		},
	})
}

func (session *workerSession) serveList() error {
	if _, err := session.expect(kindList); err != nil {
		return err
	}
	result, err := session.runGo(session.request.ListArgv)
	if err != nil {
		return err
	}
	return writeMessage(session.output, workerMessage{Kind: kindListResult, Nonce: session.nonce, Result: result})
}

func (session *workerSession) serveBuild() error {
	message, err := session.expect(kindPermit)
	if err != nil {
		return err
	}
	if !validPermit(message.Permit, session.request.Secret, session.nonce, session.request.BuildArgv) {
		return diagnostic(CodeWorkerProtocolInvalid, "build permit is not authenticated for this session and vector")
	}
	result, err := session.runGo(session.request.BuildArgv)
	if err != nil {
		return err
	}
	return writeMessage(session.output, workerMessage{Kind: kindBuildResult, Nonce: session.nonce, Result: result})
}

func (session *workerSession) awaitShutdown() error {
	if _, err := session.expect(kindShutdown); err != nil {
		return err
	}
	return nil
}

// expect enforces the session ordering and nonce binding. Any other kind,
// including a second list or a second build request, tears the session down
// without starting a compiler.
func (session *workerSession) expect(kind string) (workerMessage, error) {
	message, err := readMessage(session.input)
	if err != nil {
		return workerMessage{}, err
	}
	if message.Kind != kind {
		return workerMessage{}, diagnostic(CodeWorkerProtocolInvalid, "session expected %q, got %q", kind, message.Kind)
	}
	if message.Nonce != session.nonce {
		return workerMessage{}, diagnostic(CodeWorkerProtocolInvalid, "session message carries a replayed or unknown nonce")
	}
	return message, nil
}

func (session *workerSession) limits() ResourceLimits {
	wire := session.request.Limits
	return ResourceLimits{
		Timeout:       time.Duration(wire.TimeoutMillis) * time.Millisecond,
		OutputBytes:   wire.OutputBytes,
		ArtifactBytes: wire.ArtifactBytes,
		FileBytes:     wire.FileBytes,
		DiskBytes:     wire.DiskBytes,
		MemoryBytes:   wire.MemoryBytes,
		Processes:     wire.Processes,
	}
}

func (session *workerSession) protocolFiles() []*os.File {
	files := make([]*os.File, 0, 2)
	if file, ok := session.input.(*os.File); ok {
		files = append(files, file)
	}
	if file, ok := session.output.(*os.File); ok {
		files = append(files, file)
	}
	return files
}

// workerExecutable resolves the running worker binary from the process itself.
func workerExecutable() string {
	if path, err := os.Executable(); err == nil {
		return path
	}
	return os.Args[0]
}

// validateRequest is the worker's own guard over everything it was asked to do.
// A mutated executable, argument vector, working directory, environment, root,
// limit, or control profile is rejected before any compiler starts.
func (session *workerSession) validateRequest() error {
	request := session.request
	if request.Platform != InventoryPlatform(runtime.GOOS) || request.Platform == "" {
		return diagnostic(CodeCapabilityEvidenceInvalid, "request platform %q is not this host", request.Platform)
	}
	if err := validateProbeSet(request.Platform, request.Probes); err != nil {
		return err
	}
	if len(request.ReadOnlyRoots) != 2 || len(request.PrivateRoots) == 0 {
		return diagnostic(CodeWorkerProtocolInvalid, "request must carry the frozen source root, GOROOT, and private roots")
	}
	sourceRoot, goroot := request.ReadOnlyRoots[0], request.ReadOnlyRoots[1]
	if !filepath.IsAbs(sourceRoot) || !filepath.IsAbs(goroot) || goroot != request.GOROOT {
		return diagnostic(CodeWorkerProtocolInvalid, "request roots are not absolute manager-owned roots")
	}
	if request.GoExecutable != filepath.Join(goroot, "bin", platformGoName) {
		return diagnostic(CodeWorkerIdentityInvalid, "request selects a program other than the fingerprinted Go launcher")
	}
	if request.ToolDirectory != filepath.Join(goroot, "pkg", "tool", toolPlatform(request.Environment)) {
		return diagnostic(CodeWorkerIdentityInvalid, "request selects a tool directory outside the fingerprinted GOROOT")
	}
	if !filepath.IsAbs(request.Directory) || !isWithin(request.Directory, sourceRoot) && request.Directory != sourceRoot {
		return diagnostic(CodeWorkerProtocolInvalid, "request working directory is not inside the frozen source snapshot")
	}
	for _, root := range request.PrivateRoots {
		if !filepath.IsAbs(root) || root == sourceRoot || root == goroot || isWithin(root, sourceRoot) || isWithin(root, goroot) {
			return diagnostic(CodeWorkerProtocolInvalid, "private root %q overlaps the frozen source or GOROOT", root)
		}
	}
	if !reflect.DeepEqual(request.ListArgv, listArguments) {
		return diagnostic(CodeWorkerProtocolInvalid, "request carries a non-protocol go list vector")
	}
	if err := validateFixedBuildArgv(request.BuildArgv, request.ArtifactPath, request.PrivateRoots); err != nil {
		return err
	}
	if err := validateWorkerEnvironment(request.Environment, goroot, request.PrivateRoots); err != nil {
		return err
	}
	if _, err := normalizeBuildLimits(session.limits()); err != nil {
		return err
	}
	return nil
}

// validateProbeSet is the worker's defence in depth over the probe record it was
// handed. The normative availability decision already happened in the parent
// before this process was released to execute, so a probe record that
// contradicts rc5-native-control-inventory-v1 here is an evidence fault and
// never a mandatory-control rejection.
func validateProbeSet(platform string, probes []ControlProbe) error {
	records := nativeControlPlatforms[platform]
	if records == nil {
		return diagnostic(CodeCapabilityEvidenceInvalid, "no native-control inventory record for platform %q", platform)
	}
	if len(probes) != len(nativeControlInventory) {
		return diagnostic(CodeCapabilityEvidenceInvalid, "request carries %d probes, want exactly the inventory", len(probes))
	}
	for index, probe := range probes {
		if probe.Name != nativeControlInventory[index] {
			return diagnostic(CodeCapabilityEvidenceInvalid, "probe %d is %q, want %q", index, probe.Name, nativeControlInventory[index])
		}
		if probe.ProbedAt != ProbeTiming {
			return diagnostic(CodeCapabilityEvidenceInvalid, "probe %q reports probed_at %q", probe.Name, probe.ProbedAt)
		}
		record := records[probe.Name]
		switch probe.Availability {
		case AvailabilityAvailable:
			if record.Availability != AvailabilityAvailable {
				return diagnostic(CodeCapabilityEvidenceInvalid,
					"probe claims %q is available, but rc5-native-control-inventory-v1 marks it unavailable on %s", probe.Name, platform)
			}
		case AvailabilityUnavailable:
			if record.Availability == AvailabilityAvailable {
				return diagnostic(CodeCapabilityEvidenceInvalid,
					"rc5-native-control-inventory-v1 marks %q available on %s but the request reports it unavailable", probe.Name, platform)
			}
		default:
			return diagnostic(CodeCapabilityEvidenceInvalid, "probe %q reports availability %q", probe.Name, probe.Availability)
		}
	}
	return nil
}

func validateFixedBuildArgv(argv []string, artifact string, privateRoots []string) error {
	prefix := buildArgumentPrefix
	if len(argv) != len(prefix)+2 || !reflect.DeepEqual(argv[:len(prefix)], prefix) || argv[len(argv)-1] != "." {
		return diagnostic(CodeWorkerProtocolInvalid, "request carries a non-protocol go build vector")
	}
	output := argv[len(prefix)]
	if output != artifact || !filepath.IsAbs(output) {
		return diagnostic(CodeWorkerProtocolInvalid, "go build output is not the manager-derived absolute staging path")
	}
	for _, root := range privateRoots {
		if strictlyBelow(output, root) {
			return nil
		}
	}
	return diagnostic(CodeWorkerProtocolInvalid, "go build output escapes operation-private staging")
}

func toolPlatform(environment []string) string {
	values := environmentMap(environment)
	return values["GOOS"] + "_" + values["GOARCH"]
}

// validateWorkerEnvironment re-proves the closed offline environment inside the
// worker, so a mutated request cannot widen the compiler environment.
func validateWorkerEnvironment(environment []string, goroot string, privateRoots []string) error {
	values := make(map[string]string, len(environment))
	for _, item := range environment {
		key, value, present := strings.Cut(item, "=")
		if !present || key == "" {
			return diagnostic(CodeWorkerProtocolInvalid, "compiler environment contains a malformed entry")
		}
		if _, duplicate := values[key]; duplicate {
			return diagnostic(CodeWorkerProtocolInvalid, "compiler environment repeats %s", key)
		}
		values[key] = value
	}
	required := map[string]string{
		"GOENV": "off", "GOTOOLCHAIN": "local", "LC_ALL": "C", "LANG": "C",
		"GO111MODULE": "on", "GOFLAGS": "", "GOPROXY": "off", "GOSUMDB": "off", "GOPRIVATE": "",
		"GONOPROXY": "none", "GONOSUMDB": "none", "GOVCS": "*:off", "GOWORK": "off",
		"CGO_ENABLED": "0", "GO_EXTLINK_ENABLED": "0", "GOEXPERIMENT": "",
		"GOROOT": goroot,
	}
	for key, want := range required {
		if got, present := values[key]; !present || got != want {
			return diagnostic(CodeWorkerProtocolInvalid, "compiler environment has unexpected %s", key)
		}
	}
	for _, key := range []string{"GOPATH", "GOMODCACHE", "GOCACHE", "GOTMPDIR", "HOME", "XDG_CONFIG_HOME", "PATH", "TMPDIR"} {
		value, present := values[key]
		if !present || value == "" {
			return diagnostic(CodeWorkerProtocolInvalid, "compiler environment is missing operation-private %s", key)
		}
		if key == "PATH" {
			if !privateOrSibling(value, privateRoots) {
				return diagnostic(CodeWorkerProtocolInvalid, "compiler PATH is not an operation-private empty directory")
			}
			continue
		}
		if !containsExact(privateRoots, value) {
			return diagnostic(CodeWorkerProtocolInvalid, "compiler environment %s is not operation-private", key)
		}
	}
	for _, forbidden := range []string{
		"CC", "CXX", "FC", "AR", "PKG_CONFIG", "GOFLAGS_FILE", "GOAUTH",
		"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY", "SSH_AUTH_SOCK",
		"GIT_CONFIG_GLOBAL", "GIT_CONFIG_SYSTEM", "GOTOOLDIR", "GOROOT_FINAL", "GOBIN",
	} {
		if _, present := values[forbidden]; present {
			return diagnostic(CodeWorkerProtocolInvalid, "compiler environment contains forbidden %s", forbidden)
		}
	}
	if values["GOOS"] == "" || values["GOARCH"] == "" {
		return diagnostic(CodeWorkerProtocolInvalid, "compiler environment is missing the native target")
	}
	allowed := make(map[string]bool, len(required)+16)
	for key := range required {
		allowed[key] = true
	}
	for _, key := range []string{"GOOS", "GOARCH", "GOPATH", "GOMODCACHE", "GOCACHE", "GOTMPDIR", "HOME", "XDG_CONFIG_HOME", "PATH", "TMPDIR"} {
		allowed[key] = true
	}
	if tuning := tuningVariable(values["GOARCH"]); tuning != "" {
		if values[tuning] == "" {
			return diagnostic(CodeWorkerProtocolInvalid, "compiler environment is missing %s", tuning)
		}
		allowed[tuning] = true
	}
	for _, key := range platformEnvironmentNames() {
		allowed[key] = true
	}
	for key := range values {
		if !allowed[key] {
			return diagnostic(CodeWorkerProtocolInvalid, "compiler environment contains non-protocol variable %s", key)
		}
	}
	entries, err := os.ReadDir(values["PATH"])
	if err != nil || len(entries) != 0 {
		return diagnosticErr(CodeWorkerProtocolInvalid, err, "compiler PATH is not an empty manager directory")
	}
	return nil
}

func privateOrSibling(path string, privateRoots []string) bool {
	if containsExact(privateRoots, path) {
		return true
	}
	for _, root := range privateRoots {
		if filepath.Dir(root) == filepath.Dir(path) {
			return true
		}
	}
	return false
}

// runGo starts exactly the fingerprinted Go launcher with a manager-owned
// argument vector. It is the worker's only process-creation site.
//
// A launcher that fails its identity check ends the session with that stable
// diagnostic and creates no process. A process the operating system refuses to
// create has no exit status at all, so the refusal is reported as a bounded
// result rather than read from a nil process state.
func (session *workerSession) runGo(argv []string) (*workerResult, error) {
	request := session.request
	if err := verifyGoLauncher(request.GoExecutable); err != nil {
		return nil, err
	}
	session.started++
	limits := session.limits()
	ctx, cancel := context.WithTimeout(context.Background(), limits.Timeout)
	defer cancel()

	command := exec.CommandContext(ctx, request.GoExecutable, argv...) // #nosec G204 -- fixed manager-owned vector against a verified launcher
	command.Dir = request.Directory
	command.Env = append([]string(nil), request.Environment...)
	command.Stdin = nil
	command.SysProcAttr = compilerSysProcAttr()
	budget := &outputBudget{remaining: limits.OutputBytes}
	stdout := &boundedBuffer{budget: budget}
	stderr := &boundedBuffer{budget: budget}
	command.Stdout = stdout
	command.Stderr = stderr

	runErr := command.Run()
	result := &workerResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), Started: session.started}
	if ctx.Err() != nil {
		result.TimedOut = true
	}
	if errors.Is(stdout.err, errOutputLimit) || errors.Is(stderr.err, errOutputLimit) || errors.Is(budget.err, errOutputLimit) {
		result.Overflow = true
	}
	if runErr != nil {
		result.Detail = runErr.Error()
		if state := command.ProcessState; state != nil {
			result.ExitCode = state.ExitCode()
		} else {
			// Process creation failed, so no child ever ran: there is no exit
			// status to read and the phase carries no output to interpret.
			result.StartFailed = true
			result.ExitCode = -1
		}
	}
	return result, nil
}

// verifyGoLauncher re-proves the identity of the only program the worker
// starts, immediately before every exec. The exact content identity of the
// tree is carried by the parent's curator-go-toolchain-v1 fingerprint, which is
// re-verified after every child and before publication.
func verifyGoLauncher(path string) error {
	if !filepath.IsAbs(path) {
		return diagnostic(CodeWorkerIdentityInvalid, "the Go launcher path is not absolute")
	}
	// The parent selected this path with physicalPath, so the worker re-proves
	// it with the same resolver: any other one would decide canonicality by a
	// different rule than the one the path was chosen under.
	resolved, err := physicalPath(filepath.Clean(path))
	if err != nil || resolved != path {
		return diagnosticErr(CodeWorkerIdentityInvalid, err, "the Go launcher path is not canonical and link-free")
	}
	if err := validateLauncher(path); err != nil {
		return diagnosticErr(CodeWorkerIdentityInvalid, err, "the Go launcher is not the fingerprinted native executable")
	}
	return nil
}
