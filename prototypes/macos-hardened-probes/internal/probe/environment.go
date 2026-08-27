package probe

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/inside"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/seatbelt"
)

// Environment is the operation-private scaffolding a probe run needs: stand-in
// source and toolchain views, a private build root, an outside directory that
// escapes are aimed at, and loopback endpoints for the network probes.
//
// Nothing here comes from package data. Every path is created by this package
// under a directory the caller owns, which is the property section 5.3 requires
// of a real build domain and which the probes therefore have to reproduce.
type Environment struct {
	Root       string
	SourceDir  string
	GorootDir  string
	BuildRoot  string
	OutsideDir string
	ProfileDir string
	SelfPath   string
	MarkerDir  string

	LoopbackAddr string
	UnixSocket   string
	// UnixSocketRelocated reports that the unix endpoint could not be placed in
	// OutsideDir because the resulting path exceeded the platform limit, so it
	// lives in a private short-path directory instead. It is still outside the
	// build domain, which is the only property the probe depends on, but the
	// relocation is recorded rather than hidden.
	UnixSocketRelocated bool

	listener      net.Listener
	unixListener  net.Listener
	unixSocketDir string
	closeOnce     sync.Once
}

// maxUnixSocketPath is the usable length of sockaddr_un.sun_path on macOS: the
// field is 104 bytes and the last one holds the terminator.
//
// A bind past that limit fails with EINVAL, which the probes would read as an
// ordinary error rather than as a broken measurement. The length is therefore
// checked before binding instead of being discovered from the errno.
const maxUnixSocketPath = 103

// NewEnvironment builds the scaffolding under root.
//
// root and selfPath are resolved to absolute, symlink-free paths first. A
// seatbelt filter matches the resolved path the kernel sees, so a relative root
// or one reached through a symlink (which is what /tmp and /var are on macOS)
// matches nothing and silently turns every probe into a denial. That failure
// mode looks exactly like perfect enforcement, so it is rejected here rather
// than left to be noticed in the results.
func NewEnvironment(root, selfPath string) (*Environment, error) {
	root, err := resolvePath(root)
	if err != nil {
		return nil, fmt.Errorf("probe environment: work directory: %w", err)
	}
	selfPath, err = resolveExisting(selfPath)
	if err != nil {
		return nil, fmt.Errorf("probe environment: probe binary: %w", err)
	}

	env := &Environment{
		Root:       root,
		SourceDir:  filepath.Join(root, "source"),
		GorootDir:  filepath.Join(root, "goroot"),
		BuildRoot:  filepath.Join(root, "buildroot"),
		OutsideDir: filepath.Join(root, "outside"),
		ProfileDir: filepath.Join(root, "profiles"),
		MarkerDir:  filepath.Join(root, "markers"),
		SelfPath:   selfPath,
	}
	for _, dir := range []string{env.SourceDir, env.GorootDir, env.BuildRoot, env.OutsideDir, env.ProfileDir, env.MarkerDir} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("probe environment: %w", err)
		}
	}

	// Stand-in source snapshot and toolchain. Their content is irrelevant; what
	// matters is that a mutation attempt has something real to aim at.
	files := map[string]string{
		filepath.Join(env.SourceDir, "main.go"):         "package main\n\nfunc main() {}\n",
		filepath.Join(env.SourceDir, "go.mod"):          "module probe\n\ngo 1.25.5\n",
		filepath.Join(env.GorootDir, "bin", "go"):       "#!/bin/sh\nexit 0\n",
		filepath.Join(env.GorootDir, "VERSION"):         "go1.25.5\n",
		filepath.Join(env.OutsideDir, "outside-target"): "outside\n",
	}
	for path, content := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return nil, fmt.Errorf("probe environment: %w", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			return nil, fmt.Errorf("probe environment: %w", err)
		}
	}

	listener, listenErr := net.Listen("tcp", "127.0.0.1:0")
	if listenErr != nil {
		return nil, fmt.Errorf("probe environment: loopback listener: %w", listenErr)
	}
	env.listener = listener
	env.LoopbackAddr = listener.Addr().String()
	go acceptForever(listener)

	if err := env.listenUnix(); err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("probe environment: unix listener: %w", err)
	}

	return env, nil
}

// listenUnix binds the unix endpoint the network probes dial.
//
// It is placed in OutsideDir so it shares the lifetime of the rest of the
// scaffolding, but a work directory deep enough to push the socket path past
// maxUnixSocketPath makes that impossible. Rather than fail the whole run over
// a path length, the endpoint moves to a short private directory, which is
// still outside the domain, and the move is recorded on the Environment.
func (e *Environment) listenUnix() error {
	path := filepath.Join(e.OutsideDir, "probe.sock")
	if len(path) > maxUnixSocketPath {
		dir, err := os.MkdirTemp("", "hp")
		if err != nil {
			return err
		}
		e.unixSocketDir = dir
		e.UnixSocketRelocated = true
		path = filepath.Join(dir, "p.sock")
		if len(path) > maxUnixSocketPath {
			// Best effort: the caller is about to fail anyway, and a stranded
			// temporary directory is not part of any observation.
			_ = os.RemoveAll(dir)
			return fmt.Errorf("no usable socket path: %q is %d bytes, limit is %d",
				path, len(path), maxUnixSocketPath)
		}
	}
	listener, err := net.Listen("unix", path)
	if err != nil {
		if e.unixSocketDir != "" {
			_ = os.RemoveAll(e.unixSocketDir)
			e.unixSocketDir = ""
			e.UnixSocketRelocated = false
		}
		return err
	}
	e.UnixSocket = path
	e.unixListener = listener
	go acceptForever(listener)
	return nil
}

// resolvePath makes a directory path absolute and symlink-free, creating it
// first when it does not exist yet so EvalSymlinks has something to resolve.
func resolvePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("empty path")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		if err := os.MkdirAll(abs, 0o700); err != nil {
			return "", err
		}
	}
	return filepath.EvalSymlinks(abs)
}

// resolveExisting resolves a path that must already exist as a regular file.
//
// It is deliberately not resolvePath: creating a missing agent binary as an
// empty directory would let NewEnvironment succeed and turn every later class
// probe into an exec failure reported as "the host cannot do this". A probe
// binary that is not there is a harness fault and has to be named as one here.
func resolveExisting(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("empty path")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a regular file", resolved)
	}
	if info.Mode()&0o111 == 0 {
		return "", fmt.Errorf("%s is not executable", resolved)
	}
	return resolved, nil
}

// acceptForever drains whatever the probe endpoints receive.
//
// The probes only need the endpoint to exist and to complete a connect; nothing
// reads what a domain member sent, so both the copy and the close are best
// effort and their errors carry no observation.
func acceptForever(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		go func() {
			_, _ = io.Copy(io.Discard, conn)
			_ = conn.Close()
		}()
	}
}

// Close releases the listeners. Files under Root are left in place so the
// caller can keep them as evidence; a relocated socket directory is not part of
// the evidence and is removed.
func (e *Environment) Close() {
	e.closeOnce.Do(func() {
		if e.listener != nil {
			_ = e.listener.Close()
		}
		if e.unixListener != nil {
			_ = e.unixListener.Close()
		}
		if e.unixSocketDir != "" {
			_ = os.RemoveAll(e.unixSocketDir)
		}
	})
}

// AgentEnv is the environment handed to the in-domain agent. It is a fixed,
// closed list: the agent inherits nothing else, so no ambient value can change
// what it attempts.
func (e *Environment) AgentEnv(extra ...string) []string {
	env := []string{
		inside.EnvSource + "=" + e.SourceDir,
		inside.EnvGoroot + "=" + e.GorootDir,
		inside.EnvBuildRoot + "=" + e.BuildRoot,
		inside.EnvOutside + "=" + e.OutsideDir,
		inside.EnvSelf + "=" + e.SelfPath,
		inside.EnvLoopback + "=" + e.LoopbackAddr,
		inside.EnvUnixSocket + "=" + e.UnixSocket,
		"HOME=" + os.Getenv("HOME"),
		"TMPDIR=" + os.Getenv("TMPDIR"),
		"PATH=",
	}
	return append(env, extra...)
}

// Runner returns a seatbelt runner writing profiles into this environment.
func (e *Environment) Runner() seatbelt.Runner {
	return seatbelt.Runner{ProfileDir: e.ProfileDir}
}

// ResetBuildRoot empties and recreates the private build root between probes so
// a leftover file from one probe cannot make the next one pass.
func (e *Environment) ResetBuildRoot() error {
	if err := os.RemoveAll(e.BuildRoot); err != nil {
		return err
	}
	return os.MkdirAll(e.BuildRoot, 0o700)
}

// CleanEscapes removes files an escape attempt may have created outside the
// build root, so a rerun starts from the same state.
func (e *Environment) CleanEscapes() {
	for _, path := range []string{
		filepath.Join(e.OutsideDir, "escape-absolute"),
		filepath.Join(e.OutsideDir, "escape-through-symlink"),
		filepath.Join(os.Getenv("HOME"), ".probe-escape"),
		"/tmp/probe-escape",
		"/private/tmp/probe-escape",
		filepath.Join(e.Root, "escape-parent"),
		filepath.Join(e.Root, "escape-traversal"),
	} {
		// Best effort: most of these never existed, and the write-confinement
		// probe restates the same question by stat-ing them after the run.
		_ = os.Remove(path)
	}
	if tmpdir := os.Getenv("TMPDIR"); tmpdir != "" {
		_ = os.Remove(filepath.Join(tmpdir, "probe-escape"))
	}
}

// DescribeHost collects the host facts an evidence reader needs.
func DescribeHost() HostInfo {
	info := HostInfo{
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		UID:       os.Getuid(),
	}
	info.ProductName = swVers("productName")
	info.ProductVersion = swVers("productVersion")
	info.BuildVersion = swVers("buildVersion")
	info.KernelVersion = trimmedOutput("/usr/bin/uname", "-v")
	info.SIPStatus = trimmedOutput("/usr/bin/csrutil", "status")
	return info
}

func swVers(field string) string {
	return trimmedOutput("/usr/bin/sw_vers", "--"+field)
}

func trimmedOutput(path string, args ...string) string {
	out, err := exec.Command(path, args...).Output()
	if err != nil {
		return "unavailable: " + err.Error()
	}
	return strings.TrimSpace(strings.ReplaceAll(string(out), "\n", " "))
}
