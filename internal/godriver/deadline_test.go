package godriver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

// The same closed response channel is a timeout only after the real context
// deadline. A live session retains the exact protocol-invalid diagnosis.
func TestWorkerChannelFailureDistinguishesParentDeadline(t *testing.T) {
	for _, expired := range []bool{false, true} {
		t.Run(map[bool]string{false: "protocol", true: "deadline"}[expired], func(t *testing.T) {
			ctx := context.Background()
			if expired {
				var cancel context.CancelFunc
				ctx, cancel = context.WithDeadline(ctx, time.Now().Add(-time.Second))
				defer cancel()
			}
			client := &workerClient{runContext: ctx, stdout: io.NopCloser(bytes.NewReader(nil)), stderr: &boundedBuffer{budget: &outputBudget{remaining: 1024}}}
			_, err := client.receive(kindBuildResult)
			want := CodeWorkerProtocolInvalid
			if expired {
				want = "process_timeout"
			}
			if DiagnosticCode(err) != want {
				t.Fatalf("error=%v want %s", err, want)
			}
			if expired && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("deadline cause lost: %v", err)
			}
		})
	}
}

// A deadline expires after a real worker acknowledgement. The live control
// lists successfully; expiry closes the production channel with process_timeout.
func TestRealWorkerParentDeadline(t *testing.T) {
	for _, expire := range []bool{false, true} {
		t.Run(map[bool]string{false: "live", true: "expired"}[expire], func(t *testing.T) {
			scenario := newWorkerScenario(t)
			req := scenario.request
			plan := workerPlan{Executable: scenario.identity, GoExecutable: req.GoExecutable, GOROOT: req.GOROOT, ToolDirectory: req.ToolDirectory,
				Directory: req.Directory, Environment: req.Environment, ListArgv: req.ListArgv, BuildArgv: req.BuildArgv, ArtifactPath: req.ArtifactPath,
				ReadOnlyRoots: req.ReadOnlyRoots, PrivateRoots: req.PrivateRoots, Platform: req.Platform, Probes: scenario.probes, Limits: scenario.limits}
			deadline := time.Now().Add(time.Minute)
			if expire {
				deadline = time.Now().Add(3 * time.Second)
			}
			ctx, cancel := context.WithDeadline(t.Context(), deadline)
			defer cancel()
			client, err := launchWorker(ctx, plan)
			if err != nil {
				t.Fatal(err)
			}
			defer client.teardown()
			if expire {
				<-ctx.Done()
			}
			_, err = client.list()
			if expire {
				if DiagnosticCode(err) != "process_timeout" {
					t.Fatalf("deadline error=%v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Two separately admissible children cannot each reset the whole-session
// deadline. The short valid control still completes through the same Build.
func TestWorkerSharesWholeSessionBudget(t *testing.T) {
	for _, slow := range []bool{false, true} {
		t.Run(map[bool]string{false: "valid", true: "combined-overrun"}[slow], func(t *testing.T) {
			fixture := newSnapshotFixture(t)
			delay := 20
			if slow {
				delay = 800
			}
			fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact", ListSleepMS: delay, BuildSleepMS: delay})
			result, err := Build(t.Context(), fixture.request(ResourceLimits{Timeout: 1300 * time.Millisecond}))
			if slow {
				if DiagnosticCode(err) != "process_timeout" || result.Artifact.StagedPath != "" {
					t.Fatalf("combined budget was renewed: result=%+v err=%v", result, err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}
