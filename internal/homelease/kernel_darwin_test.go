//go:build darwin

package homelease

import (
	"errors"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/homelease/state"
	"golang.org/x/sys/unix"
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

// darwinZombieOracle reports, independently of the production proc_pidinfo
// read, whether pid is SZOMB per the kern.proc.pid sysctl. The zombie test
// asserts the production refusal only once this oracle says zombie, so the
// refusal is proven to be the SZOMB branch deciding, not a lookup that
// cannot reach zombies at all.
func darwinZombieOracle(pid int64) (bool, error) {
	info, err := unix.SysctlKinfoProc("kern.proc.pid", int(pid))
	if err != nil {
		return false, err
	}
	return int64(info.Proc.P_stat) == procStatusZombie, nil
}

// Claim: the darwin kernel read refuses a zombie child with the exact
// exited refusal (errProcessExited, which VerifyBind maps to
// environment_lease_authority_refused/authority_unavailable), while the
// same child alive beside it reads normally. Any other error, ESRCH
// included, fails this test by name, as does a returned tuple.
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
	if zombie, err := darwinZombieOracle(pid); err != nil || zombie {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("live child oracle control: zombie=%t err=%v", zombie, err)
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	// Unreaped, the child is a zombie. Poll the independent sysctl oracle
	// until the kernel says SZOMB, then the production read must give
	// exactly the exited refusal rather than its retained tuple, and
	// rather than the ESRCH a zombie-blind proc_pidinfo arg would give.
	deadline := time.Now().Add(5 * time.Second)
	for {
		zombie, err := darwinZombieOracle(pid)
		if err == nil && zombie {
			break
		}
		if time.Now().After(deadline) {
			_ = cmd.Wait()
			t.Fatalf("killed child never became a zombie per kern.proc.pid oracle: zombie=%t err=%v", zombie, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := readKernelProcess(pid); err == nil {
		_ = cmd.Wait()
		t.Fatal("zombie child still produced a tuple")
	} else if !errors.Is(err, errProcessExited) {
		_ = cmd.Wait()
		t.Fatalf("zombie refusal is not the exact exited refusal: %v", err)
	}
	_ = cmd.Wait()
	if _, err := readKernelProcess(pid); err == nil {
		t.Fatal("reaped child still produced a tuple")
	}
}
