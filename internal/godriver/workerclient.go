package godriver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"reflect"
	"time"
)

const (
	sessionNonceLength  = 64
	sessionSecretLength = 64

	// workerStderrLimit bounds the worker's own diagnostic stream.
	workerStderrLimit = int64(64 * 1024)
	// workerShutdownGrace bounds the join after the session completes.
	workerShutdownGrace = 5 * time.Second
)

// workerPlan is the complete manager-owned launch input. Nothing here is
// derived from package bytes.
type workerPlan struct {
	Executable    ExecutableIdentity
	GoExecutable  string
	GOROOT        string
	ToolDirectory string
	Directory     string
	Environment   []string
	ListArgv      []string
	BuildArgv     []string
	ArtifactPath  string
	ReadOnlyRoots []string
	PrivateRoots  []string
	Platform      string
	Probes        []ControlProbe
	Limits        ResourceLimits
}

// workerClient is one parent-side worker session. It performs exactly one
// list, one validation gap, one authenticated permit, and one build.
type workerClient struct {
	command *exec.Cmd
	domain  *controlDomain
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  *boundedBuffer

	plan     workerPlan
	nonce    string
	secret   string
	evidence CapabilityEvidence

	cancel   context.CancelFunc
	finished bool
	drained  bool
}

// launchWorker performs steps 3 to 6 of the worker session: it re-verifies the
// executable identity at the launch boundary, creates and installs the
// manager-owned native control domain, starts exactly that executable in the
// fixed hidden mode over anonymous inherited pipes, sends one canonical bounded
// request, and requires an identity proof plus a capability-evidence record
// identical to the one the manager derived from what it installed.
//
// Every control-availability decision and every control installation happens
// here, before the worker is released to execute, so CodeControlUnavailable
// cannot first appear once the worker is running.
func launchWorker(ctx context.Context, plan workerPlan) (_ *workerClient, resultErr error) {
	nonce, err := sessionToken()
	if err != nil {
		return nil, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot derive a session nonce")
	}
	secret, err := sessionToken()
	if err != nil {
		return nil, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot derive a session secret")
	}
	// Re-verify immediately before launch so a replacement race between the
	// first check and exec cannot widen the process graph.
	if err := plan.Executable.Verify(); err != nil {
		return nil, err
	}

	// The worker bounds each Go child at Limits.Timeout. The parent bounds the
	// complete worker domain at the same value plus one bounded join grace and
	// terminates the whole domain when it expires, so the operation cannot
	// outlive its wall-clock budget even if the worker itself hangs.
	runCtx, cancel := context.WithTimeout(ctx, plan.Limits.Timeout+workerShutdownGrace)
	client := &workerClient{plan: plan, nonce: nonce, secret: secret, cancel: cancel}
	defer func() {
		if resultErr != nil {
			client.teardown()
		}
	}()

	// Create every mechanism that can exist before a process does. A control
	// that cannot be created rejects here, with no worker in existence.
	domain, err := prepareControlDomain(plan.Limits, plan.Probes)
	if err != nil {
		return nil, err
	}
	client.domain = domain

	command := exec.CommandContext(runCtx, plan.Executable.Path, WorkerMode) // #nosec G204 -- identity-verified re-execution of this manager
	command.Dir = filepath.Dir(plan.Executable.Path)
	command.Env = workerEnvironment()
	command.SysProcAttr = workerSysProcAttr()
	command.Cancel = func() error {
		terminateWorkerDomain(command)
		return nil
	}
	command.WaitDelay = workerShutdownGrace
	client.stderr = &boundedBuffer{budget: &outputBudget{remaining: workerStderrLimit}}
	command.Stderr = client.stderr
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot open the worker request channel")
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot open the worker response channel")
	}
	client.command, client.stdin, client.stdout = command, stdin, stdout
	// The domain owns the launch: it installs the remaining mechanisms around
	// process creation and releases the worker only when every one of them is in
	// place. A failure destroys the worker before it executes a session.
	if err := domain.launch(command); err != nil {
		return nil, client.destroyBeforeExecution(err)
	}

	// Only now, with the domain complete, may an applied status exist.
	evidence := evidenceFromApplied(plan.Platform, plan.Probes, domain.installedControls())
	if err := validateCapabilityEvidence(evidence, plan.Platform, plan.Probes); err != nil {
		return nil, err
	}

	request := workerRequest{
		Version:          protocolVersion,
		Secret:           secret,
		ExecutablePath:   plan.Executable.Path,
		ExecutableSHA256: plan.Executable.SHA256,
		ExecutableSize:   plan.Executable.Size,
		GoExecutable:     plan.GoExecutable,
		GOROOT:           plan.GOROOT,
		ToolDirectory:    plan.ToolDirectory,
		Directory:        plan.Directory,
		Environment:      plan.Environment,
		ListArgv:         plan.ListArgv,
		BuildArgv:        plan.BuildArgv,
		ArtifactPath:     plan.ArtifactPath,
		ReadOnlyRoots:    plan.ReadOnlyRoots,
		PrivateRoots:     plan.PrivateRoots,
		Platform:         plan.Platform,
		Probes:           plan.Probes,
		Limits: wireLimits{
			TimeoutMillis: plan.Limits.Timeout.Milliseconds(),
			OutputBytes:   plan.Limits.OutputBytes,
			ArtifactBytes: plan.Limits.ArtifactBytes,
			FileBytes:     plan.Limits.FileBytes,
			DiskBytes:     plan.Limits.DiskBytes,
			MemoryBytes:   plan.Limits.MemoryBytes,
			Processes:     plan.Limits.Processes,
		},
	}
	if err := client.send(workerMessage{Kind: kindRequest, Nonce: nonce, Request: &request}); err != nil {
		return nil, err
	}
	ready, err := client.receive(kindReady)
	if err != nil {
		return nil, err
	}
	if ready.Ready == nil {
		return nil, diagnostic(CodeWorkerProtocolInvalid, "worker acknowledgement carries no identity proof")
	}
	if err := plan.Executable.matches(ready.Ready.ExecutablePath, ready.Ready.ExecutableSHA256, ready.Ready.ExecutableSize); err != nil {
		return nil, err
	}
	// The worker derives the same closed record independently from what it can
	// confirm is in effect. A difference means a control the manager installed is
	// not really governing the worker, which is an evidence fault rather than a
	// control-availability decision.
	if !reflect.DeepEqual(ready.Ready.Evidence, evidence) {
		return nil, diagnostic(CodeCapabilityEvidenceInvalid,
			"the worker capability-evidence record differs from the record the manager installed")
	}
	if err := matchAppliedControls(ready.Ready.Applied, evidence); err != nil {
		return nil, err
	}
	client.evidence = evidence
	return client, nil
}

// destroyBeforeExecution converts a failed control installation into the stable
// pre-worker rejection. It proves at runtime that the worker produced no session
// output, so the failure boundary is verified rather than assumed.
func (client *workerClient) destroyBeforeExecution(cause error) error {
	produced := client.probeForSessionOutput()
	client.teardown()
	if produced {
		return diagnosticErr(CodeControlUnavailable, cause,
			"the worker produced session output before its native control domain was installed")
	}
	if DiagnosticCode(cause) == "" {
		return diagnosticErr(CodeControlUnavailable, cause, "cannot install the manager-owned native control domain")
	}
	return cause
}

// probeForSessionOutput reports whether the worker wrote anything on the session
// channel. A worker destroyed before its controls were installed must not have.
func (client *workerClient) probeForSessionOutput() bool {
	if client.command == nil || client.command.Process == nil || client.stdout == nil {
		// The worker was never created, so it cannot have produced output.
		return false
	}
	client.drained = true
	produced := make(chan bool, 1)
	go func() {
		var probe [1]byte
		count, _ := io.ReadFull(client.stdout, probe[:])
		produced <- count > 0
	}()
	select {
	case value := <-produced:
		return value
	case <-time.After(workerShutdownGrace):
		return false
	}
}

// matchAppliedControls cross-checks the worker's applied-control report against
// the record it emitted, so a report and its evidence cannot drift.
func matchAppliedControls(applied []string, evidence CapabilityEvidence) error {
	reported := make(map[string]bool, len(applied))
	for _, name := range applied {
		if reported[name] {
			return diagnostic(CodeCapabilityEvidenceInvalid, "worker reports control %q applied twice", name)
		}
		reported[name] = true
	}
	for _, entry := range evidence.Controls {
		if (entry.Status == StatusApplied) != reported[entry.Name] {
			return diagnostic(CodeCapabilityEvidenceInvalid,
				"worker applied-control report contradicts evidence for %q", entry.Name)
		}
		delete(reported, entry.Name)
	}
	for name := range reported {
		return diagnostic(CodeCapabilityEvidenceInvalid, "worker applied control %q outside its evidence record", name)
	}
	return nil
}

// list runs the single fixed go list vector inside the session.
func (client *workerClient) list() (Output, error) {
	if err := client.send(workerMessage{Kind: kindList, Nonce: client.nonce}); err != nil {
		return Output{}, err
	}
	return client.result(kindListResult, 1)
}

// build sends the single authenticated fixed build permit and runs the single
// fixed go build vector. It is reachable only after the parent validated the
// complete package graph.
func (client *workerClient) build() (Output, error) {
	permit := buildPermit(client.secret, client.nonce, client.plan.BuildArgv)
	if err := client.send(workerMessage{Kind: kindPermit, Nonce: client.nonce, Permit: permit}); err != nil {
		return Output{}, err
	}
	return client.result(kindBuildResult, 2)
}

// result reads one bounded child result and requires that the worker has used
// its single permitted process-creation site exactly the expected number of
// times in this session, so an extra program below the worker is detected
// rather than assumed impossible.
func (client *workerClient) result(kind string, started int) (Output, error) {
	message, err := client.receive(kind)
	if err != nil {
		return Output{}, err
	}
	if message.Result == nil {
		return Output{}, diagnostic(CodeWorkerProtocolInvalid, "worker %s carries no result", kind)
	}
	result := message.Result
	if result.Started != started {
		return Output{}, diagnostic(CodeWorkerIdentityInvalid,
			"worker started %d programs in this session, want exactly %d", result.Started, started)
	}
	output := Output{Stdout: result.Stdout, Stderr: result.Stderr}
	switch {
	case result.TimedOut:
		return output, diagnostic("process_timeout", "Go child exceeded its deadline")
	case result.Overflow:
		return output, diagnostic("process_output_limit", "Go child exceeded its combined output bound")
	case result.StartFailed:
		// The operating system refused to create the Go child, so there is no
		// exit status. The phase failed; the caller classifies it as the fixed
		// command it was and publishes nothing.
		return output, errors.New("the fixed Go command could not be started: " + result.Detail)
	case result.ExitCode != 0:
		return output, errors.New(result.Detail)
	}
	if int64(len(output.Stdout)+len(output.Stderr)) > client.plan.Limits.OutputBytes {
		return output, diagnostic("process_output_limit", "Go child exceeded its combined output bound")
	}
	return output, nil
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
		return workerMessage{}, diagnostic(message.Failure.Code, "%s", message.Failure.Detail)
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
// session channel. It is safe to call more than once. The domain is terminated
// before the worker is reaped, so a group kill can never reach a reused
// process identifier.
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
			// Drain to end of file first: the worker exit closes the response pipe
			// without reaping the process, so the process-group identifier is still
			// reserved when the domain is terminated below.
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
		terminateWorkerDomain(client.command)
		_ = client.command.Wait()
	}
	if client.stdout != nil {
		_ = client.stdout.Close()
	}
	// Releasing the domain last terminates anything still inside it, so no
	// compiler or tool child can outlive the operation.
	client.domain.close()
	if client.cancel != nil {
		client.cancel()
	}
}

// Evidence returns the closed capability-evidence-v1 record for this session.
func (client *workerClient) Evidence() CapabilityEvidence { return client.evidence }

func sessionToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

// workerEnvironment is the fixed bootstrap environment of the worker process
// itself. It carries only indispensable operating-system process variables.
func workerEnvironment() []string {
	values := indispensableEnvironment()
	if values == nil {
		values = map[string]string{}
	}
	return environmentSlice(values)
}
