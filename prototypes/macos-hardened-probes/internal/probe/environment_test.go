package probe

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/inside"
)

// newTestEnvironment builds a probe environment under a fresh directory, using
// a harmless stand-in for the agent binary where the test does not need to run
// anything.
func newTestEnvironment(t *testing.T) *Environment {
	t.Helper()
	env, err := NewEnvironment(t.TempDir(), "/bin/echo")
	if err != nil {
		t.Fatalf("NewEnvironment: %v", err)
	}
	t.Cleanup(env.Close)
	return env
}

func TestNewEnvironmentBuildsTheScaffolding(t *testing.T) {
	env := newTestEnvironment(t)

	for name, dir := range map[string]string{
		"source": env.SourceDir, "goroot": env.GorootDir, "build root": env.BuildRoot,
		"outside": env.OutsideDir, "profiles": env.ProfileDir, "markers": env.MarkerDir,
	} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("%s directory missing: %v", name, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", name)
		}
	}
	// A mutation attempt needs something real to aim at, or every denial would
	// be an ENOENT that proves nothing.
	for _, path := range []string{
		filepath.Join(env.SourceDir, "main.go"),
		filepath.Join(env.SourceDir, "go.mod"),
		filepath.Join(env.GorootDir, "bin", "go"),
		filepath.Join(env.GorootDir, "VERSION"),
		filepath.Join(env.OutsideDir, "outside-target"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("stand-in file missing: %v", err)
		}
	}
}

// A seatbelt filter matches the path the kernel resolved. A root reached
// through a symlink — which is what /tmp and /var are on macOS — would match
// nothing, and every probe would report a denial that came from the mismatch
// rather than from the profile.
func TestNewEnvironmentResolvesSymlinks(t *testing.T) {
	resolved := t.TempDir()
	link := filepath.Join(t.TempDir(), "via-a-symlink")
	if err := os.Symlink(resolved, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	env, err := NewEnvironment(link, "/bin/echo")
	if err != nil {
		t.Fatalf("NewEnvironment: %v", err)
	}
	defer env.Close()

	resolvedReal, err := filepath.EvalSymlinks(resolved)
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}
	if env.Root != resolvedReal {
		t.Errorf("root %q, want the resolved %q", env.Root, resolvedReal)
	}
	// The symlink's own name must be gone from the resolved root. It is checked
	// against the final component rather than the whole path, because t.TempDir
	// embeds the test's name and would otherwise match its own description.
	if filepath.Base(env.Root) == filepath.Base(link) {
		t.Errorf("root %q still ends in the symlink component", env.Root)
	}
	// Every derived path has to sit under the resolved root too; a single
	// unresolved child would be a filter that silently matches nothing.
	for _, dir := range []string{env.SourceDir, env.GorootDir, env.BuildRoot, env.OutsideDir} {
		if !strings.HasPrefix(dir, resolvedReal+string(filepath.Separator)) {
			t.Errorf("derived path %q is not under the resolved root %q", dir, resolvedReal)
		}
	}
}

func TestNewEnvironmentCreatesAMissingRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "not", "yet", "there")
	env, err := NewEnvironment(root, "/bin/echo")
	if err != nil {
		t.Fatalf("NewEnvironment: %v", err)
	}
	defer env.Close()

	if _, err := os.Stat(env.BuildRoot); err != nil {
		t.Errorf("build root not created: %v", err)
	}
}

// An unusable agent path is a harness fault, not an observation about the host,
// so it has to be refused up front. The dangerous case is the silent one: if a
// missing agent were created on demand, every later class probe would fail to
// exec and the run would report the host as incapable.
func TestNewEnvironmentRejectsUnusablePaths(t *testing.T) {
	notExecutable := filepath.Join(t.TempDir(), "not-executable")
	if err := os.WriteFile(notExecutable, []byte("#!/bin/sh\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	cases := map[string]struct {
		emptyRoot bool
		self      string
	}{
		"empty root":              {emptyRoot: true, self: "/bin/echo"},
		"empty agent":             {emptyRoot: true, self: ""},
		"missing agent":           {self: filepath.Join(t.TempDir(), "not-a-program")},
		"agent is nowhere":        {self: "/definitely/not/here"},
		"agent is a directory":    {self: t.TempDir()},
		"agent is not executable": {self: notExecutable},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			root := ""
			if !tc.emptyRoot {
				root = t.TempDir()
			}
			env, err := NewEnvironment(root, tc.self)
			if err == nil {
				env.Close()
				t.Fatalf("NewEnvironment(%q, %q) returned no error", root, tc.self)
			}
			if !strings.Contains(err.Error(), "probe environment") {
				t.Errorf("error %q is not attributed to the probe environment", err)
			}
		})
	}

	// The missing agent must also be left missing. Creating it would be the
	// failure this test exists to prevent.
	missing := filepath.Join(t.TempDir(), "not-a-program")
	if _, err := NewEnvironment(t.TempDir(), missing); err == nil {
		t.Fatal("NewEnvironment accepted a missing agent")
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Errorf("the missing agent path was created: stat err = %v", err)
	}
}

// macOS caps sockaddr_un.sun_path at 104 bytes. A work directory deep enough to
// push the endpoint past it must relocate the socket rather than fail the run
// with an EINVAL that reads like an ordinary bind error.
func TestNewEnvironmentRelocatesAnOverlongSocketPath(t *testing.T) {
	root := t.TempDir()
	for len(filepath.Join(root, "outside", "probe.sock")) <= maxUnixSocketPath {
		root = filepath.Join(root, strings.Repeat("d", 16))
	}

	env, err := NewEnvironment(root, "/bin/echo")
	if err != nil {
		t.Fatalf("NewEnvironment: %v", err)
	}
	defer env.Close()

	if !env.UnixSocketRelocated {
		t.Error("an overlong socket path was not reported as relocated")
	}
	if len(env.UnixSocket) > maxUnixSocketPath {
		t.Errorf("socket path %q is %d bytes, over the %d limit",
			env.UnixSocket, len(env.UnixSocket), maxUnixSocketPath)
	}
	// Relocated or not, the endpoint has to be a real listening socket, or the
	// network probe's connect-unix denial would prove nothing.
	conn, err := net.Dial("unix", env.UnixSocket)
	if err != nil {
		t.Fatalf("dial relocated socket: %v", err)
	}
	_ = conn.Close()

	// It must also still be outside the build root, which is the only property
	// the probe depends on.
	if strings.HasPrefix(env.UnixSocket, env.BuildRoot+string(filepath.Separator)) {
		t.Errorf("relocated socket %q landed inside the build root", env.UnixSocket)
	}

	dir := env.unixSocketDir
	env.Close()
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("relocated socket directory survived Close: stat err = %v", err)
	}
}

// A socket that fits stays in the scaffolding, so the common case keeps the
// whole environment under one root.
//
// The root is built with MkdirTemp rather than t.TempDir because t.TempDir
// embeds the test's own name, which on macOS is already long enough to force
// the relocation this test exists to rule out.
func TestNewEnvironmentKeepsAShortSocketPathInPlace(t *testing.T) {
	root, err := os.MkdirTemp("", "hp")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(root) })

	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}
	if got := len(filepath.Join(resolved, "outside", "probe.sock")); got > maxUnixSocketPath {
		t.Skipf("this host's temporary directory is too deep for a short-path case (%d bytes)", got)
	}

	env, err := NewEnvironment(root, "/bin/echo")
	if err != nil {
		t.Fatalf("NewEnvironment: %v", err)
	}
	t.Cleanup(env.Close)

	if env.UnixSocketRelocated {
		t.Fatalf("short path %q was relocated", env.UnixSocket)
	}
	if want := filepath.Join(env.OutsideDir, "probe.sock"); env.UnixSocket != want {
		t.Errorf("socket = %q, want %q", env.UnixSocket, want)
	}
}

func TestResolvePath(t *testing.T) {
	if _, err := resolvePath(""); err == nil {
		t.Error("resolvePath(\"\") returned no error")
	}
	dir := t.TempDir()
	got, err := resolvePath(dir)
	if err != nil {
		t.Fatalf("resolvePath: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("resolvePath = %q, want an absolute path", got)
	}
}

// The agent inherits exactly this list and nothing else, so no ambient value
// from the operator's shell can change what it attempts.
func TestAgentEnvIsAClosedList(t *testing.T) {
	env := newTestEnvironment(t)
	got := env.AgentEnv()

	values := map[string]string{}
	for _, entry := range got {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			t.Errorf("environment entry %q is not a key=value pair", entry)
			continue
		}
		if _, duplicate := values[key]; duplicate {
			t.Errorf("environment sets %q twice", key)
		}
		values[key] = value
	}

	for key, want := range map[string]string{
		inside.EnvSource:     env.SourceDir,
		inside.EnvGoroot:     env.GorootDir,
		inside.EnvBuildRoot:  env.BuildRoot,
		inside.EnvOutside:    env.OutsideDir,
		inside.EnvSelf:       env.SelfPath,
		inside.EnvLoopback:   env.LoopbackAddr,
		inside.EnvUnixSocket: env.UnixSocket,
	} {
		if values[key] != want {
			t.Errorf("%s = %q, want %q", key, values[key], want)
		}
	}
	// An empty PATH is deliberate: the agent must reach programs by absolute
	// path only, so a PATH lookup can never pick a different binary.
	if path, ok := values["PATH"]; !ok || path != "" {
		t.Errorf("PATH = %q (present=%v), want an empty PATH", path, ok)
	}
	// Anything the harness did not choose must not be there.
	allowed := map[string]bool{
		inside.EnvSource: true, inside.EnvGoroot: true, inside.EnvBuildRoot: true,
		inside.EnvOutside: true, inside.EnvSelf: true, inside.EnvLoopback: true,
		inside.EnvUnixSocket: true, "HOME": true, "TMPDIR": true, "PATH": true,
	}
	for key := range values {
		if !allowed[key] {
			t.Errorf("the agent environment carries an unexpected variable %q", key)
		}
	}
}

func TestAgentEnvAppendsExtras(t *testing.T) {
	env := newTestEnvironment(t)
	got := env.AgentEnv(inside.EnvHold+"=3", inside.EnvMarker+"=/tmp/marker")

	var sawHold, sawMarker bool
	for _, entry := range got {
		switch entry {
		case inside.EnvHold + "=3":
			sawHold = true
		case inside.EnvMarker + "=/tmp/marker":
			sawMarker = true
		}
	}
	if !sawHold || !sawMarker {
		t.Errorf("extras were not appended: %v", got)
	}
	// The extras must not displace the closed base list.
	if len(got) != len(env.AgentEnv())+2 {
		t.Errorf("environment has %d entries, want the base plus two", len(got))
	}
}

// The loopback and unix endpoints are what the network probe dials. If they
// were not actually listening, a denial in-domain would be indistinguishable
// from nothing being there.
func TestEnvironmentEndpointsAreListening(t *testing.T) {
	env := newTestEnvironment(t)

	conn, err := net.Dial("tcp", env.LoopbackAddr)
	if err != nil {
		t.Errorf("the loopback endpoint is not listening: %v", err)
	} else {
		_ = conn.Close()
	}
	unixConn, err := net.Dial("unix", env.UnixSocket)
	if err != nil {
		t.Errorf("the unix endpoint is not listening: %v", err)
	} else {
		_ = unixConn.Close()
	}
}

func TestEnvironmentCloseIsIdempotent(t *testing.T) {
	env, err := NewEnvironment(t.TempDir(), "/bin/echo")
	if err != nil {
		t.Fatalf("NewEnvironment: %v", err)
	}
	env.Close()
	env.Close()
	env.Close()

	if conn, err := net.Dial("tcp", env.LoopbackAddr); err == nil {
		_ = conn.Close()
		t.Error("the loopback endpoint is still listening after Close")
	}
	// Close keeps the files: they are the evidence the caller came for.
	if _, err := os.Stat(env.SourceDir); err != nil {
		t.Errorf("Close removed the probe environment: %v", err)
	}
}

// A leftover file from one probe must not be able to make the next one pass.
func TestResetBuildRootEmptiesIt(t *testing.T) {
	env := newTestEnvironment(t)
	stale := filepath.Join(env.BuildRoot, "stale")
	if err := os.WriteFile(stale, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := env.ResetBuildRoot(); err != nil {
		t.Fatalf("ResetBuildRoot: %v", err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("the stale file survived the reset: %v", err)
	}
	entries, err := os.ReadDir(env.BuildRoot)
	if err != nil {
		t.Fatalf("read build root: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("the build root still holds %d entries after a reset", len(entries))
	}
}

func TestCleanEscapesRemovesTheTargets(t *testing.T) {
	env := newTestEnvironment(t)
	home := t.TempDir()
	tmpdir := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("TMPDIR", tmpdir)

	created := []string{
		filepath.Join(env.OutsideDir, "escape-absolute"),
		filepath.Join(env.OutsideDir, "escape-through-symlink"),
		filepath.Join(home, ".probe-escape"),
		filepath.Join(env.Root, "escape-parent"),
		filepath.Join(env.Root, "escape-traversal"),
		filepath.Join(tmpdir, "probe-escape"),
	}
	for _, path := range created {
		if err := os.WriteFile(path, []byte("escape"), 0o600); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	env.CleanEscapes()

	for _, path := range created {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%s survived CleanEscapes: %v", path, err)
		}
	}
}

// The control run makes the view writable and the mutations then really happen.
// Without a restore, the next probe would find a view that no longer matches
// what the profile describes.
func TestRestoreViewRebuildsTheStandInViews(t *testing.T) {
	env := newTestEnvironment(t)

	for _, view := range []string{env.SourceDir, env.GorootDir} {
		// Simulate what a successful control run leaves behind.
		if err := os.WriteFile(filepath.Join(view, "probe-created"), []byte("x"), 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(view, "probe-created-dir"), 0o700); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.Symlink("/etc/hosts", filepath.Join(view, "probe-symlink")); err != nil {
			t.Fatalf("symlink: %v", err)
		}
	}
	// The unlink attempt really removes the sample file.
	if err := os.Remove(filepath.Join(env.SourceDir, "main.go")); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if err := os.Remove(filepath.Join(env.GorootDir, "VERSION")); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if err := os.WriteFile(filepath.Join(env.SourceDir, "go.mod.renamed"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := env.restoreView(env.SourceDir); err != nil {
		t.Fatalf("restoreView(source): %v", err)
	}
	if err := env.restoreView(env.GorootDir); err != nil {
		t.Fatalf("restoreView(toolchain): %v", err)
	}

	for _, view := range []string{env.SourceDir, env.GorootDir} {
		for _, leftover := range []string{"probe-created", "probe-created-dir", "probe-symlink"} {
			if _, err := os.Lstat(filepath.Join(view, leftover)); !os.IsNotExist(err) {
				t.Errorf("%s survived the restore in %s: %v", leftover, filepath.Base(view), err)
			}
		}
	}
	for _, path := range []string{
		filepath.Join(env.SourceDir, "main.go"),
		filepath.Join(env.SourceDir, "go.mod"),
		filepath.Join(env.GorootDir, "VERSION"),
		filepath.Join(env.GorootDir, "bin", "go"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("the restore did not rebuild %s: %v", path, err)
		}
	}
	if _, err := os.Stat(filepath.Join(env.SourceDir, "go.mod.renamed")); !os.IsNotExist(err) {
		t.Errorf("the restore left a renamed leftover behind: %v", err)
	}
}

// A view the environment does not own must not be rewritten by a restore.
func TestRestoreViewIgnoresUnknownDirectories(t *testing.T) {
	env := newTestEnvironment(t)
	stranger := t.TempDir()
	if err := os.WriteFile(filepath.Join(stranger, "keep"), []byte("keep"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := env.restoreView(stranger); err != nil {
		t.Fatalf("restoreView(stranger): %v", err)
	}

	if _, err := os.Stat(filepath.Join(stranger, "keep")); err != nil {
		t.Errorf("restoreView disturbed a directory it does not own: %v", err)
	}
	if _, err := os.Stat(filepath.Join(stranger, "main.go")); err == nil {
		t.Error("restoreView wrote stand-in files into a directory it does not own")
	}
}

func TestEnvironmentRunnerWritesIntoTheProfileDirectory(t *testing.T) {
	env := newTestEnvironment(t)
	if env.Runner().ProfileDir != env.ProfileDir {
		t.Errorf("runner profile directory %q, want %q", env.Runner().ProfileDir, env.ProfileDir)
	}
}
