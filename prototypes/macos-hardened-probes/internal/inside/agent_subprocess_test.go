package inside

import (
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
)

// Two agent operations cannot be exercised in the test process itself:
//
//   - inherited-fd writes to file descriptors 3 and 9 by number, and in a test
//     binary those belong to the Go runtime;
//   - resource-bounds lowers RLIMIT_NOFILE and then opens descriptors until the
//     kernel refuses, which would starve the runtime while it holds.
//
// Both are therefore driven the way the harness drives them: as a subprocess of
// the built agent. The coverage cost is real and deliberate — running them
// in-process would test a situation the harness never creates.

var probeBinary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "inside-agent-")
	if err != nil {
		panic("agent tests: temp dir: " + err.Error())
	}
	defer func() { _ = os.RemoveAll(dir) }()

	binary := filepath.Join(dir, "hardened-probe")
	build := exec.Command("go", "build", "-o", binary, "./cmd/hardened-probe")
	build.Dir = filepath.Join("..", "..")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		panic("agent tests: build the probe binary: " + err.Error())
	}
	probeBinary = binary

	os.Exit(m.Run())
}

// buildProbeBinary returns the agent every subprocess test drives.
//
// It fails rather than skips when the binary is missing. A suite that skips its
// subprocess tests because the thing under test would not build reports green
// while measuring nothing, which is the one failure an evidence harness must
// not have.
func buildProbeBinary(t *testing.T) string {
	t.Helper()
	if probeBinary == "" {
		t.Fatal("the probe binary was not built, so no subprocess test measured anything")
	}
	return probeBinary
}

// runAgent invokes the built agent exactly as the harness does and parses the
// report it wrote to stdout.
func runAgent(t *testing.T, op string, env []string, extraFiles []*os.File) Report {
	t.Helper()
	cmd := exec.Command(buildProbeBinary(t), "__inside", op)
	cmd.Env = env
	cmd.ExtraFiles = extraFiles
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		t.Fatalf("agent %s failed with no report: %v", op, err)
	}
	report, parseErr := ParseReport(string(out))
	if parseErr != nil {
		t.Fatalf("agent %s report: %v (stdout %q)", op, parseErr, out)
	}
	if report.Op != op {
		t.Fatalf("agent reported op %q, want %q", report.Op, op)
	}
	return report
}

// The inherited-endpoint probe only means something if a passed descriptor is
// usable when nothing revokes it: that is the negative-control side of the
// class, and it is what makes an in-domain denial evidence of revocation.
func TestAgentInheritedFDUncontained(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()
	accepted := make(chan struct{}, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		accepted <- struct{}{}
		_ = conn.Close()
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	<-accepted

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		t.Fatalf("dial did not yield a TCP connection")
	}
	file, err := tcpConn.File()
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	defer func() { _ = file.Close() }()

	report := runAgent(t, OpInheritedFD, []string{}, []*os.File{file})

	passed, ok := report.Find("write-inherited-endpoint")
	if !ok {
		t.Fatal("report has no write-inherited-endpoint")
	}
	if passed.Outcome != OutcomeAllowed {
		t.Errorf("an inherited endpoint was unusable with no domain applied: %q (%s: %s)",
			passed.Outcome, passed.Errno, passed.Detail)
	}

	// The unpassed descriptor is the agent's proof that it observes write
	// failures at all. Without an EBADF here, an "allowed" above could just be
	// the agent ignoring errors.
	unpassed, ok := report.Find("write-unpassed-descriptor")
	if !ok {
		t.Fatal("report has no write-unpassed-descriptor")
	}
	if unpassed.Errno != "EBADF" {
		t.Errorf("write to an unpassed descriptor = %q (%s), want EBADF", unpassed.Outcome, unpassed.Errno)
	}
	if unpassed.Outcome == OutcomeAllowed {
		t.Error("a descriptor that was never passed reported a successful write")
	}
}

// With no descriptor passed at all, fd 3 is not a socket either, so the probe
// must report a failure rather than a usable endpoint.
func TestAgentInheritedFDWithNothingPassed(t *testing.T) {
	report := runAgent(t, OpInheritedFD, []string{}, nil)

	passed, ok := report.Find("write-inherited-endpoint")
	if !ok {
		t.Fatal("report has no write-inherited-endpoint")
	}
	if passed.Outcome == OutcomeAllowed {
		t.Errorf("writing to an unpassed fd 3 reported success: %+v", passed)
	}
}

// The resource probe's finding is that macOS descriptor budgets are per
// process: a child gets a fresh allowance under the same soft limit, so the two
// add up past it. That is measured here uncontained, where no sandbox can be
// blamed for the result.
func TestAgentResourceBoundsMeasuresPerProcessBudgets(t *testing.T) {
	if testing.Short() {
		t.Skip("the resource probe opens descriptors and writes megabytes")
	}
	buildRoot := t.TempDir()
	const noFileCap = 64
	const budget = 1 << 20

	report := runAgent(t, OpResourceBounds, []string{
		EnvBuildRoot + "=" + buildRoot,
		EnvSelf + "=" + buildProbeBinary(t),
		EnvNoFileCap + "=" + strconv.Itoa(noFileCap),
		EnvWriteBytes + "=" + strconv.Itoa(budget),
		"HOME=" + t.TempDir(),
	}, nil)

	for _, key := range []string{"rlimit_nofile_soft", "rlimit_nofile_hard", "self_descriptors", "child_descriptors", "bytes_written", "byte_budget"} {
		if _, ok := report.Values[key]; !ok {
			t.Errorf("resource report has no %q: %v", key, report.Values)
		}
	}

	self, err := strconv.Atoi(report.Values["self_descriptors"])
	if err != nil {
		t.Fatalf("self_descriptors %q: %v", report.Values["self_descriptors"], err)
	}
	child, err := strconv.Atoi(report.Values["child_descriptors"])
	if err != nil {
		t.Fatalf("child_descriptors %q: %v", report.Values["child_descriptors"], err)
	}
	if self <= 0 {
		t.Fatalf("the agent opened %d descriptors under a soft limit of %d", self, noFileCap)
	}
	if child <= 0 {
		t.Fatalf("the child reported %d descriptors; the aggregate claim cannot be evaluated", child)
	}
	// The whole point: the two budgets are independent, so together they exceed
	// the single soft limit. If this ever fails, macOS grew aggregate accounting
	// and the class verdict must be re-measured, not assumed.
	if self+child <= noFileCap {
		t.Errorf("parent %d + child %d = %d descriptors under a soft limit of %d: "+
			"this host appears to account descriptors in aggregate", self, child, self+child, noFileCap)
	}

	written, err := strconv.ParseInt(report.Values["bytes_written"], 10, 64)
	if err != nil {
		t.Fatalf("bytes_written %q: %v", report.Values["bytes_written"], err)
	}
	if written <= budget {
		t.Errorf("the agent wrote %d bytes against a declared budget of %d; "+
			"a host that enforced the budget would have refused earlier", written, budget)
	}

	a, ok := report.Find("write-past-declared-byte-budget")
	if !ok {
		t.Fatal("report has no write-past-declared-byte-budget")
	}
	if a.Outcome == OutcomeDenied {
		t.Errorf("the byte budget was enforced with no mechanism applied: %+v", a)
	}
	if _, ok := report.Find("setrlimit-nofile"); !ok {
		t.Error("report has no setrlimit-nofile, so the soft limit was never established")
	}
}

// Without a build root there is nothing to write into, so the byte-budget
// attempt must be absent rather than aimed at a bare filename.
func TestAgentResourceBoundsWithoutABuildRoot(t *testing.T) {
	if testing.Short() {
		t.Skip("the resource probe opens descriptors")
	}
	report := runAgent(t, OpResourceBounds, []string{
		EnvSelf + "=" + buildProbeBinary(t),
		EnvNoFileCap + "=64",
		"HOME=" + t.TempDir(),
	}, nil)

	if _, ok := report.Find("write-past-declared-byte-budget"); ok {
		t.Error("the byte-budget attempt ran with no build root to write into")
	}
	if _, ok := report.Values["bytes_written"]; ok {
		t.Errorf("bytes_written was published with no build root: %v", report.Values)
	}
}

// The hello op doubles as the descriptor-counting child. It must count only
// when asked, so the domain smoke test stays cheap.
func TestAgentHelloCountsDescriptorsOnlyOnRequest(t *testing.T) {
	plain := runAgent(t, OpHello, []string{}, nil)
	if plain.Values["hello"] != "ok" {
		t.Errorf("hello values = %v, want an acknowledgement", plain.Values)
	}
	if _, ok := plain.Values["descriptors"]; ok {
		t.Errorf("hello counted descriptors unasked: %v", plain.Values)
	}

	counted := runAgent(t, OpHello, []string{"PROBE_COUNT_DESCRIPTORS=64"}, nil)
	raw, ok := counted.Values["descriptors"]
	if !ok {
		t.Fatalf("hello did not count when asked: %v", counted.Values)
	}
	count, err := strconv.Atoi(raw)
	if err != nil {
		t.Fatalf("descriptors %q: %v", raw, err)
	}
	if count <= 0 || count > 64 {
		t.Errorf("descriptors = %d, want a positive count under the requested soft limit of 64", count)
	}
}

// A descendant that calls setsid leaves the session its parent created. The
// harness reads this from the marker file, so the marker has to be written even
// though the child is no longer in the parent's session.
func TestAgentDetachedChildLeavesTheSession(t *testing.T) {
	markerDir := t.TempDir()
	path := filepath.Join(markerDir, "child")

	cmd := exec.Command(buildProbeBinary(t), "__inside", OpDetachedChild)
	cmd.Env = []string{EnvMarker + "=" + path, EnvHold + "=1"}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	pid := cmd.Process.Pid
	t.Cleanup(func() { syscall.Kill(pid, syscall.SIGKILL) }) //nolint:errcheck // best-effort cleanup

	if !waitForFile(path, 10_000_000_000) {
		t.Fatalf("the detached child never wrote %s", path)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("wait: %v", err)
	}

	m := readMarker(t, path)
	if m.pid != pid {
		t.Errorf("marker pid %d, want %d", m.pid, pid)
	}
	if m.sid == getsid() {
		t.Errorf("the child is in session %d, the same as the parent; setsid did not take effect", m.sid)
	}
	if m.sid != m.pid {
		t.Errorf("a session leader has sid == pid; got sid %d pid %d", m.sid, m.pid)
	}
}
