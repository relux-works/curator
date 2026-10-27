//go:build unix

package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/manifest"
)

// This fixture-only shell barrier forwards real Git's arguments, streams and
// status without reexecuting a race-instrumented Go binary for each Git call.
// Noclobber claims the ready file once; only that invocation waits for release.
func installNativeGitBarrier(t *testing.T, published, ready, release string) {
	t.Helper()
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
	script := "#!/bin/sh\npublished=" + quote(published) + "\nready=" + quote(ready) + "\nrelease=" + quote(release) + "\n"
	script += "if [ -e \"$published\" ] && (set -C; : > \"$ready\") 2>/dev/null; then\n  while [ ! -e \"$release\" ]; do sleep 0.01; done\nfi\n"
	script += "exec " + quote(git) + " \"$@\"\n"
	path := filepath.Join(dir, "git")
	writeFile(t, path, script)
	if err := os.Chmod(path, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// Child entry restores only the parent's fixture Git mapping after TestMain's
// ambient-config isolation, then invokes the production CLI without fake I/O.
func TestNativeCommandProcess(t *testing.T) {
	if os.Getenv("CURATOR_NATIVE_COMMAND_HELPER") != "1" {
		return
	}
	restoreCLIUnitHomes(t)
	_ = os.Setenv("GIT_CONFIG_GLOBAL", os.Getenv("CURATOR_NATIVE_GIT_CONFIG"))
	for i, arg := range os.Args {
		if arg == "--" {
			os.Exit(run(os.Args[i+1:], fileConfigSource(os.Getenv(cliHelperConfig)), os.Stdout, os.Stderr))
		}
	}
	os.Exit(94)
}

// A CLI can run the production two-minute compiler before verification and
// publication. This fixture watchdog is not a CLI latency requirement.
const nativeCommandWatchdog = 3 * time.Minute

type nativeCommand struct {
	cmd      *exec.Cmd
	done     chan error
	output   bytes.Buffer
	finished bool
}

func startNativeCommand(t *testing.T, fixture nativeGlobalFixture, args ...string) *nativeCommand {
	t.Helper()
	child := &nativeCommand{done: make(chan error, 1)}
	child.cmd = exec.Command(os.Args[0], append([]string{"-test.run=^TestNativeCommandProcess$", "--"}, args...)...)
	child.cmd.Env = append(os.Environ(), cliHelperProcess+"=1", cliHelperConfig+"="+fixture.config,
		cliHelperHomes+"="+forwardCLIUnitHomes(t, nil),
		"CURATOR_NATIVE_COMMAND_HELPER=1", "CURATOR_NATIVE_GIT_CONFIG="+os.Getenv("GIT_CONFIG_GLOBAL"))
	child.cmd.Stdout = &child.output
	child.cmd.Stderr = &child.output
	child.cmd.WaitDelay = 5 * time.Second
	if err := child.cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Logf("started production CLI PID %d: %v", child.cmd.Process.Pid, args)
	go func() { child.done <- child.cmd.Wait() }()
	t.Cleanup(func() {
		if !child.finished {
			_ = child.cmd.Process.Kill()
			<-child.done
			child.finished = true
		}
	})
	return child
}
func (child *nativeCommand) wait(t *testing.T) {
	t.Helper()
	child.waitWithin(t, nativeCommandWatchdog)
}

func (child *nativeCommand) waitWithin(t *testing.T, timeout time.Duration) {
	t.Helper()
	select {
	case err := <-child.done:
		child.finished = true
		if err != nil {
			t.Fatalf("production CLI failed: %v\n%s", err, child.output.String())
		}
	case <-time.After(timeout):
		t.Fatal(child.timeoutDiagnostic())
	}
}
func (child *nativeCommand) pending(t *testing.T) {
	t.Helper()
	select {
	case err := <-child.done:
		child.finished = true
		t.Fatalf("remove completed during paused install (or install during paused remove): %v\n%s", err, child.output.String())
	case <-time.After(200 * time.Millisecond):
	}
}
func waitNativePath(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(nativeCommandWatchdog)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("production Git pause did not reach native publication")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
func setupNativeConcurrent(t *testing.T, required bool) nativeGlobalFixture {
	t.Helper()
	fixture := newNativeGlobalFixture(t)
	profile := t.TempDir()
	writeContextPackage(t, profile, "selected-profile", "1.0.0", "profile context\n")
	if required {
		writeFile(t, filepath.Join(profile, "agent-context.json"), `{"schema_version":1,"name":"selected-profile","version":"1.0.0","context":{"modules":[{"path":"a.md"}]},"requires":{"skills":{"shared-skill":{"git":"https://example.com/shared-skill","tag":"v1.0.0"}}}}`)
	}
	fixture.command(t, exitOK, "profile", "install", profile, "--use")
	fixture.declare(t, "v1.0.0")
	return fixture
}
func requireNativeDeclaration(t *testing.T, fixture nativeGlobalFixture, want bool) {
	t.Helper()
	m, err := manifest.Load(filepath.Join(fixture.home, ".curator", "global"))
	if err != nil || m == nil {
		t.Fatalf("Skillfile unavailable: %v", err)
	}
	if (len(m.Skills) == 1) != want {
		t.Fatalf("declaration state: %+v, want present=%v", m.Skills, want)
	}
}
func requireNativeClosure(t *testing.T, fixture nativeGlobalFixture, want bool) {
	t.Helper()
	lock, _, err := contextlock.Read(filepath.Join(fixture.home, ".curator", "profiles", "selected-profile", "lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, present := lock.Find(contextlock.KindSkill, "shared-skill")
	if present != want {
		t.Fatalf("profile used another declaration generation: skill present=%v, want %v", present, want)
	}
	if want {
		fixture.requireProfileOwner(t)
	} else {
		fixture.requireAbsent(t)
	}
}
func requireNativeCanonical(t *testing.T, fixture nativeGlobalFixture, want bool) {
	t.Helper()
	_, err := os.Lstat(filepath.Join(fixture.home, ".curator", "global", "skills", "shared-skill"))
	if (err == nil) != want || err != nil && !os.IsNotExist(err) {
		t.Fatalf("canonical state: %v, want present=%v", err, want)
	}
}

// A Git pause after native publication cannot let global remove complete inside
// the paired install. Install-before-remove leaves both publishers installed;
// the next install retires a direct-only skill, but preserves a live requirement.
func TestGlobalNativeOwnerConcurrentRemove(t *testing.T) {
	for _, required := range []bool{false, true} {
		t.Run(map[bool]string{false: "direct-only", true: "profile-required"}[required], func(t *testing.T) {
			fixture := setupNativeConcurrent(t, required)
			// Managed publication follows native publication and is absent before this
			// install, even when the profile already requires the skill.
			published := filepath.Join(envprofile.ManagedHomeDir(filepath.Join(fixture.home, ".curator"), "selected-profile", envregistry.CodexCLI), envmarker.Name)
			if _, err := os.Lstat(published); !os.IsNotExist(err) {
				t.Fatalf("barrier needs a fresh managed publication: %v", err)
			}
			dir := t.TempDir()
			ready, release := filepath.Join(dir, "ready"), filepath.Join(dir, "release")
			installNativeGitBarrier(t, published, ready, release)
			installing := startNativeCommand(t, fixture, "global", "install")
			// Release before child cleanup even on assertion failure.
			defer func() { _ = os.WriteFile(release, []byte("release"), 0600) }()
			waitNativePath(t, ready)
			fixture.requireProfileOwner(t)
			removing := startNativeCommand(t, fixture, "global", "remove", "shared-skill")
			removing.pending(t)
			requireNativeDeclaration(t, fixture, true)
			writeFile(t, release, "release")
			installing.wait(t)
			removing.wait(t)
			requireNativeDeclaration(t, fixture, false)
			requireNativeClosure(t, fixture, true)
			requireNativeCanonical(t, fixture, true)
			fixture.command(t, exitOK, "global", "install")
			requireNativeClosure(t, fixture, required)
			requireNativeCanonical(t, fixture, false)
		})
	}
}

// managerlock selects on Done only after an OS-lock attempt reports contention.
// Cancel at that event so the probe tests lock ownership without a clock budget.
type nativeLockContentionContext struct {
	context.Context
	cancel context.CancelFunc
}

func (ctx nativeLockContentionContext) Done() <-chan struct{} {
	ctx.cancel()
	return ctx.Context.Done()
}

// A FIFO pauses the real removal read after it holds the operation lock. The
// path is then restored to an ordinary file, so the competing install can read
// normally and must wait on that lock. Remove-before-install leaves both native
// and canonical entries absent. No product pause hook or fake Git result is used.
func TestGlobalNativeOwnerRemoveBeforeInstall(t *testing.T) {
	fixture := setupNativeConcurrent(t, false)
	path := manifest.PathIn(filepath.Join(fixture.home, ".curator", "global"))
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	removing := startNativeCommand(t, fixture, "global", "remove", "shared-skill")
	connect := func() *os.File {
		deadline := time.Now().Add(nativeCommandWatchdog)
		for {
			fd, err := syscall.Open(path, syscall.O_WRONLY|syscall.O_NONBLOCK, 0600)
			if err == nil {
				syscall.CloseOnExec(fd)
				return os.NewFile(uintptr(fd), path)
			}
			if time.Now().After(deadline) {
				t.Fatalf("removal did not enter its manifest read: %v", err)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	// The first reader is cmdGlobal's read-only schema preflight.
	preflight := connect()
	// Give the second read a distinct pipe inode; it cannot accidentally
	// connect to the schema reader before that reader closes its descriptor.
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := preflight.Write(payload); err != nil {
		t.Fatal(err)
	}
	_ = preflight.Close()
	// The second read is RemoveDecl, after acquisition of the operation lock.
	writer := connect()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	writeFile(t, path, string(payload))
	released := false
	release := func() {
		if !released {
			_, _ = writer.Write(payload)
			_ = writer.Close()
			released = true
		}
	}
	defer release()
	// Probe before starting the competing install: only the paused removal
	// may own this lock. Holding the home lock alone does not prove this
	// ownership; an installer could mask a missing or prematurely released lock.
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	manager, err := managerlock.New(filepath.Join(fixture.home, ".curator"))
	if err != nil {
		t.Fatal(err)
	}
	operation := manager.NewOperation(false)
	defer func() { _ = operation.Close() }()
	_, err = operation.AcquireProjects(nativeLockContentionContext{ctx, cancel}, filepath.Join(fixture.home, ".curator", "global"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("paused remove did not hold the global operation lock: %v", err)
	}
	installing := startNativeCommand(t, fixture, "global", "install")
	installing.pending(t)
	requireNativeClosure(t, fixture, false)
	release()
	removing.wait(t)
	installing.wait(t)
	requireNativeDeclaration(t, fixture, false)
	requireNativeClosure(t, fixture, false)
	requireNativeCanonical(t, fixture, false)
}

// With no concurrent writer the declaration is installed by both publishers;
// when removal finishes before install starts, neither publisher installs it.
func TestGlobalNativeOwnerRemovalOrderControls(t *testing.T) {
	for _, early := range []bool{false, true} {
		t.Run(map[bool]string{false: "unchanged", true: "early-removal"}[early], func(t *testing.T) {
			fixture := setupNativeConcurrent(t, false)
			if early {
				fixture.command(t, exitOK, "global", "remove", "shared-skill")
			}
			fixture.command(t, exitOK, "global", "install")
			requireNativeDeclaration(t, fixture, !early)
			requireNativeClosure(t, fixture, !early)
			requireNativeCanonical(t, fixture, !early)
		})
	}
}
