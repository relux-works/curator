//go:build darwin || linux

package godriver

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Paired real-worker witnesses reject omitted domain termination at the
// survival assertion. A setup failure, hang, skip or unrelated error never
// counts as detection. Linux also probes descriptor leaks and pre-fork refusals.
// The Go witness rejects restored readiness sleeps.
func TestBuildDescendantNarrowingControls(t *testing.T) {
	goTool := os.Getenv("CURATOR_GO")
	if goTool == "" {
		goTool = "go"
	}
	plants := []struct{ name, file, before, after, witness, message string }{
		{"omit-group-kill", "controls_" + runtime.GOOS + ".go", "_ = unix.Kill(-pid, unix.SIGKILL)", "_ = pid", "TestBuildTerminatesTheCompleteWorkerDomain", "outlived the worker domain teardown: survival marker"},
		{"restore-readiness-sleep", "testdata/stubgo/main.go", "_, readyErr := io.ReadFull(ready, acknowledgement[:])", "_, readyErr := 0, error(nil)\n time.Sleep(500*time.Millisecond)\n acknowledgement[0] = 1\n _ = io.ReadFull", "TestBuildWaitsForDelayedDescendantReadiness", "tool children did not publish their identities"},
		{"stop-EOF-watch-at-shutdown", "workerserver.go", "if err != nil {\n\t\t\t\treturn\n\t\t\t}\n\t\t}\n\t}()\n\treturn messages", "if err != nil || message.Kind == kindShutdown {\n\t\t\t\treturn\n\t\t\t}\n\t\t}\n\t}()\n\treturn messages", "TestWorkerEOFWhileShutdownQueued", "queued shutdown hid EOF"},
		{"omit-acceptance-deadline", "build.go", "if err := client.deadlineFailure(); err != nil {\n\t\treturn Result{}, err\n\t}\n\tfinishTeardown", "finishTeardown", "TestBuildRejectsArtifactDuringAcceptance", "post-bracket deadline accepted or misdiagnosed"},
	}
	if runtime.GOOS == "linux" {
		plants = append(plants,
			struct{ name, file, before, after, witness, message string }{"omit-deadline-domain-termination", "orphan_unix.go", "orphanKill(-os.Getpid(), unix.SIGKILL)", "error(nil)", "TestBuildDeadlineTearsDownDescendants", "timed-out descendant published survival marker"},
			struct{ name, file, before, after, witness, message string }{"omit-descriptor-seal", "controls_linux.go", "restrictManagerDescriptors(domain.descriptorRoot)", "error(nil)", "TestLinuxNativeControls/descriptor", "private descriptor leaked"},
			struct{ name, file, before, after, witness, message string }{"extra-file-admitted", "controls_linux.go", "len(command.ExtraFiles) != 0", "len(command.ExtraFiles) > 1", "TestLinuxControlLaunchRefusals/extra-files", "extra-files was admitted"},
			struct{ name, file, before, after, witness, message string }{"sealing-error-admitted", "controls_linux.go", "if err := restrictManagerDescriptors(domain.descriptorRoot); err != nil {", "if err := restrictManagerDescriptors(domain.descriptorRoot); err != nil && false {", "TestLinuxControlLaunchRefusals/sealing-failure", "sealing-failure was admitted"},
		)
	}
	for _, plant := range plants {
		t.Run(plant.name, func(t *testing.T) {
			sourcePath, err := filepath.Abs(plant.file)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(sourcePath)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(raw), "\r\n", "\n")
			if strings.Count(source, plant.before) != 1 {
				t.Fatal("descendant plant anchor not unique")
			}
			run := func(overlay string) ([]byte, error) {
				// The 90-second hang guard is below the former 120-second natural
				// lifetime, but above the fixture's 60-second worker allowance plus
				// 30 seconds for warm setup. Only the assertion supplies a verdict.
				args := []string{"test", "-count=1", "-timeout=90s", "-v", "-run", "^" + plant.witness + "$", "."}
				if overlay != "" {
					args = append([]string{"test", "-overlay", overlay}, args[1:]...)
				}
				command := exec.CommandContext(t.Context(), goTool, args...)
				for _, item := range os.Environ() {
					if !strings.HasPrefix(item, "GIT_DIR=") && !strings.HasPrefix(item, "GIT_WORK_TREE=") && !strings.HasPrefix(item, "GOFLAGS=") {
						command.Env = append(command.Env, item)
					}
				}
				if overlay != "" && strings.HasPrefix(plant.file, "testdata/") {
					// The stub is compiled by the witness in a separate Go command;
					// a parent -overlay flag alone does not reach that build.
					command.Env = append(command.Env, "GOFLAGS=-overlay="+overlay)
				}
				return command.CombinedOutput()
			}
			output, err := run("")
			if err != nil || !strings.Contains(string(output), "--- PASS: "+plant.witness) {
				t.Fatalf("positive descendant control: %v\n%s", err, output)
			}
			scratch := t.TempDir()
			replacement := filepath.Join(scratch, filepath.Base(plant.file))
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, plant.before, plant.after, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			replacements := map[string]string{sourcePath: replacement}
			if plant.name == "omit-group-kill" {
				// Parent teardown and the worker's EOF/deadline watcher independently
				// kill this group. Suppress both to expose a genuine surviving domain.
				worker, err := filepath.Abs("orphan_unix.go")
				if err != nil {
					t.Fatal(err)
				}
				raw, err := os.ReadFile(worker)
				if err != nil {
					t.Fatal(err)
				}
				const anchor = "orphanKill(-os.Getpid(), unix.SIGKILL)"
				if strings.Count(string(raw), anchor) != 1 {
					t.Fatal("worker group plant anchor not unique")
				}
				replacementWorker := filepath.Join(scratch, "orphan_unix.go")
				if err := os.WriteFile(replacementWorker, []byte(strings.Replace(string(raw), anchor, "error(nil)", 1)), 0600); err != nil {
					t.Fatal(err)
				}
				replacements[worker] = replacementWorker
			}
			if plant.name == "omit-deadline-domain-termination" {
				parent, err := filepath.Abs("controls_linux.go")
				if err != nil {
					t.Fatal(err)
				}
				raw, err := os.ReadFile(parent)
				if err != nil {
					t.Fatal(err)
				}
				const anchor = "_ = unix.Kill(-pid, unix.SIGKILL)"
				if strings.Count(string(raw), anchor) != 1 {
					t.Fatal("parent deadline plant anchor not unique")
				}
				replacementParent := filepath.Join(scratch, "controls_linux.go")
				if err := os.WriteFile(replacementParent, []byte(strings.Replace(string(raw), anchor, "_ = pid", 1)), 0600); err != nil {
					t.Fatal(err)
				}
				replacements[parent] = replacementParent
			}
			payload, err := json.Marshal(map[string]any{"Replace": replacements})
			if err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(scratch, "overlay.json")
			if err := os.WriteFile(overlay, payload, 0600); err != nil {
				t.Fatal(err)
			}
			output, err = run(overlay)
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+plant.witness) || !strings.Contains(string(output), plant.message) {
				t.Fatalf("descendant mutant survived or failed outside its assertion: %v\n%s", err, output)
			}
			t.Logf("intended assertion detected %s:\n%s", plant.name, output)
		})
	}
}
