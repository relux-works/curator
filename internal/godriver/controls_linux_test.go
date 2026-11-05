//go:build linux

package godriver

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// Real kernel witnesses pair permitted execution with forbidden effects. These
// run on Linux CI; compilation on another host is not runtime evidence.
func TestLinuxNativeControls(t *testing.T) {
	limits, err := normalizeBuildLimits(ResourceLimits{FileBytes: 1024 * 1024, ArtifactBytes: 1024})
	if err != nil {
		t.Fatal(err)
	}
	_, probes, err := probeNativeControls(limits)
	if err != nil {
		t.Fatal(err)
	}
	launch := func(cmd *exec.Cmd) *controlDomain {
		t.Helper()
		domain, err := prepareControlDomain(limits, probes)
		if err != nil {
			t.Fatal(err)
		}
		if err := domain.launch(cmd); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { terminateWorkerDomain(cmd); domain.close() })
		return domain
	}
	t.Run("file-bound", func(t *testing.T) {
		// An under-limit write succeeds; an over-limit write fails and is bounded.
		for _, blocks := range []string{"1", "2048"} {
			target := filepath.Join(t.TempDir(), "output")
			cmd := exec.Command("/bin/sh", "-c", `exec dd if=/dev/zero of="$1" bs=1024 count="$2"`, "control", target, blocks)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			launch(cmd)
			err := cmd.Wait()
			if blocks == "1" && err != nil {
				t.Fatal(err)
			}
			if blocks == "2048" && err == nil {
				t.Fatal("oversized write succeeded")
			}
			info, statErr := os.Stat(target)
			if statErr != nil {
				t.Fatal(statErr)
			}
			if blocks == "2048" {
				var exit *exec.ExitError
				if !errors.As(err, &exit) {
					t.Fatalf("not a worker exit: %v", err)
				}
				status, ok := exit.Sys().(syscall.WaitStatus)
				signaled := ok && status.Signaled() && status.Signal() == syscall.SIGXFSZ
				if info.Size() != limits.FileBytes || (!signaled && !strings.Contains(stderr.String(), "File too large")) {
					t.Fatalf("write did not reach intended file-size rejection: size=%d err=%v stderr=%s", info.Size(), err, stderr.String())
				}
			}

			if info.Size() > limits.FileBytes {
				t.Fatalf("file exceeds bound: %d", info.Size())
			}
		}
	})
	t.Run("descriptor", func(t *testing.T) {
		// Even a deliberately inheritable manager FD must not reach the worker.
		target := filepath.Join(t.TempDir(), "private")
		fd, err := unix.Open(target, unix.O_CREAT|unix.O_RDWR, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = unix.Close(fd) }()
		if _, err := unix.FcntlInt(uintptr(fd), unix.F_SETFD, 0); err != nil {
			t.Fatal(err)
		}
		// -ef compares object identity, including symlinked TMPDIR components.
		// Exit 42 means the private object is inherited; other failures are not
		// evidence that the isolation witness detected a leak.
		probe := func() *exec.Cmd {
			return exec.Command("/bin/sh", "-c", `if [ /proc/self/fd/"$1" -ef "$2" ]; then exit 42; fi`, "control", strconv.Itoa(fd), target)
		}
		unsealed := probe()
		var detected *exec.ExitError
		if err := unsealed.Run(); !errors.As(err, &detected) || detected.ExitCode() != 42 {
			t.Fatalf("unsealed descriptor detection control missed the leak: %v", err)
		}
		cmd := probe()
		launch(cmd)
		if err := cmd.Wait(); err != nil {
			var exit *exec.ExitError
			if errors.As(err, &exit) && exit.ExitCode() == 42 {
				t.Fatalf("private descriptor leaked: %v", err)
			}
			t.Fatalf("descriptor probe failed outside the leak assertion: %v", err)
		}
		// The parent retains a usable descriptor after worker isolation.
		if _, err := unix.Write(fd, []byte("parent retains custody")); err != nil {
			t.Fatalf("parent descriptor was closed: %v", err)
		}
	})
	t.Run("descendant", func(t *testing.T) {
		// The shell exits first; its live child remains until group teardown.
		pidFile := filepath.Join(t.TempDir(), "pid")
		cmd := exec.Command("/bin/sh", "-c", `sleep 60 >/dev/null 2>&1 & echo $! > "$1"`, "control", pidFile)
		launch(cmd)
		if err := cmd.Wait(); err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(pidFile)
		if err != nil {
			t.Fatal(err)
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = unix.Kill(pid, unix.SIGKILL) }()
		if !processAlive(pid) {
			t.Fatal("positive descendant is not alive before teardown")
		}
		terminateWorkerDomain(cmd)
		deadline := time.Now().Add(5 * time.Second)
		for {
			raw, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
			// A zombie is dead and cannot act, although init may reap it later.
			if errors.Is(err, os.ErrNotExist) || (err == nil && strings.Contains(string(raw), ") Z ")) {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("descendant %d survived teardown: %s (%v)", pid, raw, err)
			}
			time.Sleep(10 * time.Millisecond)
		}
	})
}

// Rejections leave no worker and no applied evidence; the positive launch uses
// the same manager domain and command, with its real descriptor enumeration.
func TestLinuxControlLaunchRefusals(t *testing.T) {
	limits, err := normalizeBuildLimits(ResourceLimits{})
	if err != nil {
		t.Fatal(err)
	}
	_, probes, err := probeNativeControls(limits)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"valid", "extra-files", "sealing-failure"} {
		t.Run(name, func(t *testing.T) {
			domain, err := prepareControlDomain(limits, probes)
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("/bin/sh", "-c", "exit 0")
			t.Cleanup(func() { domain.destroy(cmd) })
			if name == "extra-files" {
				file, err := os.Open(os.DevNull)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = file.Close() })
				cmd.ExtraFiles = []*os.File{file}
			}
			if name == "sealing-failure" {
				// Exercise an actual enumeration error without changing live procfs.
				domain.descriptorRoot = filepath.Join(t.TempDir(), "absent-proc-fd")
			}
			err = domain.launch(cmd)
			if name == "valid" {
				if err != nil {
					t.Fatal(err)
				}
				if len(domain.installedControls()) != 3 {
					t.Fatal("valid launch lacks installed evidence")
				}
				if err := cmd.Wait(); err != nil {
					t.Fatal(err)
				}
				return
			}
			if DiagnosticCode(err) != CodeControlUnavailable {
				t.Fatalf("%s was admitted or reported the wrong refusal: %v", name, err)
			}
			if cmd.Process != nil || len(domain.installedControls()) != 0 {
				t.Fatalf("%s refusal created a worker or applied evidence", name)
			}
		})
	}
}

// Actual procfs admission and exact get/set fcntl errors fail closed; the valid
// descriptor is retained in the manager with close-on-exec installed.
func TestLinuxDescriptorProbeAndSyscallFailures(t *testing.T) {
	if available, err := probeNativeControlWithDescriptorRoot(ControlInheritedHandleRestriction, ResourceLimits{}, "/proc/self/fd"); !available || err != nil {
		t.Fatalf("actual procfs: %v %v", available, err)
	}
	if available, err := probeNativeControlWithDescriptorRoot(ControlInheritedHandleRestriction, ResourceLimits{}, filepath.Join(t.TempDir(), "absent")); available || !os.IsNotExist(err) {
		t.Fatalf("missing procfs admitted: %v %v", available, err)
	}
	for _, operation := range []int{0, unix.F_GETFD, unix.F_SETFD} {
		t.Run(strconv.Itoa(operation), func(t *testing.T) {
			file, err := os.CreateTemp(t.TempDir(), "descriptor")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			if _, err := unix.FcntlInt(file.Fd(), unix.F_SETFD, 0); err != nil {
				t.Fatal(err)
			}
			injected := false
			fcntl := func(fd uintptr, op, arg int) (int, error) {
				if fd == file.Fd() && op == operation && operation != 0 {
					injected = true
					return 0, unix.EPERM
				}
				return unix.FcntlInt(fd, op, arg)
			}
			err = restrictManagerDescriptorsWith("/proc/self/fd", fcntl)
			if operation == 0 {
				flags, flagErr := unix.FcntlInt(file.Fd(), unix.F_GETFD, 0)
				if err != nil || flagErr != nil || flags&unix.FD_CLOEXEC == 0 {
					t.Fatalf("positive seal: %v %v %d", err, flagErr, flags)
				}
			} else if !injected || !errors.Is(err, unix.EPERM) {
				t.Fatalf("fcntl error admitted: injected=%v err=%v", injected, err)
			}
		})
	}
}
