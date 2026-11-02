package envprofile

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/homelease"
)

type homeLeaseRaceInput struct {
	Home, Profile, Digest, Launch, XDG string
	Native                             map[string]string
}

// This child is a real second process using the same OS manager lock and the
// production materializer authority. It neither mocks the lock nor starts a
// provider/session manager. Ordinary package runs have no child input.
func TestHomeLeaseRaceProcess(t *testing.T) {
	if os.Getenv("CURATOR_TEST_HOME_LEASE_PROCESS") != "1" {
		return
	}
	var in homeLeaseRaceInput
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		t.Fatal(err)
	}
	req := ResolveRequest{Home: in.Home, Profile: in.Profile, EnvID: envregistry.CodexCLI, Machine: envregistry.DefaultMachineConfig(), LaunchDir: in.Launch, OperatorXDG: in.XDG, Detect: func(envregistry.Adapter) string { return "unknown" }, NativeHomeOf: func(id string) (string, error) { return in.Native[id], nil }}
	m, err := HomeLeaseManager(req, nil)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("ready")
	r := homelease.Request{Environment: envregistry.CodexCLI, Profile: in.Profile, LockSHA256: in.Digest, RequestID: "racing-process", Holder: homelease.Holder{Kind: "prepared", Launcher: "curator-run", ManagerID: "race-manager", ManagerRegisteredAt: "2026-10-27T00:00:00Z", RequestID: "racing-process"}}
	out, reason, err := m.Acquire(context.Background(), r)
	if err != nil || !out.Changed || reason != "acquired" {
		t.Fatalf("child admission: %+v %s %v", out, reason, err)
	}
	fmt.Println("acquired")
}

// HL02/W01: a real acquire cannot enter between a repair's preflight and last
// write. After that same acquisition releases, the child sees the repaired
// generation and reserves it; a subsequent drift repair refuses unchanged.
func TestHomeLeaseRepairAcquireProcessRace(t *testing.T) {
	isolateHomeLeaseWriters(t)
	fx, _ := captureFixture(t)
	op, err := beginOperationGuarded(fx.home, nil, homeLeaseProfileTargets(fx.home, fx.profile, envregistry.CodexCLI))
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = op.close()
		}
	}()
	path := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI), "AGENTS.md")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("repair-race-drift"), 0600); err != nil {
		t.Fatal(err)
	}
	in := homeLeaseRaceInput{Home: fx.home, Profile: fx.profile, Digest: strings.TrimPrefix(fx.lockHash, "sha256:"), Launch: fx.launch, XDG: fx.xdg, Native: fx.native}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestHomeLeaseRaceProcess$", "-test.timeout=15s")
	// Native fixture paths travel in the closed stdin payload. The child is
	// an ambient test process, so do not forward the parent's per-test provider
	// overrides across the neutral helper's fresh process boundary.
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		switch name {
		case "CODEX_HOME", "CLAUDE_CONFIG_DIR", "PI_CODING_AGENT_DIR", "XDG_DATA_HOME", "XDG_STATE_HOME", "XDG_CACHE_HOME":
			continue
		}
		cmd.Env = append(cmd.Env, entry)
	}
	cmd.Env = append(cmd.Env, "CURATOR_TEST_HOME_LEASE_PROCESS=1")
	cmd.Stdin = bytes.NewReader(raw)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	// CommandContext owns only this recorded child PID and kills it on timeout.
	lines := make(chan string, 4)
	go func() {
		scan := bufio.NewScanner(stdout)
		for scan.Scan() {
			lines <- scan.Text()
		}
		close(lines)
	}()
	select {
	case line := <-lines:
		if line != "ready" {
			t.Fatalf("child readiness: %q %s", line, stderr.String())
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	select {
	case line := <-lines:
		t.Fatalf("child crossed the held repair lock: %q", line)
	case <-time.After(40 * time.Millisecond):
	}
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	req.borrowedOperation = op
	if _, err := Resolve(req); err != nil {
		t.Fatalf("borrowed repair: %v", err)
	}
	select {
	case line := <-lines:
		t.Fatalf("child admitted before last repair write: %q", line)
	default:
	}
	if err := op.close(); err != nil {
		t.Fatal(err)
	}
	closed = true
	select {
	case line := <-lines:
		if line != "acquired" {
			t.Fatalf("child result: %q %s", line, stderr.String())
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	// Drain the test harness's final PASS line before joining the child.
	for range lines {
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("child PID %d: %v %s", cmd.Process.Pid, err, stderr.String())
	}
	m, err := HomeLeaseManager(fx.request(envregistry.CodexCLI), nil)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := m.List(context.Background(), "", "")
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows: %+v %v", rows, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("post-acquire-drift"), 0600); err != nil {
		t.Fatal(err)
	}
	before, _ := captureFixtureState(t, fx)
	req.borrowedOperation = nil
	_, err = Resolve(req)
	requireHomeLeaseRefusal(t, err, rows[0].LeaseID)
	after, _ := captureFixtureState(t, fx)
	if !bytes.Equal(before, after) {
		t.Fatal("racing refusal changed preimage")
	}
}
