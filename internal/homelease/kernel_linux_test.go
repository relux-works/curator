//go:build linux

package homelease

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/homelease/state"
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
	pid, ticks, err := parseLinuxStat([]byte(line))
	if err != nil || pid != 4242 || ticks != "987654321" {
		t.Fatalf("valid control with hostile comm: %d %q %v", pid, ticks, err)
	}
	for name, bad := range map[string]string{
		"truncated":      "4242 (sleep) S 1 2 3\n",
		"no parenthesis": "4242 sleep S 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20\n",
		"empty":          "",
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := parseLinuxStat([]byte(bad)); err == nil {
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
			statuses: []string{"Name:\tfixture\nUid:\t1001\t1001\t1001\t1001\n"},
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
		t.Skip("no sleep binary for an owned child fixture")
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
