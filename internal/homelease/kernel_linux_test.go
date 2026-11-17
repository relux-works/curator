//go:build linux

package homelease

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/homelease/state"
	"golang.org/x/sys/unix"
)

// Claim: stat field 22 is counted after the last ")" of comm, so a comm
// containing spaces and ")" cannot shift the starttime token, and a
// truncated line refuses instead of guessing.
func TestLinuxStatCommLastParenthesis(t *testing.T) {
	// Suffix tokens: field 3 first, so starttime (field 22) is token 20.
	suffix := make([]string, 25)
	for i := range suffix {
		suffix[i] = "0"
	}
	suffix[19] = "987654321"
	line := "4242 (weird ) comm name) " + strings.Join(suffix, " ") + "\n"
	pid, _, ticks, err := parseLinuxStat([]byte(line))
	if err != nil || pid != 4242 || ticks != "987654321" {
		t.Fatalf("valid control with hostile comm: %d %q %v", pid, ticks, err)
	}
	for name, bad := range map[string]string{
		"truncated":      "4242 (sleep) S 1 2 3\n",
		"no parenthesis": "4242 sleep S 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20\n",
		"empty":          "",
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, _, err := parseLinuxStat([]byte(bad)); err == nil {
				t.Fatal("unparseable stat accepted")
			}
		})
	}
}

type fakeLinuxPass struct {
	boots    []string
	stats    []string
	statuses []string
	bootErr  error
	statErr  error
	hzErr    error
}

func (f *fakeLinuxPass) bootID() (string, error) {
	if f.bootErr != nil {
		return "", f.bootErr
	}
	next := f.boots[0]
	f.boots = f.boots[1:]
	return next, nil
}
func (f *fakeLinuxPass) stat() ([]byte, error) {
	if f.statErr != nil {
		return nil, f.statErr
	}
	next := f.stats[0]
	f.stats = f.stats[1:]
	return []byte(next), nil
}
func (f *fakeLinuxPass) status() ([]byte, error) {
	next := f.statuses[0]
	f.statuses = f.statuses[1:]
	return []byte(next), nil
}
func (f *fakeLinuxPass) hz() (int64, error) { return 100, f.hzErr }
func (f *fakeLinuxPass) close() error       { return nil }

func linuxStatLine(pid int64, ticks string) string {
	suffix := make([]string, 25)
	for i := range suffix {
		suffix[i] = "0"
	}
	suffix[0] = "S"
	suffix[19] = ticks
	return fmt.Sprintf("%d (fixture) %s\n", pid, strings.Join(suffix, " "))
}

// Claim: one pass refuses when pid, start_time or boot_id changes between
// its paired reads, even though each individual reading is well-formed.
func TestLinuxPassIncarnationChanged(t *testing.T) {
	boot := "6f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10"
	stable := func() *fakeLinuxPass {
		return &fakeLinuxPass{
			boots:    []string{boot, boot},
			stats:    []string{linuxStatLine(4242, "111"), linuxStatLine(4242, "111")},
			statuses: []string{"Name:\tfixture\nTgid:\t4242\nUid:\t1001\t1001\t1001\t1001\n"},
		}
	}
	process, err := readLinuxProcess(4242, stable())
	if err != nil || process.PID != 4242 || process.UID != 1001 || process.BootID != boot ||
		process.StartTime == nil || process.StartTime.Ticks != "111" || process.StartTime.Hz != 100 {
		t.Fatalf("stable control: %+v %v", process, err)
	}
	cases := map[string]func(*fakeLinuxPass){
		"start-time": func(f *fakeLinuxPass) { f.stats[1] = linuxStatLine(4242, "112") },
		"pid":        func(f *fakeLinuxPass) { f.stats[1] = linuxStatLine(4243, "111") },
		"boot-id":    func(f *fakeLinuxPass) { f.boots[1] = "7f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			pass := stable()
			mutate(pass)
			if _, err := readLinuxProcess(4242, pass); !errors.Is(err, errKernelChanged) {
				t.Fatalf("changed incarnation not refused as changed: %v", err)
			}
		})
	}
	t.Run("unreadable stat", func(t *testing.T) {
		pass := stable()
		pass.statErr = os.ErrPermission
		if _, err := readLinuxProcess(4242, pass); err == nil || errors.Is(err, errKernelChanged) {
			t.Fatalf("unreadable stat not a plain refusal: %v", err)
		}
	})
}

// Claim: the real observer reads the exact tuple of the test's own child,
// and the same read after the child is reaped refuses as vanished.
func TestLinuxKernelReadsOwnChild(t *testing.T) {
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
		process.StartTime.Kind != state.StartLinux || process.StartTime.Ticks == "0" || process.StartTime.Hz < 1 {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("tuple does not name the owned child: %+v", process)
	}
	// A second independent pass over the same live child is byte-equal.
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

// parkedThreadID holds dedicated OS threads with runtime.LockOSThread,
// parked until the returned release runs, and hands back a non-leader tid
// that is guaranteed alive for the read. Picking an arbitrary entry from
// /proc/self/task instead could name a Go runtime thread that exits before
// the read, refusing with ENOENT for the wrong reason and sparing the
// thread plant by flake.
func parkedThreadID(t *testing.T) (int64, func()) {
	t.Helper()
	self := int64(os.Getpid())
	type parked struct {
		release chan struct{}
		done    chan struct{}
	}
	var threads []parked
	releaseAll := func() {
		for _, th := range threads {
			close(th.release)
			<-th.done
		}
	}
	// Only one OS thread has tid == pid (the leader), so parking several
	// locked threads yields a live non-leader tid among them.
	tid := int64(-1)
	for i := 0; i < 4 && tid < 0; i++ {
		th := parked{release: make(chan struct{}), done: make(chan struct{})}
		idCh := make(chan int64, 1)
		go func() {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			idCh <- int64(unix.Gettid())
			<-th.release
			close(th.done)
		}()
		threads = append(threads, th)
		if id := <-idCh; id != self {
			tid = id
		}
	}
	if tid < 0 {
		releaseAll()
		t.Fatal("no parked non-leader thread to refuse")
	}
	return tid, releaseAll
}

// Claim: the kernel read refuses a non-leader thread id and a zombie child with their exact refusals, while the thread-group leader and the live child beside them read normally.
func TestLinuxKernelRefusesThreadAndZombie(t *testing.T) {
	// Valid control: this process is its own thread-group leader.
	self, err := readKernelProcess(int64(os.Getpid()))
	if err != nil || self.PID != int64(os.Getpid()) {
		t.Fatalf("leader control: %+v %v", self, err)
	}
	tid, releaseThreads := parkedThreadID(t)
	_, threadErr := readKernelProcess(tid)
	releaseThreads()
	if threadErr == nil {
		t.Fatalf("thread id %d produced a process tuple", tid)
	}
	if !strings.Contains(threadErr.Error(), "process is a thread") {
		t.Fatalf("thread id %d refusal is not the exact thread refusal: %v", tid, threadErr)
	}

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
	// The unreaped child is a zombie; poll its stat until the kernel says
	// so, then the production read must refuse rather than reuse its tuple.
	deadline := time.Now().Add(5 * time.Second)
	for {
		raw, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if err == nil {
			if fields := strings.Fields(string(raw[strings.LastIndex(string(raw), ")")+1:])); len(fields) > 0 && fields[0] == "Z" {
				break
			}
		}
		if time.Now().After(deadline) {
			_ = cmd.Wait()
			t.Fatal("killed child never became a zombie")
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
}
