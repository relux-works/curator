package shell

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/hookapproval"
)

// hookTrustVector is the subset of conformance/v1/vectors/shell-hook-trust.json
// this execution consumes.
type hookTrustVector struct {
	Fixtures map[string]hookTrustFixture `json:"fixtures"`
	Cases    []hookTrustCase             `json:"cases"`
}

type hookTrustFixture struct {
	File        string `json:"file"`
	BytesBase64 string `json:"bytes_base64"`
	SHA256      string `json:"sha256"`
}

type hookTrustRecord struct {
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	ApprovedBy string `json:"approved_by"`
	ApprovedAt string `json:"approved_at"`
}

type hookTrustCase struct {
	Name                  string           `json:"name"`
	File                  string           `json:"file"`
	CandidateFixture      string           `json:"candidate_fixture"`
	RolloutProfile        string           `json:"rollout_profile"`
	ManagerRecord         *hookTrustRecord `json:"manager_approval_record"`
	ProjectSuppliedRecord *struct {
		Source string          `json:"source"`
		Record hookTrustRecord `json:"record"`
	} `json:"project_supplied_record"`
	ObservedSHA256 string  `json:"observed_sha256"`
	Diagnostic     *string `json:"diagnostic"`
	Sourced        bool    `json:"sourced"`
	// SourcedMarker is the harness-only expected ${CURATOR_PROJECT_ENV}
	// value when Sourced is true; empty means "1". Every vector fixture
	// exports 1 (even the changed-bytes fixtures, which add a second
	// variable), so decoded cases keep the zero value; hand-built cases
	// whose bytes export another value set it explicitly.
	SourcedMarker               string `json:"-"`
	WarningFirst                bool   `json:"warning_first_activation"`
	WarningSecond               bool   `json:"warning_second_activation_same_session"`
	WarningsTotal               int    `json:"warnings_total_across_two_activations"`
	WarningNamesPath            bool   `json:"warning_names_path"`
	WarningNamesApprovalCommand bool   `json:"warning_names_approval_command"`
}

// TestShellHookTrustVectors executes every shell-hook-trust conformance case
// against the emitted hook, following the §8.7 recipe: materialize the
// fixture bytes, seed exactly the case's manager-home record, place a forged
// record only in project data, generate the hook for the case's profile, run
// activation twice in one shell session, and compare sourced, diagnostic, and
// warning count with the case's expectations.
//
// The vector's absolute candidate paths (/project/...) cannot be materialized
// on a test host, so the harness remaps them below a resolved temporary
// directory: the record path and the materialized path move together, and the
// hook observes the remapped absolute path.
func TestShellHookTrustVectors(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "shell-hook-trust.json")) // #nosec G304 -- explicit conformance input
	if os.IsNotExist(err) {
		t.Skipf("%s publishes no shell-hook-trust vector", root)
	}
	if err != nil {
		t.Fatal(err)
	}
	var vectors hookTrustVector
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors.Cases) == 0 {
		t.Fatal("shell-hook-trust vector publishes no cases")
	}
	conformancecoverage.Run(t, "shell-hook-trust/vectors", vectors.Cases,
		func(tc hookTrustCase) string { return tc.Name }, func(t *testing.T, c hookTrustCase) {
			runHookTrustCase(t, c, vectors.Fixtures)
		})
}

func runHookTrustCase(t *testing.T, c hookTrustCase, fixtures map[string]hookTrustFixture) {
	t.Helper()
	fixture, known := fixtures[c.CandidateFixture]
	if !known {
		t.Fatalf("case names unknown fixture %q", c.CandidateFixture)
	}
	if fixture.File != c.File {
		t.Fatalf("fixture %q is for %q, case wants %q", c.CandidateFixture, fixture.File, c.File)
	}
	content, err := base64.StdEncoding.DecodeString(fixture.BytesBase64)
	if err != nil {
		t.Fatal(err)
	}
	if hookapproval.Digest(content) != fixture.SHA256 || fixture.SHA256 != c.ObservedSHA256 {
		t.Fatal("fixture bytes do not match the vector digests")
	}
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	project := filepath.Join(base, "project")
	agentsDir := filepath.Join(project, ".agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	candidate := filepath.Join(agentsDir, filepath.Base(c.File))
	if err := os.WriteFile(candidate, content, 0o600); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(base, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	if c.ManagerRecord != nil {
		stamp, err := time.Parse(time.RFC3339, c.ManagerRecord.ApprovedAt)
		if err != nil {
			t.Fatal(err)
		}
		record := hookapproval.Record{
			Path:       candidate,
			SHA256:     c.ManagerRecord.SHA256,
			ApprovedBy: c.ManagerRecord.ApprovedBy,
			ApprovedAt: stamp,
		}
		if err := hookapproval.Upsert(home, record); err != nil {
			t.Fatal(err)
		}
	}
	if c.ProjectSuppliedRecord != nil {
		forged, err := json.Marshal(c.ProjectSuppliedRecord.Record)
		if err != nil {
			t.Fatal(err)
		}
		// The forged record lives only in project data; the hook must ignore it.
		if err := os.WriteFile(filepath.Join(agentsDir, "hook-approvals.json"), forged, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var profile TrustProfile
	switch c.RolloutProfile {
	case string(TrustProfileAWarning):
		profile = TrustProfileAWarning
	case string(TrustProfileBEnforcing):
		profile = TrustProfileBEnforcing
	default:
		t.Fatalf("case names unknown rollout profile %q", c.RolloutProfile)
	}
	switch c.File {
	case ".agents/env.sh":
		// The emitted POSIX hook must run under every POSIX interpreter,
		// not only bash/zsh: sh and dash (when present) execute the same
		// activation through the bash hook flavor.
		for _, shell := range posixTrustShells(t) {
			t.Run(shell.name, func(t *testing.T) {
				stdout, stderr := runPosixTrustActivation(t, shell, project, home, profile)
				assertTrustOutcome(t, c, candidate, stdout, stderr)
			})
		}
	case ".agents/env.ps1":
		stdout, stderr := runPowerShellTrustActivation(t, project, home, profile)
		assertTrustOutcome(t, c, candidate, stdout, stderr)
	default:
		t.Fatalf("case names unknown file %q", c.File)
	}
}

// assertTrustOutcome compares one two-activation observation with the case's
// expectations: sourced markers on stdout, diagnostics and warning counts on
// stderr, first/second split at the PROBE-1 marker.
func assertTrustOutcome(t *testing.T, c hookTrustCase, candidate, stdout, stderr string) {
	t.Helper()
	// PowerShell on Windows frames stderr lines with CRLF; normalize so the
	// diagnostic counts and the activation-probe split hold on every runner.
	stdout = strings.ReplaceAll(stdout, "\r\n", "\n")
	stderr = strings.ReplaceAll(stderr, "\r\n", "\n")
	wantSourced := "no"
	if c.Sourced {
		wantSourced = "1"
		if c.SourcedMarker != "" {
			wantSourced = c.SourcedMarker
		}
	}
	for _, marker := range []string{"sourced1=" + wantSourced, "sourced2=" + wantSourced} {
		if !strings.Contains(stdout, marker) {
			t.Fatalf("stdout lacks %q:\nstdout:\n%s\nstderr:\n%s", marker, stdout, stderr)
		}
	}
	if c.Diagnostic == nil {
		if strings.Contains(stderr, "shell_hook_env_") {
			t.Fatalf("trusted candidate warned:\n%s", stderr)
		}
	} else {
		if *c.Diagnostic != DiagnosticEnvUnapproved && *c.Diagnostic != DiagnosticEnvChanged {
			t.Fatalf("case expects unknown diagnostic %q", *c.Diagnostic)
		}
		if !strings.Contains(stderr, *c.Diagnostic) {
			t.Fatalf("stderr lacks diagnostic %q:\n%s", *c.Diagnostic, stderr)
		}
	}
	parts := strings.SplitN(stderr, "PROBE-1\n", 2)
	if len(parts) != 2 {
		t.Fatalf("stderr lacks the activation probe:\n%s", stderr)
	}
	if runtime.GOOS == "windows" {
		assertTrustWarningWindows(t, c, candidate, stderr, parts)
		return
	}
	approvalCommand := "curator hook approve " + candidate
	if got := strings.Count(stderr, approvalCommand); got != c.WarningsTotal {
		t.Fatalf("warning count = %d, want %d:\n%s", got, c.WarningsTotal, stderr)
	}
	if c.WarningNamesPath && !strings.Contains(stderr, candidate) {
		t.Fatalf("warning does not name %q:\n%s", candidate, stderr)
	}
	if c.WarningNamesApprovalCommand && !strings.Contains(stderr, approvalCommand) {
		t.Fatalf("warning does not name the approval command:\n%s", stderr)
	}
	for i, want := range []bool{c.WarningFirst, c.WarningSecond} {
		got := strings.Contains(parts[i], approvalCommand)
		if got != want {
			t.Fatalf("activation %d warned = %v, want %v:\n%s", i+1, got, want, stderr)
		}
	}
}

// assertTrustWarningWindows checks the warning half of a vector case by the
// spelling the hook itself prints on Windows: the sourced/diagnostic/warning
// outcomes stay exact, while the warned path is compared modulo the native
// vs MSYS spelling, separator, and case differences the filesystem treats
// as one file (hookapproval.Canonicalize identity rule). The count is taken
// on the approval-command prefix and every warned path token must normalize
// to the expected candidate, so a warning for the wrong path still fails.
func assertTrustWarningWindows(t *testing.T, c hookTrustCase, candidate, stderr string, parts []string) {
	t.Helper()
	const prefix = "curator hook approve "
	if got := strings.Count(stderr, prefix); got != c.WarningsTotal {
		t.Fatalf("warning count = %d, want %d:\n%s", got, c.WarningsTotal, stderr)
	}
	nativeNorm, msysNorm := windowsWarningSpellings(candidate)
	flat := strings.ToLower(strings.ReplaceAll(stderr, "\\", "/"))
	if c.WarningsTotal > 0 || c.WarningNamesPath || c.WarningNamesApprovalCommand {
		if c.WarningNamesPath || c.WarningsTotal > 0 {
			if !strings.Contains(flat, nativeNorm) && (msysNorm == "" || !strings.Contains(flat, msysNorm)) {
				t.Fatalf("warning does not name %q:\n%s", candidate, stderr)
			}
		}
	}
	for _, warned := range warnedApprovalPaths(stderr) {
		if normalizeWindowsPathToken(warned) != nativeNorm {
			t.Fatalf("warning names %q, want %q:\n%s", warned, candidate, stderr)
		}
	}
	for i, want := range []bool{c.WarningFirst, c.WarningSecond} {
		got := strings.Contains(parts[i], prefix)
		if got != want {
			t.Fatalf("activation %d warned = %v, want %v:\n%s", i+1, got, want, stderr)
		}
	}
}

// windowsWarningSpellings returns the normalized native spelling of a Go
// candidate path plus its MSYS equivalent (/c/...), both lowercased with
// forward slashes, for warning comparison on Windows.
func windowsWarningSpellings(candidate string) (nativeNorm, msysNorm string) {
	nativeNorm = strings.ToLower(strings.ReplaceAll(candidate, "\\", "/"))
	if len(nativeNorm) >= 3 && isWindowsDriveLetter(nativeNorm[0]) && nativeNorm[1] == ':' && nativeNorm[2] == '/' {
		msysNorm = "/" + string(nativeNorm[0]) + "/" + nativeNorm[3:]
	}
	return nativeNorm, msysNorm
}

// normalizeWindowsPathToken folds one warned path token to the native
// normalized form: separators unified, a leading MSYS /c/... or
// /cygdrive/c/... spelling converted to drive form, then lowercased.
func normalizeWindowsPathToken(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if len(p) >= 3 && p[0] == '/' && isWindowsDriveLetter(p[1]) && p[2] == '/' {
		p = string(p[1]) + ":/" + p[3:]
	}
	if strings.HasPrefix(p, "/cygdrive/") && len(p) >= 12 && isWindowsDriveLetter(p[10]) && p[11] == '/' {
		p = string(p[10]) + ":/" + p[12:]
	}
	return strings.ToLower(p)
}

func isWindowsDriveLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

// warnedApprovalPaths extracts the path token following every
// "curator hook approve " prefix in stderr.
func warnedApprovalPaths(stderr string) []string {
	const prefix = "curator hook approve "
	var out []string
	for {
		i := strings.Index(stderr, prefix)
		if i < 0 {
			return out
		}
		rest := stderr[i+len(prefix):]
		end := len(rest)
		if j := strings.Index(rest, " to approve"); j >= 0 && j < end {
			end = j
		}
		if j := strings.Index(rest, "\n"); j >= 0 && j < end {
			end = j
		}
		out = append(out, strings.TrimSpace(rest[:end]))
		stderr = rest[end:]
	}
}

func runPosixTrustActivation(t *testing.T, shell trustShell, project, home string, profile TrustProfile) (stdout, stderr string) {
	t.Helper()
	hook, err := HookWithProfile(shell.flavor, false, profile)
	if err != nil {
		t.Fatal(err)
	}
	hookPath := filepath.Join(t.TempDir(), "hook")
	if err := os.WriteFile(hookPath, []byte(hook), 0o600); err != nil {
		t.Fatal(err)
	}
	script := `cd "$PROJ"
. "$HOOK"
printf 'sourced1=%s\n' "${CURATOR_PROJECT_ENV:-no}"
printf 'PROBE-1\n' >&2
_curator_auto_env
printf 'sourced2=%s\n' "${CURATOR_PROJECT_ENV:-no}"
`
	return runTrustShell(t, shell.executable, script, map[string]string{
		"PROJ": project, "HOOK": hookPath,
		"CURATOR_CONFIG": filepath.Join(home, "config.json"),
	})
}

func runPowerShellTrustActivation(t *testing.T, project, home string, profile TrustProfile) (stdout, stderr string) {
	t.Helper()
	powerShell, err := exec.LookPath("pwsh")
	if err != nil {
		t.Skip("PowerShell is unavailable (no pwsh on this runner)")
	}
	hook, err := HookWithProfile("powershell", false, profile)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	hookPath := filepath.Join(dir, "hook.ps1")
	if err := os.WriteFile(hookPath, []byte(hook), 0o600); err != nil {
		t.Fatal(err)
	}
	script := "Set-Location $env:PROJ\n" +
		". $env:HOOK\n" +
		"if ($env:CURATOR_PROJECT_ENV) { Write-Output (\"sourced1=\" + $env:CURATOR_PROJECT_ENV) } else { Write-Output \"sourced1=no\" }\n" +
		"[Console]::Error.WriteLine(\"PROBE-1\")\n" +
		"Invoke-CuratorAutoEnv\n" +
		"if ($env:CURATOR_PROJECT_ENV) { Write-Output (\"sourced2=\" + $env:CURATOR_PROJECT_ENV) } else { Write-Output \"sourced2=no\" }\n"
	scriptPath := filepath.Join(dir, "activate.ps1")
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, powerShell, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	command.Env = append(os.Environ(),
		"PROJ="+project, "HOOK="+hookPath,
		"CURATOR_CONFIG="+filepath.Join(home, "config.json"))
	var out, errOut strings.Builder
	command.Stdout = &out
	command.Stderr = &errOut
	if err := command.Run(); err != nil {
		t.Fatalf("powershell activation: %v\nstdout:\n%s\nstderr:\n%s", err, out.String(), errOut.String())
	}
	return out.String(), errOut.String()
}

// trustShell is one POSIX interpreter the trust tests execute the emitted
// hook under, with the hook flavor it runs.
type trustShell struct {
	name       string
	executable string
	flavor     string
}

// posixTrustShells resolves every POSIX interpreter on this runner. Plain sh
// and dash (when present) run the bash hook flavor: the emitted hook is
// strictly POSIX-parseable, so the activation path must work there, not only
// under bash/zsh. On Windows the same probe finds Git Bash (bash.exe, sh.exe
// on PATH); when PATH lacks it the well-known Git for Windows install roots
// are checked, so the vector, hostile, malformed, and alias cases run
// wherever Git Bash is actually present. Interpreters resolving to the same
// binary run once. The skip names the absent interpreter (host-capability),
// never the GOOS alone.
func posixTrustShells(t *testing.T) []trustShell {
	t.Helper()
	var shells []trustShell
	seen := map[string]bool{}
	names := []string{"sh", "dash", "bash", "zsh"}
	if runtime.GOOS == "windows" {
		names = []string{"bash", "sh", "bash.exe", "sh.exe"}
	}
	for _, name := range names {
		executable, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		identity := executable
		if resolved, err := filepath.EvalSymlinks(executable); err == nil {
			identity = resolved
		}
		if seen[identity] {
			continue
		}
		seen[identity] = true
		flavor := "bash"
		if strings.TrimSuffix(strings.ToLower(name), ".exe") == "zsh" {
			flavor = "zsh"
		}
		shells = append(shells, trustShell{name: strings.TrimSuffix(strings.ToLower(name), ".exe"), executable: executable, flavor: flavor})
	}
	if runtime.GOOS == "windows" {
		for _, candidate := range []string{
			`C:\Program Files\Git\bin\bash.exe`,
			`C:\Program Files\Git\usr\bin\bash.exe`,
			`C:\Program Files\Git\bin\sh.exe`,
			`C:\Program Files\Git\usr\bin\sh.exe`,
		} {
			if _, err := os.Stat(candidate); err != nil {
				continue
			}
			identity := candidate
			if resolved, err := filepath.EvalSymlinks(candidate); err == nil {
				identity = resolved
			}
			if seen[identity] {
				continue
			}
			seen[identity] = true
			name := "bash"
			if strings.HasPrefix(strings.ToLower(filepath.Base(candidate)), "sh.") {
				name = "sh"
			}
			shells = append(shells, trustShell{name: name, executable: candidate, flavor: "bash"})
		}
	}
	if len(shells) == 0 {
		if runtime.GOOS == "windows" {
			t.Skip("bash is unavailable (Git Bash absent on this runner)")
		}
		t.Skip("bash is unavailable (no POSIX shell on this runner)")
	}
	return shells
}

func runTrustShell(t *testing.T, executable, script string, environment map[string]string) (stdout, stderr string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// Git Bash reports bash.exe/sh.exe; fold the suffix so the Windows
	// interpreters take the same argument shapes as their Unix names.
	base := strings.TrimSuffix(strings.ToLower(filepath.Base(executable)), ".exe")
	args := []string{"--noprofile", "--norc", "-c", script}
	switch base {
	case "zsh":
		args = []string{"-dfc", script}
	case "sh", "dash":
		// dash accepts no profile-suppression flags; bash-as-sh reads no
		// profile for -c either, but honors $ENV, which is cleared below.
		args = []string{"-c", script}
	}
	command := exec.CommandContext(ctx, executable, args...)
	command.Env = append([]string{}, os.Environ()...)
	command.Env = append(command.Env, "SHELL="+executable)
	for name, value := range environment {
		command.Env = append(command.Env, name+"="+value)
	}
	if base == "sh" {
		filtered := command.Env[:0]
		for _, entry := range command.Env {
			if entry == "ENV" || strings.HasPrefix(entry, "ENV=") {
				continue
			}
			filtered = append(filtered, entry)
		}
		command.Env = filtered
	}
	var out, errOut strings.Builder
	command.Stdout = &out
	command.Stderr = &errOut
	if err := command.Run(); err != nil {
		t.Fatalf("%s activation: %v\nstdout:\n%s\nstderr:\n%s",
			filepath.Base(executable), err, out.String(), errOut.String())
	}
	return out.String(), errOut.String()
}

// TestShellHookRefusesHostileCheckout simulates a hostile project checkout:
// foreign env-file bytes with no approval record. Under B-enforcing the
// payload must never execute; under A-warning it executes but warns once with
// the migration hint. Both profiles keep activation itself green.
func TestShellHookRefusesHostileCheckout(t *testing.T) {
	t.Run("posix", func(t *testing.T) {
		// Runs through Git Bash on Windows; posixTrustShells skips only
		// when no POSIX interpreter is present (host-capability).
		base, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		project := filepath.Join(base, "hostile")
		agentsDir := filepath.Join(project, ".agents")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		candidate := filepath.Join(agentsDir, "env.sh")
		payload := "printf 'pwned' > \"$HOSTILE_CANARY\"\n" + "export HOSTILE_MARKER=pwned\n"
		if err := os.WriteFile(candidate, []byte(payload), 0o600); err != nil {
			t.Fatal(err)
		}
		home := filepath.Join(base, "home")
		if err := os.MkdirAll(home, 0o755); err != nil {
			t.Fatal(err)
		}
		script := `cd "$PROJ"
. "$HOOK"
printf 'marker1=%s\n' "${HOSTILE_MARKER:-no}"
printf 'PROBE-1\n' >&2
_curator_auto_env
printf 'marker2=%s\n' "${HOSTILE_MARKER:-no}"
`
		for _, shell := range posixTrustShells(t) {
			t.Run(shell.name, func(t *testing.T) {
				canary := filepath.Join(t.TempDir(), "canary")
				run := func(t *testing.T, profile TrustProfile) (stdout, stderr string) {
					t.Helper()
					_ = os.Remove(canary)
					hook, err := HookWithProfile(shell.flavor, false, profile)
					if err != nil {
						t.Fatal(err)
					}
					hookPath := filepath.Join(t.TempDir(), "hook")
					if err := os.WriteFile(hookPath, []byte(hook), 0o600); err != nil {
						t.Fatal(err)
					}
					return runTrustShell(t, shell.executable, script, map[string]string{
						"PROJ": project, "HOOK": hookPath, "HOSTILE_CANARY": canary,
						"CURATOR_CONFIG": filepath.Join(home, "config.json"),
					})
				}

				stdout, stderr := run(t, TrustProfileBEnforcing)
				if strings.Contains(stdout, "marker1=pwned") || strings.Contains(stdout, "marker2=pwned") {
					t.Fatalf("B-enforcing sourced the hostile payload:\n%s\n%s", stdout, stderr)
				}
				if _, err := os.Stat(canary); !os.IsNotExist(err) {
					t.Fatalf("B-enforcing executed the hostile payload (canary: %v)", err)
				}
				assertHostileWarning(t, candidate, stderr)

				stdout, stderr = run(t, TrustProfileAWarning)
				if !strings.Contains(stdout, "marker1=pwned") {
					t.Fatalf("A-warning did not source with a warning:\n%s\n%s", stdout, stderr)
				}
				if _, err := os.Stat(canary); err != nil {
					t.Fatalf("A-warning did not execute the payload it sources: %v", err)
				}
				assertHostileWarning(t, candidate, stderr)
			})
		}
	})

	t.Run("powershell", func(t *testing.T) {
		powerShell, err := exec.LookPath("pwsh")
		if err != nil {
			t.Skip("PowerShell is unavailable (no pwsh on this runner)")
		}
		base, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		project := filepath.Join(base, "hostile")
		agentsDir := filepath.Join(project, ".agents")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		candidate := filepath.Join(agentsDir, "env.ps1")
		canary := filepath.Join(base, "canary")
		payload := "Set-Content -LiteralPath $env:HOSTILE_CANARY -Value \"pwned\"\n" +
			"$env:HOSTILE_MARKER = \"pwned\"\n"
		if err := os.WriteFile(candidate, []byte(payload), 0o600); err != nil {
			t.Fatal(err)
		}
		home := filepath.Join(base, "home")
		if err := os.MkdirAll(home, 0o755); err != nil {
			t.Fatal(err)
		}
		run := func(t *testing.T, profile TrustProfile) (stdout, stderr string) {
			t.Helper()
			_ = os.Remove(canary)
			hook, err := HookWithProfile("powershell", false, profile)
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			hookPath := filepath.Join(dir, "hook.ps1")
			if err := os.WriteFile(hookPath, []byte(hook), 0o600); err != nil {
				t.Fatal(err)
			}
			script := "Set-Location $env:PROJ\n" +
				". $env:HOOK\n" +
				"if ($env:HOSTILE_MARKER) { Write-Output (\"marker1=\" + $env:HOSTILE_MARKER) } else { Write-Output \"marker1=no\" }\n" +
				"[Console]::Error.WriteLine(\"PROBE-1\")\n" +
				"Invoke-CuratorAutoEnv\n" +
				"if ($env:HOSTILE_MARKER) { Write-Output (\"marker2=\" + $env:HOSTILE_MARKER) } else { Write-Output \"marker2=no\" }\n"
			scriptPath := filepath.Join(dir, "activate.ps1")
			if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, powerShell, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
			command.Env = append(os.Environ(),
				"PROJ="+project, "HOOK="+hookPath, "HOSTILE_CANARY="+canary,
				"CURATOR_CONFIG="+filepath.Join(home, "config.json"))
			var out, errOut strings.Builder
			command.Stdout = &out
			command.Stderr = &errOut
			if err := command.Run(); err != nil {
				t.Fatalf("powershell activation: %v\nstdout:\n%s\nstderr:\n%s", err, out.String(), errOut.String())
			}
			return out.String(), errOut.String()
		}

		stdout, stderr := run(t, TrustProfileBEnforcing)
		if strings.Contains(stdout, "marker1=pwned") || strings.Contains(stdout, "marker2=pwned") {
			t.Fatalf("B-enforcing sourced the hostile payload:\n%s\n%s", stdout, stderr)
		}
		if _, err := os.Stat(canary); !os.IsNotExist(err) {
			t.Fatalf("B-enforcing executed the hostile payload (canary: %v)", err)
		}
		assertHostileWarning(t, candidate, stderr)

		stdout, stderr = run(t, TrustProfileAWarning)
		if !strings.Contains(stdout, "marker1=pwned") {
			t.Fatalf("A-warning did not source with a warning:\n%s\n%s", stdout, stderr)
		}
		if _, err := os.Stat(canary); err != nil {
			t.Fatalf("A-warning did not execute the payload it sources: %v", err)
		}
		assertHostileWarning(t, candidate, stderr)
	})
}

// assertHostileWarning proves the hostile payload raised exactly one
// unapproved warning naming the path and the approval command: on the first
// activation, never on the second.
func assertHostileWarning(t *testing.T, candidate, stderr string) {
	t.Helper()
	// PowerShell on Windows frames stderr lines with CRLF; normalize so the
	// warning count and the activation-probe split hold on every runner.
	stderr = strings.ReplaceAll(stderr, "\r\n", "\n")
	if !strings.Contains(stderr, DiagnosticEnvUnapproved) {
		t.Fatalf("stderr lacks %q:\n%s", DiagnosticEnvUnapproved, stderr)
	}
	if strings.Contains(stderr, DiagnosticEnvChanged) {
		t.Fatalf("unapproved payload raised the changed diagnostic:\n%s", stderr)
	}
	parts := strings.SplitN(stderr, "PROBE-1\n", 2)
	if len(parts) != 2 {
		t.Fatalf("stderr lacks the activation probe:\n%s", stderr)
	}
	if runtime.GOOS == "windows" {
		const prefix = "curator hook approve "
		if got := strings.Count(stderr, prefix); got != 1 {
			t.Fatalf("warning count = %d, want 1:\n%s", got, stderr)
		}
		nativeNorm, msysNorm := windowsWarningSpellings(candidate)
		flat := strings.ToLower(strings.ReplaceAll(stderr, "\\", "/"))
		if !strings.Contains(flat, nativeNorm) && (msysNorm == "" || !strings.Contains(flat, msysNorm)) {
			t.Fatalf("warning does not name %q:\n%s", candidate, stderr)
		}
		for _, warned := range warnedApprovalPaths(stderr) {
			if normalizeWindowsPathToken(warned) != nativeNorm {
				t.Fatalf("warning names %q, want %q:\n%s", warned, candidate, stderr)
			}
		}
		if !strings.Contains(parts[0], prefix) || strings.Contains(parts[1], prefix) {
			t.Fatalf("warning is not first-activation-only:\n%s", stderr)
		}
		return
	}
	approvalCommand := "curator hook approve " + candidate
	if got := strings.Count(stderr, approvalCommand); got != 1 {
		t.Fatalf("warning count = %d, want 1:\n%s", got, stderr)
	}
	if !strings.Contains(parts[0], approvalCommand) || strings.Contains(parts[1], approvalCommand) {
		t.Fatalf("warning is not first-activation-only:\n%s", stderr)
	}
}

// TestGeneratedHookParsesUnderPOSIXShells syntax-checks every generated POSIX
// hook (both flavors, both profiles, with and without the global section)
// under each interpreter on the runner. It names the interpreters explicitly
// instead of depending on the host shell, and skips only the interpreters
// the runner lacks.
func TestGeneratedHookParsesUnderPOSIXShells(t *testing.T) {
	type checker struct {
		name       string
		executable string
	}
	var checkers []checker
	for _, name := range []string{"sh", "dash", "bash", "zsh"} {
		executable, err := exec.LookPath(name)
		if err != nil {
			t.Logf("%s is unavailable, skipping its syntax check", name)
			continue
		}
		checkers = append(checkers, checker{name: name, executable: executable})
	}
	if len(checkers) == 0 {
		t.Skip("no POSIX shell on this runner")
	}
	for _, flavor := range []string{"bash", "zsh"} {
		for _, profile := range []TrustProfile{TrustProfileAWarning, TrustProfileBEnforcing} {
			for _, includeGlobal := range []bool{false, true} {
				hook, err := HookWithProfile(flavor, includeGlobal, profile)
				if err != nil {
					t.Fatal(err)
				}
				if hook == "" {
					t.Fatalf("HookWithProfile(%s, %v, %s) is empty", flavor, includeGlobal, profile)
				}
				hookPath := filepath.Join(t.TempDir(), "hook")
				if err := os.WriteFile(hookPath, []byte(hook), 0o600); err != nil {
					t.Fatal(err)
				}
				for _, checker := range checkers {
					t.Run(flavor+"/"+string(profile)+"/global-"+boolName(includeGlobal)+"/"+checker.name, func(t *testing.T) {
						ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
						defer cancel()
						output, err := exec.CommandContext(ctx, checker.executable, "-n", hookPath).CombinedOutput()
						if err != nil {
							t.Fatalf("%s -n %s %s (global=%v): %v\n%s",
								checker.name, flavor, profile, includeGlobal, err, output)
						}
					})
				}
			}
		}
	}
}

func boolName(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

// TestHookEmbedsClosedRecordGrammar proves the emitted hooks validate the
// same closed record the Go reader enforces: the generated text carries the
// hookapproval grammar constants verbatim, and no unexpanded placeholder
// survives generation.
func TestHookEmbedsClosedRecordGrammar(t *testing.T) {
	fieldCount := strconv.Itoa(hookapproval.RecordFieldCount)
	digestLength := strconv.Itoa(hookapproval.SHA256HexLength)
	for _, flavor := range []string{"bash", "zsh", "powershell"} {
		for _, profile := range []TrustProfile{TrustProfileAWarning, TrustProfileBEnforcing} {
			hook, err := HookWithProfile(flavor, false, profile)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(hook, "__CURATOR_") {
				t.Fatalf("%s %s hook carries an unexpanded placeholder", flavor, profile)
			}
			needles := []string{hookapproval.TimestampShape, fieldCount, digestLength}
			if flavor == "powershell" {
				needles = append(needles,
					"'"+hookapproval.ApprovedByManager+"'",
					"'"+hookapproval.ApprovedByOperator+"'")
			} else {
				needles = append(needles,
					`"`+hookapproval.ApprovedByManager+`"`,
					`"`+hookapproval.ApprovedByOperator+`"`)
			}
			for _, needle := range needles {
				if !strings.Contains(hook, needle) {
					t.Fatalf("%s %s hook lacks grammar %q", flavor, profile, needle)
				}
			}
		}
	}
}

// TestShellHookRejectsMalformedRecords seeds manager-home state lines that
// carry the candidate path with the matching digest but a malformed record
// (open shape, forged approver, bad timestamp), and proves the emitted hooks
// treat the candidate as unapproved under both rollout profiles: refused
// under B-enforcing, sourced with exactly one first-activation warning under
// A-warning. The Go reader must reject every seeded line, proving the test
// premise and the reader/hook agreement.
func TestShellHookRejectsMalformedRecords(t *testing.T) {
	unapproved := DiagnosticEnvUnapproved
	cases := []struct {
		name  string
		state func(candidate, digest string) string
	}{
		{"two-fields", func(candidate, digest string) string {
			return candidate + "\t" + digest + "\n"
		}},
		{"five-fields", func(candidate, digest string) string {
			return candidate + "\t" + digest + "\tmanager\t2026-09-16T00:00:00Z\textra\n"
		}},
		{"forged-approver", func(candidate, digest string) string {
			return candidate + "\t" + digest + "\tproject\t2026-09-16T00:00:00Z\n"
		}},
		{"empty-approver", func(candidate, digest string) string {
			return candidate + "\t" + digest + "\t\t2026-09-16T00:00:00Z\n"
		}},
		{"uppercase-approver", func(candidate, digest string) string {
			return candidate + "\t" + digest + "\tManager\t2026-09-16T00:00:00Z\n"
		}},
		{"uppercase-digest", func(candidate, digest string) string {
			return candidate + "\t" + strings.ToUpper(digest) + "\tmanager\t2026-09-16T00:00:00Z\n"
		}},
		{"short-digest", func(candidate, _ string) string {
			return candidate + "\tshort\tmanager\t2026-09-16T00:00:00Z\n"
		}},
		{"nonhex-digest", func(candidate, _ string) string {
			return candidate + "\t" + strings.Repeat("z", 64) + "\tmanager\t2026-09-16T00:00:00Z\n"
		}},
		{"bad-timestamp", func(candidate, digest string) string {
			return candidate + "\t" + digest + "\tmanager\tnot-a-time\n"
		}},
		{"empty-timestamp", func(candidate, digest string) string {
			return candidate + "\t" + digest + "\tmanager\t\n"
		}},
		{"month-13", func(candidate, digest string) string {
			return candidate + "\t" + digest + "\tmanager\t2026-13-01T00:00:00Z\n"
		}},
		{"february-30", func(candidate, digest string) string {
			return candidate + "\t" + digest + "\tmanager\t2026-02-30T00:00:00Z\n"
		}},
		{"second-60", func(candidate, digest string) string {
			return candidate + "\t" + digest + "\tmanager\t2026-09-16T00:00:60Z\n"
		}},
		{"lowercase-timestamp", func(candidate, digest string) string {
			return candidate + "\t" + digest + "\tmanager\t2026-09-16t00:00:00z\n"
		}},
	}
	for _, malformed := range cases {
		t.Run(malformed.name, func(t *testing.T) {
			t.Run("posix", func(t *testing.T) {
				// Runs through Git Bash on Windows; posixTrustShells
				// skips only when no POSIX interpreter is present.
				for _, shell := range posixTrustShells(t) {
					t.Run(shell.name, func(t *testing.T) {
						base, err := filepath.EvalSymlinks(t.TempDir())
						if err != nil {
							t.Fatal(err)
						}
						project := filepath.Join(base, "project")
						agentsDir := filepath.Join(project, ".agents")
						if err := os.MkdirAll(agentsDir, 0o755); err != nil {
							t.Fatal(err)
						}
						content := []byte("export CURATOR_PROJECT_ENV=1\n")
						candidate := filepath.Join(agentsDir, "env.sh")
						if err := os.WriteFile(candidate, content, 0o600); err != nil {
							t.Fatal(err)
						}
						home := filepath.Join(base, "home")
						if err := os.MkdirAll(home, 0o755); err != nil {
							t.Fatal(err)
						}
						digest := hookapproval.Digest(content)
						if err := os.WriteFile(hookapproval.ApprovalsPath(home),
							[]byte(malformed.state(candidate, digest)), 0o600); err != nil {
							t.Fatal(err)
						}
						if _, err := hookapproval.List(home); err == nil {
							t.Fatalf("Go reader accepted the %s state line", malformed.name)
						}
						for _, profile := range []TrustProfile{TrustProfileBEnforcing, TrustProfileAWarning} {
							sourced := profile == TrustProfileAWarning
							expect := hookTrustCase{
								Name:                        malformed.name + "/" + string(profile),
								Diagnostic:                  &unapproved,
								Sourced:                     sourced,
								WarningFirst:                true,
								WarningSecond:               false,
								WarningsTotal:               1,
								WarningNamesPath:            true,
								WarningNamesApprovalCommand: true,
							}
							stdout, stderr := runPosixTrustActivation(t, shell, project, home, profile)
							assertTrustOutcome(t, expect, candidate, stdout, stderr)
						}
					})
				}
			})
			t.Run("powershell", func(t *testing.T) {
				if _, err := exec.LookPath("pwsh"); err != nil {
					t.Skip("PowerShell is unavailable (no pwsh on this runner)")
				}
				base, err := filepath.EvalSymlinks(t.TempDir())
				if err != nil {
					t.Fatal(err)
				}
				project := filepath.Join(base, "project")
				agentsDir := filepath.Join(project, ".agents")
				if err := os.MkdirAll(agentsDir, 0o755); err != nil {
					t.Fatal(err)
				}
				content := []byte("$env:CURATOR_PROJECT_ENV = \"1\"\n")
				candidate := filepath.Join(agentsDir, "env.ps1")
				if err := os.WriteFile(candidate, content, 0o600); err != nil {
					t.Fatal(err)
				}
				home := filepath.Join(base, "home")
				if err := os.MkdirAll(home, 0o755); err != nil {
					t.Fatal(err)
				}
				digest := hookapproval.Digest(content)
				if err := os.WriteFile(hookapproval.ApprovalsPath(home),
					[]byte(malformed.state(candidate, digest)), 0o600); err != nil {
					t.Fatal(err)
				}
				if _, err := hookapproval.List(home); err == nil {
					t.Fatalf("Go reader accepted the %s state line", malformed.name)
				}
				for _, profile := range []TrustProfile{TrustProfileBEnforcing, TrustProfileAWarning} {
					sourced := profile == TrustProfileAWarning
					expect := hookTrustCase{
						Name:                        malformed.name + "/" + string(profile),
						Diagnostic:                  &unapproved,
						Sourced:                     sourced,
						WarningFirst:                true,
						WarningSecond:               false,
						WarningsTotal:               1,
						WarningNamesPath:            true,
						WarningNamesApprovalCommand: true,
					}
					stdout, stderr := runPowerShellTrustActivation(t, project, home, profile)
					assertTrustOutcome(t, expect, candidate, stdout, stderr)
				}
			})
		})
	}
}

// TestShellHookTrustResolvesSymlinkedProject opens a manager-approved project
// through a directory alias and proves the emitted hooks source it silently
// under both rollout profiles: two spellings of one file resolve to the one
// record, with no warning on either activation.
func TestShellHookTrustResolvesSymlinkedProject(t *testing.T) {
	t.Run("posix", func(t *testing.T) {
		// Runs through Git Bash on Windows; posixTrustShells skips only
		// when no POSIX interpreter is present (host-capability).
		for _, shell := range posixTrustShells(t) {
			t.Run(shell.name, func(t *testing.T) {
				base, err := filepath.EvalSymlinks(t.TempDir())
				if err != nil {
					t.Fatal(err)
				}
				realProject := filepath.Join(base, "real-project")
				agentsDir := filepath.Join(realProject, ".agents")
				if err := os.MkdirAll(agentsDir, 0o755); err != nil {
					t.Fatal(err)
				}
				content := []byte("export CURATOR_PROJECT_ENV=1\n")
				candidate := filepath.Join(agentsDir, "env.sh")
				if err := os.WriteFile(candidate, content, 0o600); err != nil {
					t.Fatal(err)
				}
				alias := filepath.Join(base, "alias-project")
				if err := os.Symlink(realProject, alias); err != nil {
					t.Skipf("cannot stage a symlinked project directory: %v", err)
				}
				home := filepath.Join(base, "home")
				if err := os.MkdirAll(home, 0o755); err != nil {
					t.Fatal(err)
				}
				if _, err := hookapproval.ApproveFile(home, candidate, hookapproval.ApprovedByManager, time.Now().UTC()); err != nil {
					t.Fatal(err)
				}
				for _, profile := range []TrustProfile{TrustProfileBEnforcing, TrustProfileAWarning} {
					expect := hookTrustCase{
						Name:         "symlinked-project/" + string(profile),
						Sourced:      true,
						WarningFirst: false, WarningSecond: false,
					}
					stdout, stderr := runPosixTrustActivation(t, shell, alias, home, profile)
					assertTrustOutcome(t, expect, candidate, stdout, stderr)
				}
			})
		}
	})
	t.Run("powershell", func(t *testing.T) {
		if _, err := exec.LookPath("pwsh"); err != nil {
			t.Skip("PowerShell is unavailable (no pwsh on this runner)")
		}
		base, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		realProject := filepath.Join(base, "real-project")
		agentsDir := filepath.Join(realProject, ".agents")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		content := []byte("$env:CURATOR_PROJECT_ENV = \"1\"\n")
		candidate := filepath.Join(agentsDir, "env.ps1")
		if err := os.WriteFile(candidate, content, 0o600); err != nil {
			t.Fatal(err)
		}
		alias := filepath.Join(base, "alias-project")
		if err := os.Symlink(realProject, alias); err != nil {
			t.Skipf("cannot stage a symlinked project directory: %v", err)
		}
		home := filepath.Join(base, "home")
		if err := os.MkdirAll(home, 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := hookapproval.ApproveFile(home, candidate, hookapproval.ApprovedByManager, time.Now().UTC()); err != nil {
			t.Fatal(err)
		}
		for _, profile := range []TrustProfile{TrustProfileBEnforcing, TrustProfileAWarning} {
			expect := hookTrustCase{
				Name:         "symlinked-project/" + string(profile),
				Sourced:      true,
				WarningFirst: false, WarningSecond: false,
			}
			stdout, stderr := runPowerShellTrustActivation(t, alias, home, profile)
			assertTrustOutcome(t, expect, candidate, stdout, stderr)
		}
	})
}

// TestShellHookTrustNativeRecordAuthorizesMSYSSpelling proves the Windows
// identity rule end to end: one native manager record authorizes the same
// file reached through MSYS spelling under Git Bash. The shell is shown to
// observe MSYS spelling (its PWD starts with "/"), the approved bytes
// source silently under both rollout profiles, and changed bytes are
// refused under B-enforcing (sourced with a changed warning under
// A-warning). It runs only on Windows; the interpreter probe inside
// posixTrustShells skips with a host-capability reason when Git Bash is
// genuinely absent.
func TestShellHookTrustNativeRecordAuthorizesMSYSSpelling(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("native/MSYS spelling identity is exercised on Windows")
	}
	for _, shell := range posixTrustShells(t) {
		t.Run(shell.name, func(t *testing.T) {
			base, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			project := filepath.Join(base, "project")
			agentsDir := filepath.Join(project, ".agents")
			if err := os.MkdirAll(agentsDir, 0o755); err != nil {
				t.Fatal(err)
			}
			content := []byte("export CURATOR_PROJECT_ENV=1\n")
			candidate := filepath.Join(agentsDir, "env.sh")
			if err := os.WriteFile(candidate, content, 0o600); err != nil {
				t.Fatal(err)
			}
			home := filepath.Join(base, "home")
			if err := os.MkdirAll(home, 0o755); err != nil {
				t.Fatal(err)
			}
			// The record is stored under the native spelling the Go
			// writer canonicalizes; the hook below reaches the same
			// file through MSYS spelling.
			if _, err := hookapproval.ApproveFile(home, candidate, hookapproval.ApprovedByManager, time.Now().UTC()); err != nil {
				t.Fatal(err)
			}
			pwdOut, _ := runTrustShell(t, shell.executable, `cd "$PROJ" && pwd`, map[string]string{"PROJ": project})
			pwdOut = strings.ReplaceAll(strings.TrimSpace(pwdOut), "\r\n", "\n")
			if !strings.HasPrefix(pwdOut, "/") {
				t.Fatalf("Git Bash does not observe MSYS spelling, PWD = %q", pwdOut)
			}
			for _, profile := range []TrustProfile{TrustProfileAWarning, TrustProfileBEnforcing} {
				expect := hookTrustCase{
					Name:         "native-record-msys-lookup/" + string(profile),
					Sourced:      true,
					WarningFirst: false, WarningSecond: false,
				}
				stdout, stderr := runPosixTrustActivation(t, shell, project, home, profile)
				assertTrustOutcome(t, expect, candidate, stdout, stderr)
			}
			changed := DiagnosticEnvChanged
			if err := os.WriteFile(candidate, []byte("export CURATOR_PROJECT_ENV=2\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			expectB := hookTrustCase{
				Name:                        "native-record-msys-lookup/changed/B-enforcing",
				Diagnostic:                  &changed,
				Sourced:                     false,
				WarningFirst:                true,
				WarningSecond:               false,
				WarningsTotal:               1,
				WarningNamesPath:            true,
				WarningNamesApprovalCommand: true,
			}
			stdout, stderr := runPosixTrustActivation(t, shell, project, home, TrustProfileBEnforcing)
			assertTrustOutcome(t, expectB, candidate, stdout, stderr)
			expectA := expectB
			expectA.Name = "native-record-msys-lookup/changed/A-warning"
			expectA.Sourced = true
			// The rotated bytes export 2, so a sourced activation prints
			// sourced1=2/sourced2=2: the marker proves the NEW bytes were
			// sourced under A-warning, not stale state.
			expectA.SourcedMarker = "2"
			stdout, stderr = runPosixTrustActivation(t, shell, project, home, TrustProfileAWarning)
			assertTrustOutcome(t, expectA, candidate, stdout, stderr)
		})
	}
}

// TestWindowsWarningNormalization proves the Windows warning comparison
// folds the spellings the hook may print: native backslashes, forward
// slashes, MSYS /c/... and /cygdrive/c/... forms, and letter case all
// name one file. It runs on every platform because it exercises the
// comparison helpers directly, not the Windows-only activation path.
func TestWindowsWarningNormalization(t *testing.T) {
	candidate := `C:\Users\me\proj\.agents\env.sh`
	nativeNorm, msysNorm := windowsWarningSpellings(candidate)
	if nativeNorm != "c:/users/me/proj/.agents/env.sh" {
		t.Fatalf("native spelling = %q", nativeNorm)
	}
	if msysNorm != "/c/users/me/proj/.agents/env.sh" {
		t.Fatalf("msys spelling = %q", msysNorm)
	}
	for _, spelling := range []string{
		`C:\Users\me\proj\.agents\env.sh`,
		`c:\users\me\proj\.agents\env.sh`,
		`C:/Users/me/proj/.agents/env.sh`,
		`/c/Users/me/proj/.agents/env.sh`,
		`/C/Users/me/proj/.agents/env.sh`,
		`/cygdrive/c/Users/me/proj/.agents/env.sh`,
	} {
		if got := normalizeWindowsPathToken(spelling); got != nativeNorm {
			t.Fatalf("normalize(%q) = %q, want %q", spelling, got, nativeNorm)
		}
	}
	if got := normalizeWindowsPathToken(`D:\unrelated\env.sh`); got == nativeNorm {
		t.Fatalf("a different drive folded to the candidate: %q", got)
	}
	stderr := "curator: shell_hook_env_unapproved: /c/Users/me/proj/.agents/env.sh is not approved; " +
		"run curator hook approve /c/Users/me/proj/.agents/env.sh to approve it\n"
	paths := warnedApprovalPaths(stderr)
	if len(paths) != 1 || paths[0] != "/c/Users/me/proj/.agents/env.sh" {
		t.Fatalf("warned paths = %q", paths)
	}
	if normalizeWindowsPathToken(paths[0]) != nativeNorm {
		t.Fatalf("warned MSYS path does not name the native candidate")
	}
}

// TestPosixHookEmitsWindowsIdentity proves the generated POSIX hook carries
// the single-identity translation: MSYS detection, cygpath -w mapping with
// a documented refusal when cygpath is absent, drive-letter normalization,
// and the case-folded record lookup. Functional proof runs on Windows CI
// through Git Bash; this shape test pins the emission on every platform.
func TestPosixHookEmitsWindowsIdentity(t *testing.T) {
	for _, profile := range []TrustProfile{TrustProfileAWarning, TrustProfileBEnforcing} {
		hook, err := HookWithProfile("bash", false, profile)
		if err != nil {
			t.Fatal(err)
		}
		for _, needle := range []string{
			"_curator_trust_identity",
			"_curator_trust_is_windows_shell",
			"MINGW*|MSYS*|CYGWIN*",
			"cygpath -w",
			"tolower($1) == tolower(want)",
			"hookapproval.Canonicalize",
		} {
			if !strings.Contains(hook, needle) {
				t.Fatalf("%s POSIX hook lacks the Windows identity %q", profile, needle)
			}
		}
	}
}
