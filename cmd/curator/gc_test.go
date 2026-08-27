package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/scopes"
)

// installTestMarker writes one valid install marker so a project counts as a
// live consumer.
func installTestMarker(t *testing.T, skillsDir, name string) {
	t.Helper()
	dir := filepath.Join(skillsDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	hash, _ := hashing.ContentSHA256(dir, nil)
	if err := marker.Write(dir, &marker.Marker{
		Name: name, Source: name, RefKind: "tag", Ref: "v1",
		Commit: strings.Repeat("1", 40), ContentSHA256: hash,
		Agents: []string{}, Commands: []string{}, Dependencies: []string{},
		InstalledAt: "2026-07-20T00:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
}

// gcHome bootstraps a manager home the gc command can run against.
func gcHome(t *testing.T) string {
	t.Helper()
	home := filepath.Join(t.TempDir(), "manager home")
	if err := os.MkdirAll(home, 0o700); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(home, "config.json")
	payload, err := json.Marshal(map[string]any{
		"schema_version": 1,
		"skills_root":    filepath.Join(home, "skills"),
		"projects":       map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_CONFIG", configPath)
	return home
}

// TestGCPrunesDeadConsumersUnderTheHomeLock proves the standalone command does
// the same serialized maintenance an install does.
func TestGCPrunesDeadConsumersUnderTheHomeLock(t *testing.T) {
	home := gcHome(t)
	dead := filepath.Join(t.TempDir(), "gone")
	if err := scopes.RecordConsumer(home, dead); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, "runtime", "skill-x", strings.Repeat("3", 40)), 0o755); err != nil {
		t.Fatal(err)
	}

	if code := run([]string{"gc"}); code != exitOK {
		t.Fatalf("gc = %d", code)
	}
	if consumers := scopes.LoadConsumers(home); len(consumers) != 0 {
		t.Fatalf("dead consumer survived gc: %v", consumers)
	}
	if _, err := os.Stat(filepath.Join(home, "runtime", "skill-x")); err == nil {
		t.Fatal("unreferenced runtime entry survived gc")
	}
}

// TestGCWaitsForTheHomeLock proves maintenance serializes on the same lock
// install, rollback, and recovery take, instead of running beside them.
func TestGCWaitsForTheHomeLock(t *testing.T) {
	home := gcHome(t)
	dead := filepath.Join(t.TempDir(), "gone")
	if err := scopes.RecordConsumer(home, dead); err != nil {
		t.Fatal(err)
	}
	manager, err := managerlock.New(home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := manager.AcquireHomeOnly(t.Context(), false)
	if err != nil {
		t.Fatal(err)
	}
	released := false
	defer func() {
		if !released {
			_ = lock.Close()
		}
	}()

	finished := make(chan int, 1)
	go func() { finished <- run([]string{"gc"}) }()

	select {
	case code := <-finished:
		t.Fatalf("gc ran while the home lock was held (exit %d)", code)
	case <-time.After(250 * time.Millisecond):
	}
	if consumers := scopes.LoadConsumers(home); len(consumers) != 1 {
		t.Fatalf("a blocked gc still pruned consumers: %v", consumers)
	}

	released = true
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case code := <-finished:
		if code != exitOK {
			t.Fatalf("gc after the lock was released = %d", code)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("gc did not resume after the home lock was released")
	}
	if consumers := scopes.LoadConsumers(home); len(consumers) != 0 {
		t.Fatalf("gc did not prune after acquiring the lock: %v", consumers)
	}
}

// TestGCRunsSerializedAcrossConcurrentInvocations proves concurrent maintenance
// passes over one home cannot lose a consumer update: each pass observes a
// complete registry, so the live consumer always survives.
func TestGCRunsSerializedAcrossConcurrentInvocations(t *testing.T) {
	home := gcHome(t)
	live := t.TempDir()
	installTestMarker(t, filepath.Join(live, ".agents", "skills"), "skill-a")
	if err := scopes.RecordConsumer(home, live); err != nil {
		t.Fatal(err)
	}
	dead := filepath.Join(t.TempDir(), "gone")
	if err := scopes.RecordConsumer(home, dead); err != nil {
		t.Fatal(err)
	}

	var wait sync.WaitGroup
	codes := make([]int, 4)
	for index := range codes {
		wait.Add(1)
		go func() {
			defer wait.Done()
			codes[index] = runGC(t, home)
		}()
	}
	wait.Wait()
	for index, code := range codes {
		if code != exitOK {
			t.Fatalf("concurrent gc %d = %d", index, code)
		}
	}
	consumers := scopes.LoadConsumers(home)
	if len(consumers) != 1 || consumers[0] != live {
		t.Fatalf("concurrent maintenance lost a consumer update: %v", consumers)
	}
}

// runGC exercises the locked maintenance path directly, so it is safe to call
// from several goroutines at once (the CLI entry point mutates process state).
func runGC(t *testing.T, home string) int {
	t.Helper()
	manager, err := managerlock.New(home)
	if err != nil {
		t.Error(err)
		return exitFail
	}
	lock, err := manager.AcquireHomeOnly(t.Context(), false)
	if err != nil {
		t.Error(err)
		return exitFail
	}
	_, collectErr := collectUnderLock(home, lock)
	closeErr := lock.Close()
	if collectErr != nil || closeErr != nil {
		t.Errorf("collect = %v, close = %v", collectErr, closeErr)
		return exitFail
	}
	return exitOK
}
