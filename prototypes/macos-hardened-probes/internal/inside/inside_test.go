package inside

import (
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
	"testing"
	"time"
)

// The agent's contract is that it reports what it observed and never what it
// expected. Most of these tests are therefore about classification: an ENOENT
// reported as a denial would manufacture evidence of enforcement out of a
// missing file.

// ------------------------------------------------------------------ classify

func TestClassifyOnlyCallsPermissionErrorsDenials(t *testing.T) {
	cases := []struct {
		name        string
		err         error
		wantOutcome string
		wantErrno   string
	}{
		{"success", nil, OutcomeAllowed, ""},
		{"EPERM", syscall.EPERM, OutcomeDenied, "EPERM"},
		{"EACCES", syscall.EACCES, OutcomeDenied, "EACCES"},
		{"ENOENT", syscall.ENOENT, OutcomeInconclusive, "ENOENT"},
		{"EBADF", syscall.EBADF, OutcomeInconclusive, "EBADF"},
		{"EMFILE", syscall.EMFILE, OutcomeInconclusive, "EMFILE"},
		{"ECONNREFUSED", syscall.ECONNREFUSED, OutcomeInconclusive, "ECONNREFUSED"},
		{"EAFNOSUPPORT", syscall.EAFNOSUPPORT, OutcomeInconclusive, "EAFNOSUPPORT"},
		{"unnamed errno", syscall.Errno(0x7f), OutcomeInconclusive, "errno(127)"},
		{"not an errno", errors.New("context deadline exceeded"), OutcomeInconclusive, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			outcome, errno := classify(tc.err)
			if outcome != tc.wantOutcome {
				t.Errorf("outcome %q, want %q", outcome, tc.wantOutcome)
			}
			if errno != tc.wantErrno {
				t.Errorf("errno %q, want %q", errno, tc.wantErrno)
			}
		})
	}
}

// The errors these probes actually see are wrapped several layers deep by
// os and net. If classify only recognised a bare errno, every real denial
// would be filed as inconclusive.
func TestClassifyUnwrapsRealErrorTypes(t *testing.T) {
	cases := []struct {
		name        string
		err         error
		wantOutcome string
		wantErrno   string
	}{
		{
			"os.PathError",
			&os.PathError{Op: "open", Path: "/etc/passwd", Err: syscall.EPERM},
			OutcomeDenied, "EPERM",
		},
		{
			"os.LinkError",
			&os.LinkError{Op: "rename", Old: "a", New: "b", Err: syscall.EACCES},
			OutcomeDenied, "EACCES",
		},
		{
			"net.OpError",
			&net.OpError{Op: "dial", Net: "tcp", Err: syscall.EPERM},
			OutcomeDenied, "EPERM",
		},
		{
			"doubly wrapped",
			fmt.Errorf("probe: %w", &os.PathError{Op: "open", Err: syscall.EACCES}),
			OutcomeDenied, "EACCES",
		},
		{
			"missing file stays inconclusive when wrapped",
			&os.PathError{Op: "stat", Path: "/nope", Err: syscall.ENOENT},
			OutcomeInconclusive, "ENOENT",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			outcome, errno := classify(tc.err)
			if outcome != tc.wantOutcome || errno != tc.wantErrno {
				t.Errorf("classify = (%q, %q), want (%q, %q)", outcome, errno, tc.wantOutcome, tc.wantErrno)
			}
		})
	}
}

func TestErrnoNameCoversTheReportedSet(t *testing.T) {
	named := map[syscall.Errno]string{
		syscall.EPERM: "EPERM", syscall.EACCES: "EACCES", syscall.ENOENT: "ENOENT",
		syscall.EBADF: "EBADF", syscall.EMFILE: "EMFILE", syscall.ENOTDIR: "ENOTDIR",
		syscall.EROFS: "EROFS", syscall.ECONNREFUSED: "ECONNREFUSED", syscall.ENETDOWN: "ENETDOWN",
		syscall.EHOSTUNREACH: "EHOSTUNREACH", syscall.EAFNOSUPPORT: "EAFNOSUPPORT", syscall.ENOSYS: "ENOSYS",
	}
	for errno, want := range named {
		if got := errnoName(errno); got != want {
			t.Errorf("errnoName(%d) = %q, want %q", int(errno), got, want)
		}
	}
	if got := errnoName(syscall.Errno(4242)); got != "errno(4242)" {
		t.Errorf("errnoName(4242) = %q, want %q", got, "errno(4242)")
	}
}

func TestAttemptRecordsTheError(t *testing.T) {
	a := attempt("probe", "/target", syscall.EPERM)
	if a.Name != "probe" || a.Target != "/target" {
		t.Errorf("attempt did not carry name and target: %+v", a)
	}
	if a.Outcome != OutcomeDenied || a.Errno != "EPERM" {
		t.Errorf("attempt = %+v, want a denial", a)
	}
	if a.Detail == "" {
		t.Error("a failed attempt carries no detail")
	}

	ok := attempt("probe", "/target", nil)
	if ok.Outcome != OutcomeAllowed || ok.Errno != "" || ok.Detail != "" {
		t.Errorf("successful attempt = %+v, want a bare allowed", ok)
	}
}

// --------------------------------------------------------------- ParseReport

func TestParseReportRejectsUnusableOutput(t *testing.T) {
	cases := []struct {
		name       string
		stdout     string
		wantReason string
	}{
		{"empty", "", "no report"},
		{"whitespace only", "   \n\t\n ", "no report"},
		{"not JSON", "sandbox-exec: operation not permitted\n", "not JSON"},
		{"truncated JSON", `{"op":"hello"`, "not JSON"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseReport(tc.stdout)
			if err == nil {
				t.Fatalf("ParseReport(%q) returned no error", tc.stdout)
			}
			if !strings.Contains(err.Error(), tc.wantReason) {
				t.Errorf("error %q does not mention %q", err, tc.wantReason)
			}
		})
	}
}

// A descendant can write into the same pipe, so the agent's own report is the
// last line, not the first. Taking the first would misread the whole run.
func TestParseReportTakesTheLastLine(t *testing.T) {
	stdout := "a descendant said something\n" +
		`{"op":"stale","pid":1,"attempts":[]}` + "\n" +
		`{"op":"hello","pid":42,"pgid":7,"sid":9,"attempts":[{"name":"x","outcome":"denied"}],"values":{"hello":"ok"}}` + "\n"

	report, err := ParseReport(stdout)
	if err != nil {
		t.Fatalf("ParseReport: %v", err)
	}
	if report.Op != "hello" || report.PID != 42 || report.PGID != 7 || report.SID != 9 {
		t.Errorf("report = %+v, want the last line", report)
	}
	if report.Values["hello"] != "ok" {
		t.Errorf("values = %v", report.Values)
	}
}

func TestParseReportToleratesTrailingWhitespace(t *testing.T) {
	report, err := ParseReport("\n  " + `{"op":"hello","attempts":[]}` + "  \n\n")
	if err != nil {
		t.Fatalf("ParseReport: %v", err)
	}
	if report.Op != "hello" {
		t.Errorf("op %q, want hello", report.Op)
	}
}

func TestReportFind(t *testing.T) {
	report := Report{Attempts: []Attempt{
		{Name: "first", Outcome: OutcomeDenied},
		{Name: "second", Outcome: OutcomeAllowed},
	}}
	if a, ok := report.Find("second"); !ok || a.Outcome != OutcomeAllowed {
		t.Errorf("Find(second) = (%+v, %v)", a, ok)
	}
	if a, ok := report.Find("absent"); ok {
		t.Errorf("Find(absent) = (%+v, true), want not found", a)
	}
	if _, ok := (Report{}).Find("anything"); ok {
		t.Error("Find on an empty report reported a hit")
	}
}

// ------------------------------------------------------------ small helpers

func TestHoldSecondsFallsBackToASafeDefault(t *testing.T) {
	cases := map[string]int{
		"":     20,
		"0":    20,
		"-5":   20,
		"junk": 20,
		"1":    1,
		"42":   42,
	}
	for raw, want := range cases {
		t.Setenv(EnvHold, raw)
		if got := holdSeconds(); got != want {
			t.Errorf("holdSeconds() with %s=%q = %d, want %d", EnvHold, raw, got, want)
		}
	}
	unsetEnv(t, EnvHold)
	if got := holdSeconds(); got != 20 {
		t.Errorf("holdSeconds() unset = %d, want 20", got)
	}
}

func TestFirstRegularFilePrefersTheShallowestEntry(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "deep", "deeper")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	buried := filepath.Join(nested, "buried.txt")
	if err := os.WriteFile(buried, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Only a nested file exists: the walk must still find it.
	if got := firstRegularFile(root); got != buried {
		t.Errorf("firstRegularFile = %q, want %q", got, buried)
	}

	// Once a shallow file exists it must win, so the probe does not walk a
	// large tree on every run.
	shallow := filepath.Join(root, "shallow.txt")
	if err := os.WriteFile(shallow, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if got := firstRegularFile(root); got != shallow {
		t.Errorf("firstRegularFile = %q, want the shallow %q", got, shallow)
	}
}

func TestFirstRegularFileWithNothingToFind(t *testing.T) {
	empty := t.TempDir()
	if got := firstRegularFile(empty); got != "" {
		t.Errorf("firstRegularFile(empty) = %q, want \"\"", got)
	}
	if got := firstRegularFile(filepath.Join(empty, "missing")); got != "" {
		t.Errorf("firstRegularFile(missing) = %q, want \"\"", got)
	}
	// A directory holding only directories is still nothing to aim at.
	if err := os.MkdirAll(filepath.Join(empty, "a", "b"), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if got := firstRegularFile(empty); got != "" {
		t.Errorf("firstRegularFile(dirs only) = %q, want \"\"", got)
	}
}

func TestReachErrDistinguishesReachableFromMissing(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, []byte("content"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	empty := filepath.Join(dir, "empty")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := reachErr(dir); err != nil {
		t.Errorf("reachErr(dir) = %v, want nil", err)
	}
	if err := reachErr(file); err != nil {
		t.Errorf("reachErr(file) = %v, want nil", err)
	}
	// An empty file is reachable: EOF on the first byte is not a refusal.
	if err := reachErr(empty); err != nil {
		t.Errorf("reachErr(empty file) = %v, want nil", err)
	}
	if err := reachErr(filepath.Join(dir, "missing")); !errors.Is(err, syscall.ENOENT) {
		t.Errorf("reachErr(missing) = %v, want ENOENT", err)
	}
	// An unset HOME reaches the empty path, which must not look like success.
	if err := reachErr(""); !errors.Is(err, syscall.ENOENT) {
		t.Errorf("reachErr(\"\") = %v, want ENOENT", err)
	}
}

func TestReadDirErr(t *testing.T) {
	if err := readDirErr(t.TempDir()); err != nil {
		t.Errorf("readDirErr(dir) = %v, want nil", err)
	}
	if err := readDirErr(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("readDirErr(missing) = nil, want an error")
	}
}

func TestWaitForFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "marker")

	if waitForFile(path, 50*time.Millisecond) {
		t.Error("waitForFile reported a file that was never written")
	}

	go func() {
		time.Sleep(30 * time.Millisecond)
		os.WriteFile(path, []byte("x"), 0o600) //nolint:errcheck // best effort in a test goroutine
	}()
	if !waitForFile(path, 3*time.Second) {
		t.Error("waitForFile timed out on a file that was written")
	}
}

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	if err := os.WriteFile(src, []byte("payload"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(data) != "payload" {
		t.Errorf("copy = %q, want %q", data, "payload")
	}
	// The copy is written executable on purpose: the exec probe has to be able
	// to try to run it, otherwise "cannot execute what it wrote" is vacuous.
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode()&0o100 == 0 {
		t.Errorf("copy mode %v is not owner-executable", info.Mode())
	}

	if err := copyFile(filepath.Join(dir, "missing"), dst); err == nil {
		t.Error("copyFile from a missing source returned no error")
	}
}

func TestMmapExecReportsAMissingFile(t *testing.T) {
	if err := mmapExec(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("mmapExec of a missing file returned no error")
	}
}

// The executable-mapping escape has no negative control on Apple Silicon: the
// platform refuses PROT_EXEC on every file, a signed system binary included,
// unless the process carries a JIT entitlement. This test records which regime
// the host is in, because the class probe attributes the denial differently in
// each. If it ever starts succeeding uncontained, the exec-allowlist verdict
// has to be re-derived rather than inherited.
func TestMmapExecPlatformRegime(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "page")
	if err := os.WriteFile(path, make([]byte, syscall.Getpagesize()), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	ownFile := mmapExec(path)
	signedBinary := mmapExec("/bin/echo")

	switch {
	case ownFile == nil:
		t.Logf("this host permits PROT_EXEC on an unsigned file: the mmap escape has a real negative control here")
	case errors.Is(ownFile, syscall.EPERM):
		// The platform, not the profile, is refusing. Confirm it by mapping a
		// file that is beyond any doubt executable: if that is refused too, the
		// refusal cannot be about the file under test.
		if signedBinary == nil {
			t.Errorf("PROT_EXEC refused an owned file (%v) but allowed a signed binary; "+
				"the refusal is about the file, not a blanket W^X policy", ownFile)
		}
		t.Logf("this host refuses PROT_EXEC for every file (own file: %v, /bin/echo: %v); "+
			"the mmap escape is not attributable to any sandbox profile", ownFile, signedBinary)
	default:
		t.Errorf("mmapExec failed for an unexpected reason: %v", ownFile)
	}
}

func TestWritePastOvershootsTheBudget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bytes.bin")
	budget := int64(64 << 10)

	written, err := writePast(path, budget)
	if err != nil {
		t.Fatalf("writePast: %v", err)
	}
	// Nothing on this host bounds the bytes, so the probe must run past the
	// declared budget: that overshoot is the observation.
	if written < budget*2 {
		t.Errorf("wrote %d bytes, want at least %d", written, budget*2)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("writePast left %s behind: %v", path, err)
	}
}

func TestWritePastReportsAnUncreatableTarget(t *testing.T) {
	_, err := writePast(filepath.Join(t.TempDir(), "missing-dir", "bytes.bin"), 1024)
	if err == nil {
		t.Fatal("writePast into a missing directory returned no error")
	}
}

func TestRunProgram(t *testing.T) {
	if err := runProgram("/bin/echo", "hello"); err != nil {
		t.Errorf("runProgram(/bin/echo) = %v, want nil", err)
	}
	// A program that exits nonzero still started: runProgram reports launch
	// failures, not exit status, because only the launch is under test.
	if err := runProgram("/usr/bin/false"); err != nil {
		t.Errorf("runProgram(/usr/bin/false) = %v, want nil", err)
	}
	err := runProgram(filepath.Join(t.TempDir(), "not-a-program"))
	if err == nil {
		t.Fatal("runProgram of a missing path returned no error")
	}
	if outcome, errno := classify(err); outcome != OutcomeInconclusive || errno != "ENOENT" {
		t.Errorf("a missing program classified as (%q, %q), want inconclusive/ENOENT", outcome, errno)
	}
}

// A program that never exits must not hang the probe: the agent has to report
// something for every attempt.
func TestRunProgramGivesUpOnAHangingProgram(t *testing.T) {
	if testing.Short() {
		t.Skip("the give-up path takes the full five-second timeout")
	}
	start := time.Now()
	if err := runProgram("/bin/sleep", "120"); err != nil {
		t.Errorf("runProgram(sleep) = %v, want nil after the timeout", err)
	}
	elapsed := time.Since(start)
	if elapsed > 30*time.Second {
		t.Errorf("runProgram took %v; it should give up after about five seconds", elapsed)
	}
}

// os/exec opens /dev/null for a nil stream, and a deny-default profile refuses
// that open. The resulting EPERM is indistinguishable from an exec denial, so
// every stream must be wired explicitly.
func TestPipeStdioLeavesNoNilStream(t *testing.T) {
	cmd := exec.Command("/bin/echo")
	pipeStdio(cmd)
	if cmd.Stdin == nil || cmd.Stdout == nil || cmd.Stderr == nil {
		t.Errorf("pipeStdio left a nil stream: in=%v out=%v err=%v", cmd.Stdin, cmd.Stdout, cmd.Stderr)
	}
	if cmd.Stdout != io.Discard || cmd.Stderr != io.Discard {
		t.Error("pipeStdio did not discard the output streams")
	}
}

// ---------------------------------------------------------------- setxattr

func TestSetxattr(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Uncontained, setting an extended attribute on an owned file succeeds.
	// Section 5.2 lists it among the mutations a read-only view must refuse, so
	// a probe that could never succeed would be a false denial.
	if err := setxattr(path, "user.probe", []byte("value")); err != nil {
		t.Errorf("setxattr uncontained = %v, want nil", err)
	}
	// An empty value is padded rather than dereferencing an empty slice.
	if err := setxattr(path, "user.empty", nil); err != nil {
		t.Errorf("setxattr with an empty value = %v, want nil", err)
	}
	err := setxattr(filepath.Join(dir, "missing"), "user.probe", []byte("v"))
	if !errors.Is(err, syscall.ENOENT) {
		t.Errorf("setxattr on a missing file = %v, want ENOENT", err)
	}
	// A NUL in the path cannot be turned into a C string; it must be an error,
	// never a silently truncated path.
	if err := setxattr("/tmp/na\x00me", "user.probe", []byte("v")); err == nil {
		t.Error("setxattr with a NUL in the path returned no error")
	}
	if err := setxattr(path, "user.na\x00me", []byte("v")); err == nil {
		t.Error("setxattr with a NUL in the name returned no error")
	}
}

// ------------------------------------------------------- attempt generators

// Uncontained, every mutation of a writable directory must succeed. If any of
// them failed here, the corresponding in-domain denial would prove nothing.
func TestMutationsSucceedOnAWritableDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "target"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	attempts := mutations(dir)
	byName := map[string]Attempt{}
	for _, a := range attempts {
		byName[a.Name] = a
	}
	for _, name := range []string{
		"open-write-existing", "open-truncate-existing", "chmod-existing", "chown-existing",
		"setxattr-existing", "rename-existing", "unlink-existing", "hardlink-existing",
		"create-file", "mkdir", "symlink",
	} {
		a, ok := byName[name]
		if !ok {
			t.Errorf("mutations did not attempt %q", name)
			continue
		}
		if a.Outcome != OutcomeAllowed {
			t.Errorf("%s on a writable directory = %q (%s: %s), want allowed",
				name, a.Outcome, a.Errno, a.Detail)
		}
	}
}

func TestMutationsAreDeniedOnAReadOnlyDirectory(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root ignores directory permissions, so nothing would be refused")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "target"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o700) }) //nolint:errcheck // restoring so TempDir cleanup can run

	byName := map[string]Attempt{}
	for _, a := range mutations(dir) {
		byName[a.Name] = a
	}
	// These need write permission on the directory itself, so a plain POSIX mode
	// is enough to refuse them. The rest need the sandbox and are covered by the
	// in-domain probes.
	for _, name := range []string{"create-file", "mkdir", "symlink", "rename-existing", "unlink-existing", "hardlink-existing"} {
		a, ok := byName[name]
		if !ok {
			t.Errorf("mutations did not attempt %q", name)
			continue
		}
		if a.Outcome != OutcomeDenied {
			t.Errorf("%s on a read-only directory = %q (%s), want denied", name, a.Outcome, a.Errno)
		}
	}
}

// A view with no regular file cannot be mutated meaningfully; the probe must
// say so rather than report a denial it never observed.
func TestMutationsReportAnEmptyViewAsInconclusive(t *testing.T) {
	attempts := mutations(t.TempDir())
	var found bool
	for _, a := range attempts {
		if a.Name != "open-write-existing" {
			continue
		}
		found = true
		if a.Outcome != OutcomeInconclusive {
			t.Errorf("open-write-existing on an empty view = %q, want inconclusive", a.Outcome)
		}
		if !strings.Contains(a.Detail, "no readable regular file") {
			t.Errorf("detail %q does not explain the empty view", a.Detail)
		}
	}
	if !found {
		t.Error("mutations did not report open-write-existing for an empty view")
	}
}

func TestReadOnlyViewAttemptsLabelsEachView(t *testing.T) {
	source, goroot := t.TempDir(), t.TempDir()
	for _, dir := range []string{source, goroot} {
		if err := os.WriteFile(filepath.Join(dir, "target"), []byte("x"), 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	t.Setenv(EnvSource, source)
	t.Setenv(EnvGoroot, goroot)

	byName := map[string]Attempt{}
	for _, a := range readOnlyViewAttempts() {
		byName[a.Name] = a
	}
	for _, name := range []string{
		"source:read-dir", "source:create-file", "source:chmod-existing",
		"toolchain:read-dir", "toolchain:create-file", "toolchain:chmod-existing",
	} {
		if _, ok := byName[name]; !ok {
			t.Errorf("readOnlyViewAttempts did not report %q", name)
		}
	}
	if a := byName["source:read-dir"]; a.Outcome != OutcomeAllowed {
		t.Errorf("source:read-dir = %q, want allowed uncontained", a.Outcome)
	}
}

// An unset view variable must produce no attempts for that view rather than
// attempts against the empty path, which would be reported as ENOENT denials of
// a view that was never presented.
func TestReadOnlyViewAttemptsSkipsUnsetViews(t *testing.T) {
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "target"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Setenv(EnvSource, source)
	unsetEnv(t, EnvGoroot)

	for _, a := range readOnlyViewAttempts() {
		if strings.HasPrefix(a.Name, "toolchain:") {
			t.Errorf("an unset toolchain view still produced %q", a.Name)
		}
	}
}

func TestViewRestrictionAttemptsCoversTheUndeclaredPaths(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	byName := map[string]Attempt{}
	for _, a := range viewRestrictionAttempts() {
		byName[a.Name] = a
	}
	for _, name := range []string{
		"read-etc-passwd", "read-etc-hosts", "readdir-users", "readdir-home",
		"readdir-applications", "read-ssh-dir", "stat-home", "readdir-root",
	} {
		if _, ok := byName[name]; !ok {
			t.Errorf("viewRestrictionAttempts did not report %q", name)
		}
	}
	// Uncontained, the root namespace is enumerable. The class probe scores this
	// as a failure on macOS precisely because it stays enumerable in-domain.
	if a := byName["readdir-root"]; a.Outcome != OutcomeAllowed {
		t.Errorf("readdir-root uncontained = %q (%s), want allowed", a.Outcome, a.Errno)
	}
	if a := byName["read-etc-hosts"]; a.Outcome != OutcomeAllowed {
		t.Errorf("read-etc-hosts uncontained = %q (%s), want allowed", a.Outcome, a.Errno)
	}
}

func TestWriteConfinementAttemptsAimsAtEveryEscape(t *testing.T) {
	root := t.TempDir()
	buildRoot := filepath.Join(root, "buildroot")
	outside := filepath.Join(root, "outside")
	for _, dir := range []string{buildRoot, outside} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(outside, "outside-target"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	home := t.TempDir()
	tmpdir := t.TempDir()
	t.Setenv(EnvBuildRoot, buildRoot)
	t.Setenv(EnvOutside, outside)
	t.Setenv("HOME", home)
	t.Setenv("TMPDIR", tmpdir)

	// The probe writes to two fixed system paths by design. Remove only what
	// this test created, so a pre-existing file is never destroyed.
	for _, fixed := range []string{"/tmp/probe-escape", "/private/tmp/probe-escape"} {
		if _, err := os.Lstat(fixed); errors.Is(err, os.ErrNotExist) {
			path := fixed
			t.Cleanup(func() { _ = os.Remove(path) })
		}
	}

	byName := map[string]Attempt{}
	for _, a := range writeConfinementAttempts() {
		byName[a.Name] = a
	}
	for _, name := range []string{
		"write-inside-build-root", "write-absolute-outside", "write-home", "write-tmp",
		"write-private-tmp", "write-build-root-parent", "write-relative-traversal",
		"write-tmpdir-env", "create-escape-symlink", "write-through-symlink",
		"create-escape-hardlink", "write-through-hardlink",
	} {
		a, ok := byName[name]
		if !ok {
			t.Fatalf("writeConfinementAttempts did not report %q", name)
		}
		// Uncontained every one of them succeeds; that is what makes the
		// in-domain denials evidence rather than an artifact.
		if a.Outcome != OutcomeAllowed {
			t.Errorf("%s uncontained = %q (%s: %s), want allowed", name, a.Outcome, a.Errno, a.Detail)
		}
	}
	// The escape targets really were reached, which is the point of running the
	// same attempts uncontained.
	if _, err := os.Stat(filepath.Join(outside, "escape-absolute")); err != nil {
		t.Errorf("the uncontained escape did not land: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".probe-escape")); err != nil {
		t.Errorf("the uncontained home escape did not land: %v", err)
	}
}

// With no TMPDIR there is no inherited temp directory to escape into, so the
// attempt must be absent rather than aimed at a bare filename.
func TestWriteConfinementAttemptsSkipsAnUnsetTmpdir(t *testing.T) {
	root := t.TempDir()
	buildRoot := filepath.Join(root, "buildroot")
	outside := filepath.Join(root, "outside")
	for _, dir := range []string{buildRoot, outside} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	t.Setenv(EnvBuildRoot, buildRoot)
	t.Setenv(EnvOutside, outside)
	t.Setenv("HOME", t.TempDir())
	unsetEnv(t, "TMPDIR")

	for _, fixed := range []string{"/tmp/probe-escape", "/private/tmp/probe-escape"} {
		if _, err := os.Lstat(fixed); errors.Is(err, os.ErrNotExist) {
			path := fixed
			t.Cleanup(func() { os.Remove(path) }) //nolint:errcheck // best-effort cleanup
		}
	}

	for _, a := range writeConfinementAttempts() {
		if a.Name == "write-tmpdir-env" {
			t.Errorf("an unset TMPDIR still produced %q against %q", a.Name, a.Target)
		}
	}
}

func TestExecAllowlistAttemptsCoversEveryRoute(t *testing.T) {
	buildRoot := t.TempDir()
	t.Setenv(EnvBuildRoot, buildRoot)
	// /bin/echo stands in for the allowlisted agent: it is small, harmless and
	// exits immediately, and the probe only cares whether exec was refused.
	t.Setenv(EnvSelf, "/bin/echo")

	byName := map[string]Attempt{}
	for _, a := range execAllowlistAttempts() {
		byName[a.Name] = a
	}
	for _, name := range []string{
		"exec-shell", "exec-bash", "exec-zsh", "exec-interpreter", "exec-host-binary",
		"exec-perl", "exec-dyld-as-program", "write-program-into-build-root",
		"exec-self-written-copy", "mmap-exec-self-written", "exec-symlink-to-allowlisted",
		"exec-allowlisted-self",
	} {
		if _, ok := byName[name]; !ok {
			t.Errorf("execAllowlistAttempts did not report %q", name)
		}
	}
	// Uncontained, the domain can write a program into its build root and start
	// it. Both must be true or the in-domain denial is vacuous.
	if a := byName["write-program-into-build-root"]; a.Outcome != OutcomeAllowed {
		t.Errorf("write-program-into-build-root = %q (%s), want allowed", a.Outcome, a.Detail)
	}
	if a := byName["exec-self-written-copy"]; a.Outcome != OutcomeAllowed {
		t.Errorf("exec-self-written-copy uncontained = %q (%s), want allowed", a.Outcome, a.Detail)
	}
	if a := byName["exec-allowlisted-self"]; a.Outcome != OutcomeAllowed {
		t.Errorf("exec-allowlisted-self uncontained = %q (%s), want allowed", a.Outcome, a.Detail)
	}
}

// Without a writable build root the exec probe cannot stage its program. It has
// to report that as inconclusive rather than let a missing file read as a
// denial of execution.
func TestExecAllowlistAttemptsReportsAnUnwritableBuildRoot(t *testing.T) {
	t.Setenv(EnvBuildRoot, filepath.Join(t.TempDir(), "missing-dir"))
	t.Setenv(EnvSelf, "/bin/echo")

	byName := map[string]Attempt{}
	for _, a := range execAllowlistAttempts() {
		byName[a.Name] = a
	}
	for _, name := range []string{"exec-self-written-copy", "mmap-exec-self-written"} {
		a, ok := byName[name]
		if !ok {
			t.Fatalf("execAllowlistAttempts did not report %q", name)
		}
		if a.Outcome != OutcomeInconclusive {
			t.Errorf("%s with no build root = %q, want inconclusive", name, a.Outcome)
		}
	}
}

// With no allowlisted program to point at, the symlink and self-exec attempts
// have no subject and must be omitted rather than aimed at "".
func TestExecAllowlistAttemptsSkipsAnUnsetSelf(t *testing.T) {
	t.Setenv(EnvBuildRoot, t.TempDir())
	unsetEnv(t, EnvSelf)

	for _, a := range execAllowlistAttempts() {
		if a.Name == "exec-symlink-to-allowlisted" || a.Name == "exec-allowlisted-self" {
			t.Errorf("an unset %s still produced %q", EnvSelf, a.Name)
		}
	}
}

func TestNetworkAttemptsCoversEveryFamilyAndDirection(t *testing.T) {
	if testing.Short() {
		t.Skip("the off-host dials wait for their timeout")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	sock := shortSocketPath(t)
	unixListener, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen unix: %v", err)
	}
	defer func() { _ = unixListener.Close() }()
	go func() {
		for {
			conn, err := unixListener.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	t.Setenv(EnvLoopback, listener.Addr().String())
	t.Setenv(EnvUnixSocket, sock)

	byName := map[string]Attempt{}
	for _, a := range networkAttempts() {
		byName[a.Name] = a
	}
	for _, name := range []string{
		"socket-inet-stream", "socket-inet6-stream", "socket-inet-dgram", "socket-unix-stream",
		"connect-loopback-tcp", "connect-offhost-tcp", "connect-offhost-udp", "connect-unix",
		"listen-loopback-tcp", "socketpair-unix",
	} {
		if _, ok := byName[name]; !ok {
			t.Errorf("networkAttempts did not report %q", name)
		}
	}
	// The negative-control shape: uncontained, loopback and unix connects and an
	// inbound listen all succeed.
	for _, name := range []string{"connect-loopback-tcp", "connect-unix", "listen-loopback-tcp", "socketpair-unix"} {
		if a := byName[name]; a.Outcome != OutcomeAllowed {
			t.Errorf("%s uncontained = %q (%s: %s), want allowed", name, a.Outcome, a.Errno, a.Detail)
		}
	}
	// Never denied uncontained: a denial here would mean the host, not the
	// profile, refused, and every in-domain denial would be unattributable.
	for name, a := range byName {
		if a.Outcome == OutcomeDenied {
			t.Errorf("%s was denied with no profile applied (%s: %s)", name, a.Errno, a.Detail)
		}
	}
}

// A loopback endpoint the harness never set must not be dialled: an attempt
// against "" would be reported under a name the class probe looks for.
func TestNetworkAttemptsSkipsUnsetEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("the off-host dials wait for their timeout")
	}
	unsetEnv(t, EnvLoopback)
	unsetEnv(t, EnvUnixSocket)

	for _, a := range networkAttempts() {
		if a.Name == "connect-loopback-tcp" || a.Name == "connect-unix" {
			t.Errorf("an unset endpoint still produced %q against %q", a.Name, a.Target)
		}
	}
}

func TestDialAttemptReportsARefusedConnection(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := listener.Addr().String()
	_ = listener.Close()

	// Nothing is listening: the connection is refused, which is not a denial.
	a := dialAttempt("connect-closed-port", "tcp", addr)
	if a.Outcome == OutcomeDenied {
		t.Errorf("a refused connection was reported as a denial: %+v", a)
	}
}

// ------------------------------------------------------------------- Main

// captureStdout swaps os.Stdout for a pipe so Main's report can be read back
// in-process. The agent writes to the file descriptor directly, so redirecting
// the package-level variable is the only way to see it without a subprocess.
func captureStdout(t *testing.T, fn func() int) (int, string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	original := os.Stdout
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		data, _ := io.ReadAll(r)
		done <- string(data)
	}()

	code := fn()

	os.Stdout = original
	_ = w.Close()
	out := <-done
	_ = r.Close()
	return code, out
}

func TestMainRejectsUnusableInvocations(t *testing.T) {
	cases := map[string][]string{
		"no op":      {},
		"unknown op": {"take-over-the-host"},
		"empty op":   {""},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			code, out := captureStdout(t, func() int { return Main(args) })
			if code != 2 {
				t.Errorf("Main(%v) = %d, want 2", args, code)
			}
			if strings.TrimSpace(out) != "" {
				t.Errorf("Main(%v) wrote a report on stdout: %q", args, out)
			}
		})
	}
}

func TestMainHelloReportsIdentity(t *testing.T) {
	unsetEnv(t, "PROBE_COUNT_DESCRIPTORS")

	code, out := captureStdout(t, func() int { return Main([]string{OpHello}) })
	if code != 0 {
		t.Fatalf("Main(hello) = %d, want 0", code)
	}
	report, err := ParseReport(out)
	if err != nil {
		t.Fatalf("ParseReport: %v (stdout %q)", err, out)
	}
	if report.Op != OpHello || report.Values["hello"] != "ok" {
		t.Errorf("report = %+v, want a hello acknowledgement", report)
	}
	if report.PID != os.Getpid() {
		t.Errorf("report pid %d, want %d", report.PID, os.Getpid())
	}
	if report.PGID <= 0 || report.SID <= 0 {
		t.Errorf("report pgid=%d sid=%d, want both positive", report.PGID, report.SID)
	}
	// Without the descriptor-counting request, hello must not do the expensive
	// count: the class probe relies on hello being the cheap smoke test.
	if _, ok := report.Values["descriptors"]; ok {
		t.Errorf("hello counted descriptors without being asked: %v", report.Values)
	}
}

func TestMainEmitsOneJSONLinePerOp(t *testing.T) {
	source, goroot := t.TempDir(), t.TempDir()
	for _, dir := range []string{source, goroot} {
		if err := os.WriteFile(filepath.Join(dir, "target"), []byte("x"), 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	t.Setenv(EnvSource, source)
	t.Setenv(EnvGoroot, goroot)
	t.Setenv(EnvBuildRoot, t.TempDir())
	t.Setenv(EnvOutside, t.TempDir())
	t.Setenv(EnvSelf, "/bin/echo")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("TMPDIR", t.TempDir())
	unsetEnv(t, EnvLoopback)
	unsetEnv(t, EnvUnixSocket)

	for _, fixed := range []string{"/tmp/probe-escape", "/private/tmp/probe-escape"} {
		if _, err := os.Lstat(fixed); errors.Is(err, os.ErrNotExist) {
			path := fixed
			t.Cleanup(func() { os.Remove(path) }) //nolint:errcheck // best-effort cleanup
		}
	}

	ops := []string{OpReadOnlyView, OpWriteConfinement, OpViewRestriction, OpExecAllowlist}
	for _, op := range ops {
		t.Run(op, func(t *testing.T) {
			code, out := captureStdout(t, func() int { return Main([]string{op}) })
			if code != 0 {
				t.Fatalf("Main(%s) = %d, want 0", op, code)
			}
			if lines := strings.Count(strings.TrimSpace(out), "\n"); lines != 0 {
				t.Errorf("Main(%s) wrote %d extra lines; the report must be one line", op, lines+1)
			}
			if !json.Valid([]byte(strings.TrimSpace(out))) {
				t.Errorf("Main(%s) did not emit valid JSON: %q", op, out)
			}
			report, err := ParseReport(out)
			if err != nil {
				t.Fatalf("ParseReport: %v", err)
			}
			if report.Op != op {
				t.Errorf("report op %q, want %q", report.Op, op)
			}
			if len(report.Attempts) == 0 {
				t.Errorf("Main(%s) reported no attempts", op)
			}
			// Every attempt must carry a name and one of the three outcomes:
			// the class probes look attempts up by name and score by outcome.
			for _, a := range report.Attempts {
				if a.Name == "" {
					t.Errorf("an attempt has no name: %+v", a)
				}
				switch a.Outcome {
				case OutcomeAllowed, OutcomeDenied, OutcomeInconclusive:
				default:
					t.Errorf("attempt %q has outcome %q, which is not one of the three", a.Name, a.Outcome)
				}
			}
		})
	}
}

// The descendant op spawns two real children of the allowlisted program and
// reports their identities. It is the input to both domain classes, so the
// values it publishes are part of the contract.
func TestMainDescendantReportsBothChildren(t *testing.T) {
	if testing.Short() {
		t.Skip("the descendant op spawns children and waits for their markers")
	}
	probeBin := buildProbeBinary(t)
	markerDir := t.TempDir()
	markerBase := filepath.Join(markerDir, "descendant")

	t.Setenv(EnvSelf, probeBin)
	t.Setenv(EnvMarker, markerBase)
	t.Setenv(EnvHold, "1")

	code, out := captureStdout(t, func() int { return Main([]string{OpDescendant}) })
	if code != 0 {
		t.Fatalf("Main(descendant) = %d, want 0", code)
	}
	report, err := ParseReport(out)
	if err != nil {
		t.Fatalf("ParseReport: %v (stdout %q)", err, out)
	}

	for _, name := range []string{"spawn-detached-descendant", "spawn-attached-descendant"} {
		a, ok := report.Find(name)
		if !ok {
			t.Fatalf("descendant report has no %q", name)
		}
		if a.Outcome != OutcomeAllowed {
			t.Errorf("%s = %q (%s), want allowed", name, a.Outcome, a.Detail)
		}
	}

	detached := mustAtoi(t, report.Values["detached_pid"])
	attached := mustAtoi(t, report.Values["attached_pid"])
	if detached <= 0 || attached <= 0 {
		t.Fatalf("child pids detached=%d attached=%d, want both positive", detached, attached)
	}
	t.Cleanup(func() {
		syscall.Kill(detached, syscall.SIGKILL) //nolint:errcheck // best-effort cleanup
		syscall.Kill(attached, syscall.SIGKILL) //nolint:errcheck // best-effort cleanup
	})

	if mustAtoi(t, report.Values["domain_sid"]) != getsid() {
		t.Errorf("domain_sid %q does not match the running session %d", report.Values["domain_sid"], getsid())
	}

	// The markers are what the harness reads to learn whether the detached child
	// left the session. Both must exist by the time the op returns.
	detachedMarker := readMarker(t, markerBase+".detached")
	attachedMarker := readMarker(t, markerBase+".attached")

	// setsid on the detached child is the whole escape being measured: its
	// session must differ from ours, and the attached child's must not.
	if detachedMarker.sid == getsid() {
		t.Errorf("the detached child stayed in session %d; setsid did not take effect", detachedMarker.sid)
	}
	if attachedMarker.sid != getsid() {
		t.Errorf("the attached child is in session %d, want the domain session %d", attachedMarker.sid, getsid())
	}
	if detachedMarker.pid != detached {
		t.Errorf("detached marker pid %d, want the reported %d", detachedMarker.pid, detached)
	}
}

type marker struct{ pid, pgid, sid, ppid int }

func readMarker(t *testing.T, path string) marker {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read marker %s: %v", path, err)
	}
	var m marker
	if _, err := fmt.Sscanf(string(data), "pid=%d pgid=%d sid=%d ppid=%d", &m.pid, &m.pgid, &m.sid, &m.ppid); err != nil {
		t.Fatalf("marker %s is malformed (%q): %v", path, data, err)
	}
	return m
}

func mustAtoi(t *testing.T, raw string) int {
	t.Helper()
	n, err := strconv.Atoi(raw)
	if err != nil {
		t.Fatalf("value %q is not a number: %v", raw, err)
	}
	return n
}

func TestStartChildReportsAnUnstartableProgram(t *testing.T) {
	pid, err := startChild(filepath.Join(t.TempDir(), "not-a-program"), filepath.Join(t.TempDir(), "marker"), 1, false)
	if err == nil {
		t.Fatalf("startChild of a missing program returned pid %d and no error", pid)
	}
	if pid != 0 {
		t.Errorf("startChild returned pid %d on failure, want 0", pid)
	}
}

// Without the environment the harness sets, the descendant op has nothing to
// spawn and must say so instead of reporting a spawn that never happened.
func TestDescendantAttemptsWithoutEnvironment(t *testing.T) {
	unsetEnv(t, EnvSelf)
	unsetEnv(t, EnvMarker)

	report := Report{Values: map[string]string{}}
	attempts := descendantAttempts(&report)
	if len(attempts) != 1 {
		t.Fatalf("descendantAttempts returned %d attempts, want 1", len(attempts))
	}
	if attempts[0].Outcome != OutcomeInconclusive {
		t.Errorf("outcome %q, want inconclusive", attempts[0].Outcome)
	}
	if len(report.Values) != 0 {
		t.Errorf("values %v were published without any child", report.Values)
	}
}

func TestDetachedChildWithoutAMarker(t *testing.T) {
	unsetEnv(t, EnvMarker)
	if code := detachedChild(); code != 2 {
		t.Errorf("detachedChild with no marker = %d, want 2", code)
	}
}

func TestDetachedChildWritesItsIdentityAndHolds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "marker")
	t.Setenv(EnvMarker, path)
	t.Setenv(EnvHold, "1")

	start := time.Now()
	if code := detachedChild(); code != 0 {
		t.Fatalf("detachedChild = %d, want 0", code)
	}
	if elapsed := time.Since(start); elapsed < 900*time.Millisecond {
		t.Errorf("detachedChild returned after %v; it must hold for the requested second", elapsed)
	}

	m := readMarker(t, path)
	if m.pid != os.Getpid() || m.sid != getsid() {
		t.Errorf("marker %+v does not describe this process (pid %d sid %d)", m, os.Getpid(), getsid())
	}
	// The survived file is only written when the hold elapsed without a kill,
	// which is what an escape looks like after the fact.
	if _, err := os.Stat(path + ".survived"); err != nil {
		t.Errorf("detachedChild did not record that the hold elapsed: %v", err)
	}
}

func TestDetachedChildReportsAnUnwritableMarker(t *testing.T) {
	t.Setenv(EnvMarker, filepath.Join(t.TempDir(), "missing-dir", "marker"))
	t.Setenv(EnvHold, "1")
	if code := detachedChild(); code != 2 {
		t.Errorf("detachedChild with an unwritable marker = %d, want 2", code)
	}
}

func TestCountForHelloIgnoresUnusableRequests(t *testing.T) {
	for _, raw := range []string{"", "junk", "0", "8", "-1"} {
		t.Setenv("PROBE_COUNT_DESCRIPTORS", raw)
		if count, ok := countForHello(); ok {
			t.Errorf("countForHello with %q = (%d, true), want no count", raw, count)
		}
	}
	unsetEnv(t, "PROBE_COUNT_DESCRIPTORS")
	if _, ok := countForHello(); ok {
		t.Error("countForHello with the variable unset reported a count")
	}
}

func TestGetpgidAndGetsid(t *testing.T) {
	if pgid := getpgid(); pgid <= 0 {
		t.Errorf("getpgid() = %d, want a positive process group", pgid)
	}
	if sid := getsid(); sid <= 0 {
		t.Errorf("getsid() = %d, want a positive session", sid)
	}
}

// unsetEnv removes a variable for the duration of one test and puts it back
// afterwards. os.Unsetenv alone would leak the removal into every later test in
// the same binary, which is how one probe's environment silently becomes
// another's.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	previous, had := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
	t.Cleanup(func() {
		if had {
			_ = os.Setenv(key, previous)
			return
		}
		_ = os.Unsetenv(key)
	})
}

// shortSocketPath returns a unix-socket path that fits in sockaddr_un.sun_path,
// which holds 104 bytes on macOS including the terminator.
//
// t.TempDir embeds the test's name, and a descriptive name plus the per-run
// suffix is enough to push a socket path past the limit. The bind then fails
// with EINVAL, which reads as a broken test rather than as the path-length
// problem it is, so the socket gets its own short directory instead.
func shortSocketPath(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "ins")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	path := filepath.Join(dir, "p.sock")
	if len(path) > 103 {
		t.Skipf("no usable socket path: %q is %d bytes", path, len(path))
	}
	return path
}
