package main

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/scopes"
)

// TestGCRetainsLiveProcessBuild exercises run("gc") -> collectUnderLock ->
// scopes.Collect -> Store.Sweep with a real process executing an old cache
// image. No marker, journal reference or publication grace can protect it.
func TestGCRetainsLiveProcessBuild(t *testing.T) {
	requireNativeControlInventoryPlatform(t)
	project, home := compiledProject(t)
	skill := filepath.Join(filepath.Dir(home), "skills", "build-skill")
	writeFile(t, filepath.Join(skill, "assets", "build-tool", "cmd", "tool", "main.go"),
		"package main\nimport (\"fmt\"; \"io\"; \"os\")\nfunc main() { fmt.Println(\"ready\"); _, _ = io.Copy(io.Discard, os.Stdin) }\n")
	runGit(t, skill, "add", ".")
	runGit(t, skill, "commit", "-qm", "waiting executable fixture")
	runGit(t, skill, "tag", "-f", "v1")
	if code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), "install", "app"); code != exitOK {
		t.Fatalf("install = %d\n%s\n%s", code, stdout, stderr)
	}
	entries := cacheEntries(t, home)
	if len(entries) != 1 {
		t.Fatalf("cache entries = %v", entries)
	}
	bin := filepath.Join(entries[0], "bin")
	files, err := os.ReadDir(bin)
	if err != nil || len(files) != 1 {
		t.Fatalf("bin files = %v, %v", files, err)
	}
	child := exec.Command(filepath.Join(bin, files[0].Name()))
	stdin, err := child.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := child.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = stdin.Close()
		if err := child.Wait(); err != nil {
			t.Errorf("cache executable exit: %v", err)
		}
	})
	ready := make(chan bool, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		ready <- scanner.Scan() && scanner.Text() == "ready"
	}()
	select {
	case ok := <-ready:
		if !ok {
			t.Fatal("cache executable did not signal readiness")
		}
	case <-time.After(30 * time.Second):
		t.Fatal("cache executable readiness timed out")
	}
	if err := os.Remove(filepath.Join(project, ".agents", "skills", "build-skill", marker.Name)); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-30 * 24 * time.Hour)
	if err := os.Chtimes(entries[0], old, old); err != nil {
		t.Fatal(err)
	}
	code, output, stderr := capture(t, filepath.Join(home, "config.json"), "gc")
	if code != exitOK || !strings.Contains(output, "0 build entries removed") {
		t.Fatalf("gc = %d\n%s\n%s", code, output, stderr)
	}
	if _, err := os.Stat(entries[0]); err != nil {
		t.Fatalf("gc removed a live process executable: %v", err)
	}
	if strings.Contains(stderr, "live process executables could not be enumerated") {
		t.Logf("native enumeration could not inspect every process; exercised fail-safe retention: %s", stderr)
	}
}

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
	return home
}

func mustLoadConsumers(t *testing.T, home string) []string {
	t.Helper()
	consumers, err := scopes.LoadConsumers(home)
	if err != nil {
		t.Fatal(err)
	}
	return consumers
}

// TestGCPrunesDeadConsumersUnderTheHomeLock proves the standalone command does
// the same serialized maintenance an install does.
func TestGCPrunesDeadConsumersUnderTheHomeLock(t *testing.T) {
	t.Parallel()
	home := gcHome(t)
	dead := filepath.Join(t.TempDir(), "gone")
	if err := scopes.RecordConsumer(home, dead); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, "runtime", "skill-x", strings.Repeat("3", 40)), 0o755); err != nil {
		t.Fatal(err)
	}

	if code := run([]string{"gc"}, fileConfigSource(filepath.Join(home, "config.json")), io.Discard, io.Discard); code != exitOK {
		t.Fatalf("gc = %d", code)
	}
	if consumers := mustLoadConsumers(t, home); len(consumers) != 0 {
		t.Fatalf("dead consumer survived gc: %v", consumers)
	}
	if _, err := os.Stat(filepath.Join(home, "runtime", "skill-x")); err == nil {
		t.Fatal("unreferenced runtime entry survived gc")
	}
}

func TestGCFailsClosedForUntrustedCurrentPathSource(t *testing.T) {
	source, home := profileHome(t)
	packageRoot := t.TempDir()
	writeContextPackage(t, packageRoot, "gc-path-source", "1.0.0", "context\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", packageRoot); code != exitOK {
		t.Fatalf("path profile install = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "profile", "use", "gc-path-source"); code != exitOK {
		t.Fatalf("use path profile = %d\nstderr:\n%s", code, stderr)
	}
	if err := os.RemoveAll(packageRoot); err != nil {
		t.Fatal(err)
	}
	dead := filepath.Join(t.TempDir(), "gone")
	if err := scopes.RecordConsumer(home, dead); err != nil {
		t.Fatal(err)
	}

	code, _, stderr := runProfile(t, source, "gc")
	if code != exitFail || !strings.Contains(stderr, "environment_store_untrusted") || !strings.Contains(stderr, "regular_types") {
		t.Fatalf("gc = %d\nstderr:\n%s\nwanted environment_store_untrusted naming regular_types", code, stderr)
	}
	if consumers := mustLoadConsumers(t, home); len(consumers) != 1 || consumers[0] != dead {
		t.Fatalf("gc pruned consumers before refusing the untrusted path source: %v", consumers)
	}
}

// TestGCWaitsForTheHomeLock proves maintenance serializes on the same lock
// install, rollback, and recovery take, instead of running beside them.
func TestGCWaitsForTheHomeLock(t *testing.T) {
	t.Parallel()
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
	go func() {
		finished <- run([]string{"gc"}, fileConfigSource(filepath.Join(home, "config.json")), io.Discard, io.Discard)
	}()

	select {
	case code := <-finished:
		t.Fatalf("gc ran while the home lock was held (exit %d)", code)
	case <-time.After(250 * time.Millisecond):
	}
	if consumers := mustLoadConsumers(t, home); len(consumers) != 1 {
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
	if consumers := mustLoadConsumers(t, home); len(consumers) != 0 {
		t.Fatalf("gc did not prune after acquiring the lock: %v", consumers)
	}
}

// TestGCRunsSerializedAcrossConcurrentInvocations proves concurrent maintenance
// passes over one home cannot lose a consumer update: each pass observes a
// complete registry, so the live consumer always survives.
func TestGCRunsSerializedAcrossConcurrentInvocations(t *testing.T) {
	t.Parallel()
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
	consumers := mustLoadConsumers(t, home)
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
	_, collectErr := collectUnderLock(home, lock, envprofile.Policy{})
	closeErr := lock.Close()
	if collectErr != nil || closeErr != nil {
		t.Errorf("collect = %v, close = %v", collectErr, closeErr)
		return exitFail
	}
	return exitOK
}
