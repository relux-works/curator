// Package inside implements the in-domain agent: the process that runs inside a
// probe domain and attempts the operations each guarantee claims are denied.
//
// The agent reports what it observed, never what it expected. An attempt that
// fails with ENOENT or EBADF is reported as inconclusive, not as a denial,
// because "the file was not there" is not evidence that the kernel refused.
// Deciding what an observation means is the caller's job, in package probe.
package inside

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Outcome classifies one attempt.
const (
	// OutcomeAllowed means the operation succeeded: the guarantee does not hold
	// for this attempt.
	OutcomeAllowed = "allowed"
	// OutcomeDenied means the kernel refused with a permission error.
	OutcomeDenied = "denied"
	// OutcomeInconclusive means the operation failed for a reason that is not a
	// refusal, so it proves nothing either way.
	OutcomeInconclusive = "inconclusive"
)

// Attempt is one observed operation.
type Attempt struct {
	Name    string `json:"name"`
	Target  string `json:"target,omitempty"`
	Outcome string `json:"outcome"`
	Errno   string `json:"errno,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

// Report is what the in-domain agent writes to stdout.
type Report struct {
	Op       string            `json:"op"`
	PID      int               `json:"pid"`
	PGID     int               `json:"pgid"`
	SID      int               `json:"sid"`
	Attempts []Attempt         `json:"attempts"`
	Values   map[string]string `json:"values,omitempty"`
	// Bounds carries the aggregate-bound measurements. It is empty for every op
	// but bound-matrix and bound-stress; those measurements are structured
	// rather than flattened into Values because the caller reduces four numbers
	// per bound and a string map would make that reduction unreadable.
	Bounds []BoundMeasurement `json:"bounds,omitempty"`
}

// Env names the environment variables the harness sets for the agent. The agent
// reads paths only from these; nothing it touches is chosen by data under test.
const (
	EnvSource     = "PROBE_SOURCE"
	EnvGoroot     = "PROBE_GOROOT"
	EnvBuildRoot  = "PROBE_BUILDROOT"
	EnvOutside    = "PROBE_OUTSIDE"
	EnvLoopback   = "PROBE_LOOPBACK"
	EnvUnixSocket = "PROBE_UNIXSOCK"
	EnvSelf       = "PROBE_SELF"
	EnvMarker     = "PROBE_MARKER"
	EnvHold       = "PROBE_HOLD_SECONDS"
	EnvRootHold   = "PROBE_ROOT_HOLD_SECONDS"
	EnvNoFileCap  = "PROBE_NOFILE_CAP"
	EnvWriteBytes = "PROBE_WRITE_BYTES"
)

// Op names the operations the agent can perform.
const (
	OpNetwork          = "network"
	OpInheritedFD      = "inherited-fd"
	OpReadOnlyView     = "read-only-view"
	OpWriteConfinement = "write-confinement"
	OpViewRestriction  = "view-restriction"
	OpExecAllowlist    = "exec-allowlist"
	OpDescendant       = "descendant"
	OpDetachedChild    = "detached-child"
	OpResourceBounds   = "resource-bounds"
	OpBoundMatrix      = "bound-matrix"
	OpBoundStress      = "bound-stress"
	OpBoundEscape      = "bound-escape"
	OpHello            = "hello"
)

// Main is the entry point of the in-domain agent. It always writes a report and
// returns 0 for "I ran and reported"; the parent interprets the observations.
// A nonzero return means the agent itself could not run, which the parent
// treats as an inconclusive probe.
func Main(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "inside: missing op")
		return 2
	}
	op := args[0]
	report := Report{
		Op:     op,
		PID:    os.Getpid(),
		PGID:   getpgid(),
		SID:    getsid(),
		Values: map[string]string{},
	}

	switch op {
	case OpHello:
		report.Values["hello"] = "ok"
		if count, ok := countForHello(); ok {
			report.Values["descriptors"] = strconv.Itoa(count)
		}
	case OpNetwork:
		report.Attempts = networkAttempts()
	case OpInheritedFD:
		report.Attempts = inheritedFDAttempts()
	case OpReadOnlyView:
		report.Attempts = readOnlyViewAttempts()
	case OpWriteConfinement:
		report.Attempts = writeConfinementAttempts()
	case OpViewRestriction:
		report.Attempts = viewRestrictionAttempts()
	case OpExecAllowlist:
		report.Attempts = execAllowlistAttempts()
	case OpDescendant:
		report.Attempts = descendantAttempts(&report)
	case OpDetachedChild:
		return detachedChild()
	case OpResourceBounds:
		report.Attempts = resourceBoundAttempts(&report)
	case OpBoundMatrix:
		report.Attempts = boundMatrix(&report)
	case OpBoundStress:
		report.Attempts = boundStress(&report)
	case OpBoundEscape:
		report.Attempts = boundEscape(&report)
	default:
		fmt.Fprintf(os.Stderr, "inside: unknown op %q\n", op)
		return 2
	}

	data, err := json.Marshal(report)
	if err != nil {
		fmt.Fprintf(os.Stderr, "inside: marshal: %v\n", err)
		return 2
	}
	if _, err := os.Stdout.Write(append(data, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "inside: write report: %v\n", err)
		return 2
	}
	return 0
}

func getpgid() int {
	pgid, err := syscall.Getpgid(0)
	if err != nil {
		return -1
	}
	return pgid
}

func getsid() int {
	sid, err := syscall.Getsid(0)
	if err != nil {
		return -1
	}
	return sid
}

// classify turns a syscall error into an outcome. Only a permission refusal
// counts as a denial; everything else is inconclusive.
func classify(err error) (string, string) {
	if err == nil {
		return OutcomeAllowed, ""
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		switch errno {
		case syscall.EPERM, syscall.EACCES:
			return OutcomeDenied, errnoName(errno)
		default:
			return OutcomeInconclusive, errnoName(errno)
		}
	}
	// net.OpError and friends wrap an errno; errors.As above finds it. What is
	// left is a timeout or a Go-level failure.
	return OutcomeInconclusive, ""
}

func errnoName(errno syscall.Errno) string {
	switch errno {
	case syscall.EPERM:
		return "EPERM"
	case syscall.EACCES:
		return "EACCES"
	case syscall.ENOENT:
		return "ENOENT"
	case syscall.EBADF:
		return "EBADF"
	case syscall.EMFILE:
		return "EMFILE"
	case syscall.EAGAIN:
		return "EAGAIN"
	case syscall.ENOMEM:
		return "ENOMEM"
	case syscall.EINVAL:
		return "EINVAL"
	case syscall.ENOTDIR:
		return "ENOTDIR"
	case syscall.EROFS:
		return "EROFS"
	case syscall.ECONNREFUSED:
		return "ECONNREFUSED"
	case syscall.ENETDOWN:
		return "ENETDOWN"
	case syscall.EHOSTUNREACH:
		return "EHOSTUNREACH"
	case syscall.EAFNOSUPPORT:
		return "EAFNOSUPPORT"
	case syscall.ENOSYS:
		return "ENOSYS"
	default:
		return fmt.Sprintf("errno(%d)", int(errno))
	}
}

func attempt(name, target string, err error) Attempt {
	outcome, code := classify(err)
	a := Attempt{Name: name, Target: target, Outcome: outcome, Errno: code}
	if err != nil {
		a.Detail = err.Error()
	}
	return a
}

// ---------------------------------------------------------------- network

const dialTimeout = 750 * time.Millisecond

func networkAttempts() []Attempt {
	var out []Attempt

	// Socket creation itself. Seatbelt denies the operation, not necessarily
	// the descriptor, so this is recorded separately from connect.
	for _, sock := range []struct {
		name   string
		domain int
		typ    int
	}{
		{"socket-inet-stream", syscall.AF_INET, syscall.SOCK_STREAM},
		{"socket-inet6-stream", syscall.AF_INET6, syscall.SOCK_STREAM},
		{"socket-inet-dgram", syscall.AF_INET, syscall.SOCK_DGRAM},
		{"socket-unix-stream", syscall.AF_UNIX, syscall.SOCK_STREAM},
	} {
		fd, err := syscall.Socket(sock.domain, sock.typ, 0)
		if err == nil {
			_ = syscall.Close(fd)
		}
		out = append(out, attempt(sock.name, "", err))
	}

	loopback := os.Getenv(EnvLoopback)
	if loopback != "" {
		out = append(out, dialAttempt("connect-loopback-tcp", "tcp", loopback))
	}
	// A documentation-range address (RFC 5737 TEST-NET-1). It is never routed,
	// so this attempt cannot reach a third party even when the domain is open.
	out = append(out, dialAttempt("connect-offhost-tcp", "tcp", "192.0.2.1:80"))
	out = append(out, dialAttempt("connect-offhost-udp", "udp", "192.0.2.1:53"))

	if sock := os.Getenv(EnvUnixSocket); sock != "" {
		out = append(out, dialAttempt("connect-unix", "unix", sock))
	}

	// Inbound: binding a listener is a network operation too.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err == nil {
		_ = listener.Close()
	}
	out = append(out, attempt("listen-loopback-tcp", "127.0.0.1:0", err))

	// A socketpair is not a network endpoint; it is included so a profile that
	// denies too much (breaking ordinary IPC) is distinguishable from one that
	// denies exactly the network.
	pair, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err == nil {
		_ = syscall.Close(pair[0])
		_ = syscall.Close(pair[1])
	}
	out = append(out, attempt("socketpair-unix", "", err))

	return out
}

func dialAttempt(name, network, address string) Attempt {
	conn, err := net.DialTimeout(network, address, dialTimeout)
	if err == nil {
		_ = conn.Close()
	}
	return attempt(name, address, err)
}

// ------------------------------------------------------- inherited endpoint

func inheritedFDAttempts() []Attempt {
	var out []Attempt

	// fd 3 is the connected socket the harness passed in through ExtraFiles.
	inherited := os.NewFile(3, "inherited-endpoint")
	_, err := inherited.Write([]byte("probe\n"))
	out = append(out, attempt("write-inherited-endpoint", "fd:3", err))

	// fd 9 was never passed. Its EBADF proves the agent's write path really
	// distinguishes a usable descriptor from an unusable one, so an "allowed"
	// above is not an artifact of the agent ignoring errors.
	unpassed := os.NewFile(9, "never-passed")
	_, err = unpassed.Write([]byte("probe\n"))
	out = append(out, attempt("write-unpassed-descriptor", "fd:9", err))

	return out
}

// ------------------------------------------------------------ filesystem

// mutations are the operations section 5.2 requires the kernel to refuse
// against a read-only view, and section 5.3 requires it to refuse outside the
// private build root.
func mutations(dir string) []Attempt {
	var out []Attempt

	existing := firstRegularFile(dir)
	if existing != "" {
		f, err := os.OpenFile(existing, os.O_WRONLY, 0) //nolint:gosec // the path is a file the harness created inside its own stand-in view
		if err == nil {
			_ = f.Close()
		}
		out = append(out, attempt("open-write-existing", existing, err))

		f, err = os.OpenFile(existing, os.O_WRONLY|os.O_TRUNC, 0) //nolint:gosec // same harness-owned path as above
		if err == nil {
			_ = f.Close()
		}
		out = append(out, attempt("open-truncate-existing", existing, err))

		out = append(out, attempt("chmod-existing", existing, os.Chmod(existing, 0o600)))
		out = append(out, attempt("chown-existing", existing, os.Chown(existing, os.Getuid(), os.Getgid())))
		out = append(out, attempt("setxattr-existing", existing,
			setxattr(existing, "user.probe", []byte("x"))))

		// The destructive mutations each undo themselves, and unlink runs last.
		// Against a read-only view none of them succeeds, so the order is
		// invisible; against a writable one — which is how the negative control
		// proves these operations can succeed at all — an un-undone rename would
		// leave the later attempts with no target, and their ENOENT would be
		// recorded as "could not be evaluated" instead of "permitted".
		renamed := existing + ".renamed"
		renameErr := os.Rename(existing, renamed)
		out = append(out, attempt("rename-existing", existing, renameErr))
		if renameErr == nil {
			if err := os.Rename(renamed, existing); err != nil {
				// The view is now missing its sample file. Say so rather than
				// let every later attempt report a bare ENOENT.
				out = append(out, Attempt{
					Name:    "restore-renamed",
					Target:  renamed,
					Outcome: OutcomeInconclusive,
					Detail:  err.Error(),
				})
			}
		}

		link := existing + ".link"
		linkErr := os.Link(existing, link)
		out = append(out, attempt("hardlink-existing", existing, linkErr))
		if linkErr == nil {
			_ = os.Remove(link) // best effort: the view is restored by the caller
		}

		out = append(out, attempt("unlink-existing", existing, os.Remove(existing)))
	} else {
		out = append(out, Attempt{
			Name:    "open-write-existing",
			Target:  dir,
			Outcome: OutcomeInconclusive,
			Detail:  "no readable regular file found in the view",
		})
	}

	created := filepath.Join(dir, "probe-created")
	f, err := os.OpenFile(created, os.O_CREATE|os.O_WRONLY, 0o600) //nolint:gosec // the path is built from a directory the harness chose, never from data under test
	if err == nil {
		_ = f.Close()
		_ = os.Remove(created)
	}
	out = append(out, attempt("create-file", created, err))

	newDir := filepath.Join(dir, "probe-created-dir")
	err = os.Mkdir(newDir, 0o700)
	if err == nil {
		_ = os.Remove(newDir)
	}
	out = append(out, attempt("mkdir", newDir, err))

	link := filepath.Join(dir, "probe-symlink")
	err = os.Symlink("/etc/hosts", link)
	if err == nil {
		_ = os.Remove(link)
	}
	out = append(out, attempt("symlink", link, err))

	return out
}

// firstRegularFile returns a readable regular file inside dir, preferring the
// shallowest one so the walk stays cheap.
func firstRegularFile(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			return filepath.Join(dir, entry.Name())
		}
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if found := firstRegularFile(filepath.Join(dir, entry.Name())); found != "" {
				return found
			}
		}
	}
	return ""
}

func readOnlyViewAttempts() []Attempt {
	var out []Attempt
	for _, view := range []struct{ label, env string }{
		{"source", EnvSource},
		{"toolchain", EnvGoroot},
	} {
		dir := os.Getenv(view.env)
		if dir == "" {
			continue
		}
		// A view that cannot be read is not a read-only view; record the read
		// side so a profile that simply hid the tree is not mistaken for one
		// that presented it read-only.
		out = append(out, attempt(view.label+":read-dir", dir, readDirErr(dir)))
		for _, a := range mutations(dir) {
			a.Name = view.label + ":" + a.Name
			out = append(out, a)
		}
	}
	return out
}

func readDirErr(dir string) error {
	_, err := os.ReadDir(dir)
	return err
}

func writeConfinementAttempts() []Attempt {
	buildRoot := os.Getenv(EnvBuildRoot)
	outside := os.Getenv(EnvOutside)

	var out []Attempt

	// Positive control: the private build root must be writable, otherwise the
	// confinement result is trivially "everything is denied".
	inside := filepath.Join(buildRoot, "inside-write")
	out = append(out, attempt("write-inside-build-root", inside,
		os.WriteFile(inside, []byte("ok"), 0o600)))

	targets := []struct{ name, path string }{
		{"write-absolute-outside", filepath.Join(outside, "escape-absolute")},
		{"write-home", filepath.Join(os.Getenv("HOME"), ".probe-escape")},
		{"write-tmp", "/tmp/probe-escape"},
		{"write-private-tmp", "/private/tmp/probe-escape"},
		{"write-build-root-parent", filepath.Join(filepath.Dir(buildRoot), "escape-parent")},
		{"write-relative-traversal", filepath.Join(buildRoot, "..", "escape-traversal")},
	}
	if tmpdir := os.Getenv("TMPDIR"); tmpdir != "" {
		targets = append(targets, struct{ name, path string }{"write-tmpdir-env", filepath.Join(tmpdir, "probe-escape")})
	}
	for _, target := range targets {
		out = append(out, attempt(target.name, target.path,
			os.WriteFile(target.path, []byte("escape"), 0o600)))
	}

	// Symlink escape: a link created inside the build root that points out of
	// it. If the link can be created and written through, confinement is by
	// path string, not by resolved target.
	link := filepath.Join(buildRoot, "escape-symlink")
	_ = os.Remove(link)
	linkErr := os.Symlink(filepath.Join(outside, "escape-through-symlink"), link)
	out = append(out, attempt("create-escape-symlink", link, linkErr))
	if linkErr == nil {
		out = append(out, attempt("write-through-symlink", link,
			os.WriteFile(link, []byte("escape"), 0o600)))
	}

	// Hard link escape: link an outside file into the build root and write it.
	outsideFile := filepath.Join(outside, "outside-target")
	hard := filepath.Join(buildRoot, "escape-hardlink")
	_ = os.Remove(hard)
	hardErr := os.Link(outsideFile, hard)
	out = append(out, attempt("create-escape-hardlink", hard, hardErr))
	if hardErr == nil {
		out = append(out, attempt("write-through-hardlink", hard,
			os.WriteFile(hard, []byte("escape"), 0o600)))
	}

	return out
}

func viewRestrictionAttempts() []Attempt {
	var out []Attempt

	undeclared := []struct{ name, path string }{
		{"read-etc-passwd", "/etc/passwd"},
		{"read-etc-hosts", "/etc/hosts"},
		{"readdir-users", "/Users"},
		{"readdir-home", os.Getenv("HOME")},
		{"readdir-applications", "/Applications"},
		{"read-ssh-dir", filepath.Join(os.Getenv("HOME"), ".ssh")},
		{"stat-home", os.Getenv("HOME")},
	}
	for _, target := range undeclared {
		out = append(out, attempt(target.name, target.path, reachErr(target.path)))
	}

	// The root directory is granted read access because the loader needs it.
	// Enumerating it is therefore expected to succeed, and the probe records
	// that explicitly rather than leaving it unmeasured.
	out = append(out, attempt("readdir-root", "/", readDirErr("/")))

	return out
}

// reachErr reports whether a path can be reached at all: read its content when
// it is a file, list it when it is a directory, stat it otherwise.
func reachErr(path string) error {
	if path == "" {
		return syscall.ENOENT
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return readDirErr(path)
	}
	f, err := os.Open(path) //nolint:gosec // the path comes from the harness's fixed undeclared-target list, never from data under test
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	buf := make([]byte, 1)
	if _, err := f.Read(buf); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

// -------------------------------------------------------------- execution

func execAllowlistAttempts() []Attempt {
	buildRoot := os.Getenv(EnvBuildRoot)
	self := os.Getenv(EnvSelf)

	var out []Attempt

	hostPrograms := []struct{ name, path string }{
		{"exec-shell", "/bin/sh"},
		{"exec-bash", "/bin/bash"},
		{"exec-zsh", "/bin/zsh"},
		{"exec-interpreter", "/usr/bin/python3"},
		{"exec-host-binary", "/bin/ls"},
		{"exec-perl", "/usr/bin/perl"},
		{"exec-dyld-as-program", "/usr/lib/dyld"},
	}
	for _, program := range hostPrograms {
		out = append(out, attempt(program.name, program.path, runProgram(program.path, "--version")))
	}

	// A file the domain just wrote must not be executable from inside. The
	// content is a byte-identical copy of the one allowlisted program, which
	// makes this the sharpest form of the test: the allowlist must bind the
	// path, not the bytes.
	written := filepath.Join(buildRoot, "written-program")
	writeErr := copyFile(self, written)
	out = append(out, attempt("write-program-into-build-root", written, writeErr))
	if writeErr == nil {
		out = append(out, attempt("exec-self-written-copy", written, runProgram(written, "__inside", OpHello)))
		out = append(out, attempt("mmap-exec-self-written", written, mmapExec(written)))
	} else {
		out = append(out,
			Attempt{Name: "exec-self-written-copy", Target: written, Outcome: OutcomeInconclusive, Detail: writeErr.Error()},
			Attempt{Name: "mmap-exec-self-written", Target: written, Outcome: OutcomeInconclusive, Detail: writeErr.Error()})
	}

	// A symbolic link inside the build root that points at an allowlisted path.
	// If exec resolves the allowlist before resolving the link, the allowlist is
	// by name, not by identity.
	if self != "" {
		link := filepath.Join(buildRoot, "self-symlink")
		_ = os.Remove(link)
		if err := os.Symlink(self, link); err == nil {
			out = append(out, attempt("exec-symlink-to-allowlisted", link, runProgram(link, "__inside", OpHello)))
		} else {
			out = append(out, Attempt{Name: "exec-symlink-to-allowlisted", Target: link, Outcome: OutcomeInconclusive, Detail: err.Error()})
		}

		// Positive control: the allowlisted path itself must start, otherwise a
		// clean sweep of denials proves nothing.
		out = append(out, attempt("exec-allowlisted-self", self, runProgram(self, "__inside", OpHello)))
	}

	return out
}

// pipeStdio wires a command to pipes rather than leaving a stream nil.
//
// os/exec opens /dev/null for any nil stream, and a (deny default) profile
// denies that open. The failure surfaces as EPERM from Start, which is
// indistinguishable from the execution denial these probes are trying to
// observe. Wiring every stream explicitly keeps the observation attributable to
// exec alone.
func pipeStdio(cmd *exec.Cmd) {
	cmd.Stdin = bytes.NewReader(nil)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
}

func runProgram(path string, args ...string) error {
	cmd := exec.Command(path, args...)
	pipeStdio(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		<-done
		return nil
	}
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src) //nolint:gosec // src is the allowlisted agent path the harness passed in
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o700) //nolint:gosec // the copy must stay executable: the probe is whether it can be started
}

// mmapExec maps a file with PROT_EXEC. It is the loader-free way to get bytes
// from the private build root into an executable mapping, so it is the escape
// that an exec-only allowlist would miss.
//
// Measured caveat: on Apple Silicon this returns EPERM for every file, a signed
// system binary included, unless the calling process carries a JIT entitlement.
// The escape therefore fails for a platform reason and no profile can be
// credited with stopping it, so the class probe records the attribution rather
// than counting it as an allowlist result. On a host where an executable
// mapping is obtainable, this attempt is the one that would expose an
// exec-only allowlist.
func mmapExec(path string) error {
	f, err := os.Open(path) //nolint:gosec // path is the file the agent just wrote into its own build root
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	data, err := syscall.Mmap(int(f.Fd()), 0, syscall.Getpagesize(),
		syscall.PROT_READ|syscall.PROT_EXEC, syscall.MAP_PRIVATE)
	if err != nil {
		return err
	}
	// The mapping succeeded, which is the whole observation; failing to unmap
	// it would not change what was observed.
	_ = syscall.Munmap(data)
	return nil
}

// ------------------------------------------------- descendants and domains

func descendantAttempts(report *Report) []Attempt {
	self := os.Getenv(EnvSelf)
	marker := os.Getenv(EnvMarker)
	hold := holdSeconds()

	var out []Attempt
	if self == "" || marker == "" {
		return append(out, Attempt{
			Name:    "spawn-detached-descendant",
			Outcome: OutcomeInconclusive,
			Detail:  "PROBE_SELF or PROBE_MARKER not set",
		})
	}

	// A descendant that calls setsid leaves the process group and session the
	// supervisor created. If it survives a group-directed kill, group and
	// session membership are not a domain.
	detached, detachedErr := startChild(self, marker+".detached", hold, true)
	out = append(out, attempt("spawn-detached-descendant", marker+".detached", detachedErr))

	// Negative control: an identical descendant that does not detach. If the
	// group kill fails to reach this one too, the measurement method is broken
	// and the detached result proves nothing.
	attached, attachedErr := startChild(self, marker+".attached", hold, false)
	out = append(out, attempt("spawn-attached-descendant", marker+".attached", attachedErr))

	report.Values["detached_pid"] = strconv.Itoa(detached)
	report.Values["attached_pid"] = strconv.Itoa(attached)
	report.Values["domain_pgid"] = strconv.Itoa(getpgid())
	report.Values["domain_sid"] = strconv.Itoa(getsid())

	// Give both children time to write their markers before the parent looks.
	waitForFile(marker+".detached", 3*time.Second)
	waitForFile(marker+".attached", 3*time.Second)

	// Hold the domain root open when the supervisor asked for it. The wall-clock
	// probe needs a domain that is still working when its deadline arrives:
	// without the hold, the root would exit on its own and a supervisor could
	// not tell a deadline that fired from work that simply finished.
	//
	// The report is written by the caller after this returns, so a root that is
	// killed mid-hold produces no report. That is deliberate: the descendants
	// have already written their markers, and the supervisor reads those.
	if seconds := rootHoldSeconds(); seconds > 0 {
		time.Sleep(time.Duration(seconds) * time.Second)
	}
	return out
}

func rootHoldSeconds() int {
	if raw := os.Getenv(EnvRootHold); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return n
		}
	}
	return 0
}

func startChild(self, marker string, hold int, detach bool) (int, error) {
	cmd := exec.Command(self, "__inside", OpDetachedChild)
	cmd.Env = append(os.Environ(),
		EnvMarker+"="+marker,
		EnvHold+"="+strconv.Itoa(hold),
	)
	pipeStdio(cmd)
	if detach {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	}
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	pid := cmd.Process.Pid
	// Release the child without reaping it, so the parent harness can observe
	// whether it survives a domain-directed kill.
	go func() { _ = cmd.Wait() }()
	return pid, nil
}

// detachedChild is the descendant itself. It records its identity and then
// holds, so the harness can try to kill it as part of a domain.
func detachedChild() int {
	marker := os.Getenv(EnvMarker)
	hold := holdSeconds()
	if marker == "" {
		return 2
	}
	content := fmt.Sprintf("pid=%d pgid=%d sid=%d ppid=%d\n",
		os.Getpid(), getpgid(), getsid(), os.Getppid())
	if err := os.WriteFile(marker, []byte(content), 0o600); err != nil {
		return 2
	}
	deadline := time.Now().Add(time.Duration(hold) * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
	}
	// Record that the hold elapsed without the process being killed. The parent
	// checks liveness well before this, so this file only ever confirms an
	// escape that already happened.
	_ = os.WriteFile(marker+".survived", []byte("survived\n"), 0o600)
	return 0
}

func holdSeconds() int {
	if raw := os.Getenv(EnvHold); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return n
		}
	}
	return 20
}

func waitForFile(path string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

// ------------------------------------------------------------- resources

// Bounds on the descriptor cap the harness may declare. They are validated here
// rather than trusted, so a malformed value cannot turn the descriptor count
// into an unbounded loop or a negative conversion.
const (
	defaultNoFileCap = 64
	maxNoFileCap     = 1 << 16
)

func resourceBoundAttempts(report *Report) []Attempt {
	var out []Attempt

	// The declared cap is bounded on both sides: below 8 the agent could not
	// open its own report, and above maxNoFileCap the count loop would run long
	// enough to look like a hang rather than a measurement.
	cap64 := defaultNoFileCap
	if raw := os.Getenv(EnvNoFileCap); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 8 && n <= maxNoFileCap {
			cap64 = n
		}
	}

	var original syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &original); err != nil {
		out = append(out, Attempt{Name: "getrlimit-nofile", Outcome: OutcomeInconclusive, Detail: err.Error()})
		return out
	}
	report.Values["rlimit_nofile_soft"] = strconv.FormatUint(original.Cur, 10)
	report.Values["rlimit_nofile_hard"] = strconv.FormatUint(original.Max, 10)

	limit := syscall.Rlimit{Cur: uint64(cap64), Max: original.Max} //nolint:gosec // cap64 is bounded to (8, maxNoFileCap] above, so the conversion cannot overflow or go negative
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		out = append(out, Attempt{Name: "setrlimit-nofile", Outcome: OutcomeInconclusive, Detail: err.Error()})
		return out
	}
	out = append(out, Attempt{Name: "setrlimit-nofile", Outcome: OutcomeAllowed,
		Detail: fmt.Sprintf("soft limit set to %d", cap64)})

	self := countOpenable()
	report.Values["self_descriptors"] = strconv.Itoa(self)

	// A child inherits the same soft limit and gets its own budget under it.
	// If the child can open about as many descriptors as the parent did, the
	// limit is per process and the two budgets add up: there is no aggregate.
	child := -1
	if selfPath := os.Getenv(EnvSelf); selfPath != "" {
		child = childDescriptorCount(selfPath, cap64)
	}
	report.Values["child_descriptors"] = strconv.Itoa(child)
	if child > 0 {
		out = append(out, Attempt{
			Name:    "descriptor-budget-is-per-process",
			Outcome: OutcomeAllowed,
			Detail: fmt.Sprintf("soft limit %d; this process opened %d and a child opened %d, aggregate %d",
				cap64, self, child, self+child),
		})
	} else {
		out = append(out, Attempt{
			Name:    "descriptor-budget-is-per-process",
			Outcome: OutcomeInconclusive,
			Detail:  "child descriptor count unavailable",
		})
	}

	// Bytes written below the private build root. macOS exposes no per-directory
	// byte cap, so this attempt measures whether anything stops a write past a
	// declared budget.
	if buildRoot := os.Getenv(EnvBuildRoot); buildRoot != "" {
		budget := int64(4 << 20)
		if raw := os.Getenv(EnvWriteBytes); raw != "" {
			if n, err := strconv.ParseInt(raw, 10, 64); err == nil && n > 0 {
				budget = n
			}
		}
		written, err := writePast(filepath.Join(buildRoot, "byte-budget.bin"), budget)
		report.Values["bytes_written"] = strconv.FormatInt(written, 10)
		report.Values["byte_budget"] = strconv.FormatInt(budget, 10)
		out = append(out, attempt("write-past-declared-byte-budget",
			filepath.Join(buildRoot, "byte-budget.bin"), err))
	}

	// Restore the inherited limit so a later probe in the same process is not
	// measured under this one's cap. A failure here cannot invalidate what was
	// already observed, and the agent exits immediately after.
	_ = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &original)
	return out
}

// countOpenable opens descriptors until the per-process limit refuses one, then
// closes them all and reports how many it got.
func countOpenable() int {
	var open []*os.File
	defer func() {
		for _, f := range open {
			_ = f.Close()
		}
	}()
	for len(open) < 1<<16 {
		f, err := os.Open("/")
		if err != nil {
			break
		}
		open = append(open, f)
	}
	return len(open)
}

func childDescriptorCount(self string, cap64 int) int {
	cmd := exec.Command(self, "__inside", OpHello)
	cmd.Env = append(os.Environ(), "PROBE_COUNT_DESCRIPTORS="+strconv.Itoa(cap64))
	cmd.Stdin = bytes.NewReader(nil)
	cmd.Stderr = io.Discard
	output, err := cmd.Output()
	if err != nil {
		return -1
	}
	var report Report
	if err := json.Unmarshal(output, &report); err != nil {
		return -1
	}
	count, err := strconv.Atoi(report.Values["descriptors"])
	if err != nil {
		return -1
	}
	return count
}

func writePast(path string, budget int64) (int64, error) {
	f, err := os.Create(path) //nolint:gosec // path is built from the harness-chosen build root
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(path)
	}()
	chunk := make([]byte, 64<<10)
	var total int64
	// Deliberately overshoot the budget: a host that enforces an aggregate
	// write-byte bound refuses somewhere before the end of this loop.
	for total < budget*2 {
		n, err := f.Write(chunk)
		total += int64(n)
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// countForHello lets the hello op double as the descriptor-counting child, so
// the resource probe does not need a second allowlisted executable.
func countForHello() (int, bool) {
	raw := os.Getenv("PROBE_COUNT_DESCRIPTORS")
	if raw == "" {
		return 0, false
	}
	cap64, err := strconv.Atoi(raw)
	if err != nil || cap64 <= 8 {
		return 0, false
	}
	var original syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &original); err != nil {
		return 0, false
	}
	limit := syscall.Rlimit{Cur: uint64(cap64), Max: original.Max}
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return 0, false
	}
	return countOpenable(), true
}

// ParseReport decodes an agent report emitted on stdout. The agent may print
// nothing when it was killed, which is itself an observation, so a decode
// failure returns the raw text for the caller to record.
func ParseReport(stdout string) (Report, error) {
	trimmed := strings.TrimSpace(stdout)
	if trimmed == "" {
		return Report{}, fmt.Errorf("agent produced no report")
	}
	// The agent writes exactly one JSON line last; a descendant may have
	// written earlier lines into the same pipe.
	lines := strings.Split(trimmed, "\n")
	last := lines[len(lines)-1]
	var report Report
	if err := json.Unmarshal([]byte(last), &report); err != nil {
		return Report{}, fmt.Errorf("agent report is not JSON: %w", err)
	}
	return report, nil
}

// Find returns the attempt with the given name.
func (r Report) Find(name string) (Attempt, bool) {
	for _, a := range r.Attempts {
		if a.Name == name {
			return a, true
		}
	}
	return Attempt{}, false
}
