//go:build !windows

package testcli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// These controls exercise the same subprocess runner as Run and Output:
// ordinary exit and streams survive; cancellation stops a stalled shell and
// its pipe-holding child. They do not cover descendants that leave the group.
func TestRunCommandProcessLifetime(t *testing.T) {
	t.Run("ordinary_exit_and_streams", func(t *testing.T) {
		code, out, stderr, err := runCommand(t.Context(), "", nil, "input", "sh", "-c", "cat; printf diagnostic >&2; exit 7")
		if err != nil || code != 7 || out != "input" || stderr != "diagnostic" {
			t.Fatalf("got (%d, %q, %q, %v)", code, out, stderr, err)
		}
	})
	t.Run("expired_context_starts_nothing", func(t *testing.T) {
		marker := filepath.Join(t.TempDir(), "started")
		ctx, cancel := context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
		defer cancel()
		_, _, _, err := runCommand(ctx, "", nil, "", "sh", "-c", "touch \"$1\"", "sh", marker)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("error = %v, want deadline", err)
		}
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatalf("expired command wrote marker: %v", err)
		}
	})
	// Normal leader exit must also retire a quiet background descendant.
	t.Run("return_cleans_quiet_descendant", func(t *testing.T) {
		marker := filepath.Join(t.TempDir(), "pid")
		code, _, _, err := runCommand(t.Context(), "", nil, "", "sh", "-c", `sleep 60 >/dev/null 2>&1 & printf '%s' $! >"$1"`, "sh", marker)
		if err != nil || code != 0 {
			t.Fatalf("normal leader exit = %d, %v", code, err)
		}
		data, err := os.ReadFile(marker)
		if err != nil {
			t.Fatal(err)
		}
		pid, err := strconv.Atoi(string(data))
		if err != nil || pid <= 0 {
			t.Fatalf("invalid descendant pid %q", data)
		}
		t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
		until := time.Now().Add(time.Second)
		for syscall.Kill(pid, 0) == nil && time.Now().Before(until) {
			time.Sleep(10 * time.Millisecond)
		}
		if err := syscall.Kill(pid, 0); err != syscall.ESRCH {
			t.Fatalf("quiet descendant pid %d remains: %v", pid, err)
		}
	})

	t.Run("cancel_stalled_descendants", func(t *testing.T) {
		marker := filepath.Join(t.TempDir(), "pids")
		ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
		defer cancel()
		type result struct {
			code int
			err  error
		}
		done := make(chan result, 1)
		go func() {
			code, _, _, err := runCommand(ctx, "", nil, "", "sh", "-c", "sleep 60 & child=$!; printf '%s %s' $$ $child > \"$1\"; wait", "sh", marker)
			done <- result{code, err}
		}()
		var pids []int
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			data, _ := os.ReadFile(marker)
			fields := strings.Fields(string(data))
			if len(fields) == 2 {
				for _, s := range fields {
					n, err := strconv.Atoi(s)
					if err != nil || n <= 0 {
						t.Fatalf("invalid pid %q", s)
					}
					pids = append(pids, n)
				}
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if len(pids) != 2 {
			cancel()
			<-done
			t.Fatal("fixture never started both processes")
		}
		// Exact recorded fixture PIDs only, including when the negative plant fails.
		t.Cleanup(func() {
			for _, pid := range pids {
				_ = syscall.Kill(pid, syscall.SIGKILL)
			}
		})
		before := time.Now()
		cancel()
		select {
		case got := <-done:
			if got.err != nil || got.code != -1 {
				t.Errorf("cancel result = %+v", got)
			}
		case <-time.After(4 * time.Second):
			t.Fatal("cancellation did not finish owned wait and pipes")
		}
		if time.Since(before) > 4*time.Second {
			t.Fatal("cancellation exceeded bound")
		}
		for _, pid := range pids {
			until := time.Now().Add(time.Second)
			for syscall.Kill(pid, 0) == nil && time.Now().Before(until) {
				time.Sleep(10 * time.Millisecond)
			}
			if err := syscall.Kill(pid, 0); err != syscall.ESRCH {
				t.Errorf("fixture pid %d remains after cancellation: %v", pid, err)
			}
		}
	})
}
