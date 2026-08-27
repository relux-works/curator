package seatbelt

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func requireDarwinSandbox(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "darwin" {
		t.Skipf("seatbelt is a macOS mechanism; this host is %s", runtime.GOOS)
	}
	if ok, why := Available(); !ok {
		t.Skipf("no dynamic sandbox launcher on this host: %s", why)
	}
}

// ------------------------------------------------------------------ Render

// A profile that rendered differently between the run and the report would make
// every observation unattributable to a known rule set, so byte stability is a
// property of the artifact, not a nicety.
func TestRenderIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	profile := Profile{
		ReadOnlyPaths:      []string{filepath.Join(dir, "source"), filepath.Join(dir, "goroot")},
		WritablePaths:      []string{filepath.Join(dir, "build")},
		ExecPaths:          []string{filepath.Join(dir, "agent")},
		MapExecutablePaths: []string{filepath.Join(dir, "agent")},
		AllowSysctlRead:    true,
		AllowProcessInfo:   true,
		AllowSignals:       true,
		AllowForkExec:      true,
	}
	first := profile.Render()
	for i := 0; i < 16; i++ {
		if got := profile.Render(); got != first {
			t.Fatalf("render %d differs from the first:\n--- first ---\n%s\n--- got ---\n%s", i, first, got)
		}
	}
}

func TestRenderAlwaysDeniesByDefault(t *testing.T) {
	for name, profile := range map[string]Profile{
		"empty":       {},
		"permissive":  {AllowNetwork: true, AllowMachLookup: true, AllowSignals: true},
		"with a path": {WritablePaths: []string{t.TempDir()}},
	} {
		text := profile.Render()
		if !strings.HasPrefix(text, "(version 1)\n(deny default)\n") {
			t.Errorf("%s profile does not open with a default denial:\n%s", name, text)
		}
	}
}

// The loader paths are not optional: a profile that omits read access to them
// aborts every dynamically linked program during startup, which looks exactly
// like perfect enforcement.
func TestRenderAlwaysGrantsTheLoaderPaths(t *testing.T) {
	text := Profile{}.Render()
	if !strings.Contains(text, `(allow file-read* (literal "/")`) {
		t.Errorf("profile does not grant read on the root literal:\n%s", text)
	}
	for _, path := range loaderReadPaths {
		if !strings.Contains(text, `"`+path+`"`) {
			t.Errorf("profile does not grant read on loader path %q:\n%s", path, text)
		}
	}
	// The root is granted as a literal, never a subpath: a subpath would make
	// every top-level entry readable through it.
	if strings.Contains(text, `(subpath "/")`) {
		t.Errorf("profile grants the root as a subpath, which opens the whole namespace:\n%s", text)
	}
}

func TestRenderOmitsRulesForEmptyAllowances(t *testing.T) {
	text := Profile{}.Render()
	for _, rule := range []string{"file-write*", "process-exec*", "process-fork", "network*", "sysctl-read", "process-info*", "(allow signal)", "mach-lookup"} {
		if strings.Contains(text, rule) {
			t.Errorf("empty profile still emits %q:\n%s", rule, text)
		}
	}
	// With no bounded list, executable mapping is granted wherever reading is.
	// That is a real hole, so the unbounded form must be visible in the profile.
	if !strings.Contains(text, "(allow file-map-executable)\n") {
		t.Errorf("empty profile does not emit the unbounded file-map-executable:\n%s", text)
	}
}

func TestRenderEmitsEachOptionalRule(t *testing.T) {
	cases := []struct {
		name    string
		profile Profile
		want    string
	}{
		{"network", Profile{AllowNetwork: true}, "(allow network*)"},
		{"sysctl", Profile{AllowSysctlRead: true}, "(allow sysctl-read)"},
		{"process info", Profile{AllowProcessInfo: true}, "(allow process-info*)"},
		{"signals", Profile{AllowSignals: true}, "(allow signal)"},
		{"fork", Profile{AllowForkExec: true}, "(allow process-fork)"},
		{"mach lookup", Profile{AllowMachLookup: true}, "(allow mach-lookup)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if text := tc.profile.Render(); !strings.Contains(text, tc.want) {
				t.Errorf("profile does not emit %q:\n%s", tc.want, text)
			}
		})
	}
}

// A writable path must also be readable; a build root the domain can write but
// not read is not a build root.
func TestRenderMakesWritablePathsReadable(t *testing.T) {
	build := t.TempDir()
	text := Profile{WritablePaths: []string{build}}.Render()

	readLine, writeLine := "", ""
	for _, line := range strings.Split(text, "\n") {
		switch {
		case strings.HasPrefix(line, "(allow file-read*"):
			readLine = line
		case strings.HasPrefix(line, "(allow file-write*"):
			writeLine = line
		}
	}
	if !strings.Contains(readLine, build) {
		t.Errorf("writable path %q is not readable:\n%s", build, text)
	}
	if !strings.Contains(writeLine, build) {
		t.Errorf("writable path %q is not writable:\n%s", build, text)
	}
}

// Read-only means read-only: a path listed only under ReadOnlyPaths must never
// reach the write allowance.
func TestRenderNeverMakesReadOnlyPathsWritable(t *testing.T) {
	source := t.TempDir()
	build := t.TempDir()
	text := Profile{ReadOnlyPaths: []string{source}, WritablePaths: []string{build}}.Render()

	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "(allow file-write*") && strings.Contains(line, source) {
			t.Errorf("read-only path %q appears in the write allowance:\n%s", source, line)
		}
	}
}

func TestRenderBoundsExecutableMappingWhenAsked(t *testing.T) {
	agent := filepath.Join(t.TempDir(), "agent")
	if err := os.WriteFile(agent, []byte("x"), 0o600); err != nil {
		t.Fatalf("write agent: %v", err)
	}
	text := Profile{MapExecutablePaths: []string{agent}}.Render()

	if strings.Contains(text, "(allow file-map-executable)\n") {
		t.Errorf("bounded mapping still emits the unbounded rule:\n%s", text)
	}
	if !strings.Contains(text, `(allow file-map-executable`) || !strings.Contains(text, agent) {
		t.Errorf("bounded mapping does not name the agent:\n%s", text)
	}
	// The loader still has to be mappable or nothing starts.
	for _, path := range loaderReadPaths {
		if !strings.Contains(text, `"`+path+`"`) {
			t.Errorf("bounded mapping drops loader path %q:\n%s", path, text)
		}
	}
}

func TestRenderExecAllowlistUsesLiterals(t *testing.T) {
	agent := filepath.Join(t.TempDir(), "agent")
	if err := os.WriteFile(agent, []byte("x"), 0o600); err != nil {
		t.Fatalf("write agent: %v", err)
	}
	text := Profile{ExecPaths: []string{agent}}.Render()

	if !strings.Contains(text, `(allow process-exec* (literal "`+agent+`")`) {
		t.Errorf("exec allowance is not a literal path:\n%s", text)
	}
	// A subpath allowance would let anything under the directory start.
	if strings.Contains(text, `(allow process-exec* (subpath`) {
		t.Errorf("exec allowance uses a subpath, which is not an exact allowlist:\n%s", text)
	}
}

// -------------------------------------------------------------- path helpers

func TestQuoteEscapesProfileSyntax(t *testing.T) {
	cases := map[string]string{
		`/tmp/plain`:                   `"/tmp/plain"`,
		`/tmp/wi"th`:                   `"/tmp/wi\"th"`,
		`/tmp/back\slash`:              `"/tmp/back\\slash"`,
		`/tmp/"); (allow default) ; "`: `"/tmp/\"); (allow default) ; \""`,
	}
	for input, want := range cases {
		if got := quote(input); got != want {
			t.Errorf("quote(%q) = %s, want %s", input, got, want)
		}
	}
}

// stripLiterals removes every quoted string, honouring backslash escapes, so
// what is left is the profile's rule structure. Searching the raw text would
// find injected text that is in fact inert inside a literal.
func stripLiterals(text string) string {
	var out strings.Builder
	inLiteral, escaped := false, false
	for _, r := range text {
		switch {
		case escaped:
			escaped = false
		case r == '\\' && inLiteral:
			escaped = true
		case r == '"':
			inLiteral = !inLiteral
		case !inLiteral:
			out.WriteRune(r)
		}
	}
	return out.String()
}

// A path that could terminate its own literal would let a crafted directory
// name append rules to the profile.
func TestRenderCannotBeEscapedByAPathName(t *testing.T) {
	hostile := `/tmp/x") (allow default) (allow file-read* (literal "/`
	text := Profile{ReadOnlyPaths: []string{hostile}}.Render()

	if structure := stripLiterals(text); strings.Contains(structure, "allow default") {
		t.Fatalf("a path name injected a rule into the profile structure %q:\n%s", structure, text)
	}
	// The hostile characters must still be present, escaped, inside the literal:
	// silently dropping them would be a different bug.
	if !strings.Contains(text, `\"`) {
		t.Errorf("the hostile path was not escaped into the literal:\n%s", text)
	}
}

func TestStripLiteralsLeavesOnlyStructure(t *testing.T) {
	got := stripLiterals(`(allow file-read* (literal "/a\") (allow default) b") (subpath "/c"))`)
	want := `(allow file-read* (literal ) (subpath ))`
	if got != want {
		t.Errorf("stripLiterals = %q, want %q", got, want)
	}
}

func TestDedupeCollapsesEquivalentPaths(t *testing.T) {
	got := dedupe([]string{"/a/b", "/a/b/", "/a/./b", "/a/c", "/a/b", "/a/c/../c"})
	want := []string{"/a/b", "/a/c"}
	if len(got) != len(want) {
		t.Fatalf("dedupe = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("dedupe[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// A relative or empty path matches nothing once the kernel resolves it, so it
// must never reach a rule: a granted allowance that can never fire is
// indistinguishable from enforcement.
func TestDedupeDropsPathsThatCanNeverMatch(t *testing.T) {
	for _, path := range []string{"", ".", "..", "relative/dir", "./x"} {
		if got := dedupe([]string{path}); len(got) != 0 {
			t.Errorf("dedupe(%q) = %v, want nothing", path, got)
		}
	}
}

func TestRenderDropsRelativePaths(t *testing.T) {
	text := Profile{ReadOnlyPaths: []string{"relative/source"}, WritablePaths: []string{"relative/build"}}.Render()
	if strings.Contains(text, "relative/source") || strings.Contains(text, "relative/build") {
		t.Errorf("a relative path reached the profile:\n%s", text)
	}
}

// In the seatbelt language an operation with no filter after it allows the
// operation everywhere. A path list whose entries were all dropped must
// therefore emit no rule at all: emitting the bare keyword would turn a
// confinement request into a blanket grant.
func TestRenderNeverEmitsAnUnfilteredAllowance(t *testing.T) {
	profiles := map[string]Profile{
		"all writable paths dropped":  {WritablePaths: []string{"relative/build", "", "."}},
		"all exec paths dropped":      {ExecPaths: []string{"relative/agent", "", "."}},
		"both dropped":                {WritablePaths: []string{""}, ExecPaths: []string{".."}},
		"no paths at all":             {},
		"absolute alongside relative": {WritablePaths: []string{"relative/build"}, ExecPaths: []string{"relative/agent"}},
	}
	for name, profile := range profiles {
		t.Run(name, func(t *testing.T) {
			text := profile.Render()
			for _, bare := range []string{"(allow file-write*)", "(allow process-exec*)"} {
				if strings.Contains(text, bare) {
					t.Errorf("unfiltered allowance %s emitted:\n%s", bare, text)
				}
			}
			// The same shape with a trailing newline right after the keyword.
			for _, line := range strings.Split(text, "\n") {
				if line == "(allow file-write*" || line == "(allow process-exec*" {
					t.Errorf("allowance %q has no filter:\n%s", line, text)
				}
			}
		})
	}
}

func TestSubpathOrLiteralFollowsTheFilesystem(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	missing := filepath.Join(dir, "missing")

	if got := subpathOrLiteral(dir); !strings.HasPrefix(got, "(subpath ") {
		t.Errorf("subpathOrLiteral(dir) = %s, want a subpath", got)
	}
	if got := subpathOrLiteral(file); !strings.HasPrefix(got, "(literal ") {
		t.Errorf("subpathOrLiteral(file) = %s, want a literal", got)
	}
	// A path that does not exist is not a directory, so it must not become a
	// subpath allowance that would later cover a whole tree.
	if got := subpathOrLiteral(missing); !strings.HasPrefix(got, "(literal ") {
		t.Errorf("subpathOrLiteral(missing) = %s, want a literal", got)
	}
}

// ---------------------------------------------------------------- Available

func TestAvailableOnDarwin(t *testing.T) {
	ok, why := Available()
	if runtime.GOOS != "darwin" {
		if ok {
			t.Errorf("Available() = true on %s", runtime.GOOS)
		}
		return
	}
	if !ok {
		t.Errorf("Available() = false on darwin: %s", why)
	}
	if ok && why != "" {
		t.Errorf("Available() = true with reason %q, want an empty reason", why)
	}
}

// --------------------------------------------------------------- Runner.Run

func TestRunRejectsEmptyArgv(t *testing.T) {
	_, err := Runner{ProfileDir: t.TempDir()}.Run(context.Background(), "empty", Profile{}, nil, nil, nil)
	if err == nil {
		t.Fatal("Run with no argv returned no error")
	}
	if !strings.Contains(err.Error(), "empty argv") {
		t.Errorf("error %q does not mention the empty argv", err)
	}
}

func TestRunReportsAnUnwritableProfileDir(t *testing.T) {
	_, err := Runner{ProfileDir: filepath.Join(t.TempDir(), "does-not-exist")}.
		Run(context.Background(), "unwritable", Profile{}, []string{"/bin/echo"}, nil, nil)
	if err == nil {
		t.Fatal("Run with an unwritable profile directory returned no error")
	}
	if !strings.Contains(err.Error(), "write profile") {
		t.Errorf("error %q does not name the profile write", err)
	}
}

func TestRunExecutesAnAllowlistedProgram(t *testing.T) {
	requireDarwinSandbox(t)
	dir := t.TempDir()

	profile := Profile{
		ExecPaths:          []string{"/bin/echo"},
		MapExecutablePaths: []string{"/bin/echo"},
		AllowSysctlRead:    true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res, err := Runner{ProfileDir: dir}.Run(ctx, "echo", profile, []string{"/bin/echo", "contained"}, []string{}, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.LaunchErr != nil {
		t.Fatalf("probe domain did not start: %v (stderr %q)", res.LaunchErr, res.Stderr)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit %d, stderr %q", res.ExitCode, res.Stderr)
	}
	if strings.TrimSpace(res.Stdout) != "contained" {
		t.Errorf("stdout %q, want %q", res.Stdout, "contained")
	}

	// The profile that was actually applied has to be recoverable from the
	// result, otherwise the observation cannot be attributed to a rule set.
	if res.ProfilePath != filepath.Join(dir, "echo.sb") {
		t.Errorf("profile path %q, want %q", res.ProfilePath, filepath.Join(dir, "echo.sb"))
	}
	written, readErr := os.ReadFile(res.ProfilePath)
	if readErr != nil {
		t.Fatalf("read back the profile: %v", readErr)
	}
	if string(written) != res.ProfileText || res.ProfileText != profile.Render() {
		t.Errorf("the profile on disk is not the one reported")
	}
	if len(res.Argv) < 3 || res.Argv[0] != ExecPath || res.Argv[1] != "-f" {
		t.Errorf("argv %v does not record the launcher invocation", res.Argv)
	}
}

// A contained program that exits nonzero is not a launch failure: the domain
// worked, the program refused. Confusing the two would turn an observation into
// an inconclusive probe.
func TestRunDistinguishesExitStatusFromLaunchFailure(t *testing.T) {
	requireDarwinSandbox(t)
	if _, err := os.Stat("/usr/bin/false"); err != nil {
		t.Skipf("/usr/bin/false is not present: %v", err)
	}

	profile := Profile{
		ExecPaths:          []string{"/usr/bin/false"},
		MapExecutablePaths: []string{"/usr/bin/false"},
		AllowSysctlRead:    true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res, err := Runner{ProfileDir: t.TempDir()}.Run(ctx, "false", profile, []string{"/usr/bin/false"}, []string{}, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.LaunchErr != nil {
		t.Fatalf("a nonzero exit was reported as a launch failure: %v", res.LaunchErr)
	}
	if res.ExitCode == 0 {
		t.Errorf("exit %d, want nonzero", res.ExitCode)
	}
}

// The point of the whole package: a program that is not on the allowlist must
// not start inside the domain.
func TestRunDeniesAProgramOutsideTheAllowlist(t *testing.T) {
	requireDarwinSandbox(t)

	profile := Profile{
		ExecPaths:          []string{"/bin/echo"},
		MapExecutablePaths: []string{"/bin/echo"},
		AllowSysctlRead:    true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res, err := Runner{ProfileDir: t.TempDir()}.Run(ctx, "denied", profile, []string{"/bin/ls", "/"}, []string{}, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.ExitCode == 0 {
		t.Errorf("an unlisted program started: exit %d, stdout %q", res.ExitCode, res.Stdout)
	}
}

func TestRunRespectsACancelledContext(t *testing.T) {
	requireDarwinSandbox(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res, err := Runner{ProfileDir: t.TempDir()}.
		Run(ctx, "cancelled", Profile{ExecPaths: []string{"/bin/echo"}}, []string{"/bin/echo", "hi"}, []string{}, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// A cancelled context means the launcher never ran to completion. That is a
	// launch failure, not a host observation.
	if res.LaunchErr == nil && res.ExitCode == 0 {
		t.Errorf("a cancelled run reported success: exit %d", res.ExitCode)
	}
}
