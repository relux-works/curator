//go:build darwin

package homelease

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/homelease/state"
)

// Claim: the real observer reads the exact proc_pidinfo tuple of the test's
// own child, and the same read after the child is reaped refuses as vanished.
func TestDarwinKernelReadsOwnChild(t *testing.T) {
	sleep, err := exec.LookPath("sleep")
	if err != nil {
		t.Fatalf("owned child fixture requires a sleep binary: %v", err)
	}
	cmd := exec.Command(sleep, "30")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	pid := int64(cmd.Process.Pid)
	process, err := readKernelProcess(pid)
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("read own child: %v", err)
	}
	if process.PID != pid || process.UID != int64(os.Getuid()) || process.StartTime == nil ||
		process.StartTime.Kind != state.StartDarwin || process.StartTime.Sec == "0" || process.StartTime.Usec > 999999 {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("tuple does not name the owned child: %+v", process)
	}
	again, err := readKernelProcess(pid)
	if err != nil || !state.ProcessEqual(&process, &again) {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("repeat read drifted: %+v vs %+v (%v)", process, again, err)
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = cmd.Wait()
	if _, err := readKernelProcess(pid); err == nil {
		t.Fatal("vanished child still produced a tuple")
	}
}

// Claim: the darwin kernel read refuses a zombie child, while the same child alive beside it reads normally.
func TestDarwinKernelRefusesZombie(t *testing.T) {
	sleep, err := exec.LookPath("sleep")
	if err != nil {
		t.Fatalf("owned child fixture requires a sleep binary: %v", err)
	}
	cmd := exec.Command(sleep, "30")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	pid := int64(cmd.Process.Pid)
	// Valid control: the live child reads normally before it is killed.
	if _, err := readKernelProcess(pid); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("live child control: %v", err)
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	// Unreaped, the child is a zombie; the production read must start
	// refusing instead of returning its retained tuple.
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := readKernelProcess(pid); err != nil {
			break
		}
		if time.Now().After(deadline) {
			_ = cmd.Wait()
			t.Fatal("zombie child still produced a tuple")
		}
		time.Sleep(10 * time.Millisecond)
	}
	_ = cmd.Wait()
	if _, err := readKernelProcess(pid); err == nil {
		t.Fatal("reaped child still produced a tuple")
	}
}
