package testhomeguard

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/testgitenv"
)

const deadlineMode = "CURATOR_TEST_GUARD_DEADLINE_MODE"
const deadlineRoot = "CURATOR_TEST_GUARD_DEADLINE_ROOT"

// Only this sentinel-checked test binary can deliberately stall before its
// observer handshake. Its descendant holds stdout/stderr open until killed;
// EOF therefore witnesses termination of both fixture processes. No path is
// selected from an account home or provider environment.
func init() {
	mode := os.Getenv(deadlineMode)
	if mode != "descendant" && ((mode != "stalled" && mode != "early-exit") || os.Getenv(supervisorPID) != strconv.Itoa(os.Getppid())) {
		return
	}
	root := os.Getenv(deadlineRoot)
	marker, err := os.ReadFile(filepath.Join(root, "deadline-fixture"))
	if err != nil || !filepath.IsAbs(root) || string(marker) != "synthetic deadline fixture" {
		os.Exit(94)
	}
	name := "child.pid"
	if mode == "descendant" {
		name = "descendant.pid"
	}
	if err := os.WriteFile(filepath.Join(root, name), []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
		os.Exit(94)
	}
	if mode != "descendant" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestSupervisorDeadlineChild$") // #nosec G204,G702 -- Fixed test-only child reexec; no shell.
		for _, entry := range os.Environ() {
			if !strings.HasPrefix(entry, deadlineMode+"=") {
				cmd.Env = append(cmd.Env, entry)
			}
		}
		cmd.Env = append(cmd.Env, deadlineMode+"=descendant")
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Start(); err != nil {
			os.Exit(94)
		}
	}
	fmt.Fprintln(os.Stderr, "synthetic never-connecting "+mode+" started")
	if mode == "descendant" {
		if err := os.WriteFile(filepath.Join(root, "descendant.ready"), nil, 0600); err != nil {
			os.Exit(94)
		}
	}
	if mode == "early-exit" {
		// Wait only for the controlled descendant's startup marker, so this
		// case cannot pass merely by killing it before its code was reached.
		until := time.Now().Add(time.Second)
		for time.Now().Before(until) {
			if _, err := os.Stat(filepath.Join(root, "descendant.ready")); err == nil {
				os.Exit(0)
			}
			time.Sleep(time.Millisecond)
		}
		os.Exit(94)
	}
	for {
		time.Sleep(time.Hour)
	}
}

// Missing authentication must stop the owned child and its pipe-holding
// descendant within a bounded total time, then check the final postimage.
// The same supervisor admits an authenticated, isolated child.
func TestSupervisorAuthenticationDeadline(t *testing.T) {
	for _, mode := range []string{"connected", "stalled", "early-exit"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "deadline-fixture"), []byte("synthetic deadline fixture"), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.v", "-test.run=^TestSupervisorDeadlineChild$") // #nosec G204,G702 -- Fixed test-only child reexec; no shell.
			for _, entry := range os.Environ() {
				key, _, _ := strings.Cut(entry, "=")
				if key == "HOME" || key == "USERPROFILE" || strings.HasPrefix(key, "XDG_") || strings.HasPrefix(key, "CURATOR_TEST_HOME_GUARD_") || key == "CODEX_HOME" || key == "CLAUDE_CONFIG_DIR" || key == "PI_CODING_AGENT_DIR" || key == deadlineMode || key == deadlineRoot {
					continue
				}
				cmd.Env = append(cmd.Env, entry)
			}
			cmd.Env = append(cmd.Env, "HOME="+root, "USERPROFILE="+root, deadlineRoot+"="+root, deadlineMode+"="+mode)
			// Failure-path cleanup names only PIDs planted by this fixture.
			needsCleanup := true
			defer func() {
				if !needsCleanup {
					return
				}
				for _, name := range []string{"child.pid", "descendant.pid"} {
					raw, err := os.ReadFile(filepath.Join(root, name))
					if err != nil {
						continue
					}
					pid, err := strconv.Atoi(string(raw))
					if err == nil && pid > 1 {
						if process, err := os.FindProcess(pid); err == nil {
							_ = process.Kill()
						}
					}
				}
			}()
			// Use our own pipe: exec.WaitDelay can mask a surviving descendant
			// when the leader already returned a nonzero exit. Require real EOF.
			reader, writer, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = reader.Close() }()
			cmd.Stdout, cmd.Stderr = writer, writer
			start := time.Now()
			if err := cmd.Start(); err != nil {
				_ = writer.Close()
				t.Fatal(err)
			}
			_ = writer.Close()
			output := make(chan []byte, 1)
			go func() { raw, _ := io.ReadAll(reader); output <- raw }()
			err = cmd.Wait()
			var out []byte
			missingEOF := false
			select {
			case out = <-output:
			case <-time.After(time.Second):
				missingEOF = true
				_ = reader.Close()
				out = <-output
			}
			elapsed := time.Since(start)
			needsCleanup = missingEOF || ctx.Err() != nil
			if missingEOF {
				t.Errorf("descendant retained output after supervisor exit")
			}
			if ctx.Err() != nil || elapsed >= 12*time.Second {
				t.Fatalf("supervisor exceeded bounded total time (%s): %v\n%s", elapsed, err, out)
			}
			if !strings.Contains(string(out), "protected native homes unchanged") {
				t.Fatalf("final snapshot missing: %v\n%s", err, out)
			}
			if mode == "connected" {
				if err != nil || !strings.Contains(string(out), "synthetic authenticated dispatch") {
					t.Fatalf("authenticated control refused: %v\n%s", err, out)
				}
			} else {
				exit, ok := err.(*exec.ExitError)
				if !ok || exit.ExitCode() != 2 || !strings.Contains(string(out), "child observer handoff not verified") || strings.Contains(string(out), "synthetic authenticated dispatch") || !strings.Contains(string(out), "synthetic never-connecting descendant started") {
					t.Fatalf("deadline did not refuse its intended stalled tree: %v\n%s", err, out)
				}
				if mode == "stalled" && elapsed < handoffTimeout/2 {
					t.Fatal("control did not reach the authentication deadline")
				}
			}
			t.Logf("%s: final snapshot, elapsed %s", mode, elapsed.Round(time.Millisecond))
		})
	}
}

func TestSupervisorDeadlineChild(t *testing.T) {
	if os.Getenv(deadlineMode) == "" {
		return
	}
	if code, parent := Supervise(false); parent {
		os.Exit(code)
	}
	previous, _ := os.UserHomeDir()
	restore := testgitenv.Isolate()
	defer restore()
	if err := CheckIsolation(previous); err != nil {
		t.Fatal(err)
	}
	t.Log("synthetic authenticated dispatch")
}
