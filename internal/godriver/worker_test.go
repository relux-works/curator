package godriver

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// rawWorker drives a real hidden-mode worker process over the session channel
// with complete control over the framing, so every identity and protocol
// rejection is proved against the shipped worker rather than a mock.
type rawWorker struct {
	t       *testing.T
	command *exec.Cmd
	domain  *controlDomain
	stdin   io.WriteCloser
	stdout  io.ReadCloser
}

// startRawWorker launches the worker through the same manager-owned control
// domain the production parent installs, so every raw session observes the real
// controls rather than an unconstrained process.
func startRawWorker(t *testing.T, identity ExecutableIdentity, limits ResourceLimits, probes []ControlProbe) *rawWorker {
	t.Helper()
	return startRawWorkerFrom(t, identity.Path, limits, probes)
}

// startRawWorkerFrom launches the worker from launchPath, which is the physical
// executable for every ordinary test and a package-manager link naming that
// same installed file when a test reproduces an operator launch shape. The
// worker resolves its own identity from whatever path it was started through.
func startRawWorkerFrom(t *testing.T, launchPath string, limits ResourceLimits, probes []ControlProbe) *rawWorker {
	t.Helper()
	command := exec.Command(launchPath, WorkerMode) // #nosec G204 -- identity-verified re-execution of the test binary
	command.Env = workerEnvironment()
	command.SysProcAttr = workerSysProcAttr()
	stdin, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	command.Stderr = os.Stderr
	domain, err := prepareControlDomain(limits, probes)
	if err != nil {
		t.Fatal(err)
	}
	if err := domain.launch(command); err != nil {
		domain.close()
		t.Fatal(err)
	}
	worker := &rawWorker{t: t, command: command, stdin: stdin, stdout: stdout, domain: domain}
	t.Cleanup(worker.close)
	return worker
}

func (worker *rawWorker) send(message workerMessage) {
	worker.t.Helper()
	if err := writeMessage(worker.stdin, message); err != nil {
		worker.t.Fatalf("cannot send %s: %v", message.Kind, err)
	}
}

// sendFrame writes a raw frame so a test can present an oversize declared
// length or a non-message payload.
func (worker *rawWorker) sendFrame(length uint32, payload []byte) {
	worker.t.Helper()
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], length)
	if _, err := worker.stdin.Write(header[:]); err != nil {
		worker.t.Fatal(err)
	}
	if len(payload) != 0 {
		_, _ = worker.stdin.Write(payload)
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

// expectFailure requires the worker to reject with the given stable diagnostic.
// The mandatory-control rejection is never one of them: that boundary belongs
// exclusively to the manager parent, before the worker is released to execute.
func (worker *rawWorker) expectFailure(code string) {
	worker.t.Helper()
	if code == CodeControlUnavailable {
		worker.t.Fatalf("no worker-side rejection may be %s", CodeControlUnavailable)
	}
	message := worker.receive()
	if message.Kind != kindFailure || message.Failure == nil {
		worker.t.Fatalf("worker sent %q, want a failure", message.Kind)
	}
	if message.Failure.Code == CodeControlUnavailable {
		worker.t.Fatalf("the worker raised the pre-worker mandatory-control rejection: %s", message.Failure.Detail)
	}
	if message.Failure.Code != code {
		worker.t.Fatalf("worker failure code = %q (%s), want %q", message.Failure.Code, message.Failure.Detail, code)
	}
}

func (worker *rawWorker) close() {
	_ = worker.stdin.Close()
	terminateWorkerDomain(worker.command)
	_ = worker.command.Wait()
	worker.domain.close()
}

// workerScenario is one live fixture plus the exact valid request the manager
// parent would send, so a test can mutate exactly one field.
type workerScenario struct {
	fixture  *workerFixture
	identity ExecutableIdentity
	request  workerRequest
	nonce    string
	stage    string
	limits   ResourceLimits
	probes   []ControlProbe

	// launchPath is the path the worker is started from. Empty means the
	// physical executable, which is what the production parent uses.
	launchPath string
}

func newWorkerScenario(t *testing.T) *workerScenario {
	t.Helper()
	fixture := newSnapshotFixture(t)
	fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})

	identity, err := resolveExecutableIdentity(managerExecutable())
	if err != nil {
		t.Fatal(err)
	}
	limits, err := normalizeBuildLimits(ResourceLimits{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	platform, probes, err := probeNativeControls(limits)
	if err != nil {
		t.Fatal(err)
	}
	stage, err := os.MkdirTemp(fixture.session.operation, ".curator-go-build-")
	if err != nil {
		t.Fatal(err)
	}
	artifact := filepath.Join(stage, "bin", "golden-tool")
	if err := os.MkdirAll(filepath.Dir(artifact), 0o700); err != nil {
		t.Fatal(err)
	}
	target := fixture.session.Target()
	nonce, err := sessionToken()
	if err != nil {
		t.Fatal(err)
	}
	secret, err := sessionToken()
	if err != nil {
		t.Fatal(err)
	}
	// The domain is installed from this operation's own probe set. The request
	// carries an independent copy so a test can tamper with what the worker is
	// told without changing what the manager actually installed.
	return &workerScenario{
		fixture: fixture, identity: identity, nonce: nonce, stage: stage, limits: limits,
		probes: append([]ControlProbe(nil), probes...),
		request: workerRequest{
			Version:          protocolVersion,
			Secret:           secret,
			ExecutablePath:   identity.Path,
			ExecutableSHA256: identity.SHA256,
			ExecutableSize:   identity.Size,
			GoExecutable:     fixture.session.Executable(),
			GOROOT:           fixture.session.GOROOT(),
			ToolDirectory:    filepath.Join(fixture.session.GOROOT(), "pkg", "tool", target.GOOS+"_"+target.GOARCH),
			Directory:        fixture.sourceDir,
			Environment:      fixture.session.Environment(),
			ListArgv:         append([]string(nil), listArguments...),
			BuildArgv:        append(append([]string(nil), buildArgumentPrefix...), artifact, "."),
			ArtifactPath:     artifact,
			ReadOnlyRoots:    []string{fixture.root, fixture.session.GOROOT()},
			PrivateRoots:     privateRoots(fixture.session.Environment(), stage),
			Platform:         platform,
			Probes:           probes,
			Limits: wireLimits{
				TimeoutMillis: limits.Timeout.Milliseconds(), OutputBytes: limits.OutputBytes,
				ArtifactBytes: limits.ArtifactBytes, FileBytes: limits.FileBytes,
				DiskBytes: limits.DiskBytes, MemoryBytes: limits.MemoryBytes, Processes: limits.Processes,
			},
		},
	}
}

// start launches a worker inside the real control domain and sends the
// (possibly mutated) request.
func (scenario *workerScenario) start() *rawWorker {
	launchPath := scenario.launchPath
	if launchPath == "" {
		launchPath = scenario.identity.Path
	}
	worker := startRawWorkerFrom(scenario.fixture.t, launchPath, scenario.limits, scenario.probes)
	request := scenario.request
	worker.send(workerMessage{Kind: kindRequest, Nonce: scenario.nonce, Request: &request})
	return worker
}

// requireReady drives the session to the acknowledged state.
func (scenario *workerScenario) requireReady() *rawWorker {
	scenario.fixture.t.Helper()
	worker := scenario.start()
	message := worker.receive()
	if message.Kind != kindReady || message.Ready == nil {
		scenario.fixture.t.Fatalf("worker sent %q, want ready", message.Kind)
	}
	return worker
}

func (scenario *workerScenario) requireNoCompilerStarted() {
	scenario.fixture.t.Helper()
	if calls := scenario.fixture.sourceAwareCalls(); len(calls) != 0 {
		scenario.fixture.t.Fatalf("compiler calls = %+v, want none", calls)
	}
}

func TestWorkerRejectsEveryIdentityAndRequestMutation(t *testing.T) {
	for _, testCase := range []struct {
		name, code string
		mutate     func(*workerScenario)
	}{
		{name: "executable digest mismatch", code: CodeWorkerIdentityInvalid, mutate: func(scenario *workerScenario) {
			scenario.request.ExecutableSHA256 = "sha256:" + string(make([]byte, 0)) + "0000000000000000000000000000000000000000000000000000000000000000"
		}},
		{name: "executable path substitution", code: CodeWorkerIdentityInvalid, mutate: func(scenario *workerScenario) {
			scenario.request.ExecutablePath = filepath.Join(filepath.Dir(scenario.identity.Path), "substituted")
		}},
		{name: "executable size mismatch", code: CodeWorkerIdentityInvalid, mutate: func(scenario *workerScenario) {
			scenario.request.ExecutableSize++
		}},
		{name: "unexpected program below the worker", code: CodeWorkerIdentityInvalid, mutate: func(scenario *workerScenario) {
			scenario.request.GoExecutable = scenario.identity.Path
		}},
		{name: "tool directory outside GOROOT", code: CodeWorkerIdentityInvalid, mutate: func(scenario *workerScenario) {
			scenario.request.ToolDirectory = filepath.Join(scenario.fixture.root, "pkg", "tool")
		}},
		{name: "mutated list vector", code: CodeWorkerProtocolInvalid, mutate: func(scenario *workerScenario) {
			scenario.request.ListArgv = []string{"run", "."}
		}},
		{name: "mutated build vector", code: CodeWorkerProtocolInvalid, mutate: func(scenario *workerScenario) {
			scenario.request.BuildArgv = append(append([]string(nil), buildArgumentPrefix...), scenario.request.ArtifactPath, "./...")
		}},
		{name: "output outside private staging", code: CodeWorkerProtocolInvalid, mutate: func(scenario *workerScenario) {
			escaped := filepath.Join(scenario.fixture.root, "escaped")
			scenario.request.ArtifactPath = escaped
			scenario.request.BuildArgv = append(append([]string(nil), buildArgumentPrefix...), escaped, ".")
		}},
		{name: "poisoned environment", code: CodeWorkerProtocolInvalid, mutate: func(scenario *workerScenario) {
			scenario.request.Environment = append(scenario.request.Environment, "GOAUTH=netrc")
		}},
		{name: "widened environment", code: CodeWorkerProtocolInvalid, mutate: func(scenario *workerScenario) {
			environment := environmentMap(scenario.request.Environment)
			environment["GOPROXY"] = "https://proxy.example"
			scenario.request.Environment = environmentSlice(environment)
		}},
		{name: "working directory outside the snapshot", code: CodeWorkerProtocolInvalid, mutate: func(scenario *workerScenario) {
			scenario.request.Directory = scenario.stage
		}},
		{name: "private root overlapping the snapshot", code: CodeWorkerProtocolInvalid, mutate: func(scenario *workerScenario) {
			scenario.request.PrivateRoots = append(scenario.request.PrivateRoots, scenario.fixture.sourceDir)
		}},
		{name: "unsupported protocol version", code: CodeWorkerProtocolInvalid, mutate: func(scenario *workerScenario) {
			scenario.request.Version = "curator-go-worker-v2"
		}},
		{name: "probe claims an unavailable control", code: CodeCapabilityEvidenceInvalid, mutate: func(scenario *workerScenario) {
			for index := range scenario.request.Probes {
				if scenario.request.Probes[index].Availability == AvailabilityUnavailable {
					scenario.request.Probes[index].Availability = AvailabilityAvailable
					return
				}
			}
		}},
		{name: "probe set outside the inventory", code: CodeCapabilityEvidenceInvalid, mutate: func(scenario *workerScenario) {
			scenario.request.Probes = append(scenario.request.Probes, ControlProbe{
				Name: "host-firewall-profile", Availability: AvailabilityAvailable, ProbedAt: ProbeTiming,
			})
		}},
		{name: "platform other than this host", code: CodeCapabilityEvidenceInvalid, mutate: func(scenario *workerScenario) {
			if scenario.request.Platform == PlatformMacOS {
				scenario.request.Platform = PlatformWindows
			} else {
				scenario.request.Platform = PlatformMacOS
			}
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			scenario := newWorkerScenario(t)
			testCase.mutate(scenario)
			worker := scenario.start()
			worker.expectFailure(testCase.code)
			scenario.requireNoCompilerStarted()
		})
	}
}

func TestWorkerRejectsEveryProtocolDeviationBeforeTheCompiler(t *testing.T) {
	t.Run("build permit before list validation", func(t *testing.T) {
		scenario := newWorkerScenario(t)
		worker := scenario.requireReady()
		worker.send(workerMessage{
			Kind: kindPermit, Nonce: scenario.nonce,
			Permit: buildPermit(scenario.request.Secret, scenario.nonce, scenario.request.BuildArgv),
		})
		worker.expectFailure(CodeWorkerProtocolInvalid)
		scenario.requireNoCompilerStarted()
	})

	t.Run("replayed session nonce", func(t *testing.T) {
		scenario := newWorkerScenario(t)
		worker := scenario.requireReady()
		other, err := sessionToken()
		if err != nil {
			t.Fatal(err)
		}
		worker.send(workerMessage{Kind: kindList, Nonce: other})
		worker.expectFailure(CodeWorkerProtocolInvalid)
		scenario.requireNoCompilerStarted()
	})

	t.Run("out of order message", func(t *testing.T) {
		scenario := newWorkerScenario(t)
		worker := scenario.requireReady()
		worker.send(workerMessage{Kind: kindShutdown, Nonce: scenario.nonce})
		worker.expectFailure(CodeWorkerProtocolInvalid)
		scenario.requireNoCompilerStarted()
	})

	t.Run("unknown message kind", func(t *testing.T) {
		scenario := newWorkerScenario(t)
		worker := scenario.requireReady()
		payload, err := json.Marshal(map[string]string{"kind": "run", "nonce": scenario.nonce})
		if err != nil {
			t.Fatal(err)
		}
		worker.sendFrame(uint32(len(payload)), payload)
		worker.expectFailure(CodeWorkerProtocolInvalid)
		scenario.requireNoCompilerStarted()
	})

	t.Run("oversize frame", func(t *testing.T) {
		scenario := newWorkerScenario(t)
		worker := scenario.requireReady()
		worker.sendFrame(uint32(maxProtocolFrame)+1, nil)
		worker.expectFailure(CodeWorkerProtocolInvalid)
		scenario.requireNoCompilerStarted()
	})

	t.Run("unauthenticated build permit", func(t *testing.T) {
		scenario := newWorkerScenario(t)
		worker := scenario.requireReady()
		worker.send(workerMessage{Kind: kindList, Nonce: scenario.nonce})
		if message := worker.receive(); message.Kind != kindListResult {
			t.Fatalf("worker sent %q, want a list result", message.Kind)
		}
		worker.send(workerMessage{Kind: kindPermit, Nonce: scenario.nonce, Permit: "0000"})
		worker.expectFailure(CodeWorkerProtocolInvalid)
		if calls := scenario.fixture.sourceAwareCalls(); len(calls) != 1 {
			t.Fatalf("calls = %+v, the compiler must not build without an authenticated permit", calls)
		}
	})

	t.Run("permit bound to another vector", func(t *testing.T) {
		scenario := newWorkerScenario(t)
		worker := scenario.requireReady()
		worker.send(workerMessage{Kind: kindList, Nonce: scenario.nonce})
		worker.receive()
		worker.send(workerMessage{
			Kind: kindPermit, Nonce: scenario.nonce,
			Permit: buildPermit(scenario.request.Secret, scenario.nonce, []string{"build", "."}),
		})
		worker.expectFailure(CodeWorkerProtocolInvalid)
		if calls := scenario.fixture.sourceAwareCalls(); len(calls) != 1 {
			t.Fatalf("calls = %+v", calls)
		}
	})

	t.Run("second build request in one session", func(t *testing.T) {
		scenario := newWorkerScenario(t)
		worker := scenario.requireReady()
		worker.send(workerMessage{Kind: kindList, Nonce: scenario.nonce})
		worker.receive()
		permit := buildPermit(scenario.request.Secret, scenario.nonce, scenario.request.BuildArgv)
		worker.send(workerMessage{Kind: kindPermit, Nonce: scenario.nonce, Permit: permit})
		if message := worker.receive(); message.Kind != kindBuildResult {
			t.Fatalf("worker sent %q, want a build result", message.Kind)
		}
		worker.send(workerMessage{Kind: kindPermit, Nonce: scenario.nonce, Permit: permit})
		worker.expectFailure(CodeWorkerProtocolInvalid)
		if calls := scenario.fixture.sourceAwareCalls(); len(calls) != 2 {
			t.Fatalf("calls = %+v, exactly one list and one build may run in a session", calls)
		}
	})

	t.Run("second list in one session", func(t *testing.T) {
		scenario := newWorkerScenario(t)
		worker := scenario.requireReady()
		worker.send(workerMessage{Kind: kindList, Nonce: scenario.nonce})
		worker.receive()
		worker.send(workerMessage{Kind: kindList, Nonce: scenario.nonce})
		worker.expectFailure(CodeWorkerProtocolInvalid)
		if calls := scenario.fixture.sourceAwareCalls(); len(calls) != 1 {
			t.Fatalf("calls = %+v", calls)
		}
	})
}

func TestWorkerRunsTheCompleteHappySession(t *testing.T) {
	scenario := newWorkerScenario(t)
	worker := scenario.requireReady()
	worker.send(workerMessage{Kind: kindList, Nonce: scenario.nonce})
	listResult := worker.receive()
	if listResult.Kind != kindListResult || listResult.Result == nil || listResult.Result.ExitCode != 0 {
		t.Fatalf("list result = %+v", listResult)
	}
	worker.send(workerMessage{
		Kind: kindPermit, Nonce: scenario.nonce,
		Permit: buildPermit(scenario.request.Secret, scenario.nonce, scenario.request.BuildArgv),
	})
	buildResult := worker.receive()
	if buildResult.Kind != kindBuildResult || buildResult.Result == nil || buildResult.Result.ExitCode != 0 {
		t.Fatalf("build result = %+v", buildResult)
	}
	if _, err := os.Lstat(scenario.request.ArtifactPath); err != nil {
		t.Fatalf("worker did not stage the manager-derived artifact: %v", err)
	}
	worker.send(workerMessage{Kind: kindShutdown, Nonce: scenario.nonce})
	if err := worker.command.Wait(); err != nil {
		t.Fatalf("worker exited with %v", err)
	}
}

func TestProbeRejectsAnUncoveredPlatformBeforeTheWorker(t *testing.T) {
	limits, err := normalizeBuildLimits(ResourceLimits{})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = probeNativeControlsFor("", limits)
	if DiagnosticCode(err) != CodeControlUnavailable {
		t.Fatalf("error = %v, want %s", err, CodeControlUnavailable)
	}
	_, _, err = probeNativeControlsFor("linux", limits)
	if DiagnosticCode(err) != CodeControlUnavailable {
		t.Fatalf("error = %v, want %s", err, CodeControlUnavailable)
	}
}

func TestWorkerModeIsNotReachableThroughPackageData(t *testing.T) {
	// The hidden mode is selected only by the manager parent's own fixed argv.
	// A package cannot present it: the build-command surface is closed and the
	// worker executable is this manager, never a package path.
	fixture := newSnapshotFixture(t)
	fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
	request := fixture.request(ResourceLimits{Timeout: 30 * time.Second})
	request.CommandObject["worker_mode"] = WorkerMode
	if _, err := Build(context.Background(), request); DiagnosticCode(err) != CodePackageInfluenceForbidden {
		t.Fatalf("error = %v, want %s", err, CodePackageInfluenceForbidden)
	}
}

// TestWorkerRejectsAProbeThatDropsAPlatformAvailableControl proves the worker's
// defence in depth over a tampered probe record. The normative availability
// decision already happened in the parent before this process was released, so
// the worker reports an evidence fault and never the mandatory-control
// rejection, which is reserved for the pre-worker boundary.
func TestWorkerRejectsAProbeThatDropsAPlatformAvailableControl(t *testing.T) {
	scenario := newWorkerScenario(t)
	for index := range scenario.request.Probes {
		if scenario.request.Probes[index].Availability == AvailabilityAvailable {
			scenario.request.Probes[index].Availability = AvailabilityUnavailable
			break
		}
	}
	worker := scenario.start()
	worker.expectFailure(CodeCapabilityEvidenceInvalid)
	scenario.requireNoCompilerStarted()
}

// TestWorkerReturnsABoundedResultWhenTheGoChildCannotStart injects a real
// process-creation failure at the worker's single permitted creation site: the
// manager-owned working directory is inside the frozen snapshot but does not
// exist, so the operating system refuses the child and leaves no process state.
// The worker must report a bounded result rather than dereference it.
func TestWorkerReturnsABoundedResultWhenTheGoChildCannotStart(t *testing.T) {
	scenario := newWorkerScenario(t)
	scenario.request.Directory = filepath.Join(scenario.fixture.sourceDir, "absent")
	worker := scenario.requireReady()
	worker.send(workerMessage{Kind: kindList, Nonce: scenario.nonce})

	message := worker.receive()
	if message.Kind != kindListResult || message.Result == nil {
		t.Fatalf("worker sent %+v, want a bounded list result", message)
	}
	result := message.Result
	if !result.StartFailed || result.ExitCode != -1 || result.Detail == "" {
		t.Fatalf("result = %+v, want a refused process creation", result)
	}
	if result.Started != 1 {
		t.Fatalf("started = %d, a refused creation still consumes the single permitted site", result.Started)
	}
	if len(result.Stdout) != 0 || len(result.Stderr) != 0 {
		t.Fatalf("result carries output from a child that never ran: %+v", result)
	}
	// The worker survived: no panic, and the session still tears down cleanly.
	worker.send(workerMessage{Kind: kindPermit, Nonce: scenario.nonce,
		Permit: buildPermit(scenario.request.Secret, scenario.nonce, scenario.request.BuildArgv)})
	build := worker.receive()
	if build.Kind != kindBuildResult || build.Result == nil || !build.Result.StartFailed {
		t.Fatalf("build result = %+v, want the same bounded refusal", build)
	}
	if _, err := os.Lstat(scenario.request.ArtifactPath); err == nil {
		t.Fatal("a refused build produced an artifact")
	}
	worker.send(workerMessage{Kind: kindShutdown, Nonce: scenario.nonce})
	if err := worker.command.Wait(); err != nil {
		t.Fatalf("worker exited with %v, want a clean shutdown after a refused child", err)
	}
	if calls := scenario.fixture.sourceAwareCalls(); len(calls) != 0 {
		t.Fatalf("calls = %+v, no Go child may have run", calls)
	}
}
