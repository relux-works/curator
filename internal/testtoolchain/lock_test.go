package testtoolchain

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A separate process holds the real OS lock until explicitly released. A
// deadline names that holder without modifying its record or admitting a waiter;
// after release the same waiter can acquire, report timing, and clean up.
func TestHostGOROOTLockContention(t *testing.T) {
	root := t.TempDir()
	pid, release := startHostGOROOTHolder(t, root)
	path := filepath.Join(root, "holder.json")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("stuck-holder-named-within-bound", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		started := time.Now()
		lock, err := acquireHostGOROOT(ctx, root, t.Name(), t.Logf)
		if lock != nil {
			_ = lock.Close()
			t.Fatal("waiter acquired the stuck holder's lock")
		}
		var failure *hostGOROOTWaitError
		if !errors.Is(err, context.DeadlineExceeded) || !errors.As(err, &failure) {
			t.Fatalf("error = %v, want named deadline refusal", err)
		}
		if elapsed := time.Since(started); elapsed < 50*time.Millisecond || elapsed > 5*time.Second {
			t.Fatalf("stuck-holder refusal took %s, want configured deadline within 5s watchdog", elapsed)
		}
		if failure.holder == nil || failure.holder.Name != "TestHostGOROOTLockHolderProcess/stuck" || failure.holder.PID != pid || failure.held < failure.waited || failure.waited < 50*time.Millisecond {
			t.Fatalf("deadline omitted the real holder or timing: %+v", failure)
		}
		message := err.Error()
		for _, part := range []string{"context deadline exceeded", "waiter=" + fmt.Sprintf("%q", t.Name()), `holder="TestHostGOROOTLockHolderProcess/stuck"`, fmt.Sprintf("pid=%d", pid), "process=" + fmt.Sprintf("%q", filepath.Base(os.Args[0])), "held=", "waited="} {
			if !strings.Contains(message, part) {
				t.Fatalf("deadline message %q lacks %q", message, part)
			}
		}
		if strings.Contains(message, root) {
			t.Fatalf("deadline disclosed a fixture path: %s", message)
		}
		after, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatalf("refused waiter changed holder metadata: %v", err)
		}
		t.Log(message)
	})

	t.Run("released-holder-admits-waiter", func(t *testing.T) {
		release()
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("released process retained holder metadata: %v", err)
		}
		var logs bytes.Buffer
		logf := func(format string, args ...any) { fmt.Fprintf(&logs, format+"\n", args...) }
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		lock, err := acquireHostGOROOT(ctx, root, t.Name(), logf)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := lock.Close(); err != nil {
				t.Error(err)
			}
		}()
		owner := readHostGOROOTOwner(path)
		if owner == nil || owner.Name != t.Name() || owner.PID != os.Getpid() {
			t.Fatalf("new acquisition did not replace holder identity: %+v", owner)
		}
		if err := lock.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("holder metadata survived release: %v", err)
		}
		for _, part := range []string{"lock waiting:", "lock acquired:", "waited=", "lock released:", "held="} {
			if !strings.Contains(logs.String(), part) {
				t.Fatalf("lifecycle telemetry lacks %q: %s", part, &logs)
			}
		}
		// Closing the old wrapper twice cannot erase a successor's metadata.
		next, err := acquireHostGOROOT(ctx, root, "successor", logf)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := next.Close(); err != nil {
				t.Error(err)
			}
		}()
		if err := lock.Close(); err != nil {
			t.Fatal(err)
		}
		if owner := readHostGOROOTOwner(path); owner == nil || owner.Name != "successor" {
			t.Fatalf("old release changed successor metadata: %+v", owner)
		}
	})
}

// Missing or malformed diagnostic metadata stays explicitly unknown; it never
// grants the OS lock. Restoring the real record recovers a named refusal, and
// releasing the process still permits a valid acquisition.
func TestHostGOROOTLockUnknownHolder(t *testing.T) {
	root := t.TempDir()
	_, release := startHostGOROOTHolder(t, root)
	path := filepath.Join(root, "holder.json")
	record, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []string{"missing", "malformed"} {
		t.Run(row, func(t *testing.T) {
			if row == "missing" {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(path, []byte("not-json"), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			lock, err := acquireHostGOROOT(ctx, root, t.Name(), t.Logf)
			if lock != nil {
				_ = lock.Close()
				t.Fatal("missing metadata admitted a waiter")
			}
			if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "holder=unknown (metadata unavailable)") || strings.Contains(err.Error(), root) {
				t.Fatalf("want path-free unknown-holder deadline, got %v", err)
			}
			if row == "missing" {
				if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("waiter created metadata: %v", err)
				}
			} else if data, err := os.ReadFile(path); err != nil || string(data) != "not-json" {
				t.Fatalf("waiter changed invalid metadata: %v", err)
			}
			if err := os.WriteFile(path, record, 0600); err != nil {
				t.Fatal(err)
			}
		})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	lock, err := acquireHostGOROOT(ctx, root, "restored-record", t.Logf)
	if lock != nil {
		_ = lock.Close()
		t.Fatal("restoring metadata granted the held lock")
	}
	if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), `holder="TestHostGOROOTLockHolderProcess/stuck"`) {
		t.Fatalf("restored record did not restore named refusal: %v", err)
	}
	release()
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	lock, err = acquireHostGOROOT(ctx, root, "after-release", t.Logf)
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
}

// A metadata-publication failure refuses without exposing its path or retaining
// the OS lock; repairing the fixture admits the next named owner.
func TestHostGOROOTLockMetadataFailureReleases(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "holder.json")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	lock, err := acquireHostGOROOT(ctx, root, "metadata-failure", t.Logf)
	if lock != nil {
		_ = lock.Close()
		t.Fatal("metadata failure admitted a holder")
	}
	if err == nil || err.Error() != "record host GOROOT test lock holder failed" {
		t.Fatalf("metadata refusal = %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".holder-") {
			t.Fatal("failed publication leaked its staging file")
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	lock, err = acquireHostGOROOT(ctx, root, "after-repair", t.Logf)
	if err != nil {
		t.Fatalf("failed publication retained the OS lock: %v", err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
}

// This child is a real cross-process holder, not a simulated lock result.
func TestHostGOROOTLockHolderProcess(t *testing.T) {
	root := os.Getenv("CURATOR_TEST_GOROOT_HOLDER_ROOT")
	if root == "" {
		return
	}
	lock, err := acquireHostGOROOT(context.Background(), root, "TestHostGOROOTLockHolderProcess/stuck", t.Logf)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("HOST_GOROOT_READY")
	_, err = io.Copy(io.Discard, os.Stdin)
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
}

func startHostGOROOTHolder(t *testing.T, root string) (int, func()) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestHostGOROOTLockHolderProcess$")
	command.Env = append(os.Environ(), "CURATOR_TEST_GOROOT_HOLDER_ROOT="+root)
	command.Stderr = os.Stderr
	input, err := command.StdinPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	output, err := command.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	released := false
	release := func() {
		t.Helper()
		if released {
			return
		}
		released = true
		_ = input.Close()
		err := command.Wait()
		cancel()
		if err != nil {
			t.Fatalf("holder process failed: %v", err)
		}
	}
	t.Cleanup(release)
	ready := make(chan bool, 1)
	go func() {
		scanner := bufio.NewScanner(output)
		found := false
		for scanner.Scan() {
			if scanner.Text() == "HOST_GOROOT_READY" {
				ready <- true
				found = true
			}
		}
		if !found {
			ready <- false
		}
	}()
	select {
	case ok := <-ready:
		if !ok {
			t.Fatal("holder process did not acquire the lock")
		}
	case <-ctx.Done():
		t.Fatal("holder readiness watchdog expired")
	}
	return command.Process.Pid, release
}
