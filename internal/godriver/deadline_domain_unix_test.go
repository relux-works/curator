//go:build darwin || linux

package godriver

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The runaway real worker acknowledges two live tool descendants before its
// deadline. After timeout and teardown, releasing them cannot publish a survival
// marker. The normal completion control uses the same causal barrier elsewhere.
func TestBuildDeadlineTearsDownDescendants(t *testing.T) {
	fixture := newSnapshotFixture(t)
	barrier := t.TempDir()
	pidFile := filepath.Join(barrier, "pids")
	t.Cleanup(func() { _ = os.WriteFile(filepath.Join(barrier, "release"), nil, 0600) })
	fixture.start(stubScript{ListStdout: string(encodePackages(t, fixture.rootPackage())), Artifact: "artifact", BuildSleepMS: 60000, SpawnChildren: 2, SpawnPidFile: pidFile, SpawnBarrier: barrier})
	result, err := Build(t.Context(), fixture.request(ResourceLimits{Timeout: 5 * time.Second}))
	if DiagnosticCode(err) != "process_timeout" || result.Artifact.StagedPath != "" {
		t.Fatalf("runaway result=%+v err=%v", result, err)
	}
	raw, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("descendant readiness missing: %v", err)
	}
	pids := strings.Fields(string(raw))
	if len(pids) != 2 {
		t.Fatalf("descendant identities=%v", pids)
	}
	if err := os.WriteFile(filepath.Join(barrier, "release"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	guard, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	for _, raw := range pids {
		pid, err := strconv.Atoi(raw)
		if err != nil {
			t.Fatal(err)
		}
		for processAlive(pid) && guard.Err() == nil {
			time.Sleep(10 * time.Millisecond)
		}
		if processAlive(pid) {
			t.Fatalf("descendant %d survived timeout teardown", pid)
		}
		if _, err := os.Stat(filepath.Join(barrier, "survived-"+raw)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("timed-out descendant published survival marker: %v", err)
		}
	}
}
