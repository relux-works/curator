package godriver

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// injectControlSeamFault installs a real failure at exactly one seam of the
// manager-owned control domain for the duration of one test.
func injectControlSeamFault(t *testing.T, target controlSeam) *int {
	t.Helper()
	if controlSeamFault != nil {
		t.Fatal("a control seam fault is already installed")
	}
	fired := 0
	controlSeamFault = func(seam controlSeam) error {
		if seam != target {
			return nil
		}
		fired++
		return errors.New("injected " + string(seam) + " failure")
	}
	t.Cleanup(func() { controlSeamFault = nil })
	return &fired
}

// requireNoPrivateArtifact proves the operation published nothing: no staged
// build output survives anywhere below the operation-private root.
func requireNoPrivateArtifact(t *testing.T, session *Session) {
	t.Helper()
	err := filepath.WalkDir(session.operation, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && filepath.Base(filepath.Dir(path)) == "bin" {
			return errors.New("operation-private state carries staged output " + path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("publication check failed: %v", err)
	}
}

// TestEveryControlSeamFailureRejectsBeforeTheWorkerExecutes injects a real
// failure at each seam of the manager-owned control domain — the per-operation
// applicability probe, the creation of the domain, and its installation around
// the worker launch — and requires the exact rc.5 failure boundary at every one
// of them: build_execution_control_unavailable, no worker session, no compiler,
// no evidence claiming an applied control, and nothing published.
func TestEveryControlSeamFailureRejectsBeforeTheWorkerExecutes(t *testing.T) {
	for _, seam := range []controlSeam{seamProbe, seamPrepare, seamInstall} {
		t.Run(string(seam), func(t *testing.T) {
			fixture := newSnapshotFixture(t)
			fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
			before := len(fixture.calls())

			fired := injectControlSeamFault(t, seam)
			result, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 30 * time.Second}))

			if *fired == 0 {
				t.Fatalf("the %s seam was never crossed", seam)
			}
			if DiagnosticCode(err) != CodeControlUnavailable {
				t.Fatalf("error = %v, want %s", err, CodeControlUnavailable)
			}
			// The manager verifies at runtime that a worker destroyed before its
			// domain was installed produced no session output, and says so when
			// it did. It must not have.
			if strings.Contains(err.Error(), "produced session output") {
				t.Fatalf("the worker ran a session before its controls were installed: %v", err)
			}
			// worker_started=false and compiler_started=false: the manager
			// released no program after the failure, so the stub launcher
			// recorded nothing beyond the session-establishment probes.
			if after := len(fixture.calls()); after != before {
				t.Fatalf("launcher calls went from %d to %d; no program may run after a control failure", before, after)
			}
			if len(result.Evidence.Controls) != 0 || result.Evidence.RecordVersion != "" {
				t.Fatalf("a rejected operation emitted capability evidence: %+v", result.Evidence)
			}
			if result.Artifact.StagedPath != "" {
				t.Fatalf("a rejected operation returned an artifact: %+v", result.Artifact)
			}
			requireNoPrivateArtifact(t, fixture.session)
		})
	}
}

// TestControlSeamFailureNeverDriftsIntoAHardenedClaim proves the reworked
// failure boundary did not widen the portable profile: an injected control
// failure rejects, while an inventory control the platform marks unavailable and
// every deferred hardened guarantee still permit the build.
func TestControlSeamFailureNeverDriftsIntoAHardenedClaim(t *testing.T) {
	fixture := newSnapshotFixture(t)
	fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})

	fired := injectControlSeamFault(t, seamPrepare)
	if _, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 30 * time.Second})); DiagnosticCode(err) != CodeControlUnavailable {
		t.Fatalf("error = %v, want %s", err, CodeControlUnavailable)
	}
	if *fired == 0 {
		t.Fatal("the domain-preparation seam was never crossed")
	}
	controlSeamFault = nil

	result, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 30 * time.Second}))
	if err != nil {
		t.Fatalf("an unavailable inventory control or a deferred guarantee rejected the build: %v", err)
	}
	unavailable := 0
	for _, entry := range result.Evidence.Controls {
		if entry.Availability == AvailabilityUnavailable {
			unavailable++
		}
		if isDeferredHardenedGuarantee(entry.Name) {
			t.Fatalf("evidence claims deferred hardened guarantee %q", entry.Name)
		}
	}
	if unavailable == 0 {
		t.Fatal("this platform has no normatively unavailable inventory control to prove the boundary with")
	}
	if result.Evidence.ExecutionPolicy != ExecutionPolicy {
		t.Fatalf("evidence execution policy = %q", result.Evidence.ExecutionPolicy)
	}
}

// TestControlDomainInstallsBeforeTheWorkerIsReleased proves the ordering the
// failure boundary depends on: a domain reports no installed control until it
// has been installed, so an applied status cannot exist before the mechanism
// does, and the worker is released only afterwards.
func TestControlDomainInstallsBeforeTheWorkerIsReleased(t *testing.T) {
	limits, err := normalizeBuildLimits(ResourceLimits{})
	if err != nil {
		t.Fatal(err)
	}
	platform, probes, err := probeNativeControls(limits)
	if err != nil {
		t.Fatal(err)
	}
	domain, err := prepareControlDomain(limits, probes)
	if err != nil {
		t.Fatal(err)
	}
	defer domain.close()
	if names := domain.installedControls(); len(names) != 0 {
		t.Fatalf("a prepared but uninstalled domain reports %q as installed", names)
	}
	empty := evidenceFromApplied(platform, probes, domain.installedControls())
	for _, entry := range empty.Controls {
		if entry.Status == StatusApplied {
			t.Fatalf("control %q reports applied before its mechanism was installed", entry.Name)
		}
	}
	// An uninstalled domain can never produce a valid record, because every
	// platform-available control would be reported as not applied.
	if DiagnosticCode(validateCapabilityEvidence(empty, platform, probes)) != CodeCapabilityEvidenceInvalid {
		t.Fatal("an uninstalled domain produced an acceptable capability-evidence record")
	}

	identity, err := resolveExecutableIdentity(managerExecutable())
	if err != nil {
		t.Fatal(err)
	}
	worker := startRawWorker(t, identity, limits, probes)
	installed := worker.domain.installedControls()
	wanted := installableControls(probes)
	if strings.Join(installed, ",") != strings.Join(wanted, ",") {
		t.Fatalf("installed controls = %q, want every probed-available control %q", installed, wanted)
	}
}

// TestParentRejectsAWorkerRecordThatDiffersFromWhatWasInstalled proves the
// manager, not the worker, owns the applied status: a worker record that does
// not match the manager's installed set is an evidence fault.
func TestParentRejectsAWorkerRecordThatDiffersFromWhatWasInstalled(t *testing.T) {
	platform := PlatformWindows
	probes := syntheticProbes(platform)
	installed := installableControls(probes)
	manager := evidenceFromApplied(platform, probes, installed)

	if len(manager.Controls) != len(nativeControlInventory) {
		t.Fatalf("manager record has %d entries", len(manager.Controls))
	}
	drifted := evidenceFromApplied(platform, probes, installed[:len(installed)-1])
	if DiagnosticCode(validateCapabilityEvidence(drifted, platform, probes)) != CodeCapabilityEvidenceInvalid {
		t.Fatal("a worker record dropping an installed control was accepted")
	}
	if err := matchAppliedControls(installed[:len(installed)-1], manager); DiagnosticCode(err) != CodeCapabilityEvidenceInvalid {
		t.Fatalf("error = %v, want %s", err, CodeCapabilityEvidenceInvalid)
	}
}

// TestParentClassifiesAChildStartFailureWithoutBuildContinuation proves the
// parent's half of the refused-creation path: a bounded start-failure result is
// classified as the fixed command it was, carries no exit status, and stops the
// operation before the build phase.
func TestParentClassifiesAChildStartFailureWithoutBuildContinuation(t *testing.T) {
	for _, testCase := range []struct {
		name, kind, code string
		started          int
		classify         func(error, []byte) error
	}{
		{name: "list", kind: kindListResult, code: "go_list_failed", started: 1, classify: classifyListFailure},
		{name: "build", kind: kindBuildResult, code: "go_build_failed", started: 2, classify: classifyBuildFailure},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			nonce, err := sessionToken()
			if err != nil {
				t.Fatal(err)
			}
			frame := encodeWorkerFrame(t, workerMessage{
				Kind: testCase.kind, Nonce: nonce,
				Result: &workerResult{
					Started: testCase.started, ExitCode: -1, StartFailed: true,
					Detail: "fork/exec /goroot/bin/go: no such file or directory",
				},
			})
			client := &workerClient{
				nonce:  nonce,
				stdout: io.NopCloser(bytes.NewReader(frame)),
				plan:   workerPlan{Limits: ResourceLimits{OutputBytes: defaultBuildOutput}},
			}
			output, err := client.result(testCase.kind, testCase.started)
			if err == nil {
				t.Fatal("a refused process creation was reported as success")
			}
			if !strings.Contains(err.Error(), "could not be started") {
				t.Fatalf("error = %v, want the refused-creation detail", err)
			}
			classified := testCase.classify(err, output.Stderr)
			if DiagnosticCode(classified) != testCase.code {
				t.Fatalf("classified = %v, want %s", classified, testCase.code)
			}
		})
	}
}

func encodeWorkerFrame(t *testing.T, message workerMessage) []byte {
	t.Helper()
	payload, err := json.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
	return append(header[:], payload...)
}

// TestBuildFailsClosedWhenTheGoChildCannotStart drives a real refused process
// creation through the complete operation. The Windows active-process bound is
// the only inventory control that can refuse the compiler natively, so the
// aggregate domain is set to admit the worker alone and the compiler creation is
// refused by the kernel.
func TestBuildFailsClosedWhenTheGoChildCannotStart(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("active-process-count-limit is available only on Windows in rc5-native-control-inventory-v1")
	}
	fixture := newSnapshotFixture(t)
	fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
	before := len(fixture.calls())

	_, err := Build(context.Background(), fixture.request(ResourceLimits{Timeout: 30 * time.Second, Processes: 1}))
	if DiagnosticCode(err) != "go_list_failed" {
		t.Fatalf("error = %v, want the fixed go list command to fail closed", err)
	}
	if after := len(fixture.calls()); after != before {
		t.Fatalf("launcher calls went from %d to %d; the refused child must never run", before, after)
	}
	requireNoPrivateArtifact(t, fixture.session)
}

// TestWorkerDomainTeardownSurvivesARefusedChild proves the complete domain is
// still terminated and joined after a refused process creation.
func TestWorkerDomainTeardownSurvivesARefusedChild(t *testing.T) {
	scenario := newWorkerScenario(t)
	scenario.request.Directory = filepath.Join(scenario.fixture.sourceDir, "absent")
	worker := scenario.requireReady()
	worker.send(workerMessage{Kind: kindList, Nonce: scenario.nonce})
	if message := worker.receive(); message.Result == nil || !message.Result.StartFailed {
		t.Fatalf("message = %+v, want a refused creation", message)
	}
	pid := worker.command.Process.Pid
	worker.close()
	deadline := time.Now().Add(15 * time.Second)
	for processAlive(pid) && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if processAlive(pid) {
		t.Fatalf("worker %d outlived the domain teardown", pid)
	}
	if _, err := os.Lstat(scenario.request.ArtifactPath); err == nil {
		t.Fatal("a refused build produced an artifact")
	}
}
