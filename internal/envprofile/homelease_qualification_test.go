package envprofile

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/managerlock"
)

// W08: a borrowed real lock cannot bypass refusal or invoke the legacy
// continuation. After an explicit fixture break, the same joined operation
// reaches its continuation once and publishes its real profile/native phase.
func TestHomeLeaseBorrowedGlobalContinuation(t *testing.T) {
	fx, m, lease := leasedManagedFixture(t)
	called := 0
	write := func() error {
		locks, err := managerlock.New(fx.home)
		if err != nil {
			return err
		}
		lock, err := locks.AcquireHomeOnly(context.Background(), false)
		if err != nil {
			return err
		}
		defer func() { _ = lock.Close() }()
		_, err = GlobalSkillTransaction(fx.home, Policy{}, envregistry.DefaultMachineConfig(), nil, fx.request(envregistry.CodexCLI).NativeHomeOf, lock, func() error {
			called++
			return os.WriteFile(filepath.Join(fx.home, "legacy-continuation-witness"), []byte("completed"), 0600)
		})
		return err
	}
	before, _ := captureFixtureState(t, fx)
	requireHomeLeaseRefusal(t, write(), lease.LeaseID)
	after, _ := captureFixtureState(t, fx)
	if called != 0 || !bytes.Equal(before, after) {
		t.Fatal("refused borrowed phase changed state or called continuation")
	}
	breakFixtureLease(t, m, lease)
	if err := write(); err != nil {
		t.Fatal(err)
	}
	if called != 1 {
		t.Fatalf("continuation count: %d", called)
	}
	if b, err := os.ReadFile(filepath.Join(fx.home, "legacy-continuation-witness")); err != nil || string(b) != "completed" {
		t.Fatal("control did not publish continuation")
	}
}

// W11: a native MCP batch for Claude and Codex refuses while the same profile
// has a leased Muse home, including when entered as repair. After release,
// the same operation installs actual native configuration and ownership.
// A legacy MCP symlink is separately refused by the existing ownership gate.
func TestHomeLeaseNativeMCPBatch(t *testing.T) {
	for _, mode := range []string{"install", "repair"} {
		t.Run(mode, func(t *testing.T) {
			isolateHomeLeaseWriters(t)
			fx, req := nativeWriterFixture(t)
			seedLiveNativeCredentials(t, fx)
			fx.native[envregistry.Muse] = t.TempDir()
			if err := os.WriteFile(filepath.Join(fx.native[envregistry.Muse], "auth.json"), []byte("synthetic-native\n"), 0600); err != nil {
				t.Fatal(err)
			}
			provision := fx.request(envregistry.Muse)
			provision.Repair = true
			if _, err := Resolve(provision); err != nil {
				t.Fatal(err)
			}
			m, lease := leaseManagedFixture(t, fx, envregistry.Muse)
			req.Environments = []string{envregistry.ClaudeCode, envregistry.CodexCLI}
			req.Mode = mode
			before, _ := captureFixtureState(t, fx)
			result, err := ProfileMCP(req)
			requireHomeLeaseRefusal(t, err, lease.LeaseID)
			after, _ := captureFixtureState(t, fx)
			if result.Published || !bytes.Equal(before, after) {
				t.Fatal("refused MCP batch changed preimage")
			}
			breakFixtureLease(t, m, lease)
			result, err = ProfileMCP(req)
			if err != nil || !result.Published {
				t.Fatalf("unleased native MCP control: %+v %v", result, err)
			}
			for _, id := range req.Environments {
				if len(nativeRead(t, nativeWriterPath(fx, id))) == 0 || len(nativeRead(t, nativeMCPLedger(fx.home, fx.profile, id))) == 0 {
					t.Fatal("control omitted a batch target")
				}
			}
		})
	}
}
