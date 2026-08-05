package crossmanager

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// SampleCommand is one read-only host observation command.
type SampleCommand struct {
	Executable string
	Args       []string
}

// CommandBoundaryProbe detects new process and network rows between samples.
// Callers can replace the native commands with platform instrumentation that
// is stricter for their CI environment.
type CommandBoundaryProbe struct {
	Processes SampleCommand
	Network   SampleCommand
	Ignore    []string
	Interval  time.Duration
}

type boundaryToken struct {
	probe    CommandBoundaryProbe
	baseline boundarySnapshot
	cancel   context.CancelFunc
	done     chan struct{}
	mu       sync.Mutex
	seen     boundarySnapshot
	err      error
}

type boundarySnapshot struct {
	processes map[string]struct{}
	network   map[string]struct{}
}

// Begin captures a baseline and starts continuous process and network sampling.
func (probe CommandBoundaryProbe) Begin(ctx context.Context, _ Invocation) (any, error) {
	baseline, err := probe.sample(ctx)
	if err != nil {
		return nil, err
	}
	interval := probe.Interval
	if interval <= 0 {
		interval = 20 * time.Millisecond
	}
	sampleCtx, cancel := context.WithCancel(ctx)
	token := &boundaryToken{
		probe: probe, baseline: baseline, cancel: cancel, done: make(chan struct{}),
		seen: boundarySnapshot{processes: make(map[string]struct{}), network: make(map[string]struct{})},
	}
	go func() {
		defer close(token.done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-sampleCtx.Done():
				return
			case <-ticker.C:
				current, err := probe.sample(sampleCtx)
				token.mu.Lock()
				if err != nil {
					if sampleCtx.Err() == nil && token.err == nil {
						token.err = err
					}
				} else {
					mergeDifference(token.seen.processes, current.processes, baseline.processes, probe.Ignore)
					mergeDifference(token.seen.network, current.network, baseline.network, probe.Ignore)
				}
				token.mu.Unlock()
			}
		}
	}()
	return token, nil
}

// End stops sampling and returns every process and network row first observed
// after Begin.
func (probe CommandBoundaryProbe) End(ctx context.Context, token any) (BoundaryEvents, error) {
	state, ok := token.(*boundaryToken)
	if !ok {
		return BoundaryEvents{}, fmt.Errorf("invalid boundary snapshot token")
	}
	state.cancel()
	<-state.done
	after, err := probe.sample(ctx)
	if err != nil {
		return BoundaryEvents{}, err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	mergeDifference(state.seen.processes, after.processes, state.baseline.processes, probe.Ignore)
	mergeDifference(state.seen.network, after.network, state.baseline.network, probe.Ignore)
	if state.err != nil {
		return BoundaryEvents{}, state.err
	}
	return BoundaryEvents{
		UnexpectedProcesses: sortedSet(state.seen.processes),
		NetworkConnections:  sortedSet(state.seen.network),
	}, nil
}

func (probe CommandBoundaryProbe) sample(ctx context.Context) (boundarySnapshot, error) {
	processes, err := sampleLines(ctx, probe.Processes)
	if err != nil {
		return boundarySnapshot{}, fmt.Errorf("sample processes: %w", err)
	}
	network, err := sampleLines(ctx, probe.Network)
	if err != nil {
		return boundarySnapshot{}, fmt.Errorf("sample network: %w", err)
	}
	return boundarySnapshot{processes: processes, network: network}, nil
}

func sampleLines(ctx context.Context, sample SampleCommand) (map[string]struct{}, error) {
	if sample.Executable == "" {
		return nil, fmt.Errorf("sample executable is required")
	}
	// #nosec G204 -- the probe command is explicit runner configuration and is
	// executed directly without shell interpretation.
	payload, err := exec.CommandContext(ctx, sample.Executable, sample.Args...).Output()
	if err != nil {
		return nil, err
	}
	result := make(map[string]struct{})
	for _, line := range strings.Split(strings.ReplaceAll(string(payload), "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			result[line] = struct{}{}
		}
	}
	return result, nil
}

func setDifference(after, before map[string]struct{}, ignore []string) []string {
	result := make([]string, 0)
	for line := range after {
		if _, existed := before[line]; existed || containsAny(line, ignore) {
			continue
		}
		result = append(result, line)
	}
	sort.Strings(result)
	return result
}

func mergeDifference(target, after, before map[string]struct{}, ignore []string) {
	for _, line := range setDifference(after, before, ignore) {
		target[line] = struct{}{}
	}
}

func sortedSet(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func containsAny(value string, fragments []string) bool {
	for _, fragment := range fragments {
		if fragment != "" && strings.Contains(value, fragment) {
			return true
		}
	}
	return false
}

// NativeBoundaryProbe supplies read-only process/network samplers for the two
// required qualification platforms.
func NativeBoundaryProbe() (BoundaryProbe, error) {
	switch runtime.GOOS {
	case "darwin":
		return CommandBoundaryProbe{
			Processes: SampleCommand{Executable: "ps", Args: []string{"-axo", "pid=,ppid=,command="}},
			Network:   SampleCommand{Executable: "lsof", Args: []string{"-nP", "-iTCP", "-sTCP:ESTABLISHED"}},
			Ignore:    []string{"ps -axo", "lsof -nP"},
		}, nil
	case "windows":
		return CommandBoundaryProbe{
			Processes: SampleCommand{Executable: "powershell.exe", Args: []string{"-NoProfile", "-NonInteractive", "-Command", "Get-CimInstance Win32_Process | Sort-Object ProcessId | ForEach-Object { \"$($_.ProcessId) $($_.ParentProcessId) $($_.CommandLine)\" }"}},
			Network:   SampleCommand{Executable: "powershell.exe", Args: []string{"-NoProfile", "-NonInteractive", "-Command", "Get-NetTCPConnection -State Established | Sort-Object OwningProcess,RemoteAddress,RemotePort | ForEach-Object { \"$($_.OwningProcess) $($_.RemoteAddress):$($_.RemotePort)\" }"}},
			Ignore:    []string{"Get-CimInstance Win32_Process", "Get-NetTCPConnection"},
		}, nil
	default:
		return nil, fmt.Errorf("native boundary probe is unsupported on %s", runtime.GOOS)
	}
}
