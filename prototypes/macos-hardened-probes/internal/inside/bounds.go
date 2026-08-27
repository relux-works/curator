package inside

// This file implements the aggregate-bound measurements: CPU time, address
// space, data segment and process count, each measured by executing a real
// stress against a real declared bound rather than by declaring what the
// platform is believed to do.
//
// The shape is the same for every bound and mirrors the rest of the harness:
//
//   - the bounded run installs the declared bound on itself and then tries to
//     exceed it. It is the positive capability test.
//   - the control run makes the identical attempt with no bound installed. It
//     is the negative control: without it a refusal could come from a stress
//     that never reached anything rather than from the bound.
//   - the nested run is a descendant of the bounded run that inherits the same
//     limit and repeats the attempt. It is the adversarial escape: a fresh
//     budget under one declared bound means the bound is accounted per process,
//     not over the domain.
//
// A separate escape process asks the other half of the question: whether a
// domain member can simply raise its own soft limit back to the hard limit it
// inherited, and whether lowering the hard limit closes that route.

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// Bound kinds. Each names one resource the hardened profile requires to be
// bounded over the whole descendant tree.
const (
	BoundCPU          = "cpu-milliseconds"
	BoundAddressSpace = "address-space-bytes"
	BoundDataSegment  = "data-segment-bytes"
	BoundProcessCount = "process-count"
)

// BoundKinds is the fixed order the matrix is measured and reported in.
func BoundKinds() []string {
	return []string{BoundCPU, BoundAddressSpace, BoundDataSegment, BoundProcessCount}
}

// Environment the bound processes read. Like the rest of the agent, every value
// comes from the harness; nothing under test chooses any of them.
const (
	EnvBoundKind     = "PROBE_BOUND_KIND"
	EnvBoundDeclared = "PROBE_BOUND_DECLARED"
	EnvBoundCeiling  = "PROBE_BOUND_CEILING"
	EnvBoundInstall  = "PROBE_BOUND_INSTALL"
	EnvBoundNest     = "PROBE_BOUND_NEST"
)

// The declared bounds and the ceilings the stress stops at. They are chosen by
// the harness and stated once, so the bound a process installs and the bound
// the report is scored against cannot drift apart.
//
// Every ceiling is above its declared bound on purpose: a stress that stopped
// at the bound could never show whether the bound or the stress ended the run.
const (
	// One CPU-second is the smallest bound RLIMIT_CPU can express, and the
	// smallest one keeps the probe fast.
	declaredCPUMillis = 1000
	ceilingCPUMillis  = 1600

	declaredMemoryBytes = 256 << 20
	ceilingMemoryBytes  = 512 << 20

	// A build domain of at most this many processes. The ceiling is above the
	// declared bound, like every other ceiling here, and both are small on
	// purpose: the control run really does start this many processes.
	declaredProcessCount = 4
	ceilingProcessCount  = 6
)

// BoundRun is what one stress process reached before it stopped.
type BoundRun struct {
	// Ran reports that the process started and produced a measurement. A run
	// that did not run says nothing, and must not be read as "reached nothing".
	Ran bool `json:"ran"`
	// Reached is the quantity the process obtained, in the kind's unit.
	Reached int64 `json:"reached"`
	// Refused reports that the kernel ended the attempt, as opposed to the
	// stress stopping at its own ceiling.
	Refused bool   `json:"refused"`
	Refusal string `json:"refusal,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

// BoundMeasurement is everything one bound kind produced in one run.
type BoundMeasurement struct {
	Kind     string `json:"kind"`
	Resource string `json:"resource"`
	Declared int64  `json:"declared"`
	Ceiling  int64  `json:"ceiling"`

	// Installed reports whether the declared bound could be installed at all.
	// On a platform that refuses the value there is nothing to enforce, which
	// is a different result from a bound that installs and does not bind.
	Installed    bool   `json:"installed"`
	InstallErrno string `json:"install_errno,omitempty"`
	// SoftLimitFloor is the lowest soft limit this kernel accepted for the
	// resource, measured only when the declared value was refused. -1 means it
	// was not measured.
	SoftLimitFloor int64 `json:"soft_limit_floor"`
	// InheritedSoft and InheritedHard are the limits the domain handed the
	// bounded process, so a reader can tell a bound that was already in force
	// from one this run installed.
	InheritedSoft int64 `json:"inherited_soft"`
	InheritedHard int64 `json:"inherited_hard"`

	Bounded BoundRun `json:"bounded"`
	Nested  BoundRun `json:"nested"`
	Control BoundRun `json:"control"`

	// SoftRaiseHardPreserved is what happened when a bounded process tried to
	// raise its own soft limit back to the hard limit it inherited.
	SoftRaiseHardPreserved string `json:"soft_raise_hard_preserved"`
	// SoftRaiseHardLowered is the same attempt after the hard limit was lowered
	// to the declared value. It is the matched control for the line above: if
	// both succeed, the refusal in the second is not attributable to the hard
	// limit.
	SoftRaiseHardLowered string `json:"soft_raise_hard_lowered"`

	// Error records why a measurement is missing rather than leaving a zero
	// value that reads like an observation.
	Error string `json:"error,omitempty"`
}

// Escape results, as reported by the escape process.
const (
	EscapeRaised       = "raised"
	EscapeNotAttempted = "not-attempted"
)

func escapeRefused(err error) string {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return "refused:" + errnoName(errno)
	}
	return "refused:" + err.Error()
}

// ------------------------------------------------------------ the matrix

// boundMatrix orchestrates the whole matrix from inside the probe domain. Every
// measurement is made by a separate process because three of the four bounds
// can end the process that hit them, and a measurement that dies with its
// measurer is not a measurement.
func boundMatrix(report *Report) []Attempt {
	self := os.Getenv(EnvSelf)
	if self == "" {
		return []Attempt{{
			Name:    "bound-matrix",
			Outcome: OutcomeInconclusive,
			Detail:  EnvSelf + " is not set, so no bound process could be started",
		}}
	}

	kinds := BoundKinds()
	measurements := make([]BoundMeasurement, len(kinds))
	var wg sync.WaitGroup
	for i, kind := range kinds {
		wg.Add(1)
		go func(i int, kind string) {
			defer wg.Done()
			measurements[i] = measureBound(self, kind)
		}(i, kind)
	}
	wg.Wait()
	report.Bounds = measurements

	out := make([]Attempt, 0, len(measurements))
	for _, m := range measurements {
		out = append(out, Attempt{
			Name:    "bound-measured:" + m.Kind,
			Target:  m.Resource,
			Outcome: measurementOutcome(m),
			Detail:  m.Error,
		})
	}
	return out
}

// measurementOutcome says whether the kind produced a usable measurement at
// all. It deliberately does not say whether the bound held: that reduction
// belongs to the caller, in package probe.
func measurementOutcome(m BoundMeasurement) string {
	if m.Error != "" || !m.Control.Ran {
		return OutcomeInconclusive
	}
	return OutcomeAllowed
}

// measureBound runs the three processes one kind needs and merges what they
// reported.
func measureBound(self, kind string) BoundMeasurement {
	declared, ceiling := boundBudget(kind)
	m := BoundMeasurement{
		Kind:           kind,
		Resource:       boundResourceName(kind),
		Declared:       declared,
		Ceiling:        ceiling,
		SoftLimitFloor: -1,
		InheritedSoft:  -1,
		InheritedHard:  -1,
	}

	bounded, err := runBoundProcess(self, kind, true, true)
	if err != nil {
		m.Error = "bounded run: " + err.Error()
	} else {
		m.Installed = bounded.Installed
		m.InstallErrno = bounded.InstallErrno
		m.SoftLimitFloor = bounded.SoftLimitFloor
		m.InheritedSoft = bounded.InheritedSoft
		m.InheritedHard = bounded.InheritedHard
		m.Bounded = bounded.Bounded
		m.Nested = bounded.Nested
	}

	control, err := runBoundProcess(self, kind, false, false)
	if err != nil {
		m.Error = appendError(m.Error, "control run: "+err.Error())
	} else {
		m.Control = control.Bounded
	}

	preserved, lowered, err := runEscapeProcess(self, kind)
	if err != nil {
		m.Error = appendError(m.Error, "escape run: "+err.Error())
		m.SoftRaiseHardPreserved = EscapeNotAttempted
		m.SoftRaiseHardLowered = EscapeNotAttempted
	} else {
		m.SoftRaiseHardPreserved = preserved
		m.SoftRaiseHardLowered = lowered
	}
	return m
}

func appendError(existing, next string) string {
	if existing == "" {
		return next
	}
	return existing + "; " + next
}

// runBoundProcess starts one stress process and reads the measurement back.
//
// A process that installed a hard bound may be killed by the kernel before it
// can report, so the wait status is read as a fallback: a death by signal is
// itself the refusal the probe was looking for.
func runBoundProcess(self, kind string, install, nest bool) (BoundMeasurement, error) {
	declared, ceiling := boundBudget(kind)
	cmd := exec.Command(self, "__inside", OpBoundStress) //nolint:gosec // self is the allowlisted agent path the harness passed in
	cmd.Env = append(os.Environ(),
		EnvBoundKind+"="+kind,
		EnvBoundDeclared+"="+strconv.FormatInt(declared, 10),
		EnvBoundCeiling+"="+strconv.FormatInt(ceiling, 10),
		EnvBoundInstall+"="+boolEnv(install),
		EnvBoundNest+"="+boolEnv(nest),
	)
	cmd.Stdin = bytes.NewReader(nil)
	cmd.Stderr = io.Discard
	out, runErr := cmd.Output()

	report, parseErr := ParseReport(string(out))
	if parseErr == nil && len(report.Bounds) == 1 {
		return report.Bounds[0], nil
	}
	if signal, ok := deathSignal(runErr); ok {
		// The stress process was killed by the bound it installed. That is a
		// refusal, and saying so is more truthful than reporting the missing
		// report as a broken probe.
		return BoundMeasurement{
			Kind:     kind,
			Resource: boundResourceName(kind),
			Declared: declared,
			Ceiling:  ceiling,
			Bounded: BoundRun{
				Ran:     true,
				Reached: -1,
				Refused: true,
				Refusal: signal,
				Detail:  "the process was killed before it could report; the quantity it reached is unknown",
			},
			SoftLimitFloor: -1,
			InheritedSoft:  -1,
			InheritedHard:  -1,
		}, nil
	}
	if runErr != nil {
		return BoundMeasurement{}, fmt.Errorf("%s: %w", kind, runErr)
	}
	return BoundMeasurement{}, fmt.Errorf("%s: %w", kind, parseErr)
}

// deathSignal names the signal that killed a process, when one did.
func deathSignal(err error) (string, bool) {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return "", false
	}
	status, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok || !status.Signaled() {
		return "", false
	}
	return signalName(status.Signal()), true
}

func signalName(sig syscall.Signal) string {
	switch sig {
	case syscall.SIGXCPU:
		return "SIGXCPU"
	case syscall.SIGKILL:
		return "SIGKILL"
	case syscall.SIGSEGV:
		return "SIGSEGV"
	case syscall.SIGBUS:
		return "SIGBUS"
	case syscall.SIGABRT:
		return "SIGABRT"
	default:
		return fmt.Sprintf("signal(%d)", int(sig))
	}
}

func runEscapeProcess(self, kind string) (preserved, lowered string, err error) {
	cmd := exec.Command(self, "__inside", OpBoundEscape) //nolint:gosec // self is the allowlisted agent path the harness passed in
	declared, _ := boundBudget(kind)
	cmd.Env = append(os.Environ(),
		EnvBoundKind+"="+kind,
		EnvBoundDeclared+"="+strconv.FormatInt(declared, 10),
	)
	cmd.Stdin = bytes.NewReader(nil)
	cmd.Stderr = io.Discard
	out, runErr := cmd.Output()
	report, parseErr := ParseReport(string(out))
	if parseErr != nil {
		if runErr != nil {
			return "", "", runErr
		}
		return "", "", parseErr
	}
	return report.Values["soft_raise_hard_preserved"], report.Values["soft_raise_hard_lowered"], nil
}

func boolEnv(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func boundBudget(kind string) (declared, ceiling int64) {
	switch kind {
	case BoundCPU:
		return declaredCPUMillis, ceilingCPUMillis
	case BoundAddressSpace, BoundDataSegment:
		return declaredMemoryBytes, ceilingMemoryBytes
	case BoundProcessCount:
		return declaredProcessCount, ceilingProcessCount
	default:
		return 0, 0
	}
}

// rlimitValue converts a declared bound into the unit the POSIX resource takes.
//
// It exists because RLIMIT_CPU is the one resource whose unit is not the unit
// the bound is measured in: it counts whole seconds, while CPU consumption is
// only observable at millisecond resolution through getrusage. Passing the
// millisecond figure straight through would install a limit a thousand times
// larger than the declared one and report a bound that never bound anything.
func rlimitValue(kind string, declared int64) uint64 {
	if kind != BoundCPU {
		return uint64(declared) //nolint:gosec // declared is a non-negative harness constant
	}
	seconds := declared / 1000
	if declared%1000 != 0 {
		seconds++
	}
	if seconds < 1 {
		seconds = 1
	}
	return uint64(seconds) //nolint:gosec // bounded below by 1
}

// boundResource maps a kind to the POSIX resource that expresses it. The second
// result is false for a kind this platform has no resource for, which is a real
// answer and not an error.
func boundResource(kind string) (int, bool) {
	switch kind {
	case BoundCPU:
		return syscall.RLIMIT_CPU, true
	case BoundAddressSpace:
		return syscall.RLIMIT_AS, true
	case BoundDataSegment:
		return syscall.RLIMIT_DATA, true
	case BoundProcessCount:
		return rlimitNPROC, true
	default:
		return 0, false
	}
}

func boundResourceName(kind string) string {
	switch kind {
	case BoundCPU:
		return "RLIMIT_CPU"
	case BoundAddressSpace:
		return "RLIMIT_AS"
	case BoundDataSegment:
		return "RLIMIT_DATA"
	case BoundProcessCount:
		return nprocResourceName
	default:
		return "unknown"
	}
}

// --------------------------------------------------------- the stress process

// boundStress is one stress process: it optionally installs the declared bound,
// optionally starts one descendant that inherits it, and then tries to exceed
// the bound.
func boundStress(report *Report) []Attempt {
	kind := os.Getenv(EnvBoundKind)
	declared := envInt64(EnvBoundDeclared, 0)
	ceiling := envInt64(EnvBoundCeiling, 0)
	install := os.Getenv(EnvBoundInstall) == "1"
	nest := os.Getenv(EnvBoundNest) == "1"

	m := BoundMeasurement{
		Kind:           kind,
		Resource:       boundResourceName(kind),
		Declared:       declared,
		Ceiling:        ceiling,
		SoftLimitFloor: -1,
		InheritedSoft:  -1,
		InheritedHard:  -1,
	}

	resource, known := boundResource(kind)
	if !known {
		m.Error = "no resource is defined for kind " + strconv.Quote(kind)
		report.Bounds = []BoundMeasurement{m}
		return []Attempt{{Name: "install-bound", Outcome: OutcomeInconclusive, Detail: m.Error}}
	}

	var original syscall.Rlimit
	if err := syscall.Getrlimit(resource, &original); err != nil {
		m.Error = "getrlimit: " + err.Error()
		report.Bounds = []BoundMeasurement{m}
		return []Attempt{{Name: "install-bound", Outcome: OutcomeInconclusive, Detail: m.Error}}
	}
	m.InheritedSoft = clampToInt64(original.Cur)
	m.InheritedHard = clampToInt64(original.Max)

	var out []Attempt
	if install {
		value := rlimitValue(kind, declared)
		installErr := syscall.Setrlimit(resource, &syscall.Rlimit{Cur: value, Max: original.Max})
		m.Installed = installErr == nil
		if installErr != nil {
			m.InstallErrno = errnoOf(installErr)
			// The kernel refused the value, not the call. Find the lowest value
			// it does accept, so the report can say how far above a usable
			// build budget this platform's floor sits instead of only that the
			// declared bound failed.
			m.SoftLimitFloor = measureSoftLimitFloor(resource, value, original)
		}
		out = append(out, attempt("install-bound", m.Resource, installErr))
	} else {
		out = append(out, Attempt{
			Name:    "install-bound",
			Target:  m.Resource,
			Outcome: OutcomeInconclusive,
			Detail:  "this is the unbounded control run; no bound was installed",
		})
	}

	// The descendant is started after the bound is installed so it inherits it.
	// Starting it first would measure a process that was never in the domain's
	// budget, which is a different question.
	var nested *nestedStress
	if nest {
		var nestedErr error
		nested, nestedErr = startNestedStress(kind, declared, ceiling)
		if nestedErr != nil {
			// The descendant could not be created. That is not a broken probe:
			// on a bound that binds process creation it is the bound itself
			// refusing, so the refusal is recorded even though no nested
			// measurement exists.
			m.Nested = BoundRun{Ran: false, Reached: -1, Refused: errnoOf(nestedErr) != "",
				Refusal: errnoOf(nestedErr), Detail: nestedErr.Error()}
			out = append(out, attempt("start-nested-descendant", kind, nestedErr))
		}
	}

	m.Bounded = stress(kind, ceiling)
	out = append(out, Attempt{
		Name:    "stress:" + kind,
		Target:  m.Resource,
		Outcome: runOutcome(m.Bounded),
		Errno:   m.Bounded.Refusal,
		Detail:  m.Bounded.Detail,
	})

	if nested != nil {
		m.Nested = nested.collect()
		out = append(out, Attempt{
			Name:    "nested-descendant-stress:" + kind,
			Target:  m.Resource,
			Outcome: runOutcome(m.Nested),
			Errno:   m.Nested.Refusal,
			Detail:  m.Nested.Detail,
		})
	}

	// Put the inherited limit back. It cannot change what was already observed,
	// and the process exits immediately after, but a probe that leaves a
	// lowered limit behind in a reused process would poison the next one.
	_ = syscall.Setrlimit(resource, &original)

	report.Bounds = []BoundMeasurement{m}
	return out
}

func runOutcome(r BoundRun) string {
	switch {
	case !r.Ran:
		return OutcomeInconclusive
	case r.Refused:
		return OutcomeDenied
	default:
		return OutcomeAllowed
	}
}

// nestedStress is a descendant that was started but not yet waited for. It
// holds its own stdout pipe so the parent can burn its own budget while the
// descendant burns the one it inherited, which is the whole point: two
// concurrent claims on a single declared bound.
type nestedStress struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

func startNestedStress(kind string, declared, ceiling int64) (*nestedStress, error) {
	self := os.Getenv(EnvSelf)
	if self == "" {
		return nil, fmt.Errorf("%s is not set", EnvSelf)
	}
	cmd := exec.Command(self, "__inside", OpBoundStress) //nolint:gosec // self is the allowlisted agent path the harness passed in
	cmd.Env = append(os.Environ(),
		EnvBoundKind+"="+kind,
		EnvBoundDeclared+"="+strconv.FormatInt(declared, 10),
		EnvBoundCeiling+"="+strconv.FormatInt(ceiling, 10),
		// The descendant must not install anything: the whole question is what
		// the limit it inherited allows it to reach.
		EnvBoundInstall+"=0",
		EnvBoundNest+"=0",
	)
	cmd.Stdin = bytes.NewReader(nil)
	cmd.Stderr = io.Discard
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &nestedStress{cmd: cmd, stdout: stdout}, nil
}

func (n *nestedStress) collect() BoundRun {
	data, _ := io.ReadAll(n.stdout)
	waitErr := n.cmd.Wait()

	report, parseErr := ParseReport(string(data))
	if parseErr == nil && len(report.Bounds) == 1 {
		return report.Bounds[0].Bounded
	}
	if signal, ok := deathSignal(waitErr); ok {
		return BoundRun{
			Ran:     true,
			Reached: -1,
			Refused: true,
			Refusal: signal,
			Detail:  "the descendant was killed before it could report",
		}
	}
	detail := "the descendant produced no measurement"
	if parseErr != nil {
		detail = parseErr.Error()
	}
	return BoundRun{Ran: false, Reached: -1, Detail: detail}
}

// stress performs the attempt for one kind.
//
// It takes only the ceiling, not the declared bound: a stress must not know
// what it is being measured against, or it could stop exactly where the bound
// would have stopped it and the two would be indistinguishable.
func stress(kind string, ceiling int64) BoundRun {
	switch kind {
	case BoundCPU:
		return stressCPU(ceiling)
	case BoundAddressSpace, BoundDataSegment:
		return stressMemory(ceiling)
	case BoundProcessCount:
		return stressProcesses(ceiling)
	default:
		return BoundRun{Ran: false, Reached: -1, Detail: "no stress is defined for " + kind}
	}
}

// --------------------------------------------------------------- CPU time

// cpuWallGuard bounds the burn by wall time as well as by CPU time. A host that
// delivered neither SIGXCPU nor CPU accounting would otherwise spin forever, and
// a probe that can hang is not usable evidence.
const cpuWallGuard = 20 * time.Second

// stressCPU burns CPU until the kernel refuses, or until the ceiling is
// reached. The refusal is SIGXCPU, which RLIMIT_CPU delivers at the soft limit
// and which is caught here so the process can report the CPU time it had
// actually accumulated when the bound bound.
func stressCPU(ceilingMS int64) BoundRun {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGXCPU)
	defer signal.Stop(sig)

	deadline := time.Now().Add(cpuWallGuard)
	var sink float64
	var run BoundRun
	for i := 0; ; i++ {
		sink += float64(i%7) * 1.000001
		if i%(1<<18) != 0 {
			continue
		}
		select {
		case <-sig:
			run = BoundRun{Ran: true, Reached: cpuMillis(), Refused: true, Refusal: "SIGXCPU",
				Detail: "the kernel signalled the CPU-time soft limit"}
		default:
			used := cpuMillis()
			switch {
			case used < 0:
				run = BoundRun{Ran: false, Reached: -1,
					Detail: "getrusage is unavailable, so CPU time could not be measured"}
			case used >= ceilingMS:
				run = BoundRun{Ran: true, Reached: used,
					Detail: fmt.Sprintf("burned %d CPU-milliseconds with no refusal and stopped at the ceiling", used)}
			case time.Now().After(deadline):
				run = BoundRun{Ran: true, Reached: used,
					Detail: fmt.Sprintf("stopped by the %s wall-clock guard after %d CPU-milliseconds", cpuWallGuard, used)}
			}
		}
		if run.Detail != "" {
			break
		}
	}

	// The sum is read here so the arithmetic above is not a dead store the
	// compiler may drop. It cannot fire: a running total of non-negative terms
	// is never NaN, and a burn loop that had been optimised away would consume
	// no CPU and quietly turn every CPU measurement into a wall-clock timeout.
	if math.IsNaN(sink) {
		return BoundRun{Ran: false, Reached: -1, Detail: "the burn loop did not accumulate"}
	}
	return run
}

// cpuMillis is the CPU time this process has consumed across all its threads.
func cpuMillis() int64 {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return -1
	}
	return timevalMillis(ru.Utime) + timevalMillis(ru.Stime)
}

func timevalMillis(tv syscall.Timeval) int64 {
	return tv.Sec*1000 + int64(tv.Usec)/1000
}

// ----------------------------------------------------------------- memory

// memoryChunk is the size of one anonymous mapping. It is large enough that the
// loop is short and small enough that the refusal point is reported with useful
// resolution.
const memoryChunk = 16 << 20

// stressMemory maps anonymous memory and touches a page of each mapping, so the
// pages are committed rather than merely reserved, until the kernel refuses or
// the ceiling is reached.
func stressMemory(ceilingBytes int64) BoundRun {
	var mappings [][]byte
	defer func() {
		for _, m := range mappings {
			_ = syscall.Munmap(m)
		}
	}()

	var mapped int64
	for mapped < ceilingBytes {
		chunk, err := syscall.Mmap(-1, 0, memoryChunk,
			syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_ANON|syscall.MAP_PRIVATE)
		if err != nil {
			return BoundRun{Ran: true, Reached: mapped, Refused: true, Refusal: errnoOf(err),
				Detail: fmt.Sprintf("the kernel refused a mapping after %d bytes", mapped)}
		}
		chunk[0] = 1
		mappings = append(mappings, chunk)
		mapped += memoryChunk
	}
	return BoundRun{Ran: true, Reached: mapped,
		Detail: fmt.Sprintf("mapped and touched %d bytes with no refusal and stopped at the ceiling", mapped)}
}

// ---------------------------------------------------------- process count

// stressProcesses starts descendants until the kernel refuses or the ceiling is
// reached, then stops every one it started.
func stressProcesses(ceiling int64) BoundRun {
	self := os.Getenv(EnvSelf)
	if self == "" {
		return BoundRun{Ran: false, Reached: -1, Detail: EnvSelf + " is not set, so no descendant could be started"}
	}
	marker := os.Getenv(EnvMarker)

	var started []*exec.Cmd
	defer func() {
		for _, cmd := range started {
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			_ = cmd.Wait()
		}
	}()

	var count int64
	for count < ceiling {
		cmd := exec.Command(self, "__inside", OpDetachedChild) //nolint:gosec // self is the allowlisted agent path the harness passed in
		cmd.Env = append(os.Environ(),
			EnvMarker+"="+marker+".proc-"+strconv.FormatInt(count, 10),
			EnvHold+"=2",
		)
		pipeStdio(cmd)
		if err := cmd.Start(); err != nil {
			return BoundRun{Ran: true, Reached: count, Refused: true, Refusal: errnoOf(err),
				Detail: fmt.Sprintf("the kernel refused descendant %d: %v", count+1, err)}
		}
		started = append(started, cmd)
		count++
	}

	// Wait for the last descendant to announce itself before the deferred
	// cleanup above tears them all down. A successful fork already proves the
	// process exists, but a count backed by processes that never reached their
	// own code is weaker evidence than one backed by processes that did, and
	// the wait costs a few milliseconds.
	if count > 0 && marker != "" {
		waitForFile(marker+".proc-"+strconv.FormatInt(count-1, 10), 2*time.Second)
	}
	return BoundRun{Ran: true, Reached: count,
		Detail: fmt.Sprintf("started %d descendants with no refusal and stopped at the ceiling", count)}
}

// --------------------------------------------------------- the escape probe

// boundEscape asks whether a domain member can undo the bound it was given.
//
// A soft limit is only a floor a process agreed to; POSIX lets any process
// raise its own soft limit back up to its hard limit. A supervisor that lowers
// only the soft limit has therefore bounded nothing that the bounded process
// does not consent to. The matched control lowers the hard limit too, so a
// refusal in the second case is attributable to the hard limit and not to some
// other reason the call might fail.
func boundEscape(report *Report) []Attempt {
	kind := os.Getenv(EnvBoundKind)
	declared := envInt64(EnvBoundDeclared, 0)
	report.Values["bound_kind"] = kind
	report.Values["soft_raise_hard_preserved"] = EscapeNotAttempted
	report.Values["soft_raise_hard_lowered"] = EscapeNotAttempted

	resource, known := boundResource(kind)
	if !known {
		return []Attempt{{Name: "raise-soft-limit", Outcome: OutcomeInconclusive,
			Detail: "no resource is defined for kind " + strconv.Quote(kind)}}
	}
	return escapeAttempts(resource, boundResourceName(kind), rlimitValue(kind, declared), report)
}

// escapeAttempts is the escape measurement itself, against a resource the
// caller names.
//
// It is separated from the environment parsing above so it can be exercised
// against a resource whose hard limit is harmless to lower — the last step here
// is irreversible for the process that runs it, which is why the real
// measurement always happens in a process with nothing left to measure.
func escapeAttempts(resource int, resourceName string, value uint64, report *Report) []Attempt {
	var original syscall.Rlimit
	if err := syscall.Getrlimit(resource, &original); err != nil {
		return []Attempt{{Name: "raise-soft-limit", Outcome: OutcomeInconclusive, Detail: err.Error()}}
	}
	report.Values["inherited_soft"] = strconv.FormatInt(clampToInt64(original.Cur), 10)
	report.Values["inherited_hard"] = strconv.FormatInt(clampToInt64(original.Max), 10)

	var out []Attempt

	// A soft limit lowered below the hard limit. The escape is to put it back.
	if err := syscall.Setrlimit(resource, &syscall.Rlimit{Cur: value, Max: original.Max}); err != nil {
		out = append(out, Attempt{Name: "lower-soft-limit-only", Target: resourceName,
			Outcome: OutcomeInconclusive, Errno: errnoOf(err),
			Detail: "the kernel refused the declared soft limit, so there was nothing to escape from"})
		return out
	}
	out = append(out, Attempt{Name: "lower-soft-limit-only", Target: resourceName, Outcome: OutcomeAllowed})

	raiseErr := syscall.Setrlimit(resource, &syscall.Rlimit{Cur: original.Max, Max: original.Max})
	if raiseErr == nil {
		report.Values["soft_raise_hard_preserved"] = EscapeRaised
	} else {
		report.Values["soft_raise_hard_preserved"] = escapeRefused(raiseErr)
	}
	out = append(out, attempt("raise-soft-limit-when-hard-preserved", resourceName, raiseErr))

	// The matched control. Lowering the hard limit is irreversible for this
	// process, which is exactly why it is done last and in a process that has
	// nothing left to measure.
	if err := syscall.Setrlimit(resource, &syscall.Rlimit{Cur: value, Max: value}); err != nil {
		report.Values["soft_raise_hard_lowered"] = "unavailable:" + errnoOf(err)
		out = append(out, Attempt{Name: "lower-hard-limit", Target: resourceName,
			Outcome: OutcomeInconclusive, Errno: errnoOf(err),
			Detail: "the kernel refused the declared hard limit, so the matched control could not be run"})
		return out
	}
	out = append(out, Attempt{Name: "lower-hard-limit", Target: resourceName, Outcome: OutcomeAllowed})

	controlErr := syscall.Setrlimit(resource, &syscall.Rlimit{Cur: original.Max, Max: original.Max})
	if controlErr == nil {
		report.Values["soft_raise_hard_lowered"] = EscapeRaised
	} else {
		report.Values["soft_raise_hard_lowered"] = escapeRefused(controlErr)
	}
	out = append(out, attempt("raise-soft-limit-when-hard-lowered", resourceName, controlErr))
	return out
}

// ------------------------------------------------------------------ helpers

// measureSoftLimitFloor finds the lowest soft limit this kernel accepts for a
// resource, given a value it already refused. It restores the original limit
// before returning; a floor measurement that changed the limit it was measuring
// under would be worthless.
//
// It returns -1 when no value below the hard limit is accepted, or when the
// search cannot be bracketed.
func measureSoftLimitFloor(resource int, refused uint64, original syscall.Rlimit) int64 {
	high := original.Max
	if high <= refused {
		return -1
	}
	// The hard limit must itself be acceptable, otherwise there is no bracket.
	if err := syscall.Setrlimit(resource, &syscall.Rlimit{Cur: high, Max: original.Max}); err != nil {
		return -1
	}
	if err := syscall.Setrlimit(resource, &original); err != nil {
		return -1
	}

	low := refused
	const resolution = 1 << 20
	for i := 0; i < 64 && high-low > resolution; i++ {
		mid := low + (high-low)/2
		if err := syscall.Setrlimit(resource, &syscall.Rlimit{Cur: mid, Max: original.Max}); err == nil {
			_ = syscall.Setrlimit(resource, &original)
			high = mid
			continue
		}
		low = mid
	}
	_ = syscall.Setrlimit(resource, &original)
	return clampToInt64(high)
}

// clampToInt64 renders an rlimit value as a signed integer for the report.
// RLIM_INFINITY is already the largest signed value on the platforms this
// prototype builds for, so the clamp only guards a future one where it is not.
func clampToInt64(v uint64) int64 {
	const maxInt64 = uint64(1)<<63 - 1
	if v > maxInt64 {
		return int64(maxInt64)
	}
	return int64(v) //nolint:gosec // bounded above
}

func errnoOf(err error) string {
	if err == nil {
		return ""
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errnoName(errno)
	}
	return ""
}

func envInt64(key string, fallback int64) int64 {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}
