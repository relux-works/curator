package godriver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
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
			timeout := 5 * time.Second
			if slow {
				timeout = 1300 * time.Millisecond
			}
			result, err := Build(t.Context(), fixture.request(ResourceLimits{Timeout: timeout}))
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

// A successful compiler result is still rejected when the parent deadline
// expires before verification/acceptance. The live control publishes normally.
func TestBuildRejectsArtifactAfterParentDeadline(t *testing.T) {
	for _, expire := range []bool{false, true} {
		t.Run(map[bool]string{false: "live", true: "expired"}[expire], func(t *testing.T) {
			fixture := newSnapshotFixture(t)
			fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
			duration := time.Minute
			if expire {
				duration = 2 * time.Second
			}
			ctx, cancel := context.WithTimeout(t.Context(), duration)
			defer cancel()
			var completed bool
			fixture.session.observePhase = func(phase string, _ time.Duration) {
				if phase == "build" {
					completed = true
					if expire {
						<-ctx.Done()
					}
				}
			}
			result, err := Build(ctx, fixture.request(ResourceLimits{}))
			if !completed {
				t.Fatal("compiler did not complete before the acceptance barrier")
			}
			if expire {
				if DiagnosticCode(err) != "process_timeout" || result.Artifact.StagedPath != "" {
					t.Fatalf("expired artifact accepted: result=%+v err=%v", result, err)
				}
			} else if err != nil || result.Artifact.StagedPath == "" {
				t.Fatalf("live artifact rejected: result=%+v err=%v", result, err)
			}
		})
	}
}

// Expiry is held until after the build bracket: the final guard alone prevents
// successful acceptance. A verification failure after expiry must also say timeout.
func TestBuildRejectsArtifactDuringAcceptance(t *testing.T) {
	for _, mode := range []string{"live", "expired", "expired-artifact-error"} {
		t.Run(mode, func(t *testing.T) {
			fixture := newSnapshotFixture(t)
			fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact"})
			reached := false
			fixture.session.beforeAcceptance = func(ctx context.Context) {
				reached = true
				if mode == "live" {
					return
				}
				<-ctx.Done()
				if mode == "expired-artifact-error" {
					matches, err := filepath.Glob(filepath.Join(fixture.session.operation, ".curator-go-build-*", "bin", "*"))
					if err != nil || len(matches) != 1 {
						t.Fatalf("artifact acceptance setup: %v %v", matches, err)
					}
					if err := os.Remove(matches[0]); err != nil {
						t.Fatal(err)
					}
				}
			}
			result, err := Build(t.Context(), fixture.request(ResourceLimits{Timeout: 5 * time.Second}))
			if !reached {
				t.Fatalf("did not reach post-bracket verification: %v", err)
			}
			if mode == "live" {
				if err != nil || result.Artifact.StagedPath == "" {
					t.Fatalf("live: %+v %v", result, err)
				}
			} else if DiagnosticCode(err) != "process_timeout" || result.Artifact.StagedPath != "" {
				t.Fatalf("post-bracket deadline accepted or misdiagnosed: %+v %v", result, err)
			}
		})
	}
}

// Caller intent stays distinct from identity/protocol faults before and during builds.
func TestBuildCallerContextDiagnostics(t *testing.T) {
	for _, when := range []string{"before", "during"} {
		for _, kind := range []string{"cancel", "deadline"} {
			t.Run(when+"-"+kind, func(t *testing.T) {
				fixture := newSnapshotFixture(t)
				fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact", BuildSleepMS: 60000})
				var ctx context.Context
				var cancel context.CancelFunc
				if kind == "deadline" {
					duration := 3 * time.Second
					if when == "before" {
						duration = -time.Second
					}
					ctx, cancel = context.WithTimeout(t.Context(), duration)
				} else {
					ctx, cancel = context.WithCancel(t.Context())
					if when == "before" {
						cancel()
					} else {
						time.AfterFunc(3*time.Second, cancel)
					}
				}
				defer cancel()
				result, err := Build(ctx, fixture.request(ResourceLimits{}))
				want := "process_canceled"
				cause := context.Canceled
				if kind == "deadline" {
					want = "process_timeout"
					cause = context.DeadlineExceeded
				}
				if DiagnosticCode(err) != want || !errors.Is(err, cause) || result.Artifact.StagedPath != "" {
					t.Fatalf("context diagnostic: %+v %v want%s", result, err, want)
				}
			})
		}
	}
}
