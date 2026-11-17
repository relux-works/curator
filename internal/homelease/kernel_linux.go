//go:build linux

package homelease

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/homelease/state"
	"golang.org/x/sys/unix"
)

// errKernelChanged marks an incarnation change inside one observation pass;
// the contract maps it to the mismatch row, never to a retried success.
var errKernelChanged = errors.New("kernel incarnation changed during observation pass")

func nativeStartTimeKind() string { return state.StartLinux }

// linuxPass is the injectable source set for one Linux observation pass, so
// pass-order controls can be tested without a second live process.
type linuxPass interface {
	bootID() (string, error)
	stat() ([]byte, error)
	status() ([]byte, error)
	hz() (int64, error)
	close() error
}

func readKernelProcess(pid int64) (state.Process, error) {
	pass, err := openLinuxPass(pid)
	if err != nil {
		return state.Process{}, err
	}
	defer func() { _ = pass.close() }()
	return readLinuxProcess(pid, pass)
}

// readLinuxProcess performs the single swap-free pass the contract requires:
// boot_id, stat, status, stat again, boot_id again. Compared readings that
// change refuse instead of combining two incarnations into one tuple.
func readLinuxProcess(pid int64, pass linuxPass) (state.Process, error) {
	bootBefore, err := pass.bootID()
	if err != nil {
		return state.Process{}, err
	}
	statBefore, err := pass.stat()
	if err != nil {
		return state.Process{}, err
	}
	pidBefore, ticksBefore, err := parseLinuxStat(statBefore)
	if err != nil {
		return state.Process{}, err
	}
	status, err := pass.status()
	if err != nil {
		return state.Process{}, err
	}
	uid, err := parseLinuxStatusUID(status)
	if err != nil {
		return state.Process{}, err
	}
	statAfter, err := pass.stat()
	if err != nil {
		return state.Process{}, err
	}
	pidAfter, ticksAfter, err := parseLinuxStat(statAfter)
	if err != nil {
		return state.Process{}, err
	}
	bootAfter, err := pass.bootID()
	if err != nil {
		return state.Process{}, err
	}
	if pidBefore != pidAfter || ticksBefore != ticksAfter || bootBefore != bootAfter {
		return state.Process{}, errKernelChanged
	}
	if pidBefore != pid {
		return state.Process{}, errKernelChanged
	}
	hz, err := pass.hz()
	if err != nil {
		return state.Process{}, err
	}
	process := state.Process{PID: pid, UID: uid, BootID: bootBefore,
		StartTime: &state.StartTime{Kind: state.StartLinux, Ticks: ticksBefore, Hz: hz}}
	if err := state.ValidateProcessSemantics(&process); err != nil {
		return state.Process{}, err
	}
	return process, nil
}

// parseLinuxStat extracts field 1 (pid) and field 22 (starttime ticks).
// Field 2 (comm) may contain spaces and ")", so fields after it are counted
// from the last ")": the first suffix token is field 3 and field 22 is the
// twentieth suffix token.
func parseLinuxStat(content []byte) (int64, string, error) {
	open := bytes.IndexByte(content, '(')
	closeParen := bytes.LastIndexByte(content, ')')
	if open < 1 || closeParen < open || closeParen+1 >= len(content) {
		return 0, "", fmt.Errorf("unparseable process stat")
	}
	pidText := strings.TrimSpace(string(content[:open]))
	pid, err := strconv.ParseInt(pidText, 10, 64)
	if err != nil || pid < 1 {
		return 0, "", fmt.Errorf("unparseable process stat pid")
	}
	suffix := strings.Fields(string(content[closeParen+1:]))
	if len(suffix) < 20 {
		return 0, "", fmt.Errorf("truncated process stat")
	}
	return pid, suffix[19], nil
}

func parseLinuxStatusUID(content []byte) (int64, error) {
	for _, line := range strings.Split(string(content), "\n") {
		if !strings.HasPrefix(line, "Uid:") {
			continue
		}
		fields := strings.Fields(strings.TrimPrefix(line, "Uid:"))
		if len(fields) < 1 {
			return 0, fmt.Errorf("unparseable process status uid")
		}
		uid, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("unparseable process status uid")
		}
		return uid, nil
	}
	return 0, fmt.Errorf("process status lacks uid")
}

type linuxProcPass struct{ dir *os.File }

func openLinuxPass(pid int64) (*linuxProcPass, error) {
	// #nosec G304 -- the path is /proc/<numeric pid> for the exact
	// incarnation the caller asked to observe; no caller text enters it.
	dir, err := os.Open(fmt.Sprintf("/proc/%d", pid))
	if err != nil {
		return nil, err
	}
	return &linuxProcPass{dir: dir}, nil
}

func (p *linuxProcPass) close() error { return p.dir.Close() }

func (p *linuxProcPass) read(name string) ([]byte, error) {
	fd, err := unix.Openat(int(p.dir.Fd()), name, unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), name)
	defer func() { _ = file.Close() }()
	return io.ReadAll(io.LimitReader(file, 1<<20))
}

func (p *linuxProcPass) stat() ([]byte, error)   { return p.read("stat") }
func (p *linuxProcPass) status() ([]byte, error) { return p.read("status") }

func (p *linuxProcPass) bootID() (string, error) {
	raw, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(raw)), nil
}

// hz reports sysconf(_SC_CLK_TCK) through its kernel source, the AT_CLKTCK
// auxiliary-vector entry glibc's sysconf returns on Linux.
func (p *linuxProcPass) hz() (int64, error) {
	raw, err := os.ReadFile("/proc/self/auxv")
	if err != nil {
		return 0, err
	}
	return parseAuxvClockTicks(raw)
}

const atClockTicks = 17 // AT_CLKTCK

func parseAuxvClockTicks(raw []byte) (int64, error) {
	word := strconv.IntSize / 8
	if word != 4 && word != 8 {
		return 0, fmt.Errorf("unsupported auxv word size")
	}
	if len(raw) == 0 || len(raw)%(2*word) != 0 {
		return 0, fmt.Errorf("unparseable auxv")
	}
	read := func(b []byte) uint64 {
		if word == 8 {
			return binary.NativeEndian.Uint64(b)
		}
		return uint64(binary.NativeEndian.Uint32(b))
	}
	for off := 0; off+2*word <= len(raw); off += 2 * word {
		key := read(raw[off : off+word])
		value := read(raw[off+word : off+2*word])
		if key == 0 {
			break
		}
		if key == atClockTicks {
			if value < 1 || value > 100000 {
				return 0, fmt.Errorf("clock tick rate outside range")
			}
			return int64(value), nil
		}
	}
	return 0, fmt.Errorf("auxv lacks clock tick rate")
}
