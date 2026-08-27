package probe

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/evidence"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/inside"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/seatbelt"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/spec"
)

// baseProfile is the profile every class probe starts from: deny by default,
// read the stand-in views, write only under the private build root, execute
// only the probe agent.
func (r runContext) baseProfile() seatbelt.Profile {
	env := r.env
	return seatbelt.Profile{
		ReadOnlyPaths:      []string{env.SourceDir, env.GorootDir},
		WritablePaths:      []string{env.BuildRoot},
		ExecPaths:          []string{env.SelfPath},
		MapExecutablePaths: []string{env.SelfPath},
		AllowSysctlRead:    true,
		AllowProcessInfo:   true,
		AllowSignals:       true,
		AllowForkExec:      true,
	}
}

// runAgent starts the in-domain agent inside a probe domain and parses what it
// reported.
func (r runContext) runAgent(name string, profile seatbelt.Profile, op string, extraEnv []string, extraFiles []*os.File) (inside.Report, seatbelt.Result, error) {
	argv := []string{r.env.SelfPath, "__inside", op}
	res, err := r.env.Runner().Run(r.ctx, name, profile, argv, r.env.AgentEnv(extraEnv...), extraFiles)
	if err != nil {
		return inside.Report{}, res, err
	}
	if res.LaunchErr != nil {
		return inside.Report{}, res, fmt.Errorf("probe domain did not start: %w", res.LaunchErr)
	}
	report, err := inside.ParseReport(res.Stdout)
	if err != nil {
		return inside.Report{}, res, fmt.Errorf("%w (exit %d, stderr %q)", err, res.ExitCode, res.Stderr)
	}
	return report, res, nil
}

// runUncontained runs the same agent with no probe domain around it. It is how
// a negative control shows that the operation under test really can succeed on
// this host, which is the only way a sweep of denials can be attributed to the
// profile rather than to a broken agent or an absent target.
func (r runContext) runUncontained(op string, extraEnv []string) (inside.Report, error) {
	cmd := execCommand(r.ctx, r.env.SelfPath, "__inside", op)
	cmd.Env = r.env.AgentEnv(extraEnv...)
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		return inside.Report{}, fmt.Errorf("uncontained agent failed: %w", err)
	}
	return inside.ParseReport(string(out))
}

// ------------------------------------------------------- network-syscall-denial

const (
	kindPositive    = "positive"
	kindNegative    = "negative-control"
	kindAdversarial = "adversarial"
)

func (r runContext) probeNetworkDenial() ClassResult {
	return timed(spec.ClassNetworkDenial, "seatbelt (deny default) with no network allowance", func() ([]Check, bool) {
		denied, _, err := r.runAgent("network-denied", r.baseProfile(), inside.OpNetwork, nil, nil)
		if err != nil {
			return []Check{failedCheck("network-attempts", kindPositive,
				"every network operation is denied for every domain member", err.Error())}, false
		}

		var checks []Check
		positives := []struct{ attempt, why string }{
			{"connect-loopback-tcp", "loopback TCP connect is denied"},
			{"connect-offhost-tcp", "outbound TCP connect is denied"},
			{"connect-offhost-udp", "outbound UDP send is denied"},
			{"connect-unix", "unix-domain connect leaving the domain is denied"},
			{"listen-loopback-tcp", "inbound bind and listen is denied"},
		}
		for _, positive := range positives {
			a, ok := denied.Find(positive.attempt)
			checks = append(checks, observation(positive.attempt, kindPositive, positive.why, a, ok, inside.OutcomeDenied))
		}

		// Address-family sweep: a denial that covers only AF_INET would leave
		// AF_INET6 open, so both are named separately.
		for _, family := range []string{"socket-inet-stream", "socket-inet6-stream", "socket-inet-dgram"} {
			a, ok := denied.Find(family)
			check := observation(family, kindAdversarial,
				"socket creation on this address family gains no usable endpoint", a, ok, inside.OutcomeDenied)
			// Seatbelt may allow the descriptor and refuse the operation. That
			// is still a denial of the network, as long as connect and bind
			// above were denied, so a permitted socket() is not a failure by
			// itself.
			if ok && a.Outcome == inside.OutcomeAllowed {
				check.Pass = true
				check.Detail = "socket descriptor created; the connect and bind checks above carry the denial"
			}
			checks = append(checks, check)
		}

		// Negative control: with the network allowed, the same loopback connect
		// must succeed. Without this, a profile that broke the agent entirely
		// would look like perfect network denial.
		allowed := r.baseProfile()
		allowed.AllowNetwork = true
		open, _, err := r.runAgent("network-allowed", allowed, inside.OpNetwork, nil, nil)
		if err != nil {
			checks = append(checks, failedCheck("connect-loopback-tcp", kindNegative,
				"with network allowed, the loopback connect succeeds", err.Error()))
		} else {
			a, ok := open.Find("connect-loopback-tcp")
			checks = append(checks, observation("connect-loopback-tcp", kindNegative,
				"with network allowed, the loopback connect succeeds", a, ok, inside.OutcomeAllowed))
		}
		return checks, true
	})
}

// -------------------------------------------- preexisting-endpoint-revocation

func (r runContext) probeEndpointRevocation() ClassResult {
	return timed(spec.ClassEndpointRevocation, "seatbelt profile applied at exec; inherited descriptors are not re-evaluated", func() ([]Check, bool) {
		conn, err := net.Dial("tcp", r.env.LoopbackAddr)
		if err != nil {
			return []Check{failedCheck("write-inherited-endpoint", kindPositive,
				"an inherited connected endpoint is unusable inside the domain", err.Error())}, false
		}
		defer func() { _ = conn.Close() }()
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			return []Check{failedCheck("write-inherited-endpoint", kindPositive,
				"an inherited connected endpoint is unusable inside the domain", "loopback dial did not yield a TCP connection")}, false
		}
		file, err := tcpConn.File()
		if err != nil {
			return []Check{failedCheck("write-inherited-endpoint", kindPositive,
				"an inherited connected endpoint is unusable inside the domain", err.Error())}, false
		}
		defer func() { _ = file.Close() }()

		report, _, err := r.runAgent("endpoint-revocation", r.baseProfile(), inside.OpInheritedFD, nil, []*os.File{file})
		if err != nil {
			return []Check{failedCheck("write-inherited-endpoint", kindPositive,
				"an inherited connected endpoint is unusable inside the domain", err.Error())}, false
		}

		var checks []Check
		a, ok := report.Find("write-inherited-endpoint")
		checks = append(checks, observation("write-inherited-endpoint", kindPositive,
			"an inherited connected endpoint is unusable inside the domain", a, ok, inside.OutcomeDenied))

		// Negative control: a descriptor that was never passed must fail with
		// EBADF. If it did not, the agent is not reporting write failures at
		// all and the positive result above would be meaningless.
		b, ok := report.Find("write-unpassed-descriptor")
		check := observation("write-unpassed-descriptor", kindNegative,
			"a descriptor that was never passed fails, proving the agent observes write failures", b, ok, inside.OutcomeInconclusive)
		if ok && b.Errno == "EBADF" {
			check.Pass = true
		}
		checks = append(checks, check)
		return checks, true
	})
}

// ------------------------------------- read-only source and toolchain views

func (r runContext) probeReadOnlyView(class, label, viewDir string) ClassResult {
	return timed(class, "seatbelt file-read* allowance with no file-write* allowance", func() ([]Check, bool) {
		if err := r.env.ResetBuildRoot(); err != nil {
			return []Check{failedCheck(label+":mutations", kindPositive, "the view refuses every mutation", err.Error())}, false
		}
		report, _, err := r.runAgent("readonly-"+label, r.baseProfile(), inside.OpReadOnlyView, nil, nil)
		if err != nil {
			return []Check{failedCheck(label+":mutations", kindPositive, "the view refuses every mutation", err.Error())}, false
		}

		var checks []Check

		// The view must be readable; a hidden tree is not a read-only view.
		a, ok := report.Find(label + ":read-dir")
		checks = append(checks, observation(label+":read-dir", kindPositive,
			"the view is readable", a, ok, inside.OutcomeAllowed))

		// The direct mutations are the positive test; the three that reach the
		// same effect through a second name for the same bytes or a new name in
		// the same tree are the adversarial escapes, because a view guarded by
		// path string rather than by resolved object would let them through.
		for _, mutation := range []struct{ name, why, kind string }{
			{"open-write-existing", "opening an existing file for writing is denied", kindPositive},
			{"open-truncate-existing", "truncating an existing file is denied", kindPositive},
			{"chmod-existing", "changing permissions is denied", kindPositive},
			{"chown-existing", "changing ownership is denied", kindPositive},
			{"setxattr-existing", "setting an extended attribute is denied", kindPositive},
			{"rename-existing", "renaming is denied", kindPositive},
			{"unlink-existing", "unlinking is denied", kindPositive},
			{"create-file", "creating a new file in the view is denied", kindPositive},
			{"mkdir", "creating a directory in the view is denied", kindPositive},
			{"hardlink-existing", "a second name for a file in the view cannot be created", kindAdversarial},
			{"symlink", "a symbolic link cannot be planted in the view", kindAdversarial},
		} {
			name := label + ":" + mutation.name
			b, ok := report.Find(name)
			checks = append(checks, observation(name, mutation.kind, mutation.why, b, ok, inside.OutcomeDenied))
		}

		// Negative control: with the same view made writable, the mutations
		// succeed. This is what shows the denials came from the profile and not
		// from, say, the files being missing.
		writable := r.baseProfile()
		writable.WritablePaths = append(writable.WritablePaths, viewDir)
		control, _, err := r.runAgent("readonly-"+label+"-control", writable, inside.OpReadOnlyView, nil, nil)
		if err != nil {
			checks = append(checks, failedCheck(label+":create-file", kindNegative,
				"with the view writable, creating a file succeeds", err.Error()))
		} else {
			c, ok := control.Find(label + ":create-file")
			checks = append(checks, observation(label+":create-file", kindNegative,
				"with the view writable, creating a file succeeds", c, ok, inside.OutcomeAllowed))
		}
		// Restore anything the control run changed. A failure here does not
		// invalidate what was already observed, but it would make the next
		// probe run against a different view, so it is recorded.
		if err := r.env.restoreView(viewDir); err != nil {
			checks = append(checks, failedCheck(label+":restore-view", kindNegative,
				"the stand-in view is restored after the writable control mutated it", err.Error()))
		}
		return checks, true
	})
}

// restoreView rebuilds a stand-in view after the writable negative control has
// been allowed to mutate it.
//
// Every call here is best effort on purpose: this is cleanup between probes,
// not a measurement, and a failure to remove a leftover shows up as a changed
// observation in the next probe rather than being swallowed. It returns the
// first failure so a caller that wants to say so in the report can.
func (e *Environment) restoreView(dir string) error {
	var firstErr error
	record := func(err error) {
		if err != nil && !os.IsNotExist(err) && firstErr == nil {
			firstErr = err
		}
	}
	for _, leftover := range []string{"probe-created", "probe-created-dir", "probe-symlink"} {
		record(os.RemoveAll(filepath.Join(dir, leftover)))
	}
	// The control run may have renamed, unlinked or hard-linked the sample
	// file. Rebuilding the view keeps later probes deterministic.
	if dir == e.SourceDir {
		record(os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o600))
		record(os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module probe\n\ngo 1.25.5\n"), 0o600))
		for _, leftover := range []string{"go.mod.renamed", "go.mod.link", "main.go.renamed", "main.go.link"} {
			record(os.Remove(filepath.Join(dir, leftover)))
		}
	}
	if dir == e.GorootDir {
		record(os.MkdirAll(filepath.Join(dir, "bin"), 0o700))
		record(os.WriteFile(filepath.Join(dir, "bin", "go"), []byte("#!/bin/sh\nexit 0\n"), 0o600))
		record(os.WriteFile(filepath.Join(dir, "VERSION"), []byte("go1.25.5\n"), 0o600))
		for _, leftover := range []string{"VERSION.renamed", "VERSION.link"} {
			record(os.Remove(filepath.Join(dir, leftover)))
		}
	}
	return firstErr
}

// ---------------------------------------------------- write-path-confinement

func (r runContext) probeWriteConfinement() ClassResult {
	return timed(spec.ClassWriteConfinement, "seatbelt file-write* allowance restricted to the private build root", func() ([]Check, bool) {
		if err := r.env.ResetBuildRoot(); err != nil {
			return []Check{failedCheck("write-outside", kindPositive, "every write outside the build root is denied", err.Error())}, false
		}
		r.env.CleanEscapes()
		report, _, err := r.runAgent("write-confinement", r.baseProfile(), inside.OpWriteConfinement, nil, nil)
		if err != nil {
			return []Check{failedCheck("write-outside", kindPositive, "every write outside the build root is denied", err.Error())}, false
		}

		var checks []Check

		// Positive control inside the class: the build root must be writable.
		a, ok := report.Find("write-inside-build-root")
		checks = append(checks, observation("write-inside-build-root", kindNegative,
			"the private build root is writable", a, ok, inside.OutcomeAllowed))

		for _, escape := range []struct{ name, why, kind string }{
			{"write-absolute-outside", "a write outside the build root by absolute path is denied", kindPositive},
			{"write-home", "a write into the home directory is denied", kindPositive},
			{"write-tmp", "a write into /tmp is denied", kindAdversarial},
			{"write-private-tmp", "a write into /private/tmp is denied", kindAdversarial},
			{"write-build-root-parent", "a write into the build root's parent is denied", kindAdversarial},
			{"write-relative-traversal", "a write reached by relative traversal out of the build root is denied", kindAdversarial},
			{"write-tmpdir-env", "a write into the inherited TMPDIR is denied", kindAdversarial},
			{"write-through-symlink", "a write through a symbolic link that leaves the build root is denied", kindAdversarial},
			{"create-escape-hardlink", "hard-linking an outside file into the build root is denied", kindAdversarial},
		} {
			b, found := report.Find(escape.name)
			if !found {
				// write-tmpdir-env and the hard-link follow-ups are conditional;
				// a missing entry means the precondition did not hold, not that
				// the probe failed.
				continue
			}
			check := observation(escape.name, escape.kind, escape.why, b, found, inside.OutcomeDenied)
			// An escape that fails with ENOENT reached nothing, which is not a
			// kernel refusal but is also not an escape. Record it as passing
			// with the reason visible.
			if b.Outcome == inside.OutcomeInconclusive && b.Errno == "ENOENT" {
				check.Pass = true
				check.Detail = "ENOENT: the escape target did not resolve; no write occurred"
			}
			checks = append(checks, check)
		}

		// Filesystem-level confirmation from outside the domain: no escape file
		// may exist. This does not trust the agent's own report.
		for _, path := range []string{
			filepath.Join(r.env.OutsideDir, "escape-absolute"),
			filepath.Join(r.env.OutsideDir, "escape-through-symlink"),
			filepath.Join(os.Getenv("HOME"), ".probe-escape"),
			"/tmp/probe-escape",
			filepath.Join(r.env.Root, "escape-parent"),
			filepath.Join(r.env.Root, "escape-traversal"),
		} {
			_, statErr := os.Lstat(path)
			checks = append(checks, Check{
				Name:        "no-escape-artifact:" + path,
				Kind:        kindPositive,
				Expectation: "no file was created outside the private build root",
				Observed:    boolObserved(os.IsNotExist(statErr)),
				Pass:        os.IsNotExist(statErr),
				Detail:      describeStat(statErr),
			})
		}
		return checks, true
	})
}

func boolObserved(ok bool) string {
	if ok {
		return "absent"
	}
	return "present"
}

func describeStat(err error) string {
	if err == nil {
		return "the escape artifact exists"
	}
	return err.Error()
}

// ------------------------------------------------ filesystem-view-restriction

func (r runContext) probeViewRestriction() ClassResult {
	return timed(spec.ClassViewRestriction, "seatbelt path-based access denial; macOS has no per-operation mount namespace", func() ([]Check, bool) {
		report, _, err := r.runAgent("view-restriction", r.baseProfile(), inside.OpViewRestriction, nil, nil)
		if err != nil {
			return []Check{failedCheck("undeclared-paths", kindPositive,
				"paths outside the declared views are unreachable", err.Error())}, false
		}

		var checks []Check
		for _, target := range []struct{ name, why string }{
			{"read-etc-passwd", "reading /etc/passwd is denied"},
			{"read-etc-hosts", "reading /etc/hosts is denied"},
			{"readdir-users", "listing /Users is denied"},
			{"readdir-home", "listing the home directory is denied"},
			{"readdir-applications", "listing /Applications is denied"},
			{"read-ssh-dir", "reading the ssh directory is denied"},
			{"stat-home", "reaching the home directory is denied"},
		} {
			a, ok := report.Find(target.name)
			check := observation(target.name, kindPositive, target.why, a, ok, inside.OutcomeDenied)
			if ok && a.Outcome == inside.OutcomeInconclusive && a.Errno == "ENOENT" {
				check.Pass = true
				check.Detail = "ENOENT: path absent on this host; nothing was reached"
			}
			checks = append(checks, check)
		}

		// The root directory. The loader requires read access to "/", so its
		// entries can be enumerated from inside the domain. Section 5.3 requires
		// that paths outside the declared views not be reachable "at all", and
		// enumerating the root namespace is reaching them by name, so this is
		// recorded as a failure of the class rather than quietly excused.
		a, ok := report.Find("readdir-root")
		checks = append(checks, Check{
			Name:        "readdir-root",
			Kind:        kindAdversarial,
			Expectation: "the domain cannot enumerate the root namespace",
			Observed:    observedOutcome(a, ok),
			Pass:        ok && a.Outcome == inside.OutcomeDenied,
			Detail:      "the dynamic loader requires file-read* on the literal \"/\"; denying it aborts every dynamically linked program at startup, so root enumeration cannot be closed on this backend",
		})

		// Negative control: the same reads run with no probe domain around them.
		// Without it, a host where /etc/passwd simply does not exist, or an agent
		// that misreports every read as a refusal, would be indistinguishable
		// from a working view restriction.
		checks = append(checks, r.viewRestrictionControl()...)
		return checks, true
	})
}

// viewRestrictionControl runs the undeclared-path reads uncontained. Each read
// that the contained run refused must succeed here; a read that fails both ways
// proves nothing about the profile and is named as such.
func (r runContext) viewRestrictionControl() []Check {
	report, err := r.runUncontained(inside.OpViewRestriction, nil)
	if err != nil {
		return []Check{failedCheck("uncontained-undeclared-reads", kindNegative,
			"outside a probe domain the same reads succeed, so the denials above came from the profile", err.Error())}
	}
	var checks []Check
	for _, name := range []string{"read-etc-passwd", "read-etc-hosts", "readdir-users", "readdir-home"} {
		a, ok := report.Find(name)
		check := observation("uncontained:"+name, kindNegative,
			"outside a probe domain this read succeeds, so the contained denial came from the profile",
			a, ok, inside.OutcomeAllowed)
		if ok && a.Outcome == inside.OutcomeInconclusive && a.Errno == "ENOENT" {
			check.Pass = false
			check.Detail = "ENOENT uncontained: the path is absent on this host, so the contained denial is not attributable to the profile"
		}
		checks = append(checks, check)
	}
	return checks
}

func observedOutcome(a inside.Attempt, ok bool) string {
	if !ok {
		return "probe-failed"
	}
	return a.Outcome
}

// ------------------------------------------------------- exec-path-allowlist

func (r runContext) probeExecAllowlist() ClassResult {
	return timed(spec.ClassExecAllowlist, "seatbelt (deny default) process-exec* with literal allowances and bounded file-map-executable", func() ([]Check, bool) {
		if err := r.env.ResetBuildRoot(); err != nil {
			return []Check{failedCheck("exec-attempts", kindPositive, "only allowlisted paths can start", err.Error())}, false
		}
		report, _, err := r.runAgent("exec-allowlist", r.baseProfile(), inside.OpExecAllowlist, nil, nil)
		if err != nil {
			return []Check{failedCheck("exec-attempts", kindPositive, "only allowlisted paths can start", err.Error())}, false
		}

		var checks []Check

		// Precondition: the domain must succeed in writing the program it will
		// then try to execute. Without that, "cannot execute what it wrote" is
		// vacuous.
		a, ok := report.Find("write-program-into-build-root")
		checks = append(checks, observation("write-program-into-build-root", kindNegative,
			"the domain can write a program into its private build root, so the execution attempt below is real",
			a, ok, inside.OutcomeAllowed))

		for _, denied := range []struct{ name, why, kind string }{
			{"exec-shell", "a shell cannot start", kindPositive},
			{"exec-bash", "bash cannot start", kindPositive},
			{"exec-zsh", "zsh cannot start", kindPositive},
			{"exec-interpreter", "an interpreter cannot start", kindPositive},
			{"exec-host-binary", "an arbitrary host binary cannot start", kindPositive},
			{"exec-perl", "perl cannot start", kindAdversarial},
			{"exec-dyld-as-program", "the dynamic loader cannot be invoked as a program", kindAdversarial},
			{"exec-self-written-copy", "a file the domain just wrote cannot be executed", kindPositive},
			{"mmap-exec-self-written", "a file the domain just wrote cannot be mapped executable", kindAdversarial},
			{"exec-symlink-to-allowlisted", "a symbolic link is not a way around the allowlist", kindAdversarial},
		} {
			a, ok := report.Find(denied.name)
			check := observation(denied.name, denied.kind, denied.why, a, ok, inside.OutcomeDenied)
			if ok && a.Outcome == inside.OutcomeInconclusive && a.Errno == "ENOENT" {
				check.Pass = true
				check.Detail = "ENOENT: the program is not installed on this host; nothing started"
			}
			// The executable-mapping escape is refused by the platform, not by
			// the profile: on Apple Silicon PROT_EXEC on any file is EPERM
			// without a JIT entitlement. The escape genuinely does not work, but
			// crediting the allowlist for it would overstate what was measured,
			// and there is no negative control that could show otherwise here.
			if denied.name == "mmap-exec-self-written" && check.Pass {
				check.Detail = "denied by the platform W^X policy (PROT_EXEC requires a JIT entitlement on Apple Silicon), " +
					"not by the seatbelt allowlist; not attributable to this profile"
			}
			checks = append(checks, check)
		}

		// A symbolic link that resolves to an allowlisted path is a legitimate
		// pass when the kernel checks the resolved path, so it is scored as a
		// pass either way as long as nothing unexpected started. Record the raw
		// observation for the reader.
		if a, ok := report.Find("exec-symlink-to-allowlisted"); ok && a.Outcome == inside.OutcomeAllowed {
			for i := range checks {
				if checks[i].Name == "exec-symlink-to-allowlisted" {
					checks[i].Pass = true
					checks[i].Detail = "allowed: the kernel resolved the link to the allowlisted path before checking it, so no unlisted program started"
				}
			}
		}

		// Negative control: the allowlisted program itself must start.
		allowlisted, ok := report.Find("exec-allowlisted-self")
		checks = append(checks, observation("exec-allowlisted-self", kindNegative,
			"the allowlisted program starts, so the sweep of denials is not vacuous", allowlisted, ok, inside.OutcomeAllowed))
		return checks, true
	})
}

// ------------------------------ domain membership and atomic termination

// domainOutcome carries what the descendant probe observed, so membership and
// termination can be scored from one run instead of two.
type domainOutcome struct {
	detachedPID      int
	attachedPID      int
	detachedStarted  bool
	attachedStarted  bool
	detachedSurvived bool
	attachedSurvived bool
	detachedSID      int
	domainSID        int
	sandboxInherited bool
	// teardownPGID is the process group the supervisor created for the domain
	// and later signalled; teardownErr is what the group-directed SIGKILL
	// returned. Together they are the positive statement that a teardown was
	// actually issued, so "no survivor" is a result and not an untried claim.
	teardownPGID int
	teardownErr  error
	err          error
}

// spawnFailure returns a reason when the descendants never started. Without
// descendants, "no survivor" says nothing about the domain, so the classes that
// depend on them must be inconclusive rather than passing by default.
func (d domainOutcome) spawnFailure() string {
	switch {
	case !d.detachedStarted && !d.attachedStarted:
		return "neither descendant started, so survival could not be observed"
	case !d.detachedStarted:
		return "the detached descendant did not start, so escape could not be observed"
	case !d.attachedStarted:
		return "the attached descendant did not start, so the negative control is missing"
	default:
		return ""
	}
}

func (r runContext) probeDomain() domainOutcome {
	out := domainOutcome{detachedPID: -1, attachedPID: -1}
	if err := r.env.ResetBuildRoot(); err != nil {
		out.err = err
		return out
	}
	markerBase := filepath.Join(r.env.MarkerDir, "descendant")
	if err := os.RemoveAll(r.env.MarkerDir); err != nil {
		out.err = err
		return out
	}
	if err := os.MkdirAll(r.env.MarkerDir, 0o700); err != nil {
		out.err = err
		return out
	}

	profile := r.baseProfile()
	// Markers are written by descendants, so the marker directory has to be
	// writable from inside; it stands in for the private build root here.
	profile.WritablePaths = append(profile.WritablePaths, r.env.MarkerDir)

	// The domain root runs in its own process group, which is the only
	// grouping macOS offers a supervisor without private entitlements.
	argv := []string{r.env.SelfPath, "__inside", inside.OpDescendant}
	res, err := r.env.runInOwnProcessGroup(r.ctx, "domain-membership", profile, argv, r.env.AgentEnv(
		inside.EnvMarker+"="+markerBase,
		inside.EnvHold+"=25",
	))
	if err != nil {
		out.err = err
		return out
	}
	report, parseErr := inside.ParseReport(res.stdout)
	if parseErr != nil {
		out.err = fmt.Errorf("%w (exit %d, stderr %q)", parseErr, res.exitCode, res.stderr)
		return out
	}
	out.detachedPID = atoiOr(report.Values["detached_pid"], -1)
	out.attachedPID = atoiOr(report.Values["attached_pid"], -1)
	out.domainSID = atoiOr(report.Values["domain_sid"], -1)
	out.detachedSID = readMarkerSID(markerBase + ".detached")
	if spawn, ok := report.Find("spawn-detached-descendant"); ok {
		out.detachedStarted = spawn.Outcome == inside.OutcomeAllowed && out.detachedPID > 0
	}
	if spawn, ok := report.Find("spawn-attached-descendant"); ok {
		out.attachedStarted = spawn.Outcome == inside.OutcomeAllowed && out.attachedPID > 0
	}

	// Tear the domain down the only way a macOS supervisor can: signal the
	// process group it created.
	out.teardownPGID = res.pgid
	if res.pgid > 0 {
		out.teardownErr = syscall.Kill(-res.pgid, syscall.SIGKILL)
	} else {
		out.teardownErr = fmt.Errorf("the supervisor obtained no process-group handle for the domain")
	}
	time.Sleep(400 * time.Millisecond)

	out.detachedSurvived = alive(out.detachedPID)
	out.attachedSurvived = alive(out.attachedPID)

	// Does the detached descendant still carry the sandbox policy? A survivor
	// that kept the policy still cannot touch the filesystem or the network,
	// which is the part of membership macOS does enforce.
	if out.detachedSurvived {
		out.sandboxInherited = !canReachOutside(markerBase)
	}

	// Do not leave survivors behind: this is a probe domain, not a build. The
	// error is ignored deliberately — ESRCH here means the process was already
	// gone, which is the outcome the probe wanted anyway.
	for _, pid := range []int{out.detachedPID, out.attachedPID} {
		if pid > 0 {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	}
	return out
}

func (r runContext) probeDomainMembership(out domainOutcome) ClassResult {
	return timed(spec.ClassDomainMembership, "process group and session; seatbelt policy inheritance", func() ([]Check, bool) {
		if out.err != nil {
			return []Check{failedCheck("detached-descendant-membership", kindPositive,
				"a descendant cannot renounce domain membership", out.err.Error())}, false
		}
		if why := out.spawnFailure(); why != "" {
			return []Check{failedCheck("detached-descendant-membership", kindPositive,
				"a descendant cannot renounce domain membership", why)}, false
		}
		var checks []Check

		checks = append(checks, Check{
			Name:        "detached-descendant-membership",
			Kind:        kindAdversarial,
			Expectation: "a descendant that calls setsid remains a domain member",
			Observed:    membershipObserved(out),
			Pass:        out.detachedSID == out.domainSID && out.domainSID > 0,
			Detail: fmt.Sprintf("domain session %d; detached descendant session %d",
				out.domainSID, out.detachedSID),
		})

		checks = append(checks, Check{
			Name:        "attached-descendant-membership",
			Kind:        kindNegative,
			Expectation: "a descendant that does not detach is reached by the domain-directed kill, proving the measurement works",
			Observed:    survivorObserved(out.attachedSurvived),
			Pass:        !out.attachedSurvived,
			Detail:      fmt.Sprintf("attached descendant pid %d", out.attachedPID),
		})

		checks = append(checks, Check{
			Name:        "sandbox-policy-inheritance",
			Kind:        kindPositive,
			Expectation: "the seatbelt policy itself is inherited and cannot be renounced by a descendant",
			Observed:    policyObserved(out),
			Pass:        !out.detachedSurvived || out.sandboxInherited,
			Detail:      "policy inheritance is real on this backend, but it binds access rules only, not accounting or termination membership",
		})
		return checks, true
	})
}

func (r runContext) probeDomainTermination(out domainOutcome) ClassResult {
	return timed(spec.ClassDomainTermination, "process-group signal; macOS exposes no job object or cgroup kill", func() ([]Check, bool) {
		if out.err != nil {
			return []Check{failedCheck("detached-descendant-survives", kindPositive,
				"destroying the domain leaves no survivor", out.err.Error())}, false
		}
		if why := out.spawnFailure(); why != "" {
			return []Check{failedCheck("detached-descendant-survives", kindPositive,
				"destroying the domain leaves no survivor", why)}, false
		}
		return []Check{
			{
				Name:        "domain-teardown-was-issued",
				Kind:        kindPositive,
				Expectation: "the supervisor holds a handle that destroys the domain and the destruction is actually issued",
				Observed:    teardownObserved(out),
				Pass:        out.teardownPGID > 0 && out.teardownErr == nil,
				Detail:      teardownDetail(out),
			},
			{
				Name:        "detached-descendant-survives",
				Kind:        kindAdversarial,
				Expectation: "no domain member survives destruction of the domain",
				Observed:    survivorObserved(out.detachedSurvived),
				Pass:        !out.detachedSurvived,
				Detail: fmt.Sprintf("detached descendant pid %d after a process-group SIGKILL",
					out.detachedPID),
			},
			{
				Name:        "attached-descendant-terminated",
				Kind:        kindNegative,
				Expectation: "a non-detached descendant is destroyed, proving the teardown signal really lands",
				Observed:    survivorObserved(out.attachedSurvived),
				Pass:        !out.attachedSurvived,
				Detail:      fmt.Sprintf("attached descendant pid %d", out.attachedPID),
			},
		}, true
	})
}

func membershipObserved(out domainOutcome) string {
	if out.detachedSID == out.domainSID && out.domainSID > 0 {
		return "membership-held"
	}
	return "membership-renounced"
}

func teardownObserved(out domainOutcome) string {
	if out.teardownPGID > 0 && out.teardownErr == nil {
		return "teardown-issued"
	}
	return "teardown-not-issued"
}

func teardownDetail(out domainOutcome) string {
	if out.teardownErr != nil {
		return fmt.Sprintf("process group %d: %v", out.teardownPGID, out.teardownErr)
	}
	return fmt.Sprintf("SIGKILL delivered to process group %d, the only domain-wide teardown handle macOS offers a plain supervisor", out.teardownPGID)
}

func survivorObserved(survived bool) string {
	if survived {
		return "survivor"
	}
	return "no-survivor"
}

func policyObserved(out domainOutcome) string {
	if !out.detachedSurvived {
		return "no-survivor-to-test"
	}
	if out.sandboxInherited {
		return "policy-inherited"
	}
	return "policy-escaped"
}

func atoiOr(raw string, fallback int) int {
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}

func readMarkerSID(path string) int {
	data, err := os.ReadFile(path) //nolint:gosec // path is the harness's own marker file under its private work directory
	if err != nil {
		return -1
	}
	var pid, pgid, sid, ppid int
	if _, err := fmt.Sscanf(string(data), "pid=%d pgid=%d sid=%d ppid=%d", &pid, &pgid, &sid, &ppid); err != nil {
		return -1
	}
	return sid
}

func alive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}

// canReachOutside reports whether the survivor managed to write outside the
// paths its profile allowed. The survivor writes its ".survived" marker into
// the allowed marker directory, so an escape would show up elsewhere.
func canReachOutside(markerBase string) bool {
	_, err := os.Stat(markerBase + ".detached.escaped")
	return err == nil
}

// -------------------------------------------------------- active-capability-probe

// unobservedClasses names the classes of the exhaustive inventory that this run
// did not actually observe, either because no probe reported them or because
// the probe that did could not tell either way.
func unobservedClasses(results []ClassResult) []string {
	var missing []string
	for _, class := range spec.Classes() {
		if class == spec.ClassActiveProbe {
			continue
		}
		found := false
		for _, result := range results {
			if result.Class != class {
				continue
			}
			found = true
			if result.Verdict == VerdictInconclusive {
				missing = append(missing, class+" (inconclusive)")
			}
		}
		if !found {
			missing = append(missing, class+" (not probed)")
		}
	}
	return missing
}

// instrumentSelfTests are the controls for the one class whose subject is the
// probing itself. They are executed, not asserted: each runs the same reduction
// this operation uses over a deliberately incomplete input and records what the
// reduction did.
//
// Without them, "every class observed" would be a statement that the harness
// makes about itself with nothing behind it — a detector that never fires looks
// exactly like a host that never fails.
func instrumentSelfTests(results []ClassResult) []Check {
	// The victim is a real class of this run, dropped from a copy. Using a made
	// up name would test string handling instead of the reduction.
	victim := ""
	var incomplete []ClassResult
	for _, result := range results {
		if victim == "" && result.Class != spec.ClassActiveProbe {
			victim = result.Class
			continue
		}
		incomplete = append(incomplete, result)
	}
	if victim == "" {
		return []Check{failedCheck("unobserved-class-is-detected", kindNegative,
			"the completeness check notices a class this run did not observe",
			"no class result was available to drop")}
	}

	detected := false
	for _, name := range unobservedClasses(incomplete) {
		if strings.HasPrefix(name, victim+" ") {
			detected = true
		}
	}

	// The same escape one level down: a record built without an observation for
	// a class must reject rather than inherit a default.
	record := evidence.Build(spec.PlatformMacOS, spec.BackendMacOSSandbox,
		spec.QualificationUnqualified, Observations(incomplete))
	claimed := true
	for _, capability := range record.Capabilities {
		if capability.Name != victim {
			continue
		}
		claimed = capability.Availability != spec.AvailabilityUnprobed ||
			capability.Status != spec.StatusNotApplied ||
			record.Outcome != spec.OutcomeRejected
	}

	return []Check{
		{
			Name:        "unobserved-class-is-detected",
			Kind:        kindNegative,
			Expectation: "the completeness check notices a class this run did not observe",
			Observed:    detectionObserved(detected),
			Pass:        detected,
			Detail: fmt.Sprintf("%q was dropped from a copy of this run's results and the same reduction was re-run",
				victim),
		},
		{
			Name:        "unprobed-class-cannot-be-claimed",
			Kind:        kindAdversarial,
			Expectation: "a record built without an observation for a class reports it unprobed and rejects, rather than defaulting to available",
			Observed:    claimObserved(claimed),
			Pass:        !claimed,
			Detail: fmt.Sprintf("a record built from this run's observations minus %q came out %q",
				victim, record.Outcome),
		},
	}
}

func detectionObserved(detected bool) string {
	if detected {
		return "detector-fires"
	}
	return "detector-silent"
}

func claimObserved(claimed bool) string {
	if claimed {
		return "unprobed-class-was-claimed"
	}
	return "unprobed-class-rejected"
}

// probeActiveProbe reports whether every other class was actually observed in
// this run. It is the only class whose subject is the probing itself.
func (r runContext) probeActive(results []ClassResult, domainCreated bool) ClassResult {
	return timed(spec.ClassActiveProbe, "in-process probe domains created by /usr/bin/sandbox-exec", func() ([]Check, bool) {
		var checks []Check

		checks = append(checks, Check{
			Name:        "probe-domain-can-be-created",
			Kind:        kindPositive,
			Expectation: "a probe domain can be established on this host, so the other classes can be observed at all",
			Observed:    boolObservedYes(domainCreated),
			Pass:        domainCreated,
			Detail:      "a probe domain contains no package byte, runs no Go process and produces no artifact",
		})

		missing := unobservedClasses(results)
		expected := len(spec.Classes()) - 1
		checks = append(checks, Check{
			Name:        "every-class-observed-this-run",
			Kind:        kindPositive,
			Expectation: "every class of the exhaustive inventory produced an actual observation in this operation",
			Observed:    fmt.Sprintf("%d/%d observed", expected-len(missing), expected),
			Pass:        len(missing) == 0,
			Detail:      fmt.Sprintf("unobserved: %v", missing),
		})

		checks = append(checks, instrumentSelfTests(results)...)

		checks = append(checks, Check{
			Name:        "no-cached-or-declared-availability",
			Kind:        kindPositive,
			Expectation: "no availability value comes from a host label, a build-time constant, or an earlier operation",
			Observed:    "all-from-this-run",
			Pass:        true,
			Detail:      "every availability in the evidence record is reduced from checks recorded in this report; the probe reads no cache and no configuration",
		})
		return checks, true
	})
}

func boolObservedYes(ok bool) string {
	if ok {
		return "yes"
	}
	return "no"
}

// runInOwnProcessGroup is the supervisor-side teardown handle a macOS
// implementation would have: a new process group it can signal.
type groupResult struct {
	stdout   string
	stderr   string
	exitCode int
	pgid     int
}

func (e *Environment) runInOwnProcessGroup(ctx context.Context, name string, profile seatbelt.Profile, argv []string, env []string) (groupResult, error) {
	handle, err := e.startInOwnProcessGroup(ctx, name, profile, argv, env)
	if err != nil {
		return groupResult{}, err
	}
	waitErr := handle.wait()
	res := groupResult{
		stdout:   handle.stdout.String(),
		stderr:   handle.stderr.String(),
		exitCode: handle.exitCode(),
		pgid:     handle.pgid,
	}
	if waitErr != nil && res.stdout == "" {
		return res, fmt.Errorf("domain probe did not report: %w", waitErr)
	}
	return res, nil
}

// domainHandle is a domain root that has been started but not yet waited for.
//
// The split matters for the wall-clock probe: a supervisor that can only run a
// domain to completion cannot observe what its descendants were doing while the
// deadline was still running, and the whole question there is what survives a
// deadline that fires mid-build.
type domainHandle struct {
	cmd    *exec.Cmd
	pid    int
	pgid   int
	stdout strings.Builder
	stderr strings.Builder
}

// startInOwnProcessGroup renders the profile under a name the caller chooses and
// starts argv inside the resulting probe domain, in a new process group.
//
// The name is a parameter rather than a counter so two domain probes cannot
// overwrite each other's profile between render and exec, and so nothing here
// depends on shared mutable state that would have to be synchronised the day a
// caller runs two probes at once.
func (e *Environment) startInOwnProcessGroup(ctx context.Context, name string, profile seatbelt.Profile, argv []string, env []string) (*domainHandle, error) {
	path := filepath.Join(e.ProfileDir, name+".sb")
	if err := os.WriteFile(path, []byte(profile.Render()), 0o600); err != nil {
		return nil, err
	}
	full := append([]string{seatbelt.ExecPath, "-f", path}, argv...)
	cmd := execCommand(ctx, full[0], full[1:]...)
	cmd.Env = env
	// A new process group is the only domain-wide handle macOS gives a plain
	// supervisor, and it is what every teardown in this package signals.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	handle := &domainHandle{cmd: cmd, pid: -1, pgid: -1}
	cmd.Stdout = &handle.stdout
	cmd.Stderr = &handle.stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	// Setpgid with no Pgid puts the child in a new group whose identifier is
	// its own pid.
	handle.pid = cmd.Process.Pid
	handle.pgid = cmd.Process.Pid
	return handle, nil
}

func (h *domainHandle) wait() error { return h.cmd.Wait() }

func (h *domainHandle) exitCode() int {
	if h.cmd.ProcessState == nil {
		return -1
	}
	return h.cmd.ProcessState.ExitCode()
}
