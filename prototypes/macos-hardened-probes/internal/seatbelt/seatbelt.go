// Package seatbelt builds macOS sandbox (seatbelt) profiles and runs a probe
// domain under /usr/bin/sandbox-exec.
//
// Mechanism status, measured on the host recorded in the outcome document:
//
//   - /usr/bin/sandbox-exec is a supported shipped binary, but its underlying
//     interface (sandbox_init and friends in libsystem_sandbox) is declared
//     deprecated in <sandbox.h>. The seatbelt profile language itself is not a
//     published, versioned interface.
//   - The App Sandbox path (entitlement plus code signature) is the supported
//     alternative and is packaging- and entitlement-dependent, so it cannot be
//     applied to an arbitrary already-built toolchain binary.
//
// This package therefore builds a probe domain, never a production build
// domain, and its results are recorded as capability observations, not as an
// enforcement claim.
package seatbelt

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExecPath is the shipped dynamic-sandbox launcher.
const ExecPath = "/usr/bin/sandbox-exec"

// Loader paths that every dynamically linked Mach-O needs before main runs.
//
// Determined empirically on macOS 26.5 (arm64): a (deny default) profile that
// omits read access to the root directory itself aborts the process during
// dyld startup with SIGABRT, even when every other needed subpath is allowed.
// Metadata-only access to "/" is not enough; the loader needs file-read-data.
var loaderReadPaths = []string{
	"/usr/lib",
	"/System/Library/dyld",
	"/System/Volumes/Preboot/Cryptexes/OS",
	"/private/var/db/dyld",
}

// Profile is a (deny default) seatbelt profile described by what it allows.
// Empty slices mean "allow nothing of that kind", which is the point: every
// allowance is explicit and every path is chosen by this package, never by the
// program that runs inside.
type Profile struct {
	// ReadOnlyPaths are presented read-only: file-read* with no file-write*.
	ReadOnlyPaths []string
	// WritablePaths are the only subtrees where mutation is allowed.
	WritablePaths []string
	// ExecPaths is the exact executable allowlist, by literal path.
	ExecPaths []string
	// MapExecutablePaths bounds where executable mappings may come from. When
	// empty, mapping is allowed wherever reading is, which leaves a dlopen path
	// open; the exec probe uses a bounded list to close it.
	MapExecutablePaths []string
	// AllowNetwork disables the default network denial. Used only by negative
	// controls that must show the probe can observe a success.
	AllowNetwork bool
	// AllowSysctlRead permits sysctl reads, which the Go runtime performs
	// during startup on some configurations.
	AllowSysctlRead bool
	// AllowProcessInfo permits process-info operations used by the domain
	// membership and termination probes.
	AllowProcessInfo bool
	// AllowSignals permits signalling other processes from inside the domain.
	AllowSignals bool
	// AllowForkExec permits fork; without it a domain member cannot create a
	// descendant at all, which would make the descendant probes vacuous.
	AllowForkExec bool
	// AllowMachLookup permits Mach service lookup. Off by default; the Go
	// runtime does not need it for the operations these probes perform.
	AllowMachLookup bool
}

// Render returns the seatbelt profile source.
//
// Every rule is emitted in a fixed order so the same Profile always renders the
// same bytes: a profile that changed between the probe and the report would
// make the observation unattributable.
func (p Profile) Render() string {
	var b strings.Builder
	b.WriteString("(version 1)\n")
	b.WriteString("(deny default)\n")

	readPaths := append([]string{}, loaderReadPaths...)
	readPaths = append(readPaths, p.ReadOnlyPaths...)
	readPaths = append(readPaths, p.WritablePaths...)
	readPaths = append(readPaths, p.ExecPaths...)

	// The root directory must be readable for dyld; it is granted as a literal
	// so no other top-level entry becomes readable through it.
	b.WriteString("(allow file-read* (literal \"/\")")
	for _, path := range dedupe(readPaths) {
		fmt.Fprintf(&b, " %s", subpathOrLiteral(path))
	}
	b.WriteString(")\n")

	// Each allowance is gated on the paths that survive dedupe, never on the
	// paths that were passed in. An operation keyword with no filter after it is
	// an unrestricted allowance, so a list whose every entry was dropped must
	// emit no rule at all rather than a bare "(allow file-write*)".
	if writable := dedupe(p.WritablePaths); len(writable) > 0 {
		b.WriteString("(allow file-write*")
		for _, path := range writable {
			fmt.Fprintf(&b, " (subpath %s)", quote(path))
		}
		b.WriteString(")\n")
	}

	if len(p.MapExecutablePaths) > 0 {
		mappable := dedupe(append(append([]string{}, loaderReadPaths...), p.MapExecutablePaths...))
		b.WriteString("(allow file-map-executable")
		for _, path := range mappable {
			fmt.Fprintf(&b, " %s", subpathOrLiteral(path))
		}
		b.WriteString(")\n")
	} else {
		b.WriteString("(allow file-map-executable)\n")
	}

	if execPaths := dedupe(p.ExecPaths); len(execPaths) > 0 {
		b.WriteString("(allow process-exec*")
		for _, path := range execPaths {
			fmt.Fprintf(&b, " (literal %s)", quote(path))
		}
		b.WriteString(")\n")
	}
	if p.AllowForkExec {
		b.WriteString("(allow process-fork)\n")
	}
	if p.AllowNetwork {
		b.WriteString("(allow network*)\n")
	}
	if p.AllowSysctlRead {
		b.WriteString("(allow sysctl-read)\n")
	}
	if p.AllowProcessInfo {
		b.WriteString("(allow process-info*)\n")
	}
	if p.AllowSignals {
		b.WriteString("(allow signal)\n")
	}
	if p.AllowMachLookup {
		b.WriteString("(allow mach-lookup)\n")
	}
	return b.String()
}

func subpathOrLiteral(path string) string {
	if isDir(path) {
		return fmt.Sprintf("(subpath %s)", quote(path))
	}
	return fmt.Sprintf("(literal %s)", quote(path))
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// quote renders a seatbelt string literal. Seatbelt paths are absolute and this
// package never accepts a path from package data, but quoting is still explicit
// so a path containing a quote or backslash cannot terminate the literal early.
func quote(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`)
	return `"` + replacer.Replace(s) + `"`
}

func dedupe(paths []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, path := range paths {
		clean := filepath.Clean(path)
		// A seatbelt filter matches the absolute path the kernel resolved, so a
		// relative or empty path (filepath.Clean turns "" into ".") matches
		// nothing. Rendering it would add a rule that can never fire while
		// looking like a granted allowance, so it is dropped here.
		if !filepath.IsAbs(clean) || seen[clean] {
			continue
		}
		seen[clean] = true
		out = append(out, clean)
	}
	return out
}

// Runner executes a command inside a probe domain described by a Profile.
type Runner struct {
	// ProfileDir is where rendered profiles are written. It must be readable
	// by sandbox-exec, which reads the profile before applying it.
	ProfileDir string
}

// Result is the observable outcome of one probe-domain run.
type Result struct {
	ProfilePath string
	ProfileText string
	Argv        []string
	ExitCode    int
	Stdout      string
	Stderr      string
	// LaunchErr is set when the domain could not be created at all, as opposed
	// to the contained program exiting nonzero. The distinction matters: a
	// domain that cannot be created is an inconclusive probe, and section 5.6
	// requires an inconclusive probe to reject.
	LaunchErr error
}

// Run renders the profile, writes it under ProfileDir, and executes argv inside
// the resulting probe domain.
func (r Runner) Run(ctx context.Context, name string, profile Profile, argv []string, env []string, extraFiles []*os.File) (Result, error) {
	if len(argv) == 0 {
		return Result{}, fmt.Errorf("seatbelt: empty argv")
	}
	text := profile.Render()
	path := filepath.Join(r.ProfileDir, name+".sb")
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		return Result{}, fmt.Errorf("seatbelt: write profile: %w", err)
	}

	full := append([]string{ExecPath, "-f", path}, argv...)
	// The executable is the fixed ExecPath constant and argv is chosen by the
	// harness; no probe data and no package byte reaches this call.
	cmd := exec.CommandContext(ctx, full[0], full[1:]...) //nolint:gosec // fixed executable, harness-chosen argv
	cmd.Env = env
	cmd.ExtraFiles = extraFiles
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	res := Result{
		ProfilePath: path,
		ProfileText: text,
		Argv:        full,
		ExitCode:    cmd.ProcessState.ExitCode(),
		Stdout:      stdout.String(),
		Stderr:      stderr.String(),
	}
	if runErr != nil {
		var exitErr *exec.ExitError
		if !asExitError(runErr, &exitErr) {
			res.LaunchErr = runErr
		}
	}
	return res, nil
}

func asExitError(err error, target **exec.ExitError) bool {
	exitErr, ok := err.(*exec.ExitError)
	if ok {
		*target = exitErr
	}
	return ok
}

// Available reports whether the dynamic sandbox launcher exists and is
// executable on this host.
func Available() (bool, string) {
	info, err := os.Stat(ExecPath)
	if err != nil {
		return false, fmt.Sprintf("%s: %v", ExecPath, err)
	}
	if info.Mode()&0o111 == 0 {
		return false, fmt.Sprintf("%s is not executable", ExecPath)
	}
	return true, ""
}
